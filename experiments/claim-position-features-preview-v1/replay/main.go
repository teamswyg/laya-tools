// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Replay reports feature mechanics on the twelve approved original texts.
// It generates no semantic references and never loads or calls a model.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"unicode/utf8"
	"unsafe"

	preview "github.com/teamswyg/laya-tools/pkg/statehintpositionpreview"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

type Case struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	SHA256 string `json:"exact_utf8_sha256"`
}

type Pair struct {
	ID string `json:"id"`
	A  Case   `json:"a"`
	B  Case   `json:"b"`
}

type CaseResult struct {
	ID                    string `json:"id"`
	TextSHA256            string `json:"exact_utf8_sha256"`
	Bytes                 int    `json:"utf8_bytes"`
	Runes                 int    `json:"runes"`
	WordCount             int    `json:"actual_v2_word_count"`
	BaseActivatedBins     int    `json:"base_activated_bins_including_zero"`
	PositionActivatedBins int    `json:"position_activated_bins_including_zero"`
	BaseExactBitParity    bool   `json:"base_exact_float32_bit_parity"`
	BaseSparseOrderParity bool   `json:"base_exact_sparse_order_parity"`
	BaseSHA256            string `json:"base_dense_float32le_sha256"`
	PositionSHA256        string `json:"position_dense_float32le_sha256"`
	PreviewSHA256         string `json:"preview_dense_float32le_sha256"`
}

type PairResult struct {
	ID                  string     `json:"id"`
	A                   CaseResult `json:"a"`
	B                   CaseResult `json:"b"`
	BaseDenseEqual      bool       `json:"base_dense_exact_float32_bits_equal"`
	WordCountEqual      bool       `json:"actual_v2_word_count_equal"`
	PositionDenseEqual  bool       `json:"position_dense_exact_float32_bits_equal"`
	ChangedPositionBins int        `json:"changed_normalized_position_bins"`
}

type Results struct {
	Status              string            `json:"status"`
	FeatureSchema       string            `json:"feature_schema"`
	BaseSchema          string            `json:"base_schema"`
	GoVersion           string            `json:"go_version"`
	GOOS                string            `json:"goos"`
	GOARCH              string            `json:"goarch"`
	GOMAXPROCS          int               `json:"gomaxprocs"`
	CasesSHA256         string            `json:"public_cases_sha256"`
	DesignSHA256        string            `json:"pre_extraction_design_sha256"`
	SourceSHA256        map[string]string `json:"public_source_sha256"`
	Pairs               []PairResult      `json:"pairs"`
	Aggregate           map[string]int    `json:"aggregate"`
	BaseWorkspaceBytes  uintptr           `json:"v2_workspace_sizeof_bytes"`
	WorkspaceBytes      uintptr           `json:"preview_workspace_sizeof_bytes"`
	BaseLinearBytes     int               `json:"theoretical_9x2048x4_linear_weight_bytes"`
	PreviewLinearBytes  int               `json:"theoretical_9x3072x4_linear_weight_bytes"`
	MeasurementBoundary string            `json:"measurement_boundary"`
}

func fileDigest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return digest(data), nil
}

func digest(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func denseDigest(bits []uint32) string {
	var bytes [preview.FeatureBins * 4]byte
	for i, bit := range bits {
		binary.LittleEndian.PutUint32(bytes[i*4:], bit)
	}
	return digest(bytes[:len(bits)*4])
}

func extract(c Case) (CaseResult, [preview.FeatureBins]uint32, error) {
	var bits [preview.FeatureBins]uint32
	result := CaseResult{ID: c.ID, TextSHA256: c.SHA256, Bytes: len(c.Text), Runes: utf8.RuneCountInString(c.Text)}
	if digest([]byte(c.Text)) != c.SHA256 {
		return result, bits, fmt.Errorf("case %s original text digest mismatch", c.ID)
	}
	var baseWorkspace statehintwide.Workspace
	base, err := statehintwide.ExtractContextual(c.Text, &baseWorkspace)
	if err != nil {
		return result, bits, err
	}
	var workspace preview.Workspace
	view, err := preview.Extract(c.Text, &workspace)
	if err != nil {
		return result, bits, err
	}
	result.WordCount = view.WordCount()
	result.BaseActivatedBins = base.Len()
	result.PositionActivatedBins = view.Len() - base.Len()
	result.BaseExactBitParity = true
	result.BaseSparseOrderParity = true
	var baseline [preview.BaseFeatureBins]uint32
	for i := 0; i < base.Len(); i++ {
		f := base.At(i)
		baseline[f.Index] = math.Float32bits(f.Value)
		actual := view.At(i)
		if actual.Index != f.Index || math.Float32bits(actual.Value) != math.Float32bits(f.Value) {
			result.BaseSparseOrderParity = false
		}
	}
	for i := 0; i < view.Len(); i++ {
		f := view.At(i)
		bits[f.Index] = math.Float32bits(f.Value)
	}
	for i, expected := range baseline {
		if bits[i] != expected {
			result.BaseExactBitParity = false
		}
	}
	if !result.BaseExactBitParity || !result.BaseSparseOrderParity || result.WordCount != base.WordCount() {
		return result, bits, fmt.Errorf("case %s actual base contract drift", c.ID)
	}
	result.BaseSHA256 = denseDigest(bits[:preview.BaseFeatureBins])
	result.PositionSHA256 = denseDigest(bits[preview.BaseFeatureBins:])
	result.PreviewSHA256 = denseDigest(bits[:])
	return result, bits, nil
}

const replaySourceKey = "experiments/claim-position-features-preview-v1/replay/main.go"

// checkRecorded compares the reproducible feature evidence. Execution-host
// metadata and this reporting tool's historical source pin are not feature
// outcomes; the encoding-source pins and all feature receipts remain checked.
func checkRecorded(result Results, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var recorded Results
	if err := json.Unmarshal(raw, &recorded); err != nil {
		return err
	}
	recorded.GoVersion, recorded.GOOS, recorded.GOARCH, recorded.GOMAXPROCS = result.GoVersion, result.GOOS, result.GOARCH, result.GOMAXPROCS
	delete(recorded.SourceSHA256, replaySourceKey)
	delete(result.SourceSHA256, replaySourceKey)
	if !reflect.DeepEqual(result, recorded) {
		return fmt.Errorf("recorded feature evidence differs; no files were written")
	}
	return nil
}

func run(bundle, output string, check bool, stdout io.Writer) error {
	if check && output != "-" {
		return fmt.Errorf("--check cannot be combined with --output; no files were written")
	}
	casesBytes, err := os.ReadFile(filepath.Join(bundle, "cases.json"))
	if err != nil {
		return err
	}
	var packet struct {
		Pairs []Pair `json:"pairs"`
	}
	if err := json.Unmarshal(casesBytes, &packet); err != nil {
		return err
	}
	if len(packet.Pairs) != 6 {
		return fmt.Errorf("must retain exactly all six original pairs")
	}
	designBytes, err := os.ReadFile(filepath.Join(bundle, "DESIGN.lock.json"))
	if err != nil {
		return err
	}
	var lock struct {
		CasesSHA256  string            `json:"input_projection_sha256"`
		SourceSHA256 map[string]string `json:"source_sha256"`
	}
	if err := json.Unmarshal(designBytes, &lock); err != nil {
		return err
	}
	if digest(casesBytes) != lock.CasesSHA256 {
		return fmt.Errorf("public cases differ from pre-extraction design lock")
	}
	result := Results{
		Status:        "EXPERIMENTAL feature-only mechanics; no semantic accuracy or production replacement",
		FeatureSchema: preview.FeatureSchema, BaseSchema: statehintwide.ContextualFeatureSchema,
		GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, GOMAXPROCS: runtime.GOMAXPROCS(0),
		CasesSHA256: digest(casesBytes), DesignSHA256: digest(designBytes),
		SourceSHA256: make(map[string]string), Pairs: make([]PairResult, 0, 6),
		Aggregate:          map[string]int{"pairs_retained": 6, "texts": 12, "actual_base_bit_parity_texts": 0, "base_dense_equal_pairs": 0, "position_dense_different_pairs": 0, "position_dense_equal_pairs": 0, "model_calls": 0, "reference_generation_calls": 0},
		BaseWorkspaceBytes: unsafe.Sizeof(statehintwide.Workspace{}), WorkspaceBytes: unsafe.Sizeof(preview.Workspace{}),
		BaseLinearBytes: 9 * 2048 * 4, PreviewLinearBytes: 9 * 3072 * 4,
		MeasurementBoundary: "Workspace sizeof and theoretical 9-logit float32 linear weight payload only; not actual model payload, process RSS, GPU memory or accuracy. Changed normalized bins may include shared scale changes.",
	}
	for _, path := range []string{"pkg/statehintwide/features.go", "pkg/statehintwide/contextual_api.go", "pkg/statehintpositionpreview/features.go", "pkg/statehintpositionpreview/features_test.go", "pkg/statehintpositionpreview/features_bench_test.go", filepath.Join(bundle, "replay/main.go")} {
		hash, err := fileDigest(path)
		if err != nil {
			return err
		}
		key := path
		if path == filepath.Join(bundle, "replay/main.go") {
			key = replaySourceKey
		}
		result.SourceSHA256[key] = hash
		if expected, found := lock.SourceSHA256[key]; found && hash != expected {
			return fmt.Errorf("locked base source changed: %s", path)
		}
	}
	for _, pair := range packet.Pairs {
		a, bitsA, err := extract(pair.A)
		if err != nil {
			return err
		}
		b, bitsB, err := extract(pair.B)
		if err != nil {
			return err
		}
		p := PairResult{ID: pair.ID, A: a, B: b, BaseDenseEqual: true, WordCountEqual: a.WordCount == b.WordCount, PositionDenseEqual: true}
		for i := range bitsA {
			if bitsA[i] != bitsB[i] {
				if i < preview.BaseFeatureBins {
					p.BaseDenseEqual = false
				} else {
					p.PositionDenseEqual = false
					p.ChangedPositionBins++
				}
			}
		}
		result.Aggregate["actual_base_bit_parity_texts"] += 2
		if p.BaseDenseEqual {
			result.Aggregate["base_dense_equal_pairs"]++
		}
		if p.PositionDenseEqual {
			result.Aggregate["position_dense_equal_pairs"]++
		} else {
			result.Aggregate["position_dense_different_pairs"]++
		}
		result.Pairs = append(result.Pairs, p)
	}
	if check {
		return checkRecorded(result, filepath.Join(bundle, "results.json"))
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if output == "-" {
		_, err := stdout.Write(encoded)
		return err
	}
	return os.WriteFile(output, encoded, 0644)
}

func runCLI(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("position-feature-replay", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	bundle := flags.String("bundle", "experiments/claim-position-features-preview-v1", "public bundle directory; run from repository root")
	output := flags.String("output", "-", "feature-only result JSON; - writes stdout, a path opts in to writing a file")
	check := flags.Bool("check", false, "compare recorded bundle/results.json without writing; ignores execution-host metadata and replay tool source pin")
	if err := flags.Parse(args); err != nil {
		return err
	}
	return run(*bundle, *output, *check, stdout)
}

func main() {
	if err := runCLI(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
