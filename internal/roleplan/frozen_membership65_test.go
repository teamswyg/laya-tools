// SPDX-License-Identifier: Apache-2.0
package roleplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Archive preservation only: no membership hashing, ordering, allocation,
// assignment, source execution, feature extraction or training is performed.
func TestFrozenWholeGroupPreparation65(t *testing.T) {
	type pin struct {
		name string
		size int
		sha  string
	}
	pins := [...]pin{
		{"membership-proposal-65.json", 1986200, "ae2794224ab8fdb03aa879b2522777beb5065005286b6394db3ba4a8abcdd1ce"},
		{"binding-receipt-65.json", 15776, "c6be0a869af3946e9c4c52b602be49f08a74fe7bce9abe810f9f93b7e747ab9b"},
		{"ledger-65.json", 2118, "f0697c66d77fa82bfcba1423a4d2ba590d737c21fffeee6d8fd9170aec542e37"},
		{"HANDOFF-65.ko.md", 3851, "32c4368087b4e5f6fd0736c945bf7647e28041c848ae74983eb0fae83f7ca057"},
		{"HANDOFF-65.en.md", 3642, "bf8d2b0efd1fffc49bf3ce7917e112d64923dc756799456c6a1cd3688768f920"},
		{"INDEPENDENT-AUDIT-65.json", 13400, "fb1314fa6bc35d297fa12f1c2c8e3e4528a0bef33198cded05eb4da71cf42a0d"},
		{"INDEPENDENT-AUDIT-65.attempt1.json", 550, "586b9c85cce1e72ed2c99f92cdb68af924ed32c33c8c7e1797bb7b20ba07d0e3"},
		{"ATTEMPT-1-FAILURE-LEDGER.json", 773, "eeca04913ab9c2c833cb491bc21c297a880fabbd76d3089af8c78c461781016c"},
		{"ATTEMPT-2-LEDGER.json", 1282, "83e07d2e86710bf0f5a6547139a80e132bb62d41b11a4524eaf891c839fc4f50"},
		{"QA-65.ko.md", 3179, "db691235fdaea0284cd51735b823d591e7d85e79616224273f6194524bd3bce1"},
		{"QA-65.en.md", 3139, "85afb82bed295ddb5108bd779eae6ea9129c3a40be3bc2c9b4af20437a433865"},
	}
	const maxFile = 2 << 20
	root := filepath.Join("..", "..", "experiments", "short-claim", "whole-group-preparation")
	for _, p := range pins {
		f, err := os.Open(filepath.Join(root, p.name))
		if err != nil {
			t.Fatal(err)
		}
		b, readErr := io.ReadAll(io.LimitReader(f, maxFile+1))
		closeErr := f.Close()
		h := sha256.Sum256(b)
		if readErr != nil || closeErr != nil || len(b) > maxFile || len(b) != p.size || hex.EncodeToString(h[:]) != p.sha {
			t.Fatal("whole-group archive changed", p.name)
		}
		if p.name != "membership-proposal-65.json" {
			continue
		}
		var x struct {
			Ready    bool `json:"training_ready"`
			Seed     bool `json:"seed_assigned"`
			Roles    int  `json:"roles_assigned"`
			Assign   int  `json:"Assign_calls"`
			Order    int  `json:"OrderDigest_calls"`
			Allocate int  `json:"AllocateCounts_calls"`
			Fits     int  `json:"fits"`
			Groups   []struct {
				ID      int  `json:"original_group_id"`
				Known   bool `json:"has_known_members"`
				Parents []struct {
					State string `json:"saved_truth_state"`
					Count int    `json:"candidate_positions"`
				} `json:"parents"`
				Members   []string          `json:"canonical_members"`
				Relations []json.RawMessage `json:"relationships"`
			} `json:"groups"`
		}
		if json.Unmarshal(b, &x) != nil || x.Ready || x.Seed || x.Roles != 0 || x.Assign != 0 || x.Order != 0 || x.Allocate != 0 || x.Fits != 0 || len(x.Groups) != 17 {
			t.Fatal("archived proposal execution status changed")
		}
		var states [3]int
		parents, candidates, members, relations, knownGroups, unknownOnly := 0, 0, 0, 0, 0, 0
		for i, g := range x.Groups {
			for j := 0; j < i; j++ {
				if x.Groups[j].ID == g.ID {
					t.Fatal("duplicate original group")
				}
			}
			known := false
			for _, p := range g.Parents {
				parents++
				candidates += p.Count
				switch p.State {
				case "known":
					states[0]++
					known = true
				case "no_answer":
					states[1]++
					known = true
				case "unknown":
					states[2]++
				default:
					t.Fatal("invalid archived truth state")
				}
			}
			if known != g.Known {
				t.Fatal("archived known flag changed")
			}
			if known {
				knownGroups++
			} else {
				unknownOnly++
				if g.ID != 64 || len(g.Parents) != 4 {
					t.Fatal("unknown-only group changed")
				}
			}
			members += len(g.Members)
			relations += len(g.Relations)
		}
		if states != [3]int{34, 17, 21} || parents != 72 || candidates != 216 || members != 453 || relations != 1922 || knownGroups != 16 || unknownOnly != 1 {
			t.Fatal("archived graph cardinality changed")
		}
	}
}
