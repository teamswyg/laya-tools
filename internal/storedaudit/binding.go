// SPDX-License-Identifier: Apache-2.0
// Stage-58 preparation binding for the existing public Apache-2.0 laya-tools
// 56/56b records. No original candidate implementation is imported or executed.
// Stored-record binding is not a fresh source-truth audit or a ranking result.
package storedaudit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Schema = "riido-stored-truth-preparation-binding-58-v1"
const LegacySHA256 = "0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0"
const TypedSHA256 = "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"
const ResultSHA256 = "4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4"
const (
	ParentCount   = 72
	MaxCandidates = 4
	GroupCount    = 17
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrPin          Error = "stored_truth_input_pin_mismatch"
	ErrJSON         Error = "stored_truth_invalid_json"
	ErrDuplicate    Error = "stored_truth_duplicate_json_key"
	ErrUnknownField Error = "stored_truth_unknown_json_field"
	ErrNumber       Error = "stored_truth_invalid_integer"
	ErrUnicode      Error = "stored_truth_invalid_unicode"
	ErrSchema       Error = "stored_truth_schema_mismatch"
	ErrIdentity     Error = "stored_truth_parent_candidate_binding_mismatch"
	ErrText         Error = "stored_truth_raw_text_bounds"
	ErrTruth        Error = "stored_truth_inconsistent_outcome"
	ErrSource       Error = "stored_truth_source_binding_mismatch"
	ErrGroups       Error = "stored_truth_group_membership_mismatch"
	ErrCounts       Error = "stored_truth_aggregate_count_mismatch"
)

// Text is the complete future baseline input: raw prose and original order.
// There are no labels, IDs, groups, roles, literal outcomes or source metadata.
// Only Candidates[:CandidateCount] are present. No text is truncated, rewritten
// or normalized here. The future frozen evaluator must use the original
// shortclaim validation contract, including its 32 normalized-word limit.
type Text struct {
	Request        string
	CandidateCount int
	Candidates     [MaxCandidates]string
}
type CandidateTruth struct {
	CandidateID    string `json:"candidate_id"`
	State          string `json:"state"`
	VectorsChecked int    `json:"vectors_checked"`
	FailedVectors  int    `json:"failed_vectors"`
}
type Truth struct {
	State             string
	Reason            string
	AcceptableCount   int
	AcceptableIndices [MaxCandidates]int
	CandidateCount    int
	Candidates        [MaxCandidates]CandidateTruth
}
type CandidateAudit struct{ ID, SourceID, CodeSHA256, BundleSHA256, CoreTemplate string }
type Audit struct {
	ParentID, Prototype, ContractID, Cohort string
	CandidateCount                          int
	Candidates                              [MaxCandidates]CandidateAudit
	GroupID                                 int
}
type Row struct {
	Text  Text
	Truth Truth
	Audit Audit
}

// Group IDs are the original sparse IDs, not a renumbered 0..16 sequence.
// Every member, including unknown-only parents, remains in the metadata graph.
type Group struct {
	ID            int      `json:"id"`
	Parents       []string `json:"parents"`
	Prototypes    []string `json:"prototypes"`
	CoreTemplates []string `json:"core_templates"`
	Sources       []string `json:"sources"`
}
type GroupEdge struct {
	Left    string   `json:"left"`
	Right   string   `json:"right"`
	Reasons []string `json:"reasons"`
}
type Metadata struct {
	LegacyInputSHA256, TypedInputSHA256, StoredResultSHA256 string
	GroupPolicy, Scope                                      string
	StopReasons                                             []string
	Edges                                                   []GroupEdge
	// Exact old evidence preserves provenance/controls/contracts without
	// projecting any of those fields into Text. It is not a new audit result.
	StoredEvidence json.RawMessage
}
type Dataset struct {
	Schema                                       string
	Rows                                         [ParentCount]Row
	Groups                                       [GroupCount]Group
	Answerable, NoAnswer, Unknown, LabeledGroups int
	Candidates                                   int
	CandidateCountHistogram                      [5]int
	Metadata                                     Metadata
}

// Wire types intentionally use slices rather than fixed arrays: encoding/json
// silently discards extra fixed-array elements, which must never launder input.
type wireCandidate struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	SourceID     string `json:"source_id"`
	CodeSHA256   string `json:"code_sha256"`
	BundleSHA256 string `json:"source_bundle_sha256,omitempty"`
}
type wireParent struct {
	ID         string          `json:"id"`
	Prototype  string          `json:"prototype"`
	ContractID string          `json:"contract_id"`
	Request    string          `json:"request"`
	Candidates []wireCandidate `json:"candidates"`
}
type wireProbes struct {
	Schema  string       `json:"schema"`
	Origin  string       `json:"origin"`
	Parents []wireParent `json:"parents"`
}
type wireOutcome struct {
	ParentID   string           `json:"parent_id"`
	State      string           `json:"state"`
	Acceptable []int            `json:"acceptable_candidate_indices"`
	Candidates []CandidateTruth `json:"candidate_truth"`
	Reason     string           `json:"reason,omitempty"`
}
type sourcePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type sourceArtifact struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}
type component struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	SHA256     string `json:"sha256"`
	Normalized string `json:"normalized_behavior_sha256,omitempty"`
}
type legacySource struct {
	ID         string `json:"id"`
	Prototype  string `json:"prototype"`
	Core       string `json:"core_template"`
	Code       string `json:"code_sha256"`
	Normalized string `json:"normalized_code_sha256"`
}
type typedSource struct {
	ID         string      `json:"id"`
	Prototype  string      `json:"prototype"`
	Core       string      `json:"core_template"`
	Function   string      `json:"function"`
	Code       string      `json:"code_sha256"`
	Normalized string      `json:"normalized_code_sha256"`
	Bundle     string      `json:"source_bundle_sha256"`
	Components []component `json:"semantic_components"`
	Imports    []string    `json:"standard_imports"`
}
type control struct {
	ID       string `json:"source_id"`
	Expected string `json:"expected_control"`
	Checked  int    `json:"vectors_checked"`
	Failed   int    `json:"failed_vectors"`
	TruthSHA string `json:"truth_table_sha256"`
}
type contract struct {
	Prototype string `json:"prototype"`
	Core      string `json:"core_template"`
	Semantics string `json:"input_semantics"`
	Vectors   int    `json:"literal_vectors"`
	TruthSHA  string `json:"truth_table_sha256"`
}
type commonTruth struct {
	Schema         string        `json:"schema"`
	DatasetSHA     string        `json:"typed_dataset_sha256"`
	Parents        int           `json:"parents"`
	Candidates     int           `json:"candidates"`
	Known          int           `json:"known_parents"`
	Unknown        int           `json:"unknown_parents"`
	NoAnswer       int           `json:"no_answer_parents"`
	Prototypes     int           `json:"prototype_count"`
	ReviewRequired bool          `json:"natural_language_review_required"`
	Training       bool          `json:"training_ready"`
	Partitioned    bool          `json:"partitioned"`
	Final          bool          `json:"final_eligible"`
	Outcomes       []wireOutcome `json:"outcomes"`
	Controls       []control     `json:"controls"`
	Contracts      []contract    `json:"contracts"`
	Limitations    []string      `json:"limitations"`
}
type legacyTruth struct {
	commonTruth
	SourceSHA string         `json:"source_artifact_sha256"`
	Connected int            `json:"connected_groups"`
	Minimum   int            `json:"minimum_operational_groups"`
	Meets     bool           `json:"meets_group_minimum"`
	Sources   []legacySource `json:"sources"`
	Edges     []GroupEdge    `json:"group_edges"`
	Groups    []Group        `json:"groups"`
}
type typedTruth struct {
	commonTruth
	Artifacts      []sourceArtifact `json:"source_artifacts"`
	Sources        []typedSource    `json:"sources"`
	Infrastructure string           `json:"infrastructure_policy"`
}
type combinedGroups struct {
	Schema       string      `json:"schema"`
	Parents      int         `json:"parents"`
	Prototypes   int         `json:"prototype_count"`
	Connected    int         `json:"connected_groups"`
	Minimum      int         `json:"minimum_operational_groups"`
	Meets        bool        `json:"meets_group_minimum"`
	Labeled      int         `json:"labeled_connected_groups"`
	MinLabeled   int         `json:"minimum_labeled_connected_groups"`
	MeetsLabeled bool        `json:"meets_labeled_group_minimum"`
	Training     bool        `json:"training_ready"`
	Edges        []GroupEdge `json:"group_edges"`
	Groups       []Group     `json:"groups"`
	Policy       string      `json:"policy"`
}
type storedResult struct {
	Schema             string         `json:"schema"`
	Stage              string         `json:"stage"`
	PlanSHA            string         `json:"plan_sha256"`
	InputSHA           string         `json:"input_sha256"`
	LegacySHA          string         `json:"legacy_input_sha256"`
	Runtime            string         `json:"runtime_go"`
	CPU                int            `json:"cpu_threads"`
	Heap               int            `json:"go_heap_soft_limit_bytes"`
	HeapScope          string         `json:"go_heap_limit_scope"`
	Implementation     []sourcePin    `json:"implementation_files"`
	Legacy             legacyTruth    `json:"legacy_truth"`
	Typed              typedTruth     `json:"typed_truth"`
	Combined           combinedGroups `json:"combined_groups"`
	Parents            int            `json:"parents"`
	Candidates         int            `json:"candidates"`
	VectorChecks       int            `json:"source_vector_checks"`
	LegacyVectorChecks int            `json:"legacy_source_vector_checks"`
	TypedVectorChecks  int            `json:"typed_source_vector_checks"`
	Connected          int            `json:"connected_groups"`
	Labeled            int            `json:"labeled_connected_groups"`
	Minimum            int            `json:"minimum_connected_groups"`
	MinLabeled         int            `json:"minimum_labeled_connected_groups"`
	WholeGate          bool           `json:"whole_group_gate_pass"`
	LabeledGate        bool           `json:"labeled_group_gate_pass"`
	StopReasons        []string       `json:"stop_reasons"`
	Roles              bool           `json:"roles_assigned"`
	Partitions         int            `json:"partitions"`
	Fits               int            `json:"fits"`
	Weights            int            `json:"weight_artifacts"`
	ModelCalls         int            `json:"model_calls"`
	Rankings           int            `json:"baseline_rankings"`
	Performance        int            `json:"performance_runs"`
	Training           bool           `json:"training_execution_ready"`
	Production         bool           `json:"production_activation"`
	Final              bool           `json:"final_eligible"`
	Protected          bool           `json:"protected_final_read"`
	Scope              string         `json:"scope"`
}

func sum(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func validSHA(s string) bool {
	raw, e := hex.DecodeString(s)
	return e == nil && len(raw) == 32 && s == strings.ToLower(s)
}
func validID(s string) bool {
	if len(s) == 0 || len(s) > 96 {
		return false
	}
	for _, b := range []byte(s) {
		if !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '-' || b == '_') {
			return false
		}
	}
	return true
}
func validText(s string) bool {
	return len(s) > 0 && len(s) <= 512 && utf8.ValidString(s) && strings.TrimSpace(s) != "" && !strings.ContainsRune(s, 0)
}

// strictJSON rejects duplicate decoded/case-colliding keys, unknown fields,
// invalid UTF-8/UTF-16, nulls, exponent/fraction/negative/overflow integers and
// excessive structures. No decoder diagnostics containing input text escape.
func strictJSON(raw []byte, target any) error {
	if len(raw) == 0 || len(raw) > 4<<20 || !utf8.Valid(raw) {
		return ErrJSON
	}
	if !validEscapedUnicode(raw) {
		return ErrUnicode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	budget := 100000
	if e := scanJSON(d, 0, &budget); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return ErrJSON
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	d.UseNumber()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		return ErrUnknownField
	}
	return nil
}
func scanJSON(d *json.Decoder, depth int, budget *int) error {
	if depth > 24 || *budget <= 0 {
		return ErrJSON
	}
	*budget--
	t, e := d.Token()
	if e != nil {
		return ErrJSON
	}
	switch v := t.(type) {
	case json.Delim:
		switch v {
		case '{':
			seen := make(map[string]bool)
			keys := 0
			for d.More() {
				keyToken, e := d.Token()
				if e != nil {
					return ErrJSON
				}
				key, ok := keyToken.(string)
				if !ok || len(key) > 128 {
					return ErrJSON
				}
				folded := strings.ToLower(key)
				if seen[folded] {
					return ErrDuplicate
				}
				seen[folded] = true
				keys++
				// Frozen schema keys are ASCII lowercase, digits or underscore.
				for _, b := range []byte(key) {
					if !(b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_') {
						return ErrUnknownField
					}
				}
				if keys > 64 {
					return ErrJSON
				}
				if e := scanJSON(d, depth+1, budget); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return ErrJSON
			}
		case '[':
			items := 0
			for d.More() {
				items++
				if items > 4096 {
					return ErrJSON
				}
				if e := scanJSON(d, depth+1, budget); e != nil {
					return e
				}
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return ErrJSON
			}
		default:
			return ErrJSON
		}
	case json.Number:
		s := v.String()
		if s == "" || len(s) > 1 && s[0] == '0' {
			return ErrNumber
		}
		for _, b := range []byte(s) {
			if b < '0' || b > '9' {
				return ErrNumber
			}
		}
		if _, e := strconv.ParseUint(s, 10, 64); e != nil {
			return ErrNumber
		}
	case string, bool:
	case nil:
		return ErrJSON
	default:
		return ErrJSON
	}
	return nil
}
func fourHex(raw []byte, i int) (uint64, bool) {
	if i+4 > len(raw) {
		return 0, false
	}
	n, e := strconv.ParseUint(string(raw[i:i+4]), 16, 16)
	return n, e == nil
}
func validEscapedUnicode(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		i++
		for ; i < len(raw) && raw[i] != '"'; i++ {
			if raw[i] != '\\' {
				continue
			}
			i++
			if i >= len(raw) {
				return false
			}
			if raw[i] != 'u' {
				continue
			}
			n, ok := fourHex(raw, i+1)
			if !ok {
				return false
			}
			i += 4
			if n >= 0xdc00 && n <= 0xdfff {
				return false
			}
			if n >= 0xd800 && n <= 0xdbff {
				if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
					return false
				}
				low, ok := fourHex(raw, i+3)
				if !ok || low < 0xdc00 || low > 0xdfff {
					return false
				}
				i += 6
			}
		}
		if i >= len(raw) {
			return false
		}
	}
	return true
}

// Bind is a pure stored-record consistency check, not an original candidate
// execution or a fresh truth audit. Exact three reviewed byte pins are required.
// Errors return a zero Dataset; they never create rejected/false labels. Every
// original parent and candidate remains in original order, including unknowns.
func Bind(legacyRaw, typedRaw, resultRaw []byte) (Dataset, error) {
	if len(legacyRaw) != 49941 || len(typedRaw) != 43598 || len(resultRaw) != 266818 || sum(legacyRaw) != LegacySHA256 || sum(typedRaw) != TypedSHA256 || sum(resultRaw) != ResultSHA256 {
		return Dataset{}, ErrPin
	}
	var legacy, typed wireProbes
	var result storedResult
	for _, item := range []struct {
		raw []byte
		dst any
	}{{legacyRaw, &legacy}, {typedRaw, &typed}, {resultRaw, &result}} {
		if e := strictJSON(item.raw, item.dst); e != nil {
			return Dataset{}, e
		}
	}
	if legacy.Schema != "riido-behavior-probes-v1" || legacy.Origin != "original_authored_go_microcontracts_56_apache2_development_only" || len(legacy.Parents) != 48 || typed.Schema != "riido-typed-behavior-probes-v2" || typed.Origin != "original_authored_typed_go_microcontracts_56b_apache2_development_only" || len(typed.Parents) != 24 {
		return Dataset{}, ErrSchema
	}
	if result.Schema != "riido-typed-truth-development-result-v1" || result.Stage != "preparation_truth_groups_only" || result.InputSHA != TypedSHA256 || result.LegacySHA != LegacySHA256 || !validSHA(result.PlanSHA) || result.Runtime != "go1.27.1" || result.CPU != 1 || result.Heap != 268435456 {
		return Dataset{}, ErrSchema
	}
	if result.Roles || result.Partitions != 0 || result.Fits != 0 || result.Weights != 0 || result.ModelCalls != 0 || result.Rankings != 0 || result.Performance != 0 || result.Training || result.Production || result.Final || result.Protected {
		return Dataset{}, ErrSchema
	}
	if result.Parents != 72 || result.Candidates != 216 || result.VectorChecks != 524 || result.LegacyVectorChecks != 204 || result.TypedVectorChecks != 320 || result.Connected != 17 || result.Labeled != 16 || result.Minimum != 15 || result.MinLabeled != 15 || !result.WholeGate || !result.LabeledGate {
		return Dataset{}, ErrCounts
	}
	if result.Legacy.Schema != "riido-behavior-probe-audit-v1" || result.Typed.Schema != "riido-typed-behavior-truth-v2" {
		return Dataset{}, ErrSchema
	}
	if e := validateCommon(result.Legacy.commonTruth, 48, 144, 36, 12, 12, 12); e != nil {
		return Dataset{}, e
	}
	if e := validateCommon(result.Typed.commonTruth, 24, 72, 15, 9, 5, 6); e != nil {
		return Dataset{}, e
	}
	if !validSHA(result.Legacy.SourceSHA) || result.Legacy.Connected != 11 || result.Legacy.Minimum != 15 || result.Legacy.Meets || len(result.Legacy.Groups) != 11 {
		return Dataset{}, ErrGroups
	}
	if e := validateSources(result); e != nil {
		return Dataset{}, e
	}
	var out Dataset
	out.Schema = Schema
	for _, cohort := range []struct {
		parents  []wireParent
		outcomes []wireOutcome
		name     string
		start    int
	}{{legacy.Parents, result.Legacy.Outcomes, "legacy", 0}, {typed.Parents, result.Typed.Outcomes, "typed", 48}} {
		if len(cohort.parents) != len(cohort.outcomes) {
			return Dataset{}, ErrIdentity
		}
		for i, p := range cohort.parents {
			if !validID(p.ID) || !validID(p.Prototype) || !validID(p.ContractID) || !validText(p.Request) || len(p.Candidates) < 2 || len(p.Candidates) > 4 {
				return Dataset{}, ErrIdentity
			}
			index := cohort.start + i
			for j := 0; j < index; j++ {
				if out.Rows[j].Audit.ParentID == p.ID {
					return Dataset{}, ErrIdentity
				}
			}
			r := Row{Text: Text{Request: p.Request, CandidateCount: len(p.Candidates)}, Audit: Audit{ParentID: p.ID, Prototype: p.Prototype, ContractID: p.ContractID, Cohort: cohort.name, CandidateCount: len(p.Candidates)}}
			for j, c := range p.Candidates {
				if !validID(c.ID) || !validID(c.SourceID) || !validText(c.Text) || !validSHA(c.CodeSHA256) {
					return Dataset{}, ErrText
				}
				for k := 0; k < j; k++ {
					if p.Candidates[k].ID == c.ID {
						return Dataset{}, ErrIdentity
					}
				}
				core, e := bindSource(c, p.Prototype, cohort.name, result)
				if e != nil {
					return Dataset{}, e
				}
				r.Text.Candidates[j] = c.Text
				r.Audit.Candidates[j] = CandidateAudit{c.ID, c.SourceID, c.CodeSHA256, c.BundleSHA256, core}
			}
			truth, e := bindOutcome(p, cohort.outcomes[i])
			if e != nil {
				return Dataset{}, e
			}
			r.Truth = truth
			switch truth.State {
			case "known":
				out.Answerable++
			case "no_answer":
				out.NoAnswer++
			case "unknown":
				out.Unknown++
			}
			out.Candidates += len(p.Candidates)
			out.CandidateCountHistogram[len(p.Candidates)]++
			out.Rows[index] = r
		}
	}
	if out.Answerable != 34 || out.NoAnswer != 17 || out.Unknown != 21 || out.Candidates != 216 || out.CandidateCountHistogram[2] != 12 || out.CandidateCountHistogram[3] != 48 || out.CandidateCountHistogram[4] != 12 {
		return Dataset{}, ErrCounts
	}
	if e := bindGroups(&out, result.Combined); e != nil {
		return Dataset{}, e
	}
	out.Metadata = Metadata{LegacyInputSHA256: LegacySHA256, TypedInputSHA256: TypedSHA256, StoredResultSHA256: ResultSHA256, GroupPolicy: result.Combined.Policy, Scope: result.Scope, StopReasons: append([]string(nil), result.StopReasons...), Edges: result.Combined.Edges, StoredEvidence: append(json.RawMessage(nil), resultRaw...)}
	return out, nil
}
func validateCommon(r commonTruth, parents, candidates, known, unknown, noanswer, prototypes int) error {
	if !validSHA(r.DatasetSHA) || r.Parents != parents || r.Candidates != candidates || r.Known != known || r.Unknown != unknown || r.NoAnswer != noanswer || r.Prototypes != prototypes || len(r.Outcomes) != parents || !r.ReviewRequired || r.Training || r.Partitioned || r.Final || len(r.Contracts) != prototypes {
		return ErrCounts
	}
	gotKnown, gotUnknown, gotNoAnswer := 0, 0, 0
	for _, o := range r.Outcomes {
		switch o.State {
		case "known":
			gotKnown++
		case "no_answer":
			gotKnown++
			gotNoAnswer++
		case "unknown":
			gotUnknown++
		default:
			return ErrTruth
		}
	}
	if gotKnown != known || gotUnknown != unknown || gotNoAnswer != noanswer {
		return ErrCounts
	}
	for _, c := range r.Contracts {
		if !validID(c.Prototype) || c.Core == "" || c.Semantics == "" || c.Vectors < 1 || !validSHA(c.TruthSHA) {
			return ErrSource
		}
	}
	return nil
}
func validateSources(r storedResult) error {
	for _, pin := range r.Implementation {
		if pin.Path == "" || !validSHA(pin.SHA256) {
			return ErrSource
		}
	}
	for _, pin := range r.Typed.Artifacts {
		if pin.Name == "" || !validSHA(pin.SHA256) {
			return ErrSource
		}
	}
	for _, s := range r.Legacy.Sources {
		if !validID(s.ID) || !validID(s.Prototype) || s.Core == "" || !validSHA(s.Code) || !validSHA(s.Normalized) {
			return ErrSource
		}
	}
	for _, s := range r.Typed.Sources {
		if !validID(s.ID) || !validID(s.Prototype) || s.Core == "" || s.Function == "" || !validSHA(s.Code) || !validSHA(s.Normalized) || !validSHA(s.Bundle) || len(s.Components) == 0 {
			return ErrSource
		}
		for _, c := range s.Components {
			if c.ID == "" || c.Kind == "" || !validSHA(c.SHA256) || c.Normalized != "" && !validSHA(c.Normalized) {
				return ErrSource
			}
		}
	}
	for _, controls := range [][]control{r.Legacy.Controls, r.Typed.Controls} {
		for _, c := range controls {
			if !validID(c.ID) || c.Checked < 1 || c.Failed < 0 || c.Failed > c.Checked || !validSHA(c.TruthSHA) || (c.Expected != "correct" && c.Expected != "wrong") || c.Expected == "correct" && c.Failed != 0 || c.Expected == "wrong" && c.Failed == 0 {
				return ErrSource
			}
		}
	}
	return nil
}
func bindSource(c wireCandidate, prototype, cohort string, r storedResult) (string, error) {
	matches := 0
	core := ""
	if cohort == "legacy" {
		if c.BundleSHA256 != "" {
			return "", ErrSource
		}
		for _, s := range r.Legacy.Sources {
			if s.ID == c.SourceID {
				if s.Prototype != prototype || s.Code != c.CodeSHA256 {
					return "", ErrSource
				}
				matches++
				core = s.Core
			}
		}
	} else {
		if !validSHA(c.BundleSHA256) {
			return "", ErrSource
		}
		for _, s := range r.Typed.Sources {
			if s.ID == c.SourceID {
				if s.Prototype != prototype || s.Code != c.CodeSHA256 || s.Bundle != c.BundleSHA256 {
					return "", ErrSource
				}
				matches++
				core = s.Core
			}
		}
	}
	if matches != 1 {
		return "", ErrSource
	}
	return core, nil
}
func bindOutcome(p wireParent, o wireOutcome) (Truth, error) {
	if o.ParentID != p.ID || len(o.Candidates) != len(p.Candidates) || len(o.Acceptable) > len(p.Candidates) {
		return Truth{}, ErrIdentity
	}
	out := Truth{State: o.State, Reason: o.Reason, AcceptableCount: len(o.Acceptable), CandidateCount: len(o.Candidates)}
	var acceptable [MaxCandidates]bool
	for i, index := range o.Acceptable {
		if index < 0 || index >= len(p.Candidates) || acceptable[index] || i > 0 && index <= o.Acceptable[i-1] {
			return Truth{}, ErrTruth
		}
		acceptable[index] = true
		out.AcceptableIndices[i] = index
	}
	if o.State == "known" && len(o.Acceptable) == 0 || (o.State == "no_answer" || o.State == "unknown") && len(o.Acceptable) != 0 {
		return Truth{}, ErrTruth
	}
	if o.State != "known" && o.State != "no_answer" && o.State != "unknown" {
		return Truth{}, ErrTruth
	}
	if o.State == "unknown" && o.Reason == "" || o.State != "unknown" && o.Reason != "" {
		return Truth{}, ErrTruth
	}
	for i, c := range o.Candidates {
		if c.CandidateID != p.Candidates[i].ID || c.VectorsChecked < 0 || c.FailedVectors < 0 || c.FailedVectors > c.VectorsChecked {
			return Truth{}, ErrTruth
		}
		if o.State == "unknown" {
			if c.State != "unknown" || c.VectorsChecked != 0 || c.FailedVectors != 0 {
				return Truth{}, ErrTruth
			}
		} else {
			if c.VectorsChecked == 0 || c.State != "acceptable" && c.State != "rejected" || acceptable[i] != (c.State == "acceptable") || c.State == "acceptable" && c.FailedVectors != 0 || c.State == "rejected" && c.FailedVectors == 0 {
				return Truth{}, ErrTruth
			}
		}
		out.Candidates[i] = c
	}
	return out, nil
}
func parentIndex(out *Dataset, id string) int {
	for i := range out.Rows {
		if out.Rows[i].Audit.ParentID == id {
			return i
		}
	}
	return -1
}
func stringsUnique(values []string) bool {
	for i, s := range values {
		if s == "" {
			return false
		}
		for j := 0; j < i; j++ {
			if values[j] == s {
				return false
			}
		}
	}
	return true
}
func contains(values []string, want string) bool {
	for _, s := range values {
		if s == want {
			return true
		}
	}
	return false
}
func bindGroups(out *Dataset, r combinedGroups) error {
	if r.Schema != "riido-combined-behavior-groups-v2" || r.Parents != 72 || r.Prototypes != 18 || r.Connected != 17 || r.Minimum != 15 || !r.Meets || r.Labeled != 16 || r.MinLabeled != 15 || !r.MeetsLabeled || r.Training || len(r.Groups) != 17 || r.Policy == "" {
		return ErrGroups
	}
	var seen [ParentCount]bool
	for i, g := range r.Groups {
		if g.ID < 0 || g.ID >= 72 || len(g.Parents) == 0 || !stringsUnique(g.Parents) || !stringsUnique(g.Prototypes) || !stringsUnique(g.CoreTemplates) || !stringsUnique(g.Sources) {
			return ErrGroups
		}
		for j := 0; j < i; j++ {
			if r.Groups[j].ID == g.ID {
				return ErrGroups
			}
		}
		labeled := false
		for _, id := range g.Parents {
			index := parentIndex(out, id)
			if index < 0 || seen[index] {
				return ErrGroups
			}
			seen[index] = true
			row := &out.Rows[index]
			row.Audit.GroupID = g.ID
			if row.Truth.State != "unknown" {
				labeled = true
			}
			if !contains(g.Prototypes, row.Audit.Prototype) {
				return ErrGroups
			}
			prefix := "v1:"
			if row.Audit.Cohort == "typed" {
				prefix = "v2:"
			}
			for j := 0; j < row.Audit.CandidateCount; j++ {
				c := row.Audit.Candidates[j]
				if !contains(g.Sources, prefix+c.SourceID) || !contains(g.CoreTemplates, c.CoreTemplate) {
					return ErrGroups
				}
			}
		}
		if labeled {
			out.LabeledGroups++
		}
		out.Groups[i] = g
	}
	for _, member := range seen {
		if !member {
			return ErrGroups
		}
	}
	if out.LabeledGroups != 16 {
		return ErrGroups
	}
	for i, e := range r.Edges {
		left, right := parentIndex(out, e.Left), parentIndex(out, e.Right)
		if left < 0 || right < 0 || left == right || out.Rows[left].Audit.GroupID != out.Rows[right].Audit.GroupID || len(e.Reasons) == 0 || !stringsUnique(e.Reasons) {
			return ErrGroups
		}
		for j := 0; j < i; j++ {
			previous := r.Edges[j]
			if previous.Left == e.Left && previous.Right == e.Right || previous.Left == e.Right && previous.Right == e.Left {
				return ErrGroups
			}
		}
	}
	return nil
}
