package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/publicbehavior"
)

// Added after collection. Replay on Linux/macOS is a regression test, not a
// claim that the historical Darwin binary ran on another platform or new data.
func TestPublishedFiniteAuditReplay(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	raw, err := bounded(root, "experiments/public-behavior/results-57.json")
	if err != nil {
		t.Fatal(err)
	}
	if sha(raw) != "850c60266eceb613565390430adca12cea9b1a56207b803b06cb681f563978c5" {
		t.Fatal("published observation bytes changed")
	}
	var r record
	if err := decodeCanonical(raw, &r); err != nil {
		t.Fatal(err)
	}
	if r.SourceCommit != "9e99914f6d1e35aa9d97413f80fc6e03768e97ce" || r.InputCommit != "1462c705087f24666049f8bd31d317490c05dbeb" || r.GoVersion != "go1.27.1" || r.GOOS != "darwin" || r.GOARCH != "arm64" || r.VerifiedGitBlobs != 28 {
		t.Fatal("historical provenance changed")
	}
	input, err := bounded(root, probesPath)
	if err != nil || sha(input) != r.ProbesSHA256 {
		t.Fatal("probes linkage changed")
	}
	var probes publicbehavior.Probes
	if err := decodeCanonical(input, &probes); err != nil {
		t.Fatal(err)
	}
	planRaw, err := bounded(root, planPath)
	if err != nil || sha(planRaw) != r.PlanSHA256 {
		t.Fatal("plan linkage changed")
	}
	var p plan
	if err := decodeCanonical(planRaw, &p); err != nil {
		t.Fatal(err)
	}
	if p.BinaryBytes != 5096818 || p.BinarySHA256 != r.BinarySHA256 || p.BinarySHA256 != "09e5f6079b8a436c432d313a884376a5bde2267e7afdde88c5d422f1777cc5d7" || p.SourceCommit != r.SourceCommit || p.ProbesSHA256 != r.ProbesSHA256 || p.ProbesBytes != len(input) {
		t.Fatal("binary/input plan linkage changed")
	}
	if !reflect.DeepEqual(sourceArtifacts(), p.Sources) {
		t.Fatal("compiled source evidence changed")
	}
	sup, err := support(root)
	if err != nil || !reflect.DeepEqual(sup, p.Support) {
		t.Fatal("pre-collection support evidence changed")
	}
	if err := verifyGoClosure(root, p.Sources); err != nil {
		t.Fatal(err)
	}
	got, err := publicbehavior.Audit(probes)
	if err != nil {
		t.Fatal(err)
	}
	a, err := publicbehavior.Canonical(got)
	if err != nil {
		t.Fatal(err)
	}
	b, err := publicbehavior.Canonical(r.Audit)
	if err != nil || !bytes.Equal(a, b) {
		t.Fatal("finite replay disagreed with preserved original observations")
	}
	if got.CompletedObservations != 62 || got.PrimaryAPICalls != 90 || got.MatchedExpectations != 59 || got.MismatchedExpectations != 3 || got.UnknownObservations != 0 || got.TrainingReady || got.ModelCalls != 0 || got.Fits != 0 || got.IndependentFinalTasks != 0 {
		t.Fatal("denominator or eligibility drift")
	}
	var mismatches []string
	getters := 0
	for _, row := range got.Rows {
		getters += row.Observation.UpstreamGetterCalls
		if !row.Matches {
			mismatches = append(mismatches, row.ID)
		}
	}
	if getters != 224 || !reflect.DeepEqual(mismatches, []string{"p08-both-overflow-normative-discrepancy", "p09-reverse-overflow-discrepancy", "b05-success-shortcircuit-malformed-later-alt"}) {
		t.Fatal("disagreement evidence changed")
	}
}

func TestPublishedAuditRejectsIncorrectOracleAndInputBindings(t *testing.T) {
	raw, err := os.ReadFile("../../" + probesPath)
	if err != nil {
		t.Fatal(err)
	}
	var p publicbehavior.Probes
	if err := decodeCanonical(raw, &p); err != nil {
		t.Fatal(err)
	}
	p.Vectors[0].SourceHypothesis[0].Field = "versions.0.comparison"
	if publicbehavior.ValidateProbes(p) == nil {
		t.Fatal("unrelated nested field accepted")
	}
}
