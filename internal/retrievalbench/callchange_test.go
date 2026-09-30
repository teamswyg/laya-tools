package retrievalbench

import "testing"

func TestCallChangeConservationAndConcentration(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 21, InterleavedRank: 41}, {BaselineRank: 101, InterleavedRank: 1}, {BaselineRank: 441, InterleavedRank: 1}, {BaselineRank: 441, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}}
	old := applyPageBudget(xs, []float64{0, 1, 0, 1, 1, 0}, BudgetThreshold{MinimumScore: 1})
	new := applyPageBudget(xs, []float64{1, 0, 1, 0, 1, 0}, BudgetThreshold{MinimumScore: 1})
	r, e := summarizeCallChanges(xs, old, new, "synthetic", 1729)
	if e != nil {
		t.Fatal(e)
	}
	if r.IncreasedCases != 3 || r.IncreasedPages != 28 || r.DecreasedCases != 1 || r.DecreasedPages != 22 || r.New.Pages-r.Old.Pages != 6 || r.UnchangedCases != 2 {
		t.Fatalf("wrong accounting %+v", r)
	}
	if r.LargestPositiveSums != [3]int{22, 28, 28} || r.Buckets[0].Cases != 1 || r.Buckets[1].Cases != 1 || r.Buckets[2].Cases != 0 || r.Buckets[3].Cases != 1 {
		t.Fatal("wrong concentration/buckets")
	}
	if r.Transitions[0].Cases != 1 || r.Transitions[1].Cases != 2 || r.Transitions[2].Cases != 2 || r.Transitions[3].Cases != 1 {
		t.Fatal("wrong transition partition")
	}
	if r.Transitions[1].Top10Gains != 1 || r.Transitions[2].Top10Losses != 2 {
		t.Fatal("top10 accounting")
	}
	unchanged, e := summarizeCallChanges(xs, old, old, "synthetic", 1729)
	if e != nil || unchanged.IncreasedPages != 0 || unchanged.LargestPositiveShares != [3]float64{} {
		t.Fatal("no-loss handling")
	}
	new[0].InterleavedRank = 1
	if _, e := summarizeCallChanges(xs, old, new, "synthetic", 1729); e == nil {
		t.Fatal("accepted inconsistent applied policy")
	}
}
