// Copyright 2026 teamswyg. Licensed under the Apache License, Version 2.0.
// Source-only owned collector; Root separately admits any future invocation.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	core "riido.local/next60gjsontwooutside/twoliteral"
	"runtime"
	"runtime/debug"
)

type Artifact struct {
	Name   string `json:"name"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
type OutsideResult struct {
	Schema                 string         `json:"schema"`
	State                  string         `json:"state"`
	Failure                string         `json:"failure"`
	ConfigSHA256           string         `json:"config_sha256"`
	PlanSHA256             string         `json:"plan_sha256"`
	WorkerSHA256           string         `json:"worker_sha256"`
	ControllerSHA256       string         `json:"controller_sha256"`
	Process                ProcessReceipt `json:"process"`
	DurablePrefix          DurableReceipt `json:"durable_prefix"`
	FramesACKWritten       int            `json:"frames_ack_written"`
	LastDurableCounts      *core.Counts   `json:"last_durable_counts"`
	LastACKWrittenCounts   *core.Counts   `json:"last_ack_written_counts"`
	TerminalRowsACKWritten int            `json:"terminal_rows_ack_written"`
	FinalACKWritten        bool           `json:"final_ack_written"`
	FinalChildFileIdentity bool           `json:"final_child_file_identity"`
	CompleteCountersKnown  bool           `json:"complete_counters_known"`
	CompleteRows           *int           `json:"complete_rows"`
	RemainingCallsUnknown  bool           `json:"remaining_calls_unknown"`
	InitializationCalls    *int           `json:"initialization_calls"`
	NestedOriginalCalls    *int           `json:"nested_original_calls"`
	Authority              *bool          `json:"source_rights_compiler_resource_authority"`
	TruthAssigned          int            `json:"truth_assigned"`
	LabelsAssigned         int            `json:"labels_assigned"`
	FitOrModelCalls        int            `json:"fit_or_model_calls"`
	Artifacts              []Artifact     `json:"artifacts"`
	OutputBytesRetained    *int64         `json:"output_bytes_retained"`
	AggregateCap           int            `json:"aggregate_cap_bytes"`
	NoAutomaticRetry       bool           `json:"no_automatic_retry"`
}

func fixedFailure(e error) string {
	if e == nil {
		return ""
	}
	v, ok := e.(failureCode)
	if !ok || len(v) < 1 || len(v) > 96 {
		return "controller_failed"
	}
	for _, c := range v {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return "controller_failed"
		}
	}
	return string(v)
}
func inventory(dir string) (n int64, a []Artifact, e error) {
	scan := func(root, prefix string, names []string, caps []int64) error {
		entries, x := os.ReadDir(root)
		if x != nil {
			return errCode("output_census_failed")
		}
		for _, v := range entries {
			if prefix == "" && v.Name() == "worker-attempt" {
				continue
			}
			at := -1
			for i, name := range names {
				if v.Name() == name {
					at = i
				}
			}
			if at < 0 {
				return errCode("unexpected_output_entry")
			}
			b, x := readOwnedOutput(filepath.Join(root, v.Name()), caps[at])
			if x != nil {
				return x
			}
			n += int64(len(b))
			a = append(a, Artifact{prefix + v.Name(), int64(len(b)), digest(b)})
		}
		return nil
	}
	if e = scan(dir, "", []string{"reservation.json", "journal.jsonl", "stderr.txt", "outside.json"}, []int64{16384, journalCap, stderrCap, 16384}); e != nil {
		return
	}
	child := filepath.Join(dir, "worker-attempt")
	st, x := os.Lstat(child)
	if os.IsNotExist(x) {
		return
	}
	if x != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 || st.Mode().Perm() != 0700 {
		e = errCode("child_output_not_private_directory")
		return
	}
	e = scan(child, "worker-attempt/", []string{"reservation.json", "results.json", "failure.json"}, []int64{16384, 65536, 16384})
	return
}
func downgrade(r OutsideResult, e error) OutsideResult {
	r.State = "incomplete_observation_transport"
	r.Failure = fixedFailure(e)
	r.CompleteCountersKnown = false
	r.CompleteRows = nil
	r.RemainingCallsUnknown = true
	r.OutputBytesRetained = nil
	return r
}
func persistResult(dir string, b []byte, r OutsideResult, save func(string, string, []byte, int) error) (OutsideResult, error) {
	if save == nil {
		e := errCode("outside_persistence_missing")
		return downgrade(r, e), e
	}
	if e := save(dir, "outside.json", b, 16384); e != nil {
		return downgrade(r, e), e
	}
	return r, nil
}
func collect(c Config, raw []byte) (r OutsideResult, e error) {
	r = OutsideResult{Schema: "riido-gjson-two-native-outside-result-v1", State: "not_started", ConfigSHA256: digest(raw), PlanSHA256: c.Plan.SHA256, WorkerSHA256: c.Worker.Pin.SHA256, ControllerSHA256: c.Controller.Pin.SHA256, RemainingCallsUnknown: true, AggregateCap: aggregateCap, NoAutomaticRetry: true}
	if e = validateConfig(c); e != nil {
		return
	}
	p, x := loadCountPlan(c)
	if x != nil {
		return r, x
	}
	carrier, x := loadCarrier(p)
	if x != nil {
		return r, x
	}
	if e = requireAbsentOutput(c.Output); e != nil {
		return
	}
	if os.Mkdir(c.Output, 0700) != nil {
		return r, errCode("exclusive_output_directory_failed")
	}
	if e = syncDir(filepath.Dir(c.Output)); e != nil {
		return
	}
	reserve := struct {
		Schema     string `json:"schema"`
		Config     string `json:"config_sha256"`
		Plan       string `json:"plan_sha256"`
		Worker     string `json:"worker_sha256"`
		Controller string `json:"controller_sha256"`
		Inputs     int    `json:"inputs"`
		Rows       int    `json:"rows"`
		Methods    int    `json:"methods"`
		Callbacks  int    `json:"callbacks"`
		Frames     int    `json:"frames"`
		Aggregate  int    `json:"aggregate_cap_bytes"`
		Retry      bool   `json:"automatic_retry"`
	}{"riido-gjson-two-native-outside-reservation-v1", digest(raw), c.Plan.SHA256, c.Worker.Pin.SHA256, c.Controller.Pin.SHA256, 11, 33, 61, 33, 384, aggregateCap, false}
	b, x := json.Marshal(reserve)
	if x != nil {
		return r, errCode("reservation_encode")
	}
	if e = durableFile(c.Output, "reservation.json", append(b, '\n'), 16384); e != nil {
		return
	}
	j, x := openJournal(c.Output, 384)
	if x != nil {
		return r, x
	}
	state := ProtocolState{Carrier: carrier}
	attempt := filepath.Join(c.Output, "worker-attempt")
	if e = requireAbsentOutput(attempt); e != nil {
		j.Close()
		return
	}
	args := []string{"--plan", c.Plan.Path, "--plan-sha256", c.Plan.SHA256, "--attempt-dir", attempt}
	process, stderr, runErr := runChild(c, args, func(in io.Reader, out io.Writer) error {
		return consumeProtocol(in, out, j, &state, c.Plan.SHA256, func() error { return checkChildReservation(attempt, p, c.Plan.SHA256) })
	})
	r.Process = process
	r.DurablePrefix = j.Receipt
	r.FramesACKWritten = state.FramesACKWritten
	r.LastDurableCounts = state.LastDurableCounts
	r.TerminalRowsACKWritten = state.NextDispatch
	r.FinalACKWritten = state.Final != nil && state.FramesACKWritten == j.Receipt.FramesSynced
	if state.FramesACKWritten > 0 {
		v := state.Counts
		r.LastACKWrittenCounts = &v
	}
	closeErr := j.Close()
	if x = durableFile(c.Output, "stderr.txt", stderr, stderrCap); x != nil && runErr == nil {
		runErr = x
	}
	if closeErr != nil && runErr == nil {
		runErr = closeErr
	}
	if runErr == nil {
		if x = exactFinalResult(filepath.Join(attempt, "results.json"), state); x != nil {
			runErr = x
		} else {
			r.FinalChildFileIdentity = true
		}
	}
	if x = validateConfig(c); x != nil && runErr == nil {
		runErr = x
	}
	if _, x = loadCountPlan(c); x != nil && runErr == nil {
		runErr = x
	}
	if runErr == nil && r.FinalACKWritten && process.StartAttempts == 1 && process.Started && process.WaitAttempts == 1 && process.Reaped != nil && *process.Reaped && process.ExitCode != nil && *process.ExitCode == 0 {
		r.State = "complete_observation_transport"
		r.CompleteCountersKnown = true
		r.RemainingCallsUnknown = false
		v := 33
		r.CompleteRows = &v
	} else {
		if runErr == nil {
			runErr = errCode("incomplete_ack_or_process")
		}
		r.State = "incomplete_observation_transport"
	}
	r.Failure = fixedFailure(runErr)
	current, a, x := inventory(c.Output)
	r.Artifacts = a
	if x != nil {
		return downgrade(r, x), x
	}
	retained := current
	r.OutputBytesRetained = &retained
	for i := 0; i < 6; i++ {
		b, x = json.Marshal(r)
		if x != nil {
			return downgrade(r, errCode("outside_encode")), errCode("outside_encode")
		}
		b = append(b, '\n')
		v := current + int64(len(b))
		if v == retained {
			break
		}
		retained = v
	}
	// One future captured stdout summary (<=16KiB) is also charged prospectively;
	// no wire-stdout duplicate, ACK file, cursor, time file or retry temp is made.
	if len(b) > 16384 || current+int64(len(b))+16384 > aggregateCap {
		return downgrade(r, errCode("aggregate_saved_output_cap")), errCode("aggregate_saved_output_cap")
	}
	r, x = persistResult(c.Output, b, r, durableFile)
	if x != nil {
		return r, x
	}
	return r, runErr
}
func run(args []string) int {
	if len(args) != 2 || args[0] != "--config" {
		fmt.Fprintln(os.Stderr, "usage_config_only")
		return 2
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(256 << 20)
	raw, e := readRegular(args[1], metadataCap)
	if e != nil {
		fmt.Fprintln(os.Stderr, "config_read_failed")
		return 1
	}
	var c Config
	if strictCanonical(raw, &c) != nil {
		fmt.Fprintln(os.Stderr, "config_decode_failed")
		return 1
	}
	r, e := collect(c, raw)
	if e != nil && r.Failure == "" {
		r = downgrade(r, e)
	}
	b, x := json.Marshal(r)
	if x != nil || len(b)+1 > 16384 {
		fmt.Fprintln(os.Stderr, "outside_summary_cap_or_encode")
		return 1
	}
	b = append(b, '\n')
	if n, w := os.Stdout.Write(b); w != nil || n != len(b) {
		return 1
	}
	if e != nil {
		return 1
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:])) }
