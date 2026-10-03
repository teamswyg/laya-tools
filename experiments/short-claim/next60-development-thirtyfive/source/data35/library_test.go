// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
// Inert controls. No actual data, Root artifacts, Reader or originals are used.
package data35

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func boolp(v bool) *bool { return &v }
func intp(v int) *int    { return &v }
func syntheticRow() Row {
	return Row{Schema: "riido-finite-development-request-row-v1", ID: "fake-parent", Request: "Choose the bounded value.", Candidates: []Candidate{{"fake-a", "Returns the bounded value.", boolp(true), intp(1)}, {"fake-b", "Returns a different value.", boolp(false), intp(1)}}, Role: "development_train", Group: 83, Family: "fake-family", Revision: "0123456789012345678901234567890123456789", Scope: Scope{1, 2, "A fixed synthetic input only.", boolp(false), boolp(false)}, TextRevision: "fake-v1", Policy: "request_and_candidates_text_only"}
}
func loaded(r Row) Loaded {
	v := Loaded{Valid: true, UnusedZero: true, ReaderReturned: true, ValidateAttempted: true, Schema: "riido-short-behavior-claim-v1", Provenance: "finite-development-v1", ID: r.ID, Request: r.Request, Role: r.Role, Family: r.Family, Revision: r.Revision, TextRevision: r.TextRevision, Policy: r.Policy, ScopeDescription: r.Scope.Description, Group: r.Group, Inputs: r.Scope.Inputs, Observations: r.Scope.Observations, Count: len(r.Candidates), SupervisionCount: len(r.Candidates)}
	for i, c := range r.Candidates {
		v.Candidates[i] = LoadedCandidate{c.ID, c.Text, *c.Label, *c.Weight}
	}
	return v
}
func fakeSlot(states [5]string, label bool) (Slot, Comparison) {
	s := Slot{Position: intp(0), Internal: intp(0), ID: candidateIDs[0][0], Text: "Returns an owned finite value.", Label: boolp(label), Weight: intp(1), Selected: boolp(true), Ordinals: make([]int, 0, 5), Vector: make([]string, 0, 5), Known: make([]int, 0, 5), Unknown: make([]int, 0, 5), Retained: make([]Unknown, 0, 5)}
	c := Comparison{Rows: make([]SavedRow, 15)}
	for f, state := range states {
		o := 3 * f
		s.Ordinals = append(s.Ordinals, o)
		s.Vector = append(s.Vector, word(state))
		p := Predicate{Key: "synthetic_channel", State: state, Reason: "synthetic_only"}
		c.Rows[o] = SavedRow{State: state, Count: 1, Predicates: []Predicate{p}}
		if state == "F" {
			s.Known = append(s.Known, f)
		}
		if state == "U" {
			s.Unknown = append(s.Unknown, f)
			s.Retained = append(s.Retained, Unknown{f, o, p.Key, p.Reason})
		}
	}
	return s, c
}
func TestAggregateKnownCounterexampleAndUnknown(t *testing.T) {
	s, c := fakeSlot([5]string{"F", "U", "T", "T", "T"}, false)
	if e := checkSlot(Adoption{}, s, 0, 0, c); e != nil {
		t.Fatal(e)
	}
	s.Retained = nil
	if checkSlot(Adoption{}, s, 0, 0, c) == nil {
		t.Fatal("unknown metadata dropped")
	}
}
func TestUnknownOnlyIsNotNegative(t *testing.T) {
	s, c := fakeSlot([5]string{"U", "T", "T", "T", "T"}, false)
	if checkSlot(Adoption{}, s, 0, 0, c) == nil {
		t.Fatal("unknown-only became false")
	}
}
func TestPositiveRequiresEveryKnownTrue(t *testing.T) {
	s, c := fakeSlot([5]string{"T", "T", "T", "T", "T"}, true)
	if e := checkSlot(Adoption{}, s, 0, 0, c); e != nil {
		t.Fatal(e)
	}
	s, c = fakeSlot([5]string{"U", "T", "T", "T", "T"}, true)
	if checkSlot(Adoption{}, s, 0, 0, c) == nil {
		t.Fatal("positive has unavailable channel")
	}
}
func TestRootFieldPresenceOrderAndWeights(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Slot)
	}{{"labelnull", func(s *Slot) { s.Label = nil }}, {"weightnull", func(s *Slot) { s.Weight = nil }}, {"weighttwo", func(s *Slot) { s.Weight = intp(2) }}, {"internal_swapped", func(s *Slot) { s.Internal = intp(2) }}, {"ordinal_swapped", func(s *Slot) { s.Ordinals[1] = 0 }}, {"vector_changed", func(s *Slot) { s.Vector[1] = "unknown" }}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, c := fakeSlot([5]string{"T", "T", "T", "T", "T"}, true)
			test.edit(&s)
			if checkSlot(Adoption{}, s, 0, 0, c) == nil {
				t.Fatal("changed adoption accepted")
			}
		})
	}
	s, c := fakeSlot([5]string{"T", "T", "T", "T", "T"}, true)
	if checkSlot(Adoption{}, s, 2, 0, c) == nil {
		t.Fatal("array bound")
	}
}
func TestRootFamilyAndRoleCannotMove(t *testing.T) {
	a := Adoption{ID: requestIDs[0], Request: "Inspect fixed inputs only.", Family: "tidwall-gjson", Repository: "tidwall/gjson", Revision: "9378d3bb93e20854e1677e0e0248e1d71ae5712f", Group: 83, Role: "development_train", TextRevision: "gjson-two-finite-v1", Scope: Scope{5, 15, "Finite only.", boolp(false), boolp(false)}, Candidates: make([]Slot, 3), Selected: []int{0, 1, 2}, Excluded: []int{}}
	if e := checkRequest(a, 0); e != nil {
		t.Fatal(e)
	}
	a.Group = 82
	if checkRequest(a, 0) == nil {
		t.Fatal("group moved")
	}
	a.Group = 83
	a.Role = "validation"
	if checkRequest(a, 0) == nil {
		t.Fatal("role moved")
	}
}
func TestRawRowMissingAndNullSupervision(t *testing.T) {
	r := syntheticRow()
	raw, _ := json.Marshal(r)
	if _, e := readRow(raw); e != nil {
		t.Fatal(e)
	}
	for _, bad := range [][]byte{bytes.Replace(raw, []byte(`"label":true`), []byte(`"label":null`), 1), bytes.Replace(raw, []byte(`"sample_weight":1`), []byte(`"sample_weight":1.0`), 1), bytes.Replace(raw, []byte(`"label":true,`), nil, 1), bytes.Replace(raw, []byte(`"schema":`), []byte(`"Schema":`), 1), bytes.Replace(raw, []byte(`"label":true`), []byte(`"label":true,"LABEL":false`), 1)} {
		if _, e := readRow(bad); e == nil {
			t.Fatal("missing/null/alias supervision accepted")
		}
	}
}
func TestPrefixAndCountControls(t *testing.T) {
	r := syntheticRow()
	raw, _ := json.Marshal(r)
	raw = append(raw, '\n')
	if _, e := parseRows(raw, 1); e != nil {
		t.Fatal(e)
	}
	if _, e := parseRows(raw[:len(raw)-1], 1); e == nil {
		t.Fatal("no terminal LF")
	}
	if _, e := parseRows(append(append([]byte(nil), raw...), raw...), 2); e == nil {
		t.Fatal("duplicate parent")
	}
	receipt := NewReceipt("")
	in := Inputs{Previous: raw, ExpectedQualificationSHA: "bad"}
	if _, _, e := Compose(in, &receipt); FailureCode(e) != "external_Root_qualification_pin" || receipt.AdoptionBound {
		t.Fatal("external authority gate did not precede parsing")
	}
	in.ExpectedQualificationSHA = QualificationSHA
	if _, _, e := Compose(in, &receipt); FailureCode(e) != "fixed_input_pin" {
		t.Fatal("altered old prefix accepted")
	}
}
func TestReaderFailureRetainsPriorCounters(t *testing.T) {
	var rows [35]Row
	var data []byte
	for i := range rows {
		r := syntheticRow()
		r.ID = "fake-" + strconv.Itoa(i)
		rows[i] = r
		raw, _ := json.Marshal(r)
		data = append(data, raw...)
		data = append(data, '\n')
	}
	receipt := NewReceipt("")
	receipt.AdoptionBound = true
	receipt.PrefixExact = true
	receipt.DataBytes = len(data)
	receipt.DataSHA = Digest(data)
	calls := 0
	e := VerifyReader(data, rows, &receipt, func(raw []byte) (Loaded, error) {
		r, err := readRow(raw)
		if err != nil {
			return Loaded{}, err
		}
		v := loaded(r)
		calls++
		if calls == 3 {
			v.Candidates[1].Label = true
		}
		return v, nil
	})
	if e == nil || calls != 3 || receipt.ReaderAttempts != 3 || receipt.ReaderReturns != 3 || receipt.ReaderMatches != 2 || receipt.ExtraValidate != 3 {
		t.Fatal("partial mismatch counters")
	}
	receipt = NewReceipt("")
	receipt.AdoptionBound = true
	receipt.PrefixExact = true
	receipt.DataBytes = len(data)
	receipt.DataSHA = Digest(data)
	e = VerifyReader(data, rows, &receipt, func(raw []byte) (Loaded, error) {
		return Loaded{ReaderReturned: true}, code("synthetic_after_Reader_before_extra_validation")
	})
	if e == nil || receipt.ReaderAttempts != 1 || receipt.ReaderReturns != 1 || receipt.ExtraValidate != 0 || receipt.ReaderMatches != 0 {
		t.Fatal("Reader return confused with later bridge failure")
	}
}
func TestReaderMustMatchAllMetadataAndUnusedTail(t *testing.T) {
	r := syntheticRow()
	v := loaded(r)
	if !matchLoaded(r, v) {
		t.Fatal("synthetic match")
	}
	v.ScopeDescription = "Changed scope."
	if matchLoaded(r, v) {
		t.Fatal("scope lost")
	}
	v = loaded(r)
	v.Candidates[7] = LoadedCandidate{ID: "hidden", Weight: 1}
	if matchLoaded(r, v) {
		t.Fatal("unused tail leak")
	}
}
func TestExclusiveReservationBeforeFailure(t *testing.T) {
	base, e := filepath.EvalSymlinks(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(base, "attempt")
	c := Config{Previous: filepath.Join(base, "fake1"), Metadata: filepath.Join(base, "fake2"), Qualification: filepath.Join(base, "fake3"), Comparison: filepath.Join(base, "fake4")}
	r, e := Reserve(out, c)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	receipt := NewReceipt("bad")
	if e = r.Finish(nil, &receipt, code("synthetic_pin_failure")); FailureCode(e) != "synthetic_pin_failure" {
		t.Fatal(e)
	}
	data, e := os.ReadFile(filepath.Join(out, "train.jsonl"))
	if e != nil || len(data) != 0 {
		t.Fatal("failure wrote data")
	}
	saved, e := os.ReadFile(filepath.Join(out, "MATERIALIZATION.v1.json"))
	if e != nil || !bytes.Contains(saved, []byte("failed_partial_attempt_preserved")) {
		t.Fatal("failure missing")
	}
	if _, e = Reserve(out, c); e == nil {
		t.Fatal("exclusive output reused")
	}
	after, _ := os.ReadFile(filepath.Join(out, "MATERIALIZATION.v1.json"))
	if !bytes.Equal(saved, after) {
		t.Fatal("old failure overwritten")
	}
}
