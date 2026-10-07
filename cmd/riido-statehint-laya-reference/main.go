// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Aggregate pinned frozen-reference scores; never load, predict or train a model.
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
	"sort"
	"strings"

	"github.com/teamswyg/laya-tools/internal/statehintcorpus"
	"github.com/teamswyg/laya-tools/internal/statehintfamily"
	"github.com/teamswyg/laya-tools/pkg/statehint"
)

const (
	rubricSHA            = "b76a0e59bf046e82b0b868c56b9aba21a9d607c94bc42d5f56471a3947556215"
	baseSHA              = "891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"
	deltaSHA             = "e7037a4c92460dd0c77facb643a26a2f3ce99240e6a59bd78c15ad91df314fcc"
	baseTemperature      = 1.0000158548355103
	deltaTemperature     = .65
	instructionSHA       = "255725ec00a4526fefedfad729ddd75acc9681181b81fc0f8f96b89b0a400fd3"
	scoreBudget          = 2 << 20
	probabilityTolerance = 1e-6
	anchor               = ".cache/statehint-laya-reference-report"
)

var errReference = errors.New("reference aggregation input, checksum or output invalid")

type pin struct {
	Path string `json:"path"`
	SHA  string `json:"sha256"`
}
type scoreRow struct {
	ID                 string    `json:"id"`
	Locale             string    `json:"locale"`
	TextSHA            string    `json:"text_sha256"`
	BaseLogits         []float64 `json:"base_logits"`
	DeltaLogits        []float64 `json:"delta_logits"`
	BaseProbabilities  []float64 `json:"base_probabilities"`
	DeltaProbabilities []float64 `json:"delta_probabilities"`
	Truncated          bool      `json:"truncated"`
}
type scoreFile struct {
	Schema             string             `json:"schema"`
	Status             string             `json:"status"`
	ValidationSHA      string             `json:"validation_sha256"`
	IntentOrder        []statehint.Intent `json:"intent_order"`
	BaseSHA            string             `json:"base_sha256"`
	DeltaSHA           string             `json:"delta_sha256"`
	BaseTemperature    float64            `json:"base_temperature"`
	DeltaTemperature   float64            `json:"delta_temperature"`
	BaseOrigin         string             `json:"base_origin"`
	DeltaOrigin        string             `json:"delta_origin"`
	InstructionSHA     string             `json:"instruction_sha256"`
	FeatureCacheSHA    string             `json:"feature_cache_sha256"`
	TrainingStepsKnown *bool              `json:"training_steps_known"`
	TrainingSteps      json.RawMessage    `json:"training_steps"`
	Rows               []scoreRow         `json:"rows"`
}
type prediction struct {
	probabilities      [8]float64
	winner             int
	confidence, margin float64
}
type localeReport struct {
	Rows             int        `json:"rows"`
	RawCorrect       int        `json:"raw_eight_correct"`
	RawAccuracy      float64    `json:"raw_eight_accuracy"`
	EightNLL         float64    `json:"eight_nll"`
	Targets          [3]int     `json:"targets_pcq"`
	Proposed         [3]int     `json:"proposed_pcq"`
	Correct          [3]int     `json:"correct_pcq"`
	CorrectFamilies  [3]int     `json:"correct_families_pcq"`
	CorrectLineages  [3]int     `json:"correct_declared_lineages_pcq"`
	Precision        [3]float64 `json:"precision_pcq"`
	PrecisionDefined [3]bool    `json:"precision_defined_pcq"`
	Coverage         [3]float64 `json:"correct_coverage_pcq"`
}
type armReport struct {
	Origin                     string          `json:"origin"`
	ModelSHA                   string          `json:"model_sha256"`
	Temperature                float64         `json:"temperature"`
	TrainingStepsKnown         bool            `json:"training_steps_known"`
	TrainingSteps              *uint64         `json:"training_steps"`
	Eligible                   bool            `json:"eligible"`
	EligibilityAssessed        bool            `json:"eligibility_assessed"`
	Rows                       int             `json:"rows"`
	Families                   int             `json:"paired_families"`
	Lineages                   int             `json:"declared_lineages"`
	RawCorrect                 int             `json:"raw_eight_correct"`
	RawAccuracy                float64         `json:"raw_eight_accuracy"`
	EightNLL                   float64         `json:"eight_nll"`
	NLLClippedRows             int             `json:"eight_nll_clipped_rows"`
	RawConfusion               [8][8]int       `json:"raw_eight_confusion"`
	GatedConfusion             [4][4]int       `json:"gated_pcq_none_confusion"`
	GatedProposals             int             `json:"gated_proposals"`
	CorrectGated               int             `json:"correct_gated_proposals"`
	OutsideScope               int             `json:"outside_scope_rows"`
	BelowConfidence            int             `json:"below_confidence_rows"`
	BelowMargin                int             `json:"below_margin_rows"`
	CompletionCorrect          int             `json:"completion_correct"`
	CompletionProposals        int             `json:"completion_proposals"`
	CompletionPrecision        float64         `json:"completion_precision"`
	CompletionPrecisionDefined bool            `json:"completion_precision_defined"`
	WrongFamilies              [3]int          `json:"wrong_candidate_families_pcq"`
	WrongLineages              [3]int          `json:"wrong_candidate_declared_lineages_pcq"`
	Locales                    [2]localeReport `json:"locales_ko_en"`
}
type report struct {
	Schema                      string              `json:"schema"`
	Status                      string              `json:"status"`
	ReferenceOnly               bool                `json:"reference_only"`
	DevelopmentOnly             bool                `json:"development_only"`
	PreviouslyExposed           bool                `json:"validation_previously_exposed"`
	HumanTruth                  bool                `json:"human_truth"`
	FreshQualification          bool                `json:"fresh_qualification"`
	ConfidenceNotCalibratedOnV4 bool                `json:"confidence_not_calibrated_on_v4"`
	EligibilityAssessed         bool                `json:"eligibility_assessed"`
	Selected                    string              `json:"selected_arm"`
	SelectionWeight             int                 `json:"validation_selection_weight"`
	Promotion                   bool                `json:"promotion_performed"`
	DeploymentQualified         bool                `json:"deployment_qualified"`
	CalibrationCalls            int                 `json:"calibration_forwards"`
	TestCalls                   int                 `json:"test_forwards"`
	AdapterModelCalls           int                 `json:"adapter_model_calls"`
	Validation                  pin                 `json:"validation_input"`
	Scores                      pin                 `json:"scores_input"`
	IntentOrder                 [8]statehint.Intent `json:"probability_intent_order"`
	CandidateOrder              [3]statehint.Intent `json:"candidate_order"`
	Confidence                  float64             `json:"confidence_floor"`
	Margin                      float64             `json:"margin_floor"`
	TiePolicy                   string              `json:"raw_tie_policy"`
	InstructionSHA              string              `json:"instruction_sha256"`
	FeatureCacheSHA             string              `json:"feature_cache_sha256"`
	Arms                        [2]armReport        `json:"arms"`
}

func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func validHash(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == sha256.Size && hex.EncodeToString(b) == s
}
func local(s string) bool {
	return filepath.IsLocal(s) && filepath.Clean(s) == s && s != "." && !strings.Contains(s, "\\")
}
func readPin(root *os.Root, p pin, limit int64) ([]byte, error) {
	if !local(p.Path) || !validHash(p.SHA) {
		return nil, errReference
	}
	s, e := root.Lstat(p.Path)
	if e != nil || !s.Mode().IsRegular() || s.Size() < 1 || s.Size() > limit {
		return nil, errReference
	}
	f, e := root.Open(p.Path)
	if e != nil {
		return nil, errReference
	}
	defer f.Close()
	a, e := f.Stat()
	if e != nil || !a.Mode().IsRegular() || !os.SameFile(s, a) {
		return nil, errReference
	}
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	if e != nil || int64(len(b)) > limit || digest(b) != p.SHA {
		return nil, errReference
	}
	return b, nil
}
func decodeScores(b []byte) (scoreFile, error) {
	var s scoreFile
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(&s) != nil {
		return s, errReference
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return s, errReference
	}
	return s, nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Keep producer probabilities unchanged; check them against stable softmax logits.
func checkedPrediction(logits, probs []float64, t float64) (prediction, error) {
	var p prediction
	if len(logits) != 8 || len(probs) != 8 || !finite(t) || t <= 0 {
		return p, errReference
	}
	maxLogit := logits[0]
	for _, v := range logits {
		if !finite(v) {
			return p, errReference
		}
		maxLogit = math.Max(maxLogit, v)
	}
	var soft [8]float64
	total, sum := 0.0, 0.0
	for i, v := range logits {
		soft[i] = math.Exp((v - maxLogit) / t)
		total += soft[i]
	}
	if !finite(total) || total <= 0 {
		return p, errReference
	}
	for i, v := range probs {
		if !finite(v) || v < 0 || v > 1 || math.Abs(v-soft[i]/total) > probabilityTolerance {
			return p, errReference
		}
		p.probabilities[i] = v
		sum += v
	}
	if math.Abs(sum-1) > probabilityTolerance {
		return p, errReference
	}
	p.winner = 0
	for i, v := range p.probabilities {
		if v > p.probabilities[p.winner] {
			p.winner = i
		}
	}
	runner := -1
	for i, v := range p.probabilities {
		if i != p.winner && (runner < 0 || v > p.probabilities[runner]) {
			runner = i
		}
	}
	p.confidence = p.probabilities[p.winner]
	p.margin = p.confidence - p.probabilities[runner]
	return p, nil
}
func alignScores(c statehintcorpus.Corpus, s scoreFile, validationSHA string) ([2][]prediction, error) {
	var out [2][]prediction
	if s.Schema != "riido-statehint-laya-reference-scores-v1" || s.Status != "reference_only" || s.ValidationSHA != validationSHA || !validHash(validationSHA) || s.BaseSHA != baseSHA || s.DeltaSHA != deltaSHA || s.BaseTemperature != baseTemperature || s.DeltaTemperature != deltaTemperature || s.BaseOrigin != "base_pretrained" || s.DeltaOrigin != "published_v1_delta" || s.InstructionSHA != instructionSHA || !validHash(s.FeatureCacheSHA) || s.TrainingStepsKnown == nil || *s.TrainingStepsKnown || !bytes.Equal(bytes.TrimSpace(s.TrainingSteps), []byte("null")) || len(s.IntentOrder) != 8 {
		return out, errReference
	}
	for i, v := range statehint.Intents() {
		if s.IntentOrder[i] != v {
			return out, errReference
		}
	}
	rows := c.Rows()
	if len(rows) == 0 || len(rows) > 240 || len(s.Rows) != len(rows) {
		return out, errReference
	}
	indices := make([]int, len(s.Rows))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool { return s.Rows[indices[i]].ID < s.Rows[indices[j]].ID })
	for i := 1; i < len(indices); i++ {
		if s.Rows[indices[i-1]].ID == s.Rows[indices[i]].ID {
			return out, errReference
		}
	}
	out[0] = make([]prediction, len(rows))
	out[1] = make([]prediction, len(rows))
	for i, row := range rows {
		j := sort.Search(len(indices), func(j int) bool { return s.Rows[indices[j]].ID >= row.ID })
		if j == len(indices) {
			return out, errReference
		}
		r := s.Rows[indices[j]]
		if r.ID != row.ID || r.Locale != row.Locale || r.TextSHA != digest([]byte(row.Text)) || r.Truncated {
			return out, errReference
		}
		var e error
		out[0][i], e = checkedPrediction(r.BaseLogits, r.BaseProbabilities, s.BaseTemperature)
		if e != nil {
			return out, e
		}
		out[1][i], e = checkedPrediction(r.DeltaLogits, r.DeltaProbabilities, s.DeltaTemperature)
		if e != nil {
			return out, e
		}
	}
	return out, nil
}
func display(i int) int {
	switch statehint.Intents()[i] {
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

type support struct {
	locale, class int
	group         string
}
type wrong struct {
	class int
	group string
}

func summarize(c statehintcorpus.Corpus, p []prediction, origin, modelSHA string, t float64) (armReport, error) {
	rows, pairs := c.Rows(), c.Pairs()
	r := armReport{Origin: origin, ModelSHA: modelSHA, Temperature: t, Rows: len(rows), Families: len(pairs)}
	if len(p) != len(rows) || len(rows) != 2*len(pairs) || len(rows) == 0 {
		return r, errReference
	}
	var supports []support
	var wrongGroups []wrong
	var groups []string
	for _, pair := range pairs {
		groups = append(groups, pair.Family.Lineage)
		var familyWrong [3]bool
		for locale, index := range pair.Rows {
			row, pred := rows[index], p[index]
			target, _ := statehint.IntentIndex(row.Expected)
			td, pd := display(target), display(pred.winner)
			gate := 3
			l := &r.Locales[locale]
			l.Rows++
			r.RawConfusion[target][pred.winner]++
			if target == pred.winner {
				r.RawCorrect++
				l.RawCorrect++
			}
			value := pred.probabilities[target]
			if value < statehintfamily.NLLFloor {
				r.NLLClippedRows++
			}
			loss := -math.Log(math.Max(value, statehintfamily.NLLFloor))
			r.EightNLL += loss
			l.EightNLL += loss
			if td < 3 {
				l.Targets[td]++
			}
			switch {
			case pd == 3:
				r.OutsideScope++
			case pred.confidence < statehintfamily.ConfidenceFloor:
				r.BelowConfidence++
			case pred.margin < statehintfamily.MarginFloor:
				r.BelowMargin++
			default:
				gate = pd
			}
			r.GatedConfusion[td][gate]++
			if gate < 3 {
				r.GatedProposals++
				l.Proposed[gate]++
				if gate == td {
					r.CorrectGated++
					l.Correct[gate]++
					l.CorrectFamilies[gate]++
					supports = append(supports, support{locale, gate, pair.Family.Lineage})
				} else {
					familyWrong[gate] = true
				}
			}
		}
		for class, bad := range familyWrong {
			if bad {
				r.WrongFamilies[class]++
				wrongGroups = append(wrongGroups, wrong{class, pair.Family.Lineage})
			}
		}
	}
	sort.Strings(groups)
	for i, g := range groups {
		if i == 0 || g != groups[i-1] {
			r.Lineages++
		}
	}
	sort.Slice(supports, func(i, j int) bool {
		a, b := supports[i], supports[j]
		if a.locale != b.locale {
			return a.locale < b.locale
		}
		if a.class != b.class {
			return a.class < b.class
		}
		return a.group < b.group
	})
	for i, v := range supports {
		if i == 0 || v != supports[i-1] {
			r.Locales[v.locale].CorrectLineages[v.class]++
		}
	}
	sort.Slice(wrongGroups, func(i, j int) bool {
		if wrongGroups[i].class != wrongGroups[j].class {
			return wrongGroups[i].class < wrongGroups[j].class
		}
		return wrongGroups[i].group < wrongGroups[j].group
	})
	for i, v := range wrongGroups {
		if i == 0 || v != wrongGroups[i-1] {
			r.WrongLineages[v.class]++
		}
	}
	r.RawAccuracy = float64(r.RawCorrect) / float64(r.Rows)
	r.EightNLL /= float64(r.Rows)
	for i := range r.Locales {
		l := &r.Locales[i]
		l.RawAccuracy = float64(l.RawCorrect) / float64(l.Rows)
		l.EightNLL /= float64(l.Rows)
		for class := 0; class < 3; class++ {
			if l.Proposed[class] > 0 {
				l.PrecisionDefined[class] = true
				l.Precision[class] = float64(l.Correct[class]) / float64(l.Proposed[class])
			}
			if l.Targets[class] > 0 {
				l.Coverage[class] = float64(l.Correct[class]) / float64(l.Targets[class])
			}
		}
		r.CompletionCorrect += l.Correct[1]
		r.CompletionProposals += l.Proposed[1]
	}
	if r.CompletionProposals > 0 {
		r.CompletionPrecisionDefined = true
		r.CompletionPrecision = float64(r.CompletionCorrect) / float64(r.CompletionProposals)
	}
	return r, nil
}
func aggregate(c statehintcorpus.Corpus, s scoreFile, validation, scorePin pin) (report, error) {
	r := report{Schema: "riido-statehint-laya-reference-aggregate-v1", Status: "reference_only", ReferenceOnly: true, DevelopmentOnly: true, PreviouslyExposed: true, ConfidenceNotCalibratedOnV4: true, Validation: validation, Scores: scorePin, IntentOrder: statehint.Intents(), CandidateOrder: [3]statehint.Intent{statehint.Progress, statehint.CompletionReport, statehint.Question}, Confidence: statehintfamily.ConfidenceFloor, Margin: statehintfamily.MarginFloor, TiePolicy: "first in frozen intent order; zero-margin ties cannot gate", InstructionSHA: s.InstructionSHA, FeatureCacheSHA: s.FeatureCacheSHA}
	p, e := alignScores(c, s, validation.SHA)
	if e != nil {
		return r, e
	}
	r.Arms[0], e = summarize(c, p[0], s.BaseOrigin, s.BaseSHA, s.BaseTemperature)
	if e != nil {
		return r, e
	}
	r.Arms[1], e = summarize(c, p[1], s.DeltaOrigin, s.DeltaSHA, s.DeltaTemperature)
	return r, e
}
func saveReport(root *os.Root, name string, r report) error {
	if !local(name) || filepath.Dir(name) != anchor {
		return errReference
	}
	for _, dir := range []string{".cache", anchor} {
		if e := root.Mkdir(dir, 0700); e != nil && !errors.Is(e, os.ErrExist) {
			return errReference
		}
		st, e := root.Lstat(dir)
		if e != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || dir == anchor && st.Mode().Perm() != 0700 {
			return errReference
		}
	}
	if root.Mkdir(name, 0700) != nil {
		return errReference
	}
	out, e := root.OpenRoot(name)
	if e != nil {
		return errReference
	}
	defer out.Close()
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return errReference
	}
	f, e := out.OpenFile("report.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return errReference
	}
	b = append(b, '\n')
	n, e := f.Write(b)
	if e == nil && n != len(b) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e != nil || ce != nil {
		return errReference
	}
	return nil
}
func run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riido-statehint-laya-reference", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	val := f.String("validation", "", "explicit exposed validation120 JSONL")
	valHash := f.String("validation-sha256", "", "validation SHA-256")
	scores := f.String("scores", "", "body-free frozen reference scores")
	scoreHash := f.String("scores-sha256", "", "scores SHA-256")
	dest := f.String("out", "", "new private aggregate run directory")
	check := f.Bool("check", false, "validate inputs without outputs")
	if e := f.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			_, e = fmt.Fprintln(errOut, "riido-statehint-laya-reference --validation FILE --validation-sha256 SHA --scores FILE --scores-sha256 SHA --check\nReplace --check with --out .cache/statehint-laya-reference-report/NEW-RUN to save descriptive aggregates. No model, Fit, native runtime, calibration or final-test input.")
			return e
		}
		return errReference
	}
	if f.NArg() != 0 || *check && *dest != "" || !*check && (!local(*dest) || filepath.Dir(*dest) != anchor) || *val == *scores {
		return errReference
	}
	root, e := os.OpenRoot(".")
	if e != nil {
		return errReference
	}
	defer root.Close()
	vp, sp := pin{*val, *valHash}, pin{*scores, *scoreHash}
	vb, e := readPin(root, vp, statehintcorpus.MaxFileBytes)
	if e != nil {
		return e
	}
	c, summary, e := statehintcorpus.Read(bytes.NewReader(vb), statehintcorpus.Options{Partition: "validation", RubricSHA: rubricSHA})
	if e != nil || summary.Families != 120 || summary.Rows != 240 {
		return errReference
	}
	for _, n := range summary.IntentFamilies {
		if n != 15 {
			return errReference
		}
	}
	sb, e := readPin(root, sp, scoreBudget)
	if e != nil {
		return e
	}
	s, e := decodeScores(sb)
	if e != nil {
		return e
	}
	r, e := aggregate(c, s, vp, sp)
	if e != nil {
		return e
	}
	if *check {
		return json.NewEncoder(out).Encode(struct {
			Status     string `json:"status"`
			Rows       int    `json:"rows"`
			ModelCalls int    `json:"model_calls"`
		}{"checked; no native/model calls or output files", 240, 0})
	}
	if e := saveReport(root, *dest, r); e != nil {
		return e
	}
	return json.NewEncoder(out).Encode(r)
}
func main() {
	if run(os.Args[1:], os.Stdout, os.Stderr) != nil {
		fmt.Fprintln(os.Stderr, "reference aggregation failed; check explicit pinned validation, scores and fresh private output")
		os.Exit(1)
	}
}
