package pairlearn

import (
	"math"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
)

func fixture(name string, group int) Dataset {
	return Dataset{Split: name, Offsets: []int{0, 1, 2, 3, 4}, Indices: []uint16{1, 1, 1, 1}, Values: []float64{-2, -1, 1, 2}, Labels: []float64{0, 0, 1, 1}, Groups: []int{group, group, group, group}}
}
func TestLearningAndReproducibility(t *testing.T) {
	a, b := fixture("development", 1), fixture("validation", 2)
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: .2, L2: .0001, Epochs: 20, Batch: 4}
	r, e := Fit(a, b, c)
	if e != nil {
		t.Fatal(e)
	}
	if r.ValidationNLL >= .3 {
		t.Fatalf("did not learn separable labels: %v", r.ValidationNLL)
	}
	r2, e := Fit(a, b, c)
	if e != nil || !reflect.DeepEqual(r, r2) {
		t.Fatal("non reproducible", e)
	}
	b.Groups[0] = 1
	if _, e = Fit(a, b, c); e == nil {
		t.Fatal("accepted group leakage")
	}
	b = fixture("reserve1", 2)
	if _, e = Fit(a, b, c); e == nil {
		t.Fatal("accepted final selection")
	}
}
func TestStableLoss(t *testing.T) {
	for _, v := range []float64{-1000, 1000} {
		for _, y := range []float64{0, 1} {
			if math.IsInf(loss(v, y), 0) || math.IsNaN(loss(v, y)) {
				t.Fatal("unstable BCE")
			}
		}
	}
}
func TestPrepareRejectsHoldout(t *testing.T) {
	if _, e := Prepare(nil, paireval.Split{}, "reserve1"); e == nil {
		t.Fatal("accepted holdout")
	}
}
func TestInvalidColumns(t *testing.T) {
	a, b := fixture("development", 1), fixture("validation", 2)
	a.Offsets[2] = 100
	if _, e := Fit(a, b, Config{Mode: "fp32", LearningRate: .1, Epochs: 1, Batch: 4}); e == nil {
		t.Fatal("accepted broken columns")
	}
}

func TestPrepareHoldoutIsolationAndRuntimeParity(t *testing.T) {
	zero, one := 0, 1
	rows := []paireval.Row{{Query: "find names", Code: "list names", Label: &one}, {Query: "other", Code: "unrelated", Label: &zero}, {Query: "hidden", Code: "held out", Label: &one}}
	split := paireval.Split{Rows: []paireval.Membership{{Group: 1, Split: "development"}, {Group: 2, Split: "validation"}, {Group: 3, Split: "reserve1"}}}
	a, e := Prepare(rows, split, "development")
	if e != nil {
		t.Fatal(e)
	}
	rows[2] = paireval.Row{Query: "changed", Code: "changed", Label: nil}
	b, e := Prepare(rows, split, "development")
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("holdout influenced training", e)
	}
	w := make([]float64, hintlearn.Dimension)
	for i := range w {
		w[i] = float64(i%7) / 13
	}
	want := hintlearn.Score(w, hintlearn.Features(rows[0].Query, rows[0].Code))
	if got := a.score(w, 0); got != want {
		t.Fatalf("cache changed scoring: %g != %g", got, want)
	}
}
