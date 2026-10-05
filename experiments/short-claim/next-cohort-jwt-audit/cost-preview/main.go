// Author-only draft: not authorization to compile, import-init, or execute.
// Root freezes source, selected closure, input, binary and first-run budgets first.
package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const sourceRevision = "73c870b18e68b6e654b2b03f485aa3c9fab32cea"
const inputSHA = "47686828bbb0cd4d273db3b2087311357ea853b767f86fd787953ed76d452ed6"
const fixedUnix = int64(1700000000)
const maxInput = 64 << 10
const maxOutput = 192 << 10

var ownInit int

func init() { ownInit++ }

type Fixture struct {
	ID           string `json:"id"`
	Token        string `json:"token"`
	HeaderJSON   string `json:"header_json"`
	ClaimsJSON   string `json:"claims_json"`
	ClaimsSHA256 string `json:"claims_sha256"`
}
type Request struct {
	ID         string   `json:"id"`
	FixtureIDs []string `json:"fixture_ids"`
}
type Input struct {
	Schema         string    `json:"schema"`
	SourceRevision string    `json:"source_revision"`
	Clock          int64     `json:"clock_unix_seconds"`
	PublicKeyPEM   string    `json:"public_key_pem"`
	Fixtures       []Fixture `json:"fixtures"`
	Requests       []Request `json:"requests"`
}

// These fixed controls were declared before the first original API observations.
// They are exposed development semantics, not learned hints or independent labels.
var manualFirst = [4]int{4, 5, 2, 0}
var wanted = [4][5]bool{
	{true, true, true, false, false},
	{false, false, true, false, false},
	{false, true, false, false, true},
	{true, false, true, false, true},
}
var positions = [4][5]int{{0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}, {5, 2, 6, 7, 8}, {9, 10, 11, 12, 8}}
var requestNames = [4]string{"fiction-lantern", "fiction-harbor", "fiction-moss", "fiction-cove"}

type Counts struct {
	OptionConstructors           int `json:"option_constructors"`
	OptionApplications           int `json:"option_applications"`
	DirectNewParser              int `json:"direct_new_parser"`
	ParseWithClaims              int `json:"parse_with_claims"`
	NewValidator                 int `json:"new_validator"`
	ImplicitParserInNewValidator int `json:"implicit_parser_in_new_validator"`
	Validate                     int `json:"validate"`
	KeyCallbacks                 int `json:"key_callbacks"`
	TimeCallbacks                int `json:"time_callbacks"`
	CacheLookups                 int `json:"cache_lookups"`
	CacheComparisons             int `json:"cache_key_comparisons"`
	CacheHits                    int `json:"cache_hits"`
	CacheMisses                  int `json:"cache_misses"`
	ClaimsCopies                 int `json:"deep_claims_copies"`
	PublicKeyCopies              int `json:"public_key_copies"`
	CandidateChecks              int `json:"candidate_checks"`
	FixtureChecks                int `json:"fixture_checks"`
}
type Timings struct {
	SetupNS         int64  `json:"setup_ns"`
	LookupNS        int64  `json:"cache_lookup_ns"`
	CopyNS          int64  `json:"claims_and_key_copy_ns"`
	CacheStoreNS    int64  `json:"cache_claims_store_ns"`
	RankNS          int64  `json:"control_order_construction_ns"`
	AuthParseNS     *int64 `json:"authentication_and_parse_ns"`
	PolicyNS        *int64 `json:"fresh_policy_validate_ns"`
	CombinedNS      *int64 `json:"combined_parse_signature_policy_ns"`
	QualificationNS int64  `json:"outcome_qualification_ns"`
	TotalNS         *int64 `json:"complete_parent_work_ns"`
}
type Stats struct {
	Counts  Counts  `json:"counts"`
	Timings Timings `json:"timings"`
}

func newStats(cached bool) Stats {
	s := Stats{}
	if cached {
		a, p := int64(0), int64(0)
		s.Timings.AuthParseNS, s.Timings.PolicyNS = &a, &p
	} else {
		c := int64(0)
		s.Timings.CombinedNS = &c
	}
	return s
}

type Result struct {
	Known           bool   `json:"known"`
	Accept          bool   `json:"accept"`
	PolicyErrorMask uint16 `json:"policy_error_mask"`
	ClaimsSHA       string `json:"canonical_claims_sha256"`
	Code            string `json:"code"`
}
type Pair struct {
	Parent    int    `json:"parent_position"`
	Fixture   int    `json:"fixture_position"`
	Candidate int    `json:"candidate_position"`
	Reference Result `json:"uncached"`
	Cached    Result `json:"cached"`
	Equal     bool   `json:"equal"`
}
type Equivalence struct {
	PlannedPairs   int      `json:"planned_pairs"`
	AttemptedPairs int      `json:"attempted_pairs"`
	EqualPairs     int      `json:"equal_pairs"`
	AllEqual       bool     `json:"all_equal"`
	Records        []Pair   `json:"records"`
	Reference      [4]Stats `json:"uncached_parent_stats"`
	Cached         [4]Stats `json:"cached_parent_stats"`
}
type Run struct {
	Round            int     `json:"round"`
	Control          string  `json:"control"`
	Cached           bool    `json:"cached"`
	Parent           int     `json:"parent_position"`
	Order            [8]int  `json:"candidate_order"`
	Checked          [8]bool `json:"candidate_attempted"`
	FixtureChecks    [8]int  `json:"fixture_checks_by_candidate"`
	CandidateMatches [8]bool `json:"candidate_matches_exact_wanted"`
	Found            int     `json:"first_matching_candidate"`
	Complete         bool    `json:"complete"`
	Stats            Stats   `json:"stats"`
}
type Output struct {
	Schema                 string      `json:"schema"`
	SourceRevision         string      `json:"source_revision"`
	InputSHA               string      `json:"input_sha256"`
	PublicKeyDERSHA        string      `json:"public_key_der_sha256"`
	Family                 string      `json:"source_family"`
	Role                   string      `json:"role"`
	Scope                  string      `json:"scope"`
	OwnInit                int         `json:"own_init_calls"`
	ImportedInitObserved   bool        `json:"imported_package_init_count_observed"`
	MainStartsAfterImports bool        `json:"main_timing_excludes_import_startup"`
	PolicyMaskClasses      [8]string   `json:"normalized_error_classes_bit_order"`
	Equivalence            Equivalence `json:"exhaustive_equivalence"`
	Runs                   []Run       `json:"operational_runs"`
	Complete               bool        `json:"complete"`
	Code                   string      `json:"code"`
	MainNS                 int64       `json:"main_elapsed_ns"`
	InputSetupNS           int64       `json:"input_and_public_key_setup_ns"`
	SignaturesGenerated    int         `json:"signatures_generated"`
	KeyGenerations         int         `json:"key_generations"`
	ModelCalls             int         `json:"model_calls"`
	FitCalls               int         `json:"fit_calls"`
	RSAVerifyCallsObserved bool        `json:"rsa_verify_call_count_observed"`
	InputUnchanged         bool        `json:"input_unmodified"`
}

// Full DER identity, not just its hash, is part of every exact cache lookup.
// The bound is five fixture positions in one parent. No cross-parent/control state.
type CacheKey struct {
	Token   string
	DER     string
	Method  string
	Decoder string
	Source  string
}
type Entry struct {
	Key    CacheKey
	Claims jwt.RegisteredClaims
}
type Cache struct {
	Entries [5]Entry
	Used    int
}

func hash(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func cloneClaims(c jwt.RegisteredClaims) jwt.RegisteredClaims {
	if c.Audience != nil {
		c.Audience = append(jwt.ClaimStrings{}, c.Audience...)
	}
	if c.ExpiresAt != nil {
		n := *c.ExpiresAt
		c.ExpiresAt = &n
	}
	if c.NotBefore != nil {
		n := *c.NotBefore
		c.NotBefore = &n
	}
	if c.IssuedAt != nil {
		n := *c.IssuedAt
		c.IssuedAt = &n
	}
	return c
}
func dateValue(n *jwt.NumericDate) (*int64, error) {
	if n == nil {
		return nil, nil
	}
	if n.Nanosecond() != 0 {
		return nil, errors.New("fractional_date")
	}
	value := n.Unix()
	return &value, nil
}
func claimsHash(c jwt.RegisteredClaims) (string, error) {
	exp, e := dateValue(c.ExpiresAt)
	if e != nil {
		return "", e
	}
	nbf, e := dateValue(c.NotBefore)
	if e != nil {
		return "", e
	}
	iat, e := dateValue(c.IssuedAt)
	if e != nil {
		return "", e
	}
	canonical := struct {
		Audience []string `json:"aud,omitempty"`
		Exp      *int64   `json:"exp,omitempty"`
		Iat      *int64   `json:"iat,omitempty"`
		Iss      string   `json:"iss,omitempty"`
		ID       string   `json:"jti,omitempty"`
		Nbf      *int64   `json:"nbf,omitempty"`
		Sub      string   `json:"sub,omitempty"`
	}{append([]string(nil), c.Audience...), exp, iat, c.Issuer, c.ID, nbf, c.Subject}
	raw, e := json.Marshal(canonical)
	if e != nil {
		return "", e
	}
	return hash(raw), nil
}

// Only recognized policy leaves are Boolean false. Generic parser wrapper is
// retained as bit7 in the mask; alone it remains U. Cached policy errors receive
// the same sentinel wrapper before classification, without comparing messages.
// Unknown children, crypto/input errors, unwrap limits or panic remain unknown.
func policyStatus(err error) (known, accept bool, mask uint16) {
	if err == nil {
		return true, true, 0
	}
	leaves := [7]error{jwt.ErrTokenRequiredClaimMissing, jwt.ErrTokenInvalidAudience,
		jwt.ErrTokenExpired, jwt.ErrTokenUsedBeforeIssued, jwt.ErrTokenInvalidIssuer,
		jwt.ErrTokenInvalidSubject, jwt.ErrTokenNotValidYet}
	defer func() {
		if recover() != nil {
			known, accept, mask = false, false, 0
		}
	}()
	type frame struct {
		err   error
		depth int
	}
	var stack [32]frame
	stack[0] = frame{err, 0}
	remaining, visited := 1, 0
	for remaining > 0 {
		if visited == 32 {
			return false, false, mask
		}
		remaining--
		f := stack[remaining]
		visited++
		if f.depth > 8 || f.err == nil {
			return false, false, mask
		}
		matched := false
		for i, e := range leaves {
			if f.err == e {
				mask |= 1 << i
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if f.err == jwt.ErrTokenInvalidClaims {
			mask |= 1 << 7
			continue
		}
		if f.err == context.Canceled || f.err == context.DeadlineExceeded {
			return false, false, mask
		}
		switch w := f.err.(type) {
		case interface{ Unwrap() []error }:
			children := w.Unwrap()
			if len(children) == 0 || len(children) > 32-remaining {
				return false, false, mask
			}
			for i := len(children) - 1; i >= 0; i-- {
				stack[remaining] = frame{children[i], f.depth + 1}
				remaining++
			}
		case interface{ Unwrap() error }:
			child := w.Unwrap()
			if child == nil || remaining == 32 {
				return false, false, mask
			}
			stack[remaining] = frame{child, f.depth + 1}
			remaining++
		default:
			return false, false, mask
		}
	}
	return mask&127 != 0, false, mask
}

func candidateOption(pos int) jwt.ParserOption {
	switch pos {
	case 0:
		return jwt.WithLeeway(2 * time.Second)
	case 1:
		return jwt.WithIssuedAt()
	case 2:
		return jwt.WithExpirationRequired()
	case 3:
		return jwt.WithNotBeforeRequired()
	case 4:
		return jwt.WithAudience("alpha", "beta")
	case 5:
		return jwt.WithAllAudiences("alpha", "beta")
	case 6:
		return jwt.WithIssuer("issuer-A")
	case 7:
		return jwt.WithSubject("user-A")
	default:
		panic("candidate_bound")
	}
}
func counted(opt jwt.ParserOption, s *Stats) jwt.ParserOption {
	s.Counts.OptionConstructors++
	return func(p *jwt.Parser) { s.Counts.OptionApplications++; opt(p) }
}
func options(s *Stats, pos int, without bool) []jwt.ParserOption {
	opts := []jwt.ParserOption{
		counted(jwt.WithValidMethods([]string{"RS256"}), s),
		counted(jwt.WithTimeFunc(func() time.Time { s.Counts.TimeCallbacks++; return time.Unix(fixedUnix, 0).UTC() }), s),
		counted(jwt.WithStrictDecoding(), s),
	}
	if without {
		opts = append(opts, counted(jwt.WithoutClaimsValidation(), s))
	}
	if pos >= 0 {
		opts = append(opts, counted(candidateOption(pos), s))
	}
	return opts
}
func keyCopy(key *rsa.PublicKey, s *Stats) *rsa.PublicKey {
	started := time.Now()
	copy := &rsa.PublicKey{N: new(big.Int).Set(key.N), E: key.E}
	s.Counts.PublicKeyCopies++
	s.Timings.CopyNS += time.Since(started).Nanoseconds()
	return copy
}
func qualifyResult(f Fixture, c jwt.RegisteredClaims, err error, tokenOK bool, s *Stats) Result {
	started := time.Now()
	defer func() { s.Timings.QualificationNS += time.Since(started).Nanoseconds() }()
	known, accept, mask := policyStatus(err)
	claimSHA, e := claimsHash(c)
	result := Result{Known: known, Accept: accept, PolicyErrorMask: mask, ClaimsSHA: claimSHA, Code: "observed"}
	if e != nil || claimSHA != f.ClaimsSHA256 {
		result.Known = false
		result.Code = "unknown_payload"
	}
	if !tokenOK {
		result.Known = false
		result.Code = "unknown_return"
	}
	if !known {
		result.Code = "unknown_error"
	}
	return result
}

func trial(f Fixture, pos int, cached bool, cache *Cache, key *rsa.PublicKey, der string, s *Stats, deadline time.Time) (result Result) {
	result = Result{Code: "unknown_observer"}
	defer func() {
		if recover() != nil {
			result = Result{Code: "unknown_panic"}
		}
	}()
	if time.Now().After(deadline) {
		return Result{Code: "unknown_deadline"}
	}
	s.Counts.FixtureChecks++
	if !cached {
		claims := jwt.RegisteredClaims{}
		public := keyCopy(key, s)
		setup := time.Now()
		opts := options(s, pos, false)
		s.Counts.DirectNewParser++
		parser := jwt.NewParser(opts...)
		s.Timings.SetupNS += time.Since(setup).Nanoseconds()
		started := time.Now()
		s.Counts.ParseWithClaims++
		token, err := parser.ParseWithClaims(f.Token, &claims, func(_ *jwt.Token) (any, error) { s.Counts.KeyCallbacks++; return public, nil })
		*s.Timings.CombinedNS += time.Since(started).Nanoseconds()
		if public.N == nil || public.E != key.E || public.N.Cmp(key.N) != 0 {
			return Result{Code: "unknown_key_mutation"}
		}
		consistent := token != nil && ((err == nil && token.Valid) || (err != nil && !token.Valid))
		return qualifyResult(f, claims, err, consistent, s)
	}
	lookup := time.Now()
	wantedKey := CacheKey{f.Token, der, "RS256", "raw-url-base64-strict;RegisteredClaims;defaultJSON", "73c870b18e68b6e654b2b03f485aa3c9fab32cea"}
	s.Counts.CacheLookups++
	entry := -1
	for i := 0; i < cache.Used; i++ {
		s.Counts.CacheComparisons++
		if cache.Entries[i].Key == wantedKey {
			entry = i
			break
		}
	}
	s.Timings.LookupNS += time.Since(lookup).Nanoseconds()
	if entry < 0 {
		s.Counts.CacheMisses++
		if cache.Used == len(cache.Entries) {
			return Result{Code: "unknown_cache_capacity"}
		}
		claims := jwt.RegisteredClaims{}
		public := keyCopy(key, s)
		setup := time.Now()
		opts := options(s, -1, true)
		s.Counts.DirectNewParser++
		parser := jwt.NewParser(opts...)
		s.Timings.SetupNS += time.Since(setup).Nanoseconds()
		auth := time.Now()
		s.Counts.ParseWithClaims++
		token, err := parser.ParseWithClaims(f.Token, &claims, func(_ *jwt.Token) (any, error) { s.Counts.KeyCallbacks++; return public, nil })
		*s.Timings.AuthParseNS += time.Since(auth).Nanoseconds()
		if public.N == nil || public.E != key.E || public.N.Cmp(key.N) != 0 {
			return Result{Code: "unknown_key_mutation"}
		}
		// Authentication success is scoped to frozen token/key/decoder/method/source.
		// Token.Valid is deliberately not cached or used as candidate policy result.
		if err != nil || token == nil {
			return Result{Code: "unknown_authentication"}
		}
		store := time.Now()
		claimSHA, e := claimsHash(claims)
		if e != nil || claimSHA != f.ClaimsSHA256 {
			return Result{Code: "unknown_auth_payload"}
		}
		entry = cache.Used
		cache.Entries[entry] = Entry{Key: wantedKey, Claims: cloneClaims(claims)}
		cache.Used++
		s.Counts.ClaimsCopies++
		s.Timings.CacheStoreNS += time.Since(store).Nanoseconds()
	} else {
		s.Counts.CacheHits++
	}
	copying := time.Now()
	claims := cloneClaims(cache.Entries[entry].Claims)
	s.Counts.ClaimsCopies++
	s.Timings.CopyNS += time.Since(copying).Nanoseconds()
	setup := time.Now()
	opts := options(s, pos, false)
	s.Counts.NewValidator++
	s.Counts.ImplicitParserInNewValidator++
	validator := jwt.NewValidator(opts...)
	s.Timings.SetupNS += time.Since(setup).Nanoseconds()
	policy := time.Now()
	s.Counts.Validate++
	err := validator.Validate(&claims)
	*s.Timings.PolicyNS += time.Since(policy).Nanoseconds()
	qualification := time.Now()
	qualificationBeforeNS := s.Timings.QualificationNS
	if err != nil {
		err = errors.Join(jwt.ErrTokenInvalidClaims, err)
	}
	result = qualifyResult(f, claims, err, true, s)
	storedSHA, e := claimsHash(cache.Entries[entry].Claims)
	if e != nil || storedSHA != f.ClaimsSHA256 {
		result.Known = false
		result.Code = "unknown_cache_mutation"
	}
	// qualifyResult measured its own inner work; add only remaining checks.
	qualificationInnerNS := s.Timings.QualificationNS - qualificationBeforeNS
	s.Timings.QualificationNS += time.Since(qualification).Nanoseconds() - qualificationInnerNS
	return result
}

func equivalence(in Input, key *rsa.PublicKey, der string, deadline time.Time) Equivalence {
	out := Equivalence{PlannedPairs: 160, AllEqual: true, Records: make([]Pair, 0, 160)}
	for parent := 0; parent < 4; parent++ {
		a, b := newStats(false), newStats(true)
		var cache Cache
		for fixture := 0; fixture < 5; fixture++ {
			f := in.Fixtures[positions[parent][fixture]]
			for candidate := 0; candidate < 8; candidate++ {
				x := trial(f, candidate, false, nil, key, der, &a, deadline)
				y := trial(f, candidate, true, &cache, key, der, &b, deadline)
				equal := x.Known && y.Known && x == y
				out.Records = append(out.Records, Pair{parent, fixture, candidate, x, y, equal})
				out.AttemptedPairs++
				if equal {
					out.EqualPairs++
				} else {
					out.AllEqual = false
				}
			}
		}
		// Interleaved parent total cannot be attributed to either path.
		// Per-phase additive timers and counts are kept; parent TotalNS is null.
		out.Reference[parent], out.Cached[parent] = a, b
	}
	return out
}
func orderFor(parent int, manual bool) [8]int {
	var order [8]int
	if !manual {
		for i := 0; i < 8; i++ {
			order[i] = i
		}
		return order
	}
	order[0] = manualFirst[parent]
	n := 1
	for i := 0; i < 8; i++ {
		if i != order[0] {
			order[n] = i
			n++
		}
	}
	return order
}
func operation(in Input, key *rsa.PublicKey, der string, round, parent int, manual, cached bool, deadline time.Time) (out Run) {
	started := time.Now()
	out = Run{Round: round, Parent: parent, Cached: cached, Found: -1, Complete: true, Stats: newStats(cached)}
	if manual {
		out.Control = "manual-first"
	} else {
		out.Control = "fixed-source-order"
	}
	rank := time.Now()
	out.Order = orderFor(parent, manual)
	out.Stats.Timings.RankNS = time.Since(rank).Nanoseconds()
	var cache Cache
	defer func() { value := time.Since(started).Nanoseconds(); out.Stats.Timings.TotalNS = &value }()
	for _, candidate := range out.Order {
		out.Checked[candidate] = true
		out.Stats.Counts.CandidateChecks++
		matches := true
		for fixture := 0; fixture < 5; fixture++ {
			f := in.Fixtures[positions[parent][fixture]]
			result := trial(f, candidate, cached, &cache, key, der, &out.Stats, deadline)
			out.FixtureChecks[candidate]++
			if !result.Known {
				out.Complete = false
				return out
			}
			if result.Accept != wanted[parent][fixture] {
				matches = false
				break
			}
		}
		out.CandidateMatches[candidate] = matches
		if matches {
			out.Found = candidate
			return out
		}
	}
	return out
}

func parseInput(raw []byte) (Input, *rsa.PublicKey, []byte, error) {
	if len(raw) != 13474 || hash(raw) != inputSHA {
		return Input{}, nil, nil, errors.New("input_pin")
	}
	var in Input
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&in); e != nil {
		return Input{}, nil, nil, e
	}
	if in.Schema != "riido-jwt-public-observer-input-v1" || in.SourceRevision != sourceRevision || in.Clock != fixedUnix || len(in.Fixtures) != 13 || len(in.Requests) != 4 {
		return Input{}, nil, nil, errors.New("input_shape")
	}
	for parent, r := range in.Requests {
		if r.ID != requestNames[parent] || len(r.FixtureIDs) != 5 {
			return Input{}, nil, nil, errors.New("request_shape")
		}
		for i, id := range r.FixtureIDs {
			if id != in.Fixtures[positions[parent][i]].ID {
				return Input{}, nil, nil, errors.New("fixture_order")
			}
		}
	}
	for _, f := range in.Fixtures {
		if f.HeaderJSON != "{\"alg\":\"RS256\",\"typ\":\"JWT\"}" || f.ID != f.ClaimsSHA256 || hash([]byte(f.ClaimsJSON)) != f.ID {
			return Input{}, nil, nil, errors.New("fixture_pin")
		}
	}
	block, rest := pem.Decode([]byte(in.PublicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" || len(bytes.TrimSpace(rest)) != 0 {
		return Input{}, nil, nil, errors.New("key_pem")
	}
	public, e := x509.ParsePKIXPublicKey(block.Bytes)
	if e != nil {
		return Input{}, nil, nil, e
	}
	key, ok := public.(*rsa.PublicKey)
	if !ok || key.N == nil || key.N.BitLen() != 2048 || key.E != 65537 {
		return Input{}, nil, nil, errors.New("key_shape")
	}
	return in, key, block.Bytes, nil
}
func emit(out Output) {
	raw, e := json.Marshal(out)
	if e != nil || len(raw)+1 > maxOutput {
		_, _ = os.Stdout.Write([]byte("{\"schema\":\"riido-jwt-operational-cost-preview-v1\",\"complete\":false,\"code\":\"output_bound\"}\n"))
		return
	}
	_, _ = os.Stdout.Write(append(raw, '\n'))
}
func main() {
	started := time.Now()
	out := Output{Schema: "riido-jwt-operational-cost-preview-v1", SourceRevision: sourceRevision,
		InputSHA: inputSHA, Family: "jwt-go-lineage-whole-family-v1", Role: "development_validation",
		Scope:   "Exposed finite four-parent development diagnostic only. Cache auth/decoded claims is per parent/control and exact identity; candidate policy always fresh. Three fixed rounds follow exhaustive paired equivalence, so operational runs are process-warm. Fixed schedule is not randomized or independent replication. Direct internal uncached auth/policy times are unavailable, not estimated. Imported init precedes main. Soft memory target only; outer process wall/RSS captured separately. No generalization, learned model, Fit or savings claim.",
		OwnInit: ownInit, MainStartsAfterImports: true, Code: "author_scoped_initial",
		PolicyMaskClasses: [8]string{"required_claim_missing", "invalid_audience", "expired", "used_before_issued", "invalid_issuer", "invalid_subject", "not_valid_yet", "invalid_claims_wrapper"},
		Runs:              make([]Run, 0, 48)}
	defer func() {
		if recover() != nil {
			out.Complete = false
			out.Code = "unknown_panic"
		}
		out.MainNS = time.Since(started).Nanoseconds()
		emit(out)
	}()
	if len(os.Args) != 1 || runtime.Version() != "go1.27.1" || runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		out.Code = "runtime_or_arguments"
		return
	}
	runtime.GOMAXPROCS(1)
	debug.SetMemoryLimit(64 << 20)
	if int64(jwt.TimePrecision) != int64(time.Second) {
		out.Code = "unsupported_time_precision"
		return
	}
	deadline := started.Add(5 * time.Second)
	setup := time.Now()
	raw, e := io.ReadAll(io.LimitReader(os.Stdin, maxInput+1))
	if e != nil || len(raw) > maxInput {
		out.Code = "input_bound"
		return
	}
	in, key, der, e := parseInput(raw)
	if e != nil {
		out.Code = "input_qualification"
		return
	}
	out.PublicKeyDERSHA = hash(der)
	derIdentity := string(der)
	beforeN := append([]byte(nil), key.N.Bytes()...)
	out.InputSetupNS = time.Since(setup).Nanoseconds()
	out.Equivalence = equivalence(in, key, derIdentity, deadline)
	if !out.Equivalence.AllEqual || out.Equivalence.AttemptedPairs != 160 {
		out.Code = "equivalence_failed"
		return
	}
	for round := 0; round < 3; round++ {
		for ci := 0; ci < 2; ci++ {
			manual := ci == 1
			if round == 1 {
				manual = !manual
			}
			for mi := 0; mi < 2; mi++ {
				cached := mi == 1
				if round == 1 {
					cached = !cached
				}
				for parent := 0; parent < 4; parent++ {
					run := operation(in, key, derIdentity, round, parent, manual, cached, deadline)
					out.Runs = append(out.Runs, run)
					if !run.Complete || run.Found < 0 {
						out.Code = "operational_unknown_or_no_match"
						return
					}
				}
			}
		}
	}
	out.InputUnchanged = hash(raw) == inputSHA && bytes.Equal(key.N.Bytes(), beforeN) && key.E == 65537
	out.Complete = out.InputUnchanged && len(out.Runs) == 48
	out.Code = "completed_finite_development_preview"
}
