# Preparing a fair comparison on external Go tasks

[한국어](fair-upstream-comparison.ko.md) · [Two source contracts](public-go-contracts-54.en.md) · [Scale and partition rules](golden-set-scale.en.md)

This document describes **evaluation framework preparation** for PDCA55. It makes the requested changes visible to the model and aligns the Go language conditions used by model self-checks and independent verification. There are no actual model-comparison results for these two external requests yet. Local verifier controls and fake-executor budget checks do not establish model performance or passing public CI.

## Separate the counts first

The 24 synthetic semantic-retrieval probes are development data for checking an early mechanism. The next performance-claim target is **at least 2,400 distinct protected final requests per evaluation domain**. Retrieval, actual model routing, repository selection and decomposition each need their own ground truth and outcomes. Renaming a request already read during development does not make it protected final data.

| Scope | Count and meaning at preparation |
| --- | --- |
| Existing actual coding development observations | 5 distinct requests, 16 CLI records, three code families, one laya-tools repository |
| Existing compiled verifier tasks | 7 distinct tasks across 3 source repositories; separate from the attempt records above |
| Preserved external source inventory | 120 original candidates; not ground truth or 120 executable contracts |
| This v2 | Clarified input/evaluation versions of 2 existing external requests, not 2 additional distinct requests |
| Actual model attempts on the two external requests | 0 |
| Training runs, training labels and final-eligible requests for these two external requests | 0 each |
| Newly trained or released weights in this preparation | 0 |

Changing models, repeating a request, adding test inputs or translating it does not increase distinct-request counts. The v1 and v2 IDs share an original request through `logical_task_id` and `previous_task_id`. Two additional version IDs leave the existing seven logical verifier tasks unchanged. Do not add the [actual candidate records](../benchmarks/training/public-task-candidates.json) to the [120-source inventory](../benchmarks/training/public-go-acquisition-53.json).

## Why introduce v2?

The previous humanize Prompt was one sentence and did not fully disclose legacy preservation, mutable files or supported code shapes. UUID disclosed many constraints but not its exact allowed-call list. A comparison is unfair when the independent verifier demands conditions the model did not actually receive.

The new v2 places the complete requirements in `Spec.Prompt`. That exact Prompt is model stdin; the framework does not assume the model received separate Acceptance fields or read GitHub documentation. The actual Prompt-byte SHA, Spec, Definition, independent Contract and execution recipe are bound in plans and records. The previous seven tasks retain their Spec, Definition, Prompt, Contract and original-fixture hashes. Historical plans and attempt records are not rewritten.

| Request | v2 ID | Existing logical request | Example behavior requirement |
| --- | --- | --- | --- |
| Full int64 ordinals | `go55-humanize-ordinal64-v2` | `go53-humanize-ordinal64` | Preserve both the new API's `-21st` and the legacy `Ordinal(-21)` result `-21th` |
| Lowercase UUID parser | `go55-uuid-canonical-parse-v2` | `go53-uuid-canonical-parse` | Restrict the new API to lowercase 36-byte input while retaining broad legacy Parse/ParseBytes behavior |

The humanize contract checks the exact callable type `func(int64) string`; it does not distinguish a function declaration from a function-valued variable. UUID supports the narrower shape of appending exactly one `func ParseCanonical(s string) (UUID, error)` after a pinned original prefix.

## Toolchain and language versions differ

Even with a Go 1.27.1 executable, the `go` line in `go.mod` determines the compiler's language condition. An omitted line implies language version 1.16. [Official Go toolchain explanation](https://go.dev/doc/toolchain), [go.mod reference](https://go.dev/doc/modules/gomod-ref)

| Source | Pinned revision | Original module language | Executable |
| --- | --- | --- | --- |
| dustin/go-humanize | `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e` | Explicit Go 1.21 | `go1.27.1` |
| google/uuid | `2d3c2a9cc518326daf99a383f07c4d3c44317e4d` | Implicit Go 1.16 because the go directive is absent | `go1.27.1` |

The v2 verifier uses internally fixed original `go.mod` bytes. It validates the original file SHA, module identity and dependency-free module/go directive scope; it does not use candidate module configuration. Changing candidate `go.mod` is rejected as an unrelated source change. Candidate tests and `go.work` are not independent compiler inputs either. After execution, the temporary original module must remain unchanged.

The old v1 temporary minimal module keeps language version 1.27.1. Only v2 opts into the original-language recipe. Module SHA, language version, executable version and recipe SHA are recorded separately. No new actor `-modfile` path is introduced.

Actual local controls accepted both v2 correct implementations. Humanize recorded 6 terminal passes; UUID recorded 218, including its original suite. One wrong implementation per task was rejected. UUID entropy calls, original-header changes and additional helpers, and humanize import changes, were unknown before execution. Adding Go 1.22 integer-range syntax to the same otherwise correct implementation passed v1 but failed v2 at compilation. These are development controls establishing language-condition differences, not model attempts.

## Scope of accepted and unknown

The verifier does not use candidate-authored assertions as success evidence. Original tests and a separately pinned contract must run, with the required named tests terminating in the exact package. The complete mutable file must match trusted Go 1.27.1 gofmt.

The UUID checker supports one function after an exact original or fixed-formatter prefix, local-value writes and an explicit pure-call list. Calls through function aliases, pointers, additional helpers, concurrency and global mutation can be `verifier_unknown` even when their output would be correct. The humanize task does not inherit these UUID-specific restrictions. Unsupported shapes, unavailable verification, service errors and missing usage are not relabeled as behavioral model failures or zero usage.

Original MIT and BSD-3-Clause LICENSE bytes are separately SHA-validated and copied unchanged into the verification directory. **Candidate LICENSE contents are not assessed by the success label.** Acceptance does not certify candidate or downstream attribution compliance. UUID's original copyright header is part of the assessed source prefix and is preserved by the strict shape check. Source-license verification for these copies does not establish permission for model training, weights or dataset release.

## Two child plans and four total slots

The two sources have different revisions, so each task receives its own child plan. Each child contains one attempt per compared profile, while a parent manifest fixes both child-plan SHAs and the order of four total slots. Here, parent groups an execution budget; it does not mean training new parent/child models. The parent references child SHAs without a circular hash dependency.

For one fixed manifest and **one private durable ledger**, at most 4 slots may be reserved and only 1 controller may be active. Reservation occurs before model launch. Duplicate reservations, wrong order, mismatched task/profile/Prompt/recipe bindings and a fifth slot are rejected. Failed slots are not refunded or automatically retried. Crashes or missing launch evidence remain uncertain and stop through fail-closed handling rather than guessed recovery. Reservations, durable start markers, terminal receipts and unknown launch states are separate counts.

**This cap applies only to execution using that ledger.** It is not a host-wide or account-wide budget preventing another ledger, and it neither limits nor measures provider-internal requests/retries. Do not evade an experiment's cap by replaying it through another ledger. No unsafe recovery is provided for deleting reservation, lease or stop files, or manually resetting a ledger to resume.

## Gates before actual model execution

1. Confirm public CI and independent review for the new source and pin clean executor, Go and CLI hashes. A preparation document does not establish passing checks.
2. Verify original closure/license pins and v1-preservation regressions; freeze the v2 Spec, complete stdin, Contract and recipe. Confirm actor self-checks also use the original language and leave `go.mod` unchanged.
3. Freeze both child plans and the parent manifest before observing outcomes. Check duplicate, concurrency, order and no-refund enforcement without models, and verify the actual CLI integration separately.
4. Record the fixed-input pre-coding router observation and requested profiles before starting bounded attempts. Distinguish requested settings from served identity, and whole usage/time/failure state from verification status.

Comparative performance, money saved, subscription consumption and the ultra-small-resource target remain unknown for these two requests. Finite development controls and a four-attempt plan cannot replace 2,400 distinct final requests per domain. Labels for costs or the successful-model set require separate actual attempt evidence and eligibility review.

## Inspect the input without calling a model

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskverify ./cmd/riido-taskverify
.cache/bin/riido-taskverify --task go55-humanize-ordinal64-v2 --spec
.cache/bin/riido-taskverify --task go55-uuid-canonical-parse-v2 --spec
```

See [v2 definitions](../internal/taskverify/upstream_definitions_v2.go) for full Prompts and version links, [evaluation recipe](../internal/taskverify/evaluation_recipe.go) for recipes, and [v2 tests](../internal/taskverify/upstream_v2_test.go) for fixed hashes and local controls. [Parent budget](../internal/taskrun/parent_budget.go) and executor integration have a separate verification scope. This document contains no raw model traces, authentication, personal code or weights.
