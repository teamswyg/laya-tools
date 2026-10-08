// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayDefaultsAndCheckPreserveRecordedResults(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	original := "experiments/claim-position-features-preview-v1"
	bundle := t.TempDir()
	if err := os.Mkdir(filepath.Join(bundle, "replay"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cases.json", "DESIGN.lock.json", "results.json", "replay/main.go"} {
		raw, err := os.ReadFile(filepath.Join(original, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(bundle, name), raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	recordedPath := filepath.Join(bundle, "results.json")
	recorded, err := os.ReadFile(recordedPath)
	if err != nil {
		t.Fatal(err)
	}
	assertUnchanged := func() {
		t.Helper()
		actual, err := os.ReadFile(recordedPath)
		if err != nil || !bytes.Equal(actual, recorded) {
			t.Fatal("replay overwrote the recorded evidence", err)
		}
	}
	var stdout bytes.Buffer
	if err := runCLI([]string{"--bundle", bundle}, &stdout); err != nil || !json.Valid(stdout.Bytes()) {
		t.Fatal("default replay must emit JSON to stdout", err)
	}
	assertUnchanged()
	stdout.Reset()
	if err := runCLI([]string{"--bundle", bundle, "--output", "-"}, &stdout); err != nil || !json.Valid(stdout.Bytes()) {
		t.Fatal("explicit stdout output failed", err)
	}
	assertUnchanged()
	stdout.Reset()
	if err := runCLI([]string{"--bundle", bundle, "--check"}, &stdout); err != nil || stdout.Len() != 0 {
		t.Fatal("read-only check failed or wrote output", err)
	}
	assertUnchanged()
	optInPath := filepath.Join(bundle, "opt-in.json")
	if err := runCLI([]string{"--bundle", bundle, "--check", "--output", optInPath}, &stdout); err == nil {
		t.Fatal("ambiguous check/output combination accepted")
	}
	if _, err := os.Stat(optInPath); !os.IsNotExist(err) {
		t.Fatal("check created an output file", err)
	}
	assertUnchanged()
	if err := runCLI([]string{"--bundle", bundle, "--output", optInPath}, &stdout); err != nil {
		t.Fatal("explicit output opt-in failed", err)
	}
	if raw, err := os.ReadFile(optInPath); err != nil || !json.Valid(raw) {
		t.Fatal("explicit output did not create result JSON", err)
	}
	assertUnchanged()
	var altered Results
	if err := json.Unmarshal(recorded, &altered); err != nil {
		t.Fatal(err)
	}
	altered.Pairs[0].A.PositionSHA256 = "altered feature evidence"
	recorded, err = json.MarshalIndent(altered, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(recordedPath, recorded, 0644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := runCLI([]string{"--bundle", bundle, "--check"}, &stdout); err == nil || stdout.Len() != 0 {
		t.Fatal("check did not reject altered evidence without output")
	}
	assertUnchanged()
}
