# Retaining complete file candidate catalogs

An interim observation of 264 available catalogs found 102 above the existing 4,096-document limit, with a maximum of 6,924 files. This is not the final 2,400-task result. Freeze the [scaling plan](scaling-plan-33.json) before implementation rather than selecting small repositories or truncating files.

## Opt-in Go API

`hintsearch.NewPathIndex(paths)` accepts up to 100,000 unique canonical relative UTF-8 paths, 4,096 bytes each and 16MiB total, matching collector ceilings. Empty, duplicate or invalid paths fail explicitly. Input order is retained; files and symlinks are never opened. Permissions and historical completeness remain caller responsibilities.

```go
idx, err := hintsearch.NewPathIndex(paths)
if err != nil { return err }
ranking, err := idx.RankLongInto(completeRequest, hintsearch.Ranking{})
if err != nil { return err }
order, err := hintsearch.InterleavePathBaselineFirst(ranking.Order, hintIDs)
if err != nil { return err }
// order refers to original paths positions and retains every candidate.
```

Use global BM25 statistics across the complete catalog, not incompatible per-shard scores. The index remains immutable with array vocabulary/postings and caller-owned reusable score/order buffers. Concurrent calls require separate buffers. Large-catalog merging preserves baseline-first behavior and every candidate. With accurate equal-cost verification, a baseline target at rank r appears within min(n, 2r−1) inspections. This is not a time/token or multi-evidence completion guarantee.

Existing constructors and merge APIs retain their small bounds. Model, feature, CLI, session and deferred-helper contracts are unchanged. This supplies a file-path lexical baseline and candidate ordering, not code-body semantics or task-solving quality.

## Validation and profiling

Find the useful target after the 4,096-candidate boundary in a 6,000-path catalog and compare its score with an independent global-IDF calculation. Also test the last candidate at the maximum 100,000-path bound. Small inputs preserve exact scores/order. Reverse hints preserve the complete permutation and baseline rank bound. Invalid paths, duplicates and excessive catalogs are rejected.

On M4 Pro CPU, run three repetitions using 10,000 original synthetic file paths and a repeated single-word query. CPU/allocation profiles remain local. Observed ordering comparison/swap costs motivated replacing result ordering's `sort.Slice` with `slices.SortFunc`; scoring is unchanged.

| Measurement | Before | After |
|---|---:|---:|
| Reused-query ns/op range | 43,054–43,559 | 35,039–35,084 |
| Reused-query B/op | 72 | 16 |
| Reused-query allocations | 3 | 1 |

Query time decreased about 19% in this fixture. Both runs used the same profiling options, but this is not a causal experiment eliminating system-load/order effects. Do not extrapolate a synthetic single-word query to real long requests, repositories or LLM workloads.

Construction measured about 5.33–5.84ms before and 5.01–5.08ms after, with approximately 9.47MB cumulative allocation and 10,113 allocations unchanged. Construction was not the target of the ordering change, so its timing variation is not claimed as an improvement. Cumulative allocation is not resident memory. No SIMD/GPU speedup is claimed. Full collection is complete; actual file-localization scoring follows separately.
