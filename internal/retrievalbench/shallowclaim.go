package retrievalbench

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"

	"github.com/teamswyg/laya-tools/internal/claimtree"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const ShallowClaimPlanSHA256 = "c3d7206244ef6fe9b35cd1a62613549508e8e684949d147c304dc67d119f7291"
const ShallowControlManifestSHA256 = "d718de763f93d19adca8c6d18b77dcb091057434b29d5a34c1eda459b3032637"

type ClaimEffect struct {
	BeneficialCalls, MissedBeneficial, HarmfulCalls, AvoidedHarmful, TieCalls, SkippedTies int
	CapturedPages, MissedPages, IncurredHarmPages, AvoidedHarmPages                        int
}
type ShallowHead struct {
	Schema, PlanSHA256, TrainingRepository, ValidationRepository, EvaluationRepository string
	Depth, MinLeaf                                                                     int
	Tree                                                                               claimtree.Model
	Threshold                                                                          BudgetThreshold
}
type ShallowCandidate struct {
	File, SHA256, EvaluationRepository string
	Bytes, Depth, MinLeaf, Nodes       int
	Threshold                          BudgetThreshold
}
type ShallowFold struct {
	Policy, EvaluationRepository, File string
	Metrics                            PageClaimMetrics
	Effect                             ClaimEffect
}
type ShallowAggregate struct {
	Policy                string
	Metrics               PageClaimMetrics
	Effect                ClaimEffect
	PassesExploratoryGate bool
}
type ShallowReport struct {
	Schema, PlanSHA256, ControlManifestSHA256     string
	Questions                                     int
	Baseline, AlwaysHelper                        PageClaimMetrics
	Candidates                                    []ShallowCandidate
	Folds                                         []ShallowFold
	Aggregates                                    []ShallowAggregate
	PrimaryPassesExploratoryGate, ProductionReady bool
}

func effect(xs, selected []ClaimExample) ClaimEffect {
	var e ClaimEffect
	for i, x := range xs {
		delta := (x.BaselineRank+19)/20 - (x.InterleavedRank+19)/20
		called := selected[i].AuxiliaryRank == 1
		if delta > 0 {
			if called {
				e.BeneficialCalls++
				e.CapturedPages += delta
			} else {
				e.MissedBeneficial++
				e.MissedPages += delta
			}
		} else if delta < 0 {
			if called {
				e.HarmfulCalls++
				e.IncurredHarmPages -= delta
			} else {
				e.AvoidedHarmful++
				e.AvoidedHarmPages -= delta
			}
		} else {
			if called {
				e.TieCalls++
			} else {
				e.SkippedTies++
			}
		}
	}
	return e
}
func treeRows(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, repo string) ([]claimtree.Vector, []float64) {
	var x []claimtree.Vector
	var y []float64
	for i, row := range xs {
		if row.Repository != repo {
			continue
		}
		var v claimtree.Vector
		copy(v[:], row.Features[:])
		copy(v[12:], extra[i][:])
		x = append(x, v)
		y = append(y, float64((row.BaselineRank+19)/20-(row.InterleavedRank+19)/20))
	}
	return x, y
}
func treeScores(m *claimtree.Model, x []claimtree.Vector) []float64 {
	s := make([]float64, len(x))
	for i := range x {
		s[i] = m.Score(x[i])
	}
	return s
}
func betterTree(a, b ShallowCandidate) bool {
	if a.Threshold.ValidationPages != b.Threshold.ValidationPages {
		return a.Threshold.ValidationPages < b.Threshold.ValidationPages
	}
	if a.Threshold.ValidationCalls != b.Threshold.ValidationCalls {
		return a.Threshold.ValidationCalls < b.Threshold.ValidationCalls
	}
	if a.Depth != b.Depth {
		return a.Depth < b.Depth
	}
	return a.MinLeaf > b.MinLeaf
}

// TrainShallowClaims keeps training rows, validation selection and evaluation
// separate. Frozen controls must first pass archive verification at the caller.
func TrainShallowClaims(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, controls []CostClaimHead, out, plan string) (ShallowReport, error) {
	r := ShallowReport{Schema: "riido-shallow-claim-report-v1", PlanSHA256: plan, ControlManifestSHA256: ShallowControlManifestSHA256, Questions: len(xs)}
	if plan != ShallowClaimPlanSHA256 || len(xs) != 2948 || len(extra) != len(xs) || len(controls) != 6 {
		return r, fmt.Errorf("frozen input shape mismatch")
	}
	var counts [3]int
	seen := make([]bool, len(xs))
	for i, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] || x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > min(3009, 2*x.BaselineRank-1) {
			return r, fmt.Errorf("invalid membership/ranks")
		}
		seen[x.Group] = true
		counts[repo]++
		for _, v := range x.Features {
			if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e6 {
				return r, fmt.Errorf("invalid feature")
			}
		}
		for _, v := range extra[i] {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return r, fmt.Errorf("invalid spread feature")
			}
		}
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("repository counts")
	}
	// Bind every control to the frozen fold and feature contract before writing.
	var frozen [3][2]CostClaimHead
	for fold, repo := range claimRepositories {
		for si, seed := range []uint64{1729, 2718} {
			n := 0
			for _, h := range controls {
				if h.Head.EvaluationRepository == repo && h.Head.Seed == seed {
					n++
					frozen[fold][si] = h
					if h.Head.Schema != "riido-spread-claim-v1" || h.Head.PlanSHA256 != SpreadClaimPlanSHA256 || len(h.Head.Weights) != 16 || h.Head.TrainingRepository != claimRepositories[(fold+1)%3] || h.Head.ValidationRepository != claimRepositories[(fold+2)%3] {
						return r, fmt.Errorf("control contract")
					}
				}
			}
			if n != 1 {
				return r, fmt.Errorf("control membership")
			}
		}
	}
	if e := os.Mkdir(out, 0700); e != nil {
		return r, e
	}
	r.Baseline = measurePageClaims(xs, nil, "baseline")
	r.AlwaysHelper = measurePageClaims(xs, nil, "always_interleave")
	policies := [5]string{"gap_heuristic", "spread_1729", "spread_2718", "shallow_selected", "gold_page_oracle"}
	var combined [5][]ClaimExample
	var ordered []ClaimExample
	for fold, repo := range claimRepositories {
		trainRepo, valRepo := claimRepositories[(fold+1)%3], claimRepositories[(fold+2)%3]
		tx, ty := treeRows(xs, extra, trainRepo)
		vx, _ := treeRows(xs, extra, valRepo)
		ex, _ := treeRows(xs, extra, repo)
		var validation, test []ClaimExample
		var testExtra [][searchclaim.SpreadDimension]float64
		for i, x := range xs {
			if x.Repository == valRepo {
				validation = append(validation, x)
			}
			if x.Repository == repo {
				test = append(test, x)
				testExtra = append(testExtra, extra[i])
			}
		}
		var cand [8]ShallowCandidate
		var models [8]claimtree.Model
		ci := 0
		for _, depth := range []int{1, 2, 3, 4} {
			for _, minLeaf := range []int{16, 64} {
				m, e := claimtree.Fit(tx, ty, depth, minLeaf)
				if e != nil {
					return r, e
				}
				threshold, e := CalibratePageBudget(validation, treeScores(&m, vx), 90)
				if e != nil {
					return r, e
				}
				h := ShallowHead{"riido-shallow-claim-v1", plan, trainRepo, valRepo, repo, depth, minLeaf, m, threshold}
				b, e := json.MarshalIndent(h, "", "  ")
				if e != nil {
					return r, e
				}
				b = append(b, '\n')
				file := fmt.Sprintf("fold%d-depth%d-leaf%d.tree.json", fold, depth, minLeaf)
				if e = os.WriteFile(filepath.Join(out, file), b, 0600); e != nil {
					return r, e
				}
				cand[ci] = ShallowCandidate{file, paireval.Hash(b), repo, len(b), depth, minLeaf, m.Nodes, threshold}
				models[ci] = m
				r.Candidates = append(r.Candidates, cand[ci])
				ci++
			}
		}
		best := 0
		for i := 1; i < len(cand); i++ {
			if betterTree(cand[i], cand[best]) {
				best = i
			}
		}
		gt, e := CalibratePageBudget(validation, costScores(validation, nil), 90)
		if e != nil {
			return r, e
		}
		var selected [5][]ClaimExample
		selected[0] = applyPageBudget(test, costScores(test, nil), gt)
		for si := 0; si < 2; si++ {
			h := frozen[fold][si]
			selected[1+si] = applyPageBudget(test, costScoresExtra(test, testExtra, h.Head.Weights), h.Threshold)
		}
		selected[3] = applyPageBudget(test, treeScores(&models[best], ex), cand[best].Threshold)
		oracleScores := make([]float64, len(test))
		for i, x := range test {
			oracleScores[i] = float64((x.BaselineRank+19)/20 - (x.InterleavedRank+19)/20)
		}
		selected[4] = applyPageBudget(test, oracleScores, BudgetThreshold{MinimumScore: 1})
		ordered = append(ordered, test...)
		for slot, sel := range selected {
			file := ""
			if slot == 3 {
				file = cand[best].File
			}
			r.Folds = append(r.Folds, ShallowFold{policies[slot], repo, file, budgetMetrics(sel), effect(test, sel)})
			combined[slot] = append(combined[slot], sel...)
		}
	}
	gap := budgetMetrics(combined[0])
	for slot, sel := range combined {
		m := budgetMetrics(sel)
		pass := m.Pages <= r.AlwaysHelper.Pages && m.Quality.AuxiliaryCalls*10 <= r.Questions*9 && m.Quality.Recall1 >= r.Baseline.Quality.Recall1 && m.Quality.Recall10 >= r.Baseline.Quality.Recall10
		if slot > 0 {
			pass = pass && m.Pages <= gap.Pages && m.Quality.AuxiliaryCalls <= gap.Quality.AuxiliaryCalls && (m.Pages < gap.Pages || m.Quality.AuxiliaryCalls < gap.Quality.AuxiliaryCalls)
		}
		// The oracle is a label-reading bound, never an eligible learned policy.
		if slot == 4 {
			pass = false
		}
		r.Aggregates = append(r.Aggregates, ShallowAggregate{policies[slot], m, effect(ordered, sel), pass})
		if slot == 3 {
			r.PrimaryPassesExploratoryGate = pass
		}
	}
	return r, nil
}
