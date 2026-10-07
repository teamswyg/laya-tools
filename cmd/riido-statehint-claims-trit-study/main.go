// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only fixed ternary study. Development diagnostics do not select.
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
	"github.com/teamswyg/laya-tools/pkg/statehintclaimtrit"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const (
	anchor      = ".cache/statehint-claims-trit-study"
	inputBudget = 8 << 20
	maxRows     = 4096
	rowSchema   = "riido-three-claims-training-row-v1"
	nllFloor    = 1e-15
	parentBytes = 73988
	tritBytes   = 4023
	devRows     = 240
)

var errStudy = errors.New("ternary claims study pinned input or private output invalid")

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
	FitLineages int    `json:"fit_lineages"`
	DevLineages int    `json:"dev_lineages"`
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
	return len(name) <= 4096 && filepath.IsLocal(name) && name != "." && filepath.Clean(name) == name && !strings.Contains(name, "\\")
}
func bounded(s string, limit int) bool { return len(s) <= limit && utf8.ValidString(s) }
func boundary(text string, i int) bool {
	return i >= 0 && i <= len(text) && (i == len(text) || utf8.RuneStart(text[i]))
}

// Hold a root for every opened path component. os.Root prevents escape, while
// Lstat/SameFile reject internal symlinks and replacements during each open.
func openDirectory(root *os.Root, name string) (*os.Root, error) {
	before, err := root.Lstat(name)
	if err != nil || !before.IsDir() || before.Mode()&os.ModeSymlink != 0 {
		return nil, errStudy
	}
	child, err := root.OpenRoot(name)
	if err != nil {
		return nil, errStudy
	}
	after, err := child.Stat(".")
	if err != nil || !after.IsDir() || !os.SameFile(before, after) {
		child.Close()
		return nil, errStudy
	}
	return child, nil
}
func openInput(root *os.Root, name string, limit int, exact int) (*os.File, error) {
	if !local(name) {
		return nil, errStudy
	}
	parts := strings.Split(name, string(filepath.Separator))
	current := root
	var owned *os.Root
	defer func() {
		if owned != nil {
			owned.Close()
		}
	}()
	for _, part := range parts[:len(parts)-1] {
		next, err := openDirectory(current, part)
		if err != nil {
			return nil, errStudy
		}
		if owned != nil {
			owned.Close()
		}
		owned, current = next, next
	}
	leaf := parts[len(parts)-1]
	before, err := current.Lstat(leaf)
	if err != nil || !before.Mode().IsRegular() || before.Size() < 1 || before.Size() > int64(limit) || exact > 0 && before.Size() != int64(exact) {
		return nil, errStudy
	}
	f, err := current.Open(leaf)
	if err != nil {
		return nil, errStudy
	}
	after, err := f.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) || before.Size() != after.Size() {
		f.Close()
		return nil, errStudy
	}
	return f, nil
}
func readPinned(root *os.Root, name, sha string, limit, exact int) ([]byte, error) {
	if !validSHA(sha) {
		return nil, errStudy
	}
	f, err := openInput(root, name, limit, exact)
	if err != nil {
		return nil, errStudy
	}
	b, readErr := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	st, statErr := f.Stat()
	closeErr := f.Close()
	if readErr != nil || statErr != nil || closeErr != nil || len(b) > limit || len(b) < 1 || exact > 0 && len(b) != exact || st.Size() != int64(len(b)) || digest(b) != sha {
		return nil, errStudy
	}
	return b, nil
}

// Check verifies pinned bytes, expected container type/size and checksum without
// constructing or projecting a model. The actual run additionally uses Load's
// full numeric/header validation before creating an output directory.
func parentShape(data []byte) bool {
	if len(data) != parentBytes || parentBytes != statehintclaims.ArtifactBytes || string(data[:4]) != "RSC\x00" {
		return false
	}
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	return bytes.Equal(sum[:], data[end:])
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
			if p.rows[index].Split == "fit" {
				p.counts.FitLineages++
			} else {
				p.counts.DevLineages++
			}
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

// Observations are ephemeral numeric snapshots, never serialized by row.
type observation struct {
	Heads         [3]statehintclaims.HeadPrediction
	Logits        [3][3]float64
	Probabilities [3][3]float64
	WordCount     int
}

func observeFloat(m *statehintclaims.Model, p prepared) ([]observation, []statehintclaims.Prediction, []statehintclaims.ScoreResult, error) {
	result := make([]observation, len(p.dev))
	predictions := make([]statehintclaims.Prediction, len(p.dev))
	scores := make([]statehintclaims.ScoreResult, len(p.dev))
	var w statehintclaims.Workspace
	for i, index := range p.dev {
		text := p.rows[index].Text
		prediction, err := m.Predict(text, &w)
		if err != nil || prediction.Source != statehintclaims.Learned || prediction.TrainingSteps != m.TrainingSteps() {
			return nil, nil, nil, errStudy
		}
		score, err := m.Scores(text, &w)
		if err != nil || score.WordCount == 0 {
			return nil, nil, nil, errStudy
		}
		predictions[i], scores[i] = prediction, score
		result[i] = observation{prediction.Heads, score.Logits, score.Probabilities, score.WordCount}
	}
	return result, predictions, scores, nil
}
func observeTrit(m *statehintclaimtrit.Model, p prepared) ([]observation, []statehintclaimtrit.Prediction, []statehintclaimtrit.ScoreResult, error) {
	result := make([]observation, len(p.dev))
	predictions := make([]statehintclaimtrit.Prediction, len(p.dev))
	scores := make([]statehintclaimtrit.ScoreResult, len(p.dev))
	var w statehintclaimtrit.Workspace
	meta := m.Metadata()
	for i, index := range p.dev {
		text := p.rows[index].Text
		prediction, err := m.Predict(text, &w)
		if err != nil || prediction.Source != statehintclaimtrit.Learned || prediction.TrainingSteps != meta.TrainingSteps || prediction.Mode != meta.Mode || prediction.ParentSHA256 != meta.ParentSHA256 {
			return nil, nil, nil, errStudy
		}
		score, err := m.Scores(text, &w)
		if err != nil || score.WordCount == 0 {
			return nil, nil, nil, errStudy
		}
		predictions[i], scores[i] = prediction, score
		result[i] = observation{prediction.Heads, score.Logits, score.Probabilities, score.WordCount}
	}
	return result, predictions, scores, nil
}
func evaluate(p prepared, observed []observation) (devReport, error) {
	if len(observed) != len(p.dev) || len(observed) == 0 {
		return devReport{}, errStudy
	}
	r := devReport{Rows: len(p.dev), ProbabilityFloor: nllFloor, ReasonOrder: [5]string{"semantic_unknown", "low_confidence", "low_margin", "untrained", "no_word_content"}}
	var targetGroups, correctGroups [2][3][]string
	for i, index := range p.dev {
		input, prediction := p.rows[index], observed[i]
		locale := 0
		if input.Locale == "en" {
			locale = 1
		}
		l := &r.Locales[locale]
		l.Rows++
		for h, head := range prediction.Heads {
			target, ok := statehintclaims.StateIndex(input.labels[h])
			winner, winnerOK := statehintclaims.StateIndex(head.Winner)
			_, stateOK := statehintclaims.StateIndex(head.State)
			prob := prediction.Probabilities[h][target]
			if !ok || !winnerOK || !stateOK || math.IsNaN(prob) || math.IsInf(prob, 0) || prob < 0 || prob > 1 || head.Probabilities != prediction.Probabilities[h] {
				return devReport{}, errStudy
			}
			l.RawConfusion[h][target][winner]++
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
					l.TrueCorrectFamilies[h]++
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
	return r, nil
}

type localeDifference struct {
	Rows                   int          `json:"rows"`
	RawWinnerChanges       [3]int       `json:"raw_winner_changes"`
	RawWinnerTransitions   [3][3][3]int `json:"raw_winner_transitions_head_float_trit"`
	GateStateChanges       [3]int       `json:"gated_state_changes"`
	GateStateTransitions   [3][3][3]int `json:"gated_state_transitions_head_float_trit"`
	TrueAdded              [3]int       `json:"gated_true_added"`
	TrueRemoved            [3]int       `json:"gated_true_removed"`
	TrueAddedOnFalse       [3]int       `json:"gated_true_added_on_false_target"`
	TrueAddedOnUnknown     [3]int       `json:"gated_true_added_on_unknown_target"`
	MaxAbsLogitDelta       [3]float64   `json:"max_absolute_logit_delta"`
	MaxAbsProbabilityDelta [3]float64   `json:"max_absolute_probability_delta"`
}
type differenceReport struct {
	Rows    int                 `json:"rows"`
	Locales [2]localeDifference `json:"locales_ko_en"`
}

func compare(p prepared, baseline, changed []observation) (differenceReport, error) {
	if len(baseline) != len(p.dev) || len(changed) != len(p.dev) {
		return differenceReport{}, errStudy
	}
	r := differenceReport{Rows: len(p.dev)}
	for i, index := range p.dev {
		input := p.rows[index]
		locale := 0
		if input.Locale == "en" {
			locale = 1
		}
		l := &r.Locales[locale]
		l.Rows++
		for h := 0; h < 3; h++ {
			before, after := baseline[i].Heads[h], changed[i].Heads[h]
			bw, bOK := statehintclaims.StateIndex(before.Winner)
			aw, aOK := statehintclaims.StateIndex(after.Winner)
			bs, bsOK := statehintclaims.StateIndex(before.State)
			as, asOK := statehintclaims.StateIndex(after.State)
			if !bOK || !aOK || !bsOK || !asOK {
				return differenceReport{}, errStudy
			}
			l.RawWinnerTransitions[h][bw][aw]++
			l.GateStateTransitions[h][bs][as]++
			if bw != aw {
				l.RawWinnerChanges[h]++
			}
			if bs != as {
				l.GateStateChanges[h]++
			}
			if before.State != statehintclaims.True && after.State == statehintclaims.True {
				l.TrueAdded[h]++
				if input.labels[h] == statehintclaims.False {
					l.TrueAddedOnFalse[h]++
				}
				if input.labels[h] == statehintclaims.Unknown {
					l.TrueAddedOnUnknown[h]++
				}
			}
			if before.State == statehintclaims.True && after.State != statehintclaims.True {
				l.TrueRemoved[h]++
			}
			for c := 0; c < 3; c++ {
				ld := math.Abs(baseline[i].Logits[h][c] - changed[i].Logits[h][c])
				pd := math.Abs(baseline[i].Probabilities[h][c] - changed[i].Probabilities[h][c])
				if math.IsNaN(ld) || math.IsInf(ld, 0) || math.IsNaN(pd) || math.IsInf(pd, 0) {
					return differenceReport{}, errStudy
				}
				l.MaxAbsLogitDelta[h] = math.Max(l.MaxAbsLogitDelta[h], ld)
				l.MaxAbsProbabilityDelta[h] = math.Max(l.MaxAbsProbabilityDelta[h], pd)
			}
		}
	}
	return r, nil
}

func privateOutput(root *os.Root, name string) (*os.Root, error) {
	if !local(name) || filepath.Dir(name) != anchor {
		return nil, errStudy
	}
	if err := root.Mkdir(".cache", 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, errStudy
	}
	cache, err := openDirectory(root, ".cache")
	if err != nil {
		return nil, errStudy
	}
	defer cache.Close()
	base := filepath.Base(anchor)
	if err := cache.Mkdir(base, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, errStudy
	}
	study, err := openDirectory(cache, base)
	if err != nil {
		return nil, errStudy
	}
	defer study.Close()
	st, err := study.Stat(".")
	if err != nil || st.Mode().Perm() != 0700 {
		return nil, errStudy
	}
	child := filepath.Base(name)
	if err := study.Mkdir(child, 0700); err != nil {
		return nil, errStudy
	}
	out, err := openDirectory(study, child)
	if err != nil {
		return nil, errStudy
	}
	st, err = out.Stat(".")
	if err != nil || st.Mode().Perm() != 0700 {
		out.Close()
		return nil, errStudy
	}
	return out, nil
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

type recipe struct {
	Epochs         int        `json:"epochs"`
	Batch          int        `json:"batch_size"`
	Rate           float64    `json:"learning_rate"`
	Decay          float64    `json:"weight_decay"`
	Seed           int64      `json:"seed"`
	Temperature    float64    `json:"temperature"`
	Gate           [2]float64 `json:"confidence_margin"`
	Initialization string     `json:"initialization"`
	STE            string     `json:"straight_through_estimator"`
}

func fixedRecipe() recipe {
	return recipe{40, 32, .001, .01, 1729, 1, [2]float64{statehintclaims.ConfidenceFloor, statehintclaims.MarginFloor}, "parent_float_copy_fresh_adamw", "identity_dequantized_weight_no_scale_mask_or_scale_gradient"}
}

type parentReport struct {
	SHA                string    `json:"sha256"`
	ArtifactBytes      int       `json:"artifact_bytes"`
	TrainingSteps      uint64    `json:"training_steps"`
	InitializationSeed int64     `json:"initialization_seed"`
	ReadOnlyExact      bool      `json:"read_only_exact_save_parity"`
	Dev                devReport `json:"development_diagnostic"`
	DevCalls           int       `json:"dev_prediction_calls"`
	ScoreCalls         int       `json:"dev_score_calls"`
}
type floatVariantReport struct {
	Mode                     string           `json:"mode"`
	ParentSHA                string           `json:"parent_sha256"`
	BaseTrainingSteps        uint64           `json:"base_training_steps"`
	NewOptimizerSteps        uint64           `json:"new_optimizer_steps"`
	TrainingSteps            uint64           `json:"training_steps"`
	ParentInitializationSeed int64            `json:"parent_initialization_seed"`
	AdaptationSeed           int64            `json:"adaptation_seed"`
	Temperature              float64          `json:"temperature"`
	FeatureSchema            string           `json:"feature_schema"`
	Quantized                bool             `json:"quantized"`
	Dev                      devReport        `json:"development_diagnostic"`
	Difference               differenceReport `json:"difference_from_float_parent"`
	ModelSHA                 string           `json:"model_sha256"`
	ArtifactBytes            int              `json:"artifact_bytes"`
	ReloadByteExact          bool             `json:"reload_exact_byte_parity"`
	ReloadPredictExact       bool             `json:"reload_exact_prediction_parity"`
	ReloadScoresExact        bool             `json:"reload_exact_score_parity"`
	DevCalls                 int              `json:"dev_prediction_calls"`
	ScoreCalls               int              `json:"dev_score_calls"`
	ReloadCalls              int              `json:"reload_prediction_calls"`
	ReloadScoreCalls         int              `json:"reload_score_calls"`
	FitCalls                 int              `json:"actual_fit_calls"`
}
type variantReport struct {
	Metadata           statehintclaimtrit.ModelMetadata `json:"model_metadata"`
	Dev                devReport                        `json:"development_diagnostic"`
	Difference         differenceReport                 `json:"difference_from_float_parent"`
	DifferenceMatched  *differenceReport                `json:"difference_from_matched_float,omitempty"`
	ModelSHA           string                           `json:"model_sha256"`
	ArtifactBytes      int                              `json:"artifact_bytes"`
	ReloadByteExact    bool                             `json:"reload_exact_byte_parity"`
	ReloadPredictExact bool                             `json:"reload_exact_prediction_parity"`
	ReloadScoresExact  bool                             `json:"reload_exact_score_parity"`
	DevCalls           int                              `json:"dev_prediction_calls"`
	ScoreCalls         int                              `json:"dev_score_calls"`
	ReloadCalls        int                              `json:"reload_prediction_calls"`
	ReloadScoreCalls   int                              `json:"reload_score_calls"`
	FitCalls           int                              `json:"actual_fit_calls"`
	ProjectionFitCalls int                              `json:"projection_fit_calls"`
}
type studyReport struct {
	Schema               string                        `json:"schema"`
	Status               string                        `json:"status"`
	InputSHA             string                        `json:"input_sha256"`
	AnnotationSource     string                        `json:"annotation_source"`
	Counts               counts                        `json:"input_counts"`
	Heads                [3]string                     `json:"head_order"`
	States               [3]statehintclaims.State      `json:"state_order"`
	Recipe               recipe                        `json:"fixed_recipe"`
	Parent               parentReport                  `json:"float_parent"`
	MatchedFloat         floatVariantReport            `json:"matched_float"`
	FloatFit             statehintclaims.WarmFitReport `json:"matched_float_fit"`
	PTQ                  variantReport                 `json:"ptq"`
	QAT                  variantReport                 `json:"qat"`
	Fit                  statehintclaimtrit.FitReport  `json:"qat_fit"`
	FitCalls             int                           `json:"actual_new_fit_calls"`
	FloatFitCalls        int                           `json:"actual_new_float_fit_calls"`
	QATFitCalls          int                           `json:"actual_new_qat_fit_calls"`
	ProjectionFitCalls   int                           `json:"projection_fit_calls"`
	DevPreviouslyExposed bool                          `json:"development_diagnostics_previously_exposed"`
	Qualified            bool                          `json:"semantic_quality_qualified"`
	Selection            bool                          `json:"selection_performed"`
	Calibration          bool                          `json:"probability_calibration_performed"`
	CalTestAccessed      bool                          `json:"calibration_or_test_accessed"`
	StateWrites          int                           `json:"application_state_writes"`
	ElapsedNS            int64                         `json:"fit_evaluation_save_reload_nanoseconds"`
	GoHeap               uint64                        `json:"go_heap_alloc_bytes_not_os_rss"`
	OSRSSMeasured        bool                          `json:"os_rss_measured_in_process"`
	OSRSSMeasurement     string                        `json:"os_rss_measurement"`
	CPUProfile           bool                          `json:"private_cpu_profile_written"`
	CPUProfileScope      string                        `json:"private_cpu_profile_scope,omitempty"`
}

func validMetadata(meta statehintclaimtrit.ModelMetadata, parent *statehintclaims.Model, sha, mode string, newSteps uint64) bool {
	return meta.Adaptation == "experimental_weight_only_mixed_precision_three_claim_heads" && meta.FeatureSchema == statehintclaims.FeatureSchema && meta.InformationBitsPerTrit == math.Log2(3) && meta.Mode == mode && meta.ParentSHA256 == sha && meta.BaseTrainingSteps == parent.TrainingSteps() && meta.NewOptimizerSteps == newSteps && meta.TrainingSteps == parent.TrainingSteps()+newSteps && meta.ParentInitializationSeed == parent.InitializationSeed() && meta.AdaptationSeed == 1729 && meta.Temperature == 1 && meta.ConfidenceFloor == statehintclaims.ConfidenceFloor && meta.MarginFloor == statehintclaims.MarginFloor && meta.ScaleFloor == 1e-5 && meta.WeightTritCount == 18432 && meta.PackedWeightBytes == 3687 && meta.ScaleCount == 3 && meta.BiasCount == 9 && meta.PhysicalBitsPerTrit == 1.6 && !meta.BitNetLLM && !meta.W158A8 && !meta.QualityQualified && !meta.StateAuthority
}
func saveMatchedFloat(out *os.Root, m, parent *statehintclaims.Model, parentSHA string, fit statehintclaims.WarmFitReport, p prepared, baseline []observation) (floatVariantReport, []observation, error) {
	observed, predictions, scores, err := observeFloat(m, p)
	if err != nil {
		return floatVariantReport{}, nil, errStudy
	}
	dev, err := evaluate(p, observed)
	if err != nil {
		return floatVariantReport{}, nil, errStudy
	}
	difference, err := compare(p, baseline, observed)
	if err != nil {
		return floatVariantReport{}, nil, errStudy
	}
	var b bytes.Buffer
	if m.Save(&b) != nil || b.Len() != parentBytes || writeBytes(out, "matched_float.rsc", b.Bytes()) != nil {
		return floatVariantReport{}, nil, errStudy
	}
	stored, err := readPinned(out, "matched_float.rsc", digest(b.Bytes()), parentBytes, parentBytes)
	if err != nil {
		return floatVariantReport{}, nil, errStudy
	}
	loaded, err := statehintclaims.Load(bytes.NewReader(stored))
	if err != nil || loaded.TrainingSteps() != m.TrainingSteps() || loaded.InitializationSeed() != m.InitializationSeed() || loaded.Temperature() != m.Temperature() {
		return floatVariantReport{}, nil, errStudy
	}
	var roundtrip bytes.Buffer
	if loaded.Save(&roundtrip) != nil || !bytes.Equal(b.Bytes(), roundtrip.Bytes()) {
		return floatVariantReport{}, nil, errStudy
	}
	var w statehintclaims.Workspace
	for i, index := range p.dev {
		prediction, err := loaded.Predict(p.rows[index].Text, &w)
		if err != nil || prediction != predictions[i] {
			return floatVariantReport{}, nil, errStudy
		}
		score, err := loaded.Scores(p.rows[index].Text, &w)
		if err != nil || score != scores[i] {
			return floatVariantReport{}, nil, errStudy
		}
	}
	r := floatVariantReport{Mode: "matched_float_warm_continuation", ParentSHA: parentSHA, BaseTrainingSteps: fit.BaseTrainingSteps, NewOptimizerSteps: fit.NewOptimizerSteps, TrainingSteps: m.TrainingSteps(), ParentInitializationSeed: parent.InitializationSeed(), AdaptationSeed: fit.Seed, Temperature: m.Temperature(), FeatureSchema: statehintclaims.FeatureSchema, Dev: dev, Difference: difference, ModelSHA: digest(b.Bytes()), ArtifactBytes: b.Len(), ReloadByteExact: true, ReloadPredictExact: true, ReloadScoresExact: true, DevCalls: len(p.dev), ScoreCalls: len(p.dev), ReloadCalls: len(p.dev), ReloadScoreCalls: len(p.dev), FitCalls: 1}
	return r, observed, nil
}
func saveVariant(out *os.Root, name string, m *statehintclaimtrit.Model, p prepared, parent, matched []observation) (variantReport, error) {
	observed, predictions, scores, err := observeTrit(m, p)
	if err != nil {
		return variantReport{}, errStudy
	}
	dev, err := evaluate(p, observed)
	if err != nil {
		return variantReport{}, errStudy
	}
	difference, err := compare(p, parent, observed)
	if err != nil {
		return variantReport{}, errStudy
	}
	var b bytes.Buffer
	if statehintclaimtrit.ArtifactBytes != tritBytes || m.Save(&b) != nil || b.Len() != tritBytes || writeBytes(out, name, b.Bytes()) != nil {
		return variantReport{}, errStudy
	}
	stored, err := readPinned(out, name, digest(b.Bytes()), tritBytes, tritBytes)
	if err != nil {
		return variantReport{}, errStudy
	}
	loaded, err := statehintclaimtrit.Load(bytes.NewReader(stored))
	if err != nil || loaded.Metadata() != m.Metadata() {
		return variantReport{}, errStudy
	}
	var roundtrip bytes.Buffer
	if loaded.Save(&roundtrip) != nil || !bytes.Equal(b.Bytes(), roundtrip.Bytes()) {
		return variantReport{}, errStudy
	}
	var w statehintclaimtrit.Workspace
	for i, index := range p.dev {
		prediction, err := loaded.Predict(p.rows[index].Text, &w)
		if err != nil || prediction != predictions[i] {
			return variantReport{}, errStudy
		}
		score, err := loaded.Scores(p.rows[index].Text, &w)
		if err != nil || score != scores[i] {
			return variantReport{}, errStudy
		}
	}
	var matchedDifference *differenceReport
	if matched != nil {
		d, err := compare(p, matched, observed)
		if err != nil {
			return variantReport{}, errStudy
		}
		matchedDifference = &d
	}
	return variantReport{Metadata: m.Metadata(), Dev: dev, Difference: difference, DifferenceMatched: matchedDifference, ModelSHA: digest(b.Bytes()), ArtifactBytes: b.Len(), ReloadByteExact: true, ReloadPredictExact: true, ReloadScoresExact: true, DevCalls: len(p.dev), ScoreCalls: len(p.dev), ReloadCalls: len(p.dev), ReloadScoreCalls: len(p.dev)}, nil
}
func fitStudy(out *os.Root, parent *statehintclaims.Model, parentData []byte, p prepared, parentSHA, inputSHA string, cpuProfile bool) (studyReport, error) {
	started := time.Now()
	var profile *os.File
	finishProfile := func() error {
		if profile == nil {
			return nil
		}
		pprof.StopCPUProfile()
		err := profile.Close()
		profile = nil
		return err
	}
	if cpuProfile {
		var err error
		profile, err = out.OpenFile("cpu.pprof", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return studyReport{}, errStudy
		}
		if err := pprof.StartCPUProfile(profile); err != nil {
			profile.Close()
			profile = nil
			return studyReport{}, errStudy
		}
		defer finishProfile()
	}
	// Match the extra optimization before evaluating either continuation. Both
	// controls start from the same float parent with fresh moments and the fixed
	// seed/order. The parent and PTQ projection remain independent receivers.
	expected := uint64(40 * ((len(p.fit) + 31) / 32))
	matchedFloat := parent.Clone()
	var floatTraining statehintclaims.TrainingWorkspace
	floatFit, err := matchedFloat.WarmFit(parent, p.fit, &floatTraining)
	if err != nil || floatFit.Samples != len(p.fit) || floatFit.Epochs != 40 || floatFit.Batches != int(expected) || floatFit.BaseTrainingSteps != parent.TrainingSteps() || floatFit.NewOptimizerSteps != expected || floatFit.TrainingSteps != parent.TrainingSteps()+expected || floatFit.Seed != 1729 || floatFit.Initialization != fixedRecipe().Initialization || matchedFloat.TrainingSteps() != floatFit.TrainingSteps || matchedFloat.InitializationSeed() != parent.InitializationSeed() || matchedFloat.Temperature() != 1 {
		return studyReport{}, errStudy
	}
	// The two conversions own separate receivers. Projection makes no Fit call.
	ptq, err := statehintclaimtrit.FromFloat(parent, parentSHA)
	if err != nil || !validMetadata(ptq.Metadata(), parent, parentSHA, "ptq", 0) {
		return studyReport{}, errStudy
	}
	qat, err := statehintclaimtrit.FromFloat(parent, parentSHA)
	if err != nil || !validMetadata(qat.Metadata(), parent, parentSHA, "ptq", 0) {
		return studyReport{}, errStudy
	}
	var training statehintclaimtrit.TrainingWorkspace
	fit, err := qat.WarmFit(parent, p.fit, &training)
	if err != nil || fit.Samples != len(p.fit) || fit.Epochs != 40 || fit.Batches != int(expected) || fit.NewOptimizerSteps != expected || fit.BaseTrainingSteps != parent.TrainingSteps() || fit.TrainingSteps != parent.TrainingSteps()+expected || fit.Seed != 1729 || fit.Initialization != fixedRecipe().Initialization || !validMetadata(qat.Metadata(), parent, parentSHA, "qat", expected) || !validMetadata(ptq.Metadata(), parent, parentSHA, "ptq", 0) {
		return studyReport{}, errStudy
	}
	baseline, _, _, err := observeFloat(parent, p)
	if err != nil {
		return studyReport{}, errStudy
	}
	parentDev, err := evaluate(p, baseline)
	if err != nil {
		return studyReport{}, errStudy
	}
	matchedReport, matchedObservations, err := saveMatchedFloat(out, matchedFloat, parent, parentSHA, floatFit, p, baseline)
	if err != nil {
		return studyReport{}, errStudy
	}
	ptqReport, err := saveVariant(out, "ptq.rqt", ptq, p, baseline, nil)
	if err != nil {
		return studyReport{}, errStudy
	}
	qatReport, err := saveVariant(out, "qat.rqt", qat, p, baseline, matchedObservations)
	if err != nil {
		return studyReport{}, errStudy
	}
	qatReport.FitCalls = 1
	var unchanged bytes.Buffer
	if parent.Save(&unchanged) != nil || !bytes.Equal(parentData, unchanged.Bytes()) {
		return studyReport{}, errStudy
	}
	if finishProfile() != nil {
		return studyReport{}, errStudy
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	r := studyReport{Schema: "riido-three-claims-trit-study-report-v1", Status: "development_research_only", InputSHA: inputSHA, AnnotationSource: "ai_semantic_reference", Counts: p.counts, States: statehintclaims.States(), Recipe: fixedRecipe(), Parent: parentReport{parentSHA, len(parentData), parent.TrainingSteps(), parent.InitializationSeed(), true, parentDev, len(p.dev), len(p.dev)}, MatchedFloat: matchedReport, FloatFit: floatFit, PTQ: ptqReport, QAT: qatReport, Fit: fit, FitCalls: 2, FloatFitCalls: 1, QATFitCalls: 1, DevPreviouslyExposed: true, ElapsedNS: time.Since(started).Nanoseconds(), GoHeap: memory.HeapAlloc, OSRSSMeasurement: "external_process_measurement_required_go_heap_is_not_os_rss", CPUProfile: cpuProfile}
	if cpuProfile {
		r.CPUProfileScope = "matched_float_and_ternary_qat_new_fits_plus_development_evaluation_save_reload"
	}
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
	f := flag.NewFlagSet("riido-statehint-claims-trit-study", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	parentName := f.String("parent", "", "explicit original float RSC parent")
	parentSHA := f.String("parent-sha256", "", "exact float parent SHA-256")
	input := f.String("input", "", "explicit canonical JSONL")
	sha := f.String("input-sha256", "", "exact input SHA-256")
	check := f.Bool("check", false, "validate pins and counts only")
	out := f.String("out", "", "fresh private study directory")
	profile := f.Bool("cpu-profile", false, "private CPU profile of matched float/QAT fits and evaluation")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err = fmt.Fprintln(errorOutput, "riido-statehint-claims-trit-study --parent FILE --parent-sha256 SHA --input FILE --input-sha256 SHA --check\nFixed PTQ, matched float and warm QAT: --out .cache/statehint-claims-trit-study/NEW-RUN [--cpu-profile]. Check reads pins/schema/counts only; no model construction or predictions. No recipe, calibration, test or application inputs.")
			return err
		}
		return errStudy
	}
	if f.NArg() != 0 || !local(*parentName) || !validSHA(*parentSHA) || !local(*input) || !validSHA(*sha) || *check && (*out != "" || *profile) || !*check && (!local(*out) || filepath.Dir(*out) != anchor) {
		return errStudy
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		return errStudy
	}
	defer root.Close()
	parentData, err := readPinned(root, *parentName, *parentSHA, parentBytes, parentBytes)
	if err != nil || !parentShape(parentData) {
		return errStudy
	}
	data, err := readPinned(root, *input, *sha, inputBudget, 0)
	if err != nil {
		return errStudy
	}
	p, err := prepare(data)
	if err != nil {
		return errStudy
	}
	if *check {
		return json.NewEncoder(output).Encode(struct {
			Status string `json:"status"`
			Counts counts `json:"counts"`
		}{"checked pins/schema/counts; no model construction, FromFloat, Fit, predictions or outputs", p.counts})
	}
	if p.counts.FitRows != 1680 || p.counts.FitFamilies != 840 || p.counts.FitLineages != 280 || p.counts.DevRows != devRows || p.counts.DevFamilies != 120 || p.counts.DevLineages != 40 {
		return errStudy
	}
	parent, err := statehintclaims.Load(bytes.NewReader(parentData))
	if err != nil || parent.TrainingSteps() == 0 {
		return errStudy
	}
	var canonical bytes.Buffer
	if parent.Save(&canonical) != nil || !bytes.Equal(parentData, canonical.Bytes()) {
		return errStudy
	}
	private, err := privateOutput(root, *out)
	if err != nil {
		return errStudy
	}
	defer private.Close()
	r, err := fitStudy(private, parent, parentData, p, *parentSHA, *sha, *profile)
	if err != nil {
		return errStudy
	}
	return json.NewEncoder(output).Encode(struct {
		Status    string `json:"status"`
		Qualified bool   `json:"semantic_quality_qualified"`
	}{r.Status, false})
}
func main() {
	runtime.GOMAXPROCS(2)
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "ternary claims study failed; check pinned canonical inputs and fresh private output")
		os.Exit(1)
	}
}
