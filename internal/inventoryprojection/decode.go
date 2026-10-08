package inventoryprojection

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// Preflight enforces depth and collection ceilings before a typed decoder can
// allocate whole lists. The raw document ceiling bounds token string work.
func preflight(raw []byte) error {
	if len(raw) == 0 || len(raw) > MaxBytes || !utf8.Valid(raw) {
		return Error("json_bounds")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 32 {
			return Error("json_depth")
		}
		t, err := d.Token()
		if err != nil {
			return Error("json_shape")
		}
		if s, ok := t.(string); ok && len(s) > 65536 {
			return Error("string_bounds")
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		if delim != '{' && delim != '[' {
			return Error("json_shape")
		}
		count := 0
		for d.More() {
			if count >= MaxItems {
				return Error("collection_bounds")
			}
			count++
			if delim == '{' {
				key, err := d.Token()
				if err != nil {
					return Error("json_shape")
				}
				s, ok := key.(string)
				if !ok || len(s) > 128 {
					return Error("key_bounds")
				}
			}
			if err := walk(depth + 1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || delim == '{' && end != json.Delim('}') || delim == '[' && end != json.Delim(']') {
			return Error("json_shape")
		}
		return nil
	}
	if err := walk(0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return Error("json_trailing")
	}
	return nil
}
func closed(raw []byte, out any) error {
	if err := preflight(raw); err != nil {
		return err
	}
	if sourcecohort.Decode(raw, out) != nil {
		return Error("closed_json")
	}
	return nil
}

func decodeFull(raw []byte) (Full, error) {
	var f Full
	if err := preflight(raw); err != nil {
		return f, err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return f, Error("full_shape")
	}
	v := reflect.ValueOf(&f).Elem()
	typ := v.Type()
	seen := make([]bool, v.NumField())
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return Full{}, Error("full_shape")
		}
		key, ok := t.(string)
		if !ok {
			return Full{}, Error("full_shape")
		}
		index := -1
		for i := range seen {
			if typ.Field(i).Tag.Get("json") == key {
				index = i
				break
			}
		}
		if index < 0 || seen[index] {
			return Full{}, Error("full_key")
		}
		seen[index] = true
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return Full{}, Error("full_shape")
		}
		if key == "official_field_support" {
			fields, err := decodeFields(value)
			if err != nil {
				return Full{}, err
			}
			f.OfficialFields = fields
		} else if sourcecohort.Decode(value, v.Field(index).Addr().Interface()) != nil {
			return Full{}, Error("full_value")
		}
	}
	for _, ok := range seen {
		if !ok {
			return Full{}, Error("full_missing")
		}
	}
	if t, err := d.Token(); err != nil || t != json.Delim('}') {
		return Full{}, Error("full_shape")
	}
	if _, err := d.Token(); err != io.EOF {
		return Full{}, Error("full_trailing")
	}
	return f, nil
}
func decodeFields(raw []byte) ([]OfficialField, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return nil, Error("support_shape")
	}
	out := make([]OfficialField, 0)
	for d.More() {
		if len(out) >= MaxItems {
			return nil, Error("support_count")
		}
		t, err := d.Token()
		if err != nil {
			return nil, Error("support_shape")
		}
		name, ok := t.(string)
		if !ok || !identifier(name) {
			return nil, Error("support_key")
		}
		for _, f := range out {
			if f.Name == name || strings.EqualFold(f.Name, name) {
				return nil, Error("support_duplicate_key")
			}
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return nil, Error("support_shape")
		}
		var s Support
		if sourcecohort.Decode(value, &s) != nil {
			return nil, Error("support_value")
		}
		out = append(out, OfficialField{Name: name, Raw: append([]byte{}, value...), Support: s})
	}
	if t, err := d.Token(); err != nil || t != json.Delim('}') {
		return nil, Error("support_shape")
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, Error("support_shape")
	}
	return out, nil
}
