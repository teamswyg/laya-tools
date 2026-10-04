//go:build darwin && arm64

// SPDX-License-Identifier: Apache-2.0
// Unexecuted authoring draft. No JWT or model import, loop, retry or profiler.
package main

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
	"syscall"
	"time"
	"unicode/utf8"
)

const (
	configLimit  = 8 << 10
	targetLimit  = 32 << 20
	captureLimit = 280 << 10
	gateLimit    = 4 << 10
	rssLimit     = 256 << 20
	childTimeout = 30 * time.Second
	pipeWait     = time.Second
)

type config struct {
	Schema      string `json:"schema"`
	Profile     string `json:"profile"`
	TargetPath  string `json:"target_path"`
	TargetSHA   string `json:"target_sha256"`
	InputPath   string `json:"input_path"`
	InputSHA    string `json:"input_sha256"`
	CapturePath string `json:"capture_path"`
}

type profile struct {
	Name        string `json:"name"`
	InputLimit  int    `json:"input_file_limit_bytes"`
	StdinLimit  int    `json:"stdin_limit_bytes"`
	StdoutLimit int    `json:"stdout_limit_bytes"`
	StderrLimit int    `json:"stderr_limit_bytes"`
}

var profiles = [2]profile{
	{Name: "signer", InputLimit: 32 << 10, StdinLimit: 0, StdoutLimit: 32 << 10, StderrLimit: 1 << 10},
	{Name: "observer", InputLimit: 64 << 10, StdinLimit: 64 << 10, StdoutLimit: 192 << 10, StderrLimit: 4 << 10},
}

var childEnv = [5]string{
	"GOMAXPROCS=1", "GOMEMLIMIT=64MiB", "GOGC=100", "GOTRACEBACK=none", "GODEBUG=",
}

type streamRecord struct {
	Raw                   []byte `json:"raw_base64"`
	RetainedBytes         int    `json:"retained_bytes"`
	RetainedSHA           string `json:"retained_sha256"`
	WriteObservedBytes    int64  `json:"write_observed_bytes"`
	LimitBytes            int    `json:"limit_bytes"`
	Overflow              bool   `json:"overflow"`
	FullResponsePreserved bool   `json:"full_response_preserved"`
	Scope                 string `json:"scope"`
}

type resource struct {
	CPUAvailable      bool            `json:"cpu_available"`
	UserNS            *int64          `json:"user_cpu_ns"`
	SystemNS          *int64          `json:"system_cpu_ns"`
	RusageAvailable   bool            `json:"syscall_rusage_available"`
	Rusage            *syscall.Rusage `json:"syscall_rusage"`
	DarwinMaxRSSBytes *int64          `json:"darwin_maxrss_bytes"`
	Scope             string          `json:"scope"`
}

type intent struct {
	Schema              string    `json:"schema"`
	Config              config    `json:"private_config"`
	ConfigSHA           string    `json:"actual_config_sha256"`
	TargetBytes         int64     `json:"actual_target_bytes"`
	InputBytes          int       `json:"actual_input_bytes"`
	Profile             profile   `json:"fixed_profile"`
	Environment         [5]string `json:"fixed_child_environment"`
	ExtraChildArguments int       `json:"extra_child_arguments"`
	TimeoutNS           int64     `json:"timeout_ns"`
	PipeWaitNS          int64     `json:"pipe_cleanup_wait_ns"`
	PostRunRSSLimit     int64     `json:"post_run_darwin_maxrss_limit_bytes"`
	Go                  string    `json:"controller_go"`
	State               string    `json:"state"`
	Scope               string    `json:"scope"`
}

type capture struct {
	Schema                 string       `json:"schema"`
	IntentSHA              string       `json:"intent_sha256"`
	Profile                string       `json:"profile"`
	TargetSHA              string       `json:"actual_target_sha256"`
	InputSHA               string       `json:"actual_input_sha256"`
	InputBytes             int          `json:"actual_input_bytes"`
	StdinBytes             int          `json:"stdin_buffer_bytes"`
	StdinWrittenBytes      int          `json:"stdin_written_bytes"`
	StdinComplete          bool         `json:"stdin_write_and_close_complete"`
	StdinWriterJoined      bool         `json:"stdin_writer_joined"`
	StdinCode              string       `json:"stdin_code"`
	StartAttempts          int          `json:"explicit_start_attempts"`
	Started                bool         `json:"started"`
	StartCode              string       `json:"start_code"`
	WaitCalls              int          `json:"wait_calls"`
	WaitReturned           bool         `json:"wait_returned"`
	WaitCode               string       `json:"wait_code"`
	WatcherJoined          bool         `json:"watcher_joined"`
	DeadlineObserved       bool         `json:"deadline_observed"`
	OverflowCancelObserved bool         `json:"overflow_cancel_observed"`
	StartUTC               string       `json:"start_attempt_utc"`
	WaitReturnUTC          string       `json:"wait_return_utc"`
	EndUTC                 string       `json:"lifecycle_joined_utc"`
	WallNS                 int64        `json:"start_to_lifecycle_join_wall_ns"`
	ExitAvailable          bool         `json:"exit_available"`
	ExitCode               *int         `json:"exit_code"`
	Signal                 *int         `json:"signal"`
	Stdout                 streamRecord `json:"stdout"`
	Stderr                 streamRecord `json:"stderr"`
	Resource               resource     `json:"resource"`
	Interpretation         string       `json:"interpretation"`
	Scope                  string       `json:"scope"`
}

type gate struct {
	Schema         string `json:"schema"`
	CaptureSHA     string `json:"complete_capture_sha256"`
	CaptureBytes   int    `json:"complete_capture_bytes"`
	State          string `json:"state"`
	Reason         string `json:"reason"`
	JSONChecked    bool   `json:"stdout_json_syntax_checked_after_capture"`
	JSONObject     bool   `json:"stdout_one_json_object"`
	RSSChecked     bool   `json:"darwin_maxrss_checked"`
	RSSWithinLimit bool   `json:"darwin_maxrss_within_256MiB"`
	Scope          string `json:"scope"`
}

// Each stream writer has exactly one os/exec copy-goroutine owner. Main reads
// it only after Wait has joined copying; no shared writer lock or map exists.
type boundedWriter struct {
	Buffer   []byte
	Observed int64
	Limit    int
	Overflow bool
	Alert    chan<- struct{}
}

type stdinReceipt struct {
	Bytes    int
	Complete bool
	Code     string
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	w.Observed += int64(len(p))
	keep := len(p)
	if keep > w.Limit-len(w.Buffer) {
		keep = w.Limit - len(w.Buffer)
	}
	w.Buffer = append(w.Buffer, p[:keep]...)
	if keep != len(p) {
		w.Overflow = true
		select {
		case w.Alert <- struct{}{}:
		default:
		}
		return keep, io.ErrShortBuffer
	}
	return len(p), nil
}

func (w *boundedWriter) record(waitClean bool) streamRecord {
	return streamRecord{
		Raw: w.Buffer, RetainedBytes: len(w.Buffer), RetainedSHA: digest(w.Buffer),
		WriteObservedBytes: w.Observed, LimitBytes: w.Limit, Overflow: w.Overflow,
		FullResponsePreserved: waitClean && !w.Overflow,
		Scope:                 "Private exact retained prefix. Beyond-cap bytes are not retained; overflow is unknown and never a complete response. Observed byte count is only bytes delivered to writer, not all child output.",
	}
}

func main() {
	defer func() {
		if recover() != nil {
			publicFailure("runner_panic")
			os.Exit(1)
		}
	}()
	if len(os.Args) != 2 {
		publicFailure("runner_arguments_invalid")
		os.Exit(1)
	}
	code := run(os.Args[1])
	if code != "" {
		publicFailure(code)
		os.Exit(1)
	}
	// Child raw bytes, paths, keys and environment never reach public output.
	_, _ = os.Stdout.Write([]byte("{\"schema\":\"riido-one-shot-private-capture-status-v1\",\"code\":\"first_capture_and_gate_persisted\"}\n"))
}

func run(configPath string) string {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" || runtime.Version() != "go1.27.1" {
		return "controller_toolchain_mismatch"
	}
	cfgBytes, code := readOwned(configPath, configLimit)
	if code != "" {
		return "config_read_failed"
	}
	cfg, code := decodeConfig(cfgBytes)
	if code != "" {
		return code
	}
	var selected profile
	for _, p := range profiles {
		if p.Name == cfg.Profile {
			selected = p
		}
	}
	if selected.Name == "" {
		return "profile_invalid"
	}
	input, code := readOwned(cfg.InputPath, selected.InputLimit)
	if code != "" || len(input) == 0 || digest(input) != cfg.InputSHA {
		return "input_pin_or_size_failed"
	}
	// For signer this is the immutable embedded-source snapshot verification,
	// not stdin. Root must independently bind the exact embed bytes to binary.
	target, code := openOwned(cfg.TargetPath, targetLimit)
	if code != "" {
		return "target_open_failed"
	}
	defer target.Close()
	targetStat, err := target.Stat()
	if err != nil || targetStat.Size() <= 0 || targetStat.Mode().Perm()&0o111 == 0 {
		return "target_shape_failed"
	}
	var hashBuffer [32 << 10]byte
	hasher := sha256.New()
	n, err := io.CopyBuffer(hasher, io.LimitReader(target, targetLimit+1), hashBuffer[:])
	if err != nil || n != targetStat.Size() || n > targetLimit || hex.EncodeToString(hasher.Sum(nil)) != cfg.TargetSHA {
		return "target_pin_or_size_failed"
	}
	afterHash, err := target.Stat()
	if err != nil || !sameStat(targetStat, afterHash) {
		return "target_changed_during_hash"
	}
	if !safePath(cfg.CapturePath) {
		return "capture_path_invalid"
	}
	parent, err := os.Lstat(filepath.Dir(cfg.CapturePath))
	if err != nil || !parent.IsDir() || parent.Mode().Perm() != 0o700 || !owned(parent) {
		return "capture_parent_must_be_owned_0700"
	}
	// Reserve all three first-attempt names before Start. Never overwrite or
	// remove a partial/failed artifact; an existing name prevents another run.
	intentFile, err := os.OpenFile(cfg.CapturePath+".intent.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "intent_exclusive_create_failed"
	}
	defer intentFile.Close()
	captureFile, err := os.OpenFile(cfg.CapturePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "capture_exclusive_create_failed"
	}
	defer captureFile.Close()
	gateFile, err := os.OpenFile(cfg.CapturePath+".gate.json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "gate_exclusive_create_failed"
	}
	defer gateFile.Close()
	in := intent{
		Schema: "riido-one-shot-private-process-intent-v1", Config: cfg, ConfigSHA: digest(cfgBytes),
		TargetBytes: n, InputBytes: len(input), Profile: selected, Environment: childEnv,
		ExtraChildArguments: 0, TimeoutNS: int64(childTimeout), PipeWaitNS: int64(pipeWait),
		PostRunRSSLimit: rssLimit, Go: runtime.Version(), State: "PREPARED_BEFORE_FIRST_START_ATTEMPT",
		Scope: "Private immutable one-attempt intent. No retries, sampler, profiler, shell or credentials. Root must reserve worst-case files/binary/input/controller and freeze source, compiler, config, binary and embed provenance before invoking.",
	}
	intentBytes, code := persistJSON(intentFile, in, gateLimit)
	if code != "" {
		return "intent_persist_failed"
	}
	if syncParent(cfg.CapturePath) != nil {
		return "intent_parent_sync_failed"
	}
	// Path-based exec has an unavoidable final stat-to-exec race. Hold the
	// verified fd, require stable path stat, and rely on Root's frozen ownership.
	beforeStart, err := os.Lstat(cfg.TargetPath)
	if err != nil || !sameStat(targetStat, beforeStart) {
		return "target_changed_before_start"
	}
	ctx, cancel := context.WithTimeout(context.Background(), childTimeout)
	defer cancel()
	alert := make(chan struct{}, 2)
	stdout := boundedWriter{Buffer: make([]byte, 0, selected.StdoutLimit), Limit: selected.StdoutLimit, Alert: alert}
	stderr := boundedWriter{Buffer: make([]byte, 0, selected.StderrLimit), Limit: selected.StderrLimit, Alert: alert}
	cmd := exec.CommandContext(ctx, cfg.TargetPath)
	cmd.Env = childEnv[:]
	cmd.Dir = filepath.Dir(cfg.TargetPath)
	cmd.WaitDelay = pipeWait
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	stdinBytes := 0
	var childStdin io.WriteCloser
	if selected.StdinLimit > 0 {
		childStdin, err = cmd.StdinPipe()
		if err != nil {
			return "stdin_pipe_preparation_failed"
		}
		defer childStdin.Close()
		stdinBytes = len(input)
	}
	// nil stdin is /dev/null; signer receives zero stdin bytes and no arguments.
	rec := capture{
		Schema: "riido-one-shot-private-first-capture-v2", IntentSHA: digest(intentBytes),
		Profile: selected.Name, TargetSHA: cfg.TargetSHA, InputSHA: digest(input), InputBytes: len(input),
		StdinBytes: stdinBytes, Interpretation: "UNINTERPRETED_RAW_FIRST_RESPONSE",
		Scope: "Exactly one child Start attempt. Wait return UTC is observed immediately after cmd.Wait; lifecycle joined UTC/wall follow stdin/watcher joins and wait classification. Start wall begins immediately before Start attempt after input/config/target hashing, intent persistence, context/pipe/buffer preparation; resource extraction and capture/gate persistence are outside this wall. Startup/init/work/pipe/wait are included but unattributed. No inference, labels, roles or performance conclusion; response syntax and resource gate follow durable capture.",
	}
	start := time.Now()
	rec.StartUTC = start.UTC().Format(time.RFC3339Nano)
	rec.StartAttempts = 1
	err = cmd.Start()
	if err != nil {
		rec.StartCode, rec.WaitCode = "child_start_failed", "not_waited_not_started"
		rec.StdinCode = "child_not_started"
	} else {
		rec.Started = true
		stdinDone := make(chan stdinReceipt, 1)
		if childStdin != nil {
			go func() { stdinDone <- supplyStdin(childStdin, input) }()
		} else {
			stdinDone <- stdinReceipt{Complete: true, Code: "no_stdin_dev_null"}
		}
		done, watcher := make(chan struct{}), make(chan [2]bool, 1)
		go func() {
			var result [2]bool
			select {
			case <-ctx.Done():
				result[0] = errors.Is(ctx.Err(), context.DeadlineExceeded)
			case <-alert:
				result[1] = true
				cancel()
			case <-done:
			}
			watcher <- result
		}()
		rec.WaitCalls = 1
		err = cmd.Wait()
		rec.WaitReturnUTC = time.Now().UTC().Format(time.RFC3339Nano)
		rec.WaitReturned = true
		stdinResult := <-stdinDone
		rec.StdinWrittenBytes, rec.StdinComplete, rec.StdinCode = stdinResult.Bytes, stdinResult.Complete, stdinResult.Code
		rec.StdinWriterJoined = true
		close(done)
		seen := <-watcher
		rec.WatcherJoined = true
		rec.DeadlineObserved = seen[0] || errors.Is(ctx.Err(), context.DeadlineExceeded)
		rec.OverflowCancelObserved = seen[1]
		switch {
		case err == nil:
			rec.WaitCode = "wait_clean"
		case errors.Is(err, exec.ErrWaitDelay):
			rec.WaitCode = "pipe_cleanup_wait_exceeded"
		default:
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				rec.WaitCode = "child_exit_error"
			} else {
				rec.WaitCode = "wait_or_stream_io_error"
			}
		}
	}
	end := time.Now()
	rec.EndUTC, rec.WallNS = end.UTC().Format(time.RFC3339Nano), end.Sub(start).Nanoseconds()
	if cmd.ProcessState != nil {
		rec.ExitAvailable = true
		exitCode := cmd.ProcessState.ExitCode()
		rec.ExitCode = &exitCode
		if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			signal := int(status.Signal())
			rec.Signal = &signal
		}
		rec.Resource = resourceForState(cmd.ProcessState)
	} else {
		rec.Resource.Scope = "Unavailable because no reaped child ProcessState exists."
	}
	clean := rec.WaitReturned && rec.WaitCode == "wait_clean"
	rec.Stdout, rec.Stderr = stdout.record(clean), stderr.record(clean)
	// No stdout syntax or target result interpretation has occurred yet.
	packed, code := persistJSON(captureFile, rec, captureLimit)
	if code != "" {
		return "first_capture_persist_failed"
	}
	if syncParent(cfg.CapturePath) != nil {
		return "capture_parent_sync_failed"
	}
	// Only now classify lifecycle/resource and generic syntax. No schema,
	// policy acceptance, candidate truth, label or model admission is inferred.
	g := gate{
		Schema: "riido-one-shot-private-post-capture-gate-v1", CaptureSHA: digest(packed), CaptureBytes: len(packed),
		State: "unknown", Reason: "process_lifecycle_incomplete",
		Scope: "A passing gate means only reaped clean process, bounded complete stream, one JSON object and observed Darwin lifetime Maxrss<=256MiB. Output policy/candidate semantics remain unreviewed. GOMEMLIMIT is soft; no hard RSS enforcement, sampler, CPU profile, GPU or savings claim.",
	}
	if rec.Resource.DarwinMaxRSSBytes != nil {
		g.RSSChecked = true
		g.RSSWithinLimit = *rec.Resource.DarwinMaxRSSBytes <= rssLimit
	}
	if clean && rec.Started && rec.StdinComplete && rec.StdinWriterJoined && rec.ExitCode != nil && *rec.ExitCode == 0 && !rec.DeadlineObserved && !stdout.Overflow && !stderr.Overflow && rec.Stdout.FullResponsePreserved && rec.Stderr.FullResponsePreserved {
		switch {
		case !rec.Resource.CPUAvailable || !g.RSSChecked:
			g.Reason = "resource_usage_unavailable"
		case !g.RSSWithinLimit:
			g.Reason = "darwin_peak_rss_over_limit"
		default:
			g.JSONChecked = true
			trim := bytes.TrimSpace(rec.Stdout.Raw)
			g.JSONObject = len(trim) >= 2 && trim[0] == '{' && trim[len(trim)-1] == '}' && utf8.Valid(trim) && json.Valid(trim)
			if g.JSONObject {
				g.State, g.Reason = "captured_process_completed_json_object_only", "semantic_review_pending"
			} else {
				g.Reason = "stdout_json_malformed_or_not_object"
			}
		}
	} else if stdout.Overflow || stderr.Overflow {
		g.Reason = "output_overflow_first_prefix_preserved"
	} else if rec.DeadlineObserved {
		g.Reason = "child_timeout_first_bytes_preserved"
	}
	if _, code = persistJSON(gateFile, g, gateLimit); code != "" {
		return "post_capture_gate_persist_failed"
	}
	if syncParent(cfg.CapturePath) != nil {
		return "gate_parent_sync_failed"
	}
	return ""
}

// Derived from repository riido-residentperf/process.go resourceForState and
// resourceRSSForOS Darwin branch; resource units remain Darwin bytes. This
// owned draft stores integer CPU durations/null availability and no raw errors.
func resourceForState(state *os.ProcessState) resource {
	r := resource{Scope: "Exact OS-reported direct-child lifetime counters including init and work; Darwin Maxrss unit bytes, not Go heap or sampled RSS. No GPU/resource attribution or live hard limit."}
	if state == nil {
		return r
	}
	u, s := state.UserTime(), state.SystemTime()
	if u >= 0 && s >= 0 {
		un, sn := u.Nanoseconds(), s.Nanoseconds()
		r.UserNS, r.SystemNS, r.CPUAvailable = &un, &sn, true
	}
	if usage, ok := state.SysUsage().(*syscall.Rusage); ok {
		r.RusageAvailable = true
		r.Rusage = usage
		if usage.Maxrss > 0 {
			value := usage.Maxrss
			r.DarwinMaxRSSBytes = &value
		}
	}
	return r
}

// One goroutine owns the pipe and exact acknowledged write count. These are
// kernel pipe writes, not proof the child consumed or interpreted all bytes.
func supplyStdin(pipe io.WriteCloser, input []byte) stdinReceipt {
	r := stdinReceipt{}
	for r.Bytes < len(input) {
		n, err := pipe.Write(input[r.Bytes:])
		if n > 0 {
			r.Bytes += n
		}
		if err != nil || n == 0 {
			r.Code = "stdin_write_failed"
			_ = pipe.Close()
			return r
		}
	}
	if pipe.Close() != nil {
		r.Code = "stdin_close_failed"
		return r
	}
	r.Complete, r.Code = true, "stdin_full_write_and_close"
	return r
}

func decodeConfig(raw []byte) (config, string) {
	var cfg config
	if !utf8.Valid(raw) {
		return cfg, "config_utf8_invalid"
	}
	// Flat strict object, all seven string keys exactly once, no maps. A
	// duplicate expected hash/path is rejected instead of last-value wins.
	keys := [7]string{"schema", "profile", "target_path", "target_sha256", "input_path", "input_sha256", "capture_path"}
	var seen [7]bool
	d := json.NewDecoder(bytes.NewReader(raw))
	first, err := d.Token()
	if err != nil || first != json.Delim('{') {
		return cfg, "config_object_invalid"
	}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return cfg, "config_key_invalid"
		}
		name, ok := key.(string)
		if !ok {
			return cfg, "config_key_invalid"
		}
		index := -1
		for i, k := range keys {
			if k == name {
				index = i
				break
			}
		}
		if index < 0 || seen[index] {
			return cfg, "config_unknown_or_duplicate_key"
		}
		seen[index] = true
		var value string
		if d.Decode(&value) != nil || len(value) > 1024 {
			return cfg, "config_value_invalid"
		}
		switch index {
		case 0:
			cfg.Schema = value
		case 1:
			cfg.Profile = value
		case 2:
			cfg.TargetPath = value
		case 3:
			cfg.TargetSHA = value
		case 4:
			cfg.InputPath = value
		case 5:
			cfg.InputSHA = value
		case 6:
			cfg.CapturePath = value
		}
	}
	last, err := d.Token()
	if err != nil || last != json.Delim('}') {
		return cfg, "config_object_end_invalid"
	}
	var extra json.RawMessage
	if d.Decode(&extra) != io.EOF {
		return cfg, "config_trailing_data"
	}
	for _, present := range seen {
		if !present {
			return cfg, "config_missing_key"
		}
	}
	if cfg.Schema != "riido-one-shot-private-process-config-v1" || !validSHA(cfg.TargetSHA) || !validSHA(cfg.InputSHA) || !safePath(cfg.TargetPath) || !safePath(cfg.InputPath) || !safePath(cfg.CapturePath) {
		return cfg, "config_contract_invalid"
	}
	return cfg, ""
}

func safePath(p string) bool {
	if len(p) == 0 || len(p) > 1024 || !utf8.ValidString(p) || !filepath.IsAbs(p) || filepath.Clean(p) != p {
		return false
	}
	for _, r := range p {
		if r < 32 || r == 127 {
			return false
		}
	}
	// Reject symlink ancestors as well as leaf symlinks. Paths are frozen
	// private configuration; no shell/string command construction occurs.
	dir := filepath.Dir(p)
	for {
		st, err := os.Lstat(dir)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return false
		}
		if dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return true
}

func owned(st os.FileInfo) bool {
	u, ok := st.Sys().(*syscall.Stat_t)
	return ok && int(u.Uid) == os.Getuid()
}

func openOwned(path string, limit int64) (*os.File, string) {
	if !safePath(path) {
		return nil, "private_path_invalid"
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, "private_open_failed"
	}
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || !owned(st) || st.Size() < 0 || st.Size() > limit || st.Mode().Perm()&0o022 != 0 {
		f.Close()
		return nil, "private_file_shape_failed"
	}
	return f, ""
}

func readOwned(path string, limit int) ([]byte, string) {
	f, code := openOwned(path, int64(limit))
	if code != "" {
		return nil, code
	}
	defer f.Close()
	before, err := f.Stat()
	if err != nil {
		return nil, "private_stat_failed"
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil || len(b) > limit || int64(len(b)) != before.Size() {
		return nil, "private_read_or_size_failed"
	}
	after, err := f.Stat()
	if err != nil || !sameStat(before, after) {
		return nil, "private_file_changed"
	}
	return b, ""
}

func sameStat(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Size() == b.Size() && a.Mode() == b.Mode() && a.ModTime().Equal(b.ModTime()) && owned(a) && owned(b)
}

func validSHA(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func persistJSON(f *os.File, value any, limit int) ([]byte, string) {
	packed, err := json.Marshal(value)
	if err != nil || len(packed) >= limit {
		return nil, "private_record_encoding_or_size_failed"
	}
	packed = append(packed, '\n')
	n, err := f.Write(packed)
	if err != nil || n != len(packed) || f.Sync() != nil {
		return nil, "private_record_write_or_sync_failed"
	}
	return packed, ""
}

func syncParent(p string) error {
	f, err := os.Open(filepath.Dir(p))
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func publicFailure(code string) {
	// Every caller supplies a fixed authored code, no error/panic/path/env.
	packed, err := json.Marshal(struct {
		Schema string `json:"schema"`
		Code   string `json:"code"`
	}{"riido-one-shot-private-capture-failure-v1", code})
	if err == nil && len(packed) < 256 {
		_, _ = os.Stderr.Write(append(packed, '\n'))
	}
}
