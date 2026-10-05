// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// This is an authored Go shape calculator over the actual Output/14-column row
// types, never a worker run, observed result or ranking. Deliberately impossible
// simultaneous maxima conservatively cover bounded serialization paths.
func retrievalWorstOutputShape() Output {
	token := strings.Repeat("x", 64)
	hex64 := strings.Repeat("f", 64)
	ns := retrievalScalarNSMax
	var c Counts
	c = Counts{99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999, 99999}
	timing := Timings{SetupNS: ns, LookupNS: ns, CopyNS: ns, CacheStoreNS: ns, RankNS: ns, AuthParseNS: &ns, PolicyNS: &ns, CombinedNS: &ns, QualificationNS: ns, TotalNS: &ns}
	result := Result{Known: false, Accept: false, PolicyErrorMask: 65535, ClaimsSHA: hex64, Code: token}
	o := Output{Schema: token, SourceRevision: strings.Repeat("f", 40), InputSHA: hex64, TextInputSHA: hex64, PreparationSHA: hex64, PublicKeyDERSHA: hex64, Family: token, Role: token, Scope: strings.Repeat("x", 1024), OwnInit: 99999, Code: token, MainNS: ns, InputSetupNS: ns, ReferenceSetupNS: ns, OperationalParityChecks: 99999, TextRankCalls: 99999, SignaturesGenerated: 99999, KeyGenerations: 99999, ModelCalls: 99999, FitCalls: 99999}
	for _, list := range [][]string{o.PolicyMaskClasses[:], o.Controls[:], o.CounterColumns[:], o.TimingColumns[:], o.OperationalColumns[:], o.StateDictionary[:], o.UnknownDictionary[:]} {
		for i := range list {
			list[i] = token
		}
	}
	o.Equivalence = Equivalence{PlannedPairs: 160, AttemptedPairs: 160, EqualPairs: 160, AllEqual: false, Records: make([]Pair, 160)}
	for i := range o.Equivalence.Records {
		o.Equivalence.Records[i] = Pair{Parent: 3, Fixture: 4, Candidate: 7, Reference: result, Cached: result, Equal: false}
	}
	for i := 0; i < 4; i++ {
		o.Equivalence.Reference[i] = Stats{Counts: c, Timings: timing}
		o.Equivalence.Cached[i] = Stats{Counts: c, Timings: timing}
	}
	o.Runs = make([]retrievalOperationRow, 128)
	for i := range o.Runs {
		var order [8]int
		var scores [8]float64
		var checked [8]bool
		var fixtures, states, codes [8]int
		for j := 0; j < 8; j++ {
			order[j] = j
			scores[j] = -math.MaxFloat64
			fixtures[j] = 5
			states[j] = 3
			codes[j] = 31
		}
		o.Runs[i] = retrievalOperationRow{3, 3, false, 3, order, scores, checked, fixtures, states, codes, -1, false, retrievalCounts(c), retrievalTimings(timing)}
	}
	o.Failure = &retrievalFailure{Phase: token, Code: token, Round: -1, Control: -1, Parent: -1, Cached: false, Candidate: -1, Fixture: -1, Actual: &result}
	return o
}
func TestRetrievalActualGoCompleteWorstShapeCalculator(t *testing.T) {
	o := retrievalWorstOutputShape()
	raw, code := marshalQualifiedRetrievalOutput(o)
	if code != retrievalWireOK || len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatal("full actual typed shape did not qualify")
	}
	scalar, e := json.Marshal(-math.MaxFloat64)
	if e != nil || len(scalar) > 32 {
		t.Fatal("32-byte finite scalar premise failed")
	}
	// Replace the exemplar width with32 for every128x8 score; retain the separate
	// 4096-byte extension reserve even though actual new metadata is already here.
	padding := 128 * 8 * (32 - len(scalar))
	upper := len(raw) + padding + 4096
	if upper > retrievalStdoutLimit {
		t.Fatalf("actual shape requires a prospectively revised schema/reserve: %d > %d", upper, retrievalStdoutLimit)
	}
	sum := sha256.Sum256(raw)
	report := struct {
		Schema                                                                   string `json:"schema"`
		Rows, Pairs, Wire, ScoreScalar, Padding, Extension, Upper, Cap, Headroom int
		SHA                                                                      string `json:"synthetic_wire_sha256"`
		JWT, Rank, Model                                                         int
	}{Schema: "riido-retrieval-actual-go-shape-proof-v1", Rows: 128, Pairs: 160, Wire: len(raw), ScoreScalar: len(scalar), Padding: padding, Extension: 4096, Upper: upper, Cap: retrievalStdoutLimit, Headroom: retrievalStdoutLimit - upper, SHA: hex.EncodeToString(sum[:])}
	b, e := json.Marshal(report)
	if e != nil {
		t.Fatal("shape report")
	}
	t.Log(string(b))
}
func TestRetrievalOutputUnknownAndScalarBranches(t *testing.T) {
	o := retrievalWorstOutputShape()
	raw, code := marshalQualifiedRetrievalOutput(o)
	if code != retrievalWireOK || len(raw) == 0 {
		t.Fatal("sizing placeholder rejected")
	}
	for _, change := range []func(*Output){
		func(o *Output) { o.Code = "host/path" }, func(o *Output) { o.Scope = "x<y" }, func(o *Output) { o.Scope = "a&b" }, func(o *Output) { o.MainNS = -1 }, func(o *Output) { o.MainNS = retrievalScalarNSMax + 1 }, func(o *Output) { o.ModelCalls = 100000 },
		func(o *Output) {
			row := o.Runs[0]
			scores := row[5].([8]float64)
			scores[0] = math.NaN()
			row[5] = scores
			o.Runs[0] = row
		},
		func(o *Output) {
			row := o.Runs[0]
			order := row[4].([8]int)
			order[1] = order[0]
			row[4] = order
			o.Runs[0] = row
		},
		func(o *Output) { row := o.Runs[0]; row[12] = [16]int{}; o.Runs[0] = row },
	} {
		o := retrievalWorstOutputShape()
		change(&o)
		raw, code := marshalQualifiedRetrievalOutput(o)
		if code == retrievalWireOK || raw != nil {
			t.Fatal("unrepresentable output clipped or emitted")
		}
	}
	// Incomplete/unknown paths retain all earlier rows and the first actual result;
	// U fields are not omitted merely to make output fit.
	o = retrievalWorstOutputShape()
	o.Complete = false
	o.Failure.Actual.Known = false
	raw, code = marshalQualifiedRetrievalOutput(o)
	if code != retrievalWireOK || !json.Valid(raw) {
		t.Fatal("unknown response not representable")
	}
	var saved struct {
		Rows    []json.RawMessage `json:"operational_runs"`
		Failure *retrievalFailure `json:"first_failure"`
	}
	if json.Unmarshal(raw, &saved) != nil || len(saved.Rows) != 128 || saved.Failure == nil || saved.Failure.Actual == nil || saved.Failure.Actual.Known {
		t.Fatal("unknown traces dropped or changed")
	}
}
