# PDCA 02: stronger retrieval baseline and bounded damage from bad hints

[한국어](CYCLE-02.ko.md) · [Full results](results-02.json) · [Plan](PLAN.en.md)

## Usable interface

`pkg/hintsearch` provides pure-Go BM25 retrieval and combination with external hint ordering. `cmd/riido-hints` exposes JSON stdin/stdout. No model download, Python or GPU; existing router settings remain unchanged.

```sh
go run ./cmd/riido-hints < examples/hints/request.json
```

The example intentionally supplies an irrelevant cache hint: cache comes first, followed by BM25's payments candidate. Every candidate remains and status is always `unverified`. `bm25_score` is the original lexical score, not confidence in the combined ordering. Omit `hints` for plain BM25.

Humans can edit the JSON catalog; agents can use the same contract. The CLI neither reads repository files nor calls APIs. Callers must provide permission-filtered documents and update snapshot IDs whenever content or permissions change; IDs are not cryptographic verification of content. Query/snapshot mismatch, duplicate or unknown IDs fail without partial output. Returned candidates are not executable instructions. A downstream verifier should return incomplete on budget exhaustion; this CLI does not yet implement verification, budgets or caching.

## Why alternate?

Alternate one unvisited hint H candidate and one unvisited baseline B candidate, skipping duplicates. H may be partial; B must preserve all candidates.

**Proof scope:** a candidate at baseline rank r has r−1 baseline predecessors. Each round emits at most one H candidate and advances B to its next unvisited candidate. H visiting earlier B entries only accelerates baseline progress. Consequently the target appears by round r, within `min(N, 2r)` inspections, with no duplicates or omissions.

This bound concerns **accurate, unit-cost verification of a single candidate**. It is not a bound on variable verification costs, multi-evidence completion, hint generation, LLM tokens or wall time, and does not repair verifier errors. External hint generation needs separate resource limits. This limits damage; it does not guarantee improvement.

## Implementation and tests

BM25 k1=1.2, b=0.75 uses term frequency, length normalization and rarity; repeated query terms count once. This is lexical, not semantic. Sorted vocabulary and contiguous posting columns store document positions/frequencies, with a normalization array. Shared indexes are immutable: no maps or locks on the query path; each caller owns its result. Go runtime synchronization still exists.

Limits: 1..4096 documents, 2MiB combined text, 8192-byte queries, 16MiB encoded CLI JSON. Input limits are not total RSS limits. Tests check formula values, empty documents, Korean token matching, bounds, concurrent snapshots and stale hints. Token matching does not establish Korean semantic quality. Exhaustive hint permutations and all prefix lengths for up to seven candidates check the 2r bound and preservation, independently of model evaluation.

## Same 52 development requests

| Method | Recall@1 | Recall@4 | Mean oracle checks | Maximum per-case check ratio vs BM25 |
|---|---:|---:|---:|---:|
| BM25 | 52.08% | 75.00% | 4.615 | 1.0 |
| Tokens interleaved with BM25 | 52.08% | 75.00% | 4.615 | 1.0 |
| Fixed shuffle interleaved with BM25 | 6.25% | 75.00% | 5.308 | 2.0 |
| Reverse BM25 interleaved with BM25 | 2.08% | 75.00% | 5.385 | 2.0 |

All four absent-answer requests require all 16 checks. BM25 gets direct 16/16, paraphrase 1/16 and contrast 8/16 at rank 1, matching exact tokens. No semantic improvement. Bad hints increase average cost: the evidence does not support adding arbitrary hints. Reused development data is not a new final evaluation.

This single run: BM25 warm p95 0.625µs; full nine-method process maximum RSS 11,862,016B (11.31MiB), peak footprint 9,732,576B, wall 0.48s, user CPU 0.10s, zero swaps. Query timing includes tokenization/scoring/sorting for 16 candidates, excludes index construction and verification. Fixed method ordering and noise preclude strong speedup claims.

A separate dense-posting microbenchmark with 64 identical documents reports 523.7ns/op, 1128B/op, five allocations/op. It is not representative of real repositories or workloads. CPU pprof samples were mostly Go runtime waits/memory management, insufficient evidence for new SIMD/lock optimizations. Raw CPU/heap profiles remain local. GPU usage, energy savings and LLM costs were not measured.

## Next cycle

A learned ordering can now use the same CLI and verification contract. There is no learned model or HF release in this cycle. Next: public family-separated data, BM25 and lightweight learned controls, FP32/INT8/ternary comparisons, and a new sealed semantic-family final evaluation. The overall usefulness goal is not complete.
