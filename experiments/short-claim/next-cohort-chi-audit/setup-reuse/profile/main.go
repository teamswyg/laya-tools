// SPDX-License-Identifier: Apache-2.0
// Maintainer profiling only. These loops are separate from timing comparisons.
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
	"runtime"
	"runtime/debug"
	"runtime/pprof"

	"github.com/teamswyg/laya-tools/pkg/hintprepared"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const inputPin = "39e0aadbfe91c0b15ad3da257725de8d74c7a82c9d4222f73da8f937e72bf4c2"
const modelPin = "dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004"

func readPin(path string, size int, pin string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("profile_read")
	}
	b, err := io.ReadAll(io.LimitReader(f, int64(size+1)))
	c := f.Close()
	h := sha256.Sum256(b)
	if err != nil || c != nil || len(b) != size || hex.EncodeToString(h[:]) != pin {
		return nil, errors.New("profile_pin")
	}
	return b, nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error()) // fixed diagnostics, never paths
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("setup-reuse-profile", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	in := fs.String("input", "", "pinned public input")
	model := fs.String("model", "", "explicit existing inactive model")
	out := fs.String("out", "", "new private raw profile path")
	kind := fs.String("kind", "", "cpu or alloc; separate processes")
	phase := fs.String("phase", "", "prepare or rank")
	n := fs.Int("iterations", 0, "explicit bounded repetitions")
	if fs.Parse(os.Args[1:]) != nil || fs.NArg() != 0 || *in == "" || *model == "" || *out == "" || (*kind != "cpu" && *kind != "alloc") || (*phase != "prepare" && *phase != "rank") || *n < 1 || *n > 262144 || (*phase == "prepare" && *n > 512) {
		return errors.New("profile_arguments")
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(96 << 20) // Go GC target, not a hard RSS limit
	if *kind == "alloc" {
		runtime.MemProfileRate = 1
	}
	b, err := readPin(*in, 1417, inputPin)
	if err != nil {
		return err
	}
	current, err := shortclaim.LoadValidated(bytes.NewReader(b))
	if err != nil {
		return errors.New("profile_input")
	}
	raw, err := readPin(*model, 32792, modelPin)
	if err != nil {
		return err
	}
	v, err := hintweights.New(raw)
	if err != nil {
		return errors.New("profile_view")
	}
	p, err := hintprepared.Prepare(current.Prepared())
	if err != nil {
		return errors.New("profile_prepare")
	}
	anchor, err := p.Rank(current, v)
	if err != nil || anchor.Count != 5 || anchor.Kind != hintprepared.Kind || anchor.FallbackReason != "" {
		return errors.New("profile_anchor")
	}
	want := [5]uint64{0xbfc4dd509998bba7, 0xbfc40874752fd85e, 0xbfb69c0f88bf8dcb, 0xbfc45cd2f7e3a36f, 0xbfbc36f53a763f34}
	order := [8]int{2, 4, 1, 3, 0}
	if anchor.Order != order {
		return errors.New("profile_anchor_order")
	}
	for i, bits := range want {
		if math.Float64bits(anchor.Scores[i]) != bits {
			return errors.New("profile_anchor_bits")
		}
	}
	for i := 5; i < len(anchor.Scores); i++ {
		if math.Float64bits(anchor.Scores[i]) != 0 {
			return errors.New("profile_anchor_unused")
		}
	}
	f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.New("profile_output")
	}
	closed, stopped, cpuStarted := false, false, false
	defer func() {
		if cpuStarted && !stopped {
			pprof.StopCPUProfile()
		}
		if !closed {
			f.Close()
		}
	}()
	if *kind == "cpu" {
		if err := pprof.StartCPUProfile(f); err != nil {
			return errors.New("profile_cpu_start")
		}
		cpuStarted = true
	}
	prepareCalls, rankCalls := 1, 1 // initial setup and pinned anchor
	for i := 0; i < *n; i++ {
		if *phase == "prepare" {
			prepareCalls++
			p, err = hintprepared.Prepare(current.Prepared())
			if err != nil {
				return errors.New("profile_loop_prepare")
			}
		}
		rankCalls++
		r, e := p.Rank(current, v)
		if e != nil || r != anchor {
			return errors.New("profile_loop_rank")
		}
		for j, bits := range want {
			if math.Float64bits(r.Scores[j]) != bits {
				return errors.New("profile_loop_bits")
			}
		}
	}
	if *kind == "cpu" {
		pprof.StopCPUProfile()
		stopped = true
	} else {
		runtime.GC()
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			return errors.New("profile_alloc_write")
		}
	}
	if err := f.Sync(); err != nil {
		return errors.New("profile_sync")
	}
	err = f.Close()
	closed = true
	if err != nil {
		return errors.New("profile_close")
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Kind, Phase             string
		Iterations              int
		PrepareCalls, RankCalls int
		Parity                  bool
		AllocationIncludesSetup bool
		Scope                   string
	}{*kind, *phase, *n, prepareCalls, rankCalls, true, *kind == "alloc", "Private profiling with parity guards; not unprofiled timing, quality, verification savings, RSS/GPU, training or activation. Prepare phase includes Rank. Alloc profile includes process setup; no baseline subtraction."}); err != nil {
		return errors.New("profile_stdout")
	}
	return nil
}
