// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
// Unexecuted owned synthetic controls; never read real saved files or Want carriers.
package compare

import (
	"encoding/json"
	core "riido.local/next60gjsonsavedcomparison/twoliteral"
	"strings"
	"testing"
)

func encoded(v any) json.RawMessage {
	b, e := json.Marshal(v)
	if e != nil {
		panic("owned_synthetic_encode")
	}
	return b
}
func appendExample() (core.Row, AppendWant) {
	w := AppendWant{Normal: true, Hex: "5b", Active: "5b", Backing: "5b" + strings.Repeat("a5", 31), ErrorNil: false, Identity: "invalid_utf8"}
	r := core.Row{Request: 1, Done: true, CandidateInvoked: true, Want: encoded(w), Got: core.Got{NormalKnown: true, Normal: true, Panic: core.Panic{Known: true}, BufferKnown: true, ReturnedHex: "5b", ActiveInputAfterKnown: true, ActiveInputAfterHex: "5b", BackingAfterKnown: true, BackingAfterHex: w.Backing, Error: core.ErrorState{Available: true, Known: true, Present: true, OwnIdentityKnown: true, OwnIdentity: "invalid_utf8"}}}
	return r, w
}
func get(r RowResult, key string) Predicate {
	for i := 0; i < r.PredicateCount; i++ {
		if r.Predicates[i].Key == key {
			return r.Predicates[i]
		}
	}
	panic("synthetic_missing_key")
}
func TestUnknownErrorNeverNegativeButKnownMutationStillFalse(t *testing.T) {
	r, _ := appendExample()
	r.Got.Error = core.ErrorState{}
	v, e := rowCompare(r)
	if e != nil || v.State != "U" || v.Counts.F != 0 || get(v, "error_nil").State != "U" {
		t.Fatal("unavailable error became negative")
	}
	r.Got.BackingAfterHex = "5b22" + strings.Repeat("a5", 30)
	v, e = rowCompare(r)
	if e != nil || v.State != "F" || v.Counts.U != 2 || get(v, "full_owned_backing_after_hex").State != "F" {
		t.Fatal("knownF+U erased")
	}
}
func TestIdentityAvailabilityIsNotErrorPresence(t *testing.T) {
	for _, x := range []struct {
		known bool
		id    string
		state string
	}{{false, "", "U"}, {true, "invalid_line", "F"}, {true, "invalid_utf8", "T"}} {
		r, _ := appendExample()
		r.Got.Error.OwnIdentityKnown = x.known
		r.Got.Error.OwnIdentity = x.id
		v, e := rowCompare(r)
		if e != nil || get(v, "error_nil").State != "T" || get(v, "owned_error_identity").State != x.state {
			t.Fatal("identity was guessed from presence or type")
		}
	}
}
func TestKnownPanicHasNormalFalseAndOtherPredicatesUnknown(t *testing.T) {
	r, _ := appendExample()
	r.Got.Normal = false
	r.Got.Panic = core.Panic{Known: true, Present: true, TypeKnown: true, Type: "synthetic", Scope: "owned_candidate"}
	v, e := rowCompare(r)
	if e != nil || v.Counts.F != 2 || v.Counts.U != 6 || get(v, "full_owned_backing_after_hex").State != "U" {
		t.Fatal("panic fabricated normal-return facts")
	}
}
func TestOriginalIndexZeroKeepsPrimitiveButLineAndTailUnknown(t *testing.T) {
	w := LineWant{Normal: true, Callbacks: []ExpectedCallback{{Line: 2, Raw: "1", Type: 2}}, Identity: "none", ErrorNil: true, Terminal: "exhausted"}
	r := core.Row{Request: 0, Done: true, CandidateInvoked: true, Want: encoded(w), Got: core.Got{NormalKnown: true, Normal: true, Panic: core.Panic{Known: true}, CallbackN: 1, MappingStopped: true}}
	r.Got.Callbacks[0] = core.Callback{PrimitiveKnown: true, Primitive: core.Primitive{Raw: "1", Type: 2, Index: "0"}, ContinueKnown: true}
	v, e := rowCompare(r)
	if e != nil || v.State != "U" || get(v, "callbacks.0.raw").State != "T" || get(v, "callbacks.0.line").State != "U" || get(v, "callbacks.length").State != "U" {
		t.Fatal("unmapped span became false or full list")
	}
	w.Callbacks = []ExpectedCallback{}
	r.Want = encoded(w)
	v, e = rowCompare(r)
	if e != nil || get(v, "callbacks.length").State != "F" || v.Counts.U == 0 {
		t.Fatal("observed callback beyond wanted list lost")
	}
}
func TestExplicitFalseAndAllKnownPositiveDoNotDependOnPosition(t *testing.T) {
	r, _ := appendExample()
	r.Display = 0
	v, e := rowCompare(r)
	if e != nil || v.Counts.T != 8 || v.Counts.U != 0 || v.State != "T" || get(v, "returned_destination_nil").State != "T" {
		t.Fatal("false literal or positive got source-position privilege")
	}
	r.Got.ReturnedNil = true
	v, e = rowCompare(r)
	if e != nil || get(v, "returned_destination_nil").State != "F" {
		t.Fatal("false interpreted as skip")
	}
}
func TestMalformedTypeAndInactivePaddingRejected(t *testing.T) {
	r, _ := appendExample()
	r.Want = json.RawMessage(`{"returned_normally":true,"panic_present":false,"returned_destination_hex":"5b","returned_destination_nil":null,"active_input_destination_after_hex":"5b","full_owned_backing_after_hex":"5b","error_nil":false,"owned_error_identity":"invalid_utf8"}`)
	if _, e := rowCompare(r); e == nil {
		t.Fatal("null bool or wrong backing accepted")
	}
	r, _ = appendExample()
	r.Got.Callbacks[2] = core.Callback{PrimitiveKnown: true}
	if _, e := rowCompare(r); e == nil {
		t.Fatal("inactive callback padding ignored")
	}
	r, _ = appendExample()
	r.Got.BackingAfterHex = "5bff"
	if _, e := rowCompare(r); e == nil {
		t.Fatal("full backing missing bytes")
	}
}
func TestUTF8ReplacementHexIsNotMalformedInputHex(t *testing.T) {
	w := AppendWant{Normal: true, Hex: "5b22efbfbd22", Active: "5b", Backing: "5b22efbfbd22" + strings.Repeat("a5", 26), ErrorNil: true, Identity: "none"}
	r := core.Row{Request: 1, Done: true, CandidateInvoked: true, Want: encoded(w), Got: core.Got{NormalKnown: true, Normal: true, Panic: core.Panic{Known: true}, BufferKnown: true, ReturnedHex: w.Hex, ActiveInputAfterKnown: true, ActiveInputAfterHex: "5b", BackingAfterKnown: true, BackingAfterHex: w.Backing, Error: core.ErrorState{Available: true, Known: true, OwnIdentityKnown: true, OwnIdentity: "none"}}}
	v, e := rowCompare(r)
	if e != nil || v.State != "T" {
		t.Fatal("valid replacement character lost")
	}
	r.Got.ReturnedHex = "5b22ff22"
	v, e = rowCompare(r)
	if e != nil || get(v, "returned_destination_hex").State != "F" {
		t.Fatal("hex malformed bytes equated with U+FFFD")
	}
}

// This fabricated DTO tests completion gates only; it is never used by Run,
// has no actual fixture or scientific Got, and does not establish source truth.
func completedEnvelopeSynthetic() (Outside, core.Result, []byte, Config) {
	zero := 0
	yes := true
	over := false
	rss := int64(1)
	rows := 33
	c := Config{PlanSHA256: strings.Repeat("a", 64), WorkerSHA256: strings.Repeat("b", 64), ControllerSHA256: strings.Repeat("c", 64), Outside: Pin{Bytes: 100}}
	n := core.Counts{Row: [4]int{33, 33, 33, 0}, Candidate: [4]int{33, 33, 33, 0}, Original: [4]int{1, 1, 1, 0}}
	frames := 2 + 4*33 + 2*n.Original[0]
	r := core.Result{Schema: "riido-gjson-two-native-result-v1", State: "complete", PlanSHA256: c.PlanSHA256, Completed: 33, Counts: n, LastACKSequence: uint32(frames - 1), LastACKSHA256: strings.Repeat("d", 64)}
	raw := encoded(r)
	c.Child.Bytes = int64(len(raw))
	c.Child.SHA256 = Digest(raw)
	final := core.Frame{Version: 2, Sequence: r.LastACKSequence + 1, Previous: r.LastACKSHA256, Stage: 4, Dispatch: -1, Request: -1, Fixture: -1, Display: -1, Internal: -1, Counts: n, Result: raw}
	finalLine := append(encoded(final), '\n')
	journalBytes := int64(len(finalLine) + 3000)
	o := Outside{Schema: "riido-gjson-two-native-outside-result-v1", State: "complete_observation_transport", ConfigSHA256: strings.Repeat("e", 64), PlanSHA256: c.PlanSHA256, WorkerSHA256: c.WorkerSHA256, ControllerSHA256: c.ControllerSHA256,
		Process: Process{StartAttempts: 1, Started: true, WaitAttempts: 1, Reaped: &yes, ExitCode: &zero, DarwinTimeMaxRSSBytes: &rss, RSSCapExceeded: &over}, DurablePrefix: Durable{FramesSynced: frames, BytesSynced: journalBytes, LastSHA256: Digest(finalLine)}, FramesACKWritten: frames, LastDurableCounts: &n, LastACKWrittenCounts: &n, TerminalRowsACKWritten: 33, FinalACKWritten: true, FinalChildFileIdentity: true, CompleteCountersKnown: true, CompleteRows: &rows, AggregateCap: 2 << 20, NoAutomaticRetry: true}
	h := Digest(nil)
	o.Artifacts = []Artifact{{"reservation.json", 100, h}, {"journal.jsonl", journalBytes, h}, {"stderr.txt", 0, h}, {"worker-attempt/reservation.json", 100, h}, {"worker-attempt/results.json", c.Child.Bytes, c.Child.SHA256}}
	retained := int64(300) + journalBytes + c.Child.Bytes
	o.OutputBytesRetained = &retained
	return o, r, raw, c
}
func TestCompletionRequiresACKReapFileAndFrameIdentity(t *testing.T) {
	o, r, b, c := completedEnvelopeSynthetic()
	if completion(o, r, b, c) != nil {
		t.Fatal("completion control basis rejected")
	}
	for _, f := range []func(*Outside){func(v *Outside) { v.FinalACKWritten = false }, func(v *Outside) { v.FinalChildFileIdentity = false }, func(v *Outside) { v.Process.Reaped = nil }, func(v *Outside) { v.LastDurableCounts = nil }, func(v *Outside) { v.DurablePrefix.LastSHA256 = strings.Repeat("f", 64) }, func(v *Outside) { v.FramesACKWritten++ }, func(v *Outside) { v.Artifacts = append(v.Artifacts, Artifact{"unexpected", 0, Digest(nil)}) }} {
		o, r, b, c = completedEnvelopeSynthetic()
		f(&o)
		if completion(o, r, b, c) != ErrCompletion {
			t.Fatal("completion did not reject missing ACK or contradictory correspondence")
		}
	}
}
func TestAbsentCallbackAndNilErrorIdentityStayUnknown(t *testing.T) {
	w := LineWant{Normal: true, Callbacks: []ExpectedCallback{{Line: 1, Raw: "1", Type: 2}}, Identity: "none", ErrorNil: true, Terminal: "exhausted"}
	r := core.Row{Request: 0, Done: true, CandidateInvoked: true, Want: encoded(w), Got: core.Got{NormalKnown: true, Normal: true, Panic: core.Panic{Known: true}, CallbackN: 1, CallbackListCompleteKnown: true, CallbackListComplete: true, Error: core.ErrorState{Available: true, Known: true}}}
	v, e := rowCompare(r)
	if e != nil || get(v, "callbacks.0.raw").State != "U" || get(v, "callbacks.0.type").State != "U" || get(v, "error_nil").State != "T" || get(v, "owned_error_identity").State != "U" {
		t.Fatal("absent channel became malformed or guessed")
	}
	r, _ = appendExample()
	r.Got.ActiveInputAfterHex = "5d"
	v, e = rowCompare(r)
	if e != nil || get(v, "active_input_destination_after_hex").State != "F" {
		t.Fatal("wrong known active byte was rejected instead of compared")
	}
}
func TestStrictJSONRejectsAliasedDuplicateAndExtraArray(t *testing.T) {
	var counts core.Counts
	for _, raw := range []string{`{"r":[0,0,0,0],"b":[0,0,0,0],"o":[0,0,0,0],"k":[0,0,0,0,99]}`, `{"r":[0,0,0,0],"r":[0,0,0,0],"b":[0,0,0,0],"o":[0,0,0,0],"k":[0,0,0,0]}`, `{"R":[0,0,0,0],"b":[0,0,0,0],"o":[0,0,0,0],"k":[0,0,0,0]}`} {
		if compact([]byte(raw), &counts) == nil {
			t.Fatal("decoder discarded undeclared or extra array values")
		}
	}
}
