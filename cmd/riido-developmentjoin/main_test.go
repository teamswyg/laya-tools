package main

import (
	"context"
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/githubmeta"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) (sweaudit.TaskRole, filelabels.DevelopmentLabel, []byte) {
	t.Helper()
	m := sweaudit.TaskRole{Role: "train"}
	m.Task.ID = "task"
	m.Task.Repository = "example/repo"
	m.Task.BaseCommit = strings.Repeat("c", 40)
	l := filelabels.DevelopmentLabel{Label: filelabels.Label{ID: "task", Result: filelabels.Result{OldPaths: []string{"file.go"}}}}
	c := githubmeta.Catalog{SHA: strings.Repeat("a", 40), Tree: []sweaudit.TreeEntry{{Path: "file.go", Type: "blob", Mode: "100644", SHA: strings.Repeat("b", 40)}}}
	raw, e := json.Marshal(c)
	if e != nil {
		t.Fatal(e)
	}
	return m, l, raw
}
func TestJoinMissingCorruptAndSnapshot(t *testing.T) {
	m, l, raw := fixture(t)
	roots, cats := t.TempDir(), t.TempDir()
	c, v, e := join(m, l, roots, cats)
	if e != nil || c.Tasks != 1 || c.SupportedTargetTasks != 1 || c.UnavailableRoots != 1 || v.Status != "root_unavailable" {
		t.Fatalf("%+v %+v %v", c, v, e)
	}
	fetch := func(context.Context, string) ([]byte, error) { return raw, nil }
	if _, e := githubmeta.Root(context.Background(), m.Task.Repository, m.Task.BaseCommit, roots, fetch); e != nil {
		t.Fatal(e)
	}
	c, v, e = join(m, l, roots, cats)
	if e != nil || c.Roots != 1 || c.Catalogs != 0 || c.UnavailableCatalogs != 1 || v.Status != "catalog_unavailable" {
		t.Fatal(c, v, e)
	}
	if _, e := githubmeta.RecursiveCatalog(context.Background(), m.Task.Repository, strings.Repeat("a", 40), cats, fetch); e != nil {
		t.Fatal(e)
	}
	c, v, e = join(m, l, roots, cats)
	if e != nil || c.Catalogs != 1 || c.AllFileTargetsMapped != 1 || c.RegularFiles != 1 || v.RootSHA256 == "" || v.CatalogSHA256 == "" {
		t.Fatal(c, v, e)
	}
	other := m
	other.Task.BaseCommit = strings.Repeat("d", 40)
	c, v, e = join(other, l, roots, cats)
	if e != nil || c.UnavailableRoots != 1 || v.Status != "root_unavailable" {
		t.Fatal("used another snapshot", c, v, e)
	}
	for _, mode := range []string{"blank", "unsupported", "no-old"} {
		bad := l
		switch mode {
		case "blank":
			bad.ParseError = true
		case "unsupported":
			bad.Result.UnsupportedBlocks = 1
		case "no-old":
			bad.Result.OldPaths = nil
		}
		c, v, e = join(m, bad, roots, cats)
		if e != nil || c.AllFileTargetsMapped != 0 || c.MatchedTasks != 0 || c.Catalogs != 1 || v.Status != "unusable_target" {
			t.Fatal("unusable label counted as success", c, v, e)
		}
	}
	l.Result.OldPaths = []string{"absent"}
	c, v, e = join(m, l, roots, cats)
	if e != nil || c.TasksWithMissingPaths != 1 || c.MissingPaths != 1 || c.AllFileTargetsMapped != 0 {
		t.Fatal(c, v, e)
	}
	names, _ := filepath.Glob(filepath.Join(cats, "*.json.gz"))
	if len(names) != 1 {
		t.Fatal(names)
	}
	if e = os.WriteFile(names[0], []byte("corrupt"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, _, e = join(m, l, roots, cats); e == nil {
		t.Fatal("corrupt catalog counted unavailable")
	}
	names, _ = filepath.Glob(filepath.Join(roots, "*.json"))
	if e = os.WriteFile(names[0], []byte("corrupt"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, _, e = join(m, l, roots, cats); e == nil {
		t.Fatal("corrupt root counted unavailable")
	}
	m.Role = "final"
	if _, _, e = join(m, l, t.TempDir(), t.TempDir()); e == nil {
		t.Fatal("final admitted")
	}
}
func TestAggregateFullDenominators(t *testing.T) {
	rows := []coverage{{Role: "train", Repository: "a/r", Tasks: 1, SupportedTargetTasks: 1, UnavailableRoots: 1, UnavailableCatalogs: 1}, {Role: "train", Repository: "a/r", Tasks: 1, LabelParseFailures: 1, Catalogs: 1, Roots: 1}, {Role: "validation", Repository: "b/r", Tasks: 1, SupportedTargetTasks: 1, MatchedTasks: 1, AllFileTargetsMapped: 1, OldTargets: 1, RegularFiles: 1, Catalogs: 1, Roots: 1}}
	got := aggregate(rows)
	var total coverage
	for _, c := range got {
		add(&total, c)
	}
	if len(got) != 2 || total.Tasks != 3 || total.Catalogs != 2 || total.SupportedTargetTasks != 2 || total.AllFileTargetsMapped != 1 || total.LabelParseFailures != 1 || total.UnavailableCatalogs != 1 {
		t.Fatal(total)
	}
}
