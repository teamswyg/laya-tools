// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintsharedprobe"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fixturePair(id, group string, label statehint.Intent) []statehintcorpus.Row {
	var rows []statehintcorpus.Row
	for _, locale := range []string{"ko", "en"} {
		forms := []string{"asserted"}
		if label == statehint.Unclear {
			forms = []string{}
		}
		rows = append(rows, statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", ID: id + "-" + locale, Family: id, Lineage: group, Partition: "train", Locale: locale, Wording: 1, Role: "prose", Applicable: true, Text: "Original temporary arithmetic fixture " + id + " " + locale, Unit: "Original temporary bounded arithmetic unit.", ClauseScopes: []string{"current_unit"}, AssertionForms: forms, Expected: label, Ambiguous: label == statehint.Unclear, AnnotationSource: "Original artificial test, not semantic truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: rubricSHA, License: "Apache-2.0"})
	}
	return rows
}
func fixtureCorpus(t *testing.T, rows []statehintcorpus.Row) (statehintcorpus.Corpus, []byte) {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if r.AssertionForms == nil {
			r.AssertionForms = []string{}
		}
		if json.NewEncoder(&b).Encode(r) != nil {
			t.Fatal("encode")
		}
	}
	c, _, e := statehintcorpus.Read(bytes.NewReader(b.Bytes()), statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		t.Fatal(e)
	}
	return c, b.Bytes()
}
func fixtureSidecar(c statehintcorpus.Corpus, corpusSHA, featuresSHA string) sidecar {
	order := statehint.Intents()
	s := sidecar{Schema: "riido-statehint-laya-feature-sidecar-v1", BaseSHA: statehintsharedprobe.BaseSHA, InstructionSHA: statehintsharedprobe.InstructionSHA, IntentOrder: append([]statehint.Intent(nil), order[:]...), CorpusSHA: corpusSHA, FeatureSHA: featuresSHA, RowsCount: len(c.Rows()), Shape: [3]int{len(c.Rows()), 8, 1024}, Dtype: "float32_little_endian"}
	rows := c.Rows()
	for i := len(rows) - 1; i >= 0; i-- {
		r := rows[i]
		s.Rows = append(s.Rows, featureIdentity{r.ID, r.Locale, digest([]byte(r.Text)), r.Expected})
	}
	return s
}

func TestSidecarOrderLabelsAndToyTrainedFeatureEvaluation(t *testing.T) {
	var rows []statehintcorpus.Row
	for i, label := range statehint.Intents() {
		rows = append(rows, fixturePair(fmt.Sprintf("toy-%d", i), fmt.Sprintf("toy-group-%d", i), label)...)
	}
	c, b := fixtureCorpus(t, rows)
	data := make([]byte, len(rows)*statehintsharedprobe.RowBytes)
	s := fixtureSidecar(c, digest(b), digest(data))
	mapping, e := mapSidecar(c, s, digest(b), digest(data))
	if e != nil {
		t.Fatal(e)
	}
	for index, id := range s.Rows {
		class, _ := statehint.IntentIndex(id.Expected)
		offset := index*statehintsharedprobe.RowBytes + class*1024*4
		binary.LittleEndian.PutUint32(data[offset:offset+4], math.Float32bits(1))
		if mapping[len(rows)-1-index].RowIndex != index {
			t.Fatal("feature row order mapping")
		}
	}
	store, e := statehintsharedprobe.NewStore(bytes.NewReader(data), len(rows))
	if e != nil {
		t.Fatal(e)
	}
	m := statehintsharedprobe.NewModel()
	if _, e := m.Fit(store, mapping, statehintsharedprobe.FitOptions{Epochs: 2, BatchSize: 4, LearningRate: .001, WeightDecay: .01, Seed: 1729}); e != nil {
		t.Fatal(e)
	}
	p, e := score(m, store, c, mapping)
	if e != nil || len(p) != len(rows) {
		t.Fatal("feature score routing", e)
	}
	for _, v := range p {
		if v.Source != statehint.Learned || v.TrainingSteps != 8 {
			t.Fatal("new head history not real")
		}
	}
	wrong := s
	wrong.Rows = append([]featureIdentity(nil), s.Rows...)
	wrong.Rows[0].Expected = statehint.Question
	if _, e := mapSidecar(c, wrong, digest(b), digest(make([]byte, len(data)))); e == nil {
		t.Fatal("teacher/changed label metadata accepted")
	}
	wrong = s
	wrong.Rows = append([]featureIdentity(nil), s.Rows...)
	wrong.Rows[0].TextSHA = strings.Repeat("0", 64)
	if _, e := mapSidecar(c, wrong, digest(b), s.FeatureSHA); e == nil {
		t.Fatal("wrong body binding")
	}
	wrong = s
	wrong.Shape = [3]int{len(rows), 1024, 8}
	if _, e := mapSidecar(c, wrong, digest(b), s.FeatureSHA); e == nil {
		t.Fatal("transposed feature shape")
	}
	vm := validationScores{Schema: "riido-statehint-laya-reference-scores-v1", Status: "reference_only", ValidationSHA: digest(b), BaseSHA: s.BaseSHA, InstructionSHA: s.InstructionSHA, FeatureSHA: s.FeatureSHA, IntentOrder: s.IntentOrder, Rows: append([]featureIdentity(nil), s.Rows...)}
	for i := range vm.Rows {
		vm.Rows[i].Expected = ""
	}
	got, e := mapValidation(c, vm, digest(b), s.FeatureSHA)
	if e != nil || !reflect.DeepEqual(got, mapping) {
		t.Fatal("validation score mapping required native predictions as labels")
	}
}

func TestFrozen663177SplitAndDevLabelsCannotEnterFit(t *testing.T) {
	counts := [8]int{86, 83, 85, 89, 80, 82, 78, 80}
	s := splitMetadata{Schema: "statehint-completion-contrast-train-only-group-split-v1", Families: 840, FitFamilies: 663, DevFamilies: 177}
	var rows []statehintcorpus.Row
	for class, label := range statehint.Intents() {
		candidate := 0
		for i := 0; i < 105; i++ {
			use := "fit"
			if i >= counts[class] {
				use = "dev"
			}
			group := ""
			for {
				group = fmt.Sprintf("shared-fixture-group-%d-%04d", class, candidate)
				candidate++
				if splitUse(group) == use {
					break
				}
			}
			id := fmt.Sprintf("shared-fixture-%d-%03d", class, i)
			s.Assignments = append(s.Assignments, assignment{id, group, use})
			rows = append(rows, fixturePair(id, group, label)...)
		}
	}
	c, b := fixtureCorpus(t, rows)
	fit, dev, e := splitCorpus(c, s)
	if e != nil || len(fit.Pairs()) != 663 || len(dev.Pairs()) != 177 {
		t.Fatal("frozen split", e)
	}
	side := fixtureSidecar(c, digest(b), strings.Repeat("1", 64))
	all, e := mapSidecar(c, side, digest(b), side.FeatureSHA)
	if e != nil {
		t.Fatal(e)
	}
	training, e := subsetMapping(c, fit, all)
	if e != nil || len(training) != 1326 {
		t.Fatal("fit mapping")
	}
	eval, e := evaluationMapping(c, dev, all)
	if e != nil || len(eval) != 354 {
		t.Fatal("dev mapping")
	}
	sourceRows := c.Rows()
	for _, x := range training {
		id := side.Rows[x.RowIndex].ID
		family := strings.TrimSuffix(strings.TrimSuffix(id, "-ko"), "-en")
		found := false
		for _, a := range s.Assignments {
			if a.Family == family {
				found = true
				if a.Use != "fit" {
					t.Fatal("dev source feature entered fit")
				}
				break
			}
		}
		if !found {
			t.Fatal("missing source family")
		}
		if side.Rows[x.RowIndex].Expected != x.Label {
			t.Fatal("training target mismatch")
		}
	}
	for i := range sourceRows {
		if splitUse(sourceRows[i].Lineage) == "dev" {
			class, _ := statehint.IntentIndex(sourceRows[i].Expected)
			sourceRows[i].Expected = statehint.Intents()[(class+1)%8]
			sourceRows[i].Ambiguous = sourceRows[i].Expected == statehint.Unclear
			sourceRows[i].AssertionForms = []string{"asserted"}
		}
	}
	changed, cb := fixtureCorpus(t, sourceRows)
	otherFit, _, e := splitCorpus(changed, s)
	if e != nil {
		t.Fatal(e)
	}
	otherSide := fixtureSidecar(changed, digest(cb), side.FeatureSHA)
	otherAll, e := mapSidecar(changed, otherSide, digest(cb), side.FeatureSHA)
	if e != nil {
		t.Fatal(e)
	}
	otherTraining, e := subsetMapping(changed, otherFit, otherAll)
	if e != nil || !reflect.DeepEqual(training, otherTraining) {
		t.Fatal("development annotations entered fit")
	}
	for _, corpus := range []statehintcorpus.Corpus{fit, dev} {
		for _, r := range corpus.Rows() {
			if r.Expected == statehint.Unclear && len(r.AssertionForms) != 0 {
				t.Fatal("empty assertions invented")
			}
		}
	}
}

func TestFeatureStreamingHashBoundsAndNoImplicitFit(t *testing.T) {
	dir := t.TempDir()
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	data := make([]byte, statehintsharedprobe.RowBytes)
	if os.WriteFile(filepath.Join(dir, "features.f32le"), data, 0600) != nil {
		t.Fatal("write")
	}
	p := pin{"features.f32le", digest(data)}
	file, e := openFeatures(root, p, 1)
	if e != nil {
		t.Fatal(e)
	}
	var w statehintsharedprobe.Workspace
	if file.store.ReadRow(0, &w) != nil {
		t.Fatal("ReaderAt at EOF position")
	}
	file.file.Close()
	if _, e := openFeatures(root, p, 2); e == nil {
		t.Fatal("incomplete feature file")
	}
	if _, e := openFeatures(root, pin{p.Path, strings.Repeat("0", 64)}, 1); e == nil {
		t.Fatal("wrong feature hash")
	}
	if os.Symlink(p.Path, filepath.Join(dir, "link.f32le")) != nil {
		t.Fatal("symlink")
	}
	if _, e := openFeatures(root, pin{"link.f32le", p.SHA}, 1); e == nil {
		t.Fatal("symlink feature input")
	}
	for _, args := range [][]string{nil, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--model", "private-marker"}, {"--learning-rate", ".1"}, {"--fit"}, {"--check"}} {
		var stdout, stderr bytes.Buffer
		e := run(args, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(e.Error()+stderr.String(), "private-marker") {
			t.Fatal("implicit fit or private argument echo")
		}
	}
}
