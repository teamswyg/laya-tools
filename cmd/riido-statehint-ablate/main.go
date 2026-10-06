// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only feature ablation on previously exposed validation data.
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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	rubricSHA = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	anchor    = ".cache/statehint-feature-ablation"
)

var errAblation = errors.New("development ablation input, checksum or output invalid")

type pin struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}

type input struct {
	pin     pin
	bytes   []byte
	corpus  statehintcorpus.Corpus
	summary statehintcorpus.Summary
}

type recipe struct {
	Epochs       int        `json:"epochs"`
	Batch        int        `json:"batch_size"`
	Rate         float64    `json:"learning_rate"`
	Decay        float64    `json:"weight_decay"`
	Seed         int64      `json:"seed"`
	Temperature  float64    `json:"temperature"`
	Gate         [2]float64 `json:"confidence_margin"`
	Modes        [2]string  `json:"fresh_modes"`
	FeatureBins  int        `json:"feature_bins"`
	BaselineSize int        `json:"v1_reference_artifact_bytes_not_loaded"`
}

func fixedRecipe() recipe {
	return recipe{40, 32, .02, .001, 1729, 1, [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, [2]string{"simple2048", "contextual2048"}, statehintwide.FeatureBins, statehint.ArtifactBytes}
}

func fitOptions() statehintwide.FitOptions {
	r := fixedRecipe()
	return statehintwide.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed}
}

func digest(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func validPin(p pin) bool {
	b, err := hex.DecodeString(p.SHA)
	return local(p.Path) && err == nil && len(b) == sha256.Size && hex.EncodeToString(b) == p.SHA
}

func local(name string) bool {
	return filepath.IsLocal(name) && filepath.Clean(name) == name && name != "." && !strings.Contains(name, "\\")
}

// Read only the caller's two explicitly pinned regular files, within a budget.
// There is no manifest, calibration, final-test or parent-model input.
func readPin(root *os.Root, p pin, limit int64) ([]byte, error) {
	if !validPin(p) {
		return nil, errAblation
	}
	before, err := root.Lstat(p.Path)
	if err != nil || !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 || before.Size() < 1 || before.Size() > limit {
		return nil, errAblation
	}
	f, err := root.Open(p.Path)
	if err != nil {
		return nil, errAblation
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !os.SameFile(before, after) || !after.Mode().IsRegular() {
		return nil, errAblation
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errAblation
	}
	return b, nil
}

func loadInput(root *os.Root, p pin, partition string, families int) (input, error) {
	b, err := readPin(root, p, statehintcorpus.MaxFileBytes)
	if err != nil {
		return input{}, err
	}
	c, s, err := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if err != nil || s.Families != families || s.Rows != 2*families {
		return input{}, errAblation
	}
	for _, count := range s.IntentFamilies {
		if count != families/statehint.IntentCount {
			return input{}, errAblation
		}
	}
	return input{p, b, c, s}, nil
}

func disjoint(a, b input) error {
	metadata := make([]statehintcorpus.Family, 0, a.summary.Families+b.summary.Families)
	ids := make([]string, 0, a.summary.Rows+b.summary.Rows)
	texts := make([]string, 0, cap(ids))
	for _, in := range []input{a, b} {
		for _, pair := range in.corpus.Pairs() {
			metadata = append(metadata, pair.Family)
		}
		for _, r := range in.corpus.Rows() {
			ids = append(ids, r.ID)
			texts = append(texts, digest([]byte(r.Text)))
		}
	}
	for _, values := range [][]string{ids, texts} {
		sort.Strings(values)
		for i := 1; i < len(values); i++ {
			if values[i-1] == values[i] {
				return errAblation
			}
		}
	}
	if _, err := statehintcorpus.ValidateMetadata(metadata); err != nil {
		return errAblation
	}
	return nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errAblation
	}
	for _, dir := range []string{".cache", anchor} {
		if err := root.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, errAblation
		}
		st, err := root.Lstat(dir)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || dir == anchor && st.Mode().Perm() != 0700 {
			return nil, errAblation
		}
	}
	if err := root.Mkdir(name, 0700); err != nil {
		return nil, errAblation
	}
	return root.OpenRoot(name)
}

func writeBytes(root *os.Root, name string, b []byte) error {
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errAblation
	}
	n, err := f.Write(b)
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errAblation
	}
	return nil
}

func writeJSON(root *os.Root, name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errAblation
	}
	return writeBytes(root, name, append(b, '\n'))
}

type scoredRow struct {
	ID         string               `json:"id"`
	Family     string               `json:"family_id"`
	Locale     string               `json:"locale"`
	Expected   statehint.Intent     `json:"expected_intent"`
	Prediction statehint.Prediction `json:"prediction"`
}

type armReport struct {
	Mode             string                  `json:"mode"`
	Fit              statehintwide.FitReport `json:"fit"`
	Validation       statehintfamily.Report  `json:"exposed_validation_development"`
	Model            pin                     `json:"model"`
	ArtifactBytes    int                     `json:"artifact_bytes"`
	FitNS            int64                   `json:"fit_nanoseconds"`
	ValidationNS     int64                   `json:"validation_prediction_nanoseconds"`
	ReloadParity     bool                    `json:"trained_save_load_exact_parity"`
	ValidationCalls  int                     `json:"validation_forwards_including_reload_check"`
	ProbabilityOrder [8]statehint.Intent     `json:"probability_intent_order"`
}

type result struct {
	Schema            string      `json:"schema"`
	Status            string      `json:"status"`
	DevelopmentOnly   bool        `json:"development_only"`
	ValidationExposed bool        `json:"validation_previously_exposed"`
	HumanTruth        bool        `json:"human_truth"`
	Recipe            recipe      `json:"fixed_recipe"`
	Train             pin         `json:"train_input"`
	Validation        pin         `json:"validation_input"`
	Arms              []armReport `json:"arms"`
	CalibrationCalls  int         `json:"calibration_forwards"`
	TestCalls         int         `json:"test_forwards"`
	Selected          bool        `json:"model_selection_performed"`
	Promoted          bool        `json:"promotion_performed"`
	Qualified         bool        `json:"deployment_qualified"`
	GoVersion         string      `json:"go_version"`
}

func predict(m *statehintwide.Model, c statehintcorpus.Corpus) ([]statehint.Prediction, error) {
	rows := c.Rows()
	p := make([]statehint.Prediction, len(rows))
	var w statehintwide.Workspace
	for i, r := range rows {
		v, err := m.Predict(r.Text, &w)
		if err != nil {
			return nil, errAblation
		}
		p[i] = v
	}
	return p, nil
}

func fitArm(out *os.Root, mode statehintwide.FeatureMode, name string, samples []statehintwide.Sample, validation statehintcorpus.Corpus) (armReport, error) {
	report := armReport{Mode: name, ProbabilityOrder: statehint.Intents()}
	m := statehintwide.NewModel(mode)
	started := time.Now()
	fit, err := m.Fit(samples, fitOptions())
	report.FitNS, report.Fit = time.Since(started).Nanoseconds(), fit
	if err != nil || m.Temperature() != 1 {
		return report, errAblation
	}
	started = time.Now()
	p, err := predict(m, validation)
	report.ValidationNS = time.Since(started).Nanoseconds()
	if err != nil {
		return report, err
	}
	report.Validation, err = statehintfamily.Evaluate(validation, p)
	if err != nil {
		return report, errAblation
	}
	rows := validation.Rows()
	scores := make([]scoredRow, len(rows))
	for i, r := range rows {
		scores[i] = scoredRow{r.ID, r.Family, r.Locale, r.Expected, p[i]}
	}
	if err := writeJSON(out, name+".predictions.json", scores); err != nil {
		return report, err
	}
	var artifact bytes.Buffer
	if err := m.Save(&artifact); err != nil {
		return report, errAblation
	}
	report.Model = pin{name + ".rsh", digest(artifact.Bytes())}
	report.ArtifactBytes = artifact.Len()
	if err := writeBytes(out, report.Model.Path, artifact.Bytes()); err != nil {
		return report, err
	}
	saved, err := readPin(out, report.Model, statehintwide.ArtifactBytes)
	if err != nil {
		return report, err
	}
	loaded, err := statehintwide.Load(bytes.NewReader(saved))
	if err != nil {
		return report, errAblation
	}
	q, err := predict(loaded, validation)
	if err != nil || len(q) != len(p) || loaded.Temperature() != 1 || loaded.TrainingSteps() != m.TrainingSteps() {
		return report, errAblation
	}
	for i := range p {
		if p[i] != q[i] {
			return report, errAblation
		}
	}
	report.ReloadParity, report.ValidationCalls = true, 2*len(p)
	return report, writeJSON(out, name+".report.json", report)
}

func fitAll(out *os.Root, train, validation input) (r result, err error) {
	r = result{Schema: "riido-statehint-feature-ablation-development-v1", Status: "started", DevelopmentOnly: true, ValidationExposed: true, Recipe: fixedRecipe(), Train: train.pin, Validation: validation.pin, Arms: []armReport{}, GoVersion: runtime.Version()}
	defer func() {
		if err != nil {
			r.Status = "failed_development_ablation; partial artifacts retained"
			if e := writeJSON(out, "FAILED.json", r); e != nil {
				err = e
			}
		}
	}()
	if err = writeJSON(out, "recipe.json", r); err != nil {
		return r, err
	}
	for _, item := range []struct {
		name string
		data []byte
	}{{"train.source.jsonl", train.bytes}, {"validation.source.jsonl", validation.bytes}} {
		if err = writeBytes(out, item.name, item.data); err != nil {
			return r, err
		}
	}
	rows := train.corpus.Rows()
	samples := make([]statehintwide.Sample, len(rows))
	for i, row := range rows {
		samples[i] = statehintwide.Sample{Text: row.Text, Label: row.Expected}
	}
	for i, mode := range []statehintwide.FeatureMode{statehintwide.Simple, statehintwide.Contextual} {
		arm, e := fitArm(out, mode, r.Recipe.Modes[i], samples, validation.corpus)
		r.Arms = append(r.Arms, arm)
		if e != nil {
			return r, e
		}
	}
	r.Status = "completed_development_ablation; no selection or promotion"
	err = writeJSON(out, "report.json", r)
	return r, err
}

func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-ablate", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	trainName := f.String("train", "", "local original train840 JSONL")
	trainSHA := f.String("train-sha256", "", "exact lowercase input SHA-256")
	validationName := f.String("validation", "", "local exposed validation120 JSONL")
	validationSHA := f.String("validation-sha256", "", "exact lowercase input SHA-256")
	output := f.String("out", "", "new .cache/statehint-feature-ablation/RUN directory")
	check := f.Bool("check", false, "validate only; no model calls or outputs")
	fit := f.Bool("fit", false, "explicit fixed two-mode development comparison")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(errOut, "riido-statehint-ablate --train FILE --train-sha256 SHA --validation FILE --validation-sha256 SHA --check\nUse --fit --out .cache/statehint-feature-ablation/NEW-RUN for the fixed two-mode comparison. Exposed validation is development data. No calibration, test, selection or promotion.")
			return err
		}
		return errAblation
	}
	a, b := pin{*trainName, *trainSHA}, pin{*validationName, *validationSHA}
	if f.NArg() != 0 || *check == *fit || !validPin(a) || !validPin(b) || a.Path == b.Path || *check && *output != "" || *fit && (!local(*output) || filepath.Dir(*output) != anchor) {
		return errAblation
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return errAblation
	}
	defer root.Close()
	ai, ea := root.Lstat(a.Path)
	bi, eb := root.Lstat(b.Path)
	if ea != nil || eb != nil || os.SameFile(ai, bi) {
		return errAblation
	}
	train, err := loadInput(root, a, "train", 840)
	if err != nil {
		return err
	}
	validation, err := loadInput(root, b, "validation", 120)
	if err != nil || disjoint(train, validation) != nil {
		return errAblation
	}
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status            string `json:"status"`
			DevelopmentOnly   bool   `json:"development_only"`
			ValidationExposed bool   `json:"validation_previously_exposed"`
		}{"checked_train840_exposed_validation120; no fitting or model calls", true, true})
	}
	private, err := privateOutput(root, *output)
	if err != nil {
		return err
	}
	defer private.Close()
	r, err := fitAll(private, train, validation)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(struct {
		Status            string    `json:"status"`
		Modes             int       `json:"fresh_modes"`
		ValidationExposed bool      `json:"validation_previously_exposed"`
		SeverityCosts     []float64 `json:"development_severity_costs"`
	}{r.Status, len(r.Arms), true, []float64{r.Arms[0].Validation.SeverityCost, r.Arms[1].Validation.SeverityCost}})
}

func main() {
	runtime.GOMAXPROCS(2)
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "feature ablation failed; check pinned train/validation inputs and fresh private output")
		os.Exit(1)
	}
}
