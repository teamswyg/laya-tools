// Package hintsearch provides a bounded lexical baseline and a scheduler for
// fallible hints. It does not verify relevance, execute actions or check access.
// Callers must supply only documents the current principal may read.
package hintsearch

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"unicode"
)

const MaxDocuments = 4096
const MaxQueryBytes = 8192
const MaxCatalogBytes = 2 << 20

type occurrence struct {
	term  string
	doc   int
	count int
}

// Index is an immutable BM25 index (k1=1.2, b=0.75). Posting columns are
// contiguous slices, with no shared mutation or locks during queries.
// Strings, postings and normalization all count toward resident index cost.
type Index struct {
	terms       []string
	starts      []int
	docs        []int
	frequencies []float64
	norms       []float64
}

func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// New snapshots the text into a sorted vocabulary and posting arrays. No model,
// network, tokenizer assets or Python runtime is required.
func New(documents []string) (*Index, error) {
	if len(documents) == 0 || len(documents) > MaxDocuments {
		return nil, fmt.Errorf("require 1..%d documents", MaxDocuments)
	}
	total := 0
	for _, s := range documents {
		if len(s) > MaxCatalogBytes-total {
			return nil, fmt.Errorf("catalog exceeds %d bytes", MaxCatalogBytes)
		}
		total += len(s)
	}
	idx := &Index{norms: make([]float64, len(documents))}
	var all []occurrence
	sum := 0
	for d, s := range documents {
		ts := words(s)
		sort.Strings(ts)
		idx.norms[d] = float64(len(ts))
		sum += len(ts)
		for i := 0; i < len(ts); {
			j := i + 1
			for j < len(ts) && ts[j] == ts[i] {
				j++
			}
			all = append(all, occurrence{ts[i], d, j - i})
			i = j
		}
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].term == all[j].term {
			return all[i].doc < all[j].doc
		}
		return all[i].term < all[j].term
	})
	for i, o := range all {
		if i == 0 || all[i-1].term != o.term {
			idx.terms = append(idx.terms, o.term)
			idx.starts = append(idx.starts, len(idx.docs))
		}
		idx.docs = append(idx.docs, o.doc)
		idx.frequencies = append(idx.frequencies, float64(o.count))
	}
	idx.starts = append(idx.starts, len(idx.docs))
	average := float64(sum) / float64(len(documents))
	if average == 0 {
		average = 1
	}
	for i, l := range idx.norms {
		idx.norms[i] = 1.2 * (.25 + .75*l/average)
	}
	return idx, nil
}

// Ranking retains even zero-score candidates; scores are uncalibrated lexical
// scores indexed by original document position, not probabilities.
type Ranking struct {
	Order  []int
	Scores []float64
}

// Rank returns all document positions, with stable catalog order for ties.
// The caller owns the result and can rank concurrently against the same Index.
func (idx *Index) Rank(query string) (Ranking, error) {
	return idx.RankInto(query, Ranking{})
}

// RankInto reuses caller-owned result arrays when their capacity is sufficient.
// Every score and position is reset on success, including zero-score candidates.
// The returned slices may alias dst and are overwritten by its next reuse.
// Keep separate buffers for concurrent calls; the Index remains immutable.
// Invalid queries return before modifying dst. Rank retains independently owned
// results for callers that need to hold multiple rankings simultaneously.
func (idx *Index) RankInto(query string, dst Ranking) (Ranking, error) {
	if strings.TrimSpace(query) == "" || len(query) > MaxQueryBytes {
		return Ranking{}, fmt.Errorf("require nonempty query of at most %d bytes", MaxQueryBytes)
	}
	n := len(idx.norms)
	if cap(dst.Order) < n {
		dst.Order = make([]int, n)
	} else {
		dst.Order = dst.Order[:n]
	}
	if cap(dst.Scores) < n {
		dst.Scores = make([]float64, n)
	} else {
		dst.Scores = dst.Scores[:n]
		clear(dst.Scores)
	}
	r := dst
	ts := words(query)
	sort.Strings(ts)
	ts = slices.Compact(ts)
	for _, t := range ts {
		v, found := slices.BinarySearch(idx.terms, t)
		if !found {
			continue
		}
		start, end := idx.starts[v], idx.starts[v+1]
		df := float64(end - start)
		idf := math.Log(1 + (float64(len(idx.norms))-df+.5)/(df+.5))
		for p := start; p < end; p++ {
			d := idx.docs[p]
			tf := idx.frequencies[p]
			r.Scores[d] += idf * tf * 2.2 / (tf + idx.norms[d])
		}
	}
	for i := range r.Order {
		r.Order[i] = i
	}
	sort.Slice(r.Order, func(i, j int) bool {
		a, b := r.Order[i], r.Order[j]
		if r.Scores[a] == r.Scores[b] {
			return a < b
		}
		return r.Scores[a] > r.Scores[b]
	})
	return r, nil
}

// Interleave alternates a hint candidate and a baseline candidate, skipping
// previously emitted candidates. Baseline must be a permutation of 0..n-1;
// hints may be an empty or partial permutation of those same positions.
//
// With unit-cost, accurate candidate verification, a target at baseline rank r
// is inspected within min(n, 2*r) checks, regardless of hint quality. This is
// NOT a bound on wall time, token costs, multiple-evidence completion, or hint
// construction. Callers must bound those costs separately. No candidate is lost.
func Interleave(baseline, hints []int) ([]int, error) {
	return interleave(baseline, hints, false)
}

// InterleaveBaselineFirst preserves the baseline's first candidate and then
// alternates baseline and hint candidates. Partial hints are supported. With
// accurate unit-cost verification, baseline rank r is reached within
// min(n, 2*r-1) checks. This does not bound time, tokens or hint construction.
func InterleaveBaselineFirst(baseline, hints []int) ([]int, error) {
	return interleave(baseline, hints, true)
}

func interleave(baseline, hints []int, baselineFirst bool) ([]int, error) {
	n := len(baseline)
	if n == 0 || n > MaxDocuments || len(hints) > n {
		return nil, fmt.Errorf("invalid candidate counts")
	}
	seen := make([]bool, n)
	for _, d := range baseline {
		if d < 0 || d >= n || seen[d] {
			return nil, fmt.Errorf("baseline must be a full permutation")
		}
		seen[d] = true
	}
	clear(seen)
	for _, d := range hints {
		if d < 0 || d >= n || seen[d] {
			return nil, fmt.Errorf("hints must be a partial permutation")
		}
		seen[d] = true
	}
	clear(seen)
	out := make([]int, 0, n)
	h, b := 0, 0
	emitHint := func() {
		for h < len(hints) && seen[hints[h]] {
			h++
		}
		if h < len(hints) {
			d := hints[h]
			h++
			seen[d] = true
			out = append(out, d)
		}
	}
	emitBaseline := func() {
		for b < n && seen[baseline[b]] {
			b++
		}
		if b < n {
			d := baseline[b]
			b++
			seen[d] = true
			out = append(out, d)
		}
	}
	for len(out) < n {
		if baselineFirst {
			emitBaseline()
			emitHint()
		} else {
			emitHint()
			emitBaseline()
		}
	}
	return out, nil
}
