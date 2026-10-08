// Package inventoryprojection mechanically derives a graph while retaining the
// exact full inventory. It generates no historical acceptance or semantic QA.
package inventoryprojection

import (
	"github.com/teamswyg/laya-tools/internal/semanticframe"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

const (
	Version        = "riido-inventoryprojection-v1"
	MaxBytes       = 128 << 10
	MaxSourceBytes = 16384
	MaxTotalBytes  = 8 << 20
	MaxItems       = 256
)

type File = sourcecohort.File
type Span = sourcecohort.Span
type PinnedBytes = semanticframe.PinnedBytes
type Config struct {
	Full      File   `json:"selected_full_inventory"`
	Adapter   File   `json:"adapter_source"`
	Version   string `json:"adapter_version"`
	Mapping   File   `json:"mapping_contract"`
	Native    File   `json:"native_contract"`
	GraphPath string `json:"graph_output_path"`
}
type NativeContract struct {
	Schema string   `json:"native_inventory_schema"`
	Fields []string `json:"official_field_keys"`
}
type MappingContract struct {
	GraphSchema    string   `json:"graph_schema"`
	CopiedFields   []string `json:"copied_fields"`
	FullOnlyFields []string `json:"full_only_fields"`
}
type Inputs struct {
	Config  PinnedBytes
	Full    PinnedBytes
	Native  PinnedBytes
	Mapping PinnedBytes
	Adapter PinnedBytes
	Source  *PinnedBytes // optional: nil means actual UTF-8 span checks pending
}
type Role struct {
	ID    string `json:"id"`
	Agent string `json:"canonical_agent"`
}
type AdditionalRead struct {
	Source File   `json:"source"`
	Start  File   `json:"read_start"`
	Result File   `json:"read_result"`
	Actor  string `json:"actor"`
}
type Support struct {
	Status       string   `json:"status"`
	Value        *string  `json:"value"`
	Reason       *string  `json:"reason"`
	Inspected    bool     `json:"relevant_scope_inspected"`
	Description  string   `json:"support_description_private"`
	Propositions []string `json:"proposition_ids"`
	Referents    []string `json:"referent_ids"`
	Oppositions  []string `json:"opposition_ids"`
	Temporal     []string `json:"temporal_ids"`
	Evidence     []Span   `json:"evidence"`
}
type OfficialField struct {
	Name    string
	Raw     []byte
	Support Support
}

// Full preserves every field in the generic native declaration. OfficialFields
// retains original object order; callers can build separate sorted lookup slices.
// This type's JSON representation is not a replacement native inventory.
type Full struct {
	Schema              string                      `json:"schema"`
	ID                  string                      `json:"source_id"`
	Source              File                        `json:"source"`
	Creation            File                        `json:"creation_receipt"`
	ReadBinding         File                        `json:"read_binding"`
	AdditionalReads     []AdditionalRead            `json:"recorded_additional_source_reads"`
	Coder               Role                        `json:"coder"`
	Phase               string                      `json:"phase"`
	Scope               string                      `json:"observation_scope"`
	Boundary            Span                        `json:"complete_source_boundary"`
	Complete            bool                        `json:"inventory_complete_declared"`
	Components          []semanticframe.Component   `json:"inspected_components"`
	Propositions        []semanticframe.Proposition `json:"propositions"`
	NegativeIDs         []string                    `json:"distinct_negative_proposition_ids"`
	NegationDescription string                      `json:"combined_negation_scope_description_private"`
	Oppositions         []semanticframe.Opposition  `json:"opposition_relations"`
	Temporal            []semanticframe.Temporal    `json:"temporal_support"`
	Referents           []semanticframe.Referent    `json:"referent_support"`
	Constraints         []semanticframe.Constraint  `json:"reply_quote_adoption_retraction_negation_constraints"`
	OfficialFields      []OfficialField             `json:"official_field_support"`
	Dependencies        []string                    `json:"required_dependencies"`
	Unresolved          []string                    `json:"unresolved_mapping_or_meaning_questions_private"`
	DesiredLabels       bool                        `json:"desired_labels_in_packet"`
}
type Binding struct {
	Full    File   `json:"selected_full_inventory"`
	Graph   File   `json:"graph_projection"`
	Adapter File   `json:"adapter_source"`
	Version string `json:"adapter_version"`
	Mapping File   `json:"mapping_contract"`
	Native  File   `json:"native_contract"`
	Config  File   `json:"projection_config"`
}
type Result struct {
	RetainedFullBytes []byte
	Full              Full
	Graph             semanticframe.InventoryGraph
	GraphBytes        []byte
	Binding           Binding
	UTF8SpansVerified bool
	MeaningProven     bool
	TrainingEligible  bool
}
type Error string

func (e Error) Error() string { return string(e) }

// SupportedMapping is a content-free public mapping, not a native schema or
// official-field keyset. A caller must pin its exact encoded contract bytes.
func SupportedMapping() MappingContract {
	return MappingContract{
		GraphSchema:    semanticframe.InventorySchema,
		CopiedFields:   []string{"source_id", "source", "complete_source_boundary", "inspected_components", "propositions", "opposition_relations", "temporal_support", "referent_support", "reply_quote_adoption_retraction_negation_constraints", "required_dependencies"},
		FullOnlyFields: []string{"creation_receipt", "read_binding", "recorded_additional_source_reads", "coder", "phase", "observation_scope", "inventory_complete_declared", "distinct_negative_proposition_ids", "combined_negation_scope_description_private", "official_field_support", "unresolved_mapping_or_meaning_questions_private", "desired_labels_in_packet"},
	}
}
