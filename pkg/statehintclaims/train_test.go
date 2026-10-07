// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"bytes"
	"math"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

func ownedView(t *testing.T, w *statehintwide.Workspace) statehintwide.ContextualFeatureView {
	t.Helper()
	v, err := statehintwide.ExtractContextual("Original owned numeric fixture: alpha beta beta.", w)
	if err != nil || v.Len() < 3 {
		t.Fatal("owned features")
	}
	return v
}
func ownedSamples() []Sample {
	return []Sample{
		{"Please explain the owned boundary.", [3]State{True, False, Unknown}},
		{"I am checking an owned boundary.", [3]State{False, True, False}},
		{"Owned ambiguous scope without a resolved claim.", [3]State{Unknown, Unknown, Unknown}},
		{"An owned close claim can also request an answer.", [3]State{True, False, True}},
	}
}
func saved(t *testing.T, m *Model) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := m.Save(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestThreeHeadMeanCEFiniteDifferenceAndUnknownTarget(t *testing.T) {
	var w statehintwide.Workspace
	v := ownedView(t, &w)
	m := NewModel()
	for i := 0; i < 3; i++ {
		f := v.At(i)
		for h := range m.bias {
			for c := range m.bias[h] {
				m.weights[f.Index][h][c] = float32(i+h-c) * .17
				m.bias[h][c] = float32(h-c) * .09
			}
		}
	}
	targets := [3]State{True, False, Unknown}
	var g gradient
	if _, err := m.accumulate(v, targets, &g); err != nil {
		t.Fatal(err)
	}
	for h := range m.bias {
		for c := range m.bias[h] {
			index := v.At(h).Index
			for _, item := range []struct {
				value *float32
				grad  float64
			}{{&m.weights[index][h][c], g.weights[index][h][c]}, {&m.bias[h][c], g.bias[h][c]}} {
				original := *item.value
				*item.value = original + .001
				hi := float64(*item.value)
				var scratch gradient
				plus, err := m.accumulate(v, targets, &scratch)
				if err != nil {
					t.Fatal(err)
				}
				*item.value = original - .001
				lo := float64(*item.value)
				scratch = gradient{}
				minus, err := m.accumulate(v, targets, &scratch)
				*item.value = original
				if err != nil || math.Abs((plus-minus)/(hi-lo)-item.grad) > 2e-7 {
					t.Fatal("finite difference disagrees", h, c, item.grad, (plus-minus)/(hi-lo))
				}
			}
		}
	}
	zero := NewModel()
	g = gradient{}
	loss, err := zero.accumulate(v, targets, &g)
	if err != nil || math.Abs(loss-math.Log(3)) > 1e-14 || math.Abs(g.bias[2][2]+2.0/9) > 1e-14 {
		t.Fatal("unknown was masked or head loss was not averaged", loss, g.bias[2])
	}
}

func TestOtherHeadTargetsDoNotCoupleGradients(t *testing.T) {
	var w statehintwide.Workspace
	v := ownedView(t, &w)
	m := NewModel()
	var a, b gradient
	if _, err := m.accumulate(v, [3]State{True, False, Unknown}, &a); err != nil {
		t.Fatal(err)
	}
	if _, err := m.accumulate(v, [3]State{False, False, Unknown}, &b); err != nil {
		t.Fatal(err)
	}
	for h := 1; h < 3; h++ {
		if a.bias[h] != b.bias[h] {
			t.Fatal("categorical heads share a normalizer")
		}
		for i := range a.weights {
			if a.weights[i][h] != b.weights[i][h] {
				t.Fatal("target in one head changed another gradient")
			}
		}
	}
	if a.bias[0] == b.bias[0] {
		t.Fatal("changed target had no effect")
	}
}

func TestAdamWKnownTwoUpdatesAndBiasNoDecay(t *testing.T) {
	value := float32(2)
	var first, second float64
	if err := update(&value, .5, &first, &second, .25, .001, .99999, .1, .001); err != nil {
		t.Fatal(err)
	}
	want := float32(2*.99999 - .001*(.125/(.125+1e-8)))
	if value != want || math.Abs(first-.0125) > 1e-14 || math.Abs(second-.000015625) > 1e-14 {
		t.Fatal("first AdamW update", value, want, first, second)
	}
	prior := float64(value)
	if err := update(&value, -.25, &first, &second, 1, .001, .99999, .19, .001999); err != nil {
		t.Fatal(err)
	}
	want = float32(prior*.99999 - .001*((-.01375/.19)/(math.Sqrt(.000078109375/.001999)+1e-8)))
	if value != want || math.Abs(first-(-.01375)) > 1e-14 || math.Abs(second-.000078109375) > 1e-14 {
		t.Fatal("second AdamW bias correction", value, want, first, second)
	}
	m := NewModel()
	m.weights[0][0][0], m.bias[0][0] = 2, 2
	var a adamWorkspace
	if a.update(m, FitOptions{LearningRate: .001, WeightDecay: .01}, 1, 1) != nil || m.weights[0][0][0] != float32(1.99998) || m.bias[0][0] != 2 {
		t.Fatal("decay must affect weights only")
	}
}

func TestFitIsFreshDeterministicAndInvalidInputTransactional(t *testing.T) {
	o := FitOptions{Epochs: 3, BatchSize: 2}
	a, b := NewModel(), NewModel()
	var wa, wb TrainingWorkspace
	ra, err := a.Fit(ownedSamples(), o, &wa)
	if err != nil || ra.TrainingSteps != 6 || ra.Initialization != "fresh_zero" || ra.Seed != DefaultSeed {
		t.Fatal("owned numeric fit", ra, err)
	}
	b.weights[17][0][0], b.steps = 11, 20
	rb, err := b.Fit(ownedSamples(), o, &wb)
	if err != nil || ra != rb || !bytes.Equal(saved(t, a), saved(t, b)) {
		t.Fatal("Fit warm-started or was nondeterministic", rb, err)
	}
	before := saved(t, a)
	for _, bad := range []FitOptions{{LearningRate: math.NaN()}, {LearningRate: math.Inf(1)}, {WeightDecay: math.Inf(-1)}, {Epochs: -1}, {BatchSize: 4097}} {
		if _, err := a.Fit(ownedSamples(), bad, &wa); err == nil || !bytes.Equal(before, saved(t, a)) {
			t.Fatal("invalid options changed model")
		}
	}
	for _, bad := range []Sample{{"owned text", [3]State{True, False, ""}}, {"owned\x00text", [3]State{True, False, Unknown}}, {"...", [3]State{True, False, Unknown}}} {
		if _, err := a.Fit([]Sample{bad}, o, &wa); err == nil || !bytes.Equal(before, saved(t, a)) {
			t.Fatal("bad target/input changed model")
		}
	}
	if _, err := a.Fit(ownedSamples(), o, nil); err == nil {
		t.Fatal("training workspace was not caller-owned")
	}
}
