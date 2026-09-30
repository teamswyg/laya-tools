package githubmeta

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

func catalogFixture() Catalog {
	return Catalog{SHA: strings.Repeat("a", 40), Tree: []sweaudit.TreeEntry{{Path: "src", Mode: "040000", Type: "tree", SHA: strings.Repeat("b", 40)}, {Path: "src/main.go", Mode: "100644", Type: "blob", SHA: strings.Repeat("c", 40)}}}
}
func TestCatalogValidation(t *testing.T) {
	valid := catalogFixture()
	b, _ := json.Marshal(valid)
	if _, e := decodeCatalog(b, valid.SHA); e != nil {
		t.Fatal(e)
	}
	for _, change := range []func(*Catalog){func(c *Catalog) { c.Truncated = true }, func(c *Catalog) { c.SHA = strings.Repeat("d", 40) }, func(c *Catalog) { c.Tree = append(c.Tree, c.Tree[0]) }, func(c *Catalog) { c.Tree[1].Path = "../escape" }, func(c *Catalog) { c.Tree[1].Path = "other/main.go" }, func(c *Catalog) { c.Tree[1].Mode = "040000" }, func(c *Catalog) { c.Tree[1].Path = strings.Repeat("x", 4097) }} {
		c := catalogFixture()
		change(&c)
		b, _ := json.Marshal(c)
		if _, e := decodeCatalog(b, valid.SHA); e == nil {
			t.Fatal("accepted malformed catalog")
		}
	}
	missingFlag := strings.Replace(string(b), `"truncated":false,`, "", 1)
	if _, e := decodeCatalog([]byte(missingFlag), valid.SHA); e == nil {
		t.Fatal("missing truncation flag")
	}
	if _, e := decodeCatalog([]byte(strings.Repeat("x", MaxCatalogResponseBytes+1)), valid.SHA); e == nil {
		t.Fatal("oversized response")
	}
}
func TestCatalogCacheIntegrity(t *testing.T) {
	cache := t.TempDir()
	c := catalogFixture()
	b, _ := json.Marshal(c)
	calls := 0
	fetch := func(context.Context, string) ([]byte, error) { calls++; return b, nil }
	for range 2 {
		got, e := RecursiveCatalog(context.Background(), "owner/repo", c.SHA, cache, fetch)
		if e != nil || len(got.Tree) != 2 {
			t.Fatalf("%v", e)
		}
	}
	if calls != 1 {
		t.Fatal("cache miss")
	}
	names, e := filepath.Glob(filepath.Join(cache, "*.json.gz"))
	if e != nil || len(names) != 1 {
		t.Fatal("cache shape")
	}
	if e = os.WriteFile(names[0], []byte("corrupt"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = RecursiveCatalog(context.Background(), "owner/repo", c.SHA, cache, fetch); e == nil {
		t.Fatal("accepted corruption")
	}
	if calls != 1 {
		t.Fatal("corruption silently refetched")
	}
	badCache := t.TempDir()
	c.Truncated = true
	bad, _ := json.Marshal(c)
	if _, e = RecursiveCatalog(context.Background(), "owner/repo", c.SHA, badCache, func(context.Context, string) ([]byte, error) { return bad, nil }); e == nil {
		t.Fatal("accepted truncated result")
	}
	entries, _ := os.ReadDir(badCache)
	if len(entries) != 0 {
		t.Fatal("cached failed result")
	}
}
