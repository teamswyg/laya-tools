package retrievalbench

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const PageBudgetPlanSHA256 = "6ba948249bde6760295a7d4156a3c00f6ca837334f3886644625b3333fed3871"
const PageBudgetManifestSHA256 = "28ac81989edbe98f0551f6e470f58cdce7114019dddda803974a801ecd7bc2b5"

type BudgetThreshold struct {
	Disabled                                                             bool
	MinimumScore                                                         float64
	BudgetPercent, ValidationQuestions, ValidationCalls, ValidationPages int
}

// CalibratePageBudget accepts validation outcomes only. A tied score is an
// indivisible group: an online threshold cannot rank identical scores apart.
func CalibratePageBudget(validation []ClaimExample, scores []float64, budget int) (BudgetThreshold, error) {
	best := BudgetThreshold{Disabled: true, BudgetPercent: budget, ValidationQuestions: len(validation)}
	if len(validation) == 0 || len(validation) != len(scores) || budget < 0 || budget > 100 {
		return best, fmt.Errorf("invalid calibration inputs")
	}
	type entry struct {
		score float64
		delta int
	}
	entries := make([]entry, len(validation))
	for i, x := range validation {
		if x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > 3009 || math.IsNaN(scores[i]) || math.IsInf(scores[i], 0) {
			return best, fmt.Errorf("invalid calibration observation")
		}
		base, helper := (x.BaselineRank+19)/20, (x.InterleavedRank+19)/20
		best.ValidationPages += base
		entries[i] = entry{scores[i], helper - base}
	}
	slices.SortFunc(entries, func(a, b entry) int { return cmp.Compare(b.score, a.score) })
	pages := best.ValidationPages
	cap := len(entries) * budget / 100
	for start := 0; start < len(entries); {
		end := start
		for end < len(entries) && entries[end].score == entries[start].score {
			pages += entries[end].delta
			end++
		}
		if end > cap {
			break
		}
		if pages < best.ValidationPages || (pages == best.ValidationPages && end < best.ValidationCalls) {
			best.Disabled = false
			best.MinimumScore = entries[start].score
			best.ValidationPages = pages
			best.ValidationCalls = end
		}
		start = end
	}
	return best, nil
}

func applyPageBudget(xs []ClaimExample, scores []float64, t BudgetThreshold) []ClaimExample {
	selected := slices.Clone(xs)
	for i := range selected {
		selected[i].AuxiliaryRank = 0 // Count actual calls independently of rank ties.
		if !t.Disabled && scores[i] >= t.MinimumScore {
			selected[i].AuxiliaryRank = 1
		} else {
			selected[i].InterleavedRank = selected[i].BaselineRank
		}
	}
	return selected
}
func budgetMetrics(xs []ClaimExample) PageClaimMetrics {
	m := measurePageClaims(xs, nil, "always_interleave")
	m.Quality.AuxiliaryCalls = 0
	for _, x := range xs {
		m.Quality.AuxiliaryCalls += x.AuxiliaryRank
	}
	return m
}

type BudgetHead struct {
	File, SHA256 string
	Head         ClaimHead
}
type BudgetFold struct {
	Mode                                       string
	Seed                                       uint64
	EvaluationRepository, HeadFile, HeadSHA256 string
	Threshold                                  BudgetThreshold
	Evaluation                                 PageClaimMetrics
}
type BudgetAggregate struct {
	Mode                  string
	Seed                  uint64
	BudgetPercent         int
	Metrics               PageClaimMetrics
	PassesExploratoryGate bool
}
type BudgetReport struct {
	Schema, PlanSHA256, SourceManifestSHA256                                        string
	Questions                                                                       int
	Baseline, AlwaysHelper                                                          PageClaimMetrics
	Folds                                                                           []BudgetFold
	Aggregates                                                                      []BudgetAggregate
	PrimaryPassesExploratoryGate, ReplicationPassesExploratoryGate, ProductionReady bool
}

func ProbePageBudgets(xs []ClaimExample, heads []BudgetHead) (BudgetReport, error) {
	r := BudgetReport{Schema: "riido-page-budget-report-v1", PlanSHA256: PageBudgetPlanSHA256, SourceManifestSHA256: PageBudgetManifestSHA256, Questions: len(xs)}
	if len(xs) != 2948 || len(heads) != 24 {
		return r, fmt.Errorf("unexpected frozen input size")
	}
	counts := [3]int{}
	seen := make([]bool, len(xs))
	for _, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] {
			return r, fmt.Errorf("invalid question membership")
		}
		if x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > min(3009, 2*x.BaselineRank-1) {
			return r, fmt.Errorf("invalid baseline-first ranks")
		}
		for _, f := range x.Features {
			if math.IsNaN(f) || math.IsInf(f, 0) {
				return r, fmt.Errorf("invalid feature")
			}
		}
		counts[repo]++
		seen[x.Group] = true
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("unexpected repository counts")
	}
	r.Baseline = measurePageClaims(xs, nil, "baseline")
	r.AlwaysHelper = measurePageClaims(xs, nil, "always_interleave")
	modes := []string{"gap_heuristic", "fp32", "int8", "ternary_ptq", "ternary_ste"}
	budgets := []int{25, 50, 75, 90}
	seeds := []uint64{1729, 2718}
	// Slot 0 is the heuristic; slots 1..8 are mode/seed combinations.
	var combined [9][4][]ClaimExample
	for fold, testRepo := range claimRepositories {
		valRepo := claimRepositories[(fold+2)%3]
		var validation, test []ClaimExample
		for _, x := range xs {
			if x.Repository == valRepo {
				validation = append(validation, x)
			}
			if x.Repository == testRepo {
				test = append(test, x)
			}
		}
		for mi, mode := range modes {
			for si, seed := range seeds {
				if mi == 0 && si > 0 {
					continue
				}
				slot := 0
				var source BudgetHead
				if mi > 0 {
					slot = 1 + (mi-1)*2 + si
					matches := 0
					for _, h := range heads {
						if h.Head.Mode == mode && h.Head.Seed == seed && h.Head.EvaluationRepository == testRepo {
							source = h
							matches++
						}
					}
					if matches != 1 || source.Head.Schema != "riido-page-claim-v1" || source.Head.PlanSHA256 != PageClaimPlanSHA256 || source.Head.ValidationRepository != valRepo || source.Head.TrainingRepository != claimRepositories[(fold+1)%3] || len(source.Head.Weights) != searchclaim.Dimension {
						return r, fmt.Errorf("head fold contract mismatch")
					}
					for _, w := range source.Head.Weights {
						if math.IsNaN(w) || math.IsInf(w, 0) {
							return r, fmt.Errorf("invalid weight")
						}
					}
				} else {
					seed = 0
				}
				score := func(rows []ClaimExample) []float64 {
					values := make([]float64, len(rows))
					for i, x := range rows {
						if mi == 0 {
							values[i] = -x.Features[9]
						} else {
							values[i] = searchclaim.Score(source.Head.Weights, x.Features)
						}
					}
					return values
				}
				vs, ts := score(validation), score(test)
				for bi, budget := range budgets {
					threshold, err := CalibratePageBudget(validation, vs, budget)
					if err != nil {
						return r, err
					}
					selected := applyPageBudget(test, ts, threshold)
					combined[slot][bi] = append(combined[slot][bi], selected...)
					r.Folds = append(r.Folds, BudgetFold{mode, seed, testRepo, source.File, source.SHA256, threshold, budgetMetrics(selected)})
				}
			}
		}
	}
	for mi, mode := range modes {
		for si, seed := range seeds {
			if mi == 0 && si > 0 {
				continue
			}
			slot := 0
			if mi > 0 {
				slot = 1 + (mi-1)*2 + si
			} else {
				seed = 0
			}
			for bi, budget := range budgets {
				m := budgetMetrics(combined[slot][bi])
				control := budgetMetrics(combined[0][bi])
				pass := m.Pages <= r.AlwaysHelper.Pages && m.Quality.AuxiliaryCalls*100 <= r.Questions*budget && m.Quality.Recall1 >= r.Baseline.Quality.Recall1 && m.Quality.Recall10 >= r.Baseline.Quality.Recall10
				if mi > 0 {
					pass = pass && m.Pages <= control.Pages && m.Quality.AuxiliaryCalls <= control.Quality.AuxiliaryCalls && (m.Pages < control.Pages || m.Quality.AuxiliaryCalls < control.Quality.AuxiliaryCalls)
				}
				r.Aggregates = append(r.Aggregates, BudgetAggregate{mode, seed, budget, m, pass})
				if mode == "fp32" && budget == 90 {
					if si == 0 {
						r.PrimaryPassesExploratoryGate = pass
					} else {
						r.ReplicationPassesExploratoryGate = pass
					}
				}
			}
		}
	}
	return r, nil
}
