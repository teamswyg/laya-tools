package semanticframe

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

// Preflight enforces depth and collection ceilings before a typed decoder can
// allocate whole lists. The raw document ceiling bounds token string work.
func preflight(raw []byte) error {
	if len(raw) == 0 || len(raw) > MaxInventoryBytes || !utf8.Valid(raw) {
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
