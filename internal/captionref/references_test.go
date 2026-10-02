// SPDX-License-Identifier: Apache-2.0
package captionref

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// These fixtures exercise parser boundaries without loading or executing the
// stored72 corpus. Official generation and later CI replay have separate records.
func typedFixture() string {
	return "package owned\r\n" +
		"//line fictional.go:900\r\n" +
		"var atomicVectors = []atomicVector{{\"눈\"}}\r\n" +
		"var identityVectors = []identityVector{}\r\n" +
		"var snapshotVectors = []snapshotVector{}\r\n" +
		"func local() { var atomicVectors = 3; _ = atomicVectors }\r\n"
}

func legacyFixture() string {
	var out strings.Builder
	out.WriteString("package owned\nvar contracts = []contract{\n")
	for i := range 12 {
		fmt.Fprintf(&out, "{\"p%02d\", \"core\", []vector{{[]int{1}, []int{2}}}},\n", i)
	}
	out.WriteString("}\n")
	return out.String()
}

func TestReferencesPreservePhysicalSourceSpans(t *testing.T) {
	raw := []byte(typedFixture())
	refs, err := literalReferences("owned.go", raw, 1)
	if err != nil || len(refs) != 3 {
		t.Fatalf("references: %v; length %d", err, len(refs))
	}
	for i, ref := range refs {
		s := ref.span
		if s.StartLine != 3+i || s.EndLine != 3+i || s.StartByte < 0 || s.EndByte > len(raw) || s.StartByte >= s.EndByte {
			t.Fatalf("physical span: %+v", s)
		}
		if s.RawSHA256 != digest(raw[s.StartByte:s.EndByte]) {
			t.Fatal("raw span hash changed")
		}
		if !strings.HasPrefix(string(raw[s.StartByte:s.EndByte]), "[]") {
			t.Fatal("span expanded beyond initializer")
		}
	}
	if refs[0].prototype != "atomic-commit" || refs[0].span.VectorExpressions != 1 || refs[1].span.VectorExpressions != 0 {
		t.Fatal("target order/count changed")
	}
}

func TestLegacyElementSpansAndUniqueIdentities(t *testing.T) {
	raw := []byte(legacyFixture())
	refs, err := literalReferences("legacy.go", raw, 0)
	if err != nil || len(refs) != 12 {
		t.Fatalf("legacy refs: %v/%d", err, len(refs))
	}
	for i, ref := range refs {
		s := ref.span
		if ref.prototype != fmt.Sprintf("p%02d", i) || s.Kind != "legacy_contract_element" || s.VectorExpressions != 1 || s.StartLine != i+3 {
			t.Fatalf("legacy position: %+v", ref)
		}
		if string(raw[s.StartByte:s.EndByte])[0] != '{' || s.RawSHA256 != digest(raw[s.StartByte:s.EndByte]) {
			t.Fatal("element raw bytes changed")
		}
	}
	duplicate := strings.Replace(legacyFixture(), "p01", "p00", 1)
	if _, err := literalReferences("legacy.go", []byte(duplicate), 0); err != ErrTarget {
		t.Fatalf("duplicate identity: %v", err)
	}
}

func TestClosedTopLevelTargetShapes(t *testing.T) {
	cases := []struct {
		name, raw string
		source    int
		want      error
	}{
		{"missing", strings.Replace(typedFixture(), "var atomicVectors", "var renamed", 1), 1, ErrTarget},
		{"duplicate", typedFixture() + "var atomicVectors = []atomicVector{}\n", 1, ErrTarget},
		{"callable", strings.Replace(typedFixture(), "[]atomicVector{{\"눈\"}}", "compute()", 1), 1, ErrTarget},
		{"alias", strings.Replace(typedFixture(), "[]atomicVector{{\"눈\"}}", "otherTable", 1), 1, ErrTarget},
		{"fixed-array", strings.Replace(typedFixture(), "[]atomicVector", "[1]atomicVector", 1), 1, ErrTarget},
		{"wrong-element-type", strings.Replace(typedFixture(), "[]atomicVector", "[]otherVector", 1), 1, ErrTarget},
		{"multiple-declaration", strings.Replace(typedFixture(), "atomicVectors =", "atomicVectors, other =", 1), 1, ErrTarget},
		{"syntax", "package owned\nvar contracts = {", 0, ErrAST},
		{"legacy-short", strings.Replace(legacyFixture(), "{\"p11\", \"core\", []vector{{[]int{1}, []int{2}}}},\n", "", 1), 0, ErrTarget},
		{"legacy-not-string", strings.Replace(legacyFixture(), "\"p00\"", "123", 1), 0, ErrTarget},
		{"legacy-wrong-table-type", strings.Replace(legacyFixture(), "[]vector", "[]other", 1), 0, ErrTarget},
		{"invalid-source", typedFixture(), 3, ErrTarget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := literalReferences("owned.go", []byte(tc.raw), tc.source); err != tc.want {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
	flow := "package owned\nvar lifecycleVectors = [...]lifecycleVector{}\nvar quotedVectors = [...]quotedVector{}\nvar graphVectors = [...]graphVector{}\n"
	if refs, err := literalReferences("flow.go", []byte(flow), 2); err != nil || len(refs) != 3 {
		t.Fatalf("inferred arrays: %v/%d", err, len(refs))
	}
	if _, err := literalReferences("flow.go", []byte(strings.Replace(flow, "[...]lifecycleVector", "[]lifecycleVector", 1)), 2); err != ErrTarget {
		t.Fatalf("flow shape drift: %v", err)
	}
}

func TestDecodedUTF8ReferenceDiffersFromSourceLiteral(t *testing.T) {
	raw := []byte(`package owned; var value = "\uB208"`)
	fs := token.NewFileSet()
	file, err := parser.ParseFile(fs, "owned.go", raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	node := file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Values[0]
	span, err := expressionSpan(fs, node, "owned.go", "value", "fixture", raw)
	if err != nil {
		t.Fatal(err)
	}
	decoded := textReference("fixture.json", "/parents/0/request", "눈")
	if decoded.Bytes != 3 || decoded.SHA256 != digest([]byte("눈")) || decoded.SHA256 == span.RawSHA256 {
		t.Fatal("raw and decoded hashes conflated")
	}
	otherPosition := textReference("fixture.json", "/parents/0/candidates/1/text", "눈")
	if decoded.Pointer == otherPosition.Pointer || decoded.SHA256 != otherPosition.SHA256 {
		t.Fatal("equal text collapsed its positions")
	}
}

func TestUnpinnedInputsNeverReachBind(t *testing.T) {
	var counts Counters
	if _, err := Generate(Inputs{}, &counts); err != ErrPin || counts != (Counters{}) {
		t.Fatalf("unpinned input: %v/%+v", err, counts)
	}
	if _, err := Generate(Inputs{}, nil); err != ErrCounts {
		t.Fatalf("missing counters: %v", err)
	}
	counts.BindAttempts = 1
	if _, err := Generate(Inputs{}, &counts); err != ErrCounts || counts.BindAttempts != 1 {
		t.Fatalf("counter reuse: %v/%+v", err, counts)
	}
}

func TestProvenanceRecordsCannotMutateCompiledPins(t *testing.T) {
	a := CompiledFiles("internal/captionref")
	b := CompiledFiles("internal/captionref")
	if len(a) != 2 || len(b) != 2 || a[0].Bytes == 0 || a[0].SHA256 != digest(compiledReferences) {
		t.Fatal("compiled source unavailable")
	}
	a[0].SHA256 = "changed"
	if b[0].SHA256 == "changed" || CompiledFiles("internal/captionref")[0].SHA256 == "changed" {
		t.Fatal("metadata aliased")
	}
	pins := InputPins()
	pins[0].SHA256 = "changed"
	if InputPins()[0].SHA256 == "changed" {
		t.Fatal("input pins aliased")
	}
}
