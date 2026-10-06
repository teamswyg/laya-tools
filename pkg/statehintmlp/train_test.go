// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

func toySamples() []Sample {
	var v []Sample
	for i, label := range Intents() {
		for n := 0; n < 3; n++ {
			v = append(v, Sample{fmt.Sprintf("Original artificial class %s example %d number%d", label, i, n), label})
		}
	}
	return v
}
func saved(t *testing.T, m *Model) []byte {
	t.Helper()
	var b bytes.Buffer
	if e := m.Save(&b); e != nil {
		t.Fatal(e)
	}
	return b.Bytes()
}
func sampleLoss(t *testing.T, m *Model, ss []Sample) float64 {
	t.Helper()
	var sum float64
	var w Workspace
	var g gradient
	for _, s := range ss {
		if extract(s.Text, &w) != nil {
			t.Fatal("fixture")
		}
		target, _ := IntentIndex(s.Label)
		loss, e := m.accumulate(&w, target, &g)
		if e != nil {
			t.Fatal(e)
		}
		sum += loss
	}
	return sum / float64(len(ss))
}

func TestFiniteDifferenceActiveAndInactiveReLUDerivatives(t *testing.T) {
	m := &Model{temperature: 1}
	m.input[0][0] = .7
	m.input[1][0] = .2
	m.hiddenBias[0] = .15
	m.input[0][1] = -.2
	m.input[1][1] = .5
	m.hiddenBias[1] = -.07
	m.output[0][0] = .6
	m.output[0][1] = -.4
	m.output[1][0] = .9
	w := Workspace{count: 2, wordCount: 2}
	w.indices[0], w.indices[1] = 0, 1
	w.values[0], w.values[1] = .4, -.3
	var g gradient
	if _, e := m.accumulate(&w, 1, &g); e != nil {
		t.Fatal(e)
	}
	checks := []struct {
		name     string
		value    *float32
		analytic float64
	}{{"input active", &m.input[0][0], g.input[0][0]}, {"input negative feature", &m.input[1][0], g.input[1][0]}, {"hidden bias", &m.hiddenBias[0], g.hiddenBias[0]}, {"output target", &m.output[0][1], g.output[0][1]}, {"output other", &m.output[0][0], g.output[0][0]}, {"output bias", &m.outputBias[7], g.outputBias[7]}, {"inactive input", &m.input[0][1], g.input[0][1]}, {"inactive output", &m.output[1][0], g.output[1][0]}, {"inactive hidden bias", &m.hiddenBias[1], g.hiddenBias[1]}}
	for _, c := range checks {
		original := *c.value
		*c.value = original + 1e-3
		plus := float64(*c.value)
		var dummy gradient
		upper, e := m.accumulate(&w, 1, &dummy)
		if e != nil {
			t.Fatal(e)
		}
		*c.value = original - 1e-3
		minus := float64(*c.value)
		dummy = gradient{}
		lower, e := m.accumulate(&w, 1, &dummy)
		if e != nil {
			t.Fatal(e)
		}
		*c.value = original
		numeric := (upper - lower) / (plus - minus)
		if math.Abs(numeric-c.analytic) > 2e-6 {
			t.Fatalf("%s numerical=%g analytic=%g", c.name, numeric, c.analytic)
		}
	}
}

func TestOwnedFitLearnsFreshDeterministicallyAndInvalidInputsAreTransactional(t *testing.T) {
	ss := toySamples()
	a, b := NewModel(), NewModel()
	before := sampleLoss(t, a, ss)
	options := FitOptions{Epochs: 20, BatchSize: 8, LearningRate: .02, WeightDecay: .001, Seed: 1729}
	r, e := a.Fit(ss, options)
	if e != nil {
		t.Fatal(e)
	}
	q, e := b.Fit(ss, options)
	if e != nil || !reflect.DeepEqual(r, q) || !bytes.Equal(saved(t, a), saved(t, b)) {
		t.Fatal("seeded fit not deterministic", e)
	}
	if r.TrainingSteps != 60 || r.Batches != 60 || r.Samples != 24 || r.Initialization != Initialization || r.Seed != DefaultSeed || a.Temperature() != 1 || sampleLoss(t, a, ss) >= before/2 {
		t.Fatal("real toy fit did not lower CE/retain state")
	}
	// A second Fit deliberately resets weights/optimizer rather than warm starts.
	if a.SetTemperature(2) != nil {
		t.Fatal("temperature")
	}
	r2, e := a.Fit(ss, options)
	if e != nil || !reflect.DeepEqual(r, r2) || !bytes.Equal(saved(t, a), saved(t, b)) {
		t.Fatal("fresh Fit retained earlier weights/steps/calibration")
	}
	kept := saved(t, a)
	for _, bad := range [][]Sample{nil, {{"fixture", Intent("not-an-intent")}}, {{strings.Repeat("x", MaxTextBytes+1), Question}}, {{"invalid\x00fixture", Question}}} {
		if _, e := a.Fit(bad, options); e == nil || !bytes.Equal(saved(t, a), kept) {
			t.Fatal("invalid sample mutated learned model")
		}
	}
	for _, bad := range []FitOptions{{Epochs: -1}, {BatchSize: 4097}, {LearningRate: math.NaN()}, {WeightDecay: math.Inf(1)}} {
		if _, e := a.Fit(ss, bad); e == nil || !bytes.Equal(saved(t, a), kept) {
			t.Fatal("invalid option mutated learned model")
		}
	}
	defaults, e := (FitOptions{}).normalized()
	if e != nil || defaults.Epochs != 40 || defaults.BatchSize != 32 || defaults.LearningRate != .02 || defaults.WeightDecay != .001 || defaults.Seed != 1729 {
		t.Fatal("documented defaults changed")
	}
}

func TestStableLossAndAdamWWeightOnlyDecay(t *testing.T) {
	m := &Model{temperature: 1}
	m.outputBias[0] = 10000
	m.outputBias[1] = -10000
	var w Workspace
	var g gradient
	loss, e := m.accumulate(&w, 1, &g)
	if e != nil || !finite(loss) || math.Abs(loss-20000) > 1e-8 {
		t.Fatal("log-sum-exp loss overflowed", loss, e)
	}
	m = &Model{temperature: 1}
	m.input[0][0] = 2
	m.output[0][0] = 3
	m.hiddenBias[0] = 4
	m.outputBias[0] = 5
	a := &adamWorkspace{}
	if e := a.update(m, FitOptions{LearningRate: .02, WeightDecay: .001}, 1, 1); e != nil {
		t.Fatal(e)
	}
	if m.input[0][0] != float32(2*(1-.02*.001)) || m.output[0][0] != float32(3*(1-.02*.001)) || m.hiddenBias[0] != 4 || m.outputBias[0] != 5 {
		t.Fatal("AdamW decayed bias or skipped weight decay")
	}
}
