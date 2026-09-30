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

## Scope of this development pilot

[Precommitted plan 50](../experiments/task-outcomes/plan-50.json) stopped before
coding-model launch when the packaged executor could not resolve its Go toolchain.
[Those results](../experiments/task-outcomes/RESULTS-50.en.md) are retained; modified
code does not overwrite that plan with successful outcomes.

The separately [precommitted plan 51](../experiments/task-outcomes/plan-51.json)
includes the toolchain repair and preserves the same task/profile selection,
order, threshold, and limits. This guide uses plan 51, committed at
`14fdad230ad6cebbee194b09ed8b6df7916a19cc`. It compares three
existing public development tasks with two explicit requested profiles.
`luna-low` requests `gpt-6-luna` with `low`; `sol-low` requests `gpt-6.1-sol` with
`low`. The plan fixes the order and allows at most **six Codex CLI invocations**,
one at a time, with a 120-second main-attempt deadline and a separate 45-second
verification deadline.

This does not mean six internal LLM calls. One CLI invocation may contain
multiple model requests and provider retries; those counts are currently
`unknown`. The executor performs no retries, resume, or fallback. A started
failure or timeout counts toward the pilot's limit. The command below owns only
one attempt, so the maintainer must also enforce the complete plan's count and
order and retain entries that were not executed.

| Task ID | Requirement checked independently | Declared file scope |
| --- | --- | --- |
| `comment-preview-authority` | Make the exact comment change stating that a repository suggestion does not authorize execution | One file, `pkg/reporouter/router.go`; exact change and gofmt |
| `comment-budget-period` | Replace the budget-period comment with the specified sentence | Three pinned files; exact change, gofmt, and pinned tests |
| `catalog-min-context` | Add optional `MinContext`, reject negative values, apply minimum-context eligibility, and preserve existing behavior | Nine pinned files; boundary/JSON contracts and pinned catalog/planner tests |

The shared public base is revision
`6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`, represented by
`internal/taskverify/testdata/base`. The executor checks hashes of its files,
LICENSE and NOTICE, task specifications, and independent contract tests. Added
files and runtime caches outside the declared scope are `unassessed`.
`accepted` establishes compliance **within that task scope**, not approval of the
entire repository or arbitrary code.

These three tasks come from two code families. Running two profiles does not
turn them into six distinct tasks. They do not establish population success
rates or sufficient training labels and cannot replace the
[2,400-case final evaluation design](golden-set-scale.en.md).

## What Laya predicted before coding outcomes

The [recorded predictions](../experiments/task-outcomes/routing-predictions-50.json)
used the unchanged pinned Laya INT8 checkpoint on CPU with one thread and the
default threshold of `0.9`. No threshold adjustment or training on these three
tasks took place. Plan 51 references these unchanged predictions; it does not
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
the router's abstention. `gpt-6-astra` is not executed in this plan. These values
are not calibrated probabilities of task success.

Maximum process RSS was about **1.48 GB (1.38 GiB)** in three cold Laya router
processes. Each observation includes loading the model and native session for
one request. It is not warm request latency, Go heap usage, or GPU memory. These
measurements do not establish that the very small memory target has been met.

## Usage

Build the Go binary from the repository root. The executor has no Python runtime
requirement.

```sh
mkdir -p .cache/bin
go build -trimpath -o .cache/bin/riido-taskrun ./cmd/riido-taskrun
.cache/bin/riido-taskrun --help
.cache/bin/riido-taskrun --task comment-preview-authority --spec
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
checks the resolved `bin/go`, version, and executable-byte SHA-256 against plan
51's pins. Without a usable default or explicit root, it refuses before
inference with `trusted_toolchain_unavailable`. It does not follow an arbitrary
host PATH or download another toolchain.

The following example **preserves the invocation used for plan 51**. That plan
[stopped after two attempts](../experiments/task-outcomes/RESULTS-51.en.md) and must
not be resumed. Freeze a separate plan before a new actual run. Executable pins
refer to the observed Mac installation; another installation requires newly
verified tool hashes too.
`/absolute/...` values are placeholders; supply your own absolute paths. The
private output directory must not exist yet. Unlike reading an example, this
invocation may consume account usage.
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
  --codex-sha256 788a818fbb9596869c7a487554507cb8bdca17584b8671112b23f9e225ba35c8 \
  --codex-version 'codex-cli 0.158.0' \
  --go-root /absolute/trusted/go-toolchain \
  --plan-file /absolute/laya-tools/experiments/task-outcomes/plan-51.json \
  --plan-sha256 acf4786f7330529e04af12d4f7dbf40b0bbc2b6a9c527ecb879aee9d8c1510b0 \
  --attempt-ordinal 1 \
  --task-spec-sha256 a01f1c95e818511d6bcac35a0eb20b0b6c943a73b4fe7652a550bb0036dbd747 \
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

The examples in this guide explain execution; they do not report actual coding
model outcomes. Keep the [stopped plan 50 results](../experiments/task-outcomes/RESULTS-50.en.md),
[new plan 51](../experiments/task-outcomes/plan-51.json), and
[prior predictions](../experiments/task-outcomes/routing-predictions-50.json)
separate from post-execution records. This stage performs no new training,
model-weight release, production routing activation, or final 2,400-case scoring.

[Results 51](../experiments/task-outcomes/RESULTS-51.en.md) preserve two attempts
and the measurement-compatibility stop. The example above documents syntax;
it does not authorize resuming that stopped plan. A new actual run needs verified
available profiles, a separate frozen plan and fresh output directories.

The small development pilot first checks that success, failure, and unknown
evidence can be recorded correctly. Later, distinct public tasks and paired
profile outcomes can support comparisons of success and usage across every
attempt under the same conditions. Only then can upward routing, downward
routing, and abstention be judged. Read
[task usage and independent verification](task-outcomes.en.md) for interpretation
and the [golden-set design](golden-set-scale.en.md) for required scale and split
separation.
