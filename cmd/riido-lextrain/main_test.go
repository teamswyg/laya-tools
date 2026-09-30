package main

import (
	"github.com/teamswyg/laya-tools/internal/paireval"
	"reflect"
	"testing"
)

func TestTrainingBoundary(t *testing.T) {
	one := 1
	rows := []paireval.Row{{Query: "parse data", Code: "def parse_data(x): return x", Label: &one}, {Query: "unseen", Code: "hidden", Label: &one}}
	s := paireval.Split{Rows: []paireval.Membership{{Group: 1, Split: "development"}, {Group: 2, Split: "reserve1"}}}
	a, e := prepare(rows, s, "development", true)
	if e != nil {
		t.Fatal(e)
	}
	rows[1] = paireval.Row{}
	b, e := prepare(rows, s, "development", true)
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("holdout dependency", e)
	}
	if _, e = prepare(rows, s, "reserve1", true); e == nil {
		t.Fatal("holdout accepted")
	}
}
