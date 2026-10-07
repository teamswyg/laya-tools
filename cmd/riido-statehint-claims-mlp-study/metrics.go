// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"math"
	"math/rand/v2"
	"sort"
)

type localeReport struct {
	Rows                int           `json:"rows"`
	ReferenceCounts     [3][3]int     `json:"reference_counts_head_true_false_unknown"`
	ReferenceFamilies   [3][3]int     `json:"reference_family_support_head_true_false_unknown"`
	ReferenceLineages   [3][3]int     `json:"reference_lineage_support_head_true_false_unknown"`
	RawConfusion        [3][3][3]int  `json:"raw_confusion_head_target_winner"`
	MeanCE              float64       `json:"mean_three_head_cross_entropy"`
	HeadCE              [3]float64    `json:"mean_cross_entropy_by_head"`
	TargetCE            [3][3]float64 `json:"mean_cross_entropy_by_head_target"`
	TargetCEDefined     [3][3]bool    `json:"cross_entropy_by_head_target_defined"`
	UnknownTargetCE     float64       `json:"mean_unknown_target_cross_entropy"`
	UnknownTargetCount  int           `json:"unknown_target_count"`
	TrueProposed        [3]int        `json:"gated_true_proposed"`
	TrueCorrect         [3]int        `json:"gated_true_correct"`
	TrueFalsePositive   [3]int        `json:"gated_true_false_positive"`
	TrueOnFalse         [3]int        `json:"gated_true_on_false_reference"`
	TrueOnUnknown       [3]int        `json:"gated_true_on_unknown_reference"`
	TrueCorrectFamilies [3]int        `json:"gated_true_correct_family_support"`
	TrueCorrectLineages [3]int        `json:"gated_true_correct_lineage_support"`
	Precision           [3]float64    `json:"gated_true_precision"`
	PrecisionDefined    [3]bool       `json:"gated_true_precision_defined"`
	TrueCoverage        [3]float64    `json:"gated_true_coverage_rows"`
	UnknownReasons      [3][5]int     `json:"emitted_unknown_reasons"`
	ClippedTargets      [3]int        `json:"target_probabilities_clipped_at_floor"`
}
type evaluationReport struct {
	Rows                       int             `json:"rows"`
	MeanCE                     float64         `json:"mean_three_head_cross_entropy"`
	ProbabilityFloor           float64         `json:"cross_entropy_probability_floor"`
	ReasonOrder                [5]string       `json:"unknown_reason_order"`
	Locales                    [2]localeReport `json:"locales_ko_en"`
	CompletionTrueProposed     int             `json:"pooled_completion_gated_true_proposed"`
	CompletionTrueCorrect      int             `json:"pooled_completion_gated_true_correct"`
	CompletionFalsePositive    int             `json:"pooled_completion_gated_true_false_positive"`
	CompletionPrecision        float64         `json:"pooled_completion_gated_true_precision"`
	CompletionPrecisionDefined bool            `json:"pooled_completion_gated_true_precision_defined"`
}
type observation struct {
	prediction statehintclaims.Prediction
	score      statehintclaims.ScoreResult
}

func rowCE(r row, o observation) (float64, error) {
	var total float64
	for h, t := range r.labels {
		k, ok := statehintclaims.StateIndex(t)
		if !ok {
			return 0, errStudy
		}
		p := o.score.Probabilities[h][k]
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return 0, errStudy
		}
		total -= math.Log(math.Max(p, nllFloor)) / 3
	}
	return total, nil
}
func evaluate(p prepared, obs []observation) (evaluationReport, error) {
	if len(obs) != len(p.rows) || len(obs) == 0 {
		return evaluationReport{}, errStudy
	}
	r := evaluationReport{Rows: len(obs), ProbabilityFloor: nllFloor, ReasonOrder: [5]string{"semantic_unknown", "low_confidence", "low_margin", "untrained", "no_word_content"}}
	var groups [2][3][3][]string
	var correct [2][3][]string
	for i, input := range p.rows {
		l := localeIndex(input.Locale)
		report := &r.Locales[l]
		report.Rows++
		for h, head := range obs[i].prediction.Heads {
			k, ok := statehintclaims.StateIndex(input.labels[h])
			winner, wok := statehintclaims.StateIndex(head.Winner)
			_, sok := statehintclaims.StateIndex(head.State)
			prob := obs[i].score.Probabilities[h][k]
			if !ok || !wok || !sok || head.Head != statehintclaims.Head(h).String() || head.Probabilities != obs[i].score.Probabilities[h] || math.IsNaN(prob) || math.IsInf(prob, 0) || prob < 0 || prob > 1 {
				return evaluationReport{}, errStudy
			}
			ce := -math.Log(math.Max(prob, nllFloor))
			report.ReferenceCounts[h][k]++
			report.ReferenceFamilies[h][k]++
			groups[l][h][k] = append(groups[l][h][k], input.Group)
			report.RawConfusion[h][k][winner]++
			report.MeanCE += ce / 3
			report.HeadCE[h] += ce
			report.TargetCE[h][k] += ce
			if prob < nllFloor {
				report.ClippedTargets[h]++
			}
			if k == 2 {
				report.UnknownTargetCount++
				report.UnknownTargetCE += ce
			}
			if head.State == statehintclaims.Unknown {
				found := false
				for j, reason := range r.ReasonOrder {
					if head.UnknownReason == reason {
						report.UnknownReasons[h][j]++
						found = true
						break
					}
				}
				if !found {
					return evaluationReport{}, errStudy
				}
			} else if head.UnknownReason != "" {
				return evaluationReport{}, errStudy
			}
			if head.State == statehintclaims.True {
				report.TrueProposed[h]++
				if k == 0 {
					report.TrueCorrect[h]++
					report.TrueCorrectFamilies[h]++
					correct[l][h] = append(correct[l][h], input.Group)
				} else {
					report.TrueFalsePositive[h]++
					if k == 1 {
						report.TrueOnFalse[h]++
					} else {
						report.TrueOnUnknown[h]++
					}
				}
			}
		}
	}
	for l := range r.Locales {
		report := &r.Locales[l]
		if report.Rows == 0 {
			return evaluationReport{}, errStudy
		}
		r.MeanCE += report.MeanCE
		report.MeanCE /= float64(report.Rows)
		for h := range report.HeadCE {
			report.HeadCE[h] /= float64(report.Rows)
			report.TrueCoverage[h] = float64(report.TrueProposed[h]) / float64(report.Rows)
			report.TrueCorrectLineages[h] = unique(correct[l][h])
			if report.TrueProposed[h] > 0 {
				report.PrecisionDefined[h] = true
				report.Precision[h] = float64(report.TrueCorrect[h]) / float64(report.TrueProposed[h])
			}
			for k := range report.TargetCE[h] {
				report.ReferenceLineages[h][k] = unique(groups[l][h][k])
				if report.ReferenceCounts[h][k] > 0 {
					report.TargetCEDefined[h][k] = true
					report.TargetCE[h][k] /= float64(report.ReferenceCounts[h][k])
				}
			}
		}
		if report.UnknownTargetCount > 0 {
			report.UnknownTargetCE /= float64(report.UnknownTargetCount)
		}
		r.CompletionTrueProposed += report.TrueProposed[2]
		r.CompletionTrueCorrect += report.TrueCorrect[2]
		r.CompletionFalsePositive += report.TrueFalsePositive[2]
	}
	r.MeanCE /= float64(r.Rows)
	if r.CompletionTrueProposed > 0 {
		r.CompletionPrecisionDefined = true
		r.CompletionPrecision = float64(r.CompletionTrueCorrect) / float64(r.CompletionTrueProposed)
	}
	return r, nil
}

type pairedReport struct {
	Difference           float64    `json:"mlp16_minus_cold_linear_mean_three_head_ce"`
	StrataDifference     [2]float64 `json:"mean_delta_contrast_general"`
	ConfidenceInterval   [2]float64 `json:"paired_lineage_bootstrap_percentile_95pct_ci"`
	Lineages             [2]int     `json:"resampled_lineages_contrast_general"`
	Weights              [2]int     `json:"fixed_strata_weights_contrast_general"`
	Resamples            int        `json:"resamples"`
	Seed                 int64      `json:"seed"`
	Unit                 string     `json:"resampling_unit"`
	PercentileConvention string     `json:"percentile_convention"`
}
type lineageDelta struct {
	sum     float64
	rows    int
	stratum int
}

func paired(p prepared, s strata, candidate, comparator []observation) (pairedReport, error) {
	if len(candidate) != len(p.rows) || len(comparator) != len(p.rows) {
		return pairedReport{}, errStudy
	}
	groups := make([]lineageDelta, len(s.lineageStrata))
	for i, stratum := range s.lineageStrata {
		groups[i].stratum = stratum
	}
	for i, r := range p.rows {
		index, ok := s.byID[r.ID]
		if !ok || index < 0 || index >= len(s.rowLineage) || s.rowLineage[index] < 0 || s.rowLineage[index] >= len(groups) {
			return pairedReport{}, errStudy
		}
		a, e := rowCE(r, candidate[i])
		if e != nil {
			return pairedReport{}, e
		}
		b, e := rowCE(r, comparator[i])
		if e != nil {
			return pairedReport{}, e
		}
		g := &groups[s.rowLineage[index]]
		g.rows++
		g.sum += a - b
	}
	var values [2][]float64
	for _, g := range groups {
		if g.rows != 6 || g.stratum < 0 || g.stratum > 1 {
			return pairedReport{}, errStudy
		}
		values[g.stratum] = append(values[g.stratum], g.sum/float64(g.rows))
	}
	return bootstrap(values)
}
func bootstrap(values [2][]float64) (pairedReport, error) {
	r := pairedReport{Weights: [2]int{2, 1}, Resamples: bootstrapSamples, Seed: 1729, Unit: "paired_leakage_group_mean_of_six_rows", PercentileConvention: "linear_interpolation_at_q_times_n_minus_one_q_0.025_0.975"}
	for s, v := range values {
		if len(v) == 0 {
			return pairedReport{}, errStudy
		}
		r.Lineages[s] = len(v)
		for _, x := range v {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return pairedReport{}, errStudy
			}
			r.StrataDifference[s] += x
		}
		r.StrataDifference[s] /= float64(len(v))
	}
	r.Difference = (2*r.StrataDifference[0] + r.StrataDifference[1]) / 3
	random := rand.New(rand.NewPCG(1729, 1729^0x9e3779b97f4a7c15))
	draws := make([]float64, bootstrapSamples)
	for i := range draws {
		var means [2]float64
		for s, v := range values {
			for range len(v) {
				means[s] += v[random.IntN(len(v))]
			}
			means[s] /= float64(len(v))
		}
		draws[i] = (2*means[0] + means[1]) / 3
	}
	sort.Float64s(draws)
	r.ConfidenceInterval = [2]float64{interpolatedPercentile(draws, .025), interpolatedPercentile(draws, .975)}
	return r, nil
}

// sorted is nonempty, sorted and finite; q is in [0,1]. Bootstrap fixes its
// endpoints at .025/.975, interpolating at q*(N-1), never nearest rank.
func interpolatedPercentile(sorted []float64, q float64) float64 {
	at := q * float64(len(sorted)-1)
	low := int(at)
	high := min(low+1, len(sorted)-1)
	return sorted[low] + (at-float64(low))*(sorted[high]-sorted[low])
}

type progressReport struct {
	PrimaryCIUpperBelowZero   bool       `json:"primary_ci_upper_below_zero"`
	UnknownCELower            [2]bool    `json:"unknown_target_ce_lower_ko_en"`
	NoAddedTrueFalsePositives [2][3]bool `json:"no_increased_gated_true_false_positive_locale_head"`
	GeneralCENonIncreasing    bool       `json:"general_stratum_ce_delta_non_increasing"`
	ResearchProgress          bool       `json:"research_progress_all_frozen_conditions"`
}

func progress(delta pairedReport, candidate, comparator evaluationReport) progressReport {
	r := progressReport{PrimaryCIUpperBelowZero: delta.ConfidenceInterval[1] < 0}
	r.GeneralCENonIncreasing = delta.StrataDifference[1] <= 0
	r.ResearchProgress = r.PrimaryCIUpperBelowZero && r.GeneralCENonIncreasing
	for l := range r.UnknownCELower {
		a, b := candidate.Locales[l], comparator.Locales[l]
		r.UnknownCELower[l] = a.UnknownTargetCount > 0 && b.UnknownTargetCount == a.UnknownTargetCount && a.UnknownTargetCE < b.UnknownTargetCE
		r.ResearchProgress = r.ResearchProgress && r.UnknownCELower[l]
		for h := range r.NoAddedTrueFalsePositives[l] {
			r.NoAddedTrueFalsePositives[l][h] = a.TrueFalsePositive[h] <= b.TrueFalsePositive[h]
			r.ResearchProgress = r.ResearchProgress && r.NoAddedTrueFalsePositives[l][h]
		}
	}
	return r
}
