package shortclaim

import (
	"math"
	"strings"
)

const (
	FixedOrderKind     = "fixed_order"
	BM25Kind           = "bm25"
	LexicalOrderedKind = "lexical_ordered"
	NarrowRuleKind     = "narrow_rule"

	// NarrowRulePrefix declares a small ASCII grammar, not ordinary prose or
	// code. It must occur at byte zero of both request and every candidate.
	// After it, one to four ordered clauses have the form:
	//
	//   if [not] NAME [OP NAME] then [not] VERB [OBJECT] else [not] VERB [OBJECT]
	//
	// Clauses are separated by semicolons. NAME/OBJECT are lowercase ASCII
	// identifiers of at most 32 bytes. OP is <, <=, >, >=, ==, or !=; verbs are
	// keep, remove, accept, reject, preserve, and reverse. Operators require
	// surrounding whitespace. Unary condition negation swaps branches, while
	// action negation remains an exact flag. No antonym, operand-reordering,
	// implication, or natural-language equivalence is inferred. Clause order
	// matters. This syntax alone does not establish behavior of actual code.
	NarrowRulePrefix = "behavior-v1:"

	ErrBaselineKind  Error = "shortclaim_unknown_baseline"
	ErrBaselineScore Error = "shortclaim_nonfinite_score"
)

// Ranking is an unverified suggestion for independent verification order.
// Order[:Count] contains every original candidate index exactly once. Scores
// are indexed by original candidate index, are not calibrated probabilities,
// and authorize no action. Unused array slots are zero.
type Ranking struct {
	Kind           string                 `json:"kind"`
	Order          [MaxCandidates]int     `json:"order"`
	Scores         [MaxCandidates]float64 `json:"scores"`
	Count          int                    `json:"count"`
	FallbackReason string                 `json:"fallback_reason,omitempty"`
}

// Baselines computes four nonlearned controls. It reads only request/candidate
// text for features, never IDs, provenance, labels, or dataset roles. Validation
// rechecks exported Prepared values so callers cannot supply forged normal forms.
// All valid inputs retain all candidates, including unsupported rule syntax.
func Baselines(p Prepared) ([4]Ranking, error) {
	var out [4]Ranking
	if err := ValidatePrepared(p); err != nil {
		return out, err
	}
	query, documents, err := baselineTokenize(p)
	if err != nil {
		return out, err
	}
	bm25 := baselineBM25(query, documents, p.Count)
	lexical := baselineLexical(query, documents, p.Count)
	out[0] = baselineRanking(FixedOrderKind, [MaxCandidates]float64{}, p.Count)
	out[1] = baselineRanking(BM25Kind, bm25, p.Count)
	out[2] = baselineRanking(LexicalOrderedKind, lexical, p.Count)
	out[3] = baselineRules(p, bm25)
	for _, ranking := range out {
		if !baselineFinite(ranking) {
			return [4]Ranking{}, ErrBaselineScore
		}
	}
	return out, nil
}

// Rank computes only the selected control, retaining the same validation and
// permutation guarantees as Baselines. Narrow rules still compute BM25 for
// their disclosed bounded tie term and whole-request fallback. Fixed order
// does not tokenize; BM25 and lexical controls do not parse rules or compute
// each other's scores. Unknown kinds return a fixed error and zero output.
func Rank(p Prepared, kind string) (Ranking, error) {
	switch kind {
	case FixedOrderKind, BM25Kind, LexicalOrderedKind, NarrowRuleKind:
	default:
		return Ranking{}, ErrBaselineKind
	}
	if err := ValidatePrepared(p); err != nil {
		return Ranking{}, err
	}
	if kind == FixedOrderKind {
		return baselineRanking(kind, [MaxCandidates]float64{}, p.Count), nil
	}
	query, documents, err := baselineTokenize(p)
	if err != nil {
		return Ranking{}, err
	}
	var out Ranking
	switch kind {
	case BM25Kind:
		out = baselineRanking(kind, baselineBM25(query, documents, p.Count), p.Count)
	case LexicalOrderedKind:
		out = baselineRanking(kind, baselineLexical(query, documents, p.Count), p.Count)
	case NarrowRuleKind:
		out = baselineRules(p, baselineBM25(query, documents, p.Count))
	}
	if !baselineFinite(out) {
		return Ranking{}, ErrBaselineScore
	}
	return out, nil
}

func baselineFinite(ranking Ranking) bool {
	for i := 0; i < ranking.Count; i++ {
		if math.IsNaN(ranking.Scores[i]) || math.IsInf(ranking.Scores[i], 0) {
			return false
		}
	}
	return true
}

type baselineTokens struct {
	Words [MaxNormalizedWords]string
	Count int
}

func baselineTokenize(p Prepared) (baselineTokens, [MaxCandidates]baselineTokens, error) {
	var documents [MaxCandidates]baselineTokens
	query, ok := baselineWords(p.NormalizedRequest)
	if !ok || query.Count == 0 {
		return baselineTokens{}, documents, ErrPrepared
	}
	for i := 0; i < p.Count; i++ {
		documents[i], ok = baselineWords(p.Candidates[i].NormalizedText)
		if !ok || documents[i].Count == 0 {
			return baselineTokens{}, [MaxCandidates]baselineTokens{}, ErrPrepared
		}
	}
	return query, documents, nil
}

// Normalization produces single ASCII-space separators. This scanner retains
// immutable string views and order/repetition without allocating token slices.
func baselineWords(s string) (baselineTokens, bool) {
	var out baselineTokens
	for s != "" {
		word, rest, found := strings.Cut(s, " ")
		if word == "" || out.Count == len(out.Words) {
			return baselineTokens{}, false
		}
		out.Words[out.Count] = word
		out.Count++
		if !found {
			break
		}
		s = rest
		if s == "" {
			return baselineTokens{}, false
		}
	}
	return out, true
}

func baselineContains(words baselineTokens, word string) bool {
	for i := 0; i < words.Count; i++ {
		if words.Words[i] == word {
			return true
		}
	}
	return false
}

func baselineUnique(words baselineTokens) baselineTokens {
	var out baselineTokens
	for i := 0; i < words.Count; i++ {
		if !baselineContains(out, words.Words[i]) {
			out.Words[out.Count] = words.Words[i]
			out.Count++
		}
	}
	return out
}

// BM25 uses candidate-local document frequencies, k1=1.2 and b=0.75.
// Repeated query tokens contribute once; document term frequency is retained.
// The vocabulary is bounded by the 32 query tokens, independent of any dataset.
func baselineBM25(query baselineTokens, docs [MaxCandidates]baselineTokens, count int) [MaxCandidates]float64 {
	query = baselineUnique(query)
	var frequency [MaxCandidates][MaxNormalizedWords]int
	var documentFrequency [MaxNormalizedWords]int
	totalLength := 0
	for i := 0; i < count; i++ {
		totalLength += docs[i].Count
		for q := 0; q < query.Count; q++ {
			for j := 0; j < docs[i].Count; j++ {
				if docs[i].Words[j] == query.Words[q] {
					frequency[i][q]++
				}
			}
			if frequency[i][q] != 0 {
				documentFrequency[q]++
			}
		}
	}
	averageLength := float64(totalLength) / float64(count)
	var scores [MaxCandidates]float64
	for q := 0; q < query.Count; q++ {
		df := float64(documentFrequency[q])
		idf := math.Log1p((float64(count) - df + 0.5) / (df + 0.5))
		for i := 0; i < count; i++ {
			tf := float64(frequency[i][q])
			if tf != 0 {
				norm := 1.2 * (1 - 0.75 + 0.75*float64(docs[i].Count)/averageLength)
				scores[i] += idf * tf * 2.2 / (tf + norm)
			}
		}
	}
	return scores
}

type baselinePair struct{ Left, Right string }

func baselinePairs(words baselineTokens) ([MaxNormalizedWords - 1]baselinePair, int) {
	var out [MaxNormalizedWords - 1]baselinePair
	count := 0
	for i := 1; i < words.Count; i++ {
		pair := baselinePair{words.Words[i-1], words.Words[i]}
		seen := false
		for j := 0; j < count; j++ {
			if out[j] == pair {
				seen = true
				break
			}
		}
		if !seen {
			out[count] = pair
			count++
		}
	}
	return out, count
}

// Lexical control equally weights query-word recall and ordered query-bigram
// recall. A one-word query uses word recall alone. This is not a semantic model.
func baselineLexical(query baselineTokens, docs [MaxCandidates]baselineTokens, count int) [MaxCandidates]float64 {
	pairs, pairCount := baselinePairs(query)
	unique := baselineUnique(query)
	var scores [MaxCandidates]float64
	for i := 0; i < count; i++ {
		hits := 0
		for q := 0; q < unique.Count; q++ {
			if baselineContains(docs[i], unique.Words[q]) {
				hits++
			}
		}
		scores[i] = float64(hits) / float64(unique.Count)
		if pairCount == 0 {
			continue
		}
		orderedHits := 0
		for q := 0; q < pairCount; q++ {
			for j := 1; j < docs[i].Count; j++ {
				if pairs[q].Left == docs[i].Words[j-1] && pairs[q].Right == docs[i].Words[j] {
					orderedHits++
					break
				}
			}
		}
		scores[i] = 0.5*scores[i] + 0.5*float64(orderedHits)/float64(pairCount)
	}
	return scores
}

func baselineRanking(kind string, scores [MaxCandidates]float64, count int) Ranking {
	out := Ranking{Kind: kind, Scores: scores, Count: count}
	for i := 0; i < count; i++ {
		j := i
		for j > 0 && scores[out.Order[j-1]] < scores[i] {
			out.Order[j] = out.Order[j-1]
			j--
		}
		out.Order[j] = i
	}
	return out
}

type baselineRuleAction struct {
	Verb, Object string
	Negated      bool
}

type baselineRuleClause struct {
	Left, Operator, Right string
	Then, Else            baselineRuleAction
}

type baselineRule struct {
	Clauses [4]baselineRuleClause
	Count   int
}

func baselineRuleName(s string) bool {
	if len(s) == 0 || len(s) > 32 || s[0] < 'a' || s[0] > 'z' || s == "if" || s == "then" || s == "else" || s == "not" {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func baselineRuleVerb(s string) bool {
	switch s {
	case "keep", "remove", "accept", "reject", "preserve", "reverse":
		return true
	}
	return false
}

func baselineRuleOperator(s string) bool {
	switch s {
	case "<", "<=", ">", ">=", "==", "!=":
		return true
	}
	return false
}

func baselineRuleActionParse(words []string) (baselineRuleAction, bool) {
	var out baselineRuleAction
	if len(words) != 0 && words[0] == "not" {
		out.Negated = true
		words = words[1:]
	}
	if len(words) < 1 || len(words) > 2 || !baselineRuleVerb(words[0]) {
		return baselineRuleAction{}, false
	}
	out.Verb = words[0]
	if len(words) == 2 {
		if !baselineRuleName(words[1]) {
			return baselineRuleAction{}, false
		}
		out.Object = words[1]
	}
	return out, true
}

func baselineRuleClauseParse(text string) (baselineRuleClause, bool) {
	var words [16]string
	n := 0
	for text != "" {
		text = strings.TrimLeft(text, " \t\r\n")
		if text == "" {
			break
		}
		end := strings.IndexAny(text, " \t\r\n")
		if end < 0 {
			end = len(text)
		}
		if n == len(words) {
			return baselineRuleClause{}, false
		}
		words[n] = text[:end]
		n++
		text = text[end:]
	}
	if n < 6 || words[0] != "if" {
		return baselineRuleClause{}, false
	}
	then, otherwise := -1, -1
	for i := 1; i < n; i++ {
		if words[i] == "then" {
			if then != -1 {
				return baselineRuleClause{}, false
			}
			then = i
		}
		if words[i] == "else" {
			if otherwise != -1 {
				return baselineRuleClause{}, false
			}
			otherwise = i
		}
	}
	if then <= 1 || otherwise <= then+1 || otherwise >= n-1 {
		return baselineRuleClause{}, false
	}
	predicate := words[1:then]
	negated := predicate[0] == "not"
	if negated {
		predicate = predicate[1:]
	}
	var out baselineRuleClause
	if (len(predicate) != 1 && len(predicate) != 3) || !baselineRuleName(predicate[0]) {
		return out, false
	}
	out.Left = predicate[0]
	if len(predicate) == 3 {
		if !baselineRuleOperator(predicate[1]) || !baselineRuleName(predicate[2]) {
			return baselineRuleClause{}, false
		}
		out.Operator, out.Right = predicate[1], predicate[2]
	}
	var ok bool
	out.Then, ok = baselineRuleActionParse(words[then+1 : otherwise])
	if !ok {
		return baselineRuleClause{}, false
	}
	out.Else, ok = baselineRuleActionParse(words[otherwise+1 : n])
	if !ok {
		return baselineRuleClause{}, false
	}
	if negated {
		out.Then, out.Else = out.Else, out.Then
	}
	return out, true
}

func baselineRuleParse(text string) (baselineRule, bool) {
	var out baselineRule
	if !strings.HasPrefix(text, NarrowRulePrefix) {
		return out, false
	}
	text = text[len(NarrowRulePrefix):]
	for text != "" {
		clause, rest, found := strings.Cut(text, ";")
		if out.Count == len(out.Clauses) {
			return baselineRule{}, false
		}
		parsed, ok := baselineRuleClauseParse(clause)
		if !ok {
			return baselineRule{}, false
		}
		out.Clauses[out.Count] = parsed
		out.Count++
		if !found {
			return out, true
		}
		text = rest
	}
	return baselineRule{}, false
}

func baselineRules(p Prepared, bm25 [MaxCandidates]float64) Ranking {
	fallback := baselineRanking(NarrowRuleKind, bm25, p.Count)
	request, ok := baselineRuleParse(p.Request)
	if !ok {
		fallback.FallbackReason = "unsupported_rule_request"
		return fallback
	}
	var candidates [MaxCandidates]baselineRule
	for i := 0; i < p.Count; i++ {
		candidates[i], ok = baselineRuleParse(p.Candidates[i].Text)
		if !ok {
			fallback.FallbackReason = "unsupported_rule_candidate"
			return fallback
		}
	}
	var scores [MaxCandidates]float64
	for i := 0; i < p.Count; i++ {
		matches := 0
		for j := 0; j < min(request.Count, candidates[i].Count); j++ {
			if request.Clauses[j] == candidates[i].Clauses[j] {
				matches++
			}
		}
		// Ordered canonical-clause agreement dominates a bounded BM25 term.
		// This remains an unverified ordering score, never a behavior verdict.
		scores[i] = float64(matches)/float64(max(request.Count, candidates[i].Count)) + 0.001*bm25[i]/(1+bm25[i])
	}
	return baselineRanking(NarrowRuleKind, scores, p.Count)
}
