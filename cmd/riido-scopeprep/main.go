// riido-scopeprep emits scoped-property proposals without assigning truth.
// It is an offline maintainer tool, not a ranking or inference service.
package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/scopedproperty"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

//go:embed main.go
var compiledMain string

const (
	sourceInputPath   = "experiments/short-claim/probes-56b.json"
	sourceInputSHA256 = "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"
	maxSourceBytes    = 1 << 20
	maxPreparedBytes  = 1 << 20
)

var requiredPaths = [...]string{
	"cmd/riido-scopeprep/main.go",
	"internal/behaviorprobe/audit_provenance.go",
	"internal/behaviorprobe/data.go",
	"internal/lexicalhint/audit_provenance.go",
	"internal/lexicalhint/features.go",
	"internal/scopedproperty/dataset.go",
	"internal/scopedproperty/spec.go",
	"internal/typedbehavior/audit.go",
	"internal/typedbehavior/dataset.go",
	"internal/typedbehavior/fixtures.go",
	"internal/typedbehavior/flow.go",
	"internal/typedbehavior/groups.go",
	"internal/typedbehavior/property_observation.go",
	"internal/typedbehavior/source.go",
	"internal/typedbehavior/spec.go",
	"internal/typedbehavior/state.go",
	"pkg/shortclaim/audit_provenance.go",
	"pkg/shortclaim/baseline.go",
	"pkg/shortclaim/input.go",
}

type artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type preparation struct {
	Schema                  string     `json:"schema"`
	State                   string     `json:"state"`
	GoVersion               string     `json:"go_version"`
	CPUThreads              int        `json:"go_max_procs"`
	GoHeapSoftLimitBytes    int64      `json:"go_heap_soft_limit_bytes"`
	SourceInputSHA256       string     `json:"source_input_sha256"`
	ProposalsSHA256         string     `json:"proposal_bytes_sha256"`
	ProposalsBytes          int        `json:"proposal_bytes"`
	ProposalRows            int        `json:"proposal_rows"`
	CandidateDescriptions   int        `json:"candidate_descriptions"`
	Properties              int        `json:"properties"`
	LiteralDefinitions      int        `json:"literal_definitions"`
	CaptionReview           string     `json:"caption_review"`
	CandidateEvaluations    int        `json:"candidate_evaluations"`
	TruthLabelsAssigned     int        `json:"truth_labels_assigned"`
	GroupsAssigned          int        `json:"groups_assigned"`
	RolesAssigned           int        `json:"roles_assigned"`
	Fits                    int        `json:"fits"`
	ModelCalls              int        `json:"model_calls"`
	RankingRuns             int        `json:"ranking_runs"`
	PerformanceRuns         int        `json:"performance_runs"`
	ProtectedFinalRead      bool       `json:"protected_final_read"`
	TrainingReady           bool       `json:"training_ready"`
	OfficialExecutionFrozen bool       `json:"official_execution_frozen"`
	CompiledSources         []artifact `json:"compiled_sources"`
	StopReasons             []string   `json:"stop_reasons"`
	Limitations             []string   `json:"limitations"`
}

func digest(raw []byte) string {
	s := sha256.Sum256(raw)
	return hex.EncodeToString(s[:])
}

func compiledArtifacts() ([]artifact, error) {
	out := []artifact{{"cmd/riido-scopeprep/main.go", digest([]byte(compiledMain))}}
	add := func(prefix string, entries []typedbehavior.SourceArtifact) {
		for _, e := range entries {
			out = append(out, artifact{prefix + e.Name, e.SHA256})
		}
	}
	add("internal/typedbehavior/", typedbehavior.SourceArtifacts())
	add("internal/typedbehavior/", typedbehavior.PropertyObservationSourceArtifacts())
	add("internal/scopedproperty/", scopedproperty.SourceArtifacts())
	for _, e := range behaviorprobe.AuditSourceArtifacts() {
		out = append(out, artifact{"internal/behaviorprobe/" + e.Name, e.SHA256})
	}
	for _, e := range lexicalhint.AuditSourceArtifacts() {
		out = append(out, artifact{"internal/lexicalhint/" + e.Name, e.SHA256})
	}
	for _, e := range shortclaim.AuditSourceArtifacts() {
		out = append(out, artifact{"pkg/shortclaim/" + e.Name, e.SHA256})
	}
	slices.SortFunc(out, func(a, b artifact) int { return strings.Compare(a.Path, b.Path) })
	if len(out) != len(requiredPaths) {
		return nil, errors.New("scopeprep_compiled_manifest_invalid")
	}
	for i, e := range out {
		if e.Path != requiredPaths[i] || len(e.SHA256) != 64 {
			return nil, errors.New("scopeprep_compiled_manifest_invalid")
		}
		if _, err := hex.DecodeString(e.SHA256); err != nil {
			return nil, errors.New("scopeprep_compiled_manifest_invalid")
		}
	}
	return out, nil
}

func boundedSource(root *os.Root, path string) ([]byte, error) {
	f, err := root.Open(path)
	if err != nil {
		return nil, errors.New("scopeprep_source_unavailable")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxSourceBytes {
		return nil, errors.New("scopeprep_source_invalid")
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxSourceBytes+1))
	if err != nil || len(raw) > maxSourceBytes {
		return nil, errors.New("scopeprep_source_invalid")
	}
	return raw, nil
}

func verifySources(root *os.Root, sources []artifact) error {
	for _, a := range sources {
		raw, err := boundedSource(root, a.Path)
		if err != nil {
			return err
		}
		if digest(raw) != a.SHA256 {
			return errors.New("scopeprep_stale_compiled_source")
		}
	}
	// Additional package files can introduce initialization effects. Refuse
	// unlisted non-test Go files even if the selected candidate sources match.
	for _, dir := range []string{"cmd/riido-scopeprep", "internal/behaviorprobe", "internal/lexicalhint", "internal/scopedproperty", "internal/typedbehavior", "pkg/shortclaim"} {
		f, err := root.Open(dir)
		if err != nil {
			return errors.New("scopeprep_source_unavailable")
		}
		entries, err := f.ReadDir(64)
		if err == nil {
			more, nextErr := f.ReadDir(1)
			if len(more) != 0 || nextErr != io.EOF {
				_ = f.Close()
				return errors.New("scopeprep_source_directory_bounds")
			}
		}
		_ = f.Close()
		if err != nil && err != io.EOF {
			return errors.New("scopeprep_source_unavailable")
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			if !slices.Contains(requiredPaths[:], dir+"/"+name) {
				return errors.New("scopeprep_unlisted_package_source")
			}
		}
	}
	raw, err := boundedSource(root, sourceInputPath)
	if err != nil {
		return err
	}
	if digest(raw) != sourceInputSHA256 {
		return errors.New("scopeprep_source_input_changed")
	}
	return nil
}

func writeNew(path string, raw []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("scopeprep_output_create_failed")
	}
	_, writeErr := f.Write(raw)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		return errors.New("scopeprep_output_write_failed")
	}
	return nil
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("riido-scopeprep", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	stage := fs.String("stage", "prepare", "only prepare is supported; no truth assignment")
	dest := fs.String("out", "", "new output directory")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, _ = fmt.Fprintln(stdout, "Usage: riido-scopeprep --stage prepare --out NEW_DIRECTORY\nRun from the public repository with installed or cached Go1.27.1. This emits proposals only; captions remain pending.")
			return flag.ErrHelp
		}
		return errors.New("scopeprep_invalid_arguments")
	}
	if fs.NArg() != 0 || *dest == "" {
		return errors.New("scopeprep_invalid_arguments")
	}
	if *stage != "prepare" {
		return errors.New("scopeprep_unsupported_stage")
	}
	if runtime.Version() != "go1.27.1" {
		return errors.New("scopeprep_toolchain_not_pinned")
	}
	oldProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldProcs)
	oldHeap := debug.SetMemoryLimit(256 << 20)
	defer debug.SetMemoryLimit(oldHeap)
	root, err := os.OpenRoot(".")
	if err != nil {
		return errors.New("scopeprep_source_unavailable")
	}
	defer root.Close()
	sources, err := compiledArtifacts()
	if err != nil {
		return err
	}
	if err = verifySources(root, sources); err != nil {
		return err
	}
	// Keep the independently pinned disk input and the preparation factory's
	// canonical-original binding on the same immutable version.
	if scopedproperty.OriginalDatasetSHA256 != sourceInputSHA256 {
		return errors.New("scopeprep_original_version_binding_invalid")
	}
	// This operation binds authored captions and source pins. It must not
	// execute candidate implementations or assign property truth.
	dataset, err := scopedproperty.CreateDataset()
	if err != nil {
		return errors.New("scopeprep_dataset_prepare_failed")
	}
	raw, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil || len(raw)+1 > maxPreparedBytes {
		return errors.New("scopeprep_dataset_encoding_failed")
	}
	raw = append(raw, '\n')
	if os.Mkdir(*dest, 0700) != nil {
		return errors.New("scopeprep_output_directory_not_new")
	}
	r := preparation{
		Schema: "riido-scoped-property-preparation-v1", State: "proposals_only_captions_pending",
		GoVersion: runtime.Version(), CPUThreads: 1, GoHeapSoftLimitBytes: 256 << 20,
		SourceInputSHA256: sourceInputSHA256, ProposalsSHA256: digest(raw), ProposalsBytes: len(raw),
		ProposalRows: len(dataset.Parents), Properties: len(scopedproperty.Definitions()), CaptionReview: "pending", CompiledSources: sources,
		StopReasons: []string{"captions_pending", "synthetic_single_pipeline", "no_roles_plan", "no_official_execution_manifest"},
		Limitations: []string{
			"Twelve proposal rows reuse four authored parents and their unchanged candidates; they are not twelve independent tasks or groups.",
			"Literal definitions are authored expected observations, not candidate evaluations or actual truth labels.",
			"Source verification and text budgets do not complete independent caption fidelity review.",
			"This prepares no groups, roles, ranking, fits, model calls, resource measurements or protected-final access.",
			"Go heap is a soft limit, not an OS/native/GPU RSS hard cap; Go P count is not OS thread count or CPU affinity.",
		},
	}
	for _, p := range dataset.Parents {
		r.CandidateDescriptions += len(p.Candidates)
	}
	for _, d := range scopedproperty.Definitions() {
		r.LiteralDefinitions += d.LiteralCount
	}
	report, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return errors.New("scopeprep_summary_encoding_failed")
	}
	report = append(report, '\n')
	if err = writeNew(*dest+"/probes.json", raw); err != nil {
		return err
	}
	if err = writeNew(*dest+"/preparation.json", report); err != nil {
		return err
	}
	// Emit the same bounded public summary; supplied paths and arguments never
	// enter diagnostics. The proposal payload is written only to the new dir.
	if _, err = io.Copy(stdout, bytes.NewReader(report)); err != nil {
		return errors.New("scopeprep_summary_write_failed")
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
