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
	planPath := "experiments/short-claim/execution-plan-56b.json"
	planRaw, err := os.ReadFile(filepath.Join(root, planPath))
	if err != nil {
		t.Fatal(err)
	}
	planHash := sha256.Sum256(planRaw)
	t.Chdir(root)
	out := filepath.Join(t.TempDir(), "replay")
	var status bytes.Buffer
	if err := run([]string{"--stage", "audit", "--plan", planPath, "--plan-sha256", hex.EncodeToString(planHash[:]), "--input", "experiments/short-claim/probes-56b.json", "--legacy", "experiments/short-claim/probes-56.json", "--out", out}, &status); err != nil {
		t.Fatal(err)
	}
	gotRaw, err := os.ReadFile(filepath.Join(out, "results.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want, got any
	if json.Unmarshal(wantRaw, &want) != nil || json.Unmarshal(gotRaw, &got) != nil || !reflect.DeepEqual(want, got) {
		t.Fatal("frozen truth, source controls, label gates or group relations changed")
	}
}
