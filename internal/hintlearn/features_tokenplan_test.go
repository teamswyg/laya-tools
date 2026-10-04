// SPDX-License-Identifier: Apache-2.0
package hintlearn

import (
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Synthetic controls; original Features is the independent oracle.
func tokenPlanPair(t *testing.T, q, d string) (FeatureQuery, FeatureText) {
	t.Helper()
	text, err := PrepareFeatureText(q)
	if err != nil {
		t.Fatal(err)
	}
	query, err := PrepareFeatureQuery(text)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := PrepareFeatureText(d)
	if err != nil {
		t.Fatal(err)
	}
	return query, doc
}
func TestFeatureTokenPlanOracleAndImmutability(t *testing.T) {
	for _, pair := range [][2]string{
		{"z a z b", "b z z a"},
		{"RawPath", "rawpath"},
		{"Å β 한글 １２", "å Β 한글 １２"},
		{"left\x00right", "right left"},
		{"", "alpha"}, {"alpha", "!!!🙂"},
		{strings.Repeat("repeat ", 64), "repeat another repeat"},
	} {
		q, d := tokenPlanPair(t, pair[0], pair[1])
		terms := append([]string(nil), q.text.ordered...)
		prefixes := append([]uint64(nil), q.prefixes...)
		unique := append([]string(nil), q.text.uniqueWords...)
		docTerms := append([]string(nil), d.ordered...)
		docUnique := append([]string(nil), d.uniqueWords...)
		for i, term := range terms {
			if prefixes[i] != hashFrom(hash(term), "\x00") {
				t.Fatal("prefix order/separator changed")
			}
		}
		want := Features(pair[0], pair[1])
		count, slots, err := FeatureCardinalityPrepared(q, d)
		if err != nil || count != len(want) {
			t.Fatal("cardinality", count, slots, err)
		}
		expectedSlots := 0
		if len(terms) != 0 && len(docTerms) != 0 {
			expectedSlots = len(terms)*len(docTerms) + 1
		}
		if slots != expectedSlots {
			t.Fatal("raw term capacity changed")
		}
		s, err := NewFeatureScratch(slots)
		if err != nil {
			t.Fatal(err)
		}
		for repeat := 0; repeat < 3; repeat++ {
			count, currentSlots, err := FeatureCardinalityPrepared(q, d)
			if err != nil || count != len(want) || currentSlots != slots {
				t.Fatal("repeat cardinality", err)
			}
			got, err := s.FeaturesPrepared(q, d)
			if err != nil || !scratchBitsEqual(got, want) {
				t.Fatal("index/value bits/order/nilness", err)
			}
			if !reflect.DeepEqual(q.text.ordered, terms) || !reflect.DeepEqual(q.prefixes, prefixes) ||
				!reflect.DeepEqual(q.text.uniqueWords, unique) || !reflect.DeepEqual(d.ordered, docTerms) ||
				!reflect.DeepEqual(d.uniqueWords, docUnique) {
				t.Fatal("consumer mutated immutable plans")
			}
		}
	}
	q, d := tokenPlanPair(t, "z a z b", "a z")
	if !reflect.DeepEqual(q.text.ordered, []string{"z", "a", "z", "b", "z_a", "a_z", "z_b"}) {
		t.Fatal("sorting/compaction changed cross term order")
	}
	s, _ := NewFeatureScratch(22)
	got, err := s.FeaturesPrepared(q, d)
	if err != nil || got[len(got)-1].Value != 2.0/3.0 {
		t.Fatal("overlap must divide by unique query words", err)
	}
	if scratchBitsEqual(Features("RawPath", "rawpath"), Features("raw path", "rawpath")) {
		t.Fatal("raw/normalized distinction control collapsed")
	}
}
func TestFeatureTokenPlanCancelledZero(t *testing.T) {
	doc, index := scratchCancelledPair(t)
	q, d := tokenPlanPair(t, "q", doc)
	count, slots, err := FeatureCardinalityPrepared(q, d)
	if err != nil || count != 3 || slots != 4 {
		t.Fatal("cancelled cardinality", count, slots, err)
	}
	s, _ := NewFeatureScratch(slots)
	got, err := s.FeaturesPrepared(q, d)
	if err != nil || !scratchBitsEqual(got, Features("q", doc)) {
		t.Fatal("cancelled parity", err)
	}
	found := false
	for _, f := range got {
		if f.Index == index {
			found = true
			if math.Float64bits(f.Value) != 0 {
				t.Fatal("cancelled positive zero changed")
			}
		}
	}
	if !found {
		t.Fatal("cancelled zero pruned")
	}
}
func TestFeatureTokenPlanBoundsAndErrorPrecedence(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	if n, err := tokenPlanTermSlots(0); n != 0 || err != nil {
		t.Fatal("empty term size")
	}
	for _, words := range []int{-1, maxInt} {
		if n, err := tokenPlanTermSlots(words); n != 0 || err != ErrFeatureScratchSize {
			t.Fatal("term arithmetic/storage overflow admitted")
		}
	}
	maxWords := (maxInt/int(reflect.TypeOf("").Size()))/2 + 1
	if _, err := tokenPlanTermSlots(maxWords); err != nil {
		t.Fatal("valid term size boundary")
	}
	if _, err := tokenPlanTermSlots(maxWords + 1); err != ErrFeatureScratchSize {
		t.Fatal("term storage overflow")
	}
	maxPrefixes := maxInt / int(reflect.TypeOf(uint64(0)).Size())
	if tokenPlanPrefixSlots(maxPrefixes) != nil || tokenPlanPrefixSlots(maxPrefixes+1) != ErrFeatureScratchSize || tokenPlanPrefixSlots(-1) != ErrFeatureScratchSize {
		t.Fatal("prefix size boundary")
	}
	for _, counts := range [][2]int{{-1, 0}, {maxInt, 1}, {maxInt/2 + 1, 2}, {maxInt / int(reflect.TypeOf(Feature{}).Size()), 1}} {
		if n, err := tokenPlanCrossSlots(counts[0], counts[1]); n != 0 || err != ErrFeatureScratchSize {
			t.Fatal("cross count/product/storage overflow")
		}
	}
	if n, err := tokenPlanCrossSlots(0, maxInt); n != 0 || err != nil {
		t.Fatal("empty cross size")
	}
	if q, err := PrepareFeatureQuery(FeatureText{}); q.ready || err != ErrFeatureScratchSize {
		t.Fatal("zero text accepted")
	}
	emptyQ, emptyD := tokenPlanPair(t, "", "")
	q, d := tokenPlanPair(t, "alpha", "beta")
	var absent *FeatureScratch
	if fs, err := absent.FeaturesPrepared(emptyQ, emptyD); fs != nil || err != ErrFeatureScratchSize {
		t.Fatal("nil scratch precedes empty input")
	}
	s, _ := NewFeatureScratch(0)
	if fs, err := s.FeaturesPrepared(emptyQ, d); fs != nil || err != nil {
		t.Fatal("prepared empty text is valid")
	}
	if fs, err := s.FeaturesPrepared(FeatureQuery{}, emptyD); fs != nil || err != ErrFeatureScratchSize {
		t.Fatal("zero query precedes empty row")
	}
	if fs, err := s.FeaturesPrepared(q, FeatureText{}); fs != nil || err != ErrFeatureScratchSize {
		t.Fatal("zero document precedes capacity")
	}
	bad := q
	bad.prefixes = nil
	if _, _, err := FeatureCardinalityPrepared(bad, emptyD); err != ErrFeatureScratchSize {
		t.Fatal("prefix shape precedes empty row")
	}
	s, _ = NewFeatureScratch(2)
	borrowed, err := s.FeaturesPrepared(q, d)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]Feature(nil), borrowed...)
	longQ, _ := tokenPlanPair(t, "alpha beta", "beta")
	if fs, err := s.FeaturesPrepared(longQ, d); fs != nil || err != ErrFeatureScratchCapacity || !scratchBitsEqual(borrowed, before) {
		t.Fatal("capacity failure wrote borrowed row")
	}
	if fs, err := s.FeaturesPrepared(q, d); err != nil || !scratchBitsEqual(fs, before) {
		t.Fatal("capacity error poisoned reuse")
	}
}

func TestFeatureTokenPlanSharedReadLocalScratch(t *testing.T) {
	const rawQuery = "z a z b"
	queryText, err := PrepareFeatureText(rawQuery)
	if err != nil {
		t.Fatal(err)
	}
	query, err := PrepareFeatureQuery(queryText)
	if err != nil {
		t.Fatal(err)
	}
	rawDocs := [7]string{"", "!!!🙂", "b a z", "z", "z z z a a b", "Å β 한글 １２", "left\x00right"}
	var docs [7]FeatureText
	var wanted [7][]Feature
	var slots [7]int
	var termsBefore, uniqueBefore [7][]string
	maxSlots := 0
	for i, raw := range rawDocs {
		docs[i], err = PrepareFeatureText(raw)
		if err != nil {
			t.Fatal(err)
		}
		wanted[i] = Features(rawQuery, raw)
		count, reserved, err := FeatureCardinalityPrepared(query, docs[i])
		if err != nil || count != len(wanted[i]) {
			t.Fatal("shared plan cardinality", err)
		}
		slots[i] = reserved
		maxSlots = max(maxSlots, reserved)
		termsBefore[i] = append([]string(nil), docs[i].ordered...)
		uniqueBefore[i] = append([]string(nil), docs[i].uniqueWords...)
	}
	queryTerms := append([]string(nil), query.text.ordered...)
	queryWords := append([]string(nil), query.text.uniqueWords...)
	queryPrefixes := append([]uint64(nil), query.prefixes...)
	var scratches [4]*FeatureScratch
	for i := range scratches {
		scratches[i], err = NewFeatureScratch(maxSlots)
		if err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	failures := make(chan string, len(scratches))
	for worker, scratch := range scratches {
		wg.Add(1)
		go func(worker int, scratch *FeatureScratch, shared FeatureQuery) {
			defer wg.Done()
			for step := 0; step < 3*len(docs); step++ {
				i := (worker + step) % len(docs)
				count, reserved, err := FeatureCardinalityPrepared(shared, docs[i])
				if err != nil || count != len(wanted[i]) || reserved != slots[i] {
					failures <- "concurrent cardinality changed"
					return
				}
				row, err := scratch.FeaturesPrepared(shared, docs[i])
				if err != nil || !scratchBitsEqual(row, wanted[i]) {
					failures <- "concurrent index/value bits/order/nilness changed"
					return
				}
			}
		}(worker, scratch, query)
	}
	wg.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
	if !reflect.DeepEqual(query.text.ordered, queryTerms) ||
		!reflect.DeepEqual(query.text.uniqueWords, queryWords) ||
		!reflect.DeepEqual(query.prefixes, queryPrefixes) {
		t.Fatal("concurrent readers mutated shared query plan")
	}
	for i := range docs {
		if !reflect.DeepEqual(docs[i].ordered, termsBefore[i]) ||
			!reflect.DeepEqual(docs[i].uniqueWords, uniqueBefore[i]) {
			t.Fatal("concurrent readers mutated shared document plan")
		}
	}
}
