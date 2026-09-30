package hintsearch

import (
	"fmt"
	"slices"
	"strings"
)

const MaxCoverageDocuments = 20

// Coverage summarizes direct unique-query-term presence in caller-selected
// candidates. Fractions are in the same order as the supplied document IDs.
// It is lexical evidence, not relevance confidence.
type Coverage struct {
	QueryTerms, KnownTerms, Documents int
	Fractions                         [MaxCoverageDocuments]float64
}

// QueryCoverage reuses immutable baseline posting arrays; it neither ranks an
// auxiliary index nor scans original code. Query tokenization allocates; the
// bounded candidate counters and result use fixed arrays. Concurrent callers
// share no mutable query state. IDs must belong to this permission-filtered index.
func (idx *Index) QueryCoverage(query string, ids []int) (Coverage, error) {
	var out Coverage
	if idx == nil || len(ids) == 0 || len(ids) > MaxCoverageDocuments || strings.TrimSpace(query) == "" || len(query) > MaxQueryBytes {
		return out, fmt.Errorf("invalid coverage request")
	}
	for i, id := range ids {
		if id < 0 || id >= len(idx.norms) {
			return out, fmt.Errorf("invalid coverage document id")
		}
		for _, earlier := range ids[:i] {
			if earlier == id {
				return out, fmt.Errorf("duplicate coverage document id")
			}
		}
	}
	out.Documents = len(ids)
	terms := words(query)
	slices.Sort(terms)
	terms = slices.Compact(terms)
	out.QueryTerms = len(terms)
	if len(terms) == 0 {
		return out, nil
	}
	var hits [MaxCoverageDocuments]int
	for _, term := range terms {
		t, ok := slices.BinarySearch(idx.terms, term)
		if !ok {
			continue
		}
		out.KnownTerms++
		docs := idx.docs[idx.starts[t]:idx.starts[t+1]]
		for j, id := range ids {
			if _, found := slices.BinarySearch(docs, id); found {
				hits[j]++
			}
		}
	}
	for i := range ids {
		out.Fractions[i] = float64(hits[i]) / float64(len(terms))
	}
	return out, nil
}
