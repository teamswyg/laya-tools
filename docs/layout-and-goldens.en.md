# Array-oriented search and difficulty goldens: 2026-09-30

[한국어](layout-and-goldens.ko.md) · English

This experiment reduces repeated search work and records how Laya classifies real development tasks. Actions file caching is unchanged.

## Layout and measured tradeoff

Previously, search visited every chunk's string map and repeated query deduplication and BM25 logarithm/normalization work. The new immutable index maps words to integer IDs and stores separate `chunks []int` and `weights []float64` posting arrays. It precomputes each BM25 contribution and only accumulates postings for query terms. Sorting uses small ID/score records; source text is materialized for the top candidates.

This is SoA separation of hot numeric data from cold source text, without an ECS framework. Cache locality is a design intention, not a measured hardware cache-miss result. String-to-ID dictionaries, temporary build-time frequency maps, and tokenizer vocabulary/BPE maps remain: arbitrary key lookup is useful there, while dense arrays or linear search could cost more.

Apple M4 Pro, Go 1.27.1, 256 synthetic source files plus two existing fixture files; medians of five runs. Search timing excludes fixture/index construction; index build is measured separately. Baseline search implementation is main `2467752`.

| Metric | Before | After | Interpretation |
|---|---:|---:|---|
| Repeated search | 51,459 ns/op | 25,221 ns/op | About 51% lower time / 2.04× throughput |
| Search allocated bytes | 66,802 B/op | 23,250 B/op | About 65% lower |
| Search allocations | 22 | 25 | Three more allocations; distinct from byte volume |
| Index construction | 66.42 ms/op | 71.53 ms/op | About 7.7% slower |
| Index build allocations | 15.81 MB/op | 15.91 MB/op | About 0.6% higher; not retained heap |

This improves repeated retrieval, not total one-shot CLI or model inference latency by 2×. [Raw results](../benchmarks/results/layout-goldens-20260930) are public. Local pprof showed map access/assignment and the search path, but substantial macOS runtime/GC work too. Raw profiles are not published and do not establish a single dominant cause.

## SIMD and synchronization review

- Search reads sequential weights but scatters additions by document ID. Precomputation and removal of repeated hash lookup are the demonstrated improvements. No custom ARM64/x86 SIMD assembly was added. A future numeric kernel needs separate profiling, scalar parity, boundary tests, and architecture dispatch evidence.
- ONNX Runtime performs INT8 matrix operations. Existing x86 quantization precision protection remains. [ORT quantization documentation](https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html) describes architecture-specific instructions and saturation issues; this experiment did not trace a particular SIMD instruction.
- `Engine.mu` protects mutable tokenizer cache writes and the native session lifetime across `Predict` and `Close`. Predict is not read-only, so RWMutex is inappropriate. Shortening the critical section requires owned encoders and safe session lifetime management; replicated model workers consume additional RAM.
- `sync.Once` initializes the process-wide ORT environment; it is not considered a repeated-inference bottleneck. The process does not support switching runtime library paths after initialization.
- `App` belongs to a sequential CLI/JSONL/MCP handler. Engine's lock does not protect App initialization. Its concurrency contract is now explicit; parallel callers need separate ownership or external serialization.
- Search Index is immutable after construction, including caller-visible Chunks. Scores belong to each request. No shared scratch buffer or global lock was introduced.

Predict after Close now returns an error instead of accessing a nil native session. Native concurrent Predict/Close was tested with the race detector. Search scores/order are checked against an independent legacy map-based calculation, including parallel readers. Safety was not traded for speed.

## Fixed goldens and real development use

The [36-task fixture](../benchmarks/routing-golden.json) has expected difficulty and rationale fixed before inference. Labels are author judgments, not independent expert consensus or measured model coding success. Runs grow cumulatively 12 → 24 → 36 tasks, reusing one loaded engine per stage. Earlier results did not change labels, prompts, or the 0.9 confidence threshold.

| Tasks | English model judgments | Raw label matches | Accepted recommendations | Korean policy abstentions |
|---|---:|---:|---:|---:|
| 12 | 9 | 5/9 | 0 | 3 |
| 24 | 21 | 14/21 | 0 | 3 |
| 36 | 32 | 20/32 (62.5%) | 0 | 4 |

Of 11 expected-strong tasks in the final stage, raw predictions were strong twice, standard eight times, and fast once. Underestimation matters. Zero accepted under-routes does not establish quality when acceptance is also zero. Korean policy abstention is not counted as a correct classification. Each row records expected/raw/applied difficulty, confidence, abstention, and timing.

Several fixtures were used in this development work. Laya classified difficulty; this session's coding agent implemented changes. This was not an experiment assigning work to selected downstream models and measuring their success.

| Actual task ID | Expected | Laya raw | Outcome |
|---|---|---|---|
| closed-engine | standard | standard | Closed-engine error plus lifecycle/race coverage |
| golden-loader | standard | standard | Input/label validation and fixture tests |
| json-report | standard | standard | Confusion matrix and cumulative reports |
| search-soa | strong | standard | SoA search, reference parity, before/after benchmarks |
| lock-lifetime | strong | standard | Ownership/lifetime review and native concurrent shutdown test |
| simd-dispatch | strong | standard | Feasibility reviewed only; SIMD implementation not performed |

Not all 36 coding tasks were executed. Classification-only cases are distinct from completed implementation tasks. Future samples should fix labels before inference, add ambiguous tasks, short high-risk edits, and multilingual cases, and version changed fixtures.

## Reproduce and run in CI

```sh
go test ./internal/search -run '^$' -bench 'Benchmark(Search|Index)Layout' -benchmem -count 5
go build -o bin/routeeval ./cmd/routeeval
./bin/riidolaya setup
./bin/routeeval --stage 1
./bin/routeeval --stage 2
./bin/routeeval --stage 3
```

Run from the repository root; setup needs an existing or source-built riidolaya. Performance replay Actions also uploads all three JSON stages and search benchmarks. Probabilities can differ across platforms; local numeric outputs are not exact CI assertions. Regular CI checks fixture validity, behavior preservation, and races; performance replay records native decisions and metrics.
