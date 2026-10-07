// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintsharedprobe

import (
	"math"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const DefaultSeed int64 = 1729

// Model stores only1024 shared weights. A shared bias cancels in softmax and is
// fixed at zero. Fit/SetTemperature require exclusive ownership; scoring an
// otherwise immutable model is lock-free with a separate workspace per caller.
type Model struct {
	weights     [FeatureDimensions]float32
	temperature float64
	steps       uint64
	seed        int64
}

func NewModel() *Model { return &Model{temperature: 1, seed: DefaultSeed} }
func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}
	v := *m
	return &v
}
func (m *Model) Temperature() float64 {
	if m == nil {
		return 0
	}
	return m.temperature
}
func (m *Model) TrainingSteps() uint64 {
	if m == nil {
		return 0
	}
	return m.steps
}
func (m *Model) InitializationSeed() int64 {
	if m == nil {
		return 0
	}
	return m.seed
}
func (m *Model) SetTemperature(t float64) error {
	if m == nil || !finite(t) || t < .05 || t > 20 {
		return ErrModel
	}
	m.temperature = t
	return nil
}
func (m *Model) valid() bool {
	if m == nil || !finite(m.temperature) || m.temperature < .05 || m.temperature > 20 {
		return false
	}
	for _, w := range m.weights {
		if !finite(float64(w)) {
			return false
		}
	}
	return true
}
func (m *Model) logits(r *Row) [8]float64 {
	var out [8]float64
	for option, v := range r {
		for feature, x := range v {
			out[option] += float64(m.weights[feature]) * float64(x)
		}
	}
	return out
}
func softmax(z [8]float64, t float64) (p [8]float64, ok bool) {
	if !finite(t) || t <= 0 {
		return p, false
	}
	maximum := z[0]
	for _, v := range z {
		if !finite(v) {
			return p, false
		}
		maximum = math.Max(maximum, v)
	}
	total := 0.0
	for i, v := range z {
		p[i] = math.Exp((v - maximum) / t)
		total += p[i]
	}
	if !finite(total) || total <= 0 {
		return p, false
	}
	for i := range p {
		p[i] /= total
	}
	return p, true
}

// ScoreFeatures accepts frozen per-option vectors, never text. TrainingSteps
// describe only this Go head's updates, not the backbone's unknown history.
func (m *Model) ScoreFeatures(r *Row, w *Workspace) (statehint.Prediction, error) {
	if !m.valid() || !validRow(r) {
		return statehint.Prediction{}, ErrModel
	}
	if w == nil {
		w = &Workspace{}
	}
	w.logits = m.logits(r)
	p, ok := softmax(w.logits, m.temperature)
	if !ok {
		return statehint.Prediction{}, ErrModel
	}
	winner, runner := 7, -1
	for i, v := range p {
		if v > p[winner] {
			winner = i
		}
	}
	for i, v := range p {
		if i != winner && (runner < 0 || v > p[runner]) {
			runner = i
		}
	}
	source := statehint.Learned
	if m.steps == 0 {
		source = statehint.Untrained
	}
	return statehint.Prediction{Intent: statehint.Intents()[winner], Probabilities: p, Confidence: p[winner], Margin: p[winner] - p[runner], Source: source, TrainingSteps: m.steps}, nil
}
