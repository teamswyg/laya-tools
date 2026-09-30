package retrievalbench

import "testing"

func TestSourceExclusions(t *testing.T) {
	r := Row{Path: "pkg/x.go", Code: "func X() {}", Query: "find item"}
	for _, tt := range []struct {
		file    string
		notices []string
		want    string
	}{
		{"// Apache License\npackage x\nfunc X() {}", nil, ""},
		{"// Code generated. DO NOT EDIT.\npackage x\nfunc X() {}", nil, "generated"},
		{"package x\nfunc X() {}", []string{"pkg/LICENSE"}, "notice"},
		{"// BSD-style license\npackage x\nfunc X() {}", nil, "header"},
		{"package x\nfunc Y() {}", nil, "code"},
	} {
		if got := Exclusion(r, tt.file, tt.notices); got != tt.want {
			t.Fatalf("got %q want %q", got, tt.want)
		}
	}
}
func TestRepeatedQueryHasMultipleKnownTargets(t *testing.T) {
	rows := []Row{{Repository: "a", Name: "Alpha", Code: "func Alpha() {}", Query: " SAME Query "}, {Repository: "b", Name: "Beta", Code: "func Beta() {}", Query: "same query"}, {Repository: "c", Name: "Gamma", Code: "func Gamma() {}", Query: "Gamma"}}
	qs := queries(rows)
	if len(qs) != 2 {
		t.Fatal("counted candidates as questions")
	}
	for _, q := range qs {
		if q.text == "same query" && (len(q.targets) != 2 || q.repository != "multiple") {
			t.Fatal("lost known targets")
		}
	}
	r, e := Evaluate(rows, "raw")
	if e != nil {
		t.Fatal(e)
	}
	if r.All.Queries != 2 || r.NameAbsent.Queries != 1 || r.NamePresent.Queries != 1 {
		t.Fatalf("bad strata %+v", r)
	}
}
func TestMetrics(t *testing.T) {
	m := summarize([]observation{{1, 1}, {2, 2}, {10, 3}, {20, 4}})
	if m.Recall1 != .25 || m.Recall5 != .5 || m.Recall10 != .75 || m.MeanRank != 8.25 || m.P95Rank != 20 {
		t.Fatalf("bad metrics %+v", m)
	}
}
