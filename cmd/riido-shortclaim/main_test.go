package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func example() string {
	return `{"schema":"riido-short-behavior-claim-v1","request":"preserve input order","provenance":"original-example","candidates":[{"id":"a","text":"Reverse input order."},{"id":"b","text":"Preserve input order."}]}`
}

func TestResidentStreamKeepsRecordsSeparate(t *testing.T) {
	input := example()
	var stdout, diagnostic bytes.Buffer
	if err := run([]string{"--stream", "--baseline", "lexical_ordered"}, strings.NewReader(input+"\r\n"+input), &stdout, &diagnostic); err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(&stdout)
	for i := 0; i < 2; i++ {
		var result output
		if err := decoder.Decode(&result); err != nil {
			t.Fatal(err)
		}
		if result.Status != "unverified_heuristic" || len(result.Candidates) != 2 || result.Candidates[0].ID != "b" || len(result.InputSHA256) != 64 {
			t.Fatalf("unexpected record: %+v", result)
		}
	}
	var extra output
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("extra record: %v", err)
	}
	if diagnostic.Len() != 0 {
		t.Fatal("unexpected diagnostics")
	}
}

func TestStreamRejectsBeforeLaterRecordsAndDoesNotEchoInput(t *testing.T) {
	for _, bad := range []string{
		`{"schema":"riido-short-behavior-claim-v1","schema":"sensitive-example"}`,
		strings.Repeat("sensitive-example", shortclaim.MaxJSONBytes),
		"",
	} {
		var stdout, diagnostic bytes.Buffer
		err := run([]string{"--stream"}, strings.NewReader(example()+"\n"+bad+"\n"+example()+"\n"), &stdout, &diagnostic)
		if err == nil {
			t.Fatal("invalid record accepted")
		}
		if strings.Count(stdout.String(), "\n") != 1 {
			t.Fatal("later record emitted")
		}
		if strings.Contains(err.Error()+diagnostic.String()+stdout.String(), "sensitive-example") {
			t.Fatal("input echoed")
		}
	}
}

func TestAllControlsRetainCandidatesAndHideRawText(t *testing.T) {
	for _, kind := range []string{"fixed_order", "bm25", "lexical_ordered", "narrow_rule"} {
		var stdout, diagnostic bytes.Buffer
		if err := run([]string{"--baseline", kind}, strings.NewReader(example()), &stdout, &diagnostic); err != nil {
			t.Fatal(err)
		}
		var result output
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Candidates) != 2 || result.Candidates[0].ID == result.Candidates[1].ID {
			t.Fatal("candidate loss")
		}
		if strings.Contains(stdout.String(), "preserve input order") || strings.Contains(stdout.String(), "original-example") {
			t.Fatal("raw input retained")
		}
	}
}

func TestInvalidFlagDoesNotEchoSuppliedValues(t *testing.T) {
	for _, args := range [][]string{{"--iterations=authored-private-marker"}, {"--authored-private-marker"}, {"--baseline", "authored-private-marker"}} {
		var out, diagnostic bytes.Buffer
		err := run(args, strings.NewReader(example()), &out, &diagnostic)
		if err == nil || strings.Contains(err.Error()+diagnostic.String()+out.String(), "authored-private-marker") {
			t.Fatal("supplied flag value escaped fixed diagnostics")
		}
	}
}

func TestStreamWireLimitAndCRLFFraming(t *testing.T) {
	full := example() + strings.Repeat(" ", shortclaim.MaxJSONBytes-len(example()))
	for _, framed := range []string{full, full + "\n", full + "\r\n"} {
		var out, diagnostic bytes.Buffer
		if err := run([]string{"--stream"}, strings.NewReader(framed), &out, &diagnostic); err != nil {
			t.Fatal(err)
		}
	}
	var out, diagnostic bytes.Buffer
	if err := run([]string{"--stream"}, strings.NewReader(full+"\r"), &out, &diagnostic); err == nil || out.Len() != 0 {
		t.Fatal("oversized unterminated CR record accepted")
	}
}
