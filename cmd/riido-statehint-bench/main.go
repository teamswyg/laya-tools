// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Warm classifier profiling only; profiles remain local and are not publication assets.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/teamswyg/laya-tools/pkg/statehint"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func run() error {
	model := flag.String("model", "", "local generated .rsh")
	n := flag.Int("iterations", 500000, "warm calls,1..10000000")
	cpu := flag.String("cpu-profile", "", "optional new private pprof file")
	heap := flag.String("heap-profile", "", "optional new private Go heap pprof file")
	flag.Parse()
	if *model == "" || *n < 1 || *n > 10000000 {
		return fmt.Errorf("model/iteration bounds")
	}
	f, e := os.Open(*model)
	if e != nil {
		return e
	}
	m, e := statehint.Load(f)
	ce := f.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	text := "자료를 확인한 후 다음 작업을 시작하고 있습니다. FictionalProject-00 FictionalItem-00"
	var work statehint.Workspace
	for i := 0; i < 32; i++ {
		if _, e = m.Predict(text, &work); e != nil {
			return e
		}
	}
	if *cpu != "" {
		p, e := os.OpenFile(*cpu, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		if e = pprof.StartCPUProfile(p); e != nil {
			p.Close()
			return e
		}
		defer p.Close()
		defer pprof.StopCPUProfile()
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	var sink statehint.Prediction
	for i := 0; i < *n; i++ {
		sink, e = m.Predict(text, &work)
		if e != nil {
			return e
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(sink)
	if *heap != "" {
		p, e := os.OpenFile(*heap, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		runtime.GC()
		e = pprof.WriteHeapProfile(p)
		ce = p.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Schema            string  `json:"schema"`
		Calls             int     `json:"calls"`
		ElapsedNS         int64   `json:"elapsed_ns"`
		MeanNS            float64 `json:"mean_ns"`
		Mallocs           uint64  `json:"mallocs"`
		AllocBytes        uint64  `json:"allocated_bytes"`
		HeapAllocBytes    uint64  `json:"sampled_go_heap_alloc_bytes"`
		ProfilesLocalOnly bool    `json:"profiles_local_only"`
		Scope             string  `json:"scope"`
	}{"riido-statehint-warm-profile-v1", *n, elapsed.Nanoseconds(), float64(elapsed.Nanoseconds()) / float64(*n), after.Mallocs - before.Mallocs, after.TotalAlloc - before.TotalAlloc, after.HeapAlloc, true, "Repeated one fixed synthetic short input; full feature extraction and scoring. Profiles include profiling overhead; excludes model load,JSON,policy,startup,OS RSS and GPU memory. No throughput generalization."})
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, "riido-statehint-bench:", e)
		os.Exit(1)
	}
}
