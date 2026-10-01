// Package taskrun owns one explicit, bounded maintainer attempt on a pinned
// public task. Its reports never contain prompts, paths, credentials or traces.
package taskrun

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/taskverify"
	"github.com/teamswyg/laya-tools/pkg/taskoutcome"
)

const Schema = "riido-owned-task-attempt-v1"
const SupportedVersion = "codex-cli 0.158.0"
const DefaultTimeout = 120 * time.Second
const MaxTimeout = 180 * time.Second
const MaxStderrBytes = 1 << 20
const maxBinaryBytes = 256 << 20

// Request describes a single requested attempt; none of these names claim a
// provider-observed model identity. Execute is mandatory even through the API.
type Request struct {
	Execute            bool
	TaskID             string
	Model              string
	Reasoning          string
	BaseDir            string
	PrivateDir         string
	CodexBinary        string
	AuthSourceDir      string
	PlanSHA256         string
	PlanFile           string
	AttemptOrdinal     int
	ExpectedSpecSHA256 string
	ExpectedCLIHash    string
	ExpectedCLIVersion string
	GoRoot             string
	Timeout            time.Duration
}

type Capture struct {
	SHA256        string `json:"sha256"`
	Bytes         int64  `json:"bytes"`
	RetainedBytes int64  `json:"retained_bytes"`
	Complete      bool   `json:"complete"`
	Status        string `json:"status"`
}

type AppliedRequest struct {
	Model      string `json:"model"`
	Reasoning  string `json:"reasoning"`
	Evidence   string `json:"evidence"`
	Profile    string `json:"permissions_profile"`
	Approvals  string `json:"approval_policy"`
	Network    bool   `json:"model_command_network"`
	RetryLimit int    `json:"executor_retries"`
}

type Record struct {
	Schema                    string                `json:"schema"`
	PlanSHA256                string                `json:"plan_sha256"`
	PlanEvidence              string                `json:"plan_evidence"`
	AttemptOrdinal            int                   `json:"attempt_ordinal"`
	ProfileID                 string                `json:"profile_id"`
	Provenance                string                `json:"provenance"`
	TaskID                    string                `json:"task_id"`
	TaskSpecSHA256            string                `json:"task_spec_sha256"`
	PromptSHA256              string                `json:"prompt_sha256"`
	BaseRevision              string                `json:"base_revision"`
	BaseSHA256                string                `json:"base_sha256"`
	WorkspaceSHA256           string                `json:"initial_workspace_sha256"`
	CandidateSHA256           string                `json:"candidate_sha256,omitempty"`
	CandidateStatus           string                `json:"candidate_status"`
	CandidateFiles            []taskverify.FileHash `json:"candidate_files,omitempty"`
	CandidateFileCount        int                   `json:"candidate_file_count"`
	OutsideClosure            string                `json:"outside_closure"`
	CodexVersion              string                `json:"codex_version"`
	ExecutableSHA256          string                `json:"executable_sha256"`
	ExecutableEvidence        string                `json:"executable_evidence"`
	GoVersion                 string                `json:"go_version"`
	GoBinarySHA256            string                `json:"go_binary_sha256"`
	InvocationSHA256          string                `json:"canonical_invocation_sha256"`
	PermissionsSHA256         string                `json:"canonical_permissions_sha256"`
	PermissionsProbe          string                `json:"permissions_probe"`
	Applied                   AppliedRequest        `json:"applied_request"`
	ObservedModel             string                `json:"observed_model"`
	AttemptScope              string                `json:"attempt_scope"`
	Started                   bool                  `json:"started"`
	ExitCode                  *int                  `json:"exit_code"`
	ProcessStatus             string                `json:"process_status"`
	WallMillis                int64                 `json:"wall_millis"`
	GroupCleanup              string                `json:"process_group_cleanup"`
	AuthCleanup               string                `json:"auth_cleanup"`
	DurableStart              bool                  `json:"durable_start_record"`
	Stdout                    Capture               `json:"stdout"`
	Stderr                    Capture               `json:"stderr"`
	SummaryStatus             string                `json:"summary_status"`
	WholeAttemptUsageComplete bool                  `json:"whole_attempt_usage_complete"`
	Summary                   *taskoutcome.Summary  `json:"usage_summary,omitempty"`
	Verification              taskverify.Report     `json:"verification"`
	VerificationStatus        string                `json:"verification_status"`
	VerificationWallMillis    *int64                `json:"verification_wall_millis"`
}

// Error is a fixed enum; filesystem and subprocess errors never escape.
type Error string

func (e Error) Error() string { return string(e) }

func digest(b []byte) string  { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func digestJSON(v any) string { b, _ := json.Marshal(v); return digest(b) }

func validateRequest(req Request) error {
	if !req.Execute {
		return Error("execute_opt_in_required")
	}
	if _, err := taskverify.TaskSpec(req.TaskID); err != nil {
		return Error("invalid_task")
	}
	if req.Model == "" || req.Reasoning == "" {
		return Error("explicit_model_and_reasoning_required")
	}
	if b, err := hex.DecodeString(req.PlanSHA256); err != nil || len(b) != sha256.Size || strings.ToLower(req.PlanSHA256) != req.PlanSHA256 {
		return Error("explicit_plan_sha256_required")
	}
	if b, err := hex.DecodeString(req.ExpectedCLIHash); err != nil || len(b) != sha256.Size || strings.ToLower(req.ExpectedCLIHash) != req.ExpectedCLIHash || req.ExpectedCLIVersion != SupportedVersion {
		return Error("explicit_supported_cli_pins_required")
	}
	if b, err := hex.DecodeString(req.ExpectedSpecSHA256); err != nil || len(b) != sha256.Size || strings.ToLower(req.ExpectedSpecSHA256) != req.ExpectedSpecSHA256 {
		return Error("explicit_task_spec_pin_required")
	}
	if _, err := taskoutcome.Summarize(strings.NewReader(""), taskoutcome.Metadata{PublicTaskLabel: req.TaskID, RequestedModel: req.Model, RequestedReasoning: req.Reasoning}); err != nil {
		return Error("invalid_profile_identifier")
	}
	if !slices.Contains([]string{"minimal", "low", "medium", "high", "xhigh"}, req.Reasoning) {
		return Error("unsupported_reasoning")
	}
	if req.Timeout <= 0 || req.Timeout > MaxTimeout {
		return Error("invalid_timeout")
	}
	for _, p := range []string{req.BaseDir, req.PrivateDir, req.CodexBinary, req.PlanFile} {
		if p == "" || !filepath.IsAbs(p) || strings.ContainsAny(p, "\x00\r\n") {
			return Error("absolute_paths_required")
		}
	}
	if req.AuthSourceDir != "" && (!filepath.IsAbs(req.AuthSourceDir) || strings.ContainsAny(req.AuthSourceDir, "\x00\r\n")) {
		return Error("invalid_auth_source")
	}
	if req.GoRoot != "" && (!filepath.IsAbs(req.GoRoot) || strings.ContainsAny(req.GoRoot, "\x00\r\n")) {
		return Error("invalid_go_root")
	}
	return nil
}

// LoadBase reads only the immutable public closure and attribution files. The
// verifier validates closure pins before the executor can launch anything.
func LoadBase(dir, task string) ([]taskverify.BaseFile, []taskverify.BaseFile, error) {
	spec, err := taskverify.TaskSpec(task)
	if err != nil {
		return nil, nil, Error("invalid_task")
	}
	paths, err := taskverify.BasePaths(task)
	if err != nil {
		return nil, nil, Error("invalid_task")
	}
	attribution, err := taskverify.AttributionPins(task)
	if err != nil || len(attribution) == 0 {
		return nil, nil, Error("invalid_attribution_manifest")
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, nil, Error("base_root_unavailable")
	}
	defer root.Close()
	stagingPaths := slices.Clone(paths)
	for _, pin := range attribution {
		stagingPaths = append(stagingPaths, pin.Path)
	}
	all := make([]taskverify.BaseFile, 0, len(stagingPaths))
	for _, path := range stagingPaths {
		b, err := readRootFile(root, path, taskverify.MaxFileBytes)
		if err != nil {
			return nil, nil, Error("base_file_unavailable")
		}
		all = append(all, taskverify.BaseFile{Path: path, SHA256: digest(b), Data: b})
	}
	for i, pin := range attribution {
		if all[len(paths)+i].SHA256 != pin.SHA256 {
			return nil, nil, Error("attribution_pin_mismatch")
		}
	}
	base := all[:len(paths)]
	if _, err = taskverify.Verify(context.Background(), taskverify.Request{TaskID: task, BaseRevision: spec.BaseRevision, BaseFiles: base, AttributionFiles: all[len(paths):], CandidateDir: dir}); err != nil {
		return nil, nil, Error("base_pin_mismatch")
	}
	return base, all, nil
}

func readRootFile(root *os.Root, path string, limit int64) ([]byte, error) {
	parts := strings.Split(path, "/")
	for i := range parts {
		st, e := root.Lstat(strings.Join(parts[:i+1], "/"))
		if e != nil || st.Mode()&os.ModeSymlink != 0 || (i+1 < len(parts) && !st.IsDir()) {
			return nil, Error("file_unavailable")
		}
	}
	f, e := root.Open(path)
	if e != nil {
		return nil, Error("file_unavailable")
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || !st.Mode().IsRegular() || st.Size() > limit {
		return nil, Error("file_limit")
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit {
		return nil, Error("file_limit")
	}
	return b, nil
}

func writeFiles(dir string, files []taskverify.BaseFile) error {
	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f.Path))
		if os.MkdirAll(filepath.Dir(path), 0700) != nil || os.WriteFile(path, f.Data, 0600) != nil {
			return Error("private_stage_failed")
		}
	}
	return nil
}
func hashes(files []taskverify.BaseFile) []taskverify.FileHash {
	out := make([]taskverify.FileHash, len(files))
	for i, f := range files {
		out[i] = taskverify.FileHash{Path: f.Path, SHA256: digest(f.Data)}
	}
	slices.SortFunc(out, func(a, b taskverify.FileHash) int { return strings.Compare(a.Path, b.Path) })
	return out
}

func stageAuth(source, home string) error {
	if source == "" {
		return nil
	}
	r, e := os.OpenRoot(source)
	if e != nil {
		return Error("auth_source_unavailable")
	}
	defer r.Close()
	b, e := readRootFile(r, "auth.json", 1<<20)
	if e != nil {
		return Error("auth_source_unavailable")
	}
	var a struct {
		Mode   string          `json:"auth_mode"`
		APIKey *string         `json:"OPENAI_API_KEY"`
		Tokens json.RawMessage `json:"tokens"`
	}
	if json.Unmarshal(b, &a) != nil || a.Mode != "chatgpt" || (a.APIKey != nil && *a.APIKey != "") {
		return Error("chatgpt_auth_required")
	}
	var tokens struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		ID      string `json:"id_token"`
	}
	if json.Unmarshal(a.Tokens, &tokens) != nil || tokens.Access == "" || tokens.Refresh == "" || tokens.ID == "" {
		return Error("chatgpt_auth_required")
	}
	// Copy only a known schema; user-supplied extra auth-file keys never propagate.
	var known struct {
		Mode        string          `json:"auth_mode"`
		Tokens      json.RawMessage `json:"tokens"`
		LastRefresh json.RawMessage `json:"last_refresh,omitempty"`
	}
	if json.Unmarshal(b, &known) != nil {
		return Error("chatgpt_auth_required")
	}
	canonical, e := json.Marshal(known)
	if e != nil || os.WriteFile(filepath.Join(home, "auth.json"), canonical, 0600) != nil {
		return Error("auth_stage_failed")
	}
	return nil
}

func ownedEnvironment(private, home, workspace, goRoot string) []string {
	// No inherited PATH, user config, API keys, token variables or shell startup.
	return []string{"HOME=" + filepath.Join(private, "home"), "CODEX_HOME=" + home, "TMPDIR=" + filepath.Join(private, "tmp"), "PATH=" + filepath.Join(goRoot, "bin") + ":/usr/bin:/bin:/usr/sbin:/sbin", "GOROOT=" + goRoot, "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8", "NO_COLOR=1", "TERM=dumb", "GOTOOLCHAIN=local", "GOPROXY=off", "GOSUMDB=off", "CGO_ENABLED=0", "GOCACHE=" + filepath.Join(workspace, ".gocache")}
}

func executableIdentity(binary string, env []string, expected string) (string, string, error) {
	sha, e := executableDigest(binary)
	if e != nil {
		return "", "", e
	}
	if sha != expected {
		return "", "", Error("trusted_executable_pin_mismatch")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--version")
	cmd.Env = env
	configureProcess(cmd)
	cmd.WaitDelay = time.Second
	var output boundedBuffer
	output.limit = 1024
	cmd.Stdout = &output
	cmd.Stderr = &output
	versionErr := cmd.Run()
	cleanupProcess(cmd)
	if versionErr != nil || output.exceeded {
		return "", "", Error("executable_version_unavailable")
	}
	version := strings.TrimSpace(output.buf.String())
	if version != SupportedVersion {
		return "", "", Error("unsupported_codex_version")
	}
	return sha, version, nil
}

func executableDigest(binary string) (string, error) {
	st, e := os.Lstat(binary)
	if e != nil || !st.Mode().IsRegular() || st.Mode()&0111 == 0 || st.Size() > maxBinaryBytes {
		return "", Error("trusted_executable_unavailable")
	}
	f, e := os.Open(binary)
	if e != nil {
		return "", Error("trusted_executable_unavailable")
	}
	h := sha256.New()
	_, e = io.Copy(h, io.LimitReader(f, maxBinaryBytes+1))
	f.Close()
	if e != nil {
		return "", Error("executable_hash_failed")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type canonicalInvocation struct {
	Args          []string `json:"args"`
	Task          string   `json:"task"`
	Model         string   `json:"model"`
	Reasoning     string   `json:"reasoning"`
	Version       string   `json:"version"`
	Executable    string   `json:"executable_sha256"`
	Prompt        string   `json:"prompt_sha256"`
	Workspace     string   `json:"workspace_sha256"`
	TimeoutMillis int64    `json:"timeout_millis"`
	Permissions   string   `json:"permissions_sha256"`
	Plan          string   `json:"plan_sha256"`
	GoVersion     string   `json:"go_version"`
	GoBinary      string   `json:"go_binary_sha256"`
}

// Run never retries, resumes or falls back. An error means no main attempt was
// started. After Start succeeds, all failures are represented in its record.
func Run(ctx context.Context, req Request) (Record, error) {
	if req.Timeout == 0 {
		req.Timeout = DefaultTimeout
	}
	if ctx == nil {
		return Record{}, Error("invalid_context")
	}
	if e := validateRequest(req); e != nil {
		return Record{}, e
	}
	if !executionSupported() {
		return Record{}, Error("execution_isolation_unavailable")
	}
	toolchain, e := resolveGoRoot(ctx, req.GoRoot)
	if e != nil {
		return Record{}, e
	}
	req.GoRoot = toolchain.root
	resolvedBinary, e := filepath.EvalSymlinks(req.CodexBinary)
	if e != nil {
		return Record{}, Error("trusted_executable_unavailable")
	}
	req.CodexBinary = resolvedBinary
	private, e := isolatedPrivatePath(req.PrivateDir)
	if e != nil {
		return Record{}, e
	}
	req.PrivateDir = private
	profileID, e := validatePlan(req, toolchain)
	if e != nil {
		return Record{}, e
	}
	base, all, e := LoadBase(req.BaseDir, req.TaskID)
	if e != nil {
		return Record{}, e
	}
	// Refuse existing paths, including symlinks. Caller owns and retains all
	// private traces for review; the executor never reuses or deletes a run.
	if e = os.Mkdir(req.PrivateDir, 0700); e != nil {
		return Record{}, Error("private_directory_must_be_new")
	}
	for _, sub := range []string{"home", "codex-home", "tmp", "workspace", "candidate"} {
		if os.Mkdir(filepath.Join(req.PrivateDir, sub), 0700) != nil {
			return Record{}, Error("private_stage_failed")
		}
	}
	workspace := filepath.Join(req.PrivateDir, "workspace")
	candidate := filepath.Join(req.PrivateDir, "candidate")
	home := filepath.Join(req.PrivateDir, "codex-home")
	if e = writeFiles(workspace, all); e != nil {
		return Record{}, e
	}
	defer os.Remove(filepath.Join(home, "auth.json"))
	if e = stageAuth(req.AuthSourceDir, home); e != nil {
		return Record{}, e
	}
	env := ownedEnvironment(req.PrivateDir, home, workspace, req.GoRoot)
	executableSHA, version, e := executableIdentity(req.CodexBinary, env, req.ExpectedCLIHash)
	if e != nil {
		return Record{}, e
	}
	if executableSHA != req.ExpectedCLIHash || version != req.ExpectedCLIVersion {
		return Record{}, Error("trusted_executable_pin_mismatch")
	}
	spec, _ := taskverify.TaskSpec(req.TaskID)
	profile, canonicalPermissions, e := permissionsConfig(req.GoRoot)
	if e != nil {
		return Record{}, e
	}
	if e = permissionProbe(ctx, req.CodexBinary, workspace, req.PrivateDir, home, env, profile, req.GoRoot); e != nil {
		return Record{}, e
	}
	if e = os.MkdirAll(filepath.Join(workspace, ".riido-runtime", "tmp"), 0700); e != nil {
		return Record{}, Error("private_stage_failed")
	}
	if e = os.MkdirAll(filepath.Join(workspace, ".riido-runtime", "home"), 0700); e != nil {
		return Record{}, Error("private_stage_failed")
	}
	args := invocationArgs(req.Model, req.Reasoning, workspace, profile, req.GoRoot)
	r := Record{Schema: Schema, PlanSHA256: req.PlanSHA256, PlanEvidence: "supplied_plan_bytes_and_ordered_attempt_validated_precommit_not_attested", AttemptOrdinal: req.AttemptOrdinal, ProfileID: profileID, Provenance: "executor_owned_single_attempt", TaskID: req.TaskID, TaskSpecSHA256: digestJSON(spec), PromptSHA256: digest([]byte(spec.Prompt)), BaseRevision: spec.BaseRevision, BaseSHA256: digestJSON(hashes(base)), WorkspaceSHA256: digestJSON(hashes(all)), CodexVersion: version, ExecutableSHA256: executableSHA, PermissionsSHA256: digest([]byte(canonicalPermissions)), PermissionsProbe: "workspace_allowed_sibling_read_write_and_network_denied", ObservedModel: "unknown", AttemptScope: "one_owned_process_no_retries_other_attempts_unassessed", SummaryStatus: "unknown", CandidateStatus: "unknown", VerificationStatus: "unknown", OutsideClosure: "unassessed_including_runtime_cache_and_added_files"}
	r.Applied = AppliedRequest{Model: req.Model, Reasoning: req.Reasoning, Evidence: "explicit_executor_request_not_provider_identity", Profile: profileName, Approvals: "never"}
	r.AuthCleanup = "pending"
	r.ExecutableEvidence = "resolved_path_and_prelaunch_byte_pins_not_host_attestation"
	r.GoVersion = toolchain.version
	r.GoBinarySHA256 = toolchain.hash
	r.InvocationSHA256 = digestJSON(canonicalInvocation{Args: invocationArgs(req.Model, req.Reasoning, "$WORKSPACE", canonicalPermissions, "$GOROOT"), Task: req.TaskID, Model: req.Model, Reasoning: req.Reasoning, Version: version, Executable: executableSHA, Prompt: r.PromptSHA256, Workspace: r.WorkspaceSHA256, TimeoutMillis: req.Timeout.Milliseconds(), Permissions: r.PermissionsSHA256, Plan: req.PlanSHA256, GoVersion: toolchain.version, GoBinary: toolchain.hash})
	childCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()
	cmd := exec.CommandContext(childCtx, req.CodexBinary, args...)
	cmd.Env = env
	cmd.Dir = workspace
	cmd.Stdin = strings.NewReader(spec.Prompt)
	configureProcess(cmd)
	stdout, e := newCapture(filepath.Join(req.PrivateDir, "stdout.jsonl"), taskoutcome.MaxTraceBytes, cancel)
	if e != nil {
		return Record{}, e
	}
	defer stdout.file.Close()
	stderr, e := newCapture(filepath.Join(req.PrivateDir, "stderr.txt"), MaxStderrBytes, cancel)
	if e != nil {
		return Record{}, e
	}
	defer stderr.file.Close()
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = 2 * time.Second
	// Trust requires a locally immutable supplied executable. Detect ordinary
	// updates during preflight too; this is not an attestation against a hostile
	// host replacing executable bytes between this check and the OS exec call.
	if current, e := executableDigest(req.CodexBinary); e != nil || current != req.ExpectedCLIHash {
		return Record{}, Error("trusted_executable_changed_during_preflight")
	}
	start := time.Now()
	if e = cmd.Start(); e != nil {
		return Record{}, Error("attempt_start_failed")
	}
	r.Started = true
	r.ProcessStatus = "running"
	r.GroupCleanup = "pending"
	r.DurableStart = true
	startRecordErr := writeRecord(req.PrivateDir, "started-record.json", r)
	if startRecordErr != nil {
		r.DurableStart = false
		cancel()
	}
	e = cmd.Wait()
	r.GroupCleanup = cleanupProcess(cmd)
	r.WallMillis = time.Since(start).Milliseconds()
	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		r.ExitCode = &code
	}
	switch {
	case stdout.exceeded || stderr.exceeded:
		r.ProcessStatus = "capture_limit"
	case childCtx.Err() != nil:
		r.ProcessStatus = "timeout_or_canceled"
	case errors.Is(e, exec.ErrWaitDelay):
		r.ProcessStatus = "inherited_pipe_timeout"
	case e != nil:
		r.ProcessStatus = "nonzero_exit"
	default:
		r.ProcessStatus = "exited_zero"
	}
	r.Stdout = stdout.report()
	r.Stderr = stderr.report()
	if stdout.file.Sync() != nil || stderr.file.Sync() != nil {
		r.Stdout.Complete = false
		r.Stderr.Complete = false
		r.Stdout.Status = "capture_sync_failed"
		r.Stderr.Status = "capture_sync_failed"
	}
	if errors.Is(e, exec.ErrWaitDelay) {
		r.Stdout.Complete = false
		r.Stderr.Complete = false
		r.Stdout.Status = "pipe_completion_unknown"
		r.Stderr.Status = "pipe_completion_unknown"
	}
	if stdout.writeFailed || stderr.writeFailed {
		r.ProcessStatus = "capture_write_failed"
	}
	if startRecordErr != nil {
		r.ProcessStatus = "start_record_write_failed"
	}
	if r.Stdout.Complete {
		f, err := os.Open(filepath.Join(req.PrivateDir, "stdout.jsonl"))
		if err == nil {
			summary, err := taskoutcome.Summarize(f, taskoutcome.Metadata{PublicTaskLabel: req.TaskID, RequestedModel: req.Model, RequestedReasoning: req.Reasoning})
			f.Close()
			if err == nil {
				r.Summary = &summary
				r.SummaryStatus = "parsed"
			} else {
				r.SummaryStatus = "parse_failed"
			}
		} else {
			r.SummaryStatus = "read_failed"
		}
	} else {
		r.SummaryStatus = "capture_incomplete"
	}
	r.WholeAttemptUsageComplete = r.ProcessStatus == "exited_zero" && r.GroupCleanup == "group_terminated" && r.Stdout.Complete && r.Stderr.Complete && r.Summary != nil && r.Summary.UsageComplete
	if r.GroupCleanup != "group_terminated" {
		r.CandidateStatus = "cleanup_unknown"
	} else {
		files, err := snapshot(workspace, candidate, base)
		if err != nil {
			r.CandidateStatus = "capture_failed"
		} else {
			r.CandidateFiles = hashes(files)
			r.CandidateSHA256 = digestJSON(r.CandidateFiles)
			r.CandidateFileCount = len(files)
			r.CandidateStatus = "captured_task_closure"
			verifyStart := time.Now()
			report, err := taskverify.Verify(ctx, taskverify.Request{TaskID: req.TaskID, BaseRevision: spec.BaseRevision, BaseFiles: base, AttributionFiles: all[len(base):], CandidateDir: candidate, Timeout: taskverify.DefaultTimeout, GoRoot: req.GoRoot})
			verifyMillis := time.Since(verifyStart).Milliseconds()
			r.VerificationWallMillis = &verifyMillis
			if err == nil {
				r.Verification = report
				r.VerificationStatus = report.Status
			} else {
				r.VerificationStatus = "verification_failed"
			}
		}
	}
	// A private immutable output copy is convenient for audits. Failure cannot
	// erase the started-attempt record returned to the caller.
	if e = os.Remove(filepath.Join(home, "auth.json")); e == nil || errors.Is(e, os.ErrNotExist) {
		r.AuthCleanup = "removed_or_not_present"
	} else {
		r.AuthCleanup = "cleanup_failed"
	}
	if writeRecord(req.PrivateDir, "record.json", r) != nil {
		r.VerificationStatus = "record_write_failed"
	}
	return r, nil
}

func writeRecord(dir, name string, r Record) error {
	b, e := json.Marshal(r)
	if e != nil {
		return Error("record_write_failed")
	}
	f, e := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return Error("record_write_failed")
	}
	_, e = f.Write(append(b, '\n'))
	if e == nil {
		e = f.Sync()
	}
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		return Error("record_write_failed")
	}
	return nil
}

func isolatedPrivatePath(path string) (string, error) {
	parent, e := filepath.EvalSymlinks(filepath.Dir(path))
	if e != nil {
		return "", Error("private_parent_unavailable")
	}
	allowed := false
	for _, candidate := range []string{os.TempDir(), "/tmp", "/private/tmp"} {
		temp, e := filepath.EvalSymlinks(candidate)
		if e != nil {
			continue
		}
		rel, e := filepath.Rel(temp, parent)
		if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", Error("private_directory_requires_os_temp")
	}
	for ancestor := parent; ; ancestor = filepath.Dir(ancestor) {
		for _, marker := range []string{".git", "AGENTS.md", ".codex", ".agents"} {
			if _, e := os.Lstat(filepath.Join(ancestor, marker)); e == nil {
				return "", Error("private_project_ancestor_forbidden")
			} else if !errors.Is(e, os.ErrNotExist) {
				return "", Error("private_parent_unavailable")
			}
		}
		if ancestor == filepath.Dir(ancestor) {
			break
		}
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

func snapshot(source, target string, base []taskverify.BaseFile) ([]taskverify.BaseFile, error) {
	root, e := os.OpenRoot(source)
	if e != nil {
		return nil, Error("candidate_capture_failed")
	}
	defer root.Close()
	files := []taskverify.BaseFile{}
	total := 0
	for _, f := range base {
		b, e := readRootFile(root, f.Path, taskverify.MaxFileBytes)
		if e != nil {
			return nil, Error("candidate_capture_failed")
		}
		total += len(b)
		if total > taskverify.MaxTotalBytes || len(files) >= taskverify.MaxFiles {
			return nil, Error("candidate_byte_limit")
		}
		files = append(files, taskverify.BaseFile{Path: f.Path, SHA256: digest(b), Data: b})
	}
	if e = writeFiles(target, files); e != nil {
		return nil, e
	}
	return files, nil
}

type boundedBuffer struct {
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func (w *boundedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	if w.buf.Len()+n > w.limit {
		w.exceeded = true
		left := max(0, w.limit-w.buf.Len())
		w.buf.Write(p[:left])
		return n, nil
	}
	return w.buf.Write(p)
}

// HostSupport is static and launches no model or CLI process.
func HostSupport() string {
	if runtime.GOOS != "darwin" || !executionSupported() {
		return "unavailable"
	}
	return "darwin_custom_profile_required"
}
