package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/behaviorprobe"
	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
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
	// Historical bytes remain exact; current optimized code is checked below
	// as a separate numerical replay, not represented as the original runtime.
	for _, pin := range plan.ImplementationFiles {
		path := pin.Path
		switch path {
		case "pkg/shortclaim/input.go":
			path = "testdata/shortclaim-source-9d204/input.go.txt"
		case "pkg/shortclaim/baseline.go":
			path = "testdata/shortclaim-source-9d204/baseline.go.txt"
		case "cmd/riido-shortclaim/main.go":
			path = "testdata/shortclaim-source-9d204/cli-main.go.txt"
		case "cmd/riido-shortclaim/bench.go":
			path = "testdata/shortclaim-source-9d204/cli-bench.go.txt"
		}
		b, err := os.ReadFile(root + path)
		if err != nil || hash(b) != pin.SHA256 {
			t.Fatal("historical implementation pin changed")
		}
	}
	// Compiled package-source exports must describe today's on-disk runtime.
	var current []sourcePin
	for _, a := range shortclaim.AuditSourceArtifacts() {
		current = append(current, sourcePin{"pkg/shortclaim/" + a.Name, a.SHA256})
	}
	for _, a := range lexicalhint.AuditSourceArtifacts() {
		current = append(current, sourcePin{"internal/lexicalhint/" + a.Name, a.SHA256})
	}
	for _, a := range behaviorprobe.AuditSourceArtifacts() {
		current = append(current, sourcePin{"internal/behaviorprobe/" + a.Name, a.SHA256})
	}
	for _, pin := range current {
		b, err := os.ReadFile(root + pin.Path)
		if err != nil || hash(b) != pin.SHA256 {
			t.Fatal("current compiled runtime pin changed")
		}
	}
	// The actual CLI continues to refuse the old source-bound plan. This fails
	// before behavior evaluation, output reservation, ranking or observations.
	t.Chdir(root)
	var rejected bytes.Buffer
	out := filepath.Join(t.TempDir(), "old-plan")
	if err := run([]string{"--plan-sha256", hash(planBytes), "--out", out}, &rejected); err == nil || err.Error() != "implementation_hash_mismatch" || rejected.Len() != 0 {
		t.Fatal("current CLI accepted historical runtime pins")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("historical-plan refusal reserved output")
	}
	root = ""
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
	if err != nil || hash(savedBytes) != "c02c01f9ecce3ac16fd850d4239aeed2132fed32286d2628c472d873234b9ac6" {
		t.Fatal("historical result changed")
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
