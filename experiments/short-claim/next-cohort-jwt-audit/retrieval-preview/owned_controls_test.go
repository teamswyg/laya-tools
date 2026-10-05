// Synthetic owned controls only. Authoring is not test/startup authorization.
// No real tokens, fixture outcomes, keys, signing, Parse or Validate calls.
package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestClonePreservesAbsentAndEmptyValues(t *testing.T) {
	empty := jwt.RegisteredClaims{}
	copy := cloneClaims(empty)
	if copy.Audience != nil || copy.ExpiresAt != nil || copy.NotBefore != nil || copy.IssuedAt != nil {
		t.Fatal("absent policy fields became present")
	}
	original := jwt.RegisteredClaims{Audience: jwt.ClaimStrings{}}
	cloned := cloneClaims(original)
	if cloned.Audience == nil || len(cloned.Audience) != 0 {
		t.Fatal("present empty audience lost its identity")
	}
	if original.Audience == nil {
		t.Fatal("caller changed")
	}
}

func TestCloneSeparatesAllMutableOwnership(t *testing.T) {
	shared := &jwt.NumericDate{Time: time.Unix(111, 0).UTC()}
	before := &jwt.NumericDate{Time: time.Unix(222, 0).UTC()}
	audience := make(jwt.ClaimStrings, 2, 4)
	audience[0], audience[1] = "red", "blue"
	source := jwt.RegisteredClaims{Issuer: "issuer", Subject: "subject", ID: "id",
		Audience: audience, ExpiresAt: shared, IssuedAt: shared, NotBefore: before}
	copy := cloneClaims(source)
	if copy.Issuer != source.Issuer || copy.Subject != source.Subject || copy.ID != source.ID {
		t.Fatal("immutable value fields changed")
	}
	if copy.ExpiresAt == shared || copy.IssuedAt == shared || copy.NotBefore == before || copy.ExpiresAt == copy.IssuedAt {
		t.Fatal("numeric dates retain source or cross-field aliases")
	}
	copy.Audience[0] = "copy-red"
	copy.ExpiresAt.Time = time.Unix(333, 0).UTC()
	copy.NotBefore.Time = time.Unix(444, 0).UTC()
	if source.Audience[0] != "red" || source.ExpiresAt.Unix() != 111 || source.NotBefore.Unix() != 222 || copy.IssuedAt.Unix() != 111 {
		t.Fatal("copy mutation escaped its owned values")
	}
	source.Audience[1] = "source-blue"
	source.IssuedAt.Time = time.Unix(555, 0).UTC()
	if copy.Audience[1] != "blue" || copy.IssuedAt.Unix() != 111 || copy.ExpiresAt.Unix() != 333 {
		t.Fatal("source mutation escaped into independent copies")
	}
	source.Audience = append(source.Audience, "source-append")
	copy.Audience = append(copy.Audience, "copy-append")
	if source.Audience[2] != "source-append" || copy.Audience[2] != "copy-append" {
		t.Fatal("append reused caller-owned audience capacity")
	}
}

func TestPublicKeyCopyOwnsModulus(t *testing.T) {
	// Deliberately tiny synthetic integers, never used with an RSA operation.
	source := &rsa.PublicKey{N: big.NewInt(12345), E: 17}
	stats := newStats(false)
	copy := keyCopy(source, &stats)
	if copy == source || copy.N == source.N || copy.E != 17 || copy.N.Cmp(source.N) != 0 {
		t.Fatal("key copy is not an equal independently owned value")
	}
	copy.N.SetInt64(9)
	if source.N.Int64() != 12345 {
		t.Fatal("copied modulus aliases caller")
	}
	source.N.SetInt64(8)
	if copy.N.Int64() != 9 {
		t.Fatal("caller mutation reaches copied modulus")
	}
	if stats.Counts.PublicKeyCopies != 1 || stats.Timings.CopyNS < 0 {
		t.Fatal("copy accounting was not retained")
	}
}

func TestEveryCacheIdentityDimensionSeparatesContexts(t *testing.T) {
	base := CacheKey{Token: "synthetic.token.a", DER: "\x00der-context-a", Method: "RS256",
		Decoder: "strict-decoder-a", Source: "source-a"}
	tests := []struct {
		name   string
		change func(*CacheKey)
	}{
		{"token", func(k *CacheKey) { k.Token = "synthetic.token.b" }},
		{"der_tail", func(k *CacheKey) { k.DER = "\x00der-context-b" }},
		{"method", func(k *CacheKey) { k.Method = "RS512" }},
		{"decoder", func(k *CacheKey) { k.Decoder = "strict-decoder-b" }},
		{"source", func(k *CacheKey) { k.Source = "source-b" }},
	}
	var cache Cache
	cache.Entries[0].Key = base
	cache.Used = 1
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			other := base
			tt.change(&other)
			if cache.Entries[0].Key == other {
				t.Fatal("different authenticated context collided")
			}
			if cache.Entries[0].Key != base {
				t.Fatal("caller-owned context changed")
			}
		})
	}
	anotherParent := Cache{}
	if anotherParent.Used != 0 || anotherParent.Entries[0].Key == base {
		t.Fatal("a fresh parent inherits cache state")
	}
}

type syntheticSingle struct{ child error }

func (*syntheticSingle) Error() string   { return "synthetic wrapper" }
func (s *syntheticSingle) Unwrap() error { return s.child }

type syntheticMulti struct{ children []error }

func (*syntheticMulti) Error() string     { return "synthetic branches" }
func (m *syntheticMulti) Unwrap() []error { return m.children }

type syntheticHostileLeaf struct{ calls *int }

func (h *syntheticHostileLeaf) Error() string { (*h.calls)++; panic("Error must not run") }
func (h *syntheticHostileLeaf) Is(error) bool { (*h.calls)++; return true }

type syntheticCycle struct{}

func (*syntheticCycle) Error() string   { return "synthetic cycle" }
func (c *syntheticCycle) Unwrap() error { return c }

type syntheticPanic struct{}

func (*syntheticPanic) Error() string { return "synthetic panic wrapper" }
func (*syntheticPanic) Unwrap() error { panic("synthetic unwrap panic") }

func requirePolicy(t *testing.T, err error, accept bool, mask uint16) {
	t.Helper()
	known, gotAccept, gotMask := policyStatus(err)
	if !known || gotAccept != accept || gotMask != mask {
		t.Fatalf("policy facts differ: known=%v accept=%v mask=%d", known, gotAccept, gotMask)
	}
}
func requireUnknown(t *testing.T, err error) {
	t.Helper()
	known, accept, _ := policyStatus(err)
	if known || accept {
		t.Fatal("unqualified error became observed Boolean")
	}
}

func TestPolicySentinelsAndWrapperParity(t *testing.T) {
	requirePolicy(t, nil, true, 0)
	// Bit identities are output-schema facts, independently asserted here.
	cases := []struct {
		err  error
		mask uint16
	}{
		{jwt.ErrTokenRequiredClaimMissing, 1},
		{jwt.ErrTokenInvalidAudience, 2},
		{jwt.ErrTokenExpired, 4},
		{jwt.ErrTokenUsedBeforeIssued, 8},
		{jwt.ErrTokenInvalidIssuer, 16},
		{jwt.ErrTokenInvalidSubject, 32},
		{jwt.ErrTokenNotValidYet, 64},
	}
	for _, tt := range cases {
		requirePolicy(t, tt.err, false, tt.mask)
		joined := errors.Join(jwt.ErrTokenInvalidClaims, tt.err)
		requirePolicy(t, joined, false, tt.mask|128)
		requirePolicy(t, fmt.Errorf("synthetic context: %w", joined), false, tt.mask|128)
	}
	requirePolicy(t, errors.Join(jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired, jwt.ErrTokenInvalidAudience), false, 134)
	requireUnknown(t, jwt.ErrTokenInvalidClaims)
	requireUnknown(t, fmt.Errorf("generic only: %w", jwt.ErrTokenInvalidClaims))
}

func TestUnknownCausesCannotBorrowRecognizedRefusal(t *testing.T) {
	calls := 0
	hostile := &syntheticHostileLeaf{calls: &calls}
	cases := []error{
		hostile,
		errors.Join(jwt.ErrTokenExpired, hostile),
		&syntheticMulti{children: []error{jwt.ErrTokenExpired, nil}},
		&syntheticMulti{},
		&syntheticSingle{},
		errors.Join(jwt.ErrTokenInvalidClaims, context.Canceled),
		errors.Join(jwt.ErrTokenExpired, context.DeadlineExceeded),
		jwt.ErrTokenSignatureInvalid,
		errors.Join(jwt.ErrTokenExpired, jwt.ErrTokenMalformed),
		&syntheticCycle{},
		errors.Join(jwt.ErrTokenExpired, &syntheticPanic{}),
	}
	for i, err := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) { requireUnknown(t, err) })
	}
	if calls != 0 {
		t.Fatal("normalization executed arbitrary Error/Is behavior")
	}
}

func TestUnwrapBudgetsAreInclusiveAndFailUnknown(t *testing.T) {
	depth8 := error(jwt.ErrTokenExpired)
	for i := 0; i < 8; i++ {
		depth8 = &syntheticSingle{child: depth8}
	}
	requirePolicy(t, depth8, false, 4)
	requireUnknown(t, &syntheticSingle{child: depth8})
	// One branch node plus 31 known leaves is exactly 32 visited nodes.
	children := make([]error, 31)
	for i := range children {
		children[i] = jwt.ErrTokenExpired
	}
	requirePolicy(t, &syntheticMulti{children: children}, false, 4)
	children = append(children, jwt.ErrTokenExpired)
	requireUnknown(t, &syntheticMulti{children: children})
	requireUnknown(t, &syntheticMulti{children: append(children, jwt.ErrTokenExpired)})
}

func TestCanonicalClaimsStayOwnedAndRejectUnsupportedPrecision(t *testing.T) {
	audience := jwt.ClaimStrings{"red", "blue"}
	source := jwt.RegisteredClaims{Issuer: "synthetic", Audience: audience,
		ExpiresAt: &jwt.NumericDate{Time: time.Unix(100, 0).UTC()}}
	first, err := claimsHash(source)
	if err != nil {
		t.Fatal("valid synthetic claims rejected")
	}
	copied := cloneClaims(source)
	second, err := claimsHash(copied)
	if err != nil || first != second {
		t.Fatal("copy lost claim meaning")
	}
	copied.Audience[0] = "changed"
	third, err := claimsHash(source)
	if err != nil || third != first || source.Audience[0] != "red" {
		t.Fatal("canonical qualification mutated source")
	}
	fractional := jwt.RegisteredClaims{ExpiresAt: &jwt.NumericDate{Time: time.Unix(100, 1).UTC()}}
	if value, err := claimsHash(fractional); err == nil || value != "" {
		t.Fatal("unsupported precision produced a usable claim identity")
	}
}

func TestExpiredDeadlineHaltsBeforeNativeWorkAndKeepsTotal(t *testing.T) {
	// Synthetic empty fixtures cannot be valid JWTs. If the deadline guard moves
	// after parsing, zero native/callback/copy counters expose the regression.
	in := Input{Fixtures: make([]Fixture, 13)}
	past := time.Now().Add(-time.Second)
	for _, cached := range []bool{false, true} {
		// Assert the guard's reason directly: a recovered nil-key/cache panic
		// must not be mistaken for an expired deadline with no work attempted.
		probe := newStats(cached)
		result := trial(Fixture{}, 0, cached, nil, nil, "synthetic-der", &probe, past)
		if result.Known || result.Code != "unknown_deadline" || probe.Counts != (Counts{}) {
			t.Fatal("expired trial did not return its exact pre-work deadline state")
		}
		for _, manual := range []bool{false, true} {
			run := operation(in, nil, "synthetic-der", 0, 0, manual, cached, past)
			if run.Complete || run.Found != -1 {
				t.Fatal("unknown deadline became a completed choice")
			}
			if run.Stats.Timings.TotalNS == nil || *run.Stats.Timings.TotalNS < 0 {
				t.Fatal("deferred complete-parent elapsed record was lost")
			}
			if run.Stats.Counts != (Counts{CandidateChecks: 1}) {
				t.Fatal("deadline rejection performed native/setup/cache work")
			}
			if run.Stats.Timings.SetupNS != 0 || run.Stats.Timings.LookupNS != 0 || run.Stats.Timings.CopyNS != 0 || run.Stats.Timings.QualificationNS != 0 {
				t.Fatal("deadline rejection entered a measured native phase")
			}
			if run.Stats.Timings.AuthParseNS != nil && *run.Stats.Timings.AuthParseNS != 0 || run.Stats.Timings.PolicyNS != nil && *run.Stats.Timings.PolicyNS != 0 || run.Stats.Timings.CombinedNS != nil && *run.Stats.Timings.CombinedNS != 0 {
				t.Fatal("deadline rejection retained native phase time")
			}
			first := run.Order[0]
			for candidate := 0; candidate < 8; candidate++ {
				if run.Checked[candidate] != (candidate == first) || run.CandidateMatches[candidate] {
					t.Fatal("unknown deadline advanced candidate search")
				}
				checks := 0
				if candidate == first {
					checks = 1
				}
				if run.FixtureChecks[candidate] != checks {
					t.Fatal("attempted fixture accounting differs")
				}
			}
		}
	}
}
