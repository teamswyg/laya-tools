package sweaudit

import (
	"reflect"
	"slices"
	"testing"
)

func partitionFixture() ([]Row, []Row, []RepositoryIdentity) {
	train := []Row{
		trainingRow("a", "old/a", "alpha", 'a'), trainingRow("b", "copy/a", "fork query", 'b'),
		trainingRow("c", "b/b", "bridge", 'c'), trainingRow("d", "c/c", " BRIDGE ", 'd'),
		trainingRow("e", "c/c", "separate c", 'e'), trainingRow("f", "d/d", "independent d", 'f'),
		trainingRow("g", "renamed/eval", "different query", 'a'),
	}
	eval := []Row{trainingRow("eval", "eval/x", "protected", 'b')}
	ids := []RepositoryIdentity{
		{Requested: "old/a", Canonical: "new/a", ID: 1, NetworkID: 1},
		{Requested: "copy/a", Canonical: "copy/a", ID: 2, NetworkID: 1, Fork: true},
		{Requested: "b/b", Canonical: "b/b", ID: 3, NetworkID: 3},
		{Requested: "c/c", Canonical: "c/c", ID: 4, NetworkID: 4},
		{Requested: "d/d", Canonical: "d/d", ID: 5, NetworkID: 5},
		{Requested: "renamed/eval", Canonical: "held/x", ID: 99, NetworkID: 99},
		{Requested: "eval/x", Canonical: "held/x", ID: 99, NetworkID: 99},
	}
	return train, eval, ids
}
func TestPartitionProtectsAliasesForksAndSiblingBridges(t *testing.T) {
	train, eval, ids := partitionFixture()
	before := slices.Clone(train)
	r, tasks, e := partitionTraining(train, eval, nil, ids, 1, 1, 1)
	if e != nil || !r.Complete || r.Candidates != 5 || r.Units != 3 || len(tasks) != 5 {
		t.Fatalf("%+v %v", r, e)
	}
	unitOf := func(repo string) string {
		for _, u := range r.RepositoryUnits {
			if slices.Contains(u.Repositories, repo) {
				return u.SHA256
			}
		}
		return ""
	}
	if unitOf("new/a") != unitOf("copy/a") || unitOf("b/b") != unitOf("c/c") {
		t.Fatal("fork or sibling bridge split")
	}
	for _, x := range tasks {
		if x.Task.ID == "g" || x.Task.ID == "d" {
			t.Fatal("protected alias or duplicate representative survived")
		}
	}
	if !reflect.DeepEqual(train, before) {
		t.Fatal("caller rows mutated")
	}
	slices.Reverse(train)
	slices.Reverse(eval)
	slices.Reverse(ids)
	rr, tt, e := partitionTraining(train, eval, nil, ids, 1, 1, 1)
	if e != nil || !reflect.DeepEqual(r, rr) || !reflect.DeepEqual(tasks, tt) {
		t.Fatal("input order changed split")
	}
	if r.TrainingApproved || r.SemanticIndependenceEstablished || r.ProductionReady {
		t.Fatal("identity audit approved quality/license")
	}
}
func TestPartitionRejectsIncompleteAndInconsistentIdentities(t *testing.T) {
	train, eval, ids := partitionFixture()
	if _, _, e := partitionTraining(train, eval, nil, ids[:len(ids)-1], 1, 1, 1); e == nil {
		t.Fatal("missing identity")
	}
	bad := slices.Clone(ids)
	bad[6].Canonical = "other/x"
	if _, _, e := partitionTraining(train, eval, nil, bad, 1, 1, 1); e == nil {
		t.Fatal("one ID with inconsistent names")
	}
	bad = slices.Clone(ids)
	bad[6].NetworkID = 100
	if _, _, e := partitionTraining(train, eval, nil, bad, 1, 1, 1); e == nil {
		t.Fatal("inconsistent fork ancestry")
	}
	r, tasks, e := partitionTraining(train, eval, nil, ids, 10, 10, 10)
	if e != nil || r.Complete || len(tasks) != 0 || r.MembershipSHA256 != "" {
		t.Fatal("shortfall fabricated membership")
	}
}

func TestRoleVerifierRejectsCrossRoleLeaksAndDuplicates(t *testing.T) {
	train, eval, ids := partitionFixture()
	r, tasks, e := partitionTraining(train, eval, nil, ids, 1, 1, 1)
	if e != nil {
		t.Fatal(e)
	}
	ct := slices.Clone(train)
	for i := range ct {
		for _, x := range ids {
			if x.Requested == ct[i].Repository {
				ct[i].Repository = x.Canonical
			}
		}
	}
	if e = verifyRoles(ct, ids, r, tasks); e != nil {
		t.Fatal(e)
	}
	bad := slices.Clone(tasks)
	bad[0] = bad[1]
	if e = verifyRoles(ct, ids, r, bad); e == nil {
		t.Fatal("duplicate membership accepted")
	}
	bad = slices.Clone(tasks)
	bad[0].Role = "unknown"
	if e = verifyRoles(ct, ids, r, bad); e == nil {
		t.Fatal("wrong role accepted")
	}
	// Inject a shared request across repositories assigned to different roles.
	var a, b int
	found := false
	for i := range ct {
		for j := range ct {
			if ct[i].ID == tasks[0].Task.ID && ct[j].ID == tasks[len(tasks)-1].Task.ID {
				a, b = i, j
				found = true
			}
		}
	}
	if !found {
		t.Fatal("fixture identities missing")
	}
	ct[b].Request = ct[a].Request
	if e = verifyRoles(ct, ids, r, tasks); e == nil {
		t.Fatal("cross-role request accepted")
	}
}
