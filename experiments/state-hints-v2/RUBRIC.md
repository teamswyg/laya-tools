# V2 authored contrast evaluation rubric

This set evaluates the single primary communicative intent of fictional text.
It contains 120 manually authored freeform cases with one row per assigned
incident-family ID. It changes no V1 data, model, runtime, or training source.
It does not select model size, reasoning level, or routing tier.

이 자료는 원본 허구 문장의 주된 의도를 판별하는 개발 평가 자료입니다.
제품의 실제 상태나 사용자 권한을 증명하는 정답 자료가 아닙니다. 문장 안에
작업 완료가 주장되어도 별도의 실제 이벤트 확인 없이 상태 변경을 허용하지
않습니다.

## Label definitions

| Intent | Positive definition | Boundary |
| --- | --- | --- |
| `question` | Requests information, an explanation, or procedural instructions. | Mentioning completion, cancellation, an error, or an attachment inside the question does not turn it into a report or execution request. Asking how cancellation works is a question; explicitly asking to cancel is a cancellation request. |
| `blocker` | Reports a current task that cannot proceed because a required input, access, resource, or prerequisite is missing. | A possible future obstacle or a general question about blocked work is insufficient. |
| `reference` | Shares or presents background material, terminology, an index, chronology, or explanatory table for readers. | A request for an answer about that material is a question. A bare quoted status without a clear sharing purpose is insufficient. |
| `progress` | Reports current activity on the scoped task, including completed subtasks followed by active remaining checks or edits. | Finishing one component does not mean the entire scoped task is complete. |
| `completion_report` | Explicitly reports that the scoped assigned work and its final handoff are finished, with no active remaining work on that assignment. | This identifies an assertion, not verified lifecycle evidence. Ongoing checks, remaining preparation, or a hypothetical ending do not qualify. |
| `cancel_request` | Explicitly requests withdrawal or cancellation of the scoped work/order/booking. | A question about cancellation, quoted cancellation wording, a historical cancellation, or a statement denying cancellation is insufficient. |
| `planned` | Expresses intended or scheduled future work whose execution has not started. | A current blocked task is a blocker; current execution is progress. An explicit future intention differs from a counterfactual with no stated plan. |
| `unclear` | The message does not support one current concrete intent under this taxonomy. | Includes unscoped fragments, quotation with uncertain live authority, counterfactual/hypothetical claims without a stated plan, contradictory whole-task accounts, and multiple independent intents with no primary target. |

Use the scoped main assignment, rather than matching keywords. Incidental
mentions of another object are allowed: cancelling a duplicate order while
retaining its original order still has one primary cancellation intent.
Two independent actionable requests, or an activity report and an unrelated
request with no primary target, have no unique single-intent answer and are
labelled `unclear`. These boundaries are author judgments; external adjudication
is required before treating them as product ground truth.

Caller-provided current state, permissions, catalog entries, revision, command
ID, and trusted event evidence are not inserted into the classifier text. No
fixture message is an attestation of those inputs. Classification and the Go
policy's permission to propose a lifecycle transition are separate checks.

## Partition protocol

- `development.jsonl`: 40 rows (`split: validation`) for checking predefined hypotheses and comparing validation candidates. Never use these rows for gradient updates.
- `calibration.jsonl`: 40 rows (`split: calibration`) for calibration after weights are selected. Use a predeclared scalar calibration method and retain the fixed 0.9 acceptance threshold.
- `locked-test.jsonl`: 40 rows (`split: test`) for final evaluation only after the candidate weights, calibration, decision gate, and metrics are locked. Its body was withheld from the parent evaluator when authored; only its hash/count metadata was returned.

Each partition has 20 Korean and 20 English cases and five cases per intent
pooled across languages. Because five is odd, each language/intent cell has
two or three cases. All 120 texts and family IDs are unique, and no assigned
group crosses partitions. `manifest.json` pins the exact files. Freeze the
manifest and rubric hashes in the next plan before any new model fit or
prediction on this set. Source integrity checks may count records and verify
hashes, but must not expose the locked-test body to the parent before selection.

Training data for a new candidate must be a separate original collection.
Check exact text and assigned-group overlap, and review scenario-level overlap.
Matching vocabulary and related conceptual contrasts cannot be eliminated by
unique IDs. Do not fit weights on any of these 120 cases, add their labels as
training targets, or use final-test outcomes to choose a candidate. Keep V1
artifacts unchanged for the preregistered baseline. Once test results are
observed, later changes must treat this set as development evidence.

The sealed file is a process boundary, not a cryptographic or access-control
guarantee. Its local owner-only permission is reversible, and the author has
seen its contents. The corpus is intended to be public after evaluation and is
not a secret dataset. The parent must implement the promised read/evaluation
order in its evaluator and record the lock and first evaluation.

## Reporting and limits

Report pooled and per-language confusion matrices, per-intent correct counts,
NLL/Brier, accepted precision, coverage, and counts of concrete accepted
predictions. Acceptance uses confidence at least 0.9 and excludes `unclear`.
Do not treat deterministic rule one-hot values as calibrated probabilities.
Keep genuine questions, completion reports, current activity with completed
subtasks, and ambiguous multi-intent cases visible as error categories.
Completion-report errors must also be examined in the separate event policy:
text alone cannot supply a trusted completion event.

The author knew the V1 failures and authored falsification cases in response.
This is new held-out authored evidence, not a blinded external benchmark or
independent product sample. Common syntax and conceptual overlap remain. Five
cases per class per partition and 40 calibration rows support only coarse
development comparisons; a small number of accepted cases cannot establish
production precision. No real user data, private prompts, real incidents,
external annotation dataset, or model-generated annotation calls were used.

All original fixture text and this rubric are supplied under Apache-2.0; the
full license is in `LICENSE`. Deployment qualification remains false.
