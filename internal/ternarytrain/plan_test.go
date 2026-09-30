package ternarytrain

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFrozenFinalAndGates(t *testing.T) {
	root := "../../experiments/ternary-qat/"
	b, e := os.ReadFile(root + "final-families.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		ID     string `json:"id"`
		Family string `json:"family"`
		Label  int    `json:"label"`
		State  string `json:"state"`
		Split  string `json:"split"`
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	counts := [3]int{}
	if len(rows) != 36 {
		t.Fatal("final changed")
	}
	for _, r := range rows {
		if seen[r.ID] || seen[r.State] || r.Split != "final" || r.Family == "" || r.Label < 0 || r.Label > 2 {
			t.Fatal("invalid final case")
		}
		seen[r.ID] = true
		seen[r.State] = true
		counts[r.Label]++
	}
	if counts != [3]int{12, 12, 12} {
		t.Fatal("label counts changed")
	}
	for _, file := range []string{"plan.json", "plan-01.json"} {
		p, e := os.ReadFile(root + file)
		if e != nil {
			t.Fatal(e)
		}
		var x struct {
			Hash       string  `json:"final_families_sha256"`
			Confidence float64 `json:"confidence_gate"`
			Acceptance struct {
				Bytes     int     `json:"max_head_bytes"`
				Accuracy  float64 `json:"minimum_accuracy"`
				Drop      float64 `json:"max_accuracy_drop_vs_parent"`
				Strong    float64 `json:"minimum_strong_recall"`
				Errors    int     `json:"strong_to_fast_errors"`
				Precision float64 `json:"minimum_accepted_precision"`
				Coverage  float64 `json:"minimum_coverage"`
				Both      bool    `json:"both_seeds_required"`
			} `json:"acceptance"`
		}
		if e = json.Unmarshal(p, &x); e != nil {
			t.Fatal(e)
		}
		a := x.Acceptance
		if x.Hash != Hash(b) || x.Confidence != .9 || a.Bytes != 600 || a.Accuracy != .8 || a.Drop != .03 || a.Strong != .875 || a.Errors != 0 || a.Precision != .9 || a.Coverage != .25 || !a.Both {
			t.Fatal("frozen plan/gates drifted")
		}
	}
}
