// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"unsafe"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimtrit"
)

// Everything created here is an original, explicitly synthetic test fixture.
// Tests neither find nor read any real model, dataset, fit output, or task text.
type testInputs struct {
	float, matched, ptq, qat []byte
	fixture                  []byte
	opts                     options
}

func testSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func testRows(count int) []map[string]any {
	rows := make([]map[string]any, count)
	for i := range rows {
		locale, text := "en", fmt.Sprintf("Original synthetic English benchmark fixture number %02d requests a status update.", i)
		if i%2 == 0 {
			locale, text = "ko", fmt.Sprintf("독창적인 합성 벤치마크 예문 %02d번은 현재 상황의 설명을 요청합니다.", i)
		}
		rows[i] = map[string]any{
			"id": fmt.Sprintf("original_synthetic_test_id_%02d", i), "text": text, "locale": locale, "internal_split": "fit",
			"targets":    []string{"unknown", "unknown", "unknown"},
			"annotation": "original test-only ignored annotation marker",
			"raw_source": "original test-only ignored raw source marker",
		}
	}
	return rows
}

func encodeRows(t *testing.T, rows []map[string]any) []byte {
	t.Helper()
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	for _, row := range rows {
		if err := encoder.Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	return data.Bytes()
}

func writeTestInput(t *testing.T, directory, name string, data []byte) pinnedFile {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return pinnedFile{path: path, pin: testSHA(data)}
}

func makeTestInputs(t *testing.T) testInputs {
	t.Helper()
	parent := statehintclaims.NewModel()
	var floatData bytes.Buffer
	if err := parent.Save(&floatData); err != nil {
		t.Fatal(err)
	}
	ptq, err := statehintclaimtrit.FromFloat(parent, testSHA(floatData.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	var ptqData bytes.Buffer
	if err := ptq.Save(&ptqData); err != nil {
		t.Fatal(err)
	}
	qat := ptq.Clone()
	samples := []statehintclaims.Sample{{
		Text: "This original synthetic test sentence describes a toy benchmark.",
		Targets: [statehintclaims.HeadCount]statehintclaims.State{
			statehintclaims.Unknown, statehintclaims.Unknown, statehintclaims.Unknown,
		},
	}}
	if _, err := qat.WarmFit(parent, samples, new(statehintclaimtrit.TrainingWorkspace)); err != nil {
		t.Fatal(err)
	}
	var qatData bytes.Buffer
	if err := qat.Save(&qatData); err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	// This fixture checks the CLI's float loading path and makes no claim of
	// matched training; production matched identity comes from explicit pins.
	inputs := testInputs{float: floatData.Bytes(), matched: bytes.Clone(floatData.Bytes()), ptq: ptqData.Bytes(), qat: qatData.Bytes(), fixture: encodeRows(t, testRows(24))}
	inputs.opts = options{
		mode: "float", iterations: 16,
		floatModel:   writeTestInput(t, directory, "original-float.rsc", inputs.float),
		matchedModel: writeTestInput(t, directory, "original-matched.rsc", inputs.matched),
		ptqModel:     writeTestInput(t, directory, "original-ptq.rqt", inputs.ptq),
		qatModel:     writeTestInput(t, directory, "original-qat.rqt", inputs.qat),
		fixture:      writeTestInput(t, directory, "original-fit.jsonl", inputs.fixture),
	}
	return inputs
}

func argsFor(opts options) []string {
	return []string{
		"--mode", opts.mode,
		"--float-model", opts.floatModel.path, "--float-sha256", opts.floatModel.pin,
		"--matched-model", opts.matchedModel.path, "--matched-sha256", opts.matchedModel.pin,
		"--ptq-model", opts.ptqModel.path, "--ptq-sha256", opts.ptqModel.pin,
		"--qat-model", opts.qatModel.path, "--qat-sha256", opts.qatModel.pin,
		"--fixture", opts.fixture.path, "--fixture-sha256", opts.fixture.pin,
		"--iterations", fmt.Sprint(opts.iterations),
	}
}

func TestRunModesAndOutputExclusion(t *testing.T) {
	inputs := makeTestInputs(t)
	var first fixtureSummary
	for _, mode := range []string{"float", "matched", "ptq", "qat"} {
		t.Run(mode, func(t *testing.T) {
			opts := inputs.opts
			opts.mode = mode
			var output bytes.Buffer
			if err := run(argsFor(opts), &output); err != nil {
				t.Fatal(err)
			}
			var got report
			if err := json.Unmarshal(output.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Mode != mode || got.Iterations != 16 || got.WarmupIterations != 32 || got.Fixture.SelectedCount != 16 || !got.DevelopmentOnly {
				t.Fatalf("unexpected report mode, counts, or development scope: %+v", got)
			}
			if mode == "float" || mode == "matched" {
				if mode == "float" {
					first = got.Fixture
				} else if got.Fixture != first || got.ParentSHA256 != "" {
					t.Fatal("matched fixture order differed or an unverified parent was asserted")
				}
				if got.ModelArtifactBytes != statehintclaims.ArtifactBytes || got.ModelRuntimeFixedBytes != uint64(unsafe.Sizeof(statehintclaims.Model{})) || got.ModelImmutableBytes != 0 || got.SharedStaticDecoderBytes != 0 || got.WorkspaceFixedBytes != uint64(unsafe.Sizeof(statehintclaims.Workspace{})) {
					t.Fatal("float model storage accounting mismatch")
				}
			} else {
				if got.Fixture != first || got.ParentSHA256 != opts.floatModel.pin {
					t.Fatal("fixture order or parent identity differed between modes")
				}
				if got.ModelArtifactBytes != statehintclaimtrit.ArtifactBytes || got.ModelRuntimeFixedBytes != uint64(unsafe.Sizeof(statehintclaimtrit.Model{})) || got.ModelImmutableBytes != 64 || got.SharedStaticDecoderBytes != 1215 || got.WorkspaceFixedBytes != uint64(unsafe.Sizeof(statehintclaimtrit.Workspace{})) {
					t.Fatal("ternary model storage accounting mismatch")
				}
			}
			if got.ModelStorageBytes != got.ModelRuntimeFixedBytes+got.ModelImmutableBytes || got.WorkspaceFixedBytes == 0 || got.LoadElapsedNS < 0 || got.ElapsedNS <= 0 || got.NSPerOp <= 0 {
				t.Fatal("storage or timing accounting missing")
			}
			if len(got.PredictionResultChecksum) != 16 || got.FloatSHA256 != opts.floatModel.pin || got.MatchedSHA256 != opts.matchedModel.pin || got.PTQSHA256 != opts.ptqModel.pin || got.QATSHA256 != opts.qatModel.pin || got.FixtureSHA256 != opts.fixture.pin {
				t.Fatal("report checksum or input pins missing")
			}
			for _, forbidden := range []string{
				opts.floatModel.path, opts.matchedModel.path, opts.ptqModel.path, opts.qatModel.path, opts.fixture.path,
				"original_synthetic_test_id", "Original synthetic English", "독창적인 합성",
				"ignored annotation marker", "ignored raw source marker", "probabilities", "unknown_reason",
			} {
				if strings.Contains(output.String(), forbidden) {
					t.Fatalf("report included raw input or prediction material: %q", forbidden)
				}
			}
		})
	}
}

func TestFixtureSelectionIndependentOfAnnotationsAndOrder(t *testing.T) {
	rows := testRows(24)
	wantRequests, wantSummary, err := selectFixtures(encodeRows(t, rows))
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range rows {
		row["targets"] = map[string]any{"invalid_for_training_but_ignored_here": i}
		row["annotation"] = []any{nil, i, "different original ignored test marker"}
		row["raw_source"] = "a different original ignored raw source test marker"
	}
	slices.Reverse(rows)
	gotRequests, gotSummary, err := selectFixtures(encodeRows(t, rows))
	if err != nil {
		t.Fatal(err)
	}
	if gotRequests != wantRequests || gotSummary != wantSummary {
		t.Fatal("fixture selection depended on labels, annotations, raw sources, or file order")
	}
	// Independently sort the hex SHA-256 ordering keys and check the first 16.
	sort.Slice(rows, func(i, j int) bool {
		return testSHA([]byte("claims-trit-bench-v1:"+rows[i]["id"].(string))) < testSHA([]byte("claims-trit-bench-v1:"+rows[j]["id"].(string)))
	})
	for i := 0; i < fixtureCount; i++ {
		if gotRequests[i].text != rows[i]["text"] || gotSummary.LocaleOrder[i] != rows[i]["locale"] || gotSummary.TextByteCounts[i] != len(gotRequests[i].text) {
			t.Fatalf("selected request %d did not follow the fixed hash order", i)
		}
	}
	rows[0]["text"] = "A changed original synthetic text keeps the ID but changes the fixture identity."
	_, changed, err := selectFixtures(encodeRows(t, rows))
	if err != nil || changed.SelectedSHA256 == gotSummary.SelectedSHA256 {
		t.Fatal("selected fixture hash failed to bind selected text")
	}
}

func TestFixtureScopeAndBounds(t *testing.T) {
	cases := []struct {
		name   string
		mutate func([]map[string]any) []map[string]any
	}{
		{"dev after selected quota", func(rows []map[string]any) []map[string]any { rows[23]["internal_split"] = "dev"; return rows }},
		{"eval", func(rows []map[string]any) []map[string]any { rows[0]["internal_split"] = "eval"; return rows }},
		{"test", func(rows []map[string]any) []map[string]any { rows[0]["internal_split"] = "test"; return rows }},
		{"train alias", func(rows []map[string]any) []map[string]any { rows[0]["internal_split"] = "train"; return rows }},
		{"split alias", func(rows []map[string]any) []map[string]any {
			delete(rows[0], "internal_split")
			rows[0]["split"] = "fit"
			return rows
		}},
		{"missing split", func(rows []map[string]any) []map[string]any { delete(rows[0], "internal_split"); return rows }},
		{"too few", func(rows []map[string]any) []map[string]any { return rows[:15] }},
		{"duplicate ID", func(rows []map[string]any) []map[string]any { rows[23]["id"] = rows[0]["id"]; return rows }},
		{"missing ID", func(rows []map[string]any) []map[string]any { delete(rows[0], "id"); return rows }},
		{"missing text", func(rows []map[string]any) []map[string]any { delete(rows[0], "text"); return rows }},
		{"oversize text", func(rows []map[string]any) []map[string]any {
			rows[0]["text"] = strings.Repeat("x", statehintclaims.MaxTextBytes+1)
			return rows
		}},
		{"other locale", func(rows []map[string]any) []map[string]any { rows[0]["locale"] = "fr"; return rows }},
		{"missing locale", func(rows []map[string]any) []map[string]any { delete(rows[0], "locale"); return rows }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := selectFixtures(encodeRows(t, tc.mutate(testRows(24)))); err == nil {
				t.Fatal("invalid fixture scope or bounds accepted")
			}
		})
	}
	for _, data := range [][]byte{nil, []byte("not JSON\n"), {255}, append(encodeRows(t, testRows(24)), '\n'), []byte(strings.Repeat("x", maxFixtureLine+1))} {
		if _, _, err := selectFixtures(data); err == nil {
			t.Fatal("invalid or unbounded JSONL accepted")
		}
	}
}

func TestModeAndParentMismatch(t *testing.T) {
	inputs := makeTestInputs(t)
	for _, tc := range []struct {
		data []byte
		mode string
		pin  string
	}{
		{inputs.ptq, "qat", inputs.opts.floatModel.pin},
		{inputs.qat, "ptq", inputs.opts.floatModel.pin},
		{inputs.ptq, "ptq", strings.Repeat("0", 64)},
		{inputs.qat, "qat", strings.Repeat("0", 64)},
		{inputs.float, "ptq", inputs.opts.floatModel.pin},
		{inputs.ptq, "float", inputs.opts.floatModel.pin},
		{inputs.ptq, "matched", inputs.opts.floatModel.pin},
	} {
		if _, err := loadPredictor(tc.data, tc.mode, tc.pin); err == nil {
			t.Fatalf("accepted wrong mode %s or parent", tc.mode)
		}
	}
}

func TestRequiredPinsAndIterationBounds(t *testing.T) {
	valid := options{
		mode: "float", iterations: defaultIterations,
		floatModel:   pinnedFile{"float.test", strings.Repeat("a", 64)},
		matchedModel: pinnedFile{"matched.test", strings.Repeat("e", 64)},
		ptqModel:     pinnedFile{"ptq.test", strings.Repeat("b", 64)},
		qatModel:     pinnedFile{"qat.test", strings.Repeat("c", 64)},
		fixture:      pinnedFile{"fit.test", strings.Repeat("d", 64)},
	}
	for _, count := range []int{1, defaultIterations, maxIterations} {
		opts := valid
		opts.iterations = count
		got, help, err := parseOptions(argsFor(opts))
		if err != nil || help || got.iterations != count {
			t.Fatalf("valid iterations %d: %+v, %v, %v", count, got, help, err)
		}
	}
	for _, count := range []int{-1, 0, maxIterations + 1} {
		opts := valid
		opts.iterations = count
		if _, _, err := parseOptions(argsFor(opts)); err == nil {
			t.Fatalf("iterations %d accepted", count)
		}
	}
	for _, mutate := range []func(*options){
		func(o *options) { o.mode = "invalid_test_mode" },
		func(o *options) { o.floatModel.path = "" },
		func(o *options) { o.matchedModel.path = "" },
		func(o *options) { o.ptqModel.path = "" },
		func(o *options) { o.qatModel.path = "" },
		func(o *options) { o.fixture.path = "" },
		func(o *options) { o.floatModel.pin = "" },
		func(o *options) { o.matchedModel.pin = "" },
		func(o *options) { o.ptqModel.pin = "" },
		func(o *options) { o.qatModel.pin = "" },
		func(o *options) { o.fixture.pin = "invalid" },
		func(o *options) { o.fixture.path = "https://invalid.example/original-test-only" },
		func(o *options) { o.cpuProfile = true },
	} {
		opts := valid
		mutate(&opts)
		args := argsFor(opts)
		if opts.cpuProfile {
			args = append(args, "--cpu-profile")
		}
		if _, _, err := parseOptions(args); err == nil {
			t.Fatal("missing input, invalid pin/mode, remote input, or unsafe profile configuration accepted")
		}
	}
	// Omitting the flag uses the fixed 20,000 default rather than deriving a
	// count from a model or dataset.
	args := argsFor(valid)
	got, _, err := parseOptions(args[:len(args)-2])
	if err != nil || got.iterations != 20000 {
		t.Fatal("iteration default changed")
	}
	var output bytes.Buffer
	if err := run([]string{"--iterations", "/original/test-only/path-marker"}, &output); err == nil || strings.Contains(err.Error(), "path-marker") || output.Len() != 0 {
		t.Fatal("invalid flags exposed a caller argument or emitted a report")
	}
}

func TestPinnedFileVerification(t *testing.T) {
	directory := t.TempDir()
	data := []byte("Original synthetic local file content for hashing tests.")
	input := writeTestInput(t, directory, "original.test", data)
	for _, keep := range []bool{false, true} {
		got, size, err := readPinned(input, int64(len(data)), keep)
		if err != nil || size != len(data) || (keep && !bytes.Equal(got, data)) || (!keep && len(got) != 0) {
			t.Fatal("valid local pinned file failed")
		}
	}
	symlink := filepath.Join(directory, "symlink.test")
	if err := os.Symlink(input.path, symlink); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		input pinnedFile
		limit int64
	}{
		{pinnedFile{input.path, strings.Repeat("0", 64)}, int64(len(data))},
		{input, int64(len(data) - 1)},
		{pinnedFile{symlink, input.pin}, int64(len(data))},
		{pinnedFile{directory, input.pin}, int64(len(data))},
		{pinnedFile{filepath.Join(directory, "missing.test"), input.pin}, int64(len(data))},
	} {
		if _, _, err := readPinned(tc.input, tc.limit, true); err == nil || strings.Contains(err.Error(), directory) {
			t.Fatal("bad hash/nonregular/missing/oversize file accepted or path exposed")
		}
	}
}

func TestAllInputsPinnedBeforeUse(t *testing.T) {
	inputs := makeTestInputs(t)
	for _, change := range []func(*options){
		func(o *options) { o.floatModel.pin = strings.Repeat("0", 64) },
		func(o *options) { o.matchedModel.pin = strings.Repeat("0", 64) },
		func(o *options) { o.ptqModel.pin = strings.Repeat("0", 64) },
		func(o *options) { o.qatModel.pin = strings.Repeat("0", 64) },
		func(o *options) { o.fixture.pin = strings.Repeat("0", 64) },
	} {
		opts := inputs.opts
		change(&opts)
		var output bytes.Buffer
		if err := run(argsFor(opts), &output); err == nil || output.Len() != 0 {
			t.Fatal("pin mismatch accepted or emitted a partial report")
		}
	}
	// Unselected files are byte-hashed, not deserialized into resident models.
	opts := inputs.opts
	opts.qatModel = writeTestInput(t, t.TempDir(), "original-unselected.test", []byte("original synthetic unselected bytes"))
	var output bytes.Buffer
	if err := run(argsFor(opts), &output); err != nil {
		t.Fatal("unselected model was deserialized:", err)
	}
}

func TestPredictorAndChecksumWarmAllocations(t *testing.T) {
	inputs := makeTestInputs(t)
	for _, tc := range []struct {
		mode string
		data []byte
	}{{"float", inputs.float}, {"matched", inputs.matched}, {"ptq", inputs.ptq}, {"qat", inputs.qat}} {
		selected, err := loadPredictor(tc.data, tc.mode, inputs.opts.floatModel.pin)
		if err != nil {
			t.Fatal(err)
		}
		checksum := checksumOffset
		allocs := testing.AllocsPerRun(32, func() {
			prediction, err := selected.predict("An original synthetic warm-loop test asks for an update.")
			if err != nil {
				t.Fatal(err)
			}
			checksum = predictionChecksum(checksum, &prediction)
		})
		if allocs != 0 || checksum == checksumOffset {
			t.Fatalf("mode %s warm-loop allocations=%v or unconsumed results", tc.mode, allocs)
		}
	}
}

func TestPredictionChecksumConsumesNumericAndDecisionFields(t *testing.T) {
	var heads [statehintclaims.HeadCount]statehintclaims.HeadPrediction
	baseline := predictionChecksum(checksumOffset, &heads)
	changes := []func(*[statehintclaims.HeadCount]statehintclaims.HeadPrediction){
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) { h[2].Probabilities[1] = 0.5 },
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) { h[1].Confidence = 0.9 },
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) { h[0].Margin = 0.05 },
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) { h[0].Winner = statehintclaims.True },
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) {
			h[0].State = statehintclaims.Unknown
		},
		func(h *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) {
			h[0].UnknownReason = "original_test_reason"
		},
	}
	for _, change := range changes {
		modified := heads
		change(&modified)
		if predictionChecksum(checksumOffset, &modified) == baseline {
			t.Fatal("checksum failed to consume changed prediction output")
		}
	}
}

func TestPrivateProfileCreationAndExclusiveOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	opts := options{cpuProfile: true, heapProfile: true, profileOutput: "original-test-run"}
	files, err := createProfiles(opts)
	if err != nil {
		t.Fatal(err)
	}
	files.close()
	base := filepath.Join(".cache", "statehint-claims-trit-bench")
	child := filepath.Join(base, opts.profileOutput)
	for _, path := range []string{base, child} {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
			t.Fatal("profile directory is not private")
		}
	}
	for _, name := range []string{"cpu.pprof", "heap.pprof"} {
		info, err := os.Stat(filepath.Join(child, name))
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
			t.Fatal("profile file is not private")
		}
	}
	if _, err := createProfiles(opts); err == nil {
		t.Fatal("existing profile output reused")
	}
	for _, name := range []string{"", ".", "..", "../escaped", "nested/child", "/absolute", "a b", strings.Repeat("x", 65)} {
		opts.profileOutput = name
		if _, err := createProfiles(opts); err == nil {
			t.Fatalf("unsafe profile child name accepted: %q", name)
		}
	}
}

func TestProfileBaseRejectsSymlinkAndPublicDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Symlink(t.TempDir(), ".cache"); err != nil {
		t.Fatal(err)
	}
	opts := options{heapProfile: true, profileOutput: "original-test-run"}
	if _, err := createProfiles(opts); err == nil {
		t.Fatal("symlink profile base accepted")
	}
	if err := os.Remove(".cache"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(".cache", 0700); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(".cache", "statehint-claims-trit-bench")
	if err := os.Mkdir(base, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := createProfiles(opts); err == nil {
		t.Fatal("public profile base accepted")
	}
}

func TestRunWritesOnlyFreshPrivateProfiles(t *testing.T) {
	inputs := makeTestInputs(t)
	t.Chdir(t.TempDir())
	args := append(argsFor(inputs.opts), "--cpu-profile", "--heap-profile", "--profile-out", "original-test-run")
	var output bytes.Buffer
	if err := run(args, &output); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cpu.pprof", "heap.pprof"} {
		path := filepath.Join(".cache", "statehint-claims-trit-bench", "original-test-run", name)
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 || info.Mode().Perm() != 0600 {
			t.Fatal("private profile was not written")
		}
	}
	if strings.Contains(output.String(), "original-test-run") || strings.Contains(output.String(), ".cache") {
		t.Fatal("profile path appeared in report")
	}
	output.Reset()
	if err := run(args, &output); err == nil || output.Len() != 0 {
		t.Fatal("profile output was overwritten or a partial report emitted")
	}
}
