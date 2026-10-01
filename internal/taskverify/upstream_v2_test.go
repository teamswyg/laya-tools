package taskverify

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestUpstreamV2PreservesSevenFrozenSpecs(t *testing.T) {
	wants := []struct{ id, spec, definition, contract, prompt string }{
		{"comment-preview-authority", "a01f1c95e818511d6bcac35a0eb20b0b6c943a73b4fe7652a550bb0036dbd747", "", "", ""},
		{"comment-budget-period", "00333f0814e4bcb28bc06b9172f21672dace61eea3f6537af40dfe77a3c59fe4", "", "", ""},
		{"catalog-min-context", "884eb30b9310817c78fe8071e20881b865bb61a603a75327e83d39f6625d7e25", "", "", ""},
		{keywordGuardTask, "fb71960dd44348b9f7531ba3c7b44e8bb88c89196bd43dc362328120efd2a644", "f1782b575818b0ea4e8c29e610c21f25a2c44d4cfdca8270c2664d1cddda7bb2", "e44ea077ad9fdb09775e49694da4aecaa630070a302b3763a7599651775bc89a", "bd43cebca6b00049f88e9010405068e8c57908e91285a115f23b30db7b3e1ab7"},
		{eventKeyBoundsTask, "169cedefd41d3a5ba05d11bbb4685761a0dde6adf07401b846b77bf5a84c27db", "408255ae28f7d6bfb8994d4a9cbcbcaa46edd95c1cf2f606fe847cd28e0599a9", "c9873cabd6b921516103bd8bd8bde8987d02c9aca748522c141aff1005896b71", "f395e7c85eb8398764ce189fca23709465c372a8710ba1be2583f0d99690a567"},
		{humanizeOrdinalTask, "5c6fd6058a48bd0e6441a68c351e7b1ed491aec20baff78c6942e377ae79f5d5", "99a7c0723e3c9f40a069003ccf3f9c2e290c3cbe8e67be3a4ced02d4a31a29b8", "3b1351162b746fa61f7837af2923644eb2f4df100de643114c9d9f0eb064fd06", "107ae4d6da023114cc615ebdef6709053be8471bb0ee37f9763af73f279a94a9"},
		{uuidCanonicalTask, "c8e3a2edcc2024ce68ba7ea7b5fed00554b2e1c36a12174f942f32b5a5a5eec3", "9413ed54a10ee9a1576f71a68cf664134570cf84e735f8daeb88c66c1f4ad7c8", "259f3fcd56ac9bf0d4d307cc1b8ed77968b6b3bf21327160e09efd440f71ffd5", "5c6456db8da12a5e9510cf9f39a448298df4dd73bead5706a41288f567f30cb2"},
	}
	for _, want := range wants {
		t.Run(want.id, func(t *testing.T) {
			s, err := TaskSpec(want.id)
			b, marshalErr := json.Marshal(s)
			if err != nil || marshalErr != nil || digest(b) != want.spec || s.LogicalTaskID != "" || s.PreviousTaskID != "" || s.EvaluationRecipeSHA256 != "" {
				t.Fatal("frozen task spec changed in place")
			}
			if want.definition != "" {
				d, ok := TaskDefinition(want.id)
				db, _ := json.Marshal(d)
				contract, found := definitionContractSource(want.id)
				if !ok || !found || digest(db) != want.definition || digest([]byte(contract)) != want.contract || digest([]byte(s.Prompt)) != want.prompt {
					t.Fatal("frozen definition/contract/prompt changed")
				}
			}
			recipe, err := TaskEvaluationRecipe(want.id)
			if err != nil || recipe.Kind != minimalModuleRecipe || recipe.LanguageVersion != "1.27.1" || recipe.ToolchainVersion != "go1.27.1" || recipe.ModuleSHA256 != digest([]byte("module "+recipe.ModulePath+"\n\ngo 1.27.1\n")) {
				t.Fatal("legacy module recipe changed")
			}
		})
	}
}

func TestUpstreamV2IdentityAndCompletePrompt(t *testing.T) {
	for _, pair := range []struct{ id, old, language string }{
		{humanizeOrdinalTaskV2, humanizeOrdinalTask, "1.21"},
		{uuidCanonicalTaskV2, uuidCanonicalTask, "1.16"},
	} {
		d, ok := TaskDefinition(pair.id)
		old, _ := TaskDefinition(pair.old)
		s, err := TaskSpec(pair.id)
		recipe, recipeErr := TaskEvaluationRecipe(pair.id)
		db, _ := json.Marshal(d)
		contract, found := definitionContractSource(pair.id)
		if !ok || err != nil || recipeErr != nil || !found || s.Version != 2 || d.Version != 2 || s.LogicalTaskID != pair.old || s.PreviousTaskID != pair.old || d.LogicalTaskID != pair.old || d.PreviousTaskID != pair.old {
			t.Fatal("clarified versions not linked to same logical request")
		}
		if !slices.Equal(d.Files, old.Files) || !slices.Equal(d.RequiredTests, old.RequiredTests) || d.BaseRevision != old.BaseRevision || d.ContractSHA256 != old.ContractSHA256 || digest([]byte(contract)) != old.ContractSHA256 || s.DefinitionSHA256 != digest(db) || d.PromptSHA256 != digest([]byte(s.Prompt)) || s.EvaluationRecipeSHA256 != recipe.RecipeSHA256 || d.EvaluationRecipeSHA256 != recipe.RecipeSHA256 {
			t.Fatal("v2 changed original closure/contracts or broke identities")
		}
		if recipe.Kind != upstreamModuleRecipe || recipe.ModulePath != d.ModulePath || recipe.ModuleSHA256 != definitionPin(old, "go.mod") || recipe.LanguageVersion != pair.language || recipe.ToolchainVersion != "go1.27.1" {
			t.Fatal("upstream language/toolchain not bound")
		}
		withoutSHA := recipe
		withoutSHA.RecipeSHA256 = ""
		rb, _ := json.Marshal(withoutSHA)
		if digest(rb) != recipe.RecipeSHA256 {
			t.Fatal("recipe has a self-hash cycle")
		}
		for _, required := range d.RequiredTests {
			if !strings.Contains(s.Prompt, required.Test) || !strings.Contains(s.Prompt, required.Package) {
				t.Fatal("terminal test requirement hidden from stdin")
			}
		}
		for _, text := range []string{"Only " + d.SourcePath + " is mutable", "Candidate-authored", "Candidate LICENSE contents are outside candidate source acceptance", "acceptance does not certify candidate or downstream distribution attribution compliance", "go1.27.1", pair.language, "verifier_unknown", pair.old, d.BaseRevision, d.License} {
			if !strings.Contains(s.Prompt, text) {
				t.Fatal("evaluation condition hidden from actual Prompt", text)
			}
		}
		sb, _ := json.Marshal(s)
		t.Logf("v2 id=%s spec=%s definition=%s contract=%s prompt=%s recipe=%s module=%s language=%s", pair.id, digest(sb), digest(db), d.ContractSHA256, d.PromptSHA256, recipe.RecipeSHA256, recipe.ModuleSHA256, recipe.LanguageVersion)
	}
	if _, err := TaskEvaluationRecipe("unknown"); err == nil {
		t.Fatal("unregistered recipe accepted")
	}
	// Exact named-call disclosure must track the retained checker, not vaguely
	// describe every operation from an otherwise allowed package as supported.
	for _, pkg := range []string{"strings", "bytes", "hex", "fmt", "errors"} {
		for _, name := range []string{"ToLower", "ToUpper", "TrimSpace", "TrimPrefix", "TrimSuffix", "Trim", "ReplaceAll", "Replace", "EqualFold", "HasPrefix", "HasSuffix", "Index", "IndexByte", "Contains", "Count", "Repeat", "Decode", "DecodeString", "EncodeToString", "DecodedLen", "Equal", "Errorf", "Sprintf", "New"} {
			if uuidCanonicalPureSelector(&ast.Ident{Name: pkg}, name, nil) && !strings.Contains(uuidCanonicalPromptV2, pkg+"."+name) {
				t.Fatal("supported package call undisclosed", pkg, name)
			}
		}
	}
}

func TestUpstreamV2ModuleTrustBoundary(t *testing.T) {
	humanize, humanizeLicense := humanizeOrdinalFixture(t)
	uuid := uuidCanonicalFixture(t)
	for _, tc := range []struct {
		id    string
		files []BaseFile
	}{
		{humanizeOrdinalTaskV2, append(slices.Clone(humanize), humanizeLicense...)},
		{uuidCanonicalTaskV2, uuid},
	} {
		module, err := trustedEvaluationModule(tc.id, tc.files)
		if err != nil || !bytes.Equal(module, fileAt(tc.files, "go.mod")) {
			t.Fatal("original trusted module unavailable")
		}
		for _, mutation := range []string{"missing", "duplicate", "wrong-sha", "module", "require", "replace", "toolchain", "go-version"} {
			t.Run(tc.id+"/"+mutation, func(t *testing.T) {
				files := slices.Clone(tc.files)
				for i, f := range files {
					if f.Path != "go.mod" {
						continue
					}
					switch mutation {
					case "missing":
						files = slices.Delete(files, i, i+1)
					case "duplicate":
						files = append(files, f)
					case "wrong-sha":
						files[i].SHA256 = strings.Repeat("0", 64)
					default:
						body := string(module)
						switch mutation {
						case "module":
							body = "module example.com/other\n"
						case "require":
							body += "require example.com/private v1.0.0\n"
						case "replace":
							body += "replace example.com/x => ./external\n"
						case "toolchain":
							body += "toolchain go1.27.1\n"
						case "go-version":
							body += "go 1.27.1\n"
						}
						files[i].Data, files[i].SHA256 = []byte(body), digest([]byte(body))
					}
					break
				}
				out := isolatedTests(context.Background(), files, tc.id, MaxTimeout, "")
				if !out.unknown || out.isolated || out.passed {
					t.Fatal("supplied config reached sandbox", mutation, out)
				}
			})
		}
	}
	for _, body := range []string{"", "module", "module example.com/x require y v1.0.0", "module example.com/x replace y => z", "module example.com/x go 1.21 toolchain go1.27.1"} {
		if validUpstreamModule([]byte(body), "example.com/x", "1.21") {
			t.Fatal("dependency/configuration grammar accepted")
		}
	}
}

func TestUpstreamV2FrozenIdentity(t *testing.T) {
	for _, want := range []struct{ id, spec, definition, prompt, recipe string }{
		{humanizeOrdinalTaskV2, "48f3862192f86845027e31260073cda9d1960e5c4bf9cbe94ff0af5edc9a2d1b", "82c39fbb6f20fcc9b4fba7b08e7f85b85f3071d902c1f183da04794566f56245", "6a613343570d4cc40ec24dc2a8fb4a3bd1e5b52a80132811cb7d6cc6af92f387", "f01aecc61bc3e070675a636df28b19e2bec2466359d439771bb1d8820c9406a1"},
		{uuidCanonicalTaskV2, "d764e69a4a9634487c6e9b0f22a3009d446efa1284af10057f38da4e0e93ba34", "1c1df765a8b2020e5050069653ce5a2a1a6f94a456ef863c4c096644dfaa51a0", "d900f561e67bca653032d3974b7afa49f8041cf9ef497b112d559bd0699c49a5", "55cbd876d1cfb6200ff96abb3bd8b0b1c2f16d0c882ee1b2b9a22bfa26c458ad"},
	} {
		s, err := TaskSpec(want.id)
		d, ok := TaskDefinition(want.id)
		r, recipeErr := TaskEvaluationRecipe(want.id)
		sb, _ := json.Marshal(s)
		db, _ := json.Marshal(d)
		if err != nil || !ok || recipeErr != nil || digest(sb) != want.spec || digest(db) != want.definition || digest([]byte(s.Prompt)) != want.prompt || r.RecipeSHA256 != want.recipe {
			t.Fatal("v2 clarified evaluation identity changed in place")
		}
	}
}

func TestUpstreamV2UnchangedBaseline(t *testing.T) {
	humanize, attribution := humanizeOrdinalFixture(t)
	for _, tc := range []struct {
		id    string
		files []BaseFile
	}{
		{humanizeOrdinalTaskV2, append(slices.Clone(humanize), attribution...)},
		{uuidCanonicalTaskV2, uuidCanonicalFixture(t)},
	} {
		d, _ := TaskDefinition(tc.id)
		var base, attribution []BaseFile
		for _, f := range tc.files {
			if f.Path == "LICENSE" {
				attribution = append(attribution, f)
			} else {
				base = append(base, f)
			}
		}
		r, err := Verify(context.Background(), Request{TaskID: tc.id, BaseRevision: d.BaseRevision, BaseFiles: base, AttributionFiles: attribution, CandidateDir: candidateFixture(t, tc.files)})
		if err != nil || r.Accepted || r.Status != "rejected" || r.ExecutionIsolated || r.Checks[len(r.Checks)-1].Code != "unchanged_base" || r.EvaluationRecipe == nil {
			t.Fatal("v2 unchanged baseline accepted or incorrectly attributed", r, err)
		}
	}
}

// The positive control uses safe unsigned magnitude arithmetic, independently
// of the contract's big.Int oracle and the v1 control's string-tail approach.
const humanizeV2Correct = `
func Ordinal64(x int64) string {
 magnitude := uint64(x)
 if x < 0 { magnitude = uint64(-(x+1)) + 1 }
 suffix := "th"
 residue := magnitude % 100
 if residue < 11 || residue > 13 {
  switch magnitude % 10 { case 1: suffix="st"; case 2: suffix="nd"; case 3: suffix="rd" }
 }
 return strconv.FormatInt(x,10) + suffix
}
`

func upstreamV2Request(t *testing.T, id string, files []BaseFile, source string) Request {
	t.Helper()
	d, _ := TaskDefinition(id)
	var base, attribution []BaseFile
	for _, f := range files {
		if f.Path == "LICENSE" || f.Path == "NOTICE" {
			attribution = append(attribution, f)
		} else {
			base = append(base, f)
		}
	}
	dir := candidateFixture(t, files)
	formatted, err := format.Source([]byte(source))
	if err != nil {
		t.Fatal("invalid authored candidate", err)
	}
	if err := os.WriteFile(filepath.Join(dir, d.SourcePath), formatted, 0600); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Path, "_test.go") {
			if err := os.WriteFile(filepath.Join(dir, f.Path), []byte("deliberately invalid candidate assertions\n"), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Unassessed candidate configuration must never enter the isolated compiler.
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte("deliberately invalid candidate workspace\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return Request{TaskID: id, BaseRevision: d.BaseRevision, BaseFiles: base, AttributionFiles: attribution, CandidateDir: dir, Timeout: MaxTimeout}
}

func checkUpstreamV2Report(t *testing.T, req Request, status string, isolated bool) Report {
	t.Helper()
	before, err := os.ReadFile(filepath.Join(req.CandidateDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filepath.Join(req.CandidateDir, "go.mod"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("verification changed candidate module")
	}
	if isolated && !sandboxSupported() {
		status, isolated = "verifier_unknown", false
	}
	if r.Status != status || r.ExecutionIsolated != isolated || r.Accepted != (status == "accepted") || r.CandidateTestsUsed || r.EvaluationRecipe == nil || r.EvaluationRecipe.Kind != upstreamModuleRecipe {
		t.Fatal("v2 independent status, isolation or recipe wrong", r)
	}
	if status == "accepted" && r.IndependentTests < 6 {
		t.Fatal("named contract tests did not terminate")
	}
	t.Logf("v2 development task=%s status=%s isolated=%t terminal_passes=%d candidate_tests_used=false models=0", req.TaskID, r.Status, r.ExecutionIsolated, r.IndependentTests)
	return r
}

func TestUpstreamV2HumanizeControls(t *testing.T) {
	base, attribution := humanizeOrdinalFixture(t)
	files := append(slices.Clone(base), attribution...)
	original := string(fileAt(base, "ordinals.go"))
	for _, mode := range []string{"correct", "teen-error", "import-unknown", "changed-module"} {
		t.Run(mode, func(t *testing.T) {
			source, status, isolated := original+humanizeV2Correct, "rejected", true
			switch mode {
			case "correct":
				status = "accepted"
			case "teen-error":
				source = strings.Replace(source, "residue < 11 || residue > 13", "residue != 1000", 1)
			case "import-unknown":
				source = strings.Replace(source, `import "strconv"`, "import (\"strconv\";\"fmt\")", 1)
				status, isolated = "verifier_unknown", false
			case "changed-module":
				isolated = false
			}
			req := upstreamV2Request(t, humanizeOrdinalTaskV2, files, source)
			if mode == "changed-module" {
				if err := os.WriteFile(filepath.Join(req.CandidateDir, "go.mod"), []byte("module example.com/other\n\ngo 1.27.1\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			checkUpstreamV2Report(t, req, status, isolated)
		})
	}
}

func TestUpstreamV2UUIDControls(t *testing.T) {
	files := uuidCanonicalFixture(t)
	original := uuidCanonicalMutable(t, files)
	for _, mode := range []string{"correct", "permissive", "entropy-unknown", "header-unknown", "helper-unknown", "changed-module"} {
		t.Run(mode, func(t *testing.T) {
			source, status, isolated := original+uuidCanonicalCorrect, "rejected", true
			switch mode {
			case "correct":
				status = "accepted"
			case "permissive":
				source = original + "\nfunc ParseCanonical(s string)(UUID,error){ return Parse(s) }\n"
			case "entropy-unknown":
				source = strings.Replace(source, "if len(s)!=36", "b:=make([]byte,1);_,_=rand.Read(b); if len(s)!=36", 1)
				status, isolated = "verifier_unknown", false
			case "header-unknown":
				source = strings.TrimPrefix(source, "// Copyright 2018 Google Inc.  All rights reserved.\n")
				status, isolated = "verifier_unknown", false
			case "helper-unknown":
				source += "\nfunc unusedCanonicalHelper(s string)(UUID,error){return Parse(s)}\n"
				status, isolated = "verifier_unknown", false
			case "changed-module":
				isolated = false
			}
			req := upstreamV2Request(t, uuidCanonicalTaskV2, files, source)
			if mode == "changed-module" {
				if err := os.WriteFile(filepath.Join(req.CandidateDir, "go.mod"), []byte(uuidOriginalModule+"go 1.27.1\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			checkUpstreamV2Report(t, req, status, isolated)
		})
	}
}

func TestUpstreamV2OriginalLanguageBoundary(t *testing.T) {
	if !sandboxSupported() {
		t.Skip("language boundary requires real offline isolation")
	}
	humanize, attribution := humanizeOrdinalFixture(t)
	for _, tc := range []struct {
		id, old        string
		files          []BaseFile
		source, anchor string
	}{
		{humanizeOrdinalTaskV2, humanizeOrdinalTask, append(slices.Clone(humanize), attribution...), string(fileAt(humanize, "ordinals.go")) + humanizeV2Correct, "magnitude := uint64(x)"},
		{uuidCanonicalTaskV2, uuidCanonicalTask, uuidCanonicalFixture(t), uuidCanonicalMutable(t, uuidCanonicalFixture(t)) + uuidCanonicalCorrect, "if len(s)!=36"},
	} {
		t.Run(tc.id, func(t *testing.T) {
			// Go 1.22 integer range is semantically harmless here. The retained
			// shape checker permits its loop, but original language versions do not.
			source := strings.Replace(tc.source, tc.anchor, "for i:=range 1 { _=i }; "+tc.anchor, 1)
			req := upstreamV2Request(t, tc.id, tc.files, source)
			d, _ := TaskDefinition(tc.id)
			candidate, err := os.ReadFile(filepath.Join(req.CandidateDir, d.SourcePath))
			if err != nil || !supportedDefinitionSource(d, candidate) {
				t.Fatal("language counterexample changed supported AST shape")
			}
			r := checkUpstreamV2Report(t, req, "rejected", true)
			if r.IndependentTests != 0 {
				t.Fatal("language-incompatible candidate ran terminal tests")
			}
			req.TaskID = tc.old
			old, err := Verify(context.Background(), req)
			if err != nil || !old.Accepted || !old.ExecutionIsolated || old.EvaluationRecipe != nil || old.IndependentTests < 6 {
				t.Fatal("legacy 1.27 module recipe unexpectedly rejected valid newer syntax", old, err)
			}
			t.Logf("same source v2-original-language rejected before tests; v1-normalized-language accepted=%t passes=%d; models=0", old.Accepted, old.IndependentTests)
		})
	}
}
