// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 teamswyg contributors.
// One exposed Chi input: setup attribution and repeated scorer costs, not
// independent examples or end-to-end verifier utility. The inactive failed79
// scorer stays inactive. Profile collection is deliberately absent.
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
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/hintprepared"
	"github.com/teamswyg/laya-tools/pkg/hintweights"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const probeSchema = "riido-chi-setup-reuse-v1"
const probeInputSHA = "39e0aadbfe91c0b15ad3da257725de8d74c7a82c9d4222f73da8f937e72bf4c2"
const probeModelSHA = "dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004"
const probeModelRef = "JooYoon/riidolaya-shortclaim-data-effect-failed-79@5bef215895b69d3f2ef4b82bb3f1970279d67f46"
const probeScope = "one exposed development parent; scorer costs only; instrumented calls exclude Prepared accessor/runtime/stdlib/hidden calls; diagnostic setup is not added twice; 96MiB is a soft target; Go allocation not retained heap/RSS/GPU; no profiles, Fit, qualification or activation"
const probeBound = 262144

var probeCandidateIDs = [5]string{"post_path", "post_query", "pre_path", "twice_path", "unchanged"}
var probeNs = [4]int{1, 2, 4, 8}
var probeOrders = [6][3]string{
	{"reuse", "rebuild", "bm25"}, {"reuse", "bm25", "rebuild"},
	{"rebuild", "reuse", "bm25"}, {"rebuild", "bm25", "reuse"},
	{"bm25", "reuse", "rebuild"}, {"bm25", "rebuild", "reuse"},
}

type probeCost struct {
	ElapsedNS       int64  `json:"elapsed_ns"`
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`
	Mallocs         uint64 `json:"mallocs"`
}
type probeEvent struct {
	API       string `json:"api"`
	Attempted bool   `json:"attempted"`
	Returned  bool   `json:"returned"`
	Error     bool   `json:"error"`
	Panic     bool   `json:"panic"`
	ErrorCode string `json:"error_code"`
	PanicType string `json:"panic_type"`
}
type probeCounts struct {
	Attempted       int `json:"attempted"`
	Returned        int `json:"returned"`
	Errors          int `json:"errors"`
	Panics          int `json:"panics"`
	PrepareAttempts int `json:"prepare_attempts"`
	RankAttempts    int `json:"rank_attempts"`
}
type probeSnapshot struct {
	Kind           string    `json:"kind"`
	Count          int       `json:"count"`
	Order          [8]int    `json:"order"`
	ScoreBits      [8]string `json:"score_bits"`
	CandidateIDs   [5]string `json:"candidate_ids"`
	OrderedIDs     [5]string `json:"ordered_ids"`
	FallbackReason string    `json:"fallback_reason"`
}
type probeResult struct {
	Event         probeEvent     `json:"event"`
	Ranking       *probeSnapshot `json:"ranking"`
	MatchesAnchor *bool          `json:"matches_anchor"`
}
type probeIntervalResult struct {
	N              int           `json:"n"`
	Permutation    int           `json:"permutation"`
	Position       int           `json:"position"`
	Method         string        `json:"method"`
	State          string        `json:"state"`
	Prepares       []probeEvent  `json:"prepares"`
	Calls          []probeResult `json:"calls"`
	RemainingRanks int           `json:"remaining_ranks"`
	Cost           probeCost     `json:"cost"`
	Counts         probeCounts   `json:"counts"`
}
type probeStage struct {
	Name  string     `json:"name"`
	State string     `json:"state"`
	Event probeEvent `json:"event"`
	Cost  probeCost  `json:"cost"`
}
type probeReport struct {
	Schema              string                `json:"schema"`
	State               string                `json:"state"`
	InputSHA            string                `json:"input_sha256"`
	ModelSHA            string                `json:"model_sha256"`
	ModelRef            string                `json:"model_ref"`
	Scope               string                `json:"scope"`
	InputBytes          int                   `json:"input_bytes"`
	ModelBytes          int                   `json:"model_bytes"`
	Parents             int                   `json:"parents"`
	GOMAXPROCS          int                   `json:"gomaxprocs"`
	MemoryLimitBytes    int64                 `json:"memory_limit_bytes"`
	RuntimeGo           string                `json:"runtime_go"`
	Fit                 int                   `json:"fit"`
	NewLabels           int                   `json:"new_labels"`
	NewRoles            int                   `json:"new_roles"`
	Qualified           bool                  `json:"qualified"`
	DefaultActivated    bool                  `json:"default_activated"`
	GPU                 bool                  `json:"gpu"`
	ProtectedEvaluation bool                  `json:"protected_evaluation"`
	Profiled            bool                  `json:"profiled"`
	Stages              []probeStage          `json:"stages"`
	ModelAnchor         *probeSnapshot        `json:"model_anchor"`
	BM25Anchor          *probeSnapshot        `json:"bm25_anchor"`
	Intervals           []probeIntervalResult `json:"intervals"`
	Totals              probeCounts           `json:"totals"`
}
type probeCallbacks struct {
	prepare func() error
	rank    func(method string) (shortclaim.Ranking, error)
}

// Explicit wrapper events distinguish errors from panic(nil). No error text or
// panic values are serialized. A returned-error Ranking is still retained.
func probeCall[T any](api string, fn func() (T, error)) (out T, event probeEvent) {
	event = probeEvent{API: api, Attempted: true}
	complete := false
	defer func() {
		if !complete {
			p := recover()
			event.Panic = true
			event.PanicType = fmt.Sprintf("%T", p)
		}
	}()
	var err error
	out, err = fn()
	complete = true
	event.Returned = true
	if err != nil {
		event.Error = true
		event.ErrorCode = "returned_error"
	}
	return
}
func probeOK(e probeEvent) bool { return e.Attempted && e.Returned && !e.Error && !e.Panic }
func probeAdd(c *probeCounts, e probeEvent) {
	if e.Attempted {
		c.Attempted++
		if e.API == "prepare" {
			c.PrepareAttempts++
		}
		if e.API == "model_rank" || e.API == "bm25_rank" {
			c.RankAttempts++
		}
	}
	if e.Returned {
		c.Returned++
	}
	if e.Error {
		c.Errors++
	}
	if e.Panic {
		c.Panics++
	}
}
func probeSum(c *probeCounts, x probeCounts) {
	c.Attempted += x.Attempted
	c.Returned += x.Returned
	c.Errors += x.Errors
	c.Panics += x.Panics
	c.PrepareAttempts += x.PrepareAttempts
	c.RankAttempts += x.RankAttempts
}
func probeMeasure(fn func()) probeCost {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	fn()
	elapsed := time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&after)
	return probeCost{elapsed, after.TotalAlloc - before.TotalAlloc, after.Mallocs - before.Mallocs}
}
func probeSnap(r shortclaim.Ranking) probeSnapshot {
	s := probeSnapshot{Kind: r.Kind, Count: r.Count, Order: r.Order, CandidateIDs: probeCandidateIDs, FallbackReason: r.FallbackReason}
	for i := 0; i < 8; i++ {
		s.ScoreBits[i] = fmt.Sprintf("%016x", math.Float64bits(r.Scores[i]))
	}
	for i := 0; i < 5 && i < r.Count; i++ {
		x := r.Order[i]
		if x >= 0 && x < 5 {
			s.OrderedIDs[i] = probeCandidateIDs[x]
		}
	}
	return s
}
func probeFrozenModel() probeSnapshot {
	return probeSnapshot{Kind: hintprepared.Kind, Count: 5, Order: [8]int{2, 4, 1, 3, 0},
		ScoreBits:    [8]string{"bfc4dd509998bba7", "bfc40874752fd85e", "bfb69c0f88bf8dcb", "bfc45cd2f7e3a36f", "bfbc36f53a763f34", "0000000000000000", "0000000000000000", "0000000000000000"},
		CandidateIDs: probeCandidateIDs, OrderedIDs: [5]string{"pre_path", "unchanged", "post_query", "twice_path", "post_path"}}
}
func probeFrozenBM25() probeSnapshot {
	return probeSnapshot{Kind: shortclaim.BM25Kind, Count: 5, Order: [8]int{0, 1, 4, 3, 2},
		ScoreBits:    [8]string{"40163cc795d1d7bd", "40161b65c1868508", "400b0168d0fa2058", "40141187fc8a5cf1", "4015a721dda3640c", "0000000000000000", "0000000000000000", "0000000000000000"},
		CandidateIDs: probeCandidateIDs, OrderedIDs: [5]string{"post_path", "post_query", "unchanged", "twice_path", "pre_path"}}
}
func probeSnapshotValid(s probeSnapshot, kind string) bool {
	if s.Kind != kind || s.Count != 5 || s.CandidateIDs != probeCandidateIDs || s.FallbackReason != "" {
		return false
	}
	var scores [8]float64
	for i, b := range s.ScoreBits {
		bits, err := strconv.ParseUint(b, 16, 64)
		if err != nil || len(b) != 16 || fmt.Sprintf("%016x", bits) != b {
			return false
		}
		scores[i] = math.Float64frombits(bits)
		if math.IsNaN(scores[i]) || math.IsInf(scores[i], 0) {
			return false
		}
		if i >= 5 && (bits != 0 || s.Order[i] != 0) {
			return false
		}
	}
	var seen [5]bool
	for i := 0; i < 5; i++ {
		x := s.Order[i]
		if x < 0 || x >= 5 || seen[x] || s.OrderedIDs[i] != probeCandidateIDs[x] {
			return false
		}
		seen[x] = true
		if i > 0 {
			a := s.Order[i-1]
			if scores[a] < scores[x] || scores[a] == scores[x] && a > x {
				return false
			}
		}
	}
	return true
}

// Fixed raw arrays receive every actual returned value inside the interval;
// conversion, parity checks and JSON happen afterward. No output padding
// masquerades as a call. Setup diagnostic Prepare is a separate observation.
func probeInterval(n, permutation, position int, method string, callbacks probeCallbacks, measure func(func()) probeCost, modelAnchor, bm25Anchor probeSnapshot) probeIntervalResult {
	out := probeIntervalResult{N: n, Permutation: permutation, Position: position, Method: method, State: "unavailable", Prepares: []probeEvent{}, Calls: []probeResult{}}
	if n < 1 || n > 8 || (method != "reuse" && method != "rebuild" && method != "bm25") {
		return out
	}
	var prepEvents, rankEvents [8]probeEvent
	var ranks [8]shortclaim.Ranking
	preps, calls := 0, 0
	out.Cost = measure(func() {
		prepare := func() bool {
			_, e := probeCall("prepare", func() (struct{}, error) { return struct{}{}, callbacks.prepare() })
			prepEvents[preps] = e
			preps++
			return probeOK(e)
		}
		if method == "reuse" && !prepare() {
			return
		}
		for i := 0; i < n; i++ {
			if method == "rebuild" && !prepare() {
				return
			}
			api := "model_rank"
			if method == "bm25" {
				api = "bm25_rank"
			}
			r, e := probeCall(api, func() (shortclaim.Ranking, error) { return callbacks.rank(method) })
			ranks[calls], rankEvents[calls] = r, e
			calls++
			if !probeOK(e) {
				return
			}
		}
	})
	out.RemainingRanks = n - calls
	complete := calls == n
	for i := 0; i < preps; i++ {
		out.Prepares = append(out.Prepares, prepEvents[i])
		probeAdd(&out.Counts, prepEvents[i])
		if !probeOK(prepEvents[i]) {
			complete = false
		}
	}
	parity := true
	anchor := modelAnchor
	kind := hintprepared.Kind
	if method == "bm25" {
		anchor = bm25Anchor
		kind = shortclaim.BM25Kind
	}
	for i := 0; i < calls; i++ {
		e := rankEvents[i]
		r := probeResult{Event: e}
		probeAdd(&out.Counts, e)
		if e.Returned {
			s := probeSnap(ranks[i])
			r.Ranking = &s
		}
		if probeOK(e) {
			equal := probeSnapshotValid(*r.Ranking, kind) && *r.Ranking == anchor
			r.MatchesAnchor = &equal
			if !equal {
				parity = false
			}
		} else {
			complete = false
		}
		out.Calls = append(out.Calls, r)
	}
	if complete {
		out.State = "completed"
		if !parity {
			out.State = "parity_failure"
		}
	}
	return out
}
func probeRead(path string, n int, pin string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.New("read_unavailable")
	}
	raw, err := io.ReadAll(io.LimitReader(f, int64(n)+1))
	closeErr := f.Close()
	if err != nil || closeErr != nil || len(raw) != n {
		return nil, errors.New("read_bound")
	}
	hash := sha256.Sum256(raw)
	if hex.EncodeToString(hash[:]) != pin {
		return nil, errors.New("read_pin")
	}
	return raw, nil
}
func probeStageCall[T any](r *probeReport, name, api string, fn func() (T, error)) (value T, ok bool) {
	var e probeEvent
	c := probeMeasure(func() { value, e = probeCall(api, fn) })
	state := "completed"
	if !probeOK(e) {
		state = "unavailable"
	}
	r.Stages = append(r.Stages, probeStage{name, state, e, c})
	probeAdd(&r.Totals, e)
	return value, state == "completed"
}
func probeLive(inputPath, modelPath string) probeReport {
	oldP := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(oldP)
	oldM := debug.SetMemoryLimit(96 << 20)
	defer debug.SetMemoryLimit(oldM)
	r := probeReport{Schema: probeSchema, State: "unavailable", InputSHA: probeInputSHA, ModelSHA: probeModelSHA, ModelRef: probeModelRef, Scope: probeScope, InputBytes: 1417, ModelBytes: 32792, Parents: 1, GOMAXPROCS: runtime.GOMAXPROCS(0), MemoryLimitBytes: 96 << 20, RuntimeGo: runtime.Version(), Stages: []probeStage{}, Intervals: []probeIntervalResult{}}
	input, ok := probeStageCall(&r, "input_read", "input_read", func() ([]byte, error) { return probeRead(inputPath, 1417, probeInputSHA) })
	if !ok {
		return r
	}
	raw, ok := probeStageCall(&r, "model_read", "model_read", func() ([]byte, error) { return probeRead(modelPath, 32792, probeModelSHA) })
	if !ok {
		return r
	}
	valid, ok := probeStageCall(&r, "load_validated", "load_validated", func() (shortclaim.ValidatedInput, error) { return shortclaim.LoadValidated(bytes.NewReader(input)) })
	if !ok {
		return r
	}
	p := valid.Prepared()
	if p.Count != 5 {
		return r
	}
	for i, id := range probeCandidateIDs {
		if p.Candidates[i].ID != id {
			return r
		}
	}
	view, ok := probeStageCall(&r, "view_new", "view_new", func() (*hintweights.View, error) { return hintweights.New(raw) })
	if !ok {
		return r
	}
	owner, ok := probeStageCall(&r, "prepare", "prepare", func() (*hintprepared.Prepared, error) { return hintprepared.Prepare(p) })
	if !ok {
		return r
	}
	model, ok := probeStageCall(&r, "model_anchor", "model_rank", func() (shortclaim.Ranking, error) { return owner.Rank(valid, view) })
	if !ok {
		return r
	}
	ma := probeSnap(model)
	r.ModelAnchor = &ma
	bm25, ok := probeStageCall(&r, "bm25_anchor", "bm25_rank", func() (shortclaim.Ranking, error) { return valid.Rank(shortclaim.BM25Kind) })
	if !ok {
		return r
	}
	ba := probeSnap(bm25)
	r.BM25Anchor = &ba
	if ma != probeFrozenModel() || ba != probeFrozenBM25() || !probeSnapshotValid(ma, hintprepared.Kind) || !probeSnapshotValid(ba, shortclaim.BM25Kind) {
		return r
	}
	owner = nil // Diagnostic owner is not a reusable owner in a main interval.
	for _, n := range probeNs {
		for order, row := range probeOrders {
			for position, method := range row {
				var prepared *hintprepared.Prepared
				callbacks := probeCallbacks{prepare: func() error { var err error; prepared, err = hintprepared.Prepare(p); return err }, rank: func(which string) (shortclaim.Ranking, error) {
					if which == "bm25" {
						return valid.Rank(shortclaim.BM25Kind)
					}
					return prepared.Rank(valid, view)
				}}
				interval := probeInterval(n, order, position, method, callbacks, probeMeasure, ma, ba)
				r.Intervals = append(r.Intervals, interval)
				probeSum(&r.Totals, interval.Counts)
			}
		}
	}
	r.State = "completed"
	for _, x := range r.Intervals {
		if x.State != "completed" {
			r.State = "unavailable"
		}
	}
	// Common borrowed input/model bytes and View intentionally coexist in all
	// intervals, including BM25; this is not a cold first-use baseline comparison.
	runtime.KeepAlive(input)
	runtime.KeepAlive(raw)
	runtime.KeepAlive(valid)
	runtime.KeepAlive(view)
	return r
}

func probeEventValid(e probeEvent, api string) bool {
	if e.API != api || !e.Attempted || e.Returned == e.Panic {
		return false
	}
	if e.Panic {
		return !e.Error && e.ErrorCode == "" && e.PanicType != ""
	}
	return e.PanicType == "" && ((e.Error && e.ErrorCode == "returned_error") || (!e.Error && e.ErrorCode == ""))
}
func probeValidateInterval(x probeIntervalResult, ma, ba probeSnapshot) error {
	fail := errors.New("invalid_interval")
	if x.Cost.ElapsedNS < 0 || x.RemainingRanks != x.N-len(x.Calls) || len(x.Calls) > x.N || len(x.Prepares) > x.N {
		return fail
	}
	anchor := ma
	api := "model_rank"
	kind := hintprepared.Kind
	if x.Method == "bm25" {
		anchor = ba
		api = "bm25_rank"
		kind = shortclaim.BM25Kind
	}
	var counts probeCounts
	prepareFailed := false
	for i, e := range x.Prepares {
		if !probeEventValid(e, "prepare") || prepareFailed {
			return fail
		}
		if !probeOK(e) {
			if i != len(x.Prepares)-1 {
				return fail
			}
			prepareFailed = true
		}
		probeAdd(&counts, e)
	}
	rankFailed, parityFailed := false, false
	for i, c := range x.Calls {
		if !probeEventValid(c.Event, api) || rankFailed {
			return fail
		}
		probeAdd(&counts, c.Event)
		if c.Event.Returned {
			if c.Ranking == nil || c.Ranking.CandidateIDs != probeCandidateIDs {
				return fail
			}
		} else if c.Ranking != nil {
			return fail
		}
		if probeOK(c.Event) {
			if !probeSnapshotValid(*c.Ranking, kind) {
				return fail
			}
			equal := *c.Ranking == anchor
			if c.MatchesAnchor == nil || *c.MatchesAnchor != equal {
				return fail
			}
			if !equal {
				parityFailed = true
			}
		} else {
			if i != len(x.Calls)-1 || c.MatchesAnchor != nil {
				return fail
			}
			rankFailed = true
		}
	}
	switch x.Method {
	case "reuse":
		if len(x.Prepares) != 1 || prepareFailed && len(x.Calls) != 0 {
			return fail
		}
	case "rebuild":
		extra := 0
		if prepareFailed {
			extra = 1
		}
		if len(x.Prepares) != len(x.Calls)+extra {
			return fail
		}
	case "bm25":
		if len(x.Prepares) != 0 {
			return fail
		}
	default:
		return fail
	}
	if prepareFailed && rankFailed {
		return fail
	} // The actual loop stops on either.
	state := "completed"
	if prepareFailed || rankFailed {
		state = "unavailable"
	} else if len(x.Calls) != x.N {
		return fail
	} else if parityFailed {
		state = "parity_failure"
	}
	if x.State != state || counts != x.Counts {
		return fail
	}
	return nil
}
func probeValidateReport(r probeReport) error {
	fail := errors.New("invalid_report")
	if r.Schema != probeSchema || r.InputSHA != probeInputSHA || r.ModelSHA != probeModelSHA || r.ModelRef != probeModelRef || r.Scope != probeScope || r.InputBytes != 1417 || r.ModelBytes != 32792 || r.Parents != 1 || r.GOMAXPROCS != 1 || r.MemoryLimitBytes != 96<<20 || r.RuntimeGo != "go1.27.1" || r.Fit != 0 || r.NewLabels != 0 || r.NewRoles != 0 || r.Qualified || r.DefaultActivated || r.GPU || r.ProtectedEvaluation || r.Profiled {
		return fail
	}
	if r.ModelAnchor == nil || r.BM25Anchor == nil || *r.ModelAnchor != probeFrozenModel() || *r.BM25Anchor != probeFrozenBM25() {
		return fail
	}
	if len(r.Stages) != 7 || len(r.Intervals) != 72 {
		return fail
	}
	names := [7]string{"input_read", "model_read", "load_validated", "view_new", "prepare", "model_anchor", "bm25_anchor"}
	var totals probeCounts
	for i, s := range r.Stages {
		api := names[i]
		if i == 5 {
			api = "model_rank"
		}
		if i == 6 {
			api = "bm25_rank"
		}
		if s.Name != names[i] || s.State != "completed" || s.Cost.ElapsedNS < 0 || !probeEventValid(s.Event, api) || !probeOK(s.Event) {
			return fail
		}
		probeAdd(&totals, s.Event)
	}
	index := 0
	state := "completed"
	for _, n := range probeNs {
		for permutation, row := range probeOrders {
			for position, method := range row {
				x := r.Intervals[index]
				index++
				if x.N != n || x.Permutation != permutation || x.Position != position || x.Method != method || probeValidateInterval(x, *r.ModelAnchor, *r.BM25Anchor) != nil {
					return fail
				}
				probeSum(&totals, x.Counts)
				if x.State != "completed" {
					state = "unavailable"
				}
			}
		}
	}
	if r.State != state || r.Totals != totals {
		return fail
	}
	return nil
}

// This strict reader is for this evidence schema, not canonical JSON. It
// rejects duplicate decoded keys (including escaped keys), wrong-case/unknown
// or missing fields, trailing values and fixed-array length changes.
func probeKeys(d *json.Decoder, depth int) error {
	if depth > 32 {
		return errors.New("json_depth")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		var keys [64]string
		n := 0
		for d.More() {
			t, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := t.(string)
			if !ok || n == len(keys) {
				return errors.New("json_keys")
			}
			for i := 0; i < n; i++ {
				if keys[i] == key {
					return errors.New("json_duplicate")
				}
			}
			keys[n] = key
			n++
			if err := probeKeys(d, depth+1); err != nil {
				return err
			}
		}
		t, err := d.Token()
		if err != nil || t != json.Delim('}') {
			return errors.New("json_object")
		}
		return nil
	}
	if delimiter == '[' {
		for d.More() {
			if err := probeKeys(d, depth+1); err != nil {
				return err
			}
		}
		t, err := d.Token()
		if err != nil || t != json.Delim(']') {
			return errors.New("json_array")
		}
		return nil
	}
	return errors.New("json_delimiter")
}
func probeShape(raw []byte, t reflect.Type) error {
	fail := errors.New("json_shape")
	raw = bytes.TrimSpace(raw)
	if t.Kind() == reflect.Pointer {
		if bytes.Equal(raw, []byte("null")) {
			return nil
		}
		return probeShape(raw, t.Elem())
	}
	if bytes.Equal(raw, []byte("null")) {
		return fail
	}
	switch t.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil || len(fields) != t.NumField() {
			return fail
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			value, ok := fields[f.Tag.Get("json")]
			if !ok || probeShape(value, f.Type) != nil {
				return fail
			}
		}
	case reflect.Array, reflect.Slice:
		var items []json.RawMessage
		if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &items) != nil {
			return fail
		}
		if t.Kind() == reflect.Array && len(items) != t.Len() {
			return fail
		}
		for _, item := range items {
			if probeShape(item, t.Elem()) != nil {
				return fail
			}
		}
	}
	return nil
}
func probeDecode(raw []byte) (r probeReport, err error) {
	fail := errors.New("saved_audit_failed")
	if len(raw) == 0 || len(raw) > probeBound || !utf8.Valid(raw) {
		return r, fail
	}
	keys := json.NewDecoder(bytes.NewReader(raw))
	keys.UseNumber()
	if probeKeys(keys, 0) != nil {
		return r, fail
	}
	if _, err := keys.Token(); err != io.EOF {
		return r, fail
	}
	if probeShape(raw, reflect.TypeOf(r)) != nil {
		return r, fail
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&r) != nil {
		return probeReport{}, fail
	}
	var extra any
	if d.Decode(&extra) != io.EOF || probeValidateReport(r) != nil {
		return probeReport{}, fail
	}
	return r, nil
}
func probeSaved(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("saved_read")
	}
	raw, err := io.ReadAll(io.LimitReader(f, probeBound+1))
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return errors.New("saved_read")
	}
	r, err := probeDecode(raw)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "setup/reuse saved audit valid: state=%s intervals=72 parents=1 Fit=0; saved validation calls no model APIs\n", r.State)
	return err
}
func probeMain() error {
	fs := flag.NewFlagSet("setup-reuse", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	input := fs.String("input", "", "pinned public input")
	model := fs.String("model", "", "existing inactive model")
	saved := fs.String("saved", "", "strict plain-JSON evidence audit, no model calls")
	if fs.Parse(os.Args[1:]) != nil || fs.NArg() != 0 {
		return errors.New("arguments")
	}
	// Saved mode branches before runtime limits, input/model reads or SDK calls.
	if *saved != "" {
		if *input != "" || *model != "" {
			return errors.New("arguments")
		}
		return probeSaved(*saved)
	}
	if *input == "" || *model == "" {
		return errors.New("arguments")
	}
	r := probeLive(*input, *model)
	raw, err := json.Marshal(r)
	if err != nil || len(raw) > probeBound {
		return errors.New("output_bound")
	}
	n, err := os.Stdout.Write(raw)
	if err != nil || n != len(raw) {
		return errors.New("output_write")
	}
	if r.State != "completed" {
		return errors.New("live_unavailable")
	}
	return nil
}
func main() {
	if probeMain() != nil {
		fmt.Fprintln(os.Stderr, "setup/reuse probe failed")
		os.Exit(1)
	}
}
