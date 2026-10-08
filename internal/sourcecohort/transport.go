package sourcecohort

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"unicode/utf8"
)

// Transport checks opaque external schema/observation JSON without inventing
// schema interpretation. Content conformance remains a declared checker claim.
func Transport(data []byte) error {
	if len(data) == 0 || len(data) > 8<<20 || !utf8.Valid(data) || !unicodeEscapes(data) {
		return errCode("external_json")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if e := walk(d, 0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return errCode("external_json")
	}
	return nil
}
func walk(d *json.Decoder, depth int) error {
	if depth > 32 {
		return errCode("external_json")
	}
	t, e := d.Token()
	if e != nil {
		return errCode("external_json")
	}
	switch t {
	case json.Delim('{'):
		keys := []string{}
		for d.More() {
			k, e := d.Token()
			key, ok := k.(string)
			if e != nil || !ok || len(keys) >= 4096 {
				return errCode("external_json")
			}
			i := sort.SearchStrings(keys, key)
			if i < len(keys) && keys[i] == key {
				return errCode("external_json")
			}
			keys = append(keys, "")
			copy(keys[i+1:], keys[i:])
			keys[i] = key
			if e := walk(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim('}') {
			return errCode("external_json")
		}
	case json.Delim('['):
		n := 0
		for d.More() {
			n++
			if n > 4096 {
				return errCode("external_json")
			}
			if e := walk(d, depth+1); e != nil {
				return e
			}
		}
		t, e = d.Token()
		if e != nil || t != json.Delim(']') {
			return errCode("external_json")
		}
	case json.Delim('}'), json.Delim(']'):
		return errCode("external_json")
	}
	return nil
}
