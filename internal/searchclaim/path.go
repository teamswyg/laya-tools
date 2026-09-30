package searchclaim

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

const PathDimension = Dimension + SpreadDimension
const PathSchema = "riido-path-claim-v1"

// PathFeatureNames returns an owned schema array, not mutable shared scratch.
func PathFeatureNames() [PathDimension]string {
	var names [PathDimension]string
	copy(names[:Dimension], Names[:])
	copy(names[Dimension:], SpreadNames[:])
	return names
}

// PathFeatures is an opt-in contract for index-produced path rankings and full
// UTF-8 queries. It reads no labels, repository IDs, paths or auxiliary ranking.
// The byte feature saturates at the legacy 8KiB scale. Large inputs require a
// newly validated model; old small-catalog weights are not implicitly enabled.
func PathFeatures(query string, baseline hintsearch.Ranking) ([PathDimension]float64, error) {
	var out [PathDimension]float64
	if !utf8.ValidString(query) {
		return out, fmt.Errorf("invalid UTF8 path query")
	}
	base, e := features(query, baseline, hintsearch.MaxLongQueryBytes, hintsearch.MaxPathDocuments)
	if e != nil {
		return out, e
	}
	spread, e := scoreSpread(baseline, hintsearch.MaxPathDocuments)
	if e != nil {
		return out, e
	}
	copy(out[:Dimension], base[:])
	copy(out[Dimension:], spread[:])
	for _, v := range out {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return [PathDimension]float64{}, fmt.Errorf("invalid normalized path feature")
		}
	}
	return out, nil
}
