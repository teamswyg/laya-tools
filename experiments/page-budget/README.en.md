# Budget-calibrated claim experiment 15

**Both primary and replication fail.** Validation-only threshold selection reduces helper calls, but the learned heads read more pages than a simple score-gap rule under the same procedure. No learned policy is promoted.

## What changed

Reuse all 24 immutable experiment-14 heads without refitting or changing weights. Propose the helper when `score >= threshold`. Select that threshold only on each model's original validation repository; target outcomes inform calibration, never runtime scoring.

Precommit validation call caps of 25/50/75/90%. Enumerate distinct score thresholds plus disabled; among policies making at most floor(validation count × budget/100) calls, minimize total pages. Break ties in favor of fewer calls; disable when there is no benefit. Never split tied scores. Apply the selected threshold unchanged to the held-out repository.

The cap constrains the **validation rate**, not a hard runtime quota on new repositories. Report evaluation overruns and fail them. For example, the heuristic's 50% setting calls 1,492/2,948 evaluation queries, exceeding 50%.

The nonlearned control favors smaller raw-BM25 top-two gaps: score is feature 9 negated, `-(top-second)/(1+top)`. It is available before auxiliary search and needs no learned weights. Calibrate it with the same validation data, caps and selection procedure.

## Primary 90% setting

| Policy | Actual helper calls | Total pages | Recall@10 |
|---|---:|---:|---:|
| Raw BM25 | 0 | 9,488 | 74.83% |
| Always helper | 2,948 | 6,287 | 81.38% |
| Score-gap rule | 2,548 | 6,323 | 80.97% |
| FP32 / 1729 (primary) | 2,558 | 7,033 | 80.50% |
| FP32 / 2718 (replication) | 2,543 | 7,079 | 80.50% |
| Ternary PTQ/STE, both seeds | 2,595 | 6,905 | 80.66% |

The gap rule trades 400 fewer calls for 36 additional pages versus always-helper. This is a tradeoff, not unconditional improvement; it fails the precommitted no-extra-pages gate. Learned policies must also have no more calls and pages than the calibrated gap rule, with one strict improvement. They fail too. No gate relaxation or winning-seed selection follows scoring. All 36 aggregate conditions are reported and all fail.

## Evidence and limitations

There are 2,948 distinct questions, not 108 fold settings or 36 independent query sets. These are previously observed documentation proxies from three repositories. All code candidates are visible across folds, and epoch and threshold selection share validation data, permitting validation overfitting. The corresponding evaluation labels inform neither selection step. This is not 2,400 independent real-user questions, actual agent success or LLM billing evidence.

Inputs are the fully hash-verified package from HF revision `4e0addaa24d407bb11282f193e039c943a08bd0a`. The entire report replayed byte-identically. Source checks, retrieval, calibration and evaluation took 1.79 s wall, 1.41 s user CPU and 98,975,744 bytes peak RSS on Apple M4 Pro, without GPU. Tests cover tied scores, shifted call rates and unchanged thresholds when held-out labels change.

This evidence does not favor producing more variants of the same 12-feature linear model as the next priority. Keep the simple rule as a strong control; require independent motivation for new features or model structures. Whether the heuristic's page/call tradeoff is useful needs actual execution and verification-cost measurements.

```sh
go run ./cmd/riido-pagebudget --models /path/to/verified-experiment-14-package > results.json
```

The pinned projection/ZIPs from experiment 07 are also required. Supply the 31 verified package files without HF download metadata. No new weights were produced, so no duplicate HF model archive is published. Git stores aggregate results, selected thresholds and verification code, not weights, source text, queries or per-row labels.
