package retrievalbench

import (
	"reflect"
	"testing"
)

func TestClaimFeaturesDoNotUseGoldNames(t *testing.T) {
	rows := []Row{{Repository: "a", Code: "func ReadFile() {}", Name: "ReadFile", Query: "read file"}, {Repository: "b", Code: "func DeleteCache() {}", Name: "DeleteCache", Query: "delete cache"}}
	a, e := PrepareClaims(rows)
	if e != nil {
		t.Fatal(e)
	}
	rows[0].Name = "irrelevant gold name"
	rows[1].Name = "another gold name"
	b, e := PrepareClaims(rows)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("gold name affected runtime features")
	}
}

func TestHeldOutExamplesDoNotEnterTrainingColumns(t *testing.T) {
	xs := []ClaimExample{{Repository: "train", Group: 1, BaselineRank: 3, InterleavedRank: 2}, {Repository: "validation", Group: 2, BaselineRank: 1, InterleavedRank: 2}, {Repository: "test", Group: 3, BaselineRank: 7, InterleavedRank: 3}}
	train := claimDataset(xs, "train", "development")
	val := claimDataset(xs, "validation", "validation")
	xs[2].Features[0] = 999
	xs[2].InterleavedRank = 100
	xs[2].BaselineRank = 1
	if !reflect.DeepEqual(train, claimDataset(xs, "train", "development")) || !reflect.DeepEqual(val, claimDataset(xs, "validation", "validation")) {
		t.Fatal("held-out metadata changed fitting data")
	}
}

func TestClaimMetricsCompareAgainstBaseline(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 1, InterleavedRank: 2}, {BaselineRank: 10, InterleavedRank: 3}, {BaselineRank: 3, InterleavedRank: 3}}
	b := measureClaims(xs, nil, "baseline")
	a := measureClaims(xs, nil, "always_interleave")
	if b.MeanRank != 14./3 || b.Wins != 0 || b.AuxiliaryCalls != 0 || a.MeanRank != 8./3 || a.Wins != 1 || a.Losses != 1 || a.Ties != 1 || a.AuxiliaryCalls != 3 {
		t.Fatalf("incorrect policy metrics: %+v %+v", b, a)
	}
}
