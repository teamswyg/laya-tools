// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

func number(v int) *int { return &v }
func ownedRows() []rawRow {
	var rows []rawRow
	for i, family := range []string{"owned-fit", "owned-dev-a", "owned-dev-b"} {
		for _, locale := range []string{"ko", "en"} {
			text := "직접 작성한 숫자시험 문장입니다. 의미 정답이나 품질 자료가 아닙니다."
			if locale == "en" {
				text = "Original owned numeric fixture; not human semantic truth or quality evidence."
			}
			r := rawRow{Schema: rowSchema, ID: family + "-" + locale, Family: family, Group: "owned-fit-group", Split: "fit", Locale: locale, Text: text, Targets: []string{"true", "false", "unknown"}, RubricSHA: strings.Repeat("a", 64), AnnotationSource: "ai_semantic_reference"}
			if i > 0 {
				r.Group, r.Split = "owned-dev-group", "dev"
			}
			for h := range r.Targets {
				end := len(text)
				if r.Targets[h] == "false" {
					end = 0
				}
				r.Evidence = append(r.Evidence, evidenceInput{number(0), number(end), "Owned numeric annotation fixture, no human reference quality claim."})
			}
			rows = append(rows, r)
		}
	}
	return rows
}
func encode(t *testing.T, rows []rawRow) []byte {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if err := json.NewEncoder(&b).Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}
func writeOwned(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(name, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestWholeLineageSplitAndPairedLocaleValidation(t *testing.T) {
	p, err := prepare(encode(t, ownedRows()))
	if err != nil || p.counts.Rows != 6 || p.counts.Families != 3 || p.counts.Lineages != 2 || p.counts.FitRows != 2 || p.counts.DevRows != 4 {
		t.Fatal("owned paired data", p.counts, err)
	}
	for _, change := range []func([]rawRow){
		func(r []rawRow) { r[2].Group, r[3].Group = r[0].Group, r[0].Group },
		func(r []rawRow) { r[1].Locale, r[1].ID = "ko", r[0].ID },
		func(r []rawRow) { r[2].ID = r[0].ID },
		func(r []rawRow) { r[1].Group = "different-group" },
		func(r []rawRow) { r[1].Split = "dev" },
		func(r []rawRow) { r[1].RubricSHA = strings.Repeat("b", 64) },
	} {
		rows := ownedRows()
		change(rows)
		if _, err := prepare(encode(t, rows)); err == nil {
			t.Fatal("split/locale/identity/rubric conflict accepted")
		}
	}
	// Locale-specific annotations remain independent rather than silently forced.
	rows := ownedRows()
	rows[1].Targets[2] = "true"
	if _, err := prepare(encode(t, rows)); err != nil {
		t.Fatal("distinct valid locale targets were forced together", err)
	}
}

func TestRequiredTargetsEvidenceAndUTF8SpanBoundaries(t *testing.T) {
	for _, change := range []func([]rawRow){
		func(r []rawRow) { r[0].Targets = nil },
		func(r []rawRow) { r[0].Targets = []string{"true", "false"} },
		func(r []rawRow) { r[0].Targets[0] = "planned" },
		func(r []rawRow) { r[0].Evidence = nil },
		func(r []rawRow) { r[0].Evidence[0].Start = nil },
		func(r []rawRow) { r[0].Evidence[0].End = nil },
		func(r []rawRow) { r[0].Evidence[0].Start = number(1) },
		func(r []rawRow) { r[0].Evidence[0].End = number(len(r[0].Text) + 1) },
		func(r []rawRow) { r[0].Evidence[2].End = number(0) },
		func(r []rawRow) { r[0].Evidence[1].Start, r[0].Evidence[1].End = number(3), number(3) },
		func(r []rawRow) { r[0].Evidence[0].Reason = "" },
		func(r []rawRow) { r[0].Evidence[0].Reason = strings.Repeat("x", 513) },
		func(r []rawRow) { r[0].Text = strings.Repeat("x", 4097) },
	} {
		rows := ownedRows()
		change(rows)
		if _, err := prepare(encode(t, rows)); err == nil {
			t.Fatal("missing/malformed target or evidence accepted")
		}
	}
	data := encode(t, ownedRows())
	missing := bytes.Replace(data, []byte(`"targets":["true","false","unknown"],`), nil, 1)
	if bytes.Equal(missing, data) {
		t.Fatal("owned missing-field mutation failed")
	}
	for _, bad := range [][]byte{missing, append([]byte{0xff}, data...), bytes.Repeat([]byte{'x'}, inputBudget+1)} {
		if _, err := prepare(bad); err == nil {
			t.Fatal("missing field, invalid UTF8 or budget accepted")
		}
	}
}

func TestPinnedCheckCreatesNoOutputOrProfile(t *testing.T) {
	t.Chdir(t.TempDir())
	data := encode(t, ownedRows())
	writeOwned(t, "owned.jsonl", data)
	args := []string{"--input", "owned.jsonl", "--input-sha256", digest(data), "--check"}
	var output, errorOutput bytes.Buffer
	if err := run(args, &output, &errorOutput); err != nil || !strings.Contains(output.String(), "no Fit, model predictions or outputs") {
		t.Fatal("owned check", err)
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("check created cache/model/profile")
	}
	for _, bad := range [][]string{
		{"--input", "owned.jsonl", "--input-sha256", strings.Repeat("0", 64), "--check"},
		append(append([]string(nil), args...), "--cpu-profile"),
		append(append([]string(nil), args...), "--out", anchor+"/bad"),
		{"--model", "private-marker"},
		{"--epochs", "1"},
	} {
		var out, errOut bytes.Buffer
		err := run(bad, &out, &errOut)
		if err == nil || out.Len() != 0 || strings.Contains(errOut.String()+err.Error(), "private-marker") {
			t.Fatal("bad pin/operation accepted or details leaked")
		}
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("invalid check wrote output")
	}
	if err := os.Symlink("owned.jsonl", "link.jsonl"); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := readInput(root, "link.jsonl", digest(data)); err == nil {
		t.Fatal("pinned symlink accepted")
	}
}

func readReport(t *testing.T, path string) studyReport {
	t.Helper()
	data, err := os.ReadFile(path)
	var r studyReport
	if err != nil || json.Unmarshal(data, &r) != nil {
		t.Fatal("owned report", err)
	}
	return r
}
func TestOriginalSyntheticFullRunIsPrivateNotQualifiedAndDevLabelsCannotTrain(t *testing.T) {
	t.Chdir(t.TempDir())
	rows := ownedRows()
	data := encode(t, rows)
	writeOwned(t, "owned-a.jsonl", data)
	var output, errOut bytes.Buffer
	args := []string{"--input", "owned-a.jsonl", "--input-sha256", digest(data), "--out", anchor + "/owned-a", "--cpu-profile"}
	if err := run(args, &output, &errOut); err != nil {
		t.Fatal("owned full run", err)
	}
	a := readReport(t, anchor+"/owned-a/report.json")
	if a.AnnotationSource != "ai_semantic_reference" || a.Qualified || a.Selection || a.Calibration || a.StateWrites != 0 || !a.ReloadExact || !a.CPUProfile || a.FitCalls != 1 || a.Fit.Samples != 2 || a.Fit.TrainingSteps != 40 || a.DevCalls != 4 || a.ReloadCalls != 4 || a.ArtifactBytes != statehintclaims.ArtifactBytes || a.Recipe != fixedRecipe() {
		t.Fatal("synthetic pipeline became quality proof or recipe changed", a)
	}
	for _, l := range a.Dev.Locales {
		if l.Rows != 2 || l.TrueTargetFamilies[0] != 2 || l.TrueTargetLineages[0] != 1 {
			t.Fatal("row/family/lineage support was conflated", l)
		}
		for h := range l.RawConfusion {
			n := 0
			for _, target := range l.RawConfusion[h] {
				for _, count := range target {
					n += count
				}
			}
			if n != 2 || l.TrueProposed[h] == 0 && l.PrecisionDefined[h] {
				t.Fatal("confusion denominator or undefined precision", l)
			}
		}
	}
	for _, name := range []string{"claims.rsc", "report.json", "cpu.pprof"} {
		st, err := os.Stat(anchor + "/owned-a/" + name)
		if err != nil || st.Mode().Perm() != 0600 || st.Size() == 0 {
			t.Fatal("output/profile is not private", name, err)
		}
	}
	st, err := os.Stat(anchor + "/owned-a")
	if err != nil || st.Mode().Perm() != 0700 {
		t.Fatal("run directory is not private")
	}
	var rawReport []byte
	rawReport, err = os.ReadFile(anchor + "/owned-a/report.json")
	if err != nil || bytes.Contains(rawReport, []byte(rows[0].Text)) || bytes.Contains(rawReport, []byte(rows[0].ID)) || bytes.Contains(rawReport, []byte(rows[0].Group)) {
		t.Fatal("raw text/identity entered aggregate report")
	}
	// Change only development targets, preserving positive evidence spans. The
	// trained artifact must stay identical even though dev metrics may change.
	for i := 2; i < len(rows); i++ {
		rows[i].Targets[2] = "true"
	}
	changed := encode(t, rows)
	writeOwned(t, "owned-b.jsonl", changed)
	output.Reset()
	if err := run([]string{"--input", "owned-b.jsonl", "--input-sha256", digest(changed), "--out", anchor + "/owned-b"}, &output, &errOut); err != nil {
		t.Fatal(err)
	}
	b := readReport(t, anchor+"/owned-b/report.json")
	if a.ModelSHA != b.ModelSHA || a.Fit != b.Fit || a.InputSHA == b.InputSHA {
		t.Fatal("dev annotation changed trained weights or input pin")
	}
	before, err := os.ReadFile(filepath.Join(anchor, "owned-a", "claims.rsc"))
	if err != nil {
		t.Fatal(err)
	}
	if err := run(args, &output, &errOut); err == nil {
		t.Fatal("existing private run overwritten")
	}
	after, err := os.ReadFile(filepath.Join(anchor, "owned-a", "claims.rsc"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed overwrite changed existing artifact")
	}
}
