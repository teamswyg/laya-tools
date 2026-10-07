// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"math"
	"strings"
	"testing"
)

func TestDigestLengthGuardHasNoOversizedDecodeAllocation(t *testing.T) {
	oversized := strings.Repeat("a", 1<<20)
	if validSHA(oversized) || validSHA(strings.Repeat("A", 64)) || validSHA(strings.Repeat("g", 64)) || !validSHA(strings.Repeat("a", 64)) {
		t.Fatal("digest admission changed")
	}
	if allocations := testing.AllocsPerRun(10, func() {
		if validSHA(oversized) {
			panic("oversized digest")
		}
	}); allocations != 0 {
		t.Fatal("oversized digest decoded before bound", allocations)
	}
}

func TestStrictRowsGroupingAndUTF8Evidence(t *testing.T) {
	good := ownedRows("fit", 3)
	if p, e := prepare(encodeRows(t, good), "fit"); e != nil || p.counts != (counts{Rows: 6, Families: 3, Lineages: 1, Locales: [2]int{3, 3}}) {
		t.Fatal("owned grouping", e)
	}
	mutations := []func([]rawRow){
		func(r []rawRow) { r[0].Targets[0] = "True" }, func(r []rawRow) { r[0].Targets = r[0].Targets[:2] }, func(r []rawRow) { r[0].Evidence = r[0].Evidence[:2] }, func(r []rawRow) { r[0].Evidence[0].Start = nil }, func(r []rawRow) { r[0].Evidence[0].Start = number(1) }, func(r []rawRow) { r[0].Evidence[0].End = number(-1) }, func(r []rawRow) { r[0].Evidence[0].End = number(len(r[0].Text) + 1) }, func(r []rawRow) { r[0].Evidence[0].End = number(0) }, func(r []rawRow) { r[0].Evidence[0].Reason = "" }, func(r []rawRow) { r[0].Evidence[1].Start = number(3); r[0].Evidence[1].End = number(3) }, func(r []rawRow) { r[0].Text = " \t" }, func(r []rawRow) { r[0].Text = "!!!" }, func(r []rawRow) { r[0].Text = string([]byte{0xff}) }, func(r []rawRow) { r[0].Schema = "alternate" }, func(r []rawRow) { r[0].AnnotationSource = "human_verified" }, func(r []rawRow) { r[0].Split = "dev" }, func(r []rawRow) { r[0].RubricSHA = strings.Repeat("c", 64) }, func(r []rawRow) { r[0].ID = r[1].ID }, func(r []rawRow) { r[0].Group = "other-group" }, func(r []rawRow) { r[0].Locale = "fr" }, func(r []rawRow) { r[0].Family = "case/private" },
	}
	for i, mutate := range mutations {
		rows := ownedRows("fit", 3)
		mutate(rows)
		if _, e := prepare(encodeRows(t, rows), "fit"); e == nil {
			t.Fatalf("accepted malformed case %d", i)
		}
	}
	data := encodeRows(t, good)
	for _, bad := range [][]byte{bytes.Replace(data, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1), bytes.Replace(data, []byte(`"schema":`), []byte(`"Schema":`), 1), bytes.Replace(data, []byte(`"schema":`), []byte(`"extra":0,"schema":`), 1), append(data, '\n'), append([]byte{0xff}, data...), append([]byte("null\n"), data...)} {
		if _, e := prepare(bad, "fit"); e == nil {
			t.Fatal("accepted noncanonical JSONL")
		}
	}
	if _, e := prepare(encodeRows(t, good[:5]), "fit"); e == nil {
		t.Fatal("unpaired family accepted")
	}
	if _, e := prepare(encodeRows(t, ownedRows("fit", 4)), "fit"); e == nil {
		t.Fatal("unequal lineage size accepted")
	}
	// Paired locale membership does not imply identical wording or references.
	different := ownedRows("fit", 3)
	different[1].Targets[0] = "unknown"
	different[1].Evidence[0].End = number(len(different[1].Text))
	paired, e := prepare(encodeRows(t, different), "fit")
	if e != nil || paired.rows[0].labels[0] == paired.rows[1].labels[0] {
		t.Fatal("independent locale reference differences rejected", e)
	}
}

func TestStrataReceiptAndIndependentSplits(t *testing.T) {
	f := ownedFixture(t)
	train, e := prepare(f.training, "fit")
	if e != nil {
		t.Fatal(e)
	}
	s, e := prepareStrata(f.metadata, train)
	if e != nil {
		t.Fatal(e)
	}
	r, e := validateReceipt(f.receipt, digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s)
	if e != nil {
		t.Fatal(e)
	}
	p, e := prepare(f.evaluation, "dev")
	if e != nil || validateEvaluation(p, train, s, r) != nil {
		t.Fatal("valid owned evaluation rejected", e)
	}
	var meta strataInput
	if json.Unmarshal(f.metadata, &meta) != nil {
		t.Fatal("owned metadata")
	}
	meta.Rows[0] = stratumRow{train.rows[0].ID, train.rows[0].Family, train.rows[0].Group, train.rows[0].Locale, "contrast"}
	if _, e := prepareStrata(encode(t, meta), train); e == nil {
		t.Fatal("training identity leaked across split")
	}
	if json.Unmarshal(f.metadata, &meta) != nil {
		t.Fatal("owned metadata")
	}
	meta.Rows[0].Stratum = "general"
	if _, e := prepareStrata(encode(t, meta), train); e == nil {
		t.Fatal("stratum split within lineage accepted")
	}
	for _, change := range []func(*receipt){func(r *receipt) { r.ContrastFamilies[0][0][2] = 19 }, func(r *receipt) { r.ContrastLineages[1][2][0] = 9 }, func(r *receipt) { r.States[0] = statehintclaims.Unknown }, func(r *receipt) { r.Heads[0] = "completion_claimed" }, func(r *receipt) { r.EvaluationSHA = strings.Repeat("0", 64) }, func(r *receipt) { r.EvaluationBytes = 0 }, func(r *receipt) { r.ReferenceStatus = "human_verified" }} {
		r := f.refs
		change(&r)
		if _, e := validateReceipt(encode(t, r), digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s); e == nil {
			t.Fatal("invalid receipt accepted")
		}
	}
	badArrays := bytes.Replace(f.receipt, []byte(`"locale_order":["ko","en"]`), []byte(`"locale_order":["ko","en","ko"]`), 1)
	if _, e := validateReceipt(badArrays, digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s); e == nil {
		t.Fatal("extra canonical-array entries accepted")
	}
	p.rows[0].labels[0] = statehintclaims.Unknown
	if validateEvaluation(p, train, s, r) == nil {
		t.Fatal("evaluation support differs from writer receipt")
	}
}

func TestLineageBootstrapUsesFixedWeightAndPairedLocales(t *testing.T) {
	values := [2][]float64{{-2, -2}, {1, 1}}
	r, e := bootstrap(values)
	if e != nil || r.Difference != -1 || r.ConfidenceInterval != [2]float64{-1, -1} || r.Weights != [2]int{2, 1} || r.Resamples != 10000 || r.Seed != 1729 || r.PercentileConvention != "linear_interpolation_at_q_times_n_minus_one_q_0.025_0.975" {
		t.Fatal("fixed bootstrap convention/weight changed", r, e)
	}
	varying := [2][]float64{{-3, -1}, {1, 4}}
	a, e := bootstrap(varying)
	b, e2 := bootstrap(varying)
	if e != nil || e2 != nil || a != b || a.Difference != -0.5 || a.ConfidenceInterval != [2]float64{-5.0 / 3, 2.0 / 3} {
		t.Fatal("deterministic known group distribution changed", a, e, e2)
	}
	f := ownedFixture(t)
	train, _ := prepare(f.training, "fit")
	s, _ := prepareStrata(f.metadata, train)
	p, _ := prepare(f.evaluation, "dev")
	candidate := make([]observation, len(p.rows))
	comparator := make([]observation, len(p.rows))
	for i, r := range p.rows {
		delta := -2.0
		if s.rows[s.byID[r.ID]].Stratum == "general" {
			delta = 1
		}
		if r.Locale == "ko" {
			delta += .75
		} else {
			delta -= .75
		}
		for h, target := range r.labels {
			k, _ := statehintclaims.StateIndex(target)
			candidate[i].score.Probabilities[h][k] = math.Exp(-(4 + delta))
			comparator[i].score.Probabilities[h][k] = math.Exp(-4)
		}
	}
	grouped, e := paired(p, s, candidate, comparator)
	if e != nil || math.Abs(grouped.Difference+1) > 1e-12 || math.Abs(grouped.ConfidenceInterval[0]+1) > 1e-12 || math.Abs(grouped.ConfidenceInterval[1]+1) > 1e-12 || grouped.Lineages != [2]int{40, 20} {
		t.Fatal("cross-language rows were resampled separately", grouped, e)
	}
	if _, e := paired(prepared{rows: p.rows[:len(p.rows)-1]}, s, candidate[:len(candidate)-1], comparator[:len(comparator)-1]); e == nil {
		t.Fatal("incomplete paired group accepted")
	}
	if _, e := bootstrap([2][]float64{{math.NaN()}, {0}}); e == nil {
		t.Fatal("nonfinite bootstrap accepted")
	}
}

func TestPercentileConventionInterpolatesBetweenSortedDraws(t *testing.T) {
	sorted := []float64{0, 10, 20, 30}
	if got := interpolatedPercentile(sorted, .025); math.Abs(got-.75) > 1e-12 {
		t.Fatal("lower percentile rounded to an array entry", got)
	}
	if got := interpolatedPercentile(sorted, .975); math.Abs(got-29.25) > 1e-12 {
		t.Fatal("upper percentile rounded to an array entry", got)
	}
	if interpolatedPercentile(sorted, 0) != 0 || interpolatedPercentile(sorted, 1) != 30 || interpolatedPercentile([]float64{7}, .025) != 7 {
		t.Fatal("percentile endpoint/singleton changed")
	}
}

func TestUnknownCEAndFalsePositiveProgressConditions(t *testing.T) {
	candidate, comparator := evaluationReport{}, evaluationReport{}
	for l := 0; l < 2; l++ {
		candidate.Locales[l].UnknownTargetCount = 1
		candidate.Locales[l].UnknownTargetCE = 1
		comparator.Locales[l].UnknownTargetCount = 1
		comparator.Locales[l].UnknownTargetCE = 2
	}
	delta := pairedReport{ConfidenceInterval: [2]float64{-1, -.1}}
	if !progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("all conditions should pass")
	}
	candidate.Locales[1].TrueFalsePositive[2] = 1
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("one locale/head FP increase ignored")
	}
	candidate.Locales[1].TrueFalsePositive[2] = 0
	candidate.Locales[0].UnknownTargetCE = 2
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("equal unknown CE counted as improvement")
	}
	candidate.Locales[0].UnknownTargetCount = 0
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("undefined unknown CE counted as improvement")
	}
}

func TestAggregateFalsePositiveIncludesFalseAndUnknownReferences(t *testing.T) {
	p, e := prepare(encodeRows(t, ownedRows("dev", 3)), "dev")
	if e != nil {
		t.Fatal(e)
	}
	obs := make([]observation, len(p.rows))
	for i := range obs {
		for h := 0; h < 3; h++ {
			prob := [3]float64{.95, .025, .025}
			obs[i].score.Probabilities[h] = prob
			obs[i].prediction.Heads[h] = statehintclaims.HeadPrediction{Head: statehintclaims.Head(h).String(), Winner: statehintclaims.True, State: statehintclaims.True, Probabilities: prob}
		}
	}
	r, e := evaluate(p, obs)
	if e != nil {
		t.Fatal(e)
	}
	for _, l := range r.Locales {
		if l.TrueProposed != [3]int{3, 3, 3} || l.TrueCorrect != [3]int{1, 1, 1} || l.TrueFalsePositive != [3]int{2, 2, 2} || l.TrueOnFalse != [3]int{1, 1, 1} || l.TrueOnUnknown != [3]int{1, 1, 1} || l.UnknownTargetCount != 3 {
			t.Fatal("false/unknown targets excluded from gated true FP", l)
		}
		if l.TrueCoverage != [3]float64{1, 1, 1} || l.Precision != [3]float64{1.0 / 3, 1.0 / 3, 1.0 / 3} {
			t.Fatal("precision/coverage denominators", l)
		}
	}
	if r.CompletionTrueProposed != 6 || r.CompletionTrueCorrect != 2 || r.CompletionFalsePositive != 4 || !r.CompletionPrecisionDefined || r.CompletionPrecision != 1.0/3 {
		t.Fatal("pooled completion counts/precision", r)
	}
}

func TestGeneralRegressionIndependentlyBlocksResearchProgress(t *testing.T) {
	delta := pairedReport{ConfidenceInterval: [2]float64{-2, -1}, StrataDifference: [2]float64{-2, .001}}
	var candidate, linear evaluationReport
	for l := range candidate.Locales {
		candidate.Locales[l].UnknownTargetCount = 1
		linear.Locales[l].UnknownTargetCount = 1
		candidate.Locales[l].UnknownTargetCE = 1
		linear.Locales[l].UnknownTargetCE = 2
	}
	r := progress(delta, candidate, linear)
	if !r.PrimaryCIUpperBelowZero || !r.UnknownCELower[0] || !r.UnknownCELower[1] || r.GeneralCENonIncreasing || r.ResearchProgress {
		t.Fatal("general regression escaped independent frozen condition", r)
	}
	delta.StrataDifference[1] = 0
	if r := progress(delta, candidate, linear); !r.GeneralCENonIncreasing || !r.ResearchProgress {
		t.Fatal("general equality rejected", r)
	}
}
