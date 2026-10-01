package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func hashBytes(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func validSHA(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// Separate identities: no source, target, ID or provenance enters featureSHA.
// A tagged count and uint64 big-endian byte length precede ordered texts.
func featureSHA(p shortclaim.Prepared) string {
	h := sha256.New()
	writeLengthText(h, "riido-resident-features-v1")
	var count [8]byte
	binary.BigEndian.PutUint64(count[:], uint64(p.Count))
	_, _ = h.Write(count[:])
	writeLengthText(h, p.NormalizedRequest)
	for i := 0; i < p.Count; i++ {
		writeLengthText(h, p.Candidates[i].NormalizedText)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writeLengthText(w io.Writer, text string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(text)))
	_, _ = w.Write(size[:])
	_, _ = io.WriteString(w, text)
}

// Exactly the pinned child's digest contract, which includes raw metadata.
// Integration must test independently literal digest vectors before freeze.
func inputDigest(p shortclaim.Prepared) string {
	h := sha256.New()
	for _, text := range [3]string{p.Schema, p.Request, p.Provenance} {
		writeLengthText(h, text)
	}
	for i := 0; i < p.Count; i++ {
		writeLengthText(h, p.Candidates[i].ID)
		writeLengthText(h, p.Candidates[i].Text)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Reject encoding/json's case folding and duplicate decoded names before
// strict struct decoding. Token scanning is depth-bounded and array-backed.
func scanValue(d *json.Decoder, depth int) error {
	if depth > 16 {
		return errors.New("json_depth_limit")
	}
	t, e := d.Token()
	if e != nil {
		return errors.New("json_invalid")
	}
	delim, composite := t.(json.Delim)
	if !composite {
		return nil
	}
	switch delim {
	case '{':
		var keys [64]string
		n := 0
		for d.More() {
			t, e = d.Token()
			key, ok := t.(string)
			if e != nil || !ok || len(key) == 0 || n == len(keys) {
				return errors.New("json_keys_invalid")
			}
			for _, c := range key {
				if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
					return errors.New("json_keys_invalid")
				}
			}
			for i := 0; i < n; i++ {
				if keys[i] == key {
					return errors.New("json_keys_invalid")
				}
			}
			keys[n], n = key, n+1
			if e := scanValue(d, depth+1); e != nil {
				return e
			}
		}
		if end, e := d.Token(); e != nil || end != json.Delim('}') {
			return errors.New("json_invalid")
		}
	case '[':
		for d.More() {
			if e := scanValue(d, depth+1); e != nil {
				return e
			}
		}
		if end, e := d.Token(); e != nil || end != json.Delim(']') {
			return errors.New("json_invalid")
		}
	default:
		return errors.New("json_invalid")
	}
	return nil
}

func strictJSON(raw []byte, target any) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' || !utf8.Valid(raw) || !json.Valid(raw) || !scalarJSONEscapes(raw) {
		return errors.New("json_invalid")
	}
	check := json.NewDecoder(bytes.NewReader(raw))
	check.UseNumber()
	if e := scanValue(check, 0); e != nil {
		return e
	}
	if _, e := check.Token(); e != io.EOF {
		return errors.New("json_multiple_values")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil || d.Decode(new(any)) != io.EOF {
		return errors.New("json_shape_invalid")
	}
	return nil
}

// json.Valid has already checked escape lengths and hexadecimal digits.
// Reject unpaired UTF-16 escapes before encoding/json can replace them with
// U+FFFD. Escaped backslashes and an explicit U+FFFD remain valid scalars.
func scalarJSONEscapes(raw []byte) bool {
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
			u := jsonHex16(raw[i+1 : i+5])
			i += 4
			if u >= 0xdc00 && u <= 0xdfff {
				return false
			}
			if u >= 0xd800 && u <= 0xdbff {
				if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
					return false
				}
				low := jsonHex16(raw[i+3 : i+7])
				if low < 0xdc00 || low > 0xdfff {
					return false
				}
				i += 6
			}
		}
	}
	return true
}

func jsonHex16(b []byte) uint16 {
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

func expectedResponse(p shortclaim.Prepared, kind string) ([]byte, ExpectedResponse, error) {
	r, e := shortclaim.Rank(p, kind)
	if e != nil {
		return nil, ExpectedResponse{}, errors.New("expectation_rank_failed")
	}
	out := WireOutput{Schema: responseSchema, Status: "unverified_heuristic", Baseline: kind, InputSHA256: inputDigest(p), FallbackReason: r.FallbackReason, Candidates: make([]ScoredCandidate, p.Count)}
	for i := 0; i < r.Count; i++ {
		index := r.Order[i]
		out.Candidates[i] = ScoredCandidate{ID: p.Candidates[index].ID, Score: r.Scores[index]}
	}
	var encoded bytes.Buffer
	if json.NewEncoder(&encoded).Encode(out) != nil {
		return nil, ExpectedResponse{}, errors.New("expectation_encoding_failed")
	}
	raw := encoded.Bytes()
	if len(raw) > wireLimit {
		return nil, ExpectedResponse{}, errors.New("expectation_size_invalid")
	}
	return slices.Clone(raw), ExpectedResponse{Baseline: kind, RawSHA256: hashBytes(raw), RawBytes: len(raw), FallbackReason: r.FallbackReason}, nil
}

func checkResponse(raw []byte, payload Payload, kind string, expected ExpectedResponse) error {
	if len(raw) < 2 || len(raw) > wireLimit || raw[len(raw)-1] != '\n' || bytes.IndexAny(raw[:len(raw)-1], "\r\n") >= 0 {
		return errors.New("response_framing_invalid")
	}
	kinds := baselineKinds()
	if !slices.Contains(kinds[:], kind) || expected.Baseline != kind || !validSHA(payload.InputDigest) || !validSHA(expected.RawSHA256) || expected.RawBytes < 2 || expected.RawBytes > wireLimit {
		return errors.New("response_expectation_invalid")
	}
	// Pointer fields distinguish a missing/null score from the legitimate
	// zero score. The byte pin alone must not stand in for this wire contract.
	var r struct {
		Schema         *string         `json:"schema"`
		Status         *string         `json:"status"`
		Baseline       *string         `json:"baseline"`
		InputSHA256    *string         `json:"input_sha256"`
		FallbackReason json.RawMessage `json:"fallback_reason"`
		Candidates     *[]struct {
			ID    *string  `json:"id"`
			Score *float64 `json:"score"`
		} `json:"verification_order"`
	}
	if strictJSON(raw, &r) != nil {
		return errors.New("response_json_invalid")
	}
	if r.Schema == nil || r.Status == nil || r.Baseline == nil || r.InputSHA256 == nil || r.Candidates == nil || *r.Schema != responseSchema || *r.Status != "unverified_heuristic" || *r.Baseline != kind || *r.InputSHA256 != payload.InputDigest || len(*r.Candidates) != 3 {
		return errors.New("response_contract_invalid")
	}
	fallback := ""
	if len(r.FallbackReason) != 0 {
		if r.FallbackReason[0] != '"' || json.Unmarshal(r.FallbackReason, &fallback) != nil {
			return errors.New("response_fallback_invalid")
		}
	}
	if fallback != expected.FallbackReason || kind != "narrow_rule" && fallback != "" {
		return errors.New("response_fallback_invalid")
	}
	if fallback != "" && fallback != "unsupported_rule_request" && fallback != "unsupported_rule_candidate" {
		return errors.New("response_fallback_invalid")
	}
	ids := candidateIDs()
	var seen [3]bool
	for _, c := range *r.Candidates {
		if c.ID == nil || c.Score == nil {
			return errors.New("response_candidates_invalid")
		}
		index := slices.Index(ids[:], *c.ID)
		if index < 0 || seen[index] || math.IsNaN(*c.Score) || math.IsInf(*c.Score, 0) {
			return errors.New("response_candidates_invalid")
		}
		seen[index] = true
	}
	if hashBytes(raw) != expected.RawSHA256 || len(raw) != expected.RawBytes {
		return errors.New("response_bytes_mismatch")
	}
	return nil
}

func words(s string) int {
	if s == "" {
		return 0
	}
	return 1 + strings.Count(s, " ")
}

func distribution(samples []int64) Summary {
	if len(samples) == 0 {
		return Summary{}
	}
	ordered := slices.Clone(samples)
	slices.Sort(ordered)
	n := len(ordered)
	return Summary{Count: n, P50NS: ordered[(n-1)/2], P95NS: ordered[(95*n+99)/100-1], MaxNS: ordered[n-1]}
}
