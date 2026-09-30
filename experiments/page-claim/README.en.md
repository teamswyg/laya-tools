# Page-saving claim model 14

**Trained, not promoted.** FP32, INT8, ternary PTQ and ternary STE proposed the helper for every query under both seeds. They matched the always-helper control, establishing no practical benefit from adding a learned head.

## Objective and model family

There are 2,948 distinct documentation queries and 3,009 function candidates. The 24 exports are three folds × two seeds × four representations, not 24 questions. FP32 and ternary STE were fitted separately (12 fits); INT8 and ternary PTQ derive from each FP32 parent. Unlike experiment 09, this experiment uses baseline-first interleaving and a page-cost objective, so differences cannot be attributed to the loss alone.

Page size is fixed at 20. The binary label is whether the helper needs fewer pages through the first known target; the sample weight is the absolute page-count difference. Ties have zero weight. Saving ten pages contributes ten times as much loss weight as saving one. Inputs remain the original 12 runtime-only query-shape and baseline-score features. Gold ranks/names, repository identities and auxiliary search outcomes are never features.

Training minimizes weighted BCE. Training weights are normalized globally to mean one, with gradients divided by batch sample count; validation also uses weighted BCE. Settings and gates were committed in plan-14.json before fitting. CPU cost is not converted to output cost using an invented exchange rate. Omitted sample weights preserve the legacy objective; replaying experiment 09 produced byte-identical results and all 24 model files.

## Evaluation and results

The repository counts are 806 etcd-io/etcd, 1,152 kubernetes/test-infra and 990 lxc/lxd queries. Training, validation and evaluation rotate by repository. A model's held-out query labels are not used to fit it or select its epoch. All code candidates are visible in every fold. This is previously observed data used for experiment design, not an independent final evaluation; three source repositories remain a limitation.

| Policy | Helper calls | Total pages | Emitted candidates | Recall@10 |
|---|---:|---:|---:|---:|
| Raw BM25 | 0 | 9,488 | 189,760 | 74.83% |
| Always helper | 2,948 | 6,287 | 125,740 | 81.38% |
| Every trained representation/seed | 2,948 | 6,287 | 125,740 | 81.38% |

Both primary FP32/1729 and replication FP32/2718 fail the fixed exploratory gate: no more pages than always-helper, at least 10% fewer helper calls, and Recall@1/10 no worse than baseline. No threshold adjustment or winning-seed selection follows evaluation. Scores are not calibrated probabilities or action authority.

On Apple M4 Pro CPU, source verification, search, 12 fits and 24 exports took 1.78 s wall, 1.42 s user CPU, and 90,750,976 bytes peak RSS. GPU was unused. All exports and the full report replayed byte-identically. Models are numeric JSON, not packed 1.58-bit runtime artifacts. They are not compressed Laya models and have no pretrained encoder.

## Next evidence needed

Of 2,948 queries, 2,470 have equal page counts and therefore zero weight. The loss does not reward skipping the helper on those queries; call reduction appears only in the acceptance gate, not the training loss. Feature insufficiency is therefore not established as the cause. A separate next experiment should precommit helper-call budgets and threshold selection rules, selecting only on the validation repository. This experiment will not adjust its threshold after observing results.

Changing the objective did not produce useful discrimination with these features/settings. This does not prove semantic learning impossible. Before fitting another variant, test whether additional runtime features contain a signal separating helper benefit and harm. Features requiring auxiliary search cannot justify avoiding that search. At least 2,400 independent real-user questions and actual verification-cost evaluation remain outstanding.

```sh
go run ./cmd/riido-pageclaim --out .cache/new-page-claim-run
```

The pinned source projection and ZIPs from experiment 07 are required. Weights stay outside Git and may be archived separately on Hugging Face after exact-source CI and public-package verification. Router/session defaults remain unchanged. No raw source, queries or private data are distributed, and upstream data is not relicensed. Not production ready. No LLM savings established.
