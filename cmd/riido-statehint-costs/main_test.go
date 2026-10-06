// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

func TestBadPinAndUntrainedArtifactProduceNoProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "untrained.rsh")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := statehint.NewModel().Save(file); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	profiles := filepath.Join(dir, "profiles")
	var out bytes.Buffer
	if err := run([]string{"--model", path, "--model-sha256", strings.Repeat("0", 64), "--profiles-dir", profiles}, &out, &out); err == nil {
		t.Fatal("bad model pin accepted")
	}
	if out.Len() != 0 {
		t.Fatal("failed benchmark reported measurements")
	}
	if _, err := os.Stat(profiles); !os.IsNotExist(err) {
		t.Fatal("failure created profile output")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if err := run([]string{"--model", path, "--model-sha256", hex.EncodeToString(sum[:]), "--profiles-dir", profiles}, &out, &out); err == nil {
		t.Fatal("untrained artifact accepted as a trained inference benchmark")
	}
	if _, err := os.Stat(profiles); !os.IsNotExist(err) {
		t.Fatal("untrained model created profile output")
	}
}
