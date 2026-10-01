package typedbehavior

import (
	"slices"
	"strings"
	"testing"
)

func TestSourceClosureIncludesAliasesGlobalsMethodsAndIota(t *testing.T) {
	source := `package typedbehavior
type box struct { N int }
const (first = 10 + iota; second)
var base = 3
var alias = plus
func plus(v int) int { return v + base }
func (b box) value() int { return b.N + second }
func root(v int) int { return (box{alias(v)}).value() }
`
	pins, e := pinSources([]sourceFile{{"sample.go", source}}, []sourceSpec{{"sample", "sample", "sample", "root", true}})
	if e != nil || len(pins) != 1 {
		t.Fatalf("closure: %v", e)
	}
	components := pins[0].Components
	for _, symbol := range []string{"root", "box", "base", "alias", "plus", "second", "box.value"} {
		if !slices.ContainsFunc(components, func(c ComponentPin) bool { return strings.HasSuffix(c.ID, ":"+symbol) }) {
			t.Fatalf("uncovered authored dependency %s", symbol)
		}
	}
	changed := strings.Replace(source, "b.N + second", "b.N - second", 1)
	after, e := pinSources([]sourceFile{{"sample.go", changed}}, []sourceSpec{{"sample", "sample", "sample", "root", true}})
	if e != nil {
		t.Fatal(e)
	}
	if pins[0].CodeSHA256 != after[0].CodeSHA256 || pins[0].BundleSHA256 == after[0].BundleSHA256 {
		t.Fatal("receiver method change was not bound independently of unchanged root")
	}
	changed = strings.Replace(source, "first = 10 + iota", "first = 20 + iota", 1)
	after, e = pinSources([]sourceFile{{"sample.go", changed}}, []sourceSpec{{"sample", "sample", "sample", "root", true}})
	if e != nil || pins[0].BundleSHA256 == after[0].BundleSHA256 {
		t.Fatal("inherited constant initializer not pinned", e)
	}
}

func TestSourceClosureDoesNotTreatShadowingLocalsAsGlobals(t *testing.T) {
	source := `package typedbehavior
var helper = 7
func root(v int) int { helper := v; return helper }
`
	pins, e := pinSources([]sourceFile{{"sample.go", source}}, []sourceSpec{{"sample", "sample", "sample", "root", true}})
	if e != nil || len(pins) != 1 {
		t.Fatal(e)
	}
	if slices.ContainsFunc(pins[0].Components, func(c ComponentPin) bool { return strings.HasSuffix(c.ID, ":helper") }) {
		t.Fatal("shadowing local falsely pinned as authored global")
	}
}

func TestSourceClosureRejectsHiddenInitializationAndUnlistedImports(t *testing.T) {
	cases := []string{
		"package typedbehavior;func init(){};func root(v int)int{return v}",
		"package typedbehavior;var seed=bump();func bump()int{return 7};func root(v int)int{return v}",
		"package typedbehavior;var data=[]int{1};var _=copy(data,[]int{2});func root(v int)int{return data[0]}",
		"package typedbehavior;var data=[]int{1};var _=append(data[:0],2);func root(v int)int{return data[0]}",
		"package typedbehavior;import \"net/http\";func root(v int)int{_ = http.MethodGet;return v}",
		"package typedbehavior;func root(v int)int{return unresolved(v)}",
	}
	for _, source := range cases {
		if _, e := pinSources([]sourceFile{{"sample.go", source}}, []sourceSpec{{"sample", "sample", "sample", "root", true}}); e == nil {
			t.Fatal("unsupported closure accepted")
		}
	}
}

func TestActualClosedRegistryPinsAndLiteralControls(t *testing.T) {
	pins, e := SourcePins()
	if e != nil {
		t.Fatal(e)
	}
	if len(pins) != 24 {
		t.Fatal("authored closed registry count", len(pins))
	}
	for i, s := range sourceSpecs() {
		if pins[i].ID != s.id || len(pins[i].CodeSHA256) != 64 || len(pins[i].BundleSHA256) != 64 || len(pins[i].Components) == 0 {
			t.Fatal("source identity not pinned")
		}
		c := checkSource(s.id)
		contract := slices.IndexFunc(Contracts(), func(c ContractSpec) bool { return c.Prototype == s.prototype })
		if contract < 0 || c.Unknown || c.Checked != Contracts()[contract].Vectors || s.correct && c.Failed != 0 || !s.correct && c.Failed == 0 {
			t.Fatal("independent literal control", s.id, c)
		}
	}
}
