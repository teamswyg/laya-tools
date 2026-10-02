package main

import (
	"context"
	"math"
	"slices"
	"testing"

	"github.com/teamswyg/laya-tools/internal/storedaudit"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// These controls contain literal permutations, not computed corpus rankings.
func literalRankings(n int, order [4]int) [4]shortclaim.Ranking {
	var ranks [4]shortclaim.Ranking
	for i, kind := range controlKinds {
		ranks[i].Kind, ranks[i].Count = kind, n
		copy(ranks[i].Order[:n], order[:n])
	}
	return ranks
}

func TestCounterfactualCostPreservesMultipleAnswersAndUnknown(t *testing.T) {
	ranks := literalRankings(4, [4]int{0, 2, 1, 3})
	truth := storedaudit.Truth{State: "known", AcceptableCount: 2, AcceptableIndices: [4]int{1, 3}}
	checks, top1, top3, oracle, err := measureRankings(4, truth, ranks)
	if err != nil || checks != [4]int{3, 3, 3, 3} || top1 != [4]bool{} || top3 != [4]bool{true, true, true, true} || oracle != 1 {
		t.Fatal("first acceptable set was not preserved")
	}
	truth = storedaudit.Truth{State: "no_answer"}
	checks, top1, top3, oracle, err = measureRankings(4, truth, ranks)
	if err != nil || checks != [4]int{4, 4, 4, 4} || top1 != [4]bool{} || top3 != [4]bool{} || oracle != 4 {
		t.Fatal("no-answer must inspect every candidate")
	}
	truth = storedaudit.Truth{State: "unknown"}
	checks, top1, top3, oracle, err = measureRankings(4, truth, ranks)
	if err != nil || checks != [4]int{} || top1 != [4]bool{} || top3 != [4]bool{} || oracle != 0 {
		t.Fatal("unknown must remain excluded from cost and top metrics")
	}
	ranks = literalRankings(2, [4]int{0, 1})
	truth = storedaudit.Truth{State: "known", AcceptableCount: 1, AcceptableIndices: [4]int{1}}
	checks, top1, top3, oracle, err = measureRankings(2, truth, ranks)
	if err != nil || checks != [4]int{2, 2, 2, 2} || top1 != [4]bool{} || top3 != [4]bool{true, true, true, true} || oracle != 1 {
		t.Fatal("two-candidate Top3 bounds were changed")
	}
}

func TestCounterfactualCostRejectsInvalidOutcomesAndPermutations(t *testing.T) {
	validTruth := storedaudit.Truth{State: "known", AcceptableCount: 1, AcceptableIndices: [4]int{1}}
	for _, truth := range []storedaudit.Truth{
		{State: "known"},
		{State: "unknown", AcceptableCount: 1, AcceptableIndices: [4]int{0}},
		{State: "no_answer", AcceptableCount: 1, AcceptableIndices: [4]int{0}},
		{State: "known", AcceptableCount: 2, AcceptableIndices: [4]int{1, 1}},
		{State: "known", AcceptableCount: 1, AcceptableIndices: [4]int{2}},
	} {
		if _, _, _, _, err := measureRankings(2, truth, literalRankings(2, [4]int{0, 1})); err == nil {
			t.Fatal("invalid stored outcome was accepted")
		}
	}
	for i := 0; i < 6; i++ {
		ranks := literalRankings(2, [4]int{0, 1})
		switch i {
		case 0:
			ranks[1].Order[1] = 0
		case 1:
			ranks[1].Order[1] = 2
		case 2:
			ranks[1].Count = 3
		case 3:
			ranks[1].Scores[0] = math.NaN()
		case 4:
			ranks[1].Scores[0] = math.Inf(1)
		case 5:
			ranks[1].Scores[2] = 1
		}
		if _, _, _, _, err := measureRankings(2, validTruth, ranks); err == nil {
			t.Fatal("malformed control was accepted")
		}
	}
}

func metadataOnlyDataset() storedaudit.Dataset {
	d := storedaudit.Dataset{Schema: storedaudit.Schema, Answerable: 34, NoAnswer: 17, Unknown: 21, LabeledGroups: 16, Candidates: 216, CandidateCountHistogram: [5]int{0, 0, 12, 48, 12}}
	d.Metadata.StopReasons = []string{"no_roles_plan", "synthetic_single_pipeline"}
	for i := range d.Groups {
		d.Groups[i].ID = i
	}
	return d
}

func TestCancellationAndInvalidBoundsPrecedeAnyControlCall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, err := evaluateContext(ctx, metadataOnlyDataset())
	if err == nil || r.State != "incomplete" || r.UtilityEvaluated || r.DispatchAttempts != 0 || r.ValidatedRequests != 0 || r.BaselineCalls != 0 || r.RankingOutputs != 0 || r.CompletedRows != 0 {
		t.Fatal("cancellation allowed scoring")
	}
	if !slices.Contains(r.StopReasons, "no_roles_plan") || !slices.Contains(r.StopReasons, "synthetic_single_pipeline") || !slices.Contains(r.StopReasons, "no_partition_or_training_execution_plan") {
		t.Fatal("partial report lost prior limitations")
	}
	d := metadataOnlyDataset()
	d.Rows[0].Audit.ParentID = "literal-invalid-bound"
	d.Rows[0].Text.CandidateCount = 5
	d.Rows[0].Truth.CandidateCount = 5
	d.Rows[0].Audit.CandidateCount = 5
	r, err = evaluateContext(context.Background(), d)
	if err == nil || r.State != "incomplete" || r.UtilityEvaluated || r.DispatchAttempts != 1 || r.ValidatedRequests != 0 || r.BaselineCalls != 0 || r.RankingOutputs != 0 || r.CompletedRows != 0 {
		t.Fatal("invalid bounds reached text preparation")
	}
}
