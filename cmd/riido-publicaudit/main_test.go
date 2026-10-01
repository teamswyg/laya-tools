package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/publicbehavior"
)

func TestCanonicalRejectsDuplicateUnknownTrailingAndLargeIntegerLoss(t *testing.T) {
	type value struct {
		Number uint64 `json:"number"`
	}
	valid := []byte("{\n  \"number\": 18446744073709551615\n}\n")
	var v value
	if err := decodeCanonical(valid, &v); err != nil || v.Number != ^uint64(0) {
		t.Fatal("exact uint64 was lost", err)
	}
	for _, raw := range []string{
		"{\n  \"number\": 1,\n  \"number\": 18446744073709551615\n}\n",
		"{\n  \"number\": 18446744073709551615,\n  \"unknown\": true\n}\n",
		string(valid) + "{}\n",
		"{\n  \"number\": 18446744073709551616\n}\n",
	} {
		if err := decodeCanonical([]byte(raw), &v); err == nil {
			t.Fatal("accepted invalid canonical data")
		}
	}
}

func TestOutputIsExclusiveAndPreflightRejectsUnlistedSource(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "receipt")
	if err := writeNew(file, []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := writeNew(file, []byte("replacement")); err == nil {
		t.Fatal("replaced an existing receipt")
	}
	got, err := os.ReadFile(file)
	if err != nil || string(got) != "original" {
		t.Fatal("original receipt changed")
	}
	for _, path := range []string{"cmd/riido-publicaudit", "internal/publicbehavior", "internal/publicbehavior/testdata/upstream/semver", "internal/publicbehavior/testdata/upstream/doublestar"} {
		if err := os.MkdirAll(filepath.Join(dir, path), 0700); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := verifyGoClosure(root, nil); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/publicbehavior/extra.go"), []byte("package publicbehavior\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyGoClosure(root, nil); err == nil {
		t.Fatal("unlisted compiled source accepted")
	}
}

func TestFrozenProbePreflightDoesNotObserve(t *testing.T) {
	raw, err := os.ReadFile("../../" + probesPath)
	if err != nil {
		t.Fatal(err)
	}
	var p publicbehavior.Probes
	if err := decodeCanonical(raw, &p); err != nil {
		t.Fatal(err)
	}
	if err := publicbehavior.ValidateProbes(p); err != nil {
		t.Fatal(err)
	}
	if len(p.Vectors) != 62 {
		t.Fatal("input count drift")
	}
	original := p.Vectors[0].LeftHex
	p.Vectors[0].LeftHex = "ff"
	if publicbehavior.ValidateProbes(p) == nil {
		t.Fatal("raw input/hex mismatch accepted")
	}
	p.Vectors[0].LeftHex = original
	p.Vectors[0].Checks = []publicbehavior.Check{{Field: "supported", Want: json.RawMessage("true")}}
	if publicbehavior.ValidateProbes(p) == nil {
		t.Fatal("vacuous acceptance was permitted")
	}
	if valid := hash(strings.Repeat("A", 40), 40); valid {
		t.Fatal("noncanonical pin accepted")
	}
}
