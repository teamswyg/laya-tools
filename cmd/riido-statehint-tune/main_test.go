package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintpilot"
)

func TestStrictAuthoredRowRejectsAmbiguity(t *testing.T) {
	valid := `{"id":"toy1","family":"family1","locale":"en","text":"Where is the setting?","expected_intent":"question"}`
	if _, e := strictRow([]byte(valid)); e != nil {
		t.Fatal(e)
	}
	for _, s := range []string{valid + ` {}`, `{"id":"a","id":"b","family":"f","locale":"en","text":"x","expected_intent":"question"}`, `{"id":"a","family":"f","locale":"en","text":null,"expected_intent":"question"}`, `{"id":"a","family":"f","locale":"en","text":"x","Expected_intent":"question"}`} {
		if _, e := strictRow([]byte(s)); e == nil {
			t.Fatal("ambiguous row accepted")
		}
	}
}
func TestDisjointRejectsSharedContextEvenWithDifferentID(t *testing.T) {
	a := []statehintpilot.Case{{ID: "a", Family: "shared-context", Text: "first message", Expected: statehint.Progress}}
	b := []statehintpilot.Case{{ID: "b", Family: "shared-context", Text: "different message", Expected: statehint.CompletionReport}}
	if e := disjoint(a, b); e == nil {
		t.Fatal("context leaked across partitions")
	}
}
func TestPinnedInputAndActualSavedArtifact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "row.jsonl")
	row := statehintpilot.Case{ID: "toy", Family: "family", Locale: "en", Text: "Where is the setting?", Expected: statehint.Question}
	b, e := json.Marshal(row)
	if e != nil {
		t.Fatal(e)
	}
	b = append(b, '\n')
	if e = os.WriteFile(path, b, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = readPin(Pin{Path: path, SHA256: digest(b), Rows: 1}); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, append(b, ' '), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = readPin(Pin{Path: path, SHA256: digest(b), Rows: 1}); e == nil {
		t.Fatal("changed source file accepted")
	}
	payload := []byte("retained-first-artifact")
	out := filepath.Join(dir, "artifact")
	if e = writeBytes(out, payload); e != nil {
		t.Fatal(e)
	}
	if e = writeBytes(out, []byte("replacement")); e == nil {
		t.Fatal("artifact overwritten")
	}
	saved, e := os.ReadFile(out)
	if e != nil || digest(saved) != digest(payload) {
		t.Fatal("saved original not preserved")
	}
}
