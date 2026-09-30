// Package researchbundle verifies an explicit public tinyhead artifact allowlist.
// Integrity checks do not replace review of data rights or publication suitability.
package researchbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/teamswyg/laya-tools/internal/tinyhead"
)

var names = []string{"LICENSE", "NOTICE", "README.md", "features.json", "results.json", "folded-f32.rdh", "row-int8.rdh", "ternary-0.5.rdh", "ternary-0.7.rdh", "ternary-1.rdh", "ternary-1.3.rdh"}
var sha = regexp.MustCompile(`^[0-9a-f]{40}$`)
var repo = regexp.MustCompile(`^JooYoon/riidolaya-tinyhead-experiment-v[0-9]+\.[0-9]+$`)

type Manifest struct {
	Schema          string            `json:"schema"`
	Repository      string            `json:"repository"`
	SourceRevision  string            `json:"source_revision"`
	ParentRevision  string            `json:"parent_revision"`
	Origin          string            `json:"origin"`
	License         string            `json:"license"`
	ProductionReady bool              `json:"production_ready"`
	Files           map[string]string `json:"files"`
}

func Names() []string { return append([]string(nil), names...) }
func read(path string, limit int64) ([]byte, error) {
	i, e := os.Lstat(path)
	if e != nil {
		return nil, e
	}
	if !i.Mode().IsRegular() || i.Size() > limit {
		return nil, fmt.Errorf("non-regular or oversized artifact")
	}
	return os.ReadFile(path)
}
func Verify(dir string) (Manifest, error) {
	var m Manifest
	b, e := read(filepath.Join(dir, "MANIFEST.json"), 16384)
	if e != nil {
		return m, e
	}
	if e = json.Unmarshal(b, &m); e != nil {
		return m, e
	}
	if m.Schema == "riido-qat-bundle-v1" {
		return VerifyQAT(dir, m)
	}
	if m.Schema != "riido-tinyhead-bundle-v1" || !repo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "f648bad968b0e8c2b04cf0596b1c664721060fef" || m.Origin != "original_synthetic_pdca06_development" || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != len(names) {
		return m, fmt.Errorf("invalid public experimental manifest")
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return m, e
	}
	if len(entries) != len(names)+1 {
		return m, fmt.Errorf("unexpected package files")
	}
	for _, n := range names {
		b, e := read(filepath.Join(dir, n), 64<<20)
		if e != nil {
			return m, e
		}
		h := sha256.Sum256(b)
		if m.Files[n] != hex.EncodeToString(h[:]) {
			return m, fmt.Errorf("hash mismatch: %s", n)
		}
		if filepath.Ext(n) == ".rdh" {
			model, e := tinyhead.Decode(b)
			if e != nil {
				return m, e
			}
			if model.Width() != 1024 {
				return m, fmt.Errorf("unexpected head width")
			}
		}
		if n == "LICENSE" && m.Files[n] != "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9" {
			return m, fmt.Errorf("expected complete parent Apache license")
		}
		if n == "features.json" {
			var x struct {
				Schema     string `json:"schema"`
				Origin     string `json:"origin"`
				Provenance struct {
					Hash string `json:"parent_head_sha256"`
				} `json:"provenance"`
			}
			if e = json.Unmarshal(b, &x); e != nil {
				return m, e
			}
			if x.Schema != "riido-tiny-features-v1" || x.Origin != m.Origin || x.Provenance.Hash != "29cc0c479fb60fc038ef9555b0e082f1c84555d353ce7f89c40eca989a365627" {
				return m, fmt.Errorf("unexpected feature source")
			}
		}
		if n == "results.json" {
			var r struct {
				Experiment string `json:"experiment"`
				Parity     bool   `json:"fp32_reference_parity_pass"`
				FeatureSHA string `json:"feature_sha256"`
				Results    []struct {
					Name string `json:"name"`
					SHA  string `json:"sha256"`
				} `json:"results"`
			}
			if e = json.Unmarshal(b, &r); e != nil {
				return m, e
			}
			if r.Experiment != "tinyhead-01" || !r.Parity || r.FeatureSHA != m.Files["features.json"] || len(r.Results) != 6 {
				return m, fmt.Errorf("invalid experiment evidence")
			}
			seen := map[string]bool{}
			for _, v := range r.Results {
				if seen[v.Name] || m.Files[v.Name+".rdh"] != v.SHA || v.SHA == "" {
					return m, fmt.Errorf("result artifact mismatch")
				}
				seen[v.Name] = true
			}
		}
	}
	return m, nil
}
