package main

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

func scoreAt(rank int) fileeval.Score {
	return fileeval.Score{Targets: 1, Mapped: 1, FirstRank: rank, Hit1: rank == 1, Hit10: rank <= 10, Hit100: rank <= 100, All10: rank <= 10, ReciprocalRank: 1 / float64(rank)}
}
func TestPairedHeadroomAndRepositoryAccounting(t *testing.T) {
	// A ten-page gain can coexist with a nine-page loss. Oracle headroom is
	// deliberately distinct from the actual always-helper net result.
	cases := []privateCase{
		{ID: "a", Repository: "repo-a", Baseline: scoreAt(201), Legacy: scoreAt(181), Helpers: [3]fileeval.Score{scoreAt(201), scoreAt(1), scoreAt(201)}},
		{ID: "b", Repository: "repo-b", Baseline: scoreAt(21), Legacy: scoreAt(41), Helpers: [3]fileeval.Score{scoreAt(21), scoreAt(201), scoreAt(21)}},
		{ID: "c", Repository: "repo-b", Baseline: scoreAt(1), Legacy: scoreAt(1), Helpers: [3]fileeval.Score{scoreAt(1), scoreAt(1), scoreAt(1)}},
	}
	cases[1].Fallbacks[2] = "explicit_path_comparison_budget"
	cases[0].HintCounts[1] = 2
	cases[0].Work[0] = hintsearch.HelperWork{IndexBuildAttempts: 4}
	cases[0].Work[1] = hintsearch.HelperWork{IndexBuildAttempts: 8, IndexSearchAttempts: 2}
	cases[0].Work[2] = hintsearch.HelperWork{AnchorComparisons: 9}
	before := append([]privateCase(nil), cases...)
	policies, repos, e := summarize(cases, [3]hintsearch.HelperWork{{IndexBuildAttempts: 4}, {IndexBuildAttempts: 8}, {}})
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(before, cases) {
		t.Fatal("summarizer changed case evidence")
	}
	base, legacy, fused, anchors := policies[0], policies[1], policies[3], policies[4]
	if base.Tasks != 3 || base.Pages != 14 || base.Hit10 != 1 || base.MacroHit10 != .25 {
		t.Fatalf("baseline aggregate: %+v", base)
	}
	if legacy.Pages != 14 || legacy.Wins != 1 || legacy.Losses != 1 || legacy.Ties != 1 || legacy.PositiveGain != 1 || legacy.NegativeGain != 1 {
		t.Fatalf("paired legacy aggregate: %+v", legacy)
	}
	if fused.Pages != 13 || fused.PositiveGain != 10 || fused.NegativeGain != 9 || !fused.NecessaryFivePercentHeadroom || math.Abs(fused.OraclePageReduction-10.0/14) > 1e-12 {
		t.Fatalf("optimistic headroom conflated with net gain: %+v", fused)
	}
	if fused.Hints != 2 || fused.IndexBuildAttempts != 8 || fused.IndexSearchAttempts != 2 || fused.MacroHit10 != .75 {
		t.Fatalf("fused work/macro counts: %+v", fused)
	}
	if anchors.Fallbacks != 1 || anchors.Tasks != 3 || anchors.Pages != 14 || anchors.AnchorComparisons != 9 || anchors.NecessaryFivePercentHeadroom {
		t.Fatalf("fallback row omitted: %+v", anchors)
	}
	if len(repos) != 10 {
		t.Fatalf("require every policy in both repositories, got %d", len(repos))
	}
	var repoBuilds int
	for _, r := range repos {
		if r.Policy == fused.Policy {
			repoBuilds += r.IndexBuildAttempts
			if r.Repository == "repo-a" && r.IndexBuildAttempts != 8 {
				t.Fatalf("first catalog construction omitted from repository: %+v", r)
			}
			if r.Repository == "repo-b" && (r.Pages != 12 || r.IndexBuildAttempts != 0) {
				t.Fatalf("repository harm or reused catalog accounting wrong: %+v", r)
			}
		}
	}
	if repoBuilds != fused.IndexBuildAttempts {
		t.Fatal("repository constructor counts do not sum to the actual global work")
	}
	if _, _, e := summarize(cases, [3]hintsearch.HelperWork{{IndexBuildAttempts: 4}, {IndexBuildAttempts: 7}, {}}); e == nil {
		t.Fatal("unaccounted constructor work accepted")
	}
}
func TestIncompleteTargetCannotEnterScreen(t *testing.T) {
	valid := scoreAt(1)
	c := privateCase{Baseline: valid, Legacy: valid, Helpers: [3]fileeval.Score{valid, valid, valid}}
	c.Helpers[1].Mapped = 0
	if _, _, e := summarize([]privateCase{c}, [3]hintsearch.HelperWork{}); e == nil {
		t.Fatal("unmapped helper target was accepted")
	}
}
func TestOutputAndPinnedInputBoundaries(t *testing.T) {
	for _, p := range []string{"", "/tmp/out", "results", ".cache/", ".cache/../public", ".cache/x/../out"} {
		if _, e := parse([]string{"--out", p}); e == nil {
			t.Fatalf("accepted output %q", p)
		}
	}
	if _, e := parse([]string{"--out", ".cache/preview48"}); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(t.TempDir(), "private-prompt-marker.json")
	data := []byte("{}\n")
	if e := os.WriteFile(file, data, 0600); e != nil {
		t.Fatal(e)
	}
	if e := pinned(file, digest(data), int64(len(data)), nil); e != nil {
		t.Fatal(e)
	}
	for _, limit := range []int64{1, 2} {
		if e := pinned(file, digest(data), limit, nil); e == nil {
			t.Fatal("oversized pinned input accepted")
		}
	}
	if e := pinned(file, strings.Repeat("0", 64), 64, nil); e == nil || strings.Contains(e.Error(), "private-prompt-marker") {
		t.Fatal("hash mismatch accepted or input name echoed")
	}
}

func TestPrivateOutputCannotFollowPublicSymlink(t *testing.T) {
	t.Chdir(t.TempDir())
	if e := os.Mkdir(".cache", 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Mkdir("public", 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink("../public", ".cache/shortcut"); e != nil {
		t.Fatal(e)
	}
	if e := privateAncestors(".cache/shortcut/run"); e == nil {
		t.Fatal("private records could escape through a symlink")
	}
	if e := privateAncestors(".cache/missing/run"); e == nil {
		t.Fatal("unverified ancestors were allowed")
	}
	if e := privateAncestors(".cache/new-run"); e != nil {
		t.Fatal(e)
	}
}
