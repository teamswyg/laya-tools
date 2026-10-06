// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// A frozen, small, source-separated synthetic comparison; not product truth.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"
)

type row struct {
	ID     string           `json:"id"`
	Group  string           `json:"group_id"`
	Locale string           `json:"locale"`
	Intent statehint.Intent `json:"intent"`
	Split  string           `json:"split"`
	Text   string           `json:"text"`
	Origin string           `json:"origin"`
}
type pin struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
	Rows int    `json:"rows"`
}
type plan struct {
	Schema      string            `json:"schema"`
	Files       map[string]pin    `json:"files"`
	ParentSHA   string            `json:"parent_sha256"`
	Rates       []float64         `json:"learning_rates"`
	Epochs      int               `json:"epochs"`
	Batch       int               `json:"batch_size"`
	Decay       float64           `json:"weight_decay"`
	Seed        int64             `json:"seed"`
	Threshold   float64           `json:"confidence_threshold"`
	SourcePins  map[string]string `json:"sources"`
	ManifestSHA string            `json:"evaluation_manifest_sha256"`
	RubricSHA   string            `json:"rubric_sha256"`
	Arms        []string          `json:"arms"`
}
type metrics struct {
	N, Correct, Accepted, AcceptedCorrect int
	Accuracy, NLL, Coverage               float64
	Precision                             *float64
	Confusion                             [8][8]int
}
type observed struct {
	ID, Group, Locale string
	Expected          statehint.Intent
	Prediction        statehint.Prediction
}
type candidate struct {
	name   string
	narrow *statehint.Model
	wide   *statehintwide.Model
}

func (c *candidate) predict(text string, w *statehint.Workspace, v *statehintwide.Workspace) (statehint.Prediction, error) {
	if c.narrow != nil {
		return c.narrow.Predict(text, w)
	}
	return c.wide.Predict(text, v)
}
func (c *candidate) temperature(t float64) error {
	if c.narrow != nil {
		return c.narrow.SetTemperature(t)
	}
	return c.wide.SetTemperature(t)
}
func (c *candidate) save(w io.Writer) error {
	if c.narrow != nil {
		return c.narrow.Save(w)
	}
	return c.wide.Save(w)
}
func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func pinned(p pin) ([]byte, error) {
	b, e := os.ReadFile(p.Path)
	if e != nil {
		return nil, e
	}
	if len(b) > 2<<20 || hash(b) != p.SHA {
		return nil, fmt.Errorf("frozen file mismatch")
	}
	return b, nil
}
func rowsFrom(p pin, split string) ([]row, error) {
	b, e := pinned(p)
	if e != nil {
		return nil, e
	}
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Buffer(make([]byte, 4096), 16384)
	var rows []row
	ids, texts, groups := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for s.Scan() {
		var r row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if e = d.Decode(&r); e != nil {
			return nil, e
		}
		var extra any
		if d.Decode(&extra) != io.EOF {
			return nil, fmt.Errorf("trailing row")
		}
		_, ok := statehint.IntentIndex(r.Intent)
		if !ok || r.ID == "" || r.Group == "" || r.Text == "" || r.Split != split || r.Origin != "original_authored_synthetic_development" || (r.Locale != "ko" && r.Locale != "en") || ids[r.ID] || groups[r.Group] || texts[r.Text] {
			return nil, fmt.Errorf("invalid/duplicated source case")
		}
		ids[r.ID], groups[r.Group], texts[r.Text] = true, true, true
		rows = append(rows, r)
	}
	if e = s.Err(); e != nil {
		return nil, e
	}
	if len(rows) != p.Rows {
		return nil, fmt.Errorf("row count drift")
	}
	return rows, nil
}
func disjoint(a, b []row) error {
	for _, x := range a {
		for _, y := range b {
			if x.ID == y.ID || x.Group == y.Group || x.Text == y.Text {
				return fmt.Errorf("cross-partition source overlap")
			}
		}
	}
	return nil
}
func evaluate(rows []row, predict func(string) (statehint.Prediction, error)) (metrics, []observed, error) {
	var m metrics
	var out []observed
	for _, r := range rows {
		p, e := predict(r.Text)
		if e != nil {
			return m, nil, e
		}
		i, _ := statehint.IntentIndex(r.Intent)
		j, ok := statehint.IntentIndex(p.Intent)
		if !ok {
			return m, nil, fmt.Errorf("unknown predicted intent")
		}
		m.N++
		m.Confusion[i][j]++
		correct := i == j
		if correct {
			m.Correct++
		}
		m.NLL -= math.Log(math.Max(p.Probabilities[i], 1e-15))
		if p.Intent != statehint.Unclear && p.Confidence >= .9 && p.GuardReason == "" {
			m.Accepted++
			if correct {
				m.AcceptedCorrect++
			}
		}
		out = append(out, observed{r.ID, r.Group, r.Locale, r.Intent, p})
	}
	if m.N == 0 {
		return m, nil, fmt.Errorf("empty evaluation")
	}
	m.Accuracy = float64(m.Correct) / float64(m.N)
	m.NLL /= float64(m.N)
	m.Coverage = float64(m.Accepted) / float64(m.N)
	if m.Accepted > 0 {
		v := float64(m.AcceptedCorrect) / float64(m.Accepted)
		m.Precision = &v
	}
	return m, out, nil
}
func predictor(c *candidate) func(string) (statehint.Prediction, error) {
	var w statehint.Workspace
	var v statehintwide.Workspace
	return func(s string) (statehint.Prediction, error) { return c.predict(s, &w, &v) }
}
func write(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(p, append(b, '\n'), 0600)
}
func mainrun() error {
	f := flag.NewFlagSet("statehint-pdca", flag.ContinueOnError)
	pf := f.String("plan", "", "frozen source/data plan")
	out := f.String("out", "", "new local output version")
	parentPath := f.String("parent", "", "pinned v1 local artifact")
	if e := f.Parse(os.Args[1:]); e != nil {
		return e
	}
	if *pf == "" || *out == "" || *parentPath == "" {
		return fmt.Errorf("plan/out/parent required")
	}
	b, e := os.ReadFile(*pf)
	if e != nil {
		return e
	}
	var p plan
	if e = json.Unmarshal(b, &p); e != nil {
		return e
	}
	if p.Schema != "riido-statehint-pdca-v2" || p.Epochs != 40 || p.Batch != 32 || p.Decay != .001 || p.Seed != 1729 || p.Threshold != .9 || len(p.Rates) != 1 || p.Rates[0] != .02 || len(p.Files) != 4 || len(p.SourcePins) < 5 {
		return fmt.Errorf("unsupported frozen recipe")
	}
	for _, entry := range []struct{ path, sha string }{{"experiments/state-hints-v2/manifest.json", p.ManifestSHA}, {"experiments/state-hints-v2/RUBRIC.md", p.RubricSHA}} {
		data, e := os.ReadFile(entry.path)
		if e != nil || hash(data) != entry.sha {
			return fmt.Errorf("manifest or rubric drift")
		}
	}
	wantedArms := []string{"v1_warm_start", "v1_fresh", "wide_simple_fresh", "wide_context_fresh"}
	if len(p.Arms) != len(wantedArms) {
		return fmt.Errorf("candidate arms changed")
	}
	for i, want := range wantedArms {
		if p.Arms[i] != want {
			return fmt.Errorf("candidate order changed")
		}
	}
	for file, pin := range p.SourcePins {
		data, e := os.ReadFile(file)
		if e != nil || hash(data) != pin {
			return fmt.Errorf("source drift")
		}
	}
	parentBytes, e := os.ReadFile(*parentPath)
	if e != nil || hash(parentBytes) != p.ParentSHA {
		return fmt.Errorf("parent pin mismatch")
	}
	parent, e := statehint.Load(bytes.NewReader(parentBytes))
	if e != nil {
		return e
	}
	// The final file is neither read nor parsed until the model/temperature lock.
	train, e := rowsFrom(p.Files["train"], "train")
	if e != nil {
		return e
	}
	val, e := rowsFrom(p.Files["validation"], "validation")
	if e != nil {
		return e
	}
	cal, e := rowsFrom(p.Files["calibration"], "calibration")
	if e != nil {
		return e
	}
	for _, pair := range [][2][]row{{train, val}, {train, cal}, {val, cal}} {
		if e = disjoint(pair[0], pair[1]); e != nil {
			return e
		}
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	started := time.Now()
	narrowSamples := make([]statehint.Sample, len(train))
	wideSamples := make([]statehintwide.Sample, len(train))
	for i, r := range train {
		narrowSamples[i] = statehint.Sample{Text: r.Text, Label: r.Intent}
		wideSamples[i] = statehintwide.Sample{Text: r.Text, Label: r.Intent}
	}
	candidates := []*candidate{{name: "v1_warm_start", narrow: parent.Clone()}, {name: "v1_fresh", narrow: statehint.NewModel()}, {name: "wide_simple_fresh", wide: statehintwide.NewModel(statehintwide.Simple)}, {name: "wide_context_fresh", wide: statehintwide.NewModel(statehintwide.Contextual)}}
	history := []any{}
	var selected *candidate
	best := math.Inf(1)
	for _, c := range candidates {
		if c.narrow != nil {
			_, e = c.narrow.Fit(narrowSamples, statehint.FitOptions{Epochs: p.Epochs, BatchSize: p.Batch, LearningRate: .02, WeightDecay: p.Decay, Seed: p.Seed})
		} else {
			_, e = c.wide.Fit(wideSamples, statehintwide.FitOptions{Epochs: p.Epochs, BatchSize: p.Batch, LearningRate: .02, WeightDecay: p.Decay, Seed: p.Seed})
		}
		if e != nil {
			return e
		}
		m, _, e := evaluate(val, predictor(c))
		if e != nil {
			return e
		}
		var model bytes.Buffer
		if e = c.save(&model); e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(*out, c.name+".rsh"), model.Bytes(), 0600); e != nil {
			return e
		}
		entry := struct {
			Name        string
			Validation  metrics
			ArtifactSHA string
			Bytes       int
		}{c.name, m, hash(model.Bytes()), model.Len()}
		history = append(history, entry)
		if e = json.NewEncoder(os.Stdout).Encode(entry); e != nil {
			return e
		}
		if m.NLL < best {
			best = m.NLL
			selected = c
		}
	}
	bestT, bestCal := 1., math.Inf(1)
	for i := 0; i <= 45; i++ {
		t := .5 + float64(i)*.1
		if e = selected.temperature(t); e != nil {
			return e
		}
		m, _, e := evaluate(cal, predictor(selected))
		if e != nil {
			return e
		}
		if m.NLL < bestCal {
			bestT, bestCal = t, m.NLL
		}
	}
	if e = selected.temperature(bestT); e != nil {
		return e
	}
	var selectedBytes bytes.Buffer
	if e = selected.save(&selectedBytes); e != nil {
		return e
	}
	if e = os.WriteFile(filepath.Join(*out, "selected.rsh"), selectedBytes.Bytes(), 0600); e != nil {
		return e
	}
	lock := struct {
		Schema, Selected, ArtifactSHA, PlanSHA string
		Temperature, Threshold                 float64
	}{"riido-statehint-v2-before-test-lock", selected.name, hash(selectedBytes.Bytes()), hash(b), bestT, .9}
	if e = write(filepath.Join(*out, "LOCK.before-test.json"), lock); e != nil {
		return e
	}
	// Only now parse the new final test. It is never used to select a candidate.
	test, e := rowsFrom(p.Files["test"], "test")
	if e != nil {
		return e
	}
	for _, used := range [][]row{train, val, cal} {
		if e = disjoint(used, test); e != nil {
			return e
		}
	}
	controls := map[string]func(string) (statehint.Prediction, error){"parent_v1": predictor(&candidate{name: "parent_v1", narrow: parent}), "selected": predictor(selected), "rules_v1": statehint.RuleBaseline, "rules_speechact": statehint.RuleSpeechAct}
	reports := map[string]metrics{}
	observations := map[string][]observed{}
	for name, fn := range controls {
		m, observed, e := evaluate(test, fn)
		if e != nil {
			return e
		}
		reports[name+"/all"] = m
		observations[name] = observed
		for _, locale := range []string{"ko", "en"} {
			var filtered []row
			for _, r := range test {
				if r.Locale == locale {
					filtered = append(filtered, r)
				}
			}
			m, _, e = evaluate(filtered, fn)
			if e != nil {
				return e
			}
			reports[name+"/"+locale] = m
		}
	}
	report := struct {
		Schema              string
		PlanSHA             string
		History             []any
		Lock                any
		Metrics             map[string]metrics
		Observations        map[string][]observed
		ElapsedNS           int64
		DeploymentQualified bool
		Scope               string
	}{"riido-statehint-v2-development-results", hash(b), history, lock, reports, observations, time.Since(started).Nanoseconds(), false, "Original synthetic free-context cases;v1 exposed diagnostics informed hypotheses. Separate320training/40validation/40calibration/40sealedfinal cases, no filler variants. Agent-authored labels are not independent product truth. Confidence.9 unchanged; full failures retained. Model/rules did not execute Riido writes."}
	if e = write(filepath.Join(*out, "report.json"), report); e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Selected    string
		Test        metrics
		Temperature float64
	}{selected.name, reports["selected/all"], bestT})
}
func main() {
	if e := mainrun(); e != nil {
		fmt.Fprintln(os.Stderr, "statehint-pdca:", e)
		os.Exit(1)
	}
}
