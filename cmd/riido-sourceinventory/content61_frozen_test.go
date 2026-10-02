// SPDX-License-Identifier: Apache-2.0
package main

import (
	"cmp"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// An archive guard only: no original inventory/behavior/model API is invoked.
// Hashes preserve the first assessments, including their unresolved meanings.
func TestFrozenRemainingContentReview61(t *testing.T) {
	type pin struct {
		Path  string `json:"path"`
		Bytes int    `json:"bytes"`
		SHA   string `json:"sha256"`
	}
	pins := [...]pin{
		{"content-review-manifest-61.json", 467067, "f8f0b2320bb214f102575384cace477067a0f075b98ee09c071e778d9285224b"},
		{"content-review-evidence-61.json", 315807, "bffd66f6586ad3991697c49c491e949f1895537de95a8aabc7dfcff205bfa28a"},
		{"actual-content-review-attempt-1-ledger.json", 2068, "fba824de52d5469645200297ca0c5d671c72a42f0be4936c05dfb8bc361f5b83"},
		{"content-review-reader-byte-receipt-61.json", 5570, "dfed9c03bea82b99505e36921f4f7a09d5592028dd474ef84edee66244dedcf7"},
		{"reader-byte-QA-attempt-ledger.json", 1833, "1d641d7fd51b825a223b39afcc155d8c98220494b428f1c6c5319446cf6833c2"},
		{"SERIALIZER-BUILD-LEDGER-61.json", 1777, "2d1943ab98e09453932819e8f48fad95e457b05a1171367ede76a387bd767381"},
		{"content-review-mechanics-61.json", 9669, "1f8ee9bcf1f62cc18a575c86e413fd50384634bc213dd0e0dd215fc1890ab5d8"},
		{"content-review-mechanics-ledger-61.json", 4400, "fd9a23d99c8c86e61ea12288a264c517da6ccdd10b35f8236f63502baafa5d42"},
		{"content-review-mechanics-failed-attempt-1-61.json", 798, "5bf264360b1c980d7aa87d2384af6f8e9ad2c1ce7e7a2d7d9561304b8c8a963e"},
		{"content-review-mechanics-plan-attempt-1-61.json", 1305, "8ae8ca76def82ce6195dc55baac64ae8094a3533f36a6734e0c38e8aefd99721"},
		{"content-review-mechanics-plan-attempt-2-61.json", 1075, "7dd82aff9e35382a7066d7314e6f86cbc16850508d05973c95a1bb23c50d66f4"},
	}
	read := func(p pin) []byte {
		t.Helper()
		f, err := os.Open(filepath.Join("..", "..", "experiments", "short-claim", p.Path))
		if err != nil {
			t.Fatal(err)
		}
		// Each frozen shard is below 1 MiB; this does not change runtime caps.
		b, readErr := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || len(b) > 1<<20 || len(b) != p.Bytes || sha(b) != p.SHA {
			t.Fatal("frozen content61 artifact changed", p.Path, readErr, closeErr)
		}
		return b
	}
	type join struct {
		ID    string `json:"claim_id"`
		SHA   string `json:"record_sha256"`
		State string `json:"state"`
	}
	var manifest struct {
		Schema string `json:"schema"`
		Ready  bool   `json:"training_ready"`
		Fits   int    `json:"fits"`
		Labels int    `json:"new_labels"`
		Roles  int    `json:"roles_assigned"`
		Counts struct {
			Records   int `json:"all_record_objects"`
			Original  int `json:"assessed_original59_entries"`
			Approvals int `json:"candidate_approvals"`
		} `json:"counts"`
		Preservation struct {
			Parents         int `json:"original_parents"`
			Candidates      int `json:"original_candidate_positions"`
			Unknown         int `json:"unknown_global_original"`
			SelectedUnknown int `json:"unknown_selected_original"`
			Groups          int `json:"connected_groups"`
			KnownGroups     int `json:"known_containing_groups"`
			FinalReads      int `json:"protected_final_reads"`
		} `json:"preservation"`
		Shards        []pin  `json:"family_shards"`
		Selected      []join `json:"selected_existing59_entries"`
		Supplementary []join `json:"supplementary_observation_and_negative_axis_records"`
	}
	if err := json.Unmarshal(read(pins[0]), &manifest); err != nil {
		t.Fatal(err)
	}
	for _, p := range pins[1:] {
		read(p)
	}
	if manifest.Schema != "riido-caption-content-scoped-review-manifest-61-v1" || manifest.Ready || manifest.Fits != 0 || manifest.Labels != 0 || manifest.Roles != 0 || manifest.Counts.Approvals != 0 || manifest.Counts.Records != 975 || manifest.Counts.Original != 615 || len(manifest.Shards) != 15 || len(manifest.Selected) != 615 || len(manifest.Supplementary) != 360 {
		t.Fatal("scope or training boundary changed")
	}
	p := manifest.Preservation
	if p.Parents != 72 || p.Candidates != 216 || p.Unknown != 21 || p.SelectedUnknown != 18 || p.Groups != 17 || p.KnownGroups != 16 || p.FinalReads != 0 {
		t.Fatal("preservation boundary changed")
	}
	type record struct {
		ID       string `json:"claim_id"`
		SHA      string `json:"review_record_sha256"`
		State    string `json:"state"`
		Approval bool   `json:"aggregate_candidate_approval"`
	}
	records := make([]record, 0, 975)
	for _, p := range manifest.Shards {
		if filepath.Base(p.Path) != p.Path {
			t.Fatal("shard path must be a basename")
		}
		var shard struct {
			Schema   string   `json:"schema"`
			Ready    bool     `json:"training_ready"`
			Approval bool     `json:"aggregate_candidate_approval"`
			Records  []record `json:"review_records"`
		}
		if err := json.Unmarshal(read(p), &shard); err != nil {
			t.Fatal(err)
		}
		if shard.Schema != "riido-caption-content-family-review-61-v1" || shard.Ready || shard.Approval || len(shard.Records) != 65 {
			t.Fatal("family scope changed", p.Path)
		}
		records = append(records, shard.Records...)
	}
	slices.SortFunc(records, func(a, b record) int { return cmp.Compare(a.ID, b.ID) })
	for i, r := range records {
		if r.ID == "" || r.Approval || (i > 0 && records[i-1].ID == r.ID) {
			t.Fatal("nonunique or approved record", r.ID)
		}
	}
	joins := append(slices.Clone(manifest.Selected), manifest.Supplementary...)
	slices.SortFunc(joins, func(a, b join) int { return cmp.Compare(a.ID, b.ID) })
	if len(joins) != len(records) {
		t.Fatal("join cardinality changed")
	}
	for i, j := range joins {
		r := records[i]
		if j.ID != r.ID || j.SHA != r.SHA || j.State != r.State {
			t.Fatal("record bijection changed", j.ID)
		}
	}
	states := [...]string{"consistent_with_scoped_evidence", "contradicts_scoped_evidence", "omits_required_scope", "unsupported_or_uncertain"}
	var counts [4]int
	for _, j := range manifest.Selected {
		i := slices.Index(states[:], j.State)
		if i < 0 {
			t.Fatal("unknown assessment state", j.ID)
		}
		counts[i]++
	}
	if counts != [4]int{446, 94, 10, 65} {
		t.Fatal("first assessment states changed", counts)
	}
}
