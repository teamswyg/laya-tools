package statehintcatalog

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

// Propose reads metadata for explicit opaque references, then classifies
// caller-supplied text. All output is shadow-only. This bridge deliberately
// accepts no verified events or user-provided trusted flags. Therefore no
// returned proposal can execute or authorize a state change.
func (a Adapter) Propose(ctx context.Context, reader SnapshotReader, predictor Predictor, scope Scope, requests []Request) (Result, error) {
	if ctx == nil || reader == nil || predictor == nil || !validRef(scope.WorkspaceRef) || !validRef(scope.OwnerRef) ||
		len(requests) == 0 || len(requests) > MaxWorks {
		return Result{}, ErrInput
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if a.MinConfidence == 0 {
		a.MinConfidence = .9
	}
	if a.MinMargin == 0 {
		a.MinMargin = .05
	}
	if !finite(a.MinConfidence) || a.MinConfidence < .9 || a.MinConfidence > 1 || !finite(a.MinMargin) || a.MinMargin < 0 || a.MinMargin > 1 {
		return Result{}, ErrInput
	}
	requested := map[string]bool{}
	commands := map[string]bool{}
	refs := make([]string, 0, len(requests))
	for _, request := range requests {
		if !validRef(request.WorkRef) || !validRef(request.CommandRef) || requested[request.WorkRef] || commands[request.CommandRef] ||
			len(request.Text) > statehint.MaxTextBytes || !utf8.ValidString(request.Text) || strings.ContainsRune(request.Text, 0) {
			return Result{}, ErrInput
		}
		requested[request.WorkRef], commands[request.CommandRef] = true, true
		refs = append(refs, request.WorkRef)
	}
	snapshot, err := reader.Read(ctx, scope, refs)
	if err != nil {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		// Reader errors can contain native IDs or credentials. Those details
		// belong to the private reader's logging boundary, not this response.
		return Result{}, ErrSnapshot
	}
	if err := validateSnapshot(snapshot, scope, requested); err != nil {
		return Result{}, err
	}
	labels, issues := resolveLabels(snapshot.Catalog)
	emojis, emojiIssues := resolveEmojis(snapshot.Catalog)
	issues = append(issues, emojiIssues...)
	if issues == nil {
		issues = []MappingIssue{}
	}
	result := Result{Mode: "shadow", CatalogConsistency: "unqualified", Scope: scope, ReadTimes: snapshot.ReadTimes,
		Proposals: make([]Proposal, 0, len(requests)), MappingIssues: issues}
	works := map[string]Work{}
	for _, work := range snapshot.Works {
		works[work.Ref] = work
	}
	for _, request := range requests {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		proposal := Proposal{WorkRef: request.WorkRef, CommandRef: request.CommandRef, LabelRefs: []string{}, NoOp: true, StateReason: "work_unavailable"}
		work, found := works[request.WorkRef]
		if !found {
			result.Proposals = append(result.Proposals, proposal)
			continue
		}
		p, err := predictor.Predict(request.Text)
		if err != nil || !validPrediction(p) {
			return Result{}, ErrPredictor
		}
		proposal.Prediction = &p
		proposal.CurrentState, proposal.NextState = work.State, work.State
		if eligible(p, a) {
			proposal.LabelRefs, proposal.EmojiCode = annotations(work, labels, emojis, p.Intent)
		}
		proposal.NoOp = len(proposal.LabelRefs) == 0 && proposal.EmojiCode == ""
		version, canonical := canonicalVersion(work.CanonicalVersion)
		if canonical {
			proposal.CanonicalVersion = work.CanonicalVersion
		}
		switch {
		case p.GuardReason != "":
			proposal.StateReason = "unavailable_prediction_guard"
		case work.CanonicalVersion == "":
			proposal.StateReason = "unavailable_no_canonical_version"
		case !canonical:
			proposal.StateReason = "unavailable_invalid_canonical_version"
		case !work.StatePlanningAvailable:
			proposal.StateReason = "unavailable_reader_state_planning"
		case !validState(work.State):
			proposal.StateReason = "unavailable_current_state"
		default:
			ref, statusReason := resolveStatus(snapshot.Catalog, targetState(p.Intent))
			if targetState(p.Intent) != "" && ref == "" {
				proposal.StateReason = statusReason
				break
			}
			// Only canonical source-provided versions reach the Plan call.
			// No placeholder revision is invented for annotations-only work.
			plan, err := statehint.Plan(statehint.PlanInput{WorkID: work.Ref, CommandID: request.CommandRef, ExpectedVersion: version,
				CurrentState: work.State, Labels: labels, ExistingLabelIDs: work.ExistingOpaqueLabels, Emojis: emojis,
				CurrentEmojiCode: work.CurrentEmoji, MinConfidence: a.MinConfidence, MinMargin: a.MinMargin}, p)
			if err != nil {
				return Result{}, ErrSnapshot
			}
			proposal.Plan = &plan
			proposal.LabelRefs, proposal.EmojiCode = plan.LabelIDs, plan.EmojiCode
			proposal.NextState, proposal.NoOp = plan.NextState, plan.NoOp
			proposal.StatePlanningAvailable = true
			proposal.StateReason = plan.Reason
			if eligible(p, a) && plan.Reason != "terminal_state_preserved" {
				proposal.CandidateStatusRef = ref
			}
		}
		result.Proposals = append(result.Proposals, proposal)
	}
	return result, nil
}

func eligible(p statehint.Prediction, a Adapter) bool {
	return p.Source != statehint.Untrained && p.Intent != statehint.Unclear && p.GuardReason == "" && p.Confidence >= a.MinConfidence && p.Margin >= a.MinMargin
}

func annotations(work Work, labels []statehint.CatalogLabel, emojis []statehint.EmojiCandidate, intent statehint.Intent) ([]string, string) {
	refs := make([]string, 0)
	for _, label := range labels {
		if label.Intent != intent {
			continue
		}
		existing := false
		for _, ref := range work.ExistingOpaqueLabels {
			if ref == label.ID {
				existing = true
				break
			}
		}
		if !existing {
			refs = append(refs, label.ID)
		}
	}
	for _, emoji := range emojis {
		if emoji.Intent == intent && emoji.Code != work.CurrentEmoji {
			return refs, emoji.Code
		}
	}
	return refs, ""
}
