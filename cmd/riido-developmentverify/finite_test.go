// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// These controls manipulate owned synthetic DTOs. They never read a real saved
// result, call an original function or rerun the public111 projection verifier.
func TestUnavailableChannelCannotBecomeFalse(t *testing.T) {
	p := finitePredicate{Key: "owned-unavailable", State: "unknown", Got: finiteObserved{Reason: "unavailable_or_unobserved_saved_channel"}}
	if !finiteAvailable(p) {
		t.Fatal("unknown unavailable shape rejected")
	}
	p.State = "unsatisfied"
	if finiteAvailable(p) {
		t.Fatal("unavailable channel became a known counterexample")
	}
	p.State = "unknown"
	b := false
	p.Got.Boolean = &b
	if finiteAvailable(p) {
		t.Fatal("unknown acquired a fabricated bool")
	}
}
func TestKnownAvailabilityAndConflictingChannels(t *testing.T) {
	b := false
	p := finitePredicate{State: "unsatisfied", Got: finiteObserved{Known: true, Boolean: &b}}
	if !finiteAvailable(p) {
		t.Fatal("known false channel rejected")
	}
	text := "owned"
	p.Got.Text = &text
	if finiteAvailable(p) {
		t.Fatal("multiple typed channels accepted")
	}
	p.Got.Text = nil
	p.Got.Reason = "invented"
	if finiteAvailable(p) {
		t.Fatal("known channel with unknown reason accepted")
	}
}
func TestKnownFalseAndUnknownRemainDistinct(t *testing.T) {
	if finiteAggregate(finiteCounts{F: 1, U: 1}) != "unsatisfied" || finiteAggregate(finiteCounts{U: 1}) != "unknown" || finiteAggregate(finiteCounts{T: 1}) != "satisfied" {
		t.Fatal("known counterexample and missing observation conflated")
	}
}
func ownedFiniteRow() (finiteRow, finiteGoldenRow) {
	// Literal values overlap a public finite request only as a small format
	// example; owned IDs and one synthetic predicate carry no source judgment.
	r := finiteRow{Ordinal: 0, Request: 0, Fixture: 0, Candidate: 0, RequestID: "owned-request", CaseID: "owned-case", CandidateID: "owned-candidate", Want: json.RawMessage(expectedFiniteWants[0]), Counts: finiteCounts{T: 1}, State: "satisfied"}
	p := finitePredicate{Key: "owned-key", Want: json.RawMessage(`"owned"`), Operator: "eq", Channel: "owned.text", State: "satisfied"}
	text := "owned"
	p.Got = finiteObserved{Known: true, Text: &text}
	raw, _ := json.Marshal(p)
	r.Predicates = []json.RawMessage{raw}
	x, _ := finiteCompact(raw)
	w := finiteGoldenRow{0, 0, 0, 0, "owned-request", "owned-case", "owned-candidate", 1, "satisfied", finiteCounts{T: 1}, [4]string{x}}
	return r, w
}
func TestCandidateOrderAndLiteralWantBinding(t *testing.T) {
	r, w := ownedFiniteRow()
	var out finiteSummary
	if e := checkFiniteRow(r, w, &out); e != nil || out.RowsChecked != 1 || out.PredicatesChecked != 1 {
		t.Fatal("owned row correspondence")
	}
	r.Candidate = 1
	out = finiteSummary{}
	if checkFiniteRow(r, w, &out) != failure("row_correspondence") || out.RowsChecked != 0 {
		t.Fatal("candidate permutation accepted")
	}
	r, w = ownedFiniteRow()
	r.Want = json.RawMessage(`{"a":"altered","error":null}`)
	if checkFiniteRow(r, w, &out) != failure("literal_Want") {
		t.Fatal("altered literal Want accepted")
	}
}
func TestExactArityAndRowCountState(t *testing.T) {
	r, w := ownedFiniteRow()
	r.Predicates = append(r.Predicates, r.Predicates[0])
	var out finiteSummary
	if checkFiniteRow(r, w, &out) != failure("row_correspondence") {
		t.Fatal("extra predicate silently discarded")
	}
	r, w = ownedFiniteRow()
	r.Counts = finiteCounts{F: 1}
	if checkFiniteRow(r, w, &out) != failure("row_state_counts") || out.RowsChecked != 0 {
		t.Fatal("mismatched counters accepted")
	}
}
func TestStrictJSONDuplicateUnknownAndBounds(t *testing.T) {
	for _, raw := range []string{`{"satisfied":1,"SATISFIED":1}`, `{"satisfied":1,"extra":0}`, `{"satisfied":1} {}`, strings.Repeat("[", 18) + "0" + strings.Repeat("]", 18)} {
		var c finiteCounts
		if finiteDecode([]byte(raw), &c) {
			t.Fatal("duplicate, unknown, trailing or excessive nesting accepted")
		}
	}
	var x []int
	if !finiteDecode([]byte(`[1,2]`), &x) || len(x) != 2 {
		t.Fatal("owned valid JSON")
	}
}
func TestPinMutationBeforeBookkeeping(t *testing.T) {
	var out finiteSummary
	if verifyFiniteDocument(bytes.Repeat([]byte{'x'}, finiteEvidenceBytes), &out) != failure("evidence_pin") || out.RowsChecked != 0 || out.PredicatesChecked != 0 || out.BookkeepingVerified {
		t.Fatal("unpinned input processed")
	}
}
