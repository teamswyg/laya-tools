package taskoutcome

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
)

const start = "{\"type\":\"thread.started\",\"thread_id\":\"private-id\"}\n{\"type\":\"turn.started\"}\n"
const completed = "{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":10,\"cached_input_tokens\":4,\"output_tokens\":3}}\n"

func read(t *testing.T, trace string) Summary {
	t.Helper()
	s, err := Summarize(strings.NewReader(trace), Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestMultiTurnExactHashAndUnknownIdentity(t *testing.T) {
	trace := start + completed + "{\"type\":\"turn.started\"}\n" +
		"{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":20,\"cached_input_tokens\":5,\"output_tokens\":7,\"reasoning_output_tokens\":2}}"
	s, err := Summarize(strings.NewReader(trace), Metadata{PublicTaskLabel: "authored-test", RequestedModel: "example-model", RequestedReasoning: "high"})
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256([]byte(trace))
	if s.TraceSHA256 != hex.EncodeToString(h[:]) || s.TraceBytes != int64(len(trace)) || s.Events != 5 || s.TurnsCompleted != 2 || !s.UsageComplete || !s.LifecycleComplete {
		t.Fatalf("bad aggregate: %+v", s)
	}
	if *s.Usage.Input.Total != 30 || *s.Usage.Cached.Total != 9 || *s.Usage.Output.Total != 10 || s.Usage.Reasoning.Total != nil || *s.Usage.Reasoning.ObservedTotal != 2 {
		t.Fatal("reasoning was counted twice or optional observations invented")
	}
	if s.ObservedModel != "unknown" || s.TaskAcceptance != "unknown" || s.ProcessExit != "unknown" || s.TraceProvenance != "unknown" || s.ExecutionTime != "unknown" || s.Metadata.RequestedModel != "example-model" {
		t.Fatal("request metadata became observed execution or acceptance")
	}
}

func TestMissingUsagePreservesKnownFields(t *testing.T) {
	for _, tc := range []struct{ name, terminal string }{
		{"missing", `{"type":"turn.completed"}`},
		{"null", `{"type":"turn.completed","usage":null}`},
		{"empty", `{"type":"turn.completed","usage":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := read(t, start+tc.terminal)
			if s.UsageComplete || s.Usage.Input.ObservedTotal != nil || s.Usage.Output.Total != nil || s.Usage.Input.MissingTurns != 1 {
				t.Fatal("missing usage invented zero")
			}
		})
	}
	s := read(t, start+`{"type":"turn.completed","usage":{"input_tokens":10,"output_tokens":0}}`)
	if s.UsageComplete || !s.Usage.Input.Complete || *s.Usage.Input.Total != 10 || !s.Usage.Output.Complete || *s.Usage.Output.Total != 0 || s.Usage.Cached.ObservedTotal != nil {
		t.Fatal("missing cached usage erased measured input/output or invented cached zero")
	}
	s = read(t, start+completed+"{\"type\":\"turn.started\"}\n{\"type\":\"turn.completed\",\"usage\":null}")
	if *s.Usage.Input.ObservedTotal != 10 || s.Usage.Input.Total != nil || s.Usage.Input.ObservedTurns != 1 || s.Usage.Input.MissingTurns != 1 {
		t.Fatal("partial prefix was lost or reported as whole trace usage")
	}
}

func TestFailureIncompleteAndUnknownControl(t *testing.T) {
	for _, tc := range []struct{ name, suffix, status string }{
		{"failed", "{\"type\":\"turn.started\"}\n{\"type\":\"turn.failed\",\"error\":{\"message\":\"private detail\"}}", "observed_failure"},
		{"error", `{"type":"error","message":"private detail"}`, "observed_failure"},
		{"incomplete", `{"type":"turn.started"}`, "incomplete"},
		{"future", `{"type":"future.control","payload":"private detail"}`, "unknown_controls"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := read(t, start+completed+tc.suffix)
			if s.UsageComplete || s.Usage.Input.Total != nil || *s.Usage.Input.ObservedTotal != 10 || s.TraceStatus != tc.status {
				t.Fatalf("bad incomplete/failure handling: %+v", s)
			}
		})
	}
	s := read(t, start+`{"type":"turn.failed"}`+"\n"+start+completed)
	if s.ThreadsStarted != 2 || s.TurnsStarted != 2 || s.TurnsFailed != 1 || s.TurnsCompleted != 1 || s.UsageComplete {
		t.Fatal("sequential failed attempt was omitted")
	}
	s = read(t, "\n \t\n")
	if s.UsageComplete || s.LifecycleComplete || s.Usage.Input.ObservedTotal != nil {
		t.Fatal("empty trace invented completion or usage")
	}
}

func TestMalformedRedactedErrors(t *testing.T) {
	for _, tc := range []struct {
		name, trace string
		code        ErrorCode
	}{
		{"truncated", start + `{"type":"turn.completed","usage":`, ErrJSON},
		{"trailing", start + completed + `{"type":"error"} {"type":"error"}`, ErrJSON},
		{"nonobject", `[]`, ErrJSON},
		{"missingtype", `{"private":"/Users/private/secret"}`, ErrJSON},
		{"duplicate-type", `{"type":"thread.started","\u0074ype":"error"}`, ErrDuplicate},
		{"duplicate-usage", start + `{"type":"turn.completed","usage":null,"usage":{}}`, ErrDuplicate},
		{"duplicate-input", start + `{"type":"turn.completed","usage":{"input_tokens":1,"input_tokens":2}}`, ErrDuplicate},
		{"duplicate-unknown-usage", start + `{"type":"turn.completed","usage":{"future":1,"future":2}}`, ErrDuplicate},
		{"duplicate-item", start + `{"type":"item.updated","item":{},"item":{}}`, ErrDuplicate},
		{"negative", start + `{"type":"turn.completed","usage":{"input_tokens":-1}}`, ErrUsage},
		{"float", start + `{"type":"turn.completed","usage":{"input_tokens":1.0}}`, ErrUsage},
		{"exponent", start + `{"type":"turn.completed","usage":{"input_tokens":1e2}}`, ErrUsage},
		{"string", start + `{"type":"turn.completed","usage":{"input_tokens":"10"}}`, ErrUsage},
		{"value-overflow", start + `{"type":"turn.completed","usage":{"input_tokens":9223372036854775808}}`, ErrUsage},
		{"cached-subset", start + `{"type":"turn.completed","usage":{"input_tokens":1,"cached_input_tokens":2}}`, ErrUsage},
		{"reasoning-subset", start + `{"type":"turn.completed","usage":{"output_tokens":1,"reasoning_output_tokens":2}}`, ErrUsage},
		{"completed-without-start", `{"type":"turn.completed"}`, ErrLifecycle},
		{"overlap", start + `{"type":"turn.started"}`, ErrLifecycle},
		{"second-terminal", start + completed + `{"type":"turn.failed"}`, ErrLifecycle},
		{"orphan-item", `{"type":"item.completed","item":{}}`, ErrLifecycle},
		{"invalid-utf8", "{\"type\":\"thread.started\",\"id\":\"\xff\"}", ErrJSON},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := Summarize(strings.NewReader(tc.trace), Metadata{})
			if err != tc.code || s.Schema != "" {
				t.Fatalf("expected fixed %s, got %v", tc.code, err)
			}
			if strings.Contains(err.Error(), "/Users/") || strings.Contains(err.Error(), "secret") {
				t.Fatal("untrusted diagnostic echoed")
			}
		})
	}
}

func TestUsageTotalOverflow(t *testing.T) {
	trace := start + `{"type":"turn.completed","usage":{"input_tokens":` + strconv.FormatInt(math.MaxInt64, 10) + `}}` +
		"\n{\"type\":\"turn.started\"}\n{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":1}}"
	if _, err := Summarize(strings.NewReader(trace), Metadata{}); err != ErrOverflow {
		t.Fatalf("expected total overflow, got %v", err)
	}
}

func TestBounds(t *testing.T) {
	trace := start + completed
	base := limits{len(trace), MaxLineBytes, 3, 1}
	if _, err := summarize(strings.NewReader(trace), Metadata{}, base); err != nil {
		t.Fatal("exact bounds rejected", err)
	}
	for _, tc := range []struct {
		name  string
		bound limits
		code  ErrorCode
	}{
		{"whole", limits{len(trace) - 1, MaxLineBytes, 3, 1}, ErrTraceSize},
		{"line", limits{len(trace), 16, 3, 1}, ErrLineSize},
		{"events", limits{len(trace), MaxLineBytes, 2, 1}, ErrEvents},
		{"turns", limits{len(trace) * 2, MaxLineBytes, 6, 1}, ErrTurns},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := trace
			if tc.name == "turns" {
				input += start + completed
			}
			if _, err := summarize(strings.NewReader(input), Metadata{}, tc.bound); err != tc.code {
				t.Fatalf("expected %s, got %v", tc.code, err)
			}
		})
	}
	// Exercise the real ReadSlice multi-buffer boundary and default line limit.
	if _, err := Summarize(strings.NewReader(strings.Repeat(" ", MaxLineBytes)), Metadata{}); err != nil {
		t.Fatal("exact maximum whitespace line rejected", err)
	}
	if _, err := Summarize(strings.NewReader(strings.Repeat(" ", MaxLineBytes+1)), Metadata{}); err != ErrLineSize {
		t.Fatal("default line bound not enforced", err)
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) {
	return 0, errors.New("/Users/private/token=private-secret")
}

func TestNoRawRetentionAndReaderErrors(t *testing.T) {
	secret := "/Users/private/ghp_private_secret"
	terminal := strings.Replace(completed, `"usage":`, `"model":"private-model","task_accepted":true,"usage":`, 1)
	trace := strings.ReplaceAll(start, "private-id", secret) + `{"type":"item.future","item":{"command":"` + secret + `","message":"private source"}}` + "\n" + terminal
	s := read(t, trace)
	b, err := json.Marshal(s)
	if err != nil || strings.Contains(string(b), secret) || strings.Contains(string(b), "private source") || strings.Contains(string(b), "private-model") || strings.Contains(string(b), "command") || s.DiscardedItemEvents != 1 || !s.UsageComplete || s.TaskAcceptance != "unknown" || s.ObservedModel != "unknown" {
		t.Fatal("raw item/thread content retained")
	}
	if _, err := Summarize(brokenReader{}, Metadata{}); err != ErrRead {
		t.Fatal("reader error not redacted", err)
	}
	for _, label := range []string{"/Users/private", "token=private", "ghp_private", "hf_private", strings.Repeat("a", 97)} {
		if _, err := Summarize(strings.NewReader(trace), Metadata{RequestedModel: label}); err != ErrMetadata {
			t.Fatal("unsafe metadata accepted")
		}
	}
}

func TestIndependentCallOwnership(t *testing.T) {
	for i := range 24 {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			a := read(t, start+completed)
			b := read(t, start+completed)
			*a.Usage.Input.ObservedTotal = 999
			if *a.Usage.Input.Total != 10 || *b.Usage.Input.ObservedTotal != 10 || *b.Usage.Input.Total != 10 {
				t.Fatal("calls or complete/observed values share mutable storage")
			}
		})
	}
}

var _ io.Reader = brokenReader{}

// Authored schema fixtures reproduce the observed event classes/order only;
// no real CLI trace, provider message or account ID is embedded here.
const startupThread = `{"type":"thread.started","thread_id":"authored-startup-thread"}` + "\n"
const startupDiagnostic = `{"type":"item.completed","item":{"id":"authored-diagnostic-item","message":"authored startup diagnostic body","type":"error"}}` + "\n"

func TestStartupErrorBeforeFirstTurnSummarizesObservedFailure(t *testing.T) {
	trace := startupThread + startupDiagnostic + `{"type":"turn.started"}` + "\n" + `{"type":"error","message":"authored terminal diagnostic body"}` + "\n" + `{"type":"turn.failed"}` + "\n"
	s := read(t, trace)
	if s.Events != 5 || s.ThreadsStarted != 1 || s.StartupErrorItems != 1 || s.DiscardedItemEvents != 0 || s.TurnsStarted != 1 || s.TurnsFailed != 1 || s.ErrorEvents != 1 || !s.LifecycleComplete || s.TraceStatus != "observed_failure" || s.UsageComplete {
		t.Fatal("startup diagnostic was rejected or promoted to completion")
	}
	for _, field := range []FieldUsage{s.Usage.Input, s.Usage.Cached, s.Usage.Output, s.Usage.Reasoning} {
		if field.ObservedTotal != nil || field.Total != nil || field.Complete || field.ObservedTurns != 0 {
			t.Fatal("missing usage invented zero")
		}
	}
	b, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	for _, discarded := range []string{"authored-startup-thread", "authored-diagnostic-item", "authored startup diagnostic body", "authored terminal diagnostic body"} {
		if strings.Contains(string(b), discarded) {
			t.Fatal("startup ID/message retained")
		}
	}
	h := sha256.Sum256([]byte(trace))
	if s.TraceSHA256 != hex.EncodeToString(h[:]) || s.TraceBytes != int64(len(trace)) {
		t.Fatal("startup event bytes omitted from trace binding")
	}
	if s.ObservedModel != "unknown" || s.ProcessExit != "unknown" || s.TaskAcceptance != "unknown" {
		t.Fatal("startup evidence became provider/process/acceptance attestation")
	}
}

func TestStartupErrorKeepsPartialObservationsButWholeUsageUnknown(t *testing.T) {
	for _, suffix := range []string{completed, completed + `{"type":"turn.started"}` + "\n" + `{"type":"turn.failed"}`, completed + `{"type":"turn.started"}` + "\n" + `{"type":"turn.completed","usage":null}`} {
		s := read(t, startupThread+startupDiagnostic+`{"type":"turn.started"}`+"\n"+suffix)
		if s.StartupErrorItems != 1 || s.UsageComplete || s.Usage.Input.Total != nil || s.Usage.Cached.Total != nil || s.Usage.Output.Total != nil || s.Usage.Input.ObservedTotal == nil || *s.Usage.Input.ObservedTotal != 10 || *s.Usage.Cached.ObservedTotal != 4 || *s.Usage.Output.ObservedTotal != 3 || s.TraceStatus != "observed_failure" {
			t.Fatal("later completed turn erased startup failure or partial usage")
		}
	}
	s := read(t, startupThread+startupDiagnostic)
	if s.LifecycleComplete || s.UsageComplete || s.Usage.Input.Total != nil || s.TraceStatus != "observed_failure" {
		t.Fatal("startup-only diagnostic invented a turn")
	}
}

func TestStartupExceptionRejectsOtherItemsAndWrongPhase(t *testing.T) {
	for _, trace := range []string{
		startupDiagnostic,
		startupThread + strings.Replace(startupDiagnostic, `"type":"error"`, `"type":"agent_message"`, 1),
		startupThread + strings.Replace(startupDiagnostic, `"type":"item.completed"`, `"type":"item.updated"`, 1),
		start + completed + startupDiagnostic,
		startupThread + startupDiagnostic + `{"type":"thread.started"}`,
	} {
		if s, e := Summarize(strings.NewReader(trace), Metadata{}); e != ErrLifecycle || s.Schema != "" {
			t.Fatal("startup exception admitted orphan/other/late item or empty duplicate thread")
		}
	}
	// The same opaque diagnostic during an active turn keeps its existing item
	// behavior; startup-specific controls do not inspect or reclassify it.
	s := read(t, start+startupDiagnostic+completed)
	if s.StartupErrorItems != 0 || s.DiscardedItemEvents != 1 || !s.UsageComplete {
		t.Fatal("active item behavior changed")
	}
}

func TestStartupDiagnosticControlShapeAndDuplicates(t *testing.T) {
	for _, tc := range []struct {
		item string
		code ErrorCode
	}{
		{`null`, ErrJSON}, {`[]`, ErrJSON}, {`{}`, ErrJSON},
		{`{"type":"error","id":1,"message":"authored"}`, ErrJSON},
		{`{"type":"error","id":"authored","message":{}}`, ErrJSON},
		{`{"type":"error","id":"authored"}`, ErrJSON},
		{`{"type":null,"id":"authored","message":"authored"}`, ErrJSON},
		{`{"type":"error","\u0074ype":"error","id":"authored","message":"authored"}`, ErrDuplicate},
		{`{"type":"error","id":"authored","id":"other","message":"authored"}`, ErrDuplicate},
		{`{"type":"error","id":"authored","message":"authored","message":"other"}`, ErrDuplicate},
	} {
		trace := startupThread + `{"type":"item.completed","item":` + tc.item + `}`
		if s, e := Summarize(strings.NewReader(trace), Metadata{}); e != tc.code || s.Schema != "" {
			t.Fatalf("startup control expected %s, got %v", tc.code, e)
		}
	}
	// Existing top-level duplicate and usage checks still apply before startup
	// lifecycle handling, rather than silently ignoring diagnostic metadata.
	for _, tc := range []struct {
		event string
		code  ErrorCode
	}{
		{`{"type":"item.completed","item":{"id":"a","message":"b","type":"error"},"item":{"id":"c","message":"d","type":"error"}}`, ErrDuplicate},
		{`{"type":"item.completed","item":{"id":"a","message":"b","type":"error"},"usage":{"input_tokens":-1}}`, ErrUsage},
	} {
		if _, e := Summarize(strings.NewReader(startupThread+tc.event), Metadata{}); e != tc.code {
			t.Fatal("startup exception bypassed existing duplicate/usage guard")
		}
	}
}

func TestStartupDiagnosticRemainsWithinTraceBounds(t *testing.T) {
	trace := startupThread + startupDiagnostic + `{"type":"turn.started"}` + "\n" + completed
	if _, e := summarize(strings.NewReader(trace), Metadata{}, limits{len(trace), MaxLineBytes, 4, 1}); e != nil {
		t.Fatal("bounded startup sequence rejected", e)
	}
	for _, tc := range []struct {
		bound limits
		code  ErrorCode
	}{{limits{len(trace) - 1, MaxLineBytes, 4, 1}, ErrTraceSize}, {limits{len(trace), 16, 4, 1}, ErrLineSize}, {limits{len(trace), MaxLineBytes, 3, 1}, ErrEvents}} {
		if _, e := summarize(strings.NewReader(trace), Metadata{}, tc.bound); e != tc.code {
			t.Fatal("startup sequence bypassed byte/line/event bound")
		}
	}
}
