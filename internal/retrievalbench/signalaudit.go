package retrievalbench

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const SignalAuditPlanSHA256 = "c7ac47857642df4e5e50b990f8b6a93f1c6f1d9d2a8bafa3adde853fac7116ca"

type SignalMetric struct {
	Cases, Positives                  int
	Prevalence, AUC, AveragePrecision float64
}
type SignalFold struct {
	Target, Feature, TrainingRepository, ValidationRepository, EvaluationRepository string
	Sign                                                                            int
	Training, Validation, Evaluation                                                SignalMetric
}
type SignalAggregate struct {
	Target, Feature                               string
	Cases, Positives                              int
	MeanAUC, MeanAveragePrecision, MeanPrevalence float64
}
type SignalAuditReport struct {
	Schema, PlanSHA256 string
	Questions          int
	Folds              []SignalFold
	Aggregates         []SignalAggregate
	ProductionReady    bool
}

// signalMetric uses score groups for AP, so catalog ordering cannot split ties.
func signalMetric(ps []paireval.Prediction) (SignalMetric, error) {
	m := SignalMetric{Cases: len(ps)}
	for _, p := range ps {
		if !p.Eligible || (p.Label != 0 && p.Label != 1) || math.IsNaN(p.Score) || math.IsInf(p.Score, 0) {
			return m, fmt.Errorf("invalid signal observation")
		}
		m.Positives += p.Label
	}
	if m.Positives == 0 || m.Positives == m.Cases {
		return m, fmt.Errorf("signal metric requires both classes")
	}
	m.Prevalence = float64(m.Positives) / float64(m.Cases)
	m.AUC = *paireval.AUC(ps)
	xs := slices.Clone(ps)
	slices.SortFunc(xs, func(a, b paireval.Prediction) int { return cmp.Compare(b.Score, a.Score) })
	positive := 0
	for start := 0; start < len(xs); {
		end := start
		hits := 0
		for end < len(xs) && xs[end].Score == xs[start].Score {
			hits += xs[end].Label
			end++
		}
		positive += hits
		m.AveragePrecision += float64(hits) / float64(m.Positives) * float64(positive) / float64(end)
		start = end
	}
	return m, nil
}

// orientSignal reads training observations only; holdout labels never choose sign.
func orientSignal(train []paireval.Prediction) (int, error) {
	m, e := signalMetric(train)
	if e != nil {
		return 0, e
	}
	if m.AUC < .5 {
		return -1, nil
	}
	return 1, nil
}
func signalRows(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, repo, target string, feature, sign int) []paireval.Prediction {
	var ps []paireval.Prediction
	for i, x := range xs {
		if x.Repository != repo {
			continue
		}
		d := (x.BaselineRank+19)/20 - (x.InterleavedRank+19)/20
		if target == "harm_vs_benefit" && d == 0 {
			continue
		}
		label := 0
		if target == "benefit_vs_rest" {
			if d > 0 {
				label = 1
			}
		} else {
			if d < 0 {
				label = 1
			}
		}
		value := 0.0
		if feature < 12 {
			value = x.Features[feature]
		} else {
			value = extra[i][feature-12]
		}
		ps = append(ps, paireval.Prediction{Label: label, Score: float64(sign) * value, Eligible: true})
	}
	return ps
}
func AuditClaimSignals(xs []ClaimExample, extra [][searchclaim.SpreadDimension]float64, plan string) (SignalAuditReport, error) {
	r := SignalAuditReport{Schema: "riido-claim-signal-audit-v1", PlanSHA256: plan, Questions: len(xs)}
	if plan != SignalAuditPlanSHA256 || len(xs) != 2948 || len(extra) != len(xs) {
		return r, fmt.Errorf("frozen shape mismatch")
	}
	var counts [3]int
	seen := make([]bool, len(xs))
	for i, x := range xs {
		repo := slices.Index(claimRepositories, x.Repository)
		if repo < 0 || x.Group < 0 || x.Group >= len(xs) || seen[x.Group] || x.BaselineRank < 1 || x.BaselineRank > 3009 || x.InterleavedRank < 1 || x.InterleavedRank > min(3009, 2*x.BaselineRank-1) {
			return r, fmt.Errorf("invalid frozen membership")
		}
		seen[x.Group] = true
		counts[repo]++
		for _, v := range x.Features {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return r, fmt.Errorf("invalid feature")
			}
		}
		for _, v := range extra[i] {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return r, fmt.Errorf("invalid spread")
			}
		}
	}
	if counts != [3]int{806, 1152, 990} {
		return r, fmt.Errorf("repository counts mismatch")
	}
	var names [16]string
	copy(names[:], searchclaim.Names[:])
	copy(names[12:], searchclaim.SpreadNames[:])
	for _, target := range []string{"harm_vs_rest", "benefit_vs_rest", "harm_vs_benefit"} {
		for f, name := range names {
			a := SignalAggregate{Target: target, Feature: name}
			for fold, repo := range claimRepositories {
				trainRepo, valRepo := claimRepositories[(fold+1)%3], claimRepositories[(fold+2)%3]
				sign, e := orientSignal(signalRows(xs, extra, trainRepo, target, f, 1))
				if e != nil {
					return r, e
				}
				train, e := signalMetric(signalRows(xs, extra, trainRepo, target, f, sign))
				if e != nil {
					return r, e
				}
				val, e := signalMetric(signalRows(xs, extra, valRepo, target, f, sign))
				if e != nil {
					return r, e
				}
				test, e := signalMetric(signalRows(xs, extra, repo, target, f, sign))
				if e != nil {
					return r, e
				}
				r.Folds = append(r.Folds, SignalFold{target, name, trainRepo, valRepo, repo, sign, train, val, test})
				a.Cases += test.Cases
				a.Positives += test.Positives
				a.MeanAUC += test.AUC / 3
				a.MeanAveragePrecision += test.AveragePrecision / 3
				a.MeanPrevalence += test.Prevalence / 3
			}
			r.Aggregates = append(r.Aggregates, a)
		}
	}
	return r, nil
}
