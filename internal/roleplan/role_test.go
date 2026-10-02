package roleplan

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var syntheticSeed = []byte("synthetic-external-seed-v1")

func toyComponents(known, unknown int) []Component {
	cs := make([]Component, known+unknown)
	for i := range cs {
		cs[i] = Component{fmt.Sprintf("toy-component-%02d", i), int64(100 + 7*i), sha256.Sum256([]byte(fmt.Sprintf("synthetic-membership-%02d", i))), i < known}
	}
	return cs
}
func noExtraCoverage() CoveragePlan { return CoveragePlan{Declared: true} }
func errorIs(t *testing.T, e error, want string) {
	t.Helper()
	if e == nil || e.Error() != want {
		t.Fatalf("got %v, want %s", e, want)
	}
}
func TestSyntheticLargestRemainderLiteralTable(t *testing.T) {
	cases := []struct {
		n    int
		want [3]int
	}{{0, [3]int{0, 0, 0}}, {1, [3]int{1, 0, 0}}, {2, [3]int{1, 1, 0}}, {3, [3]int{2, 1, 0}}, {4, [3]int{2, 1, 1}}, {5, [3]int{3, 1, 1}}, {14, [3]int{8, 3, 3}}, {15, [3]int{9, 3, 3}}, {16, [3]int{10, 3, 3}}, {17, [3]int{10, 4, 3}}, {18, [3]int{11, 4, 3}}, {19, [3]int{11, 4, 4}}, {20, [3]int{12, 4, 4}}}
	for _, tc := range cases {
		got, e := AllocateCounts(tc.n)
		if e != nil || got != tc.want {
			t.Fatalf("n=%d got %v/%v want %v", tc.n, got, e, tc.want)
		}
	}
	_, e := AllocateCounts(-1)
	errorIs(t, e, "component_bounds")
	_, e = AllocateCounts(MaxComponents + 1)
	errorIs(t, e, "component_bounds")
}
func TestSyntheticLengthPrefixLiteralBytes(t *testing.T) {
	// Full bytes and SHA were derived separately from this Go implementation
	// using an explicit u64-BE byte construction, then fixed as literals.
	wantHex := "0000000000000023726969646f2d77686f6c652d636f6d706f6e656e742d6d656d626572736869702d7631000000000000000200000000000000016100000000000000016200000000000000010000000000000001610000000000000006736f75726365000000000000000162"
	m := Membership{[]string{"b", "a"}, []Relationship{{"a", "source", "b"}}}
	before := Membership{append([]string(nil), m.Members...), append([]Relationship(nil), m.Relationships...)}
	h, b, e := MembershipDigest(m)
	if e != nil || hex.EncodeToString(b) != wantHex || hex.EncodeToString(h[:]) != "57b4f7bd5cc092cdc4d6e09f815689974bdce87eee5e00db28b8e4bf472c480f" {
		t.Fatalf("encoding mismatch: %x/%x/%v", b, h, e)
	}
	if !reflect.DeepEqual(m, before) {
		t.Fatal("membership input mutated")
	}
	reordered := Membership{[]string{"a", "b"}, []Relationship{{"a", "source", "b"}}}
	h2, b2, e := MembershipDigest(reordered)
	if e != nil || h2 != h || !bytes.Equal(b, b2) {
		t.Fatal("canonical membership depends on input order")
	}
	// A changed directed relationship remains distinct, not silently undirected.
	h3, _, e := MembershipDigest(Membership{[]string{"a", "b"}, []Relationship{{"b", "source", "a"}}})
	if e != nil || h3 == h {
		t.Fatal("relationship direction lost")
	}
	var d [32]byte
	for i := range d {
		d[i] = byte(i)
	}
	orderedHash, e := OrderDigest(syntheticSeed, d)
	if e != nil || hex.EncodeToString(orderedHash[:]) != "53ec156f4d3b7a5b42d389476078f7af7e9eca4130b7a1a8dbd004f6a076fd8b" {
		t.Fatalf("order encoding mismatch: %x/%v", orderedHash, e)
	}
}
func TestSyntheticLengthPrefixesAvoidConcatenationCollisions(t *testing.T) {
	h1, _, e := MembershipDigest(Membership{Members: []string{"ab", "c"}})
	if e != nil {
		t.Fatal(e)
	}
	h2, _, e := MembershipDigest(Membership{Members: []string{"a", "bc"}})
	if e != nil {
		t.Fatal(e)
	}
	if h1 == h2 {
		t.Fatal("ambiguous concatenation")
	}
	h3, _, e := MembershipDigest(Membership{Members: []string{"a", "b", "c"}})
	if e != nil {
		t.Fatal(e)
	}
	if h3 == h1 || h3 == h2 {
		t.Fatal("count field lost")
	}
}
func TestSyntheticDeterminismLiteralRolesAndNoMutation(t *testing.T) {
	cs := toyComponents(16, 1)
	before := append([]Component(nil), cs...)
	seedBefore := append([]byte(nil), syntheticSeed...)
	got, e := Assign(cs, syntheticSeed, noExtraCoverage())
	if e != nil {
		t.Fatal(e)
	}
	// One externally specified synthetic seed, no seed search. The fixed role
	// vector comes from an independent SHA/LP implementation of the recipe.
	want := [16]Role{2, 0, 2, 1, 2, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0}
	for i, r := range want {
		if got.Assignments[i].Role != r || got.Assignments[i].OriginalGroupID != cs[i].OriginalGroupID || got.Assignments[i].MembershipSHA256 != cs[i].MembershipSHA256 {
			t.Fatalf("role %d: %+v", i, got.Assignments[i])
		}
	}
	if got.PlannedKnownCounts != ([3]int{10, 3, 3}) || got.ProvenanceCounts != ([3]int{11, 3, 3}) || got.UnknownOnlyComponents != 1 || got.Assignments[16].Role != DevelopmentTrain || !got.Assignments[16].ProvenanceOnly {
		t.Fatal("unknown provenance accounting changed")
	}
	if got.TrainingReady || got.ScorerFeatureProjectionPermitted || !got.UnknownMembersRemainUnknown || got.TransferDevelopmentGroups != 0 {
		t.Fatal("eligibility or feature claim fabricated")
	}
	if !reflect.DeepEqual(cs, before) || !bytes.Equal(syntheticSeed, seedBefore) {
		t.Fatal("caller inputs mutated")
	}
	again, e := Assign(cs, syntheticSeed, noExtraCoverage())
	if e != nil || !reflect.DeepEqual(again, got) {
		t.Fatal("same frozen input not deterministic")
	}
	reverse := append([]Component(nil), cs...)
	for i, j := 0, len(reverse)-1; i < j; i, j = i+1, j-1 {
		reverse[i], reverse[j] = reverse[j], reverse[i]
	}
	reversed, e := Assign(reverse, syntheticSeed, noExtraCoverage())
	if e != nil {
		t.Fatal(e)
	}
	lookup := componentIDs(cs)
	for _, a := range reversed.Assignments {
		index := componentIndex(lookup, a.ComponentID)
		if index < 0 || a.Role != got.Assignments[index].Role {
			t.Fatal("input-order dependent role")
		}
	}
	if got.Counters.OrderHashesAttempted != 16 || got.Counters.OrderHashesCompleted != 16 || got.Counters.AssignmentsReturned != 17 {
		t.Fatal("synthetic counters incorrect")
	}
}
func TestSyntheticTieComparatorRecipe(t *testing.T) {
	a := ordered{GroupID: 30}
	b := ordered{GroupID: 40}
	if !orderedLess(a, b) || orderedLess(b, a) {
		t.Fatal("group tie order")
	}
	b.Membership[31] = 1
	if !orderedLess(a, b) || orderedLess(b, a) {
		t.Fatal("membership digest tie order")
	}
	a.OrderHash[0] = 1
	if orderedLess(a, b) || !orderedLess(b, a) {
		t.Fatal("primary hash order")
	}
}
func TestSyntheticFloorsExcludeUnknownOnly(t *testing.T) {
	r, e := Assign(toyComponents(14, 10), syntheticSeed, noExtraCoverage())
	errorIs(t, e, "existing_labeled_floor_unmet")
	if r.KnownComponents != 14 || r.UnknownOnlyComponents != 10 || r.PlannedKnownCounts != ([3]int{8, 3, 3}) || len(r.Assignments) != 0 || r.Counters.AssignmentsReturned != 0 {
		t.Fatal("unknown-only counted as labeled or failure returned roles")
	}
	r, e = Assign(toyComponents(15, 0), syntheticSeed, noExtraCoverage())
	if e != nil || r.PlannedKnownCounts != ([3]int{9, 3, 3}) {
		t.Fatal("unchanged 9/3/3 floor not supported")
	}
	r, e = Assign(toyComponents(0, 16), syntheticSeed, noExtraCoverage())
	errorIs(t, e, "existing_labeled_floor_unmet")
	if r.Counters.OrderHashesAttempted != 0 || r.KnownComponents != 0 {
		t.Fatal("unknown-only hashed as labeled")
	}
}
func TestSyntheticCoverageExplicitAndFailClosed(t *testing.T) {
	cs := toyComponents(16, 1)
	ids := make([]string, 16)
	for i := range ids {
		ids[i] = cs[i].ID
	}
	p := CoveragePlan{Declared: true, Requirements: []CoverageRequirement{{"toy-declared-coverage", DevelopmentTrain, ids, 10}}}
	r, e := Assign(cs, syntheticSeed, p)
	if e != nil || r.Counters.CoverageRequirementsPassed != 1 {
		t.Fatal("explicit supported coverage rejected")
	}
	p.Requirements[0].MinimumKnownComponents = 11
	r, e = Assign(cs, syntheticSeed, p)
	errorIs(t, e, "predeclared_coverage_unmet")
	if r.Counters.CoverageRequirementsAttempted != 1 || r.Counters.CoverageRequirementsPassed != 0 || r.Counters.AssignmentsReturned != 0 || len(r.Assignments) != 0 {
		t.Fatal("failed coverage published assignments")
	}
	p = CoveragePlan{Declared: true, Requirements: []CoverageRequirement{{"toy-unknown-not-label", DevelopmentTrain, []string{cs[16].ID}, 1}}}
	r, e = Assign(cs, syntheticSeed, p)
	errorIs(t, e, "predeclared_coverage_unmet")
	if r.KnownComponents != 16 || r.UnknownOnlyComponents != 1 {
		t.Fatal("unknown counted toward coverage")
	}
	_, e = Assign(cs, syntheticSeed, CoveragePlan{})
	errorIs(t, e, "coverage_plan_undeclared")
}
func TestSyntheticComponentRefusals(t *testing.T) {
	cases := []struct {
		name   string
		change func([]Component) []Component
		want   string
	}{
		{"nil", func(c []Component) []Component { return nil }, "component_bounds"},
		{"too many", func(c []Component) []Component { return make([]Component, MaxComponents+1) }, "component_bounds"},
		{"empty id", func(c []Component) []Component { c[0].ID = ""; return c }, "component_metadata_invalid"},
		{"invalid utf8", func(c []Component) []Component { c[0].ID = string([]byte{255}); return c }, "component_metadata_invalid"},
		{"id bounds", func(c []Component) []Component { c[0].ID = strings.Repeat("a", MaxIDBytes+1); return c }, "component_metadata_invalid"},
		{"negative group", func(c []Component) []Component { c[0].OriginalGroupID = -1; return c }, "component_metadata_invalid"},
		{"empty digest", func(c []Component) []Component { c[0].MembershipSHA256 = [32]byte{}; return c }, "component_metadata_invalid"},
		{"duplicate id", func(c []Component) []Component { c[1].ID = c[0].ID; return c }, "component_id_duplicate"},
		{"duplicate digest", func(c []Component) []Component { c[1].MembershipSHA256 = c[0].MembershipSHA256; return c }, "membership_digest_duplicate"},
		{"split original group", func(c []Component) []Component { c[1].OriginalGroupID = c[0].OriginalGroupID; return c }, "original_group_split"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, e := Assign(tc.change(toyComponents(16, 0)), syntheticSeed, noExtraCoverage())
			errorIs(t, e, tc.want)
			if len(r.Assignments) != 0 || r.Counters.AssignmentsReturned != 0 || r.FailureCode != tc.want {
				t.Fatal("failure lost or published assignments")
			}
		})
	}
	_, e := Assign(toyComponents(16, 0), nil, noExtraCoverage())
	errorIs(t, e, "seed_bounds")
	_, e = Assign(toyComponents(16, 0), make([]byte, MaxSeedBytes+1), noExtraCoverage())
	errorIs(t, e, "seed_bounds")
}
func TestSyntheticCoverageRefusals(t *testing.T) {
	cs := toyComponents(16, 0)
	cases := []struct {
		name string
		p    CoveragePlan
		want string
	}{
		{"invalid role", CoveragePlan{true, []CoverageRequirement{{"toy", Role(3), []string{cs[0].ID}, 0}}}, "coverage_requirement_invalid"},
		{"negative requirement", CoveragePlan{true, []CoverageRequirement{{"toy", DevelopmentTrain, []string{cs[0].ID}, -1}}}, "coverage_requirement_invalid"},
		{"oversized requirement", CoveragePlan{true, []CoverageRequirement{{"toy", DevelopmentTrain, nil, 1}}}, "coverage_requirement_invalid"},
		{"absent component", CoveragePlan{true, []CoverageRequirement{{"toy", DevelopmentTrain, []string{"absent"}, 1}}}, "coverage_component_missing"},
		{"duplicate component", CoveragePlan{true, []CoverageRequirement{{"toy", DevelopmentTrain, []string{cs[0].ID, cs[0].ID}, 1}}}, "coverage_component_duplicate"},
		{"duplicate requirement", CoveragePlan{true, []CoverageRequirement{{"toy", DevelopmentTrain, nil, 0}, {"toy", DevelopmentTrain, nil, 0}}}, "coverage_requirement_duplicate"},
		{"requirement bounds", CoveragePlan{true, make([]CoverageRequirement, MaxComponents+1)}, "coverage_bounds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, e := Assign(cs, syntheticSeed, tc.p)
			errorIs(t, e, tc.want)
			if r.Counters.OrderHashesAttempted != 0 {
				t.Fatal("invalid plan executed ordering")
			}
		})
	}
}
func TestSyntheticMembershipRefusalsAndSplitFreeze(t *testing.T) {
	cases := []struct {
		name string
		m    Membership
		want string
	}{
		{"empty", Membership{}, "membership_bounds"},
		{"members bounds", Membership{Members: make([]string, MaxMembershipEntries+1)}, "membership_bounds"},
		{"relations bounds", Membership{Members: []string{"a"}, Relationships: make([]Relationship, MaxMembershipEntries+1)}, "membership_bounds"},
		{"empty member", Membership{Members: []string{""}}, "membership_id_invalid"},
		{"duplicate member", Membership{Members: []string{"a", "a"}}, "membership_member_duplicate"},
		{"unknown endpoint", Membership{[]string{"a"}, []Relationship{{"a", "source", "b"}}}, "membership_relationship_invalid"},
		{"duplicate relationship", Membership{[]string{"a", "b"}, []Relationship{{"a", "source", "b"}, {"a", "source", "b"}}}, "membership_relationship_duplicate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, b, e := MembershipDigest(tc.m)
			errorIs(t, e, tc.want)
			if b != nil {
				t.Fatal("failed encoding published bytes")
			}
		})
	}
	complete := Membership{[]string{"a", "b"}, []Relationship{{"a", "source", "b"}}}
	h, _, e := MembershipDigest(complete)
	if e != nil {
		t.Fatal(e)
	}
	frozen := []Component{{"toy-whole", 8, h, true}}
	if e = VerifyMembershipSet(frozen, []Membership{complete}); e != nil {
		t.Fatal(e)
	}
	split1 := Membership{Members: []string{"a"}}
	split2 := Membership{Members: []string{"b"}}
	h1, _, _ := MembershipDigest(split1)
	h2, _, _ := MembershipDigest(split2)
	split := []Component{{"toy-part-a", 8, h1, true}, {"toy-part-b", 8, h2, false}}
	errorIs(t, VerifyMembershipSet(split, []Membership{split1, split2}), "original_group_split")
	errorIs(t, VerifyAgainstFreeze(split, frozen), "frozen_components_mismatch")
	// If an attacker renames group IDs and omits an edge, opaque input alone
	// cannot discover it. Exact prior freeze identity is the trust boundary.
	split[1].OriginalGroupID = 9
	errorIs(t, VerifyAgainstFreeze(split, frozen), "frozen_components_mismatch")
	overlap1 := Membership{Members: []string{"a", "b"}}
	overlap2 := Membership{Members: []string{"b", "c"}}
	o1, _, _ := MembershipDigest(overlap1)
	o2, _, _ := MembershipDigest(overlap2)
	overlap := []Component{{"toy-overlap-a", 8, o1, true}, {"toy-overlap-b", 9, o2, true}}
	errorIs(t, VerifyMembershipSet(overlap, []Membership{overlap1, overlap2}), "membership_cross_component_overlap")
	altered := append([]Component(nil), frozen...)
	altered[0].MembershipSHA256 = h1
	errorIs(t, VerifyMembershipSet(altered, []Membership{complete}), "membership_pin_mismatch")
	unknownFrozen := toyComponents(16, 1)
	unknownAltered := append([]Component(nil), unknownFrozen...)
	unknownAltered[16].HasKnownMembers = true
	errorIs(t, VerifyAgainstFreeze(unknownAltered, unknownFrozen), "frozen_components_mismatch")
	if e = VerifyAgainstFreeze(unknownFrozen, append([]Component(nil), unknownFrozen...)); e != nil {
		t.Fatal(e)
	}
}

func TestSyntheticEncodedBytePreflight(t *testing.T) {
	simple := Membership{[]string{"a", "b"}, []Relationship{{"a", "source", "b"}}}
	n, e := encodedMembershipBytes(simple)
	if e != nil || n != 109 {
		t.Fatalf("literal length got %d/%v want109", n, e)
	}
	utf8Case := Membership{Members: []string{"é"}}
	h, b, e := MembershipDigest(utf8Case)
	if e != nil || len(b) != 69 || h == ([32]byte{}) {
		t.Fatal("UTF8 encoded byte length incorrect")
	}
	total := MaxEncodedMembershipBytes - 1
	if e = checkedEncodedAdd(&total, 1); e != nil || total != MaxEncodedMembershipBytes {
		t.Fatal("inclusive payload bound")
	}
	errorIs(t, checkedEncodedAdd(&total, 1), "membership_encoded_bounds")
	maxInt := int(^uint(0) >> 1)
	total = maxInt
	errorIs(t, checkedEncodedAdd(&total, 1), "membership_encoded_bounds")
	if total != maxInt {
		t.Fatal("overflow guard mutated length")
	}
	total = 1
	errorIs(t, checkedEncodedAdd(&total, maxInt), "membership_encoded_bounds")
	total = 0
	errorIs(t, checkedEncodedAdd(&total, -1), "membership_encoded_bounds")
	// Shared strings keep this refusal fixture small; its declared LP payload
	// would exceed 64MiB. Refusal precedes copies, Buffer.Grow and encoding.
	members := make([]string, 256)
	for i := range members {
		members[i] = fmt.Sprintf("%03d", i) + strings.Repeat("x", 509)
	}
	kind := strings.Repeat("k", 512)
	relations := make([]Relationship, 0, 65536)
	for _, from := range members {
		for _, to := range members {
			relations = append(relations, Relationship{from, kind, to})
		}
	}
	oversized := Membership{members, relations}
	_, encoded, e := MembershipDigest(oversized)
	errorIs(t, e, "membership_encoded_bounds")
	if encoded != nil {
		t.Fatal("oversized encoding emitted payload")
	}
}
func TestSyntheticFlatMembershipOverlap(t *testing.T) {
	first := Membership{Members: make([]string, 1024)}
	second := Membership{Members: make([]string, 1024)}
	for i := 0; i < 1024; i++ {
		first.Members[i] = fmt.Sprintf("toy-a-%04d", 1023-i)
		second.Members[i] = fmt.Sprintf("toy-b-%04d", i)
	}
	h1, _, e := MembershipDigest(first)
	if e != nil {
		t.Fatal(e)
	}
	h2, _, e := MembershipDigest(second)
	if e != nil {
		t.Fatal(e)
	}
	cs := []Component{{"toy-flat-a", 10, h1, true}, {"toy-flat-b", 20, h2, false}}
	if e = VerifyMembershipSet(cs, []Membership{first, second}); e != nil {
		t.Fatal(e)
	}
	second.Members[1023] = first.Members[42]
	h2, _, e = MembershipDigest(second)
	if e != nil {
		t.Fatal(e)
	}
	cs[1].MembershipSHA256 = h2
	errorIs(t, VerifyMembershipSet(cs, []Membership{first, second}), "membership_cross_component_overlap")
}

func TestSyntheticCoverageReferenceAggregateBoundAndIndexedOwnership(t *testing.T) {
	// A full bounded synthetic plan exercises ID resolution without executing
	// original data, finding a seed, or timing the role policy as a benchmark.
	cs := toyComponents(MaxComponents, 0)
	ids := make([]string, len(cs))
	for i := range ids {
		ids[i] = cs[len(cs)-1-i].ID
	}
	before := append([]string(nil), ids...)
	plan := CoveragePlan{Declared: true}
	for i := 0; i < MaxCoverageReferences/MaxComponents; i++ {
		plan.Requirements = append(plan.Requirements, CoverageRequirement{
			Key: fmt.Sprintf("toy-bound-%02d", i), Role: DevelopmentTrain,
			ComponentIDs: ids,
		})
	}
	compiled, e := validateCoverage(cs, plan)
	if e != nil || len(compiled.InputIndices) != MaxCoverageReferences || len(compiled.Offsets) != len(plan.Requirements)+1 || compiled.Offsets[len(compiled.Offsets)-1] != MaxCoverageReferences {
		t.Fatalf("inclusive reference bound refused: %v", e)
	}
	lookup := componentIDs(cs)
	for qi := range plan.Requirements {
		indices := compiled.InputIndices[compiled.Offsets[qi]:compiled.Offsets[qi+1]]
		for j, index := range indices {
			if index < 0 || index >= len(cs) || cs[index].ID != lookup[j].ID {
				t.Fatal("sorted ID did not bind original input index")
			}
		}
	}
	if !reflect.DeepEqual(ids, before) {
		t.Fatal("coverage IDs reordered in caller storage")
	}
	indexBefore := compiled.InputIndices[0]
	ids[0] = "caller-mutated-after-preparation"
	if compiled.InputIndices[0] != indexBefore {
		t.Fatal("compiled coverage aliases caller IDs")
	}
	// One extra reference is refused before missing/duplicate ID processing or
	// ordering. This is a resource limit, not a known-group readiness gate.
	plan.Requirements = append(plan.Requirements, CoverageRequirement{
		Key: "toy-over-bound", Role: DevelopmentTrain, ComponentIDs: []string{"missing"},
	})
	compiled, e = validateCoverage(cs, plan)
	errorIs(t, e, "coverage_reference_bounds")
	if compiled.Offsets != nil || compiled.InputIndices != nil {
		t.Fatal("over-bound plan returned partially compiled references")
	}
	result, e := Assign(cs, syntheticSeed, plan)
	errorIs(t, e, "coverage_reference_bounds")
	if result.Counters.OrderHashesAttempted != 0 || result.Counters.AssignmentsReturned != 0 || len(result.Assignments) != 0 {
		t.Fatal("over-bound coverage plan proceeded to ordering or returned roles")
	}
}

func TestSyntheticCoverageSortedDuplicatesAndRequirementIdentity(t *testing.T) {
	cs := toyComponents(16, 1)
	cases := []struct {
		name string
		plan CoveragePlan
		want string
	}{
		{"nonadjacent ID duplicate", CoveragePlan{true, []CoverageRequirement{{"toy-ids", DevelopmentTrain, []string{cs[15].ID, cs[1].ID, cs[7].ID, cs[1].ID}, 0}}}, "coverage_component_duplicate"},
		{"nonadjacent key-role duplicate", CoveragePlan{true, []CoverageRequirement{{"toy-key", DevelopmentTrain, nil, 0}, {"toy-between", DevelopmentCalibration, nil, 0}, {"toy-key", DevelopmentTrain, nil, 0}}}, "coverage_requirement_duplicate"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, e := Assign(cs, syntheticSeed, tc.plan)
			errorIs(t, e, tc.want)
			if result.Counters.OrderHashesAttempted != 0 || len(result.Assignments) != 0 {
				t.Fatal("duplicate plan reached ordering")
			}
		})
	}
	// A shared key in distinct roles and case-distinct keys are valid declared
	// requirements. Zero minima add no new required coverage or transfer gate.
	p := CoveragePlan{true, []CoverageRequirement{
		{"toy-shared", DevelopmentTrain, nil, 0},
		{"toy-shared", DevelopmentValidation, nil, 0},
		{"Toy-shared", DevelopmentTrain, nil, 0},
	}}
	r, e := Assign(cs, syntheticSeed, p)
	if e != nil || r.Counters.CoverageRequirementsPassed != 3 || r.TransferDevelopmentGroups != 0 {
		t.Fatalf("exact key-role identity or transfer0 changed: %v", e)
	}
}

func TestSyntheticDeclaredCoverageReorderingAndResultOwnership(t *testing.T) {
	cs := toyComponents(16, 1)
	ids := make([]string, 16)
	for i := range ids {
		ids[i] = cs[15-i].ID
	}
	p := CoveragePlan{true, []CoverageRequirement{
		{"toy-z-validation", DevelopmentValidation, ids, 3},
		{"toy-a-train", DevelopmentTrain, ids, 10},
	}}
	before := append([]string(nil), ids...)
	r, e := Assign(cs, syntheticSeed, p)
	if e != nil || r.Counters.CoverageRequirementsPassed != 2 || !reflect.DeepEqual(ids, before) {
		t.Fatalf("requirement offsets changed original plan association: %v", e)
	}
	assignment := r.Assignments[0]
	cs[0].ID = "caller-mutated-component"
	cs[0].MembershipSHA256[0] ^= 1
	ids[0] = "caller-mutated-coverage"
	if r.Assignments[0] != assignment {
		t.Fatal("returned assignment aliases caller metadata")
	}
}

func TestSyntheticAggregateEncodingPreflightBeforeMemberOrTupleCopies(t *testing.T) {
	// Each declared payload fits individually but the set exceeds64MiB. The
	// shared duplicate tuples intentionally become irrelevant: aggregate byte
	// preflight must refuse before per-membership encoding/duplicate work.
	memberA, memberB := strings.Repeat("a", 512), strings.Repeat("b", 512)
	memberC, memberD := strings.Repeat("c", 512), strings.Repeat("d", 512)
	kind := strings.Repeat("k", 512)
	first := Membership{Members: []string{memberA, memberB}, Relationships: make([]Relationship, 22000)}
	second := Membership{Members: []string{memberC, memberD}, Relationships: make([]Relationship, 22000)}
	for i := range first.Relationships {
		first.Relationships[i] = Relationship{memberA, kind, memberB}
		second.Relationships[i] = Relationship{memberC, kind, memberD}
	}
	for _, m := range []Membership{first, second} {
		if n, e := encodedMembershipBytes(m); e != nil || n >= MaxEncodedMembershipBytes {
			t.Fatalf("individual payload unexpectedly over bound: %d/%v", n, e)
		}
	}
	errorIs(t, VerifyMembershipSet(toyComponents(2, 0), []Membership{first, second}), "membership_encoded_bounds")
}

func TestSyntheticCoverageIDByteBoundsBeforeSortedProjection(t *testing.T) {
	cs := toyComponents(16, 0)
	for _, tc := range []struct {
		name string
		ids  []string
	}{
		{"empty", []string{""}},
		{"invalid UTF8", []string{string([]byte{'a', 255})}},
		{"one byte over", []string{strings.Repeat("a", MaxIDBytes+1)}},
		{"large common prefix", []string{strings.Repeat("a", 1<<20) + "x", strings.Repeat("a", 1<<20) + "y"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := CoveragePlan{true, []CoverageRequirement{{"toy-ID-bound", DevelopmentTrain, tc.ids, 0}}}
			compiled, e := validateCoverage(cs, p)
			errorIs(t, e, "coverage_component_id_invalid")
			if compiled.Offsets != nil || compiled.InputIndices != nil {
				t.Fatal("invalid ID allocated or exposed sorted projection")
			}
			result, e := Assign(cs, syntheticSeed, p)
			errorIs(t, e, "coverage_component_id_invalid")
			if result.Counters.OrderHashesAttempted != 0 || result.Counters.AssignmentsReturned != 0 || len(result.Assignments) != 0 {
				t.Fatal("invalid coverage ID reached ordering or assignment")
			}
		})
	}
	for _, tc := range []struct{ name, id string }{
		{"ASCII exact512", strings.Repeat("x", MaxIDBytes)},
		{"UTF8 exact512", strings.Repeat("é", MaxIDBytes/2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owned := append([]Component(nil), cs...)
			owned[0].ID = tc.id
			p := CoveragePlan{true, []CoverageRequirement{{"toy-inclusive-ID-bound", DevelopmentTrain, []string{tc.id}, 0}}}
			compiled, e := validateCoverage(owned, p)
			if e != nil || len(tc.id) != MaxIDBytes || len(compiled.InputIndices) != 1 || compiled.InputIndices[0] != 0 {
				t.Fatalf("exact byte bound did not bind original index: %v", e)
			}
			if _, e = Assign(owned, syntheticSeed, p); e != nil {
				t.Fatalf("valid exact512 ID refused: %v", e)
			}
		})
	}
}
