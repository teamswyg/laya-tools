package retrievalbench

import (
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"math"
	"reflect"
	"testing"
)

func TestCoverageAuditPreservesControls(t *testing.T) {
	var xs []ClaimExample
	for repo, n := range []int{806, 1152, 990} {
		for i := 0; i < n; i++ {
			b, h := 1, 1
			switch i % 3 {
			case 0:
				b, h = 21, 41
			case 1:
				b, h = 41, 1
			}
			x := ClaimExample{Repository: claimRepositories[repo], Group: len(xs), BaselineRank: b, InterleavedRank: h}
			x.Features[0] = 1
			xs = append(xs, x)
		}
	}
	spread := make([][searchclaim.SpreadDimension]float64, len(xs))
	cov := make([][searchclaim.CoverageDimension]float64, len(xs))
	old, e := AuditClaimSignals(xs, spread, SignalAuditPlanSHA256)
	if e != nil {
		t.Fatal(e)
	}
	got, e := AuditCoverageSignals(xs, spread, cov, CoverageAuditPlanSHA256)
	if e != nil {
		t.Fatal(e)
	}
	if len(got.Folds) != 180 || len(got.Aggregates) != 60 || !reflect.DeepEqual(old.Folds, got.Folds[:144]) || !reflect.DeepEqual(old.Aggregates, got.Aggregates[:48]) {
		t.Fatal("changed control audit")
	}
	cov[0][0] = math.NaN()
	if _, e := AuditCoverageSignals(xs, spread, cov, CoverageAuditPlanSHA256); e == nil {
		t.Fatal("accepted nan")
	}
}
