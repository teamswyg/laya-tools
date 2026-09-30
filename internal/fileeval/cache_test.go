package fileeval

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

// referenceRank is the pre-cache two-index path, independent of Ranker.
func referenceRank(q string, paths []string) (out [3][]int, fallback bool, err error) {
	if !slices.IsSorted(paths) {
		return out, false, fmt.Errorf("unsorted")
	}
	a, e := hintsearch.NewPathIndex(paths)
	if e != nil {
		return out, false, e
	}
	ar, e := a.RankLongInto(q, hintsearch.Ranking{})
	if e != nil {
		return out, false, e
	}
	out[0] = ar.Order
	texts := make([]string, len(paths))
	for i, p := range paths {
		texts[i] = lexicalhint.NormalizeText(p)
	}
	b, e := hintsearch.NewPathTextIndex(paths, texts)
	var br hintsearch.Ranking
	if e == nil {
		br, e = b.RankLongInto(lexicalhint.NormalizeText(q), hintsearch.Ranking{})
	}
	if e != nil {
		out[1] = slices.Clone(ar.Order)
		out[2] = slices.Clone(ar.Order)
		return out, true, nil
	}
	out[1] = br.Order
	out[2], e = hintsearch.InterleavePathBaselineFirst(ar.Order, br.Order)
	return out, false, e
}

func TestCacheExactIdentityAndRecovery(t *testing.T) {
	var r Ranker
	first := []string{"src/HTTPReader.go", "src/read_file.go", "테스트/파일.go"}
	second := []string{"src/HTTPReader.go", "src/write_file.go", "테스트/파일.go"}
	var held [3][]int
	var snapshot [3][]int
	cases := []struct {
		q     string
		paths []string
	}{
		{"read http", first}, {"write_file", first}, {"", first},
		{strings.Repeat("aB", 50000), first}, {"read", first},
		{"file", second}, {"http", []string{"z", "a"}}, {"file", []string{"a", "a"}},
		{"테스트", first}, {"!!!", first}, {"file", nil},
	}
	for i, c := range cases {
		got, fb, e := r.Rank(c.q, c.paths)
		want, wfb, we := referenceRank(c.q, c.paths)
		if (e == nil) != (we == nil) || fb != wfb || (e == nil && !reflect.DeepEqual(got, want)) {
			t.Fatalf("case %d: fallback %v/%v error %v/%v", i, fb, wfb, e, we)
		}
		if i == 0 {
			held = got
			for j := range got {
				snapshot[j] = slices.Clone(got[j])
			}
		}
	}
	if !reflect.DeepEqual(held, snapshot) {
		t.Fatal("later query overwrote retained result")
	}
	if !slices.Equal(r.paths, first) || r.Stats().CatalogBuilds != 3 {
		t.Fatalf("cache failed replacement: %+v", r.Stats())
	}
	// Change the caller's slice in place, retaining its backing array.
	first[1] = "src/z.go"
	got, _, e := r.Rank("z", first)
	want, _, we := referenceRank("z", first)
	if e != nil || we != nil || !reflect.DeepEqual(got, want) || r.Stats().CatalogBuilds != 4 {
		t.Fatal("caller mutation reused stale index")
	}
}

func TestIndependentCacheWorkers(t *testing.T) {
	paths := []string{"src/HTTPReader.go", "src/read_file.go", "테스트/파일.go"}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			var r Ranker
			for i := range 20 {
				q := []string{"http", "read file", "테스트"}[i%3]
				got, fb, e := r.Rank(q, paths)
				want, wfb, we := referenceRank(q, paths)
				if e != nil || we != nil || fb != wfb || !reflect.DeepEqual(got, want) {
					t.Error("worker mismatch")
					return
				}
			}
			if r.Stats().CatalogBuilds != 1 || r.Stats().CatalogHits != 19 {
				t.Error("unexpected worker cache counters")
			}
		})
	}
	wg.Wait()
}

func BenchmarkRepeatedPathCatalog(b *testing.B) {
	paths := make([]string, 10000)
	for i := range paths {
		paths[i] = fmt.Sprintf("src/module%05d/HTTPReader.go", i)
	}
	for _, reuse := range []bool{false, true} {
		name := "reference"
		if reuse {
			name = "reuse"
		}
		b.Run(name, func(b *testing.B) {
			var r Ranker
			b.ReportAllocs()
			for b.Loop() {
				var err error
				if reuse {
					_, _, err = r.Rank("read http module42", paths)
				} else {
					_, _, err = referenceRank("read http module42", paths)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
