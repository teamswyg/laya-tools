// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// Original temporary metadata fixtures exercise IO and actual fitting, not
// learned model quality or independently verified semantic labels.
func fixtureRows(partition string, families int) []statehintcorpus.Row {
	rows := make([]statehintcorpus.Row, 0, 2*families)
	for i := 0; i < families; i++ {
		label := statehint.Intents()[i%8]
		for _, locale := range []string{"ko", "en"} {
			family := fmt.Sprintf("%s-fixture-%04d", partition, i)
			rows = append(rows, statehintcorpus.Row{
				Schema: "statehint-v4-original-train-seed-row-v1", ID: family + "-" + locale,
				Family: family, Lineage: family, Partition: partition, Locale: locale,
				Wording: 1, Role: "prose", Applicable: true,
				Text: fmt.Sprintf("Original temporary %s arithmetic observation %d %s %s", partition, i, locale, label),
				Unit: "Temporary arithmetic fixture, not a semantic benchmark.", ClauseScopes: []string{"current_unit"},
				AssertionForms: []string{"asserted"}, Expected: label, Ambiguous: label == statehint.Unclear,
				AnnotationSource: "Original owned test metadata; no human truth claim.",
				OntologyVersion:  "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: rubricSHA, License: "Apache-2.0",
			})
		}
	}
	return rows
}

func encodeRows(t *testing.T, rows []statehintcorpus.Row) []byte {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if err := json.NewEncoder(&b).Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}

func fixtureInput(t *testing.T, partition string) input {
	t.Helper()
	b := encodeRows(t, fixtureRows(partition, 8))
	c, s, err := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if err != nil {
		t.Fatal(err)
	}
	return input{pin{partition + ".jsonl", digest(b)}, b, c, s}
}

func TestPinnedInputBoundsChecksumAndPartition(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	b := encodeRows(t, fixtureRows("train", 8))
	if err := os.WriteFile(filepath.Join(dir, "train.jsonl"), b, 0600); err != nil {
		t.Fatal(err)
	}
	p := pin{"train.jsonl", digest(b)}
	if got, err := readPin(root, p, int64(len(b))); err != nil || !bytes.Equal(got, b) {
		t.Fatal("exact bound failed", err)
	}
	if _, err := readPin(root, p, int64(len(b)-1)); err == nil {
		t.Fatal("over-budget bytes accepted despite correct checksum")
	}
	for _, bad := range []string{strings.Repeat("0", 64), strings.ToUpper(p.SHA), "not-a-checksum"} {
		if _, err := loadInput(root, pin{p.Path, bad}, "train", 8); err == nil {
			t.Fatal("bad input checksum accepted")
		}
	}
	if _, err := loadInput(root, p, "validation", 8); err == nil {
		t.Fatal("wrong partition accepted")
	}
	if _, err := loadInput(root, p, "train", 840); err == nil {
		t.Fatal("tiny corpus bypassed production quota")
	}
	rows := fixtureRows("train", 8)
	rows[0].Text = strings.Repeat("x", 4097)
	b = encodeRows(t, rows)
	if err := os.WriteFile(filepath.Join(dir, p.Path), b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadInput(root, pin{p.Path, digest(b)}, "train", 8); err == nil {
		t.Fatal("oversized text bypassed corpus reader")
	}
}

func TestCrossInputLineageAndExactTextReuseRejected(t *testing.T) {
	a, b := fixtureInput(t, "train"), fixtureInput(t, "validation")
	if err := disjoint(a, b); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func([]statehintcorpus.Row){
		func(rows []statehintcorpus.Row) {
			rows[0].Lineage, rows[1].Lineage = a.corpus.Rows()[0].Lineage, a.corpus.Rows()[0].Lineage
		},
		func(rows []statehintcorpus.Row) { rows[0].Text = a.corpus.Rows()[0].Text },
	} {
		rows := fixtureRows("validation", 8)
		mutate(rows)
		data := encodeRows(t, rows)
		c, s, err := statehintcorpus.Read(bytes.NewReader(data), statehintcorpus.Options{Partition: "validation", RubricSHA: rubricSHA})
		if err != nil {
			t.Fatal("fixture failed before cross-input check", err)
		}
		if err := disjoint(a, input{corpus: c, summary: s}); err == nil {
			t.Fatal("cross-input source reuse accepted")
		}
	}
}

func TestActualBothModeFitsPreserveBytesScoresAndReload(t *testing.T) {
	dir := t.TempDir()
	out, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	a, b := fixtureInput(t, "train"), fixtureInput(t, "validation")
	r, err := fitAll(out, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Arms) != 2 || !r.DevelopmentOnly || !r.ValidationExposed || r.CalibrationCalls != 0 || r.TestCalls != 0 || r.Selected || r.Promoted || r.Qualified {
		t.Fatal("development boundary changed")
	}
	for name, want := range map[string][]byte{"train.source.jsonl": a.bytes, "validation.source.jsonl": b.bytes} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatal("source bytes changed", err)
		}
	}
	for _, arm := range r.Arms {
		if !arm.ReloadParity || arm.Fit.TrainingSteps != 40 || arm.ValidationCalls != 32 || arm.ArtifactBytes != statehintwide.ArtifactBytes || arm.ProbabilityOrder != statehint.Intents() {
			t.Fatal("trained mode record incomplete")
		}
		artifact, err := os.ReadFile(filepath.Join(dir, arm.Model.Path))
		if err != nil || digest(artifact) != arm.Model.SHA {
			t.Fatal("saved model pin mismatch", err)
		}
		loaded, err := statehintwide.Load(bytes.NewReader(artifact))
		if err != nil || loaded.Temperature() != 1 || loaded.TrainingSteps() != 40 {
			t.Fatal("trained artifact lost state", err)
		}
		var saved bytes.Buffer
		if err := loaded.Save(&saved); err != nil || !bytes.Equal(saved.Bytes(), artifact) {
			t.Fatal("trained artifact serialization drift", err)
		}
		p, err := predict(loaded, b.corpus)
		if err != nil {
			t.Fatal(err)
		}
		metrics, err := statehintfamily.Evaluate(b.corpus, p)
		if err != nil || !reflect.DeepEqual(metrics, arm.Validation) {
			t.Fatal("reloaded family report drift", err)
		}
		data, err := os.ReadFile(filepath.Join(dir, arm.Mode+".predictions.json"))
		if err != nil {
			t.Fatal(err)
		}
		var scored []scoredRow
		if err := json.Unmarshal(data, &scored); err != nil || len(scored) != len(p) {
			t.Fatal("complete score vectors missing", err)
		}
		for i := range p {
			if scored[i].Prediction != p[i] || scored[i].ID != b.corpus.Rows()[i].ID {
				t.Fatal("source-order eight-column scores changed")
			}
		}
		corrupt := append([]byte(nil), artifact...)
		corrupt[len(corrupt)-1] ^= 1
		for _, bad := range [][]byte{artifact[:len(artifact)-1], append(append([]byte(nil), artifact...), 0), corrupt} {
			if _, err := statehintwide.Load(bytes.NewReader(bad)); err == nil {
				t.Fatal("invalid trained artifact accepted")
			}
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		st, err := entry.Info()
		if err != nil || st.Mode().Perm() != 0600 {
			t.Fatal("output is not private", err)
		}
	}
}

func TestPrivateOutputExclusiveAndNoImplicitFitting(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	out, err := privateOutput(root, anchor+"/run01")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeBytes(out, "kept", []byte("original")); err != nil {
		t.Fatal(err)
	}
	if err := writeBytes(out, "kept", []byte("replacement")); err == nil {
		t.Fatal("overwrote an existing artifact")
	}
	out.Close()
	if _, err := privateOutput(root, anchor+"/run01"); err == nil {
		t.Fatal("existing run accepted")
	}
	st, err := root.Stat(anchor + "/run01")
	if err != nil || st.Mode().Perm() != 0700 {
		t.Fatal("run directory is not private", err)
	}
	for _, args := range [][]string{nil, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--learning-rate", ".1"}} {
		var stdout, stderr bytes.Buffer
		err := run(args, &stdout, &stderr)
		if err == nil || strings.Contains(stdout.String()+stderr.String()+err.Error(), "private-marker") {
			t.Fatal("implicit fitting, unsupported inputs or option echo")
		}
	}
}

func TestCheckCLIUsesFullQuotasWithoutOutputs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	train := encodeRows(t, fixtureRows("train", 840))
	validation := encodeRows(t, fixtureRows("validation", 120))
	for name, data := range map[string][]byte{"train.jsonl": train, "validation.jsonl": validation} {
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"--train", "train.jsonl", "--train-sha256", digest(train), "--validation", "validation.jsonl", "--validation-sha256", digest(validation), "--check"}
	var stdout, stderr bytes.Buffer
	if err := run(args, &stdout, &stderr); err != nil || !strings.Contains(stdout.String(), "no fitting or model calls") {
		t.Fatal("full-quota check failed", err)
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("check created output state")
	}
	for _, index := range []int{3, 7} {
		bad := append([]string(nil), args...)
		bad[index] = strings.Repeat("0", 64)
		stdout.Reset()
		if err := run(bad, &stdout, &stderr); err == nil || stdout.Len() != 0 {
			t.Fatal("bad train/validation pin yielded a success report")
		}
	}
}

func TestValidationTargetsAndAnnotationsCannotChangeTrainedWeights(t *testing.T) {
	train, validation := fixtureInput(t, "train"), fixtureInput(t, "validation")
	changed := validation.corpus.Rows()
	for i := range changed {
		index, _ := statehint.IntentIndex(changed[i].Expected)
		changed[i].Expected = statehint.Intents()[(index+1)%8]
		changed[i].Ambiguous = changed[i].Expected == statehint.Unclear
		changed[i].Unit = "Changed annotation; never a model feature."
	}
	c, _, err := statehintcorpus.Read(bytes.NewReader(encodeRows(t, changed)), statehintcorpus.Options{Partition: "validation", RubricSHA: rubricSHA})
	if err != nil {
		t.Fatal(err)
	}
	samples := make([]statehintwide.Sample, len(train.corpus.Rows()))
	for i, row := range train.corpus.Rows() {
		samples[i] = statehintwide.Sample{Text: row.Text, Label: row.Expected}
	}
	var hashes [2]string
	for i, corpus := range []statehintcorpus.Corpus{validation.corpus, c} {
		out, err := os.OpenRoot(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		arm, err := fitArm(out, statehintwide.Contextual, "contextual2048", samples, corpus)
		out.Close()
		if err != nil {
			t.Fatal(err)
		}
		hashes[i] = arm.Model.SHA
	}
	if hashes[0] != hashes[1] {
		t.Fatal("validation labels or annotations changed training")
	}
}
