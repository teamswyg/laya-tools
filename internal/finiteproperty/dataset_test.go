package finiteproperty

import (
	"reflect"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/scopedproperty"
	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// These preparation controls never call Audit or any candidate function.
func TestFiniteScopeAndOriginalCandidateOrder(t *testing.T) {
	old, err := scopedproperty.CreateDataset()
	if err != nil {
		t.Fatal(err)
	}
	d, err := Prepare()
	if err != nil {
		t.Fatal(err)
	}
	for i, p := range d.Parents {
		if p.OriginalParentID != old.Parents[i].OriginalParentID {
			t.Fatal("parent order changed")
		}
		for j, c := range p.Candidates {
			previous := old.Parents[i].Candidates[j]
			if c.ID != previous.ID || c.SourceID != previous.SourceID || c.CodeSHA256 != previous.CodeSHA256 || c.BundleSHA256 != previous.BundleSHA256 {
				t.Fatal("candidate identity changed")
			}
		}
	}
	for _, p := range old.Parents {
		if p.CaptionReview != scopedproperty.PendingReview {
			t.Fatal("promoted old pending caption")
		}
	}
	if d.Oracle != scopedproperty.Definitions() || d.NewIndependentParents != 0 {
		t.Fatal("oracle/independence changed")
	}
	if err := Validate(d); err != nil {
		t.Fatal(err)
	}
	d.Parents[0].Request += " Extra scope."
	if err := Validate(d); err == nil {
		t.Fatal("unfrozen prose accepted")
	}
}

func TestScopeBindingRejectsHexTypoOrMissingInput(t *testing.T) {
	d, err := Prepare()
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Dataset){
		func(x *Dataset) {
			x.Parents[0].Request = strings.Replace(x.Parents[0].Request, "0x22612c62222c63", "0x22612c62222c64", 1)
		},
		func(x *Dataset) {
			x.Parents[0].Request = strings.Replace(x.Parents[0].Request, "0x22612c62222c63", "other", 1)
		},
		func(x *Dataset) {
			x.Parents[0].Request = strings.Replace(x.Parents[0].Request, "0x22612c62222c63", "0xZZ", 1)
		},
	} {
		changed := d
		change(&changed)
		if err := validateScopeText(changed); err == nil {
			t.Fatal("scope mismatch accepted")
		}
	}
}

func TestProspectiveTextEnvelopeHasNoMetadataProjection(t *testing.T) {
	d, err := Prepare()
	if err != nil {
		t.Fatal(err)
	}
	p := d.Parents[0]
	want := FeatureInput(p)
	p.ID, p.OriginalParentID, p.PropertyID = "metadata-change", "metadata-change", "metadata-change"
	for i := range p.Candidates {
		p.Candidates[i].ID, p.Candidates[i].SourceID, p.Candidates[i].CodeSHA256, p.Candidates[i].BundleSHA256 = "changed", "changed", "changed", "changed"
	}
	if !reflect.DeepEqual(FeatureInput(p), want) {
		t.Fatal("metadata leaked into text envelope")
	}
	for _, p := range d.Parents {
		prepared, err := shortclaim.Validate(FeatureInput(p))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(prepared.NormalizedRequest, "0x") != d.Oracle[propertyIndex(p.PropertyID)].LiteralCount {
			t.Fatal("hex scope lost after normalization")
		}
	}
}

func propertyIndex(id string) int {
	for i, expected := range [3]string{"finite-input-56e-v2-1", "finite-input-56e-v2-2", "finite-input-56e-v2-3"} {
		if id == expected {
			return i
		}
	}
	panic("unexpected test property")
}

func TestExactMatchRequiresErrorAndAllOutputSlots(t *testing.T) {
	want := Observation{Error: typedbehavior.PropertyQuotedSyntax}
	for _, wrong := range []Observation{
		{}, {Error: typedbehavior.PropertyQuotedSyntax, Count: 1},
		{Error: typedbehavior.PropertyQuotedSyntax, Tokens: [8]string{7: "residue"}},
		{Error: typedbehavior.PropertyQuotedSyntax, Panicked: true},
		{Error: typedbehavior.PropertyQuotedBounds},
	} {
		if ExactMatch(wrong, want) {
			t.Fatal("vacuous/partial/panic match")
		}
	}
	if !ExactMatch(want, want) {
		t.Fatal("exact observation rejected")
	}
	if ExactMatch(Observation{Count: 1}, Observation{Count: 2}) {
		t.Fatal("successful Count mismatch")
	}
}
