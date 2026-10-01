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

func TestVersionedSpecAndOriginalAttribution(t *testing.T) {
	const task = "repo-keyword-language-guard"
	spec, err := taskverify.TaskSpec(task)
	if err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	if code := execute([]string{"--task", task, "--spec"}, &out, &diagnostics); code != 0 || diagnostics.Len() != 0 {
		t.Fatal("versioned spec CLI failed")
	}
	var got taskverify.Spec
	if json.Unmarshal(out.Bytes(), &got) != nil || got.BaseRevision != spec.BaseRevision || got.DefinitionSHA256 == "" || got.AcceptanceSourceSHA256 == "" {
		t.Fatal("versioned CLI specification lost snapshot or independent contract")
	}
	base := filepath.Join("..", "..", "internal", "taskverify", "testdata", "repo-keyword-language-guard-v1")
	out.Reset()
	diagnostics.Reset()
	if code := execute([]string{"--task", task, "--base-dir", base, "--candidate-dir", base}, &out, &diagnostics); code != 1 || diagnostics.Len() != 0 {
		t.Fatalf("versioned baseline CLI must reject the unchanged candidate using its own revision, code=%d", code)
	}
	var report taskverify.Report
	if json.Unmarshal(out.Bytes(), &report) != nil || report.BaseRevision != spec.BaseRevision || report.Status != "rejected" || report.Accepted || len(report.AttributionFiles) != 2 || report.AttributionSHA256 == "" {
		t.Fatal("CLI did not bind original public attribution and versioned snapshot")
	}
	// A caller-controlled LICENSE with identical size is still rejected by its
	// digest. Only the separately pinned original staging bytes are accepted.
	dir := t.TempDir()
	paths, _ := taskverify.BasePaths(task)
	pins, _ := taskverify.AttributionPins(task)
	for _, pin := range pins {
		paths = append(paths, pin.Path)
	}
	for _, path := range paths {
		b, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if path == "LICENSE" {
			b[0] ^= 1
		}
		if err := os.WriteFile(p, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	out.Reset()
	diagnostics.Reset()
	if code := execute([]string{"--task", task, "--base-dir", dir, "--candidate-dir", base}, &out, &diagnostics); code != 2 || out.Len() != 0 || strings.TrimSpace(diagnostics.String()) != "attribution_pin_mismatch" {
		t.Fatal("CLI accepted or exposed mismatched original attribution")
	}
}

func TestIndependentParserFamilySpecAndBaseline(t *testing.T) {
	const task = "taskoutcome-event-key-bounds"
	var out, diagnostics bytes.Buffer
	if code := execute([]string{"--task", task, "--spec", "--candidate-dir", "/missing-private-marker"}, &out, &diagnostics); code != 0 || diagnostics.Len() != 0 {
		t.Fatal("independent parser spec failed")
	}
	var spec taskverify.Spec
	if json.Unmarshal(out.Bytes(), &spec) != nil || spec.ID != task || spec.BaseRevision != "ee72334166e2962b0821d5198a50fdd13f92ab29" || spec.DefinitionSHA256 != "408255ae28f7d6bfb8994d4a9cbcbcaa46edd95c1cf2f606fe847cd28e0599a9" || strings.Contains(out.String(), "private-marker") {
		t.Fatal("parser CLI did not select its own pinned definition")
	}
	base := filepath.Join("..", "..", "internal", "taskverify", "testdata", "taskoutcome-event-key-bounds-v1")
	out.Reset()
	diagnostics.Reset()
	if code := execute([]string{"--task", task, "--base-dir", base, "--candidate-dir", base}, &out, &diagnostics); code != 1 || diagnostics.Len() != 0 {
		t.Fatal("independent parser baseline was not rejected at its own revision")
	}
	var report taskverify.Report
	if json.Unmarshal(out.Bytes(), &report) != nil || report.TaskID != task || report.BaseRevision != spec.BaseRevision || report.Status != "rejected" || report.Accepted || report.SpecSHA256 != "169cedefd41d3a5ba05d11bbb4685761a0dde6adf07401b846b77bf5a84c27db" || len(report.AttributionFiles) != 2 || report.AttributionSHA256 == "" {
		t.Fatal("parser CLI report lost task/source/contract/attribution identity")
	}
}
