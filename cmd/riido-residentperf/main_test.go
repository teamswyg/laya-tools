package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func policyPlan() Plan {
	p := Plan{Schema: planSchema, State: planState, Policy: policy, SourceCommit: strings.Repeat("a", 40), Go: goVersion, OS: runtime.GOOS, Arch: runtime.GOARCH, ChildBinarySHA256: strings.Repeat("a", 64), DriverBinarySHA256: strings.Repeat("b", 64), BuildRecipeSHA256: strings.Repeat("c", 64), CorpusSHA256: strings.Repeat("d", 64), LegacyInputSHA256: legacySHA, TypedInputSHA256: typedSHA, Rows: plannedRows(), FirstRequests: 1, WarmupRequests: 20, TimedRequests: timedPerRow, CPUThreads: 1, HeapSoftLimitBytes: 256 << 20, ChildTimeoutSeconds: 15, GlobalTimeoutSeconds: 60, CleanupGraceMilliseconds: 1000, Retries: 0, InFlight: 1, PublicResultLimitBytes: resultLimit, StopOnFirstFailedRow: true}
	return p
}

func TestFrozenPolicyRejectsChangedBudgetsOrderAndIdentity(t *testing.T) {
	p := policyPlan()
	check := func(p Plan, want bool) {
		t.Helper()
		raw, e := encodePublic(p, maxPlanBytes)
		if e != nil {
			t.Fatal("test plan encoding failed")
		}
		_, e = decodePlan(raw, hashBytes(raw))
		if (e == nil) != want {
			t.Fatal("frozen plan acceptance differs")
		}
	}
	check(p, true)
	for _, change := range []func(*Plan){
		func(p *Plan) { p.CleanupGraceMilliseconds = 3000 },
		func(p *Plan) { p.Retries = 1 },
		func(p *Plan) { p.InFlight = 2 },
		func(p *Plan) { p.TimedRequests-- },
		func(p *Plan) { p.StopOnFirstFailedRow = false },
		func(p *Plan) { p.SourceCommit = "main" },
		func(p *Plan) { p.ChildBinarySHA256 = "latest" },
		func(p *Plan) { p.Rows[0], p.Rows[1] = p.Rows[1], p.Rows[0] },
	} {
		q := p
		change(&q)
		check(q, false)
	}
	raw, _ := encodePublic(p, maxPlanBytes)
	if _, e := decodePlan(raw, strings.Repeat("0", 64)); e == nil {
		t.Fatal("wrong frozen bytes accepted")
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		t.Fatal("test plan unavailable")
	}
	object["unexpected"] = json.RawMessage(`true`)
	extra, _ := json.Marshal(object)
	if _, e := decodePlan(extra, hashBytes(extra)); e == nil {
		t.Fatal("unknown plan field accepted")
	}
}

func TestRowsBalanceBeforeObservation(t *testing.T) {
	rows := plannedRows()
	for _, kind := range baselineKinds() {
		for _, work := range [2]string{"same", "distinct"} {
			for repeat := 1; repeat <= 3; repeat++ {
				n := 0
				for _, row := range rows {
					if row == (RowPlan{kind, work, repeat}) {
						n++
					}
				}
				if n != 1 {
					t.Fatal("planned matrix is not balanced")
				}
			}
		}
	}
	if rows[0] != (RowPlan{"fixed_order", "same", 1}) || rows[8].Baseline != "narrow_rule" {
		t.Fatal("precommitted counterbalance changed")
	}
}

func TestCanonicalPlanRefusesHiddenFixedArrayRows(t *testing.T) {
	raw, e := encodePublic(policyPlan(), maxPlanBytes)
	if e != nil {
		t.Fatal("test plan unavailable")
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil {
		t.Fatal("test plan decode failed")
	}
	var rows []RowPlan
	if json.Unmarshal(object["rows"], &rows) != nil || len(rows) != rowCount {
		t.Fatal("test rows unavailable")
	}
	for _, count := range [2]int{rowCount - 1, rowCount + 1} {
		changed := append([]RowPlan(nil), rows...)
		if count < rowCount {
			changed = changed[:count]
		} else {
			changed = append(changed, rows[0])
		}
		array, _ := json.MarshalIndent(changed, "  ", "  ")
		// Retain the Plan's original field order and indentation elsewhere.
		altered := bytes.Replace(raw, object["rows"], array, 1)
		if bytes.Equal(altered, raw) {
			t.Fatal("test row alteration failed")
		}
		_, e := decodePlan(altered, hashBytes(altered))
		if e == nil {
			t.Fatal("hidden or padded frozen plan rows accepted")
		}
	}
}

func TestArgumentErrorsAndOutputNeverExposePrivateInput(t *testing.T) {
	secret := "private-sensitive-value"
	for _, args := range [][]string{{"--unknown-" + secret}, {"--stage", "replay", "--out", secret}, {"--stage", "prepare", "--plan", secret, "--out", secret}, {"--stage", "prepare", "--out", secret, secret}} {
		var out bytes.Buffer
		e := run(args, &out)
		if e == nil || strings.Contains(e.Error(), secret) || strings.Contains(out.String(), secret) {
			t.Fatal("argument diagnostic leaked private input")
		}
	}
	var help bytes.Buffer
	if !errors.Is(run([]string{"--help"}, &help), flag.ErrHelp) || !strings.Contains(help.String(), "CPU-only") {
		t.Fatal("public help unavailable")
	}
}

func TestBoundedFilesAndExclusiveOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data")
	if e := writeNew(path, []byte("original")); e != nil {
		t.Fatal("initial output failed")
	}
	if e := writeNew(path, []byte("replacement")); e == nil {
		t.Fatal("existing output overwritten")
	}
	raw, e := readBounded(path, 8)
	if e != nil || string(raw) != "original" {
		t.Fatal("exclusive output changed")
	}
	if _, e := readBounded(path, 7); e == nil {
		t.Fatal("oversized input accepted")
	}
	if _, e := readBounded(dir, 32); e == nil {
		t.Fatal("directory accepted as input")
	}
	if _, e := readBounded(filepath.Join(dir, "missing-private"), 32); e == nil || strings.Contains(e.Error(), dir) {
		t.Fatal("missing input diagnostic leaked path")
	}
	if _, e := encodePublic(struct{ V float64 }{math.Inf(1)}, 64); e == nil {
		t.Fatal("nonfinite public metric accepted")
	}
	if _, e := encodePublic("1234", 6); e == nil {
		t.Fatal("public bound omitted final LF")
	}
	if e := os.Remove(path); e != nil {
		t.Fatal("test cleanup failed")
	}
}

// This is a CLI preparation control: no child, model, fit, or official replay.
func TestPreparationCLIEmitsOnlyBoundPublicArtifacts(t *testing.T) {
	t.Chdir(residentRepoRoot(t))
	dest := filepath.Join(t.TempDir(), "new-preparation")
	var out bytes.Buffer
	if e := run([]string{"--stage", "prepare", "--out", dest}, &out); e != nil {
		t.Fatal("packaged preparation control failed")
	}
	var summary struct {
		State          string `json:"state"`
		CorpusSHA      string `json:"corpus_sha256"`
		BuildRecipeSHA string `json:"build_recipe_sha256"`
		Child          int    `json:"measured_child_requests"`
		Models         int    `json:"model_calls"`
		Fits           int    `json:"fits"`
		Final          int    `json:"protected_final_reads"`
	}
	if json.Unmarshal(out.Bytes(), &summary) != nil || summary.State != "protocol_only_no_measurement" || summary.Child != 0 || summary.Models != 0 || summary.Fits != 0 || summary.Final != 0 {
		t.Fatal("preparation scope differs")
	}
	corpus, e := readBounded(filepath.Join(dest, "corpus.json"), maxCorpusBytes)
	if e != nil || hashBytes(corpus) != summary.CorpusSHA {
		t.Fatal("prepared wire bytes are not bound")
	}
	recipe, e := readBounded(filepath.Join(dest, "build-recipe.json"), maxRecipeBytes)
	if e != nil || hashBytes(recipe) != summary.BuildRecipeSHA {
		t.Fatal("prepared build recipe is not bound")
	}
	prepared, e := readBounded(filepath.Join(dest, "preparation.json"), maxPreparationBytes)
	if e != nil || !bytes.Equal(prepared, out.Bytes()) {
		t.Fatal("prepared summary differs from stdout")
	}
	entries, e := os.ReadDir(dest)
	if e != nil || len(entries) != 3 || len(corpus)+len(recipe)+len(prepared) > resultLimit {
		t.Fatal("prepared output scope or budget differs")
	}
	if run([]string{"--stage", "prepare", "--out", dest}, io.Discard) == nil {
		t.Fatal("existing preparation overwritten")
	}
}

func TestReplayDispatchFailurePreservesAllPlannedRows(t *testing.T) {
	p := policyPlan()
	// Two intentionally incomplete unit payloads refuse dispatch before Start.
	// This is not a resource run of the original public cohort.
	r := replayVerified(p, ReadyCorpus{Payloads: make([]ReadyPayload, 2)}, "never-started", strings.Repeat("a", 64))
	if r.PlannedRows != rowCount || r.PlannedRequests != rowCount*requestsPerRow || r.SourceCommit != p.SourceCommit || r.Models != 0 || r.Fits != 0 || r.ProtectedFinalReads != 0 {
		t.Fatal("failed replay lost frozen scope")
	}
	for i, row := range r.Rows {
		if row.Index != i || row.Plan != p.Rows[i] || row.Status != "not_started" || row.Started || !countsValid(row.Phases) || row.CPUPerRequestAvailable {
			t.Fatal("failed replay hid unattempted rows")
		}
		if i == 0 && row.Error != "prepared_dispatch_invalid" || i > 0 && row.Error != "prior_row_failure_no_retry" {
			t.Fatal("failed replay resumed or discarded failure reason")
		}
	}
}
