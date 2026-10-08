package nativecontract

import (
	"bytes"
	"encoding/json"
	"io"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func identifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}
func actorLabel(s string) string {
	if len(s) > 128 {
		return ""
	}
	label := strings.TrimPrefix(s, "/root/")
	if !identifier(label) {
		return ""
	}
	return label
}
func validFile(f File) bool {
	if len(f.Path) == 0 || len(f.Path) > 4096 || !utf8.ValidString(f.Path) || strings.ContainsAny(f.Path, "\\:\x00") || path.IsAbs(f.Path) || path.Clean(f.Path) != f.Path || f.Path == "." || f.Path == ".." || strings.HasPrefix(f.Path, "../") || len(f.SHA256) != 64 || f.Bytes < 1 || f.Bytes > MaxBytes {
		return false
	}
	for _, c := range f.SHA256 {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func verify(p PinnedBytes) error {
	if len(p.Bytes) < 1 || len(p.Bytes) > MaxBytes || !validFile(p.File) || p.File.Bytes != int64(len(p.Bytes)) || sourcecohort.Hash(p.Bytes) != p.File.SHA256 {
		return Error("byte_pin")
	}
	return nil
}
func checkPins(pins []File) error {
	for _, f := range pins {
		if !validFile(f) {
			return Error("reference_pin")
		}
	}
	sorted := append([]File{}, pins...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Path == sorted[i-1].Path && sorted[i] != sorted[i-1] {
			return Error("conflicting_reference_pin")
		}
	}
	return nil
}
func preflight(b []byte) error {
	if len(b) == 0 || len(b) > MaxBytes {
		return Error("json_bounds")
	}
	if sourcecohort.Transport(b) != nil {
		return Error("json_transport")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var walk func(int) error
	walk = func(depth int) error {
		if depth > 32 {
			return Error("json_depth")
		}
		t, e := d.Token()
		if e != nil {
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
		n := 0
		for d.More() {
			if n >= MaxItems {
				return Error("collection_bounds")
			}
			n++
			if delim == '{' {
				k, e := d.Token()
				s, ok := k.(string)
				if e != nil || !ok || len(s) > 128 {
					return Error("key_bounds")
				}
			}
			if e := walk(depth + 1); e != nil {
				return e
			}
		}
		end, e := d.Token()
		if e != nil || delim == '{' && end != json.Delim('}') || delim == '[' && end != json.Delim(']') {
			return Error("json_shape")
		}
		return nil
	}
	if e := walk(0); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return Error("json_trailing")
	}
	return nil
}
func closed(b []byte, v any) error {
	if e := preflight(b); e != nil {
		return e
	}
	if sourcecohort.Decode(b, v) != nil {
		return Error("closed_json")
	}
	return nil
}
func UTC(s string) (time.Time, bool) {
	t, e := time.Parse(time.RFC3339Nano, s)
	return t, e == nil && strings.HasSuffix(s, "Z") && t.Year() > 1970
}
func keysSame(a, b []string) bool {
	if len(a) != 7 || len(b) != 7 {
		return false
	}
	a = append([]string{}, a...)
	b = append([]string{}, b...)
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if !identifier(a[i]) || a[i] != b[i] || i > 0 && (a[i] == a[i-1] || b[i] == b[i-1]) {
			return false
		}
	}
	return true
}
func registered(id string, ids []string) bool {
	if !identifier(id) || len(ids) < 1 || len(ids) > 400 {
		return false
	}
	copyIDs := append([]string{}, ids...)
	sort.Strings(copyIDs)
	for i, s := range copyIDs {
		if !identifier(s) || i > 0 && s == copyIDs[i-1] {
			return false
		}
	}
	i := sort.SearchStrings(copyIDs, id)
	return i < len(copyIDs) && copyIDs[i] == id
}
func pinnedCopy(p PinnedBytes) PinnedBytes {
	return PinnedBytes{File: p.File, Bytes: append([]byte{}, p.Bytes...)}
}
