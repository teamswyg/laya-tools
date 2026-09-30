# Joint query-coverage claim training 25

**The 20-feature combination worsens page cost for both initializations; do not adopt these additions in the default model.** Primary pages rise 6,326→6,439 while calls fall 2,408→2,363: 45 fewer calls at the cost of 113 extra pages. The unchanged gate fails.

## Controlled comparison

Weak individual signals in experiment 24 did not exclude interactions. Commit the [plan](plan-25.json) before fitting and append all four coverage features, without selecting favorable evaluation features.

Keep the original 12 signals and four score-distribution signals, then append first-candidate query-term coverage, top-20 mean/stddev and vocabulary absence: 20 total. They use baseline search/index information before helper retrieval, never repository identity, gold or helper outcomes.

Retain the sign of baseline pages minus helper pages minus call penalty as target, absolute magnitude as weight. Same penalties {0,0.25,1,4,16}, seeds {1729,2718}, learning rate 0.1/L2 0.0001/100 epochs/batch64, minimum-validation-loss epoch and validation-only 90% budget/penalty selection. Three repository folds ×two initializations ×five penalties yield **30 numerical heads**, evaluated on **2,948 distinct queries**.

Changing dimension can alter initialization and optimizer trajectory. A shared seed does not establish identical paired optimization paths; two seeds are not independent data evidence.

## Results

| Policy | Seed | Calls | Pages | Recall@10 |
|---|---:|---:|---:|---:|
| Raw | — | 0 | 9,488 | 0.7483 |
| Always helper | — | 2,948 | 6,287 | 0.8138 |
| Gap rule | — | 2,548 | 6,323 | 0.8097 |
| Previous 16 features | 1729 | 2,408 | 6,326 | 0.8100 |
| Previous 16 features | 2718 | 2,317 | 6,331 | 0.8094 |
| New 20, cost selected | 1729 | 2,363 | 6,439 | 0.8111 |
| New 20, cost selected | 2718 | 2,328 | 6,440 | 0.8104 |
| New 20, λ=0 | 1729 | 2,457 | 6,434 | 0.8111 |
| New 20, λ=0 | 2718 | 2,432 | 6,429 | 0.8107 |

All preserve raw Recall@1 of 0.4179. Both seeds select λ=1/0/0 for held-out etcd/test-infra/lxd. Do not switch the primary policy after observing better λ=0 evaluation pages. Top-10 quality rises slightly, but page cost worsens and violates the original objective.

| Primary evaluation repository | Old→new pages | Old→new calls |
|---|---:|---:|
| etcd | 1,602→1,599 | 618→540 |
| test-infra | 2,300→2,418 | 852→894 |
| lxd | 2,424→2,422 | 938→929 |

The second seed also worsens test-infra pages, 2,305→2,419. Effects differ across repositories; local gains do not establish overall improvement. This does not prove coverage useless for every task, but gives no support for these additions under the current objective/split. Do not keep searching the same data until a favorable number appears.

## Implementation, replay and resources

A shared Go trainer uses fixed eight-column extra arrays with active widths 0/4/8 for 12/16/20 dimensions. Preserve old arithmetic order and schemas. Archive verification requires `riido-coverage-claim-v1` and exactly 20 coefficients, rejecting confusion with 16-feature heads. These are FP32-rounded linear coefficients, not ternary models or Laya checkpoints.

The 30 new heads/report reproduce **31/31 files byte-for-byte**. After refactoring, experiments 21 (16 features) and 20 (12 features) each reproduce 31/31 files. Previous HF archive integrity checks also pass. Tests cover column order, weighted sum, unchanged targets/weights, held-out isolation, invalid values/dimensions and model contracts.

One Apple M4 Pro/macOS/Go 1.27.1 CPU source-check/retrieval/feature/30-fit/evaluation/export process took wall 2.04s, user 1.67s, peak RSS 90,374,144 bytes (~86.2MiB). This is whole-process cost, not per-inference/GPU memory or proof of RSS improvement. Each JSON head is 1,152–1,171 bytes. See [experiment 24](../query-coverage/README.en.md) for scoped extraction microbenchmarks. No LLM savings established.

```sh
go run ./cmd/riido-coverageclaim --out .cache/coverage-claim-25
```

[Full results](results-25.json). Research heads go only to HF after CI/license/hash verification; raw sources, queries, row labels, weights and profiles stay out of Git. No default activation.

Defer further coverage tuning. Next separate repository-specific failures and rare large page losses, distinguishing input/distribution issues from objective issues in another plan. Fresh real-user ≥2,400 requests and actual agent success/total-cost evidence remain outstanding. Repeated observation of these three repositories is not an independent final test. CoSQA reserves remain unscored.
