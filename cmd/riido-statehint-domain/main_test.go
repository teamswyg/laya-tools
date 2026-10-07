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
				group = fmt.Sprintf("domain-fixture-group-%d-%04d", class, candidate)
				candidate++
				if splitUse(group) == use {
					break
				}
			}
			family := fmt.Sprintf("domain-fixture-source-%d-%03d", class, i)
			s.Assignments = append(s.Assignments, assignment{family, group, use})
			rows = append(rows, fixturePair(family, group, "train", label)...)
		}
	}
	// Two fit families share one original group, exercising full-member checks.
	s.Assignments[1].Group = s.Assignments[0].Group
	rows[2].Lineage, rows[3].Lineage = rows[0].Lineage, rows[0].Lineage
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

func domainFixture(t *testing.T, train input, split metadataInput) (input, metadataInput) {
	t.Helper()
	var sm splitMetadata
	if decodeMetadata(split.bytes, &sm) != nil {
		t.Fatal("split")
	}
	fit, _, a, e := splitCorpus(train.corpus, sm)
	if e != nil {
		t.Fatal(e)
	}
	pairs := fit.Pairs()
	var rows []statehintcorpus.Row
	var source []assignment
	for class, label := range statehint.Intents() {
		source = source[:0]
		for _, p := range pairs {
			if p.Family.Expected == label {
				v, ok := findAssignment(a, p.Family.ID)
				if !ok {
					t.Fatal("fit source")
				}
				source = append(source, v)
			}
		}
		for i := 0; i < 8; i++ {
			x := source[i]
			family := fmt.Sprintf("domain-fixture-aug-%d-%02d", class, i)
			rows = append(rows, fixturePair(family, x.Group, "train", label)...)
		}
	}
	aug := fixtureInput(t, rows, "train", "domain.jsonl")
	audit := sourceAudit{Schema: "statehint-dev-comment-domain-source-groups-v1", AugmentationSHA: aug.pin.SHA, TrainSHA: train.pin.SHA, SplitSHA: split.pin.SHA}
	ar := aug.corpus.Rows()
	for _, p := range aug.corpus.Pairs() {
		entry := sourceEntry{Family: p.Family.ID, Pair: p.Family.ID, Expected: p.Family.Expected, Group: p.Family.Lineage, Groups: []string{p.Family.Lineage}}
		for _, x := range a {
			if x.Group == entry.Group {
				entry.Members = append(entry.Members, x.Family)
			}
		}
		entry.Sources = []string{entry.Members[0]}
		for locale, index := range p.Rows {
			entry.RowIDs[locale] = ar[index].ID
			entry.TextSHA[locale] = digest([]byte(ar[index].Text))
		}
		audit.Entries = append(audit.Entries, entry)
	}
	raw, e := json.Marshal(audit)
	if e != nil {
		t.Fatal(e)
	}
	return aug, metadataInput{pin{"audit.json", digest(raw)}, raw}
}
func cloneAudit(t *testing.T, a sourceAudit) sourceAudit {
	t.Helper()
	b, e := json.Marshal(a)
	if e != nil {
		t.Fatal(e)
	}
	var c sourceAudit
	if json.Unmarshal(b, &c) != nil {
		t.Fatal("audit clone")
	}
	return c
}
func TestBalancedWholeFamilyProfilesAndDiagnosticIsolation(t *testing.T) {
	train, split := fullFixture(t)
	aug, audit := domainFixture(t, train, split)
	p, e := prepare(train, aug, split, audit, nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Fit.Pairs()) != 663 || len(p.Dev.Pairs()) != 177 || len(p.Draws) != 64 || len(p.Samples[0]) != 1454 || len(p.Samples[1]) != 1454 {
		t.Fatal("fixed split/profile sizes")
	}
	var counts [8]int
	pairs, rows := p.Fit.Pairs(), p.Fit.Rows()
	for index, d := range p.Draws {
		class, ok := statehint.IntentIndex(d.Expected)
		if !ok {
			t.Fatal("class")
		}
		counts[class]++
		i := sort.Search(len(pairs), func(i int) bool { return pairs[i].Family.ID >= d.Family })
		if i == len(pairs) || pairs[i].Family.ID != d.Family || pairs[i].Family.Lineage != d.Group {
			t.Fatal("nonfit draw")
		}
		h := sha256.Sum256([]byte("dev-comment-domain-control-1729:" + d.Family))
		if d.Rank != hex.EncodeToString(h[:]) {
			t.Fatal("new frozen draw prefix")
		}
		for locale, rowIndex := range pairs[i].Rows {
			r := rows[rowIndex]
			sample := p.Samples[0][1326+2*index+locale]
			if d.RowIDs[locale] != r.ID || sample.Text != r.Text || sample.Label != r.Expected {
				t.Fatal("whole KOEN family/order")
			}
		}
	}
	if counts != balanced(8) {
		t.Fatal("balanced8 per intent")
	}
	for profile := 0; profile < 2; profile++ {
		if !reflect.DeepEqual(p.Samples[profile][:1326], samples(p.Fit)) {
			t.Fatal("common base order")
		}
	}
	_, m0, e := orderedReceipt(p, 0)
	if e != nil {
		t.Fatal(e)
	}
	_, m1, e := orderedReceipt(p, 1)
	if e != nil {
		t.Fatal(e)
	}
	if m0.ClassRows != m1.ClassRows || m0.LocaleRows != [2]int{727, 727} || m1.LocaleRows != m0.LocaleRows || m0.SamplesSHA == m1.SamplesSHA {
		t.Fatal("matched priors/locale or profile receipt")
	}
	// Reordering source rows changes neither complete family pairing nor samples.
	shuffledRows := train.corpus.Rows()
	sort.Slice(shuffledRows, func(i, j int) bool { return shuffledRows[i].ID > shuffledRows[j].ID })
	shuffled := fixtureInput(t, shuffledRows, "train", train.pin.Path)
	var sa sourceAudit
	if decodeMetadata(audit.bytes, &sa) != nil {
		t.Fatal("audit")
	}
	sa.TrainSHA = shuffled.pin.SHA
	raw, _ := json.Marshal(sa)
	q, e := prepare(shuffled, aug, split, metadataInput{pin{audit.pin.Path, digest(raw)}, raw}, nil)
	if e != nil || !reflect.DeepEqual(p.Samples, q.Samples) || !reflect.DeepEqual(p.Draws, q.Draws) {
		t.Fatal("source order changes training")
	}
	// Already exposed internal-dev labels and units are never training inputs.
	changedRows := train.corpus.Rows()
	for i := range changedRows {
		if splitUse(changedRows[i].Lineage) == "dev" {
			class, _ := statehint.IntentIndex(changedRows[i].Expected)
			changedRows[i].Expected = statehint.Intents()[(class+1)%8]
			changedRows[i].Ambiguous = changedRows[i].Expected == statehint.Unclear
			changedRows[i].Unit = "Changed diagnostic annotation."
			changedRows[i].AssertionForms = []string{"asserted"}
		}
	}
	changed := fixtureInput(t, changedRows, "train", train.pin.Path)
	sa.TrainSHA = changed.pin.SHA
	raw, _ = json.Marshal(sa)
	diagnostic := toyInput(t, "validation", "external")
	otherDiagnostic := rotateAnnotations(t, diagnostic, "validation")
	q, e = prepare(changed, aug, split, metadataInput{pin{audit.pin.Path, digest(raw)}, raw}, &otherDiagnostic)
	if e != nil {
		t.Fatal(e)
	}
	withDiag, e := prepare(train, aug, split, audit, &diagnostic)
	if e != nil || !reflect.DeepEqual(p.Samples, q.Samples) || !reflect.DeepEqual(p.Samples, withDiag.Samples) || !reflect.DeepEqual(p.Draws, q.Draws) {
		t.Fatal("diagnostic labels affect profiles")
	}
	for _, c := range []statehintcorpus.Corpus{p.Fit, p.Dev} {
		for _, r := range c.Rows() {
			if r.Expected == statehint.Unclear && len(r.AssertionForms) != 0 {
				t.Fatal("valid empty assertions altered")
			}
		}
	}
}
func TestDrawCyclingIsWholeFamilyAndDeterministic(t *testing.T) {
	in := toyInput(t, "train", "single")
	extra, draws, e := controlDraws(in.corpus)
	if e != nil || len(draws) != 64 || len(extra) != 128 {
		t.Fatal("cycling")
	}
	for class := 0; class < 8; class++ {
		for i := 0; i < 8; i++ {
			d := draws[class*8+i]
			if d.Family != draws[class*8].Family || d.Expected != statehint.Intents()[class] {
				t.Fatal("class cycle drift")
			}
			if !reflect.DeepEqual(extra[class*16+2*i:class*16+2*i+2], extra[class*16:class*16+2]) {
				t.Fatal("cycled family split")
			}
		}
	}
}
func TestSourceAuditRequiresFitFullMembersAndOneCanonical(t *testing.T) {
	train, split := fullFixture(t)
	aug, audit := domainFixture(t, train, split)
	var sm splitMetadata
	var sa sourceAudit
	decodeMetadata(split.bytes, &sm)
	decodeMetadata(audit.bytes, &sa)
	_, _, a, e := splitCorpus(train.corpus, sm)
	if e != nil {
		t.Fatal(e)
	}
	if e := checkSources(aug, a, sa, train.pin, split.pin); e != nil {
		t.Fatal(e)
	}
	var dev assignment
	var unusedFit assignment
	for _, x := range a {
		if x.Use == "dev" && dev.Family == "" {
			dev = x
		}
		if x.Use == "fit" && strings.HasSuffix(x.Family, "-020") {
			unusedFit = x
		}
	}
	cases := []struct {
		name   string
		mutate func(*sourceAudit)
	}{
		{"binding", func(s *sourceAudit) { s.AugmentationSHA = strings.Repeat("0", 64) }},
		{"schema", func(s *sourceAudit) { s.Schema = "statehint-completion-contrast-source-groups-v1" }},
		{"pair-not-family", func(s *sourceAudit) { s.Entries[0].Pair = "matched-other-family" }},
		{"row-hash", func(s *sourceAudit) { s.Entries[0].TextSHA[0] = strings.Repeat("0", 64) }},
		{"dev-source", func(s *sourceAudit) {
			x := &s.Entries[0]
			x.Sources = append(x.Sources, dev.Family)
			x.Groups = append(x.Groups, dev.Group)
			x.Members = append(x.Members, dev.Family)
		}},
		{"omitted-member", func(s *sourceAudit) {
			x := &s.Entries[0]
			if len(x.Members) < 2 {
				t.Fatal("fixture sibling")
			}
			x.Members = x.Members[:1]
		}},
		{"partial-connected-component", func(s *sourceAudit) {
			x := &s.Entries[0]
			x.Groups = append(x.Groups, unusedFit.Group)
			x.Members = append(x.Members, unusedFit.Family)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := cloneAudit(t, sa)
			tc.mutate(&changed)
			if checkSources(aug, a, changed, train.pin, split.pin) == nil {
				t.Fatal("invalid audit accepted")
			}
		})
	}
	// A second canonical for a reused original group is invalid even when both
	// canonical groups and every listed original member are otherwise fit.
	changed := cloneAudit(t, sa)
	x := &changed.Entries[1]
	x.Group = unusedFit.Group
	x.Groups = append(x.Groups, unusedFit.Group)
	x.Members = append(x.Members, unusedFit.Family)
	rows := aug.corpus.Rows()
	for i := range rows {
		if rows[i].Family == x.Family {
			rows[i].Lineage = x.Group
		}
	}
	other := fixtureInput(t, rows, "train", aug.pin.Path)
	changed.AugmentationSHA = other.pin.SHA
	if checkSources(other, a, changed, train.pin, split.pin) == nil {
		t.Fatal("same original group split across canonical components")
	}
	rows = aug.corpus.Rows()
	for i := range rows {
		rows[i].Partition = "validation"
	}
	other = fixtureInput(t, rows, "validation", aug.pin.Path)
	if disjoint(train, other) == nil {
		t.Fatal("source group crosses partition")
	}

}
func TestBothActualTrainedFormatsReloadAndNoEvaluationLabelTraining(t *testing.T) {
	train := toyInput(t, "train", "training")
	dev := toyInput(t, "train", "development")
	diagnostic := toyInput(t, "validation", "diagnostic")
	changedDev, changedDiagnostic := rotateAnnotations(t, dev, "train"), rotateAnnotations(t, diagnostic, "validation")
	for _, kind := range []armKind{linearArm, mlpArm} {
		t.Run(fixedRecipe().ArmOrder[int(kind)*2], func(t *testing.T) {
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
				r, e := fitArm(out, kind, 0, samples(train.corpus), d.corpus, v)
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

func TestFreshProfilesAndNoDiagnosticSelection(t *testing.T) {
	training := samples(toyInput(t, "train", "fresh").corpus)
	for _, kind := range []armKind{linearArm, mlpArm} {
		first, a, e := fitModel(kind, training)
		if e != nil {
			t.Fatal(e)
		}
		var b0 bytes.Buffer
		if first.save(&b0) != nil {
			t.Fatal("save")
		}
		second, b, e := fitModel(kind, training)
		if e != nil {
			t.Fatal(e)
		}
		var b1 bytes.Buffer
		if second.save(&b1) != nil || !bytes.Equal(b0.Bytes(), b1.Bytes()) || a != b || a.LabelSmoothing != 0 {
			t.Fatal("fresh hard-CE fit reproducibility")
		}
	}
	train, split := fullFixture(t)
	aug, audit := domainFixture(t, train, split)
	val := toyInput(t, "validation", "disjoint")
	rows := val.corpus.Rows()
	rows[0].Text = train.corpus.Rows()[0].Text
	copied := fixtureInput(t, rows, "validation", val.pin.Path)
	if _, e := prepare(train, aug, split, audit, &copied); e == nil {
		t.Fatal("training text used as diagnostic")
	}
	rows = val.corpus.Rows()
	for i := 0; i < 2; i++ {
		rows[i].Lineage = aug.corpus.Pairs()[0].Family.Lineage
	}
	copied = fixtureInput(t, rows, "validation", val.pin.Path)
	if _, e := prepare(train, aug, split, audit, &copied); e == nil {
		t.Fatal("connected fit group used as diagnostic")
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
	if _, _, _, e := splitCorpus(train.corpus, s); e == nil {
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
	for _, args := range [][]string{nil, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--source-overlay", "private-marker"}, {"--parent", "private-marker"}, {"--learning-rate", ".1"}, {"--validation", "private-marker", "--check"}} {
		var stdout, stderr bytes.Buffer
		e := run(args, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(stderr.String()+e.Error(), "private-marker") {
			t.Fatal("implicit fit, unsupported recipe or private error echo")
		}
	}
}

// This full synthetic run checks report wiring and policy, never real data accuracy.
func TestFourFixedArmsHaveNoSelectionAndShareProfileReceipt(t *testing.T) {
	train, split := fullFixture(t)
	coded := func(in input) input {
		rows := in.corpus.Rows()
		for i := range rows {
			rows[i].Text = fmt.Sprintf("Owned artificial pipeline category %s fixture %s locale %s", rows[i].Expected, rows[i].Family, rows[i].Locale)
		}
		return fixtureInput(t, rows, "train", in.pin.Path)
	}
	train = coded(train)
	aug, audit := domainFixture(t, train, split)
	aug = coded(aug)
	var a sourceAudit
	decodeMetadata(audit.bytes, &a)
	a.AugmentationSHA = aug.pin.SHA
	rows := aug.corpus.Rows()
	for i := range a.Entries {
		for _, r := range rows {
			if r.Family == a.Entries[i].Family {
				locale := 0
				if r.Locale == "en" {
					locale = 1
				}
				a.Entries[i].TextSHA[locale] = digest([]byte(r.Text))
			}
		}
	}
	raw, _ := json.Marshal(a)
	audit = metadataInput{pin{audit.pin.Path, digest(raw)}, raw}
	diagnostic := toyInput(t, "validation", "four-arm-diagnostic")
	p, e := prepare(train, aug, split, audit, &diagnostic)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	out, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	r, e := fitAll(out, p)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Arms) != 4 || r.SelectedArm != "" || r.SelectionPerformed || r.InternalDevSelectionWeight != 0 || r.ValidationSelectionWeight != 0 || r.Promotion || r.DeploymentQualified || r.CalibrationCalls != 0 || r.TestCalls != 0 || !r.InternalDevPreviouslyExposed || r.FreshQualification {
		t.Fatal("four-arm diagnostic-only policy")
	}
	for i, arm := range r.Arms {
		if arm.Name != fixedRecipe().ArmOrder[i] || arm.Fit.Samples != 1454 || arm.Fit.TrainingSteps != 1840 || arm.Fit.Batches != 1840 || arm.Fit.LabelSmoothing != 0 || !arm.ReloadParity || arm.Dev.Rows != 354 || arm.Diagnostic == nil || arm.Diagnostic.Rows != 16 {
			t.Fatal("fixed backend/profile reports")
		}
		profile := i % 2
		if arm.TrainingSamplesSHA != r.Training[profile].SamplesSHA {
			t.Fatal("architecture training inputs differ within profile")
		}
		data, e := os.ReadFile(filepath.Join(dir, arm.Name+".report.json"))
		var saved armReport
		if e != nil || json.Unmarshal(data, &saved) != nil || !reflect.DeepEqual(saved, arm) {
			t.Fatal("arm receipt differs from final report")
		}
	}
	if r.Training[0].ClassRows != r.Training[1].ClassRows || r.Training[0].LocaleRows != [2]int{727, 727} || r.Training[1].LocaleRows != [2]int{727, 727} {
		t.Fatal("class/locale profiles differ")
	}
	for _, name := range []string{"train840.source.jsonl", "domain64.source.jsonl", "frozen-split.source.json", "source-groups.audit.json", "exposed-validation.source.jsonl", "control.whole-family-draws.json", "balanced_old_fit_control.ordered-samples.json", "new_domain_training64.ordered-samples.json", "report.json"} {
		st, e := os.Stat(filepath.Join(dir, name))
		if e != nil || st.Mode().Perm() != 0600 {
			t.Fatal("missing/nonprivate fixed-run receipt")
		}
	}
}
