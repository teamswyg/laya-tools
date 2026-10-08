// Package acceptedfullbridge joins a pinned FULL inventory's independent
// acceptance to its mechanically derived graph and communication plan. It never
// generates historical judgments or opens caller references.
package acceptedfullbridge

import (
	"github.com/teamswyg/laya-tools/internal/inventoryprojection"
	"github.com/teamswyg/laya-tools/internal/semanticframe"
)

const (
	VersionSchema    = "riido-full-selected-version-declaration-v1"
	ProducerSchema   = "riido-full-producer-declaration-v1"
	AcceptanceSchema = "riido-full-independent-acceptance-declaration-v1"
	HistorySchema    = "riido-full-retained-history-declaration-v1"
	FrameSchema      = "riido-full-projected-family-frame-v1"
	AcceptanceScope  = "complete_full_inventory_official_and_support"
	MaxBytes         = 128 << 10
	MaxItems         = 256
	MaxTotalBytes    = 8 << 20
)

type File = semanticframe.File
type PinnedBytes = semanticframe.PinnedBytes

// These are new explicit generic wire contracts, not historical native decoders.
// Actual native adapters must preserve/pin their original records separately.
type Version struct {
	Schema        string `json:"schema"`
	SourceID      string `json:"source_id"`
	Selection     string `json:"selection"`
	ReviewLineage string `json:"review_lineage"`
	Source        File   `json:"source"`
	Observation   File   `json:"observation"`
	Full          File   `json:"selected_full_inventory"`
	Graph         File   `json:"graph_projection"`
	Projection    File   `json:"projection_binding"`
	Producer      File   `json:"producer_binding"`
	Review        File   `json:"standard_source_review"`
	Acceptance    []File `json:"independent_full_acceptance"`
	History       []File `json:"retained_history"`
}
type Producer struct {
	Schema        string `json:"schema"`
	SourceID      string `json:"source_id"`
	ProducerID    string `json:"producer_id"`
	ProducerActor string `json:"producer_actor"`
	Selection     string `json:"selection"`
	Source        File   `json:"source"`
	Observation   File   `json:"observation"`
	Full          File   `json:"full_inventory"`
	History       []File `json:"retained_history"`
}
type Acceptance struct {
	Schema         string   `json:"schema"`
	SourceID       string   `json:"source_id"`
	Selection      string   `json:"selection"`
	Source         File     `json:"source"`
	Observation    File     `json:"observation"`
	Full           File     `json:"full_inventory"`
	Producer       File     `json:"producer_binding"`
	Review         File     `json:"standard_source_review"`
	ReviewerID     string   `json:"reviewer_id"`
	ReviewedUTC    string   `json:"reviewed_utc"`
	Verdict        string   `json:"declared_full_verdict"`
	Scope          string   `json:"review_scope"`
	OfficialFields []string `json:"complete_official_field_keys"`
	History        []File   `json:"retained_history"`
}

// Retained history carries explicit old-version pins and disposition. Its
// referenced records remain unopened; a retained hold is never made a pass.
type History struct {
	Schema      string `json:"schema"`
	SourceID    string `json:"source_id"`
	Source      File   `json:"source"`
	Observation File   `json:"observation"`
	Full        File   `json:"full_inventory"`
	Producer    File   `json:"producer_binding"`
	Review      File   `json:"standard_source_review"`
	Version     File   `json:"historical_version_record"`
	Disposition string `json:"retained_disposition"`
	Relation    string `json:"history_relation"`
}
type Frame struct {
	Schema     string              `json:"schema"`
	Full       File                `json:"accepted_full_inventory"`
	Projection File                `json:"projection_binding"`
	Plan       semanticframe.Frame `json:"communication_frame"`
}
type Inputs struct {
	Projection          inventoryprojection.Inputs
	ProjectionBinding   PinnedBytes
	Graph               PinnedBytes
	Version             PinnedBytes
	Observation         PinnedBytes
	Producer            PinnedBytes
	Review              PinnedBytes
	CheckBinding        PinnedBytes
	CheckReport         PinnedBytes
	ReadStart           PinnedBytes
	ReadResult          PinnedBytes
	Acceptance          []PinnedBytes
	History             []PinnedBytes
	Frame               PinnedBytes
	RegisteredSourceIDs []string
}
type Summary struct {
	State                  string                `json:"state"`
	AcceptanceDeclarations int                   `json:"acceptance_declarations"`
	RetainedHolds          int                   `json:"retained_holds"`
	Plan                   semanticframe.Summary `json:"plan_structure"`
	MeaningProven          bool                  `json:"meaning_proven"`
	TrainingEligible       bool                  `json:"training_eligible"`
	Limits                 string                `json:"limits"`
}
type Result struct {
	Projection inventoryprojection.Result
	Summary    Summary
}
type Error string

func (e Error) Error() string { return string(e) }
