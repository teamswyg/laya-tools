// SPDX-License-Identifier: Apache-2.0
// riido-captionref prepares references to existing stored development records.
// It never executes candidate implementations, ranking, models or fitting.
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

	"github.com/teamswyg/laya-tools/internal/captionref"
	"github.com/teamswyg/laya-tools/internal/storedaudit"
)

//go:embed main.go
var compiledMain []byte

const (
	planPath     = "experiments/short-claim/execution-plan-59.json"
	maxBytes     = 1 << 20
	maxBinary    = 64 << 20
	planSchema   = "riido-captionref-execution-plan-59-v1"
	recipeSchema = "riido-captionref-build-recipe-59-v1"
)

type artifact = captionref.Artifact

func supportPaths() [9]string {
	return [9]string{
		"experiments/short-claim/PLAN-REFERENCES-59.ko.md",
		"experiments/short-claim/PLAN-REFERENCES-59.en.md",
		"experiments/short-claim/build-recipe-59.json",
		"experiments/short-claim/oracle-review-59.json",
		"cmd/riido-captionref/main_test.go",
		"internal/captionref/references_test.go",
		"go.mod", "go.sum", "LICENSE",
	}
}

func sourceDirs() [3]string {
	return [3]string{"cmd/riido-captionref", "internal/captionref", "internal/storedaudit"}
}

type plan struct {
	Schema              string     `json:"schema"`
	SourceCommit        string     `json:"source_freeze_commit"`
	GoVersion           string     `json:"go_version"`
	CGOEnabled          string     `json:"cgo_enabled"`
	TrimPath            bool       `json:"trimpath"`
	GOOS                string     `json:"goos"`
	GOARCH              string     `json:"goarch"`
	BinarySHA256        string     `json:"binary_sha256"`
	BinaryBytes         int        `json:"binary_bytes"`
	Inputs              []artifact `json:"input_artifacts"`
	Sources             []artifact `json:"compiled_source_artifacts"`
	Support             []artifact `json:"support_artifacts"`
	Parents             int        `json:"parents"`
	CandidateCaptions   int        `json:"candidate_captions"`
	Contracts           int        `json:"contracts"`
	Answerable          int        `json:"answerable_parents"`
	NoAnswer            int        `json:"no_answer_parents"`
	Unknown             int        `json:"unknown_parents"`
	ConnectedGroups     int        `json:"connected_groups"`
	LabeledGroups       int        `json:"labeled_connected_groups"`
	OfficialAttempts    int        `json:"planned_official_attempts"`
	Retries             int        `json:"planned_official_retries"`
	GitBlobBudget       int        `json:"git_blob_budget_seconds_per_file"`
	CPUThreads          int        `json:"go_max_procs"`
	HeapSoftLimit       int64      `json:"go_heap_soft_limit_bytes"`
	PreparationOnly     bool       `json:"preparation_only"`
	ContentReview       string     `json:"content_review"`
	LiteralReified      bool       `json:"literal_payload_reified"`
	TrainingReady       bool       `json:"training_ready"`
	NewLabels           int        `json:"new_labels_authorized"`
	NewParents          int        `json:"new_independent_parents_authorized"`
	Roles               int        `json:"roles_authorized"`
	Rankings            int        `json:"ranking_calls_authorized"`
	SourceCalls         int        `json:"source_api_calls_authorized"`
	ModelCalls          int        `json:"model_calls_authorized"`
	PaidCalls           int        `json:"paid_calls_authorized"`
	Fits                int        `json:"fits_authorized"`
	FinalReads          int        `json:"protected_final_reads_authorized"`
	Activation          bool       `json:"production_activation_authorized"`
	ResourceMeasurement bool       `json:"resource_measurement_authorized"`
}

type buildRecipe struct {
	Schema     string   `json:"schema"`
	GoVersion  string   `json:"go_version"`
	CGOEnabled string   `json:"cgo_enabled"`
	Arguments  []string `json:"arguments"`
}

type record struct {
	Schema           string              `json:"schema"`
	State            string              `json:"state"`
	FailureCode      string              `json:"failure_code"`
	FailureCause     string              `json:"failure_cause"`
	SourceCommit     string              `json:"source_freeze_commit"`
	InputCommit      string              `json:"input_freeze_commit"`
	Inputs           []artifact          `json:"input_artifacts"`
	Sources          []artifact          `json:"compiled_source_artifacts"`
	Support          []artifact          `json:"support_artifacts"`
	PlanSHA256       string              `json:"plan_sha256"`
	BinarySHA256     string              `json:"binary_sha256"`
	BinaryBytes      int                 `json:"binary_bytes"`
	GoVersion        string              `json:"go_version"`
	CGOEnabled       string              `json:"cgo_enabled"`
	TrimPath         bool                `json:"trimpath"`
	GOOS             string              `json:"goos"`
	GOARCH           string              `json:"goarch"`
	CPUThreads       int                 `json:"go_max_procs"`
	HeapSoftLimit    int64               `json:"go_heap_soft_limit_bytes"`
	VerifiedGitBlobs int                 `json:"git_blobs_verified_before_generation"`
	Counters         captionref.Counters `json:"generation_counters"`
	References       captionref.Report   `json:"references"`
	PreparationOnly  bool                `json:"preparation_only"`
	ContentReview    string              `json:"content_review"`
	LiteralReified   bool                `json:"literal_payload_reified"`
	TrainingReady    bool                `json:"training_ready"`
	Limitations      []string            `json:"limitations"`
}

func sha(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }

func hash(s string, n int) bool {
	_, err := hex.DecodeString(s)
	return len(s) == n && strings.ToLower(s) == s && err == nil
}

func safePath(path string) bool {
	return path != "" && !strings.ContainsAny(path, "\\:\x00") && !filepath.IsAbs(path) && fs.ValidPath(path) && path != "."
}

func bounded(root *os.Root, path string) ([]byte, error) {
	if !safePath(path) {
		return nil, errors.New("captionref_path_invalid")
	}
	f, err := root.Open(path)
	if err != nil {
		return nil, errors.New("captionref_file_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > maxBytes {
		return nil, errors.New("captionref_file_bounds")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil || len(raw) > maxBytes {
		return nil, errors.New("captionref_file_bounds")
	}
	return raw, nil
}

func canonical(value any) ([]byte, error) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil || len(raw)+1 > maxBytes {
		return nil, errors.New("captionref_json_encoding_or_bounds")
	}
	return append(raw, '\n'), nil
}

func decodeCanonical(raw []byte, value any) error {
	if len(raw) == 0 || len(raw) > maxBytes || !utf8.Valid(raw) {
		return errors.New("captionref_invalid_json")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return errors.New("captionref_invalid_json")
	}
	want, err := canonical(value)
	if err != nil || !bytes.Equal(want, raw) {
		return errors.New("captionref_noncanonical_json")
	}
	return nil
}

func sourceArtifacts() []artifact {
	out := captionref.CompiledFiles("internal/captionref")
	out = append(out, storedaudit.CompiledFiles("internal/storedaudit")...)
	return append(out, artifact{Path: "cmd/riido-captionref/main.go", SHA256: sha(compiledMain), Bytes: len(compiledMain)})
}

func support(root *os.Root) ([]artifact, error) {
	out := make([]artifact, 0, 9)
	for _, path := range supportPaths() {
		raw, err := bounded(root, path)
		if err != nil {
			return nil, err
		}
		if path == "experiments/short-claim/build-recipe-59.json" {
			if err := validateRecipe(raw); err != nil {
				return nil, err
			}
		}
		out = append(out, artifact{Path: path, SHA256: sha(raw), Bytes: len(raw)})
	}
	return out, nil
}

func validateRecipe(raw []byte) error {
	var got buildRecipe
	if err := decodeCanonical(raw, &got); err != nil {
		return err
	}
	want := buildRecipe{Schema: recipeSchema, GoVersion: "go1.27.1", CGOEnabled: "0", Arguments: []string{"build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "<private_binary>", "./cmd/riido-captionref"}}
	wantRaw, err := canonical(want)
	if err != nil || !bytes.Equal(raw, wantRaw) {
		return errors.New("captionref_build_recipe_mismatch")
	}
	return nil
}

func readInputs(root *os.Root) (captionref.Inputs, []artifact, error) {
	var inputs captionref.Inputs
	pins := captionref.InputPins()
	for i, pin := range pins {
		raw, err := bounded(root, pin.Path)
		if err != nil || len(raw) != pin.Bytes || sha(raw) != pin.SHA256 {
			return captionref.Inputs{}, nil, errors.New("captionref_input_pin_mismatch")
		}
		if i < 3 {
			inputs.Stored[i] = raw
		} else {
			inputs.Literals[i-3] = raw
		}
	}
	return inputs, slices.Clone(pins[:]), nil
}

func validateArtifacts(entries []artifact) error {
	seen := make([]string, 0, len(entries))
	for _, a := range entries {
		if !safePath(a.Path) || !hash(a.SHA256, 64) || a.Bytes < 0 || a.Bytes > maxBytes || slices.Contains(seen, a.Path) {
			return errors.New("captionref_pin_invalid")
		}
		seen = append(seen, a.Path)
	}
	return nil
}

func verifyBlobs(repo string, root *os.Root, commit string, entries []artifact) error {
	if !hash(commit, 40) {
		return errors.New("captionref_commit_invalid")
	}
	if err := validateArtifacts(entries); err != nil {
		return err
	}
	for _, a := range entries {
		raw, err := bounded(root, a.Path)
		if err != nil || len(raw) != a.Bytes || sha(raw) != a.SHA256 {
			return errors.New("captionref_disk_pin_mismatch")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "git", "show", commit+":"+a.Path)
		cmd.Dir = repo
		var output boundedBuffer
		cmd.Stdout, cmd.Stderr = &output, io.Discard
		err = cmd.Run()
		cancel()
		if err != nil || !bytes.Equal(output.raw, raw) {
			return errors.New("captionref_git_pin_mismatch")
		}
	}
	return nil
}

func verifyGoClosure(root *os.Root, sources []artifact) error {
	if len(sources) != 5 || validateArtifacts(sources) != nil {
		return errors.New("captionref_compiled_source_invalid")
	}
	for _, dir := range sourceDirs() {
		entries, err := fs.ReadDir(root.FS(), dir)
		if err != nil {
			return errors.New("captionref_source_directory_unavailable")
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
				return errors.New("captionref_unlisted_go_source")
			}
		}
	}
	return nil
}

type boundedBuffer struct{ raw []byte }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > maxBytes-len(b.raw) {
		return 0, errors.New("captionref_git_blob_bounds")
	}
	b.raw = append(b.raw, p...)
	return len(p), nil
}

func validBuild(info *debug.BuildInfo) bool {
	if info == nil || info.GoVersion != "go1.27.1" || info.Path != "github.com/teamswyg/laya-tools/cmd/riido-captionref" {
		return false
	}
	trim, cgo := false, false
	for _, s := range info.Settings {
		if s.Key == "-trimpath" {
			trim = s.Value == "true"
		}
		if s.Key == "CGO_ENABLED" {
			cgo = s.Value == "0"
		}
	}
	return trim && cgo
}

func binaryArtifact() (artifact, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || !validBuild(info) {
		return artifact{}, errors.New("captionref_binary_build_mismatch")
	}
	path, err := os.Executable()
	if err != nil {
		return artifact{}, errors.New("captionref_binary_unavailable")
	}
	f, err := os.Open(path)
	if err != nil {
		return artifact{}, errors.New("captionref_binary_unavailable")
	}
	defer f.Close()
	infoFile, err := f.Stat()
	if err != nil || !infoFile.Mode().IsRegular() || infoFile.Size() < 0 || infoFile.Size() > maxBinary {
		return artifact{}, errors.New("captionref_binary_bounds")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxBinary+1))
	if err != nil || n != infoFile.Size() || n > maxBinary {
		return artifact{}, errors.New("captionref_binary_unavailable")
	}
	return artifact{Bytes: int(n), SHA256: hex.EncodeToString(h.Sum(nil))}, nil
}

func openNew(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("captionref_output_exists_or_unavailable")
	}
	return f, nil
}

type writeCloser interface {
	Write([]byte) (int, error)
	Close() error
}

func finishNew(f writeCloser, raw []byte) error {
	if len(raw) > maxBytes {
		_ = f.Close()
		return errors.New("captionref_output_bounds")
	}
	n, err := f.Write(raw)
	cerr := f.Close()
	if err != nil || cerr != nil || n != len(raw) {
		return errors.New("captionref_output_write_failed")
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

func reserveResult(path string) (*os.File, error) {
	if filepath.Base(path) != "results.json" || filepath.Dir(path) == "." {
		return nil, errors.New("captionref_output_path_invalid")
	}
	if err := os.Mkdir(filepath.Dir(path), 0700); err != nil {
		return nil, errors.New("captionref_output_directory_exists_or_unavailable")
	}
	return openNew(path)
}

func failureCounters(kind string, c captionref.Counters) error {
	return fmt.Errorf("%s:metadata_bind_attempts=%d,metadata_bind_completed=%d,ast_parse_attempts=%d,ast_parses_completed=%d,parents_generated=%d,captions_generated=%d,contracts_generated=%d", kind, c.BindAttempts, c.BindCompleted, c.ASTParseAttempts, c.ASTParsesCompleted, c.ParentsGenerated, c.CaptionsGenerated, c.ContractsGenerated)
}

// Only fixed package error enums may enter the public failed record. Unexpected
// diagnostics never expose parser text, raw records, paths or arbitrary errors.
func generationFailureCause(err error) string {
	if err == nil {
		return ""
	}
	for _, known := range []error{
		captionref.ErrPin, captionref.ErrMetadata, captionref.ErrAST,
		captionref.ErrTarget, captionref.ErrSpan, captionref.ErrCounts,
		storedaudit.ErrPin, storedaudit.ErrJSON, storedaudit.ErrDuplicate,
		storedaudit.ErrUnknownField, storedaudit.ErrNumber, storedaudit.ErrUnicode,
		storedaudit.ErrSchema, storedaudit.ErrIdentity, storedaudit.ErrText,
		storedaudit.ErrTruth, storedaudit.ErrSource, storedaudit.ErrGroups,
		storedaudit.ErrCounts,
	} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	return "captionref_generation_error_unclassified"
}

func persistResult(f writeCloser, r record) error {
	raw, err := canonical(r)
	if err != nil {
		_ = f.Close()
		return failureCounters("captionref_result_encoding_failed_after_generation", r.Counters)
	}
	if err := finishNew(f, raw); err != nil {
		return failureCounters("captionref_result_write_failed_after_generation", r.Counters)
	}
	return nil
}

func newPlan(commit string, binary artifact, inputs, sources, sup []artifact) plan {
	return plan{Schema: planSchema, SourceCommit: commit, GoVersion: "go1.27.1", CGOEnabled: "0", TrimPath: true, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, BinarySHA256: binary.SHA256, BinaryBytes: binary.Bytes, Inputs: slices.Clone(inputs), Sources: slices.Clone(sources), Support: slices.Clone(sup), Parents: 72, CandidateCaptions: 216, Contracts: 18, Answerable: 34, NoAnswer: 17, Unknown: 21, ConnectedGroups: 17, LabeledGroups: 16, OfficialAttempts: 1, GitBlobBudget: 5, CPUThreads: 1, HeapSoftLimit: 256 << 20, PreparationOnly: true, ContentReview: "pending"}
}

func validatePlan(got, want plan) error {
	if !hash(got.SourceCommit, 40) || !hash(got.BinarySHA256, 64) || got.BinaryBytes <= 0 || got.BinaryBytes > maxBinary || len(got.Inputs) != 6 || len(got.Sources) != 5 || len(got.Support) != 9 {
		return errors.New("captionref_plan_invalid")
	}
	all := append(slices.Clone(got.Inputs), got.Sources...)
	all = append(all, got.Support...)
	if validateArtifacts(all) != nil {
		return errors.New("captionref_plan_invalid")
	}
	raw, err := canonical(got)
	wantRaw, wantErr := canonical(want)
	if err != nil || wantErr != nil || !bytes.Equal(raw, wantRaw) {
		return errors.New("captionref_plan_mismatch")
	}
	return nil
}

func run(args []string) error {
	if len(args) == 0 || (args[0] != "prepare" && args[0] != "references") {
		return errors.New("captionref_mode_required_prepare_or_references")
	}
	f := flag.NewFlagSet("riido-captionref", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	repo := f.String("repo", ".", "repository directory")
	source := f.String("source-commit", "", "full source freeze commit")
	input := f.String("input-commit", "", "full input freeze commit")
	out := f.String("out", "", "new plan file or new directory/results.json")
	if f.Parse(args[1:]) != nil || f.NArg() != 0 || *out == "" {
		return errors.New("captionref_arguments_invalid")
	}
	if (args[0] == "prepare" && (!hash(*source, 40) || *input != "")) || (args[0] == "references" && (!hash(*input, 40) || *source != "")) {
		return errors.New("captionref_commit_arguments_invalid")
	}
	if runtime.Version() != "go1.27.1" {
		return errors.New("captionref_toolchain_mismatch")
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	root, err := os.OpenRoot(*repo)
	if err != nil {
		return errors.New("captionref_repository_unavailable")
	}
	defer root.Close()
	inputs, pins, err := readInputs(root)
	if err != nil {
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
	sourceEntries := append(slices.Clone(sources), sup...)
	if args[0] == "prepare" {
		all := append(slices.Clone(sourceEntries), pins...)
		if err := verifyBlobs(*repo, root, *source, all); err != nil {
			return err
		}
		p := newPlan(*source, binary, pins, sources, sup)
		if err := validatePlan(p, p); err != nil {
			return err
		}
		raw, err := canonical(p)
		if err != nil {
			return err
		}
		if err := writeNew(*out, raw); err != nil {
			return err
		}
		fmt.Printf("plan prepared; Git blobs20; metadata Bind0; references0; ranking/sourceAPI/model/fit0\n")
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
	if err := validatePlan(p, newPlan(p.SourceCommit, binary, pins, sources, sup)); err != nil {
		return err
	}
	if err := verifyBlobs(*repo, root, p.SourceCommit, sourceEntries); err != nil {
		return err
	}
	inputEntries := append(slices.Clone(pins), artifact{Path: planPath, SHA256: sha(rawPlan), Bytes: len(rawPlan)})
	if err := verifyBlobs(*repo, root, *input, inputEntries); err != nil {
		return err
	}
	resultFile, err := reserveResult(*out)
	if err != nil {
		return err
	}
	r := record{Schema: "riido-caption-reference-preparation-record-59-v1", State: "incomplete", SourceCommit: p.SourceCommit, InputCommit: *input, Inputs: slices.Clone(pins), Sources: slices.Clone(sources), Support: slices.Clone(sup), PlanSHA256: sha(rawPlan), BinarySHA256: binary.SHA256, BinaryBytes: binary.Bytes, GoVersion: runtime.Version(), CGOEnabled: "0", TrimPath: true, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, CPUThreads: 1, HeapSoftLimit: 256 << 20, VerifiedGitBlobs: len(sourceEntries) + len(inputEntries), PreparationOnly: true, ContentReview: "pending", Limitations: []string{
		"Metadata references to previously observed public development records; no new truth or independent requests.",
		"Raw AST expression SHA is distinct from historical serialized literal-table, formatted code and bundle SHA.",
		"Literal Input/Want/Got values are not reified; source fidelity, request/contract coverage and observation-field reviews remain pending.",
		"No roles, rankings, source API calls, model or paid calls, fitting, weights, protected-final access, activation or savings evidence.",
		"Original no_roles_plan and synthetic_single_pipeline limitations remain unresolved.",
		"Fresh directory/exclusive output prevents accidental reuse; official attempt/retry policy is external, not a host-global lock.",
		"Go scheduling1 and heap soft limit256MiB are settings, not CPU, RSS, GPU or latency measurements.",
	}}
	r.References, err = captionref.Generate(inputs, &r.Counters)
	if err != nil {
		r.FailureCode = "captionref_generation_failed"
		r.FailureCause = generationFailureCause(err)
	} else {
		r.State = "references_generated_content_review_pending"
	}
	if err := persistResult(resultFile, r); err != nil {
		return err
	}
	if r.FailureCode != "" {
		return failureCounters(r.FailureCode+":"+r.FailureCause, r.Counters)
	}
	fmt.Printf("references prepared; parents%d captions%d contracts%d; content review pending; sourceAPI/ranking/model/fit0\n", r.Counters.ParentsGenerated, r.Counters.CaptionsGenerated, r.Counters.ContractsGenerated)
	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
