package typedbehavior

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
)

func TestStateAuthoredSourcesAndLiteralControls(t *testing.T) {
	original, err := os.ReadFile("state.go")
	if err != nil {
		t.Fatal(err)
	}
	if string(original) != stateSourceText {
		t.Fatal("embedded original state source differs")
	}
	sources := sourcesState()
	if len(sources) != 12 || len(atomicSources) != 4 || len(identitySources) != 4 || len(snapshotSources) != 4 {
		t.Fatal("three families each require one correct and three wrong controls")
	}
	var correct, wrong, checked int
	for _, s := range sources {
		t.Run(s.id, func(t *testing.T) {
			got := checkState(s.id)
			if got.Unknown || got.Checked == 0 || s.correct && got.Failed != 0 || !s.correct && got.Failed == 0 {
				t.Fatalf("compiled control does not match independent table: %+v", got)
			}
			if next := checkState(s.id); next != got {
				t.Fatalf("fresh invocations changed result: %+v -> %+v", got, next)
			}
			checked += got.Checked
			if s.correct {
				correct++
			} else {
				wrong++
			}
		})
	}
	if correct != 3 || wrong != 9 || checked != 140 {
		t.Fatalf("control counts correct=%d wrong=%d vector checks=%d", correct, wrong, checked)
	}
	for _, s := range atomicSources {
		assertStateFunctionBinding(t, s.spec.function, s.run)
	}
	for _, s := range identitySources {
		assertStateFunctionBinding(t, s.spec.function, s.run)
	}
	for _, s := range snapshotSources {
		assertStateFunctionBinding(t, s.spec.function, s.run)
	}
	if got := checkState("not-an-authored-source"); !got.Unknown || got.Checked != 0 || got.Failed != 0 {
		t.Fatal("unsupported source must remain unknown")
	}
	copyOfRegistry := sourcesState()
	copyOfRegistry[0].id = "modified"
	if sourcesState()[0].id != "atomic-correct" {
		t.Fatal("registry returned shared mutable memory")
	}
}

func assertStateFunctionBinding(t *testing.T, function string, fn any) {
	t.Helper()
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	if !strings.HasSuffix(name, "."+function) {
		t.Fatalf("source name / compiled function mismatch: %s / %s", function, name)
	}
}

func TestStateAtomicLateFailureAndExactError(t *testing.T) {
	v := atomicVector{
		Raw: "54,x", Before: atomicConfig{8, true},
		Want: atomicObservation{atomicConfig{0, false}, atomicConfig{8, true}, atomicSyntaxError, false},
	}
	if got := observeAtomic(atomicCorrect, v); got != v.Want {
		t.Fatalf("late error must preserve state and return zero: %+v", got)
	}
	if got := observeAtomic(atomicPartialCommit, v); got.After != (atomicConfig{54, true}) || got.Returned != (atomicConfig{}) || got.Error != atomicSyntaxError {
		t.Fatalf("partial-commit control did not exhibit its distinct failure: %+v", got)
	}
	if got := observeAtomic(atomicPartialResult, v); got.After != v.Before || got.Returned != (atomicConfig{54, false}) || got.Error != atomicSyntaxError {
		t.Fatalf("partial-result control did not exhibit its distinct failure: %+v", got)
	}
	for _, err := range []error{errors.New(ErrSyntax.Error()), fmt.Errorf("wrapped: %w", ErrSyntax), nil} {
		candidate := func(string, *atomicConfig) (atomicConfig, error) { return atomicConfig{}, err }
		if got := checkAtomic(candidate, []atomicVector{v}); got.Unknown || got.Checked != 1 || got.Failed != 1 {
			t.Fatalf("wrong error identity must be a known mismatch: %+v", got)
		}
	}
}

func TestStateCandidatePanicsAreKnownMismatches(t *testing.T) {
	atomic := checkAtomic(func(string, *atomicConfig) (atomicConfig, error) { panic("candidate") }, atomicVectors)
	identity := checkIdentity(func(error, error, error) identityFacts { panic(nil) }, identityVectors)
	snapshot := checkSnapshot(func([]int64) []int64 { panic("candidate") }, snapshotVectors)
	for _, got := range []controlCheck{atomic, identity, snapshot} {
		if got.Unknown || got.Checked == 0 || got.Failed != got.Checked {
			t.Fatalf("candidate panic is a finite no-panic contract violation: %+v", got)
		}
	}
}

func TestStateErrorIdentityAndFreshInputCauses(t *testing.T) {
	joined, ok := buildIdentityFixture(identityJoinedAB)
	if !ok || joined.TargetA == joined.TargetB || joined.TargetA.Error() != joined.TargetB.Error() {
		t.Fatal("equal-text distinct sentinels were not constructed")
	}
	if got := identityCorrect(joined.Input, joined.TargetA, joined.TargetB); got != (identityFacts{true, true, false, 0}) {
		t.Fatalf("joined identities lost: %+v", got)
	}
	first, ok := buildIdentityFixture(identityDirectCause)
	if !ok {
		t.Fatal("missing cause fixture")
	}
	second, ok := buildIdentityFixture(identityDirectCause)
	if !ok || first.Cause == second.Cause || first.TargetA == second.TargetA {
		t.Fatal("each invocation must have fresh cause and target identities")
	}
	got := identityMutateCause(first.Input, first.TargetA, first.TargetB)
	if got != (identityFacts{false, false, true, 7}) || first.Cause.Code != 8 || second.Cause.Code != 7 {
		t.Fatal("cause mutation must be observed independently of otherwise correct output")
	}
	vectors := []identityVector{{identityDirectCause, identityFacts{false, false, true, 7}, 7}}
	if got := checkIdentity(identityMutateCause, vectors); got.Unknown || got.Checked != 1 || got.Failed != 1 {
		t.Fatalf("input cause mutation was not a known mismatch: %+v", got)
	}
}

func TestStateSnapshotOwnershipAndNilness(t *testing.T) {
	if got := checkSnapshot(snapshotCorrect, snapshotVectors); got.Unknown || got.Checked != 6 || got.Failed != 0 {
		t.Fatalf("owned copy failed: %+v", got)
	}
	if got := checkSnapshot(snapshotAlias, snapshotVectors); got.Unknown || got.Checked != 6 || got.Failed != 4 {
		t.Fatalf("each nonempty alias must fail: %+v", got)
	}
	if got := checkSnapshot(snapshotEmptyNil, snapshotVectors); got.Unknown || got.Checked != 6 || got.Failed != 1 {
		t.Fatalf("non-nil empty input must not collapse to nil: %+v", got)
	}
	wrongLength := func([]int64) []int64 { return make([]int64, 100) }
	if got := checkSnapshot(wrongLength, snapshotVectors); got.Unknown || got.Checked != 6 || got.Failed != 6 {
		t.Fatalf("wrong output length is a known mismatch, not observer unknown: %+v", got)
	}
	// Inspector mutations must never alter the literal source table itself.
	before := snapshotVectors[5]
	_ = checkSnapshot(snapshotAlias, snapshotVectors)
	_ = checkSnapshot(snapshotMutates, snapshotVectors)
	if snapshotVectors[5] != before {
		t.Fatal("candidate/inspector writes reached literal expected memory")
	}
	input := []int64{1, 2, 3, 4}
	out := snapshotCorrect(input)
	for i := range out {
		out[i] = int64(70 + i)
	}
	if input[0] != 1 || input[1] != 2 || input[2] != 3 || input[3] != 4 {
		t.Fatal("returned-value mutation changed original input")
	}
	for i := range input {
		input[i] = -int64(30 + i)
	}
	if out[0] != 70 || out[1] != 71 || out[2] != 72 || out[3] != 73 {
		t.Fatal("original-input mutation changed returned value")
	}
}

func TestStateUnsupportedObserverInputsStayUnknown(t *testing.T) {
	for _, got := range []controlCheck{
		checkAtomic(nil, atomicVectors), checkAtomic(atomicCorrect, nil),
		checkIdentity(nil, identityVectors), checkIdentity(identityCorrect, nil),
		checkIdentity(identityCorrect, []identityVector{{Case: identityCase(255)}}),
		checkSnapshot(nil, snapshotVectors), checkSnapshot(snapshotCorrect, nil),
		checkSnapshot(snapshotCorrect, []snapshotVector{{Input: sliceState{Len: 5}}}),
		checkSnapshot(snapshotCorrect, []snapshotVector{{Input: sliceState{Len: 1, Nil: true}}}),
		checkSnapshot(snapshotCorrect, []snapshotVector{{Input: sliceState{Values: [4]int64{0, 1, 0, 0}}}}),
	} {
		if !got.Unknown || got.Checked != 0 || got.Failed != 0 {
			t.Fatalf("unsupported observer input gained a failure label: %+v", got)
		}
	}
	if _, supported := sliceStateOf(make([]int64, 5)); supported {
		t.Fatal("fixed observer accepted an unsupported extent")
	}
}

func TestStateFiniteContractEvidence(t *testing.T) {
	evidence := stateContractEvidence()
	if len(evidence) != 3 {
		t.Fatal("finite contract count")
	}
	want := []struct {
		prototype string
		core      string
		vectors   int
		sha       string
	}{
		{"atomic-commit", "staged-commit", 18, "1db8a7b5c077fd02f734b5c9e89739cd1c83d718c71c3cf8b4a1395d75b8075d"},
		{"error-identity", "error-cause-identity", 11, "a3f82cc54e5ed58840fff8938107ce5c5636f96eb268141c8d96ca191fa0c230"},
		{"owned-snapshot", "ownership-graph", 6, "cf0d2577491044ead9f63bc2f0530fb2396ca6658c0d4715b9bb107cf6cf24e4"},
	}
	for i, e := range evidence {
		if e.Prototype != want[i].prototype || e.Core != want[i].core || e.Vectors != want[i].vectors || e.InputSemantics == "" || e.TruthTableSHA256 != want[i].sha {
			t.Fatalf("incomplete contract evidence: %+v", e)
		}
	}
}

// These complete examples are proposals for the separately authored fixture.
// They are not inputs to source truth, controls, or a model in these tests.
func TestStateLanguageExamplesFitBudget(t *testing.T) {
	samples := []struct {
		id, text string
	}{
		{"atomic-request", "Nonnil destination: parse exact ASCII DD,F: Count=DD; Enabled=(F=1); DD two digits; F=0/1. Success commits/returns Config. Failure returns zero/exact ErrSyntax, preserving destination. Never panic."},
		{"atomic-correct", "Validate exact DD,F syntax before writing. Success returns and commits parsed Config; failure returns zero and ErrSyntax without modifying destination."},
		{"atomic-partial-commit", "Update destination Count from a leading ASCII digit pair before validating DD,F. Then commit full valid Config or return zero and ErrSyntax."},
		{"atomic-partial-result", "Validate DD,F. On failure, preserve destination; return Count from the first two ASCII digits, or zero if invalid, plus ErrSyntax. Success commits and returns Config."},
		{"atomic-no-commit", "Validate DD,F; return parsed Config on success without writing destination. Invalid input returns zero and ErrSyntax, leaving destination unchanged."},
		{"identity-request", "Report errors.Is(A/B) and errors.As(Cause)/original code through wrapping/joining. Targets and Cause pointers are nonnil. Equal messages alone never match. Preserve input causes; never panic."},
		{"identity-correct", "Use errors.Is for both targets and errors.As for Cause. Return its original code without changing the error chain."},
		{"identity-by-message", "Match targets by finding their messages in the input error text. Use errors.As for Cause and return its code unchanged."},
		{"identity-direct-only", "Compare input directly with each target and use a direct Cause type assertion. Return the direct cause code without modifying it."},
		{"identity-mutate-cause", "Use errors.Is and errors.As; capture the original Cause code, increment that Cause's stored code, and return the captured facts."},
		{"snapshot-request", "Return an independent slice with identical length, values, and nilness. Mutating either input or result must not alter the other. Leave input unchanged during copying; never panic."},
		{"snapshot-correct", "Return nil for nil input; otherwise allocate and copy every value, preserving non-nil empty slices and leaving input unchanged."},
		{"snapshot-alias", "Return the input slice directly with the same nilness, length, and values, so input and result share their backing array."},
		{"snapshot-mutates", "For nonempty input, increment its first value, then allocate and copy all values. Preserve nilness and non-nil empty slices."},
		{"snapshot-empty-nil", "Return nil whenever length is zero; otherwise allocate and copy every value without changing input."},
	}
	for _, sample := range samples {
		normalized := lexicalhint.NormalizeText(sample.text)
		words := len(strings.Fields(normalized))
		if len(sample.text) > 512 || len(normalized) > 512 || words > 32 {
			t.Fatalf("%s: bytes=%d normalized bytes=%d words=%d", sample.id, len(sample.text), len(normalized), words)
		}
	}
}
