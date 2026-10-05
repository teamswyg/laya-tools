// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"testing"
	"time"
)

func retrievalSyntheticText() retrievalTextInput {
	return retrievalTextInput{Requests: [4]string{"red green", "blue", "purple", "yellow"}, Captions: retrievalSyntheticCaptions()}
}
func TestRetrievalRankUnknownAbortsBeforeJWTOrCandidateFalse(t *testing.T) {
	for _, cached := range [2]bool{false, true} {
		text := retrievalSyntheticText()
		text.Captions[7] = ""
		// Invalid text fails before the nil fixture/key/reference can reach a trial.
		got := retrievalOperation(Input{}, nil, "synthetic-der", &text, nil, 0, 0, 2, cached, time.Now().Add(time.Hour))
		if got.Complete || got.Found != -1 || got.Failure == nil || got.Failure.Phase != "rank" || got.Failure.Code != "unknown_retrieval_text" || got.Failure.Actual != nil {
			t.Fatal("rank unknown became success or lost explicit failure")
		}
		if got.Stats.Counts != (Counts{}) || got.Checked != ([8]bool{}) || got.States != ([8]int{}) || got.Fixtures != ([8]int{}) || got.ParityChecks != 0 || got.TextRankCalls != 1 {
			t.Fatal("unknown rank ran a verifier or marked a candidate false")
		}
		if got.Stats.Timings.TotalNS == nil || *got.Stats.Timings.TotalNS < got.Stats.Timings.RankNS || got.Stats.Timings.RankNS < 0 {
			t.Fatal("unknown ranking preparation time omitted")
		}
	}
}
func TestRetrievalInvalidContextDoesNotTouchVerifier(t *testing.T) {
	text := retrievalSyntheticText()
	for _, tc := range []struct{ parent, control int }{{-1, 0}, {4, 0}, {0, -1}, {0, 4}} {
		got := retrievalOperation(Input{}, nil, "synthetic", &text, nil, 0, tc.parent, tc.control, false, time.Now().Add(time.Hour))
		if got.Complete || got.Failure == nil || got.Failure.Code != "unknown_operation_context" || got.Stats.Counts != (Counts{}) || got.States != ([8]int{}) {
			t.Fatal("invalid context authorized work")
		}
	}
}
func TestRetrievalOperationalParityUnknownPrecedesWanted(t *testing.T) {
	known := Result{Known: true, Accept: false, PolicyErrorMask: 132, ClaimsSHA: "synthetic-claim", Code: "observed"}
	if retrievalParityCode(known, known) != "" {
		t.Fatal("equal known false result was lost")
	}
	changed := known
	changed.Accept = true
	changed.PolicyErrorMask = 0
	if retrievalParityCode(changed, known) != "unknown_operational_parity" {
		t.Fatal("accept drift was not unknown")
	}
	changed = known
	changed.PolicyErrorMask = 129
	if retrievalParityCode(changed, known) != "unknown_operational_parity" {
		t.Fatal("error class drift was hidden")
	}
	changed = known
	changed.ClaimsSHA = "different"
	if retrievalParityCode(changed, known) != "unknown_operational_parity" {
		t.Fatal("claims drift was hidden")
	}
	changed = known
	changed.Known = false
	changed.Code = "unknown_payload"
	if retrievalParityCode(changed, known) != "unknown_payload" {
		t.Fatal("original unknown reason lost")
	}
	changed.Code = "unlisted-host-like-value"
	if retrievalParityCode(changed, known) != "unknown_unlisted_code" {
		t.Fatal("raw unknown code leaked")
	}
	missing := Result{}
	if retrievalParityCode(known, missing) != "unknown_operational_parity" {
		t.Fatal("missing reference authorized a mismatch label")
	}
}
func TestRetrievalRotating128ScheduleWithoutRanksOrFixtures(t *testing.T) {
	var starts [4]int
	var firstModes [4][2]int
	var controlCounts [4]int
	n := 0
	for round := 0; round < 4; round++ {
		for step := 0; step < 4; step++ {
			control := (step + round) % 4
			if step == 0 {
				starts[control]++
			}
			for mode := 0; mode < 2; mode++ {
				cached := (mode == 1) != (round%2 == 1)
				if mode == 0 {
					i := 0
					if cached {
						i = 1
					}
					firstModes[control][i]++
				}
				for parent := 0; parent < 4; parent++ {
					controlCounts[control]++
					n++
				}
			}
		}
	}
	if n != 128 || starts != [4]int{1, 1, 1, 1} || controlCounts != [4]int{32, 32, 32, 32} {
		t.Fatal("schedule did not balance all controls")
	}
	for _, modes := range firstModes {
		if modes != [2]int{2, 2} {
			t.Fatal("first cache mode was not counterbalanced")
		}
	}
}

func TestRetrievalRawTextOwnershipPredicateWithoutRanking(t *testing.T) {
	before := retrievalSyntheticText()
	current := before
	// Hashing the frozen raw envelope here does not normalize or rank its texts.
	if !retrievalTextUnmodified(before, current, retrievalFrozenTextInput) {
		t.Fatal("unchanged raw/text value rejected")
	}
	current.Requests[0] = "changed"
	if retrievalTextUnmodified(before, current, retrievalFrozenTextInput) {
		t.Fatal("request mutation missed")
	}
	current = before
	current.Captions[7] = "changed"
	if retrievalTextUnmodified(before, current, retrievalFrozenTextInput) {
		t.Fatal("caption mutation missed")
	}
	if retrievalTextUnmodified(before, before, retrievalFrozenTextInput+" ") {
		t.Fatal("raw embed mutation missed")
	}
}
