# Shared raw token preparation: fewer allocations, preserved results

[한국어](README.ko.md) · [Optional Go API](../../../../pkg/hintprepared/README.en.md) · [Prewritten plan](PLAN.public.v1.md) · [Complete observations](OBSERVATIONS.actual.public.v1.json.gz)

`hintprepared.Prepare` prepares ordered raw tokens once and reuses query hash prefixes for cardinality and features. It does not substitute normalized text. Both cross-hash passes and the final copy remain paid work. Plans stay constructor-local; there is no new global cache, pool or lock.

Existing optional `Prepare` → `Rank` callers keep the same interface, reuse conditions and error contracts. Runtime code remains Go. Node only materializes and checks this experiment's pinned sources for maintainers.

## First measurement

The baseline is the previous [scratch constructor](../constructor-scratch/README.en.md), rather than the initial legacy implementation. One already exposed public Chi development parent supplies a request and five candidates. N is Rank calls after Prepare. Six pairs per N alternate order three times each: 24 intervals in total.

| Metric | Baseline scratch | Token plan |
|---|---:|---:|
| Go cumulative allocation per interval | 281,000 B | 244,792 B |
| Mallocs per interval | 662 | 241 |
| N=1 median paired elapsed ratio | 1 | 0.99017 |
| N=8 median paired elapsed ratio | 1 | 0.97795 |

All twelve pairs allocated **12.89%** fewer bytes and **421 fewer allocations (63.60%)**. Median elapsed time was about 0.98% lower at N=1 and 2.20% lower at N=8. The prewritten gates passed: no pair allocates more bytes, every pair saves at least 144 Mallocs, and each N's median paired time ratio is at most 1.00.

Three pairs were slower, with a maximum new/baseline ratio of about 1.0910. Baseline elapsed times also declined during repetition. Alternating order cannot eliminate warming and drift, so this is not evidence of universal speed improvement. These are unprofiled elapsed measurements, not CPU utilization.

Final **logical storage stays 170,648 B**: 10,589 features (169,424 B), 1,000 text bytes and a 224 B owner. Cumulative allocation is not peak memory, retained heap or RSS. Feature bits/order/cancelled zeros, five candidates, score/tie order, offsets, tight capacity and ownership are preserved. Original Features remains the oracle.

All 140 explicit wrappers returned without errors or panics: 26 Prepare and 110 Rank calls. Timed intervals include 24 Prepare and 108 Rank calls; input/model reads, validation, View creation and anchors are recorded separately. Later model-free feature checks and hidden internal calls are outside the 140-call count.

## Check saved results without a model

Run the saved check from the repository root with Go1.27.1 and Node:

```sh
scripts/verify-chi-token-plan.sh
```

Pinned temporary sources check complete gzip/CRC/EOF/JSON, call schedule, all feature/score bits and gates. No model file, download or new model scoring is needed. Use this saved check instead of rerunning the model. See [freeze](FREEZE.public.v1.json) and [first execution](EXECUTION.actual.public.v1.json). All 1,047,402 raw JSON bytes are preserved in 71,015 gzip bytes.

**The correct candidate remains fifth.** The existing inactive failed79 RIIDOH01 FP32 scorer has an immutable [Hugging Face archive reference](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46). There is no new training, Fit, checkpoint or default activation. This does not execute a Laya encoder, establish MPS/GPU performance or demonstrate LLM-cost savings. Twenty-four repeated intervals are not twenty-four independent Golden examples.

At writing, local controls and race checks passed; public CI and publication remain pending. Required CI and the bot gate govern merging; human approval does not replace them.
