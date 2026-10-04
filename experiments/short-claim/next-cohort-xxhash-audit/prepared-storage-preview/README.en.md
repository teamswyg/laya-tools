# Prepare and reuse features for the same request

One public request and five captions were connected to the existing packed Go model using caller-owned arrays. **All 15 score bits and every order stay identical, including the bad legacy-first recommendation.** This is a repeated-computation cost experiment, not model quality, training or Codex savings.

Fresh rebuilds the original Features each use, copies into reused scratch and calls View.Score. AoS keeps index/value pairs together; SoA keeps uint16 indices and float64 values in separate columns. All methods use the same result tuple, bit/order checks and whole-invocation counters. Each caller owns its buffers; there is no shared map, lock or global cache. The View is common.

| Logical persistent storage | Fresh scratch | AoS | SoA |
| --- | ---: | ---: | ---: |
| Data length/capacity bytes | 20,976B | 89,168B | 55,730B |
| Fixed owner layout | 32B | 48B | 72B |
| Interface handle | 16B | 16B | 16B |

Fixed owner bytes already include slice headers, offsets and pointers. Do not add those breakdowns again. SoA feature payload is 33,438B (37.5%) smaller than AoS, but holds 34,754B more than fresh scratch. These are logical 64-bit Go layout/capacity calculations, excluding allocator padding, View containers, input/result/report storage, heap/RSS and GPU. The common encoded View payload is 32,792B; the borrowed original model remains live at len 32,792/cap 40,960. Reference Decode is used only for an initial gate and released before measurement. Initial coexistence peak reduction is not established.

## Include preparation in the comparison

One fixed input, four alternating-order rounds. Median window times below exclude input validation and model loading. Reuse measures 1,024 uses after preparation; construct_repeat includes one construction plus N uses. This does not guarantee warm CPU caches.

| Construction + uses | Fresh total time / Go allocation | AoS total time / Go allocation | SoA total time / Go allocation |
| --- | ---: | ---: | ---: |
| 1 | 490µs / 149,360B | 478µs / 217,728B | 478µs / 189,088B |
| 2 | 910µs / 276,912B | 459µs / 217,728B | 471µs / 189,088B |
| 8 | 3.65ms / 1,042,224B | 467µs / 217,728B | 523µs / 189,088B |
| 64 | 29.71ms / ~8,185,160B | 752µs / 217,728B | 1.42ms / 189,088B |
| Prepared reuse: per five-score use | 453µs / ~127,552B | 4.76µs / 0B | 12.52µs / 0B |

One-off construction allocates more with either prepared form. For this input, total allocation improves from two uses; the difference grows at 64. AoS was faster here, while SoA retained a smaller feature store. SoA calls a bounds-checked coefficient API per feature, so speed differences cannot be attributed to layout/cache hits alone. General performance, production latency, whole-task success and LLM usage remain untested.

Reuse only when request, all caption text and candidate order are unchanged; otherwise construct a new owner. Do not keep an unbounded request cache. Explicit capacity/lifetime limits would precede any long-lived cache.

## Saved verification and optional actual replay

Go 1.27.1 saved-only verification reads no model, downloads or training assets. The additional CI step uses this mode; existing separate native Laya CI remains unchanged.

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/prepared-storage-preview/verify.sh
```

With an existing exactly pinned FP32 file, repeat the measurement:

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/prepared-storage-preview/reproduce.sh
```

Timing, allocations and heap snapshots may vary; bits, order, call schedule and logical storage remain fixed. Preserve the [pre-execution freeze](FREEZE.public.v1.json), [original observation](observations.actual.public.v1.json) and [readback/summary](READBACK.actual.public.v1.json) separately. Successful setup/nested counts are source-derived; constructor/five-score attempt/normal-success pairs are explicit counters. No full per-call panic ledger exists. Heap snapshots include all common process objects and are not peak/RSS or per-method memory. No new pprof/GPU run.

One nonblind post-hoc parent; repeated uses are not independent samples. No new roles/labels/Fit/protected evaluation/default activation. The existing failed reference is a scratch Apache-2.0 research asset with retained provenance, not complete unresolved ancestry qualification. No weights, raw profiles or private input are committed to GitHub.
