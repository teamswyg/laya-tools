package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"slices"
	"testing"
	"time"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

type stageMeasurement struct {
	Stage            string  `json:"stage"`
	Repetitions      int     `json:"repetitions"`
	P50NS            int64   `json:"p50_ns"`
	P95NS            int64   `json:"p95_ns"`
	AllocationsPerOp float64 `json:"go_allocations_per_op"`
}

func measureStage(name string, n int, f func() error) (stageMeasurement, error) {
	// Warm-up and allocation probes are disclosed separately from the n latency
	// samples. Every invocation does real work; there is no result cache.
	for i := 0; i < 20; i++ {
		if err := f(); err != nil {
			return stageMeasurement{}, err
		}
	}
	var allocationError error
	allocations := testing.AllocsPerRun(100, func() {
		if err := f(); err != nil {
			allocationError = err
		}
	})
	if allocationError != nil {
		return stageMeasurement{}, allocationError
	}
	times := make([]int64, n)
	for i := range times {
		start := time.Now()
		if err := f(); err != nil {
			return stageMeasurement{}, err
		}
		times[i] = time.Since(start).Nanoseconds()
	}
	slices.Sort(times)
	return stageMeasurement{Stage: name, Repetitions: n, P50NS: times[(n-1)/2], P95NS: times[(95*n+99)/100-1], AllocationsPerOp: allocations}, nil
}

func benchmark(in io.Reader, out io.Writer, kind string, repetitions int, profile string) error {
	if repetitions < 100 || repetitions > 10000 {
		return errors.New("benchmark_iteration_bounds")
	}
	raw, err := io.ReadAll(io.LimitReader(in, shortclaim.MaxJSONBytes+1))
	if err != nil {
		return errors.New("benchmark_read_failed")
	}
	p, err := shortclaim.Load(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	if p.Count != shortclaim.MaxCandidates {
		return errors.New("benchmark_requires_eight_candidates")
	}
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)
	if profile != "" {
		f, e := os.OpenFile(profile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if e != nil {
			return errors.New("profile_create_failed")
		}
		defer f.Close()
		if e = pprof.StartCPUProfile(f); e != nil {
			return errors.New("profile_start_failed")
		}
		defer pprof.StopCPUProfile()
	}
	start := time.Now()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	ranking, err := shortclaim.Rank(p, kind)
	if err != nil {
		return err
	}
	var sink shortclaim.Prepared
	stages := []struct {
		name string
		f    func() error
	}{
		{"parse_validate_normalize", func() error { var e error; sink, e = shortclaim.Load(bytes.NewReader(raw)); return e }},
		{"rank_validate_features_score_order", func() error { var e error; ranking, e = shortclaim.Rank(p, kind); return e }},
		{"input_digest", func() error { _ = digest(p); return nil }},
		{"ranking_serialization", func() error { return json.NewEncoder(io.Discard).Encode(ranking) }},
		{"full_go_request", func() error { return emit(bytes.NewReader(raw), io.Discard, kind) }},
	}
	measurements := make([]stageMeasurement, 0, len(stages))
	for _, stage := range stages {
		r, e := measureStage(stage.name, repetitions, stage.f)
		if e != nil {
			return e
		}
		measurements = append(measurements, r)
	}
	runtime.KeepAlive(sink)
	runtime.KeepAlive(ranking)
	runtime.ReadMemStats(&after)
	r := struct {
		Schema              string             `json:"schema"`
		Kind                string             `json:"baseline"`
		InputSHA256         string             `json:"canonical_input_sha256"`
		OS                  string             `json:"os"`
		Arch                string             `json:"arch"`
		Go                  string             `json:"go"`
		CandidateCount      int                `json:"candidate_count"`
		GOMAXPROCS          int                `json:"gomaxprocs"`
		GPUUsed             bool               `json:"gpu_used"`
		EncoderUsed         bool               `json:"encoder_used"`
		CacheHits           int                `json:"cache_hits"`
		CPUProfileEnabled   bool               `json:"cpu_profile_enabled"`
		WarmupsPerStage     int                `json:"warmups_per_stage"`
		AllocProbesPerStage int                `json:"allocation_probes_per_stage_including_warmup"`
		GoHeapAllocBytes    uint64             `json:"go_heap_alloc_bytes_at_end"`
		GoAllocBytes        uint64             `json:"go_total_alloc_bytes_delta"`
		WallSeconds         float64            `json:"measurement_wall_seconds"`
		Measurements        []stageMeasurement `json:"stages"`
		Scope               string             `json:"scope"`
	}{"riido-shortclaim-benchmark-v1", kind, digest(p), runtime.GOOS, runtime.GOARCH, runtime.Version(), p.Count, 1, false, false, 0, profile != "", 20, 101, after.HeapAlloc, after.TotalAlloc - before.TotalAlloc, time.Since(start).Seconds(), measurements, "Original repeated fixture; rank combines validation, features, scoring and ordering. HeapAlloc includes uncollected garbage, without forced GC, and is not OS peak RSS; timer/GC/serialization overhead is included. Optional profile covers warmups, allocation probes and timed stages; profile flush/close is outside measurement_wall_seconds. No learned model or quality/savings claim."}
	return json.NewEncoder(out).Encode(r)
}
