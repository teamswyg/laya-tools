package main

import (
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"math"
	"testing"
)

func TestRejectFixtureAmbiguity(t *testing.T) {
	in := input{Schema: "riido-tiny-features-v1", Origin: "original_synthetic_pdca06_development", Source: tinyhead.Source{Gamma: []float64{1}}}
	for i, s := range []string{"train", "validation", "calibration", "test"} {
		in.Rows = append(in.Rows, row{ID: s, Split: s, Label: i % 3, Feature: []float64{1}, Reference: [3]float64{1, 0, 0}})
	}
	if e := check(in); e != nil {
		t.Fatal(e)
	}
	in.Rows[0].Reference[0] = math.NaN()
	if e := check(in); e == nil {
		t.Fatal("NaN accepted")
	}
	in.Rows[0].Reference = [3]float64{1, 0, 0}
	in.Rows[0].ID = "test"
	if e := check(in); e == nil {
		t.Fatal("duplicate accepted")
	}
}
