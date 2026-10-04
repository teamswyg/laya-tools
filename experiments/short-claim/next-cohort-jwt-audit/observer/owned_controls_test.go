// SPDX-License-Identifier: Apache-2.0
// Synthetic protocol controls only. No original fixture, key, signature,
// ParserOption, NewParser, ParseWithClaims, model, Fit, label or role is used.
// Authoring does not execute imported-package init. A future Go test binary
// will initialize its imports; Root must freeze and count that separately.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
)

func requireStatus(t *testing.T, err error, want string, codes ...string) []string {
	t.Helper()
	got, status := normalizedErrors(err)
	if status != want {
		t.Fatalf("status = %q, want %q; codes = %q", status, want, got)
	}
	for _, code := range codes {
		found := false
		for _, observed := range got {
			if observed == code {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing code %q in %q", code, got)
		}
	}
	return got
}

func TestScalarAndMalformedProtocolControls(t *testing.T) {
	cases := []struct {
		name, wire string
		accept     bool
	}{
		{"plain", "{\"x\":\"synthetic\"}", true},
		{"raw_utf8", "{\"x\":\"한😀\"}", true},
		{"surrogate_pair", "\"\\ud83d\\ude00\"", true},
		{"literal_backslash_u", "\"\\\\ud800\"", true},
		{"escaped_quote", "\"a\\\"b\"", true},
		{"escaped_duplicate", "{\"a\":1,\"\\u0061\":2}", false},
		{"duplicate", "{\"x\":true,\"x\":false}", false},
		{"invalid_utf8", "\"\xff\"", false},
		{"truncated_utf8", "\"\xc3\"", false},
		{"lone_high", "\"\\ud800\"", false},
		{"lone_low", "\"\\udc00\"", false},
		{"high_non_low", "\"\\ud800\\u0041\"", false},
		{"pair_extra_low", "\"\\ud83d\\ude00\\udc00\"", false},
		{"broken_escape", "\"\\u12\"", false},
		{"raw_control", "\"\x00\"", false},
		{"truncated_object", "{\"x\":", false},
		{"trailing_value", "{} []", false},
		{"bad_escape", "\"\\q\"", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wire := []byte(c.wire)
			before := append([]byte(nil), wire...)
			if err := checkJSON(wire); (err == nil) != c.accept {
				t.Fatalf("accept = %v, want %v", err == nil, c.accept)
			}
			if !bytes.Equal(wire, before) {
				t.Fatal("validator mutated caller bytes")
			}
		})
	}
}

func TestProtocolLimitsAreAtomicAndInclusive(t *testing.T) {
	depthWire := func(n int) []byte { return []byte(strings.Repeat("[", n) + "0" + strings.Repeat("]", n)) }
	if checkJSON(depthWire(maxJSONDepth)) != nil {
		t.Fatal("maximum nesting rejected")
	}
	if checkJSON(depthWire(maxJSONDepth+1)) == nil {
		t.Fatal("excess nesting accepted")
	}
	arrayWire := func(n int) []byte { return []byte("[" + strings.TrimSuffix(strings.Repeat("null,", n), ",") + "]") }
	// One array value plus its children counts towards the documented node cap.
	if checkJSON(arrayWire(maxJSONNodes-1)) != nil {
		t.Fatal("maximum node count rejected")
	}
	if checkJSON(arrayWire(maxJSONNodes)) == nil {
		t.Fatal("excess node count accepted")
	}
	objectWire := func(n int) []byte {
		var b strings.Builder
		b.WriteByte('{')
		for i := 0; i < n; i++ {
			if i != 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Quote("k" + strconv.Itoa(i)))
			b.WriteString(":null")
		}
		b.WriteByte('}')
		return []byte(b.String())
	}
	if checkJSON(objectWire(32)) != nil {
		t.Fatal("maximum unique key count rejected")
	}
	if checkJSON(objectWire(33)) == nil {
		t.Fatal("excess unique key count accepted")
	}
	for _, c := range []struct {
		name, key string
		accept    bool
	}{
		{"ascii_key_limit", strings.Repeat("k", 64), true},
		{"ascii_key_excess", strings.Repeat("k", 65), false},
		{"utf8_key_byte_limit", strings.Repeat("é", 32), true},
		{"utf8_key_byte_excess", strings.Repeat("é", 33), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			wire := []byte("{" + strconv.Quote(c.key) + ":null}")
			if err := checkJSON(wire); (err == nil) != c.accept {
				t.Fatalf("accept = %v, want %v", err == nil, c.accept)
			}
		})
	}
	for _, c := range []struct {
		name, value string
		accept      bool
	}{
		{"ascii_limit", strings.Repeat("x", maxJSONString), true},
		{"ascii_excess", strings.Repeat("x", maxJSONString+1), false},
		{"utf8_byte_limit", strings.Repeat("é", maxJSONString/2), true},
		{"utf8_byte_excess", strings.Repeat("é", maxJSONString/2+1), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			wire := []byte(strconv.Quote(c.value))
			if err := checkJSON(wire); (err == nil) != c.accept {
				t.Fatalf("accept = %v, want %v", err == nil, c.accept)
			}
		})
	}
	// The public envelope rejects these before PKIX decoding or JWT calls.
	for _, wire := range [][]byte{
		[]byte("null"), []byte("[]"), []byte("{\"schema\":null}"),
		[]byte(strings.Repeat(" ", maxInputBytes+1)),
	} {
		_, key, der, err := validateInput(wire)
		if err == nil || key != nil || der != nil {
			t.Fatal("malformed/budgeted envelope produced usable input")
		}
	}
	if decoded, err := canonicalLiteral([]byte(strings.Repeat(" ", maxPayloadBytes+1))); err == nil || decoded != nil {
		t.Fatal("oversized literal produced decoded output")
	}
}

// Error stringification and custom Is are intentionally hostile. The bounded
// observer must retain an unknown leaf without invoking either method.
type nonStringifiableLeaf struct{}

func (nonStringifiableLeaf) Error() string { panic("synthetic Error must not run") }
func (nonStringifiableLeaf) Is(error) bool { panic("synthetic Is must not run") }

type syntheticWrapper struct{ child error }

func (syntheticWrapper) Error() string   { return "synthetic wrapper" }
func (e syntheticWrapper) Unwrap() error { return e.child }

type syntheticCycle struct{ next error }

func (*syntheticCycle) Error() string   { return "synthetic cycle" }
func (e *syntheticCycle) Unwrap() error { return e.next }

type syntheticUnwrapPanic struct{}

func (syntheticUnwrapPanic) Error() string { return "synthetic panic" }
func (syntheticUnwrapPanic) Unwrap() error { panic("synthetic unwrap") }

type syntheticChildren struct{ children []error }

func (syntheticChildren) Error() string     { return "synthetic children" }
func (e syntheticChildren) Unwrap() []error { return e.children }

func TestUnknownErrorPreservation(t *testing.T) {
	if codes := requireStatus(t, nil, "observed"); len(codes) != 0 {
		t.Fatal("nil error produced codes")
	}
	requireStatus(t, fmt.Errorf("synthetic outer: %w", jwt.ErrTokenExpired), "observed", "expired")
	requireStatus(t, errors.Join(jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired), "observed", "invalid_claims", "expired")
	requireStatus(t, jwt.ErrTokenInvalidClaims, "unknown_error", "wrapper_only_invalid_claims")
	// A recognized rejection does not conceal an unrelated unknown cause.
	requireStatus(t, errors.Join(jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired, nonStringifiableLeaf{}), "unknown_error", "expired", "unrecognized_leaf")
	requireStatus(t, nonStringifiableLeaf{}, "unknown_error", "unrecognized_leaf")
	requireStatus(t, syntheticWrapper{}, "unknown_error", "unrecognized_leaf")
	requireStatus(t, syntheticChildren{}, "unknown_error", "unrecognized_leaf")
	requireStatus(t, syntheticChildren{[]error{nil, jwt.ErrTokenExpired}}, "unknown_error", "expired", "unrecognized_leaf")
	requireStatus(t, errors.Join(jwt.ErrTokenExpired, context.DeadlineExceeded), "unknown_resource", "resource_context")
	requireStatus(t, jwt.ErrTokenSignatureInvalid, "unknown_crypto", "signature_invalid")
	requireStatus(t, jwt.ErrInvalidType, "unknown_input", "invalid_claim_type")
	requireStatus(t, jwt.ErrTokenMalformed, "unknown_input", "token_malformed")
	requireStatus(t, syntheticUnwrapPanic{}, "unknown_panic", "unwrap_panic")
	cycle := &syntheticCycle{}
	cycle.next = cycle
	requireStatus(t, cycle, "unknown_resource", "unwrap_budget")
}

func TestErrorTreeLimitsAndCallerOwnership(t *testing.T) {
	wrap := func(n int) error {
		var err error = jwt.ErrTokenExpired
		for i := 0; i < n; i++ {
			err = syntheticWrapper{err}
		}
		return err
	}
	requireStatus(t, wrap(8), "observed", "expired")
	requireStatus(t, wrap(9), "unknown_resource", "unwrap_budget")
	knownChildren := func(n int) error {
		a := make([]error, n)
		for i := range a {
			a[i] = jwt.ErrTokenExpired
		}
		return syntheticChildren{a}
	}
	// A multi-edge wrapper counts as one node, alongside all of its leaves.
	requireStatus(t, knownChildren(31), "observed", "expired")
	requireStatus(t, knownChildren(32), "unknown_resource", "unwrap_budget")
	requireStatus(t, knownChildren(33), "unknown_resource", "unwrap_budget")
	first := requireStatus(t, jwt.ErrTokenExpired, "observed", "expired")
	second := requireStatus(t, jwt.ErrTokenExpired, "observed", "expired")
	first[0] = "caller-mutated"
	if second[0] != "expired" {
		t.Fatal("normalizer shared caller-visible backing storage")
	}
	requireStatus(t, jwt.ErrTokenExpired, "observed", "expired")
	wire := []byte("{\"sub\":\"syn\",\"iss\":\"synth\"}")
	before := append([]byte(nil), wire...)
	canonical, err := canonicalLiteral(wire)
	if err != nil || string(canonical) != "{\"iss\":\"synth\",\"sub\":\"syn\"}" {
		t.Fatal("literal canonicalization lost synthetic fields")
	}
	canonical[0] = '!'
	if !bytes.Equal(wire, before) {
		t.Fatal("canonical output aliases caller bytes")
	}
	again, err := canonicalLiteral(wire)
	if err != nil || string(again) != "{\"iss\":\"synth\",\"sub\":\"syn\"}" {
		t.Fatal("caller mutation contaminated later result")
	}
}
