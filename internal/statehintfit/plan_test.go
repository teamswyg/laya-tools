// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package statehintfit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

// Pins and manifest fixtures describe parser boundaries only. They are not
// invented training sentences, learned predictions or independent cases.
func planFixture() Plan {
	sha := strings.Repeat("a", 64)
	p := Plan{Schema: "riido-statehint-v4-ready-plan-v1", SourceCommit: strings.Repeat("a", 40), BinarySHA: sha, Sources: []Pin{{"go.mod", sha}}, Rubric: Pin{"rubric.json", rubricSHA}, Definitions: [2]Pin{{"rubric.ko.md", rubricKO}, {"rubric.en.md", rubricEN}}, Manifest: Pin{"manifest.json", sha}, Audit: Pin{"audit.json", sha}, Parent: Pin{"parent.rsh", parentSHA}, Seed: 1729, Batch: 32, Epochs: 40, Decay: .001, Confidence: .9, Margin: .05, NLLFloor: 1e-15, FalseCost: [3]int{1, 10, 3}, Coverage: .2, Support: 5, Lineages: 2, Precision: .98, AbstainCost: .375, Temperature: [2]int{5, 50}}
	p.Arms = [4]Arm{{"warm_lr005", true, .005}, {"warm_lr010", true, .01}, {"warm_lr020", true, .02}, {"fresh_lr020", false, .02}}
	for i, name := range [4]string{"train", "validation", "calibration", "test"} {
		p.Partitions[i] = Pin{name + ".jsonl", sha}
	}
	return p
}

func TestStrictReadyPlanRejectsAmbiguousOrIncompleteWire(t *testing.T) {
	p := planFixture()
	data, err := json.Marshal(p)
	var decoded Plan
	if err != nil || strictJSON(data, &decoded) != nil || validPlan(decoded) != nil {
		t.Fatal("valid parser fixture rejected")
	}
	bad := []string{
		strings.Replace(string(data), `"schema":`, `"schema":"duplicate","schema":`, 1),
		strings.Replace(string(data), `"schema":`, `"Schema":`, 1),
		strings.Replace(string(data), `"schema":`, `"extra":false,"schema":`, 1),
		strings.Replace(string(data), `"seed":1729`, `"seed":"1729"`, 1),
		strings.Replace(string(data), `"batch_size":32`, `"batch_size":32.0`, 1),
		strings.Replace(string(data), `"sources":[`, `"sources":null,"discarded":[`, 1),
		strings.Replace(string(data), `"confidence":0.9,`, "", 1),
		strings.Replace(string(data), `"warm":true`, `"Warm":true`, 1),
		strings.Replace(string(data), `"warm":true`, `"warm":true,"warm":false`, 1),
		strings.Replace(string(data), `"learning_rate":0.005`, `"learning_rate":null`, 1),
		strings.Replace(string(data), `"path":"go.mod"`, `"path":"go.mod","path":"go.sum"`, 1),
		string(data) + "{}", "null", "[]", "\xff",
	}
	for i, input := range bad {
		var target Plan
		if strictJSON([]byte(input), &target) == nil {
			t.Fatalf("ambiguous wire %d accepted", i)
		}
	}
	for _, edit := range []func(*Plan){
		func(p *Plan) { p.Confidence = .8 }, func(p *Plan) { p.Margin = 0 },
		func(p *Plan) { p.Lineages = 1 }, func(p *Plan) { p.Precision = .9 },
		func(p *Plan) { p.Arms[0].Rate = .006 }, func(p *Plan) { p.Epochs = 4 },
		func(p *Plan) { p.Manifest.SHA256 = "" }, func(p *Plan) { p.Parent.SHA256 = strings.Repeat("a", 64) },
		func(p *Plan) { p.Partitions[3].Path = p.Audit.Path }, func(p *Plan) { p.Partitions[3].Path = "../outside" },
	} {
		p := planFixture()
		edit(&p)
		if validPlan(p) != ErrStudy {
			t.Fatal("fixed configuration or pin altered")
		}
	}
}

func manifestFixture() Manifest {
	m := Manifest{Schema: "riido-statehint-v4-family-manifest-v1", RubricSHA: rubricSHA}
	for class, intent := range statehint.Intents() {
		for i := range 150 {
			part := "train"
			if i >= 105 {
				part = [3]string{"validation", "calibration", "test"}[(i-105)/15]
			}
			id := fmt.Sprintf("arithmetic-%d-%03d", class, i)
			m.Families = append(m.Families, FamilyPin{ID: id, Lineage: id, Partition: part, Expected: intent, RowIDs: [2]string{id + "-ko", id + "-en"}, TextSHA: [2]string{digest([]byte(id + "ko")), digest([]byte(id + "en"))}})
		}
	}
	return m
}

func TestManifestGlobalQuotaAndLeakageBeforeAnyTextRead(t *testing.T) {
	m := manifestFixture()
	metadata, err := validateManifest(m)
	if err != nil || len(metadata) != 1200 {
		t.Fatal("synthetic metadata fixture")
	}
	for _, edit := range []func(*Manifest){
		func(m *Manifest) { m.Families[1199].Partition = "train" },
		func(m *Manifest) { m.Families[105].Lineage = m.Families[0].Lineage },
		func(m *Manifest) { m.Families[105].RowIDs[0] = m.Families[0].RowIDs[0] },
		func(m *Manifest) { m.Families[105].TextSHA[0] = m.Families[0].TextSHA[0] },
		func(m *Manifest) { m.Families[105].ID = m.Families[0].ID },
		func(m *Manifest) { m.Families[105].Expected = statehint.Progress },
		func(m *Manifest) { m.Families = m.Families[:1199] },
	} {
		m := manifestFixture()
		edit(&m)
		if _, err := validateManifest(m); err != ErrStudy {
			t.Fatal("broken quota/lineage/global uniqueness accepted")
		}
	}
}
