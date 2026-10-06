package statehint

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

type State string

const (
	Todo      State = "todo"
	Active    State = "active"
	Done      State = "done"
	Cancelled State = "cancelled"
)

type EventKind string

const (
	StartedEvent   EventKind = "started"
	CompletedEvent EventKind = "completed"
	CancelledEvent EventKind = "cancelled"
)

// CatalogLabel is caller-supplied current catalog evidence. IDs are opaque;
// no application catalog is embedded in the public classifier.
type CatalogLabel struct {
	ID     string `json:"id"`
	Intent Intent `json:"intent"`
	Active bool   `json:"active"`
}
type EmojiCandidate struct {
	Intent Intent `json:"intent"`
	Code   string `json:"code"`
}

// EventEvidence must come from a trusted runtime event source, not model text.
// Trusted is an explicit input trust boundary, not a cryptographic attestation.
type EventEvidence struct {
	ID      string    `json:"id"`
	Kind    EventKind `json:"kind"`
	WorkID  string    `json:"work_id"`
	Version uint64    `json:"version"`
	Trusted bool      `json:"trusted"`
}
type PlanInput struct {
	WorkID           string           `json:"work_id"`
	CommandID        string           `json:"command_id"`
	ExpectedVersion  uint64           `json:"expected_version"`
	CurrentState     State            `json:"current_state"`
	Labels           []CatalogLabel   `json:"labels"`
	ExistingLabelIDs []string         `json:"existing_label_ids"`
	Emojis           []EmojiCandidate `json:"emojis"`
	CurrentEmojiCode string           `json:"current_emoji_code,omitempty"`
	Evidence         []EventEvidence  `json:"evidence"`
	MinConfidence    float64          `json:"min_confidence"`
	MinMargin        float64          `json:"min_margin"`
}
type PlanResult struct {
	Mode             string   `json:"mode"`
	MutationExecuted bool     `json:"mutation_executed"`
	Intent           Intent   `json:"intent"`
	WorkID           string   `json:"work_id"`
	CommandID        string   `json:"command_id"`
	ExpectedVersion  uint64   `json:"expected_version"`
	LabelIDs         []string `json:"label_ids"`
	EmojiCode        string   `json:"emoji_code,omitempty"`
	CurrentState     State    `json:"current_state"`
	NextState        State    `json:"next_state"`
	StateChange      bool     `json:"state_change"`
	NoOp             bool     `json:"no_op"`
	Reason           string   `json:"reason"`
	EvidenceID       string   `json:"evidence_id,omitempty"`
}

func validOpaqueID(value string) bool {
	return value != "" && len(value) <= 256 && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsRune(value, 0)
}
func validState(value State) bool {
	return value == Todo || value == Active || value == Done || value == Cancelled
}

// ValidEmojiCode validates the bounded canonical lowercase unified-code shape.
// Membership and intent meaning come exclusively from the supplied allowlist.
func ValidEmojiCode(code string) bool {
	if len(code) < 4 || len(code) > 128 || code != strings.ToLower(code) {
		return false
	}
	parts := strings.Split(code, "-")
	if len(parts) > 16 {
		return false
	}
	for _, part := range parts {
		if len(part) < 4 || len(part) > 6 {
			return false
		}
		for _, c := range part {
			if !strings.ContainsRune("0123456789abcdef", c) {
				return false
			}
		}
		value, err := strconv.ParseUint(part, 16, 32)
		if err != nil || value < 0x20 || value > utf8.MaxRune || value >= 0xd800 && value <= 0xdfff {
			return false
		}
		if len(part) > 4 && part[0] == '0' {
			return false
		}
	}
	return true
}

func DefaultEmojis() []EmojiCandidate {
	return []EmojiCandidate{
		{Intent: Question, Code: "2753"}, {Intent: Blocker, Code: "1f6a7"},
		{Intent: Reference, Code: "1f4ce"}, {Intent: Progress, Code: "1f6e0-fe0f"},
		{Intent: CompletionReport, Code: "2705"}, {Intent: CancelRequest, Code: "1f6d1"},
		{Intent: Planned, Code: "1f4cb"},
	}
}

func validPrediction(p Prediction) bool {
	index, ok := IntentIndex(p.Intent)
	if !ok || !finite(p.Confidence) || !finite(p.Margin) || p.Margin < 0 || p.Margin > 1 {
		return false
	}
	if p.Source != Learned && p.Source != Untrained && p.Source != RuleSource {
		return false
	}
	if (p.Source == Learned) != (p.TrainingSteps > 0) {
		return false
	}
	var sum, second float64
	for class, probability := range p.Probabilities {
		if !finite(probability) || probability < 0 || probability > 1 || probability > p.Probabilities[index]+1e-12 {
			return false
		}
		sum += probability
		if class != index && probability > second {
			second = probability
		}
	}
	return math.Abs(sum-1) < 1e-9 && math.Abs(p.Confidence-p.Probabilities[index]) < 1e-9 && math.Abs(p.Margin-(p.Confidence-second)) < 1e-9
}

// Plan produces proposals only. It neither calls an application mutation nor
// sends reactions, notifications, or external messages. A later adapter must
// revalidate catalog freshness, permission, revision, and command replay.
func Plan(input PlanInput, p Prediction) (PlanResult, error) {
	if !validOpaqueID(input.WorkID) || !validOpaqueID(input.CommandID) || input.ExpectedVersion == 0 || input.ExpectedVersion == ^uint64(0) ||
		!validState(input.CurrentState) || !validPrediction(p) ||
		!finite(input.MinConfidence) || input.MinConfidence < .9 || input.MinConfidence > 1 ||
		!finite(input.MinMargin) || input.MinMargin < 0 || input.MinMargin > 1 ||
		len(input.Labels) > 128 || len(input.ExistingLabelIDs) > 128 || len(input.Emojis) > 32 || len(input.Evidence) > 64 {
		return PlanResult{}, ErrPlan
	}
	if input.CurrentEmojiCode != "" && !ValidEmojiCode(input.CurrentEmojiCode) {
		return PlanResult{}, ErrPlan
	}
	for i, label := range input.Labels {
		if _, ok := IntentIndex(label.Intent); !ok || !validOpaqueID(label.ID) {
			return PlanResult{}, ErrPlan
		}
		for _, other := range input.Labels[:i] {
			if other.ID == label.ID {
				return PlanResult{}, ErrPlan
			}
		}
	}
	for _, id := range input.ExistingLabelIDs {
		if !validOpaqueID(id) {
			return PlanResult{}, ErrPlan
		}
	}
	for i, emoji := range input.Emojis {
		if _, ok := IntentIndex(emoji.Intent); !ok || !ValidEmojiCode(emoji.Code) {
			return PlanResult{}, ErrPlan
		}
		for _, other := range input.Emojis[:i] {
			if other.Intent == emoji.Intent {
				return PlanResult{}, ErrPlan
			}
		}
	}
	for _, evidence := range input.Evidence {
		if !validOpaqueID(evidence.ID) || !validOpaqueID(evidence.WorkID) || evidence.Version == 0 ||
			(evidence.Kind != StartedEvent && evidence.Kind != CompletedEvent && evidence.Kind != CancelledEvent) {
			return PlanResult{}, ErrPlan
		}
	}
	result := PlanResult{Mode: "shadow", Intent: p.Intent, WorkID: input.WorkID, CommandID: input.CommandID,
		ExpectedVersion: input.ExpectedVersion, LabelIDs: []string{}, CurrentState: input.CurrentState,
		NextState: input.CurrentState, NoOp: true, Reason: "no_matching_trusted_event"}
	if p.Source == Untrained {
		result.Reason = "untrained_model"
		return result, nil
	}
	if p.Intent == Unclear {
		result.Reason = "unclear"
		return result, nil
	}
	if p.Confidence < input.MinConfidence || p.Margin < input.MinMargin {
		result.Reason = "below_threshold"
		return result, nil
	}
	for _, label := range input.Labels {
		if !label.Active || label.Intent != p.Intent {
			continue
		}
		existing := false
		for _, id := range input.ExistingLabelIDs {
			if id == label.ID {
				existing = true
				break
			}
		}
		if !existing {
			result.LabelIDs = append(result.LabelIDs, label.ID)
		}
	}
	for _, emoji := range input.Emojis {
		if emoji.Intent == p.Intent && emoji.Code != input.CurrentEmojiCode {
			result.EmojiCode = emoji.Code
			break
		}
	}
	var expected EventKind
	var destination State
	switch p.Intent {
	case Progress:
		expected, destination = StartedEvent, Active
	case CompletionReport:
		expected, destination = CompletedEvent, Done
	case CancelRequest:
		expected, destination = CancelledEvent, Cancelled
	}
	// Reopening a terminal work item requires a separate application command;
	// an ordinary content hint cannot silently undo completion/cancellation.
	if (input.CurrentState == Done || input.CurrentState == Cancelled) && destination != "" && destination != input.CurrentState {
		result.Reason = "terminal_state_preserved"
		result.NoOp = len(result.LabelIDs) == 0 && result.EmojiCode == ""
		return result, nil
	}
	var chosen *EventEvidence
	for i := range input.Evidence {
		e := &input.Evidence[i]
		if !e.Trusted || e.WorkID != input.WorkID || e.Version != input.ExpectedVersion {
			continue
		}
		// Conflicting trusted lifecycle facts cannot be resolved by text.
		if chosen != nil && chosen.Kind != e.Kind {
			result.Reason = "conflicting_trusted_events"
			result.NoOp = len(result.LabelIDs) == 0 && result.EmojiCode == ""
			return result, nil
		}
		chosen = e
	}
	if chosen != nil && expected != "" && chosen.Kind == expected {
		result.EvidenceID = chosen.ID
		result.NextState = destination
		result.StateChange = destination != input.CurrentState
		result.Reason = "matching_trusted_event"
		if !result.StateChange {
			result.Reason = "state_already_selected"
		}
	}
	result.NoOp = !result.StateChange && len(result.LabelIDs) == 0 && result.EmojiCode == ""
	return result, nil
}
