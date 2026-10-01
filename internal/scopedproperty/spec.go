// Package scopedproperty prepares finite property contracts and draft captions.
// Preparation never executes candidate code, labels outcomes or ranks text.
package scopedproperty

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"

	"github.com/teamswyg/laya-tools/internal/typedbehavior"
)

const (
	QuotedCommaProperty   = "quoted-comma-content-v1"
	OutsideEscapeProperty = "outside-escaped-comma-v1"
	SyntaxZeroProperty    = "syntax-zero-output-v1"
	PendingReview         = "pending"
	PropertyVersion       = 1
)

// These are observations, not control roles. In particular, a whole-function
// wrong source can agree with one property's independently declared literals.
type Observation = typedbehavior.PropertyQuotedObservation

type Literal struct {
	Input string      `json:"input"`
	Want  Observation `json:"want"`
}

// Literals are a fixed capacity of five; only LiteralCount entries are in scope.
// Remaining entries are zero. Tokens includes all eight slots, not a prefix.
// Exact error-kind comparison includes successful OK (zero), not just errors.
type PropertyDefinition struct {
	ID                     string                            `json:"id"`
	Version                int                               `json:"version"`
	Scope                  string                            `json:"scope"`
	Exclusions             string                            `json:"exclusions"`
	ObservationFields      [4]string                         `json:"observation_fields"`
	NoPanic                bool                              `json:"no_panic"`
	RequireExactErrorKind  bool                              `json:"require_exact_error_kind"`
	ExpectedErrorKind      typedbehavior.PropertyQuotedError `json:"expected_error_kind"`
	WholeArrayZeroRequired bool                              `json:"whole_array_zero_required"`
	UnknownPolicy          string                            `json:"unknown_policy"`
	CaptionReview          string                            `json:"caption_review"`
	LiteralCount           int                               `json:"literal_count"`
	Literals               [5]Literal                        `json:"literals"`
	TruthTableSHA256       string                            `json:"truth_table_sha256"`
}

// The closed registry is returned by value. No input can introduce code,
// callbacks or an expected role, or change the original source order.
func SourceIDs() [4]string {
	return [4]string{"quoted-correct", "quoted-literal-comma", "quoted-inside-escape", "quoted-partial-error"}
}

func OriginalParentIDs() [4]string {
	return [4]string{"typed56b-p05-1", "typed56b-p05-2", "typed56b-p05-3", "typed56b-p05-4"}
}

// Definitions transfers the original independent literals without invoking a
// candidate or consulting whole-function roles or existing acceptable sets.
// Each return owns its arrays and contains no writable shared slices.
func Definitions() [3]PropertyDefinition {
	out := [3]PropertyDefinition{
		{
			ID: QuotedCommaProperty, Version: PropertyVersion,
			Scope:        "Only the two listed valid ASCII literals without backslashes, within128 raw bytes,8 tokens and8 decoded bytes per token. Preserve commas inside balanced quotes, strip quotes and split outside commas. Require exact successful Error, Count, all8 Tokens and no panic; this is finite evidence only.",
			Exclusions:   "All unlisted inputs, escapes, error behavior, non-ASCII and bound overflow are outside this property. No whole-function or general-input certificate.",
			LiteralCount: 2,
			Literals: [5]Literal{
				{Input: `"a,b",c`, Want: Observation{Tokens: [8]string{"a,b", "c"}, Count: 2, Error: typedbehavior.PropertyQuotedOK}},
				{Input: `ab"c,d"ef`, Want: Observation{Tokens: [8]string{"abc,def"}, Count: 1, Error: typedbehavior.PropertyQuotedOK}},
			},
		},
		{
			ID: OutsideEscapeProperty, Version: PropertyVersion,
			Scope:        "Only the two listed valid ASCII literals without quotes or trailing escapes, within128 raw bytes,8 tokens and8 decoded bytes per token. Backslash consumes the next byte as content; an unescaped comma splits. Require exact successful Error, Count, all8 Tokens and no panic; this is finite evidence only.",
			Exclusions:   "All unlisted inputs, escapes inside quotes, error behavior, non-ASCII and bound overflow are outside this property. No whole-function or general-input certificate.",
			LiteralCount: 2,
			Literals: [5]Literal{
				{Input: `a\,b,c`, Want: Observation{Tokens: [8]string{"a,b", "c"}, Count: 2, Error: typedbehavior.PropertyQuotedOK}},
				{Input: `a\\,b`, Want: Observation{Tokens: [8]string{`a\`, "b"}, Count: 2, Error: typedbehavior.PropertyQuotedOK}},
			},
		},
		{
			ID: SyntaxZeroProperty, Version: PropertyVersion,
			Scope:             "Only the five listed ASCII unclosed-quote/trailing-escape literals, within128 raw bytes,8 tokens and8 decoded bytes per token. Every literal independently requires exact syntax error, Count0, the entire zero8-string array and no panic. An omitted error is a mismatch, never a vacuous pass; this is finite evidence only.",
			Exclusions:        "All unlisted inputs, successful parsing, ASCII errors and bounds errors are outside this property. Clearing initial errors does not certify clearing later syntax errors or all errors.",
			ExpectedErrorKind: typedbehavior.PropertyQuotedSyntax, WholeArrayZeroRequired: true,
			LiteralCount: 5,
			Literals: [5]Literal{
				{Input: `"a,b`, Want: Observation{Error: typedbehavior.PropertyQuotedSyntax}},
				{Input: `a\`, Want: Observation{Error: typedbehavior.PropertyQuotedSyntax}},
				{Input: `a,b\`, Want: Observation{Error: typedbehavior.PropertyQuotedSyntax}},
				{Input: `"a\`, Want: Observation{Error: typedbehavior.PropertyQuotedSyntax}},
				{Input: `a,"b`, Want: Observation{Error: typedbehavior.PropertyQuotedSyntax}},
			},
		},
	}
	for i := range out {
		out[i].ObservationFields = [4]string{"error", "count", "tokens", "panicked"}
		out[i].NoPanic, out[i].RequireExactErrorKind = true, true
		out[i].UnknownPolicy = "Unlisted/unsupported inputs, incomplete or pending captions, and observer failures remain unknown; safely observed candidate panic under this no-panic contract is a mismatch. Corrupt bindings stop preparation."
		out[i].CaptionReview = PendingReview
		out[i].TruthTableSHA256 = tableSHA(out[i])
	}
	return out
}

func tableSHA(d PropertyDefinition) string {
	d.TruthTableSHA256 = ""
	raw, _ := json.Marshal(d)
	return sha(raw)
}

func sha(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

// SourceArtifacts binds these two exact compiled preparation files separately
// from the immutable old eight typed files and the new observation adapter.
//
//go:embed spec.go dataset.go
var preparationSources embed.FS

func SourceArtifacts() []typedbehavior.SourceArtifact {
	names := [2]string{"spec.go", "dataset.go"}
	out := make([]typedbehavior.SourceArtifact, len(names))
	for i, name := range names {
		raw, _ := preparationSources.ReadFile(name)
		out[i] = typedbehavior.SourceArtifact{Name: name, SHA256: sha(raw)}
	}
	return out
}
