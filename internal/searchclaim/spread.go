package searchclaim

import (
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/hintsearch"
	"math"
)

const SpreadDimension = 4

var SpreadNames = [SpreadDimension]string{"top20_relative_mean", "top20_relative_stddev", "top20_first_score_share", "top20_last_first_ratio"}

// ScoreSpread reads only the top 20 baseline scores. It requires descending
// index-produced ranks, uses no outcomes or auxiliary ranking, and allocates no
// storage on valid input. Scores outside the inspected prefix are not read.
func ScoreSpread(b hintsearch.Ranking) ([SpreadDimension]float64, error) {
	var f [SpreadDimension]float64
	if len(b.Order) == 0 || len(b.Order) != len(b.Scores) || len(b.Order) > hintsearch.MaxDocuments {
		return f, fmt.Errorf("invalid ranking shape")
	}
	n := min(20, len(b.Order))
	var values [20]float64
	top, prev, sum := 0., math.Inf(1), 0.
	for i, id := range b.Order[:n] {
		if id < 0 || id >= len(b.Scores) {
			return f, fmt.Errorf("invalid rank index")
		}
		v := b.Scores[id]
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > prev {
			return f, fmt.Errorf("invalid descending score")
		}
		prev = v
		if i == 0 {
			top = v
		}
		if top > 0 {
			values[i] = v / top
		}
		sum += values[i]
	}
	if top == 0 {
		return f, nil
	}
	mean := sum / float64(n)
	variance := 0.
	for _, v := range values[:n] {
		d := v - mean
		variance += d * d
	}
	f = [SpreadDimension]float64{mean, math.Sqrt(variance / float64(n)), 1 / sum, values[n-1]}
	return f, nil
}
