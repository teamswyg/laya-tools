package retrievalbench

import (
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"math"
	"reflect"
	"testing"
)

func TestSignalMetricsTiesAndOrientation(t *testing.T) {
	ps := []paireval.Prediction{{Label: 1, Score: 2, Eligible: true}, {Label: 0, Score: 2, Eligible: true}, {Label: 1, Score: 1, Eligible: true}, {Label: 0, Score: 0, Eligible: true}}
	m, e := signalMetric(ps)
	if e != nil {
		t.Fatal(e)
	}
	if m.AUC != .625 || math.Abs(m.AveragePrecision-7.0/12) > 1e-12 {
		t.Fatalf("tie accounting %+v", m)
	}
	ps[0], ps[1] = ps[1], ps[0]
	n, e := signalMetric(ps)
	if e != nil || m != n {
		t.Fatal("tie order changed metrics")
	}
	for i := range ps {
		ps[i].Score = 1
	}
	m, e = signalMetric(ps)
	if e != nil || m.AUC != .5 || m.AveragePrecision != .5 {
		t.Fatal("constant feature")
	}
	s, e := orientSignal(ps)
	if e != nil || s != 1 {
		t.Fatal("constant sign")
	}
	ps = []paireval.Prediction{{Label: 1, Score: 0, Eligible: true}, {Label: 0, Score: 1, Eligible: true}}
	s, e = orientSignal(ps)
	if e != nil || s != -1 {
		t.Fatal("inverted sign")
	}
	ps[0].Score = math.NaN()
	if _, e := signalMetric(ps); e == nil {
		t.Fatal("accepted nan")
	}
	if _, e := signalMetric([]paireval.Prediction{{Label: 1, Eligible: true}}); e == nil {
		t.Fatal("accepted single class")
	}
}
func TestSignalSubgroupAndHeldoutIsolation(t *testing.T) {
	xs := []ClaimExample{{Repository: "train", BaselineRank: 21, InterleavedRank: 41}, {Repository: "train", BaselineRank: 41, InterleavedRank: 1}, {Repository: "train", BaselineRank: 1, InterleavedRank: 1}, {Repository: "test", BaselineRank: 21, InterleavedRank: 1}}
	xs[0].Features[1] = .9
	xs[1].Features[1] = .1
	extra := make([][searchclaim.SpreadDimension]float64, len(xs))
	train := signalRows(xs, extra, "train", "harm_vs_benefit", 1, 1)
	if len(train) != 2 || train[0].Label != 1 || train[1].Label != 0 {
		t.Fatal("wrong subgroup")
	}
	sign, e := orientSignal(train)
	if e != nil || sign != 1 {
		t.Fatal(e)
	}
	xs[3].Features[1] = 999
	xs[3].InterleavedRank = 40
	again := signalRows(xs, extra, "train", "harm_vs_benefit", 1, 1)
	if !reflect.DeepEqual(train, again) {
		t.Fatal("heldout mutation changed training")
	}
	if len(signalRows(xs, extra, "train", "harm_vs_rest", 1, 1)) != 3 {
		t.Fatal("rest omitted ties")
	}
}
