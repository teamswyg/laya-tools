// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package compare

import (
	"encoding/json"
	"reflect"
	core "riido.local/next60gjsonsavedcomparison/twoliteral"
	"strconv"
)

type ExpectedCallback struct {
	Line int    `json:"line"`
	Raw  string `json:"raw"`
	Type uint8  `json:"type"`
}
type LineWant struct {
	Normal      bool               `json:"returned_normally"`
	Panic       bool               `json:"panic_present"`
	Callbacks   []ExpectedCallback `json:"callbacks"`
	Stop        bool               `json:"callback_stop_requested"`
	ErrorNil    bool               `json:"error_nil"`
	Identity    string             `json:"owned_error_identity"`
	FailingLine int                `json:"returned_failing_line"`
	Terminal    string             `json:"returned_terminal_reason"`
}
type AppendWant struct {
	Normal   bool   `json:"returned_normally"`
	Panic    bool   `json:"panic_present"`
	Hex      string `json:"returned_destination_hex"`
	Nil      bool   `json:"returned_destination_nil"`
	Active   string `json:"active_input_destination_after_hex"`
	Backing  string `json:"full_owned_backing_after_hex"`
	ErrorNil bool   `json:"error_nil"`
	Identity string `json:"owned_error_identity"`
}

func tally(t *Tally, s string) {
	switch s {
	case "T":
		t.T++
	case "F":
		t.F++
	case "U":
		t.U++
	}
}
func add(a *Tally, b Tally) { a.T += b.T; a.F += b.F; a.U += b.U }
func overall(t Tally) string {
	if t.F > 0 {
		return "F"
	}
	if t.U > 0 {
		return "U"
	}
	return "T"
}
func (r *RowResult) predicate(key string, want, got any, known bool, reason string) {
	w, _ := json.Marshal(want)
	p := Predicate{Key: key, Want: w, Got: json.RawMessage("null"), State: "U", Reason: reason}
	if known {
		b, _ := json.Marshal(got)
		p.Got = b
		p.State = "F"
		p.Reason = "known_mismatch"
		if reflect.DeepEqual(want, got) {
			p.State = "T"
			p.Reason = "known_match"
		}
	}
	r.Predicates[r.PredicateCount] = p
	r.PredicateCount++
	tally(&r.Counts, p.State)
}
func rowCompare(row core.Row) (r RowResult, e error) {
	r = RowResult{Ordinal: row.Ordinal, Request: row.Request, Fixture: row.Fixture, Display: row.Display, Internal: row.Internal, RequestID: row.RequestID, FixtureID: row.FixtureID, InputPointer: row.InputPointer, LiteralWant: append(json.RawMessage(nil), row.Want...), SavedGot: row.Got, CallbackEvidence: row.Got.Callbacks, CallbackCount: row.Got.CallbackN, MappingStopped: row.Got.MappingStopped}
	g := row.Got
	if validateGot(g) != nil {
		return r, ErrShape
	}
	observed := row.Done && row.CandidateInvoked
	var normal, panicWant bool
	var l LineWant
	var a AppendWant
	if row.Request == 0 {
		if compact(row.Want, &l) != nil || l.Callbacks == nil || len(l.Callbacks) > 3 || l.Identity != "none" && l.Identity != "invalid_line" || l.FailingLine < 0 || l.FailingLine > 4 || l.Terminal != "exhausted" && l.Terminal != "invalid_line" && l.Terminal != "callback_stop" || l.ErrorNil != (l.Identity == "none") {
			return r, ErrShape
		}
		normal, panicWant = l.Normal, l.Panic
		for _, c := range l.Callbacks {
			if c.Line < 1 || c.Line > 4 || c.Type > 5 || len(c.Raw) > 12 {
				return r, ErrShape
			}
		}
	} else if row.Request == 1 {
		if compact(row.Want, &a) != nil || !hexValue(a.Hex, 32) || !hexValue(a.Active, 1) || len(a.Backing) != 64 || !hexValue(a.Backing, 32) || a.Identity != "none" && a.Identity != "invalid_utf8" || a.ErrorNil != (a.Identity == "none") {
			return r, ErrShape
		}
		normal, panicWant = a.Normal, a.Panic
	} else {
		return r, ErrShape
	}
	r.predicate("returned_normally", normal, g.Normal, observed && g.NormalKnown, "normal_return_unavailable")
	r.predicate("panic_present", panicWant, g.Panic.Present, observed && g.Panic.Known, "panic_presence_unavailable")
	// Captured side facts remain in callback metadata. A known panic has no normal
	// candidate return: other Wanted return/mutation predicates stay U, never guessed.
	panicObserved := observed && g.Panic.Known && g.Panic.Present
	usable := observed && !panicObserved
	if row.Request == 0 {
		complete := g.CallbackListCompleteKnown && g.CallbackListComplete
		lengthKnown := usable && (complete || g.CallbackN > len(l.Callbacks))
		r.predicate("callbacks.length", len(l.Callbacks), g.CallbackN, lengthKnown, "callback_list_incomplete_or_unmapped")
		for i, w := range l.Callbacks {
			prefix := "callbacks." + strconv.Itoa(i) + "."
			present := usable && i < g.CallbackN
			missing := usable && complete && i >= g.CallbackN
			if missing {
				r.predicate(prefix+"line", w.Line, nil, true, "")
				r.predicate(prefix+"raw", w.Raw, nil, true, "")
				r.predicate(prefix+"type", w.Type, nil, true, "")
				continue
			}
			var c core.Callback
			if present {
				c = g.Callbacks[i]
			}
			r.predicate(prefix+"line", w.Line, c.Line, present && c.LineKnown, "physical_line_unavailable")
			r.predicate(prefix+"raw", w.Raw, c.Primitive.Raw, present && c.PrimitiveKnown, "callback_primitive_unavailable")
			r.predicate(prefix+"type", w.Type, c.Primitive.Type, present && c.PrimitiveKnown, "callback_primitive_unavailable")
		}
		r.predicate("callback_stop_requested", l.Stop, g.StopRequested, usable && g.StopKnown, "callback_stop_unavailable")
		r.predicate("error_nil", l.ErrorNil, !g.Error.Present, usable && g.Error.Available && g.Error.Known, "error_channel_unavailable")
		r.predicate("owned_error_identity", l.Identity, g.Error.OwnIdentity, usable && g.Error.Available && g.Error.Known && g.Error.OwnIdentityKnown, "error_identity_unavailable")
		r.predicate("returned_failing_line", l.FailingLine, g.FailingLine, usable && g.FailingLineKnown, "failing_line_channel_unavailable")
		r.predicate("returned_terminal_reason", l.Terminal, g.Terminal, usable && g.TerminalKnown, "terminal_channel_unavailable")
	} else {
		r.predicate("returned_destination_hex", a.Hex, g.ReturnedHex, usable && g.BufferKnown, "returned_buffer_unavailable")
		r.predicate("returned_destination_nil", a.Nil, g.ReturnedNil, usable && g.BufferKnown, "returned_buffer_unavailable")
		r.predicate("active_input_destination_after_hex", a.Active, g.ActiveInputAfterHex, usable && g.ActiveInputAfterKnown, "active_destination_unavailable")
		r.predicate("full_owned_backing_after_hex", a.Backing, g.BackingAfterHex, usable && g.BackingAfterKnown, "full_backing_unavailable")
		r.predicate("error_nil", a.ErrorNil, !g.Error.Present, usable && g.Error.Available && g.Error.Known, "error_channel_unavailable")
		r.predicate("owned_error_identity", a.Identity, g.Error.OwnIdentity, usable && g.Error.Available && g.Error.Known && g.Error.OwnIdentityKnown, "error_identity_unavailable")
	}
	if r.PredicateCount < 1 || r.PredicateCount > MaxPredicates {
		return r, ErrBounds
	}
	r.State = overall(r.Counts)
	return r, nil
}
func buildReport(child core.Result, o Outside, c Config) (r Report, e error) {
	r = Report{Schema: "riido-gjson-two-saved-finite-comparison-v1", State: "complete_saved_comparison", Scope: "Same frozen finite Want for each display candidate; saved primitive facts only, no original reexecution, no qualification or label adoption.", OutsideSHA256: c.Outside.SHA256, ChildSHA256: c.Child.SHA256, PlanSHA256: c.PlanSHA256, RootFreezeSHA256: c.RootFreeze.SHA256, FixturesSHA256: c.Fixtures.SHA256, WantsSHA256: c.Wants.SHA256, CaptionsSHA256: c.Captions.SHA256, CorrectionSHA256: c.Correction.SHA256, FamilySHA256: c.Family.SHA256, WorkerSHA256: c.WorkerSHA256, ControllerSHA256: c.ControllerSHA256, ExplicitSavedCounters: child.Counts, FramesACKWritten: o.FramesACKWritten}
	for q := 0; q < 2; q++ {
		for x := 0; x < 3; x++ {
			j := [2][3]int{{0, 2, 1}, {1, 2, 0}}[q][x]
			n := 5
			if q == 1 {
				n = 6
			}
			r.Candidates[q][x] = Candidate{Request: q, Display: x, Internal: j, Rows: n}
		}
	}
	for i, row := range child.Records {
		v, e := rowCompare(row)
		if e != nil {
			return r, e
		}
		r.Rows[i] = v
		tally(&r.RowCounts, v.State)
		add(&r.PredicateCounts, v.Counts)
		s := &r.Candidates[row.Request][row.Display]
		s.Vector[row.Fixture] = v.State
		tally(&s.Counts, v.State)
		add(&s.PredicateCounts, v.Counts)
		if v.State == "F" {
			s.KnownCounterexampleFixtures[s.CounterexampleCount] = row.Fixture
			s.CounterexampleCount++
		}
		if v.State == "U" {
			s.UnknownFixtures[s.UnknownCount] = row.Fixture
			s.UnknownCount++
		}
		if v.Counts.U > 0 {
			s.UnknownPredicateFixtures[s.UnknownPredicateFixtureCount] = row.Fixture
			s.UnknownPredicateFixtureCount++
		}
	}
	for q := 0; q < 2; q++ {
		for x := 0; x < 3; x++ {
			v := &r.Candidates[q][x]
			if v.Counts.F > 0 {
				v.Proposal = "known_counterexample_candidate"
			} else if v.Counts.U > 0 {
				v.Proposal = "unknown_only_excluded_candidate"
			} else if v.Counts.T == v.Rows && v.PredicateCounts.U == 0 {
				v.Proposal = "all_required_known_true_candidate"
			} else {
				return r, ErrShape
			}
		}
	}
	return r, nil
}
