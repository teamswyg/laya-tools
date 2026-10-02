// SPDX-License-Identifier: Apache-2.0
// Bounded maintainer runner; no candidate or model execution.
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

	"github.com/teamswyg/laya-tools/internal/sourceinventory"
)

//go:embed main.go
var compiledMain []byte

//go:embed metadata.go
var compiledMetadata []byte

const (
	maxBytes        = 1 << 20
	maxBinary       = 64 << 20
	planPath        = "experiments/short-claim/execution-plan-60.json"
	planSchema      = "riido-source-inventory-execution-plan-60-v1"
	recordSchema    = "riido-source-inventory-record-60-v1"
	buildRecipePath = "experiments/short-claim/build-recipe-60.json"
	buildPath       = "github.com/teamswyg/laya-tools/cmd/riido-sourceinventory"
)

type codeError string

func (e codeError) Error() string { return string(e) }
func fail(s string) error         { return codeError(s) }

type artifact = sourceinventory.SourceArtifact

func inputPins() [5]artifact {
	return [5]artifact{
		{Path: "internal/behaviorprobe/data.go", SHA256: "b377343a64c019c205f69865286a21ce5d91dda3051c31aa5c69b75a40300559", Bytes: 28420},
		{Path: "internal/typedbehavior/spec.go", SHA256: "013a486ecc1d28913c2dda8cf74b3068dcb193d8b77e534dc62e2f87b9ee6b80", Bytes: 1228},
		{Path: "internal/typedbehavior/state.go", SHA256: "35324ed595bcf384c484f33d98c6866164bebb2b636986ee8ea51bb979bdad7c", Bytes: 19402},
		{Path: "internal/typedbehavior/flow.go", SHA256: "ea7f09d31e15e561173932ac0bb61db17746cdce929e0f976d97994d282d898b", Bytes: 18242},
		{Path: "experiments/short-claim/results-56b.json", SHA256: "4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4", Bytes: 266818},
	}
}

func compiledPaths() [5]string {
	return [5]string{"cmd/riido-sourceinventory/main.go", "cmd/riido-sourceinventory/metadata.go", "internal/sourceinventory/inventory.go", "internal/sourceinventory/registry.go", "internal/sourceinventory/provenance.go"}
}

func supportPaths() [16]string {
	return [16]string{
		"experiments/short-claim/PLAN-INVENTORY-60.ko.md", "experiments/short-claim/PLAN-INVENTORY-60.en.md",
		"experiments/short-claim/source-inventory-recipe-60.json", buildRecipePath,
		"cmd/riido-sourceinventory/main_test.go", "cmd/riido-sourceinventory/metadata_test.go",
		"internal/sourceinventory/inventory_test.go", "go.mod", "go.sum", "LICENSE",
		"experiments/short-claim/USAGE-INVENTORY-60.ko.md", "experiments/short-claim/USAGE-INVENTORY-60.en.md",
		"experiments/short-claim/CONTENT-REVIEW-PROTOCOL-60.ko.md", "experiments/short-claim/CONTENT-REVIEW-PROTOCOL-60.en.md",
		"experiments/short-claim/content-review-protocol-60.json", "experiments/short-claim/content-review-scope-references-60.json",
	}
}

type buildRecipe struct {
	Schema string   `json:"schema"`
	Go     string   `json:"go_version"`
	CGO    string   `json:"cgo_enabled"`
	Args   []string `json:"arguments"`
}

func fixedRecipe() buildRecipe {
	return buildRecipe{"riido-source-inventory-build-recipe-60-v1", "go1.27.1", "0", []string{"build", "-trimpath", "-buildvcs=false", "-p=1", "-o", "<private_binary>", "./cmd/riido-sourceinventory"}}
}

type expectedWork struct {
	Provenance string `json:"provenance"`
	Parse      int    `json:"parse_calls"`
	Roots      int    `json:"root_relations"`
	Components int    `json:"component_relations"`
	Format     int    `json:"format_calls"`
	Scan       int    `json:"scanner_calls"`
	Bundle     int    `json:"bundle_calls"`
}

func expected() expectedWork {
	return expectedWork{"retained_metadata_expectation_not_observed", 4, 60, 204, 264, 112, 24}
}

type plan struct {
	Schema           string       `json:"schema"`
	SourceCommit     string       `json:"source_freeze_commit"`
	Go               string       `json:"go_version"`
	CGO              string       `json:"cgo_enabled"`
	TrimPath         bool         `json:"trimpath"`
	BuildVCS         bool         `json:"buildvcs"`
	GOOS             string       `json:"goos"`
	GOARCH           string       `json:"goarch"`
	Binary           artifact     `json:"binary"`
	Inputs           []artifact   `json:"input_artifacts"`
	CompileSource    []artifact   `json:"compiled_source_artifacts"`
	Support          []artifact   `json:"support_artifacts"`
	CompileBinding   string       `json:"compile_binding"`
	Expected         expectedWork `json:"expected_work_not_observed"`
	Attempts         int          `json:"planned_inventory_attempts"`
	Retries          int          `json:"planned_inventory_retries"`
	GitBudgetSeconds int          `json:"git_blob_budget_seconds_per_file"`
	GOMAXPROCS       int          `json:"go_max_procs"`
	HeapSoftBytes    int64        `json:"go_heap_soft_limit_bytes"`
	Policy           string       `json:"execution_policy"`
	ContentReview    string       `json:"content_review"`
}

type invocation struct {
	Attempts          int  `json:"rebind_attempts"`
	Returned          int  `json:"rebind_calls_returned"`
	CountersAvailable bool `json:"partial_inventory_counters_available"`
}

type record struct {
	Schema        string                 `json:"schema"`
	State         string                 `json:"state"`
	FailureCode   string                 `json:"failure_code"`
	FailureCause  string                 `json:"failure_cause"`
	PlanSHA       string                 `json:"plan_sha256"`
	InputCommit   string                 `json:"input_freeze_commit"`
	Plan          plan                   `json:"plan"`
	VerifiedBlobs int                    `json:"git_blobs_verified_before_inventory"`
	Invocation    invocation             `json:"invocation"`
	Inventory     sourceinventory.Report `json:"inventory"`
	ContentReview string                 `json:"content_review"`
	TrainingReady bool                   `json:"training_ready"`
	Limitations   []string               `json:"limitations"`
}

type arguments struct{ Mode, Repo, SourceCommit, InputCommit, Out string }

func parse(args []string) (arguments, error) {
	var c arguments
	if len(args) == 0 || args[0] != "prepare" && args[0] != "inventory" {
		return c, fail("inventory60_mode_invalid")
	}
	c.Mode = args[0]
	f := flag.NewFlagSet("riido-sourceinventory", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.StringVar(&c.Repo, "repo", ".", "repository directory")
	f.StringVar(&c.SourceCommit, "source-commit", "", "full source freeze commit for prepare")
	f.StringVar(&c.InputCommit, "input-commit", "", "full input freeze commit for inventory")
	f.StringVar(&c.Out, "out", "", "new private file under .cache")
	if f.Parse(args[1:]) != nil || f.NArg() != 0 || !safePath(c.Out) || !strings.HasPrefix(c.Out, ".cache/") {
		return c, fail("inventory60_arguments_invalid")
	}
	if c.Mode == "prepare" && (!hash(c.SourceCommit, 40) || c.InputCommit != "") || c.Mode == "inventory" && (!hash(c.InputCommit, 40) || c.SourceCommit != "" || filepath.Base(c.Out) != "results.json") {
		return c, fail("inventory60_commit_arguments_invalid")
	}
	return c, nil
}

func sha(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func hash(s string, n int) bool {
	_, e := hex.DecodeString(s)
	return len(s) == n && strings.ToLower(s) == s && e == nil
}
func safePath(p string) bool {
	return p != "" && p != "." && !strings.ContainsAny(p, "\\:\x00") && !filepath.IsAbs(p) && fs.ValidPath(p)
}

func bounded(root *os.Root, p string, limit int) ([]byte, error) {
	if !safePath(p) {
		return nil, fail("inventory60_path_invalid")
	}
	before, e := root.Lstat(p)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 0 || before.Size() > int64(limit) {
		return nil, fail("inventory60_file_bounds")
	}
	f, e := root.Open(p)
	if e != nil {
		return nil, fail("inventory60_file_unavailable")
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) || after.Size() != before.Size() {
		return nil, fail("inventory60_file_bounds")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if e != nil || len(b) > limit || int64(len(b)) != after.Size() {
		return nil, fail("inventory60_file_bounds")
	}
	return b, nil
}

func canonical(v any) ([]byte, error) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil || len(b)+1 > maxBytes {
		return nil, fail("inventory60_json_bounds")
	}
	return append(b, '\n'), nil
}
func decodeCanonical(raw []byte, v any) error {
	if len(raw) == 0 || len(raw) > maxBytes || !utf8.Valid(raw) {
		return fail("inventory60_json_invalid")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		return fail("inventory60_json_invalid")
	}
	b, e := canonical(v)
	if e != nil || !bytes.Equal(b, raw) {
		return fail("inventory60_json_noncanonical")
	}
	return nil
}

func compiledSources() []artifact {
	out := []artifact{{Path: "cmd/riido-sourceinventory/main.go", SHA256: sha(compiledMain), Bytes: len(compiledMain)}, {Path: "cmd/riido-sourceinventory/metadata.go", SHA256: sha(compiledMetadata), Bytes: len(compiledMetadata)}}
	return append(out, sourceinventory.SourceArtifacts()...)
}

func artifactsValid(pins []artifact) bool {
	var seen []string
	for _, a := range pins {
		if !safePath(a.Path) || !hash(a.SHA256, 64) || a.Bytes <= 0 || a.Bytes > maxBytes || slices.Contains(seen, a.Path) {
			return false
		}
		seen = append(seen, a.Path)
	}
	return true
}

func verifyCompileClosure(root *os.Root, pins []artifact) error {
	paths := compiledPaths()
	if len(pins) != len(paths) || !artifactsValid(pins) {
		return fail("inventory60_compile_source_invalid")
	}
	for i, p := range paths {
		if pins[i].Path != p {
			return fail("inventory60_compile_source_invalid")
		}
	}
	for _, dir := range []string{"cmd/riido-sourceinventory", "internal/sourceinventory"} {
		entries, e := fs.ReadDir(root.FS(), dir)
		if e != nil {
			return fail("inventory60_compile_directory_invalid")
		}
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !slices.Contains(paths[:], dir+"/"+entry.Name()) {
				return fail("inventory60_unlisted_compiled_go_source")
			}
		}
	}
	return nil
}

// No promoted ReadFrom method: io.Copy must pass through bounded Write.
type boundedBuffer struct{ raw []byte }

func (b *boundedBuffer) Write(p []byte) (int, error) {
	if len(p) > maxBytes-len(b.raw) {
		return 0, fail("inventory60_git_blob_bounds")
	}
	b.raw = append(b.raw, p...)
	return len(p), nil
}

type gitRead func(context.Context, string, string, string) ([]byte, error)

func readGit(ctx context.Context, repo, commit, path string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", "show", commit+":"+path)
	cmd.Dir = repo
	var b boundedBuffer
	cmd.Stdout = &b
	cmd.Stderr = io.Discard
	if cmd.Run() != nil {
		return nil, fail("inventory60_git_blob_unavailable")
	}
	return b.raw, nil
}
func verifyBlobs(repo string, root *os.Root, commit string, pins []artifact, git gitRead) error {
	if !hash(commit, 40) || !artifactsValid(pins) {
		return fail("inventory60_git_pin_invalid")
	}
	for _, a := range pins {
		raw, e := bounded(root, a.Path, maxBytes)
		if e != nil || len(raw) != a.Bytes || sha(raw) != a.SHA256 {
			return fail("inventory60_disk_pin_mismatch")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		got, e := git(ctx, repo, commit, a.Path)
		cancel()
		if e != nil || len(got) > maxBytes || !bytes.Equal(raw, got) {
			return fail("inventory60_git_pin_mismatch")
		}
	}
	return nil
}

func validBuild(info *debug.BuildInfo) bool {
	if info == nil || info.GoVersion != "go1.27.1" || info.Path != buildPath {
		return false
	}
	var keys []string
	trim, cgo, osOK, archOK := false, false, false, false
	for _, s := range info.Settings {
		if slices.Contains(keys, s.Key) || strings.HasPrefix(s.Key, "vcs.") || s.Key == "vcs" {
			return false
		}
		keys = append(keys, s.Key)
		switch s.Key {
		case "-trimpath":
			trim = s.Value == "true"
		case "CGO_ENABLED":
			cgo = s.Value == "0"
		case "GOOS":
			osOK = s.Value == runtime.GOOS
		case "GOARCH":
			archOK = s.Value == runtime.GOARCH
		case "-buildvcs":
			if s.Value != "false" {
				return false
			}
		}
	}
	return trim && cgo && osOK && archOK
}

func binaryArtifact() (artifact, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || !validBuild(info) || runtime.Version() != "go1.27.1" {
		return artifact{}, fail("inventory60_binary_build_mismatch")
	}
	p, e := os.Executable()
	if e != nil {
		return artifact{}, fail("inventory60_binary_unavailable")
	}
	f, e := os.Open(p)
	if e != nil {
		return artifact{}, fail("inventory60_binary_unavailable")
	}
	defer f.Close()
	s, e := f.Stat()
	if e != nil || !s.Mode().IsRegular() || s.Size() <= 0 || s.Size() > maxBinary {
		return artifact{}, fail("inventory60_binary_bounds")
	}
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, maxBinary+1))
	if e != nil || n != s.Size() || n > maxBinary {
		return artifact{}, fail("inventory60_binary_bounds")
	}
	return artifact{SHA256: hex.EncodeToString(h.Sum(nil)), Bytes: int(n)}, nil
}

type dependencies struct {
	CompileBindingVerified bool
	Sources                []artifact
	Pins                   [5]artifact
	Binary                 func() (artifact, error)
	Git                    gitRead
	Rebind                 func(sourceinventory.Input) (sourceinventory.Report, error)
	WriteResult            func(*os.File, []byte) error
}

func productionDependencies() dependencies {
	// Every runtime file is embedded and checked against both disk and Git
	// before Rebind. This proves compiled bytes, not semantic source closure.
	return dependencies{CompileBindingVerified: true, Sources: compiledSources(), Pins: inputPins(), Binary: binaryArtifact, Git: readGit, Rebind: sourceinventory.Rebind, WriteResult: finish}
}

func loadInputs(root *os.Root, pins [5]artifact) ([5][]byte, error) {
	var raw [5][]byte
	paths := inputPins()
	for i, a := range pins {
		if a.Path != paths[i].Path || !artifactsValid([]artifact{a}) {
			return raw, fail("inventory60_input_manifest_invalid")
		}
		b, e := bounded(root, a.Path, maxBytes)
		if e != nil || len(b) != a.Bytes || sha(b) != a.SHA256 {
			return raw, fail("inventory60_input_pin_mismatch")
		}
		raw[i] = b
	}
	return raw, nil
}

func loadSupport(root *os.Root) ([]artifact, error) {
	paths := supportPaths()
	out := make([]artifact, 0, len(paths))
	for _, p := range paths {
		b, e := bounded(root, p, maxBytes)
		if e != nil || len(b) == 0 {
			return nil, fail("inventory60_support_invalid")
		}
		if p == buildRecipePath {
			var got buildRecipe
			if decodeCanonical(b, &got) != nil {
				return nil, fail("inventory60_build_recipe_invalid")
			}
			want, _ := canonical(fixedRecipe())
			if !bytes.Equal(b, want) {
				return nil, fail("inventory60_build_recipe_invalid")
			}
		}
		out = append(out, artifact{Path: p, SHA256: sha(b), Bytes: len(b)})
	}
	return out, nil
}

func newPlan(commit string, binary artifact, pins [5]artifact, sources, support []artifact) plan {
	return plan{Schema: planSchema, SourceCommit: commit, Go: "go1.27.1", CGO: "0", TrimPath: true, BuildVCS: false, GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Binary: binary, Inputs: slices.Clone(pins[:]), CompileSource: slices.Clone(sources), Support: slices.Clone(support), CompileBinding: "compiled_package_source_bytes_verified_ast_only", Expected: expected(), Attempts: 1, Retries: 0, GitBudgetSeconds: 5, GOMAXPROCS: 1, HeapSoftBytes: 256 << 20, Policy: "preflight_then_exclusive_output_then_single_rebind_no_automatic_retry", ContentReview: "pending"}
}

func validatePlan(got, want plan) error {
	if !hash(got.SourceCommit, 40) || !hash(got.Binary.SHA256, 64) || got.Binary.Bytes <= 0 || got.Binary.Bytes > maxBinary || got.Binary.Path != "" || len(got.Inputs) != 5 || len(got.CompileSource) != 5 || len(got.Support) != len(supportPaths()) {
		return fail("inventory60_plan_invalid")
	}
	all := append(append(slices.Clone(got.Inputs), got.CompileSource...), got.Support...)
	if !artifactsValid(all) {
		return fail("inventory60_plan_invalid")
	}
	a, e := canonical(got)
	b, w := canonical(want)
	if e != nil || w != nil || !bytes.Equal(a, b) {
		return fail("inventory60_plan_mismatch")
	}
	return nil
}

func parentsRegular(root *os.Root, p string) error {
	parent := filepath.Dir(p)
	var prefix string
	if parent == "." {
		return nil
	}
	for _, part := range strings.Split(parent, "/") {
		if prefix == "" {
			prefix = part
		} else {
			prefix += "/" + part
		}
		s, e := root.Lstat(prefix)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 {
			return fail("inventory60_output_parent_invalid")
		}
	}
	return nil
}

func openNew(root *os.Root, p string) (*os.File, error) {
	if !safePath(p) || parentsRegular(root, p) != nil {
		return nil, fail("inventory60_output_parent_invalid")
	}
	f, e := root.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return nil, fail("inventory60_output_exists_or_unavailable")
	}
	return f, nil
}
func reserve(root *os.Root, p string) (*os.File, error) {
	if !safePath(p) || filepath.Base(p) != "results.json" || parentsRegular(root, filepath.Dir(p)) != nil {
		return nil, fail("inventory60_output_parent_invalid")
	}
	if root.Mkdir(filepath.Dir(p), 0700) != nil {
		return nil, fail("inventory60_output_directory_exists_or_unavailable")
	}
	return openNew(root, p)
}
func finish(f *os.File, b []byte) error {
	if len(b) > maxBytes {
		_ = f.Close()
		return fail("inventory60_result_bounds")
	}
	n, e := f.Write(b)
	syncErr := f.Sync()
	closeErr := f.Close()
	if e != nil || syncErr != nil || closeErr != nil || n != len(b) {
		return fail("inventory60_result_write_failed")
	}
	return nil
}

func safeCause(err error) string {
	for _, known := range []error{sourceinventory.ErrInput, sourceinventory.ErrFilePin, sourceinventory.ErrParse, sourceinventory.ErrShape, sourceinventory.ErrRegistry, sourceinventory.ErrDeclaration, sourceinventory.ErrHistoricalPin, sourceinventory.ErrRecipe} {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	return "inventory60_rebind_error_unclassified"
}

// A write/encode failure still returns safe counters rather than raw errors.
type persistenceError struct {
	Code       string
	Invocation invocation
	Counters   sourceinventory.Counters
}

func (e persistenceError) Error() string { b, _ := json.Marshal(e); return string(b) }

func persist(f *os.File, r record, write func(*os.File, []byte) error) error {
	b, e := canonical(r)
	if e != nil {
		_ = f.Close()
		return persistenceError{"inventory60_result_encoding_failed", r.Invocation, r.Inventory.Counters}
	}
	if write(f, b) != nil {
		_ = f.Close()
		return persistenceError{"inventory60_result_persistence_failed", r.Invocation, r.Inventory.Counters}
	}
	return nil
}

// Validate returned evidence without executing the engine again or replacing
// its actual counters with the predeclared expectation.
func completeScope(in sourceinventory.Input, got sourceinventory.Report) bool {
	if len(in.Roots) != 60 || len(got.Files) != 4 || len(got.Roots) != len(in.Roots) {
		return false
	}
	wantCounters := sourceinventory.Counters{
		FileHashAttempts: 4, FileHashesCompleted: 4, FilePinsMatched: 4,
		ParseAttempts: 4, ParsesCompleted: 4, ParsesSucceeded: 4, RegistryEntries: 60,
		RootAttempts: 60, RootsCompleted: 60, ComponentAttempts: 204, ComponentsCompleted: 204,
		RawSpanHashAttempts: 264, RawSpanHashesCompleted: 264,
		FormatCallsAttempted: 264, FormatCallsCompleted: 264,
		ScanCallsAttempted: 112, ScanCallsCompleted: 112,
		BundleCallsAttempted: 24, BundleCallsCompleted: 24,
	}
	if got.Counters != wantCounters {
		return false
	}
	for i, f := range got.Files {
		if f.Path != in.Files[i].Path || f.Bytes != len(in.Files[i].Raw) || f.SHA256 != in.Files[i].ExpectedSHA256 {
			return false
		}
	}
	for i, r := range got.Roots {
		want := in.Roots[i]
		if r.ID != want.ID || r.Cohort != want.Cohort || !r.MetadataPinsMatched || r.ContentReview != "pending" || r.Root.FormattedSHA256 != want.CodeSHA256 || r.Root.NormalizedSHA256 != want.NormalizedCodeSHA256 || r.HistoricalBundleSHA256 != want.BundleSHA256 || len(r.Components) != len(want.Components) {
			return false
		}
		for j, c := range r.Components {
			expected := want.Components[j]
			if c.ID != expected.ID || c.Kind != expected.Kind || c.FormattedSHA256 != expected.SHA256 || c.NormalizedSHA256 != expected.NormalizedBehaviorSHA256 {
				return false
			}
		}
	}
	return true
}

func invoke(in sourceinventory.Input, r *record, engine func(sourceinventory.Input) (sourceinventory.Report, error)) {
	r.Invocation.Attempts++
	defer func() {
		if recover() != nil {
			r.FailureCode = "inventory60_rebind_failed"
			r.FailureCause = "inventory60_rebind_panic_unclassified"
			r.State = "incomplete"
			r.Invocation.CountersAvailable = false
		}
	}()
	got, e := engine(in)
	r.Invocation.Returned++
	r.Invocation.CountersAvailable = true
	r.Inventory = got
	if e != nil {
		r.FailureCode = "inventory60_rebind_failed"
		r.FailureCause = safeCause(e)
		return
	}
	if got.Schema != "riido-source-declaration-reference-draft-v1" || got.State != "metadata_rebound_content_review_pending" || got.ContentReview != "pending" || got.ObjectBindingPerformed || got.ClosureDiscoveryPerformed || got.ReusePolicy != "no_cache_each_relation_v1" || !completeScope(in, got) {
		r.FailureCode = "inventory60_rebind_failed"
		r.FailureCause = "inventory60_report_scope_invalid"
		return
	}
	r.State = "metadata_rebound_content_review_pending"
}

func runWith(args []string, deps dependencies) error {
	c, e := parse(args)
	if e != nil {
		return e
	}
	if !deps.CompileBindingVerified {
		return fail("inventory60_compile_binding_unverified")
	}
	if runtime.Version() != "go1.27.1" {
		return fail("inventory60_toolchain_mismatch")
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	root, e := os.OpenRoot(c.Repo)
	if e != nil {
		return fail("inventory60_repository_unavailable")
	}
	defer root.Close()
	raw, e := loadInputs(root, deps.Pins)
	if e != nil {
		return e
	}
	if e := verifyCompileClosure(root, deps.Sources); e != nil {
		return e
	}
	support, e := loadSupport(root)
	if e != nil {
		return e
	}
	binary, e := deps.Binary()
	if e != nil {
		return e
	}
	sourceEntries := append(slices.Clone(deps.Sources), support...)
	if c.Mode == "prepare" {
		all := append(slices.Clone(sourceEntries), deps.Pins[:]...)
		if e := verifyBlobs(c.Repo, root, c.SourceCommit, all, deps.Git); e != nil {
			return e
		}
		p := newPlan(c.SourceCommit, binary, deps.Pins, deps.Sources, support)
		if e := validatePlan(p, p); e != nil {
			return e
		}
		b, e := canonical(p)
		if e != nil {
			return e
		}
		f, e := openNew(root, c.Out)
		if e != nil {
			return e
		}
		return finish(f, b)
	}
	planRaw, e := bounded(root, planPath, maxBytes)
	if e != nil {
		return e
	}
	var p plan
	if e := decodeCanonical(planRaw, &p); e != nil {
		return e
	}
	if e := validatePlan(p, newPlan(p.SourceCommit, binary, deps.Pins, deps.Sources, support)); e != nil {
		return e
	}
	if e := verifyBlobs(c.Repo, root, p.SourceCommit, sourceEntries, deps.Git); e != nil {
		return e
	}
	inputEntries := append(slices.Clone(deps.Pins[:]), artifact{Path: planPath, SHA256: sha(planRaw), Bytes: len(planRaw)})
	if e := verifyBlobs(c.Repo, root, c.InputCommit, inputEntries, deps.Git); e != nil {
		return e
	}
	roots, e := historicalRoots(raw[4])
	if e != nil {
		return e
	}
	in := sourceinventory.Input{GoVersion: runtime.Version(), Roots: roots}
	for i := range in.Files {
		in.Files[i] = sourceinventory.File{Path: deps.Pins[i].Path, Raw: raw[i], ExpectedSHA256: deps.Pins[i].SHA256}
	}
	f, e := reserve(root, c.Out)
	if e != nil {
		return e
	}
	r := record{Schema: recordSchema, State: "incomplete", PlanSHA: sha(planRaw), InputCommit: c.InputCommit, Plan: p, VerifiedBlobs: len(sourceEntries) + len(inputEntries), ContentReview: "pending", Limitations: []string{
		"Stored declaration/digest correspondence only; syntax is not runtime object binding or semantic closure inference.",
		"Expected operation counts are metadata forecasts; observed counters start at zero and come only from the returned inventory report.",
		"All caption fidelity, contract and observation-field review remains pending; no labels, roles, historical SourcePins/Bind/Generate APIs, candidates, models, fitting or protected-final access.",
		"Fresh output prevents reuse at this path; one-off attempt policy remains externally frozen, with no automatic retry or host-global lock.",
		"After a panic, inner partial counters are unavailable; no zero-counter completion or success is inferred. Process termination or disk failure cannot guarantee a persisted envelope.",
		"Go scheduling and heap soft limits are settings, not measured CPU/GPU/RSS/latency or savings.",
	}}
	invoke(in, &r, deps.Rebind)
	if e := persist(f, r, deps.WriteResult); e != nil {
		return e
	}
	if r.FailureCode != "" {
		return persistenceError{r.FailureCode + ":" + r.FailureCause, r.Invocation, r.Inventory.Counters}
	}
	return nil
}

func main() {
	if e := runWith(os.Args[1:], productionDependencies()); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
