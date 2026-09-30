# Experiment 46: over 2,400 examples per role, usefulness gates failed

All ten registered candidates were fitted on **4,456 training and 2,599 validation examples**. This is a larger comparison than the initial 24-task pilot, but the current claim head on 16 numeric features failed its usefulness gates. Completed training is distinct from improved performance. No policy was activated and no new model weights were released on Hugging Face.

| Validation policy | Total pages | Auxiliary selections | Tasks with a target in top 10 | Equal-repository Hit10 |
|---|---:|---:|---:|---:|
| Baseline | 18,023 | 0 | 1,014 | 41.273% |
| Always helper | 18,179 | 2,599 | 998 | 40.549% |
| Gap heuristic | 17,967 | 720 | 1,010 | 41.091% |
| Learned selection, seed 1729 | 18,010 | 55 | 1,013 | 41.245% |
| Learned selection, seed 2718 | 17,978 | 1,248 | 1,007 | 40.986% |

Pages mean **pages of 20 candidates until the first target file**. They are not measured money, LLM tokens or task completion time. Auxiliary selections and `CounterfactualWork` replay a policy over two stored search outcomes; they do not measure time saved by actually skipping the helper. The [aggregate report](results-46.json) retains every candidate, control, epoch and repository result.

Both seeds selected penalty 4 under the fixed procedure. Page reductions against baseline were **0.072% / 0.250%**, below the required 5%. Pooled and repository-average Hit10 declined. With seed 2718, cryptography pages increased from 1,925 to 2,140, or 11.17%, exceeding the 5% repository regression cap. A small page reduction is not a pass. The gap heuristic also reduces quality, so it is not evidence for deployment.

A post-hoc calculation on actual cost rows found that **an oracle knowing all labels and selecting every beneficial case could save at most 561 pages, or 3.113%**. The 5% gate requires at least 902 fewer pages. Even this optimistic bound ignores quality and call constraints and cannot reach the gate: the helper itself lacked enough headroom on this input. The learned policies' 13 / 45 saved pages also fall far below that bound. The [headroom diagnostic](diagnostic-46.json) explains already observed validation; it is not a deployable oracle or a new success gate. The preregistered gate is unchanged and the experiment remains failed.

## Data and golden-set scale

| Stage | Tasks | Purpose |
|---|---:|---|
| Fixed training coverage | 7,335 | Includes unavailable and pending members |
| Actual eligible training | 4,456 | Fit coefficients on 16 generic numeric features |
| Fixed validation coverage | 5,686 | Original development-comparison scope |
| Actual eligible validation | 2,599 | Select epoch, penalty and threshold |
| Protected final | 2,402 | No final outcomes read or scored |

Eligible validation counts are Lightning 287, Google Cloud Python 672, NumPy 679, pandas 634 and cryptography 327. They exceed 100 per repository and 2,400 per role. Repeated seeds, epochs or configurations of the same question do not add independent examples. These counts do not prove semantic independence or freedom from pretraining contamination. Only about 45.7% of fixed validation is currently usable, limiting claims about the full development distribution.

Always-helper validation has 116 wins, 283 losses and 2,200 page ties. Informative cases can be much fewer than the full count. At penalty zero, 3,722 training and 2,200 validation rows have zero weight but remain in coverage. Positive penalties make tied outcomes informative about avoiding helper cost. Different penalties change labels and weights, so a smaller weighted loss is not a comparable success rate.

Actual Go verification joined numeric inputs to [source review 47](../path-cost-data/SOURCE-REVIEW-47.en.md). The 221 ambiguous rows were excluded from fitting. The 7,144 source-root candidates and 7,055 eligible training-plus-validation examples count different stages. Source text/code, patches, query/path strings, repository/task IDs and vocabulary are not coefficient inputs. Conditional review for this local numeric pilot is not public-release approval.

## Reproduction and memory

The [plan](plan-46.json) was committed before costs were inspected. Runner code `7d9c64d9dd6e949f6d80e5f5d914b3515c91a711` was committed first, and a binary built from that clean revision prepared the actual inputs. The [input seal](input-46.json) was committed separately as `d8a4e29` before fitting with **the same binary**. Rebuilding after the seal commit changes the runner revision and intentionally fails the pin. There is no override for changed inputs or inadequate counts.

All five penalties × two seeds and all 100 epochs of every fit are recorded. Equal validation NLL selects the earliest epoch. FP32 describes coefficient storage; features, shadow weights, gradients and loss accumulation use float64. Cost-weighted scores are not calibrated success probabilities.

An ordinary replay produced **11 byte-identical JSON files**, including aggregates and all models. Ordinary and race cost-47 replays matched both the original aggregate and private examples; six source-review outputs also matched offline. Code race tests, vet, formatting and redacted secret checks were run separately. The frozen cost input has zero fallbacks. This trainer requires complete paired outcomes and fails closed on incomplete auxiliary execution rather than inventing normal training labels.

One isolated warm offline replay of **all input verification, ten fits, policy evaluation and file writing took 1.45 seconds with about 108.6MiB peak RSS**. No GPU was used. This is one full training-runner observation, not single-inference latency or a cold/warm distribution. macOS process RSS is distinct from Go heap pprof; pprof and isolated inference were not measured here. See the [resource observation and scope](resources-46.json).

Model JSON files are 1,956–1,985 bytes; each FP32 coefficient payload alone is 64 bytes. The decoded model, arrays, catalogs, caches and search pipeline are not 64 bytes. Model bodies remain in private cache; Git contains explanations, aggregates and hashes.

## Direction for a separate experiment

Preserve this family as a failed experiment. Do not change penalties, thresholds or gates on the same validation data or retroactively choose another candidate as a winner. First analyze **attainable helper headroom and training-only evidence for features that separate benefit from harm**, then precommit a new hypothesis and the scope of already observed validation. Insufficient representation and conflict between page cost and quality remain hypotheses.

INT8/ternary compression is not evidence of improved usefulness before a useful FP32 signal exists. Later stages must measure actually skipped helper execution, context size, LLM calls/retries and completion. Model escalation/de-escalation, repository selection and dependency-aware decomposition each need separate golden sets; file-search tasks do not supply their ground truth.
