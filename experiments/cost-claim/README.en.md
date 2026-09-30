# Call-cost-aware claim learning 20

**Changing the objective improves on the previous learned model, but fails the precommitted promotion gate.** Primary pages fall from 7,033 to 6,337 and helper calls from 2,558 to 2,470. Relative to the cheap score-gap rule, it saves 78 calls but reads 14 extra pages. A beneficial overall trade-off is not established.

## Controlled change

Experiment 14 labels the sign of baseline-minus-helper pages and weights by its absolute value. Its 2,470/2,948 page ties have zero loss weight, so unnecessary calls on ties are not penalized. Subtract a call penalty λ: net benefit = saved pages − λ. Positive λ gives page ties a nonzero negative-target weight.

Freeze λ in {0, 0.25, 1, 4, 16}. These are **experimental page-equivalent penalties**, not measured CPU, token or billing conversions. Calibrate a threshold under the validation-only 90% call cap, then select the penalty with fewest validation pages, breaking ties by fewer calls and then smaller λ. Held-out outcomes never select the penalty.

Keep all 12 features, trainer/hyperparameters, three repository folds and two seeds unchanged. No new features or ternary compression in this ablation. Train 3 folds × 2 seeds × 5 penalties = **30 FP32 numerical heads** in Go on CPU. These are experimental conditions of one 12-coefficient architecture, not 30 distinct architectures.

| Policy | Seed | Helper calls | Pages | Recall@10 | Gate |
|---|---:|---:|---:|---:|---|
| Raw search | — | 0 | 9,488 | 0.7483 | control |
| Always helper | — | 2,948 | 6,287 | 0.8138 | control |
| Score-gap rule | — | 2,548 | 6,323 | 0.8097 | fail |
| Zero-penalty learned | 1729 | 2,558 | 7,033 | 0.8050 | fail |
| Zero-penalty learned | 2718 | 2,543 | 7,079 | 0.8050 | fail |
| Validation-selected cost | 1729 | 2,470 | 6,337 | 0.8094 | fail |
| Validation-selected cost | 2718 | 2,489 | 6,338 | 0.8094 | fail |

Each row counts the same **2,948 distinct docstring queries** once over 3,009 candidates. Baseline-first policies preserve Recall@1. Both seeds select λ=1, 4, 0.25 for held-out etcd, test-infra, lxd respectively. Two seeds are not independent data replications.

Retain experiment 15's 90% gate: pages no greater than always-helper, observed calls ≤90%, Recall@1/10 at least raw; learned policies must also Pareto-dominate the gap rule in pages/calls with one strict gain. Selected policies need 50/51 extra pages versus always-helper and fail to dominate the gap rule. Do not relax the gate after results.

## Reproduction, resources and limits

Commit the [plan](plan-20.json) before fitting. Keep fixed repository splits, pre-helper features and page accounting. Report actual held-out calls; a validation cap is not a deployment hard quota. Tied scores remain indivisible. Features do not inspect gold targets or auxiliary outcomes.

All six λ=0 models exactly reproduce hash-verified experiment 14 archive weights, epochs, NLL and repository roles. A complete rerun reproduces all 30 heads plus aggregate JSON **byte-for-byte (31/31 files)**. Tests cover cost-weighted ties, target sign changes, held-out outcome isolation, validation selection order and export-policy integrity.

Apple M4 Pro/macOS/Go 1.27.1/CPU: external wall 1.95 s, user CPU 1.59 s, peak RSS 90,898,432 bytes (~86.7 MiB). This single process includes source loading, candidates, 30 fits, validation/evaluation and writes. Not per-inference, GPU or Go-heap measurements.

```sh
go run ./cmd/riido-costclaim --out .cache/cost-claim-20
# Choose a new output directory for replay.
```

[Results](results-20.json) contain every candidate's validation metrics, hashes and selection. Keep raw corpus, queries, row labels and numerical weights out of Git. Publish numerical research heads only after suitability, manifest and exact-source CI verification. Defaults and opt-in behavior remain unchanged.

Previously observed documentation queries from three repositories are not a fresh independent user-task test. CoSQA reserves remain unscored and would not automatically validate this retrieval policy: they address pair relevance. Actual agent completion, total time and LLM-use savings remain unverified.

Next hold the cost target fixed and isolate **cheap pre-helper feature contributions**, counting feature overhead alongside pages and calls. Return to INT8/ternary comparison after establishing a useful full-precision candidate.
