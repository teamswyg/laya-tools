// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"math"
	"math/rand/v2"
)

type Sample struct {
	Text  string `json:"text"`
	Label Intent `json:"label"`
}

// Zero values select40 epochs, batch32, lr.02, decay.001 and seed1729. Every
// valid Fit reinitializes weights and optimizer; warm-start is not supported.
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
		o.LearningRate = .02
	}
	if o.WeightDecay == 0 {
		o.WeightDecay = .001
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
	input      [FeatureBins][HiddenUnits]float64
	hiddenBias [HiddenUnits]float64
	output     [HiddenUnits][IntentCount]float64
	outputBias [IntentCount]float64
}
type adamWorkspace struct{ gradient, first, second gradient }

// gradient adds one untempered eight-way cross-entropy derivative. ReLU uses
// zero derivative at zero. Weights/activations are unchanged until batch update.
func (m *Model) accumulate(w *Workspace, target int, g *gradient) (float64, error) {
	logits := m.logits(w)
	p, ok := softmax(logits, 1)
	if !ok || target < 0 || target >= IntentCount {
		return 0, ErrTraining
	}
	maxLogit := logits[0]
	for _, v := range logits {
		maxLogit = math.Max(maxLogit, v)
	}
	var sum float64
	for _, v := range logits {
		sum += math.Exp(v - maxLogit)
	}
	loss := maxLogit + math.Log(sum) - logits[target]
	if !finite(loss) {
		return 0, ErrTraining
	}
	var hiddenDelta [HiddenUnits]float64
	for c, value := range p {
		delta := value
		if c == target {
			delta--
		}
		g.outputBias[c] += delta
		for h, activation := range w.hidden {
			g.output[h][c] += delta * activation
			hiddenDelta[h] += delta * float64(m.output[h][c])
		}
	}
	for h, z := range w.preactivation {
		if z <= 0 {
			continue
		}
		delta := hiddenDelta[h]
		g.hiddenBias[h] += delta
		for _, index := range w.indices[:w.count] {
			g.input[index][h] += delta * float64(w.values[index])
		}
	}
	return loss, nil
}
func update(value *float32, g float64, first, second *float64, scale, rate, decay, c1, c2 float64) error {
	const beta1, beta2, epsilon = .9, .999, 1e-8
	g *= scale
	*first = beta1*(*first) + (1-beta1)*g
	*second = beta2*(*second) + (1-beta2)*g*g
	v := float64(*value)*decay - rate*((*first/c1)/(math.Sqrt(*second/c2)+epsilon))
	if !finite(v) || math.Abs(v) > math.MaxFloat32 {
		return ErrTraining
	}
	*value = float32(v)
	return nil
}
func (a *adamWorkspace) update(m *Model, o FitOptions, batch int, step uint64) error {
	scale := 1 / float64(batch)
	decay := 1 - o.LearningRate*o.WeightDecay
	c1 := 1 - math.Pow(.9, float64(step))
	c2 := 1 - math.Pow(.999, float64(step))
	for i := range m.input {
		for h := range m.input[i] {
			if e := update(&m.input[i][h], a.gradient.input[i][h], &a.first.input[i][h], &a.second.input[i][h], scale, o.LearningRate, decay, c1, c2); e != nil {
				return e
			}
		}
	}
	for h := range m.output {
		for c := range m.output[h] {
			if e := update(&m.output[h][c], a.gradient.output[h][c], &a.first.output[h][c], &a.second.output[h][c], scale, o.LearningRate, decay, c1, c2); e != nil {
				return e
			}
		}
	}
	for h := range m.hiddenBias {
		if e := update(&m.hiddenBias[h], a.gradient.hiddenBias[h], &a.first.hiddenBias[h], &a.second.hiddenBias[h], scale, o.LearningRate, 1, c1, c2); e != nil {
			return e
		}
	}
	for c := range m.outputBias {
		if e := update(&m.outputBias[c], a.gradient.outputBias[c], &a.first.outputBias[c], &a.second.outputBias[c], scale, o.LearningRate, 1, c1, c2); e != nil {
			return e
		}
	}
	return nil
}

// Fit validates all caller samples before replacing any existing weights.
// A numerical failure also leaves the previous model intact. Features are text
// only; no annotation/provenance field or development score is accepted here.
func (m *Model) Fit(samples []Sample, options FitOptions) (FitReport, error) {
	if !m.valid() || len(samples) == 0 || len(samples) > 1_000_000 {
		return FitReport{}, ErrTraining
	}
	o, e := options.normalized()
	if e != nil {
		return FitReport{}, e
	}
	var features Workspace
	for _, s := range samples {
		if _, ok := IntentIndex(s.Label); !ok {
			return FitReport{}, ErrTraining
		}
		if e := extract(s.Text, &features); e != nil {
			return FitReport{}, ErrTraining
		}
	}
	working := initialized(o.Seed)
	optimizer := &adamWorkspace{}
	order := make([]int, len(samples))
	for i := range order {
		order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(o.Seed), uint64(o.Seed)^0x9e3779b97f4a7c15))
	report := FitReport{Samples: len(samples), Epochs: o.Epochs, Initialization: Initialization, Seed: o.Seed}
	var lossSum float64
	for epoch := 0; epoch < o.Epochs; epoch++ {
		random.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for start := 0; start < len(order); start += o.BatchSize {
			end := min(start+o.BatchSize, len(order))
			optimizer.gradient = gradient{}
			for _, index := range order[start:end] {
				s := samples[index]
				if extract(s.Text, &features) != nil {
					return FitReport{}, ErrTraining
				}
				target, _ := IntentIndex(s.Label)
				loss, e := working.accumulate(&features, target, &optimizer.gradient)
				if e != nil {
					return FitReport{}, e
				}
				lossSum += loss
			}
			if e := optimizer.update(working, o, end-start, working.steps+1); e != nil {
				return FitReport{}, e
			}
			working.steps++
			report.Batches++
		}
	}
	report.TrainingSteps = working.steps
	report.MeanLoss = lossSum / (float64(len(samples)) * float64(o.Epochs))
	if !finite(report.MeanLoss) || !working.valid() {
		return FitReport{}, ErrTraining
	}
	*m = *working
	return report, nil
}
