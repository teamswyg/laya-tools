package retrievalbench

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"

	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const CostClaimPlanSHA256 = "451b90fb88ed318d75f61a2ff7ccc83e3f32376cafaba507da36514fa69e78dc"

var costPenalties = [...]float64{0, .25, 1, 4, 16}

type CostClaimHead struct {
	Head      ClaimHead
	Penalty   float64
	Threshold BudgetThreshold
}
type CostCandidate struct {
	File, SHA256, EvaluationRepository string
	Bytes, Epoch                       int
	Seed                               uint64
	Penalty, ValidationWeightedNLL     float64
	Threshold                          BudgetThreshold
}
type CostFold struct {
	Policy, EvaluationRepository, File, SHA256 string
	Seed                                       uint64
	Penalty                                    float64
	Threshold                                  BudgetThreshold
	Metrics                                    PageClaimMetrics
}
type CostAggregate struct {
	Policy                string
	Seed                  uint64
	Metrics               PageClaimMetrics
	PassesExploratoryGate bool
}
type CostClaimReport struct {
	Schema, PlanSHA256                                                              string
	Questions, Dimension, PageSize, PageTies                                        int
	Baseline, AlwaysHelper                                                          PageClaimMetrics
	Candidates                                                                      []CostCandidate
	Folds                                                                           []CostFold
	Aggregates                                                                      []CostAggregate
	PrimaryPassesExploratoryGate, ReplicationPassesExploratoryGate, ProductionReady bool
}

func costClaimDataset(xs []ClaimExample, repo, split string, penalty float64) (pairlearn.Dataset, error) {
	if math.IsNaN(penalty) || math.IsInf(penalty, 0) || penalty < 0 || penalty > 16 {
		return pairlearn.Dataset{}, fmt.Errorf("invalid call penalty")
	}
	d := claimDataset(xs, repo, split)
	d.SampleWeights = make([]float64, len(d.Labels))
	i := 0
	for _, x := range xs {
		if x.Repository != repo {
			continue
		}
		delta := float64((x.BaselineRank+19)/20-(x.InterleavedRank+19)/20) - penalty
		d.Labels[i] = 0
		if delta > 0 {
			d.Labels[i] = 1
		}
		d.SampleWeights[i] = math.Abs(delta)
		i++
	}
	return d, nil
}
func betterCostCandidate(a, b CostCandidate) bool {
	if a.Threshold.ValidationPages != b.Threshold.ValidationPages {
		return a.Threshold.ValidationPages < b.Threshold.ValidationPages
	}
	if a.Threshold.ValidationCalls != b.Threshold.ValidationCalls {
		return a.Threshold.ValidationCalls < b.Threshold.ValidationCalls
	}
	return a.Penalty < b.Penalty
}
func costScores(xs []ClaimExample, w []float64) []float64 {
	r := make([]float64, len(xs))
	for i, x := range xs {
		if w == nil {
			r[i] = -x.Features[9]
		} else {
			r[i] = searchclaim.Score(w, x.Features)
		}
	}
	return r
}
func TrainCostClaims(xs []ClaimExample, out, plan string) (CostClaimReport, error) {
	r := CostClaimReport{Schema: "riido-cost-claim-report-v1", PlanSHA256: plan, Questions: len(xs), Dimension: searchclaim.Dimension, PageSize: 20}
	if plan != CostClaimPlanSHA256 || len(xs) != 2948 {
		return r, fmt.Errorf("plan or question count mismatch")
	}
	counts := [3]int{}
	seen := make([]bool, len(xs))
	for _, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] || x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > min(3009, 2*x.BaselineRank-1) {
			return r, fmt.Errorf("invalid frozen membership/ranks")
		}
		for _, v := range x.Features {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return r, fmt.Errorf("invalid feature")
			}
		}
		seen[x.Group] = true
		counts[repo]++
		if (x.BaselineRank+19)/20 == (x.InterleavedRank+19)/20 {
			r.PageTies++
		}
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("repository count mismatch")
	}
	if err := os.Mkdir(out, 0700); err != nil {
		return r, err
	}
	r.Baseline = measurePageClaims(xs, nil, "baseline")
	r.AlwaysHelper = measurePageClaims(xs, nil, "always_interleave")
	var combined [5][]ClaimExample
	for fold, testRepo := range claimRepositories {
		trainRepo, valRepo := claimRepositories[(fold+1)%3], claimRepositories[(fold+2)%3]
		var validation, test []ClaimExample
		for _, x := range xs {
			if x.Repository == valRepo {
				validation = append(validation, x)
			}
			if x.Repository == testRepo {
				test = append(test, x)
			}
		}
		gapThreshold, err := CalibratePageBudget(validation, costScores(validation, nil), 90)
		if err != nil {
			return r, err
		}
		selected := applyPageBudget(test, costScores(test, nil), gapThreshold)
		combined[0] = append(combined[0], selected...)
		r.Folds = append(r.Folds, CostFold{Policy: "gap_heuristic", EvaluationRepository: testRepo, Threshold: gapThreshold, Metrics: budgetMetrics(selected)})
		for si, seed := range []uint64{1729, 2718} {
			var candidates [5]CostCandidate
			var weights [5][]float64
			for pi, penalty := range costPenalties {
				train, err := costClaimDataset(xs, trainRepo, "development", penalty)
				if err != nil {
					return r, err
				}
				val, err := costClaimDataset(xs, valRepo, "validation", penalty)
				if err != nil {
					return r, err
				}
				fit, err := pairlearn.Fit(train, val, pairlearn.Config{Seed: seed, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 100, Batch: 64, Dimension: searchclaim.Dimension})
				if err != nil {
					return r, err
				}
				threshold, err := CalibratePageBudget(validation, costScores(validation, fit.Weights), 90)
				if err != nil {
					return r, err
				}
				nll := pairlearn.NLL(val, fit.Weights)
				head := CostClaimHead{ClaimHead{"riido-cost-claim-v1", "fp32", trainRepo, valRepo, testRepo, plan, seed, fit.Epoch, nll, fit.Weights}, penalty, threshold}
				data, err := json.MarshalIndent(head, "", "  ")
				if err != nil {
					return r, err
				}
				data = append(data, '\n')
				file := fmt.Sprintf("fold%d-penalty%d-fp32-%d.claim.json", fold, pi, seed)
				if err = os.WriteFile(filepath.Join(out, file), data, 0600); err != nil {
					return r, err
				}
				candidates[pi] = CostCandidate{file, paireval.Hash(data), testRepo, len(data), fit.Epoch, seed, penalty, nll, threshold}
				weights[pi] = fit.Weights
				r.Candidates = append(r.Candidates, candidates[pi])
			}
			best := 0
			for i := 1; i < len(candidates); i++ {
				if betterCostCandidate(candidates[i], candidates[best]) {
					best = i
				}
			}
			for policy, pi := range []int{0, best} {
				c := candidates[pi]
				selected := applyPageBudget(test, costScores(test, weights[pi]), c.Threshold)
				slot := 1 + policy*2 + si
				combined[slot] = append(combined[slot], selected...)
				r.Folds = append(r.Folds, CostFold{[]string{"zero_penalty", "cost_selected"}[policy], testRepo, c.File, c.SHA256, seed, c.Penalty, c.Threshold, budgetMetrics(selected)})
			}
		}
	}
	control := budgetMetrics(combined[0])
	for slot, selected := range combined {
		m := budgetMetrics(selected)
		policy := "gap_heuristic"
		var seed uint64
		if slot > 0 {
			policy = []string{"zero_penalty", "cost_selected"}[(slot-1)/2]
			seed = []uint64{1729, 2718}[(slot-1)%2]
		}
		pass := m.Pages <= r.AlwaysHelper.Pages && m.Quality.AuxiliaryCalls*10 <= r.Questions*9 && m.Quality.Recall1 >= r.Baseline.Quality.Recall1 && m.Quality.Recall10 >= r.Baseline.Quality.Recall10
		if slot > 0 {
			pass = pass && m.Pages <= control.Pages && m.Quality.AuxiliaryCalls <= control.Quality.AuxiliaryCalls && (m.Pages < control.Pages || m.Quality.AuxiliaryCalls < control.Quality.AuxiliaryCalls)
		}
		r.Aggregates = append(r.Aggregates, CostAggregate{policy, seed, m, pass})
		if slot == 3 {
			r.PrimaryPassesExploratoryGate = pass
		}
		if slot == 4 {
			r.ReplicationPassesExploratoryGate = pass
		}
	}
	return r, nil
}
