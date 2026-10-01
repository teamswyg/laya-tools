# PDCA55: actual profile comparison on two external Go requests

[한국어](RESULTS-55.ko.md)

Sol6 low and Luna low each attempted **two distinct development requests** once: a full-int64 ordinal API in humanize and a strict lowercase UUID parser. Both humanize candidates were accepted. For UUID, the Sol6 candidate was accepted and the Luna candidate was rejected. All four CLI processes exited zero and reported complete core usage. Successful process exit and fulfillment of requirements are separate observations.

Four profile attempts and the 6 humanize / 218 UUID tests do not create more unique requests. One observation per profile on two requests does not establish a model ranking or monetary savings.

## Conditions fixed before outcomes

The [humanize plan](plan-55-humanize.json) and [UUID plan](plan-55-uuid.json) use different upstream revisions, so a [parent budget](parent-budget-55.json) binds their four slots. Plans were committed at `fe7f3c68df1b299d04a9115081c52fd4d81afd9b` before outcomes. The order was **humanize/Sol6 → UUID/Luna → humanize/Luna → UUID/Sol6**, concurrency 1, with 120 seconds for the model attempt and 45 seconds for separate independent verification. The executor performed no retry, resume or fallback.

A user-requested pause and read-only upstream review took place between slots 2 and 3. The remaining slots ran in the original order after explicit resume. The pause did not refund a failed slot or create a new ledger. This gap and the single observation per profile mean elapsed times are not a controlled general benchmark.

[Final receipts](parent-receipts-55.json) record **4 reservations, 4 durable start markers, 4 terminal receipts and 0 unresolved launch states**. The budget is exhausted with no active execution. All four attempts confirmed child-process-group cleanup and the copied-auth removal state. No slot was refunded. This cap applies to one shared private durable ledger, not the entire host/account or internal provider calls/retries.

The executor was built cleanly from merge `e30da75c5d1ea381924ac5d54e8bcbb9a2f462c8`, whose tree matches CI-passed reviewed head `c8d1cf66413b696b7c0872c01df8155e3cca29d4`. [That CI](https://github.com/teamswyg/laya-tools/actions/runs/36816995319) verifies the framework. The four actual coding attempts ran through Codex CLI on a local Mac using the existing ChatGPT login. No new API key, paid cloud job or endpoint was created.

## Requirements and Go conditions received by the models

Actual v2 stdin was the entire frozen `Spec.Prompt`, including mutable files, legacy behavior, required format and supported verifier scope. Humanize received 3,150 bytes and UUID 5,745 bytes, bound to the Prompt SHA in each plan. These v2 IDs clarify two previously prepared logical requests. Versioning added no new logical requests; both were first actually attempted here.

| Source request | Pinned revision | Original language | Go executable | Contract scope |
| --- | --- | --- | --- | --- |
| humanize full-int64 ordinals | `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e` | Go 1.21 | Go 1.27.1 | New API signs, endpoints and suffixes; legacy Ordinal preserved |
| UUID lowercase parser | `2d3c2a9cc518326daf99a383f07c4d3c44317e4d` | Implicit Go 1.16 without a directive | Go 1.27.1 | Lowercase 36-byte grammar, zero UUID on error; legacy APIs and state preserved |

Independent verification used pinned original `go.mod` bytes and module/recipe hashes. Candidate module settings and candidate-authored tests did not determine success. The UUID gate supports one appended function after the exact original or trusted formatter prefix. Extra helpers, calls through aliases, concurrency and global writes are outside support and yield `verifier_unknown` even for otherwise correct output. Unknown is not relabeled as behavioral failure. [Evaluation conditions](../../docs/fair-upstream-comparison.en.md)

Original MIT and BSD-3 LICENSE files were separately validated and copied from trusted bytes. Candidate LICENSE contents are outside success-label assessment; acceptance does not certify downstream attribution compliance. Finite contract acceptance does not establish complete upstream correctness or rights to train and distribute model weights.

## Actual observations

| Order | Request | Requested profile | Main ms | Verification ms | Input | Included cached input | Output | Included reasoning | Independent outcome |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | humanize | Sol6 low | 76,404 | 4,170 | 59,824 | 44,928 | 1,005 | 83 | Accepted; 6 terminal passes |
| 2 | UUID | Luna low | 98,471 | 5,852 | 152,941 | 129,536 | 2,126 | 0 | Rejected; pass count not aggregated |
| 3 | humanize | Luna low | 20,836 | 4,646 | 45,866 | 39,936 | 649 | 0 | Accepted; 6 terminal passes |
| 4 | UUID | Sol6 low | 43,474 | 5,421 | 82,301 | 66,304 | 1,347 | 231 | Accepted; 218 terminal passes |

Total input **340,932** includes cached input **280,704**, and output **5,127** includes reasoning **314**. These are subsets, not additional tokens. Sum of model-attempt time **239,185 ms** and separate verification **20,089 ms** excludes parent preparation, native predictions and the pause.

On humanize, the Luna attempt used less reported input/output and less elapsed time while passing the same contract. On UUID, the Luna candidate was rejected and the Sol6 candidate accepted, with different observed elapsed times. These observations show that outcomes can vary by request; they are not general profile rankings or success-probability labels. Repair or escalation of the failed candidate was not executed. Subscription consumption, billing rates and internal provider calls/retries remain unknown, so monetary savings are not claimed.

[Public numeric records and hashes](results-55.json) bind the raw bytes of [attempt 1](attempt-55-01.json) and its receipt; the remaining records are [attempt 2](attempt-55-02.json), [attempt 3](attempt-55-03.json) and [attempt 4](attempt-55-04.json). Requested profiles are explicit, while provider-attested served identity is `unknown` for all four. Raw model traces, credentials, private code, local paths and weights are excluded.

### Why the UUID Luna candidate was rejected

A source-only review after the terminal outcome found that a valid nonzero prefix followed by malformed hex or a later separator returns a **partially populated UUID** with an error. The disclosed requirement is zero UUID on every error, so this violates the contract. This analysis did not rerun the model or repair and replace the recorded outcome.

The raw verifier counter `independent_tests: 0` occurs because a nonzero independent `go test` exit returns before terminal-pass aggregation. **It does not mean no tests ran or all 218 failed.** Public aggregation represents the pass count as `null`. The candidate passed isolation, supported-shape and formatting gates before the behavioral tests rejected it.

## Laya observations made before coding

[Pre-outcome predictions](routing-predictions-55.json) were committed at `1b87612d7848ada919356c9dbe5f1e04f2993ea8`. They route the entire original Prompt bytes through unchanged INT8 base, one CPU thread, threshold 0.9 and a 512-token budget.

| Request | Full input bytes | Standard score | Truncated / abstained | Cold process ms | App observation ms | Maximum RSS bytes |
| --- | ---: | ---: | --- | ---: | ---: | ---: |
| humanize | 3,150 | 0.502708097 | Both true | 1,595 | 1,551.182 | 1,544,044,544 |
| UUID | 5,745 | 0.477323446 | Both true | 1,337 | 1,304.806 | 1,581,891,584 |

Both inputs were truncated within the 512-token total budget, which also includes question/header/options tokens, and Laya abstained. Configured fallback preserves strong/Astra; the actual comparison followed the separately fixed Sol6/Luna order. Astra was not executed and live routing was not activated. A full-Prompt input hash does not mean the encoder read the entire Prompt. Scores are not calibrated coding-success probabilities.

Maximum RSS of about 1.54 / 1.58 GB and elapsed time are **cold whole-process CPU** observations including native loading. They are distinct from Go heap, GPU memory, remote coding-model resources and warm latency. These observations do not meet the ultra-small-resource goal.

## Upstream review and next direction

The pinned Laya [router preset](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/laya/presets.py#L177) asks about difficulty, domain, tool needs and sensitivity. Its [routing example](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/examples/29_presets_model_router.py) applies threshold policy to three authored inputs. The [model-routing benchmark](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/research/scripts/bench_apps.py#L243) classifies domains for 399 inputs; it does not supply completion or usage labels for particular Codex profiles.

Our assessment is that this structure can inform scoring of short claims and hints, while model-selection utility still needs actual completion outcomes and failure costs. Because both full Prompts truncated and abstained, **short claim inputs and Go rule / small-coefficient baselines without an encoder** are a separate PDCA56 plan. Shortening this input later will not turn this experiment into a routing success. That new experiment has no measured performance yet.

## Scale and eligibility

Actual cumulative development observations cover **7 distinct requests, 20 CLI records, 5 code families and 3 repositories**. The compiled logical-request count did not grow; actual observation coverage increased from 5 to 7 when these two external requests were first attempted. Historical failures, timeouts, support errors and the [original 120-candidate inventory](../../docs/public-go-acquisition-53.en.md) remain unchanged. Those 120 candidates are not 120 labels or executable contracts.

Both requests have already been read during development. **Training runs, training-eligible labels, protected-final-eligible requests, new weight releases and production-policy changes are all zero.** Tests, repetitions, translations and sibling tasks do not count as distinct final requests.

After the initial 24 development probes, expand to **120 or more** independently verifiable development requests across repositories and **240 or more** for training-signal assessment. Separately acquire and protect **at least 2,400 distinct final requests per claimed domain**. These stages have not been achieved yet. Follow the [scale guide](../../docs/golden-set-scale.en.md) and [acquisition plan](../../docs/golden-set-acquisition.en.md) for data and group separation. This cycle performed no fitting, calibration, new model release or protected-final scoring.
