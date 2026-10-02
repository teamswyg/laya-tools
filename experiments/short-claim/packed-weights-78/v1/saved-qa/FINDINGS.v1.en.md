# Review of saved packed78 benchmark results

On these public synthetic inputs, median packed scoring took about4.08–5.10 times as long as reference FP64-array scoring. Construction was faster and allocated fewer bytes per operation, but repeated scoring of already-constructed coefficients did not show a speed benefit. This measurement does not approve a default-path change or model quality.

The reviewer did not author packed78 and is nonblind because of prior role62, input and observer work. The earlier independent source-review artifacts remain unchanged. This review checked saved numbers and execution bindings; it did not newly run trained models, Features, Fit, Decode, Score, benchmarks or tests.

## Scoring: the same266 public feature entries

Each entry is the median of five samples within one original invocation. Ratio is packed ns/op divided by legacy ns/op; a positive time delta means slower.

| Format | Legacy ns/op | Packed ns/op | Time ratio | Time increase |
|---|---:|---:|---:|---:|
| FP32 | 155.8 | 635.9 | 4.0815× | +308.15% |
| INT8 | 156.7 | 680.6 | 4.3433× | +334.33% |
| Ternary13 nonzeros | 155.8 | 794.1 | 5.0969× | +409.69% |
| Ternary8192 nonzeros | 155.7 | 776.2 | 4.9852× | +398.52% |

Both implementations report0 B/op and0 allocs/op in every scoring sample. Packed coefficient conversion, prefix/popcount, kind dispatch and bounds checks have different costs from array loads. This experiment did not isolate those components and therefore does not prove that any one caused the slowdown.

## Construction: different owned representations

| Format | Legacy ns/op | Packed ns/op | Time delta | Legacy→Packed B/op | Allocated-byte delta | Legacy→Packed allocs/op |
|---|---:|---:|---:|---:|---:|---:|
| FP32 | 15,888 | 6,772 | -57.38% | 65,536→41,024 | -37.40% | 1→2 |
| INT8 | 8,883 | 4,240 | -52.27% | 65,536→9,536 | -85.45% | 1→2 |
| Ternary13 | 9,521 | 366.8 | -96.15% | 65,536→1,504 | -97.71% | 1→3 |
| Ternary8192 | 12,906 | 419.4 | -96.75% | 65,536→2,656 | -95.95% | 1→3 |

Reference construction owns8192 float64 values; packed construction owns a copy of wire bytes plus a ternary prefix. This compares kernels that later read the same coefficients from different representations, not functions producing identical representations. Fewer allocated bytes and fewer allocations are different measures: packed allocated fewer bytes but made more allocations.

Calculated owned data lengths are legacy65,536 B, packed FP32 32,792 B, INT8 8,216 B, ternary13 1,308 B and dense2,330 B. These slice lengths include the24 B wire header and258 B prefix but exclude Go object headers, allocator rounding, caller input and feature buffers. B/op is observed cumulative allocated bytes per operation. For example, FP32 data length32,792 B differs from measured41,024 B/op. Neither is total live heap or OS RSS.

## Execution bindings and limits

Saved ledger and before-start records confirm one original root benchmark invocation, Start1, joined Wait, exit0, retry0 and no timeout/overflow. All16 benchmark cases×5 repeats=80 rows are present in the frozen legacy→packed order. METRICS preserves every sample's original iteration count. Those reported loop counts do not reconstruct exact original API calls including calibration/setup or count independent data cases.

One successful stdlib metadata helper, with zero failures, verified14 file bindings. Pins match for the frozen plan, prototype/test/benchmark source, reference Decoder/Score source, handoff, prior independent receipt, actual test-binary bytes/SHA and buildInfo, actual ledger/before-start, and raw benchmark/OS-record hashes. BuildInfo matches Go1.27.1/darwin arm64/CGO0/trimpath with no VCS metadata. Arguments match CPU1,256MiB soft Go heap,60-second outside timeout,50-second Go testing timeout, run `^$`, count5/200ms/benchmem. BuildInfo and saved-record bindings are not independent dynamic environment instrumentation or a compiler-trust proof.

The combined process recorded real19.76 seconds/user18.73 seconds/system0.24 seconds, controller wall19.764369208 seconds, maximum RSS11,173,888 B and peak footprint8,520,184 B. This single RSS includes both implementations, setup and the Go testing runtime. There were no isolated representation processes, so it does not establish RAM saving. The256MiB soft heap setting is not an OS RSS hard cap.

The target was Apple M4 Pro with Go1.27.1/darwin arm64. Repeated warm loops ran in a fixed order in one process; file/startup state was not independently controlled and background load was unmeasured. Five-sample medians are not statistics across independent hosts or data. Results do not generalize to other CPUs, serving, file I/O, cold start or GPU/SIMD. Model quality, training success, LLM usage/cost saving and production/default activation were not established. First-fit71 counters were not overwritten.

Safe artifacts are the [numeric metrics and raw hashes](METRICS.v1.json), [receipt](RECEIPT.v1.json), [separate ledger](ATTEMPT-LEDGER.v1.json) and these bilingual findings. Private raw stdout/time, binary, helper source, host paths and trained-model contents are not copied into them.
