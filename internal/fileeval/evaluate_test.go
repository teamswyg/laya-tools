package fileeval

import (
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestRealCandidateRankingSeparateFromLabels(t *testing.T) {
	paths := []string{"src/FooBar.go", "src/other.go", "test/FooBar_test.go"}
	before := slices.Clone(paths)
	orders, failed, e := Rank("foo bar", paths)
	if e != nil || failed {
		t.Fatalf("%v", e)
	}
	if len(orders[Baseline]) != 3 || orders[Interleaved][0] != orders[Baseline][0] || !slices.Equal(paths, before) {
		t.Fatal("catalog lost/mutated")
	}
	a, e := Measure(paths, orders[Interleaved], []string{"src/FooBar.go"})
	if e != nil || a.Mapped != 1 {
		t.Fatal(e)
	}
	_, e = Measure(paths, orders[Interleaved], []string{"missing.go"})
	if e != nil {
		t.Fatal(e)
	}
	repeat, _, e := Rank("foo bar", paths)
	if e != nil || !reflect.DeepEqual(orders, repeat) {
		t.Fatal("label scoring changed ranking")
	}
}
func TestScoringMissingMultipleAndNewOnly(t *testing.T) {
	paths := []string{"a", "b", "c"}
	order := []int{2, 0, 1}
	s, e := Measure(paths, order, []string{"a", "c", "missing"})
	if e != nil || s.Targets != 3 || s.Mapped != 2 || s.FirstRank != 1 || !s.Hit1 || s.All10 || s.ReciprocalRank != 1 {
		t.Fatalf("%+v %v", s, e)
	}
	s, e = Measure(paths, order, nil)
	if e != nil || s.FirstRank != 0 || s.Hit100 || s.All10 {
		t.Fatal("new-only fabricated success")
	}
	if _, e = Measure(paths, []int{0, 0, 2}, []string{"a"}); e == nil {
		t.Fatal("invalid permutation")
	}
	if _, e = Measure(paths, order, []string{"a", "a"}); e == nil {
		t.Fatal("duplicate label")
	}
}
func TestExpandedQueryFallsBackWithoutDroppingBaseline(t *testing.T) {
	// Camel-case expansion exceeds128KiB although original input is within it.
	q := strings.Repeat("aB", 50000)
	orders, failed, e := Rank(q, []string{"a.go", "b.go"})
	if e != nil || !failed || !reflect.DeepEqual(orders[Baseline], orders[Interleaved]) {
		t.Fatalf("fallback %v %v", failed, e)
	}
}
