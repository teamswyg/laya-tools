// SPDX-License-Identifier: Apache-2.0
package hintrelation

import (
	"errors"
	"strings"
	"testing"
)

// These are original synthetic grammar probes, not native API fixtures,
// supervision, saved evaluation cases, or an assertion of interval correctness.
func syntheticTraversal(method, interval string) string {
	return method + " calls the iterator for every value in the tree within the range " + interval + ", until iterator returns false."
}

func TestRequestRelations(t *testing.T) {
	const filterCoverage = HasDirection | HasLower | HasUpper
	tests := []struct {
		name, text string
		want       Attributes
	}{
		{"closed_open", "Return keys in increasing order keeping x >= lower and x < upper.", Attributes{Ascending, Closed, Open, false, filterCoverage}},
		{"open_closed", "Return keys in decreasing order with x > lower and x <= upper", Attributes{Descending, Open, Closed, false, filterCoverage}},
		{"both_open", "Return keys in decreasing order keeping lower < x and upper > x", Attributes{Descending, Open, Open, false, filterCoverage}},
		{"both_closed", "Return keys in increasing order with upper >= x and lower <= x", Attributes{Ascending, Closed, Closed, false, filterCoverage}},
		{"predicate_order", "Return keys in increasing order, with x < upper and x >= lower", Attributes{Ascending, Closed, Open, false, filterCoverage}},
		{"reverse_operands", "Return keys in decreasing order with upper >= x and lower < x.", Attributes{Descending, Open, Closed, false, filterCoverage}},
		{"reverse_closed_open", "Return keys in increasing order keeping lower <= x and upper > x", Attributes{Ascending, Closed, Open, false, filterCoverage}},
		{"lower_open_only", "Return keys in increasing order with lower < x", Attributes{Ascending, Open, Unbounded, false, filterCoverage}},
		{"lower_closed_only", "Return keys in decreasing order keeping x >= lower", Attributes{Descending, Closed, Unbounded, false, filterCoverage}},
		{"upper_open_only", "Return keys in decreasing order with upper > x", Attributes{Descending, Unbounded, Open, false, filterCoverage}},
		{"upper_closed_only", "Return keys in increasing order keeping x <= upper", Attributes{Ascending, Unbounded, Closed, false, filterCoverage}},
		{"all_increasing", "Return all keys in increasing order.", Attributes{Ascending, Unbounded, Unbounded, false, filterCoverage}},
		{"all_decreasing", "Return all keys in decreasing order", Attributes{Descending, Unbounded, Unbounded, false, filterCoverage}},
		{"first_two", "Return the first two qualifying keys in increasing order with upper > x and lower <= x; stop the callback after two.", Attributes{Ascending, Closed, Open, true, filterCoverage | HasStop}},
		{"first_two_one_sided", "Return the first two qualifying keys in decreasing order keeping x > lower; stop the callback after two", Attributes{Descending, Open, Unbounded, true, filterCoverage | HasStop}},
		{"case_whitespace", "RETURN\tKEYS\nIN DECREASING ORDER, KEEPING UPPER >= X AND LOWER < X.\r\n", Attributes{Descending, Open, Closed, false, filterCoverage}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequest(tt.text)
			if err != nil || got != tt.want {
				t.Fatalf("ParseRequest = %+v, %v; want %+v, nil", got, err, tt.want)
			}
		})
	}
}

func TestKeepingAndWithAreEquivalent(t *testing.T) {
	tails := []string{
		"x > lower", "x >= lower", "upper > x", "upper >= x",
		"x >= lower and x < upper", "upper >= x and lower < x",
	}
	for _, direction := range []string{"increasing", "decreasing"} {
		for _, tail := range tails {
			for _, prefix := range []string{"Return keys", "Return the first two qualifying keys"} {
				base := prefix + " in " + direction + " order, "
				if strings.Contains(prefix, "first two") {
					tail += "; stop the callback after two"
				}
				keeping, err := ParseRequest(base + "keeping " + tail + ".")
				if err != nil || keeping.Coverage == 0 {
					t.Fatalf("supported keeping probe = %+v, %v", keeping, err)
				}
				with, err := ParseRequest(base + "with " + tail + ".")
				if err != nil || with != keeping {
					t.Fatalf("with = %+v, %v; keeping = %+v", with, err, keeping)
				}
			}
		}
	}
}

func TestUnsupportedRequestsDoNotGuess(t *testing.T) {
	tests := []string{
		"", " ", "Return keys in increasing order", "Return keys in increasing order keeping",
		"Return keys in increasing order keeping x = lower", "Return keys in increasing order keeping x < lower",
		"Return keys in increasing order keeping x >= upper", "Return keys in increasing order keeping lower > x",
		"Return keys in increasing order keeping lower < upper", "Return keys in increasing order keeping x != lower",
		"Return keys in increasing order keeping x >= lower or x < upper",
		"Return keys in increasing order keeping x >= lower and x > lower",
		"Return keys in increasing order keeping x >= lower and x < upper and x <= upper",
		"Return keys in increasing order keeping not x >= lower",
		"Do not return keys in increasing order keeping x >= lower",
		"\"Return keys in increasing order keeping x >= lower\"",
		"Return keys in increasing order keeping 'x' >= lower",
		"Return keys in increasing order keeping x >= lower unless empty",
		"If ready, return keys in increasing order keeping x >= lower",
		"Return keys in increasing order keeping x >= lower, but x < lower",
		"Return keys in increasing order keeping x >= lower. Otherwise use decreasing order.",
		"Return keys in increasing order keeping x >= lower..",
		"Return all keys in increasing order keeping x >= lower",
		"Return all keys in increasing order; stop the callback after two",
		"Return the first two qualifying keys in increasing order keeping x >= lower",
		"Return the first two qualifying keys in increasing order keeping x >= lower; stop the callback after three",
		"Return keys in increasing order keeping x >= lower; stop the callback after two",
		"Return keys in increasing order on an empty tree",
		"Return keys in increasing order keeping x >= lower 그리고 x < upper",
		"Return keys in increasing order keeping x >= lower — inclusive",
	}
	for _, raw := range tests {
		got, err := ParseRequest(raw)
		if err != nil || got != (Attributes{}) {
			t.Errorf("unsupported %q = %+v, %v; want zero recognition and nil", raw, got, err)
		}
	}
}

func TestCandidateIntervalPositions(t *testing.T) {
	const full = HasDirection | HasLower | HasUpper | HasStop
	const partial = HasDirection | HasStop
	tests := []struct {
		name, method, interval string
		want                   Attributes
	}{
		{"ascending_closed_open", "AscendRange", "[pivot, pivot)", Attributes{Ascending, Closed, Open, true, full}},
		{"ascending_open_closed", "AscendRange", "(pivot, pivot]", Attributes{Ascending, Open, Closed, true, full}},
		{"descending_closed_open_positions", "DescendRange", "[pivot, pivot)", Attributes{Descending, Open, Closed, true, full}},
		{"descending_open_closed_positions", "DescendRange", "(pivot, pivot]", Attributes{Descending, Closed, Open, true, full}},
		{"ascending_named_bounds", "AscendRange", "[greaterOrEqual, lessThan)", Attributes{Ascending, Closed, Open, true, full}},
		{"descending_named_bounds", "DescendRange", "[lessOrEqual, greaterThan)", Attributes{Descending, Open, Closed, true, full}},
		{"suffix_does_not_supply_bounds", "AscendGreaterThan", "[pivot, pivot)", Attributes{Ascending, Closed, Open, true, full}},
		{"less_than_suffix", "DescendLessThan", "(pivot, pivot]", Attributes{Descending, Closed, Open, true, full}},
		{"less_or_equal_suffix", "AscendLessOrEqual", "[pivot, pivot]", Attributes{Ascending, Closed, Closed, true, full}},
		{"greater_or_equal_suffix", "DescendGreaterOrEqual", "(pivot, pivot)", Attributes{Descending, Open, Open, true, full}},
		{"bare_ascending", "Ascend", "[first, last]", Attributes{Ascending, Unbounded, Unbounded, true, full}},
		{"bare_descending", "Descend", "[last, first]", Attributes{Descending, Unbounded, Unbounded, true, full}},
		{"misoriented_extrema_ascending", "AscendRange", "[last, first]", Attributes{Ascending, BoundUnknown, BoundUnknown, true, partial}},
		{"misoriented_extrema_descending", "DescendRange", "[first, last]", Attributes{Descending, BoundUnknown, BoundUnknown, true, partial}},
		{"excluded_first", "AscendRange", "(first, last]", Attributes{Ascending, BoundUnknown, Unbounded, true, partial | HasUpper}},
		{"excluded_last", "AscendRange", "[first, last)", Attributes{Ascending, Unbounded, BoundUnknown, true, partial | HasLower}},
		{"excluded_descending_last", "DescendRange", "(last, first]", Attributes{Descending, Unbounded, BoundUnknown, true, partial | HasLower}},
		{"excluded_descending_first", "DescendRange", "[last, first)", Attributes{Descending, BoundUnknown, Unbounded, true, partial | HasUpper}},
		{"one_unknown_endpoint", "AscendRange", "[unrecognized, pivot)", Attributes{Ascending, BoundUnknown, Open, true, partial | HasUpper}},
		{"unknown_right_endpoint", "DescendRange", "[pivot, unrecognized)", Attributes{Descending, BoundUnknown, Closed, true, partial | HasUpper}},
		{"closed_less_than_is_unknown", "DescendRange", "[lessThan, greaterThan)", Attributes{Descending, Open, BoundUnknown, true, partial | HasLower}},
		{"opposite_named_positions", "DescendRange", "[greaterOrEqual, lessThan)", Attributes{Descending, BoundUnknown, BoundUnknown, true, partial}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCandidate(syntheticTraversal(tt.method, tt.interval))
			if err != nil || got != tt.want {
				t.Fatalf("ParseCandidate = %+v, %v; want %+v, nil", got, err, tt.want)
			}
		})
	}
}

func TestCandidateRequiresWholeTraversalScope(t *testing.T) {
	valid := syntheticTraversal("AscendRange", "[pivot, pivot)")
	tests := []string{
		"", "AscendRange", "AscendRange [pivot, pivot)",
		strings.Replace(valid, "every value", "some value", 1),
		strings.Replace(valid, "within the range", "outside the range", 1),
		strings.Replace(valid, ", until iterator returns false.", ".", 1),
		strings.Replace(valid, "returns false", "returns true", 1),
		strings.Replace(valid, "AscendRange", "AscendEqual", 1),
		strings.Replace(valid, "AscendRange", "TraverseRange", 1),
		strings.Replace(valid, "[pivot, pivot)", "{pivot, pivot}", 1),
		strings.Replace(valid, "[pivot, pivot)", "[pivot, pivot, pivot)", 1),
		strings.Replace(valid, "[pivot, pivot)", "[, pivot)", 1),
		strings.Replace(valid, "[pivot, pivot)", "[pivot, )", 1),
		strings.Replace(valid, "[pivot, pivot)", "[pivot, (pivot))", 1),
		"\"" + valid + "\"", "If enabled, " + valid,
		valid + " But the iterator cannot stop.", valid + " 그리고 계속합니다",
	}
	for _, raw := range tests {
		got, err := ParseCandidate(raw)
		if err != nil || got != (Attributes{}) {
			t.Errorf("unsupported candidate %q = %+v, %v; want zero recognition and nil", raw, got, err)
		}
	}
}

func TestExtractUsesOnlyJointRecognition(t *testing.T) {
	request := "Return the first two qualifying keys in increasing order with x >= lower and x < upper; stop the callback after two."
	candidate := syntheticTraversal("AscendRange", "(pivot, pivot]")
	got, q, d, err := Extract(request, candidate)
	want := [Dimension]float32{1, 0, 0, 1, 0, 1, 1, 0}
	if err != nil || got != want || q.Coverage == 0 || d.Coverage == 0 {
		t.Fatalf("Extract = %v, %+v, %+v, %v; want %v", got, q, d, err, want)
	}

	partialCandidate := syntheticTraversal("AscendRange", "[unknown, pivot)")
	got, _, d, err = Extract(request, partialCandidate)
	want = [Dimension]float32{1, 0, 0, 0, 1, 0, 1, 0}
	if err != nil || got != want || d.Coverage&HasLower != 0 {
		t.Fatalf("partial endpoint generated a guessed lower feature: %v, %+v, %v", got, d, err)
	}

	got, q, d, err = Extract("unsupported request", candidate)
	if err != nil || got != ([Dimension]float32{}) || q != (Attributes{}) || d.Coverage == 0 {
		t.Fatalf("unsupported request leaked candidate recognition into features: %v, %+v, %+v, %v", got, q, d, err)
	}
	got, q, d, err = Extract(request, "unsupported candidate")
	if err != nil || got != ([Dimension]float32{}) || q.Coverage == 0 || d != (Attributes{}) {
		t.Fatalf("unsupported candidate leaked request recognition into features: %v, %+v, %+v, %v", got, q, d, err)
	}

	got, _, _, err = Extract("Return keys in increasing order keeping x >= lower and x < upper", candidate)
	if err != nil || got[StopMatch] != 0 || got[StopConflict] != 0 {
		t.Fatalf("omitted stop invented a negative stop constraint: %v, %v", got, err)
	}
	got, q, d, err = Extract("", "")
	if err != nil || got != ([Dimension]float32{}) || q != (Attributes{}) || d != (Attributes{}) {
		t.Fatalf("empty text = %v, %+v, %+v, %v", got, q, d, err)
	}
}

func TestColumnsOwnSoAAndOnlyLiveSlots(t *testing.T) {
	request := "Return keys in increasing order with x >= lower and x < upper"
	candidates := [MaxCandidates]string{
		syntheticTraversal("AscendRange", "[pivot, pivot)"),
		syntheticTraversal("DescendRange", "[pivot, pivot)"),
		"unsupported candidate",
	}
	// A bounds-invalid unused string must not turn into a live input or error.
	candidates[MaxCandidates-1] = strings.Repeat("x", MaxBytes+1)
	columns, err := Build(request, candidates, 3)
	if err != nil || columns.Count() != 3 {
		t.Fatalf("Build count = %d, %v; want 3, nil", columns.Count(), err)
	}
	query, err := columns.Attributes(-1)
	if err != nil || query != (Attributes{Ascending, Closed, Open, false, HasDirection | HasLower | HasUpper}) {
		t.Fatalf("query slot = %+v, %v", query, err)
	}
	values := columns.Values()
	for i := 0; i < 3; i++ {
		want, _, wantAttributes, err := Extract(request, candidates[i])
		if err != nil {
			t.Fatal(err)
		}
		gotAttributes, err := columns.Attributes(i)
		if err != nil || gotAttributes != wantAttributes {
			t.Fatalf("candidate slot %d = %+v, %v; want %+v", i, gotAttributes, err, wantAttributes)
		}
		for column := range want {
			if values[column][i] != want[column] {
				t.Fatalf("SoA[%d][%d] = %v; want %v", column, i, values[column][i], want[column])
			}
		}
	}
	for i := 3; i < MaxCandidates; i++ {
		for column := range values {
			if values[column][i] != 0 {
				t.Fatalf("inactive feature [%d][%d] = %v", column, i, values[column][i])
			}
		}
		if columns.direction[i+1] != DirectionUnknown || columns.lower[i+1] != BoundUnknown || columns.upper[i+1] != BoundUnknown || columns.stop[i+1] || columns.coverage[i+1] != 0 {
			t.Fatalf("inactive attribute slot %d is not zero", i+1)
		}
	}
	owned := columns
	values[DirectionMatch][0] = 123
	candidates[0] = "changed source array"
	query.Direction = Descending
	if columns != owned || columns.Values()[DirectionMatch][0] != 1 {
		t.Fatal("a returned value or source-array mutation changed owned columns")
	}
	other, err := Build("unsupported query", candidates, 1)
	if err != nil || other.Values() != ([Dimension][MaxCandidates]float32{}) || columns != owned {
		t.Fatalf("another batch changed owned columns: %+v, %v", other, err)
	}
	for _, index := range []int{-2, 3, MaxCandidates} {
		got, err := columns.Attributes(index)
		if !errors.Is(err, ErrInput) || got != (Attributes{}) {
			t.Errorf("invalid attribute index %d = %+v, %v", index, got, err)
		}
	}
}

func TestAgreementOrderStableTiesAndPermutation(t *testing.T) {
	request := "Return the first two qualifying keys in increasing order with x >= lower and x < upper; stop the callback after two"
	candidates := [MaxCandidates]string{
		syntheticTraversal("AscendRange", "[pivot, pivot)"),
		syntheticTraversal("DescendRange", "[pivot, pivot)"),
		"unsupported candidate A",
		syntheticTraversal("Ascend", "[first, last]"),
		syntheticTraversal("AscendRange", "[pivot, pivot)"),
		syntheticTraversal("AscendRange", "(pivot, pivot]"),
		syntheticTraversal("DescendRange", "(pivot, pivot]"),
		"unsupported candidate B",
	}
	columns, err := Build(request, candidates, MaxCandidates)
	if err != nil {
		t.Fatal(err)
	}
	order, scores, err := columns.AgreementOrder()
	wantOrder := [MaxCandidates]int{0, 4, 6, 2, 3, 5, 7, 1}
	wantScores := [MaxCandidates]int{4, -2, 0, 0, 4, 0, 2, 0}
	if err != nil || order != wantOrder || scores != wantScores {
		t.Fatalf("order = %v, scores = %v, err = %v; want %v, %v", order, scores, err, wantOrder, wantScores)
	}
	var seen [MaxCandidates]bool
	for rank, candidate := range order {
		if candidate < 0 || candidate >= MaxCandidates || seen[candidate] {
			t.Fatalf("not a live permutation at rank %d: %v", rank, order)
		}
		seen[candidate] = true
		if rank > 0 && scores[order[rank-1]] < scores[candidate] {
			t.Fatal("scores increased with rank")
		}
	}
	// Unsupported syntax is a zero-feature stable tie, not removal of candidates.
	columns, err = Build("unsupported query", candidates, 3)
	if err != nil {
		t.Fatal(err)
	}
	order, scores, err = columns.AgreementOrder()
	if err != nil || order != ([MaxCandidates]int{0, 1, 2, -1, -1, -1, -1, -1}) || scores != ([MaxCandidates]int{}) {
		t.Fatalf("zero-feature order = %v, scores = %v, %v", order, scores, err)
	}
}

func TestInputBoundsAndAtomicZeroResults(t *testing.T) {
	// Boundary probes use unsupported text deliberately: budget acceptance must
	// be distinct from grammar recognition and never imply supervision.
	accepted := []string{
		strings.Repeat("x", MaxBytes),
		strings.TrimSpace(strings.Repeat("word ", MaxWords)),
		strings.Repeat(";", MaxTokens),
		strings.Repeat("Ax", MaxTokens), // one ASCII word, 64 camel-case tokens
	}
	for _, raw := range accepted {
		for _, parse := range []func(string) (Attributes, error){ParseRequest, ParseCandidate} {
			got, err := parse(raw)
			if err != nil || got != (Attributes{}) {
				t.Errorf("accepted boundary (%d bytes) = %+v, %v", len(raw), got, err)
			}
		}
	}
	invalid := []string{
		string([]byte{0xff}), string([]byte{'a', 0xc3}),
		strings.Repeat("x", MaxBytes+1),
		strings.TrimSpace(strings.Repeat("word ", MaxWords+1)),
		strings.Repeat(";", MaxTokens+1),
		strings.Repeat("Ax", MaxTokens+1),
	}
	validRequest := "Return all keys in increasing order"
	validCandidate := syntheticTraversal("Ascend", "[first, last]")
	for _, raw := range invalid {
		for _, parse := range []func(string) (Attributes, error){ParseRequest, ParseCandidate} {
			got, err := parse(raw)
			if !errors.Is(err, ErrInput) || got != (Attributes{}) {
				t.Errorf("invalid boundary (%d bytes) = %+v, %v", len(raw), got, err)
			}
		}
		for _, inputs := range [][2]string{{raw, validCandidate}, {validRequest, raw}} {
			v, q, d, err := Extract(inputs[0], inputs[1])
			if !errors.Is(err, ErrInput) || v != ([Dimension]float32{}) || q != (Attributes{}) || d != (Attributes{}) {
				t.Errorf("Extract left partial state on input failure: %v, %+v, %+v, %v", v, q, d, err)
			}
		}
		candidates := [MaxCandidates]string{validCandidate, raw}
		columns, err := Build(validRequest, candidates, 2)
		if !errors.Is(err, ErrInput) || columns != (Columns{}) {
			t.Errorf("Build left first candidate on later failure: %+v, %v", columns, err)
		}
		columns, err = Build(raw, [MaxCandidates]string{validCandidate}, 1)
		if !errors.Is(err, ErrInput) || columns != (Columns{}) {
			t.Errorf("Build left state on request failure: %+v, %v", columns, err)
		}
	}
	for _, count := range []int{-1, 0, MaxCandidates + 1} {
		columns, err := Build(validRequest, [MaxCandidates]string{validCandidate}, count)
		if !errors.Is(err, ErrInput) || columns != (Columns{}) {
			t.Errorf("invalid count %d = %+v, %v", count, columns, err)
		}
	}
	var zero Columns
	if zero.Count() != 0 || zero.Values() != ([Dimension][MaxCandidates]float32{}) {
		t.Fatal("zero Columns exposes live storage")
	}
	for _, index := range []int{-1, 0} {
		got, err := zero.Attributes(index)
		if !errors.Is(err, ErrInput) || got != (Attributes{}) {
			t.Errorf("zero Attributes(%d) = %+v, %v", index, got, err)
		}
	}
	order, scores, err := zero.AgreementOrder()
	if !errors.Is(err, ErrInput) || order != ([MaxCandidates]int{-1, -1, -1, -1, -1, -1, -1, -1}) || scores != ([MaxCandidates]int{}) {
		t.Fatalf("zero order = %v, scores = %v, %v", order, scores, err)
	}
}
