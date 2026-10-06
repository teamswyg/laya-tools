// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintmlp

import (
	"math"
	"math/rand/v2"
)

const (
	DefaultSeed    int64 = 1729
	Initialization       = "glorot-uniform-pcg-separate-shuffle-v3"
)

// Model is 2048 contextual inputs ->16 ReLU units ->8 softmax columns. Its
// arrays are exclusively mutated by Fit/SetTemperature, never by Predict.
// Fit/SetTemperature require exclusive ownership. Concurrent Predict calls
// need separate workspaces and an otherwise immutable model.
type Model struct {
	input       [FeatureBins][HiddenUnits]float32
	hiddenBias  [HiddenUnits]float32
	output      [HiddenUnits][IntentCount]float32
	outputBias  [IntentCount]float32
	temperature float64
	steps       uint64
	seed        int64
}

func NewModel() *Model { return initialized(DefaultSeed) }
func initialized(seed int64) *Model {
	m := &Model{temperature: 1, seed: seed}
	random := rand.New(rand.NewPCG(uint64(seed), uint64(seed)^0x696e69746d6c7033))
	inputLimit := math.Sqrt(6 / float64(FeatureBins+HiddenUnits))
	outputLimit := math.Sqrt(6 / float64(HiddenUnits+IntentCount))
	for i := range m.input {
		for h := range m.input[i] {
			m.input[i][h] = float32((2*random.Float64() - 1) * inputLimit)
		}
	}
	for h := range m.output {
		for c := range m.output[h] {
			m.output[h][c] = float32((2*random.Float64() - 1) * outputLimit)
		}
	}
	return m
}
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
	for _, row := range m.input {
		for _, v := range row {
			if !finite(float64(v)) {
				return false
			}
		}
	}
	for _, row := range m.output {
		for _, v := range row {
			if !finite(float64(v)) {
				return false
			}
		}
	}
	for _, v := range m.hiddenBias {
		if !finite(float64(v)) {
			return false
		}
	}
	for _, v := range m.outputBias {
		if !finite(float64(v)) {
			return false
		}
	}
	return true
}
func (m *Model) logits(w *Workspace) [IntentCount]float64 {
	for h, b := range m.hiddenBias {
		w.preactivation[h] = float64(b)
	}
	for _, index := range w.indices[:w.count] {
		value := float64(w.values[index])
		for h, weight := range m.input[index] {
			w.preactivation[h] += value * float64(weight)
		}
	}
	for h, z := range w.preactivation {
		w.hidden[h] = math.Max(0, z)
	}
	var logits [IntentCount]float64
	for c, b := range m.outputBias {
		logits[c] = float64(b)
	}
	for h, activation := range w.hidden {
		for c, weight := range m.output[h] {
			logits[c] += activation * float64(weight)
		}
	}
	return logits
}
func softmax(logits [IntentCount]float64, t float64) ([IntentCount]float64, bool) {
	var p [IntentCount]float64
	if !finite(t) || t < .05 || t > 20 {
		return p, false
	}
	maxLogit := logits[0]
	for _, v := range logits {
		if !finite(v) {
			return p, false
		}
		maxLogit = math.Max(maxLogit, v)
	}
	var total float64
	for c, v := range logits {
		p[c] = math.Exp((v - maxLogit) / t)
		total += p[c]
	}
	if !finite(total) || total <= 0 {
		return p, false
	}
	for c := range p {
		p[c] /= total
	}
	return p, true
}
func prediction(p [IntentCount]float64, source Source, steps uint64) Prediction {
	winner, runner := IntentCount-1, -1
	for c, v := range p {
		if v > p[winner] {
			winner = c
		}
	}
	for c := range p {
		if c != winner && (runner < 0 || p[c] > p[runner]) {
			runner = c
		}
	}
	return Prediction{Intent: Intents()[winner], Probabilities: p, Confidence: p[winner], Margin: p[winner] - p[runner], Source: source, TrainingSteps: steps}
}
func (m *Model) Predict(text string, w *Workspace) (Prediction, error) {
	if m == nil || !finite(m.temperature) || m.temperature < .05 || m.temperature > 20 {
		return Prediction{}, ErrModel
	}
	if w == nil {
		w = &Workspace{}
	}
	if e := extract(text, w); e != nil {
		return Prediction{}, e
	}
	source := Learned
	if m.steps == 0 {
		source = Untrained
	}
	if w.wordCount == 0 {
		var p [IntentCount]float64
		for c := range p {
			p[c] = 1.0 / IntentCount
		}
		v := prediction(p, source, m.steps)
		v.GuardReason = "no_word_content"
		return v, nil
	}
	p, ok := softmax(m.logits(w), m.temperature)
	if !ok {
		return Prediction{}, ErrModel
	}
	return prediction(p, source, m.steps), nil
}
