package typedbehavior

import _ "embed"

// PropertyQuotedError preserves the closed tokenizer's error observations.
// It is neither a whole-contract label nor a property acceptance decision.
type PropertyQuotedError uint8

const (
	PropertyQuotedOK PropertyQuotedError = iota
	PropertyQuotedSyntax
	PropertyQuotedBounds
	PropertyQuotedASCII
)

// PropertyQuotedObservation is an owned, comparable observation value. A safely
// captured candidate panic is explicit; the caller's scoped contract determines
// whether that violates a required no-panic property. Source support failures
// return an error rather than an observation that can provide a boolean label.
type PropertyQuotedObservation struct {
	Error    PropertyQuotedError `json:"error"`
	Count    int                 `json:"count"`
	Tokens   [8]string           `json:"tokens"`
	Panicked bool                `json:"panicked"`
}

type propertyObservationSourceError struct{}

func (propertyObservationSourceError) Error() string {
	return "property_observation_source_not_registered"
}

// ObserveQuotedProperty observes only the four original compiled quoted
// controls. It accepts no function, code, registry update or expected label.
// The original literal tables and full-contract control roles are not read.
func ObserveQuotedProperty(sourceID, input string) (PropertyQuotedObservation, error) {
	var candidate func(string) quotedTokens
	switch sourceID {
	case "quoted-correct":
		candidate = quotedCorrect
	case "quoted-literal-comma":
		candidate = quotedLiteralComma
	case "quoted-inside-escape":
		candidate = quotedInsideEscape
	case "quoted-partial-error":
		candidate = quotedPartialError
	default:
		return PropertyQuotedObservation{}, propertyObservationSourceError{}
	}
	got, panicked := observeQuoted(candidate, input)
	return PropertyQuotedObservation{
		Error: PropertyQuotedError(got.Error), Count: got.Count,
		Tokens: got.Tokens, Panicked: panicked,
	}, nil
}

// Bind this adapter's actual compiled source independently of the historical
// eight SourceArtifacts. The coordinator also pins the original source/helper
// closure; this one-file evidence is not a replacement for that evidence.
//
//go:embed property_observation.go
var propertyObservationSourceText string

func PropertyObservationSourceArtifacts() []SourceArtifact {
	return []SourceArtifact{{
		Name: "property_observation.go", SHA256: sum([]byte(propertyObservationSourceText)),
	}}
}
