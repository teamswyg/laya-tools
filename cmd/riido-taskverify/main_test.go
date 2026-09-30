package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/taskverify"
)

func TestSpecDoesNotReadOrExecuteCandidate(t *testing.T) {
	var out, diagnostics bytes.Buffer
	code := execute([]string{"--task", "catalog-min-context", "--spec", "--candidate-dir", "/missing-private-marker"}, &out, &diagnostics)
	var spec taskverify.Spec
	if code != 0 || diagnostics.Len() != 0 || json.Unmarshal(out.Bytes(), &spec) != nil || spec.ID != "catalog-min-context" || !strings.Contains(spec.Prompt, "MinContext") {
		t.Fatalf("spec failed: code=%d", code)
	}
	if strings.Contains(out.String(), "private-marker") {
		t.Fatal("candidate path leaked into public specification")
	}
}
func TestInvalidInputsHaveRedactedFixedErrors(t *testing.T) {
	for _, args := range [][]string{{"--task", "private-marker"}, {"--task", "catalog-min-context", "--timeout", "61s"}, {"--secret-private-marker", "a"}, {"--task", "catalog-min-context"}} {
		var out, diagnostics bytes.Buffer
		if execute(args, &out, &diagnostics) != 2 || out.Len() != 0 || strings.Contains(diagnostics.String(), "private-marker") {
			t.Fatal("invalid input accepted or echoed")
		}
	}
}
func TestBaseCannotEscapeThroughSymlink(t *testing.T) {
	dir := t.TempDir()
	if e := os.MkdirAll(filepath.Join(dir, "pkg"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(t.TempDir(), filepath.Join(dir, "pkg/reporouter")); e != nil {
		t.Fatal(e)
	}
	if _, e := readBase(dir, "comment-preview-authority"); e == nil {
		t.Fatal("symlinked base closure was read")
	}
}
