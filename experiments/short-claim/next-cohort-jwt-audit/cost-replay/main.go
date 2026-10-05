// Copyright 2026 SWYG. SPDX-License-Identifier: Apache-2.0
// Saved finite evidence replay only: no JWT dependency, signature verification,
// source execution, model, Fit, performance gate, or role admission.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"unicode/utf8"
)

const base = "experiments/short-claim/next-cohort-jwt-audit/"
const revision = "73c870b18e68b6e654b2b03f485aa3c9fab32cea"
const inputSHA = "47686828bbb0cd4d273db3b2087311357ea853b767f86fd787953ed76d452ed6"
const keySHA = "6c6ef54b2b0fa9af5bb3db8b8280a132371a3733adb5068e8ed40efabbc8c92d"
const limit = 192 << 10

type object map[string]json.RawMessage // JSON boundary only; simulation uses owned arrays.
type fault string

func need(ok bool, code string) {
	if !ok {
		panic(fault(code))
	}
}
func field[T any](o object, k string) T {
	var v T
	b, ok := o[k]
	need(ok && !bytes.Equal(bytes.TrimSpace(b), []byte("null")), "unknown_missing_or_null_field")
	need(json.Unmarshal(b, &v) == nil, "unknown_field_type")
	return v
}
func hash(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }

// Restricted frozen encoding: these four exact files have no Unicode escapes.
// Reject them instead of letting encoding/json repair unpaired UTF-16 surrogates.
func strict(b []byte) {
	need(len(b) <= limit && utf8.Valid(b) && !bytes.Contains(b, []byte(`\u`)), "unknown_encoding_or_bound")
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	nodes := 0
	var value func(int)
	value = func(depth int) {
		nodes++
		need(depth <= 16 && nodes <= 32768, "unknown_json_bound")
		t, err := d.Token()
		need(err == nil, "unknown_json_syntax")
		if s, ok := t.(string); ok {
			need(len(s) <= 4096, "unknown_json_string_bound")
		}
		if start, ok := t.(json.Delim); ok {
			need(start == '[' || start == '{', "unknown_json_syntax")
			var seen [64]string
			n := 0
			for d.More() {
				if start == '{' {
					k, e := d.Token()
					s, ok := k.(string)
					need(e == nil && ok && len(s) <= 128 && n < len(seen), "unknown_json_object_bound")
					for i := 0; i < n; i++ {
						need(seen[i] != s, "unknown_duplicate_key")
					}
					seen[n] = s
					n++
				}
				value(depth + 1)
			}
			end, e := d.Token()
			need(e == nil && ((start == '[' && end == json.Delim(']')) || (start == '{' && end == json.Delim('}'))), "unknown_json_syntax")
		}
	}
	value(0)
	_, err := d.Token()
	need(err == io.EOF, "unknown_trailing_json")
}
func read(path string, size int, sha string, packed bool, rawSize int, rawSHA string) object {
	f, e := os.Open(base + path)
	need(e == nil, "unknown_saved_file")
	b, e := io.ReadAll(io.LimitReader(f, limit+1))
	closeErr := f.Close()
	need(e == nil && closeErr == nil && len(b) == size && hash(b) == sha, "unknown_saved_pin")
	if packed {
		r := bytes.NewReader(b)
		z, e := gzip.NewReader(r)
		need(e == nil, "unknown_gzip")
		z.Multistream(false)
		b, e = io.ReadAll(io.LimitReader(z, limit+1))
		ce := z.Close()
		need(e == nil && ce == nil && r.Len() == 0 && len(b) == rawSize && hash(b) == rawSHA, "unknown_gzip_or_raw_pin")
	}
	strict(b)
	var o object
	need(json.Unmarshal(b, &o) == nil && o != nil, "unknown_object")
	return o
}

var columns = [17]string{"option_constructors", "option_applications", "direct_new_parser", "parse_with_claims", "new_validator", "implicit_parser_in_new_validator", "validate", "key_callbacks", "time_callbacks", "cache_lookups", "cache_key_comparisons", "cache_hits", "cache_misses", "deep_claims_copies", "public_key_copies", "candidate_checks", "fixture_checks"}
var clocks = [10]string{"setup_ns", "cache_lookup_ns", "claims_and_key_copy_ns", "cache_claims_store_ns", "control_order_construction_ns", "authentication_and_parse_ns", "fresh_policy_validate_ns", "combined_parse_signature_policy_ns", "outcome_qualification_ns", "complete_parent_work_ns"}
var errorsByBit = [8]string{"required_claim_missing", "invalid_audience", "expired", "used_before_issued", "invalid_issuer", "invalid_subject", "not_valid_yet", "invalid_claims_wrapper"}
var candidateIDs = [8]string{"option-leeway", "option-issued-at", "option-exp-required", "option-nbf-required", "option-audience-any", "option-audience-all", "option-issuer", "option-subject"}
var manualFirst = [4]int{4, 5, 2, 0} // Prospective request contract; not a learned prediction.
type counts [17]int64
type fact struct {
	Accept bool
	Mask   int
	SHA    string
}
type cacheKey [5]string

type simulation struct {
	Order    [8]int
	Checked  [8]bool
	Fixtures [8]int
	Matches  [8]bool
	Found    int
	Counts   counts
}

func expected(n, candidates, misses, hits, comparisons int, cached bool) counts {
	parse, validate := n, 0
	if cached {
		parse, validate = misses, n
	} else {
		misses, hits, comparisons = 0, 0, 0
	}
	lookup, copies := 0, 0
	if cached {
		lookup, copies = n, n+misses
	}
	return counts{int64(4 * (parse + validate)), int64(4 * (parse + validate)), int64(parse), int64(parse), int64(validate), int64(validate), int64(validate), int64(parse), int64(n), int64(lookup), int64(comparisons), int64(hits), int64(misses), int64(copies), int64(parse), int64(candidates), int64(n)}
}
func lookup(keys *[5]cacheKey, length *int, k cacheKey) (bool, int) {
	comparisons := 0
	for i := 0; i < *length; i++ {
		comparisons++
		if keys[i] == k {
			return true, comparisons
		}
	}
	need(*length < len(keys), "unknown_cache_capacity")
	keys[*length] = k
	*length++
	return false, comparisons
}
func simulate(got [5][8]bool, want [5]bool, keys [5]cacheKey, first int, cached bool) simulation {
	out := simulation{Found: -1}
	n := 0
	if first >= 0 {
		need(first < 8, "unknown_order")
		out.Order[n] = first
		n++
	}
	for c := 0; c < 8; c++ {
		if c != first {
			out.Order[n] = c
			n++
		}
	}
	var cache [5]cacheKey
	size, fixtures, candidates, misses, hits, comparisons := 0, 0, 0, 0, 0, 0
	for _, c := range out.Order {
		candidates++
		out.Checked[c] = true
		matches := true
		for f := 0; f < 5; f++ {
			fixtures++
			out.Fixtures[c]++
			if cached {
				hit, cmp := lookup(&cache, &size, keys[f])
				comparisons += cmp
				if hit {
					hits++
				} else {
					misses++
				}
			}
			if got[f][c] != want[f] {
				matches = false
				break
			}
		}
		out.Matches[c] = matches
		if matches {
			out.Found = c
			break
		}
	}
	out.Counts = expected(fixtures, candidates, misses, hits, comparisons, cached)
	return out
}
func cacheEquivalence(keys [5]cacheKey) counts {
	var cache [5]cacheKey
	size, misses, hits, comparisons := 0, 0, 0, 0
	for f := 0; f < 5; f++ {
		for c := 0; c < 8; c++ {
			hit, cmp := lookup(&cache, &size, keys[f])
			comparisons += cmp
			if hit {
				hits++
			} else {
				misses++
			}
		}
	}
	return expected(40, 0, misses, hits, comparisons, true)
}
func add(a *counts, b counts) {
	for i := range a {
		a[i] += b[i]
	}
}
func checkCounts(o object, want counts) {
	need(len(o) == len(columns), "unknown_counter_shape")
	for i, k := range columns {
		need(field[int64](o, k) == want[i], "unknown_counter_mismatch")
	}
}
func checkTiming(o object, cached, operation bool) {
	need(len(o) == len(clocks), "unknown_timing_shape")
	var v [10]int64
	for i, k := range clocks {
		b, ok := o[k]
		need(ok, "unknown_missing_timing")
		isNull := bytes.Equal(bytes.TrimSpace(b), []byte("null"))
		shouldNull := (!cached && (i == 5 || i == 6)) || (cached && i == 7) || (!operation && i == 9)
		need(isNull == shouldNull, "unknown_timing_scope")
		if !isNull {
			v[i] = field[int64](o, k)
			need(v[i] >= 0 && v[i] <= 30000000000, "unknown_timing_bound")
		}
	}
	need(v[0] > 0 && v[2] > 0 && v[8] > 0, "unknown_active_timing")
	if cached {
		need(v[1] > 0 && v[3] > 0 && v[5] > 0 && v[6] > 0, "unknown_active_timing")
	} else {
		need(v[1] == 0 && v[3] == 0 && v[7] > 0, "unknown_timing_scope")
	}
	if operation {
		var sum int64
		for i := 0; i < 9; i++ {
			sum += v[i]
		}
		// A measured short order construction can round to zero. Its field
		// remains required, non-null and nonnegative; complete work is positive.
		need(v[9] > 0 && sum <= v[9], "unknown_parent_timing")
	} else {
		need(v[4] == 0, "unknown_timing_scope")
	}
	// Additive phase totals fit their parent timer; no interval timestamps or causal claim.
}
func checkStats(o object, want counts, cached, operation bool) {
	checkCounts(field[object](o, "counts"), want)
	checkTiming(field[object](o, "timings"), cached, operation)
}
func checkMask(accept bool, mask int) {
	need(mask >= 0 && mask < 256 && ((accept && mask == 0) || (!accept && mask&128 != 0 && mask&127 != 0)), "unknown_error_qualification")
}
func result(o object, sha string) fact {
	need(field[bool](o, "known") && field[string](o, "code") == "observed", "unknown_result")
	f := fact{field[bool](o, "accept"), field[int](o, "policy_error_mask"), field[string](o, "canonical_claims_sha256")}
	checkMask(f.Accept, f.Mask)
	need(f.SHA == sha, "unknown_claims_identity")
	return f
}
func prior(o object, sha, canonical string) fact {
	need(field[string](o, "status") == "observed", "unknown_prior_status")
	for _, k := range []string{"token_present", "parse_returned", "payload_match", "claims_zero_before", "claims_input_unchanged", "key_input_unchanged"} {
		need(field[bool](o, k), "unknown_prior_qualification")
	}
	accept := field[bool](o, "valid")
	need(field[bool](o, "err_nil") == accept, "unknown_prior_flags")
	mask := 0
	names := field[[]string](o, "error_classes")
	for _, name := range names {
		bit := -1
		for i, known := range errorsByBit {
			if name == known || (i == 7 && name == "invalid_claims") {
				bit = i
			}
		}
		need(bit >= 0 && mask&(1<<bit) == 0, "unknown_prior_error")
		mask |= 1 << bit
	}
	checkMask(accept, mask)
	payload, e := json.Marshal(field[object](o, "decoded_canonical_payload"))
	need(e == nil && string(payload) == canonical && hash(payload) == sha, "unknown_prior_payload")
	calls := field[object](o, "calls")
	need(len(calls) == 6, "unknown_prior_call_shape")
	for i, k := range []string{"option_constructors", "option_applications", "new_parser", "parse_with_claims", "key_callbacks", "time_callbacks"} {
		want := 1
		if i < 2 {
			want = 4
		}
		need(field[int](calls, k) == want, "unknown_prior_calls")
	}
	return fact{accept, mask, sha}
}

type summary struct {
	Schema      string     `json:"schema"`
	Status      string     `json:"status"`
	Code        string     `json:"code"`
	Pairs       int        `json:"equal_pairs"`
	Runs        int        `json:"operational_runs"`
	Columns     [17]string `json:"counter_columns"`
	Equivalence counts     `json:"equivalence_counts"`
	Operations  counts     `json:"operational_counts"`
	Whole       counts     `json:"whole_work_counts"`
	NativeCalls int        `json:"new_jwt_api_calls"`
	ModelCalls  int        `json:"model_calls"`
	Scope       string     `json:"scope"`
}

func audit() summary {
	preparation := read("INPUTS.reviewed.v2.json", 23123, "9f050acc724284c34b57da70216a894633c2fa5d7d56a4dcdd9f3eb10b8db4d0", false, 0, "")
	input := read("OBSERVER-INPUT.actual.public.v1.json", 13474, inputSHA, false, 0, "")
	old := read("OBSERVATIONS.actual.public.v1.json.gz", 4273, "f080f9d1b6c0c50cf475086683b4879d3be3742832be6ec0914cefda356205f5", true, 116255, "b46d887b7f8e720a575cf307fabb11a3021b6259fa03b71b55c4655760d88b6f")
	current := read("cost-preview/OBSERVATIONS.actual.public.v1.json.gz", 5721, "0d3aa30350e9fb47a0857f31cdab3ba273901c5e326dae3adf60ff891a990bca", true, 125979, "e03b5a24788ed81d4bbc8f3d8e802525b79309c1d83c3212185f6dbca3877894")
	need(field[string](preparation, "schema") == "riido-JWT-input-preparation-review-v2" && field[string](preparation, "source_revision") == revision, "unknown_preparation_identity")
	need(field[string](input, "schema") == "riido-jwt-public-observer-input-v1" && field[int64](input, "clock_unix_seconds") == 1700000000 && field[string](input, "source_revision") == revision, "unknown_input_identity")
	for _, o := range []object{old, current} {
		need(field[string](o, "source_revision") == revision && field[string](o, "input_sha256") == inputSHA && field[string](o, "public_key_der_sha256") == keySHA && field[bool](o, "complete"), "unknown_saved_identity")
	}
	need(field[int](old, "distinct_fixture_inputs") == 13 && field[int](old, "fixture_positions") == 20, "unknown_prior_input_counts")
	oldCalls := field[object](old, "calls")
	need(len(oldCalls) == 6, "unknown_prior_call_shape")
	for i, k := range []string{"option_constructors", "option_applications", "new_parser", "parse_with_claims", "key_callbacks", "time_callbacks"} {
		want := 160
		if i < 2 {
			want = 640
		}
		need(field[int](oldCalls, k) == want, "unknown_prior_calls")
	}
	need(field[string](old, "schema") == "riido-jwt-direct-observer-output-v1" && field[bool](old, "input_unchanged") && field[int](old, "planned_trials") == 160, "unknown_prior_identity")
	need(field[string](current, "schema") == "riido-jwt-operational-cost-preview-v1" && field[bool](current, "input_unmodified") && field[string](current, "code") == "completed_finite_development_preview" && field[string](current, "source_family") == "jwt-go-lineage-whole-family-v1" && field[string](current, "role") == "development_validation", "unknown_current_identity")
	need(field[int](current, "own_init_calls") == 1 && !field[bool](current, "imported_package_init_count_observed") && field[bool](current, "main_timing_excludes_import_startup") && !field[bool](current, "rsa_verify_call_count_observed"), "unknown_startup_scope")
	for _, k := range []string{"signatures_generated", "key_generations", "model_calls", "fit_calls"} {
		need(field[int](current, k) == 0, "unknown_execution_scope")
	}
	need(field[int64](current, "main_elapsed_ns") > 0 && field[int64](current, "input_and_public_key_setup_ns") > 0, "unknown_main_timing")
	bitOrder := field[[]string](current, "normalized_error_classes_bit_order")
	need(len(bitOrder) == 8, "unknown_error_bit_order")
	for i, name := range bitOrder {
		need(name == errorsByBit[i], "unknown_error_bit_order")
	}
	prepared := field[[]object](preparation, "requests")
	requests := field[[]object](input, "requests")
	fixtures := field[[]object](input, "fixtures")
	previous := field[[]object](old, "records")
	need(len(prepared) == 4 && len(requests) == 4 && len(fixtures) == 13 && len(previous) == 160, "unknown_input_shape")
	var allFacts [4][5][8]fact
	var allWant [4][5]bool
	var allKeys [4][5]cacheKey
	for i, f := range fixtures {
		sha := field[string](f, "id")
		need(sha == field[string](f, "claims_sha256") && sha == hash([]byte(field[string](f, "claims_json"))) && field[string](f, "header_json") == `{"alg":"RS256","typ":"JWT"}`, "unknown_fixture_pin")
		for j := 0; j < i; j++ {
			need(field[string](fixtures[j], "id") != sha && field[string](fixtures[j], "token") != field[string](f, "token"), "unknown_fixture_duplicate")
		}
	}
	for p := 0; p < 4; p++ {
		name := field[string](prepared[p], "id")
		need(field[string](requests[p], "id") == name, "unknown_parent_identity")
		wants := field[[]object](prepared[p], "fixtures")
		ids := field[[]string](requests[p], "fixture_ids")
		need(len(wants) == 5 && len(ids) == 5, "unknown_parent_shape")
		for f := 0; f < 5; f++ {
			allWant[p][f] = field[bool](wants[f], "want_accept")
			canonical := field[string](wants[f], "claims_json")
			need(field[string](wants[f], "claims_sha256") == ids[f] && hash([]byte(canonical)) == ids[f], "unknown_want_binding")
			found := -1
			for j, fixture := range fixtures {
				if field[string](fixture, "id") == ids[f] {
					found = j
				}
			}
			need(found >= 0, "unknown_fixture_binding")
			need(field[string](fixtures[found], "claims_json") == canonical, "unknown_claims_binding")
			allKeys[p][f] = cacheKey{field[string](fixtures[found], "token"), keySHA, "RS256", "raw-url-base64-strict;RegisteredClaims;defaultJSON", revision}
			for c := 0; c < 8; c++ {
				o := previous[(p*5+f)*8+c]
				need(field[string](o, "request_id") == name && field[int](o, "fixture_position") == f && field[int](o, "candidate_position") == c && field[string](o, "candidate_id") == candidateIDs[c] && field[string](o, "fixture_id") == ids[f], "unknown_prior_order")
				allFacts[p][f][c] = prior(o, ids[f], canonical)
			}
		}
	}
	eq := field[object](current, "exhaustive_equivalence")
	pairs := field[[]object](eq, "records")
	refs := field[[]object](eq, "uncached_parent_stats")
	caches := field[[]object](eq, "cached_parent_stats")
	need(len(pairs) == 160 && len(refs) == 4 && len(caches) == 4 && field[int](eq, "planned_pairs") == 160 && field[int](eq, "attempted_pairs") == 160 && field[int](eq, "equal_pairs") == 160 && field[bool](eq, "all_equal"), "unknown_equivalence_shape")
	out := summary{Schema: "riido-jwt-saved-cost-replay-v1", Status: "pass", Code: "saved_finite_evidence_consistent", Pairs: 160, Runs: 48, Columns: columns, Scope: "Pinned saved public evidence only; exposed four-parent development diagnostic, not independent replication. Missing or contradictory evidence is unknown. Additive phase sums checked, not causal speed or interval-disjointness proof. No JWT execution, signatures, model, Fit or role admission."}
	for i, o := range pairs {
		p, f, c := i/40, (i/8)%5, i%8
		want := allFacts[p][f][c]
		need(field[int](o, "parent_position") == p && field[int](o, "fixture_position") == f && field[int](o, "candidate_position") == c && field[bool](o, "equal"), "unknown_pair_order")
		need(result(field[object](o, "uncached"), want.SHA) == want && result(field[object](o, "cached"), want.SHA) == want, "unknown_pair_parity")
	}
	for p := 0; p < 4; p++ {
		a, b := expected(40, 0, 0, 0, 0, false), cacheEquivalence(allKeys[p])
		checkStats(refs[p], a, false, false)
		checkStats(caches[p], b, true, false)
		add(&out.Equivalence, a)
		add(&out.Equivalence, b)
	}
	runs := field[[]object](current, "operational_runs")
	need(len(runs) == 48, "unknown_run_shape")
	for i, o := range runs {
		round, ci, mi, p := i/16, (i%16)/8, (i%8)/4, i%4
		manual, cached := (ci == 1) != (round == 1), (mi == 1) != (round == 1)
		control, first := "fixed-source-order", -1
		if manual {
			control, first = "manual-first", manualFirst[p]
		}
		var got [5][8]bool
		for f := 0; f < 5; f++ {
			for c := 0; c < 8; c++ {
				got[f][c] = allFacts[p][f][c].Accept
			}
		}
		want := simulate(got, allWant[p], allKeys[p], first, cached)
		need(field[int](o, "round") == round && field[int](o, "parent_position") == p && field[bool](o, "cached") == cached && field[string](o, "control") == control && field[bool](o, "complete"), "unknown_run_schedule")
		order := field[[]int](o, "candidate_order")
		checked := field[[]bool](o, "candidate_attempted")
		fc := field[[]int](o, "fixture_checks_by_candidate")
		matches := field[[]bool](o, "candidate_matches_exact_wanted")
		need(len(order) == 8 && len(checked) == 8 && len(fc) == 8 && len(matches) == 8, "unknown_candidate_shape")
		for c := 0; c < 8; c++ {
			need(order[c] == want.Order[c] && checked[c] == want.Checked[c] && fc[c] == want.Fixtures[c] && matches[c] == want.Matches[c], "unknown_failfast_mismatch")
		}
		need(want.Found >= 0 && field[int](o, "first_matching_candidate") == want.Found, "unknown_found_candidate")
		checkStats(field[object](o, "stats"), want.Counts, cached, true)
		add(&out.Operations, want.Counts)
	}
	out.Whole = out.Equivalence
	add(&out.Whole, out.Operations)
	return out
}
func main() {
	defer func() {
		if v := recover(); v != nil {
			code := "unknown_internal_replay"
			if f, ok := v.(fault); ok {
				code = string(f)
			}
			_ = json.NewEncoder(os.Stdout).Encode(summary{Schema: "riido-jwt-saved-cost-replay-v1", Status: "unknown", Code: code, Scope: "Saved evidence was not fully qualified; no candidate is relabeled and no native call or retry is made."})
			os.Exit(1)
		}
	}()
	need(len(os.Args) == 1, "unknown_arguments")
	_ = json.NewEncoder(os.Stdout).Encode(audit())
}
