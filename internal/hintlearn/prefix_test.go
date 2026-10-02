package hintlearn

import (
	"hash/fnv"
	"math"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestFeaturePrefixPreservesByteHash(t *testing.T) {
	inputs := []string{"", "a", "a_b", "한글", "🙂", "a\x00b", strings.Repeat("long", 32)}
	for _, q := range inputs {
		for _, d := range inputs {
			want := fnv.New64a()
			want.Write([]byte(q))
			want.Write([]byte{0})
			want.Write([]byte(d))
			if got := hashFrom(hashFrom(hash(q), "\x00"), d); got != want.Sum64() {
				t.Fatalf("prefix changed byte hash for %q / %q", q, d)
			}
		}
	}
}

// The original concatenating feature path is retained only as a compatibility
// and benchmark control. Keep the original hash statically callable, as it was
// in production, so the control adds no per-pair indirect-call overhead.
func originalByteHash(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

func concatenatingFeatures(query, document string) []Feature {
	qw, dw := words(query), words(document)
	qt, dt := terms(qw), terms(dw)
	if len(qt) == 0 || len(dt) == 0 {
		return nil
	}
	fs := make([]Feature, 0, len(qt)*len(dt)+1)
	scale := 1 / math.Sqrt(float64(len(qt)*len(dt)))
	for _, q := range qt {
		for _, d := range dt {
			h := originalByteHash(q + "\x00" + d)
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
	return append(fs, Feature{0, float64(hits) / float64(max(1, len(qw)))})
}

func TestFeaturePrefixPreservesSparseValuesAndScores(t *testing.T) {
	fixtures := [][2]string{
		{"", "empty input"},
		{"retain active entries", "cache keeps active entries and removes inactive entries"},
		{"retain active entries", "cache removes active entries and keeps inactive entries"},
		{"한글 오류 반환 🙂", "한글 오류 반환과 빈 값 보존"},
		{strings.Repeat("repeat ", 32), strings.Repeat("repeat ", 32)},
		{"left\x00right long_word", "right left long word"},
		{"Find the string splitter that keeps empty quoted words and returns completed tokens before an unclosed quote or trailing escape error.", "Returns completed words before malformed quote errors and preserves empty quoted words."},
	}
	w := make([]float64, Dimension)
	for i := range w {
		w[i] = math.Sin(float64(i))
	}
	for _, fixture := range fixtures {
		want := concatenatingFeatures(fixture[0], fixture[1])
		got := Features(fixture[0], fixture[1])
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("sparse order, indices or values changed for %q", fixture[0])
		}
		if math.Float64bits(Score(w, got)) != math.Float64bits(Score(w, want)) {
			t.Fatalf("score bits changed for %q", fixture[0])
		}
	}
}

var featureBenchmarkResult []Feature

func BenchmarkFeaturePrefix(b *testing.B) {
	// Original public synthetic strings; no corpus, model, source API or fit.
	fixtures := []struct{ name, query, document string }{
		{"short", "retain active entries", "cache keeps active entries and removes inactive entries"},
		{"bounded32", "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho sigma tau upsilon phi chi psi omega red green blue white black gray orange violet", "violet orange gray black white blue green red omega psi chi phi upsilon tau sigma rho pi omicron xi nu mu lambda kappa iota theta eta zeta epsilon delta gamma beta alpha"},
	}
	for _, fixture := range fixtures {
		b.Run(fixture.name+"/concatenating", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				featureBenchmarkResult = concatenatingFeatures(fixture.query, fixture.document)
			}
		})
		b.Run(fixture.name+"/prefix", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				featureBenchmarkResult = Features(fixture.query, fixture.document)
			}
		})
	}
}
