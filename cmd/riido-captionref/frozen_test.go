// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/captionref"
)

// This post-observation regression follows the exact replay policy frozen in
// PLAN-REFERENCES-59 and oracle-review-59. It replays metadata once per test run;
// it is not another official collection, source observation or independent task.
func TestFrozenReference59ExactBodyAndCounters(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	raw, err := bounded(root, "experiments/short-claim/caption-coverage-59.json")
	if err != nil || len(raw) != 360591 || sha(raw) != "7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29" {
		t.Fatal("frozen result pin mismatch", err)
	}
	var saved record
	if err := decodeCanonical(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Schema != "riido-caption-reference-preparation-record-59-v1" || saved.State != "references_generated_content_review_pending" || saved.FailureCode != "" || saved.FailureCause != "" ||
		saved.SourceCommit != "75a9776230f7eaef3294247cc10bb7f13342f102" || saved.InputCommit != "69e9ddc51e218da029572e2bb463500dc36cd143" ||
		saved.PlanSHA256 != "a321d625e986ec7e621b827a0616b7492291f624bd395bc32b424a0ad1178819" || saved.BinarySHA256 != "03efc5808316128e0c336887907ee61d4aa9bab59cc764b8401e53d4786a21ec" || saved.BinaryBytes != 5202770 ||
		saved.GoVersion != "go1.27.1" || saved.CGOEnabled != "0" || !saved.TrimPath || saved.GOOS != "darwin" || saved.GOARCH != "arm64" ||
		saved.CPUThreads != 1 || saved.HeapSoftLimit != 256<<20 || saved.VerifiedGitBlobs != 21 || !saved.PreparationOnly || saved.ContentReview != "pending" || saved.LiteralReified || saved.TrainingReady {
		t.Fatal("historical envelope changed")
	}
	if !reflect.DeepEqual(sourceArtifacts(), saved.Sources) || len(saved.Support) != 9 || len(saved.Inputs) != 6 {
		t.Fatal("compiled source/support/input set changed")
	}
	if err := verifyGoClosure(root, saved.Sources); err != nil {
		t.Fatal("compiled Go source closure changed", err)
	}
	for _, pin := range append(append([]artifact{}, saved.Sources...), saved.Support...) {
		current, err := bounded(root, pin.Path)
		if err != nil || len(current) != pin.Bytes || sha(current) != pin.SHA256 {
			t.Fatal("source/support pin changed", err)
		}
	}
	inputs, pins, err := readInputs(root)
	if err != nil || !reflect.DeepEqual(pins, saved.Inputs) {
		t.Fatal("input set changed", err)
	}
	planRaw, err := bounded(root, planPath)
	if err != nil || len(planRaw) != 4681 || sha(planRaw) != saved.PlanSHA256 {
		t.Fatal("plan pin changed", err)
	}
	var p plan
	if err := decodeCanonical(planRaw, &p); err != nil {
		t.Fatal(err)
	}
	if err := validatePlan(p, newPlan(saved.SourceCommit, artifact{SHA256: saved.BinarySHA256, Bytes: saved.BinaryBytes}, saved.Inputs, saved.Sources, saved.Support)); err != nil {
		// newPlan uses the current platform. Reconcile only those environment
		// fields to the historical Darwin record before exact protocol checking.
		want := newPlan(saved.SourceCommit, artifact{SHA256: saved.BinarySHA256, Bytes: saved.BinaryBytes}, saved.Inputs, saved.Sources, saved.Support)
		want.GOOS, want.GOARCH = saved.GOOS, saved.GOARCH
		if err := validatePlan(p, want); err != nil {
			t.Fatal("frozen plan contract changed", err)
		}
	}
	var counters captionref.Counters
	got, err := captionref.Generate(inputs, &counters)
	if err != nil {
		t.Fatal(err)
	}
	gotRaw, err := canonical(got)
	if err != nil {
		t.Fatal(err)
	}
	wantRaw, err := canonical(saved.References)
	if err != nil || !bytes.Equal(gotRaw, wantRaw) || counters != saved.Counters {
		t.Fatal("reference body/counters changed", err)
	}
	if got.TrainingReady || got.ContentReview != "pending" || got.NewLabels != 0 || got.NewParents != 0 || got.Roles != 0 || got.Fits != 0 || got.SourceCalls != 0 || got.Rankings != 0 || got.ModelCalls != 0 {
		t.Fatal("reference replay acquired semantic or training eligibility")
	}
}
