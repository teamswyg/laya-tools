package familycohort

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/teamswyg/laya-tools/internal/sourcecohort"
)

func TestVolumeRootReachesPinnedBundleDecode(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "software-bundle.json")
	data := []byte("{}")
	if err := os.WriteFile(inputPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	root := filepath.VolumeName(inputPath) + string(os.PathSeparator)
	rel, err := filepath.Rel(root, inputPath)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "fresh-output")
	s, err := Run(root, File{Path: filepath.ToSlash(rel), SHA256: sourcecohort.Hash(data), Bytes: int64(len(data))}, out)
	if err == nil || s.Code != "bundle_json" {
		t.Fatalf("valid root descendant was refused before closed bundle decoding: code=%q error=%v", s.Code, err)
	}
	for _, name := range []string{"000001.start.json", "000001.start.json.result"} {
		if _, err := os.Stat(filepath.Join(out, "receipts", name)); err != nil {
			t.Fatal("root descendant did not produce its bounded read receipt", err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "STRUCTURE.private.json")); !os.IsNotExist(err) {
		t.Fatal("invalid software bundle produced a verified structure")
	}
}
