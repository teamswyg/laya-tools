// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"unicode/utf8"
)

const retrievalTextInputBytes = 1575
const retrievalTextInputSHA = "585a236f648397f9ff2c35cebf048d64137d3a7d746f1e151a648151493d816e"
const retrievalPreparationSHA = "9f050acc724284c34b57da70216a894633c2fa5d7d56a4dcdd9f3eb10b8db4d0"
const retrievalSignedInputSHA = "47686828bbb0cd4d273db3b2087311357ea853b767f86fd787953ed76d452ed6"
const retrievalSourceRevision = "73c870b18e68b6e654b2b03f485aa3c9fab32cea"
const retrievalTextSchema = "riido-jwt-text-retrieval-input-v1"

//go:embed TEXT-INPUT.actual.public.v1.json
var retrievalFrozenTextInput string

type retrievalTextInput struct {
	Requests [4]string
	Captions [8]string
}
type retrievalTextCode string

const (
	retrievalTextOK             retrievalTextCode = ""
	retrievalTextPinUnknown     retrievalTextCode = "unknown_text_pin"
	retrievalTextShapeUnknown   retrievalTextCode = "unknown_text_shape"
	retrievalTextBindingUnknown retrievalTextCode = "unknown_text_binding"
)

// Call within input_setup after the old signed-stdin parser verifies its exact
// original bytes. No normalization/features/catalog preparation happens here.
// Embedded raw bytes belong in source closure, binary/startup/RSS and input SHA
// audit. Requests/captions are immutable string values, not learned examples.
func readRetrievalTextInput(signedInputSHA string) (retrievalTextInput, retrievalTextCode) {
	return decodeRetrievalTextInput([]byte(retrievalFrozenTextInput), signedInputSHA)
}
func decodeRetrievalTextInput(raw []byte, signedInputSHA string) (retrievalTextInput, retrievalTextCode) {
	digest := sha256.Sum256(raw)
	if len(raw) != retrievalTextInputBytes || hex.EncodeToString(digest[:]) != retrievalTextInputSHA {
		return retrievalTextInput{}, retrievalTextPinUnknown
	}
	if signedInputSHA != retrievalSignedInputSHA {
		return retrievalTextInput{}, retrievalTextBindingUnknown
	}
	return retrievalDecodeTextShape(raw)
}

// The exact pin already binds the complete JSON. This additional bounded
// decoder preserves case-sensitive/duplicate/schema and exact array checks.
// Literal UTF8 is required; no frozen source text uses Unicode escapes. Never
// silently repair unpaired escaped UTF16 or trim/normalize input captions.
func retrievalDecodeTextShape(raw []byte) (retrievalTextInput, retrievalTextCode) {
	if len(raw) > 8192 || !utf8.Valid(raw) || bytes.Contains(raw, []byte(`\u`)) {
		return retrievalTextInput{}, retrievalTextShapeUnknown
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return retrievalTextInput{}, retrievalTextShapeUnknown
	}
	var out retrievalTextInput
	var metadata [4]string
	var seen [6]bool
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return retrievalTextInput{}, retrievalTextShapeUnknown
		}
		index := -1
		switch key {
		case "schema":
			index = 0
		case "source_revision":
			index = 1
		case "source_preparation_sha256":
			index = 2
		case "signed_input_sha256":
			index = 3
		case "requests":
			index = 4
		case "captions":
			index = 5
		}
		if index < 0 || seen[index] {
			return retrievalTextInput{}, retrievalTextShapeUnknown
		}
		seen[index] = true
		if index < 4 {
			token, err := d.Token()
			value, ok := token.(string)
			if err != nil || !ok || len(value) > 128 {
				return retrievalTextInput{}, retrievalTextShapeUnknown
			}
			metadata[index] = value
		} else {
			count := 8
			if index == 4 {
				count = 4
			}
			texts, ok := retrievalDecodeTextArray(d, count)
			if !ok {
				return retrievalTextInput{}, retrievalTextShapeUnknown
			}
			if index == 4 {
				copy(out.Requests[:], texts[:4])
			} else {
				out.Captions = texts
			}
		}
	}
	t, err = d.Token()
	if err != nil || t != json.Delim('}') {
		return retrievalTextInput{}, retrievalTextShapeUnknown
	}
	if _, err = d.Token(); err != io.EOF {
		return retrievalTextInput{}, retrievalTextShapeUnknown
	}
	for _, present := range seen {
		if !present {
			return retrievalTextInput{}, retrievalTextShapeUnknown
		}
	}
	if metadata != [4]string{retrievalTextSchema, retrievalSourceRevision, retrievalPreparationSHA, retrievalSignedInputSHA} {
		return retrievalTextInput{}, retrievalTextBindingUnknown
	}
	return out, retrievalTextOK
}
func retrievalDecodeTextArray(d *json.Decoder, count int) ([8]string, bool) {
	var out [8]string
	if count != 4 && count != 8 {
		return out, false
	}
	t, err := d.Token()
	if err != nil || t != json.Delim('[') {
		return out, false
	}
	n := 0
	for d.More() {
		if n >= count {
			return [8]string{}, false
		}
		token, err := d.Token()
		value, ok := token.(string)
		if err != nil || !ok || len(value) == 0 || len(value) > retrievalTextLimit || !utf8.ValidString(value) {
			return [8]string{}, false
		}
		out[n] = value
		n++
	}
	t, err = d.Token()
	if err != nil || t != json.Delim(']') || n != count {
		return [8]string{}, false
	}
	return out, true
}
