package githubmeta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompleteSubtreeAssemblyAndCache(t *testing.T) {
	for _, trigger := range []string{"truncated", "limit"} {
		t.Run(trigger, func(t *testing.T) {
			root, child, blob := strings.Repeat("a", 40), strings.Repeat("b", 40), strings.Repeat("c", 40)
			cache, trees := t.TempDir(), t.TempDir()
			calls := 0
			fetch := func(_ context.Context, endpoint string) ([]byte, error) {
				calls++
				switch endpoint {
				case "repos/owner/repo/git/trees/" + root + "?recursive=1":
					if trigger == "limit" {
						return nil, ErrResponseLimit
					}
					return json.Marshal(Catalog{SHA: root, Truncated: true})
				case "repos/owner/repo/git/trees/" + root:
					return json.Marshal(Catalog{SHA: root, Tree: []sweaudit.TreeEntry{{Path: "src", Mode: "040000", Type: "tree", SHA: child}}})
				case "repos/owner/repo/git/trees/" + child + "?recursive=1":
					return json.Marshal(Catalog{SHA: child, Tree: []sweaudit.TreeEntry{{Path: "main.go", Mode: "100644", Type: "blob", SHA: blob}}})
				default:
					return nil, fmt.Errorf("unexpected endpoint")
				}
			}
			c, e := CompleteCatalog(context.Background(), "owner/repo", root, cache, trees, fetch)
			if e != nil || len(c.Tree) != 2 || c.Tree[1].Path != "src/main.go" || c.Truncated {
				t.Fatalf("%+v %v", c, e)
			}
			if calls != 3 {
				t.Fatalf("calls=%d", calls)
			}
			_, e = CompleteCatalog(context.Background(), "owner/repo", root, cache, trees, func(context.Context, string) ([]byte, error) { t.Fatal("cache fetched"); return nil, nil })
			if e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestCompleteDoesNotGuessOrPersistMissingChild(t *testing.T) {
	root, child := strings.Repeat("a", 40), strings.Repeat("b", 40)
	cache, trees := t.TempDir(), t.TempDir()
	calls := 0
	_, e := CompleteCatalog(context.Background(), "owner/repo", root, cache, trees, func(_ context.Context, endpoint string) ([]byte, error) {
		calls++
		if strings.HasSuffix(endpoint, root+"?recursive=1") {
			return nil, ErrResponseLimit
		}
		if strings.HasSuffix(endpoint, root) {
			return json.Marshal(Catalog{SHA: root, Tree: []sweaudit.TreeEntry{{Path: "src", Mode: "040000", Type: "tree", SHA: child}}})
		}
		return nil, fmt.Errorf("child unavailable")
	})
	if e == nil || calls != 3 {
		t.Fatalf("%v calls%d", e, calls)
	}
	files, _ := filepath.Glob(filepath.Join(cache, "*.json.gz"))
	if len(files) != 0 {
		t.Fatal("partial root saved")
	}
	calls = 0
	sentinel := errors.New("ordinary request failure")
	_, e = CompleteCatalog(context.Background(), "owner/repo", root, t.TempDir(), t.TempDir(), func(context.Context, string) ([]byte, error) { calls++; return nil, sentinel })
	if !errors.Is(e, sentinel) || calls != 1 {
		t.Fatal("ordinary failure expanded")
	}
}

func TestSubtreeCycleStopsAtDepthBound(t *testing.T) {
	root := strings.Repeat("a", 40)
	cache, trees := t.TempDir(), t.TempDir()
	_, e := CompleteCatalog(context.Background(), "owner/repo", root, cache, trees, func(_ context.Context, endpoint string) ([]byte, error) {
		if strings.Contains(endpoint, "?") {
			return nil, ErrResponseLimit
		}
		return json.Marshal(Catalog{SHA: root, Tree: []sweaudit.TreeEntry{{Path: "again", Mode: "040000", Type: "tree", SHA: root}}})
	})
	if e == nil || !strings.Contains(e.Error(), "depth bound") {
		t.Fatalf("unbounded cycle: %v", e)
	}
	files, _ := filepath.Glob(filepath.Join(cache, "*.json.gz"))
	if len(files) != 0 {
		t.Fatal("cyclic partial root saved")
	}
}
