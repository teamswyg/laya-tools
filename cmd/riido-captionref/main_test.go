// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/captionref"
	"github.com/teamswyg/laya-tools/internal/storedaudit"
)

// Tests use owned toy files and protocol metadata only. None invokes Generate,
// storedaudit.Bind, original source APIs, ranking, models or actual raw inputs.
func TestCanonicalRejectsAmbiguousAndLossyJSON(t *testing.T) {
	type value struct {
		Number uint64 `json:"number"`
		Text   string `json:"text"`
	}
	want := value{^uint64(0), "<public>"}
	raw, err := canonical(want)
	if err != nil || !bytes.Contains(raw, []byte(`\u003cpublic\u003e`)) {
		t.Fatal("canonical escaped bytes unavailable", err)
	}
	var got value
	if err := decodeCanonical(raw, &got); err != nil || got != want {
		t.Fatal("exact integer/text roundtrip failed", err)
	}
	for _, malformed := range [][]byte{
		[]byte("{\n  \"number\": 1,\n  \"number\": 18446744073709551615,\n  \"text\": \"\\u003cpublic\\u003e\"\n}\n"),
		[]byte("{\n  \"Number\": 18446744073709551615,\n  \"text\": \"\\u003cpublic\\u003e\"\n}\n"),
		[]byte("{\n  \"number\": 18446744073709551616,\n  \"text\": \"public\"\n}\n"),
		[]byte("{\n  \"number\": 1,\n  \"text\": \"public\",\n  \"extra\": true\n}\n"),
		append(slices.Clone(raw), []byte("{}\n")...),
		[]byte("null\n"),
		bytes.ReplaceAll(raw, []byte(`\u003cpublic\u003e`), []byte("<public>")),
		append([]byte("{\n  \"number\": 1,\n  \"text\": \""), 0xff, '"', '\n', '}', '\n'),
		bytes.Repeat([]byte{' '}, maxBytes+1),
	} {
		var v value
		if err := decodeCanonical(malformed, &v); err == nil {
			t.Fatal("noncanonical/ambiguous/lossy JSON accepted")
		}
	}
	if _, err := canonical(math.NaN()); err == nil {
		t.Fatal("nonfinite JSON accepted")
	}
}

func TestBuildRecipeContract(t *testing.T) {
	want := buildRecipe{recipeSchema, "go1.27.1", "0", []string{"build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "<private_binary>", "./cmd/riido-captionref"}}
	raw, err := canonical(want)
	if err != nil || validateRecipe(raw) != nil {
		t.Fatal("owned correct recipe refused", err)
	}
	bad := want
	bad.Arguments = slices.Clone(want.Arguments)
	bad.Arguments[1] = "-race"
	raw, _ = canonical(bad)
	if validateRecipe(raw) == nil {
		t.Fatal("changed build recipe accepted")
	}
	raw, _ = canonical(want)
	if validateRecipe(bytes.ReplaceAll(raw, []byte(`\u003cprivate_binary\u003e`), []byte("<private_binary>"))) == nil {
		t.Fatal("unescaped recipe bypassed canonical contract")
	}
}

func TestCheckedInBuildRecipeIsCanonical(t *testing.T) {
	raw, err := os.ReadFile("../../experiments/short-claim/build-recipe-59.json")
	if err != nil || validateRecipe(raw) != nil {
		t.Fatal("checked-in build recipe failed its runtime contract", err)
	}
}

func TestBoundedRootRejectsEscapesNonregularAndOversize(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := os.WriteFile(filepath.Join(dir, "toy.json"), []byte("owned toy"), 0600); err != nil {
		t.Fatal(err)
	}
	if raw, err := bounded(root, "toy.json"); err != nil || string(raw) != "owned toy" {
		t.Fatal("owned regular file refused", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "oversize.json"), make([]byte, maxBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "directory"), 0700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("private-marker"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "escape.json")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../secret", "/secret", "a/../toy.json", "./toy.json", "a\\private", "c:private", ".", "directory", "oversize.json", "escape.json"} {
		if _, err := bounded(root, path); err == nil {
			t.Fatalf("unsafe/unbounded input accepted: %q", path)
		} else if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), path) {
			t.Fatal("input path/content escaped diagnostics")
		}
	}
}

func TestInputPinMismatchBeforeBinding(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	path := captionref.InputPins()[0].Path
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, path), []byte("owned incorrect toy pin"), 0600); err != nil {
		t.Fatal(err)
	}
	in, pins, err := readInputs(root)
	if err == nil || err.Error() != "captionref_input_pin_mismatch" || pins != nil {
		t.Fatal("incorrect owned pin was accepted", err)
	}
	for _, raw := range append(in.Stored[:], in.Literals[:]...) {
		if raw != nil {
			t.Fatal("failed input read exposed partial inputs")
		}
	}
}

func TestPinValidationRejectsDuplicatesAndInvalidValues(t *testing.T) {
	want := artifact{Path: "toy/a.json", Bytes: 3, SHA256: strings.Repeat("a", 64)}
	if validateArtifacts([]artifact{want}) != nil {
		t.Fatal("valid toy pin refused")
	}
	for _, edit := range []func(*artifact){
		func(a *artifact) { a.Path = "../private" },
		func(a *artifact) { a.SHA256 = strings.Repeat("A", 64) },
		func(a *artifact) { a.Bytes = -1 },
		func(a *artifact) { a.Bytes = maxBytes + 1 },
	} {
		bad := want
		edit(&bad)
		if validateArtifacts([]artifact{bad}) == nil {
			t.Fatal("invalid toy pin accepted")
		}
	}
	if validateArtifacts([]artifact{want, want}) == nil {
		t.Fatal("duplicate pin accepted")
	}
}

func TestInvalidPinAndCommitNeverReachGit(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, commit := range []string{"main", "abc", strings.Repeat("A", 40)} {
		if err := verifyBlobs(dir, root, commit, nil); err == nil || err.Error() != "captionref_commit_invalid" {
			t.Fatal("invalid revision reached Git", err)
		}
	}
	bad := artifact{Path: "../private", Bytes: 1, SHA256: strings.Repeat("a", 64)}
	if err := verifyBlobs(dir, root, strings.Repeat("a", 40), []artifact{bad}); err == nil || err.Error() != "captionref_pin_invalid" {
		t.Fatal("invalid pin reached Git", err)
	}
	valid := artifact{Path: "toy.json", Bytes: 1, SHA256: strings.Repeat("a", 64)}
	if err := verifyBlobs(dir, root, strings.Repeat("a", 40), []artifact{valid}); err == nil || err.Error() != "captionref_disk_pin_mismatch" {
		t.Fatal("missing toy file reached Git", err)
	}
}

func TestExclusiveResultPreservesPartialEvidenceAndRefusesReuse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attempt", "results.json")
	f, err := reserveResult(path)
	if err != nil {
		t.Fatal(err)
	}
	r := record{Schema: "toy-partial", State: "incomplete", FailureCode: "captionref_generation_failed", FailureCause: string(captionref.ErrAST), Counters: captionref.Counters{BindAttempts: 1, BindCompleted: 1, ASTParseAttempts: 2, ASTParsesCompleted: 1, ParentsGenerated: 72, CaptionsGenerated: 216, ContractsGenerated: 12}}
	if err := persistResult(f, r); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got record
	if err := decodeCanonical(raw, &got); err != nil || got.Counters != r.Counters || got.State != "incomplete" || got.FailureCode != r.FailureCode || got.FailureCause != r.FailureCause {
		t.Fatal("partial counters/state lost", err)
	}
	if _, err := reserveResult(path); err == nil {
		t.Fatal("existing attempt directory reused")
	}
	if err := writeNew(path, []byte("replacement")); err == nil {
		t.Fatal("existing partial record replaced")
	}
	again, _ := os.ReadFile(path)
	if !bytes.Equal(raw, again) {
		t.Fatal("stored partial bytes changed")
	}
	if _, err := reserveResult("results.json"); err == nil {
		t.Fatal("unreserved current directory permitted")
	}
	if _, err := reserveResult(filepath.Join(t.TempDir(), "other.json")); err == nil {
		t.Fatal("non-result filename permitted")
	}
}

type toyWriter struct {
	writeErr error
	closeErr error
	short    bool
	closed   bool
}

func (f *toyWriter) Write(raw []byte) (int, error) {
	if f.short {
		return len(raw) - 1, nil
	}
	return len(raw), f.writeErr
}
func (f *toyWriter) Close() error { f.closed = true; return f.closeErr }

func TestWriteAndEncodingFailuresRetainAllCountersWithoutText(t *testing.T) {
	c := captionref.Counters{BindAttempts: 2, BindCompleted: 1, ASTParseAttempts: 3, ASTParsesCompleted: 2, ParentsGenerated: 4, CaptionsGenerated: 5, ContractsGenerated: 6}
	suffix := ":metadata_bind_attempts=2,metadata_bind_completed=1,ast_parse_attempts=3,ast_parses_completed=2,parents_generated=4,captions_generated=5,contracts_generated=6"
	r := record{State: "private-state-marker", Counters: c, Limitations: []string{"private-text-marker"}}
	for _, w := range []*toyWriter{{short: true}, {writeErr: errors.New("private-writer-marker")}, {closeErr: errors.New("private-close-marker")}} {
		err := persistResult(w, r)
		if err == nil || err.Error() != "captionref_result_write_failed_after_generation"+suffix || !w.closed || strings.Contains(err.Error(), "private") {
			t.Fatal("write failure leaked text or lost counters", err)
		}
	}
	w := &toyWriter{}
	r.Limitations = []string{strings.Repeat("x", maxBytes+1)}
	if err := persistResult(w, r); err == nil || err.Error() != "captionref_result_encoding_failed_after_generation"+suffix || !w.closed {
		t.Fatal("encoding failure lost counters or writer", err)
	}
}

func TestGenerationFailureCausePreservesSafeEnumsOnly(t *testing.T) {
	if generationFailureCause(nil) != "" {
		t.Fatal("nil error acquired failure cause")
	}
	for _, err := range []error{captionref.ErrAST, captionref.ErrMetadata, storedaudit.ErrTruth, storedaudit.ErrGroups} {
		if got := generationFailureCause(err); got != err.Error() {
			t.Fatal("fixed error cause lost", got)
		}
	}
	if got := generationFailureCause(errors.New("private-parser-path-marker")); got != "captionref_generation_error_unclassified" {
		t.Fatal("unclassified diagnostics leaked", got)
	}
}

func TestOversizedOutputLeavesExclusiveReservation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversize.json")
	if err := writeNew(path, make([]byte, maxBytes+1)); err == nil {
		t.Fatal("oversized output allowed")
	}
	if info, err := os.Stat(path); err != nil || info.Size() != 0 {
		t.Fatal("failed output reservation missing", err)
	}
	if _, err := openNew(path); err == nil {
		t.Fatal("failed reserved file reused")
	}
}

func toyPlan() plan {
	makePins := func(prefix string, count int) []artifact {
		out := make([]artifact, count)
		for i := range out {
			out[i] = artifact{Path: prefix + "/" + string(rune('a'+i)) + ".json", Bytes: i + 1, SHA256: strings.Repeat("a", 64)}
		}
		return out
	}
	return newPlan(strings.Repeat("b", 40), artifact{Bytes: 1024, SHA256: strings.Repeat("c", 64)}, makePins("input", 6), makePins("source", 5), makePins("support", 9))
}

func TestPlanIsCanonicalAndFrozenWithoutGenerating(t *testing.T) {
	want := toyPlan()
	raw, err := canonical(want)
	if err != nil {
		t.Fatal(err)
	}
	var got plan
	if err := decodeCanonical(raw, &got); err != nil || validatePlan(got, want) != nil {
		t.Fatal("canonical toy plan refused", err)
	}
	for _, edit := range []func(*plan){
		func(p *plan) { p.Schema = "different" },
		func(p *plan) { p.SourceCommit = strings.Repeat("A", 40) },
		func(p *plan) { p.BinaryBytes = maxBinary + 1 },
		func(p *plan) { p.GoVersion = "go1.26.0" },
		func(p *plan) { p.CGOEnabled = "1" },
		func(p *plan) { p.TrimPath = false },
		func(p *plan) { p.OfficialAttempts = 2 },
		func(p *plan) { p.Retries = 1 },
		func(p *plan) { p.ContentReview = "approved" },
		func(p *plan) { p.LiteralReified = true },
		func(p *plan) { p.TrainingReady = true },
		func(p *plan) { p.NewLabels = 1 },
		func(p *plan) { p.Rankings = 1 },
		func(p *plan) { p.SourceCalls = 1 },
		func(p *plan) { p.Roles = 1 },
		func(p *plan) { p.Fits = 1 },
		func(p *plan) { p.HeapSoftLimit++ },
		func(p *plan) { p.Inputs = p.Inputs[:5] },
		func(p *plan) { p.Sources = p.Sources[:4] },
		func(p *plan) { p.Support = p.Support[:8] },
	} {
		bad := want
		edit(&bad)
		if validatePlan(bad, want) == nil {
			t.Fatal("changed plan accepted")
		}
	}
}

func TestCompiledSourceMetadataClosedToFiveFiles(t *testing.T) {
	sources := sourceArtifacts()
	if len(sources) != 5 || validateArtifacts(sources) != nil {
		t.Fatal("compiled source metadata mismatch")
	}
	dir := t.TempDir()
	for _, a := range sources {
		path := filepath.Join(dir, a.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("owned toy package"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := verifyGoClosure(root, sources); err != nil {
		t.Fatal("closed toy source list refused", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/captionref/extra.go"), []byte("owned extra toy"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyGoClosure(root, sources); err == nil || err.Error() != "captionref_unlisted_go_source" {
		t.Fatal("unlisted source accepted", err)
	}
}

func TestBuildSettingsArePinned(t *testing.T) {
	want := debug.BuildInfo{GoVersion: "go1.27.1", Path: "github.com/teamswyg/laya-tools/cmd/riido-captionref", Settings: []debug.BuildSetting{{Key: "-trimpath", Value: "true"}, {Key: "CGO_ENABLED", Value: "0"}}}
	if !validBuild(&want) || validBuild(nil) {
		t.Fatal("build metadata contract mismatch")
	}
	for _, edit := range []func(*debug.BuildInfo){
		func(b *debug.BuildInfo) { b.GoVersion = "go1.26.0" },
		func(b *debug.BuildInfo) { b.Path = "another-command" },
		func(b *debug.BuildInfo) { b.Settings = nil },
		func(b *debug.BuildInfo) {
			b.Settings = []debug.BuildSetting{{Key: "-trimpath", Value: "false"}, {Key: "CGO_ENABLED", Value: "0"}}
		},
		func(b *debug.BuildInfo) {
			b.Settings = []debug.BuildSetting{{Key: "-trimpath", Value: "true"}, {Key: "CGO_ENABLED", Value: "1"}}
		},
	} {
		bad := want
		edit(&bad)
		if validBuild(&bad) {
			t.Fatal("changed build metadata accepted")
		}
	}
}

func TestBoundedGitOutputStopsBeforeGrowing(t *testing.T) {
	var b boundedBuffer
	if n, err := b.Write([]byte("toy")); n != 3 || err != nil {
		t.Fatal("toy bounded output refused")
	}
	if n, err := b.Write(make([]byte, maxBytes)); n != 0 || err == nil || string(b.raw) != "toy" {
		t.Fatal("oversized output grew buffer")
	}
}
