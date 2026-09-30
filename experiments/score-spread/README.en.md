# Cheap score-spread feature ablation 21

**Four added features reduce pages and calls under the same cost objective, but the precommitted gate still fails.** Primary pages improve 6,337→6,326 and calls 2,470→2,408 relative to the 12-feature model. Versus the cheap gap rule it saves 140 calls but reads three extra pages. Defaults remain unchanged.

## Controlled change

Add mean, population standard deviation, first-candidate score share, and last/first score ratio over the top 20 baseline scores. Mean/stddev use scores divided by the maximum first. Use actual count when fewer than 20 candidates exist; return four zeros when the maximum is zero. Computing ratios first avoids overflow from large score sums.

Features accept **only the already-computed baseline ranking**, never gold targets, repository identities or auxiliary results. Freeze them before auxiliary/target ranks. Keep the four extras in a parallel fixed-width array, preserving old example and 12-feature contracts. New heads use `riido-spread-claim-v1` with 16 coefficients; publication verification rejects interpreting them as old heads.

Freeze the [plan](plan-21.json) before execution. Keep experiment 20's penalties {0, 0.25, 1, 4, 16}, seeds, fitter, repository folds and validation-only 90% threshold/penalty selection. Train 30 new 16-dimensional heads. Penalties are experimental page-equivalent units, not measured billing/token conversions.

| Policy | Seed | Calls | Pages | Recall@10 |
|---|---:|---:|---:|---:|
| Raw | — | 0 | 9,488 | 0.7483 |
| Always helper | — | 2,948 | 6,287 | 0.8138 |
| Gap rule | — | 2,548 | 6,323 | 0.8097 |
| Old 12-feature selected cost | 1729 | 2,470 | 6,337 | 0.8094 |
| Old 12-feature selected cost | 2718 | 2,489 | 6,338 | 0.8094 |
| New 16-feature selected cost | 1729 | 2,408 | 6,326 | 0.8100 |
| New 16-feature selected cost | 2718 | 2,317 | 6,331 | 0.8094 |
| New 16-feature zero penalty | 1729 | 2,434 | 6,348 | 0.8083 |
| New 16-feature zero penalty | 2718 | 2,339 | 6,342 | 0.8073 |

Each row covers the same 2,948 distinct docstring queries once. Baseline-first preserves Recall@1. Both seeds select λ=4/0/0.25 for held-out etcd/test-infra/lxd; held-out results do not select penalties. Seeds are not independent data replications.

The primary policy improves both metrics over the old learned policy, but uses 39 extra pages versus always-helper and three versus the gap rule. It fails the unchanged gate: pages≤always-helper, calls≤90%, Recall@1/10≥raw, and learned policy Pareto-dominates the gap rule in pages/calls. Do not lower this gate or select a favorable seed.

## Cost and reproducibility

M4 Pro/macOS/Go 1.27.1, CPU. One full preparation/training/evaluation/export process: 2.09 s external wall, 1.71 s user CPU, peak RSS 92,291,072 bytes (~88.0 MiB). Not per-inference memory.

Three synthetic additive-function benchmark runs: 42.27/39.31/39.65 ns, 0 B/op and 0 allocs/op. Valid-input allocation assertion passes. CPU pprof attributes 96.3% of samples to the feature function including its validation. This profiles repeated microbenchmark calls, not whole-app bottlenecks or GPU work. Raw profiles remain local.

Fixed-order combined benchmarks paradoxically report the 16-feature path faster than 12 features; do not interpret that timing variation as a speedup. Rely on measured additive-function cost/allocation only; service latency remains untested.

All 30 new heads and report replay byte-for-byte (31/31). Experiment 20 replay after shared-code refactoring also matches all 31 old files. Existing HF 20/14 archives still verify. Tests cover zero/uniform/concentrated/large scores, malformed inputs, bounded prefix reads, split isolation and incompatible 12/16-dimensional export contracts.

```sh
go run ./cmd/riido-spreadclaim --out .cache/spread-claim-21
go test ./internal/searchclaim -run '^$' -bench 'Benchmark(SpreadExtra|FeatureVariants)' -benchmem -benchtime=300ms -count=3
```

[Results](results-21.json), [benchmark aggregates](benchmark-21.json). Publish numerical research heads only to HF after exact-source CI/integrity checks. No raw corpus, queries, row labels, weights or profiles in Git. Scores are cost surrogates, not calibrated probabilities or action authority.

Previously observed docstrings from three repositories are not fresh independent final evidence. Real requests, agent completion and LLM savings remain unverified. CoSQA reserves remain unscored and do not automatically validate this retrieval policy. Further work must separate cheap information value from actual task costs.
