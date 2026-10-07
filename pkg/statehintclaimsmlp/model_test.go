// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

import (
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

func numericModel() *Model { return &Model{seed: DefaultSeed} }
func markSyntheticTrained(m *Model) {
	m.samples, m.steps = 1, FitEpochs
}

func TestArchitectureOwnedCopiesInitializationAndMetadata(t *testing.T) {
	if ParameterCount != 32937 || WeightCount != 32912 || ParameterCount*4 != 131748 || ArtifactBytes != 131972 {
		t.Fatal("architecture or byte count drift")
	}
	a, b := NewModel(), NewModel()
	if *a != *b || a.TrainingSteps() != 0 || a.InitializationSeed() != 1729 || a.Temperature() != 1 {
		t.Fatal("fresh initialization drift")
	}
	inputLimit, outputLimit := math.Sqrt(6.0/(2048+16)), math.Sqrt(6.0/(16+3))
	var nonzeroInput, nonzeroOutput int
	for _, row := range a.input {
		for _, value := range row {
			if math.Abs(float64(value)) > inputLimit || !finite(float64(value)) {
				t.Fatal("Glorot input bound")
			}
			if value != 0 {
				nonzeroInput++
			}
		}
	}
	for _, row := range a.output {
		for _, head := range row {
			for _, value := range head {
				if math.Abs(float64(value)) > outputLimit || !finite(float64(value)) {
					t.Fatal("per-head Glorot output bound")
				}
				if value != 0 {
					nonzeroOutput++
				}
			}
		}
	}
	if nonzeroInput != FeatureBins*HiddenUnits || nonzeroOutput != HiddenUnits*HeadCount*StateCount || a.hiddenBias != [HiddenUnits]float32{} || a.outputBias != [HeadCount][StateCount]float32{} {
		t.Fatal("initialization was zero, sparse or biased")
	}
	metadata := a.Metadata()
	if metadata.Architecture != Architecture || metadata.FeatureSchema != statehintwide.ContextualFeatureSchema || metadata.ParameterFloats != ParameterCount || metadata.WeightFloats != WeightCount || metadata.FeatureBins != 2048 || metadata.HiddenUnits != 16 || metadata.Heads != 3 || metadata.StatesPerHead != 3 || metadata.ParameterPrecision != "float32" || metadata.Accumulator != "float64" || metadata.Initialization != Initialization || metadata.InitializationSeed != 1729 || metadata.TrainingSteps != 0 || metadata.TrainingSamples != 0 || metadata.ConfidenceFloor != .9 || metadata.MarginFloor != .05 || metadata.Temperature != 1 || metadata.Qualified || metadata.StateAuthority {
		t.Fatal("metadata omitted or granted a contract", metadata)
	}
	p, err := a.Parameters()
	if err != nil || p.Input != a.input || p.HiddenBias != a.hiddenBias || p.Output != a.output || p.OutputBias != a.outputBias {
		t.Fatal("parameter copy drift", err)
	}
	clone := a.Clone()
	p.Input[0][0], p.HiddenBias[0], p.Output[0][0][0], p.OutputBias[0][0] = 42, 43, 44, 45
	clone.input[0][0], clone.hiddenBias[0], clone.output[0][0][0], clone.outputBias[0][0] = 51, 52, 53, 54
	if *a != *b {
		t.Fatal("owned copies aliased live parameters")
	}
	var nilModel *Model
	if nilModel.Clone() != nil || nilModel.TrainingSteps() != 0 || nilModel.InitializationSeed() != 0 {
		t.Fatal("nil accessors")
	}
	if _, err := nilModel.Parameters(); err != ErrModel {
		t.Fatal("nil parameters accepted")
	}
	if Heads() != statehintclaims.Heads() || States() != statehintclaims.States() {
		t.Fatal("common order drift")
	}
}

func TestIndependentHeadSoftmaxAndClaimGateCompatibility(t *testing.T) {
	m := numericModel()
	markSyntheticTrained(m)
	m.outputBias = [HeadCount][StateCount]float32{{8, 0, 0}, {0, 8, 0}, {0, 0, 8}}
	var w Workspace
	a, err := m.Scores("Original owned numeric inference fixture.", &w)
	if err != nil || a.WordCount != 5 {
		t.Fatal(a, err)
	}
	for head, p := range a.Probabilities {
		if math.Abs(p[0]+p[1]+p[2]-1) > 3e-16 || p[head] < .99 || a.Logits[head][head] != 8 {
			t.Fatal("heads do not normalize independently", head, a)
		}
	}
	m.outputBias[0][0] = -12
	b, err := m.Scores("Original owned numeric inference fixture.", &w)
	if err != nil || a.Probabilities[1] != b.Probabilities[1] || a.Probabilities[2] != b.Probabilities[2] || a.Probabilities[0] == b.Probabilities[0] {
		t.Fatal("one head altered other normalizers", err)
	}
	m.outputBias[0][0] = 8
	p, err := m.Predict("Original owned numeric inference fixture.", &w)
	if err != nil || p.Source != Learned || p.TrainingSteps != 40 || p.Heads[0].State != True || p.Heads[1].State != False || p.Heads[2].State != Unknown || p.Heads[2].UnknownReason != "semantic_unknown" {
		t.Fatal("claim state/gate mismatch", p, err)
	}
	// Freeze the common claim gate surface, including tie handling and reason
	// precedence. No quality or state authority follows from passing a gate.
	for _, tc := range []struct {
		p      [StateCount]float64
		steps  uint64
		words  int
		winner State
		state  State
		reason string
	}{
		{[3]float64{.96, .02, .02}, 40, 1, True, True, ""},
		{[3]float64{.02, .96, .02}, 40, 1, False, False, ""},
		{[3]float64{.02, .02, .96}, 40, 1, Unknown, Unknown, "semantic_unknown"},
		{[3]float64{.7, .2, .1}, 40, 1, True, Unknown, "low_confidence"},
		{[3]float64{.5, .5, 0}, 40, 1, True, Unknown, "low_confidence"},
		{[3]float64{1.0 / 3, 1.0 / 3, 1.0 / 3}, 40, 1, Unknown, Unknown, "semantic_unknown"},
		{[3]float64{.96, .02, .02}, 0, 1, True, Unknown, "untrained"},
		{[3]float64{.02, .02, .96}, 0, 1, Unknown, Unknown, "untrained"},
		{[3]float64{.96, .02, .02}, 0, 0, True, Unknown, "no_word_content"},
		{[3]float64{.9, .05, .05}, 40, 1, True, True, ""},
		// low_margin is mathematically unreachable for normalized 3-way p
		// that satisfies .9; exercise branch precedence directly anyway.
		{[3]float64{.91, .89, .0}, 40, 1, True, Unknown, "low_margin"},
	} {
		r := headPrediction(CompletionClaimed, tc.p, tc.steps, tc.words)
		if r.Head != "completion_claimed" || r.Winner != tc.winner || r.State != tc.state || r.UnknownReason != tc.reason || r.Probabilities != tc.p {
			t.Fatal("fixed claim gate precedence", tc, r)
		}
	}
	m.samples, m.steps = 0, 0
	p, err = m.Predict("Original owned fixture.", &w)
	if err != nil || p.Source != Untrained {
		t.Fatal("untrained source", err)
	}
	for _, head := range p.Heads {
		if head.UnknownReason != "untrained" || head.State != Unknown {
			t.Fatal("fresh model emitted a claim state")
		}
	}
	p, err = m.Predict("... 🔧", &w)
	if err != nil {
		t.Fatal(err)
	}
	for _, head := range p.Heads {
		if head.Winner != Unknown || head.State != Unknown || head.UnknownReason != "no_word_content" || head.Confidence != 1.0/3 || head.Margin != 0 {
			t.Fatal("no-word neutralization", head)
		}
	}
}

func TestInputAndNonfiniteInferenceErrors(t *testing.T) {
	m := NewModel()
	var w Workspace
	for _, text := range []string{"bad\x00word", string([]byte{0xff}), strings.Repeat("x", MaxTextBytes+1)} {
		if _, err := m.Predict(text, &w); err != ErrInput {
			t.Fatal("invalid input accepted", err)
		}
	}
	if _, err := m.Predict("owned", nil); err != ErrInput {
		t.Fatal("nil inference workspace accepted", err)
	}
	var nilModel *Model
	if _, err := nilModel.Predict("owned", &w); err != ErrModel {
		t.Fatal("nil model accepted", err)
	}
	m.hiddenBias[0] = float32(math.NaN())
	if _, err := m.Scores("owned", &w); err != ErrModel {
		t.Fatal("nonfinite activation accepted", err)
	}
	m = NewModel()
	m.output[0][0][0] = float32(math.Inf(1))
	if _, err := m.Scores("owned", &w); err != ErrModel {
		t.Fatal("nonfinite output accepted", err)
	}
	m = NewModel()
	m.outputBias[0][0] = float32(math.Inf(-1))
	if _, err := m.Predict("owned", &w); err != ErrModel {
		t.Fatal("nonfinite bias accepted", err)
	}
	if _, ok := softmax([3]float64{math.NaN(), 0, 0}); ok {
		t.Fatal("nonfinite softmax")
	}
}

func TestWarmPredictZeroAllocationsAndConcurrentReadOnly(t *testing.T) {
	m := NewModel()
	var w Workspace
	const text = "Original synthetic concurrent fixture with repeated repeated terms."
	want, err := m.Predict(text, &w)
	if err != nil {
		t.Fatal(err)
	}
	before := *m
	if allocations := testing.AllocsPerRun(100, func() {
		if _, err := m.Predict(text, &w); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatal("warm Predict allocated", allocations)
	}
	var wg sync.WaitGroup
	for caller := 0; caller < 8; caller++ {
		wg.Go(func() {
			var local Workspace
			for i := 0; i < 50; i++ {
				got, err := m.Predict(text, &local)
				if err != nil || got != want {
					t.Error("concurrent prediction drift", err)
					return
				}
			}
		})
	}
	wg.Wait()
	if *m != before {
		t.Fatal("inference mutated the model")
	}
}
