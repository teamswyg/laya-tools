package hintsearch

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func pathFixture(n int) []string {
	r := make([]string, n)
	for i := range r {
		r[i] = fmt.Sprintf("src/file%05d.go", i)
	}
	r[n-1] = "src/needle.go"
	return r
}
func TestPathIndexRetainsLateCandidateAndGlobalScores(t *testing.T) {
	paths := pathFixture(6000)
	before := slices.Clone(paths)
	idx, e := NewPathIndex(paths)
	if e != nil {
		t.Fatal(e)
	}
	r, e := idx.RankLongInto("needle", Ranking{})
	if e != nil {
		t.Fatal(e)
	}
	want := math.Log(1 + (6000-1+.5)/(1+.5))
	if len(r.Order) != 6000 || r.Order[0] != 5999 || math.Abs(r.Scores[5999]-want) > 1e-12 {
		t.Fatal("candidate lost or non-global score")
	}
	if !slices.Equal(paths, before) {
		t.Fatal("input mutated")
	}
	if _, e := New(paths); e == nil {
		t.Fatal("old constructor limit changed")
	}
	small := paths[:10]
	old, _ := New(small)
	next, _ := NewPathIndex(small)
	for _, q := range []string{"src go", "file00005", "unknown"} {
		a, _ := old.Rank(q)
		b, _ := next.Rank(q)
		if !reflect.DeepEqual(a, b) {
			t.Fatal("small catalog parity")
		}
	}
}
func TestPathIndexInputBounds(t *testing.T) {
	for _, paths := range [][]string{nil, make([]string, MaxPathDocuments+1), {"same.go", "same.go"}, {"../bad"}, {"/absolute"}, {"a/../b"}, {"."}, {"a\x00b"}, {string([]byte{0xff})}, {strings.Repeat("x", 4097)}} {
		if _, e := NewPathIndex(paths); e == nil {
			t.Fatal("accepted invalid paths")
		}
	}
	paths := make([]string, 4097)
	for i := range paths {
		paths[i] = fmt.Sprintf("%04d", i) + strings.Repeat("x", 4092)
	}
	if _, e := NewPathIndex(paths); e == nil {
		t.Fatal("accepted excessive catalog bytes")
	}
}
func TestLargePathMergePreservesRankBound(t *testing.T) {
	base := make([]int, 6000)
	for i := range base {
		base[i] = i
	}
	hints := slices.Clone(base)
	slices.Reverse(hints)
	got, e := InterleavePathBaselineFirst(base, hints)
	if e != nil {
		t.Fatal(e)
	}
	seen := make([]bool, len(base))
	for rank, d := range got {
		if d < 0 || d >= len(base) || seen[d] {
			t.Fatal("invalid permutation")
		}
		seen[d] = true
		if rank+1 > min(len(base), 2*(d+1)-1) {
			t.Fatal("baseline rank bound")
		}
	}
	if len(got) != len(base) || got[0] != 0 || base[0] != 0 || hints[0] != 5999 {
		t.Fatal("incomplete or mutated inputs")
	}
	if _, e := InterleaveBaselineFirst(base, hints); e == nil {
		t.Fatal("old merge bound changed")
	}
	if _, e := InterleavePathBaselineFirst(base, []int{1, 1}); e == nil {
		t.Fatal("duplicate hint accepted")
	}
	smallBase := base[:10]
	smallHints := []int{9, 8, 7}
	a, _ := InterleaveBaselineFirst(smallBase, smallHints)
	b, _ := InterleavePathBaselineFirst(smallBase, smallHints)
	if !slices.Equal(a, b) {
		t.Fatal("merge parity")
	}
}
func BenchmarkPathIndex10000(b *testing.B) {
	paths := pathFixture(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, e := NewPathIndex(paths); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkPathRank10000(b *testing.B) {
	idx, e := NewPathIndex(pathFixture(10000))
	if e != nil {
		b.Fatal(e)
	}
	var dst Ranking
	dst, e = idx.RankLongInto("needle", dst)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		dst, e = idx.RankLongInto("needle", dst)
		if e != nil {
			b.Fatal(e)
		}
	}
}

func TestPathIndexAtMaximumCount(t *testing.T) {
	idx, err := NewPathIndex(pathFixture(MaxPathDocuments))
	if err != nil {
		t.Fatal(err)
	}
	result, err := idx.RankLongInto("needle", Ranking{})
	if err != nil || len(result.Order) != MaxPathDocuments || result.Order[0] != MaxPathDocuments-1 {
		t.Fatalf("maximum-count search failed: %v", err)
	}
}
