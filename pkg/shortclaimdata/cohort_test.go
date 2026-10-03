// SPDX-License-Identifier: Apache-2.0
package shortclaimdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"
)

// Synthetic declarations and opaque bindings, never qualified Golden samples.
type cohortTestCandidate struct {
	ID                 string `json:"metadata_id"`
	Text               string `json:"text"`
	State              string `json:"state"`
	Label              *bool  `json:"label"`
	LossWeight         int    `json:"loss_weight"`
	EvaluationEligible bool   `json:"evaluation_eligible"`
}
type cohortTestBindings struct {
	Source   string `json:"source_sha256"`
	Evidence string `json:"evidence_sha256"`
	Masks    string `json:"mask_policy_sha256"`
	Roles    string `json:"roles_groups_sha256"`
}
type cohortTestRow struct {
	Schema         string                `json:"schema"`
	StableID       string                `json:"stable_id"`
	Request        string                `json:"request"`
	Candidates     []cohortTestCandidate `json:"candidates"`
	Role           string                `json:"role"`
	WholeGroup     int                   `json:"whole_group"`
	Family         string                `json:"source_family"`
	Revision       string                `json:"source_revision"`
	TextRevision   string                `json:"text_revision"`
	Policy         string                `json:"feature_policy"`
	Interpretation string                `json:"interpretation"`
	Bindings       cohortTestBindings    `json:"bindings"`
}

func cohortTestLabel(value bool) *bool    { return &value }
func cohortTestHex(value [32]byte) string { return hex.EncodeToString(value[:]) }
func cohortTestFixture(n int) (cohortTestRow, CohortBinding) {
	binding := CohortBinding{
		SourceSHA256: [32]byte{0xab}, EvidenceSHA256: [32]byte{0xcd},
		MaskPolicySHA256: [32]byte{0xef}, RolesGroupsSHA256: [32]byte{0x12},
		Metadata: CohortMetadata{StableID: "owned-cohort-case", SourceFamily: "owned-family", SourceRevision: strings.Repeat("b", 40), TextRevision: "owned-text-v1", Role: "development_train", WholeGroup: 101},
	}
	var candidates [8]cohortTestCandidate
	for i := 0; i < n; i++ {
		candidates[i] = cohortTestCandidate{ID: "owned-" + strconv.Itoa(i), Text: "Keep candidate values after validation.", State: "F", Label: cohortTestLabel(false), LossWeight: 1}
	}
	m := binding.Metadata
	return cohortTestRow{
		Schema: CohortSchema, StableID: m.StableID, Request: "Keep the first value after validation.", Candidates: candidates[:n], Role: m.Role, WholeGroup: m.WholeGroup,
		Family: m.SourceFamily, Revision: m.SourceRevision, TextRevision: m.TextRevision, Policy: FeaturePolicy, Interpretation: "unambiguous",
		Bindings: cohortTestBindings{cohortTestHex(binding.SourceSHA256), cohortTestHex(binding.EvidenceSHA256), cohortTestHex(binding.MaskPolicySHA256), cohortTestHex(binding.RolesGroupsSHA256)},
	}, binding
}
func cohortTestEncode(t testing.TB, row cohortTestRow, binding CohortBinding) ([]byte, CohortBinding) {
	t.Helper()
	raw, err := json.Marshal(row)
	if err != nil {
		t.Fatal("synthetic fixture encoding")
	}
	binding.RowSHA256 = sha256.Sum256(raw)
	return raw, binding
}
func cohortTestLoad(t testing.TB, row cohortTestRow, binding CohortBinding) CohortExample {
	t.Helper()
	raw, binding := cohortTestEncode(t, row, binding)
	example, err := LoadCohortRow(bytes.NewReader(raw), binding)
	if err != nil || !example.Valid() {
		t.Fatal("synthetic valid row rejected", err)
	}
	return example
}
func cohortTestReplace(t testing.TB, raw []byte, old, replacement string) []byte {
	t.Helper()
	if !bytes.Contains(raw, []byte(old)) {
		t.Fatal("synthetic replacement anchor missing")
	}
	return bytes.Replace(raw, []byte(old), []byte(replacement), 1)
}
func cohortTestReject(t testing.TB, raw []byte, binding CohortBinding, want Error) {
	t.Helper()
	// Reach parsing with a fresh row pin; stale pins have a separate control.
	binding.RowSHA256 = sha256.Sum256(raw)
	example, err := LoadCohortRow(bytes.NewReader(raw), binding)
	if err != want || example != (CohortExample{}) || example.Valid() || example.Class() != "invalid" {
		t.Fatal("rejected row retained data or returned wrong diagnostic", err, want)
	}
}

func TestCohortDeclaredClassesFiveAndEight(t *testing.T) {
	for _, n := range []int{5, 8} {
		for _, class := range []string{"known_none", "known_one", "known_many", "known_all", "unknown_containing", "unknown_only"} {
			t.Run(strconv.Itoa(n)+"/"+class, func(t *testing.T) {
				row, binding := cohortTestFixture(n)
				for i := range row.Candidates {
					c := &row.Candidates[i]
					positive := class == "known_all" || (class == "known_one" && i == n-1) || (class == "known_many" && (i == 0 || i == n-1))
					if positive {
						c.State, c.Label = "T", cohortTestLabel(true)
					}
					if class == "unknown_only" || (class == "unknown_containing" && i == n-1) {
						c.State, c.Label, c.LossWeight = "U", nil, 0
					}
				}
				example := cohortTestLoad(t, row, binding)
				wantClass := class
				if class == "unknown_only" {
					wantClass = "unknown_containing"
				}
				if example.Class() != wantClass {
					t.Fatal("class", example.Class())
				}
				s := example.Supervision()
				if s.Count != n || s.Ambiguous {
					t.Fatal("framing")
				}
				for i, c := range row.Candidates {
					state := CohortFalse
					if c.State == "T" {
						state = CohortTrue
					}
					if c.State == "U" {
						state = CohortUnknown
					}
					known, positive := c.Label != nil, c.Label != nil && *c.Label
					if s.States[i] != state || s.LabelKnown[i] != known || s.Labels[i] != positive || s.LossWeights[i] != uint8(c.LossWeight) || s.EvaluationEligible[i] {
						t.Fatal("truth or mask changed")
					}
				}
				for i := n; i < 8; i++ {
					if s.States[i] != 0 || s.LabelKnown[i] || s.Labels[i] || s.LossWeights[i] != 0 || s.EvaluationEligible[i] {
						t.Fatal("unused slot")
					}
				}
			})
		}
	}
}

func TestCohortTruthSurvivesAmbiguityAndZeroLoss(t *testing.T) {
	for _, n := range []int{5, 8} {
		row, binding := cohortTestFixture(n)
		row.Candidates[0].State, row.Candidates[0].Label, row.Candidates[0].LossWeight = "T", cohortTestLabel(true), 0
		example := cohortTestLoad(t, row, binding)
		s := example.Supervision()
		if example.Class() != "known_one" || !s.LabelKnown[0] || !s.Labels[0] || s.LossWeights[0] != 0 || s.LossWeights[1] != 1 {
			t.Fatal("zero weight changed truth")
		}
		row.Interpretation = "ambiguous"
		for i := range row.Candidates {
			row.Candidates[i].LossWeight = 0
		}
		example = cohortTestLoad(t, row, binding)
		s = example.Supervision()
		if example.Class() != "ambiguous" || !s.Ambiguous || !s.LabelKnown[0] || !s.Labels[0] || !s.LabelKnown[1] || s.Labels[1] {
			t.Fatal("ambiguous truth")
		}
		for i := 0; i < n; i++ {
			if s.LossWeights[i] != 0 || s.EvaluationEligible[i] {
				t.Fatal("ambiguous mask")
			}
		}
		row.Candidates[n-1].State, row.Candidates[n-1].Label = "U", nil
		example = cohortTestLoad(t, row, binding)
		if example.Class() != "ambiguous" || example.Supervision().LabelKnown[n-1] || example.Supervision().Labels[n-1] {
			t.Fatal("ambiguous U label")
		}
	}
}

func TestCohortRoleAndUnknownMasks(t *testing.T) {
	for _, role := range []string{"development_train", "development_validation", "development_calibration"} {
		row, binding := cohortTestFixture(5)
		row.Role, binding.Metadata.Role = role, role
		for i := range row.Candidates {
			row.Candidates[i].LossWeight = 0
		}
		if role == "development_validation" {
			row.Candidates[0].EvaluationEligible = true
		}
		cohortTestLoad(t, row, binding)
		bad := row
		bad.Candidates = append([]cohortTestCandidate(nil), row.Candidates...)
		if role == "development_train" {
			bad.Candidates[0].EvaluationEligible = true
		} else {
			bad.Candidates[0].LossWeight = 1
		}
		raw, expected := cohortTestEncode(t, bad, binding)
		cohortTestReject(t, raw, expected, ErrCohortMask)
	}
	for _, mode := range []string{"unknown_loss", "unknown_evaluation", "ambiguous_loss", "ambiguous_evaluation", "calibration_evaluation"} {
		row, binding := cohortTestFixture(5)
		for i := range row.Candidates {
			row.Candidates[i].LossWeight = 0
		}
		switch mode {
		case "unknown_loss", "unknown_evaluation":
			row.Candidates[0].State, row.Candidates[0].Label = "U", nil
		case "ambiguous_loss", "ambiguous_evaluation":
			row.Interpretation = "ambiguous"
		case "calibration_evaluation":
			row.Role, binding.Metadata.Role = "development_calibration", "development_calibration"
		}
		if mode == "unknown_evaluation" || mode == "ambiguous_evaluation" {
			row.Role, binding.Metadata.Role = "development_validation", "development_validation"
		}
		if strings.HasSuffix(mode, "loss") {
			row.Candidates[0].LossWeight = 1
		} else {
			row.Candidates[0].EvaluationEligible = true
		}
		raw, expected := cohortTestEncode(t, row, binding)
		cohortTestReject(t, raw, expected, ErrCohortMask)
	}
}

func TestCohortValueCopiesAndMetadataIsolation(t *testing.T) {
	row, binding := cohortTestFixture(5)
	for i := range row.Candidates {
		row.Candidates[i].LossWeight = 0
	}
	example := cohortTestLoad(t, row, binding)
	originalInput, originalSupervision, originalBinding := example.Input().Prepared(), example.Supervision(), example.Binding()
	s := example.Supervision()
	s.Labels[0], s.LabelKnown[0], s.LossWeights[0], s.States[7] = true, false, 255, CohortTrue
	b := example.Binding()
	b.Metadata.Role, b.SourceSHA256[0] = "changed", 0
	m := example.Metadata()
	m.WholeGroup = 999
	p := example.Input().Prepared()
	p.Request, p.Candidates[0].NormalizedText = "changed", "forged"
	if example.Supervision() != originalSupervision || example.Binding() != originalBinding || example.Input().Prepared() != originalInput {
		t.Fatal("accessor alias")
	}
	row.StableID, row.Family, row.Revision, row.TextRevision, row.Role, row.WholeGroup = "other-owned-case", "other-owned-family", strings.Repeat("c", 40), "other-text-v2", "development_calibration", 202
	binding.Metadata = CohortMetadata{row.StableID, row.Family, row.Revision, row.TextRevision, row.Role, row.WholeGroup}
	for i := range row.Candidates {
		row.Candidates[i].ID = "other-" + strconv.Itoa(i)
	}
	other := cohortTestLoad(t, row, binding).Input().Prepared()
	if other.NormalizedRequest != originalInput.NormalizedRequest || other.Provenance != CohortInputProvenance {
		t.Fatal("metadata in input")
	}
	for i := 0; i < originalInput.Count; i++ {
		if other.Candidates[i].NormalizedText != originalInput.Candidates[i].NormalizedText {
			t.Fatal("metadata in text")
		}
	}
}

type cohortTestReadCounter struct {
	Reads, Consumed int
	Data            *bytes.Reader
}

func (r *cohortTestReadCounter) Read(dst []byte) (int, error) {
	r.Reads++
	if r.Data == nil {
		return 0, errors.New("owned private-path secret error control")
	}
	n, err := r.Data.Read(dst)
	r.Consumed += n
	return n, err
}

func TestCohortExternalBindingPrecedesReading(t *testing.T) {
	row, binding := cohortTestFixture(5)
	_, binding = cohortTestEncode(t, row, binding)
	for i := 0; i < 9; i++ {
		bad := binding
		pins := [...]*[32]byte{&bad.RowSHA256, &bad.SourceSHA256, &bad.EvidenceSHA256, &bad.MaskPolicySHA256, &bad.RolesGroupsSHA256}
		if i < 5 {
			*pins[i] = [32]byte{}
		} else {
			switch i {
			case 5:
				bad.Metadata.WholeGroup = 0
			case 6:
				bad.Metadata.Role = "protected_final"
			case 7:
				bad.Metadata.StableID = "../private"
			case 8:
				bad.Metadata.SourceRevision = strings.Repeat("B", 40)
			}
		}
		r := &cohortTestReadCounter{}
		example, err := LoadCohortRow(r, bad)
		if err != ErrCohortPin || r.Reads != 0 || example != (CohortExample{}) {
			t.Fatal("invalid external authority consumed row bytes")
		}
	}
	if example, err := LoadCohortRow(nil, binding); err != ErrRead || example != (CohortExample{}) {
		t.Fatal("valid binding hid missing reader")
	}
}

func TestCohortFrozenBindingMismatches(t *testing.T) {
	row, binding := cohortTestFixture(5)
	raw, binding := cohortTestEncode(t, row, binding)
	for i := 0; i < 11; i++ {
		bad := binding
		pins := [...]*[32]byte{&bad.RowSHA256, &bad.SourceSHA256, &bad.EvidenceSHA256, &bad.MaskPolicySHA256, &bad.RolesGroupsSHA256}
		if i < 5 {
			pins[i][0] ^= 1
		} else {
			switch i {
			case 5:
				bad.Metadata.Role = "development_validation"
			case 6:
				bad.Metadata.WholeGroup++
			case 7:
				bad.Metadata.SourceFamily = "other-family"
			case 8:
				bad.Metadata.SourceRevision = strings.Repeat("c", 40)
			case 9:
				bad.Metadata.TextRevision = "other-v2"
			case 10:
				bad.Metadata.StableID = "other-case"
			}
		}
		example, err := LoadCohortRow(bytes.NewReader(raw), bad)
		if err != ErrCohortPin || example != (CohortExample{}) {
			t.Fatal("frozen binding mismatch admitted a row")
		}
	}
	// Identical opaque bindings cannot admit changed mask bytes under a stale pin.
	changed := cohortTestReplace(t, raw, `"loss_weight":1`, `"loss_weight":0`)
	if example, err := LoadCohortRow(bytes.NewReader(changed), binding); err != ErrCohortPin || example != (CohortExample{}) {
		t.Fatal("stale row pin admitted changed masks")
	}
}

func TestCohortStrictSyntaxAndTypes(t *testing.T) {
	row, binding := cohortTestFixture(5)
	raw, binding := cohortTestEncode(t, row, binding)
	cases := []struct {
		old, replacement string
		want             Error
	}{
		{`"schema":`, `"schema":"ignored","schema":`, ErrDuplicate},
		{`"stable_id":`, `"sche\u006da":"ignored","stable_id":`, ErrDuplicate},
		{`"stable_id":`, `"SCHEMA":"ignored","stable_id":`, ErrDuplicate},
		{`"schema":`, `"Schema":`, ErrField},
		{`"stable_id":`, `"extra":null,"stable_id":`, ErrField},
		{`"state":"F"`, `"state":"F","st\u0061te":"U"`, ErrDuplicate},
		{`"evidence_sha256":`, `"SOURCE_SHA256":"ignored","evidence_sha256":`, ErrDuplicate},
		{`"label":false,`, ``, ErrMissing},
		{`"label":false`, `"label":null`, ErrLabel},
		{`"label":false`, `"label":0`, ErrLabel},
		{`"loss_weight":1,`, ``, ErrMissing},
		{`"evaluation_eligible":false`, `"evaluation_eligible":null`, ErrCohortMask},
		{`"interpretation":"unambiguous",`, ``, ErrMissing},
		{`"interpretation":"unambiguous"`, `"interpretation":"pending"`, ErrCohortState},
		{`"state":"F"`, `"state":"false"`, ErrCohortState},
		{`"state":"F"`, `"state":null`, ErrJSON},
		{`"state":"F"`, `"state":"T"`, ErrLabel},
		{`"state":"F"`, `"state":"U"`, ErrLabel},
		{row.Revision, `../private`, ErrMetadata},
		{FeaturePolicy, `whole_json`, ErrMetadata},
		{CohortSchema, Schema, ErrSchema},
		{`Keep the first value`, `Keep \ud800 value`, ErrUnicode},
		{`Keep the first value`, `Keep \udc00 value`, ErrUnicode},
		{`Keep the first value`, "Keep " + string([]byte{0xff}) + " value", ErrUnicode},
		{row.Bindings.Source, strings.ToUpper(row.Bindings.Source), ErrCohortPin},
		{row.Bindings.Evidence, `abcd`, ErrCohortPin},
		{row.Bindings.Masks, strings.Repeat("0", 64), ErrCohortPin},
		{row.Bindings.Roles, `../private`, ErrCohortPin},
	}
	for i, tc := range cases {
		t.Run("syntax_"+strconv.Itoa(i), func(t *testing.T) {
			cohortTestReject(t, cohortTestReplace(t, raw, tc.old, tc.replacement), binding, tc.want)
		})
	}
	for _, value := range []string{"null", "1.0", "1e0", "-0", "2"} {
		cohortTestReject(t, cohortTestReplace(t, raw, `"loss_weight":1`, `"loss_weight":`+value), binding, ErrWeight)
	}
	for _, value := range []string{"0", "65536", "101.0", "101e0"} {
		cohortTestReject(t, cohortTestReplace(t, raw, `"whole_group":101`, `"whole_group":`+value), binding, ErrMetadata)
	}
	for _, invalid := range [][]byte{nil, []byte("null"), append(append([]byte(nil), raw...), []byte("{}")...), append(append([]byte(nil), raw...), []byte("garbage")...), append(append([]byte{'['}, raw...), ']')} {
		cohortTestReject(t, invalid, binding, ErrJSON)
	}
}

func TestCohortCandidateBoundsAndTruthConsistency(t *testing.T) {
	for _, n := range []int{0, 1, 4, 6, 7, 9} {
		row, binding := cohortTestFixture(8)
		if n <= 8 {
			row.Candidates = row.Candidates[:n]
		} else {
			row.Candidates = append(row.Candidates, cohortTestCandidate{ID: "ninth", Text: "Keep values.", State: "F", Label: cohortTestLabel(false)})
		}
		raw, expected := cohortTestEncode(t, row, binding)
		cohortTestReject(t, raw, expected, ErrCandidates)
	}
	row, binding := cohortTestFixture(5)
	row.Candidates[1].ID = row.Candidates[0].ID
	raw, expected := cohortTestEncode(t, row, binding)
	cohortTestReject(t, raw, expected, ErrInput)
	for _, c := range []cohortTestCandidate{
		{ID: "owned-0", Text: "Keep values.", State: "T", Label: nil},
		{ID: "owned-0", Text: "Keep values.", State: "F", Label: cohortTestLabel(true)},
		{ID: "owned-0", Text: "Keep values.", State: "U", Label: cohortTestLabel(false)},
	} {
		row, binding = cohortTestFixture(5)
		row.Candidates[0] = c
		raw, expected = cohortTestEncode(t, row, binding)
		cohortTestReject(t, raw, expected, ErrLabel)
	}
}

func TestCohortReadBoundsRedactionAndUnicodeScalars(t *testing.T) {
	row, binding := cohortTestFixture(5)
	raw, binding := cohortTestEncode(t, row, binding)
	oversized := append(append([]byte(nil), raw...), bytes.Repeat([]byte{' '}, MaxRowBytes+1-len(raw))...)
	binding.RowSHA256 = sha256.Sum256(oversized)
	r := &cohortTestReadCounter{Data: bytes.NewReader(oversized)}
	if example, err := LoadCohortRow(r, binding); err != ErrBounds || example != (CohortExample{}) || r.Consumed != MaxRowBytes+1 {
		t.Fatal("oversized read was not bounded at cap+1")
	}
	r = &cohortTestReadCounter{}
	example, err := LoadCohortRow(r, binding)
	if err != ErrRead || example != (CohortExample{}) || strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "secret") {
		t.Fatal("reader error leaked private diagnostic")
	}
	// Scalars and escaped literal backslashes are valid Unicode.
	for _, request := range []string{"Keep 😀 values after validation.", "Keep � values after validation.", `Keep \ud800 text after validation.`} {
		row.Request = request
		cohortTestLoad(t, row, binding)
	}
	if example, err := LoadDevelopmentRow(bytes.NewReader(raw)); err == nil || example != (Example{}) {
		t.Fatal("legacy Reader silently accepted opt-in cohort schema")
	}
	if _, err := LoadCohortRow(io.LimitReader(bytes.NewReader(raw), int64(len(raw)-1)), binding); err != ErrCohortPin {
		t.Fatal("truncated row bypassed exact row binding")
	}
}

func BenchmarkCohortRow(b *testing.B) {
	for _, n := range []int{5, 8} {
		b.Run("candidates_"+strconv.Itoa(n), func(b *testing.B) {
			b.StopTimer()
			row, binding := cohortTestFixture(n)
			raw, binding := cohortTestEncode(b, row, binding)
			var reader bytes.Reader
			b.ReportAllocs()
			b.SetBytes(int64(len(raw)))
			b.ResetTimer()
			b.StartTimer()
			for i := 0; i < b.N; i++ {
				// Reset framing; fixture and bindings remain outside the timer.
				reader.Reset(raw)
				example, err := LoadCohortRow(&reader, binding)
				if err != nil || !example.Valid() {
					b.Fatal("synthetic cohort load failed")
				}
			}
		})
	}
}
