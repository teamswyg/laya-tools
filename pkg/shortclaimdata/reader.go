// Package shortclaimdata reads finite development supervision. It grants no
// semantic truth, rights, qualification, evaluation role or training authority.
package shortclaimdata

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	Schema           = "riido-finite-development-request-row-v1"
	Role             = "development_train"
	FeaturePolicy    = "request_and_candidates_text_only"
	InputProvenance  = "finite-development-v1"
	MaxRowBytes      = 16 << 10
	MaxGroup         = 65535
	MaxFiniteInputs  = 4096
	MaxMetadataBytes = 128
)

// Error is a fixed diagnostic: no decoder, reader, supplied text or path is
// retained by an error. Failed loads return a completely zero Example.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrRead       Error = "shortclaimdata_read_failed"
	ErrBounds     Error = "shortclaimdata_row_bounds"
	ErrJSON       Error = "shortclaimdata_invalid_json"
	ErrUnicode    Error = "shortclaimdata_invalid_unicode"
	ErrField      Error = "shortclaimdata_unknown_field"
	ErrDuplicate  Error = "shortclaimdata_duplicate_field"
	ErrMissing    Error = "shortclaimdata_missing_field"
	ErrSchema     Error = "shortclaimdata_unsupported_schema"
	ErrRole       Error = "shortclaimdata_invalid_role"
	ErrCandidates Error = "shortclaimdata_candidate_count"
	ErrLabel      Error = "shortclaimdata_invalid_label"
	ErrWeight     Error = "shortclaimdata_invalid_weight"
	ErrMetadata   Error = "shortclaimdata_invalid_metadata"
	ErrScope      Error = "shortclaimdata_invalid_finite_scope"
	ErrInput      Error = "shortclaimdata_invalid_input"
)

// FiniteScope is bookkeeping, never encoder input. A count is a declaration
// checked for internal shape, not proof that observations happened or are true.
type FiniteScope struct {
	InputCount, CandidateObservations        int
	Description                              string
	AllInputGuarantee, UnseenSourceGuarantee bool
}

type Metadata struct {
	StableID, SourceFamily, SourceRevision, TextRevision string
	Role, FeaturePolicy                                  string
	WholeGroup                                           int
	FiniteScope                                          FiniteScope
}

// Supervision owns SoA arrays in candidate order. Unused slots are zero. A
// returned value can be edited without changing its Example or validated input.
type Supervision struct {
	Labels  [shortclaim.MaxCandidates]bool
	Weights [shortclaim.MaxCandidates]int
	Count   int
}

// Example owns immutable strings and fixed arrays, with no mutable slices or
// maps retained. Its zero value is invalid. Metadata and supervision cannot be
// implicitly passed to the shortclaim feature path by this API.
type Example struct {
	input       shortclaim.ValidatedInput
	metadata    Metadata
	supervision Supervision
	ready       bool
}

func (e Example) Valid() bool                      { return e.ready }
func (e Example) Input() shortclaim.ValidatedInput { return e.input }
func (e Example) Metadata() Metadata               { return e.metadata }
func (e Example) Supervision() Supervision         { return e.supervision }

type wireRow struct {
	schema, request string
	metadata        Metadata
	candidates      [shortclaim.MaxCandidates]shortclaim.Candidate
	supervision     Supervision
}

// LoadDevelopmentRow accepts exactly one bounded JSON object, including optional
// surrounding whitespace. A JSONL caller supplies one already bounded row, not
// the entire file. This function never runs inference, fitting or scoring.
func LoadDevelopmentRow(r io.Reader) (Example, error) {
	if r == nil {
		return Example{}, ErrRead
	}
	raw, err := io.ReadAll(io.LimitReader(r, MaxRowBytes+1))
	if err != nil {
		return Example{}, ErrRead
	}
	if len(raw) > MaxRowBytes {
		return Example{}, ErrBounds
	}
	if !utf8.Valid(raw) {
		return Example{}, ErrUnicode
	}
	if !json.Valid(raw) {
		return Example{}, ErrJSON
	}
	if !scalarEscapes(raw) {
		return Example{}, ErrUnicode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var w wireRow
	err = object(d, [12]string{"schema", "stable_id", "request", "candidates", "role", "whole_group", "source_family", "source_revision", "finite_scope", "text_revision", "feature_policy"}, 11, func(key string) error {
		switch key {
		case "schema":
			return text(d, &w.schema)
		case "stable_id":
			return text(d, &w.metadata.StableID)
		case "request":
			return text(d, &w.request)
		case "candidates":
			return candidates(d, &w)
		case "role":
			return text(d, &w.metadata.Role)
		case "whole_group":
			return positiveInt(d, MaxGroup, &w.metadata.WholeGroup, ErrMetadata)
		case "source_family":
			return text(d, &w.metadata.SourceFamily)
		case "source_revision":
			return text(d, &w.metadata.SourceRevision)
		case "finite_scope":
			return scope(d, &w.metadata.FiniteScope)
		case "text_revision":
			return text(d, &w.metadata.TextRevision)
		case "feature_policy":
			return text(d, &w.metadata.FeaturePolicy)
		}
		return ErrField
	})
	if err != nil {
		return Example{}, err
	}
	if _, err = d.Token(); err != io.EOF {
		return Example{}, ErrJSON
	}
	if w.schema != Schema {
		return Example{}, ErrSchema
	}
	if w.metadata.Role != Role {
		return Example{}, ErrRole
	}
	if w.metadata.FeaturePolicy != FeaturePolicy {
		return Example{}, ErrMetadata
	}
	if !identifier(w.metadata.StableID, shortclaim.MaxIDBytes) ||
		!identifier(w.metadata.SourceFamily, MaxMetadataBytes) ||
		!identifier(w.metadata.TextRevision, MaxMetadataBytes) ||
		!revision(w.metadata.SourceRevision) {
		return Example{}, ErrMetadata
	}
	n := w.supervision.Count
	if w.metadata.FiniteScope.CandidateObservations != w.metadata.FiniteScope.InputCount*n {
		return Example{}, ErrScope
	}
	// ValidateInput owns a fixed array. Only request and candidate text are
	// normalized. Metadata IDs remain inspection metadata; generic provenance
	// deliberately differs from the original public source Input provenance.
	in, err := shortclaim.ValidateInput(shortclaim.Input{
		Schema: shortclaim.Schema, Request: w.request,
		Candidates: w.candidates[:n], Provenance: InputProvenance,
	})
	if err != nil {
		return Example{}, ErrInput
	}
	return Example{input: in, metadata: w.metadata, supervision: w.supervision, ready: true}, nil
}

func candidates(d *json.Decoder, w *wireRow) error {
	t, err := d.Token()
	if err != nil || t != json.Delim('[') {
		return ErrJSON
	}
	n := 0
	for d.More() {
		if n == shortclaim.MaxCandidates {
			return ErrCandidates
		}
		err = object(d, [12]string{"metadata_id", "text", "label", "sample_weight"}, 4, func(key string) error {
			switch key {
			case "metadata_id":
				return text(d, &w.candidates[n].ID)
			case "text":
				return text(d, &w.candidates[n].Text)
			case "label":
				return boolean(d, &w.supervision.Labels[n], ErrLabel)
			case "sample_weight":
				t, e := d.Token()
				number, ok := t.(json.Number)
				if e != nil || !ok || number.String() != "1" {
					return ErrWeight
				}
				w.supervision.Weights[n] = 1
			}
			return nil
		})
		if err != nil {
			return err
		}
		n++
	}
	if t, err = d.Token(); err != nil || t != json.Delim(']') {
		return ErrJSON
	}
	if n == 0 {
		return ErrCandidates
	}
	w.supervision.Count = n
	return nil
}

func scope(d *json.Decoder, s *FiniteScope) error {
	err := object(d, [12]string{"input_count", "candidate_observations", "description", "all_input_guarantee", "unseen_source_guarantee"}, 5, func(key string) error {
		switch key {
		case "input_count":
			return positiveInt(d, MaxFiniteInputs, &s.InputCount, ErrScope)
		case "candidate_observations":
			return positiveInt(d, MaxFiniteInputs*shortclaim.MaxCandidates, &s.CandidateObservations, ErrScope)
		case "description":
			return text(d, &s.Description)
		case "all_input_guarantee":
			return boolean(d, &s.AllInputGuarantee, ErrScope)
		case "unseen_source_guarantee":
			return boolean(d, &s.UnseenSourceGuarantee, ErrScope)
		}
		return ErrField
	})
	if err != nil {
		return err
	}
	if s.AllInputGuarantee || s.UnseenSourceGuarantee || len(s.Description) > shortclaim.MaxTextBytes ||
		strings.TrimSpace(s.Description) == "" || strings.ContainsRune(s.Description, 0) {
		return ErrScope
	}
	return nil
}

func text(d *json.Decoder, dst *string) error {
	t, err := d.Token()
	s, ok := t.(string)
	if err != nil || !ok {
		return ErrJSON
	}
	*dst = s
	return nil
}
func boolean(d *json.Decoder, dst *bool, failure Error) error {
	t, err := d.Token()
	b, ok := t.(bool)
	if err != nil || !ok {
		return failure
	}
	*dst = b
	return nil
}
func positiveInt(d *json.Decoder, max int, dst *int, failure Error) error {
	t, err := d.Token()
	number, ok := t.(json.Number)
	if err != nil || !ok {
		return failure
	}
	s := number.String()
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return failure
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > max {
		return failure
	}
	*dst = n
	return nil
}

// Object fields use fixed arrays: exact decoded names, no struct case folding.
// Both escaped duplicates and case aliases are rejected before their values.
// Every whitelisted field is mandatory, including label=false and weight=1.
func object(d *json.Decoder, allowed [12]string, count int, value func(string) error) error {
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return ErrJSON
	}
	var seen [12]string
	n := 0
	for d.More() {
		t, err = d.Token()
		key, ok := t.(string)
		if err != nil || !ok {
			return ErrJSON
		}
		for j := 0; j < n; j++ {
			if strings.EqualFold(seen[j], key) {
				return ErrDuplicate
			}
		}
		found := false
		for j := 0; j < count; j++ {
			if allowed[j] == key {
				found = true
				break
			}
		}
		if !found {
			return ErrField
		}
		seen[n] = key
		n++
		if err = value(key); err != nil {
			return err
		}
	}
	if t, err = d.Token(); err != nil || t != json.Delim('}') {
		return ErrJSON
	}
	if n != count {
		return ErrMissing
	}
	return nil
}

func identifier(s string, bound int) bool {
	if len(s) < 1 || len(s) > bound {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alnum && (i == 0 || c != '.' && c != '_' && c != '-') {
			return false
		}
	}
	return true
}
func revision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !(s[i] >= '0' && s[i] <= '9' || s[i] >= 'a' && s[i] <= 'f') {
			return false
		}
	}
	return true
}

// Derived from this project's shortclaim wire validator (Apache-2.0). Valid
// JSON is checked first, so slices below are safe. A UTF-16 surrogate is allowed
// only as a high+low pair; literal U+FFFD and escaped backslashes remain valid.
func scalarEscapes(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '"' {
			continue
		}
		for i++; i < len(raw) && raw[i] != '"'; i++ {
			if raw[i] != '\\' {
				continue
			}
			i++
			if raw[i] != 'u' {
				continue
			}
			u := hex4(raw[i+1 : i+5])
			i += 4
			if u >= 0xdc00 && u <= 0xdfff {
				return false
			}
			if u >= 0xd800 && u <= 0xdbff {
				if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
					return false
				}
				low := hex4(raw[i+3 : i+7])
				if low < 0xdc00 || low > 0xdfff {
					return false
				}
				i += 6
			}
		}
	}
	return true
}
func hex4(b []byte) uint16 {
	var v uint16
	for _, c := range b {
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= uint16(c - '0')
		case c >= 'a' && c <= 'f':
			v |= uint16(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			v |= uint16(c - 'A' + 10)
		}
	}
	return v
}
