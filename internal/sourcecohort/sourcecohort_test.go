package sourcecohort

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/internal/reviewpacket"
)

func save(t *testing.T, root, name string, data []byte) File {
	t.Helper()
	p := filepath.Join(root, name)
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, data, 0600); e != nil {
		t.Fatal(e)
	}
	return File{name, Hash(data), int64(len(data))}
}
func jsonFile(t *testing.T, root, name string, v any) File {
	t.Helper()
	return save(t, root, name, marshal(v))
}
func fileAt(t *testing.T, root, name string) File {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, name))
	if e != nil {
		t.Fatal(e)
	}
	return File{name, Hash(b), int64(len(b))}
}

type fixture struct {
	root     string
	plan     Plan
	planFile File
	registry Registry
	creation CreationManifest
	reviews  Reviews
}

func allocationFixture(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	recipe := DefaultRecipe()
	reg := Registry{Schema: "riido-sourcecohort-registry-v1", CohortID: "synthetic-cohort", Rows: []Row{}}
	n := 0
	for a, stratum := range recipe.Strata {
		for _, split := range recipe.Splits {
			for j := 0; j < split.PerStratum; j++ {
				style := ""
				if split.Name == "dev" {
					style = "short"
					if a >= 6 {
						style = "general"
					}
				}
				reg.Rows = append(reg.Rows, Row{ID: syntheticID(n), Stratum: stratum, Split: split.Name, DEVStyle: style, Dependencies: []string{}})
				n++
			}
		}
	}
	plan := Plan{Schema: "riido-sourcecohort-plan-v1", Registry: jsonFile(t, root, "registry-planned.json", reg), SplitConfig: jsonFile(t, root, "recipe.json", recipe), AuthorID: "synthetic-author", CheckerID: "synthetic-checker", InputMaxBytes: 4 << 20, SourceMaxBytes: 16384, FreezeMaxBytes: 16 << 20, OutputMaxBytes: 32 << 20}
	return fixture{root: root, plan: plan, planFile: jsonFile(t, root, "allocation-plan.json", plan), registry: reg}
}
func syntheticID(i int) string { return fmt.Sprintf("synthetic-%03d", i) }
func fullFixture(t *testing.T) fixture {
	return fullFixtureWithReviewer(t, func(_ int, checker string) string { return checker })
}
func fullFixtureWithReviewer(t *testing.T, reviewerFor func(int, string) string) fixture {
	t.Helper()
	f := allocationFixture(t)
	s, e := Run(f.root, f.planFile, filepath.Join(f.root, "allocation-out"), true)
	if e != nil || s.State != "allocation_ready" {
		t.Fatalf("allocation: %v %+v", e, s)
	}
	allocation := fileAt(t, f.root, "allocation-out/ALLOCATION.private.json")
	schema := jsonFile(t, f.root, "schema.json", struct {
		Kind string `json:"kind"`
	}{"public_synthetic_schema_only"})
	f.plan.Allocation = &allocation
	f.plan.SourceSchema = &schema
	f.creation = CreationManifest{Schema: "riido-sourcecohort-creation-v1", Rows: []CreationRow{}}
	f.reviews = Reviews{Schema: "riido-sourcecohort-reviews-v1", Rows: []Review{}}
	for i, row := range f.registry.Rows {
		reviewer := reviewerFor(i, f.plan.CheckerID)
		created := time.Now().UTC().Format(time.RFC3339Nano)
		data := []byte("public synthetic situation " + row.ID + ": 한글\n")
		source := save(t, f.root, "sources/"+row.ID+".txt", data)
		f.registry.Rows[i].Source = &source
		cr := CreationReceipt{"riido-sourcecohort-creation-receipt-v1", row.ID, source, f.plan.AuthorID, "ai_nonhuman", true, created, true, []File{}, []string{}}
		creationReceipt := jsonFile(t, f.root, "creation-receipts/"+row.ID+".json", cr)
		startName := "review-reads/" + row.ID + ".json"
		if e := os.MkdirAll(filepath.Join(f.root, "review-reads"), 0700); e != nil {
			t.Fatal(e)
		}
		_, out, e := reviewpacket.Capture(reviewpacket.Config{InputPath: filepath.Join(f.root, source.Path), ReceiptPath: filepath.Join(f.root, startName), ExpectedSHA256: source.SHA256, MaxBytes: 16384, Actor: reviewer})
		if e != nil || !out.ResultDurable {
			t.Fatal(e)
		}
		obs := jsonFile(t, f.root, "observations/"+row.ID+".json", struct {
			Scope string `json:"scope"`
		}{"public_synthetic_observation_only"})
		report := jsonFile(t, f.root, "checks/"+row.ID+".json", CheckReport{true, true, true, "", "public synthetic declared check only"})
		binding := CheckBinding{"riido-sourcecohort-check-binding-v1", source, obs, schema, report, time.Now().UTC().Format(time.RFC3339Nano)}
		bindingFile := jsonFile(t, f.root, "bindings/"+row.ID+".json", binding)
		f.creation.Rows = append(f.creation.Rows, CreationRow{row.ID, source, f.plan.AuthorID, "ai_nonhuman", true, created, []string{}, creationReceipt})
		f.reviews.Rows = append(f.reviews.Rows, Review{ID: row.ID, Source: source, SourceSchema: schema, ReviewerID: reviewer, ReviewedUTC: time.Now().UTC().Format(time.RFC3339Nano), Complete: true, Structural: true, SourceScope: "complete_source_situation", SemanticVerdict: "declared_pass", Holds: []Hold{}, Evidence: []Span{{"source_scope", 0, len(data)}}, Dependencies: []string{}, ReadStart: fileAt(t, f.root, startName), ReadResult: fileAt(t, f.root, startName+".result"), Observation: obs, CheckBinding: bindingFile})
	}
	f.refresh(t)
	return f
}
func (f *fixture) refresh(t *testing.T) {
	f.plan.Registry = jsonFile(t, f.root, "registry-realized.json", f.registry)
	c := jsonFile(t, f.root, "creation.json", f.creation)
	a := jsonFile(t, f.root, "reviews.json", f.reviews)
	f.plan.Creation = &c
	f.plan.Reviews = &a
	f.planFile = jsonFile(t, f.root, "freeze-plan.json", f.plan)
}
func TestWhole400AllocationBeforeMaterial(t *testing.T) {
	f := allocationFixture(t)
	s, e := Run(f.root, f.planFile, filepath.Join(f.root, "out"), true)
	if e != nil || s.Expected != 400 || s.Registered != 400 || s.Verified != 0 || s.State != "allocation_ready" || s.MeaningProven {
		t.Fatalf("allocation %+v %v", s, e)
	}
	if _, e := os.Stat(filepath.Join(f.root, "sources")); !os.IsNotExist(e) {
		t.Fatal("allocation must not require sources")
	}
}
func TestWhole400FreezeAndNoOverwrite(t *testing.T) {
	f := fullFixture(t)
	f.creation.Rows[0].Dependencies = []string{f.registry.Rows[1].ID}
	f.reviews.Rows[1].Dependencies = []string{f.registry.Rows[0].ID}
	f.refresh(t)
	out := filepath.Join(f.root, "freeze-out")
	s, e := Run(f.root, f.planFile, out, false)
	if e != nil {
		t.Fatal(e)
	}
	if s.Expected != 400 || s.Verified != 400 || s.Held != 0 || s.Components != 399 || s.State != "structural_provenance_freeze_ready" || s.MeaningProven || s.ProviderIndependence != "unknown" {
		t.Fatalf("wrong aggregate %+v", s)
	}
	b, e := os.ReadFile(filepath.Join(out, "FREEZE.private.json"))
	if e != nil {
		t.Fatal(e)
	}
	var frozen Freeze
	if e := Decode(b, &frozen); e != nil || len(frozen.Rows) != 400 || frozen.Rows[0].ObservationJSON == "" {
		t.Fatal("full joined observations lost", e)
	}
	info, _ := os.Stat(filepath.Join(out, "FREEZE.private.json"))
	if info.Mode().Perm() != 0600 {
		t.Fatal("freeze must be private")
	}
	_, e = Run(f.root, f.planFile, out, false)
	if e == nil {
		t.Fatal("existing output overwritten")
	}
	after, _ := os.ReadFile(filepath.Join(out, "FREEZE.private.json"))
	if !bytes.Equal(after, b) {
		t.Fatal("old freeze modified")
	}
}
func TestAllocationRejectsSubsetAndBadSlots(t *testing.T) {
	f := allocationFixture(t)
	for _, edit := range []func(*Registry, *Recipe){func(r *Registry, c *Recipe) { r.Rows = r.Rows[:399] }, func(r *Registry, c *Recipe) { r.Rows = append(r.Rows, r.Rows[0]) }, func(r *Registry, c *Recipe) { r.Rows[1].ID = r.Rows[0].ID }, func(r *Registry, c *Recipe) {
		c.Strata = c.Strata[:1]
		c.Splits = []Split{{"train", 1}, {"dev", 1}, {"cal", 1}, {"test", 1}}
		c.DEVShort = 1
		c.DEVGeneral = 0
	}, func(r *Registry, c *Recipe) { r.Rows[0].Split = "dev" }, func(r *Registry, c *Recipe) { r.Rows[35].DEVStyle = "general" }, func(r *Registry, c *Recipe) { r.Rows[0].Dependencies = []string{"missing"} }, func(r *Registry, c *Recipe) { r.Rows[0].Dependencies = []string{r.Rows[35].ID} }, func(r *Registry, c *Recipe) { r.Rows[0].Dependencies = []string{r.Rows[1].ID, r.Rows[1].ID} }} {
		raw := marshal(f.registry)
		var reg Registry
		Decode(raw, &reg)
		recipe := DefaultRecipe()
		edit(&reg, &recipe)
		s, e := ValidateAllocation(reg, recipe)
		if e == nil || s.State == "allocation_ready" {
			t.Fatal("bad whole allocation admitted")
		}
	}
}
func TestHoldsRetainWholeDenominatorAndPreventExport(t *testing.T) {
	f := fullFixture(t)
	f.reviews.Rows[0].Holds = []Hold{{"semantic", "public synthetic hold"}, {"source_scope", "public synthetic hold"}}
	f.reviews.Rows[399].Complete = false
	f.creation.Rows[0].RightsAllowed = false
	f.refresh(t)
	out := filepath.Join(f.root, "held")
	s, e := Run(f.root, f.planFile, out, false)
	if e == nil || s.Expected != 400 || s.Held != 2 || s.Reviewed != 400 || s.Holds[countAt(s.Holds, "semantic")].Rows != 1 || s.Holds[countAt(s.Holds, "rights")].Rows != 1 {
		t.Fatalf("hidden holds %+v %v", s, e)
	}
	if _, e := os.Stat(filepath.Join(out, "FREEZE.private.json")); !os.IsNotExist(e) {
		t.Fatal("held subset exported")
	}
}
func TestJoinPinEvidenceChronologyDependencyRightsAndSizeFailures(t *testing.T) {
	f := fullFixture(t)
	tests := []struct {
		name string
		edit func(*fixture)
	}{{"missing", func(f *fixture) { f.reviews.Rows = f.reviews.Rows[:399] }}, {"extra", func(f *fixture) { f.reviews.Rows = append(f.reviews.Rows, f.reviews.Rows[0]) }}, {"duplicate", func(f *fixture) { f.reviews.Rows[1].ID = f.reviews.Rows[0].ID }}, {"source-pin", func(f *fixture) { f.creation.Rows[0].Source.SHA256 = strings.Repeat("0", 64) }}, {"evidence", func(f *fixture) { f.reviews.Rows[0].Evidence = nil }}, {"utf8-boundary", func(f *fixture) {
		b, _ := os.ReadFile(filepath.Join(f.root, f.creation.Rows[0].Source.Path))
		a := bytes.Index(b, []byte("한"))
		f.reviews.Rows[0].Evidence = append(f.reviews.Rows[0].Evidence, Span{"proposition", a + 1, a + 3})
	}}, {"chronology", func(f *fixture) { f.reviews.Rows[0].ReviewedUTC = "2000-01-01T00:00:00Z" }}, {"unknown-dependency", func(f *fixture) { f.creation.Rows[0].Dependencies = []string{"missing"} }}, {"cross-split", func(f *fixture) { f.reviews.Rows[0].Dependencies = []string{f.registry.Rows[35].ID} }}, {"rights", func(f *fixture) { f.creation.Rows[0].CreatorKind = "human" }}, {"size", func(f *fixture) { f.plan.FreezeMaxBytes = 1 }}, {"same-role", func(f *fixture) { f.plan.CheckerID = f.plan.AuthorID }}, {"byte-pin", func(f *fixture) { f.plan.Reviews.Bytes++ }}, {"output-cap", func(f *fixture) { f.plan.OutputMaxBytes = 1 }}}
	baseReg, baseCreation, baseReviews := marshal(f.registry), marshal(f.creation), marshal(f.reviews)
	basePlan := f.plan
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := f
			Decode(baseReg, &g.registry)
			Decode(baseCreation, &g.creation)
			Decode(baseReviews, &g.reviews)
			g.plan = basePlan
			tc.edit(&g)
			g.refresh(t)
			if tc.name == "byte-pin" {
				g.plan.Reviews.Bytes++
				g.planFile = jsonFile(t, g.root, "freeze-plan.json", g.plan)
			}
			s, e := Run(g.root, g.planFile, filepath.Join(g.root, "failure-"+tc.name), false)
			if e == nil || s.State != "blocked" || s.Expected != 400 {
				t.Fatalf("bad input admitted %+v", s)
			}
			if _, e := os.Stat(filepath.Join(g.root, "failure-"+tc.name, "FREEZE.private.json")); !os.IsNotExist(e) {
				t.Fatal("failure exported freeze")
			}
		})
	}
}
func TestStrictJSONAndPointerReset(t *testing.T) {
	f := allocationFixture(t)
	good := marshal(f.plan)
	for _, b := range [][]byte{append(good, []byte("{}")...), bytes.Replace(good, []byte(`"schema":`), []byte(`"Schema":`), 1), bytes.Replace(good, []byte(`"schema":`), []byte(`"extra":1,"schema":`), 1), bytes.Replace(good, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1), bytes.Replace(good, []byte(`"author_id":"synthetic-author"`), []byte(`"author_id":"\ud800"`), 1), bytes.Replace(good, []byte(`"creation":null,`), nil, 1)} {
		var p Plan
		if Decode(b, &p) == nil {
			t.Fatal("invalid closed JSON accepted")
		}
	}
	p := f.plan
	x := File{"old", Hash([]byte("x")), 1}
	p.Creation = &x
	if e := Decode(good, &p); e != nil || p.Creation != nil {
		t.Fatal("null reused old pointer")
	}
	var v struct {
		Text string `json:"text"`
	}
	if Decode([]byte(`{"text":"\ud83d\ude00"}`), &v) != nil || v.Text != "😀" {
		t.Fatal("valid surrogate pair rejected")
	}
}
