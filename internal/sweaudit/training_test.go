package sweaudit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func trainingRow(id, repo, q string, commit byte) Row {
	return Row{id, repo, strings.Repeat(string(commit), 40), q}
}
func TestTrainingExclusionIsTransitiveAndOrderInvariant(t *testing.T) {
	eval := []Row{trainingRow("e", "held/a", "protected", 'a')}
	train := []Row{
		trainingRow("a", "one/a", " PROTECTED ", 'b'), trainingRow("b", "one/a", "other", 'b'),
		trainingRow("c", "held/a", "unique", 'c'), trainingRow("d", "next/a", "unique", 'd'),
		trainingRow("f", "good/a", "good", 'e'), trainingRow("g", "good/a", "Good", 'f'),
		trainingRow("h", "good/b", "another", 'a'), trainingRow("i", "prior/a", "prior", 'b'),
	}
	before := slices.Clone(train)
	r, selected, e := TrainingCandidates(train, eval, []string{"prior/a"})
	if e != nil || r.ExcludedTrainingRows != 5 || r.EligibleTrainingRows != 3 || r.EligibleComponents != 2 || r.DuplicateSiblingRows != 1 || r.DirectSharedRequests != 1 || len(selected) != 2 {
		t.Fatalf("%+v %v", r, e)
	}
	if r.TrainingApproved || r.SplitEstablished || r.PerformanceEstablished || r.ProductionReady {
		t.Fatal("audit fabricated approval")
	}
	if !reflect.DeepEqual(before, train) {
		t.Fatal("input mutated")
	}
	for _, s := range selected {
		if s.ID != "f" && s.ID != "h" {
			t.Fatal("wrong representative")
		}
	}
	slices.Reverse(train)
	rr, ss, e := TrainingCandidates(train, eval, []string{"prior/a"})
	if e != nil || !reflect.DeepEqual(r, rr) || !reflect.DeepEqual(selected, ss) {
		t.Fatal("input order changed membership")
	}
	// A matching ID bridges different repositories and request strings.
	train[0].ID = "e"
	rr, _, e = TrainingCandidates(train, eval, nil)
	if e != nil || rr.DirectSharedIDs != 1 {
		t.Fatal("shared ID not audited")
	}
}
func TestTrainingRejectsInvalidIdentities(t *testing.T) {
	row := trainingRow("a", "repo/name", "question", 'a')
	if _, _, e := TrainingCandidates([]Row{row, row}, []Row{row}, nil); e == nil {
		t.Fatal("duplicate ID")
	}
	bad := row
	bad.BaseCommit = "bad"
	if _, _, e := TrainingCandidates([]Row{bad}, []Row{row}, nil); e == nil {
		t.Fatal("invalid commit")
	}
	if _, _, e := TrainingCandidates([]Row{row}, nil, nil); e == nil {
		t.Fatal("missing evaluation")
	}
}
func TestTrainingReaderBoundsAreOptIn(t *testing.T) {
	b, _ := json.Marshal(trainingRow("a", "repo/name", "question", 'a'))
	b = append(b, '\n')
	data := []byte(strings.Repeat(string(b), 4097))
	p := filepath.Join(t.TempDir(), "rows.jsonl")
	if e := os.WriteFile(p, data, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Read(p, digest(data)); e == nil {
		t.Fatal("evaluation reader enlarged")
	}
	rows, e := ReadTraining(p, digest(data))
	if e != nil || len(rows) != 4097 {
		t.Fatal(e)
	}
	if _, e = ReadTraining(p, strings.Repeat("0", 64)); e == nil {
		t.Fatal("bad hash accepted")
	}
	if _, e = parseBounded(data, 4096); e == nil {
		t.Fatal("row bound ignored")
	}
	if _, e = parseBounded([]byte(`{"instance_id":"a","patch":"hidden"}`), 10); e == nil {
		t.Fatal("label field accepted")
	}
}
