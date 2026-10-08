// Package nativecontract joins two exact prospective native metadata contracts.
// It performs no file/network IO and does not authenticate semantic declarations.
package nativecontract

import (
	"github.com/teamswyg/laya-tools/internal/acceptedfullbridge"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

const (
	Original                  = "original_reconsideration"
	Amended                   = "amended_pair"
	OriginalAcceptanceSchema  = "new400-native-original-full-inventory-acceptance-v1"
	OriginalVersionSchema     = "new400-native-original-selected-review-version-v1"
	OriginalDispositionSchema = "new400-current-original-selected-review-disposition-v1"
	AmendedAcceptanceSchema   = "new400-native-amended-full-inventory-acceptance-v1"
	AmendedVersionSchema      = "new400-native-amended-selected-version-v1"
	AmendedDispositionSchema  = "new400-current-amended-selected-disposition-v1"
	Scope                     = "complete_official_and_supporting_inventory"
	MaxBytes                  = 128 << 10
	MaxTotalBytes             = 8 << 20
	MaxItems                  = 256
)

type File = sourcecohort.File
type PinnedBytes = acceptedfullbridge.PinnedBytes
type CopyRead struct {
	Input  File `json:"input"`
	Start  File `json:"read_start"`
	Result File `json:"read_result"`
}
type OriginalAcceptance struct {
	Schema       string     `json:"schema"`
	ID           string     `json:"source_id"`
	Reviewer     string     `json:"reviewer_id"`
	ReviewedUTC  string     `json:"reviewed_utc"`
	Source       File       `json:"source"`
	Observation  File       `json:"observation"`
	Full         File       `json:"full_inventory"`
	Producer     File       `json:"coder_record_binding"`
	Review       File       `json:"standard_review"`
	OwnInventory File       `json:"own_precoder_inventory"`
	ReadStart    File       `json:"original_read_start"`
	ReadResult   File       `json:"original_read_result"`
	Start        File       `json:"reconsideration_start"`
	Copies       []CopyRead `json:"copy_read_receipts"`
	Check        File       `json:"check_binding"`
	Report       File       `json:"check_report"`
	Scope        string     `json:"scope"`
	Verdict      string     `json:"verdict"`
	Unresolved   int        `json:"unresolved_count"`
	HeldReview   File       `json:"retained_held_review"`
	Method       File       `json:"method_limitation"`
	Rationale    File       `json:"semantic_rationale"`
	WholeQA      bool       `json:"whole_source_qa_pass"`
	Training     bool       `json:"training_eligible"`
}
type OriginalVersion struct {
	Schema       string `json:"schema"`
	ID           string `json:"source_id"`
	Reviewer     string `json:"reviewer_id"`
	UTC          string `json:"recorded_utc"`
	Selection    string `json:"selection"`
	Source       File   `json:"source"`
	Observation  File   `json:"observation"`
	Full         File   `json:"full_inventory"`
	Producer     File   `json:"coder_record_binding"`
	Review       File   `json:"standard_review"`
	Acceptance   File   `json:"full_inventory_acceptance"`
	HeldReview   File   `json:"retained_held_review"`
	Start        File   `json:"reconsideration_start"`
	ReadStart    File   `json:"original_read_start"`
	ReadResult   File   `json:"original_read_result"`
	Check        File   `json:"check_binding"`
	Report       File   `json:"check_report"`
	OwnInventory File   `json:"own_precoder_inventory"`
	Method       File   `json:"method_limitation"`
	Unchanged    bool   `json:"original_source_and_coder_records_unchanged"`
	OldPass      bool   `json:"old_held_record_pass"`
	WholeQA      bool   `json:"whole_source_qa_pass"`
	Training     bool   `json:"training_eligible"`
}
type AmendedAcceptance struct {
	Schema              string     `json:"schema"`
	ID                  string     `json:"source_id"`
	Reviewer            string     `json:"reviewer_id"`
	ReviewedUTC         string     `json:"reviewed_utc"`
	Source              File       `json:"source"`
	OriginalObservation File       `json:"original_observation"`
	OriginalFull        File       `json:"original_full_inventory"`
	OriginalProducer    File       `json:"original_coder_record_binding"`
	Observation         File       `json:"candidate_observation"`
	Full                File       `json:"candidate_full_inventory"`
	Pair                File       `json:"candidate_pair_binding"`
	PairScope           File       `json:"root_pair_scope"`
	AcceptanceScope     File       `json:"root_acceptance_scope"`
	Review              File       `json:"standard_review"`
	OwnInventory        File       `json:"own_precoder_inventory"`
	ReadStart           File       `json:"original_read_start"`
	ReadResult          File       `json:"original_read_result"`
	Start               File       `json:"acceptance_start"`
	Check               File       `json:"check_binding"`
	Report              File       `json:"check_report"`
	HeldReview          File       `json:"retained_held_review"`
	Rationale           File       `json:"semantic_rationale"`
	Copies              []CopyRead `json:"copy_read_receipts"`
	OfficialKeys        []string   `json:"complete_official_field_keys"`
	Scope               string     `json:"scope"`
	Verdict             string     `json:"verdict"`
	Unresolved          int        `json:"unresolved_count"`
	OldPass             bool       `json:"old_held_record_pass"`
	WholeQA             bool       `json:"whole_source_qa_pass"`
	Training            bool       `json:"training_eligible"`
}
type AmendedVersion struct {
	Schema              string `json:"schema"`
	ID                  string `json:"source_id"`
	Reviewer            string `json:"reviewer_id"`
	UTC                 string `json:"recorded_utc"`
	Selection           string `json:"selection"`
	Source              File   `json:"source"`
	OriginalObservation File   `json:"original_observation"`
	OriginalFull        File   `json:"original_full_inventory"`
	OriginalProducer    File   `json:"original_coder_record_binding"`
	Observation         File   `json:"candidate_observation"`
	Full                File   `json:"candidate_full_inventory"`
	Pair                File   `json:"candidate_pair_binding"`
	Review              File   `json:"standard_review"`
	Acceptance          File   `json:"full_inventory_acceptance"`
	HeldReview          File   `json:"retained_held_review"`
	PairScope           File   `json:"root_pair_scope"`
	AcceptanceScope     File   `json:"root_acceptance_scope"`
	Start               File   `json:"acceptance_start"`
	ReadStart           File   `json:"original_read_start"`
	ReadResult          File   `json:"original_read_result"`
	Check               File   `json:"check_binding"`
	Report              File   `json:"check_report"`
	OwnInventory        File   `json:"own_precoder_inventory"`
	Unchanged           bool   `json:"original_source_unchanged"`
	OldPass             bool   `json:"old_held_record_pass"`
	WholeQA             bool   `json:"whole_source_qa_pass"`
	Training            bool   `json:"training_eligible"`
}
type Disposition struct {
	Schema         string `json:"schema"`
	ID             string `json:"source_id"`
	Reviewer       string `json:"reviewer_id"`
	UTC            string `json:"recorded_utc"`
	Version        File   `json:"selected_version"`
	Review         File   `json:"selected_review"`
	SemanticHolds  int    `json:"current_semantic_holds"`
	TechnicalHolds int    `json:"current_technical_holds"`
	HeldReview     File   `json:"retained_held_review"`
	OldPass        bool   `json:"old_held_record_pass"`
	WholeQA        bool   `json:"whole_source_qa_pass"`
	Training       bool   `json:"training_eligible"`
}

// Anchors are caller declarations. Equal pins do not authenticate their target.
type Anchors struct {
	ID, ReviewerID, ProducerID, ProducerActor                                 string
	Source, SourceSchema, OriginalObservation, OriginalFull, OriginalProducer File
	SelectedObservation, SelectedFull, SelectedProducer                       File
	OwnInventory, ReadStart, ReadResult, Start, Check, Report, HeldReview     File
	Method, PairScope, AcceptanceScope                                        File
	OfficialKeys                                                              []string
}

// Optional witnesses use exactly the public readpacket/check/Review contracts.
// They must be supplied as bytes; no referenced path is opened by this package.
type Witnesses struct{ ReadStart, ReadResult, Check, Report, HeldReview PinnedBytes }
type Inputs struct {
	Kind                                     string
	Review, Acceptance, Version, Disposition PinnedBytes
	Expected                                 Anchors
	RegisteredSourceIDs                      []string
	Witnesses                                *Witnesses
}
type Selected struct {
	SourceID, ReviewerID, Kind, Verdict, Scope              string
	Source, Observation, Full, Producer                     File
	OriginalObservation, OriginalFull, OriginalProducer     File
	Review, Acceptance, Version, Disposition, HeldReview    File
	ReadStart, ReadResult, Check, Report                    File
	Unresolved, CurrentSemanticHolds, CurrentTechnicalHolds int
}
type Summary struct {
	State                       string `json:"state"`
	DeclaredVerdict             string `json:"declared_verdict"`
	HeldReviewReferenceRetained bool   `json:"held_review_reference_retained"`
	OfficialKeysVerified        bool   `json:"explicit_native_official_keys_verified"`
	WitnessContentVerified      bool   `json:"caller_supplied_witness_content_verified"`
	MeaningProven               bool   `json:"meaning_proven"`
	TrainingEligible            bool   `json:"training_eligible"`
	Limits                      string `json:"limits"`
}
type Result struct {
	Selected                                             Selected
	Review                                               sourcecohort.Review
	RawReview, RawAcceptance, RawVersion, RawDisposition PinnedBytes
	RawWitnesses                                         *Witnesses
	OriginalAcceptance                                   *OriginalAcceptance
	OriginalVersion                                      *OriginalVersion
	AmendedAcceptance                                    *AmendedAcceptance
	AmendedVersion                                       *AmendedVersion
	Disposition                                          Disposition
	Summary                                              Summary
}
type Error string

func (e Error) Error() string { return string(e) }
