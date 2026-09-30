// Package taskoutcome summarizes private Codex JSONL without retaining item
// content. A completed turn is neither task acceptance nor model identity proof.
package taskoutcome

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	MaxTraceBytes = 64 << 20
	MaxLineBytes  = 1 << 20
	MaxEvents     = 100000
	MaxTurns      = 4096
)

// ErrorCode deliberately contains no source lines, paths, or reader errors.
type ErrorCode string

func (e ErrorCode) Error() string { return string(e) }

const (
	ErrMetadata  ErrorCode = "invalid_metadata"
	ErrRead      ErrorCode = "trace_read_failed"
	ErrTraceSize ErrorCode = "trace_byte_limit"
	ErrLineSize  ErrorCode = "line_byte_limit"
	ErrEvents    ErrorCode = "event_limit"
	ErrTurns     ErrorCode = "turn_limit"
	ErrJSON      ErrorCode = "invalid_event_json"
	ErrDuplicate ErrorCode = "duplicate_control_field"
	ErrUsage     ErrorCode = "invalid_usage"
	ErrOverflow  ErrorCode = "usage_total_overflow"
	ErrLifecycle ErrorCode = "invalid_lifecycle"
)

// Metadata describes the caller's requested profile only. It is not observed
// execution evidence. Labels are bounded identifiers, never free-form prompts.
type Metadata struct {
	PublicTaskLabel    string `json:"public_task_label,omitempty"`
	RequestedModel     string `json:"requested_model,omitempty"`
	RequestedReasoning string `json:"requested_reasoning,omitempty"`
}

// FieldUsage preserves partial observations. Total is null unless every turn
// has a terminal completed event and this field is present in every such turn.
// Optional reasoning may be unknown while the three core fields are complete.
type FieldUsage struct {
	ObservedTotal *int64 `json:"observed_total"`
	ObservedTurns int    `json:"observed_turns"`
	MissingTurns  int    `json:"missing_completed_turns"`
	Total         *int64 `json:"total"`
	Complete      bool   `json:"complete"`
}

type Usage struct {
	Input     FieldUsage `json:"input_tokens"`
	Cached    FieldUsage `json:"cached_input_tokens"`
	Output    FieldUsage `json:"output_tokens"`
	Reasoning FieldUsage `json:"reasoning_output_tokens"`
}

type Summary struct {
	Schema               string   `json:"schema"`
	Metadata             Metadata `json:"requested_metadata"`
	MetadataEvidence     string   `json:"metadata_evidence"`
	TraceProvenance      string   `json:"trace_provenance"`
	ObservedModel        string   `json:"observed_model"`
	ProcessExit          string   `json:"process_exit"`
	ExecutionTime        string   `json:"execution_time"`
	TaskAcceptance       string   `json:"task_acceptance"`
	TraceSHA256          string   `json:"trace_bytes_sha256"`
	TraceBytes           int64    `json:"trace_bytes"`
	Events               int      `json:"events"`
	ThreadsStarted       int      `json:"threads_started"`
	TurnsStarted         int      `json:"turns_started"`
	TurnsCompleted       int      `json:"turns_completed"`
	TurnsFailed          int      `json:"turns_failed"`
	ErrorEvents          int      `json:"error_events"`
	StartupErrorItems    int      `json:"startup_error_items"`
	DiscardedItemEvents  int      `json:"discarded_item_events"`
	UnknownControlEvents int      `json:"unknown_control_events"`
	LifecycleComplete    bool     `json:"lifecycle_complete"`
	TraceStatus          string   `json:"trace_status"`
	UsageComplete        bool     `json:"usage_complete"`
	Usage                Usage    `json:"usage"`
}

func validLabel(s string) bool {
	if len(s) > 96 {
		return false
	}
	if s == "" {
		return true
	}
	for i := range len(s) {
		b := s[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || (i > 0 && strings.ContainsRune("_.:-", rune(b)))) {
			return false
		}
	}
	lower := strings.ToLower(s)
	for _, prefix := range []string{"ghp_", "github_pat_", "hf_", "sk-", "akia"} {
		if strings.Contains(lower, prefix) {
			return false
		}
	}
	return true
}

type limits struct{ trace, line, events, turns int }
type event struct {
	kind  string
	usage [4]*int64
	item  json.RawMessage
}

// Summarize consumes the trace once. Limits include all trace bytes, including
// blank lines and newlines; the line limit excludes its final newline.
// Sequential thread.started blocks are allowed, but turns cannot overlap.
func Summarize(in io.Reader, metadata Metadata) (Summary, error) {
	return summarize(in, metadata, limits{MaxTraceBytes, MaxLineBytes, MaxEvents, MaxTurns})
}

func summarize(in io.Reader, m Metadata, bound limits) (Summary, error) {
	if in == nil || !validLabel(m.PublicTaskLabel) || !validLabel(m.RequestedModel) || !validLabel(m.RequestedReasoning) {
		return Summary{}, ErrMetadata
	}
	s := Summary{Schema: "riido-codex-jsonl-summary-v1", Metadata: m, MetadataEvidence: "caller_requested_unverified", TraceProvenance: "unknown", ObservedModel: "unknown", ProcessExit: "unknown", ExecutionTime: "unknown", TaskAcceptance: "unknown", TraceStatus: "incomplete"}
	h := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(io.LimitReader(in, int64(bound.trace)+1), h), 16<<10)
	var line []byte
	thread, threadHasTurn, active := false, false, false
	for {
		part, err := r.ReadSlice('\n')
		s.TraceBytes += int64(len(part))
		if s.TraceBytes > int64(bound.trace) {
			return Summary{}, ErrTraceSize
		}
		line = append(line, part...)
		body := line
		if len(body) > 0 && body[len(body)-1] == '\n' {
			body = body[:len(body)-1]
		}
		if len(body) > bound.line {
			return Summary{}, ErrLineSize
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil && err != io.EOF {
			return Summary{}, ErrRead
		}
		if len(bytes.TrimSpace(body)) != 0 {
			if s.Events == bound.events {
				return Summary{}, ErrEvents
			}
			e, e2 := parseEvent(body)
			if e2 != nil {
				return Summary{}, e2
			}
			s.Events++
			switch e.kind {
			case "thread.started":
				if active || (thread && !threadHasTurn) {
					return Summary{}, ErrLifecycle
				}
				thread, threadHasTurn = true, false
				s.ThreadsStarted++
			case "turn.started":
				if !thread || active {
					return Summary{}, ErrLifecycle
				}
				if s.TurnsStarted == bound.turns {
					return Summary{}, ErrTurns
				}
				active, threadHasTurn = true, true
				s.TurnsStarted++
			case "turn.completed":
				if !active {
					return Summary{}, ErrLifecycle
				}
				active = false
				s.TurnsCompleted++
				fields := []*FieldUsage{&s.Usage.Input, &s.Usage.Cached, &s.Usage.Output, &s.Usage.Reasoning}
				for i, value := range e.usage {
					if e2 := add(fields[i], value); e2 != nil {
						return Summary{}, e2
					}
				}
			case "turn.failed":
				if !active {
					return Summary{}, ErrLifecycle
				}
				active = false
				s.TurnsFailed++
			case "error":
				s.ErrorEvents++
			default:
				if strings.HasPrefix(e.kind, "item.") {
					if !active {
						// Codex0.158 can emit a completed error diagnostic after
						// thread creation and before its first turn. This narrow
						// startup exception never admits orphan/late/other items.
						if e.kind != "item.completed" || !thread || threadHasTurn {
							return Summary{}, ErrLifecycle
						}
						startupError, err := startupErrorItem(e.item)
						if err != nil {
							return Summary{}, err
						}
						if !startupError {
							return Summary{}, ErrLifecycle
						}
						s.StartupErrorItems++
					} else {
						s.DiscardedItemEvents++
					}
				} else {
					s.UnknownControlEvents++
				}
			}
		}
		line = line[:0]
		if err == io.EOF {
			break
		}
	}
	s.TraceSHA256 = hex.EncodeToString(h.Sum(nil))
	s.LifecycleComplete = thread && threadHasTurn && !active && s.TurnsStarted == s.TurnsCompleted+s.TurnsFailed
	if s.LifecycleComplete {
		s.TraceStatus = "completed_turns"
	}
	if s.TurnsFailed > 0 || s.ErrorEvents > 0 || s.StartupErrorItems > 0 {
		s.TraceStatus = "observed_failure"
	}
	if s.UnknownControlEvents > 0 {
		s.TraceStatus = "unknown_controls"
	}
	complete := s.LifecycleComplete && s.TurnsFailed == 0 && s.ErrorEvents == 0 && s.StartupErrorItems == 0 && s.UnknownControlEvents == 0
	for _, f := range []*FieldUsage{&s.Usage.Input, &s.Usage.Cached, &s.Usage.Output, &s.Usage.Reasoning} {
		f.Complete = complete && f.ObservedTurns == s.TurnsStarted
		if f.Complete {
			v := *f.ObservedTotal
			f.Total = &v
		}
	}
	s.UsageComplete = s.Usage.Input.Complete && s.Usage.Cached.Complete && s.Usage.Output.Complete
	return s, nil
}

func add(f *FieldUsage, v *int64) error {
	if v == nil {
		f.MissingTurns++
		return nil
	}
	prior := int64(0)
	if f.ObservedTotal != nil {
		prior = *f.ObservedTotal
	}
	if *v > math.MaxInt64-prior {
		return ErrOverflow
	}
	sum := prior + *v
	f.ObservedTotal = &sum
	f.ObservedTurns++
	return nil
}

func parseEvent(body []byte) (event, error) {
	var e event
	if !utf8.Valid(body) {
		return e, ErrJSON
	}
	d := json.NewDecoder(bytes.NewReader(body))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return e, ErrJSON
	}
	var seen uint8
	var rawUsage json.RawMessage
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return e, ErrJSON
		}
		name, ok := key.(string)
		if !ok {
			return e, ErrJSON
		}
		var bit uint8
		switch name {
		case "type":
			bit = 1
		case "usage":
			bit = 2
		case "thread_id":
			bit = 4
		case "item":
			bit = 8
		case "error":
			bit = 16
		}
		if bit != 0 && seen&bit != 0 {
			return e, ErrDuplicate
		}
		seen |= bit
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return e, ErrJSON
		}
		switch name {
		case "type":
			if json.Unmarshal(value, &e.kind) != nil || e.kind == "" || len(e.kind) > 128 {
				return e, ErrJSON
			}
		case "usage":
			rawUsage = value
		case "item":
			e.item = value
		}
	}
	if _, err := d.Token(); err != nil || seen&1 == 0 {
		return e, ErrJSON
	}
	if _, err := d.Token(); err != io.EOF {
		return e, ErrJSON
	}
	// Validate even unused usage objects to reject duplicate relevant usage keys;
	// only turn.completed contributes observations.
	if len(rawUsage) != 0 {
		values, err := parseUsage(rawUsage)
		if err != nil {
			return e, err
		}
		if e.kind == "turn.completed" {
			e.usage = values
		}
	}
	return e, nil
}

// Decode only the startup diagnostic's control type. ID/message values are
// checked for string shape and immediately discarded, never decoded or stored.
// Existing active-turn items retain their opaque, bounded discard behavior.
func startupErrorItem(body []byte) (bool, error) {
	d := json.NewDecoder(bytes.NewReader(body))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return false, ErrJSON
	}
	var seen uint8
	kind := ""
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return false, ErrJSON
		}
		name, ok := key.(string)
		if !ok {
			return false, ErrJSON
		}
		var bit uint8
		switch name {
		case "type":
			bit = 1
		case "id":
			bit = 2
		case "message":
			bit = 4
		}
		if bit != 0 && seen&bit != 0 {
			return false, ErrDuplicate
		}
		seen |= bit
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return false, ErrJSON
		}
		if name == "type" {
			if json.Unmarshal(value, &kind) != nil || kind == "" || len(kind) > 128 {
				return false, ErrJSON
			}
		}
		if name == "id" || name == "message" {
			trimmed := bytes.TrimSpace(value)
			if len(trimmed) == 0 || trimmed[0] != '"' {
				return false, ErrJSON
			}
		}
	}
	if _, err := d.Token(); err != nil || seen&7 != 7 {
		return false, ErrJSON
	}
	if _, err := d.Token(); err != io.EOF {
		return false, ErrJSON
	}
	return kind == "error", nil
}

func parseUsage(body []byte) ([4]*int64, error) {
	var values [4]*int64
	if bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
		return values, nil
	}
	d := json.NewDecoder(bytes.NewReader(body))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return values, ErrUsage
	}
	var seen uint8
	var keys [64]string
	keyCount := 0
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return values, ErrUsage
		}
		key, ok := t.(string)
		if !ok {
			return values, ErrUsage
		}
		// Also reject repeated future/unknown usage keys. This small fixed array
		// bounds metadata work; no item contents or untrusted names escape parsing.
		if keyCount == len(keys) || len(key) > 128 {
			return values, ErrUsage
		}
		for _, prior := range keys[:keyCount] {
			if prior == key {
				return values, ErrDuplicate
			}
		}
		keys[keyCount], keyCount = key, keyCount+1
		index := -1
		switch key {
		case "input_tokens":
			index = 0
		case "cached_input_tokens":
			index = 1
		case "output_tokens":
			index = 2
		case "reasoning_output_tokens":
			index = 3
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return values, ErrUsage
		}
		if index < 0 {
			continue
		}
		bit := uint8(1 << index)
		if seen&bit != 0 {
			return values, ErrDuplicate
		}
		seen |= bit
		if bytes.Equal(raw, []byte("null")) {
			continue
		}
		for _, b := range raw {
			if b < '0' || b > '9' {
				return values, ErrUsage
			}
		}
		v, err := strconv.ParseInt(string(raw), 10, 64)
		if err != nil {
			return values, ErrUsage
		}
		values[index] = &v
	}
	if _, err := d.Token(); err != nil {
		return values, ErrUsage
	}
	if _, err := d.Token(); err != io.EOF {
		return values, ErrUsage
	}
	if values[0] != nil && values[1] != nil && *values[1] > *values[0] {
		return values, ErrUsage
	}
	if values[2] != nil && values[3] != nil && *values[3] > *values[2] {
		return values, ErrUsage
	}
	return values, nil
}
