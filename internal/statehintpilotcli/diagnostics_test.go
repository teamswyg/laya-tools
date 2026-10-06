package statehintpilotcli

import (
	"math"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintpilot"
)

func TestExcludedMeaningMassIsGroupedWithoutRenormalization(t *testing.T) {
	blocker, _ := statehint.IntentIndex(statehint.Blocker)
	reference, _ := statehint.IntentIndex(statehint.Reference)
	question, _ := statehint.IntentIndex(statehint.Question)
	var probabilities [statehint.IntentCount]float64
	probabilities[blocker], probabilities[reference], probabilities[question] = .3, .6, .1
	r := statehintpilot.Report{IntentOrder: statehint.Intents(), Observations: []statehintpilot.Observation{{Expected: statehint.Blocker, Observed: statehint.Reference, Confidence: .6, Probability: probabilities}}}
	d := probabilityDiagnostics(r)
	if math.Abs(d.NLL8+math.Log(.3)) > 1e-12 || math.Abs(d.DisplayNLL4+math.Log(.9)) > 1e-12 {
		t.Fatalf("wrong probability losses: %+v", d)
	}
	if math.Abs(d.DisplayBrier4-.02) > 1e-12 {
		t.Fatalf("excluded mass was changed: %+v", d)
	}
	if math.Abs(d.ECE10-.6) > 1e-12 || r.Observations[0].Observed != statehint.Reference {
		t.Fatal("diagnostic changed classification or confidence")
	}
}

func TestUnavailableAndGuardedObservationsStayVisible(t *testing.T) {
	unclear, _ := statehint.IntentIndex(statehint.Unclear)
	var p [statehint.IntentCount]float64
	p[unclear] = 1
	r := statehintpilot.Report{IntentOrder: statehint.Intents(), Observations: []statehintpilot.Observation{{Expected: statehint.Question}, {Expected: statehint.Question, Observed: statehint.Unclear, Confidence: 1, Probability: p, Guard: "empty_text"}}}
	d := probabilityDiagnostics(r)
	if d.Rows != 2 || d.Unavailable != 1 || d.Guarded != 1 || d.Classified != 1 || d.NLL8Clamped != 1 || d.DisplayNLL4Clamped != 1 {
		t.Fatalf("lost failed observations: %+v", d)
	}
	if math.IsInf(d.NLL8, 0) || math.IsNaN(d.NLL8) || d.ConfidenceBins[9].Cases != 1 {
		t.Fatal("non-finite or missing confidence=1")
	}
}
