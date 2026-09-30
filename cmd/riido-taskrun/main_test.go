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
