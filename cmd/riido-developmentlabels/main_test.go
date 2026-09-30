package main

import (
	"github.com/teamswyg/laya-tools/internal/filelabels"
	"github.com/teamswyg/laya-tools/internal/sweaudit"
	"testing"
)

func TestSummaryKeepsUnusableTargets(t *testing.T) {
	var members []sweaudit.TaskRole
	for _, id := range []string{"a", "b", "c", "d"} {
		m := sweaudit.TaskRole{Role: "train"}
		m.Task.ID = id
		m.Task.Repository = "example/repo"
		members = append(members, m)
	}
	labels := []filelabels.DevelopmentLabel{
		{Label: filelabels.Label{ID: "a", Result: filelabels.Result{OldPaths: []string{"a", "b"}}}},
		{Label: filelabels.Label{ID: "b", ParseError: true}, Oversized: true},
		{Label: filelabels.Label{ID: "c", Result: filelabels.Result{NewFiles: 1}}},
		{Label: filelabels.Label{ID: "d", Result: filelabels.Result{OldPaths: []string{"d"}, UnsupportedBlocks: 2}}},
	}
	out, e := summarize(members, labels)
	if e != nil {
		t.Fatal(e)
	}
	c := out[0]
	if c.Tasks != 4 || c.OversizedTasks != 1 || c.ParseErrors != 1 || c.NoOldPaths != 1 || c.UnsupportedTasks != 1 || c.UnsupportedBlocks != 2 || c.OldPathReferences != 3 || c.MaxOldPaths != 2 {
		t.Fatalf("bad summary %+v", c)
	}
	members[0].Role = "final"
	if _, e := summarize(members, labels); e == nil {
		t.Fatal("accepted final")
	}
}
