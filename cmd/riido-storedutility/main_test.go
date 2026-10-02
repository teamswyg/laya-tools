package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
)

// All tests below exercise protocol/provenance gates or tiny metadata only.
// None calls evaluate, shortclaim ranking, storedaudit.Bind or original APIs.
func TestCanonicalJSONRejectsAmbiguousOrLossyValues(t *testing.T) {
	type value struct {
		Number uint64 `json:"number"`
		Text   string `json:"text"`
	}
	valid := value{Number: ^uint64(0), Text: "public"}
	raw, err := canonical(valid)
	if err != nil {
		t.Fatal(err)
	}
	var got value
	if err := decodeCanonical(raw, &got); err != nil || got != valid {
		t.Fatal("exact integer/byte roundtrip failed", err)
	}
	for _, malformed := range [][]byte{
		[]byte("{\n  \"number\": 1,\n  \"number\": 18446744073709551615,\n  \"text\": \"public\"\n}\n"),
		[]byte("{\n  \"Number\": 18446744073709551615,\n  \"text\": \"public\"\n}\n"),
		[]byte("{\n  \"number\": 18446744073709551616,\n  \"text\": \"public\"\n}\n"),
		[]byte("{\n  \"number\": 18446744073709551615,\n  \"text\": \"public\",\n  \"unknown\": true\n}\n"),
		append(slices.Clone(raw), []byte("{}\n")...),
		[]byte("null\n"),
		[]byte("{\"number\":18446744073709551615,\"text\":\"public\"}\n"),
		append([]byte("{\n  \"number\": 18446744073709551615,\n  \"text\": \""), 0xff, '"', '\n', '}', '\n'),
		bytes.Repeat([]byte(" "), maxBytes+1),
	} {
		var decoded value
		if err := decodeCanonical(malformed, &decoded); err == nil {
			t.Fatal("ambiguous/noncanonical input accepted")
		}
	}
	if _, err := canonical(math.NaN()); err == nil {
		t.Fatal("nonfinite result encoded")
	}
}

func TestExclusiveOutputPreservesPartialRecordAndBounds(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "results.json")
	f, err := openNew(file)
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic partial counters test persistence, never the actual evaluator.
	partial := record{Schema: "toy_partial", Evaluation: report{State: "incomplete", DispatchAttempts: 1, CompletedRows: 0}}
	raw, err := canonical(partial)
	if err != nil || finishNew(f, raw) != nil {
		t.Fatal("partial metadata was not preserved", err)
	}
	if err := writeNew(file, []byte("replacement")); err == nil {
		t.Fatal("existing partial result was replaced")
	}
	got, err := os.ReadFile(file)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatal("partial record bytes changed")
	}
	if err := os.Mkdir(dir, 0700); err == nil {
		t.Fatal("an existing output directory was treated as new")
	}
	large := filepath.Join(dir, "oversized.json")
	if err := writeNew(large, bytes.Repeat([]byte{'x'}, maxBytes+1)); err == nil {
		t.Fatal("oversized output was allowed")
	}
	if info, err := os.Stat(large); err != nil || info.Size() != 0 {
		t.Fatal("failed output was not reserved without content")
	}
	if _, err := openNew(large); err == nil {
		t.Fatal("a failed reserved output was reused")
	}
}

func TestFailureDiagnosticsKeepDistinctPartialCountersWithoutText(t *testing.T) {
	r := report{State: "private-state-marker", BestBaseline: "private-baseline-marker", DispatchAttempts: 5, ValidatedRequests: 4, BaselineCalls: 3, RankingOutputs: 8, CompletedRows: 2}
	for _, encoding := range []bool{true, false} {
		text := resultFailure(encoding, r).Error()
		want := "storedutility_result_write_failed_after_evaluation"
		if encoding {
			want = "storedutility_result_encoding_failed_after_evaluation"
		}
		want += ":dispatch_attempts=5,validated_requests=4,baseline_calls=3,ranking_outputs=8,completed_rows=2"
		if text != want || strings.Contains(text, "private-") {
			t.Fatal("partial counters lost or report text leaked")
		}
	}
}

func TestBoundedRootRefusesPrivateEscapeAndNonRegularInputs(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := os.WriteFile(filepath.Join(dir, "public.json"), []byte("public"), 0600); err != nil {
		t.Fatal(err)
	}
	if raw, err := bounded(root, "public.json"); err != nil || string(raw) != "public" {
		t.Fatal("regular bounded input failed", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "large.json"), make([]byte, maxBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "directory"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../secret", "/secret", "a/../public.json", "a\\private", "c:private", "./public.json", "directory", "large.json"} {
		if _, err := bounded(root, path); err == nil {
			t.Fatal("escaped/unbounded/nonregular input accepted")
		} else if strings.Contains(err.Error(), path) && len(path) > 3 {
			t.Fatal("input path appeared in diagnostic")
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("private-test-marker"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "escape.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := bounded(root, "escape.json"); err == nil || strings.Contains(err.Error(), "private-test-marker") {
		t.Fatal("root escape was not safely rejected")
	}
}

func TestExtraNonTestGoSourceRejected(t *testing.T) {
	dir := t.TempDir()
	for _, path := range sourceDirs() {
		if err := os.MkdirAll(filepath.Join(dir, path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := verifyGoClosure(root, nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg/shortclaim/extra_test.go"), []byte("package shortclaim"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyGoClosure(root, nil); err != nil {
		t.Fatal("test metadata was treated as compiled runtime source")
	}
	extra := filepath.Join(dir, "pkg/shortclaim/extra.go")
	if err := os.WriteFile(extra, []byte("package shortclaim"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyGoClosure(root, nil); err == nil {
		t.Fatal("unlisted executable source accepted")
	}
	if err := verifyGoClosure(root, []artifact{{Path: "pkg/shortclaim/extra.go"}}); err != nil {
		t.Fatal("explicit toy source was not recognized", err)
	}
}

func TestCompiledSourceClosureRejectsStaleDiskBeforeDispatch(t *testing.T) {
	paths := []string{
		"cmd/riido-storedutility/main.go", "cmd/riido-storedutility/evaluator.go",
		"internal/storedaudit/binding.go", "internal/storedaudit/binding_provenance.go",
		"pkg/shortclaim/input.go", "pkg/shortclaim/baseline.go", "pkg/shortclaim/audit_provenance.go",
		"internal/lexicalhint/features.go", "internal/lexicalhint/audit_provenance.go",
	}
	dir := t.TempDir()
	for _, path := range paths {
		raw, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil {
			t.Fatal(err)
		}
		to := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(to, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	pins, err := sourceArtifacts(root)
	if err != nil || len(pins) != 9 {
		t.Fatal("exact compiled source closure rejected", err)
	}
	if err := verifyGoClosure(root, pins); err != nil {
		t.Fatal(err)
	}
	for _, pin := range pins {
		raw, err := bounded(root, pin.Path)
		if err != nil || pin.Bytes != len(raw) || pin.SHA256 != sha(raw) {
			t.Fatal("compiled byte identity was lost")
		}
		if err := os.WriteFile(filepath.Join(dir, pin.Path), append(slices.Clone(raw), []byte("\n// stale test source\n")...), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := sourceArtifacts(root); err == nil {
			t.Fatal("stale compiled dependency was accepted")
		}
		if err := os.WriteFile(filepath.Join(dir, pin.Path), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGitPreflightRejectsInvalidPinsBeforeStartingGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "public.txt"), []byte("public"), 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	// No git executable is reachable. Every case must fail its local validation.
	t.Setenv("PATH", t.TempDir())
	for _, pin := range []artifact{
		{Path: "../escape", Bytes: 6, SHA256: sha([]byte("public"))},
		{Path: "public.txt", Bytes: 6, SHA256: strings.Repeat("A", 64)},
		{Path: "public.txt", Bytes: -1, SHA256: sha([]byte("public"))},
		{Path: "public.txt", Bytes: maxBytes + 1, SHA256: sha([]byte("public"))},
		{Path: "public.txt", Bytes: 7, SHA256: sha([]byte("public"))},
		{Path: "public.txt", Bytes: 6, SHA256: sha([]byte("different"))},
	} {
		err := verifyBlobs(dir, root, strings.Repeat("a", 40), []artifact{pin})
		if err == nil || (err.Error() != "storedutility_pin_invalid" && err.Error() != "storedutility_disk_pin_mismatch") {
			t.Fatal("invalid pin reached Git or was accepted", err)
		}
	}
	if err := verifyBlobs(dir, root, strings.Repeat("A", 40), nil); err == nil || err.Error() != "storedutility_commit_invalid" {
		t.Fatal("noncanonical revision accepted")
	}
}

func TestBoundedGitBufferDoesNotAcceptOversizedChunk(t *testing.T) {
	var b boundedBuffer
	if n, err := b.Write([]byte("toy")); err != nil || n != 3 {
		t.Fatal(err)
	}
	if n, err := b.Write(make([]byte, maxBytes)); err == nil || n != 0 || string(b.raw) != "toy" {
		t.Fatal("oversized blob partially entered buffer")
	}
}

func toyPlan() plan {
	binary := artifact{SHA256: strings.Repeat("1", 64), Bytes: 1024}
	inputs := []artifact{{Path: "public/input.json", SHA256: strings.Repeat("2", 64), Bytes: 7}}
	sources := []artifact{{Path: "public/source.go", SHA256: strings.Repeat("3", 64), Bytes: 13}}
	support := []artifact{{Path: "public/plan.md", SHA256: strings.Repeat("4", 64), Bytes: 17}}
	return newPlan(strings.Repeat("a", 40), binary, inputs, sources, support)
}

func clonePlan(t *testing.T, value plan) plan {
	t.Helper()
	raw, err := canonical(value)
	if err != nil {
		t.Fatal(err)
	}
	var out plan
	if err := decodeCanonical(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFrozenPlanRejectsPrivilegeCountAndArtifactDrift(t *testing.T) {
	want := toyPlan()
	if err := validatePlan(want, want); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*plan){
		func(p *plan) { p.Schema = "other" },
		func(p *plan) { p.SourceCommit = strings.Repeat("B", 40) },
		func(p *plan) { p.BinaryBytes++ },
		func(p *plan) { p.BinarySHA256 = strings.Repeat("5", 64) },
		func(p *plan) { p.Inputs[0].SHA256 = strings.Repeat("6", 64) },
		func(p *plan) { p.Sources = append(p.Sources, p.Sources[0]) },
		func(p *plan) { p.Support = nil },
		func(p *plan) { p.Parents++ },
		func(p *plan) { p.CandidateCaptions++ },
		func(p *plan) { p.Answerable++ },
		func(p *plan) { p.NoAnswer++ },
		func(p *plan) { p.Unknown-- },
		func(p *plan) { p.ConnectedGroups++ },
		func(p *plan) { p.LabeledGroups++ },
		func(p *plan) { p.Baselines[0], p.Baselines[1] = p.Baselines[1], p.Baselines[0] },
		func(p *plan) { p.RequiredRelativeGain = .01 },
		func(p *plan) { p.OfficialAttempts = 2 },
		func(p *plan) { p.Retries = 1 },
		func(p *plan) { p.RankingBudgetSeconds = 11 },
		func(p *plan) { p.GitBlobBudgetSeconds = 6 },
		func(p *plan) { p.CPUThreads = 2 },
		func(p *plan) { p.HeapSoftLimit++ },
		func(p *plan) { p.Roles = 1 },
		func(p *plan) { p.Fits = 1 },
		func(p *plan) { p.ModelCalls = 1 },
		func(p *plan) { p.PaidCalls = 1 },
		func(p *plan) { p.FinalReads = 1 },
		func(p *plan) { p.Weights = 1 },
		func(p *plan) { p.Activation = true },
		func(p *plan) { p.ResourceMeasurement = true },
		func(p *plan) { p.GoVersion = "go1.27.0" },
		func(p *plan) { p.GOOS = "different" },
		func(p *plan) { p.GOARCH = "different" },
	} {
		got := clonePlan(t, want)
		change(&got)
		if err := validatePlan(got, want); err == nil {
			t.Fatal("frozen policy or provenance drift accepted")
		}
	}
}

func TestPlanDoesNotAliasArtifactInputsOrRegistry(t *testing.T) {
	inputs := []artifact{{Path: "toy.json", SHA256: strings.Repeat("1", 64), Bytes: 3}}
	sources := []artifact{{Path: "toy.go", SHA256: strings.Repeat("2", 64), Bytes: 4}}
	support := []artifact{{Path: "toy.md", SHA256: strings.Repeat("3", 64), Bytes: 5}}
	p := newPlan(strings.Repeat("a", 40), artifact{}, inputs, sources, support)
	inputs[0].Path, sources[0].Path, support[0].Path = "altered", "altered", "altered"
	p.Baselines[0] = "altered"
	if p.Inputs[0].Path != "toy.json" || p.Sources[0].Path != "toy.go" || p.Support[0].Path != "toy.md" || baselineKinds()[0] != "fixed_order" {
		t.Fatal("caller could mutate prepared provenance or global registry")
	}
}

func TestBuildAndRecipeRequirementsRejectUntestedExecution(t *testing.T) {
	valid := &debug.BuildInfo{GoVersion: "go1.27.1", Path: "github.com/teamswyg/laya-tools/cmd/riido-storedutility", Settings: []debug.BuildSetting{{Key: "-trimpath", Value: "true"}, {Key: "CGO_ENABLED", Value: "0"}}}
	if !validBuild(valid) {
		t.Fatal("fixed build metadata rejected")
	}
	for _, bad := range []*debug.BuildInfo{nil, {}, {GoVersion: "go1.27.0", Path: valid.Path, Settings: valid.Settings}, {GoVersion: valid.GoVersion, Path: "other", Settings: valid.Settings}, {GoVersion: valid.GoVersion, Path: valid.Path, Settings: []debug.BuildSetting{{Key: "CGO_ENABLED", Value: "0"}}}, {GoVersion: valid.GoVersion, Path: valid.Path, Settings: []debug.BuildSetting{{Key: "-trimpath", Value: "true"}, {Key: "CGO_ENABLED", Value: "1"}}}} {
		if validBuild(bad) {
			t.Fatal("unreviewed build metadata accepted")
		}
	}
	r := buildRecipe{Schema: recipeSchema, GoVersion: "go1.27.1", CGOEnabled: "0", Arguments: []string{"build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "<private_binary>", "./cmd/riido-storedutility"}}
	raw, err := canonical(r)
	if err != nil || validateRecipe(raw) != nil {
		t.Fatal("fixed recipe rejected", err)
	}
	r.Arguments[2] = "-buildvcs=true"
	raw, _ = canonical(r)
	if validateRecipe(raw) == nil {
		t.Fatal("different build recipe accepted")
	}
}

func TestFixedArgumentDiagnosticsNeverEchoPrivateValues(t *testing.T) {
	marker := "private-argument-marker"
	for _, args := range [][]string{
		{marker},
		{"plan", "--unknown=" + marker},
		{"plan", "--out=" + marker, "--source-commit=" + marker},
		{"audit", "--out=" + marker, "--input-commit=" + marker},
		{"audit", "--out=" + marker, "--input-commit=" + strings.Repeat("a", 40), marker},
	} {
		if err := execute(args); err == nil || strings.Contains(err.Error(), marker) {
			t.Fatal("private argument leaked or invalid invocation accepted")
		}
	}
}

func TestCanonicalFixedArrayOverflowCannotBeLaundered(t *testing.T) {
	type fixed struct {
		Items [2]int `json:"items"`
	}
	raw := []byte("{\n  \"items\": [\n    1,\n    2,\n    3\n  ]\n}\n")
	var got fixed
	if err := decodeCanonical(raw, &got); err == nil {
		t.Fatal("discarded fixed-array element was laundered")
	}
	// Confirm the control targets encoding/json's actual truncation behavior.
	if err := json.Unmarshal(raw, &got); err != nil || got.Items != [2]int{1, 2} {
		t.Fatal("fixed-array control no longer reproduces silent truncation", err)
	}
}
