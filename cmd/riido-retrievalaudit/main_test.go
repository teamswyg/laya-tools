package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTransitiveLeakageGroups(t *testing.T) {
	r := audit([]row{
		{Repository: "a", Query: "shared query", Code: "func A() {}"},
		{Repository: "b", Query: " Shared   QUERY ", Code: "func B() {}"},
		{Repository: "b", Query: "different", Code: "func C() {}"},
		{Repository: "c", Query: "other", Code: "func C() { /* layout only */ }"},
		{Repository: "d", Query: "independent", Code: "func D() {}"},
	})
	if len(r.Groups) != 2 || r.Groups[0] != (group{4, 3}) || r.Groups[1] != (group{1, 1}) {
		t.Fatalf("groups: %+v", r.Groups)
	}
}

func TestCodeTokenIdentity(t *testing.T) {
	a, ok := codeKey("func A() { println(\"a  b\") }")
	if !ok {
		t.Fatal("valid code")
	}
	b, _ := codeKey("func A() { println(\"a b\") }")
	if a == b {
		t.Fatal("collapsed string literal")
	}
	c, _ := codeKey("func A() { /* not code */ println(\"a  b\") }")
	if a != c {
		t.Fatal("comment changed code identity")
	}
	if _, ok := codeKey("func A() { \"unterminated"); ok {
		t.Fatal("bad code accepted")
	}
}

func TestNameWordBoundary(t *testing.T) {
	r := audit([]row{
		{Repository: "a", Name: "Get", Query: "get the item", Code: "func Get() {}"},
		{Repository: "a", Name: "Get", Query: "target item", Code: "func Get() {}"},
	})
	if r.NameInQuery != 1 {
		t.Fatalf("name count %d", r.NameInQuery)
	}
}

func TestRejectMalformedProjection(t *testing.T) {
	for _, data := range []string{"", "{}", `{"repository":"r","path":"x","code":"func A() {}","extra":1}`, `{"repository":"r","path":"x","code":"func A() {}"} {}`} {
		p := filepath.Join(t.TempDir(), "rows.jsonl")
		if err := os.WriteFile(p, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if err := run(p); err == nil {
			t.Fatal("invalid projection accepted")
		}
	}
}
