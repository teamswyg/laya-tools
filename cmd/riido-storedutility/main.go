// riido-storedutility ranks existing public captions against frozen stored
// truth. Preparation only binds metadata; audit never re-executes source truth.
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
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/storedaudit"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

//go:embed main.go
var compiledMain []byte

//go:embed evaluator.go
var compiledEvaluator []byte

const (
	planPath     = "experiments/short-claim/execution-plan-58.json"
	maxBytes     = 1 << 20
	planSchema   = "riido-storedutility-execution-plan-58-v1"
	recipeSchema = "riido-storedutility-build-recipe-58-v1"
)

type artifact = storedaudit.FilePin

func rawPaths() [3]string {
	return [3]string{
		"experiments/short-claim/probes-56.json",
		"experiments/short-claim/probes-56b.json",
		"experiments/short-claim/results-56b.json",
	}
}

func supportPaths() [13]string {
	return [13]string{
		"experiments/short-claim/PLAN-STORED-58.ko.md",
		"experiments/short-claim/PLAN-STORED-58.en.md",
		"experiments/short-claim/build-recipe-58.json",
		"experiments/short-claim/oracle-review-58.json",
		"cmd/riido-storedutility/main_test.go",
		"cmd/riido-storedutility/evaluator_test.go",
		"internal/storedaudit/binding_test.go",
		"pkg/shortclaim/input_test.go",
		"pkg/shortclaim/baseline_test.go",
		"internal/lexicalhint/features_test.go",
		"go.mod", "go.sum", "LICENSE",
	}
}

func sourceDirs() [4]string {
	return [4]string{"cmd/riido-storedutility", "internal/storedaudit", "pkg/shortclaim", "internal/lexicalhint"}
}

func baselineKinds() [4]string {
	return [4]string{shortclaim.FixedOrderKind, shortclaim.BM25Kind, shortclaim.LexicalOrderedKind, shortclaim.NarrowRuleKind}
}

type plan struct {
	Schema               string     `json:"schema"`
	SourceCommit         string     `json:"source_freeze_commit"`
	GoVersion            string     `json:"go_version"`
	GOOS                 string     `json:"goos"`
	GOARCH               string     `json:"goarch"`
	BinarySHA256         string     `json:"binary_sha256"`
	BinaryBytes          int        `json:"binary_bytes"`
	Inputs               []artifact `json:"stored_input_artifacts"`
	Sources              []artifact `json:"compiled_source_artifacts"`
	Support              []artifact `json:"support_artifacts"`
	Parents              int        `json:"parents"`
	CandidateCaptions    int        `json:"candidate_captions"`
	Answerable           int        `json:"answerable_parents"`
	NoAnswer             int        `json:"no_answer_parents"`
	Unknown              int        `json:"unknown_parents"`
	ConnectedGroups      int        `json:"connected_groups"`
	LabeledGroups        int        `json:"labeled_connected_groups"`
	Baselines            []string   `json:"baselines"`
	RequiredRelativeGain float64    `json:"necessary_utility_reduction"`
	OfficialAttempts     int        `json:"planned_official_attempts"`
	Retries              int        `json:"planned_official_retries"`
	RankingBudgetSeconds int        `json:"ranking_budget_seconds"`
	GitBlobBudgetSeconds int        `json:"git_blob_budget_seconds_per_file"`
	CPUThreads           int        `json:"go_max_procs"`
	HeapSoftLimit        int64      `json:"go_heap_soft_limit_bytes"`
	Roles                int        `json:"roles_assigned"`
	Fits                 int        `json:"fits_authorized"`
	ModelCalls           int        `json:"model_calls_authorized"`
	PaidCalls            int        `json:"paid_calls_authorized"`
	FinalReads           int        `json:"protected_final_reads_authorized"`
	Weights              int        `json:"new_weights_authorized"`
	Activation           bool       `json:"production_activation_authorized"`
	ResourceMeasurement  bool       `json:"resource_measurement_authorized"`
}

type buildRecipe struct {
	Schema     string   `json:"schema"`
	GoVersion  string   `json:"go_version"`
	CGOEnabled string   `json:"cgo_enabled"`
	Arguments  []string `json:"arguments"`
}

type record struct {
	Schema           string     `json:"schema"`
	SourceCommit     string     `json:"source_freeze_commit"`
	InputCommit      string     `json:"input_freeze_commit"`
	Inputs           []artifact `json:"stored_input_artifacts"`
	PlanSHA256       string     `json:"plan_sha256"`
	BinarySHA256     string     `json:"binary_sha256"`
	BinaryBytes      int        `json:"binary_bytes"`
	GoVersion        string     `json:"go_version"`
	GOOS             string     `json:"goos"`
	GOARCH           string     `json:"goarch"`
	VerifiedGitBlobs int        `json:"git_blobs_verified_before_ranking"`
	Evaluation       report     `json:"evaluation"`
	Limitations      []string   `json:"limitations"`
}

func sha(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }

func hash(s string, n int) bool {
	_, err := hex.DecodeString(s)
	return len(s) == n && strings.ToLower(s) == s && err == nil
}

func safePath(path string) bool {
	if path == "" || strings.ContainsAny(path, "\\:\x00") || filepath.IsAbs(path) || !fs.ValidPath(path) {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "." || part == ".." || part == "" {
			return false
		}
	}
	return true
}

func bounded(root *os.Root, path string) ([]byte, error) {
	if !safePath(path) {
		return nil, errors.New("storedutility_path_invalid")
	}
	f, err := root.Open(path)
	if err != nil {
		return nil, errors.New("storedutility_file_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxBytes {
		return nil, errors.New("storedutility_file_bounds")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(raw) > maxBytes {
		return nil, errors.New("storedutility_file_bounds")
	}
	return raw, nil
}

func canonical(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil || len(raw)+1 > maxBytes {
		return nil, errors.New("storedutility_json_encoding_or_bounds")
	}
	return append(raw, '\n'), nil
}

func decodeCanonical(raw []byte, value any) error {
	if len(raw) == 0 || len(raw) > maxBytes || !utf8.Valid(raw) {
		return errors.New("storedutility_invalid_json")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return errors.New("storedutility_invalid_json")
	}
	canon, err := canonical(value)
	if err != nil || !bytes.Equal(canon, raw) {
		return errors.New("storedutility_noncanonical_json")
	}
	return nil
}

func sourceArtifacts(root *os.Root) ([]artifact, error) {
	out := storedaudit.CompiledFiles("internal/storedaudit")
	out = append(out, artifact{Path: "cmd/riido-storedutility/main.go", Bytes: len(compiledMain), SHA256: sha(compiledMain)}, artifact{Path: "cmd/riido-storedutility/evaluator.go", Bytes: len(compiledEvaluator), SHA256: sha(compiledEvaluator)})
	for _, a := range shortclaim.AuditSourceArtifacts() {
		raw, err := bounded(root, "pkg/shortclaim/"+a.Name)
		if err != nil || sha(raw) != a.SHA256 {
			return nil, errors.New("storedutility_compiled_source_mismatch")
		}
		out = append(out, artifact{Path: "pkg/shortclaim/" + a.Name, SHA256: a.SHA256, Bytes: len(raw)})
	}
	for _, a := range lexicalhint.AuditSourceArtifacts() {
		raw, err := bounded(root, "internal/lexicalhint/"+a.Name)
		if err != nil || sha(raw) != a.SHA256 {
			return nil, errors.New("storedutility_compiled_source_mismatch")
		}
		out = append(out, artifact{Path: "internal/lexicalhint/" + a.Name, SHA256: a.SHA256, Bytes: len(raw)})
	}
	slices.SortFunc(out, func(a, b artifact) int { return strings.Compare(a.Path, b.Path) })
	if len(out) != 9 {
		return nil, errors.New("storedutility_source_closure_mismatch")
	}
	for _, a := range out {
		raw, err := bounded(root, a.Path)
		if err != nil || len(raw) != a.Bytes || sha(raw) != a.SHA256 {
			return nil, errors.New("storedutility_compiled_source_mismatch")
		}
	}
	return out, nil
}

func support(root *os.Root) ([]artifact, error) {
	paths := supportPaths()
	out := make([]artifact, len(paths))
	for i, path := range paths {
		raw, err := bounded(root, path)
		if err != nil {
			return nil, err
		}
		out[i] = artifact{Path: path, Bytes: len(raw), SHA256: sha(raw)}
		if strings.HasSuffix(path, "/build-recipe-58.json") {
			if err := validateRecipe(raw); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func validateRecipe(raw []byte) error {
	var r buildRecipe
	if err := decodeCanonical(raw, &r); err != nil {
		return err
	}
	if r.Schema != recipeSchema || r.GoVersion != "go1.27.1" || r.CGOEnabled != "0" || !slices.Equal(r.Arguments, []string{"build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "<private_binary>", "./cmd/riido-storedutility"}) {
		return errors.New("storedutility_build_recipe_mismatch")
	}
	return nil
}

func rawInputs(root *os.Root) ([3][]byte, []artifact, error) {
	var raw [3][]byte
	paths := rawPaths()
	want := [3]string{storedaudit.LegacySHA256, storedaudit.TypedSHA256, storedaudit.ResultSHA256}
	pins := make([]artifact, len(paths))
	for i, path := range paths {
		b, err := bounded(root, path)
		if err != nil || sha(b) != want[i] {
			return raw, nil, errors.New("storedutility_stored_input_pin_mismatch")
		}
		raw[i] = b
		pins[i] = artifact{Path: path, Bytes: len(b), SHA256: want[i]}
	}
	return raw, pins, nil
}

func verifyBlobs(repo string, root *os.Root, commit string, entries []artifact) error {
	if !hash(commit, 40) {
		return errors.New("storedutility_commit_invalid")
	}
	seen := make([]string, 0, len(entries))
	for _, a := range entries {
		if !safePath(a.Path) || !hash(a.SHA256, 64) || a.Bytes < 0 || a.Bytes > maxBytes || slices.Contains(seen, a.Path) {
			return errors.New("storedutility_pin_invalid")
		}
		seen = append(seen, a.Path)
		raw, err := bounded(root, a.Path)
		if err != nil || len(raw) != a.Bytes || sha(raw) != a.SHA256 {
			return errors.New("storedutility_disk_pin_mismatch")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "git", "show", commit+":"+a.Path)
		cmd.Dir = repo
		var output boundedBuffer
		cmd.Stdout, cmd.Stderr = &output, io.Discard
		err = cmd.Run()
		cancel()
		if err != nil || !bytes.Equal(output.raw, raw) {
			return errors.New("storedutility_git_pin_mismatch")
		}
	}
	return nil
}

func verifyGoClosure(root *os.Root, sources []artifact) error {
	for _, dir := range sourceDirs() {
		entries, err := fs.ReadDir(root.FS(), dir)
		if err != nil {
			return errors.New("storedutility_source_directory_unavailable")
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
				return errors.New("storedutility_unlisted_go_source")
			}
		}
	}
	return nil
}

type boundedBuffer struct{ raw []byte }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > maxBytes-len(b.raw) {
		return 0, errors.New("storedutility_git_blob_bounds")
	}
	b.raw = append(b.raw, p...)
	return len(p), nil
}

func validBuild(info *debug.BuildInfo) bool {
	if info == nil || info.GoVersion != "go1.27.1" || info.Path != "github.com/teamswyg/laya-tools/cmd/riido-storedutility" {
		return false
	}
	trim, cgo := false, false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "-trimpath":
			trim = setting.Value == "true"
		case "CGO_ENABLED":
			cgo = setting.Value == "0"
		}
	}
	return trim && cgo
}

func binaryArtifact() (artifact, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || !validBuild(info) {
		return artifact{}, errors.New("storedutility_binary_build_mismatch")
	}
	path, err := os.Executable()
	if err != nil {
		return artifact{}, errors.New("storedutility_binary_unavailable")
	}
	f, err := os.Open(path)
	if err != nil {
		return artifact{}, errors.New("storedutility_binary_unavailable")
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() < 0 || stat.Size() > 64<<20 {
		return artifact{}, errors.New("storedutility_binary_bounds")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, (64<<20)+1))
	if err != nil || n != stat.Size() || n > 64<<20 {
		return artifact{}, errors.New("storedutility_binary_unavailable")
	}
	return artifact{Bytes: int(n), SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func openNew(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("storedutility_output_exists_or_unavailable")
	}
	return f, nil
}

func finishNew(f *os.File, raw []byte) error {
	if len(raw) > maxBytes {
		_ = f.Close()
		return errors.New("storedutility_output_bounds")
	}
	n, err := f.Write(raw)
	cerr := f.Close()
	if err != nil || cerr != nil || n != len(raw) {
		return errors.New("storedutility_output_write_failed")
	}
	return nil
}

func writeNew(path string, raw []byte) error {
	f, err := openNew(path)
	if err != nil {
		return err
	}
	return finishNew(f, raw)
}

func resultFailure(encoding bool, r report) error {
	kind := "storedutility_result_write_failed_after_evaluation"
	if encoding {
		kind = "storedutility_result_encoding_failed_after_evaluation"
	}
	return fmt.Errorf("%s:dispatch_attempts=%d,validated_requests=%d,baseline_calls=%d,ranking_outputs=%d,completed_rows=%d", kind, r.DispatchAttempts, r.ValidatedRequests, r.BaselineCalls, r.RankingOutputs, r.CompletedRows)
}

func newPlan(commit string, binary artifact, inputs, sources, support []artifact) plan {
	kinds := baselineKinds()
	return plan{Schema: planSchema, SourceCommit: commit, GoVersion: "go1.27.1", GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, BinarySHA256: binary.SHA256, BinaryBytes: binary.Bytes, Inputs: slices.Clone(inputs), Sources: slices.Clone(sources), Support: slices.Clone(support), Parents: 72, CandidateCaptions: 216, Answerable: 34, NoAnswer: 17, Unknown: 21, ConnectedGroups: 17, LabeledGroups: 16, Baselines: slices.Clone(kinds[:]), RequiredRelativeGain: .05, OfficialAttempts: 1, RankingBudgetSeconds: 10, GitBlobBudgetSeconds: 5, CPUThreads: 1, HeapSoftLimit: 256 << 20}
}

func validatePlan(p, expected plan) error {
	if !hash(p.SourceCommit, 40) || !hash(p.BinarySHA256, 64) || p.BinaryBytes <= 0 || p.BinaryBytes > 64<<20 {
		return errors.New("storedutility_execution_plan_mismatch")
	}
	raw, err := canonical(p)
	want, wantErr := canonical(expected)
	if err != nil || wantErr != nil || !bytes.Equal(raw, want) {
		return errors.New("storedutility_execution_plan_mismatch")
	}
	return nil
}

func execute(args []string) error {
	if len(args) == 0 || (args[0] != "plan" && args[0] != "audit") {
		return errors.New("usage: riido-storedutility plan|audit --repo PATH --out NEW_PATH; plan --source-commit SHA, audit --input-commit SHA")
	}
	flags := flag.NewFlagSet("riido-storedutility", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	repo := flags.String("repo", ".", "public repository checkout")
	out := flags.String("out", "", "new plan file or new audit directory")
	source := flags.String("source-commit", "", "source freeze for plan")
	input := flags.String("input-commit", "", "input freeze for audit")
	if flags.Parse(args[1:]) != nil || flags.NArg() != 0 || *out == "" {
		return errors.New("storedutility_arguments")
	}
	if args[0] == "plan" && (*input != "" || !hash(*source, 40)) {
		return errors.New("storedutility_plan_arguments")
	}
	if args[0] == "audit" && (*source != "" || !hash(*input, 40)) {
		return errors.New("storedutility_input_freeze_required")
	}
	if runtime.Version() != "go1.27.1" {
		return errors.New("storedutility_toolchain_mismatch")
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	root, err := os.OpenRoot(*repo)
	if err != nil {
		return errors.New("storedutility_repository_unavailable")
	}
	defer root.Close()
	raw, inputs, err := rawInputs(root)
	if err != nil {
		return err
	}
	sources, err := sourceArtifacts(root)
	if err != nil {
		return err
	}
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
	all := append(slices.Clone(sources), sup...)
	if args[0] == "plan" {
		if err := verifyBlobs(*repo, root, *source, all); err != nil {
			return err
		}
		if err := verifyBlobs(*repo, root, *source, inputs); err != nil {
			return err
		}
		if _, err := storedaudit.Bind(raw[0], raw[1], raw[2]); err != nil {
			return errors.New("storedutility_metadata_preflight_failed")
		}
		p := newPlan(*source, binary, inputs, sources, sup)
		encoded, err := canonical(p)
		if err != nil {
			return err
		}
		if err := writeNew(*out, encoded); err != nil {
			return err
		}
		fmt.Printf("plan prepared; source artifacts %d; ranking calls 0; original API calls 0\n", len(all))
		return nil
	}
	rawPlan, err := bounded(root, planPath)
	if err != nil {
		return err
	}
	var p plan
	if err := decodeCanonical(rawPlan, &p); err != nil {
		return err
	}
	if err := validatePlan(p, newPlan(p.SourceCommit, binary, inputs, sources, sup)); err != nil {
		return err
	}
	if err := verifyBlobs(*repo, root, p.SourceCommit, all); err != nil {
		return err
	}
	inputEntries := append(slices.Clone(inputs), artifact{Path: planPath, Bytes: len(rawPlan), SHA256: sha(rawPlan)})
	if err := verifyBlobs(*repo, root, *input, inputEntries); err != nil {
		return err
	}
	d, err := storedaudit.Bind(raw[0], raw[1], raw[2])
	if err != nil {
		return errors.New("storedutility_metadata_preflight_failed")
	}
	if err := os.Mkdir(*out, 0700); err != nil {
		return errors.New("storedutility_output_directory_exists_or_unavailable")
	}
	resultFile, err := openNew(filepath.Join(*out, "results.json"))
	if err != nil {
		return err
	}
	// Reserve the result before the first ranking. A failure leaves its partial
	// evidence in the exclusive directory; it never silently retries.
	result, evaluationErr := evaluate(d)
	r := record{Schema: "riido-storedutility-official-record-58-v1", SourceCommit: p.SourceCommit, InputCommit: *input, Inputs: slices.Clone(inputs), PlanSHA256: sha(rawPlan), BinarySHA256: binary.SHA256, BinaryBytes: binary.Bytes, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, VerifiedGitBlobs: len(all) + len(inputEntries), Evaluation: result, Limitations: []string{
		"Stored development truth only; no original implementation/API calls or fresh truth labels.",
		"Four nonlearned rankings over 72 existing requests; repetitions and assertions are not independent final requests.",
		"No roles, fitting, weights, model or paid calls, protected-final reads, activation or savings evidence.",
		"Necessary 5% utility headroom is not training readiness, calibrated confidence or measured LLM savings.",
		"Fresh directory and exclusive result guard accidental reuse; official attempt1/retry0 is an external ledger policy, not a host-global lock.",
		"Ranking uses a cooperative 10-second context at bounded row boundaries; each local Git blob verification separately permits 5 seconds.",
		"Go scheduling and heap soft limit are settings, not CPU, RSS, GPU or latency measurements.",
	}}
	encoded, encodeErr := canonical(r)
	if encodeErr != nil {
		_ = resultFile.Close()
		return resultFailure(true, result)
	}
	if err := finishNew(resultFile, encoded); err != nil {
		return resultFailure(false, result)
	}
	fmt.Printf("completed rows %d; baseline calls %d; known requests %d; best baseline %s; possible reduction %.6f; necessary gate %t\n", result.CompletedRows, result.BaselineCalls, result.KnownRequests, result.BestBaseline, result.PossibleRelativeGain, result.UtilityUpperBoundPass)
	if evaluationErr != nil {
		return errors.New("storedutility_evaluation_incomplete")
	}
	return nil
}

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
