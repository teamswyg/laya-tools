package pairlearn

import (
	"fmt"
	"math"
)

const TraceSchema = "riido-pairlearn-epoch-trace-v1"

// WeightSummary counts every validated row, including zero-weight rows. The sum
// is of the supplied sample weights, not an effective number of tasks. Nil
// SampleWeights means unit weights. A pathclaim caller supplies weights divided
// by its fixed scale; pairlearn never infers or changes that source convention.
type WeightSummary struct {
	Rows, ZeroWeightRows int
	SampleWeightSum      float64
}

// EpochTrace contains only finite numeric observations from the coefficients
// used for scoring at the end of this epoch. NLL uses the weighted objective
// when SampleWeights is nonnil. It is neither the shadow-weight loss nor the
// selected epoch's loss copied into later epochs. No coefficients or row IDs
// are exposed by the trace.
type EpochTrace struct {
	Epoch                      int
	TrainingNLL, ValidationNLL float64
	IsNewBest                  bool
	BestEpochSoFar             int
}

// Trace retains an owned, ordered prefix of fully finite epoch observations.
// An input failure has FailureEpoch zero and no dataset summary or epochs.
// A numerical failure names its epoch; the failing epoch is not in the finite
// prefix. Complete requires successful training and a record for every epoch.
type Trace struct {
	Schema, Mode, Precision string
	Training, Validation    WeightSummary
	Epochs                  []EpochTrace
	SelectedEpoch           int
	Complete                bool
	FailureEpoch            int
	FailureKind             string
}

// FitWithTrace runs the same fit as Fit, observing current epoch coefficients
// through a private value-only recorder. Inputs have one owner and must remain
// immutable for the call, as with Fit. The recorder cannot alter weights, choose
// epochs, invoke outside callbacks, write files or access any source identities.
//
// Extra train NLL calculations happen only in this opt-in path. Fit retains its
// historical validation failure behavior and final best-weight training NLL.
// If only an observed training NLL is nonfinite, the underlying fit still runs
// to its original conclusion; this wrapper returns the identical Result plus a
// trace error and the finite prefix. Never treat that incomplete trace as a
// successful recorded experiment.
func FitWithTrace(train, validation Dataset, c Config) (Result, Trace, error) {
	t := Trace{Schema: TraceSchema}
	result, err := fit(train, validation, c, &fitObserver{trace: &t})
	t.SelectedEpoch = result.Epoch
	if err != nil {
		if t.FailureKind == "" {
			t.FailureKind = "input_validation"
		}
		return result, t, err
	}
	if t.FailureKind != "" {
		return result, t, fmt.Errorf("incomplete epoch trace: %s", t.FailureKind)
	}
	if len(t.Epochs) != c.Epochs {
		t.FailureKind = "incomplete_epoch_trace"
		return result, t, fmt.Errorf("incomplete epoch trace")
	}
	t.Complete = true
	return result, t, nil
}

// Only the private trainer can emit observations. These callbacks are not an
// exported observer API and never receive slices of mutable fitted weights.
type fitObserver struct{ trace *Trace }

func (o *fitObserver) start(train, validation Dataset, c Config) {
	o.trace.Mode = c.Mode
	o.trace.Precision = "Coefficients are rounded to float32 before scoring; shadow weights, gradients and loss accumulation use float64."
	if c.Mode == "ternary_ste" {
		o.trace.Precision = "Coefficients use ternary values with a float32-rounded global scale; shadow weights, gradients and loss accumulation use float64."
	}
	o.trace.Training, o.trace.Validation = summarizeWeights(train), summarizeWeights(validation)
	o.trace.Epochs = make([]EpochTrace, 0, c.Epochs)
}

func (o *fitObserver) fail(epoch int, kind string) {
	if o.trace.FailureKind == "" {
		o.trace.FailureEpoch, o.trace.FailureKind = epoch, kind
	}
}

func (o *fitObserver) record(e EpochTrace) {
	if o.trace.FailureKind != "" {
		return
	}
	if math.IsNaN(e.TrainingNLL) || math.IsInf(e.TrainingNLL, 0) {
		o.fail(e.Epoch, "nonfinite_training_nll")
		return
	}
	o.trace.Epochs = append(o.trace.Epochs, e)
}

func summarizeWeights(d Dataset) WeightSummary {
	s := WeightSummary{Rows: len(d.Labels)}
	if d.SampleWeights == nil {
		s.SampleWeightSum = float64(s.Rows)
		return s
	}
	for _, w := range d.SampleWeights {
		s.SampleWeightSum += w
		if w == 0 {
			s.ZeroWeightRows++
		}
	}
	return s
}
