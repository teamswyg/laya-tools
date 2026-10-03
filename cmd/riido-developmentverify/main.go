// SPDX-License-Identifier: Apache-2.0
// Proposed public development-data CI verifier. Root reviews and stages this
// source before execution. It never loads a model or runs an original worker.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
	"github.com/teamswyg/laya-tools/pkg/shortclaimdata"
)

const dataBytes = 41428
const dataSHA = "9cdfb758f03adc34a9fb5e00e3c1525921df26d0912f6b554e4ed4c890fd10a2"
const metadataBytes = 2458
const metadataSHA = "01fe3dc513830427cad4020714a1eab25f9c5a025b4f74d5beaabf2b2ee94593"

type failure string

func (e failure) Error() string { return string(e) }

type rowSpec struct {
	SHA, ID                             string
	Group, Inputs, Candidates, Positive int
}
type totals struct {
	Rows, Labels, Positive, Negative, Inputs, Selected, PositiveIndex1 int
	GroupRows                                                          [6]int
}
type summary struct {
	Schema                  string         `json:"schema"`
	State                   string         `json:"state"`
	Failure                 string         `json:"failure_code"`
	DataSHA                 string         `json:"data_sha256"`
	MetadataSHA             string         `json:"metadata_sha256"`
	Attempts                int            `json:"Reader_dispatches"`
	Returns                 int            `json:"Reader_returns"`
	Matches                 int            `json:"Reader_value_matches"`
	ProjectionDispatches    int            `json:"ValidateInput_projection_dispatches"`
	ProjectionReturns       int            `json:"ValidateInput_projection_returns"`
	Rows                    int            `json:"requests"`
	Labels                  int            `json:"candidate_labels"`
	Positive                int            `json:"positive_labels"`
	Negative                int            `json:"negative_labels"`
	Inputs                  int            `json:"frozen_input_fixtures"`
	Selected                int            `json:"selected_candidate_observations"`
	Full                    int            `json:"original_evidence_observations"`
	PositiveIndex1          int            `json:"positive_selected_index1_requests"`
	Groups                  [6]int         `json:"source_group_ids"`
	GroupRows               [6]int         `json:"rows_per_source_group"`
	RetainedUnknown         int            `json:"retained_unknown_predicates"`
	UnknownMetadataVerified bool           `json:"retained_unknown_metadata_verified"`
	FiniteGoldenVerified    bool           `json:"finite_predicate_golden_verified"`
	Finite                  *finiteSummary `json:"finite_predicate_integrity,omitempty"`
	OriginalCalls           int            `json:"original_candidate_calls"`
	ModelCalls              int            `json:"model_calls"`
	Features                int            `json:"Features_calls"`
	Score                   int            `json:"Score_calls"`
	Project                 int            `json:"Project_calls"`
	Fit                     int            `json:"Fit_calls"`
	LabelsAssigned          bool           `json:"labels_assigned_by_verifier"`
	Qualified               bool           `json:"qualified_by_verifier"`
	Scope                   string         `json:"scope"`
}

type wireCandidate struct {
	ID     string `json:"metadata_id"`
	Text   string `json:"text"`
	Label  *bool  `json:"label"`
	Weight int    `json:"sample_weight"`
}
type wireScope struct {
	Inputs       int    `json:"input_count"`
	Observations int    `json:"candidate_observations"`
	Description  string `json:"description"`
	All          bool   `json:"all_input_guarantee"`
	Unseen       bool   `json:"unseen_source_guarantee"`
}
type wireRow struct {
	Schema       string          `json:"schema"`
	ID           string          `json:"stable_id"`
	Request      string          `json:"request"`
	Candidates   []wireCandidate `json:"candidates"`
	Role         string          `json:"role"`
	Group        int             `json:"whole_group"`
	Family       string          `json:"source_family"`
	Revision     string          `json:"source_revision"`
	Scope        wireScope       `json:"finite_scope"`
	TextRevision string          `json:"text_revision"`
	Policy       string          `json:"feature_policy"`
}

func digest(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func decode(raw []byte, dst any) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(dst) != nil {
		return false
	}
	return errors.Is(d.Decode(new(any)), io.EOF)
}

// The reader enforces presence, duplicate keys, exact schema, candidate limits,
// Unicode and unit weights. The independently decoded view checks every public
// accessor against the frozen raw row; row SHA binds all eleven wire fields.
func compareProjection(w wireRow, ex shortclaimdata.Example, want rowSpec, out *summary) (totals, error) {
	var t totals
	if !ex.Valid() || w.ID != want.ID || w.Group != want.Group || w.Scope.Inputs != want.Inputs || len(w.Candidates) != want.Candidates || want.Positive < 0 || want.Positive >= want.Candidates {
		return t, failure("row_correspondence")
	}
	m, p, s := ex.Metadata(), ex.Input().Prepared(), ex.Supervision()
	var candidates [shortclaim.MaxCandidates]shortclaim.Candidate
	if len(w.Candidates) > len(candidates) {
		return t, failure("row_correspondence")
	}
	for i, c := range w.Candidates {
		candidates[i] = shortclaim.Candidate{ID: c.ID, Text: c.Text}
	}
	out.ProjectionDispatches++
	validated, e := shortclaim.ValidateInput(shortclaim.Input{Schema: shortclaim.Schema, Request: w.Request, Candidates: candidates[:len(w.Candidates)], Provenance: shortclaimdata.InputProvenance})
	out.ProjectionReturns++
	if e != nil || p != validated.Prepared() {
		return t, failure("reader_prepared")
	}
	wantMetadata := shortclaimdata.Metadata{
		StableID: w.ID, SourceFamily: w.Family, SourceRevision: w.Revision, TextRevision: w.TextRevision,
		Role: w.Role, FeaturePolicy: w.Policy, WholeGroup: w.Group,
		FiniteScope: shortclaimdata.FiniteScope{InputCount: w.Scope.Inputs, CandidateObservations: w.Scope.Observations, Description: w.Scope.Description, AllInputGuarantee: w.Scope.All, UnseenSourceGuarantee: w.Scope.Unseen},
	}
	if w.Schema != shortclaimdata.Schema || w.Role != shortclaimdata.Role || w.Policy != shortclaimdata.FeaturePolicy || w.Scope.All || w.Scope.Unseen || m != wantMetadata || p.Schema != shortclaim.Schema || p.Provenance != shortclaimdata.InputProvenance || p.Request != w.Request || p.NormalizedRequest == "" || p.Count != len(w.Candidates) || s.Count != p.Count || w.Scope.Observations != want.Inputs*want.Candidates {
		return t, failure("reader_metadata")
	}
	for i, c := range w.Candidates {
		if c.Label == nil || c.Weight != 1 || p.Candidates[i].ID != c.ID || p.Candidates[i].Text != c.Text || p.Candidates[i].NormalizedText == "" || s.Labels[i] != *c.Label || s.Weights[i] != c.Weight || *c.Label != (i == want.Positive) {
			return t, failure("reader_candidate")
		}
		t.Labels++
		if *c.Label {
			t.Positive++
			if i == 1 {
				t.PositiveIndex1++
			}
		} else {
			t.Negative++
		}
	}
	for i := p.Count; i < shortclaim.MaxCandidates; i++ {
		if p.Candidates[i] != (shortclaim.PreparedCandidate{}) || s.Labels[i] || s.Weights[i] != 0 {
			return totals{}, failure("reader_unused_slot")
		}
	}
	if w.Group < 76 || w.Group > 81 {
		return totals{}, failure("reader_group")
	}
	t.Rows = 1
	t.Inputs = w.Scope.Inputs
	t.Selected = w.Scope.Observations
	t.GroupRows[w.Group-76] = 1
	return t, nil
}

type rowLoader func(io.Reader) (shortclaimdata.Example, error)

func verifyRow(raw []byte, want rowSpec, load rowLoader, out *summary) (totals, error) {
	if len(raw) < 1 || len(raw) > shortclaimdata.MaxRowBytes || digest(raw) != want.SHA {
		return totals{}, failure("row_pin")
	}
	var w wireRow
	if !decode(raw, &w) {
		return totals{}, failure("row_json")
	}
	out.Attempts++
	ex, e := load(bytes.NewReader(raw))
	out.Returns++
	if e != nil {
		return totals{}, failure("project_reader")
	}
	t, e := compareProjection(w, ex, want, out)
	if e == nil {
		out.Matches++
	}
	return t, e
}

func add(a *totals, b totals) {
	a.Rows += b.Rows
	a.Labels += b.Labels
	a.Positive += b.Positive
	a.Negative += b.Negative
	a.Inputs += b.Inputs
	a.Selected += b.Selected
	a.PositiveIndex1 += b.PositiveIndex1
	for i := range a.GroupRows {
		a.GroupRows[i] += b.GroupRows[i]
	}
}
func streamRows(r io.Reader, wants []rowSpec, load rowLoader, out *summary) (totals, error) {
	var t totals
	b := bufio.NewReaderSize(r, shortclaimdata.MaxRowBytes+1)
	for i, want := range wants {
		line, e := b.ReadSlice('\n')
		if e != nil || len(line) < 2 || len(line)-1 > shortclaimdata.MaxRowBytes {
			return t, failure("data_row_bounds")
		}
		z, e := verifyRow(line[:len(line)-1], want, load, out)
		if e != nil {
			return t, e
		}
		add(&t, z)
		if t.Rows != i+1 {
			return t, failure("data_row_count")
		}
	}
	if _, e := b.ReadByte(); !errors.Is(e, io.EOF) {
		return t, failure("data_extra_row")
	}
	return t, nil
}

// These small regular-file reads never print a path or caller text. File pins
// are checked before Reader dispatch; a second digest and stat reject mutation.
func openRegular(path string, size int64) (*os.File, os.FileInfo, error) {
	if path == "" {
		return nil, nil, failure("input_file")
	}
	p, e := filepath.Abs(path)
	if e != nil {
		return nil, nil, failure("input_file")
	}
	st, e := os.Lstat(p)
	if e != nil || !st.Mode().IsRegular() || st.Size() != size {
		return nil, nil, failure("input_file")
	}
	f, e := os.Open(p)
	if e != nil {
		return nil, nil, failure("input_file")
	}
	fs, e := f.Stat()
	if e != nil || !os.SameFile(st, fs) || !fs.Mode().IsRegular() || fs.Size() != size {
		f.Close()
		return nil, nil, failure("input_file")
	}
	return f, fs, nil
}
func hashFile(f *os.File, size int64, sha string) error {
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, size+1))
	if e != nil || n != size || hex.EncodeToString(h.Sum(nil)) != sha {
		return failure("input_pin")
	}
	return nil
}
func stable(f *os.File, before os.FileInfo) bool {
	after, e := f.Stat()
	return e == nil && os.SameFile(before, after) && after.Mode().IsRegular() && after.Size() == before.Size() && after.ModTime() == before.ModTime()
}
func readMetadata(path string) ([]byte, error) {
	f, st, e := openRegular(path, metadataBytes)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, metadataBytes+1))
	if e != nil || len(raw) != metadataBytes || digest(raw) != metadataSHA || !stable(f, st) {
		return nil, failure("metadata_pin")
	}
	return raw, nil
}

func verifyFiles(data, metadata string, out *summary) error {
	raw, e := readMetadata(metadata)
	if e != nil {
		return e
	}
	if e = verifyMetadata(raw); e != nil {
		return e
	}
	out.MetadataSHA = metadataSHA
	out.RetainedUnknown = 8
	out.UnknownMetadataVerified = true
	out.Full = 434
	f, st, e := openRegular(data, dataBytes)
	if e != nil {
		return e
	}
	defer f.Close()
	if e = hashFile(f, dataBytes, dataSHA); e != nil {
		return e
	}
	if _, e = f.Seek(0, io.SeekStart); e != nil {
		return failure("data_seek")
	}
	h := sha256.New()
	t, e := streamRows(io.TeeReader(io.LimitReader(f, dataBytes+1), h), expectedRows[:], shortclaimdata.LoadDevelopmentRow, out)
	if e != nil {
		return e
	}
	if hex.EncodeToString(h.Sum(nil)) != dataSHA || !stable(f, st) {
		return failure("data_changed")
	}
	if t != (totals{Rows: 30, Labels: 88, Positive: 30, Negative: 58, Inputs: 146, Selected: 430, PositiveIndex1: 27, GroupRows: [6]int{4, 13, 6, 3, 1, 3}}) || out.Attempts != 30 || out.Returns != 30 || out.Matches != 30 || out.ProjectionDispatches != 30 || out.ProjectionReturns != 30 {
		return failure("data_counts")
	}
	out.DataSHA = dataSHA
	out.Rows = t.Rows
	out.Labels = t.Labels
	out.Positive = t.Positive
	out.Negative = t.Negative
	out.Inputs = t.Inputs
	out.Selected = t.Selected
	out.PositiveIndex1 = t.PositiveIndex1
	out.GroupRows = t.GroupRows
	return nil
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "--help" {
		_, e := io.WriteString(stdout, "riido-developmentverify --data DATA30 --metadata MATERIALIZATION [--finite SAVED_PREDICATES]\nReader and saved predicate bookkeeping integrity only; no original worker, feature encoding, scoring or fit.\n")
		if e != nil {
			return 1
		}
		return 0
	}
	if (len(args) != 4 && len(args) != 6) || args[0] != "--data" || args[2] != "--metadata" || args[1] == "" || args[3] == "" || (len(args) == 6 && (args[4] != "--finite" || args[5] == "")) {
		io.WriteString(stderr, "developmentverify_arguments\n")
		return 2
	}
	out := summary{Schema: "riido-public30-reader-integrity-v1", State: "rejected", Groups: [6]int{76, 77, 78, 79, 80, 81}, Scope: "Pinned development30 Reader/raw-value and eight retained-unknown metadata agreement; optional finite check verifies saved predicate bookkeeping only. No original reexecution, new semantic, role, rights, qualification, training, model quality or resource authority."}
	e := verifyFiles(args[1], args[3], &out)
	if e == nil && len(args) == 6 {
		finite, finiteErr := verifyFinite(args[5])
		out.Finite = &finite
		out.FiniteGoldenVerified = finiteErr == nil && finite.BookkeepingVerified
		e = finiteErr
	}
	if e != nil {
		out.Failure = e.Error()
	} else {
		out.State = "verified_reader_and_retained_metadata"
		if out.FiniteGoldenVerified {
			out.State = "verified_reader_metadata_and_saved_predicates"
		}
	}
	raw, me := json.Marshal(out)
	if me != nil || len(raw)+1 > 4096 {
		io.WriteString(stderr, "developmentverify_output\n")
		return 1
	}
	if n, we := stdout.Write(append(raw, '\n')); we != nil || n != len(raw)+1 {
		io.WriteString(stderr, "developmentverify_output\n")
		return 1
	}
	if e != nil {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
