package typedbehavior

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
)

// The source and literal tables are authored here. They establish finite
// observations, not a general language oracle or arbitrary-code execution.
//
//go:embed flow.go
var flowSourceText string

func sourcesFlow() []sourceSpec {
	return []sourceSpec{
		{"lifecycle-correct", "cancellation-lifecycle", "lifecycle-events", "lifecycleCorrect", true},
		{"lifecycle-preacquire", "cancellation-lifecycle", "lifecycle-events", "lifecyclePreacquire", false},
		{"lifecycle-failure-leak", "cancellation-lifecycle", "lifecycle-events", "lifecycleFailureLeak", false},
		{"lifecycle-release-first", "cancellation-lifecycle", "lifecycle-events", "lifecycleReleaseFirst", false},
		{"quoted-correct", "quoted-delimiters", "lexical-state", "quotedCorrect", true},
		{"quoted-literal-comma", "quoted-delimiters", "lexical-state", "quotedLiteralComma", false},
		{"quoted-inside-escape", "quoted-delimiters", "lexical-state", "quotedInsideEscape", false},
		{"quoted-partial-error", "quoted-delimiters", "lexical-state", "quotedPartialError", false},
		{"graph-correct", "ancestor-cycle", "graph-identity-path", "graphCorrect", true},
		{"graph-revisit-cycle", "ancestor-cycle", "graph-identity-path", "graphRevisitCycle", false},
		{"graph-ignore-cycle", "ancestor-cycle", "graph-identity-path", "graphIgnoreCycle", false},
		{"graph-mutates-input", "ancestor-cycle", "graph-identity-path", "graphMutatesInput", false},
	}
}

func checkFlow(id string) (out controlCheck) {
	// A failure of dispatch/table support remains unknown. The narrower call
	// observers below separately capture a candidate panic as a known mismatch.
	defer func() {
		if recover() != nil {
			out.Unknown = true
		}
	}()
	switch id {
	case "lifecycle-correct":
		return checkLifecycle(lifecycleCorrect, lifecycleVectors[:])
	case "lifecycle-preacquire":
		return checkLifecycle(lifecyclePreacquire, lifecycleVectors[:])
	case "lifecycle-failure-leak":
		return checkLifecycle(lifecycleFailureLeak, lifecycleVectors[:])
	case "lifecycle-release-first":
		return checkLifecycle(lifecycleReleaseFirst, lifecycleVectors[:])
	case "quoted-correct":
		return checkQuoted(quotedCorrect, quotedVectors[:])
	case "quoted-literal-comma":
		return checkQuoted(quotedLiteralComma, quotedVectors[:])
	case "quoted-inside-escape":
		return checkQuoted(quotedInsideEscape, quotedVectors[:])
	case "quoted-partial-error":
		return checkQuoted(quotedPartialError, quotedVectors[:])
	case "graph-correct":
		return checkGraph(graphCorrect, graphVectors[:])
	case "graph-revisit-cycle":
		return checkGraph(graphRevisitCycle, graphVectors[:])
	case "graph-ignore-cycle":
		return checkGraph(graphIgnoreCycle, graphVectors[:])
	case "graph-mutates-input":
		return checkGraph(graphMutatesInput, graphVectors[:])
	default:
		return controlCheck{Unknown: true}
	}
}

func flowContractEvidence() []ContractSpec {
	return []ContractSpec{
		{"cancellation-lifecycle", "lifecycle-events", "Finite event traces, count0..8, events begin/finish/fail/cancel. Finish/fail occur only after begin. Begin duplicates and all post-terminal events are ignored. Pre-cancel records cancellation without acquisition. Acquired completion records its success/failure/cancellation before one release; status is pending/success/failure/cancellation. No real resources, clocks, goroutines or network.", len(lifecycleVectors), flowTruthHash(lifecycleVectors)},
		{"quoted-delimiters", "lexical-state", "Byte tokenizer: ASCII input, at most128 bytes, comma delimiter outside double quotes, quotes stripped anywhere, backslash escapes any following byte inside or outside quotes, empty tokens preserved. At most8 tokens and8 decoded bytes per token. Unclosed quote/trailing escape returns syntax error, bound overflow returns bounds error, non-ASCII returns ASCII error; every error returns zero count and zero tokens. The table covers only its literal strings.", len(quotedVectors), flowTruthHash(quotedVectors)},
		{"ancestor-cycle", "graph-identity-path", "Valid directed graphs have1..4 nodes, root0, each0..4 child indices in0..nodecount-1. Only reachable ancestor cycles reject; shared descendants, duplicate edges and unreachable cycles are allowed. Input value must remain unchanged. Out-of-domain counts/indices are excluded from semantic scoring and are observer-support failures, never candidate labels. No error/path trace beyond the finite ok/cycle enum is claimed.", len(graphVectors), flowTruthHash(graphVectors)},
	}
}

func flowTruthHash(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("flow_literal_table_encoding_failed")
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

type lifecycleInputEvent uint8
type lifecycleOutputEvent uint8
type lifecycleStatus uint8

const (
	lifeBegin lifecycleInputEvent = iota + 1
	lifeFinish
	lifeFail
	lifeCancel
)

const (
	lifeAcquired lifecycleOutputEvent = iota + 1
	lifeSucceeded
	lifeFailed
	lifeCancelled
	lifeReleased
)

const (
	lifePending lifecycleStatus = iota
	lifeSuccess
	lifeFailure
	lifeCancellation
	lifeInvalid
)

type lifecycleTrace struct {
	Events [8]lifecycleInputEvent `json:"events"`
	Count  int                    `json:"count"`
}

type lifecycleObservation struct {
	Events [8]lifecycleOutputEvent `json:"events"`
	Count  int                     `json:"count"`
	Status lifecycleStatus         `json:"status"`
}

type lifecycleVector struct {
	Input lifecycleTrace       `json:"input"`
	Want  lifecycleObservation `json:"want"`
}

// Expected traces are literal and never obtained from a candidate function.
var lifecycleVectors = [...]lifecycleVector{
	{lifecycleTrace{}, lifecycleObservation{}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeCancel}, 1}, lifecycleObservation{[8]lifecycleOutputEvent{lifeCancelled}, 1, lifeCancellation}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeCancel, lifeBegin}, 2}, lifecycleObservation{[8]lifecycleOutputEvent{lifeCancelled}, 1, lifeCancellation}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin}, 1}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired}, 1, lifePending}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeBegin}, 2}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired}, 1, lifePending}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeFinish}, 2}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeSucceeded, lifeReleased}, 3, lifeSuccess}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeFail}, 2}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeFailed, lifeReleased}, 3, lifeFailure}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeCancel}, 2}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeCancelled, lifeReleased}, 3, lifeCancellation}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeFinish, lifeFinish, lifeCancel}, 4}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeSucceeded, lifeReleased}, 3, lifeSuccess}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeFail, lifeFinish, lifeBegin}, 4}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeFailed, lifeReleased}, 3, lifeFailure}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeCancel, lifeCancel, lifeBegin}, 3}, lifecycleObservation{[8]lifecycleOutputEvent{lifeCancelled}, 1, lifeCancellation}},
	{lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeCancel, lifeBegin, lifeCancel, lifeFinish, lifeFail, lifeBegin, lifeCancel}, 8}, lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeCancelled, lifeReleased}, 3, lifeCancellation}},
}

func lifecycleCorrect(in lifecycleTrace) lifecycleObservation {
	return runLifecycle(in, false, false, false)
}
func lifecyclePreacquire(in lifecycleTrace) lifecycleObservation {
	return runLifecycle(in, true, false, false)
}
func lifecycleFailureLeak(in lifecycleTrace) lifecycleObservation {
	return runLifecycle(in, false, true, false)
}
func lifecycleReleaseFirst(in lifecycleTrace) lifecycleObservation {
	return runLifecycle(in, false, false, true)
}

func runLifecycle(in lifecycleTrace, preacquire, failureLeak, releaseFirst bool) lifecycleObservation {
	if !validLifecycle(in) {
		return lifecycleObservation{Status: lifeInvalid}
	}
	var out lifecycleObservation
	acquired, done := false, false
	emit := func(e lifecycleOutputEvent) {
		out.Events[out.Count] = e
		out.Count++
	}
	for i := 0; i < in.Count; i++ {
		if done {
			continue
		}
		e := in.Events[i]
		if e == lifeBegin {
			if !acquired {
				emit(lifeAcquired)
				acquired = true
			}
			continue
		}
		if e == lifeCancel && !acquired && preacquire {
			emit(lifeAcquired)
			acquired = true
		}
		var result lifecycleOutputEvent
		switch e {
		case lifeFinish:
			result, out.Status = lifeSucceeded, lifeSuccess
		case lifeFail:
			result, out.Status = lifeFailed, lifeFailure
		case lifeCancel:
			result, out.Status = lifeCancelled, lifeCancellation
		}
		release := acquired && !(failureLeak && e == lifeFail)
		if release && releaseFirst {
			emit(lifeReleased)
		}
		emit(result)
		if release && !releaseFirst {
			emit(lifeReleased)
		}
		done = true
	}
	return out
}

func validLifecycle(in lifecycleTrace) bool {
	if in.Count < 0 || in.Count > len(in.Events) {
		return false
	}
	begun, terminal := false, false
	for i := 0; i < in.Count; i++ {
		e := in.Events[i]
		if e < lifeBegin || e > lifeCancel {
			return false
		}
		if terminal {
			continue
		}
		if (e == lifeFinish || e == lifeFail) && !begun {
			return false
		}
		if e == lifeBegin {
			begun = true
		} else {
			terminal = true
		}
	}
	return true
}

func observeLifecycle(fn func(lifecycleTrace) lifecycleObservation, in lifecycleTrace) (got lifecycleObservation, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	got = fn(in)
	return
}

func checkLifecycle(fn func(lifecycleTrace) lifecycleObservation, vectors []lifecycleVector) controlCheck {
	var out controlCheck
	for _, v := range vectors {
		if !validLifecycle(v.Input) {
			out.Unknown = true
			return out
		}
		got, panicked := observeLifecycle(fn, v.Input)
		out.Checked++
		if panicked || got != v.Want {
			out.Failed++
		}
	}
	return out
}

type quotedError uint8

const (
	quotedOK quotedError = iota
	quotedSyntax
	quotedBounds
	quotedASCII
)

type quotedTokens struct {
	Tokens [8]string   `json:"tokens"`
	Count  int         `json:"count"`
	Error  quotedError `json:"error"`
}

type quotedVector struct {
	Input string       `json:"input"`
	Want  quotedTokens `json:"want"`
}

var quotedVectors = [...]quotedVector{
	{"", quotedTokens{[8]string{""}, 1, quotedOK}},
	{",", quotedTokens{[8]string{"", ""}, 2, quotedOK}},
	{",a,", quotedTokens{[8]string{"", "a", ""}, 3, quotedOK}},
	{"a,b", quotedTokens{[8]string{"a", "b"}, 2, quotedOK}},
	{`"a,b",c`, quotedTokens{[8]string{"a,b", "c"}, 2, quotedOK}},
	{`ab"c,d"ef`, quotedTokens{[8]string{"abc,def"}, 1, quotedOK}},
	{`a\,b,c`, quotedTokens{[8]string{"a,b", "c"}, 2, quotedOK}},
	{`"a\"b",c`, quotedTokens{[8]string{`a"b`, "c"}, 2, quotedOK}},
	{`a\\,b`, quotedTokens{[8]string{`a\`, "b"}, 2, quotedOK}},
	{`""`, quotedTokens{[8]string{""}, 1, quotedOK}},
	{`"a,b`, quotedTokens{Error: quotedSyntax}},
	{`a\`, quotedTokens{Error: quotedSyntax}},
	{`a,b\`, quotedTokens{Error: quotedSyntax}},
	{`"a\`, quotedTokens{Error: quotedSyntax}},
	{`a,"b`, quotedTokens{Error: quotedSyntax}},
	{"12345678", quotedTokens{[8]string{"12345678"}, 1, quotedOK}},
	{"123456789", quotedTokens{Error: quotedBounds}},
	{",,,,,,,", quotedTokens{[8]string{}, 8, quotedOK}},
	{",,,,,,,,", quotedTokens{Error: quotedBounds}},
	{"1,2,3,4,5,6,7,8", quotedTokens{[8]string{"1", "2", "3", "4", "5", "6", "7", "8"}, 8, quotedOK}},
	{"é", quotedTokens{Error: quotedASCII}},
	{`\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\aaaaaaaa`, quotedTokens{[8]string{"aaaaaaaa", "aaaaaaaa", "aaaaaaaa", "aaaaaaaa", "aaaaaaaa", "aaaaaaaa", "aaaaaaaa", "aaaaaaaa"}, 8, quotedOK}},
	{`\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\a\a\a\a\a\a\a,\a\aaaaaaa`, quotedTokens{Error: quotedBounds}},
}

type quotedMode uint8

const (
	quotedNormal quotedMode = iota
	quotedCommaLiteral
	quotedQuotedEscapes
	quotedPartial
)

func quotedCorrect(in string) quotedTokens      { return scanQuoted(in, quotedNormal) }
func quotedLiteralComma(in string) quotedTokens { return scanQuoted(in, quotedCommaLiteral) }
func quotedInsideEscape(in string) quotedTokens { return scanQuoted(in, quotedQuotedEscapes) }
func quotedPartialError(in string) quotedTokens { return scanQuoted(in, quotedPartial) }

func scanQuoted(in string, mode quotedMode) quotedTokens {
	if len(in) > 128 {
		return quotedTokens{Error: quotedBounds}
	}
	for i := 0; i < len(in); i++ {
		if in[i] > 127 {
			return quotedTokens{Error: quotedASCII}
		}
	}
	var out quotedTokens
	inQuote := false
	fail := func(kind quotedError) quotedTokens {
		if mode == quotedPartial {
			out.Error = kind
			if out.Count < len(out.Tokens) {
				out.Count++
			}
			return out
		}
		return quotedTokens{Error: kind}
	}
	for i := 0; i < len(in); i++ {
		c := in[i]
		if mode != quotedCommaLiteral {
			if c == '\\' && (inQuote || mode != quotedQuotedEscapes) {
				i++
				if i == len(in) {
					return fail(quotedSyntax)
				}
				c = in[i]
			} else if c == '"' {
				inQuote = !inQuote
				continue
			} else if c == ',' && !inQuote {
				out.Count++
				if out.Count == len(out.Tokens) {
					return fail(quotedBounds)
				}
				continue
			}
		} else if c == ',' {
			out.Count++
			if out.Count == len(out.Tokens) {
				return fail(quotedBounds)
			}
			continue
		}
		if len(out.Tokens[out.Count]) == 8 {
			return fail(quotedBounds)
		}
		out.Tokens[out.Count] += string(c)
	}
	if inQuote {
		return fail(quotedSyntax)
	}
	out.Count++
	return out
}

func observeQuoted(fn func(string) quotedTokens, in string) (got quotedTokens, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	got = fn(in)
	return
}

func checkQuoted(fn func(string) quotedTokens, vectors []quotedVector) controlCheck {
	var out controlCheck
	for _, v := range vectors {
		got, panicked := observeQuoted(fn, v.Input)
		out.Checked++
		if panicked || got != v.Want {
			out.Failed++
		}
	}
	return out
}

type flowGraphNode struct {
	Children [4]int `json:"children"`
	Count    int    `json:"count"`
}

type flowGraph struct {
	Nodes [4]flowGraphNode `json:"nodes"`
	Count int              `json:"count"`
}

type graphError uint8

const (
	graphOK graphError = iota
	graphCycle
	graphInvalid
)

type graphObservation struct {
	Error graphError `json:"error"`
}

type graphVector struct {
	Input flowGraph        `json:"input"`
	Want  graphObservation `json:"want"`
}

var graphVectors = [...]graphVector{
	{flowGraph{Count: 1}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{1}, 1}, {[4]int{2}, 1}, {[4]int{3}, 1}, {}}, 4}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{1, 2}, 2}, {[4]int{3}, 1}, {[4]int{3}, 1}, {}}, 4}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{1, 1}, 2}, {}}, 2}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{0}, 1}}, 1}, graphObservation{graphCycle}},
	{flowGraph{[4]flowGraphNode{{[4]int{1}, 1}, {[4]int{0}, 1}}, 2}, graphObservation{graphCycle}},
	{flowGraph{[4]flowGraphNode{{[4]int{1}, 1}, {[4]int{2}, 1}, {[4]int{1}, 1}}, 3}, graphObservation{graphCycle}},
	{flowGraph{[4]flowGraphNode{{}, {[4]int{2}, 1}, {[4]int{1}, 1}}, 3}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{1, 2}, 2}, {[4]int{2}, 1}, {}}, 3}, graphObservation{graphOK}},
	{flowGraph{[4]flowGraphNode{{[4]int{1, 1, 1, 1}, 4}, {[4]int{2, 2, 2, 2}, 4}, {[4]int{3, 3, 3, 3}, 4}, {}}, 4}, graphObservation{graphOK}},
}

func validFlowGraph(g *flowGraph) bool {
	if g == nil || g.Count < 1 || g.Count > len(g.Nodes) {
		return false
	}
	for i := 0; i < g.Count; i++ {
		n := g.Nodes[i]
		if n.Count < 0 || n.Count > len(n.Children) {
			return false
		}
		for j := 0; j < n.Count; j++ {
			if n.Children[j] < 0 || n.Children[j] >= g.Count {
				return false
			}
		}
	}
	return true
}

func graphCorrect(g *flowGraph) graphObservation {
	if !validFlowGraph(g) {
		return graphObservation{graphInvalid}
	}
	var path [4]bool
	return graphObservation{walkFlowGraph(g, 0, &path, false)}
}

func graphRevisitCycle(g *flowGraph) graphObservation {
	if !validFlowGraph(g) {
		return graphObservation{graphInvalid}
	}
	var visited [4]bool
	return graphObservation{walkFlowGraph(g, 0, &visited, true)}
}

func graphIgnoreCycle(g *flowGraph) graphObservation {
	if !validFlowGraph(g) {
		return graphObservation{graphInvalid}
	}
	return graphObservation{graphOK}
}

func graphMutatesInput(g *flowGraph) graphObservation {
	out := graphCorrect(g)
	if validFlowGraph(g) {
		g.Nodes[0].Count = 0
		g.Count--
	}
	return out
}

// Marking an ancestor before recursion bounds every call path to at most four
// nodes. The revisit mutant also terminates; it simply keeps marks too long.
func walkFlowGraph(g *flowGraph, node int, marks *[4]bool, keepMarks bool) graphError {
	if marks[node] {
		return graphCycle
	}
	marks[node] = true
	for i := 0; i < g.Nodes[node].Count; i++ {
		if err := walkFlowGraph(g, g.Nodes[node].Children[i], marks, keepMarks); err != graphOK {
			return err
		}
	}
	if !keepMarks {
		marks[node] = false
	}
	return graphOK
}

func observeGraph(fn func(*flowGraph) graphObservation, in *flowGraph) (got graphObservation, panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	got = fn(in)
	return
}

func checkGraph(fn func(*flowGraph) graphObservation, vectors []graphVector) controlCheck {
	var out controlCheck
	for _, v := range vectors {
		if !validFlowGraph(&v.Input) {
			out.Unknown = true
			return out
		}
		// Every call gets its own fixed-array value; mutations cannot contaminate
		// the literal table or another candidate, and are themselves mismatches.
		input := v.Input
		got, panicked := observeGraph(fn, &input)
		out.Checked++
		if panicked || input != v.Input || got != v.Want {
			out.Failed++
		}
	}
	return out
}
