// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

// Original temporary data and manufactured logits test arithmetic and bindings,
// not native inference, independent labels or semantic performance.
func fixture(t *testing.T, perClass int) (statehintcorpus.Corpus, []byte) {
	t.Helper()
	var b bytes.Buffer
	for class, label := range statehint.Intents() {
		for number := 0; number < perClass; number++ {
			family := fmt.Sprintf("fixture-%d-%02d", class, number)
			for _, locale := range []string{"ko", "en"} {
				r := statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", ID: family + "-" + locale, Family: family, Lineage: fmt.Sprintf("fixture-group-%d", class), Partition: "validation", Locale: locale, Wording: 1, Role: "prose", Applicable: true, Text: "Original temporary arithmetic fixture " + family + " " + locale, Unit: "Original temporary bounded unit.", ClauseScopes: []string{"current_unit"}, AssertionForms: []string{"asserted"}, Expected: label, Ambiguous: label == statehint.Unclear, AnnotationSource: "Original fake arithmetic fixture; not semantic truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: rubricSHA, License: "Apache-2.0"}
				if json.NewEncoder(&b).Encode(r) != nil {
					t.Fatal("fixture encoding")
				}
			}
		}
	}
	c, _, e := statehintcorpus.Read(bytes.NewReader(b.Bytes()), statehintcorpus.Options{Partition: "validation", RubricSHA: rubricSHA})
	if e != nil {
		t.Fatal(e)
	}
	return c, b.Bytes()
}
func scoresFor(c statehintcorpus.Corpus, corpusSHA string) scoreFile {
	order := statehint.Intents()
	s := scoreFile{Schema: "riido-statehint-laya-reference-scores-v1", Status: "reference_only", ValidationSHA: corpusSHA, IntentOrder: append([]statehint.Intent(nil), order[:]...), BaseSHA: baseSHA, DeltaSHA: deltaSHA, BaseTemperature: baseTemperature, DeltaTemperature: deltaTemperature, BaseOrigin: "base_pretrained", DeltaOrigin: "published_v1_delta", InstructionSHA: instructionSHA, TrainingStepsKnown: new(bool), TrainingSteps: json.RawMessage("null"), FeatureCacheSHA: digest([]byte("Original fake feature receipt; no native features"))}
	for _, row := range c.Rows() {
		target, _ := statehint.IntentIndex(row.Expected)
		logits := make([]float64, 8)
		logits[target] = 10
		s.Rows = append(s.Rows, scoreRow{ID: row.ID, Locale: row.Locale, TextSHA: digest([]byte(row.Text)), BaseLogits: logits, DeltaLogits: append([]float64(nil), logits...), BaseProbabilities: fixtureSoftmax(logits, baseTemperature), DeltaProbabilities: fixtureSoftmax(logits, deltaTemperature)})
	}
	return s
}
func fixtureSoftmax(logits []float64, t float64) []float64 {
	values := make([]float64, len(logits))
	sum := 0.0
	for i, v := range logits {
		values[i] = math.Exp(v / t)
		sum += values[i]
	}
	for i := range values {
		values[i] /= sum
	}
	return values
}
func setWinner(row *scoreRow, arm, winner int) {
	v := make([]float64, 8)
	v[winner] = 10
	if arm == 0 {
		row.BaseLogits = v
		row.BaseProbabilities = fixtureSoftmax(v, baseTemperature)
	} else {
		row.DeltaLogits = v
		row.DeltaProbabilities = fixtureSoftmax(v, deltaTemperature)
	}
}
func cloneScores(t *testing.T, s scoreFile) scoreFile {
	t.Helper()
	b, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	v, e := decodeScores(b)
	if e != nil {
		t.Fatal(e)
	}
	return v
}

func TestPairedWrongCompletionAndLineageCountsWithoutQualification(t *testing.T) {
	c, b := fixture(t, 2)
	s := scoresFor(c, digest(b))
	rows := c.Rows()
	for i, row := range rows {
		if strings.HasPrefix(row.Family, "fixture-3-") || row.Family == "fixture-0-00" {
			setWinner(&s.Rows[i], 0, 4)
		}
		setWinner(&s.Rows[i], 1, 7)
	}
	r, e := aggregate(c, s, pin{"validation.jsonl", digest(b)}, pin{"scores.json", strings.Repeat("1", 64)})
	if e != nil {
		t.Fatal(e)
	}
	a := r.Arms[0]
	if a.Rows != 32 || a.Families != 16 || a.Lineages != 8 || a.RawCorrect != 26 || a.GatedProposals != 12 || a.CorrectGated != 6 || a.CompletionCorrect != 4 || a.CompletionProposals != 10 || a.CompletionPrecision != .4 || !a.CompletionPrecisionDefined || a.WrongFamilies != [3]int{0, 3, 0} || a.WrongLineages != [3]int{0, 2, 0} {
		t.Fatal("wrong family/row/lineage denominator", a)
	}
	for _, l := range a.Locales {
		if l.Rows != 16 || l.CorrectFamilies[1] != 2 || l.CorrectLineages[1] != 1 || l.Proposed[1] != 5 || l.Precision[1] != .4 || l.Coverage[1] != 1 {
			t.Fatal("locale completion support", l)
		}
	}
	d := r.Arms[1]
	if d.GatedProposals != 0 || d.CompletionPrecisionDefined || d.CompletionProposals != 0 || d.RawCorrect != 4 {
		t.Fatal("zero completion must be undefined", d)
	}
	if !r.ReferenceOnly || !r.PreviouslyExposed || !r.ConfidenceNotCalibratedOnV4 || r.HumanTruth || r.FreshQualification || r.EligibilityAssessed || r.Selected != "" || r.Promotion || r.DeploymentQualified || r.AdapterModelCalls != 0 || r.CalibrationCalls != 0 || r.TestCalls != 0 {
		t.Fatal("reference silently qualified or selected")
	}
	for _, arm := range r.Arms {
		if arm.TrainingStepsKnown || arm.TrainingSteps != nil || arm.Eligible || arm.EligibilityAssessed {
			t.Fatal("fabricated training history or eligibility")
		}
	}
}

func TestRowPermutationPreservesPairedCountsAndNativeProbabilities(t *testing.T) {
	c, b := fixture(t, 2)
	s := scoresFor(c, digest(b))
	p, e := alignScores(c, s, digest(b))
	if e != nil {
		t.Fatal(e)
	}
	for i, row := range s.Rows {
		for class, v := range row.BaseProbabilities {
			if p[0][i].probabilities[class] != v {
				t.Fatal("native probabilities changed")
			}
		}
	}
	want, e := aggregate(c, s, pin{"v", digest(b)}, pin{"s", strings.Repeat("1", 64)})
	if e != nil {
		t.Fatal(e)
	}
	sort.Slice(s.Rows, func(i, j int) bool { return s.Rows[i].ID > s.Rows[j].ID })
	got, e := aggregate(c, s, pin{"v", digest(b)}, pin{"s", strings.Repeat("1", 64)})
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("score row order altered aggregate")
	}
	rowBytes := bytes.Split(bytes.TrimSpace(b), []byte{'\n'})
	var reverse bytes.Buffer
	for i := len(rowBytes) - 1; i >= 0; i-- {
		reverse.Write(rowBytes[i])
		reverse.WriteByte('\n')
	}
	reversed := append([]byte(nil), reverse.Bytes()...)
	other, _, e := statehintcorpus.Read(bytes.NewReader(reversed), statehintcorpus.Options{Partition: "validation", RubricSHA: rubricSHA})
	if e != nil {
		t.Fatal(e)
	}
	// Source bytes need a new pin, while unchanged ID/text bindings retain outcomes.
	s.ValidationSHA = digest(reversed)
	r, e := aggregate(other, s, pin{"v", s.ValidationSHA}, pin{"s", strings.Repeat("1", 64)})
	if e != nil || !reflect.DeepEqual(r.Arms, want.Arms) {
		t.Fatal("source row order altered paired metrics")
	}
}

func TestInvalidFramesArraysScoresAndUnknownTrainingSteps(t *testing.T) {
	c, b := fixture(t, 1)
	s := scoresFor(c, digest(b))
	changes := []func(*scoreFile){func(v *scoreFile) { v.Rows[0].BaseLogits = v.Rows[0].BaseLogits[:7] }, func(v *scoreFile) { v.Rows[0].BaseProbabilities = append(v.Rows[0].BaseProbabilities, 0) }, func(v *scoreFile) { v.Rows[0].BaseProbabilities[0] = 1.2 }, func(v *scoreFile) { v.Rows[0].DeltaLogits[0] = math.NaN() }, func(v *scoreFile) { v.Rows[0].BaseLogits[0] = math.Inf(1) }, func(v *scoreFile) { v.Rows[0].TextSHA = strings.Repeat("0", 64) }, func(v *scoreFile) { v.Rows[0].Locale = "other" }, func(v *scoreFile) { v.Rows[0].Truncated = true }, func(v *scoreFile) { v.Rows[0].ID = v.Rows[1].ID }, func(v *scoreFile) { v.Rows[0].DeltaProbabilities[0] += .01 }, func(v *scoreFile) { v.BaseTemperature = 1 }, func(v *scoreFile) { v.DeltaOrigin = "trained_go" }, func(v *scoreFile) { v.BaseSHA = deltaSHA }, func(v *scoreFile) { v.IntentOrder[0] = statehint.Progress }, func(v *scoreFile) { v.TrainingStepsKnown = new(bool); *v.TrainingStepsKnown = true }, func(v *scoreFile) { v.TrainingSteps = json.RawMessage("1") }, func(v *scoreFile) { v.TrainingStepsKnown = nil }, func(v *scoreFile) { v.TrainingSteps = nil }, func(v *scoreFile) { v.InstructionSHA = strings.Repeat("0", 64) }, func(v *scoreFile) { v.FeatureCacheSHA = "malformed" }, func(v *scoreFile) { v.Rows = v.Rows[:len(v.Rows)-1] }, func(v *scoreFile) { v.ValidationSHA = strings.Repeat("0", 64) }}
	for i, change := range changes {
		v := cloneScores(t, s)
		change(&v)
		if _, e := alignScores(c, v, digest(b)); e == nil {
			t.Fatalf("invalid case%d accepted", i)
		}
	}
	data, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := decodeScores(append(data, []byte(" {}")...)); e == nil {
		t.Fatal("trailing scores object")
	}
	if _, e := decodeScores(bytes.Replace(data, []byte("\"status\":"), []byte("\"raw_prose\":\"private-marker\",\"status\":"), 1)); e == nil {
		t.Fatal("raw prose or unknown fields admitted")
	}
}

func TestStableTiesAndNLLUnderflowAreDescriptive(t *testing.T) {
	p, e := checkedPrediction(make([]float64, 8), []float64{.125, .125, .125, .125, .125, .125, .125, .125}, .65)
	if e != nil || p.winner != 0 || p.margin != 0 || p.confidence != .125 {
		t.Fatal("raw reference tie rule")
	}
	c, b := fixture(t, 1)
	s := scoresFor(c, digest(b))
	for i := range s.Rows {
		s.Rows[i].BaseLogits = make([]float64, 8)
		for j := range s.Rows[i].BaseLogits {
			s.Rows[i].BaseLogits[j] = -1e6
		}
		s.Rows[i].BaseLogits[7] = 0
		s.Rows[i].BaseProbabilities = []float64{0, 0, 0, 0, 0, 0, 0, 1}
	}
	r, e := aggregate(c, s, pin{"v", digest(b)}, pin{"s", strings.Repeat("1", 64)})
	if e != nil || r.Arms[0].NLLClippedRows != 14 || !finite(r.Arms[0].EightNLL) || r.Arms[0].CompletionPrecisionDefined {
		t.Fatal("underflow clipping or undefined precision", e)
	}
}

func TestCLIOnlyAggregatesExplicitPinnedExposedCorpusAndPrivateOutputs(t *testing.T) {
	dir := t.TempDir()
	root, e := os.OpenRoot(dir)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	c, b := fixture(t, 15)
	s := scoresFor(c, digest(b))
	data, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	for _, item := range []struct {
		name string
		data []byte
	}{{"validation.jsonl", b}, {"scores.json", data}} {
		if e := os.WriteFile(filepath.Join(dir, item.name), item.data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if got, e := readPin(root, pin{"scores.json", digest(data)}, int64(len(data))); e != nil || !bytes.Equal(got, data) {
		t.Fatal("exact bounded score read")
	}
	if _, e := readPin(root, pin{"scores.json", digest(data)}, int64(len(data)-1)); e == nil {
		t.Fatal("over-budget scores")
	}
	if _, e := readPin(root, pin{"scores.json", strings.Repeat("0", 64)}, scoreBudget); e == nil {
		t.Fatal("wrong score pin")
	}
	if e := os.Symlink("scores.json", filepath.Join(dir, "linked.json")); e != nil {
		t.Fatal(e)
	}
	if _, e := readPin(root, pin{"linked.json", digest(data)}, scoreBudget); e == nil {
		t.Fatal("symlink input")
	}
	r, e := aggregate(c, s, pin{"validation.jsonl", digest(b)}, pin{"scores.json", digest(data)})
	if e != nil {
		t.Fatal(e)
	}
	if saveReport(root, anchor+"/run01", r) != nil {
		t.Fatal("private output")
	}
	if saveReport(root, anchor+"/run01", r) == nil {
		t.Fatal("existing run replaced")
	}
	st, e := root.Stat(anchor + "/run01/report.json")
	if e != nil || st.Mode().Perm() != 0600 {
		t.Fatal("report not private")
	}
	st, e = root.Stat(anchor + "/run01")
	if e != nil || st.Mode().Perm() != 0700 {
		t.Fatal("directory not private")
	}
	// No argv reaches a model runner. Unsupported options/errors do not echo input.
	for _, args := range [][]string{nil, {"--fit"}, {"--model", "private-marker"}, {"--test", "private-marker"}, {"--calibration", "private-marker"}, {"--validation", "private-marker", "--check"}} {
		var stdout, stderr bytes.Buffer
		e := run(args, &stdout, &stderr)
		if e == nil || stdout.Len() != 0 || strings.Contains(stderr.String()+e.Error(), "private-marker") {
			t.Fatal("implicit run or private error echo")
		}
	}
	t.Chdir(dir)
	var stdout, stderr bytes.Buffer
	if e := run([]string{"--validation", "validation.jsonl", "--validation-sha256", digest(b), "--scores", "scores.json", "--scores-sha256", digest(data), "--check"}, &stdout, &stderr); e != nil || !strings.Contains(stdout.String(), "\"model_calls\":0") {
		t.Fatal("explicit owned-fixture check", e)
	}
}
