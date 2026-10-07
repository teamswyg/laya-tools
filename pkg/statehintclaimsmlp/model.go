// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintclaimsmlp is a source-only shared-ReLU claim experiment.
// Its three categorical outputs are claim hints and have no state authority.
package statehintclaimsmlp

import (
	"errors"
	"math"
	"math/rand/v2"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	FeatureBins     = statehintwide.FeatureBins
	MaxTextBytes    = statehintwide.MaxTextBytes
	FeatureSchema   = statehintwide.ContextualFeatureSchema
	HiddenUnits     = 16
	HeadCount       = statehintclaims.HeadCount
	StateCount      = statehintclaims.StateCount
	WeightCount     = FeatureBins*HiddenUnits + HiddenUnits*HeadCount*StateCount
	ParameterCount  = WeightCount + HiddenUnits + HeadCount*StateCount
	ConfidenceFloor = statehintclaims.ConfidenceFloor
	MarginFloor     = statehintclaims.MarginFloor
	DefaultSeed     = statehintclaims.DefaultSeed
	Initialization  = "fresh-glorot-uniform-pcg-separate-shuffle-rcm-v1"
	Architecture    = "2048-contextual->16-shared-relu->3-independent-3-state-categorical-heads"
	initStream      = uint64(0x696e697472636d31)
)

type Head = statehintclaims.Head
type State = statehintclaims.State
type Source = statehintclaims.Source
type HeadPrediction = statehintclaims.HeadPrediction
type Prediction = statehintclaims.Prediction
type ScoreResult = statehintclaims.ScoreResult

const (
	ResponseRequested      = statehintclaims.ResponseRequested
	CurrentActivityClaimed = statehintclaims.CurrentActivityClaimed
	CompletionClaimed      = statehintclaims.CompletionClaimed
	True                   = statehintclaims.True
	False                  = statehintclaims.False
	Unknown                = statehintclaims.Unknown
	Learned                = statehintclaims.Learned
	Untrained              = statehintclaims.Untrained
)

var (
	ErrInput    = statehintwide.ErrInput
	ErrModel    = errors.New("invalid shared-ReLU three-claim model")
	ErrTraining = errors.New("invalid shared-ReLU three-claim training")
	ErrArtifact = errors.New("invalid shared-ReLU three-claim artifact")
)

func Heads() [HeadCount]Head            { return statehintclaims.Heads() }
func States() [StateCount]State         { return statehintclaims.States() }
func ParseState(s string) (State, bool) { return statehintclaims.ParseState(s) }
func StateIndex(s State) (int, bool)    { return statehintclaims.StateIndex(s) }
func finite(v float64) bool             { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Model owns float32 parameters. Scores/Predict only read it. Fit requires
// exclusive ownership; concurrent inference requires separate Workspaces.
// The shared representation feeds three separately normalized heads, not one
// nine-way softmax. There are no generative outputs, caches, maps or locks.
type Model struct {
	input      [FeatureBins][HiddenUnits]float32
	hiddenBias [HiddenUnits]float32
	output     [HiddenUnits][HeadCount][StateCount]float32
	outputBias [HeadCount][StateCount]float32
	steps      uint64
	seed       int64
	samples    uint32
}

// Workspace is caller-owned and must be supplied. The borrowed contextual
// features and float64 activations are reused without a hot-path allocation.
type Workspace struct {
	features              statehintwide.Workspace
	preactivation, hidden [HiddenUnits]float64
}

// Metadata describes the numerical contract, never an evaluation or authority
// grant. Qualified and StateAuthority are always false in this experiment.
type Metadata struct {
	Architecture       string  `json:"architecture"`
	FeatureSchema      string  `json:"feature_schema"`
	FeatureBins        int     `json:"feature_bins"`
	HiddenUnits        int     `json:"hidden_units"`
	Heads              int     `json:"heads"`
	StatesPerHead      int     `json:"states_per_head"`
	ParameterFloats    int     `json:"parameter_floats"`
	WeightFloats       int     `json:"weight_floats"`
	ParameterPrecision string  `json:"parameter_precision"`
	Accumulator        string  `json:"accumulator_precision"`
	Initialization     string  `json:"initialization"`
	InitializationSeed int64   `json:"initialization_seed"`
	TrainingSteps      uint64  `json:"training_steps"`
	TrainingSamples    uint32  `json:"training_samples"`
	ConfidenceFloor    float64 `json:"confidence_floor"`
	MarginFloor        float64 `json:"margin_floor"`
	Temperature        float64 `json:"temperature"`
	Qualified          bool    `json:"qualified"`
	StateAuthority     bool    `json:"state_authority"`
}

func (m *Model) Metadata() Metadata {
	r := Metadata{Architecture: Architecture, FeatureSchema: FeatureSchema, FeatureBins: FeatureBins, HiddenUnits: HiddenUnits, Heads: HeadCount, StatesPerHead: StateCount, ParameterFloats: ParameterCount, WeightFloats: WeightCount, ParameterPrecision: "float32", Accumulator: "float64", Initialization: Initialization, ConfidenceFloor: ConfidenceFloor, MarginFloor: MarginFloor, Temperature: 1}
	if m != nil {
		r.InitializationSeed, r.TrainingSteps, r.TrainingSamples = m.seed, m.steps, m.samples
	}
	return r
}

func NewModel() *Model {
	m := &Model{}
	m.initialize()
	return m
}

func (m *Model) initialize() {
	*m = Model{seed: DefaultSeed}
	random := rand.New(rand.NewPCG(uint64(DefaultSeed), uint64(DefaultSeed)^initStream))
	inputLimit := math.Sqrt(6 / float64(FeatureBins+HiddenUnits))
	// Each independent head is a 16-to-3 affine layer.
	outputLimit := math.Sqrt(6 / float64(HiddenUnits+StateCount))
	for i := range m.input {
		for h := range m.input[i] {
			m.input[i][h] = float32((2*random.Float64() - 1) * inputLimit)
		}
	}
	for h := range m.output {
		for head := range m.output[h] {
			for c := range m.output[h][head] {
				m.output[h][head][c] = float32((2*random.Float64() - 1) * outputLimit)
			}
		}
	}
}

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

func validHistory(steps uint64, samples uint32, seed int64) bool {
	if seed != DefaultSeed || samples > maxSamples {
		return false
	}
	if samples == 0 {
		return steps == 0
	}
	return steps == uint64(FitEpochs)*((uint64(samples)+FitBatchSize-1)/FitBatchSize)
}

func (m *Model) valid() bool {
	if m == nil || !validHistory(m.steps, m.samples, m.seed) {
		return false
	}
	for _, row := range m.input {
		for _, value := range row {
			if !finite(float64(value)) {
				return false
			}
		}
	}
	for _, value := range m.hiddenBias {
		if !finite(float64(value)) {
			return false
		}
	}
	for _, row := range m.output {
		for _, head := range row {
			for _, value := range head {
				if !finite(float64(value)) {
					return false
				}
			}
		}
	}
	for _, head := range m.outputBias {
		for _, value := range head {
			if !finite(float64(value)) {
				return false
			}
		}
	}
	return true
}

func (m *Model) logits(view statehintwide.ContextualFeatureView, w *Workspace) ([HeadCount][StateCount]float64, bool) {
	for h, b := range m.hiddenBias {
		w.preactivation[h] = float64(b)
	}
	for i := 0; i < view.Len(); i++ {
		feature := view.At(i)
		value := float64(feature.Value)
		for h, weight := range m.input[feature.Index] {
			w.preactivation[h] += value * float64(weight)
		}
	}
	for h, z := range w.preactivation {
		if !finite(z) {
			return [HeadCount][StateCount]float64{}, false
		}
		w.hidden[h] = math.Max(0, z)
	}
	var result [HeadCount][StateCount]float64
	for head := range result {
		for c := range result[head] {
			result[head][c] = float64(m.outputBias[head][c])
		}
	}
	for h, activation := range w.hidden {
		for head := range result {
			for c := range result[head] {
				result[head][c] += activation * float64(m.output[h][head][c])
			}
		}
	}
	return result, true
}

func softmax(logits [StateCount]float64) ([StateCount]float64, bool) {
	var p [StateCount]float64
	maximum := logits[0]
	for _, value := range logits {
		if !finite(value) {
			return p, false
		}
		maximum = math.Max(maximum, value)
	}
	var total float64
	for c, value := range logits {
		p[c] = math.Exp(value - maximum)
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

// Scores uses the exact existing contextual sparse text extractor, float64
// accumulation and a separate three-way T=1 softmax for each claim head.
func (m *Model) Scores(text string, workspace *Workspace) (ScoreResult, error) {
	if m == nil || !validHistory(m.steps, m.samples, m.seed) {
		return ScoreResult{}, ErrModel
	}
	if workspace == nil {
		return ScoreResult{}, ErrInput
	}
	view, err := statehintwide.ExtractContextual(text, &workspace.features)
	if err != nil {
		return ScoreResult{}, err
	}
	logits, ok := m.logits(view, workspace)
	if !ok {
		return ScoreResult{}, ErrModel
	}
	r := ScoreResult{Logits: logits, WordCount: view.WordCount()}
	for head := range r.Probabilities {
		p, ok := softmax(r.Logits[head])
		if !ok {
			return ScoreResult{}, ErrModel
		}
		r.Probabilities[head] = p
	}
	return r, nil
}

// Keep the claim package's exact tie rule and gate precedence. Semantic
// unknown is a supervised class; untrained/no-word/numeric abstention are
// separately named output reasons. These gates confer no state authority.
func headPrediction(head Head, p [StateCount]float64, steps uint64, words int) HeadPrediction {
	winner, runner := StateCount-1, -1
	for c, value := range p {
		if value > p[winner] {
			winner = c
		}
	}
	for c := range p {
		if c != winner && (runner < 0 || p[c] > p[runner]) {
			runner = c
		}
	}
	r := HeadPrediction{Head: head.String(), Winner: States()[winner], State: States()[winner], Probabilities: p, Confidence: p[winner], Margin: p[winner] - p[runner]}
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

func (m *Model) Predict(text string, workspace *Workspace) (Prediction, error) {
	scores, err := m.Scores(text, workspace)
	if err != nil {
		return Prediction{}, err
	}
	r := Prediction{Source: Learned, TrainingSteps: m.steps}
	if m.steps == 0 {
		r.Source = Untrained
	}
	for head := range r.Heads {
		p := scores.Probabilities[head]
		if scores.WordCount == 0 {
			p = [StateCount]float64{1.0 / StateCount, 1.0 / StateCount, 1.0 / StateCount}
		}
		r.Heads[head] = headPrediction(Head(head), p, m.steps, scores.WordCount)
	}
	return r, nil
}
