// riido-typedaudit prepares original public typed probes or audits only their
// finite truth and connected groups. It never ranks, partitions, fits, invokes
// a model, accesses protected-final data, or activates a policy.
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
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	planSchema      = "riido-typed-truth-execution-plan-v1"
	planState       = "preparation_truth_groups_only"
	planPolicy      = "truth_groups_only_no_roles_no_fits_no_models"
	resultSchema    = "riido-typed-truth-development-result-v1"
	frozenLegacySHA = "0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0"
	maxFileBytes    = 64 << 20
	goHeapSoftLimit = 256 << 20
	requiredRuntime = "go1.27.1"
)

// Pin the running driver bytes, rather than trusting a mutable filesystem copy.
//
//go:embed main.go
var driverSourceText string

type failure string

func (e failure) Error() string { return string(e) }

const (
	errArguments  failure = "invalid_arguments"
	errRuntime    failure = "runtime_version_unpinned"
	errPlanRead   failure = "plan_read_failed"
	errPlanHash   failure = "plan_hash_mismatch"
	errPlanJSON   failure = "plan_json_invalid"
	errPlan       failure = "unsupported_execution_plan"
	errManifest   failure = "implementation_manifest_invalid"
	errSource     failure = "implementation_source_hash_mismatch"
	errEmbedded   failure = "compiled_source_artifact_mismatch"
	errInput      failure = "typed_input_hash_or_json_invalid"
	errLegacy     failure = "legacy_input_hash_or_json_invalid"
	errPrepare    failure = "typed_preparation_failed"
	errTypecheck  failure = "source_typecheck_failed"
	errLegacyEval failure = "legacy_truth_audit_failed"
	errTypedEval  failure = "typed_truth_audit_failed"
	errGroups     failure = "combined_group_audit_failed"
	errReport     failure = "truth_or_group_report_invalid"
	errOutput     failure = "new_output_directory_or_write_failed"
	errStatus     failure = "status_write_failed"
)

type sourcePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type auditPlan struct {
	Schema                    string      `json:"schema"`
	State                     string      `json:"state"`
	InputSHA256               string      `json:"input_sha256"`
	LegacyInputSHA256         string      `json:"legacy_input_sha256"`
	Parents                   int         `json:"parents"`
	LegacyParents             int         `json:"legacy_parents"`
	MinConnectedGroups        int         `json:"min_connected_groups"`
	MinLabeledConnectedGroups int         `json:"min_labeled_connected_groups"`
	Fits                      int         `json:"fits_authorized_in_this_plan"`
	CPUThreads                int         `json:"cpu_threads"`
	ImplementationFiles       []sourcePin `json:"implementation_files"`
	Policy                    string      `json:"policy"`
}

var requiredFiles = [...]string{
	"internal/typedbehavior/spec.go", "internal/typedbehavior/state.go",
	"internal/typedbehavior/flow.go", "internal/typedbehavior/source.go",
	"internal/typedbehavior/dataset.go", "internal/typedbehavior/fixtures.go",
	"internal/typedbehavior/audit.go", "internal/typedbehavior/groups.go",
	"internal/behaviorprobe/data.go", "internal/lexicalhint/features.go",
	"pkg/shortclaim/input.go", "cmd/riido-typedaudit/main.go",
	"pkg/shortclaim/baseline.go", "internal/behaviorprobe/audit_provenance.go",
	"internal/lexicalhint/audit_provenance.go", "pkg/shortclaim/audit_provenance.go",
}

// Optional tests remain explicit public paths.
// No caller-selected repository, directory wildcard or arbitrary .go file is read.
var optionalFiles = [...]string{
	"internal/typedbehavior/spec_test.go", "internal/typedbehavior/state_test.go",
	"internal/typedbehavior/flow_test.go", "internal/typedbehavior/source_test.go",
	"internal/typedbehavior/dataset_test.go", "internal/typedbehavior/fixtures_test.go",
	"internal/typedbehavior/audit_test.go", "internal/typedbehavior/groups_test.go",
	"internal/behaviorprobe/data_test.go", "internal/lexicalhint/features_test.go",
	"pkg/shortclaim/input_test.go",
	"pkg/shortclaim/baseline_test.go", "cmd/riido-typedaudit/main_test.go",
	"cmd/riido-typedaudit/replay_test.go",
}

type report struct {
	Schema                   string                    `json:"schema"`
	Stage                    string                    `json:"stage"`
	PlanSHA256               string                    `json:"plan_sha256"`
	InputSHA256              string                    `json:"input_sha256"`
	LegacyInputSHA256        string                    `json:"legacy_input_sha256"`
	RuntimeGo                string                    `json:"runtime_go"`
	CPUThreads               int                       `json:"cpu_threads"`
	GoHeapSoftLimitBytes     int64                     `json:"go_heap_soft_limit_bytes"`
	GoHeapLimitScope         string                    `json:"go_heap_limit_scope"`
	ImplementationFiles      []sourcePin               `json:"implementation_files"`
	LegacyTruth              behaviorprobe.Report      `json:"legacy_truth"`
	TypedTruth               typedbehavior.Report      `json:"typed_truth"`
	CombinedGroups           typedbehavior.GroupReport `json:"combined_groups"`
	Parents                  int                       `json:"parents"`
	Candidates               int                       `json:"candidates"`
	SourceVectorChecks       int                       `json:"source_vector_checks"`
	LegacySourceVectorChecks int                       `json:"legacy_source_vector_checks"`
	TypedSourceVectorChecks  int                       `json:"typed_source_vector_checks"`
	ConnectedGroups          int                       `json:"connected_groups"`
	LabeledConnectedGroups   int                       `json:"labeled_connected_groups"`
	MinimumConnectedGroups   int                       `json:"minimum_connected_groups"`
	MinimumLabeledGroups     int                       `json:"minimum_labeled_connected_groups"`
	WholeGroupGatePass       bool                      `json:"whole_group_gate_pass"`
	LabeledGroupGatePass     bool                      `json:"labeled_group_gate_pass"`
	StopReasons              []string                  `json:"stop_reasons"`
	RolesAssigned            bool                      `json:"roles_assigned"`
	Partitions               int                       `json:"partitions"`
	Fits                     int                       `json:"fits"`
	Weights                  int                       `json:"weight_artifacts"`
	ModelCalls               int                       `json:"model_calls"`
	BaselineRankings         int                       `json:"baseline_rankings"`
	PerformanceRuns          int                       `json:"performance_runs"`
	TrainingExecutionReady   bool                      `json:"training_execution_ready"`
	ProductionActivation     bool                      `json:"production_activation"`
	FinalEligible            bool                      `json:"final_eligible"`
	ProtectedFinalRead       bool                      `json:"protected_final_read"`
	Scope                    string                    `json:"scope"`
}

func hash(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func validSHA(s string) bool {
	if len(s) != 64 || s != strings.ToLower(s) {
		return false
	}
	_, e := hex.DecodeString(s)
	return e == nil
}

func readBounded(name string) ([]byte, error) {
	f, e := os.Open(name)
	if e != nil {
		return nil, errInput
	}
	defer f.Close()
	return readRegular(f)
}

func readRegular(f *os.File) ([]byte, error) {
	info, e := f.Stat()
	if e != nil || !info.Mode().IsRegular() || info.Size() > maxFileBytes {
		return nil, errInput
	}
	raw, e := io.ReadAll(io.LimitReader(f, maxFileBytes+1))
	if e != nil || len(raw) > maxFileBytes {
		return nil, errInput
	}
	return raw, nil
}

// Transport checking rejects null, duplicate object keys, excess nesting and
// extra JSON values before the typed decoder. Schema keys use a lowercase ASCII
// namespace, preventing encoding/json's case-insensitive field-name matching.
// No token value enters diagnostics.
func checkJSON(raw []byte) bool {
	if len(raw) == 0 || len(raw) > maxFileBytes || !utf8.Valid(raw) {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	first, e := d.Token()
	if e != nil || first != json.Delim('{') || !checkJSONContainer(d, '{', 0) {
		return false
	}
	_, e = d.Token()
	return e == io.EOF
}

func checkJSONContainer(d *json.Decoder, kind json.Delim, depth int) bool {
	if depth > 64 {
		return false
	}
	var keys []string
	for d.More() {
		if kind == '{' {
			key, e := d.Token()
			name, ok := key.(string)
			if e != nil || !ok || !schemaKey(name) || len(keys) >= 128 || slices.Contains(keys, name) {
				return false
			}
			keys = append(keys, name)
		}
		value, e := d.Token()
		if e != nil || value == nil {
			return false
		}
		if nested, ok := value.(json.Delim); ok {
			if nested != '{' && nested != '[' || !checkJSONContainer(d, nested, depth+1) {
				return false
			}
		}
	}
	end, e := d.Token()
	return e == nil && (kind == '{' && end == json.Delim('}') || kind == '[' && end == json.Delim(']'))
}

func schemaKey(name string) bool {
	if len(name) == 0 || len(name) > 128 {
		return false
	}
	for i := range len(name) {
		c := name[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func strictDecode(raw []byte, value any) bool {
	if !checkJSON(raw) {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		return false
	}
	return d.Decode(new(any)) == io.EOF
}

func decodePlan(raw []byte, p *auditPlan) bool {
	if !strictDecode(raw, p) {
		return false
	}
	// Unknown and duplicate fields were already rejected. Requiring all twelve
	// keys prevents an omitted zero-valued fits authorization from being inferred.
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return false
	}
	fields := 0
	for d.More() {
		if _, e := d.Token(); e != nil || d.Decode(new(json.RawMessage)) != nil {
			return false
		}
		fields++
	}
	return fields == 12
}

func validatePlan(p auditPlan) error {
	if p.Schema != planSchema || p.State != planState || p.Policy != planPolicy || !validSHA(p.InputSHA256) || p.LegacyInputSHA256 != frozenLegacySHA || p.Parents != 24 || p.LegacyParents != 48 || p.MinConnectedGroups != 15 || p.MinLabeledConnectedGroups != 15 || p.Fits != 0 || p.CPUThreads != 1 {
		return errPlan
	}
	if len(p.ImplementationFiles) < len(requiredFiles) || len(p.ImplementationFiles) > 32 {
		return errManifest
	}
	for i, pin := range p.ImplementationFiles {
		if !validSHA(pin.SHA256) || filepath.IsAbs(pin.Path) || path.Clean(pin.Path) != pin.Path || strings.Contains(pin.Path, "\\") || strings.Contains(pin.Path, "..") || !strings.HasSuffix(pin.Path, ".go") || !slices.Contains(requiredFiles[:], pin.Path) && !slices.Contains(optionalFiles[:], pin.Path) {
			return errManifest
		}
		for j := 0; j < i; j++ {
			if p.ImplementationFiles[j].Path == pin.Path {
				return errManifest
			}
		}
	}
	for _, required := range requiredFiles {
		if !slices.ContainsFunc(p.ImplementationFiles, func(pin sourcePin) bool { return pin.Path == required }) {
			return errManifest
		}
	}
	return nil
}

func pinSHA(p auditPlan, name string) string {
	for _, pin := range p.ImplementationFiles {
		if pin.Path == name {
			return pin.SHA256
		}
	}
	return ""
}

func verifySources(p auditPlan, repo string) error {
	root, e := os.OpenRoot(repo)
	if e != nil {
		return errSource
	}
	defer root.Close()
	for _, pin := range p.ImplementationFiles {
		f, e := root.Open(filepath.FromSlash(pin.Path))
		if e != nil {
			return errSource
		}
		raw, readErr := readRegular(f)
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || hash(raw) != pin.SHA256 {
			return errSource
		}
	}
	if pinSHA(p, "cmd/riido-typedaudit/main.go") != hash([]byte(driverSourceText)) {
		return errEmbedded
	}
	artifacts := typedbehavior.SourceArtifacts()
	if len(artifacts) != 8 {
		return errEmbedded
	}
	for _, artifact := range artifacts {
		if pinSHA(p, "internal/typedbehavior/"+artifact.Name) != artifact.SHA256 {
			return errEmbedded
		}
	}
	for _, artifact := range behaviorprobe.AuditSourceArtifacts() {
		if pinSHA(p, "internal/behaviorprobe/"+artifact.Name) != artifact.SHA256 {
			return errEmbedded
		}
	}
	for _, artifact := range lexicalhint.AuditSourceArtifacts() {
		if pinSHA(p, "internal/lexicalhint/"+artifact.Name) != artifact.SHA256 {
			return errEmbedded
		}
	}
	for _, artifact := range shortclaim.AuditSourceArtifacts() {
		if pinSHA(p, "pkg/shortclaim/"+artifact.Name) != artifact.SHA256 {
			return errEmbedded
		}
	}
	return nil
}

func controlChecks(controls []behaviorprobe.ControlTruth) (int, bool) {
	total := 0
	for _, c := range controls {
		if c.VectorsChecked < 1 || c.FailedVectors < 0 || c.FailedVectors > c.VectorsChecked || c.ExpectedControl != "correct" && c.ExpectedControl != "wrong" || c.ExpectedControl == "correct" && c.FailedVectors != 0 || c.ExpectedControl == "wrong" && c.FailedVectors == 0 {
			return 0, false
		}
		total += c.VectorsChecked
	}
	return total, len(controls) > 0
}

func labeledGroups(groups typedbehavior.GroupReport, outcomes []behaviorprobe.Outcome) (int, bool) {
	if groups.ConnectedGroups != len(groups.Groups) || groups.Parents != len(outcomes) {
		return 0, false
	}
	seen := make([]bool, len(outcomes))
	labeled := 0
	for i, g := range groups.Groups {
		if g.ID < 0 || g.ID >= len(outcomes) || len(g.Parents) == 0 {
			return 0, false
		}
		for j := 0; j < i; j++ {
			if groups.Groups[j].ID == g.ID {
				return 0, false
			}
		}
		hasKnown := false
		for _, id := range g.Parents {
			n := slices.IndexFunc(outcomes, func(o behaviorprobe.Outcome) bool { return o.ParentID == id })
			if n < 0 || seen[n] {
				return 0, false
			}
			seen[n] = true
			switch outcomes[n].State {
			case "known":
				if len(outcomes[n].Acceptable) == 0 {
					return 0, false
				}
				hasKnown = true
			case "no_answer":
				if len(outcomes[n].Acceptable) != 0 {
					return 0, false
				}
				hasKnown = true
			case "unknown":
				if len(outcomes[n].Acceptable) != 0 {
					return 0, false
				}
			default:
				return 0, false
			}
		}
		if hasKnown {
			labeled++
		}
	}
	return labeled, !slices.Contains(seen, false)
}

func buildReport(p auditPlan, planSHA string, legacy behaviorprobe.Report, typed typedbehavior.Report, groups typedbehavior.GroupReport) (report, error) {
	if validatePlan(p) != nil || !validSHA(planSHA) || legacy.Parents != p.LegacyParents || typed.Parents != p.Parents || len(legacy.Outcomes) != legacy.Parents || len(typed.Outcomes) != typed.Parents || legacy.TrainingReady || legacy.Partitioned || legacy.FinalEligible || typed.TrainingReady || typed.Partitioned || typed.FinalEligible || groups.TrainingReady || groups.MinimumGroups != p.MinConnectedGroups || groups.MinimumLabeledGroups != p.MinLabeledConnectedGroups {
		return report{}, errReport
	}
	outcomes := append(slices.Clone(legacy.Outcomes), typed.Outcomes...)
	labeled, ok := labeledGroups(groups, outcomes)
	if !ok || labeled != groups.LabeledConnectedGroups || groups.MeetsGroupMinimum != (groups.ConnectedGroups >= p.MinConnectedGroups) || groups.MeetsLabeledGroupMinimum != (labeled >= p.MinLabeledConnectedGroups) {
		return report{}, errReport
	}
	legacyChecks, ok := controlChecks(legacy.Controls)
	if !ok {
		return report{}, errReport
	}
	typedControls := make([]behaviorprobe.ControlTruth, len(typed.Controls))
	for i, c := range typed.Controls {
		typedControls[i] = behaviorprobe.ControlTruth{SourceID: c.SourceID, ExpectedControl: c.ExpectedControl, VectorsChecked: c.VectorsChecked, FailedVectors: c.FailedVectors, TruthTableSHA256: c.TruthTableSHA256}
	}
	typedChecks, ok := controlChecks(typedControls)
	if !ok {
		return report{}, errReport
	}
	r := report{
		Schema: resultSchema, Stage: planState, PlanSHA256: planSHA, InputSHA256: p.InputSHA256, LegacyInputSHA256: p.LegacyInputSHA256,
		RuntimeGo: requiredRuntime, CPUThreads: 1, GoHeapSoftLimitBytes: goHeapSoftLimit,
		GoHeapLimitScope:    "Go heap soft target only; not a total-process RSS/native/GPU hard limit or a performance measurement.",
		ImplementationFiles: slices.Clone(p.ImplementationFiles), LegacyTruth: legacy, TypedTruth: typed, CombinedGroups: groups,
		Parents: legacy.Parents + typed.Parents, Candidates: legacy.Candidates + typed.Candidates,
		LegacySourceVectorChecks: legacyChecks, TypedSourceVectorChecks: typedChecks, SourceVectorChecks: legacyChecks + typedChecks,
		ConnectedGroups: groups.ConnectedGroups, LabeledConnectedGroups: labeled, MinimumConnectedGroups: p.MinConnectedGroups, MinimumLabeledGroups: p.MinLabeledConnectedGroups,
		WholeGroupGatePass: groups.MeetsGroupMinimum, LabeledGroupGatePass: groups.MeetsLabeledGroupMinimum,
		Scope: "Original finite development controls from one coordinated synthetic pipeline. Source-vector checks are assertions, not distinct requests. Group minima do not establish role readiness, generalization, LLM savings or final eligibility. Acquire at least 2400 distinct protected-final requests per domain separately.",
	}
	if !r.WholeGroupGatePass {
		r.StopReasons = append(r.StopReasons, "insufficient_whole_connected_groups")
	}
	if !r.LabeledGroupGatePass {
		r.StopReasons = append(r.StopReasons, "insufficient_labeled_connected_groups")
	}
	r.StopReasons = append(r.StopReasons, "no_roles_plan", "synthetic_single_pipeline")
	return r, nil
}

func newOutput(name string, raw []byte, filename string) error {
	if e := os.Mkdir(name, 0700); e != nil {
		return errOutput
	}
	f, e := os.OpenFile(filepath.Join(name, filename), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errOutput
	}
	_, writeErr := f.Write(append(raw, '\n'))
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		return errOutput
	}
	return nil
}

func prepare(out string, stdout io.Writer) error {
	d, e := typedbehavior.PrepareDataset()
	if e != nil {
		if e.Error() == "source_typecheck_failed" {
			return errTypecheck
		}
		return errPrepare
	}
	raw, e := json.MarshalIndent(d, "", "  ")
	if e != nil || newOutput(out, raw, "probes.json") != nil {
		return errOutput
	}
	status := struct {
		Schema      string `json:"schema"`
		Stage       string `json:"stage"`
		InputSHA256 string `json:"input_sha256"`
		Parents     int    `json:"parents"`
		Evaluations int    `json:"candidate_evaluations"`
		GroupAudits int    `json:"group_audits"`
		Rankings    int    `json:"rankings"`
		Fits        int    `json:"fits"`
	}{"riido-typed-truth-stage-status-v1", "prepare", hash(append(raw, '\n')), len(d.Parents), 0, 0, 0, 0}
	if json.NewEncoder(stdout).Encode(status) != nil {
		return errStatus
	}
	return nil
}

func run(args []string, stdout io.Writer) error { return runIn(args, stdout, ".") }

func runIn(args []string, stdout io.Writer, repo string) error {
	fs := flag.NewFlagSet("riido-typedaudit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	stage := fs.String("stage", "", "prepare or audit")
	planFile := fs.String("plan", "", "frozen public execution plan")
	planSHA := fs.String("plan-sha256", "", "exact plan SHA-256")
	input := fs.String("input", "", "typed public input")
	legacyFile := fs.String("legacy", "", "immutable legacy public input")
	out := fs.String("out", "", "new output directory")
	if fs.Parse(args) != nil || fs.NArg() != 0 || *out == "" || *stage != "prepare" && *stage != "audit" {
		return errArguments
	}
	if *stage == "prepare" && (*planFile != "" || *planSHA != "" || *input != "" || *legacyFile != "") || *stage == "audit" && (*planFile == "" || !validSHA(*planSHA) || *input == "" || *legacyFile == "") {
		return errArguments
	}
	if runtime.Version() != requiredRuntime {
		return errRuntime
	}
	if _, e := os.Lstat(*out); e == nil || !errors.Is(e, os.ErrNotExist) {
		return errOutput
	}
	previousThreads := runtime.GOMAXPROCS(1)
	previousLimit := debug.SetMemoryLimit(goHeapSoftLimit)
	defer runtime.GOMAXPROCS(previousThreads)
	defer debug.SetMemoryLimit(previousLimit)
	if *stage == "prepare" {
		return prepare(*out, stdout)
	}
	planRaw, e := readBounded(*planFile)
	if e != nil {
		return errPlanRead
	}
	if hash(planRaw) != *planSHA {
		return errPlanHash
	}
	var p auditPlan
	if !decodePlan(planRaw, &p) {
		return errPlanJSON
	}
	if e = validatePlan(p); e != nil {
		return e
	}
	if e = verifySources(p, repo); e != nil {
		return e
	}
	inputRaw, e := readBounded(*input)
	if e != nil || hash(inputRaw) != p.InputSHA256 || !checkJSON(inputRaw) {
		return errInput
	}
	// LoadBytes decodes exactly this verified immutable byte slice, not a path.
	d, e := typedbehavior.LoadBytes(inputRaw)
	if e != nil || len(d.Parents) != p.Parents {
		return errInput
	}
	legacyRaw, e := readBounded(*legacyFile)
	if e != nil || hash(legacyRaw) != p.LegacyInputSHA256 {
		return errLegacy
	}
	var legacy behaviorprobe.Dataset
	if !strictDecode(legacyRaw, &legacy) || len(legacy.Parents) != p.LegacyParents {
		return errLegacy
	}
	// Evaluate validates schema, origin, registered contract/source hashes,
	// duplicate IDs, UTF-8 and text/candidate bounds on these decoded bytes.
	lr, e := behaviorprobe.Evaluate(legacy)
	if e != nil || lr.SourceArtifactSHA256 != pinSHA(p, "internal/behaviorprobe/data.go") {
		return errLegacyEval
	}
	tr, e := typedbehavior.Evaluate(d)
	if e != nil {
		return errTypedEval
	}
	groups, e := typedbehavior.CombinedGroups(legacy, lr, d, tr)
	if e != nil {
		return errGroups
	}
	r, e := buildReport(p, *planSHA, lr, tr, groups)
	if e != nil {
		return e
	}
	raw, e := json.MarshalIndent(r, "", "  ")
	if e != nil || newOutput(*out, raw, "results.json") != nil {
		return errOutput
	}
	status := struct {
		Schema                 string `json:"schema"`
		Stage                  string `json:"stage"`
		Parents                int    `json:"parents"`
		ConnectedGroups        int    `json:"connected_groups"`
		LabeledConnectedGroups int    `json:"labeled_connected_groups"`
		TrainingExecutionReady bool   `json:"training_execution_ready"`
		Fits                   int    `json:"fits"`
	}{"riido-typed-truth-stage-status-v1", "audit", r.Parents, r.ConnectedGroups, r.LabeledConnectedGroups, false, 0}
	if json.NewEncoder(stdout).Encode(status) != nil {
		return errStatus
	}
	return nil
}

func main() {
	if e := run(os.Args[1:], os.Stdout); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
