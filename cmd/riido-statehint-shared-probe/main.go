// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// One supervised shared-head fit on frozen features; never runs the backbone.
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
	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintsharedprobe"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	rubricSHA            = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	trainSHA             = "586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e"
	splitSHA             = "f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9"
	validationSHA        = "1cfe4911ff697f9c6c77f7be627a75eb14c87061b87207edf4b009c2f21e00cc"
	validationFeatureSHA = "da8c37ebcb4466d03779f29a949ee402b394821c8d849c62837134249a40792d"
	anchor               = ".cache/statehint-shared-probe"
)

var errProbe = errors.New("shared feature probe input, checksum or output invalid")

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
		return nil, errProbe
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > limit {
		return nil, errProbe
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errProbe
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errProbe
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errProbe
	}
	return b, nil
}

func decodeMetadata(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if d.Decode(v) != nil {
		return errProbe
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errProbe
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
		return input{}, errProbe
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
			return statehintcorpus.Corpus{}, errProbe
		}
	}
	c, _, e := statehintcorpus.Read(&b, statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		return c, errProbe
	}
	return c, nil
}

func splitCorpus(c statehintcorpus.Corpus, s splitMetadata) (statehintcorpus.Corpus, statehintcorpus.Corpus, error) {
	if s.Schema != "statehint-completion-contrast-train-only-group-split-v1" || s.Families != 840 || s.FitFamilies != 663 || s.DevFamilies != 177 || len(s.Assignments) != 840 {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errProbe
	}
	a := append([]assignment(nil), s.Assignments...)
	sort.Slice(a, func(i, j int) bool { return a[i].Family < a[j].Family })
	pairs, rows := c.Pairs(), c.Rows()
	if len(pairs) != len(a) {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errProbe
	}
	var fit, dev []statehintcorpus.Row
	for i, p := range pairs {
		x := a[i]
		if x.Family != p.Family.ID || x.Group != p.Family.Lineage || x.Use != splitUse(x.Group) {
			return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errProbe
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
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, errProbe
	}
	f, e := corpusFromRows(fit)
	if e != nil {
		return f, statehintcorpus.Corpus{}, e
	}
	d, e := corpusFromRows(dev)
	return f, d, e
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errProbe
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errProbe
		}
		s, e := root.Lstat(dir)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || dir == anchor && s.Mode().Perm() != 0700 {
			return nil, errProbe
		}
	}
	if e := root.Mkdir(name, 0700); e != nil {
		return nil, errProbe
	}
	return root.OpenRoot(name)
}

func writeBytes(root *os.Root, name string, b []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errProbe
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
		return errProbe
	}
	return nil
}

func writeJSON(root *os.Root, name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return errProbe
	}
	return writeBytes(root, name, append(b, '\n'))
}

type featureIdentity struct {
	ID       string           `json:"id"`
	Locale   string           `json:"locale"`
	TextSHA  string           `json:"text_sha256"`
	Expected statehint.Intent `json:"expected_intent"`
}
type sidecar struct {
	Schema         string             `json:"schema"`
	BaseSHA        string             `json:"base_sha256"`
	InstructionSHA string             `json:"instruction_sha256"`
	IntentOrder    []statehint.Intent `json:"intent_order"`
	CorpusSHA      string             `json:"corpus_sha256"`
	FeatureSHA     string             `json:"feature_sha256"`
	RowsCount      int                `json:"rows_count"`
	Shape          [3]int             `json:"shape"`
	Dtype          string             `json:"dtype"`
	Rows           []featureIdentity  `json:"rows"`
}
type validationScores struct {
	Schema         string             `json:"schema"`
	Status         string             `json:"status"`
	ValidationSHA  string             `json:"validation_sha256"`
	BaseSHA        string             `json:"base_sha256"`
	InstructionSHA string             `json:"instruction_sha256"`
	FeatureSHA     string             `json:"feature_cache_sha256"`
	IntentOrder    []statehint.Intent `json:"intent_order"`
	Rows           []featureIdentity  `json:"rows"`
}
type featureFile struct {
	pin   pin
	file  *os.File
	store *statehintsharedprobe.Store
}

func openFeatures(root *os.Root, p pin, rows int) (featureFile, error) {
	var out featureFile
	if !validPin(p) || rows < 1 || rows > statehintsharedprobe.MaxRows {
		return out, errProbe
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() != int64(rows*statehintsharedprobe.RowBytes) {
		return out, errProbe
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return out, errProbe
	}
	success := false
	defer func() {
		if !success {
			f.Close()
		}
	}()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return out, errProbe
	}
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, after.Size()+1))
	if e != nil || n != after.Size() || hex.EncodeToString(h.Sum(nil)) != p.SHA {
		return out, errProbe
	}
	store, e := statehintsharedprobe.NewStore(f, rows)
	if e != nil {
		return out, errProbe
	}
	success = true
	return featureFile{p, f, store}, nil
}
func checkOrder(order []statehint.Intent) bool {
	if len(order) != 8 {
		return false
	}
	for i, v := range statehint.Intents() {
		if order[i] != v {
			return false
		}
	}
	return true
}
func mapIdentities(c statehintcorpus.Corpus, identities []featureIdentity, requireLabels bool) ([]statehintsharedprobe.Sample, error) {
	rows := c.Rows()
	if len(rows) != len(identities) {
		return nil, errProbe
	}
	order := make([]int, len(identities))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return identities[order[i]].ID < identities[order[j]].ID })
	for i := 1; i < len(order); i++ {
		if identities[order[i-1]].ID == identities[order[i]].ID {
			return nil, errProbe
		}
	}
	samples := make([]statehintsharedprobe.Sample, len(rows))
	for i, row := range rows {
		j := sort.Search(len(order), func(j int) bool { return identities[order[j]].ID >= row.ID })
		if j == len(order) {
			return nil, errProbe
		}
		index := order[j]
		id := identities[index]
		if id.ID != row.ID || id.Locale != row.Locale || id.TextSHA != digest([]byte(row.Text)) || requireLabels && id.Expected != row.Expected {
			return nil, errProbe
		}
		samples[i] = statehintsharedprobe.Sample{RowIndex: index, Label: row.Expected}
	}
	return samples, nil
}
func mapSidecar(c statehintcorpus.Corpus, s sidecar, corpusSHA, featureSHA string) ([]statehintsharedprobe.Sample, error) {
	if s.Schema != "riido-statehint-laya-feature-sidecar-v1" || s.BaseSHA != statehintsharedprobe.BaseSHA || s.InstructionSHA != statehintsharedprobe.InstructionSHA || !checkOrder(s.IntentOrder) || s.CorpusSHA != corpusSHA || s.FeatureSHA != featureSHA || s.RowsCount != len(c.Rows()) || s.Shape != [3]int{s.RowsCount, 8, 1024} || s.Dtype != "float32_little_endian" {
		return nil, errProbe
	}
	return mapIdentities(c, s.Rows, true)
}
func mapValidation(c statehintcorpus.Corpus, s validationScores, corpusSHA, featureSHA string) ([]statehintsharedprobe.Sample, error) {
	if s.Schema != "riido-statehint-laya-reference-scores-v1" || s.Status != "reference_only" || s.ValidationSHA != corpusSHA || s.BaseSHA != statehintsharedprobe.BaseSHA || s.InstructionSHA != statehintsharedprobe.InstructionSHA || s.FeatureSHA != featureSHA || !checkOrder(s.IntentOrder) {
		return nil, errProbe
	}
	return mapIdentities(c, s.Rows, false)
}
func subsetMapping(full, subset statehintcorpus.Corpus, mapping []statehintsharedprobe.Sample) ([]statehintsharedprobe.Sample, error) {
	rows := full.Rows()
	if len(rows) != len(mapping) {
		return nil, errProbe
	}
	indices := make([]int, len(rows))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool { return rows[indices[i]].ID < rows[indices[j]].ID })
	sr := subset.Rows()
	out := make([]statehintsharedprobe.Sample, 0, len(sr))
	for _, pair := range subset.Pairs() {
		for _, index := range pair.Rows {
			row := sr[index]
			j := sort.Search(len(indices), func(j int) bool { return rows[indices[j]].ID >= row.ID })
			if j == len(indices) || rows[indices[j]].ID != row.ID {
				return nil, errProbe
			}
			out = append(out, mapping[indices[j]])
		}
	}
	return out, nil
}

type scoredRow struct {
	ID         string               `json:"id"`
	Locale     string               `json:"locale"`
	Expected   statehint.Intent     `json:"expected_intent"`
	Prediction statehint.Prediction `json:"prediction"`
}

func score(m *statehintsharedprobe.Model, store *statehintsharedprobe.Store, c statehintcorpus.Corpus, mapping []statehintsharedprobe.Sample) ([]statehint.Prediction, error) {
	rows := c.Rows()
	if len(rows) != len(mapping) {
		return nil, errProbe
	}
	out := make([]statehint.Prediction, len(rows))
	var w statehintsharedprobe.Workspace
	for i, s := range mapping {
		if s.Label != rows[i].Expected || store.ReadRow(s.RowIndex, &w) != nil {
			return nil, errProbe
		}
		p, e := m.ScoreFeatures(&w.Row, &w)
		if e != nil {
			return nil, errProbe
		}
		out[i] = p
	}
	return out, nil
}

// Training order is paired-family order; evaluation must retain corpus row order.
func evaluationMapping(full, subset statehintcorpus.Corpus, mapping []statehintsharedprobe.Sample) ([]statehintsharedprobe.Sample, error) {
	rows := full.Rows()
	if len(rows) != len(mapping) {
		return nil, errProbe
	}
	indices := make([]int, len(rows))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool { return rows[indices[i]].ID < rows[indices[j]].ID })
	out := make([]statehintsharedprobe.Sample, len(subset.Rows()))
	for i, row := range subset.Rows() {
		j := sort.Search(len(indices), func(j int) bool { return rows[indices[j]].ID >= row.ID })
		if j == len(indices) || rows[indices[j]].ID != row.ID {
			return nil, errProbe
		}
		out[i] = mapping[indices[j]]
	}
	return out, nil
}
func evaluate(out *os.Root, name string, m *statehintsharedprobe.Model, store *statehintsharedprobe.Store, c statehintcorpus.Corpus, mapping []statehintsharedprobe.Sample) (statehintfamily.Report, []statehint.Prediction, error) {
	p, e := score(m, store, c, mapping)
	if e != nil {
		return statehintfamily.Report{}, nil, e
	}
	r, e := statehintfamily.Evaluate(c, p)
	if e != nil {
		return r, nil, errProbe
	}
	rows := c.Rows()
	saved := make([]scoredRow, len(rows))
	for i, row := range rows {
		saved[i] = scoredRow{row.ID, row.Locale, row.Expected, p[i]}
	}
	return r, p, writeJSON(out, name+".predictions.json", saved)
}

type referenceArm struct {
	Origin              string  `json:"origin"`
	Temperature         float64 `json:"temperature"`
	RawCorrect          int     `json:"raw_eight_correct"`
	EightNLL            float64 `json:"eight_nll"`
	CompletionCorrect   int     `json:"completion_correct"`
	CompletionProposals int     `json:"completion_proposals"`
	TrainingStepsKnown  bool    `json:"training_steps_known"`
	TrainingSteps       *uint64 `json:"training_steps"`
}
type referenceReport struct {
	Schema     string          `json:"schema"`
	Status     string          `json:"status"`
	FeatureSHA string          `json:"feature_cache_sha256"`
	Validation pin             `json:"validation_input"`
	Arms       [2]referenceArm `json:"arms"`
}
type result struct {
	Schema              string                         `json:"schema"`
	Status              string                         `json:"status"`
	ResearchOnly        bool                           `json:"research_only"`
	RequiresBackbone    bool                           `json:"requires_frozen_backbone_for_novel_text"`
	StandaloneTextModel bool                           `json:"standalone_text_model"`
	HumanTruth          bool                           `json:"human_truth"`
	BaseSHA             string                         `json:"base_sha256"`
	InstructionSHA      string                         `json:"instruction_sha256"`
	BiasFixedZero       bool                           `json:"shared_bias_fixed_zero"`
	Parameters          int                            `json:"shared_parameters"`
	Initialization      string                         `json:"initialization"`
	Fit                 statehintsharedprobe.FitReport `json:"fit"`
	FitNS               int64                          `json:"fit_nanoseconds"`
	Train               pin                            `json:"train_input"`
	Split               pin                            `json:"split_input"`
	Features            pin                            `json:"training_feature_file"`
	Sidecar             pin                            `json:"training_sidecar"`
	Validation          pin                            `json:"validation_input"`
	ValidationFeatures  pin                            `json:"validation_feature_file"`
	ValidationMap       pin                            `json:"validation_mapping_scores"`
	Model               pin                            `json:"model"`
	ArtifactBytes       int                            `json:"artifact_bytes"`
	ReloadParity        bool                           `json:"trained_save_load_exact_parity"`
	Dev                 statehintfamily.Report         `json:"already_exposed_internal_dev_descriptive"`
	Diagnostic          statehintfamily.Report         `json:"already_exposed_validation_descriptive"`
	ReferenceInput      pin                            `json:"historical_reference_aggregate"`
	Reference           [2]referenceArm                `json:"historical_validation_reference"`
	SelectionWeight     int                            `json:"diagnostic_selection_weight"`
	Selected            string                         `json:"selected_arm"`
	Promotion           bool                           `json:"promotion_performed"`
	DeploymentQualified bool                           `json:"deployment_qualified"`
	CalibrationCalls    int                            `json:"calibration_forwards"`
	TestCalls           int                            `json:"test_forwards"`
	BackboneCalls       int                            `json:"go_command_backbone_calls"`
	DevCalls            int                            `json:"dev_head_forwards_including_reload"`
	ValidationCalls     int                            `json:"validation_head_forwards_including_reload"`
	GoHeap              uint64                         `json:"go_heap_after_fit_not_os_rss"`
	GoVersion           string                         `json:"go_version"`
}
type prepared struct {
	train, validation                        input
	split, sidecar, validationMap, reference metadataInput
	features, validationFeatures             featureFile
	fit, dev                                 statehintcorpus.Corpus
	training, devMapping, validationMapping  []statehintsharedprobe.Sample
	referenceArms                            [2]referenceArm
}

func fitOne(out *os.Root, p prepared) (r result, err error) {
	r = result{Schema: "riido-statehint-shared-probe-development-v1", Status: "started", ResearchOnly: true, RequiresBackbone: true, BaseSHA: statehintsharedprobe.BaseSHA, InstructionSHA: statehintsharedprobe.InstructionSHA, BiasFixedZero: true, Parameters: 1024, Initialization: "fresh zero shared head; no teacher pseudo labels", Train: p.train.pin, Split: p.split.pin, Features: p.features.pin, Sidecar: p.sidecar.pin, Validation: p.validation.pin, ValidationFeatures: p.validationFeatures.pin, ValidationMap: p.validationMap.pin, ReferenceInput: p.reference.pin, Reference: p.referenceArms, GoVersion: runtime.Version()}
	defer func() {
		if err != nil {
			r.Status = "failed; partial private artifacts retained; no selection/promotion"
			if e := writeJSON(out, "FAILED.json", r); e != nil {
				err = e
			}
		}
	}()
	if err = writeJSON(out, "recipe.json", struct {
		Epochs, Batch   int
		Rate, Decay     float64
		Seed            int64
		Temperature     float64
		SelectionWeight int
		BiasFixedZero   bool
	}{40, 32, .001, .01, 1729, 1, 0, true}); err != nil {
		return r, err
	}
	if err = writeJSON(out, "fit.source-row-indices-labels.json", p.training); err != nil {
		return r, err
	}
	m := statehintsharedprobe.NewModel()
	started := time.Now()
	r.Fit, err = m.Fit(p.features.store, p.training, statehintsharedprobe.FitOptions{Epochs: 40, BatchSize: 32, LearningRate: .001, WeightDecay: .01, Seed: 1729})
	r.FitNS = time.Since(started).Nanoseconds()
	if err != nil {
		return r, err
	}
	if r.Fit.Samples != 1326 || r.Fit.Batches != 1680 || r.Fit.TrainingSteps != 1680 || m.Temperature() != 1 {
		return r, errProbe
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r.GoHeap = memory.HeapAlloc
	var devPred, valPred []statehint.Prediction
	r.Dev, devPred, err = evaluate(out, "internal-dev", m, p.features.store, p.dev, p.devMapping)
	if err != nil {
		return r, err
	}
	r.Diagnostic, valPred, err = evaluate(out, "exposed-validation", m, p.validationFeatures.store, p.validation.corpus, p.validationMapping)
	if err != nil {
		return r, err
	}
	var binaryModel bytes.Buffer
	if m.Save(&binaryModel) != nil || binaryModel.Len() != statehintsharedprobe.ArtifactBytes {
		return r, errProbe
	}
	r.Model = pin{"shared-head.rsp", digest(binaryModel.Bytes())}
	r.ArtifactBytes = binaryModel.Len()
	if err = writeBytes(out, r.Model.Path, binaryModel.Bytes()); err != nil {
		return r, err
	}
	saved, e := readPin(out, r.Model, statehintsharedprobe.ArtifactBytes)
	if e != nil {
		return r, e
	}
	loaded, e := statehintsharedprobe.Load(bytes.NewReader(saved))
	if e != nil {
		return r, errProbe
	}
	for _, check := range []struct {
		store   *statehintsharedprobe.Store
		c       statehintcorpus.Corpus
		mapping []statehintsharedprobe.Sample
		want    []statehint.Prediction
	}{{p.features.store, p.dev, p.devMapping, devPred}, {p.validationFeatures.store, p.validation.corpus, p.validationMapping, valPred}} {
		got, e := score(loaded, check.store, check.c, check.mapping)
		if e != nil || len(got) != len(check.want) {
			return r, errProbe
		}
		for i := range got {
			if got[i] != check.want[i] {
				return r, errProbe
			}
		}
	}
	var again bytes.Buffer
	if loaded.Save(&again) != nil || !bytes.Equal(saved, again.Bytes()) {
		return r, errProbe
	}
	r.ReloadParity = true
	r.DevCalls = 2 * len(devPred)
	r.ValidationCalls = 2 * len(valPred)
	r.Status = "completed; one supervised shared-head probe; descriptive only, no selection/promotion"
	err = writeJSON(out, "report.json", r)
	return r, err
}
func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-shared-probe", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	names := []string{"train", "split", "features", "sidecar", "validation", "validation-features", "validation-scores", "reference-report"}
	paths := make([]*string, len(names))
	hashes := make([]*string, len(names))
	for i, name := range names {
		paths[i] = f.String(name, "", "explicit pinned local input")
		hashes[i] = f.String(name+"-sha256", "", "input SHA-256")
	}
	check := f.Bool("check", false, "validate only, no head fitting or outputs")
	fit := f.Bool("fit", false, "one fixed pure-Go shared-head fit")
	dest := f.String("out", "", "fresh private output")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-shared-probe requires explicit FILE/SHA pairs for train, split, features, sidecar, validation, validation-features, validation-scores and reference-report. Use --check, or --fit --out .cache/statehint-shared-probe/NEW-RUN. No native runtime, new feature extraction, calibration, final-test or recipe override.")
			return e
		}
		return errProbe
	}
	if f.NArg() != 0 || *check == *fit || *check && *dest != "" || *fit && (!local(*dest) || filepath.Dir(*dest) != anchor) {
		return errProbe
	}
	pins := make([]pin, len(names))
	for i := range names {
		pins[i] = pin{*paths[i], *hashes[i]}
		if !validPin(pins[i]) {
			return errProbe
		}
		for _, old := range pins[:i] {
			if old.Path == pins[i].Path {
				return errProbe
			}
		}
	}
	if pins[0].SHA != trainSHA || pins[1].SHA != splitSHA || pins[4].SHA != validationSHA || pins[5].SHA != validationFeatureSHA {
		return errProbe
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errProbe
	}
	defer root.Close()
	p := prepared{}
	p.train, e = loadInput(root, pins[0], "train", 840, balanced(105))
	if e != nil {
		return e
	}
	p.validation, e = loadInput(root, pins[4], "validation", 120, balanced(15))
	if e != nil {
		return e
	}
	for _, item := range []struct {
		index  int
		target *metadataInput
	}{{1, &p.split}, {3, &p.sidecar}, {6, &p.validationMap}, {7, &p.reference}} {
		b, e := readPin(root, pins[item.index], 2<<20)
		if e != nil {
			return e
		}
		*item.target = metadataInput{pins[item.index], b}
	}
	var split splitMetadata
	var side sidecar
	var mapping validationScores
	var reference referenceReport
	if decodeMetadata(p.split.bytes, &split) != nil || decodeMetadata(p.sidecar.bytes, &side) != nil || decodeMetadata(p.validationMap.bytes, &mapping) != nil || decodeMetadata(p.reference.bytes, &reference) != nil {
		return errProbe
	}
	p.fit, p.dev, e = splitCorpus(p.train.corpus, split)
	if e != nil {
		return e
	}
	all, e := mapSidecar(p.train.corpus, side, p.train.pin.SHA, pins[2].SHA)
	if e != nil {
		return e
	}
	p.training, e = subsetMapping(p.train.corpus, p.fit, all)
	if e != nil {
		return e
	}
	p.devMapping, e = evaluationMapping(p.train.corpus, p.dev, all)
	if e != nil {
		return e
	}
	p.validationMapping, e = mapValidation(p.validation.corpus, mapping, p.validation.pin.SHA, pins[5].SHA)
	if e != nil {
		return e
	}
	if reference.Schema != "riido-statehint-laya-reference-aggregate-v1" || reference.Status != "reference_only" || reference.FeatureSHA != pins[5].SHA || reference.Validation.SHA != p.validation.pin.SHA || reference.Arms[0].Origin != "base_pretrained" || reference.Arms[1].Origin != "published_v1_delta" {
		return errProbe
	}
	for _, arm := range reference.Arms {
		if arm.TrainingStepsKnown || arm.TrainingSteps != nil {
			return errProbe
		}
	}
	p.referenceArms = reference.Arms
	p.features, e = openFeatures(root, pins[2], 1680)
	if e != nil {
		return e
	}
	defer p.features.file.Close()
	p.validationFeatures, e = openFeatures(root, pins[5], 240)
	if e != nil {
		return e
	}
	defer p.validationFeatures.file.Close()
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status                           string `json:"status"`
			FitRows, DevRows, ValidationRows int
			BackboneCalls                    int
		}{"checked; no head Fit/backbone calls or outputs", len(p.training), len(p.devMapping), len(p.validationMapping), 0})
	}
	private, e := privateOutput(root, *dest)
	if e != nil {
		return e
	}
	defer private.Close()
	r, e := fitOne(private, p)
	if e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(struct {
		Status    string `json:"status"`
		Steps     uint64 `json:"new_go_head_training_steps"`
		Promotion bool   `json:"promotion_performed"`
	}{r.Status, r.Fit.TrainingSteps, false})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "shared feature probe failed; check complete pinned features, source mapping and frozen split")
		os.Exit(1)
	}
}
