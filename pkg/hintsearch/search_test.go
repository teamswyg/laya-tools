package hintsearch

import (
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
