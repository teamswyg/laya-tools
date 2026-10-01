package typedbehavior

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// This original source contains concrete candidate implementations, observer
// code, and independently authored literal vectors. It is development evidence,
// never a text-scoring feature or a protected-final sample.
//
//go:embed state.go
var stateSourceText string

type atomicConfig struct {
	Count   uint8
	Enabled bool
}

type atomicFn func(string, *atomicConfig) (atomicConfig, error)

// ErrSyntax is the exact error required by the finite atomic-commit contract.
// An error with equal text, or a wrapping of this sentinel, is not exact.
var ErrSyntax = errors.New("invalid DD,F syntax")

type atomicError uint8

const (
	atomicNoError atomicError = iota
	atomicSyntaxError
	atomicOtherError
)

type atomicObservation struct {
	Returned atomicConfig
	After    atomicConfig
	Error    atomicError
	Panicked bool
}

type atomicVector struct {
	Raw    string
	Before atomicConfig
	Want   atomicObservation
}

// All expected states below are literal, not calculated by a parser/candidate.
// The grammar has exactly four ASCII bytes: two digits, comma, flag 0 or 1.
var atomicVectors = []atomicVector{
	{"00,0", atomicConfig{87, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{0, false}, atomicNoError, false}},
	{"01,1", atomicConfig{77, false}, atomicObservation{atomicConfig{1, true}, atomicConfig{1, true}, atomicNoError, false}},
	{"12,1", atomicConfig{77, false}, atomicObservation{atomicConfig{12, true}, atomicConfig{12, true}, atomicNoError, false}},
	{"99,0", atomicConfig{3, true}, atomicObservation{atomicConfig{99, false}, atomicConfig{99, false}, atomicNoError, false}},
	{"", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12,2", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12,x", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12;1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"123,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12,1x", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"1a,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"/2,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{":2,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"１２,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12,１", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{"12,1\n", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
	{" 12,1", atomicConfig{77, true}, atomicObservation{atomicConfig{0, false}, atomicConfig{77, true}, atomicSyntaxError, false}},
}

func decodeAtomic(raw string) (atomicConfig, bool) {
	if len(raw) != 4 || raw[0] < '0' || raw[0] > '9' || raw[1] < '0' || raw[1] > '9' || raw[2] != ',' || raw[3] != '0' && raw[3] != '1' {
		return atomicConfig{}, false
	}
	return atomicConfig{(raw[0]-'0')*10 + raw[1] - '0', raw[3] == '1'}, true
}

func atomicCorrect(raw string, dst *atomicConfig) (atomicConfig, error) {
	next, ok := decodeAtomic(raw)
	if !ok {
		return atomicConfig{}, ErrSyntax
	}
	*dst = next
	return next, nil
}

func atomicPartialCommit(raw string, dst *atomicConfig) (atomicConfig, error) {
	if len(raw) >= 2 && raw[0] >= '0' && raw[0] <= '9' && raw[1] >= '0' && raw[1] <= '9' {
		dst.Count = (raw[0]-'0')*10 + raw[1] - '0'
	}
	next, ok := decodeAtomic(raw)
	if !ok {
		return atomicConfig{}, ErrSyntax
	}
	*dst = next
	return next, nil
}

func atomicPartialResult(raw string, dst *atomicConfig) (atomicConfig, error) {
	next, ok := decodeAtomic(raw)
	if !ok {
		partial := atomicConfig{}
		if len(raw) >= 2 && raw[0] >= '0' && raw[0] <= '9' && raw[1] >= '0' && raw[1] <= '9' {
			partial.Count = (raw[0]-'0')*10 + raw[1] - '0'
		}
		return partial, ErrSyntax
	}
	*dst = next
	return next, nil
}

func atomicNoCommit(raw string, dst *atomicConfig) (atomicConfig, error) {
	next, ok := decodeAtomic(raw)
	if !ok {
		return atomicConfig{}, ErrSyntax
	}
	return next, nil
}

// Only the candidate call is recovered. Observer faults are not relabeled as
// candidate failures. Each invocation receives its own destination state.
func invokeAtomic(fn atomicFn, raw string, dst *atomicConfig) (out atomicConfig, err error, panicked bool) {
	panicked = true
	defer func() { _ = recover() }()
	out, err = fn(raw, dst)
	panicked = false
	return
}

func observeAtomic(fn atomicFn, v atomicVector) atomicObservation {
	dst := v.Before
	out, err, panicked := invokeAtomic(fn, v.Raw, &dst)
	e := atomicOtherError
	if err == nil {
		e = atomicNoError
	} else if err == ErrSyntax {
		e = atomicSyntaxError
	}
	return atomicObservation{out, dst, e, panicked}
}

func checkAtomic(fn atomicFn, vectors []atomicVector) controlCheck {
	if fn == nil || len(vectors) == 0 {
		return controlCheck{Unknown: true}
	}
	var result controlCheck
	for _, v := range vectors {
		got := observeAtomic(fn, v)
		result.Checked++
		if got != v.Want {
			result.Failed++
		}
	}
	return result
}

type identityCause struct {
	Code uint8
}

func (c *identityCause) Error() string { return "typed cause" }

type identityFacts struct {
	HasA, HasB, HasCause bool
	CauseCode            uint8
}

type identityFn func(input, targetA, targetB error) identityFacts

type identityCase uint8

const (
	identityNil identityCase = iota
	identityDirectA
	identityDirectB
	identityWrappedA
	identityJoinedAB
	identitySameMessage
	identityDirectCause
	identityWrappedCause
	identityJoinedACause
	identityJoinedUnrelatedCause
	identityUnrelated
)

type identityVector struct {
	Case           identityCase
	Want           identityFacts
	WantCauseAfter uint8
}

// Expected facts and cause states are literals independent of errors.Is/As and
// all candidate implementations. No fixture contains an ambiguous typed nil.
var identityVectors = []identityVector{
	{identityNil, identityFacts{false, false, false, 0}, 0},
	{identityDirectA, identityFacts{true, false, false, 0}, 0},
	{identityDirectB, identityFacts{false, true, false, 0}, 0},
	{identityWrappedA, identityFacts{true, false, false, 0}, 0},
	{identityJoinedAB, identityFacts{true, true, false, 0}, 0},
	{identitySameMessage, identityFacts{false, false, false, 0}, 0},
	{identityDirectCause, identityFacts{false, false, true, 7}, 7},
	{identityWrappedCause, identityFacts{false, false, true, 9}, 9},
	{identityJoinedACause, identityFacts{true, false, true, 7}, 7},
	{identityJoinedUnrelatedCause, identityFacts{false, false, true, 9}, 9},
	{identityUnrelated, identityFacts{false, false, false, 0}, 0},
}

type identityFixture struct {
	Input, TargetA, TargetB error
	Cause                   *identityCause
}

func buildIdentityFixture(kind identityCase) (identityFixture, bool) {
	// Equal text is deliberate: equality of messages must not imply identity.
	a, b := errors.New("same-text"), errors.New("same-text")
	f := identityFixture{TargetA: a, TargetB: b}
	switch kind {
	case identityNil:
	case identityDirectA:
		f.Input = a
	case identityDirectB:
		f.Input = b
	case identityWrappedA:
		f.Input = fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", a))
	case identityJoinedAB:
		f.Input = errors.Join(a, b)
	case identitySameMessage:
		f.Input = errors.New("same-text")
	case identityDirectCause:
		f.Cause = &identityCause{7}
		f.Input = f.Cause
	case identityWrappedCause:
		f.Cause = &identityCause{9}
		f.Input = fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", f.Cause))
	case identityJoinedACause:
		f.Cause = &identityCause{7}
		f.Input = errors.Join(a, f.Cause)
	case identityJoinedUnrelatedCause:
		f.Cause = &identityCause{9}
		f.Input = errors.Join(errors.New("unrelated"), fmt.Errorf("inner: %w", f.Cause))
	case identityUnrelated:
		f.Input = errors.New("unrelated")
	default:
		return identityFixture{}, false
	}
	return f, true
}

func identityCorrect(input, a, b error) identityFacts {
	facts := identityFacts{HasA: errors.Is(input, a), HasB: errors.Is(input, b)}
	var cause *identityCause
	facts.HasCause = errors.As(input, &cause)
	if facts.HasCause {
		facts.CauseCode = cause.Code
	}
	return facts
}

func identityByMessage(input, a, b error) identityFacts {
	var facts identityFacts
	if input != nil {
		facts.HasA = strings.Contains(input.Error(), a.Error())
		facts.HasB = strings.Contains(input.Error(), b.Error())
	}
	var cause *identityCause
	facts.HasCause = errors.As(input, &cause)
	if facts.HasCause {
		facts.CauseCode = cause.Code
	}
	return facts
}

func identityDirectOnly(input, a, b error) identityFacts {
	facts := identityFacts{HasA: input == a, HasB: input == b}
	cause, ok := input.(*identityCause)
	facts.HasCause = ok
	if ok {
		facts.CauseCode = cause.Code
	}
	return facts
}

func identityMutateCause(input, a, b error) identityFacts {
	facts := identityFacts{HasA: errors.Is(input, a), HasB: errors.Is(input, b)}
	var cause *identityCause
	facts.HasCause = errors.As(input, &cause)
	if facts.HasCause {
		facts.CauseCode = cause.Code
		cause.Code++
	}
	return facts
}

func invokeIdentity(fn identityFn, f identityFixture) (out identityFacts, panicked bool) {
	panicked = true
	defer func() { _ = recover() }()
	out = fn(f.Input, f.TargetA, f.TargetB)
	panicked = false
	return
}

func checkIdentity(fn identityFn, vectors []identityVector) controlCheck {
	if fn == nil || len(vectors) == 0 {
		return controlCheck{Unknown: true}
	}
	var result controlCheck
	for _, v := range vectors {
		f, ok := buildIdentityFixture(v.Case)
		if !ok {
			result.Unknown = true
			return result
		}
		got, panicked := invokeIdentity(fn, f)
		result.Checked++
		changed := f.Cause != nil && f.Cause.Code != v.WantCauseAfter
		if panicked || got != v.Want || changed {
			result.Failed++
		}
	}
	return result
}

type snapshotFn func([]int64) []int64

// A fixed observation value distinguishes nil from non-nil empty slices without
// retaining candidate slices or using reflection. Values outside Len are zero.
type sliceState struct {
	Values [4]int64
	Len    uint8
	Nil    bool
}

type snapshotTrace struct {
	Returned              sliceState
	InputAfterCall        sliceState
	InputAfterResultWrite sliceState
	ResultAfterInputWrite sliceState
}

type snapshotVector struct {
	Input sliceState
	Want  snapshotTrace
}

// Mutations in this literal trace are inspector actions, not candidate writes:
// result[i]=73+i, then input[i]=-29-i. Every logical index is checked in both
// directions; the expected post-mutation result values remain literal below.
var snapshotVectors = []snapshotVector{
	{sliceState{[4]int64{}, 0, true}, snapshotTrace{
		sliceState{[4]int64{}, 0, true}, sliceState{[4]int64{}, 0, true},
		sliceState{[4]int64{}, 0, true}, sliceState{[4]int64{}, 0, true},
	}},
	{sliceState{[4]int64{}, 0, false}, snapshotTrace{
		sliceState{[4]int64{}, 0, false}, sliceState{[4]int64{}, 0, false},
		sliceState{[4]int64{}, 0, false}, sliceState{[4]int64{}, 0, false},
	}},
	{sliceState{[4]int64{5, 0, 0, 0}, 1, false}, snapshotTrace{
		sliceState{[4]int64{5, 0, 0, 0}, 1, false}, sliceState{[4]int64{5, 0, 0, 0}, 1, false},
		sliceState{[4]int64{5, 0, 0, 0}, 1, false}, sliceState{[4]int64{73, 0, 0, 0}, 1, false},
	}},
	{sliceState{[4]int64{5, 8, 0, 0}, 2, false}, snapshotTrace{
		sliceState{[4]int64{5, 8, 0, 0}, 2, false}, sliceState{[4]int64{5, 8, 0, 0}, 2, false},
		sliceState{[4]int64{5, 8, 0, 0}, 2, false}, sliceState{[4]int64{73, 74, 0, 0}, 2, false},
	}},
	{sliceState{[4]int64{-3, 0, 7, 0}, 3, false}, snapshotTrace{
		sliceState{[4]int64{-3, 0, 7, 0}, 3, false}, sliceState{[4]int64{-3, 0, 7, 0}, 3, false},
		sliceState{[4]int64{-3, 0, 7, 0}, 3, false}, sliceState{[4]int64{73, 74, 75, 0}, 3, false},
	}},
	{sliceState{[4]int64{1, 2, 3, 4}, 4, false}, snapshotTrace{
		sliceState{[4]int64{1, 2, 3, 4}, 4, false}, sliceState{[4]int64{1, 2, 3, 4}, 4, false},
		sliceState{[4]int64{1, 2, 3, 4}, 4, false}, sliceState{[4]int64{73, 74, 75, 76}, 4, false},
	}},
}

func snapshotCorrect(input []int64) []int64 {
	if input == nil {
		return nil
	}
	out := make([]int64, len(input))
	copy(out, input)
	return out
}

func snapshotAlias(input []int64) []int64 { return input }

func snapshotMutates(input []int64) []int64 {
	if input == nil {
		return nil
	}
	if len(input) > 0 {
		input[0]++
	}
	out := make([]int64, len(input))
	copy(out, input)
	return out
}

func snapshotEmptyNil(input []int64) []int64 {
	if len(input) == 0 {
		return nil
	}
	out := make([]int64, len(input))
	copy(out, input)
	return out
}

func invokeSnapshot(fn snapshotFn, input []int64) (out []int64, panicked bool) {
	panicked = true
	defer func() { _ = recover() }()
	out = fn(input)
	panicked = false
	return
}

func validSliceState(s sliceState) bool {
	if s.Len > 4 || s.Nil && s.Len != 0 {
		return false
	}
	for i := int(s.Len); i < len(s.Values); i++ {
		if s.Values[i] != 0 {
			return false
		}
	}
	return true
}

func sliceStateOf(s []int64) (sliceState, bool) {
	if len(s) > 4 {
		return sliceState{}, false
	}
	out := sliceState{Len: uint8(len(s)), Nil: s == nil}
	copy(out.Values[:], s)
	return out, true
}

func checkSnapshot(fn snapshotFn, vectors []snapshotVector) controlCheck {
	if fn == nil || len(vectors) == 0 {
		return controlCheck{Unknown: true}
	}
	var result controlCheck
	for _, v := range vectors {
		if !validSliceState(v.Input) || !validSliceState(v.Want.Returned) || !validSliceState(v.Want.InputAfterCall) || !validSliceState(v.Want.InputAfterResultWrite) || !validSliceState(v.Want.ResultAfterInputWrite) {
			result.Unknown = true
			return result
		}
		var input []int64
		if !v.Input.Nil {
			input = make([]int64, int(v.Input.Len), int(v.Input.Len)+2)
			copy(input, v.Input.Values[:v.Input.Len])
		}
		out, panicked := invokeSnapshot(fn, input)
		result.Checked++
		// Wrong length/nilness are known contract failures, not an unsupported
		// observation. Skip unsafe inspector indexing after such a failure.
		if panicked || len(out) != int(v.Input.Len) || (out == nil) != v.Input.Nil {
			result.Failed++
			continue
		}
		var got snapshotTrace
		var ok bool
		if got.Returned, ok = sliceStateOf(out); !ok {
			result.Unknown = true
			return result
		}
		if got.InputAfterCall, ok = sliceStateOf(input); !ok {
			result.Unknown = true
			return result
		}
		for i := range out {
			out[i] = 73 + int64(i)
		}
		if got.InputAfterResultWrite, ok = sliceStateOf(input); !ok {
			result.Unknown = true
			return result
		}
		for i := range input {
			input[i] = -29 - int64(i)
		}
		if got.ResultAfterInputWrite, ok = sliceStateOf(out); !ok {
			result.Unknown = true
			return result
		}
		if got != v.Want {
			result.Failed++
		}
	}
	return result
}

var atomicSources = []struct {
	spec sourceSpec
	run  atomicFn
}{
	{sourceSpec{"atomic-correct", "atomic-commit", "staged-commit", "atomicCorrect", true}, atomicCorrect},
	{sourceSpec{"atomic-partial-commit", "atomic-commit", "staged-commit", "atomicPartialCommit", false}, atomicPartialCommit},
	{sourceSpec{"atomic-partial-result", "atomic-commit", "staged-commit", "atomicPartialResult", false}, atomicPartialResult},
	{sourceSpec{"atomic-no-commit", "atomic-commit", "staged-commit", "atomicNoCommit", false}, atomicNoCommit},
}

var identitySources = []struct {
	spec sourceSpec
	run  identityFn
}{
	{sourceSpec{"identity-correct", "error-identity", "error-cause-identity", "identityCorrect", true}, identityCorrect},
	{sourceSpec{"identity-by-message", "error-identity", "error-cause-identity", "identityByMessage", false}, identityByMessage},
	{sourceSpec{"identity-direct-only", "error-identity", "error-cause-identity", "identityDirectOnly", false}, identityDirectOnly},
	{sourceSpec{"identity-mutate-cause", "error-identity", "error-cause-identity", "identityMutateCause", false}, identityMutateCause},
}

var snapshotSources = []struct {
	spec sourceSpec
	run  snapshotFn
}{
	{sourceSpec{"snapshot-correct", "owned-snapshot", "ownership-graph", "snapshotCorrect", true}, snapshotCorrect},
	{sourceSpec{"snapshot-alias", "owned-snapshot", "ownership-graph", "snapshotAlias", false}, snapshotAlias},
	{sourceSpec{"snapshot-mutates", "owned-snapshot", "ownership-graph", "snapshotMutates", false}, snapshotMutates},
	{sourceSpec{"snapshot-empty-nil", "owned-snapshot", "ownership-graph", "snapshotEmptyNil", false}, snapshotEmptyNil},
}

func sourcesState() []sourceSpec {
	out := make([]sourceSpec, 0, len(atomicSources)+len(identitySources)+len(snapshotSources))
	for _, s := range atomicSources {
		out = append(out, s.spec)
	}
	for _, s := range identitySources {
		out = append(out, s.spec)
	}
	for _, s := range snapshotSources {
		out = append(out, s.spec)
	}
	return out
}

func checkState(id string) controlCheck {
	for _, s := range atomicSources {
		if s.spec.id == id {
			return checkAtomic(s.run, atomicVectors)
		}
	}
	for _, s := range identitySources {
		if s.spec.id == id {
			return checkIdentity(s.run, identityVectors)
		}
	}
	for _, s := range snapshotSources {
		if s.spec.id == id {
			return checkSnapshot(s.run, snapshotVectors)
		}
	}
	return controlCheck{Unknown: true}
}

func stateTruthHash(b []byte, err error) string {
	// Callers serialize only their statically typed independent literal table.
	if err != nil {
		panic("unsupported authored state truth table")
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func stateContractEvidence() []ContractSpec {
	return []ContractSpec{
		{
			Prototype: "atomic-commit", Core: "staged-commit", Vectors: len(atomicVectors), TruthTableSHA256: stateTruthHash(json.Marshal(atomicVectors)),
			InputSemantics: "Exact ASCII DD,F: two digits, comma, flag 0/1; fresh non-nil destination. Success returns/commits parsed Config. Invalid input returns zero/exact ErrSyntax and preserves destination; no candidate panic.",
		},
		{
			Prototype: "error-identity", Core: "error-cause-identity", Vectors: len(identityVectors), TruthTableSHA256: stateTruthHash(json.Marshal(identityVectors)),
			InputSemantics: "Fresh A/B sentinels with equal text and fresh non-nil Cause values. Direct, multiple wrapping, joining, same-message unrelated, and nil error cases; report errors.Is(A/B), errors.As(Cause)/code without changing causes; no candidate panic.",
		},
		{
			Prototype: "owned-snapshot", Core: "ownership-graph", Vectors: len(snapshotVectors), TruthTableSHA256: stateTruthHash(json.Marshal(snapshotVectors)),
			InputSemantics: "Fresh slices of 0..4 int64 values, distinguishing nil/non-nil empty. Preserve initial input, nilness, length and values; inspector writes each result[i]=73+i then each input[i]=-29-i to check both ownership directions; no candidate panic.",
		},
	}
}
