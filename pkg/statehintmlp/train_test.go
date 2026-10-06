// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"bytes"
	"crypto/sha256"
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

// Freeze a small owned hard-CE artifact before adding the opt-in objective.
// The fixture tests arithmetic compatibility, not semantic model quality.
func TestOwnedHardCEArtifactFingerprint(t *testing.T) {
	m := NewModel()
	options := FitOptions{Epochs: 4, BatchSize: 8, LearningRate: .02, WeightDecay: .001, Seed: 1729, LabelSmoothing: 0}
	if report, err := m.Fit(toySamples(), options); err != nil || report.LabelSmoothing != 0 {
		t.Fatal(err)
	}
	// This digest was captured from the previous hard-only implementation on
	// the same original owned fixture before adding the smoothing option.
	const previous = "bbba42878930c4f1eaea52cc08ca2a4c3b63cd469bef6ff5beb2ce5ccbc22ba4"
	if got := fmt.Sprintf("%x", sha256.Sum256(saved(t, m))); got != previous {
		t.Fatalf("alpha0 changed previous hard-CE artifact: %s", got)
	}
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

func TestUniformSmoothedTargetsAndInvalidAlphaPreserveModel(t *testing.T) {
	for _, alpha := range []float64{0, .05, .2} {
		for target := 0; target < IntentCount; target++ {
			q, ok := smoothedTargets(target, alpha)
			if !ok {
				t.Fatal("valid alpha/target rejected")
			}
			var sum float64
			for c, v := range q {
				want := alpha / IntentCount
				if c == target {
					want += 1 - alpha
				}
				if v != want || v < 0 || v > 1 {
					t.Fatal("incorrect uniform target")
				}
				sum += v
			}
			if math.Abs(sum-1) > 1e-15 {
				t.Fatal("soft targets do not sum to one")
			}
		}
	}
	m := NewModel()
	if _, err := m.Fit(toySamples(), FitOptions{Epochs: 1, BatchSize: 8}); err != nil {
		t.Fatal(err)
	}
	kept := saved(t, m)
	for _, alpha := range []float64{-.01, .2001, math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, ok := smoothedTargets(0, alpha); ok {
			t.Fatal("invalid alpha accepted by target builder")
		}
		if _, err := m.Fit(toySamples(), FitOptions{Epochs: 1, LabelSmoothing: alpha}); err == nil || !bytes.Equal(saved(t, m), kept) {
			t.Fatal("invalid alpha mutated trained model")
		}
	}
	for _, target := range []int{-1, IntentCount} {
		if _, ok := smoothedTargets(target, .05); ok {
			t.Fatal("invalid target accepted")
		}
	}
}

func TestSoftTargetFiniteDifferenceGradientAndZeroExactDelegation(t *testing.T) {
	m := &Model{temperature: 1}
	m.input[0][0], m.input[1][0], m.hiddenBias[0] = .7, .2, .15
	m.input[0][1], m.input[1][1], m.hiddenBias[1] = -.2, .5, -.07
	m.output[0][0], m.output[0][1], m.output[1][0] = .6, -.4, .9
	w := Workspace{count: 2, wordCount: 2}
	w.indices[0], w.indices[1], w.values[0], w.values[1] = 0, 1, .4, -.3
	var hard, zero gradient
	lh, err := m.accumulate(&w, 1, &hard)
	if err != nil {
		t.Fatal(err)
	}
	lz, err := m.accumulateSmoothed(&w, 1, 0, &zero)
	if err != nil || lh != lz || hard != zero {
		t.Fatal("alpha0 arithmetic drifted from original hard CE")
	}
	var soft gradient
	if _, err := m.accumulateSmoothed(&w, 1, .05, &soft); err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		name     string
		value    *float32
		analytic float64
	}{{"input", &m.input[0][0], soft.input[0][0]}, {"negative feature", &m.input[1][0], soft.input[1][0]}, {"hidden bias", &m.hiddenBias[0], soft.hiddenBias[0]}, {"target output", &m.output[0][1], soft.output[0][1]}, {"other output", &m.output[0][0], soft.output[0][0]}, {"output bias", &m.outputBias[7], soft.outputBias[7]}, {"inactive input", &m.input[0][1], soft.input[0][1]}, {"inactive output", &m.output[1][0], soft.output[1][0]}, {"inactive bias", &m.hiddenBias[1], soft.hiddenBias[1]}}
	for _, c := range checks {
		original := *c.value
		*c.value = original + 1e-3
		plus := float64(*c.value)
		var dummy gradient
		upper, err := m.accumulateSmoothed(&w, 1, .05, &dummy)
		if err != nil {
			t.Fatal(err)
		}
		*c.value = original - 1e-3
		minus := float64(*c.value)
		dummy = gradient{}
		lower, err := m.accumulateSmoothed(&w, 1, .05, &dummy)
		*c.value = original
		if err != nil || math.Abs((upper-lower)/(plus-minus)-c.analytic) > 2e-6 {
			t.Fatalf("soft %s gradient mismatch: numeric=%g analytic=%g err=%v", c.name, (upper-lower)/(plus-minus), c.analytic, err)
		}
	}
}

func TestOwnedSmoothedFitUsesTargetsAndKeepsArtifactContract(t *testing.T) {
	ss := toySamples()
	meanLoss := func(m *Model) float64 {
		var sum float64
		var w Workspace
		for _, s := range ss {
			if extract(s.Text, &w) != nil {
				t.Fatal("fixture input")
			}
			target, _ := IntentIndex(s.Label)
			var g gradient
			loss, err := m.accumulateSmoothed(&w, target, .05, &g)
			if err != nil {
				t.Fatal(err)
			}
			sum += loss
		}
		return sum / float64(len(ss))
	}
	m := NewModel()
	before := meanLoss(m)
	options := FitOptions{Epochs: 4, BatchSize: 8, LearningRate: .02, WeightDecay: .001, Seed: 1729, LabelSmoothing: .05}
	report, err := m.Fit(ss, options)
	if err != nil || report.LabelSmoothing != .05 || report.TrainingSteps != 12 || m.Temperature() != 1 || meanLoss(m) >= before {
		t.Fatal("smoothed Fit did not use its objective", err)
	}
	first := saved(t, m)
	if len(first) != ArtifactBytes || ArtifactBytes != 131872 {
		t.Fatal("training option changed artifact schema")
	}
	loaded, err := Load(bytes.NewReader(first))
	if err != nil || !bytes.Equal(first, saved(t, loaded)) {
		t.Fatal("smoothed artifact reload drifted", err)
	}
	for _, s := range ss {
		p, err := m.Predict(s.Text, nil)
		if err != nil {
			t.Fatal(err)
		}
		q, err := loaded.Predict(s.Text, nil)
		if err != nil || p != q {
			t.Fatal("smoothed full prediction reload drifted", err)
		}
	}
	if _, err := m.Fit(ss, options); err != nil || !bytes.Equal(first, saved(t, m)) {
		t.Fatal("smoothed fresh Fit is not deterministic", err)
	}
}
