// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only, fixed architecture comparison on already exposed development data.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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
	"github.com/teamswyg/laya-tools/pkg/statehintmlp"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	rubricSHA      = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	trainSHA       = "586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e"
	splitSHA       = "f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9"
	anchor         = ".cache/statehint-capacity"
	metadataBudget = 1 << 20
)

var errCapacity = errors.New("capacity comparison input, checksum or output invalid")

type pin struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}
type input struct {
	pin    pin
	bytes  []byte
	corpus statehintcorpus.Corpus
}
type metadataInput struct {
	pin   pin
	bytes []byte
}
type assignment struct {
	Family string `json:"family_id"`
	Group  string `json:"leakage_group_id"`
	Use    string `json:"internal_split"`
}
type splitMetadata struct {
	Schema      string       `json:"schema"`
	Families    int          `json:"families"`
	FitFamilies int          `json:"fitFamilies"`
	DevFamilies int          `json:"devFamilies"`
	Assignments []assignment `json:"assignments"`
}
type draw struct {
	Family   string           `json:"family_id"`
	Group    string           `json:"leakage_group_id"`
	Expected statehint.Intent `json:"expected_intent"`
	RowIDs   [2]string        `json:"row_ids_ko_en"`
	Rank     string           `json:"sha256_rank"`
}
type prepared struct {
	Train      input
	Split      metadataInput
	Validation *input
	Fit, Dev   statehintcorpus.Corpus
	Samples    []statehintwide.Sample
	Draws      []draw
}
type recipe struct {
	Epochs        int        `json:"epochs"`
	Batch         int        `json:"batch_size"`
	Rate          float64    `json:"learning_rate"`
	Decay         float64    `json:"weight_decay"`
	Seed          int64      `json:"seed"`
	Temperature   float64    `json:"temperature"`
	Gate          [2]float64 `json:"confidence_margin"`
	ArmOrder      [2]string  `json:"fresh_arm_order"`
	RowsPerArm    int        `json:"rows_per_arm"`
	UpdatesPerArm int        `json:"updates_per_arm"`
	DrawFamilies  [8]int     `json:"extra_families_per_intent_persisted_order"`
}

func classCounts() [8]int { return [8]int{0, 8, 8, 8, 32, 0, 8, 0} }
func fixedRecipe() recipe {
	return recipe{40, 32, .02, .001, 1729, 1, [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, [2]string{"fresh_contextual_linear", "fresh_contextual_mlp16"}, 1454, 1840, classCounts()}
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func local(name string) bool {
	return filepath.IsLocal(name) && filepath.Clean(name) == name && name != "." && !strings.Contains(name, "\\")
}
func validPin(p pin) bool {
	b, e := hex.DecodeString(p.SHA)
	return local(p.Path) && e == nil && len(b) == sha256.Size && hex.EncodeToString(b) == p.SHA
}
func readPin(root *os.Root, p pin, limit int64) ([]byte, error) {
	if !validPin(p) || limit < 1 {
		return nil, errCapacity
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > limit {
		return nil, errCapacity
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errCapacity
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errCapacity
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errCapacity
	}
	return b, nil
}
func decodeMetadata(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if d.Decode(v) != nil {
		return errCapacity
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errCapacity
	}
	return nil
}
func balanced(n int) (counts [8]int) {
	for i := range counts {
		counts[i] = n
	}
	return
}
func loadInput(root *os.Root, p pin, partition string, families int, counts [8]int) (input, error) {
	b, e := readPin(root, p, statehintcorpus.MaxFileBytes)
	if e != nil {
		return input{}, e
	}
	c, s, e := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if e != nil || s.Families != families || s.Rows != 2*families || s.IntentFamilies != counts {
		return input{}, errCapacity
	}
	return input{p, b, c}, nil
}

// Preserve the frozen controlled-run group split and whole-family draw exactly.
func splitUse(group string) string {
	s := sha256.Sum256([]byte("completion-contrast-internal-split-1729:" + group))
	if binary.BigEndian.Uint32(s[:4])%5 == 0 {
		return "dev"
	}
	return "fit"
}
func corpusFromRows(rows []statehintcorpus.Row) (statehintcorpus.Corpus, error) {
	var b bytes.Buffer
	for _, r := range rows {
		if r.AssertionForms == nil {
			r.AssertionForms = []string{}
		}
		if json.NewEncoder(&b).Encode(r) != nil {
			return statehintcorpus.Corpus{}, errCapacity
		}
	}
	c, _, e := statehintcorpus.Read(&b, statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		return c, errCapacity
	}
	return c, nil
}
func splitCorpus(c statehintcorpus.Corpus, s splitMetadata) (statehintcorpus.Corpus, statehintcorpus.Corpus, error) {
	if s.Schema != "statehint-completion-contrast-train-only-group-split-v1" || s.Families != 840 || s.FitFamilies != 663 || s.DevFamilies != 177 || len(s.Assignments) != 840 {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errCapacity
	}
	a := append([]assignment(nil), s.Assignments...)
	sort.Slice(a, func(i, j int) bool { return a[i].Family < a[j].Family })
	pairs, rows := c.Pairs(), c.Rows()
	if len(pairs) != len(a) {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errCapacity
	}
	var fit, dev []statehintcorpus.Row
	for i, p := range pairs {
		x := a[i]
		if x.Family != p.Family.ID || x.Group != p.Family.Lineage || x.Use != splitUse(x.Group) {
			return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errCapacity
		}
		target := &fit
		if x.Use == "dev" {
			target = &dev
		}
		for _, index := range p.Rows {
			*target = append(*target, rows[index])
		}
	}
	if len(fit) != 1326 || len(dev) != 354 {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errCapacity
	}
	f, e := corpusFromRows(fit)
	if e != nil {
		return f, statehintcorpus.Corpus{}, e
	}
	d, e := corpusFromRows(dev)
	return f, d, e
}
func samples(c statehintcorpus.Corpus) []statehintwide.Sample {
	rows := c.Rows()
	v := make([]statehintwide.Sample, 0, len(rows))
	for _, p := range c.Pairs() {
		for _, i := range p.Rows {
			v = append(v, statehintwide.Sample{Text: rows[i].Text, Label: rows[i].Expected})
		}
	}
	return v
}
func controlDraws(c statehintcorpus.Corpus) ([]statehintwide.Sample, []draw, error) {
	rows, pairs := c.Rows(), c.Pairs()
	var extra []statehintwide.Sample
	var draws []draw
	for class, need := range classCounts() {
		if need == 0 {
			continue
		}
		label := statehint.Intents()[class]
		type rankedPair struct {
			pair statehintcorpus.Pair
			rank string
		}
		var ranked []rankedPair
		for _, p := range pairs {
			if p.Family.Expected == label {
				ranked = append(ranked, rankedPair{p, digest([]byte("completion-contrast-class-matched-control-1729:" + p.Family.ID))})
			}
		}
		if len(ranked) == 0 {
			return nil, nil, errCapacity
		}
		sort.Slice(ranked, func(i, j int) bool {
			if ranked[i].rank != ranked[j].rank {
				return ranked[i].rank < ranked[j].rank
			}
			return ranked[i].pair.Family.ID < ranked[j].pair.Family.ID
		})
		for i := 0; i < need; i++ {
			chosen := ranked[i%len(ranked)]
			p := chosen.pair
			d := draw{Family: p.Family.ID, Group: p.Family.Lineage, Expected: label, Rank: chosen.rank}
			for locale, index := range p.Rows {
				r := rows[index]
				d.RowIDs[locale] = r.ID
				extra = append(extra, statehintwide.Sample{Text: r.Text, Label: r.Expected})
			}
			draws = append(draws, d)
		}
	}
	return extra, draws, nil
}
func disjoint(train, validation input) error {
	var metadata []statehintcorpus.Family
	var ids, texts []string
	for _, in := range []input{train, validation} {
		for _, p := range in.corpus.Pairs() {
			metadata = append(metadata, p.Family)
		}
		for _, r := range in.corpus.Rows() {
			ids = append(ids, r.ID)
			texts = append(texts, digest([]byte(r.Text)))
		}
	}
	for _, v := range [][]string{ids, texts} {
		sort.Strings(v)
		for i := 1; i < len(v); i++ {
			if v[i-1] == v[i] {
				return errCapacity
			}
		}
	}
	if _, e := statehintcorpus.ValidateMetadata(metadata); e != nil {
		return errCapacity
	}
	return nil
}
func prepare(train input, split metadataInput, validation *input) (prepared, error) {
	p := prepared{Train: train, Split: split, Validation: validation}
	var s splitMetadata
	if decodeMetadata(split.bytes, &s) != nil {
		return prepared{}, errCapacity
	}
	var e error
	p.Fit, p.Dev, e = splitCorpus(train.corpus, s)
	if e != nil {
		return prepared{}, e
	}
	if validation != nil && disjoint(train, *validation) != nil {
		return prepared{}, errCapacity
	}
	extra, draws, e := controlDraws(p.Fit)
	if e != nil {
		return prepared{}, e
	}
	p.Draws = draws
	p.Samples = append(samples(p.Fit), extra...)
	if len(p.Samples) != 1454 || len(draws) != 64 {
		return prepared{}, errCapacity
	}
	return p, nil
}
func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errCapacity
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errCapacity
		}
		s, e := root.Lstat(dir)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || dir == anchor && s.Mode().Perm() != 0700 {
			return nil, errCapacity
		}
	}
	if e := root.Mkdir(name, 0700); e != nil {
		return nil, errCapacity
	}
	return root.OpenRoot(name)
}
func writeBytes(root *os.Root, name string, b []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errCapacity
	}
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil || ce != nil {
		return errCapacity
	}
	return nil
}
func writeJSON(root *os.Root, name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return errCapacity
	}
	return writeBytes(root, name, append(b, '\n'))
}

type armKind int

const (
	linearArm armKind = iota
	mlpArm
)

type fitReport struct {
	Samples       int     `json:"samples"`
	Epochs        int     `json:"epochs"`
	Batches       int     `json:"batches"`
	TrainingSteps uint64  `json:"training_steps"`
	MeanLoss      float64 `json:"mean_loss"`
}

// Different workspace types stay caller-owned; Prediction is the SDK type alias.
type armModel struct {
	linear     *statehintwide.Model
	mlp        *statehintmlp.Model
	linearWork statehintwide.Workspace
	mlpWork    statehintmlp.Workspace
}

func (m *armModel) predict(text string) (statehint.Prediction, error) {
	if m.linear != nil {
		return m.linear.Predict(text, &m.linearWork)
	}
	return m.mlp.Predict(text, &m.mlpWork)
}
func (m *armModel) save(w io.Writer) error {
	if m.linear != nil {
		return m.linear.Save(w)
	}
	return m.mlp.Save(w)
}
func (m *armModel) temperature() float64 {
	if m.linear != nil {
		return m.linear.Temperature()
	}
	return m.mlp.Temperature()
}
func (m *armModel) steps() uint64 {
	if m.linear != nil {
		return m.linear.TrainingSteps()
	}
	return m.mlp.TrainingSteps()
}
func fitModel(kind armKind, training []statehintwide.Sample) (*armModel, fitReport, error) {
	r := fixedRecipe()
	m := &armModel{}
	if kind == linearArm {
		m.linear = statehintwide.NewModel(statehintwide.Contextual)
		f, e := m.linear.Fit(training, statehintwide.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed})
		return m, fitReport{f.Samples, f.Epochs, f.Batches, f.TrainingSteps, f.MeanLoss}, e
	}
	if kind != mlpArm {
		return nil, fitReport{}, errCapacity
	}
	m.mlp = statehintmlp.NewModel()
	typed := make([]statehintmlp.Sample, len(training))
	for i, s := range training {
		typed[i] = statehintmlp.Sample{Text: s.Text, Label: s.Label}
	}
	f, e := m.mlp.Fit(typed, statehintmlp.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed})
	return m, fitReport{f.Samples, f.Epochs, f.Batches, f.TrainingSteps, f.MeanLoss}, e
}
func loadModel(kind armKind, b []byte) (*armModel, error) {
	m := &armModel{}
	var e error
	if kind == linearArm {
		m.linear, e = statehintwide.Load(bytes.NewReader(b))
	} else if kind == mlpArm {
		m.mlp, e = statehintmlp.Load(bytes.NewReader(b))
	} else {
		return nil, errCapacity
	}
	return m, e
}

type scoredRow struct {
	ID         string               `json:"id"`
	Family     string               `json:"family_id"`
	Locale     string               `json:"locale"`
	Expected   statehint.Intent     `json:"expected_intent"`
	Prediction statehint.Prediction `json:"prediction"`
}
type armReport struct {
	Name             string                  `json:"arm"`
	Initialization   string                  `json:"initialization"`
	Seed             int64                   `json:"initialization_seed"`
	ArtifactFormat   string                  `json:"artifact_format"`
	Loader           string                  `json:"loader"`
	Parameters       int                     `json:"parameters"`
	Fit              fitReport               `json:"fit"`
	Dev              statehintfamily.Report  `json:"internal_dev"`
	Diagnostic       *statehintfamily.Report `json:"exposed_validation_diagnostic,omitempty"`
	Model            pin                     `json:"model"`
	ArtifactBytes    int                     `json:"artifact_bytes"`
	FitNS            int64                   `json:"fit_nanoseconds"`
	DevNS            int64                   `json:"internal_dev_prediction_nanoseconds"`
	DiagnosticNS     int64                   `json:"diagnostic_prediction_nanoseconds"`
	ReloadParity     bool                    `json:"trained_save_load_exact_parity"`
	DevCalls         int                     `json:"internal_dev_forwards_including_reload_check"`
	DiagnosticCalls  int                     `json:"exposed_validation_forwards_including_reload_check"`
	GoHeapAfterFit   uint64                  `json:"go_heap_alloc_after_fit_bytes_not_os_rss"`
	ProbabilityOrder [8]statehint.Intent     `json:"probability_intent_order"`
}
type result struct {
	Schema                       string      `json:"schema"`
	Status                       string      `json:"status"`
	DevelopmentOnly              bool        `json:"development_only"`
	HumanTruth                   bool        `json:"human_truth"`
	InternalDevPreviouslyExposed bool        `json:"internal_dev_previously_exposed"`
	FreshQualification           bool        `json:"fresh_qualification"`
	ComparisonScope              string      `json:"comparison_scope"`
	Recipe                       recipe      `json:"fixed_recipe"`
	Train                        pin         `json:"train_input"`
	Split                        pin         `json:"frozen_split_input"`
	Validation                   *pin        `json:"optional_exposed_validation_input,omitempty"`
	DrawSHA                      string      `json:"whole_family_draws_sha256"`
	TrainingSamplesSHA           string      `json:"ordered_training_samples_receipt_sha256"`
	Arms                         []armReport `json:"arms"`
	SelectedArm                  string      `json:"internal_dev_selected_arm"`
	SelectionSource              string      `json:"selection_source"`
	Promotion                    bool        `json:"promotion_performed"`
	DeploymentQualified          bool        `json:"deployment_qualified"`
	CalibrationCalls             int         `json:"calibration_forwards"`
	TestCalls                    int         `json:"test_forwards"`
	ValidationSelectionWeight    int         `json:"exposed_validation_selection_weight"`
	GoVersion                    string      `json:"go_version"`
	RSSNote                      string      `json:"resource_measurement_note"`
}

func predict(m *armModel, c statehintcorpus.Corpus) ([]statehint.Prediction, error) {
	rows := c.Rows()
	p := make([]statehint.Prediction, len(rows))
	for i, r := range rows {
		v, e := m.predict(r.Text)
		if e != nil {
			return nil, errCapacity
		}
		p[i] = v
	}
	return p, nil
}
func evaluateAndSave(out *os.Root, name string, m *armModel, c statehintcorpus.Corpus) (statehintfamily.Report, []statehint.Prediction, int64, error) {
	started := time.Now()
	p, e := predict(m, c)
	duration := time.Since(started).Nanoseconds()
	if e != nil {
		return statehintfamily.Report{}, nil, duration, e
	}
	r, e := statehintfamily.Evaluate(c, p)
	if e != nil {
		return r, nil, duration, errCapacity
	}
	rows := c.Rows()
	scores := make([]scoredRow, len(rows))
	for i, row := range rows {
		scores[i] = scoredRow{row.ID, row.Family, row.Locale, row.Expected, p[i]}
	}
	return r, p, duration, writeJSON(out, name+".predictions.json", scores)
}
func parity(m *armModel, c statehintcorpus.Corpus, want []statehint.Prediction) bool {
	got, e := predict(m, c)
	if e != nil || len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
func fitArm(out *os.Root, kind armKind, training []statehintwide.Sample, dev statehintcorpus.Corpus, validation *input) (armReport, error) {
	if kind != linearArm && kind != mlpArm {
		return armReport{}, errCapacity
	}
	r := armReport{Name: fixedRecipe().ArmOrder[kind], Seed: 1729, ProbabilityOrder: statehint.Intents(), Initialization: "zero-initialized-linear", ArtifactFormat: "RSH v2", Loader: "pkg/statehintwide.Load", Parameters: statehintwide.FeatureBins*statehintwide.IntentCount + statehintwide.IntentCount}
	suffix, expectedBytes := ".rsh", statehintwide.ArtifactBytes
	if kind == mlpArm {
		r.Initialization = statehintmlp.Initialization
		r.ArtifactFormat = "RSM v3"
		r.Loader = "pkg/statehintmlp.Load"
		r.Parameters = statehintmlp.ParameterCount
		suffix = ".rsm"
		expectedBytes = statehintmlp.ArtifactBytes
	}
	started := time.Now()
	m, f, e := fitModel(kind, training)
	r.Fit, r.FitNS = f, time.Since(started).Nanoseconds()
	if e != nil || m.temperature() != 1 {
		return r, errCapacity
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r.GoHeapAfterFit = memory.HeapAlloc
	report, p, ns, e := evaluateAndSave(out, r.Name+".internal-dev", m, dev)
	r.Dev, r.DevNS = report, ns
	if e != nil {
		return r, e
	}
	var q []statehint.Prediction
	if validation != nil {
		report, scores, ns, e := evaluateAndSave(out, r.Name+".exposed-validation", m, validation.corpus)
		if e != nil {
			return r, e
		}
		r.Diagnostic, r.DiagnosticNS = &report, ns
		q = scores
	}
	var artifact bytes.Buffer
	if m.save(&artifact) != nil || artifact.Len() != expectedBytes {
		return r, errCapacity
	}
	r.Model = pin{r.Name + suffix, digest(artifact.Bytes())}
	r.ArtifactBytes = artifact.Len()
	if writeBytes(out, r.Model.Path, artifact.Bytes()) != nil {
		return r, errCapacity
	}
	saved, e := readPin(out, r.Model, int64(expectedBytes))
	if e != nil {
		return r, e
	}
	loaded, e := loadModel(kind, saved)
	if e != nil || loaded.temperature() != 1 || loaded.steps() != m.steps() || !parity(loaded, dev, p) || validation != nil && !parity(loaded, validation.corpus, q) {
		return r, errCapacity
	}
	var again bytes.Buffer
	if loaded.save(&again) != nil || !bytes.Equal(again.Bytes(), artifact.Bytes()) {
		return r, errCapacity
	}
	r.ReloadParity = true
	r.DevCalls = 2 * len(p)
	r.DiagnosticCalls = 2 * len(q)
	return r, writeJSON(out, r.Name+".report.json", r)
}
func selectArm(arms []armReport) string {
	best := -1
	for i, a := range arms {
		if !a.Dev.Eligible {
			continue
		}
		if best < 0 || a.Dev.SeverityCost < arms[best].Dev.SeverityCost || a.Dev.SeverityCost == arms[best].Dev.SeverityCost && a.Dev.EightNLL < arms[best].Dev.EightNLL {
			best = i
		}
	}
	if best < 0 {
		return ""
	}
	return arms[best].Name
}
func fitAll(out *os.Root, p prepared) (r result, err error) {
	r = result{Schema: "riido-statehint-capacity-development-v1", Status: "started", DevelopmentOnly: true, InternalDevPreviouslyExposed: true, ComparisonScope: "architecture bundle: identical Contextual2048 features and ordered control samples; zero linear vs seeded Glorot16ReLU MLP; not isolated initialization, Laya or ternary", Recipe: fixedRecipe(), Train: p.Train.pin, Split: p.Split.pin, Arms: []armReport{}, SelectionSource: "already exposed internal-dev177 development only; eligible severity cost, eight NLL, fixed arm order; no fresh qualification", GoVersion: runtime.Version(), RSSNote: "Go heap is not process peak RSS. Measure OS peak RSS externally for the whole fixed run; arm wall times are sequential. No native/GPU runtime."}
	if p.Validation != nil {
		v := p.Validation.pin
		r.Validation = &v
	}
	defer func() {
		if err != nil {
			r.Status = "failed; no selected arm or promotion; partial artifacts retained"
			r.SelectedArm = ""
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
	}{{"train840.source.jsonl", p.Train.bytes}, {"frozen-split.source.json", p.Split.bytes}} {
		if err = writeBytes(out, item.name, item.data); err != nil {
			return r, err
		}
	}
	if p.Validation != nil {
		if err = writeBytes(out, "exposed-validation.source.jsonl", p.Validation.bytes); err != nil {
			return r, err
		}
	}
	if err = writeJSON(out, "control.whole-family-draws.json", p.Draws); err != nil {
		return r, err
	}
	drawBytes, e := json.MarshalIndent(p.Draws, "", "  ")
	if e != nil {
		return r, errCapacity
	}
	r.DrawSHA = digest(append(drawBytes, '\n'))
	type sampleReceipt struct {
		Index   int              `json:"index"`
		Label   statehint.Intent `json:"label"`
		TextSHA string           `json:"text_sha256"`
	}
	receipt := make([]sampleReceipt, len(p.Samples))
	for i, s := range p.Samples {
		receipt[i] = sampleReceipt{i, s.Label, digest([]byte(s.Text))}
	}
	if err = writeJSON(out, "training.ordered-samples.json", receipt); err != nil {
		return r, err
	}
	receiptBytes, e := json.MarshalIndent(receipt, "", "  ")
	if e != nil {
		return r, errCapacity
	}
	r.TrainingSamplesSHA = digest(append(receiptBytes, '\n'))
	for _, kind := range []armKind{linearArm, mlpArm} {
		arm, e := fitArm(out, kind, p.Samples, p.Dev, p.Validation)
		r.Arms = append(r.Arms, arm)
		if e != nil {
			return r, e
		}
		if arm.Fit.Samples != 1454 || arm.Fit.TrainingSteps != 1840 || arm.Fit.Batches != 1840 {
			return r, errCapacity
		}
	}
	r.SelectedArm = selectArm(r.Arms)
	r.Status = "completed; exposed internal-dev selection only; no promotion"
	if r.SelectedArm == "" {
		r.Status = "completed; neither arm eligible; no selection or promotion"
	}
	err = writeJSON(out, "report.json", r)
	return r, err
}
func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-capacity", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	trainName := f.String("train", "", "pinned original train840 JSONL")
	trainHash := f.String("train-sha256", "", "frozen train840 SHA-256")
	splitName := f.String("split", "", "frozen663fit177dev metadata")
	splitHash := f.String("split-sha256", "", "frozen split SHA-256")
	valName := f.String("validation", "", "optional exposed validation120 diagnostic")
	valHash := f.String("validation-sha256", "", "optional diagnostic SHA-256")
	output := f.String("out", "", "fresh private run directory")
	check := f.Bool("check", false, "check only, no model calls or outputs")
	fit := f.Bool("fit", false, "explicit fixed fresh two-arm fit")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-capacity --train FILE --train-sha256 SHA --split FILE --split-sha256 SHA --check\nOptional --validation FILE --validation-sha256 SHA is exposed diagnostic only. Use --fit --out .cache/statehint-capacity/NEW-RUN for the fixed two-arm study. No augmentation, parent model, calibration, final test or recipe override.")
			return e
		}
		return errCapacity
	}
	trainPin, splitPin := pin{*trainName, *trainHash}, pin{*splitName, *splitHash}
	pins := []pin{trainPin, splitPin}
	if *valName != "" || *valHash != "" {
		pins = append(pins, pin{*valName, *valHash})
	}
	if f.NArg() != 0 || *check == *fit || trainPin.SHA != trainSHA || splitPin.SHA != splitSHA || *check && *output != "" || *fit && (!local(*output) || filepath.Dir(*output) != anchor) {
		return errCapacity
	}
	for i, p := range pins {
		if !validPin(p) {
			return errCapacity
		}
		for _, old := range pins[:i] {
			if old.Path == p.Path {
				return errCapacity
			}
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errCapacity
	}
	defer root.Close()
	train, e := loadInput(root, trainPin, "train", 840, balanced(105))
	if e != nil {
		return e
	}
	splitBytes, e := readPin(root, splitPin, metadataBudget)
	if e != nil {
		return e
	}
	var validation *input
	if len(pins) == 3 {
		v, e := loadInput(root, pins[2], "validation", 120, balanced(15))
		if e != nil {
			return e
		}
		validation = &v
	}
	p, e := prepare(train, metadataInput{splitPin, splitBytes}, validation)
	if e != nil {
		return e
	}
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status       string `json:"status"`
			FitFamilies  int    `json:"base_fit_families"`
			DevFamilies  int    `json:"already_exposed_internal_dev_families"`
			Draws        int    `json:"extra_whole_family_draws"`
			TrainingRows int    `json:"training_rows_per_arm"`
		}{"checked; no fitting, model calls or outputs", 663, 177, 64, len(p.Samples)})
	}
	private, e := privateOutput(root, *output)
	if e != nil {
		return e
	}
	defer private.Close()
	r, e := fitAll(private, p)
	if e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(struct {
		Status    string `json:"status"`
		Selected  string `json:"internal_dev_selected_arm"`
		Promotion bool   `json:"promotion_performed"`
	}{r.Status, r.SelectedArm, false})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "capacity study failed; check pinned training, frozen split and fresh private output")
		os.Exit(1)
	}
}
