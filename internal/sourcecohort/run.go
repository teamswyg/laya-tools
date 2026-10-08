package sourcecohort

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/reviewpacket"
)

type reader struct {
	root, out      string
	max, freeze    int64
	calls          int
	used, retained int64
}

func (r *reader) capture(f File, limit int64, target any) []byte {
	require(validFile(f) && f.Bytes <= limit && limit > 0 && limit <= reviewpacket.MaximumMaxBytes, "input_pin")
	require(!filepath.IsAbs(f.Path) && filepath.Clean(f.Path) == f.Path && f.Path != "." && !strings.HasPrefix(f.Path, ".."), "input_path")
	parent, e := filepath.EvalSymlinks(filepath.Dir(filepath.Join(r.root, f.Path)))
	require(e == nil && (parent == r.root || strings.HasPrefix(parent, r.root+string(os.PathSeparator))), "input_root")
	require(r.used+4096+r.freeze+65536 <= r.max, "output_preflight")
	r.calls++
	receipt := filepath.Join(r.out, "receipts", fmt.Sprintf("%06d.start.json", r.calls))
	b, outcome, e := reviewpacket.Capture(reviewpacket.Config{InputPath: filepath.Join(r.root, f.Path), ReceiptPath: receipt, ExpectedSHA256: f.SHA256, MaxBytes: limit, Actor: "sourcecohort-adapter"})
	for _, name := range []string{receipt, receipt + ".result"} {
		info, se := os.Stat(name)
		if se == nil {
			r.used += info.Size()
		}
	}
	require(e == nil && outcome.StartDurable && outcome.ResultDurable, "capture_failed")
	require(int64(len(b)) == f.Bytes && Hash(b) == f.SHA256, "input_bytes")
	if target != nil {
		require(Decode(b, target) == nil, "input_json")
	}
	return b
}
func diskBytes(dir string) int64 {
	var n int64
	e := filepath.WalkDir(dir, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			i, e := d.Info()
			if e != nil {
				return e
			}
			n += i.Size()
		}
		return nil
	})
	require(e == nil, "output_accounting")
	return n
}
func marshal(v any) []byte {
	b, e := json.Marshal(v)
	require(e == nil, "output_json")
	return append(b, '\n')
}
func publish(dir, name string, data []byte) {
	tmp := filepath.Join(dir, name+".pending")
	f, e := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	require(e == nil, "output_create")
	n, we := f.Write(data)
	se := f.Sync()
	ce := f.Close()
	require(we == nil && se == nil && ce == nil && n == len(data), "output_write")
	require(os.Link(tmp, filepath.Join(dir, name)) == nil && os.Remove(tmp) == nil, "output_publish")
	d, e := os.Open(dir)
	require(e == nil, "output_directory")
	se = d.Sync()
	ce = d.Close()
	require(se == nil && ce == nil, "output_sync")
}

// Run creates a fresh private output directory. Only its returned aggregate is
// suitable for stdout. All referenced content is captured with new receipts.
func Run(root string, planFile File, out string, allocationOnly bool) (s Summary, err error) {
	s = initial()
	created := false
	err = guarded(func() {
		var e error
		root, e = filepath.Abs(root)
		require(e == nil, "root_path")
		root, e = filepath.EvalSymlinks(root)
		require(e == nil, "root_path")
		require(os.Mkdir(out, 0700) == nil, "new_output_required")
		created = true
		require(os.Mkdir(filepath.Join(out, "receipts"), 0700) == nil, "output_directory")
		r := reader{root: root, out: out, max: 64 << 20}
		var p Plan
		r.capture(planFile, 1<<20, &p)
		require(p.Schema == "riido-sourcecohort-plan-v1" && p.InputMaxBytes > 0 && p.InputMaxBytes <= 8<<20 && p.SourceMaxBytes > 0 && p.SourceMaxBytes <= 16384 && p.FreezeMaxBytes > 0 && p.FreezeMaxBytes <= 64<<20 && p.OutputMaxBytes >= 1<<20 && p.OutputMaxBytes <= 64<<20 && p.FreezeMaxBytes+65536+4096 <= p.OutputMaxBytes, "plan")
		r.max, r.freeze = p.OutputMaxBytes, p.FreezeMaxBytes
		var recipe Recipe
		var reg Registry
		r.capture(p.SplitConfig, p.InputMaxBytes, &recipe)
		r.capture(p.Registry, p.InputMaxBytes, &reg)
		rows := validateRegistry(reg, recipe, &s)
		if allocationOnly {
			require(p.Creation == nil && p.Reviews == nil && p.Allocation == nil, "allocation_future_refs")
			for _, row := range rows {
				require(row.Source == nil, "allocation_source_must_be_null")
			}
			b := marshal(Allocation{"riido-sourcecohort-allocation-v1", p.Registry, p.SplitConfig, reg.CohortID, rows})
			require(int64(len(b)) <= p.FreezeMaxBytes && diskBytes(out)+int64(len(b))+65536 <= r.max, "allocation_size")
			publish(out, "ALLOCATION.private.json", b)
			s.State = "allocation_ready"
			s.Unresolved = 0
			return
		}
		require(p.Creation != nil && p.Reviews != nil && p.SourceSchema != nil && p.Allocation != nil && identifier(p.AuthorID) && identifier(p.CheckerID) && p.AuthorID != p.CheckerID, "freeze_binding")
		var allocation Allocation
		r.capture(*p.Allocation, p.InputMaxBytes, &allocation)
		require(allocation.Schema == "riido-sourcecohort-allocation-v1" && allocation.CohortID == reg.CohortID && allocation.SplitConfig == p.SplitConfig && len(allocation.Rows) == len(rows), "allocation_binding")
		var prospective Registry
		r.capture(allocation.Registry, p.InputMaxBytes, &prospective)
		priorSummary := initial()
		priorRows := validateRegistry(prospective, recipe, &priorSummary)
		require(prospective.CohortID == reg.CohortID, "allocation_registry_binding")
		for _, row := range priorRows {
			require(row.Source == nil, "allocation_source_must_be_null")
		}
		sort.Slice(allocation.Rows, func(i, j int) bool { return allocation.Rows[i].ID < allocation.Rows[j].ID })
		for i, a := range allocation.Rows {
			b := rows[i]
			prior := priorRows[i]
			require(a.Source == nil && a.ID == prior.ID && a.Stratum == prior.Stratum && a.Split == prior.Split && a.DEVStyle == prior.DEVStyle && sameIDs(a.Dependencies, prior.Dependencies), "allocation_registry_binding")
			require(a.ID == b.ID && a.Stratum == b.Stratum && a.Split == b.Split && a.DEVStyle == b.DEVStyle && sameIDs(a.Dependencies, b.Dependencies), "allocation_changed")
		}
		var creation CreationManifest
		var reviews Reviews
		r.capture(*p.Creation, p.InputMaxBytes, &creation)
		r.capture(*p.Reviews, p.InputMaxBytes, &reviews)
		s.Reviewed = len(reviews.Rows)
		reviewHolds(reviews.Rows, &s)
		require(creation.Schema == "riido-sourcecohort-creation-v1" && reviews.Schema == "riido-sourcecohort-reviews-v1" && len(creation.Rows) == len(rows) && len(reviews.Rows) == len(rows), "whole_join_count")
		sort.Slice(creation.Rows, func(i, j int) bool { return creation.Rows[i].ID < creation.Rows[j].ID })
		sort.Slice(reviews.Rows, func(i, j int) bool { return reviews.Rows[i].ID < reviews.Rows[j].ID })
		for i, row := range rows {
			c, v := creation.Rows[i], reviews.Rows[i]
			require(c.ID == row.ID && v.ID == row.ID, "whole_join_id")
			if !c.RightsAllowed || c.CreatorKind != "ai_nonhuman" {
				if len(v.Holds) == 0 && v.Complete && v.Structural && v.SourceScope == "complete_source_situation" && v.SemanticVerdict == "declared_pass" {
					s.Held++
				}
				s.Holds[countAt(s.Holds, "rights")].Rows++
			}
		}
		require(s.Held == 0, "review_holds")
		schema := r.capture(*p.SourceSchema, p.InputMaxBytes, nil)
		require(Transport(schema) == nil, "schema_json")
		joined := make([]Joined, len(rows))
		parents := make([]int, len(rows))
		for i := range parents {
			parents[i] = i
		}
		for i, row := range rows {
			c, v := creation.Rows[i], reviews.Rows[i]
			require(row.Source != nil && *row.Source == c.Source && c.Source == v.Source && v.SourceSchema == *p.SourceSchema && c.AuthorID == p.AuthorID && v.ReviewerID == p.CheckerID, "row_binding")
			var cr CreationReceipt
			r.capture(c.Receipt, p.InputMaxBytes, &cr)
			require(cr.Schema == "riido-sourcecohort-creation-receipt-v1" && cr.SourceID == row.ID && cr.Source == c.Source && cr.AuthorID == c.AuthorID && cr.CreatorKind == c.CreatorKind && cr.RightsAllowed && cr.CreatedUTC == c.CreatedUTC && cr.ConsultedNone == (len(cr.Consulted) == 0) && len(cr.Consulted) <= 32, "creation_receipt")
			for _, consulted := range cr.Consulted {
				r.capture(consulted, p.InputMaxBytes, nil)
			}
			require(r.retained+c.Source.Bytes+v.Observation.Bytes <= p.FreezeMaxBytes, "retained_freeze_bound")
			r.retained += c.Source.Bytes + v.Observation.Bytes
			data := r.capture(c.Source, p.SourceMaxBytes, nil)
			checkSpans(data, v.Evidence)
			decided := checkReadReceipts(&r, v, c)
			observation := r.capture(v.Observation, p.InputMaxBytes, nil)
			require(Transport(observation) == nil, "observation_json")
			var binding CheckBinding
			r.capture(v.CheckBinding, p.InputMaxBytes, &binding)
			require(binding.Schema == "riido-sourcecohort-check-binding-v1" && binding.Source == c.Source && binding.Observation == v.Observation && binding.SourceSchema == *p.SourceSchema && !timeUTC(binding.ExecutedUTC).Before(decided) && !timeUTC(binding.ExecutedUTC).After(timeUTC(v.ReviewedUTC)), "check_binding")
			var check CheckReport
			r.capture(binding.Report, p.InputMaxBytes, &check)
			if !check.StructuralValid || !check.MandatoryComplete || !check.SourcePhaseScope || check.FailureCode != "" || strings.TrimSpace(check.Scope) == "" {
				s.Held++
				s.Holds[countAt(s.Holds, "structural")].Rows++
				require(false, "declared_check_hold")
			}
			deps := []string{}
			for _, list := range [][]string{row.Dependencies, c.Dependencies, cr.Dependencies, v.Dependencies} {
				checkDependencies(rows, i, list)
				deps = append(deps, list...)
			}
			sort.Strings(deps)
			last := ""
			for _, id := range deps {
				if id != last {
					union(parents, i, rowIndex(rows, id))
					last = id
				}
			}
			joined[i] = Joined{row, c, v, string(data), string(observation), cr, binding, check}
			s.Verified++
			s.Unresolved = s.Expected - s.Verified
		}
		for i := range parents {
			if find(parents, i) == i {
				s.Components++
			}
		}
		s.State = "structural_provenance_freeze_ready"
		b := marshal(Freeze{"riido-sourcecohort-freeze-v1", planFile, p, reg.CohortID, joined, s})
		require(int64(len(b)) <= p.FreezeMaxBytes && diskBytes(out)+int64(len(b))+65536 <= r.max, "freeze_size")
		publish(out, "FREEZE.private.json", b)
	})
	if err != nil {
		s.State = "blocked"
		s.Code = err.Error()
		s.Unresolved = s.Expected - s.Verified
		if s.Code != "review_holds" && s.Code != "declared_check_hold" {
			s.Holds[countAt(s.Holds, "technical")].Rows++
		}
	}
	if created {
		if e := guarded(func() { publish(out, "SUMMARY.json", marshal(s)) }); e != nil {
			err = e
			s.Code = e.Error()
			s.State = "blocked"
		}
	}
	return
}
func sameIDs(a, b []string) bool {
	a = sortedIDs(a)
	b = sortedIDs(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func find(p []int, i int) int {
	for p[i] != i {
		p[i] = p[p[i]]
		i = p[i]
	}
	return i
}
func union(p []int, a, b int) {
	a, b = find(p, a), find(p, b)
	if a > b {
		a, b = b, a
	}
	p[b] = a
}

type start struct {
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
type terminal struct {
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

func checkReadReceipts(r *reader, v Review, c CreationRow) time.Time {
	require(v.ReadResult.Path == v.ReadStart.Path+".result", "receipt_path")
	var a start
	var b terminal
	raw := r.capture(v.ReadStart, 8192, &a)
	r.capture(v.ReadResult, 8192, &b)
	require(a.Version == "riido-reviewpacket/v1" && a.State == "started" && a.Expected == c.Source.SHA256 && a.Max >= c.Source.Bytes && a.Max <= 16384 && a.Limit == 1 && (a.Actor == "" || a.Actor == v.ReviewerID) && a.Policy == "regular-file; no final symlink; nonblocking open; identity checked; bounded read; no retries" && b.Version == a.Version && b.State == "verified" && b.StartSHA == Hash(raw) && b.Expected == c.Source.SHA256 && b.Actual == c.Source.SHA256 && b.Bytes == c.Source.Bytes && b.Attempts == 1 && len(b.Errors) == 0, "receipt_binding")
	require(!timeUTC(a.UTC).Before(timeUTC(c.CreatedUTC)) && !timeUTC(b.ReadUTC).Before(timeUTC(a.UTC)) && !timeUTC(b.UTC).Before(timeUTC(b.ReadUTC)) && !timeUTC(v.ReviewedUTC).Before(timeUTC(b.UTC)), "chronology")
	return timeUTC(b.UTC)
}
