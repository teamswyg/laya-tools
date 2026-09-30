package sweaudit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func fixture(id, request string, snapshot int) Row {
	return Row{id, "public/repo", fmt.Sprintf("%040x", snapshot), request}
}
func TestTransitiveEvaluationGroups(t *testing.T) {
	full := []Row{fixture("a", " SAME request ", 1), fixture("b", "same\trequest", 2), fixture("c", "third", 2), fixture("d", "separate", 4)}
	multi := []Row{fixture("c", "fourth", 3)}
	before := slices.Clone(full)
	r, selection, e := groupEvaluation(full, multi, 2)
	if e != nil {
		t.Fatal(e)
	}
	if r.Components != 2 || r.NonSingletonComponents != 1 || r.LargestComponent != 4 || r.SelectedComponents != 2 || !r.EvaluationOnly || r.ProductionReady {
		t.Fatalf("wrong grouping: %+v", r)
	}
	if !reflect.DeepEqual(before, full) {
		t.Fatal("mutated input")
	}
	for _, s := range selection {
		if s.ID != "a" && s.ID != "d" {
			t.Fatal("representative not lexical minimum")
		}
	}
	slices.Reverse(full)
	rr, ss, e := groupEvaluation(full, multi, 2)
	if e != nil || !reflect.DeepEqual(r, rr) || !reflect.DeepEqual(selection, ss) {
		t.Fatal("order-dependent grouping")
	}
	short, none, e := GroupEvaluation(full, multi)
	if e != nil || short.Shortfall != 2398 || len(none) != 0 || short.SelectedComponents != 0 {
		t.Fatal("weakened 2400 threshold")
	}
	b, _ := json.Marshal(r)
	if strings.Contains(string(b), "SAME request") || strings.Contains(string(b), "fourth") {
		t.Fatal("request text exported")
	}
	if _, _, e := groupEvaluation([]Row{fixture("a", "x", 1), fixture("a", "y", 2)}, nil, 1); e == nil {
		t.Fatal("duplicate source identity accepted")
	}
}
func TestFrozen2400SelectionAndInputBudget(t *testing.T) {
	rows := make([]Row, 2401)
	for i := range rows {
		rows[i] = fixture(fmt.Sprint(i), fmt.Sprint("request ", i), i+1)
	}
	// Every row is over the existing runtime query budget; none may be filtered.
	for i := range rows {
		rows[i].Request += strings.Repeat("x", 8193)
	}
	r, selection, e := GroupEvaluation(rows, nil)
	if e != nil {
		t.Fatal(e)
	}
	if r.Components != 2401 || len(selection) != 2400 || r.RequestsOver8192 != 2400 || r.SelectedCounts.DistinctRequests != 2400 || r.Shortfall != 0 {
		t.Fatalf("selection: %+v", r)
	}
	slices.Reverse(rows)
	rr, ss, e := GroupEvaluation(rows, nil)
	if e != nil || r.SelectedSHA256 != rr.SelectedSHA256 || !reflect.DeepEqual(selection, ss) {
		t.Fatal("unstable selection")
	}
	rows[0].BaseCommit = "invalid"
	if _, _, e := GroupEvaluation(rows, nil); e == nil {
		t.Fatal("invalid grouping key accepted")
	}
}
