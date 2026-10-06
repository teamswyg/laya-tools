// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Known-method supervised warm starts; fixed validation selection and held-out calibration.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintpilot"
)

type Pin struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Rows   int    `json:"rows"`
}
type Plan struct {
	Schema                               string `json:"schema"`
	Train, Validation, Calibration, Test Pin
	ParentSHA                            string            `json:"parent_sha256"`
	DriverSHA                            string            `json:"driver_sha256"`
	Sources                              map[string]string `json:"sources"`
	Arms                                 []Arm             `json:"arms"`
	Seed                                 int64             `json:"seed"`
	Batch                                int               `json:"batch"`
	Decay                                float64           `json:"decay"`
	Confidence                           float64           `json:"confidence"`
	Margin                               float64           `json:"margin"`
}
type Arm struct {
	Name   string  `json:"name"`
	Warm   bool    `json:"warm"`
	Rate   float64 `json:"learning_rate"`
	Epochs int     `json:"epochs"`
}
type Trial struct {
	Arm               Arm                 `json:"arm"`
	Updates           statehint.FitReport `json:"fit"`
	ValidationLoss    float64             `json:"validation_display_nll"`
	ValidationCorrect int                 `json:"validation_correct"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readPin(pin Pin) ([]statehintpilot.Case, error) {
	b, e := os.ReadFile(pin.Path)
	if e != nil {
		return nil, e
	}
	if len(b) > 8<<20 || digest(b) != pin.SHA256 {
		return nil, errors.New("data pin mismatch")
	}
	lines := bytes.Split(bytes.TrimSpace(b), []byte{'\n'})
	if len(lines) != pin.Rows {
		return nil, errors.New("row count mismatch")
	}
	rows := make([]statehintpilot.Case, 0, len(lines))
	ids, families, texts := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, line := range lines {
		row, e := strictRow(line)
		if e != nil {
			return nil, e
		}
		if _, ok := statehint.IntentIndex(row.Expected); !ok || row.ID == "" || row.Family == "" || row.Text == "" || len(row.Text) > statehint.MaxTextBytes || !utf8.ValidString(row.Text) || strings.ContainsRune(row.Text, 0) || !opaque(row.ID) || !opaque(row.Family) || ids[row.ID] || families[row.Family] || texts[row.Text] || row.Locale != "ko" && row.Locale != "en" {
			return nil, errors.New("invalid source row")
		}
		ids[row.ID], families[row.Family], texts[row.Text] = true, true, true
		rows = append(rows, row)
	}
	return rows, nil
}
func disjoint(groups ...[]statehintpilot.Case) error {
	ids, families, texts := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, rows := range groups {
		for _, r := range rows {
			if ids[r.ID] || families[r.Family] || texts[r.Text] {
				return errors.New("source overlap")
			}
			ids[r.ID], families[r.Family], texts[r.Text] = true, true, true
		}
	}
	return nil
}
func group(intent statehint.Intent) int {
	switch intent {
	case statehint.Progress:
		return 0
	case statehint.CompletionReport:
		return 1
	case statehint.Question:
		return 2
	default:
		return 3
	}
}
func loss(model *statehint.Model, rows []statehintpilot.Case) (float64, int, error) {
	var total float64
	correct := 0
	var w statehint.Workspace
	for _, r := range rows {
		p, e := model.Predict(r.Text, &w)
		if e != nil {
			return 0, 0, e
		}
		var groups [4]float64
		for i, value := range p.Probabilities {
			groups[group(statehint.Intents()[i])] += value
		}
		total -= math.Log(math.Max(groups[group(r.Expected)], 1e-15))
		if p.Intent == r.Expected {
			correct++
		}
	}
	return total / float64(len(rows)), correct, nil
}
func saveJSON(path string, value any) error {
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		return e
	}
	return f.Sync()
}
func artifact(model *statehint.Model) ([]byte, error) {
	var out bytes.Buffer
	e := model.Save(&out)
	return out.Bytes(), e
}
func predictor(model *statehint.Model) func(string) (statehint.Prediction, error) {
	var w statehint.Workspace
	return func(s string) (statehint.Prediction, error) { return model.Predict(s, &w) }
}
func run() error {
	planPath := flag.String("plan", "", "frozen plan")
	parentPath := flag.String("parent", "", "pinned parent")
	out := flag.String("out", "", "new output version")
	driver := flag.String("driver", "", "exact driver source")
	flag.Parse()
	if *planPath == "" || *parentPath == "" || *out == "" || *driver == "" {
		return errors.New("required arguments")
	}
	planBytes, e := os.ReadFile(*planPath)
	if e != nil {
		return e
	}
	var p Plan
	if e = json.Unmarshal(planBytes, &p); e != nil {
		return e
	}
	if p.Schema != "riido-statehint-v3-predeclared-v1" || p.Confidence != .9 || p.Margin != .05 || p.Seed != 1729 || p.Batch != 32 || p.Decay != .001 || len(p.Arms) != 4 || p.Train.Rows != 400 || p.Validation.Rows != 80 || p.Calibration.Rows != 80 || p.Test.Rows != 200 {
		return errors.New("unexpected frozen configuration")
	}
	expectedArms := []Arm{{Name: "warm_lr005", Warm: true, Rate: .005, Epochs: 40}, {Name: "warm_lr010", Warm: true, Rate: .01, Epochs: 40}, {Name: "warm_lr020", Warm: true, Rate: .02, Epochs: 40}, {Name: "fresh_lr020", Warm: false, Rate: .02, Epochs: 40}}
	for i, arm := range p.Arms {
		if arm != expectedArms[i] {
			return errors.New("frozen arm order or parameters changed")
		}
	}
	source, e := os.ReadFile(*driver)
	if e != nil || digest(source) != p.DriverSHA {
		return errors.New("driver pin mismatch")
	}
	required := []string{"go.mod", "go.sum"}
	for _, dir := range []string{"pkg/statehint", "pkg/statehintcatalog", "pkg/statehintpilot"} {
		entries, e := os.ReadDir(dir)
		if e != nil {
			return e
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
				required = append(required, filepath.Join(dir, entry.Name()))
			}
		}
	}
	if len(p.Sources) != len(required) {
		return errors.New("incomplete source pin set")
	}
	for _, path := range required {
		if len(p.Sources[path]) != 64 {
			return errors.New("missing required source pin")
		}
	}
	for path, hash := range p.Sources {
		b, e := os.ReadFile(path)
		if e != nil || digest(b) != hash {
			return errors.New("source pin mismatch")
		}
	}
	parentBytes, e := os.ReadFile(*parentPath)
	if e != nil || digest(parentBytes) != p.ParentSHA {
		return errors.New("parent pin mismatch")
	}
	parent, e := statehint.Load(bytes.NewReader(parentBytes))
	if e != nil || parent.TrainingSteps() == 0 {
		return errors.New("invalid trained parent")
	}
	train, e := readPin(p.Train)
	if e != nil {
		return e
	}
	validation, e := readPin(p.Validation)
	if e != nil {
		return e
	}
	calibration, e := readPin(p.Calibration)
	if e != nil {
		return e
	}
	if e = disjoint(train, validation, calibration); e != nil {
		return e
	}
	if _, e = os.Stat(*out); e == nil || !os.IsNotExist(e) {
		return errors.New("output version already exists or cannot be checked")
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	if e = saveJSON(filepath.Join(*out, "PLAN.before-fit.json"), p); e != nil {
		return e
	}
	samples := make([]statehint.Sample, len(train))
	for i, r := range train {
		samples[i] = statehint.Sample{Text: r.Text, Label: r.Expected}
	}
	var selected *statehint.Model
	selectedIndex := -1
	best := math.Inf(1)
	trials := make([]Trial, 0, len(p.Arms))
	for i, arm := range p.Arms {
		m := statehint.NewModel()
		if arm.Warm {
			m = parent.Clone()
		}
		fit, e := m.Fit(samples, statehint.FitOptions{Epochs: arm.Epochs, BatchSize: p.Batch, LearningRate: arm.Rate, WeightDecay: p.Decay, Seed: p.Seed})
		if e != nil {
			return e
		}
		nll, correct, e := loss(m, validation)
		if e != nil {
			return e
		}
		trials = append(trials, Trial{arm, fit, nll, correct})
		if nll < best {
			selected, best, selectedIndex = m, nll, i
		}
		b, e := artifact(m)
		if e != nil {
			return e
		}
		if e = writeBytes(filepath.Join(*out, arm.Name+".rsh"), b); e != nil {
			return e
		}
	}
	// Temperature grid is fixed before fitting. Only the selected model sees calibration.
	bestTemp, bestCal := 1., math.Inf(1)
	for step := 5; step <= 50; step++ {
		temp := float64(step) / 10
		if e = selected.SetTemperature(temp); e != nil {
			return e
		}
		nll, _, e := loss(selected, calibration)
		if e != nil {
			return e
		}
		if nll < bestCal {
			bestTemp, bestCal = temp, nll
		}
	}
	if e = selected.SetTemperature(bestTemp); e != nil {
		return e
	}
	modelBytes, e := artifact(selected)
	if e != nil {
		return e
	}
	if e = writeBytes(filepath.Join(*out, "selected.rsh"), modelBytes); e != nil {
		return e
	}
	lock := map[string]any{"schema": "v3-selected-model-lock-before-test-v1", "plan_sha256": digest(planBytes), "selected_arm_index": selectedIndex, "selected_arm": p.Arms[selectedIndex], "artifact_sha256": digest(modelBytes), "artifact_bytes": len(modelBytes), "temperature": bestTemp, "confidence": p.Confidence, "margin": p.Margin, "validation_display_nll": best, "calibration_display_nll": bestCal, "training_steps": selected.TrainingSteps(), "test_parsed_yet": false}
	if e = saveJSON(filepath.Join(*out, "LOCK.before-test.json"), lock); e != nil {
		return e
	}
	// The final rows are first parsed only after the selected artifact and lock exist.
	test, e := readPin(p.Test)
	if e != nil {
		return e
	}
	if e = disjoint(train, validation, calibration, test); e != nil {
		return e
	}
	child, e := statehintpilot.Evaluate(test, predictor(selected))
	if e != nil {
		return e
	}
	baseline, e := statehintpilot.Evaluate(test, predictor(parent))
	if e != nil {
		return e
	}
	rules, e := statehintpilot.Evaluate(test, statehint.RuleSpeechAct)
	if e != nil {
		return e
	}
	// Reload the saved artifact and compare every probability, not only labels.
	savedBytes, e := os.ReadFile(filepath.Join(*out, "selected.rsh"))
	if e != nil || digest(savedBytes) != digest(modelBytes) {
		return errors.New("saved file checksum mismatch")
	}
	reloaded, e := statehint.Load(bytes.NewReader(savedBytes))
	if e != nil {
		return e
	}
	var a, b statehint.Workspace
	for _, r := range test {
		x, e := selected.Predict(r.Text, &a)
		if e != nil {
			return e
		}
		y, e := reloaded.Predict(r.Text, &b)
		if e != nil || x != y {
			return errors.New("saved artifact parity failure")
		}
	}
	result := map[string]any{"schema": "v3-known-method-fit-and-development-evaluation-v1", "plan_sha256": digest(planBytes), "trials": trials, "lock": lock, "selected_test": child, "parent_test": baseline, "speechact_control": rules, "reload_probability_parity_cases": len(test), "data": "AI-authored original synthetic; not independent product ground truth", "live_reads": 0, "mutations": 0, "selection": "fixed validation four-group NLL; earlier arm on tie; test not used to choose weights/temperature"}
	if e = saveJSON(filepath.Join(*out, "RESULTS.first.json"), result); e != nil {
		return e
	}
	fmt.Println(strings.Join([]string{"completed", p.Arms[selectedIndex].Name, digest(modelBytes)}, " "))
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func opaque(s string) bool {
	if s == "" || len(s) > 256 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}
func strictRow(line []byte) (statehintpilot.Case, error) {
	var row statehintpilot.Case
	if !utf8.Valid(line) {
		return row, errors.New("invalid UTF8 row")
	}
	d := json.NewDecoder(bytes.NewReader(line))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') {
		return row, errors.New("invalid row object")
	}
	seen := [5]bool{}
	for d.More() {
		k, e := d.Token()
		if e != nil {
			return row, e
		}
		var target *string
		index := 0
		switch k {
		case "id":
			target = &row.ID
		case "family":
			target = &row.Family
			index = 1
		case "locale":
			target = &row.Locale
			index = 2
		case "text":
			target = &row.Text
			index = 3
		case "expected_intent":
			index = 4
		default:
			return row, errors.New("unknown row key")
		}
		if seen[index] {
			return row, errors.New("duplicate row key")
		}
		seen[index] = true
		v, e := d.Token()
		if e != nil {
			return row, e
		}
		value, ok := v.(string)
		if !ok {
			return row, errors.New("non-string row value")
		}
		if index == 4 {
			row.Expected = statehint.Intent(value)
		} else {
			*target = value
		}
	}
	t, e = d.Token()
	if e != nil || t != json.Delim('}') {
		return row, errors.New("row object not closed")
	}
	if _, e = d.Token(); e != io.EOF {
		return row, errors.New("trailing row value")
	}
	for _, v := range seen {
		if !v {
			return row, errors.New("missing row key")
		}
	}
	return row, nil
}

func writeBytes(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.Write(b); e != nil {
		return e
	}
	return f.Sync()
}
