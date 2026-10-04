// Prospective observer only. Authoring this file is not execution authorization.
// Root must freeze the complete source, inputs, compiler closure and budgets first.
package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	sourceRevision  = "73c870b18e68b6e654b2b03f485aa3c9fab32cea"
	inputSchema     = "riido-jwt-public-observer-input-v1"
	outputSchema    = "riido-jwt-direct-observer-output-v1"
	fixedUnix       = int64(1700000000)
	fixedHeader     = "{\"alg\":\"RS256\",\"typ\":\"JWT\"}"
	maxInputBytes   = 64 << 10
	maxOutputBytes  = 192 << 10
	maxJSONDepth    = 8
	maxJSONNodes    = 2048
	maxJSONString   = 2048
	maxKeyPEMBytes  = 2048
	maxKeyDERBytes  = 512
	maxTokenBytes   = 1024
	maxPayloadBytes = 256
	maxRunWall      = 5 * time.Second
	maxTrialWall    = time.Second
)

// This only counts this observer's own init. Imported-package startup happens
// before main and is explicitly outside all API-call and timing counters below.
var observerInitCalls int

func init() { observerInitCalls++ }

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
	Schema           string    `json:"schema"`
	SourceRevision   string    `json:"source_revision"`
	ClockUnixSeconds int64     `json:"clock_unix_seconds"`
	PublicKeyPEM     string    `json:"public_key_pem"`
	Fixtures         []Fixture `json:"fixtures"`
	Requests         []Request `json:"requests"`
}

// These are frozen literal input identities, not candidate truth, roles or masks.
var fixtureHashes = [13]string{
	"16f8a77e96006cac58231d3509b5651dcd9995c275631c0f800f6924255e254b",
	"2e0a8b9b36b129a4ffcb85323c40dc57fb1463527176fff2a68fb3177731331c",
	"8aa0f8ac1a2bb4c306aed7164064a9714d9361e4a2074b532d661fca8f4687fa",
	"665fe80b46bfc7a9dbff64dc8839001557695978f7035a76420ceeee1c039e2c",
	"e0e2e07e1bc15cb60ee280933e8eca6400a826ab2e1a1a6ea4015a7e30ea612f",
	"47a77101b36b0115e0a87b1fa7f75b99d21182ce813cb6e54fefe24b1a3d85dc",
	"dfd6c6ab8b5f2939853073db4b986776b415b7f61006a393b2154d9ed2be22d3",
	"ab0b3b2ffe886451c20f7daf10f1700cc4926623b0888d89f825a5e169698acb",
	"b6ea1118e326eb2d7ea8de3eae3b319bd070301a832f78f654e2e997b62e1ecb",
	"68bef3dac81307b47a4ff5a104a631b394d66b16174e0948409a6c76c966e54f",
	"5dfb4f5552d2ec1f05b27afc095d19529fcd53bf5f7208a66c44db151889109b",
	"28dcf7d130ad96d187a44ef048a4daac74c8d0927d20df6419fb990f0a86b96d",
	"c8711414a9bf6f80ab2b0c6fdc9de9b76bf5c08232f499dc127ebe0c17070c69",
}

var requestIDs = [4]string{"fiction-lantern", "fiction-harbor", "fiction-moss", "fiction-cove"}
var requestFixtureIndices = [4][5]int{
	{0, 1, 2, 3, 4},
	{0, 1, 2, 3, 4},
	{5, 2, 6, 7, 8},
	{9, 10, 11, 12, 8},
}

// Declaration order in the exact reviewed parser_option.go, not outcome order.
var candidateIDs = [8]string{
	"option-leeway", "option-issued-at", "option-exp-required", "option-nbf-required",
	"option-audience-any", "option-audience-all", "option-issuer", "option-subject",
}

type Counts struct {
	OptionConstructors int `json:"option_constructors"`
	OptionApplications int `json:"option_applications"`
	NewParser          int `json:"new_parser"`
	ParseWithClaims    int `json:"parse_with_claims"`
	KeyCallbacks       int `json:"key_callbacks"`
	TimeCallbacks      int `json:"time_callbacks"`
}

type Record struct {
	RequestID               string          `json:"request_id"`
	FixturePosition         int             `json:"fixture_position"`
	CandidatePosition       int             `json:"candidate_position"`
	FixtureID               string          `json:"fixture_id"`
	CandidateID             string          `json:"candidate_id"`
	Status                  string          `json:"status"`
	TokenPresent            bool            `json:"token_present"`
	Valid                   bool            `json:"valid"`
	ErrNil                  *bool           `json:"err_nil"`
	ParseReturned           bool            `json:"parse_returned"`
	ErrorClasses            []string        `json:"error_classes"`
	DecodedCanonicalPayload json.RawMessage `json:"decoded_canonical_payload"`
	PayloadMatch            bool            `json:"payload_match"`
	ClaimsZeroBefore        bool            `json:"claims_zero_before"`
	ClaimsInputUnchanged    bool            `json:"claims_input_unchanged"`
	KeyInputUnchanged       bool            `json:"key_input_unchanged"`
	Counts                  Counts          `json:"calls"`
	ElapsedNS               int64           `json:"elapsed_ns"`
	ParseElapsedNS          int64           `json:"parse_elapsed_ns"`
}

type StartupScope struct {
	ObserverInitCalls                int  `json:"observer_init_calls"`
	ImportedPackageInitCountObserved bool `json:"imported_package_init_count_observed"`
	APIIncludesImportedInit          bool `json:"api_counts_include_imported_init"`
	TimingStartsAtMain               bool `json:"timing_starts_at_main"`
}

type Output struct {
	Schema                string       `json:"schema"`
	SourceRevision        string       `json:"source_revision"`
	InputSHA256           string       `json:"input_sha256"`
	PublicKeyDERSHA256    string       `json:"public_key_der_sha256"`
	FixedClock            int64        `json:"fixed_clock_unix_seconds"`
	TimePrecisionNS       int64        `json:"time_precision_ns"`
	Startup               StartupScope `json:"startup_scope"`
	FixtureInputs         int          `json:"distinct_fixture_inputs"`
	FixturePositions      int          `json:"fixture_positions"`
	PlannedTrials         int          `json:"planned_trials"`
	Records               []Record     `json:"records"`
	Counts                Counts       `json:"calls"`
	CommonConstructors    [3]int       `json:"common_option_constructor_calls"`
	CandidateConstructors [8]int       `json:"candidate_option_constructor_calls"`
	Complete              bool         `json:"complete"`
	InputUnchanged        bool         `json:"input_unchanged"`
	MainElapsedNS         int64        `json:"main_elapsed_ns"`
}

func hash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Reject duplicate object keys before decoding structs; ordinary Unmarshal
// alone would silently accept a duplicate key. All work is bounded first.
func scanValue(d *json.Decoder, depth int, nodes *int) error {
	if depth > maxJSONDepth {
		return errors.New("budget")
	}
	*nodes = *nodes + 1
	if *nodes > maxJSONNodes {
		return errors.New("budget")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	if s, ok := t.(string); ok && len(s) > maxJSONString {
		return errors.New("budget")
	}
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			var seen [32]string
			seenCount := 0
			for d.More() {
				k, err := d.Token()
				if err != nil {
					return err
				}
				s, ok := k.(string)
				if !ok || len(s) > 64 || seenCount == len(seen) {
					return errors.New("shape")
				}
				for i := 0; i < seenCount; i++ {
					if seen[i] == s {
						return errors.New("duplicate")
					}
				}
				seen[seenCount] = s
				seenCount++
				if err := scanValue(d, depth+1, nodes); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim('}') {
				return errors.New("shape")
			}
		case '[':
			for d.More() {
				if err := scanValue(d, depth+1, nodes); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return errors.New("shape")
			}
		default:
			return errors.New("shape")
		}
	}
	return nil
}

func checkJSON(b []byte) error {
	if !validScalars(b) {
		return errors.New("scalar")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	nodes := 0
	if err := scanValue(d, 0, &nodes); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing")
	}
	return nil
}

// Exact reviewed scalar routine from cmd/riido-hintpreview: encoding/json
// otherwise replaces invalid UTF-8 and unpaired UTF-16 escapes with U+FFFD.
func validScalars(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	hex4 := func(start int) (uint16, bool) {
		if start+4 > len(b) {
			return 0, false
		}
		var n uint16
		for _, c := range b[start : start+4] {
			n <<= 4
			switch {
			case c >= '0' && c <= '9':
				n += uint16(c - '0')
			case c >= 'a' && c <= 'f':
				n += uint16(c-'a') + 10
			case c >= 'A' && c <= 'F':
				n += uint16(c-'A') + 10
			default:
				return 0, false
			}
		}
		return n, true
	}
	inside := false
	for i := 0; i < len(b); i++ {
		if b[i] == '"' {
			inside = !inside
			continue
		}
		if !inside || b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue // The JSON decoder validates all other escapes and syntax.
		}
		n, ok := hex4(i + 1)
		if !ok {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n >= 0xd800 && n <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return false
			}
			low, ok := hex4(i + 3)
			if !ok || low < 0xdc00 || low > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return !inside
}

// Canonicalize a restricted registered-claim map with the standard JSON
// encoder. This never invokes JWT's ClaimStrings/NumericDate marshal methods.
func canonicalLiteral(b []byte) ([]byte, error) {
	if len(b) > maxPayloadBytes || checkJSON(b) != nil {
		return nil, errors.New("claims")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var m map[string]any
	if err := d.Decode(&m); err != nil || m == nil || len(m) > 6 {
		return nil, errors.New("claims")
	}
	for k, v := range m {
		switch k {
		case "iss", "sub":
			s, ok := v.(string)
			if !ok || len(s) > 32 {
				return nil, errors.New("claims")
			}
		case "aud":
			a, ok := v.([]any)
			if !ok || len(a) < 1 || len(a) > 2 {
				return nil, errors.New("claims")
			}
			for _, x := range a {
				s, ok := x.(string)
				if !ok || len(s) > 16 {
					return nil, errors.New("claims")
				}
			}
		case "exp", "nbf", "iat":
			n, ok := v.(json.Number)
			if !ok {
				return nil, errors.New("claims")
			}
			x, err := n.Int64()
			if err != nil || x < fixedUnix-100 || x > fixedUnix+100 {
				return nil, errors.New("claims")
			}
			m[k] = x
		default:
			return nil, errors.New("claims")
		}
	}
	return json.Marshal(m)
}

func validateInput(raw []byte) (Input, *rsa.PublicKey, []byte, error) {
	var in Input
	if len(raw) == 0 || len(raw) > maxInputBytes || checkJSON(raw) != nil {
		return in, nil, nil, errors.New("input")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return in, nil, nil, err
	}
	if in.Schema != inputSchema || in.SourceRevision != sourceRevision || in.ClockUnixSeconds != fixedUnix || len(in.Fixtures) != 13 || len(in.Requests) != 4 {
		return in, nil, nil, errors.New("input")
	}
	if len(in.PublicKeyPEM) > maxKeyPEMBytes {
		return in, nil, nil, errors.New("key")
	}
	block, rest := pem.Decode([]byte(in.PublicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" || len(block.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 || len(block.Bytes) > maxKeyDERBytes {
		return in, nil, nil, errors.New("key")
	}
	keyAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return in, nil, nil, err
	}
	key, ok := keyAny.(*rsa.PublicKey)
	if !ok || key.N == nil || key.N.Sign() <= 0 || key.N.BitLen() != 2048 || key.E != 65537 {
		return in, nil, nil, errors.New("key")
	}
	for i, fixture := range in.Fixtures {
		if fixture.ID != fixtureHashes[i] || fixture.ClaimsSHA256 != fixtureHashes[i] || fixture.HeaderJSON != fixedHeader || hash([]byte(fixture.ClaimsJSON)) != fixtureHashes[i] || len(fixture.Token) > maxTokenBytes {
			return in, nil, nil, errors.New("fixture")
		}
		canonical, err := canonicalLiteral([]byte(fixture.ClaimsJSON))
		if err != nil || !bytes.Equal(canonical, []byte(fixture.ClaimsJSON)) {
			return in, nil, nil, errors.New("fixture")
		}
		parts := strings.Split(fixture.Token, ".")
		if len(parts) != 3 {
			return in, nil, nil, errors.New("token")
		}
		header, e0 := base64.RawURLEncoding.Strict().DecodeString(parts[0])
		payload, e1 := base64.RawURLEncoding.Strict().DecodeString(parts[1])
		signature, e2 := base64.RawURLEncoding.Strict().DecodeString(parts[2])
		if e0 != nil || e1 != nil || e2 != nil || !bytes.Equal(header, []byte(fixedHeader)) || !bytes.Equal(payload, []byte(fixture.ClaimsJSON)) || len(signature) != 256 {
			return in, nil, nil, errors.New("token")
		}
	}
	for i, request := range in.Requests {
		if request.ID != requestIDs[i] || len(request.FixtureIDs) != 5 {
			return in, nil, nil, errors.New("request")
		}
		for j, id := range request.FixtureIDs {
			if id != fixtureHashes[requestFixtureIndices[i][j]] {
				return in, nil, nil, errors.New("request")
			}
		}
	}
	return in, key, append([]byte(nil), block.Bytes...), nil
}

func canonicalParsed(c jwt.RegisteredClaims) ([]byte, error) {
	m := make(map[string]any, 7)
	if c.Issuer != "" {
		m["iss"] = c.Issuer
	}
	if c.Subject != "" {
		m["sub"] = c.Subject
	}
	if c.ID != "" {
		m["jti"] = c.ID
	}
	if len(c.Audience) != 0 {
		m["aud"] = append([]string(nil), c.Audience...)
	}
	for _, field := range []struct {
		name  string
		value *jwt.NumericDate
	}{{"exp", c.ExpiresAt}, {"nbf", c.NotBefore}, {"iat", c.IssuedAt}} {
		if field.value == nil {
			continue
		}
		if field.value.Nanosecond() != 0 {
			return nil, errors.New("fractional")
		}
		m[field.name] = field.value.Unix()
	}
	return json.Marshal(m)
}

// Walk only bounded unwrap edges. Standard pinned-source wrapping gives the
// same sentinel facts as errors.Is, without delegating to unbounded/custom Is.
func normalizedErrors(err error) (classes []string, status string) {
	classes = make([]string, 0, 21)
	if err == nil {
		return classes, "observed"
	}
	known := [...]struct {
		err  error
		code string
	}{
		{jwt.ErrInvalidKey, "invalid_key"}, {jwt.ErrInvalidKeyType, "invalid_key_type"},
		{jwt.ErrHashUnavailable, "hash_unavailable"}, {jwt.ErrTokenMalformed, "token_malformed"},
		{jwt.ErrTokenUnverifiable, "token_unverifiable"}, {jwt.ErrTokenSignatureInvalid, "signature_invalid"},
		{jwt.ErrTokenRequiredClaimMissing, "required_claim_missing"}, {jwt.ErrTokenInvalidAudience, "invalid_audience"},
		{jwt.ErrTokenExpired, "expired"}, {jwt.ErrTokenUsedBeforeIssued, "used_before_issued"},
		{jwt.ErrTokenInvalidIssuer, "invalid_issuer"}, {jwt.ErrTokenInvalidSubject, "invalid_subject"},
		{jwt.ErrTokenNotValidYet, "not_valid_yet"}, {jwt.ErrTokenInvalidId, "invalid_id"},
		{jwt.ErrTokenInvalidClaims, "invalid_claims"}, {jwt.ErrInvalidType, "invalid_claim_type"},
	}
	var seen [16]bool
	unknownLeaf, budget, resource, unwrapPanic := false, false, false, false
	defer func() {
		if recover() != nil {
			unwrapPanic = true
		}
		classes = classes[:0]
		for i, entry := range known {
			if seen[i] {
				classes = append(classes, entry.code)
			}
		}
		policyCause := false
		for _, i := range [...]int{6, 7, 8, 9, 10, 11, 12} {
			if seen[i] {
				policyCause = true
			}
		}
		wrapperOnly := seen[14] && !policyCause
		if unknownLeaf {
			classes = append(classes, "unrecognized_leaf")
		}
		if budget {
			classes = append(classes, "unwrap_budget")
		}
		if wrapperOnly {
			classes = append(classes, "wrapper_only_invalid_claims")
		}
		if resource {
			classes = append(classes, "resource_context")
		}
		if unwrapPanic {
			classes = append(classes, "unwrap_panic")
		}
		switch {
		case unwrapPanic:
			status = "unknown_panic"
		case budget || resource:
			status = "unknown_resource"
		case unknownLeaf || wrapperOnly:
			status = "unknown_error"
		case seen[0] || seen[1] || seen[2] || seen[4] || seen[5]:
			status = "unknown_crypto"
		case seen[3] || seen[15]:
			status = "unknown_input"
		case policyCause:
			status = "observed"
		default:
			status = "unknown_error"
		}
	}()
	type frame struct {
		err   error
		depth int
	}
	var stack [32]frame
	stack[0] = frame{err, 0}
	remaining, visited := 1, 0
	for remaining != 0 {
		if visited == len(stack) {
			budget = true
			break
		}
		remaining--
		node := stack[remaining]
		visited++
		if node.depth > 8 {
			budget = true
			break
		}
		if node.err == nil {
			unknownLeaf = true
			continue
		}
		matched := false
		for i, entry := range known {
			if node.err == entry.err {
				seen[i] = true
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if node.err == context.Canceled || node.err == context.DeadlineExceeded {
			resource = true
			continue
		}
		switch wrapped := node.err.(type) {
		case interface{ Unwrap() []error }:
			children := wrapped.Unwrap()
			if len(children) == 0 {
				unknownLeaf = true
				continue
			}
			if len(children) > len(stack)-remaining {
				budget = true
				break
			}
			for i := len(children) - 1; i >= 0; i-- {
				stack[remaining] = frame{children[i], node.depth + 1}
				remaining++
			}
		case interface{ Unwrap() error }:
			child := wrapped.Unwrap()
			if child == nil {
				unknownLeaf = true
				continue
			}
			if remaining == len(stack) {
				budget = true
				break
			}
			stack[remaining] = frame{child, node.depth + 1}
			remaining++
		default:
			unknownLeaf = true
		}
		if budget {
			break
		}
	}
	return classes, status
}

func countedOption(opt jwt.ParserOption, r *Record) jwt.ParserOption {
	return func(p *jwt.Parser) { r.Counts.OptionApplications++; opt(p) }
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
		panic("candidate_position")
	}
}

func observe(requestID string, fixturePosition, candidatePosition int, fixture Fixture, publicKey *rsa.PublicKey, deadline time.Time, commonCounts *[3]int, candidateCounts *[8]int) (r Record) {
	r = Record{RequestID: requestID, FixturePosition: fixturePosition, CandidatePosition: candidatePosition, FixtureID: fixture.ID, CandidateID: candidateIDs[candidatePosition], Status: "unknown_observer", ErrorClasses: []string{}, DecodedCanonicalPayload: json.RawMessage("null")}
	started := time.Now()
	// ParseWithClaims must receive this fresh variable by pointer. Its mutation
	// is decoded output, not mutation of the immutable fixture JSON input.
	claims := jwt.RegisteredClaims{}
	r.ClaimsZeroBefore = claims.Issuer == "" && claims.Subject == "" && claims.ID == "" && claims.Audience == nil && claims.ExpiresAt == nil && claims.NotBefore == nil && claims.IssuedAt == nil
	key := rsa.PublicKey{N: new(big.Int).Set(publicKey.N), E: publicKey.E}
	keyBefore := append([]byte(nil), key.N.Bytes()...)
	tokenHash := hash([]byte(fixture.Token))
	var parseStart time.Time
	defer func() {
		if recover() != nil {
			r.Status = "unknown_panic"
			r.ErrorClasses = []string{"panic"}
		}
		// Also retain an unknown record if qualification itself panics. This
		// nested guard completes the named result and the remaining audit loop.
		defer func() {
			if recover() != nil {
				r.Status = "unknown_panic"
				r.ErrorClasses = []string{"panic"}
			}
			r.ElapsedNS = time.Since(started).Nanoseconds()
			if r.ElapsedNS > maxTrialWall.Nanoseconds() && r.Status == "observed" {
				r.Status = "unknown_resource"
			}
		}()
		if !r.ParseReturned && !parseStart.IsZero() {
			r.ParseElapsedNS = time.Since(parseStart).Nanoseconds()
		}
		r.ClaimsInputUnchanged = hash([]byte(fixture.ClaimsJSON)) == fixture.ClaimsSHA256 && hash([]byte(fixture.Token)) == tokenHash
		r.KeyInputUnchanged = key.N != nil && key.E == publicKey.E && bytes.Equal(key.N.Bytes(), keyBefore)
		if r.ParseReturned {
			decoded, err := canonicalParsed(claims)
			if err == nil {
				r.DecodedCanonicalPayload = decoded
				r.PayloadMatch = bytes.Equal(decoded, []byte(fixture.ClaimsJSON))
			}
			if (err != nil || !r.PayloadMatch) && r.Status == "observed" {
				r.Status = "unknown_payload"
			}
		}
		if (!r.ClaimsInputUnchanged || !r.KeyInputUnchanged) && r.Status == "observed" {
			r.Status = "unknown_input_mutation"
		}
	}()
	if started.After(deadline) {
		r.Status = "unknown_resource"
		r.ErrorClasses = []string{"worker_deadline"}
		return r
	}
	r.Counts.OptionConstructors++
	commonCounts[0]++
	method := countedOption(jwt.WithValidMethods([]string{"RS256"}), &r)
	r.Counts.OptionConstructors++
	commonCounts[1]++
	clock := countedOption(jwt.WithTimeFunc(func() time.Time { r.Counts.TimeCallbacks++; return time.Unix(fixedUnix, 0).UTC() }), &r)
	r.Counts.OptionConstructors++
	commonCounts[2]++
	strict := countedOption(jwt.WithStrictDecoding(), &r)
	r.Counts.OptionConstructors++
	candidateCounts[candidatePosition]++
	option := countedOption(candidateOption(candidatePosition), &r)
	r.Counts.NewParser++
	parser := jwt.NewParser(method, clock, strict, option)
	r.Counts.ParseWithClaims++
	parseStart = time.Now()
	token, err := parser.ParseWithClaims(fixture.Token, &claims, func(_ *jwt.Token) (any, error) { r.Counts.KeyCallbacks++; return &key, nil })
	r.ParseElapsedNS = time.Since(parseStart).Nanoseconds()
	r.ParseReturned = true
	r.TokenPresent = token != nil
	r.Valid = token != nil && token.Valid
	errNil := err == nil
	r.ErrNil = &errNil
	r.ErrorClasses, r.Status = normalizedErrors(err)
	if (errNil && (!r.TokenPresent || !r.Valid)) || (!errNil && r.Valid) {
		r.Status = "unknown_observer"
	}
	return r
}

func addCounts(dst *Counts, src Counts) {
	dst.OptionConstructors += src.OptionConstructors
	dst.OptionApplications += src.OptionApplications
	dst.NewParser += src.NewParser
	dst.ParseWithClaims += src.ParseWithClaims
	dst.KeyCallbacks += src.KeyCallbacks
	dst.TimeCallbacks += src.TimeCallbacks
}

func run() (exit int) {
	started := time.Now()
	defer func() {
		if recover() != nil {
			_, _ = io.WriteString(os.Stderr, "observer_failure\n")
			exit = 3
		}
	}()
	if len(os.Args) != 1 {
		_, _ = io.WriteString(os.Stderr, "observer_input_rejected\n")
		return 2
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, maxInputBytes+1))
	if err != nil {
		_, _ = io.WriteString(os.Stderr, "observer_input_rejected\n")
		return 2
	}
	in, key, der, err := validateInput(raw)
	if err != nil {
		_, _ = io.WriteString(os.Stderr, "observer_input_rejected\n")
		return 2
	}
	// Do not assign global TimePrecision: the exact source default is part of
	// qualification, and changing it would hide a startup/source discrepancy.
	if jwt.TimePrecision != time.Second {
		_, _ = io.WriteString(os.Stderr, "observer_source_state_rejected\n")
		return 3
	}
	before, err := json.Marshal(in)
	if err != nil {
		_, _ = io.WriteString(os.Stderr, "observer_failure\n")
		return 3
	}
	out := Output{Schema: outputSchema, SourceRevision: sourceRevision, InputSHA256: hash(raw), PublicKeyDERSHA256: hash(der), FixedClock: fixedUnix, TimePrecisionNS: int64(jwt.TimePrecision), Startup: StartupScope{ObserverInitCalls: observerInitCalls, ImportedPackageInitCountObserved: false, APIIncludesImportedInit: false, TimingStartsAtMain: true}, FixtureInputs: 13, FixturePositions: 20, PlannedTrials: 160, Records: make([]Record, 0, 160), Complete: true}
	deadline := started.Add(maxRunWall)
	for requestPosition := 0; requestPosition < 4; requestPosition++ {
		for fixturePosition := 0; fixturePosition < 5; fixturePosition++ {
			fixture := in.Fixtures[requestFixtureIndices[requestPosition][fixturePosition]]
			for candidatePosition := 0; candidatePosition < 8; candidatePosition++ {
				r := observe(in.Requests[requestPosition].ID, fixturePosition, candidatePosition, fixture, key, deadline, &out.CommonConstructors, &out.CandidateConstructors)
				if r.Status != "observed" {
					out.Complete = false
				}
				addCounts(&out.Counts, r.Counts)
				out.Records = append(out.Records, r)
			}
		}
	}
	after, err := json.Marshal(in)
	out.InputUnchanged = err == nil && bytes.Equal(before, after)
	if !out.InputUnchanged || jwt.TimePrecision != time.Second {
		out.Complete = false
	}
	out.MainElapsedNS = time.Since(started).Nanoseconds()
	encoded, err := json.Marshal(out)
	if err != nil || len(encoded)+1 > maxOutputBytes {
		_, _ = io.WriteString(os.Stderr, "observer_output_rejected\n")
		return 3
	}
	encoded = append(encoded, '\n')
	n, err := os.Stdout.Write(encoded)
	if err != nil || n != len(encoded) {
		_, _ = io.WriteString(os.Stderr, "observer_output_write_failed\n")
		return 3
	}
	if !out.Complete {
		_, _ = io.WriteString(os.Stderr, "observer_incomplete\n")
		return 3
	}
	return 0
}

func main() { os.Exit(run()) }

// No signing, key generation, source fetching, model, Rank, Fit, labels, masks,
// roles, candidate aggregation or request-Wanted data exist in this observer.
