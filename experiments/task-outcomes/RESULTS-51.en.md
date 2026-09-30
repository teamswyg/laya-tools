# Experiment 51: two actual attempts recorded, comparison stopped

[한국어](RESULTS-51.ko.md) · [Numeric record](results-51.json) · [Precommitted plan](plan-51.json)

**Two owned Codex CLI attempts were connected to their results, but no comparison
of the two profiles' capability was possible.** The first `luna-low` attempt
modified code and recorded usage but did not meet the exact comment-change
requirement. The second `sol-low` attempt returned a service error indicating
that the explicit profile was unsupported and exposed a gap in initial-error
event handling. The plan's integrity stop left the remaining four entries
unexecuted.

## Planned scope and actual execution

[Plan 50](RESULTS-50.en.md) stopped before its main coding-model launch because
of executor packaging. Its record was preserved. After repairing toolchain
resolution, the separately precommitted [plan 51](plan-51.json) kept the same
tasks, profiles, order, threshold, and limits.

| Scope | Count |
| --- | ---: |
| Planned Codex CLI main invocations | 6 |
| Owned CLI main invocations started and collected | 2 |
| Remaining planned entries not executed | 4 |
| Distinct public development tasks scheduled | 3 |
| Code families scheduled | 2 |
| Task actually executed | One: `comment-preview-authority` |
| Requested profiles producing a modified candidate | One: `luna-low` |
| Independently accepted candidates | 0 |

The invocations used different requested profiles on the same public comment
task. They are neither two independent tasks nor two capability outcomes.
Internal model-request and provider-retry counts remain unknown.

## First attempt: process success is not task success

The [first original numeric record](attempt-51-01.json) is ordinal 1,
`luna-low`: an explicit `gpt-6-luna` / `low` request. The CLI exited **0** and
recorded `wall_millis` **17,401 ms**. Process-group termination and copied-auth
cleanup were confirmed. Requested settings were passed directly to the child
CLI, but no provider-attested model identity was available: `observed_model`
remains `unknown`.

The candidate changed the comment but **omitted the required final period**.
The precommitted verifier checks an exact-change contract rather than approximate
semantic similarity, so it returned `rejected` / `not_exact_comment_change`.
The contract was not relaxed after observing the result, and this was not
reclassified as success.
This comment task uses a static contract check without executing candidate code.
The verification fields `independent_tests: 0` and `execution_isolated: false`
indicate that no candidate behavioral tests ran; they are separate from the
executor's permissions probe.

| Completely observed usage in the first attempt | Tokens |
| --- | ---: |
| Input | 53,150 |
| Cached input | 43,776 |
| Output | 338 |
| Reasoning output | 0 |

Complete usage evidence and task acceptance are separate. Cached input is
already included in input, and reasoning is included in output; neither is
added again. These numbers are not money, subscription allowance consumption,
or remaining quota. They also do not represent only the brief comment prompt
or pure generation latency. CLI process time and independent verification time
remain separate.

This is one observation that the requested configuration did not satisfy this
exact comment contract. It does not establish a smaller model's overall failure
rate, difficulty calibration, or performance in other repositories.

## Second attempt: unsupported profile and an early diagnostic

The [second original numeric record](attempt-51-02.json) is ordinal 2,
`sol-low`: an explicit `gpt-6.1-sol` / `low` request. The CLI exited **1** and
recorded `wall_millis` **1,798 ms**. The private error text is not published;
only these observations are retained:

- `provider_error_matches_unsupported_pattern`: `true`
- `provider_error_mentions_chatgpt_account`: `true`
- `authentication_error_observed`: `false`
- `timeout_error_observed`: `false`
- `observed_model`: `unknown`

This applies to the observed login, client, and request combination. It does not
mean `gpt-6.1-sol` is unsupported for all accounts or establish the specific cause
of availability failure. Official documentation says
availability depends on rollout, sign-in method, and client. See
[OpenAI's model documentation](https://learn.chatgpt.com/docs/models).

The trace contained `thread.started` followed by an `item.completed` error item
before any `turn.started`. The existing summarizer could not handle this initial
diagnostic shape, yielding `summary_status: parse_failed` with
`invalid_lifecycle`. Core usage was not observed and remains **`null`**, not
zero tokens or a free invocation.
The complete control sequence was `thread.started` → `item.completed` →
`turn.started` → `error` → `turn.failed`. The initial completed item preceded the
turn, exposing parser incompatibility; the later failed turn remains a separate
execution observation.

Candidate files remained at the base, so independent verification rejected the
unchanged candidate. That verdict establishes only that the declared files were
unchanged. It must not become a label that an unsupported profile tried the
coding task and failed through insufficient capability.

Profile availability and parser compatibility are separate issues. Correcting
the early diagnostic parser would not improve the existing candidate or create
new observed usage. Original attempt records must remain intact; any reanalysis
needs a separately versioned record referencing them.

After the stop, parser revision `67c583173d0653aa4d5732ffeadd87f4fd124099`
read the same private trace in a [separate reanalysis](reanalysis-51-02.json).
It identified one startup error item, one failed turn and one top-level error.
The trace-byte hash matches the original owned record. All usage fields remain
`null`; no model was invoked again and the original `parse_failed` record remains.

## Why the remaining four entries stopped

The first candidate's rejection alone did not stop the fixed comparison. The
second attempt exposed profile availability failure and an event-compatibility
gap, triggering the measurement-integrity stop. Four budget-comment and
minimum-context entries remain `not_executed_measurement_integrity_stop`. No available model
was silently substituted, and failed requests were not retried to retain only
better results.

The profiles therefore cannot be compared for task success, usage, or time under
the same conditions. First-attempt usage is known, second-attempt usage is
unknown, and only the first requested profile produced a modified candidate.
These results do not establish useful upward/downward routing, token savings,
money savings, or calibrated task difficulty.

## Executor and prior Laya evidence

Executor source revision was
`14fdad230ad6cebbee194b09ed8b6df7916a19cc`. Its clean-source `-trimpath` binary
SHA-256 was
`e7164ebe838a85a8d3fe87a82c5a32ceed65c98e5aa1b0643cb34f682be5e65c`.
This is the `riido-taskrun` executable hash. Separately, the validated Go
toolchain was Go 1.27.1, with pinned executable SHA-256
`132b69336a1f809932a8a20b0201dbbb980e86e3a323ae32e893639d83d71598`.
The public task base was `6b2c1bdfd38bea5a9a2c11433ffcfcb94c4143a3`.

Plan 51 reused [the three prior Laya CPU predictions](routing-predictions-50.json)
unchanged. All three abstained at threshold `0.9`; each cold router process had
maximum RSS of approximately **1.48 GB (1.38 GiB)**. These observations include
native model/session loading and are not Go heap, GPU memory, or coding-model
resources. They are not three new predictions, new independent tasks, or evidence
that the very small memory target was achieved.

## Next action and remaining goal

The initial-diagnostic parser is fixed and passed authored positive, negative,
input-bound and redaction checks, plus full Go race and vet checks. The packaged
standalone verifier accepted an authored budget-comment candidate with 16 isolated
test events; this is tool verification, separate from model outcomes. Next,
confirm explicit profiles that can actually run and freeze **a separate pilot
plan** before restarting. New attempts need new records, preserving the originals.

Current evidence concerns measurement preparation on three development tasks
in two families. The target of at least **2,400 distinct final requests per
evaluated domain** remains intact. Once executor and profile verification work,
distinct development tasks and independent sources can be expanded. No new
training, weight publication, final scoring, calibration, threshold adjustment,
or production routing activation took place.

Read the [execution guide](../../docs/task-execution.en.md),
[usage and independent verification guide](../../docs/task-outcomes.en.md), and
[golden-set scale and separation](../../docs/golden-set-scale.en.md).
Plan SHA-256 is
`acf4786f7330529e04af12d4f7dbf40b0bbc2b6a9c527ecb879aee9d8c1510b0`.
Public records omit authentication, private error text, local absolute paths,
raw execution traces, and model weights.
