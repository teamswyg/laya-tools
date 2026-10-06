// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Local content classification and shadow plans; no external mutation or model router.
package statehintcli

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

type request struct {
	Text    string               `json:"text"`
	Context *statehint.PlanInput `json:"context,omitempty"`
}
type response struct {
	Schema           string                `json:"schema"`
	Status           string                `json:"status"`
	Prediction       *statehint.Prediction `json:"prediction,omitempty"`
	Plan             *statehint.PlanResult `json:"plan,omitempty"`
	MutationExecuted bool                  `json:"mutation_executed"`
	Reason           string                `json:"reason,omitempty"`
	EmojiCode        string                `json:"emoji_code,omitempty"`
	Mode             string                `json:"mode"`
}

// JSON objects reject duplicate and case-alias keys before struct decoding.
// All maps are bounded input parsing; prediction state uses caller arrays.
func validRequest(b []byte) bool {
	if !utf8.Valid(b) || len(b) > 16384 || len(bytes.TrimSpace(b)) == 0 || bytes.TrimSpace(b)[0] != '{' {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var value func(int) bool
	value = func(depth int) bool {
		if depth > 12 {
			return false
		}
		t, e := d.Token()
		if e != nil {
			return false
		}
		if delimiter, ok := t.(json.Delim); ok {
			if delimiter != '{' && delimiter != '[' {
				return false
			}
			names := map[string]bool{}
			for d.More() {
				if delimiter == '{' {
					k, e := d.Token()
					s, ok := k.(string)
					if e != nil || !ok {
						return false
					}
					s = strings.ToLower(s)
					if names[s] {
						return false
					}
					names[s] = true
				}
				if !value(depth + 1) {
					return false
				}
			}
			end, e := d.Token()
			return e == nil && ((delimiter == '{' && end == json.Delim('}')) || (delimiter == '[' && end == json.Delim(']')))
		}
		return true
	}
	if !value(0) {
		return false
	}
	if _, e := d.Token(); e != io.EOF {
		return false
	}
	var top map[string]json.RawMessage
	if json.Unmarshal(b, &top) != nil {
		return false
	}
	for name := range top {
		if name != "text" && name != "context" {
			return false
		}
	}
	text, ok := top["text"]
	if !ok || bytes.Equal(bytes.TrimSpace(text), []byte("null")) {
		return false
	}
	var s string
	if json.Unmarshal(text, &s) != nil || s == "" {
		return false
	}
	if context, ok := top["context"]; ok && bytes.Equal(bytes.TrimSpace(context), []byte("null")) {
		return false
	}
	return true
}

// Run handles only local inputs. Without --model it uses an explicitly unlearned
// rule control; a trained artifact is loaded once for the complete JSONL stream.
func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riidolaya state-hint", flag.ContinueOnError)
	f.SetOutput(errOut)
	modelPath := f.String("model", "", "optional local .rsh model; no automatic download")
	text := f.String("text", "", "one content item")
	jsonl := f.Bool("jsonl", false, "read one bounded JSON request per line")
	_ = f.Bool("json", true, "JSON output for people and agents")
	if err := f.Parse(args); err != nil {
		return err
	}
	query := strings.Join(f.Args(), " ")
	if *text != "" && query != "" {
		return fmt.Errorf("use --text or positional text")
	}
	if *text != "" {
		query = *text
	}
	if *jsonl && query != "" {
		return fmt.Errorf("JSONL and one text are separate modes")
	}
	var model *statehint.Model
	var wide *statehintwide.Model
	if *modelPath != "" {
		file, e := os.Open(*modelPath)
		if e != nil {
			return fmt.Errorf("cannot open local statehint model")
		}
		var header [8]byte
		_, e = io.ReadFull(file, header[:])
		if e == nil {
			reader := io.MultiReader(bytes.NewReader(header[:]), file)
			switch binary.LittleEndian.Uint16(header[4:6]) {
			case 1:
				model, e = statehint.Load(reader)
			case 2:
				wide, e = statehintwide.Load(reader)
			default:
				e = statehint.ErrArtifact
			}
		}
		ce := file.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	var work statehint.Workspace
	var wideWork statehintwide.Workspace
	encoder := json.NewEncoder(out)
	process := func(r request) error {
		var p statehint.Prediction
		var e error
		if wide != nil {
			p, e = wide.Predict(r.Text, &wideWork)
		} else if model == nil {
			p, e = statehint.RuleBaseline(r.Text)
		} else {
			p, e = model.Predict(r.Text, &work)
		}
		if e != nil {
			return encoder.Encode(response{Schema: "riido-statehint-response-v1", Status: "unknown", Reason: "input_or_model_out_of_scope"})
		}
		x := response{Schema: "riido-statehint-response-v1", Status: "suggested", Prediction: &p, Mode: "shadow"}
		if p.Confidence >= .9 && p.Intent != statehint.Unclear && p.GuardReason == "" {
			for _, emoji := range statehint.DefaultEmojis() {
				if emoji.Intent == p.Intent {
					x.EmojiCode = emoji.Code
					break
				}
			}
		}
		if r.Context != nil {
			// Context owns its allowlist and existing display. A default intent
			// icon must not contradict or bypass the guarded annotation plan.
			x.EmojiCode = ""
			ctx := *r.Context
			if ctx.MinConfidence == 0 {
				ctx.MinConfidence = .9
			}
			if ctx.MinMargin == 0 {
				ctx.MinMargin = .05
			}
			plan, e := statehint.Plan(ctx, p)
			if e != nil {
				x.Status = "unknown"
				x.Reason = "invalid_context"
			} else {
				x.Plan = &plan
				x.EmojiCode = plan.EmojiCode
			}
		}
		return encoder.Encode(x)
	}
	if !*jsonl {
		if query == "" {
			return fmt.Errorf("provide text or --jsonl")
		}
		return process(request{Text: query})
	}
	s := bufio.NewScanner(in)
	s.Buffer(make([]byte, 4096), 16384)
	for s.Scan() {
		if len(s.Bytes()) == 0 {
			continue
		}
		var r request
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		var e error
		if !validRequest(s.Bytes()) {
			e = fmt.Errorf("invalid typed request")
		} else {
			e = d.Decode(&r)
		}
		if e == nil {
			var extra any
			if d.Decode(&extra) != io.EOF {
				e = fmt.Errorf("trailing data")
			}
		}
		if e != nil {
			if e = encoder.Encode(response{Schema: "riido-statehint-response-v1", Status: "unknown", Reason: "invalid_request"}); e != nil {
				return e
			}
			continue
		}
		if e = process(r); e != nil {
			return e
		}
	}
	return s.Err()
}
