package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func checkSelection(s selection) error {
	if s.Schema != "riido-hf33-directory-selection-v1" || s.PublicationCommit != "2a1c2224e89f60eef9e381f98c0117206f60e0d2" || len(s.Targets) != 2 || !s.CanonicalScienceExcluded || !s.ModelsTokensRawObservationsExcluded {
		return errContract
	}
	inodes := map[string]bool{}
	for i, t := range s.Targets {
		if t.ID != targetNames[i] || len(t.Files) != fileCounts[i] || len(t.Directories) != directoryCounts[i] {
			return errContract
		}
		rm, e := modeOf(t.Root.Mode)
		if e != nil || rm&0700 != 0700 || t.Root.Device == "" || t.Root.Inode == "" {
			return errContract
		}
		seen := map[string]bool{}
		last := ""
		for _, d := range t.Directories {
			if !cleanRelative(d.Relative) || d.Relative <= last || seen[d.Relative] {
				return errContract
			}
			dm, e := modeOf(d.Mode)
			if e != nil || dm&0700 != 0700 {
				return errContract
			}
			seen[d.Relative] = true
			last = d.Relative
		}
		dirs := map[string]bool{}
		for _, d := range t.Directories {
			dirs[d.Relative] = true
		}
		total := int64(0)
		last = ""
		for _, f := range t.Files {
			if !cleanRelative(f.Relative) || f.Relative <= last || seen[f.Relative] || f.Relative == markerName || f.Bytes < 0 || f.Bytes > restoreCap || !validSHA(f.SHA256) || f.Nlink != 1 || f.Device == "" || f.Inode == "" {
				return errContract
			}
			if _, e := modeOf(f.Mode); e != nil {
				return e
			}
			key := f.Device + ":" + f.Inode
			if inodes[key] {
				return errContract
			}
			inodes[key] = true
			parent := filepath.ToSlash(filepath.Dir(f.Relative))
			if parent != "." && !dirs[parent] {
				return errContract
			}
			seen[f.Relative] = true
			last = f.Relative
			total += f.Bytes
		}
		for _, d := range t.Directories {
			parent := filepath.ToSlash(filepath.Dir(d.Relative))
			if parent != "." && !dirs[parent] {
				return errContract
			}
		}
		if total != payloadSizes[i] || total > restoreCap {
			return errContract
		}
	}
	return nil
}

func checkPaths(c config, configPath string, a admission) error {
	if c.Schema != "riido-hf33-directory-archive-config-v1" || !cleanAbsolute(c.Root) || filepath.Base(c.Root) != "riido-next60-remaining-three-root-outside-review-v1" || !freshBase(c.OutputBase) || len(c.RestoreBases) != 2 {
		return errContract
	}
	if e := noSymlinkAbsolute(c.Root); e != nil {
		return e
	}
	ri, e := os.Lstat(c.Root)
	if e != nil || !ri.IsDir() || !plainPermissions(ri) {
		return errArtifact
	}
	names := map[string]bool{}
	for _, s := range targetNames {
		names[strings.ToLower(s)] = true
	}
	for _, s := range append([]string{c.OutputBase}, c.RestoreBases...) {
		if !freshBase(s) || names[strings.ToLower(s)] {
			return errContract
		}
		names[strings.ToLower(s)] = true
		_, e := os.Lstat(filepath.Join(c.Root, s))
		if !os.IsNotExist(e) {
			return errArtifact
		}
	}
	pins := append([]pin{c.Selection, c.Admission, c.PriorAccount, a.Census}, a.HelperSources...)
	pins = append(pins, pin{Path: configPath})
	// Existing helpers, config, frozen selection and census must not be deleted
	// with either source, or be placed inside a future output/restore directory.
	for _, p := range pins {
		if !cleanAbsolute(p.Path) {
			return errContract
		}
		for _, base := range append(append([]string{}, targetNames[:]...), append([]string{c.OutputBase}, c.RestoreBases...)...) {
			q := filepath.Join(c.Root, base)
			if p.Path == q || strings.HasPrefix(p.Path, q+string(filepath.Separator)) {
				return errContract
			}
		}
	}
	// Existing pin ancestors are compared physically as well as lexically,
	// including case aliases on case-insensitive filesystems.
	for _, p := range pins {
		if e := noSymlinkAbsolute(p.Path); e != nil {
			return e
		}
		for q := p.Path; ; q = filepath.Dir(q) {
			fi, e := os.Lstat(q)
			if e != nil {
				return e
			}
			for _, name := range targetNames {
				si, e := os.Lstat(filepath.Join(c.Root, name))
				if e != nil || os.SameFile(fi, si) {
					return errArtifact
				}
			}
			if filepath.Dir(q) == q {
				break
			}
		}
	}
	for _, name := range targetNames {
		p := filepath.Join(c.Root, name)
		if e := noSymlinkAbsolute(p); e != nil {
			return e
		}
		fi, e := os.Lstat(p)
		if e != nil || !fi.IsDir() || os.SameFile(ri, fi) {
			return errArtifact
		}
	}
	first, e := os.Lstat(filepath.Join(c.Root, targetNames[0]))
	if e != nil {
		return e
	}
	second, e := os.Lstat(filepath.Join(c.Root, targetNames[1]))
	if e != nil {
		return e
	}
	if os.SameFile(first, second) {
		return errArtifact
	}
	return nil
}

func checkAdmission(c config, a admission) error {
	if a.Schema != "riido-hf33-directory-archive-admission-v1" || a.Selection != c.Selection || a.PriorAccount != c.PriorAccount || a.CapBytes != globalCap || a.CurrentRetainedBytes <= 0 || a.CurrentRetainedBytes > globalCap || len(a.ArchiveCaps) != 2 || a.RestoreCap != restoreCap || a.ControlCap != controlCap || a.InflateCap != inflateCap || !a.WritersQuiescent || !a.FixedPublicCopiesOnly || len(a.HelperSources) != 6 {
		return errContract
	}
	for _, n := range a.ArchiveCaps {
		if n <= 0 || n > maxArchiveCap {
			return errContract
		}
	}
	// Logical-path accounting: do not hash-deduplicate, retroactively reset a
	// budget, or count expected compression as reclaim. Phase two conservatively
	// retains phase-one archive cap and the entire shared new-control reserve.
	if e := checkPeak(a.CurrentRetainedBytes, a.ArchiveCaps); e != nil {
		return e
	}
	helperNames := []string{"main.go", "types.go", "contract.go", "archive.go", "run.go", "go.mod"}
	helperParent := filepath.Dir(a.HelperSources[0].Path)
	for i, p := range a.HelperSources {
		if filepath.Base(p.Path) != helperNames[i] || filepath.Dir(p.Path) != helperParent {
			return errContract
		}
	}
	seen := map[string]bool{}
	for _, p := range append([]pin{a.Census, a.PriorAccount}, a.HelperSources...) {
		if seen[p.Path] {
			return errContract
		}
		seen[p.Path] = true
		if _, e := readPin(p, metadataCap); e != nil {
			return e
		}
	}
	return nil
}

func inventorySHA(t target) string {
	b, _ := json.Marshal(t)
	return digest(b)
}

func checkTree(root string, t target, restored bool, allowMarker bool) error {
	if e := noSymlinkAbsolute(root); e != nil {
		return e
	}
	ri, e := os.Lstat(root)
	if e != nil || !ri.IsDir() || !plainPermissions(ri) {
		return errArtifact
	}
	rm, e := modeOf(t.Root.Mode)
	if e != nil || ri.Mode().Perm() != rm {
		return errChanged
	}
	if !restored {
		rootMeta, e := infoMeta(ri)
		if e != nil || !sameRootMeta(rootMeta, t.Root, allowMarker) {
			return errChanged
		}
	}
	files := map[string]entry{}
	dirs := map[string]directory{}
	for _, f := range t.Files {
		files[f.Relative] = f
	}
	for _, d := range t.Directories {
		dirs[d.Relative] = d
	}
	seenFiles, seenDirs := 0, 0
	seenMarker := false
	inodes := map[string]bool{}
	e = filepath.WalkDir(root, func(p string, de os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == root {
			return nil
		}
		rel, er := filepath.Rel(root, p)
		if er != nil {
			return er
		}
		rel = filepath.ToSlash(rel)
		fi, er := os.Lstat(p)
		if er != nil || fi.Mode()&os.ModeSymlink != 0 || !plainPermissions(fi) {
			return errArtifact
		}
		if allowMarker && rel == markerName {
			seenMarker = true
			if !fi.Mode().IsRegular() {
				return errArtifact
			}
			return nil
		}
		if fi.IsDir() {
			want, ok := dirs[rel]
			if !ok {
				return errArtifact
			}
			dm, er := modeOf(want.Mode)
			if er != nil || fi.Mode().Perm() != dm || (!restored && !sameMeta(fi, want.meta, true)) {
				return errChanged
			}
			seenDirs++
			return nil
		}
		want, ok := files[rel]
		if !ok || !fi.Mode().IsRegular() {
			return errArtifact
		}
		got, before, er := hashRegular(p, restoreCap)
		if er != nil || got.Bytes != want.Bytes || got.SHA256 != want.SHA256 {
			return errChanged
		}
		m, er := infoMeta(before)
		fm, mer := modeOf(want.Mode)
		if er != nil || mer != nil || before.Mode().Perm() != fm || m.Nlink != 1 || (!restored && !sameMeta(before, want.meta, true)) {
			return errChanged
		}
		key := m.Device + ":" + m.Inode
		if inodes[key] {
			return errArtifact
		}
		inodes[key] = true
		seenFiles++
		return nil
	})
	if e != nil {
		return e
	}
	if seenFiles != len(t.Files) || seenDirs != len(t.Directories) || (allowMarker && !seenMarker) {
		return errArtifact
	}
	return nil
}

func orderedDirectories(t target, descending bool) []directory {
	ds := append([]directory(nil), t.Directories...)
	sort.Slice(ds, func(i, j int) bool {
		a, b := strings.Count(ds[i].Relative, "/"), strings.Count(ds[j].Relative, "/")
		if a == b {
			if descending {
				return ds[i].Relative > ds[j].Relative
			}
			return ds[i].Relative < ds[j].Relative
		}
		if descending {
			return a > b
		}
		return a < b
	})
	return ds
}

func checkPeak(current int64, caps []int64) error {
	if current <= 0 || current > globalCap || len(caps) != 2 {
		return errContract
	}
	for _, n := range caps {
		if n <= 0 || n > maxArchiveCap {
			return errContract
		}
	}
	p1 := current + caps[0] + restoreCap + controlCap
	p2 := current - payloadSizes[0] + caps[0] + caps[1] + restoreCap + controlCap
	if p1 > globalCap || p2 > globalCap {
		return errCap
	}
	return nil
}

// Adding the one admitted marker changes APFS directory nlink by one;
// Linux directory nlink can stay equal. All other root identity facts remain exact.
func sameRootMeta(got, want meta, marker bool) bool {
	if got.Mode != want.Mode || got.Device != want.Device || got.Inode != want.Inode {
		return false
	}
	if !marker {
		return got.Nlink == want.Nlink && got.MtimeNS == want.MtimeNS
	}
	return got.Nlink == want.Nlink || (want.Nlink != ^uint64(0) && got.Nlink == want.Nlink+1)
}
