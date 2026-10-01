package typedbehavior

import (
	"slices"
	"testing"
)

func TestRelationsTransitivelyIncludeNegativeSourcesAndSemanticClosures(t *testing.T) {
	shared := ComponentPin{ID: "shared-helper", Kind: "function", SHA256: "component"}
	parents := []RelationParent{
		{ID: "a", Prototype: "p1", Core: "c1", Request: "first requirement", Sources: []RelationSource{{ID: "negative"}}},
		{ID: "b", Prototype: "p2", Core: "c2", Request: "second requirement", Sources: []RelationSource{{ID: "negative"}, {ID: "other", Components: []ComponentPin{shared}}}},
		{ID: "c", Prototype: "p3", Core: "c3", Request: "third requirement", Sources: []RelationSource{{ID: "distinct", Components: []ComponentPin{shared}}}},
	}
	r, e := auditRelations(parents)
	if e != nil {
		t.Fatal(e)
	}
	if r.ConnectedGroups != 1 || len(r.Groups[0].Parents) != 3 {
		t.Fatal("transitive source/closure dependence lost")
	}
	if r.TrainingReady || r.MeetsGroupMinimum {
		t.Fatal("relation evidence incorrectly qualified training")
	}
}

func TestRelationsIncludeRenamedBehavioralHelperCopies(t *testing.T) {
	parents := []RelationParent{
		{ID: "a", Prototype: "p1", Core: "c1", Request: "first requirement", Sources: []RelationSource{{ID: "a-src", Components: []ComponentPin{{ID: "helper-a", Kind: "function", SHA256: "raw-a", NormalizedBehaviorSHA256: "copy"}}}}},
		{ID: "b", Prototype: "p2", Core: "c2", Request: "second requirement", Sources: []RelationSource{{ID: "b-src", Components: []ComponentPin{{ID: "helper-b", Kind: "function", SHA256: "raw-b", NormalizedBehaviorSHA256: "copy"}}}}},
	}
	r, e := auditRelations(parents)
	if e != nil || r.ConnectedGroups != 1 {
		t.Fatal("renamed helper copy unconnected", e)
	}
	if !slices.Contains(r.Edges[0].Reasons, "normalized_behavioral_component_copy") {
		t.Fatal("copy reason missing")
	}
}

func TestRelationReportsDoNotClaimIndependenceFromGroupCount(t *testing.T) {
	var parents []RelationParent
	for i := 0; i < 15; i++ {
		s := string(rune('a' + i))
		parents = append(parents, RelationParent{ID: s, Prototype: s, Core: s, Request: "unique " + s})
	}
	r, e := auditRelations(parents)
	if e != nil || !r.MeetsGroupMinimum || r.TrainingReady {
		t.Fatal("operational minimum incorrectly became training readiness", e)
	}
}

func TestUnknownOnlyGroupCannotSatisfyLabeledMinimum(t *testing.T) {
	var parents []RelationParent
	var labeled []string
	for i := 0; i < 15; i++ {
		id := string(rune('a' + i))
		parents = append(parents, RelationParent{ID: id, Prototype: id, Core: id, Request: "unique " + id})
		if i < 14 {
			labeled = append(labeled, id)
		}
	}
	r, e := auditRelations(parents)
	if e != nil {
		t.Fatal(e)
	}
	applyLabelEvidence(&r, labeled)
	if !r.MeetsGroupMinimum || r.LabeledConnectedGroups != 14 || r.MeetsLabeledGroupMinimum || r.TrainingReady {
		t.Fatal("unknown-only group supplied fictitious learning evidence")
	}
}
