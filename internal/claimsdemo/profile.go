// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package claimsdemo

import (
	"errors"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

const maxProfileBytes = 16 << 20

// Profiling is explicitly local and optional. New 0600 files prevent accidental
// overwrites. No HTTP profiling endpoints or input-derived labels are registered.
type localProfiles struct {
	cpu, heap  *profileFile
	cpuStarted bool
}

type profileFile struct {
	file      *os.File
	remaining int
	err       error
}

func (f *profileFile) Write(data []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	length := len(data)
	if length > f.remaining {
		data = data[:f.remaining]
	}
	n, err := f.file.Write(data)
	f.remaining -= n
	if err == nil && n != length {
		err = io.ErrShortWrite
	}
	f.err = err
	return n, err
}

func openProfile(path string) (*profileFile, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, errors.New("demo: cannot create profile; use a new local file in an existing directory")
	}
	return &profileFile{file: f, remaining: maxProfileBytes}, nil
}

func startProfiles(cpuPath, heapPath string) (*localProfiles, error) {
	p := &localProfiles{}
	var err error
	if p.cpu, err = openProfile(cpuPath); err != nil {
		return nil, err
	}
	if p.heap, err = openProfile(heapPath); err != nil {
		if p.cpu != nil {
			p.cpu.file.Close()
		}
		return nil, err
	}
	if p.cpu != nil {
		if err := pprof.StartCPUProfile(p.cpu); err != nil {
			p.cpu.file.Close()
			if p.heap != nil {
				p.heap.file.Close()
			}
			return nil, errors.New("demo: cannot start local Go CPU profiling")
		}
		p.cpuStarted = true
	}
	return p, nil
}

func (p *localProfiles) finish(keepAlive any) error {
	failed := false
	if p.cpuStarted {
		pprof.StopCPUProfile()
	}
	if p.cpu != nil {
		if p.cpu.err != nil {
			failed = true
		}
		if p.cpu.file.Close() != nil {
			failed = true
		}
	}
	if p.heap != nil {
		// The server/model remain reachable during this post-GC live-heap snapshot.
		runtime.GC()
		if pprof.WriteHeapProfile(p.heap) != nil {
			failed = true
		}
		if p.heap.file.Close() != nil {
			failed = true
		}
	}
	runtime.KeepAlive(keepAlive)
	if failed {
		return errors.New("demo: local profile could not be completed within its 16 MiB limit")
	}
	return nil
}
