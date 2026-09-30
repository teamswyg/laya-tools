package pathclaim

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRegisteredTracedFitsPreserveAllTenResults(t *testing.T) {
	rows, ready := ownedReadyRows(t)
	plain, err := Fit(rows, ready)
	if err != nil {
		t.Fatal(err)
	}
	traced, traces, err := FitWithTrace(rows, ready)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plain, traced) {
		t.Fatal("tracing changed registered fits")
	}
	for i, trace := range traces {
		trial := traced.Trials[i]
		if !trace.Complete || len(trace.Epochs) != 100 || trace.FailureKind != "" || trace.SelectedEpoch != trial.Epoch || trace.Training.Rows != trial.Training.Rows || trace.Validation.Rows != trial.Validation.Rows || trace.Training.ZeroWeightRows != trial.Training.ZeroWeightRows || trace.Validation.ZeroWeightRows != trial.Validation.ZeroWeightRows {
			t.Fatal("incomplete or misbound trial trace", i)
		}
	}
	before, err := json.Marshal(traces)
	if err != nil {
		t.Fatal(err)
	}
	_, replay, err := FitWithTrace(rows, ready)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(replay)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("registered trace replay changed")
	}
}

func TestTracedFitsRejectInsufficientValidationBeforeAttempts(t *testing.T) {
	rows, ready := ownedReadyRows(t)
	ready.ValidationEligible = 2399
	fits, traces, err := FitWithTrace(rows, ready)
	if err == nil {
		t.Fatal("short validation accepted")
	}
	for i, trial := range fits.Trials {
		if trial.Attempted || trial.Fitted || len(traces[i].Epochs) != 0 {
			t.Fatal("fit attempted before readiness", i)
		}
	}
}
