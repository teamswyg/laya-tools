package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func protocolPrepared(t *testing.T, request string, texts [3]string) shortclaim.Prepared {
	t.Helper()
	ids := candidateIDs()
	in := shortclaim.Input{Schema: textSchema, Request: request, Provenance: "unit-vector", Candidates: make([]shortclaim.Candidate, 3)}
	for i := range texts {
		in.Candidates[i] = shortclaim.Candidate{ID: ids[i], Text: texts[i]}
	}
	p, e := shortclaim.Validate(in)
	if e != nil {
		t.Fatal(e)
	}
	return p
}

func TestProtocolLiteralDigestVectors(t *testing.T) {
	p := protocolPrepared(t, "A,B", [3]string{"ab", "a b", "b a"})
	// Literal bytes were independently encoded and checked with SHA-256. These
	// vectors do not call writeLengthText, featureSHA, or inputDigest to produce
	// their expected messages, and do not use any source IDs or truth labels.
	vectors := []struct {
		name, messageHex, want, got string
	}{
		{"features", "000000000000001a726969646f2d7265736964656e742d66656174757265732d7631000000000000000300000000000000036120620000000000000002616200000000000000036120620000000000000003622061", "2429291077f9101f2ee4a98b480a7e6841afe36b35a504469d890167bdc33127", featureSHA(p)},
		{"input", "000000000000001d726969646f2d73686f72742d6265686176696f722d636c61696d2d76310000000000000003412c42000000000000000b756e69742d766563746f72000000000000000b63616e6469646174652d3000000000000000026162000000000000000b63616e6469646174652d310000000000000003612062000000000000000b63616e6469646174652d320000000000000003622061", "e8cf4bcd6768cdaff29a9ea3671e9392482dd4bcf1fe1aedf56e3d90ea4859b0", inputDigest(p)},
	}
	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			raw, e := hex.DecodeString(v.messageHex)
			if e != nil {
				t.Fatal(e)
			}
			sum := sha256.Sum256(raw)
			if hex.EncodeToString(sum[:]) != v.want || v.got != v.want {
				t.Fatalf("literal digest disagreement: %s", v.name)
			}
		})
	}
	metadata := p
	metadata.Provenance, metadata.Candidates[0].ID = "other-provenance", "other-id"
	if featureSHA(metadata) != featureSHA(p) || inputDigest(metadata) == inputDigest(p) {
		t.Fatal("metadata leaked into features or disappeared from wire digest")
	}
	whitespace := protocolPrepared(t, "  A,B  ", [3]string{" ab ", "a  b", "b a"})
	if featureSHA(whitespace) != featureSHA(p) || inputDigest(whitespace) == inputDigest(p) {
		t.Fatal("raw and normalized identities were conflated")
	}
	ordered := p
	ordered.Candidates[0], ordered.Candidates[1] = ordered.Candidates[1], ordered.Candidates[0]
	if featureSHA(ordered) == featureSHA(p) || inputDigest(ordered) == inputDigest(p) {
		t.Fatal("candidate order is not bound")
	}
	changed := protocolPrepared(t, "A,B", [3]string{"abc", "a b", "b a"})
	if featureSHA(changed) == featureSHA(p) {
		t.Fatal("actual candidate text change disappeared")
	}
	shorter := p
	shorter.Count, shorter.Candidates[2] = 2, shortclaim.PreparedCandidate{}
	if shortclaim.ValidatePrepared(shorter) != nil || featureSHA(shorter) == featureSHA(p) {
		t.Fatal("candidate count is not bound")
	}
	left := protocolPrepared(t, "request", [3]string{"a b", "c", "d"})
	right := protocolPrepared(t, "request", [3]string{"a", "b c", "d"})
	if featureSHA(left) == featureSHA(right) {
		t.Fatal("length-prefixed ordered texts lost their boundaries")
	}
	// Normalization drops raw punctuation. Such identity is preparation's pool
	// selection policy, not a safe semantic/model/rule-response cache key.
	quoted := protocolPrepared(t, `"a,b",c`, [3]string{"ab", "a b", "b a"})
	escaped := protocolPrepared(t, `a\,b,c`, [3]string{"ab", "a b", "b a"})
	if featureSHA(quoted) != featureSHA(escaped) || inputDigest(quoted) == inputDigest(escaped) {
		t.Fatal("punctuation-loss boundary was not represented")
	}
}

func TestStrictJSONBoundaries(t *testing.T) {
	type shape struct {
		Name   string    `json:"name"`
		Values []float64 `json:"values"`
	}
	valid := []string{`{"name":"plain","values":[0,1.25]}`, `{"name":"\ud83d\ude00"}`, `{"name":"\uFFFD"}`, `{"name":"\\ud800"}`, `{"n\u0061me":"ok"}`}
	for _, raw := range valid {
		var got shape
		if e := strictJSON([]byte(raw), &got); e != nil {
			t.Fatalf("valid scalar/name rejected: %s", raw)
		}
	}
	invalid := []string{
		``, `[]`, `null`, `{"name":"one"} {"name":"two"}`,
		`{"Name":"casefold"}`, `{"NAME":"casefold"}`, `{"name":"one","name":"two"}`,
		`{"name":"one","n\u0061me":"two"}`, `{"unknown":"value"}`,
		`{"name":"\ud800"}`, `{"name":"\udfff"}`, `{"name":"\ud800\u0041"}`,
		`{"name":"\ud800\\udc00"}`, `{"name":"\udc00\ud800"}`,
		`{"name":"x","values":[1e999]}`, `{"name":"x","values":[NaN]}`,
		`{"name":"x","values":[{"score":1,"sco\u0072e":2}]}`,
		`{"name":"x","values":` + strings.Repeat("[", 17) + `0` + strings.Repeat("]", 17) + `}`,
		string([]byte{'{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 0xff, '"', '}'}),
	}
	for i, raw := range invalid {
		var got shape
		if e := strictJSON([]byte(raw), &got); e == nil {
			t.Fatalf("invalid case %d was accepted", i)
		}
	}
}

func TestExpectedResponseAndExactBytes(t *testing.T) {
	p := protocolPrepared(t, "keep active rows", [3]string{"keep active rows", "remove active rows", "keep expired rows"})
	payload := Payload{InputDigest: inputDigest(p)}
	for _, kind := range baselineKinds() {
		raw, expected, e := expectedResponse(p, kind)
		if e != nil || checkResponse(raw, payload, kind, expected) != nil {
			t.Fatalf("deterministic protocol expectation failed for %s", kind)
		}
		if expected.Baseline != kind || expected.RawBytes != len(raw) || expected.RawSHA256 != hashBytes(raw) || bytes.Count(raw, []byte{'\n'}) != 1 {
			t.Fatal("response byte contract not retained")
		}
		if e := checkResponse(append([]byte{' '}, raw...), payload, kind, expected); e == nil {
			t.Fatal("a differently serialized response escaped the exact byte pin")
		}
	}
	if _, got, e := expectedResponse(p, "unsupported"); e == nil || got != (ExpectedResponse{}) {
		t.Fatal("unknown baseline manufactured an expectation")
	}
	for _, sample := range []struct {
		request  string
		texts    [3]string
		fallback string
	}{
		{"plain request", [3]string{"keep rows", "remove rows", "keep active rows"}, "unsupported_rule_request"},
		{"behavior-v1: if active then keep else remove", [3]string{"plain candidate", "behavior-v1: if active then keep else remove", "behavior-v1: if active then remove else keep"}, "unsupported_rule_candidate"},
		{"behavior-v1: if active then keep else remove", [3]string{"behavior-v1: if active then keep else remove", "behavior-v1: if active then remove else keep", "behavior-v1: if not active then remove else keep"}, ""},
	} {
		p := protocolPrepared(t, sample.request, sample.texts)
		raw, expectation, e := expectedResponse(p, "narrow_rule")
		if e != nil || expectation.FallbackReason != sample.fallback || checkResponse(raw, Payload{InputDigest: inputDigest(p)}, "narrow_rule", expectation) != nil {
			t.Fatal("explicit narrow-rule fallback contract failed")
		}
	}
}

func TestResponseRejectsMalformedContractEvenWithMatchingHash(t *testing.T) {
	p := protocolPrepared(t, "active keep", [3]string{"keep active", "remove active", "keep expired"})
	payload := Payload{InputDigest: inputDigest(p)}
	raw, expected, e := expectedResponse(p, "fixed_order")
	if e != nil {
		t.Fatal(e)
	}
	// Rebinding the raw hash below is intentional: a valid byte pin does not
	// excuse structural/semantic protocol faults, including missing zero scores.
	replace := func(old, next string) []byte { return bytes.Replace(raw, []byte(old), []byte(next), 1) }
	invalid := [][]byte{
		replace(`"schema":`, `"Schema":`),
		replace(`"schema":`, `"extra":"x","schema":`),
		replace(`"status":"unverified_heuristic"`, `"status":"approved"`),
		replace(`"schema":"`+responseSchema+`"`, `"schema":null`),
		replace(`"schema":"`+responseSchema+`",`, ``),
		replace(`"baseline":"fixed_order"`, `"baseline":"bm25"`),
		replace(payload.InputDigest, strings.Repeat("0", 64)),
		replace(`"candidate-1"`, `"candidate-0"`),
		replace(`"candidate-1"`, `"unknown-id"`),
		replace(`"id":"candidate-0"`, `"id":null`),
		replace(`"id":"candidate-0"`, `"id":"\ud800"`),
		replace(`"score":0`, `"score":null`),
		replace(`,"score":0`, ``),
		replace(`"score":0`, `"score":1e999`),
		replace(`"score":0`, `"score":0,"sco\u0072e":0`),
		replace(`"verification_order":`, `"fallback_reason":null,"verification_order":`),
		replace(`"verification_order":`, `"fallback_reason":"unsupported_rule_request","verification_order":`),
		replace(`"verification_order":`, `"fallback_reason":"unknown","verification_order":`),
		replace(`"verification_order":`, `"baseline":"fixed_order","verification_order":`),
		replace(`"verification_order":`, `"base\u006cine":"fixed_order","verification_order":`),
		append(slices.Clone(raw), '\n'),
		append(slices.Clone(raw[:len(raw)-1]), '\r', '\n'),
		replace(`,"status"`, ",\n\"status\""),
		slices.Clone(raw[:len(raw)-1]),
		[]byte(strings.Repeat(" ", wireLimit) + "\n"),
	}
	for i, bad := range invalid {
		pin := expected
		pin.RawSHA256, pin.RawBytes = hashBytes(bad), len(bad)
		if e := checkResponse(bad, payload, "fixed_order", pin); e == nil {
			t.Fatalf("malformed response case %d passed", i)
		}
	}
	for _, changed := range []ExpectedResponse{
		{Baseline: "bm25", RawSHA256: expected.RawSHA256, RawBytes: expected.RawBytes},
		{Baseline: "fixed_order", RawSHA256: "bad", RawBytes: expected.RawBytes},
		{Baseline: "fixed_order", RawSHA256: expected.RawSHA256, RawBytes: 0},
	} {
		if checkResponse(raw, payload, "fixed_order", changed) == nil {
			t.Fatal("unbound response expectation passed")
		}
	}
	if checkResponse(raw, Payload{}, "fixed_order", expected) == nil || checkResponse(raw, payload, "unknown", expected) == nil {
		t.Fatal("missing digest or unknown baseline passed")
	}
	permutation := bytes.ReplaceAll(raw, []byte("candidate-0"), []byte("candidate-temp"))
	permutation = bytes.ReplaceAll(permutation, []byte("candidate-1"), []byte("candidate-0"))
	permutation = bytes.ReplaceAll(permutation, []byte("candidate-temp"), []byte("candidate-1"))
	if checkResponse(permutation, payload, "fixed_order", expected) == nil {
		t.Fatal("different full permutation escaped the exact expectation pin")
	}
	rebound := expected
	rebound.RawSHA256, rebound.RawBytes = hashBytes(permutation), len(permutation)
	if checkResponse(permutation, payload, "fixed_order", rebound) != nil {
		t.Fatal("structural permutation contract incorrectly assumes semantic ordering")
	}
}

func TestProtocolSmallHelpers(t *testing.T) {
	for _, valid := range []string{strings.Repeat("0", 64), strings.Repeat("a", 64), strings.Repeat("f", 64)} {
		if !validSHA(valid) {
			t.Fatal("valid SHA rejected")
		}
	}
	for _, invalid := range []string{"", strings.Repeat("0", 63), strings.Repeat("F", 64), strings.Repeat("g", 64), strings.Repeat("/", 64), strings.Repeat(":", 64)} {
		if validSHA(invalid) {
			t.Fatal("invalid SHA accepted")
		}
	}
	if words("") != 0 || words("one two three") != 3 || distribution(nil) != (Summary{}) {
		t.Fatal("empty/count helper contract failed")
	}
	input := []int64{4, 1, 3, 2}
	if got := distribution(input); got != (Summary{Count: 4, P50NS: 2, P95NS: 4, MaxNS: 4}) || !slices.Equal(input, []int64{4, 1, 3, 2}) {
		t.Fatal("nearest-rank summary changed caller-owned samples")
	}
}
