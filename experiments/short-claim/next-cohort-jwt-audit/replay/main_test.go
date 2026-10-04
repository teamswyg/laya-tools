// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"testing"
)

// These are owned synthetic audit controls, not JWT execution or source truth.
func boolPtr(v bool) *bool { return &v }
func syntheticRecord() (record, fixture) {
	f := fixture{ClaimsJSON: `{"sub":"synthetic"}`}
	f.ClaimsSHA = digest([]byte(f.ClaimsJSON))
	r := record{Status: "observed", TokenPresent: boolPtr(true), Valid: boolPtr(true), ErrNil: boolPtr(true), ParseReturned: boolPtr(true), Errors: []string{}, Payload: json.RawMessage(f.ClaimsJSON), PayloadMatch: boolPtr(true), ZeroBefore: boolPtr(true), InputUnchanged: boolPtr(true), KeyUnchanged: boolPtr(true)}
	return r, f
}
func TestSyntheticUnknownIsPreserved(t *testing.T) {
	cases := []struct {
		name   string
		change func(*record)
	}{
		{"false_valid_nil_error", func(r *record) { r.Valid = boolPtr(false) }},
		{"parse_missing", func(r *record) { r.ParseReturned = nil }},
		{"claims_mutation", func(r *record) { r.InputUnchanged = boolPtr(false) }},
		{"payload_hash_contradiction", func(r *record) { r.Payload = json.RawMessage(`{"sub":"different"}`) }},
		{"unknown_error_under_policy", func(r *record) {
			r.Valid = boolPtr(false)
			r.ErrNil = boolPtr(false)
			r.Errors = []string{"expired", "invalid_claims", "unrecognized_leaf"}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, f := syntheticRecord()
			c.change(&r)
			state, reasons := recordState(r, f)
			if state != "U" || len(reasons) == 0 {
				t.Fatalf("tamper must stay U, got %q", state)
			}
		})
	}
}
func TestKnownMismatchDoesNotEraseUnknownTrace(t *testing.T) {
	state, u := aggregate([5]string{"F", "U", "T", "T", "T"}, [5]bool{true, true, true, true, true})
	if state != "F" || u != 1 {
		t.Fatal("known F must retain one U trace")
	}
	state, u = aggregate([5]string{"T", "U", "T", "T", "T"}, [5]bool{true, true, true, true, true})
	if state != "U" || u != 1 {
		t.Fatal("missing observation must not synthesize T")
	}
}
func TestMissingRecordRejectedBeforeAggregation(t *testing.T) {
	o := observations{Schema: "riido-jwt-direct-observer-output-v1", Revision: revision, InputSHA: inputSHA, KeySHA: "synthetic", Clock: fixedClock, Precision: 1000000000, Complete: true, Unchanged: true, Distinct: 13, Positions: 20, Planned: 160, Records: make([]record, 159)}
	if qualify(o, observerInput{}, "synthetic") == nil {
		t.Fatal("159 records must not become a complete matrix")
	}
}
func TestStrictJSONRefusesAmbiguousScalarsAndDuplicateKeys(t *testing.T) {
	for _, b := range []string{`{"x":1,"\u0078":2}`, `"\ud800"`, string([]byte{'"', 0xff, '"'}), `{} []`} {
		var target json.RawMessage
		if strict([]byte(b), &target) == nil {
			t.Fatal("ambiguous JSON accepted")
		}
	}
}

func TestFreshUnknownStatusDoesNotEchoCallerData(t *testing.T) {
	if publicStatus("caller-controlled-value") != "unknown_observer" {
		t.Fatal("raw status escaped")
	}
	if publicStatus("observed") != "observed" {
		t.Fatal("known status lost")
	}
}

func TestMissingErrorChannelAndVerifiedTokenRemainUnknown(t *testing.T) {
	r, f := syntheticRecord()
	r.Errors = nil
	if got, _ := recordState(r, f); got != "U" {
		t.Fatal("missing channel accepted")
	}
	r, f = syntheticRecord()
	r.TokenPresent = boolPtr(false)
	r.Valid = boolPtr(false)
	r.ErrNil = boolPtr(false)
	r.Errors = []string{"expired", "invalid_claims"}
	if got, _ := recordState(r, f); got != "U" {
		t.Fatal("nil verified token rejected as knownF")
	}
}
func TestNormalizedClassComparisonAndNoRawEcho(t *testing.T) {
	if sameClasses([]string{"expired"}, []string{"invalid_audience"}) {
		t.Fatal("changed cause erased")
	}
	if !sameClasses([]string{"expired", "invalid_claims"}, []string{"invalid_claims", "expired"}) {
		t.Fatal("set ordering changed meaning")
	}
	got := publicErrorClasses([]string{"caller-controlled-class"})
	if len(got) != 1 || got[0] != "unknown_error_class" {
		t.Fatal("raw unknown class escaped")
	}
}
