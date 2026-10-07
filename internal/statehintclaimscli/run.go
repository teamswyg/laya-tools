// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintclaimscli serves explicit local claim models without app writes.
package statehintclaimscli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

var errInput = errors.New("claims requires a valid local RSC model and bounded text or JSONL")

type response struct {
	Schema     string                      `json:"schema"`
	Mode       string                      `json:"mode"`
	Prediction *statehintclaims.Prediction `json:"prediction,omitempty"`
	Reason     string                      `json:"reason,omitempty"`
	Qualified  bool                        `json:"semantic_quality_qualified"`
	Writes     bool                        `json:"mutation_executed"`
}

// Exactly one text field: duplicates, aliases, extra fields and trailing values
// cannot silently change a request. No parsing map or per-stream cache is used.
func request(line []byte) (string, bool) {
	if len(line) > 32768 || !utf8.Valid(line) {
		return "", false
	}
	d := json.NewDecoder(bytes.NewReader(line))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') || !d.More() {
		return "", false
	}
	k, e := d.Token()
	if e != nil || k != "text" {
		return "", false
	}
	v, e := d.Token()
	s, ok := v.(string)
	if e != nil || !ok || d.More() {
		return "", false
	}
	t, e = d.Token()
	if e != nil || t != json.Delim('}') {
		return "", false
	}
	_, e = d.Token()
	return s, e == io.EOF && len(s) <= statehintclaims.MaxTextBytes && utf8.ValidString(s)
}

// Run loads once and reuses a caller-owned workspace for the JSONL stream.
// A model is mandatory: no download, rule fallback or default registration.
func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riidolaya claims", flag.ContinueOnError)
	f.SetOutput(errOut)
	path := f.String("model", "", "explicit local research .rsc model")
	text := f.String("text", "", "one bounded content item")
	jsonl := f.Bool("jsonl", false, "warm stream of {text:string} requests")
	_ = f.Bool("json", true, "JSON research output")
	if f.Parse(args) != nil || *path == "" {
		return errInput
	}
	query := strings.Join(f.Args(), " ")
	if *text != "" {
		if query != "" {
			return errInput
		}
		query = *text
	}
	if *jsonl && query != "" || !*jsonl && query == "" {
		return errInput
	}
	file, e := os.Open(*path)
	if e != nil {
		return errInput
	}
	model, e := statehintclaims.Load(file)
	closeErr := file.Close()
	if e != nil || closeErr != nil {
		return errInput
	}
	var workspace statehintclaims.Workspace
	encoder := json.NewEncoder(out)
	process := func(s string, valid bool) error {
		r := response{Schema: "riido-claim-hints-response-v1", Mode: "research_preview"}
		if !valid {
			r.Reason = "invalid_request"
		} else {
			p, e := model.Predict(s, &workspace)
			if e != nil {
				r.Reason = "input_or_model_out_of_scope"
			} else {
				r.Prediction = &p
			}
		}
		return encoder.Encode(r)
	}
	if !*jsonl {
		return process(query, true)
	}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 32769)
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		text, valid := request(scanner.Bytes())
		if e := process(text, valid); e != nil {
			return e
		}
	}
	if scanner.Err() != nil {
		return errInput
	}
	return nil
}
