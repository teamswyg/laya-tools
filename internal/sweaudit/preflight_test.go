package sweaudit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPreflightSamplingAndRootInspection(t *testing.T) {
	sha := strings.Repeat("a", 40)
	input := []Selection{{Source: "multilingual", ID: "a", Repository: "a/b", BaseCommit: sha}, {Source: "full", ID: "z", Repository: "a/b", BaseCommit: sha}, {Source: "full", ID: "a", Repository: "c/d", BaseCommit: sha}}
	samples, e := PreflightSamples(input)
	if e != nil || len(samples) != 2 || samples[0].ID != "z" || input[0].Source != "multilingual" {
		t.Fatal("identity sample or input mutation")
	}
	tree := rootTree{SHA: sha}
	tree.Tree = append(tree.Tree, struct{ Path, Mode, Type, SHA string }{"LICENSE", "100644", "blob", sha}, struct{ Path, Mode, Type, SHA string }{"src", "040000", "tree", sha})
	b, _ := json.Marshal(tree)
	r, e := InspectRoot("a/b", b)
	if e != nil || !r.Available || r.RootEntries != 2 || !reflect.DeepEqual(r.LicenseCandidates, []string{"LICENSE"}) {
		t.Fatalf("root: %+v %v", r, e)
	}
	tree.Truncated = true
	b, _ = json.Marshal(tree)
	if _, e = InspectRoot("a/b", b); e == nil {
		t.Fatal("accepted truncated tree")
	}
	tree.Truncated = false
	tree.Tree[1].Path = "LICENSE"
	b, _ = json.Marshal(tree)
	if _, e = InspectRoot("a/b", b); e == nil {
		t.Fatal("accepted duplicate paths")
	}
	if _, e = InspectRoot("a/b", []byte(`{"message":"Not Found"}`)); e == nil {
		t.Fatal("accepted API error response")
	}
}
