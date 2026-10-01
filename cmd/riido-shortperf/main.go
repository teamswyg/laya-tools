// riido-shortperf measures a pinned, public CPU-only short-claim fixture.
// Each child is a fresh process. No native model or external service is used.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type scenario struct {
	Kind      string          `json:"baseline"`
	Mode      string          `json:"mode"`
	Repeat    int             `json:"repeat"`
	WallMS    float64         `json:"process_wall_ms"`
	UserCPU   float64         `json:"process_user_cpu_seconds"`
	SystemCPU float64         `json:"process_system_cpu_seconds"`
	RSS       int64           `json:"process_peak_rss_bytes"`
	Output    json.RawMessage `json:"output"`
}

type boundedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, errors.New("child_output_limit")
	}
	return b.Buffer.Write(p)
}

func readBounded(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("file_read_failed")
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(b)) > limit {
		return nil, errors.New("file_read_bounds_or_io")
	}
	return b, nil
}

func measure(parent context.Context, binary string, input []byte, kind, mode, inputDigest, childGo string, repeat, iterations int) (scenario, error) {
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	args := []string{"--baseline", kind}
	if mode == "warm_stages" {
		args = append(args, "--benchmark", "--iterations", fmt.Sprint(iterations))
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = append(os.Environ(), "GOMAXPROCS=1")
	cmd.Stdin = bytes.NewReader(input)
	stdout, stderr := boundedBuffer{limit: 32 << 10}, boundedBuffer{limit: 4 << 10}
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	start := time.Now()
	err := cmd.Run()
	r := scenario{Kind: kind, Mode: mode, Repeat: repeat, WallMS: float64(time.Since(start).Nanoseconds()) / 1e6}
	if err != nil || ctx.Err() != nil {
		return r, errors.New("shortclaim_child_failed")
	}
	if stderr.Len() != 0 || !json.Valid(stdout.Bytes()) || stdout.Len() > 32<<10 {
		return r, errors.New("invalid_child_output")
	}
	var header struct {
		Schema          string `json:"schema"`
		Baseline        string `json:"baseline"`
		Profile         bool   `json:"cpu_profile_enabled"`
		Candidates      int    `json:"candidate_count"`
		InputDigest     string `json:"input_sha256"`
		CanonicalDigest string `json:"canonical_input_sha256"`
		Go              string `json:"go"`
		Threads         int    `json:"gomaxprocs"`
		GPU             bool   `json:"gpu_used"`
		Encoder         bool   `json:"encoder_used"`
		CacheHits       int    `json:"cache_hits"`
		Stages          []struct {
			Stage       string `json:"stage"`
			Repetitions int    `json:"repetitions"`
		} `json:"stages"`
		Warmups          int `json:"warmups_per_stage"`
		AllocationProbes int `json:"allocation_probes_per_stage_including_warmup"`
		Order            []struct {
			ID string `json:"id"`
		} `json:"verification_order"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &header); err != nil || header.Baseline != kind {
		return r, errors.New("child_contract_mismatch")
	}
	if mode == "warm_stages" {
		if header.Schema != "riido-shortclaim-benchmark-v1" || header.Candidates != 8 || header.Profile || header.Go != childGo || header.Threads != 1 || header.GPU || header.Encoder || header.CacheHits != 0 || header.CanonicalDigest != inputDigest || len(header.Stages) != 5 || header.Warmups != 20 || header.AllocationProbes != 101 {
			return r, errors.New("warm_child_contract_mismatch")
		}
		for i, name := range []string{"parse_validate_normalize", "rank_validate_features_score_order", "input_digest", "ranking_serialization", "full_go_request"} {
			if header.Stages[i].Stage != name || header.Stages[i].Repetitions != iterations {
				return r, errors.New("warm_stage_count_mismatch")
			}
		}
	} else if header.Schema != "riido-shortclaim-order-v1" || len(header.Order) != 8 {
		return r, errors.New("cold_child_contract_mismatch")
	} else {
		if header.InputDigest != inputDigest {
			return r, errors.New("cold_digest_mismatch")
		}
		var seen [8]bool
		for _, candidate := range header.Order {
			idx := -1
			for i := 0; i < 8; i++ {
				if candidate.ID == fmt.Sprintf("candidate-%d", i) {
					idx = i
					break
				}
			}
			if idx < 0 || seen[idx] {
				return r, errors.New("cold_candidate_permutation_mismatch")
			}
			seen[idx] = true
		}
	}
	u, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok || u.Maxrss <= 0 {
		return r, errors.New("missing_peak_rss")
	}
	r.UserCPU, r.SystemCPU = cmd.ProcessState.UserTime().Seconds(), cmd.ProcessState.SystemTime().Seconds()
	r.RSS = u.Maxrss
	if runtime.GOOS == "linux" {
		r.RSS *= 1024
	}
	r.Output = append(json.RawMessage(nil), stdout.Bytes()...)
	return r, nil
}

func run() error {
	fs := flag.NewFlagSet("riido-shortperf", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	binary := fs.String("binary", ".cache/riido-shortclaim-56", "compiled trimmed Go binary")
	planFile := fs.String("plan", "experiments/short-claim/execution-plan-56a.json", "pinned diagnostic plan")
	planSHA := fs.String("plan-sha256", "", "required frozen plan SHA-256")
	inputFile := fs.String("input", "experiments/short-claim/benchmark-input-56.json", "original pinned fixture, not user input")
	out := fs.String("out", "", "new result file, never overwritten")
	if fs.Parse(os.Args[1:]) != nil {
		return errors.New("invalid_arguments")
	}
	if fs.NArg() != 0 || len(*planSHA) != 64 || *out == "" || runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return errors.New("require_pinned_plan_new_output_and_supported_os")
	}
	p, err := readBounded(*planFile, 64<<10)
	if err != nil || digest(p) != *planSHA {
		return errors.New("plan_hash_mismatch")
	}
	var plan struct {
		Schema            string `json:"schema"`
		State             string `json:"state"`
		InputSHA          string `json:"benchmark_input_sha256"`
		Iterations        int    `json:"benchmark_iterations"`
		Repeats           int    `json:"benchmark_cold_repeats"`
		Threads           int    `json:"cpu_threads"`
		Fits              int    `json:"fits_authorized_in_this_plan"`
		BinarySHA         string `json:"benchmark_binary_sha256"`
		BuildGo           string `json:"benchmark_build_go"`
		CanonicalInputSHA string `json:"benchmark_canonical_input_sha256"`
		Sources           []struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
		} `json:"implementation_files"`
	}
	if json.Unmarshal(p, &plan) != nil || plan.Schema != "riido-shortclaim-nonlearned-execution-plan-v1" || plan.State != "nonlearned_diagnostics_only" || plan.Iterations != 2000 || plan.Repeats != 3 || plan.Threads != 1 || plan.Fits != 0 || len(plan.Sources) == 0 || len(plan.BinarySHA) != 64 || len(plan.CanonicalInputSHA) != 64 {
		return errors.New("unsupported_plan")
	}
	for _, source := range plan.Sources {
		b, e := readBounded(source.Path, 1<<20)
		if e != nil || digest(b) != source.SHA {
			return errors.New("implementation_hash_mismatch")
		}
	}
	input, err := readBounded(*inputFile, 12<<10)
	if err != nil || digest(input) != plan.InputSHA {
		return errors.New("fixture_hash_mismatch")
	}
	binaryBytes, err := readBounded(*binary, 32<<20)
	if err != nil || digest(binaryBytes) != plan.BinarySHA {
		return errors.New("binary_read_failed")
	}
	info, err := buildinfo.Read(bytes.NewReader(binaryBytes))
	if err != nil || info.GoVersion != plan.BuildGo || info.Path != "github.com/teamswyg/laya-tools/cmd/riido-shortclaim" {
		return errors.New("binary_build_identity_mismatch")
	}
	// Execute a private copy made from the verified bytes, avoiding a second
	// read of a replaceable --binary path between verification and execution.
	dir, err := os.MkdirTemp("", "riido-shortperf-")
	if err != nil {
		return errors.New("binary_copy_directory_failed")
	}
	defer os.RemoveAll(dir)
	executable := filepath.Join(dir, "riido-shortclaim")
	if os.WriteFile(executable, binaryBytes, 0500) != nil {
		return errors.New("binary_copy_failed")
	}
	rows := []scenario{}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for _, kind := range []string{"fixed_order", "bm25", "lexical_ordered", "narrow_rule"} {
		for repeat := 1; repeat <= plan.Repeats; repeat++ {
			r, e := measure(ctx, executable, input, kind, "cold_process", plan.CanonicalInputSHA, plan.BuildGo, repeat, plan.Iterations)
			if e != nil {
				return e
			}
			rows = append(rows, r)
		}
		r, e := measure(ctx, executable, input, kind, "warm_stages", plan.CanonicalInputSHA, plan.BuildGo, 1, plan.Iterations)
		if e != nil {
			return e
		}
		rows = append(rows, r)
		if time.Since(start) > time.Minute {
			return errors.New("performance_wall_budget_exceeded")
		}
	}
	r := struct {
		Schema      string     `json:"schema"`
		PlanSHA     string     `json:"plan_sha256"`
		InputSHA    string     `json:"fixture_file_sha256"`
		BinarySHA   string     `json:"binary_sha256"`
		OS          string     `json:"os"`
		Arch        string     `json:"arch"`
		Go          string     `json:"go"`
		ChildGo     string     `json:"child_build_go"`
		WallSeconds float64    `json:"replay_wall_seconds"`
		Rows        []scenario `json:"rows"`
		Scope       string     `json:"scope"`
	}{"riido-shortclaim-process-performance-v1", *planSHA, digest(input), digest(binaryBytes), runtime.GOOS, runtime.GOARCH, runtime.Version(), info.GoVersion, time.Since(start).Seconds(), rows, "Sequential fresh-process replay, 3 cold repetitions and one five-stage warm benchmark per baseline, original eight-candidate 32-word request fixture, GOMAXPROCS=1. Executed private copy of expected digest/build identity; OS child peak RSS includes all child allocations; CPU profile disabled. This is not an arrival-rate load test, learned inference, GPU result, or LLM savings measurement."}
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("output_create_failed")
	}
	if _, err := f.Write(append(encoded, '\n')); err != nil {
		_ = f.Close()
		return errors.New("output_write_failed")
	}
	if f.Close() != nil {
		return errors.New("output_close_failed")
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
