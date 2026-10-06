// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfamily

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

// Metadata/math fixtures only. No learned model is constructed or run.
func corpusFixture(t *testing.T, labels []statehint.Intent, lineage func(int) string) statehintcorpus.Corpus {
	t.Helper()
	var out bytes.Buffer
	sha := strings.Repeat("a", 64)
	for i, label := range labels {
		for _, locale := range []string{"ko", "en"} {
			r := statehintcorpus.Row{Schema: "statehint-v4-original-train-seed-row-v1", ID: fmt.Sprintf("fixture-%03d-%s", i, locale), Family: fmt.Sprintf("fixture-%03d", i), Lineage: lineage(i), Partition: "validation", Locale: locale, Wording: 1, Role: "prose", Applicable: true, Text: fmt.Sprintf("Invented arithmetic fixture %d %s", i, locale), Unit: "Fictional bounded arithmetic observation.", ClauseScopes: []string{"current_unit"}, AssertionForms: []string{"asserted"}, Expected: label, AnnotationSource: "Original metadata fixture, not product truth.", OntologyVersion: "statehint-intent-scope-v4-1200x2-v1", OntologyFreezeSHA: sha, License: "Apache-2.0"}
			if err := json.NewEncoder(&out).Encode(r); err != nil {
				t.Fatal(err)
			}
		}
	}
	c, _, err := statehintcorpus.Read(&out, statehintcorpus.Options{Partition: "validation", RubricSHA: sha})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func confident(intent, alternative statehint.Intent) statehint.Prediction {
	i, _ := statehint.IntentIndex(intent)
	j, _ := statehint.IntentIndex(alternative)
	var p [8]float64
	p[i] = .97
	p[j] = .03
	return statehint.Prediction{Intent: intent, Probabilities: p, Confidence: p[i], Margin: p[i] - p[j], Source: statehint.Learned, TrainingSteps: 1}
}
func uniform() statehint.Prediction {
	p := statehint.Prediction{Intent: statehint.Unclear, Confidence: .125, Source: statehint.Learned, TrainingSteps: 1}
	for i := range p.Probabilities {
		p.Probabilities[i] = .125
	}
	return p
}
func closeTo(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("got %.16g want %.16g", got, want)
	}
}

func TestHandComputedFamilyCostLossAndConfusion(t *testing.T) {
	c := corpusFixture(t, []statehint.Intent{statehint.CompletionReport, statehint.Question}, func(i int) string { return fmt.Sprintf("lineage-%d", i) })
	correct := confident(statehint.CompletionReport, statehint.Question)
	wrongQuestion := confident(statehint.Question, statehint.CompletionReport)
	wrongCompletion := confident(statehint.CompletionReport, statehint.Question)
	p := []statehint.Prediction{correct, wrongQuestion, wrongCompletion, wrongCompletion}
	r, err := Evaluate(c, p)
	if err != nil {
		t.Fatal(err)
	}
	// First family: false question3 + one missed row/2. Second: false completion10 + both missed/2.
	closeTo(t, r.SeverityCost, (3.5+11)/2)
	closeTo(t, r.EightNLL, (-math.Log(.97)-3*math.Log(.03))/4)
	closeTo(t, r.FourNLL, r.EightNLL)
	closeTo(t, r.EightBrier, (.0018+3*1.8818)/4)
	closeTo(t, r.FourBrier, r.EightBrier)
	if r.FamilyErrors != [3]int{0, 1, 1} || r.DeclaredLineageErrors != [3]int{0, 1, 1} || r.GatedProposals != 4 || r.CorrectGatedProposals != 1 || r.MissedTargetRows != 3 || r.RawEightConfusion[4][4] != 1 || r.RawEightConfusion[4][0] != 1 || r.RawEightConfusion[0][4] != 2 || r.Eligible {
		t.Fatalf("count contract: %+v", r)
	}
	closeTo(t, r.CompletionPrecision, 1.0/3)
	if !r.CompletionPrecisionDefined || r.Locales[0].CorrectFamilies[1] != 1 || r.Locales[1].CorrectFamilies[1] != 0 || r.Stability.OneCorrectGatedFamilies != 1 || r.Stability.GatedAgreementFamilies != 1 {
		t.Fatal("locale/support/stability counts")
	}
}

func TestNoneProbabilityIsSumButGateUsesEightWinner(t *testing.T) {
	c := corpusFixture(t, []statehint.Intent{statehint.Reference}, func(int) string { return "lineage-none" })
	values := [8]float64{.31, .30, .20, .10, .09, 0, 0, 0}
	p := statehint.Prediction{Intent: statehint.Question, Probabilities: values, Confidence: values[0], Margin: values[0] - values[1], Source: statehint.Learned, TrainingSteps: 1}
	r, err := Evaluate(c, []statehint.Prediction{p, p})
	if err != nil {
		t.Fatal(err)
	}
	closeTo(t, r.EightNLL, -math.Log(.2))
	closeTo(t, r.FourNLL, -math.Log(.5))
	closeTo(t, r.EightBrier, .8442)
	closeTo(t, r.FourBrier, .3642)
	if r.RawFourConfusion[3][2] != 2 || r.GatedFourConfusion[3][3] != 2 || r.BelowConfidenceRows != 2 || r.SeverityCost != 0 {
		t.Fatal("none summation changed actual gate or created missing positives")
	}
}

func balancedFixture(t *testing.T, sameLineage bool) (statehintcorpus.Corpus, []statehint.Prediction) {
	var labels []statehint.Intent
	for _, label := range statehint.Intents() {
		for range 15 {
			labels = append(labels, label)
		}
	}
	c := corpusFixture(t, labels, func(i int) string {
		if sameLineage {
			return "one-lineage"
		}
		return fmt.Sprintf("lineage-%d", i%3)
	})
	var p []statehint.Prediction
	for _, label := range labels {
		index, _ := statehint.IntentIndex(label)
		value := confident(label, statehint.Intents()[(index+1)%8])
		p = append(p, value, value)
	}
	return c, p
}

func TestEligibilityAndAllAbstainDoNotPretendUtility(t *testing.T) {
	c, p := balancedFixture(t, false)
	r, err := Evaluate(c, p)
	if err != nil || !r.Eligible || len(r.EligibilityFailures) != 0 || r.SeverityCost != 0 || r.DeclaredLineages != 3 {
		t.Fatalf("eligible metadata fixture: %+v %v", r, err)
	}
	for _, locale := range r.Locales {
		if locale.CorrectFamilies != [3]int{15, 15, 15} || locale.CorrectDeclaredLineages != [3]int{3, 3, 3} || locale.CorrectCoverage != [3]float64{1, 1, 1} {
			t.Fatal("support denominators")
		}
	}
	for i := range p {
		p[i] = uniform()
	}
	r, err = Evaluate(c, p)
	if err != nil || r.Eligible || r.CompletionPrecisionDefined {
		t.Fatal("all-abstain qualification")
	}
	closeTo(t, r.SeverityCost, .375)
	c, p = balancedFixture(t, true)
	r, err = Evaluate(c, p)
	if err != nil || r.Eligible || r.Locales[0].CorrectDeclaredLineages != [3]int{1, 1, 1} {
		t.Fatal("family IDs invented declared lineage support")
	}
}

func TestRuleAndUntrainedControlsAreDescriptiveOnly(t *testing.T) {
	c, p := balancedFixture(t, false)
	for i := range p {
		index, _ := statehint.IntentIndex(p[i].Intent)
		p[i].Probabilities = [8]float64{}
		p[i].Probabilities[index] = 1
		p[i].Confidence = 1
		p[i].Margin = 1
		p[i].Source = statehint.RuleSource
		p[i].TrainingSteps = 0
	}
	r, err := Evaluate(c, p)
	if err != nil || r.Eligible || r.RuleRows != 240 || r.GatedProposals != 0 || r.EightNLL != 0 || r.EightBrier != 0 || r.ProbabilityMetricsCalibrated {
		t.Fatal("onehot rule became a learned/calibrated gate")
	}
	closeTo(t, r.SeverityCost, .375)
	for i := range p {
		p[i] = uniform()
		p[i].Source = statehint.Untrained
		p[i].TrainingSteps = 0
		p[i].GuardReason = "no_word_content"
	}
	r, err = Evaluate(c, p)
	if err != nil || r.Eligible || r.UntrainedRows != 240 || r.GuardedRows != 240 {
		t.Fatal("untrained/nonword guards")
	}
}

func TestStrictPredictionFailureHasNoPartialOrPrivateReport(t *testing.T) {
	c := corpusFixture(t, []statehint.Intent{statehint.Question}, func(int) string { return "PRIVATE-LINEAGE-MARKER" })
	good := confident(statehint.Question, statehint.Reference)
	cases := []struct {
		name string
		edit func(*statehint.Prediction)
	}{
		{"nan", func(p *statehint.Prediction) { p.Probabilities[1] = math.NaN() }},
		{"inf", func(p *statehint.Prediction) { p.Confidence = math.Inf(1) }},
		{"negative", func(p *statehint.Prediction) { p.Probabilities[1] = -.01 }},
		{"sum", func(p *statehint.Prediction) { p.Probabilities[1] = .1 }},
		{"winner", func(p *statehint.Prediction) { p.Intent = statehint.Reference }},
		{"confidence", func(p *statehint.Prediction) { p.Confidence = .96 }},
		{"margin", func(p *statehint.Prediction) { p.Margin = .5 }},
		{"source", func(p *statehint.Prediction) { p.Source = "PRIVATE-MARKER" }},
		{"steps", func(p *statehint.Prediction) { p.TrainingSteps = 0 }},
		{"rule-step", func(p *statehint.Prediction) { p.Source = statehint.RuleSource }},
		{"soft-rule", func(p *statehint.Prediction) { p.Source = statehint.RuleSource; p.TrainingSteps = 0 }},
		{"guard", func(p *statehint.Prediction) { p.GuardReason = "PRIVATE-MARKER" }},
		{"nonword-score", func(p *statehint.Prediction) { p.GuardReason = "no_word_content" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := good
			tc.edit(&bad)
			r, err := Evaluate(c, []statehint.Prediction{good, bad})
			if err != ErrInput || !reflect.DeepEqual(r, Report{}) || strings.Contains(err.Error(), "PRIVATE") {
				t.Fatal("invalid score returned partial/private output")
			}
		})
	}
	for _, p := range [][]statehint.Prediction{nil, {good}, {good, good, good}} {
		if _, err := Evaluate(c, p); err != ErrInput {
			t.Fatal("prediction length")
		}
	}
	if _, err := Evaluate(statehintcorpus.Corpus{}, nil); err != ErrInput {
		t.Fatal("empty corpus")
	}
	p := []statehint.Prediction{good, good}
	before := append([]statehint.Prediction(nil), p...)
	r, err := Evaluate(c, p)
	if err != nil || !reflect.DeepEqual(p, before) {
		t.Fatal("mutated cached predictions")
	}
	b, _ := json.Marshal(r)
	for _, value := range []string{"PRIVATE", "fixture-", "Invented arithmetic", "Fictional bounded"} {
		if bytes.Contains(b, []byte(value)) {
			t.Fatal("report leaked source metadata")
		}
	}
}

func TestUnclearTieAndNLLFloor(t *testing.T) {
	c := corpusFixture(t, []statehint.Intent{statehint.Unclear}, func(int) string { return "tie-lineage" })
	p := uniform()
	r, err := Evaluate(c, []statehint.Prediction{p, p})
	if err != nil || r.RawEightConfusion[7][7] != 2 {
		t.Fatal("SDK uncertainty tie")
	}
	closeTo(t, r.EightNLL, math.Log(8))
	closeTo(t, r.FourNLL, -math.Log(.625))
	wrong := confident(statehint.Question, statehint.Reference)
	r, err = Evaluate(c, []statehint.Prediction{wrong, wrong})
	if err != nil || r.EightNLLClippedRows != 2 {
		t.Fatal("clipped NLL")
	}
	closeTo(t, r.EightNLL, -math.Log(NLLFloor))
}
