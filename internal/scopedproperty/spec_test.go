package scopedproperty

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/typedbehavior"
)

func TestIndependentLiteralDefinitionsAndFullObservations(t *testing.T) {
	defs := Definitions()
	wantInputs := [3][5]string{
		{`"a,b",c`, `ab"c,d"ef`},
		{`a\,b,c`, `a\\,b`},
		{`"a,b`, `a\`, `a,b\`, `"a\`, `a,"b`},
	}
	wantCounts := [3]int{2, 2, 5}
	total := 0
	for i, def := range defs {
		if def.Version != 1 || def.CaptionReview != "pending" || !def.NoPanic || !def.RequireExactErrorKind || def.ObservationFields != [4]string{"error", "count", "tokens", "panicked"} || def.LiteralCount != wantCounts[i] || def.Scope == "" || def.Exclusions == "" || def.UnknownPolicy == "" {
			t.Fatal("definition omitted scope, finite observations or pending review")
		}
		for j, literal := range def.Literals {
			if j >= def.LiteralCount {
				if literal != (Literal{}) {
					t.Fatal("unused literal capacity gained a hidden input")
				}
				continue
			}
			total++
			if literal.Input != wantInputs[i][j] || literal.Want.Panicked {
				t.Fatal("literal inputs or no-panic observation differ from independent table")
			}
			if i == 2 {
				if literal.Want != (Observation{Error: typedbehavior.PropertyQuotedSyntax}) || !def.WholeArrayZeroRequired || def.ExpectedErrorKind != typedbehavior.PropertyQuotedSyntax {
					t.Fatal("syntax property permits omitted error, partial Count or residual slots")
				}
			} else if literal.Want.Error != typedbehavior.PropertyQuotedOK || def.ExpectedErrorKind != typedbehavior.PropertyQuotedOK || def.WholeArrayZeroRequired {
				t.Fatal("successful property changed the exact expected OK kind")
			}
		}
	}
	if total != 9 || defs[0].Literals[0].Want != (Observation{Tokens: [8]string{"a,b", "c"}, Count: 2}) || defs[0].Literals[1].Want != (Observation{Tokens: [8]string{"abc,def"}, Count: 1}) || defs[1].Literals[0].Want != (Observation{Tokens: [8]string{"a,b", "c"}, Count: 2}) || defs[1].Literals[1].Want != (Observation{Tokens: [8]string{`a\`, "b"}, Count: 2}) {
		t.Fatal("independent full Count/token expectations changed")
	}
}

func TestTruthHashBindsVersionFieldsPolicyAndEverySlot(t *testing.T) {
	d := Definitions()[2]
	if len(d.TruthTableSHA256) != 64 || tableSHA(d) != d.TruthTableSHA256 {
		t.Fatal("table hash was not bound to the actual definition")
	}
	for _, edit := range []func(*PropertyDefinition){
		func(d *PropertyDefinition) { d.Version++ },
		func(d *PropertyDefinition) { d.Scope += "changed" },
		func(d *PropertyDefinition) { d.Exclusions += "changed" },
		func(d *PropertyDefinition) { d.RequireExactErrorKind = false },
		func(d *PropertyDefinition) { d.ExpectedErrorKind = typedbehavior.PropertyQuotedOK },
		func(d *PropertyDefinition) { d.NoPanic = false },
		func(d *PropertyDefinition) { d.WholeArrayZeroRequired = false },
		func(d *PropertyDefinition) { d.ObservationFields[2] = "prefix_only" },
		func(d *PropertyDefinition) { d.UnknownPolicy = "failed_label" },
		func(d *PropertyDefinition) { d.Literals[4].Input = "different" },
		func(d *PropertyDefinition) { d.Literals[0].Want.Error = typedbehavior.PropertyQuotedOK },
		func(d *PropertyDefinition) { d.Literals[0].Want.Count = 1 },
		func(d *PropertyDefinition) { d.Literals[0].Want.Tokens[7] = "residual" },
		func(d *PropertyDefinition) { d.Literals[0].Want.Panicked = true },
	} {
		changed := d
		edit(&changed)
		if tableSHA(changed) == d.TruthTableSHA256 {
			t.Fatal("property meaning changed without changing the table pin")
		}
	}
}

func TestRegistryAndDefinitionReturnsOwnTheirArrays(t *testing.T) {
	sources, parents, defs := SourceIDs(), OriginalParentIDs(), Definitions()
	sources[0], parents[0], defs[0].Literals[0].Want.Tokens[0] = "changed", "changed", "changed"
	if SourceIDs()[0] != "quoted-correct" || OriginalParentIDs()[0] != "typed56b-p05-1" || Definitions()[0].Literals[0].Want.Tokens[0] != "a,b" {
		t.Fatal("returned metadata exposed mutable shared state")
	}
	raw, e := json.Marshal(Definitions())
	if e != nil {
		t.Fatal(e)
	}
	for _, forbidden := range []string{"expected_control", "acceptable_candidate_indices", "failed_vectors", "training_ready", "whole_contract_role", "error_required", "required_error"} {
		if strings.Contains(string(raw), `"`+forbidden+`"`) {
			t.Fatal("property definitions reused roles or calculated outcome labels")
		}
	}
	if strings.Count(string(raw), `"require_exact_error_kind":true`) != 3 || strings.Count(string(raw), `"expected_error_kind":0`) != 2 || strings.Count(string(raw), `"expected_error_kind":1`) != 1 {
		t.Fatal("wire definitions omit exact OK and syntax comparison policies")
	}
}

func TestPreparationSourceArtifactsMatchExactCompiledFiles(t *testing.T) {
	artifacts := SourceArtifacts()
	if len(artifacts) != 2 || artifacts[0].Name != "spec.go" || artifacts[1].Name != "dataset.go" {
		t.Fatal("preparation compiled manifest did not bind exactly two owned files")
	}
	for _, artifact := range artifacts {
		raw, e := os.ReadFile(artifact.Name)
		if e != nil || artifact.SHA256 != sha(raw) {
			t.Fatal("disk and compiled preparation source differ")
		}
	}
	artifacts[0].SHA256 = "changed"
	if SourceArtifacts()[0].SHA256 == "changed" {
		t.Fatal("artifact results share mutable storage")
	}
}
