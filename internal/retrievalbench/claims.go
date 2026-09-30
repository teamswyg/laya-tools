package retrievalbench

import (
	"fmt"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
)

// ClaimExample contains no query/code text or gold function names. Repository
// and outcome ranks are split/label metadata, never runtime feature inputs.
type ClaimExample struct {
	Features                                     [searchclaim.Dimension]float64
	Repository                                   string
	Group                                        int
	BaselineRank, AuxiliaryRank, InterleavedRank int
}

func PrepareClaims(rows []Row) ([]ClaimExample, error) {
	return prepareClaims(rows, false)
}

func PrepareBaselineFirstClaims(rows []Row) ([]ClaimExample, error) {
	return prepareClaims(rows, true)
}

func prepareClaims(rows []Row, baselineFirst bool) ([]ClaimExample, error) {
	return prepareClaimsWithSpread(rows, baselineFirst, nil)
}

// PrepareBaselineFirstSpreadClaims keeps additional features in a parallel
// fixed-width array, preserving the v1 example and runtime feature contracts.
func PrepareBaselineFirstSpreadClaims(rows []Row) ([]ClaimExample, [][searchclaim.SpreadDimension]float64, error) {
	var extra [][searchclaim.SpreadDimension]float64
	xs, err := prepareClaimsWithSpread(rows, true, &extra)
	if err != nil {
		return nil, nil, err
	}
	return xs, extra, nil
}
func prepareClaimsWithSpread(rows []Row, baselineFirst bool, extra *[][searchclaim.SpreadDimension]float64) ([]ClaimExample, error) {
	docs := make([]string, len(rows))
	normalized := make([]string, len(rows))
	for i, r := range rows {
		docs[i] = r.Code
		normalized[i] = lexicalhint.NormalizeText(r.Code)
	}
	base, e := hintsearch.New(docs)
	if e != nil {
		return nil, e
	}
	aux, e := hintsearch.New(normalized)
	if e != nil {
		return nil, e
	}
	var baseline, additional hintsearch.Ranking
	qs := queries(rows)
	out := make([]ClaimExample, 0, len(qs))
	for i, q := range qs {
		if q.repository == "multiple" {
			return nil, fmt.Errorf("cross-repository query group requires a new split plan")
		}
		baseline, e = base.RankInto(q.text, baseline)
		if e != nil {
			return nil, e
		}
		// Features are frozen before any target rank or auxiliary outcome is read.
		f, e := searchclaim.Features(q.text, baseline)
		if e != nil {
			return nil, e
		}
		if extra != nil {
			spread, err := searchclaim.ScoreSpread(baseline)
			if err != nil {
				return nil, err
			}
			*extra = append(*extra, spread)
		}
		additional, e = aux.RankInto(lexicalhint.NormalizeText(q.text), additional)
		if e != nil {
			return nil, e
		}
		var order []int
		if baselineFirst {
			order, e = hintsearch.InterleaveBaselineFirst(baseline.Order, additional.Order)
		} else {
			order, e = hintsearch.Interleave(baseline.Order, additional.Order)
		}
		if e != nil {
			return nil, e
		}
		rank := func(order []int) int {
			for p, id := range order {
				for _, target := range q.targets {
					if id == target {
						return p + 1
					}
				}
			}
			return len(order) + 1
		}
		br, ar, ir := rank(baseline.Order), rank(additional.Order), rank(order)
		bound := 2 * br
		if baselineFirst {
			bound--
		}
		if ir > min(len(rows), bound) {
			return nil, fmt.Errorf("interleave bound violated")
		}
		out = append(out, ClaimExample{f, q.repository, i, br, ar, ir})
	}
	return out, nil
}
