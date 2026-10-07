// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
)

func number(i int) *int { return &i }
func encodeRows(t *testing.T, rows []rawRow) []byte {
	t.Helper()
	var b bytes.Buffer
	e := json.NewEncoder(&b)
	for _, r := range rows {
		if e.Encode(r) != nil {
			t.Fatal("encode owned fixture")
		}
	}
	return b.Bytes()
}
func encode(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func writeOwned(t *testing.T, name string, data []byte) {
	t.Helper()
	if os.WriteFile(name, data, 0600) != nil {
		t.Fatal("write owned fixture")
	}
}
func refreshChecksum(data []byte) {
	end := len(data) - sha256.Size
	s := sha256.Sum256(data[:end])
	copy(data[end:], s[:])
}
func ownedModel(t *testing.T, steps uint64) []byte {
	t.Helper()
	var b bytes.Buffer
	if statehintclaims.NewModel().Save(&b) != nil {
		t.Fatal("owned model")
	}
	data := append([]byte(nil), b.Bytes()...)
	binary.LittleEndian.PutUint64(data[16:24], steps)
	refreshChecksum(data)
	return data
}

// The test generates every word, ID and label itself. Repetition exercises
// owned numeric plumbing and has no semantic or model-quality meaning.
func ownedRows(split string, families int) []rawRow {
	rows := make([]rawRow, 0, 2*families)
	for i := 0; i < families; i++ {
		family := fmt.Sprintf("owned-%s-%03d", split, i)
		group := fmt.Sprintf("owned-%s-group-%03d", split, i/3)
		for _, locale := range []string{"ko", "en"} {
			text := "직접 작성한 숫자시험 문장입니다."
			if locale == "en" {
				text = "Amber kite original owned numerical fixture."
			}
			r := rawRow{Schema: rowSchema, ID: family + "-" + locale, Family: family, Group: group, Split: split, Locale: locale, Text: text, RubricSHA: strings.Repeat("b", 64), AnnotationSource: "ai_semantic_reference"}
			for h := 0; h < 3; h++ {
				target := statehintclaims.States()[(i+h)%3]
				r.Targets = append(r.Targets, string(target))
				end := len(text)
				if target == statehintclaims.False {
					end = 0
				}
				r.Evidence = append(r.Evidence, evidenceInput{number(0), number(end), "Owned numerical fixture; no semantic quality evidence."})
			}
			rows = append(rows, r)
		}
	}
	return rows
}

type fixture struct {
	parent, comparator, training, evaluation, metadata, receipt []byte
	refs                                                        receipt
	manifest                                                    manifest
}

func ownedFixture(t *testing.T) fixture {
	t.Helper()
	train := ownedRows("fit", 840)
	evaluation := ownedRows("dev", 180)
	meta := strataInput{Schema: strataSchema}
	for i, r := range evaluation {
		stratum := "contrast"
		if i >= 240 {
			stratum = "general"
		}
		meta.Rows = append(meta.Rows, stratumRow{r.ID, r.Family, r.Group, r.Locale, stratum})
	}
	f := fixture{parent: ownedModel(t, 2120), comparator: ownedModel(t, 4240), training: encodeRows(t, train), evaluation: encodeRows(t, evaluation), metadata: encode(t, meta)}
	c, e := metadataCounts(meta.Rows, true)
	if e != nil {
		t.Fatal(e)
	}
	f.refs = receipt{Schema: receiptSchema, TrainingSHA: digest(f.training), EvaluationSHA: digest(f.evaluation), EvaluationBytes: int64(len(f.evaluation)), StrataSHA: digest(f.metadata), RowSchema: rowSchema, AnnotationSource: "ai_semantic_reference", ReferenceStatus: "ai_authored_independently_ai_reviewed_human_pending", Locales: [2]string{"ko", "en"}, Heads: headOrder(), States: statehintclaims.States(), EvaluationCounts: c}
	for l := 0; l < 2; l++ {
		for h := 0; h < 3; h++ {
			f.refs.NewTrainingUnknownLineages[l][h] = 40
			for k := 0; k < 3; k++ {
				f.refs.ContrastFamilies[l][h][k] = 40
				f.refs.ContrastLineages[l][h][k] = 40
				f.refs.NewTrainingFamilies[l][h][k] = 40
			}
		}
	}
	f.receipt = encode(t, f.refs)
	f.manifest = manifest{digest(f.parent), digest(f.comparator)}
	return f
}
func (f fixture) write(t *testing.T) {
	t.Helper()
	for _, file := range []struct {
		name string
		data []byte
	}{{"parent.rsc", f.parent}, {"comparator.rsc", f.comparator}, {"training.jsonl", f.training}, {"evaluation.jsonl", f.evaluation}, {"strata.json", f.metadata}, {"receipt.json", f.receipt}} {
		writeOwned(t, file.name, file.data)
	}
}
func (f fixture) args() []string {
	return []string{"--parent", "parent.rsc", "--parent-sha256", digest(f.parent), "--comparator", "comparator.rsc", "--comparator-sha256", digest(f.comparator), "--training", "training.jsonl", "--training-sha256", digest(f.training), "--evaluation", "evaluation.jsonl", "--evaluation-sha256", f.refs.EvaluationSHA, "--strata", "strata.json", "--strata-sha256", digest(f.metadata), "--receipt", "receipt.json", "--receipt-sha256", digest(f.receipt)}
}
func runOwned(t *testing.T, f fixture, extra ...string) ([]byte, error) {
	t.Helper()
	var output, errorOutput bytes.Buffer
	e := runWithManifest(append(f.args(), extra...), &output, &errorOutput, f.manifest)
	if strings.Contains(errorOutput.String(), "parent.rsc") {
		t.Fatal("path in diagnostics")
	}
	return output.Bytes(), e
}

func TestCheckDoesNotReadEvaluationOrInitializeNumericModels(t *testing.T) {
	t.Chdir(t.TempDir())
	f := ownedFixture(t)
	// Its existing body is malformed and does not match the declared hash.
	// Its size matches the sealed receipt. Check must neither parse nor hash it.
	f.evaluation = bytes.Repeat([]byte{'x'}, len(f.evaluation))
	binary.LittleEndian.PutUint32(f.parent[192:196], math.Float32bits(float32(math.NaN())))
	refreshChecksum(f.parent)
	binary.LittleEndian.PutUint32(f.comparator[192:196], math.Float32bits(float32(math.NaN())))
	refreshChecksum(f.comparator)
	f.manifest = manifest{digest(f.parent), digest(f.comparator)}
	f.write(t)
	if os.Geteuid() != 0 {
		if os.Chmod("evaluation.jsonl", 0000) != nil {
			t.Fatal("owned evaluation mode")
		}
		if file, e := os.Open("evaluation.jsonl"); e == nil {
			file.Close()
			t.Fatal("unreadable synthetic leaf unexpectedly readable")
		}
	}
	output, e := runOwned(t, f, "--check")
	if e != nil {
		t.Fatal("check accessed body/numeric model", e)
	}
	if !bytes.Contains(output, []byte("no Load, Predict, Fit, evaluation body access or outputs")) {
		t.Fatal("unclear check scope")
	}
	if _, e := os.Stat(".cache"); !os.IsNotExist(e) {
		t.Fatal("check created outputs")
	}
	for _, extra := range [][]string{{"--check", "--cpu-profile"}, {"--check", "--out", anchor + "/bad"}, {"--epochs", "1"}, {"--seed", "1"}, {"--temperature", "2"}, {"--calibration", "private-marker"}, {"--model", "private-marker"}, {"--selection"}, {"--check", "--manifest", "private-marker"}} {
		if output, e := runOwned(t, f, extra...); e == nil || len(output) != 0 || strings.Contains(e.Error(), "private-marker") {
			t.Fatal("invalid flag accepted/private path exposed")
		}
	}
	var outputBuffer, errorBuffer bytes.Buffer
	if run(append(f.args(), "--check"), &outputBuffer, &errorBuffer) == nil {
		t.Fatal("CLI accepted an alternate synthetic manifest")
	}
	// The same numeric-invalid parent fails actual mode before output/Fit.
	if _, e := runOwned(t, f, "--out", anchor+"/invalid-numeric"); e == nil {
		t.Fatal("numeric parent accepted")
	}
	if _, e := os.Stat(".cache"); !os.IsNotExist(e) {
		t.Fatal("numeric parent failure created output")
	}
	if os.Remove("evaluation.jsonl") != nil {
		t.Fatal("remove owned file")
	}
	if _, e := runOwned(t, f, "--check"); e == nil {
		t.Fatal("missing evaluation accepted")
	}
}

func TestDigestLengthGuardHasNoOversizedDecodeAllocation(t *testing.T) {
	oversized := strings.Repeat("a", 1<<20)
	if validSHA(oversized) || validSHA(strings.Repeat("A", 64)) || validSHA(strings.Repeat("g", 64)) || !validSHA(strings.Repeat("a", 64)) {
		t.Fatal("digest admission changed")
	}
	if allocations := testing.AllocsPerRun(10, func() {
		if validSHA(oversized) {
			panic("oversized digest")
		}
	}); allocations != 0 {
		t.Fatal("oversized digest decoded before bound", allocations)
	}
}

func TestPreflightRejectsEvaluationPathAndHardlinkAliasesBeforeAnyRead(t *testing.T) {
	t.Chdir(t.TempDir())
	f := ownedFixture(t)
	// Model-sized malformed bytes pass every role's file budget. An unreadable
	// leaf and deliberately inconsistent pins make a body-read failure distinct
	// from the alias guard's sentinel error; neither check nor actual may read.
	f.evaluation = bytes.Repeat([]byte{'x'}, parentBytes)
	f.write(t)
	if os.Chmod("evaluation.jsonl", 0000) != nil {
		t.Fatal("owned evaluation mode")
	}
	if os.Link("evaluation.jsonl", "hardlink-evaluation.jsonl") != nil {
		t.Fatal("owned evaluation hardlink")
	}
	for _, role := range []string{"--parent", "--training", "--strata", "--receipt", "--comparator"} {
		for _, alias := range []string{"evaluation.jsonl", "hardlink-evaluation.jsonl"} {
			for _, mode := range []string{"check", "actual"} {
				args := f.args()
				for i := 0; i < len(args); i += 2 {
					if args[i] == role {
						args[i+1] = alias
					}
				}
				if mode == "check" {
					args = append(args, "--check")
				} else {
					args = append(args, "--out", anchor+"/alias-rejected")
				}
				var output, errorOutput bytes.Buffer
				if e := runWithManifest(args, &output, &errorOutput, f.manifest); !errors.Is(e, errEvaluationAlias) || output.Len() != 0 || errorOutput.Len() != 0 {
					t.Fatalf("%s %s %s did not reject before reading: %v", role, alias, mode, e)
				}
				if _, e := os.Stat(".cache"); !os.IsNotExist(e) {
					t.Fatal("alias admission created output or consumed fit")
				}
			}
		}
	}
}

func TestStrictRowsGroupingAndUTF8Evidence(t *testing.T) {
	good := ownedRows("fit", 3)
	if p, e := prepare(encodeRows(t, good), "fit"); e != nil || p.counts != (counts{Rows: 6, Families: 3, Lineages: 1, Locales: [2]int{3, 3}}) {
		t.Fatal("owned grouping", e)
	}
	mutations := []func([]rawRow){
		func(r []rawRow) { r[0].Targets[0] = "True" }, func(r []rawRow) { r[0].Targets = r[0].Targets[:2] }, func(r []rawRow) { r[0].Evidence = r[0].Evidence[:2] }, func(r []rawRow) { r[0].Evidence[0].Start = nil }, func(r []rawRow) { r[0].Evidence[0].Start = number(1) }, func(r []rawRow) { r[0].Evidence[0].End = number(-1) }, func(r []rawRow) { r[0].Evidence[0].End = number(len(r[0].Text) + 1) }, func(r []rawRow) { r[0].Evidence[0].End = number(0) }, func(r []rawRow) { r[0].Evidence[0].Reason = "" }, func(r []rawRow) { r[0].Evidence[1].Start = number(3); r[0].Evidence[1].End = number(3) }, func(r []rawRow) { r[0].Text = " \t" }, func(r []rawRow) { r[0].Text = "!!!" }, func(r []rawRow) { r[0].Text = string([]byte{0xff}) }, func(r []rawRow) { r[0].Schema = "alternate" }, func(r []rawRow) { r[0].AnnotationSource = "human_verified" }, func(r []rawRow) { r[0].Split = "dev" }, func(r []rawRow) { r[0].RubricSHA = strings.Repeat("c", 64) }, func(r []rawRow) { r[0].ID = r[1].ID }, func(r []rawRow) { r[0].Group = "other-group" }, func(r []rawRow) { r[0].Locale = "fr" }, func(r []rawRow) { r[0].Family = "case/private" },
	}
	for i, mutate := range mutations {
		rows := ownedRows("fit", 3)
		mutate(rows)
		if _, e := prepare(encodeRows(t, rows), "fit"); e == nil {
			t.Fatalf("accepted malformed case %d", i)
		}
	}
	data := encodeRows(t, good)
	for _, bad := range [][]byte{bytes.Replace(data, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1), bytes.Replace(data, []byte(`"schema":`), []byte(`"Schema":`), 1), bytes.Replace(data, []byte(`"schema":`), []byte(`"extra":0,"schema":`), 1), append(data, '\n'), append([]byte{0xff}, data...), append([]byte("null\n"), data...)} {
		if _, e := prepare(bad, "fit"); e == nil {
			t.Fatal("accepted noncanonical JSONL")
		}
	}
	if _, e := prepare(encodeRows(t, good[:5]), "fit"); e == nil {
		t.Fatal("unpaired family accepted")
	}
	if _, e := prepare(encodeRows(t, ownedRows("fit", 4)), "fit"); e == nil {
		t.Fatal("unequal lineage size accepted")
	}
	// Paired locale membership does not imply identical wording or references.
	different := ownedRows("fit", 3)
	different[1].Targets[0] = "unknown"
	different[1].Evidence[0].End = number(len(different[1].Text))
	paired, e := prepare(encodeRows(t, different), "fit")
	if e != nil || paired.rows[0].labels[0] == paired.rows[1].labels[0] {
		t.Fatal("independent locale reference differences rejected", e)
	}
}

func TestStrataReceiptAndIndependentSplits(t *testing.T) {
	f := ownedFixture(t)
	train, e := prepare(f.training, "fit")
	if e != nil {
		t.Fatal(e)
	}
	s, e := prepareStrata(f.metadata, train)
	if e != nil {
		t.Fatal(e)
	}
	r, e := validateReceipt(f.receipt, digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s)
	if e != nil {
		t.Fatal(e)
	}
	p, e := prepare(f.evaluation, "dev")
	if e != nil || validateEvaluation(p, train, s, r) != nil {
		t.Fatal("valid owned evaluation rejected", e)
	}
	var meta strataInput
	if json.Unmarshal(f.metadata, &meta) != nil {
		t.Fatal("owned metadata")
	}
	meta.Rows[0] = stratumRow{train.rows[0].ID, train.rows[0].Family, train.rows[0].Group, train.rows[0].Locale, "contrast"}
	if _, e := prepareStrata(encode(t, meta), train); e == nil {
		t.Fatal("training identity leaked across split")
	}
	if json.Unmarshal(f.metadata, &meta) != nil {
		t.Fatal("owned metadata")
	}
	meta.Rows[0].Stratum = "general"
	if _, e := prepareStrata(encode(t, meta), train); e == nil {
		t.Fatal("stratum split within lineage accepted")
	}
	for _, change := range []func(*receipt){func(r *receipt) { r.ContrastFamilies[0][0][2] = 19 }, func(r *receipt) { r.ContrastLineages[1][2][0] = 9 }, func(r *receipt) { r.NewTrainingFamilies[0][0][0] = 29 }, func(r *receipt) { r.NewTrainingUnknownLineages[1][0] = 14 }, func(r *receipt) { r.States[0] = statehintclaims.Unknown }, func(r *receipt) { r.Heads[0] = "completion_claimed" }, func(r *receipt) { r.EvaluationSHA = strings.Repeat("0", 64) }, func(r *receipt) { r.EvaluationBytes = 0 }, func(r *receipt) { r.ReferenceStatus = "human_verified" }} {
		r := f.refs
		change(&r)
		if _, e := validateReceipt(encode(t, r), digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s); e == nil {
			t.Fatal("invalid receipt accepted")
		}
	}
	badArrays := bytes.Replace(f.receipt, []byte(`"locale_order":["ko","en"]`), []byte(`"locale_order":["ko","en","ko"]`), 1)
	if _, e := validateReceipt(badArrays, digest(f.training), f.refs.EvaluationSHA, digest(f.metadata), s); e == nil {
		t.Fatal("extra canonical-array entries accepted")
	}
	p.rows[0].labels[0] = statehintclaims.Unknown
	if validateEvaluation(p, train, s, r) == nil {
		t.Fatal("evaluation support differs from writer receipt")
	}
}

func TestLineageBootstrapUsesFixedWeightAndPairedLocales(t *testing.T) {
	values := [2][]float64{{-2, -2}, {1, 1}}
	r, e := bootstrap(values)
	if e != nil || r.Difference != -1 || r.ConfidenceInterval != [2]float64{-1, -1} || r.Weights != [2]int{2, 1} || r.Resamples != 10000 || r.Seed != 1729 || r.PercentileConvention != "linear_interpolation_at_q_times_n_minus_one_q_0.025_0.975" {
		t.Fatal("fixed bootstrap convention/weight changed", r, e)
	}
	varying := [2][]float64{{-3, -1}, {1, 4}}
	a, e := bootstrap(varying)
	b, e2 := bootstrap(varying)
	if e != nil || e2 != nil || a != b || a.Difference != -0.5 || a.ConfidenceInterval != [2]float64{-5.0 / 3, 2.0 / 3} {
		t.Fatal("deterministic known group distribution changed", a, e, e2)
	}
	f := ownedFixture(t)
	train, _ := prepare(f.training, "fit")
	s, _ := prepareStrata(f.metadata, train)
	p, _ := prepare(f.evaluation, "dev")
	candidate := make([]observation, len(p.rows))
	comparator := make([]observation, len(p.rows))
	for i, r := range p.rows {
		delta := -2.0
		if s.rows[s.byID[r.ID]].Stratum == "general" {
			delta = 1
		}
		if r.Locale == "ko" {
			delta += .75
		} else {
			delta -= .75
		}
		for h, target := range r.labels {
			k, _ := statehintclaims.StateIndex(target)
			candidate[i].score.Probabilities[h][k] = math.Exp(-(4 + delta))
			comparator[i].score.Probabilities[h][k] = math.Exp(-4)
		}
	}
	grouped, e := paired(p, s, candidate, comparator)
	if e != nil || math.Abs(grouped.Difference+1) > 1e-12 || math.Abs(grouped.ConfidenceInterval[0]+1) > 1e-12 || math.Abs(grouped.ConfidenceInterval[1]+1) > 1e-12 || grouped.Lineages != [2]int{40, 20} {
		t.Fatal("cross-language rows were resampled separately", grouped, e)
	}
	if _, e := paired(prepared{rows: p.rows[:len(p.rows)-1]}, s, candidate[:len(candidate)-1], comparator[:len(comparator)-1]); e == nil {
		t.Fatal("incomplete paired group accepted")
	}
	if _, e := bootstrap([2][]float64{{math.NaN()}, {0}}); e == nil {
		t.Fatal("nonfinite bootstrap accepted")
	}
}

func TestPercentileConventionInterpolatesBetweenSortedDraws(t *testing.T) {
	sorted := []float64{0, 10, 20, 30}
	if got := interpolatedPercentile(sorted, .025); math.Abs(got-.75) > 1e-12 {
		t.Fatal("lower percentile rounded to an array entry", got)
	}
	if got := interpolatedPercentile(sorted, .975); math.Abs(got-29.25) > 1e-12 {
		t.Fatal("upper percentile rounded to an array entry", got)
	}
	if interpolatedPercentile(sorted, 0) != 0 || interpolatedPercentile(sorted, 1) != 30 || interpolatedPercentile([]float64{7}, .025) != 7 {
		t.Fatal("percentile endpoint/singleton changed")
	}
}

func TestUnknownCEAndFalsePositiveProgressConditions(t *testing.T) {
	candidate, comparator := evaluationReport{}, evaluationReport{}
	for l := 0; l < 2; l++ {
		candidate.Locales[l].UnknownTargetCount = 1
		candidate.Locales[l].UnknownTargetCE = 1
		comparator.Locales[l].UnknownTargetCount = 1
		comparator.Locales[l].UnknownTargetCE = 2
	}
	delta := pairedReport{ConfidenceInterval: [2]float64{-1, -.1}}
	if !progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("all conditions should pass")
	}
	candidate.Locales[1].TrueFalsePositive[2] = 1
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("one locale/head FP increase ignored")
	}
	candidate.Locales[1].TrueFalsePositive[2] = 0
	candidate.Locales[0].UnknownTargetCE = 2
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("equal unknown CE counted as improvement")
	}
	candidate.Locales[0].UnknownTargetCount = 0
	if progress(delta, candidate, comparator).ResearchProgress {
		t.Fatal("undefined unknown CE counted as improvement")
	}
}

func TestAggregateFalsePositiveIncludesFalseAndUnknownReferences(t *testing.T) {
	p, e := prepare(encodeRows(t, ownedRows("dev", 3)), "dev")
	if e != nil {
		t.Fatal(e)
	}
	obs := make([]observation, len(p.rows))
	for i := range obs {
		for h := 0; h < 3; h++ {
			prob := [3]float64{.95, .025, .025}
			obs[i].score.Probabilities[h] = prob
			obs[i].prediction.Heads[h] = statehintclaims.HeadPrediction{Head: statehintclaims.Head(h).String(), Winner: statehintclaims.True, State: statehintclaims.True, Probabilities: prob}
		}
	}
	r, e := evaluate(p, obs)
	if e != nil {
		t.Fatal(e)
	}
	for _, l := range r.Locales {
		if l.TrueProposed != [3]int{3, 3, 3} || l.TrueCorrect != [3]int{1, 1, 1} || l.TrueFalsePositive != [3]int{2, 2, 2} || l.TrueOnFalse != [3]int{1, 1, 1} || l.TrueOnUnknown != [3]int{1, 1, 1} || l.UnknownTargetCount != 3 {
			t.Fatal("false/unknown targets excluded from gated true FP", l)
		}
	}
}

func TestPinnedInputsAndEvaluationSnapshotRejectSymlinksAndDrift(t *testing.T) {
	t.Chdir(t.TempDir())
	f := ownedFixture(t)
	f.write(t)
	root, e := os.OpenRoot(".")
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	if os.Symlink("evaluation.jsonl", "linked-evaluation.jsonl") != nil {
		t.Fatal("owned symlink")
	}
	if _, e := inspectInput(root, "linked-evaluation.jsonl", inputBudget); e == nil {
		t.Fatal("leaf symlink accepted")
	}
	if os.Mkdir("regular", 0700) != nil {
		t.Fatal("owned dir")
	}
	writeOwned(t, "regular/model.rsc", f.parent)
	if os.Symlink("regular", "linked") != nil {
		t.Fatal("owned symlink")
	}
	if _, e := readPinned(root, "linked/model.rsc", digest(f.parent), parentBytes, parentBytes); e == nil {
		t.Fatal("ancestor symlink accepted")
	}
	if _, e := inspectInput(root, "../evaluation.jsonl", inputBudget); e == nil {
		t.Fatal("escape accepted")
	}
	before, e := inspectInput(root, "evaluation.jsonl", inputBudget)
	if e != nil {
		t.Fatal(e)
	}
	p := pins{evaluationName: "evaluation.jsonl", evaluationSHA: digest(f.evaluation)}
	if _, e := readEvaluation(root, p, before); e != nil {
		t.Fatal(e)
	}
	// Same bytes and inode with a changed timestamp still invalidate the seal.
	changed := before.ModTime().Add(time.Second)
	if os.Chtimes("evaluation.jsonl", changed, changed) != nil {
		t.Fatal("owned time")
	}
	if _, e := readEvaluation(root, p, before); e == nil {
		t.Fatal("evaluation snapshot drift accepted")
	}
	if os.Symlink("regular", ".cache") != nil {
		t.Fatal("owned cache symlink")
	}
	if out, e := privateOutput(root, anchor+"/bad"); e == nil {
		out.Close()
		t.Fatal("symlink output accepted")
	}
}

func TestOwnedFixedRunParityAndFitBeforeAnyEvaluationValidation(t *testing.T) {
	t.Chdir(t.TempDir())
	f := ownedFixture(t)
	f.write(t)
	if _, e := runOwned(t, f, "--out", anchor+"/owned-a", "--cpu-profile"); e != nil {
		t.Fatal("owned synthetic run", e)
	}
	data, e := os.ReadFile(anchor + "/owned-a/report.json")
	var report studyReport
	if e != nil || json.Unmarshal(data, &report) != nil {
		t.Fatal("owned report", e)
	}
	if report.FitCalls != 1 || report.Fit.BaseTrainingSteps != 2120 || report.Fit.NewOptimizerSteps != 2120 || report.Fit.TrainingSteps != 4240 || report.Fit.Samples != 1680 || report.Fit.Epochs != 40 || report.Fit.Batches != 2120 || report.Recipe != fixedRecipe() || report.Qualified || report.QualificationChanged || report.Selection || report.Calibration || report.CalTestAccessed || report.StateWrites != 0 || !report.FitBeforeEvaluation || report.OSRSSMeasured || !report.CPUProfile || report.FitLossMeaning != "mean_online_fit_trajectory_loss_across_all_epochs_not_final_evaluation_ce" || report.Status != "ai_reference_research_only_not_human_or_product_evidence" {
		t.Fatal("fixed research boundary/counters changed", report.Fit)
	}
	if !report.Candidate.ReloadByteExact || !report.Candidate.ReloadPredictExact || !report.Candidate.ReloadScoresExact || report.Candidate.ArtifactBytes != 73988 || report.Candidate.PredictionCalls != 360 || report.Candidate.ScoreCalls != 360 || report.Candidate.ReloadPredictionCalls != 360 || report.Candidate.ReloadScoreCalls != 360 || report.Primary.Lineages != [2]int{40, 20} || !report.Parent.ReadOnlyExact || !report.Comparator.ReadOnlyExact {
		t.Fatal("candidate provenance/parity changed")
	}
	for _, evaluation := range []evaluationReport{report.CandidateEvaluation, report.ComparatorEvaluation} {
		if evaluation.Rows != 360 {
			t.Fatal("evaluation denominator")
		}
		for _, l := range evaluation.Locales {
			if l.Rows != 180 || l.UnknownTargetCount != 180 {
				t.Fatal("locale/unknown denominator")
			}
			for h := 0; h < 3; h++ {
				if l.ReferenceCounts[h] != [3]int{60, 60, 60} || l.ReferenceFamilies[h] != [3]int{60, 60, 60} || l.ReferenceLineages[h] != [3]int{60, 60, 60} {
					t.Fatal("reference supports")
				}
				total := 0
				for _, target := range l.RawConfusion[h] {
					for _, n := range target {
						total += n
					}
				}
				if total != 180 || l.TrueFalsePositive[h] != l.TrueOnFalse[h]+l.TrueOnUnknown[h] {
					t.Fatal("confusion/false positive denominator")
				}
			}
		}
	}
	for _, marker := range []string{"owned-dev-000", "owned-fit-000", "Amber kite", "직접 작성한", "Owned numerical fixture", "parent.rsc", "evaluation.jsonl"} {
		if bytes.Contains(data, []byte(marker)) {
			t.Fatal("private row/path exposed")
		}
	}
	for _, file := range []string{"candidate.rsc", "report.json", "cpu.pprof"} {
		st, e := os.Stat(anchor + "/owned-a/" + file)
		if e != nil || st.Mode().Perm() != 0600 || st.Size() == 0 {
			t.Fatal("file not private", file, e)
		}
	}
	st, e := os.Stat(anchor + "/owned-a")
	if e != nil || st.Mode().Perm() != 0700 {
		t.Fatal("directory not private")
	}
	candidate, e := os.ReadFile(anchor + "/owned-a/candidate.rsc")
	if e != nil || len(candidate) != 73988 || digest(candidate) != report.Candidate.SHA {
		t.Fatal("saved artifact identity")
	}
	if _, e := canonicalModel(candidate, 4240); e != nil {
		t.Fatal("numeric reload", e)
	}
	for _, file := range []struct {
		name string
		data []byte
	}{{"parent.rsc", f.parent}, {"comparator.rsc", f.comparator}, {"training.jsonl", f.training}, {"evaluation.jsonl", f.evaluation}} {
		after, e := os.ReadFile(file.name)
		if e != nil || !bytes.Equal(after, file.data) {
			t.Fatal("input mutated")
		}
	}
	if _, e := runOwned(t, f, "--out", anchor+"/owned-a"); e == nil {
		t.Fatal("output reused")
	}
	// Contradictory fresh labels/body and a NaN comparator cannot be validated
	// before the sole fit: its exact artifact must already exist on failure.
	// Their changes cannot affect training because text/targets come only from
	// the separate pinned training corpus.
	f.evaluation = bytes.Repeat([]byte{'x'}, len(f.evaluation))
	f.refs.EvaluationSHA = digest(f.evaluation)
	f.receipt = encode(t, f.refs)
	binary.LittleEndian.PutUint32(f.comparator[192:196], math.Float32bits(float32(math.NaN())))
	refreshChecksum(f.comparator)
	f.manifest.comparator = digest(f.comparator)
	f.write(t)
	if _, e := runOwned(t, f, "--out", anchor+"/owned-b"); e == nil {
		t.Fatal("malformed fresh evaluation accepted")
	}
	second, e := os.ReadFile(anchor + "/owned-b/candidate.rsc")
	if e != nil || !bytes.Equal(candidate, second) {
		t.Fatal("evaluation/comparator affected fit or were accessed before fit", e)
	}
	if _, e := os.Stat(anchor + "/owned-b/report.json"); !os.IsNotExist(e) {
		t.Fatal("failed evaluation published report")
	}
	consumed, e := os.ReadFile(anchor + "/owned-b/fit-consumption.json")
	if e != nil || !bytes.Contains(consumed, []byte(`"actual_new_fit_calls": 1`)) || !bytes.Contains(consumed, []byte(`"failure_phase": "evaluation_schema_parse"`)) {
		t.Fatal("consumed fit/failure phase not preserved", e)
	}
	// With a valid evaluation body, comparator numeric rejection still occurs
	// after the one completed fit, with the same isolated candidate bytes.
	f.evaluation = encodeRows(t, ownedRows("dev", 180))
	f.refs.EvaluationSHA = digest(f.evaluation)
	f.receipt = encode(t, f.refs)
	f.write(t)
	if _, e := runOwned(t, f, "--out", anchor+"/owned-c"); e == nil {
		t.Fatal("numeric comparator accepted")
	}
	third, e := os.ReadFile(anchor + "/owned-c/candidate.rsc")
	if e != nil || !bytes.Equal(candidate, third) {
		t.Fatal("comparator loaded before fit", e)
	}
	consumed, e = os.ReadFile(anchor + "/owned-c/fit-consumption.json")
	if e != nil || !bytes.Contains(consumed, []byte(`"failure_phase": "comparator_numeric_load"`)) {
		t.Fatal("comparator failure consumption missing", e)
	}
}
