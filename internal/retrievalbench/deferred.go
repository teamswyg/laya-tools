package retrievalbench

import (
	"fmt"
	"slices"
)

const DeferredPlanSHA256 = "0dbd0fe7e270e857f5b36b89d064b7340cbcde8e52e7a892f6122bdb387e7c5b"

type DeferredSummary struct {
	Policy, Repository                                                               string
	Seed                                                                             uint64
	Static, Deferred                                                                 PageClaimMetrics
	InitialCalls, ContinuationCalls                                                  int
	ImprovedCases, SavedPages, WorsenedCases, AddedPages, TiedCases, LargestIncrease int
	PassesExploratoryGate                                                            bool
}
type DeferredReport struct {
	Schema, PlanSHA256, PreviousManifestSHA256, CurrentManifestSHA256               string
	Questions, PageSize                                                             int
	Baseline, AlwaysHelper                                                          PageClaimMetrics
	Folds, Aggregates                                                               []DeferredSummary
	PrimaryPassesExploratoryGate, ReplicationPassesExploratoryGate, ProductionReady bool
}

// simulateContinuation uses gold ranks ONLY to emulate a perfect verifier's
// stopping behavior. Runtime must receive a client continuation, not these ranks.
func simulateContinuation(xs []ClaimExample, ranks []int, policy, repo string, seed uint64) (DeferredSummary, []ClaimExample, error) {
	r := DeferredSummary{Policy: policy, Repository: repo, Seed: seed}
	if len(xs) == 0 || len(xs) != len(ranks) {
		return r, nil, fmt.Errorf("deferred row shape")
	}
	selected := slices.Clone(xs)
	for i, x := range xs {
		if x.BaselineRank < 1 || x.BaselineRank > 3009 || ranks[i] < 1 || ranks[i] > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > 3009 || x.AuxiliaryRank < 0 || x.AuxiliaryRank > 1 {
			return r, nil, fmt.Errorf("invalid simulation ranks or call")
		}
		if (x.BaselineRank <= 20 && ranks[i] != x.BaselineRank) || (x.BaselineRank > 20 && ranks[i] <= 20) {
			return r, nil, fmt.Errorf("deferred prefix changed")
		}
		if x.AuxiliaryRank == 1 {
			r.InitialCalls++
		} else {
			if x.InterleavedRank != x.BaselineRank {
				return r, nil, fmt.Errorf("skipped policy changed baseline")
			}
			if x.BaselineRank > 20 {
				selected[i].InterleavedRank = ranks[i]
				selected[i].AuxiliaryRank = 1
				r.ContinuationCalls++
			}
		}
		delta := (selected[i].InterleavedRank+19)/20 - (x.InterleavedRank+19)/20
		if delta > 0 {
			r.WorsenedCases++
			r.AddedPages += delta
			r.LargestIncrease = max(r.LargestIncrease, delta)
		} else if delta < 0 {
			r.ImprovedCases++
			r.SavedPages -= delta
		} else {
			r.TiedCases++
		}
	}
	r.Static = budgetMetrics(xs)
	r.Deferred = budgetMetrics(selected)
	if r.InitialCalls+r.ContinuationCalls != r.Deferred.Quality.AuxiliaryCalls || r.AddedPages-r.SavedPages != r.Deferred.Pages-r.Static.Pages {
		return r, nil, fmt.Errorf("deferred conservation")
	}
	return r, selected, nil
}

// ProbeDeferredHelpers performs no fitting or threshold selection. Counterfactual
// helper rankings are prepared for every query; reported calls are simulated
// policy calls, NOT counts of actual RankInto executions during preparation.
func ProbeDeferredHelpers(xs []ClaimExample, spread, coverage [][4]float64, ranks []int, old, current FrozenClaimPolicy, plan string) (DeferredReport, error) {
	r := DeferredReport{Schema: "riido-deferred-helper-report-v1", PlanSHA256: plan, PreviousManifestSHA256: ShallowControlManifestSHA256, CurrentManifestSHA256: Coverage25ManifestSHA256, Questions: len(xs), PageSize: 20}
	if plan != DeferredPlanSHA256 || len(ranks) != len(xs) {
		return r, fmt.Errorf("deferred plan or shape")
	}
	// Also validates memberships, finite features, archive head contracts and exact
	// reproduction of every published static16/static20 fold and aggregate.
	if _, err := AuditCallChanges(xs, spread, coverage, old, current, CallChangePlanSHA256); err != nil {
		return r, err
	}
	for i, x := range xs {
		if ranks[i] > min(3009, x.InterleavedRank+20) {
			return r, fmt.Errorf("deferred ordering bound")
		}
	}
	r.Baseline = measurePageClaims(xs, nil, "baseline")
	r.AlwaysHelper = measurePageClaims(xs, nil, "always_interleave")
	type variant struct {
		name  string
		seed  uint64
		width int
	}
	variants := []variant{{"pure_deferred", 0, 0}, {"gap_deferred", 0, 0}, {"spread16_deferred", 1729, 4}, {"spread16_deferred", 2718, 4}, {"coverage20_deferred", 1729, 8}, {"coverage20_deferred", 2718, 8}}
	var gap PageClaimMetrics
	for _, v := range variants {
		var all []ClaimExample
		var allRanks []int
		for _, repo := range claimRepositories {
			var test []ClaimExample
			var extra [][8]float64
			var rr []int
			for i, x := range xs {
				if x.Repository == repo {
					test = append(test, x)
					rr = append(rr, ranks[i])
					var e [8]float64
					copy(e[:4], spread[i][:])
					copy(e[4:], coverage[i][:])
					extra = append(extra, e)
				}
			}
			var selected []ClaimExample
			if v.name == "pure_deferred" {
				selected = applyPageBudget(test, make([]float64, len(test)), BudgetThreshold{Disabled: true})
			} else if v.name == "gap_deferred" {
				var f CostFold
				matches := 0
				for _, candidate := range old.Report.Folds {
					if candidate.Policy == "gap_heuristic" && candidate.EvaluationRepository == repo && candidate.Seed == 0 {
						f = candidate
						matches++
					}
				}
				if matches != 1 {
					return r, fmt.Errorf("frozen gap fold membership")
				}
				selected = applyPageBudget(test, costScores(test, nil), f.Threshold)
				if budgetMetrics(selected) != f.Metrics {
					return r, fmt.Errorf("frozen gap fold replay")
				}
			} else {
				source, dim, pin, schema := old, 16, SpreadClaimPlanSHA256, "riido-spread-claim-v1"
				if v.width == 8 {
					source, dim, pin, schema = current, 20, CoverageClaimPlanSHA256, "riido-coverage-claim-v1"
				}
				head, err := frozenHead(source, repo, v.seed, dim, pin, schema)
				if err != nil {
					return r, err
				}
				selected = applyPageBudget(test, costScoresColumns(test, extra, v.width, head.Head.Weights), head.Threshold)
			}
			summary, _, err := simulateContinuation(selected, rr, v.name, repo, v.seed)
			if err != nil {
				return r, err
			}
			r.Folds = append(r.Folds, summary)
			all = append(all, selected...)
			allRanks = append(allRanks, rr...)
		}
		summary, _, err := simulateContinuation(all, allRanks, v.name, "all", v.seed)
		if err != nil {
			return r, err
		}
		if v.name == "gap_deferred" {
			gap = summary.Static
			matches := 0
			for _, a := range old.Report.Aggregates {
				if a.Policy == "gap_heuristic" && a.Seed == 0 {
					matches++
					if a.Metrics != gap {
						return r, fmt.Errorf("frozen gap aggregate replay")
					}
				}
			}
			if matches != 1 {
				return r, fmt.Errorf("gap aggregate membership")
			}
		}
		m := summary.Deferred
		pass := m.Pages <= r.AlwaysHelper.Pages && m.Quality.AuxiliaryCalls*10 <= len(xs)*9 && m.Quality.Recall1 >= r.Baseline.Quality.Recall1 && m.Quality.Recall10 >= r.Baseline.Quality.Recall10
		if v.width > 0 {
			pass = pass && m.Pages <= gap.Pages && m.Quality.AuxiliaryCalls <= gap.Quality.AuxiliaryCalls && (m.Pages < gap.Pages || m.Quality.AuxiliaryCalls < gap.Quality.AuxiliaryCalls)
		}
		summary.PassesExploratoryGate = pass
		r.Aggregates = append(r.Aggregates, summary)
		if v.name == "spread16_deferred" {
			if v.seed == 1729 {
				r.PrimaryPassesExploratoryGate = pass
			} else {
				r.ReplicationPassesExploratoryGate = pass
			}
		}
	}
	return r, nil
}
