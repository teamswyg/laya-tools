// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only, fixed data-profile x architecture study; diagnostics never select.
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
	anchor         = ".cache/statehint-domain"
	metadataBudget = 1 << 20
)

var errDomain = errors.New("domain study input, checksum or output invalid")

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
	Epochs         int        `json:"epochs"`
	Batch          int        `json:"batch_size"`
	Rate           float64    `json:"learning_rate"`
	Decay          float64    `json:"weight_decay"`
	Seed           int64      `json:"seed"`
	Temperature    float64    `json:"temperature"`
	Gate           [2]float64 `json:"confidence_margin"`
	ArmOrder       [4]string  `json:"fresh_arm_order"`
	RowsPerArm     int        `json:"rows_per_arm"`
	UpdatesPerArm  int        `json:"updates_per_arm"`
	LabelSmoothing float64    `json:"label_smoothing"`
	DrawFamilies   [8]int     `json:"extra_families_per_intent_persisted_order"`
}

func classCounts() [8]int { return balanced(8) }
func fixedRecipe() recipe {
	return recipe{40, 32, .02, .001, 1729, 1, [2]float64{statehintfamily.ConfidenceFloor, statehintfamily.MarginFloor}, [4]string{"linear_balanced_control", "linear_domain64", "mlp16_balanced_control", "mlp16_domain64"}, 1454, 1840, 0, classCounts()}
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
		return nil, errDomain
	}
	before, e := root.Lstat(p.Path)
	if e != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > limit {
		return nil, errDomain
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errDomain
	}
	defer f.Close()
	after, e := f.Stat()
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errDomain
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errDomain
	}
	return b, nil
}
func decodeMetadata(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if d.Decode(v) != nil {
		return errDomain
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return errDomain
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
		return input{}, errDomain
	}
	return input{p, b, c}, nil
}

// Preserve the frozen controlled-run group split and whole-family draw exactly.
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
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errDomain
	}
	a := append([]assignment(nil), s.Assignments...)
	sort.Slice(a, func(i, j int) bool { return a[i].Family < a[j].Family })
	pairs := c.Pairs()
	rows := c.Rows()
	if len(pairs) != len(a) {
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errDomain
	}
	var fit, dev []statehintcorpus.Row
	for i, p := range pairs {
		x := a[i]
		if x.Family != p.Family.ID || x.Group != p.Family.Lineage || x.Use != splitUse(x.Group) {
			return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errDomain
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
		return statehintcorpus.Corpus{}, statehintcorpus.Corpus{}, nil, errDomain
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
			return statehintcorpus.Corpus{}, errDomain
		}
	}
	c, _, e := statehintcorpus.Read(&b, statehintcorpus.Options{Partition: "train", RubricSHA: rubricSHA})
	if e != nil {
		return c, errDomain
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
	if s.Schema != "statehint-dev-comment-domain-source-groups-v1" || s.AugmentationSHA != aug.pin.SHA || s.TrainSHA != trainPin.SHA || s.SplitSHA != splitPin.SHA || len(s.Entries) != 64 {
		return errDomain
	}
	entries := append([]sourceEntry(nil), s.Entries...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Family < entries[j].Family })
	pairs := aug.corpus.Pairs()
	rows := aug.corpus.Rows()
	if len(entries) != len(pairs) {
		return errDomain
	}
	type edge struct{ source, group string }
	var edges []edge
	for i, p := range pairs {
		x := entries[i]
		if x.Family != p.Family.ID || x.Expected != p.Family.Expected || x.Group != p.Family.Lineage || x.Pair != x.Family {
			return errDomain
		}
		for locale, index := range p.Rows {
			if x.RowIDs[locale] != rows[index].ID || x.TextSHA[locale] != digest([]byte(rows[index].Text)) {
				return errDomain
			}
		}
		sources, ok := uniqueSorted(x.Sources)
		if !ok {
			return errDomain
		}
		groups, ok := uniqueSorted(x.Groups)
		if !ok || !contains(groups, x.Group) {
			return errDomain
		}
		members, ok := uniqueSorted(x.Members)
		if !ok {
			return errDomain
		}
		for _, id := range sources {
			v, ok := findAssignment(a, id)
			if !ok || v.Use != "fit" || !contains(groups, v.Group) || !contains(members, id) {
				return errDomain
			}
		}
		for _, id := range members {
			v, ok := findAssignment(a, id)
			if !ok || v.Use != "fit" || !contains(groups, v.Group) {
				return errDomain
			}
		}
		for _, group := range groups {
			found := false
			for _, v := range a {
				if v.Group == group {
					found = true
					if v.Use != "fit" || !contains(members, v.Family) {
						return errDomain
					}
				}
			}
			if !found {
				return errDomain
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
			return errDomain
		}
	}
	// Every entry sharing a canonical augmentation component must describe its
	// full connected original groups/members, not just a local partial branch.
	for i, x := range entries {
		for _, y := range entries[:i] {
			if x.Group != y.Group {
				continue
			}
			xg, _ := uniqueSorted(x.Groups)
			yg, _ := uniqueSorted(y.Groups)
			xm, _ := uniqueSorted(x.Members)
			ym, _ := uniqueSorted(y.Members)
			if !equalStrings(xg, yg) || !equalStrings(xm, ym) {
				return errDomain
			}
		}
	}
	return nil
}
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
				return errDomain
			}
		}
	}
	if _, e := statehintcorpus.ValidateMetadata(metadata); e != nil {
		return errDomain
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
				ranked = append(ranked, rankedPair{p, digest([]byte("dev-comment-domain-control-1729:" + p.Family.ID))})
			}
		}
		if len(ranked) == 0 {
			return nil, nil, errDomain
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
		return prepared{}, errDomain
	}
	var e error
	p.Fit, p.Dev, p.Assignments, e = splitCorpus(train.corpus, s)
	if e != nil || checkSources(aug, p.Assignments, source, train.pin, split.pin) != nil {
		return prepared{}, errDomain
	}
	ins := []input{train, aug}
	if validation != nil {
		ins = append(ins, *validation)
	}
	if disjoint(ins...) != nil {
		return prepared{}, errDomain
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
		return prepared{}, errDomain
	}
	return p, nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errDomain
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return nil, errDomain
		}
		s, e := root.Lstat(dir)
		if e != nil || !s.IsDir() || s.Mode()&os.ModeSymlink != 0 || dir == anchor && s.Mode().Perm() != 0700 {
			return nil, errDomain
		}
	}
	if e := root.Mkdir(name, 0700); e != nil {
		return nil, errDomain
	}
	return root.OpenRoot(name)
}
func writeBytes(root *os.Root, name string, b []byte) error {
	f, e := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return errDomain
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
		return errDomain
	}
	return nil
}
func writeJSON(root *os.Root, name string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return errDomain
	}
	return writeBytes(root, name, append(b, '\n'))
}

type armKind int

const (
	linearArm armKind = iota
	mlpArm
)

type fitReport struct {
	Samples        int     `json:"samples"`
	Epochs         int     `json:"epochs"`
	Batches        int     `json:"batches"`
	TrainingSteps  uint64  `json:"training_steps"`
	MeanLoss       float64 `json:"mean_loss"`
	LabelSmoothing float64 `json:"label_smoothing"`
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
		return m, fitReport{f.Samples, f.Epochs, f.Batches, f.TrainingSteps, f.MeanLoss, 0}, e
	}
	if kind != mlpArm {
		return nil, fitReport{}, errDomain
	}
	m.mlp = statehintmlp.NewModel()
	typed := make([]statehintmlp.Sample, len(training))
	for i, s := range training {
		typed[i] = statehintmlp.Sample{Text: s.Text, Label: s.Label}
	}
	f, e := m.mlp.Fit(typed, statehintmlp.FitOptions{Epochs: r.Epochs, BatchSize: r.Batch, LearningRate: r.Rate, WeightDecay: r.Decay, Seed: r.Seed, LabelSmoothing: 0})
	if f.LabelSmoothing != 0 {
		return nil, fitReport{}, errDomain
	}
	return m, fitReport{f.Samples, f.Epochs, f.Batches, f.TrainingSteps, f.MeanLoss, f.LabelSmoothing}, e
}
func loadModel(kind armKind, b []byte) (*armModel, error) {
	m := &armModel{}
	var e error
	if kind == linearArm {
		m.linear, e = statehintwide.Load(bytes.NewReader(b))
	} else if kind == mlpArm {
		m.mlp, e = statehintmlp.Load(bytes.NewReader(b))
	} else {
		return nil, errDomain
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
	Name               string                  `json:"arm"`
	Architecture       string                  `json:"architecture"`
	DataProfile        string                  `json:"training_data_profile"`
	TrainingSamplesSHA string                  `json:"ordered_training_samples_receipt_sha256"`
	Initialization     string                  `json:"initialization"`
	Seed               int64                   `json:"initialization_seed"`
	ArtifactFormat     string                  `json:"artifact_format"`
	Loader             string                  `json:"loader"`
	Parameters         int                     `json:"parameters"`
	Fit                fitReport               `json:"fit"`
	Dev                statehintfamily.Report  `json:"internal_dev"`
	Diagnostic         *statehintfamily.Report `json:"exposed_validation_diagnostic,omitempty"`
	Model              pin                     `json:"model"`
	ArtifactBytes      int                     `json:"artifact_bytes"`
	FitNS              int64                   `json:"fit_nanoseconds"`
	DevNS              int64                   `json:"internal_dev_prediction_nanoseconds"`
	DiagnosticNS       int64                   `json:"diagnostic_prediction_nanoseconds"`
	ReloadParity       bool                    `json:"trained_save_load_exact_parity"`
	DevCalls           int                     `json:"internal_dev_forwards_including_reload_check"`
	DiagnosticCalls    int                     `json:"exposed_validation_forwards_including_reload_check"`
	GoHeapAfterFit     uint64                  `json:"go_heap_alloc_after_fit_bytes_not_os_rss"`
	ProbabilityOrder   [8]statehint.Intent     `json:"probability_intent_order"`
}
type trainingReceipt struct {
	Profile       string `json:"data_profile"`
	Rows          int    `json:"rows"`
	ClassRows     [8]int `json:"class_rows_in_intent_order"`
	LocaleRows    [2]int `json:"locale_rows_ko_en"`
	SamplesSHA    string `json:"ordered_samples_sha256"`
	ExtraFamilies int    `json:"extra_whole_family_draws_or_new_families"`
}
type result struct {
	Schema                       string             `json:"schema"`
	Status                       string             `json:"status"`
	DevelopmentOnly              bool               `json:"development_only"`
	HumanTruth                   bool               `json:"human_truth"`
	InternalDevPreviouslyExposed bool               `json:"internal_dev_previously_exposed"`
	FreshQualification           bool               `json:"fresh_qualification"`
	ComparisonScope              string             `json:"comparison_scope"`
	Recipe                       recipe             `json:"fixed_recipe"`
	Train                        pin                `json:"train_input"`
	Augmentation                 pin                `json:"augmentation_input"`
	Split                        pin                `json:"frozen_split_input"`
	Audit                        pin                `json:"source_groups_input"`
	Validation                   *pin               `json:"optional_exposed_validation_input,omitempty"`
	DrawSHA                      string             `json:"whole_family_draws_sha256"`
	Training                     [2]trainingReceipt `json:"training_profiles"`
	Arms                         []armReport        `json:"arms"`
	SelectedArm                  string             `json:"selected_arm"`
	SelectionPerformed           bool               `json:"selection_performed"`
	InternalDevSelectionWeight   int                `json:"internal_dev_selection_weight"`
	ValidationSelectionWeight    int                `json:"exposed_validation_selection_weight"`
	Promotion                    bool               `json:"promotion_performed"`
	DeploymentQualified          bool               `json:"deployment_qualified"`
	CalibrationCalls             int                `json:"calibration_forwards"`
	TestCalls                    int                `json:"test_forwards"`
	GoVersion                    string             `json:"go_version"`
	RSSNote                      string             `json:"resource_measurement_note"`
}

func predict(m *armModel, c statehintcorpus.Corpus) ([]statehint.Prediction, error) {
	rows := c.Rows()
	p := make([]statehint.Prediction, len(rows))
	for i, r := range rows {
		v, e := m.predict(r.Text)
		if e != nil {
			return nil, errDomain
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
		return r, nil, duration, errDomain
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
func fitArm(out *os.Root, kind armKind, profile int, training []statehintwide.Sample, dev statehintcorpus.Corpus, validation *input) (armReport, error) {
	if (kind != linearArm && kind != mlpArm) || profile < 0 || profile > 1 {
		return armReport{}, errDomain
	}
	r := armReport{Name: fixedRecipe().ArmOrder[int(kind)*2+profile], Seed: 1729, ProbabilityOrder: statehint.Intents(), Initialization: "zero-initialized-linear", ArtifactFormat: "RSH v2", Loader: "pkg/statehintwide.Load", Parameters: statehintwide.FeatureBins*statehintwide.IntentCount + statehintwide.IntentCount}
	r.Architecture = "Contextual2048 linear"
	r.DataProfile = profileName(profile)
	suffix, expectedBytes := ".rsh", statehintwide.ArtifactBytes
	if kind == mlpArm {
		r.Architecture = "Contextual2048 MLP16 ReLU"
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
		return r, errDomain
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
		return r, errDomain
	}
	r.Model = pin{r.Name + suffix, digest(artifact.Bytes())}
	r.ArtifactBytes = artifact.Len()
	if writeBytes(out, r.Model.Path, artifact.Bytes()) != nil {
		return r, errDomain
	}
	saved, e := readPin(out, r.Model, int64(expectedBytes))
	if e != nil {
		return r, e
	}
	loaded, e := loadModel(kind, saved)
	if e != nil || loaded.temperature() != 1 || loaded.steps() != m.steps() || !parity(loaded, dev, p) || validation != nil && !parity(loaded, validation.corpus, q) {
		return r, errDomain
	}
	var again bytes.Buffer
	if loaded.save(&again) != nil || !bytes.Equal(again.Bytes(), artifact.Bytes()) {
		return r, errDomain
	}
	r.ReloadParity = true
	r.DevCalls = 2 * len(p)
	r.DiagnosticCalls = 2 * len(q)
	return r, nil
}
func profileName(i int) string {
	if i == 0 {
		return "balanced_old_fit_control"
	}
	return "new_domain_training64"
}

type sampleReceipt struct {
	Index   int              `json:"index"`
	Origin  string           `json:"origin"`
	Family  string           `json:"family_id"`
	Group   string           `json:"leakage_group_id"`
	RowID   string           `json:"row_id"`
	Locale  string           `json:"locale"`
	Label   statehint.Intent `json:"label"`
	TextSHA string           `json:"text_sha256"`
}

func orderedReceipt(p prepared, profile int) ([]sampleReceipt, trainingReceipt, error) {
	if profile < 0 || profile > 1 {
		return nil, trainingReceipt{}, errDomain
	}
	receipt := make([]sampleReceipt, 0, len(p.Samples[profile]))
	meta := trainingReceipt{Profile: profileName(profile), ExtraFamilies: 64}
	add := func(row statehintcorpus.Row, origin string) error {
		index, ok := statehint.IntentIndex(row.Expected)
		if !ok || len(receipt) >= len(p.Samples[profile]) {
			return errDomain
		}
		sample := p.Samples[profile][len(receipt)]
		if sample.Label != row.Expected || sample.Text != row.Text {
			return errDomain
		}
		locale := 0
		if row.Locale == "en" {
			locale = 1
		} else if row.Locale != "ko" {
			return errDomain
		}
		meta.ClassRows[index]++
		meta.LocaleRows[locale]++
		receipt = append(receipt, sampleReceipt{len(receipt), origin, row.Family, row.Lineage, row.ID, row.Locale, row.Expected, digest([]byte(row.Text))})
		return nil
	}
	addCorpus := func(c statehintcorpus.Corpus, origin string) error {
		rows := c.Rows()
		for _, pair := range c.Pairs() {
			for _, index := range pair.Rows {
				if e := add(rows[index], origin); e != nil {
					return e
				}
			}
		}
		return nil
	}
	if e := addCorpus(p.Fit, "original_fit663"); e != nil {
		return nil, meta, e
	}
	if profile == 0 {
		pairs, rows := p.Fit.Pairs(), p.Fit.Rows()
		for _, d := range p.Draws {
			i := sort.Search(len(pairs), func(i int) bool { return pairs[i].Family.ID >= d.Family })
			if i == len(pairs) || pairs[i].Family.ID != d.Family {
				return nil, meta, errDomain
			}
			for _, index := range pairs[i].Rows {
				if e := add(rows[index], "old_fit_whole_family_draw"); e != nil {
					return nil, meta, e
				}
			}
		}
	} else if e := addCorpus(p.Augmentation.corpus, "new_domain64"); e != nil {
		return nil, meta, e
	}
	if len(receipt) != len(p.Samples[profile]) {
		return nil, meta, errDomain
	}
	meta.Rows = len(receipt)
	raw, e := json.MarshalIndent(receipt, "", "  ")
	if e != nil {
		return nil, meta, errDomain
	}
	meta.SamplesSHA = digest(append(raw, '\n'))
	return receipt, meta, nil
}
func fitAll(out *os.Root, p prepared) (r result, err error) {
	r = result{Schema: "riido-statehint-domain-development-v1", Status: "started", DevelopmentOnly: true, InternalDevPreviouslyExposed: true, ComparisonScope: "fixed2x2 data-profile x architecture bundles; within-architecture data effect is descriptive, not causality or fresh gold; both diagnostic sets select0", Recipe: fixedRecipe(), Train: p.Train.pin, Augmentation: p.Augmentation.pin, Split: p.Split.pin, Audit: p.Audit.pin, Arms: []armReport{}, GoVersion: runtime.Version(), RSSNote: "Go heap is not process peak RSS. Measure OS peak RSS externally for the whole sequential four-arm run; no native/GPU runtime."}
	if p.Validation != nil {
		v := p.Validation.pin
		r.Validation = &v
	}
	defer func() {
		if err != nil {
			r.Status = "failed; no selection or promotion; partial artifacts retained"
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
	}{
		{"train840.source.jsonl", p.Train.bytes}, {"domain64.source.jsonl", p.Augmentation.bytes}, {"frozen-split.source.json", p.Split.bytes}, {"source-groups.audit.json", p.Audit.bytes},
	} {
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
	raw, e := json.MarshalIndent(p.Draws, "", "  ")
	if e != nil {
		return r, errDomain
	}
	r.DrawSHA = digest(append(raw, '\n'))
	for profile := 0; profile < 2; profile++ {
		receipt, meta, e := orderedReceipt(p, profile)
		if e != nil {
			return r, e
		}
		r.Training[profile] = meta
		if err = writeJSON(out, profileName(profile)+".ordered-samples.json", receipt); err != nil {
			return r, err
		}
	}
	if r.Training[0].ClassRows != r.Training[1].ClassRows || r.Training[0].LocaleRows != r.Training[1].LocaleRows {
		return r, errDomain
	}
	for _, kind := range []armKind{linearArm, mlpArm} {
		for profile := 0; profile < 2; profile++ {
			arm, e := fitArm(out, kind, profile, p.Samples[profile], p.Dev, p.Validation)
			arm.TrainingSamplesSHA = r.Training[profile].SamplesSHA
			r.Arms = append(r.Arms, arm)
			if e != nil {
				return r, e
			}
			if arm.Fit.Samples != 1454 || arm.Fit.TrainingSteps != 1840 || arm.Fit.Batches != 1840 || arm.Fit.LabelSmoothing != 0 {
				return r, errDomain
			}
			if err = writeJSON(out, arm.Name+".report.json", arm); err != nil {
				return r, err
			}
		}
	}
	r.Status = "completed; four fixed arms, diagnostic-only; no selection or promotion"
	err = writeJSON(out, "report.json", r)
	return r, err
}

func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-domain", flag.ContinueOnError)
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
	fit := f.Bool("fit", false, "explicit fixed fresh four-arm comparison")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-domain --train FILE --train-sha256 SHA --augmentation FILE --augmentation-sha256 SHA --split FILE --split-sha256 SHA --source-overlay FILE --source-overlay-sha256 SHA --check\nOptional --validation FILE --validation-sha256 SHA is diagnostic only. Use authorized --fit --out .cache/statehint-domain/NEW-RUN. No calibration, final-test or parent-model input.")
			return e
		}
		return errDomain
	}
	trainPin, augPin, splitPin, auditPin := pin{*trainName, *trainHash}, pin{*augName, *augHash}, pin{*splitName, *splitHash}, pin{*auditName, *auditHash}
	pins := []pin{trainPin, augPin, splitPin, auditPin}
	if *valName != "" || *valHash != "" {
		pins = append(pins, pin{*valName, *valHash})
	}
	if f.NArg() != 0 || *check == *fit || trainPin.SHA != trainSHA || splitPin.SHA != splitSHA || *check && *output != "" || *fit && (!local(*output) || filepath.Dir(*output) != anchor) {
		return errDomain
	}
	for i, p := range pins {
		if !validPin(p) {
			return errDomain
		}
		for _, old := range pins[:i] {
			if old.Path == p.Path {
				return errDomain
			}
		}
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errDomain
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
		Selected  string `json:"selected_arm"`
		Promotion bool   `json:"promotion_performed"`
	}{r.Status, r.SelectedArm, false})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "domain study failed; check pinned training, source audit, frozen split and fresh private output")
		os.Exit(1)
	}
}
