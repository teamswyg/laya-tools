// riido-residentperf prepares and measures a frozen public JSONL workload.
// It never loads a model, labels candidate truth or grants execution approval.
package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

//go:embed main.go types.go protocol.go prepare.go process.go artifacts.go
var compiledSources embed.FS

const (
	legacyPath          = "experiments/short-claim/probes-56.json"
	typedPath           = "experiments/short-claim/probes-56b.json"
	recipePath          = "experiments/short-claim/build-recipe-56d.json"
	maxCorpusBytes      = 4 << 20
	maxPlanBytes        = 256 << 10
	maxReportBytes      = 512 << 10
	maxPreparationBytes = 256 << 10
	maxRecipeBytes      = 64 << 10
)

func loadPayload(raw []byte) (shortclaim.Prepared, error) {
	p, e := shortclaim.Load(bytes.NewReader(raw))
	if e != nil {
		return shortclaim.Prepared{}, errors.New("corpus_payload_invalid")
	}
	return p, nil
}

func readBounded(path string, cap int64) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, errors.New("file_read_failed")
	}
	defer f.Close()
	return readRegular(f, cap)
}

func readRegular(f *os.File, cap int64) ([]byte, error) {
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > cap {
		return nil, errors.New("file_bounds_or_type_invalid")
	}
	raw, e := io.ReadAll(io.LimitReader(f, cap+1))
	if e != nil || int64(len(raw)) > cap {
		return nil, errors.New("file_read_limit_or_io")
	}
	return raw, nil
}

func decodePlan(raw []byte, expectedSHA string) (Plan, error) {
	var p Plan
	if len(raw) > maxPlanBytes || !validSHA(expectedSHA) || hashBytes(raw) != expectedSHA || strictJSON(raw, &p) != nil {
		return p, errors.New("plan_hash_or_shape_invalid")
	}
	// encoding/json silently truncates excess fixed-array elements. Require the
	// canonical complete public bytes so hidden rows or padded arrays cannot be
	// erased by decoding before policy checks.
	canonical, e := encodePublic(p, maxPlanBytes)
	if e != nil || !bytes.Equal(raw, canonical) {
		return p, errors.New("plan_canonical_shape_invalid")
	}
	if p.Schema != planSchema || p.State != planState || p.Policy != policy || p.Go != goVersion || runtime.Version() != goVersion || p.OS != runtime.GOOS || p.Arch != runtime.GOARCH || p.OS != "darwin" && p.OS != "linux" || p.FirstRequests != 1 || p.WarmupRequests != 20 || p.TimedRequests != timedPerRow || p.CPUThreads != 1 || p.HeapSoftLimitBytes != 256<<20 || p.ChildTimeoutSeconds != 15 || p.GlobalTimeoutSeconds != 60 || p.CleanupGraceMilliseconds != 1000 || p.Retries != 0 || p.InFlight != 1 || p.PublicResultLimitBytes != resultLimit || !p.StopOnFirstFailedRow || p.LegacyInputSHA256 != legacySHA || p.TypedInputSHA256 != typedSHA {
		return p, errors.New("plan_policy_invalid")
	}
	if !validCommit(p.SourceCommit) {
		return p, errors.New("plan_source_revision_invalid")
	}
	for _, sha := range []string{p.ChildBinarySHA256, p.DriverBinarySHA256, p.BuildRecipeSHA256, p.CorpusSHA256} {
		if !validSHA(sha) {
			return p, errors.New("plan_artifact_pin_invalid")
		}
	}
	if p.Rows != plannedRows() {
		return p, errors.New("plan_rows_invalid")
	}
	return p, nil
}

func plannedRows() [rowCount]RowPlan {
	var out [rowCount]RowPlan
	kinds := baselineKinds()
	i := 0
	for repeat := 1; repeat <= 3; repeat++ {
		for step := 0; step < 4; step++ {
			k := step
			if repeat == 2 {
				k = 3 - step
			}
			work := [2]string{"same", "distinct"}
			if (repeat+step)%2 == 0 {
				work = [2]string{"distinct", "same"}
			}
			for _, w := range work {
				out[i] = RowPlan{Baseline: kinds[k], Workload: w, Repeat: repeat}
				i++
			}
		}
	}
	return out
}

func compiledPins() ([]FilePin, error) {
	var pins []FilePin
	for _, name := range [6]string{"main.go", "types.go", "protocol.go", "prepare.go", "process.go", "artifacts.go"} {
		raw, e := compiledSources.ReadFile(name)
		if e != nil {
			return nil, errors.New("compiled_source_missing")
		}
		pins = append(pins, FilePin{Path: "cmd/riido-residentperf/" + name, SHA256: hashBytes(raw)})
	}
	for _, p := range behaviorprobe.AuditSourceArtifacts() {
		pins = append(pins, FilePin{Path: "internal/behaviorprobe/" + p.Name, SHA256: p.SHA256})
	}
	for _, p := range typedbehavior.SourceArtifacts() {
		pins = append(pins, FilePin{Path: "internal/typedbehavior/" + p.Name, SHA256: p.SHA256})
	}
	for _, p := range typedbehavior.PropertyObservationSourceArtifacts() {
		pins = append(pins, FilePin{Path: "internal/typedbehavior/" + p.Name, SHA256: p.SHA256})
	}
	for _, p := range lexicalhint.AuditSourceArtifacts() {
		pins = append(pins, FilePin{Path: "internal/lexicalhint/" + p.Name, SHA256: p.SHA256})
	}
	for _, p := range shortclaim.AuditSourceArtifacts() {
		pins = append(pins, FilePin{Path: "pkg/shortclaim/" + p.Name, SHA256: p.SHA256})
	}
	slices.SortFunc(pins, func(a, b FilePin) int { return strings.Compare(a.Path, b.Path) })
	return pins, nil
}

func bindCorpus(raw, legacyRaw, typedRaw []byte, p Plan) (ReadyCorpus, error) {
	var got Corpus
	if len(raw) > maxCorpusBytes || hashBytes(raw) != p.CorpusSHA256 || strictJSON(raw, &got) != nil || validateCorpus(got) != nil {
		return ReadyCorpus{}, errors.New("corpus_hash_or_shape_invalid")
	}
	want, e := prepareCorpus(legacyRaw, typedRaw)
	if e != nil {
		return ReadyCorpus{}, e
	}
	wantBytes, e := encodePublic(want, maxCorpusBytes)
	if e != nil || !bytes.Equal(raw, wantBytes) {
		return ReadyCorpus{}, errors.New("corpus_original_projection_mismatch")
	}
	return prepareWireTable(got)
}

func replayVerified(plan Plan, ready ReadyCorpus, verifiedPrivateChild, planSHA string) Report {
	r := Report{Schema: resultSchema, SourceCommit: plan.SourceCommit, PlanSHA256: planSHA, CorpusSHA256: plan.CorpusSHA256, ChildBinarySHA256: plan.ChildBinarySHA256, DriverBinarySHA256: plan.DriverBinarySHA256, OS: runtime.GOOS, Arch: runtime.GOARCH, Go: runtime.Version(), PlannedRows: rowCount, PlannedRequests: rowCount * requestsPerRow, UniqueFeaturePayloads: len(ready.Payloads), PureStartup: "unobserved", PureChildProcessing: "unobserved", PurePipe: "unobserved", ControllerRSSScope: "whole controller lifetime including preflight; CPU delta covers replay, row expectation hashes and cleanup; no per-row RSS attribution", Scope: "Original public repeated inputs, CPU-only uncached actual JSONL. Counts are repeated wire observations, not unique training/final requests. Validation is receipt-to-end including delivery gap; work and gap are separately observed. Per-row expectation hashing is before spawn but inside global replay CPU/wall. Protocol expectations are not semantic truth. No model/fit/savings assertion."}
	for i, row := range plan.Rows {
		r.Rows[i] = pendingRow(i, row)
	}
	before, beforeOK := selfUsage()
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(plan.GlobalTimeoutSeconds)*time.Second)
	defer cancel()
	stopReason := ""
	for i := range r.Rows {
		if stopReason != "" {
			r.Rows[i].Error = stopReason
			continue
		}
		if ctx.Err() != nil {
			stopReason = "global_deadline_before_start"
			r.Rows[i].Error = stopReason
			continue
		}
		r.Rows[i] = measureRow(ctx, verifiedPrivateChild, plan, ready, i)
		if r.Rows[i].Status != "complete" {
			stopReason = "prior_row_failure_no_retry"
		}
	}
	r.ReplayWallNS = time.Since(start).Nanoseconds()
	r.GlobalDeadlineExceeded = ctx.Err() != nil || r.ReplayWallNS > int64(plan.GlobalTimeoutSeconds)*int64(time.Second)
	after, afterOK := selfUsage()
	r.Controller = selfResource(before, after, beforeOK && afterOK)
	return r
}

func writeExact(f *os.File, raw []byte) error {
	n, e := f.Write(raw)
	if e != nil || n != len(raw) {
		return errors.New("output_write_failed")
	}
	if f.Close() != nil {
		return errors.New("output_close_failed")
	}
	return nil
}

func writeNew(path string, raw []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errors.New("new_output_create_failed")
	}
	defer f.Close()
	return writeExact(f, raw)
}

func encodePublic(v any, cap int) ([]byte, error) {
	raw, e := json.MarshalIndent(v, "", "  ")
	if e != nil || len(raw)+1 > cap {
		return nil, errors.New("public_result_encoding_or_limit_failed")
	}
	return append(raw, '\n'), nil
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("riido-residentperf", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	stage := fs.String("stage", "", "prepare or replay")
	planFile := fs.String("plan", "", "frozen machine plan")
	planSHA := fs.String("plan-sha256", "", "exact frozen plan SHA256")
	corpusFile := fs.String("corpus", "", "frozen public corpus")
	binaryFile := fs.String("binary", "", "actual packaged shortclaim child")
	dest := fs.String("out", "", "new output directory")
	if e := fs.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			fmt.Fprintln(stdout, "Usage: riido-residentperf --stage prepare --out NEW_DIRECTORY\nriido-residentperf --stage replay --plan PLAN --plan-sha256 SHA --corpus CORPUS --binary CHILD --out NEW_DIRECTORY\nRequires public repository sources and packaged Go1.27.1 CGO0 trimpath binaries. CPU-only repeated wire measurements; no models or semantic truth.")
			return flag.ErrHelp
		}
		return errors.New("invalid_arguments")
	}
	if fs.NArg() != 0 || *dest == "" || (*stage != "prepare" && *stage != "replay") {
		return errors.New("invalid_arguments")
	}
	if runtime.Version() != goVersion || (runtime.GOOS != "darwin" && runtime.GOOS != "linux") {
		return errors.New("unsupported_toolchain_or_os")
	}
	oldP := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldP)
	oldHeap := debug.SetMemoryLimit(256 << 20)
	defer debug.SetMemoryLimit(oldHeap)
	if *stage == "prepare" {
		if *planFile != "" || *planSHA != "" || *corpusFile != "" || *binaryFile != "" {
			return errors.New("prepare_arguments_conflict")
		}
		if e := verifySources(nil); e != nil {
			return e
		}
		legacy, e := readBounded(legacyPath, 1<<20)
		if e != nil {
			return e
		}
		typed, e := readBounded(typedPath, 1<<20)
		if e != nil {
			return e
		}
		c, e := prepareCorpus(legacy, typed)
		if e != nil {
			return e
		}
		raw, e := encodePublic(c, maxCorpusBytes)
		if e != nil {
			return e
		}
		pins, e := currentManifest()
		if e != nil {
			return e
		}
		recipe, e := recipeBytes()
		if e != nil {
			return e
		}
		summary := struct {
			Schema               string    `json:"schema"`
			State                string    `json:"state"`
			CorpusSHA            string    `json:"corpus_sha256"`
			Bytes                int       `json:"corpus_bytes"`
			OriginalParents      int       `json:"original_parents"`
			Payloads             int       `json:"distinct_feature_payloads"`
			ProtocolExpectations int       `json:"prepared_protocol_expectations"`
			Sources              []FilePin `json:"implementation_files"`
			BuildRecipeSHA       string    `json:"build_recipe_sha256"`
			ChildRequests        int       `json:"measured_child_requests"`
			Models               int       `json:"model_calls"`
			Fits                 int       `json:"fits"`
			Final                int       `json:"protected_final_reads"`
			Scope                string    `json:"scope"`
		}{"riido-resident-resource-preparation-v1", "protocol_only_no_measurement", hashBytes(raw), len(raw), 72, len(c.Payloads), len(c.Payloads) * 4, pins, hashBytes(recipe), 0, 0, 0, 0, "Prepare only public source projections and deterministic protocol expectations. Distinct normalized features are not independent labels or safe cache keys; no child resource measurements."}
		summaryRaw, e := encodePublic(summary, maxPreparationBytes)
		if e != nil {
			return e
		}
		if len(raw)+len(summaryRaw)+len(recipe) > resultLimit {
			return errors.New("public_total_result_budget_exceeded")
		}
		if os.Mkdir(*dest, 0700) != nil {
			return errors.New("output_directory_not_new")
		}
		if e := writeNew(filepath.Join(*dest, "corpus.json"), raw); e != nil {
			return e
		}
		if e := writeNew(filepath.Join(*dest, "preparation.json"), summaryRaw); e != nil {
			return e
		}
		if e := writeNew(filepath.Join(*dest, "build-recipe.json"), recipe); e != nil {
			return e
		}
		if _, e := stdout.Write(summaryRaw); e != nil {
			return errors.New("summary_write_failed")
		}
		return nil
	}
	if *planFile == "" || *corpusFile == "" || *binaryFile == "" || !validSHA(*planSHA) {
		return errors.New("replay_arguments_incomplete")
	}
	planRaw, e := readBounded(*planFile, maxPlanBytes)
	if e != nil {
		return e
	}
	p, e := decodePlan(planRaw, *planSHA)
	if e != nil {
		return e
	}
	if e := verifySources(&p); e != nil {
		return e
	}
	corpusRaw, e := readBounded(*corpusFile, maxCorpusBytes)
	if e != nil {
		return e
	}
	legacy, e := readBounded(legacyPath, 1<<20)
	if e != nil {
		return e
	}
	typed, e := readBounded(typedPath, 1<<20)
	if e != nil {
		return e
	}
	ready, e := bindCorpus(corpusRaw, legacy, typed, p)
	if e != nil {
		return e
	}
	child, e := readBounded(*binaryFile, 32<<20)
	if e != nil {
		return e
	}
	if e := verifyBinary(child, p.ChildBinarySHA256, "github.com/teamswyg/laya-tools/cmd/riido-shortclaim"); e != nil {
		return e
	}
	selfPath, e := os.Executable()
	if e != nil {
		return errors.New("driver_identity_unavailable")
	}
	self, e := readBounded(selfPath, 32<<20)
	if e != nil {
		return e
	}
	if e := verifyBinary(self, p.DriverBinarySHA256, "github.com/teamswyg/laya-tools/cmd/riido-residentperf"); e != nil {
		return e
	}
	if len(corpusRaw)+len(planRaw)+maxReportBytes+maxPreparationBytes+maxRecipeBytes > resultLimit {
		return errors.New("public_total_result_budget_exceeded")
	}
	privateDir, e := os.MkdirTemp("", "riido-resident-child-")
	if e != nil {
		return errors.New("binary_copy_directory_failed")
	}
	defer os.RemoveAll(privateDir)
	childPath := filepath.Join(privateDir, "child")
	if os.WriteFile(childPath, child, 0500) != nil {
		return errors.New("binary_copy_failed")
	}
	if os.Mkdir(*dest, 0700) != nil {
		return errors.New("output_directory_not_new")
	}
	// Reserve every result before the first process starts, preventing an
	// existing-output error from causing a post-measurement automatic rerun.
	f, e := os.OpenFile(filepath.Join(*dest, "results.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errors.New("new_output_create_failed")
	}
	defer f.Close()
	r := replayVerified(p, ready, childPath, *planSHA)
	raw, e := encodePublic(r, maxReportBytes)
	if e != nil {
		return e
	}
	if e := writeExact(f, raw); e != nil {
		return e
	}
	completed := 0
	for _, row := range r.Rows {
		if row.Status == "complete" {
			completed++
		}
	}
	fmt.Fprintf(stdout, "resident replay recorded: %d/%d complete rows; public raw result retained\n", completed, rowCount)
	if completed != rowCount || r.GlobalDeadlineExceeded {
		return errors.New("resident_replay_incomplete_recorded")
	}
	return nil
}

func main() {
	if e := run(os.Args[1:], os.Stdout); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, e.Error())
		os.Exit(1)
	}
}
