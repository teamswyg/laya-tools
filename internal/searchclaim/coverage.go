package searchclaim

import (
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
)

const CoverageDimension = 4

var CoverageNames = [CoverageDimension]string{"first_query_term_coverage", "top20_mean_query_term_coverage", "top20_stddev_query_term_coverage", "query_vocabulary_absent_fraction"}

// QueryCoverage reads the baseline index and top candidate IDs only. All query
// terms use the same tokenizer as BM25; no helper outcomes or labels are inputs.
func QueryCoverage(query string, b hintsearch.Ranking, idx *hintsearch.Index) ([CoverageDimension]float64, error) {
	var f [CoverageDimension]float64
	if len(b.Order) == 0 || len(b.Order) != len(b.Scores) || len(b.Order) > hintsearch.MaxDocuments {
		return f, fmt.Errorf("invalid coverage ranking shape")
	}
	c, e := idx.QueryCoverage(query, b.Order[:min(hintsearch.MaxCoverageDocuments, len(b.Order))])
	if e != nil {
		return f, e
	}
	if c.QueryTerms == 0 {
		return f, nil
	}
	mean := 0.0
	for _, v := range c.Fractions[:c.Documents] {
		mean += v
	}
	mean /= float64(c.Documents)
	variance := 0.0
	for _, v := range c.Fractions[:c.Documents] {
		d := v - mean
		variance += d * d
	}
	return [CoverageDimension]float64{c.Fractions[0], mean, math.Sqrt(variance / float64(c.Documents)), 1 - float64(c.KnownTerms)/float64(c.QueryTerms)}, nil
}
