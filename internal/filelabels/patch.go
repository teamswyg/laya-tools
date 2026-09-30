// Package filelabels extracts evaluation-only paths; it never applies patches.
package filelabels

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxPatchBytes = 2 << 20

type Result struct {
	OldPaths                                            []string
	Blocks, NewFiles, UnsupportedBlocks, DuplicatePaths int
}

func cleanPath(s string) (string, error) {
	if s == "" || len(s) > 4096 || !utf8.ValidString(s) || strings.ContainsRune(s, 0) || strings.HasPrefix(s, "/") || s == "." || s == ".." || strings.HasPrefix(s, "../") || path.Clean(s) != s {
		return "", fmt.Errorf("invalid relative path")
	}
	return s, nil
}
func decodedPath(s string) (string, error) {
	if strings.HasPrefix(s, `"`) {
		v, e := strconv.Unquote(s)
		if e != nil {
			return "", fmt.Errorf("invalid quoted path")
		}
		s = v
	}
	return cleanPath(s)
}
func fileHeader(s, prefix string) (string, bool, error) {
	s, _, _ = strings.Cut(s, "\t")
	if strings.HasPrefix(s, `"`) {
		v, e := strconv.Unquote(s)
		if e != nil {
			return "", false, e
		}
		s = v
	}
	if s == "/dev/null" {
		return "", true, nil
	}
	if !strings.HasPrefix(s, prefix) {
		return "", false, fmt.Errorf("file path prefix")
	}
	p, e := cleanPath(strings.TrimPrefix(s, prefix))
	return p, false, e
}

// headerOld accepts only unambiguous two-field Git headers. Text diff headers
// and rename metadata can still resolve paths containing unquoted spaces.
func headerOld(s string) (string, error) {
	var fields []string
	for s != "" {
		s = strings.TrimLeft(s, " ")
		if s == "" {
			break
		}
		if s[0] == '"' {
			end := 1
			for ; end < len(s); end++ {
				if s[end] == '\\' {
					end++
					continue
				}
				if s[end] == '"' {
					break
				}
			}
			if end >= len(s) {
				return "", fmt.Errorf("invalid quoted header")
			}
			v, e := strconv.Unquote(s[:end+1])
			if e != nil {
				return "", e
			}
			fields = append(fields, v)
			s = s[end+1:]
		} else {
			v, rest, _ := strings.Cut(s, " ")
			fields = append(fields, v)
			s = rest
		}
	}
	if len(fields) != 2 || !strings.HasPrefix(fields[0], "a/") || !strings.HasPrefix(fields[1], "b/") {
		return "", fmt.Errorf("ambiguous diff header")
	}
	return cleanPath(strings.TrimPrefix(fields[0], "a/"))
}
func Parse(patch string) (Result, error) {
	var r Result
	if len(patch) == 0 || len(patch) > MaxPatchBytes || !utf8.ValidString(patch) {
		return r, fmt.Errorf("invalid patch input")
	}
	type block struct {
		header, old, rename                                string
		active, body, binary, hasOld, hasNew, newFile, bad bool
	}
	var b block
	finish := func() {
		if !b.active {
			return
		}
		r.Blocks++
		if b.bad {
			r.UnsupportedBlocks++
			return
		}
		if b.newFile {
			r.NewFiles++
			return
		}
		var p string
		var e error
		switch {
		case b.hasOld && b.hasNew:
			p = b.old
		case b.hasOld || b.hasNew:
			r.UnsupportedBlocks++
			return
		case b.rename != "":
			p = b.rename
		case b.body && !b.binary:
			r.UnsupportedBlocks++
			return
		default:
			p, e = headerOld(b.header)
		}
		if e != nil || p == "" {
			r.UnsupportedBlocks++
			return
		}
		r.OldPaths = append(r.OldPaths, p)
	}
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			finish()
			b = block{active: true, header: strings.TrimPrefix(line, "diff --git ")}
			continue
		}
		if !b.active {
			if strings.TrimSpace(line) != "" {
				return Result{}, fmt.Errorf("unexpected patch preamble")
			}
			continue
		}
		if b.body {
			continue
		}
		if strings.HasPrefix(line, "@@") || line == "GIT binary patch" {
			b.body = true
			b.binary = line == "GIT binary patch"
			continue
		}
		switch {
		case strings.HasPrefix(line, "new file mode "):
			b.newFile = true
		case strings.HasPrefix(line, "rename from "):
			var e error
			b.rename, e = decodedPath(strings.TrimPrefix(line, "rename from "))
			b.bad = b.bad || e != nil
		case strings.HasPrefix(line, "--- "):
			if b.hasOld {
				b.bad = true
			}
			b.hasOld = true
			p, added, e := fileHeader(strings.TrimPrefix(line, "--- "), "a/")
			b.old = p
			b.newFile = b.newFile || added
			b.bad = b.bad || e != nil
		case strings.HasPrefix(line, "+++ "):
			if b.hasNew {
				b.bad = true
			}
			b.hasNew = true
			_, _, e := fileHeader(strings.TrimPrefix(line, "+++ "), "b/")
			b.bad = b.bad || e != nil
		}
	}
	finish()
	if r.Blocks == 0 {
		return r, fmt.Errorf("no Git diff blocks")
	}
	slices.Sort(r.OldPaths)
	before := len(r.OldPaths)
	r.OldPaths = slices.Compact(r.OldPaths)
	r.DuplicatePaths = before - len(r.OldPaths)
	return r, nil
}
