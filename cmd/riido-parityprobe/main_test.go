package main

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	d, flip, err := distance([]float64{.5, .5}, []float64{.4, .6})
	if err != nil || math.Abs(d-.1) > 1e-12 || !flip {
		t.Fatalf("unexpected comparison: %v %v %v", d, flip, err)
	}
	for _, p := range [][]float64{{.2}, {math.NaN(), .5}, {math.Inf(1), 0}, {-.1, 1.1}, {.1, .2}} {
		if _, _, err := distance(p, []float64{.5, .5}); err == nil {
			t.Fatal("accepted malformed probability")
		}
	}
}
