// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
package compare

import (
	"bytes"
	"encoding/json"
	core "riido.local/next60gjsonsavedcomparison/twoliteral"
)

func fixedPins(c Config) error {
	if c.Schema != "riido-gjson-two-saved-comparison-config-v1" || c.CollectorExitCode == nil || *c.CollectorExitCode != 0 || !hashShape(c.PlanSHA256) || !hashShape(c.WorkerSHA256) || !hashShape(c.ControllerSHA256) {
		return ErrCompletion
	}
	refs := []struct {
		p Pin
		b int64
		h string
	}{{c.Fixtures, 6142, "bbaf3ad1856a66809ec51818a347b3bb46734d68a7e0bee2a36591879e7ef560"}, {c.Wants, 5714, "6cef54345a729233d4de7591b513227b0bcc8fe463d471e3ffd0e12215652236"}, {c.Captions, 2935, "a0cc56b4058d3b3d24325abbf863bd0ac3c26a578305cacf0b9fa962c24d8c5f"}, {c.RootFreeze, 6889, "87285c2c5efe2f5de7539c96934e36629a35cdcb432f45b11361462c09dd40c2"}, {c.Correction, 8510, "b5144e78100a9c4b235812c25bb13dc880fcb8ed4f4bbc2ccf85942ffbd8cf43"}, {c.Family, 3089, "1677d33aeb32c18c809e9ea717f375d5086d5b371e68f09f48fd24ae2f254b7f"}}
	for _, v := range refs {
		if v.p.Bytes != v.b || v.p.SHA256 != v.h {
			return ErrPin
		}
	}
	for _, v := range refs {
		if _, e := readPin(v.p, 128<<10, false); e != nil {
			return e
		}
	}
	all := []Pin{c.Outside, c.Child, c.Fixtures, c.Wants, c.Captions, c.RootFreeze, c.Correction, c.Family}
	for i, p := range all {
		for _, v := range all[:i] {
			if p.Path == v.Path {
				return ErrPin
			}
		}
	}
	return nil
}
func carriers(l, w []byte) (fs [11]core.Fixture, ws [11]core.WantRecord, e error) {
	var a, b map[string]json.RawMessage
	if json.Unmarshal(l, &a) != nil || json.Unmarshal(w, &b) != nil {
		return fs, ws, ErrShape
	}
	var x []core.Fixture
	var y []core.WantRecord
	if json.Unmarshal(a["fixtures"], &x) != nil || json.Unmarshal(b["records"], &y) != nil || len(x) != 11 || len(y) != 11 {
		return fs, ws, ErrShape
	}
	for i, v := range x {
		q, f := 0, i
		if i >= 5 {
			q = 1
			f = i - 5
		}
		if v.Ordinal != i || v.Request != q || v.Index != f || y[i].Ordinal != i || v.RequestID != y[i].RequestID || v.ID != y[i].FixtureID {
			return fs, ws, ErrShape
		}
		fs[i] = v
		ws[i] = y[i]
	}
	return
}
func equalRaw(a, b []byte) bool {
	var x, y bytes.Buffer
	return json.Compact(&x, a) == nil && json.Compact(&y, b) == nil && bytes.Equal(x.Bytes(), y.Bytes())
}
func completion(o Outside, r core.Result, rawChild []byte, c Config) error {
	if o.Schema != "riido-gjson-two-native-outside-result-v1" || o.State != "complete_observation_transport" || o.Failure != "" || o.PlanSHA256 != c.PlanSHA256 || o.WorkerSHA256 != c.WorkerSHA256 || o.ControllerSHA256 != c.ControllerSHA256 || !hashShape(o.ConfigSHA256) || o.Process.StartAttempts != 1 || !o.Process.Started || o.Process.WaitAttempts != 1 || o.Process.Reaped == nil || !*o.Process.Reaped || o.Process.ExitCode == nil || *o.Process.ExitCode != 0 || o.Process.DeadlineExceeded || o.Process.StderrOverflow || o.Process.RSSCapExceeded == nil || *o.Process.RSSCapExceeded || o.Process.DarwinTimeMaxRSSBytes == nil || *o.Process.DarwinTimeMaxRSSBytes < 0 || *o.Process.DarwinTimeMaxRSSBytes > 256<<20 || o.Process.WholeChildWallNanoseconds < 0 {
		return ErrCompletion
	}
	if !o.FinalACKWritten || !o.FinalChildFileIdentity || !o.CompleteCountersKnown || o.CompleteRows == nil || *o.CompleteRows != 33 || o.TerminalRowsACKWritten != 33 || o.RemainingCallsUnknown || !o.NoAutomaticRetry || o.AggregateCap != 2<<20 || o.OutputBytesRetained == nil || *o.OutputBytesRetained < 0 || *o.OutputBytesRetained > 2<<20 || o.InitializationCalls != nil || o.NestedOriginalCalls != nil || o.Authority != nil || o.Process.InitCalls != nil || o.Process.NestedOriginalCalls != nil || o.TruthAssigned != 0 || o.LabelsAssigned != 0 || o.FitOrModelCalls != 0 {
		return ErrCompletion
	}
	if r.Schema != "riido-gjson-two-native-result-v1" || r.State != "complete" || r.Failure != "" || r.Completed != 33 || r.PlanSHA256 != c.PlanSHA256 || r.Labels != 0 || r.Qualified != 0 || r.NewParents != 0 || r.TrainingAuthorized || r.OriginalInitialization != nil || r.OriginalNested != nil || r.Counts.Row != [4]int{33, 33, 33, 0} || r.Counts.Candidate != [4]int{33, 33, 33, 0} || !normalTuple(r.Counts.Original, 61) || !normalTuple(r.Counts.Callback, 33) {
		return ErrCompletion
	}
	if o.FramesACKWritten != 2+4*33+2*r.Counts.Original[0]+2*r.Counts.Callback[0] || o.FramesACKWritten > 384 || o.FramesACKWritten != o.DurablePrefix.FramesSynced || r.LastACKSequence+1 != uint32(o.FramesACKWritten) || !hashShape(r.LastACKSHA256) || o.LastDurableCounts == nil || o.LastACKWrittenCounts == nil || *o.LastDurableCounts != r.Counts || *o.LastACKWrittenCounts != r.Counts || o.DurablePrefix.BytesSynced < 1 || o.DurablePrefix.BytesSynced > 1634304 {
		return ErrCompletion
	}
	f := core.Frame{Version: 2, Sequence: r.LastACKSequence + 1, Previous: r.LastACKSHA256, Stage: 4, Dispatch: -1, Request: -1, Fixture: -1, Display: -1, Internal: -1, Counts: r.Counts, Result: rawChild}
	b, e := json.Marshal(f)
	if e != nil || len(b)+1 > 65536 || o.DurablePrefix.BytesSynced < int64(len(b)+1) || Digest(append(b, '\n')) != o.DurablePrefix.LastSHA256 {
		return ErrCompletion
	}
	if len(o.Artifacts) != 5 {
		return ErrCompletion
	}
	names := [5]string{"reservation.json", "journal.jsonl", "stderr.txt", "worker-attempt/reservation.json", "worker-attempt/results.json"}
	seen := [5]bool{}
	caps := [5]int64{16384, 1634304, 65536, 16384, 65536}
	sum := int64(0)
	for _, v := range o.Artifacts {
		at := -1
		for i, n := range names {
			if v.Name == n {
				at = i
			}
		}
		if at < 0 || seen[at] || v.Bytes < 0 || v.Bytes > caps[at] || !hashShape(v.SHA256) {
			return ErrCompletion
		}
		seen[at] = true
		sum += v.Bytes
		if at == 1 && v.Bytes != o.DurablePrefix.BytesSynced || at == 4 && (v.Bytes != c.Child.Bytes || v.SHA256 != c.Child.SHA256) {
			return ErrCompletion
		}
	}
	if sum+int64(c.Outside.Bytes) != *o.OutputBytesRetained {
		return ErrCompletion
	}
	return nil
}
func exactChild(b []byte, r core.Result) bool {
	v, e := json.Marshal(r)
	return e == nil && len(b) <= 60000 && bytes.Equal(b, v)
}
func Run(c Config) (Report, error) {
	if e := fixedPins(c); e != nil {
		return Report{}, e
	}
	outside, e := readPin(c.Outside, 16384, true)
	if e != nil {
		return Report{}, e
	}
	child, e := readPin(c.Child, 65536, true)
	if e != nil {
		return Report{}, e
	}
	var o Outside
	var r core.Result
	if Canonical(outside, &o) != nil || compact(child, &r) != nil || !exactChild(child, r) {
		return Report{}, ErrShape
	}
	if e = completion(o, r, child, c); e != nil {
		return Report{}, e
	}
	l, e := readPin(c.Fixtures, 128<<10, false)
	if e != nil {
		return Report{}, e
	}
	w, e := readPin(c.Wants, 128<<10, false)
	if e != nil {
		return Report{}, e
	}
	fs, ws, e := carriers(l, w)
	if e != nil {
		return Report{}, e
	}
	for d, row := range r.Records {
		q, fi := 0, d/3
		if fi >= 5 {
			q = 1
			fi -= 5
		}
		x := d % 3
		j := [2][3]int{{0, 2, 1}, {1, 2, 0}}[q][x]
		i := fi
		if q == 1 {
			i += 5
		}
		if row.Ordinal != d || row.Request != q || row.Fixture != fi || row.Display != x || row.Internal != j || row.RequestID != fs[i].RequestID || row.FixtureID != fs[i].ID || row.InputPointer != fs[i].InputReference.Pointer || !equalRaw(row.Input, fs[i].Input) || !equalRaw(row.Want, ws[i].Want) || !row.Done || !row.CandidateInvoked || row.Truth != nil || row.Role != nil || row.Weight != nil {
			return Report{}, ErrShape
		}
		if validateGot(row.Got) != nil {
			return Report{}, ErrShape
		}
	}
	return buildReport(r, o, c)
}
