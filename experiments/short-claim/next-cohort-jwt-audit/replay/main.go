// SPDX-License-Identifier: Apache-2.0
// Saved observation audit only: no JWT dependency, parser, signing, model or Fit.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	base           = "experiments/short-claim/next-cohort-jwt-audit/"
	revision       = "73c870b18e68b6e654b2b03f485aa3c9fab32cea"
	preparationSHA = "9f050acc724284c34b57da70216a894633c2fa5d7d56a4dcdd9f3eb10b8db4d0"
	inputSHA       = "47686828bbb0cd4d273db3b2087311357ea853b767f86fd787953ed76d452ed6"
	savedSHA       = "b46d887b7f8e720a575cf307fabb11a3021b6259fa03b71b55c4655760d88b6f"
	gzipSHA        = "f080f9d1b6c0c50cf475086683b4879d3be3742832be6ec0914cefda356205f5"
	outputLimit    = 192 << 10
	fixedClock     = int64(1700000000)
	fixedHeader    = `{"alg":"RS256","typ":"JWT"}`
)

var requestIDs = [4]string{"fiction-lantern", "fiction-harbor", "fiction-moss", "fiction-cove"}
var candidateIDs = [8]string{"option-leeway", "option-issued-at", "option-exp-required", "option-nbf-required", "option-audience-any", "option-audience-all", "option-issuer", "option-subject"}
var fixtureIDs = [13]string{
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
var bindings = [4][5]int{{0, 1, 2, 3, 4}, {0, 1, 2, 3, 4}, {5, 2, 6, 7, 8}, {9, 10, 11, 12, 8}}

type preparedFixture struct {
	ID          string          `json:"id"`
	Claims      json.RawMessage `json:"claims"`
	Want        *bool           `json:"want_accept"`
	ClaimsJSON  string          `json:"claims_json"`
	ClaimsBytes int             `json:"claims_json_bytes"`
	ClaimsSHA   string          `json:"claims_sha256"`
}
type preparedRequest struct {
	ID        string            `json:"id"`
	Text      string            `json:"text"`
	Fixtures  []preparedFixture `json:"fixtures"`
	TextBytes int               `json:"text_bytes"`
	Words     int               `json:"ascii_whitespace_words"`
}
type preparedCandidate struct {
	ID             string `json:"id"`
	Option         string `json:"option"`
	Caption        string `json:"caption"`
	SourceLine     int    `json:"source_line"`
	CaptionOrigin  string `json:"caption_origin"`
	SourceURL      string `json:"source_url"`
	PreviousLine   int    `json:"previous_source_line"`
	LineConvention string `json:"source_line_convention"`
}
type preparation struct {
	Schema         string              `json:"schema"`
	Status         string              `json:"status"`
	Revision       string              `json:"source_revision"`
	Family         string              `json:"family"`
	Role           json.RawMessage     `json:"role"`
	Group          json.RawMessage     `json:"group_id"`
	RoleAdopted    bool                `json:"role_authority_adopted"`
	ProposedRole   string              `json:"proposed_future_role"`
	Common         json.RawMessage     `json:"common"`
	CandidateOrder string              `json:"candidate_order"`
	Candidates     []preparedCandidate `json:"candidates"`
	Requests       []preparedRequest   `json:"requests"`
	Counts         json.RawMessage     `json:"counts"`
	WantScope      string              `json:"want_scope"`
	GotToRecord    []string            `json:"got_to_record"`
	Acceptance     string              `json:"acceptance_predicate"`
	UnknownPolicy  []string            `json:"unknown_policy"`
	Excluded       []string            `json:"excluded"`
	Pending        []string            `json:"pending_before_execution"`
	Previous       json.RawMessage     `json:"previous_input"`
	Limitations    []string            `json:"finite_scope_limitations"`
	CacheProtocol  string              `json:"cached_control_protocol"`
	LiteralReview  json.RawMessage     `json:"independent_literal_review"`
}
type fixture struct {
	ID         string `json:"id"`
	Token      string `json:"token"`
	HeaderJSON string `json:"header_json"`
	ClaimsJSON string `json:"claims_json"`
	ClaimsSHA  string `json:"claims_sha256"`
}
type inputRequest struct {
	ID         string   `json:"id"`
	FixtureIDs []string `json:"fixture_ids"`
}
type observerInput struct {
	Schema    string         `json:"schema"`
	Revision  string         `json:"source_revision"`
	Clock     int64          `json:"clock_unix_seconds"`
	PublicKey string         `json:"public_key_pem"`
	Fixtures  []fixture      `json:"fixtures"`
	Requests  []inputRequest `json:"requests"`
}
type counts struct {
	Constructors int `json:"option_constructors"`
	Applications int `json:"option_applications"`
	Parsers      int `json:"new_parser"`
	Parses       int `json:"parse_with_claims"`
	Keys         int `json:"key_callbacks"`
	Times        int `json:"time_callbacks"`
}
type startup struct {
	OwnInit          int   `json:"observer_init_calls"`
	ImportedObserved *bool `json:"imported_package_init_count_observed"`
	IncludesInit     *bool `json:"api_counts_include_imported_init"`
	StartsAtMain     bool  `json:"timing_starts_at_main"`
}
type record struct {
	RequestID         string          `json:"request_id"`
	FixturePosition   int             `json:"fixture_position"`
	CandidatePosition int             `json:"candidate_position"`
	FixtureID         string          `json:"fixture_id"`
	CandidateID       string          `json:"candidate_id"`
	Status            string          `json:"status"`
	TokenPresent      *bool           `json:"token_present"`
	Valid             *bool           `json:"valid"`
	ErrNil            *bool           `json:"err_nil"`
	ParseReturned     *bool           `json:"parse_returned"`
	Errors            []string        `json:"error_classes"`
	Payload           json.RawMessage `json:"decoded_canonical_payload"`
	PayloadMatch      *bool           `json:"payload_match"`
	ZeroBefore        *bool           `json:"claims_zero_before"`
	InputUnchanged    *bool           `json:"claims_input_unchanged"`
	KeyUnchanged      *bool           `json:"key_input_unchanged"`
	Counts            counts          `json:"calls"`
	ElapsedNS         int64           `json:"elapsed_ns"`
	ParseElapsedNS    int64           `json:"parse_elapsed_ns"`
}
type observations struct {
	Schema                string   `json:"schema"`
	Revision              string   `json:"source_revision"`
	InputSHA              string   `json:"input_sha256"`
	KeySHA                string   `json:"public_key_der_sha256"`
	Clock                 int64    `json:"fixed_clock_unix_seconds"`
	Precision             int64    `json:"time_precision_ns"`
	Startup               startup  `json:"startup_scope"`
	Distinct              int      `json:"distinct_fixture_inputs"`
	Positions             int      `json:"fixture_positions"`
	Planned               int      `json:"planned_trials"`
	Records               []record `json:"records"`
	Counts                counts   `json:"calls"`
	CommonConstructors    []int    `json:"common_option_constructor_calls"`
	CandidateConstructors []int    `json:"candidate_option_constructor_calls"`
	Complete              bool     `json:"complete"`
	Unchanged             bool     `json:"input_unchanged"`
	MainNS                int64    `json:"main_elapsed_ns"`
}

// Alphabetical explicit columns produce the same canonical field order as the
// frozen literals, without a map or JWT RegisteredClaims marshaler.
type claims struct {
	Audience  []string `json:"aud,omitempty"`
	Expires   *int64   `json:"exp,omitempty"`
	Issued    *int64   `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	ID        string   `json:"jti,omitempty"`
	NotBefore *int64   `json:"nbf,omitempty"`
	Subject   string   `json:"sub,omitempty"`
}
type fixtureAudit struct {
	Position int      `json:"fixture_position"`
	ID       string   `json:"fixture_id"`
	Want     bool     `json:"literal_want_accept"`
	Got      string   `json:"got_accept_T_F_U"`
	State    string   `json:"observed_or_unknown"`
	Reasons  []string `json:"unknown_reasons"`
	Errors   []string `json:"normalized_error_classes"`
}
type candidateAudit struct {
	ID       string          `json:"candidate_id"`
	Caption  string          `json:"caption"`
	Match    string          `json:"match_literal_vector_T_F_U"`
	Unknown  int             `json:"unknown_fixture_count"`
	Fixtures [5]fixtureAudit `json:"fixtures"`
}
type requestAudit struct {
	ID         string            `json:"request_id"`
	Text       string            `json:"text"`
	Candidates [8]candidateAudit `json:"candidates"`
}
type result struct {
	Schema          string          `json:"schema"`
	Revision        string          `json:"source_revision"`
	PreparationSHA  string          `json:"preparation_sha256"`
	InputSHA        string          `json:"observer_input_sha256"`
	SavedSHA        string          `json:"checked_observation_sha256"`
	SavedPin        bool            `json:"first_saved_observation_pin_verified"`
	Fresh           bool            `json:"fresh_saved_comparison"`
	MatchesGolden   bool            `json:"matches_first_saved_predicates_without_timing"`
	TimingsCompared bool            `json:"timings_compared"`
	Records         int             `json:"records_checked"`
	Matrix          [4][8]string    `json:"match_matrix_T_F_U"`
	T               int             `json:"T"`
	F               int             `json:"F"`
	U               int             `json:"U"`
	UnknownRecords  int             `json:"unknown_records_preserved"`
	Requests        [4]requestAudit `json:"requests"`
	JWTCalls        int             `json:"checker_JWT_API_calls"`
	ModelCalls      int             `json:"checker_model_calls"`
	FitCalls        int             `json:"checker_Fit_calls"`
	Scope           string          `json:"scope"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func readBounded(p string, limit int) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, errors.New("file_read_failed")
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil || len(b) == 0 || len(b) > limit {
		return nil, errors.New("file_size_or_read_failed")
	}
	return b, nil
}
func readSaved(p string, first bool) ([]byte, error) {
	b, err := readBounded(p, outputLimit)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(p, ".gz") {
		if first && (len(b) != 4273 || digest(b) != gzipSHA) {
			return nil, errors.New("saved_gzip_pin_failed")
		}
		r := bytes.NewReader(b)
		g, err := gzip.NewReader(r)
		if err != nil {
			return nil, errors.New("gzip_header_failed")
		}
		g.Multistream(false)
		decoded, e := io.ReadAll(io.LimitReader(g, outputLimit+1))
		ce := g.Close()
		if e != nil || ce != nil || len(decoded) > outputLimit || r.Len() != 0 {
			return nil, errors.New("gzip_full_EOF_or_budget_failed")
		}
		b = decoded
	}
	if first && (len(b) != 116255 || digest(b) != savedSHA) {
		return nil, errors.New("first_saved_raw_pin_failed")
	}
	return b, nil
}

func scanJSON(d *json.Decoder, depth int, nodes *int) error {
	*nodes++
	if depth > 12 || *nodes > 16384 {
		return errors.New("json_budget")
	}
	t, e := d.Token()
	if e != nil {
		return e
	}
	if s, ok := t.(string); ok && len(s) > 4096 {
		return errors.New("json_string_budget")
	}
	if v, ok := t.(json.Delim); ok {
		switch v {
		case '{':
			var keys [64]string
			count := 0
			for d.More() {
				k, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := k.(string)
				if !ok || len(s) > 128 || count == len(keys) {
					return errors.New("json_key_budget")
				}
				for i := 0; i < count; i++ {
					if keys[i] == s {
						return errors.New("json_duplicate")
					}
				}
				keys[count] = s
				count++
				if e := scanJSON(d, depth+1, nodes); e != nil {
					return e
				}
			}
			t, e := d.Token()
			if e != nil || t != json.Delim('}') {
				return errors.New("json_object")
			}
		case '[':
			for d.More() {
				if e := scanJSON(d, depth+1, nodes); e != nil {
					return e
				}
			}
			t, e := d.Token()
			if e != nil || t != json.Delim(']') {
				return errors.New("json_array")
			}
		default:
			return errors.New("json_delimiter")
		}
	}
	return nil
}
func strict(b []byte, v any) error {
	if !validScalars(b) {
		return errors.New("json_scalar")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	n := 0
	if e := scanJSON(d, 0, &n); e != nil {
		return errors.New("json_shape_or_budget")
	}
	if _, e := d.Token(); e != io.EOF {
		return errors.New("json_trailing")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return errors.New("json_schema")
	}
	return nil
}

// Scalar boundary adapted from the already reviewed cmd/riido-hintpreview.
// encoding/json otherwise replaces invalid UTF-8 and lone UTF-16 escapes.
func validScalars(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	in := false
	hexValue := func(c byte) (uint16, bool) {
		switch {
		case c >= '0' && c <= '9':
			return uint16(c - '0'), true
		case c >= 'a' && c <= 'f':
			return uint16(c - 'a' + 10), true
		case c >= 'A' && c <= 'F':
			return uint16(c - 'A' + 10), true
		}
		return 0, false
	}
	unit := func(p int) (uint16, bool) {
		if p+4 > len(b) {
			return 0, false
		}
		var n uint16
		for i := p; i < p+4; i++ {
			x, ok := hexValue(b[i])
			if !ok {
				return 0, false
			}
			n = n*16 + x
		}
		return n, true
	}
	for i := 0; i < len(b); i++ {
		if b[i] == '"' {
			in = !in
			continue
		}
		if !in || b[i] != '\\' {
			continue
		}
		i++
		if i >= len(b) {
			return false
		}
		if b[i] != 'u' {
			continue
		}
		u, ok := unit(i + 1)
		if !ok {
			return false
		}
		i += 4
		if u >= 0xdc00 && u <= 0xdfff {
			return false
		}
		if u >= 0xd800 && u <= 0xdbff {
			if i+6 >= len(b) || b[i+1] != '\\' || b[i+2] != 'u' {
				return false
			}
			v, ok := unit(i + 3)
			if !ok || v < 0xdc00 || v > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}
func canonical(b []byte) ([]byte, error) {
	var c claims
	if len(b) > 256 || strict(b, &c) != nil {
		return nil, errors.New("claims_shape")
	}
	return json.Marshal(c)
}

func loadInputs(preparationPath, inputPath string) (preparation, observerInput, string, error) {
	var p preparation
	var in observerInput
	pb, e := readBounded(preparationPath, 32<<10)
	if e != nil || len(pb) != 23123 || digest(pb) != preparationSHA {
		return p, in, "", errors.New("literal_preparation_pin_failed")
	}
	ib, e := readBounded(inputPath, 64<<10)
	if e != nil || len(ib) != 13474 || digest(ib) != inputSHA {
		return p, in, "", errors.New("observer_input_pin_failed")
	}
	if strict(pb, &p) != nil || strict(ib, &in) != nil {
		return p, in, "", errors.New("input_schema_failed")
	}
	if p.Schema != "riido-JWT-input-preparation-review-v2" || p.Revision != revision || in.Schema != "riido-jwt-public-observer-input-v1" || in.Revision != revision || in.Clock != fixedClock || len(p.Candidates) != 8 || len(p.Requests) != 4 || len(in.Requests) != 4 || len(in.Fixtures) != 13 {
		return p, in, "", errors.New("frozen_input_shape_failed")
	}
	key, rest := pem.Decode([]byte(in.PublicKey))
	if key == nil || key.Type != "PUBLIC KEY" || len(key.Headers) != 0 || len(bytes.TrimSpace(rest)) != 0 || len(key.Bytes) > 512 {
		return p, in, "", errors.New("public_key_envelope_failed")
	}
	for c := 0; c < 8; c++ {
		if p.Candidates[c].ID != candidateIDs[c] || p.Candidates[c].Caption == "" {
			return p, in, "", errors.New("candidate_order_failed")
		}
	}
	for f := 0; f < 13; f++ {
		x := in.Fixtures[f]
		b, e := canonical([]byte(x.ClaimsJSON))
		parts := strings.Split(x.Token, ".")
		if x.ID != fixtureIDs[f] || x.ClaimsSHA != fixtureIDs[f] || x.HeaderJSON != fixedHeader || e != nil || !bytes.Equal(b, []byte(x.ClaimsJSON)) || digest(b) != x.ClaimsSHA || len(parts) != 3 || len(x.Token) > 1024 {
			return p, in, "", errors.New("fixture_pin_failed")
		}
		head, e0 := base64.RawURLEncoding.Strict().DecodeString(parts[0])
		body, e1 := base64.RawURLEncoding.Strict().DecodeString(parts[1])
		sig, e2 := base64.RawURLEncoding.Strict().DecodeString(parts[2])
		if e0 != nil || e1 != nil || e2 != nil || string(head) != fixedHeader || !bytes.Equal(body, b) || len(sig) != 256 {
			return p, in, "", errors.New("signed_public_fixture_envelope_failed")
		}
	}
	for r := 0; r < 4; r++ {
		a, b := p.Requests[r], in.Requests[r]
		if a.ID != requestIDs[r] || b.ID != requestIDs[r] || len(a.Fixtures) != 5 || len(b.FixtureIDs) != 5 || len(a.Text) != a.TextBytes {
			return p, in, "", errors.New("request_binding_failed")
		}
		for f := 0; f < 5; f++ {
			x, y := a.Fixtures[f], in.Fixtures[bindings[r][f]]
			if x.Want == nil || x.ClaimsSHA != y.ID || b.FixtureIDs[f] != y.ID || x.ClaimsJSON != y.ClaimsJSON || x.ClaimsBytes != len(x.ClaimsJSON) {
				return p, in, "", errors.New("literal_Want_or_fixture_binding_failed")
			}
		}
	}
	return p, in, digest(key.Bytes), nil
}
func add(a, b counts) counts {
	return counts{a.Constructors + b.Constructors, a.Applications + b.Applications, a.Parsers + b.Parsers, a.Parses + b.Parses, a.Keys + b.Keys, a.Times + b.Times}
}
func qualify(o observations, in observerInput, keySHA string) error {
	if o.Schema != "riido-jwt-direct-observer-output-v1" || o.Revision != revision || o.InputSHA != inputSHA || o.KeySHA != keySHA || o.Clock != fixedClock || o.Precision != 1000000000 || !o.Complete || !o.Unchanged || o.Distinct != 13 || o.Positions != 20 || o.Planned != 160 || len(o.Records) != 160 {
		return errors.New("complete_observation_integrity_failed")
	}
	if o.Startup.OwnInit != 1 || o.Startup.ImportedObserved == nil || *o.Startup.ImportedObserved || o.Startup.IncludesInit == nil || *o.Startup.IncludesInit || !o.Startup.StartsAtMain || len(o.CommonConstructors) != 3 || len(o.CandidateConstructors) != 8 {
		return errors.New("startup_or_constructor_shape_failed")
	}
	for _, n := range o.CommonConstructors {
		if n != 160 {
			return errors.New("common_constructor_count_failed")
		}
	}
	for _, n := range o.CandidateConstructors {
		if n != 20 {
			return errors.New("candidate_constructor_count_failed")
		}
	}
	want := counts{640, 640, 160, 160, 160, 160}
	var sum counts
	for r := 0; r < 4; r++ {
		for f := 0; f < 5; f++ {
			for c := 0; c < 8; c++ {
				x := o.Records[(r*5+f)*8+c]
				if x.RequestID != requestIDs[r] || x.FixturePosition != f || x.CandidatePosition != c || x.CandidateID != candidateIDs[c] || x.FixtureID != in.Requests[r].FixtureIDs[f] {
					return errors.New("record_order_or_binding_failed")
				}
				if x.Counts != (counts{4, 4, 1, 1, 1, 1}) {
					return errors.New("per_record_calls_failed")
				}
				sum = add(sum, x.Counts)
			}
		}
	}
	if sum != want || o.Counts != want {
		return errors.New("complete_API_counter_integrity_failed")
	}
	return nil
}

func isTrue(b *bool) bool { return b != nil && *b }
func recordState(r record, f fixture) (string, []string) {
	reasons := make([]string, 0, 8)
	if r.Errors == nil {
		reasons = append(reasons, "normalized_error_channel_missing")
	}
	if !isTrue(r.TokenPresent) {
		reasons = append(reasons, "verified_token_missing")
	}
	if r.Status != "observed" {
		reasons = append(reasons, "observer_status_unknown")
	}
	if r.TokenPresent == nil || r.Valid == nil || r.ErrNil == nil || r.ParseReturned == nil {
		reasons = append(reasons, "raw_parse_fields_missing")
	}
	if !isTrue(r.ParseReturned) {
		reasons = append(reasons, "parse_not_returned")
	}
	if r.ErrNil != nil && r.Valid != nil && r.TokenPresent != nil {
		if (*r.ErrNil && (!*r.TokenPresent || !*r.Valid)) || (!*r.ErrNil && *r.Valid) {
			reasons = append(reasons, "raw_parse_contradiction")
		}
	}
	if !isTrue(r.ZeroBefore) || !isTrue(r.InputUnchanged) || !isTrue(r.KeyUnchanged) {
		reasons = append(reasons, "claims_or_key_ownership_unknown")
	}
	payload, e := canonical(r.Payload)
	if e != nil || !bytes.Equal(payload, []byte(f.ClaimsJSON)) || digest(payload) != f.ClaimsSHA || !isTrue(r.PayloadMatch) {
		reasons = append(reasons, "decoded_payload_or_hash_contradiction")
	}
	policy := false
	wrapper := false
	seen := make([]string, 0, 21)
	if len(r.Errors) > 21 {
		reasons = append(reasons, "normalized_error_budget")
	}
	for _, code := range r.Errors {
		duplicate := false
		for _, old := range seen {
			if old == code {
				duplicate = true
			}
		}
		if duplicate {
			reasons = append(reasons, "duplicate_normalized_error")
			continue
		}
		seen = append(seen, code)
		switch code {
		case "required_claim_missing", "invalid_audience", "expired", "used_before_issued", "invalid_issuer", "invalid_subject", "not_valid_yet":
			policy = true
		case "invalid_claims":
			wrapper = true
		default:
			reasons = append(reasons, "untrusted_or_unknown_error_class")
		}
	}
	if r.ErrNil != nil {
		if *r.ErrNil && len(r.Errors) != 0 {
			reasons = append(reasons, "nil_error_with_classes")
		}
		if !*r.ErrNil && (!policy || wrapper && !policy) {
			reasons = append(reasons, "error_without_known_policy_cause")
		}
	}
	if len(reasons) != 0 {
		return "U", reasons
	}
	if isTrue(r.TokenPresent) && isTrue(r.Valid) && isTrue(r.ErrNil) {
		return "T", reasons
	}
	return "F", reasons
}
func aggregate(got [5]string, want [5]bool) (string, int) {
	unknown := 0
	mismatch := false
	for f, g := range got {
		if g == "U" {
			unknown++
			continue
		}
		if (g == "T") != want[f] {
			mismatch = true
		}
	}
	if mismatch {
		return "F", unknown
	}
	if unknown != 0 {
		return "U", unknown
	}
	return "T", 0
}
func publicStatus(s string) string {
	if s == "observed" {
		return s
	}
	return "unknown_observer"
}
func publicErrorClasses(a []string) []string {
	out := make([]string, 0, len(a))
	for _, x := range a {
		switch x {
		case "required_claim_missing", "invalid_audience", "expired", "used_before_issued", "invalid_issuer", "invalid_subject", "not_valid_yet", "invalid_claims":
			out = append(out, x)
		default:
			out = append(out, "unknown_error_class")
		}
	}
	return out
}
func sameClasses(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, x := range a {
		found := false
		for _, y := range b {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func audit(p preparation, in observerInput, o observations) result {
	out := result{Schema: "riido-JWT-saved-replay-v1", Revision: revision, PreparationSHA: preparationSHA, InputSHA: inputSHA, Records: 160, Scope: "Saved finite public fixtures only; literal Wants are read unchanged. F means a trustworthy mismatch even if U traces remain; T requires all five known matches. No signature verification, JWT calls, policy sufficiency, roles, training or performance claim; imported startup remains unattributed. Timings are not compared."}
	for r := 0; r < 4; r++ {
		out.Requests[r].ID = p.Requests[r].ID
		out.Requests[r].Text = p.Requests[r].Text
		for c := 0; c < 8; c++ {
			a := candidateAudit{ID: candidateIDs[c], Caption: p.Candidates[c].Caption}
			var got [5]string
			var want [5]bool
			for f := 0; f < 5; f++ {
				x := o.Records[(r*5+f)*8+c]
				g, reasons := recordState(x, in.Fixtures[bindings[r][f]])
				got[f] = g
				want[f] = *p.Requests[r].Fixtures[f].Want
				a.Fixtures[f] = fixtureAudit{f, x.FixtureID, want[f], g, publicStatus(x.Status), reasons, publicErrorClasses(x.Errors)}
				if g == "U" {
					out.UnknownRecords++
				}
			}
			a.Match, a.Unknown = aggregate(got, want)
			out.Requests[r].Candidates[c] = a
			out.Matrix[r][c] = a.Match
			switch a.Match {
			case "T":
				out.T++
			case "F":
				out.F++
			case "U":
				out.U++
			}
		}
	}
	return out
}
func samePredicates(a, b result) bool {
	if a.Matrix != b.Matrix || a.UnknownRecords != b.UnknownRecords {
		return false
	}
	for r := 0; r < 4; r++ {
		for c := 0; c < 8; c++ {
			for f := 0; f < 5; f++ {
				x, y := a.Requests[r].Candidates[c].Fixtures[f], b.Requests[r].Candidates[c].Fixtures[f]
				if x.Got != y.Got || x.State != y.State || len(x.Reasons) != len(y.Reasons) || !sameClasses(x.Errors, y.Errors) {
					return false
				}
				for i := range x.Reasons {
					if x.Reasons[i] != y.Reasons[i] {
						return false
					}
				}
			}
		}
	}
	return true
}

func run(args []string) (result, error) {
	var empty result
	f := flag.NewFlagSet("jwt-saved-replay", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	p := f.String("input", base+"INPUTS.reviewed.v2.json", "unchanged literal preparation")
	inPath := f.String("observer-input", base+"OBSERVER-INPUT.actual.public.v1.json", "unchanged public observer envelope")
	saved := f.String("saved", "", "fresh JSON or gzip stdout to compare with the pinned first saved observation")
	if f.Parse(args) != nil || f.NArg() != 0 {
		return empty, errors.New("arguments_invalid")
	}
	prep, in, key, e := loadInputs(*p, *inPath)
	if e != nil {
		return empty, e
	}
	first, e := readSaved(base+"OBSERVATIONS.actual.public.v1.json.gz", true)
	if e != nil {
		return empty, e
	}
	var old observations
	if strict(first, &old) != nil {
		return empty, errors.New("first_observation_schema_failed")
	}
	if e = qualify(old, in, key); e != nil {
		return empty, e
	}
	gold := audit(prep, in, old)
	out := gold
	checked := first
	if *saved != "" {
		checked, e = readSaved(*saved, false)
		if e != nil {
			return empty, e
		}
		var fresh observations
		if strict(checked, &fresh) != nil {
			return empty, errors.New("fresh_observation_schema_failed")
		}
		if e = qualify(fresh, in, key); e != nil {
			return empty, e
		}
		out = audit(prep, in, fresh)
		out.Fresh = true
	}
	out.SavedSHA = digest(checked)
	out.SavedPin = true
	out.MatchesGolden = samePredicates(gold, out)
	return out, nil
}
func main() {
	out, e := run(os.Args[1:])
	if e != nil {
		b, _ := json.Marshal(struct {
			Schema string `json:"schema"`
			Code   string `json:"code"`
		}{"riido-JWT-saved-replay-failure-v1", e.Error()})
		_, _ = os.Stderr.Write(append(b, '\n'))
		os.Exit(1)
	}
	b, e := json.Marshal(out)
	if e != nil || len(b) > 64<<10 {
		_, _ = os.Stderr.Write([]byte("{\"code\":\"output_encoding_or_budget_failed\"}\n"))
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(append(b, '\n'))
	if !out.MatchesGolden {
		os.Exit(1)
	}
}
