// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintclaimtrit implements an experimental weight-only,
// mixed-precision ternary adaptation of three independent categorical claim
// heads. It is not a BitNet language model or a W1.58A8 implementation. Features,
// scales, biases and accumulators remain floating point. Outputs are attributed
// claim hints and grant no task or application state authority.
package statehintclaimtrit

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"github.com/teamswyg/laya-tools/pkg/tritpack"
)

const (
	HeadCount         = statehintclaims.HeadCount
	StateCount        = statehintclaims.StateCount
	FeatureBins       = statehintclaims.FeatureBins
	MaxTextBytes      = statehintclaims.MaxTextBytes
	FeatureSchema     = statehintclaims.FeatureSchema
	WeightTritCount   = FeatureBins * HeadCount * StateCount
	PackedWeightBytes = (WeightTritCount + 4) / 5
	// SharedStaticDecoderBytes belongs to the read-only decoder shared by all
	// models, not to each Model's fixed storage or parent-SHA backing string.
	SharedStaticDecoderBytes = decoderTableBytes
	ScaleCount               = HeadCount
	BiasCount                = HeadCount * StateCount
	ConfidenceFloor          = statehintclaims.ConfidenceFloor
	MarginFloor              = statehintclaims.MarginFloor
	DefaultSeed              = statehintclaims.DefaultSeed
	ScaleFloor               = 1e-5
)

type Head = statehintclaims.Head
type State = statehintclaims.State
type Source = statehintclaims.Source
type HeadPrediction = statehintclaims.HeadPrediction

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

func Heads() [HeadCount]Head            { return statehintclaims.Heads() }
func States() [StateCount]State         { return statehintclaims.States() }
func ParseState(s string) (State, bool) { return statehintclaims.ParseState(s) }
func StateIndex(s State) (int, bool)    { return statehintclaims.StateIndex(s) }

type Mode uint8

const (
	PTQ Mode = 1
	QAT Mode = 2
)

func (m Mode) String() string {
	switch m {
	case PTQ:
		return "ptq"
	case QAT:
		return "qat"
	default:
		return ""
	}
}

var (
	ErrInput    = statehintwide.ErrInput
	ErrModel    = errors.New("invalid ternary three-claim model")
	ErrTraining = errors.New("invalid ternary three-claim training")
	ErrArtifact = errors.New("invalid ternary three-claim artifact")
)

// Model owns only packed trits, three float32 scales, nine float32 biases, and
// provenance. The immutable parentSHAHex string is derived once when constructed
// or loaded; Clone and training-workspace copies share its 64-byte backing data.
// The string header is part of Model's struct size, while the backing data must
// be counted separately when reporting resident memory. Model retains no float
// master weights, unpacked trit array, optimizer, mutable inference cache, or
// locks. Inference is read-only; WarmFit requires exclusive receiver ownership.
// Every inference caller must own its Workspace.
type Model struct {
	packed       [PackedWeightBytes]byte
	scales       [HeadCount]float32
	bias         [HeadCount][StateCount]float32
	baseSteps    uint64
	newSteps     uint64
	parentSeed   int64
	parentSHA    [sha256.Size]byte
	parentSHAHex string
	mode         Mode
}

type Workspace struct{ features statehintwide.Workspace }

type ScoreResult struct {
	Logits        [HeadCount][StateCount]float64 `json:"logits"`
	Probabilities [HeadCount][StateCount]float64 `json:"probabilities"`
	WordCount     int                            `json:"word_count"`
}

type Prediction struct {
	Heads         [HeadCount]HeadPrediction `json:"heads"`
	Source        Source                    `json:"source"`
	TrainingSteps uint64                    `json:"training_steps"`
	Mode          string                    `json:"adaptation_mode"`
	ParentSHA256  string                    `json:"parent_sha256"`
}

// ModelMetadata describes numerical storage and provenance, not qualification.
// ParentSHA256 is verified against the canonical parent Save bytes on conversion
// and warm continuation. A digest provides identity, not publisher authenticity.
type ModelMetadata struct {
	Adaptation               string  `json:"adaptation"`
	Mode                     string  `json:"mode"`
	FeatureSchema            string  `json:"feature_schema"`
	WeightTritCount          int     `json:"weight_trit_count"`
	PackedWeightBytes        int     `json:"packed_weight_bytes"`
	SharedStaticDecoderBytes int     `json:"shared_static_decoder_bytes"`
	ScaleCount               int     `json:"scale_count"`
	BiasCount                int     `json:"bias_count"`
	PhysicalBitsPerTrit      float64 `json:"physical_bits_per_trit"`
	InformationBitsPerTrit   float64 `json:"information_bits_per_trit"`
	TrainingSteps            uint64  `json:"training_steps"`
	BaseTrainingSteps        uint64  `json:"base_training_steps"`
	NewOptimizerSteps        uint64  `json:"new_optimizer_steps"`
	ParentSHA256             string  `json:"parent_sha256"`
	ParentInitializationSeed int64   `json:"parent_initialization_seed"`
	AdaptationSeed           int64   `json:"adaptation_seed"`
	Temperature              float64 `json:"temperature"`
	ConfidenceFloor          float64 `json:"confidence_floor"`
	MarginFloor              float64 `json:"margin_floor"`
	ScaleFloor               float64 `json:"scale_floor"`
	BitNetLLM                bool    `json:"bitnet_llm"`
	W158A8                   bool    `json:"w1_58_a8"`
	QualityQualified         bool    `json:"quality_qualified"`
	StateAuthority           bool    `json:"state_authority"`
}

func (m *Model) Metadata() ModelMetadata {
	if m == nil {
		return ModelMetadata{}
	}
	return ModelMetadata{
		Adaptation: "experimental_weight_only_mixed_precision_three_claim_heads",
		Mode:       m.mode.String(), FeatureSchema: FeatureSchema,
		WeightTritCount: WeightTritCount, PackedWeightBytes: PackedWeightBytes,
		SharedStaticDecoderBytes: SharedStaticDecoderBytes,
		ScaleCount:               ScaleCount, BiasCount: BiasCount, PhysicalBitsPerTrit: 8.0 / 5,
		InformationBitsPerTrit: math.Log2(3), TrainingSteps: m.TrainingSteps(),
		BaseTrainingSteps: m.baseSteps, NewOptimizerSteps: m.newSteps,
		ParentSHA256: m.parentSHAHex, ParentInitializationSeed: m.parentSeed,
		AdaptationSeed: DefaultSeed, Temperature: 1, ConfidenceFloor: ConfidenceFloor,
		MarginFloor: MarginFloor, ScaleFloor: ScaleFloor,
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
	return m.baseSteps + m.newSteps
}
func (m *Model) InitializationSeed() int64 {
	if m == nil {
		return 0
	}
	return m.parentSeed
}
func (m *Model) Temperature() float64 { return 1 }

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// usable checks small numeric/provenance fields only. Constructors and Load
// validate the whole packed encoding once; scoring decodes addressed trits only.
func (m *Model) usable() bool {
	if m == nil || (m.mode != PTQ && m.mode != QAT) || m.newSteps > math.MaxUint64-m.baseSteps ||
		(m.mode == PTQ && m.newSteps != 0) || (m.mode == QAT && (m.newSteps == 0 || m.newSteps%TrainingEpochs != 0)) {
		return false
	}
	for _, s := range m.scales {
		if !finite(float64(s)) || s < float32(ScaleFloor) {
			return false
		}
	}
	for _, head := range m.bias {
		for _, b := range head {
			if !finite(float64(b)) {
				return false
			}
		}
	}
	return true
}
func (m *Model) valid() bool {
	return m.usable() && m.parentSHAHexMatches() && tritpack.Validate(m.packed[:], WeightTritCount) == nil
}

// parentSHAHexMatches checks the private derived-field invariant without
// allocating. Only full model validation needs this check; Predict reads the
// immutable constructor/loader result directly and performs no hex conversion.
func (m *Model) parentSHAHexMatches() bool {
	const digits = "0123456789abcdef"
	if len(m.parentSHAHex) != sha256.Size*2 {
		return false
	}
	for i, b := range m.parentSHA {
		if m.parentSHAHex[2*i] != digits[b>>4] || m.parentSHAHex[2*i+1] != digits[b&15] {
			return false
		}
	}
	return true
}

func parentDigest(parent *statehintclaims.Model) ([sha256.Size]byte, error) {
	var digest [sha256.Size]byte
	if parent == nil {
		return digest, ErrModel
	}
	h := sha256.New()
	if err := parent.Save(h); err != nil {
		return digest, ErrModel
	}
	copy(digest[:], h.Sum(nil))
	return digest, nil
}

// FromFloat performs deterministic post-training quantization with no examples
// or optimizer updates. The parent must be exclusively readable: it must not be
// undergoing Fit. Its Parameters API returns an owned value copy. parentSHA
// must identify the actual canonical parent artifact; it is not an attestation.
func FromFloat(parent *statehintclaims.Model, parentSHA string) (*Model, error) {
	if len(parentSHA) != sha256.Size*2 {
		return nil, ErrModel
	}
	raw, err := hex.DecodeString(parentSHA)
	if err != nil || len(raw) != sha256.Size {
		return nil, ErrModel
	}
	actual, err := parentDigest(parent)
	if err != nil || string(raw) != string(actual[:]) {
		return nil, ErrModel
	}
	parameters, err := parent.Parameters()
	if err != nil {
		return nil, ErrModel
	}
	m := &Model{mode: PTQ, baseSteps: parent.TrainingSteps(), parentSeed: parent.InitializationSeed(), parentSHA: actual, parentSHAHex: hex.EncodeToString(actual[:])}
	if err := m.quantize(&parameters); err != nil || !m.valid() {
		return nil, ErrModel
	}
	return m, nil
}

// quantize uses float64 means, then stores float32 per-head scales before
// rounding W/s with Go's RoundToEven and clipping to {-1,0,+1}. This is a
// deterministic adaptation, without claiming exact PyTorch bit parity.
func (m *Model) quantize(p *statehintclaims.Parameters) error {
	var sums [HeadCount]float64
	for _, feature := range p.Weights {
		for h, head := range feature {
			for _, w := range head {
				if !finite(float64(w)) {
					return ErrModel
				}
				sums[h] += math.Abs(float64(w))
			}
		}
	}
	for h, sum := range sums {
		m.scales[h] = float32(math.Max(sum/float64(FeatureBins*StateCount), ScaleFloor))
	}
	var trits [WeightTritCount]int8
	i := 0
	for _, feature := range p.Weights {
		for h, head := range feature {
			for _, w := range head {
				trits[i] = quantizedTrit(w, m.scales[h])
				i++
			}
		}
	}
	for _, head := range p.Bias {
		for _, b := range head {
			if !finite(float64(b)) {
				return ErrModel
			}
		}
	}
	m.bias = p.Bias
	if _, err := tritpack.Pack(m.packed[:], trits[:]); err != nil {
		return ErrModel
	}
	return nil
}

func quantizedTrit(w, scale float32) int8 {
	q := math.RoundToEven(float64(w) / float64(scale))
	return int8(math.Max(-1, math.Min(1, q)))
}

func (m *Model) logits(view statehintwide.ContextualFeatureView) ([HeadCount][StateCount]float64, error) {
	var result [HeadCount][StateCount]float64
	for h := range result {
		for c := range result[h] {
			result[h][c] = float64(m.bias[h][c])
		}
	}
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		for h := range result {
			for c := range result[h] {
				index := int(f.Index)*HeadCount*StateCount + h*StateCount + c
				trit := packedTritDigits[m.packed[index/5]][index%5]
				result[h][c] += float64(trit) * float64(m.scales[h]) * float64(f.Value)
			}
		}
	}
	return result, nil
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

func (m *Model) Scores(text string, workspace *Workspace) (ScoreResult, error) {
	if !m.usable() {
		return ScoreResult{}, ErrModel
	}
	if workspace == nil {
		return ScoreResult{}, ErrInput
	}
	view, err := statehintwide.ExtractContextual(text, &workspace.features)
	if err != nil {
		return ScoreResult{}, err
	}
	logits, err := m.logits(view)
	if err != nil {
		return ScoreResult{}, err
	}
	r := ScoreResult{Logits: logits, WordCount: view.WordCount()}
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
	winner, runner := StateCount-1, -1
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

// Predict retains the float parent's fixed T1, .9/.05 floors, word guard and
// explicit unknown class. PTQ retains learned attribution when parent steps are
// present, while separate mode/base/new metadata makes projection transparent.
func (m *Model) Predict(text string, workspace *Workspace) (Prediction, error) {
	scores, err := m.Scores(text, workspace)
	if err != nil {
		return Prediction{}, err
	}
	r := Prediction{Source: Learned, TrainingSteps: m.TrainingSteps(), Mode: m.mode.String(), ParentSHA256: m.parentSHAHex}
	if r.TrainingSteps == 0 {
		r.Source = Untrained
	}
	for h := range r.Heads {
		p := scores.Probabilities[h]
		if scores.WordCount == 0 {
			p = [StateCount]float64{1.0 / StateCount, 1.0 / StateCount, 1.0 / StateCount}
		}
		r.Heads[h] = headPrediction(Head(h), p, r.TrainingSteps, scores.WordCount)
	}
	return r, nil
}
