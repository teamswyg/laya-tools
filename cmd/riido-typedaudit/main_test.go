package main

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
)

// Exercise the maintainer executable, including trimmed build paths, rather
// than only a go-test binary whose GOROOT metadata may mask exporter failures.
// This prepares metadata only: no candidate evaluator or group audit is called.
func TestPackagedTrimpathPrepareWithPinnedGo(t *testing.T) {
	repo, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal("could not identify the public repository")
	}
	directory := t.TempDir()
	name := "riido-typedaudit"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(directory, name)
	// Resolve the explicitly pinned cached toolchain without contacting module
	// services. go test itself supplies GOROOT, so explicitly omit that test-only
	// environment hint. A trimmed standalone tool must resolve its own exports.
	settings := []string{"GOTOOLCHAIN=go1.27.1", "GOPROXY=off", "GOSUMDB=off", "GOENV=off", "GOFLAGS=", "GOWORK=off", "CGO_ENABLED=0", "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}
	environment := slices.DeleteFunc(slices.Clone(os.Environ()), func(value string) bool { return strings.HasPrefix(value, "GOROOT=") })
	for _, setting := range settings {
		key, _, _ := strings.Cut(setting, "=")
		environment = slices.DeleteFunc(environment, func(value string) bool { return strings.HasPrefix(value, key+"=") })
	}
	environment = append(environment, settings...)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	version := exec.CommandContext(ctx, "go", "version")
	version.Dir, version.Env = repo, environment
	var versionOutput bytes.Buffer
	version.Stdout, version.Stderr = &versionOutput, io.Discard
	if version.Run() != nil || !strings.HasPrefix(versionOutput.String(), "go version go1.27.1 ") {
		t.Fatal("maintainer build toolchain is not Go 1.27.1")
	}
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./cmd/riido-typedaudit")
	build.Dir, build.Env = repo, environment
	build.Stdout, build.Stderr = io.Discard, io.Discard
	if build.Run() != nil {
		t.Fatal("offline trimmed-path maintainer build failed")
	}
	info, e := buildinfo.ReadFile(binary)
	if e != nil || info.GoVersion != "go1.27.1" {
		t.Fatal("packaged executable was not built with Go 1.27.1")
	}
	trimmed, pureGo := false, false
	for _, setting := range info.Settings {
		trimmed = trimmed || setting.Key == "-trimpath" && setting.Value == "true"
		pureGo = pureGo || setting.Key == "CGO_ENABLED" && setting.Value == "0"
	}
	if !trimmed || !pureGo {
		t.Fatal("packaged build metadata does not prove the pinned packaging recipe")
	}
	out := filepath.Join(directory, "prepared")
	prepare := exec.CommandContext(ctx, binary, "--stage", "prepare", "--out", out)
	prepare.Dir, prepare.Env = repo, environment
	var status, diagnostics bytes.Buffer
	prepare.Stdout, prepare.Stderr = &status, &diagnostics
	if prepare.Run() != nil || diagnostics.Len() != 0 {
		t.Fatal("packaged trimmed-path metadata preparation failed")
	}
	entries, e := os.ReadDir(out)
	if e != nil || len(entries) != 1 || entries[0].Name() != "probes.json" {
		t.Fatal("packaged prepare must emit only the transport dataset")
	}
	raw, e := os.ReadFile(filepath.Join(out, "probes.json"))
	if e != nil {
		t.Fatal("could not read packaged public preparation")
	}
	d, e := typedbehavior.LoadBytes(raw)
	if e != nil || len(d.Parents) != 24 {
		t.Fatal("packaged preparation is not source-bound valid transport")
	}
	var observed struct {
		Schema      string `json:"schema"`
		Stage       string `json:"stage"`
		InputSHA256 string `json:"input_sha256"`
		Parents     int    `json:"parents"`
		Evaluations int    `json:"candidate_evaluations"`
		GroupAudits int    `json:"group_audits"`
		Rankings    int    `json:"rankings"`
		Fits        int    `json:"fits"`
	}
	if !strictDecode(status.Bytes(), &observed) || observed.Schema != "riido-typed-truth-stage-status-v1" || observed.Stage != "prepare" || observed.InputSHA256 != hash(raw) || observed.Parents != 24 || observed.Evaluations != 0 || observed.GroupAudits != 0 || observed.Rankings != 0 || observed.Fits != 0 {
		t.Fatal("packaged status does not bind a preparation-only operation")
	}
	for _, forbidden := range []string{directory, repo, "acceptable_candidate_indices", "vectors_checked", "connected_groups", "training_ready"} {
		if bytes.Contains(raw, []byte(forbidden)) || strings.Contains(status.String(), forbidden) {
			t.Fatal("packaged preparation emitted paths, labels or audit state")
		}
	}
}

func testPlan() auditPlan {
	p := auditPlan{
		Schema: planSchema, State: planState, Policy: planPolicy,
		InputSHA256: hash([]byte("synthetic transport")), LegacyInputSHA256: frozenLegacySHA,
		Parents: 24, LegacyParents: 48, MinConnectedGroups: 15, MinLabeledConnectedGroups: 15,
		Fits: 0, CPUThreads: 1,
	}
	for _, file := range requiredFiles {
		p.ImplementationFiles = append(p.ImplementationFiles, sourcePin{file, hash([]byte(file))})
	}
	return p
}

func TestPrepareEmitsOnlyTransportAndFreshDirectory(t *testing.T) {
	out := filepath.Join(t.TempDir(), "prepare")
	var status bytes.Buffer
	if e := run([]string{"--stage", "prepare", "--out", out}, &status); e != nil {
		t.Fatal(e)
	}
	entries, e := os.ReadDir(out)
	if e != nil || len(entries) != 1 || entries[0].Name() != "probes.json" {
		t.Fatal("prepare must emit one transport file, no truth/group/report artifacts")
	}
	raw, e := os.ReadFile(filepath.Join(out, "probes.json"))
	if e != nil {
		t.Fatal(e)
	}
	d, e := typedbehavior.LoadBytes(raw)
	if e != nil || len(d.Parents) != 24 {
		t.Fatal("prepared bytes do not pass the source-bound transport loader")
	}
	var s struct {
		Schema      string `json:"schema"`
		Stage       string `json:"stage"`
		InputSHA256 string `json:"input_sha256"`
		Parents     int    `json:"parents"`
		Evaluations int    `json:"candidate_evaluations"`
		GroupAudits int    `json:"group_audits"`
		Rankings    int    `json:"rankings"`
		Fits        int    `json:"fits"`
	}
	if !strictDecode(status.Bytes(), &s) || s.Stage != "prepare" || s.InputSHA256 != hash(raw) || s.Parents != 24 || s.Evaluations != 0 || s.GroupAudits != 0 || s.Rankings != 0 || s.Fits != 0 {
		t.Fatal("preparation status has incorrect binding or execution claims")
	}
	for _, forbidden := range []string{out, "acceptable_candidate_indices", "vectors_checked", "connected_groups", "training_ready"} {
		if bytes.Contains(raw, []byte(forbidden)) || strings.Contains(status.String(), forbidden) {
			t.Fatal("preparation exposed paths, labels or audit state")
		}
	}
	if e := run([]string{"--stage", "prepare", "--out", out}, io.Discard); e != errOutput {
		t.Fatal("an existing output directory must never be replaced")
	}
	after, e := os.ReadFile(filepath.Join(out, "probes.json"))
	if e != nil || !bytes.Equal(raw, after) {
		t.Fatal("rejected output reuse changed the prepared bytes")
	}
}

func TestArgumentAndReaderErrorsAreFixedAndRedacted(t *testing.T) {
	marker := "synthetic-private-marker"
	for _, args := range [][]string{
		{"--" + marker},
		{"--stage", marker, "--out", marker},
		{"--stage", "prepare", "--out", marker, "--input", marker},
		{"--stage", "audit", "--out", marker, "--plan", marker, "--plan-sha256", marker, "--input", marker, "--legacy", marker},
		{"--stage", "prepare", "--out", marker, marker},
	} {
		var output bytes.Buffer
		e := run(args, &output)
		if e != errArguments || output.Len() != 0 || strings.Contains(e.Error(), marker) {
			t.Fatal("raw flags or values escaped fixed diagnostics")
		}
	}
	var output bytes.Buffer
	out := filepath.Join(t.TempDir(), "audit")
	e := run([]string{"--stage", "audit", "--out", out, "--plan", marker, "--plan-sha256", hash([]byte(marker)), "--input", marker, "--legacy", marker}, &output)
	if e != errPlanRead || output.Len() != 0 || strings.Contains(e.Error(), marker) {
		t.Fatal("reader diagnostics disclosed an argument path")
	}
}

func TestStrictJSONRejectsNullDuplicateUnknownAndExtraValues(t *testing.T) {
	p := testPlan()
	raw, e := json.Marshal(p)
	if e != nil || !decodePlan(raw, new(auditPlan)) {
		t.Fatal("valid plan transport rejected")
	}
	for _, bad := range [][]byte{
		[]byte("null"), []byte("[]"), []byte("{} {}"), append(slices.Clone(raw), '{', '}'),
		bytes.Replace(raw, []byte(`"cpu_threads":1`), []byte(`"cpu_threads":null`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1),
		bytes.Replace(raw, []byte(`"state":`), []byte(`"training_execution_ready":true,"state":`), 1),
		bytes.Replace(raw, []byte(`"sha256":`), []byte(`"sha256":"duplicate","sha256":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"SCHEMA":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"duplicate","SCHEMA":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"ScHeMa":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"\u0053CHEMA":`), 1),
		bytes.Replace(raw, []byte(`"sha256":`), []byte(`"SHA256":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"schéma":`), 1),
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"":`), 1),
		append([]byte{0xff}, raw...),
	} {
		if decodePlan(bad, new(auditPlan)) {
			t.Fatal("null, ambiguous or unauthorized transport was accepted")
		}
	}
	missingFits := bytes.Replace(raw, []byte(`"fits_authorized_in_this_plan":0,`), nil, 1)
	if decodePlan(missingFits, new(auditPlan)) {
		t.Fatal("missing authorization field was inferred from a Go zero value")
	}
	// Nested null is rejected in legacy transport as well, rather than becoming
	// an empty string/slice or an implicit zero-value configuration.
	if strictDecode([]byte(`{"schema":"x","origin":"x","parents":[{"id":"p","prototype":"x","contract_id":"x","request":null,"candidates":[]}]}`), new(behaviorprobe.Dataset)) {
		t.Fatal("legacy nested null was silently coerced")
	}
}

func TestEveryRuntimeDependencyRejectsStaleCompiledSource(t *testing.T) {
	if len(requiredFiles) != 16 {
		t.Fatal("compiled manifest must include all sixteen runtime files")
	}
	for _, name := range requiredFiles {
		t.Run(name, func(t *testing.T) {
			repo, p := copyPublicSourceRepo(t)
			file := filepath.Join(repo, filepath.FromSlash(name))
			raw, e := os.ReadFile(file)
			if e != nil {
				t.Fatal("could not read public temporary source")
			}
			changed := append(raw, '\n')
			if os.WriteFile(file, changed, 0600) != nil {
				t.Fatal("could not modify public temporary source")
			}
			for i := range p.ImplementationFiles {
				if p.ImplementationFiles[i].Path == name {
					p.ImplementationFiles[i].SHA256 = hash(changed)
				}
			}
			if e := verifySources(p, repo); e != errEmbedded {
				t.Fatal("rebound on-disk source described stale compiled runtime code")
			}
		})
	}
}

func TestPlanRejectsChangedAuthorityAndIncompleteManifests(t *testing.T) {
	tests := []struct {
		name string
		edit func(*auditPlan)
	}{
		{"schema", func(p *auditPlan) { p.Schema = "different" }},
		{"state", func(p *auditPlan) { p.State = "fit" }},
		{"policy", func(p *auditPlan) { p.Policy = "activate" }},
		{"fits", func(p *auditPlan) { p.Fits = 1 }},
		{"cpu", func(p *auditPlan) { p.CPUThreads = 2 }},
		{"parents", func(p *auditPlan) { p.Parents = 23 }},
		{"legacy parents", func(p *auditPlan) { p.LegacyParents = 47 }},
		{"legacy SHA", func(p *auditPlan) { p.LegacyInputSHA256 = hash([]byte("changed")) }},
		{"input SHA", func(p *auditPlan) { p.InputSHA256 = strings.Repeat("z", 64) }},
		{"whole minimum", func(p *auditPlan) { p.MinConnectedGroups = 14 }},
		{"labeled minimum", func(p *auditPlan) { p.MinLabeledConnectedGroups = 14 }},
		{"missing required", func(p *auditPlan) { p.ImplementationFiles = p.ImplementationFiles[1:] }},
		{"duplicate required", func(p *auditPlan) { p.ImplementationFiles[1] = p.ImplementationFiles[0] }},
		{"absolute", func(p *auditPlan) { p.ImplementationFiles[0].Path = "/private/source.go" }},
		{"traversal", func(p *auditPlan) { p.ImplementationFiles[0].Path = "../source.go" }},
		{"clean alias", func(p *auditPlan) { p.ImplementationFiles[0].Path = "internal/typedbehavior/../typedbehavior/spec.go" }},
		{"backslash", func(p *auditPlan) { p.ImplementationFiles[0].Path = `internal\typedbehavior\spec.go` }},
		{"unowned", func(p *auditPlan) { p.ImplementationFiles[0].Path = "other/source.go" }},
		{"source SHA", func(p *auditPlan) { p.ImplementationFiles[0].SHA256 = "invalid" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := testPlan()
			tt.edit(&p)
			if validatePlan(p) == nil {
				t.Fatal("changed authority or unbound source was accepted")
			}
		})
	}
}

func copyPublicSourceRepo(t *testing.T) (string, auditPlan) {
	t.Helper()
	repo := t.TempDir()
	p := testPlan()
	for i, pin := range p.ImplementationFiles {
		raw, e := os.ReadFile(filepath.Join("../..", filepath.FromSlash(pin.Path)))
		if e != nil {
			t.Fatal("could not copy an owned public source fixture")
		}
		name := filepath.Join(repo, filepath.FromSlash(pin.Path))
		if os.MkdirAll(filepath.Dir(name), 0700) != nil || os.WriteFile(name, raw, 0600) != nil {
			t.Fatal("could not create a public temporary source fixture")
		}
		p.ImplementationFiles[i].SHA256 = hash(raw)
	}
	return repo, p
}

func TestSourceManifestAndCompiledArtifactBinding(t *testing.T) {
	repo, p := copyPublicSourceRepo(t)
	if e := verifySources(p, repo); e != nil {
		t.Fatal(e)
	}
	changed := filepath.Join(repo, "internal/typedbehavior/state.go")
	if e := os.WriteFile(changed, []byte("package typedbehavior\n"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := verifySources(p, repo); e != errSource {
		t.Fatal("source bytes changing after plan binding were accepted")
	}
	for i := range p.ImplementationFiles {
		if p.ImplementationFiles[i].Path == "internal/typedbehavior/state.go" {
			p.ImplementationFiles[i].SHA256 = hash([]byte("package typedbehavior\n"))
		}
	}
	if e := verifySources(p, repo); e != errEmbedded {
		t.Fatal("new on-disk source hash was allowed to describe stale compiled code")
	}
	outside := filepath.Join(t.TempDir(), "public-source.go")
	if os.WriteFile(outside, []byte("package typedbehavior\n"), 0600) != nil || os.Remove(changed) != nil || os.Symlink(outside, changed) != nil {
		t.Fatal("could not create a public source escape control")
	}
	if e := verifySources(p, repo); e != errSource {
		t.Fatal("owned source path escaped the repository root")
	}
}

func writeTestPlan(t *testing.T, p auditPlan, directory string) (string, string) {
	t.Helper()
	raw, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	name := filepath.Join(directory, "plan.json")
	if os.WriteFile(name, raw, 0600) != nil {
		t.Fatal("could not write synthetic validation plan")
	}
	return name, hash(raw)
}

// These audit rejections stop before either candidate evaluator is called.
// The official 24+48 truth/groups report is collected only after root freezes
// its actual plan; this test never runs that report or observes its group count.
func TestAuditRejectsPlanAndSourceBundleForgeryBeforeTruth(t *testing.T) {
	repo, p := copyPublicSourceRepo(t)
	d, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	d.Parents[0].Candidates[0].BundleSHA256 = strings.Repeat("0", 64)
	inputRaw, e := json.Marshal(d)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	if os.WriteFile(input, inputRaw, 0600) != nil {
		t.Fatal("could not write synthetic forged transport")
	}
	p.InputSHA256 = hash(inputRaw)
	planFile, planSHA := writeTestPlan(t, p, dir)
	out := filepath.Join(dir, "out")
	args := []string{"--stage", "audit", "--plan", planFile, "--plan-sha256", planSHA, "--input", input, "--legacy", "synthetic-unread-legacy-marker", "--out", out}
	var stdout bytes.Buffer
	if e := runIn(args, &stdout, repo); e != errInput || stdout.Len() != 0 {
		t.Fatal("forged source closure reached truth or exposed the unread legacy argument")
	}
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		t.Fatal("a failed preflight must not publish result artifacts")
	}
	args[5] = hash([]byte("wrong plan bytes"))
	if e := runIn(args, &stdout, repo); e != errPlanHash {
		t.Fatal("changed plan digest was accepted")
	}
	// A malformed/unknown/null input remains a rejection even if its raw hash
	// is correctly bound by a new synthetic validation plan.
	for _, raw := range [][]byte{[]byte("null"), []byte("{} {}"), []byte(`{"schema":null}`), []byte{0xff}} {
		p.InputSHA256 = hash(raw)
		if os.WriteFile(input, raw, 0600) != nil {
			t.Fatal("could not write malformed public control")
		}
		planFile, planSHA = writeTestPlan(t, p, dir)
		args[3], args[5] = planFile, planSHA
		if e := runIn(args, &stdout, repo); e != errInput {
			t.Fatal("malformed input acquired a truth label")
		}
	}
}

func TestReadBoundsAndOutputNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "oversized.json")
	f, e := os.Create(name)
	if e != nil || f.Truncate(maxFileBytes+1) != nil || f.Close() != nil {
		t.Fatal("could not create bounded-reader control")
	}
	if _, e := readBounded(name); e == nil {
		t.Fatal("oversized file accepted")
	}
	if _, e := readBounded(dir); e == nil {
		t.Fatal("directory accepted as an input stream")
	}
	out := filepath.Join(dir, "new")
	if e := newOutput(out, []byte(`{"original":true}`), "results.json"); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(filepath.Join(out, "results.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e := newOutput(out, []byte(`{"replace":true}`), "results.json"); e != errOutput {
		t.Fatal("existing results were overwritten")
	}
	after, e := os.ReadFile(filepath.Join(out, "results.json"))
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("failed replacement modified an existing result")
	}
}

// Report-only controls are deliberately synthetic. No candidate functions,
// production dataset, natural-language ranking or empirical groups are run.
func syntheticReports(labeled int) (auditPlan, behaviorprobe.Report, typedbehavior.Report, typedbehavior.GroupReport) {
	p := testPlan()
	legacy := behaviorprobe.Report{Parents: 48, Candidates: 48, Controls: []behaviorprobe.ControlTruth{
		{SourceID: "synthetic-correct", ExpectedControl: "correct", VectorsChecked: 3},
		{SourceID: "synthetic-wrong", ExpectedControl: "wrong", VectorsChecked: 3, FailedVectors: 1},
	}}
	typed := typedbehavior.Report{Parents: 24, Candidates: 24, Controls: []typedbehavior.ControlTruth{
		{SourceID: "synthetic-correct", ExpectedControl: "correct", VectorsChecked: 5},
		{SourceID: "synthetic-wrong", ExpectedControl: "wrong", VectorsChecked: 5, FailedVectors: 2},
	}}
	groups := typedbehavior.GroupReport{Parents: 72, ConnectedGroups: 15, MinimumGroups: 15, LabeledConnectedGroups: labeled, MinimumLabeledGroups: 15, MeetsGroupMinimum: true, MeetsLabeledGroupMinimum: labeled >= 15}
	for i := 0; i < 15; i++ {
		// Sparse IDs are original parent root indexes, not dense group ordinals.
		groups.Groups = append(groups.Groups, behaviorprobe.Group{ID: i * 4})
	}
	for i := 0; i < 72; i++ {
		group := i % 15
		id := fmt.Sprintf("synthetic-parent-%02d", i)
		o := behaviorprobe.Outcome{ParentID: id, State: "unknown", Acceptable: []int{}, Candidates: []behaviorprobe.CandidateTruth{{CandidateID: "synthetic", State: "unknown"}}}
		if group < labeled {
			if group == 0 {
				o.State = "no_answer"
				o.Candidates[0] = behaviorprobe.CandidateTruth{CandidateID: "synthetic", State: "rejected", VectorsChecked: 1, FailedVectors: 1}
			} else {
				o.State, o.Acceptable = "known", []int{0}
				o.Candidates[0] = behaviorprobe.CandidateTruth{CandidateID: "synthetic", State: "acceptable", VectorsChecked: 1}
			}
		}
		groups.Groups[group].Parents = append(groups.Groups[group].Parents, id)
		if i < 48 {
			legacy.Outcomes = append(legacy.Outcomes, o)
		} else {
			typed.Outcomes = append(typed.Outcomes, o)
		}
	}
	return p, legacy, typed, groups
}

func TestSparseGroupsKnownNoAnswerAndUnknownOnlyGate(t *testing.T) {
	p, legacy, typed, groups := syntheticReports(14)
	r, e := buildReport(p, hash([]byte("synthetic plan")), legacy, typed, groups)
	if e != nil || r.Parents != 72 || r.Candidates != 72 || r.SourceVectorChecks != 16 || !r.WholeGroupGatePass || r.LabeledGroupGatePass || r.LabeledConnectedGroups != 14 || !slices.Contains(r.StopReasons, "insufficient_labeled_connected_groups") {
		t.Fatalf("sparse/known/no-answer/unknown-only distinction failed: %v", e)
	}
	p, legacy, typed, groups = syntheticReports(15)
	r, e = buildReport(p, hash([]byte("synthetic plan")), legacy, typed, groups)
	if e != nil || !r.WholeGroupGatePass || !r.LabeledGroupGatePass || r.TrainingExecutionReady || r.RolesAssigned || r.FinalEligible || r.ProtectedFinalRead || r.ProductionActivation || r.Fits != 0 || r.Partitions != 0 || r.Weights != 0 || r.ModelCalls != 0 || r.BaselineRankings != 0 || r.PerformanceRuns != 0 || !slices.Contains(r.StopReasons, "no_roles_plan") || !slices.Contains(r.StopReasons, "synthetic_single_pipeline") {
		t.Fatal("passing operational minima must never grant eligibility or execution")
	}
}

func TestReportRejectsForgedEligibilityCountsAndGroupMembership(t *testing.T) {
	tests := []struct {
		name string
		edit func(*auditPlan, *behaviorprobe.Report, *typedbehavior.Report, *typedbehavior.GroupReport)
	}{
		{"plan fits", func(p *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			p.Fits = 1
		}},
		{"legacy final", func(_ *auditPlan, r *behaviorprobe.Report, _ *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.FinalEligible = true
		}},
		{"typed ready", func(_ *auditPlan, _ *behaviorprobe.Report, r *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.TrainingReady = true
		}},
		{"typed partition", func(_ *auditPlan, _ *behaviorprobe.Report, r *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.Partitioned = true
		}},
		{"group ready", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.TrainingReady = true
		}},
		{"labeled inflation", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.LabeledConnectedGroups = 16
		}},
		{"whole count", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.ConnectedGroups = 16
		}},
		{"duplicate group id", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.Groups[1].ID = r.Groups[0].ID
		}},
		{"group id range", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.Groups[0].ID = 72
		}},
		{"missing member", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.Groups[0].Parents = r.Groups[0].Parents[1:]
		}},
		{"duplicate member", func(_ *auditPlan, _ *behaviorprobe.Report, _ *typedbehavior.Report, r *typedbehavior.GroupReport) {
			r.Groups[1].Parents[0] = r.Groups[0].Parents[0]
		}},
		{"unknown answer", func(_ *auditPlan, r *behaviorprobe.Report, _ *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.Outcomes[0].State = "unknown"
			r.Outcomes[0].Acceptable = []int{0}
		}},
		{"wrong control pass", func(_ *auditPlan, r *behaviorprobe.Report, _ *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.Controls[1].FailedVectors = 0
		}},
		{"correct control fail", func(_ *auditPlan, _ *behaviorprobe.Report, r *typedbehavior.Report, _ *typedbehavior.GroupReport) {
			r.Controls[0].FailedVectors = 1
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, legacy, typed, groups := syntheticReports(15)
			tt.edit(&p, &legacy, &typed, &groups)
			if _, e := buildReport(p, hash([]byte("synthetic plan")), legacy, typed, groups); e != errReport {
				t.Fatal("forged authority or group evidence reached a public report")
			}
		})
	}
}
