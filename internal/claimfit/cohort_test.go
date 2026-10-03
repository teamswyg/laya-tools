// SPDX-License-Identifier: Apache-2.0
package claimfit

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/roleplan"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
	cohort "github.com/teamswyg/laya-tools/pkg/shortclaimdata"
)

// Synthetic JSON with external pins for the public reader. Evidence hashes bind
// opaque records by equality; they do not prove truth or Fit readiness.
const (
	bridgeTrain = "development_train"
	bridgeVal   = "development_validation"
	bridgeCal   = "development_calibration"
)

type bridgeSpec struct {
	id, family, role, states, loss, evaluation string
	group                                      int
	ambiguous                                  bool
	mask, roles                                [32]byte
}

func bridgeCase(id string, group int, role, states, loss, evaluation string) bridgeSpec {
	return bridgeSpec{id: id, family: "family-" + id, role: role, group: group, states: states, loss: loss, evaluation: evaluation}
}

func bridgeInput(n int) shortclaim.Input {
	texts := [8]string{
		"KeepEqual toy entries in their original order.",
		"Reverse the entries.",
		"Preserve original order.",
		"Drop repeated entries.",
		"Sort names.",
		"Leave entries unchanged.",
		"Append an entry.",
		"Rotate entries.",
	}
	cs := make([]shortclaim.Candidate, n)
	for i := range cs {
		cs[i] = shortclaim.Candidate{ID: "bridge-" + strconv.Itoa(i), Text: texts[i]}
	}
	return shortclaim.Input{Schema: shortclaim.Schema, Request: "KeepEqual toy entries in originalOrder.", Candidates: cs, Provenance: cohort.CohortInputProvenance}
}

func bridgeRow(t testing.TB, s bridgeSpec) cohort.CohortExample {
	t.Helper()
	n := len(s.states)
	if (n != 5 && n != 8) || len(s.loss) != n || len(s.evaluation) != n {
		t.Fatal("invalid synthetic control specification")
	}
	// Supply pins before marshaling; pin exact row bytes separately.
	b := cohort.CohortBinding{
		SourceSHA256:      sha256.Sum256([]byte("synthetic-cohort-control-source-v1")),
		EvidenceSHA256:    sha256.Sum256([]byte("synthetic-cohort-control-evidence-v1")),
		MaskPolicySHA256:  sha256.Sum256([]byte("synthetic-cohort-control-masks-v1")),
		RolesGroupsSHA256: sha256.Sum256([]byte("synthetic-cohort-control-roles-v1")),
		Metadata:          cohort.CohortMetadata{StableID: s.id, SourceFamily: s.family, SourceRevision: strings.Repeat("a", 40), TextRevision: "synthetic-text-v1", Role: s.role, WholeGroup: s.group},
	}
	if s.mask != ([32]byte{}) {
		b.MaskPolicySHA256 = s.mask
	}
	if s.roles != ([32]byte{}) {
		b.RolesGroupsSHA256 = s.roles
	}
	in := bridgeInput(n)
	cs := make([]map[string]any, n)
	for i, c := range in.Candidates {
		var label any
		switch s.states[i] {
		case 'T':
			label = true
		case 'F':
			label = false
		case 'U':
		default:
			t.Fatal("invalid synthetic state")
		}
		if (s.loss[i] != '0' && s.loss[i] != '1') || (s.evaluation[i] != '0' && s.evaluation[i] != '1') {
			t.Fatal("invalid synthetic mask")
		}
		cs[i] = map[string]any{"metadata_id": c.ID, "text": c.Text, "state": string(s.states[i]), "label": label, "loss_weight": int(s.loss[i] - '0'), "evaluation_eligible": s.evaluation[i] == '1'}
	}
	interpretation := "unambiguous"
	if s.ambiguous {
		interpretation = "ambiguous"
	}
	wire := map[string]any{
		"schema": cohort.CohortSchema, "stable_id": s.id, "request": in.Request, "candidates": cs,
		"role": s.role, "whole_group": s.group, "source_family": s.family,
		"source_revision": b.Metadata.SourceRevision, "text_revision": b.Metadata.TextRevision,
		"feature_policy": cohort.FeaturePolicy, "interpretation": interpretation,
		"bindings": map[string]string{
			"source_sha256": hex.EncodeToString(b.SourceSHA256[:]), "evidence_sha256": hex.EncodeToString(b.EvidenceSHA256[:]),
			"mask_policy_sha256": hex.EncodeToString(b.MaskPolicySHA256[:]), "roles_groups_sha256": hex.EncodeToString(b.RolesGroupsSHA256[:]),
		},
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	b.RowSHA256 = sha256.Sum256(raw)
	e, err := cohort.LoadCohortRow(bytes.NewReader(raw), b)
	if err != nil || !e.Valid() || e.Binding() != b {
		t.Fatalf("synthetic public reader binding failed: %v", err)
	}
	return e
}

func bridgeParent(t testing.TB, n int, truth TruthState, role roleplan.Role, group int, acceptable, eligible []int) Parent {
	t.Helper()
	text, err := shortclaim.Validate(bridgeInput(n))
	if err != nil {
		t.Fatal(err)
	}
	p := Parent{Text: text, Truth: truth, Role: role, WholeGroup: group, Acceptable: acceptable}
	for _, c := range eligible {
		p.LossEligible[c] = true
	}
	return p
}

func bridgeReject(t *testing.T, rows []cohort.CohortExample) {
	t.Helper()
	out, err := ProjectCohort(rows)
	if err != ErrCohort || !reflect.DeepEqual(out, CohortProjection{}) {
		t.Fatalf("rejection output is not zero: err=%v count=%d selected=%d scans=%d", err, out.Count, out.SelectedCount, out.Projection.FeatureScans)
	}
}

func TestCohortSelectedTruthMasksMappingAndProjectColumns(t *testing.T) {
	partial := bridgeCase("partial", 1, bridgeTrain, "TFUFF", "01000", "00000")
	train := bridgeCase("train", 2, bridgeTrain, "TFTFF", "01100", "00000")
	ambiguous := bridgeCase("ambiguous", 3, bridgeVal, "TFFTFFFF", "00000000", "00000000")
	ambiguous.ambiguous = true
	validation := bridgeCase("validation", 4, bridgeVal, "FTFFFFTF", "00000000", "10000010")
	specs := []bridgeSpec{partial, train, ambiguous, validation}
	rows := make([]cohort.CohortExample, len(specs))
	for i, s := range specs {
		rows[i] = bridgeRow(t, s)
	}
	out, err := ProjectCohort(rows)
	if err != nil {
		t.Fatal(err)
	}
	if out.Count != 4 || out.SelectedCount != 2 || out.SelectedAuditIndices[0] != 1 || out.SelectedAuditIndices[1] != 3 {
		t.Fatal("selected-to-audit mapping changed")
	}
	for i, row := range rows {
		if out.Audit[i] != row {
			t.Fatalf("audit row %d changed", i)
		}
	}
	// Independent parents retain every T, including T at weight 0.
	parents := []Parent{
		bridgeParent(t, 5, Known, roleplan.DevelopmentTrain, 2, []int{0, 2}, []int{1, 2}),
		bridgeParent(t, 8, Known, roleplan.DevelopmentValidation, 4, []int{1, 6}, []int{0, 6}),
	}
	want, err := Project(parents)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out.Projection, want) {
		t.Fatal("bridge differs from Project columns or counters")
	}
	p := out.Projection
	if !reflect.DeepEqual(p.Development.Labels, []float64{1, 0, 1, 0, 0}) || !reflect.DeepEqual(p.Development.SampleWeights, []float64{0, 1, 1, 0, 0}) ||
		!reflect.DeepEqual(p.Validation.Labels, []float64{0, 1, 0, 0, 0, 0, 1, 0}) || !reflect.DeepEqual(p.Validation.SampleWeights, []float64{1, 0, 0, 0, 0, 0, 1, 0}) {
		t.Fatal("zero-weight truth or role-specific eligibility changed")
	}
	if p.FeatureScans != 26 || p.Development.Excluded != 0 || p.Validation.Excluded != 0 {
		t.Fatal("withheld audit entered scans or denominators")
	}
	if p.Counts != ([3]RoleCounts{
		{Parents: 1, Candidates: 5, FitRows: 5, ZeroWeightFitRows: 3, WholeGroups: 1, PositiveWeightGroups: 1},
		{Parents: 1, Candidates: 8, FitRows: 8, ZeroWeightFitRows: 6, WholeGroups: 1, PositiveWeightGroups: 1},
		{},
	}) || p.DevelopmentAUC.Denominators != (Denominators{OriginalFitRows: 5, PositiveWeightRows: 2, ZeroWeightRows: 3, PositiveLabels: 1, NegativeLabels: 1}) ||
		p.ValidationAUC.Denominators != (Denominators{OriginalFitRows: 8, PositiveWeightRows: 2, ZeroWeightRows: 6, PositiveLabels: 1, NegativeLabels: 1}) {
		t.Fatal("selected-only counters changed")
	}
	if p.DevelopmentRefs[0] != (RowRef{0, 0}) || p.ValidationRefs[0] != (RowRef{1, 0}) ||
		!reflect.DeepEqual(p.DevelopmentAUC.Refs, []RowRef{{0, 1}, {0, 2}}) || !reflect.DeepEqual(p.ValidationAUC.Refs, []RowRef{{1, 0}, {1, 6}}) {
		t.Fatal("explicit selected parent references changed")
	}
}

func TestCohortAuditOnlyHasZeroProjectionAndPreservesNullableTruth(t *testing.T) {
	specs := []bridgeSpec{
		bridgeCase("partial-audit", 11, bridgeTrain, "TFUFF", "01000", "00000"),
		bridgeCase("ambiguous-audit", 12, bridgeVal, "TFFFF", "00000", "00000"),
		bridgeCase("calibration-audit", 13, bridgeCal, "TTFFF", "00000", "00000"),
		bridgeCase("zero-audit", 14, bridgeTrain, "TFFFF", "00000", "00000"),
	}
	specs[1].ambiguous = true
	rows := make([]cohort.CohortExample, len(specs))
	classes := []string{"unknown_containing", "ambiguous", "known_many", "known_one"}
	for i, s := range specs {
		rows[i] = bridgeRow(t, s)
		t.Run(s.id, func(t *testing.T) {
			out, err := ProjectCohort([]cohort.CohortExample{rows[i]})
			if err != nil || out.Count != 1 || out.SelectedCount != 0 || !reflect.DeepEqual(out.Projection, Projection{}) || out.Audit[0] != rows[i] || out.Audit[0].Class() != classes[i] {
				t.Fatal("audit-only truth or zero projection changed", err)
			}
		})
	}
	out, err := ProjectCohort(rows)
	if err != nil || out.Count != 4 || out.SelectedCount != 0 || !reflect.DeepEqual(out.Projection, Projection{}) {
		t.Fatal("zero-selected audit must succeed", err)
	}
	s := out.Audit[0].Supervision()
	if s.Count != 5 || s.States[0] != cohort.CohortTrue || !s.LabelKnown[0] || !s.Labels[0] || s.LossWeights[0] != 0 ||
		s.States[2] != cohort.CohortUnknown || s.LabelKnown[2] || s.Labels[2] || s.LossWeights[2] != 0 || s.EvaluationEligible[2] || s.LossWeights[1] != 1 {
		t.Fatal("partial-U truth or masks changed")
	}
	for i := s.Count; i < shortclaim.MaxCandidates; i++ {
		if s.States[i] != 0 || s.LabelKnown[i] || s.Labels[i] || s.LossWeights[i] != 0 || s.EvaluationEligible[i] {
			t.Fatal("unused supervision slot is not zero")
		}
	}
	if !out.Audit[1].Supervision().Ambiguous || !out.Audit[1].Supervision().Labels[0] || !out.Audit[2].Supervision().Labels[1] {
		t.Fatal("ambiguous or calibration declared T was erased")
	}
}

func TestCohortFiveAndEightCandidateTruth(t *testing.T) {
	tests := []struct {
		name, states, class string
		truth               TruthState
		acceptable          []int
	}{
		{"five-none", "FFFFF", "known_none", NoAnswer, nil},
		{"five-all", "TTTTT", "known_all", Known, []int{0, 1, 2, 3, 4}},
		{"five-many", "FTFFT", "known_many", Known, []int{1, 4}},
		{"eight-none", "FFFFFFFF", "known_none", NoAnswer, nil},
		{"eight-all", "TTTTTTTT", "known_all", Known, []int{0, 1, 2, 3, 4, 5, 6, 7}},
		{"eight-many", "FTFFFFFT", "known_many", Known, []int{1, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := len(tt.states)
			s := bridgeCase(tt.name, 21, bridgeTrain, tt.states, strings.Repeat("1", n), strings.Repeat("0", n))
			e := bridgeRow(t, s)
			eligible := make([]int, n)
			labels := make([]float64, n)
			for i := range eligible {
				eligible[i] = i
			}
			for _, i := range tt.acceptable {
				labels[i] = 1
			}
			want, err := Project([]Parent{bridgeParent(t, n, tt.truth, roleplan.DevelopmentTrain, 21, tt.acceptable, eligible)})
			if err != nil {
				t.Fatal(err)
			}
			out, err := ProjectCohort([]cohort.CohortExample{e})
			if err != nil || out.Count != 1 || out.SelectedCount != 1 || out.SelectedAuditIndices[0] != 0 || out.Audit[0].Class() != tt.class || !reflect.DeepEqual(out.Projection, want) || !reflect.DeepEqual(out.Projection.Development.Labels, labels) {
				t.Fatal("5/8 declared truth differs", err)
			}
			p := out.Projection.Parents[0]
			if p.Truth != tt.truth || p.AcceptableCount != len(tt.acceptable) || out.Projection.FeatureScans != 2*n ||
				out.Projection.DevelopmentAUC.Denominators != (Denominators{OriginalFitRows: n, PositiveWeightRows: n, PositiveLabels: len(tt.acceptable), NegativeLabels: n - len(tt.acceptable)}) {
				t.Fatal("truth or counters changed")
			}
		})
	}
}

func TestCohortRejectsHiddenCrossRoleAuditRows(t *testing.T) {
	base := bridgeCase("visible-train", 31, bridgeTrain, "TFFFF", "01000", "00000")
	withheld := []bridgeSpec{
		bridgeCase("hidden-partial", 32, bridgeVal, "TFUFF", "00000", "01000"),
		bridgeCase("hidden-ambiguous", 32, bridgeVal, "TFFFF", "00000", "00000"),
		bridgeCase("hidden-calibration", 32, bridgeCal, "TFFFF", "00000", "00000"),
		bridgeCase("hidden-zero", 32, bridgeVal, "TFFFF", "00000", "00000"),
	}
	withheld[1].ambiguous = true
	for _, hidden := range withheld {
		for _, shared := range []string{"whole-group", "source-family"} {
			t.Run(hidden.id+"/"+shared, func(t *testing.T) {
				s := hidden
				if shared == "whole-group" {
					s.group = base.group
				} else {
					s.family = base.family
				}
				a, b := bridgeRow(t, base), bridgeRow(t, s)
				bridgeReject(t, []cohort.CohortExample{a, b})
				bridgeReject(t, []cohort.CohortExample{b, a})
			})
		}
	}
	same := base
	same.id = "same-role"
	same.group = 33
	out, err := ProjectCohort([]cohort.CohortExample{bridgeRow(t, base), bridgeRow(t, same)})
	if err != nil || out.SelectedCount != 2 {
		t.Fatal("same-role shared source family rejected", err)
	}
}

func TestCohortRejectsPinsDuplicateInvalidAndParentBounds(t *testing.T) {
	base := bridgeCase("binding-base", 41, bridgeTrain, "TFFFF", "10000", "00000")
	a := bridgeRow(t, base)
	for _, pin := range []string{"mask", "roles-groups"} {
		t.Run(pin, func(t *testing.T) {
			s := bridgeCase("binding-other", 42, bridgeCal, "TFFFF", "00000", "00000")
			if pin == "mask" {
				s.mask = sha256.Sum256([]byte("different-synthetic-mask-policy"))
			} else {
				s.roles = sha256.Sum256([]byte("different-synthetic-role-membership"))
			}
			b := bridgeRow(t, s)
			bridgeReject(t, []cohort.CohortExample{a, b})
			bridgeReject(t, []cohort.CohortExample{b, a})
		})
	}
	bridgeReject(t, []cohort.CohortExample{a, a})
	bridgeReject(t, []cohort.CohortExample{{}, a})
	bridgeReject(t, []cohort.CohortExample{a, {}})
	bridgeReject(t, nil)
	bridgeReject(t, []cohort.CohortExample{})
	tooMany := make([]cohort.CohortExample, 61)
	for i := range tooMany {
		s := base
		s.id = "bound-" + strconv.Itoa(i)
		s.family = s.id
		s.group = i + 1
		tooMany[i] = bridgeRow(t, s)
	}
	bridgeReject(t, tooMany)
}

func TestCohortAuditAccessorsAndSnapshotsAreCopies(t *testing.T) {
	s := bridgeCase("copy-audit", 51, bridgeTrain, "TFUFF", "01000", "00000")
	e := bridgeRow(t, s)
	rows := []cohort.CohortExample{e}
	out, err := ProjectCohort(rows)
	if err != nil {
		t.Fatal(err)
	}
	text := out.Audit[0].Input().Prepared()
	supervision := out.Audit[0].Supervision()
	binding := out.Audit[0].Binding()
	metadata := out.Audit[0].Metadata()
	wantText, wantSupervision, wantBinding, wantMetadata := text, supervision, binding, metadata
	text.Request = "changed copied request"
	text.Candidates[0].Text = "changed copied candidate"
	text.Candidates[0].ID = "changed-id"
	text.Count = 8
	supervision.States[2] = cohort.CohortFalse
	supervision.LabelKnown[2] = true
	supervision.Labels[0] = false
	supervision.LossWeights[0] = 1
	supervision.EvaluationEligible[0] = true
	supervision.Count = 8
	supervision.Ambiguous = true
	binding.RowSHA256[0] ^= 1
	binding.SourceSHA256[0] ^= 1
	binding.EvidenceSHA256[0] ^= 1
	binding.MaskPolicySHA256[0] ^= 1
	binding.RolesGroupsSHA256[0] ^= 1
	binding.Metadata.StableID = "changed-binding-id"
	metadata.Role = bridgeVal
	metadata.WholeGroup = 999
	metadata.SourceFamily = "changed-family"
	if out.Audit[0].Input().Prepared() != wantText || out.Audit[0].Supervision() != wantSupervision || out.Audit[0].Binding() != wantBinding || out.Audit[0].Metadata() != wantMetadata || out.Audit[0] != e {
		t.Fatal("accessor mutation changed audit")
	}
	rows[0] = cohort.CohortExample{}
	if out.Audit[0] != e {
		t.Fatal("caller mutation changed audit")
	}
	out.Audit[0] = cohort.CohortExample{}
	if e.Input().Prepared() != wantText || e.Supervision() != wantSupervision || e.Binding() != wantBinding {
		t.Fatal("audit mutation changed original")
	}
}

var bridgeResult CohortProjection

// Fixtures and expectations stay outside the timer. Project-owned-B reports
// Project payload, excluding RSS, scratch and fixed-60 cohort audit arrays.
func BenchmarkProjectCohort(b *testing.B) {
	b.Run("audit-only-zero-selected", func(b *testing.B) {
		s := bridgeCase("bench-audit", 61, bridgeTrain, "TFUFF", "01000", "00000")
		rows := []cohort.CohortExample{bridgeRow(b, s)}
		want := CohortProjection{Count: 1}
		want.Audit[0] = rows[0]
		bridgeBench(b, rows, want)
	})
	b.Run("known-five-train-eight-validation", func(b *testing.B) {
		rows := []cohort.CohortExample{
			bridgeRow(b, bridgeCase("bench-train", 62, bridgeTrain, "TFTFF", "01100", "00000")),
			bridgeRow(b, bridgeCase("bench-validation", 63, bridgeVal, "FTFFFFTF", "00000000", "10000010")),
		}
		p, err := Project([]Parent{
			bridgeParent(b, 5, Known, roleplan.DevelopmentTrain, 62, []int{0, 2}, []int{1, 2}),
			bridgeParent(b, 8, Known, roleplan.DevelopmentValidation, 63, []int{1, 6}, []int{0, 6}),
		})
		if err != nil {
			b.Fatal(err)
		}
		want := CohortProjection{Count: 2, SelectedCount: 2, Projection: p}
		copy(want.Audit[:], rows)
		want.SelectedAuditIndices[1] = 1
		bridgeBench(b, rows, want)
	})
}

func bridgeBench(b *testing.B, rows []cohort.CohortExample, want CohortProjection) {
	b.Helper()
	check, err := ProjectCohort(rows)
	if err != nil || !reflect.DeepEqual(check, want) {
		b.Fatal("synthetic benchmark control differs", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridgeResult, err = ProjectCohort(rows)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(bridgeResult.Projection.PayloadBytes), "project-owned-B")
}
