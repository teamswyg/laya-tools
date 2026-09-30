# Execute auxiliary path search only when selected — experiment 42

**A new Go path genuinely omits auxiliary index construction and ranking when auxiliary search is declined.** This differs from merely counting a precomputed result as unused. No useful selector has been trained here, and actual LLM latency/token savings remain unmeasured.

The [plan](plan-42.json) was committed first. Existing `Rank` computes baseline, normalized and interleaved orders for comparison. New `RankSelected` ranks the baseline and passes only experiment 41's 16 numerical features, by value, to a selector. Auxiliary indexing/ranking happens only after a true result. A nil selector skips feature extraction too and returns baseline-only ranking.

| Path | Result | Actual auxiliary work |
|---|---|---|
| Nil selector | Full baseline order | 0 builds, 0 rankings |
| False result | Full baseline order | 0 builds, 0 rankings |
| Selector error | Baseline with `selector_error` | 0 builds, 0 rankings |
| True, successful auxiliary execution | Full interleaving preserving the baseline's first candidate | Build when needed, then rank |
| True, auxiliary failure | Baseline with `auxiliary_error` | Count actual failed attempts |
| Invalid baseline | Error | Selector is not called |

Feature/interleaving errors also have explicit baseline fallbacks. Private selector error messages are not copied into the result. No candidates are removed and no answer is treated as certain; auxiliary results only reorder the same candidate set.

`WorkStats` counts **actual attempted operations**: baseline ranking, auxiliary index construction and auxiliary ranking, including failures. These are not model-call counts or estimates of time/memory/tokens saved. Future total-cost measurements must also include baseline index construction, features and the selector itself.

Shared baseline/auxiliary operations were factored out of the existing `Ranker`. Old ranking results are checked against an independent pre-refactor reference. The standalone function retains no index between calls. Only `Ranker.RankSelected` opts into exact-full-catalog index reuse. Experiment 35 did not establish a real-workload benefit from reuse, so it is not activated by default.

A synthetic reused-worker sequence `skip → use → use → skip → change catalog and use` records five baseline rankings, two auxiliary builds and three auxiliary rankings. After the first skip, no auxiliary index exists. Skipping with a warm auxiliary index still performs zero auxiliary rankings. Reused objects may retain the previous index memory; skipping does not imply freeing that memory.

Tests cover always-on/off and selector-error policies, lazy first construction, reuse and catalog mutation, independent prior results/feature arrays, long queries/10,000 candidates and independent workers under race detection. Punctuation-only queries normalize to empty text and fall back in the existing reference too; the new path preserves this. Queries that exceed 128KiB only after normalization also retain the full baseline order. Full race/vet/format/redacted scans gate publication.

One sequential worker owns each `Ranker`. Selectors must not mutate input paths or re-enter the same object. Independent workers use separate objects and immutable models; no global cache or lock is added.

This remains an internal research API and does not automatically connect learned weights to existing CLI commands. Next, fit a selector under fixed development/source-eligibility rules and compare baseline-only, always-auxiliary and learned selection using actual computation and quality. These synthetic execution checks are not a new 2,400-task user evaluation or a training success. With acquisition running separately, they are not latency/memory benchmarks. No final evaluation or new HF release occurs here.
