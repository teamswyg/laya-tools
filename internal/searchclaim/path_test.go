package searchclaim

import (
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

func pathRanks(n int) hintsearch.Ranking {
	r := hintsearch.Ranking{Order: make([]int, n), Scores: make([]float64, n)}
	for i := range r.Order {
		r.Order[i] = i
		r.Scores[i] = float64(n - i)
	}
	return r
}
func TestPathFeatureLegacyParity(t *testing.T) {
	for _, n := range []int{1, 2, 19, 20, 21, 4096} {
		for _, q := range []string{"someName_123 한글", strings.Repeat("a", hintsearch.MaxQueryBytes), "😀 hello", "!!!"} {
			rank := pathRanks(n)
			want, e := Features(q, rank)
			if e != nil {
				t.Fatal(e)
			}
			spread, e := ScoreSpread(rank)
			if e != nil {
				t.Fatal(e)
			}
			got, e := PathFeatures(q, rank)
			if e != nil {
				t.Fatal(e)
			}
			if !slices.Equal(got[:Dimension], want[:]) || !slices.Equal(got[Dimension:], spread[:]) {
				t.Fatal("short input changed")
			}
		}
	}
}
func TestPathFeatureLargeInputAndImmutability(t *testing.T) {
	q := strings.Repeat("긴문장 Query42_ ", 4096)
	rank := pathRanks(10000)
	before := hintsearch.Ranking{Order: slices.Clone(rank.Order), Scores: slices.Clone(rank.Scores)}
	got, e := PathFeatures(q, rank)
	if e != nil {
		t.Fatal(e)
	}
	again, e := PathFeatures(q, rank)
	if e != nil || got != again {
		t.Fatal("nondeterministic", e)
	}
	if got[2] != 1 || !reflect.DeepEqual(rank, before) {
		t.Fatal("saturation or caller mutation")
	}
	for _, v := range got {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			t.Fatal("out of range")
		}
	}
	if _, e := Features(q, rank); e == nil {
		t.Fatal("widened old Features")
	}
	if _, e := ScoreSpread(rank); e == nil {
		t.Fatal("widened old ScoreSpread")
	}
	names := PathFeatureNames()
	names[0] = "changed"
	if PathFeatureNames()[0] != "bias" {
		t.Fatal("schema aliases shared array")
	}
}
func TestPathFeatureBounds(t *testing.T) {
	if _, e := PathFeatures(strings.Repeat("a", hintsearch.MaxLongQueryBytes), pathRanks(hintsearch.MaxPathDocuments)); e != nil {
		t.Fatal(e)
	}
	for _, q := range []string{"", " \t\n", string([]byte{255}), strings.Repeat("a", hintsearch.MaxLongQueryBytes+1)} {
		if _, e := PathFeatures(q, pathRanks(2)); e == nil {
			t.Fatal("accepted invalid query")
		}
	}
	if _, e := PathFeatures("q", pathRanks(hintsearch.MaxPathDocuments+1)); e == nil {
		t.Fatal("accepted oversized ranking")
	}
	for _, change := range []func(*hintsearch.Ranking){
		func(r *hintsearch.Ranking) { r.Order[0] = -1 },
		func(r *hintsearch.Ranking) { r.Scores[0] = math.NaN() },
		func(r *hintsearch.Ranking) { r.Scores[30] = math.Inf(1) },
		func(r *hintsearch.Ranking) { r.Scores[3] = -1 },
		func(r *hintsearch.Ranking) { r.Scores[2] = r.Scores[0] + 1 },
	} {
		r := pathRanks(100)
		change(&r)
		if _, e := PathFeatures("q", r); e == nil {
			t.Fatal("accepted malformed ranking")
		}
	}
	r := pathRanks(21)
	clear(r.Scores)
	v, e := PathFeatures("q", r)
	if e != nil {
		t.Fatal(e)
	}
	for _, x := range v[8:] {
		if x != 0 {
			t.Fatal("zero scores produced nonzero signal")
		}
	}
}
