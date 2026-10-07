// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

func ownedSamples() []Sample {
	return []Sample{
		{Text: "Original synthetic amber boundary explanation.", Targets: [3]State{True, False, Unknown}},
		{Text: "Original synthetic cobalt boundary checking.", Targets: [3]State{False, True, False}},
		{Text: "Original synthetic violet boundary ambiguity.", Targets: [3]State{Unknown, Unknown, Unknown}},
		{Text: "Original synthetic saffron boundary closed and answer requested.", Targets: [3]State{True, False, True}},
	}
}

func ownedView(t *testing.T, w *Workspace) statehintwide.ContextualFeatureView {
	t.Helper()
	view, err := statehintwide.ExtractContextual("Original owned alpha beta beta numeric fixture.", &w.features)
	if err != nil || view.Len() < 4 {
		t.Fatal("invalid original numeric fixture", err)
	}
	return view
}

// independentLoss deliberately bypasses production logits, softmax and
// accumulate. It explicitly evaluates three log-sum-exp categorical losses.
func independentLoss(m *Model, view statehintwide.ContextualFeatureView, targets [HeadCount]State) float64 {
	var activation [HiddenUnits]float64
	for hidden := 0; hidden < HiddenUnits; hidden++ {
		z := float64(m.hiddenBias[hidden])
		for i := 0; i < view.Len(); i++ {
			f := view.At(i)
			z += float64(f.Value) * float64(m.input[f.Index][hidden])
		}
		if z > 0 {
			activation[hidden] = z
		}
	}
	var loss float64
	for head := 0; head < HeadCount; head++ {
		var logit [StateCount]float64
		for state := 0; state < StateCount; state++ {
			logit[state] = float64(m.outputBias[head][state])
			for hidden := 0; hidden < HiddenUnits; hidden++ {
				logit[state] += activation[hidden] * float64(m.output[hidden][head][state])
			}
		}
		maximum := math.Max(logit[0], math.Max(logit[1], logit[2]))
		partition := math.Exp(logit[0]-maximum) + math.Exp(logit[1]-maximum) + math.Exp(logit[2]-maximum)
		target := 2
		if targets[head] == True {
			target = 0
		} else if targets[head] == False {
			target = 1
		}
		loss += (math.Log(partition) + maximum - logit[target]) / 3
	}
	return loss
}

func TestIndependentFiniteDifferencesAllFourBlocksAndReLUZeroNegative(t *testing.T) {
	var w Workspace
	view := ownedView(t, &w)
	m := numericModel()
	for h := 0; h < HiddenUnits; h++ {
		if h < 8 {
			m.hiddenBias[h] = .7 + float32(h)*.025
		} else if h < 15 {
			m.hiddenBias[h] = -.7
		}
		for i := 0; i < 3 && h < 15; i++ {
			m.input[view.At(i).Index][h] = float32((i+1)*(h+1)) * .001
		}
		for head := 0; head < HeadCount; head++ {
			for c := 0; c < StateCount; c++ {
				m.output[h][head][c] = float32(h+1)*.017 + float32(head-c)*.12
				m.outputBias[head][c] = float32(head-c) * .09
			}
		}
	}
	targets := [3]State{True, False, Unknown}
	var g gradient
	loss, err := m.accumulate(view, &w, targets, &g)
	if err != nil || math.Abs(loss-independentLoss(m, view, targets)) > 5e-16 || w.preactivation[15] != 0 {
		t.Fatal("mean CE or zero preactivation drift", loss, err)
	}
	check := func(name string, value *float32, analytic float64) {
		t.Helper()
		original := *value
		*value = original + .001
		plusValue, plusLoss := float64(*value), independentLoss(m, view, targets)
		*value = original - .001
		minusValue, minusLoss := float64(*value), independentLoss(m, view, targets)
		*value = original
		numeric := (plusLoss - minusLoss) / (plusValue - minusValue)
		if math.Abs(numeric-analytic) > 3e-8 {
			t.Fatalf("%s: independent numerical=%g analytic=%g", name, numeric, analytic)
		}
	}
	// Include signed active features, all active/inactive hidden units, every
	// head/state output and bias, plus a truly absent input row.
	var negativeFeature bool
	for i := 0; i < view.Len(); i++ {
		if view.At(i).Value < 0 {
			negativeFeature = true
			for h := 0; h < 15; h++ {
				index := view.At(i).Index
				check("signed input", &m.input[index][h], g.input[index][h])
			}
			break
		}
	}
	if !negativeFeature {
		t.Fatal("numeric fixture lacks a signed feature")
	}
	for h := 0; h < HiddenUnits; h++ {
		if h < 15 { // Central differences at ReLU's kink cannot define its convention.
			check("hidden bias", &m.hiddenBias[h], g.hiddenBias[h])
			for i := 0; i < 3; i++ {
				index := view.At(i).Index
				check("input", &m.input[index][h], g.input[index][h])
			}
		}
		for head := 0; head < HeadCount; head++ {
			for c := 0; c < StateCount; c++ {
				check("output", &m.output[h][head][c], g.output[h][head][c])
			}
		}
		if h >= 8 {
			if g.hiddenBias[h] != 0 {
				t.Fatal("ReLU derivative at zero/negative was nonzero", h)
			}
			for i := 0; i < view.Len(); i++ {
				if g.input[view.At(i).Index][h] != 0 {
					t.Fatal("ReLU zero/negative propagated input gradient", h)
				}
			}
		}
	}
	for head := 0; head < HeadCount; head++ {
		for c := 0; c < StateCount; c++ {
			check("output bias", &m.outputBias[head][c], g.outputBias[head][c])
		}
	}
	var present [FeatureBins]bool
	for i := 0; i < view.Len(); i++ {
		present[view.At(i).Index] = true
	}
	for index, active := range present {
		if !active {
			check("absent input", &m.input[index][0], g.input[index][0])
			break
		}
	}
}

func TestMeanThreeHeadLossUnknownTargetAndSharedRepresentation(t *testing.T) {
	m := numericModel()
	m.hiddenBias[0] = 1
	m.output[0] = [HeadCount][StateCount]float32{{.5, -.2, .1}, {-.3, .7, .1}, {0, .2, -.4}}
	var w Workspace
	view := ownedView(t, &w)
	var a, b gradient
	if _, err := m.accumulate(view, &w, [3]State{True, False, Unknown}, &a); err != nil {
		t.Fatal(err)
	}
	if _, err := m.accumulate(view, &w, [3]State{False, False, Unknown}, &b); err != nil {
		t.Fatal(err)
	}
	for head := 1; head < 3; head++ {
		if a.outputBias[head] != b.outputBias[head] {
			t.Fatal("output heads share a target/normalizer")
		}
		for h := range a.output {
			if a.output[h][head] != b.output[h][head] {
				t.Fatal("another head target changed output-head gradient")
			}
		}
	}
	if a.hiddenBias[0] == b.hiddenBias[0] || a.outputBias[0] == b.outputBias[0] {
		t.Fatal("shared hidden representation did not receive changed-head loss")
	}
	m = numericModel()
	a = gradient{}
	loss, err := m.accumulate(view, &w, [3]State{True, False, Unknown}, &a)
	if err != nil || math.Abs(loss-math.Log(3)) > 1e-14 || math.Abs(a.outputBias[2][2]+2.0/9) > 1e-14 {
		t.Fatal("unknown masked or CE not averaged", loss, a.outputBias, err)
	}
	m.outputBias = [3][3]float32{{10000, -10000, 0}, {0, 10000, -10000}, {-10000, 0, 10000}}
	a = gradient{}
	loss, err = m.accumulate(view, &w, [3]State{False, Unknown, True}, &a)
	if err != nil || loss != 20000 || !finite(loss) {
		t.Fatal("log-sum-exp loss was not stable", loss, err)
	}
}

func TestScalarManualAdamFirstAndFortyStepFixtureAndBiasNoDecay(t *testing.T) {
	value := float32(2)
	var first, second float64
	if err := update(&value, .5, &first, &second, .25, .99999, .1, .001); err != nil {
		t.Fatal(err)
	}
	want := float32(2*.99999 - .001*(.125/(.125+1e-8)))
	if value != want || math.Abs(first-.0125) > 1e-14 || math.Abs(second-.000015625) > 1e-14 {
		t.Fatal("manual scalar first Adam step", value, want, first, second)
	}
	value, first, second = 2, 0, 0
	manual := float32(2)
	var mean, variance float64
	power1, power2 := 1.0, 1.0
	for step := 1; step <= 40; step++ {
		derivative := [3]float64{.125, -.25, .5}[(step-1)%3]
		power1, power2 = power1*.9, power2*.999
		mean = .9*mean + .1*derivative
		variance = .999*variance + .001*derivative*derivative
		correctedMean, correctedVariance := mean/(1-power1), variance/(1-power2)
		manual = float32(float64(manual)*.99999 - .001*correctedMean/(math.Sqrt(correctedVariance)+1e-8))
		if err := update(&value, derivative, &first, &second, 1, .99999, 1-math.Pow(.9, float64(step)), 1-math.Pow(.999, float64(step))); err != nil || value != manual || math.Abs(first-mean) > 1e-15 || math.Abs(second-variance) > 1e-15 {
			t.Fatal("scalar manual 40-step recurrence", step, value, manual, err)
		}
	}
	// Freeze the independent recurrence's float32 terminal value; this is a
	// numeric fixture, not a training/semantic-quality result.
	const terminalBits = uint32(0x3ffdfcd9)
	if math.Float32bits(value) != terminalBits {
		t.Fatal("fixed 40-step scalar fixture changed", value)
	}
	m := numericModel()
	m.input[0][0], m.output[0][0][0], m.hiddenBias[0], m.outputBias[0][0] = 2, 3, 4, 5
	var adam adamWorkspace
	if err := adam.update(m, 32, 1); err != nil || m.input[0][0] != float32(2*.99999) || m.output[0][0][0] != float32(3*.99999) || m.hiddenBias[0] != 4 || m.outputBias[0][0] != 5 {
		t.Fatal("both weights must decay; both biases must not", err)
	}
	for _, args := range []struct {
		batch int
		step  uint64
	}{{0, 1}, {33, 1}, {1, 0}, {1, math.MaxUint64}} {
		if adam.update(m, args.batch, args.step) != ErrTraining {
			t.Fatal("invalid optimizer update accepted", args)
		}
	}
	for _, derivative := range []float64{math.NaN(), math.Inf(1), math.MaxFloat64} {
		value, first, second = 2, 0, 0
		if err := update(&value, derivative, &first, &second, 1, .99999, .1, .001); err != ErrTraining || value != 2 || first != 0 || second != 0 {
			t.Fatal("nonfinite scalar update was not transactional", err)
		}
	}
}

func meanLoss(t *testing.T, m *Model, samples []Sample) float64 {
	t.Helper()
	var w Workspace
	var total float64
	for _, sample := range samples {
		view, err := statehintwide.ExtractContextual(sample.Text, &w.features)
		if err != nil {
			t.Fatal(err)
		}
		total += independentLoss(m, view, sample.Targets)
	}
	return total / float64(len(samples))
}

func TestFixedColdFitFreshMomentsRepeatedFitNoAliasAndTransactionalErrors(t *testing.T) {
	samples := ownedSamples()
	a, b := NewModel(), NewModel()
	beforeLoss := meanLoss(t, a, samples)
	var wa, wb TrainingWorkspace
	ra, err := a.Fit(samples, &wa)
	if err != nil || ra.Samples != 4 || ra.Epochs != 40 || ra.Batches != 40 || ra.TrainingSteps != 40 || ra.Initialization != Initialization || ra.Seed != 1729 || ra.BatchSize != 32 || ra.LearningRate != .001 || ra.WeightDecay != .01 || a.Temperature() != 1 || meanLoss(t, a, samples) >= beforeLoss {
		t.Fatal("fixed recipe did not reduce owned CE", ra, err)
	}
	b.input[17][0], b.hiddenBias[0], b.output[0][0][0], b.outputBias[0][0] = 11, 12, 13, 14
	markSyntheticTrained(b)
	wb.optimizer.first.input[0][0], wb.optimizer.second.hiddenBias[0] = math.NaN(), math.Inf(1)
	wb.working.steps = math.MaxUint64
	rb, err := b.Fit(samples, &wb)
	if err != nil || ra != rb || *a != *b || !bytes.Equal(saved(t, a), saved(t, b)) {
		t.Fatal("Fit retained old weights, moments or steps", rb, err)
	}
	if wa.inference != (Workspace{}) || wb.inference != (Workspace{}) {
		t.Fatal("training workspace retained last text/features")
	}
	kept := *a
	wa.working.input[0][0] = 42
	wa.optimizer.first.input[0][0], wa.optimizer.second.input[0][0] = math.Inf(-1), math.NaN()
	if *a != kept {
		t.Fatal("trained receiver aliases working model")
	}
	ra2, err := a.Fit(samples, &wa)
	if err != nil || ra2 != ra || *a != kept {
		t.Fatal("repeated Fit was not fresh and deterministic", ra2, err)
	}
	for _, invalid := range [][]Sample{
		nil,
		{{Text: "owned", Targets: [3]State{True, False, ""}}},
		{{Text: "... 🔧", Targets: [3]State{True, False, Unknown}}},
		{{Text: "owned\x00word", Targets: [3]State{True, False, Unknown}}},
		{{Text: string([]byte{0xff}), Targets: [3]State{True, False, Unknown}}},
		{{Text: strings.Repeat("x", MaxTextBytes+1), Targets: [3]State{True, False, Unknown}}},
		{samples[0], {Text: "invalid later sample", Targets: [3]State{"invalid", False, Unknown}}},
	} {
		working, optimizer := wa.working, wa.optimizer
		if _, err := a.Fit(invalid, &wa); err != ErrTraining || *a != kept || wa.working != working || wa.optimizer != optimizer {
			t.Fatal("validation changed receiver or began optimizer updates", err)
		}
	}
	if _, err := a.Fit(samples, nil); err != ErrTraining || *a != kept {
		t.Fatal("missing training workspace mutated model", err)
	}
	var nilModel *Model
	if _, err := nilModel.Fit(samples, &wa); err != ErrTraining {
		t.Fatal("nil training receiver accepted", err)
	}
	bad := a.Clone()
	bad.input[0][0] = float32(math.NaN())
	if _, err := bad.Fit(samples, &wa); err != ErrTraining || !math.IsNaN(float64(bad.input[0][0])) {
		t.Fatal("invalid receiver bypassed validation", err)
	}
	if wa.inference != (Workspace{}) {
		t.Fatal("failed Fit retained last text/features")
	}
}

func TestFixedStepCountsPartialBatchAndSyntheticArtifactPin(t *testing.T) {
	if !validHistory(2120, 1680, 1729) || validHistory(2080, 1680, 1729) || validHistory(2121, 1680, 1729) || validHistory(math.MaxUint64, 1680, 1729) || validHistory(0, 1, 1729) || validHistory(40, 0, 1729) || validHistory(0, 0, 1) || validHistory(40, maxSamples+1, 1729) {
		t.Fatal("fixed sample/step counts or bounds drift")
	}
	samples := make([]Sample, 33)
	for i := range samples {
		samples[i] = Sample{Text: fmt.Sprintf("Original synthetic partial batch fixture number %d.", i), Targets: [3]State{True, False, Unknown}}
	}
	m := NewModel()
	var w TrainingWorkspace
	r, err := m.Fit(samples, &w)
	if err != nil || r.TrainingSteps != 80 || r.Batches != 80 || r.Samples != 33 || m.samples != 33 {
		t.Fatal("partial batch changed fixed epoch/update relation", r, err)
	}
	// Exactly this original four-row synthetic fixture pins initialization,
	// shuffle, mean CE, float32 updates and artifact contract together.
	m = NewModel()
	if _, err := m.Fit(ownedSamples(), &w); err != nil {
		t.Fatal(err)
	}
	const pin = "259d8d0326ba17d8c281b05b79d129796961694ac0b0a33ef8b5c0247a1bc795"
	got := fmt.Sprintf("%x", sha256.Sum256(saved(t, m)))
	if got != pin {
		t.Fatal("synthetic fixture numeric artifact drift", got)
	}
}
