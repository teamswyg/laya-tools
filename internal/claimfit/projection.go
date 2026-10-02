// SPDX-License-Identifier: Apache-2.0
// Package claimfit projects caller-frozen supervision into owned sparse columns.
// It is neither a loader, an eligibility policy, a role allocator nor a trainer.
package claimfit

import (
	"math"
	"sort"
	"unsafe"

	"github.com/teamswyg/laya-tools/internal/hintlearn"
	"github.com/teamswyg/laya-tools/internal/pairlearn"
	"github.com/teamswyg/laya-tools/internal/roleplan"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

const MaxPreparedPayloadBytes uint64 = 64 << 20

// PrototypeSourceSHA256 pins the unchanged private preparation64 source archive.
// It is historical provenance, not a feature or a readiness certificate.
const PrototypeSourceSHA256 = "572e33a815b9a123c0e521d92567d6f254526250ab4d82e202487ddbe1d517e5"

// RoleEnumContract is a proposed bridge contract, not a membership/source pin.
// Callers must separately bind frozen roleplan source and full membership.
const RoleEnumContract = "riido-claimfit-role-bridge-v1:train=0,validation=1,calibration=2"

type TruthState uint8

const (
	Known TruthState = iota + 1
	NoAnswer
	Unknown
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrTruth      Error = "claimfit_invalid_truth"
	ErrAcceptable Error = "claimfit_invalid_acceptable"
	ErrRole       Error = "claimfit_invalid_role"
	ErrGroup      Error = "claimfit_invalid_group"
	ErrLeakage    Error = "claimfit_group_role_leakage"
	ErrFeatures   Error = "claimfit_invalid_features"
	ErrPayload    Error = "claimfit_prepared_payload_bounds"
)

// Parent consumes already-bound original truth and masks. Acceptable preserves
// caller order; it must be nonempty for Known and empty otherwise. Masks do not
// change truth. Unknown/calibration masks are retained without producing rows.
type Parent struct {
	Text         shortclaim.Prepared
	Truth        TruthState
	Acceptable   []int
	Role         roleplan.Role
	WholeGroup   int
	LossEligible [shortclaim.MaxCandidates]bool
}

// NullableLabel uses Known=false for null. Positive=false alone is not a label.
type NullableLabel struct{ Known, Positive bool }

// RuntimeParent owns bounded value arrays. Strings share immutable Go bytes as
// shortclaim.Prepared does. Caller slices cannot mutate this snapshot.
type RuntimeParent struct {
	Text            shortclaim.Prepared
	Truth           TruthState
	Acceptable      [shortclaim.MaxCandidates]int
	AcceptableCount int
	Role            roleplan.Role
	WholeGroup      int
	LossEligible    [shortclaim.MaxCandidates]bool
	Labels          [shortclaim.MaxCandidates]NullableLabel
}

type RowRef struct{ ParentIndex, CandidateIndex int }

type Denominators struct {
	OriginalFitRows, PositiveWeightRows, ZeroWeightRows int
	UnknownAuditCandidates                              int
	PositiveLabels, NegativeLabels                      int
}

// DiagnosticView contains only positive-weight rows. Calling the unchanged
// pairlearn.AUC on Data is an unweighted diagnostic over this explicit view.
type DiagnosticView struct {
	Data         pairlearn.Dataset
	Refs         []RowRef
	Denominators Denominators
}

type RoleCounts struct {
	Parents, Candidates, UnknownParents, UnknownCandidates int
	FitRows, ZeroWeightFitRows                             int
	WholeGroups, PositiveWeightGroups                      int
}

type Projection struct {
	Parents         []RuntimeParent
	Development     pairlearn.Dataset
	Validation      pairlearn.Dataset
	DevelopmentRefs []RowRef
	ValidationRefs  []RowRef
	DevelopmentAUC  DiagnosticView
	ValidationAUC   DiagnosticView
	Counts          [3]RoleCounts
	// Payload includes owned arrays/headers and conservatively counts retained
	// immutable string bytes per parent. It excludes allocator and scratch costs.
	PayloadBytes uint64
	FeatureScans int
}

type rowPlan struct {
	parent, candidate, features int
	weight                      bool
}
type groupRole struct {
	group    int
	role     roleplan.Role
	positive bool
}
type splitPlan struct{ rows, features, positiveRows, positiveFeatures, unknown int }

// Project never opens files, checks SHA/readiness, allocates roles or fits a
// model. Only original request/candidate Text enters hintlearn.Features.
// A successful all-zero-weight projection is not a trainable dataset claim.
func Project(parents []Parent) (Projection, error) {
	return project(parents, hintlearn.Features, MaxPreparedPayloadBytes)
}

func roleIndex(r roleplan.Role) (int, bool) {
	// Keep the enum bridge explicit; do not call roleplan planning APIs.
	switch r {
	case roleplan.DevelopmentTrain:
		return 0, uint8(r) == 0
	case roleplan.DevelopmentValidation:
		return 1, uint8(r) == 1
	case roleplan.DevelopmentCalibration:
		return 2, uint8(r) == 2
	}
	return 0, false
}

func addPayload(total *uint64, count, width, limit uint64) error {
	if *total > limit || width != 0 && count > (limit-*total)/width {
		return ErrPayload
	}
	*total += count * width
	return nil
}

func textBytes(p shortclaim.Prepared) uint64 {
	n := uint64(len(p.Schema) + len(p.Request) + len(p.NormalizedRequest) + len(p.Provenance))
	for _, c := range p.Candidates {
		n += uint64(len(c.ID) + len(c.Text) + len(c.NormalizedText))
	}
	return n
}

func validateParent(p Parent) error {
	if err := shortclaim.ValidatePrepared(p.Text); err != nil {
		return err
	}
	if p.Truth != Known && p.Truth != NoAnswer && p.Truth != Unknown {
		return ErrTruth
	}
	if _, ok := roleIndex(p.Role); !ok {
		return ErrRole
	}
	if p.WholeGroup < 0 {
		return ErrGroup
	}
	if len(p.Acceptable) > p.Text.Count || (p.Truth == Known) != (len(p.Acceptable) > 0) {
		return ErrAcceptable
	}
	var seen [shortclaim.MaxCandidates]bool
	for _, i := range p.Acceptable {
		if i < 0 || i >= p.Text.Count || seen[i] {
			return ErrAcceptable
		}
		seen[i] = true
	}
	return nil
}

func validFeatures(fs []hintlearn.Feature) bool {
	if len(fs) > hintlearn.Dimension {
		return false
	}
	var seen [hintlearn.Dimension]bool
	for _, f := range fs {
		if f.Index < 0 || f.Index >= hintlearn.Dimension || seen[f.Index] || math.IsNaN(f.Value) || math.IsInf(f.Value, 0) {
			return false
		}
		seen[f.Index] = true
	}
	return true
}

func dataset(split string, rows, features int) pairlearn.Dataset {
	return pairlearn.Dataset{Split: split, Offsets: make([]int, rows+1), Indices: make([]uint16, features), Values: make([]float64, features), Labels: make([]float64, rows), Groups: make([]int, rows), SampleWeights: make([]float64, rows)}
}

func datasetPayload(total *uint64, rows, features int, limit uint64) error {
	intBytes := uint64(unsafe.Sizeof(int(0)))
	// Offsets, uint16 indices, float64 values, labels, groups, explicit weights,
	// RowRef backing arrays are included; owning structs/slice headers are
	// already counted once in sizeof(Projection).
	for _, x := range [][2]uint64{{uint64(rows) + 1, intBytes}, {uint64(features), 10}, {uint64(rows), 16 + intBytes + uint64(unsafe.Sizeof(RowRef{}))}} {
		if err := addPayload(total, x[0], x[1], limit); err != nil {
			return err
		}
	}
	return nil
}

func project(parents []Parent, extract func(string, string) []hintlearn.Feature, limit uint64) (Projection, error) {
	var zero Projection
	var payload uint64
	if err := addPayload(&payload, 1, uint64(unsafe.Sizeof(Projection{})), limit); err != nil {
		return zero, err
	}
	if err := addPayload(&payload, uint64(len(parents)), uint64(unsafe.Sizeof(RuntimeParent{})), limit); err != nil {
		return zero, err
	}
	groups := make([]groupRole, len(parents))
	var counts [3]RoleCounts
	var plans [2]splitPlan
	rows := 0
	for i, p := range parents {
		if err := validateParent(p); err != nil {
			return zero, err
		}
		if err := addPayload(&payload, textBytes(p.Text), 1, limit); err != nil {
			return zero, err
		}
		r, _ := roleIndex(p.Role)
		counts[r].Parents++
		counts[r].Candidates += p.Text.Count
		positive := false
		if p.Truth == Unknown {
			counts[r].UnknownParents++
			counts[r].UnknownCandidates += p.Text.Count
			if r < 2 {
				plans[r].unknown += p.Text.Count
			}
		} else if r < 2 {
			plans[r].rows += p.Text.Count
			rows += p.Text.Count
			for c := 0; c < p.Text.Count; c++ {
				positive = positive || p.LossEligible[c]
			}
		}
		groups[i] = groupRole{p.WholeGroup, p.Role, positive}
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].group < groups[j].group })
	for start := 0; start < len(groups); {
		end := start + 1
		positive := groups[start].positive
		for end < len(groups) && groups[end].group == groups[start].group {
			if groups[end].role != groups[start].role {
				return zero, ErrLeakage
			}
			positive = positive || groups[end].positive
			end++
		}
		r, _ := roleIndex(groups[start].role)
		counts[r].WholeGroups++
		if positive {
			counts[r].PositiveWeightGroups++
		}
		start = end
	}
	// Preflight all non-feature output columns before allocating row plans.
	base := payload
	for r := range plans {
		if err := datasetPayload(&base, plans[r].rows, 0, limit); err != nil {
			return zero, err
		}
		if err := datasetPayload(&base, 0, 0, limit); err != nil {
			return zero, err
		}
	}
	// Scratch is bounded separately; it does not falsely consume output cap.
	var scratch uint64
	if err := addPayload(&scratch, uint64(rows), uint64(unsafe.Sizeof(rowPlan{})), limit); err != nil {
		return zero, err
	}
	rowPlans := make([]rowPlan, 0, rows)
	scans := 0
	// Exact preflight pass retains only small row metadata, never all features.
	for pi, p := range parents {
		r, _ := roleIndex(p.Role)
		if r == 2 || p.Truth == Unknown {
			continue
		}
		for c := 0; c < p.Text.Count; c++ {
			fs := extract(p.Text.Request, p.Text.Candidates[c].Text)
			scans++
			if !validFeatures(fs) {
				return zero, ErrFeatures
			}
			plans[r].features += len(fs)
			if p.LossEligible[c] {
				plans[r].positiveRows++
				plans[r].positiveFeatures += len(fs)
			}
			rowPlans = append(rowPlans, rowPlan{pi, c, len(fs), p.LossEligible[c]})
			// Stop promptly before an oversized complete output is allocated.
			trial := payload
			for k := range plans {
				if err := datasetPayload(&trial, plans[k].rows, plans[k].features, limit); err != nil {
					return zero, err
				}
				if err := datasetPayload(&trial, plans[k].positiveRows, plans[k].positiveFeatures, limit); err != nil {
					return zero, err
				}
			}
		}
	}
	for r := range plans {
		if err := datasetPayload(&payload, plans[r].rows, plans[r].features, limit); err != nil {
			return zero, err
		}
		if err := datasetPayload(&payload, plans[r].positiveRows, plans[r].positiveFeatures, limit); err != nil {
			return zero, err
		}
	}
	out := Projection{Parents: make([]RuntimeParent, len(parents)), PayloadBytes: payload, FeatureScans: scans}
	var d, auc [2]pairlearn.Dataset
	var refs, aucRefs [2][]RowRef
	var denom [2]Denominators
	for r, split := range [2]string{"development", "validation"} {
		d[r] = dataset(split, plans[r].rows, plans[r].features)
		d[r].Excluded = plans[r].unknown
		auc[r] = dataset(split, plans[r].positiveRows, plans[r].positiveFeatures)
		refs[r] = make([]RowRef, plans[r].rows)
		aucRefs[r] = make([]RowRef, plans[r].positiveRows)
		denom[r] = Denominators{OriginalFitRows: plans[r].rows, PositiveWeightRows: plans[r].positiveRows, ZeroWeightRows: plans[r].rows - plans[r].positiveRows, UnknownAuditCandidates: plans[r].unknown}
		counts[r].FitRows = plans[r].rows
		counts[r].ZeroWeightFitRows = plans[r].rows - plans[r].positiveRows
	}
	for i, p := range parents {
		v := RuntimeParent{Text: p.Text, Truth: p.Truth, AcceptableCount: len(p.Acceptable), Role: p.Role, WholeGroup: p.WholeGroup, LossEligible: p.LossEligible}
		copy(v.Acceptable[:], p.Acceptable)
		if p.Truth != Unknown {
			for c := 0; c < p.Text.Count; c++ {
				v.Labels[c].Known = true
			}
			for _, c := range p.Acceptable {
				v.Labels[c].Positive = true
			}
		}
		out.Parents[i] = v
	}
	var rowAt, featureAt, aucRowAt, aucFeatureAt [2]int
	for _, rp := range rowPlans {
		p := parents[rp.parent]
		r, _ := roleIndex(p.Role)
		fs := extract(p.Text.Request, p.Text.Candidates[rp.candidate].Text)
		out.FeatureScans++
		if len(fs) != rp.features || !validFeatures(fs) {
			return zero, ErrFeatures
		}
		y := 0.
		if out.Parents[rp.parent].Labels[rp.candidate].Positive {
			y = 1
		}
		ref := RowRef{rp.parent, rp.candidate}
		put := func(dst *pairlearn.Dataset, row, at int, weight float64) int {
			for _, f := range fs {
				dst.Indices[at] = uint16(f.Index)
				dst.Values[at] = f.Value
				at++
			}
			dst.Offsets[row+1] = at
			dst.Labels[row] = y
			dst.Groups[row] = p.WholeGroup
			dst.SampleWeights[row] = weight
			return at
		}
		weight := 0.
		if rp.weight {
			weight = 1
		}
		featureAt[r] = put(&d[r], rowAt[r], featureAt[r], weight)
		refs[r][rowAt[r]] = ref
		rowAt[r]++
		if rp.weight {
			aucFeatureAt[r] = put(&auc[r], aucRowAt[r], aucFeatureAt[r], 1)
			aucRefs[r][aucRowAt[r]] = ref
			aucRowAt[r]++
			if y == 1 {
				denom[r].PositiveLabels++
			} else {
				denom[r].NegativeLabels++
			}
		}
	}
	out.Development, out.Validation = d[0], d[1]
	out.DevelopmentRefs, out.ValidationRefs = refs[0], refs[1]
	out.DevelopmentAUC = DiagnosticView{auc[0], aucRefs[0], denom[0]}
	out.ValidationAUC = DiagnosticView{auc[1], aucRefs[1], denom[1]}
	out.Counts = counts
	return out, nil
}
