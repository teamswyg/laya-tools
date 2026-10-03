// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
// Package data35 copies pinned Root supervision. It does not determine truth.
package data35

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
)

const PreviousBytes = 45390
const PreviousSHA = "7b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919"
const QualificationBytes = 17473
const QualificationSHA = "024650dc9cb1e9c1c965aea40a8d31de31ebe459b645a24836eaa358d69b194f"
const MetadataBytes = 4339
const MetadataSHA = "d3f4b4ecc8a33841d553435d81cf39854a8b3821e746ed90d710c2c3db20572a"
const ComparisonBytes = 106426
const ComparisonSHA = "394017f724d59b6fd13f92e6ada53c20ccd9425f9a3d6b48ccb8dc8835ac87f4"
const MaxData = 128 << 10
const MaxRow = 16 << 10
const MaxReceipt = 32 << 10
const MaxInput = 1 << 20

type Error string

func (e Error) Error() string { return string(e) }
func code(s string) error     { return Error(s) }
func FailureCode(e error) string {
	if v, ok := e.(Error); ok {
		return string(v)
	}
	if e != nil {
		return "data35_reader_callback_failed"
	}
	return ""
}
func Digest(b []byte) string                { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func pinned(b []byte, n int, h string) bool { return len(b) == n && Digest(b) == h }

type Pin struct {
	Bytes int    `json:"bytes"`
	SHA   string `json:"sha256"`
}
type Scope struct {
	Inputs       int    `json:"input_count"`
	Observations int    `json:"candidate_observations"`
	Description  string `json:"description"`
	All          *bool  `json:"all_input_guarantee"`
	Unseen       *bool  `json:"unseen_source_guarantee"`
}
type Candidate struct {
	ID     string `json:"metadata_id"`
	Text   string `json:"text"`
	Label  *bool  `json:"label"`
	Weight *int   `json:"sample_weight"`
}
type Row struct {
	Schema       string      `json:"schema"`
	ID           string      `json:"stable_id"`
	Request      string      `json:"request"`
	Candidates   []Candidate `json:"candidates"`
	Role         string      `json:"role"`
	Group        int         `json:"whole_group"`
	Family       string      `json:"source_family"`
	Revision     string      `json:"source_revision"`
	Scope        Scope       `json:"finite_scope"`
	TextRevision string      `json:"text_revision"`
	Policy       string      `json:"feature_policy"`
}
type Unknown struct {
	Fixture int    `json:"fixture"`
	Ordinal int    `json:"ordinal"`
	Key     string `json:"key"`
	Reason  string `json:"reason"`
}
type Slot struct {
	Position *int      `json:"source_candidate_position"`
	Internal *int      `json:"internal_implementation_position"`
	ID       string    `json:"metadata_id"`
	Text     string    `json:"text"`
	Label    *bool     `json:"label"`
	Weight   *int      `json:"sample_weight"`
	Selected *bool     `json:"selected"`
	Ordinals []int     `json:"row_ordinals"`
	Vector   []string  `json:"row_vector"`
	Known    []int     `json:"known_counterexample_fixture_positions"`
	Unknown  []int     `json:"aggregate_unknown_fixture_positions"`
	Retained []Unknown `json:"unknown_predicates_retained"`
}
type Adoption struct {
	ID           string `json:"stable_id"`
	Request      string `json:"request"`
	Family       string `json:"source_family"`
	Repository   string `json:"upstream_repository"`
	Revision     string `json:"source_revision"`
	Group        int    `json:"whole_group"`
	Role         string `json:"role"`
	TextRevision string `json:"text_revision"`
	Scope        Scope  `json:"finite_scope"`
	Candidates   []Slot `json:"candidates"`
	Selected     []int  `json:"selected_candidate_positions"`
	Excluded     []int  `json:"excluded_candidate_positions"`
}
type AuthorityPool struct {
	Requests int `json:"requests"`
	Labels   int `json:"candidate_labels"`
	Positive int `json:"positive_labels"`
	Negative int `json:"negative_labels"`
}
type Qualification struct {
	Schema    string `json:"schema"`
	State     string `json:"state"`
	Authority string `json:"authority"`
	Basis     struct {
		Comparison Pin `json:"saved_comparison"`
	} `json:"basis"`
	Before            AuthorityPool `json:"pool_before"`
	After             AuthorityPool `json:"pool_after_if_materialized"`
	Requests          []Adoption    `json:"requests"`
	UnknownNeverFalse *bool         `json:"unknown_never_false"`
	Coexist           *bool         `json:"known_F_and_U_coexist_preserved"`
	Roles             *bool         `json:"whole_source_roles_preserved"`
	Immutable         *bool         `json:"old33_data_immutable"`
	Retained          int           `json:"selected_retained_unknown_predicates"`
	NewFamilies       int           `json:"new_family_assignments"`
	NewRequests       int           `json:"new_qualified_requests"`
	NewLabels         int           `json:"new_labels_adopted"`
	Reexecution       int           `json:"new_original_reexecution"`
	Materialized      int           `json:"new_materialized_data_rows"`
	Model             int           `json:"new_Fit_or_model"`
	Protected         *bool         `json:"protected_evaluation"`
}
type TFU struct {
	T int `json:"satisfied"`
	F int `json:"unsatisfied"`
	U int `json:"unknown"`
}
type Predicate struct {
	Key    string          `json:"key"`
	Want   json.RawMessage `json:"Want"`
	Got    json.RawMessage `json:"Got"`
	State  string          `json:"state"`
	Reason string          `json:"reason"`
}
type SavedRow struct {
	Ordinal    int         `json:"ordinal"`
	Request    int         `json:"request_index"`
	Fixture    int         `json:"fixture_index"`
	Display    int         `json:"display_position"`
	Internal   int         `json:"internal_position"`
	RequestID  string      `json:"request_id"`
	FixtureID  string      `json:"fixture_id"`
	Count      int         `json:"predicate_count"`
	Predicates []Predicate `json:"predicates"`
	Counts     TFU         `json:"predicate_counts"`
	State      string      `json:"state"`
	Label      *bool       `json:"label"`
	Weight     *int        `json:"sample_weight"`
}
type Comparison struct {
	Schema          string          `json:"schema"`
	State           string          `json:"state"`
	Rows            []SavedRow      `json:"rows"`
	RowCounts       TFU             `json:"row_counts"`
	PredicateCounts TFU             `json:"predicate_counts"`
	Journal         *bool           `json:"journal_content_read"`
	Initialization  *int            `json:"initialization_calls"`
	Nested          *int            `json:"nested_original_calls"`
	Labels          int             `json:"labels_assigned"`
	Weights         int             `json:"weights_assigned"`
	Roles           int             `json:"roles_assigned"`
	Calls           int             `json:"original_model_Fit_calls"`
	Authority       json.RawMessage `json:"qualification_authority"`
}
type Retained struct {
	Request  string `json:"request_id"`
	Position int    `json:"source_candidate_position"`
	Fixture  int    `json:"fixture"`
	Key      string `json:"key"`
	Ordinal  *int   `json:"ordinal"`
	Reason   string `json:"reason,omitempty"`
}
type Prior struct {
	Schema       string          `json:"schema"`
	State        string          `json:"state"`
	Bytes        int             `json:"data_bytes"`
	SHA          string          `json:"data_sha256"`
	False        *bool           `json:"unknown_converted_to_false"`
	WholeHistory *bool           `json:"whole_history_unknown_total_claimed"`
	Retained     []Retained      `json:"post23_plus_new_selected_unknown_predicates"`
	Excluded     json.RawMessage `json:"new_excluded_unknown_candidates"`
	UnknownScope struct {
		Selected       int   `json:"post23_plus_new_selected"`
		NoWholeHistory *bool `json:"whole_history_unknown_total_not_claimed"`
	} `json:"retained_unknown_metadata_scope"`
}
type Inputs struct {
	Previous, Metadata, Qualification, Comparison []byte
	ExpectedQualificationSHA                      string
}
type Config struct {
	Schema                   string `json:"schema"`
	Previous                 string `json:"previous_data"`
	Metadata                 string `json:"prior_metadata"`
	Qualification            string `json:"qualification"`
	ExpectedQualificationSHA string `json:"qualification_sha256"`
	Comparison               string `json:"comparison"`
}
type LoadedCandidate struct {
	ID, Text string
	Label    bool
	Weight   int
}
type Loaded struct {
	Valid, UnusedZero, ReaderReturned, ValidateAttempted                                            bool
	Schema, Provenance, ID, Request, Role, Family, Revision, TextRevision, Policy, ScopeDescription string
	Group, Inputs, Observations, Count, SupervisionCount                                            int
	All, Unseen                                                                                     bool
	Candidates                                                                                      [8]LoadedCandidate
}
type Pool struct {
	Requests          int    `json:"requests"`
	Labels            int    `json:"known_labels"`
	Positive          int    `json:"positive"`
	Negative          int    `json:"negative"`
	Inputs            int    `json:"input_fixtures"`
	Original          int    `json:"original_evidence_observations"`
	Selected          int    `json:"selected_candidate_observations"`
	PositivePositions [3]int `json:"positive_selected_position_histogram"`
}
type Receipt struct {
	Schema            string          `json:"schema"`
	State             string          `json:"state"`
	Failure           string          `json:"failure_code"`
	ConfigSHA         string          `json:"config_sha256"`
	DataBytes         int             `json:"data_bytes"`
	DataSHA           string          `json:"data_sha256"`
	DataDurable       bool            `json:"data_file_sync_and_directory_sync_returned"`
	PreviousPin       Pin             `json:"previous33_pin"`
	MetadataPin       Pin             `json:"prior33_metadata_pin"`
	QualificationPin  Pin             `json:"Root_qualification_pin"`
	ComparisonPin     Pin             `json:"saved_comparison_pin"`
	PinsVerified      bool            `json:"input_pins_verified"`
	AdoptionBound     bool            `json:"external_Root_adoption_bound"`
	PrefixExact       bool            `json:"previous33_prefix_byte_exact"`
	SupervisionCopied bool            `json:"Root_supervision_copied"`
	Pool              Pool            `json:"pool"`
	Retained          []Retained      `json:"post23_plus_selected33_plus_new35_unknown_predicates"`
	PriorExcluded     json.RawMessage `json:"prior33_excluded_unknown_candidates_preserved"`
	PriorUnknown      int             `json:"existing33_post23_scope_selected_unknown_predicates"`
	NewUnknown        int             `json:"new_selected_unknown_predicates"`
	ScopedUnknown     int             `json:"post23_plus_selected33_plus_new35_unknown_predicates_count"`
	WholeHistory      bool            `json:"whole_history_unknown_total_claimed"`
	UnknownFalse      bool            `json:"unknown_converted_to_false"`
	ReaderAttempts    int             `json:"LoadDevelopmentRow_attempts"`
	ReaderReturns     int             `json:"LoadDevelopmentRow_returns"`
	ReaderMatches     int             `json:"LoadDevelopmentRow_matches"`
	ExtraValidate     int             `json:"additional_ValidateInput_calls_in_reader_bridge"`
	Assigned          int             `json:"new_truth_roles_labels_weights_assigned_by_helper"`
	Original          int             `json:"original_calls"`
	Features          int             `json:"Features_calls"`
	Score             int             `json:"Score_calls"`
	Project           int             `json:"Project_calls"`
	Fit               int             `json:"Fit_calls"`
	Model             int             `json:"model_calls"`
	Initialization    *int            `json:"original_initialization_calls"`
	Nested            *int            `json:"nested_original_calls"`
	Scope             string          `json:"scope"`
}

var requestIDs = [2]string{"source-next60-strict-json-lines-prefix", "source-next60-reject-invalid-utf8-append"}
var candidateIDs = [2][3]string{{"original_ForEachLine", "validate_all_before_callbacks", "per_line_validate_then_callback"}, {"preflight_utf8_before_write", "quote_before_utf8_check", "original_AppendJSONString"}}
var fixtures = [2]int{5, 6}
var bases = [2]int{0, 15}
var internalOrder = [2][3]int{{0, 2, 1}, {1, 2, 0}}

func NewReceipt(configSHA string) Receipt {
	return Receipt{Schema: "riido-development35-correspondence-v1", State: "reserved_unvalidated", ConfigSHA: configSHA,
		PreviousPin: Pin{PreviousBytes, PreviousSHA}, MetadataPin: Pin{MetadataBytes, MetadataSHA}, QualificationPin: Pin{QualificationBytes, QualificationSHA}, ComparisonPin: Pin{ComparisonBytes, ComparisonSHA},
		Scope: "Pinned Root development supervision and unchanged project Reader value correspondence only. Requires successful CLI exit and Root durable receipt readback; a surviving receipt alone is not completion authority. Unknown scope starts after23; older original artifacts remain preserved. No whole-history count, model quality, Fit or training authorization."}
}
func sources(in Inputs, receipt *Receipt) (Qualification, Comparison, Prior, error) {
	var q Qualification
	var c Comparison
	var p Prior
	// External SHA authorization is checked before decoding or accessing labels.
	if in.ExpectedQualificationSHA != QualificationSHA {
		return q, c, p, code("external_Root_qualification_pin")
	}
	for _, v := range []struct {
		b []byte
		n int
		h string
	}{{in.Previous, PreviousBytes, PreviousSHA}, {in.Metadata, MetadataBytes, MetadataSHA}, {in.Qualification, QualificationBytes, QualificationSHA}, {in.Comparison, ComparisonBytes, ComparisonSHA}} {
		if !pinned(v.b, v.n, v.h) {
			return q, c, p, code("fixed_input_pin")
		}
	}
	receipt.PinsVerified = true
	if e := decode(in.Qualification, &q); e != nil {
		return q, c, p, e
	}
	if q.Schema != "riido-Root-gjson-two-finite-qualification-v1" || q.State != "Root_adopted_two_finite_requests_materialization_separate" || !strings.HasPrefix(q.Authority, "Root ") || len(q.Requests) != 2 || !yes(q.UnknownNeverFalse) || !yes(q.Coexist) || !yes(q.Roles) || !yes(q.Immutable) || !no(q.Protected) || q.NewFamilies != 0 || q.NewRequests != 2 || q.NewLabels != 6 || q.Reexecution != 0 || q.Materialized != 0 || q.Model != 0 || q.Retained != 36 || q.Basis.Comparison != (Pin{ComparisonBytes, ComparisonSHA}) || q.Before != (AuthorityPool{33, 96, 33, 63}) || q.After != (AuthorityPool{35, 102, 35, 67}) {
		return q, c, p, code("adoption_contract")
	}
	receipt.AdoptionBound = true
	if e := decode(in.Metadata, &p); e != nil {
		return q, c, p, e
	}
	if p.Schema != "riido-native-three-development33-correspondence-v1" || p.State != "materialized_Root_adopted_data_only" || p.Bytes != PreviousBytes || p.SHA != PreviousSHA || !no(p.False) || !no(p.WholeHistory) || !yes(p.UnknownScope.NoWholeHistory) || p.UnknownScope.Selected != 10 || len(p.Retained) != 10 || len(p.Excluded) == 0 || bytes.Equal(p.Excluded, []byte("null")) {
		return q, c, p, code("prior_unknown_scope")
	}
	if e := decode(in.Comparison, &c); e != nil {
		return q, c, p, e
	}
	if e := comparisonShape(c); e != nil {
		return q, c, p, e
	}
	return q, c, p, nil
}
func add(v *TFU, s string) error {
	switch s {
	case "T":
		v.T++
	case "F":
		v.F++
	case "U":
		v.U++
	default:
		return code("saved_state")
	}
	return nil
}
func overall(v TFU) string {
	if v.F > 0 {
		return "F"
	}
	if v.U > 0 {
		return "U"
	}
	return "T"
}
func word(s string) string {
	switch s {
	case "T":
		return "satisfied"
	case "F":
		return "unsatisfied"
	case "U":
		return "unknown"
	}
	return ""
}
func comparisonShape(c Comparison) error {
	if c.Schema != "riido-gjson-two-saved-finite-comparison-v1" || c.State != "complete_saved_comparison" || len(c.Rows) != 33 || !no(c.Journal) || c.Initialization != nil || c.Nested != nil || c.Labels != 0 || c.Weights != 0 || c.Roles != 0 || c.Calls != 0 || !bytes.Equal(c.Authority, []byte("null")) {
		return code("comparison_contract")
	}
	var rows, preds TFU
	for i, r := range c.Rows {
		req := 0
		if i >= 15 {
			req = 1
		}
		f := (i - bases[req]) / 3
		d := (i - bases[req]) % 3
		count := 8
		if req == 0 {
			count = [5]int{14, 11, 8, 11, 11}[f]
		}
		if r.Ordinal != i || r.Request != req || r.Fixture != f || r.Display != d || r.Internal != internalOrder[req][d] || r.RequestID != requestIDs[req] || r.Count != count || len(r.Predicates) != 17 || r.Label != nil || r.Weight != nil || r.FixtureID == "" {
			return code("saved_row_order_shape")
		}
		var v TFU
		for j, p := range r.Predicates {
			if j >= r.Count {
				if p.Key != "" || p.State != "" || p.Reason != "" || !bytes.Equal(p.Want, []byte("null")) || !bytes.Equal(p.Got, []byte("null")) {
					return code("saved_padding")
				}
				continue
			}
			if p.Key == "" || p.Reason == "" {
				return code("saved_predicate")
			}
			for _, old := range r.Predicates[:j] {
				if old.Key == p.Key {
					return code("saved_duplicate_predicate")
				}
			}
			if e := add(&v, p.State); e != nil {
				return e
			}
			if e := add(&preds, p.State); e != nil {
				return e
			}
		}
		if v != r.Counts || v.T+v.F+v.U != r.Count || r.State != overall(v) {
			return code("saved_row_counts")
		}
		if e := add(&rows, r.State); e != nil {
			return e
		}
	}
	if rows != c.RowCounts || preds != c.PredicateCounts || rows != (TFU{18, 9, 6}) || preds != (TFU{255, 18, 36}) {
		return code("saved_totals")
	}
	return nil
}
func checkRequest(a Adoption, req int) error {
	if req < 0 || req >= len(requestIDs) {
		return code("request_index_bound")
	}
	if a.ID != requestIDs[req] || a.Family != "tidwall-gjson" || a.Repository != "tidwall/gjson" || a.Revision != "9378d3bb93e20854e1677e0e0248e1d71ae5712f" || a.Group != 83 || a.Role != "development_train" || a.TextRevision != "gjson-two-finite-v1" || !textOK(a.Request) || a.Scope.Inputs != fixtures[req] || a.Scope.Observations != 3*fixtures[req] || len(a.Candidates) != 3 || !eqInts(a.Selected, []int{0, 1, 2}) || a.Excluded == nil || len(a.Excluded) != 0 {
		return code("request_family_role_selection")
	}
	return nil
}
func checkSlot(a Adoption, s Slot, request, position int, c Comparison) error {
	if request < 0 || request >= 2 || position < 0 || position >= 3 {
		return code("candidate_index_bound")
	}
	n := fixtures[request]
	if s.Position == nil || *s.Position != position || s.Internal == nil || *s.Internal != internalOrder[request][position] || s.ID != candidateIDs[request][position] || !textOK(s.Text) || s.Label == nil || s.Weight == nil || *s.Weight != 1 || !yes(s.Selected) || len(s.Ordinals) != n || len(s.Vector) != n || s.Known == nil || s.Unknown == nil || s.Retained == nil {
		return code("candidate_adoption_presence")
	}
	known := make([]int, 0, n)
	unknown := make([]int, 0, n)
	retained := make([]Unknown, 0, 36)
	allT := true
	for f := 0; f < n; f++ {
		ord := bases[request] + 3*f + position
		if ord < 0 || ord >= len(c.Rows) {
			return code("candidate_ordinal_bound")
		}
		r := c.Rows[ord]
		if s.Ordinals[f] != ord || s.Vector[f] != word(r.State) {
			return code("candidate_vector")
		}
		if r.State != "T" {
			allT = false
		}
		if r.State == "F" {
			known = append(known, f)
		}
		if r.State == "U" {
			unknown = append(unknown, f)
		}
		for _, p := range r.Predicates[:r.Count] {
			if p.State == "U" {
				retained = append(retained, Unknown{f, ord, p.Key, p.Reason})
			}
		}
	}
	if !eqInts(s.Known, known) || !eqInts(s.Unknown, unknown) || !reflect.DeepEqual(s.Retained, retained) {
		return code("candidate_counterexample_unknown_correspondence")
	}
	// False is Root's aggregate finite label: known counterexample plus U is allowed.
	// Unknown-only candidates cannot receive a false label or be silently weighted.
	if *s.Label {
		if !allT || len(retained) != 0 {
			return code("positive_requires_all_known_T")
		}
	} else if len(known) == 0 {
		return code("negative_requires_known_F")
	}
	return nil
}
func parseRows(raw []byte, want int) ([35]Row, error) {
	var rows [35]Row
	if want < 1 || want > len(rows) || len(raw) < 1 || len(raw) > MaxData || raw[len(raw)-1] != '\n' {
		return rows, code("data_bounds_LF")
	}
	scan := bufio.NewScanner(bytes.NewReader(raw))
	scan.Buffer(make([]byte, 1024), MaxRow+1)
	n := 0
	for scan.Scan() {
		if n >= want {
			return rows, code("data_row_count")
		}
		r, e := readRow(scan.Bytes())
		if e != nil {
			return rows, e
		}
		for _, old := range rows[:n] {
			if old.ID == r.ID {
				return rows, code("data_duplicate_parent")
			}
		}
		rows[n] = r
		n++
	}
	if scan.Err() != nil || n != want {
		return rows, code("data_row_bound_count")
	}
	return rows, nil
}
func pool(rows [35]Row, n int, original int) (Pool, error) {
	var p Pool
	p.Original = original
	for _, r := range rows[:n] {
		p.Requests++
		p.Inputs += r.Scope.Inputs
		p.Selected += r.Scope.Observations
		p.Labels += len(r.Candidates)
		positives := 0
		for j, c := range r.Candidates {
			if c.Label == nil || c.Weight == nil || *c.Weight != 1 {
				return p, code("pool_supervision")
			}
			if *c.Label {
				p.Positive++
				positives++
				if j >= 3 {
					return p, code("positive_position_bound")
				}
				p.PositivePositions[j]++
			} else {
				p.Negative++
			}
		}
		if positives != 1 {
			return p, code("one_finite_positive_per_current_request")
		}
	}
	return p, nil
}
func Compose(in Inputs, receipt *Receipt) ([]byte, [35]Row, error) {
	var zero [35]Row
	if receipt == nil {
		return nil, zero, code("receipt_required")
	}
	q, c, p, e := sources(in, receipt)
	if e != nil {
		return nil, zero, e
	}
	rows, e := parseRows(in.Previous, 33)
	if e != nil {
		return nil, zero, e
	}
	before, e := pool(rows, 33, 476)
	if e != nil {
		return nil, zero, e
	}
	if before.Requests != 33 || before.Labels != 96 || before.Positive != 33 || before.Negative != 63 || before.Inputs != 160 || before.Selected != 467 || before.PositivePositions != ([3]int{4, 28, 1}) {
		return nil, zero, code("previous33_pool")
	}
	data := append([]byte(nil), in.Previous...)
	retained := append([]Retained(nil), p.Retained...)
	for req, a := range q.Requests {
		if e = checkRequest(a, req); e != nil {
			return nil, zero, e
		}
		for _, old := range rows[:33+req] {
			if old.ID == a.ID {
				return nil, zero, code("new_parent_duplicate")
			}
		}
		row := Row{Schema: "riido-finite-development-request-row-v1", ID: a.ID, Request: a.Request, Role: a.Role, Group: a.Group, Family: a.Family, Revision: a.Revision, Scope: a.Scope, TextRevision: a.TextRevision, Policy: "request_and_candidates_text_only", Candidates: make([]Candidate, 0, 3)}
		for d, s := range a.Candidates {
			if e = checkSlot(a, s, req, d, c); e != nil {
				return nil, zero, e
			}
			label, weight := *s.Label, *s.Weight
			row.Candidates = append(row.Candidates, Candidate{s.ID, s.Text, &label, &weight})
			for _, u := range s.Retained {
				ordinal := u.Ordinal
				retained = append(retained, Retained{a.ID, d, u.Fixture, u.Key, &ordinal, u.Reason})
			}
		}
		raw, err := json.Marshal(row)
		if err != nil {
			return nil, zero, code("row_encode")
		}
		checked, err := readRow(raw)
		if err != nil {
			return nil, zero, err
		}
		rows[33+req] = checked
		if len(data)+len(raw)+1 > MaxData {
			return nil, zero, code("data_output_cap")
		}
		data = append(data, raw...)
		data = append(data, '\n')
	}
	after, e := pool(rows, 35, 509)
	if e != nil {
		return nil, zero, e
	}
	if after != (Pool{35, 102, 35, 67, 171, 509, 500, [3]int{5, 28, 2}}) || len(retained) != 46 {
		return nil, zero, code("new_pool_unknown_scope")
	}
	if !bytes.Equal(data[:PreviousBytes], in.Previous) {
		return nil, zero, code("previous_prefix")
	}
	receipt.PrefixExact = true
	receipt.SupervisionCopied = true
	receipt.Pool = after
	receipt.Retained = retained
	receipt.PriorExcluded = append(json.RawMessage(nil), p.Excluded...)
	receipt.PriorUnknown = 10
	receipt.NewUnknown = 36
	receipt.ScopedUnknown = 46
	receipt.DataBytes = len(data)
	receipt.DataSHA = Digest(data)
	receipt.State = "composed_Reader_pending"
	return data, rows, nil
}
func matchLoaded(r Row, v Loaded) bool {
	if len(r.Candidates) < 1 || len(r.Candidates) > 8 {
		return false
	}
	if !v.Valid || !v.UnusedZero || v.Schema != "riido-short-behavior-claim-v1" || v.Provenance != "finite-development-v1" || v.ID != r.ID || v.Request != r.Request || v.Role != r.Role || v.Family != r.Family || v.Revision != r.Revision || v.TextRevision != r.TextRevision || v.Policy != r.Policy || v.Group != r.Group || v.Inputs != r.Scope.Inputs || v.Observations != r.Scope.Observations || v.ScopeDescription != r.Scope.Description || v.All || v.Unseen || v.Count != len(r.Candidates) || v.SupervisionCount != len(r.Candidates) {
		return false
	}
	for j, c := range r.Candidates {
		if c.Label == nil || c.Weight == nil || v.Candidates[j] != (LoadedCandidate{c.ID, c.Text, *c.Label, *c.Weight}) {
			return false
		}
	}
	for _, c := range v.Candidates[len(r.Candidates):] {
		if c != (LoadedCandidate{}) {
			return false
		}
	}
	return true
}
func VerifyReader(data []byte, rows [35]Row, receipt *Receipt, load func([]byte) (Loaded, error)) error {
	if receipt == nil || load == nil || !receipt.AdoptionBound || !receipt.PrefixExact || receipt.ReaderAttempts != 0 || receipt.ReaderReturns != 0 || receipt.ReaderMatches != 0 || len(data) != receipt.DataBytes || Digest(data) != receipt.DataSHA {
		return code("reader_bound_input")
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 1024), MaxRow+1)
	n := 0
	for scan.Scan() {
		if n >= 35 {
			return code("reader_count")
		}
		receipt.ReaderAttempts++
		v, e := load(scan.Bytes())
		if v.ReaderReturned {
			receipt.ReaderReturns++
		}
		if v.ValidateAttempted {
			receipt.ExtraValidate++
		}
		if e != nil {
			return code("reader_callback_failed")
		}
		if !v.ReaderReturned || !v.ValidateAttempted {
			return code("reader_bridge_state")
		}
		if !matchLoaded(rows[n], v) {
			return code("reader_value_correspondence")
		}
		receipt.ReaderMatches++
		n++
	}
	if scan.Err() != nil || n != 35 {
		return code("reader_count")
	}
	receipt.State = "complete_data_and_project_Reader_correspondence"
	return nil
}
func pathOK(p string) bool {
	return filepath.IsAbs(p) && filepath.Clean(p) == p && !strings.ContainsAny(p, "\x00\r\n")
}
func ReadPinned(path string, n int, h string) ([]byte, error) {
	if !pathOK(path) || n < 1 || n > MaxInput || !hashOK(h, 64) {
		return nil, code("input_path_pin")
	}
	real, e := filepath.EvalSymlinks(path)
	if e != nil || real != path {
		return nil, code("input_symlink")
	}
	before, e := os.Lstat(path)
	if e != nil || !before.Mode().IsRegular() || before.Size() != int64(n) {
		return nil, code("input_regular_size")
	}
	fd, e := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if e != nil {
		return nil, code("input_open")
	}
	f := os.NewFile(uintptr(fd), path)
	defer f.Close()
	opened, e := f.Stat()
	if e != nil || !os.SameFile(before, opened) {
		return nil, code("input_identity")
	}
	raw, e := io.ReadAll(io.LimitReader(f, int64(n)+1))
	if e != nil || !pinned(raw, n, h) {
		return nil, code("input_sha")
	}
	after, e := os.Lstat(path)
	if e != nil || !os.SameFile(before, after) || after.Size() != before.Size() || !after.ModTime().Equal(before.ModTime()) {
		return nil, code("input_changed")
	}
	return raw, nil
}
func LoadConfig(path, externalSHA string) (Config, error) {
	var c Config
	if !pathOK(path) || !hashOK(externalSHA, 64) {
		return c, code("config_arguments")
	}
	st, e := os.Lstat(path)
	if e != nil || !st.Mode().IsRegular() || st.Size() < 1 || st.Size() > 8192 {
		return c, code("config_regular_cap")
	}
	raw, e := ReadPinned(path, int(st.Size()), externalSHA)
	if e != nil {
		return c, e
	}
	if e = decode(raw, &c); e != nil {
		return c, e
	}
	if e = exactKeys(raw, []string{"schema", "previous_data", "prior_metadata", "qualification", "qualification_sha256", "comparison"}); e != nil {
		return c, e
	}
	if c.Schema != "riido-development35-config-v1" || !hashOK(c.ExpectedQualificationSHA, 64) {
		return c, code("config_Root_adoption_pin")
	}
	for _, p := range []string{c.Previous, c.Metadata, c.Qualification, c.Comparison} {
		if !pathOK(p) {
			return c, code("config_input_paths")
		}
	}
	return c, nil
}
func Load(c Config) (Inputs, error) {
	in := Inputs{ExpectedQualificationSHA: c.ExpectedQualificationSHA}
	if in.ExpectedQualificationSHA != QualificationSHA {
		return in, code("external_Root_qualification_pin")
	}
	for _, p := range []struct {
		path string
		n    int
		h    string
		dst  *[]byte
	}{{c.Previous, PreviousBytes, PreviousSHA, &in.Previous}, {c.Metadata, MetadataBytes, MetadataSHA, &in.Metadata}, {c.Qualification, QualificationBytes, QualificationSHA, &in.Qualification}, {c.Comparison, ComparisonBytes, ComparisonSHA, &in.Comparison}} {
		raw, e := ReadPinned(p.path, p.n, p.h)
		if e != nil {
			return in, e
		}
		*p.dst = raw
	}
	return in, nil
}

type Reservation struct {
	dir           string
	data, receipt *os.File
	used          bool
}

func syncDir(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return code("directory_open")
	}
	defer f.Close()
	if f.Sync() != nil {
		return code("directory_sync")
	}
	return nil
}
func Reserve(out string, c Config) (*Reservation, error) {
	return reserveWithFileSync(out, c, func(f *os.File) error { return f.Sync() })
}

// The hook isolates failure controls; production supplies the real file Sync.
func reserveWithFileSync(out string, c Config, fileSync func(*os.File) error) (*Reservation, error) {
	if fileSync == nil {
		return nil, code("reservation_sync_hook")
	}
	if !pathOK(out) || out == string(filepath.Separator) {
		return nil, code("output_path")
	}
	parent := filepath.Dir(out)
	real, e := filepath.EvalSymlinks(parent)
	if e != nil || real != parent {
		return nil, code("output_parent_symlink")
	}
	for _, p := range []string{c.Previous, c.Metadata, c.Qualification, c.Comparison} {
		if p == out || strings.HasPrefix(p, out+string(filepath.Separator)) {
			return nil, code("output_input_overlap")
		}
	}
	if os.Mkdir(out, 0700) != nil {
		return nil, code("output_not_fresh")
	}
	if e = syncDir(parent); e != nil {
		return nil, e
	}
	r := &Reservation{dir: out}
	r.data, e = os.OpenFile(filepath.Join(out, "train.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return r, code("data_reservation")
	}
	r.receipt, e = os.OpenFile(filepath.Join(out, "MATERIALIZATION.v1.json"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return r, code("receipt_reservation")
	}
	if fileSync(r.data) != nil {
		return r, code("data_reservation_sync")
	}
	if fileSync(r.receipt) != nil {
		return r, code("receipt_reservation_sync")
	}
	if e = syncDir(out); e != nil {
		return r, e
	}
	return r, nil
}
func (r *Reservation) Close() {
	if r == nil {
		return
	}
	if r.data != nil {
		r.data.Close()
	}
	if r.receipt != nil {
		r.receipt.Close()
	}
}
func writeReserved(f *os.File, b []byte, cap int) error {
	if f == nil || len(b) < 1 || len(b) > cap {
		return code("output_size_or_reservation")
	}
	n, e := f.Write(b)
	if e != nil || n != len(b) {
		return code("output_write")
	}
	if f.Sync() != nil {
		return code("output_sync")
	}
	if f.Close() != nil {
		return code("output_close")
	}
	return nil
}
func (r *Reservation) Finish(data []byte, receipt *Receipt, cause error) error {
	if r == nil || r.used || r.data == nil || r.receipt == nil || receipt == nil {
		return code("output_reservation_state")
	}
	r.used = true
	if cause == nil {
		if receipt.State != "complete_data_and_project_Reader_correspondence" || receipt.ReaderMatches != 35 || receipt.DataBytes != len(data) || receipt.DataSHA != Digest(data) {
			cause = code("output_completion_state")
		} else if e := writeReserved(r.data, data, MaxData); e != nil {
			cause = e
		} else if e = syncDir(r.dir); e != nil {
			cause = e
		} else {
			receipt.DataDurable = true
		}
	}
	if cause != nil {
		receipt.State = "failed_partial_attempt_preserved"
		receipt.Failure = FailureCode(cause)
		if !receipt.DataDurable {
			receipt.DataBytes = 0
			receipt.DataSHA = ""
		}
	}
	raw, e := json.MarshalIndent(receipt, "", "  ")
	if e != nil {
		return code("receipt_encode")
	}
	raw = append(raw, '\n')
	if e = writeReserved(r.receipt, raw, MaxReceipt); e != nil {
		return e
	}
	if e = syncDir(r.dir); e != nil {
		return e
	}
	return cause
}
func Arguments(args []string, names []string) ([]string, error) {
	if len(args) != 2*len(names) {
		return nil, code("arguments")
	}
	v := make([]string, len(names))
	for i := 0; i < len(args); i += 2 {
		at := -1
		for j, n := range names {
			if args[i] == n {
				at = j
				break
			}
		}
		if at < 0 || v[at] != "" || args[i+1] == "" {
			return nil, code("arguments")
		}
		v[at] = args[i+1]
	}
	return v, nil
}
