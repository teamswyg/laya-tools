package hintsearch

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"unicode/utf8"
)

const MaxPathDocuments = 100000
const MaxPathCatalogBytes = 16 << 20

// NewPathIndex indexes complete file paths with global BM25 statistics. It is
// opt-in and does not change the small catalog, model or session contracts.
// Paths are text: no files are opened and no symlinks are followed. Callers
// remain responsible for permissions and historical catalog completeness.
func NewPathIndex(paths []string) (*Index, error) {
	return NewPathTextIndex(paths, paths)
}

// NewPathTextIndex keeps unique path identities while indexing aligned text,
// for example identifier-normalized paths. Equal text for different files is
// allowed. Both path and text catalogs are independently bounded to 16MiB.
func NewPathTextIndex(paths, texts []string) (*Index, error) {
	if len(paths) != len(texts) {
		return nil, fmt.Errorf("path and text counts differ")
	}
	if len(paths) == 0 || len(paths) > MaxPathDocuments {
		return nil, fmt.Errorf("require 1..%d file paths", MaxPathDocuments)
	}
	total := 0
	for _, p := range paths {
		if p == "" || len(p) > 4096 || !utf8.ValidString(p) || strings.ContainsRune(p, 0) || strings.HasPrefix(p, "/") || p == "." || p == ".." || strings.HasPrefix(p, "../") || path.Clean(p) != p {
			return nil, fmt.Errorf("require canonical relative UTF-8 file paths")
		}
		if len(p) > MaxPathCatalogBytes-total {
			return nil, fmt.Errorf("path catalog exceeds %d bytes", MaxPathCatalogBytes)
		}
		total += len(p)
	}
	unique := slices.Clone(paths)
	slices.Sort(unique)
	if len(slices.Compact(unique)) != len(paths) {
		return nil, fmt.Errorf("duplicate file path")
	}
	for _, text := range texts {
		if !utf8.ValidString(text) {
			return nil, fmt.Errorf("require UTF-8 path text")
		}
	}
	return newIndex(texts, MaxPathDocuments, MaxPathCatalogBytes)
}

// InterleavePathBaselineFirst preserves every candidate for large path indexes.
// With unit-cost accurate verification, baseline rank r is reached within
// min(n,2*r-1) checks. This is not a time/token bound or a claim of relevance.
// All input permutations are validated; neither input is modified.
func InterleavePathBaselineFirst(baseline, hints []int) ([]int, error) {
	return interleaveBounded(baseline, hints, true, MaxPathDocuments)
}
