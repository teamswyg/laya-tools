// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintclaims implements three independent categorical claim heads.
// Outputs are attributed claim hints, never verified task or application state.
package statehintclaims

import (
	"errors"
	"math"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	HeadCount       = 3
	StateCount      = 3
	FeatureBins     = statehintwide.FeatureBins
	MaxTextBytes    = statehintwide.MaxTextBytes
	FeatureSchema   = statehintwide.ContextualFeatureSchema
	ParameterCount  = FeatureBins*HeadCount*StateCount + HeadCount*StateCount
	ConfidenceFloor = .9
	MarginFloor     = .05
	DefaultSeed     = 1729
)

type Head uint8

const (
	ResponseRequested Head = iota
	CurrentActivityClaimed
	CompletionClaimed
)

func Heads() [HeadCount]Head {
	return [HeadCount]Head{ResponseRequested, CurrentActivityClaimed, CompletionClaimed}
}
func (h Head) String() string {
	switch h {
	case ResponseRequested:
		return "response_requested"
	case CurrentActivityClaimed:
		return "current_activity_claimed"
	case CompletionClaimed:
		return "completion_claimed"
	default:
		return ""
	}
}

// State is a supervised class. Unknown is a real categorical target, not a
// masked example. Output abstention is distinguished by UnknownReason.
type State string

const (
	True    State = "true"
	False   State = "false"
	Unknown State = "unknown"
)

func States() [StateCount]State { return [StateCount]State{True, False, Unknown} }
func (s State) String() string  { return string(s) }
func ParseState(s string) (State, bool) {
	state := State(s)
	_, ok := StateIndex(state)
	if !ok {
		return "", false
	}
	return state, true
}
func StateIndex(s State) (int, bool) {
	switch s {
	case True:
		return 0, true
	case False:
		return 1, true
	case Unknown:
		return 2, true
	default:
		return 0, false
	}
}

type Source string

const (
	Learned   Source = "learned_supervised_claims"
	Untrained Source = "untrained_claims"
)

var (
	ErrInput    = statehintwide.ErrInput
	ErrModel    = errors.New("invalid three-claim model")
	ErrTraining = errors.New("invalid three-claim training")
	ErrArtifact = errors.New("invalid three-claim artifact")
)

// Model stores nine contiguous columns per feature. Active features are read
// once per dot product. Inference is read-only; Fit requires exclusive model
// ownership. This layout makes no SIMD or performance claim.
type Model struct {
	weights [FeatureBins][HeadCount][StateCount]float32
	bias    [HeadCount][StateCount]float32
	steps   uint64
	seed    int64
}

// Workspace belongs to one inference caller, with no locks or global cache.
type Workspace struct{ features statehintwide.Workspace }

type ScoreResult struct {
	Logits        [HeadCount][StateCount]float64 `json:"logits"`
	Probabilities [HeadCount][StateCount]float64 `json:"probabilities"`
	WordCount     int                            `json:"word_count"`
}
type HeadPrediction struct {
	Head          string              `json:"head"`
	Winner        State               `json:"winner"`
	State         State               `json:"state"`
	Probabilities [StateCount]float64 `json:"probabilities"`
	Confidence    float64             `json:"confidence"`
	Margin        float64             `json:"margin"`
	UnknownReason string              `json:"unknown_reason,omitempty"`
}
type Prediction struct {
	Heads         [HeadCount]HeadPrediction `json:"heads"`
	Source        Source                    `json:"source"`
	TrainingSteps uint64                    `json:"training_steps"`
}

func NewModel() *Model { return &Model{seed: DefaultSeed} }
func (m *Model) Clone() *Model {
	if m == nil {
		return nil
	}
	c := *m
	return &c
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
func (m *Model) Temperature() float64 { return 1 }
func finite(v float64) bool           { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (m *Model) valid() bool {
	if m == nil {
		return false
	}
	for _, feature := range m.weights {
		for _, head := range feature {
			for _, v := range head {
				if !finite(float64(v)) {
					return false
				}
			}
		}
	}
	for _, head := range m.bias {
		for _, v := range head {
			if !finite(float64(v)) {
				return false
			}
		}
	}
	return true
}
func (m *Model) logits(view statehintwide.ContextualFeatureView) [HeadCount][StateCount]float64 {
	var result [HeadCount][StateCount]float64
	for h := range result {
		for c := range result[h] {
			result[h][c] = float64(m.bias[h][c])
		}
	}
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		v := float64(f.Value)
		for h := range result {
			for c := range result[h] {
				result[h][c] += float64(m.weights[f.Index][h][c]) * v
			}
		}
	}
	return result
}
func softmax(logits [StateCount]float64) ([StateCount]float64, bool) {
	var p [StateCount]float64
	maxValue := logits[0]
	for _, v := range logits {
		if !finite(v) {
			return p, false
		}
		maxValue = math.Max(maxValue, v)
	}
	var sum float64
	for c, v := range logits {
		p[c] = math.Exp(v - maxValue)
		sum += p[c]
	}
	if !finite(sum) || sum <= 0 {
		return p, false
	}
	for c := range p {
		p[c] /= sum
	}
	return p, true
}

// Scores and Predict accept text only. Annotation, provenance, event facts and
// task state never enter these numeric features. Temperature is fixed at one.
func (m *Model) Scores(text string, workspace *Workspace) (ScoreResult, error) {
	if m == nil {
		return ScoreResult{}, ErrModel
	}
	if workspace == nil {
		return ScoreResult{}, ErrInput
	}
	view, err := statehintwide.ExtractContextual(text, &workspace.features)
	if err != nil {
		return ScoreResult{}, err
	}
	r := ScoreResult{Logits: m.logits(view), WordCount: view.WordCount()}
	for h := range r.Probabilities {
		p, ok := softmax(r.Logits[h])
		if !ok {
			return ScoreResult{}, ErrModel
		}
		r.Probabilities[h] = p
	}
	return r, nil
}

func headPrediction(h Head, p [StateCount]float64, steps uint64, words int) HeadPrediction {
	winner, runner := StateCount-1, -1 // Exact ties prefer the unknown class.
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
	r := HeadPrediction{Head: h.String(), Winner: States()[winner], State: States()[winner], Probabilities: p, Confidence: p[winner], Margin: p[winner] - p[runner]}
	switch {
	case words == 0:
		r.UnknownReason = "no_word_content"
	case steps == 0:
		r.UnknownReason = "untrained"
	case r.Winner == Unknown:
		r.UnknownReason = "semantic_unknown"
	case r.Confidence < ConfidenceFloor:
		r.UnknownReason = "low_confidence"
	case r.Margin < MarginFloor:
		r.UnknownReason = "low_margin"
	}
	if r.UnknownReason != "" {
		r.State = Unknown
	}
	return r
}

// An emitted true/false requires the fixed .9/.05 numeric floors, trained
// weights and word content. This mechanism establishes no quality qualification
// and grants no application/state authority, including for completion claims.
func (m *Model) Predict(text string, workspace *Workspace) (Prediction, error) {
	scores, err := m.Scores(text, workspace)
	if err != nil {
		return Prediction{}, err
	}
	r := Prediction{Source: Learned, TrainingSteps: m.steps}
	if m.steps == 0 {
		r.Source = Untrained
	}
	for h := range r.Heads {
		p := scores.Probabilities[h]
		if scores.WordCount == 0 {
			p = [StateCount]float64{1.0 / StateCount, 1.0 / StateCount, 1.0 / StateCount}
		}
		r.Heads[h] = headPrediction(Head(h), p, m.steps, scores.WordCount)
	}
	return r, nil
}
