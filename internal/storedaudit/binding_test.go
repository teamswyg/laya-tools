// SPDX-License-Identifier: Apache-2.0
package storedaudit

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These tests only bind existing public stored records or inspect JSON and
// metadata consistency boundaries. They do not execute original candidates,
// validate/normalize features, rank, train, read final data or call models.
func storedFixtures(t *testing.T) [3][]byte {
	t.Helper()
	names := [3]string{"probes-56.json", "probes-56b.json", "results-56b.json"}
	wantBytes := [3]int{49941, 43598, 266818}
	var raw [3][]byte
	for i, name := range names {
		f, err := os.Open(filepath.Join("..", "..", "experiments", "short-claim", name))
		if err != nil {
			t.Fatalf("open fixed public fixture %s: %v", name, err)
		}
		info, err := f.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() != int64(wantBytes[i]) {
			_ = f.Close()
			t.Fatalf("fixed public fixture %s has unexpected type or size", name)
		}
		raw[i], err = io.ReadAll(io.LimitReader(f, int64(wantBytes[i]+1)))
		closeErr := f.Close()
		if err != nil || closeErr != nil || len(raw[i]) != wantBytes[i] {
			t.Fatalf("bounded read of fixed public fixture %s failed", name)
		}
	}
	return raw
}

func TestBindFixedStoredRecords(t *testing.T) {
	raw := storedFixtures(t)
	dataset, err := Bind(raw[0], raw[1], raw[2])
	if err != nil {
		t.Fatalf("fixed stored-record binding: %v", err)
	}
	if dataset.Schema != Schema || dataset.Answerable != 34 || dataset.NoAnswer != 17 || dataset.Unknown != 21 || dataset.Candidates != 216 || dataset.LabeledGroups != 16 || dataset.CandidateCountHistogram != [5]int{0, 0, 12, 48, 12} {
		t.Fatal("stored preparation counts changed")
	}
	if dataset.Metadata.LegacyInputSHA256 != LegacySHA256 || dataset.Metadata.TypedInputSHA256 != TypedSHA256 || dataset.Metadata.StoredResultSHA256 != ResultSHA256 || !bytes.Equal(dataset.Metadata.StoredEvidence, raw[2]) || !reflect.DeepEqual(dataset.Metadata.StopReasons, []string{"no_roles_plan", "synthetic_single_pipeline"}) {
		t.Fatal("stored evidence or preparation limitations changed")
	}
	var legacy, typed wireProbes
	var result storedResult
	for _, item := range []struct {
		raw []byte
		dst any
	}{{raw[0], &legacy}, {raw[1], &typed}, {raw[2], &result}} {
		// This independent decode is for comparison with the exact fixture. Bind
		// performs the strict decode; tests do not derive any source truth here.
		if err := json.Unmarshal(item.raw, item.dst); err != nil {
			t.Fatalf("read stored comparison metadata: %v", err)
		}
	}
	parents := append(legacy.Parents, typed.Parents...)
	outcomes := append(result.Legacy.Outcomes, result.Typed.Outcomes...)
	unknown, labeled := 0, 0
	var membership [ParentCount]bool
	for i, row := range dataset.Rows {
		p, o := parents[i], outcomes[i]
		if row.Text.Request != p.Request || row.Text.CandidateCount != len(p.Candidates) || row.Audit.ParentID != p.ID || row.Audit.Prototype != p.Prototype || row.Audit.ContractID != p.ContractID || row.Audit.CandidateCount != len(p.Candidates) || row.Truth.State != o.State || row.Truth.Reason != o.Reason || row.Truth.AcceptableCount != len(o.Acceptable) {
			t.Fatalf("stored parent ordering or state changed at row %d", i)
		}
		for j, c := range p.Candidates {
			if row.Text.Candidates[j] != c.Text || row.Audit.Candidates[j].ID != c.ID || row.Audit.Candidates[j].SourceID != c.SourceID || row.Audit.Candidates[j].CodeSHA256 != c.CodeSHA256 || row.Audit.Candidates[j].BundleSHA256 != c.BundleSHA256 || row.Truth.Candidates[j] != o.Candidates[j] {
				t.Fatalf("stored candidate ordering or metadata changed at row %d candidate %d", i, j)
			}
		}
		for j, index := range o.Acceptable {
			if row.Truth.AcceptableIndices[j] != index {
				t.Fatalf("acceptable candidate index changed at row %d", i)
			}
		}
		for j := len(p.Candidates); j < MaxCandidates; j++ {
			if row.Text.Candidates[j] != "" || row.Audit.Candidates[j] != (CandidateAudit{}) || row.Truth.Candidates[j] != (CandidateTruth{}) {
				t.Fatalf("unused candidate slot is populated at row %d", i)
			}
		}
		if row.Truth.State == "unknown" {
			unknown++
			if row.Truth.AcceptableCount != 0 {
				t.Fatal("unknown stored record was converted into an answer")
			}
			for _, c := range row.Truth.Candidates[:row.Truth.CandidateCount] {
				if c.State != "unknown" || c.VectorsChecked != 0 || c.FailedVectors != 0 {
					t.Fatal("unknown stored record was converted into a rejected candidate")
				}
			}
		}
	}
	for i, group := range dataset.Groups {
		if !reflect.DeepEqual(group, result.Combined.Groups[i]) {
			t.Fatalf("original sparse group or membership changed at index %d", i)
		}
		groupLabeled := false
		for _, id := range group.Parents {
			index := -1
			for j, row := range dataset.Rows {
				if row.Audit.ParentID == id {
					index = j
				}
			}
			if index < 0 || membership[index] || dataset.Rows[index].Audit.GroupID != group.ID {
				t.Fatal("stored group does not preserve unique parent membership")
			}
			membership[index] = true
			groupLabeled = groupLabeled || dataset.Rows[index].Truth.State != "unknown"
		}
		if groupLabeled {
			labeled++
		}
	}
	for _, present := range membership {
		if !present {
			t.Fatal("stored grouping omitted a parent")
		}
	}
	if unknown != 21 || labeled != 16 {
		t.Fatal("unknown membership or labeled group count changed")
	}
	if !reflect.DeepEqual(dataset.Metadata.Edges, result.Combined.Edges) {
		t.Fatal("stored metadata graph edges changed")
	}
	// The source/contract/group/outcome metadata cannot become baseline inputs
	// through an accidental field addition to Text.
	typ := reflect.TypeOf(Text{})
	if typ.NumField() != 3 || typ.Field(0).Name != "Request" || typ.Field(1).Name != "CandidateCount" || typ.Field(2).Name != "Candidates" {
		t.Fatal("Text gained non-prose metadata fields")
	}
	t.Log("stored-record metadata binding: 1 successful Bind, 72 parents, 51 known including 17 no-answer, 21 unknown; source calls/rankings/models/fits: 0")
}

func TestBindRejectsModifiedPinnedInputs(t *testing.T) {
	raw := storedFixtures(t)
	for i, name := range [3]string{"legacy", "typed", "result"} {
		t.Run(name, func(t *testing.T) {
			changed := raw
			changed[i] = bytes.Clone(raw[i])
			// A same-size whitespace change is still a different exact input.
			if len(changed[i]) < 2 || changed[i][1] != '\n' {
				t.Fatal("unexpected frozen fixture layout")
			}
			changed[i][1] = ' '
			dataset, err := Bind(changed[0], changed[1], changed[2])
			if err != ErrPin || !reflect.DeepEqual(dataset, Dataset{}) {
				t.Fatal("modified pinned bytes were accepted or synthesized labels")
			}
		})
	}
	// These failures are pin checks, not newly observed source-truth failures.
	t.Log("stored-record metadata binding: 3 altered-input Bind calls, all rejected at the exact pin gate")
}

func TestStrictStoredJSONBoundaries(t *testing.T) {
	type fields struct {
		Count int    `json:"count"`
		Text  string `json:"text"`
	}
	cases := []struct {
		name, raw string
		want      error
	}{
		{"duplicate", `{"count":1,"count":2,"text":"safe"}`, ErrDuplicate},
		{"escaped_duplicate", `{"count":1,"co\u0075nt":2,"text":"safe"}`, ErrDuplicate},
		{"case_duplicate", `{"count":1,"Count":2,"text":"safe"}`, ErrDuplicate},
		{"unknown", `{"count":1,"text":"safe","unexpected":2}`, ErrUnknownField},
		{"case_key", `{"Count":1,"text":"safe"}`, ErrUnknownField},
		{"fraction", `{"count":1.5,"text":"safe"}`, ErrNumber},
		{"exponent", `{"count":1e0,"text":"safe"}`, ErrNumber},
		{"negative", `{"count":-1,"text":"safe"}`, ErrNumber},
		{"overflow", `{"count":18446744073709551616,"text":"safe"}`, ErrNumber},
		{"unpaired_high_surrogate", `{"count":1,"text":"\ud800"}`, ErrUnicode},
		{"unpaired_low_surrogate", `{"count":1,"text":"\udc00"}`, ErrUnicode},
		{"null", `{"count":1,"text":null}`, ErrJSON},
		{"trailing_value", `{"count":1,"text":"safe"} {}`, ErrJSON},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var decoded fields
			if err := strictJSON([]byte(tc.raw), &decoded); err != tc.want {
				t.Fatalf("strict stored JSON error = %v; want %v", err, tc.want)
			}
		})
	}
	var decoded fields
	if err := strictJSON([]byte(`{"count":1,"text":"\ud83d\ude00"}`), &decoded); err != nil || decoded.Text != "😀" {
		t.Fatal("valid Unicode surrogate pair was not preserved")
	}
	if err := strictJSON([]byte(`{"count":1,"text":"\\ud800"}`), &decoded); err != nil || decoded.Text != `\ud800` {
		t.Fatal("literal escaped backslash was treated as a Unicode escape")
	}
	invalidUTF8 := append([]byte(`{"count":1,"text":"`), 0xff)
	invalidUTF8 = append(invalidUTF8, []byte(`"}`)...)
	if err := strictJSON(invalidUTF8, &decoded); err != ErrJSON {
		t.Fatal("invalid raw UTF-8 was accepted")
	}
}

func TestStoredWireArraysKeepExtraElements(t *testing.T) {
	var outcome wireOutcome
	raw := []byte(`{"parent_id":"parent","state":"known","acceptable_candidate_indices":[0,1,2,3,4],"candidate_truth":[],"reason":""}`)
	if err := strictJSON(raw, &outcome); err != nil || len(outcome.Acceptable) != 5 || outcome.Acceptable[4] != 4 {
		t.Fatal("wire decoding silently discarded an extra array element")
	}
	p := wireParent{ID: "parent", Candidates: []wireCandidate{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}}
	if _, err := bindOutcome(p, outcome); err != ErrIdentity {
		t.Fatal("extra acceptable indices escaped the original candidate-count boundary")
	}
}

func TestStoredOutcomeStatesAndOrder(t *testing.T) {
	p := wireParent{ID: "parent", Candidates: []wireCandidate{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}}}
	good := wireOutcome{ParentID: "parent", State: "known", Acceptable: []int{1, 3}, Candidates: []CandidateTruth{{"a", "rejected", 2, 1}, {"b", "acceptable", 2, 0}, {"c", "rejected", 2, 1}, {"d", "acceptable", 2, 0}}}
	if truth, err := bindOutcome(p, good); err != nil || truth.AcceptableCount != 2 || truth.AcceptableIndices != [4]int{1, 3, 0, 0} {
		t.Fatal("multiple stored acceptable candidates were not preserved")
	}
	cases := []struct {
		name   string
		change func(*wireOutcome)
		want   error
	}{
		{"parent_identity", func(o *wireOutcome) { o.ParentID = "other" }, ErrIdentity},
		{"negative_index", func(o *wireOutcome) { o.Acceptable = []int{-1} }, ErrTruth},
		{"out_of_range", func(o *wireOutcome) { o.Acceptable = []int{4} }, ErrTruth},
		{"duplicate_index", func(o *wireOutcome) { o.Acceptable = []int{1, 1} }, ErrTruth},
		{"unsorted_indices", func(o *wireOutcome) { o.Acceptable = []int{3, 1} }, ErrTruth},
		{"known_without_answer", func(o *wireOutcome) { o.Acceptable = nil }, ErrTruth},
		{"candidate_order", func(o *wireOutcome) { o.Candidates[0], o.Candidates[1] = o.Candidates[1], o.Candidates[0] }, ErrTruth},
		{"candidate_count", func(o *wireOutcome) { o.Candidates = o.Candidates[:3] }, ErrIdentity},
		{"contradictory_label", func(o *wireOutcome) { o.Candidates[1].State = "rejected"; o.Candidates[1].FailedVectors = 1 }, ErrTruth},
		{"acceptable_with_failed_vector", func(o *wireOutcome) { o.Candidates[1].FailedVectors = 1 }, ErrTruth},
		{"no_answer_with_acceptable", func(o *wireOutcome) { o.State = "no_answer" }, ErrTruth},
		{"unknown_with_answer", func(o *wireOutcome) { o.State = "unknown"; o.Reason = "policy_unavailable" }, ErrTruth},
		{"unexpected_state", func(o *wireOutcome) { o.State = "false" }, ErrTruth},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := good
			changed.Acceptable = append([]int(nil), good.Acceptable...)
			changed.Candidates = append([]CandidateTruth(nil), good.Candidates...)
			tc.change(&changed)
			truth, err := bindOutcome(p, changed)
			if err != tc.want || !reflect.DeepEqual(truth, Truth{}) {
				t.Fatalf("inconsistent stored outcome error = %v; want %v and zero truth", err, tc.want)
			}
		})
	}
	p.Candidates = p.Candidates[:2]
	noAnswer := wireOutcome{ParentID: "parent", State: "no_answer", Candidates: []CandidateTruth{{"a", "rejected", 1, 1}, {"b", "rejected", 1, 1}}}
	if truth, err := bindOutcome(p, noAnswer); err != nil || truth.State != "no_answer" || truth.AcceptableCount != 0 || truth.CandidateCount != 2 {
		t.Fatal("two-candidate no-answer metadata was not preserved")
	}
	p.Candidates = append(p.Candidates, wireCandidate{ID: "c"})
	unknown := wireOutcome{ParentID: "parent", State: "unknown", Reason: "policy_unavailable", Candidates: []CandidateTruth{{"a", "unknown", 0, 0}, {"b", "unknown", 0, 0}, {"c", "unknown", 0, 0}}}
	if truth, err := bindOutcome(p, unknown); err != nil || truth.State != "unknown" || truth.CandidateCount != 3 || truth.Reason != unknown.Reason {
		t.Fatal("three-candidate unknown metadata was not preserved")
	}
	unknown.Candidates[0].VectorsChecked = 1
	if _, err := bindOutcome(p, unknown); err != ErrTruth {
		t.Fatal("unknown metadata with observed vectors was accepted")
	}
	unknown.Candidates[0].VectorsChecked = 0
	unknown.Reason = ""
	if _, err := bindOutcome(p, unknown); err != ErrTruth {
		t.Fatal("unknown metadata without its reason was accepted")
	}
}

func TestRawTextAndCompiledProvenance(t *testing.T) {
	if !validText(strings.Repeat("x", 512)) || validText(strings.Repeat("x", 513)) || validText(" \n\t") || validText("x\x00y") || validText(string([]byte{0xff})) {
		t.Fatal("raw stored prose bounds changed")
	}
	pins := CompiledFiles("internal/storedaudit")
	if len(pins) != 2 {
		t.Fatal("compiled source list is not closed to two files")
	}
	for _, pin := range pins {
		raw, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(pin.Path)))
		if err != nil || sum(raw) != pin.SHA256 || len(raw) != pin.Bytes {
			t.Fatal("compiled source provenance differs from disk")
		}
	}
	pins[0] = FilePin{}
	if CompiledFiles("internal/storedaudit")[0].Path != "internal/storedaudit/binding.go" {
		t.Fatal("returned source metadata can mutate compiled provenance")
	}
}
