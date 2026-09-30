package hintsearch

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestCoverageUniqueTermsAndAbsent(t *testing.T) {
	idx, e := New([]string{"Read FILE 파일", "file write", "other"})
	if e != nil {
		t.Fatal(e)
	}
	got, e := idx.QueryCoverage("FILE file 파일 missing", []int{0, 1, 2})
	if e != nil {
		t.Fatal(e)
	}
	if got.QueryTerms != 3 || got.KnownTerms != 2 || got.Documents != 3 || got.Fractions[0] != 2.0/3 || got.Fractions[1] != 1.0/3 || got.Fractions[2] != 0 {
		t.Fatalf("wrong coverage %+v", got)
	}
	zero, e := idx.QueryCoverage("!!!", []int{0})
	if e != nil || zero.QueryTerms != 0 || zero.Fractions[0] != 0 {
		t.Fatal("zero tokens")
	}
	for _, ids := range [][]int{nil, {0, 0}, {-1}, {3}, make([]int, 21)} {
		if _, e := idx.QueryCoverage("file", ids); e == nil {
			t.Fatal("accepted invalid ids")
		}
	}
	if _, e := idx.QueryCoverage(strings.Repeat("a", MaxQueryBytes+1), []int{0}); e == nil {
		t.Fatal("oversize query")
	}
	var absent *Index
	if _, e := absent.QueryCoverage("x", []int{0}); e == nil {
		t.Fatal("nil index")
	}
}
func TestCoverageMatchesDirectScanAndConcurrentRead(t *testing.T) {
	docs := []string{"alpha beta beta", "beta Gamma", "파일 alpha", "", "delta"}
	idx, e := New(docs)
	if e != nil {
		t.Fatal(e)
	}
	ids := []int{4, 2, 0, 3, 1}
	query := "ALPHA alpha 파일 beta missing"
	want := Coverage{QueryTerms: 4, KnownTerms: 3, Documents: 5}
	terms := words(query)
	slices.Sort(terms)
	terms = slices.Compact(terms)
	for j, id := range ids {
		ws := words(docs[id])
		n := 0
		for _, term := range terms {
			if slices.Contains(ws, term) {
				n++
			}
		}
		want.Fractions[j] = float64(n) / float64(len(terms))
	}
	before, e := idx.Rank(query)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 10 {
				got, e := idx.QueryCoverage(query, ids)
				if e != nil || got != want {
					t.Errorf("coverage mismatch: %+v %v", got, e)
					return
				}
			}
		})
	}
	wg.Wait()
	after, e := idx.Rank(query)
	if e != nil || !slices.Equal(before.Order, after.Order) || !slices.Equal(before.Scores, after.Scores) {
		t.Fatal("coverage mutated index")
	}
}

var coverageSink Coverage
var rankSink Ranking

func BenchmarkCoverageAndRanking(b *testing.B) {
	docs := make([]string, 3009)
	for i := range docs {
		docs[i] = fmt.Sprintf("func ReadFile%d path buffer contents error return close", i)
	}
	idx, e := New(docs)
	if e != nil {
		b.Fatal(e)
	}
	query := "read file contents error"
	var ids [20]int
	for i := range ids {
		ids[i] = i
	}
	b.Run("coverage20", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var err error
			coverageSink, err = idx.QueryCoverage(query, ids[:])
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("rank_reused", func(b *testing.B) {
		dst, e := idx.Rank(query)
		if e != nil {
			b.Fatal(e)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst, e = idx.RankInto(query, dst)
			if e != nil {
				b.Fatal(e)
			}
		}
		rankSink = dst
	})
}
