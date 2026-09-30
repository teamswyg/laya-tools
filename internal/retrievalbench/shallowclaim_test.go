package retrievalbench

import (
	"github.com/teamswyg/laya-tools/internal/searchclaim"
	"reflect"
	"testing"
)

func TestTreeRowsDoNotReadOtherRepository(t *testing.T) {
	xs := []ClaimExample{{Repository: "train", BaselineRank: 41, InterleavedRank: 2}, {Repository: "test", BaselineRank: 50, InterleavedRank: 99}}
	extra := make([][searchclaim.SpreadDimension]float64, 2)
	extra[0][0] = .5
	x, y := treeRows(xs, extra, "train")
	if len(x) != 1 || x[0][12] != .5 || y[0] != 2 {
		t.Fatal("wrong projection")
	}
	xs[1].BaselineRank = 1
	xs[1].Features[0] = 999
	extra[1][0] = 999
	xx, yy := treeRows(xs, extra, "train")
	if !reflect.DeepEqual(x, xx) || !reflect.DeepEqual(y, yy) {
		t.Fatal("held-out mutation changed training")
	}
}
func TestClaimEffectAccounting(t *testing.T) {
	xs := []ClaimExample{{BaselineRank: 41, InterleavedRank: 1}, {BaselineRank: 21, InterleavedRank: 1}, {BaselineRank: 21, InterleavedRank: 41}, {BaselineRank: 21, InterleavedRank: 41}, {BaselineRank: 1, InterleavedRank: 1}, {BaselineRank: 1, InterleavedRank: 1}}
	sel := applyPageBudget(xs, []float64{1, 0, 1, 0, 1, 0}, BudgetThreshold{MinimumScore: 1})
	e := effect(xs, sel)
	if e != (ClaimEffect{1, 1, 1, 1, 1, 1, 2, 1, 1, 1}) {
		t.Fatalf("wrong effect %+v", e)
	}
}
func TestTreeSelectionUsesValidationCostAndSimplicity(t *testing.T) {
	b := ShallowCandidate{Depth: 4, MinLeaf: 16, Threshold: BudgetThreshold{ValidationPages: 100, ValidationCalls: 90}}
	a := b
	a.Threshold.ValidationPages = 99
	if !betterTree(a, b) {
		t.Fatal("pages priority")
	}
	a = b
	a.Threshold.ValidationCalls = 89
	if !betterTree(a, b) {
		t.Fatal("calls priority")
	}
	a = b
	a.Depth = 3
	if !betterTree(a, b) {
		t.Fatal("depth priority")
	}
	a = b
	a.MinLeaf = 64
	if !betterTree(a, b) {
		t.Fatal("leaf priority")
	}
	a.Threshold.ValidationPages = 101
	if betterTree(a, b) {
		t.Fatal("simplicity overrode cost")
	}
}
