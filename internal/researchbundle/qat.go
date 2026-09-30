package researchbundle

import (
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/ternarytrain"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"math"
	"os"
	"path/filepath"
	"regexp"
)

var qatNames = []string{"LICENSE", "NOTICE", "README.md", "PLAN.json", "plan-01.json", "selection.json", "development-cycle-01.json", "results.json", "benchmark.json", "development-features.json", "final-features.json", "final-families.json", "seed-1729.rdh", "seed-2718.rdh", "cycle01-seed-1729.rdh", "cycle01-seed-2718.rdh"}
var qatRepo = regexp.MustCompile(`^JooYoon/riidolaya-ternary-qat-v[0-9]+\.[0-9]+$`)

func VerifyQAT(dir string, m Manifest) (Manifest, error) {
	if !qatRepo.MatchString(m.Repository) || !sha.MatchString(m.SourceRevision) || m.ParentRevision != "9921e2bafeeb6429d773b03f4d95f8a348634c0a" || m.Origin != "original_synthetic_ternary_qat" || m.License != "apache-2.0" || m.ProductionReady || len(m.Files) != len(qatNames) {
		return m, fmt.Errorf("invalid QAT manifest")
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		return m, e
	}
	if len(entries) != len(qatNames)+1 {
		return m, fmt.Errorf("unexpected QAT files")
	}
	for _, n := range qatNames {
		b, e := read(filepath.Join(dir, n), 64<<20)
		if e != nil {
			return m, e
		}
		if ternarytrain.Hash(b) != m.Files[n] {
			return m, fmt.Errorf("QAT hash mismatch: %s", n)
		}
	}
	if m.Files["LICENSE"] != "a6cba85bc92e0cff7a450b1d873c0eaa2e9fc96bf472df0247a26bec77bf3ff9" {
		return m, fmt.Errorf("expected parent Apache license")
	}
	parse := func(name string, v any) error {
		b, e := read(filepath.Join(dir, name), 64<<20)
		if e != nil {
			return e
		}
		return json.Unmarshal(b, v)
	}
	var plan struct {
		Experiment string `json:"experiment"`
		Dev        string `json:"development_feature_sha256"`
		Final      string `json:"final_families_sha256"`
	}
	if e = parse("PLAN.json", &plan); e != nil {
		return m, e
	}
	if plan.Experiment != "ternary-qat-02" || plan.Dev != m.Files["development-features.json"] || plan.Final != m.Files["final-families.json"] {
		return m, fmt.Errorf("QAT plan/data mismatch")
	}
	type checkpoint struct {
		File  string `json:"file"`
		SHA   string `json:"sha256"`
		Seed  int64  `json:"seed"`
		Bytes int    `json:"bytes"`
	}
	type seal struct {
		Experiment string       `json:"experiment"`
		Plan       string       `json:"plan_sha256"`
		Dev        string       `json:"development_feature_sha256"`
		Final      string       `json:"final_families_sha256"`
		Opened     bool         `json:"final_opened"`
		Models     []checkpoint `json:"models"`
	}
	var selected seal
	if e = parse("selection.json", &selected); e != nil {
		return m, e
	}
	if selected.Experiment != plan.Experiment || selected.Plan != m.Files["PLAN.json"] || selected.Dev != plan.Dev || selected.Final != plan.Final || selected.Opened || len(selected.Models) != 2 {
		return m, fmt.Errorf("QAT selection seal mismatch")
	}
	var archived seal
	if e = parse("development-cycle-01.json", &archived); e != nil {
		return m, e
	}
	if archived.Experiment != "ternary-qat-01" || archived.Plan != m.Files["plan-01.json"] || archived.Dev != plan.Dev || archived.Final != plan.Final || archived.Opened || len(archived.Models) != 2 {
		return m, fmt.Errorf("archived cycle seal mismatch")
	}
	for i, c := range archived.Models {
		if c.Seed != []int64{1729, 2718}[i] || c.File != fmt.Sprintf("seed-%d.rdh", c.Seed) || m.Files["cycle01-"+c.File] != c.SHA {
			return m, fmt.Errorf("archived checkpoint mismatch")
		}
		b, e := read(filepath.Join(dir, "cycle01-"+c.File), 1<<20)
		if e != nil {
			return m, e
		}
		h, e := tinyhead.Decode(b)
		if e != nil || h.Width() != 1024 || h.Bytes() != c.Bytes {
			return m, fmt.Errorf("invalid archived model")
		}
	}
	var report struct {
		Experiment string               `json:"experiment"`
		Plan       string               `json:"plan_sha256"`
		Selection  string               `json:"selection_sha256"`
		Final      string               `json:"final_features_sha256"`
		Pass       bool                 `json:"both_seeds_pass"`
		Parent     ternarytrain.Metrics `json:"parent_fp32"`
		Results    []struct {
			Model       checkpoint                `json:"model"`
			Metrics     ternarytrain.Metrics      `json:"metrics"`
			Pass        bool                      `json:"gate_pass"`
			Predictions []ternarytrain.Prediction `json:"predictions"`
		} `json:"results"`
	}
	if e = parse("results.json", &report); e != nil {
		return m, e
	}
	if report.Experiment != plan.Experiment || report.Plan != m.Files["PLAN.json"] || report.Selection != m.Files["selection.json"] || report.Final != m.Files["final-features.json"] || len(report.Results) != 2 || !report.Pass {
		return m, fmt.Errorf("QAT result seal/gate mismatch")
	}
	dev, _, e := ternarytrain.Load(filepath.Join(dir, "development-features.json"))
	if e != nil {
		return m, e
	}
	final, _, e := ternarytrain.Load(filepath.Join(dir, "final-features.json"))
	if e != nil {
		return m, e
	}
	if len(final.Rows) != 36 || final.Schema != "riido-qat-final-v1" {
		return m, fmt.Errorf("wrong final dataset")
	}
	// Link feature IDs and labels to the reviewed public authored cases.
	var cases []struct {
		ID    string `json:"id"`
		Label int    `json:"label"`
	}
	if e = parse("final-families.json", &cases); e != nil {
		return m, e
	}
	if len(cases) != len(final.Rows) {
		return m, fmt.Errorf("final family count mismatch")
	}
	for i, r := range final.Rows {
		if r.ID != cases[i].ID || r.Label != cases[i].Label {
			return m, fmt.Errorf("final labels/identities changed")
		}
	}
	parent, e := tinyhead.Build(dev.Source, tinyhead.Float32, 0)
	if e != nil {
		return m, e
	}
	baseline, _, e := ternarytrain.Evaluate(parent, final.Rows)
	if e != nil {
		return m, e
	}
	if !sameMetrics(baseline, report.Parent) {
		return m, fmt.Errorf("parent reference recomputation failed")
	}
	for i, c := range selected.Models {
		if c.Seed != []int64{1729, 2718}[i] || c.File != fmt.Sprintf("seed-%d.rdh", c.Seed) || m.Files[c.File] != c.SHA || c != report.Results[i].Model {
			return m, fmt.Errorf("selected QAT checkpoint mismatch")
		}
		b, e := read(filepath.Join(dir, c.File), 1<<20)
		if e != nil {
			return m, e
		}
		h, e := tinyhead.Decode(b)
		if e != nil {
			return m, e
		}
		if h.Width() != 1024 || h.Bytes() != c.Bytes {
			return m, fmt.Errorf("invalid QAT head")
		}
		actual, pred, e := ternarytrain.Evaluate(h, final.Rows)
		if e != nil {
			return m, e
		}
		r := report.Results[i]
		if !ternarytrain.Pass(actual, baseline, h.Bytes()) || !r.Pass || !sameMetrics(actual, r.Metrics) || len(pred) != len(r.Predictions) {
			return m, fmt.Errorf("QAT quality recomputation failed")
		}
		for j, p := range pred {
			v := r.Predictions[j]
			if p.ID != v.ID || p.Label != v.Label || p.Winner != v.Winner {
				return m, fmt.Errorf("prediction identity mismatch")
			}
			for k := 0; k < 3; k++ {
				if math.Abs(p.Probability[k]-v.Probability[k]) > 1e-9 {
					return m, fmt.Errorf("probability mismatch")
				}
			}
		}
	}
	return m, nil
}
func sameMetrics(a, b ternarytrain.Metrics) bool {
	if a.Cases != b.Cases || a.Correct != b.Correct || a.Accepted != b.Accepted || a.AcceptedCorrect != b.AcceptedCorrect || a.Strong != b.Strong || a.StrongCorrect != b.StrongCorrect || a.StrongToFast != b.StrongToFast {
		return false
	}
	for _, p := range [][2]float64{{a.Accuracy, b.Accuracy}, {a.Precision, b.Precision}, {a.Coverage, b.Coverage}, {a.Recall, b.Recall}, {a.NLL, b.NLL}, {a.Brier, b.Brier}} {
		if math.IsNaN(p[1]) || math.IsInf(p[1], 0) || math.Abs(p[0]-p[1]) > 1e-9 {
			return false
		}
	}
	return true
}
