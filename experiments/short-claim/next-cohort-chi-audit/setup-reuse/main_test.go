// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

// Fake callbacks; no APIs or model reads.
func fakeRank(s probeSnapshot) shortclaim.Ranking {
	r := shortclaim.Ranking{Kind: s.Kind, Count: s.Count, Order: s.Order, FallbackReason: s.FallbackReason}
	for i, b := range s.ScoreBits {
		u, _ := strconv.ParseUint(b, 16, 64)
		r.Scores[i] = math.Float64frombits(u)
	}
	return r
}
func fakeMeasure(f func()) probeCost { f(); return probeCost{} }
func fakeCalls(m, b probeSnapshot) probeCallbacks {
	return probeCallbacks{prepare: func() error { return nil }, rank: func(which string) (shortclaim.Ranking, error) {
		if which == "bm25" {
			return fakeRank(b), nil
		}
		return fakeRank(m), nil
	}}
}
func fakeReport() probeReport {
	m, b := probeFrozenModel(), probeFrozenBM25()
	r := probeReport{Schema: probeSchema, State: "completed", InputSHA: probeInputSHA, ModelSHA: probeModelSHA, ModelRef: probeModelRef, Scope: probeScope, InputBytes: 1417, ModelBytes: 32792, Parents: 1, GOMAXPROCS: 1, MemoryLimitBytes: 96 << 20, RuntimeGo: "go1.27.1", ModelAnchor: &m, BM25Anchor: &b}
	for _, name := range []string{"input_read", "model_read", "load_validated", "view_new", "prepare", "model_anchor", "bm25_anchor"} {
		api := name
		if name == "model_anchor" {
			api = "model_rank"
		}
		if name == "bm25_anchor" {
			api = "bm25_rank"
		}
		e := probeEvent{API: api, Attempted: true, Returned: true}
		r.Stages = append(r.Stages, probeStage{Name: name, State: "completed", Event: e})
		probeAdd(&r.Totals, e)
	}
	for _, n := range probeNs {
		for p, row := range probeOrders {
			for pos, method := range row {
				x := probeInterval(n, p, pos, method, fakeCalls(m, b), fakeMeasure, m, b)
				r.Intervals = append(r.Intervals, x)
				probeSum(&r.Totals, x.Counts)
			}
		}
	}
	return r
}

func TestProbePrepareAndAnchors(t *testing.T) {
	m, b := probeFrozenModel(), probeFrozenBM25()
	for _, n := range []int{1, 8} {
		for _, method := range []string{"reuse", "rebuild", "bm25"} {
			preps, ranks := 0, 0
			c := fakeCalls(m, b)
			c.prepare = func() error { preps++; return nil }
			old := c.rank
			c.rank = func(which string) (shortclaim.Ranking, error) { ranks++; return old(which) }
			x := probeInterval(n, 0, 0, method, c, fakeMeasure, m, b)
			want := 1
			if method == "rebuild" {
				want = n
			}
			if method == "bm25" {
				want = 0
			}
			if preps != want || ranks != n || x.Counts.PrepareAttempts != want || x.Counts.RankAttempts != n || x.Counts.Attempted != want+n || x.Counts.Returned != want+n || x.State != "completed" || x.RemainingRanks != 0 {
				t.Fatalf("%s n%d: %#v", method, n, x)
			}
			for _, call := range x.Calls {
				if call.MatchesAnchor == nil || !*call.MatchesAnchor {
					t.Fatal("own anchor")
				}
			}
		}
	}
}

func TestProbePartialReturns(t *testing.T) {
	m, b := probeFrozenModel(), probeFrozenBM25()
	for _, mode := range []string{"rank_error", "rank_panic", "prepare_error", "prepare_panic"} {
		c := fakeCalls(m, b)
		calls := 0
		if mode == "prepare_error" {
			c.prepare = func() error { return errors.New("fake") }
		}
		if mode == "prepare_panic" {
			c.prepare = func() error { panic(nil) }
		}
		if mode == "rank_error" {
			c.rank = func(string) (shortclaim.Ranking, error) {
				calls++
				if calls == 2 {
					return fakeRank(m), errors.New("fake")
				}
				return fakeRank(m), nil
			}
		}
		if mode == "rank_panic" {
			c.rank = func(string) (shortclaim.Ranking, error) { panic(nil) }
		}
		x := probeInterval(4, 0, 0, "reuse", c, fakeMeasure, m, b)
		if x.State != "unavailable" || x.RemainingRanks != 4-len(x.Calls) || probeValidateInterval(x, m, b) != nil {
			t.Fatal(mode, x)
		}
		if mode == "rank_error" {
			if len(x.Calls) != 2 {
				t.Fatal("calls", x)
			}
			last := x.Calls[1]
			if x.Counts.Attempted != 3 || x.Counts.Returned != 3 || x.Counts.Errors != 1 || last.Ranking == nil || *last.Ranking != m || last.MatchesAnchor != nil {
				t.Fatal("error return", x)
			}
		} else if mode == "rank_panic" {
			if len(x.Calls) != 1 {
				t.Fatal("panic calls", x)
			}
			e := x.Calls[0]
			if x.Counts.Panics != 1 || x.Counts.Returned != 1 || e.Event.Returned || !e.Event.Panic || e.Ranking != nil || e.MatchesAnchor != nil {
				t.Fatal("panic return", x)
			}
		} else {
			e := x.Prepares[0]
			if len(x.Calls) != 0 || x.Counts.Attempted != 1 || x.Counts.RankAttempts != 0 || !e.Attempted {
				t.Fatal("invented rank", x)
			}
			if mode == "prepare_panic" && (!e.Panic || e.Returned || x.Counts.Returned != 0) {
				t.Fatal("prepare panic", x)
			}
			if mode == "prepare_error" && (!e.Error || !e.Returned || x.Counts.Errors != 1) {
				t.Fatal("prepare error", x)
			}
		}
	}
	c := fakeCalls(m, b)
	c.rank = func(string) (shortclaim.Ranking, error) { return fakeRank(m), errors.New("fake") }
	x := probeInterval(4, 0, 0, "rebuild", c, fakeMeasure, m, b)
	e := probeEvent{API: "prepare", Attempted: true, Returned: true, Error: true, ErrorCode: "returned_error"}
	x.Prepares = append(x.Prepares, e)
	probeAdd(&x.Counts, e)
	if probeValidateInterval(x, m, b) == nil {
		t.Fatal("prepare after rank failure")
	}
}

func TestProbeRankingIntegrity(t *testing.T) {
	m, b := probeFrozenModel(), probeFrozenBM25()
	for _, change := range []func(*shortclaim.Ranking){
		func(r *shortclaim.Ranking) { r.Scores[0] += 1 },
		func(r *shortclaim.Ranking) { r.Order[1] = r.Order[0] },
		func(r *shortclaim.Ranking) { r.Scores[5] = math.Copysign(0, -1) },
		func(r *shortclaim.Ranking) { r.Scores[0] = math.NaN() },
	} {
		c := fakeCalls(m, b)
		r := fakeRank(m)
		change(&r)
		c.rank = func(string) (shortclaim.Ranking, error) { return r, nil }
		x := probeInterval(1, 0, 0, "reuse", c, fakeMeasure, m, b)
		if x.State != "parity_failure" || x.Calls[0].Ranking == nil || x.Calls[0].MatchesAnchor == nil || *x.Calls[0].MatchesAnchor {
			t.Fatal("bad ranking", x)
		}
	}
}

func TestProbeStrictCompleteReport(t *testing.T) {
	good := fakeReport()
	raw, err := json.Marshal(good)
	if err != nil || probeValidateReport(good) != nil {
		t.Fatal("fake report", err)
	}
	if _, err = probeDecode(raw); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*probeReport){
		func(r *probeReport) { r.Intervals = r.Intervals[:71] },
		func(r *probeReport) { r.Intervals[1] = r.Intervals[0] },
		func(r *probeReport) { r.Intervals[0].Counts.Returned++ },
		func(r *probeReport) { r.Totals.RankAttempts++ },
		func(r *probeReport) { r.Intervals[0].Calls[0].Ranking.CandidateIDs[0] = "lost" },
	} {
		r := fakeReport()
		mutate(&r)
		if probeValidateReport(r) == nil {
			t.Fatal("corrupt report")
		}
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	delete(fields, "fit")
	missing, _ := json.Marshal(fields)
	for _, bad := range [][]byte{
		bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1),
		append(append([]byte{}, raw[:len(raw)-1]...), []byte(`,"unknown":0}`)...),
		append(append([]byte{}, raw...), []byte(` {}`)...), missing,
		bytes.Replace(raw, []byte(`"fit":0`), []byte(`"fit":null`), 1),
		bytes.Replace(raw, []byte(`"qualified":false`), []byte(`"qualified":null`), 1),
		bytes.Replace(raw, []byte(`"order":[2,4,1,3,0,0,0,0]`), []byte(`"order":[2,4,1,3,0,0,0]`), 1),
	} {
		if _, err := probeDecode(bad); err == nil {
			t.Fatal("bad JSON")
		}
	}
}
