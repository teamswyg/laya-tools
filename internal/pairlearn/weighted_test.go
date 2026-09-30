package pairlearn

import (
	"math"
	"reflect"
	"slices"
	"testing"
)

func weightedFixture(split string, offset int) Dataset {
	return Dataset{Split: split, Offsets: []int{0, 1, 2, 3, 4}, Indices: []uint16{0, 0, 0, 0}, Values: []float64{1, 1, 1, 1}, Labels: []float64{0, 0, 0, 1}, Groups: []int{offset, offset + 1, offset + 2, offset + 3}}
}
func TestWeightedFitChangesCostPreference(t *testing.T) {
	train, val := weightedFixture("development", 0), weightedFixture("validation", 10)
	cfg := Config{Seed: 1729, Mode: "fp32", LearningRate: .1, Epochs: 100, Batch: 4, Dimension: 1}
	plain, err := Fit(train, val, cfg)
	if err != nil {
		t.Fatal(err)
	}
	train.SampleWeights = []float64{1, 1, 1, 1}
	val.SampleWeights = slices.Clone(train.SampleWeights)
	ones, err := Fit(train, val, cfg)
	if err != nil || !reflect.DeepEqual(ones, plain) {
		t.Fatal("unit weights changed legacy result", err)
	}
	train.SampleWeights = []float64{1, 1, 1, 12}
	val.SampleWeights = slices.Clone(train.SampleWeights)
	weighted, err := Fit(train, val, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if plain.Weights[0] >= 0 || weighted.Weights[0] <= 0 {
		t.Fatal("weights did not change cost preference")
	}
	train.SampleWeights = []float64{2, 2, 2, 24}
	val.SampleWeights = slices.Clone(train.SampleWeights)
	scaled, err := Fit(train, val, cfg)
	if err != nil || !reflect.DeepEqual(scaled, weighted) {
		t.Fatal("global scale changed objective", err)
	}
}
func TestWeightedNLLAndZeroWeight(t *testing.T) {
	d := weightedFixture("development", 0)
	d.SampleWeights = []float64{0, 0, 0, 3}
	want := loss(2, 1)
	if got := NLL(d, []float64{2}); math.Abs(got-want) > 1e-15 {
		t.Fatalf("got %g want %g", got, want)
	}
	d.Labels[0] = 1
	if got := NLL(d, []float64{2}); math.Abs(got-want) > 1e-15 {
		t.Fatal("zero weight changed objective")
	}
}
func TestWeightedFitRejectsInvalid(t *testing.T) {
	cfg := Config{Seed: 1, Mode: "fp32", LearningRate: .1, Epochs: 1, Batch: 4, Dimension: 1}
	for _, w := range [][]float64{{}, {1}, {0, 0, 0, 0}, {-1, 1, 1, 1}, {math.NaN(), 1, 1, 1}, {math.Inf(1), 1, 1, 1}, {4097, 1, 1, 1}, {math.SmallestNonzeroFloat64, 0, 0, 0}} {
		train, val := weightedFixture("development", 0), weightedFixture("validation", 10)
		train.SampleWeights = w
		if _, err := Fit(train, val, cfg); err == nil {
			t.Fatal("accepted invalid weights", w)
		}
	}
	train, val := weightedFixture("development", 0), weightedFixture("validation", 0)
	train.SampleWeights = []float64{1, 1, 1, 1}
	if _, err := Fit(train, val, cfg); err == nil {
		t.Fatal("weights bypassed group isolation")
	}
}
