package githubmeta

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

const MaxCatalogResponseBytes = 8 << 20
const MaxStoredCatalogBytes = 32 << 20
const MaxTreeEntries = 100000
const MaxTreePathBytes = 16 << 20

var ErrTruncatedTree = errors.New("recursive tree is truncated")

type Catalog struct {
	SHA       string               `json:"sha"`
	Truncated bool                 `json:"truncated"`
	Tree      []sweaudit.TreeEntry `json:"tree"`
}

func FetchCatalog(ctx context.Context, endpoint string) ([]byte, error) {
	return fetchBounded(ctx, endpoint, MaxCatalogResponseBytes)
}

func decodeCatalog(b []byte, expected string) (Catalog, error) {
	if len(b) > MaxCatalogResponseBytes {
		return Catalog{}, ErrResponseLimit
	}
	return decodeCatalogLimit(b, expected, MaxCatalogResponseBytes)
}

func decodeCatalogLimit(b []byte, expected string, limit int) (Catalog, error) {
	var raw struct {
		SHA       string
		Truncated *bool
		Tree      []sweaudit.TreeEntry
	}
	if len(b) > limit || !revisionPattern.MatchString(expected) {
		return Catalog{}, fmt.Errorf("catalog response bound or identity")
	}
	if e := json.Unmarshal(b, &raw); e != nil {
		return Catalog{}, fmt.Errorf("invalid catalog JSON")
	}
	if raw.SHA != expected || raw.Truncated == nil {
		return Catalog{}, fmt.Errorf("incomplete or mismatched catalog")
	}
	if *raw.Truncated {
		return Catalog{}, ErrTruncatedTree
	}
	if len(raw.Tree) == 0 || len(raw.Tree) > MaxTreeEntries {
		return Catalog{}, fmt.Errorf("catalog entry bound")
	}
	total := 0
	for _, v := range raw.Tree {
		if v.Path == "" || len(v.Path) > 4096 || !utf8.ValidString(v.Path) || strings.ContainsRune(v.Path, 0) || strings.HasPrefix(v.Path, "/") || v.Path == "." || v.Path == ".." || strings.HasPrefix(v.Path, "../") || path.Clean(v.Path) != v.Path || !revisionPattern.MatchString(v.SHA) {
			return Catalog{}, fmt.Errorf("invalid catalog entry")
		}
		valid := (v.Type == "blob" && (v.Mode == "100644" || v.Mode == "100755" || v.Mode == "120000")) || (v.Type == "tree" && v.Mode == "040000") || (v.Type == "commit" && v.Mode == "160000")
		if !valid {
			return Catalog{}, fmt.Errorf("invalid catalog mode")
		}
		total += len(v.Path)
		if total > MaxTreePathBytes {
			return Catalog{}, fmt.Errorf("catalog path budget")
		}
	}
	slices.SortFunc(raw.Tree, func(a, b sweaudit.TreeEntry) int { return strings.Compare(a.Path, b.Path) })
	for i, v := range raw.Tree {
		if i > 0 && raw.Tree[i-1].Path == v.Path {
			return Catalog{}, fmt.Errorf("duplicate catalog path")
		}
		parent := path.Dir(v.Path)
		if parent != "." {
			j, ok := slices.BinarySearchFunc(raw.Tree, parent, func(a sweaudit.TreeEntry, b string) int { return strings.Compare(a.Path, b) })
			if !ok || raw.Tree[j].Type != "tree" {
				return Catalog{}, fmt.Errorf("missing parent directory")
			}
		}
	}
	return Catalog{SHA: raw.SHA, Tree: raw.Tree}, nil
}

// RecursiveCatalog retains complete validated path metadata, not file contents.
// Cache entries are normalized compressed JSON; revalidate on every read.
func RecursiveCatalog(ctx context.Context, repo, treeID, cache string, fetch func(context.Context, string) ([]byte, error)) (Catalog, error) {
	if !repositoryPattern.MatchString(repo) || !revisionPattern.MatchString(treeID) {
		return Catalog{}, fmt.Errorf("invalid catalog identity")
	}
	key := sha256.Sum256([]byte(repo + "\x00" + treeID))
	name := filepath.Join(cache, hex.EncodeToString(key[:])+".json.gz")
	f, e := os.Open(name)
	if e == nil {
		defer f.Close()
		z, e := gzip.NewReader(f)
		if e != nil {
			return Catalog{}, fmt.Errorf("invalid catalog cache")
		}
		defer z.Close()
		b, e := io.ReadAll(io.LimitReader(z, MaxStoredCatalogBytes+1))
		if e != nil {
			return Catalog{}, fmt.Errorf("invalid compressed catalog")
		}
		c, e := decodeCatalogLimit(b, treeID, MaxStoredCatalogBytes)
		if e != nil {
			// Corrupt cached data must not trigger a network repair silently.
			return Catalog{}, fmt.Errorf("invalid cached catalog: %v", e)
		}
		return c, nil
	}
	if !os.IsNotExist(e) {
		return Catalog{}, e
	}
	b, e := fetch(ctx, "repos/"+repo+"/git/trees/"+treeID+"?recursive=1")
	if e != nil {
		return Catalog{}, e
	}
	c, e := decodeCatalog(b, treeID)
	if e != nil {
		return Catalog{}, e
	}
	return saveCatalog(repo, treeID, cache, c)
}

func saveCatalog(repo, treeID, cache string, c Catalog) (Catalog, error) {
	b, e := json.Marshal(c)
	if e != nil {
		return Catalog{}, e
	}
	c, e = decodeCatalogLimit(b, treeID, MaxStoredCatalogBytes)
	if e != nil {
		return Catalog{}, e
	}
	key := sha256.Sum256([]byte(repo + "\x00" + treeID))
	name := filepath.Join(cache, hex.EncodeToString(key[:])+".json.gz")
	if e = os.MkdirAll(cache, 0700); e != nil {
		return Catalog{}, e
	}
	f, e := os.CreateTemp(cache, ".catalog-*.tmp")
	if e != nil {
		return Catalog{}, e
	}
	temp := f.Name()
	defer os.Remove(temp)
	z := gzip.NewWriter(f)
	if _, e = z.Write(b); e != nil {
		z.Close()
		f.Close()
		return Catalog{}, e
	}
	if e = z.Close(); e != nil {
		f.Close()
		return Catalog{}, e
	}
	if e = f.Close(); e != nil {
		return Catalog{}, e
	}
	if e = os.Rename(temp, name); e != nil {
		return Catalog{}, e
	}
	return c, nil
}
