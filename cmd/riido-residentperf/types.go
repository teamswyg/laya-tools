// Resource records describe public development protocol observations only.
// They do not represent semantic truth, training samples, or model savings.
package main

const (
	planSchema     = "riido-resident-resource-plan-v1"
	planState      = "uncached_public_jsonl_resource_only"
	corpusSchema   = "riido-resident-public-wire-corpus-v2"
	resultSchema   = "riido-resident-resource-result-v1"
	policy         = "public_original_inputs_no_truth_no_models_no_fits_no_cache_no_retry"
	goVersion      = "go1.27.1"
	legacySHA      = "0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0"
	typedSHA       = "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"
	rowCount       = 24
	requestsPerRow = 1045
	timedPerRow    = 1024
	resultLimit    = 8 << 20
	wireLimit      = 12 << 10
	textSchema     = "riido-short-behavior-claim-v1"
	responseSchema = "riido-shortclaim-order-v1"
	provenance     = "resident56d-original-public"
)

func baselineKinds() [4]string {
	return [4]string{"fixed_order", "bm25", "lexical_ordered", "narrow_rule"}
}
func candidateIDs() [3]string { return [3]string{"candidate-0", "candidate-1", "candidate-2"} }

type FilePin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type RowPlan struct {
	Baseline string `json:"baseline"`
	Workload string `json:"workload"`
	Repeat   int    `json:"repeat"`
}

// The plan is authored and hash-frozen separately. No draft code manufactures
// an execution-authorizing plan or fills in unverified artifact hashes.
type Plan struct {
	Schema                   string            `json:"schema"`
	State                    string            `json:"state"`
	Policy                   string            `json:"policy"`
	SourceCommit             string            `json:"preparation_source_commit"`
	Go                       string            `json:"runtime_go"`
	OS                       string            `json:"os"`
	Arch                     string            `json:"arch"`
	ChildBinarySHA256        string            `json:"child_binary_sha256"`
	DriverBinarySHA256       string            `json:"driver_binary_sha256"`
	BuildRecipeSHA256        string            `json:"build_recipe_sha256"`
	CorpusSHA256             string            `json:"corpus_sha256"`
	LegacyInputSHA256        string            `json:"legacy_input_sha256"`
	TypedInputSHA256         string            `json:"typed_input_sha256"`
	ImplementationFiles      []FilePin         `json:"implementation_files"`
	Rows                     [rowCount]RowPlan `json:"rows"`
	FirstRequests            int               `json:"first_requests"`
	WarmupRequests           int               `json:"warmup_requests"`
	TimedRequests            int               `json:"timed_requests"`
	CPUThreads               int               `json:"cpu_threads"`
	HeapSoftLimitBytes       int64             `json:"heap_soft_limit_bytes"`
	ChildTimeoutSeconds      int               `json:"child_timeout_seconds"`
	GlobalTimeoutSeconds     int               `json:"global_timeout_seconds"`
	CleanupGraceMilliseconds int               `json:"cleanup_grace_milliseconds"`
	Retries                  int               `json:"retries"`
	PublicResultLimitBytes   int               `json:"public_result_limit_bytes"`
	InFlight                 int               `json:"inflight"`
	StopOnFirstFailedRow     bool              `json:"stop_on_first_failed_row"`
}

type ScoredCandidate struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

// Field order and encoder settings deliberately match the frozen child CLI.
// This is a protocol expectation, never an independent semantic truth label.
type WireOutput struct {
	Schema         string            `json:"schema"`
	Status         string            `json:"status"`
	Baseline       string            `json:"baseline"`
	InputSHA256    string            `json:"input_sha256"`
	FallbackReason string            `json:"fallback_reason,omitempty"`
	Candidates     []ScoredCandidate `json:"verification_order"`
}

type ExpectedResponse struct {
	Baseline       string `json:"baseline"`
	RawSHA256      string `json:"raw_sha256"`
	RawBytes       int    `json:"raw_bytes"`
	FallbackReason string `json:"fallback_reason"`
}

type Payload struct {
	Wire                []byte              `json:"wire_base64"` // exact compact JSON bytes; LF is added once
	FeatureSHA256       string              `json:"feature_sha256"`
	WireSHA256          string              `json:"wire_sha256"` // includes exactly one LF
	InputDigest         string              `json:"input_digest"`
	RawTextBytes        [4]int              `json:"raw_text_bytes"`
	NormalizedTextBytes [4]int              `json:"normalized_text_bytes"`
	NormalizedWords     [4]int              `json:"normalized_words"`
	Expected            [4]ExpectedResponse `json:"expected"`
}

type OriginLink struct {
	Fixture              string    `json:"fixture"` // public names, never local paths
	ParentID             string    `json:"parent_id"`
	OriginalIndex        int       `json:"original_index"`
	CandidateCount       int       `json:"candidate_count"`
	PayloadIndex         int       `json:"payload_index"` // -1 only for unaltered non-3 sets
	ExcludedReason       string    `json:"excluded_reason,omitempty"`
	FeatureSHA256        string    `json:"feature_sha256,omitempty"`
	OriginalCandidateIDs [8]string `json:"original_candidate_ids"`
}

type Corpus struct {
	Schema            string       `json:"schema"`
	State             string       `json:"state"`
	OS                string       `json:"os"`
	Arch              string       `json:"arch"`
	LegacyInputSHA256 string       `json:"legacy_input_sha256"`
	TypedInputSHA256  string       `json:"typed_input_sha256"`
	FeatureEncoding   string       `json:"feature_encoding"`
	WireEncoding      string       `json:"wire_encoding"`
	OriginalParents   int          `json:"original_parents"`
	Payloads          []Payload    `json:"payloads"`
	Origins           []OriginLink `json:"origins"`
}

type PhaseCounts struct {
	Planned      int `json:"planned"`
	Attempted    int `json:"attempted"`
	FullyWritten int `json:"fully_written"`
	Received     int `json:"received"`
	Validated    int `json:"validated"`
	Failed       int `json:"failed"`
	Incomplete   int `json:"incomplete"`
	NotAttempted int `json:"not_attempted"`
}

type Summary struct {
	Count int   `json:"count"`
	P50NS int64 `json:"p50_ns"`
	P95NS int64 `json:"p95_ns"`
	MaxNS int64 `json:"max_ns"`
}

type Resource struct {
	Available            bool    `json:"available"`
	CPUAvailable         bool    `json:"cpu_available"`
	RSSAvailable         bool    `json:"rss_available"`
	CPUError             string  `json:"cpu_error,omitempty"`
	RSSError             string  `json:"rss_error,omitempty"`
	UserCPUSeconds       float64 `json:"user_cpu_seconds"`
	SystemCPUSeconds     float64 `json:"system_cpu_seconds"`
	LifetimePeakRSSRaw   int64   `json:"lifetime_peak_rss_raw"`
	LifetimePeakRSSUnit  string  `json:"lifetime_peak_rss_unit"`
	LifetimePeakRSSBytes int64   `json:"lifetime_peak_rss_bytes"`
}

type RowResult struct {
	Index                      int            `json:"index"`
	Plan                       RowPlan        `json:"plan"`
	Status                     string         `json:"status"`
	Error                      string         `json:"error,omitempty"`
	Phases                     [3]PhaseCounts `json:"phases"` // first, warmup, timed
	Started                    bool           `json:"started"`
	ExpectedHashPreparationNS  int64          `json:"expected_hash_preparation_ns"`
	SpawnCallNS                int64          `json:"spawn_call_ns"`
	StartupAndFirstResponseNS  int64          `json:"startup_and_first_response_ns"`
	LifetimeWallNS             int64          `json:"lifetime_wall_ns"`
	TimedRTT                   Summary        `json:"timed_wire_rtt"`
	TimedWrite                 Summary        `json:"timed_write_blocking"`
	TimedValidation            Summary        `json:"timed_controller_validation"`
	TimedValidationWork        Summary        `json:"timed_controller_validation_work"`
	TimedResponseDeliveryGap   Summary        `json:"timed_response_delivery_gap"`
	InputBytes                 int64          `json:"input_bytes"`
	OutputBytes                int64          `json:"output_bytes"`
	StderrBytes                int64          `json:"stderr_bytes"`
	ExpectedInputStreamSHA256  string         `json:"expected_input_stream_sha256"`
	ExpectedOutputStreamSHA256 string         `json:"expected_output_stream_sha256"`
	ActualInputStreamSHA256    string         `json:"actual_input_stream_sha256"`
	ActualOutputStreamSHA256   string         `json:"actual_output_stream_sha256"`
	EOFObserved                bool           `json:"eof_observed"`
	StdoutJoined               bool           `json:"stdout_joined"`
	StdoutReceiptAvailable     bool           `json:"stdout_receipt_available"`
	StdoutError                string         `json:"stdout_error,omitempty"`
	StderrJoined               bool           `json:"stderr_joined"`
	StderrReceiptAvailable     bool           `json:"stderr_receipt_available"`
	StderrError                string         `json:"stderr_error,omitempty"`
	StderrLimitExceeded        bool           `json:"stderr_limit_exceeded"`
	WatcherJoined              bool           `json:"watcher_joined"`
	CleanupComplete            bool           `json:"cleanup_complete"`
	CleanupWallNS              int64          `json:"cleanup_wall_ns"`
	CleanupError               string         `json:"cleanup_error,omitempty"`
	WaitCalls                  int            `json:"wait_calls"`
	WaitError                  string         `json:"wait_error,omitempty"`
	WaitObservedWallNS         int64          `json:"wait_observed_wall_ns"`
	Reaped                     bool           `json:"reaped"`
	ExitAvailable              bool           `json:"exit_available"`
	ExitCode                   int            `json:"exit_code"`
	Cancelled                  bool           `json:"cancelled"`
	ChildDeadlineExceeded      bool           `json:"child_deadline_exceeded"`
	ContextDeadlineExceeded    bool           `json:"context_deadline_exceeded"`
	Resource                   Resource       `json:"child_lifetime_resource"`
	CPUPerRequestAvailable     bool           `json:"cpu_per_request_available"`
	CPURequestDenominator      int            `json:"cpu_request_denominator"`
	CPUSecondsPerRequest       float64        `json:"cpu_seconds_per_request"`
}

type Report struct {
	Schema                 string              `json:"schema"`
	PlanSHA256             string              `json:"plan_sha256"`
	SourceCommit           string              `json:"preparation_source_commit"`
	CorpusSHA256           string              `json:"corpus_sha256"`
	ChildBinarySHA256      string              `json:"child_binary_sha256"`
	DriverBinarySHA256     string              `json:"driver_binary_sha256"`
	OS                     string              `json:"os"`
	Arch                   string              `json:"arch"`
	Go                     string              `json:"go"`
	PlannedRows            int                 `json:"planned_rows"`
	PlannedRequests        int                 `json:"planned_requests"`
	UniqueFeaturePayloads  int                 `json:"unique_feature_payloads"`
	Rows                   [rowCount]RowResult `json:"rows"`
	ReplayWallNS           int64               `json:"replay_wall_ns"`
	GlobalDeadlineExceeded bool                `json:"global_deadline_exceeded"`
	Controller             Resource            `json:"controller_replay_resource"`
	ControllerRSSScope     string              `json:"controller_rss_scope"`
	PureStartup            string              `json:"pure_startup"`
	PureChildProcessing    string              `json:"pure_child_processing"`
	PurePipe               string              `json:"pure_pipe"`
	Fits                   int                 `json:"fits"`
	Models                 int                 `json:"model_calls"`
	ProtectedFinalReads    int                 `json:"protected_final_reads"`
	CacheHits              int                 `json:"result_cache_hits"`
	SavingsClaim           bool                `json:"savings_claim"`
	Scope                  string              `json:"scope"`
}
