// Package statehintpilot evaluates three shadow display intents. It neither
// trains models nor exposes any application write, notification or event API.
package statehintpilot

import (
	"errors"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

const (
	BatchSize         = statehintcatalog.MaxWorks
	MaxCases          = 2400
	MaxTotalTextBytes = 8 << 20
)

var (
	ErrInput      = errors.New("statehintpilot: invalid or oversized cases")
	ErrBudget     = errors.New("statehintpilot: resource budget exceeded")
	ErrEvaluation = errors.New("statehintpilot: shadow evaluation failed")
)

// Expected labels and IDs are evaluation metadata, never classifier input.
// Case IDs also serve as opaque reader references. A private caller reader may
// resolve them internally; this package never receives or publishes native IDs.
type Case struct {
	ID       string           `json:"id"`
	Family   string           `json:"family"`
	Locale   string           `json:"locale"`
	Text     string           `json:"text"`
	Expected statehint.Intent `json:"expected_intent"`
}

type Options struct {
	Reader            statehintcatalog.SnapshotReader
	Scope             statehintcatalog.Scope
	MinConfidence     float64
	MinMargin         float64
	MaxCases          int
	MaxTotalTextBytes int
	MaxReadCalls      int
}

// Eligibility is a model/rule gate; proposal creation also needs current
// annotation configuration. Missing counts cover all AnnotationChecked rows,
// independently of eligibility; a correct emoji cannot hide a missing label.
type TargetMetric struct {
	Intent                 statehint.Intent `json:"intent"`
	Expected               int              `json:"expected"`
	Predicted              int              `json:"predicted"`
	Correct                int              `json:"correct"`
	Eligible               int              `json:"eligible"`
	CorrectEligible        int              `json:"correct_eligible"`
	WrongEligible          int              `json:"wrong_eligible"`
	ProposalsCreated       int              `json:"proposals_created"`
	CorrectProposals       int              `json:"correct_proposals"`
	WrongProposals         int              `json:"wrong_proposals"`
	AnnotationChecked      int              `json:"annotation_checked"`
	ExistingAnnotationNoOp int              `json:"existing_annotation_no_op"`
	CurrentLabelMissing    int              `json:"current_label_missing"`
	CurrentEmojiMissing    int              `json:"current_emoji_missing"`
}

type Metrics struct {
	Rows                    int                                               `json:"rows"`
	Classified              int                                               `json:"classified"`
	Correct                 int                                               `json:"correct"`
	Eligible                int                                               `json:"eligible"`
	Abstained               int                                               `json:"abstained"`
	ProposalsCreated        int                                               `json:"proposals_created"`
	WrongEligible           int                                               `json:"wrong_eligible"`
	WrongProposals          int                                               `json:"wrong_proposals"`
	AnnotationChecked       int                                               `json:"annotation_checked"`
	ExistingAnnotationNoOp  int                                               `json:"existing_annotation_no_op"`
	UnavailableWork         int                                               `json:"unavailable_work"`
	CurrentLabelMissing     int                                               `json:"current_label_missing"`
	CurrentEmojiMissing     int                                               `json:"current_emoji_missing"`
	EligibleWithoutProposal int                                               `json:"eligible_without_proposal"`
	ExpectedCounts          [statehint.IntentCount]int                        `json:"expected_counts"`
	PredictedCounts         [statehint.IntentCount]int                        `json:"predicted_counts"`
	Confusion               [statehint.IntentCount][statehint.IntentCount]int `json:"confusion"`
	Targets                 [3]TargetMetric                                   `json:"targets"`
}

type Observation struct {
	CaseID                 string                         `json:"case_id"`
	Family                 string                         `json:"family"`
	Locale                 string                         `json:"locale"`
	Expected               statehint.Intent               `json:"expected"`
	Observed               statehint.Intent               `json:"observed,omitempty"`
	Confidence             float64                        `json:"confidence"`
	Probability            [statehint.IntentCount]float64 `json:"probabilities"`
	Margin                 float64                        `json:"margin"`
	Source                 statehint.Source               `json:"source,omitempty"`
	TrainingSteps          uint64                         `json:"training_steps"`
	Eligible               bool                           `json:"eligible"`
	ProposalCreated        bool                           `json:"proposal_created"`
	WrongEligible          bool                           `json:"wrong_eligible"`
	WrongProposal          bool                           `json:"wrong_proposal"`
	ExistingAnnotationNoOp bool                           `json:"existing_annotation_no_op"`
	CurrentLabelMissing    bool                           `json:"current_label_missing"`
	CurrentEmojiMissing    bool                           `json:"current_emoji_missing"`
	// Missing flags are known only when AnnotationChecked is true. Catalog
	// diagnostics are recorded independently of the confidence gate.
	AnnotationChecked               bool                `json:"annotation_checked"`
	AnnotationReason                string              `json:"annotation_reason"`
	MappingDiagnosticsChecked       bool                `json:"mapping_diagnostics_checked"`
	MappingIssues                   []MappingDiagnostic `json:"mapping_issues"`
	LabelProposalCount              int                 `json:"label_proposal_count"`
	EmojiProposed                   bool                `json:"emoji_proposed"`
	Guard                           string              `json:"guard,omitempty"`
	Reason                          string              `json:"reason"`
	StateReason                     string              `json:"state_reason,omitempty"`
	RawReaderStatePlanningAvailable bool                `json:"raw_reader_state_planning_available"`
	RawReaderStatePlanningChecked   bool                `json:"raw_reader_state_planning_checked"`
	PilotStatePlanningDisabled      bool                `json:"pilot_state_planning_disabled"`
	// Candidate resolution/selection is disabled, although snapshot metadata
	// still undergoes structural validation before filtering.
	StateCandidatesChecked bool `json:"state_candidates_checked"`
	MutationExecuted       bool `json:"mutation_executed"`
	StateChange            bool `json:"state_change"`
}

// Catalog issue codes exclude native/opaque catalog references and source text.
type MappingDiagnostic struct {
	Reason string           `json:"reason"`
	Intent statehint.Intent `json:"intent,omitempty"`
}

type Budgets struct {
	CaseLimit           int `json:"case_limit"`
	TextByteLimit       int `json:"text_byte_limit"`
	ReadCallLimit       int `json:"read_call_limit"`
	CasesUsed           int `json:"cases_used"`
	TextBytesUsed       int `json:"text_bytes_used"`
	ReadCallsUsed       int `json:"read_calls_used"`
	PredictionCallsUsed int `json:"prediction_calls_used"`
	BatchSize           int `json:"batch_size"`
}

// Reports contain no source-text field or native catalog/scope references.
// Caller case/family identifiers can still be sensitive. Publication remains an
// application decision; this package has no automatic publication facility.
type Report struct {
	Schema                    string                                  `json:"schema"`
	Mode                      string                                  `json:"mode"`
	ReaderMode                string                                  `json:"reader_mode"`
	FixtureCatalog            string                                  `json:"fixture_catalog,omitempty"`
	SyntheticCanonicalVersion string                                  `json:"synthetic_canonical_version,omitempty"`
	CatalogConsistency        string                                  `json:"catalog_consistency"`
	MappingDiagnosticScope    string                                  `json:"mapping_diagnostic_scope"`
	PlacementVerified         bool                                    `json:"placement_verified"`
	MutationExecuted          bool                                    `json:"mutation_executed"`
	StateChanges              int                                     `json:"state_changes"`
	MinConfidence             float64                                 `json:"min_confidence"`
	MinMargin                 float64                                 `json:"min_margin"`
	AllowedIntents            [3]statehint.Intent                     `json:"allowed_intents"`
	IntentOrder               [statehint.IntentCount]statehint.Intent `json:"intent_order"`
	FamilyCount               int                                     `json:"family_count"`
	FamilyCounts              map[string]int                          `json:"family_counts"`
	RepeatedFamilies          int                                     `json:"repeated_families"`
	IndependentProductSamples bool                                    `json:"independent_product_samples"`
	Budgets                   Budgets                                 `json:"budgets"`
	Overall                   Metrics                                 `json:"overall"`
	PerLanguage               map[string]Metrics                      `json:"per_language"`
	Observations              []Observation                           `json:"observations"`
}

func AllowedIntents() [3]statehint.Intent {
	return [3]statehint.Intent{statehint.Progress, statehint.CompletionReport, statehint.Question}
}

func allowed(intent statehint.Intent) bool {
	return intent == statehint.Progress || intent == statehint.CompletionReport || intent == statehint.Question
}

func targetIndex(intent statehint.Intent) (int, bool) {
	switch intent {
	case statehint.Progress:
		return 0, true
	case statehint.CompletionReport:
		return 1, true
	case statehint.Question:
		return 2, true
	default:
		return 0, false
	}
}
