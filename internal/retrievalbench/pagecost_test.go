package retrievalbench

import "testing"

func TestPageCostBoundaries(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 20, InterleavedRank: 21}, {BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 61, InterleavedRank: 61}}
	r, err := MeasurePageCost(xs, 61, 20)
	if err != nil {
		t.Fatal(err)
	}
	if r.BaselinePages != 8 || r.HelperPages != 7 || r.BaselineEmitted != 141 || r.HelperEmitted != 121 || r.PageWins != 1 || r.PageLosses != 1 || r.PageTies != 1 || r.WorkWins != 1 || r.WorkLosses != 2 || r.HelperWorkProxy != 14 {
		t.Fatalf("wrong accounting: %+v", r)
	}
	for _, size := range []int{1, 5, 20, 50, 4096} {
		for b := 1; b <= 61; b++ {
			r, err := MeasurePageCost([]ClaimExample{{BaselineRank: b, InterleavedRank: b}}, 61, size)
			if err != nil || r.PageTies != 1 || r.WorkLosses != 1 || r.BaselineEmitted < b || r.BaselineEmitted > 61 {
				t.Fatalf("boundary %d/%d: %+v %v", b, size, r, err)
			}
		}
	}
}

func TestPageCostRejectsInvalid(t *testing.T) {
	for _, rank := range []int{-1, 0, 62} {
		if _, err := MeasurePageCost([]ClaimExample{{BaselineRank: rank, InterleavedRank: 1}}, 61, 20); err == nil {
			t.Fatal("accepted bad baseline")
		}
		if _, err := MeasurePageCost([]ClaimExample{{BaselineRank: 1, InterleavedRank: rank}}, 61, 20); err == nil {
			t.Fatal("accepted bad helper")
		}
	}
	if _, err := MeasurePageCost(nil, 61, 20); err == nil {
		t.Fatal("accepted empty set")
	}
	if _, err := MeasurePageCost([]ClaimExample{{BaselineRank: 1, InterleavedRank: 1}}, 61, 0); err == nil {
		t.Fatal("accepted zero size")
	}
}
