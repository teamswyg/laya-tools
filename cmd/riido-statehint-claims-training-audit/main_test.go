// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureRows() []inputRow {
	rows := make([]inputRow, 6)
	for i := range rows {
		locale := []string{"ko", "en"}[i%2]
		family := fmt.Sprintf("fixture-family-%d", i/2)
		rows[i] = inputRow{Schema: "riido-three-claims-training-row-v1", ID: family + "-" + locale, Family: family, Group: "fixture-group", Split: "fit", Locale: locale, Text: fmt.Sprintf("Private fixture text number %d", i), Targets: []string{"true", "false", "unknown"}, Rubric: rubricSHA, Annotation: "ai_semantic_reference"}
		setEvidence(&rows[i])
	}
	return rows
}
func setEvidence(r *inputRow) {
	r.Evidence = make([]evidence, 3)
	for h, target := range r.Targets {
		start, end := 0, len(r.Text)
		if target == "false" {
			end = 0
		}
		r.Evidence[h] = evidence{&start, &end, "fixture explanation"}
	}
}
func corpus(t *testing.T, rows []inputRow) ([]byte, manifest) {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if e := json.NewEncoder(&b).Encode(r); e != nil {
			t.Fatal(e)
		}
	}
	data := b.Bytes()
	return data, manifest{digest(data), len(rows), len(rows) / 2, len(rows) / 6}
}
func auditFixture(t *testing.T, rows []inputRow) report {
	t.Helper()
	b, pin := corpus(t, rows)
	parsed, e := prepare(b, pin)
	if e != nil {
		t.Fatal(e)
	}
	r, e := summarize(parsed, pin)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func fixtureRoot(t *testing.T, b []byte) *os.Root {
	t.Helper()
	dir := t.TempDir()
	if e := os.WriteFile(filepath.Join(dir, "training.jsonl"), b, 0600); e != nil {
		t.Fatal(e)
	}
	r, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

func TestFixedCLIManifestAndDigestBeforeDecode(t *testing.T) {
	b, fixture := corpus(t, fixtureRows())
	root := fixtureRoot(t, b)
	var out bytes.Buffer
	if e := run([]string{"--training", "training.jsonl", "--training-sha256", fixture.sha}, &out, root, frozenManifest()); e != errAudit || out.Len() != 0 {
		t.Fatal("external CLI accepted an arbitrary fixture digest")
	}
	if e := run([]string{"--training", "missing.jsonl", "--training-sha256", "wrong"}, &out, nil, frozenManifest()); e != errAudit {
		t.Fatal("advertised digest was not rejected before file access")
	}
	wrong := fixture
	wrong.sha = trainingSHA
	if _, e := readPinned(root, "training.jsonl", trainingSHA, wrong); e != errAudit {
		t.Fatal("body digest mismatch accepted")
	}
	if _, e := prepare([]byte("not JSON"), wrong); e != errAudit {
		t.Fatal("unpinned body reached parsing")
	}
	if e := run([]string{"--training", "training.jsonl", "--training-sha256", fixture.sha}, &out, root, fixture); e != nil {
		t.Fatal(e)
	}
	if !json.Valid(out.Bytes()) {
		t.Fatal("fixture result is not JSON")
	}
	for _, unknown := range []string{"--evaluation", "--model", "--out", "--seed", "--fit"} {
		out.Reset()
		if e := run([]string{unknown, "x"}, &out, root, fixture); e != errAudit || out.Len() != 0 {
			t.Fatal("unsupported CLI option accepted", unknown)
		}
	}
}

func TestBoundedRegularInputAndAliases(t *testing.T) {
	b, pin := corpus(t, fixtureRows())
	root := fixtureRoot(t, b)
	for _, name := range []string{"", ".", "./training.jsonl", "x/../training.jsonl", "../training.jsonl", "/training.jsonl", "x\\training.jsonl", "training.jsonl/"} {
		if _, e := readPinned(root, name, pin.sha, pin); e != errAudit {
			t.Fatal("path alias accepted", name)
		}
	}
	if e := root.Mkdir("nested", 0700); e != nil {
		t.Fatal(e)
	}
	if e := root.WriteFile("nested/training.jsonl", b, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := readPinned(root, "nested/training.jsonl", pin.sha, pin); e != nil {
		t.Fatal("regular nested file rejected", e)
	}
	if e := root.Symlink("training.jsonl", "leaf-link"); e != nil {
		t.Fatal(e)
	}
	if _, e := readPinned(root, "leaf-link", pin.sha, pin); e != errAudit {
		t.Fatal("leaf symlink accepted")
	}
	if e := root.Symlink("nested", "dir-link"); e != nil {
		t.Fatal(e)
	}
	if _, e := readPinned(root, "dir-link/training.jsonl", pin.sha, pin); e != errAudit {
		t.Fatal("component symlink accepted")
	}
	if e := root.Link("training.jsonl", "hardlink"); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"training.jsonl", "hardlink"} {
		if _, e := readPinned(root, name, pin.sha, pin); e != errAudit {
			t.Fatal("hardlink alias accepted")
		}
	}
	if _, e := readPinned(root, "nested", pin.sha, pin); e != errAudit {
		t.Fatal("directory accepted")
	}
	if e := root.WriteFile("empty", nil, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := readPinned(root, "empty", pin.sha, pin); e != errAudit {
		t.Fatal("empty input accepted")
	}
	large := bytes.Repeat([]byte("x"), inputBudget+1)
	if e := root.WriteFile("large", large, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := readPinned(root, "large", pin.sha, pin); e != errAudit {
		t.Fatal("oversized input accepted")
	}
}

func TestStrictJSONRejectsAmbiguousOrMissingFields(t *testing.T) {
	b, _ := corpus(t, fixtureRows())
	line := bytes.Split(b, []byte("\n"))[0]
	variants := [][]byte{
		bytes.Replace(line, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1),
		bytes.Replace(line, []byte(`"start_byte":0`), []byte(`"start_byte":0,"start_byte":0`), 1),
		append(append([]byte(nil), line...), []byte(" {}")...),
		bytes.Replace(line, []byte(`"reason":`), []byte(`"extra":true,"reason":`), 1),
		bytes.Replace(line, []byte(`"start_byte":0,`), nil, 1),
		bytes.Replace(line, []byte(`"start_byte":0`), []byte(`"start_byte":null`), 1),
		bytes.Replace(line, []byte(`"schema":`), []byte(`"unexpected":1,"schema":`), 1),
	}
	for i, v := range variants {
		var r inputRow
		if decodeStrict(v, &r) != errAudit {
			t.Fatal("malformed JSON accepted", i)
		}
	}
}

func TestRowAndFamilyGroupInvariants(t *testing.T) {
	cases := []struct {
		name   string
		change func([]inputRow)
	}{
		{"split", func(r []inputRow) { r[0].Split = "dev" }},
		{"rubric", func(r []inputRow) { r[0].Rubric = strings.Repeat("0", 64) }},
		{"annotation", func(r []inputRow) { r[0].Annotation = "human" }},
		{"locale", func(r []inputRow) { r[0].Locale = "ja" }},
		{"label", func(r []inputRow) { r[0].Targets[0] = "invalid" }},
		{"targets", func(r []inputRow) { r[0].Targets = r[0].Targets[:2] }},
		{"evidence", func(r []inputRow) { r[0].Evidence = r[0].Evidence[:2] }},
		{"empty_true", func(r []inputRow) { x := 0; r[0].Evidence[0].End = &x }},
		{"negative_span", func(r []inputRow) { x := -1; r[0].Evidence[0].Start = &x }},
		{"beyond_span", func(r []inputRow) { x := len(r[0].Text) + 1; r[0].Evidence[0].End = &x }},
		{"empty_reason", func(r []inputRow) { r[0].Evidence[0].Reason = " " }},
		{"text_limit", func(r []inputRow) { r[0].Text = strings.Repeat("x", 4097); setEvidence(&r[0]) }},
		{"wordless", func(r []inputRow) { r[0].Text = "!!!"; setEvidence(&r[0]) }},
		{"duplicate_id", func(r []inputRow) { r[1] = r[0] }},
		{"unpaired_family", func(r []inputRow) { r[0].Family = "other"; r[0].ID = "other-ko" }},
		{"split_group", func(r []inputRow) { r[0].Group = "other" }},
		{"utf8_boundary", func(r []inputRow) { r[0].Text = "한글"; setEvidence(&r[0]); x := 1; r[0].Evidence[0].Start = &x }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows := fixtureRows()
			tc.change(rows)
			b, pin := corpus(t, rows)
			parsed, e := prepare(b, pin)
			if e == nil {
				_, e = summarize(parsed, pin)
			}
			if e != errAudit {
				t.Fatal("invalid row metadata accepted")
			}
		})
	}
	rows := fixtureRows()
	b, pin := corpus(t, rows)
	pin.rows++
	if _, e := prepare(b, pin); e != errAudit {
		t.Fatal("incorrect corpus shape accepted")
	}
}

func TestLengthBucketBoundariesAndFiniteFractions(t *testing.T) {
	rows := fixtureRows()
	for i, n := range []int{64, 65, 128, 129, 256, 257} {
		rows[i].Text = strings.Repeat("x", n)
		setEvidence(&rows[i])
	}
	r := auditFixture(t, rows)
	if r.LengthRows != [4]int{1, 2, 2, 1} {
		t.Fatal("byte boundary buckets", r.LengthRows)
	}
	for h := range 3 {
		for s := range 3 {
			f := r.SpanFractions[h][s]
			if math.IsNaN(f.Mean) || math.IsInf(f.Mean, 0) || f.Min < 0 || f.Max > 1 {
				t.Fatal("invalid fraction")
			}
		}
	}
	for bucket := range 4 {
		for h := range 3 {
			total := 0
			for l := range 2 {
				for s := range 3 {
					total += r.LengthStates[bucket][l][h][s]
				}
			}
			if total != r.LengthRows[bucket] {
				t.Fatal("bucket support inconsistent")
			}
		}
	}
	rows[5].Text = strings.Repeat("x", 4096)
	setEvidence(&rows[5])
	if r := auditFixture(t, rows); r.LengthRows != [4]int{1, 2, 2, 1} {
		t.Fatal("4096-byte text boundary rejected")
	}
}

func TestGroupMustContainExactlyThreeCompletePairs(t *testing.T) {
	rows := append(fixtureRows(), fixtureRows()...)
	for i := 6; i < len(rows); i++ {
		rows[i].Family += "-second"
		rows[i].ID = rows[i].Family + "-" + rows[i].Locale
	}
	b, pin := corpus(t, rows)
	parsed, e := prepare(b, pin)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := summarize(parsed, pin); e != errAudit {
		t.Fatal("twelve-row group accepted")
	}
	for i := 6; i < len(rows); i++ {
		rows[i].Group = "second-group"
	}
	if r := auditFixture(t, rows); r.Groups != 2 {
		t.Fatal("two valid groups rejected")
	}
}

func TestUnicodeRuneBoundariesAndLocaleByteDivergence(t *testing.T) {
	rows := append(fixtureRows(), fixtureRows()...)
	for i := 6; i < len(rows); i++ {
		rows[i].Family += "-second"
		rows[i].ID = rows[i].Family + "-" + rows[i].Locale
		rows[i].Group = "second-group"
	}
	for family, n := range []int{16, 17, 32, 33, 64, 65} {
		rows[family*2].Text = strings.Repeat("한", n)
		rows[family*2+1].Text = strings.Repeat("a", n)
		setEvidence(&rows[family*2])
		setEvidence(&rows[family*2+1])
	}
	r := auditFixture(t, rows)
	if r.RuneBounds != [4]int{16, 32, 64, 4096} || r.RuneRows != [4]int{2, 4, 4, 2} || r.LocaleRuneRows != [2][4]int{{1, 2, 2, 1}, {1, 2, 2, 1}} {
		t.Fatal("Unicode rune boundary counters", r.RuneRows, r.LocaleRuneRows)
	}
	if r.LocaleLengthRows != [2][4]int{{2, 2, 2, 0}, {5, 1, 0, 0}} || r.LengthRows != [4]int{7, 3, 2, 0} {
		t.Fatal("Korean/English byte divergence not preserved", r.LocaleLengthRows)
	}
	for bucket := range 4 {
		for locale := range 2 {
			for h := range 3 {
				for s := range 3 {
					want := 0
					if s == h {
						want = r.LocaleRuneRows[locale][bucket]
					}
					if r.RuneStates[bucket][locale][h][s] != want {
						t.Fatal("rune head/state counts")
					}
				}
			}
		}
	}
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	for _, text := range []string{strings.Repeat("한", 16), strings.Repeat("a", 16), "fixture-family", "fixture-group"} {
		if bytes.Contains(b, []byte(text)) {
			t.Fatal("rune audit leaked fixture content")
		}
	}
}

func TestNonemptyFalseEvidenceAndPartialSpan(t *testing.T) {
	rows := fixtureRows()
	start, end := 0, 7
	rows[0].Evidence[1].Start, rows[0].Evidence[1].End = &start, &end
	r := auditFixture(t, rows)
	if r.EmptyEvidence[1][1] != 5 || r.NonemptyEvidence[1][1] != 1 {
		t.Fatal("false evidence presence counters")
	}
	f := r.SpanFractions[1][1]
	want := float64(end) / float64(len(rows[0].Text)) / 6
	if f.Count != 6 || math.Abs(f.Mean-want) > 1e-15 || f.Min != 0 || math.Abs(f.Max-want*6) > 1e-15 {
		t.Fatal("row-weighted partial span summaries", f)
	}
}

func TestSemanticAggregatesAndTextFreeOutput(t *testing.T) {
	rows := fixtureRows()
	rows[0].Text = "Alpha  Beta"
	rows[1].Text = " alpha beta "
	labels := [][3]string{{"true", "false", "unknown"}, {"false", "false", "unknown"}, {"true", "true", "true"}, {"true", "true", "true"}, {"unknown", "unknown", "false"}, {"unknown", "false", "false"}}
	for i := range rows {
		rows[i].Targets = append([]string(nil), labels[i][:]...)
		setEvidence(&rows[i])
	}
	r := auditFixture(t, rows)
	if r.Rows != 6 || r.Families != 3 || r.Groups != 1 || r.LocaleRows != [2]int{3, 3} || r.GroupsWithLocale != [2]int{1, 1} {
		t.Fatal("shape counters")
	}
	want := [2][3][3]int{{{2, 0, 1}, {1, 1, 1}, {1, 1, 1}}, {{1, 1, 1}, {1, 2, 0}, {1, 1, 1}}}
	if r.HeadStates != want {
		t.Fatal("head-state counts", r.HeadStates)
	}
	for l := range 2 {
		for h := range 3 {
			for s := range 3 {
				support := 0
				if want[l][h][s] > 0 {
					support = 1
				}
				if r.GroupSupports[l][h][s] != support {
					t.Fatal("group support counts")
				}
			}
		}
	}
	if r.PairDifferences != [3]int{1, 1, 0} || r.Duplicates != (duplicates{1, 1, 1, 1}) {
		t.Fatal("paired or duplicate semantic counters", r.PairDifferences, r.Duplicates)
	}
	if r.EmptyEvidence[0][1] != 1 || r.EmptyEvidence[1][1] != 3 || r.EmptyEvidence[2][1] != 2 {
		t.Fatal("empty false evidence counts")
	}
	for h := range 3 {
		for s := range 3 {
			f := r.SpanFractions[h][s]
			expected := 1.0
			if s == 1 {
				expected = 0
			}
			if f.Mean != expected {
				t.Fatal("span fraction includes wrong evidence state")
			}
		}
	}
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	for _, secret := range []string{"Alpha", "alpha beta", "Private fixture", "fixture-family", "fixture-group", "fixture explanation", "normalized_text_sha256", "text_sha256"} {
		if bytes.Contains(b, []byte(secret)) {
			t.Fatal("row content or identifier leaked", secret)
		}
	}
	for h := range 3 {
		rows[1].Targets[h] = rows[0].Targets[h]
	}
	setEvidence(&rows[1])
	r = auditFixture(t, rows)
	if r.Duplicates != (duplicates{1, 1, 0, 0}) {
		t.Fatal("matching duplicated targets incorrectly conflict")
	}
}
