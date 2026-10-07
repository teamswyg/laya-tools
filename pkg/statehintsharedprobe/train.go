// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintsharedprobe

import (
	"math"
	"math/rand/v2"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

type Sample struct {
	RowIndex int
	Label    statehint.Intent
}
type FitOptions struct {
	Epochs, BatchSize         int
	LearningRate, WeightDecay float64
	Seed                      int64
}
type FitReport struct {
	Samples       int     `json:"samples"`
	Epochs        int     `json:"epochs"`
	Batches       int     `json:"batches"`
	TrainingSteps uint64  `json:"new_go_head_training_steps"`
	MeanLoss      float64 `json:"mean_loss"`
	Seed          int64   `json:"seed"`
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

// accumulate differentiates one eight-choice CE: sum_i(p_i-y_i)*phi_i.
// No class-specific weights, trainable bias, labels in features or temperature
// adjustment is involved. Stable subtraction preserves log loss at large logits.
func (m *Model) accumulate(row *Row, target int, g *[FeatureDimensions]float64) (float64, error) {
	if target < 0 || target >= 8 || !validRow(row) {
		return 0, ErrTraining
	}
	z := m.logits(row)
	p, ok := softmax(z, 1)
	if !ok {
		return 0, ErrTraining
	}
	maximum := z[0]
	for _, v := range z {
		maximum = math.Max(maximum, v)
	}
	total := 0.0
	for _, v := range z {
		total += math.Exp(v - maximum)
	}
	loss := (maximum - z[target]) + math.Log(total)
	if !finite(loss) {
		return 0, ErrTraining
	}
	for option, v := range row {
		delta := p[option]
		if option == target {
			delta--
		}
		for feature, x := range v {
			g[feature] += delta * float64(x)
		}
	}
	return loss, nil
}

type optimizer struct{ gradient, first, second [FeatureDimensions]float64 }

func (a *optimizer) update(m *Model, o FitOptions, batch int, step uint64) error {
	const b1, b2, epsilon = .9, .999, 1e-8
	scale := 1 / float64(batch)
	c1, c2 := 1-math.Pow(b1, float64(step)), 1-math.Pow(b2, float64(step))
	decay := 1 - o.LearningRate*o.WeightDecay
	for i, g := range a.gradient {
		g *= scale
		a.first[i] = b1*a.first[i] + (1-b1)*g
		a.second[i] = b2*a.second[i] + (1-b2)*g*g
		v := float64(m.weights[i])*decay - o.LearningRate*(a.first[i]/c1)/(math.Sqrt(a.second[i]/c2)+epsilon)
		if !finite(v) || math.Abs(v) > math.MaxFloat32 {
			return ErrTraining
		}
		m.weights[i] = float32(v)
	}
	return nil
}

// Fit streams one32KiB feature row per sample. It starts from fresh zeros and
// fresh AdamW every time. Read/numeric errors leave the previous model unchanged.
func (m *Model) Fit(store *Store, samples []Sample, options FitOptions) (FitReport, error) {
	if !m.valid() || store == nil || len(samples) == 0 || len(samples) > MaxRows {
		return FitReport{}, ErrTraining
	}
	o, e := options.normalized()
	if e != nil {
		return FitReport{}, e
	}
	var w Workspace
	for _, s := range samples {
		if _, ok := statehint.IntentIndex(s.Label); !ok {
			return FitReport{}, ErrTraining
		}
		if store.ReadRow(s.RowIndex, &w) != nil {
			return FitReport{}, ErrTraining
		}
	}
	working := NewModel()
	working.seed = o.Seed
	var a optimizer
	order := make([]int, len(samples))
	for i := range order {
		order[i] = i
	}
	random := rand.New(rand.NewPCG(uint64(o.Seed), uint64(o.Seed)^0x9e3779b97f4a7c15))
	r := FitReport{Samples: len(samples), Epochs: o.Epochs, Seed: o.Seed}
	sum := 0.0
	for epoch := 0; epoch < o.Epochs; epoch++ {
		random.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
		for start := 0; start < len(order); start += o.BatchSize {
			end := min(start+o.BatchSize, len(order))
			a.gradient = [FeatureDimensions]float64{}
			for _, index := range order[start:end] {
				s := samples[index]
				if store.ReadRow(s.RowIndex, &w) != nil {
					return FitReport{}, ErrTraining
				}
				target, _ := statehint.IntentIndex(s.Label)
				loss, e := working.accumulate(&w.Row, target, &a.gradient)
				if e != nil {
					return FitReport{}, e
				}
				sum += loss
			}
			if a.update(working, o, end-start, working.steps+1) != nil {
				return FitReport{}, ErrTraining
			}
			working.steps++
			r.Batches++
		}
	}
	r.TrainingSteps = working.steps
	r.MeanLoss = sum / (float64(len(samples)) * float64(o.Epochs))
	if !finite(r.MeanLoss) || !working.valid() {
		return FitReport{}, ErrTraining
	}
	*m = *working
	return r, nil
}
