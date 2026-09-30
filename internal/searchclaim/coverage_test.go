package searchclaim

import (
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
	"testing"
)

func TestCoverageFeatureValues(t *testing.T) {
	idx, e := hintsearch.New([]string{"alpha beta", "beta", "other"})
	if e != nil {
		t.Fatal(e)
	}
	r, e := idx.Rank("alpha beta absent")
	if e != nil {
		t.Fatal(e)
	}
	f, e := QueryCoverage("alpha beta absent", r, idx)
	if e != nil {
		t.Fatal(e)
	}
	want := [4]float64{2.0 / 3, 1.0 / 3, math.Sqrt(2.0 / 27), 1.0 / 3}
	for i := range f {
		if math.Abs(f[i]-want[i]) > 1e-12 {
			t.Fatalf("wrong feature %v want %v", f, want)
		}
	}
	// Scores are not inputs to term-presence features.
	for i := range r.Scores {
		r.Scores[i] = math.NaN()
	}
	again, e := QueryCoverage("alpha beta absent", r, idx)
	if e != nil || f != again {
		t.Fatal("coverage read scores")
	}
}
