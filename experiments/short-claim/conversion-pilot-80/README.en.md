# Source80: connecting behavior evidence to training candidates

This small pilot connects [24 verified function inputs](../native-observation-80/ACTUAL-RESULTS.en.md) to **three user-like request drafts and nine comparison candidates**. Fixture count does not become request count. Actual candidate execution and supervision promotion are not yet complete.

| Request | Three comparison candidates |
|---|---|
| Preserve UUID errors and partial bytes | Clear bytes on error / forward results / turn errors into panics |
| Preserve database UUID state and errors | Forward Scan / clear state first / discard errors |
| Format ordinal suffixes with exceptions | Last digit only / always th / retain the 11–13 exception |

Each candidate has [concrete Go code](preparation/candidates.go.txt). Even an incorrect candidate's caption faithfully describes its code. [Four separate fields](preparation/CONVERSION-PROPOSAL.v1.json) track request satisfaction, caption fidelity, supervision eligibility and required value/state/error channels. Actual truth, role, mask, weight, Got and acceptable indices are currently null.

## Planned verification scope

The [historical preplan](preparation/CANDIDATE-VERIFICATION-PREPLAN.v1.json) remains unfrozen and unexecuted. A new observation plans **72 candidate calls** across every candidate and the same original inputs. Explicit Error calls on returned errors are expected to be 10, conservatively capped at 48. Candidate and Error dispatches together are expected to be 82, capped at 120.

Source inspection maps 56 paths to one original API call inside a candidate and 16 paths to owned Ordinal code. Without dynamic instrumentation of those inner calls, actual original API counts remain null. Seventy-two candidate dispatches are not seventy-two original API executions. Error-valued panics expose presence and type without an additional error-message call. New expectations, code and plan must be frozen before the one child process preserves all failures and differences.

## Data and roles

The [source/role review](family-review/REVIEW.v1.en.md) preserves the existing 76 requests and 19 groups. UUID Parse, Scan and all derivatives stay in one source group; Ordinal derivatives stay together too. Their declared source identities do not currently connect to existing groups, so new roles remain null. Previous surveys and coding work exposed these sources; they are not newly independent held-out data.

After candidate execution and caption supervision eligibility are verified, the four fixed controls and oracle use the same complete candidate sets. Existing groups, roles, truth, masks, weights and thresholds remain unchanged. Three requests do not establish utility or satisfy the final minimum of 2,400 requests. One further FP32 development fit will have a separate plan after evidence and whole-scope headroom checks.

Full BSD/MIT notices and derivative-code attribution are linked in [NOTICE](preparation/NOTICE.md). Original function observations and new candidate observations remain distinct; this does not rerun the old Source80 worker. [한국어](README.ko.md)
