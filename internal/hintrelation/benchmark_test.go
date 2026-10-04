// SPDX-License-Identifier: Apache-2.0
package hintrelation

import (
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/hintsymbol"
)

// This microbenchmark repeats one supported public-shaped text eight times.
// It measures extraction APIs, not discrimination, learning, source behavior,
// native/GPU memory, end-to-end verification, or production savings. The warm
// legacy control excludes one-time token plans and scratch allocation.
const benchRequest = "Return keys in increasing order, keeping x >= lower and x < upper."
const benchCandidate = "Ascend calls the iterator for every value in the tree within the range [pivot, last], until iterator returns false."

var benchmarkEntries int
var benchmarkColumns Columns

func BenchmarkRepresentations(b *testing.B) {
	var candidates [MaxCandidates]string
	for i := range candidates {
		candidates[i] = benchCandidate
	}
	b.Run("legacy_raw_eight", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			n := 0
			for _, text := range candidates {
				n += len(hintlearn.Features(benchRequest, text))
			}
			benchmarkEntries = n
		}
	})
	b.Run("symbol_raw_eight", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			n := 0
			for _, text := range candidates {
				features, err := hintsymbol.Extract(benchRequest, text)
				if err != nil {
					b.Fatal(err)
				}
				n += len(features)
			}
			benchmarkEntries = n
		}
	})
	b.Run("legacy_prepared_warm_eight", func(b *testing.B) {
		q, err := hintlearn.PrepareFeatureText(benchRequest)
		if err != nil {
			b.Fatal(err)
		}
		query, err := hintlearn.PrepareFeatureQuery(q)
		if err != nil {
			b.Fatal(err)
		}
		document, err := hintlearn.PrepareFeatureText(benchCandidate)
		if err != nil {
			b.Fatal(err)
		}
		_, slots, err := hintlearn.FeatureCardinalityPrepared(query, document)
		if err != nil {
			b.Fatal(err)
		}
		scratch, err := hintlearn.NewFeatureScratch(slots)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			n := 0
			for range MaxCandidates {
				features, err := scratch.FeaturesPrepared(query, document)
				if err != nil {
					b.Fatal(err)
				}
				n += len(features)
			}
			benchmarkEntries = n
		}
	})
	b.Run("relation_parse_eight", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			columns, err := Build(benchRequest, candidates, len(candidates))
			if err != nil {
				b.Fatal(err)
			}
			benchmarkColumns = columns
		}
	})
}
