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
)

const (
	requiredBaselineSHA = "43d41cb923b33bbbd96879488376ad5c3f600bdd53860d6d6fba852751c0c0b4"
	rubricSHA           = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	trainSHA            = "586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e"
	splitSHA            = "f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9"
	anchor              = ".cache/statehint-label-smoothing"
	metadataBudget      = 1 << 20
)

var errSmooth = errors.New("label-smoothing comparison input, checksum or output invalid")

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
	Samples    []statehintmlp.Sample
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
	Alphas        [2]float64 `json:"fixed_label_smoothing_alphas"`
}

func classCounts() [8]int { return [8]int{0, 8, 8, 8, 32, 0, 8, 0} }
func fixedRecipe() recipe {
	return recipe{40, 32, .02, .001, 1729, 1, [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, [2]string{"hard_ce_alpha000", "uniform_smoothing_alpha005"}, 1454, 1840, classCounts(), [2]float64{0, .05}}
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
		return nil, errSmooth
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > limit {
		return nil, errSmooth
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errSmooth
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errSmooth
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errSmooth
	}
	return b, nil
}
func decodeMetadata(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if d.Decode(v) != nil {
		return errSmooth
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errSmooth
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
		return input{}, errSmooth
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
			return statehintcorpus.Corpus{}, errSmooth
		}
	}
	c, _, e := statehintcorpus.Read(&b, statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		return c, errSmooth
	}
	return c, nil
}
func splitCorpus(c statehintcorpus.Corpus, s splitMetadata) (statehintcorpus.Corpus, statehintcorpus.Corpus, error) {
	if s.Schema != "statehint-completion-contrast-train-only-group-split-v1" || s.Families != 840 || s.FitFamilies != 663 || s.DevFamilies != 177 || len(s.Assignments) != 840 {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errSmooth
	}
	a := append([]assignment(nil), s.Assignments...)
	sort.Slice(a, func(i, j int) bool { return a[i].Family < a[j].Family })
	pairs, rows := c.Pairs(), c.Rows()
	if len(pairs) != len(a) {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errSmooth
	}
	var fit, dev []statehintcorpus.Row
	for i, p := range pairs {
		x := a[i]
		if x.Family != p.Family.ID || x.Group != p.Family.Lineage || x.Use != splitUse(x.Group) {
			return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errSmooth
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
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errSmooth
	}
	f, e := corpusFromRows(fit)
	if e != nil {
		return f, statehintcorpus.Corpus{}, e
	}
	d, e := corpusFromRows(dev)
	return f, d, e
}
func samples(c statehintcorpus.Corpus) []statehintmlp.Sample {
	rows := c.Rows()
	v := make([]statehintmlp.Sample, 0, len(rows))
	for _, p := range c.Pairs() {
		for _, i := range p.Rows {
			v = append(v, statehintmlp.Sample{Text: rows[i].Text, Label: rows[i].Expected})
		}
	}
	return v
}
func controlDraws(c statehintcorpus.Corpus) ([]statehintmlp.Sample, []draw, error) {
	rows, pairs := c.Rows(), c.Pairs()
	var extra []statehintmlp.Sample
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
			return nil, nil, errSmooth
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
				extra = append(extra, statehintmlp.Sample{Text: r.Text, Label: r.Expected})
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
				return errSmooth
			}
		}
	}
	if _, e := statehintcorpus.ValidateMetadata(metadata); e != nil {
		return errSmooth
	}
	return nil
}
func prepare(train input, split metadataInput, validation *input) (prepared, error) {
	p := prepared{Train: train, Split: split, Validation: validation}
	var s splitMetadata
	if decodeMetadata(split.bytes, &s) != nil {
		return prepared{}, errSmooth
	}
	var e error
	p.Fit, p.Dev, e = splitCorpus(train.corpus, s)
	if e != nil {
		return prepared{}, e
	}
	if validation != nil && disjoint(train, *validation) != nil {
		return prepared{}, errSmooth
	}
	extra, draws, e := controlDraws(p.Fit)
	if e != nil {
		return prepared{}, e
	}
	p.Draws = draws
	p.Samples = append(samples(p.Fit), extra...)
	if len(p.Samples) != 1454 || len(draws) != 64 {
		return prepared{}, errSmooth
	}
	return p, nil
}
func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errSmooth
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errSmooth
		}
		s, e := root.Lstat(dir)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || dir == anchor && s.Mode().Perm() != 0700 {
			return nil, errSmooth
		}
	}
	if e := root.Mkdir(name, 0700); e != nil {
		return nil, errSmooth
	}
	return root.OpenRoot(name)
}
func writeBytes(root *os.Root, name string, b []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errSmooth
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
		return errSmooth
	}
	return nil
}
func writeJSON(root *os.Root, name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return errSmooth
	}
	return writeBytes(root, name, append(b, '\n'))
}

type armKind int

const (
	hardArm armKind = iota
	smoothArm
)

// The same MLP and caller-owned workspace serve both fixed objectives.
type armModel struct {
	model     *statehintmlp.Model
	workspace statehintmlp.Workspace
}

func (m *armModel) predict(text string) (statehint.Prediction, error) {
	return m.model.Predict(text, &m.workspace)
}
func (m *armModel) save(w io.Writer) error { return m.model.Save(w) }
func (m *armModel) temperature() float64   { return m.model.Temperature() }
func (m *armModel) steps() uint64          { return m.model.TrainingSteps() }
func fitModel(kind armKind, training []statehintmlp.Sample) (*armModel, statehintmlp.FitReport, error) {
	if kind != hardArm && kind != smoothArm {
		return nil, statehintmlp.FitReport{}, errSmooth
	}
	r := fixedRecipe()
	m := &armModel{model: statehintmlp.NewModel()}
	f, e := m.model.Fit(training, statehintmlp.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed, LabelSmoothing: r.Alphas[kind]})
	return m, f, e
}
func loadModel(b []byte) (*armModel, error) {
	m, e := statehintmlp.Load(bytes.NewReader(b))
	return &armModel{model: m}, e
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
	LabelSmoothing   float64                 `json:"label_smoothing"`
	Seed             int64                   `json:"initialization_seed"`
	ArtifactFormat   string                  `json:"artifact_format"`
	Loader           string                  `json:"loader"`
	Parameters       int                     `json:"parameters"`
	Fit              statehintmlp.FitReport  `json:"fit"`
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
	RequiredBaselineSHA          string      `json:"required_alpha0_artifact_sha256"`
	BaselineReproduced           bool        `json:"alpha0_baseline_exactly_reproduced"`
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
			return nil, errSmooth
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
		return r, nil, duration, errSmooth
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
func fitArm(out *os.Root, kind armKind, training []statehintmlp.Sample, dev statehintcorpus.Corpus, validation *input) (armReport, error) {
	if kind != hardArm && kind != smoothArm {
		return armReport{}, errSmooth
	}
	r := armReport{Name: fixedRecipe().ArmOrder[kind], LabelSmoothing: fixedRecipe().Alphas[kind], Seed: 1729, ProbabilityOrder: statehint.Intents(), Initialization: statehintmlp.Initialization, ArtifactFormat: "RSM v3", Loader: "pkg/statehintmlp.Load", Parameters: statehintmlp.ParameterCount}
	started := time.Now()
	m, f, e := fitModel(kind, training)
	r.Fit, r.FitNS = f, time.Since(started).Nanoseconds()
	if e != nil || m.temperature() != 1 || f.LabelSmoothing != r.LabelSmoothing {
		return r, errSmooth
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
	if m.save(&artifact) != nil || artifact.Len() != statehintmlp.ArtifactBytes {
		return r, errSmooth
	}
	r.Model = pin{r.Name + ".rsm", digest(artifact.Bytes())}
	r.ArtifactBytes = artifact.Len()
	if writeBytes(out, r.Model.Path, artifact.Bytes()) != nil {
		return r, errSmooth
	}
	saved, e := readPin(out, r.Model, statehintmlp.ArtifactBytes)
	if e != nil {
		return r, e
	}
	loaded, e := loadModel(saved)
	if e != nil || loaded.temperature() != 1 || loaded.steps() != m.steps() || loaded.model.InitializationSeed() != m.model.InitializationSeed() || !parity(loaded, dev, p) || validation != nil && !parity(loaded, validation.corpus, q) {
		return r, errSmooth
	}
	var again bytes.Buffer
	if loaded.save(&again) != nil || !bytes.Equal(again.Bytes(), artifact.Bytes()) {
		return r, errSmooth
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
	r = result{Schema: "riido-statehint-label-smoothing-development-v1", Status: "started", DevelopmentOnly: true, InternalDevPreviouslyExposed: true, RequiredBaselineSHA: requiredBaselineSHA, ComparisonScope: "fixed label-smoothing objective comparison on the same fresh Contextual2048/16ReLU MLP and ordered control samples; alpha0 vs alpha0.05; exposed development, not Laya, ternary or fresh qualification", Recipe: fixedRecipe(), Train: p.Train.pin, Split: p.Split.pin, Arms: []armReport{}, SelectionSource: "already exposed internal-dev177 development only; eligible severity cost, eight NLL, fixed arm order; no fresh qualification", GoVersion: runtime.Version(), RSSNote: "Go heap is not process peak RSS. Measure OS peak RSS externally for the whole fixed run; arm wall times are sequential. No native/GPU runtime."}
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
		return r, errSmooth
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
		return r, errSmooth
	}
	r.TrainingSamplesSHA = digest(append(receiptBytes, '\n'))
	for _, kind := range []armKind{hardArm, smoothArm} {
		arm, e := fitArm(out, kind, p.Samples, p.Dev, p.Validation)
		r.Arms = append(r.Arms, arm)
		if e != nil {
			return r, e
		}
		if kind == hardArm {
			if arm.Model.SHA != requiredBaselineSHA {
				return r, errSmooth
			}
			r.BaselineReproduced = true
		}
		if arm.Fit.Samples != 1454 || arm.Fit.TrainingSteps != 1840 || arm.Fit.Batches != 1840 {
			return r, errSmooth
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
	f := flag.NewFlagSet("riido-statehint-smooth", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	trainName := f.String("train", "", "pinned original train840 JSONL")
	trainHash := f.String("train-sha256", "", "frozen train840 SHA-256")
	splitName := f.String("split", "", "frozen663fit177dev metadata")
	splitHash := f.String("split-sha256", "", "frozen split SHA-256")
	valName := f.String("validation", "", "optional exposed validation120 diagnostic")
	valHash := f.String("validation-sha256", "", "optional diagnostic SHA-256")
	output := f.String("out", "", "fresh private run directory")
	check := f.Bool("check", false, "check only, no model calls or outputs")
	fit := f.Bool("fit", false, "explicit fixed fresh alpha0/alpha0.05 comparison")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-smooth --train FILE --train-sha256 SHA --split FILE --split-sha256 SHA --check\nOptional --validation FILE --validation-sha256 SHA is exposed diagnostic only. Use --fit --out .cache/statehint-label-smoothing/NEW-RUN for the fixed two-arm smoothing study. Alpha0 must reproduce its pinned artifact before alpha0.05 is fitted. No alpha override, augmentation, parent model, calibration, final test or recipe override.")
			return e
		}
		return errSmooth
	}
	trainPin, splitPin := pin{*trainName, *trainHash}, pin{*splitName, *splitHash}
	pins := []pin{trainPin, splitPin}
	if *valName != "" || *valHash != "" {
		pins = append(pins, pin{*valName, *valHash})
	}
	if f.NArg() != 0 || *check == *fit || trainPin.SHA != trainSHA || splitPin.SHA != splitSHA || *check && *output != "" || *fit && (!local(*output) || filepath.Dir(*output) != anchor) {
		return errSmooth
	}
	for i, p := range pins {
		if !validPin(p) {
			return errSmooth
		}
		for _, old := range pins[:i] {
			if old.Path == p.Path {
				return errSmooth
			}
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errSmooth
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
		fmt.Fprintln(os.Stderr, "label-smoothing study failed; check pinned training, frozen split, baseline reproduction and fresh private output")
		os.Exit(1)
	}
}
