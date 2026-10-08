// Package familycohort checks a supplied pre-reference metadata graph. Source,
// review, frame and schema pins are opaque declarations, not verified QA.
package familycohort

import "github.com/teamswyg/laya-tools/internal/sourcecohort"

const (
	Schema           = "riido-familycohort-bundle-v1"
	SourceCount      = 400
	FamilyCount      = 1200
	CommentCount     = 2400
	MaxInputBytes    = 8 << 20
	MaxCommentBytes  = 4096
	MaxArtifactBytes = 16 << 20
	MaxOutputBytes   = 64 << 20
)

type File = sourcecohort.File

type SourceBinding struct {
	ID           string `json:"source_id"`
	Split        string `json:"split"`
	DEVStyle     string `json:"dev_style"`
	Source       File   `json:"source"`
	SourceReview File   `json:"source_review"`
}

type FamilyBinding struct {
	ID          string `json:"family_id"`
	SourceID    string `json:"source_id"`
	Frame       File   `json:"frame"`
	FrameSchema File   `json:"frame_schema"`
}

// CompleteComment carries the entire decoded UTF-8 text. It has no labels,
// independently assigned partition, truncation option or semantic frame values.
type CompleteComment struct {
	ID       string `json:"comment_id"`
	FamilyID string `json:"family_id"`
	SourceID string `json:"source_id"`
	Source   File   `json:"source"`
	Frame    File   `json:"frame"`
	Locale   string `json:"locale"`
	Text     string `json:"text"`
}

type Bundle struct {
	Schema             string            `json:"schema"`
	FrameSchema        File              `json:"frame_schema"`
	FrameSchemaVersion string            `json:"frame_schema_version"`
	Sources            []SourceBinding   `json:"sources"`
	Families           []FamilyBinding   `json:"families"`
	Comments           []CompleteComment `json:"comments"`
}

type Counts struct {
	Name     string `json:"name"`
	Sources  int    `json:"source_bindings"`
	Families int    `json:"family_bindings"`
	Comments int    `json:"comments"`
}

// Summary contains aggregate structure and explicit pending checks only. A
// successful graph join never promotes any QA, rights or training claim.
type Summary struct {
	Schema               string   `json:"schema"`
	State                string   `json:"state"`
	Code                 string   `json:"code"`
	ExpectedSources      int      `json:"expected_source_bindings"`
	ExpectedFamilies     int      `json:"expected_family_bindings"`
	ExpectedComments     int      `json:"expected_comments"`
	Sources              int      `json:"source_bindings"`
	Families             int      `json:"family_bindings"`
	Comments             int      `json:"comments"`
	KORows               int      `json:"ko_comments"`
	ENRows               int      `json:"en_comments"`
	Splits               []Counts `json:"inherited_splits"`
	DEVStyles            []Counts `json:"inherited_dev_styles"`
	InputBytes           int64    `json:"encoded_bundle_bytes"`
	InputMaxBytes        int64    `json:"encoded_bundle_max_bytes"`
	CommentMaxBytes      int      `json:"decoded_comment_max_utf8_bytes"`
	MeaningProven        bool     `json:"source_meaning_proven"`
	TrainingEligible     bool     `json:"training_eligible"`
	ProviderIndependence string   `json:"provider_independence"`
	PendingChecks        []string `json:"pending_checks"`
	Limits               string   `json:"limits"`
}

type Structure struct {
	Schema  string  `json:"schema"`
	Input   File    `json:"input"`
	Bundle  File    `json:"retained_bundle"`
	Summary Summary `json:"summary"`
}

// Error never includes caller text, identifiers, paths or decoder diagnostics.
type Error string

func (e Error) Error() string { return string(e) }

func initial() Summary {
	return Summary{
		Schema: "riido-familycohort-summary-v1", State: "blocked",
		ExpectedSources: SourceCount, ExpectedFamilies: FamilyCount, ExpectedComments: CommentCount,
		InputMaxBytes: MaxInputBytes, CommentMaxBytes: MaxCommentBytes,
		ProviderIndependence: "unknown",
		Splits:               []Counts{{Name: "train"}, {Name: "dev"}, {Name: "cal"}, {Name: "test"}},
		DEVStyles:            []Counts{{Name: "short"}, {Name: "general"}},
		PendingChecks:        []string{"whole_source_qa", "source_read_provenance", "frame_schema_semantics", "meaning", "provider_identity_and_independence", "rights", "source_fidelity", "bilingual_fidelity", "naturalness", "blind_references", "reference_support", "all_workflow_gates"},
		Limits:               "Structural joins only. Referenced Source, review, frame and schema files are opaque, unopened declarations. Whole Source QA remains an external authoring prerequisite. All QA and semantic checks are pending. The 8 MiB encoded-bundle cap is separate from each complete comment's 4096-byte UTF-8 cap; oversized bundles are rejected whole without truncation or subset success.",
	}
}
