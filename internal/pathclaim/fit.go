package pathclaim

import (
	"fmt"

	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

type Trial struct {
	Lambda                     float64
	Seed                       uint64
	Epoch                      int
	TrainingNLL, ValidationNLL float64
	Training, Validation       DatasetCounts
	Attempted, Fitted          bool
	Failure                    string
	// Weights are definition-level numeric coefficients, populated only by an
	// actual successful fit. Keep them out of ordinary aggregate JSON reports.
	Weights [searchclaim.PathDimension]float32 `json:"-"`
	// Private verified provenance prevents construction from a toggled fit flag.
	input Readiness
}

type Fits struct {
	Readiness Readiness
	Trials    [10]Trial
}

func frozenConfig(seed uint64) pairlearn.Config {
	return pairlearn.Config{Seed: seed, Mode: "fp32", LearningRate: .1, L2: .0001, Epochs: 100, Batch: 64, Dimension: searchclaim.PathDimension}
}

// Fit makes exactly the ten precommitted candidate attempts, sequentially, after
// validating actual rows against the readiness seal. It offers no configuration
// or readiness override. It neither selects thresholds/winners nor evaluates
// final tasks, proves permissions, publishes models or activates a policy.
//
// pairlearn currently exposes selected epoch/NLL only; full per-epoch traces are
// not supplied by this adapter. A real experiment runner must address that plan
// requirement before claiming complete plan-46 execution.
func Fit(rows []Row, readiness Readiness) (Fits, error) {
	result := Fits{Readiness: readiness}
	if err := VerifyReadiness(rows, readiness); err != nil {
		return result, err
	}
	failed := false
	i := 0
	for _, lambda := range Penalties() {
		train, tc, trainErr := Prepare(rows, "train", lambda)
		validation, vc, validationErr := Prepare(rows, "validation", lambda)
		for _, seed := range Seeds() {
			trial := Trial{Lambda: lambda, Seed: seed, Training: tc, Validation: vc, input: readiness}
			switch {
			case trainErr != nil:
				trial.Failure = trainErr.Error()
			case validationErr != nil:
				trial.Failure = validationErr.Error()
			default:
				trial.Attempted = true
				fit, err := pairlearn.Fit(train, validation, frozenConfig(seed))
				if err != nil {
					trial.Failure = err.Error()
				} else {
					trial.Epoch, trial.TrainingNLL, trial.ValidationNLL = fit.Epoch, fit.TrainingNLL, fit.ValidationNLL
					for j, w := range fit.Weights {
						trial.Weights[j] = float32(w)
					}
					trial.Fitted = true
				}
			}
			if !trial.Fitted {
				failed = true
			}
			result.Trials[i], i = trial, i+1
		}
	}
	if failed {
		return result, fmt.Errorf("one or more registered path fits failed")
	}
	return result, nil
}
