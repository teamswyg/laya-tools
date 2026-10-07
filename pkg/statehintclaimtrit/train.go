// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintclaimtrit

import (
	"math"
	"math/rand/v2"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// The experimental recipe is fixed, with no hyperparameter selection API.
const (
	TrainingEpochs    = 40
	TrainingBatchSize = 32
	LearningRate      = .001
	WeightDecay       = .01
)

type Sample = statehintclaims.Sample

type FitReport struct {
	Samples           int     `json:"samples"`
	Epochs            int     `json:"epochs"`
	Batches           int     `json:"batches"`
	TrainingSteps     uint64  `json:"training_steps"`
	BaseTrainingSteps uint64  `json:"base_training_steps"`
	NewOptimizerSteps uint64  `json:"new_optimizer_steps"`
	MeanLoss          float64 `json:"mean_loss"`
	Initialization    string  `json:"initialization"`
	Seed              int64   `json:"adaptation_seed"`
}

type gradient struct {
	weights [FeatureBins][HeadCount][StateCount]float64
	bias    [HeadCount][StateCount]float64
}

type adamWorkspace struct{ gradient, first, second gradient }

// TrainingWorkspace owns float shadow parameters and fresh Adam moments. Those
// arrays are never resident in Model or used by inference. One WarmFit call
// exclusively owns a workspace; sequential reuse is allowed. No text or feature
// corpus is retained. The order slice grows only to the supplied sample count.
type TrainingWorkspace struct {
	master    statehintclaims.Parameters
	forward   Model
	optimizer adamWorkspace
	features  statehintwide.Workspace
	order     []int
}

func (w *TrainingWorkspace) initialize(m *Model, parameters statehintclaims.Parameters) {
	w.master = parameters // Parent FLOAT masters, never dequantized PTQ weights.
	w.forward = *m
	w.forward.mode = QAT
	w.forward.newSteps = 0
	w.optimizer = adamWorkspace{}
}

// accumulate treats each head as an independent categorical distribution and
// each target, including unknown, as an explicit class. The objective is mean
// three-head CE. The identity STE sends delta*x to the float shadow weights:
// there is no scale factor, clipping mask, or scale derivative in this path.
func (m *Model) accumulate(view statehintwide.ContextualFeatureView, targets [HeadCount]State, g *gradient) (float64, error) {
	logits, err := m.logits(view)
	if err != nil {
		return 0, ErrTraining
	}
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

func update(value *float32, g float64, first, second *float64, batchScale, decay, correction1, correction2 float64) error {
	const beta1, beta2, epsilon = .9, .999, 1e-8
	g *= batchScale
	*first = beta1*(*first) + (1-beta1)*g
	*second = beta2*(*second) + (1-beta2)*g*g
	v := float64(*value)*decay - LearningRate*((*first/correction1)/(math.Sqrt(*second/correction2)+epsilon))
	if !finite(v) || math.Abs(v) > math.MaxFloat32 {
		return ErrTraining
	}
	*value = float32(v)
	return nil
}

func (a *adamWorkspace) update(p *statehintclaims.Parameters, batch int, step uint64) error {
	if batch <= 0 || step == 0 {
		return ErrTraining
	}
	scale := 1 / float64(batch)
	decay := 1 - LearningRate*WeightDecay
	c1, c2 := 1-math.Pow(.9, float64(step)), 1-math.Pow(.999, float64(step))
	for i := range p.Weights {
		for h := range p.Weights[i] {
			for c := range p.Weights[i][h] {
				if err := update(&p.Weights[i][h][c], a.gradient.weights[i][h][c], &a.first.weights[i][h][c], &a.second.weights[i][h][c], scale, decay, c1, c2); err != nil {
					return err
				}
			}
		}
	}
	for h := range p.Bias {
		for c := range p.Bias[h] {
			if err := update(&p.Bias[h][c], a.gradient.bias[h][c], &a.first.bias[h][c], &a.second.bias[h][c], scale, 1, c1, c2); err != nil {
				return err
			}
		}
	}
	return nil
}

// WarmFit copies the matching parent FLOAT parameters and starts AdamW moments
// and the optimizer step at zero. Each optimizer batch recomputes per-head
// scales and quantized forward weights once from unchanged shadows; the shadows
// and float biases update after the whole batch. The final model is repacked.
//
// Parent must match this model's verified parent artifact and must not be fitted
// concurrently. Every call restarts from that parent, including when the
// receiver previously completed QAT. Validation or numeric errors leave both
// the receiver and parent unchanged. Only caller-supplied samples are accepted;
// this API performs no data reading, scoring evaluation, or model selection.
func (m *Model) WarmFit(parent *statehintclaims.Model, samples []Sample, workspace *TrainingWorkspace) (FitReport, error) {
	if !m.valid() || workspace == nil || len(samples) == 0 || len(samples) > 1_000_000 {
		return FitReport{}, ErrTraining
	}
	digest, err := parentDigest(parent)
	if err != nil || digest != m.parentSHA || parent.TrainingSteps() != m.baseSteps || parent.InitializationSeed() != m.parentSeed {
		return FitReport{}, ErrTraining
	}
	parameters, err := parent.Parameters()
	if err != nil {
		return FitReport{}, ErrTraining
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
	batches := uint64((len(samples)+TrainingBatchSize-1)/TrainingBatchSize) * TrainingEpochs
	if batches > math.MaxUint64-m.baseSteps {
		return FitReport{}, ErrTraining
	}
	workspace.initialize(m, parameters)
	if cap(workspace.order) < len(samples) {
		workspace.order = make([]int, len(samples))
	} else {
		workspace.order = workspace.order[:len(samples)]
	}
	for i := range workspace.order {
		workspace.order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(DefaultSeed), uint64(DefaultSeed)^0x9e3779b97f4a7c15))
	r := FitReport{Samples: len(samples), Epochs: TrainingEpochs, BaseTrainingSteps: m.baseSteps, Seed: DefaultSeed, Initialization: "parent_float_copy_fresh_adamw"}
	var lossSum float64
	for epoch := 0; epoch < TrainingEpochs; epoch++ {
		random.Shuffle(len(workspace.order), func(i, j int) { workspace.order[i], workspace.order[j] = workspace.order[j], workspace.order[i] })
		for start := 0; start < len(samples); start += TrainingBatchSize {
			end := min(start+TrainingBatchSize, len(samples))
			workspace.optimizer.gradient = gradient{}
			if err := workspace.forward.quantize(&workspace.master); err != nil {
				return FitReport{}, ErrTraining
			}
			for _, index := range workspace.order[start:end] {
				sample := samples[index]
				view, err := statehintwide.ExtractContextual(sample.Text, &workspace.features)
				if err != nil {
					return FitReport{}, ErrTraining
				}
				loss, err := workspace.forward.accumulate(view, sample.Targets, &workspace.optimizer.gradient)
				if err != nil {
					return FitReport{}, err
				}
				lossSum += loss
			}
			step := workspace.forward.newSteps + 1 // Never parent steps + 1.
			if err := workspace.optimizer.update(&workspace.master, end-start, step); err != nil {
				return FitReport{}, err
			}
			workspace.forward.newSteps = step
			r.Batches++
		}
	}
	if err := workspace.forward.quantize(&workspace.master); err != nil || !workspace.forward.valid() {
		return FitReport{}, ErrTraining
	}
	r.NewOptimizerSteps = workspace.forward.newSteps
	r.TrainingSteps = workspace.forward.TrainingSteps()
	r.MeanLoss = lossSum / (float64(len(samples)) * TrainingEpochs)
	if !finite(r.MeanLoss) {
		return FitReport{}, ErrTraining
	}
	*m = workspace.forward
	return r, nil
}
