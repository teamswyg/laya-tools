package fileeval

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/teamswyg/laya-tools/internal/searchclaim"
)

func choose(use bool) Selector {
	return func([searchclaim.PathDimension]float64) (bool, error) { return use, nil }
}
func permutation(t *testing.T, order []int, n int) {
	t.Helper()
	v := slices.Clone(order)
	slices.Sort(v)
	if len(v) != n {
		t.Fatal("lost candidates")
	}
	for i, x := range v {
		if i != x {
			t.Fatal("not a permutation")
		}
	}
}
func TestSelectedReferenceParityAndSkippedWork(t *testing.T) {
	paths := []string{"src/HTTPReader.go", "src/read_file.go", "test/HTTPReader_test.go", "한글/파일.go"}
	for _, q := range []string{"http reader", "read_file", "한글", "!!!", strings.Repeat("query ", 2000)} {
		ref, fb, e := referenceRank(q, paths)
		if e != nil {
			t.Fatal("fixture reference", e)
		}
		off, e := RankSelected(q, paths, choose(false))
		if e != nil || !slices.Equal(off.Order, ref[Baseline]) || off.Work != (WorkStats{BaselineRankAttempts: 1}) || !off.FeaturesAvailable || off.AuxiliaryRequested || off.AuxiliaryUsed {
			t.Fatalf("skip %+v %v", off, e)
		}
		nilPolicy, e := RankSelected(q, paths, nil)
		if e != nil || nilPolicy.FeaturesAvailable || nilPolicy.Work != (WorkStats{BaselineRankAttempts: 1}) || !slices.Equal(nilPolicy.Order, ref[Baseline]) {
			t.Fatal("nil selector did work", nilPolicy, e)
		}
		on, e := RankSelected(q, paths, choose(true))
		if e != nil || !slices.Equal(on.Order, ref[Interleaved]) || on.AuxiliaryUsed == fb || on.Work != (WorkStats{1, 1, 1}) {
			t.Fatalf("on %+v %v", on, e)
		}
		if on.Order[0] != ref[Baseline][0] {
			t.Fatal("lost first baseline")
		}
		for _, out := range []Selection{off, nilPolicy, on} {
			if !slices.Equal(out.BaselineOrder, ref[Baseline]) {
				t.Fatal("paired baseline differs from independent reference")
			}
		}
		permutation(t, on.Order, len(paths))
	}
}
func TestSelectedLazyReuseMutationAndOwnership(t *testing.T) {
	var r Ranker
	paths := []string{"a/ReadFile.go", "b/WriteFile.go"}
	first, e := r.RankSelected("read file", paths, choose(false))
	if e != nil || r.aux != nil || r.auxReady {
		t.Fatal("eager helper", e)
	}
	held := slices.Clone(first.Order)
	heldBaseline := slices.Clone(first.BaselineOrder)
	second, e := r.RankSelected("read file", paths, func(f [searchclaim.PathDimension]float64) (bool, error) { f[0] = 0; return true, nil })
	if e != nil || r.aux == nil || !r.auxReady || second.Work != (WorkStats{1, 1, 1}) || second.Features[0] != 1 {
		t.Fatal("lazy first use or feature alias", second, e)
	}
	third, e := r.RankSelected("write", paths, choose(true))
	if e != nil || third.Work != (WorkStats{1, 0, 1}) {
		t.Fatal("reuse", third, e)
	}
	fourth, e := r.RankSelected("read", paths, choose(false))
	if e != nil || fourth.Work != (WorkStats{1, 0, 0}) {
		t.Fatal("skipped warm helper executed", fourth, e)
	}
	paths[1] = "b/NewFile.go"
	fifth, e := r.RankSelected("new file", paths, choose(true))
	if e != nil || fifth.Work != (WorkStats{1, 1, 1}) {
		t.Fatal("stale catalog", fifth, e)
	}
	ref, _, e := referenceRank("new file", paths)
	if e != nil || !slices.Equal(fifth.Order, ref[Interleaved]) {
		t.Fatal("changed snapshot mismatch")
	}
	if !slices.Equal(first.Order, held) || !slices.Equal(first.BaselineOrder, heldBaseline) {
		t.Fatal("later work changed earlier output")
	}
	if r.WorkStats() != (WorkStats{5, 2, 3}) {
		t.Fatal("actual work totals", r.WorkStats())
	}
}
func TestSelectedFallbackAndInvalidBaseline(t *testing.T) {
	paths := []string{"a.go", "b.go"}
	failed, e := RankSelected("query", paths, func([searchclaim.PathDimension]float64) (bool, error) {
		return true, errors.New("private selector error")
	})
	base, _, _ := referenceRank("query", paths)
	if e != nil || failed.Fallback != "selector_error" || failed.AuxiliaryRequested || failed.AuxiliaryUsed || failed.Work != (WorkStats{1, 0, 0}) || !slices.Equal(failed.Order, base[Baseline]) {
		t.Fatal(failed, e)
	}
	q := strings.Repeat("aB", 50000)
	ref, fb, e := referenceRank(q, paths)
	if e != nil || !fb {
		t.Fatal("fixture must exceed normalized query bound")
	}
	fallback, e := RankSelected(q, paths, choose(true))
	if e != nil || fallback.Fallback != "auxiliary_error" || !fallback.AuxiliaryRequested || fallback.AuxiliaryUsed || fallback.Work != (WorkStats{1, 1, 1}) || !slices.Equal(fallback.Order, ref[Baseline]) {
		t.Fatal(fallback, e)
	}
	for _, bad := range [][]string{nil, {"same", "same"}, {"z", "a"}} {
		calls := 0
		if _, e := RankSelected("query", bad, func([searchclaim.PathDimension]float64) (bool, error) { calls++; return true, nil }); e == nil || calls != 0 {
			t.Fatal("invalid baseline reached selector")
		}
	}
}
func TestSelectedLargeCatalogAndIndependentWorkers(t *testing.T) {
	paths := make([]string, 10000)
	for i := range paths {
		paths[i] = fmt.Sprintf("src/File%05d.go", i)
	}
	q := strings.Repeat("file ", 2000)
	ref, fb, e := referenceRank(q, paths)
	if e != nil || fb {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var r Ranker
			for _, use := range []bool{false, true, false, true} {
				out, e := r.RankSelected(q, paths, choose(use))
				if e != nil {
					t.Error(e)
					return
				}
				want := ref[Baseline]
				if use {
					want = ref[Interleaved]
				}
				if !reflect.DeepEqual(out.Order, want) {
					t.Error("worker ranking mismatch")
				}
				if !use && (out.Work.AuxiliaryBuildAttempts != 0 || out.Work.AuxiliaryRankAttempts != 0) {
					t.Error("unselected work executed")
				}
			}
		}()
	}
	wg.Wait()
}
