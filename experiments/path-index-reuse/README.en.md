# Reusing one index pair — experiment 35

**Complete 2,400-task evaluation did not become faster and peak memory increased. Withhold default activation; provide an explicit experiment option only.** Distinguish the repeated-identical-catalog hypothesis from this evaluation mixing repositories and historical snapshots.

Freeze the [plan](plan-35.json) before implementation and measurement, against merged experiment 34. Preserve all original requests, complete catalogs, policies and task order. No training or label-based tuning occurs.

`fileeval.Ranker` reuses original and identifier-normalized BM25 indexes only when every original path matches in order. Neither a hash alone nor repository names establish identity. Retain one catalog, replacing it on changes. Cache no queries, rankings or labels. Returned rankings are independently owned and never overwritten by later calls. Each worker owns its instance, so no locks are needed; never call one instance concurrently. Preserve input bounds and fallback behavior.

| Execution order | Policy | Wall seconds | User / system seconds | Maximum RSS bytes |
|---|---|---:|---:|---:|
| 1 | Baseline | 52.59 | 53.88 / 2.06 | 108,150,784 |
| 2 | Reuse | 52.94 | 55.09 / 1.83 | 122,044,416 |
| 3 | Reuse | 52.72 | 55.33 / 1.88 | 128,991,232 |
| 4 | Baseline | 52.65 | 55.25 / 1.94 | 109,920,256 |

Execute sequentially on Apple M4 Pro CPU without overlapping benchmark jobs. Two-observation median wall time is 52.62 → 52.83 seconds; maximum observed RSS is about 104.8 → 123.0MiB. Do not infer statistical significance from such a small timing difference. Each reuse run has only three hits and builds 2,397 catalogs. Whole-process costs include input/cache validation, file reads, ranking, scoring and output, not pure inference. [All observations](results-35.json).

Every run reproduces experiment 34 aggregates byte-for-byte and the complete three-policy ranking-stream fingerprint for all 2,400 tasks. Frame task-ID length, policy ID, candidate count and every ordered candidate position into SHA-256. This verifies more than matching top-hit counts. Publish no requests or per-task rankings.

Precommitted acceptance requires identical full rankings/aggregates, lower median wall time and no increase in maximum observed RSS. Performance and memory criteria fail. Do not reorder tasks or enlarge the cache after observing this result and reinterpret it as the same successful experiment. Any batching change needs a separate experiment.

```sh
# Default: no retained index between calls
go run ./cmd/riido-fileeval --out .cache/file-eval-new
# Explicit experiment: retain only the previous complete catalog
go run ./cmd/riido-fileeval --reuse-path-indexes --out .cache/file-eval-reuse-new
```

Requires experiment 34's pinned local query/label projections and complete caches. Default execution also emits the full-ranking digest. Test invalid/changed inputs, caller array mutation, expanded-query fallback, preserved previous outputs and independent concurrent workers. Standalone `Rank` retains no cache between calls.

This concerns search execution costs only, not Laya/ternary inference or training, GPU, actual agent token savings or new quality evidence. Existing final reserves and model assets remain unchanged; no Hugging Face release occurs. Publish original synthetic fixtures, code and aggregates only; requests, path catalogs, labels and raw pprof files stay local.

A separate original synthetic fixture repeats the same 10,000 paths and query. Three benchmark samples with pprof enabled show baseline 18.90–19.77ms/op, about 29.16MB allocated/op and 80,249 allocations/op; reuse 0.1127–0.1130ms/op, about 0.423MB allocated/op and 21 allocations/op. Each benchmark includes one initial index build amortized across its iterations. Do not generalize this favorable repetition pattern to the actual 2,400-task result.

Combined pprof cumulative allocations also concentrate in ranking arrays and interleaving. Iteration counts differ substantially between policies, so combined allocation shares do not compare their memory efficiency. CPU samples heavily include runtime waiting and memory-return costs; they do not establish a need for SIMD or additional locks. Cumulative allocations are neither resident memory nor GPU memory. Full race, vet and formatting checks passed.
