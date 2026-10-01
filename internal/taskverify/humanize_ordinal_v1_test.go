package taskverify

import (
	"context"
	"crypto/sha1" // Git's public blob identity, not an authentication primitive.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func humanizeOrdinalFixture(t *testing.T) ([]BaseFile, []BaseFile) {
	t.Helper()
	d, ok := TaskDefinition(humanizeOrdinalTask)
	if !ok {
		t.Fatal("missing external ordinal definition")
	}
	gitBlobs := []struct{ path, sha1 string }{
		{"go.mod", "611d24fa1a8560ca99ba83263ec3d6a12d85c3e9"},
		{"ordinals.go", "43d88a861950eac85b0f742a59621f92345d7109"},
		{"ordinals_test.go", "c478d5c9bdafae580e0bc7f6283766ae6c48d4a6"},
		{"common_test.go", "fc7db151640401c8b56bce644e294076bdff9fe0"},
		{"LICENSE", "8d9a94a90680d9fc114a1b3a2b4123c233c324af"},
	}
	var source, attribution []BaseFile
	for _, pin := range d.Files {
		b, err := os.ReadFile(filepath.Join("testdata/go53-humanize-ordinal64-v1", pin.Path))
		if err != nil || digest(b) != pin.SHA256 {
			t.Fatal("external original SHA256 mismatch")
		}
		var wantBlob string
		for _, original := range gitBlobs {
			if original.path == pin.Path {
				wantBlob = original.sha1
			}
		}
		blob := sha1.New()
		fmt.Fprintf(blob, "blob %d%c", len(b), 0)
		blob.Write(b)
		if wantBlob == "" || hex.EncodeToString(blob.Sum(nil)) != wantBlob {
			t.Fatal("external public Git blob mismatch")
		}
		f := BaseFile{Path: pin.Path, SHA256: pin.SHA256, Data: b}
		if pin.Path == "LICENSE" {
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

const humanizeOrdinalAuthoredCorrect = `
func Ordinal64(x int64) string {
	prefix := strconv.FormatInt(x, 10)
	tail := prefix
	if len(tail) > 2 {
		tail = tail[len(tail)-2:]
	}
	suffix := "th"
	if tail != "11" && tail != "12" && tail != "13" {
		switch prefix[len(prefix)-1] {
		case '1':
			suffix = "st"
		case '2':
			suffix = "nd"
		case '3':
			suffix = "rd"
		}
	}
	return prefix + suffix
}
`

func humanizeOrdinalTemporaryFixture(t *testing.T, source []BaseFile) string {
	t.Helper()
	b, err := os.ReadFile("testdata/go53-humanize-ordinal64-v1/LICENSE")
	if err != nil || digest(b) != "a973b4498c13eb74baa2a8e5c351426a6826f2fcdd909916dbe53ee2e755fd71" {
		t.Fatal("temporary source copy lost original MIT license")
	}
	files := append(slices.Clone(source), BaseFile{Path: "LICENSE", SHA256: digest(b), Data: b})
	return candidateFixture(t, files)
}

func humanizeOrdinalCandidate(t *testing.T, source []BaseFile, mode string) string {
	t.Helper()
	dir := humanizeOrdinalTemporaryFixture(t, source)
	p := "ordinals.go"
	if mode == "comment-only" {
		replaceCandidate(t, dir, p, "// Ordinal gives you", "// The original Ordinal gives you")
		return dir
	}
	b := fileAt(source, p)
	if err := os.WriteFile(filepath.Join(dir, p), append(slices.Clone(b), []byte(humanizeOrdinalAuthoredCorrect)...), 0600); err != nil {
		t.Fatal(err)
	}
	// Format the complete candidate before applying uniquely matched mutations.
	replaceCandidate(t, dir, p, "func Ordinal64", "func Ordinal64")
	switch mode {
	case "correct":
	case "no-teen-exception":
		replaceCandidate(t, dir, p, `if tail != "11" && tail != "12" && tail != "13" {`, `if tail != "never" {`)
	case "teen-exception-only-small":
		replaceCandidate(t, dir, p, "tail = tail[len(tail)-2:]", "if x < 0 { tail = tail[1:] }")
	case "negative-always-th":
		replaceCandidate(t, dir, p, "prefix := strconv.FormatInt(x, 10)", "prefix := strconv.FormatInt(x, 10)\nif x < 0 { return prefix + \"th\" }")
	case "sign-dropped":
		replaceCandidate(t, dir, p, "prefix := strconv.FormatInt(x, 10)", "prefix := strconv.FormatInt(x, 10)\nif prefix[0] == '-' { prefix = prefix[1:] }")
	case "absolute-min-overflow":
		replaceCandidate(t, dir, p, "prefix := strconv.FormatInt(x, 10)", "sign := \"\"\nif x < 0 { sign = \"-\"; x = -x }\nprefix := strconv.FormatInt(x, 10)")
		replaceCandidate(t, dir, p, "return prefix + suffix", "return sign + prefix + suffix")
	case "int32-truncation":
		replaceCandidate(t, dir, p, "strconv.FormatInt(x, 10)", "strconv.FormatInt(int64(int32(x)), 10)")
	case "float-rounded-digits":
		replaceCandidate(t, dir, p, "strconv.FormatInt(x, 10)", "strconv.FormatFloat(float64(x), 'f', 0, 64)")
	case "unsigned-digits":
		replaceCandidate(t, dir, p, "strconv.FormatInt(x, 10)", "strconv.FormatUint(uint64(x), 10)")
	case "swapped-st-nd":
		replaceCandidate(t, dir, p, "case '1':\n\t\t\tsuffix = \"st\"", "case '1':\n\t\t\tsuffix = \"nd\"")
	case "legacy-negative-changed":
		replaceCandidate(t, dir, p, "func Ordinal(x int) string {", "func Ordinal(x int) string {\nif x < 0 { return Ordinal64(int64(x)) }")
	case "wrong-signature":
		replaceCandidate(t, dir, p, "func Ordinal64(x int64) string", "func Ordinal64(x int) string")
	case "shared-last-result":
		replaceCandidate(t, dir, p, "func Ordinal64(x int64) string {", "var authoredOrdinalMemo string\nfunc Ordinal64(x int64) string {\nif authoredOrdinalMemo != \"\" { return authoredOrdinalMemo }")
		replaceCandidate(t, dir, p, "return prefix + suffix", "authoredOrdinalMemo = prefix + suffix\nreturn authoredOrdinalMemo")
	default:
		t.Fatal("unknown independent ordinal candidate")
	}
	return dir
}

func verifyHumanizeOrdinal(t *testing.T, source, attribution []BaseFile, dir string) Report {
	t.Helper()
	r, err := Verify(context.Background(), Request{TaskID: humanizeOrdinalTask, BaseRevision: humanizeOrdinalRevision, BaseFiles: source, AttributionFiles: attribution, CandidateDir: dir, Timeout: MaxTimeout})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestFrozenEventKeyDefinitionV1(t *testing.T) {
	d, _ := TaskDefinition(eventKeyBoundsTask)
	s, err := TaskSpec(eventKeyBoundsTask)
	b, marshalErr := json.Marshal(s)
	contract, present := definitionContractSource(eventKeyBoundsTask)
	if err != nil || marshalErr != nil || !present || digest(b) != "169cedefd41d3a5ba05d11bbb4685761a0dde6adf07401b846b77bf5a84c27db" || s.DefinitionSHA256 != "408255ae28f7d6bfb8994d4a9cbcbcaa46edd95c1cf2f606fe847cd28e0599a9" || d.ContractSHA256 != "c9873cabd6b921516103bd8bd8bde8987d02c9aca748522c141aff1005896b71" || d.PromptSHA256 != "f395e7c85eb8398764ce189fca23709465c372a8710ba1be2583f0d99690a567" || digest([]byte(contract)) != d.ContractSHA256 {
		t.Fatal("external-module support changed frozen parser v1 identity")
	}
}

func TestHumanizeOrdinalDefinitionPins(t *testing.T) {
	source, attribution := humanizeOrdinalFixture(t)
	d, _ := TaskDefinition(humanizeOrdinalTask)
	s, err := TaskSpec(d.ID)
	b, marshalErr := json.Marshal(d)
	contract, present := definitionContractSource(d.ID)
	if err != nil || marshalErr != nil || !present || s.DefinitionSHA256 != digest(b) || d.ContractSHA256 != digest([]byte(contract)) || d.PromptSHA256 != digest([]byte(s.Prompt)) || s.AcceptanceSourceSHA256 != d.ContractSHA256 {
		t.Fatal("external ordinal source/prompt/contract identity is not bound")
	}
	specBytes, specErr := json.Marshal(s)
	if specErr != nil || digest(specBytes) != "5c6fd6058a48bd0e6441a68c351e7b1ed491aec20baff78c6942e377ae79f5d5" || s.DefinitionSHA256 != "99a7c0723e3c9f40a069003ccf3f9c2e290c3cbe8e67be3a4ced02d4a31a29b8" || d.ContractSHA256 != "3b1351162b746fa61f7837af2923644eb2f4df100de643114c9d9f0eb064fd06" || d.PromptSHA256 != "107ae4d6da023114cc615ebdef6709053be8471bb0ee37f9763af73f279a94a9" {
		t.Fatal("ordinal version1 source/spec/contract identity changed in place")
	}
	paths, _ := BasePaths(d.ID)
	pins, pinErr := AttributionPins(d.ID)
	if pinErr != nil || len(pins) != 1 || pins[0].Path != "LICENSE" || pins[0].SHA256 != d.Files[4].SHA256 || !slices.Equal(paths, []string{"go.mod", "ordinals.go", "ordinals_test.go", "common_test.go"}) || d.ModulePath != "github.com/dustin/go-humanize" || d.License != "MIT" || len(d.RequiredTests) != 6 || s.BaseRevision != humanizeOrdinalRevision || s.Prompt != humanizeOrdinalPrompt {
		t.Fatal("external task lost its minimal closure, actual license or module identity")
	}
	// The exact public baseline has only the original int API. It has no new
	// Ordinal64 symbol even though its original legacy test source is preserved.
	f, parseErr := parser.ParseFile(token.NewFileSet(), "ordinals.go", fileAt(source, "ordinals.go"), 0)
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	legacy, newAPI := false, false
	for _, declaration := range f.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok {
			legacy = legacy || fn.Name.Name == "Ordinal"
			newAPI = newAPI || fn.Name.Name == "Ordinal64"
		}
	}
	if !legacy || newAPI {
		t.Fatal("the frozen source's baseline gap is not the claimed new API")
	}
	r := verifyHumanizeOrdinal(t, source, attribution, humanizeOrdinalTemporaryFixture(t, source))
	if r.Accepted || r.Status != "rejected" || r.Checks[len(r.Checks)-1].Code != "unchanged_base" || r.AttributionSHA256 == "" {
		t.Fatal("external unchanged baseline accepted or license unbound", r)
	}
	if sandboxSupported() {
		licensedSource := append(slices.Clone(source), attribution...)
		outcome := isolatedTests(context.Background(), licensedSource, humanizeOrdinalTask, MaxTimeout, "")
		if outcome.passed || outcome.unknown || !outcome.isolated {
			t.Fatal("actual independent contract did not reject missing Ordinal64", outcome)
		}
	}
	d.Files[0].SHA256, d.Packages[0], d.RequiredTests[0].Test = "changed", "changed", "changed"
	fresh, _ := TaskDefinition(humanizeOrdinalTask)
	if fresh.Files[0].SHA256 == "changed" || fresh.Packages[0] == "changed" || fresh.RequiredTests[0].Test == "changed" {
		t.Fatal("external registry shares caller-mutable slices")
	}
}

func TestHumanizeOrdinalBehavior(t *testing.T) {
	source, attribution := humanizeOrdinalFixture(t)
	for _, mode := range []string{"correct", "comment-only", "no-teen-exception", "teen-exception-only-small", "negative-always-th", "sign-dropped", "absolute-min-overflow", "int32-truncation", "float-rounded-digits", "unsigned-digits", "swapped-st-nd", "legacy-negative-changed", "wrong-signature", "shared-last-result"} {
		t.Run(mode, func(t *testing.T) {
			dir := humanizeOrdinalCandidate(t, source, mode)
			// Every candidate's deliberately broken assertion is ignored; immutable
			// original tests and the independent contract determine acceptance.
			if err := os.WriteFile(filepath.Join(dir, "ordinals_test.go"), []byte("package humanize\ninvalid candidate assertions\n"), 0600); err != nil {
				t.Fatal(err)
			}
			r := verifyHumanizeOrdinal(t, source, attribution, dir)
			if !sandboxSupported() {
				if r.Accepted || r.Status != "verifier_unknown" {
					t.Fatal("unsupported platform claimed external behavior outcome", r)
				}
				return
			}
			if mode == "correct" {
				if !r.Accepted || r.Status != "accepted" || !r.ExecutionIsolated || r.CandidateTestsUsed || r.IndependentTests < 6 {
					t.Fatal("independent correct external candidate did not pass", r)
				}
			} else if r.Accepted || r.Status != "rejected" || !r.ExecutionIsolated || r.CandidateTestsUsed {
				t.Fatal("seeded ordinal error was accepted or untested", mode, r)
			}
		})
	}
}

func TestHumanizeOrdinalBaseAndUnsupportedShape(t *testing.T) {
	source, attribution := humanizeOrdinalFixture(t)
	dir := humanizeOrdinalCandidate(t, source, "correct")
	for _, mode := range []string{"wrong-revision", "wrong-source", "wrong-license", "invented-notice"} {
		req := Request{TaskID: humanizeOrdinalTask, BaseRevision: humanizeOrdinalRevision, BaseFiles: slices.Clone(source), AttributionFiles: slices.Clone(attribution), CandidateDir: dir}
		switch mode {
		case "wrong-revision":
			req.BaseRevision = eventKeyBoundsRevision
		case "wrong-source":
			req.BaseFiles[0].Data = []byte("wrong public module bytes")
		case "wrong-license":
			req.AttributionFiles[0].SHA256 = strings.Repeat("0", 64)
		case "invented-notice":
			req.AttributionFiles = append(req.AttributionFiles, BaseFile{Path: "NOTICE", SHA256: digest(nil)})
		}
		if r, err := Verify(context.Background(), req); err == nil || r.Accepted {
			t.Fatal("incorrect external source/license manifest trusted", mode)
		}
	}
	for _, mode := range []string{"import", "init", "directive", "package"} {
		dir := humanizeOrdinalCandidate(t, source, "correct")
		switch mode {
		case "import":
			replaceCandidate(t, dir, "ordinals.go", `import "strconv"`, "import (\"strconv\";\"fmt\")")
		case "init":
			replaceCandidate(t, dir, "ordinals.go", "func Ordinal(x int)", "func init() {}\nfunc Ordinal(x int)")
		case "directive":
			replaceCandidate(t, dir, "ordinals.go", "func Ordinal(x int)", "//go:generate unused\nfunc Ordinal(x int)")
		case "package":
			replaceCandidate(t, dir, "ordinals.go", "package humanize", "package unrelated")
		}
		r := verifyHumanizeOrdinal(t, source, attribution, dir)
		if r.Accepted || r.Status != "verifier_unknown" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "unsupported_candidate_shape" {
			t.Fatal("unsupported external source shape mislabeled", mode, r)
		}
	}
}
