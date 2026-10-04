# Packed costs with an actual reference model

The existing Go `hintweights` reader was connected to the same public FP32 file, request and five captions. All 15 float64 score bits and three ranking orders equal the Decode path. **The bad legacy-first recommendation also remains unchanged.** Representation cost and recommendation quality are separate; keep the model inactive.

| Scope | Decode/reference | Packed Coefficient | Packed Score + copy scratch |
| --- | ---: | ---: | ---: |
| Model preparation Go allocation/op | 65,536B;1alloc | 41,024B;2allocs | Same View constructor |
| Prepared five-score allocation/op | 0B;0alloc | 0B;0alloc | 0B;0alloc |
| Full five-rank allocation/op | ~130,616B;253allocs | ~127,704B;239allocs | ~127,704B;239allocs |
| Logical coefficient payload | 65,536B | 32,792B | 32,792B |
| Additional conversion scratch | None | None | 20,976B |

View payload includes the 24-byte encoded header. Borrowed original file bytes, Go containers, allocator padding, common prepared-feature storage, whole heap/RSS and GPU memory are excluded from logical payloads. Both representations coexist in one process, so this does not demonstrate whole-memory reduction. Including scratch leaves an 11,768B logical difference; the no-conversion path differs by 32,744B.

One fixed input was measured across four alternating-order rounds and 32 windows: 64 load, 1,024 prepared-score and 128 full-rank operations per window. Prepared five-score median observations were 4.24/18.23/6.78µs; full ranks approximately 448/456/445µs. These are **not intrinsic representation speed comparisons**: per-feature versus per-score instrumentation is asymmetric, reference Rank repeats validation and the packed adapter supports only the frozen validated input. Repetitions are not independent samples. No production latency, whole-task or Codex cost improvement is established.

## pprof and the next target

A separate replay generated CPU/allocs profiles and successfully parsed both with Go pprof. Unprofiled evidence remains unchanged. Sampled alloc_space attributed 74.72% flat and 85.07% cumulative to Features. This includes initialization, gates, all 32 windows and profiling; it is a cumulative allocation estimate, not exact live/peak/RSS. The short CPU profile is heavily affected by runtime kevent/madvise and cannot establish a CPU/SIMD bottleneck. Only [aggregates and limits](PROFILE-READBACK.actual.public.v1.json) are public; raw profiles stay private.

## Verify and optionally reproduce

With Go 1.27.1, verify saved evidence without reading/downloading models or training:

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/packed-real-preview/verify.sh
```

With an existing exactly pinned model file, repeat the actual measurement:

```sh
MODEL_FILE=/absolute/path/data-only8192-fp32-seed1729.hbin \
  bash experiments/short-claim/next-cohort-xxhash-audit/packed-real-preview/reproduce.sh
```

Times may differ; the script checks all score bits, orders, call schedule and logical payload. The [additive HF follow-up](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/6ed4e42ea3bbd0549752451e1320a434717d4c8c/followups/xxhash135-136-134b76a) links both projects. All 22 original file metadata and the weight blob stayed unchanged; only four public evidence/docs files were added and read back byte-for-byte. No new checkpoint or Laya/MPS/GPU training. Existing scratch-model provenance and Apache-2.0 code scope apply; complete unresolved ancestry qualification is not claimed.

## Next PDCA

Separate an opt-in caller-owned prepared-feature experiment for repeated identical request/candidates. Preserve original tokenization, sort/collision accumulation and float64 values. Compare tightly sized AoS first, then uint16-index/float64-value SoA: 16→10 logical bytes per feature, with uint32 cumulative offsets. Each caller owns its storage, without shared maps/locks. Record preparation plus 1/2/8/64-repeat costs and actual len/cap; holding prepared features can increase memory for one-off use.

One nonblind post-hoc parent, no new roles/labels/Fit/protected 2,400 evaluation/default activation. Existing 37/HF37 remain unchanged. Read the complete 4,557B observations, [pre-execution freeze](FREEZE.public.v2.json) and [readback](READBACK.actual.public.v1.json) separately.
