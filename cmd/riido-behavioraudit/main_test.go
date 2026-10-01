package main

import (
	"testing"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
)

func diagnosticFixture() (behaviorprobe.Dataset, behaviorprobe.Report, auditPlan) {
	p := behaviorprobe.Parent{ID: "answerable", Request: "preserve input order", Candidates: []behaviorprobe.SourceCandidate{{ID: "a", Text: "Reverse the sequence."}, {ID: "b", Text: "Preserve input order."}, {ID: "c", Text: "Drop the sequence."}}}
	d := behaviorprobe.Dataset{Parents: []behaviorprobe.Parent{p, p, p}}
	d.Parents[1].ID, d.Parents[2].ID = "no-answer", "unknown"
	truth := behaviorprobe.Report{ConnectedGroups: 1, Outcomes: []behaviorprobe.Outcome{
		{ParentID: "answerable", State: "known", Acceptable: []int{1}},
		{ParentID: "no-answer", State: "no_answer", Acceptable: []int{}},
		{ParentID: "unknown", State: "unknown", Acceptable: []int{}},
	}}
	return d, truth, auditPlan{Parents: 3, MinConnectedGroups: 15, MinPossibleRelativeGain: .05}
}

func TestUnknownExclusionNoAnswerExhaustionAndOracleBound(t *testing.T) {
	d, truth, plan := diagnosticFixture()
	r, err := evaluate(d, truth, plan)
	if err != nil {
		t.Fatal(err)
	}
	fixed := r.Metrics[0]
	if fixed.KnownRequests != 2 || fixed.AnswerableRequests != 1 || fixed.NoAnswerRequests != 1 || fixed.Checks != 5 || fixed.AnswerableChecks != 2 || fixed.Top1Correct != 0 || fixed.Top3Correct != 1 {
		t.Fatalf("fixed denominator/ordering mismatch: %+v", fixed)
	}
	if r.OracleChecks != 4 || r.OracleAnswerableChecks != 1 || r.PossibleRelativeGain != 0 || r.GroupGatePass || r.UtilityUpperBoundPass || r.Fits != 0 || r.TrainingExecutionReady || len(r.ParentRankings) != 3 {
		t.Fatalf("incorrect scope or oracle bound: %+v", r)
	}
}

func TestAllowedSetAndTruthFailures(t *testing.T) {
	d, truth, plan := diagnosticFixture()
	truth.Outcomes[0].Acceptable = []int{0, 2}
	r, err := evaluate(d, truth, plan)
	if err != nil || r.Metrics[0].AnswerableChecks != 1 || r.Metrics[0].Top1Correct != 1 {
		t.Fatalf("multiple answers unsupported: %v", err)
	}
	for _, invalid := range [][]int{{0, 0}, {-1}, {3}, {}} {
		truth.Outcomes[0].Acceptable = invalid
		if _, err := evaluate(d, truth, plan); err == nil {
			t.Fatal("invalid known truth accepted")
		}
	}
	truth.Outcomes[0].Acceptable = []int{1}
	truth.Outcomes[1].Acceptable = []int{0}
	if _, err := evaluate(d, truth, plan); err == nil {
		t.Fatal("no-answer contradictory truth accepted")
	}
}
