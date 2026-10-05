// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Prospective helper for the frozen JWT development experiment; not executed.
// Normalization and BM25 are derived from this repository's lexicalhint and
// shortclaim sources. Scores suggest verification order and certify nothing.
package main

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

const retrievalCandidateLimit = 8
const retrievalTextLimit = 512
const retrievalWordLimit = 32

type retrievalKind uint8

const (
	retrievalBM25 retrievalKind = iota
	retrievalJaccard
)

type retrievalCode string

const (
	retrievalOK                 retrievalCode = ""
	retrievalUnknownKind        retrievalCode = "unknown_retrieval_kind"
	retrievalUnknownText        retrievalCode = "unknown_retrieval_text"
	retrievalUnknownUnicode     retrievalCode = "unknown_retrieval_unicode"
	retrievalUnknownNormalized  retrievalCode = "unknown_retrieval_normalized"
	retrievalUnknownScore       retrievalCode = "unknown_retrieval_score"
	retrievalUnknownPermutation retrievalCode = "unknown_retrieval_permutation"
	retrievalUnknownPanic       retrievalCode = "unknown_retrieval_panic"
)

type retrievalRanking struct {
	Order      [retrievalCandidateLimit]int
	Scores     [retrievalCandidateLimit]float64 // Original declaration positions.
	WordCounts [retrievalCandidateLimit + 1]int // Query, then eight captions.
}
type retrievalWords struct {
	Values [retrievalWordLimit]string
	Count  int
}

func retrievalDeclaration() retrievalRanking {
	var out retrievalRanking
	for i := range out.Order {
		out.Order[i] = i
	}
	return out
}

// The complete call belongs inside the operation's rank and parent timers.
// It accepts only raw query/caption text. There is no ID, Want, fixture, option,
// result, source-line, model, global scratch, map, lock or prepared-score input.
// A nonempty code is unknown; caller must not treat its declaration order as a
// successful retrieval fallback. Every returned value is independently owned.
func rankRetrievalText(query string, captions [retrievalCandidateLimit]string, kind retrievalKind) (out retrievalRanking, code retrievalCode) {
	out = retrievalDeclaration()
	defer func() {
		if recover() != nil {
			out, code = retrievalDeclaration(), retrievalUnknownPanic
		}
	}()
	if kind != retrievalBM25 && kind != retrievalJaccard {
		return out, retrievalUnknownKind
	}
	q, code := retrievalPrepare(query)
	if code != retrievalOK {
		return out, code
	}
	out.WordCounts[0] = q.Count
	var docs [retrievalCandidateLimit]retrievalWords
	for i, caption := range captions {
		docs[i], code = retrievalPrepare(caption)
		if code != retrievalOK {
			return out, code
		}
		out.WordCounts[i+1] = docs[i].Count
	}
	var scores [retrievalCandidateLimit]float64
	if kind == retrievalBM25 {
		scores = retrievalBM25Scores(q, docs)
	} else {
		scores = retrievalJaccardScores(q, docs)
	}
	for _, score := range scores {
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 {
			return out, retrievalUnknownScore
		}
	}
	out.Scores = scores
	// Insertion sort uses strict greater-than. Equal scores retain declaration
	// order, independent of IDs, labels and prior observed candidate success.
	for i := range out.Order {
		j := i
		for j > 0 && scores[out.Order[j-1]] < scores[i] {
			out.Order[j] = out.Order[j-1]
			j--
		}
		out.Order[j] = i
	}
	var seen [retrievalCandidateLimit]bool
	for _, index := range out.Order {
		if index < 0 || index >= len(seen) || seen[index] {
			return retrievalDeclaration(), retrievalUnknownPermutation
		}
		seen[index] = true
	}
	return out, retrievalOK
}

func retrievalPrepare(raw string) (retrievalWords, retrievalCode) {
	if len(raw) == 0 || len(raw) > retrievalTextLimit {
		return retrievalWords{}, retrievalUnknownText
	}
	if !utf8.ValidString(raw) {
		return retrievalWords{}, retrievalUnknownUnicode
	}
	normalized := retrievalNormalize(raw)
	if normalized == "" || len(normalized) > retrievalTextLimit {
		return retrievalWords{}, retrievalUnknownNormalized
	}
	var words retrievalWords
	for normalized != "" {
		word, rest, found := strings.Cut(normalized, " ")
		if word == "" || words.Count == len(words.Values) {
			return retrievalWords{}, retrievalUnknownNormalized
		}
		words.Values[words.Count] = word
		words.Count++
		if !found {
			break
		}
		normalized = rest
		if normalized == "" {
			return retrievalWords{}, retrievalUnknownNormalized
		}
	}
	return words, retrievalOK
}

// Same word/CamelCase splitting, Unicode lowercasing and punctuation removal
// as lexicalhint.NormalizeText. This control does not infer operator direction,
// negation, quantifier scope, policy semantics or validity of actual code.
func retrievalNormalize(raw string) string {
	rs := []rune(raw)
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
	return strings.Join(strings.Fields(b.String()), " ")
}
func retrievalContains(words retrievalWords, value string) bool {
	for i := 0; i < words.Count; i++ {
		if words.Values[i] == value {
			return true
		}
	}
	return false
}
func retrievalUnique(words retrievalWords) retrievalWords {
	var out retrievalWords
	for i := 0; i < words.Count; i++ {
		if !retrievalContains(out, words.Values[i]) {
			out.Values[out.Count] = words.Values[i]
			out.Count++
		}
	}
	return out
}

// Candidate-local BM25, k1=1.2/b=0.75. The math operation order matches the
// existing shortclaim baseline; query repetitions contribute once, while
// document token repetitions and length remain relevant.
func retrievalBM25Scores(query retrievalWords, docs [retrievalCandidateLimit]retrievalWords) [retrievalCandidateLimit]float64 {
	query = retrievalUnique(query)
	var frequency [retrievalCandidateLimit][retrievalWordLimit]int
	var df [retrievalWordLimit]int
	totalLength := 0
	for i := range docs {
		totalLength += docs[i].Count
		for q := 0; q < query.Count; q++ {
			for j := 0; j < docs[i].Count; j++ {
				if docs[i].Values[j] == query.Values[q] {
					frequency[i][q]++
				}
			}
			if frequency[i][q] != 0 {
				df[q]++
			}
		}
	}
	averageLength := float64(totalLength) / float64(retrievalCandidateLimit)
	var scores [retrievalCandidateLimit]float64
	for q := 0; q < query.Count; q++ {
		n := float64(df[q])
		idf := math.Log1p((float64(retrievalCandidateLimit) - n + 0.5) / (n + 0.5))
		for i := range docs {
			tf := float64(frequency[i][q])
			if tf != 0 {
				norm := 1.2 * (1 - 0.75 + 0.75*float64(docs[i].Count)/averageLength)
				scores[i] += idf * tf * 2.2 / (tf + norm)
			}
		}
	}
	return scores
}

// Exact unique-token intersection/union, not recall or ordered-bigram scoring.
func retrievalJaccardScores(query retrievalWords, docs [retrievalCandidateLimit]retrievalWords) [retrievalCandidateLimit]float64 {
	query = retrievalUnique(query)
	var scores [retrievalCandidateLimit]float64
	for i, document := range docs {
		document = retrievalUnique(document)
		hits := 0
		for q := 0; q < query.Count; q++ {
			if retrievalContains(document, query.Values[q]) {
				hits++
			}
		}
		union := query.Count + document.Count - hits
		if union > 0 {
			scores[i] = float64(hits) / float64(union)
		}
	}
	return scores
}
