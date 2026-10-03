// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"reflect"
	core "riido.local/next60gjsontwooutside/twoliteral"
	"strconv"
)

const zeroSHA = "0000000000000000000000000000000000000000000000000000000000000000"

type ProtocolState struct {
	Sequence          uint32
	Previous          string
	Counts            core.Counts
	NextDispatch      int
	RowActive         bool
	InvocationActive  bool
	InvocationEnded   bool
	MethodActive      bool
	CallbackActive    bool
	Method            int
	Current           core.Row
	Stopped           bool
	MethodPanic       bool
	Rows              [33]core.Row
	Known             [33]bool
	Final             *core.Result
	Carrier           Carrier
	FramesACKWritten  int
	LastDurableCounts *core.Counts
}

func schedule(d int) (q, f, x, j int, ok bool) {
	if d < 0 || d >= 33 {
		return
	}
	q = 0
	fi := d / 3
	f = fi
	if fi >= 5 {
		q = 1
		f = fi - 5
	}
	x = d % 3
	j = [2][3]int{{0, 2, 1}, {1, 2, 0}}[q][x]
	return q, f, x, j, true
}
func tuple(t [4]int, max int) bool {
	return t[0] >= 0 && t[0] <= max && t[1] >= 0 && t[1] <= t[0] && t[2] >= 0 && t[2] <= max && t[3] >= 0 && t[3] <= max && t[2]+t[3] <= t[1]
}
func bounded(c core.Counts) bool {
	return tuple(c.Row, 33) && tuple(c.Candidate, 33) && tuple(c.Original, 61) && tuple(c.Callback, 33)
}
func catchDispatch(t *[4]int) {
	if t[1] < t[0] {
		t[1]++
	}
}
func allowed(q, j, m int) bool {
	if q == 1 {
		return m == 4
	}
	if j == 0 {
		return m == 1
	}
	return m == 2 || m == 3
}
func primitive(v core.Primitive) bool {
	n, e := strconv.ParseInt(v.Index, 10, 64)
	return e == nil && strconv.FormatInt(n, 10) == v.Index && len(v.Raw) <= 12 && v.Type <= 5
}
func panicShape(v core.Panic) bool {
	if !v.Known {
		return v == (core.Panic{})
	}
	if !v.Present {
		return !v.TypeKnown && v.Type == "" && v.Scope == ""
	}
	return len(v.Type) <= 128 && (!v.TypeKnown && v.Type == "" || v.TypeKnown && v.Type != "") && (v.Scope == "owned_candidate" || v.Scope == "direct_original_method")
}
func errorShape(v core.ErrorState) bool {
	if !v.Available {
		return v == (core.ErrorState{})
	}
	if !v.Known {
		return !v.Present && !v.TypeKnown && v.Type == "" && !v.OwnIdentityKnown && v.OwnIdentity == ""
	}
	if !v.Present {
		return v.OwnIdentityKnown && v.OwnIdentity == "none" && !v.TypeKnown && v.Type == ""
	}
	return (!v.TypeKnown && v.Type == "" || v.TypeKnown && len(v.Type) > 0 && len(v.Type) <= 128) && (!v.OwnIdentityKnown && v.OwnIdentity == "" || v.OwnIdentityKnown && (v.OwnIdentity == "invalid_line" || v.OwnIdentity == "invalid_utf8"))
}
func byteHex(v string, max int) bool {
	if len(v) > 2*max || len(v)%2 != 0 {
		return false
	}
	b, e := hex.DecodeString(v)
	return e == nil && hex.EncodeToString(b) == v
}
func gotShape(g core.Got) bool {
	if !panicShape(g.Panic) || !errorShape(g.Error) || g.CallbackN < 0 || g.CallbackN > 3 || g.MappedPrefix < 0 || g.MappedPrefix > g.CallbackN || !g.NormalKnown && g.Normal || !g.StopKnown && g.StopRequested || !g.FailingLineKnown && g.FailingLine != 0 || g.FailingLine < 0 || g.FailingLine > 4 || !g.TerminalKnown && g.Terminal != "" {
		return false
	}
	if g.TerminalKnown && g.Terminal != "exhausted" && g.Terminal != "invalid_line" && g.Terminal != "callback_stop" {
		return false
	}
	if !g.CallbackListCompleteKnown && g.CallbackListComplete {
		return false
	}
	for i, c := range g.Callbacks {
		if i >= g.CallbackN {
			if c != (core.Callback{}) {
				return false
			}
			continue
		}
		if !c.PrimitiveKnown || !primitive(c.Primitive) || !c.LineKnown && c.Line != 0 || c.Line < 0 || c.Line > 4 || !c.ContinueKnown && c.Continue {
			return false
		}
	}
	if !g.BufferKnown && (g.ReturnedHex != "" || g.ReturnedNil) || g.BufferKnown && !byteHex(g.ReturnedHex, 32) {
		return false
	}
	if g.BackingBeforeKnown {
		if len(g.BackingBeforeHex) != 64 || !byteHex(g.BackingBeforeHex, 32) {
			return false
		}
	} else if g.BackingBeforeHex != "" {
		return false
	}
	if g.BackingAfterKnown {
		if len(g.BackingAfterHex) != 64 || !byteHex(g.BackingAfterHex, 32) {
			return false
		}
	} else if g.BackingAfterHex != "" {
		return false
	}
	if g.ActiveInputAfterKnown {
		if g.ActiveInputAfterHex != "5b" {
			return false
		}
	} else if g.ActiveInputAfterHex != "" {
		return false
	}
	return true
}
func preserves(a, b core.Row) bool {
	if a.CandidateInvoked && !b.CandidateInvoked {
		return false
	}
	x, y := a.Got, b.Got
	if y.CallbackN < x.CallbackN || y.MappedPrefix < x.MappedPrefix || x.MappingStopped && !y.MappingStopped || x.StopRequested && !y.StopRequested {
		return false
	}
	for i := 0; i < x.CallbackN; i++ {
		old, now := x.Callbacks[i], y.Callbacks[i]
		if old.Primitive != now.Primitive || old.PrimitiveKnown != now.PrimitiveKnown || old.LineKnown && (!now.LineKnown || old.Line != now.Line) || old.ContinueKnown && (!now.ContinueKnown || old.Continue != now.Continue) {
			return false
		}
	}
	if x.NormalKnown && (!y.NormalKnown || x.Normal != y.Normal) || x.Error.Known && x.Error != y.Error {
		return false
	}
	return true
}
func factShape(m int, f *core.Fact) bool {
	if f == nil || !panicShape(f.Panic) || !f.Panic.Known || f.ErrorChannelAvailable || f.Returned == f.Panic.Present {
		return false
	}
	if !f.Returned {
		return f.Bool == nil && f.Primitive == nil && f.ReturnedHex == nil && f.ReturnedNil == nil && f.CallbackReturnComplete == nil && (f.BackingHex == nil || byteHex(*f.BackingHex, 32))
	}
	if f.OwnedBoundUnsupported {
		return !f.PrimitiveKnown && f.Bool == nil && f.Primitive == nil && f.ReturnedHex == nil && f.ReturnedNil == nil && f.CallbackReturnComplete == nil
	}
	if !f.PrimitiveKnown {
		return false
	}
	switch m {
	case 1:
		return f.CallbackReturnComplete != nil && *f.CallbackReturnComplete && f.Bool == nil && f.Primitive == nil && f.ReturnedHex == nil && f.ReturnedNil == nil && f.BackingHex == nil
	case 2:
		return f.Bool != nil && f.Primitive == nil && f.ReturnedHex == nil && f.ReturnedNil == nil && f.BackingHex == nil && f.CallbackReturnComplete == nil
	case 3:
		return f.Primitive != nil && primitive(*f.Primitive) && f.Bool == nil && f.ReturnedHex == nil && f.ReturnedNil == nil && f.BackingHex == nil && f.CallbackReturnComplete == nil
	case 4:
		return f.ReturnedHex != nil && byteHex(*f.ReturnedHex, 32) && f.ReturnedNil != nil && f.BackingHex != nil && len(*f.BackingHex) == 64 && byteHex(*f.BackingHex, 32) && f.Bool == nil && f.Primitive == nil && f.CallbackReturnComplete == nil
	}
	return false
}
func (s ProtocolState) Next(f core.Frame, plan string) (ProtocolState, error) {
	if s.Stopped || s.Final != nil || f.Version != 2 || f.Sequence != s.Sequence+1 || f.Sequence > 384 || !bounded(f.Counts) {
		return s, errCode("frame_header")
	}
	previous := s.Previous
	if previous == "" {
		previous = zeroSHA
	}
	if f.Previous != previous {
		return s, errCode("frame_chain")
	}
	if f.Stage == 9 {
		if s.Sequence != 0 || f.Dispatch != -1 || f.Request != -1 || f.Fixture != -1 || f.Display != -1 || f.Internal != -1 || f.Method != 0 || f.Counts != (core.Counts{}) || f.Row != nil || f.Fact != nil || f.Callback != nil || len(f.Result) != 0 || f.DisableEscapeHTML == nil || *f.DisableEscapeHTML {
			return s, errCode("ready_shape")
		}
		s.Sequence = f.Sequence
		return s, nil
	}
	if f.DisableEscapeHTML != nil || s.Sequence == 0 {
		return s, errCode("ready_missing_or_repeated")
	}
	if f.Stage == 4 {
		if s.RowActive || s.MethodActive || s.CallbackActive || s.InvocationActive || s.NextDispatch != 33 || f.Dispatch != -1 || f.Request != -1 || f.Fixture != -1 || f.Display != -1 || f.Internal != -1 || f.Method != 0 || f.Row != nil || f.Fact != nil || f.Callback != nil || f.Counts != s.Counts {
			return s, errCode("final_header")
		}
		var z core.Result
		if strictCompact(f.Result, &z) != nil {
			return s, errCode("final_canonical")
		}
		if z.Schema != "riido-gjson-two-native-result-v1" || z.State != "complete" || z.Failure != "" || z.PlanSHA256 != plan || z.Completed != 33 || z.Counts != s.Counts || z.LastACKSequence != s.Sequence || z.LastACKSHA256 != previous || z.OriginalInitialization != nil || z.OriginalNested != nil || z.Labels != 0 || z.Qualified != 0 || z.NewParents != 0 || z.TrainingAuthorized {
			return s, errCode("final_binding")
		}
		if z.Counts.Row != [4]int{33, 33, 33, 0} || z.Counts.Candidate != [4]int{33, 33, 33, 0} || z.Counts.Original[0] != z.Counts.Original[1] || z.Counts.Original[1] != z.Counts.Original[2] || z.Counts.Original[3] != 0 || z.Counts.Callback[0] != z.Counts.Callback[1] || z.Counts.Callback[1] != z.Counts.Callback[2] || z.Counts.Callback[3] != 0 {
			return s, errCode("final_counts")
		}
		for i, row := range z.Records {
			if !s.Known[i] || !reflect.DeepEqual(row, s.Rows[i]) || !bindRow(row, i, s.Carrier) {
				return s, errCode("final_rows")
			}
		}
		s.Final = &z
		s.Sequence = f.Sequence
		return s, nil
	}
	q, fi, x, j, ok := schedule(s.NextDispatch)
	if !ok || f.Dispatch != s.NextDispatch || f.Request != q || f.Fixture != fi || f.Display != x || f.Internal != j || len(f.Result) != 0 {
		return s, errCode("row_schedule")
	}
	if f.Row != nil {
		if !bindRow(*f.Row, s.NextDispatch, s.Carrier) || !gotShape(f.Row.Got) || s.RowActive && !preserves(s.Current, *f.Row) {
			return s, errCode("row_shape_or_rollback")
		}
	}
	want := s.Counts
	switch f.Stage {
	case 0:
		if s.RowActive || f.Method != 0 || f.Row == nil || f.Fact != nil || f.Callback != nil || f.Row.Done || f.Row.CandidateInvoked || f.Row.Got != (core.Got{}) {
			return s, errCode("before_row")
		}
		want.Row[0]++
		s.RowActive = true
		s.InvocationEnded = false
		s.MethodPanic = false
	case 5:
		if !s.RowActive || s.InvocationActive || s.InvocationEnded || s.MethodActive || s.CallbackActive || f.Method != 0 || f.Row == nil || f.Row.Done || f.Row.CandidateInvoked || f.Fact != nil || f.Callback != nil {
			return s, errCode("before_candidate")
		}
		catchDispatch(&want.Row)
		want.Candidate[0]++
		s.InvocationActive = true
	case 1:
		if !s.RowActive || !s.InvocationActive || s.MethodActive || s.CallbackActive || !allowed(q, j, f.Method) || f.Row != nil || f.Fact != nil || f.Callback != nil || s.MethodPanic {
			return s, errCode("before_method")
		}
		catchDispatch(&want.Candidate)
		want.Original[0]++
		s.MethodActive = true
		s.Method = f.Method
	case 7:
		if !s.InvocationActive || s.CallbackActive || q != 0 || f.Method != 0 || f.Row == nil || f.Callback == nil || f.Fact != nil || f.Row.Done || !f.Row.CandidateInvoked || f.Row.Got.NormalKnown || f.Row.Got.Error.Available || f.Row.Got.FailingLineKnown || f.Row.Got.TerminalKnown || j == 0 && (!s.MethodActive || s.Method != 1) || j != 0 && s.MethodActive {
			return s, errCode("before_callback")
		}
		catchDispatch(&want.Candidate)
		if s.MethodActive {
			catchDispatch(&want.Original)
		}
		want.Callback[0]++
		if f.Row.Got.CallbackN != s.Current.Got.CallbackN+1 || *f.Callback != f.Row.Got.Callbacks[f.Row.Got.CallbackN-1] || f.Callback.ContinueKnown {
			return s, errCode("callback_entry")
		}
		s.CallbackActive = true
	case 8:
		if !s.CallbackActive || f.Method != 0 || f.Row == nil || f.Callback == nil || f.Fact != nil || f.Row.Done || f.Row.Got.NormalKnown || f.Row.Got.Error.Available || f.Row.Got.FailingLineKnown || f.Row.Got.TerminalKnown || f.Row.Got.CallbackN != s.Current.Got.CallbackN || !f.Callback.ContinueKnown || *f.Callback != f.Row.Got.Callbacks[f.Row.Got.CallbackN-1] {
			return s, errCode("after_callback")
		}
		catchDispatch(&want.Callback)
		want.Callback[2]++
		s.CallbackActive = false
	case 2:
		if !s.MethodActive || s.CallbackActive || f.Method != s.Method || f.Row != nil || f.Callback != nil || !factShape(f.Method, f.Fact) {
			return s, errCode("after_method")
		}
		catchDispatch(&want.Original)
		if f.Fact.Panic.Present {
			want.Original[3]++
			s.MethodPanic = true
		} else {
			want.Original[2]++
		}
		s.MethodActive = false
		s.Method = 0
		if f.Fact.OwnedBoundUnsupported {
			s.Stopped = true
		}
	case 6:
		if !s.InvocationActive || s.MethodActive || s.CallbackActive || f.Method != 0 || f.Row == nil || f.Row.Done || !f.Row.CandidateInvoked || !f.Row.Got.NormalKnown || !f.Row.Got.Panic.Known || f.Row.Got.Normal == f.Row.Got.Panic.Present || f.Fact != nil || f.Callback != nil || s.MethodPanic && !f.Row.Got.Panic.Present {
			return s, errCode("after_candidate")
		}
		catchDispatch(&want.Candidate)
		if f.Row.Got.Panic.Present {
			want.Candidate[3]++
		} else {
			want.Candidate[2]++
		}
		s.InvocationActive = false
		s.InvocationEnded = true
		if f.Row.Got.Normal {
			available := j != 0
			if q == 1 {
				available = j != 0
			}
			if f.Row.Got.Error.Available != available || available && !f.Row.Got.Error.Known {
				return s, errCode("native_absent_error_channel")
			}
		}
	case 3:
		if !s.RowActive || !s.InvocationEnded || s.MethodActive || s.CallbackActive || f.Method != 0 || f.Row == nil || !f.Row.Done || f.Fact != nil || f.Callback != nil {
			return s, errCode("after_row")
		}
		old := s.Current
		old.Done = true
		if !reflect.DeepEqual(old, *f.Row) {
			return s, errCode("after_row_snapshot")
		}
		want.Row[2]++
		s.Rows[s.NextDispatch] = *f.Row
		s.Known[s.NextDispatch] = true
		s.NextDispatch++
		s.RowActive = false
		if f.Row.Got.Panic.Present {
			s.Stopped = true
		}
	default:
		return s, errCode("stage_unknown")
	}
	if f.Counts != want {
		return s, errCode("counter_transition")
	}
	s.Counts = f.Counts
	s.Sequence = f.Sequence
	if f.Row != nil {
		s.Current = *f.Row
	}
	return s, nil
}
func consumeProtocol(in io.Reader, out io.Writer, j *DurableJournal, s *ProtocolState, plan string, first func() error) error {
	b := bufio.NewReaderSize(in, 65537)
	for {
		raw, e := readLine(b)
		if e != nil {
			if e == io.EOF && s.Final != nil {
				return nil
			}
			return errCode("protocol_partial_or_eof")
		}
		var f core.Frame
		if e = strictCanonical(raw, &f); e != nil {
			return e
		}
		if f.Stage != 4 && len(raw) > 4096 {
			return errCode("ordinary_frame_cap")
		}
		next, e := s.Next(f, plan)
		if e != nil {
			return e
		}
		if s.Sequence == 0 {
			if first == nil {
				return errCode("child_reservation_missing")
			}
			if e = first(); e != nil {
				return e
			}
		}
		cursor, e := j.Append(raw)
		if e != nil {
			return e
		}
		counts := f.Counts
		s.LastDurableCounts = &counts
		next.LastDurableCounts = &counts
		next.Previous = cursor.LastSHA256
		ack, _ := json.Marshal(core.Ack{Version: 2, Sequence: f.Sequence, SHA256: cursor.LastSHA256})
		ack = append(ack, '\n')
		if len(ack) > 256 {
			return errCode("ack_cap")
		}
		n, e := out.Write(ack)
		if e != nil || n != len(ack) {
			return errCode("ack_write_uncertain")
		}
		next.FramesACKWritten = s.FramesACKWritten + 1
		*s = next
	}
}
func exactFinalResult(path string, s ProtocolState) error {
	if s.Final == nil {
		return errCode("final_missing")
	}
	expected, e := json.Marshal(s.Final)
	if e != nil || len(expected) > 60000 {
		return errCode("final_size")
	}
	got, e := readOwnedOutput(path, 65536)
	if e != nil {
		return e
	}
	if !bytes.Equal(expected, got) {
		return errCode("final_file_diff")
	}
	return nil
}
