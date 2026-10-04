// SPDX-License-Identifier: Apache-2.0
// This unexecuted preparation program emits public fictional signature fixtures.
// It does not import JWT, evaluate a candidate, inspect Wants, or assign roles.
package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"os"
	"runtime"
	"strings"
	"unicode/utf8"
)

const (
	inputSHA        = "1edde7ed9945af038d0060122b7503d882790e1142707396cf6e44e5d67e1cf7"
	canonicalHeader = `{"alg":"RS256","typ":"JWT"}`
	maxInputBytes   = 32 << 10
	maxClaimsBytes  = 256
	maxTokenBytes   = 1024
	maxPublicBytes  = 1024
	maxOutputBytes  = 32 << 10
	requestCount    = 4
	fixturesEach    = 5
	positionCount   = requestCount * fixturesEach
	uniqueCount     = 13
	rsaBits         = 2048
	rsaExponent     = 65537
)

//go:embed INPUTS.draft.json
var frozenInput []byte

var expectedClaimsSHA = [uniqueCount]string{
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

// Only public claim preparation fields are projected. The full snapshot hash
// guards every byte, including all unchanged Wants and unrelated metadata.
type inputPacket struct {
	Schema     string            `json:"schema"`
	Common     inputCommon       `json:"common"`
	Candidates []json.RawMessage `json:"candidates"`
	Requests   []inputRequest    `json:"requests"`
}

type inputCommon struct {
	ClockUnixSeconds int64  `json:"clock_unix_seconds"`
	HeaderJSON       string `json:"header_json"`
}

type inputRequest struct {
	Fixtures []inputFixture `json:"fixtures"`
}

type inputFixture struct {
	Claims          json.RawMessage `json:"claims"`
	ClaimsJSON      string          `json:"claims_json"`
	ClaimsJSONBytes int             `json:"claims_json_bytes"`
	ClaimsSHA       string          `json:"claims_sha256"`
}

// Declaration order matches the frozen canonical JSON key ordering. Pointers
// preserve absent versus numeric-zero timestamps; no interpretation is made.
type literalClaims struct {
	Aud []string `json:"aud,omitempty"`
	Exp *int64   `json:"exp,omitempty"`
	Iat *int64   `json:"iat,omitempty"`
	Iss string   `json:"iss,omitempty"`
	Nbf *int64   `json:"nbf,omitempty"`
	Sub string   `json:"sub,omitempty"`
}

type positionBinding struct {
	RequestIndex int    `json:"request_index"`
	FixtureIndex int    `json:"fixture_index"`
	InputIndex   int    `json:"input_index"`
	FixtureID    string `json:"fixture_id"`
}

type preparedInput struct {
	Index          int    `json:"index"`
	ID             string `json:"id"`
	HeaderJSON     string `json:"header_json"`
	ClaimsJSON     string `json:"claims_json"`
	ClaimsBytes    int    `json:"claims_json_bytes"`
	ClaimsSHA      string `json:"claims_sha256"`
	Token          string `json:"token"`
	TokenBytes     int    `json:"token_bytes"`
	TokenSHA       string `json:"token_sha256"`
	SignatureBytes int    `json:"signature_bytes"`
}

type counts struct {
	GenerateKeyCalls  int `json:"rsa_generate_key_explicit_calls"`
	SignCalls         int `json:"rsa_sign_pkcs1v15_explicit_calls"`
	SelfVerifyCalls   int `json:"rsa_verify_pkcs1v15_preparation_calls"`
	JWTCalls          int `json:"JWT_calls"`
	JWTInit           int `json:"JWT_package_init"`
	CandidateOutcomes int `json:"candidate_outcomes"`
	Models            int `json:"models"`
	Fits              int `json:"Fits"`
	Labels            int `json:"labels"`
	RoleAssignments   int `json:"role_assignments"`
	PaidCalls         int `json:"paid_calls"`
	GPUCalls          int `json:"GPU_calls"`
}

type publicOutput struct {
	Schema               string                         `json:"schema"`
	Scope                string                         `json:"scope"`
	InputSHA             string                         `json:"input_snapshot_sha256"`
	HeaderJSON           string                         `json:"header_json"`
	HeaderSegment        string                         `json:"header_base64url"`
	KeyBits              int                            `json:"RSA_public_modulus_bits"`
	KeyExponent          int                            `json:"RSA_public_exponent"`
	PublicKeyPEM         string                         `json:"public_key_pem"`
	PublicKeyPEMSHA      string                         `json:"public_key_pem_sha256"`
	PublicKeyFingerprint string                         `json:"public_key_pkix_der_sha256"`
	Inputs               [uniqueCount]preparedInput     `json:"fixtures"`
	Bindings             [positionCount]positionBinding `json:"position_bindings"`
	Counts               counts                         `json:"counts"`
}

type failureReceipt struct {
	Schema string `json:"schema"`
	Code   string `json:"code"`
	Counts counts `json:"counts"`
}

func main() {
	var actual counts
	// A recovered panic prints no panic value, stack, arguments or key material.
	// Fatal runtime failures and external dumps are outside this guarantee.
	defer func() {
		if recover() != nil {
			emitFailure("preparation_panic", actual)
			os.Exit(1)
		}
	}()
	if len(os.Args) != 1 {
		emitFailure("unexpected_arguments", actual)
		os.Exit(1)
	}
	output, code := prepare(&actual)
	if code != "" {
		emitFailure(code, actual)
		os.Exit(1)
	}
	// Nothing reaches stdout until every fixture has passed self-verification
	// and the entire public-only payload has passed its size check.
	n, err := os.Stdout.Write(output)
	if err != nil || n != len(output) {
		emitFailure("public_output_write_failed", actual)
		os.Exit(1)
	}
}

func prepare(actual *counts) ([]byte, string) {
	inputs, bindings, code := loadInputs()
	if code != "" {
		return nil, code
	}
	actual.GenerateKeyCalls++
	// Exactly one explicit attempt: no retry, deterministic seed, imported
	// private fixture, key serialization, environment dump, or private logging.
	privateKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if privateKey != nil {
		defer wipePrivateKey(privateKey)
	}
	if err != nil || privateKey == nil {
		return nil, "RSA_generation_failed"
	}
	if privateKey.N == nil || privateKey.N.BitLen() != rsaBits || privateKey.E != rsaExponent {
		return nil, "RSA_public_shape_mismatch"
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil || len(publicDER) > maxPublicBytes {
		return nil, "public_PKIX_encoding_failed"
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	if len(publicPEM) == 0 || len(publicPEM) > maxPublicBytes {
		return nil, "public_PEM_encoding_failed"
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(canonicalHeader))
	for i := range inputs {
		message := header + "." + base64.RawURLEncoding.EncodeToString([]byte(inputs[i].ClaimsJSON))
		digest := sha256.Sum256([]byte(message))
		actual.SignCalls++
		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
		if err != nil || len(signature) != rsaBits/8 {
			return nil, "RSA_signing_failed"
		}
		token := message + "." + base64.RawURLEncoding.EncodeToString(signature)
		if len(token) > maxTokenBytes {
			return nil, "token_size_exceeded"
		}
		// Round-trip the actual public token and verify its signature. This is
		// preparation self-checking with the same standard library, not an
		// independent candidate oracle or a claim-policy approval.
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			return nil, "token_structure_mismatch"
		}
		decodedHeader, hErr := base64.RawURLEncoding.Strict().DecodeString(parts[0])
		decodedClaims, cErr := base64.RawURLEncoding.Strict().DecodeString(parts[1])
		decodedSignature, sErr := base64.RawURLEncoding.Strict().DecodeString(parts[2])
		if hErr != nil || cErr != nil || sErr != nil || string(decodedHeader) != canonicalHeader || string(decodedClaims) != inputs[i].ClaimsJSON || !bytes.Equal(decodedSignature, signature) {
			return nil, "token_roundtrip_mismatch"
		}
		checkDigest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
		actual.SelfVerifyCalls++
		if rsa.VerifyPKCS1v15(&privateKey.PublicKey, crypto.SHA256, checkDigest[:], decodedSignature) != nil {
			return nil, "RSA_signature_selfcheck_failed"
		}
		inputs[i].Token = token
		inputs[i].TokenBytes = len(token)
		inputs[i].TokenSHA = hashString(token)
		inputs[i].SignatureBytes = len(signature)
	}
	if actual.GenerateKeyCalls != 1 || actual.SignCalls != uniqueCount || actual.SelfVerifyCalls != uniqueCount {
		return nil, "explicit_crypto_count_mismatch"
	}
	out := publicOutput{
		Schema:               "riido-jwt-public-signature-fixtures-v1",
		Scope:                "fictional_fixture_preparation_not_candidate_outcomes",
		InputSHA:             inputSHA,
		HeaderJSON:           canonicalHeader,
		HeaderSegment:        header,
		KeyBits:              rsaBits,
		KeyExponent:          rsaExponent,
		PublicKeyPEM:         string(publicPEM),
		PublicKeyPEMSHA:      hashBytes(publicPEM),
		PublicKeyFingerprint: hashBytes(publicDER),
		Inputs:               inputs,
		Bindings:             bindings,
		Counts:               *actual,
	}
	encoded, err := json.Marshal(out)
	if err != nil || len(encoded)+1 > maxOutputBytes {
		return nil, "public_output_encoding_or_size_failed"
	}
	return append(encoded, '\n'), ""
}

func loadInputs() ([uniqueCount]preparedInput, [positionCount]positionBinding, string) {
	var inputs [uniqueCount]preparedInput
	var bindings [positionCount]positionBinding
	if len(frozenInput) > maxInputBytes || !utf8.Valid(frozenInput) || hashBytes(frozenInput) != inputSHA {
		return inputs, bindings, "frozen_input_pin_or_size_mismatch"
	}
	var packet inputPacket
	decoder := json.NewDecoder(bytes.NewReader(frozenInput))
	if decoder.Decode(&packet) != nil {
		return inputs, bindings, "frozen_input_decode_failed"
	}
	var extra json.RawMessage
	if decoder.Decode(&extra) != io.EOF {
		return inputs, bindings, "frozen_input_trailing_data"
	}
	if packet.Schema != "riido-jwt-finite-cost-screen-input-draft-v1" || packet.Common.HeaderJSON != canonicalHeader || packet.Common.ClockUnixSeconds != 1700000000 || len(packet.Candidates) != 8 || len(packet.Requests) != requestCount {
		return inputs, bindings, "frozen_input_shape_mismatch"
	}
	live, position := 0, 0
	for r, request := range packet.Requests {
		if len(request.Fixtures) != fixturesEach {
			return inputs, bindings, "fixture_count_mismatch"
		}
		for f, fixture := range request.Fixtures {
			if len(fixture.ClaimsJSON) == 0 || len(fixture.ClaimsJSON) > maxClaimsBytes || len(fixture.ClaimsJSON) != fixture.ClaimsJSONBytes || hashString(fixture.ClaimsJSON) != fixture.ClaimsSHA {
				return inputs, bindings, "canonical_claims_pin_or_size_mismatch"
			}
			if !canonicalClaimsEqual(fixture.ClaimsJSON, fixture.Claims) {
				return inputs, bindings, "canonical_claims_projection_mismatch"
			}
			index := -1
			for i := 0; i < live; i++ {
				if inputs[i].ClaimsJSON == fixture.ClaimsJSON {
					index = i
					break
				}
			}
			if index < 0 {
				if live == uniqueCount || fixture.ClaimsSHA != expectedClaimsSHA[live] {
					return inputs, bindings, "distinct_claim_order_or_count_mismatch"
				}
				index = live
				inputs[live] = preparedInput{Index: live, ID: fixture.ClaimsSHA, HeaderJSON: canonicalHeader, ClaimsJSON: fixture.ClaimsJSON, ClaimsBytes: len(fixture.ClaimsJSON), ClaimsSHA: fixture.ClaimsSHA}
				live++
			}
			if position == positionCount {
				return inputs, bindings, "position_limit_exceeded"
			}
			bindings[position] = positionBinding{RequestIndex: r, FixtureIndex: f, InputIndex: index, FixtureID: inputs[index].ID}
			position++
		}
	}
	if live != uniqueCount || position != positionCount {
		return inputs, bindings, "distinct_claim_or_position_count_mismatch"
	}
	return inputs, bindings, ""
}

func canonicalClaimsEqual(canonical string, original json.RawMessage) bool {
	for i := 0; i < len(canonical); i++ {
		if canonical[i] >= utf8.RuneSelf {
			return false
		}
	}
	var fromOriginal literalClaims
	decoder := json.NewDecoder(bytes.NewReader(original))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&fromOriginal) != nil {
		return false
	}
	var extra json.RawMessage
	if decoder.Decode(&extra) != io.EOF {
		return false
	}
	encoded, err := json.Marshal(fromOriginal)
	return err == nil && string(encoded) == canonical
}

func hashString(value string) string { return hashBytes([]byte(value)) }

func hashBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func emitFailure(code string, actual counts) {
	// All codes are fixed source literals. Error values, keys, input bytes and
	// local paths never enter a failure receipt. No retry or partial success.
	receipt := failureReceipt{"riido-jwt-signature-preparation-failure-v1", code, actual}
	encoded, err := json.Marshal(receipt)
	if err == nil && len(encoded)+1 <= 1024 {
		_, _ = os.Stderr.Write(append(encoded, '\n'))
	}
}

func wipeBigInt(value *big.Int) {
	if value == nil {
		return
	}
	words := value.Bits()
	for i := range words {
		words[i] = 0
	}
	value.SetInt64(0)
	runtime.KeepAlive(words)
}

func wipePrivateKey(key *rsa.PrivateKey) {
	if key == nil {
		return
	}
	// Best effort only: Go's allocator, compiler, stack and opaque FIPS state
	// may retain copies. This is not a secure-erasure or no-dump promise.
	wipeBigInt(key.D)
	for _, prime := range key.Primes {
		wipeBigInt(prime)
	}
	wipeBigInt(key.Precomputed.Dp)
	wipeBigInt(key.Precomputed.Dq)
	wipeBigInt(key.Precomputed.Qinv)
	for i := range key.Precomputed.CRTValues {
		wipeBigInt(key.Precomputed.CRTValues[i].Exp)
		wipeBigInt(key.Precomputed.CRTValues[i].Coeff)
		wipeBigInt(key.Precomputed.CRTValues[i].R)
	}
	wipeBigInt(key.N)
	clear(key.Primes)
	clear(key.Precomputed.CRTValues)
	*key = rsa.PrivateKey{}
	runtime.KeepAlive(key)
}
