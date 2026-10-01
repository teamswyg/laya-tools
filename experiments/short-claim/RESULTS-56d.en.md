# Actual round-trip cost of a resident Go hint tool: 56d results

[한국어](RESULTS-56d.ko.md) · [Raw results](results-56d.json) · [Numeric summary](summary-56d.json) · [Pre-execution design](PLAN-RESIDENT-56d.en.md) · [Execution plan](execution-plan-56d.json) · [Freeze record](freeze-56d.json)

**All 24 frozen execution rows completed. Without retries, 25,080 repeated requests were validated, and child lifetime peak RSS ranged from 9.21875 to 10.515625 MiB.** This measures sending a JSON line to a small Go process kept running and receiving a proposed verification order. It does not measure Laya model inference, training, semantic accuracy, or actual LLM usage savings. This local record was written before CI execution. After verification, actual CI completion links and automatic merge confirmation are recorded separately in the completion comment on [issue19](https://github.com/teamswyg/laya-tools/issues/19).

## Reading this result as a person or agent

The hypothetical user is an agent that repeatedly calls a small hint tool. It sends a request and 3 candidate captions to `riido-shortclaim --stream`, which suggests a verification order while retaining every candidate. Independent external checks must establish the actual meaning; responses remain `unverified_heuristic`. This experiment checks the cost of that calling path and the correspondence between inputs and outputs.

The child runs 4 **nonlearned Go baselines**: `fixed_order`, BM25, lexical unigram/bigram matching, and a narrow rule. It recomputes preparation, validation, ordering, digest, and output serialization for every request; result cache hits are 0. It uses no GPU, encoder, Laya weights, or model calls. These memory numbers therefore do not describe a complete Laya model. All 48 pool inputs are outside the narrow grammar, so every `narrow_rule` response records `unsupported_rule_request` and falls back to BM25. Successful narrow-grammar processing was not observed in this workload.

## Repeated observations versus independent data

From the 72 existing public original parents, only the **48 originally containing 3 candidates were selected**. The other 24 originally contained 2 or 4 candidates and were excluded; no candidate set was trimmed to 3. There were 0 duplicate normalized text features, so the pool also contains 48 payloads. The [corpus](wire-corpus-56d.json) retains all 72 origin links with original parent/candidate IDs and ordering. IDs and source provenance do not enter ranking features.

| Source | Original parents | Selected | Excluded |
|---|---:|---:|---:|
| Existing `probes-56` | 48 | 24 | 24: twelve 2-candidate and twelve 4-candidate sets |
| Typed `probes-56b` | 24 | 24 | 0 |
| Total | 72 | 48 | 24 |

`same` repeats the first payload. `distinct` cycles through the 48-payload pool in original order. Each baseline/workload uses 3 fresh child processes, giving `4 × 2 × 3 = 24` rows. Each row has 1 first response, 20 warmup requests, and 1,024 timed requests: 1,045 requests in total. Across all rows there are 24 first responses, 480 warmup requests, and 24,576 timed observations.

**The 25,080 requests and 24,576 timings are repeated processing observations, not new independent requests or training/final evaluation samples.** New independent requests, fits, model calls, new models, and protected-final reads are all 0 in this stage. The target of at least 2,400 distinct protected-final requests per domain remains unmet.

## Observed time, CPU, and memory

The native CPU path ran on an Apple M4 Pro with 24 GiB RAM, macOS 26.6.2, darwin/arm64, and Go 1.27.1. Builds use CGO=0 and trimpath. Both controller and child use GOMAXPROCS=1 and a 256 MiB Go heap soft limit. There is 1 child and 1 request in flight. These settings do not establish hard limits on OS threads, CPU affinity, or total RSS.

Each p95 below comes from a row's 1,024 timed **wire RTT** observations. The range is the minimum–maximum over 3 fresh-process rows under the same condition; it is not a pooled percentile or confidence interval. Wire RTT starts before writing the request and ends when a complete response line is received. Subsequent controller validation is recorded separately. RSS and CPU/request columns are rounded for display; the raw JSON retains exact values.

| Baseline | Workload | Range of 3 row wire RTT p95s, µs | Child peak RSS range, MiB | Child lifetime CPU/request range, µs |
|---|---|---:|---:|---:|
| `fixed_order` | same | 21.875–37.292 | 9.219–9.281 | 15.307–32.073 |
| `fixed_order` | distinct | 39.875–43.375 | 10.094–10.391 | 23.844–24.723 |
| `bm25` | same | 24.125–27.334 | 9.250–9.469 | 17.424–17.767 |
| `bm25` | distinct | 45.833–47.375 | 10.109–10.266 | 27.104–27.762 |
| `lexical_ordered` | same | 23.208–24.708 | 9.281–9.438 | 17.414–17.559 |
| `lexical_ordered` | distinct | 48.709–50.333 | 10.031–10.359 | 28.340–28.801 |
| `narrow_rule` | same | 23.916–25.666 | 9.250–9.453 | 17.111–17.605 |
| `narrow_rule` | distinct | 46.834–49.125 | 10.219–10.516 | 27.447–27.847 |

Summed user+system CPU across the 24 child lifetimes is **0.568013 seconds**. A row's CPU/request divides its entire lifetime CPU by **1,045**, including first response and warmup. It is not timed-phase CPU or pure inference time. Original Darwin RSS units are bytes, and the converted values are identical. The full child range is 9,666,560–11,026,432 bytes.

The separate controller used **0.2652 user seconds + 0.240851 system seconds = 0.506051 replay CPU seconds**. Its lifetime peak RSS was **34,717,696 bytes = 33.109375 MiB**. Controller RSS includes preparation and is neither per-row RAM nor a replay-only memory increase. Separate child and controller peaks cannot be added to claim a simultaneous total peak. Whole replay wall time was **1.482910833 seconds**. This wall/CPU includes per-row expected-stream hash preparation and cleanup. Hash preparation occurs before child startup and was separately recorded at 204.166–427.041 µs per row.

The ranges of per-row timed controller p95s were 8.083–17.458 µs from receipt through validation completion, 7.125–15.750 µs for validation work, and 1.042–2.125 µs for the receipt-to-validation-start delivery gap. Write-blocking p95s ranged from 1.375 to 2.333 µs. Do not add/subtract overlapping clocks or percentiles of different components to infer pure application time. **Pure startup, child processing, and pipe costs remain unobserved.**

The first row, `fixed_order/same/repeat 1`, took **577.13 ms** from the start call through its first response. The remaining 23 rows took 2.784792–3.267459 ms. The first observation was retained; its cause is unknown. A fresh process is not described as OS-cache-cold, and no cache flush ran.

## Did the result complete intact?

For all 24 rows, planned/attempted/fully-written/received/validated counts equal the phase targets. Failed, incomplete, not-attempted, and stderr-byte counts are all 0. Actual input/output stream SHAs match expected SHAs in every row. EOF, joined stdout/stderr readers, joined watcher, completed cleanup, and process reaping are recorded. Each row called `Wait` exactly once and exited with code 0. No cancellation, timeout, or cleanup error was recorded.

Recorded row wall times, from the start call through cleanup, ranged from 29.215084 to 617.744 ms, within the 15-second child budget. Replay completed within 60 seconds. Cleanup took 3.250–6.708 µs, within the shared 1,000 ms cleanup budget. Automatic retries were 0. The published raw result is **81,471 bytes** with SHA-256 `d24a50cf71968bb0336f58812ce02a647facff28cb40595e6cebbcd9587c4eaf`. JSON result size is within the 8 MiB publication budget; total repeatedly transmitted wire bytes are a separate quantity.

Expected protocol responses were prepared by the same public implementation. Matching hashes and complete rows demonstrate delivery/deterministic-output consistency, not independent semantic truth or a correct verification order.

## Freeze and limits of interpretation

Preparation sources are fixed at [60b0350](https://github.com/teamswyg/laya-tools/commit/60b03500fab417470d48d1a096be4ca1f43fafc2). Inputs, binaries, and the 24-row matrix were fixed before official execution at [891e52d](https://github.com/teamswyg/laya-tools/commit/891e52d83e9aaf39c87cb5ef55c6d492724bebf1). The [freeze record](freeze-56d.json) contains **31 implementation pins, including 5 pre-collection test files**, both binary SHAs, and recipe/corpus/plan SHAs. A separate freeze operation verified the 31 Git blobs; the CLI checks only commit format. A regression guard was added after collection and must not be retroactively described as a pre-execution test or frozen source. Record CI completion and automatic merge separately on [issue19](https://github.com/teamswyg/laya-tools/issues/19).

Input sizes differ between workloads. The `same` line is 385 bytes including LF. Pool lines range from 385 to 1,165 bytes, averaging 675.6875 bytes; normalized requests and candidates each range from 4 to 32 words. The first request has 8 words and its candidates have 5/7/8 words. The same/distinct difference therefore does not isolate input reuse or cache effects. Raw requests span 36–236 bytes and captions 28–252 bytes, so this is not a worst-case measurement across every possible 512-byte input or language. Normalization removes punctuation; feature SHA does not establish a safe future semantic/rule/model cache key.

[56a](RESULTS-56a.en.md) retains its first cold observation of **370.932333 ms** and the separate profile's `runtime.kevent` **83.02%** flat share with unresolved cause/attribution. That experiment used 8 candidates and an `io.Discard`/stage benchmark; 56d uses selected 3-candidate inputs and actual resident pipe round trips. A direct speedup percentage is not justified. This experiment does not establish improvements from caching, SIMD, SoA, or locks, or any LLM savings.

Semantic/utility work remains separate in [the 56c plan](PLAN-56c.en.md). Open-loop concurrency, matched-byte echo controls, a ready signal, and caching/invalidation need separate follow-up designs. Preserve the scope of **measuring the calling cost of a small nondecisive hint**. Semantic accuracy, whole LLM usage, and a low-resource model's weights/runtime memory each need their own evidence.
