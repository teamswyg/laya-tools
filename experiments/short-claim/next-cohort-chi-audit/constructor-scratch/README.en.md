# Constructor scratch reuse: fewer allocated bytes at a time cost

[한국어](README.ko.md) · [Optional Go API](../../../../pkg/hintprepared/README.en.md) · [Prewritten adoption gates](PLAN.public.v1.md) · [Raw observations](OBSERVATIONS.actual.public.v1.json.gz) · [Numeric summary](SUMMARY.actual.public.v1.json)

`hintprepared.Prepare` now reuses one constructor-local cross-feature buffer instead of retaining a large temporary buffer per candidate. A cardinality prepass sizes the final array exactly, then each compact row is copied immediately. The returned owner never shares mutable scratch. There is no global cache, pool or new lock.

The comparison uses **one already exposed public Chi development parent**. Each N has six pairs, with three executions in each order. N is the number of Rank calls after preparation. Every interval pays full Prepare and N Rank calls, including the new tokenization and cardinality prepass.

| Interval | Legacy Go cumulative allocation | Scratch Go cumulative allocation | Bytes saved | Median paired scratch/legacy time | Allocation count |
|---|---:|---:|---:|---:|---:|
| N=1 | 445,880 B | 281,032 B | 36.97% | 1.0773×, about 7.7% slower | 376 → 664 |
| N=8 | 445,992 B | 281,144 B | 36.96% | 1.0751×, about 7.5% slower | 383 → 671 |

All twelve pairs saved allocated bytes. The experiment passed its prewritten memory-first gates: at least 20% fewer allocated bytes in every pair and no more than 25% median paired time increase for either N. This is not a speed improvement. Paired time ratios range from 0.9679–1.1537 for N=1 and 1.0366–1.1163 for N=8. Allocation counts increased by 288. Reducing repeated tokenization and small string allocations is the next hypothesis.

The final owner is unchanged: 10,589 features in a tight 169,424 B array, 1,000 cloned text bytes and a 224 B fixed owner. **170,648 B is logical storage**, not allocator usage, measured retained heap or RSS. The table reports interval `TotalAlloc` and `Mallocs`, including recording-wrapper costs. It does not establish a 37% decrease in peak memory, whole-process memory or CPU/GPU utilization.

The first actual run preserves all 24 intervals and every candidate score. All 140 explicit wrappers returned without errors or panics: 26 Prepare and 110 Rank calls, including separately recorded setup/anchors. Main intervals contain 24 Prepare and 108 Rank calls. Input/model reads, input validation and View construction are separately recorded. Original Features remains unchanged; the legacy Prepare body and shared Rank retain the prior bytes. All feature indices, float64 bits and order, score bits and tie ordering match.

**The correct candidate remains fifth for the existing model.** These inactive RIIDOH01 FP32 hashed-feature weights are not the Laya encoder. There is no new fitting, checkpoint, default routing activation, independent Golden parent, protected evaluation, GPU or MPS execution. Twenty-four repeated intervals are not twenty-four independent tasks. This does not demonstrate agent completion-time or LLM-cost savings.

## Check the saved result without a model

From the repository root with Go1.27.1:

```sh
go test ./pkg/hintprepared -run '^TestConstructorSaved$' -count=1 -v \
  -constructor-saved "$PWD/experiments/short-claim/next-cohort-chi-audit/constructor-scratch/OBSERVATIONS.actual.public.v1.json.gz"
```

The checker verifies a single gzip frame, CRC and complete EOF; exact JSON fields and array lengths; the complete call schedule, rankings and cost gates. It separately checks the public input hash and recomputes **all features with unchanged Features**, binding rows and logical sizes to an independent witness. These later model-free oracle calls are separate from the 140 live wrappers. No model file is read. Default tests use synthetic numeric weights and failure controls; actual costs require all three explicit input/model/output flags.

See [freeze](FREEZE.public.v1.json), [source pins](SOURCE-PINS.public.v1.json), [independent source review](SOURCE-REVIEW.public.v1.json), [synthetic checks](FAKE-CONTROLS.actual.public.v1.json) and [first execution](EXECUTION.actual.public.v1.json). All 1,046,200 raw JSON bytes are preserved in 70,899 gzip bytes. No raw profiles, private paths or model weights are stored here. The earlier [pprof aggregates](../setup-reuse/PROFILES-AGGREGATE.actual.public.v1.json) belong to a separate experiment; this run collected no profiles.
