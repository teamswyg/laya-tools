// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Fixed source-schema/control-flow checks, using only this owned source file.
func TestFixedSourceScope(t *testing.T) {
	b, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := goparser.ParseFile(token.NewFileSet(), "main.go", b, 0)
	if err != nil {
		t.Fatal(err)
	}
	imports := []string{}
	for _, v := range f.Imports {
		s, e := strconv.Unquote(v.Path.Value)
		if e != nil {
			t.Fatal(e)
		}
		imports = append(imports, s)
	}
	sort.Strings(imports)
	want := []string{"bytes", "crypto/sha256", "encoding/hex", "encoding/json", "errors", "flag", "fmt", "io", "os", "path/filepath", "reflect", "regexp", "runtime", "sort", "strconv", "strings", "unicode/utf8"}
	if !reflect.DeepEqual(imports, want) {
		t.Fatalf("source imports changed: %v", imports)
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, d := range f.Decls {
		if d, ok := d.(*ast.FuncDecl); ok {
			funcs[d.Name.Name] = d
		}
	}
	callName := func(c *ast.CallExpr) string {
		switch x := c.Fun.(type) {
		case *ast.Ident:
			return x.Name
		case *ast.SelectorExpr:
			if p, ok := x.X.(*ast.Ident); ok {
				return p.Name + "." + x.Sel.Name
			}
		}
		return ""
	}
	calls := func(name string) map[string]token.Pos {
		r := map[string]token.Pos{}
		ast.Inspect(funcs[name], func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				key := callName(c)
				if _, found := r[key]; !found {
					r[key] = c.Pos()
				}
			}
			return true
		})
		return r
	}
	first, ok := funcs["loadInputs"].Body.List[0].(*ast.IfStmt)
	if !ok {
		t.Fatal("count guard must be first")
	}
	init, ok := first.Init.(*ast.AssignStmt)
	if !ok || len(init.Rhs) != 1 {
		t.Fatal("count guard initializer")
	}
	c, ok := init.Rhs[0].(*ast.CallExpr)
	if !ok || callName(c) != "inputCount" {
		t.Fatal("count before all preflight")
	}
	loadCalls := calls("loadInputs")
	if !(loadCalls["inputCount"] < loadCalls["preparer"] && loadCalls["preparer"] < loadCalls["inputBudget"] && loadCalls["inputBudget"] < loadCalls["loader"]) {
		t.Fatal("count/preflight/budget/load order")
	}
	for _, name := range []string{"prepare", "loadInputs", "audit", "validate"} {
		for call := range calls(name) {
			if call == "os.Open" || call == "os.ReadFile" || call == "readBounded" || call == "decodePinned" || call == "hash" {
				t.Fatalf("content operation %s in %s", call, name)
			}
		}
	}
	dc := calls("decodePinned")
	if !(dc["hash"] < dc["tree"] && dc["tree"] < dc["json.NewDecoder"]) {
		t.Fatal("pin before decode")
	}
	lc := calls("load")
	if !(lc["f.Stat"] < lc["os.SameFile"] && lc["os.SameFile"] < lc["readBounded"] && lc["readBounded"] < lc["decodePinned"]) {
		t.Fatal("opened stat/bounded read/decode order")
	}
	flagNames := []string{}
	ast.Inspect(funcs["main"], func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		key := callName(c)
		if key == "flags.Var" || key == "flags.BoolVar" {
			if len(c.Args) < 2 {
				t.Fatal("flag args")
			}
			v, ok := c.Args[1].(*ast.BasicLit)
			if !ok {
				t.Fatal("flag literal")
			}
			s, e := strconv.Unquote(v.Value)
			if e != nil {
				t.Fatal(e)
			}
			flagNames = append(flagNames, s)
		}
		return true
	})
	if !reflect.DeepEqual(flagNames, []string{"input", "self-test"}) {
		t.Fatal("unexpected input channel", flagNames)
	}
}

func TestFixedMetadataRootSchema(t *testing.T) {
	ty := reflect.TypeOf(Metadata{})
	got := []string{}
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)
		got = append(got, strings.Split(f.Tag.Get("json"), ",")[0])
		if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Interface {
			t.Fatal("schema catch-all")
		}
	}
	want := []string{"schema", "created_at_utc", "status", "scope", "metadata_boundary", "identity_scheme", "concept_design", "required_before_wording", "future_production_barriers", "groups", "across_group_relation_hypotheses", "source_boundary", "retained_provisional_metadata_correction", "relation_universe", "generation_provenance"}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("root schema changed", got)
	}
	if seed != "claims-short-core-full-v1:concepts:1729" || schema != "three-claims-short-core-full-provisional-concept-tranche-v1" || maxInput != 204800 || maxTotal != 2097152 || maxTranches != 10 || maxGroups != 400 {
		t.Fatal("fixed contract constants changed")
	}
}

// This fixture is created solely inside the owned auditor directory and removed.
func TestOwnedSyntheticReplacementBeforeRead(t *testing.T) {
	dir, err := os.MkdirTemp(".", ".synthetic-identity-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	a, b := dir+"/a.json", dir+"/b.json"
	fixture := []byte("SYNTHETIC_NOT_JSON_SENTINEL")
	if err = os.WriteFile(a, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(b, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	p := prepared{path: a, pin: hash(fixture), absolute: b, bytes: len(fixture), info: info}
	_, err = load(p)
	if err == nil || err.Error() != "preflight_input_changed" {
		t.Fatal("equal-size/equal-pin replacement must reject before JSON content decode", err)
	}
}

// Uncorrected top-level metadata omits the optional batch1 correction object.
// Initial batch0 objects occur only as predecessor objects inside batch1.
func TestUncorrectedMetadataOmitsCorrection(t *testing.T) {
	b := encode(synthetic(1))
	if strings.Contains(string(b), "retained_provisional_metadata_correction") {
		t.Fatal("uncorrected metadata must omit correction field")
	}
	m, err := decodePinned(b, hash(b))
	if err != nil || m.Correction != nil {
		t.Fatal("canonical omission must accept", err)
	}
	for _, tc := range []struct{ name, raw, want string }{
		{"explicit_null", "null", "schema_type"},
		{"top_level_initial_batch0", `{"batch_number":0,"status":"SYNTHETIC_INITIAL_METADATA_ONLY","input_rows_created_or_revised":0}`, "schema_unknown_field"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := strings.TrimSpace(string(b))
			invalid := []byte(s[:len(s)-1] + `,"retained_provisional_metadata_correction":` + tc.raw + "}\n")
			_, err := decodePinned(invalid, hash(invalid))
			if err == nil || err.Error() != tc.want {
				t.Fatalf("wanted %s; got %v", tc.want, err)
			}
		})
	}
}

func TestPresentUniverseIsExactOwnPlusPrior(t *testing.T) {
	later := []loaded{}
	for _, start := range []int{41, 81} {
		l, e := synLoad(synthetic(start))
		if e != nil {
			t.Fatal(e)
		}
		l.input.Path = fmt.Sprintf("synthetic-%d.json", start)
		later = append(later, l)
	}
	m := synthetic(1)
	m.Universe = &Universe{m.Scope.Tranche.Range, []int{1, 120}, []ExternalMetadata{{later[0].meta.Scope.Tranche.Range, later[0].input.Path, later[0].input.SHA}, {later[1].meta.Scope.Tranche.Range, later[1].input.Path, later[1].input.SHA}}, "PROVISIONAL"}
	if _, e := synLoad(m); e == nil || e.Error() != "relation_universe" {
		t.Fatal("expanded first Known with later universes omitted must reject", e)
	}
	m.Universe.Known = []int{1, 40}
	if _, e := synLoad(m); e == nil || e.Error() != "external_metadata_declaration" {
		t.Fatal("future references must reject", e)
	}
	m.Universe = nil
	first, e := synLoad(m)
	if e != nil {
		t.Fatal(e)
	}
	first.input.Path = "synthetic-1.json"
	if _, e = audit(append([]loaded{first}, later...)); e != nil {
		t.Fatal("absent-universe compatibility", e)
	}
}

func endpointCorrectionFixture(t *testing.T) Metadata {
	t.Helper()
	all, e := synTranches(3)
	if e != nil {
		t.Fatal(e)
	}
	m := all[2].meta
	priorTree, e := tree(encode(m))
	if e != nil {
		t.Fatal(e)
	}
	prior := append(compact(priorTree, nil, nil), '\n')
	oldIDs, _ := json.Marshal(m.Relations[0].IDs)
	oldOrd, _ := json.Marshal(m.Relations[0].Ordinals)
	m.Relations[0].IDs = []string{groupID(2), groupID(81)}
	m.Relations[0].Ordinals = []int{2, 81}
	newIDs, _ := json.Marshal(m.Relations[0].IDs)
	newOrd, _ := json.Marshal(m.Relations[0].Ordinals)
	m.Correction = &Correction{Batch: 1, Created: "2026-10-07T00:00:00Z", Trigger: "synthetic endpoint repair", Scope: "synthetic metadata only", PriorSHA: hash(prior), PriorBytes: len(prior), Serialization: correctionSerialization, Deltas: []Delta{{"/across_group_relation_hypotheses/0/group_ids", oldIDs, newIDs}, {"/across_group_relation_hypotheses/0/registration_ordinals", oldOrd, newOrd}}, Rows: 0}
	return m
}
func TestPairedWholeRelationEndpointCorrections(t *testing.T) {
	m := endpointCorrectionFixture(t)
	if _, e := synLoad(m); e != nil {
		t.Fatal("paired endpoint repair exact predecessor", e)
	}
	for _, tc := range []struct {
		name, want string
		mutate     func(*Metadata)
	}{
		{"missing_pair", "correction_relation_pair", func(m *Metadata) { m.Correction.Deltas = m.Correction.Deltas[:1] }},
		{"old_id_ordinal_mismatch", "correction_relation_endpoints", func(m *Metadata) { m.Correction.Deltas[0].Old = json.RawMessage(`["invalid","also_invalid"]`) }},
		{"old_cardinality_one", "correction_relation_endpoints", func(m *Metadata) {
			b, _ := json.Marshal([]string{groupID(1)})
			m.Correction.Deltas[0].Old = b
			m.Correction.Deltas[1].Old = json.RawMessage(`[1]`)
		}},
		{"old_future_namespace", "correction_relation_endpoints", func(m *Metadata) {
			b, _ := json.Marshal([]string{groupID(1), groupID(121)})
			m.Correction.Deltas[0].Old = b
			m.Correction.Deltas[1].Old = json.RawMessage(`[1,121]`)
		}},
		{"missing_declared_universe", "correction_relation_universe", func(m *Metadata) { m.Universe = nil }},
		{"canonical_index_alias", "correction_pointer", func(m *Metadata) { m.Correction.Deltas[0].Pointer = "/across_group_relation_hypotheses/00/group_ids" }},
		{"element_pointer", "correction_pointer", func(m *Metadata) { m.Correction.Deltas[0].Pointer += "/0" }},
		{"status_pointer", "correction_field", func(m *Metadata) {
			m.Correction.Deltas[0].Pointer = "/across_group_relation_hypotheses/0/review_status"
		}},
		{"bad_prior_pin", "correction_prior_pin", func(m *Metadata) { m.Correction.PriorSHA = strings.Repeat("0", 64) }},
		{"changed_rationale_outside_deltas", "correction_prior_pin", func(m *Metadata) { m.Relations[0].Rationale = "synthetic altered rationale" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := endpointCorrectionFixture(t)
			tc.mutate(&m)
			_, e := synLoad(m)
			if e == nil || e.Error() != tc.want {
				t.Fatalf("wanted %s; got %v", tc.want, e)
			}
		})
	}
}

func ownedBatch26Fixture(t *testing.T) Metadata {
	t.Helper()
	all, e := synTranches(3)
	if e != nil {
		t.Fatal(e)
	}
	m := all[2].meta
	second := m.Relations[0]
	second.IDs = []string{groupID(2), groupID(82)}
	second.Ordinals = []int{2, 82}
	m.Relations = append(m.Relations, second)
	n, e := tree(encode(m))
	if e != nil {
		t.Fatal(e)
	}
	prior := append(compact(n, nil, nil), '\n')
	deltas := []Delta{}
	for i := 0; i < 2; i++ {
		oldIDs, _ := json.Marshal(m.Relations[i].IDs)
		oldOrd, _ := json.Marshal(m.Relations[i].Ordinals)
		m.Relations[i].IDs = []string{groupID(i + 3), groupID(i + 81)}
		m.Relations[i].Ordinals = []int{i + 3, i + 81}
		newIDs, _ := json.Marshal(m.Relations[i].IDs)
		newOrd, _ := json.Marshal(m.Relations[i].Ordinals)
		root := fmt.Sprintf("/across_group_relation_hypotheses/%d/", i)
		deltas = append(deltas, Delta{root + "group_ids", oldIDs, newIDs}, Delta{root + "registration_ordinals", oldOrd, newOrd})
	}
	for i := 0; i < 22; i++ {
		g, f := i/3, i%3
		old, _ := json.Marshal(m.Groups[g].Families[f].EN)
		m.Groups[g].Families[f].EN = fmt.Sprintf("SYNTHETIC_CORRECTED_%d", i)
		neu, _ := json.Marshal(m.Groups[g].Families[f].EN)
		deltas = append(deltas, Delta{fmt.Sprintf("/groups/%d/families/%d/concept_en", g, f), old, neu})
	}
	m.Correction = &Correction{Batch: 1, Created: "2026-10-07T00:00:00Z", Trigger: "synthetic batch26", Scope: "synthetic metadata only", PriorSHA: hash(prior), PriorBytes: len(prior), Serialization: correctionSerialization, Deltas: deltas, Rows: 0}
	return m
}
func TestOwnedBatch26TwoEndpointRepairsReconstruct(t *testing.T) {
	m := ownedBatch26Fixture(t)
	if len(m.Correction.Deltas) != 26 {
		t.Fatal("fixture must retain all26 deltas")
	}
	if _, e := synLoad(m); e != nil {
		t.Fatal("all26 fields reconstruct exact predecessor", e)
	}
	m.Correction.Deltas = m.Correction.Deltas[:25]
	if _, e := synLoad(m); e == nil || e.Error() != "correction_prior_pin" {
		t.Fatal("omitted delta must invalidate predecessor reconstruction", e)
	}
}

func TestEndpointCorrectionPreservesCurrentTypedAndRaw(t *testing.T) {
	m := endpointCorrectionFixture(t)
	source := encode(m)
	before := append([]byte(nil), source...)
	decoded, e := decodePinned(source, hash(source))
	if e != nil {
		t.Fatal("exact predecessor reconstruction", e)
	}
	wantIDs, wantOrd := []string{groupID(2), groupID(81)}, []int{2, 81}
	if !reflect.DeepEqual(decoded.Relations[0].IDs, wantIDs) || !reflect.DeepEqual(decoded.Relations[0].Ordinals, wantOrd) {
		t.Fatal("returned metadata must retain NEW endpoint arrays", decoded.Relations[0].IDs, decoded.Relations[0].Ordinals)
	}
	if !reflect.DeepEqual(source, before) {
		t.Fatal("pinned raw input bytes changed")
	}
	n, e := tree(source)
	if e != nil {
		t.Fatal(e)
	}
	if e = correction(m, n); e != nil {
		t.Fatal("direct reconstruction", e)
	}
	if !reflect.DeepEqual(m.Relations[0].IDs, wantIDs) || !reflect.DeepEqual(m.Relations[0].Ordinals, wantOrd) {
		t.Fatal("predecessor decode aliased current source arrays")
	}
	if !reflect.DeepEqual(encode(m), before) || !reflect.DeepEqual(source, before) {
		t.Fatal("current source serialization or raw bytes changed")
	}
}

func TestExplicitNullNoPriorExactRecipeBatch26(t *testing.T) {
	m := ownedBatch26Fixture(t)
	m.Correction.Serialization = correctionNoPriorSerialization
	wire := func(m Metadata) []byte {
		return []byte(strings.Replace(string(encode(m)), `"field_deltas":`, `"prior_correction_object": null, "field_deltas":`, 1))
	}
	b := wire(m)
	before := append([]byte(nil), b...)
	decoded, e := decodePinned(b, hash(b))
	if e != nil {
		t.Fatal("explicit null plus exact no-prior recipe must reconstruct", e)
	}
	if decoded.Correction.Prior != nil || !reflect.DeepEqual(decoded.Relations, m.Relations) || !reflect.DeepEqual(b, before) {
		t.Fatal("absence/current endpoints/raw bytes changed")
	}
	m.Correction.PriorSHA = strings.Repeat("0", 64)
	bad := wire(m)
	if _, e = decodePinned(bad, hash(bad)); e == nil || e.Error() != "correction_prior_pin" {
		t.Fatal("wrong predecessor pin must reject", e)
	}
	m = ownedBatch26Fixture(t)
	m.Correction.Serialization = correctionNoPriorSerialization + " "
	bad = wire(m)
	if _, e = decodePinned(bad, hash(bad)); e == nil || e.Error() != "correction_provenance" {
		t.Fatal("other recipe text must reject", e)
	}
	m = ownedBatch26Fixture(t)
	m.Correction.Serialization = correctionNoPriorSerialization
	bad = []byte(strings.Replace(string(wire(m)), `"prior_correction_object": null`, `"semantic_reason": null`, 1))
	if _, e = decodePinned(bad, hash(bad)); e == nil || e.Error() != "schema_type" {
		t.Fatal("other typed null must reject", e)
	}
}
