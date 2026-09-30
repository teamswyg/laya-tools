package retrievalbench

import (
	"slices"
	"testing"
)

func TestSimulateContinuation(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 21, InterleavedRank: 21}, {BaselineRank: 81, InterleavedRank: 81}, {BaselineRank: 81, InterleavedRank: 3, AuxiliaryRank: 1}}
	ranks := []int{1, 41, 21, 25}
	before := slices.Clone(xs)
	r, out, err := simulateContinuation(xs, ranks, "test", "all", 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.InitialCalls != 1 || r.ContinuationCalls != 2 || r.Static.Pages != 9 || r.Deferred.Pages != 7 || r.ImprovedCases != 1 || r.SavedPages != 3 || r.WorsenedCases != 1 || r.AddedPages != 1 || r.TiedCases != 2 || r.LargestIncrease != 1 {
		t.Fatalf("wrong accounting: %+v", r)
	}
	if out[0].AuxiliaryRank != 0 || out[1].AuxiliaryRank != 1 || out[2].InterleavedRank != 21 || out[3].InterleavedRank != 3 || !slices.Equal(xs, before) {
		t.Fatal("wrong continuation or input mutation")
	}
	for _, bad := range [][]int{{1}, {2, 41, 21, 25}, {1, 20, 21, 25}, {1, 3010, 21, 25}} {
		if _, _, err := simulateContinuation(xs, bad, "test", "all", 0); err == nil {
			t.Fatal("accepted bad ranks")
		}
	}
	xs[1].InterleavedRank = 1
	if _, _, err := simulateContinuation(xs, ranks, "test", "all", 0); err == nil {
		t.Fatal("accepted skipped policy changing baseline")
	}
}

func TestDeferredFirstPageBoundary(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 20, InterleavedRank: 20}, {BaselineRank: 21, InterleavedRank: 21}}
	r, _, err := simulateContinuation(xs, []int{20, 21}, "test", "all", 0)
	if err != nil || r.ContinuationCalls != 1 || r.InitialCalls != 0 || r.Deferred.Pages != 3 {
		t.Fatalf("boundary: %+v %v", r, err)
	}
}
