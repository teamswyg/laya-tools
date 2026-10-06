// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
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
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// All fixtures are original temporary metadata and artificial prose. Tiny
// model fits verify execution/serialization; they are not semantic evidence.
func fixturePair(family, group, partition string, label statehint.Intent) []statehintcorpus.Row {
	var rows []statehintcorpus.Row
	for _, locale := range []string{"ko", "en"} {
		rows = append(rows, statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", ID: family + "-" + locale, Family: family, Lineage: group, Partition: partition, Locale: locale, Wording: 1, Role: "prose", Applicable: true, Text: fmt.Sprintf("Original temporary arithmetic fixture %s %s %s", family, locale, label), Unit: "Original temporary bounded arithmetic fixture.", ClauseScopes: []string{"current_unit"}, AssertionForms: []string{"asserted"}, Expected: label, Ambiguous: label == statehint.Unclear, AnnotationSource: "Owned synthetic metadata, not human semantic truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: rubricSHA, License: "Apache-2.0"})
	}
	return rows
}
func encodeRows(t *testing.T, rows []statehintcorpus.Row) []byte {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if e := json.NewEncoder(&b).Encode(r); e != nil {
			t.Fatal(e)
		}
	}
	return b.Bytes()
}
func fixtureInput(t *testing.T, rows []statehintcorpus.Row, partition, name string) input {
	t.Helper()
	b := encodeRows(t, rows)
	c, s, e := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if e != nil {
		t.Fatal(e)
	}
	return input{pin{name, digest(b)}, b, c, s}
}
func jsonBytes(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func fullFixture(t *testing.T) (input, input, metadataInput, metadataInput) {
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
				group = fmt.Sprintf("fixture-group-%d-%04d", class, candidate)
				candidate++
				if splitUse(group) == use {
					break
				}
			}
			family := fmt.Sprintf("fixture-source-%d-%03d", class, i)
			s.Assignments = append(s.Assignments, assignment{family, group, use})
			rows = append(rows, fixturePair(family, group, "train", label)...)
		}
	}
	train := fixtureInput(t, rows, "train", "train.jsonl")
	splitBytes := jsonBytes(t, s)
	split := metadataInput{pin{"split.json", digest(splitBytes)}, splitBytes}
	var fit []assignment
	for _, a := range s.Assignments {
		if a.Use == "fit" {
			fit = append(fit, a)
		}
	}
	var augRows []statehintcorpus.Row
	aud := sourceAudit{Schema: "statehint-completion-contrast-source-groups-v1", TrainSHA: train.pin.SHA, SplitSHA: split.pin.SHA}
	other := [4]statehint.Intent{statehint.Reference, statehint.Progress, statehint.Planned, statehint.Blocker}
	for i := 0; i < 32; i++ {
		source := fit[i]
		for member, label := range [2]statehint.Intent{statehint.CompletionReport, other[i%4]} {
			family := fmt.Sprintf("fixture-augmentation-%02d-%d", i, member)
			r := fixturePair(family, source.Group, "train", label)
			augRows = append(augRows, r...)
			aud.Entries = append(aud.Entries, sourceEntry{Family: family, Pair: fmt.Sprintf("fixture-pair-%02d", i), Expected: label, Group: source.Group, Sources: []string{source.Family}, Groups: []string{source.Group}, Members: []string{source.Family}, RowIDs: [2]string{r[0].ID, r[1].ID}, TextSHA: [2]string{digest([]byte(r[0].Text)), digest([]byte(r[1].Text))}})
		}
	}
	aug := fixtureInput(t, augRows, "train", "augmentation.jsonl")
	aud.AugmentationSHA = aug.pin.SHA
	auditBytes := jsonBytes(t, aud)
	return train, aug, split, metadataInput{pin{"audit.json", digest(auditBytes)}, auditBytes}
}
func toyInput(t *testing.T, partition string) input {
	t.Helper()
	var rows []statehintcorpus.Row
	for i, label := range statehint.Intents() {
		family := fmt.Sprintf("%s-toy-%d", partition, i)
		rows = append(rows, fixturePair(family, family, partition, label)...)
	}
	return fixtureInput(t, rows, partition, partition+".jsonl")
}

func TestMatchedCountsDeterministicWholeLocaleDrawsAndDevIsolation(t *testing.T) {
	train, aug, split, audit := fullFixture(t)
	p, e := prepare(train, aug, split, audit, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Fit.Pairs()) != 663 || len(p.Dev.Pairs()) != 177 || len(p.Draws) != 64 {
		t.Fatal("split/draw counts changed")
	}
	var got [2][8]int
	for arm, ss := range p.Samples {
		if len(ss) != 1454 {
			t.Fatal("unequal arm budget")
		}
		for _, s := range ss {
			class, _ := statehint.IntentIndex(s.Label)
			got[arm][class]++
		}
	}
	if got[0] != got[1] {
		t.Fatal("class priors differ")
	}
	for _, d := range p.Draws {
		a, ok := findAssignment(p.Assignments, d.Family)
		if !ok || a.Use != "fit" || a.Group != d.Group || !strings.HasSuffix(d.RowIDs[0], "-ko") || !strings.HasSuffix(d.RowIDs[1], "-en") {
			t.Fatal("draw lost whole fit-side KO/EN family")
		}
	}
	extra, draws, e := controlDraws(p.Fit)
	if e != nil || len(extra) != 128 || !reflect.DeepEqual(draws, p.Draws) {
		t.Fatal("ranking not deterministic")
	}
	changed := train.corpus.Rows()
	devID := p.Dev.Pairs()[0].Family.ID
	for i := range changed {
		if changed[i].Family == devID {
			changed[i].Text = "Changed development-only arithmetic observation " + changed[i].Locale
			changed[i].Unit = "Changed dev annotation only."
		}
	}
	modified := fixtureInput(t, changed, "train", train.pin.Path)
	var a sourceAudit
	if decodeMetadata(audit.bytes, &a) != nil {
		t.Fatal("fixture audit")
	}
	a.TrainSHA = modified.pin.SHA
	ab := jsonBytes(t, a)
	q, e := prepare(modified, aug, split, metadataInput{pin{"audit.json", digest(ab)}, ab}, nil)
	if e != nil || !reflect.DeepEqual(p.Samples, q.Samples) {
		t.Fatal("internal dev text/annotations entered training", e)
	}
	// Cycling repeats a ranked whole family when a class has too few fits.
	toy := toyInput(t, "train")
	_, cycled, e := controlDraws(toy.corpus)
	if e != nil || len(cycled) != 64 {
		t.Fatal(e)
	}
	var completions []draw
	for _, d := range cycled {
		if d.Expected == statehint.CompletionReport {
			completions = append(completions, d)
		}
	}
	if len(completions) != 32 || completions[0] != completions[31] {
		t.Fatal("draw cycling invented new family identity")
	}
}

func TestSourceDevConnectionMissingMembersPinsAndPairGroupsRejected(t *testing.T) {
	train, aug, split, audit := fullFixture(t)
	var s splitMetadata
	var a sourceAudit
	_ = decodeMetadata(split.bytes, &s)
	_ = decodeMetadata(audit.bytes, &a)
	_, _, assignments, e := splitCorpus(train.corpus, s)
	if e != nil {
		t.Fatal(e)
	}
	var dev assignment
	for _, x := range assignments {
		if x.Use == "dev" {
			dev = x
			break
		}
	}
	mutations := []func(*sourceAudit){func(x *sourceAudit) { x.AugmentationSHA = strings.Repeat("0", 64) }, func(x *sourceAudit) { x.TrainSHA = strings.Repeat("0", 64) }, func(x *sourceAudit) { x.SplitSHA = strings.Repeat("0", 64) }, func(x *sourceAudit) {
		x.Entries[0].Sources = []string{dev.Family}
		x.Entries[0].Groups = append(x.Entries[0].Groups, dev.Group)
		x.Entries[0].Members = append(x.Entries[0].Members, dev.Family)
	}, func(x *sourceAudit) { x.Entries[0].TextSHA[0] = strings.Repeat("0", 64) }, func(x *sourceAudit) { x.Entries[0].Pair = "different-pair" }, func(x *sourceAudit) { x.Entries[0].Members = nil }, func(x *sourceAudit) { x.Entries[0].Groups = nil }, func(x *sourceAudit) { x.Entries[0].Sources = []string{"unknown-source"} }, func(x *sourceAudit) { x.Entries[1] = x.Entries[0] }}
	for i, mutate := range mutations {
		var x sourceAudit
		_ = decodeMetadata(audit.bytes, &x)
		mutate(&x)
		if checkSources(aug, assignments, x, train.pin, split.pin) == nil {
			t.Fatalf("invalid source audit %d accepted", i)
		}
	}
	connected := append([]assignment(nil), assignments...)
	first := a.Entries[0]
	for i, x := range connected {
		if x.Use == "fit" && x.Family != first.Sources[0] {
			connected[i].Group = first.Group
			break
		}
	}
	if checkSources(aug, connected, a, train.pin, split.pin) == nil {
		t.Fatal("omitted connected old fit member accepted")
	}
	s.Assignments[0].Use = "wrong"
	if _, _, _, e := splitCorpus(train.corpus, s); e == nil {
		t.Fatal("malformed split accepted")
	}
	if decodeMetadata(append(append([]byte(nil), audit.bytes...), []byte(" {}")...), &a) == nil {
		t.Fatal("trailing metadata object accepted")
	}
}

func TestPinnedBoundedReaderAndInputQuotas(t *testing.T) {
	dir := t.TempDir()
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	in := toyInput(t, "train")
	path := filepath.Join(dir, in.pin.Path)
	if e := os.WriteFile(path, in.bytes, 0600); e != nil {
		t.Fatal(e)
	}
	if b, e := readPin(root, in.pin, int64(len(in.bytes))); e != nil || !bytes.Equal(b, in.bytes) {
		t.Fatal("exact bound rejected", e)
	}
	if _, e := readPin(root, in.pin, int64(len(in.bytes)-1)); e == nil {
		t.Fatal("over-budget input accepted")
	}
	for _, hash := range []string{strings.Repeat("0", 64), strings.ToUpper(in.pin.SHA), "malformed"} {
		if _, e := readPin(root, pin{in.pin.Path, hash}, statehintcorpus.MaxFileBytes); e == nil {
			t.Fatal("bad pin accepted")
		}
	}
	if _, e := loadInput(root, in.pin, "train", 840, balanced(105)); e == nil {
		t.Fatal("tiny input bypassed frozen quota")
	}
	if _, e := loadInput(root, in.pin, "validation", 8, balanced(1)); e == nil {
		t.Fatal("wrong partition accepted")
	}
	if e := os.Symlink(in.pin.Path, filepath.Join(dir, "link.jsonl")); e != nil {
		t.Fatal(e)
	}
	if _, e := readPin(root, pin{"link.jsonl", in.pin.SHA}, statehintcorpus.MaxFileBytes); e == nil {
		t.Fatal("symlink accepted")
	}
}

func TestOnlyEligibleInternalDevSelectsByCostNLLFixedOrder(t *testing.T) {
	arm := func(name string, cost, nll float64, eligible bool) armReport {
		return armReport{Name: name, Dev: statehintfamily.Report{Eligible: eligible, SeverityCost: cost, EightNLL: nll}, Diagnostic: &statehintfamily.Report{Eligible: true, SeverityCost: 0}}
	}
	a, b := arm("control", .1, .3, true), arm("contrast", .01, .01, false)
	if selectArm([]armReport{a, b}) != "control" {
		t.Fatal("ineligible/diagnostic arm selected")
	}
	b.Dev.Eligible = true
	if selectArm([]armReport{a, b}) != "contrast" {
		t.Fatal("severity ranking")
	}
	b.Dev.SeverityCost = a.Dev.SeverityCost
	if selectArm([]armReport{a, b}) != "contrast" {
		t.Fatal("NLL tie breaker")
	}
	b.Dev.EightNLL = a.Dev.EightNLL
	if selectArm([]armReport{a, b}) != "control" {
		t.Fatal("fixed arm order tie breaker")
	}
	a.Dev.Eligible = false
	b.Dev.Eligible = false
	if selectArm([]armReport{a, b}) != "" {
		t.Fatal("both ineligible still selected")
	}
}

func TestActualTrainedReloadParityAndDiagnosticLabelsCannotChangeWeights(t *testing.T) {
	train, dev, val := toyInput(t, "train"), toyInput(t, "train"), toyInput(t, "validation")
	training := samples(train.corpus)
	changed := val.corpus.Rows()
	for i := range changed {
		index, _ := statehint.IntentIndex(changed[i].Expected)
		changed[i].Expected = statehint.Intents()[(index+1)%8]
		changed[i].Ambiguous = changed[i].Expected == statehint.Unclear
		changed[i].Unit = "Different diagnostic annotation only."
	}
	other := fixtureInput(t, changed, "validation", val.pin.Path)
	var pins [2]string
	for i, v := range []*input{&val, &other} {
		dir := t.TempDir()
		out, e := os.OpenRoot(dir)
		if e != nil {
			t.Fatal(e)
		}
		arm, e := fitArm(out, "toy", training, dev.corpus, v)
		out.Close()
		if e != nil {
			t.Fatal(e)
		}
		pins[i] = arm.Model.SHA
		if !arm.ReloadParity || arm.Fit.TrainingSteps != 40 || arm.ArtifactBytes != statehintwide.ArtifactBytes || arm.DevCalls != 32 || arm.DiagnosticCalls != 32 {
			t.Fatal("real toy fit/reload record incomplete")
		}
		data, e := os.ReadFile(filepath.Join(dir, arm.Model.Path))
		if e != nil || digest(data) != pins[i] {
			t.Fatal("model pin")
		}
		m, e := statehintwide.Load(bytes.NewReader(data))
		if e != nil {
			t.Fatal(e)
		}
		pred, e := predict(m, dev.corpus)
		if e != nil {
			t.Fatal(e)
		}
		r, e := statehintfamily.Evaluate(dev.corpus, pred)
		if e != nil || !reflect.DeepEqual(r, arm.Dev) {
			t.Fatal("family report changed after reload")
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
	if pins[0] != pins[1] {
		t.Fatal("diagnostic labels or annotations affected trained weights")
	}
}

func TestPrivateExclusiveOutputAndNoImplicitModelRun(t *testing.T) {
	root, e := os.OpenRoot(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	out, e := privateOutput(root, anchor+"/test-run")
	if e != nil {
		t.Fatal(e)
	}
	if writeBytes(out, "kept", []byte("original")) != nil {
		t.Fatal("write")
	}
	if writeBytes(out, "kept", []byte("replacement")) == nil {
		t.Fatal("overwrote saved source")
	}
	out.Close()
	if _, e := privateOutput(root, anchor+"/test-run"); e == nil {
		t.Fatal("existing run reused")
	}
	st, e := root.Stat(anchor + "/test-run")
	if e != nil || st.Mode().Perm() != 0700 {
		t.Fatal("run not private")
	}
	for _, args := range [][]string{nil, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--learning-rate", ".1"}, {"--validation", "private-marker", "--check"}} {
		var stdout, stderr bytes.Buffer
		e := run(args, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(stderr.String()+e.Error(), "private-marker") {
			t.Fatal("implicit fit, unsupported input or private error echo")
		}
	}
}

func TestInputDisjointnessRejectsCopiedBodyAndDevLineageDiagnostic(t *testing.T) {
	train, aug, _, _ := fullFixture(t)
	var vr []statehintcorpus.Row
	for i := 0; i < 8; i++ {
		vr = append(vr, fixturePair(fmt.Sprintf("diagnostic-%d", i), fmt.Sprintf("diagnostic-%d", i), "validation", statehint.Intents()[i])...)
	}
	val := fixtureInput(t, vr, "validation", "validation.jsonl")
	if disjoint(train, aug, val) != nil {
		t.Fatal("unrelated fixture rejected")
	}
	vr[0].Text = train.corpus.Rows()[0].Text
	copy := fixtureInput(t, vr, "validation", val.pin.Path)
	if disjoint(train, aug, copy) == nil {
		t.Fatal("copied training text admitted as diagnostic")
	}
	vr = val.corpus.Rows()
	vr[0].Lineage = train.corpus.Pairs()[0].Family.Lineage
	vr[1].Lineage = vr[0].Lineage
	copy = fixtureInput(t, vr, "validation", val.pin.Path)
	if disjoint(train, aug, copy) == nil {
		t.Fatal("training/dev lineage admitted into diagnostic")
	}
}

func TestDrawOrderIndependentOfSourceRowOrder(t *testing.T) {
	train, aug, split, audit := fullFixture(t)
	p, e := prepare(train, aug, split, audit, nil)
	if e != nil {
		t.Fatal(e)
	}
	rows := p.Fit.Rows()
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID > rows[j].ID })
	c, e := corpusFromRows(rows)
	if e != nil {
		t.Fatal(e)
	}
	ss, dd, e := controlDraws(c)
	original, draws, e2 := controlDraws(p.Fit)
	if e != nil || e2 != nil || !reflect.DeepEqual(ss, original) || !reflect.DeepEqual(dd, draws) {
		t.Fatal("source row order changed deterministic whole-family draw")
	}
}

func TestSplitPreservesValidUnclearEmptyAssertionLists(t *testing.T) {
	train, aug, split, audit := fullFixture(t)
	rows := train.corpus.Rows()
	for i := range rows {
		if rows[i].Expected == statehint.Unclear {
			rows[i].AssertionForms = []string{}
		}
	}
	modified := fixtureInput(t, rows, "train", train.pin.Path)
	var source sourceAudit
	if decodeMetadata(audit.bytes, &source) != nil {
		t.Fatal("fixture audit")
	}
	source.TrainSHA = modified.pin.SHA
	ab := jsonBytes(t, source)
	p, err := prepare(modified, aug, split, metadataInput{pin{"audit.json", digest(ab)}, ab}, nil)
	if err != nil || len(p.Fit.Pairs()) != 663 || len(p.Dev.Pairs()) != 177 {
		t.Fatal("valid empty assertion arrays lost during split reconstruction", err)
	}
	for _, corpus := range []statehintcorpus.Corpus{p.Fit, p.Dev} {
		for _, row := range corpus.Rows() {
			if row.Expected == statehint.Unclear && len(row.AssertionForms) != 0 {
				t.Fatal("reconstruction invented an assertion")
			}
		}
	}
}
