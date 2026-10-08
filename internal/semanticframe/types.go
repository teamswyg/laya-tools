// Package semanticframe joins Source-grounded communication plans to pinned
// inventories. Successful checks prove structural links, never meaning or QA.
package semanticframe

import "github.com/teamswyg/laya-tools/internal/sourcecohort"

const (
	Schema                 = "riido-semantic-family-frame-draft-v3"
	VersionSchema          = "riido-accepted-inventory-version-binding-draft-v1"
	MaxFrameBytes          = 128 << 10
	MaxInventoryBytes      = 128 << 10
	MaxSourceBytes         = 16384
	MaxItems               = 256
	InventorySchema        = "riido-inventory-graph-declaration-v1"
	OriginalProducerSchema = "riido-inventory-original-producer-declaration-v1"
	AmendedProducerSchema  = "riido-inventory-amended-producer-declaration-v1"
	AcceptanceSchema       = "riido-inventory-acceptance-declaration-v1"
)

type File = sourcecohort.File
type Span = sourcecohort.Span

type Frame struct {
	Schema                string      `json:"schema"`
	FamilyID              string      `json:"family_id"`
	SourceID              string      `json:"source_id"`
	Slot                  int         `json:"family_slot"`
	Source                File        `json:"source"`
	Inventory             File        `json:"accepted_inventory"`
	Producer              File        `json:"accepted_inventory_binding"`
	VersionEvidence       File        `json:"accepted_version_evidence"`
	Observation           File        `json:"accepted_observation"`
	SourceReview          File        `json:"accepted_source_review"`
	Definitions           File        `json:"definitions"`
	InventorySchemaSource File        `json:"inventory_schema_source"`
	Boundary              Span        `json:"source_boundary"`
	Plan                  Plan        `json:"communication_plan"`
	Ambiguities           []Ambiguity `json:"preserved_ambiguities"`
	DesiredLabels         bool        `json:"desired_labels_in_packet"`
	TargetMode            string      `json:"target_input_mode"`
}

type Binding struct {
	Pointer string `json:"inventory_pointer"`
	Role    string `json:"role"`
}
type Plan struct {
	Bindings       []Binding `json:"bindings"`
	Evidence       []Span    `json:"source_evidence"`
	Description    string    `json:"description_private"`
	WordingFreedom string    `json:"wording_freedom"`
}
type Ambiguity struct {
	Pointers    []string `json:"inventory_pointers"`
	Evidence    []Span   `json:"source_evidence"`
	Description string   `json:"description_private"`
	Operation   string   `json:"operation"`
}

// VersionBinding declares selected-version links. Its presence alone does not
// accept an inventory; producer and independent acceptance evidence must join.
type VersionBinding struct {
	Schema              string `json:"schema"`
	Selection           string `json:"selection"`
	Source              File   `json:"source"`
	Observation         File   `json:"observation"`
	Inventory           File   `json:"inventory"`
	Producer            File   `json:"producer_binding"`
	SourceReview        File   `json:"standard_source_review"`
	InventoryAcceptance []File `json:"independent_inventory_acceptance_evidence"`
	Predecessors        []File `json:"retained_predecessor_bindings"`
}

type Component struct {
	ID          string `json:"component_id"`
	Description string `json:"scope_description_private"`
	Evidence    []Span `json:"evidence"`
}
type Proposition struct {
	ID          string   `json:"proposition_id"`
	Description string   `json:"proposition_description_private"`
	Polarity    string   `json:"polarity"`
	Scope       string   `json:"scope_description_private"`
	Subjects    []string `json:"subject_referent_ids"`
	Evidence    []Span   `json:"evidence"`
}
type Opposition struct {
	ID          string `json:"opposition_id"`
	Left        string `json:"left_proposition_id"`
	Right       string `json:"right_proposition_id"`
	Description string `json:"relation_description_private"`
	Evidence    []Span `json:"evidence"`
}
type Temporal struct {
	ID                    string   `json:"temporal_id"`
	Relation              string   `json:"relation_to_anchor"`
	Propositions          []string `json:"proposition_ids"`
	Anchor                string   `json:"anchor_description_private"`
	Evidence              []Span   `json:"anchor_evidence"`
	CommunicationEvidence []Span   `json:"communication_time_relation_evidence"`
}
type Referent struct {
	ID          string  `json:"referent_id"`
	Kind        string  `json:"referent_kind"`
	Parent      *string `json:"parent_referent_id"`
	Description string  `json:"description_private"`
	Evidence    []Span  `json:"evidence"`
}
type Constraint struct {
	Kind         string   `json:"kind"`
	Propositions []string `json:"proposition_ids"`
	Description  string   `json:"constraint_description_private"`
	Evidence     []Span   `json:"evidence"`
}

// InventoryGraph is a generic typed projection, not a decoder for historical
// private inventory variants. The caller must pin and adapt their full inventory
// with an explicit reviewed adapter before using these graph joins.
type InventoryGraph struct {
	Schema       string        `json:"schema"`
	SourceID     string        `json:"source_id"`
	Source       File          `json:"source"`
	Boundary     Span          `json:"complete_source_boundary"`
	Components   []Component   `json:"inspected_components"`
	Propositions []Proposition `json:"propositions"`
	Oppositions  []Opposition  `json:"opposition_relations"`
	Temporal     []Temporal    `json:"temporal_support"`
	Referents    []Referent    `json:"referent_support"`
	Constraints  []Constraint  `json:"reply_quote_adoption_retraction_negation_constraints"`
	Dependencies []string      `json:"required_dependencies"`
}

// ProducerDeclaration is a new generic wire contract. Original and amended
// branches have distinct schemas; neither decodes existing private producers.
// The fields are producer declarations even after their bytes are verified.
type ProducerDeclaration struct {
	Schema       string `json:"schema"`
	ProducerID   string `json:"producer_id"`
	SourceID     string `json:"source_id"`
	Source       File   `json:"source"`
	Observation  File   `json:"observation"`
	Inventory    File   `json:"inventory"`
	Predecessors []File `json:"retained_predecessor_bindings"`
}

// InventoryAcceptance declares exact inventory-version acceptance separately
// from the standard Source review. Reviewer IDs and verdict are not authenticated.
type InventoryAcceptance struct {
	Schema       string `json:"schema"`
	SourceID     string `json:"source_id"`
	Selection    string `json:"selection"`
	Source       File   `json:"source"`
	Observation  File   `json:"observation"`
	Inventory    File   `json:"inventory"`
	Producer     File   `json:"producer_binding"`
	SourceReview File   `json:"standard_source_review"`
	ReviewerID   string `json:"reviewer_id"`
	Verdict      string `json:"declared_inventory_verdict"`
	Predecessors []File `json:"retained_predecessor_bindings"`
}

// PinnedBytes is caller-captured raw content and its independently expected pin.
// Normalization verifies hash/length; it performs no filesystem opens or receipts.
type PinnedBytes struct {
	File  File
	Bytes []byte
}
type AcceptedInputs struct {
	Version             PinnedBytes
	Source              PinnedBytes
	Observation         PinnedBytes
	Inventory           PinnedBytes
	Producer            PinnedBytes
	SourceReview        PinnedBytes
	InventoryAcceptance []PinnedBytes
	Predecessors        []PinnedBytes
}

// NormalizedAccepted can only be built by NormalizeAccepted. Accessors return
// value metadata; the retained source and inventory cannot be mutated by callers.
type NormalizedAccepted struct {
	version    VersionBinding
	versionPin File
	source     []byte
	inventory  InventoryGraph
}

func (n NormalizedAccepted) SourcePin() File    { return n.version.Source }
func (n NormalizedAccepted) InventoryPin() File { return n.version.Inventory }

// Summary contains counts and the strict limit of this software check. No
// caller text, private IDs, paths or decoder diagnostics are returned in errors.
type Summary struct {
	State            string `json:"state"`
	Bindings         int    `json:"bindings"`
	Ambiguities      int    `json:"ambiguities"`
	MeaningProven    bool   `json:"meaning_proven"`
	TrainingEligible bool   `json:"training_eligible"`
	Limits           string `json:"limits"`
}

type Error string

func (e Error) Error() string { return string(e) }
