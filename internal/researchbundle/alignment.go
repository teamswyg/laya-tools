package researchbundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/alignment"
	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/paireval"
	"github.com/teamswyg/laya-tools/internal/staticembed"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const alignmentResultsSHA = "3fe1a1ce24cd110dcbe7ef54155c46e58540a42d03914837d12416ed9e8ae78c"
const alignmentPlanSHA = "0ea3776307eafa3012d03a0b969eaa2220247a336e7a4105eac74d2e75b2bd8d"
const alignmentPartitionSHA = "d90eda8e8a423eff44847ff46ef19b468a57ac2791d45c57e7e8ba39a4ee9902"

var alignmentRepo = regexp.MustCompile(`^JooYoon/riidolaya-alignment-research-v[0-9]+\.[0-9]+$`)

type alignmentArtifact struct {
	Variant, Encoder, Head, File, SHA256 string
	Seed                                 uint64
	Bytes                                int
}
type alignmentModel struct {
	Schema, SourceSHA256, PartitionSHA256, EmbeddingRevision, Variant, Encoder, Head string
	Seed                                                                             uint64
	Weights                                                                          []float64
}

func alignmentName(a alignmentArtifact) (string, error) {
	if alignment.Dimension(a.Variant) == 0 || (a.Seed != 1729 && a.Seed != 2718) {
		return "", fmt.Errorf("invalid variant/seed")
	}
	if a.Variant == "lexical" {
		if a.Encoder != "none" {
			return "", fmt.Errorf("lexical encoder mismatch")
		}
	} else if a.Encoder != "fp32" && a.Encoder != "ternary" {
		return "", fmt.Errorf("invalid encoder")
	}
	switch a.Head {
	case "fp32", "int8", "ternary_ptq", "ternary_ste":
	default:
		return "", fmt.Errorf("invalid head")
	}
	return fmt.Sprintf("%s-%s-%s-%d.align.json", a.Variant, a.Encoder, a.Head, a.Seed), nil
}
func verifyAlignmentModel(b []byte, a alignmentArtifact) error {
	var m alignmentModel
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(&m); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("extra model content")
	}
	if m.Schema != alignment.Schema || m.SourceSHA256 != paireval.SourceSHA || m.PartitionSHA256 != alignmentPartitionSHA || m.EmbeddingRevision != staticembed.Revision || m.Variant != a.Variant || m.Encoder != a.Encoder || m.Head != a.Head || m.Seed != a.Seed || len(m.Weights) != alignment.Dimension(a.Variant) {
		return fmt.Errorf("head contract mismatch")
	}
	for _, w := range m.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || math.Abs(w) > 1e6 {
			return fmt.Errorf("invalid coefficient")
		}
	}
	switch a.Head {
	case "fp32", "int8":
		q := hintlearn.Quantize(m.Weights, a.Head)
		for i, w := range q {
			if math.Abs(w-m.Weights[i]) > 1e-7*math.Max(1, math.Abs(w)) {
				return fmt.Errorf("quantization mismatch")
			}
		}
	default:
		scale := 0.
		for _, w := range m.Weights {
			if w == 0 {
				continue
			}
			if scale == 0 {
				scale = math.Abs(w)
			}
			if math.Abs(w) != scale {
				return fmt.Errorf("nonternary coefficient")
			}
		}
	}
	return nil
}

// VerifyAlignment binds every export to the immutable reviewed aggregate report
// and its source/plan. It checks release integrity, not model usefulness or a
// fresh recomputation of dataset metrics. Reproduction is separate evidence.
func VerifyAlignment(dir string, m Manifest) (Manifest, error) {
	if m.Schema != "riido-alignment-bundle-v1" || !alignmentRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != staticembed.Revision || m.Origin != "public_cosqa_computational_results_05" || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != 46 {
		return m, fmt.Errorf("invalid alignment manifest")
	}
	rb, e := read(filepath.Join(dir, "results.json"), 1<<20)
	if e != nil {
		return m, e
	}
	if paireval.Hash(rb) != alignmentResultsSHA {
		return m, fmt.Errorf("unreviewed results")
	}
	var results struct {
		Models                          []alignmentArtifact
		Promotion, PrimaryReadyForFinal bool
	}
	if e = json.Unmarshal(rb, &results); e != nil {
		return m, e
	}
	if len(results.Models) != 40 || results.Promotion || results.PrimaryReadyForFinal {
		return m, fmt.Errorf("unexpected experiment state")
	}
	expected := []string{"LICENSE", "NOTICE", "README.md", "README.ko.md", "results.json", "plan.json"}
	seen := map[string]bool{}
	for _, a := range results.Models {
		name, e := alignmentName(a)
		if e != nil {
			return m, e
		}
		if name != a.File || seen[name] {
			return m, fmt.Errorf("invalid export name")
		}
		seen[name] = true
		expected = append(expected, name)
		b, e := read(filepath.Join(dir, name), 1<<16)
		if e != nil {
			return m, e
		}
		if len(b) != a.Bytes || paireval.Hash(b) != a.SHA256 {
			return m, fmt.Errorf("model does not match recorded export: %s", name)
		}
		if e = verifyAlignmentModel(b, a); e != nil {
			return m, e
		}
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return m, e
	}
	if len(entries) != len(expected)+1 {
		return m, fmt.Errorf("extra release files")
	}
	for _, name := range expected {
		b, e := read(filepath.Join(dir, name), 1<<20)
		if e != nil {
			return m, e
		}
		if m.Files[name] != paireval.Hash(b) {
			return m, fmt.Errorf("manifest hash mismatch: %s", name)
		}
		switch name {
		case "LICENSE":
			if m.Files[name] != "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9" {
				return m, fmt.Errorf("license mismatch")
			}
		case "plan.json":
			if m.Files[name] != alignmentPlanSHA {
				return m, fmt.Errorf("plan mismatch")
			}
		case "NOTICE":
			if !strings.Contains(string(b), "C-UDA") || !strings.Contains(string(b), staticembed.Revision) {
				return m, fmt.Errorf("missing source notices")
			}
		case "README.md":
			if !strings.Contains(string(b), "Not production ready") || !strings.Contains(string(b), "No LLM savings established") {
				return m, fmt.Errorf("missing research limitations")
			}
		}
	}
	return m, nil
}
