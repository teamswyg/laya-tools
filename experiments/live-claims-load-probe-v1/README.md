# Local typing-demo load probe

[한국어](README.ko.md) · [Demo instructions](../../docs/live-claims-demo.en.md)

This probe checks the current Go classifier's basic processing cost for repeated cheap calls. **It is a local throughput measurement repeating one short sentence, not a semantic-quality evaluation.** No GPU or Laya weights are used. Other material-review work was running during the measurements.

## Conditions and results

- Apple M4 Pro, macOS 26.6.2, Go 1.27.1.
- Serial HTTP requests to the already-loaded local demo server. The public sentence is in `main.go`.
- 16 warm-up requests followed by 1,000 timed requests. Every request executes inference; predictions are discarded.
- The own Go model artifact is 73,988 bytes. Its SHA-256 is pinned in `result.json`. This is neither a new model nor a new training result.

| Metric | Run with response model-SHA verification |
|---|---:|
| Total time for 1,000 requests | 66.98 ms |
| Throughput | About 14,930 requests/s |
| Local HTTP median / p95 / p99 | 54.25 / 117.17 / 147.13 µs |
| Server model inference median / p95 / p99 | 2 / 4 / 6 µs |

`result.initial.json` retains the first measurement before response model-SHA verification was added to the probe: 67.89 ms for 1,000 requests, about 14,730 requests/s. Both runs used the same demo server. Model scores and expected labels are not evaluated in this experiment.

Server OS RSS snapshots were 10,064 KiB before the first run, 17,648 KiB after it, and 18,560 KiB after the second run. These are **point-in-time observations of server resident memory**, not peak memory, Go heap measurements, steady-state memory, or proof of input retention. Memory grew after HTTP processing; its cause is not established without separate heap/allocation measurements. Future Go heap/allocation profiling can use pprof, which does not measure GPU memory.

## Reproduce

Start the pinned model and server described in the [demo instructions](../../docs/live-claims-demo.en.md), then run from the repository root:

```sh
go run ./experiments/live-claims-load-probe-v1
```

The probe connects only to `127.0.0.1:8877`. It checks a two-second request timeout, bounded response size, response schema and model SHA, and emits no successful result on failure. The 180ms typing debounce and browser rendering are excluded. HTTP latency includes the Go client's response read and JSON decoding; server inference time excludes HTTP work. Server timing is reported at microsecond precision.

This is one input, serial calls and a warm model. Diverse lengths, concurrent users, fixed-arrival-rate open-loop load and cold model loading were not measured. It does not establish production throughput, minimum hardware, semantic accuracy or cost savings. The next performance experiment should separately examine allocations at multiple input lengths and latency/queue length under a bounded arrival rate. Semantic improvement still requires complete review of the 400 situations and subsequent training and evaluation.
