// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import "math"

const retrievalScalarCounterMax = 99999
const retrievalScalarNSMax = int64(30000000000)

func retrievalSafeToken(s string, max int) bool {
	if len(s) == 0 || len(s) > max {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' && r != '.' && r != ':' {
			return false
		}
	}
	return true
}
func retrievalSafeHex(s string, n int, empty bool) bool {
	if empty && s == "" {
		return true
	}
	if len(s) != n {
		return false
	}
	for _, r := range s {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}
func retrievalSafeScope(s string) bool {
	if len(s) > 1024 {
		return false
	}
	for _, r := range s {
		if r < 32 || r > 126 || r == '"' || r == '\\' || r == '<' || r == '>' || r == '&' {
			return false
		}
	}
	return true
}
func retrievalBoundCounter(n int) bool { return n >= 0 && n <= retrievalScalarCounterMax }
func retrievalBoundNS(n int64) bool    { return n >= 0 && n <= retrievalScalarNSMax }
func retrievalBoundResult(r Result) bool {
	return retrievalSafeToken(r.Code, 64) && retrievalSafeHex(r.ClaimsSHA, 64, true)
}
func retrievalBoundStats(s Stats) bool {
	for _, n := range retrievalCounts(s.Counts) {
		if !retrievalBoundCounter(n) {
			return false
		}
	}
	for _, n := range retrievalTimings(s.Timings) {
		if n != nil && !retrievalBoundNS(*n) {
			return false
		}
	}
	return true
}
func retrievalBoundFailure(f *retrievalFailure) bool {
	if f == nil {
		return true
	}
	if !retrievalSafeToken(f.Phase, 64) || !retrievalSafeToken(f.Code, 64) || f.Round < -1 || f.Round > 3 || f.Control < -1 || f.Control > 3 || f.Parent < -1 || f.Parent > 3 || f.Candidate < -1 || f.Candidate > 7 || f.Fixture < -1 || f.Fixture > 4 {
		return false
	}
	return f.Actual == nil || retrievalBoundResult(*f.Actual)
}
func retrievalBoundRow(row retrievalOperationRow) bool {
	for _, i := range [3]int{0, 1, 3} {
		v, ok := row[i].(int)
		if !ok || v < 0 || v > 3 {
			return false
		}
	}
	if _, ok := row[2].(bool); !ok {
		return false
	}
	order, ok := row[4].([8]int)
	if !ok {
		return false
	}
	var seen [8]bool
	for _, v := range order {
		if v < 0 || v > 7 || seen[v] {
			return false
		}
		seen[v] = true
	}
	scores, ok := row[5].([8]float64)
	if !ok {
		return false
	}
	for _, v := range scores {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	if _, ok := row[6].([8]bool); !ok {
		return false
	}
	fixtures, ok := row[7].([8]int)
	if !ok {
		return false
	}
	for _, v := range fixtures {
		if v < 0 || v > 5 {
			return false
		}
	}
	states, ok := row[8].([8]int)
	if !ok {
		return false
	}
	for _, v := range states {
		if v < 0 || v > 3 {
			return false
		}
	}
	codes, ok := row[9].([8]int)
	if !ok {
		return false
	}
	for _, v := range codes {
		if v < 0 || v > 31 {
			return false
		}
	}
	found, ok := row[10].(int)
	if !ok || found < -1 || found > 7 {
		return false
	}
	if _, ok := row[11].(bool); !ok {
		return false
	}
	counts, ok := row[12].([17]int)
	if !ok {
		return false
	}
	for _, v := range counts {
		if !retrievalBoundCounter(v) {
			return false
		}
	}
	times, ok := row[13].([10]*int64)
	if !ok {
		return false
	}
	for _, v := range times {
		if v != nil && !retrievalBoundNS(*v) {
			return false
		}
	}
	return true
}

// Numeric and frozen safe-string guards precede writing; no clipping, omitted
// unknown rows, user-controlled MarshalJSON or nonfinite score replacement.
// Semantic qualification is separate: this guard proves only representability.
func retrievalOutputFits(o Output) bool {
	for _, s := range []string{o.Schema, o.Family, o.Role, o.Code} {
		if !retrievalSafeToken(s, 64) {
			return false
		}
	}
	if !retrievalSafeHex(o.SourceRevision, 40, false) || !retrievalSafeHex(o.InputSHA, 64, false) || !retrievalSafeHex(o.TextInputSHA, 64, false) || !retrievalSafeHex(o.PreparationSHA, 64, false) || !retrievalSafeHex(o.PublicKeyDERSHA, 64, true) || !retrievalSafeScope(o.Scope) {
		return false
	}
	for _, n := range []int{o.OwnInit, o.OperationalParityChecks, o.TextRankCalls, o.SignaturesGenerated, o.KeyGenerations, o.ModelCalls, o.FitCalls} {
		if !retrievalBoundCounter(n) {
			return false
		}
	}
	for _, n := range []int64{o.MainNS, o.InputSetupNS, o.ReferenceSetupNS} {
		if !retrievalBoundNS(n) {
			return false
		}
	}
	for _, list := range [][]string{o.PolicyMaskClasses[:], o.Controls[:], o.CounterColumns[:], o.TimingColumns[:], o.OperationalColumns[:], o.StateDictionary[:], o.UnknownDictionary[:]} {
		for _, s := range list {
			if !retrievalSafeToken(s, 64) {
				return false
			}
		}
	}
	eq := o.Equivalence
	for _, n := range []int{eq.PlannedPairs, eq.AttemptedPairs, eq.EqualPairs} {
		if n < 0 || n > 160 {
			return false
		}
	}
	if len(eq.Records) > 160 || len(o.Runs) > 128 || !retrievalBoundFailure(o.Failure) {
		return false
	}
	for _, p := range eq.Records {
		if p.Parent < 0 || p.Parent > 3 || p.Fixture < 0 || p.Fixture > 4 || p.Candidate < 0 || p.Candidate > 7 || !retrievalBoundResult(p.Reference) || !retrievalBoundResult(p.Cached) {
			return false
		}
	}
	for _, s := range eq.Reference {
		if !retrievalBoundStats(s) {
			return false
		}
	}
	for _, s := range eq.Cached {
		if !retrievalBoundStats(s) {
			return false
		}
	}
	for _, row := range o.Runs {
		if !retrievalBoundRow(row) {
			return false
		}
	}
	return true
}
func marshalQualifiedRetrievalOutput(o Output) ([]byte, retrievalWireCode) {
	if !retrievalOutputFits(o) {
		return nil, retrievalWireCode("unknown_output_scalar")
	}
	return marshalRetrievalBounded(o)
}
