// SPDX-License-Identifier: Apache-2.0
package hintlearn

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// Owned synthetic controls only. Features is the unchanged oracle; no model,
// corpus, source API, Fit, benchmark or performance claim is involved.
func scratchBitsEqual(a, b []Feature) bool {
	if len(a) != len(b) || (a == nil) != (b == nil) {
		return false
	}
	for i := range a {
		if a[i].Index != b[i].Index || math.Float64bits(a[i].Value) != math.Float64bits(b[i].Value) {
			return false
		}
	}
	return true
}

func TestFeatureScratchOracleBitsAndCardinality(t *testing.T) {
	var named [64]string
	for i := range named {
		named[i] = "word" + strconv.Itoa(i)
	}
	distinct := strings.Join(named[:], " ")
	fixtures := []struct {
		q, d  string
		slots int
	}{
		{"", "alpha", 0},
		{"alpha", "!!!🙂", 0},
		{"alpha beta", "beta alpha", 10},
		{"Å β 한글 １２", "å Β 한글 １２", 50},
		{"left\x00right", "right left", 10},
		{strings.Repeat("repeat ", 32), strings.Repeat("repeat ", 32), 3970},
		{strings.Repeat("repeat ", 64), distinct, 16130},
		{distinct, distinct, 16130},
	}
	for _, f := range fixtures {
		want := Features(f.q, f.d)
		count, slots, err := FeatureCardinality(f.q, f.d)
		if err != nil || count != len(want) || slots != f.slots {
			t.Fatalf("cardinality: count=%d slots=%d error=%v", count, slots, err)
		}
		s, err := NewFeatureScratch(slots)
		if err != nil || len(s.cross) != slots || cap(s.cross) != slots {
			t.Fatal("reserved buffer", err)
		}
		for repeat := 0; repeat < 3; repeat++ {
			got, err := s.Features(f.q, f.d)
			if err != nil || !scratchBitsEqual(got, want) {
				t.Fatal("feature index/value bits/order changed", repeat, err)
			}
			if len(got) > 0 && got[len(got)-1].Index != 0 {
				t.Fatal("overlap is not last")
			}
		}
	}
}

// Find a bounded synthetic opposite-sign collision. The third (bigram) term
// must use another bucket, so cancellation cannot be confused with omission.
func scratchCancelledPair(t *testing.T) (string, int) {
	t.Helper()
	var seen [Dimension][2]string
	prefix := hashFrom(hash("q"), "\x00")
	for i := 0; i < 1<<17; i++ {
		term := "w" + strconv.Itoa(i)
		h := hashFrom(prefix, term)
		index, sign := 1+int(h%(Dimension-1)), int(h>>63)
		other := seen[index][1-sign]
		if other != "" && 1+int(hashFrom(prefix, other+"_"+term)%(Dimension-1)) != index {
			return other + " " + term, index
		}
		seen[index][sign] = term
	}
	t.Fatal("bounded synthetic collision witness unavailable")
	return "", 0
}

func TestFeatureScratchRetainsCancelledZero(t *testing.T) {
	doc, index := scratchCancelledPair(t)
	want := Features("q", doc)
	count, slots, err := FeatureCardinality("q", doc)
	if err != nil || count != 3 || count != len(want) || slots != 4 {
		t.Fatal("cancelled feature cardinality", count, slots, err)
	}
	s, err := NewFeatureScratch(slots)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Features("q", doc)
	if err != nil || !scratchBitsEqual(got, want) {
		t.Fatal("collision bits/order", err)
	}
	found := false
	for _, f := range got {
		if f.Index == index {
			found = true
			if math.Float64bits(f.Value) != 0 {
				t.Fatal("cancelled cross feature changed")
			}
		}
	}
	if !found {
		t.Fatal("zero cross feature pruned")
	}
}

func TestFeatureScratchReuseAndConstructorIndependence(t *testing.T) {
	a, err := NewFeatureScratch(10)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewFeatureScratch(10)
	if err != nil {
		t.Fatal(err)
	}
	long, err := a.Features("alpha beta", "beta alpha")
	if err != nil {
		t.Fatal(err)
	}
	kept := append([]Feature(nil), long...)
	other, err := b.Features("gamma delta", "gamma delta")
	if err != nil || !scratchBitsEqual(long, kept) || &long[0] == &other[0] {
		t.Fatal("constructors share mutable storage", err)
	}
	for _, pair := range [][2]string{{"alpha", "gamma"}, {"", "alpha"}, {"alpha beta", "beta alpha"}} {
		got, err := a.Features(pair[0], pair[1])
		if err != nil || !scratchBitsEqual(got, Features(pair[0], pair[1])) {
			t.Fatal("stale borrowed tail became an active feature", err)
		}
	}
	if !scratchBitsEqual(other, Features("gamma delta", "gamma delta")) {
		t.Fatal("using another constructor changed this row")
	}
}

func TestFeatureScratchCapacityAndErrorPrecedence(t *testing.T) {
	for _, slots := range []int{-1, int(^uint(0) >> 1)} {
		if s, err := NewFeatureScratch(slots); s != nil || err != ErrFeatureScratchSize {
			t.Fatal("invalid allocation size admitted", err)
		}
	}
	var nilScratch *FeatureScratch
	if fs, err := nilScratch.Features("", ""); fs != nil || err != ErrFeatureScratchSize {
		t.Fatal("nil receiver must fail before empty terms", err)
	}
	z, err := NewFeatureScratch(0)
	if err != nil {
		t.Fatal(err)
	}
	if fs, err := z.Features("alpha", ""); fs != nil || err != nil {
		t.Fatal("zero-capacity empty row", err)
	}
	if fs, err := z.Features("alpha", "beta"); fs != nil || err != ErrFeatureScratchCapacity {
		t.Fatal("zero-capacity nonempty row", err)
	}
	s, err := NewFeatureScratch(2)
	if err != nil {
		t.Fatal(err)
	}
	borrowed, err := s.Features("alpha", "beta")
	if err != nil {
		t.Fatal(err)
	}
	before := append([]Feature(nil), borrowed...)
	if fs, err := s.Features("alpha beta", "gamma"); fs != nil || err != ErrFeatureScratchCapacity || !scratchBitsEqual(borrowed, before) {
		t.Fatal("capacity error altered prior borrowed row", err)
	}
	if fs, err := s.Features("alpha", "beta"); err != nil || !scratchBitsEqual(fs, before) {
		t.Fatal("capacity failure poisoned later use", err)
	}
}
