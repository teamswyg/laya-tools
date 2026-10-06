package statehintpilot

import (
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

func newMetrics() Metrics {
	m := Metrics{}
	for index, intent := range AllowedIntents() {
		m.Targets[index].Intent = intent
	}
	return m
}

func observe(c Case, p statehintcatalog.Proposal, snapshot statehintcatalog.Snapshot, issues []statehintcatalog.MappingIssue, options Options) Observation {
	o := Observation{CaseID: c.ID, Family: c.Family, Locale: c.Locale, Expected: c.Expected, Reason: "work_unavailable", StateReason: "disabled_in_display_pilot",
		PilotStatePlanningDisabled: true, MappingDiagnosticsChecked: true, MappingIssues: []MappingDiagnostic{}, AnnotationReason: "annotation_work_unavailable"}
	for _, issue := range issues {
		o.MappingIssues = append(o.MappingIssues, MappingDiagnostic{Reason: issue.Reason, Intent: issue.Intent})
	}
	for _, work := range snapshot.Works {
		if work.Ref == c.ID {
			o.RawReaderStatePlanningChecked = true
			o.RawReaderStatePlanningAvailable = work.StatePlanningAvailable
			break
		}
	}
	if p.Prediction == nil {
		return o
	}
	pred := *p.Prediction
	o.Observed, o.Confidence, o.Probability, o.Margin, o.Source, o.TrainingSteps = pred.Intent, pred.Confidence, pred.Probabilities, pred.Margin, pred.Source, pred.TrainingSteps
	if pred.GuardReason == "no_word_content" {
		o.Guard = pred.GuardReason
	} else if pred.GuardReason != "" {
		o.Guard = "custom_predictor_guard"
	}
	var existing bool
	if allowed(pred.Intent) {
		labelAvailable, emojiAvailable, foundExisting := annotationAvailability(snapshot, issues, c.ID, pred.Intent)
		o.AnnotationChecked = true
		o.CurrentLabelMissing, o.CurrentEmojiMissing = !labelAvailable, !emojiAvailable
		existing = foundExisting
		switch {
		case !labelAvailable && !emojiAvailable:
			o.AnnotationReason = "current_label_and_emoji_unavailable"
		case !labelAvailable:
			o.AnnotationReason = "current_label_unavailable"
		case !emojiAvailable:
			o.AnnotationReason = "current_emoji_unavailable"
		default:
			o.AnnotationReason = "current_annotation_available"
		}
	} else {
		o.AnnotationReason = "annotation_not_applicable"
	}
	switch {
	case pred.GuardReason != "":
		o.Reason = "prediction_guard"
	case pred.Source == statehint.Untrained:
		o.Reason = "untrained_model"
	case !allowed(pred.Intent):
		o.Reason = "outside_three_intent_scope"
	case pred.Confidence < options.MinConfidence:
		o.Reason = "below_confidence"
	case pred.Margin < options.MinMargin:
		o.Reason = "below_margin"
	default:
		o.Eligible = true
	}
	if !o.Eligible {
		return o
	}
	o.LabelProposalCount = len(p.LabelRefs)
	o.EmojiProposed = p.EmojiCode != ""
	o.ProposalCreated = o.LabelProposalCount > 0 || o.EmojiProposed
	o.WrongEligible = pred.Intent != c.Expected
	o.WrongProposal = o.ProposalCreated && o.WrongEligible
	o.ExistingAnnotationNoOp = !o.ProposalCreated && existing
	switch {
	case o.ProposalCreated:
		o.Reason = "proposal_created"
	case o.ExistingAnnotationNoOp:
		o.Reason = "existing_annotation_no_op"
	case o.CurrentLabelMissing && o.CurrentEmojiMissing:
		o.Reason = "eligible_without_current_annotation"
	default:
		o.Reason = "eligible_without_new_proposal"
	}
	return o
}

func annotationAvailability(snapshot statehintcatalog.Snapshot, issues []statehintcatalog.MappingIssue, ref string, intent statehint.Intent) (bool, bool, bool) {
	var work statehintcatalog.Work
	for _, candidate := range snapshot.Works {
		if candidate.Ref == ref {
			work = candidate
			break
		}
	}
	blocked := map[string]bool{}
	labelBindingBlocked, emojiBindingBlocked := false, false
	for _, issue := range issues {
		if issue.Ref != "" {
			blocked[issue.Ref] = true
		}
		if issue.Intent == intent && issue.Reason == "duplicate_intent_binding" {
			labelBindingBlocked = true
		}
		if issue.Intent == intent && issue.Reason == "ambiguous_emoji_binding" {
			emojiBindingBlocked = true
		}
	}
	labelAvailable, emojiAvailable, existing := false, false, false
	if !labelBindingBlocked {
		for _, binding := range snapshot.Catalog.Bindings {
			if binding.Intent != intent {
				continue
			}
			for _, labelRef := range binding.LabelRefs {
				if blocked[labelRef] {
					continue
				}
				for _, label := range snapshot.Catalog.Labels {
					if label.Ref != labelRef || !label.Active || label.GroupContainer {
						continue
					}
					labelAvailable = true
					for _, current := range work.ExistingOpaqueLabels {
						if current == labelRef {
							existing = true
						}
					}
				}
			}
		}
	}
	if !emojiBindingBlocked {
		for _, emoji := range snapshot.Catalog.Emojis {
			if emoji.Intent == intent {
				emojiAvailable = true
				if emoji.Code == work.CurrentEmoji {
					existing = true
				}
			}
		}
	}
	return labelAvailable, emojiAvailable, existing
}

func (m *Metrics) add(o Observation) {
	m.Rows++
	expected, _ := statehint.IntentIndex(o.Expected)
	m.ExpectedCounts[expected]++
	if index, ok := targetIndex(o.Expected); ok {
		m.Targets[index].Expected++
	}
	if o.Observed == "" {
		m.UnavailableWork++
		m.Abstained++
		return
	}
	m.Classified++
	observed, _ := statehint.IntentIndex(o.Observed)
	m.PredictedCounts[observed]++
	m.Confusion[expected][observed]++
	correct := o.Observed == o.Expected
	if correct {
		m.Correct++
	}
	if !o.Eligible {
		m.Abstained++
	}
	index, ok := targetIndex(o.Observed)
	if !ok {
		return
	}
	target := &m.Targets[index]
	target.Predicted++
	if correct {
		target.Correct++
	}
	if o.AnnotationChecked {
		m.AnnotationChecked++
		target.AnnotationChecked++
	}
	if o.Eligible {
		m.Eligible++
		target.Eligible++
		if correct {
			target.CorrectEligible++
		} else {
			m.WrongEligible++
			target.WrongEligible++
		}
		if !o.ProposalCreated {
			m.EligibleWithoutProposal++
		}
	}
	if o.ProposalCreated {
		m.ProposalsCreated++
		target.ProposalsCreated++
		if correct {
			target.CorrectProposals++
		} else {
			m.WrongProposals++
			target.WrongProposals++
		}
	}
	if o.ExistingAnnotationNoOp {
		m.ExistingAnnotationNoOp++
		target.ExistingAnnotationNoOp++
	}
	if o.CurrentLabelMissing {
		m.CurrentLabelMissing++
		target.CurrentLabelMissing++
	}
	if o.CurrentEmojiMissing {
		m.CurrentEmojiMissing++
		target.CurrentEmojiMissing++
	}
}
