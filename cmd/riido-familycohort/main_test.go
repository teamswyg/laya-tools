package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/teamswyg/laya-tools/internal/familycohort"
	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func TestHelpAndSanitizedArguments(t *testing.T) {
	var out bytes.Buffer
	if run([]string{"--help"}, &out) != 0 || out.Len() == 0 {
		t.Fatal("help unavailable")
	}
	for _, args := range [][]string{{"--private-secret-option"}, {"private-secret-positional"}, {"--root", "private-secret-root"}} {
		out.Reset()
		if run(args, &out) != 2 || bytes.Contains(out.Bytes(), []byte("private-secret")) {
			t.Fatal("bad argument leaked supplied text", out.String())
		}
	}
}

func cliFakeBundle() familycohort.Bundle {
	file := func(name string) familycohort.File {
		data := []byte("public fake metadata bytes " + name)
		return familycohort.File{Path: name, SHA256: sourcecohort.Hash(data), Bytes: int64(len(data))}
	}
	b := familycohort.Bundle{Schema: familycohort.Schema, FrameSchema: file("unopened-fake-schema.json"), FrameSchemaVersion: "fake-version-v1"}
	frame, review := file("unopened-fake-frames.json"), file("unopened-fake-reviews.json")
	for i := 0; i < familycohort.SourceCount; i++ {
		s := familycohort.SourceBinding{ID: fmt.Sprintf("fake-source-%04d", i), Split: "train", Source: file(fmt.Sprintf("unopened-sources/%04d.txt", i)), SourceReview: review}
		b.Sources = append(b.Sources, s)
		for j := 0; j < 3; j++ {
			f := familycohort.FamilyBinding{ID: fmt.Sprintf("fake-family-%04d", i*3+j), SourceID: s.ID, Frame: frame, FrameSchema: b.FrameSchema}
			b.Families = append(b.Families, f)
			for _, locale := range []string{"ko", "en"} {
				b.Comments = append(b.Comments, familycohort.CompleteComment{ID: f.ID + "-" + locale, FamilyID: f.ID, SourceID: s.ID, Source: s.Source, Frame: frame, Locale: locale, Text: "public fake complete UTF-8 fixture 한글 " + f.ID + " " + locale})
			}
		}
	}
	return b
}

func TestOneFullFakeCLICaptureRetainsExactBytesAndNoReferenceReads(t *testing.T) {
	root := t.TempDir()
	bundle := cliFakeBundle()
	raw, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	raw = append(append([]byte(" \n"), raw...), '\n', ' ')
	inputPath := filepath.Join(root, "bundle.json")
	if err := os.WriteFile(inputPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(root, "out")
	args := []string{"--root", root, "--input", "bundle.json", "--sha256", sourcecohort.Hash(raw), "--bytes", fmt.Sprint(len(raw)), "--out", outDir}
	var stdout bytes.Buffer
	if code := run(args, &stdout); code != 0 {
		t.Fatalf("full fake CLI failed: %d %s", code, stdout.String())
	}
	if bytes.Contains(stdout.Bytes(), []byte("fake-")) || bytes.Contains(stdout.Bytes(), []byte(root)) || bytes.Contains(stdout.Bytes(), []byte("한글")) {
		t.Fatal("stdout leaked identifiers, text or local paths")
	}
	var summary familycohort.Summary
	if json.Unmarshal(stdout.Bytes(), &summary) != nil || summary.State != "STRUCTURAL_ONLY_QA_PENDING" || summary.Sources != 400 || summary.Families != 1200 || summary.Comments != 2400 || summary.InputBytes != int64(len(raw)) || summary.MeaningProven || summary.TrainingEligible {
		t.Fatal("wrong structural aggregate", summary)
	}
	retained, err := os.ReadFile(filepath.Join(outDir, "BUNDLE.private.json"))
	if err != nil || !bytes.Equal(retained, raw) {
		t.Fatal("captured input spelling or complete text changed", err)
	}
	var structure familycohort.Structure
	manifest, err := os.ReadFile(filepath.Join(outDir, "STRUCTURE.private.json"))
	if err != nil || json.Unmarshal(manifest, &structure) != nil || structure.Input.SHA256 != sourcecohort.Hash(raw) || structure.Bundle.SHA256 != structure.Input.SHA256 {
		t.Fatal("private structure lost exact input binding", err)
	}
	receipts, err := os.ReadDir(filepath.Join(outDir, "receipts"))
	if err != nil || len(receipts) != 2 {
		t.Fatal("bundle was not captured exactly once", err)
	}
	start, err := os.ReadFile(filepath.Join(outDir, "receipts", "000001.start.json"))
	if err != nil {
		t.Fatal(err)
	}
	var a struct {
		Actor    string `json:"actor"`
		Expected string `json:"expected_sha256"`
		Limit    int    `json:"read_attempt_limit"`
	}
	result, err := os.ReadFile(filepath.Join(outDir, "receipts", "000001.start.json.result"))
	var r struct {
		State    string `json:"state"`
		StartSHA string `json:"start_receipt_sha256"`
		Actual   string `json:"actual_sha256"`
		Bytes    int64  `json:"bytes_read"`
		Attempts int    `json:"input_open_attempts"`
	}
	if err != nil || json.Unmarshal(start, &a) != nil || json.Unmarshal(result, &r) != nil || a.Actor != "familycohort-adapter" || a.Expected != sourcecohort.Hash(raw) || a.Limit != 1 || r.State != "verified" || r.StartSHA != sourcecohort.Hash(start) || r.Actual != a.Expected || r.Bytes != int64(len(raw)) || r.Attempts != 1 {
		t.Fatal("read integrity receipt linkage lost", err)
	}
	// None of the declared Source/review/frame/schema files exist. Only the one
	// bundle and outputs were needed for this software structural check.
	if _, err := os.Stat(filepath.Join(root, "unopened-sources")); !os.IsNotExist(err) {
		t.Fatal("test unexpectedly realized Source files")
	}
	if runtime.GOOS != "windows" {
		if err := filepath.WalkDir(outDir, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			want := os.FileMode(0600)
			if d.IsDir() {
				want = 0700
			}
			if info.Mode().Perm() != want {
				t.Errorf("private permission mismatch %o", info.Mode().Perm())
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	stdout.Reset()
	if run(args, &stdout) != 1 {
		t.Fatal("existing output was overwritten")
	}
	after, _ := os.ReadFile(filepath.Join(outDir, "BUNDLE.private.json"))
	if !bytes.Equal(after, retained) {
		t.Fatal("exclusive output modified old captured bundle")
	}
}

func TestSmallCaptureFailuresDoNotExportStructure(t *testing.T) {
	root := t.TempDir()
	raw := []byte(`{}`)
	if err := os.WriteFile(filepath.Join(root, "small.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, sha string
		bytes     int
	}{
		{"wrong-hash", sourcecohort.Hash([]byte("different fake bytes")), len(raw)},
		{"wrong-byte-count", sourcecohort.Hash(raw), len(raw) + 1},
		{"invalid-shape", sourcecohort.Hash(raw), len(raw)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := filepath.Join(root, tc.name)
			var out bytes.Buffer
			if run([]string{"--root", root, "--input", "small.json", "--sha256", tc.sha, "--bytes", fmt.Sprint(tc.bytes), "--out", dir}, &out) != 1 {
				t.Fatal("invalid bundle capture accepted")
			}
			if _, err := os.Stat(filepath.Join(dir, "STRUCTURE.private.json")); !os.IsNotExist(err) {
				t.Fatal("failed capture exported structure")
			}
			if _, err := os.Stat(filepath.Join(dir, "BUNDLE.private.json")); !os.IsNotExist(err) {
				t.Fatal("failed capture exported bundle")
			}
			if bytes.Contains(out.Bytes(), []byte(root)) || bytes.Contains(out.Bytes(), []byte("small.json")) {
				t.Fatal("capture failure leaked input path")
			}
		})
	}
}
