package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHelpAndSpecNeverExecute(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--task", "catalog-min-context", "--spec"}} {
		var out, err bytes.Buffer
		if code := execute(args, &out, &err); code != 0 || err.Len() != 0 || out.Len() == 0 {
			t.Fatal("static command failed")
		}
		if len(args) > 1 {
			var spec map[string]any
			if json.Unmarshal(out.Bytes(), &spec) != nil || spec["id"] != "catalog-min-context" {
				t.Fatal("wrong static spec")
			}
		}
	}
}

func TestDefaultRequiresExplicitExecution(t *testing.T) {
	for _, args := range [][]string{{"--task", "catalog-min-context"}, {"--execute", "--task", "catalog-min-context"}, {"--execute", "--task", "unknown"}, {"--execute", "--task", "catalog-min-context", "--secret-field", "private-value"}} {
		var out, err bytes.Buffer
		code := execute(args, &out, &err)
		if code != 2 || out.Len() != 0 || err.Len() == 0 || strings.Contains(err.String(), "private-value") {
			t.Fatal("prelaunch refusal/redaction")
		}
	}
}

func TestParseRequiresOnlyTaskForSpec(t *testing.T) {
	c, e := parse([]string{"--task", "comment-preview-authority", "--spec"})
	if e != nil || !c.spec || c.request.Execute {
		t.Fatal("spec required execution")
	}
}

func TestVersionTwoSpecExposesRecipeWithoutLaunching(t *testing.T) {
	for _, id := range []string{"go55-humanize-ordinal64-v2", "go55-uuid-canonical-parse-v2"} {
		var out, diagnostics bytes.Buffer
		if code := execute([]string{"--task", id, "--spec"}, &out, &diagnostics); code != 0 || diagnostics.Len() != 0 {
			t.Fatal("static v2 spec unavailable")
		}
		var spec struct {
			LogicalTaskID string `json:"logical_task_id"`
			RecipeSHA256  string `json:"evaluation_recipe_sha256"`
			Prompt        string `json:"prompt"`
		}
		if json.Unmarshal(out.Bytes(), &spec) != nil || spec.LogicalTaskID == "" || len(spec.RecipeSHA256) != 64 || !strings.Contains(spec.Prompt, "Only ") {
			t.Fatal("public v2 evaluation conditions hidden")
		}
	}
}

func TestBudgetInspectionNeverRequiresExecution(t *testing.T) {
	c, err := parse([]string{"--budget-status", "--parent-file", "/public/parent.json", "--parent-sha256", strings.Repeat("a", 64), "--budget-dir", "/private/missing-ledger"})
	if err != nil || !c.budgetStatus || c.request.Execute || c.request.TaskID != "" {
		t.Fatal("inspection required a task or model execution")
	}
	for _, args := range [][]string{
		{"--budget-status"},
		{"--budget-status", "--execute"},
		{"--budget-status", "--spec"},
	} {
		var out, diagnostics bytes.Buffer
		if code := execute(args, &out, &diagnostics); code != 2 || out.Len() != 0 || diagnostics.Len() == 0 {
			t.Fatal("invalid read-only inspection was accepted")
		}
	}
}
