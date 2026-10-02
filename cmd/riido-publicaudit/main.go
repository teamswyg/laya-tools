// riido-publicaudit collects finite public API observations. It never fits or
// loads a decision model, ranks candidates, or accepts generated code.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/publicbehavior"
)

//go:embed main.go
var compiledMain []byte

const (
	probesPath = "experiments/public-behavior/probes-57.json"
	planPath   = "experiments/public-behavior/execution-plan-57.json"
	maxBytes   = 1 << 20
)

var supportPaths = [...]string{
	"experiments/public-behavior/PLAN-57.en.md",
	"experiments/public-behavior/PLAN-57.ko.md",
	"experiments/public-behavior/oracle-review-57.json",
	"experiments/public-behavior/build-recipe-57.json",
	"cmd/riido-publicaudit/main_test.go",
	"go.mod", "go.sum",
}

type artifact = publicbehavior.SourceArtifact

type plan struct {
	Schema           string     `json:"schema"`
	SourceCommit     string     `json:"source_freeze_commit"`
	GoVersion        string     `json:"go_version"`
	GOOS             string     `json:"goos"`
	GOARCH           string     `json:"goarch"`
	BinarySHA256     string     `json:"binary_sha256"`
	BinaryBytes      int        `json:"binary_bytes"`
	ProbesSHA256     string     `json:"probes_sha256"`
	ProbesBytes      int        `json:"probes_bytes"`
	Sources          []artifact `json:"compiled_source_and_upstream_artifacts"`
	Support          []artifact `json:"support_artifacts"`
	OfficialAttempts int        `json:"planned_official_attempts"`
	Retries          int        `json:"planned_official_retries"`
	CPUThreads       int        `json:"go_max_procs"`
	HeapSoftLimit    int64      `json:"go_heap_soft_limit_bytes"`
}

type record struct {
	Schema           string                     `json:"schema"`
	SourceCommit     string                     `json:"source_freeze_commit"`
	InputCommit      string                     `json:"input_freeze_commit"`
	ProbesSHA256     string                     `json:"probes_sha256"`
	PlanSHA256       string                     `json:"plan_sha256"`
	BinarySHA256     string                     `json:"binary_sha256"`
	GoVersion        string                     `json:"go_version"`
	GOOS             string                     `json:"goos"`
	GOARCH           string                     `json:"goarch"`
	VerifiedGitBlobs int                        `json:"git_blobs_verified_before_observations"`
	Audit            publicbehavior.AuditResult `json:"audit"`
	Limitations      []string                   `json:"limitations"`
}

func sha(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func hash(s string, n int) bool {
	_, err := hex.DecodeString(s)
	return len(s) == n && strings.ToLower(s) == s && err == nil
}

func bounded(root *os.Root, path string) ([]byte, error) {
	f, err := root.Open(path)
	if err != nil {
		return nil, errors.New("publicaudit_file_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return nil, errors.New("publicaudit_file_bounds")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(raw) > maxBytes {
		return nil, errors.New("publicaudit_file_bounds")
	}
	return raw, nil
}

func decodeCanonical(raw []byte, value any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return errors.New("publicaudit_invalid_json")
	}
	canon, err := publicbehavior.Canonical(value)
	if err != nil || !bytes.Equal(canon, raw) {
		return errors.New("publicaudit_noncanonical_json")
	}
	return nil
}

func sourceArtifacts() []artifact {
	upstream := publicbehavior.SourceArtifacts()
	out := make([]artifact, 0, len(upstream)+2)
	out = append(out, upstream[:]...)
	out = append(out, publicbehavior.AuditSourceArtifact(), artifact{Path: "cmd/riido-publicaudit/main.go", Bytes: len(compiledMain), SHA256: sha(compiledMain)})
	slices.SortFunc(out, func(a, b artifact) int { return strings.Compare(a.Path, b.Path) })
	return out
}

func support(root *os.Root) ([]artifact, error) {
	out := make([]artifact, len(supportPaths))
	for i, path := range supportPaths {
		raw, err := bounded(root, path)
		if err != nil {
			return nil, err
		}
		out[i] = artifact{Path: path, Bytes: len(raw), SHA256: sha(raw)}
	}
	return out, nil
}

func verifyBlobs(repo string, root *os.Root, commit string, entries []artifact) error {
	if !hash(commit, 40) {
		return errors.New("publicaudit_commit_invalid")
	}
	for _, a := range entries {
		if !hash(a.SHA256, 64) || a.Bytes < 0 || a.Bytes > maxBytes {
			return errors.New("publicaudit_pin_invalid")
		}
		raw, err := bounded(root, a.Path)
		if err != nil || len(raw) != a.Bytes || sha(raw) != a.SHA256 {
			return errors.New("publicaudit_disk_pin_mismatch")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "git", "show", commit+":"+a.Path)
		cmd.Dir = repo
		var output boundedBuffer
		cmd.Stdout = &output
		err = cmd.Run()
		cancel()
		if err != nil || !bytes.Equal(output.raw, raw) {
			return errors.New("publicaudit_git_pin_mismatch")
		}
	}
	return nil
}

func verifyGoClosure(root *os.Root, sources []artifact) error {
	for _, dir := range []string{"cmd/riido-publicaudit", "internal/publicbehavior", "internal/publicbehavior/testdata/upstream/semver", "internal/publicbehavior/testdata/upstream/doublestar"} {
		entries, err := fs.ReadDir(root.FS(), dir)
		if err != nil {
			return errors.New("publicaudit_source_directory_unavailable")
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			found := false
			for _, a := range sources {
				if a.Path == dir+"/"+e.Name() {
					found = true
				}
			}
			if e.IsDir() || !found {
				return errors.New("publicaudit_unlisted_go_source")
			}
		}
	}
	return nil
}

type boundedBuffer struct{ raw []byte }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(b.raw)+len(p) > maxBytes {
		return 0, errors.New("publicaudit_git_blob_bounds")
	}
	b.raw = append(b.raw, p...)
	return len(p), nil
}

func binaryArtifact() (artifact, error) {
	path, err := os.Executable()
	if err != nil {
		return artifact{}, errors.New("publicaudit_binary_unavailable")
	}
	f, err := os.Open(path)
	if err != nil {
		return artifact{}, errors.New("publicaudit_binary_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 64<<20 {
		return artifact{}, errors.New("publicaudit_binary_bounds")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, (64<<20)+1))
	if err != nil || n != info.Size() {
		return artifact{}, errors.New("publicaudit_binary_unavailable")
	}
	return artifact{Bytes: int(n), SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func writeNew(path string, raw []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("publicaudit_output_exists_or_unavailable")
	}
	_, err = f.Write(raw)
	cerr := f.Close()
	if err != nil || cerr != nil {
		return errors.New("publicaudit_output_write_failed")
	}
	return nil
}

func execute(args []string) error {
	if len(args) == 0 || (args[0] != "plan" && args[0] != "audit") {
		return errors.New("usage: riido-publicaudit plan|audit --repo PATH --out NEW_PATH; plan --source-commit SHA, audit --input-commit SHA")
	}
	flags := flag.NewFlagSet("riido-publicaudit", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repo := flags.String("repo", ".", "public repository checkout")
	out := flags.String("out", "", "new plan file or new audit directory")
	source := flags.String("source-commit", "", "source freeze for plan")
	input := flags.String("input-commit", "", "input freeze for audit")
	if flags.Parse(args[1:]) != nil || flags.NArg() != 0 || *out == "" {
		return errors.New("publicaudit_arguments")
	}
	if runtime.Version() != "go1.27.1" {
		return errors.New("publicaudit_toolchain_mismatch")
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	root, err := os.OpenRoot(*repo)
	if err != nil {
		return errors.New("publicaudit_repository_unavailable")
	}
	defer root.Close()
	rawProbes, err := bounded(root, probesPath)
	if err != nil {
		return err
	}
	var probes publicbehavior.Probes
	if err := decodeCanonical(rawProbes, &probes); err != nil {
		return err
	}
	if err := publicbehavior.ValidateProbes(probes); err != nil {
		return err
	}
	sources := sourceArtifacts()
	sup, err := support(root)
	if err != nil {
		return err
	}
	if err := verifyGoClosure(root, sources); err != nil {
		return err
	}
	binary, err := binaryArtifact()
	if err != nil {
		return err
	}
	if args[0] == "plan" {
		if *input != "" || len(sources) != 19 {
			return errors.New("publicaudit_plan_arguments")
		}
		if err := verifyBlobs(*repo, root, *source, append(slices.Clone(sources), sup...)); err != nil {
			return err
		}
		p := plan{Schema: "riido-public-behavior-execution-plan-57-v1", SourceCommit: *source, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, BinarySHA256: binary.SHA256, BinaryBytes: binary.Bytes, ProbesSHA256: sha(rawProbes), ProbesBytes: len(rawProbes), Sources: sources, Support: sup, OfficialAttempts: 1, CPUThreads: 1, HeapSoftLimit: 256 << 20}
		raw, err := publicbehavior.Canonical(p)
		if err != nil {
			return err
		}
		if err := writeNew(*out, raw); err != nil {
			return err
		}
		fmt.Printf("plan prepared; source artifacts%d; candidate observations0\n", len(sources)+len(sup))
		return nil
	}
	if *source != "" || !hash(*input, 40) {
		return errors.New("publicaudit_input_freeze_required")
	}
	rawPlan, err := bounded(root, planPath)
	if err != nil {
		return err
	}
	var p plan
	if err := decodeCanonical(rawPlan, &p); err != nil {
		return err
	}
	if p.Schema != "riido-public-behavior-execution-plan-57-v1" || p.GoVersion != runtime.Version() || p.GOOS != runtime.GOOS || p.GOARCH != runtime.GOARCH || p.BinarySHA256 != binary.SHA256 || p.BinaryBytes != binary.Bytes || p.ProbesSHA256 != sha(rawProbes) || p.ProbesBytes != len(rawProbes) || p.OfficialAttempts != 1 || p.Retries != 0 || p.CPUThreads != 1 || p.HeapSoftLimit != 256<<20 || !slices.Equal(p.Sources, sources) || !slices.Equal(p.Support, sup) {
		return errors.New("publicaudit_execution_plan_mismatch")
	}
	all := append(slices.Clone(sources), sup...)
	if err := verifyBlobs(*repo, root, p.SourceCommit, all); err != nil {
		return err
	}
	inputs := []artifact{{Path: probesPath, Bytes: len(rawProbes), SHA256: sha(rawProbes)}, {Path: planPath, Bytes: len(rawPlan), SHA256: sha(rawPlan)}}
	if err := verifyBlobs(*repo, root, *input, inputs); err != nil {
		return err
	}
	if err := os.Mkdir(*out, 0700); err != nil {
		return errors.New("publicaudit_output_directory_exists_or_unavailable")
	}
	result, auditErr := publicbehavior.Audit(probes)
	r := record{Schema: "riido-public-behavior-official-record-57-v1", SourceCommit: p.SourceCommit, InputCommit: *input, ProbesSHA256: sha(rawProbes), PlanSHA256: sha(rawPlan), BinarySHA256: binary.SHA256, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, VerifiedGitBlobs: len(all) + len(inputs), Audit: result, Limitations: []string{"Finite development inputs only; six properties from two source families.", "Normative, descriptive API and stronger policy expectations remain distinct.", "No ranking, fits, model calls, weights, protected-final reads or savings evidence.", "Primary entry API calls exclude getters, String/Original and internal helpers; getter calls are separately recorded.", "Fresh output directory guards accidental reuse, not globally one-shot execution. External official attempt ledger:1; retry0.", "Heap limit and Go scheduling settings are not measured CPU, RSS, GPU or latency."}}
	raw, err := publicbehavior.Canonical(r)
	if err != nil {
		return errors.New("publicaudit_result_encoding_failed_after_calls")
	}
	if err := writeNew(*out+string(os.PathSeparator)+"results.json", raw); err != nil {
		return fmt.Errorf("publicaudit_result_write_failed_after_observations:%d", result.DispatchAttempts)
	}
	fmt.Printf("finite observations%d; primary API calls%d; matches%d; mismatches%d; unknown%d\n", result.CompletedObservations, result.PrimaryAPICalls, result.MatchedExpectations, result.MismatchedExpectations, result.UnknownObservations)
	if auditErr != nil {
		return auditErr
	}
	if result.UnknownObservations != 0 {
		return errors.New("publicaudit_unknown_observations")
	}
	return nil
}

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
