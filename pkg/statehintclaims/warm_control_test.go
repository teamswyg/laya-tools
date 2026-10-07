// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// Every input here is an original tiny synthetic fixture. The 2120 counter is
// an owned metadata fixture, not evidence of historical training or a real
// parent model asset. No corpus, private input or evaluation set is read.
func warmControlParent() *Model {
	p := NewModel()
	p.steps, p.seed = 2120, 731
	for i := range p.weights {
		for h := range p.weights[i] {
			for c := range p.weights[i][h] {
				p.weights[i][h][c] = float32((i%7)-2*h+c-3) * .03
			}
		}
	}
	p.bias = [HeadCount][StateCount]float32{{.12, -.21, .34}, {-.19, .31, -.14}, {.22, -.17, .09}}
	return p
}

func warmControlSamples() []Sample {
	return []Sample{
		{Text: "cedar lantern drifts", Targets: [HeadCount]State{True, False, Unknown}},
		{Text: "ochre pebble rests", Targets: [HeadCount]State{False, Unknown, True}},
	}
}

type manualWarmMoments struct {
	weights [FeatureBins][HeadCount][StateCount]float64
	bias    [HeadCount][StateCount]float64
}

// This reference directly evaluates categorical probabilities, gradients and
// AdamW. It calls no production logits/softmax/accumulate/update/Fit methods.
// Only the unchanged public text extractor supplies the signed L2 features.
func manualWarmControl(t *testing.T, parent *Model, sample Sample, offset uint64) (Model, float64) {
	t.Helper()
	m := *parent
	var featureWorkspace statehintwide.Workspace
	view, err := statehintwide.ExtractContextual(sample.Text, &featureWorkspace)
	if err != nil || view.WordCount() == 0 {
		t.Fatal("invalid owned reference features", err)
	}
	features := make([]statehintwide.SparseFeature, view.Len())
	var values [FeatureBins]float64
	for i := range features {
		features[i] = view.At(i)
		values[features[i].Index] = float64(features[i].Value)
	}
	var first, second manualWarmMoments
	var lossSum float64
	for step := uint64(1); step <= 40; step++ {
		var logits, delta [HeadCount][StateCount]float64
		for h := range logits {
			for c := range logits[h] {
				logits[h][c] = float64(m.bias[h][c])
			}
		}
		for _, feature := range features {
			for h := range logits {
				for c := range logits[h] {
					logits[h][c] += float64(m.weights[feature.Index][h][c]) * float64(feature.Value)
				}
			}
		}
		for h := range logits {
			var target int
			switch sample.Targets[h] {
			case True:
				target = 0
			case False:
				target = 1
			case Unknown:
				target = 2
			default:
				t.Fatal("invalid owned reference target")
			}
			maxLogit := math.Max(logits[h][0], math.Max(logits[h][1], logits[h][2]))
			var exponentials [StateCount]float64
			var sum float64
			for c := range exponentials {
				exponentials[c] = math.Exp(logits[h][c] - maxLogit)
				sum += exponentials[c]
			}
			lossSum += (maxLogit + math.Log(sum) - logits[h][target]) / 3
			for c, e := range exponentials {
				delta[h][c] = e / sum
				if c == target {
					delta[h][c]--
				}
				delta[h][c] /= 3
			}
		}
		for i := range m.weights {
			for h := range m.weights[i] {
				for c := range m.weights[i][h] {
					gradient := values[i] * delta[h][c]
					m.weights[i][h][c] = manualWarmAdam(m.weights[i][h][c], gradient, &first.weights[i][h][c], &second.weights[i][h][c], step+offset, true)
				}
			}
		}
		for h := range m.bias {
			for c := range m.bias[h] {
				m.bias[h][c] = manualWarmAdam(m.bias[h][c], delta[h][c], &first.bias[h][c], &second.bias[h][c], step+offset, false)
			}
		}
	}
	m.steps = parent.steps + 40
	return m, lossSum / 40
}

func manualWarmAdam(value float32, gradient float64, first, second *float64, step uint64, weight bool) float32 {
	*first = .9*(*first) + .1*gradient
	*second = .999*(*second) + .001*gradient*gradient
	firstCorrected := *first / (1 - math.Pow(.9, float64(step)))
	secondCorrected := *second / (1 - math.Pow(.999, float64(step)))
	decay := 1.0
	if weight {
		decay = .99999 // 1 - .001*.01; float biases never decay.
	}
	return float32(float64(value)*decay - .001*firstCorrected/(math.Sqrt(secondCorrected)+1e-8))
}

func TestWarmControlIndependentFortyStepReferenceAndSerializedHistory(t *testing.T) {
	parent := warmControlParent()
	beforeParent := saved(t, parent)
	sample := warmControlSamples()[0]
	want, wantLoss := manualWarmControl(t, parent, sample, 0)
	wrong, _ := manualWarmControl(t, parent, sample, 2120)
	if want.weights == wrong.weights && want.bias == wrong.bias {
		t.Fatal("owned fixture does not distinguish Adam step 1 from 2121")
	}
	// Unrelated receiver parameters and history must not choose initialization.
	receiver := NewModel()
	receiver.weights[0][0][0], receiver.bias[1][2] = 7, -13
	receiver.steps, receiver.seed = 57, 999
	var workspace TrainingWorkspace
	workspace.optimizer.first.weights[0][0][0] = 1e9
	workspace.optimizer.second.bias[0][0] = 1e9
	report, err := receiver.WarmFit(parent, []Sample{sample}, &workspace)
	if err != nil {
		t.Fatal(err)
	}
	if *receiver != want {
		t.Fatal("matched continuation differs from independent float/AdamW reference")
	}
	if math.Abs(report.MeanLoss-wantLoss) > 1e-13 || report.Samples != 1 || report.Epochs != 40 || report.Batches != 40 || report.BaseTrainingSteps != 2120 || report.NewOptimizerSteps != 40 || report.TrainingSteps != 2160 || report.Seed != 1729 || report.Initialization != "parent_float_copy_fresh_adamw" {
		t.Fatalf("warm report %+v; reference loss %g", report, wantLoss)
	}
	if !bytes.Equal(beforeParent, saved(t, parent)) {
		t.Fatal("matched continuation mutated parent")
	}
	data := saved(t, receiver)
	if len(data) != ArtifactBytes || binary.LittleEndian.Uint64(data[16:24]) != 2160 {
		t.Fatal("float artifact did not record inherited plus new updates")
	}
	loaded, err := Load(bytes.NewReader(data))
	if err != nil || *loaded != *receiver || !bytes.Equal(data, saved(t, loaded)) || loaded.TrainingSteps() != 2160 || loaded.InitializationSeed() != parent.InitializationSeed() {
		t.Fatal("float continuation reload changed parameters or history", err)
	}
}

func TestWarmControlRestartsFromParentWithReusedWorkspace(t *testing.T) {
	parent := warmControlParent()
	beforeParent := saved(t, parent)
	samples := make([]Sample, 33)
	owned := warmControlSamples()
	for i := range samples {
		samples[i] = owned[i%len(owned)]
	}
	a, b := NewModel(), NewModel()
	a.weights[0][0][0], b.weights[1][2][2] = -12, 23
	a.bias[0][0], b.bias[2][2] = 31, -41
	a.steps, b.steps = 11, 33
	var reused TrainingWorkspace
	reused.working.weights[5][1][1] = 1e5
	reused.optimizer.first.weights[0][0][0] = 1e7
	reused.optimizer.second.bias[0][0] = 1e7
	reused.order = []int{999, 998, 997}
	first, err := a.WarmFit(parent, samples, &reused)
	if err != nil {
		t.Fatal(err)
	}
	second, err := b.WarmFit(parent, samples, &TrainingWorkspace{})
	if err != nil || first != second || *a != *b {
		t.Fatal("initial receiver or stale workspace affected continuation", err)
	}
	if first.NewOptimizerSteps != 80 || first.TrainingSteps != 2200 || first.Batches != 80 {
		t.Fatalf("partial 33-example batches did not use fixed 40/32 recipe: %+v", first)
	}
	beforeReceiver := saved(t, a)
	bad := []Sample{{Text: "cedar lantern drifts", Targets: [HeadCount]State{True, State("invalid"), Unknown}}}
	if _, err := a.WarmFit(parent, bad, &reused); err != ErrTraining || !bytes.Equal(beforeReceiver, saved(t, a)) || !bytes.Equal(beforeParent, saved(t, parent)) {
		t.Fatal("bad target mutated receiver or parent", err)
	}
	// A completed receiver and previously used moments must still restart from
	// the supplied parent, even after a failed intervening continuation.
	reused.optimizer.first.weights[0][0][0] = 1e8
	reused.optimizer.second.bias[0][0] = 1e8
	again, err := a.WarmFit(parent, samples, &reused)
	if err != nil || again != first || *a != *b || !bytes.Equal(beforeParent, saved(t, parent)) {
		t.Fatal("restart retained previous adaptation, moments or parent mutation", err)
	}
}

func TestWarmControlRejectsAliasingAndValidationIsTransactional(t *testing.T) {
	parent := warmControlParent()
	receiver := NewModel()
	receiver.weights[1][1][1] = .47
	beforeParent, beforeReceiver := saved(t, parent), saved(t, receiver)
	owned := warmControlSamples()
	if _, err := parent.WarmFit(parent, owned, &TrainingWorkspace{}); err != ErrTraining || !bytes.Equal(beforeParent, saved(t, parent)) {
		t.Fatal("aliasing continuation accepted or mutated parent", err)
	}
	for _, samples := range [][]Sample{
		nil,
		{{Text: "cedar lantern drifts", Targets: [HeadCount]State{True, False, State("invalid")}}},
		{{Text: "...", Targets: [HeadCount]State{True, False, Unknown}}},
		{{Text: "a\x00b", Targets: [HeadCount]State{True, False, Unknown}}},
		{{Text: strings.Repeat("a", MaxTextBytes+1), Targets: [HeadCount]State{True, False, Unknown}}},
	} {
		if _, err := receiver.WarmFit(parent, samples, &TrainingWorkspace{}); err != ErrTraining || !bytes.Equal(beforeReceiver, saved(t, receiver)) || !bytes.Equal(beforeParent, saved(t, parent)) {
			t.Fatal("validation error mutated receiver or parent", err)
		}
	}
	if _, err := receiver.WarmFit(parent, owned, nil); err != ErrTraining {
		t.Fatal("accepted nil workspace", err)
	}
	if _, err := receiver.WarmFit(nil, owned, &TrainingWorkspace{}); err != ErrTraining {
		t.Fatal("accepted nil parent", err)
	}
	overflow := parent.Clone()
	overflow.steps = math.MaxUint64 - 39
	beforeOverflow := saved(t, overflow)
	if _, err := receiver.WarmFit(overflow, []Sample{owned[0]}, &TrainingWorkspace{}); err != ErrTraining || !bytes.Equal(beforeOverflow, saved(t, overflow)) || !bytes.Equal(beforeReceiver, saved(t, receiver)) {
		t.Fatal("overflow accepted or mutated receiver/parent", err)
	}
}
