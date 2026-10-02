package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/teamswyg/laya-tools/internal/finiteproperty"
)

// Added after collection: this regression replay is not a pre-observation pin,
// a new official attempt or an additional independent development request.
func TestPublished56ePreservesFrozenInputsAndFiniteObservation(t *testing.T) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	read := func(path, sha string) []byte {
		t.Helper()
		raw, err := bounded(root, path)
		if err != nil {
			t.Fatal(err)
		}
		if finiteproperty.SHA256(raw) != sha {
			t.Fatalf("changed public record %s", path)
		}
		return raw
	}
	dataRaw := read("experiments/short-claim/probes-56e.json", "86556bd96599ae013441f199f90c8cd619a011d52212d1ccbfdbe98b6f6a1ec5")
	planRaw := read("experiments/short-claim/execution-plan-56e.json", "6a21ff6da7772476e070509b45bf9506b95f50473841bf3183d17a2fca7ded69")
	resultRaw := read("experiments/short-claim/results-56e.json", "4a50160b99ea059a8de6e7be92ba22b5017f94dec2d3ca2fea644bda759a1be8")
	var d finiteproperty.Dataset
	var p plan
	var r record
	for _, entry := range []struct {
		raw    []byte
		target any
	}{{dataRaw, &d}, {planRaw, &p}, {resultRaw, &r}} {
		if err := decode(entry.raw, entry.target); err != nil {
			t.Fatal(err)
		}
	}
	compiled, err := sources()
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled) != len(p.Sources) || len(p.Support) != 13 {
		t.Fatal("source closure changed")
	}
	// Preserve the source-bound historical plan, verifying archived bytes for
	// the two deliberately optimized files and unchanged bytes for the rest.
	for i, original := range p.Sources {
		if compiled[i].Path != original.Path {
			t.Fatal("compiled source path changed")
		}
		path := original.Path
		switch path {
		case "pkg/shortclaim/input.go":
			path = "testdata/shortclaim-source-9d204/input.go.txt"
		case "pkg/shortclaim/baseline.go":
			path = "testdata/shortclaim-source-9d204/baseline.go.txt"
		default:
			if compiled[i].SHA256 != original.SHA256 {
				t.Fatal("unchanged historical source differs from compiled source")
			}
		}
		read(path, original.SHA256)
	}
	// A current compiled manifest is checked separately and never replaces the
	// frozen public plan. The production verifier must still refuse old pins.
	if err := verifyFiles(root, compiled); err != nil {
		t.Fatal(err)
	}
	if err := verifyFiles(root, p.Sources); err == nil {
		t.Fatal("current source verifier accepted historical runtime pins")
	}
	if err := verifyFiles(root, p.Support); err != nil {
		t.Fatal(err)
	}
	if err := verifyDirectories(root); err != nil {
		t.Fatal(err)
	}
	if p.SourceCommit != "1f9631c2819abf24ca57e6317ed48b8397be0881" || r.SourceCommit != p.SourceCommit || r.InputCommit != "ca5ff83b847e984dab628be536cc2462ddff8221" || r.PlanSHA256 != finiteproperty.SHA256(planRaw) || p.DatasetSHA256 != finiteproperty.SHA256(dataRaw) || p.DatasetBytes != len(dataRaw) || p.BinarySHA256 != "b5b5f97acb48dd3c1d6f472effee6e56d6d094e7e2b25db4ab0c68d06c792612" || p.BinaryBytes != 8824130 || r.BinarySHA256 != p.BinarySHA256 || r.GitBlobsVerified != 36 || p.Attempts != 1 || p.Retries != 0 || r.State != "complete" {
		t.Fatal("freeze/execution linkage changed")
	}
	// The preserved Darwin binary/platform is historical evidence. Linux CI
	// replays the pure-Go semantics without claiming that binary ran on Linux.
	if r.GoVersion != "go1.27.1" || r.GOOS != "darwin" || r.GOARCH != "arm64" || p.GoVersion != r.GoVersion || p.GOOS != r.GOOS || p.GOARCH != r.GOARCH {
		t.Fatal("historical platform changed")
	}
	got, err := finiteproperty.Audit(d)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, r.Observation) {
		t.Fatal("observations/labels changed")
	}
	if got.ActualCandidateCalls != 36 || got.ObserverDispatchAttempts != 36 || got.CompletedObservations != 36 || got.ObserverFailures != 0 || got.UniqueSourceInputs != 36 || got.CandidateLabels != 36 || got.NewIndependentParents != 0 || got.PreflightTextNormalizations != 192 || got.RankingFeatureExtractions != 0 || got.RankingRuns != 0 || got.Fits != 0 || got.ModelCalls != 0 || got.ProtectedFinalRead || got.TrainingReady {
		t.Fatal("scope/counters changed")
	}
	answerable, noAnswer := 0, 0
	for _, row := range got.Rows {
		switch row.Status {
		case "finite_answerable":
			answerable++
		case "finite_no_answer":
			noAnswer++
		default:
			t.Fatal("unknown row status")
		}
	}
	if answerable != 11 || noAnswer != 1 {
		t.Fatal("finite cohort changed")
	}
}
