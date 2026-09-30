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
		additional, e = aux.RankInto(lexicalhint.NormalizeText(q.text), additional)
		if e != nil {
			return nil, e
		}
		order, e := hintsearch.Interleave(baseline.Order, additional.Order)
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
		if ir > min(len(rows), 2*br) {
			return nil, fmt.Errorf("interleave bound violated")
		}
		out = append(out, ClaimExample{f, q.repository, i, br, ar, ir})
	}
	return out, nil
}
