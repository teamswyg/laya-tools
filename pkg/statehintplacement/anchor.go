// Package statehintplacement checks conditional shadow-proposal placement.
// It provides no native transport, credential, mutation or admission-proof API.
package statehintplacement

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/statehint"
	"github.com/teamswyg/laya-tools/pkg/statehintcatalog"
)

type Scope = statehintcatalog.Scope

// Owner revisions are asserted by the application's trusted owner boundary.
// TextSHA256 is derived from bytes, not issued by the owner. An unavailable
// OwnerContentRevision stays empty; neither a digest nor a timestamp replaces it.
type Anchor struct {
	Scope                Scope  `json:"scope"`
	WorkRef              string `json:"work_ref"`
	OwnerWorkRevision    string `json:"owner_work_revision"`
	ContentRef           string `json:"content_ref"`
	TextSHA256           string `json:"text_sha256"`
	OwnerContentRevision string `json:"owner_content_revision,omitempty"`
}

// ReadAnchor receives no expected revision, user hash, or model-input text.
// Current TextSHA256 must come from current content already held by the trusted
// application boundary. If that boundary cannot access it, return no digest.
type RevisionReader interface {
	ReadAnchor(context.Context, Scope, string, string) (Anchor, error)
}
type RevisionReaderFunc func(context.Context, Scope, string, string) (Anchor, error)

func (f RevisionReaderFunc) ReadAnchor(ctx context.Context, scope Scope, workRef, contentRef string) (Anchor, error) {
	if f == nil {
		return Anchor{}, ErrUnavailable
	}
	return f(ctx, scope, workRef, contentRef)
}

type Status string

const (
	Matched     Status = "placement_matched"
	Stale       Status = "stale"
	Unavailable Status = "unavailable"
)

var (
	ErrInput       = errors.New("statehintplacement: invalid held text or request")
	ErrUnavailable = errors.New("statehintplacement: reader or predictor unavailable")
)

type Validation struct {
	Status                        Status `json:"status"`
	Reason                        string `json:"reason"`
	OwnerContentRevisionAvailable bool   `json:"owner_content_revision_available"`
	ActualVerified                bool   `json:"actual_verified"`
	MutationExecuted              bool   `json:"mutation_executed"`
}

// TextDigest hashes exact UTF-8 bytes, without trimming, case folding or other
// normalization. It makes no claim about owner revisions or authorization.
func TextDigest(text string) (string, error) {
	if len(text) > statehint.MaxTextBytes || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
		return "", ErrInput
	}
	digest := sha256.Sum256([]byte(text))
	return hex.EncodeToString(digest[:]), nil
}

func bounded(value string) bool {
	if value == "" || len(value) > 256 || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func validAnchor(a Anchor) bool {
	return bounded(a.Scope.WorkspaceRef) && bounded(a.Scope.OwnerRef) && bounded(a.WorkRef) && bounded(a.OwnerWorkRevision) && bounded(a.ContentRef) && validDigest(a.TextSHA256) &&
		(a.OwnerContentRevision == "" || bounded(a.OwnerContentRevision))
}

// ValidatePlacement compares complete owner revisions as opaque strings. A
// matched observation is conditional: it is not an authorization, owner-issued
// content revision, atomic snapshot, or verified deployment/admission proof.
func ValidatePlacement(expected, current Anchor) Validation {
	result := Validation{Status: Unavailable, Reason: "expected_metadata_unavailable"}
	if !validAnchor(expected) {
		return result
	}
	result.Reason = "current_metadata_unavailable"
	if !validAnchor(current) {
		return result
	}
	result.Status = Stale
	switch {
	case expected.Scope != current.Scope:
		result.Reason = "scope_changed"
	case expected.WorkRef != current.WorkRef:
		result.Reason = "work_ref_changed"
	case expected.ContentRef != current.ContentRef:
		result.Reason = "content_ref_changed"
	case expected.OwnerWorkRevision != current.OwnerWorkRevision:
		result.Reason = "owner_work_revision_changed"
	case expected.OwnerContentRevision != current.OwnerContentRevision:
		result.Reason = "owner_content_revision_changed"
	case expected.TextSHA256 != current.TextSHA256:
		result.Reason = "text_changed"
	default:
		result.Status = Matched
		result.Reason = "exact_anchor_observation_match"
	}
	result.OwnerContentRevisionAvailable = current.OwnerContentRevision != ""
	return result
}
