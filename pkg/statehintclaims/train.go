// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaims

import (
	"math"
	"math/rand/v2"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

type Sample struct {
	Text    string           `json:"text"`
	Targets [HeadCount]State `json:"targets"`
}

// Zero fields select epochs40, batch32, rate.001, decay.01 and seed1729.
// Every Fit starts from zero weights and new Adam moments; no warm start.
type FitOptions struct {
	Epochs, BatchSize         int
	LearningRate, WeightDecay float64
	Seed                      int64
}
type FitReport struct {
	Samples        int     `json:"samples"`
	Epochs         int     `json:"epochs"`
	Batches        int     `json:"batches"`
	TrainingSteps  uint64  `json:"training_steps"`
	MeanLoss       float64 `json:"mean_loss"`
	Initialization string  `json:"initialization"`
	Seed           int64   `json:"initialization_seed"`
}

func (o FitOptions) normalized() (FitOptions, error) {
	if o.Epochs == 0 {
		o.Epochs = 40
	}
	if o.BatchSize == 0 {
		o.BatchSize = 32
	}
	if o.LearningRate == 0 {
		o.LearningRate = .001
	}
	if o.WeightDecay == 0 {
		o.WeightDecay = .01
	}
	if o.Seed == 0 {
		o.Seed = DefaultSeed
	}
	if o.Epochs < 1 || o.Epochs > 10000 || o.BatchSize < 1 || o.BatchSize > 4096 || !finite(o.LearningRate) || o.LearningRate <= 0 || o.LearningRate > 1 || !finite(o.WeightDecay) || o.WeightDecay < 0 || o.WeightDecay > 1 {
		return FitOptions{}, ErrTraining
	}
	return o, nil
}

type gradient struct {
	weights [FeatureBins][HeadCount][StateCount]float64
	bias    [HeadCount][StateCount]float64
}
type adamWorkspace struct{ gradient, first, second gradient }

// TrainingWorkspace is exclusively owned by one Fit call. Its arrays contain
// the candidate model and fresh optimizer; they are never used for inference.
// Reusing it between completed Fits is allowed. The order slice grows only to
// the supplied sample count; no text or feature corpus is cached here.
type TrainingWorkspace struct {
	working   Model
	optimizer adamWorkspace
	features  statehintwide.Workspace
	order     []int
}

// Each example contributes mean three-head CE, with a separate normalizer and
// p-y derivative for each head. Unknown contributes exactly like true/false.
func (m *Model) accumulate(view statehintwide.ContextualFeatureView, targets [HeadCount]State, g *gradient) (float64, error) {
	logits := m.logits(view)
	var deltas [HeadCount][StateCount]float64
	var loss float64
	for h := range logits {
		target, valid := StateIndex(targets[h])
		p, ok := softmax(logits[h])
		if !valid || !ok {
			return 0, ErrTraining
		}
		maxLogit := logits[h][0]
		for _, v := range logits[h] {
			maxLogit = math.Max(maxLogit, v)
		}
		var sum float64
		for _, v := range logits[h] {
			sum += math.Exp(v - maxLogit)
		}
		loss += (math.Log(sum) - (logits[h][target] - maxLogit)) / HeadCount
		for c, v := range p {
			if c == target {
				v--
			}
			deltas[h][c] = v / HeadCount
			g.bias[h][c] += deltas[h][c]
		}
	}
	if !finite(loss) {
		return 0, ErrTraining
	}
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		for h := range deltas {
			for c, delta := range deltas[h] {
				g.weights[f.Index][h][c] += float64(f.Value) * delta
			}
		}
	}
	return loss, nil
}

func update(value *float32, g float64, first, second *float64, scale, rate, decay, correction1, correction2 float64) error {
	const beta1, beta2, epsilon = .9, .999, 1e-8
	g *= scale
	*first = beta1*(*first) + (1-beta1)*g
	*second = beta2*(*second) + (1-beta2)*g*g
	v := float64(*value)*decay - rate*((*first/correction1)/(math.Sqrt(*second/correction2)+epsilon))
	if !finite(v) || math.Abs(v) > math.MaxFloat32 {
		return ErrTraining
	}
	*value = float32(v)
	return nil
}
func (a *adamWorkspace) update(m *Model, o FitOptions, batch int, step uint64) error {
	scale := 1 / float64(batch)
	decay := 1 - o.LearningRate*o.WeightDecay
	c1, c2 := 1-math.Pow(.9, float64(step)), 1-math.Pow(.999, float64(step))
	for i := range m.weights {
		for h := range m.weights[i] {
			for c := range m.weights[i][h] {
				if err := update(&m.weights[i][h][c], a.gradient.weights[i][h][c], &a.first.weights[i][h][c], &a.second.weights[i][h][c], scale, o.LearningRate, decay, c1, c2); err != nil {
					return err
				}
			}
		}
	}
	for h := range m.bias {
		for c := range m.bias[h] {
			if err := update(&m.bias[h][c], a.gradient.bias[h][c], &a.first.bias[h][c], &a.second.bias[h][c], scale, o.LearningRate, 1, c1, c2); err != nil {
				return err
			}
		}
	}
	return nil
}

// Fit validates every target and bounded input before training a fresh working
// model. Errors leave the receiver unchanged. Only caller-supplied supervised
// targets are accepted; there is no data reader, evaluator or selection policy.
func (m *Model) Fit(samples []Sample, options FitOptions, workspace *TrainingWorkspace) (FitReport, error) {
	if m == nil || !m.valid() || workspace == nil || len(samples) == 0 || len(samples) > 1_000_000 {
		return FitReport{}, ErrTraining
	}
	o, err := options.normalized()
	if err != nil {
		return FitReport{}, err
	}
	for _, sample := range samples {
		for _, target := range sample.Targets {
			if _, ok := StateIndex(target); !ok {
				return FitReport{}, ErrTraining
			}
		}
		view, err := statehintwide.ExtractContextual(sample.Text, &workspace.features)
		if err != nil || view.WordCount() == 0 {
			return FitReport{}, ErrTraining
		}
	}
	workspace.working = Model{seed: o.Seed}
	workspace.optimizer = adamWorkspace{}
	if cap(workspace.order) < len(samples) {
		workspace.order = make([]int, len(samples))
	} else {
		workspace.order = workspace.order[:len(samples)]
	}
	for i := range workspace.order {
		workspace.order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(o.Seed), uint64(o.Seed)^0x9e3779b97f4a7c15))
	r := FitReport{Samples: len(samples), Epochs: o.Epochs, Seed: o.Seed, Initialization: "fresh_zero"}
	var lossSum float64
	for epoch := 0; epoch < o.Epochs; epoch++ {
		random.Shuffle(len(workspace.order), func(i, j int) { workspace.order[i], workspace.order[j] = workspace.order[j], workspace.order[i] })
		for start := 0; start < len(samples); start += o.BatchSize {
			end := min(start+o.BatchSize, len(samples))
			workspace.optimizer.gradient = gradient{}
			for _, index := range workspace.order[start:end] {
				sample := samples[index]
				view, err := statehintwide.ExtractContextual(sample.Text, &workspace.features)
				if err != nil {
					return FitReport{}, ErrTraining
				}
				loss, err := workspace.working.accumulate(view, sample.Targets, &workspace.optimizer.gradient)
				if err != nil {
					return FitReport{}, err
				}
				lossSum += loss
			}
			if err := workspace.optimizer.update(&workspace.working, o, end-start, workspace.working.steps+1); err != nil {
				return FitReport{}, err
			}
			workspace.working.steps++
			r.Batches++
		}
	}
	r.TrainingSteps = workspace.working.steps
	r.MeanLoss = lossSum / (float64(len(samples)) * float64(o.Epochs))
	if !finite(r.MeanLoss) || !workspace.working.valid() {
		return FitReport{}, ErrTraining
	}
	*m = workspace.working
	return r, nil
}
