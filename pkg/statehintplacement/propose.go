package statehintplacement

import (
	"context"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

type Request struct {
	Expected   Anchor
	HeldText   string
	CommandRef string
}
type Adapter struct {
	Revisions RevisionReader
	Catalog   statehintcatalog.SnapshotReader
	Predictor statehintcatalog.Predictor
}
type MappingDiagnostic struct {
	Reason string           `json:"reason"`
	Intent statehint.Intent `json:"intent,omitempty"`
}
type Proposal struct {
	Mode               string                `json:"mode"`
	Placement          Validation            `json:"placement"`
	Prediction         *statehint.Prediction `json:"prediction,omitempty"`
	Eligible           bool                  `json:"eligible"`
	LabelRefs          []string              `json:"label_refs"`
	EmojiCode          string                `json:"emoji_code,omitempty"`
	NoOp               bool                  `json:"no_op"`
	Reason             string                `json:"reason"`
	CatalogConsistency string                `json:"catalog_consistency"`
	MappingIssues      []MappingDiagnostic   `json:"mapping_issues"`
	ActualVerified     bool                  `json:"actual_verified"`
	MutationExecuted   bool                  `json:"mutation_executed"`
	StateChange        bool                  `json:"state_change"`
}

func unavailable(reason string) Proposal {
	return Proposal{Mode: "shadow", Placement: Validation{Status: Unavailable, Reason: reason}, LabelRefs: []string{}, NoOp: true, Reason: reason, CatalogConsistency: "unqualified", MappingIssues: []MappingDiagnostic{}}
}

// Propose hashes already-held input, checks independent current-owner metadata,
// builds an annotation-only proposal, then re-reads the anchor before returning
// it. No automatic retries or writes occur. Revision equality cannot prove that
// separate catalog reads are atomic; the caller must revalidate before any
// future real UI/application write.
func (a Adapter) Propose(ctx context.Context, request Request) (Proposal, error) {
	if ctx == nil || a.Revisions == nil || a.Catalog == nil || a.Predictor == nil {
		return unavailable("adapter_unavailable"), ErrInput
	}
	if err := ctx.Err(); err != nil {
		return unavailable("context_cancelled"), err
	}
	if !validAnchor(request.Expected) || !bounded(request.CommandRef) {
		return unavailable("expected_metadata_unavailable"), nil
	}
	digest, err := TextDigest(request.HeldText)
	if err != nil {
		return unavailable("held_text_invalid"), err
	}
	if digest != request.Expected.TextSHA256 {
		result := unavailable("held_text_anchor_mismatch")
		result.Placement.Status = Stale
		return result, nil
	}
	expected := request.Expected
	current, err := a.Revisions.ReadAnchor(ctx, expected.Scope, expected.WorkRef, expected.ContentRef)
	if ctx.Err() != nil {
		return unavailable("context_cancelled"), ctx.Err()
	}
	if err != nil {
		if ctx.Err() != nil {
			return unavailable("context_cancelled"), ctx.Err()
		}
		return unavailable("current_anchor_read_unavailable"), nil
	}
	placement := ValidatePlacement(expected, current)
	if placement.Status != Matched {
		result := unavailable(placement.Reason)
		result.Placement = placement
		return result, nil
	}
	reader := &annotationReader{inner: a.Catalog}
	predictor := statehintcatalog.PredictorFunc(func(text string) (statehint.Prediction, error) {
		if ctx.Err() != nil {
			return statehint.Prediction{}, ctx.Err()
		}
		p, err := a.Predictor.Predict(text)
		if ctx.Err() != nil {
			return statehint.Prediction{}, ctx.Err()
		}
		return p, err
	})
	result, err := (statehintcatalog.Adapter{MinConfidence: .9, MinMargin: .05}).Propose(ctx, reader, predictor, expected.Scope, []statehintcatalog.Request{{WorkRef: expected.WorkRef, CommandRef: request.CommandRef, Text: request.HeldText}})
	if ctx.Err() != nil {
		return unavailable("context_cancelled"), ctx.Err()
	}
	if err != nil {
		if ctx.Err() != nil {
			return unavailable("context_cancelled"), ctx.Err()
		}
		return unavailable("catalog_or_prediction_unavailable"), nil
	}
	if result.MutationExecuted || len(result.Proposals) != 1 {
		return unavailable("invalid_catalog_proposal"), nil
	}
	p := result.Proposals[0]
	if p.WorkRef != expected.WorkRef || p.Prediction == nil || p.MutationExecuted || p.StateChange || p.Plan != nil {
		return unavailable("work_or_annotation_unavailable"), nil
	}
	current, err = a.Revisions.ReadAnchor(ctx, expected.Scope, expected.WorkRef, expected.ContentRef)
	if ctx.Err() != nil {
		return unavailable("context_cancelled"), ctx.Err()
	}
	if err != nil {
		if ctx.Err() != nil {
			return unavailable("context_cancelled"), ctx.Err()
		}
		return unavailable("final_anchor_read_unavailable"), nil
	}
	placement = ValidatePlacement(expected, current)
	if placement.Status != Matched {
		out := unavailable(placement.Reason)
		out.Placement = placement
		return out, nil
	}
	out := Proposal{Mode: "shadow", Placement: placement, Prediction: p.Prediction, LabelRefs: []string{}, NoOp: true, Reason: "outside_three_intent_scope", CatalogConsistency: "unqualified", MappingIssues: []MappingDiagnostic{}}
	for _, issue := range append(result.MappingIssues, reader.issues...) {
		out.MappingIssues = append(out.MappingIssues, MappingDiagnostic{Reason: issue.Reason, Intent: issue.Intent})
	}
	pred := *p.Prediction
	if pred.GuardReason != "" && pred.GuardReason != "no_word_content" {
		pred.GuardReason = "custom_predictor_guard"
	}
	out.Prediction = &pred
	switch {
	case pred.GuardReason != "":
		out.Reason = "prediction_guard"
	case pred.Source == statehint.Untrained:
		out.Reason = "untrained_model"
	case !allowed(pred.Intent):
	case pred.Confidence < .9:
		out.Reason = "below_confidence"
	case pred.Margin < .05:
		out.Reason = "below_margin"
	default:
		out.Eligible = true
		out.LabelRefs = p.LabelRefs
		out.EmojiCode = p.EmojiCode
		out.NoOp = len(p.LabelRefs) == 0 && p.EmojiCode == ""
		out.Reason = "annotation_proposed"
		if out.NoOp {
			out.Reason = "no_new_current_annotation"
		}
	}
	return out, nil
}

func allowed(intent statehint.Intent) bool {
	return intent == statehint.Progress || intent == statehint.CompletionReport || intent == statehint.Question
}

type annotationReader struct {
	inner  statehintcatalog.SnapshotReader
	issues []statehintcatalog.MappingIssue
}

func (r *annotationReader) Read(ctx context.Context, scope statehintcatalog.Scope, refs []string) (statehintcatalog.Snapshot, error) {
	if ctx.Err() != nil {
		return statehintcatalog.Snapshot{}, ctx.Err()
	}
	snapshot, err := r.inner.Read(ctx, scope, refs)
	if ctx.Err() != nil {
		return statehintcatalog.Snapshot{}, ctx.Err()
	}
	if err != nil {
		return statehintcatalog.Snapshot{}, ErrUnavailable
	}
	if err := statehintcatalog.ValidateSnapshot(snapshot, scope, refs); err != nil {
		return statehintcatalog.Snapshot{}, ErrUnavailable
	}
	// The snapshot's work and scope are validated before the display filter.
	// OwnerWorkRevision is never inserted into any older numeric-version field.
	filtered := snapshot
	filtered.Works = append([]statehintcatalog.Work(nil), snapshot.Works...)
	for i := range filtered.Works {
		filtered.Works[i].StatePlanningAvailable = false
	}
	filtered.Catalog.Bindings = make([]statehintcatalog.Binding, 0, 3)
	filtered.Catalog.Emojis = make([]statehint.EmojiCandidate, 0, 3)
	counts := map[string]int{}
	for _, binding := range snapshot.Catalog.Bindings {
		for _, ref := range binding.LabelRefs {
			counts[ref]++
		}
	}
	r.issues = nil
	for _, binding := range snapshot.Catalog.Bindings {
		if !allowed(binding.Intent) {
			continue
		}
		kept := make([]string, 0, len(binding.LabelRefs))
		for _, ref := range binding.LabelRefs {
			if counts[ref] != 1 {
				r.issues = append(r.issues, statehintcatalog.MappingIssue{Reason: "duplicate_label_ref", Ref: ref, Intent: binding.Intent})
			} else {
				kept = append(kept, ref)
			}
		}
		filtered.Catalog.Bindings = append(filtered.Catalog.Bindings, statehintcatalog.Binding{Intent: binding.Intent, LabelRefs: kept})
	}
	for _, emoji := range snapshot.Catalog.Emojis {
		if allowed(emoji.Intent) {
			filtered.Catalog.Emojis = append(filtered.Catalog.Emojis, emoji)
		}
	}
	return filtered, nil
}
