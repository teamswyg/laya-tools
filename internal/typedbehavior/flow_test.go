package typedbehavior

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/teamswyg/laya-tools/internal/lexicalhint"
	"github.com/teamswyg/laya-tools/pkg/shortclaim"
)

func TestFlowCorrectAndWrongCompiledControls(t *testing.T) {
	sources := sourcesFlow()
	if len(sources) != 12 {
		t.Fatalf("source count=%d", len(sources))
	}
	correct, wrong := 0, 0
	for _, source := range sources {
		t.Run(source.id, func(t *testing.T) {
			got := checkFlow(source.id)
			if got.Unknown || got.Checked == 0 || source.correct && got.Failed != 0 || !source.correct && got.Failed == 0 {
				t.Fatalf("compiled control mismatch: %+v", got)
			}
		})
		if source.correct {
			correct++
		} else {
			wrong++
		}
	}
	if correct != 3 || wrong != 9 {
		t.Fatalf("control counts=%d/%d", correct, wrong)
	}
}

func TestFlowCandidatePanicsAreKnownViolations(t *testing.T) {
	lifecycle := checkLifecycle(func(lifecycleTrace) lifecycleObservation { panic("authored panic") }, lifecycleVectors[:])
	quoted := checkQuoted(func(string) quotedTokens { panic("authored panic") }, quotedVectors[:])
	graph := checkGraph(func(*flowGraph) graphObservation { panic("authored panic") }, graphVectors[:])
	for i, got := range []controlCheck{lifecycle, quoted, graph} {
		if got.Unknown || got.Checked == 0 || got.Failed != got.Checked {
			t.Fatalf("candidate panic became unknown in family%d: %+v", i, got)
		}
	}
}

func TestFlowInvalidSupportAndUnknownSourceRemainUnknown(t *testing.T) {
	if got := checkFlow("unregistered-authored-source"); !got.Unknown || got.Checked != 0 || got.Failed != 0 {
		t.Fatalf("unregistered source: %+v", got)
	}
	called := false
	life := checkLifecycle(func(lifecycleTrace) lifecycleObservation {
		called = true
		return lifecycleObservation{}
	}, []lifecycleVector{{Input: lifecycleTrace{Count: 9}}})
	if called || !life.Unknown || life.Checked != 0 || life.Failed != 0 {
		t.Fatalf("out-of-domain lifecycle was labelled: %+v", life)
	}
	for _, input := range []flowGraph{
		{Count: 0}, {Count: 5},
		{Nodes: [4]flowGraphNode{{Count: 5}}, Count: 1},
		{Nodes: [4]flowGraphNode{{Children: [4]int{1}, Count: 1}}, Count: 1},
		{Nodes: [4]flowGraphNode{{Children: [4]int{-1}, Count: 1}}, Count: 1},
	} {
		called = false
		got := checkGraph(func(*flowGraph) graphObservation {
			called = true
			return graphObservation{}
		}, []graphVector{{Input: input}})
		if called || !got.Unknown || got.Checked != 0 || got.Failed != 0 {
			t.Fatalf("out-of-domain graph was labelled: %+v", got)
		}
	}
}

func TestLifecycleCancellationReleaseOrderAndTerminalIdempotence(t *testing.T) {
	pre := lifecycleCorrect(lifecycleTrace{[8]lifecycleInputEvent{lifeCancel, lifeBegin, lifeFinish, lifeFail}, 4})
	if pre != (lifecycleObservation{[8]lifecycleOutputEvent{lifeCancelled}, 1, lifeCancellation}) {
		t.Fatalf("pre-cancel acquired/released or resumed: %+v", pre)
	}
	failed := lifecycleCorrect(lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeFail, lifeBegin, lifeCancel, lifeFinish, lifeFail, lifeBegin, lifeCancel}, 8})
	if failed != (lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeFailed, lifeReleased}, 3, lifeFailure}) {
		t.Fatalf("failure cleanup/order/idempotence: %+v", failed)
	}
	success := lifecycleCorrect(lifecycleTrace{[8]lifecycleInputEvent{lifeBegin, lifeBegin, lifeFinish, lifeCancel}, 4})
	if success != (lifecycleObservation{[8]lifecycleOutputEvent{lifeAcquired, lifeSucceeded, lifeReleased}, 3, lifeSuccess}) {
		t.Fatalf("duplicate begin or terminal cancellation: %+v", success)
	}
}

func TestQuotedRawAndDecodedBoundariesAndNoPartialErrors(t *testing.T) {
	if len(quotedVectors[len(quotedVectors)-2].Input) != 128 || len(quotedVectors[len(quotedVectors)-1].Input) != 129 {
		t.Fatal("authored raw boundary literals drifted")
	}
	for _, v := range quotedVectors {
		got := quotedCorrect(v.Input)
		if got != v.Want {
			t.Fatalf("literal tokenizer mismatch: got=%+v want=%+v", got, v.Want)
		}
		if got.Error != quotedOK && (got.Count != 0 || got.Tokens != ([8]string{})) {
			t.Fatal("an error retained partial tokens")
		}
	}
	for _, tc := range []quotedVector{
		{`a\"b`, quotedTokens{[8]string{`a"b`}, 1, quotedOK}},
		{`"a\,b"`, quotedTokens{[8]string{"a,b"}, 1, quotedOK}},
		{"a\\\x00,b", quotedTokens{[8]string{"a\x00", "b"}, 2, quotedOK}},
	} {
		if got := quotedCorrect(tc.Input); got != tc.Want {
			t.Fatalf("escape semantics mismatch: got=%+v want=%+v", got, tc.Want)
		}
	}
}

func TestGraphSharedIdentityCycleAndInputImmutability(t *testing.T) {
	diamond := flowGraph{[4]flowGraphNode{{[4]int{1, 2}, 2}, {[4]int{3}, 1}, {[4]int{3}, 1}, {}}, 4}
	original := diamond
	if got := graphCorrect(&diamond); got.Error != graphOK || diamond != original {
		t.Fatal("shared DAG rejected or input mutated")
	}
	if graphRevisitCycle(&diamond).Error != graphCycle {
		t.Fatal("global-visited negative control stopped exposing shared DAG mistake")
	}
	mutating := checkGraph(graphMutatesInput, []graphVector{{Input: diamond, Want: graphObservation{graphOK}}})
	if mutating.Unknown || mutating.Checked != 1 || mutating.Failed != 1 || diamond != original {
		t.Fatalf("mutation not detected or contaminated original: %+v", mutating)
	}
	if again := checkGraph(graphCorrect, []graphVector{{Input: diamond, Want: graphObservation{graphOK}}}); again.Unknown || again.Failed != 0 || again.Checked != 1 {
		t.Fatalf("fresh inputs not preserved after mutating candidate: %+v", again)
	}
	nested := flowGraph{[4]flowGraphNode{{[4]int{1}, 1}, {[4]int{2}, 1}, {[4]int{1}, 1}}, 3}
	if graphCorrect(&nested).Error != graphCycle || graphIgnoreCycle(&nested).Error != graphOK {
		t.Fatal("nested ancestor cycle control did not differ")
	}
}

func TestFlowLiteralEvidenceIncludesActualInputsAndExpectedObservations(t *testing.T) {
	evidence := flowContractEvidence()
	if len(evidence) != 3 {
		t.Fatal("family evidence count drifted")
	}
	tables := []any{lifecycleVectors, quotedVectors, graphVectors}
	counts := []int{len(lifecycleVectors), len(quotedVectors), len(graphVectors)}
	for i, table := range tables {
		encoded, err := json.Marshal(table)
		if err != nil || !strings.Contains(string(encoded), `"input"`) || !strings.Contains(string(encoded), `"want"`) {
			t.Fatal("table digest omitted observations")
		}
		if evidence[i].Vectors != counts[i] || evidence[i].TruthTableSHA256 != flowTruthHash(table) || len(evidence[i].TruthTableSHA256) != 64 {
			t.Fatal("table evidence pin drifted")
		}
	}
	changed := graphVectors
	changed[0].Want.Error = graphCycle
	if flowTruthHash(changed) == evidence[2].TruthTableSHA256 {
		t.Fatal("changed oracle result kept table hash")
	}
}

// These are proposals for the parent's original public fixtures, not labels or
// additional observations. Passing input bounds does not establish semantic
// completeness. In particular the quoted request omits error-kind distinctions
// and must remain unknown-incomplete-caption, despite passing this bounds test.
func TestFlowCaptionProposalsFitInputBounds(t *testing.T) {
	lifecycleRequest := "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel acquires nothing. Terminal result→release once if acquired. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."
	quotedRequest := "Split ASCII commas outside double quotes/escapes; strip syntax, preserve empties. Backslash escapes any following byte. Unclosed quotes/trailing escapes/exceeding128 input bytes/8 tokens/8 token bytes error without partial output."
	graphRequest := "Given valid nonempty directed graphs with 1–4 nodes, traverse root0. Reject reachable ancestor cycles; allow shared DAGs, duplicate edges and unreachable cycles. Return ok/cycle, preserve input, never panic."
	cases := []struct{ id, request, text string }{
		{"lifecycle-correct", lifecycleRequest, lifecycleRequest},
		{"lifecycle-preacquire", lifecycleRequest, "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel also acquires. Terminal result→release once if acquired. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
		{"lifecycle-failure-leak", lifecycleRequest, "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel acquires nothing. Result→release if acquired except fail. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
		{"lifecycle-release-first", lifecycleRequest, "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel records cancellation without acquisition. Acquired terminal release→result. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
		{"quoted-correct", quotedRequest, quotedRequest},
		{"quoted-literal-comma", quotedRequest, "Split ASCII commas even inside quotes/escapes; preserve syntax and empties. Reject input beyond128 bytes/8 tokens/8 token bytes. Unclosed quotes and trailing escapes stay literal, without syntax errors."},
		{"quoted-inside-escape", quotedRequest, "Split unquoted ASCII commas, remove double quotes, preserve empties. Backslash escapes inside quotes; elsewhere literal. Unclosed quotes/trailing quoted escapes/exceeding128 input bytes/8 tokens/8 token bytes error without partial output."},
		{"quoted-partial-error", quotedRequest, "Non-ASCII/raw>128 bytes: empty tokens. Otherwise split unquoted/unescaped commas, strip double-quote syntax/escapes, preserve empties. Syntax errors or >8 tokens/>8 decoded bytes per token retain partial tokens."},
		{"graph-correct", graphRequest, graphRequest},
		{"graph-revisit-cycle", graphRequest, "From root0 in valid nonempty 1–4-node directed graphs, reject any previously visited node, including shared DAGs/duplicate edges. Ignore unreachable cycles; return ok/cycle, preserve input, never panic."},
		{"graph-ignore-cycle", graphRequest, "Valid nonempty directed graphs with 1–4 nodes: return ok without traversal, even for reachable cycles. Preserve input, never panic; shared DAGs, duplicate edges and unreachable cycles also pass."},
		{"graph-mutates-input", graphRequest, "From root0 in valid nonempty 1–4-node graphs, reject reachable ancestor cycles; allow shared DAGs/duplicate edges/unreachable cycles. Compute ok/cycle; erase root edges, decrement count, return; never panic."},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			_, err := shortclaim.Validate(shortclaim.Input{Schema: shortclaim.Schema, Request: tc.request, Provenance: "authored-flow-caption-56b", Candidates: []shortclaim.Candidate{{ID: tc.id, Text: tc.text}}})
			if err != nil {
				t.Fatalf("caption does not fit: %v; requestWords=%d candidateWords=%d", err, len(strings.Fields(lexicalhint.NormalizeText(tc.request))), len(strings.Fields(lexicalhint.NormalizeText(tc.text))))
			}
		})
	}
}
