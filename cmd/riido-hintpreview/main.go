// SPDX-License-Identifier: Apache-2.0
// riido-hintpreview shows experimental raw-text representations. It does not
// load a model, call a generator, verify an action, or change a router default.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/hintrelation"
	"github.com/teamswyg/laya-tools/internal/hintsymbol"
)

const maxJSONBytes = 16 << 10

var errInput = errors.New("hintpreview_invalid_input")

type input struct {
	Request    string   `json:"request"`
	Candidates []string `json:"candidates"`
}

type fingerprint struct {
	Schema    string `json:"schema"`
	Dimension int    `json:"dimension"`
	Entries   int    `json:"entries"`
	SHA256    string `json:"sha256"`
}

type candidate struct {
	Index      int                     `json:"index"`
	Attributes hintrelation.Attributes `json:"grammar_recognition"`
	Legacy     fingerprint             `json:"legacy"`
	Symbol     fingerprint             `json:"symbol"`
}

type output struct {
	RelationSchema  string                                                      `json:"relation_schema"`
	ColumnNames     [hintrelation.Dimension]string                              `json:"relation_column_names"`
	AgreementPolicy string                                                      `json:"agreement_policy"`
	Schema          string                                                      `json:"schema"`
	PreviewOnly     bool                                                        `json:"preview_only"`
	ModelCalls      int                                                         `json:"model_calls"`
	FitCalls        int                                                         `json:"Fit_calls"`
	Request         hintrelation.Attributes                                     `json:"request_grammar_recognition"`
	Candidates      []candidate                                                 `json:"candidates"`
	Columns         [hintrelation.Dimension][hintrelation.MaxCandidates]float32 `json:"relation_columns"`
	Order           []int                                                       `json:"agreement_order"`
	Scores          []int                                                       `json:"agreement_scores"`
}

// The fingerprint is ordered uint16 LE index + float64 LE value, without
// decimal formatting, capacity/header bytes, source text, or coefficients.
// The legacy terminal overlap entry retains its original position.
func describe(schema string, n int, at func(int) (int, float64)) fingerprint {
	h := sha256.New()
	var buf [10]byte
	for i := 0; i < n; i++ {
		index, value := at(i)
		binary.LittleEndian.PutUint16(buf[:2], uint16(index))
		binary.LittleEndian.PutUint64(buf[2:], math.Float64bits(value))
		_, _ = h.Write(buf[:])
	}
	return fingerprint{schema, hintlearn.Dimension, n, hex.EncodeToString(h.Sum(nil))}
}

// encoding/json otherwise replaces invalid UTF-8 and unpaired UTF-16 escapes
// with U+FFFD. Validate source scalar strings before that lossy replacement.
func validScalars(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	hex4 := func(start int) (uint16, bool) {
		if start+4 > len(b) {
			return 0, false
		}
		var n uint16
		for _, c := range b[start : start+4] {
			n <<= 4
			switch {
			case c >= '0' && c <= '9':
				n += uint16(c - '0')
			case c >= 'a' && c <= 'f':
				n += uint16(c-'a') + 10
			case c >= 'A' && c <= 'F':
				n += uint16(c-'A') + 10
			default:
				return 0, false
			}
		}
		return n, true
	}
	inside := false
	for i := 0; i < len(b); i++ {
		if b[i] == '"' {
			inside = !inside
			continue
		}
		if !inside || b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue // The JSON decoder validates all other escapes and syntax.
		}
		n, ok := hex4(i + 1)
		if !ok {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return false
			}
			low, ok := hex4(i + 3)
			if !ok || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return !inside
}

// Two explicit fields avoid map allocation and reject duplicate, unknown,
// missing, null and wrong-type fields instead of silently overwriting them.
func decodeInput(b []byte) (input, error) {
	if !validScalars(b) {
		return input{}, errInput
	}
	d := json.NewDecoder(bytes.NewReader(b))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return input{}, errInput
	}
	var v input
	var seen [2]bool
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return input{}, errInput
		}
		switch key {
		case "request":
			if seen[0] {
				return input{}, errInput
			}
			seen[0] = true
			token, err = d.Token()
			text, ok := token.(string)
			if err != nil || !ok {
				return input{}, errInput
			}
			v.Request = text
		case "candidates":
			if seen[1] {
				return input{}, errInput
			}
			seen[1] = true
			token, err = d.Token()
			if err != nil || token != json.Delim('[') {
				return input{}, errInput
			}
			for d.More() {
				token, err = d.Token()
				text, ok := token.(string)
				if err != nil || !ok || len(v.Candidates) == hintrelation.MaxCandidates {
					return input{}, errInput
				}
				v.Candidates = append(v.Candidates, text)
			}
			token, err = d.Token()
			if err != nil || token != json.Delim(']') || len(v.Candidates) == 0 {
				return input{}, errInput
			}
		default:
			return input{}, errInput
		}
	}
	token, err = d.Token()
	if err != nil || token != json.Delim('}') || !seen[0] || !seen[1] {
		return input{}, errInput
	}
	if _, err := d.Token(); err != io.EOF {
		return input{}, errInput
	}
	return v, nil
}

func run(in io.Reader, out io.Writer) error {
	b, err := io.ReadAll(io.LimitReader(in, maxJSONBytes+1))
	if err != nil {
		return errInput
	}
	if len(b) > maxJSONBytes {
		return errInput
	}
	v, err := decodeInput(b)
	if err != nil {
		return errInput
	}
	var texts [hintrelation.MaxCandidates]string
	copy(texts[:], v.Candidates)
	columns, err := hintrelation.Build(v.Request, texts, len(v.Candidates))
	if err != nil {
		return errInput
	}
	result := output{Schema: "riido-hint-representation-preview-v1", RelationSchema: hintrelation.Schema,
		ColumnNames:     [hintrelation.Dimension]string{"direction_match", "direction_conflict", "lower_match", "lower_conflict", "upper_match", "upper_conflict", "stop_match", "stop_conflict"},
		AgreementPolicy: "recognized_match_plus1_conflict_minus1_stable_display_ties_all_candidates_retained", PreviewOnly: true, Columns: columns.Values()}
	result.Request, _ = columns.Attributes(-1)
	// Validate and compute bounded symbol features before the legacy extractor,
	// whose original public API does not itself enforce this preview's budgets.
	var symbols [hintrelation.MaxCandidates]fingerprint
	for i, text := range v.Candidates {
		features, extractErr := hintsymbol.Extract(v.Request, text)
		if extractErr != nil {
			return errInput
		}
		// Retain only the compact fingerprint, not eight large cross backing
		// arrays. This is an ownership change, not a measured peak-RSS bound.
		symbols[i] = describe(hintsymbol.Schema, len(features), func(j int) (int, float64) { return features[j].Index, features[j].Value })
	}
	for i, text := range v.Candidates {
		legacy := hintlearn.Features(v.Request, text)
		attrs, _ := columns.Attributes(i)
		result.Candidates = append(result.Candidates, candidate{
			Index: i, Attributes: attrs,
			Legacy: describe("riido-legacy-word-hash8192-v1", len(legacy), func(j int) (int, float64) { return legacy[j].Index, legacy[j].Value }),
			Symbol: symbols[i],
		})
	}
	order, scores, err := columns.AgreementOrder()
	if err != nil {
		return errInput
	}
	result.Order, result.Scores = order[:len(v.Candidates)], scores[:len(v.Candidates)]
	return json.NewEncoder(out).Encode(result)
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		// Do not echo raw user input, paths, or source code in errors.
		fmt.Fprintln(os.Stderr, "hintpreview: invalid bounded JSON input or output failure")
		os.Exit(1)
	}
}
