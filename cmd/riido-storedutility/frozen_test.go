// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/teamswyg/laya-tools/internal/storedaudit"
)

const (
	frozen58PlanSHA   = "c85c048831c6a3807656fccdba967251574b88b5764cbf708c6b38575d8e1893"
	frozen58ResultSHA = "48bbb4dd10ee5cf34f4edb7e746f0a0f7c85afd53b6513b62ea8442dac2768e6"
	frozen58Source    = "3c2bde94db7661bf5cdb268f0a2ede0f0b962a94"
	frozen58Input     = "bb2e6d2d137ce0626ccbfba1ee591a9814ec8153"
	frozen58Binary    = "fb7153b55b9351fcfb69ef881e1fb49a2ab69fc184adccebc5f6d2b6cc5510c8"
)

// This post-observation support test reproduces the frozen stage-58 stored
// utility record. Replays are CI regression checks, not additional official
// attempts, independent requests, source-truth observations or model evidence.
// The original plan and its 9 source / 13 support / 3 input pins stay unchanged.
func TestFrozenStoredUtilityRegression(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal("fixed public repository root unavailable")
	}
	defer root.Close()
	readPinned := func(path, wantSHA string, wantBytes int) []byte {
		t.Helper()
		raw, err := bounded(root, path)
		if err != nil || len(raw) != wantBytes || sha(raw) != wantSHA {
			t.Fatal("frozen public artifact bytes or digest changed")
		}
		return raw
	}
	historicalPath := func(path string) string {
		switch path {
		case "pkg/shortclaim/input.go":
			return "testdata/shortclaim-source-9d204/input.go.txt"
		case "pkg/shortclaim/baseline.go":
			return "testdata/shortclaim-source-9d204/baseline.go.txt"
		case "pkg/shortclaim/input_test.go":
			return "testdata/shortclaim-source-9d204/input-test.go.txt"
		case "pkg/shortclaim/baseline_test.go":
			return "testdata/shortclaim-source-9d204/baseline-test.go.txt"
		default:
			return path
		}
	}
	rawPlan := readPinned(planPath, frozen58PlanSHA, 5402)
	rawResult := readPinned("experiments/short-claim/results-58.json", frozen58ResultSHA, 184532)
	var p plan
	var saved record
	if decodeCanonical(rawPlan, &p) != nil || decodeCanonical(rawResult, &saved) != nil {
		t.Fatal("frozen record is not strict canonical JSON")
	}
	raw, inputs, err := rawInputs(root)
	if err != nil || len(inputs) != 3 || !reflect.DeepEqual(inputs, p.Inputs) || !reflect.DeepEqual(inputs, saved.Inputs) {
		t.Fatal("frozen raw input digest, size or ordering changed")
	}
	sources, err := sourceArtifacts(root)
	if err != nil || len(sources) != 9 || len(p.Sources) != len(sources) {
		t.Fatal("current compiled source bytes, digests or closure changed")
	}
	if err := verifyGoClosure(root, sources); err != nil {
		t.Fatal("current compiled Go source list changed")
	}
	for i, pin := range p.Sources {
		if sources[i].Path != pin.Path {
			t.Fatal("historical compiled source path order changed")
		}
		readPinned(historicalPath(pin.Path), pin.SHA256, pin.Bytes)
	}
	sup, err := support(root)
	if err != nil || len(sup) != 13 || len(p.Support) != len(sup) {
		t.Fatal("current support bytes or ordering changed")
	}
	for i, pin := range p.Support {
		if sup[i].Path != pin.Path {
			t.Fatal("historical support path order changed")
		}
		readPinned(historicalPath(pin.Path), pin.SHA256, pin.Bytes)
	}
	// These are historical provenance pins. The current CI test executable on
	// Linux or macOS is not required to have the original Darwin/arm64 binary's
	// digest, architecture or build recipe; this test does not run that binary.
	expected := newPlan(frozen58Source, artifact{SHA256: frozen58Binary, Bytes: 4872258}, inputs, p.Sources, p.Support)
	expected.GOOS, expected.GOARCH = "darwin", "arm64"
	if err := validatePlan(p, expected); err != nil {
		t.Fatal("historical frozen plan policy or provenance changed")
	}
	if saved.Schema != "riido-storedutility-official-record-58-v1" || saved.SourceCommit != frozen58Source || saved.InputCommit != frozen58Input || saved.PlanSHA256 != frozen58PlanSHA || saved.BinarySHA256 != frozen58Binary || saved.BinaryBytes != 4872258 || saved.GoVersion != "go1.27.1" || saved.GOOS != "darwin" || saved.GOARCH != "arm64" || saved.VerifiedGitBlobs != 26 {
		t.Fatal("historical official record provenance changed")
	}
	if !saved.Evaluation.UtilityEvaluated || saved.Evaluation.State != "complete" || saved.Evaluation.CompletedRows != 72 {
		t.Fatal("historical official evaluation was not complete")
	}
	dataset, err := storedaudit.Bind(raw[0], raw[1], raw[2])
	if err != nil {
		t.Fatal("frozen stored metadata binding failed")
	}
	actual, err := evaluate(dataset)
	if err != nil {
		t.Fatal("frozen utility regression replay failed", err)
	}
	if err := sameFrozen58Evaluation(actual, saved.Evaluation); err != nil {
		t.Fatal("frozen utility regression output changed", err)
	}
	t.Log("CI regression replay: Bind 1; evaluate 1; requests 72; baseline calls 72; ranking outputs 288; additional official attempts/independent samples/source API/model/fit/paid calls 0")
}

// Only active ranking scores get cross-platform floating-point tolerance. All
// remaining evaluation fields, including orders, fallbacks, costs, ratios,
// group metrics and unused slots, are compared exactly. Copies prevent this
// comparison from changing the original result or the reproduced evaluation.
func sameFrozen58Evaluation(actual, saved report) error {
	if len(actual.Rows) != len(saved.Rows) {
		return errors.New("frozen58_row_count_mismatch")
	}
	actual.Rows = slices.Clone(actual.Rows)
	saved.Rows = slices.Clone(saved.Rows)
	for i := range saved.Rows {
		for j := range saved.Rows[i].Rankings {
			a := &actual.Rows[i].Rankings[j]
			s := &saved.Rows[i].Rankings[j]
			if s.Count < 2 || s.Count > 4 || a.Count != s.Count {
				return errors.New("frozen58_ranking_count_mismatch")
			}
			for k := range s.Scores {
				if k >= s.Count {
					if a.Order[k] != 0 || s.Order[k] != 0 || a.Scores[k] != 0 || s.Scores[k] != 0 {
						return errors.New("frozen58_nonzero_unused_ranking_slot")
					}
					continue
				}
				if math.IsNaN(a.Scores[k]) || math.IsNaN(s.Scores[k]) || math.IsInf(a.Scores[k], 0) || math.IsInf(s.Scores[k], 0) || math.Abs(a.Scores[k]-s.Scores[k]) > 1e-12*(1+math.Abs(s.Scores[k])) {
					return errors.New("frozen58_score_mismatch")
				}
				a.Scores[k], s.Scores[k] = 0, 0
			}
		}
	}
	if !reflect.DeepEqual(actual, saved) {
		return errors.New("frozen58_exact_evaluation_mismatch")
	}
	return nil
}

func TestFrozen58ComparisonDoesNotRelaxOtherEvidence(t *testing.T) {
	// Tiny comparison controls exercise no Bind, Validate, feature or rankings.
	saved := report{Rows: []parentRanking{{CandidateCount: 2}}}
	for i := range saved.Rows[0].Rankings {
		saved.Rows[0].Rankings[i].Count = 2
		saved.Rows[0].Rankings[i].Order[1] = 1
		saved.Rows[0].Rankings[i].Scores[0] = 1
	}
	clone := func() report {
		out := saved
		out.Rows = slices.Clone(saved.Rows)
		return out
	}
	close := clone()
	close.Rows[0].Rankings[0].Scores[0] += 1e-13
	if err := sameFrozen58Evaluation(close, saved); err != nil {
		t.Fatal("bounded active-score rounding was rejected", err)
	}
	if close.Rows[0].Rankings[0].Scores[0] == 0 || saved.Rows[0].Rankings[0].Scores[0] != 1 {
		t.Fatal("comparison mutated its inputs")
	}
	for _, change := range []func(*report){
		func(r *report) { r.Rows[0].Rankings[0].Scores[0] += 1e-8 },
		func(r *report) { r.Rows[0].Rankings[0].Scores[0] = math.NaN() },
		func(r *report) { r.Rows[0].Rankings[0].Scores[0] = math.Inf(1) },
		func(r *report) { r.Rows[0].Rankings[0].Order[0] = 1 },
		func(r *report) { r.Rows[0].Rankings[0].FallbackReason = "different" },
		func(r *report) { r.Rows[0].Rankings[0].Scores[2] = 1e-30 },
		func(r *report) { r.Rows[0].Rankings[0].Order[2] = 1 },
		func(r *report) { r.Rows[0].Checks[0]++ },
		func(r *report) { r.PossibleRelativeGain += 1e-30 },
	} {
		changed := clone()
		change(&changed)
		if err := sameFrozen58Evaluation(changed, saved); err == nil {
			t.Fatal("score tolerance relaxed nonfinite, unused or non-score evidence")
		}
	}
}
