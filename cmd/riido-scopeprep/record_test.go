package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/scopedproperty"
)

// This publication guard was added after the single preparation export. It
// replays proposal construction, not candidate truth, ranking or performance.
func TestPublishedPreparation56cMatchesOriginalAndFrozenSources(t *testing.T) {
	repo := repoRoot(t)
	read := func(path string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(repo, path))
		if err != nil {
			t.Fatal("published preparation unavailable")
		}
		return raw
	}
	const base = "experiments/short-claim/"
	planRaw := read(base + "preparation-plan-56c.json")
	if digest(planRaw) != "8a47e276b0693f6a68fe1894dd6b180846478b8ddb0ccfdef71905b87ce23e37" {
		t.Fatal("frozen preparation plan changed")
	}
	type pin struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Bytes  int    `json:"bytes"`
	}
	var plan struct {
		Files []pin `json:"implementation_files"`
		Old   []pin `json:"preserved_56b_artifacts"`
		Docs  []pin `json:"preparation_documents"`
	}
	if json.Unmarshal(planRaw, &plan) != nil || len(plan.Files) != 23 || len(plan.Old) != 27 || len(plan.Docs) != 6 {
		t.Fatal("preparation manifest shape changed")
	}
	for _, pins := range [][]pin{plan.Files, plan.Old, plan.Docs} {
		for _, p := range pins {
			if strings.HasPrefix(p.Path, "/") || strings.Contains(p.Path, "..") {
				t.Fatal("unexpected public pin path")
			}
			raw := read(p.Path)
			if digest(raw) != p.SHA256 || len(raw) != p.Bytes {
				t.Fatal("frozen preparation or historical artifact changed")
			}
		}
	}
	dataset, err := scopedproperty.CreateDataset()
	if err != nil {
		t.Fatal("proposal replay failed")
	}
	want, err := json.MarshalIndent(dataset, "", "  ")
	if err != nil {
		t.Fatal("proposal replay encoding failed")
	}
	want = append(want, '\n')
	probes := read(base + "probes-56c.json")
	if !bytes.Equal(probes, want) || digest(probes) != "8832ed777ecb18c9d250e1fca112030a75260b9b3a901d460732bb8b5520978b" {
		t.Fatal("published proposals diverge from frozen original")
	}
	summary := read(base + "preparation-56c.json")
	if digest(summary) != "46729c04a44c142b2d8a3889b23d524eb4c7cbb4f48a3d7806807a70c2372204" {
		t.Fatal("published preparation summary changed")
	}
	var report preparation
	if json.Unmarshal(summary, &report) != nil {
		t.Fatal("published preparation summary invalid")
	}
	assertPreparationOnly(t, report)
	compiled, err := compiledArtifacts()
	if err != nil || len(compiled) != len(report.CompiledSources) {
		t.Fatal("compiled preparation manifest changed")
	}
	for i, p := range compiled {
		if p != report.CompiledSources[i] {
			t.Fatal("published summary is not bound to compiled sources")
		}
	}
	if report.ProposalsSHA256 != digest(probes) || report.ProposalsBytes != len(probes) {
		t.Fatal("published proposal bytes not bound to summary")
	}
	if digest(read(base+"snapshot-56c.json")) != "e1a7f38d063446269b035cd5cd3f11d7de49a725d6f5dbed271a86ebcdfc7f6c" {
		t.Fatal("published snapshot changed")
	}
}
