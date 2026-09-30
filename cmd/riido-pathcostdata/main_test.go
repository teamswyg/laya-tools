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
	"github.com/teamswyg/laya-tools/internal/trainingdata"
)

func TestConfigRequiresAllOverridesTogether(t *testing.T) {
	names := [...]string{"plan", "plan-sha256", "checkpoint", "checkpoint-sha256"}
	values := [...]string{"owned-plan.json", strings.Repeat("a", 64), "owned-checkpoint.json", strings.Repeat("b", 64)}
	for mask := 0; mask < 1<<len(names); mask++ {
		t.Run(fmt.Sprintf("combination_%04b", mask), func(t *testing.T) {
			args := []string{"--out", "private-output"}
			for i, name := range names {
				if mask&(1<<i) != 0 {
					args = append(args, "--"+name, values[i])
				}
			}
			got, err := parseConfig(args)
			if mask != 0 && mask != 15 {
				if err == nil {
					t.Fatal("partial override accepted", got)
				}
				return
			}
			if err != nil || got.out != "private-output" {
				t.Fatal(got, err)
			}
			if mask == 0 {
				if got.planPath != "experiments/path-cost-data/plan-43.json" || got.planSHA != planSHA || got.checkpointPath != ".cache/development-join-40-resume2/evidence.json" || got.checkpointSHA != checkpointSHA {
					t.Fatal("43 defaults changed", got)
				}
			} else if got.planPath != values[0] || got.planSHA != values[1] || got.checkpointPath != values[2] || got.checkpointSHA != values[3] {
				t.Fatal("explicit overrides ignored", got)
			}
		})
	}
	for _, bad := range []struct {
		name  string
		index int
		value string
	}{
		{"empty_plan", 0, ""}, {"empty_checkpoint", 2, ""},
		{"empty_plan_hash", 1, ""}, {"bad_plan_hash", 1, "not-hex"},
		{"uppercase_checkpoint_hash", 3, strings.Repeat("A", 64)}, {"short_checkpoint_hash", 3, strings.Repeat("a", 62)},
	} {
		t.Run(bad.name, func(t *testing.T) {
			vs := values
			vs[bad.index] = bad.value
			args := []string{"--out", "private-output"}
			for i, name := range names {
				args = append(args, "--"+name, vs[i])
			}
			if _, err := parseConfig(args); err == nil {
				t.Fatal("invalid explicit override accepted")
			}
		})
	}
	for _, args := range [][]string{nil, {"--out", ""}, {"--out", "private-output", "unexpected"}, {"--out", "private-output", "--unknown"}} {
		if _, err := parseConfig(args); err == nil {
			t.Fatal("invalid arguments accepted", args)
		}
	}
}

func TestPlanPinsFixedScopeAndReportsActualHash(t *testing.T) {
	ready := false
	base := plan{Schema: "riido-path-cost-data-plan-v1", MembershipSHA256: trainingdata.MembershipSHA256, QueryProjectionSHA256: trainingdata.QueryProjectionSHA256, PatchProjectionSHA256: patchSHA, AvailabilityCheckpoint: "owned checkpoint", QueryReader: "fixed reader", FeatureOrder: "fixed features", Costs: "fixed costs", Denominators: "all development members", Integrity: "offline only", Validation: "owned tests", ProductionReady: &ready}
	for _, mode := range []string{"valid", "hash", "schema", "membership", "query", "patch", "production", "missing_production", "missing_policy", "unknown", "trailing", "oversized"} {
		t.Run(mode, func(t *testing.T) {
			p := base
			switch mode {
			case "schema":
				p.Schema = "different"
			case "membership":
				p.MembershipSHA256 = strings.Repeat("a", 64)
			case "query":
				p.QueryProjectionSHA256 = strings.Repeat("a", 64)
			case "patch":
				p.PatchProjectionSHA256 = strings.Repeat("a", 64)
			case "production":
				v := true
				p.ProductionReady = &v
			case "missing_production":
				p.ProductionReady = nil
			case "missing_policy":
				p.Costs = " "
			case "oversized":
				p.Validation = strings.Repeat("v", (1<<20)+1)
			}
			raw, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "unknown" {
				raw = append(append(slices.Clone(raw[:len(raw)-1]), []byte(",\"fit\":true")...), '}')
			} else if mode == "trailing" {
				raw = append(raw, []byte(" {}")...)
			}
			name := filepath.Join(t.TempDir(), "plan.json")
			if err = os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			h := digest(raw)
			if mode == "hash" {
				h = strings.Repeat("0", 64)
			}
			actual, err := readPlan(name, h)
			if mode == "valid" {
				if err != nil || actual != digest(raw) {
					t.Fatal("verified plan digest missing", actual, err)
				}
			} else if err == nil || actual != "" {
				t.Fatal("bad plan accepted", actual, err)
			}
		})
	}
	if got, err := readPlan(filepath.Join("..", "..", "experiments", "path-cost-data", "plan-43.json"), planSHA); err != nil || got != planSHA {
		t.Fatal("original pinned plan unsupported", got, err)
	}
}

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
			got, actual, err := readCheckpoint(name, h, members)
			if mode == "valid" {
				if err != nil || len(got) != 1 || actual != digest(raw) {
					t.Fatal(got, actual, err)
				}
			} else if err == nil || actual != "" {
				t.Fatal("bad checkpoint accepted")
			}
		})
	}
}

func TestCatalogUnavailableStillRequiresPinnedRoot(t *testing.T) {
	p, _, raw := checkpointFixture(t)
	p.Status, p.CatalogSHA256 = "catalog_unavailable", ""
	roots := t.TempDir()
	// Using a regular file instead of a catalog directory proves that root-only
	// verification does not consult, create or repair a catalog cache.
	noCatalog := filepath.Join(t.TempDir(), "do-not-access")
	if err := os.WriteFile(noCatalog, []byte("not a catalog directory"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := cachedPaths(p, roots, noCatalog); err == nil {
		t.Fatal("catalog-unavailable row skipped required missing root")
	}
	if _, err := githubmeta.Root(context.Background(), p.Repository, p.BaseCommit, roots, func(context.Context, string) ([]byte, error) { return raw, nil }); err != nil {
		t.Fatal(err)
	}
	if paths, err := cachedPaths(p, roots, noCatalog); err != nil || len(paths) != 0 {
		t.Fatal("root-only verification consulted a catalog", paths, err)
	}
	for _, mode := range []string{"digest", "tree", "commit"} {
		bad := p
		switch mode {
		case "digest":
			bad.RootSHA256 = strings.Repeat("0", 64)
		case "tree":
			bad.TreeID = strings.Repeat("d", 40)
		case "commit":
			bad.BaseCommit = strings.Repeat("d", 40)
		}
		if _, err := cachedPaths(bad, roots, noCatalog); err == nil {
			t.Fatal("changed root-only reference accepted", mode)
		}
	}
	files, err := filepath.Glob(filepath.Join(roots, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatal(files, err)
	}
	if err = os.WriteFile(files[0], []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = cachedPaths(p, roots, noCatalog); err == nil {
		t.Fatal("corrupt root-only reference accepted")
	}
	p.Status, p.RootSHA256, p.TreeID = "root_unavailable", "", ""
	if paths, err := cachedPaths(p, roots, noCatalog); err != nil || len(paths) != 0 {
		t.Fatal("checkpoint-unavailable root consulted later cache", paths, err)
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
