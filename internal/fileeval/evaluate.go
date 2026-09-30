// Package fileeval separates label-free path ranking from label-based scoring.
package fileeval

import (
	"fmt"
	"slices"
	"strings"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const Baseline = 0
const Normalized = 1
const Interleaved = 2

// Ranker retains at most one immutable index pair for an exact ordered path
// catalog. It belongs to one worker: do not call its methods concurrently.
// Results are independently owned; no ranking buffers or labels are cached.
type Ranker struct {
	paths     []string
	base, aux *hintsearch.Index
	auxReady  bool
	stats     CacheStats
	work      WorkStats
}

type CacheStats struct{ CatalogBuilds, CatalogHits int }

func (r *Ranker) Stats() CacheStats { return r.stats }

// Rank has no label input. Every order retains the full sorted path catalog.
// This standalone form retains no cache between calls.
func Rank(query string, paths []string) ([3][]int, bool, error) {
	return new(Ranker).rank(query, paths, false)
}

// Rank reuses indexes only when every original path matches in order. Full
// equality, rather than repository names or a hash, establishes cache identity.
func (r *Ranker) Rank(query string, paths []string) ([3][]int, bool, error) {
	return r.rank(query, paths, true)
}

func (r *Ranker) baseline(query string, paths []string, retain bool) (hintsearch.Ranking, error) {
	if !slices.IsSorted(paths) {
		return hintsearch.Ranking{}, fmt.Errorf("require sorted path identities")
	}
	if r.base == nil || !slices.Equal(r.paths, paths) {
		idx, e := hintsearch.NewPathIndex(paths)
		if e != nil {
			return hintsearch.Ranking{}, e
		}
		// Own the slice and strings so caller replacement cannot change cache keys.
		var owned []string
		if retain {
			owned = make([]string, len(paths))
			for i, p := range paths {
				owned[i] = strings.Clone(p)
			}
		}
		r.paths, r.base, r.aux, r.auxReady = owned, idx, nil, false
		r.stats.CatalogBuilds++
	} else {
		r.stats.CatalogHits++
	}
	r.work.BaselineRankAttempts++
	base, e := r.base.RankLongInto(query, hintsearch.Ranking{})
	if e != nil {
		return hintsearch.Ranking{}, e
	}
	return base, nil
}
func (r *Ranker) auxiliary(query string, paths []string) (hintsearch.Ranking, error) {
	var e error
	if !r.auxReady {
		r.work.AuxiliaryBuildAttempts++
		texts := make([]string, len(paths))
		for i, p := range paths {
			texts[i] = lexicalhint.NormalizeText(p)
		}
		r.aux, e = hintsearch.NewPathTextIndex(paths, texts)
		r.auxReady = true
	}
	var ranked hintsearch.Ranking
	if r.aux != nil {
		r.work.AuxiliaryRankAttempts++
		ranked, e = r.aux.RankLongInto(lexicalhint.NormalizeText(query), hintsearch.Ranking{})
	} else {
		e = fmt.Errorf("normalized catalog unavailable")
	}
	return ranked, e
}
func (r *Ranker) rank(query string, paths []string, retain bool) (orders [3][]int, auxFailed bool, err error) {
	base, e := r.baseline(query, paths, retain)
	if e != nil {
		return orders, false, e
	}
	orders[Baseline] = base.Order
	ranked, e := r.auxiliary(query, paths)
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
