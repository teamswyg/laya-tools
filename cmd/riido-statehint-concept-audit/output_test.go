// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Run the owned fault fixtures under ordinary go test, including pre-content
// budgets, strict decoding, exact references and predecessor reconstruction.
func TestOwnedSyntheticChecks(t *testing.T) {
	for _, name := range []string{"GOMEMLIMIT", "GOTOOLCHAIN", "GOPROXY"} {
		t.Setenv(name, "SYNTHETIC_ENV_SENTINEL")
	}
	r, err := selfTest()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encode(r), []byte("SYNTHETIC_ENV_SENTINEL")) {
		t.Fatal("self-test must not echo environment values")
	}
}

func TestExplicitPinnedSyntheticFileAndOutputBoundary(t *testing.T) {
	dir, err := os.MkdirTemp(".", ".synthetic-pinned-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	m := synthetic(1)
	g := &m.Groups[0]
	g.Core = Pair{"SYNTHETIC_CORE_KO_SENTINEL", "SYNTHETIC_CORE_EN_SENTINEL"}
	g.Domain = "SYNTHETIC_DOMAIN_SENTINEL"
	g.Brevity = "SYNTHETIC_BREVITY_SENTINEL"
	g.Slots[0].KO, g.Slots[0].EN = "SYNTHETIC_SLOT_KO_SENTINEL", "SYNTHETIC_SLOT_EN_SENTINEL"
	g.Families[0].Caution = "SYNTHETIC_CAUTION_SENTINEL"
	m.Relations = []Relation{{[]string{groupID(1), groupID(2)}, []int{1, 2}, "synthetic_review", "SYNTHETIC_RATIONALE_SENTINEL", "SYNTHETIC_DISTINCTION_SENTINEL", "pending_independent_full_registry_lineage_review", []string{"synthetic_dimension"}}}
	m.Sources.Pins = []DeclaredSourcePin{{"synthetic-declared-source.json", strings.Repeat("e", 64), 1}}
	source := encode(m)
	path := filepath.Join(dir, "metadata.json")
	if err = os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	all, err := loadInputs([]string{path + "=" + hash(source)}, prepare, load)
	if err != nil {
		t.Fatal(err)
	}
	r, err := audit(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Inputs) != 1 || r.Inputs[0].Path != path || r.Inputs[0].SHA != hash(source) || r.Inputs[0].Bytes != len(source) || !reflect.DeepEqual(r.Range, []int{1, 40}) {
		t.Fatal("report must bind the explicitly supplied file")
	}
	report := encode(r)
	for _, forbidden := range []string{"SYNTHETIC_", g.ID, g.Families[0].ID, m.Sources.Pins[0].Path, m.Sources.Pins[0].SHA} {
		if bytes.Contains(report, []byte(forbidden)) {
			t.Fatal("report exposed per-record metadata or an unopened source declaration")
		}
	}
	if r.WithinTypes["same_core_event;discourse_variant"] != 80 || r.RelationTypes["synthetic_review"]["pending_independent_full_registry_lineage_review"] != 1 || r.Components.Max != 2 {
		t.Fatal("aggregate type spelling or review connectivity changed")
	}
	onDisk, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(source, onDisk) {
		t.Fatal("audit changed the supplied file", err)
	}
}
