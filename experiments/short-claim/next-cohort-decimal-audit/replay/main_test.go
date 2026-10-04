// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// ReaderOnly prevents WriterTo from hiding a destination ReadFrom cap bypass.
func TestOutputBoundCopyPaths(t *testing.T) {
	for _, writerTo := range []bool{false, true} {
		for _, size := range []int{8, 9, 32768} {
			var src io.Reader = strings.NewReader(strings.Repeat("x", size))
			if !writerTo {
				src = struct{ io.Reader }{src}
			}
			dst := &bounded{max: 8}
			if _, ok := any(dst).(io.ReaderFrom); ok {
				t.Fatal("unexpected ReaderFrom")
			}
			n, err := io.Copy(dst, src)
			if n > 8 || len(dst.Bytes()) > 8 {
				t.Fatal("output cap bypassed")
			}
			if size == 8 {
				if err != nil || n != 8 || string(dst.Bytes()) != "xxxxxxxx" {
					t.Fatal("exact fit failed")
				}
			} else if err == nil {
				t.Fatal("oversize accepted")
			}
		}
	}
}

// Mutations of a saved exposed protocol are controls, not new Golden examples.
// Loading and validation here do not launch the original library or worker.
func TestSavedReportCorruptions(t *testing.T) {
	p, err := loadPacket("..")
	if err != nil {
		t.Fatal(err)
	}
	serialized, err := json.Marshal(p.Saved)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*Report)
	}{
		{"authority", func(r *Report) { r.RoleAssigned = true }},
		{"counts", func(r *Report) { r.Counts.APIAttempted-- }},
		{"identity", func(r *Report) { r.Trials[0].CandidateID = "truncate_checked" }},
		{"value", func(r *Report) { v := "2"; r.Trials[3].Result.Value = &v }},
		{"ledger_snapshot", func(r *Report) { v := "2"; r.Trials[0].Calls[2].Text = &v }},
		{"extra_trial", func(r *Report) { r.Trials = append(r.Trials, r.Trials[0]) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var r Report
			if decode(serialized, &r) != nil {
				t.Fatal("clone decode")
			}
			c.mutate(&r)
			if validate(r, p.Input, p.Wants) == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
}
