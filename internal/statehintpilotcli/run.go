// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Package statehintpilotcli runs bounded local synthetic shadow comparisons.
package statehintpilotcli

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"os"
	"runtime"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
	"github.com/teamswyg/laya-tools/pkg/statehintpilot"
	"github.com/teamswyg/laya-tools/pkg/statehintwide"
)

const maxCases = 2400
const maxInputBytes = 8 << 20
const maxArtifactBytes = 1 << 20

var errCases = errors.New("state-hint-pilot: invalid or oversized JSONL cases")
var errArtifact = errors.New("state-hint-pilot: invalid or unavailable local trained model")

type Artifact struct {
	SHA256        string  `json:"sha256"`
	Bytes         int     `json:"bytes"`
	Version       uint16  `json:"format_version"`
	TrainingSteps uint64  `json:"training_steps"`
	Temperature   float64 `json:"temperature"`
}
type Variant struct {
	Name            string                  `json:"name"`
	ProbabilityKind string                  `json:"probability_kind"`
	Artifact        *Artifact               `json:"artifact,omitempty"`
	ElapsedNS       int64                   `json:"elapsed_ns"`
	Report          statehintpilot.Report   `json:"report"`
	Diagnostics     *ProbabilityDiagnostics `json:"probability_diagnostics,omitempty"`
}
type Comparison struct {
	Schema              string    `json:"schema"`
	DatasetSHA256       string    `json:"dataset_sha256"`
	InputBytes          int       `json:"input_bytes"`
	Rows                int       `json:"rows"`
	ReaderKind          string    `json:"reader_kind"`
	DevelopmentOnly     bool      `json:"development_only"`
	TrainingPerformed   bool      `json:"training_performed"`
	MutationExecuted    bool      `json:"mutation_executed"`
	ConfidenceThreshold float64   `json:"confidence_threshold"`
	MarginThreshold     float64   `json:"margin_threshold"`
	SampledGoHeapBytes  uint64    `json:"sampled_go_heap_bytes"`
	Variants            []Variant `json:"variants"`
}

// Run never downloads a model, trains, chooses a winner, or calls a live reader.
// Reports are private by default when written to a file and never echo text.
func Run(args []string, out, errOut io.Writer) error {
	f := flag.NewFlagSet("riidolaya state-hint-pilot", flag.ContinueOnError)
	f.SetOutput(errOut)
	casesPath := f.String("cases", "", "original synthetic JSONL cases; up to 2400")
	modelPath := f.String("model", "", "required local trained .rsh model; no download")
	parentPath := f.String("parent", "", "optional retained parent model for the same cases")
	modelSHA := f.String("model-sha256", "", "optional required model checksum before any predictions")
	parentSHA := f.String("parent-sha256", "", "optional required parent checksum")
	casesSHA := f.String("cases-sha256", "", "optional required complete JSONL checksum")
	compareRules := f.Bool("compare-rules", false, "also evaluate two explicitly unlearned rule controls")
	outputPath := f.String("out", "", "new private JSON file; existing files are never overwritten")
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() != 0 || *casesPath == "" || *modelPath == "" {
		return errors.New("state-hint-pilot: --cases and --model are required")
	}
	if *parentSHA != "" && *parentPath == "" {
		return errors.New("state-hint-pilot: --parent-sha256 requires --parent")
	}
	cases, input, err := readCases(*casesPath)
	if err != nil {
		return err
	}
	sha := sha256.Sum256(input)
	if !matchesSHA(*casesSHA, hex.EncodeToString(sha[:])) {
		return errors.New("state-hint-pilot: cases checksum mismatch")
	}
	predict, artifact, err := loadPredictor(*modelPath)
	if err != nil {
		return err
	}
	if !matchesSHA(*modelSHA, artifact.SHA256) {
		return errors.New("state-hint-pilot: model checksum mismatch")
	}
	type candidate struct {
		name, kind string
		predict    statehintcatalog.PredictorFunc
		artifact   *Artifact
	}
	candidates := []candidate{{"model", "synthetic_temperature_scaled_softmax", predict, &artifact}}
	if *parentPath != "" {
		p, a, err := loadPredictor(*parentPath)
		if err != nil {
			return err
		}
		if !matchesSHA(*parentSHA, a.SHA256) {
			return errors.New("state-hint-pilot: parent checksum mismatch")
		}
		candidates = append(candidates, candidate{"parent", "synthetic_temperature_scaled_softmax", p, &a})
	}
	if *compareRules {
		candidates = append(candidates, candidate{"rules_original", "unlearned_one_hot_selection_scores", statehint.RuleBaseline, nil}, candidate{"rules_speechact", "unlearned_one_hot_selection_scores", statehint.RuleSpeechAct, nil})
	}
	// Reserve explicit output before any predictions; failure never overwrites
	// previous evidence. Stdout is the default for people and agents.
	var file *os.File
	if *outputPath != "" {
		file, err = os.OpenFile(*outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return errors.New("state-hint-pilot: output must be a new writable file")
		}
		defer file.Close()
		out = file
	}
	result := Comparison{Schema: "riido-statehint-synthetic-pilot-v1", DatasetSHA256: hex.EncodeToString(sha[:]), InputBytes: len(input), Rows: len(cases), ReaderKind: "synthetic_fixture_only", DevelopmentOnly: true, ConfidenceThreshold: .9, MarginThreshold: .05, Variants: make([]Variant, 0, len(candidates))}
	for _, c := range candidates {
		start := time.Now()
		report, err := statehintpilot.Evaluate(cases, c.predict)
		if err != nil {
			return err
		}
		variant := Variant{Name: c.name, ProbabilityKind: c.kind, Artifact: c.artifact, ElapsedNS: time.Since(start).Nanoseconds(), Report: report}
		if c.artifact != nil {
			diagnostic := probabilityDiagnostics(report)
			variant.Diagnostics = &diagnostic
		}
		result.Variants = append(result.Variants, variant)
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	result.SampledGoHeapBytes = memory.HeapAlloc
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return errors.New("state-hint-pilot: cannot write complete report")
	}
	if file != nil {
		if err := file.Sync(); err != nil {
			return errors.New("state-hint-pilot: cannot sync complete report")
		}
	}
	return nil
}

func matchesSHA(expected, actual string) bool {
	if expected == "" {
		return true
	}
	b, err := hex.DecodeString(expected)
	return err == nil && len(b) == sha256.Size && hex.EncodeToString(b) == actual
}

func readCases(path string) ([]statehintpilot.Case, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, errCases
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		return nil, nil, errCases
	}
	input, err := io.ReadAll(io.LimitReader(f, maxInputBytes+1))
	if err != nil || len(input) > maxInputBytes {
		return nil, nil, errCases
	}
	scanner := bufio.NewScanner(bytes.NewReader(input))
	scanner.Buffer(make([]byte, 1024), 64<<10)
	cases := make([]statehintpilot.Case, 0)
	for scanner.Scan() {
		if len(cases) == maxCases {
			return nil, nil, errCases
		}
		row, err := decodeCase(scanner.Bytes())
		if err != nil {
			return nil, nil, err
		}
		cases = append(cases, row)
	}
	if scanner.Err() != nil || len(cases) == 0 {
		return nil, nil, errCases
	}
	return cases, input, nil
}

// Token-level string decoding rejects duplicate/alias/unknown keys, nulls,
// nested values, invalid UTF-8 and trailing JSON instead of silently coercing.
func decodeCase(line []byte) (statehintpilot.Case, error) {
	var row statehintpilot.Case
	if !utf8.Valid(line) {
		return row, errCases
	}
	d := json.NewDecoder(bytes.NewReader(line))
	t, err := d.Token()
	if err != nil || t != json.Delim('{') {
		return row, errCases
	}
	seen := [5]bool{}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return row, errCases
		}
		var target *string
		var index int
		switch key {
		case "id":
			target = &row.ID
			index = 0
		case "family":
			target = &row.Family
			index = 1
		case "locale":
			target = &row.Locale
			index = 2
		case "text":
			target = &row.Text
			index = 3
		case "expected_intent":
			index = 4
		default:
			return row, errCases
		}
		if seen[index] {
			return row, errCases
		}
		seen[index] = true
		value, err := d.Token()
		if err != nil {
			return row, errCases
		}
		s, ok := value.(string)
		if !ok {
			return row, errCases
		}
		if index == 4 {
			row.Expected = statehint.Intent(s)
		} else {
			*target = s
		}
	}
	t, err = d.Token()
	if err != nil || t != json.Delim('}') {
		return row, errCases
	}
	if _, err = d.Token(); err != io.EOF {
		return row, errCases
	}
	for _, present := range seen {
		if !present {
			return row, errCases
		}
	}
	return row, nil
}

func loadPredictor(path string) (statehintcatalog.PredictorFunc, Artifact, error) {
	var artifact Artifact
	f, err := os.Open(path)
	if err != nil {
		return nil, artifact, errArtifact
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		return nil, artifact, errArtifact
	}
	b, err := io.ReadAll(io.LimitReader(f, maxArtifactBytes+1))
	if err != nil || len(b) < 8 || len(b) > maxArtifactBytes {
		return nil, artifact, errArtifact
	}
	hash := sha256.Sum256(b)
	artifact.SHA256, artifact.Bytes, artifact.Version = hex.EncodeToString(hash[:]), len(b), binary.LittleEndian.Uint16(b[4:6])
	switch artifact.Version {
	case 1:
		model, err := statehint.Load(bytes.NewReader(b))
		if err != nil || model.TrainingSteps() == 0 {
			return nil, Artifact{}, errArtifact
		}
		artifact.TrainingSteps, artifact.Temperature = model.TrainingSteps(), model.Temperature()
		var workspace statehint.Workspace
		return func(text string) (statehint.Prediction, error) { return model.Predict(text, &workspace) }, artifact, nil
	case 2:
		model, err := statehintwide.Load(bytes.NewReader(b))
		if err != nil || model.TrainingSteps() == 0 {
			return nil, Artifact{}, errArtifact
		}
		artifact.TrainingSteps, artifact.Temperature = model.TrainingSteps(), model.Temperature()
		var workspace statehintwide.Workspace
		return func(text string) (statehint.Prediction, error) { return model.Predict(text, &workspace) }, artifact, nil
	default:
		return nil, Artifact{}, errArtifact
	}
}
