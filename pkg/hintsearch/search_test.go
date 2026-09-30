package hintsearch

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestBM25Reference(t *testing.T) {
	idx, err := New([]string{"alpha alpha beta", "alpha gamma", "delta"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := idx.Rank("alpha alpha")
	if err != nil {
		t.Fatal(err)
	}
	idf := math.Log(1 + (3.-2.+.5)/(2.+.5))
	want := []float64{idf * 2 * 2.2 / (2 + 1.2*(.25+.75*3/2)), idf * 2.2 / (1 + 1.2*(.25+.75*2/2)), 0}
	for i := range want {
		if math.Abs(r.Scores[i]-want[i]) > 1e-12 {
			t.Fatalf("score %d = %g want %g", i, r.Scores[i], want[i])
		}
	}
	if !reflect.DeepEqual(r.Order, []int{0, 1, 2}) {
		t.Fatal(r.Order)
	}
}

func TestRankIntoResetsScoresAndRetainsTies(t *testing.T) {
	idx, _ := New([]string{"alpha alpha beta", "alpha gamma", "delta"})
	var scratch Ranking
	for _, query := range []string{"alpha", "delta", "unknown", "!!!", "alpha alpha", "gamma", "beta"} {
		want, err := idx.Rank(query)
		if err != nil {
			t.Fatal(err)
		}
		scratch, err = idx.RankInto(query, scratch)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(scratch, want) {
			t.Fatalf("stale ranking for %q: %+v != %+v", query, scratch, want)
		}
	}
	orderPtr, scorePtr := &scratch.Order[0], &scratch.Scores[0]
	keptOrder := append([]int(nil), scratch.Order...)
	keptScores := append([]float64(nil), scratch.Scores...)
	if _, err := idx.RankInto(" ", scratch); err == nil {
		t.Fatal("accepted invalid query")
	}
	if !reflect.DeepEqual(scratch.Order, keptOrder) || !reflect.DeepEqual(scratch.Scores, keptScores) {
		t.Fatal("invalid input mutated result")
	}
	scratch, _ = idx.RankInto("unknown", scratch)
	if &scratch.Order[0] != orderPtr || &scratch.Scores[0] != scorePtr {
		t.Fatal("did not reuse arrays")
	}
	if !reflect.DeepEqual(scratch.Order, []int{0, 1, 2}) || !reflect.DeepEqual(scratch.Scores, []float64{0, 0, 0}) {
		t.Fatal("lost zero-score candidates")
	}
}

func TestRankIntoResizeAcrossIndexes(t *testing.T) {
	var scratch Ranking
	for _, docs := range [][]string{{"one"}, {"one", "two", "three", "one"}, {"two", "one"}} {
		idx, _ := New(docs)
		var err error
		scratch, err = idx.RankInto("one", scratch)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := idx.Rank("one")
		if !reflect.DeepEqual(scratch, want) {
			t.Fatal("resize changed ranking")
		}
	}
}

func TestConcurrentRankInto(t *testing.T) {
	idx, _ := New([]string{"alpha", "beta", "gamma"})
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Go(func() {
			var scratch Ranking
			for i := 0; i < 100; i++ {
				var err error
				scratch, err = idx.RankInto("beta", scratch)
				if err != nil || scratch.Order[0] != 1 || scratch.Scores[0] != 0 {
					t.Error("workspace interference")
				}
				scratch.Order[0] = 0
				scratch.Scores[0] = 999
			}
		})
	}
	wg.Wait()
}

func BenchmarkRankReuse(b *testing.B) {
	for _, n := range []int{64, 3009} {
		docs := make([]string, n)
		for i := range docs {
			docs[i] = fmt.Sprintf("payment idempotency retries item%d group%d", i, i%17)
		}
		idx, _ := New(docs)
		for _, reuse := range []bool{false, true} {
			b.Run(fmt.Sprintf("documents=%d/reuse=%t", n, reuse), func(b *testing.B) {
				var scratch Ranking
				if reuse {
					scratch, _ = idx.Rank("payment group3")
				}
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					var err error
					if reuse {
						scratch, err = idx.RankInto("payment group3", scratch)
					} else {
						scratch, err = idx.Rank("payment group3")
					}
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
func TestEmptyVocabularyAndUnicode(t *testing.T) {
	idx, _ := New([]string{"", "...", "결제 중복 방지"})
	r, _ := idx.Rank("결제")
	if r.Order[0] != 2 {
		t.Fatal(r.Order)
	}
	r, _ = idx.Rank("unknown")
	if !reflect.DeepEqual(r.Order, []int{0, 1, 2}) {
		t.Fatal(r.Order)
	}
	if _, err := idx.Rank(" "); err == nil {
		t.Fatal("empty accepted")
	}
	if _, err := idx.Rank(strings.Repeat("a", MaxQueryBytes+1)); err == nil {
		t.Fatal("oversize accepted")
	}
}
func TestBounds(t *testing.T) {
	for _, docs := range [][]string{nil, make([]string, MaxDocuments+1), {strings.Repeat("x", MaxCatalogBytes+1)}} {
		if _, err := New(docs); err == nil {
			t.Fatal("invalid catalog accepted")
		}
	}
	for _, args := range [][2][]int{{nil, nil}, {{0, 0}, nil}, {{1}, nil}, {{0}, {1}}, {{0}, {0, 0}}} {
		if _, err := Interleave(args[0], args[1]); err == nil {
			t.Fatal("invalid permutation accepted")
		}
	}
}
func TestInterleaveBoundExhaustive(t *testing.T) {
	for n := 1; n <= 7; n++ {
		base := make([]int, n)
		h := make([]int, n)
		for i := range base {
			base[i] = n - 1 - i
			h[i] = i
		}
		var visit func(int)
		visit = func(k int) {
			if k == n {
				for length := 0; length <= n; length++ {
					out, err := Interleave(base, h[:length])
					if err != nil {
						t.Fatal(err)
					}
					pos := make([]int, n)
					for i, d := range out {
						pos[d] = i + 1
					}
					for i, d := range base {
						if pos[d] == 0 || pos[d] > min(n, 2*(i+1)) {
							t.Fatalf("bound violation base=%v hints=%v out=%v", base, h[:length], out)
						}
					}
				}
				return
			}
			for j := k; j < n; j++ {
				h[k], h[j] = h[j], h[k]
				visit(k + 1)
				h[k], h[j] = h[j], h[k]
			}
		}
		visit(0)
	}
}
func TestConcurrentSnapshots(t *testing.T) {
	docs := []string{"payment idempotency", "cache expiry"}
	idx, _ := New(docs)
	docs[0] = "changed"
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			for j := 0; j < 100; j++ {
				r, err := idx.Rank("payment")
				if err != nil || r.Order[0] != 0 || r.Scores[0] <= 0 {
					t.Error("snapshot changed")
				}
				r.Order[0] = 1
			}
		})
	}
	wg.Wait()
}
func BenchmarkBM25(b *testing.B) {
	docs := make([]string, 64)
	for i := range docs {
		docs[i] = "payment idempotency prevents duplicate charges across retries"
	}
	idx, _ := New(docs)
	b.ReportAllocs()
	for b.Loop() {
		idx.Rank("prevent duplicate payment")
	}
}
