// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

var ErrStudy = errors.New("statehint study configuration, pins or stage invalid")

// strictJSON accepts exactly the exported JSON fields of these private study
// structs. Maps, interfaces, pointers, nulls, aliases and duplicate keys are
// excluded. Finite input/slice limits bound the reflection walk.
func strictJSON(data []byte, target any) error {
	if len(data) == 0 || len(data) > 1<<20 || !utf8.Valid(data) {
		return ErrStudy
	}
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ErrStudy
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := strictValue(decoder, v.Elem()); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrStudy
	}
	return nil
}

func strictValue(decoder *json.Decoder, value reflect.Value) error {
	token, err := decoder.Token()
	if err != nil || !value.CanSet() || token == nil {
		return ErrStudy
	}
	switch value.Kind() {
	case reflect.Struct:
		if token != json.Delim('{') || value.NumField() > 64 {
			return ErrStudy
		}
		seen := make([]bool, value.NumField())
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return ErrStudy
			}
			index := -1
			for i := range seen {
				if key == value.Type().Field(i).Tag.Get("json") {
					index = i
					break
				}
			}
			if index < 0 || seen[index] || strictValue(decoder, value.Field(index)) != nil {
				return ErrStudy
			}
			seen[index] = true
		}
		for _, found := range seen {
			if !found {
				return ErrStudy
			}
		}
		if end, err := decoder.Token(); err != nil || end != json.Delim('}') {
			return ErrStudy
		}
	case reflect.Array, reflect.Slice:
		if token != json.Delim('[') {
			return ErrStudy
		}
		if value.Kind() == reflect.Slice {
			value.Set(reflect.MakeSlice(value.Type(), 0, 0))
		}
		count := 0
		for decoder.More() {
			if count == 2400 || value.Kind() == reflect.Array && count >= value.Len() {
				return ErrStudy
			}
			if value.Kind() == reflect.Slice {
				value.Set(reflect.Append(value, reflect.Zero(value.Type().Elem())))
			}
			if strictValue(decoder, value.Index(count)) != nil {
				return ErrStudy
			}
			count++
		}
		if value.Kind() == reflect.Array && count != value.Len() {
			return ErrStudy
		}
		if end, err := decoder.Token(); err != nil || end != json.Delim(']') {
			return ErrStudy
		}
	case reflect.String:
		s, ok := token.(string)
		if !ok || !utf8.ValidString(s) || strings.ContainsRune(s, 0) {
			return ErrStudy
		}
		value.SetString(s)
	case reflect.Bool:
		b, ok := token.(bool)
		if !ok {
			return ErrStudy
		}
		value.SetBool(b)
	case reflect.Int, reflect.Int64:
		n, ok := token.(json.Number)
		if !ok {
			return ErrStudy
		}
		integer, err := n.Int64()
		if err != nil || value.OverflowInt(integer) {
			return ErrStudy
		}
		value.SetInt(integer)
	case reflect.Float64:
		n, ok := token.(json.Number)
		if !ok {
			return ErrStudy
		}
		f, err := n.Float64()
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return ErrStudy
		}
		value.SetFloat(f)
	default:
		return ErrStudy
	}
	return nil
}
