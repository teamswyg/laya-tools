package typedbehavior

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func TestPreparedPublicCaptionsFitRuntimeBounds(t *testing.T) {
	for _, f := range fixturePrototypes {
		for _, request := range f.requests {
			for _, c := range f.sources {
				if _, e := shortclaim.Validate(shortclaim.Input{Schema: shortclaim.Schema, Request: request, Provenance: "authored-development-test", Candidates: []shortclaim.Candidate{{ID: "a", Text: c.text}}}); e != nil {
					t.Fatalf("authored caption budget: %s/%s: %v", f.prototype, c.id, e)
				}
			}
		}
	}
	if _, e := PrepareDataset(); e != nil {
		t.Fatal(e)
	}
}

func TestTypedProjectionCannotIncludeMetadata(t *testing.T) {
	d, e := PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	before := FeatureInputs(d)
	for i := range d.Parents {
		d.Parents[i].ID = "unrelated-id"
		d.Parents[i].Prototype = "other-family"
		d.Parents[i].ContractID = "unknown-ambiguous"
		for j := range d.Parents[i].Candidates {
			c := &d.Parents[i].Candidates[j]
			c.ID = "other-id"
			c.SourceID = "other-source"
			c.CodeSHA256 = "other-code"
			c.BundleSHA256 = "other-bundle"
		}
	}
	if !reflect.DeepEqual(before, FeatureInputs(d)) {
		t.Fatal("truth/provenance metadata entered the scoring projection")
	}
}

func TestTypedTransportRejectsHelperBundleForgery(t *testing.T) {
	d, e := PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	d.Parents[0].Candidates[0].BundleSHA256 = "forged-bundle"
	raw, e := json.Marshal(d)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = LoadBytes(raw); e == nil {
		t.Fatal("unbound helper closure accepted")
	}
	if _, e = Evaluate(d); e == nil {
		t.Fatal("forged directly constructed dataset accepted")
	}
}

func TestIncompleteCaptionCannotBecomeBooleanTruth(t *testing.T) {
	d, e := PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	d.Parents = slices.Clone(d.Parents[:1])
	d.Parents[0].ContractID = "unknown-incomplete-caption"
	r, e := Evaluate(d)
	if e != nil {
		t.Fatal(e)
	}
	if r.UnknownParents != 1 || r.KnownParents != 0 || len(r.Outcomes[0].Acceptable) != 0 || r.Outcomes[0].Reason != "bounded_caption_does_not_express_complete_contract" {
		t.Fatal("caption uncertainty became a boolean label")
	}
	for _, c := range r.Outcomes[0].Candidates {
		if c.State != "unknown" || c.VectorsChecked != 0 {
			t.Fatal("incomplete request silently used code truth")
		}
	}
	if r.TrainingReady || r.Partitioned || r.FinalEligible {
		t.Fatal("preparation incorrectly authorized downstream work")
	}
}

func TestIncompleteOrChangedCaptionsCannotBePromotedToKnown(t *testing.T) {
	d, e := PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	quoted := slices.IndexFunc(d.Parents, func(p Parent) bool { return p.Prototype == "quoted-delimiters" })
	if quoted < 0 {
		t.Fatal("incomplete authored fixture missing")
	}
	d.Parents[quoted].ContractID = "quoted-delimiters-v2"
	raw, _ := json.Marshal(d)
	if _, e = LoadBytes(raw); e == nil {
		t.Fatal("known-incomplete caption promoted through transport")
	}
	if _, e = Evaluate(d); e == nil {
		t.Fatal("known-incomplete caption promoted through direct API")
	}
	d, _ = PrepareDataset()
	d.Parents[0].Request = "Do something useful."
	if _, e = Evaluate(d); e == nil {
		t.Fatal("unreviewed request received a boolean label")
	}
	d, _ = PrepareDataset()
	d.Parents[0].Candidates[0].Text = "Always correct."
	if _, e = Evaluate(d); e == nil {
		t.Fatal("source identity silently validated changed candidate prose")
	}
}
