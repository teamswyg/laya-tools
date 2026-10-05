// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"testing"
)

// All fixtures here are original synthetic shapes, never JWT tokens or saved outcomes.
func obj(v any) object {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	var o object
	if e = json.Unmarshal(b, &o); e != nil {
		panic(e)
	}
	return o
}
func unknown(t *testing.T, fn func()) {
	t.Helper()
	rejected := false
	func() {
		defer func() {
			if v := recover(); v != nil {
				if _, ok := v.(fault); !ok {
					t.Fatalf("unexpected panic type: %T", v)
				}
				rejected = true
			}
		}()
		fn()
	}()
	if !rejected {
		t.Fatal("unqualified evidence was accepted")
	}
}
func qualified(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if v := recover(); v != nil {
			t.Fatalf("qualified synthetic evidence rejected: %v", v)
		}
	}()
	fn()
}
func synthetics() (object, object) {
	f := obj(map[string]any{"known": true, "accept": false, "policy_error_mask": 132, "canonical_claims_sha256": "synthetic-sha", "code": "observed"})
	p := obj(map[string]any{"status": "observed", "token_present": true, "parse_returned": true, "valid": false, "err_nil": false, "payload_match": true, "claims_zero_before": true, "claims_input_unchanged": true, "key_input_unchanged": true, "error_classes": []string{"invalid_claims", "expired"}, "decoded_canonical_payload": map[string]any{"aud": []string{"synthetic"}}, "calls": map[string]int{"option_constructors": 4, "option_applications": 4, "new_parser": 1, "parse_with_claims": 1, "key_callbacks": 1, "time_callbacks": 1}})
	return f, p
}
func TestMissingNullAndFalseRemainDistinct(t *testing.T) {
	f, _ := synthetics()
	qualified(t, func() {
		got := result(f, "synthetic-sha")
		if got.Accept || got.Mask != 132 {
			t.Fatal("known false lost")
		}
	})
	for _, k := range []string{"known", "accept", "policy_error_mask", "canonical_claims_sha256", "code"} {
		t.Run(k, func(t *testing.T) {
			f, _ := synthetics()
			delete(f, k)
			unknown(t, func() { result(f, "synthetic-sha") })
			f, _ = synthetics()
			f[k] = json.RawMessage("null")
			unknown(t, func() { result(f, "synthetic-sha") })
		})
	}
	f, _ = synthetics()
	f["known"] = json.RawMessage("false")
	unknown(t, func() { result(f, "synthetic-sha") })
	f, _ = synthetics()
	unknown(t, func() { result(f, "different-claims") })
}
func TestPolicyMaskQualification(t *testing.T) {
	for _, tc := range []struct {
		accept bool
		mask   int
	}{{true, 0}, {false, 129}, {false, 132}, {false, 192}, {false, 255}} {
		qualified(t, func() { checkMask(tc.accept, tc.mask) })
	}
	for _, tc := range []struct {
		accept bool
		mask   int
	}{{true, 128}, {true, 132}, {false, 0}, {false, 128}, {false, 4}, {false, 260}, {false, -1}} {
		unknown(t, func() { checkMask(tc.accept, tc.mask) })
	}
	f, _ := synthetics()
	f["policy_error_mask"] = json.RawMessage("1.5")
	unknown(t, func() { result(f, "synthetic-sha") })
}
func TestPriorQualificationPreservesUnknown(t *testing.T) {
	canonical := `{"aud":["synthetic"]}`
	sha := hash([]byte(canonical))
	_, p := synthetics()
	qualified(t, func() {
		f := prior(p, sha, canonical)
		if f.Accept || f.Mask != 132 {
			t.Fatal("prior false/mask")
		}
	})
	for _, k := range []string{"status", "valid", "err_nil", "token_present", "parse_returned", "payload_match", "claims_zero_before", "claims_input_unchanged", "key_input_unchanged", "error_classes", "decoded_canonical_payload", "calls"} {
		t.Run(k, func(t *testing.T) {
			_, p := synthetics()
			delete(p, k)
			unknown(t, func() { prior(p, sha, canonical) })
		})
	}
	for _, k := range []string{"token_present", "parse_returned", "payload_match", "claims_zero_before", "claims_input_unchanged", "key_input_unchanged"} {
		_, p = synthetics()
		p[k] = json.RawMessage("false")
		unknown(t, func() { prior(p, sha, canonical) })
	}
	for _, classes := range [][]string{{"invalid_claims"}, {"expired"}, {"invalid_claims", "expired", "alien"}, {"invalid_claims", "expired", "expired"}} {
		_, p = synthetics()
		b, _ := json.Marshal(classes)
		p["error_classes"] = b
		unknown(t, func() { prior(p, sha, canonical) })
	}
	_, p = synthetics()
	p["err_nil"] = json.RawMessage("true")
	unknown(t, func() { prior(p, sha, canonical) })
	_, p = synthetics()
	unknown(t, func() { prior(p, hash([]byte(`{"aud":[]}`)), `{"aud":[]}`) })
	_, p = synthetics()
	calls := field[object](p, "calls")
	delete(calls, "time_callbacks")
	p["calls"], _ = json.Marshal(calls)
	unknown(t, func() { prior(p, sha, canonical) })
}
func TestStrictFrozenJSONBoundary(t *testing.T) {
	qualified(t, func() { strict([]byte(`{"a":[true,false,null,12,{"b":"literal λ"}]}`)) })
	for _, s := range []string{`{"a":0,"a":1}`, `{"a":{"b":0,"b":1}}`, `{} {}`, `{"a":`, `{"a":"\ud800"}`, `{"a":"\u0061"}`, `{"a":NaN}`} {
		unknown(t, func() { strict([]byte(s)) })
	}
	unknown(t, func() { strict([]byte{'{', '"', 'a', '"', ':', '"', 0xff, '"', '}'}) })
	b := make([]byte, limit+1)
	unknown(t, func() { strict(b) })
	s := "0"
	for i := 0; i < 18; i++ {
		s = "[" + s + "]"
	}
	unknown(t, func() { strict([]byte(s)) })
}
func TestExactCounterPresenceAndNoExtras(t *testing.T) {
	want := expected(7, 3, 5, 2, 12, true)
	values := make(map[string]int64)
	for i, k := range columns {
		values[k] = want[i]
	}
	qualified(t, func() { checkCounts(obj(values), want) })
	values["cache_key_comparisons"]++
	unknown(t, func() { checkCounts(obj(values), want) })
	values["cache_key_comparisons"]--
	values["parse_with_claims"]++
	unknown(t, func() { checkCounts(obj(values), want) })
	values["parse_with_claims"]--
	o := obj(values)
	delete(o, "candidate_checks")
	unknown(t, func() { checkCounts(o, want) })
	o = obj(values)
	o["alien"] = json.RawMessage("0")
	unknown(t, func() { checkCounts(o, want) })
	o = obj(values)
	o["cache_hits"] = json.RawMessage("null")
	unknown(t, func() { checkCounts(o, want) })
	zero := expected(5, 1, 0, 0, 0, false)
	for i, k := range columns {
		values[k] = zero[i]
	}
	o = obj(values)
	delete(o, "cache_hits")
	unknown(t, func() { checkCounts(o, zero) })
}
func syntheticTiming(cached, operation bool) object {
	values := make(map[string]any)
	for _, k := range clocks {
		values[k] = int64(1)
	}
	if cached {
		values[clocks[7]] = nil
	} else {
		values[clocks[1]] = int64(0)
		values[clocks[3]] = int64(0)
		values[clocks[5]] = nil
		values[clocks[6]] = nil
	}
	if operation {
		values[clocks[9]] = int64(100)
	} else {
		values[clocks[4]] = int64(0)
		values[clocks[9]] = nil
	}
	return obj(values)
}
func TestTimingScopeNullsAndAdditiveParent(t *testing.T) {
	for _, cached := range []bool{false, true} {
		for _, operation := range []bool{false, true} {
			qualified(t, func() { checkTiming(syntheticTiming(cached, operation), cached, operation) })
		}
	}
	for _, k := range clocks {
		o := syntheticTiming(true, true)
		delete(o, k)
		unknown(t, func() { checkTiming(o, true, true) })
	}
	o := syntheticTiming(false, true)
	o[clocks[5]] = json.RawMessage("0")
	unknown(t, func() { checkTiming(o, false, true) })
	o = syntheticTiming(true, true)
	o[clocks[7]] = json.RawMessage("0")
	unknown(t, func() { checkTiming(o, true, true) })
	o = syntheticTiming(true, false)
	o[clocks[9]] = json.RawMessage("100")
	unknown(t, func() { checkTiming(o, true, false) })
	o = syntheticTiming(true, true)
	o[clocks[0]] = json.RawMessage("-1")
	unknown(t, func() { checkTiming(o, true, true) })
	o = syntheticTiming(true, true)
	o[clocks[5]] = json.RawMessage("0")
	unknown(t, func() { checkTiming(o, true, true) })
	o = syntheticTiming(true, true)
	o[clocks[9]] = json.RawMessage("1")
	unknown(t, func() { checkTiming(o, true, true) })
	o = syntheticTiming(true, true)
	o[clocks[0]] = json.RawMessage("30000000001")
	unknown(t, func() { checkTiming(o, true, true) })
}
func TestMeasuredZeroRankIsPresentWhileInvalidRankRemainsUnknown(t *testing.T) {
	for _, cached := range []bool{false, true} {
		o := syntheticTiming(cached, true)
		o[clocks[4]] = json.RawMessage("0")
		qualified(t, func() { checkTiming(o, cached, true) })
		o[clocks[4]] = json.RawMessage("-1")
		unknown(t, func() { checkTiming(o, cached, true) })
		o[clocks[4]] = json.RawMessage("null")
		unknown(t, func() { checkTiming(o, cached, true) })
		delete(o, clocks[4])
		unknown(t, func() { checkTiming(o, cached, true) })
	}
}
func TestEveryCacheIdentityDimensionAndCapacity(t *testing.T) {
	original := cacheKey{"synthetic-token", "synthetic-public-key", "synthetic-method", "synthetic-decoder", "synthetic-revision"}
	for dimension := 0; dimension < len(original); dimension++ {
		var store [5]cacheKey
		n := 0
		hit, cmp := lookup(&store, &n, original)
		if hit || cmp != 0 || n != 1 {
			t.Fatal("first miss")
		}
		changed := original
		changed[dimension] += "-other"
		hit, cmp = lookup(&store, &n, changed)
		if hit || cmp != 1 || n != 2 {
			t.Fatal("identity dimension collapsed")
		}
		hit, cmp = lookup(&store, &n, original)
		if !hit || cmp != 1 || n != 2 {
			t.Fatal("stable hit")
		}
		hit, cmp = lookup(&store, &n, changed)
		if !hit || cmp != 2 || n != 2 {
			t.Fatal("second stable hit")
		}
	}
	var store [5]cacheKey
	n := 0
	for i := 0; i < 5; i++ {
		k := original
		k[0] = string(rune('a' + i))
		lookup(&store, &n, k)
	}
	unknown(t, func() { lookup(&store, &n, cacheKey{"sixth"}) })
}
func TestFailFastAndStableManualOrderingWithoutRealFixtures(t *testing.T) {
	var got [5][8]bool
	var want [5]bool
	var keys [5]cacheKey
	for f := 0; f < 5; f++ {
		keys[f] = cacheKey{string(rune('a' + f)), "key", "method", "decode", "revision"}
		for c := 0; c < 8; c++ {
			got[f][c] = true
		}
	}
	// Candidate zero first disagrees at fixture 1; candidate one at fixture 2;
	// candidate two agrees with five literal false Wants. Others must remain unseen.
	got[0][0] = false
	got[0][1] = false
	got[1][1] = false
	for f := 0; f < 5; f++ {
		got[f][2] = false
	}
	fixed := simulate(got, want, keys, -1, true)
	if fixed.Found != 2 || fixed.Fixtures != [8]int{2, 3, 5} || fixed.Checked != [8]bool{true, true, true} || fixed.Matches != [8]bool{false, false, true} {
		t.Fatal("fail-fast trace")
	}
	if fixed.Counts[16] != 10 || fixed.Counts[15] != 3 || fixed.Counts[11] != 5 || fixed.Counts[12] != 5 || fixed.Counts[10] != 19 {
		t.Fatal("independent cache counts")
	}
	manual := simulate(got, want, keys, 2, true)
	if manual.Order != [8]int{2, 0, 1, 3, 4, 5, 6, 7} || manual.Fixtures != [8]int{0, 0, 5} || manual.Counts[11] != 0 || manual.Counts[12] != 5 || manual.Counts[10] != 10 {
		t.Fatal("manual cold trace")
	}
	plain := simulate(got, want, keys, -1, false)
	if plain.Fixtures != fixed.Fixtures || plain.Counts[3] != 10 || plain.Counts[6] != 0 || plain.Counts[10] != 0 {
		t.Fatal("uncached reference counts")
	}
	// No match is an explicit sentinel, not falsely forced to candidate zero.
	for f := 0; f < 5; f++ {
		got[f][2] = true
	}
	none := simulate(got, want, keys, -1, false)
	if none.Found != -1 || none.Counts[15] != 8 {
		t.Fatal("no answer lost")
	}
	unknown(t, func() { simulate(got, want, keys, 8, true) })
}
func TestEquivalenceColdCacheCountsAndOwnership(t *testing.T) {
	var keys [5]cacheKey
	for f := 0; f < 5; f++ {
		keys[f] = cacheKey{string(rune('a' + f)), "key", "method", "decode", "revision"}
	}
	before := keys
	got := cacheEquivalence(keys)
	if keys != before || got != expected(40, 0, 5, 35, 115, true) {
		t.Fatal("equivalence traversal/ownership")
	}
	shared := keys
	shared[4] = shared[0]
	got = cacheEquivalence(shared)
	if got[12] != 4 || got[11] != 36 || got[10] != 84 {
		t.Fatal("full token identity reuse")
	}
}
