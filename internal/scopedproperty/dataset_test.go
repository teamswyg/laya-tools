package scopedproperty

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/typedbehavior"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func cloneDataset(d Dataset) Dataset {
	out := d
	out.Parents = slices.Clone(d.Parents)
	for i := range out.Parents {
		out.Parents[i].Candidates = slices.Clone(d.Parents[i].Candidates)
	}
	return out
}

func cloneOriginalDataset(d typedbehavior.Dataset) typedbehavior.Dataset {
	out := d
	out.Parents = slices.Clone(d.Parents)
	for i := range out.Parents {
		out.Parents[i].Candidates = slices.Clone(d.Parents[i].Candidates)
	}
	return out
}

func TestOriginalDatasetGuardMatchesEntireFrozenFile(t *testing.T) {
	original, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	raw, e := json.MarshalIndent(original, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	raw = append(raw, '\n')
	frozen, e := os.ReadFile("../../experiments/short-claim/probes-56b.json")
	if e != nil {
		t.Fatal(e)
	}
	const frozenSHA = "0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e"
	if OriginalDatasetSHA256 != frozenSHA || sha(raw) != frozenSHA || sha(frozen) != frozenSHA || !bytes.Equal(raw, frozen) || verifyOriginalDataset(original) != nil {
		t.Fatal("factory original bytes differ from the whole frozen JSON")
	}
}

func TestOriginalDatasetGuardRejectsRequestOrderAndPinsWithoutMutation(t *testing.T) {
	original, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	n := slices.IndexFunc(original.Parents, func(p typedbehavior.Parent) bool { return p.ID == OriginalParentIDs()[0] })
	if n < 0 || len(original.Parents[n].Candidates) != 3 {
		t.Fatal("original quoted parent is absent")
	}
	before, e := json.Marshal(original)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name string
		edit func(*typedbehavior.Dataset)
	}{
		{"request", func(d *typedbehavior.Dataset) { d.Parents[n].Request += " public-marker" }},
		{"candidate order", func(d *typedbehavior.Dataset) {
			d.Parents[n].Candidates[0], d.Parents[n].Candidates[1] = d.Parents[n].Candidates[1], d.Parents[n].Candidates[0]
		}},
		{"code pin", func(d *typedbehavior.Dataset) { d.Parents[n].Candidates[0].CodeSHA256 = strings.Repeat("0", 64) }},
		{"bundle pin", func(d *typedbehavior.Dataset) { d.Parents[n].Candidates[0].BundleSHA256 = strings.Repeat("0", 64) }},
		{"unselected parent", func(d *typedbehavior.Dataset) { d.Parents[0].Request += " public-marker" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := cloneOriginalDataset(original)
			tc.edit(&changed)
			if e := verifyOriginalDataset(changed); e == nil || e.Error() != "scoped_original_dataset_hash_mismatch" {
				t.Fatal("original drift did not produce the fixed hash rejection")
			}
		})
	}
	after, e := json.Marshal(original)
	if e != nil || !bytes.Equal(before, after) || verifyOriginalDataset(original) != nil {
		t.Fatal("guard controls mutated their original input")
	}
	fresh, e := typedbehavior.PrepareDataset()
	if e != nil || verifyOriginalDataset(fresh) != nil {
		t.Fatal("guard controls changed original fixture storage")
	}
}

func TestEveryDraftUsesActualRawNormalizedBudget(t *testing.T) {
	for _, draft := range captionDrafts() {
		input := shortclaim.Input{Schema: shortclaim.Schema, Request: draft.request, Provenance: "draft-budget-check"}
		for i, text := range draft.texts {
			input.Candidates = []shortclaim.Candidate{{ID: "candidate", Text: text}}
			if _, e := shortclaim.Validate(input); e != nil {
				t.Fatalf("draft %s candidate %d fails actual budget: %v", draft.property, i, e)
			}
		}
	}
}

// Preparation checks only metadata, closures and text bounds. It does not call
// ObserveQuotedProperty, evaluate the twelve proposals or build any outcome set.
func TestCreatePendingTwelvePreservesOriginalCandidatesAndPins(t *testing.T) {
	original, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	d, e := CreateDataset()
	if e != nil {
		t.Fatal(e)
	}
	if d.Schema != Schema || d.Origin != Origin || d.CaptionReview != "pending" || len(d.Parents) != 12 || ValidateDataset(d) != nil {
		t.Fatal("closed pending preparation or dataset binding failed")
	}
	for _, p := range d.Parents {
		n := slices.IndexFunc(original.Parents, func(old typedbehavior.Parent) bool { return old.ID == p.OriginalParentID })
		if n < 0 || original.Parents[n].Prototype != "quoted-delimiters" || p.CaptionReview != "pending" || len(p.Candidates) != 3 {
			t.Fatal("proposal is not bound to an original unknown quoted parent")
		}
		for j, c := range p.Candidates {
			old := original.Parents[n].Candidates[j]
			if c.ID != old.ID || c.SourceID != old.SourceID || c.CodeSHA256 != old.CodeSHA256 || c.BundleSHA256 != old.BundleSHA256 || c.CaptionReview != "pending" {
				t.Fatal("candidate set, original order or actual source closure changed")
			}
		}
	}
	raw, e := json.Marshal(d)
	if e != nil {
		t.Fatal(e)
	}
	for _, forbidden := range []string{"outcomes", "acceptable", "known_parents", "no_answer", "groups", "roles", "fits", "expected_control", "ready"} {
		if strings.Contains(string(raw), `"`+forbidden+`"`) {
			t.Fatal("preparation published labels, group results or review approval")
		}
	}
}

func TestMetadataNeverEntersFeatureProjection(t *testing.T) {
	d, e := CreateDataset()
	if e != nil {
		t.Fatal(e)
	}
	before := FeatureInputs(d)
	poisoned := cloneDataset(d)
	poisoned.Schema, poisoned.Origin, poisoned.CaptionReview = "changed", "changed", "ready"
	for i := range poisoned.Parents {
		p := &poisoned.Parents[i]
		p.ID, p.OriginalParentID, p.PropertyID, p.TruthTableSHA256 = "changed", "changed", "changed", "changed"
		p.PropertyVersion, p.CaptionReview = 99, "ready"
		for j := range p.Candidates {
			c := &p.Candidates[j]
			c.ID, c.SourceID, c.CodeSHA256, c.BundleSHA256, c.CaptionReview = "changed", "changed", "changed", "changed", "ready"
		}
	}
	if !reflect.DeepEqual(before, FeatureInputs(poisoned)) {
		t.Fatal("source, target or review metadata changed ranking inputs")
	}
	for _, input := range before {
		if _, e := shortclaim.Validate(input); e != nil {
			t.Fatal("feature projection violates runtime input bounds")
		}
	}
	before[0].Candidates[0].Text = "changed"
	before[0].Request = "changed"
	if FeatureInputs(d)[0].Candidates[0].Text == "changed" || d.Parents[0].Request == "changed" {
		t.Fatal("projected data exposes mutable candidate storage")
	}
	changedText := cloneDataset(d)
	changedText.Parents[0].Candidates[0].Text += " changed"
	if reflect.DeepEqual(FeatureInputs(d), FeatureInputs(changedText)) {
		t.Fatal("actual candidate text was accidentally omitted from features")
	}
}

func TestClosedValidationRejectsForgedPinsOrderCaptionsAndReview(t *testing.T) {
	d, e := CreateDataset()
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name string
		edit func(*Dataset)
	}{
		{"schema", func(d *Dataset) { d.Schema = "different" }},
		{"origin", func(d *Dataset) { d.Origin = "different" }},
		{"ready", func(d *Dataset) { d.CaptionReview = "ready" }},
		{"missing parent", func(d *Dataset) { d.Parents = d.Parents[:11] }},
		{"parent order", func(d *Dataset) { d.Parents[0], d.Parents[1] = d.Parents[1], d.Parents[0] }},
		{"original parent", func(d *Dataset) { d.Parents[0].OriginalParentID = "typed56b-p01-1" }},
		{"property", func(d *Dataset) { d.Parents[0].PropertyID = SyntaxZeroProperty }},
		{"version", func(d *Dataset) { d.Parents[0].PropertyVersion++ }},
		{"table pin", func(d *Dataset) { d.Parents[0].TruthTableSHA256 = strings.Repeat("0", 64) }},
		{"parent ready", func(d *Dataset) { d.Parents[0].CaptionReview = "ready" }},
		{"request", func(d *Dataset) { d.Parents[0].Request += " changed" }},
		{"raw bound", func(d *Dataset) { d.Parents[0].Request = strings.Repeat("x", 513) }},
		{"word bound", func(d *Dataset) { d.Parents[0].Request = strings.Repeat("word ", 33) }},
		{"candidate order", func(d *Dataset) {
			d.Parents[0].Candidates[0], d.Parents[0].Candidates[1] = d.Parents[0].Candidates[1], d.Parents[0].Candidates[0]
		}},
		{"candidate count", func(d *Dataset) { d.Parents[0].Candidates = d.Parents[0].Candidates[:2] }},
		{"candidate ID", func(d *Dataset) { d.Parents[0].Candidates[0].ID = "new" }},
		{"source ID", func(d *Dataset) { d.Parents[0].Candidates[0].SourceID = "unregistered" }},
		{"code pin", func(d *Dataset) { d.Parents[0].Candidates[0].CodeSHA256 = strings.Repeat("0", 64) }},
		{"closure pin", func(d *Dataset) { d.Parents[0].Candidates[0].BundleSHA256 = strings.Repeat("0", 64) }},
		{"candidate ready", func(d *Dataset) { d.Parents[0].Candidates[0].CaptionReview = "ready" }},
		{"candidate caption", func(d *Dataset) { d.Parents[0].Candidates[0].Text += " changed" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := cloneDataset(d)
			tc.edit(&changed)
			if ValidateDataset(changed) == nil {
				t.Fatal("forged or upgraded pending proposal was accepted")
			}
		})
	}
}

func TestPreparationFactoriesOwnSlicesAndPreserveFrozenOriginal(t *testing.T) {
	original, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	before, e := json.Marshal(original)
	if e != nil {
		t.Fatal(e)
	}
	d, e := CreateDataset()
	if e != nil {
		t.Fatal(e)
	}
	d.Parents[0].Candidates[0].Text = "changed"
	d.Parents[1].Request = "changed"
	again, e := CreateDataset()
	if e != nil || again.Parents[0].Candidates[0].Text == "changed" || again.Parents[1].Request == "changed" {
		t.Fatal("factory leaked writable proposal storage")
	}
	untouched, e := typedbehavior.PrepareDataset()
	if e != nil {
		t.Fatal(e)
	}
	after, e := json.Marshal(untouched)
	if e != nil || string(before) != string(after) {
		t.Fatal("property preparation changed the frozen original dataset")
	}
}
