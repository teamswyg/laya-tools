// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only controlled, source-derived training contrast.
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
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	rubricSHA      = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	trainSHA       = "586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e"
	splitSHA       = "f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9"
	anchor         = ".cache/statehint-completion-contrast"
	metadataBudget = 1 << 20
)

var errContrast = errors.New("controlled contrast input, checksum or output invalid")

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
type sourceEntry struct {
	Family   string           `json:"family_id"`
	Pair     string           `json:"pair_id"`
	Expected statehint.Intent `json:"expected_intent"`
	Group    string           `json:"effective_leakage_group_id"`
	Sources  []string         `json:"source_connected_train_family_ids"`
	Groups   []string         `json:"source_connected_train_group_ids"`
	Members  []string         `json:"all_selected_source_component_members"`
	RowIDs   [2]string        `json:"row_ids"`
	TextSHA  [2]string        `json:"text_sha256"`
}
type sourceAudit struct {
	Schema          string        `json:"schema"`
	AugmentationSHA string        `json:"augmentation_sha256"`
	TrainSHA        string        `json:"train840_sha256"`
	SplitSHA        string        `json:"frozen_internal_split_sha256"`
	Entries         []sourceEntry `json:"entries"`
}
type draw struct {
	Family   string           `json:"family_id"`
	Group    string           `json:"leakage_group_id"`
	Expected statehint.Intent `json:"expected_intent"`
	RowIDs   [2]string        `json:"row_ids_ko_en"`
	Rank     string           `json:"sha256_rank"`
}
type prepared struct {
	Train, Augmentation input
	Split, Audit        metadataInput
	Validation          *input
	Fit, Dev            statehintcorpus.Corpus
	Samples             [2][]statehintwide.Sample
	Draws               []draw
	Assignments         []assignment
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

func classCounts() [8]int {
	var counts [8]int
	for _, item := range []struct {
		label statehint.Intent
		count int
	}{{statehint.CompletionReport, 32}, {statehint.Reference, 8}, {statehint.Progress, 8}, {statehint.Planned, 8}, {statehint.Blocker, 8}} {
		index, _ := statehint.IntentIndex(item.label)
		counts[index] = item.count
	}
	return counts
}
func fixedRecipe() recipe {
	return recipe{40, 32, .02, .001, 1729, 1, [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, [2]string{"class_matched_control", "matched_semantic_contrast"}, 1454, 1840, classCounts()}
}
func fitOptions() statehintwide.FitOptions {
	r := fixedRecipe()
	return statehintwide.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed}
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
		return nil, errContrast
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > limit {
		return nil, errContrast
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errContrast
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errContrast
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errContrast
	}
	return b, nil
}
func decodeMetadata(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	if d.Decode(v) != nil {
		return errContrast
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errContrast
	}
	return nil
}
func loadInput(root *os.Root, p pin, partition string, families int, counts [8]int) (input, error) {
	b, e := readPin(root, p, statehintcorpus.MaxFileBytes)
	if e != nil {
		return input{}, e
	}
	c, s, e := statehintcorpus.Read(bytes.NewReader(b), statehintcorpus.Options{Partition: partition, RubricSHA: rubricSHA})
	if e != nil || s.Families != families || s.Rows != 2*families || s.IntentFamilies != counts {
		return input{}, errContrast
	}
	return input{p, b, c, s}, nil
}
func balanced(count int) (v [8]int) {
	for i := range v {
		v[i] = count
	}
	return
}
func findAssignment(a []assignment, id string) (assignment, bool) {
	i := sort.Search(len(a), func(i int) bool { return a[i].Family >= id })
	if i < len(a) && a[i].Family == id {
		return a[i], true
	}
	return assignment{}, false
}
func splitUse(group string) string {
	s := sha256.Sum256([]byte("completion-contrast-internal-split-1729:" + group))
	if binary.BigEndian.Uint32(s[:4])%5 == 0 {
		return "dev"
	}
	return "fit"
}
func splitCorpus(c statehintcorpus.Corpus, s splitMetadata) (statehintcorpus.Corpus, statehintcorpus.Corpus, []assignment, error) {
	if s.Schema != "statehint-completion-contrast-train-only-group-split-v1" || s.Families != 840 || s.FitFamilies != 663 || s.DevFamilies != 177 || len(s.Assignments) != 840 {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errContrast
	}
	a := append([]assignment(nil), s.Assignments...)
	sort.Slice(a, func(i, j int) bool { return a[i].Family < a[j].Family })
	pairs := c.Pairs()
	rows := c.Rows()
	if len(pairs) != len(a) {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errContrast
	}
	var fit, dev []statehintcorpus.Row
	for i, p := range pairs {
		x := a[i]
		if x.Family != p.Family.ID || x.Group != p.Family.Lineage || x.Use != splitUse(x.Group) {
			return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errContrast
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
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errContrast
	}
	f, e := corpusFromRows(fit)
	if e != nil {
		return f, statehintcorpus.Corpus{}, nil, e
	}
	d, e := corpusFromRows(dev)
	return f, d, a, e
}
func corpusFromRows(rows []statehintcorpus.Row) (statehintcorpus.Corpus, error) {
	var b bytes.Buffer
	for _, r := range rows {
		// Rows() detaches empty slices as nil. Unclear permits an empty
		// assertion list, but the strict JSONL schema requires [] rather
		// than null. Preserve that valid annotation when making subsets.
		if r.AssertionForms == nil {
			r.AssertionForms = []string{}
		}
		if json.NewEncoder(&b).Encode(r) != nil {
			return statehintcorpus.Corpus{}, errContrast
		}
	}
	c, _, e := statehintcorpus.Read(&b, statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		return c, errContrast
	}
	return c, nil
}
func uniqueSorted(values []string) ([]string, bool) {
	v := append([]string(nil), values...)
	sort.Strings(v)
	if len(v) == 0 {
		return nil, false
	}
	for i := 1; i < len(v); i++ {
		if v[i-1] == v[i] {
			return nil, false
		}
	}
	return v, true
}
func contains(sorted []string, s string) bool {
	i := sort.SearchStrings(sorted, s)
	return i < len(sorted) && sorted[i] == s
}
func checkSources(aug input, a []assignment, s sourceAudit, trainPin, splitPin pin) error {
	if s.Schema != "statehint-completion-contrast-source-groups-v1" || s.AugmentationSHA != aug.pin.SHA || s.TrainSHA != trainPin.SHA || s.SplitSHA != splitPin.SHA || len(s.Entries) != 64 {
		return errContrast
	}
	entries := append([]sourceEntry(nil), s.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Family < entries[j].Family })
	pairs := aug.corpus.Pairs()
	rows := aug.corpus.Rows()
	if len(entries) != len(pairs) {
		return errContrast
	}
	type edge struct{ source, group string }
	var edges []edge
	for i, p := range pairs {
		x := entries[i]
		if x.Family != p.Family.ID || x.Expected != p.Family.Expected || x.Group != p.Family.Lineage || x.Pair == "" {
			return errContrast
		}
		for locale, index := range p.Rows {
			if x.RowIDs[locale] != rows[index].ID || x.TextSHA[locale] != digest([]byte(rows[index].Text)) {
				return errContrast
			}
		}
		sources, ok := uniqueSorted(x.Sources)
		if !ok {
			return errContrast
		}
		groups, ok := uniqueSorted(x.Groups)
		if !ok || !contains(groups, x.Group) {
			return errContrast
		}
		members, ok := uniqueSorted(x.Members)
		if !ok {
			return errContrast
		}
		for _, id := range sources {
			v, ok := findAssignment(a, id)
			if !ok || v.Use != "fit" || !contains(groups, v.Group) || !contains(members, id) {
				return errContrast
			}
		}
		for _, id := range members {
			v, ok := findAssignment(a, id)
			if !ok || v.Use != "fit" || !contains(groups, v.Group) {
				return errContrast
			}
		}
		for _, group := range groups {
			found := false
			for _, v := range a {
				if v.Group == group {
					found = true
					if v.Use != "fit" || !contains(members, v.Family) {
						return errContrast
					}
				}
			}
			if !found {
				return errContrast
			}
			edges = append(edges, edge{group, x.Group})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].source != edges[j].source {
			return edges[i].source < edges[j].source
		}
		return edges[i].group < edges[j].group
	})
	for i := 1; i < len(edges); i++ {
		if edges[i-1].source == edges[i].source && edges[i-1].group != edges[i].group {
			return errContrast
		}
	}
	// The two members of each matched pair must stay in one group. The fixed
	// augmentation is one C and one R/P/planned/B family per pair.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Pair != entries[j].Pair {
			return entries[i].Pair < entries[j].Pair
		}
		return entries[i].Family < entries[j].Family
	})
	for i := 0; i < len(entries); i += 2 {
		left, right := entries[i], entries[i+1]
		if left.Pair != right.Pair || left.Group != right.Group || i+2 < len(entries) && entries[i+2].Pair == left.Pair {
			return errContrast
		}
		if (left.Expected == statehint.CompletionReport) == (right.Expected == statehint.CompletionReport) {
			return errContrast
		}
	}
	return nil
}
func disjoint(inputs ...input) error {
	var metadata []statehintcorpus.Family
	var ids, texts []string
	for _, in := range inputs {
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
				return errContrast
			}
		}
	}
	if _, e := statehintcorpus.ValidateMetadata(metadata); e != nil {
		return errContrast
	}
	return nil
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
	rows := c.Rows()
	pairs := c.Pairs()
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
			return nil, nil, errContrast
		}
		sort.Slice(ranked, func(i, j int) bool {
			ri, rj := ranked[i].rank, ranked[j].rank
			if ri != rj {
				return ri < rj
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
func prepare(train, aug input, split, audit metadataInput, validation *input) (prepared, error) {
	p := prepared{Train: train, Augmentation: aug, Split: split, Audit: audit, Validation: validation}
	var s splitMetadata
	var source sourceAudit
	if decodeMetadata(split.bytes, &s) != nil || decodeMetadata(audit.bytes, &source) != nil {
		return prepared{}, errContrast
	}
	var e error
	p.Fit, p.Dev, p.Assignments, e = splitCorpus(train.corpus, s)
	if e != nil || checkSources(aug, p.Assignments, source, train.pin, split.pin) != nil {
		return prepared{}, errContrast
	}
	ins := []input{train, aug}
	if validation != nil {
		ins = append(ins, *validation)
	}
	if disjoint(ins...) != nil {
		return prepared{}, errContrast
	}
	base := samples(p.Fit)
	extra, draws, e := controlDraws(p.Fit)
	if e != nil {
		return prepared{}, e
	}
	p.Draws = draws
	p.Samples[0] = append(append([]statehintwide.Sample(nil), base...), extra...)
	p.Samples[1] = append(append([]statehintwide.Sample(nil), base...), samples(aug.corpus)...)
	if len(p.Samples[0]) != 1454 || len(p.Samples[1]) != 1454 {
		return prepared{}, errContrast
	}
	return p, nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errContrast
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errContrast
		}
		s, e := root.Lstat(dir)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || dir == anchor && s.Mode().Perm() != 0700 {
			return nil, errContrast
		}
	}
	if e := root.Mkdir(name, 0700); e != nil {
		return nil, errContrast
	}
	return root.OpenRoot(name)
}
func writeBytes(root *os.Root, name string, b []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errContrast
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
		return errContrast
	}
	return nil
}
func writeJSON(root *os.Root, name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return errContrast
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
	Name             string                  `json:"arm"`
	Fit              statehintwide.FitReport `json:"fit"`
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
	Schema                    string      `json:"schema"`
	Status                    string      `json:"status"`
	DevelopmentOnly           bool        `json:"development_only"`
	HumanTruth                bool        `json:"human_truth"`
	Recipe                    recipe      `json:"fixed_recipe"`
	Train                     pin         `json:"train_input"`
	Augmentation              pin         `json:"augmentation_input"`
	Split                     pin         `json:"frozen_split_input"`
	Audit                     pin         `json:"source_group_audit_input"`
	Validation                *pin        `json:"optional_exposed_validation_input,omitempty"`
	Arms                      []armReport `json:"arms"`
	SelectedArm               string      `json:"internal_dev_selected_arm"`
	SelectionSource           string      `json:"selection_source"`
	Promotion                 bool        `json:"promotion_performed"`
	DeploymentQualified       bool        `json:"deployment_qualified"`
	CalibrationCalls          int         `json:"calibration_forwards"`
	TestCalls                 int         `json:"test_forwards"`
	ValidationSelectionWeight int         `json:"exposed_validation_selection_weight"`
	GoVersion                 string      `json:"go_version"`
	RSSNote                   string      `json:"resource_measurement_note"`
}

func predict(m *statehintwide.Model, c statehintcorpus.Corpus) ([]statehint.Prediction, error) {
	rows := c.Rows()
	p := make([]statehint.Prediction, len(rows))
	var w statehintwide.Workspace
	for i, r := range rows {
		v, e := m.Predict(r.Text, &w)
		if e != nil {
			return nil, errContrast
		}
		p[i] = v
	}
	return p, nil
}
func evaluateAndSave(out *os.Root, name string, m *statehintwide.Model, c statehintcorpus.Corpus) (statehintfamily.Report, []statehint.Prediction, int64, error) {
	started := time.Now()
	p, e := predict(m, c)
	duration := time.Since(started).Nanoseconds()
	if e != nil {
		return statehintfamily.Report{}, nil, duration, e
	}
	r, e := statehintfamily.Evaluate(c, p)
	if e != nil {
		return r, nil, duration, errContrast
	}
	rows := c.Rows()
	scores := make([]scoredRow, len(rows))
	for i, row := range rows {
		scores[i] = scoredRow{row.ID, row.Family, row.Locale, row.Expected, p[i]}
	}
	e = writeJSON(out, name+".predictions.json", scores)
	return r, p, duration, e
}
func parity(m *statehintwide.Model, c statehintcorpus.Corpus, want []statehint.Prediction) bool {
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
func fitArm(out *os.Root, name string, training []statehintwide.Sample, dev statehintcorpus.Corpus, validation *input) (armReport, error) {
	r := armReport{Name: name, ProbabilityOrder: statehint.Intents()}
	m := statehintwide.NewModel(statehintwide.Contextual)
	started := time.Now()
	fit, e := m.Fit(training, fitOptions())
	r.Fit, r.FitNS = fit, time.Since(started).Nanoseconds()
	if e != nil || m.Temperature() != 1 {
		return r, errContrast
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r.GoHeapAfterFit = memory.HeapAlloc
	report, p, ns, e := evaluateAndSave(out, name+".internal-dev", m, dev)
	r.Dev, r.DevNS = report, ns
	if e != nil {
		return r, e
	}
	var q []statehint.Prediction
	if validation != nil {
		report, scores, ns, e := evaluateAndSave(out, name+".exposed-validation", m, validation.corpus)
		if e != nil {
			return r, e
		}
		r.Diagnostic, r.DiagnosticNS = &report, ns
		q = scores
	}
	var artifact bytes.Buffer
	if m.Save(&artifact) != nil {
		return r, errContrast
	}
	r.Model = pin{name + ".rsh", digest(artifact.Bytes())}
	r.ArtifactBytes = artifact.Len()
	if writeBytes(out, r.Model.Path, artifact.Bytes()) != nil {
		return r, errContrast
	}
	saved, e := readPin(out, r.Model, statehintwide.ArtifactBytes)
	if e != nil {
		return r, e
	}
	loaded, e := statehintwide.Load(bytes.NewReader(saved))
	if e != nil || loaded.Temperature() != 1 || loaded.TrainingSteps() != m.TrainingSteps() || !parity(loaded, dev, p) {
		return r, errContrast
	}
	if validation != nil && !parity(loaded, validation.corpus, q) {
		return r, errContrast
	}
	var again bytes.Buffer
	if loaded.Save(&again) != nil || !bytes.Equal(again.Bytes(), artifact.Bytes()) {
		return r, errContrast
	}
	r.ReloadParity = true
	r.DevCalls = 2 * len(p)
	r.DiagnosticCalls = 2 * len(q)
	return r, writeJSON(out, name+".report.json", r)
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
	r = result{Schema: "riido-statehint-controlled-completion-contrast-v1", Status: "started", DevelopmentOnly: true, Recipe: fixedRecipe(), Train: p.Train.pin, Augmentation: p.Augmentation.pin, Split: p.Split.pin, Audit: p.Audit.pin, Arms: []armReport{}, SelectionSource: "internal-dev177 only; eligible severity cost, eight NLL, fixed arm order", GoVersion: runtime.Version(), RSSNote: "Go heap is not OS RSS; measure process peak RSS externally with this fixed run. No native/GPU runtime."}
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
	}{{"train840.source.jsonl", p.Train.bytes}, {"augmentation64.source.jsonl", p.Augmentation.bytes}, {"frozen-split.source.json", p.Split.bytes}, {"source-groups.source.json", p.Audit.bytes}} {
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
	for i, name := range r.Recipe.ArmOrder {
		arm, e := fitArm(out, name, p.Samples[i], p.Dev, p.Validation)
		r.Arms = append(r.Arms, arm)
		if e != nil {
			return r, e
		}
		if arm.Fit.Samples != 1454 || arm.Fit.TrainingSteps != 1840 || arm.Fit.Batches != 1840 {
			return r, errContrast
		}
	}
	r.SelectedArm = selectArm(r.Arms)
	r.Status = "completed; internal-dev selection only; no promotion"
	if r.SelectedArm == "" {
		r.Status = "completed; neither arm eligible; retain previous parent; no promotion"
	}
	err = writeJSON(out, "report.json", r)
	return r, err
}
func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-contrast", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	trainName := f.String("train", "", "explicit pinned original train840 JSONL")
	trainHash := f.String("train-sha256", "", "frozen train840 SHA-256")
	augName := f.String("augmentation", "", "explicit pinned TRAINONLY augmentation64 JSONL")
	augHash := f.String("augmentation-sha256", "", "augmentation SHA-256")
	splitName := f.String("split", "", "frozen train-only663fit177dev metadata")
	splitHash := f.String("split-sha256", "", "frozen split SHA-256")
	auditName := f.String("source-overlay", "", "pinned64-family connected-source audit")
	auditHash := f.String("source-overlay-sha256", "", "source-group audit SHA-256")
	valName := f.String("validation", "", "optional previously exposed validation120 diagnostic")
	valHash := f.String("validation-sha256", "", "optional diagnostic SHA-256")
	output := f.String("out", "", "new private run directory")
	check := f.Bool("check", false, "check inputs only, no model calls or outputs")
	fit := f.Bool("fit", false, "explicit fixed fresh two-arm comparison")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-contrast --train FILE --train-sha256 SHA --augmentation FILE --augmentation-sha256 SHA --split FILE --split-sha256 SHA --source-overlay FILE --source-overlay-sha256 SHA --check\nOptional --validation FILE --validation-sha256 SHA is diagnostic only. Use authorized --fit --out .cache/statehint-completion-contrast/NEW-RUN. No calibration, final-test or parent-model input.")
			return e
		}
		return errContrast
	}
	trainPin, augPin, splitPin, auditPin := pin{*trainName, *trainHash}, pin{*augName, *augHash}, pin{*splitName, *splitHash}, pin{*auditName, *auditHash}
	pins := []pin{trainPin, augPin, splitPin, auditPin}
	if *valName != "" || *valHash != "" {
		pins = append(pins, pin{*valName, *valHash})
	}
	if f.NArg() != 0 || *check == *fit || trainPin.SHA != trainSHA || splitPin.SHA != splitSHA || *check && *output != "" || *fit && (!local(*output) || filepath.Dir(*output) != anchor) {
		return errContrast
	}
	for i, p := range pins {
		if !validPin(p) {
			return errContrast
		}
		for _, old := range pins[:i] {
			if old.Path == p.Path {
				return errContrast
			}
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errContrast
	}
	defer root.Close()
	train, e := loadInput(root, trainPin, "train", 840, balanced(105))
	if e != nil {
		return e
	}
	aug, e := loadInput(root, augPin, "train", 64, classCounts())
	if e != nil {
		return e
	}
	splitBytes, e := readPin(root, splitPin, metadataBudget)
	if e != nil {
		return e
	}
	auditBytes, e := readPin(root, auditPin, metadataBudget)
	if e != nil {
		return e
	}
	var validation *input
	if len(pins) == 5 {
		v, e := loadInput(root, pins[4], "validation", 120, balanced(15))
		if e != nil {
			return e
		}
		validation = &v
	}
	p, e := prepare(train, aug, metadataInput{splitPin, splitBytes}, metadataInput{auditPin, auditBytes}, validation)
	if e != nil {
		return e
	}
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status      string `json:"status"`
			FitFamilies int    `json:"base_fit_families"`
			DevFamilies int    `json:"internal_dev_families"`
			NewFamilies int    `json:"augmentation_families"`
		}{"checked; no fitting, model calls or outputs", 663, 177, 64})
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
	if e := run(os.Args[1:], os.Stdout, os.Stderr); e != nil {
		fmt.Fprintln(os.Stderr, "controlled contrast failed; check explicit pinned training, split, source audit and fresh private output")
		os.Exit(1)
	}
}
