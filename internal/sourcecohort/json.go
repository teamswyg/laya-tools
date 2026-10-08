package sourcecohort

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Decode uses closed typed shapes. Required fields cannot disappear behind Go
// zero values. Only explicitly nullable pointers and optional receipt fields
// permit null; arrays are bounded and object keys are case sensitive.
func Decode(data []byte, target any) error {
	if len(data) == 0 || len(data) > 8<<20 || !utf8.Valid(data) || !unicodeEscapes(data) {
		return errCode("json_bounds")
	}
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return errCode("json_target")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	fresh := reflect.New(v.Elem().Type()).Elem()
	if e := value(d, fresh, 0, false); e != nil {
		return e
	}
	if _, e := d.Token(); e != io.EOF {
		return errCode("json_trailing")
	}
	v.Elem().Set(fresh)
	return nil
}
func value(d *json.Decoder, v reflect.Value, depth int, optional bool) error {
	if depth > 32 {
		return errCode("json_depth")
	}
	t, e := d.Token()
	if e != nil || !v.CanSet() {
		return errCode("json_type")
	}
	if t == nil {
		if v.Kind() == reflect.Pointer || optional && v.Kind() == reflect.Slice {
			return nil
		}
		return errCode("json_null")
	}
	if v.Kind() == reflect.Pointer {
		v.Set(reflect.New(v.Type().Elem()))
		return withToken(d, v.Elem(), depth, t)
	}
	return withToken(d, v, depth, t)
}
func withToken(d *json.Decoder, v reflect.Value, depth int, t json.Token) error {
	switch v.Kind() {
	case reflect.Struct:
		if t != json.Delim('{') || v.NumField() > 64 {
			return errCode("json_object")
		}
		seen := make([]bool, v.NumField())
		for d.More() {
			k, e := d.Token()
			if e != nil {
				return errCode("json_key")
			}
			index := -1
			for i := range seen {
				if k == strings.Split(v.Type().Field(i).Tag.Get("json"), ",")[0] {
					index = i
					break
				}
			}
			if index < 0 || seen[index] {
				return errCode("json_key")
			}
			seen[index] = true
			if e := value(d, v.Field(index), depth+1, strings.Contains(v.Type().Field(index).Tag.Get("json"), ",omitempty")); e != nil {
				return e
			}
		}
		for i, ok := range seen {
			if !ok && !strings.Contains(v.Type().Field(i).Tag.Get("json"), ",omitempty") {
				return errCode("json_missing")
			}
		}
		end, e := d.Token()
		if e != nil || end != json.Delim('}') {
			return errCode("json_object")
		}
	case reflect.Slice:
		if t != json.Delim('[') {
			return errCode("json_array")
		}
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		for d.More() {
			if v.Len() >= 4096 {
				return errCode("json_array_bounds")
			}
			item := reflect.New(v.Type().Elem()).Elem()
			if e := value(d, item, depth+1, false); e != nil {
				return e
			}
			v.Set(reflect.Append(v, item))
		}
		end, e := d.Token()
		if e != nil || end != json.Delim(']') {
			return errCode("json_array")
		}
	case reflect.String:
		s, ok := t.(string)
		if !ok || len(s) > 1<<16 || strings.ContainsRune(s, 0) {
			return errCode("json_string")
		}
		v.SetString(s)
	case reflect.Bool:
		b, ok := t.(bool)
		if !ok {
			return errCode("json_bool")
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int64:
		n, ok := t.(json.Number)
		if !ok {
			return errCode("json_number")
		}
		i, e := n.Int64()
		if e != nil || v.OverflowInt(i) {
			return errCode("json_number")
		}
		v.SetInt(i)
	default:
		return errCode("json_type")
	}
	return nil
}

func unicodeEscapes(data []byte) bool {
	in := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			in = !in
			continue
		}
		if !in || data[i] != '\\' {
			continue
		}
		if i+1 >= len(data) {
			return false
		}
		if data[i+1] != 'u' {
			i++
			continue
		}
		if i+6 > len(data) {
			return false
		}
		n, e := strconv.ParseUint(string(data[i+2:i+6]), 16, 16)
		if e != nil {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+12 > len(data) || string(data[i+6:i+8]) != "\\u" {
				return false
			}
			low, e := strconv.ParseUint(string(data[i+8:i+12]), 16, 16)
			if e != nil || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 11
		} else {
			if n >= 0xdc00 && n <= 0xdfff {
				return false
			}
			i += 5
		}
	}
	return !in
}
