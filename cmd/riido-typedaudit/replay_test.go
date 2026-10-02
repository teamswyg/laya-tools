package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

// Preparation may test controls before collecting the official report. Once
// published, both CI platforms must reproduce the frozen numerical evidence.
func TestFrozenPublicTruthAndGroupsReplay(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(root, "experiments/short-claim/results-56b.json")
	wantRaw, err := os.ReadFile(resultPath)
	if errors.Is(err, os.ErrNotExist) {
		for _, artifact := range []string{"execution-plan-56b.json", "probes-56b.json"} {
			if _, e := os.Stat(filepath.Join(root, "experiments/short-claim", artifact)); e == nil || !errors.Is(e, os.ErrNotExist) {
				t.Fatal("frozen inputs exist but their official truth/groups report is missing")
			}
		}
		t.Skip("official truth/groups report not yet collected")
	}
	if err != nil {
		t.Fatal(err)
	}
	if hash(wantRaw) != "4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4" {
		t.Fatal("frozen truth/groups report changed")
	}
	planPath := "experiments/short-claim/execution-plan-56b.json"
	planRaw, err := os.ReadFile(filepath.Join(root, planPath))
	if err != nil {
		t.Fatal(err)
	}
	if hash(planRaw) != "2fe7aeccd72198fda67f2b0ad9e067a2092b465bef252ba72fb464662cf40a82" {
		t.Fatal("frozen plan changed")
	}
	planHash := sha256.Sum256(planRaw)
	var original auditPlan
	if !decodePlan(planRaw, &original) || validatePlan(original) != nil {
		t.Fatal("frozen plan is invalid")
	}
	current := original
	current.ImplementationFiles = slices.Clone(original.ImplementationFiles)
	for i, pin := range original.ImplementationFiles {
		path := pin.Path
		switch path {
		case "pkg/shortclaim/input.go":
			path = "testdata/shortclaim-source-9d204/input.go.txt"
		case "pkg/shortclaim/baseline.go":
			path = "testdata/shortclaim-source-9d204/baseline.go.txt"
		case "cmd/riido-typedaudit/replay_test.go":
			path = "testdata/shortclaim-source-9d204/typed-replay-test.go.txt"
		}
		historical, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || hash(historical) != pin.SHA256 {
			t.Fatal("historical implementation pin changed")
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(pin.Path)))
		if err != nil {
			t.Fatal("current replay source unavailable")
		}
		current.ImplementationFiles[i].SHA256 = hash(raw)
	}
	if err := verifySources(current, root); err != nil {
		t.Fatal("current compiled and on-disk source differ")
	}
	t.Chdir(root)
	out := filepath.Join(t.TempDir(), "replay")
	var status bytes.Buffer
	// The real CLI still refuses the original source pins before observations.
	if err := run([]string{"--stage", "audit", "--plan", planPath, "--plan-sha256", hex.EncodeToString(planHash[:]), "--input", "experiments/short-claim/probes-56b.json", "--legacy", "experiments/short-claim/probes-56.json", "--out", out}, &status); err != errSource || status.Len() != 0 {
		t.Fatal("current CLI accepted the historical source-bound plan")
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("historical-plan refusal reserved output")
	}
	// A test-only temporary replay plan changes only implementation digests.
	// Neither it nor its report claims to be the original frozen execution.
	replayRaw, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	replayRaw = append(replayRaw, '\n')
	replayPath := filepath.Join(t.TempDir(), "current-source-replay-plan.json")
	if err := os.WriteFile(replayPath, replayRaw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"--stage", "audit", "--plan", replayPath, "--plan-sha256", hash(replayRaw), "--input", "experiments/short-claim/probes-56b.json", "--legacy", "experiments/short-claim/probes-56.json", "--out", out}, &status); err != nil {
		t.Fatal(err)
	}
	gotRaw, err := os.ReadFile(filepath.Join(out, "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want, got report
	if json.Unmarshal(wantRaw, &want) != nil || json.Unmarshal(gotRaw, &got) != nil {
		t.Fatal("report decode failed")
	}
	if want.PlanSHA256 != hex.EncodeToString(planHash[:]) || !reflect.DeepEqual(want.ImplementationFiles, original.ImplementationFiles) || got.PlanSHA256 != hash(replayRaw) || !reflect.DeepEqual(got.ImplementationFiles, current.ImplementationFiles) {
		t.Fatal("historical/current report provenance differs from its own plan")
	}
	// Only these two explicit source provenance fields differ. All numerical
	// truth, full groups, transport pins, scope, counters and flags stay exact.
	got.PlanSHA256, got.ImplementationFiles = want.PlanSHA256, want.ImplementationFiles
	if !reflect.DeepEqual(want, got) {
		t.Fatal("frozen truth, source controls, label gates or group relations changed")
	}
}
