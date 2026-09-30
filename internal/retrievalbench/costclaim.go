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

const SpreadClaimPlanSHA256 = "07962577cf323d1924246374af4360763b03d22fdb55921a981b19b3a2231883"

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
func costScores(xs []ClaimExample, w []float64) []float64 { return costScoresExtra(xs, nil, w) }
func costScoresExtra(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, w []float64) []float64 {
	r := make([]float64, len(xs))
	for i, x := range xs {
		if w == nil {
			r[i] = -x.Features[9]
		} else {
			r[i] = searchclaim.Score(w, x.Features)
			if extra != nil {
				for j, v := range extra[i] {
					r[i] += w[searchclaim.Dimension+j] * v
				}
			}
		}
	}
	return r
}
func TrainCostClaims(xs []ClaimExample, out, plan string) (CostClaimReport, error) {
	return trainCostClaims(xs, nil, out, plan)
}
func TrainSpreadClaims(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, out, plan string) (CostClaimReport, error) {
	if len(extra) != len(xs) {
		return CostClaimReport{}, fmt.Errorf("feature row count mismatch")
	}
	return trainCostClaims(xs, extra, out, plan)
}

const CoverageClaimPlanSHA256 = "57f8898268ffd08bf6c81d08a7d0461843c4238d9c0cbc7beb1aafac752fa88a"

// Extended columns preserve existing 12/16-feature schemas and arithmetic.
// Rows are fixed contiguous arrays; active width is frozen by the experiment.
func extendClaimColumns(spread [][searchclaim.SpreadDimension]float64) [][8]float64 {
	if spread == nil {
		return nil
	}
	out := make([][8]float64, len(spread))
	for i := range spread {
		copy(out[i][:4], spread[i][:])
	}
	return out
}
func TrainCoverageClaims(xs []ClaimExample, spread [][searchclaim.SpreadDimension]float64, coverage [][searchclaim.CoverageDimension]float64, out, plan string) (CostClaimReport, error) {
	if len(spread) != len(xs) || len(coverage) != len(xs) {
		return CostClaimReport{}, fmt.Errorf("coverage row count mismatch")
	}
	values := extendClaimColumns(spread)
	for i := range coverage {
		copy(values[i][4:], coverage[i][:])
	}
	return trainClaimColumns(xs, values, 8, out, plan)
}
func trainCostClaims(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, out, plan string) (CostClaimReport, error) {
	width := 0
	if extra != nil {
		width = 4
	}
	return trainClaimColumns(xs, extendClaimColumns(extra), width, out, plan)
}
func costScoresColumns(xs []ClaimExample, extra [][8]float64, width int, w []float64) []float64 {
	r := costScores(xs, w)
	if w != nil {
		for i := range xs {
			for j := 0; j < width; j++ {
				r[i] += w[searchclaim.Dimension+j] * extra[i][j]
			}
		}
	}
	return r
}
func trainClaimColumns(xs []ClaimExample, extra [][8]float64, width int, out, plan string) (CostClaimReport, error) {
	expectedPlan, headSchema, reportSchema := CostClaimPlanSHA256, "riido-cost-claim-v1", "riido-cost-claim-report-v1"
	switch width {
	case 0:
	case 4:
		expectedPlan, headSchema, reportSchema = SpreadClaimPlanSHA256, "riido-spread-claim-v1", "riido-spread-claim-report-v1"
	case 8:
		expectedPlan, headSchema, reportSchema = CoverageClaimPlanSHA256, "riido-coverage-claim-v1", "riido-coverage-claim-report-v1"
	default:
		return CostClaimReport{}, fmt.Errorf("invalid feature width")
	}
	if (width == 0 && len(extra) != 0) || (width > 0 && len(extra) != len(xs)) {
		return CostClaimReport{}, fmt.Errorf("extra column shape mismatch")
	}
	dimension := searchclaim.Dimension + width
	r := CostClaimReport{Schema: reportSchema, PlanSHA256: plan, Questions: len(xs), Dimension: dimension, PageSize: 20}
	if plan != expectedPlan || len(xs) != 2948 {
		return r, fmt.Errorf("plan or question count mismatch")
	}
	for _, row := range extra {
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return r, fmt.Errorf("invalid extra feature")
			}
		}
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
		var valExtra, testExtra [][8]float64
		for xi, x := range xs {
			if x.Repository == valRepo {
				validation = append(validation, x)
				if extra != nil {
					valExtra = append(valExtra, extra[xi])
				}
			}
			if x.Repository == testRepo {
				test = append(test, x)
				if extra != nil {
					testExtra = append(testExtra, extra[xi])
				}
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
				train, err := costClaimDatasetColumns(xs, extra, width, trainRepo, "development", penalty)
				if err != nil {
					return r, err
				}
				val, err := costClaimDatasetColumns(xs, extra, width, valRepo, "validation", penalty)
				if err != nil {
					return r, err
				}
				fit, err := pairlearn.Fit(train, val, pairlearn.Config{Seed: seed, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 100, Batch: 64, Dimension: dimension})
				if err != nil {
					return r, err
				}
				threshold, err := CalibratePageBudget(validation, costScoresColumns(validation, valExtra, width, fit.Weights), 90)
				if err != nil {
					return r, err
				}
				nll := pairlearn.NLL(val, fit.Weights)
				head := CostClaimHead{ClaimHead{headSchema, "fp32", trainRepo, valRepo, testRepo, plan, seed, fit.Epoch, nll, fit.Weights}, penalty, threshold}
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
				selected := applyPageBudget(test, costScoresColumns(test, testExtra, width, weights[pi]), c.Threshold)
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

func costClaimDatasetExtra(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, repo, split string, penalty float64) (pairlearn.Dataset, error) {
	width := 0
	if extra != nil {
		width = 4
	}
	return costClaimDatasetColumns(xs, extendClaimColumns(extra), width, repo, split, penalty)
}
func costClaimDatasetColumns(xs []ClaimExample, extra [][8]float64, width int, repo, split string, penalty float64) (pairlearn.Dataset, error) {
	if width != 0 && width != 4 && width != 8 {
		return pairlearn.Dataset{}, fmt.Errorf("invalid extra width")
	}

	if (width == 0 && len(extra) != 0) || (width > 0 && len(extra) != len(xs)) {
		return pairlearn.Dataset{}, fmt.Errorf("extra column shape mismatch")
	}
	d, err := costClaimDataset(xs, repo, split, penalty)
	if err != nil || extra == nil {
		return d, err
	}
	if len(extra) != len(xs) {
		return d, fmt.Errorf("extra feature length mismatch")
	}
	offsets := []int{0}
	indices := make([]uint16, 0, len(d.Indices)+len(d.Labels)*width)
	values := make([]float64, 0, cap(indices))
	row := 0
	for i, x := range xs {
		if x.Repository != repo {
			continue
		}
		indices = append(indices, d.Indices[d.Offsets[row]:d.Offsets[row+1]]...)
		values = append(values, d.Values[d.Offsets[row]:d.Offsets[row+1]]...)
		for j, v := range extra[i][:width] {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return d, fmt.Errorf("invalid extra feature")
			}
			if v != 0 {
				indices = append(indices, uint16(searchclaim.Dimension+j))
				values = append(values, v)
			}
		}
		offsets = append(offsets, len(indices))
		row++
	}
	d.Offsets, d.Indices, d.Values = offsets, indices, values
	return d, nil
}
