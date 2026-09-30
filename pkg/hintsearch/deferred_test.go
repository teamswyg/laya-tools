package hintsearch

import (
	"slices"
	"testing"
)

func TestInterleaveAfterPrefix(t *testing.T) {
	// Exhaust permutations and all prefix boundaries on a small catalog.
	base := []int{0, 1, 2, 3, 4, 5}
	var visit func(int)
	visit = func(start int) {
		if start < len(base) {
			for i := start; i < len(base); i++ {
				base[start], base[i] = base[i], base[start]
				visit(start + 1)
				base[start], base[i] = base[i], base[start]
			}
			return
		}
		for _, hints := range [][]int{nil, {5, 2, 0}, {5, 4, 3, 2, 1, 0}} {
			before := slices.Clone(base)
			hb := slices.Clone(hints)
			merged, err := InterleaveBaselineFirst(base, hints)
			if err != nil {
				t.Fatal(err)
			}
			for prefix := 0; prefix <= len(base); prefix++ {
				want := slices.Clone(base[:prefix])
				for _, id := range merged {
					if !slices.Contains(base[:prefix], id) {
						want = append(want, id)
					}
				}
				got, err := InterleaveAfterPrefix(base, hints, prefix)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(got, want) || !slices.Equal(base, before) || !slices.Equal(hints, hb) {
					t.Fatalf("prefix %d got %v want %v", prefix, got, want)
				}
			}
		}
	}
	visit(0)
	for _, tc := range []struct {
		base, hints []int
		prefix      int
	}{{nil, nil, 0}, {[]int{0, 0}, nil, 1}, {[]int{0, 1}, []int{1, 1}, 1}, {[]int{0, 1}, []int{2}, 1}, {[]int{0, 1}, nil, -1}, {[]int{0, 1}, nil, 3}} {
		if _, err := InterleaveAfterPrefix(tc.base, tc.hints, tc.prefix); err == nil {
			t.Fatalf("accepted malformed input: %+v", tc)
		}
	}
}

func BenchmarkInterleaveAfterPrefix3009(b *testing.B) {
	base, hints := make([]int, 3009), make([]int, 3009)
	for i := range base {
		base[i] = i
		hints[i] = len(base) - 1 - i
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := InterleaveAfterPrefix(base, hints, 20); err != nil {
			b.Fatal(err)
		}
	}
}
