package retrievalbench

import (
	"fmt"
	"math"
	"slices"

	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const CallChangePlanSHA256 = "cf3f55bd68b3514a1cfdb99f1f1f0ef16ae02ddf22ba7ae330568efed007b108"
const Coverage25ManifestSHA256 = "09abdb32f44119be1ae54f7d38be4fa35d6a8dbc0c88208a3ef6bb044f3b60cf"

type FrozenClaimPolicy struct {
	Report CostClaimReport
	Heads  []CostClaimHead
}
type CallTransition struct {
	Name                                                                           string
	Cases, OldCalls, NewCalls, OldPages, NewPages                                  int
	IncreasedCases, IncreasedPages, DecreasedCases, DecreasedPages, UnchangedCases int
	Top10Gains, Top10Losses                                                        int
}
type LossBucket struct{ Minimum, Maximum, Cases, Pages int } // maximum0 means unbounded

type CallChangeSummary struct {
	Repository                                                                     string
	Seed                                                                           uint64
	Old, New                                                                       PageClaimMetrics
	OldEffect, NewEffect                                                           ClaimEffect
	Transitions                                                                    [4]CallTransition
	IncreasedCases, IncreasedPages, DecreasedCases, DecreasedPages, UnchangedCases int
	LargestPositiveSums                                                            [3]int // fixed k=1,3,10
	LargestPositiveShares                                                          [3]float64
	Buckets                                                                        [4]LossBucket
}
type CallChangeReport struct {
	Schema, PlanSHA256, PreviousManifestSHA256, CurrentManifestSHA256 string
	Questions                                                         int
	Folds, Aggregates                                                 []CallChangeSummary
	ProductionReady                                                   bool
}

func summarizeCallChanges(xs, old, new []ClaimExample, repo string, seed uint64) (CallChangeSummary, error) {
	r := CallChangeSummary{Repository: repo, Seed: seed, Buckets: [4]LossBucket{{Minimum: 1, Maximum: 1}, {Minimum: 2, Maximum: 5}, {Minimum: 6, Maximum: 20}, {Minimum: 21}}}
	if len(xs) == 0 || len(xs) != len(old) || len(xs) != len(new) {
		return r, fmt.Errorf("change row shape mismatch")
	}
	for i, name := range []string{"neither_calls", "new_only_calls", "old_only_calls", "both_call"} {
		r.Transitions[i].Name = name
	}
	var losses []int
	for i, x := range xs {
		if x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > min(3009, 2*x.BaselineRank-1) {
			return r, fmt.Errorf("invalid source ranks")
		}
		for _, s := range []ClaimExample{old[i], new[i]} {
			expected := x.BaselineRank
			if s.AuxiliaryRank == 1 {
				expected = x.InterleavedRank
			}
			if s.AuxiliaryRank < 0 || s.AuxiliaryRank > 1 || s.InterleavedRank != expected || s.BaselineRank != x.BaselineRank {
				return r, fmt.Errorf("invalid applied policy")
			}
		}
		op, np := (old[i].InterleavedRank+19)/20, (new[i].InterleavedRank+19)/20
		t := &r.Transitions[old[i].AuxiliaryRank*2+new[i].AuxiliaryRank]
		t.Cases++
		t.OldCalls += old[i].AuxiliaryRank
		t.NewCalls += new[i].AuxiliaryRank
		t.OldPages += op
		t.NewPages += np
		if old[i].InterleavedRank > 10 && new[i].InterleavedRank <= 10 {
			t.Top10Gains++
		}
		if old[i].InterleavedRank <= 10 && new[i].InterleavedRank > 10 {
			t.Top10Losses++
		}
		d := np - op
		switch {
		case d > 0:
			r.IncreasedCases++
			r.IncreasedPages += d
			t.IncreasedCases++
			t.IncreasedPages += d
			losses = append(losses, d)
			for bi := range r.Buckets {
				b := &r.Buckets[bi]
				if d >= b.Minimum && (b.Maximum == 0 || d <= b.Maximum) {
					b.Cases++
					b.Pages += d
					break
				}
			}
		case d < 0:
			r.DecreasedCases++
			r.DecreasedPages -= d
			t.DecreasedCases++
			t.DecreasedPages -= d
		default:
			r.UnchangedCases++
			t.UnchangedCases++
		}
	}
	r.Old = budgetMetrics(old)
	r.New = budgetMetrics(new)
	r.OldEffect = effect(xs, old)
	r.NewEffect = effect(xs, new)
	slices.Sort(losses)
	slices.Reverse(losses)
	for i, k := range []int{1, 3, 10} {
		for _, v := range losses[:min(k, len(losses))] {
			r.LargestPositiveSums[i] += v
		}
		if r.IncreasedPages > 0 {
			r.LargestPositiveShares[i] = float64(r.LargestPositiveSums[i]) / float64(r.IncreasedPages)
		}
	}
	if r.IncreasedPages-r.DecreasedPages != r.New.Pages-r.Old.Pages || r.IncreasedCases+r.DecreasedCases+r.UnchangedCases != len(xs) {
		return r, fmt.Errorf("change conservation failed")
	}
	return r, nil
}
func frozenHead(p FrozenClaimPolicy, repo string, seed uint64, dimension int, plan, schema string) (CostClaimHead, error) {
	var found CostClaimHead
	n := 0
	fold := slices.Index(claimRepositories, repo)
	for _, h := range p.Heads {
		if h.Head.EvaluationRepository == repo && h.Head.Seed == seed {
			found = h
			n++
		}
	}
	if n != 1 || fold < 0 {
		return found, fmt.Errorf("frozen head membership")
	}
	h := found.Head
	if h.Schema != schema || h.PlanSHA256 != plan || h.TrainingRepository != claimRepositories[(fold+1)%3] || h.ValidationRepository != claimRepositories[(fold+2)%3] || len(h.Weights) != dimension {
		return found, fmt.Errorf("frozen head contract")
	}
	for _, w := range h.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) {
			return found, fmt.Errorf("invalid weight")
		}
	}
	if math.IsNaN(found.Threshold.MinimumScore) || math.IsInf(found.Threshold.MinimumScore, 0) {
		return found, fmt.Errorf("invalid threshold")
	}
	return found, nil
}
func publishedFold(p FrozenClaimPolicy, repo string, seed uint64, m PageClaimMetrics) bool {
	n := 0
	for _, f := range p.Report.Folds {
		if f.Policy == "cost_selected" && f.EvaluationRepository == repo && f.Seed == seed {
			if f.Metrics != m {
				return false
			}
			n++
		}
	}
	return n == 1
}
func publishedAggregate(p FrozenClaimPolicy, seed uint64, m PageClaimMetrics) bool {
	n := 0
	for _, a := range p.Report.Aggregates {
		if a.Policy == "cost_selected" && a.Seed == seed {
			if a.Metrics != m {
				return false
			}
			n++
		}
	}
	return n == 1
}

// AuditCallChanges only replays verified, frozen policies. Archive verification
// and exact manifest pins are enforced by the command before this function.
func AuditCallChanges(xs []ClaimExample, spread [][searchclaim.SpreadDimension]float64, coverage [][searchclaim.CoverageDimension]float64, previous, current FrozenClaimPolicy, plan string) (CallChangeReport, error) {
	r := CallChangeReport{Schema: "riido-call-change-audit-v1", PlanSHA256: plan, PreviousManifestSHA256: ShallowControlManifestSHA256, CurrentManifestSHA256: Coverage25ManifestSHA256, Questions: len(xs)}
	if plan != CallChangePlanSHA256 || len(xs) != 2948 || len(spread) != len(xs) || len(coverage) != len(xs) || len(previous.Heads) != 6 || len(current.Heads) != 6 {
		return r, fmt.Errorf("frozen input shape")
	}
	var counts [3]int
	seen := make([]bool, len(xs))
	for i, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] {
			return r, fmt.Errorf("invalid membership")
		}
		seen[x.Group] = true
		counts[repo]++
		for _, v := range x.Features {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return r, fmt.Errorf("invalid feature")
			}
		}
		for _, row := range [][4]float64{spread[i], coverage[i]} {
			for _, v := range row {
				if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
					return r, fmt.Errorf("invalid extra feature")
				}
			}
		}
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("repository counts")
	}
	for _, seed := range []uint64{1729, 2718} {
		var all, oldAll, newAll []ClaimExample
		for _, repo := range claimRepositories {
			oh, e := frozenHead(previous, repo, seed, 16, SpreadClaimPlanSHA256, "riido-spread-claim-v1")
			if e != nil {
				return r, e
			}
			nh, e := frozenHead(current, repo, seed, 20, CoverageClaimPlanSHA256, "riido-coverage-claim-v1")
			if e != nil {
				return r, e
			}
			var test []ClaimExample
			var ext [][8]float64
			for i, x := range xs {
				if x.Repository == repo {
					test = append(test, x)
					var v [8]float64
					copy(v[:4], spread[i][:])
					copy(v[4:], coverage[i][:])
					ext = append(ext, v)
				}
			}
			old := applyPageBudget(test, costScoresColumns(test, ext, 4, oh.Head.Weights), oh.Threshold)
			new := applyPageBudget(test, costScoresColumns(test, ext, 8, nh.Head.Weights), nh.Threshold)
			summary, e := summarizeCallChanges(test, old, new, repo, seed)
			if e != nil {
				return r, e
			}
			if !publishedFold(previous, repo, seed, summary.Old) || !publishedFold(current, repo, seed, summary.New) {
				return r, fmt.Errorf("published fold replay mismatch")
			}
			r.Folds = append(r.Folds, summary)
			all = append(all, test...)
			oldAll = append(oldAll, old...)
			newAll = append(newAll, new...)
		}
		summary, e := summarizeCallChanges(all, oldAll, newAll, "all", seed)
		if e != nil {
			return r, e
		}
		if !publishedAggregate(previous, seed, summary.Old) || !publishedAggregate(current, seed, summary.New) {
			return r, fmt.Errorf("published aggregate replay mismatch")
		}
		r.Aggregates = append(r.Aggregates, summary)
	}
	return r, nil
}
