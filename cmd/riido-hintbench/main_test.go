package main

import (
	"reflect"
	"testing"
)

func TestRankingPreservesFallbackAndDoesNotReadTarget(t *testing.T) {
	x := newIndex()
	for _, m := range []string{"catalog_order", "fixed_shuffle", "tokens", "hash256", "hash4096"} {
		for _, q := range queries() {
			a := x.rank(q.text, m)
			b := x.rank(q.text, m)
			if !reflect.DeepEqual(a, b) {
				t.Fatal("repeat changed ranking")
			}
			seen := make([]bool, len(fixtures))
			for _, i := range a {
				if i < 0 || i >= len(seen) || seen[i] {
					t.Fatal("invalid permutation")
				}
				seen[i] = true
			}
			if len(a) != len(fixtures) {
				t.Fatal("fallback lost candidate")
			}
		}
	}
}
func TestOracleCountsAbsentAndContrastErrors(t *testing.T) {
	x := newIndex()
	qs := queries()
	r := measure(x, qs, "tokens")
	errors := 0
	for _, c := range r.Cases {
		if c.Kind == "absent" && (c.Rank != 0 || c.Checks != len(fixtures)) {
			t.Fatal("absent case must exhaust catalog")
		}
		if c.Kind == "contrast" && c.Rank != 1 {
			errors++
		}
	}
	if errors == 0 {
		t.Fatal("fixture should expose inability to interpret negation")
	}
}
func TestFeatureCollisionAndDeduplication(t *testing.T) {
	if overlap(tokens("one one two"), tokens("one")) != 1 {
		t.Fatal("duplicate terms inflated score")
	}
	if intersect(encode([]string{"one"}, 256), encode([]string{"one"}, 256), 256) != 1 {
		t.Fatal("equal terms failed")
	}
}
