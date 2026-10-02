// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIFlagsNeverEchoSuppliedArguments(t *testing.T) {
	secret := "private-value-marker"
	for _, args := range [][]string{
		{"--unexpected=" + secret},
		{"--input-root=" + secret, "--plan-sha256=" + secret},
		{"--plan-sha256"},
		{"--input-root", secret, "--output", secret, "--plan-sha256", secret},
		{"--input-root", secret, "--output", secret, "--plan-sha256", planSHA, secret},
	} {
		var out, err bytes.Buffer
		code := run(args, &out, &err)
		if code != 2 || out.Len() != 0 || strings.Contains(err.String(), secret) || !strings.HasPrefix(err.String(), "scope67_") {
			t.Fatal("CLI value leaked or flag error not fixed")
		}
	}
}

func TestCLIHelpHasNoSourceOrModelSideEffect(t *testing.T) {
	var out, err bytes.Buffer
	if code := run([]string{"--help"}, &out, &err); code != 0 || out.String() != usage || err.Len() != 0 {
		t.Fatal("fixed help differs")
	}
}

func TestCLIToyEnvelopeAndExclusiveReservation(t *testing.T) {
	f := fixture67(t)
	result := filepath.Join(t.TempDir(), "toy-result.json")
	args := []string{"--input-root", f.root, "--output", result, "--plan-sha256", planSHA}
	var out, err bytes.Buffer
	if code := runWithConfig(args, &out, &err, f.cfg); code != 0 || err.Len() != 0 {
		t.Fatal("toy CLI failed", err.String())
	}
	b, e := os.ReadFile(result)
	if e != nil {
		t.Fatal(e)
	}
	var env envelope
	if e := json.Unmarshal(b, &env); e != nil {
		t.Fatal(e)
	}
	if env.State != "passed_metadata_proposed_supervision_only" || env.Failure != "" || len(env.Report.Parents) != 3 || env.Report.TrainingReady || env.Report.DiversityCleared || env.Ledger.ReviewRecordsVerified != 40 {
		t.Fatal("toy envelope semantics differ")
	}
	st, e := os.Stat(result)
	if e != nil || st.Mode().Perm() != 0600 {
		t.Fatal("result permissions differ")
	}
	out.Reset()
	err.Reset()
	if code := runWithConfig(args, &out, &err, f.cfg); code != 1 || out.Len() != 0 || err.String() != "scope67_output_reserve_failed\n" {
		t.Fatal("existing output reused")
	}
	after, e := os.ReadFile(result)
	if e != nil || !bytes.Equal(b, after) {
		t.Fatal("existing output overwritten")
	}
}

func TestCLIToyFailurePersistsPrefixWithoutRawPaths(t *testing.T) {
	f := fixture67(t)
	f.cfg.Pins[4].SHA = strings.Repeat("e", 64)
	result := filepath.Join(t.TempDir(), "toy-failed-result.json")
	var out, err bytes.Buffer
	code := runWithConfig([]string{"--input-root", f.root, "--output", result, "--plan-sha256", planSHA}, &out, &err, f.cfg)
	if code != 1 || out.Len() != 0 || err.String() != "scope67_pin\n" || strings.Contains(err.String(), f.root) {
		t.Fatal("failure diagnostic leaked values")
	}
	b, e := os.ReadFile(result)
	if e != nil {
		t.Fatal(e)
	}
	var env envelope
	if json.Unmarshal(b, &env) != nil || env.State != "failed" || len(env.Report.Parents) != 0 || env.Ledger.FilesVerified != 4 || env.Ledger.BytesRead <= env.Ledger.BytesVerified {
		t.Fatal("failure prefix missing")
	}
	if strings.Contains(string(b), f.root) || strings.Contains(string(b), result) {
		t.Fatal("raw path in failure envelope")
	}
}
