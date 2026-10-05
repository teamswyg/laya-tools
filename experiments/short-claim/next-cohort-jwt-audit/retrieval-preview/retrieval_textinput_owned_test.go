// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"strings"
	"testing"
)

type retrievalSyntheticTextShape struct {
	Schema      string   `json:"schema"`
	Revision    string   `json:"source_revision"`
	Preparation string   `json:"source_preparation_sha256"`
	Signed      string   `json:"signed_input_sha256"`
	Requests    []string `json:"requests"`
	Captions    []string `json:"captions"`
}

func retrievalSyntheticTextWire(requests, captions int) []byte {
	value := retrievalSyntheticTextShape{Schema: retrievalTextSchema, Revision: retrievalSourceRevision, Preparation: retrievalPreparationSHA, Signed: retrievalSignedInputSHA, Requests: make([]string, requests), Captions: make([]string, captions)}
	for i := range value.Requests {
		value.Requests[i] = "synthetic query"
	}
	for i := range value.Captions {
		value.Captions[i] = "synthetic caption"
	}
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}
func TestRetrievalEmbeddedInputExactPinAndSeparateBinding(t *testing.T) {
	out, code := readRetrievalTextInput(retrievalSignedInputSHA)
	if code != retrievalTextOK || len(out.Requests) != 4 || len(out.Captions) != 8 {
		t.Fatal("frozen text shape/pin failed")
	}
	if _, code := readRetrievalTextInput("wrong-signed-input"); code != retrievalTextBindingUnknown {
		t.Fatal("wrong original signed-input binding accepted")
	}
	raw := append([]byte(retrievalFrozenTextInput), ' ')
	if _, code := decodeRetrievalTextInput(raw, retrievalSignedInputSHA); code != retrievalTextPinUnknown {
		t.Fatal("changed complete embed bytes accepted")
	}
	out.Requests[0] = "caller mutation"
	again, code := readRetrievalTextInput(retrievalSignedInputSHA)
	if code != retrievalTextOK || again.Requests[0] == out.Requests[0] {
		t.Fatal("caller mutation changed embedded strings")
	}
}
func TestRetrievalBoundedShapeWithoutAnyRanking(t *testing.T) {
	raw := retrievalSyntheticTextWire(4, 8)
	if _, code := retrievalDecodeTextShape(raw); code != retrievalTextOK {
		t.Fatal("owned synthetic six-field shape rejected")
	}
	for _, shape := range [][2]int{{3, 8}, {5, 8}, {4, 7}, {4, 9}} {
		if _, code := retrievalDecodeTextShape(retrievalSyntheticTextWire(shape[0], shape[1])); code != retrievalTextShapeUnknown {
			t.Fatal("short/excess text array accepted")
		}
	}
	malformed := []string{
		strings.Replace(string(raw), `"schema":`, `"Schema":`, 1),
		strings.Replace(string(raw), `"requests":`, `"schema":"duplicate","requests":`, 1),
		strings.Replace(string(raw), `"requests":`, `"alien":0,"requests":`, 1),
		strings.Replace(string(raw), `"synthetic query"`, `null`, 1),
		strings.Replace(string(raw), `"synthetic query"`, `""`, 1),
		strings.Replace(string(raw), `"synthetic query"`, `"\ud800"`, 1),
		string(raw) + ` {}`,
	}
	for _, b := range malformed {
		if _, code := retrievalDecodeTextShape([]byte(b)); code != retrievalTextShapeUnknown {
			t.Fatal("duplicate/case/type/Unicode/trailing malformed text accepted")
		}
	}
	if _, code := retrievalDecodeTextShape([]byte{0xff}); code != retrievalTextShapeUnknown {
		t.Fatal("invalid raw UTF8 repaired")
	}
	if _, code := retrievalDecodeTextShape(make([]byte, 8193)); code != retrievalTextShapeUnknown {
		t.Fatal("raw shape bound ignored")
	}
	if _, code := retrievalDecodeTextShape([]byte(strings.Replace(string(raw), retrievalPreparationSHA, strings.Repeat("0", 64), 1))); code != retrievalTextBindingUnknown {
		t.Fatal("wrong source preparation accepted")
	}
}
