// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package claimsdemo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

// Retain the original token decoder as an independent compatibility oracle.
func legacyDecodeText(body []byte) (string, error) {
	if !utf8.Valid(body) {
		return "", errors.New("UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(body))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return "", errors.New("object")
	}
	seen := false
	var text string
	for d.More() {
		k, err := d.Token()
		if err != nil || k != "text" || seen {
			return "", errors.New("key")
		}
		v, err := d.Token()
		if err != nil {
			return "", err
		}
		var ok bool
		text, ok = v.(string)
		if !ok {
			return "", errors.New("string")
		}
		seen = true
	}
	t, err = d.Token()
	if err != nil || t != json.Delim('}') || !seen {
		return "", errors.New("end")
	}
	if _, err = d.Token(); err != io.EOF {
		return "", errors.New("trailing")
	}
	return text, nil
}

func FuzzDecodeTextCompatibility(f *testing.F) {
	for _, body := range []string{`{"text":""}`, `{"text":"한글 🛠️"}`, `{"te\u0078t":"escaped key"}`, `{"Text":"wrong case"}`, `{"text":null}`, `{"text":"one","text":"two"}`, `{"text":"one","te\u0078t":"two"}`, `{"text":"one","extra":0}`, `{"text":"\ud800"}`, `{"text":"\ud83d\ude00"}`, `{"text":"\u0000"}`, `{"text":"one"} {}`, `null`, `{}`, `{"text":[1]}`, "{\"text\":\"\xff\"}"} {
		f.Add([]byte(body))
	}
	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > MaxRequestBytes {
			return
		}
		want, oldErr := legacyDecodeText(body)
		got, newErr := decodeText(body)
		if (oldErr == nil) != (newErr == nil) || oldErr == nil && got != want {
			t.Fatalf("decoder compatibility mismatch: old accepted=%v new accepted=%v", oldErr == nil, newErr == nil)
		}
	})
}

func BenchmarkDecodeText(b *testing.B) {
	for _, size := range []int{32, 256, 1024, 4096} {
		body := []byte(`{"text":"` + strings.Repeat("x", size) + `"}`)
		for _, entry := range []struct {
			name string
			run  func([]byte) (string, error)
		}{{"legacy", legacyDecodeText}, {"direct", decodeText}} {
			b.Run(fmt.Sprintf("%d/%s", size, entry.name), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_, err := entry.run(body)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
