// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
// Saved-file comparison only. Root separately admits this source and invocation.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	cmp "riido.local/next60gjsonsavedcomparison/compare"
	"runtime"
	"runtime/debug"
)

func run(a []string) int {
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	if len(a) != 6 {
		return 1
	}
	var v [3]string
	var seen [3]bool
	for i := 0; i < 6; i += 2 {
		at := -1
		switch a[i] {
		case "--config":
			at = 0
		case "--config-sha256":
			at = 1
		case "--out":
			at = 2
		}
		if at < 0 || seen[at] || a[i+1] == "" {
			return 1
		}
		seen[at] = true
		v[at] = a[i+1]
	}
	c, e := cmp.LoadConfig(v[0], v[1])
	if e != nil {
		return 1
	}
	out, e := cmp.ReserveExclusive(v[2])
	if e != nil {
		return 1
	}
	defer out.Close()
	r, e := cmp.Run(c)
	if e != nil {
		return 1
	}
	b, e := json.Marshal(r)
	if e != nil {
		return 1
	}
	b = append(b, '\n')
	if cmp.WriteReserved(v[2], out, b) != nil {
		return 1
	}
	summary := struct {
		State      string    `json:"state"`
		Rows       int       `json:"rows"`
		Predicates cmp.Tally `json:"predicate_counts"`
		RowCounts  cmp.Tally `json:"row_counts"`
		Bytes      int       `json:"bytes"`
		SHA        string    `json:"sha256"`
		Labels     int       `json:"labels_assigned"`
		ModelFit   int       `json:"model_Fit_calls"`
	}{r.State, 33, r.PredicateCounts, r.RowCounts, len(b), cmp.Digest(b), 0, 0}
	if json.NewEncoder(os.Stdout).Encode(summary) != nil {
		return 1
	}
	return 0
}
func main() {
	v := run(os.Args[1:])
	if v != 0 {
		fmt.Fprintln(os.Stderr, "saved_comparison_failed")
	}
	os.Exit(v)
}
