// Copyright 2026 teamswyg. SPDX-License-Identifier: Apache-2.0
package data35

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

func identifier(s string, max int) bool {
	if len(s) < 1 || len(s) > max {
		return false
	}
	for i, c := range []byte(s) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') && (i == 0 || c != '.' && c != '_' && c != '-') {
			return false
		}
	}
	return true
}
func hashOK(s string, n int) bool {
	if len(s) != n {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func yes(p *bool) bool { return p != nil && *p }
func no(p *bool) bool  { return p != nil && !*p }
func textOK(s string) bool {
	return utf8.ValidString(s) && len(s) > 0 && len(s) <= 512 && strings.TrimSpace(s) != "" && !strings.ContainsRune(s, 0) && len(strings.Fields(s)) <= 32
}

// Derived from the project's Apache-2.0 shortclaimdata reader. json.Valid must
// succeed first, making escape indexing safe. Reject replacement of surrogates.
func scalars(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != '"' {
			continue
		}
		for i++; i < len(b) && b[i] != '"'; i++ {
			if b[i] != '\\' {
				continue
			}
			i++
			if b[i] != 'u' {
				continue
			}
			u := hex4(b[i+1 : i+5])
			i += 4
			if u >= 0xdc00 && u <= 0xdfff {
				return false
			}
			if u >= 0xd800 && u <= 0xdbff {
				if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
					return false
				}
				lo := hex4(b[i+3 : i+7])
				if lo < 0xdc00 || lo > 0xdfff {
					return false
				}
				i += 6
			}
		}
	}
	return true
}
func hex4(b []byte) uint16 {
	var n uint16
	for _, c := range b {
		n <<= 4
		if c >= '0' && c <= '9' {
			n |= uint16(c - '0')
		} else if c >= 'a' && c <= 'f' {
			n |= uint16(c - 'a' + 10)
		} else if c >= 'A' && c <= 'F' {
			n |= uint16(c - 'A' + 10)
		}
	}
	return n
}
func walk(d *json.Decoder, depth int, nodes *int) error {
	*nodes++
	if depth > 32 || *nodes > 65536 {
		return code("json_shape")
	}
	t, e := d.Token()
	if e != nil {
		return code("json_decode")
	}
	v, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	switch v {
	case '{':
		var keys [128]string
		n := 0
		for d.More() {
			t, e = d.Token()
			key, ok := t.(string)
			if e != nil || !ok || n == len(keys) {
				return code("json_keys")
			}
			for _, old := range keys[:n] {
				if strings.EqualFold(old, key) {
					return code("json_duplicate_alias")
				}
			}
			keys[n] = key
			n++
			if e = walk(d, depth+1, nodes); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim('}') {
			return code("json_end")
		}
	case '[':
		n := 0
		for d.More() {
			n++
			if n > 1024 {
				return code("json_array_bound")
			}
			if e = walk(d, depth+1, nodes); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim(']') {
			return code("json_end")
		}
	default:
		return code("json_delimiter")
	}
	return nil
}
func decode(raw []byte, out any) error {
	if len(raw) < 1 || len(raw) > 1<<20 || !utf8.Valid(raw) || !json.Valid(raw) || !scalars(raw) {
		return code("json_unicode_bounds")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	nodes := 0
	if e := walk(d, 0, &nodes); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return code("json_trailing")
	}
	if json.Unmarshal(raw, out) != nil {
		return code("json_type")
	}
	return nil
}
func exactKeys(raw []byte, keys []string) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') {
		return code("row_object")
	}
	var seen [12]bool
	if len(keys) > len(seen) {
		return code("row_key_bound")
	}
	for d.More() {
		t, e = d.Token()
		key, ok := t.(string)
		if e != nil || !ok {
			return code("row_key")
		}
		at := -1
		for i, k := range keys {
			if key == k {
				at = i
				break
			}
		}
		if at < 0 || seen[at] {
			return code("row_key_case_extra_duplicate")
		}
		seen[at] = true
		var v json.RawMessage
		if d.Decode(&v) != nil {
			return code("row_value")
		}
	}
	t, e = d.Token()
	if e != nil || t != json.Delim('}') {
		return code("row_end")
	}
	for _, s := range seen[:len(keys)] {
		if !s {
			return code("row_key_missing")
		}
	}
	return nil
}
func readRow(raw []byte) (Row, error) {
	var r Row
	if len(raw) > MaxRow {
		return r, code("row_cap")
	}
	if e := decode(raw, &r); e != nil {
		return r, e
	}
	if e := exactKeys(raw, []string{"schema", "stable_id", "request", "candidates", "role", "whole_group", "source_family", "source_revision", "finite_scope", "text_revision", "feature_policy"}); e != nil {
		return r, e
	}
	var v struct {
		Candidates []json.RawMessage `json:"candidates"`
		Scope      json.RawMessage   `json:"finite_scope"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return r, code("row_type")
	}
	for _, c := range v.Candidates {
		if e := exactKeys(c, []string{"metadata_id", "text", "label", "sample_weight"}); e != nil {
			return r, e
		}
	}
	if e := exactKeys(v.Scope, []string{"input_count", "candidate_observations", "description", "all_input_guarantee", "unseen_source_guarantee"}); e != nil {
		return r, e
	}
	if r.Schema != "riido-finite-development-request-row-v1" || r.Role != "development_train" || r.Policy != "request_and_candidates_text_only" || !identifier(r.ID, 64) || !identifier(r.Family, 128) || !identifier(r.TextRevision, 128) || !hashOK(r.Revision, 40) || r.Group < 1 || r.Group > 65535 || !textOK(r.Request) || len(r.Candidates) < 1 || len(r.Candidates) > 8 || r.Scope.Inputs < 1 || r.Scope.Inputs > 4096 || r.Scope.Observations != r.Scope.Inputs*len(r.Candidates) || !no(r.Scope.All) || !no(r.Scope.Unseen) || len(r.Scope.Description) > 512 || strings.TrimSpace(r.Scope.Description) == "" {
		return r, code("row_scope_metadata")
	}
	for i, c := range r.Candidates {
		if !identifier(c.ID, 64) || !textOK(c.Text) || c.Label == nil || c.Weight == nil || *c.Weight != 1 {
			return r, code("row_supervision")
		}
		for _, old := range r.Candidates[:i] {
			if old.ID == c.ID {
				return r, code("row_candidate_duplicate")
			}
		}
	}
	return r, nil
}
func eqInts(a, b []int) bool {
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}
	for i, n := range a {
		if n != b[i] {
			return false
		}
	}
	return true
}
