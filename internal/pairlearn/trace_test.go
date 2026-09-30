package pairlearn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func cloneDataset(d Dataset) Dataset {
	d.Offsets = slices.Clone(d.Offsets)
	d.Indices = slices.Clone(d.Indices)
	d.Values = slices.Clone(d.Values)
	d.Labels = slices.Clone(d.Labels)
	d.Groups = slices.Clone(d.Groups)
	d.SampleWeights = slices.Clone(d.SampleWeights)
	return d
}

func TestTraceExactFitParityAndReplay(t *testing.T) {
	for _, mode := range []string{"fp32", "ternary_ste"} {
		for _, seed := range []uint64{1729, 2718} {
			for _, weights := range []string{"nil", "unit", "weighted"} {
				t.Run(fmt.Sprintf("%s_%s_%d", mode, weights, seed), func(t *testing.T) {
					train, validation := weightedFixture("development", 0), weightedFixture("validation", 10)
					if weights == "unit" {
						train.SampleWeights = []float64{1, 1, 1, 1}
						validation.SampleWeights = []float64{1, 1, 1, 1}
					} else if weights == "weighted" {
						train.SampleWeights = []float64{0, .25, 1, 3}
						validation.SampleWeights = []float64{2, 0, .5, 1}
					}
					beforeTrain, beforeValidation := cloneDataset(train), cloneDataset(validation)
					c := Config{Seed: seed, Mode: mode, LearningRate: .1, L2: .0001, Epochs: 30, Batch: 3, Dimension: 1}
					plain, err := Fit(train, validation, c)
					if err != nil {
						t.Fatal(err)
					}
					got, trace, err := FitWithTrace(train, validation, c)
					if err != nil || !reflect.DeepEqual(got, plain) || !trace.Complete || trace.Schema != TraceSchema || trace.Mode != mode || trace.SelectedEpoch != plain.Epoch || len(trace.Epochs) != c.Epochs || trace.FailureKind != "" {
						t.Fatal("trace changed fit", got, plain, trace, err)
					}
					if !reflect.DeepEqual(train, beforeTrain) || !reflect.DeepEqual(validation, beforeValidation) {
						t.Fatal("trace mutated input columns")
					}
					bestNLL, bestEpoch := math.Inf(1), 0
					for i, epoch := range trace.Epochs {
						newBest := epoch.ValidationNLL < bestNLL
						if newBest {
							bestNLL, bestEpoch = epoch.ValidationNLL, epoch.Epoch
						}
						if epoch.Epoch != i+1 || !finite(epoch.TrainingNLL) || !finite(epoch.ValidationNLL) || epoch.IsNewBest != newBest || epoch.BestEpochSoFar != bestEpoch {
							t.Fatal("invalid epoch or best-so-far observation", epoch)
						}
						if epoch.Epoch == plain.Epoch && (epoch.TrainingNLL != plain.TrainingNLL || epoch.ValidationNLL != plain.ValidationNLL) {
							t.Fatal("selected losses differ from original result", epoch, plain)
						}
					}
					if bestEpoch != plain.Epoch {
						t.Fatal("trace did not retain the selected earliest minimum")
					}
					wantTraining, wantValidation := WeightSummary{Rows: 4, SampleWeightSum: 4}, WeightSummary{Rows: 4, SampleWeightSum: 4}
					if weights == "weighted" {
						wantTraining.ZeroWeightRows, wantTraining.SampleWeightSum = 1, 4.25
						wantValidation.ZeroWeightRows, wantValidation.SampleWeightSum = 1, 3.5
					}
					if trace.Training != wantTraining || trace.Validation != wantValidation {
						t.Fatal("weight sums replaced row coverage", trace.Training, trace.Validation)
					}
					if !bytes.Contains([]byte(trace.Precision), []byte("float32")) || !bytes.Contains([]byte(trace.Precision), []byte("float64")) {
						t.Fatal("precision convention missing", trace.Precision)
					}
					replay, repeated, err := FitWithTrace(train, validation, c)
					if err != nil || !reflect.DeepEqual(got, replay) || !reflect.DeepEqual(trace, repeated) {
						t.Fatal("trace not reproducible", err)
					}
					firstJSON, err := json.Marshal(trace)
					if err != nil {
						t.Fatal(err)
					}
					secondJSON, err := json.Marshal(repeated)
					if err != nil || !bytes.Equal(firstJSON, secondJSON) {
						t.Fatal("trace JSON changed on replay", err)
					}
					for _, forbidden := range []string{`"Weights":`, `"Values":`, `"Indices":`, `"Groups":`, `"IDs":`} {
						if bytes.Contains(firstJSON, []byte(forbidden)) {
							t.Fatal("trace exposed source or model columns", forbidden)
						}
					}
					plainJSON, err := json.Marshal(plain)
					if err != nil {
						t.Fatal(err)
					}
					tracedJSON, err := json.Marshal(got)
					if err != nil || !bytes.Equal(plainJSON, tracedJSON) || bytes.Contains(tracedJSON, []byte(`"Trace"`)) {
						t.Fatal("legacy result JSON changed", err)
					}
				})
			}
		}
	}
}

func TestTraceOneEpochWeightedLossAgainstIndependentCalculation(t *testing.T) {
	train := Dataset{Split: "development", Offsets: []int{0, 1, 2}, Indices: []uint16{0, 0}, Values: []float64{-1, 2}, Labels: []float64{0, 1}, Groups: []int{1, 2}, SampleWeights: []float64{2, .5}}
	validation := Dataset{Split: "validation", Offsets: []int{0, 1, 2}, Indices: []uint16{0, 0}, Values: []float64{.25, 2}, Labels: []float64{0, 1}, Groups: []int{3, 4}, SampleWeights: []float64{0, 3}}
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 1, Batch: 2, Dimension: 1}
	rng := rand.New(rand.NewPCG(c.Seed, c.Seed+1))
	shadow := rng.NormFloat64() * .01
	coefficient := float64(float32(shadow))
	p0, p1 := 1/(1+math.Exp(coefficient)), 1/(1+math.Exp(-2*coefficient))
	grad := p0*(2*2/2.5)*-1 + (p1-1)*(.5*2/2.5)*2
	expected := float64(float32(shadow - c.LearningRate*(grad/2+c.L2*shadow)))
	got, trace, err := FitWithTrace(train, validation, c)
	if err != nil || len(trace.Epochs) != 1 || got.Weights[0] != expected {
		t.Fatal("weighted batch or precision changed", got, trace, err, expected)
	}
	// The moderate one-dimensional logits allow a separate softplus expression;
	// no production loss or NLL helper is used to calculate these expectations.
	wantTrain := (2*math.Log1p(math.Exp(-expected)) + .5*math.Log1p(math.Exp(-2*expected))) / 2.5
	wantValidation := math.Log1p(math.Exp(-2 * expected))
	if math.Abs(trace.Epochs[0].TrainingNLL-wantTrain) > 1e-15 || math.Abs(trace.Epochs[0].ValidationNLL-wantValidation) > 1e-15 || trace.Training != (WeightSummary{Rows: 2, SampleWeightSum: 2.5}) || trace.Validation != (WeightSummary{Rows: 2, ZeroWeightRows: 1, SampleWeightSum: 3}) {
		t.Fatal("weighted loss or coverage differs from direct calculation", trace)
	}
}

func TestTraceTiesKeepEarliestEpoch(t *testing.T) {
	train, validation := weightedFixture("development", 0), weightedFixture("validation", 10)
	clear(train.Values)
	clear(validation.Values)
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 8, Batch: 3, Dimension: 1}
	result, trace, err := FitWithTrace(train, validation, c)
	if err != nil || result.Epoch != 1 || trace.SelectedEpoch != 1 {
		t.Fatal(result, trace, err)
	}
	for i, epoch := range trace.Epochs {
		if epoch.BestEpochSoFar != 1 || epoch.IsNewBest != (i == 0) || epoch.TrainingNLL != math.Log(2) || epoch.ValidationNLL != math.Log(2) {
			t.Fatal("tied epoch replaced earliest model", epoch)
		}
	}
}

func TestTraceInputFailuresAreEmpty(t *testing.T) {
	for _, mode := range []string{"role", "group", "column", "weight", "configuration", "weight_scale"} {
		t.Run(mode, func(t *testing.T) {
			train, validation := weightedFixture("development", 0), weightedFixture("validation", 10)
			c := Config{Seed: 1729, Mode: "fp32", LearningRate: .1, Epochs: 4, Batch: 3, Dimension: 1}
			switch mode {
			case "role":
				validation.Split = "final"
			case "group":
				validation.Groups[0] = train.Groups[0]
			case "column":
				train.Offsets[2] = 100
			case "weight":
				validation.SampleWeights = []float64{math.NaN(), 1, 1, 1}
			case "configuration":
				c.Epochs = 101
			case "weight_scale":
				train.SampleWeights = []float64{math.SmallestNonzeroFloat64, 0, 0, 0}
			}
			plain, oldErr := Fit(train, validation, c)
			got, trace, err := FitWithTrace(train, validation, c)
			if oldErr == nil || err == nil || oldErr.Error() != err.Error() || !reflect.DeepEqual(plain, got) || trace.Complete || trace.FailureKind != "input_validation" || trace.FailureEpoch != 0 || len(trace.Epochs) != 0 || trace.Training != (WeightSummary{}) || trace.Validation != (WeightSummary{}) {
				t.Fatal("input failure entered observation or changed behavior", got, trace, oldErr, err)
			}
			if _, err := json.Marshal(trace); err != nil {
				t.Fatal("invalid input leaked nonfinite values into trace", err)
			}
		})
	}
}

func TestTraceValidationFailureKeepsFinitePrefixAndOriginalResult(t *testing.T) {
	train := Dataset{Split: "development", Offsets: []int{0, 1}, Indices: []uint16{0}, Values: []float64{1}, Labels: []float64{1}, Groups: []int{1}}
	validation := cloneDataset(train)
	validation.Split, validation.Groups[0], validation.Values[0], validation.Labels[0] = "validation", 2, 1e308, 0
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: .2, Epochs: 100, Batch: 1, Dimension: 1}
	plain, oldErr := Fit(train, validation, c)
	got, trace, err := FitWithTrace(train, validation, c)
	if oldErr == nil || err == nil || oldErr.Error() != err.Error() || !reflect.DeepEqual(plain, got) || trace.Complete || trace.FailureKind != "nonfinite_validation_nll" || trace.FailureEpoch <= 1 || len(trace.Epochs) != trace.FailureEpoch-1 {
		t.Fatal("failure prefix or original partial result changed", got, trace, oldErr, err)
	}
	if trace.Epochs[len(trace.Epochs)-1].TrainingNLL == trace.Epochs[0].TrainingNLL || trace.Epochs[len(trace.Epochs)-1].BestEpochSoFar != 1 || trace.Epochs[len(trace.Epochs)-1].IsNewBest {
		t.Fatal("later epochs copied selected weights instead of current coefficients", trace)
	}
	if _, err := json.Marshal(trace); err != nil {
		t.Fatal("nonfinite failed epoch persisted", err)
	}
}

func TestTraceTrainingFailureLeavesLegacyFitUnchanged(t *testing.T) {
	train := Dataset{Split: "development", Offsets: []int{0, 1, 2}, Indices: []uint16{0, 0}, Values: []float64{1e308, 1}, Labels: []float64{0, 1}, Groups: []int{1, 2}, SampleWeights: []float64{0, 1}}
	validation := Dataset{Split: "validation", Offsets: []int{0, 1}, Indices: []uint16{0}, Values: []float64{1}, Labels: []float64{1}, Groups: []int{3}}
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: 1, Epochs: 20, Batch: 2, Dimension: 1}
	plain, oldErr := Fit(train, validation, c)
	got, trace, err := FitWithTrace(train, validation, c)
	if oldErr != nil || err == nil || !math.IsNaN(plain.TrainingNLL) || math.Float64bits(got.TrainingNLL) != math.Float64bits(plain.TrainingNLL) || !slices.Equal(got.Weights, plain.Weights) || got.Epoch != plain.Epoch || got.ValidationNLL != plain.ValidationNLL || trace.Complete || trace.FailureKind != "nonfinite_training_nll" || trace.FailureEpoch <= 1 || len(trace.Epochs) != trace.FailureEpoch-1 {
		t.Fatal("optional train observation changed underlying training", got, trace, oldErr, err)
	}
	if trace.SelectedEpoch <= trace.FailureEpoch || trace.Training.Rows != 2 || trace.Training.ZeroWeightRows != 1 || trace.Training.SampleWeightSum != 1 {
		t.Fatal("trace incorrectly stopped the original trainer or dropped zero-weight row", trace)
	}
	if _, err := json.Marshal(trace); err != nil {
		t.Fatal("nonfinite training observation persisted", err)
	}
}

func TestTraceOwnsObservationsAndIndependentWorkers(t *testing.T) {
	train, validation := fixture("development", 1), fixture("validation", 2)
	c := Config{Seed: 1729, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 10, Batch: 3, Dimension: 2}
	result, trace, err := FitWithTrace(train, validation, c)
	if err != nil {
		t.Fatal(err)
	}
	saved := trace
	saved.Epochs = slices.Clone(trace.Epochs)
	var workers sync.WaitGroup
	for range 3 {
		workers.Go(func() {
			got, observed, err := FitWithTrace(cloneDataset(train), cloneDataset(validation), c)
			if err != nil || !reflect.DeepEqual(got, result) || !reflect.DeepEqual(observed, saved) {
				t.Error("independent worker changed observations", err)
			}
			if len(observed.Epochs) != 0 {
				observed.Epochs[0].TrainingNLL = -1
			}
		})
	}
	workers.Wait()
	if !reflect.DeepEqual(trace, saved) {
		t.Fatal("worker reused a held trace backing array")
	}
	trace.Epochs[0].TrainingNLL = -2
	if result.TrainingNLL < 0 || saved.Epochs[0].TrainingNLL < 0 {
		t.Fatal("trace mutation changed the result or another trace")
	}
	clear(train.Values)
	clear(validation.Values)
	if saved.Training.Rows != 4 || saved.Validation.Rows != 4 || saved.Epochs[0].TrainingNLL < 0 {
		t.Fatal("trace retained mutable input columns")
	}
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
