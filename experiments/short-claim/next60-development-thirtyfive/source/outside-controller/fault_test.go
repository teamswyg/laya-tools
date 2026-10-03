// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
// Inert owned synthetic controls. No child/original/package is invoked.
package main

import (
	"bytes"
	"encoding/json"
	core "riido.local/next60gjsontwooutside/twoliteral"
	"testing"
)

type fakeJournal struct {
	bytes.Buffer
	fail bool
}

func (f *fakeJournal) Sync() error {
	if f.fail {
		return errCode("synthetic_sync")
	}
	return nil
}
func (f *fakeJournal) Close() error { return nil }

type noACK struct{ n int }

func (w *noACK) Write(b []byte) (int, error) { w.n++; return 0, errCode("synthetic_ack") }
func readyLine() []byte {
	v := false
	b, _ := json.Marshal(core.Frame{Version: 2, Sequence: 1, Previous: zeroSHA, Stage: 9, Dispatch: -1, Request: -1, Fixture: -1, Display: -1, Internal: -1, DisableEscapeHTML: &v})
	return append(b, '\n')
}
func TestSyncThenACKFailurePrefixes(t *testing.T) {
	for _, syncFails := range []bool{true, false} {
		f := &fakeJournal{fail: syncFails}
		j := &DurableJournal{file: f, dir: "synthetic", FrameBound: 384, parentSync: func(string) error { return nil }}
		s := ProtocolState{}
		w := &noACK{}
		e := consumeProtocol(bytes.NewReader(readyLine()), w, j, &s, zeroSHA, func() error { return nil })
		if e == nil || f.Len() == 0 || s.Sequence != 0 || s.FramesACKWritten != 0 {
			t.Fatal("failure cannot fabricate ACK or erase physical bytes")
		}
		if syncFails {
			if j.Receipt.FramesSynced != 0 || w.n != 0 || s.LastDurableCounts != nil {
				t.Fatal("Sync failure was acknowledged")
			}
		} else if j.Receipt.FramesSynced != 1 || w.n != 1 || s.LastDurableCounts == nil {
			t.Fatal("durable but unACKed frame was lost")
		}
	}
}
func TestNestedCallbackDecisionLifecycle(t *testing.T) {
	row := core.Row{Ordinal: 0, Request: 0, Fixture: 0, Display: 0, Internal: 0, RequestID: "synthetic", FixtureID: "s0", InputPointer: "/fake", Input: json.RawMessage(`{}`), Want: json.RawMessage(`{}`), CandidateInvoked: true}
	s := ProtocolState{Sequence: 3, RowActive: true, InvocationActive: true, MethodActive: true, Method: 1, Current: row, Counts: core.Counts{Row: [4]int{1, 1, 0, 0}, Candidate: [4]int{1, 1, 0, 0}, Original: [4]int{1, 0, 0, 0}}}
	s.Carrier.Fixtures[0] = core.Fixture{RequestID: "synthetic", ID: "s0", Input: row.Input}
	s.Carrier.Fixtures[0].InputReference.Pointer = "/fake"
	s.Carrier.Wants[0].Want = row.Want
	c := core.Callback{PrimitiveKnown: true, Primitive: core.Primitive{Type: 2, Raw: "1", Index: "0"}}
	row.Got.CallbackN = 1
	row.Got.Callbacks[0] = c
	counts := s.Counts
	counts.Original[1] = 1
	counts.Callback[0] = 1
	before := core.Frame{Version: 2, Sequence: 4, Previous: zeroSHA, Stage: 7, Counts: counts, Row: &row, Callback: &c}
	next, e := s.Next(before, zeroSHA)
	if e != nil || !next.CallbackActive || !next.MethodActive {
		t.Fatal("callback entry must retain active native method")
	}
	c.ContinueKnown = true
	c.Continue = false
	row.Got.Callbacks[0] = c
	counts.Callback = [4]int{1, 1, 1, 0}
	after := before
	after.Sequence = 5
	after.Stage = 8
	after.Counts = counts
	after.Row = &row
	after.Callback = &c
	next, e = next.Next(after, zeroSHA)
	if e != nil || next.CallbackActive || !next.MethodActive {
		t.Fatal("owned false decision closes before native method return")
	}
	after.Counts.Callback[1] = 0
	if _, e = s.Next(after, zeroSHA); e == nil {
		t.Fatal("unpaired callback must fail")
	}
}
func TestFinalPersistenceDowngrade(t *testing.T) {
	v := 33
	r := OutsideResult{State: "complete_observation_transport", CompleteRows: &v, CompleteCountersKnown: true, FramesACKWritten: 8, DurablePrefix: DurableReceipt{FramesSynced: 9}}
	calls := 0
	got, e := persistResult("synthetic", nil, r, func(string, string, []byte, int) error { calls++; return errCode("synthetic_store") })
	if e == nil || calls != 1 || got.State == r.State || got.CompleteRows != nil || got.CompleteCountersKnown || !got.RemainingCallsUnknown || got.FramesACKWritten != 8 || got.DurablePrefix.FramesSynced != 9 {
		t.Fatal("final save failure must downgrade completion while retaining prefix")
	}
}
