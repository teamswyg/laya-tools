# Complete long-query input — experiment 32

**All 2,400 frozen real requests executed without truncation**, including 53 beyond the existing 8,192-byte bound. The search catalog is synthetic and used for execution-cost checks; this is not actual task retrieval accuracy.

## Usage

The opt-in Go `pkg/hintsearch` method `Index.RankLongInto(query, dst)` accepts complete UTF-8 input up to 128KiB. It tokenizes once and uses the same unique-term BM25 scoring and tie order as the existing API. No summarization, windows or truncation occur. Invalid input is rejected before destination mutation.

```go
index, err := hintsearch.New(documents)
if err != nil { return err }
ranking, err := index.RankLongInto(completeRequest, hintsearch.Ranking{})
if err != nil { return err }
// ranking.Order contains original document positions. Reuse ranking next time.
```

Existing `Rank`/`RankInto` retain their 8KiB bound. The index is immutable and output buffers are caller-owned; use separate buffers concurrently. Model, feature, CLI and session contracts are not enlarged. This supplies a complete lexical baseline for subsequent evaluation and Go callers, not long-input model integration.

## Execution

After fixing the [plan](plan-32.json), recompute the same frozen membership from pinned projections. Use 4,096 original documents in the form `document 0000 source code ...`. Most words are shared except document numbers, making this simpler than actual repositories. No answers or patches are read.

| Item | Result |
|---|---:|
| Requests / successful / failed | 2,400 / 2,400 / 0 |
| Requests exceeding 8KiB | 53 |
| Total / maximum original bytes | 4,363,135 / 71,136 |
| Truncated requests | 0 |
| Short-input parity checks / mismatches | 2,347 / 0 |

Compare complete score arrays and ordering exactly. The [aggregate](results-32.json) reproduces byte for byte. An independent BM25 calculation verifies a useful term beyond 8KiB. Tests also cover Unicode boundaries, bounds, rejection without buffer mutation and concurrent calls. Run full Go race/vet, formatting and redacted secret checks.

## Resources and limitations

On M4 Pro CPU, the final 2,400-query loop took 109,067,791ns (about 0.109s), 43,429,152 cumulative Go allocation bytes and 16,741 allocations. Caller result buffers were prepared using an original fixture. This measures ranking only, excluding source reads, membership calculation, indexing and short-input parity runs. Cumulative allocation is not peak resident memory.

The whole process took 0.92s wall, 0.56s user, 0.02s sys and 35,749,888 bytes maximum RSS (about 34.1MiB). Timing/allocation observations can vary and remain separate from deterministic result JSON. Synthetic catalog cost does not establish real repository/model/GPU/LLM costs.

The 53 long tasks no longer require exclusion from lexical evaluation. Next construct pre-fix file catalogs and measure search/hint effects against actual candidates. No routing, difficulty or decomposition quality is established here. No training, HF release or default activation occurred; final reserves remain untouched.
