package retrievalbench

import (
	"math"
	"reflect"
	"testing"
)

func TestCalibratePageBudgetTiesAndObjective(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 21}, {BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 41, InterleavedRank: 1}}
	scores := []float64{4, 3, 2, 1}
	threshold, err := CalibratePageBudget(xs, scores, 50)
	if err != nil || threshold.Disabled || threshold.MinimumScore != 4 || threshold.ValidationCalls != 1 || threshold.ValidationPages != 8 {
		t.Fatalf("not the best prefix: %+v %v", threshold, err)
	}
	// Equal page totals choose fewer calls, including the disabled policy.
	ties := []ClaimExample{{BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}}
	threshold, err = CalibratePageBudget(ties, []float64{2, 1}, 100)
	if err != nil || !threshold.Disabled || threshold.ValidationCalls != 0 {
		t.Fatal("failed no-benefit tie", threshold, err)
	}
	// Two equal top scores cannot be split to satisfy a one-query budget.
	threshold, err = CalibratePageBudget(xs, []float64{4, 4, 2, 1}, 25)
	if err != nil || !threshold.Disabled {
		t.Fatal("split tied score group", threshold, err)
	}
}
func TestValidationCapDoesNotGuaranteeTestCap(t *testing.T) {
	val := []ClaimExample{{BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}}
	threshold, err := CalibratePageBudget(val, []float64{4, 3, 2, 1}, 25)
	if err != nil {
		t.Fatal(err)
	}
	test := []ClaimExample{{BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 41, InterleavedRank: 1}}
	m := budgetMetrics(applyPageBudget(test, []float64{5, 5}, threshold))
	if threshold.ValidationCalls != 1 || m.Quality.AuxiliaryCalls != 2 {
		t.Fatal("distribution shift unexpectedly capped", threshold, m)
	}
}
func TestCalibrateRejectsInvalid(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 1, InterleavedRank: 1}}
	for _, score := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := CalibratePageBudget(xs, []float64{score}, 90); err == nil {
			t.Fatal("accepted nonfinite")
		}
	}
	if _, err := CalibratePageBudget(nil, nil, 90); err == nil {
		t.Fatal("empty")
	}
	if _, err := CalibratePageBudget(xs, nil, 90); err == nil {
		t.Fatal("length")
	}
	if _, err := CalibratePageBudget(xs, []float64{1}, 101); err == nil {
		t.Fatal("budget")
	}
	xs[0].BaselineRank = 0
	if _, err := CalibratePageBudget(xs, []float64{1}, 90); err == nil {
		t.Fatal("rank")
	}
}
func TestBudgetFoldNeverCalibratesOnItsEvaluationLabels(t *testing.T) {
	var xs []ClaimExample
	for repo, n := range []int{806, 1152, 990} {
		for i := 0; i < n; i++ {
			x := ClaimExample{Repository: claimRepositories[repo], Group: len(xs), BaselineRank: 41, InterleavedRank: 1}
			x.Features[0] = 1
			x.Features[9] = float64(i) / float64(n)
			xs = append(xs, x)
		}
	}
	var heads []BudgetHead
	for fold, repo := range claimRepositories {
		for _, mode := range []string{"fp32", "int8", "ternary_ptq", "ternary_ste"} {
			for _, seed := range []uint64{1729, 2718} {
				w := make([]float64, 12)
				w[9] = -1
				heads = append(heads, BudgetHead{Head: ClaimHead{Schema: "riido-page-claim-v1", Mode: mode, Seed: seed, PlanSHA256: PageClaimPlanSHA256, TrainingRepository: claimRepositories[(fold+1)%3], ValidationRepository: claimRepositories[(fold+2)%3], EvaluationRepository: repo, Weights: w}})
			}
		}
	}
	before, err := ProbePageBudgets(xs, heads)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 806; i++ {
		xs[i].BaselineRank = 81
	}
	after, err := ProbePageBudgets(xs, heads)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Folds) != 108 || len(before.Aggregates) != 36 {
		t.Fatal("missing combinations")
	}
	for i, a := range before.Folds {
		if a.EvaluationRepository == claimRepositories[0] && !reflect.DeepEqual(a.Threshold, after.Folds[i].Threshold) {
			t.Fatal("evaluation-label leakage")
		}
	}
	for _, a := range after.Aggregates {
		if a.Metrics.Quality.Questions != 2948 {
			t.Fatal("duplicate/missing evaluation questions")
		}
	}
}
