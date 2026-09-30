# Auxiliary-search claim model 09: small to train, no established gain

[한국어](README.ko.md) · [Frozen plan](plan-09.json) · [Results and weight hashes](results-09.json) · [Reproduction/resources](reproduction-09.json)

**The primary candidate suggested no auxiliary search on any of 2,948 questions and did not improve the baseline.** Preserve this failure; do not adopt by default. There are 2,948 questions; 24 refers to exported model files, not question count.

## Claim and inputs

The input is query shape and raw BM25 score statistics. The output is a fallible suggestion that identifier-splitting search could reach a known target earlier. It does not establish truth, authorize actions, execute tools or choose Codex models. In this experiment a nonnegative margin selects auxiliary-first interleaving; otherwise baseline ordering remains. All candidates survive either policy.

Twelve features: bias, word count, byte count, uppercase/digit/underscore fractions, camel-case boundaries, unique-word fraction, squashed top score, top-two score gap, positive-score fraction in the top ten and positive-score fraction in the catalog. All are available at runtime. The feature function cannot receive gold function names, target ranks, repository labels or auxiliary results. Target ranks enter labels/evaluation only.

The label is strict improvement in first-known-target rank under auxiliary-first interleaving: positive on 801/2,948 questions. This binary label does not directly represent improvement magnitude or extra search CPU. The zero-margin threshold was fixed before execution and was not lowered after observing results. Margins are not calibrated probabilities.

## Training and model relationships

A 12-coefficient layer was trained from scratch in Go without a pretrained encoder. This is not compression of the full Laya/Model2Vec encoder; it is a separate minimal-signal search-control experiment.

| Evaluation repository | Training repository | Validation repository for epoch selection |
|---|---|---|
| etcd | kubernetes/test-infra | lxc/lxd |
| kubernetes/test-infra | lxc/lxd | etcd |
| lxc/lxd | etcd | kubernetes/test-infra |

Within a fold, evaluation-query labels never enter training/epoch selection. All candidate code documents remain visible as the retrieval catalog. These descriptions were previously observed in experiment 07, so this is not a fresh independent final test. A repository trains other folds; combining all fold heads invalidates a claim of independent evaluation on those same queries.

FP32 and ternary STE are separately fitted siblings per fold/seed. INT8 and ternary PTQ derive from the selected FP32 parent. Three folds × two seeds × two fitted modes = 12 fits, 24 exports with derivatives. Fixed learning rate 0.1, L2 0.0001, batch 64, 100 epochs; select minimum validation BCE. The primary candidate is FP32/seed1729.

## Results

Each question contributes once using the model whose evaluation repository matches its source.

| Policy | Auxiliary calls | Recall@1 | Recall@10 | Mean first-target rank |
|---|---:|---:|---:|---:|
| Raw BM25 | 0 | 41.79% | 74.83% | 48.99 |
| Always auxiliary-first interleave | 2,948 | 39.21% | 81.51% | 26.71 |
| Primary FP32 | 0 | 41.79% | 74.83% | 48.99 |
| INT8/seed1729 | 0 | 41.79% | 74.83% | 48.99 |
| Ternary PTQ/seed1729 | 53 | 41.82% | 75.00% | 47.73 |
| Ternary STE/seed1729 | 0 | 41.79% | 74.83% | 48.99 |

Every seed2718 variant also made zero auxiliary calls. Do not promote the one small PTQ improvement by changing the primary after the fact. No variant met the fixed gate: at least 10% mean-rank reduction without Recall@1/10 regression. The nonlearned always-interleave control reduces mean rank but hurts top-1. Small size alone does not establish usefulness.

A next hypothesis is mismatch between binary loss/threshold and actual utility. Predicting whether help occurs differs from minimizing candidates inspected. A separate experiment can investigate preserving the baseline's first candidate and cost-aware training/validation. Do not erase this failure or change its gate retrospectively.

## Resources, reproduction and publication

Apple M4 Pro/Go1.27.1 CPU: source checks, both searches, labels, 12 fits and 24 exports took 1.80s wall/1.44s user CPU, peak RSS 94,273,536 bytes. Replay reproduced all 24 models and the complete report byte-for-byte. No GPU was used.

Each model has 12 coefficients: 96 bytes for current float64 runtime coefficients, 521–669 bytes for JSON with metadata. These figures exclude program/index memory. Ternary-valued JSON is not a packed 1.58-bit runtime format. Feature extraction from an already computed 3,009-candidate ranking took about 1.75µs, 32 new bytes and one allocation in a repeated-fixture benchmark, excluding search and head inference.

```sh
# Prepare experiment 07's pinned corpus/ZIPs first; use a new output directory.
go run ./cmd/riido-claimtrain --out .cache/search-claim-new
go test -race ./internal/searchclaim ./internal/retrievalbench
```

Weights stay outside Git. A public HF research archive is a separate step gated on immutable result/file checksums and exact-source CI. Raw questions/code/label rows and ZIPs stay local. The new `riido-hubcheck` contract checks numeric coefficients, folds/dimensions/seeds, explicit filenames and notices, rejecting unknown model/manifest fields. Integrity is not proof of usefulness or exhaustive legal clearance. `.claim.json` research heads are incompatible with the existing `.hbin` input to `riido-hints --model`.

Source code/descriptions come from the three public repositories reviewed in [07](../retrieval-baseline/README.en.md), including separate-notice exclusions and stated license limitations. Heads use low-dimensional statistics and contain no raw source strings. Both CoSQA reserves remain unused for training/scoring. Real-user task success, token savings and production readiness remain unestablished.
