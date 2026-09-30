package hintsearch

import (
	"fmt"
	"path"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

const (
	NormalizedPositive64    = "normalized-positive64"
	JoinedFieldsRRF64       = "joined-fields-rrf64"
	ExplicitPath64          = "explicit-path64"
	PathHelperHintLimit     = 64
	MaxPathHelperTokens     = 262144
	MaxPathHelperVocabulary = 131072
	// The exact-anchor scan declines before matching if its complete Cartesian
	// comparison count exceeds this resource bound. No anchors are truncated.
	MaxPathHelperAnchorComparisons = 8_000_000
)

// HelperWork counts actual helper field construction attempts and BM25 search
// attempts, including failures. A construction attempt includes its preflight;
// it need not allocate a BM25 index. AnchorComparisons counts path/token pairs,
// not bytes inspected, elapsed time, downstream calls or candidate verification.
type HelperWork struct {
	IndexBuildAttempts, IndexSearchAttempts, AnchorComparisons int
}

// HelperResult owns Order independently of inputs and other results. HintCount
// counts proposed hints, including any already-first baseline candidate. A valid
// Rank call attempts every fixed helper, including no-evidence and fallback cases.
// A fallback preserves the complete supplied baseline. Work contains query work
// only; count BuildWork once per constructed PathHelpers to avoid double counting.
type HelperResult struct {
	ID        string
	Order     []int
	HintCount int
	Attempted bool
	Fallback  string
	Work      HelperWork
}

// PathHelpers is an opt-in, label-blind implementation of the three fixed
// experiment-48 path helpers. It opens no files and accepts no outcomes, labels,
// repository identities or model IDs. Indexes and cloned path identities are
// private and immutable after construction. Rank uses request-owned buffers and
// can run concurrently; callers must not mutate query/baseline inputs during it.
// No existing search, model or default routing behavior is changed.
type PathHelpers struct {
	paths                   []string
	normalized, full, names *Index
	normalizedFallback      string
	joinedFallback          string
	buildWork               [3]HelperWork
}

// PathHelperIDs returns an owned fixed-order array for BuildWork and Rank.
func PathHelperIDs() [3]string {
	return [3]string{NormalizedPositive64, JoinedFieldsRRF64, ExplicitPath64}
}

// NewPathHelpers validates the same unique canonical relative UTF-8 path
// contract as NewPathIndex, then snapshots caller identities. Each text field
// checks byte, lexical-token and vocabulary bounds before BM25 allocation.
// Auxiliary construction errors are retained per variant for Rank fallback;
// invalid shared path identities instead return an error here.
func NewPathHelpers(paths []string) (*PathHelpers, error) {
	if err := validateHelperPaths(paths); err != nil {
		return nil, err
	}
	h := &PathHelpers{paths: make([]string, len(paths))}
	for i, p := range paths {
		h.paths[i] = strings.Clone(p)
	}
	h.buildWork[0].IndexBuildAttempts++
	h.normalized, h.normalizedFallback = newHelperField(h.paths, false, false)
	if h.normalizedFallback != "" {
		h.normalizedFallback = "normalized_" + h.normalizedFallback
	}
	h.buildWork[1].IndexBuildAttempts++
	h.full, h.joinedFallback = newHelperField(h.paths, false, true)
	if h.joinedFallback != "" {
		h.joinedFallback = "joined_full_" + h.joinedFallback
		return h, nil
	}
	h.buildWork[1].IndexBuildAttempts++
	h.names, h.joinedFallback = newHelperField(h.paths, true, true)
	if h.joinedFallback != "" {
		h.full = nil
		h.joinedFallback = "joined_basename_" + h.joinedFallback
	}
	return h, nil
}

// BuildWork reports constructor work in PathHelperIDs order. This value is
// immutable and independently owned; repeated Rank calls perform no rebuilding.
func (h *PathHelpers) BuildWork() [3]HelperWork {
	if h == nil {
		return [3]HelperWork{}
	}
	return h.buildWork
}

func validateHelperPaths(paths []string) error {
	if len(paths) == 0 || len(paths) > MaxPathDocuments {
		return fmt.Errorf("require 1..%d file paths", MaxPathDocuments)
	}
	total := 0
	for _, p := range paths {
		if p == "" || len(p) > 4096 || !utf8.ValidString(p) || strings.ContainsRune(p, 0) || strings.HasPrefix(p, "/") || p == "." || p == ".." || strings.HasPrefix(p, "../") || path.Clean(p) != p {
			return fmt.Errorf("require canonical relative UTF-8 file paths")
		}
		if len(p) > MaxPathCatalogBytes-total {
			return fmt.Errorf("path catalog exceeds %d bytes", MaxPathCatalogBytes)
		}
		total += len(p)
	}
	unique := slices.Clone(paths)
	slices.Sort(unique)
	if len(slices.Compact(unique)) != len(paths) {
		return fmt.Errorf("duplicate file path")
	}
	return nil
}

func joinedHelperText(s string) string {
	return strings.ToLower(s) + " " + lexicalhint.NormalizeText(s)
}

func newHelperField(paths []string, basename, joined bool) (*Index, string) {
	texts := make([]string, len(paths))
	totalBytes, totalTokens := 0, 0
	vocabulary := make(map[string]struct{})
	for i, p := range paths {
		if basename {
			p = path.Base(p)
		}
		var text string
		if joined {
			text = joinedHelperText(p)
		} else {
			text = lexicalhint.NormalizeText(p)
		}
		if len(text) > MaxPathCatalogBytes-totalBytes {
			return nil, "catalog_text_budget"
		}
		totalBytes += len(text)
		ts := words(text)
		if len(ts) > MaxPathHelperTokens-totalTokens {
			return nil, "catalog_token_budget"
		}
		totalTokens += len(ts)
		for _, t := range ts {
			if _, found := vocabulary[t]; found {
				continue
			}
			if len(vocabulary) == MaxPathHelperVocabulary {
				return nil, "catalog_vocabulary_budget"
			}
			vocabulary[t] = struct{}{}
		}
		texts[i] = text
	}
	// No preflight map is retained by the immutable index or query path.
	vocabulary = nil
	idx, err := newIndex(texts, MaxPathDocuments, MaxPathCatalogBytes)
	if err != nil {
		return nil, "catalog_index_error"
	}
	return idx, ""
}

// Rank compares all three fixed variants against an already-produced baseline
// permutation. It accepts complete nonempty UTF-8 queries up to 128KiB, never
// truncating them. Invalid shared inputs return an error before helper searching;
// per-variant auxiliary failures return owned, complete baseline fallback orders.
// Every successful order preserves baseline[0] and the min(n, 2*r-1) rank bound.
func (h *PathHelpers) Rank(query string, baseline []int) ([]HelperResult, error) {
	if h == nil || len(h.paths) == 0 || len(baseline) != len(h.paths) {
		return nil, fmt.Errorf("invalid path helper baseline")
	}
	if !utf8.ValidString(query) || len(query) > MaxLongQueryBytes || strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("require nonempty UTF-8 query of at most %d bytes", MaxLongQueryBytes)
	}
	base, err := InterleavePathBaselineFirst(baseline, nil)
	if err != nil {
		return nil, err
	}
	ids := PathHelperIDs()
	results := make([]HelperResult, len(ids))
	for i, id := range ids {
		results[i] = HelperResult{ID: id, Order: slices.Clone(base), Attempted: true}
	}
	normalizedQuery := lexicalhint.NormalizeText(query)
	var hints []int
	if h.normalizedFallback != "" {
		results[0].Fallback = h.normalizedFallback
	} else if len(normalizedQuery) > MaxLongQueryBytes || strings.TrimSpace(normalizedQuery) == "" {
		results[0].Fallback = "normalized_query_error"
	} else {
		results[0].Work.IndexSearchAttempts++
		r, e := h.normalized.RankLongInto(normalizedQuery, Ranking{})
		if e != nil {
			results[0].Fallback = "normalized_search_error"
		} else {
			hints = positiveHelperHints(r)
			mergeHelperResult(&results[0], base, hints)
		}
	}
	joinedQuery := strings.ToLower(query) + " " + normalizedQuery
	if h.joinedFallback != "" {
		results[1].Fallback = h.joinedFallback
	} else if len(joinedQuery) > MaxLongQueryBytes || strings.TrimSpace(joinedQuery) == "" {
		results[1].Fallback = "joined_query_error"
	} else {
		results[1].Work.IndexSearchAttempts++
		full, e := h.full.RankLongInto(joinedQuery, Ranking{})
		if e != nil {
			results[1].Fallback = "joined_full_search_error"
		} else {
			results[1].Work.IndexSearchAttempts++
			names, e := h.names.RankLongInto(joinedQuery, Ranking{})
			if e != nil {
				results[1].Fallback = "joined_basename_search_error"
			} else {
				mergeHelperResult(&results[1], base, reciprocalHelperHints(full, names))
			}
		}
	}
	hints, comparisons, fallback := explicitHelperHints(query, h.paths)
	results[2].Work.AnchorComparisons = comparisons
	results[2].Fallback = fallback
	if fallback == "" {
		mergeHelperResult(&results[2], base, hints)
	}
	return results, nil
}

func positiveHelperHints(r Ranking) []int {
	hints := make([]int, 0, min(PathHelperHintLimit, len(r.Order)))
	for _, d := range r.Order {
		if r.Scores[d] <= 0 || len(hints) == PathHelperHintLimit {
			break
		}
		hints = append(hints, d)
	}
	return hints
}

func reciprocalHelperHints(full, names Ranking) []int {
	scores := make([]float64, len(full.Order))
	for _, field := range []Ranking{full, names} {
		for i, d := range field.Order {
			if field.Scores[d] <= 0 {
				break
			}
			scores[d] += 1 / (60 + float64(i+1))
		}
	}
	hints := make([]int, 0)
	for d, score := range scores {
		if score > 0 {
			hints = append(hints, d)
		}
	}
	slices.SortFunc(hints, func(a, b int) int {
		if scores[a] == scores[b] {
			return a - b
		}
		if scores[a] > scores[b] {
			return -1
		}
		return 1
	})
	return slices.Clone(hints[:min(PathHelperHintLimit, len(hints))])
}

func explicitHelperHints(query string, paths []string) ([]int, int, string) {
	anchors := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' && r != '-' && r != '/'
	})
	n := 0
	for _, token := range anchors {
		if strings.ContainsAny(token, "/.") {
			anchors[n] = token
			n++
		}
	}
	anchors = anchors[:n]
	slices.Sort(anchors)
	anchors = slices.Compact(anchors)
	if len(anchors) == 0 {
		return nil, 0, ""
	}
	if len(paths) > MaxPathHelperAnchorComparisons/len(anchors) {
		return nil, 0, "explicit_path_comparison_budget"
	}
	type evidence struct{ segments, bytes int }
	best := make([]evidence, len(paths))
	comparisons := 0
	for _, token := range anchors {
		segments := 1 + strings.Count(token, "/")
		for d, p := range paths {
			comparisons++
			if p != token && !(len(p) > len(token) && strings.HasSuffix(p, token) && p[len(p)-len(token)-1] == '/') {
				continue
			}
			if segments > best[d].segments || (segments == best[d].segments && len(token) > best[d].bytes) {
				best[d] = evidence{segments, len(token)}
			}
		}
	}
	hints := make([]int, 0)
	for d, ev := range best {
		if ev.segments > 0 {
			hints = append(hints, d)
		}
	}
	slices.SortFunc(hints, func(a, b int) int {
		if best[a].segments != best[b].segments {
			return best[b].segments - best[a].segments
		}
		if best[a].bytes != best[b].bytes {
			return best[b].bytes - best[a].bytes
		}
		return a - b
	})
	return slices.Clone(hints[:min(PathHelperHintLimit, len(hints))]), comparisons, ""
}

func mergeHelperResult(out *HelperResult, baseline, hints []int) {
	order, err := InterleavePathBaselineFirst(baseline, hints)
	if err != nil {
		out.Fallback = "interleave_error"
		return
	}
	out.Order = order
	out.HintCount = len(hints)
}
