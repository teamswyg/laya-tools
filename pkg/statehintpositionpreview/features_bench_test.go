// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintpositionpreview

import (
	"runtime"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

// Each timed operation extracts one of the same 12 locked original texts,
// cycled in public pair order. Input parsing and workspace setup are untimed.
func BenchmarkExtractOriginalTwelve(b *testing.B) {
	pairs := originalPairs(b)
	var texts [12]string
	for i, pair := range pairs {
		texts[2*i], texts[2*i+1] = pair.A.Text, pair.B.Text
	}
	b.Run("v2", func(b *testing.B) {
		var workspace statehintwide.Workspace
		index, sink := 0, 0
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			view, err := statehintwide.ExtractContextual(texts[index], &workspace)
			if err != nil {
				b.Fatal(err)
			}
			sink += view.Len() + view.WordCount()
			index = (index + 1) % len(texts)
		}
		runtime.KeepAlive(sink)
	})
	b.Run("position_preview", func(b *testing.B) {
		var workspace Workspace
		index, sink := 0, 0
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			view, err := Extract(texts[index], &workspace)
			if err != nil {
				b.Fatal(err)
			}
			sink += view.Len() + view.WordCount()
			index = (index + 1) % len(texts)
		}
		runtime.KeepAlive(sink)
	})
}
