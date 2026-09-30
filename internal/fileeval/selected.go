package fileeval

import (
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

// WorkStats counts attempted real operations, including failed attempts. It does
// not estimate time, memory, tokens or downstream model calls.
type WorkStats struct {
	BaselineRankAttempts, AuxiliaryBuildAttempts, AuxiliaryRankAttempts int
}

func (r *Ranker) WorkStats() WorkStats { return r.work }
func workDelta(after, before WorkStats) WorkStats {
	return WorkStats{after.BaselineRankAttempts - before.BaselineRankAttempts, after.AuxiliaryBuildAttempts - before.AuxiliaryBuildAttempts, after.AuxiliaryRankAttempts - before.AuxiliaryRankAttempts}
}

// Selector receives only an owned fixed-size array of inference features. An
// error declines auxiliary work and keeps baseline candidates. Selectors must
// not mutate the input path slice or recursively use the same Ranker, which
// belongs to one sequential worker.
type Selector func([searchclaim.PathDimension]float64) (bool, error)

type Selection struct {
	Order                                                []int
	Features                                             [searchclaim.PathDimension]float64
	FeaturesAvailable, AuxiliaryRequested, AuxiliaryUsed bool
	Fallback                                             string
	Work                                                 WorkStats
}

// RankSelected retains no index across calls. A nil selector is baseline-only;
// it does not extract features or build/rank the auxiliary index.
func RankSelected(query string, paths []string, selector Selector) (Selection, error) {
	return new(Ranker).selected(query, paths, selector, false)
}

// RankSelected opts into the existing exact-catalog index reuse. It must not be
// called concurrently on one Ranker; independent workers need no shared locks.
func (r *Ranker) RankSelected(query string, paths []string, selector Selector) (Selection, error) {
	return r.selected(query, paths, selector, true)
}
func (r *Ranker) selected(query string, paths []string, selector Selector, retain bool) (result Selection, err error) {
	before := r.work
	defer func() { result.Work = workDelta(r.work, before) }()
	base, e := r.baseline(query, paths, retain)
	if e != nil {
		return result, e
	}
	result.Order = base.Order
	if selector == nil {
		return result, nil
	}
	result.Features, e = searchclaim.PathFeatures(query, base)
	if e != nil {
		result.Fallback = "feature_error"
		return result, nil
	}
	result.FeaturesAvailable = true
	use, e := selector(result.Features)
	if e != nil {
		result.Fallback = "selector_error"
		return result, nil
	}
	result.AuxiliaryRequested = use
	if !use {
		return result, nil
	}
	aux, e := r.auxiliary(query, paths)
	if e != nil {
		result.Fallback = "auxiliary_error"
		return result, nil
	}
	order, e := hintsearch.InterleavePathBaselineFirst(base.Order, aux.Order)
	if e != nil {
		result.Fallback = "interleave_error"
		return result, nil
	}
	result.Order = order
	result.AuxiliaryUsed = true
	return result, nil
}
