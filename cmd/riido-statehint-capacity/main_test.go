// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintmlp"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// Temporary original fixtures exercise code and serialization, not semantic accuracy.
func fixturePair(family, group, partition string, label statehint.Intent) []statehintcorpus.Row {
	var rows []statehintcorpus.Row
	for _, locale := range []string{"ko", "en"} {
		assertions := []string{"asserted"}
		if label == statehint.Unclear {
			assertions = []string{}
		}
		rows = append(rows, statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", ID: family + "-" + locale, Family: family, Lineage: group, Partition: partition, Locale: locale, Wording: 1, Role: "prose", Applicable: true, Text: fmt.Sprintf("Owned temporary arithmetic fixture %s %s", family, locale), Unit: "Owned temporary bounded arithmetic unit.", ClauseScopes: []string{"current_unit"}, AssertionForms: assertions, Expected: label, Ambiguous: label == statehint.Unclear, AnnotationSource: "Original artificial fixture, not human truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: rubricSHA, License: "Apache-2.0"})
	}
	return rows
}
func fixtureInput(t *testing.T, rows []statehintcorpus.Row, partition, name string) input {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if r.AssertionForms == nil {
			r.AssertionForms = []string{}
		}
		if e := json.NewEncoder(&b).Encode(r); e != nil {
			t.Fatal(e)
		}
	}
	c, _, e := statehintcorpus.Read(bytes.NewReader(b.Bytes()), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if e != nil {
		t.Fatal(e)
	}
	return input{pin{name, digest(b.Bytes())}, b.Bytes(), c}
}
func fullFixture(t *testing.T) (input, metadataInput) {
	t.Helper()
	fitCounts := [8]int{86, 83, 85, 89, 80, 82, 78, 80}
	s := splitMetadata{Schema: "statehint-completion-contrast-train-only-group-split-v1", Families: 840, FitFamilies: 663, DevFamilies: 177}
	var rows []statehintcorpus.Row
	for class, label := range statehint.Intents() {
		candidate := 0
		for i := 0; i < 105; i++ {
			use := "fit"
			if i >= fitCounts[class] {
				use = "dev"
			}
			group := ""
			for {
				group = fmt.Sprintf("capacity-fixture-group-%d-%04d", class, candidate)
				candidate++
				if splitUse(group) == use {
					break
				}
			}
			family := fmt.Sprintf("capacity-fixture-source-%d-%03d", class, i)
			s.Assignments = append(s.Assignments, assignment{family, group, use})
			rows = append(rows, fixturePair(family, group, "train", label)...)
		}
	}
	in := fixtureInput(t, rows, "train", "train.jsonl")
	b, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	return in, metadataInput{pin{"split.json", digest(b)}, b}
}
func toyInput(t *testing.T, partition, prefix string) input {
	t.Helper()
	var rows []statehintcorpus.Row
	for i, label := range statehint.Intents() {
		id := fmt.Sprintf("%s-%d", prefix, i)
		rows = append(rows, fixturePair(id, id, partition, label)...)
	}
	return fixtureInput(t, rows, partition, prefix+".jsonl")
}
func rotateAnnotations(t *testing.T, in input, partition string) input {
	t.Helper()
	rows := in.corpus.Rows()
	for i := range rows {
		index, _ := statehint.IntentIndex(rows[i].Expected)
		rows[i].Expected = statehint.Intents()[(index+1)%8]
		rows[i].Ambiguous = rows[i].Expected == statehint.Unclear
		rows[i].Unit = "Changed development annotation only."
		rows[i].AssertionForms = []string{"asserted"}
	}
	return fixtureInput(t, rows, partition, in.pin.Path)
}

func TestExactControlClassMatchingWholeFamiliesAndDevIsolation(t *testing.T) {
	train, split := fullFixture(t)
	p, e := prepare(train, split, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Fit.Pairs()) != 663 || len(p.Dev.Pairs()) != 177 || len(p.Samples) != 1454 || len(p.Draws) != 64 {
		t.Fatal("fixed sample split changed")
	}
	fitRows, fitPairs := p.Fit.Rows(), p.Fit.Pairs()
	var counts [8]int
	for drawIndex, d := range p.Draws {
		index, ok := statehint.IntentIndex(d.Expected)
		if !ok {
			t.Fatal("unknown draw class")
		}
		counts[index]++
		i := sort.Search(len(fitPairs), func(i int) bool { return fitPairs[i].Family.ID >= d.Family })
		if i == len(fitPairs) || fitPairs[i].Family.ID != d.Family || fitPairs[i].Family.Lineage != d.Group {
			t.Fatal("dev/nonfit family drawn")
		}
		// This prefix belongs to the previous control, not a new capacity seed.
		h := sha256.Sum256([]byte("completion-contrast-class-matched-control-1729:" + d.Family))
		if d.Rank != hex.EncodeToString(h[:]) {
			t.Fatal("control rank changed")
		}
		for locale, rowIndex := range fitPairs[i].Rows {
			r := fitRows[rowIndex]
			s := p.Samples[1326+2*drawIndex+locale]
			if d.RowIDs[locale] != r.ID || s.Text != r.Text || s.Label != r.Expected {
				t.Fatal("draw split locales or annotation injected into text")
			}
		}
	}
	if counts != classCounts() || classCounts() != [8]int{0, 8, 8, 8, 32, 0, 8, 0} {
		t.Fatal("control class prior changed")
	}
	if !reflect.DeepEqual(p.Samples[:1326], samples(p.Fit)) {
		t.Fatal("base training order changed")
	}
	rows := train.corpus.Rows()
	for i := range rows {
		if splitUse(rows[i].Lineage) == "dev" {
			label, _ := statehint.IntentIndex(rows[i].Expected)
			rows[i].Expected = statehint.Intents()[(label+1)%8]
			rows[i].Ambiguous = rows[i].Expected == statehint.Unclear
			rows[i].Unit = "Changed internal development annotation."
			rows[i].AssertionForms = []string{"asserted"}
		}
	}
	changed := fixtureInput(t, rows, "train", train.pin.Path)
	other, e := prepare(changed, split, nil)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(p.Samples, other.Samples) || !reflect.DeepEqual(p.Draws, other.Draws) {
		t.Fatal("internal dev labels affected training or class matching")
	}
	// Source order cannot change paired locale ordering or deterministic draws.
	rows = train.corpus.Rows()
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID > rows[j].ID })
	shuffled := fixtureInput(t, rows, "train", train.pin.Path)
	other, e = prepare(shuffled, split, nil)
	if e != nil || !reflect.DeepEqual(p.Samples, other.Samples) || !reflect.DeepEqual(p.Draws, other.Draws) {
		t.Fatal("source row order affected fixed control")
	}
	for _, c := range []statehintcorpus.Corpus{p.Fit, p.Dev} {
		for _, r := range c.Rows() {
			if r.Expected == statehint.Unclear && len(r.AssertionForms) != 0 {
				t.Fatal("empty unclear assertion list was invented")
			}
		}
	}
}

func TestBothActualTrainedFormatsReloadAndNoEvaluationLabelTraining(t *testing.T) {
	train := toyInput(t, "train", "training")
	dev := toyInput(t, "train", "development")
	diagnostic := toyInput(t, "validation", "diagnostic")
	changedDev, changedDiagnostic := rotateAnnotations(t, dev, "train"), rotateAnnotations(t, diagnostic, "validation")
	for _, kind := range []armKind{linearArm, mlpArm} {
		t.Run(fixedRecipe().ArmOrder[kind], func(t *testing.T) {
			var hashes [2]string
			for i := 0; i < 2; i++ {
				d, v := dev, &diagnostic
				if i == 1 {
					d, v = changedDev, &changedDiagnostic
				}
				dir := t.TempDir()
				out, e := os.OpenRoot(dir)
				if e != nil {
					t.Fatal(e)
				}
				r, e := fitArm(out, kind, samples(train.corpus), d.corpus, v)
				out.Close()
				if e != nil {
					t.Fatal(e)
				}
				hashes[i] = r.Model.SHA
				if !r.ReloadParity || r.Fit.Samples != 16 || r.Fit.Epochs != 40 || r.Fit.Batches != 40 || r.Fit.TrainingSteps != 40 || r.DevCalls != 32 || r.DiagnosticCalls != 32 || r.ProbabilityOrder != statehint.Intents() {
					t.Fatal("fit/parity/call report incomplete")
				}
				want := statehintwide.ArtifactBytes
				if kind == mlpArm {
					want = statehintmlp.ArtifactBytes
				}
				if r.ArtifactBytes != want {
					t.Fatal("wrong artifact size")
				}
				data, e := os.ReadFile(filepath.Join(dir, r.Model.Path))
				if e != nil || digest(data) != r.Model.SHA {
					t.Fatal("artifact hash mismatch")
				}
				m, e := loadModel(kind, data)
				if e != nil {
					t.Fatal(e)
				}
				got, e := predict(m, d.corpus)
				if e != nil {
					t.Fatal(e)
				}
				report, e := statehintfamily.Evaluate(d.corpus, got)
				if e != nil || !reflect.DeepEqual(report, r.Dev) {
					t.Fatal("SDK Prediction alias or report parity drift")
				}
				if kind == mlpArm {
					if _, e := statehintwide.Load(bytes.NewReader(data)); e == nil {
						t.Fatal("MLP accepted by linear loader")
					}
				} else if _, e := statehintmlp.Load(bytes.NewReader(data)); e == nil {
					t.Fatal("linear accepted by MLP loader")
				}
				if _, e := loadModel(kind, append(data, 0)); e == nil {
					t.Fatal("trailing artifact accepted")
				}
				var scores []scoredRow
				b, e := os.ReadFile(filepath.Join(dir, r.Name+".internal-dev.predictions.json"))
				if e != nil || json.Unmarshal(b, &scores) != nil || len(scores) != 16 {
					t.Fatal("all-eight score receipt absent")
				}
				for j, s := range scores {
					if s.Prediction != got[j] || s.Expected != d.corpus.Rows()[j].Expected {
						t.Fatal("score receipt altered prediction")
					}
				}
				entries, e := os.ReadDir(dir)
				if e != nil {
					t.Fatal(e)
				}
				for _, entry := range entries {
					st, e := entry.Info()
					if e != nil || st.Mode().Perm() != 0600 {
						t.Fatal("nonprivate artifact")
					}
				}
			}
			if hashes[0] != hashes[1] {
				t.Fatal("development or diagnostic annotations changed model weights")
			}
		})
	}
}

func TestExposedDiagnosticCannotSelectOrLeakThroughPreparation(t *testing.T) {
	arm := func(name string, cost, nll float64, eligible bool) armReport {
		return armReport{Name: name, Dev: statehintfamily.Report{Eligible: eligible, SeverityCost: cost, EightNLL: nll}, Diagnostic: &statehintfamily.Report{Eligible: true, SeverityCost: 0, EightNLL: 0}}
	}
	a, b := arm("linear", .1, .3, true), arm("mlp", .01, .01, false)
	if selectArm([]armReport{a, b}) != "linear" {
		t.Fatal("diagnostic or ineligible arm selected")
	}
	b.Dev.Eligible = true
	if selectArm([]armReport{a, b}) != "mlp" {
		t.Fatal("severity selection")
	}
	b.Dev.SeverityCost = a.Dev.SeverityCost
	if selectArm([]armReport{a, b}) != "mlp" {
		t.Fatal("NLL tie break")
	}
	b.Dev.EightNLL = a.Dev.EightNLL
	if selectArm([]armReport{a, b}) != "linear" {
		t.Fatal("fixed order tie break")
	}
	a.Dev.Eligible = false
	b.Dev.Eligible = false
	if selectArm([]armReport{a, b}) != "" {
		t.Fatal("ineligible research selection")
	}
	train, split := fullFixture(t)
	val := toyInput(t, "validation", "independent-diagnostic")
	if _, e := prepare(train, split, &val); e != nil {
		t.Fatal(e)
	}
	rows := val.corpus.Rows()
	rows[0].Text = train.corpus.Rows()[0].Text
	copied := fixtureInput(t, rows, "validation", val.pin.Path)
	if _, e := prepare(train, split, &copied); e == nil {
		t.Fatal("copied training body accepted as diagnostic")
	}
	rows = val.corpus.Rows()
	rows[0].Lineage = train.corpus.Pairs()[0].Family.Lineage
	rows[1].Lineage = rows[0].Lineage
	copied = fixtureInput(t, rows, "validation", val.pin.Path)
	if _, e := prepare(train, split, &copied); e == nil {
		t.Fatal("training/internal dev lineage admitted into diagnostic")
	}
}

func TestBoundsFrozenQuotasAndExclusivePrivateOutput(t *testing.T) {
	dir := t.TempDir()
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	in := toyInput(t, "train", "bounded")
	if e := os.WriteFile(filepath.Join(dir, in.pin.Path), in.bytes, 0600); e != nil {
		t.Fatal(e)
	}
	if b, e := readPin(root, in.pin, int64(len(in.bytes))); e != nil || !bytes.Equal(b, in.bytes) {
		t.Fatal("exact input bound")
	}
	if _, e := readPin(root, in.pin, int64(len(in.bytes)-1)); e == nil {
		t.Fatal("over-budget input")
	}
	if _, e := readPin(root, pin{in.pin.Path, strings.Repeat("0", 64)}, statehintcorpus.MaxFileBytes); e == nil {
		t.Fatal("wrong SHA accepted")
	}
	if _, e := loadInput(root, in.pin, "train", 840, balanced(105)); e == nil {
		t.Fatal("tiny corpus bypassed frozen quotas")
	}
	if _, e := loadInput(root, in.pin, "validation", 8, balanced(1)); e == nil {
		t.Fatal("wrong partition accepted")
	}
	if e := os.Symlink(in.pin.Path, filepath.Join(dir, "link.jsonl")); e != nil {
		t.Fatal(e)
	}
	if _, e := readPin(root, pin{"link.jsonl", in.pin.SHA}, statehintcorpus.MaxFileBytes); e == nil {
		t.Fatal("symlink input accepted")
	}
	train, split := fullFixture(t)
	var s splitMetadata
	if decodeMetadata(split.bytes, &s) != nil {
		t.Fatal("split")
	}
	s.Assignments[0].Use = "wrong"
	if _, _, e := splitCorpus(train.corpus, s); e == nil {
		t.Fatal("changed frozen split accepted")
	}
	if decodeMetadata(append(append([]byte(nil), split.bytes...), []byte(" {}")...), &s) == nil {
		t.Fatal("trailing metadata accepted")
	}
	out, e := privateOutput(root, anchor+"/new-run")
	if e != nil {
		t.Fatal(e)
	}
	if writeBytes(out, "kept", []byte("original")) != nil || writeBytes(out, "kept", []byte("replacement")) == nil {
		t.Fatal("exclusive artifact write")
	}
	out.Close()
	if _, e := privateOutput(root, anchor+"/new-run"); e == nil {
		t.Fatal("existing run reused")
	}
	st, e := root.Stat(anchor + "/new-run")
	if e != nil || st.Mode().Perm() != 0700 {
		t.Fatal("private output mode")
	}
	for _, args := range [][]string{nil, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--augmentation", "private-marker"}, {"--parent", "private-marker"}, {"--learning-rate", ".1"}, {"--validation", "private-marker", "--check"}} {
		var stdout, stderr bytes.Buffer
		e := run(args, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(stderr.String()+e.Error(), "private-marker") {
			t.Fatal("implicit fit, unsupported recipe or private error echo")
		}
	}
}
