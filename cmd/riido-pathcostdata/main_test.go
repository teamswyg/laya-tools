package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/fileeval"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func label(paths ...string) filelabels.DevelopmentLabel {
	return filelabels.DevelopmentLabel{Label: filelabels.Label{ID: "task", Result: filelabels.Result{OldPaths: paths}}}
}
func TestPairedPageCostsAndUnusableTargets(t *testing.T) {
	paths := make([]string, 41)
	order := make([]int, len(paths))
	for i := range paths {
		paths[i] = fmt.Sprintf("file%02d.go", i)
		order[i] = i
	}
	helper := slices.Clone(order)
	helper[19], helper[20] = helper[20], helper[19]
	sel := fileeval.Selection{BaselineOrder: order, Order: helper}
	for _, tc := range []struct {
		target             string
		base, helper, gain int
	}{
		{paths[20], 2, 1, 1}, {paths[19], 1, 2, -1}, {paths[40], 3, 3, 0},
	} {
		v := example{ID: "task"}
		if err := score(&v, paths, sel, label(tc.target)); err != nil || v.Status != "scored" || v.BaselinePages != tc.base || v.HelperPages != tc.helper || v.Gain != tc.gain {
			t.Fatal("page boundary", v, err)
		}
	}
	for _, status := range []string{"parse_failure", "unsupported_target", "no_old_target", "unmapped_target"} {
		l := label(paths[0])
		switch status {
		case "parse_failure":
			l.ParseError = true
		case "unsupported_target":
			l.Result.UnsupportedBlocks = 1
		case "no_old_target":
			l.Result.OldPaths = nil
		case "unmapped_target":
			l.Result.OldPaths = []string{paths[0], "missing"}
		}
		v := example{ID: "task"}
		if err := score(&v, paths, sel, l); err != nil || v.Status != status || v.BaselinePages != 0 || v.HelperPages != 0 || v.Baseline.Targets != 0 {
			t.Fatal("invalid targets became a training outcome", v, err)
		}
	}
	bad := sel
	bad.Order = []int{0}
	if err := score(&example{ID: "task"}, paths, bad, label(paths[0])); err == nil {
		t.Fatal("broken ranking accepted")
	}
	if err := score(&example{ID: "other"}, paths, sel, label(paths[0])); err == nil {
		t.Fatal("wrong label accepted")
	}
}
func TestSummaryKeepsEntireDenominator(t *testing.T) {
	var rows []example
	for _, status := range []string{"checkpoint_unavailable", "parse_failure", "unsupported_target", "no_old_target", "unmapped_target", "ranking_error", "scored"} {
		rows = append(rows, example{ID: status, Role: "train", Repository: "a/r", Status: status})
	}
	rows = append(rows, example{ID: "v", Role: "validation", Repository: "b/r", Status: "scored", BaselinePages: 2, HelperPages: 1, Gain: 1, FeaturesAvailable: true, AuxiliaryRequested: true, AuxiliaryUsed: true, Work: fileeval.WorkStats{BaselineRankAttempts: 1, AuxiliaryBuildAttempts: 1, AuxiliaryRankAttempts: 1}})
	total, roles, repos, err := summarize(rows)
	if err != nil || total.Tasks != 8 || total.Unavailable != 1 || total.ParseFailures != 1 || total.Unsupported != 1 || total.NoOldTargets != 1 || total.Unmapped != 1 || total.RankingErrors != 1 || total.Scored != 2 || total.Wins != 1 || total.Ties != 1 || total.FeaturesAvailable != 1 || total.Work.BaselineRankAttempts != 1 || len(roles) != 2 || len(repos) != 2 || roles[0].Tasks != 7 || roles[1].Tasks != 1 {
		t.Fatal(total, roles, repos, err)
	}
	rows[0].Status = "unknown"
	if _, _, _, err = summarize(rows); err == nil {
		t.Fatal("unknown status silently omitted")
	}
}
func checkpointFixture(t *testing.T) (checkpoint, sweaudit.TaskRole, []byte) {
	t.Helper()
	m := sweaudit.TaskRole{Role: "train", Task: sweaudit.Selection{Source: "train", ID: "task", Repository: "example/repo", BaseCommit: strings.Repeat("c", 40)}}
	c := githubmeta.Catalog{SHA: strings.Repeat("a", 40), Tree: []sweaudit.TreeEntry{{Path: "file.go", Type: "blob", Mode: "100644", SHA: strings.Repeat("b", 40)}}}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	p := checkpoint{Role: m.Role, Repository: m.Task.Repository, ID: m.Task.ID, BaseCommit: m.Task.BaseCommit, RootSHA256: digest(raw), TreeID: c.SHA, CatalogSHA256: digest(raw), Status: "matched"}
	return p, m, raw
}
func TestCheckpointIntegrityAndRoles(t *testing.T) {
	p, m, _ := checkpointFixture(t)
	for _, mode := range []string{"valid", "hash", "duplicate", "final", "repo", "commit", "status", "digest", "missing"} {
		t.Run(mode, func(t *testing.T) {
			rows := []checkpoint{p}
			members := []sweaudit.TaskRole{m}
			switch mode {
			case "duplicate":
				rows = append(rows, p)
				members = append(members, m)
			case "final":
				rows[0].Role = "final"
				members[0].Role = "final"
			case "repo":
				rows[0].Repository = "other/repo"
			case "commit":
				rows[0].BaseCommit = strings.Repeat("d", 40)
			case "status":
				rows[0].Status = "unknown"
			case "digest":
				rows[0].CatalogSHA256 = ""
			case "missing":
				rows = nil
			}
			raw, err := json.Marshal(rows)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(t.TempDir(), "checkpoint.json")
			if err = os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			h := digest(raw)
			if mode == "hash" {
				h = strings.Repeat("0", 64)
			}
			got, err := readCheckpoint(name, h, members)
			if mode == "valid" {
				if err != nil || len(got) != 1 {
					t.Fatal(got, err)
				}
			} else if err == nil {
				t.Fatal("bad checkpoint accepted")
			}
		})
	}
}
func TestPinnedCacheMissingChangedAndCorruptAreFatal(t *testing.T) {
	p, _, raw := checkpointFixture(t)
	roots, cats := t.TempDir(), t.TempDir()
	if _, err := cachedPaths(p, roots, cats); err == nil {
		t.Fatal("missing required root")
	}
	fetch := func(context.Context, string) ([]byte, error) { return raw, nil }
	if _, err := githubmeta.Root(context.Background(), p.Repository, p.BaseCommit, roots, fetch); err != nil {
		t.Fatal(err)
	}
	if _, err := cachedPaths(p, roots, cats); err == nil {
		t.Fatal("missing required catalog")
	}
	if _, err := githubmeta.RecursiveCatalog(context.Background(), p.Repository, p.TreeID, cats, fetch); err != nil {
		t.Fatal(err)
	}
	paths, err := cachedPaths(p, roots, cats)
	if err != nil || !slices.Equal(paths, []string{"file.go"}) {
		t.Fatal(paths, err)
	}
	for _, mode := range []string{"root", "tree", "catalog", "commit"} {
		bad := p
		switch mode {
		case "root":
			bad.RootSHA256 = strings.Repeat("0", 64)
		case "tree":
			bad.TreeID = strings.Repeat("d", 40)
		case "catalog":
			bad.CatalogSHA256 = strings.Repeat("0", 64)
		case "commit":
			bad.BaseCommit = strings.Repeat("d", 40)
		}
		if _, err := cachedPaths(bad, roots, cats); err == nil {
			t.Fatal("changed checkpoint accepted", mode)
		}
	}
	files, err := filepath.Glob(filepath.Join(cats, "*.json.gz"))
	if err != nil || len(files) != 1 {
		t.Fatal(files, err)
	}
	if err = os.WriteFile(files[0], []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = cachedPaths(p, roots, cats); err == nil {
		t.Fatal("corrupt required catalog")
	}
}
