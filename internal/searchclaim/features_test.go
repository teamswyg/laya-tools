package searchclaim

import (
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
	"testing"
)

func TestRuntimeOnlySignals(t *testing.T) {
	idx, _ := hintsearch.New([]string{"ReadFile file contents", "cache delete"})
	r, _ := idx.Rank("ReadFile")
	f, e := Features("ReadFile", r)
	if e != nil {
		t.Fatal(e)
	}
	if f[0] != 1 || f[6] != 1 || f[7] != 1 || f[11] != .5 {
		t.Fatalf("unexpected signals %v", f)
	}
	for _, v := range f {
		if v < 0 || v > 1 || math.IsNaN(v) {
			t.Fatal("unbounded feature")
		}
	}
	r.Scores[0] = math.NaN()
	if _, e := Features("ReadFile", r); e == nil {
		t.Fatal("accepted nonfinite score")
	}
}

func TestSignalsDoNotNeedAuxiliaryResults(t *testing.T) {
	r := hintsearch.Ranking{Order: []int{0, 1}, Scores: []float64{0, 0}}
	f, e := Features("unknown query", r)
	if e != nil {
		t.Fatal(e)
	}
	if f[8] != 0 || f[9] != 0 || f[10] != 0 || f[11] != 0 {
		t.Fatal("unknown query gained evidence")
	}
}

func BenchmarkFeatures(b *testing.B) {
	docs := make([]string, 3009)
	for i := range docs {
		docs[i] = "read file contents safely"
	}
	idx, _ := hintsearch.New(docs)
	r, _ := idx.Rank("read file")
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, e := Features("read file", r); e != nil {
			b.Fatal(e)
		}
	}
}
