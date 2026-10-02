# Packed Score 78 v2 preparation

This draft makes one small change: **select the packed storage format outside the feature loop**. FP32, INT8, and ternary each use a serial loop. `New`, the coefficient API, owned-byte copying, wire format, prefix table, validity/default-state behavior, and payload accounting remain unchanged. The original v1 directory, failure logs, and test-correction history were not modified.

Each loop preserves feature order and duplicates, reads the coefficient as FP64, and accumulates `coefficient * value`. Scale is not moved after accumulation for INT8 or ternary. Even a zero ternary coefficient is multiplied, preserving `0 * Inf` and `0 * NaN`. Signed zero and NaN payloads were compared bit-for-bit with the synthetic reference. An invalid index after a valid prefix still returns `ErrIndex` and +0. No maps, locks, global cache, or SIMD were added.

On Go 1.27.1 Darwin arm64, the existing 5 top-level tests and 38 subtests passed in one race run. Vet and the library build each passed once. There were 0 new test failures. Checks cover nil/zero views, invalid indices across all formats, Inf/-Inf/NaN, signed zero, duplicates, ownership after caller mutation, and concurrent readers. Reference `Decode` and `Score` were called only on public synthetic bytes. No HF model payload or training data was read.

| Asset | Status |
|---|---|
| `view.go` | Only Score changed to three format-specific loops |
| `view_test.go` | Boundary cases strengthened within the existing 5 top-level tests |
| `view_bench_test.go` | Exact v1 bytes; 0 executions here |
| `go.mod` | Exact v1 bytes; excluded from verbatim publication because it contains a host path |
| Earlier v1 source and failure records | Original bytes and hashes preserved in their existing location |

There are **0 new performance measurements** in this preparation. The parent observed higher CPU cost for the earlier v1 packed path, but this change does not establish or isolate the cause and has no measured speedup. The parent plans one later comparison using the unchanged four-format benchmark, 266 synthetic features, count 5, and 200 ms timing. A library build creates neither a runnable experiment nor a new model.

Owned payload accounting is unchanged and does not imply lower Go heap or RSS. Caller input bytes and owned packed bytes may coexist. Synthetic bit equality on Go 1.27.1 is not a proof for every platform or compiler. Laya/GPU execution, native observation reruns, Fit, Features, Project, new weights, training labels, shared-repository changes, and remote mutations are all 0. Production readiness and default activation remain false.

Final pins are in [HANDOFF.v2.json](HANDOFF.v2.json); preparation attempts and the retained failure history are in [ATTEMPT-LEDGER.v2.json](ATTEMPT-LEDGER.v2.json). Raw test JSONL and the module file containing a host path are excluded from publication.
