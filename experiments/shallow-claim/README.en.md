# Shallow page-gain claim learning 22

**Fewer calls, but worse page cost and retrieval quality: the precommitted gate fails.** Compared with experiment 21 seed 1729, calls fall 2,408→1,884 while pages rise 6,326→6,333. No default integration. No LLM savings established.

## Design

Of 2,948 queries, helper pages improve for 332, worsen for 146, and tie for 2,470. Test whether a tiny conditional model can identify useful calls. Keep the same 16 baseline-only features and train a Go regression tree predicting baseline pages minus helper pages. Positive scores indicate predicted savings; they are neither probabilities nor action authority.

**Both architecture and objective change** versus linear classification. Do not attribute differences solely to nonlinearity. Depths 1/2/3/4 × minimum leaf sizes 16/64 × three repository folds yield 24 models, not 24 questions. There are **2,948 distinct documentation queries**. This deterministic algorithm has no random seed; replay is not independent data evidence.

The [plan](plan-22.json) was committed before fitting. Greedy splits minimize squared error on training rows; leaves store training mean savings. Select each threshold with the existing validation-only 90% call budget, then choose by validation pages, calls, smaller depth and larger minimum leaf. Evaluation outcomes do not select candidates. Controls use the verified immutable HF 21 archive without refitting.

## Results

| Policy | Calls | Pages | Recall@10 |
|---|---:|---:|---:|
| Raw | 0 | 9,488 | 0.7483 |
| Always helper | 2,948 | 6,287 | 0.8138 |
| Gap rule | 2,548 | 6,323 | 0.8097 |
| Experiment 21, seed 1729 | 2,408 | 6,326 | 0.8100 |
| Experiment 21, seed 2718 | 2,317 | 6,331 | 0.8094 |
| Shallow tree | 1,884 | 6,333 | 0.8039 |
| Gold page oracle | 332 | 6,014 | 0.7948 |

All preserve raw Recall@1 of 0.4179. Validation selects depth 4/minimum 16 for held-out etcd and depth 4/minimum 64 for test-infra and lxd. Each policy counts all 2,948 queries once. The tree reads 46 extra pages versus always-helper and 10 versus the gap rule; fewer calls do not override the fixed gate.

## Failure accounting

| Policy | Beneficial calls / 332 | Missed saved pages | Harmful calls / 146 | Incurred harmful pages | Page-tie calls / 2,470 |
|---|---:|---:|---:|---:|---:|
| Gap rule | 319 | 39 | 144 | 270 | 2,085 |
| Experiment 21, 1729 | 315 | 53 | 139 | 259 | 1,954 |
| Experiment 21, 2718 | 313 | 58 | 139 | 259 | 1,865 |
| Tree | 299 | 62 | 135 | 257 | 1,450 |

The tree skips many page ties but still calls on 135 of 146 harmful queries. Here total pages equal `6,014 oracle + missed savings + incurred harm`; for the tree, 6,014+62+257=6,333. This motivates investigating signals of harmful helper retrieval rather than only reducing calls. It does not prove the current features contain no such signal.

The oracle reads gold outcomes and is not deployable. Page ties can change ranks within a 20-candidate page and therefore Recall@10; they are not necessarily useless calls. Oracle promotion is explicitly disabled.

## Resources and reproduction

The Go model uses separate fixed arrays for at most 31 nodes' feature indices, children, cuts and values. Validated immutable scoring needs no locks, at most four comparisons and no allocations. Training does allocate temporary sorting/partition arrays. No SIMD or GPU is used. These are float64 numerical trees; ternary compression remains a separate experiment.

Apple M4 Pro/macOS/Go 1.27.1: repeating one synthetic depth-four path measured 2.501/2.483/2.407ns, 0 B/op, 0 allocs/op. This favors warm caches and branch prediction. It excludes feature extraction, search, JSON and process startup; it is not real query latency. Scoring accounted for 75.19% cumulative CPU samples in that same synthetic loop only.

One full source-verification/preparation/24-fit/evaluation/export process took wall 1.80s, user CPU 1.43s, peak RSS 99,008,512 bytes (~94.4MiB). This is whole-process CPU evidence, not per-model or GPU memory. A JSON tree is 2,059–2,562 bytes. Compact serialization is deferred until utility is established.

All 24 models plus aggregate output reproduce byte-for-byte (25/25). Both frozen experiment-21 control policies exactly reproduce all prior metrics. Tests cover leaf size, identical features, deterministic selection, malformed topology and held-out isolation.

```sh
go run ./cmd/riido-shallowclaim --control .cache/spread-claim-hf-verified-21 --out .cache/shallow-claim-22
go test ./internal/claimtree -run '^$' -bench BenchmarkScoreDepth4 -benchmem
```

[Full results](results-22.json). [Tree documentation](https://scikit-learn.org/stable/modules/tree.html) informs squared-error splits, mean leaves and depth bounds. No upstream implementation was copied and no Python runtime was added.

These repeatedly observed queries from three repositories remain development evidence. Fresh ≥2,400 real-user requests, repository choice, task decomposition, actual agent success and total cost remain unverified. CoSQA reserves remain unscored. Next isolate harmful-call signals and the quality/cost relationship in a separate plan. Failed models remain reproducibility artifacts; raw sources, query rows, labels and profiles stay unpublished.
