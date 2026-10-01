package taskverify

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func eventKeyFixture(t *testing.T) ([]BaseFile, []BaseFile) {
	t.Helper()
	d, ok := TaskDefinition(eventKeyBoundsTask)
	if !ok {
		t.Fatal("missing independent envelope definition")
	}
	var source, attribution []BaseFile
	for _, pin := range d.Files {
		b, err := os.ReadFile(filepath.Join("testdata/taskoutcome-event-key-bounds-v1", filepath.FromSlash(pin.Path)))
		if err != nil || digest(b) != pin.SHA256 {
			t.Fatal("original public envelope fixture pin mismatch")
		}
		f := BaseFile{Path: pin.Path, SHA256: pin.SHA256, Data: b}
		if pin.Path == "LICENSE" || pin.Path == "NOTICE" {
			attribution = append(attribution, f)
		} else {
			source = append(source, f)
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

func verifyEventKeys(t *testing.T, source, attribution []BaseFile, dir string) Report {
	t.Helper()
	r, err := Verify(context.Background(), Request{TaskID: eventKeyBoundsTask, BaseRevision: eventKeyBoundsRevision, BaseFiles: source, AttributionFiles: attribution, CandidateDir: dir, Timeout: MaxTimeout})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

const originalEventNameGuard = `name, ok := key.(string)
		if !ok {
			return e, ErrJSON
		}
		var bit uint8`

const authoredEventNameGuard = `name, ok := key.(string)
		if !ok || len(name) > 128 || eventKeyCount == len(eventKeys) {
			return e, ErrJSON
		}
		for _, previous := range eventKeys[:eventKeyCount] {
			if previous == name {
				return e, ErrDuplicate
			}
		}
		eventKeys[eventKeyCount] = name
		eventKeyCount++
		var bit uint8`

func eventKeyCandidate(t *testing.T, source []BaseFile, mode string) string {
	t.Helper()
	dir := candidateFixture(t, source)
	p := "pkg/taskoutcome/summary.go"
	if mode == "comment-only" {
		replaceCandidate(t, dir, p, "// Sequential thread.started blocks are allowed, but turns cannot overlap.", "// Sequential thread.started blocks are allowed; turns cannot overlap.")
		return dir
	}
	replaceCandidate(t, dir, p, "var rawUsage json.RawMessage", "var rawUsage json.RawMessage\nvar eventKeys [64]string\neventKeyCount := 0")
	replaceCandidate(t, dir, p, originalEventNameGuard, authoredEventNameGuard)
	switch mode {
	case "correct":
	case "unknown-duplicates-ignored":
		replaceCandidate(t, dir, p, "for _, previous := range eventKeys[:eventKeyCount] {", "for _, previous := range eventKeys[:eventKeyCount] {\nif name != \"type\" && name != \"usage\" && name != \"thread_id\" && name != \"item\" && name != \"error\" { break }")
	case "guard-only-after-type":
		replaceCandidate(t, dir, p, authoredEventNameGuard, `name, ok := key.(string)
        if !ok { return e, ErrJSON }
        if name == "type" || e.kind != "" {
            if len(name) > 128 || eventKeyCount == len(eventKeys) { return e, ErrJSON }
            for _, previous := range eventKeys[:eventKeyCount] {
                if previous == name { return e, ErrDuplicate }
            }
            eventKeys[eventKeyCount] = name
            eventKeyCount++
        }
        var bit uint8`)
	case "new-startup-envelope-rules":
		replaceCandidate(t, dir, p, "kind := \"\"", "kind := \"\"\nvar startupKeys [256]string\nstartupKeyCount := 0")
		replaceCandidate(t, dir, p, "name, ok := key.(string)\n\t\tif !ok {\n\t\t\treturn false, ErrJSON\n\t\t}", "name, ok := key.(string)\nif !ok || len(name) > 128 { return false, ErrJSON }\nfor _, previous := range startupKeys[:startupKeyCount] { if previous == name { return false, ErrDuplicate } }\nstartupKeys[startupKeyCount] = name\nstartupKeyCount++")
	case "usage-bound-wrong-error":
		replaceCandidate(t, dir, p, "if keyCount == len(keys) || len(key) > 128 {\n\t\t\treturn values, ErrUsage", "if keyCount == len(keys) || len(key) > 128 {\n\t\t\treturn values, ErrJSON")
	case "unknown-envelope-retained", "unknown-envelope-value-retained", "hidden-unknown-envelope-retained":
		tag := "raw_unknown,omitempty"
		if mode == "hidden-unknown-envelope-retained" {
			tag = "-"
		}
		replaceCandidate(t, dir, p, "type Summary struct {", "type Summary struct {\nRawUnknown string `json:\""+tag+"\"`")
		replaceCandidate(t, dir, p, "type event struct {", "type event struct {\nunknown string")
		replaceCandidate(t, dir, p, "s.Events++", "s.Events++\ns.RawUnknown += e.unknown")
		retention := "e.unknown += name"
		if mode == "unknown-envelope-value-retained" {
			retention = "if len(value) > 0 && value[0] != '\"' { e.unknown += string(value) }"
		}
		replaceCandidate(t, dir, p, "case \"item\":\n\t\t\te.item = value", "case \"item\":\n\t\t\te.item = value\ndefault:\n"+retention)
	case "count63-strict":
		replaceCandidate(t, dir, p, "var eventKeys [64]string", "var eventKeys [63]string")
	case "count65-loose":
		replaceCandidate(t, dir, p, "var eventKeys [64]string", "var eventKeys [65]string")
	case "control-keys-not-counted":
		replaceCandidate(t, dir, p, "eventKeys[eventKeyCount] = name\n\t\teventKeyCount++", "if name != \"usage\" && name != \"thread_id\" && name != \"item\" && name != \"error\" {\neventKeys[eventKeyCount] = name\neventKeyCount++\n}")
	case "key127-strict":
		replaceCandidate(t, dir, p, "len(name) > 128", "len(name) > 127")
	case "key129-loose":
		replaceCandidate(t, dir, p, "len(name) > 128", "len(name) > 129")
	case "runes-instead-of-bytes":
		replaceCandidate(t, dir, p, "len(name) > 128", "utf8.RuneCountInString(name) > 128")
	case "case-folded-duplicates":
		replaceCandidate(t, dir, p, "previous == name", "strings.EqualFold(previous, name)")
	case "wrong-duplicate-error":
		replaceCandidate(t, dir, p, "if previous == name {\n\t\t\t\treturn e, ErrDuplicate", "if previous == name {\n\t\t\t\treturn e, ErrJSON")
	case "shared-event-state":
		replaceCandidate(t, dir, p, "var eventKeys [64]string\n\teventKeyCount := 0", "// Incorrect shared event state is declared at package scope.")
		replaceCandidate(t, dir, p, "func parseEvent(body []byte)", "var eventKeys [64]string\nvar eventKeyCount int\n\nfunc parseEvent(body []byte)")
	default:
		t.Fatal("unknown authored envelope candidate")
	}
	return dir
}

func TestFrozenKeywordDefinitionV1(t *testing.T) {
	d, _ := TaskDefinition(keywordGuardTask)
	s, err := TaskSpec(keywordGuardTask)
	b, marshalErr := json.Marshal(s)
	contract, present := definitionContractSource(keywordGuardTask)
	if err != nil || marshalErr != nil || !present || digest(b) != "fb71960dd44348b9f7531ba3c7b44e8bb88c89196bd43dc362328120efd2a644" || s.DefinitionSHA256 != "f1782b575818b0ea4e8c29e610c21f25a2c44d4cfdca8270c2664d1cddda7bb2" || d.ContractSHA256 != "e44ea077ad9fdb09775e49694da4aecaa630070a302b3763a7599651775bc89a" || d.PromptSHA256 != "bd43cebca6b00049f88e9010405068e8c57908e91285a115f23b30db7b3e1ab7" || digest([]byte(contract)) != d.ContractSHA256 {
		t.Fatal("adding another family changed frozen keyword v1 bytes/hash")
	}
}

func TestEventKeyDefinitionPins(t *testing.T) {
	source, attribution := eventKeyFixture(t)
	d, _ := TaskDefinition(eventKeyBoundsTask)
	s, err := TaskSpec(d.ID)
	b, marshalErr := json.Marshal(d)
	contract, present := definitionContractSource(d.ID)
	if err != nil || marshalErr != nil || !present || s.DefinitionSHA256 != digest(b) || d.ContractSHA256 != digest([]byte(contract)) || d.PromptSHA256 != digest([]byte(s.Prompt)) || s.AcceptanceSourceSHA256 != d.ContractSHA256 {
		t.Fatal("envelope source/prompt/contract identity is not bound")
	}
	specBytes, err := json.Marshal(s)
	if err != nil || digest(specBytes) != "169cedefd41d3a5ba05d11bbb4685761a0dde6adf07401b846b77bf5a84c27db" || s.DefinitionSHA256 != "408255ae28f7d6bfb8994d4a9cbcbcaa46edd95c1cf2f606fe847cd28e0599a9" || d.ContractSHA256 != "c9873cabd6b921516103bd8bd8bde8987d02c9aca748522c141aff1005896b71" || d.PromptSHA256 != "f395e7c85eb8398764ce189fca23709465c372a8710ba1be2583f0d99690a567" {
		t.Fatal("published envelope version1 identity changed in place")
	}
	paths, _ := BasePaths(d.ID)
	if !slices.Equal(paths, []string{"go.mod", "pkg/taskoutcome/summary.go", "pkg/taskoutcome/summary_test.go"}) || s.BaseRevision != eventKeyBoundsRevision || d.License != "Apache-2.0" || len(d.RequiredTests) != 6 {
		t.Fatal("new task used the wrong closure or test package")
	}
	d.Files[0].SHA256 = "changed"
	d.RequiredTests[0].Test = "changed"
	d.Imports[0] = "changed"
	fresh, _ := TaskDefinition(eventKeyBoundsTask)
	if fresh.Files[0].SHA256 == "changed" || fresh.RequiredTests[0].Test == "changed" || fresh.Imports[0] == "changed" {
		t.Fatal("new registry shares caller-mutable data")
	}
	r := verifyEventKeys(t, source, attribution, candidateFixture(t, source))
	if r.Accepted || r.Status != "rejected" || r.Checks[len(r.Checks)-1].Code != "unchanged_base" || r.AttributionSHA256 == "" {
		t.Fatal("unchanged envelope baseline accepted or attribution unbound", r)
	}
	if sandboxSupported() {
		outcome := isolatedTests(context.Background(), source, eventKeyBoundsTask, MaxTimeout, "")
		if outcome.passed || outcome.unknown || !outcome.isolated {
			t.Fatal("independent envelope contract did not reject baseline", outcome)
		}
	}
}

func TestEventKeyBoundaryBehavior(t *testing.T) {
	source, attribution := eventKeyFixture(t)
	for _, mode := range []string{"correct", "comment-only", "unknown-duplicates-ignored", "guard-only-after-type", "new-startup-envelope-rules", "usage-bound-wrong-error", "unknown-envelope-retained", "unknown-envelope-value-retained", "hidden-unknown-envelope-retained", "count63-strict", "count65-loose", "control-keys-not-counted", "key127-strict", "key129-loose", "runes-instead-of-bytes", "case-folded-duplicates", "wrong-duplicate-error", "shared-event-state"} {
		t.Run(mode, func(t *testing.T) {
			dir := eventKeyCandidate(t, source, mode)
			if err := os.WriteFile(filepath.Join(dir, "pkg/taskoutcome/summary_test.go"), []byte("package taskoutcome\ninvalid candidate assertions\n"), 0600); err != nil {
				t.Fatal(err)
			}
			r := verifyEventKeys(t, source, attribution, dir)
			if !sandboxSupported() {
				if r.Accepted || r.Status != "verifier_unknown" {
					t.Fatal("unsupported platform claimed envelope outcome", r)
				}
				return
			}
			if mode == "correct" {
				if !r.Accepted || r.Status != "accepted" || !r.ExecutionIsolated || r.CandidateTestsUsed || r.IndependentTests < 6 {
					t.Fatal("independent correct envelope candidate did not pass", r)
				}
			} else if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated {
				t.Fatal("seeded parser/state boundary error was accepted or untested", mode, r)
			}
		})
	}
}

func TestEventKeyBaseAndExecutionShape(t *testing.T) {
	source, attribution := eventKeyFixture(t)
	dir := eventKeyCandidate(t, source, "correct")
	for _, mode := range []string{"old-revision", "wrong-attribution", "wrong-source"} {
		req := Request{TaskID: eventKeyBoundsTask, BaseRevision: eventKeyBoundsRevision, BaseFiles: slices.Clone(source), AttributionFiles: slices.Clone(attribution), CandidateDir: dir}
		switch mode {
		case "old-revision":
			req.BaseRevision = keywordGuardRevision
		case "wrong-attribution":
			req.AttributionFiles[0].SHA256 = strings.Repeat("0", 64)
		case "wrong-source":
			req.BaseFiles[0].Data = []byte("wrong public bytes")
		}
		if r, err := Verify(context.Background(), req); err == nil || r.Accepted {
			t.Fatal("incorrect new-family base was trusted", mode)
		}
	}
	for _, mode := range []string{"import", "init", "directive", "package"} {
		dir := eventKeyCandidate(t, source, "correct")
		p := "pkg/taskoutcome/summary.go"
		switch mode {
		case "import":
			replaceCandidate(t, dir, p, "\"bufio\"", "\"bufio\"\n\"fmt\"")
		case "init":
			replaceCandidate(t, dir, p, "type ErrorCode string", "func init() {}\ntype ErrorCode string")
		case "directive":
			replaceCandidate(t, dir, p, "type ErrorCode string", "//go:generate unused\ntype ErrorCode string")
		case "package":
			replaceCandidate(t, dir, p, "package taskoutcome", "package other")
		}
		r := verifyEventKeys(t, source, attribution, dir)
		if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "unsupported_candidate_shape" {
			t.Fatal("unsupported parser source shape mislabeled", mode, r)
		}
	}
}
