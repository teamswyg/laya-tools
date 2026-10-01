# Execute a public task and connect its evidence

[한국어](task-execution.ko.md)

`riido-taskrun` is a Go executor that gives one small public coding task to Codex
CLI and **connects that invocation to its files, usage, process termination, and
independent verification**. It does not automatically intercept ordinary Codex
usage. Execution requires `--execute`. It is currently a macOS development tool
for maintainers checking the measurement procedure.

## Why an owned executor is needed

A usage trace does not establish that the code is correct. A tested candidate
does not establish which invocation produced it. The
[existing outcome tools](task-outcomes.en.md) provide usage parsing and file
verification separately. The new executor connects a process it starts directly
to the resulting evidence.

It copies a fixed public task and base files into a fresh workspace, then makes
one invocation with an explicit model and reasoning request. After terminating
and waiting for the process group, it retains the candidate and checks the
requirements independently of the model's claims and authored test assertions.
It records unsuccessful attempts too; a savings claim cannot select only a
successful first attempt.

This evidence covers **one owned CLI invocation**. It does not establish the
provider's actual model identity, the number of internal model requests, or the
absence of other attempts.

## Current development pilot and preserved history

[Precommitted plan 50](../experiments/task-outcomes/plan-50.json) stopped before
coding-model launch when the packaged executor could not resolve its Go toolchain.
[Those results](../experiments/task-outcomes/RESULTS-50.en.md) are retained; modified
code does not overwrite that plan with successful outcomes.

[Plan 51](../experiments/task-outcomes/plan-51.json) stopped after two attempts on
one comment task. [Its results](../experiments/task-outcomes/RESULTS-51.en.md)
preserve the exact-contract rejection, requested-profile support failure, and
startup-parser compatibility gap.

The repaired tools were used under a separate
[plan 52](../experiments/task-outcomes/plan-52.json), frozen first at commit
`d202ff042106ad36209ef4cdee6ca1e1c9e64514`. Three existing public development tasks
requested `gpt-6-sol` and `gpt-6-luna`, both with `low`. **All six planned Codex
CLI attempts were recorded**, and all six candidates passed independent checks
within their declared closures. Five attempts exited zero with complete core
usage. The Sol attempt on `catalog-min-context` reached the 120-second deadline
and has unknown whole usage. Candidate acceptance and completed execution must
not be combined into six successful runs. Read
[results 52](../experiments/task-outcomes/RESULTS-52.en.md) for per-attempt evidence.

Plan 52 ran one attempt at a time, with a 120-second main-attempt deadline and a
separate 45-second verification deadline. Completed plans 52/53 and stopped plans
50/51 are not rerun. A new actual run needs a separate plan with verified current
profiles and tools.

The six attempts in 52 or four in 53 do not count internal LLM calls. One CLI invocation may contain
multiple model requests and provider retries; those counts are currently
`unknown`. The executor performs no retries, resume, or fallback. A started
failure or timeout counts toward the pilot's limit. The executor owns only
one attempt, so the maintainer must also enforce the complete plan's count and
order and retain entries that were not executed.

| Task ID | Requirement checked independently | Declared file scope |
| --- | --- | --- |
| `comment-preview-authority` | Make the exact comment change stating that a repository suggestion does not authorize execution | One file, `pkg/reporouter/router.go`; exact change and gofmt |
| `comment-budget-period` | Replace the budget-period comment with the specified sentence | Three pinned files; exact change, gofmt, and pinned tests |
| `catalog-min-context` | Add optional `MinContext`, reject negative values, apply minimum-context eligibility, and preserve existing behavior | Nine pinned files; boundary/JSON contracts and pinned catalog/planner tests |

The shared public base for those three tasks is revision
`6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`, represented by
`internal/taskverify/testdata/base`. The executor checks hashes of its files,
LICENSE and NOTICE, task specifications, and independent contract tests. Added
files and runtime caches outside the declared scope are `unassessed`.
`accepted` establishes compliance **within that task scope**, not approval of the
entire repository or arbitrary code.

The behavioral task `repo-keyword-language-guard` uses the separately versioned
base at `146b02b9c37e6d90a11386050ccfdffc036c113e`, stored in
`internal/taskverify/testdata/repo-keyword-language-guard-v1`. Its contract checks
every keyword of selected repository candidates before Judge invocation. The
three-file source scope is `go.mod`, `pkg/reporouter/router.go` and
`pkg/reporouter/router_test.go`, with LICENSE/NOTICE pinned separately.
The original three task specifications and hashes remain unchanged.

[Experiment 53](../experiments/task-outcomes/RESULTS-53.en.md) ran **this one request
twice per Luna/Sol low profile: four CLI attempts**. All four candidates passed
independent closure checks with zero exit and complete whole core usage.
Four repetitions are not four distinct requests. All four produced the same
mutable-source change; the first completed accepted candidate's keyword guard
was integrated into runtime and checked with the pinned independent contract.
Candidate-authored tests were not adopted as acceptance evidence. Authored
reference/mutant verification remains separate from model outcomes.

The historical 51/52 subtotal is eight records over three requests. Including 53,
the current total is **12 owned records, four distinct requests, two actually
attempted code families and one repository**. Preserve 52's six accepted
candidates, 53's four and 51's rejection/support error as originally recorded.
The registry has eight candidates in three families; the new parser family has
no actual attempts. A separate `taskoutcome-event-key-bounds` candidate has a prepared
versioned verifier for duplicate top-level JSONL event keys, a 64-key limit
and a 128-byte key-length limit. **It has zero actual model attempts**; its
specification/checks are versioned and separate from actual model outcomes. It does not change the
contract used by these four attempts or existing sealed records.

The [public Go inventory](public-go-acquisition-53.en.md) has 120 development
candidates from eight repositories and **zero execution-eligible tasks**.
Do not add them to attempt or training-label counts. They do not replace the
[2,400-case final design](golden-set-scale.en.md) or
[acquisition, split and execution-budget plan](golden-set-acquisition.en.md).

## What Laya predicted before coding outcomes

The [recorded predictions](../experiments/task-outcomes/routing-predictions-50.json)
used the unchanged pinned Laya INT8 checkpoint on CPU with one thread and the
default threshold of `0.9`. No threshold adjustment or training on these three
tasks took place. Plan 52 references these unchanged predictions; it does not
count them as new predictions or three new independent tasks.

| Task | Author's difficulty hypothesis | Laya suggestion | Largest probability | Applied tier |
| --- | --- | --- | ---: | --- |
| `comment-preview-authority` | fast | standard | 0.480 | strong, abstained |
| `comment-budget-period` | fast | fast | 0.486 | strong, abstained |
| `catalog-min-context` | standard | standard | 0.591 | strong, abstained |

All three inputs were untruncated but abstained because their confidence was
below `0.9`. The useful question is **which requested coding profile actually
meets the requirements**, rather than agreement with an author's expected tier.
This pilot runs the fixed luna/sol comparison, not the strong model retained by
the router's abstention. `gpt-6-astra` was not executed. The historical standard
mapping requested `gpt-6.1-sol`; it is not reinterpreted as a prediction for this
pilot's `gpt-6-sol`. These values are not calibrated probabilities of task success.

In the historical observations from 50 reused by 52, maximum process RSS was about **1.48 GB (1.38 GiB)** in three cold Laya router
processes. Each observation includes loading the model and native session for
one request. It is not warm request latency, Go heap usage, or GPU memory. These
measurements do not establish that the very small memory target has been met.

Before 53's coding outcomes, a [separate prediction](../experiments/task-outcomes/routing-predictions-53.json)
was sealed. INT8 base, one CPU thread and a 0.9 threshold suggested standard,
but a maximum score of about **0.669** caused abstention. The configured fallback
is strong/Astra; no Astra coding attempt occurred. The new standard/Sol6 mapping
applies only to this prediction. One cold maximum-RSS observation was about
**1.50 GB/1.40 GiB**; this is not warm latency, Go heap, GPU memory or a calibrated
coding-success probability.

## Usage

Build the Go binary from the repository root. The executor has no Python runtime
requirement.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskrun ./cmd/riido-taskrun
.cache/bin/riido-taskrun --help
.cache/bin/riido-taskrun --task comment-preview-authority --spec
.cache/bin/riido-taskrun --task repo-keyword-language-guard --spec
```

`--help` and `--spec` launch no model. A default invocation refuses execution.
An actual attempt requires the precommitted plan, pinned public base, trusted
Codex executable and version, trusted Go installation, and an existing local
ChatGPT login. The current
supported version is `codex-cli 0.158.0`; an executable hash that differs from
the plan is refused before inference. A changed version or plan requires a new
plan first.

A deployment-style `-trimpath` binary may have no default Go installation path.
Pass an explicit trusted Go 1.27.1 installation with `--go-root`. The executor
checks the resolved `bin/go`, version, and executable-byte SHA-256 against the
new plan's pins. Without a usable default or explicit root, it refuses before
inference with `trusted_toolchain_unavailable`. It does not follow an arbitrary
host PATH or download another toolchain.

The following example shows **the argument form for a new precommitted plan**.
`NEW_*`, `VERIFIED_*`, `PLANNED_ORDINAL`, and `/absolute/...` are placeholders,
not an executable frozen plan. Match the specification and base to the selected
task, verify the actual hashes, and freeze the new plan first. The example's
`gpt-6-luna` request must match an entry in that new plan too. The private output
directory must not exist yet. An actual invocation may consume account usage.
The output must be beneath an OS temporary directory. A repository or
`AGENTS.md`, `.codex`, or `.agents` in its ancestors causes prelaunch refusal.

```sh
.cache/bin/riido-taskrun \
  --execute \
  --task comment-preview-authority \
  --model gpt-6-luna \
  --reasoning low \
  --base-dir /absolute/laya-tools/internal/taskverify/testdata/base \
  --private-dir /private/tmp/riido-NEW-ATTEMPT \
  --codex-bin /absolute/trusted/codex \
  --codex-sha256 VERIFIED_CODEX_SHA256 \
  --codex-version 'codex-cli 0.158.0' \
  --go-root /absolute/trusted/go-toolchain \
  --plan-file /absolute/private/NEW-FROZEN-PLAN.json \
  --plan-sha256 NEW_PLAN_SHA256 \
  --attempt-ordinal PLANNED_ORDINAL \
  --task-spec-sha256 VERIFIED_TASK_SPEC_SHA256 \
  --auth-source-dir /absolute/private/codex-login \
  --timeout 120s
```

`--auth-source-dir` contains an existing ChatGPT `auth.json`. Do not paste tokens
into arguments or chat. API-key authentication is refused, and only the required
authentication file is copied. User config, rules, skills, MCP, history, and
unrelated inherited environment variables are not carried into the invocation.
Copied authentication stays outside the model workspace and is removed on
terminal cleanup.
The shared background server, apps, hooks, other agents, memories, automatic
skill installation and web search are explicitly disabled for this measurement.

The fresh workspace contains only the declared public files. A macOS permissions
profile allows workspace writes and necessary tool reads, while denying network
access by model-generated commands and access to sibling directories. A real
pre-inference probe must demonstrate the enforcement; otherwise no model attempt
starts. These command permissions are separate from the parent Codex CLI's
communication with its official service. The ordinary checkout is not modified.
Commands use the verified Go installation and a controlled environment with
toolchain/module downloads and CGO disabled. The same installation is passed to
independent verification.

Raw stdout JSONL, stderr, candidate files, and `record.json` stay in the local
private directory. The executor does not upload them to GitHub or Hugging Face.
Public records should contain only reviewed numeric values, enums, and hashes
for these public tasks. Never commit raw traces, authentication, or local
absolute paths.

## Read the result as a person or agent

| Field | Meaning |
| --- | --- |
| `provenance` | `executor_owned_single_attempt` identifies one invocation started and collected by this executor |
| `applied_request` | The explicit request passed to the child CLI; it does not attest to the provider's actual model |
| `observed_model` | Currently `unknown`; do not populate it from the requested model |
| `process_status`, `exit_code` | CLI termination evidence, separate from code correctness |
| `go_version`, `go_binary_sha256` | Version and executable hash of the local Go toolchain checked against the plan before inference |
| `verification_status` | Independent result such as `accepted`, `rejected`, or `verifier_unknown` |
| `usage_summary` | Validly parsed usage, retaining partial observations rather than replacing unknown totals with zero |
| `whole_attempt_usage_complete` | Whether successful termination and process-group cleanup, complete stdout/stderr collection, and core usage for every turn are all present |

A candidate may meet the requirements while its usage remains unknown. That is
not an implementation-failure label. Conversely, exit code zero and complete
usage do not establish task success when the independent checks reject the code.

Input, cached input, and output are reported separately. Cached input is a
subset of input, and reasoning output is a subset of output; neither is added a
second time. Tokens are not converted into money, subscription allowance usage,
or remaining allowance. The recorded wall time measures the local CLI process,
not the separate verification time or remote model GPU usage.

Stdout is limited to 64 MiB and stderr to 1 MiB. Deadline, capture, parsing, and
verification-environment failures remain visible. For example,
`timeout_or_canceled` and `capture_limit` describe execution states;
`unsupported_candidate_shape` describes the narrow verifier's unsupported
scope. These cannot all be labeled as model capability failures.

| `riido-taskrun` exit code | Meaning |
| ---: | --- |
| 0 | The CLI exited zero, independent checks accepted the declared closure, and copied authentication was cleaned up; usage may still be incomplete |
| 1 | A started unsuccessful attempt was recorded; read the JSON process and verification states to distinguish the cause |
| 2 | Prelaunch refusal for arguments, pins, authentication, isolation, or related conditions, or an output error |
| 3 | An attempt was recorded but acceptance, candidate collection, record persistence, or related evidence is unavailable, or authentication cleanup failed |
| 0 from `--help` or `--spec` | Information output succeeded; no model invocation or candidate acceptance occurred |

Agents should not create training labels from exit codes alone. Read process,
verification, and usage evidence separately. Retain prelaunch refusals and
unexecuted entries for the plan rather than replacing them with successful runs.
Validating local plan bytes and ordered attempts does not independently attest
that the plan was published before outcomes.

## What this establishes and what comes next

The command examples explain execution; they do not themselves create model
outcomes. [Experiment 52](../experiments/task-outcomes/RESULTS-52.en.md) separately
records actual executions under its frozen plan. Both profiles have acceptance
and complete usage on the two comment requests, but the Sol behavioral attempt
has unknown whole usage, preventing a complete cost comparison for that request.
Only three requests had been attempted by 52, so that comparison does not establish profile savings
or useful production routing. No new training, model-weight release, production
routing activation, or final 2,400-case scoring took place.

The stopped records for [experiment 50](../experiments/task-outcomes/RESULTS-50.en.md)
and [experiment 51](../experiments/task-outcomes/RESULTS-51.en.md) remain preserved.
A new actual run needs verified available profiles, a separate frozen plan and
fresh output directories.

The small development pilot first checks that success, failure, and unknown
evidence can be recorded correctly. Later, distinct public tasks and paired
profile outcomes can support comparisons of success and usage across every
attempt under the same conditions. Only then can upward routing, downward
routing, and abstention be judged. Read
[task usage and independent verification](task-outcomes.en.md) for interpretation
and the [golden-set design](golden-set-scale.en.md) for required scale and split
separation. The [acquisition plan](golden-set-acquisition.en.md) describes staged
collection and execution budgets.

Current [results 53](../experiments/task-outcomes/RESULTS-53.en.md) provide four
repetitions on one behavioral request. Reported sums are 339,805 input including
280,576 cached input, and 4,890 output including 76 reasoning; do not double-count
subsets. Luna had less total input, but uncached input was Luna 31,425 versus
Sol 27,804. Token totals alone do not establish lower actual cost. Four cumulative
requests do not validate general capability or production routing. The keyword
guard was integrated into runtime with no new fitting, weight release, Codex
policy activation or final scoring. The [120 source candidates](public-go-acquisition-53.en.md)
need independently verified contracts and license/dependency closures before
execution; they are not measured model labels.
