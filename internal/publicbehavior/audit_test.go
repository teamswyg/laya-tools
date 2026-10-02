package publicbehavior

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestNestedObservationCannotInheritTopLevelFields(t *testing.T) {
	if _, err := observationField([]byte(`{"comparison":1,"versions":[{},{}]}`), "versions.0.comparison"); err == nil {
		t.Fatal("top-level field leaked into a version record")
	}
}

func TestRejectsInputsBeforeCallingOriginalAPIs(t *testing.T) {
	for _, in := range []Input{
		{Operation: "arbitrary_code", Left: "1.2.3"},
		{Operation: OperationCompare, Left: strings.Repeat("a", MaxInputBytes+1), Right: "1.2.3"},
		{Operation: OperationStrictParse, Left: "1.2.3", Right: string([]byte{255})},
	} {
		got := Observe(in)
		if got.UpstreamAPICalls != 0 || got.UpstreamGetterCalls != 0 || got.Panicked || got.ComparisonCalled {
			t.Fatal("invalid input reached an upstream API")
		}
		if got.Supported && !got.RejectedInputBounds {
			t.Fatal("missing rejection state")
		}
	}
}

func TestErrorIdentityDoesNotAcceptMatchingMessagesOrWrappedNumError(t *testing.T) {
	if got := classifyError(errors.New("version string empty")); got.kind != ErrorUnclassified {
		t.Fatal("message equality masqueraded as sentinel identity")
	}
	err := &strconv.NumError{Func: "ParseUint", Num: "", Err: strconv.ErrSyntax}
	if got := classifyError(err); got.kind != ErrorStrconvNum || got.cause != ErrorStrconvSyntax {
		t.Fatal("exact NumError identity lost")
	}
	if got := classifyError(fmt.Errorf("outer: %w", err)); got.kind != ErrorUnclassified {
		t.Fatal("unexpected wrapped error type accepted")
	}
}
