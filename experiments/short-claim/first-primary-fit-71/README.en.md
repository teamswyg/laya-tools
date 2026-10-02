# First claim-model fit: completed, utility failed

The first encoder-free FP32 fit completed but required more checks than the strongest nonlearned control. Keep the existing controls for now. This experiment proposes **verification order while retaining every candidate**. It does not discard candidates or authorize work.

| Method on 15 known validation requests | Total checks | Correct first candidate | Correct within top three |
|---|---:|---:|---:|
| Original order | 31 | 1 | 10 |
| BM25 | 28 | 5 | 10 |
| Ordered lexical control | **27** | **5** | 10 |
| Narrow conditional rule | 28 | 5 | 10 |
| First fitted model | **31** | **2** | 10 |

Five unknown requests are excluded from 20 validation requests, leaving 15: ten answerable and five no-answer. No-answer retains the cost of checking every candidate. Utility includes all 45 known candidates, including masked endpoints. Select one comparator by minimum total checks with the previously fixed tie order.

Relative gain was `(27−31)/27 = −14.81%`. The fixed ≥5% improvement and Top1/Top3 conditions failed on check reduction and Top1. Oracle required 21 checks, so headroom existed without being achieved. The fully eligible diagnostic subset also worsened, 26→30. The four extra checks came from three lifecycle-acquire requests; unknown data or one masked candidate cannot explain the failure alone.

## Actual execution

Reuse the [complete76 prepared arrays](../projection-execution-76/README.en.md), retaining train91/validation45 and 18/1 zero-weight rows. No new roles, labels, features, calibration fitting, protected-final reads or paid model calls occurred. Only original request and candidate text enter features.

Go 1.27.1 ran **one** FitWithTrace with seed1729, 8192 FP32 coefficients, LR0.1, L2 0.0001, batch128 and 50epochs. The original earliest minimum validation-BCE epoch selector chose epoch50. Training NLL was 0.6514239277140677 and validation NLL 0.6657976915843424. Epoch selection and utility reused the same validation set; this is not final generalization evidence. With 91 rows below batch128, training made one update per epoch, 50 updates total.

The FP32 file is 32,792 B, SHA256 `5ecbeb9b35f67cd91de5c1674d0398e9890bae51029684f45cc955e7407c7049`. Decode expands coefficients into 65,536 B of float64 values; optimizer and loss accumulation also use float64. File size is not total runtime memory. Coefficients remain in a private file outside Git. Publication qualification and production readiness are false.

Whole-child peak RSS was 28,295,168 B (about 26.98 MiB), outside wall 0.805194625 s. This one observation includes input reading, controls, fitting, checkpoints and output. CPU1, soft Go heap256MiB and outside300s timeout were configured. It is not GPU/Laya-encoder execution, resident inference latency or LLM savings. Preserve the [actual ledger](actual/ROOT-ACTUAL-LEDGER-71.v1.json) and [full numeric result](actual/results.json). AI-assisted preparation/review cost remains unmeasured.

## One next experiment

Lower average BCE need not improve within-request order, although BCE can learn ranking when features and data are sufficient. Keep data, lexical-feature and training-scale limits separate. The [fixed next plan](NEXT-PDCA-71.v1.json) retains BCE and adds one same-request positive/negative logistic-ranking term. Fix λ1, pair mean followed by parent mean, and preserve seed, features, masks, validation-BCE epoch selection and utility conditions. Do not search coefficients, seeds or thresholds.

This is an objective sibling on the same data/features, not a compressed child. Defer ternary compression until utility is established. The next choice follows a failure on reused validation, so even an improvement remains developmental. Preserve the final goal of at least 2,400 distinct requests per domain from fresh whole-source groups.

For current use, `riido-shortclaim --stream --baseline lexical_ordered < requests.jsonl` returns order hints with every candidate and `unverified_heuristic` status. Actual-work savings require separate observation. This command uses no newly fitted coefficients and does not register with Codex automatically.

[Independent numerical review](independent-qa/FINDINGS-FIRST-FIT-71.en.md) · [Baseline tool runtime cost](../baseline-runtime-profile-71/RESULTS-BASELINE-PROFILE-71.en.md) · [한국어](README.ko.md) · [Public dataset/HF proof](../publication-proof-93/README.en.md)
