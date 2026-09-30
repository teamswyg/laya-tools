package taskverify

import (
	"bytes"
	"context"
	"encoding/json"
	"go/format"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func pinnedFixture(t *testing.T, task string) []BaseFile {
	t.Helper()
	paths, err := BasePaths(task)
	if err != nil {
		t.Fatal(err)
	}
	files := []BaseFile{}
	for _, p := range paths {
		file, err := os.Open(filepath.Join("testdata/base", filepath.FromSlash(p)))
		if err != nil {
			t.Fatal("pinned public fixture missing")
		}
		data, err := io.ReadAll(io.LimitReader(file, MaxFileBytes+1))
		file.Close()
		if err != nil || len(data) > MaxFileBytes {
			t.Fatal("invalid pinned public fixture")
		}
		files = append(files, BaseFile{p, digest(data), data})
	}
	if _, err := validateBase(task, BaseRevision, files); err != nil {
		t.Fatal(err)
	}
	return files
}

func candidateFixture(t *testing.T, files []BaseFile) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, f.Data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func replaceCandidate(t *testing.T, dir, p, old, newText string) {
	t.Helper()
	file := filepath.Join(dir, filepath.FromSlash(p))
	data, err := os.ReadFile(file)
	if err != nil || bytes.Count(data, []byte(old)) != 1 {
		t.Fatal("fixture replacement precondition")
	}
	data = bytes.Replace(data, []byte(old), []byte(newText), 1)
	if strings.HasSuffix(p, ".go") {
		data, err = format.Source(data)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func verifyFixture(t *testing.T, task string, files []BaseFile, dir string) Report {
	t.Helper()
	r, err := Verify(context.Background(), Request{TaskID: task, BaseRevision: BaseRevision, BaseFiles: files, CandidateDir: dir, Timeout: MaxTimeout})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestExactPreviewComment(t *testing.T) {
	files := pinnedFixture(t, "comment-preview-authority")
	dir := candidateFixture(t, files)
	r := verifyFixture(t, "comment-preview-authority", files, dir)
	if r.Accepted || r.Status != "rejected" || r.Checks[len(r.Checks)-1].Code != "unchanged_base" {
		t.Fatal("unchanged base accepted", r)
	}
	replaceCandidate(t, dir, "pkg/reporouter/router.go", "// candidate, suggest, abstain; never execution authority", "// candidate, suggest, or abstain; a suggestion never authorizes execution.")
	r = verifyFixture(t, "comment-preview-authority", files, dir)
	if !r.Accepted || r.Status != "accepted" || r.CandidateScope != "task_closure" || r.CandidateTestsUsed || r.IndependentTests != 0 {
		t.Fatal("exact independent comment check failed", r)
	}
	replaceCandidate(t, dir, "pkg/reporouter/router.go", "return Config{4, .9, .15}", "return Config{3, .9, .15}")
	r = verifyFixture(t, "comment-preview-authority", files, dir)
	if r.Accepted || r.Checks[len(r.Checks)-1].Code != "not_exact_comment_change" {
		t.Fatal("comment plus behavior change accepted", r)
	}
}

func TestBudgetComment(t *testing.T) {
	files := pinnedFixture(t, "comment-budget-period")
	dir := candidateFixture(t, files)
	replaceCandidate(t, dir, "pkg/catalog/catalog.go", "// Snapshot for one caller-defined budget period. This is not a reservation ledger.", "// Snapshot supplied by the caller for a single budget period; no reservation is made.")
	r := verifyFixture(t, "comment-budget-period", files, dir)
	if !sandboxSupported() {
		if r.Status != "verifier_unknown" || r.Accepted {
			t.Fatal("unsupported isolation claimed acceptance", r)
		}
		return
	}
	if !r.Accepted || !r.ExecutionIsolated || r.IndependentTests == 0 {
		t.Fatal("exact comment did not pass pinned tests", r)
	}
	replaceCandidate(t, dir, "go.mod", "go 1.27.1", "go 1.27.0")
	r = verifyFixture(t, "comment-budget-period", files, dir)
	if r.Accepted || r.Checks[len(r.Checks)-1].Code != "not_exact_comment_change" {
		t.Fatal("unrelated closure change accepted", r)
	}
}

func minimumCandidate(t *testing.T, files []BaseFile, mode string) string {
	t.Helper()
	dir := candidateFixture(t, files)
	replaceCandidate(t, dir, "pkg/catalog/catalog.go", "type Request struct {", "type Request struct {\n MinContext int `json:\"min_context,omitempty\"`")
	if mode == "field-only" {
		return dir
	}
	replaceCandidate(t, dir, "pkg/catalog/catalog.go", "if r.Tier < 0 ||", "if r.MinContext < 0 || r.Tier < 0 ||")
	op := "<"
	if mode == "wrong-boundary" {
		op = "<="
	}
	replaceCandidate(t, dir, "pkg/catalog/catalog.go", "if m.Quality < s.RequiredQuality {", "if m.Context "+op+" r.MinContext { reject(\"min_context_required\") }\n if m.Quality < s.RequiredQuality {")
	if mode == "wrong-reason" {
		replaceCandidate(t, dir, "pkg/catalog/catalog.go", "reject(\"min_context_required\")", "reject(\"context_limit\")")
	}
	return dir
}

func TestMinimumContextBehavior(t *testing.T) {
	files := pinnedFixture(t, "catalog-min-context")
	baseDir := candidateFixture(t, files)
	if r := verifyFixture(t, "catalog-min-context", files, baseDir); r.Accepted || r.Checks[len(r.Checks)-1].Code != "unchanged_base" {
		t.Fatal("unchanged baseline accepted", r)
	}
	for _, mode := range []string{"correct", "field-only", "wrong-boundary", "wrong-reason"} {
		t.Run(mode, func(t *testing.T) {
			dir := minimumCandidate(t, files, mode)
			// Candidate assertions never determine acceptance. Replacing its own
			// existing tests with a deliberate failure does not alter pinned tests.
			if err := os.WriteFile(filepath.Join(dir, "pkg/catalog/catalog_test.go"), []byte("package catalog\ninvalid candidate-owned assertion\n"), 0600); err != nil {
				t.Fatal(err)
			}
			r := verifyFixture(t, "catalog-min-context", files, dir)
			if !sandboxSupported() {
				if r.Accepted || r.Status != "verifier_unknown" {
					t.Fatal("unsupported isolation accepted", r)
				}
				return
			}
			if mode == "correct" {
				if !r.Accepted || !r.ExecutionIsolated || r.IndependentTests == 0 || r.CandidateTestsUsed {
					t.Fatal("correct behavior did not pass independent tests", r)
				}
			} else if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated {
				t.Fatal("plausible wrong behavior accepted or not tested", r)
			}
		})
	}
}

func TestPinnedInputsAndReadBounds(t *testing.T) {
	files := pinnedFixture(t, "comment-preview-authority")
	dir := candidateFixture(t, files)
	for _, mutate := range []func(*Request){
		func(r *Request) { r.BaseRevision = "untrusted" },
		func(r *Request) { r.BaseFiles[0].SHA256 = strings.Repeat("0", 64) },
		func(r *Request) { r.BaseFiles[0].Data = []byte("wrong public bytes") },
		func(r *Request) { r.BaseFiles = append(r.BaseFiles, r.BaseFiles[0]) },
		func(r *Request) { r.Timeout = MaxTimeout + time.Second },
		func(r *Request) { r.Timeout = -time.Second },
	} {
		req := Request{TaskID: "comment-preview-authority", BaseRevision: BaseRevision, BaseFiles: append([]BaseFile(nil), files...), CandidateDir: dir}
		mutate(&req)
		r, err := Verify(context.Background(), req)
		if err == nil || r.Accepted {
			t.Fatal("invalid input accepted", r, err)
		}
	}
	p := filepath.Join(dir, "pkg/reporouter/router.go")
	if err := os.WriteFile(p, bytes.Repeat([]byte("x"), MaxFileBytes+1), 0600); err != nil {
		t.Fatal(err)
	}
	r := verifyFixture(t, "comment-preview-authority", files, dir)
	if r.Accepted || r.Checks[len(r.Checks)-1].Code != "candidate_file_limit" {
		t.Fatal("oversized candidate accepted", r)
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("DO_NOT_RETURN_PRIVATE_SENTINEL"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, p); err != nil {
		t.Fatal(err)
	}
	r = verifyFixture(t, "comment-preview-authority", files, dir)
	b, err := json.Marshal(r)
	if err != nil || r.Accepted || bytes.Contains(b, []byte("PRIVATE_SENTINEL")) || bytes.Contains(b, []byte(dir)) || bytes.Contains(b, []byte(outside)) {
		t.Fatal("path escape or private report content", r)
	}
	unknown, e := Verify(context.Background(), Request{TaskID: "PRIVATE_SENTINEL_INVALID_TASK", BaseRevision: "PRIVATE_SENTINEL_INVALID_REVISION"})
	b, err = json.Marshal(unknown)
	if e == nil || err != nil || bytes.Contains(b, []byte("PRIVATE_SENTINEL")) {
		t.Fatal("unknown identifiers echoed into report")
	}
}

func TestMinimumFieldContract(t *testing.T) {
	files := pinnedFixture(t, "catalog-min-context")
	base := fileAt(files, "pkg/catalog/catalog.go")
	for _, field := range []string{
		"MinContext int `json:\"min_context,omitempty\"`",
		"MinContext int64 `json:\"min_context,omitempty\"`",
		"MinContext int `json:\"minimum_context,omitempty\"`",
		"Other int `json:\"min_context,omitempty\"`",
	} {
		data := bytes.Replace(base, []byte("type Request struct {"), []byte("type Request struct {\n"+field), 1)
		want := strings.HasPrefix(field, "MinContext int `json:\"min_context,omitempty\"`")
		if minimumField(data) != want {
			t.Fatal("minimum field contract", field)
		}
	}
}

func TestUnsupportedExecutionShapeIsUnknown(t *testing.T) {
	files := pinnedFixture(t, "catalog-min-context")
	for _, mode := range []string{"import", "init", "directive"} {
		t.Run(mode, func(t *testing.T) {
			dir := minimumCandidate(t, files, "correct")
			switch mode {
			case "import":
				replaceCandidate(t, dir, "pkg/catalog/catalog.go", "\"fmt\"", "\"fmt\"\n\"strconv\"")
			case "init":
				replaceCandidate(t, dir, "pkg/catalog/catalog.go", "type Request struct {", "func init() {}\n\ntype Request struct {")
			case "directive":
				replaceCandidate(t, dir, "pkg/catalog/catalog.go", "type Request struct {", "//go:generate taskverify-unused\ntype Request struct {")
			}
			r := verifyFixture(t, "catalog-min-context", files, dir)
			if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "unsupported_candidate_shape" {
				t.Fatal("unsupported execution shape classified as task failure or accepted", r)
			}
		})
	}
}

func TestOutputLimitCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := boundedOutput{cancel: cancel}
	if _, err := w.Write(bytes.Repeat([]byte("x"), maxOutputBytes+1)); err == nil || !w.overflow || ctx.Err() == nil || w.b.Len() != 0 {
		t.Fatal("output was not bounded and cancelled")
	}
}

func TestSandboxBlocksHostFilesNetworkAndInheritedSecrets(t *testing.T) {
	if !sandboxSupported() {
		t.Skip("behavior executor is unsupported; Verify reports verifier_unknown")
	}
	const envKey = "RIIDO_TASKVERIFY_PRIVATE_TEST_SENTINEL"
	t.Setenv(envKey, "private-authored-fixture")
	outside := filepath.Join(t.TempDir(), "outside-authored-sentinel")
	if err := os.WriteFile(outside, []byte("authored-not-a-real-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("local isolation probe unavailable")
	}
	defer listener.Close()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	toolchain, err := filepath.EvalSymlinks(runtime.GOROOT())
	if err != nil {
		t.Fatal(err)
	}
	cache, work := filepath.Join(dir, "cache"), filepath.Join(dir, "work")
	for _, p := range []string{cache, work} {
		if err := os.Mkdir(p, 0700); err != nil {
			t.Fatal(err)
		}
	}
	// Exit codes identify capability failures without printing the sentinel,
	// host paths or inherited environment into a persisted report.
	program := `package main
import ("net"; "os"; "time")
func main() {
 if _,err:=os.ReadFile(os.Args[1]);err==nil { os.Exit(42) }
 if os.Getenv("RIIDO_TASKVERIFY_PRIVATE_TEST_SENTINEL")!="" || os.Getenv("HOME")!="" || os.Getenv("CODEX_HOME")!="" { os.Exit(43) }
 c,err:=net.DialTimeout("tcp",os.Args[2],time.Second)
 if err==nil { c.Close(); os.Exit(44) }
}
`
	if err := os.WriteFile(filepath.Join(dir, "probe.go"), []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), MaxTimeout)
	defer cancel()
	cmd, err := sandboxCommand(ctx, filepath.Join(toolchain, "bin/go"), []string{"run", "probe.go", outside, listener.Addr().String()}, toolchain, dir)
	if err != nil {
		t.Fatal("sandbox setup")
	}
	cmd.Dir, cmd.Env = dir, isolatedEnv(toolchain, dir, cache, work)
	w := &boundedOutput{cancel: cancel}
	cmd.Stdout, cmd.Stderr = w, w
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil || w.overflow {
		t.Fatal("sandbox allowed a prohibited capability or failed its trusted probe")
	}
}

func TestCancellationCannotBecomeBehaviorAcceptance(t *testing.T) {
	files := pinnedFixture(t, "comment-budget-period")
	dir := candidateFixture(t, files)
	replaceCandidate(t, dir, "pkg/catalog/catalog.go", "// Snapshot for one caller-defined budget period. This is not a reservation ledger.", "// Snapshot supplied by the caller for a single budget period; no reservation is made.")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, err := Verify(ctx, Request{TaskID: "comment-budget-period", BaseRevision: BaseRevision, BaseFiles: files, CandidateDir: dir})
	if err != nil || r.Accepted || r.Status != "verifier_unknown" {
		t.Fatal("cancelled verification treated as a behavior result", r, err)
	}
}

func TestSpecBindsIndependentAcceptanceSource(t *testing.T) {
	spec, err := TaskSpec("catalog-min-context")
	if err != nil || spec.AcceptanceSourceSHA256 != digest([]byte(contractTests)) {
		t.Fatal("independent contract source is not pinned")
	}
	canonical, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	files := pinnedFixture(t, spec.ID)
	r := verifyFixture(t, spec.ID, files, candidateFixture(t, files))
	if r.SpecSHA256 != digest(canonical) {
		t.Fatal("report does not bind complete task spec")
	}
	spec.AcceptanceSourceSHA256 = strings.Repeat("0", 64)
	changed, err := json.Marshal(spec)
	if err != nil || digest(changed) == r.SpecSHA256 {
		t.Fatal("changed acceptance source does not change task-spec identity")
	}
}
