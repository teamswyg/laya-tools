// Package sourcecohort freezes whole registered cohorts from verified bytes and
// declared reviews. It does not establish source meaning or model eligibility.
package sourcecohort

type File struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}
type Split struct {
	Name       string `json:"name"`
	PerStratum int    `json:"per_stratum"`
}
type Recipe struct {
	Schema     string   `json:"schema"`
	Strata     []string `json:"strata"`
	Splits     []Split  `json:"splits"`
	DEVShort   int      `json:"dev_short"`
	DEVGeneral int      `json:"dev_general"`
}

func DefaultRecipe() Recipe {
	return Recipe{"riido-sourcecohort-recipe-v1", []string{"response_action", "time", "attribution_adoption", "negation_correction", "context_referents", "work_program_scope", "modality_hedging", "interactions"}, []Split{{"train", 35}, {"dev", 5}, {"cal", 5}, {"test", 5}}, 30, 10}
}

type Plan struct {
	Schema         string `json:"schema"`
	Registry       File   `json:"registry"`
	Creation       *File  `json:"creation"`
	Reviews        *File  `json:"reviews"`
	SourceSchema   *File  `json:"source_schema"`
	Allocation     *File  `json:"allocation"`
	SplitConfig    File   `json:"split_config"`
	AuthorID       string `json:"author_id"`
	CheckerID      string `json:"checker_id"`
	InputMaxBytes  int64  `json:"input_max_bytes"`
	SourceMaxBytes int64  `json:"source_max_bytes"`
	FreezeMaxBytes int64  `json:"freeze_max_bytes"`
	OutputMaxBytes int64  `json:"output_max_bytes"`
}
type Row struct {
	ID           string   `json:"source_id"`
	Stratum      string   `json:"stratum"`
	Split        string   `json:"split"`
	DEVStyle     string   `json:"dev_style"`
	Source       *File    `json:"source"`
	Dependencies []string `json:"dependencies"`
}
type Registry struct {
	Schema         string `json:"schema"`
	CohortID       string `json:"cohort_id"`
	IntendedLabels bool   `json:"intended_labels"`
	Rows           []Row  `json:"rows"`
}
type CreationRow struct {
	ID            string   `json:"source_id"`
	Source        File     `json:"source"`
	AuthorID      string   `json:"author_id"`
	CreatorKind   string   `json:"creator_kind"`
	RightsAllowed bool     `json:"rights_allowed"`
	CreatedUTC    string   `json:"created_utc"`
	Dependencies  []string `json:"dependencies"`
	Receipt       File     `json:"receipt"`
}
type CreationManifest struct {
	Schema string        `json:"schema"`
	Rows   []CreationRow `json:"rows"`
}
type CreationReceipt struct {
	Schema        string   `json:"schema"`
	SourceID      string   `json:"source_id"`
	Source        File     `json:"source"`
	AuthorID      string   `json:"author_id"`
	CreatorKind   string   `json:"creator_kind"`
	RightsAllowed bool     `json:"rights_allowed"`
	CreatedUTC    string   `json:"created_utc"`
	ConsultedNone bool     `json:"consulted_none"`
	Consulted     []File   `json:"consulted"`
	Dependencies  []string `json:"dependencies"`
}
type Span struct {
	Kind  string `json:"kind"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}
type Hold struct {
	Category string `json:"category"`
	Reason   string `json:"reason"`
}
type Review struct {
	ID              string   `json:"source_id"`
	Source          File     `json:"source"`
	SourceSchema    File     `json:"source_schema"`
	ReviewerID      string   `json:"reviewer_id"`
	ReviewedUTC     string   `json:"reviewed_utc"`
	Complete        bool     `json:"complete"`
	Structural      bool     `json:"structural"`
	SourceScope     string   `json:"source_scope"`
	SemanticVerdict string   `json:"independent_semantic_verdict"`
	Holds           []Hold   `json:"holds"`
	Evidence        []Span   `json:"evidence"`
	Dependencies    []string `json:"required_dependencies"`
	Observation     File     `json:"observation"`
	CheckBinding    File     `json:"check_binding"`
	ReadStart       File     `json:"read_start"`
	ReadResult      File     `json:"read_result"`
}
type Reviews struct {
	Schema string   `json:"schema"`
	Rows   []Review `json:"rows"`
}
type Count struct {
	Name string `json:"name"`
	Rows int    `json:"rows"`
}
type Summary struct {
	Schema               string  `json:"schema"`
	State                string  `json:"state"`
	Code                 string  `json:"code"`
	Expected             int     `json:"expected_rows"`
	Registered           int     `json:"registered_rows"`
	Reviewed             int     `json:"reviewed_rows"`
	Verified             int     `json:"verified_rows"`
	Held                 int     `json:"held_rows"`
	Unresolved           int     `json:"unresolved_rows"`
	Components           int     `json:"dependency_components"`
	Holds                []Count `json:"holds"`
	Strata               []Count `json:"strata"`
	Splits               []Count `json:"splits"`
	DEVStyles            []Count `json:"dev_styles"`
	MeaningProven        bool    `json:"source_meaning_proven"`
	ProviderIndependence string  `json:"provider_independence"`
	Limits               string  `json:"limits"`
}
type Joined struct {
	Registry        Row             `json:"registry"`
	Creation        CreationRow     `json:"creation"`
	Review          Review          `json:"review"`
	SourceText      string          `json:"source_text"`
	ObservationJSON string          `json:"observation_json"`
	CreationReceipt CreationReceipt `json:"creation_receipt"`
	CheckBinding    CheckBinding    `json:"check_binding"`
	CheckReport     CheckReport     `json:"check_report"`
}
type Allocation struct {
	Schema      string `json:"schema"`
	Registry    File   `json:"registry"`
	SplitConfig File   `json:"split_config"`
	CohortID    string `json:"cohort_id"`
	Rows        []Row  `json:"rows"`
}
type CheckBinding struct {
	Schema       string `json:"schema"`
	Source       File   `json:"source"`
	Observation  File   `json:"observation"`
	SourceSchema File   `json:"source_schema"`
	Report       File   `json:"report"`
	ExecutedUTC  string `json:"executed_utc"`
}
type CheckReport struct {
	StructuralValid   bool   `json:"structural_valid"`
	MandatoryComplete bool   `json:"mandatory_complete"`
	SourcePhaseScope  bool   `json:"source_phase_scope"`
	FailureCode       string `json:"failure_code,omitempty"`
	Scope             string `json:"scope"`
}
type Freeze struct {
	Schema   string   `json:"schema"`
	Plan     File     `json:"plan"`
	Inputs   Plan     `json:"inputs"`
	CohortID string   `json:"cohort_id"`
	Rows     []Joined `json:"rows"`
	Summary  Summary  `json:"summary"`
}
