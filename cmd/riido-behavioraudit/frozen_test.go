package main

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
)

// The CI oracle reruns Go behavior and every nonlearned ranking; the published
// result is a comparison target, never a scoring feature or generated label.
func TestFrozenPublic56AReproduces(t *testing.T) {
	root := "../../"
	planBytes, err := os.ReadFile(root + "experiments/short-claim/execution-plan-56a.json")
	if err != nil {
		t.Fatal(err)
	}
	if hash(planBytes) != "9c8461b9fe0a80a0f0976c8f07a68f3a47f0c534800a0a303e1cbd8a613caaeb" {
		t.Fatal("execution plan changed")
	}
	var plan auditPlan
	if err := json.Unmarshal(planBytes, &plan); err != nil {
		t.Fatal(err)
	}
	for _, pin := range plan.ImplementationFiles {
		b, err := os.ReadFile(root + pin.Path)
		if err != nil || hash(b) != pin.SHA256 {
			t.Fatal("runtime implementation pin changed")
		}
	}
	raw, err := os.ReadFile(root + "experiments/short-claim/probes-56.json")
	if err != nil || hash(raw) != plan.InputSHA256 {
		t.Fatal("original fixture pin changed")
	}
	d, err := behaviorprobe.Load(root + "experiments/short-claim/probes-56.json")
	if err != nil {
		t.Fatal(err)
	}
	truth, err := behaviorprobe.Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	if truth.SourceArtifactSHA256 != plan.SourceArtifactSHA256 {
		t.Fatal("embedded source changed")
	}
	actual, err := evaluate(d, truth, plan)
	if err != nil {
		t.Fatal(err)
	}
	actual.PlanSHA256, actual.InputSHA256 = hash(planBytes), hash(raw)
	savedBytes, err := os.ReadFile(root + "experiments/short-claim/results-56a.json")
	if err != nil {
		t.Fatal(err)
	}
	var saved report
	if err := json.Unmarshal(savedBytes, &saved); err != nil {
		t.Fatal(err)
	}
	if len(actual.ParentRankings) != len(saved.ParentRankings) {
		t.Fatal("ranking parent count changed")
	}
	// Preserve exact orders and integer evidence across Linux/macOS while
	// allowing tiny floating math rounding differences in unverified scores.
	for i := range actual.ParentRankings {
		for j := 0; j < 4; j++ {
			for k := 0; k < 8; k++ {
				a, b := actual.ParentRankings[i].Rankings[j].Scores[k], saved.ParentRankings[i].Rankings[j].Scores[k]
				if math.IsNaN(a) || math.IsInf(a, 0) || math.IsNaN(b) || math.IsInf(b, 0) || math.Abs(a-b) > 1e-12*(1+math.Abs(b)) {
					t.Fatal("baseline score changed")
				}
				actual.ParentRankings[i].Rankings[j].Scores[k] = 0
				saved.ParentRankings[i].Rankings[j].Scores[k] = 0
			}
		}
	}
	if !reflect.DeepEqual(actual, saved) {
		t.Fatal("public diagnostic truth/groups/rankings/denominators failed independent replay")
	}
}
