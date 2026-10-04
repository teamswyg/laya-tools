// SPDX-License-Identifier: Apache-2.0
// Package hintsymbol is an experimental symbol-retaining hash representation.
// It has no model loader, Fit, router, truth or approval API. Existing RIIDOH01
// coefficients have different feature meanings and must not be reused here.
package hintsymbol

import (
	"errors"
	"math"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	Schema    = "riido-symbol-hash8192-v1"
	Dimension = 8192
	MaxBytes  = 512
	MaxWords  = 32
	MaxTokens = 64
)

var ErrInput = errors.New("hintsymbol_input_bounds_or_utf8")

// Feature is intentionally a distinct type from legacy hintlearn.Feature.
// Returned storage belongs to the caller; this package keeps no shared scratch.
type Feature struct {
	Index int
	Value float64
}

func tokenize(raw string) ([]string, error) {
	if len(raw) > MaxBytes || !utf8.ValidString(raw) {
		return nil, ErrInput
	}
	s := strings.ToLower(raw)
	var tokens [MaxTokens]string
	n, words, start := 0, 0, -1
	add := func(v string) bool {
		if n == len(tokens) {
			return false
		}
		tokens[n] = v
		n++
		return true
	}
	flush := func(end int) bool {
		if start < 0 {
			return true
		}
		words++
		if words > MaxWords || !add(s[start:end]) {
			return false
		}
		start = -1
		return true
	}
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if start < 0 {
				start = i
			}
			i += size
			continue
		}
		if !flush(i) {
			return nil, ErrInput
		}
		// Longest comparison first. Control-prefixed constant tags cannot be
		// confused with an ordinary word such as "gte" from the raw text.
		var symbol string
		switch s[i] {
		case '<':
			symbol = "\x01<"
			if i+1 < len(s) && s[i+1] == '=' {
				symbol = "\x01<="
				size = 2
			}
		case '>':
			symbol = "\x01>"
			if i+1 < len(s) && s[i+1] == '=' {
				symbol = "\x01>="
				size = 2
			}
		case '[':
			symbol = "\x01["
		case ']':
			symbol = "\x01]"
		case '(':
			symbol = "\x01("
		case ')':
			symbol = "\x01)"
		}
		if symbol != "" && !add(symbol) {
			return nil, ErrInput
		}
		i += size
	}
	if !flush(len(s)) {
		return nil, ErrInput
	}
	return append([]string(nil), tokens[:n]...), nil
}

func terms(words []string) []string {
	out := append([]string(nil), words...)
	for i := 1; i < len(words); i++ {
		out = append(out, words[i-1]+"_"+words[i])
	}
	return out
}
func hashFrom(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// Extract keeps the legacy cross-hash, sign, scale, ordered bigrams, overlap and
// compaction arithmetic. Only comparison/interval token retention changes.
// Without those symbols, valid input matches legacy Features exactly. Bounds
// limit temporary cross storage to 127*127+1 Features, not peak RSS.
func Extract(query, document string) ([]Feature, error) {
	qw, err := tokenize(query)
	if err != nil {
		return nil, err
	}
	dw, err := tokenize(document)
	if err != nil {
		return nil, err
	}
	qt, dt := terms(qw), terms(dw)
	if len(qt) == 0 || len(dt) == 0 {
		return nil, nil
	}
	out := make([]Feature, 0, len(qt)*len(dt)+1)
	scale := 1 / math.Sqrt(float64(len(qt)*len(dt)))
	for _, q := range qt {
		prefix := hashFrom(hashFrom(14695981039346656037, q), "\x00")
		for _, d := range dt {
			h := hashFrom(prefix, d)
			v := scale
			if h>>63 != 0 {
				v = -v
			}
			out = append(out, Feature{1 + int(h%(Dimension-1)), v})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	n := 0
	for _, f := range out {
		if n > 0 && out[n-1].Index == f.Index {
			out[n-1].Value += f.Value
		} else {
			out[n] = f
			n++
		}
	}
	out = out[:n]
	sort.Strings(qw)
	qw = slices.Compact(qw)
	sort.Strings(dw)
	dw = slices.Compact(dw)
	hits := 0
	for _, q := range qw {
		if slices.Contains(dw, q) {
			hits++
		}
	}
	return append(out, Feature{0, float64(hits) / float64(max(1, len(qw)))}), nil
}
