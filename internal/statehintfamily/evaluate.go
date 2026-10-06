// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintfamily evaluates cached scores by paired scenario family.
package statehintfamily

import (
	"errors"
	"math"
	"sort"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const (
	ConfidenceFloor      = .9
	MarginFloor          = .05
	NLLFloor             = 1e-15
	AllAbstainComparator = .375
)

var ErrInput = errors.New("family evaluation requires valid paired corpus and cached predictions")

// Arrays use progress, completion_report, question order; locales use ko,en.
type LocaleReport struct {
	Rows                    int        `json:"rows"`
	Targets                 [3]int     `json:"targets"`
	Proposed                [3]int     `json:"proposed"`
	Correct                 [3]int     `json:"correct"`
	CorrectFamilies         [3]int     `json:"correct_families"`
	CorrectDeclaredLineages [3]int     `json:"correct_declared_lineages"`
	Precision               [3]float64 `json:"precision"`
	PrecisionDefined        [3]bool    `json:"precision_defined"`
	CorrectCoverage         [3]float64 `json:"correct_coverage"`
	EightNLL                float64    `json:"eight_nll"`
	FourNLL                 float64    `json:"four_nll"`
	EightBrier              float64    `json:"eight_brier"`
	FourBrier               float64    `json:"four_brier"`
}

type StabilityReport struct {
	RawEightAgreementFamilies       int     `json:"raw_eight_agreement_families"`
	RawFourAgreementFamilies        int     `json:"raw_four_agreement_families"`
	GatedAgreementFamilies          int     `json:"gated_agreement_families"`
	BothRawCorrectFamilies          int     `json:"both_raw_correct_families"`
	BothCorrectGatedFamilies        int     `json:"both_correct_gated_families"`
	OneCorrectGatedFamilies         int     `json:"one_correct_gated_family"`
	MeanAbsoluteEightProbabilityGap float64 `json:"mean_absolute_eight_probability_gap"`
}

type Report struct {
	Rows                         int             `json:"rows"`
	Families                     int             `json:"families"`
	DeclaredLineages             int             `json:"declared_lineages"`
	Eligible                     bool            `json:"eligible"`
	EligibilityFailures          []string        `json:"eligibility_failures"`
	SeverityCost                 float64         `json:"severity_cost"`
	EightNLL                     float64         `json:"eight_nll"`
	FourNLL                      float64         `json:"four_nll"`
	EightBrier                   float64         `json:"eight_brier"`
	FourBrier                    float64         `json:"four_brier"`
	EightNLLClippedRows          int             `json:"eight_nll_clipped_rows"`
	FourNLLClippedRows           int             `json:"four_nll_clipped_rows"`
	RawEightConfusion            [8][8]int       `json:"raw_eight_confusion"`
	RawFourConfusion             [4][4]int       `json:"raw_four_confusion"`
	GatedFourConfusion           [4][4]int       `json:"gated_four_confusion"`
	Locales                      [2]LocaleReport `json:"locales_ko_en"`
	FamilyErrors                 [3]int          `json:"family_errors_progress_completion_question"`
	DeclaredLineageErrors        [3]int          `json:"declared_lineage_errors_progress_completion_question"`
	CompletionPrecision          float64         `json:"completion_precision"`
	CompletionPrecisionDefined   bool            `json:"completion_precision_defined"`
	GatedProposals               int             `json:"gated_proposals"`
	CorrectGatedProposals        int             `json:"correct_gated_proposals"`
	MissedTargetRows             int             `json:"missed_target_rows"`
	GuardedRows                  int             `json:"guarded_rows"`
	OutsideScopeRows             int             `json:"outside_scope_rows"`
	BelowConfidenceRows          int             `json:"below_confidence_rows"`
	BelowMarginRows              int             `json:"below_margin_rows"`
	LearnedRows                  int             `json:"learned_rows"`
	UntrainedRows                int             `json:"untrained_rows"`
	RuleRows                     int             `json:"rule_rows"`
	ProbabilityMetricsCalibrated bool            `json:"probability_metrics_calibrated"`
	Stability                    StabilityReport `json:"stability"`
}

type support struct {
	locale, class int
	lineage       string
}
type wrongLineage struct {
	class   int
	lineage string
}
type rowScore struct {
	target, raw, gate int
	correct           bool
	n8, n4, b8, b4    float64
}

func display(intent statehint.Intent) int {
	switch intent {
	case statehint.Progress:
		return 0
	case statehint.CompletionReport:
		return 1
	case statehint.Question:
		return 2
	default:
		return 3
	}
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

// Validate scores without changing them. The uncertainty tie follows the SDK.
func validPrediction(p statehint.Prediction) bool {
	switch p.Source {
	case statehint.Learned:
		if p.TrainingSteps == 0 {
			return false
		}
	case statehint.Untrained, statehint.RuleSource:
		if p.TrainingSteps != 0 {
			return false
		}
	default:
		return false
	}
	if p.GuardReason != "" && p.GuardReason != "no_word_content" {
		return false
	}
	winner, runner, sum := statehint.IntentCount-1, -1, 0.0
	for i, v := range p.Probabilities {
		if !finite(v) || v < 0 || v > 1 {
			return false
		}
		sum += v
		if v > p.Probabilities[winner] {
			winner = i
		}
	}
	if math.Abs(sum-1) > 1e-9 {
		return false
	}
	for i, v := range p.Probabilities {
		if i != winner && (runner < 0 || v > p.Probabilities[runner]) {
			runner = i
		}
	}
	index, known := statehint.IntentIndex(p.Intent)
	if !known || index != winner || !finite(p.Confidence) || !finite(p.Margin) || p.Confidence != p.Probabilities[winner] || p.Margin != p.Probabilities[winner]-p.Probabilities[runner] {
		return false
	}
	if p.Source == statehint.RuleSource {
		if p.GuardReason != "" || p.Confidence != 1 {
			return false
		}
		for i, v := range p.Probabilities {
			if (i == winner && v != 1) || (i != winner && v != 0) {
				return false
			}
		}
	}
	if p.GuardReason == "no_word_content" {
		for _, v := range p.Probabilities {
			if v != .125 {
				return false
			}
		}
	}
	return true
}

// Evaluate never predicts or trains. Predictions must be in original row order.
// It produces no raw text, identifiers, hashes or semantic/rights attestation.
// A qualification here does not verify the caller's full partition quotas.
func Evaluate(corpus statehintcorpus.Corpus, predictions []statehint.Prediction) (Report, error) {
	rows, pairs := corpus.Rows(), corpus.Pairs()
	if len(rows) == 0 || len(rows) > 2400 || len(rows) != 2*len(pairs) || len(predictions) != len(rows) {
		return Report{}, ErrInput
	}
	seen := make([]bool, len(rows))
	for _, pair := range pairs {
		for locale, index := range pair.Rows {
			if index < 0 || index >= len(rows) || seen[index] {
				return Report{}, ErrInput
			}
			seen[index] = true
			r := rows[index]
			if r.Family != pair.Family.ID || r.Lineage != pair.Family.Lineage || r.Partition != pair.Family.Partition || r.Expected != pair.Family.Expected || r.Locale != [2]string{"ko", "en"}[locale] || !validPrediction(predictions[index]) {
				return Report{}, ErrInput
			}
		}
	}
	report := Report{Rows: len(rows), Families: len(pairs), EligibilityFailures: []string{}}
	scores := make([]rowScore, len(rows))
	supports := make([]support, 0, len(rows))
	wrong := make([]wrongLineage, 0, 3*len(pairs))
	lineages := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		lineages = append(lineages, pair.Family.Lineage)
		var anyWrong [3]bool
		missed, bothCorrect, rawBothCorrect := 0, 0, 0
		for locale, index := range pair.Rows {
			r, p := rows[index], predictions[index]
			target, _ := statehint.IntentIndex(r.Expected)
			raw, _ := statehint.IntentIndex(p.Intent)
			s := rowScore{target: display(r.Expected), raw: display(p.Intent), gate: 3}
			switch p.Source {
			case statehint.Learned:
				report.LearnedRows++
			case statehint.Untrained:
				report.UntrainedRows++
			case statehint.RuleSource:
				report.RuleRows++
			}
			switch {
			case p.GuardReason != "":
				report.GuardedRows++
			case p.Source != statehint.Learned:
			case s.raw == 3:
				report.OutsideScopeRows++
			case p.Confidence < ConfidenceFloor:
				report.BelowConfidenceRows++
			case p.Margin < MarginFloor:
				report.BelowMarginRows++
			default:
				s.gate = s.raw
			}
			var four [4]float64
			for i, v := range p.Probabilities {
				four[display(statehint.Intents()[i])] += v
				y := 0.0
				if i == target {
					y = 1
				}
				s.b8 += (v - y) * (v - y)
			}
			for i, v := range four {
				y := 0.0
				if i == s.target {
					y = 1
				}
				s.b4 += (v - y) * (v - y)
			}
			if p.Probabilities[target] < NLLFloor {
				report.EightNLLClippedRows++
			}
			if four[s.target] < NLLFloor {
				report.FourNLLClippedRows++
			}
			s.n8 = -math.Log(math.Max(p.Probabilities[target], NLLFloor))
			s.n4 = -math.Log(math.Max(four[s.target], NLLFloor))
			report.RawEightConfusion[target][raw]++
			report.RawFourConfusion[s.target][s.raw]++
			report.GatedFourConfusion[s.target][s.gate]++
			if raw == target {
				rawBothCorrect++
			}
			l := &report.Locales[locale]
			l.Rows++
			l.EightNLL += s.n8
			l.FourNLL += s.n4
			l.EightBrier += s.b8
			l.FourBrier += s.b4
			if s.target < 3 {
				l.Targets[s.target]++
			}
			if s.gate < 3 {
				report.GatedProposals++
				l.Proposed[s.gate]++
				if s.gate == s.target {
					s.correct = true
					bothCorrect++
					report.CorrectGatedProposals++
					l.Correct[s.gate]++
					l.CorrectFamilies[s.gate]++
					supports = append(supports, support{locale, s.gate, pair.Family.Lineage})
				} else {
					anyWrong[s.gate] = true
				}
			}
			if s.target < 3 && !s.correct {
				missed++
				report.MissedTargetRows++
			}
			scores[index] = s
		}
		for class, yes := range anyWrong {
			if yes {
				report.FamilyErrors[class]++
				wrong = append(wrong, wrongLineage{class, pair.Family.Lineage})
			}
		}
		report.SeverityCost += 10*boolFloat(anyWrong[1]) + 3*boolFloat(anyWrong[2]) + boolFloat(anyWrong[0]) + float64(missed)/2
		a, b := scores[pair.Rows[0]], scores[pair.Rows[1]]
		report.EightNLL += (a.n8 + b.n8) / 2
		report.FourNLL += (a.n4 + b.n4) / 2
		report.EightBrier += (a.b8 + b.b8) / 2
		report.FourBrier += (a.b4 + b.b4) / 2
		if predictions[pair.Rows[0]].Intent == predictions[pair.Rows[1]].Intent {
			report.Stability.RawEightAgreementFamilies++
		}
		if a.raw == b.raw {
			report.Stability.RawFourAgreementFamilies++
		}
		if a.gate == b.gate {
			report.Stability.GatedAgreementFamilies++
		}
		if rawBothCorrect == 2 {
			report.Stability.BothRawCorrectFamilies++
		}
		if bothCorrect == 2 {
			report.Stability.BothCorrectGatedFamilies++
		} else if bothCorrect == 1 {
			report.Stability.OneCorrectGatedFamilies++
		}
		for i, v := range predictions[pair.Rows[0]].Probabilities {
			report.Stability.MeanAbsoluteEightProbabilityGap += math.Abs(v-predictions[pair.Rows[1]].Probabilities[i]) / 8
		}
	}
	n := float64(len(pairs))
	report.SeverityCost /= n
	report.EightNLL /= n
	report.FourNLL /= n
	report.EightBrier /= n
	report.FourBrier /= n
	report.Stability.MeanAbsoluteEightProbabilityGap /= n
	sort.Strings(lineages)
	for i, v := range lineages {
		if i == 0 || v != lineages[i-1] {
			report.DeclaredLineages++
		}
	}
	sort.Slice(supports, func(i, j int) bool {
		a, b := supports[i], supports[j]
		if a.locale != b.locale {
			return a.locale < b.locale
		}
		if a.class != b.class {
			return a.class < b.class
		}
		return a.lineage < b.lineage
	})
	for i, v := range supports {
		if i == 0 || v != supports[i-1] {
			report.Locales[v.locale].CorrectDeclaredLineages[v.class]++
		}
	}
	sort.Slice(wrong, func(i, j int) bool {
		if wrong[i].class != wrong[j].class {
			return wrong[i].class < wrong[j].class
		}
		return wrong[i].lineage < wrong[j].lineage
	})
	for i, v := range wrong {
		if i == 0 || v != wrong[i-1] {
			report.DeclaredLineageErrors[v.class]++
		}
	}
	for locale := range report.Locales {
		l := &report.Locales[locale]
		d := float64(l.Rows)
		l.EightNLL /= d
		l.FourNLL /= d
		l.EightBrier /= d
		l.FourBrier /= d
		for class := range l.Targets {
			if l.Proposed[class] > 0 {
				l.PrecisionDefined[class] = true
				l.Precision[class] = float64(l.Correct[class]) / float64(l.Proposed[class])
			}
			if l.Targets[class] > 0 {
				l.CorrectCoverage[class] = float64(l.Correct[class]) / float64(l.Targets[class])
			}
		}
	}
	complete := report.Locales[0].Proposed[1] + report.Locales[1].Proposed[1]
	if complete > 0 {
		report.CompletionPrecisionDefined = true
		report.CompletionPrecision = float64(report.Locales[0].Correct[1]+report.Locales[1].Correct[1]) / float64(complete)
	}
	if report.LearnedRows != report.Rows {
		report.EligibilityFailures = append(report.EligibilityFailures, "nonlearned_predictions")
	}
	for locale, l := range report.Locales {
		for class := range l.Targets {
			prefix := [2]string{"ko_", "en_"}[locale] + [3]string{"progress_", "completion_report_", "question_"}[class]
			if l.Targets[class] == 0 || l.CorrectCoverage[class] < .2 {
				report.EligibilityFailures = append(report.EligibilityFailures, prefix+"correct_coverage")
			}
			if l.CorrectFamilies[class] < 5 {
				report.EligibilityFailures = append(report.EligibilityFailures, prefix+"correct_family_support")
			}
			if l.CorrectDeclaredLineages[class] < 2 {
				report.EligibilityFailures = append(report.EligibilityFailures, prefix+"correct_declared_lineage_support")
			}
		}
	}
	if !report.CompletionPrecisionDefined || report.CompletionPrecision < .98 {
		report.EligibilityFailures = append(report.EligibilityFailures, "completion_precision")
	}
	if report.SeverityCost >= AllAbstainComparator {
		report.EligibilityFailures = append(report.EligibilityFailures, "all_abstain_comparator")
	}
	report.Eligible = len(report.EligibilityFailures) == 0
	return report, nil
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
