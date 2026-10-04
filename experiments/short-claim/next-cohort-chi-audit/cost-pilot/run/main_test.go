// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Public saved traces seed synthetic protocol-tampering controls. No test calls
// Chi observe, a worker, View.New, Prepare, Rank, a model file, or a compiler.
func testPacket(t *testing.T) packet {
	t.Helper()
	root := filepath.Join("..", "..")
	var p packet
	for _, i := range []int{9, 10, 11, 14} {
		b, e := read(filepath.Join(root, pins[i].name), pins[i].size)
		if e != nil || sha(b) != pins[i].sha {
			t.Fatal("public test fixture unavailable")
		}
		p.Data[i] = b
	}
	raw, e := unpack(p.Data[11])
	if e != nil || decode(p.Data[9], &p.In) != nil || decode(p.Data[10], &p.Wants) != nil || decode(raw, &p.Saved) != nil || validate(p.Saved, p.In, p.Wants) != nil {
		t.Fatal("public fixture shape")
	}
	p.Expected, e = anchorRanks(p.Data[14], p.In)
	if e != nil {
		t.Fatal(e)
	}
	p.FreezeSHA = sha([]byte("synthetic freeze"))
	p.SourceSHA = sha([]byte("synthetic pins"))
	return p
}
func testEvents(phase string, model bool) []pilotAPIEvent {
	names := []string{"route_load", "wants_read", "score_read", "load_validated"}
	cats := []string{"io", "io", "io", "hint"}
	if model {
		names = append(names, "model_read", "view_new", "prepare")
		cats = append(cats, "io", "hint", "hint")
	}
	var es []pilotAPIEvent
	for i, n := range names {
		es = append(es, pilotAPIEvent{Category: cats[i], Phase: phase, Method: n, Attempted: true, Returned: true})
	}
	return es
}
func rankEvent(policy string) pilotAPIEvent {
	method := "baseline_rank"
	if policy == "model" {
		method = "model_rank"
	}
	return pilotAPIEvent{Category: "hint", Phase: "rank", Method: method, Attempted: true, Returned: true}
}
func recount(p *pilotPipeline) {
	var ts []trial
	p.OwnedVerificationMismatches = 0
	p.UnavailableChecks = 0
	for _, v := range p.Visits {
		for _, c := range v.Checks {
			ts = append(ts, c.Trial)
			if c.Verdict.State == "mismatch" {
				p.OwnedVerificationMismatches++
			}
			if c.Verdict.State == "unavailable" {
				p.UnavailableChecks++
			}
		}
	}
	p.ChiCounts = pilotChiCounts(ts)
	p.HintCounts, p.IOCounts, p.ModelRankAttempts, p.ModelRankReturns, _ = eventCounts(p.HintEvents)
}
func testPipeline(p packet, policy, mode string, block, pos int) pilotPipeline {
	s := p.Expected[policy]
	r := rankValue{Kind: s.Kind, Count: s.Count, Order: s.Order, FallbackReason: s.FallbackReason}
	for i, b := range s.ScoreBits {
		x, _ := strconv.ParseUint(b, 16, 64)
		r.Scores[i] = math.Float64frombits(x)
	}
	x := pilotPipeline{PlannedPolicy: policy, EffectivePolicy: policy, Block: block, Position: pos, State: "verified", Ranking: &r, RankSnapshot: &s, CandidateIDs: s.CurrentIDs, Cost: pilotCost{ElapsedNS: 1}}
	if mode == "cold" {
		x.HintEvents = testEvents("cold_setup", policy == "model")
	}
	x.HintEvents = append(x.HintEvents, rankEvent(policy))
	for _, ci := range r.Order[:5] {
		x.Visited[ci] = true
		v := pilotVisit{CandidateIndex: ci, CandidateID: s.CurrentIDs[ci], State: "verified", RemainingFixtureIDs: []string{}}
		for fi := 0; fi < 9; fi++ {
			tr := p.Saved.Trials[fi*5+ci]
			ver := pilotClassify(tr, p.In.Fixtures[fi], p.Wants.Fixtures[fi].Want, p.In.Candidates[ci])
			v.Checks = append(v.Checks, pilotCheck{fi, p.In.Fixtures[fi].ID, ver, tr})
			if ver.State != "pass" {
				v.State = ver.State
				for j := fi + 1; j < 9; j++ {
					v.RemainingFixtureIDs = append(v.RemainingFixtureIDs, p.In.Fixtures[j].ID)
				}
				break
			}
		}
		x.Visits = append(x.Visits, v)
		if v.State == "verified" {
			chosen := ci
			x.SelectedIndex = &chosen
			break
		}
	}
	recount(&x)
	return x
}
func testWire(p packet, mode, policy string, block int) pilotWire {
	x := pilotWire{Schema: "riido-chi-paired-cost-worker-v1", Mode: mode, Block: block, Status: "completed", RouteSHA: pins[9].sha, WantsSHA: pins[10].sha, ScoreSHA: pins[13].sha, ModelSHA: modelSHA, ModelRef: modelRef, Scope: "synthetic protocol control"}
	if mode != "cold" {
		x.SetupEvents = testEvents("setup", mode != "precheck")
		x.SetupCost = &pilotCost{ElapsedNS: 1}
	}
	switch mode {
	case "precheck":
		x.Precheck = &p.Saved
		for ci, c := range p.In.Candidates {
			v := pilotPreVerdict{CandidateID: c.ID, State: "verified"}
			for fi := 0; fi < 9; fi++ {
				v.Checks[fi] = pilotClassify(p.Saved.Trials[fi*5+ci], p.In.Fixtures[fi], p.Wants.Fixtures[fi].Want, c)
				if v.Checks[fi].State == "pass" {
					v.Passed++
				} else {
					v.State = "mismatch"
				}
			}
			x.PrecheckVerdicts = append(x.PrecheckVerdicts, v)
		}
	case "anchor":
		e := rankEvent("model")
		e.Phase = "anchor"
		x.SetupEvents = append(x.SetupEvents, e, pilotAPIEvent{Category: "hint", Phase: "anchor", Method: "baselines_anchor", Attempted: true, Returned: true})
		for _, n := range []string{"model", "fixed", "bm25", "lexical", "narrow"} {
			x.Anchor = append(x.Anchor, p.Expected[n])
		}
	case "cold":
		pos := -1
		for i, n := range orders[block] {
			if n == policy {
				pos = i
			}
		}
		x.Pipelines = []pilotPipeline{testPipeline(p, policy, mode, block, pos)}
	case "warm":
		for pos, n := range orders[block] {
			x.Pipelines = append(x.Pipelines, testPipeline(p, n, mode, block, pos))
		}
	}
	x.SetupHintCounts, x.SetupIOCounts, x.ModelRankAttempts, x.ModelRankReturns, _ = eventCounts(x.SetupEvents)
	for _, v := range x.Pipelines {
		x.ModelRankAttempts += v.ModelRankAttempts
		x.ModelRankReturns += v.ModelRankReturns
	}
	return x
}
func testZip(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	z := gzip.NewWriter(&out)
	if _, e = z.Write(b); e != nil {
		t.Fatal(e)
	}
	if z.Close() != nil {
		t.Fatal("gzip close")
	}
	return out.Bytes()
}
func copyPipeline(t *testing.T, p pilotPipeline) pilotPipeline {
	t.Helper()
	b, _ := json.Marshal(p)
	var q pilotPipeline
	if decode(b, &q) != nil {
		t.Fatal("copy shape")
	}
	return q
}

func TestOutputBoundCopyPaths(t *testing.T) {
	for _, writerTo := range []bool{false, true} {
		for _, n := range []int{8, 9, 32768} {
			var r io.Reader = strings.NewReader(strings.Repeat("x", n))
			if !writerTo {
				r = struct{ io.Reader }{r}
			}
			b := &bounded{max: 8}
			if _, ok := any(b).(io.ReaderFrom); ok {
				t.Fatal("ReaderFrom bypass")
			}
			got, e := io.Copy(b, r)
			if got > 8 || b.buf.Len() > 8 || (n == 8 && (e != nil || got != 8)) || (n > 8 && e == nil) {
				t.Fatal("output cap")
			}
		}
	}
}
func TestStrictJSONAndGzip(t *testing.T) {
	type example struct {
		Zero bool   `json:"zero"`
		A    [2]int `json:"a"`
	}
	for _, s := range []string{`{"zero":false,"a":[0,0],"extra":0}`, `{"a":[0,0]}`, `{"zero":null,"a":[0,0]}`, `{"zero":false,"zero":false,"a":[0,0]}`, `{"zero":false,"a":[0]}`, `{"zero":false,"a":[0,0,0]}`, `{"zero":false,"a":[0,0]} {}`} {
		var v example
		if decode([]byte(s), &v) == nil {
			t.Fatal("non-strict JSON accepted")
		}
	}
	b := testZip(t, map[string]int{"x": 1})
	bad := append([]byte(nil), b...)
	bad[len(bad)-8] ^= 1
	for _, raw := range [][]byte{b[:len(b)-1], bad, append(append([]byte(nil), b...), b...), append(append([]byte(nil), b...), 0)} {
		if _, e := unpack(raw); e == nil {
			t.Fatal("gzip integrity accepted")
		}
	}
	if _, e := unpackLimit(b, 1); e == nil {
		t.Fatal("raw bound")
	}
}
func TestFreezeRequiredZerosAndLimits(t *testing.T) {
	var v map[string]any
	if decode([]byte(freezeTemplate), &v) != nil {
		t.Fatal("template")
	}
	v["source_pins"] = map[string]any{"bytes": 1, "sha256": strings.Repeat("0", 64)}
	b, _ := json.Marshal(v)
	if _, e := checkFreeze(b); e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"labels", "default_activation", "raw_cap_bytes"} {
		var q map[string]any
		_ = decode(b, &q)
		delete(q, key)
		raw, _ := json.Marshal(q)
		if _, e := checkFreeze(raw); e == nil {
			t.Fatal("missing freeze key")
		}
	}
	v["worker_timeout_seconds"] = 21
	raw, _ := json.Marshal(v)
	if _, e := checkFreeze(raw); e == nil {
		t.Fatal("changed timeout")
	}
}
func TestCompleteTrialAndHintIntegrity(t *testing.T) {
	p := testPacket(t)
	base := testPipeline(p, "model", "cold", 0, 2)
	check := func(x pilotPipeline) error {
		return checkPipeline(x, 0, 2, "model", "cold", p.In, p.Wants, p.Saved, p.Expected, false)
	}
	if e := check(base); e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"ledger_deleted", "ledger_value", "handler_before", "hint_deleted", "rank_category"} {
		x := copyPipeline(t, base)
		tr := &x.Visits[0].Checks[0].Trial
		switch kind {
		case "ledger_deleted":
			tr.Ledger = nil
		case "ledger_value":
			v := "altered captured return"
			tr.Ledger[0].Value = &v
		case "handler_before":
			v := "altered before parameter"
			tr.Handlers[0].Before.Chi = &v
		case "hint_deleted":
			x.HintEvents = x.HintEvents[:len(x.HintEvents)-1]
		case "rank_category":
			x.HintEvents[len(x.HintEvents)-1].Category = "io"
		}
		recount(&x)
		if check(x) == nil {
			t.Fatalf("accepted tampering: %s", kind)
		}
	}
}
func TestWarmPreparationFallbackAndTerminalIO(t *testing.T) {
	p := testPacket(t)
	x := testWire(p, "warm", "", 0)
	code := "returned_error"
	x.SetupEvents[len(x.SetupEvents)-1].Error = true
	x.SetupEvents[len(x.SetupEvents)-1].ErrorCode = &code
	pos := 2
	y := testPipeline(p, "bm25", "warm", 0, pos)
	y.PlannedPolicy = "model"
	reason := "model_setup_error"
	y.FallbackReason = &reason
	failed := pilotSnap(rankValue{}, y.CandidateIDs)
	y.FailedRanking = &failed
	y.HintEvents = []pilotAPIEvent{{Category: "hint", Phase: "fallback", Method: "bm25_fallback", Attempted: true, Returned: true}}
	recount(&y)
	x.Pipelines[pos] = y
	x.SetupHintCounts, x.SetupIOCounts, x.ModelRankAttempts, x.ModelRankReturns, _ = eventCounts(x.SetupEvents)
	for _, v := range x.Pipelines {
		x.ModelRankAttempts += v.ModelRankAttempts
		x.ModelRankReturns += v.ModelRankReturns
	}
	if e := checkWire(x, "warm", "", 0, p.In, p.Wants, p.Saved, p.Expected); e != nil {
		t.Fatal(e)
	}
	for _, panicked := range []bool{false, true} {
		bad := copyPipeline(t, y)
		if panicked {
			kind := "synthetic_panic"
			bad.HintEvents[0].Returned = false
			bad.HintEvents[0].Panic = true
			bad.HintEvents[0].PanicType = &kind
		} else {
			bad.HintEvents[0].Error = true
			bad.HintEvents[0].ErrorCode = &code
		}
		recount(&bad)
		x.Pipelines[pos] = bad
		if checkWire(x, "warm", "", 0, p.In, p.Wants, p.Saved, p.Expected) == nil {
			t.Fatal("failed fallback retained a successful ranking")
		}
		bad.Ranking, bad.RankSnapshot, bad.SelectedIndex = nil, nil, nil
		bad.State, bad.Visits, bad.Visited = "unresolved", nil, [5]bool{}
		recount(&bad)
		x.Pipelines[pos] = bad
		if e := checkWire(x, "warm", "", 0, p.In, p.Wants, p.Saved, p.Expected); e != nil {
			t.Fatal("failed fallback lost its unresolved shape", e)
		}
	}
	x.Pipelines[pos] = y
	x.SetupEvents = x.SetupEvents[:5]
	x.SetupEvents[4].Error = true
	x.SetupEvents[4].ErrorCode = &code
	x.SetupHintCounts, x.SetupIOCounts, _, _, _ = eventCounts(x.SetupEvents)
	if checkWire(x, "warm", "", 0, p.In, p.Wants, p.Saved, p.Expected) == nil {
		t.Fatal("IO error converted to shared preparation")
	}
}

func testBundle(t *testing.T, p packet) bundle {
	t.Helper()
	b := bundle{Schema: "riido-chi-paired-cost-bundle-v1", Status: "completed_development_cost_audit", FreezeSHA: p.FreezeSHA, SourcePinsSHA: p.SourceSHA, Parents: 1, Cleanup: true, Scope: "synthetic bundle; no actual target/model execution"}
	add := func(stage, policy string, block int, out []byte) {
		b.Processes = append(b.Processes, processRecord{Stage: stage, Block: block, Policy: policy, ElapsedNS: 100, Exit: 0, Complete: true, Stdout: out, StdoutSHA: sha(out), StdoutReceived: len(out), StdoutComplete: true, StderrSHA: sha(nil), StderrComplete: true})
	}
	add("go_version", "", 0, []byte("go version go1.27.1 linux/arm64\n"))
	var pass string
	for _, n := range controlNames {
		pass += "--- PASS: " + n + " (0.00s)\n"
	}
	add("synthetic_controls", "", 0, []byte(pass))
	add("build", "", 0, nil)
	worker := func(mode, policy string, block int) {
		x := testWire(p, mode, policy, block)
		add(mode, policy, block, testZip(t, x))
		b.Pipelines += len(x.Pipelines)
		b.ModelAttempts += x.ModelRankAttempts
		b.ModelReturns += x.ModelRankReturns
	}
	worker("precheck", "", 0)
	worker("anchor", "", 0)
	for block, row := range orders {
		for _, policy := range row {
			worker("cold", policy, block)
		}
	}
	for block := range orders {
		worker("warm", "", block)
	}
	return b
}
func TestSavedBundleRecountsAndSchedule(t *testing.T) {
	p := testPacket(t)
	b := testBundle(t, p)
	if e := checkBundle(b, p); e != nil {
		t.Fatal(e)
	}
	for _, kind := range []string{"duplicate", "missing", "cleanup", "time", "controls"} {
		x := testBundle(t, p)
		switch kind {
		case "cleanup":
			x.Cleanup = false
		case "controls":
			x.Processes[1].Stdout = []byte("testing: warning: no tests to run\nPASS\n")
			x.Processes[1].StdoutSHA = sha(x.Processes[1].Stdout)
			x.Processes[1].StdoutReceived = len(x.Processes[1].Stdout)
		default:
			q := &x.Processes[5]
			raw, _ := unpack(q.Stdout)
			var v pilotWire
			_ = decode(raw, &v)
			if kind == "duplicate" {
				v.Pipelines = append(v.Pipelines, v.Pipelines[0])
				x.Pipelines++
			} else if kind == "missing" {
				v.Pipelines = nil
				x.Pipelines--
			} else {
				v.Pipelines[0].Cost.ElapsedNS = q.ElapsedNS + 1
			}
			q.Stdout = testZip(t, v)
			q.StdoutSHA = sha(q.Stdout)
			q.StdoutReceived = len(q.Stdout)
		}
		if checkBundle(x, p) == nil {
			t.Fatalf("saved corruption accepted: %s", kind)
		}
	}
	compressed := testZip(t, b)
	if _, e := unpackLimit(compressed[:len(compressed)-1], bundleCap); e == nil {
		t.Fatal("truncated saved gzip")
	}
}
func TestCleanupFailurePreservesReceiptAndPrimary(t *testing.T) {
	b := bundle{Schema: "riido-chi-paired-cost-bundle-v1", Status: "failed", Processes: []processRecord{{Stage: "anchor", Exit: 1}}}
	primary := errors.New("worker_protocol_failure")
	e := cleanupResult(&b, primary, func() error { return errors.New("private cleanup error withheld") })
	if e != primary || b.Cleanup || b.CleanupFailure != "cleanup_failure" || len(b.Processes) != 1 {
		t.Fatal("cleanup lost evidence")
	}
	b.Failure = e.Error()
	path := filepath.Join(t.TempDir(), "failed.gz")
	if writeBundle(path, b) != nil {
		t.Fatal("failure receipt missing")
	}
	compressed, e := read(path, bundleCap)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := unpackLimit(compressed, bundleCap)
	var got bundle
	if e != nil || decode(raw, &got) != nil || got.Status != "failed" || got.Cleanup || got.Failure != primary.Error() || got.CleanupFailure != "cleanup_failure" || len(got.Processes) != 1 {
		t.Fatal("failure receipt changed")
	}
}
