# Experiment 50: pilot stopped before coding-model launch

[한국어](RESULTS-50.ko.md) · [Numeric record](results-50.json) · [Precommitted plan](plan-50.json)

**No actual coding-model process started.** The first executor invocation returned
exit code 2 with `trusted_toolchain_unavailable` before passing the public task to
the main model process. This is not a model failure or difficulty result. It
revealed an executor packaging defect, so the remaining fixed comparisons stopped.

## Separate planned entries from execution

The plan compared three existing public development tasks with two explicit
requested profiles. `luna-low` requests `gpt-6-luna` with `low`; `sol-low` requests
`gpt-6.1-sol` with `low`. It allowed at most six sequential coding CLI main
invocations. The first scheduled entry was `comment-preview-authority` / `luna-low`.

| Item | Observed count |
| --- | ---: |
| Planned coding CLI main invocations | 6 |
| Prelaunch refusals | 1 |
| Coding CLI main invocations started | 0 |
| Remaining planned entries not executed | 5 |
| Distinct public development tasks | 3 |
| Code families | 2 |
| Actual coding-model outcomes | 0 |
| Separately executed prior Laya CPU predictions | 3 |

The five remaining entries are `not_executed_integrity_stop`, not successful or
unsuccessful model outcomes. Comparing two profiles does not turn one task into
two distinct tasks. CLI invocation counts are also separate from internal model
requests and provider retries.

## What failed

The packaged executor's source revision was
`6ae4f1a38ed3d74040f9b1ca7947b3932bb1a5d8`. In this deployment-style `-trimpath`
build, the executor could not resolve a trusted Go installation through
`runtime.GOROOT()`. It refused while preparing the narrowly allowed toolchain
read permissions, before the main coding-model process. The packaging defect
was reproduced separately offline.

Source tests and a help invocation had not established that the packaged binary
could resolve its toolchain. The required fix explicitly validates a trusted Go
installation and passes it consistently to the executor and independent verifier.
After fixing and verifying this behavior, a restart requires **a separately
precommitted plan 51**. Results from the stopped plan 50 must not be overwritten
with successful outcomes or treated as a silent retry using modified code.

The observation tool reported approximately **0.52 seconds** for the executor
invocation. This is an approximate external observation, not a coding-model
process timing or verification record. No main-model start or terminal record
existed. The inspected private directory had mode `0700`, and copied
authentication was absent after refusal. These are limited post-refusal
observations; they do not prove model execution, completion, or a successful
model-command isolation probe.

## Did Laya itself run?

**The three prior CPU predictions actually ran.** They are separate from the
coding-model comparison. They used the unchanged pinned `base-int8-v2` checkpoint
with one CPU thread and threshold `0.9`. The
[original prediction record](routing-predictions-50.json) remains intact.

| Public development task | Author's hypothesis | Laya suggestion | Largest probability | Applied tier | Cold process wall time |
| --- | --- | --- | ---: | --- | ---: |
| `comment-preview-authority` | fast | standard | 0.479528 | strong, abstained | 0.98 s |
| `comment-budget-period` | fast | fast | 0.486019 | strong, abstained | 0.70 s |
| `catalog-min-context` | standard | standard | 0.591143 | strong, abstained | 0.72 s |

All inputs were untruncated, but low confidence caused abstention. These are not
calibrated task-success probabilities. The pilot was the precommitted luna/sol
comparison; it did not plan to execute the retained strong model, `gpt-6-astra`.

Individual maximum Laya router process RSS values were `1,486,143,488`,
`1,482,211,328`, and `1,486,192,640` bytes, approximately **1.48 GB (1.38 GiB)**.
Each is one cold process including native model and session loading. These
observations are not warm request latency, Go heap, GPU memory, or remote coding
model resources. They do not establish that the very small memory target has
been achieved.

## Results this cannot support

Coding-task usage was not observed and remains `null`. It is not zero, and no
zero-cost or 100%-savings claim follows. Task acceptance was not checked and
remains `not_assessed`. Expected difficulty does not become an observed success
label. The refusal is not evidence of a smaller model's capability failure or
under-routing.

Three development tasks in two families prepare a measurement procedure. The
user's target of at least **2,400 distinct final requests per evaluated domain**
remains unchanged. This record cannot replace that evaluation or establish
actual usage savings, model difficulty calibration, or population success rates.
No new training, threshold adjustment, protected-final scoring, model-weight
publication, or production routing activation took place.

The next step is to verify toolchain resolution and isolation in the deployment
build, then check evidence binding with a small actual comparison under a new
plan. Independent public tasks can be acquired and scaled after that procedure
works. Read the [usage guide](../../docs/task-execution.en.md),
[usage and independent verification guide](../../docs/task-outcomes.en.md), and
[golden-set scale and split principles](../../docs/golden-set-scale.en.md).

This public record contains no raw traces, authentication, local private paths,
or model weights. It binds to plan SHA-256
`186a59fd11f09cce4e1e26fe3b79919a7bf1fc65db19caa25a3d2aa7aaa1a967` and prior
prediction SHA-256
`bee88a6c02ab0d9af23b57649b5d324aebd59d3e825b78f8868e823d75773dac`.
