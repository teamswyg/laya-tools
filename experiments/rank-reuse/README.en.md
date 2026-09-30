# Ranking array reuse 08: identical results with about 97% less cumulative allocation

[한국어](README.ko.md) · [Measurements](results-08.json) · [Corpus and limitations](../retrieval-baseline/README.en.md)

**Repeated-search allocation decreased; a large latency improvement or resident-memory reduction was not established.** The same 2,948 questions and 3,009 candidates were used. No new training, data selection or scoring change occurred.

## API and ownership

`pkg/hintsearch.Index.RankInto(query, previous)` reuses caller-owned order/score arrays. Existing `Rank` continues returning independently owned results. Scores, ties, zero-score documents and the complete candidate set are preserved.

```go
var scratch hintsearch.Ranking
for _, query := range queries {
    var err error
    scratch, err = index.RankInto(query, scratch)
    if err != nil {
        return err
    }
    // Consume scratch.Order and scratch.Scores before the next call.
}
```

The next reuse overwrites these arrays. Copy retained results or use `Rank`. Concurrent workers need separate buffers; the index remains immutable. No shared pool, map or mutex was introduced. Invalid queries return before mutating the buffers. Reuse across larger/smaller indexes resizes and clears scores appropriately.

The one-request `riido-hints` CLI does not gain repeated-query savings merely from this addition. Intended users are Go callers and a future daemon querying the same index repeatedly. The evaluator exposes `--reuse-rank` for direct use.

## Corpus preservation and allocation

Four process runs used fresh → reuse → reuse → fresh order. Each evaluated both raw/identifier modes. Hashes cover every candidate position and float64 score bit for every question; all four matched. All earlier 07 overall, name-stratum and repository quality metrics also matched.

| Run | Combined query-loop cumulative allocation | Full wall time | Peak RSS |
|---|---:|---:|---:|
| Fresh 1 | 298,879,080 B | 1.89s | 102,465,536 B |
| Reuse 1 | 9,176,888 B | 1.50s | 92,291,072 B |
| Reuse 2 | 9,176,888 B | 1.50s | 93,503,488 B |
| Fresh 2 | 298,884,224 B | 1.52s | 91,062,272 B |

Cumulative allocation decreased by about **96.93%**. These Go `TotalAlloc` deltas cover two loops of 2,948 queries, including preprocessing, trace hashing and metric collection. They are not live memory or model size. Source/archive checks and index construction are excluded from loop allocation but included in whole-process wall/RSS. Fresh run 2 had lower peak RSS than reuse, so no peak-RSS improvement is established. Per-query p95 also remained similar; raw values are in JSON.

## Microbenchmark

Original public fixture strings form catalogs of 64/3,009 candidates with one repeated query. Medians are from five 300ms runs per configuration. This is separate from the diverse corpus evaluation.

| Candidates | Mode | New bytes/op | Allocs/op | Median time |
|---:|---|---:|---:|---:|
| 64 | Fresh | 1,112 B | 5 | 655.3ns |
| 64 | Reuse | 88 B | 3 | 588.9ns |
| 3,009 | Fresh | 49,240 B | 5 | 59,595ns |
| 3,009 | Reuse | 88 B | 3 | 59,493ns |

At 3,009 candidates, new bytes/op decreased by about 99.82%, with negligible time change. Reusable arrays still occupy caller-owned memory after preparation; 88 bytes measures subsequent new allocation only. Tokenization and sorting costs remain. This is not a SIMD experiment.

## Reproduction and checks

Apple M4 Pro/Go1.27.1 CPU. Prepare experiment 07's pinned source and archives.

```sh
go test -race ./pkg/hintsearch ./internal/retrievalbench
go test ./pkg/hintsearch -run '^$' -bench BenchmarkRankReuse -benchmem -benchtime=300ms -count=5
go run ./cmd/riido-retrievalbench --stage evaluate
go run ./cmd/riido-retrievalbench --stage evaluate --reuse-rank
```

Regression tests cover stale-score removal across different queries, missing-term/tied candidates, changing index sizes, nonmutation on invalid input and concurrent queries with separate buffers. Existing analytical BM25 and candidate-preservation tests remain.

This establishes a runtime allocation improvement, not a new model-quality, real-agent-success or LLM-savings result. A later semantic experiment must predict auxiliary-search usefulness from runtime-observable signals with separate data and evaluation criteria.
