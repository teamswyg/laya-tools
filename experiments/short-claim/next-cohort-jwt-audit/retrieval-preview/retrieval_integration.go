// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Author-only integration: no Go/ranking/JWT calls have been executed.
package main

import (
	"crypto/rsa"
	"time"
)

const retrievalScope = "Exposed four-parent development diagnostic only; standalone source-derived BM25 and exact token Jaccard, not pkg API timing. Raw normalization, preparation, score, sort and permutation work is charged per operation. Four rotated rounds are process-warm after160 live pairs, not randomized or independent replication. Unknown ranking aborts before that operation's first JWT call. Current160 results qualify repeated trials but never enter ranking. Explicit Validate counters exclude uncached parser-internal validation. No model, Fit, role admission or savings claim."

var retrievalControls = [4]string{"fixed-source-order", "manual-first", "bm25-text", "jaccard-token-set"}
var retrievalCounterColumns = [17]string{"option_constructors", "option_applications", "direct_new_parser", "parse_with_claims", "new_validator", "implicit_parser_in_new_validator", "validate", "key_callbacks", "time_callbacks", "cache_lookups", "cache_key_comparisons", "cache_hits", "cache_misses", "deep_claims_copies", "public_key_copies", "candidate_checks", "fixture_checks"}
var retrievalTimingColumns = [10]string{"setup_ns", "cache_lookup_ns", "claims_and_key_copy_ns", "cache_claims_store_ns", "control_order_construction_ns", "authentication_and_parse_ns", "fresh_policy_validate_ns", "combined_parse_signature_policy_ns", "outcome_qualification_ns", "complete_parent_work_ns"}
var retrievalOperationColumns = [14]string{"round", "control_position", "cached", "parent_position", "candidate_order", "scores", "candidate_attempted", "fixture_checks_by_candidate", "candidate_states", "unknown_code_indices", "first_matching_candidate", "complete", "counts", "timings"}
var retrievalStates = [4]string{"unattempted", "known_match", "known_mismatch", "unknown"}
var retrievalUnknownCodes = [32]string{"none", "unknown_observer", "unknown_panic", "unknown_deadline", "unknown_key_mutation", "unknown_cache_capacity", "unknown_authentication", "unknown_auth_payload", "unknown_cache_mutation", "unknown_payload", "unknown_return", "unknown_error", "unknown_operational_parity", "unknown_retrieval_kind", "unknown_retrieval_text", "unknown_retrieval_unicode", "unknown_retrieval_normalized", "unknown_retrieval_score", "unknown_retrieval_permutation", "unknown_retrieval_panic", "unknown_text_pin", "unknown_text_shape", "unknown_text_binding", "unknown_operation_context", "unknown_no_match", "unknown_unlisted_code", "unknown_output_scalar", "unknown_output_bound", "unknown_output_marshal", "unknown_output_write", "unknown_equivalence", "unknown_runtime"}

type retrievalOperationRow [14]any

// Failure is separate from candidate states; rank-U must never label any candidate F.
type retrievalFailure struct {
	Phase     string  `json:"phase"`
	Code      string  `json:"code"`
	Round     int     `json:"round"`
	Control   int     `json:"control_position"`
	Parent    int     `json:"parent_position"`
	Cached    bool    `json:"cached"`
	Candidate int     `json:"candidate_position"`
	Fixture   int     `json:"fixture_position"`
	Actual    *Result `json:"actual_trial"`
}
type Output struct {
	Schema                  string                  `json:"schema"`
	SourceRevision          string                  `json:"source_revision"`
	InputSHA                string                  `json:"input_sha256"`
	TextInputSHA            string                  `json:"text_input_sha256"`
	PreparationSHA          string                  `json:"source_preparation_sha256"`
	PublicKeyDERSHA         string                  `json:"public_key_der_sha256"`
	Family                  string                  `json:"source_family"`
	Role                    string                  `json:"role"`
	Scope                   string                  `json:"scope"`
	OwnInit                 int                     `json:"own_init_calls"`
	ImportedInitObserved    bool                    `json:"imported_package_init_count_observed"`
	MainStartsAfterImports  bool                    `json:"main_timing_excludes_import_startup"`
	PolicyMaskClasses       [8]string               `json:"normalized_error_classes_bit_order"`
	Controls                [4]string               `json:"controls"`
	CounterColumns          [17]string              `json:"counter_columns"`
	TimingColumns           [10]string              `json:"timing_columns"`
	OperationalColumns      [14]string              `json:"operational_columns"`
	StateDictionary         [4]string               `json:"candidate_state_dictionary"`
	UnknownDictionary       [32]string              `json:"unknown_code_dictionary"`
	Equivalence             Equivalence             `json:"exhaustive_equivalence"`
	Runs                    []retrievalOperationRow `json:"operational_runs"`
	OperationalParityChecks int                     `json:"operational_parity_checks"`
	TextRankCalls           int                     `json:"text_rank_calls"`
	ReferenceSetupNS        int64                   `json:"parity_reference_setup_ns"`
	Failure                 *retrievalFailure       `json:"first_failure"`
	Complete                bool                    `json:"complete"`
	Code                    string                  `json:"code"`
	MainNS                  int64                   `json:"main_elapsed_ns"`
	InputSetupNS            int64                   `json:"input_and_public_key_setup_ns"`
	SignaturesGenerated     int                     `json:"signatures_generated"`
	KeyGenerations          int                     `json:"key_generations"`
	ModelCalls              int                     `json:"model_calls"`
	FitCalls                int                     `json:"fit_calls"`
	RSAVerifyCallsObserved  bool                    `json:"rsa_verify_call_count_observed"`
	InputUnchanged          bool                    `json:"input_unmodified"`
	TextInputUnchanged      bool                    `json:"text_input_unmodified"`
}
type retrievalOperationResult struct {
	Round, Control, Parent      int
	Cached                      bool
	Order                       [8]int
	Scores                      [8]float64
	Checked                     [8]bool
	Fixtures                    [8]int
	States                      [8]int
	UnknownCodes                [8]int
	Found                       int
	Complete                    bool
	Stats                       Stats
	Failure                     *retrievalFailure
	ParityChecks, TextRankCalls int
}

func retrievalUnknownIndex(code string) int {
	for i, known := range retrievalUnknownCodes {
		if code == known {
			return i
		}
	}
	return 25 // Unknown identifier stays U; raw observer errors are never printed.
}
func retrievalCounts(c Counts) [17]int {
	return [17]int{c.OptionConstructors, c.OptionApplications, c.DirectNewParser, c.ParseWithClaims, c.NewValidator, c.ImplicitParserInNewValidator, c.Validate, c.KeyCallbacks, c.TimeCallbacks, c.CacheLookups, c.CacheComparisons, c.CacheHits, c.CacheMisses, c.ClaimsCopies, c.PublicKeyCopies, c.CandidateChecks, c.FixtureChecks}
}
func retrievalTimings(t Timings) [10]*int64 {
	return [10]*int64{&t.SetupNS, &t.LookupNS, &t.CopyNS, &t.CacheStoreNS, &t.RankNS, t.AuthParseNS, t.PolicyNS, t.CombinedNS, &t.QualificationNS, t.TotalNS}
}
func retrievalCompact(o retrievalOperationResult) retrievalOperationRow {
	return retrievalOperationRow{o.Round, o.Control, o.Cached, o.Parent, o.Order, o.Scores, o.Checked, o.Fixtures, o.States, o.UnknownCodes, o.Found, o.Complete, retrievalCounts(o.Stats.Counts), retrievalTimings(o.Stats.Timings)}
}

// The reference is the first current live160-pair phase, not previous outputs or
// Wanted. Construction is separately charged and does not normalize any text.
func retrievalRetainParity(eq Equivalence) ([4][5][8]Result, bool) {
	var out [4][5][8]Result
	if !eq.AllEqual || eq.PlannedPairs != 160 || eq.AttemptedPairs != 160 || eq.EqualPairs != 160 || len(eq.Records) != 160 {
		return out, false
	}
	for i, pair := range eq.Records {
		p, f, c := i/40, (i/8)%5, i%8
		if pair.Parent != p || pair.Fixture != f || pair.Candidate != c || !pair.Equal || !pair.Reference.Known || !pair.Cached.Known || pair.Reference != pair.Cached {
			return [4][5][8]Result{}, false
		}
		out[p][f][c] = pair.Reference
	}
	return out, true
}
func retrievalParityCode(got, expected Result) string {
	if !got.Known {
		if got.Code == "" {
			return "unknown_observer"
		}
		return retrievalUnknownCodes[retrievalUnknownIndex(got.Code)]
	}
	if !expected.Known || got != expected {
		return "unknown_operational_parity"
	}
	return ""
}
func retrievalChooseOrder(text *retrievalTextInput, parent, control int) (retrievalRanking, retrievalCode, bool) {
	if text == nil || parent < 0 || parent >= 4 || control < 0 || control >= 4 {
		return retrievalDeclaration(), retrievalCode("unknown_operation_context"), false
	}
	if control < 2 {
		return retrievalRanking{Order: orderFor(parent, control == 1)}, retrievalOK, false
	}
	kind := retrievalBM25
	if control == 3 {
		kind = retrievalJaccard
	}
	rank, code := rankRetrievalText(text.Requests[parent], text.Captions, kind)
	return rank, code, true
}

// Every operation owns its cache/stats/arrays. The timer starts before ranking
// construction and closes after all result/error metadata. No mutable scratch is
// shared. Unknown rank returns before any trial, option, parser or key callback.
func retrievalOperation(in Input, key *rsa.PublicKey, der string, text *retrievalTextInput, expected *[4][5][8]Result, round, parent, control int, cached bool, deadline time.Time) (out retrievalOperationResult) {
	started := time.Now()
	out = retrievalOperationResult{Round: round, Parent: parent, Control: control, Cached: cached, Found: -1, Complete: true, Stats: newStats(cached)}
	activeCandidate, activeFixture := -1, -1
	defer func() {
		if recover() != nil {
			out.Complete = false
			out.Failure = &retrievalFailure{Phase: "operation", Code: "unknown_panic", Round: round, Control: control, Parent: parent, Cached: cached, Candidate: activeCandidate, Fixture: activeFixture}
			if activeCandidate >= 0 && activeCandidate < 8 {
				out.States[activeCandidate] = 3
				out.UnknownCodes[activeCandidate] = retrievalUnknownIndex("unknown_panic")
			}
		}
		elapsed := time.Since(started).Nanoseconds()
		out.Stats.Timings.TotalNS = &elapsed
	}()
	rankingStarted := time.Now()
	rank, rankCode, called := retrievalChooseOrder(text, parent, control)
	out.Order, out.Scores = rank.Order, rank.Scores
	if called {
		out.TextRankCalls = 1
	}
	if rankCode != retrievalOK {
		out.Complete = false
		out.Failure = &retrievalFailure{Phase: "rank", Code: string(rankCode), Round: round, Control: control, Parent: parent, Cached: cached, Candidate: -1, Fixture: -1}
	}
	out.Stats.Timings.RankNS = time.Since(rankingStarted).Nanoseconds()
	if !out.Complete {
		return out
	}
	if expected == nil {
		out.Complete = false
		out.Failure = &retrievalFailure{Phase: "qualification", Code: "unknown_operational_parity", Round: round, Control: control, Parent: parent, Cached: cached, Candidate: -1, Fixture: -1}
		return out
	}
	var cache Cache
	for _, candidate := range out.Order {
		activeCandidate = candidate
		out.Checked[candidate] = true
		out.Stats.Counts.CandidateChecks++
		matches := true
		for fixture := 0; fixture < 5; fixture++ {
			activeFixture = fixture
			f := in.Fixtures[positions[parent][fixture]]
			result := trial(f, candidate, cached, &cache, key, der, &out.Stats, deadline)
			out.Fixtures[candidate]++
			qualification := time.Now()
			out.ParityChecks++
			code := retrievalParityCode(result, expected[parent][fixture][candidate])
			if code != "" {
				out.Complete = false
				out.States[candidate] = 3
				out.UnknownCodes[candidate] = retrievalUnknownIndex(code)
				actual := result
				out.Failure = &retrievalFailure{Phase: "verification", Code: code, Round: round, Control: control, Parent: parent, Cached: cached, Candidate: candidate, Fixture: fixture, Actual: &actual}
			}
			out.Stats.Timings.QualificationNS += time.Since(qualification).Nanoseconds()
			if !out.Complete {
				return out
			}
			if result.Accept != wanted[parent][fixture] {
				matches = false
				break
			}
		}
		if matches {
			out.States[candidate] = 1
			out.Found = candidate
			return out
		}
		out.States[candidate] = 2
	}
	out.Complete = false
	out.Failure = &retrievalFailure{Phase: "no_match", Code: "unknown_no_match", Round: round, Control: control, Parent: parent, Cached: cached, Candidate: -1, Fixture: -1}
	return out
}

// Raw input ownership only; no normalization, scores, fixtures or Wanted.
func retrievalTextUnmodified(before, current retrievalTextInput, raw string) bool {
	return before == current && hash([]byte(raw)) == retrievalTextInputSHA
}
