# Reusing preparation for the same input

A compact claim hint can run repeatedly without rebuilding text features on every call. This experiment measures the existing Go `hintprepared` API with Go 1.27.1. Codex integration is optional.

| Same-input calls | One Prepare + N Rank, median | Prepare every call, median | BM25, median | Reuse / rebuild cumulative Go allocation |
|---:|---:|---:|---:|---:|
| 1 | 1.845ms | 2.059ms | 0.0156ms | 445,848 / 445,848B |
| 2 | 1.370ms | 2.805ms | 0.0154ms | 445,848 / 891,696B |
| 4 | 1.140ms | 4.525ms | 0.0244ms | 445,848 / 1,783,392B |
| 8 | 1.099ms | 8.158ms | 0.0434ms | 445,848 / 3,566,784B |

The paired rebuild/reuse time ratio at N=8 was7.219–7.578. At N=1, identical work varied0.843–1.581. These are six observations per method/N in one local process on one exposed development parent. Falling reuse times across N may reflect run-order drift; they are not a general scaling law. See all ranges in the [summary](SUMMARY.actual.public.v1.json).

Reuse reduces repeated setup, while BM25 remains cheaper here. The existing learned model still orders the fully passing candidate last. The earlier [whole-work pilot](../cost-pilot/README.en.md) required198 explicit verification calls with the model versus107 with BM25. This cost result does not repair ordering quality or demonstrate LLM token savings. The model remains inactive.

## Use

Audit retained results without downloading a model. Saved mode accepts plain JSON; decompress into an ordinary temporary file and clean it up.

```sh
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
gzip -dc experiments/short-claim/next-cohort-chi-audit/setup-reuse/OBSERVATIONS.actual.public.v1.json.gz > "$tmp"
go run ./experiments/short-claim/next-cohort-chi-audit/setup-reuse -saved "$tmp"
```

For a new measurement, supply the explicit existing32,792B RIIDOH01 FP32 model pinned by SHA in the [freeze](FREEZE.public.v1.json) and this [HF revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46). `model.hbin` below is a placeholder filename.

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/setup-reuse \
  -input experiments/short-claim/next-cohort-chi-audit/preview/INPUT.public.v3.json \
  -model model.hbin
```

This calls the Go hashed-feature scorer, not a Laya encoder. Retain any new output separately rather than replacing the frozen observation. Both humans and agents can distinguish saved auditing from new execution. API callers may Prepare once and Rank repeatedly for the same request and ordered candidate text. ID/provenance-only changes can reuse preparation; changed text requires a new Prepare. Scores are neither probabilities nor truth verdicts.

## Profiling and the next experiment

Four separate profiling processes used128 preparation iterations or262,144 repeated ranks. Most cumulative Go allocation was attributed to temporary `Features` construction and final `Prepare` storage. Rank CPU samples concentrated on coefficient reading/conversion. Only [public aggregates](PROFILES-AGGREGATE.actual.public.v1.json) are published; raw profiles remain local.

Next compare reducing temporary/final feature-array coexistence and an optional predecoded coefficient representation. Any array/SoA/SIMD change must preserve feature/score bits, ordering and error contracts while accounting for setup cost and retained storage. Current preparation is caller-owned immutable storage, with no shared cache or added locks.

The72 intervals and524,548 profiled rank calls repeat one parent, adding no independent Golden cases. Fit, labels and role changes are0; protected2400 evaluation was not read. TotalAlloc and pprof alloc_space measure cumulative Go allocation, not RSS/GPU memory.96MiB is a soft Go GC target. Preparation CPU sampling is too short for precise bottleneck percentages. Common raw input/model bytes, validated input and an owned View coexist even in BM25 intervals; this is not a cold first-use comparison. CPU profiling excludes initial setup; allocation profiles include process setup without subtraction, and loop parity guards are included.

[Korean](README.ko.md) · [Complete raw evidence](OBSERVATIONS.actual.public.v1.json.gz) · [Execution](EXECUTION.actual.public.v1.json) · [Caption review](QUALITY-CAPTION-REVIEW.public.v1.md) · [Existing notices](../NOTICE)
