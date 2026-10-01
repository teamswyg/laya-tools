package taskverify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func keywordFixture(t *testing.T) ([]BaseFile, []BaseFile) {
	t.Helper()
	d, ok := TaskDefinition(keywordGuardTask)
	if !ok {
		t.Fatal("missing versioned definition")
	}
	var source, attribution []BaseFile
	for _, f := range d.Files {
		b, err := os.ReadFile(filepath.Join("testdata/repo-keyword-language-guard-v1", filepath.FromSlash(f.Path)))
		if err != nil || digest(b) != f.SHA256 {
			t.Fatal("versioned public snapshot pin mismatch")
		}
		file := BaseFile{Path: f.Path, SHA256: f.SHA256, Data: b}
		if f.Path == "LICENSE" || f.Path == "NOTICE" {
			attribution = append(attribution, file)
		} else {
			source = append(source, file)
		}
	}
	if _, err := validateBase(d.ID, d.BaseRevision, source); err != nil {
		t.Fatal(err)
	}
	if _, err := validateAttribution(d.ID, attribution); err != nil {
		t.Fatal(err)
	}
	return source, attribution
}

func verifyKeyword(t *testing.T, source, attribution []BaseFile, dir string) Report {
	t.Helper()
	r, err := Verify(context.Background(), Request{TaskID: keywordGuardTask, BaseRevision: keywordGuardRevision, BaseFiles: source, AttributionFiles: attribution, CandidateDir: dir, Timeout: MaxTimeout})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

const originalSummaryGuard = `for _, repo := range choices {
		if nonLatin(repo.Summary) {
			r.Reason = "metadata_language_unvalidated"
			return r, nil
		}
	}`

const authoredKeywordGuard = `for _, repo := range choices {
		unsupported := nonLatin(repo.Summary)
		for _, keyword := range repo.Keywords {
			unsupported = unsupported || nonLatin(keyword)
		}
		if unsupported {
			r.Reason = "metadata_language_unvalidated"
			return r, nil
		}
	}`

func keywordCandidate(t *testing.T, source []BaseFile, mode string) string {
	t.Helper()
	dir := candidateFixture(t, source)
	switch mode {
	case "comment-only":
		replaceCandidate(t, dir, "pkg/reporouter/router.go", "// This preview uses the installed English checkpoint.", "// This preview uses the installed English checkpoint only.")
	case "correct", "first-keyword-only", "only-keyword-positions-0-1-31", "only-first-two-selected-candidates", "all-repositories", "nil-judge-guard", "weakened-threshold":
		replaceCandidate(t, dir, "pkg/reporouter/router.go", originalSummaryGuard, authoredKeywordGuard)
		if mode == "first-keyword-only" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "range repo.Keywords", "range repo.Keywords[:min(1, len(repo.Keywords))]")
		}
		if mode == "only-keyword-positions-0-1-31" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "for _, keyword := range repo.Keywords {", "for position, keyword := range repo.Keywords {\nif position != 0 && position != 1 && position != 31 { continue }")
		}
		if mode == "only-first-two-selected-candidates" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "for _, repo := range choices {", "for _, repo := range choices[:min(2, len(choices))] {")
		}
		if mode == "all-repositories" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "for _, repo := range choices {", "for _, document := range idx.docs {\nrepo := document.repo")
		}
		if mode == "nil-judge-guard" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "if j == nil {", authoredKeywordGuard+"\nif j == nil {")
		}
		if mode == "weakened-threshold" {
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "return Config{4, .9, .15}", "return Config{4, .5, .15}")
		}
	default:
		t.Fatal("unknown authored candidate mode")
	}
	return dir
}

func TestFrozenOriginalTaskSpecs(t *testing.T) {
	for _, tc := range []struct {
		id, sha string
		paths   []string
	}{
		{"comment-preview-authority", "a01f1c95e818511d6bcac35a0eb20b0b6c943a73b4fe7652a550bb0036dbd747", []string{"pkg/reporouter/router.go"}},
		{"comment-budget-period", "00333f0814e4bcb28bc06b9172f21672dace61eea3f6537af40dfe77a3c59fe4", []string{"go.mod", "pkg/catalog/catalog.go", "pkg/catalog/catalog_test.go"}},
		{"catalog-min-context", "884eb30b9310817c78fe8071e20881b865bb61a603a75327e83d39f6625d7e25", []string{"go.mod", "pkg/catalog/catalog.go", "pkg/catalog/catalog_test.go", "pkg/planner/planner.go", "pkg/planner/planner_test.go", "pkg/switchpolicy/policy.go", "pkg/switchpolicy/policy_test.go", "examples/planner/config.json", "examples/planner/request.json"}},
	} {
		s, err := TaskSpec(tc.id)
		b, marshalErr := json.Marshal(s)
		paths, pathErr := BasePaths(tc.id)
		if err != nil || marshalErr != nil || pathErr != nil || digest(b) != tc.sha || s.BaseRevision != "6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3" || !slices.Equal(paths, tc.paths) {
			t.Fatal("frozen original task bytes or closure changed", tc.id)
		}
	}
}

func TestVersionedDefinitionOwnsPinsAndContract(t *testing.T) {
	source, attribution := keywordFixture(t)
	d, _ := TaskDefinition(keywordGuardTask)
	spec, err := TaskSpec(d.ID)
	b, marshalErr := json.Marshal(d)
	if err != nil || marshalErr != nil || spec.DefinitionSHA256 != digest(b) || spec.AcceptanceSourceSHA256 != digest([]byte(keywordGuardContractTests)) || d.PromptSHA256 != digest([]byte(spec.Prompt)) {
		t.Fatal("definition does not bind source, prompt and independent acceptance")
	}
	specBytes, err := json.Marshal(spec)
	if err != nil || digest(specBytes) != "fb71960dd44348b9f7531ba3c7b44e8bb88c89196bd43dc362328120efd2a644" || spec.DefinitionSHA256 != "f1782b575818b0ea4e8c29e610c21f25a2c44d4cfdca8270c2664d1cddda7bb2" || d.ContractSHA256 != "e44ea077ad9fdb09775e49694da4aecaa630070a302b3763a7599651775bc89a" || d.PromptSHA256 != "bd43cebca6b00049f88e9010405068e8c57908e91285a115f23b30db7b3e1ab7" {
		t.Fatal("version 1 task definition changed in place; create a new version instead")
	}
	d.Files[0].SHA256 = "changed"
	d.RequiredTests[0].Test = "changed"
	d.Imports[0] = "changed"
	d.Packages[0] = "changed"
	d.MutablePaths[0] = "changed"
	fresh, _ := TaskDefinition(keywordGuardTask)
	if fresh.Files[0].SHA256 == "changed" || fresh.RequiredTests[0].Test == "changed" || fresh.Imports[0] == "changed" || fresh.Packages[0] == "changed" || fresh.MutablePaths[0] == "changed" {
		t.Fatal("caller mutated immutable registry")
	}
	paths, _ := BasePaths(keywordGuardTask)
	if len(paths) != 3 || slices.Contains(paths, "LICENSE") || slices.Contains(paths, "NOTICE") {
		t.Fatal("source closure includes duplicate attribution staging paths")
	}
	r := verifyKeyword(t, source, attribution, candidateFixture(t, source))
	if r.Accepted || r.Status != "rejected" || r.Checks[len(r.Checks)-1].Code != "unchanged_base" || len(r.AttributionFiles) != 2 || r.AttributionSHA256 == "" {
		t.Fatal("baseline accepted or original attribution evidence missing", r)
	}
	if sandboxSupported() {
		outcome := isolatedTests(context.Background(), source, keywordGuardTask, MaxTimeout, "")
		if outcome.passed || outcome.unknown || !outcome.isolated {
			t.Fatal("independent new behavior contract did not reject unchanged public baseline", outcome)
		}
	}
}

func TestVersionedKeywordGuardBehavior(t *testing.T) {
	source, attribution := keywordFixture(t)
	for _, mode := range []string{"correct", "comment-only", "first-keyword-only", "only-keyword-positions-0-1-31", "only-first-two-selected-candidates", "all-repositories", "nil-judge-guard", "weakened-threshold"} {
		t.Run(mode, func(t *testing.T) {
			dir := keywordCandidate(t, source, mode)
			// Candidate assertions are ignored: only immutable base tests and the
			// separately supplied contract are present in the execution module.
			if err := os.WriteFile(filepath.Join(dir, "pkg/reporouter/router_test.go"), []byte("package reporouter\ninvalid candidate-owned assertion\n"), 0600); err != nil {
				t.Fatal(err)
			}
			r := verifyKeyword(t, source, attribution, dir)
			if !sandboxSupported() {
				if r.Accepted || r.Status != "verifier_unknown" {
					t.Fatal("unsupported isolation claimed behavior result", r)
				}
				return
			}
			if mode == "correct" {
				if !r.Accepted || r.Status != "accepted" || !r.ExecutionIsolated || r.IndependentTests < 5 || r.CandidateTestsUsed {
					t.Fatal("independently authored behavior was not accepted", r)
				}
			} else if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated {
				t.Fatal("seeded behavioral error was accepted or not independently executed", r)
			}
		})
	}
}

func TestVersionedRevisionAttributionAndScope(t *testing.T) {
	source, attribution := keywordFixture(t)
	dir := keywordCandidate(t, source, "correct")
	for _, mode := range []string{"old-revision", "missing-attribution", "wrong-attribution", "duplicate-attribution", "wrong-source"} {
		req := Request{TaskID: keywordGuardTask, BaseRevision: keywordGuardRevision, BaseFiles: slices.Clone(source), AttributionFiles: slices.Clone(attribution), CandidateDir: dir}
		switch mode {
		case "old-revision":
			req.BaseRevision = BaseRevision
		case "missing-attribution":
			req.AttributionFiles = nil
		case "wrong-attribution":
			req.AttributionFiles[0].Data = []byte("wrong original public license bytes")
		case "duplicate-attribution":
			req.AttributionFiles[1] = req.AttributionFiles[0]
		case "wrong-source":
			req.BaseFiles[0].SHA256 = strings.Repeat("0", 64)
		}
		if r, err := Verify(context.Background(), req); err == nil || r.Accepted {
			t.Fatal("untrusted versioned public inputs accepted", mode)
		}
	}
	replaceCandidate(t, dir, "go.mod", "go 1.27.1", "go 1.27.0")
	r := verifyKeyword(t, source, attribution, dir)
	if r.Accepted || r.Status != "rejected" || r.Checks[len(r.Checks)-1].Code != "unrelated_closure_source_changed" {
		t.Fatal("unrelated source closure change accepted", r)
	}
	for _, id := range []string{"comment-preview-authority", "comment-budget-period", "catalog-min-context", keywordGuardTask} {
		pins, err := AttributionPins(id)
		if err != nil || len(pins) != 2 || pins[0].Path != "LICENSE" || pins[1].Path != "NOTICE" {
			t.Fatal("task attribution pins missing", id)
		}
		pins[0].SHA256 = "changed"
		again, _ := AttributionPins(id)
		if again[0].SHA256 == "changed" {
			t.Fatal("caller mutated attribution pins")
		}
	}
	if _, err := AttributionPins("unknown"); err == nil {
		t.Fatal("unknown task attribution supplied")
	}
}

func TestVersionedExecutionShapeUnknown(t *testing.T) {
	source, attribution := keywordFixture(t)
	for _, mode := range []string{"import", "init", "directive", "package"} {
		dir := keywordCandidate(t, source, "correct")
		switch mode {
		case "import":
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "\"fmt\"", "\"fmt\"\n\"strconv\"")
		case "init":
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "type Repository struct {", "func init() {}\ntype Repository struct {")
		case "directive":
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "type Repository struct {", "//go:generate unused\ntype Repository struct {")
		case "package":
			replaceCandidate(t, dir, "pkg/reporouter/router.go", "package reporouter", "package other")
		}
		r := verifyKeyword(t, source, attribution, dir)
		if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "unsupported_candidate_shape" {
			t.Fatal("unsupported versioned shape classified as behavior failure", mode, r)
		}
	}
}

func TestVersionedTerminalPasses(t *testing.T) {
	d, _ := TaskDefinition(keywordGuardTask)
	var output bytes.Buffer
	emit := func(action, pkg, test string) {
		fmt.Fprintf(&output, "{\"Action\":%q,\"Package\":%q,\"Test\":%q}\n", action, pkg, test)
	}
	emit("start", d.Packages[0], "")
	for _, tt := range d.RequiredTests {
		emit("run", tt.Package, tt.Test)
		emit("pass", tt.Package, tt.Test)
	}
	emit("pass", d.Packages[0], "")
	complete := slices.Clone(output.Bytes())
	if r := versionedTerminalPasses(d, complete); !r.passed || r.unknown || r.tests != len(d.RequiredTests) {
		t.Fatal("complete exact terminal package/test evidence rejected", r)
	}
	for _, mode := range []string{"missing-test", "skipped", "package-only", "pass-without-run", "wrong-package", "late-test", "duplicate-test", "invalid-json"} {
		data := slices.Clone(complete)
		switch mode {
		case "missing-test":
			data = bytes.ReplaceAll(data, []byte(fmt.Sprintf("{\"Action\":\"pass\",\"Package\":%q,\"Test\":%q}\n", d.RequiredTests[0].Package, d.RequiredTests[0].Test)), nil)
		case "skipped":
			data = bytes.Replace(data, []byte("\"Action\":\"pass\""), []byte("\"Action\":\"skip\""), 1)
		case "package-only":
			data = []byte(fmt.Sprintf("{\"Action\":\"start\",\"Package\":%q}\n{\"Action\":\"pass\",\"Package\":%q}\n", d.Packages[0], d.Packages[0]))
		case "pass-without-run":
			data = bytes.ReplaceAll(data, []byte(fmt.Sprintf("{\"Action\":\"run\",\"Package\":%q,\"Test\":%q}\n", d.RequiredTests[0].Package, d.RequiredTests[0].Test)), nil)
		case "wrong-package":
			data = bytes.ReplaceAll(data, []byte(d.Packages[0]), []byte("unknown/package"))
		case "late-test":
			data = append(data, []byte(fmt.Sprintf("{\"Action\":\"run\",\"Package\":%q,\"Test\":%q}\n", d.Packages[0], d.RequiredTests[0].Test))...)
		case "duplicate-test":
			data = bytes.Replace(data, []byte("\"Action\":\"pass\""), []byte("\"Action\":\"run\""), 1)
		case "invalid-json":
			data = append(data, []byte("not JSON\n")...)
		}
		if r := versionedTerminalPasses(d, data); r.passed {
			t.Fatal("incomplete/incorrect terminal evidence accepted", mode)
		}
	}
}
