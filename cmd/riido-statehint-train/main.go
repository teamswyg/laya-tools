// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Original synthetic supervised classifier training. This is not Laya fine-tuning.
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
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"
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
type metric struct {
	N               int       `json:"n"`
	Correct         int       `json:"correct"`
	Accuracy        float64   `json:"accuracy"`
	NLL             float64   `json:"nll"`
	Brier           float64   `json:"brier"`
	Accepted        int       `json:"accepted"`
	AcceptedCorrect int       `json:"accepted_correct"`
	Coverage        float64   `json:"coverage"`
	Precision       *float64  `json:"accepted_precision"`
	Confusion       [8][8]int `json:"confusion"`
}
type resultRow struct {
	ID, Group, Locale, Split string
	Target                   statehint.Intent
	Prediction               statehint.Prediction
}
type trial struct {
	LearningRate float64             `json:"learning_rate"`
	Epochs       int                 `json:"epochs"`
	Validation   metric              `json:"validation"`
	Fit          statehint.FitReport `json:"fit"`
	ModelSHA     string              `json:"model_sha256"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func load(p string) ([]row, string, error) {
	b, e := os.ReadFile(p)
	if e != nil {
		return nil, "", e
	}
	if len(b) > 2<<20 {
		return nil, "", fmt.Errorf("corpus bound")
	}
	s := bufio.NewScanner(bytes.NewReader(b))
	s.Buffer(make([]byte, 4096), 16384)
	var rows []row
	ids := map[string]bool{}
	texts := map[string]bool{}
	groups := map[string]string{}
	for s.Scan() {
		var r row
		d := json.NewDecoder(bytes.NewReader(s.Bytes()))
		d.DisallowUnknownFields()
		if e = d.Decode(&r); e != nil {
			return nil, "", e
		}
		var extra any
		if d.Decode(&extra) != io.EOF {
			return nil, "", fmt.Errorf("trailing corpus item")
		}
		_, ok := statehint.IntentIndex(r.Intent)
		if !ok || r.ID == "" || r.Group == "" || r.Text == "" || len(r.Text) > 4096 || r.Origin != "original_authored_synthetic_development" || (r.Locale != "ko" && r.Locale != "en") || ids[r.ID] || texts[r.Text] {
			return nil, "", fmt.Errorf("invalid or duplicate corpus row")
		}
		if r.Split != "train" && r.Split != "validation" && r.Split != "calibration" && r.Split != "test" {
			return nil, "", fmt.Errorf("unknown split")
		}
		if old, ok := groups[r.Group]; ok && old != r.Split {
			return nil, "", fmt.Errorf("family split leakage")
		}
		ids[r.ID] = true
		texts[r.Text] = true
		groups[r.Group] = r.Split
		rows = append(rows, r)
	}
	if e = s.Err(); e != nil {
		return nil, "", e
	}
	if len(rows) != 2400 || len(groups) != 96 {
		return nil, "", fmt.Errorf("corpus size or family count changed")
	}
	return rows, hash(b), nil
}
func subset(rows []row, split, locale string) []row {
	var out []row
	for _, r := range rows {
		if r.Split == split && (locale == "" || r.Locale == locale) {
			out = append(out, r)
		}
	}
	return out
}
func evaluate(rows []row, predict func(string) (statehint.Prediction, error)) (metric, []resultRow, error) {
	var m metric
	var out []resultRow
	for _, r := range rows {
		p, e := predict(r.Text)
		if e != nil {
			return m, nil, e
		}
		target, _ := statehint.IntentIndex(r.Intent)
		guess, ok := statehint.IntentIndex(p.Intent)
		if !ok {
			return m, nil, fmt.Errorf("invalid predicted intent")
		}
		m.N++
		m.Confusion[target][guess]++
		correct := r.Intent == p.Intent
		if correct {
			m.Correct++
		}
		m.NLL -= math.Log(math.Max(p.Probabilities[target], 1e-15))
		for i, v := range p.Probabilities {
			y := 0.
			if i == target {
				y = 1
			}
			m.Brier += (v - y) * (v - y)
		}
		if p.Confidence >= .9 && p.Intent != statehint.Unclear && p.GuardReason == "" {
			m.Accepted++
			if correct {
				m.AcceptedCorrect++
			}
		}
		out = append(out, resultRow{r.ID, r.Group, r.Locale, r.Split, r.Intent, p})
	}
	if m.N == 0 {
		return m, nil, fmt.Errorf("empty evaluation")
	}
	m.Accuracy = float64(m.Correct) / float64(m.N)
	m.NLL /= float64(m.N)
	m.Brier /= float64(m.N)
	m.Coverage = float64(m.Accepted) / float64(m.N)
	if m.Accepted > 0 {
		v := float64(m.AcceptedCorrect) / float64(m.Accepted)
		m.Precision = &v
	}
	return m, out, nil
}
func predictor(m *statehint.Model) func(string) (statehint.Prediction, error) {
	var w statehint.Workspace
	return func(s string) (statehint.Prediction, error) { return m.Predict(s, &w) }
}
func writeJSON(p string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	b = append(b, '\n')
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	ce := f.Close()
	if e != nil {
		return e
	}
	return ce
}
func save(p string, m *statehint.Model) (string, error) {
	var b bytes.Buffer
	if e := m.Save(&b); e != nil {
		return "", e
	}
	f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return "", e
	}
	_, e = f.Write(b.Bytes())
	ce := f.Close()
	if e != nil {
		return "", e
	}
	if ce != nil {
		return "", ce
	}
	return hash(b.Bytes()), nil
}
func run() error {
	f := flag.NewFlagSet("riido-statehint-train", flag.ContinueOnError)
	data := f.String("data", "experiments/state-hints/corpus.jsonl", "frozen original synthetic corpus")
	out := f.String("out", "", "new private version directory")
	plan := f.String("plan", "", "frozen training plan")
	if e := f.Parse(os.Args[1:]); e != nil {
		return e
	}
	if *out == "" || *plan == "" {
		return fmt.Errorf("--out and --plan required")
	}
	rows, dataSHA, e := load(*data)
	if e != nil {
		return e
	}
	planBytes, e := os.ReadFile(*plan)
	if e != nil {
		return e
	}
	var contract struct {
		Schema       string  `json:"schema"`
		DataSHA      string  `json:"data_sha256"`
		Confidence   float64 `json:"confidence_threshold"`
		Seed         int64   `json:"seed"`
		GoClassifier struct {
			LearningRates   []float64                                `json:"learning_rates"`
			Epochs          []int                                    `json:"epochs"`
			BatchSize       int                                      `json:"batch_size"`
			WeightDecay     float64                                  `json:"weight_decay"`
			FeatureBins     int                                      `json:"feature_bins"`
			TemperatureGrid struct{ Minimum, Maximum, Step float64 } `json:"temperature_grid"`
		} `json:"go_classifier"`
		GoSources map[string]string `json:"go_sources"`
	}
	if e = json.Unmarshal(planBytes, &contract); e != nil {
		return e
	}
	if contract.Schema != "riido-statehint-training-plan-v1" || contract.DataSHA != dataSHA || contract.Confidence != .9 || contract.Seed != 1729 {
		return fmt.Errorf("training plan mismatch")
	}
	goPlan := contract.GoClassifier
	if len(goPlan.LearningRates) != 2 || goPlan.LearningRates[0] != .01 || goPlan.LearningRates[1] != .03 || len(goPlan.Epochs) != 2 || goPlan.Epochs[0] != 20 || goPlan.Epochs[1] != 40 || goPlan.BatchSize != 32 || goPlan.WeightDecay != .001 || goPlan.FeatureBins != 1024 || goPlan.TemperatureGrid.Minimum != .5 || goPlan.TemperatureGrid.Maximum != 5 || goPlan.TemperatureGrid.Step != .1 {
		return fmt.Errorf("Go classifier plan mismatch")
	}
	if len(contract.GoSources) < 10 {
		return fmt.Errorf("missing Go source pins")
	}
	for path, pin := range contract.GoSources {
		if filepath.IsAbs(path) || filepath.Clean(path) != path || strings.HasPrefix(path, "..") || len(pin) != 64 {
			return fmt.Errorf("invalid source pin")
		}
		b, e := os.ReadFile(path)
		if e != nil || hash(b) != pin {
			return fmt.Errorf("Go source pin mismatch")
		}
	}
	if e = os.Mkdir(*out, 0700); e != nil {
		return e
	}
	started := time.Now()
	trainRows, valRows, calRows := subset(rows, "train", ""), subset(rows, "validation", ""), subset(rows, "calibration", "")
	if len(trainRows) != 1200 || len(valRows) != 400 || len(calRows) != 400 || len(subset(rows, "test", "")) != 400 {
		return fmt.Errorf("split counts changed")
	}
	samples := make([]statehint.Sample, len(trainRows))
	for i, r := range trainRows {
		samples[i] = statehint.Sample{Text: r.Text, Label: r.Intent}
	}
	base := statehint.NewModel()
	var best *statehint.Model
	bestNLL := math.Inf(1)
	var trials []trial
	for _, lr := range []float64{.01, .03} {
		for _, epochs := range []int{20, 40} {
			m := base.Clone()
			fit, e := m.Fit(samples, statehint.FitOptions{Epochs: epochs, BatchSize: 32, LearningRate: lr, WeightDecay: .001, Seed: 1729})
			if e != nil {
				return e
			}
			v, _, e := evaluate(valRows, predictor(m))
			if e != nil {
				return e
			}
			sha, e := save(filepath.Join(*out, fmt.Sprintf("candidate-%g-%d.rsh", lr, epochs)), m)
			if e != nil {
				return e
			}
			t := trial{lr, epochs, v, fit, sha}
			trials = append(trials, t)
			if e = json.NewEncoder(os.Stdout).Encode(t); e != nil {
				return e
			}
			if v.NLL < bestNLL {
				bestNLL = v.NLL
				best = m.Clone()
			}
		}
	}
	bestT, bestCal := 1., math.Inf(1)
	for i := 0; i <= 45; i++ {
		t := .5 + float64(i)*.1
		m := best.Clone()
		if e = m.SetTemperature(t); e != nil {
			return e
		}
		v, _, e := evaluate(calRows, predictor(m))
		if e != nil {
			return e
		}
		if v.NLL < bestCal {
			bestT, bestCal = t, v.NLL
		}
	}
	if e = best.SetTemperature(bestT); e != nil {
		return e
	}
	modelSHA, e := save(filepath.Join(*out, "statehint.rsh"), best)
	if e != nil {
		return e
	}
	// Final model and temperature are locked before the first test prediction.
	locked := struct {
		ModelSHA, DataSHA, PlanSHA string
		Temperature                float64
	}{modelSHA, dataSHA, hash(planBytes), bestT}
	if e = writeJSON(filepath.Join(*out, "TEST-LOCK.json"), locked); e != nil {
		return e
	}
	metrics := make(map[string]metric)
	var predictions []resultRow
	for _, split := range []string{"validation", "calibration", "test"} {
		for _, locale := range []string{"", "ko", "en"} {
			rs := subset(rows, split, locale)
			key := split + "/" + locale
			m, p, e := evaluate(rs, predictor(best))
			if e != nil {
				return e
			}
			metrics["trained/"+key] = m
			rule, _, e := evaluate(rs, statehint.RuleBaseline)
			if e != nil {
				return e
			}
			metrics["rule/"+key] = rule
			if locale == "" {
				predictions = append(predictions, p...)
			}
		}
	}
	// Serialization parity covers every original input, including both languages.
	b, e := os.ReadFile(filepath.Join(*out, "statehint.rsh"))
	if e != nil {
		return e
	}
	loaded, e := statehint.Load(bytes.NewReader(b))
	if e != nil {
		return e
	}
	left, right := predictor(best), predictor(loaded)
	for _, r := range rows {
		a, e := left(r.Text)
		if e != nil {
			return e
		}
		z, e := right(r.Text)
		if e != nil {
			return e
		}
		if a != z {
			return fmt.Errorf("reload prediction mismatch")
		}
	}
	var workspace statehint.Workspace
	var sink statehint.Prediction
	benchText := "자료를 확인한 후 다음 작업을 시작하고 있습니다. FictionalProject-00 FictionalItem-00"
	for i := 0; i < 32; i++ {
		sink, e = loaded.Predict(benchText, &workspace)
		if e != nil {
			return e
		}
	}
	latencies := make([]int64, 2000)
	for i := range latencies {
		start := time.Now()
		sink, e = loaded.Predict(benchText, &workspace)
		if e != nil {
			return e
		}
		latencies[i] = time.Since(start).Nanoseconds()
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	allocs := testing.AllocsPerRun(100, func() { sink, _ = loaded.Predict(benchText, &workspace) })
	_ = sink
	report := struct {
		Schema              string            `json:"schema"`
		DataSHA             string            `json:"data_sha256"`
		PlanSHA             string            `json:"plan_sha256"`
		ModelSHA            string            `json:"model_sha256"`
		ModelBytes          int               `json:"model_bytes"`
		TrainedParameters   int               `json:"trained_parameters"`
		Trials              []trial           `json:"trials"`
		Temperature         float64           `json:"temperature"`
		Metrics             map[string]metric `json:"metrics"`
		Predictions         []resultRow       `json:"predictions"`
		ElapsedNS           int64             `json:"training_evaluation_elapsed_ns"`
		Bench               any               `json:"local_warm_benchmark"`
		ReloadParityRows    int               `json:"reload_parity_rows"`
		DeploymentQualified bool              `json:"deployment_qualified"`
		Scope               string            `json:"scope"`
	}{"riido-statehint-go-development-v1", dataSHA, hash(planBytes), modelSHA, len(b), 1024*8 + 8, trials, bestT, metrics, predictions, time.Since(started).Nanoseconds(), struct {
		N                                      int
		P50NS, P95NS                           int64
		AllocsPerCall                          float64
		ModelStructBytes, WorkspaceStructBytes uintptr
		Scope                                  string
	}{2000, latencies[1000], latencies[1900], allocs, unsafe.Sizeof(*loaded), unsafe.Sizeof(workspace), "One process-warm short text; full feature extraction and scoring, excludes model load, policy plan and process startup. No cross-machine or GPU claim."}, len(rows), false, "Original bilingual synthetic development only:2400 variants,96 template families and48 translated concept pairs. This is supervised Go classification, not Laya fine-tuning, product truth,2400 independent cases or automatic Riido mutation. Rules and calibrated learned probabilities are separate controls."}
	if e = writeJSON(filepath.Join(*out, "report.json"), report); e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Status, ModelSHA string
		ModelBytes       int
		Test             metric
		Temperature      float64
	}{"development_complete", modelSHA, len(b), metrics["trained/test/"], bestT})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "riido-statehint-train:", e)
		os.Exit(1)
	}
}
