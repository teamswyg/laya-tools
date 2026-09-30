package researchbundle

import (
	"encoding/json"
	"fmt"
	"github.com/teamswyg/laya-tools/internal/ternarytrain"
	"github.com/teamswyg/laya-tools/internal/tinyhead"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func qatStage(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string][]byte{}
	for _, n := range qatNames {
		files[n] = []byte("public test fixture")
	}
	license, e := os.ReadFile("../../LICENSE")
	if e != nil {
		t.Fatal(e)
	}
	files["LICENSE"] = license
	marshal := func(n string, v any) {
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		files[n] = b
	}
	source := tinyhead.Source{Gamma: make([]float64, 1024), Beta: make([]float64, 1024), Weight: make([]float64, 3072), Temperature: .5}
	for i := range source.Gamma {
		source.Gamma[i] = 1
	}
	for c := 0; c < 3; c++ {
		source.Weight[c*1024+c] = 1
	}
	head, e := tinyhead.Build(source, tinyhead.Ternary, 1)
	if e != nil {
		t.Fatal(e)
	}
	parent, _ := tinyhead.Build(source, tinyhead.Float32, 0)
	modelSHA := ternarytrain.Hash(head.Encode())
	dev := ternarytrain.Dataset{Schema: "riido-tiny-features-v1", Origin: "original_synthetic_pdca06_development", Source: source}
	final := ternarytrain.Dataset{Schema: "riido-qat-final-v1", Origin: "original_synthetic_ternary_qat_final"}
	cases := []map[string]any{}
	row := func(id, split string, c int) ternarytrain.Row {
		x := make([]float64, 1024)
		x[c] = 1
		p, _ := head.Predict(x, make([]float64, 1024))
		return ternarytrain.Row{ID: id, Split: split, Label: c, Feature: x, Reference: p}
	}
	for _, split := range []string{"train", "validation", "calibration"} {
		for c := 0; c < 3; c++ {
			dev.Rows = append(dev.Rows, row(fmt.Sprintf("%s-%d", split, c), split, c))
		}
	}
	for i := 0; i < 36; i++ {
		id := fmt.Sprintf("final-%d", i)
		final.Rows = append(final.Rows, row(id, "final", i%3))
		cases = append(cases, map[string]any{"id": id, "label": i % 3})
	}
	marshal("development-features.json", dev)
	marshal("final-features.json", final)
	marshal("final-families.json", cases)
	marshal("PLAN.json", map[string]any{"experiment": "ternary-qat-02", "development_feature_sha256": ternarytrain.Hash(files["development-features.json"]), "final_families_sha256": ternarytrain.Hash(files["final-families.json"])})
	files["plan-01.json"] = []byte("{}")
	models := []map[string]any{}
	for _, seed := range []int64{1729, 2718} {
		name := fmt.Sprintf("seed-%d.rdh", seed)
		files[name] = head.Encode()
		files["cycle01-"+name] = head.Encode()
		models = append(models, map[string]any{"file": name, "seed": seed, "sha256": modelSHA, "bytes": head.Bytes()})
	}
	for _, v := range []struct{ file, experiment, plan string }{{"selection.json", "ternary-qat-02", "PLAN.json"}, {"development-cycle-01.json", "ternary-qat-01", "plan-01.json"}} {
		marshal(v.file, map[string]any{"experiment": v.experiment, "plan_sha256": ternarytrain.Hash(files[v.plan]), "development_feature_sha256": ternarytrain.Hash(files["development-features.json"]), "final_families_sha256": ternarytrain.Hash(files["final-families.json"]), "models": models, "final_opened": false})
	}
	baseline, _, _ := ternarytrain.Evaluate(parent, final.Rows)
	metrics, pred, _ := ternarytrain.Evaluate(head, final.Rows)
	results := []map[string]any{}
	for _, m := range models {
		results = append(results, map[string]any{"model": m, "metrics": metrics, "gate_pass": true, "predictions": pred})
	}
	marshal("results.json", map[string]any{"experiment": "ternary-qat-02", "plan_sha256": ternarytrain.Hash(files["PLAN.json"]), "selection_sha256": ternarytrain.Hash(files["selection.json"]), "final_features_sha256": ternarytrain.Hash(files["final-features.json"]), "both_seeds_pass": true, "parent_fp32": baseline, "results": results})
	m := Manifest{Schema: "riido-qat-bundle-v1", Repository: "JooYoon/riidolaya-ternary-qat-v0.1", SourceRevision: strings.Repeat("a", 40), ParentRevision: "9921e2bafeeb6429d773b03f4d95f8a348634c0a", Origin: "original_synthetic_ternary_qat", License: "apache-2.0", Files: map[string]string{}}
	for n, b := range files {
		m.Files[n] = ternarytrain.Hash(b)
		if e = os.WriteFile(filepath.Join(dir, n), b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	b, _ := json.Marshal(m)
	os.WriteFile(filepath.Join(dir, "MANIFEST.json"), b, 0600)
	return dir
}
func TestQATRecomputesQualityEvenIfManifestResigned(t *testing.T) {
	dir := qatStage(t)
	if _, e := Verify(dir); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(dir, "results.json")
	b, _ := os.ReadFile(path)
	var report map[string]any
	json.Unmarshal(b, &report)
	report["results"].([]any)[0].(map[string]any)["metrics"].(map[string]any)["correct"] = 35
	b, _ = json.Marshal(report)
	os.WriteFile(path, b, 0600)
	path = filepath.Join(dir, "MANIFEST.json")
	raw, _ := os.ReadFile(path)
	var m Manifest
	json.Unmarshal(raw, &m)
	m.Files["results.json"] = ternarytrain.Hash(b)
	raw, _ = json.Marshal(m)
	os.WriteFile(path, raw, 0600)
	if _, e := Verify(dir); e == nil {
		t.Fatal("forged metrics accepted after hash update")
	}
}
