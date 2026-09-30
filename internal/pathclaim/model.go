package pathclaim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

const ModelSchema = "riido-path-cost-claim-model-v1"

// Threshold describes a separately selected development threshold, never a
// final-set or production gate. Disabled is the baseline-only candidate.
type Threshold struct {
	Disabled                                                             bool
	MinimumScore                                                         float64
	BudgetPercent, ValidationQuestions, ValidationCalls, ValidationPages int
}

// Model is the private JSON definition for a 16-coefficient FP32 storage head.
// Scoring uses float64 features and accumulation. Cost-weighted scores are not
// calibrated success probabilities. Array values are copied on assignment.
type Model struct {
	Schema, FeatureSchema, CoefficientPrecision string
	Readiness                                   Readiness
	Lambda                                      float64
	Seed                                        uint64
	Epoch                                       int
	Threshold                                   Threshold
	Coefficients                                [searchclaim.PathDimension]float32
}

func NewModel(trial Trial, readiness Readiness, threshold Threshold) (Model, error) {
	if !trial.Fitted || !trial.Attempted || trial.Failure != "" {
		return Model{}, fmt.Errorf("model requires a successful actual fit")
	}
	if trial.input != readiness {
		return Model{}, fmt.Errorf("trial input provenance mismatch")
	}
	if trial.Training.Rows != readiness.TrainingEligible || trial.Validation.Rows != readiness.ValidationEligible {
		return Model{}, fmt.Errorf("trial eligible count mismatch")
	}
	if math.IsNaN(trial.TrainingNLL) || math.IsInf(trial.TrainingNLL, 0) || math.IsNaN(trial.ValidationNLL) || math.IsInf(trial.ValidationNLL, 0) {
		return Model{}, fmt.Errorf("nonfinite fitted objective")
	}
	model := Model{Schema: ModelSchema, FeatureSchema: searchclaim.PathSchema, CoefficientPrecision: "float32", Readiness: readiness, Lambda: trial.Lambda, Seed: trial.Seed, Epoch: trial.Epoch, Threshold: threshold, Coefficients: trial.Weights}
	if err := model.Validate(); err != nil {
		return Model{}, err
	}
	return model, nil
}

func (m Model) Validate() error {
	if m.Schema != ModelSchema || m.FeatureSchema != searchclaim.PathSchema || m.CoefficientPrecision != "float32" {
		return fmt.Errorf("path model schema or precision mismatch")
	}
	if err := m.Readiness.Validate(); err != nil {
		return err
	}
	if !validPenalty(m.Lambda) || (m.Seed != 1729 && m.Seed != 2718) || m.Epoch < 1 || m.Epoch > 100 {
		return fmt.Errorf("unregistered model training metadata")
	}
	for _, v := range m.Coefficients {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return fmt.Errorf("nonfinite FP32 coefficient")
		}
	}
	t := m.Threshold
	if math.IsNaN(t.MinimumScore) || math.IsInf(t.MinimumScore, 0) || t.BudgetPercent != 90 || t.ValidationQuestions != m.Readiness.ValidationEligible || t.ValidationCalls < 0 || t.ValidationCalls > t.ValidationQuestions*90/100 || t.ValidationPages < t.ValidationQuestions || t.ValidationPages > t.ValidationQuestions*MaxPages {
		return fmt.Errorf("invalid development threshold metadata")
	}
	if t.Disabled && t.ValidationCalls != 0 {
		return fmt.Errorf("disabled threshold has helper calls")
	}
	return nil
}

// UnmarshalJSON rejects truncated/expanded coefficient arrays instead of letting
// the standard fixed-array decoder silently fill or discard coefficients.
func (m *Model) UnmarshalJSON(b []byte) error {
	if len(b) > 16<<10 {
		return fmt.Errorf("path model byte bound")
	}
	var wire struct {
		Schema, FeatureSchema, CoefficientPrecision string
		Readiness                                   json.RawMessage
		Lambda                                      float64
		Seed                                        uint64
		Epoch                                       int
		Threshold                                   Threshold
		Coefficients                                []float32
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&wire); err != nil {
		return fmt.Errorf("invalid path model JSON")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("extra path model content")
	}
	if len(wire.Coefficients) != searchclaim.PathDimension {
		return fmt.Errorf("path coefficient count mismatch")
	}
	var shape struct{ ValidationRepositories []json.RawMessage }
	if err := json.Unmarshal(wire.Readiness, &shape); err != nil || len(shape.ValidationRepositories) != 5 {
		return fmt.Errorf("input seal repository count mismatch")
	}
	var next Model
	next.Schema, next.FeatureSchema, next.CoefficientPrecision = wire.Schema, wire.FeatureSchema, wire.CoefficientPrecision
	seal := json.NewDecoder(bytes.NewReader(wire.Readiness))
	seal.DisallowUnknownFields()
	if err := seal.Decode(&next.Readiness); err != nil {
		return fmt.Errorf("invalid input seal JSON")
	}
	next.Lambda, next.Seed, next.Epoch, next.Threshold = wire.Lambda, wire.Seed, wire.Epoch, wire.Threshold
	copy(next.Coefficients[:], wire.Coefficients)
	if err := next.Validate(); err != nil {
		return err
	}
	*m = next
	return nil
}

// DecodeModel bounds the complete file, including leading whitespace. Artifact
// readers should use it instead of an unbounded outer json.Unmarshal, which
// trims whitespace before invoking UnmarshalJSON.
func DecodeModel(b []byte) (Model, error) {
	var m Model
	if len(b) > 16<<10 {
		return m, fmt.Errorf("path model byte bound")
	}
	if err := m.UnmarshalJSON(b); err != nil {
		return Model{}, err
	}
	return m, nil
}

func (m Model) score(features [searchclaim.PathDimension]float64) (float64, error) {
	if err := validateFeatures(features); err != nil {
		return 0, err
	}
	var score float64
	for i, value := range features {
		score += float64(m.Coefficients[i]) * value
	}
	return score, nil
}

func (m Model) Score(features [searchclaim.PathDimension]float64) (float64, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	}
	return m.score(features)
}

// Selector validates once and snapshots every coefficient/configuration by
// value. Later mutation of the caller's Model cannot change the selector.
func (m Model) Selector() (fileeval.Selector, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return func(features [searchclaim.PathDimension]float64) (bool, error) {
		score, err := m.score(features)
		if err != nil {
			return false, err
		}
		return !m.Threshold.Disabled && score >= m.Threshold.MinimumScore, nil
	}, nil
}
