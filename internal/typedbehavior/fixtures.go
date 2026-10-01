package typedbehavior

import (
	"errors"
	"fmt"
	"slices"
)

type caption struct{ id, text string }
type prototypeFixture struct {
	prototype string
	requests  [2]string
	complete  bool
	ambiguous string
	sources   [4]caption
}

// These original captions are reviewed before ranking observations. Related
// paraphrases/candidate variants remain one connected provenance family.
// Incomplete descriptions remain unknown even though code controls are known.
var fixturePrototypes = []prototypeFixture{
	{
		prototype: "atomic-commit", complete: true,
		requests: [2]string{
			"Nonnil destination: parse exact ASCII DD,F: Count=DD; Enabled=(F=1); DD two digits; F=0/1. Success commits/returns Config. Failure returns zero/exact ErrSyntax, preserving destination. Never panic.",
			"For nonnil destination, read exact ASCII DD,F (DD two digits, F=0/1). Count=DD, Enabled=(F=1). Commit/return success; otherwise zero/exact ErrSyntax, unchanged destination. Never panic.",
		},
		ambiguous: "Update the configuration usefully while handling malformed input.",
		sources: [4]caption{
			{"atomic-correct", "Validate exact DD,F syntax before writing. Success returns and commits parsed Config; failure returns zero and ErrSyntax without modifying destination."},
			{"atomic-partial-commit", "Update destination Count from a leading ASCII digit pair before validating DD,F. Then commit full valid Config or return zero and ErrSyntax."},
			{"atomic-partial-result", "Validate DD,F. On failure, preserve destination; return Count from the first two ASCII digits, or zero if invalid, plus ErrSyntax. Success commits and returns Config."},
			{"atomic-no-commit", "Validate DD,F; return parsed Config on success without writing destination. Invalid input returns zero and ErrSyntax, leaving destination unchanged."},
		},
	},
	{
		prototype: "error-identity", complete: true,
		requests: [2]string{
			"Report errors.Is(A/B) and errors.As(Cause)/original code through wrapping/joining. Targets and Cause pointers are nonnil. Equal messages alone never match. Preserve input causes; never panic.",
			"Use Is(A/B) and As(Cause) through joined/wrapped errors; report original Cause code. Nonnil targets/Cause pointers; equal messages do not establish identity. Preserve causes, never panic.",
		},
		ambiguous: "Recognize useful errors and choose an appropriate result.",
		sources: [4]caption{
			{"identity-correct", "Use errors.Is for both targets and errors.As for Cause. Return its original code without changing the error chain."},
			{"identity-by-message", "Match targets by finding their messages in the input error text. Use errors.As for Cause and return its code unchanged."},
			{"identity-direct-only", "Compare input directly with each target and use a direct Cause type assertion. Return the direct cause code without modifying it."},
			{"identity-mutate-cause", "Use errors.Is and errors.As; capture the original Cause code, increment that Cause's stored code, and return the captured facts."},
		},
	},
	{
		prototype: "owned-snapshot", complete: true,
		requests: [2]string{
			"Return an independent slice with identical length, values, and nilness. Mutating either input or result must not alter the other. Leave input unchanged during copying; never panic.",
			"Preserve slice values, length, nilness, and input during copying. Return independent storage: writes to either slice cannot change the other. Never panic.",
		},
		ambiguous: "Return a suitable view of these values.",
		sources: [4]caption{
			{"snapshot-correct", "Return nil for nil input; otherwise allocate and copy every value, preserving non-nil empty slices and leaving input unchanged."},
			{"snapshot-alias", "Return the input slice directly with the same nilness, length, and values, so input and result share their backing array."},
			{"snapshot-mutates", "For nonempty input, increment its first value, then allocate and copy all values. Preserve nilness and non-nil empty slices."},
			{"snapshot-empty-nil", "Return nil whenever length is zero; otherwise allocate and copy every value without changing input."},
		},
	},
	{
		prototype: "cancellation-lifecycle", complete: true,
		requests: [2]string{
			"≤8 events: begin acquires once; finish/fail require begin. Pre-cancel acquires nothing. Terminal result→release once if acquired. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic.",
			"≤8 events: begin acquires once; finish/fail require begin. Ignore duplicate begin/post-terminal events. Pre-cancel acquires nothing. Terminal result→release once if acquired. Return pending/success/failure/cancellation, never panic.",
		},
		ambiguous: "Manage cancellation and resources appropriately.",
		sources: [4]caption{
			{"lifecycle-correct", "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel acquires nothing. Terminal result→release once if acquired. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
			{"lifecycle-preacquire", "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel also acquires. Terminal result→release once if acquired. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
			{"lifecycle-failure-leak", "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel acquires nothing. Result→release if acquired except fail. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
			{"lifecycle-release-first", "≤8 events: begin acquires once; finish/fail require begin. Pre-cancel records cancellation without acquisition. Acquired terminal release→result. Ignore duplicate begin/post-terminal events. Return pending/success/failure/cancellation, never panic."},
		},
	},
	{
		prototype: "quoted-delimiters", complete: false,
		requests: [2]string{
			"Split ASCII commas outside double quotes/escapes; strip syntax, preserve empties. Backslash escapes any following byte. Unclosed quotes/trailing escapes/exceeding128 input bytes/8 tokens/8 token bytes error without partial output.",
			"Tokenize commas outside double quotes; remove quotes, backslash escapes next byte; retain empties. Reject nonASCII/unclosed quotes/trailing escapes/bounds, clearing output. Limits: 128 input bytes, 8 tokens, 8 bytes/token.",
		},
		ambiguous: "Split the fields sensibly while respecting quoting.",
		sources: [4]caption{
			{"quoted-correct", "Split ASCII commas outside double quotes/escapes; strip syntax, preserve empties. Backslash escapes any following byte. Unclosed quotes/trailing escapes/exceeding128 input bytes/8 tokens/8 token bytes error without partial output."},
			{"quoted-literal-comma", "Split ASCII commas even inside quotes/escapes; preserve syntax and empties. Reject input beyond128 bytes/8 tokens/8 token bytes. Unclosed quotes and trailing escapes stay literal, without syntax errors."},
			{"quoted-inside-escape", "Split unquoted ASCII commas, remove double quotes, preserve empties. Backslash escapes inside quotes; elsewhere literal. Unclosed quotes/trailing quoted escapes/exceeding128 input bytes/8 tokens/8 token bytes error without partial output."},
			{"quoted-partial-error", "Non-ASCII/raw>128 bytes: empty tokens. Otherwise split unquoted/unescaped commas, strip double-quote syntax/escapes, preserve empties. Syntax errors or >8 tokens/>8 decoded bytes per token retain partial tokens."},
		},
	},
	{
		prototype: "ancestor-cycle", complete: true,
		requests: [2]string{
			"Given valid nonempty directed graphs with 1–4 nodes, traverse root0. Reject reachable ancestor cycles; allow shared DAGs, duplicate edges and unreachable cycles. Return ok/cycle, preserve input, never panic.",
			"Valid nonempty 1–4 node directed graph, root0: reject reachable ancestor repeats; allow shared descendants, duplicate edges, and unreachable cycles. Return ok/cycle, preserve input; never panic.",
		},
		ambiguous: "Choose a reasonable way to inspect this graph.",
		sources: [4]caption{
			{"graph-correct", "Given valid nonempty directed graphs with 1–4 nodes, traverse root0. Reject reachable ancestor cycles; allow shared DAGs, duplicate edges and unreachable cycles. Return ok/cycle, preserve input, never panic."},
			{"graph-revisit-cycle", "From root0 in valid nonempty 1–4-node directed graphs, reject any previously visited node, including shared DAGs/duplicate edges. Ignore unreachable cycles; return ok/cycle, preserve input, never panic."},
			{"graph-ignore-cycle", "Valid nonempty directed graphs with 1–4 nodes: return ok without traversal, even for reachable cycles. Preserve input, never panic; shared DAGs, duplicate edges and unreachable cycles also pass."},
			{"graph-mutates-input", "From root0 in valid nonempty 1–4-node graphs, reject reachable ancestor cycles; allow shared DAGs/duplicate edges/unreachable cycles. Compute ok/cycle; erase root edges, decrement count, return; never panic."},
		},
	},
}

// PrepareDataset generates only authored text/provenance transport. It never
// runs candidates, labels an outcome, computes group counts or scores captions.
func PrepareDataset() (Dataset, error) {
	pins, e := SourcePins()
	if e != nil {
		return Dataset{}, e
	}
	d := Dataset{Schema: Schema, Origin: Origin}
	for i, f := range fixturePrototypes {
		for variant := 0; variant < 4; variant++ {
			request := f.requests[min(variant, 1)]
			contract := f.prototype + "-v2"
			if !f.complete {
				contract = "unknown-incomplete-caption"
			}
			if variant == 2 {
				request = f.requests[0]
			}
			if variant == 3 {
				request = f.ambiguous
				contract = "unknown-ambiguous"
			}
			p := Parent{ID: fmt.Sprintf("typed56b-p%02d-%d", i+1, variant+1), Prototype: f.prototype, ContractID: contract, Request: request}
			order := [3]int{1, 0, 2}
			if variant == 1 {
				order = [3]int{2, 3, 0}
			}
			if variant == 2 {
				order = [3]int{1, 2, 3}
			}
			if variant == 3 {
				order = [3]int{0, 1, 3}
			}
			// Rotate known candidate positions by prototype; declare authored order,
			// never adjust it in response to measured ranks.
			if variant < 2 {
				r := (i + variant) % 3
				order = [3]int{order[r], order[(r+1)%3], order[(r+2)%3]}
			}
			for j, index := range order {
				c := f.sources[index]
				n := slices.IndexFunc(pins, func(s SourcePin) bool { return s.ID == c.id })
				if n < 0 || pins[n].Prototype != f.prototype {
					return Dataset{}, errors.New("fixture_registry_identity_missing")
				}
				s := pins[n]
				p.Candidates = append(p.Candidates, Candidate{ID: fmt.Sprintf("candidate-%d", j), Text: c.text, SourceID: s.ID, CodeSHA256: s.CodeSHA256, BundleSHA256: s.BundleSHA256})
			}
			d.Parents = append(d.Parents, p)
		}
	}
	if e = validateDataset(d, pins); e != nil {
		return Dataset{}, e
	}
	return d, nil
}
