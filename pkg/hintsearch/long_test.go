package hintsearch

import (
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestLongQueryTailAndIndependentScores(t *testing.T) {
	idx, err := New([]string{"needle", "hay"})
	if err != nil {
		t.Fatal(err)
	}
	q := strings.Repeat("unknown ", 1100) + "needle"
	got, err := idx.RankLongInto(q, Ranking{})
	if err != nil {
		t.Fatal(err)
	}
	// Each document has one token, so BM25's tf normalization is exactly one.
	// Repeated out-of-vocabulary words must neither affect normalization nor
	// hide the only useful term beyond the old input boundary.
	if got.Order[0] != 0 || math.Abs(got.Scores[0]-math.Log(2)) > 1e-12 || got.Scores[1] != 0 {
		t.Fatalf("unexpected scores: %+v", got)
	}
	if _, err := idx.Rank(q); err == nil {
		t.Fatal("old bound changed")
	}
}

func TestLongQueryParityAndRejectionPreservesBuffers(t *testing.T) {
	idx, _ := New([]string{"한글 Unicode needle", "hay needle", "zero"})
	for _, q := range []string{"needle", "한글 UNICODE needle needle", "unseen", strings.Repeat(" ", 8191) + "x"} {
		a, e := idx.Rank(q)
		if e != nil {
			t.Fatal(e)
		}
		b, e := idx.RankLongInto(q, Ranking{})
		if e != nil || !reflect.DeepEqual(a, b) {
			t.Fatalf("parity error %v", e)
		}
	}
	dst := Ranking{Order: []int{2, 1, 0}, Scores: []float64{7, 8, 9}}
	for _, q := range []string{"", " \t", strings.Repeat("x", MaxLongQueryBytes+1), string([]byte{0xff})} {
		if _, e := idx.RankLongInto(q, dst); e == nil {
			t.Fatal("accepted invalid input")
		}
		if !reflect.DeepEqual(dst, Ranking{Order: []int{2, 1, 0}, Scores: []float64{7, 8, 9}}) {
			t.Fatal("mutated invalid destination")
		}
	}
	q := strings.Repeat(" ", MaxLongQueryBytes-len("한글")) + "한글"
	if _, e := idx.RankLongInto(q, dst); e != nil {
		t.Fatal(e)
	}
}

func TestLongQueryConcurrent(t *testing.T) {
	idx, _ := New([]string{"needle", "hay"})
	q := strings.Repeat("unknown ", 1100) + "needle"
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var dst Ranking
			for range 10 {
				var e error
				dst, e = idx.RankLongInto(q, dst)
				if e != nil || dst.Order[0] != 0 {
					t.Error("concurrent ranking failed")
					return
				}
			}
		}()
	}
	wg.Wait()
}
