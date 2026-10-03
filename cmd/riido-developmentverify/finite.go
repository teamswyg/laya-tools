// SPDX-License-Identifier: Apache-2.0
// Proposed saved predicate bookkeeping integrity check. It executes no worker.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

const finiteEvidenceBytes = 71300
const finiteEvidenceSHA = "918a5c7602b0a03bb7bc8d7817fb6126601bd2b474adc6f398875adfbc944cb1"
const finiteSourceComparisonSHA = "9ab9c49189be82cd984058792658766f3eb438f52308602245c144757fa67b8c"

type finiteCounts struct {
	T int `json:"satisfied"`
	F int `json:"known_mismatches"`
	U int `json:"unknown"`
}
type finiteStates struct {
	T int `json:"satisfied"`
	F int `json:"unsatisfied"`
	U int `json:"unknown"`
}
type finitePin struct {
	Bytes int    `json:"bytes"`
	SHA   string `json:"sha256"`
}
type finiteRow struct {
	Ordinal     int               `json:"ordinal"`
	Request     int               `json:"request"`
	Fixture     int               `json:"fixture"`
	Candidate   int               `json:"candidate"`
	RequestID   string            `json:"request_id"`
	CaseID      string            `json:"case_id"`
	CandidateID string            `json:"candidate_id"`
	Want        json.RawMessage   `json:"full_literal_Want"`
	Predicates  []json.RawMessage `json:"predicates"`
	Counts      finiteCounts      `json:"predicate_counts"`
	State       string            `json:"state"`
}
type finiteDocument struct {
	Schema         string          `json:"schema"`
	State          string          `json:"state"`
	Source         finitePin       `json:"source_comparison_pin"`
	RowCount       int             `json:"row_count"`
	PredicateCount int             `json:"predicate_count"`
	States         finiteStates    `json:"predicate_states"`
	Rows           []finiteRow     `json:"rows"`
	Derivation     json.RawMessage `json:"derivation"`
}
type finiteObserved struct {
	Known   bool    `json:"known"`
	Null    bool    `json:"null"`
	NonNull bool    `json:"known_nonnull"`
	Text    *string `json:"text"`
	Boolean *bool   `json:"boolean"`
	Integer *int64  `json:"integer"`
	Reason  string  `json:"unknown_reason"`
}
type finitePredicate struct {
	Key      string          `json:"key"`
	Want     json.RawMessage `json:"Want"`
	Operator string          `json:"operator"`
	Channel  string          `json:"saved_typed_channel"`
	Got      finiteObserved  `json:"Got"`
	State    string          `json:"state"`
}
type finiteGoldenRow struct {
	Ordinal, Request, Fixture, Candidate int
	RequestID, CaseID, CandidateID       string
	PredicateCount                       int
	State                                string
	Counts                               finiteCounts
	Predicates                           [4]string
}
type finiteUnknown struct {
	Ordinal   int    `json:"ordinal"`
	Request   int    `json:"request"`
	Fixture   int    `json:"fixture"`
	Candidate int    `json:"candidate"`
	Key       string `json:"key"`
	Channel   string `json:"saved_typed_channel"`
	Reason    string `json:"unknown_reason"`
}
type finiteSummary struct {
	Schema              string           `json:"schema"`
	State               string           `json:"state"`
	Failure             string           `json:"failure_code"`
	EvidenceSHA         string           `json:"evidence_sha256"`
	SourceSHA           string           `json:"source_comparison_sha256"`
	RowsChecked         int              `json:"rows_checked"`
	PredicatesChecked   int              `json:"predicates_checked"`
	PredicateStates     finiteStates     `json:"predicate_states"`
	RowStates           finiteStates     `json:"row_states"`
	Vectors             [3][3][5]string  `json:"request_candidate_fixture_states"`
	Unknowns            [2]finiteUnknown `json:"retained_unknowns"`
	BookkeepingVerified bool             `json:"saved_predicate_bookkeeping_verified"`
	OriginalReexecution bool             `json:"original_worker_reexecution_verified"`
	SemanticRecomputed  bool             `json:"semantic_predicates_recomputed"`
	Original            int              `json:"original_candidate_calls"`
	Model               int              `json:"model_calls"`
	Features            int              `json:"Features_calls"`
	Score               int              `json:"Score_calls"`
	Project             int              `json:"Project_calls"`
	Fit                 int              `json:"Fit_calls"`
	Labels              bool             `json:"labels_assigned"`
	Qualified           bool             `json:"qualified_by_verifier"`
	Scope               string           `json:"scope"`
}

func finiteCompact(b []byte) (string, bool) {
	var x bytes.Buffer
	e := json.Compact(&x, b)
	return x.String(), e == nil
}

// JSON arity is checked with slices before indexed copies. No fixed-array JSON
// decode is allowed to silently discard extra records. Duplicate key detection
// uses a small owned stack rather than a map; comparisons are case insensitive.
func finiteWalk(d *json.Decoder, depth int, nodes *int) bool {
	if depth > 16 || *nodes >= 8192 {
		return false
	}
	*nodes = *nodes + 1
	t, e := d.Token()
	if e != nil {
		return false
	}
	v, ok := t.(json.Delim)
	if !ok {
		return true
	}
	switch v {
	case '{':
		var keys [16]string
		n := 0
		for d.More() {
			k, e := d.Token()
			s, ok := k.(string)
			if e != nil || !ok || n == len(keys) {
				return false
			}
			for i := 0; i < n; i++ {
				if strings.EqualFold(keys[i], s) {
					return false
				}
			}
			keys[n] = s
			n++
			if !finiteWalk(d, depth+1, nodes) {
				return false
			}
		}
		t, e = d.Token()
		return e == nil && t == json.Delim('}')
	case '[':
		n := 0
		for d.More() {
			if n >= 256 || !finiteWalk(d, depth+1, nodes) {
				return false
			}
			n++
		}
		t, e = d.Token()
		return e == nil && t == json.Delim(']')
	default:
		return false
	}
}
func finiteDecode(raw []byte, dst any) bool {
	if !utf8.Valid(raw) {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	nodes := 0
	if !finiteWalk(d, 0, &nodes) {
		return false
	}
	if _, e := d.Token(); !errors.Is(e, io.EOF) {
		return false
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil {
		return false
	}
	return errors.Is(d.Decode(new(any)), io.EOF)
}
func finiteAvailable(p finitePredicate) bool {
	g := p.Got
	if !g.Known {
		return p.State == "unknown" && g.Reason == "unavailable_or_unobserved_saved_channel" && !g.Null && !g.NonNull && g.Text == nil && g.Boolean == nil && g.Integer == nil
	}
	if p.State != "satisfied" && p.State != "unsatisfied" || g.Reason != "" {
		return false
	}
	n := 0
	if g.Null {
		n++
	}
	if g.NonNull {
		n++
	}
	if g.Text != nil {
		n++
	}
	if g.Boolean != nil {
		n++
	}
	if g.Integer != nil {
		n++
	}
	return n == 1
}
func finiteAggregate(c finiteCounts) string {
	if c.F > 0 {
		return "unsatisfied"
	}
	if c.U > 0 {
		return "unknown"
	}
	return "satisfied"
}
func finiteAddState(s *finiteStates, state string) bool {
	switch state {
	case "satisfied":
		s.T++
	case "unsatisfied":
		s.F++
	case "unknown":
		s.U++
	default:
		return false
	}
	return true
}

func checkFiniteRow(r finiteRow, w finiteGoldenRow, out *finiteSummary) error {
	if r.Ordinal != w.Ordinal || r.Request != w.Request || r.Fixture != w.Fixture || r.Candidate != w.Candidate || r.RequestID != w.RequestID || r.CaseID != w.CaseID || r.CandidateID != w.CandidateID || r.Request < 0 || r.Request >= 3 || r.Candidate < 0 || r.Candidate >= 3 || r.Fixture < 0 || r.Fixture >= 5 || r.Request == 2 && r.Fixture >= 4 || len(r.Predicates) != w.PredicateCount || len(r.Predicates) < 1 || len(r.Predicates) > 4 {
		return failure("row_correspondence")
	}
	wantIndex := r.Fixture + r.Request*5
	want, ok := finiteCompact(r.Want)
	if !ok || want != expectedFiniteWants[wantIndex] {
		return failure("literal_Want")
	}
	var c finiteCounts
	for i, raw := range r.Predicates {
		text, ok := finiteCompact(raw)
		if !ok || text != w.Predicates[i] {
			return failure("predicate_golden")
		}
		var p finitePredicate
		if !finiteDecode(raw, &p) || !finiteAvailable(p) {
			return failure("predicate_availability")
		}
		switch p.State {
		case "satisfied":
			c.T++
		case "unsatisfied":
			c.F++
		case "unknown":
			c.U++
		default:
			return failure("predicate_state")
		}
		if p.State == "unknown" {
			index := out.PredicateStates.U
			if index >= len(out.Unknowns) {
				return failure("unknown_count")
			}
			out.Unknowns[index] = finiteUnknown{r.Ordinal, r.Request, r.Fixture, r.Candidate, p.Key, p.Channel, p.Got.Reason}
		}
		finiteAddState(&out.PredicateStates, p.State)
		out.PredicatesChecked++
	}
	if c != r.Counts || c != w.Counts || finiteAggregate(c) != r.State || r.State != w.State {
		return failure("row_state_counts")
	}
	finiteAddState(&out.RowStates, r.State)
	out.Vectors[r.Request][r.Candidate][r.Fixture] = r.State
	out.RowsChecked++
	return nil
}
func verifyFiniteDocument(raw []byte, out *finiteSummary) error {
	if len(raw) != finiteEvidenceBytes || digest(raw) != finiteEvidenceSHA {
		return failure("evidence_pin")
	}
	var d finiteDocument
	if !finiteDecode(raw, &d) {
		return failure("evidence_json")
	}
	derivation, ok := finiteCompact(d.Derivation)
	if d.Schema != "riido-remaining-three-public-finite-predicate-evidence-v1" || d.State != "public_projection_of_saved_comparison_only" || d.Source != (finitePin{380995, finiteSourceComparisonSHA}) || d.RowCount != 42 || d.PredicateCount != 111 || d.States != (finiteStates{89, 20, 2}) || len(d.Rows) != 42 || !ok || derivation != expectedFiniteDerivation {
		return failure("evidence_header")
	}
	for i, r := range d.Rows {
		if e := checkFiniteRow(r, expectedFiniteRows[i], out); e != nil {
			return e
		}
	}
	expectedUnknown := [2]finiteUnknown{{12, 0, 4, 0, "error_kind_not", "INI.Error.input_byte_limit", "unavailable_or_unobserved_saved_channel"}, {33, 2, 1, 0, "exclusive_open", "Safe.Trace.Exclusive", "unavailable_or_unobserved_saved_channel"}}
	if out.RowsChecked != 42 || out.PredicatesChecked != 111 || out.PredicateStates != (finiteStates{89, 20, 2}) || out.RowStates != (finiteStates{29, 11, 2}) || out.Unknowns != expectedUnknown {
		return failure("evidence_counts")
	}
	for c := 0; c < 3; c++ {
		if out.Vectors[2][c][4] != "" {
			return failure("unused_fixture_slot")
		}
	}
	out.EvidenceSHA = finiteEvidenceSHA
	out.SourceSHA = finiteSourceComparisonSHA
	out.BookkeepingVerified = true
	return nil
}

// Existing Reader-core regular-file and stable-stat helpers are reused. This
// companion has no worker launch, stdout writer or separate process lifecycle.
func readFinitePinned(path string) ([]byte, error) {
	f, st, e := openRegular(path, finiteEvidenceBytes)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, finiteEvidenceBytes+1))
	if e != nil || len(raw) != finiteEvidenceBytes || !stable(f, st) {
		return nil, failure("finite_input_changed")
	}
	return raw, nil
}
func verifyFinite(path string) (finiteSummary, error) {
	out := finiteSummary{Schema: "riido-public111-predicate-integrity-v1", State: "rejected", Scope: "Exact saved public projection ordering, literal Wants, typed predicate values, counts and retained unknowns only. No original reexecution, source correctness, new semantic qualification, rights, role, training, model quality or resource authority."}
	raw, e := readFinitePinned(path)
	if e == nil {
		e = verifyFiniteDocument(raw, &out)
	}
	if e != nil {
		out.Failure = e.Error()
	} else {
		out.State = "verified_saved_predicate_bookkeeping"
	}
	return out, e
}
