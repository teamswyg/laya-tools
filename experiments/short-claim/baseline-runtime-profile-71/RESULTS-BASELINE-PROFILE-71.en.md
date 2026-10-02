# One CPU and memory measurement of the public baseline

The existing public `riido-shortclaim` was built once with Go1.27.1/CGO0/trimpath/buildvcs=false and executed once on the public eight-candidate fixture with `lexical_ordered`, 10,000 samples per stage and CPU profiling. Retries, models, fits, new labels and paid calls are zero. The original71 fit counters and results were unchanged. This measures neither validation quality on the actual76 corpus nor learned-model inference latency.

| Stage | p50 | p95 | Go allocations/op |
|---|---:|---:|---:|
| JSON parsing, validation, normalization | 19.708µs | 45.542µs | 184 |
| Prepared input revalidation, features, score, order | 21.334µs | 31.666µs | 81 |
| Input digest | 0.875µs | 1.167µs | 23 |
| Ranking serialization | 0.458µs | 1.250µs | 2 |
| Full Go request | 42.500µs | 83.042µs | 291 |

Besides10,000 latency samples, each stage performs20 warmups and101 allocation probes, including Go testing's one warmup and100 measurements. Each callback therefore runs10,121 times, totaling50,605. Successful-path source arithmetic yields20,243 calls each to Load/Rank/ValidatePrepared,40,486 to Validate,364,374 to NormalizeText for nine texts, and20,243 selected-baseline tokenizations. These are separate benchmark workload counts derived from source, not dynamic instrumentation. All-four `Baselines` and learned `Features` are not invoked. The existing stage name includes nonlearned lexical processing under features.

Whole-child OS observations are real1.47s, user0.99s, sys0.01s, peak RSS **13,189,120 bytes (about12.58MiB)** and peak footprint10,568,184 bytes. Controller elapsed time is1.476222375s, while CLI internal measurement is1.009231333s. Internal timing excludes profile shutdown/flush/close; outer timing includes startup and shutdown, so their scopes differ. GOMAXPROCS1, Go heap soft256MiB and an outer30-second SIGKILL guard were used, without a timeout. The guard waits after killing and does not guarantee total controller termination within exactly30 seconds.

The ending Go HeapAlloc snapshot is3,221,544 bytes; total allocation increase is1,009,991,512 bytes. Around1GB of cumulative allocations does not mean1GB was resident simultaneously. The snapshot includes uncollected garbage, no forced GC occurred, and it is distinct from OS peak RSS. Profiling was enabled for this single execution, so its overhead was not isolated.

CPU pprof was read once. The displayed top has `runtime.kevent` at flat730ms/77.66%, and `internal/lexicalhint.tokens` at flat40ms/4.26%, cumulative140ms/14.89%. The profile combines warmups, allocation probes and all five stages, showing sampled function attribution. This ranking alone does not establish the cost of lexical arithmetic or a SIMD benefit. Present evidence motivates considering revalidation/normalization and full-request allocations as separate improvement hypotheses; no functionality or normalization boundary was changed here.

No GPU or learned encoder/model was used. CPU pprof, Go heap statistics and OS RSS are not GPU telemetry and do not measure GPU time or memory. This is a repeated public fixture measurement, with no quality, token-savings, generalization or deployment approval claim.

The safe numeric summary is `COMPACT-RUNTIME-PROFILE-71.v1.json`,7,493 bytes/SHA256 `b0d0fd85fb082b50a564b0974177b857bda8e62856966518f7459c54796ef414`; symbol-only CPU top is `CPU-TOP-AGGREGATE-71.v1.txt`,672 bytes/SHA256 `d65e066a980195e2044a32a593af6446306b4caba1c91047076e5917a4d28576`. Original input text, profiles, raw logs and private paths were not published to Git/HF. One initial lookup of the supplied name `benchmark.go` failed because the actual file is `bench.go`; the correct source was read before proceeding. Build1/benchmark1/pprofread1/metadata-summary1 all succeeded.
