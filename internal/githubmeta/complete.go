package githubmeta

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

// CompleteCatalog repairs only oversized/truncated recursive responses. It
// never persists a partial root or follows symlink/submodule targets. The
// fetch callback must apply per-request timeout, spacing and invocation budget.
func CompleteCatalog(ctx context.Context, repo, treeID, cache, treeCache string, fetch func(context.Context, string) ([]byte, error)) (Catalog, error) {
	first, e := RecursiveCatalog(ctx, repo, treeID, cache, fetch)
	if e == nil {
		return first, nil
	}
	if !errors.Is(e, ErrResponseLimit) && !errors.Is(e, ErrTruncatedTree) {
		return Catalog{}, e
	}
	result := Catalog{SHA: treeID}
	expanded, totalBytes := 0, 0
	appendEntry := func(prefix string, entry sweaudit.TreeEntry) error {
		if prefix != "" {
			entry.Path = prefix + "/" + entry.Path
		}
		if len(entry.Path) > 4096 || len(result.Tree) >= MaxTreeEntries || len(entry.Path) > MaxTreePathBytes-totalBytes {
			return fmt.Errorf("assembled catalog bound")
		}
		totalBytes += len(entry.Path)
		result.Tree = append(result.Tree, entry)
		return nil
	}
	var walk func(string, string, int, bool) error
	walk = func(id, prefix string, depth int, alreadyFailed bool) error {
		if depth > 64 {
			return fmt.Errorf("subtree depth bound")
		}
		if !alreadyFailed {
			c, e := RecursiveCatalog(ctx, repo, id, cache, fetch)
			if e == nil {
				for _, entry := range c.Tree {
					if e = appendEntry(prefix, entry); e != nil {
						return e
					}
				}
				return nil
			}
			if !errors.Is(e, ErrResponseLimit) && !errors.Is(e, ErrTruncatedTree) {
				return e
			}
		}
		expanded++
		if expanded > 1000 {
			return fmt.Errorf("subtree expansion bound")
		}
		entries, e := Tree(ctx, repo, id, treeCache, fetch)
		if e != nil {
			return e
		}
		for _, entry := range entries {
			if e = appendEntry(prefix, entry); e != nil {
				return e
			}
			if entry.Type == "tree" {
				child := entry.Path
				if prefix != "" {
					child = prefix + "/" + child
				}
				if e = walk(entry.SHA, child, depth+1, false); e != nil {
					return e
				}
			}
		}
		return nil
	}
	if e = walk(treeID, "", 0, true); e != nil {
		return Catalog{}, e
	}
	return saveCatalog(repo, treeID, cache, result)
}
