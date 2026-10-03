// SPDX-License-Identifier: Apache-2.0
package claimfit

import (
	"errors"

	"github.com/teamswyg/laya-tools/internal/roleplan"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
	"github.com/teamswyg/laya-tools/pkg/shortclaimdata"
)

const MaxCohortParents = 60

var ErrCohort = errors.New("claimfit_cohort_binding")

// CohortProjection keeps the full nullable audit separate from existing fully
// known learning columns. Projected parent indexes refer to SelectedAuditIndices,
// not directly to Audit. These are structural views, never permission to Fit.
type CohortProjection struct {
	Audit                [MaxCohortParents]shortclaimdata.CohortExample
	Count                int
	SelectedAuditIndices [MaxCohortParents]int
	SelectedCount        int
	Projection           Projection
}

// ProjectCohort reuses Project, not a second trainer or feature implementation.
// All rows, including withheld ones, are checked before selection or features.
// It checks declared group/family consistency; it cannot discover shared ancestry
// or verify the opaque source/evidence/role-plan records. The caller must freeze
// qualification and whole-component roles separately before invoking any Fit.
//
// Partial-U, ambiguity, calibration and wholly ineligible parents remain only
// in Audit. No U is dropped or converted to a false supervision row. Known labels
// at zero weight remain truthful inside any otherwise selected complete parent.
func ProjectCohort(rows []shortclaimdata.CohortExample) (CohortProjection, error) {
	var zero CohortProjection
	if len(rows) < 1 || len(rows) > MaxCohortParents {
		return zero, ErrCohort
	}
	for i, e := range rows {
		if !e.Valid() {
			return zero, ErrCohort
		}
		b, m := e.Binding(), e.Metadata()
		if i > 0 {
			first := rows[0].Binding()
			if b.RolesGroupsSHA256 != first.RolesGroupsSHA256 || b.MaskPolicySHA256 != first.MaskPolicySHA256 {
				return zero, ErrCohort
			}
		}
		for j := 0; j < i; j++ {
			other := rows[j].Metadata()
			if m.StableID == other.StableID ||
				((m.WholeGroup == other.WholeGroup || m.SourceFamily == other.SourceFamily) && m.Role != other.Role) {
				return zero, ErrCohort
			}
		}
	}
	// Fixed scratch bounds 60 parents and 480 acceptable indexes. Allocate the
	// acceptable backing array only for a selected positive: audit-only calls and
	// all-negative parents need none. Project owns its returned columns.
	var parents [MaxCohortParents]Parent
	var acceptable *[MaxCohortParents][shortclaim.MaxCandidates]int
	out := CohortProjection{Count: len(rows)}
	copy(out.Audit[:], rows)
	for auditIndex, e := range rows {
		s, m := e.Supervision(), e.Metadata()
		if s.Ambiguous || m.Role == "development_calibration" {
			continue
		}
		known, anyEligible := true, false
		var eligible [shortclaim.MaxCandidates]bool
		for c := 0; c < s.Count; c++ {
			known = known && s.LabelKnown[c]
			if m.Role == "development_train" {
				eligible[c] = s.LossWeights[c] == 1
			} else {
				eligible[c] = s.EvaluationEligible[c]
			}
			anyEligible = anyEligible || eligible[c]
		}
		if !known || !anyEligible {
			continue
		}
		n := out.SelectedCount
		p := Parent{Text: e.Input().Prepared(), Truth: NoAnswer, WholeGroup: m.WholeGroup, LossEligible: eligible}
		if m.Role == "development_train" {
			p.Role = roleplan.DevelopmentTrain
		} else {
			p.Role = roleplan.DevelopmentValidation
		}
		positive := 0
		for c := 0; c < s.Count; c++ {
			if s.Labels[c] {
				if acceptable == nil {
					acceptable = new([MaxCohortParents][shortclaim.MaxCandidates]int)
				}
				acceptable[n][positive] = c
				positive++
			}
		}
		if positive != 0 {
			p.Truth = Known
			p.Acceptable = acceptable[n][:positive]
		}
		parents[n] = p
		out.SelectedAuditIndices[n] = auditIndex
		out.SelectedCount++
	}
	if out.SelectedCount == 0 {
		return out, nil
	}
	projection, err := Project(parents[:out.SelectedCount])
	if err != nil {
		return zero, err
	}
	out.Projection = projection
	return out, nil
}
