package researchbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	m := Manifest{Schema: "riido-tinyhead-bundle-v1", Repository: "JooYoon/riidolaya-tinyhead-experiment-v0.1", SourceRevision: strings.Repeat("a", 40), ParentRevision: "f648bad968b0e8c2b04cf0596b1c664721060fef", Origin: "original_synthetic_pdca06_development", License: "apache-2.0", Files: map[string]string{}}
	s := tinyhead.Source{Gamma: make([]float64, 1024), Beta: make([]float64, 1024), Weight: make([]float64, 3072), Temperature: .5}
	model, e := tinyhead.Build(s, tinyhead.Float32, 0)
	if e != nil {
		t.Fatal(e)
	}
	files := map[string][]byte{}
	for _, n := range names {
		files[n] = []byte("public research")
	}
	files["LICENSE"], e = os.ReadFile("../../LICENSE")
	if e != nil {
		t.Fatal(e)
	}
	files["features.json"] = []byte(`{"schema":"riido-tiny-features-v1","origin":"original_synthetic_pdca06_development","provenance":{"parent_head_sha256":"29cc0c479fb60fc038ef9555b0e082f1c84555d353ce7f89c40eca989a365627"}}`)
	rows := []map[string]string{}
	for _, n := range names {
		if strings.HasSuffix(n, ".rdh") {
			files[n] = model.Encode()
			h := sha256.Sum256(files[n])
			rows = append(rows, map[string]string{"name": strings.TrimSuffix(n, ".rdh"), "sha256": hex.EncodeToString(h[:])})
		}
	}
	h := sha256.Sum256(files["features.json"])
	files["results.json"], e = json.Marshal(map[string]any{"experiment": "tinyhead-01", "fp32_reference_parity_pass": true, "feature_sha256": hex.EncodeToString(h[:]), "results": rows})
	if e != nil {
		t.Fatal(e)
	}
	for n, b := range files {
		h := sha256.Sum256(b)
		m.Files[n] = hex.EncodeToString(h[:])
		if e = os.WriteFile(filepath.Join(dir, n), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	b, _ := json.Marshal(m)
	if e = os.WriteFile(filepath.Join(dir, "MANIFEST.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	return dir
}
func TestAllowlistIntegrity(t *testing.T) {
	dir := stage(t)
	if _, e := Verify(dir); e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(dir, "extra.env")
	if e := os.WriteFile(p, []byte("unreviewed"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Verify(dir); e == nil {
		t.Fatal("unexpected file accepted")
	}
	os.Remove(p)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed"), 0600)
	if _, e := Verify(dir); e == nil {
		t.Fatal("tamper accepted")
	}
}
func TestRejectSymlink(t *testing.T) {
	dir := stage(t)
	target := filepath.Join(t.TempDir(), "notice")
	os.WriteFile(target, []byte("public research"), 0600)
	os.Remove(filepath.Join(dir, "NOTICE"))
	if e := os.Symlink(target, filepath.Join(dir, "NOTICE")); e != nil {
		t.Skip(e)
	}
	if _, e := Verify(dir); e == nil {
		t.Fatal("symlink accepted")
	}
}
func TestRejectManifestPolicy(t *testing.T) {
	for _, change := range []func(*Manifest){func(m *Manifest) { m.ProductionReady = true }, func(m *Manifest) { m.ParentRevision = "main" }, func(m *Manifest) { m.SourceRevision = "HEAD" }, func(m *Manifest) { m.Repository = "someone/model" }, func(m *Manifest) { m.Files["../outside"] = "x" }} {
		dir := stage(t)
		p := filepath.Join(dir, "MANIFEST.json")
		b, _ := os.ReadFile(p)
		var m Manifest
		json.Unmarshal(b, &m)
		change(&m)
		b, _ = json.Marshal(m)
		os.WriteFile(p, b, 0600)
		if _, e := Verify(dir); e == nil {
			t.Fatal("invalid policy accepted")
		}
	}
}
