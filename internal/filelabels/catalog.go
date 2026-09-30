package filelabels

import (
	"fmt"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/sweaudit"
)

// CatalogMatch distinguishes file-name candidates from directory/gitlink targets.
// It never follows links or opens source contents.
type CatalogMatch struct {
	Targets, RegularFiles, Symlinks, Submodules, Directories, Missing int
	MissingPaths                                                      []string `json:",omitempty"`
}

func MatchCatalog(paths []string, entries []sweaudit.TreeEntry) (CatalogMatch, error) {
	var r CatalogMatch
	for i, v := range entries {
		if i > 0 && entries[i-1].Path >= v.Path {
			return r, fmt.Errorf("unsorted or duplicate catalog path")
		}
		if _, e := cleanPath(v.Path); e != nil {
			return r, e
		}
	}
	for i, p := range paths {
		if _, e := cleanPath(p); e != nil {
			return CatalogMatch{}, e
		}
		if i > 0 && paths[i-1] >= p {
			return CatalogMatch{}, fmt.Errorf("unsorted or duplicate target")
		}
		r.Targets++
		j, ok := slices.BinarySearchFunc(entries, p, func(a sweaudit.TreeEntry, b string) int { return strings.Compare(a.Path, b) })
		if !ok {
			r.Missing++
			r.MissingPaths = append(r.MissingPaths, strings.Clone(p))
			continue
		}
		v := entries[j]
		switch {
		case v.Type == "blob" && (v.Mode == "100644" || v.Mode == "100755"):
			r.RegularFiles++
		case v.Type == "blob" && v.Mode == "120000":
			r.Symlinks++
		case v.Type == "commit" && v.Mode == "160000":
			r.Submodules++
		case v.Type == "tree" && v.Mode == "040000":
			r.Directories++
		default:
			return CatalogMatch{}, fmt.Errorf("invalid catalog target mode")
		}
	}
	return r, nil
}
