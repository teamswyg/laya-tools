// Package claimreferencepilot evaluates the numerical and structural metadata
// contract for a finite AI definition/reference pilot. It reads no text, calls no
// model, authenticates no provider, and grants no execution or adoption authority.
// Receipt digests and declarations must be supplied by a separately admitted
// recorder. Their presence does not establish semantic or provenance truth.
package claimreferencepilot

import "time"

const (
	FamilyCount        = 48
	LocaleCount        = 2
	PanelCount         = 2
	RaterCount         = 2
	HeadCount          = 3
	StateCount         = 3
	StratumCount       = 8
	FamiliesPerStratum = 6
	RoleCount          = 16
	InputCount         = FamilyCount * LocaleCount
	RecordCount        = PanelCount * LocaleCount * RaterCount * FamilyCount
	JudgmentCount      = RecordCount * HeadCount
	SupportFloor       = 2
	AgreedTrueFloor    = 8
	ExactFloor         = 39
	PositiveFloor      = 44
	NaturalnessFloor   = 5
)

// Every semantic/role enum reserves zero for missing or invalid metadata.
type Locale uint8

const (
	KO Locale = iota + 1
	EN
)

type Panel uint8

const (
	First Panel = iota + 1
	FreshRepeat
)

type Rater uint8

const (
	RaterOne Rater = iota + 1
	RaterTwo
)

type Head uint8

const (
	ResponseRequested Head = iota + 1
	CurrentActivityClaimed
	CompletionClaimed
)

type State uint8

const (
	True State = iota + 1
	False
	Unknown
)

type ReviewStatus uint8

const (
	Assessed ReviewStatus = iota + 1
	InvalidOrMissingInput
	RaterUnable
)

type Confidence uint8

const (
	High Confidence = iota + 1
	Medium
	Low
	NotRecordedWithReason
)

type ActorKind uint8

const (
	DeclaredAIAuthorOrReviewer ActorKind = iota + 1
	DeclaredAIReferenceNotHumanGold
)

type Role uint8

const (
	SourceAuthor Role = iota + 1
	KOCommentAuthor
	ENCommentAuthor
	SourceFidelityChecker
	FirstKORaterOne
	FirstKORaterTwo
	FirstENRaterOne
	FirstENRaterTwo
	RepeatKORaterOne
	RepeatKORaterTwo
	RepeatENRaterOne
	RepeatENRaterTwo
	KONaturalnessReviewer
	ENNaturalnessReviewer
	BilingualFrameAuditor
	Comparer
)

// RaterRole returns zero for an invalid panel, locale or rater.
func RaterRole(panel Panel, locale Locale, rater Rater) Role {
	if panel < First || panel > FreshRepeat || locale < KO || locale > EN || rater < RaterOne || rater > RaterTwo {
		return 0
	}
	return FirstKORaterOne + Role((int(panel)-1)*4+(int(locale)-1)*2+int(rater)-1)
}

// Digest is an opaque receipt or input digest. Evaluate never emits it. Zero
// denotes missing evidence. Input digests are metadata supplied by the recorder;
// this package neither reads nor hashes the protected input bodies.
type Digest [32]byte

// Stamp orders contemporaneous UTC events; Ordinal resolves timestamp ties.
// Non-UTC, zero time, and zero ordinal are invalid. Equal stamps are not ordered.
type Stamp struct {
	UTC     time.Time
	Ordinal uint64
}

type ControllerBinding struct {
	Receipt, SourceReceipt, ToolObservation, IsolationReceipt, ExposureReceipt Digest
	Provider, ExposedModelVersion, Broker, RunID                               string
	UnderlyingModelIdentity                                                    string
	UnderlyingModelObservation                                                 Digest
	UnderlyingModelUnknown                                                     bool
	UnderlyingModelIndependenceClaimed                                         bool
	Registered                                                                 Stamp
}

// RoleBinding compares declarations with canonical tool-observed agent/context
// identifiers. ToolObservation is a binding receipt, not authentication performed
// by this package. ContextIDs includes the observed context and every context
// whose content was available; intersections and hierarchical overlaps fail.
// Unknown hidden model identity must be explicitly declared, never fabricated.
type RoleBinding struct {
	Present                                                                                                                     bool
	Position                                                                                                                    Role
	RoleID, AgentID, ContextID, ObservedAgentID, ObservedContextID                                                              string
	ContextIDs                                                                                                                  []string
	ToolObservation, SourceReceipt, ControllerReceipt, CapabilityReceipt, ExposureReceipt, InstructionReceipt, IsolationReceipt Digest
	Kind                                                                                                                        ActorKind
	DeclaredAI, ForkTurnsNone, Available                                                                                        bool
	ForbiddenExposure                                                                                                           bool
	Provider, ExposedModelVersion                                                                                               string
	UnderlyingModelIdentity                                                                                                     string
	UnderlyingModelObservation                                                                                                  Digest
	UnderlyingModelUnknown                                                                                                      bool
	UnderlyingModelIndependenceClaimed                                                                                          bool
	Registered                                                                                                                  Stamp
}

// Bindings carries independently issued prerequisites. No receipt is issued or
// endorsed by Evaluate. A missing receipt blocks dependent evaluation.
type Bindings struct {
	SchemaSource, SchemaAcceptance, CharterSource, CharterAcceptance            Digest
	MethodScope, BrokerAdmission, ExecutionAdmission, AcquisitionCAP, ReviewCAP Digest
	Controller                                                                  ControllerBinding
	Locked                                                                      Stamp
}

type Source struct {
	Receipt, CreationReceipt            Digest
	Bytes                               uint32
	Created, Saved                      Stamp
	EditorialRevisions, SelectedVersion uint8
	VersionsRetained                    bool
}

type InputScope struct {
	Present                                                              bool
	Receipt, CommentDigest, ModelEvidenceDigest, ReferenceEvidenceDigest Digest
	Bytes, Codepoints                                                    uint32
	ValidNonemptyUTF8, NoTruncation                                      bool
	ContextUnits                                                         uint8
	EditorialRevisions, SelectedVersion                                  uint8
	VersionsRetained                                                     bool
	Locked                                                               Stamp
}

// Screen joins a locked AI review to the actual source and selected bilingual
// inputs. Naturalness uses only its own locale digest; other screens use both.
// A soft naturalness rejection may count zero without a hard hold. Missing
// screens or unresolved source/input/meaning/frame holds always block.
type Screen struct {
	Present                                                       bool
	Reviewer                                                      Role
	SourceReceipt                                                 Digest
	InputDigests                                                  [LocaleCount]Digest
	Receipt                                                       Digest
	Accepted, HardHold, MaterialLossOrInvention, ReferencesHidden bool
	Locked                                                        Stamp
}

type Family struct {
	Present         bool
	Key             string
	Stratum         uint8 // 1..8, assigned before labels; six distinct families per stratum.
	Source          Source
	Inputs          [LocaleCount]InputScope
	Fidelity, Frame Screen
	Naturalness     [LocaleCount]Screen
}

type DefinitionDecision struct {
	Initial, Repeat                                                      Digest
	Revisions                                                            uint8 // zero repeats v0; one requires a distinct, locked v1.
	DefectEvidence, BilingualRepairEvidence, AlternativeReadingsEvidence Digest
	Reviewer                                                             Role
	FirstDissentPreserved, NoStructuralChanges                           bool
	HiddenFrameValuesUsed                                                bool
	Locked                                                               Stamp
}

type Lifecycle struct {
	InitialDefinitionLocked, InputsLocked, FramesLocked Stamp
	PanelLocks                                          [PanelCount]Stamp
	FrameValuesReleased, FinalReviewLocked              Stamp
	Decision                                            DefinitionDecision
}

type RecordKey struct {
	Panel  Panel
	Locale Locale
	Rater  Rater
	Family uint8
}
type Span struct{ StartByte, EndByte int }

// Judgment retains semantic state separately from confidence and inability.
// Reason/Rule receipts and evidence spans are checked structurally only.
type Judgment struct {
	Head                       Head
	FirstState                 State
	EvidenceSpans              []Span
	ReasonReceipt, RuleReceipt Digest
	Confidence                 Confidence
	PlausibleStates            []State
	FirstImmutable             bool
	ConfidenceIsSemanticState  bool
}

type ReferenceRecord struct {
	Key                                                                                                     RecordKey
	FamilyKey                                                                                               string
	RaterRole                                                                                               Role
	Kind                                                                                                    ActorKind
	Receipt, Definition, InputScopeReceipt, InputDigest                                                     Digest
	Status                                                                                                  ReviewStatus
	BlockedReasonReceipt                                                                                    Digest
	Judgments                                                                                               []Judgment // exactly three ordered heads if assessed; empty if blocked.
	Started, Locked                                                                                         Stamp
	PartnerLocaleHidden, ModelResultsHidden, DesiredStatesHidden, SourceAndStratumHidden, FrameValuesHidden bool
	HumanGoldOrOperationalAccuracyClaimed                                                                   bool
}

// Dissent resolves against one immutable reference, retaining its head, panel,
// locale, version, first state and alternatives. It supplies no replacement vote.
type Dissent struct {
	Reference                                   RecordKey
	ReferenceReceipt, Definition, ReasonReceipt Digest
	Head                                        Head
	FirstState                                  State
	PlausibleStates                             []State
}

// Distribution is optional recorded-opinion metadata. Each explicit panel,
// locale, family and head cell is unique; membership must equal the assessed
// locked rater sample. Counts always use true, false, unknown order.
type Distribution struct {
	Panel                  Panel
	Locale                 Locale
	Family                 uint8
	Head                   Head
	Definition             Digest
	Members                []RecordKey
	Counts                 [StateCount]int
	DependenceLimitReceipt Digest
}

type Audit struct {
	FirstReferencesPreserved                                                                           bool
	MajorityAutomaticRelabel, FirstReferenceOverwritten, UnresolvedCasesDropped, ConfidenceSubstituted bool
	Dissent                                                                                            []Dissent
	Distributions                                                                                      []Distribution
}

// Batch must contain every attempted record, including failures and duplicates.
// Evaluate never mutates it, selects an easier subset, or exports case metadata.
type Batch struct {
	Bindings             Bindings
	Usage                Usage
	Roles                [RoleCount]RoleBinding
	Families             [FamilyCount]Family
	SelectedInputsDigest Digest // one aggregate digest of the fixed 96-input inventory.
	PanelInputsDigest    [PanelCount]Digest
	Lifecycle            Lifecycle
	Records              []ReferenceRecord
	Audit                Audit
	ProhibitedOperations uint64 // classifier/model/fit/predict or other out-of-scope operations.
}

type Verdict string

const (
	Pass    Verdict = "PASS"
	Fail    Verdict = "FAIL"
	Blocked Verdict = "BLOCKED"
)

type Failures struct {
	Bindings, Roles, Inventory, Scopes, Screens, Chronology, Revisions, Records, Audit, HardHolds int
	SupportCells, AgreedTrueCells, ExactCells, PositiveCells, NaturalnessCells                    int
}

type Completeness struct {
	Families, Inputs, FidelityReviews, FrameReviews, NaturalnessReviews                                                        int
	RecordsRetained, UniqueRecordSlots, MissingRecordSlots, DuplicateAttempts, InvalidRecords, BlockedRecords, AssessedRecords int
	JudgmentsRetained, UsableJudgments                                                                                         int
	Passed                                                                                                                     bool
}

// Aggregate contains counts and fixed-size numerical tables only. Every
// agreement denominator remains 48, every naturalness denominator remains 6.
// PASS means these metadata/number checks passed; it does not prove semantic
// fidelity, authentic provenance, independent models, human Gold, or authority.
type Aggregate struct {
	Verdict                           Verdict
	Failures                          Failures
	Completeness                      Completeness
	StratumFamilies                   [StratumCount]int
	StateSupport                      [PanelCount][LocaleCount][RaterCount][HeadCount][StateCount]int
	FirstAgreedTrue                   [LocaleCount][HeadCount]int
	ExactAgreement, PositiveAgreement [PanelCount][LocaleCount][HeadCount]int
	Confusion                         [PanelCount][LocaleCount][HeadCount][StateCount][StateCount]int
	Naturalness                       [StratumCount]int
	QualifiedModel                    bool `json:"qualified_model"`
	HumanGold                         bool `json:"human_gold"`
	IndependentModelsEstablished      bool `json:"independent_models_established"`
}
