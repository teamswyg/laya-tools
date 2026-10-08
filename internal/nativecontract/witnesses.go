package nativecontract

import (
	"strings"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// These optional witness shapes are the exact public readpacket contract.
// They do not imply that native/historical private read shapes are equivalent.
type readStart struct {
	Version  string `json:"version"`
	State    string `json:"state"`
	UTC      string `json:"started_utc"`
	Expected string `json:"expected_sha256"`
	Max      int64  `json:"max_bytes"`
	Policy   string `json:"input_policy"`
	Limit    int    `json:"read_attempt_limit"`
	Actor    string `json:"actor,omitempty"`
}
type issue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
type readResult struct {
	Version  string  `json:"version"`
	State    string  `json:"state"`
	StartSHA string  `json:"start_receipt_sha256"`
	Expected string  `json:"expected_sha256"`
	Actual   string  `json:"actual_sha256,omitempty"`
	Bytes    int64   `json:"bytes_read"`
	Attempts int     `json:"input_open_attempts"`
	ReadUTC  string  `json:"read_completed_utc,omitempty"`
	UTC      string  `json:"decided_utc"`
	Errors   []issue `json:"errors,omitempty"`
}

func witnessJoin(r sourcecohort.Review, in Inputs, pins []File) error {
	w := in.Witnesses
	e := in.Expected
	if w.ReadStart.File != e.ReadStart || w.ReadResult.File != e.ReadResult || w.Check.File != e.Check || w.Report.File != e.Report || w.HeldReview.File != e.HeldReview {
		return Error("witness_pin_join")
	}
	var s readStart
	var result readResult
	var c sourcecohort.CheckBinding
	var report sourcecohort.CheckReport
	var old sourcecohort.Review
	if err := closed(w.ReadStart.Bytes, &s); err != nil {
		return err
	}
	if err := closed(w.ReadResult.Bytes, &result); err != nil {
		return err
	}
	if err := closed(w.Check.Bytes, &c); err != nil {
		return err
	}
	if err := closed(w.Report.Bytes, &report); err != nil {
		return err
	}
	if err := closed(w.HeldReview.Bytes, &old); err != nil {
		return err
	}
	if s.Version != "riido-reviewpacket/v1" || s.State != "started" || s.Expected != e.Source.SHA256 || s.Max < e.Source.Bytes || s.Max > 16384 || s.Limit != 1 || s.Actor != e.ReviewerID || s.Policy != "regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries" || result.Version != s.Version || result.State != "verified" || result.StartSHA != w.ReadStart.File.SHA256 || result.Expected != e.Source.SHA256 || result.Actual != e.Source.SHA256 || result.Bytes != e.Source.Bytes || result.Attempts != 1 || len(result.Errors) != 0 {
		return Error("witness_read_join")
	}
	st, sok := UTC(s.UTC)
	rt, rok := UTC(result.ReadUTC)
	dt, dok := UTC(result.UTC)
	reviewed, vok := UTC(r.ReviewedUTC)
	checked, cok := UTC(c.ExecutedUTC)
	if !sok || !rok || !dok || !vok || !cok || rt.Before(st) || dt.Before(rt) || checked.Before(dt) || checked.After(reviewed) {
		return Error("witness_chronology")
	}
	if c.Schema != "riido-sourcecohort-check-binding-v1" || c.Source != e.Source || c.Observation != e.SelectedObservation || c.SourceSchema != e.SourceSchema || c.Report != w.Report.File {
		return Error("witness_check_join")
	}
	if r.Structural && (!report.StructuralValid || !report.MandatoryComplete || !report.SourcePhaseScope || report.FailureCode != "" || strings.TrimSpace(report.Scope) == "") {
		return Error("witness_check_status")
	}
	if old.ID != e.ID || old.Source != e.Source || old.Observation != e.OriginalObservation || old.SourceSchema != e.SourceSchema || len(old.Holds) == 0 || old.SemanticVerdict != "declared_pass" && old.SemanticVerdict != "declared_hold" {
		return Error("witness_retained_hold")
	}
	if err := checkPins(append(pins, old.Source, old.SourceSchema, old.Observation, old.CheckBinding, old.ReadStart, old.ReadResult)); err != nil {
		return err
	}
	return nil
}
