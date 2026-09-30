// Package lexicalhint provides a small, corpus-independent lexical control.
// It does not claim semantic understanding. Feature schema is versioned because
// weights from other feature contracts must never be interpreted as this one.
package lexicalhint

import (
	"math"
	"slices"
	"strings"
	"unicode"
)

const Dimension = 16
const Schema = "riido-lexical-features-v1"

var Names = [Dimension]string{"bias", "query_token_recall", "token_jaccard", "token_cosine", "query_bigram_recall", "bigram_jaccard", "trigram_dice", "name_token_recall", "name_trigram_dice", "exact_recall_squared", "bigram_recall_squared", "shared_name_and_code", "query_length", "code_length", "length_ratio", "query_code_length_product"}

func tokens(s string) []string {
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if unicode.IsUpper(r) && i > 0 && (unicode.IsLower(rs[i-1]) || unicode.IsDigit(rs[i-1]) || (i+1 < len(rs) && unicode.IsLower(rs[i+1]) && unicode.IsUpper(rs[i-1]))) {
			b.WriteByte(' ')
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}
func unique(xs []string) []string { slices.Sort(xs); return slices.Compact(xs) }
func intersection(a, b []string) int {
	i, j, n := 0, 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			n++
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return n
}
func pairs(ts []string) []string {
	out := make([]string, 0, max(0, len(ts)-1))
	for i := 1; i < len(ts); i++ {
		out = append(out, ts[i-1]+"\x00"+ts[i])
	}
	return unique(out)
}
func grams(ts []string) []string {
	out := []string{}
	for _, t := range ts {
		rs := []rune("^" + t + "$")
		for i := 2; i < len(rs); i++ {
			out = append(out, string(rs[i-2:i+1]))
		}
	}
	return unique(out)
}
func recall(a, b []string) float64 { return float64(intersection(a, b)) / float64(max(1, len(a))) }
func dice(a, b []string) float64 {
	return 2 * float64(intersection(a, b)) / float64(max(1, len(a)+len(b)))
}

// Name is a lexical declaration-name heuristic, not a language parser.
// Quoted keywords and receiver declarations can confuse it. Full-code
// features still use the entire input; this heuristic does not truncate it.
func Name(code string) string {
	fields := strings.FieldsFunc(code, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' })
	for i, t := range fields {
		if (t == "def" || t == "func" || t == "function") && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

// Features expects caller-enforced byte/token budgets. Length features are an
// explicit ablation: they can learn source artifacts instead of relevance.
func Features(query, code string, withLength bool) [Dimension]float64 {
	q, d := tokens(query), tokens(code)
	qn, dn := len(q), len(d)
	qp, dp := pairs(q), pairs(d)
	q = unique(q)
	d = unique(d)
	name := unique(tokens(Name(code)))
	qg, dg := grams(q), grams(d)
	hits := float64(intersection(q, d))
	r := hits / float64(max(1, len(q)))
	br := recall(qp, dp)
	nr := recall(q, name)
	x := [Dimension]float64{1, r, hits / float64(max(1, len(q)+len(d)-int(hits))), hits / math.Sqrt(float64(max(1, len(q))*max(1, len(d)))), br, float64(intersection(qp, dp)) / float64(max(1, len(qp)+len(dp)-intersection(qp, dp))), dice(qg, dg), nr, dice(qg, grams(name)), r * r, br * br, r * nr}
	if withLength {
		x[12] = math.Log1p(float64(qn)) / 8
		x[13] = math.Log1p(float64(dn)) / 8
		x[14] = float64(min(qn, dn)) / float64(max(1, max(qn, dn)))
		x[15] = x[12] * x[13]
	}
	return x
}
func Score(w []float64, x [Dimension]float64) float64 {
	s := 0.
	for i, v := range x {
		s += w[i] * v
	}
	return s
}

// NormalizeText exposes the same identifier/word splitting used by Features.
// It removes punctuation, preserves token order/repetition and does not filter
// keywords. Callers must enforce a post-normalization byte budget too.
func NormalizeText(s string) string { return strings.Join(tokens(s), " ") }
