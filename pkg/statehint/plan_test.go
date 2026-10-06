package statehint

import (
	"errors"
	"math"
	"testing"
)

// scoredTestPrediction supplies deliberately fake scores for policy tests,
// never presented as an inference result or trained model evidence.
func scoredTestPrediction(intent Intent) Prediction {
	var probabilities [IntentCount]float64
	for class := range probabilities {
		probabilities[class] = .02 / 7
	}
	index, _ := IntentIndex(intent)
	probabilities[index] = .98
	return prediction(probabilities, Learned, 1)
}
func basePlanInput() PlanInput {
	return PlanInput{WorkID: "original-work", CommandID: "original-command", ExpectedVersion: 4, CurrentState: Todo, MinConfidence: .9, MinMargin: .05}
}

func TestStatePlanRequiresMatchingScopedTrustedEvidence(t *testing.T) {
	for _, c := range []struct {
		intent Intent
		kind   EventKind
		state  State
	}{{Progress, StartedEvent, Active}, {CompletionReport, CompletedEvent, Done}, {CancelRequest, CancelledEvent, Cancelled}} {
		input := basePlanInput()
		p := scoredTestPrediction(c.intent)
		correct := EventEvidence{ID: "original-event", Kind: c.kind, WorkID: input.WorkID, Version: input.ExpectedVersion, Trusted: true}
		for _, e := range []EventEvidence{{}, {ID: "event", Kind: c.kind, WorkID: input.WorkID, Version: 4}, {ID: "event", Kind: c.kind, WorkID: "other", Version: 4, Trusted: true}, {ID: "event", Kind: c.kind, WorkID: input.WorkID, Version: 3, Trusted: true}} {
			input.Evidence = nil
			if e.ID != "" {
				input.Evidence = []EventEvidence{e}
			}
			result, err := Plan(input, p)
			if err != nil || result.StateChange || result.NextState != Todo || result.MutationExecuted || result.Mode != "shadow" {
				t.Fatal(result, err)
			}
		}
		input.Evidence = []EventEvidence{correct}
		result, err := Plan(input, p)
		if err != nil || !result.StateChange || result.NextState != c.state || result.EvidenceID != correct.ID || result.MutationExecuted {
			t.Fatal(result, err)
		}
		input.CurrentState = c.state
		result, err = Plan(input, p)
		if err != nil || result.StateChange || !result.NoOp || result.Reason != "state_already_selected" {
			t.Fatal("repeated state change", result, err)
		}
	}
}

func TestMismatchedAndConflictingEventsCannotChangeState(t *testing.T) {
	input := basePlanInput()
	input.Evidence = []EventEvidence{{ID: "event", Kind: CompletedEvent, WorkID: input.WorkID, Version: 4, Trusted: true}}
	for _, intent := range []Intent{Question, Blocker, Reference, Progress, CancelRequest, Planned, Unclear} {
		result, err := Plan(input, scoredTestPrediction(intent))
		if err != nil || result.StateChange {
			t.Fatal(result, err)
		}
	}
	input.Evidence = append(input.Evidence, EventEvidence{ID: "conflict", Kind: CancelledEvent, WorkID: input.WorkID, Version: 4, Trusted: true})
	result, err := Plan(input, scoredTestPrediction(CompletionReport))
	if err != nil || result.StateChange || result.Reason != "conflicting_trusted_events" {
		t.Fatal(result, err)
	}
}

func TestSuppliedCatalogAnnotationsAndRepeatedProposal(t *testing.T) {
	input := basePlanInput()
	input.Labels = []CatalogLabel{{ID: "original-question", Intent: Question, Active: true}, {ID: "original-inactive", Intent: Question}, {ID: "original-reference", Intent: Reference, Active: true}}
	input.Emojis = DefaultEmojis()
	p := scoredTestPrediction(Question)
	result, err := Plan(input, p)
	if err != nil || len(result.LabelIDs) != 1 || result.LabelIDs[0] != "original-question" || result.EmojiCode != "2753" || result.MutationExecuted || result.StateChange {
		t.Fatal(result, err)
	}
	input.ExistingLabelIDs = result.LabelIDs
	input.CurrentEmojiCode = result.EmojiCode
	result, err = Plan(input, p)
	if err != nil || !result.NoOp || len(result.LabelIDs) != 0 || result.EmojiCode != "" {
		t.Fatal("repeated annotation", result, err)
	}
	p.Source = Untrained
	p.TrainingSteps = 0
	result, err = Plan(basePlanInput(), p)
	if err != nil || !result.NoOp || result.Reason != "untrained_model" {
		t.Fatal(result, err)
	}
}

func TestInvalidPlanFailsClosed(t *testing.T) {
	for _, modify := range []func(*PlanInput){
		func(i *PlanInput) { i.CommandID = "" }, func(i *PlanInput) { i.ExpectedVersion = 0 }, func(i *PlanInput) { i.CurrentState = "other" },
		func(i *PlanInput) { i.MinConfidence = math.NaN() },
		func(i *PlanInput) { i.MinConfidence = .1 },
		func(i *PlanInput) {
			i.Labels = []CatalogLabel{{ID: "same", Intent: Question}, {ID: "same", Intent: Blocker}}
		},
		func(i *PlanInput) { i.Emojis = []EmojiCandidate{{Intent: Question, Code: "❓"}} },
	} {
		input := basePlanInput()
		modify(&input)
		if _, err := Plan(input, scoredTestPrediction(Question)); !errors.Is(err, ErrPlan) {
			t.Fatal("invalid plan accepted", err)
		}
	}
	for _, code := range []string{"1F440", "1f600-", "d800", "110000", "00001f", "✅", "0020-fe0f-"} {
		if ValidEmojiCode(code) {
			t.Fatal("invalid canonical code accepted", code)
		}
	}
	for _, emoji := range DefaultEmojis() {
		if !ValidEmojiCode(emoji.Code) {
			t.Fatal("default invalid", emoji)
		}
	}
	p := scoredTestPrediction(Question)
	p.Confidence = .5
	if _, err := Plan(basePlanInput(), p); !errors.Is(err, ErrPlan) {
		t.Fatal("inconsistent score accepted")
	}
}

func TestBelowThresholdAndUnclearMakeNoProposal(t *testing.T) {
	input := basePlanInput()
	input.MinConfidence = .99
	input.Labels = []CatalogLabel{{ID: "q", Intent: Question, Active: true}}
	result, err := Plan(input, scoredTestPrediction(Question))
	if err != nil || !result.NoOp || result.Reason != "below_threshold" {
		t.Fatal(result, err)
	}
	input.MinConfidence = .9
	result, err = Plan(input, scoredTestPrediction(Unclear))
	if err != nil || !result.NoOp || result.Reason != "unclear" {
		t.Fatal(result, err)
	}
}
