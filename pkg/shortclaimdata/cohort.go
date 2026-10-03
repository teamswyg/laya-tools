// SPDX-License-Identifier: Apache-2.0
package shortclaimdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"unicode/utf8"

	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const (
	CohortSchema          = "riido-shortclaim-cohort-row-v1"
	CohortInputProvenance = "cohort-audit-v1"
)

const (
	ErrCohortPin   Error = "shortclaimdata_cohort_pin"
	ErrCohortState Error = "shortclaimdata_cohort_state"
	ErrCohortMask  Error = "shortclaimdata_cohort_mask"
)

// CohortState is declared candidate truth, independently of loss eligibility.
// U is not a negative label. Evidence predicates are not evaluated by this API.
type CohortState uint8

const (
	CohortTrue CohortState = iota + 1
	CohortFalse
	CohortUnknown
)

type CohortMetadata struct {
	StableID, SourceFamily, SourceRevision, TextRevision, Role string
	WholeGroup                                                 int
}

// CohortBinding is supplied separately, before LoadCohortRow. The row digest
// includes exact bytes, masks and interpretation. The other digests bind opaque
// records; matching them does not verify those records, rights or qualification.
// Metadata binds the caller's frozen role/group, never allocates a new role.
type CohortBinding struct {
	RowSHA256, SourceSHA256, EvidenceSHA256, MaskPolicySHA256, RolesGroupsSHA256 [32]byte
	Metadata                                                                     CohortMetadata
}

// CohortSupervision owns SoA arrays. LabelKnown=false represents JSON null;
// Labels=false alone cannot distinguish U from F. Ambiguity preserves T/F truth
// but withholds all eligibility. Unused slots remain zero.
type CohortSupervision struct {
	States             [shortclaim.MaxCandidates]CohortState
	LabelKnown, Labels [shortclaim.MaxCandidates]bool
	LossWeights        [shortclaim.MaxCandidates]uint8
	EvaluationEligible [shortclaim.MaxCandidates]bool
	Count              int
	Ambiguous          bool
}

type CohortExample struct {
	input       shortclaim.ValidatedInput
	supervision CohortSupervision
	binding     CohortBinding
	ready       bool
}

func (e CohortExample) Valid() bool                      { return e.ready }
func (e CohortExample) Input() shortclaim.ValidatedInput { return e.input }
func (e CohortExample) Supervision() CohortSupervision   { return e.supervision }
func (e CohortExample) Binding() CohortBinding           { return e.binding }
func (e CohortExample) Metadata() CohortMetadata         { return e.binding.Metadata }

// Class is structural bookkeeping from declared labels, not semantic proof.
// Ambiguity takes precedence; any candidate U prevents known_none/known_all.
func (e CohortExample) Class() string {
	if !e.ready {
		return "invalid"
	}
	s := e.supervision
	if s.Ambiguous {
		return "ambiguous"
	}
	positive := 0
	for i := 0; i < s.Count; i++ {
		if !s.LabelKnown[i] {
			return "unknown_containing"
		}
		if s.Labels[i] {
			positive++
		}
	}
	switch positive {
	case 0:
		return "known_none"
	case 1:
		return "known_one"
	case s.Count:
		return "known_all"
	default:
		return "known_many"
	}
}

func validCohortMetadata(m CohortMetadata) bool {
	return identifier(m.StableID, shortclaim.MaxIDBytes) &&
		identifier(m.SourceFamily, MaxMetadataBytes) && revision(m.SourceRevision) &&
		identifier(m.TextRevision, MaxMetadataBytes) && m.WholeGroup >= 1 && m.WholeGroup <= MaxGroup &&
		(m.Role == "development_train" || m.Role == "development_validation" || m.Role == "development_calibration")
}

func validCohortBinding(b CohortBinding) bool {
	zero := [32]byte{}
	return validCohortMetadata(b.Metadata) && b.RowSHA256 != zero && b.SourceSHA256 != zero &&
		b.EvidenceSHA256 != zero && b.MaskPolicySHA256 != zero && b.RolesGroupsSHA256 != zero
}

type cohortWire struct {
	schema, request, policy, interpretation string
	metadata                                CohortMetadata
	candidates                              [shortclaim.MaxCandidates]shortclaim.Candidate
	supervision                             CohortSupervision
	source, evidence, masks, roles          [32]byte
}

// LoadCohortRow is opt-in and separate from LoadDevelopmentRow. It reads exactly
// one JSON object <=16 KiB and 5 or 8 candidates. Pins are checked before reading
// any row; exact row bytes are checked before decoding or normalization. Only
// request/candidate text enters normalization. No feature, Fit or path is used.
func LoadCohortRow(r io.Reader, expected CohortBinding) (CohortExample, error) {
	if !validCohortBinding(expected) {
		return CohortExample{}, ErrCohortPin
	}
	if r == nil {
		return CohortExample{}, ErrRead
	}
	raw, err := io.ReadAll(io.LimitReader(r, MaxRowBytes+1))
	if err != nil {
		return CohortExample{}, ErrRead
	}
	if len(raw) > MaxRowBytes {
		return CohortExample{}, ErrBounds
	}
	if sha256.Sum256(raw) != expected.RowSHA256 {
		return CohortExample{}, ErrCohortPin
	}
	if !utf8.Valid(raw) {
		return CohortExample{}, ErrUnicode
	}
	if !json.Valid(raw) {
		return CohortExample{}, ErrJSON
	}
	if !scalarEscapes(raw) {
		return CohortExample{}, ErrUnicode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var w cohortWire
	err = object(d, [12]string{"schema", "stable_id", "request", "candidates", "role", "whole_group", "source_family", "source_revision", "text_revision", "feature_policy", "interpretation", "bindings"}, 12, func(key string) error {
		switch key {
		case "schema":
			return text(d, &w.schema)
		case "stable_id":
			return text(d, &w.metadata.StableID)
		case "request":
			return text(d, &w.request)
		case "candidates":
			return cohortCandidates(d, &w)
		case "role":
			return text(d, &w.metadata.Role)
		case "whole_group":
			return positiveInt(d, MaxGroup, &w.metadata.WholeGroup, ErrMetadata)
		case "source_family":
			return text(d, &w.metadata.SourceFamily)
		case "source_revision":
			return text(d, &w.metadata.SourceRevision)
		case "text_revision":
			return text(d, &w.metadata.TextRevision)
		case "feature_policy":
			return text(d, &w.policy)
		case "interpretation":
			return text(d, &w.interpretation)
		case "bindings":
			return object(d, [12]string{"source_sha256", "evidence_sha256", "mask_policy_sha256", "roles_groups_sha256"}, 4, func(key string) error {
				switch key {
				case "source_sha256":
					return cohortDigest(d, &w.source)
				case "evidence_sha256":
					return cohortDigest(d, &w.evidence)
				case "mask_policy_sha256":
					return cohortDigest(d, &w.masks)
				case "roles_groups_sha256":
					return cohortDigest(d, &w.roles)
				}
				return ErrField
			})
		}
		return ErrField
	})
	if err != nil {
		return CohortExample{}, err
	}
	if _, err = d.Token(); err != io.EOF {
		return CohortExample{}, ErrJSON
	}
	if w.schema != CohortSchema {
		return CohortExample{}, ErrSchema
	}
	if !validCohortMetadata(w.metadata) {
		return CohortExample{}, ErrMetadata
	}
	if w.metadata != expected.Metadata || w.source != expected.SourceSHA256 ||
		w.evidence != expected.EvidenceSHA256 || w.masks != expected.MaskPolicySHA256 || w.roles != expected.RolesGroupsSHA256 {
		return CohortExample{}, ErrCohortPin
	}
	if w.policy != FeaturePolicy {
		return CohortExample{}, ErrMetadata
	}
	if w.interpretation != "unambiguous" && w.interpretation != "ambiguous" {
		return CohortExample{}, ErrCohortState
	}
	w.supervision.Ambiguous = w.interpretation == "ambiguous"
	for c := 0; c < w.supervision.Count; c++ {
		s := &w.supervision
		if (s.Ambiguous || s.States[c] == CohortUnknown) && (s.LossWeights[c] != 0 || s.EvaluationEligible[c]) {
			return CohortExample{}, ErrCohortMask
		}
		switch w.metadata.Role {
		case "development_train":
			if s.EvaluationEligible[c] {
				return CohortExample{}, ErrCohortMask
			}
		case "development_validation":
			if s.LossWeights[c] != 0 {
				return CohortExample{}, ErrCohortMask
			}
		case "development_calibration":
			if s.LossWeights[c] != 0 || s.EvaluationEligible[c] {
				return CohortExample{}, ErrCohortMask
			}
		}
	}
	in, err := shortclaim.ValidateInput(shortclaim.Input{
		Schema: shortclaim.Schema, Request: w.request, Candidates: w.candidates[:w.supervision.Count], Provenance: CohortInputProvenance,
	})
	if err != nil {
		return CohortExample{}, ErrInput
	}
	return CohortExample{input: in, supervision: w.supervision, binding: expected, ready: true}, nil
}

func cohortDigest(d *json.Decoder, dst *[32]byte) error {
	var s string
	if err := text(d, &s); err != nil {
		return err
	}
	if len(s) != 64 {
		return ErrCohortPin
	}
	for i := range s {
		if !(s[i] >= '0' && s[i] <= '9' || s[i] >= 'a' && s[i] <= 'f') {
			return ErrCohortPin
		}
	}
	_, err := hex.Decode(dst[:], []byte(s))
	if err != nil || *dst == [32]byte{} {
		return ErrCohortPin
	}
	return nil
}

func cohortCandidates(d *json.Decoder, w *cohortWire) error {
	t, err := d.Token()
	if err != nil || t != json.Delim('[') {
		return ErrJSON
	}
	n := 0
	for d.More() {
		if n == shortclaim.MaxCandidates {
			return ErrCandidates
		}
		var state string
		var labelKnown, label bool
		err = object(d, [12]string{"metadata_id", "text", "state", "label", "loss_weight", "evaluation_eligible"}, 6, func(key string) error {
			switch key {
			case "metadata_id":
				return text(d, &w.candidates[n].ID)
			case "text":
				return text(d, &w.candidates[n].Text)
			case "state":
				return text(d, &state)
			case "label":
				t, e := d.Token()
				if e != nil {
					return ErrLabel
				}
				if t == nil {
					return nil
				}
				var ok bool
				label, ok = t.(bool)
				if !ok {
					return ErrLabel
				}
				labelKnown = true
			case "loss_weight":
				t, e := d.Token()
				number, ok := t.(json.Number)
				if e != nil || !ok || (number.String() != "0" && number.String() != "1") {
					return ErrWeight
				}
				if number.String() == "1" {
					w.supervision.LossWeights[n] = 1
				}
			case "evaluation_eligible":
				return boolean(d, &w.supervision.EvaluationEligible[n], ErrCohortMask)
			}
			return nil
		})
		if err != nil {
			return err
		}
		switch state {
		case "T":
			if !labelKnown || !label {
				return ErrLabel
			}
			w.supervision.States[n] = CohortTrue
		case "F":
			if !labelKnown || label {
				return ErrLabel
			}
			w.supervision.States[n] = CohortFalse
		case "U":
			if labelKnown {
				return ErrLabel
			}
			w.supervision.States[n] = CohortUnknown
		default:
			return ErrCohortState
		}
		w.supervision.LabelKnown[n], w.supervision.Labels[n] = labelKnown, label
		n++
	}
	if t, err = d.Token(); err != nil || t != json.Delim(']') {
		return ErrJSON
	}
	if n != 5 && n != 8 {
		return ErrCandidates
	}
	w.supervision.Count = n
	return nil
}
