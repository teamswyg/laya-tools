// Package fileeval separates label-free path ranking from label-based scoring.
package fileeval

import (
	"fmt"
	"slices"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const Baseline = 0
const Normalized = 1
const Interleaved = 2

// Rank has no label input. Every order retains the full sorted path catalog.
func Rank(query string, paths []string) (orders [3][]int, auxFailed bool, err error) {
	if !slices.IsSorted(paths) {
		return orders, false, fmt.Errorf("require sorted path identities")
	}
	idx, e := hintsearch.NewPathIndex(paths)
	if e != nil {
		return orders, false, e
	}
	base, e := idx.RankLongInto(query, hintsearch.Ranking{})
	if e != nil {
		return orders, false, e
	}
	orders[Baseline] = base.Order
	texts := make([]string, len(paths))
	for i, p := range paths {
		texts[i] = lexicalhint.NormalizeText(p)
	}
	aux, e := hintsearch.NewPathTextIndex(paths, texts)
	var ranked hintsearch.Ranking
	if e == nil {
		ranked, e = aux.RankLongInto(lexicalhint.NormalizeText(query), hintsearch.Ranking{})
	}
	if e != nil {
		orders[Normalized] = slices.Clone(base.Order)
		orders[Interleaved] = slices.Clone(base.Order)
		return orders, true, nil
	}
	orders[Normalized] = ranked.Order
	orders[Interleaved], e = hintsearch.InterleavePathBaselineFirst(base.Order, ranked.Order)
	return orders, false, e
}

type Score struct {
	Targets, Mapped, FirstRank int
	Hit1, Hit10, Hit100, All10 bool
	ReciprocalRank             float64
}

// Measure accepts labels only after rankings have been constructed.
func Measure(paths []string, order []int, gold []string) (Score, error) {
	var s Score
	if !slices.IsSorted(paths) || !slices.IsSorted(gold) || len(order) != len(paths) || len(paths) == 0 {
		return s, fmt.Errorf("invalid scoring inputs")
	}
	for i := 1; i < len(paths); i++ {
		if paths[i] == paths[i-1] {
			return s, fmt.Errorf("duplicate candidate")
		}
	}
	for i := 1; i < len(gold); i++ {
		if gold[i] == gold[i-1] {
			return s, fmt.Errorf("duplicate label")
		}
	}
	ranks := make([]int, len(paths))
	for i, d := range order {
		if d < 0 || d >= len(paths) || ranks[d] != 0 {
			return s, fmt.Errorf("invalid ranking permutation")
		}
		ranks[d] = i + 1
	}
	s.Targets = len(gold)
	last := 0
	for _, g := range gold {
		i, found := slices.BinarySearch(paths, g)
		if !found {
			continue
		}
		s.Mapped++
		r := ranks[i]
		if s.FirstRank == 0 || r < s.FirstRank {
			s.FirstRank = r
		}
		last = max(last, r)
	}
	s.Hit1 = s.FirstRank > 0 && s.FirstRank <= 1
	s.Hit10 = s.FirstRank > 0 && s.FirstRank <= 10
	s.Hit100 = s.FirstRank > 0 && s.FirstRank <= 100
	s.All10 = s.Targets > 0 && s.Mapped == s.Targets && last <= 10
	if s.FirstRank > 0 {
		s.ReciprocalRank = 1 / float64(s.FirstRank)
	}
	return s, nil
}
