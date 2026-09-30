package searchclaim

import (
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
	"testing"
)

func TestScoreSpreadCasesAndScale(t *testing.T) {
	r := hintsearch.Ranking{Order: []int{0, 1, 2, 3}, Scores: []float64{4, 4, 4, 4}}
	got, e := ScoreSpread(r)
	if e != nil || got != [4]float64{1, 0, .25, 1} {
		t.Fatalf("uniform %v %v", got, e)
	}
	r.Scores = []float64{4, 0, 0, 0}
	got, e = ScoreSpread(r)
	if e != nil || got[0] != .25 || got[2] != 1 || got[3] != 0 || math.Abs(got[1]-math.Sqrt(.1875)) > 1e-12 {
		t.Fatalf("concentrated %v %v", got, e)
	}
	r.Scores = []float64{1e308, 0, 0, 0}
	huge, e := ScoreSpread(r)
	if e != nil || huge != got {
		t.Fatal("scale invariance/overflow", huge, e)
	}
	r.Scores = []float64{0, 0, 0, 0}
	got, e = ScoreSpread(r)
	if e != nil || got != [4]float64{} {
		t.Fatal("zero scores gained evidence")
	}
	for _, scores := range [][]float64{{-1, 0, 0, 0}, {1, 2, 0, 0}, {math.NaN(), 0, 0, 0}, {math.Inf(1), 0, 0, 0}} {
		r.Scores = scores
		if _, e := ScoreSpread(r); e == nil {
			t.Fatal("accepted invalid score")
		}
	}
	r.Order[0] = 4
	if _, e := ScoreSpread(r); e == nil {
		t.Fatal("accepted invalid index")
	}
}
func spreadRanking() hintsearch.Ranking {
	r := hintsearch.Ranking{Order: make([]int, 3009), Scores: make([]float64, 3009)}
	for i := range r.Order {
		r.Order[i] = i
		r.Scores[i] = float64(len(r.Order) - i)
	}
	return r
}
func TestSpreadPrefixAndAllocation(t *testing.T) {
	r := spreadRanking()
	a, e := ScoreSpread(r)
	if e != nil {
		t.Fatal(e)
	}
	r.Scores[25] = math.NaN()
	b, e := ScoreSpread(r)
	if e != nil || a != b {
		t.Fatal("read outside top20")
	}
	if n := testing.AllocsPerRun(100, func() {
		_, e := ScoreSpread(r)
		if e != nil {
			panic(e)
		}
	}); n != 0 {
		t.Fatalf("allocated %v", n)
	}
}
func BenchmarkSpreadExtra(b *testing.B) {
	r := spreadRanking()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, e := ScoreSpread(r); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkFeatureVariants(b *testing.B) {
	r := spreadRanking()
	for _, extra := range []bool{false, true} {
		name := "base12"
		if extra {
			name = "spread16"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, e := Features("read file contents", r); e != nil {
					b.Fatal(e)
				}
				if extra {
					if _, e := ScoreSpread(r); e != nil {
						b.Fatal(e)
					}
				}
			}
		})
	}
}
