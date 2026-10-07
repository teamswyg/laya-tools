// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimtrit"
)

func number(v int) *int { return &v }
func ownedRows() []rawRow {
	var rows []rawRow
	for i, family := range []string{"owned-fit", "owned-dev-a", "owned-dev-b"} {
		for _, locale := range []string{"ko", "en"} {
			text := "직접 작성한 숫자시험 문장입니다. 의미 정답이나 품질 자료가 아닙니다."
			if locale == "en" {
				text = "Original owned numeric fixture; not human semantic truth or quality evidence."
			}
			r := rawRow{Schema: rowSchema, ID: family + "-" + locale, Family: family, Group: "owned-fit-group", Split: "fit", Locale: locale, Text: text, Targets: []string{"true", "false", "unknown"}, RubricSHA: strings.Repeat("a", 64), AnnotationSource: "ai_semantic_reference"}
			if i > 0 {
				r.Group, r.Split = "owned-dev-group", "dev"
			}
			for h := range r.Targets {
				end := len(text)
				if r.Targets[h] == "false" {
					end = 0
				}
				r.Evidence = append(r.Evidence, evidenceInput{number(0), number(end), "Owned numeric annotation fixture, no human reference quality claim."})
			}
			rows = append(rows, r)
		}
	}
	return rows
}
func encode(t *testing.T, rows []rawRow) []byte {
	t.Helper()
	var b bytes.Buffer
	for _, r := range rows {
		if err := json.NewEncoder(&b).Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}
func writeOwned(t *testing.T, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(name, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestWholeLineageSplitAndPairedLocaleValidation(t *testing.T) {
	p, err := prepare(encode(t, ownedRows()))
	if err != nil || p.counts.Rows != 6 || p.counts.Families != 3 || p.counts.Lineages != 2 || p.counts.FitRows != 2 || p.counts.DevRows != 4 {
		t.Fatal("owned paired data", p.counts, err)
	}
	for _, change := range []func([]rawRow){
		func(r []rawRow) { r[2].Group, r[3].Group = r[0].Group, r[0].Group },
		func(r []rawRow) { r[1].Locale, r[1].ID = "ko", r[0].ID },
		func(r []rawRow) { r[2].ID = r[0].ID },
		func(r []rawRow) { r[1].Group = "different-group" },
		func(r []rawRow) { r[1].Split = "dev" },
		func(r []rawRow) { r[1].RubricSHA = strings.Repeat("b", 64) },
	} {
		rows := ownedRows()
		change(rows)
		if _, err := prepare(encode(t, rows)); err == nil {
			t.Fatal("split/locale/identity/rubric conflict accepted")
		}
	}
	// Locale-specific annotations remain independent rather than silently forced.
	rows := ownedRows()
	rows[1].Targets[2] = "true"
	if _, err := prepare(encode(t, rows)); err != nil {
		t.Fatal("distinct valid locale targets were forced together", err)
	}
}

func TestRequiredTargetsEvidenceAndUTF8SpanBoundaries(t *testing.T) {
	for _, change := range []func([]rawRow){
		func(r []rawRow) { r[0].Targets = nil },
		func(r []rawRow) { r[0].Targets = []string{"true", "false"} },
		func(r []rawRow) { r[0].Targets[0] = "planned" },
		func(r []rawRow) { r[0].Evidence = nil },
		func(r []rawRow) { r[0].Evidence[0].Start = nil },
		func(r []rawRow) { r[0].Evidence[0].End = nil },
		func(r []rawRow) { r[0].Evidence[0].Start = number(1) },
		func(r []rawRow) { r[0].Evidence[0].End = number(len(r[0].Text) + 1) },
		func(r []rawRow) { r[0].Evidence[2].End = number(0) },
		func(r []rawRow) { r[0].Evidence[1].Start, r[0].Evidence[1].End = number(3), number(3) },
		func(r []rawRow) { r[0].Evidence[0].Reason = "" },
		func(r []rawRow) { r[0].Evidence[0].Reason = strings.Repeat("x", 513) },
		func(r []rawRow) { r[0].Text = strings.Repeat("x", 4097) },
	} {
		rows := ownedRows()
		change(rows)
		if _, err := prepare(encode(t, rows)); err == nil {
			t.Fatal("missing/malformed target or evidence accepted")
		}
	}
	data := encode(t, ownedRows())
	missing := bytes.Replace(data, []byte(`"targets":["true","false","unknown"],`), nil, 1)
	if bytes.Equal(missing, data) {
		t.Fatal("owned missing-field mutation failed")
	}
	for _, bad := range [][]byte{missing, append([]byte{0xff}, data...), bytes.Repeat([]byte{'x'}, inputBudget+1)} {
		if _, err := prepare(bad); err == nil {
			t.Fatal("missing field, invalid UTF8 or budget accepted")
		}
	}
}

// Build an explicitly owned numerical parent without training or importing any
// model. Public RSC fields are filled with original toy values, not semantic
// reference weights. Its inherited counter is an independent provenance fixture.
func ownedParent(t *testing.T, steps uint64) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := statehintclaims.NewModel().Save(&b); err != nil {
		t.Fatal(err)
	}
	data := b.Bytes()
	if len(data) != parentBytes {
		t.Fatal("parent fixture container size")
	}
	binary.LittleEndian.PutUint64(data[16:24], steps)
	for i := 0; i < statehintclaimtrit.WeightTritCount; i++ {
		value := float32(i%17-8) * .003
		binary.LittleEndian.PutUint32(data[192+i*4:196+i*4], math.Float32bits(value))
	}
	offset := 192 + statehintclaimtrit.WeightTritCount*4
	for i := 0; i < 9; i++ {
		binary.LittleEndian.PutUint32(data[offset+i*4:offset+i*4+4], math.Float32bits(float32(i%3-1)*.05))
	}
	refreshChecksum(data)
	return data
}
func refreshChecksum(data []byte) {
	end := len(data) - sha256.Size
	sum := sha256.Sum256(data[:end])
	copy(data[end:], sum[:])
}
func ownedArgs(input string, data []byte, parent []byte) []string {
	return []string{"--parent", "owned.rsc", "--parent-sha256", digest(parent), "--input", input, "--input-sha256", digest(data)}
}
func TestPinnedCheckIsStructuralAndCreatesNoOutputOrProfile(t *testing.T) {
	t.Chdir(t.TempDir())
	data := encode(t, ownedRows())
	// Zero parent steps deliberately prove check does not require a fitted model.
	parent := ownedParent(t, 0)
	writeOwned(t, "owned.jsonl", data)
	writeOwned(t, "owned.rsc", parent)
	args := append(ownedArgs("owned.jsonl", data, parent), "--check")
	var output, errorOutput bytes.Buffer
	if err := run(args, &output, &errorOutput); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Status string `json:"status"`
		Counts counts `json:"counts"`
	}
	if json.Unmarshal(output.Bytes(), &result) != nil || result.Counts.Rows != 6 || result.Counts.FitLineages != 1 || result.Counts.DevLineages != 1 || !strings.Contains(result.Status, "no model construction, FromFloat, Fit, predictions or outputs") {
		t.Fatal("check exceeded structural work", output.String())
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("check created output")
	}
	// A numeric NaN with a valid pinned/checksummed byte container still passes
	// structural check; an actual run must reject it via the full numeric Load.
	invalidNumeric := append([]byte(nil), parent...)
	binary.LittleEndian.PutUint32(invalidNumeric[192:196], math.Float32bits(float32(math.NaN())))
	refreshChecksum(invalidNumeric)
	writeOwned(t, "numeric-invalid.rsc", invalidNumeric)
	numericArgs := []string{"--parent", "numeric-invalid.rsc", "--parent-sha256", digest(invalidNumeric), "--input", "owned.jsonl", "--input-sha256", digest(data), "--check"}
	if err := run(numericArgs, &output, &errorOutput); err != nil {
		t.Fatal("structural check initialized numeric model", err)
	}
	for _, bad := range [][]string{
		{"--parent", "owned.rsc", "--parent-sha256", strings.Repeat("0", 64), "--input", "owned.jsonl", "--input-sha256", digest(data), "--check"},
		{"--parent", "owned.rsc", "--parent-sha256", digest(parent), "--input", "owned.jsonl", "--input-sha256", strings.Repeat("0", 64), "--check"},
		append(append([]string(nil), args...), "--cpu-profile"),
		append(append([]string(nil), args...), "--out", anchor+"/bad"),
		{"--model", "private-marker"}, {"--epochs", "1"}, {"--calibration", "private-marker"}, {"--parent", "../private-marker"},
	} {
		var out, errOut bytes.Buffer
		err := run(bad, &out, &errOut)
		if err == nil || out.Len() != 0 || strings.Contains(errOut.String()+err.Error(), "private-marker") {
			t.Fatal("bad operation accepted or private detail leaked")
		}
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("invalid check created output")
	}
}
func TestPinnedRegularFilesRejectContainerAndSymlinkChanges(t *testing.T) {
	t.Chdir(t.TempDir())
	parent := ownedParent(t, 2120)
	data := encode(t, ownedRows())
	writeOwned(t, "owned.rsc", parent)
	writeOwned(t, "owned.jsonl", data)
	if err := os.Mkdir("regular", 0700); err != nil {
		t.Fatal(err)
	}
	writeOwned(t, "regular/parent.rsc", parent)
	writeOwned(t, "regular/input.jsonl", data)
	for _, pair := range [][2]string{{"owned.rsc", "leaf.rsc"}, {"owned.jsonl", "leaf.jsonl"}, {"regular", "linked"}} {
		if err := os.Symlink(pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	for _, name := range []string{"leaf.rsc", "linked/parent.rsc", "../owned.rsc"} {
		if _, err := readPinned(root, name, digest(parent), parentBytes, parentBytes); err == nil {
			t.Fatal("symlink/nonlocal parent accepted")
		}
	}
	if _, err := readPinned(root, "leaf.jsonl", digest(data), inputBudget, 0); err == nil {
		t.Fatal("leaf symlink input accepted")
	}
	if _, err := readPinned(root, "linked/input.jsonl", digest(data), inputBudget, 0); err == nil {
		t.Fatal("ancestor symlink input accepted")
	}
	for _, mutated := range [][]byte{append(append([]byte(nil), parent...), 0), parent[:len(parent)-1], bytes.Repeat([]byte{1}, parentBytes)} {
		writeOwned(t, "bad.rsc", mutated)
		var output, errorOutput bytes.Buffer
		if err := run([]string{"--parent", "bad.rsc", "--parent-sha256", digest(mutated), "--input", "owned.jsonl", "--input-sha256", digest(data), "--check"}, &output, &errorOutput); err == nil || output.Len() != 0 {
			t.Fatal("bad parent size/type/checksum accepted")
		}
	}
	wrongType := append([]byte(nil), parent...)
	copy(wrongType[:4], "RQT\x00")
	refreshChecksum(wrongType)
	writeOwned(t, "wrong-type.rsc", wrongType)
	if parentShape(wrongType) {
		t.Fatal("wrong exact-size container type accepted")
	}
	if _, err := statehintclaims.Load(bytes.NewReader(append(append([]byte(nil), parent...), 0))); err == nil {
		t.Fatal("parent trailing bytes accepted")
	}
	if _, err := statehintclaimtrit.Load(bytes.NewReader(parent)); err == nil {
		t.Fatal("ternary Load accepted float format")
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("invalid pin created output")
	}
}
func TestFreshPrivateOutputRejectsSymlinksAndReuses(t *testing.T) {
	t.Chdir(t.TempDir())
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if err := os.Mkdir("owned-target", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("owned-target", ".cache"); err != nil {
		t.Fatal(err)
	}
	if out, err := privateOutput(root, anchor+"/bad"); err == nil {
		out.Close()
		t.Fatal("symlink cache accepted")
	}
	if err := os.Remove(".cache"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(".cache", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../owned-target", anchor); err != nil {
		t.Fatal(err)
	}
	if out, err := privateOutput(root, anchor+"/bad"); err == nil {
		out.Close()
		t.Fatal("symlink study root accepted")
	}
	if err := os.Remove(anchor); err != nil {
		t.Fatal(err)
	}
	out, err := privateOutput(root, anchor+"/fresh")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeBytes(out, "owned.marker", []byte("owned")); err != nil {
		t.Fatal(err)
	}
	out.Close()
	if reused, err := privateOutput(root, anchor+"/fresh"); err == nil {
		reused.Close()
		t.Fatal("fresh output reused")
	}
	marker, err := os.ReadFile(anchor + "/fresh/owned.marker")
	if err != nil || string(marker) != "owned" {
		t.Fatal("reuse changed output")
	}
}

// The fixed-size corpus below is generated in the test itself. Repetition and
// target assignment exercise numeric plumbing only, never semantic quality.
func fixedOwnedRows() []rawRow {
	rows := make([]rawRow, 0, 1920)
	for splitIndex, split := range []string{"fit", "dev"} {
		families := 840
		if splitIndex == 1 {
			families = 120
		}
		for i := 0; i < families; i++ {
			family := fmt.Sprintf("owned-%s-%03d", split, i)
			group := fmt.Sprintf("owned-%s-group-%03d", split, i/3)
			for _, locale := range []string{"ko", "en"} {
				text := "직접 작성한 숫자시험 문장입니다."
				if locale == "en" {
					text = "Amber kite original owned numerical fixture."
				}
				targets := []string{"true", "false", "unknown"}
				if i%2 == 1 {
					targets = []string{"false", "unknown", "true"}
				}
				r := rawRow{Schema: rowSchema, ID: family + "-" + locale, Family: family, Group: group, Split: split, Locale: locale, Text: text, Targets: targets, RubricSHA: strings.Repeat("b", 64), AnnotationSource: "ai_semantic_reference"}
				for _, target := range targets {
					end := len(text)
					if target == "false" {
						end = 0
					}
					r.Evidence = append(r.Evidence, evidenceInput{number(0), number(end), "Owned numerical fixture only; no semantic quality evidence."})
				}
				rows = append(rows, r)
			}
		}
	}
	return rows
}
func readReport(t *testing.T, path string) studyReport {
	t.Helper()
	data, err := os.ReadFile(path)
	var r studyReport
	if err != nil || json.Unmarshal(data, &r) != nil {
		t.Fatal("owned report", err)
	}
	return r
}
func TestOwnedSyntheticFixedRunSeparatesCountersAndDevCannotTrain(t *testing.T) {
	t.Chdir(t.TempDir())
	rows := fixedOwnedRows()
	data := encode(t, rows)
	parent := ownedParent(t, 2120)
	writeOwned(t, "owned.rsc", parent)
	writeOwned(t, "owned-a.jsonl", data)
	args := append(ownedArgs("owned-a.jsonl", data, parent), "--out", anchor+"/owned-a", "--cpu-profile")
	var output, errorOutput bytes.Buffer
	if err := run(args, &output, &errorOutput); err != nil {
		t.Fatal("owned synthetic fixed run", err)
	}
	a := readReport(t, anchor+"/owned-a/report.json")
	if a.Qualified || a.Selection || a.Calibration || a.CalTestAccessed || a.StateWrites != 0 || a.OSRSSMeasured || a.AnnotationSource != "ai_semantic_reference" || a.Status != "development_research_only" || !a.DevPreviouslyExposed || a.FitCalls != 2 || a.FloatFitCalls != 1 || a.QATFitCalls != 1 || a.ProjectionFitCalls != 0 || !a.CPUProfile || a.CPUProfileScope != "matched_float_and_ternary_qat_new_fits_plus_development_evaluation_save_reload" || a.Recipe != fixedRecipe() || !a.Parent.ReadOnlyExact || a.Parent.SHA != digest(parent) {
		t.Fatal("numeric fixture became qualification or recipe changed")
	}
	if a.Counts.FitRows != 1680 || a.Counts.DevRows != 240 || a.Counts.FitFamilies != 840 || a.Counts.DevFamilies != 120 || a.Counts.FitLineages != 280 || a.Counts.DevLineages != 40 {
		t.Fatal("fixed scope changed", a.Counts)
	}
	if a.Fit.Samples != 1680 || a.Fit.Batches != 2120 || a.Fit.BaseTrainingSteps != 2120 || a.Fit.NewOptimizerSteps != 2120 || a.Fit.TrainingSteps != 4240 {
		t.Fatal("base/new optimizer counters conflated", a.Fit)
	}
	if a.FloatFit.Samples != 1680 || a.FloatFit.Batches != 2120 || a.FloatFit.BaseTrainingSteps != 2120 || a.FloatFit.NewOptimizerSteps != 2120 || a.FloatFit.TrainingSteps != 4240 || a.FloatFit.Seed != a.Fit.Seed || a.FloatFit.Initialization != a.Fit.Initialization || a.FloatFit.Epochs != a.Fit.Epochs {
		t.Fatal("matched float recipe/counters diverged", a.FloatFit)
	}
	mf := a.MatchedFloat
	if mf.Mode != "matched_float_warm_continuation" || mf.Quantized || mf.ParentSHA != digest(parent) || mf.BaseTrainingSteps != 2120 || mf.NewOptimizerSteps != 2120 || mf.TrainingSteps != 4240 || mf.ParentInitializationSeed != 1729 || mf.AdaptationSeed != 1729 || mf.Temperature != 1 || mf.FeatureSchema != statehintclaims.FeatureSchema || mf.FitCalls != 1 || mf.ArtifactBytes != 73988 || !mf.ReloadByteExact || !mf.ReloadPredictExact || !mf.ReloadScoresExact || mf.DevCalls != 240 || mf.ScoreCalls != 240 || mf.ReloadCalls != 240 || mf.ReloadScoreCalls != 240 || mf.Dev.Rows != 240 || mf.Difference.Rows != 240 {
		t.Fatal("matched float provenance/parity invalid", mf)
	}
	if a.PTQ.DifferenceMatched != nil || a.QAT.DifferenceMatched == nil || a.QAT.DifferenceMatched.Rows != 240 || a.QAT.Difference.Rows != 240 {
		t.Fatal("QAT lacks separate baseline and matched-control comparisons")
	}

	if a.PTQ.Metadata.Mode != "ptq" || a.PTQ.Metadata.BaseTrainingSteps != 2120 || a.PTQ.Metadata.NewOptimizerSteps != 0 || a.PTQ.Metadata.TrainingSteps != 2120 || a.PTQ.FitCalls != 0 || a.PTQ.ProjectionFitCalls != 0 || a.QAT.Metadata.Mode != "qat" || a.QAT.Metadata.NewOptimizerSteps != 2120 || a.QAT.Metadata.TrainingSteps != 4240 || a.QAT.FitCalls != 1 || a.QAT.ProjectionFitCalls != 0 {
		t.Fatal("projection/adaptation provenance changed")
	}
	for _, v := range []variantReport{a.PTQ, a.QAT} {
		if v.ArtifactBytes != 4023 || !v.ReloadByteExact || !v.ReloadPredictExact || !v.ReloadScoresExact || v.DevCalls != 240 || v.ScoreCalls != 240 || v.ReloadCalls != 240 || v.ReloadScoreCalls != 240 || v.Metadata.ParentSHA256 != digest(parent) || v.Metadata.QualityQualified || v.Metadata.BitNetLLM || v.Metadata.W158A8 || v.Metadata.StateAuthority {
			t.Fatal("parity/diagnostic/model claims invalid", v.Metadata)
		}
		for _, l := range v.Dev.Locales {
			if l.Rows != 120 {
				t.Fatal("locale dev denominator")
			}
			for h := range l.RawConfusion {
				n := 0
				for _, target := range l.RawConfusion[h] {
					for _, count := range target {
						n += count
					}
				}
				if n != 120 || l.TrueProposed[h] == 0 && l.PrecisionDefined[h] {
					t.Fatal("confusion or undefined precision")
				}
			}
		}
	}
	beforeParent, err := os.ReadFile("owned.rsc")
	if err != nil || !bytes.Equal(beforeParent, parent) {
		t.Fatal("parent mutated")
	}
	for _, name := range []string{"matched_float.rsc", "ptq.rqt", "qat.rqt", "report.json", "cpu.pprof"} {
		st, err := os.Stat(anchor + "/owned-a/" + name)
		if err != nil || st.Mode().Perm() != 0600 || st.Size() == 0 {
			t.Fatal("output not private", name, err)
		}
	}
	st, err := os.Stat(anchor + "/owned-a")
	if err != nil || st.Mode().Perm() != 0700 {
		t.Fatal("output directory not private")
	}
	reportBytes, err := os.ReadFile(anchor + "/owned-a/report.json")
	if err != nil || bytes.Contains(reportBytes, []byte(rows[0].ID)) || bytes.Contains(reportBytes, []byte(rows[0].Group)) || bytes.Contains(reportBytes, []byte(rows[0].Text)) || bytes.Contains(reportBytes, []byte(rows[0].Evidence[0].Reason)) {
		t.Fatal("raw private detail in report")
	}
	ptqBytes, err := os.ReadFile(anchor + "/owned-a/ptq.rqt")
	if err != nil {
		t.Fatal(err)
	}
	qatBytes, err := os.ReadFile(anchor + "/owned-a/qat.rqt")
	if err != nil || bytes.Equal(ptqBytes, qatBytes) {
		t.Fatal("PTQ and QAT artifact mode indistinguishable")
	}
	if _, err := statehintclaimtrit.Load(bytes.NewReader(append(append([]byte(nil), qatBytes...), 0))); err == nil {
		t.Fatal("RQT trailing bytes accepted")
	}
	matchedBytes, err := os.ReadFile(anchor + "/owned-a/matched_float.rsc")
	if err != nil || len(matchedBytes) != parentBytes || digest(matchedBytes) != mf.ModelSHA {
		t.Fatal("matched float artifact identity/size")
	}
	matchedModel, err := statehintclaims.Load(bytes.NewReader(matchedBytes))
	if err != nil || matchedModel.TrainingSteps() != 4240 || matchedModel.InitializationSeed() != 1729 {
		t.Fatal("matched float artifact metadata", err)
	}
	originalModel, err := statehintclaims.Load(bytes.NewReader(parent))
	if err != nil {
		t.Fatal(err)
	}
	qatModel, err := statehintclaimtrit.Load(bytes.NewReader(qatBytes))
	if err != nil {
		t.Fatal(err)
	}
	preparedInput, err := prepare(data)
	if err != nil {
		t.Fatal(err)
	}
	matchedObserved, _, _, err := observeFloat(matchedModel, preparedInput)
	if err != nil {
		t.Fatal(err)
	}
	baselineObserved, _, _, err := observeFloat(originalModel, preparedInput)
	if err != nil {
		t.Fatal(err)
	}
	qatObserved, _, _, err := observeTrit(qatModel, preparedInput)
	if err != nil {
		t.Fatal(err)
	}
	wantMatched, err := compare(preparedInput, matchedObserved, qatObserved)
	if err != nil || *a.QAT.DifferenceMatched != wantMatched {
		t.Fatal("QAT matched comparison did not use matched-control predictions")
	}
	wantBaseline, err := compare(preparedInput, baselineObserved, qatObserved)
	if err != nil || a.QAT.Difference != wantBaseline {
		t.Fatal("QAT baseline comparison lost independent parent reference")
	}

	// Change only dev labels. All unknown spans are already nonempty, so this is
	// a valid annotation mutation that must not enter the warm-fit samples.
	for i := 1680; i < len(rows); i++ {
		for h, target := range rows[i].Targets {
			if target == "unknown" {
				rows[i].Targets[h] = "true"
			}
		}
	}
	changed := encode(t, rows)
	writeOwned(t, "owned-b.jsonl", changed)
	output.Reset()
	if err := run(append(ownedArgs("owned-b.jsonl", changed, parent), "--out", anchor+"/owned-b"), &output, &errorOutput); err != nil {
		t.Fatal(err)
	}
	b := readReport(t, anchor+"/owned-b/report.json")
	if a.MatchedFloat.ModelSHA != b.MatchedFloat.ModelSHA || a.PTQ.ModelSHA != b.PTQ.ModelSHA || a.QAT.ModelSHA != b.QAT.ModelSHA || a.FloatFit != b.FloatFit || a.Fit != b.Fit || a.InputSHA == b.InputSHA {
		t.Fatal("dev labels changed training or parent projection")
	}
	if err := run(args, &output, &errorOutput); err == nil {
		t.Fatal("existing run overwritten")
	}
	after, err := os.ReadFile(filepath.Join(anchor, "owned-a", "qat.rqt"))
	if err != nil || !bytes.Equal(qatBytes, after) {
		t.Fatal("failed overwrite mutated artifact")
	}
}
func TestActualRunRejectsWrongFixedScopeBeforeOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	parent := ownedParent(t, 2120)
	data := encode(t, ownedRows())
	writeOwned(t, "owned.rsc", parent)
	writeOwned(t, "owned.jsonl", data)
	var output, errorOutput bytes.Buffer
	if err := run(append(ownedArgs("owned.jsonl", data, parent), "--out", anchor+"/bad"), &output, &errorOutput); err == nil {
		t.Fatal("nonfixed corpus accepted")
	}
	if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
		t.Fatal("rejected scope created output")
	}
}
func TestAggregateGateDifferencesSeparateFalseAndUnknownTargets(t *testing.T) {
	p, err := prepare(encode(t, ownedRows()))
	if err != nil {
		t.Fatal(err)
	}
	before, after := make([]observation, len(p.dev)), make([]observation, len(p.dev))
	for i := range before {
		for h := 0; h < 3; h++ {
			before[i].Heads[h] = statehintclaims.HeadPrediction{Winner: statehintclaims.Unknown, State: statehintclaims.Unknown}
			after[i].Heads[h] = statehintclaims.HeadPrediction{Winner: statehintclaims.True, State: statehintclaims.True}
		}
	}
	report, err := compare(p, before, after)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range report.Locales {
		if l.Rows != 2 || l.TrueAdded != [3]int{2, 2, 2} || l.TrueAddedOnFalse != [3]int{0, 2, 0} || l.TrueAddedOnUnknown != [3]int{0, 0, 2} || l.RawWinnerChanges != [3]int{2, 2, 2} || l.GateStateChanges != [3]int{2, 2, 2} {
			t.Fatal("false/unknown gate differences conflated", l)
		}
	}
}

func TestActualRunRejectsUntrainedOrNonNumericParentBeforeOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	data := encode(t, fixedOwnedRows())
	writeOwned(t, "owned.jsonl", data)
	untrained := ownedParent(t, 0)
	nonNumeric := ownedParent(t, 2120)
	binary.LittleEndian.PutUint32(nonNumeric[192:196], math.Float32bits(float32(math.NaN())))
	refreshChecksum(nonNumeric)
	for _, parent := range [][]byte{untrained, nonNumeric} {
		writeOwned(t, "owned.rsc", parent)
		var output, errorOutput bytes.Buffer
		if err := run(append(ownedArgs("owned.jsonl", data, parent), "--out", anchor+"/invalid-parent"), &output, &errorOutput); err == nil || output.Len() != 0 {
			t.Fatal("invalid numeric or untrained parent accepted")
		}
		if _, err := os.Stat(".cache"); !os.IsNotExist(err) {
			t.Fatal("rejected parent created output")
		}
	}
}
