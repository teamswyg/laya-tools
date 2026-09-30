package paireval

import (
	"math"
	"reflect"
	"testing"
)

func TestThresholdAndTiedAUC(t *testing.T) {
	ps := []Prediction{{Label: 0, Score: 1, Eligible: true}, {Label: 1, Score: 1, Eligible: true}, {Label: 1, Score: 2, Eligible: true}, {Label: 0, Score: 0, Eligible: true}}
	a := AUC(ps)
	if a == nil || math.Abs(*a-.875) > 1e-12 {
		t.Fatal(a)
	}
	if Threshold(ps) != 2 {
		t.Fatal("conservative tie threshold incorrect", Threshold(ps))
	}
	for i := range ps {
		ps[i].Positive = ps[i].Score >= 2
	}
	m := Measure(ps)
	if m.BalancedAccuracy != .75 || m.Cases != 4 {
		t.Fatal(m)
	}
}
func TestScopeDoesNotRemoveCases(t *testing.T) {
	ps := []Prediction{{Label: 1, Eligible: false, Positive: true}, {Label: 0, Eligible: true, Positive: false}}
	m := Measure(ps)
	if m.Cases != 2 || m.Eligible != 1 || m.BalancedAccuracy != 1 || m.EligibleAUC != nil {
		t.Fatal(m)
	}
}
func TestGroupSplitAndLabelIndependence(t *testing.T) {
	zero, one := 0, 1
	rs := []Row{{Code: "a  b", Query: "first", Label: &zero}, {Code: "a\nb", Query: "second", Label: &one}, {Code: "different", Query: " SECOND ", Label: &zero}}
	s := Partition(rs)
	for _, r := range s.Rows {
		if r.Group != s.Rows[0].Group {
			t.Fatal("connected clone split")
		}
	}
	rs[0].Label = &one
	if !reflect.DeepEqual(s, Partition(rs)) {
		t.Fatal("labels changed grouping")
	}
}
func TestPairedBootstrapIdentity(t *testing.T) {
	ps := []Prediction{{Group: 0, Label: 0, Positive: false}, {Group: 1, Label: 1, Positive: true}, {Group: 2, Label: 1, Positive: false}, {Group: 3, Label: 0, Positive: true}}
	ci := DeltaInterval(ps, ps)
	if ci.Low != 0 || ci.High != 0 || ci.Groups != 4 || ci.Replicates == 0 {
		t.Fatal(ci)
	}
}
func TestBM25DevelopmentOnly(t *testing.T) {
	rs := []Row{{Code: "alpha alpha beta"}, {Code: "gamma"}, {Code: "secret"}}
	s := Split{Rows: []Membership{{0, "development"}, {1, "development"}, {2, "final"}}}
	l := NewLexical(rs, s)
	if l.Score("alpha", "alpha") == 0 || l.Score("secret", "secret") != 0 {
		t.Fatal("development collection boundary failed")
	}
}
