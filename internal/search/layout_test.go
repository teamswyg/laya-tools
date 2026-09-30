package search

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func largeFixture(t testing.TB) string {
	t.Helper()
	root := fixture(t)
	for i := 0; i < 256; i++ {
		body := fmt.Sprintf("package fixture\n// group%d authorization redirect handler\n", i%16) + strings.Repeat("func processRequest() {} // validate token request headers\n", 48)
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("fixture%03d.go", i)), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
func BenchmarkSearchLayout(b *testing.B) {
	idx, err := Load(largeFixture(b))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := idx.Search("group7 authorization authorization", "", 8, 3, nil); err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkIndexLayout(b *testing.B) {
	root := largeFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Load(root); err != nil {
			b.Fatal(err)
		}
	}
}

// Independent map-based BM25 oracle preserves legacy scores and ordering.
func TestLayoutMatchesReference(t *testing.T) {
	idx, err := Load(largeFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"group7 authorization authorization", "token request headers", "missingword", "한글", "processRequest"} {
		counts := make([]map[string]int, len(idx.Chunks))
		df := map[string]int{}
		lengths := make([]int, len(counts))
		avg := 0.0
		for i, c := range idx.Chunks {
			counts[i] = map[string]int{}
			ts := Terms(c.Path + "\n" + c.Text)
			lengths[i] = len(ts)
			avg += float64(len(ts))
			for _, term := range ts {
				counts[i][term]++
			}
			for term := range counts[i] {
				df[term]++
			}
		}
		avg /= float64(len(counts))
		var expected []Chunk
		for i, c := range idx.Chunks {
			seen := map[string]bool{}
			score := 0.0
			for _, term := range Terms(q) {
				if seen[term] {
					continue
				}
				seen[term] = true
				f := float64(counts[i][term])
				if f == 0 {
					continue
				}
				idf := math.Log(1 + (float64(len(counts)-df[term])+.5)/(float64(df[term])+.5))
				score += idf * f * 2.2 / (f + 1.2*(.25+.75*float64(lengths[i])/avg))
			}
			if score > 0 {
				c.Score = score
				expected = append(expected, c)
			}
		}
		sortChunks(expected)
		expected = expected[:min(8, len(expected))]
		var selected []Chunk
		for _, c := range expected {
			duplicate := false
			for _, picked := range selected {
				overlap := min(c.End, picked.End) - max(c.Start, picked.Start) + 1
				if c.Path == picked.Path && overlap > min(c.End-c.Start+1, picked.End-picked.Start+1)/2 {
					duplicate = true
					break
				}
			}
			if !duplicate {
				selected = append(selected, c)
			}
			if len(selected) == 3 {
				break
			}
		}
		got, err := idx.Search(q, "", 8, 3, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Results) != len(selected) {
			t.Fatalf("%s length mismatch", q)
		}
		for i, c := range selected {
			g := got.Results[i]
			if g.Path != c.Path || g.Start != c.Start || math.Abs(g.Score-c.Score) > 1e-12 {
				t.Fatalf("%s mismatch: %+v vs %+v", q, g, c)
			}
		}
	}
	t.Run("parallel immutable reads", func(t *testing.T) {
		for i := 0; i < 8; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				if _, err := idx.Search("authorization", "", 8, 3, nil); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}
