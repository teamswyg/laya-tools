// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaimdata"
)

// These are owned synthetic rows, never a real source observation or data30.
const toy = `{"schema":"riido-finite-development-request-row-v1","stable_id":"owned-synthetic","request":"Inspect frozen input","candidates":[{"metadata_id":"owned-a","text":"Hint one","label":false,"sample_weight":1},{"metadata_id":"owned-b","text":"Hint two","label":true,"sample_weight":1}],"role":"development_train","whole_group":76,"source_family":"owned-fixture","source_revision":"0000000000000000000000000000000000000000","finite_scope":{"input_count":1,"candidate_observations":2,"description":"Only owned synthetic input","all_input_guarantee":false,"unseen_source_guarantee":false},"text_revision":"owned-v1","feature_policy":"request_and_candidates_text_only"}`

func toySpec(raw []byte) rowSpec { return rowSpec{digest(raw), "owned-synthetic", 76, 1, 2, 1} }

func TestReaderCorrespondenceAndCandidateSwap(t *testing.T) {
	raw := []byte(toy)
	var out summary
	z, e := verifyRow(raw, toySpec(raw), shortclaimdata.LoadDevelopmentRow, &out)
	if e != nil || z.Labels != 2 || z.Positive != 1 || z.Negative != 1 || out.Attempts != 1 || out.Returns != 1 || out.Matches != 1 {
		t.Fatal("synthetic Reader correspondence")
	}
	var w wireRow
	if !decode(raw, &w) {
		t.Fatal("fixture")
	}
	ex, e := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(raw))
	if e != nil {
		t.Fatal("fixture reader")
	}
	w.Candidates[0], w.Candidates[1] = w.Candidates[1], w.Candidates[0]
	if _, e = compareProjection(w, ex, toySpec(raw), &out); e == nil {
		t.Fatal("permuted candidate mapping accepted")
	}
}

func TestMalformedRowAndAllFalseDoNotBecomeNoAnswer(t *testing.T) {
	for _, raw := range []string{
		strings.Replace(toy, `"label":true,`, "", 1),
		strings.Replace(toy, `"sample_weight":1`, `"sample_weight":null`, 1),
		strings.Replace(toy, `"schema":`, `"schema":"duplicate","schema":`, 1),
		strings.Replace(toy, `"role":"development_train"`, `"role":"validation"`, 1),
		strings.Replace(toy, `"label":true`, `"label":false`, 1),
	} {
		var out summary
		if _, e := verifyRow([]byte(raw), toySpec([]byte(raw)), shortclaimdata.LoadDevelopmentRow, &out); e == nil {
			t.Fatal("malformed or altered supervision accepted")
		}
		if out.Matches != 0 {
			t.Fatal("failure recorded as accepted row")
		}
	}
}

func TestPinnedMutationPrecedesReader(t *testing.T) {
	var out summary
	calls := 0
	load := func(io.Reader) (shortclaimdata.Example, error) {
		calls++
		return shortclaimdata.Example{}, shortclaimdata.ErrRead
	}
	if _, e := verifyRow([]byte(strings.Replace(toy, "Hint one", "Hint altered", 1)), toySpec([]byte(toy)), load, &out); e != failure("row_pin") || calls != 0 || out.Attempts != 0 {
		t.Fatal("unfrozen text dispatched")
	}
}

func TestFailedReaderStopsAtKnownPrefix(t *testing.T) {
	var out summary
	calls := 0
	raw := []byte(toy)
	want := toySpec(raw)
	load := func(r io.Reader) (shortclaimdata.Example, error) {
		calls++
		if calls == 2 {
			return shortclaimdata.Example{}, shortclaimdata.ErrRead
		}
		return shortclaimdata.LoadDevelopmentRow(r)
	}
	z, e := streamRows(strings.NewReader(toy+"\n"+toy+"\n"+toy+"\n"), []rowSpec{want, want, want}, load, &out)
	if e != failure("project_reader") || calls != 2 || out.Attempts != 2 || out.Returns != 2 || out.Matches != 1 || z.Rows != 1 {
		t.Fatal("failure/prefix counters lost or next row dispatched")
	}
}

func TestStreamRejectsExtraAndOversizeRows(t *testing.T) {
	for _, raw := range []string{toy + "\n\n", strings.Repeat("x", shortclaimdata.MaxRowBytes+1) + "\n", toy} {
		var out summary
		if _, e := streamRows(strings.NewReader(raw), []rowSpec{toySpec([]byte(toy))}, shortclaimdata.LoadDevelopmentRow, &out); e == nil {
			t.Fatal("row boundary accepted")
		}
	}
}

func TestMetadataMustUseHistoricalExactPin(t *testing.T) {
	// Fake declarations, even plausible counts, cannot stand in for Root's pin.
	raw, _ := json.Marshal(materialization{Schema: "riido-next60-27-or30-data-correspondence-v1", Requests: 30, Labels: 88})
	if verifyMetadata(raw) != failure("metadata_pin") {
		t.Fatal("invented metadata accepted")
	}
}

func TestArgumentsAreFixedAndRedacted(t *testing.T) {
	var out, err bytes.Buffer
	if run([]string{"--data", "caller-sensitive-string"}, &out, &err) != 2 || out.Len() != 0 || err.String() != "developmentverify_arguments\n" {
		t.Fatal("argument diagnostics retained supplied text")
	}
}
