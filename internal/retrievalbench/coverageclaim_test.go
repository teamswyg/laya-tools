package retrievalbench

import (
	"math"
	"reflect"
	"testing"
)

func TestCoverageColumnsOrderAndIsolation(t *testing.T) {
	xs := []ClaimExample{{Repository: "train", BaselineRank: 41, InterleavedRank: 1}, {Repository: "test", BaselineRank: 21, InterleavedRank: 41}}
	xs[0].Features[0] = 1
	extra := make([][8]float64, 2)
	extra[0][0] = .25
	extra[0][4] = .5
	extra[0][7] = .75
	d, e := costClaimDatasetColumns(xs, extra, 8, "train", "development", 1)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(d.Indices, []uint16{0, 12, 16, 19}) || !reflect.DeepEqual(d.Values, []float64{1, .25, .5, .75}) || d.Labels[0] != 1 || d.SampleWeights[0] != 1 {
		t.Fatalf("wrong projection %+v", d)
	}
	w := make([]float64, 20)
	w[0] = 1
	w[12] = 2
	w[16] = 3
	w[19] = 4
	if got := costScoresColumns(xs, extra, 8, w)[0]; got != 6 {
		t.Fatalf("wrong score %v", got)
	}
	xs[1].BaselineRank = 1
	extra[1][4] = 999
	again, e := costClaimDatasetColumns(xs, extra, 8, "train", "development", 1)
	if e != nil || !reflect.DeepEqual(d, again) {
		t.Fatal("heldout mutation altered training")
	}
	extra[0][4] = math.NaN()
	if _, e := costClaimDatasetColumns(xs, extra, 8, "train", "development", 1); e == nil {
		t.Fatal("accepted nan")
	}
	if _, e := costClaimDatasetColumns(xs, nil, 8, "train", "development", 1); e == nil {
		t.Fatal("accepted missing columns")
	}
}
