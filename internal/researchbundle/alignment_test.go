package researchbundle

import (
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/alignment"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"testing"
)

func TestAlignmentModelContract(t *testing.T) {
	a := alignmentArtifact{Variant: "lexical", Encoder: "none", Head: "fp32", Seed: 1729}
	m := alignmentModel{alignment.Schema, paireval.SourceSHA, alignmentPartitionSHA, staticembed.Revision, "lexical", "none", "fp32", 1729, make([]float64, 16)}
	b, _ := json.Marshal(m)
	if e := verifyAlignmentModel(b, a); e != nil {
		t.Fatal(e)
	}
	var object map[string]any
	json.Unmarshal(b, &object)
	object["raw_training_text"] = "must never publish"
	bad, _ := json.Marshal(object)
	if e := verifyAlignmentModel(bad, a); e == nil {
		t.Fatal("accepted unapproved content")
	}
	m.Weights = m.Weights[:15]
	b, _ = json.Marshal(m)
	if e := verifyAlignmentModel(b, a); e == nil {
		t.Fatal("accepted wrong dimension")
	}
	a.Variant = "../../raw"
	if _, e := alignmentName(a); e == nil {
		t.Fatal("accepted traversal")
	}
}
func TestAlignmentTernaryContract(t *testing.T) {
	a := alignmentArtifact{Variant: "lexical", Encoder: "none", Head: "ternary_ste", Seed: 1729}
	m := alignmentModel{alignment.Schema, paireval.SourceSHA, alignmentPartitionSHA, staticembed.Revision, "lexical", "none", "ternary_ste", 1729, make([]float64, 16)}
	m.Weights[0] = .5
	m.Weights[1] = -.5
	b, _ := json.Marshal(m)
	if e := verifyAlignmentModel(b, a); e != nil {
		t.Fatal(e)
	}
	m.Weights[1] = .25
	b, _ = json.Marshal(m)
	if e := verifyAlignmentModel(b, a); e == nil {
		t.Fatal("accepted extra ternary magnitude")
	}
}
func TestAlignmentCannotPromote(t *testing.T) {
	m := Manifest{Schema: "riido-alignment-bundle-v1", ProductionReady: true}
	if _, e := VerifyAlignment(t.TempDir(), m); e == nil {
		t.Fatal("accepted production claim")
	}
}
