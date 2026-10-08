package acceptedfullbridge

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// Exact public reviewpacket wire shapes. They are not historical inventory
// acceptance shapes, and their contents remain unauthenticated declarations.
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
type receiptIssue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
type readResult struct {
	Version  string         `json:"version"`
	State    string         `json:"state"`
	StartSHA string         `json:"start_receipt_sha256"`
	Expected string         `json:"expected_sha256"`
	Actual   string         `json:"actual_sha256,omitempty"`
	Bytes    int64          `json:"bytes_read"`
	Attempts int            `json:"input_open_attempts"`
	ReadUTC  string         `json:"read_completed_utc,omitempty"`
	UTC      string         `json:"decided_utc"`
	Errors   []receiptIssue `json:"errors,omitempty"`
}

func utc(s string) (time.Time, bool) {
	t, e := time.Parse(time.RFC3339Nano, s)
	return t, e == nil && strings.HasSuffix(s, "Z") && t.Year() > 1970
}
func reviewEvidence(r sourcecohort.Review, in Inputs) error {
	if r.CheckBinding != in.CheckBinding.File || r.ReadStart != in.ReadStart.File || r.ReadResult != in.ReadResult.File || r.ReadResult.Path != r.ReadStart.Path+".result" || !validFile(r.SourceSchema) {
		return Error("review_evidence_pin")
	}
	var start readStart
	var result readResult
	if err := closed(in.ReadStart.Bytes, &start); err != nil {
		return err
	}
	if err := closed(in.ReadResult.Bytes, &result); err != nil {
		return err
	}
	if start.Version != "riido-reviewpacket/v1" || start.State != "started" || start.Expected != r.Source.SHA256 || start.Max < r.Source.Bytes || start.Max > 16384 || start.Limit != 1 || start.Actor != r.ReviewerID || start.Policy != "regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries" || result.Version != start.Version || result.State != "verified" || result.StartSHA != in.ReadStart.File.SHA256 || result.Expected != r.Source.SHA256 || result.Actual != r.Source.SHA256 || result.Bytes != r.Source.Bytes || result.Attempts != 1 || len(result.Errors) != 0 {
		return Error("review_read_lineage")
	}
	st, sok := utc(start.UTC)
	read, rok := utc(result.ReadUTC)
	decided, dok := utc(result.UTC)
	reviewed, vok := utc(r.ReviewedUTC)
	if !sok || !rok || !dok || !vok || read.Before(st) || decided.Before(read) || reviewed.Before(decided) {
		return Error("review_chronology")
	}
	var binding sourcecohort.CheckBinding
	var report sourcecohort.CheckReport
	if err := closed(in.CheckBinding.Bytes, &binding); err != nil {
		return err
	}
	if err := closed(in.CheckReport.Bytes, &report); err != nil {
		return err
	}
	exec, eok := utc(binding.ExecutedUTC)
	if binding.Schema != "riido-sourcecohort-check-binding-v1" || binding.Source != r.Source || binding.Observation != r.Observation || binding.SourceSchema != r.SourceSchema || binding.Report != in.CheckReport.File || !eok || exec.Before(decided) || exec.After(reviewed) || !report.StructuralValid || !report.MandatoryComplete || !report.SourcePhaseScope || report.FailureCode != "" || strings.TrimSpace(report.Scope) == "" {
		return Error("review_check_lineage")
	}
	source := in.Projection.Source.Bytes
	if !utf8.Valid(source) || len(r.Evidence) == 0 || len(r.Evidence) > MaxItems {
		return Error("review_source_evidence")
	}
	whole := false
	for _, s := range r.Evidence {
		if s.Start < 0 || s.Start >= s.End || s.End > len(source) || s.Start < len(source) && !utf8.RuneStart(source[s.Start]) || s.End < len(source) && !utf8.RuneStart(source[s.End]) {
			return Error("review_source_evidence")
		}
		switch s.Kind {
		case "source_scope", "proposition", "opposition", "event_anchor", "communication_time", "absence", "not_applicable":
		default:
			return Error("review_source_evidence")
		}
		whole = whole || s.Kind == "source_scope" && s.Start == 0 && s.End == len(source)
	}
	if !whole {
		return Error("review_source_boundary")
	}
	return nil
}
