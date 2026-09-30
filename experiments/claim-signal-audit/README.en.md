# Claim input signal audit 23

**Existing score distributions signal that helper retrieval may change page cost, but individual features weakly separate benefit from harm.** Audit the information before spending more fits on the same inputs. This produces neither a new model nor a deployed policy.

## Why separate the targets?

Experiment 22's tree still calls on 135 of 146 harmful queries. A feature may distinguish harm from the rest while also scoring beneficial queries highly; suppressing calls with that feature can remove benefits too.

Audit all 16 existing features on **2,948 distinct queries**: 332 page benefits, 146 harms and 2,470 ties, using 20-candidate pages.

- Harm versus rest: 2,948 cases, 146 positives.
- Benefit versus rest: 2,948 cases, 332 positives.
- Harm versus benefit: **478-case subgroup** excluding page ties. This is not a separate ≥2,400-case final evaluation.

The [plan](plan-23.json) was committed before execution. Each feature's orientation is chosen solely by training-repository AUC: flip when below 0.5; keep otherwise. Validation and evaluation labels never choose the direction. Retain the existing etcd/test-infra/lxd rotating split and count every eligible query once per feature/target condition.

## Findings

AUC describes positive-versus-negative pair ordering; **0.72 AUC does not mean 72% accuracy**. Average precision (AP) measures positive concentration across score thresholds. Equal scores are grouped and cannot be split by input order.

| Target | Feature | Mean repository AUC | Mean AP | Mean prevalence |
|---|---|---:|---:|---:|
| Harm/rest | Top-20 last/first score ratio | 0.7217 | 0.1510 | 0.0502 |
| Benefit/rest | Same ratio | 0.7307 | 0.2330 | 0.1152 |
| Harm/benefit | Same ratio | 0.4789 | 0.3396 | 0.3067 |
| Harm/benefit | Squashed top score | 0.5910 | 0.4066 | 0.3067 |
| Harm/benefit | Unique query-word fraction | 0.5730 | 0.3500 | 0.3067 |

These are **unweighted means across three repositories**, not pooled/query-weighted AUC or AP. Prevalence is averaged the same way, so differs slightly from 146/2,948. Constant-feature AUC is 0.5 and AP equals each repository's positive prevalence.

For both harm/rest and benefit/rest, training consistently orients larger last/first ratios toward positives. Evaluation AUC ranges are approximately 0.66–0.78 and 0.71–0.74 respectively. This supports investigating overlapping score-distribution signals, while the harm/benefit mean AUC of 0.4789 is weak.

The highest observed mean harm/benefit AUC among individual features is 0.5910 for squashed top score (repositories 0.5554/0.6350/0.5827). It is the **observed maximum after multiple comparisons**, not an independently validated winner or useful policy. Weak individual signals do not establish absence of multivariate information. AP is not precision at a chosen operating threshold or actual cost saved.

## Consequence for the next experiment

Do not assume further score-distribution fitting will identify harmful calls. A next candidate is **direct query-word coverage in the baseline's top candidates**, computed before helper retrieval. Investigate reuse of the existing index's sorted posting arrays and count extraction time, throughput and memory. Information requiring helper retrieval must not be advertised as free evidence for skipping that retrieval.

Precommit the next feature change separately from objective changes. These repeatedly observed development queries are not fresh final evidence. Final evaluation still needs independent real-user ≥2,400 requests, including counts and sources of rare harmful cases. Report deliberately enriched stress cases separately from representative-distribution evaluation; no repetition or rewriting to inflate counts.

## Reproduction and resources

```sh
go run ./cmd/riido-claimaudit --out .cache/claim-signal-audit-23
```

Source and candidate hashes are verified before recomputing features/outcomes. No model fitting, threshold selection or new weight export occurs. The complete report replays byte-for-byte. Tests cover tie-order invariance, constant/reversed signals, invalid scores/single classes, subgroup membership and held-out mutation isolation.

One Apple M4 Pro/macOS/Go 1.27.1 CPU process took wall 1.81s, user 1.45s, peak RSS 94,158,848 bytes (~89.8MiB), including source verification, retrieval, feature preparation, analysis and output. This is not per-inference or GPU memory. Raw queries, label rows, predictions and model weights remain unpublished. No new HF model repository is created for a feature audit.

[All 144 fold conditions and 48 aggregates](results-23.json). No multiple-comparison-adjusted significance, new-domain/language generalization, agent completion or LLM savings is established. CoSQA reserves remain unscored.
