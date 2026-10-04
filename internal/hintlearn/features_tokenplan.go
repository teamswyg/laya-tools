// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 teamswyg contributors.
package hintlearn

import (
	"math"
	"reflect"
	"slices"
	"sort"
)

// FeatureText owns immutable raw terms and overlap words; copies share slices.
// Prepared empty text is valid; zero is invalid. Never use normalized input.
type FeatureText struct {
	ordered     []string
	uniqueWords []string
	ready       bool
}

// FeatureQuery caches ordered FNV(term+NUL) prefixes, including duplicates.
// Copies are immutable; zero is invalid. Do not retain plans in public owners.
type FeatureQuery struct {
	text     FeatureText
	prefixes []uint64
	ready    bool
}

func tokenPlanTermSlots(words int) (int, error) {
	maxInt := int(^uint(0) >> 1)
	if words < 0 || words > maxInt/2+1 {
		return 0, ErrFeatureScratchSize
	}
	if words == 0 {
		return 0, nil
	}
	n := words + (words - 1)
	if n > maxInt/int(reflect.TypeOf("").Size()) {
		return 0, ErrFeatureScratchSize
	}
	return n, nil
}
func tokenPlanPrefixSlots(terms int) error {
	maxInt := int(^uint(0) >> 1)
	if terms < 0 || terms > maxInt/int(reflect.TypeOf(uint64(0)).Size()) {
		return ErrFeatureScratchSize
	}
	return nil
}
func tokenPlanCrossSlots(query, document int) (int, error) {
	if query < 0 || document < 0 {
		return 0, ErrFeatureScratchSize
	}
	if query == 0 || document == 0 {
		return 0, nil
	}
	maxInt := int(^uint(0) >> 1)
	if query > (maxInt-1)/document {
		return 0, ErrFeatureScratchSize
	}
	n := query*document + 1
	if !validScratchSize(n) {
		return 0, ErrFeatureScratchSize
	}
	return n, nil
}

// PrepareFeatureText tokenizes once; strings may share immutable raw bytes.
// Callers enforce text bounds. Size checks do not promise an OOM limit.
func PrepareFeatureText(raw string) (FeatureText, error) {
	ws := words(raw)
	n, err := tokenPlanTermSlots(len(ws))
	if err != nil {
		return FeatureText{}, err
	}
	if n == 0 {
		return FeatureText{ready: true}, nil
	}
	ordered := make([]string, n)
	copy(ordered, ws)
	for i := 1; i < len(ws); i++ {
		ordered[len(ws)+i-1] = ws[i-1] + "_" + ws[i]
	}
	// Separate backing preserves term order while overlap words are compacted.
	sort.Strings(ws)
	ws = slices.Compact(ws)
	return FeatureText{ordered: ordered, uniqueWords: ws, ready: true}, nil
}

// PrepareFeatureQuery consumes prepared text without tokenizing again.
func PrepareFeatureQuery(text FeatureText) (FeatureQuery, error) {
	if !text.ready {
		return FeatureQuery{}, ErrFeatureScratchSize
	}
	if err := tokenPlanPrefixSlots(len(text.ordered)); err != nil {
		return FeatureQuery{}, err
	}
	var prefixes []uint64
	if len(text.ordered) != 0 {
		prefixes = make([]uint64, len(text.ordered))
		for i, term := range text.ordered {
			prefixes[i] = hashFrom(hash(term), "\x00")
		}
	}
	return FeatureQuery{text: text, prefixes: prefixes, ready: true}, nil
}
func validTokenPlans(query FeatureQuery, document FeatureText) bool {
	return query.ready && query.text.ready && document.ready &&
		len(query.prefixes) == len(query.text.ordered)
}

// FeatureCardinalityPrepared counts all indices, including cancelled zeros.
// It adds overlap; both passes still pay cross hashing.
func FeatureCardinalityPrepared(query FeatureQuery, document FeatureText) (count, crossSlots int, err error) {
	if !validTokenPlans(query, document) {
		return 0, 0, ErrFeatureScratchSize
	}
	crossSlots, err = tokenPlanCrossSlots(len(query.text.ordered), len(document.ordered))
	if err != nil || crossSlots == 0 {
		return 0, crossSlots, err
	}
	var seen [Dimension / 64]uint64
	for _, prefix := range query.prefixes {
		for _, term := range document.ordered {
			index := 1 + int(hashFrom(prefix, term)%(Dimension-1))
			word, mask := index/64, uint64(1)<<uint(index%64)
			if seen[word]&mask == 0 {
				seen[word] |= mask
				count++
			}
		}
	}
	return count + 1, crossSlots, nil
}

// FeaturesPrepared borrows the row; copy it before scratch reuse. Plans stay
// immutable, errors precede writes, and original sort/FP order is preserved.
func (s *FeatureScratch) FeaturesPrepared(query FeatureQuery, document FeatureText) ([]Feature, error) {
	if s == nil {
		return nil, ErrFeatureScratchSize
	}
	if !validTokenPlans(query, document) {
		return nil, ErrFeatureScratchSize
	}
	slots, err := tokenPlanCrossSlots(len(query.text.ordered), len(document.ordered))
	if err != nil {
		return nil, err
	}
	if slots == 0 {
		return nil, nil
	}
	if slots > cap(s.cross) {
		return nil, ErrFeatureScratchCapacity
	}
	fs := s.cross[:0]
	scale := 1 / math.Sqrt(float64(len(query.text.ordered)*len(document.ordered)))
	for _, prefix := range query.prefixes {
		for _, term := range document.ordered {
			h := hashFrom(prefix, term)
			v := scale
			if h>>63 != 0 {
				v = -v
			}
			fs = append(fs, Feature{1 + int(h%(Dimension-1)), v})
		}
	}
	sort.Slice(fs, func(i, j int) bool { return fs[i].Index < fs[j].Index })
	n := 0
	for _, f := range fs {
		if n > 0 && fs[n-1].Index == f.Index {
			fs[n-1].Value += f.Value
		} else {
			fs[n] = f
			n++
		}
	}
	fs = fs[:n]
	hits := 0
	for _, word := range query.text.uniqueWords {
		if slices.Contains(document.uniqueWords, word) {
			hits++
		}
	}
	fs = append(fs, Feature{0, float64(hits) / float64(max(1, len(query.text.uniqueWords)))})
	return fs, nil
}
