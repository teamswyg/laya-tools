// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package questioncuecli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintcue"
)

var errOptions = errors.New("invalid question-cue options; use --role and --text, or --jsonl")
var errStream = errors.New("question-cue input stream unavailable or over budget")

type request struct {
	Text string
	Role statehintcue.Role
}

type response struct {
	Schema           string              `json:"schema"`
	Mode             string              `json:"mode"`
	Status           string              `json:"status"`
	Reason           string              `json:"reason,omitempty"`
	Cue              statehintcue.Result `json:"cue"`
	MutationExecuted bool                `json:"mutation_executed"`
}

// Run emits only an experimental punctuation observation. It loads no model,
// echoes no text, and never creates an annotation or state-change plan.
func Run(args []string, in io.Reader, out, errOut io.Writer) error {
	options := flag.NewFlagSet("question-cue", flag.ContinueOnError)
	options.SetOutput(io.Discard) // Unknown arguments may contain private text.
	text := options.String("text", "", "one local content unit")
	role := options.String("role", "", "explicit prose, metadata, or code role")
	jsonl := options.Bool("jsonl", false, "bounded typed JSONL from stdin")
	_ = options.Bool("json", true, "JSON is the default output")
	if err := options.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = io.WriteString(errOut, "question-cue --role prose|metadata|code --text TEXT\nquestion-cue --jsonl: one {\"text\":\"...\",\"role\":\"prose\"} per line\nExperimental punctuation cue; no model, probability, annotation or state change.\n")
			return err
		}
		return errOptions
	}
	if options.NArg() != 0 || (*jsonl && (*text != "" || *role != "")) || (!*jsonl && (*text == "" || *role == "")) {
		return errOptions
	}
	encoder := json.NewEncoder(out)
	process := func(value request) error {
		cue, err := statehintcue.Inspect(context.Background(), value.Text, value.Role)
		result := response{Schema: "riido-question-cue-v1", Mode: "cue_preview", Status: "observed", Cue: cue}
		if err != nil {
			result.Status, result.Reason = "unknown", "input_or_role_out_of_scope"
		} else if cue.Guard != "" {
			result.Status = "guarded"
		}
		return encoder.Encode(result)
	}
	if !*jsonl {
		return process(request{Text: *text, Role: statehintcue.Role(*role)})
	}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 16384)
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		value, ok := decodeRequest(scanner.Bytes())
		if !ok {
			if err := encoder.Encode(response{Schema: "riido-question-cue-v1", Mode: "cue_preview", Status: "unknown", Reason: "invalid_typed_request"}); err != nil {
				return err
			}
			continue
		}
		if err := process(value); err != nil {
			return err
		}
	}
	if scanner.Err() != nil {
		return errStream
	}
	return nil
}

// Two explicit string fields avoid maps, duplicate/case-alias acceptance and
// permissive fallback. Role declaration is a caller claim, not owner authority.
func decodeRequest(data []byte) (request, bool) {
	var result request
	if len(data) > 16384 || !utf8.Valid(data) {
		return result, false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return result, false
	}
	var seen [2]bool
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return request{}, false
		}
		index := -1
		switch key {
		case "text":
			index = 0
		case "role":
			index = 1
		}
		if index < 0 || seen[index] {
			return request{}, false
		}
		value, err := decoder.Token()
		text, ok := value.(string)
		if err != nil || !ok {
			return request{}, false
		}
		if index == 0 {
			result.Text = text
		} else {
			result.Role = statehintcue.Role(text)
		}
		seen[index] = true
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || !seen[0] || !seen[1] || result.Text == "" || strings.TrimSpace(string(result.Role)) != string(result.Role) {
		return request{}, false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return request{}, false
	}
	return result, true
}
