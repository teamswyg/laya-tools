# Reusing validated input: public fixture measurement

The previous CLI validated and normalized the input while loading it, then repeated those steps during ranking. The new `ValidatedInput` retains the validated fixed array and strings in private fields. Its constructors use the existing validation, and ranking reuses the existing tokenization, scoring and ordering kernel. The zero value is rejected; the returned `Prepared` is a value copy that cannot mutate the retained state. Public `Rank(Prepared)` and `Baselines` keep their existing validation. The type represents a shape and normalization contract, not semantic truth or action authority.

Exact order, score, fallback and CLI schema/input digest parity passed for all four baselines. Invalid constructor inputs, JSON-forged zero objects, modified `Prepared` values, mutations of caller slices and returned copies, concurrent reads, and existing stream bounds/privacy checks also passed. Targeted Go 1.27.1 race and vet checks ran once each. No map, lock, model, training or SIMD was added.

The measurement used the existing public eight-candidate `lexical_ordered` fixture, 10,000 repetitions, one CPU and a 256 MiB Go soft memory target. There was one build, one measurement and no retry. New profiles and new pprof reads were both zero.

| Metric | Previous observation | New observation |
|---|---:|---:|
| Parse allocations/op | 184 | 184 |
| Rank allocations/op | 81 | 0 |
| Full request allocations/op | 291 | 210 |
| Rank p50 | 21,334 ns | 8,708 ns |
| Full request p50 | 42,500 ns | 29,250 ns |
| Go cumulative allocation bytes | 1,009,991,512 | 656,806,680 |
| Go HeapAlloc bytes at end | 3,221,544 | 3,911,912 |
| Whole measurement process peak RSS bytes | 13,189,120 | 10,928,128 |
| Whole measurement process peak footprint bytes | 10,568,184 | 8,421,880 |
| OS real/user/sys seconds | 1.47 / 0.99 / 0.01 | 1.05 / 0.67 / 0.01 |

The previous observation enabled CPU profiling; the new observation disabled it. Each version has one sample, with ambient host load and scheduling unrecorded. Parent full checks were coordinated to start after notification that the new measurement had completed, but other concurrent load was unobserved. Latency therefore does not establish a causal optimization speedup. End HeapAlloc increased and includes uncollected garbage without forced GC. Cumulative allocation, end heap and OS RSS are different metrics; the soft target is not an OS hard cap. There were no GPU, native encoder or learned-model calls and no GPU memory measurement.

Successful-path source arithmetic gives the same 50,605 stage callbacks. Validation calls fall from 40,486 to 20,243 and normalization calls from 364,374 to 182,187. These are source-derived counts, not dynamic instrumentation, and do not modify the first71 fit counters. This public fixture cost observation proves neither actual76 validation quality, learned-model quality nor LLM token/billing savings.

The implementation and comparison helper share an author; these are author-side mechanics checks, not independent quality approval. The source origin is `9d204c2c700505658c108297d5fa769a835a60c3`; the six changed files were [frozen](SOURCE-FREEZE-71.v1.json) before measurement. The [comparison JSON](COMPARISON-VALIDATED-INPUT-71.v1.json) and [attempt ledger](ATTEMPT-LEDGER-VALIDATED-INPUT-71.v1.json) record exact original byte/SHA references, limits and counts. Raw input, binary, raw profiles/logs and host paths remain private. This task performed no external publication or push.
