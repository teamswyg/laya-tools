package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/taskoutcome"
)

const fixture = "{\"type\":\"thread.started\"}\n{\"type\":\"turn.started\"}\n{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":1,\"cached_input_tokens\":0,\"output_tokens\":0}}\n"

func TestCLIStdinAggregateOnly(t *testing.T) {
	var out, diagnostic bytes.Buffer
	if code := run([]string{"--input", "-", "--requested-model", "example-model"}, strings.NewReader(fixture), &out, &diagnostic); code != 0 || diagnostic.Len() != 0 {
		t.Fatal(code, diagnostic.String())
	}
	var s taskoutcome.Summary
	if err := json.Unmarshal(out.Bytes(), &s); err != nil || !s.UsageComplete || s.Metadata.RequestedModel != "example-model" || s.ObservedModel != "unknown" || s.TaskAcceptance != "unknown" {
		t.Fatal("invalid aggregate response", err)
	}
}

func TestCLIRedactedFailuresAndNoApprovalOption(t *testing.T) {
	for _, args := range [][]string{
		{"--input", "/Users/private/nonexistent-trace-secret"},
		{"--requested-model", "ghp_private_secret"},
		{"--verified-acceptance", "true"},
		{"--unknown=/Users/private/secret"},
	} {
		var out, diagnostic bytes.Buffer
		code := run(args, strings.NewReader(fixture), &out, &diagnostic)
		if code == 0 || out.Len() != 0 || strings.Contains(diagnostic.String(), "/Users/") || strings.Contains(diagnostic.String(), "secret") {
			t.Fatal("failed CLI echoed private arguments or authorized acceptance", code)
		}
	}
	var out, diagnostic bytes.Buffer
	if code := run(nil, strings.NewReader(`{"type":"error","message":`), &out, &diagnostic); code == 0 || out.Len() != 0 || diagnostic.String() != "invalid_event_json\n" {
		t.Fatal("truncated input became a successful report")
	}
}

func TestCLIFileAndHelp(t *testing.T) {
	f := filepath.Join(t.TempDir(), "private-path.jsonl")
	if err := os.WriteFile(f, []byte(fixture), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostic bytes.Buffer
	if code := run([]string{"--input", f}, nil, &out, &diagnostic); code != 0 || strings.Contains(out.String(), f) {
		t.Fatal("file input failed or echoed path")
	}
	out.Reset()
	if code := run([]string{"--help"}, nil, &out, &diagnostic); code != 0 || !strings.Contains(out.String(), "No Codex execution") {
		t.Fatal("missing help")
	}
	out.Reset()
	diagnostic.Reset()
	if code := run([]string{"--input", t.TempDir()}, nil, &out, &diagnostic); code == 0 || out.Len() != 0 || diagnostic.String() != "input_not_regular\n" {
		t.Fatal("nonregular input accepted or echoed")
	}
}
