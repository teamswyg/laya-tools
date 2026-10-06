// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Maintainer-only inference costs. No training, accuracy score or external IO.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/teamswyg/laya-tools/pkg/statehint"
)

var fixtures = [...]string{
	"The original fixture parser branch is implemented; its required boundary cases are still being added.",
	"가상의 개발 댓글 분류기를 준비했습니다. 같은 기능의 입력 길이 검사는 아직 작업 중입니다.",
	"Please choose whether the fictional settings panel should preserve the previous selection after reset.",
	"이번 가상 변환기의 양방향 처리와 왕복 검사를 모두 마쳤고, 이 변환기 안에 남은 작업은 없습니다.",
}

type report struct {
	Schema          string  `json:"schema"`
	ModelSHA        string  `json:"model_sha256"`
	ArtifactBytes   int     `json:"artifact_bytes"`
	TrainingSteps   uint64  `json:"training_steps"`
	Fixtures        int     `json:"original_fixtures"`
	Iterations      int     `json:"measured_predictions"`
	LoadNS          int64   `json:"load_and_validate_nanoseconds"`
	ElapsedNS       int64   `json:"warm_prediction_nanoseconds"`
	NSPerPrediction float64 `json:"nanoseconds_per_prediction"`
	AllocatedBytes  uint64  `json:"go_allocated_bytes_during_loop"`
	Allocations     uint64  `json:"go_allocations_during_loop"`
	CPUProfile      bool    `json:"go_cpu_profile_written"`
	HeapProfile     bool    `json:"go_heap_profile_written"`
	Accuracy        bool    `json:"accuracy_measured"`
	GPU             bool    `json:"gpu_measured"`
	Fit             bool    `json:"training_executed"`
	Mutation        bool    `json:"operational_mutation"`
}

func main() {
	runtime.GOMAXPROCS(2)
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "statehint benchmark failed; check local model pin and fresh output directory")
		os.Exit(1)
	}
}

func run(args []string, out, errOut io.Writer) error {
	flags := flag.NewFlagSet("riido-statehint-bench", flag.ContinueOnError)
	flags.SetOutput(errOut)
	path := flags.String("model", "", "already-held trained version1 .rsh; never downloaded")
	pin := flags.String("model-sha256", "", "required exact lowercase SHA-256")
	iterations := flags.Int("iterations", 10000, "measured predictions, 4..1000000")
	profiles := flags.String("profiles-dir", "", "optional new local directory for CPU and heap profiles")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *path == "" || *iterations < len(fixtures) || *iterations > 1000000 || len(*pin) != 64 {
		return statehint.ErrInput
	}
	decoded, err := hex.DecodeString(*pin)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != *pin {
		return statehint.ErrInput
	}
	file, err := os.Open(*path)
	if err != nil {
		return statehint.ErrArtifact
	}
	data, readErr := io.ReadAll(io.LimitReader(file, statehint.ArtifactBytes+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil || len(data) != statehint.ArtifactBytes {
		return statehint.ErrArtifact
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != *pin {
		return statehint.ErrArtifact
	}
	started := time.Now()
	model, err := statehint.Load(bytes.NewReader(data))
	loadNS := time.Since(started).Nanoseconds()
	if err != nil || model.TrainingSteps() == 0 {
		return statehint.ErrArtifact
	}
	var workspace statehint.Workspace
	for _, text := range fixtures {
		if _, err = model.Predict(text, &workspace); err != nil {
			return err
		}
	}
	var cpu *os.File
	if *profiles != "" {
		if err = os.Mkdir(*profiles, 0700); err != nil {
			return err
		}
		cpu, err = os.OpenFile(filepath.Join(*profiles, "cpu.pprof"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		if err = pprof.StartCPUProfile(cpu); err != nil {
			_ = cpu.Close()
			return err
		}
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started = time.Now()
	for i := 0; i < *iterations; i++ {
		if _, err = model.Predict(fixtures[i%len(fixtures)], &workspace); err != nil {
			break
		}
	}
	elapsed := time.Since(started).Nanoseconds()
	runtime.ReadMemStats(&after)
	if cpu != nil {
		pprof.StopCPUProfile()
		if closeErr = cpu.Close(); err == nil {
			err = closeErr
		}
	}
	if err != nil {
		return err
	}
	if *profiles != "" {
		heap, err := os.OpenFile(filepath.Join(*profiles, "heap.pprof"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		runtime.GC()
		err = pprof.WriteHeapProfile(heap)
		closeErr = heap.Close()
		if err != nil || closeErr != nil {
			return fmt.Errorf("heap profile write failed")
		}
	}
	runtime.KeepAlive(model)
	runtime.KeepAlive(&workspace)
	return json.NewEncoder(out).Encode(report{
		Schema: "riido-statehint-inference-costs-v1", ModelSHA: *pin,
		ArtifactBytes: len(data), TrainingSteps: model.TrainingSteps(), Fixtures: len(fixtures),
		Iterations: *iterations, LoadNS: loadNS, ElapsedNS: elapsed,
		NSPerPrediction: float64(elapsed) / float64(*iterations),
		AllocatedBytes:  after.TotalAlloc - before.TotalAlloc, Allocations: after.Mallocs - before.Mallocs,
		CPUProfile: cpu != nil, HeapProfile: *profiles != "",
	})
}
