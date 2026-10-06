package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOriginalCorpusAndRejectedFamilyLeak(t *testing.T) {
	b, e := os.ReadFile("../../experiments/state-hints/templates.json")
	if e != nil {
		t.Fatal(e)
	}
	a, e := materialize(b)
	if e != nil {
		t.Fatal(e)
	}
	var in input
	if e = json.Unmarshal(b, &in); e != nil {
		t.Fatal(e)
	}
	in.Templates[1].FamilyID = in.Templates[0].FamilyID
	bad, _ := json.Marshal(in)
	if _, e = materialize(bad); e == nil {
		t.Fatal("accepted duplicated family")
	}
	if _, e = materialize(append(b, []byte(" {}")...)); e == nil {
		t.Fatal("accepted trailing template data")
	}
	if len(a) == 0 {
		t.Fatal("empty corpus")
	}
}
