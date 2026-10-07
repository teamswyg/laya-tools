// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimsmlp

import (
	"math"
	"math/rand/v2"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	FitEpochs       = 40
	FitBatchSize    = 32
	FitLearningRate = .001
	FitWeightDecay  = .01
	maxSamples      = 1_000_000
)

// Sample contains only text and three categorical targets. Annotation,
// provenance, application state, and evaluation scores are never input features.
type Sample = statehintclaims.Sample

type FitReport struct {
	Samples        int     `json:"samples"`
	Epochs         int     `json:"epochs"`
	Batches        int     `json:"batches"`
	TrainingSteps  uint64  `json:"training_steps"`
	MeanLoss       float64 `json:"mean_loss"`
	Initialization string  `json:"initialization"`
	Seed           int64   `json:"initialization_seed"`
	BatchSize      int     `json:"batch_size"`
	LearningRate   float64 `json:"learning_rate"`
	WeightDecay    float64 `json:"weight_decay"`
}

type gradient struct {
	input      [FeatureBins][HiddenUnits]float64
	hiddenBias [HiddenUnits]float64
	output     [HiddenUnits][HeadCount][StateCount]float64
	outputBias [HeadCount][StateCount]float64
}
type adamWorkspace struct{ gradient, first, second gradient }

// TrainingWorkspace belongs exclusively to one Fit call. Reuse between calls
// is allowed; each Fit resets the float32 working model, gradient, both moments
// and step count. Only integer shuffle indices survive, never a feature/text
// corpus. The last borrowed text features/activations are cleared on return.
type TrainingWorkspace struct {
	working   Model
	optimizer adamWorkspace
	inference Workspace
	order     []int
}

// accumulate adds mean three-head categorical CE. The shared hidden gradient
// sums the three heads' mean-loss contributions. Unknown is an ordinary target;
// ReLU's derivative is exactly zero for both zero and negative preactivation.
func (m *Model) accumulate(view statehintwide.ContextualFeatureView, w *Workspace, targets [HeadCount]State, g *gradient) (float64, error) {
	logits, ok := m.logits(view, w)
	if !ok {
		return 0, ErrTraining
	}
	var hiddenDelta [HiddenUnits]float64
	var loss float64
	for head := range logits {
		target, valid := StateIndex(targets[head])
		p, ok := softmax(logits[head])
		if !valid || !ok {
			return 0, ErrTraining
		}
		maximum := logits[head][0]
		for _, value := range logits[head] {
			maximum = math.Max(maximum, value)
		}
		var total float64
		for _, value := range logits[head] {
			total += math.Exp(value - maximum)
		}
		loss += (math.Log(total) - (logits[head][target] - maximum)) / HeadCount
		for c, value := range p {
			if c == target {
				value--
			}
			delta := value / HeadCount
			g.outputBias[head][c] += delta
			for h, activation := range w.hidden {
				g.output[h][head][c] += delta * activation
				hiddenDelta[h] += delta * float64(m.output[h][head][c])
			}
		}
	}
	if !finite(loss) {
		return 0, ErrTraining
	}
	for h, z := range w.preactivation {
		if z <= 0 {
			continue
		}
		delta := hiddenDelta[h]
		g.hiddenBias[h] += delta
		for i := 0; i < view.Len(); i++ {
			feature := view.At(i)
			g.input[feature.Index][h] += delta * float64(feature.Value)
		}
	}
	return loss, nil
}

func update(value *float32, g float64, first, second *float64, scale, decay, c1, c2 float64) error {
	const beta1, beta2, epsilon = .9, .999, 1e-8
	g *= scale
	if !finite(g) || !finite(*first) || !finite(*second) || *second < 0 || !finite(c1) || !finite(c2) || c1 <= 0 || c2 <= 0 {
		return ErrTraining
	}
	moment := beta1*(*first) + (1-beta1)*g
	variance := beta2*(*second) + (1-beta2)*g*g
	value64 := float64(*value)*decay - FitLearningRate*((moment/c1)/(math.Sqrt(variance/c2)+epsilon))
	if !finite(moment) || !finite(variance) || !finite(value64) || math.Abs(value64) > math.MaxFloat32 {
		return ErrTraining
	}
	*first, *second, *value = moment, variance, float32(value64)
	return nil
}

func (a *adamWorkspace) update(m *Model, batch int, step uint64) error {
	if batch < 1 || batch > FitBatchSize || step == 0 || step > uint64(FitEpochs)*((maxSamples+FitBatchSize-1)/FitBatchSize) {
		return ErrTraining
	}
	scale := 1 / float64(batch)
	decay := 1 - FitLearningRate*FitWeightDecay
	c1, c2 := 1-math.Pow(.9, float64(step)), 1-math.Pow(.999, float64(step))
	for i := range m.input {
		for h := range m.input[i] {
			if err := update(&m.input[i][h], a.gradient.input[i][h], &a.first.input[i][h], &a.second.input[i][h], scale, decay, c1, c2); err != nil {
				return err
			}
		}
	}
	for h := range m.output {
		for head := range m.output[h] {
			for c := range m.output[h][head] {
				if err := update(&m.output[h][head][c], a.gradient.output[h][head][c], &a.first.output[h][head][c], &a.second.output[h][head][c], scale, decay, c1, c2); err != nil {
					return err
				}
			}
		}
	}
	for h := range m.hiddenBias {
		if err := update(&m.hiddenBias[h], a.gradient.hiddenBias[h], &a.first.hiddenBias[h], &a.second.hiddenBias[h], scale, 1, c1, c2); err != nil {
			return err
		}
	}
	for head := range m.outputBias {
		for c := range m.outputBias[head] {
			if err := update(&m.outputBias[head][c], a.gradient.outputBias[head][c], &a.first.outputBias[head][c], &a.second.outputBias[head][c], scale, 1, c1, c2); err != nil {
				return err
			}
		}
	}
	return nil
}

// Fit is always cold: 40 epochs, batch32, AdamW rate .001, weight decay .01,
// seed1729, hard targets, fixed T1 and fresh Glorot weights/moments/step zero.
// It has no tunable options, smoothing, calibration or warm-start API. Every
// sample's bounded UTF-8, word content and all targets validate before updates.
// Validation or numerical errors leave the receiver exactly unchanged.
func (m *Model) Fit(samples []Sample, workspace *TrainingWorkspace) (FitReport, error) {
	if m == nil || !m.valid() || workspace == nil || len(samples) == 0 || len(samples) > maxSamples {
		return FitReport{}, ErrTraining
	}
	defer func() { workspace.inference = Workspace{} }()
	for _, sample := range samples {
		for _, target := range sample.Targets {
			if _, ok := StateIndex(target); !ok {
				return FitReport{}, ErrTraining
			}
		}
		view, err := statehintwide.ExtractContextual(sample.Text, &workspace.inference.features)
		if err != nil || view.WordCount() == 0 {
			return FitReport{}, ErrTraining
		}
	}
	workspace.working.initialize()
	workspace.optimizer = adamWorkspace{}
	if cap(workspace.order) < len(samples) {
		workspace.order = make([]int, len(samples))
	} else {
		workspace.order = workspace.order[:len(samples)]
	}
	for i := range workspace.order {
		workspace.order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(DefaultSeed), uint64(DefaultSeed)^0x9e3779b97f4a7c15))
	r := FitReport{Samples: len(samples), Epochs: FitEpochs, Seed: DefaultSeed, Initialization: Initialization, BatchSize: FitBatchSize, LearningRate: FitLearningRate, WeightDecay: FitWeightDecay}
	var lossSum float64
	for epoch := 0; epoch < FitEpochs; epoch++ {
		random.Shuffle(len(workspace.order), func(i, j int) { workspace.order[i], workspace.order[j] = workspace.order[j], workspace.order[i] })
		for start := 0; start < len(samples); start += FitBatchSize {
			end := min(start+FitBatchSize, len(samples))
			workspace.optimizer.gradient = gradient{}
			for _, index := range workspace.order[start:end] {
				sample := samples[index]
				view, err := statehintwide.ExtractContextual(sample.Text, &workspace.inference.features)
				if err != nil {
					return FitReport{}, ErrTraining
				}
				loss, err := workspace.working.accumulate(view, &workspace.inference, sample.Targets, &workspace.optimizer.gradient)
				if err != nil {
					return FitReport{}, err
				}
				lossSum += loss
				if !finite(lossSum) {
					return FitReport{}, ErrTraining
				}
			}
			if err := workspace.optimizer.update(&workspace.working, end-start, workspace.working.steps+1); err != nil {
				return FitReport{}, err
			}
			workspace.working.steps++
			r.Batches++
		}
	}
	workspace.working.samples = uint32(len(samples))
	r.TrainingSteps = workspace.working.steps
	r.MeanLoss = lossSum / (float64(len(samples)) * FitEpochs)
	if !finite(r.MeanLoss) || !workspace.working.valid() {
		return FitReport{}, ErrTraining
	}
	*m = workspace.working
	return r, nil
}
