package retrievalbench

import (
	"math"
	"reflect"
	"testing"
)

func TestCostObjectiveIncludesTieCalls(t *testing.T) {
	xs := []ClaimExample{{Repository: "a", Group: 1, BaselineRank: 20, InterleavedRank: 21}, {Repository: "b", Group: 2, BaselineRank: 61, InterleavedRank: 1}, {Repository: "a", Group: 3, BaselineRank: 41, InterleavedRank: 1}, {Repository: "a", Group: 4, BaselineRank: 2, InterleavedRank: 1}}
	for i := range xs {
		xs[i].Features[0] = 1
	}
	zero, err := costClaimDataset(xs, "a", "development", 0)
	if err != nil || !reflect.DeepEqual(zero, pageClaimDataset(xs, "a", "development")) {
		t.Fatal("zero penalty changed control", err)
	}
	d, err := costClaimDataset(xs, "a", "development", .25)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d.Labels, []float64{0, 1, 0}) || !reflect.DeepEqual(d.SampleWeights, []float64{1.25, 1.75, .25}) {
		t.Fatalf("wrong net-cost target %+v", d)
	}
	xs[1].BaselineRank = 3009
	again, err := costClaimDataset(xs, "a", "development", .25)
	if err != nil || !reflect.DeepEqual(d, again) {
		t.Fatal("held-out labels changed training")
	}
	d, err = costClaimDataset(xs, "a", "development", 4)
	if err != nil || !reflect.DeepEqual(d.Labels, []float64{0, 0, 0}) {
		t.Fatal("call penalty cannot reverse positive target")
	}
	for _, p := range []float64{-1, 17, math.NaN(), math.Inf(1)} {
		if _, err := costClaimDataset(xs, "a", "development", p); err == nil {
			t.Fatal("accepted invalid penalty")
		}
	}
}
func TestCostSelectionUsesValidationPagesThenCallsThenPenalty(t *testing.T) {
	base := CostCandidate{Penalty: 1, Threshold: BudgetThreshold{ValidationPages: 10, ValidationCalls: 5}}
	a := base
	a.Threshold.ValidationPages = 9
	a.Threshold.ValidationCalls = 9
	if !betterCostCandidate(a, base) {
		t.Fatal("did not prioritize pages")
	}
	a = base
	a.Threshold.ValidationCalls = 4
	if !betterCostCandidate(a, base) {
		t.Fatal("did not break ties by calls")
	}
	a = base
	a.Penalty = .25
	if !betterCostCandidate(a, base) || betterCostCandidate(base, a) {
		t.Fatal("penalty tie break failed")
	}
	a = base
	a.EvaluationRepository = "different"
	a.File = "different"
	a.ValidationWeightedNLL = -100
	if betterCostCandidate(a, base) || betterCostCandidate(base, a) {
		t.Fatal("used unrelated metadata to choose candidate")
	}
}

func TestExtraFeaturesPreserveTargetsAndSplit(t *testing.T) {
	xs := []ClaimExample{{Repository: "train", Group: 1, BaselineRank: 41, InterleavedRank: 1}, {Repository: "test", Group: 2, BaselineRank: 1, InterleavedRank: 1}}
	xs[0].Features[0] = 1
	extra := [][4]float64{{.5, 0, .25, 1}, {1, 1, 1, 1}}
	old, _ := costClaimDataset(xs, "train", "development", .25)
	d, e := costClaimDatasetExtra(xs, extra, "train", "development", .25)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(d.Labels, old.Labels) || !reflect.DeepEqual(d.SampleWeights, old.SampleWeights) || !reflect.DeepEqual(d.Groups, old.Groups) || !reflect.DeepEqual(d.Indices, []uint16{0, 12, 14, 15}) {
		t.Fatalf("unexpected extra dataset %+v", d)
	}
	xs[1].BaselineRank = 3009
	extra[1] = [4]float64{0, 0, 0, 0}
	again, e := costClaimDatasetExtra(xs, extra, "train", "development", .25)
	if e != nil || !reflect.DeepEqual(d, again) {
		t.Fatal("held-out state changed training")
	}
	if _, e := costClaimDatasetExtra(xs, extra[:1], "train", "development", .25); e == nil {
		t.Fatal("accepted short features")
	}
	extra[0][0] = math.NaN()
	if _, e := costClaimDatasetExtra(xs, extra, "train", "development", .25); e == nil {
		t.Fatal("accepted nonfinite feature")
	}
}
