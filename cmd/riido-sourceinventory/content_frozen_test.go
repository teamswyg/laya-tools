// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Post-observation archive guard: integrity does not prove interpretation.
// No inventory, AST, formatter, candidate, model or training API is executed.
func TestFrozenScopedContentReview60(t *testing.T) {
	pins := []struct {
		path   string
		bytes  int
		sha256 string
	}{
		{"content-review-assessments-60.json", 1280021, "1beb5c06d3fb70c0a5e0727de793f1890b4ab4a47f1d392a6666b068f096c526"},
		{"content-review-evidence-60.json", 37101, "be5bb734158c2148818c7f97e2965eb9b9f0d0f76b33506d3262d7671fbe6bea"},
		{"content-review-byte-receipt-60.json", 1689, "1ac37c2f2f1e8d10d2f5cda02e7e3f96d6b8ba1ba466f453cb7b39f3a4e4f170"},
		{"content-review-mechanics-60.json", 2536, "5b5daf45ca45721ef26fabc8eb476b0fac4613febca713ed5ea424149c6b2d29"},
		{"content-review-mechanics-ledger-60.json", 2904, "fe5aa7f22018c1a877b8b5518297231d46e8f253d2e3406df357efc6770689e4"},
		{"content-review-protocol-60.json", 26067, "6f4504ba04e0b5813e4dc58ec0f6dbe5101a1127e09a379172dd57d7ed11ff19"},
		{"content-review-scope-references-60.json", 96306, "0f3f0b7742bd74a45e5c263f501bcce63964e2432a3cafb3217156d694f6ffdd"},
		{"caption-coverage-59.json", 360591, "7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29"},
		{"source-inventory-60.json", 174651, "d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4"},
	}
	var assessment []byte
	for i, p := range pins {
		f, err := os.Open(filepath.Join("..", "..", "experiments", "short-claim", p.path))
		if err != nil {
			t.Fatal(err)
		}
		// Allow the 1.28 MB archive here; runtime input caps remain unchanged.
		b, readErr := io.ReadAll(io.LimitReader(f, (2<<20)+1))
		closeErr := f.Close()
		if readErr != nil || closeErr != nil || len(b) > 2<<20 || len(b) != p.bytes || sha(b) != p.sha256 {
			t.Fatal("archived content evidence changed", p.path, readErr, closeErr)
		}
		if i == 0 {
			assessment = b
		}
	}
	type join struct {
		ID    string `json:"claim_id"`
		SHA   string `json:"record_sha256"`
		State string `json:"state"`
	}
	var a struct {
		Schema        string `json:"schema"`
		TrainingReady bool   `json:"training_ready"`
		Counts        struct {
			Assessed  int `json:"assessed_original59_entries"`
			Untouched int `json:"untouched_original59_entries"`
			Approvals int `json:"aggregate_candidate_approvals"`
		} `json:"counts"`
		Preservation struct {
			Unknown         int `json:"unknown_parents"`
			SelectedUnknown int `json:"unknown_selected_parents"`
			Groups          int `json:"connected_groups"`
			KnownGroups     int `json:"known_containing_groups"`
			FinalReads      int `json:"protected_final_reads"`
		} `json:"preservation"`
		Selected      []join `json:"selected_existing59_entries"`
		Supplementary []join `json:"supplementary_observation_and_negative_axis_records"`
		Records       []struct {
			ID       string `json:"claim_id"`
			SHA      string `json:"review_record_sha256"`
			State    string `json:"state"`
			Approval bool   `json:"aggregate_candidate_approval"`
		} `json:"review_records"`
	}
	if err := json.Unmarshal(assessment, &a); err != nil {
		t.Fatal(err)
	}
	if a.Schema != "riido-caption-content-scoped-review-60-v1" || a.TrainingReady || a.Counts.Assessed != 123 || a.Counts.Untouched != 615 || a.Counts.Approvals != 0 || len(a.Selected) != 123 || len(a.Supplementary) != 72 || len(a.Records) != 195 {
		t.Fatal("review scope or training boundary changed")
	}
	if a.Preservation.Unknown != 21 || a.Preservation.SelectedUnknown != 3 || a.Preservation.Groups != 17 || a.Preservation.KnownGroups != 16 || a.Preservation.FinalReads != 0 {
		t.Fatal("historical uncertainty or group boundary changed")
	}
	states := [...]string{"consistent_with_scoped_evidence", "contradicts_scoped_evidence", "omits_required_scope", "unsupported_or_uncertain"}
	var counts [4]int
	for _, set := range [][]join{a.Selected, a.Supplementary} {
		for _, j := range set {
			matches := 0
			for _, r := range a.Records {
				if r.ID == j.ID {
					matches++
					if r.SHA != j.SHA || r.State != j.State || r.Approval {
						t.Fatal("record join changed", j.ID)
					}
				}
			}
			if matches != 1 {
				t.Fatal("record join must be unique", j.ID)
			}
		}
	}
	for _, j := range a.Selected {
		found := false
		for i, state := range states {
			if j.State == state {
				counts[i]++
				found = true
			}
		}
		if !found {
			t.Fatal("unexpected review state", j.ID)
		}
	}
	if counts != [4]int{78, 17, 2, 26} {
		t.Fatal("mapped assessment counts changed", counts)
	}
}
