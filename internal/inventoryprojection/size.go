package inventoryprojection

import (
	"reflect"
	"strconv"

	"github.com/teamswyg/laya-tools/internal/semanticframe"
)

// Only already bounded typed graphs reach this counter. It follows the compact
// encoding/json representation without allocating an oversized encoded graph.
func graphFits(g semanticframe.InventoryGraph) bool {
	remaining := MaxBytes
	add := func(n int) bool {
		if n > remaining {
			return false
		}
		remaining -= n
		return true
	}
	str := func(s string) bool {
		if !add(2 + len(s)) {
			return false
		}
		for i := 0; i < len(s); i++ {
			extra := 0
			switch s[i] {
			case '"', '\\', '\b', '\f', '\n', '\r', '\t':
				extra = 1
			case '<', '>', '&':
				extra = 5
			default:
				if s[i] < 0x20 {
					extra = 5
				} else if s[i] == 0xe2 && i+2 < len(s) && s[i+1] == 0x80 && (s[i+2] == 0xa8 || s[i+2] == 0xa9) {
					extra = 3
					i += 2
				}
			}
			if !add(extra) {
				return false
			}
		}
		return true
	}
	var visit func(reflect.Value) bool
	visit = func(v reflect.Value) bool {
		switch v.Kind() {
		case reflect.Struct:
			if !add(2) {
				return false
			}
			for i := 0; i < v.NumField(); i++ {
				if i > 0 && !add(1) {
					return false
				}
				if !str(v.Type().Field(i).Tag.Get("json")) || !add(1) || !visit(v.Field(i)) {
					return false
				}
			}
			return true
		case reflect.Slice:
			if v.IsNil() {
				return add(4)
			}
			if !add(2) {
				return false
			}
			for i := 0; i < v.Len(); i++ {
				if i > 0 && !add(1) {
					return false
				}
				if !visit(v.Index(i)) {
					return false
				}
			}
			return true
		case reflect.Pointer:
			if v.IsNil() {
				return add(4)
			}
			return visit(v.Elem())
		case reflect.String:
			return str(v.String())
		case reflect.Int, reflect.Int64:
			var b [20]byte
			return add(len(strconv.AppendInt(b[:0], v.Int(), 10)))
		case reflect.Bool:
			if v.Bool() {
				return add(4)
			}
			return add(5)
		}
		return false
	}
	return visit(reflect.ValueOf(g))
}
