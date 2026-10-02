// Package shortclaim prepares bounded, unverified behavior claims. It does not
// certify task completion or authorize an action.
package shortclaim

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

const (
	Schema             = "riido-short-behavior-claim-v1"
	MaxJSONBytes       = 12 << 10
	MaxCandidates      = 8
	MaxTextBytes       = 512
	MaxNormalizedWords = 32
	MaxIDBytes         = 64
	MaxProvenanceBytes = 128
)

// Error is a fixed public diagnostic. It never includes supplied text or a
// decoder/reader error, which could retain raw input or private paths.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrRead             Error = "shortclaim_read_failed"
	ErrJSONBounds       Error = "shortclaim_json_bounds"
	ErrJSON             Error = "shortclaim_invalid_json"
	ErrUnicode          Error = "shortclaim_invalid_unicode"
	ErrUnknownField     Error = "shortclaim_unknown_field"
	ErrDuplicateField   Error = "shortclaim_duplicate_field"
	ErrSchema           Error = "shortclaim_unsupported_schema"
	ErrCandidateCount   Error = "shortclaim_candidate_count"
	ErrIdentifier       Error = "shortclaim_invalid_identifier"
	ErrDuplicateID      Error = "shortclaim_duplicate_candidate_id"
	ErrTextBounds       Error = "shortclaim_text_bounds"
	ErrNormalizedBounds Error = "shortclaim_normalized_bounds"
	ErrEmptyText        Error = "shortclaim_empty_text"
	ErrPrepared         Error = "shortclaim_invalid_prepared"
)

type Candidate struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Input struct {
	Schema     string      `json:"schema"`
	Request    string      `json:"request"`
	Candidates []Candidate `json:"candidates"`
	Provenance string      `json:"provenance"`
}

// PreparedCandidate retains raw text as an immutable Go string. NormalizedText
// is for features; ID is metadata and must not enter features.
type PreparedCandidate struct {
	ID, Text, NormalizedText string
}

// Prepared owns a fixed candidate array, preserving input order. Copies share
// only immutable string bytes, not mutable slices. Because exported fields can
// be forged or modified, public consumers must first call ValidatePrepared.
type Prepared struct {
	Schema, Request, NormalizedRequest, Provenance string
	Candidates                                     [MaxCandidates]PreparedCandidate
	Count                                          int
}

// Load accepts one JSON object with exact, case-sensitive fields. Duplicate
// decoded keys (including escaped spellings) and unknown fields are rejected.
// JSON wire size is independently bounded; Validate has no JSON wire to bound.
// Malformed input returns a zero Prepared and must not become runtime fallback.
func Load(r io.Reader) (Prepared, error) {
	if r == nil {
		return Prepared{}, ErrRead
	}
	raw, err := io.ReadAll(io.LimitReader(r, MaxJSONBytes+1))
	if err != nil {
		return Prepared{}, ErrRead
	}
	if len(raw) > MaxJSONBytes {
		return Prepared{}, ErrJSONBounds
	}
	if !utf8.Valid(raw) {
		return Prepared{}, ErrUnicode
	}
	if !json.Valid(raw) {
		return Prepared{}, ErrJSON
	}
	// encoding/json replaces unpaired UTF-16 escapes with U+FFFD. Reject these
	// throughout the JSON before decoding, including an unknown field's value.
	if !scalarEscapes(raw) {
		return Prepared{}, ErrUnicode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	var in Input
	var candidates [MaxCandidates]Candidate
	err = object(d, [4]string{"schema", "request", "candidates", "provenance"}, 4, func(key string) error {
		switch key {
		case "schema":
			return textValue(d, &in.Schema)
		case "request":
			return textValue(d, &in.Request)
		case "provenance":
			return textValue(d, &in.Provenance)
		case "candidates":
			t, e := d.Token()
			if e != nil || t != json.Delim('[') {
				return ErrJSON
			}
			n := 0
			for d.More() {
				if n == MaxCandidates {
					return ErrCandidateCount
				}
				e = object(d, [4]string{"id", "text"}, 2, func(key string) error {
					if key == "id" {
						return textValue(d, &candidates[n].ID)
					}
					return textValue(d, &candidates[n].Text)
				})
				if e != nil {
					return e
				}
				n++
			}
			if t, e = d.Token(); e != nil || t != json.Delim(']') {
				return ErrJSON
			}
			in.Candidates = candidates[:n]
		}
		return nil
	})
	if err != nil {
		return Prepared{}, err
	}
	if _, err = d.Token(); err != io.EOF {
		return Prepared{}, ErrJSON
	}
	return Validate(in)
}

// Validate applies the same decoded-field constraints as Load. It does not
// mutate the caller's slice or strings. There is no fallback or truncation here.
func Validate(in Input) (Prepared, error) {
	if len(in.Candidates) < 1 || len(in.Candidates) > MaxCandidates {
		return Prepared{}, ErrCandidateCount
	}
	if !utf8.ValidString(in.Schema) || !utf8.ValidString(in.Request) || !utf8.ValidString(in.Provenance) {
		return Prepared{}, ErrUnicode
	}
	for _, c := range in.Candidates {
		if !utf8.ValidString(c.ID) || !utf8.ValidString(c.Text) {
			return Prepared{}, ErrUnicode
		}
	}
	if in.Schema != Schema {
		return Prepared{}, ErrSchema
	}
	if !identifier(in.Provenance, MaxProvenanceBytes) {
		return Prepared{}, ErrIdentifier
	}
	normalized, err := prepareText(in.Request)
	if err != nil {
		return Prepared{}, err
	}
	p := Prepared{Schema: in.Schema, Request: in.Request, NormalizedRequest: normalized, Provenance: in.Provenance, Count: len(in.Candidates)}
	for i, c := range in.Candidates {
		if !identifier(c.ID, MaxIDBytes) {
			return Prepared{}, ErrIdentifier
		}
		for j := 0; j < i; j++ {
			if c.ID == in.Candidates[j].ID {
				return Prepared{}, ErrDuplicateID
			}
		}
		normalized, err = prepareText(c.Text)
		if err != nil {
			return Prepared{}, err
		}
		p.Candidates[i] = PreparedCandidate{ID: c.ID, Text: c.Text, NormalizedText: normalized}
	}
	return p, nil
}

// ValidatePrepared rejects forged normalization, count, metadata, or unused
// candidate slots. Each consumer gets its own scratch; there is no shared lock.
func ValidatePrepared(p Prepared) error {
	if p.Count < 1 || p.Count > MaxCandidates {
		return ErrCandidateCount
	}
	var candidates [MaxCandidates]Candidate
	for i := 0; i < p.Count; i++ {
		candidates[i] = Candidate{ID: p.Candidates[i].ID, Text: p.Candidates[i].Text}
	}
	want, err := Validate(Input{Schema: p.Schema, Request: p.Request, Candidates: candidates[:p.Count], Provenance: p.Provenance})
	if err != nil {
		return err
	}
	if p != want {
		return ErrPrepared
	}
	return nil
}

func prepareText(raw string) (string, error) {
	if len(raw) > MaxTextBytes {
		return "", ErrTextBounds
	}
	n := lexicalhint.NormalizeText(raw)
	if n == "" {
		return "", ErrEmptyText
	}
	if len(n) > MaxTextBytes || 1+strings.Count(n, " ") > MaxNormalizedWords {
		return "", ErrNormalizedBounds
	}
	return n, nil
}

// Public identifiers are flat ASCII metadata, not paths or arbitrary payloads.
// Candidate IDs are case-sensitive; only exact duplicates are disallowed.
func identifier(s string, bound int) bool {
	if len(s) < 1 || len(s) > bound {
		return false
	}
	for i := range len(s) {
		c := s[i]
		alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alnum && (i == 0 || c != '.' && c != '_' && c != '-') {
			return false
		}
	}
	return true
}

func textValue(d *json.Decoder, dst *string) error {
	t, err := d.Token()
	s, ok := t.(string)
	if err != nil || !ok {
		return ErrJSON
	}
	*dst = s
	return nil
}

// Each object has at most four exact fields. A fixed array checks decoded key
// duplicates before dispatch; no case-insensitive struct decoding is used.
func object(d *json.Decoder, allowed [4]string, count int, value func(string) error) error {
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return ErrJSON
	}
	var seen [4]string
	n := 0
	for d.More() {
		t, err = d.Token()
		key, ok := t.(string)
		if err != nil || !ok {
			return ErrJSON
		}
		for j := 0; j < n; j++ {
			if seen[j] == key {
				return ErrDuplicateField
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
			return ErrUnknownField
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
	return nil
}

// scalarEscapes scans valid JSON without recursion or allocating decoded
// strings. Only escaped high+low surrogate pairs are Unicode scalars; literal
// U+FFFD and escaped backslashes remain valid. json.Valid runs first.
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
