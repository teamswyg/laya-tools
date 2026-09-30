package main

import (
	"github.com/teamswyg/laya-tools/internal/fileeval"
	"testing"
)

func TestPageAccountingKeepsInapplicableRowsSeparate(t *testing.T) {
	r := newAggregate("")
	r.Rows = 3
	for _, pair := range [][2]int{{21, 2}, {21, 41}, {0, 0}} {
		var scores [3]fileeval.Score
		scores[0].FirstRank = pair[0]
		scores[2].FirstRank = pair[1]
		add(&r, scores)
	}
	if r.ApplicableCostRows != 2 || r.ImprovedPages != 1 || r.WorsenedPages != 1 || r.TiedPages != 0 || r.Policies[0].FirstTargetPages != 4 || r.Policies[2].FirstTargetPages != 4 {
		t.Fatalf("incorrect page accounting: %+v", r)
	}
	r.Policies[2].Hit10 = 1
	rates(&r)
	if r.Policies[2].Hit10Rate != 1.0/3 {
		t.Fatal("all-task denominator changed")
	}
}
