package retrievalbench

import "fmt"

type PageCost struct {
	PageSize                           int
	Questions                          int
	BaselinePages, HelperPages         int
	BaselineEmitted, HelperEmitted     int
	PageWins, PageLosses, PageTies     int
	BaselineWorkProxy, HelperWorkProxy int
	WorkWins, WorkLosses, WorkTies     int
}

type PageCostProbe struct {
	Questions, Candidates int
	PrimaryPageSize       int
	Costs                 []PageCost
	ProductionReady       bool
}

// MeasurePageCost assumes perfect recognition of the first known target.
// Work counts index/search passes with equal cost, not time or billed tokens.
func MeasurePageCost(xs []ClaimExample, candidates, size int) (PageCost, error) {
	r := PageCost{PageSize: size, Questions: len(xs)}
	if candidates < 1 || candidates > 4096 || size < 1 || size > 4096 || len(xs) == 0 {
		return r, fmt.Errorf("invalid page-cost input")
	}
	for _, x := range xs {
		if x.BaselineRank < 1 || x.BaselineRank > candidates || x.InterleavedRank < 1 || x.InterleavedRank > candidates {
			return PageCost{}, fmt.Errorf("target rank outside catalog")
		}
		b, h := (x.BaselineRank+size-1)/size, (x.InterleavedRank+size-1)/size
		r.BaselinePages += b
		r.HelperPages += h
		r.BaselineEmitted += min(candidates, b*size)
		r.HelperEmitted += min(candidates, h*size)
		if h < b {
			r.PageWins++
		} else if h > b {
			r.PageLosses++
		} else {
			r.PageTies++
		}
		if 2*h < b {
			r.WorkWins++
		} else if 2*h > b {
			r.WorkLosses++
		} else {
			r.WorkTies++
		}
	}
	r.BaselineWorkProxy, r.HelperWorkProxy = r.BaselinePages, 2*r.HelperPages
	return r, nil
}

func ProbePageCost(rows []Row) (PageCostProbe, error) {
	xs, err := PrepareBaselineFirstClaims(rows)
	if err != nil {
		return PageCostProbe{}, err
	}
	r := PageCostProbe{Questions: len(xs), Candidates: len(rows), PrimaryPageSize: 20}
	if len(xs) != 2948 || len(rows) != 3009 {
		return r, fmt.Errorf("unexpected corpus size")
	}
	for _, size := range []int{1, 5, 10, 20, 50} {
		c, err := MeasurePageCost(xs, len(rows), size)
		if err != nil {
			return r, err
		}
		r.Costs = append(r.Costs, c)
	}
	return r, nil
}
