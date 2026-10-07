// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

func closeFloat(a, b float64) bool {
	return math.Abs(a-b) <= 1e-13*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

func TestAnalyticalIdentitySTEAndMeanIndependentHeadCE(t *testing.T) {
	// This analytical surrogate-gradient reference deliberately does not finite
	// difference the discontinuous quantizer. Head zero's 20-valued shadows are
	// saturated; they must still receive unmasked identity-STE gradients.
	var parameters statehintclaims.Parameters
	for i := range parameters.Weights {
		parameters.Weights[i][0] = [StateCount]float32{20, -2, 0}
		parameters.Weights[i][1] = [StateCount]float32{2, 2, 2}
	}
	parameters.Bias = [HeadCount][StateCount]float32{{.3, -.2, .7}, {.1, .4, -.1}, {-.5, .2, .8}}
	m := &Model{mode: PTQ}
	if err := m.quantize(&parameters); err != nil {
		t.Fatal(err)
	}
	var features statehintwide.Workspace
	view, err := statehintwide.ExtractContextual("amber kite waits", &features)
	if err != nil {
		t.Fatal(err)
	}
	var featureSum float64
	for i := 0; i < view.Len(); i++ {
		featureSum += float64(view.At(i).Value)
	}
	var probabilities [HeadCount][StateCount]float64
	expectedTrits := [HeadCount][StateCount]float64{{1, 0, 0}, {1, 1, 1}, {0, 0, 0}}
	for h := range probabilities {
		var sum float64
		for c := range probabilities[h] {
			probabilities[h][c] = math.Exp(float64(parameters.Bias[h][c]) + expectedTrits[h][c]*float64(m.scales[h])*featureSum)
			sum += probabilities[h][c]
		}
		for c := range probabilities[h] {
			probabilities[h][c] /= sum
		}
	}
	targets := [HeadCount]State{True, False, Unknown}
	var g gradient
	loss, err := m.accumulate(view, targets, &g)
	if err != nil {
		t.Fatal(err)
	}
	wantLoss := (-math.Log(probabilities[0][0]) - math.Log(probabilities[1][1]) - math.Log(probabilities[2][2])) / HeadCount
	if !closeFloat(loss, wantLoss) {
		t.Fatalf("mean CE %g want %g", loss, wantLoss)
	}
	var active [FeatureBins]bool
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		active[f.Index] = true
		for h := 0; h < HeadCount; h++ {
			for c := 0; c < StateCount; c++ {
				delta := probabilities[h][c]
				if c == h {
					delta--
				}
				delta /= HeadCount
				want := delta * float64(f.Value)
				if !closeFloat(g.weights[f.Index][h][c], want) {
					t.Fatalf("identity STE at %d/%d/%d: %g want %g", f.Index, h, c, g.weights[f.Index][h][c], want)
				}
				if !closeFloat(g.bias[h][c], delta) {
					t.Fatalf("bias gradient %d/%d: %g want %g", h, c, g.bias[h][c], delta)
				}
			}
		}
	}
	for i, used := range active {
		if !used && g.weights[i] != [HeadCount][StateCount]float64{} {
			t.Fatal("gradient leaked into inactive feature")
		}
	}
	var changed gradient
	targets[0] = False
	if _, err := m.accumulate(view, targets, &changed); err != nil {
		t.Fatal(err)
	}
	if g.bias[1] != changed.bias[1] || g.bias[2] != changed.bias[2] {
		t.Fatal("one head's target changed other heads")
	}
	for i := range g.weights {
		if g.weights[i][1] != changed.weights[i][1] || g.weights[i][2] != changed.weights[i][2] {
			t.Fatal("cross-head weight gradient")
		}
	}
}

func TestFloatShadowInitializationAndFreshAdamStepOne(t *testing.T) {
	parent, m := syntheticPTQ(t)
	parameters, err := parent.Parameters()
	if err != nil {
		t.Fatal(err)
	}
	var w TrainingWorkspace
	w.optimizer.first.weights[0][0][0] = 123
	w.optimizer.second.bias[0][0] = 456
	m.baseSteps = 2120 // Owned optimizer-counter fixture, not a loaded parent.
	w.initialize(m, parameters)
	if w.master != parameters || w.optimizer != (adamWorkspace{}) || w.forward.newSteps != 0 {
		t.Fatal("initialization did not copy float parent and clear optimizer")
	}
	differentFromQuantized := false
	for i := range parameters.Weights {
		for h := range parameters.Weights[i] {
			for _, value := range parameters.Weights[i][h] {
				q := float32(quantizedTrit(value, m.scales[h])) * m.scales[h]
				if value != q {
					differentFromQuantized = true
				}
			}
		}
	}
	if !differentFromQuantized {
		t.Fatal("fixture did not distinguish float initialization from quantized initialization")
	}
	// The same nonzero gradient has a different Adam correction at 2121. A
	// zero-gradient weight still decays, while a zero-gradient bias does not.
	w.master = statehintclaims.Parameters{}
	w.master.Weights[0][0][0], w.master.Weights[1][0][0], w.master.Bias[0][0], w.master.Bias[1][0] = 2, 3, 2, 3
	w.optimizer.gradient.weights[0][0][0], w.optimizer.gradient.bias[0][0] = .6, .6
	if err := w.optimizer.update(&w.master, 2, 1); err != nil {
		t.Fatal(err)
	}
	g := .3
	wantWeight := float32(2*(1-LearningRate*WeightDecay) - LearningRate*g/(math.Abs(g)+1e-8))
	wantBias := float32(2 - LearningRate*g/(math.Abs(g)+1e-8))
	if w.master.Weights[0][0][0] != wantWeight || w.master.Bias[0][0] != wantBias || w.master.Weights[1][0][0] != float32(3*(1-LearningRate*WeightDecay)) || w.master.Bias[1][0] != 3 {
		t.Fatalf("fresh first-step update: weight %g bias %g", w.master.Weights[0][0][0], w.master.Bias[0][0])
	}
	if !closeFloat(w.optimizer.first.weights[0][0][0], .1*g) || !closeFloat(w.optimizer.second.weights[0][0][0], .001*g*g) {
		t.Fatal("moments do not average over batch")
	}
}

func TestForwardRemainsFixedWithinBatchAndRequantizesAfterUpdate(t *testing.T) {
	parent, m := syntheticPTQ(t)
	parameters, err := parent.Parameters()
	if err != nil {
		t.Fatal(err)
	}
	var w TrainingWorkspace
	w.initialize(m, parameters)
	if err := w.forward.quantize(&w.master); err != nil {
		t.Fatal(err)
	}
	beforeMaster, beforeForward := w.master, w.forward
	for _, sample := range syntheticSamples() {
		view, err := statehintwide.ExtractContextual(sample.Text, &w.features)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.forward.accumulate(view, sample.Targets, &w.optimizer.gradient); err != nil {
			t.Fatal(err)
		}
		if w.master != beforeMaster || w.forward != beforeForward {
			t.Fatal("sample altered shadows or quantized forward before batch end")
		}
	}
	if err := w.optimizer.update(&w.master, 2, 1); err != nil {
		t.Fatal(err)
	}
	if w.master == beforeMaster || w.forward != beforeForward {
		t.Fatal("update failed to keep float master and packed forward separate")
	}
	if err := w.forward.quantize(&w.master); err != nil {
		t.Fatal(err)
	}
	if w.forward.scales == beforeForward.scales && w.forward.packed == beforeForward.packed {
		t.Fatal("new batch did not use updated shadow quantization")
	}
}

func TestWarmFitCounterStartsAtOneWithInherited2120Steps(t *testing.T) {
	parent, _ := syntheticParent(t)
	// This owned serialization fixture changes only the inherited counter. It
	// is not a real model asset or evidence that 2120 training updates occurred.
	data := parentBytes(t, parent)
	binary.LittleEndian.PutUint64(data[16:24], 2120)
	repairDigest(data)
	parent, err := statehintclaims.Load(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	m, err := FromFloat(parent, hex.EncodeToString(sum[:]))
	if err != nil {
		t.Fatal(err)
	}
	parameters, err := parent.Parameters()
	if err != nil {
		t.Fatal(err)
	}
	sample := syntheticSamples()[0]
	// An independent one-example schedule isolates the optimizer counter from
	// the continuation loop, while analytical gradients and Adam updates are
	// checked separately above. Both traces recompute quantization each batch.
	reference := func(offset uint64) statehintclaims.Parameters {
		var w TrainingWorkspace
		w.initialize(m, parameters)
		for step := uint64(1); step <= TrainingEpochs; step++ {
			w.optimizer.gradient = gradient{}
			if err := w.forward.quantize(&w.master); err != nil {
				t.Fatal(err)
			}
			view, err := statehintwide.ExtractContextual(sample.Text, &w.features)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.forward.accumulate(view, sample.Targets, &w.optimizer.gradient); err != nil {
				t.Fatal(err)
			}
			if err := w.optimizer.update(&w.master, 1, step+offset); err != nil {
				t.Fatal(err)
			}
		}
		return w.master
	}
	want, wrong := reference(0), reference(2120)
	if want == wrong {
		t.Fatal("fixture does not distinguish fresh versus inherited Adam corrections")
	}
	var workspace TrainingWorkspace
	report, err := m.WarmFit(parent, []Sample{sample}, &workspace)
	if err != nil {
		t.Fatal(err)
	}
	if workspace.master != want || report.BaseTrainingSteps != 2120 || report.NewOptimizerSteps != 40 || report.TrainingSteps != 2160 {
		t.Fatalf("warm optimizer did not start at step one: %+v", report)
	}
}

func TestWarmFitFixedRecipeFreshWorkspaceReuseAndImmutableParent(t *testing.T) {
	parent, m := syntheticPTQ(t)
	beforeParent := parentBytes(t, parent)
	ptq := *m
	samples := make([]Sample, 33)
	for i := range samples {
		samples[i] = syntheticSamples()[i%2]
	}
	var workspace TrainingWorkspace
	workspace.optimizer.first.weights[0][0][0] = 1e8
	report, err := m.WarmFit(parent, samples, &workspace)
	if err != nil {
		t.Fatal(err)
	}
	if report.Epochs != 40 || report.Batches != 80 || report.NewOptimizerSteps != 80 || report.BaseTrainingSteps != parent.TrainingSteps() || report.TrainingSteps != parent.TrainingSteps()+80 || report.Seed != DefaultSeed || !finite(report.MeanLoss) || report.Initialization != "parent_float_copy_fresh_adamw" {
		t.Fatalf("report %+v", report)
	}
	if m.mode != QAT || m.newSteps != 80 || m.Metadata().Mode != "qat" {
		t.Fatal("QAT provenance not recorded")
	}
	if !bytes.Equal(beforeParent, parentBytes(t, parent)) {
		t.Fatal("QAT mutated parent")
	}
	reference := ptq.Clone()
	if _, err := reference.WarmFit(parent, samples, &TrainingWorkspace{}); err != nil {
		t.Fatal(err)
	}
	if *m != *reference {
		t.Fatal("warm continuation depended on stale workspace moments")
	}
	// Reusing the same workspace and QAT receiver still restarts from float parent.
	if _, err := m.WarmFit(parent, samples, &workspace); err != nil {
		t.Fatal(err)
	}
	if *m != *reference {
		t.Fatal("warm continuation accumulated prior adaptation or optimizer state")
	}
}

func TestWarmFitValidationIsTransactional(t *testing.T) {
	parent, m := syntheticPTQ(t)
	beforeModel, beforeParent := *m, parentBytes(t, parent)
	badTarget := syntheticSamples()
	badTarget[0].Targets[0] = State("invalid")
	for _, samples := range [][]Sample{nil, badTarget, {{Text: "...", Targets: [HeadCount]State{True, False, Unknown}}}, {{Text: "a\x00b", Targets: [HeadCount]State{True, False, Unknown}}}} {
		if _, err := m.WarmFit(parent, samples, &TrainingWorkspace{}); err == nil {
			t.Fatal("accepted invalid samples")
		}
		if *m != beforeModel || !bytes.Equal(beforeParent, parentBytes(t, parent)) {
			t.Fatal("validation error mutated model or parent")
		}
	}
	if _, err := m.WarmFit(parent, syntheticSamples(), nil); err == nil {
		t.Fatal("accepted nil workspace")
	}
	if _, err := m.WarmFit(nil, syntheticSamples(), &TrainingWorkspace{}); err == nil {
		t.Fatal("accepted nil parent")
	}
	if _, err := m.WarmFit(statehintclaims.NewModel(), syntheticSamples(), &TrainingWorkspace{}); err == nil {
		t.Fatal("accepted a different parent artifact")
	}
	if *m != beforeModel || !bytes.Equal(beforeParent, parentBytes(t, parent)) {
		t.Fatal("parent mismatch mutated model or parent")
	}
}
