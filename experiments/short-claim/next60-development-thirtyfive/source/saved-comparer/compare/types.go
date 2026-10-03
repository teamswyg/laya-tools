// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
// Saved primitive facts only; no original imports, getters, launch or model calls.
package compare

import (
	"encoding/json"
	core "riido.local/next60gjsonsavedcomparison/twoliteral"
)

const MaxPredicates = 17
const DataCap = 65536
const ReportCap = 1 << 20

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrPin        Error = "saved_compare_pin"
	ErrShape      Error = "saved_compare_shape"
	ErrCompletion Error = "saved_compare_incomplete"
	ErrBounds     Error = "saved_compare_bound"
	ErrIO         Error = "saved_compare_io"
)

type Pin struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type Config struct {
	Schema            string `json:"schema"`
	Outside           Pin    `json:"outside"`
	Child             Pin    `json:"child"`
	Fixtures          Pin    `json:"fixtures"`
	Wants             Pin    `json:"wants"`
	Captions          Pin    `json:"captions"`
	RootFreeze        Pin    `json:"Root_freeze"`
	Correction        Pin    `json:"correction"`
	Family            Pin    `json:"family"`
	CollectorExitCode *int   `json:"Root_collector_exit_code"`
	PlanSHA256        string `json:"plan_sha256"`
	WorkerSHA256      string `json:"worker_sha256"`
	ControllerSHA256  string `json:"controller_sha256"`
}
type Tally struct {
	T int `json:"satisfied"`
	F int `json:"unsatisfied"`
	U int `json:"unknown"`
}
type Predicate struct {
	Key    string          `json:"key"`
	Want   json.RawMessage `json:"Want"`
	Got    json.RawMessage `json:"Got"`
	State  string          `json:"state"`
	Reason string          `json:"reason"`
}
type RowResult struct {
	Ordinal          int                      `json:"ordinal"`
	Request          int                      `json:"request_index"`
	Fixture          int                      `json:"fixture_index"`
	Display          int                      `json:"display_position"`
	Internal         int                      `json:"internal_position"`
	RequestID        string                   `json:"request_id"`
	FixtureID        string                   `json:"fixture_id"`
	InputPointer     string                   `json:"input_pointer"`
	LiteralWant      json.RawMessage          `json:"full_literal_Want"`
	SavedGot         core.Got                 `json:"saved_primitive_Got"`
	Predicates       [MaxPredicates]Predicate `json:"predicates"`
	PredicateCount   int                      `json:"predicate_count"`
	Counts           Tally                    `json:"predicate_counts"`
	State            string                   `json:"state"`
	CallbackEvidence [3]core.Callback         `json:"callback_primitive_evidence"`
	CallbackCount    int                      `json:"callback_count"`
	MappingStopped   bool                     `json:"mapping_stopped"`
	Label            *bool                    `json:"label"`
	SampleWeight     *int                     `json:"sample_weight"`
}
type Candidate struct {
	Request                      int       `json:"request_index"`
	Display                      int       `json:"display_position"`
	Internal                     int       `json:"internal_position"`
	Rows                         int       `json:"finite_rows"`
	Vector                       [6]string `json:"row_vector"`
	Counts                       Tally     `json:"row_counts"`
	PredicateCounts              Tally     `json:"predicate_counts"`
	KnownCounterexampleFixtures  [6]int    `json:"known_counterexample_fixture_positions"`
	CounterexampleCount          int       `json:"counterexample_count"`
	UnknownFixtures              [6]int    `json:"aggregate_unknown_fixture_positions"`
	UnknownCount                 int       `json:"aggregate_unknown_count"`
	UnknownPredicateFixtures     [6]int    `json:"unknown_predicate_fixture_positions"`
	UnknownPredicateFixtureCount int       `json:"unknown_predicate_fixture_count"`
	Proposal                     string    `json:"finite_evidence_proposal"`
	Label                        *bool     `json:"label"`
	SampleWeight                 *int      `json:"sample_weight"`
	Role                         *string   `json:"role"`
}
type Report struct {
	Schema                 string          `json:"schema"`
	State                  string          `json:"state"`
	Scope                  string          `json:"scope"`
	OutsideSHA256          string          `json:"outside_sha256"`
	ChildSHA256            string          `json:"child_sha256"`
	PlanSHA256             string          `json:"plan_sha256"`
	RootFreezeSHA256       string          `json:"Root_freeze_sha256"`
	FixturesSHA256         string          `json:"fixtures_sha256"`
	WantsSHA256            string          `json:"Wants_sha256"`
	CaptionsSHA256         string          `json:"captions_sha256"`
	CorrectionSHA256       string          `json:"correction_sha256"`
	FamilySHA256           string          `json:"family_sha256"`
	WorkerSHA256           string          `json:"worker_sha256"`
	ControllerSHA256       string          `json:"controller_sha256"`
	Rows                   [33]RowResult   `json:"rows"`
	RowCounts              Tally           `json:"row_counts"`
	PredicateCounts        Tally           `json:"predicate_counts"`
	Candidates             [2][3]Candidate `json:"candidates"`
	ExplicitSavedCounters  core.Counts     `json:"explicit_saved_counter_scope"`
	FramesACKWritten       int             `json:"frames_ack_written"`
	JournalContentRead     bool            `json:"journal_content_read"`
	InitializationCalls    *int            `json:"initialization_calls"`
	NestedOriginalCalls    *int            `json:"nested_original_calls"`
	LabelsAssigned         int             `json:"labels_assigned"`
	WeightsAssigned        int             `json:"weights_assigned"`
	RolesAssigned          int             `json:"roles_assigned"`
	OriginalModelFitCalls  int             `json:"original_model_Fit_calls"`
	QualificationAuthority *bool           `json:"qualification_authority"`
}
type Process struct {
	StartAttempts             int    `json:"start_attempts"`
	Started                   bool   `json:"started"`
	WaitAttempts              int    `json:"wait_attempts"`
	Reaped                    *bool  `json:"reaped"`
	ExitCode                  *int   `json:"exit_code"`
	DeadlineExceeded          bool   `json:"deadline_exceeded"`
	StderrOverflow            bool   `json:"stderr_overflow"`
	WholeChildWallNanoseconds int64  `json:"whole_child_wall_nanoseconds"`
	DarwinTimeMaxRSSBytes     *int64 `json:"darwin_time_max_rss_bytes"`
	RSSCapExceeded            *bool  `json:"rss_cap_exceeded"`
	InitCalls                 any    `json:"initialization_calls"`
	NestedOriginalCalls       any    `json:"nested_original_calls"`
	Scope                     string `json:"scope"`
}
type Durable struct {
	FramesSynced int    `json:"frames_synced"`
	BytesSynced  int64  `json:"bytes_synced"`
	LastSHA256   string `json:"last_sha256"`
}
type Artifact struct {
	Name   string `json:"name"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type Outside struct {
	Schema                 string       `json:"schema"`
	State                  string       `json:"state"`
	Failure                string       `json:"failure"`
	ConfigSHA256           string       `json:"config_sha256"`
	PlanSHA256             string       `json:"plan_sha256"`
	WorkerSHA256           string       `json:"worker_sha256"`
	ControllerSHA256       string       `json:"controller_sha256"`
	Process                Process      `json:"process"`
	DurablePrefix          Durable      `json:"durable_prefix"`
	FramesACKWritten       int          `json:"frames_ack_written"`
	LastDurableCounts      *core.Counts `json:"last_durable_counts"`
	LastACKWrittenCounts   *core.Counts `json:"last_ack_written_counts"`
	TerminalRowsACKWritten int          `json:"terminal_rows_ack_written"`
	FinalACKWritten        bool         `json:"final_ack_written"`
	FinalChildFileIdentity bool         `json:"final_child_file_identity"`
	CompleteCountersKnown  bool         `json:"complete_counters_known"`
	CompleteRows           *int         `json:"complete_rows"`
	RemainingCallsUnknown  bool         `json:"remaining_calls_unknown"`
	InitializationCalls    *int         `json:"initialization_calls"`
	NestedOriginalCalls    *int         `json:"nested_original_calls"`
	Authority              *bool        `json:"source_rights_compiler_resource_authority"`
	TruthAssigned          int          `json:"truth_assigned"`
	LabelsAssigned         int          `json:"labels_assigned"`
	FitOrModelCalls        int          `json:"fit_or_model_calls"`
	Artifacts              []Artifact   `json:"artifacts"`
	OutputBytesRetained    *int64       `json:"output_bytes_retained"`
	AggregateCap           int          `json:"aggregate_cap_bytes"`
	NoAutomaticRetry       bool         `json:"no_automatic_retry"`
}
