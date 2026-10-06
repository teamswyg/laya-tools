// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintcorpus

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

var testSHA = strings.Repeat("a", 64)

func fixture() []Row {
	base := Row{Schema: "statehint-v4-original-train-seed-row-v1", Family: "fictional-001", Lineage: "fictional-lineage-001", Partition: "train", Wording: 1, Role: "prose", Applicable: true, Unit: "Choose a fictional queue retry limit.", ClauseScopes: []string{"current_unit"}, AssertionForms: []string{"requested"}, Expected: statehint.Question, AnnotationSource: "Original fictional parser fixture, not verified truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: testSHA, License: "Apache-2.0"}
	ko, en := base, base
	ko.ID, ko.Locale, ko.Text = "fictional-001-ko", "ko", "가상 대기열에서 재시도는 몇 번으로 정해야 하나요?"
	en.ID, en.Locale, en.Text = "fictional-001-en", "en", "How many retries should the fictional queue use?"
	return []Row{ko, en}
}

func encodeRows(t *testing.T, rows []Row) string {
	t.Helper()
	var out bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&out).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	return out.String()
}

func TestReadPairedSourceOrderAndMetadata(t *testing.T) {
	rows := fixture()
	corpus, summary, err := Read(strings.NewReader(encodeRows(t, rows)), Options{"train", testSHA})
	if err != nil || summary.Rows != 2 || summary.Families != 1 || summary.Lineages != 1 || summary.LocaleRows != [2]int{1, 1} || corpus.pairs[0].Rows != [2]int{0, 1} || corpus.rows[0].Text != rows[0].Text {
		t.Fatalf("paired source contract: summary=%+v err=%v", summary, err)
	}
	metadata := []Family{corpus.pairs[0].Family, {ID: "fictional-002", Lineage: "fictional-lineage-002", Partition: "test", Expected: statehint.Progress}}
	if err := MatchMetadata(corpus, metadata, "train"); err != nil {
		t.Fatal(err)
	}
	copyRows, copyPairs := corpus.Rows(), corpus.Pairs()
	copyRows[0].Expected = statehint.CompletionReport
	copyRows[0].ClauseScopes[0] = "corrupted"
	copyPairs[0].Rows[0] = -1
	if corpus.rows[0].Expected != statehint.Question || corpus.rows[0].ClauseScopes[0] != "current_unit" || corpus.pairs[0].Rows[0] != 0 || MatchMetadata(corpus, metadata, "train") != nil {
		t.Fatal("detached snapshot corrupted validated corpus")
	}
	metadata[0].Expected = statehint.CompletionReport
	if MatchMetadata(corpus, metadata, "train") == nil {
		t.Fatal("manifest label mismatch accepted")
	}
	encoded, _ := json.Marshal(corpus)
	if string(encoded) != "{}" {
		t.Fatal("corpus JSON leaked fields")
	}
}

func TestRejectMalformedPairs(t *testing.T) {
	tests := []struct {
		name string
		edit func([]Row) []Row
	}{
		{"single row", func(r []Row) []Row { return r[:1] }},
		{"same locale", func(r []Row) []Row { r[1].Locale = "ko"; return r }},
		{"duplicate id", func(r []Row) []Row { r[1].ID = r[0].ID; return r }},
		{"duplicate text", func(r []Row) []Row { r[1].Text = r[0].Text; return r }},
		{"different family", func(r []Row) []Row { r[1].Family = "different"; return r }},
		{"different lineage", func(r []Row) []Row { r[1].Lineage = "different"; return r }},
		{"different label", func(r []Row) []Row { r[1].Expected = statehint.Progress; return r }},
		{"different unit", func(r []Row) []Row { r[1].Unit = "Different hidden unit"; return r }},
		{"different ambiguity", func(r []Row) []Row {
			r[0].Expected = statehint.Unclear
			r[1].Expected = statehint.Unclear
			r[1].Ambiguous = true
			return r
		}},
		{"partition crossing", func(r []Row) []Row { r[1].Partition = "test"; return r }},
		{"rubric mismatch", func(r []Row) []Row { r[1].OntologyFreezeSHA = strings.Repeat("b", 64); return r }},
		{"outside prose", func(r []Row) []Row { r[1].Role = "metadata"; return r }},
		{"not applicable", func(r []Row) []Row { r[1].Applicable = false; return r }},
		{"state proof", func(r []Row) []Row { r[1].TaskLevelCompletion = true; return r }},
		{"nul text", func(r []Row) []Row { r[1].Text += "\x00"; return r }},
		{"too much text", func(r []Row) []Row { r[1].Text = strings.Repeat("x", statehint.MaxTextBytes+1); return r }},
		{"empty text", func(r []Row) []Row { r[1].Text = " "; return r }},
		{"unknown scope", func(r []Row) []Row { r[1].ClauseScopes = []string{"guess"}; return r }},
		{"duplicate scope", func(r []Row) []Row { r[1].ClauseScopes = []string{"current_unit", "current_unit"}; return r }},
		{"no forms", func(r []Row) []Row { r[1].AssertionForms = nil; return r }},
		{"no definite assertion form", func(r []Row) []Row { r[1].AssertionForms = []string{}; return r }},
		{"extra wording", func(r []Row) []Row { r[1].Wording = 2; return r }},
		{"wrong license", func(r []Row) []Row { r[1].License = "unknown"; return r }},
		{"four row family", func(r []Row) []Row {
			b := fixture()
			b[0].ID += "x"
			b[1].ID += "x"
			b[0].Text += "more"
			b[1].Text += "more"
			return append(r, b...)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			corpus, summary, err := Read(strings.NewReader(encodeRows(t, test.edit(fixture()))), Options{"train", testSHA})
			if err != ErrInput || len(corpus.rows) != 0 || summary != (Summary{}) {
				t.Fatal("invalid source returned usable partial result")
			}
		})
	}
}

func TestStrictWireAndBudgets(t *testing.T) {
	valid := encodeRows(t, fixture())
	inputs := []string{
		strings.Replace(valid, `"schema":`, `"schema":"duplicate","schema":`, 1),
		strings.Replace(valid, `"schema":`, `"Schema":`, 1),
		strings.Replace(valid, `"schema":`, `"unexpected":0,"schema":`, 1),
		strings.Replace(valid, `"semantic_applicable":true`, `"semantic_applicable":null`, 1),
		strings.Replace(valid, `"wording_index":1`, `"wording_index":"1"`, 1),
		strings.Replace(valid, `"ambiguous":false,`, "", 1),
		strings.Replace(valid, `"text":`, "\xff\"text\":", 1),
		valid + "{}\n", valid + "\n", "null\n", "",
	}
	for i, input := range inputs {
		if _, _, err := Read(strings.NewReader(input), Options{"train", testSHA}); err != ErrInput {
			t.Fatalf("malformed wire %d accepted", i)
		}
	}
	for _, options := range []Options{{"unknown", testSHA}, {"train", "bad"}, {"train", strings.Repeat("A", 64)}} {
		if _, _, err := Read(strings.NewReader(valid), options); err != ErrInput {
			t.Fatal("invalid caller contract accepted")
		}
	}
	if _, _, err := Read(nil, Options{"train", testSHA}); err != ErrInput {
		t.Fatal("nil reader accepted")
	}
	if _, _, err := Read(io.MultiReader(strings.NewReader(valid), &spacesReader{remaining: MaxFileBytes + 1}), Options{"train", testSHA}); err != ErrInput {
		t.Fatal("over-budget stream accepted")
	}
}

type spacesReader struct{ remaining int }

func (r *spacesReader) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := min(len(p), r.remaining)
	for i := range p[:n] {
		p[i] = ' '
	}
	r.remaining -= n
	return n, nil
}

func TestMetadataLeakageAndDeclaredCounts(t *testing.T) {
	families := []Family{{ID: "a", Lineage: "shared", Partition: "train", Expected: statehint.Question}, {ID: "b", Lineage: "shared", Partition: "train", Expected: statehint.Progress}, {ID: "c", Lineage: "separate", Partition: "test", Expected: statehint.CompletionReport}}
	summary, err := ValidateMetadata(families)
	if err != nil || summary.Lineages != 2 || summary.Families != 3 || summary.Rows != 6 || summary.PartitionFamilies != [4]int{2, 0, 0, 1} {
		t.Fatalf("metadata counts: %+v %v", summary, err)
	}
	families[1].Partition = "validation"
	if _, err := ValidateMetadata(families); err != ErrInput {
		t.Fatal("shared incident crossed partition")
	}
	families[1].Partition, families[1].ID = "train", "a"
	if _, err := ValidateMetadata(families); err != ErrInput {
		t.Fatal("duplicate family accepted")
	}
}

func TestUnclearWithoutDeclaredAssertionForm(t *testing.T) {
	rows := fixture()
	for i := range rows {
		rows[i].Expected, rows[i].Ambiguous = statehint.Unclear, true
		rows[i].AssertionForms = []string{}
		rows[i].ClauseScopes = []string{"unknown"}
	}
	// Labels are author metadata in this syntax-only fixture, not truth inferred
	// from its wording. An empty explicit array differs from a missing/null key.
	if _, _, err := Read(strings.NewReader(encodeRows(t, rows)), Options{"train", testSHA}); err != nil {
		t.Fatal("unclear explicit empty form rejected", err)
	}
}
