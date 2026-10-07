// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// riido-statehint-claims-trit-bench measures one explicitly selected local
// three-claim model per process on pinned, original synthetic fit fixtures.
// It reports development measurements, not semantic quality or state authority.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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
	"unsafe"

	"github.com/teamswyg/laya-tools/pkg/statehintclaims"
	"github.com/teamswyg/laya-tools/pkg/statehintclaimtrit"
)

const (
	fixtureCount      = 16
	warmupCount       = 32
	defaultIterations = 20000
	maxIterations     = 1000000
	maxFixtureBytes   = 128 << 20
	maxFixtureRows    = 1000000
	maxFixtureLine    = 1 << 20
	selectionDomain   = "claims-trit-bench-v1:"
	checksumOffset    = uint64(14695981039346656037)
	checksumPrime     = uint64(1099511628211)
)

type pinnedFile struct {
	path string
	pin  string
}

type options struct {
	mode          string
	floatModel    pinnedFile
	matchedModel  pinnedFile
	ptqModel      pinnedFile
	qatModel      pinnedFile
	fixture       pinnedFile
	iterations    int
	cpuProfile    bool
	heapProfile   bool
	profileOutput string
}

type fixtureRow struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Locale string `json:"locale"`
	Split  string `json:"internal_split"`
	rank   [sha256.Size]byte
}

type request struct {
	text string
}

type fixtureSummary struct {
	Scope          string               `json:"scope"`
	InputRows      int                  `json:"input_rows"`
	SelectedCount  int                  `json:"selected_count"`
	Selection      string               `json:"selection"`
	SelectedSHA256 string               `json:"selected_sha256"`
	LocaleOrder    [fixtureCount]string `json:"locale_order"`
	TextByteCounts [fixtureCount]int    `json:"text_byte_counts"`
	MinTextBytes   int                  `json:"min_text_bytes"`
	MaxTextBytes   int                  `json:"max_text_bytes"`
	TotalTextBytes int                  `json:"total_text_bytes"`
}

type report struct {
	Schema                    string         `json:"schema"`
	DevelopmentOnly           bool           `json:"development_only"`
	Mode                      string         `json:"mode"`
	GoVersion                 string         `json:"go_version"`
	FloatSHA256               string         `json:"float_model_sha256"`
	MatchedSHA256             string         `json:"matched_model_sha256"`
	PTQSHA256                 string         `json:"ptq_model_sha256"`
	QATSHA256                 string         `json:"qat_model_sha256"`
	FixtureSHA256             string         `json:"fixture_sha256"`
	ParentSHA256              string         `json:"parent_sha256,omitempty"`
	TrainingSteps             uint64         `json:"model_training_steps"`
	Fixture                   fixtureSummary `json:"fixture"`
	ModelArtifactBytes        int            `json:"model_artifact_bytes"`
	ModelRuntimeFixedBytes    uint64         `json:"model_runtime_fixed_bytes"`
	ModelImmutableBytes       uint64         `json:"model_immutable_referenced_bytes"`
	ModelStorageBytes         uint64         `json:"model_storage_bytes"`
	WorkspaceFixedBytes       uint64         `json:"workspace_fixed_bytes"`
	RuntimeStorageScope       string         `json:"runtime_storage_scope"`
	LoadElapsedNS             int64          `json:"load_elapsed_ns"`
	LoadScope                 string         `json:"load_scope"`
	WarmupIterations          int            `json:"warmup_iterations"`
	Iterations                int            `json:"iterations"`
	ElapsedNS                 int64          `json:"elapsed_ns"`
	NSPerOp                   float64        `json:"ns_per_op"`
	TimingScope               string         `json:"timing_scope"`
	GoHeapAllocBefore         uint64         `json:"go_heap_alloc_bytes_before"`
	GoHeapAllocAfter          uint64         `json:"go_heap_alloc_bytes_after"`
	GoHeapAllocDelta          int64          `json:"go_heap_alloc_bytes_delta"`
	GoTotalAllocDelta         uint64         `json:"go_total_alloc_bytes_delta"`
	GoMallocsDelta            uint64         `json:"go_mallocs_delta"`
	GoAllocatedBytesPerOp     float64        `json:"go_allocated_bytes_per_op"`
	GoMallocsPerOp            float64        `json:"go_mallocs_per_op"`
	GoAllocationScope         string         `json:"go_allocation_scope"`
	PredictionResultChecksum  string         `json:"prediction_result_checksum"`
	ProcessPeakRSSMeasurement string         `json:"process_peak_rss_measurement"`
	CPUProfile                bool           `json:"cpu_profile"`
	HeapProfile               bool           `json:"heap_profile"`
}

// The closure owns exactly one model pointer and one caller-owned workspace.
// No model conversion, training, inference cache, or shared lock is used here.
type predictor struct {
	predict        func(string) ([statehintclaims.HeadCount]statehintclaims.HeadPrediction, error)
	modelBytes     uint64
	immutableBytes uint64
	workspaceBytes uint64
	loadNS         int64
	parentSHA      string
	trainingSteps  uint64
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	opts, help, err := parseOptions(args)
	if err != nil {
		return err
	}
	if help {
		_, err := io.WriteString(output, "Required: --mode float|matched|ptq|qat --float-model FILE --float-sha256 SHA --matched-model FILE --matched-sha256 SHA --ptq-model FILE --ptq-sha256 SHA --qat-model FILE --qat-sha256 SHA --fixture FILE --fixture-sha256 SHA\nOptional: --iterations N (1..1000000; default 20000), --cpu-profile, --heap-profile, --profile-out NAME (fresh private child name).\nFixtures must contain only original synthetic internal_split=fit rows with id, text, and ko/en locale.\n")
		return err
	}
	result, err := benchmark(opts)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(output).Encode(result); err != nil {
		return errors.New("benchmark report write failed")
	}
	return nil
}

func parseOptions(args []string) (options, bool, error) {
	var opts options
	flags := flag.NewFlagSet("riido-statehint-claims-trit-bench", flag.ContinueOnError)
	// Parsing diagnostics can echo caller paths or other raw arguments. Keep
	// every error emitted by this command independent of caller-provided data.
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.mode, "mode", "", "one of float, matched, ptq, qat")
	flags.StringVar(&opts.floatModel.path, "float-model", "", "local float RSC model")
	flags.StringVar(&opts.floatModel.pin, "float-sha256", "", "float artifact byte SHA-256")
	flags.StringVar(&opts.matchedModel.path, "matched-model", "", "local matched optimization float RSC model")
	flags.StringVar(&opts.matchedModel.pin, "matched-sha256", "", "matched float artifact byte SHA-256")
	flags.StringVar(&opts.ptqModel.path, "ptq-model", "", "local PTQ RQT model")
	flags.StringVar(&opts.ptqModel.pin, "ptq-sha256", "", "PTQ artifact byte SHA-256")
	flags.StringVar(&opts.qatModel.path, "qat-model", "", "local QAT RQT model")
	flags.StringVar(&opts.qatModel.pin, "qat-sha256", "", "QAT artifact byte SHA-256")
	flags.StringVar(&opts.fixture.path, "fixture", "", "local original synthetic fit JSONL")
	flags.StringVar(&opts.fixture.pin, "fixture-sha256", "", "fixture file byte SHA-256")
	flags.IntVar(&opts.iterations, "iterations", defaultIterations, "fixed bounded prediction count")
	flags.BoolVar(&opts.cpuProfile, "cpu-profile", false, "write a private CPU profile")
	flags.BoolVar(&opts.heapProfile, "heap-profile", false, "write a private Go heap profile")
	flags.StringVar(&opts.profileOutput, "profile-out", "", "fresh private profile child name")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, true, nil
		}
		return opts, false, errors.New("invalid benchmark flags")
	}
	if flags.NArg() != 0 || (opts.mode != "float" && opts.mode != "matched" && opts.mode != "ptq" && opts.mode != "qat") {
		return opts, false, errors.New("benchmark requires one mode: float, matched, ptq, or qat")
	}
	if opts.iterations < 1 || opts.iterations > maxIterations {
		return opts, false, errors.New("iterations must be between 1 and 1000000")
	}
	for _, input := range []*pinnedFile{&opts.floatModel, &opts.matchedModel, &opts.ptqModel, &opts.qatModel, &opts.fixture} {
		if input.path == "" || strings.Contains(input.path, "://") {
			return opts, false, errors.New("all model and fixture inputs require explicit local files")
		}
		pin, err := hex.DecodeString(input.pin)
		if err != nil || len(pin) != sha256.Size {
			return opts, false, errors.New("all model and fixture inputs require byte SHA-256 pins")
		}
		input.pin = hex.EncodeToString(pin)
	}
	if opts.cpuProfile || opts.heapProfile {
		if !validChildName(opts.profileOutput) {
			return opts, false, errors.New("profiles require a fresh private child name")
		}
	} else if opts.profileOutput != "" {
		return opts, false, errors.New("profile output requires CPU or heap profiling")
	}
	return opts, false, nil
}

// readPinned rejects symlinks and nonregular files, verifies the opened file
// identity, and checks byte SHA-256 before any artifact decoding or JSON parsing.
// Unselected model files are streamed into the hash without retaining bytes.
func readPinned(input pinnedFile, limit int64, keep bool) ([]byte, int, error) {
	info, err := os.Lstat(input.path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > limit {
		return nil, 0, errors.New("input must be a bounded local regular file")
	}
	file, err := os.Open(input.path)
	if err != nil {
		return nil, 0, errors.New("input file open failed")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, 0, errors.New("input file identity changed")
	}
	hash := sha256.New()
	var retained bytes.Buffer
	var writer io.Writer = hash
	if keep {
		writer = io.MultiWriter(hash, &retained)
	}
	n, err := io.Copy(writer, io.LimitReader(file, limit+1))
	if err != nil || n != info.Size() || n > limit {
		return nil, 0, errors.New("input file read failed or changed")
	}
	if hex.EncodeToString(hash.Sum(nil)) != input.pin {
		return nil, 0, errors.New("input SHA-256 mismatch")
	}
	if keep {
		return retained.Bytes(), int(n), nil
	}
	return nil, int(n), nil
}

func selectFixtures(data []byte) ([fixtureCount]request, fixtureSummary, error) {
	var requests [fixtureCount]request
	summary := fixtureSummary{Scope: "original_synthetic_fit_only", Selection: "first16_sha256_claims-trit-bench-v1_id", MinTextBytes: statehintclaims.MaxTextBytes}
	if !utf8.Valid(data) {
		return requests, summary, errors.New("fixture JSONL must be valid UTF-8")
	}
	rows := make([]fixtureRow, 0, fixtureCount)
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), maxFixtureLine)
	for scanner.Scan() {
		var row fixtureRow
		// Ignore every annotation/target/provenance field. Only these four
		// explicitly tagged fields participate in fixture validity or selection.
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return requests, summary, errors.New("invalid fixture JSONL row")
		}
		if row.Split != "fit" {
			return requests, summary, errors.New("fixture input must contain only internal_split=fit rows")
		}
		if row.ID == "" || len(row.ID) > 1024 || !utf8.ValidString(row.ID) ||
			row.Text == "" || len(row.Text) > statehintclaims.MaxTextBytes || !utf8.ValidString(row.Text) ||
			(row.Locale != "ko" && row.Locale != "en") {
			return requests, summary, errors.New("invalid fixture id, text, or ko/en locale")
		}
		if _, exists := seen[row.ID]; exists {
			return requests, summary, errors.New("fixture IDs must be unique")
		}
		seen[row.ID] = struct{}{}
		row.rank = sha256.Sum256([]byte(selectionDomain + row.ID))
		rows = append(rows, row)
		if len(rows) > maxFixtureRows {
			return requests, summary, errors.New("fixture row limit exceeded")
		}
	}
	if scanner.Err() != nil || len(rows) < fixtureCount {
		return requests, summary, errors.New("fixture input needs at least 16 bounded fit rows")
	}
	sort.Slice(rows, func(i, j int) bool {
		order := bytes.Compare(rows[i].rank[:], rows[j].rank[:])
		if order == 0 {
			return rows[i].ID < rows[j].ID
		}
		return order < 0
	})
	summary.InputRows, summary.SelectedCount = len(rows), fixtureCount
	hash := sha256.New()
	var length [8]byte
	for i, row := range rows[:fixtureCount] {
		requests[i] = request{text: row.Text}
		summary.LocaleOrder[i] = row.Locale
		summary.TextByteCounts[i] = len(row.Text)
		summary.MinTextBytes = min(summary.MinTextBytes, len(row.Text))
		summary.MaxTextBytes = max(summary.MaxTextBytes, len(row.Text))
		summary.TotalTextBytes += len(row.Text)
		for _, value := range []string{row.ID, row.Text, row.Locale, row.Split} {
			binary.LittleEndian.PutUint64(length[:], uint64(len(value)))
			hash.Write(length[:])
			io.WriteString(hash, value)
		}
	}
	summary.SelectedSHA256 = hex.EncodeToString(hash.Sum(nil))
	return requests, summary, nil
}

func loadPredictor(data []byte, mode, floatSHA string) (predictor, error) {
	var selected predictor
	start := time.Now()
	if mode == "float" || mode == "matched" {
		model, err := statehintclaims.Load(bytes.NewReader(data))
		if err != nil {
			return selected, errors.New("selected float model load failed")
		}
		selected.loadNS = time.Since(start).Nanoseconds()
		selected.trainingSteps = model.TrainingSteps()
		workspace := new(statehintclaims.Workspace)
		selected.modelBytes, selected.workspaceBytes = uint64(unsafe.Sizeof(*model)), uint64(unsafe.Sizeof(*workspace))
		selected.predict = func(text string) ([statehintclaims.HeadCount]statehintclaims.HeadPrediction, error) {
			prediction, err := model.Predict(text, workspace)
			return prediction.Heads, err
		}
		return selected, nil
	}
	model, err := statehintclaimtrit.Load(bytes.NewReader(data))
	if err != nil {
		return selected, errors.New("selected ternary model load failed")
	}
	selected.loadNS = time.Since(start).Nanoseconds()
	metadata := model.Metadata()
	if metadata.Mode != mode {
		return selected, errors.New("selected ternary artifact mode does not match requested mode")
	}
	if metadata.ParentSHA256 != floatSHA {
		return selected, errors.New("selected ternary artifact parent does not match pinned float bytes")
	}
	selected.parentSHA = metadata.ParentSHA256
	selected.trainingSteps = metadata.TrainingSteps
	selected.immutableBytes = uint64(len(metadata.ParentSHA256))
	workspace := new(statehintclaimtrit.Workspace)
	selected.modelBytes, selected.workspaceBytes = uint64(unsafe.Sizeof(*model)), uint64(unsafe.Sizeof(*workspace))
	selected.predict = func(text string) ([statehintclaims.HeadCount]statehintclaims.HeadPrediction, error) {
		prediction, err := model.Predict(text, workspace)
		return prediction.Heads, err
	}
	return selected, nil
}

func benchmark(opts options) (report, error) {
	var result report
	var selectedData []byte
	var selectedBytes int
	inputs := []struct {
		mode  string
		input pinnedFile
		limit int64
	}{
		{"float", opts.floatModel, statehintclaims.ArtifactBytes},
		{"matched", opts.matchedModel, statehintclaims.ArtifactBytes},
		{"ptq", opts.ptqModel, statehintclaimtrit.ArtifactBytes},
		{"qat", opts.qatModel, statehintclaimtrit.ArtifactBytes},
	}
	for _, input := range inputs {
		data, size, err := readPinned(input.input, input.limit, input.mode == opts.mode)
		if err != nil {
			return result, err
		}
		if input.mode == opts.mode {
			selectedData, selectedBytes = data, size
		}
	}
	fixtureData, _, err := readPinned(opts.fixture, maxFixtureBytes, true)
	if err != nil {
		return result, err
	}
	requests, summary, err := selectFixtures(fixtureData)
	if err != nil {
		return result, err
	}
	selected, err := loadPredictor(selectedData, opts.mode, opts.floatModel.pin)
	if err != nil {
		return result, err
	}
	// The model is read-only. Drop artifact/corpus backing bytes before the
	// warm loop; only the 16 selected texts and owned workspace remain needed.
	selectedData, fixtureData = nil, nil
	profiles, err := createProfiles(opts)
	if err != nil {
		return result, err
	}
	defer profiles.close()
	runtime.GC()
	for i := 0; i < warmupCount; i++ {
		if _, err := selected.predict(requests[i%fixtureCount].text); err != nil {
			return result, errors.New("prediction warmup failed")
		}
	}
	cpuRunning := false
	defer func() {
		if cpuRunning {
			pprof.StopCPUProfile()
		}
	}()
	if profiles.cpu != nil {
		if err := pprof.StartCPUProfile(profiles.cpu); err != nil {
			return result, errors.New("CPU profile start failed")
		}
		cpuRunning = true
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	checksum := checksumOffset
	start := time.Now()
	for i := 0; i < opts.iterations; i++ {
		prediction, err := selected.predict(requests[i%fixtureCount].text)
		if err != nil {
			return result, errors.New("prediction loop failed")
		}
		checksum = predictionChecksum(checksum, &prediction)
	}
	elapsed := time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&after)
	if profiles.cpu != nil {
		pprof.StopCPUProfile()
		cpuRunning = false
	}
	if profiles.heap != nil {
		runtime.GC()
		if err := pprof.WriteHeapProfile(profiles.heap); err != nil {
			return result, errors.New("heap profile write failed")
		}
	}
	result = report{
		Schema: "statehint-claims-trit-bench-v1", DevelopmentOnly: true, Mode: opts.mode, GoVersion: runtime.Version(),
		FloatSHA256: opts.floatModel.pin, MatchedSHA256: opts.matchedModel.pin, PTQSHA256: opts.ptqModel.pin, QATSHA256: opts.qatModel.pin,
		FixtureSHA256: opts.fixture.pin, ParentSHA256: selected.parentSHA, TrainingSteps: selected.trainingSteps, Fixture: summary,
		ModelArtifactBytes: selectedBytes, ModelRuntimeFixedBytes: selected.modelBytes, WorkspaceFixedBytes: selected.workspaceBytes,
		ModelImmutableBytes: selected.immutableBytes, ModelStorageBytes: selected.modelBytes + selected.immutableBytes,
		RuntimeStorageScope: "fixed=unsafe.Sizeof(Model); immutable=parent SHA string backing; sum excludes workspace and other heap/process memory",
		LoadElapsedNS:       selected.loadNS, LoadScope: "selected artifact decoding only; file hashing and fixtures excluded",
		WarmupIterations: warmupCount, Iterations: opts.iterations, ElapsedNS: elapsed, NSPerOp: float64(elapsed) / float64(opts.iterations),
		TimingScope:       "Predict plus allocation-free prediction checksum; round-robin first16; warmup excluded",
		GoHeapAllocBefore: before.HeapAlloc, GoHeapAllocAfter: after.HeapAlloc, GoHeapAllocDelta: int64(after.HeapAlloc) - int64(before.HeapAlloc),
		GoTotalAllocDelta: after.TotalAlloc - before.TotalAlloc, GoMallocsDelta: after.Mallocs - before.Mallocs,
		GoAllocatedBytesPerOp:    float64(after.TotalAlloc-before.TotalAlloc) / float64(opts.iterations),
		GoMallocsPerOp:           float64(after.Mallocs-before.Mallocs) / float64(opts.iterations),
		GoAllocationScope:        "process Go counters around warm loop; includes enabled profiling and background activity",
		PredictionResultChecksum: fmt.Sprintf("%016x", checksum), ProcessPeakRSSMeasurement: "external wrapper required; Go heap is not process peak RSS",
		CPUProfile: opts.cpuProfile, HeapProfile: opts.heapProfile,
	}
	return result, nil
}

// Accumulate every head's numeric outputs and decisions without serializing or
// retaining predictions. The published value is a checksum, not raw outputs.
func predictionChecksum(checksum uint64, heads *[statehintclaims.HeadCount]statehintclaims.HeadPrediction) uint64 {
	for _, head := range heads {
		for _, value := range head.Probabilities {
			checksum = checksumWord(checksum, math.Float64bits(value))
		}
		checksum = checksumWord(checksum, math.Float64bits(head.Confidence))
		checksum = checksumWord(checksum, math.Float64bits(head.Margin))
		for _, value := range []string{string(head.Winner), string(head.State), head.UnknownReason} {
			checksum = checksumWord(checksum, uint64(len(value)))
			for i := 0; i < len(value); i++ {
				checksum = (checksum ^ uint64(value[i])) * checksumPrime
			}
		}
	}
	return checksum
}

func checksumWord(checksum, word uint64) uint64 {
	for i := 0; i < 8; i++ {
		checksum = (checksum ^ (word & 255)) * checksumPrime
		word >>= 8
	}
	return checksum
}

type profileFiles struct{ cpu, heap *os.File }

func (files profileFiles) close() {
	if files.cpu != nil {
		files.cpu.Close()
	}
	if files.heap != nil {
		files.heap.Close()
	}
}

func validChildName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, value := range name {
		if !(value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '-' || value == '_') {
			return false
		}
	}
	return true
}

func createProfiles(opts options) (profileFiles, error) {
	var files profileFiles
	if !opts.cpuProfile && !opts.heapProfile {
		return files, nil
	}
	if !validChildName(opts.profileOutput) {
		return files, errors.New("invalid private profile child name")
	}
	base := filepath.Join(".cache", "statehint-claims-trit-bench")
	for _, path := range []string{".cache", base} {
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(path, 0700); err != nil {
				return files, errors.New("private profile base creation failed")
			}
			continue
		}
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || (path == base && info.Mode().Perm()&0077 != 0) {
			return files, errors.New("profile base must be a private local directory")
		}
	}
	directory := filepath.Join(base, opts.profileOutput)
	if err := os.Mkdir(directory, 0700); err != nil {
		return files, errors.New("profile output must be a fresh private directory")
	}
	open := func(name string) (*os.File, error) {
		return os.OpenFile(filepath.Join(directory, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	}
	var err error
	if opts.cpuProfile {
		files.cpu, err = open("cpu.pprof")
		if err != nil {
			return files, errors.New("private CPU profile creation failed")
		}
	}
	if opts.heapProfile {
		files.heap, err = open("heap.pprof")
		if err != nil {
			files.close()
			return profileFiles{}, errors.New("private heap profile creation failed")
		}
	}
	return files, nil
}
