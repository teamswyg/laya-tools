# Resident resource experiment 56d: uncached JSONL round trips

[한국어](PLAN-RESIDENT-56d.ko.md) · [Existing 56a resources](RESULTS-56a.en.md) · [56b audit](RESULTS-56b.en.md) · [Next semantic/utility path 56c](NEXT-56c.en.md)

**This is a pre-execution design. The final machine plan, input and binary freeze, official resource measurements and completed CI evidence do not exist yet.** Counts and budgets below are proposals, not observations. This document records **0 new measurements**. Current 56b performance runs, model calls, fits and new weights remain 0.

The fictional user is an agent repeatedly using a small Go hint tool. Instead of starting a process for every request, it sends one JSON line to a resident process and receives one proposed inspection order. This experiment asks about the whole round trip's CPU, memory and time. Semantic ranking quality and actual LLM usage reductions need separate utility evaluation.

## Existing observations and scope

[56a](performance-56a.json) used the same eight-candidate example for three cold and one warm process per control. The first cold observation was **370.932333 ms**; the other 11 were 2.564958–5.232667 ms. Initial-launch and OS/file-cache effects remain unverified. The [separate profile](profile-summary-56a.json) reported `runtime.kevent` at 0.88 flat seconds, **83.02%**; attribution remains unresolved. Do not establish an application bottleneck or shared-lock contention from it. Preserve the original records, numbers and first observation.

The current [stream](../../cmd/riido-shortclaim/main.go) performs JSON preparation, validation, ranking, input digest and actual output serialization for every line, with no result cache. The [benchmark](../../cmd/riido-shortclaim/bench.go) uses `io.Discard` and repeated stages; it cannot run with `--stream`. `--cpuprofile` is also benchmark-only. Do not rename 56a's warm p95 as resident pipe round-trip latency. Offline `go/types` provenance-audit costs are separate from runtime hint costs.

## Inputs to prepare and freeze before timing

Use the existing [48 requests](probes-56.json) and [24 requests](probes-56b.json), totaling 72 original public authored requests, as preparation sources. Their `FeatureInputs` projections extract **only the request and ordered candidate descriptions**. Prototype, source, contract, target and role are absent from ranking inputs. Typed-source and language checks happen during preparation; the resource driver reads only prepared runtime JSONL.

1. Retain projections that originally have three candidates. Do not remove, add or duplicate candidates, or truncate descriptions to force a count of three. Verify at most 512 bytes of raw and normalized text and 32 normalized words for each request/candidate, and at most 12 KiB per JSON line.
2. Propose `feature_sha256` as SHA-256 of a **length-prefixed, ordered** encoding of the normalized request, candidate count three and three normalized descriptions. Freeze its encoding and implementation hash before execution. Exclude IDs and provenance. Changing only description IDs or whitespace does not create a new feature payload. This is a conservative deduplication criterion, not proof of equivalent outputs for all controls or a safe future cache key.
3. Deduplicate feature SHA values in stable original order. Publish selection/exclusion lists, original-parent bindings, request/candidate byte and word counts, and the number of distinct payloads in a preparation manifest. **Official machine-plan freezing and execution are blocked until at least two distinct feature payloads are verified.** Additional payloads are not additional independent labels, training groups or final requests.
4. `same` repeats the frozen first payload. `distinct` cycles through the entire deduplicated list in stable order. Freeze the first list position for the timed phase too. Each workload uses the same byte sequence across all controls. Publish differences in sizes and word distributions; do not attribute differences causally to input reuse alone.
5. Prepare runtime-envelope schema, IDs and provenance as public constants or flat identifiers. Freeze each line's raw SHA and existing input digest. IDs are metadata for response matching. Record `feature_sha256`, JSON raw SHA and the metadata-inclusive input digest as different values.

A separate preparation check creates expected protocol responses and freezes platform-specific response-byte SHA values. These verify input correspondence and deterministic output; the actual resident tool recomputes every request. Record actual response SHA values too. Do not assume identical floating-point response bytes across platforms. Protocol preparation is not an experiment observing truth labels or oracle utility.

## Small execution matrix and budgets

| Proposed setting | Value and meaning |
|---|---|
| Controls | `fixed_order`, `bm25`, `lexical_ordered`, `narrow_rule` |
| Workloads and repeats | Three fresh resident processes for each same/distinct workload: `4 × 2 × 3 = 24 children` |
| Requests per child | 1 first response + 20 warmup requests + **1,024 timed requests = 1,045 requests** |
| Planned total | 25,080 requests; 24,576 timed observations. Repetitions are not new development/final data |
| Processing | One child and one inflight request; check response correspondence before sending the next request |
| CPU and Go heap | GOMAXPROCS=1 for child and controller; 256 MiB Go-heap soft limit |
| Deadlines | 15 seconds per child, 60 seconds for the whole official replay; 0 automatic retries |
| Public result storage | At most 8 MiB of new results; no unlimited accumulation of full responses |

GOMAXPROCS limits Go's simultaneously executing Ps; it does not establish OS thread count or CPU affinity. A Go-heap soft limit is not a hard cap on total RSS, native memory or GPU memory. Freeze the CPU-only path without an encoder, GPU, model or result cache. Freeze termination and pipe-cleanup policy after the driver is implemented and checked.

`distinct` also cycles a distinct-feature pool. A child's 1,024 timed requests are not necessarily 1,024 unique requests; 24,576 is the proposed timing-observation count. Add neither count to independent training or final samples.

Freeze the 24-row execution order in advance. Record the decision to alternate the existing control order and same/distinct order before timing. Preserve the first child and each child's first response separately; do not discard them or replace them with another control's warmup. A fresh process is not an OS-cache-cold process, and no cache flush is performed. Building, binary verification/copying and corpus preparation are separate preparation work.

## Observable times and unobserved costs

| Metric | Boundary to freeze |
|---|---|
| Spawn-call wall | From just before the controller's start call to its return; this is not a child-ready signal |
| First response | From just before the start call to the first complete response line; startup, first processing, pipes and controller are combined |
| Wire RTT | From starting a request write to receiving its complete response line; includes write wait, child work, pipes and receive framing |
| Write blocking | From starting a request-line write until all bytes have been written |
| Controller validation | From receiving the complete line to completing JSON, digest and candidate-preservation checks |
| Warm distribution | Wire RTT p50/p95/max and completion count for 1,024 requests after warmup |

Continuously drain bounded response framing, timestamp line completion and then check correspondence. The current stream has no ready signal or child-internal timestamps. Therefore **pure startup, child-internal-only processing latency and pure pipe cost remain unobserved**. Metrics can overlap; do not add or subtract them to manufacture application time. The existing `io.Discard` benchmark is not justification for simply subtracting pipe time either.

## CPU, memory and protocol verification

At child termination, record user/system CPU and **that child's whole-lifetime peak RSS**. Record the controller's `RUSAGE_SELF` user/system CPU increases and separate lifetime peak RSS. To compare controller RSS by row, use a fresh controller per trial and keep orchestration costs separate. Check that controller CPU does not count child CPU again. Peak RSS is not an end-minus-start delta, instantaneous warm memory or Go heap. Summed process peaks do not establish a simultaneous total peak. Distinguish Darwin bytes from Linux KiB, convert to bytes and retain the original units.

A completed child's CPU/request denominator is **1,045**, including the first request and warmup. The planned warm-latency denominator is **1,024**. Whole-process CPU is not warm-phase-only CPU. Each row retains planned, attempted, completed, failed and incomplete counts and input/output byte totals. Do not hide failures, timeouts or protocol mismatches from denominators or fill them as successes. An incomplete row is not a normal 1,024-sample result. Preserve its raw CPU and actual per-phase counts instead of filling in a completed 1,045-request CPU/request value.

Every response must be one bounded JSON line. Check schema, `unverified_heuristic` status, control kind, frozen input digest, response raw SHA, finite scores, and three candidate IDs as an exact permutation of the input. For out-of-scope `narrow_rule` inputs, verify the frozen BM25 fallback and `unsupported_rule_request`/`unsupported_rule_candidate` reason. Malformed input is not normal fallback. Publish fixed enums and termination/cleanup status, without stderr, reader errors or invocation arguments. Protocol checks do not certify the candidate functions' semantic correctness.

## Execution gates and later work

Review and check the new driver's source, tests and manifest, input projection/feature-hash encoding, platform-specific protocol expectations, 24-row order and termination policy. **Freeze the binary built with Go 1.27.1, `CGO_ENABLED=0` and trimpath: its buildinfo, binary-byte SHA, source/input/output pins, new driver SHA and machine execution plan must be fixed before official collection.** Execute the exact verified bytes to avoid replacement between reads, and require fresh outputs. Record actual CI status and links after CI runs. This document has no ready-to-execute badge or success link.

The **six matched-byte echo/framing controls** belong to a **later separate attribution plan**. Do not silently add them to this 24-row matrix or subtract their latency from actual round trips. If exact startup separation is needed, design a versioned ready signal separately. Open-loop arrivals with a bounded queue, concurrency, caching/invalidation, and lock/SIMD/SoA changes require later independent plans. Preserve unresolved kevent attribution and the first cold observation; current cache, SIMD and speed gains are unestablished.

Public outputs are original public fixtures, plans, numbers, hashes and scope explanations. Real work inputs, personal absolute paths, credentials, binaries and raw pprof/traces are not published. Training, model calls, weights, production activation and protected-final/CoSQA access are 0 in this design. The separate goal remains **at least 2,400 distinct protected-final requests per domain**.
