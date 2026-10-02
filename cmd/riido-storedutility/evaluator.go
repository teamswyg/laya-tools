package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/teamswyg/laya-tools/internal/storedaudit"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

type metric struct {
	Kind                  string  `json:"kind"`
	KnownRequests         int     `json:"known_requests"`
	AnswerableRequests    int     `json:"answerable_requests"`
	NoAnswerRequests      int     `json:"no_answer_requests"`
	Checks                int     `json:"counterfactual_known_checks"`
	AnswerableChecks      int     `json:"counterfactual_answerable_checks"`
	Top1Correct           int     `json:"top1_correct"`
	Top3Correct           int     `json:"top3_correct"`
	RuleFallbacks         int     `json:"fallbacks_including_unknown"`
	FallbackDenominator   int     `json:"fallback_denominator"`
	GroupMacroKnownChecks float64 `json:"group_macro_known_checks"`
}

type groupMetric struct {
	GroupID      int    `json:"group_id"`
	Parents      int    `json:"parents"`
	Known        int    `json:"known"`
	Answerable   int    `json:"answerable"`
	NoAnswer     int    `json:"no_answer"`
	Unknown      int    `json:"unknown"`
	Checks       [4]int `json:"counterfactual_known_checks"`
	OracleChecks int    `json:"oracle_known_checks"`
}

type parentRanking struct {
	Reason                 string                `json:"stored_truth_reason,omitempty"`
	ParentID               string                `json:"parent_id"`
	GroupID                int                   `json:"group_id"`
	State                  string                `json:"stored_truth_state"`
	CandidateCount         int                   `json:"candidate_count"`
	AcceptableCount        int                   `json:"acceptable_count"`
	AcceptableIndices      [4]int                `json:"acceptable_indices"`
	ExcludedFromCostAndTop bool                  `json:"excluded_from_cost_and_top"`
	Rankings               [4]shortclaim.Ranking `json:"nonlearned_rankings"`
	Checks                 [4]int                `json:"counterfactual_checks"`
	OracleChecks           int                   `json:"oracle_checks"`
}

type report struct {
	UtilityEvaluated       bool                                `json:"utility_evaluated"`
	Schema                 string                              `json:"schema"`
	State                  string                              `json:"state"`
	Scope                  string                              `json:"scope"`
	ParentDenominator      int                                 `json:"parent_denominator"`
	CandidateDenominator   int                                 `json:"candidate_denominator"`
	ConnectedGroups        int                                 `json:"connected_groups"`
	LabeledConnectedGroups int                                 `json:"labeled_connected_groups"`
	DispatchAttempts       int                                 `json:"dispatch_attempts"`
	ValidatedRequests      int                                 `json:"validated_requests"`
	BaselineCalls          int                                 `json:"baseline_calls"`
	RankingOutputs         int                                 `json:"completed_control_rankings"`
	CompletedRows          int                                 `json:"completed_rows"`
	KnownRequests          int                                 `json:"known_requests"`
	AnswerableRequests     int                                 `json:"answerable_requests"`
	NoAnswerRequests       int                                 `json:"no_answer_requests"`
	UnknownRequests        int                                 `json:"unknown_requests"`
	OracleChecks           int                                 `json:"oracle_known_checks"`
	OracleAnswerableChecks int                                 `json:"oracle_answerable_checks"`
	BestBaseline           string                              `json:"best_fixed_nonlearned_baseline"`
	PossibleRelativeGain   float64                             `json:"possible_relative_gain"`
	AnswerableRelativeGain float64                             `json:"answerable_relative_gain"`
	GroupMacroOracleChecks float64                             `json:"group_macro_oracle_checks"`
	MacroGroups            int                                 `json:"macro_groups_with_known_requests"`
	GroupGatePass          bool                                `json:"group_gate_pass"`
	UtilityUpperBoundPass  bool                                `json:"utility_upper_bound_pass"`
	TrainingReady          bool                                `json:"training_ready"`
	RolesAssigned          int                                 `json:"roles_assigned"`
	SourceAPICalls         int                                 `json:"source_api_calls"`
	ModelCalls             int                                 `json:"model_calls"`
	Fits                   int                                 `json:"fits"`
	PaidCalls              int                                 `json:"paid_calls"`
	NewWeights             int                                 `json:"new_weights"`
	ProtectedFinalReads    int                                 `json:"protected_final_reads"`
	Metrics                [4]metric                           `json:"metrics"`
	Groups                 [storedaudit.GroupCount]groupMetric `json:"group_metrics"`
	Rows                   []parentRanking                     `json:"rows"`
	StopReasons            []string                            `json:"stop_reasons"`
}

var controlKinds = [...]string{shortclaim.FixedOrderKind, shortclaim.BM25Kind, shortclaim.LexicalOrderedKind, shortclaim.NarrowRuleKind}

// measureRankings consumes already-computed controls. It never sees prose,
// invokes a source candidate or turns an unknown outcome into a negative label.
func measureRankings(n int, truth storedaudit.Truth, rankings [4]shortclaim.Ranking) ([4]int, [4]bool, [4]bool, int, error) {
	var checks [4]int
	var top1, top3 [4]bool
	if n < 2 || n > 4 || truth.AcceptableCount < 0 || truth.AcceptableCount > n {
		return checks, top1, top3, 0, errors.New("storedutility_invalid_candidate_or_answer_count")
	}
	var allowed [4]bool
	for _, index := range truth.AcceptableIndices[:truth.AcceptableCount] {
		if index < 0 || index >= n || allowed[index] {
			return checks, top1, top3, 0, errors.New("storedutility_invalid_acceptable_set")
		}
		allowed[index] = true
	}
	oracle := 0
	switch truth.State {
	case "known":
		if truth.AcceptableCount == 0 {
			return checks, top1, top3, 0, errors.New("storedutility_known_without_answer")
		}
		oracle = 1
	case "no_answer":
		if truth.AcceptableCount != 0 {
			return checks, top1, top3, 0, errors.New("storedutility_no_answer_with_answer")
		}
		oracle = n
	case "unknown":
		if truth.AcceptableCount != 0 {
			return checks, top1, top3, 0, errors.New("storedutility_unknown_with_answer")
		}
	default:
		return checks, top1, top3, 0, errors.New("storedutility_invalid_truth_state")
	}
	for c, rank := range rankings {
		if rank.Kind != controlKinds[c] || rank.Count != n {
			return checks, top1, top3, 0, errors.New("storedutility_invalid_control_identity")
		}
		var seen [4]bool
		for i := 0; i < n; i++ {
			index := rank.Order[i]
			if index < 0 || index >= n || seen[index] || math.IsNaN(rank.Scores[i]) || math.IsInf(rank.Scores[i], 0) {
				return checks, top1, top3, 0, errors.New("storedutility_invalid_control_permutation_or_score")
			}
			seen[index] = true
		}
		for i := n; i < shortclaim.MaxCandidates; i++ {
			if rank.Order[i] != 0 || rank.Scores[i] != 0 {
				return checks, top1, top3, 0, errors.New("storedutility_nonzero_unused_control_slot")
			}
		}
		if truth.State == "unknown" {
			continue
		}
		checks[c] = n
		if truth.State == "known" {
			for i := 0; i < n; i++ {
				if allowed[rank.Order[i]] {
					checks[c] = i + 1
					break
				}
			}
			top1[c] = allowed[rank.Order[0]]
			for i := 0; i < min(3, n); i++ {
				if allowed[rank.Order[i]] {
					top3[c] = true
					break
				}
			}
		}
	}
	return checks, top1, top3, oracle, nil
}

func evaluate(d storedaudit.Dataset) (report, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return evaluateContext(ctx, d)
}

// Cancellation is checked between bounded rows, with no detached worker.
// shortclaim separately limits each raw text to512 bytes/32 normalized words.
func evaluateContext(ctx context.Context, d storedaudit.Dataset) (report, error) {
	r := report{Schema: "riido-storedtruth-nonlearned-utility-58-v1", State: "incomplete", Scope: "Unpartitioned development diagnostic of frozen authored inputs and stored finite truth. Counterfactual check counts, not candidate execution, model accuracy, held-out generalization or LLM/resource savings.", ParentDenominator: 72, CandidateDenominator: 216, ConnectedGroups: 17, LabeledConnectedGroups: 16}
	fail := func(reason string) (report, error) {
		r.StopReasons = append(r.StopReasons, reason)
		return r, errors.New(reason)
	}
	if ctx == nil {
		return fail("storedutility_missing_context")
	}
	if d.Schema != storedaudit.Schema || d.Answerable != 34 || d.NoAnswer != 17 || d.Unknown != 21 || d.LabeledGroups != 16 || d.Candidates != 216 || d.CandidateCountHistogram != [5]int{0, 0, 12, 48, 12} {
		return fail("storedutility_binding_metadata_mismatch")
	}
	r.StopReasons = append(r.StopReasons, d.Metadata.StopReasons...)
	r.StopReasons = append(r.StopReasons, "no_partition_or_training_execution_plan")
	for i, kind := range controlKinds {
		r.Metrics[i].Kind = kind
		r.Metrics[i].FallbackDenominator = 72
	}
	for i, group := range d.Groups {
		r.Groups[i].GroupID = group.ID
		for j := 0; j < i; j++ {
			if r.Groups[j].GroupID == group.ID {
				return fail("storedutility_duplicate_group")
			}
		}
	}
	r.Rows = make([]parentRanking, 0, 72)
	for _, row := range d.Rows {
		if ctx.Err() != nil {
			return fail("storedutility_ranking_budget_exceeded")
		}
		r.DispatchAttempts++
		groupIndex := -1
		for i := range r.Groups {
			if r.Groups[i].GroupID == row.Audit.GroupID {
				groupIndex = i
				break
			}
		}
		if groupIndex < 0 || row.Audit.ParentID == "" || row.Text.CandidateCount < 2 || row.Text.CandidateCount > 4 || row.Text.CandidateCount != row.Truth.CandidateCount || row.Text.CandidateCount != row.Audit.CandidateCount {
			return fail("storedutility_row_binding_mismatch")
		}
		// Metadata is not copied to request/candidate text or scoring provenance.
		in := shortclaim.Input{Schema: shortclaim.Schema, Request: row.Text.Request, Provenance: "stored-development-58"}
		var candidates [4]shortclaim.Candidate
		for i := 0; i < row.Text.CandidateCount; i++ {
			candidates[i] = shortclaim.Candidate{ID: fmt.Sprintf("candidate-%d", i), Text: row.Text.Candidates[i]}
		}
		in.Candidates = candidates[:row.Text.CandidateCount]
		p, err := shortclaim.Validate(in)
		if err != nil {
			return fail("storedutility_text_contract_rejected")
		}
		r.ValidatedRequests++
		r.BaselineCalls++
		rankings, err := shortclaim.Baselines(p)
		if err != nil {
			return fail("storedutility_control_computation_failed")
		}
		r.RankingOutputs += 4
		checks, top1, top3, oracle, err := measureRankings(p.Count, row.Truth, rankings)
		if err != nil {
			return fail(err.Error())
		}
		entry := parentRanking{ParentID: row.Audit.ParentID, GroupID: row.Audit.GroupID, State: row.Truth.State, Reason: row.Truth.Reason, CandidateCount: p.Count, AcceptableCount: row.Truth.AcceptableCount, AcceptableIndices: row.Truth.AcceptableIndices, ExcludedFromCostAndTop: row.Truth.State == "unknown", Rankings: rankings, Checks: checks, OracleChecks: oracle}
		r.Rows = append(r.Rows, entry)
		r.CompletedRows++
		g := &r.Groups[groupIndex]
		g.Parents++
		g.OracleChecks += oracle
		r.OracleChecks += oracle
		switch row.Truth.State {
		case "known":
			r.KnownRequests++
			r.AnswerableRequests++
			r.OracleAnswerableChecks++
			g.Known++
			g.Answerable++
		case "no_answer":
			r.KnownRequests++
			r.NoAnswerRequests++
			g.Known++
			g.NoAnswer++
		case "unknown":
			r.UnknownRequests++
			g.Unknown++
		}
		for i := range rankings {
			m := &r.Metrics[i]
			if rankings[i].FallbackReason != "" {
				m.RuleFallbacks++
			}
			if row.Truth.State == "unknown" {
				continue
			}
			m.KnownRequests++
			m.Checks += checks[i]
			g.Checks[i] += checks[i]
			if row.Truth.State == "known" {
				m.AnswerableRequests++
				m.AnswerableChecks += checks[i]
				if top1[i] {
					m.Top1Correct++
				}
				if top3[i] {
					m.Top3Correct++
				}
			} else {
				m.NoAnswerRequests++
			}
		}
		if ctx.Err() != nil {
			return fail("storedutility_ranking_budget_exceeded")
		}
	}
	if r.KnownRequests != 51 || r.AnswerableRequests != 34 || r.NoAnswerRequests != 17 || r.UnknownRequests != 21 || r.CompletedRows != 72 || r.RankingOutputs != 288 {
		return fail("storedutility_completed_count_mismatch")
	}
	for i, group := range r.Groups {
		if group.Parents != len(d.Groups[i].Parents) {
			return fail("storedutility_group_denominator_mismatch")
		}
		if group.Known == 0 {
			continue
		}
		r.MacroGroups++
		r.GroupMacroOracleChecks += float64(group.OracleChecks) / float64(group.Known)
		for c := range r.Metrics {
			r.Metrics[c].GroupMacroKnownChecks += float64(group.Checks[c]) / float64(group.Known)
		}
	}
	if r.MacroGroups != 16 {
		return fail("storedutility_labeled_group_count_mismatch")
	}
	r.GroupMacroOracleChecks /= float64(r.MacroGroups)
	best := 0
	for i := range r.Metrics {
		r.Metrics[i].GroupMacroKnownChecks /= float64(r.MacroGroups)
		if r.Metrics[i].Checks < r.OracleChecks {
			return fail("storedutility_cost_below_oracle")
		}
		if r.Metrics[i].Checks < r.Metrics[best].Checks {
			best = i
		}
	}
	m := r.Metrics[best]
	if m.Checks < 1 || m.AnswerableChecks < 1 {
		return fail("storedutility_invalid_utility_denominator")
	}
	r.BestBaseline = m.Kind
	r.PossibleRelativeGain = float64(m.Checks-r.OracleChecks) / float64(m.Checks)
	r.AnswerableRelativeGain = float64(m.AnswerableChecks-r.OracleAnswerableChecks) / float64(m.AnswerableChecks)
	r.GroupGatePass = r.ConnectedGroups >= 15 && r.LabeledConnectedGroups >= 15
	// Exact integer5% comparison; the floating ratio is reporting only.
	r.UtilityUpperBoundPass = 100*(m.Checks-r.OracleChecks) >= 5*m.Checks
	if !r.UtilityUpperBoundPass {
		r.StopReasons = append(r.StopReasons, "insufficient_oracle_utility_upper_bound")
	}
	if ctx.Err() != nil {
		return fail("storedutility_ranking_budget_exceeded")
	}
	r.State = "complete"
	r.UtilityEvaluated = true
	return r, nil
}
