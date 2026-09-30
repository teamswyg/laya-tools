package retrievalbench

import (
	"reflect"
	"testing"
)

func TestPageClaimLabelsAndRepositoryIsolation(t *testing.T) {
	xs := []ClaimExample{{Repository: "a", Group: 1, BaselineRank: 20, InterleavedRank: 21}, {Repository: "b", Group: 2, BaselineRank: 61, InterleavedRank: 1}, {Repository: "a", Group: 3, BaselineRank: 41, InterleavedRank: 1}, {Repository: "a", Group: 4, BaselineRank: 2, InterleavedRank: 1}}
	for i := range xs {
		xs[i].Features[0] = 1
	}
	d := pageClaimDataset(xs, "a", "development")
	if !reflect.DeepEqual(d.Groups, []int{1, 3, 4}) || !reflect.DeepEqual(d.Labels, []float64{0, 1, 0}) || !reflect.DeepEqual(d.SampleWeights, []float64{1, 2, 0}) {
		t.Fatalf("wrong labels/split: %+v", d)
	}
	d2 := pageClaimDataset(xs, "a", "development")
	xs[1].BaselineRank = 1000
	if !reflect.DeepEqual(d2, pageClaimDataset(xs, "a", "development")) {
		t.Fatal("held-out outcome changed train dataset")
	}
}
func TestPageClaimSelectedCost(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 20, InterleavedRank: 21}}
	xs[0].Features[0] = 1
	xs[1].Features[0] = -1
	w := make([]float64, 12)
	w[0] = 1
	m := measurePageClaims(xs, w, "learned")
	if m.Pages != 2 || m.EmittedCandidates != 40 || m.Quality.AuxiliaryCalls != 1 {
		t.Fatalf("selected costs: %+v", m)
	}
	base := measurePageClaims(xs, nil, "baseline")
	all := measurePageClaims(xs, nil, "always_interleave")
	if base.Pages != 4 || all.Pages != 3 {
		t.Fatal("controls changed")
	}
}
