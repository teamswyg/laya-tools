// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only fixed three-claim study. Development scores do not select.
package main

import (
	"bufio"
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
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	anchor      = ".cache/statehint-claims-study"
	inputBudget = 8 << 20
	maxRows     = 4096
	rowSchema   = "riido-three-claims-training-row-v1"
	nllFloor    = 1e-15
)

var errStudy = errors.New("claims study pinned input or private output invalid")

type evidenceInput struct {
	Start  *int   `json:"start_byte"`
	End    *int   `json:"end_byte"`
	Reason string `json:"reason"`
}
type rawRow struct {
	Schema           string          `json:"schema"`
	ID               string          `json:"id"`
	Family           string          `json:"family_id"`
	Group            string          `json:"leakage_group_id"`
	Split            string          `json:"internal_split"`
	Locale           string          `json:"locale"`
	Text             string          `json:"text"`
	Targets          []string        `json:"targets"`
	RubricSHA        string          `json:"rubric_sha256"`
	AnnotationSource string          `json:"annotation_source"`
	Evidence         []evidenceInput `json:"evidence"`
}
type row struct {
	rawRow
	labels [3]statehintclaims.State
}
type counts struct {
	Rows        int    `json:"rows"`
	Families    int    `json:"families"`
	Lineages    int    `json:"lineages"`
	FitRows     int    `json:"fit_rows"`
	DevRows     int    `json:"dev_rows"`
	FitFamilies int    `json:"fit_families"`
	DevFamilies int    `json:"dev_families"`
	RubricSHA   string `json:"rubric_sha256"`
}
type prepared struct {
	rows   []row
	fit    []statehintclaims.Sample
	dev    []int
	counts counts
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func validSHA(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == sha256.Size && hex.EncodeToString(b) == s
}
func local(name string) bool {
	return filepath.IsLocal(name) && name != "." && filepath.Clean(name) == name && !strings.Contains(name, "\\")
}
func bounded(s string, limit int) bool { return len(s) <= limit && utf8.ValidString(s) }
func boundary(text string, i int) bool {
	return i >= 0 && i <= len(text) && (i == len(text) || utf8.RuneStart(text[i]))
}
func readInput(root *os.Root, name, sha string) ([]byte, error) {
	if !local(name) || !validSHA(sha) {
		return nil, errStudy
	}
	before, err := root.Lstat(name)
	if err != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > inputBudget {
		return nil, errStudy
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, errStudy
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errStudy
	}
	b, err := io.ReadAll(io.LimitReader(f, inputBudget+1))
	if err != nil || len(b) > inputBudget || digest(b) != sha {
		return nil, errStudy
	}
	return b, nil
}

func prepare(data []byte) (prepared, error) {
	var p prepared
	if len(data) < 1 || len(data) > inputBudget || !utf8.Valid(data) {
		return p, errStudy
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 64<<10)
	var features statehintwide.Workspace
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if len(p.rows) >= maxRows {
			return prepared{}, errStudy
		}
		d := json.NewDecoder(bytes.NewReader(line))
		d.DisallowUnknownFields()
		var r row
		if d.Decode(&r.rawRow) != nil {
			return prepared{}, errStudy
		}
		var extra any
		if d.Decode(&extra) != io.EOF || r.Schema != rowSchema || r.AnnotationSource != "ai_semantic_reference" || !validSHA(r.RubricSHA) || r.ID == "" || r.Family == "" || r.Group == "" || !bounded(r.ID, 128) || !bounded(r.Family, 128) || !bounded(r.Group, 128) || (r.Split != "fit" && r.Split != "dev") || (r.Locale != "ko" && r.Locale != "en") || r.ID != r.Family+"-"+r.Locale || len(r.Targets) != 3 || len(r.Evidence) != 3 || !bounded(r.Text, statehintclaims.MaxTextBytes) || strings.TrimSpace(r.Text) == "" {
			return prepared{}, errStudy
		}
		if p.counts.RubricSHA == "" {
			p.counts.RubricSHA = r.RubricSHA
		} else if p.counts.RubricSHA != r.RubricSHA {
			return prepared{}, errStudy
		}
		for h, target := range r.Targets {
			label, ok := statehintclaims.ParseState(target)
			e := r.Evidence[h]
			if !ok || e.Start == nil || e.End == nil || *e.Start > *e.End || !boundary(r.Text, *e.Start) || !boundary(r.Text, *e.End) || strings.TrimSpace(e.Reason) == "" || !bounded(e.Reason, 512) {
				return prepared{}, errStudy
			}
			if label != statehintclaims.False && *e.End <= *e.Start || label == statehintclaims.False && *e.Start == *e.End && *e.Start != 0 {
				return prepared{}, errStudy
			}
			if *e.End > *e.Start && strings.TrimSpace(r.Text[*e.Start:*e.End]) == "" {
				return prepared{}, errStudy
			}
			r.labels[h] = label
		}
		view, err := statehintwide.ExtractContextual(r.Text, &features)
		if err != nil || view.WordCount() == 0 {
			return prepared{}, errStudy
		}
		p.rows = append(p.rows, r)
	}
	if scanner.Err() != nil || len(p.rows) == 0 {
		return prepared{}, errStudy
	}
	order := make([]int, len(p.rows))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return p.rows[order[i]].ID < p.rows[order[j]].ID })
	for i := 1; i < len(order); i++ {
		if p.rows[order[i-1]].ID == p.rows[order[i]].ID {
			return prepared{}, errStudy
		}
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := p.rows[order[i]], p.rows[order[j]]
		if a.Family == b.Family {
			return a.Locale < b.Locale
		}
		return a.Family < b.Family
	})
	for start := 0; start < len(order); {
		end := start + 1
		for end < len(order) && p.rows[order[end]].Family == p.rows[order[start]].Family {
			end++
		}
		if end-start != 2 {
			return prepared{}, errStudy
		}
		a, b := p.rows[order[start]], p.rows[order[start+1]]
		if a.Locale != "en" || b.Locale != "ko" || a.Group != b.Group || a.Split != b.Split {
			return prepared{}, errStudy
		}
		p.counts.Families++
		if a.Split == "fit" {
			p.counts.FitFamilies++
		} else {
			p.counts.DevFamilies++
		}
		start = end
	}
	sort.Slice(order, func(i, j int) bool { return p.rows[order[i]].Group < p.rows[order[j]].Group })
	for i, index := range order {
		if i == 0 || p.rows[index].Group != p.rows[order[i-1]].Group {
			p.counts.Lineages++
		} else if p.rows[index].Split != p.rows[order[i-1]].Split {
			return prepared{}, errStudy
		}
	}
	// Retain the pinned input's fit-row order, without inspecting dev labels in Fit.
	for i, r := range p.rows {
		if r.Split == "fit" {
			p.fit = append(p.fit, statehintclaims.Sample{Text: r.Text, Targets: r.labels})
		} else {
			p.dev = append(p.dev, i)
		}
	}
	p.counts.Rows, p.counts.FitRows, p.counts.DevRows = len(p.rows), len(p.fit), len(p.dev)
	if p.counts.FitRows == 0 || p.counts.DevRows == 0 {
		return prepared{}, errStudy
	}
	return p, nil
}

type localeReport struct {
	Rows                int          `json:"rows"`
	RawConfusion        [3][3][3]int `json:"raw_confusion_head_target_winner"`
	MeanCE              float64      `json:"mean_three_head_cross_entropy"`
	ClippedTargets      [3]int       `json:"target_probabilities_clipped_at_floor"`
	TrueTargetFamilies  [3]int       `json:"true_target_families"`
	TrueTargetLineages  [3]int       `json:"true_target_lineages"`
	TrueProposed        [3]int       `json:"gated_true_proposed"`
	TrueCorrect         [3]int       `json:"gated_true_correct"`
	TrueCorrectFamilies [3]int       `json:"gated_true_correct_families"`
	TrueCorrectLineages [3]int       `json:"gated_true_correct_lineages"`
	Precision           [3]float64   `json:"gated_true_precision"`
	PrecisionDefined    [3]bool      `json:"gated_true_precision_defined"`
	UnknownReasons      [3][5]int    `json:"emitted_unknown_reasons"`
}
type devReport struct {
	Rows             int             `json:"rows"`
	MeanCE           float64         `json:"mean_three_head_cross_entropy"`
	ProbabilityFloor float64         `json:"cross_entropy_probability_floor"`
	ReasonOrder      [5]string       `json:"unknown_reason_order"`
	Locales          [2]localeReport `json:"locales_ko_en"`
}

func unique(v []string) int {
	sort.Strings(v)
	n := 0
	for i, s := range v {
		if i == 0 || s != v[i-1] {
			n++
		}
	}
	return n
}
func evaluate(m *statehintclaims.Model, p prepared) (devReport, []statehintclaims.Prediction, error) {
	r := devReport{Rows: len(p.dev), ProbabilityFloor: nllFloor, ReasonOrder: [5]string{"semantic_unknown", "low_confidence", "low_margin", "untrained", "no_word_content"}}
	var targetGroups, correctGroups [2][3][]string
	predictions := make([]statehintclaims.Prediction, 0, len(p.dev))
	var w statehintclaims.Workspace
	for _, index := range p.dev {
		input := p.rows[index]
		prediction, err := m.Predict(input.Text, &w)
		if err != nil || prediction.Source != statehintclaims.Learned || prediction.TrainingSteps != m.TrainingSteps() || prediction.TrainingSteps == 0 {
			return devReport{}, nil, errStudy
		}
		predictions = append(predictions, prediction)
		locale := 0
		if input.Locale == "en" {
			locale = 1
		}
		l := &r.Locales[locale]
		l.Rows++
		for h, head := range prediction.Heads {
			target, _ := statehintclaims.StateIndex(input.labels[h])
			winner, ok := statehintclaims.StateIndex(head.Winner)
			if !ok {
				return devReport{}, nil, errStudy
			}
			l.RawConfusion[h][target][winner]++
			prob := head.Probabilities[target]
			if prob < nllFloor {
				l.ClippedTargets[h]++
			}
			l.MeanCE -= math.Log(math.Max(prob, nllFloor)) / 3
			for reason, name := range r.ReasonOrder {
				if head.UnknownReason == name {
					l.UnknownReasons[h][reason]++
				}
			}
			if input.labels[h] == statehintclaims.True {
				l.TrueTargetFamilies[h]++
				targetGroups[locale][h] = append(targetGroups[locale][h], input.Group)
			}
			if head.State == statehintclaims.True {
				l.TrueProposed[h]++
				if input.labels[h] == statehintclaims.True {
					l.TrueCorrect[h]++
					l.TrueCorrectFamilies[h]++ // Exactly one row per family/locale.
					correctGroups[locale][h] = append(correctGroups[locale][h], input.Group)
				}
			}
		}
	}
	for locale := range r.Locales {
		l := &r.Locales[locale]
		r.MeanCE += l.MeanCE
		if l.Rows > 0 {
			l.MeanCE /= float64(l.Rows)
		}
		for h := range l.TrueProposed {
			l.TrueTargetLineages[h] = unique(targetGroups[locale][h])
			l.TrueCorrectLineages[h] = unique(correctGroups[locale][h])
			if l.TrueProposed[h] > 0 {
				l.PrecisionDefined[h] = true
				l.Precision[h] = float64(l.TrueCorrect[h]) / float64(l.TrueProposed[h])
			}
		}
	}
	r.MeanCE /= float64(r.Rows)
	return r, predictions, nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errStudy
	}
	for _, dir := range []string{".cache", anchor} {
		if err := root.Mkdir(dir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, errStudy
		}
		st, err := root.Lstat(dir)
		if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || dir == anchor && st.Mode().Perm() != 0700 {
			return nil, errStudy
		}
	}
	if err := root.Mkdir(name, 0700); err != nil {
		return nil, errStudy
	}
	return root.OpenRoot(name)
}
func writeBytes(root *os.Root, name string, data []byte) error {
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errStudy
	}
	n, err := f.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errStudy
	}
	return nil
}

type studyReport struct {
	Schema           string                    `json:"schema"`
	Status           string                    `json:"status"`
	InputSHA         string                    `json:"input_sha256"`
	AnnotationSource string                    `json:"annotation_source"`
	Counts           counts                    `json:"input_counts"`
	Heads            [3]string                 `json:"head_order"`
	States           [3]statehintclaims.State  `json:"state_order"`
	Recipe           recipe                    `json:"fixed_recipe"`
	Fit              statehintclaims.FitReport `json:"fit"`
	Dev              devReport                 `json:"development_diagnostic"`
	ModelSHA         string                    `json:"model_sha256"`
	ArtifactBytes    int                       `json:"artifact_bytes"`
	ReloadExact      bool                      `json:"trained_reload_exact_parity"`
	DevCalls         int                       `json:"dev_prediction_calls"`
	ReloadCalls      int                       `json:"reload_prediction_calls"`
	FitCalls         int                       `json:"actual_fit_calls"`
	ElapsedNS        int64                     `json:"fit_evaluation_save_reload_nanoseconds"`
	GoHeap           uint64                    `json:"go_heap_alloc_bytes_not_os_rss"`
	CPUProfile       bool                      `json:"private_cpu_profile_written"`
	Qualified        bool                      `json:"semantic_quality_qualified"`
	Selection        bool                      `json:"selection_performed"`
	Calibration      bool                      `json:"probability_calibration_performed"`
	StateWrites      int                       `json:"application_state_writes"`
}
type recipe struct {
	Epochs      int        `json:"epochs"`
	Batch       int        `json:"batch_size"`
	Rate        float64    `json:"learning_rate"`
	Decay       float64    `json:"weight_decay"`
	Seed        int64      `json:"seed"`
	Temperature float64    `json:"temperature"`
	Gate        [2]float64 `json:"confidence_margin"`
}

func fixedRecipe() recipe {
	return recipe{40, 32, .001, .01, 1729, 1, [2]float64{statehintclaims.ConfidenceFloor, statehintclaims.MarginFloor}}
}
func fitStudy(out *os.Root, p prepared, inputSHA string, cpuProfile bool) (studyReport, error) {
	started := time.Now()
	if cpuProfile {
		f, err := out.OpenFile("cpu.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return studyReport{}, errStudy
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			f.Close()
			return studyReport{}, errStudy
		}
		defer func() { pprof.StopCPUProfile(); f.Close() }()
	}
	m := statehintclaims.NewModel()
	var training statehintclaims.TrainingWorkspace
	fit, err := m.Fit(p.fit, statehintclaims.FitOptions{Epochs: 40, BatchSize: 32, LearningRate: .001, WeightDecay: .01, Seed: 1729}, &training)
	if err != nil || fit.TrainingSteps != uint64(40*((len(p.fit)+31)/32)) {
		return studyReport{}, errStudy
	}
	dev, predictions, err := evaluate(m, p)
	if err != nil {
		return studyReport{}, err
	}
	var b bytes.Buffer
	if m.Save(&b) != nil || b.Len() != statehintclaims.ArtifactBytes || writeBytes(out, "claims.rsc", b.Bytes()) != nil {
		return studyReport{}, errStudy
	}
	f, err := out.Open("claims.rsc")
	if err != nil {
		return studyReport{}, errStudy
	}
	loaded, loadErr := statehintclaims.Load(f)
	closeErr := f.Close()
	if loadErr != nil || closeErr != nil {
		return studyReport{}, errStudy
	}
	var roundtrip bytes.Buffer
	if loaded.Save(&roundtrip) != nil || !bytes.Equal(b.Bytes(), roundtrip.Bytes()) {
		return studyReport{}, errStudy
	}
	var workspace statehintclaims.Workspace
	for i, index := range p.dev {
		prediction, err := loaded.Predict(p.rows[index].Text, &workspace)
		if err != nil || predictions[i] != prediction {
			return studyReport{}, errStudy
		}
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r := studyReport{Schema: "riido-three-claims-study-report-v1", Status: "development_research_only", InputSHA: inputSHA, AnnotationSource: "ai_semantic_reference", Counts: p.counts, States: statehintclaims.States(), Recipe: fixedRecipe(), Fit: fit, Dev: dev, ModelSHA: digest(b.Bytes()), ArtifactBytes: b.Len(), ReloadExact: true, DevCalls: len(p.dev), ReloadCalls: len(p.dev), FitCalls: 1, ElapsedNS: time.Since(started).Nanoseconds(), GoHeap: memory.HeapAlloc, CPUProfile: cpuProfile}
	for i, h := range statehintclaims.Heads() {
		r.Heads[i] = h.String()
	}
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil || writeBytes(out, "report.json", append(encoded, '\n')) != nil {
		return studyReport{}, errStudy
	}
	return r, nil
}

func run(args []string, output, errorOutput io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-claims-study", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	input := f.String("input", "", "explicit canonical JSONL")
	sha := f.String("input-sha256", "", "exact input SHA-256")
	check := f.Bool("check", false, "validate only; no Fit or output directory")
	out := f.String("out", "", "fresh private study directory")
	profile := f.Bool("cpu-profile", false, "private CPU profile of Fit/evaluation")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(errorOutput, "riido-statehint-claims-study --input FILE --input-sha256 SHA --check\nAuthorized fixed Fit: --out .cache/statehint-claims-study/NEW-RUN [--cpu-profile]. No model/recipe, calibration, test or application inputs.")
			return err
		}
		return errStudy
	}
	if f.NArg() != 0 || !local(*input) || !validSHA(*sha) || *check && (*out != "" || *profile) || !*check && (!local(*out) || filepath.Dir(*out) != anchor) {
		return errStudy
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return errStudy
	}
	defer root.Close()
	data, err := readInput(root, *input, *sha)
	if err != nil {
		return err
	}
	p, err := prepare(data)
	if err != nil {
		return err
	}
	if *check {
		return json.NewEncoder(output).Encode(struct {
			Status string `json:"status"`
			Counts counts `json:"counts"`
		}{"checked; no Fit, model predictions or outputs", p.counts})
	}
	private, err := privateOutput(root, *out)
	if err != nil {
		return err
	}
	defer private.Close()
	r, err := fitStudy(private, p, *sha, *profile)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(struct {
		Status    string `json:"status"`
		Qualified bool   `json:"semantic_quality_qualified"`
	}{r.Status, false})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "claims study failed; check pinned canonical input and fresh private output")
		os.Exit(1)
	}
}
