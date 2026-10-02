// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/internal/sourceinventory"
)

// All test files and metadata are toy text. No real source, saved results or
// original Rebind is opened. The successful engine below is explicitly a fake.
type toyRepo struct {
	dir             string
	deps            dependencies
	calls, gitCalls int
}

func put(t *testing.T, dir, path string, raw []byte) {
	t.Helper()
	p := filepath.Join(dir, path)
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, raw, 0600); e != nil {
		t.Fatal(e)
	}
}

func fixture(t *testing.T) *toyRepo {
	t.Helper()
	r := &toyRepo{dir: t.TempDir()}
	if e := os.Mkdir(filepath.Join(r.dir, ".cache"), 0700); e != nil {
		t.Fatal(e)
	}
	r.deps.Pins = inputPins()
	for i, p := range r.deps.Pins {
		raw := []byte("package typedbehavior\n")
		if i == 0 {
			raw = []byte("package behaviorprobe\n")
		}
		if i == 4 {
			raw = toyMetadata(t)
		}
		put(t, r.dir, p.Path, raw)
		r.deps.Pins[i] = artifact{Path: p.Path, SHA256: sha(raw), Bytes: len(raw)}
	}
	for _, p := range compiledPaths() {
		raw := []byte("// toy runtime source: " + p + "\npackage toy\n")
		put(t, r.dir, p, raw)
		r.deps.Sources = append(r.deps.Sources, artifact{Path: p, SHA256: sha(raw), Bytes: len(raw)})
	}
	for _, p := range supportPaths() {
		raw := []byte("toy support\n")
		if p == buildRecipePath {
			var e error
			raw, e = canonical(fixedRecipe())
			if e != nil {
				t.Fatal(e)
			}
		}
		put(t, r.dir, p, raw)
	}
	r.deps.CompileBindingVerified = true
	r.deps.Binary = func() (artifact, error) { return artifact{SHA256: sha([]byte("toy binary")), Bytes: 10}, nil }
	r.deps.Git = func(ctx context.Context, repo, commit, path string) ([]byte, error) {
		r.gitCalls++
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 5*time.Second || time.Until(deadline) <= 0 {
			t.Fatal("missing per-blob5s budget")
		}
		if repo != r.dir || !hash(commit, 40) {
			t.Fatal("wrong Git binding")
		}
		return os.ReadFile(filepath.Join(repo, path))
	}
	r.deps.Rebind = func(in sourceinventory.Input) (sourceinventory.Report, error) {
		r.calls++
		if len(in.Roots) != 60 || len(in.Files) != 4 {
			t.Fatal("wrong toy projection")
		}
		return toyReport(in), nil
	}
	r.deps.WriteResult = finish
	return r
}

func prepare(t *testing.T, r *toyRepo) plan {
	t.Helper()
	e := runWith([]string{"prepare", "--repo", r.dir, "--source-commit", strings.Repeat("a", 40), "--out", ".cache/toy-plan.json"}, r.deps)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(r.dir, ".cache/toy-plan.json"))
	if e != nil {
		t.Fatal(e)
	}
	var p plan
	if e := decodeCanonical(raw, &p); e != nil {
		t.Fatal(e)
	}
	put(t, r.dir, planPath, raw)
	return p
}

func execute(r *toyRepo, out string) error {
	return runWith([]string{"inventory", "--repo", r.dir, "--input-commit", strings.Repeat("b", 40), "--out", out}, r.deps)
}
func readRecord(t *testing.T, r *toyRepo, out string) record {
	t.Helper()
	raw, e := os.ReadFile(filepath.Join(r.dir, out))
	if e != nil {
		t.Fatal(e)
	}
	var got record
	if e := decodeCanonical(raw, &got); e != nil {
		t.Fatal(e)
	}
	return got
}

func TestCompileBindingGateStopsBeforeInputs(t *testing.T) {
	d := productionDependencies()
	if !d.CompileBindingVerified || len(d.Sources) != 5 || !artifactsValid(d.Sources) {
		t.Fatal("compiled source enumeration missing")
	}
	d.CompileBindingVerified = false
	e := runWith([]string{"prepare", "--repo", "does-not-exist", "--source-commit", strings.Repeat("a", 40), "--out", ".cache/plan.json"}, d)
	if e == nil || e.Error() != "inventory60_compile_binding_unverified" {
		t.Fatal("unverified binding attempted filesystem or original operations", e)
	}
}

func TestToyPrepareOnlyHashesAndCanonicalPlan(t *testing.T) {
	r := fixture(t)
	// Prove prepare does not even project metadata: hash-pinned non-JSON is
	// acceptable for prepare, but inventory rejects it before reserving output.
	p := r.deps.Pins[4]
	raw := []byte("unparsed toy metadata\n")
	put(t, r.dir, p.Path, raw)
	r.deps.Pins[4] = artifact{Path: p.Path, SHA256: sha(raw), Bytes: len(raw)}
	got := prepare(t, r)
	if r.calls != 0 || r.gitCalls != 26 || got.Expected != expected() || got.Retries != 0 || got.Attempts != 1 || got.CompileBinding != "compiled_package_source_bytes_verified_ast_only" {
		t.Fatal("prepare did work or changed forecasts")
	}
	if e := execute(r, ".cache/invalid/results.json"); e == nil {
		t.Fatal("invalid metadata accepted")
	}
	if _, e := os.Stat(filepath.Join(r.dir, ".cache/invalid")); !os.IsNotExist(e) || r.calls != 0 {
		t.Fatal("output reserved before metadata preflight")
	}
}

func TestToyReserveBeforeExactlyOneCallAndPersist(t *testing.T) {
	r := fixture(t)
	prepare(t, r)
	out := ".cache/success/results.json"
	original := r.deps.Rebind
	r.deps.Rebind = func(in sourceinventory.Input) (sourceinventory.Report, error) {
		if s, e := os.Stat(filepath.Join(r.dir, out)); e != nil || s.Size() != 0 {
			t.Fatal("not reserved before invocation")
		}
		return original(in)
	}
	if e := execute(r, out); e != nil {
		t.Fatal(e)
	}
	got := readRecord(t, r, out)
	if r.calls != 1 || got.Invocation != (invocation{1, 1, true}) || got.VerifiedBlobs != 27 || got.State != "metadata_rebound_content_review_pending" || got.Inventory.Counters.FormatCallsCompleted != 264 || got.Plan.Expected != expected() || got.ContentReview != "pending" || got.TrainingReady {
		t.Fatal("fake report acquired forecast counts or semantic eligibility")
	}
	if e := execute(r, out); e == nil || r.calls != 1 {
		t.Fatal("existing output triggered retry")
	}
	if e := execute(r, ".cache/success/another.json"); e == nil || r.calls != 1 {
		t.Fatal("alternate output in used directory")
	}
	mode, e := os.Stat(filepath.Join(r.dir, out))
	if e != nil || mode.Mode().Perm() != 0600 {
		t.Fatal("output permissions")
	}
}

// Protocol scaffolding only: these counters and references come from a fake
// engine, never from original inventory execution or injected by the runner.
func toyReport(in sourceinventory.Input) sourceinventory.Report {
	r := sourceinventory.Report{Schema: "riido-source-declaration-reference-draft-v1", State: "metadata_rebound_content_review_pending", ContentReview: "pending", ReusePolicy: "no_cache_each_relation_v1"}
	r.Counters = sourceinventory.Counters{
		FileHashAttempts: 4, FileHashesCompleted: 4, FilePinsMatched: 4,
		ParseAttempts: 4, ParsesCompleted: 4, ParsesSucceeded: 4, RegistryEntries: 60,
		RootAttempts: 60, RootsCompleted: 60, ComponentAttempts: 204, ComponentsCompleted: 204,
		RawSpanHashAttempts: 264, RawSpanHashesCompleted: 264,
		FormatCallsAttempted: 264, FormatCallsCompleted: 264,
		ScanCallsAttempted: 112, ScanCallsCompleted: 112, BundleCallsAttempted: 24, BundleCallsCompleted: 24,
	}
	for _, f := range in.Files {
		r.Files = append(r.Files, sourceinventory.FileReference{Path: f.Path, Bytes: len(f.Raw), SHA256: f.ExpectedSHA256})
	}
	for _, want := range in.Roots {
		root := sourceinventory.RootReference{ID: want.ID, Cohort: want.Cohort, MetadataPinsMatched: true, ContentReview: "pending", HistoricalBundleSHA256: want.BundleSHA256,
			Root: sourceinventory.DeclarationReference{FormattedSHA256: want.CodeSHA256, NormalizedSHA256: want.NormalizedCodeSHA256}}
		for _, c := range want.Components {
			root.Components = append(root.Components, sourceinventory.DeclarationReference{ID: c.ID, Kind: c.Kind, FormattedSHA256: c.SHA256, NormalizedSHA256: c.NormalizedBehaviorSHA256})
		}
		r.Roots = append(r.Roots, root)
	}
	return r
}

func TestReturnedScopeRejectsEmptyOrChangedEvidenceWithoutFillingCounters(t *testing.T) {
	roots, err := historicalRoots(toyMetadata(t))
	if err != nil {
		t.Fatal(err)
	}
	in := sourceinventory.Input{Roots: roots, GoVersion: "go1.27.1"}
	paths := inputPins()
	for i := range in.Files {
		in.Files[i] = sourceinventory.File{Path: paths[i].Path, Raw: []byte("toy"), ExpectedSHA256: sha([]byte("toy"))}
	}
	mutations := []struct {
		name   string
		change func(*sourceinventory.Report)
	}{
		{"empty", func(r *sourceinventory.Report) { r.Files = nil; r.Roots = nil; r.Counters = sourceinventory.Counters{} }},
		{"file_count", func(r *sourceinventory.Report) { r.Files = r.Files[:3] }},
		{"file_hash", func(r *sourceinventory.Report) { r.Files[0].SHA256 = sha([]byte("changed")) }},
		{"root_count", func(r *sourceinventory.Report) { r.Roots = r.Roots[:59] }},
		{"root_order", func(r *sourceinventory.Report) { r.Roots[0], r.Roots[1] = r.Roots[1], r.Roots[0] }},
		{"root_hash", func(r *sourceinventory.Report) { r.Roots[0].Root.FormattedSHA256 = sha([]byte("changed")) }},
		{"pins", func(r *sourceinventory.Report) { r.Roots[0].MetadataPinsMatched = false }},
		{"component", func(r *sourceinventory.Report) { r.Roots[36].Components[0].ID = "changed" }},
		{"bundle", func(r *sourceinventory.Report) { r.Roots[36].HistoricalBundleSHA256 = sha([]byte("changed")) }},
		{"counter", func(r *sourceinventory.Report) { r.Counters.FormatCallsCompleted = 53 }},
	}
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			got := toyReport(in)
			m.change(&got)
			r := record{State: "incomplete"}
			calls := 0
			invoke(in, &r, func(sourceinventory.Input) (sourceinventory.Report, error) { calls++; return got, nil })
			if calls != 1 || r.State != "incomplete" || r.FailureCause != "inventory60_report_scope_invalid" || !reflect.DeepEqual(r.Inventory, got) || r.Invocation != (invocation{1, 1, true}) {
				t.Fatal("invalid returned evidence promoted, changed or retried", r)
			}
		})
	}
}

func TestToyReturnedFailureRetainsPrefixAndSafeEnums(t *testing.T) {
	for _, known := range []error{sourceinventory.ErrInput, sourceinventory.ErrFilePin, sourceinventory.ErrParse, sourceinventory.ErrShape, sourceinventory.ErrRegistry, sourceinventory.ErrDeclaration, sourceinventory.ErrHistoricalPin, sourceinventory.ErrRecipe, errors.New("SECRET source /Users/owner/raw.go raw parser diagnostic")} {
		t.Run(safeCause(known), func(t *testing.T) {
			r := fixture(t)
			prepare(t, r)
			prefix := sourceinventory.Counters{FileHashAttempts: 2, FileHashesCompleted: 2, FilePinsMatched: 1, ParseAttempts: 1, ParsesCompleted: 1, ParsesSucceeded: 1}
			r.deps.Rebind = func(sourceinventory.Input) (sourceinventory.Report, error) {
				r.calls++
				return sourceinventory.Report{Schema: "riido-source-declaration-reference-draft-v1", State: "incomplete", Counters: prefix, ContentReview: "pending"}, known
			}
			out := ".cache/failure/results.json"
			e := execute(r, out)
			if e == nil {
				t.Fatal("failure succeeded")
			}
			got := readRecord(t, r, out)
			if r.calls != 1 || got.Inventory.Counters != prefix || got.Invocation != (invocation{1, 1, true}) || got.State != "incomplete" || got.FailureCause != safeCause(known) {
				t.Fatal("failure lost prefix or retry", got)
			}
			b, _ := canonical(got)
			if strings.Contains(string(b), "SECRET") || strings.Contains(e.Error(), "SECRET") {
				t.Fatal("raw error leaked")
			}
		})
	}
}

func TestToyPanicMarksInnerCountersUnavailable(t *testing.T) {
	r := fixture(t)
	prepare(t, r)
	r.deps.Rebind = func(sourceinventory.Input) (sourceinventory.Report, error) { r.calls++; panic("SECRET panic raw path") }
	out := ".cache/panic/results.json"
	if e := execute(r, out); e == nil || strings.Contains(e.Error(), "SECRET") {
		t.Fatal("panic was unsafely reported", e)
	}
	got := readRecord(t, r, out)
	if r.calls != 1 || got.Invocation != (invocation{1, 0, false}) || got.FailureCause != "inventory60_rebind_panic_unclassified" || got.State != "incomplete" {
		t.Fatal("panic fabricated returned counters", got.Invocation)
	}
}

func TestToyPersistenceFailureReturnsSafeActualCounters(t *testing.T) {
	r := fixture(t)
	prepare(t, r)
	prefix := sourceinventory.Counters{ParseAttempts: 2, ParsesCompleted: 2, ParsesSucceeded: 1}
	r.deps.Rebind = func(sourceinventory.Input) (sourceinventory.Report, error) {
		r.calls++
		return sourceinventory.Report{Counters: prefix}, sourceinventory.ErrParse
	}
	r.deps.WriteResult = func(f *os.File, raw []byte) error { _ = f.Close(); return errors.New("SECRET disk path") }
	e := execute(r, ".cache/write-failure/results.json")
	var got persistenceError
	if !errors.As(e, &got) || got.Counters != prefix || got.Invocation != (invocation{1, 1, true}) || strings.Contains(e.Error(), "SECRET") || r.calls != 1 {
		t.Fatal("failed persistence lost counters or leaked", e)
	}
}

func TestToyPreflightMutationNeverCallsInventory(t *testing.T) {
	tests := []struct {
		name   string
		change func(*testing.T, *toyRepo)
	}{
		{"input", func(t *testing.T, r *toyRepo) { put(t, r.dir, r.deps.Pins[0].Path, []byte("changed toy source")) }},
		{"metadata", func(t *testing.T, r *toyRepo) { put(t, r.dir, r.deps.Pins[4].Path, []byte("changed toy metadata")) }},
		{"source", func(t *testing.T, r *toyRepo) {
			put(t, r.dir, r.deps.Sources[0].Path, []byte("changed toy compiler source"))
		}},
		{"extra_go", func(t *testing.T, r *toyRepo) {
			put(t, r.dir, "internal/sourceinventory/unlisted_linux.go", []byte("package toy"))
		}},
		{"support", func(t *testing.T, r *toyRepo) { put(t, r.dir, "LICENSE", []byte("changed toy support")) }},
		{"recipe", func(t *testing.T, r *toyRepo) { put(t, r.dir, buildRecipePath, []byte("{}\n")) }},
		{"binary", func(t *testing.T, r *toyRepo) {
			r.deps.Binary = func() (artifact, error) { return artifact{SHA256: sha([]byte("changed binary")), Bytes: 1}, nil }
		}},
		{"git", func(t *testing.T, r *toyRepo) {
			r.deps.Git = func(context.Context, string, string, string) ([]byte, error) { return []byte("wrong Git blob"), nil }
		}},
		{"plan", func(t *testing.T, r *toyRepo) { put(t, r.dir, planPath, []byte("{}\n")) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := fixture(t)
			prepare(t, r)
			tc.change(t, r)
			out := ".cache/stopped/results.json"
			if e := execute(r, out); e == nil {
				t.Fatal("mutation accepted")
			}
			if r.calls != 0 {
				t.Fatal("preflight ran inventory")
			}
			if _, e := os.Stat(filepath.Join(r.dir, ".cache/stopped")); !os.IsNotExist(e) {
				t.Fatal("preflight created output")
			}
		})
	}
}

func TestToySymlinkAndFileBounds(t *testing.T) {
	r := fixture(t)
	root, e := os.OpenRoot(r.dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	put(t, r.dir, "exact", bytes.Repeat([]byte{'x'}, maxBytes))
	if b, e := bounded(root, "exact", maxBytes); e != nil || len(b) != maxBytes {
		t.Fatal("exact byte bound rejected")
	}
	put(t, r.dir, "large", bytes.Repeat([]byte{'x'}, maxBytes+1))
	if _, e := bounded(root, "large", maxBytes); e == nil {
		t.Fatal("oversized file accepted")
	}
	if e := os.Symlink("exact", filepath.Join(r.dir, "symlink")); e != nil {
		t.Fatal(e)
	}
	if _, e := bounded(root, "symlink", maxBytes); e == nil {
		t.Fatal("symlink input accepted")
	}
	for _, p := range []string{"../exact", "/absolute", "a\\b", "a:b", "a/../exact", ""} {
		if _, e := bounded(root, p, maxBytes); e == nil {
			t.Fatal("unsafe path accepted")
		}
	}
	if e := os.Symlink(".", filepath.Join(r.dir, ".cache/link")); e != nil {
		t.Fatal(e)
	}
	if f, e := reserve(root, ".cache/link/new/results.json"); e == nil {
		f.Close()
		t.Fatal("symlink output parent accepted")
	}
}

func TestBoundedGitWriterCannotBypassLimit(t *testing.T) {
	var b boundedBuffer
	source := bytes.NewReader(bytes.Repeat([]byte{'x'}, maxBytes+1))
	if _, e := io.Copy(&b, source); e == nil || len(b.raw) > maxBytes {
		t.Fatal("io.Copy bypassed Write bound")
	}
	if _, e := b.Write(bytes.Repeat([]byte{'x'}, maxBytes)); e != nil {
		t.Fatal("exact limit rejected")
	}
	if _, e := b.Write([]byte{'x'}); e == nil {
		t.Fatal("overflow accepted")
	}
}

func buildFixture() *debug.BuildInfo {
	return &debug.BuildInfo{GoVersion: "go1.27.1", Path: buildPath, Settings: []debug.BuildSetting{{Key: "-trimpath", Value: "true"}, {Key: "CGO_ENABLED", Value: "0"}, {Key: "GOOS", Value: runtime.GOOS}, {Key: "GOARCH", Value: runtime.GOARCH}}}
}
func TestBuildInfoPlatformAndVCSFailClosed(t *testing.T) {
	if !validBuild(buildFixture()) {
		t.Fatal("correct build rejected")
	}
	if validBuild(nil) {
		t.Fatal("missing build accepted")
	}
	for _, setting := range []debug.BuildSetting{{Key: "vcs", Value: "git"}, {Key: "vcs.revision", Value: strings.Repeat("a", 40)}, {Key: "-trimpath", Value: "false"}, {Key: "CGO_ENABLED", Value: "1"}, {Key: "GOOS", Value: "wrong"}, {Key: "GOARCH", Value: "wrong"}, {Key: "-buildvcs", Value: "true"}} {
		info := buildFixture()
		if setting.Key == "-buildvcs" || strings.HasPrefix(setting.Key, "vcs") {
			info.Settings = append(info.Settings, setting)
		} else {
			for i, s := range info.Settings {
				if s.Key == setting.Key {
					info.Settings[i] = setting
				}
			}
		}
		if validBuild(info) {
			t.Fatal("invalid build accepted", setting.Key)
		}
	}
	info := buildFixture()
	info.GoVersion = "go1.27.2"
	if validBuild(info) {
		t.Fatal("wrong compiler")
	}
	info = buildFixture()
	info.Settings = append(info.Settings, info.Settings[0])
	if validBuild(info) {
		t.Fatal("duplicate setting")
	}
}

func TestCanonicalPlanRejectsShapeDuplicatesAndDrift(t *testing.T) {
	r := fixture(t)
	p := prepare(t, r)
	raw, e := canonical(p)
	if e != nil {
		t.Fatal(e)
	}
	bad := [][]byte{append(slices.Clone(raw), ' '), append(slices.Clone(raw), []byte("{}")...), []byte(strings.Replace(string(raw), "\"schema\":", "\"extra\": 0, \"schema\":", 1)), []byte(strings.Replace(string(raw), "\"planned_inventory_retries\": 0", "\"planned_inventory_retries\": 0, \"planned_inventory_retries\": 0", 1))}
	for _, b := range bad {
		var got plan
		if e := decodeCanonical(b, &got); e == nil {
			t.Fatal("noncanonical plan accepted")
		}
	}
	p.Expected.Format = 53
	if e := validatePlan(p, newPlan(p.SourceCommit, p.Binary, r.deps.Pins, r.deps.Sources, p.Support)); e == nil {
		t.Fatal("forecast changed to unique hash count")
	}
}

func TestEmbeddedSourceDescriptorsAreOwned(t *testing.T) {
	a := sourceinventory.SourceArtifacts()
	b := sourceinventory.SourceArtifacts()
	if !reflect.DeepEqual(a, b) || len(a) != 3 {
		t.Fatal("embedding descriptor count")
	}
	a[0].SHA256 = "mutated"
	if reflect.DeepEqual(a, sourceinventory.SourceArtifacts()) {
		t.Fatal("descriptor aliases shared state")
	}
}

func TestOnlySyntheticRebindReturnsRealPrefixCounters(t *testing.T) {
	// Exactly one actual Rebind call in this wrapper test suite, on four tiny
	// independently authored package declarations. Registry lookup then fails.
	var in sourceinventory.Input
	in.GoVersion = "go1.27.1"
	paths := inputPins()
	for i := range in.Files {
		raw := []byte("package typedbehavior\n")
		if i == 0 {
			raw = []byte("package behaviorprobe\n")
		}
		in.Files[i] = sourceinventory.File{Path: paths[i].Path, Raw: raw, ExpectedSHA256: sha(raw)}
	}
	in.Roots = []sourceinventory.HistoricalRoot{{Cohort: "legacy", ID: "toy", Prototype: "toy", Core: "toy", CodeSHA256: sha([]byte("toy")), NormalizedCodeSHA256: sha([]byte("toy"))}}
	got, e := sourceinventory.Rebind(in)
	if !errors.Is(e, sourceinventory.ErrRegistry) || got.Counters.FileHashAttempts != 4 || got.Counters.ParseAttempts != 4 || got.Counters.ParsesSucceeded != 4 || got.Counters.RootAttempts != 0 || got.Counters.FormatCallsAttempted != 0 {
		t.Fatal("synthetic prefix", got.Counters, e)
	}
}
