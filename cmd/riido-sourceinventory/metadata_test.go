// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func toyMetadata(t *testing.T) []byte {
	t.Helper()
	m := metadataProjection{Schema: "riido-typed-truth-development-result-v1", Go: "go1.27.1"}
	m.Legacy.Schema = "riido-behavior-probe-audit-v1"
	m.Typed.Schema = "riido-typed-behavior-truth-v2"
	for i := 0; i < 36; i++ {
		m.Legacy.Sources = append(m.Legacy.Sources, storedRoot{ID: fmt.Sprintf("toy-legacy-%02d", i), Prototype: "toy-prototype", Core: "toy-core", Code: sha([]byte("toy-code")), Normalized: sha([]byte("toy-normal"))})
	}
	for i := 0; i < 24; i++ {
		r := storedRoot{ID: fmt.Sprintf("toy-typed-%02d", i), Prototype: "toy-prototype", Core: "toy-core", Function: "toyRoot", Code: sha([]byte("toy-code")), Normalized: sha([]byte("toy-normal")), Bundle: sha([]byte("toy-bundle"))}
		n := 8
		if i < 12 {
			n = 9
		}
		for j := 0; j < n; j++ {
			r.Components = append(r.Components, storedComponent{ID: fmt.Sprintf("typedbehavior/state.go:value:toy%02d", j), Kind: "value", SHA256: sha([]byte("toy-component"))})
		}
		m.Typed.Sources = append(m.Typed.Sources, r)
	}
	raw, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	// Values exceeding int64 and not exactly representable as float64 are
	// retained only as skipped historical JSON, never arithmetic or features.
	raw = append([]byte(`{"unused_largest_uint64":18446744073709551615,"unused_vector":{"Input":9007199254740993},`), raw[1:]...)
	return raw
}

func TestMetadataProjectsOnlyStoredStringsInOriginalOrder(t *testing.T) {
	raw := toyMetadata(t)
	roots, e := historicalRoots(raw)
	if e != nil {
		t.Fatal(e)
	}
	if len(roots) != 60 || roots[0].ID != "toy-legacy-00" || roots[35].ID != "toy-legacy-35" || roots[36].ID != "toy-typed-00" || roots[59].ID != "toy-typed-23" {
		t.Fatal("metadata reordered")
	}
	count := 0
	for _, r := range roots {
		count += len(r.Components)
	}
	if count != 204 {
		t.Fatal("relation count")
	}
	if roots[0].Function != "" || roots[36].Function != "toyRoot" || roots[36].Components[0].ID != "typedbehavior/state.go:value:toy00" {
		t.Fatal("stored fields missing")
	}
	// Projection owns strings/slices after input bytes are destroyed.
	for i := range raw {
		raw[i] = 0
	}
	if roots[0].ID != "toy-legacy-00" {
		t.Fatal("byte alias")
	}
}

func TestMetadataRejectsRequiredStringShapeCountDuplicates(t *testing.T) {
	raw := toyMetadata(t)
	tests := []struct{ name, old, new string }{
		{"wrong_schema", "riido-typed-truth-development-result-v1", "wrong"},
		{"wrong_go", "go1.27.1", "go1.27.2"},
		{"numeric_id", `"id":"toy-legacy-00"`, `"id":18446744073709551615`},
		{"missing_core", `"core_template":"toy-core"`, `"ignored_core":"toy-core"`},
		{"duplicate_id", "toy-legacy-01", "toy-legacy-00"},
		{"wrong_component_order", "typedbehavior/state.go:value:toy00", "typedbehavior/state.go:value:toy99"},
		{"wrong_component_hash", sha([]byte("toy-component")), strings.Repeat("A", 64)},
		{"new_truth_key_duplicate", `"unused_largest_uint64":18446744073709551615`, `"unused_largest_uint64":1,"unused_largest_uint64":18446744073709551615`},
		{"nested_duplicate", `"Input":9007199254740993`, `"Input":1,"Input":9007199254740993`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := []byte(strings.Replace(string(raw), tc.old, tc.new, 1))
			if _, e := historicalRoots(b); e == nil {
				t.Fatal("changed metadata accepted")
			}
		})
	}
	var m metadataProjection
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	m.Legacy.Sources = m.Legacy.Sources[:35]
	b, _ := json.Marshal(m)
	if _, e := historicalRoots(b); e == nil {
		t.Fatal("root count accepted")
	}
	if e := json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	m.Typed.Sources[0].Components = m.Typed.Sources[0].Components[:8]
	b, _ = json.Marshal(m)
	if _, e := historicalRoots(b); e == nil {
		t.Fatal("changed relation cardinality accepted")
	}
}

func TestMetadataJSONBoundsAndTrailingValues(t *testing.T) {
	for _, b := range [][]byte{nil, []byte("{}{}"), []byte("{\"a\":1,}"), []byte{0xff}, []byte(strings.Repeat("[", 65) + "0" + strings.Repeat("]", 65)), []byte(strings.Repeat(" ", maxBytes+1))} {
		if e := uniqueJSON(b); e == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
	if e := uniqueJSON([]byte(`{"u64":18446744073709551615,"unknown":[true,null,"toy"]}`)); e != nil {
		t.Fatal("unneeded large integer rejected", e)
	}
}
