package sweaudit

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

type RootObservation struct {
	Repository                 string
	ResponseBytes, RootEntries int
	LicenseCandidates          []string
	Available                  bool
}
type rootTree struct {
	SHA       string                                   `json:"sha"`
	Truncated bool                                     `json:"truncated"`
	Tree      []struct{ Path, Mode, Type, SHA string } `json:"tree"`
}

// PreflightSamples chooses by identity, never by source size or outcome.
func PreflightSamples(selected []Selection) ([]Selection, error) {
	if len(selected) == 0 || len(selected) > 2400 {
		return nil, fmt.Errorf("invalid selection size")
	}
	rows := slices.Clone(selected)
	for _, s := range rows {
		if !repoPattern.MatchString(s.Repository) || !commitPattern.MatchString(s.BaseCommit) || s.ID == "" || (s.Source != "full" && s.Source != "multilingual") {
			return nil, fmt.Errorf("invalid selection identity")
		}
	}
	slices.SortFunc(rows, func(a, b Selection) int {
		if a.Repository != b.Repository {
			return cmp.Compare(a.Repository, b.Repository)
		}
		return cmp.Compare(a.Source+"\x00"+a.ID, b.Source+"\x00"+b.ID)
	})
	out := rows[:0]
	for _, s := range rows {
		if len(out) == 0 || out[len(out)-1].Repository != s.Repository {
			out = append(out, s)
		}
	}
	return out, nil
}
func InspectRoot(repo string, b []byte) (RootObservation, error) {
	r := RootObservation{Repository: repo, ResponseBytes: len(b)}
	if !repoPattern.MatchString(repo) || len(b) == 0 || len(b) > 2<<20 {
		return r, fmt.Errorf("invalid root response size or repository")
	}
	var tree rootTree
	if e := json.Unmarshal(b, &tree); e != nil {
		return r, fmt.Errorf("invalid root JSON")
	}
	if tree.Truncated || !commitPattern.MatchString(tree.SHA) || len(tree.Tree) == 0 {
		return r, fmt.Errorf("incomplete root tree")
	}
	var paths []string
	for _, entry := range tree.Tree {
		if entry.Path == "" || strings.Contains(entry.Path, "/") || !commitPattern.MatchString(entry.SHA) {
			return r, fmt.Errorf("malformed root entry")
		}
		validMode := (entry.Type == "blob" && (entry.Mode == "100644" || entry.Mode == "100755" || entry.Mode == "120000")) || (entry.Type == "tree" && entry.Mode == "040000") || (entry.Type == "commit" && entry.Mode == "160000")
		if !validMode {
			return r, fmt.Errorf("invalid root object type or mode")
		}
		paths = append(paths, entry.Path)
		p := strings.ToUpper(entry.Path)
		if entry.Type == "blob" && (p == "LICENSE" || strings.HasPrefix(p, "LICENSE.") || strings.HasPrefix(p, "LICENSE-") || p == "COPYING" || strings.HasPrefix(p, "COPYING.") || p == "NOTICE" || strings.HasPrefix(p, "NOTICE.")) {
			r.LicenseCandidates = append(r.LicenseCandidates, entry.Path)
		}
	}
	slices.Sort(paths)
	if len(slices.Compact(paths)) != len(tree.Tree) {
		return r, fmt.Errorf("duplicate root entries")
	}
	slices.Sort(r.LicenseCandidates)
	r.RootEntries = len(tree.Tree)
	r.Available = true
	return r, nil
}
