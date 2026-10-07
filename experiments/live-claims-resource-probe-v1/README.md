# Live claim-hint resource probe v1

[한국어](README.ko.md) · [Exact measured Go helper](replay/main.go)

This is a bounded development measurement of one existing, unqualified
three-claim Go model. It measures CPU resource use and repeated hint throughput
on eight newly authored, unlabelled inputs. It does not establish semantic
accuracy, generalization, a production SLA, or actual Riido/LLM cost savings.

The model was loaded and predicted with CPU-only Go code. The measured binary
was built with `CGO_ENABLED=0`; this inference path did not require a GPU.

## Results

All six fixed runs completed successfully. The table shows the median across
three runs, followed by the minimum–maximum across those same runs. Quantiles
are calculated within each run; they are not pooled across runs.

| Measure | Direct in-process `Predict` | Sequential local HTTP client |
| --- | ---: | ---: |
| Whole measured-loop throughput, calls/s | 229,626 (212,105–253,939) | 14,235 (12,890–14,246) |
| Mean per-call latency, µs | 3.599 (3.250–3.890) | 69.319 (69.264–76.588) |
| Per-run p50 latency, µs | 3.500 (3.375–3.625) | 72.833 (65.625–72.958) |
| Per-run p95 latency, µs | 5.000 (3.709–6.250) | 90.917 (90.250–118.375) |
| Per-run p99 latency, µs | 5.834 (4.667–7.625) | 109.875 (99.750–161.417) |
| Maximum whole measured-process RSS, MiB | 12.281 (12.203–12.297) | 16.766 (16.672–16.812) |
| Warm ready Go heap allocation, bytes | 687,688 (687,688–687,848) | 322,112 (321,680–322,272) |
| Retained Go heap after measured calls and GC, bytes | 688,208 (688,208–688,368) | 323,504 (323,120–324,768) |
| Net Go heap change after model verification/load and GC, bytes | 81,208 (81,208–81,208) | 81,208 (81,208–81,208) |
| Timed allocation, bytes/call | 0.000533 (0.000533–0.000533) | 8,622.347 (8,620.880–8,624.027) |
| Timed allocations/call | 0.000033 (0.000033–0.000033) | 82.060 (82.042–82.073) |

Each 30,000-call direct loop recorded **16 bytes and one allocation in total**,
including the measurement harness. This is near-zero observed allocation over
the loop, not an assertion that `Predict` alone allocates exactly zero. The HTTP
client allocation median is 8,622.347 bytes/request, or 8.420 KiB/request.

The direct latency timer brackets the Go call. Its throughput denominator also
includes per-call timers, checksums, repeated-output guards and loop work. The
HTTP latency brackets request construction, loopback transport, server work,
response reading, JSON decoding and output-pin validation. These measurements
do not isolate network overhead, and their different budgets and instrumentation
do not support a model-speedup claim.

The live server separately reported integer `inference_us` values. Across the
HTTP runs, the median of the per-run means was 4.510 µs (4.505–4.935 µs). These
values have microsecond quantization and a different timing boundary from the
client stopwatch. They must not be subtracted from client timings to infer an
isolated transport cost.

## Memory and CPU scope

`/usr/bin/time -l` measured maximum RSS for each new probe process. The direct
arm's RSS covers its in-process helper. The HTTP arm's RSS covers **the client
helper only**, including a duplicate locally verified/loaded model with an
81,208-byte net heap increment. It excludes the already-running demo server.
Neither RSS column is an isolated model-size measurement.

Go `MemStats` counters report Go-managed memory and allocation, separately from
whole-process RSS. The retained heap includes sampling arrays and bookkeeping.
The direct harness retains two 30,000-element `int64` arrays (480,000 bytes of
elements); the HTTP harness retains two 600-element arrays (9,600 bytes of
elements). Model/workspace stack memory and runtime/code pages also affect the
process measurement. Go memory counters do not measure native/GPU memory; no
native inference or GPU execution was used here.

A separate maintainer-supplied `ps` observation of the already-running demo
server showed RSS 12,064 KiB = 12,353,536 bytes = 11.78125 MiB, at approximately
16m50s process elapsed time, with displayed CPU 0.0%. An exact UTC timestamp was
not recorded. This is an informational snapshot, **not a measured peak or a
time-matched server measurement** for these six runs. It is excluded from the
main results table and does not establish loaded-server memory under this probe.

The external process CPU and wall fields are rounded to 0.01 seconds. They
include startup, model verification/loading, warm-up, explicit GC, measured
calls and output. They cover the HTTP client rather than server CPU. The first
direct process had 0.75s external wall time despite a 0.141440s measured loop;
later direct processes had 0.14s and 0.12s external wall time. No whole-host CPU
utilization or sustained-load result is claimed.

| Arm/run | Measured calls | Loop elapsed, ms | Calls/s | Mean latency, µs | External real/user/sys, s | Max RSS, bytes |
| --- | ---: | ---: | ---: | ---: | --- | ---: |
| Direct 1 | 30,000 | 141.440 | 212,104.698 | 3.890 | 0.75 / 0.15 / 0.01 | 12,877,824 |
| Direct 2 | 30,000 | 130.647 | 229,626.324 | 3.599 | 0.14 / 0.13 / 0.00 | 12,894,208 |
| Direct 3 | 30,000 | 118.139 | 253,938.607 | 3.250 | 0.12 / 0.12 / 0.00 | 12,795,904 |
| HTTP 1 | 600 | 46.547 | 12,890.324 | 76.588 | 0.06 / 0.02 / 0.02 | 17,481,728 |
| HTTP 2 | 600 | 42.118 | 14,245.663 | 69.264 | 0.05 / 0.01 / 0.02 | 17,580,032 |
| HTTP 3 | 600 | 42.150 | 14,235.044 | 69.319 | 0.05 / 0.01 / 0.02 | 17,629,184 |

## Fixed protocol and pins

- Apple M4 Pro, 14 physical/logical CPUs, 24 GiB RAM, macOS 26.6.2 (25G83),
  darwin/arm64. Other work could have been active on the host.
- Exact Go 1.27.1; `CGO_ENABLED=0`, `GOWORK=off`, `GOTOOLCHAIN=local`,
  `GOPROXY=off`, `GOMAXPROCS=2`, `GOMEMLIMIT=128MiB`.
  `GOMEMLIMIT` is a soft Go runtime memory limit, not a hard whole-process cap.
- Model: 73,988 bytes, 4,240 recorded training steps, SHA-256
  `cfd35ee70a23f94a8d470b7dea596244a91ac7f1410475e8a42959693c99864b`.
  Supplied lineage: HF semanticcontrast parent21. No model/seed/parameter
  selection or training was performed by this probe. Every process checked the
  explicit model's SHA-256 **before** calling `Load`.
- One sequential worker and one reusable direct workspace. Three new processes
  per arm. Fixed order: direct 1, HTTP 1, direct 2, HTTP 2, direct 3, HTTP 3.
- Per direct run: 128 warm-up + 30,000 measured calls. Total: 384 warm-up +
  90,000 measured = **90,384 direct calls**.
- Per HTTP run: 32 warm-up + 600 measured POSTs, using one loopback keep-alive
  client, plus status checks before/after. Total: 96 warm-up + 1,800 measured =
  **1,896 POSTs**, plus **six GET status checks = 1,902 HTTP requests**.
- The HTTP target was the existing loopback instance at `127.0.0.1:8877`.
  No server was started or restarted. All six status checks matched the model
  SHA, steps, research-preview mode, unqualified semantic status, one worker,
  fixed confidence/margin floors 0.9/0.05 and temperature 1.
- Each arm had a 55-second internal work deadline. Measured loops took
  0.042–0.141 seconds; there was no sustained-load phase.
- Eight newly authored development inputs: five English and three Korean,
  with no labels, fixed round-robin order and no content-based selection.
  Their UTF-8 byte counts are **52, 54, 50, 51, 51, 64, 63, 71** (mean 57).
  JSON body counts are **63, 65, 61, 62, 62, 75, 74, 82**.
  Length-framed corpus SHA-256:
  `0a81f50a71ff1e8e7a988188034cb1989bbe0f41fec92fe21350ed8cbfa2b3b1`.
- The first eight predictions' checksum matched across all six runs:
  `ac8b22815a94d3b3`. Full-loop checksums were
  `8e2766d9b9cfd615` for every direct run and `6d7cb4a547b59523` for every HTTP
  run. Different full-loop counts explain the different checksums. Every later
  occurrence of an input was checked against its first prediction. Checksums
  consume numeric and decision fields; they prevent unused-output benchmarking
  and check repeatability, not semantic correctness.
- Repository HEAD before measurement:
  `798edb6157fe3723a11f18f54718b58e156b6ddb`. Relevant current-source SHA-256
  pins, helper/binary pins and raw-record hashes are in
  [measurements.json](measurements.json).

These inputs are exposed development material and are excluded from the
protected final 2,400-example dataset and model/selector tuning. This resource
probe did not read protected research data. No application-state writes,
training, threshold changes, model releases or raw profiling captures occurred.

The exact measured Go helper and all eight original unlabelled phrases are now
published in [replay/main.go](replay/main.go). Its SHA matches the measured helper
pin; no diagnostic unit fix or new measurement run was used. These fixtures
contain no private work data. Model artifacts and raw local logs stay outside
Git. Use the instructions below to replay the fixed protocol; machine load and
timing results will vary.

## Replay

From the repository root, install the pinned public model using the
[demo instructions](../../docs/live-claims-demo.en.md). The exact helper reads
`.cache/live-claims-demo/model/claims.rsc` and verifies its SHA before loading it.
Start the existing demo at `127.0.0.1:8877` for the HTTP arm. The helper deliberately
preserves the original fixed path and the recorded server-unit naming issue.

```sh
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local GOPROXY=off \
  go build -trimpath -ldflags='-s -w' -o bin/claim-resource-probe \
  ./experiments/live-claims-resource-probe-v1/replay
```

Use Go 1.27.1 and `GOMAXPROCS=2 GOMEMLIMIT=128MiB`. Run the fixed sequence
`inprocess 1`, `http 1`, `inprocess 2`, `http 2`, `inprocess 3`, `http 3`, each in a
fresh process. For example, on macOS:

```sh
GOMAXPROCS=2 GOMEMLIMIT=128MiB /usr/bin/time -l \
  ./bin/claim-resource-probe -arm inprocess -repeat 1
GOMAXPROCS=2 GOMEMLIMIT=128MiB /usr/bin/time -l \
  ./bin/claim-resource-probe -arm http -repeat 1
```

Do not pool within-run quantiles, subtract server time to claim isolated network
cost, add unmatched process RSS peaks, or interpret these fixtures as accuracy
data. The public helper is a maintainer experiment and is not an application
runtime dependency. The published measurement was not rerun for publication.

## Recorded unit issue

The private measured helper reused its nanosecond distribution structure for
server-supplied integer microseconds. In the immutable raw JSON records,
`server_inference_reported_us` therefore contains subfields named `mean_ns`,
`p50_ns`, etc., whose **numerical values are actually microseconds**. The
public aggregate names those fields `mean_us`, `p50_us`, etc. The raw records,
measured source and measured binary were preserved; no source rewrite or new
measurement run was used to conceal the diagnostic issue. Numeric observations
were not changed.

[measurements.json](measurements.json) contains normalized aggregates and all
six per-run values. It summarizes the fixed runs without selecting a best run.
