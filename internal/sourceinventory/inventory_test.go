package sourceinventory

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// All source below is independently authored synthetic text. These tests never
// open retained files or import/call any project source-pin or evaluator API.
const legacyText = `package behaviorprobe
type sourceDef struct { id, prototype, core, function string; run func([]int) []int; correct bool }
var sourceCatalog = []sourceDef{
 {"toy-legacy", "toy-legacy-prototype", "toy-legacy-core", "legacyRoot", legacyRoot, true},
}
//line private.go:900
func legacyRoot(v []int) []int { return []int{v[0] + 1} }
`

const specText = `package typedbehavior
type sourceSpec struct { id, prototype, core, function string; correct bool }
type atomicFn func(int) int
type identityFn func(int) int
type snapshotFn func(int) int
`

const stateText = `package typedbehavior
type marker uint8
const (
 first marker = iota
 second
)
type cause struct { Code uint8 }
func (c *cause) Error() string { return "toy" }
var Sentinel = &cause{Code:1}
func helper(x int) int { return x + 1 }
func atomicRoot(x int) int { return helper(x) + int(first) }
func identityRoot(x int) int { return int(Sentinel.Code) }
var atomicSources = []struct { spec sourceSpec; run atomicFn }{
 {sourceSpec{"toy-atomic", "toy-atomic-prototype", "toy-atomic-core", "atomicRoot", true}, atomicRoot},
}
var identitySources = []struct { spec sourceSpec; run identityFn }{
 {sourceSpec{"toy-identity", "toy-identity-prototype", "toy-identity-core", "identityRoot", false}, identityRoot},
}
var snapshotSources = []struct { spec sourceSpec; run snapshotFn }{}
`

const flowText = `package typedbehavior
func quotedRoot(x int) int { return helper(x) }
func sourcesFlow() []sourceSpec {
 return []sourceSpec{
 {"toy-quoted", "toy-quoted-prototype", "toy-quoted-core", "quotedRoot", true},
 }
}
func checkFlow(id string) int {
 switch id {
 case "toy-quoted": return checkQuoted(quotedRoot, nil)
 default: return 0
 }
}
func checkQuoted(fn func(int) int, ignored any) int { return 0 }
`

func testSHA(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// This tiny independent table pins expected formatting, including bare
// TypeSpec, whole inherited-iota GenDecl, and pointer-receiver method syntax.
func formattedToy(symbol string) string {
	switch symbol {
	case "legacyRoot":
		return "func legacyRoot(v []int) []int { return []int{v[0] + 1} }"
	case "atomicRoot":
		return "func atomicRoot(x int) int { return helper(x) + int(first) }"
	case "identityRoot":
		return "func identityRoot(x int) int { return int(Sentinel.Code) }"
	case "quotedRoot":
		return "func quotedRoot(x int) int { return helper(x) }"
	case "helper":
		return "func helper(x int) int { return x + 1 }"
	case "marker":
		return "marker uint8"
	case "first", "second":
		return "const (\n\tfirst marker = iota\n\tsecond\n)"
	case "cause":
		return "cause struct{ Code uint8 }"
	case "*cause.Error":
		return "func (c *cause) Error() string { return \"toy\" }"
	case "Sentinel":
		return "var Sentinel = &cause{Code: 1}"
	default:
		panic("unknown synthetic declaration")
	}
}

// The expected normalized stream is written from the toy grammar, not obtained
// by parsing, formatting or normalizing candidate output. Each token spelling
// is explicit, and all identifier names outside the historical keep set become _.
func expectedNormal(symbol string) string {
	type entry struct {
		t token.Token
		s string
	}
	var tokens []entry
	funcStart := []entry{{token.FUNC, ""}, {token.IDENT, "_"}, {token.LPAREN, ""}, {token.IDENT, "_"}, {token.IDENT, "int"}, {token.RPAREN, ""}, {token.IDENT, "int"}, {token.LBRACE, ""}, {token.RETURN, ""}}
	switch symbol {
	case "helper":
		tokens = append(append([]entry{}, funcStart...), entry{token.IDENT, "_"}, entry{token.ADD, ""}, entry{token.INT, "1"}, entry{token.RBRACE, ""}, entry{token.SEMICOLON, "\n"})
	case "atomicRoot":
		tokens = append(append([]entry{}, funcStart...), entry{token.IDENT, "_"}, entry{token.LPAREN, ""}, entry{token.IDENT, "_"}, entry{token.RPAREN, ""}, entry{token.ADD, ""}, entry{token.IDENT, "int"}, entry{token.LPAREN, ""}, entry{token.IDENT, "_"}, entry{token.RPAREN, ""}, entry{token.RBRACE, ""}, entry{token.SEMICOLON, "\n"})
	case "identityRoot":
		tokens = append(append([]entry{}, funcStart...), entry{token.IDENT, "int"}, entry{token.LPAREN, ""}, entry{token.IDENT, "_"}, entry{token.PERIOD, ""}, entry{token.IDENT, "_"}, entry{token.RPAREN, ""}, entry{token.RBRACE, ""}, entry{token.SEMICOLON, "\n"})
	case "quotedRoot":
		tokens = append(append([]entry{}, funcStart...), entry{token.IDENT, "_"}, entry{token.LPAREN, ""}, entry{token.IDENT, "_"}, entry{token.RPAREN, ""}, entry{token.RBRACE, ""}, entry{token.SEMICOLON, "\n"})
	case "legacyRoot":
		tokens = []entry{{token.FUNC, ""}, {token.IDENT, "_"}, {token.LPAREN, ""}, {token.IDENT, "_"}, {token.LBRACK, ""}, {token.RBRACK, ""}, {token.IDENT, "int"}, {token.RPAREN, ""}, {token.LBRACK, ""}, {token.RBRACK, ""}, {token.IDENT, "int"}, {token.LBRACE, ""}, {token.RETURN, ""}, {token.LBRACK, ""}, {token.RBRACK, ""}, {token.IDENT, "int"}, {token.LBRACE, ""}, {token.IDENT, "_"}, {token.LBRACK, ""}, {token.INT, "0"}, {token.RBRACK, ""}, {token.ADD, ""}, {token.INT, "1"}, {token.RBRACE, ""}, {token.RBRACE, ""}, {token.SEMICOLON, "\n"}}
	case "*cause.Error":
		tokens = []entry{{token.FUNC, ""}, {token.LPAREN, ""}, {token.IDENT, "_"}, {token.MUL, ""}, {token.IDENT, "_"}, {token.RPAREN, ""}, {token.IDENT, "_"}, {token.LPAREN, ""}, {token.RPAREN, ""}, {token.IDENT, "_"}, {token.LBRACE, ""}, {token.RETURN, ""}, {token.STRING, "\"toy\""}, {token.RBRACE, ""}, {token.SEMICOLON, "\n"}}
	default:
		panic("unknown synthetic token table")
	}
	var text strings.Builder
	for _, e := range tokens {
		if e.t == token.FUNC {
			e.s = "func"
		}
		if e.t == token.RETURN {
			e.s = "return"
		}
		text.WriteString(strconv.Itoa(int(e.t)))
		text.WriteByte(':')
		text.WriteString(e.s)
		text.WriteByte(';')
	}
	return testSHA(text.String())
}

func toyComponent(path, kind, symbol string) Component {
	identity := strings.TrimPrefix(path, "internal/") + ":" + kind + ":" + symbol
	c := Component{ID: identity, Kind: kind, SHA256: testSHA(formattedToy(symbol))}
	if kind == "function" || kind == "method" {
		c.NormalizedBehaviorSHA256 = expectedNormal(symbol)
	}
	return c
}

func fixture() Input {
	paths := fixedPaths()
	texts := [4]string{legacyText, specText, stateText, flowText}
	in := Input{GoVersion: "go1.27.1"}
	for i := range in.Files {
		in.Files[i] = File{paths[i], []byte(texts[i]), testSHA(texts[i])}
	}
	legacy := HistoricalRoot{Cohort: "legacy", ID: "toy-legacy", Prototype: "toy-legacy-prototype", Core: "toy-legacy-core", CodeSHA256: testSHA(formattedToy("legacyRoot")), NormalizedCodeSHA256: expectedNormal("legacyRoot")}
	atomic := HistoricalRoot{Cohort: "typed", ID: "toy-atomic", Prototype: "toy-atomic-prototype", Core: "toy-atomic-core", Function: "atomicRoot", CodeSHA256: testSHA(formattedToy("atomicRoot")), NormalizedCodeSHA256: expectedNormal("atomicRoot"), Components: []Component{
		toyComponent(paths[2], "function", "atomicRoot"), toyComponent(paths[2], "function", "helper"), toyComponent(paths[2], "type", "marker"), toyComponent(paths[2], "value", "first"), toyComponent(paths[2], "value", "second"),
	}}
	identity := HistoricalRoot{Cohort: "typed", ID: "toy-identity", Prototype: "toy-identity-prototype", Core: "toy-identity-core", Function: "identityRoot", CodeSHA256: testSHA(formattedToy("identityRoot")), NormalizedCodeSHA256: expectedNormal("identityRoot"), Components: []Component{
		toyComponent(paths[2], "function", "identityRoot"), toyComponent(paths[2], "method", "*cause.Error"), toyComponent(paths[2], "type", "cause"), toyComponent(paths[2], "value", "Sentinel"),
	}}
	quoted := HistoricalRoot{Cohort: "typed", ID: "toy-quoted", Prototype: "toy-quoted-prototype", Core: "toy-quoted-core", Function: "quotedRoot", CodeSHA256: testSHA(formattedToy("quotedRoot")), NormalizedCodeSHA256: expectedNormal("quotedRoot"), Components: []Component{
		toyComponent(paths[3], "function", "quotedRoot"), toyComponent(paths[2], "function", "helper"),
	}}
	// Independent bundle strings explicitly order the fixed toy components.
	atomic.BundleSHA256 = testSHA("typedbehavior/state.go:function:atomicRoot\x00" + atomic.Components[0].SHA256 + "\x00typedbehavior/state.go:function:helper\x00" + atomic.Components[1].SHA256 + "\x00typedbehavior/state.go:type:marker\x00" + atomic.Components[2].SHA256 + "\x00typedbehavior/state.go:value:first\x00" + atomic.Components[3].SHA256 + "\x00typedbehavior/state.go:value:second\x00" + atomic.Components[4].SHA256 + "\x00go1.27.1;standard-library-infrastructure-v1")
	identity.BundleSHA256 = testSHA("typedbehavior/state.go:function:identityRoot\x00" + identity.Components[0].SHA256 + "\x00typedbehavior/state.go:method:*cause.Error\x00" + identity.Components[1].SHA256 + "\x00typedbehavior/state.go:type:cause\x00" + identity.Components[2].SHA256 + "\x00typedbehavior/state.go:value:Sentinel\x00" + identity.Components[3].SHA256 + "\x00go1.27.1;standard-library-infrastructure-v1")
	quoted.BundleSHA256 = testSHA("typedbehavior/flow.go:function:quotedRoot\x00" + quoted.Components[0].SHA256 + "\x00typedbehavior/state.go:function:helper\x00" + quoted.Components[1].SHA256 + "\x00go1.27.1;standard-library-infrastructure-v1")
	in.Roots = []HistoricalRoot{legacy, atomic, identity, quoted}
	return in
}

func TestSyntheticRebindingPhysicalSpansAndSharedGenDecl(t *testing.T) {
	in := fixture()
	r, err := Rebind(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.State != "metadata_rebound_content_review_pending" || r.ContentReview != "pending" || r.ClosureDiscoveryPerformed || r.ObjectBindingPerformed {
		t.Fatalf("scope: %+v", r)
	}
	if r.Counters != (Counters{FileHashAttempts: 4, FileHashesCompleted: 4, FilePinsMatched: 4, ParseAttempts: 4, ParsesCompleted: 4, ParsesSucceeded: 4, RegistryEntries: 4, RootAttempts: 4, RootsCompleted: 4, ComponentAttempts: 11, ComponentsCompleted: 11, RawSpanHashAttempts: 15, RawSpanHashesCompleted: 15, FormatCallsAttempted: 15, FormatCallsCompleted: 15, ScanCallsAttempted: 10, ScanCallsCompleted: 10, BundleCallsAttempted: 3, BundleCallsCompleted: 3}) || r.ReusePolicy != "no_cache_each_relation_v1" {
		t.Fatalf("counts: %+v", r.Counters)
	}
	if len(r.Roots) != 4 {
		t.Fatal("root count")
	}
	for _, root := range r.Roots {
		if !root.MetadataPinsMatched || root.ContentReview != "pending" {
			t.Fatal("review state")
		}
		refs := append([]DeclarationReference{root.Root}, root.Components...)
		for _, ref := range refs {
			file := slices.IndexFunc(in.Files[:], func(f File) bool { return f.Path == ref.Path })
			if file < 0 || ref.RawSHA256 != testSHA(string(in.Files[file].Raw[ref.StartByte:ref.EndByte])) {
				t.Fatal("span SHA")
			}
		}
	}
	legacy := r.Roots[0].Root
	if legacy.StartLine != 7 || legacy.EndLine != 7 || legacy.RawSHA256 != testSHA("func legacyRoot(v []int) []int { return []int{v[0] + 1} }") {
		t.Fatalf("physical //line position: %+v", legacy)
	}
	var first, second DeclarationReference
	for _, c := range r.Roots[1].Components {
		if c.Symbol == "first" {
			first = c
		}
		if c.Symbol == "second" {
			second = c
		}
	}
	if first.ID == second.ID || first.StartByte != second.StartByte || first.EndByte != second.EndByte || first.RawSHA256 != second.RawSHA256 || first.FormattedSHA256 != second.FormattedSHA256 {
		t.Fatal("multi-symbol GenDecl identity lost")
	}
	if first.NormalizedSHA256 != "" || second.NormalizedSHA256 != "" {
		t.Fatal("enum becomes normalized behavioral proof")
	}
}

func repin(in *Input, file int, text string) {
	in.Files[file].Raw = []byte(text)
	in.Files[file].ExpectedSHA256 = testSHA(text)
}

func TestSyntheticFailClosedAndPartialCounters(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Input)
		want   error
	}{
		{"altered_file", func(in *Input) { in.Files[2].Raw = append(in.Files[2].Raw, ' ') }, ErrFilePin},
		{"unapproved_path", func(in *Input) { in.Files[2].Path = "../state.go" }, ErrInput},
		{"bad_version", func(in *Input) { in.GoVersion = "go1.27.2" }, ErrInput},
		{"invalid_parse", func(in *Input) { repin(in, 2, "package typedbehavior\nfunc (") }, ErrParse},
		{"registry_callback", func(in *Input) { repin(in, 0, strings.Replace(legacyText, "legacyRoot, true", "anotherRoot, true", 1)) }, ErrRegistry},
		{"registry_local_fake", func(in *Input) {
			repin(in, 0, strings.Replace(legacyText, "var sourceCatalog = []sourceDef{", "var anotherCatalog = []sourceDef{", 1)+"func local() { var sourceCatalog = []sourceDef{}; _ = sourceCatalog }\n")
		}, ErrRegistry},
		{"flow_wrong_dispatch", func(in *Input) {
			repin(in, 3, strings.Replace(flowText, "checkQuoted(quotedRoot, nil)", "checkQuoted(helper, nil)", 1))
		}, ErrRegistry},
		{"flow_local_shadow", func(in *Input) {
			repin(in, 3, strings.Replace(flowText, " switch id {", " quotedRoot := helper\n switch id {", 1))
		}, ErrRegistry},
		{"flow_parameter_shadow", func(in *Input) {
			repin(in, 3, strings.Replace(flowText, "checkFlow(id string)", "checkFlow(id string, quotedRoot func(int)int)", 1))
		}, ErrRegistry},
		{"flow_result_shadow", func(in *Input) {
			repin(in, 3, strings.Replace(flowText, "checkFlow(id string) int", "checkFlow(id string) (quotedRoot int)", 1))
		}, ErrRegistry},
		{"flow_arbitrary_defer", func(in *Input) {
			repin(in, 3, strings.Replace(flowText, " switch id {", " defer func() { quotedRoot(0) }()\n switch id {", 1))
		}, ErrRegistry},
		{"wrong_state_run_type", func(in *Input) {
			repin(in, 2, strings.Replace(stateText, "run atomicFn", "run identityFn", 1))
		}, ErrRegistry},
		{"duplicate_root", func(in *Input) { in.Roots[3] = in.Roots[1] }, ErrInput},
		{"root_hash", func(in *Input) { in.Roots[1].CodeSHA256 = strings.Repeat("0", 64) }, ErrHistoricalPin},
		{"normalized_hash", func(in *Input) { in.Roots[1].NormalizedCodeSHA256 = strings.Repeat("0", 64) }, ErrHistoricalPin},
		{"unexpected_standard_import", func(in *Input) { in.Roots[1].StandardImports = []string{"errors"} }, ErrHistoricalPin},
		{"component_hash", func(in *Input) { in.Roots[1].Components[1].SHA256 = strings.Repeat("0", 64) }, ErrHistoricalPin},
		{"component_receiver", func(in *Input) { in.Roots[2].Components[1].ID = "typedbehavior/state.go:method:cause.Error" }, ErrDeclaration},
		{"component_missing", func(in *Input) { in.Roots[1].Components[1].ID = "typedbehavior/state.go:function:missing" }, ErrDeclaration},
		{"duplicate_component", func(in *Input) { in.Roots[1].Components[1] = in.Roots[1].Components[0] }, ErrInput},
		{"unordered_components", func(in *Input) {
			in.Roots[1].Components[1], in.Roots[1].Components[0] = in.Roots[1].Components[0], in.Roots[1].Components[1]
		}, ErrInput},
		{"conflicting_component_across_roots", func(in *Input) { in.Roots[3].Components[1].SHA256 = strings.Repeat("0", 64) }, ErrInput},
		{"bundle_hash", func(in *Input) { in.Roots[1].BundleSHA256 = strings.Repeat("0", 64) }, ErrHistoricalPin},
		{"nonallowlisted_import", func(in *Input) { repin(in, 1, specText+"\nimport \"example.org/private\"\n") }, ErrParse},
		{"external_import", func(in *Input) {
			repin(in, 1, strings.Replace(specText, "package typedbehavior", "package typedbehavior\nimport \"example.org/private\"", 1))
		}, ErrRecipe},
		{"dot_import", func(in *Input) {
			repin(in, 1, strings.Replace(specText, "package typedbehavior", "package typedbehavior\nimport . \"errors\"", 1))
		}, ErrShape},
		{"init_function", func(in *Input) { repin(in, 2, stateText+"func init() {}\n") }, ErrShape},
		{"generic_receiver", func(in *Input) { repin(in, 2, strings.Replace(stateText, "(c *cause)", "(c *cause[int])", 1)) }, ErrShape},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := fixture()
			tc.change(&in)
			r, e := Rebind(in)
			if !errors.Is(e, tc.want) {
				t.Fatalf("got %v want %v", e, tc.want)
			}
			if r.State != "incomplete" || r.ContentReview != "pending" {
				t.Fatal("failure loses state")
			}
			if r.Counters.RootsCompleted > r.Counters.RootAttempts || r.Counters.ComponentsCompleted > r.Counters.ComponentAttempts {
				t.Fatal("counter conservation")
			}
		})
	}
	in := fixture()
	in.Roots[2].BundleSHA256 = strings.Repeat("0", 64)
	r, e := Rebind(in)
	if e != ErrHistoricalPin || r.Counters.RootAttempts != 3 || r.Counters.RootsCompleted != 2 || len(r.Roots) != 2 || r.Counters.ComponentAttempts != 9 || r.Counters.ComponentsCompleted != 9 {
		t.Fatalf("partial work: %+v %v", r.Counters, e)
	}
}

func TestSyntheticWhitespaceIsNotRawIdentity(t *testing.T) {
	in := fixture()
	repin(&in, 2, strings.Replace(stateText, "func helper(x int) int { return x + 1 }", "func helper(x int) int {\n\treturn x + 1\n}", 1))
	// Existing single-line formatted hash changes if source layout changes, even
	// though normalized token hashes agree. No silent recipe correction is allowed.
	r, e := Rebind(in)
	if e != ErrHistoricalPin || r.Counters.RootsCompleted != 1 {
		t.Fatalf("got %v %+v", e, r.Counters)
	}
}

func TestSyntheticIdentifierNormalizationAndLiteralDifference(t *testing.T) {
	a, e := normalized([]byte("func a(x int) int { return x + 1 }"))
	if e != nil || a != expectedNormal("helper") {
		t.Fatal("literal token table mismatch")
	}
	b, e := normalized([]byte("func other(y int) int { return y + 1 }"))
	if e != nil || a != b {
		t.Fatal("rename normalization mismatch")
	}
	c, e := normalized([]byte("func a(x int) int { return x - 1 }"))
	if e != nil || a == c {
		t.Fatal("operator erased")
	}
	d, e := normalized([]byte("func a(x int) int { return x + 2 }"))
	if e != nil || a == d {
		t.Fatal("literal erased")
	}
}

func TestSyntheticMultinameValueAndGroupedTypeNodes(t *testing.T) {
	raw := []byte("package typedbehavior\ntype (\n A int\n B bool\n)\nvar Left, Right int\n")
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "state.go", raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	decls, err := indexDeclarations(f, fixedPaths()[2], 2)
	if err != nil {
		t.Fatal(err)
	}
	var files [4]File
	files[2] = File{Path: fixedPaths()[2], Raw: raw}
	var refs []DeclarationReference
	for _, symbol := range []string{"A", "B", "Left", "Right"} {
		kind := "type"
		if symbol == "Left" || symbol == "Right" {
			kind = "value"
		}
		d, err := lookup(decls, 2, kind, symbol)
		if err != nil {
			t.Fatal(err)
		}
		var calls Counters
		ref, err := reference(fs, files, d, symbol, false, &calls)
		if err != nil {
			t.Fatal(err)
		}
		refs = append(refs, ref)
	}
	if refs[0].FormattedSHA256 != testSHA("A int") || refs[1].FormattedSHA256 != testSHA("B bool") || refs[0].StartByte == refs[1].StartByte {
		t.Fatal("TypeSpec accidentally widened to GenDecl")
	}
	if refs[2].ID == refs[3].ID || refs[2].StartByte != refs[3].StartByte || refs[2].EndByte != refs[3].EndByte || refs[2].FormattedSHA256 != testSHA("var Left, Right int") || refs[3].FormattedSHA256 != refs[2].FormattedSHA256 {
		t.Fatal("multiname ValueSpec lost whole GenDecl identity")
	}
}

func TestSyntheticPhysicalReturnedCountersSurviveFailure(t *testing.T) {
	in := fixture()
	in.Files[2].Raw = append(in.Files[2].Raw, ' ')
	r, err := Rebind(in)
	if err != ErrFilePin || r.Counters.FileHashAttempts != 3 || r.Counters.FileHashesCompleted != 3 || r.Counters.FilePinsMatched != 2 || r.Counters.ParseAttempts != 2 || r.Counters.ParsesCompleted != 2 || r.Counters.ParsesSucceeded != 2 {
		t.Fatalf("file physical counters: %+v %v", r.Counters, err)
	}
	in = fixture()
	repin(&in, 2, "package typedbehavior\nfunc (")
	r, err = Rebind(in)
	if err != ErrParse || r.Counters.ParseAttempts != 3 || r.Counters.ParsesCompleted != 3 || r.Counters.ParsesSucceeded != 2 {
		t.Fatalf("parser returned vs success: %+v %v", r.Counters, err)
	}
	in = fixture()
	in.Roots[2].BundleSHA256 = strings.Repeat("0", 64)
	r, err = Rebind(in)
	if err != ErrHistoricalPin || r.Counters.BundleCallsAttempted != 2 || r.Counters.BundleCallsCompleted != 2 || r.Counters.RootAttempts != 3 || r.Counters.RootsCompleted != 2 || r.Counters.FormatCallsAttempted != 12 || r.Counters.FormatCallsCompleted != 12 || r.Counters.ScanCallsAttempted != 7 || r.Counters.ScanCallsCompleted != 7 {
		t.Fatalf("bundle mismatch loses work: %+v %v", r.Counters, err)
	}
}
