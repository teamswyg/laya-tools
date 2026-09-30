// Package searchclaim supplies fallible auxiliary-search hints from runtime
// signals. Features cannot accept gold targets, labels or repository identities.
package searchclaim

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const Dimension = 12
const Schema = "riido-search-claim-v1"

var Names = [Dimension]string{"bias", "query_words", "query_bytes", "uppercase_fraction", "digit_fraction", "underscore_fraction", "camel_boundaries_per_word", "unique_word_fraction", "top_score_squashed", "top_two_gap_squashed", "top_ten_positive_fraction", "positive_catalog_fraction"}

func Features(query string, baseline hintsearch.Ranking) ([Dimension]float64, error) {
	var f [Dimension]float64
	if strings.TrimSpace(query) == "" || len(query) > hintsearch.MaxQueryBytes || len(baseline.Order) == 0 || len(baseline.Order) > hintsearch.MaxDocuments || len(baseline.Order) != len(baseline.Scores) {
		return f, fmt.Errorf("invalid runtime feature input")
	}
	words := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	f[0] = 1
	f[1] = math.Min(float64(len(words))/64, 1)
	f[2] = float64(len(query)) / hintsearch.MaxQueryBytes
	count, upper, digits, under, camel := 0, 0, 0, 0, 0
	var prev rune
	for _, r := range query {
		count++
		if unicode.IsUpper(r) {
			upper++
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				camel++
			}
		}
		if unicode.IsDigit(r) {
			digits++
		}
		if r == '_' {
			under++
		}
		prev = r
	}
	if count > 0 {
		f[3] = float64(upper) / float64(count)
		f[4] = float64(digits) / float64(count)
		f[5] = float64(under) / float64(count)
	}
	if len(words) > 0 {
		f[6] = math.Min(float64(camel)/float64(len(words)), 1)
		slices.Sort(words)
		f[7] = float64(len(slices.Compact(words))) / float64(len(words))
	}
	positive := 0
	for _, s := range baseline.Scores {
		if math.IsNaN(s) || math.IsInf(s, 0) || s < 0 {
			return f, fmt.Errorf("invalid score")
		}
		if s > 0 {
			positive++
		}
	}
	// Baseline is produced by the immutable index. Check referenced positions,
	// but do not allocate an additional catalog-sized permutation buffer here.
	for _, id := range baseline.Order {
		if id < 0 || id >= len(baseline.Scores) {
			return f, fmt.Errorf("invalid rank position")
		}
	}
	top := baseline.Scores[baseline.Order[0]]
	second := 0.
	if len(baseline.Order) > 1 {
		second = baseline.Scores[baseline.Order[1]]
	}
	if top < second {
		return f, fmt.Errorf("unsorted ranking")
	}
	f[8] = top / (1 + top)
	f[9] = (top - second) / (1 + top)
	n := min(10, len(baseline.Order))
	for _, id := range baseline.Order[:n] {
		if baseline.Scores[id] > 0 {
			f[10]++
		}
	}
	f[10] /= float64(n)
	f[11] = float64(positive) / float64(len(baseline.Order))
	return f, nil
}

func Score(weights []float64, f [Dimension]float64) float64 {
	var v float64
	for i := range f {
		v += weights[i] * f[i]
	}
	return v
}
