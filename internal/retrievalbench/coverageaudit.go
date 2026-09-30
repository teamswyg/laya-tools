package retrievalbench

import (
	"fmt"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"math"
)

const CoverageAuditPlanSHA256 = "5ec0cdb6c3d5affb086c22770bab79ff1adca0aed496b77a08193f21e3ff50e2"

// AuditCoverageSignals preserves audit23's complete control analysis and adds
// every coverage feature without choosing a held-out winner or operating point.
func AuditCoverageSignals(xs []ClaimExample, spread [][searchclaim.SpreadDimension]float64, coverage [][searchclaim.CoverageDimension]float64, plan string) (SignalAuditReport, error) {
	if plan != CoverageAuditPlanSHA256 || len(coverage) != len(xs) {
		return SignalAuditReport{}, fmt.Errorf("coverage plan/shape mismatch")
	}
	for _, row := range coverage {
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return SignalAuditReport{}, fmt.Errorf("invalid coverage feature")
			}
		}
	}
	r, e := AuditClaimSignals(xs, spread, SignalAuditPlanSHA256)
	if e != nil {
		return r, e
	}
	r.Schema = "riido-coverage-signal-audit-v1"
	r.PlanSHA256 = plan
	for _, target := range []string{"harm_vs_rest", "benefit_vs_rest", "harm_vs_benefit"} {
		for f, name := range searchclaim.CoverageNames {
			a := SignalAggregate{Target: target, Feature: name}
			for fold, repo := range claimRepositories {
				trainRepo, valRepo := claimRepositories[(fold+1)%3], claimRepositories[(fold+2)%3]
				// Both extra arrays have four columns. Selecting 12+f reads only coverage.
				sign, e := orientSignal(signalRows(xs, coverage, trainRepo, target, 12+f, 1))
				if e != nil {
					return r, e
				}
				train, e := signalMetric(signalRows(xs, coverage, trainRepo, target, 12+f, sign))
				if e != nil {
					return r, e
				}
				val, e := signalMetric(signalRows(xs, coverage, valRepo, target, 12+f, sign))
				if e != nil {
					return r, e
				}
				test, e := signalMetric(signalRows(xs, coverage, repo, target, 12+f, sign))
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
