package hintweights

import (
	"runtime"
	"testing"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
)

// Public synthetic benchmark fixtures; never trained-model or HF assets.
// Controls call the reference directly, with conversion outside timed loops.
func BenchmarkPreparedConstruct(b *testing.B) {
	for _, f := range fixtures() {
		kind, raw := f.name, f.bytes
		if kind != "fp32" && kind != "int8" && kind != "ternary_edges" && kind != "ternary_dense" {
			continue
		}
		b.Run(kind+"/legacy_decode", func(b *testing.B) {
			b.ReportAllocs()
			var w []float64
			for b.Loop() {
				var e error
				w, e = hintlearn.Decode(raw)
				if e != nil {
					b.Fatal(e)
				}
			}
			runtime.KeepAlive(w)
		})
		b.Run(kind+"/packed_owned", func(b *testing.B) {
			b.ReportAllocs()
			var v *View
			for b.Loop() {
				var e error
				v, e = New(raw)
				if e != nil {
					b.Fatal(e)
				}
			}
			runtime.KeepAlive(v)
		})
	}
}
func BenchmarkPreparedScore(b *testing.B) {
	fs := syntheticFeatures()
	refFeatures := referenceFeatures(fs)
	for _, f := range fixtures() {
		kind, raw := f.name, f.bytes
		if kind != "fp32" && kind != "int8" && kind != "ternary_edges" && kind != "ternary_dense" {
			continue
		}
		w, e := hintlearn.Decode(raw)
		if e != nil {
			b.Fatal(e)
		}
		v, e := New(raw)
		if e != nil {
			b.Fatal(e)
		}
		b.Run(kind+"/legacy_direct", func(b *testing.B) {
			b.ReportAllocs()
			var score float64
			for b.Loop() {
				score = hintlearn.Score(w, refFeatures)
			}
			runtime.KeepAlive(score)
		})
		b.Run(kind+"/packed_direct", func(b *testing.B) {
			b.ReportAllocs()
			var score float64
			for b.Loop() {
				var e error
				score, e = v.Score(fs)
				if e != nil {
					b.Fatal(e)
				}
			}
			runtime.KeepAlive(score)
		})
	}
}
