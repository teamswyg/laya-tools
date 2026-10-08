package inventoryprojection

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/semanticframe"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

// Original public software fixtures: no admitted material or acceptance labels.
func raw(name string, b []byte) PinnedBytes {
	return PinnedBytes{File: File{Path: name, SHA256: sourcecohort.Hash(b), Bytes: int64(len(b))}, Bytes: b}
}
func encode(t testing.TB, name string, v any) PinnedBytes {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return raw(name, b)
}
func nativeBytes(t testing.TB, f Full) []byte {
	t.Helper()
	b, e := json.Marshal(f)
	if e != nil {
		t.Fatal(e)
	}
	fields, e := json.Marshal(f.OfficialFields)
	if e != nil {
		t.Fatal(e)
	}
	var obj bytes.Buffer
	obj.WriteByte('{')
	for i, x := range f.OfficialFields {
		if i > 0 {
			obj.WriteByte(',')
		}
		key, _ := json.Marshal(x.Name)
		value, _ := json.Marshal(x.Support)
		obj.Write(key)
		obj.WriteByte(':')
		obj.Write(value)
	}
	obj.WriteByte('}')
	return bytes.Replace(b, append([]byte(`"official_field_support":`), fields...), append([]byte(`"official_field_support":`), obj.Bytes()...), 1)
}
func fixture(t testing.TB) (Inputs, Full) {
	t.Helper()
	source := raw("source.txt", []byte("항목 A; check."))
	boundary := Span{Kind: "source_scope", Start: 0, End: len(source.Bytes)}
	es := []Span{boundary}
	ref := raw("receipt.json", []byte("opaque reference"))
	support := Support{Status: "open-status", Value: nil, Reason: nil, Inspected: false, Description: "full-only support explanation", Propositions: []string{"pA"}, Referents: []string{}, Oppositions: []string{}, Temporal: []string{}, Evidence: es}
	f := Full{Schema: "public-software-fixture-full-v1", ID: "fixture-source", Source: source.File, Creation: ref.File, ReadBinding: ref.File, AdditionalReads: []AdditionalRead{}, Coder: Role{"fixture-coder", "fixture-agent"}, Phase: "open-phase", Scope: "open-scope", Boundary: boundary, Complete: false, Components: []semanticframe.Component{{ID: "c1", Description: "component description", Evidence: es}}, Propositions: []semanticframe.Proposition{{ID: "pZ", Description: "first description", Polarity: "open polarity", Scope: "operation scope", Subjects: []string{"r1"}, Evidence: es}, {ID: "pA", Description: "second description", Polarity: "another open polarity", Scope: "scope", Subjects: []string{}, Evidence: es}}, NegativeIDs: []string{"pA"}, NegationDescription: "retained combined scope", Oppositions: []semanticframe.Opposition{}, Temporal: []semanticframe.Temporal{}, Referents: []semanticframe.Referent{{ID: "r1", Kind: "open referent kind", Parent: nil, Description: "referent description", Evidence: es}}, Constraints: []semanticframe.Constraint{}, OfficialFields: []OfficialField{{Name: "z_field", Support: support}, {Name: "a_field", Support: support}}, Dependencies: []string{}, Unresolved: []string{"retained unresolved question"}}
	in := Inputs{Full: raw("full.json", nativeBytes(t, f)), Native: encode(t, "native.json", NativeContract{Schema: f.Schema, Fields: []string{"a_field", "z_field"}}), Mapping: encode(t, "mapping.json", SupportedMapping()), Adapter: raw("adapter.go", []byte("public declared adapter identity"))}
	c := Config{Full: in.Full.File, Adapter: in.Adapter.File, Version: Version, Mapping: in.Mapping.File, Native: in.Native.File, GraphPath: "graph.json"}
	in.Config = encode(t, "config.json", c)
	return in, f
}
func repin(t testing.TB, in *Inputs, f Full) {
	t.Helper()
	in.Full = raw(in.Full.File.Path, nativeBytes(t, f))
	var c Config
	json.Unmarshal(in.Config.Bytes, &c)
	c.Full = in.Full.File
	in.Config = encode(t, in.Config.File.Path, c)
}

func TestProjectionRetainsFullAnchorOrderAndOmittedFields(t *testing.T) {
	in, _ := fixture(t)
	r, err := Project(in)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r.RetainedFullBytes, in.Full.Bytes) || r.Binding.Full != in.Full.File || r.Binding.Graph.SHA256 != sourcecohort.Hash(r.GraphBytes) {
		t.Fatal("pin/raw loss")
	}
	if r.Graph.Propositions[0].ID != "pZ" || r.Graph.Propositions[1].ID != "pA" || r.Graph.Propositions[0].Polarity != "open polarity" || r.Full.OfficialFields[0].Name != "z_field" || r.Full.OfficialFields[1].Name != "a_field" {
		t.Fatal("reordered or changed values")
	}
	if r.Full.NegationDescription != "retained combined scope" || len(r.Full.Unresolved) != 1 || r.Full.OfficialFields[0].Support.Description != "full-only support explanation" || r.Full.OfficialFields[0].Support.Value != nil {
		t.Fatal("full-only value loss")
	}
	if r.UTF8SpansVerified || r.MeaningProven || r.TrainingEligible {
		t.Fatal("unearned qualification")
	}
	in.Full.Bytes[0] = 'x'
	if r.RetainedFullBytes[0] != '{' {
		t.Fatal("retained bytes alias caller input")
	}
}
func TestSupportOnlyChangeCanKeepGraphButMustChangeFullBinding(t *testing.T) {
	in, f := fixture(t)
	before, err := Project(in)
	if err != nil {
		t.Fatal(err)
	}
	f.OfficialFields[0].Support.Description = "changed full-only explanation"
	repin(t, &in, f)
	after, err := Project(in)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before.GraphBytes, after.GraphBytes) || before.Binding.Full == after.Binding.Full || before.Binding.Config == after.Binding.Config {
		t.Fatal("wrong support-only binding")
	}
}
func TestOptionalSourceOnlyControlsActualUTF8SpanValidation(t *testing.T) {
	in, f := fixture(t)
	source := raw("source.txt", []byte("항목 A; check."))
	in.Source = &source
	r, err := Project(in)
	if err != nil || !r.UTF8SpansVerified {
		t.Fatalf("valid source: %v", err)
	}
	f.OfficialFields[0].Support.Evidence = []Span{{Kind: "proposition", Start: 1, End: 3}}
	repin(t, &in, f)
	if _, err = Project(in); err != Error("utf8_span") {
		t.Fatalf("split codepoint: %v", err)
	}
	in.Source = nil
	if r, err = Project(in); err != nil || r.UTF8SpansVerified {
		t.Fatalf("copy-only is not verified: %v", err)
	}
}
func TestConfigAndInputSubstitutionsReject(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Inputs)
	}{
		{"changed full", func(in *Inputs) { in.Full.Bytes = append(in.Full.Bytes, ' ') }},
		{"repinned full without config", func(in *Inputs) { in.Full = raw("full.json", append(in.Full.Bytes, ' ')) }},
		{"missing mapping", func(in *Inputs) { in.Mapping = PinnedBytes{} }},
		{"repinned native", func(in *Inputs) {
			in.Native = encode(t, "native.json", NativeContract{Schema: "different", Fields: []string{"a_field", "z_field"}})
		}},
		{"repinned adapter", func(in *Inputs) { in.Adapter = raw("adapter.go", []byte("different source identity")) }},
		{"unsupported version", func(in *Inputs) {
			var c Config
			json.Unmarshal(in.Config.Bytes, &c)
			c.Version = "unsupported"
			in.Config = encode(t, "config.json", c)
		}},
		{"wrong supported mapping", func(in *Inputs) {
			m := SupportedMapping()
			m.CopiedFields = m.CopiedFields[:1]
			in.Mapping = encode(t, "mapping.json", m)
			var c Config
			json.Unmarshal(in.Config.Bytes, &c)
			c.Mapping = in.Mapping.File
			in.Config = encode(t, "config.json", c)
		}},
		{"wrong field keyset", func(in *Inputs) {
			var n NativeContract
			json.Unmarshal(in.Native.Bytes, &n)
			n.Fields = []string{"a_field"}
			in.Native = encode(t, "native.json", n)
			var c Config
			json.Unmarshal(in.Config.Bytes, &c)
			c.Native = in.Native.File
			in.Config = encode(t, "config.json", c)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, _ := fixture(t)
			tc.change(&in)
			if _, err := Project(in); err == nil {
				t.Fatal("substitution accepted")
			}
		})
	}
}
func TestFullEndpointsEvidenceAndMetadataReject(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Full)
	}{
		{"dangling full-only negative ID", func(f *Full) { f.NegativeIDs = []string{"missing"} }},
		{"dangling full-only support endpoint", func(f *Full) { f.OfficialFields[0].Support.Temporal = []string{"missing"} }},
		{"dangling subject", func(f *Full) { f.Propositions[0].Subjects = []string{"missing"} }},
		{"parent cycle", func(f *Full) { id := "r1"; f.Referents[0].Parent = &id }},
		{"duplicate proposition", func(f *Full) { f.Propositions[1].ID = "pZ" }},
		{"invalid support span", func(f *Full) { f.OfficialFields[0].Support.Evidence[0].End = int(f.Source.Bytes) + 1 }},
		{"partial full boundary", func(f *Full) { f.Boundary.End-- }},
		{"self dependency", func(f *Full) { f.Dependencies = []string{f.ID} }},
		{"conflicting Source/receipt path", func(f *Full) { f.Creation.Path = f.Source.Path }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, f := fixture(t)
			tc.change(&f)
			repin(t, &in, f)
			if _, err := Project(in); err == nil {
				t.Fatal("bad full graph accepted")
			}
		})
	}
}
func TestClosedNativeDecodeAndEarlyBounds(t *testing.T) {
	in, _ := fixture(t)
	bad := [][]byte{[]byte(`{}`), []byte(strings.Replace(string(in.Full.Bytes), `"schema":`, `"Schema":`, 1)), []byte(strings.Replace(string(in.Full.Bytes), `"schema":`, `"extra":true,"schema":`, 1)), []byte(strings.Replace(string(in.Full.Bytes), `"schema":`, `"schema":"duplicate","schema":`, 1)), []byte(strings.Replace(string(in.Full.Bytes), `"z_field":`, `"z_field":null,"z_field":`, 1)), []byte(strings.Replace(string(in.Full.Bytes), `"open polarity"`, `"\ud800"`, 1)), []byte(strings.Replace(string(in.Full.Bytes), `"required_dependencies":[]`, `"required_dependencies":null`, 1)), []byte(`{"x":` + strings.Repeat(`[`, 33) + `0` + strings.Repeat(`]`, 33) + `}`)}
	for i, b := range bad {
		if _, err := decodeFull(b); err == nil {
			t.Fatalf("invalid native %d accepted", i)
		}
	}
	large := raw("full.json", []byte(strings.Repeat(" ", MaxBytes+1)))
	in.Full = large
	if _, err := Project(in); err != Error("byte_pin") {
		t.Fatalf("full bounds %v", err)
	}
	array := `[` + strings.Repeat(`0,`, MaxItems) + `0]`
	if err := preflight([]byte(array)); err != Error("collection_bounds") {
		t.Fatalf("array bounds %v", err)
	}
	largeHash := strings.Repeat("a", 1<<20)
	if allocations := testing.AllocsPerRun(10, func() {
		if validFile(File{Path: "file", SHA256: largeHash, Bytes: 1}) {
			panic("bad pin")
		}
	}); allocations != 0 {
		t.Fatalf("pin validation allocations %v", allocations)
	}
}

func TestLiteralHTMLExpansionIsRejectedBeforeGraphEncoding(t *testing.T) {
	in, f := fixture(t)
	f.Propositions[0].Description = strings.Repeat("<", MaxBytes/4)
	literal := bytes.ReplaceAll(nativeBytes(t, f), []byte(`\u003c`), []byte(`<`))
	if len(literal) > MaxBytes {
		t.Fatal("fixture input too large")
	}
	in.Full = raw("full.json", literal)
	var c Config
	json.Unmarshal(in.Config.Bytes, &c)
	c.Full = in.Full.File
	in.Config = encode(t, "config.json", c)
	if _, err := Project(in); err != Error("graph_bytes") {
		t.Fatalf("expanding graph: %v", err)
	}
}

func TestUnknownNativeSchemaAndKeysetRemainExplicitConfiguration(t *testing.T) {
	in, f := fixture(t)
	// No historical schema or official field names are hard-coded in the decoder.
	f.Schema = "another-public-native-schema"
	f.OfficialFields[0].Name = "other_z"
	repin(t, &in, f)
	in.Native = encode(t, "native.json", NativeContract{Schema: f.Schema, Fields: []string{"a_field", "other_z"}})
	var c Config
	json.Unmarshal(in.Config.Bytes, &c)
	c.Native = in.Native.File
	in.Config = encode(t, "config.json", c)
	r, err := Project(in)
	if err != nil || r.Binding.Native != in.Native.File {
		t.Fatalf("explicit unknown contract: %v", err)
	}
	if r.MeaningProven || r.TrainingEligible {
		t.Fatal("native contract qualified semantics")
	}
}
