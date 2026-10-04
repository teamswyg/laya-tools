# Actual verification cost of Chi candidate ordering

A small claim model proposes an order in which to check candidates. This experiment asks whether that proposal reduces **the actual cost of reaching a verified answer**. If only one of five URL handling candidates satisfies the contract, proposing it earlier can avoid unnecessary checks. Every selected candidate must still pass the same nine checks.

This is a development experiment on **one exposed public Chi parent**. Thirty-two repeated pipelines do not create thirty-two independent examples. It uses the existing inactive failed FP32 model, with no new training, labels, role assignment or default activation. It does not execute the Laya encoder or a GPU. Broader usefulness and Codex savings require independent cases and actual task evidence.

## Actual results

All32 pipelines verified `post_path` with all nine checks. Unavailable checks and fallbacks were zero. The model nevertheless proposed the correct candidate fifth and required more verification work. **The existing model remains inactive.**

| Order | Candidates visited | Fixture checks | Explicit Chi API calls |
|---|---:|---:|---:|
| BM25, fixed, lexical | 1 | 9 | 107 |
| Existing model | 5 | 16 | 198 |

The following are medians of four observations per policy. In-process time and Go allocations use matching boundaries.

| Policy | Fresh-process interval | Reused interval | Fresh-process Go allocation | Reused Go allocation |
|---|---:|---:|---:|---:|
| BM25 | 262.6µs | 53.5µs | 190,056B | 64,752B |
| Fixed | 261.8µs | 35.9µs | 190,264B | 64,720B |
| Lexical | 263.7µs | 56.2µs | 190,376B | 64,568B |
| Existing model | 1,606.5µs | 64.4µs | 816,424B | 114,392B |

Within-block model/BM25 interval ratios were **5.10–6.57** for fresh processes and **0.75–1.68** for reused state. The reused model interval was shorter in two pairs and longer in two, providing no consistent speedup or general performance finding. Each reused block separately required **1.37–1.60ms** and approximately **701,800B of Go allocation** for setup. A short interval that excludes setup does not establish a reduction in complete cost.

Externally measured fresh-process median time was 4.43ms for BM25 and 5.88ms for the model. Within-block complete-process model/BM25 ratios were 0.89–1.39: three longer model processes and one shorter. Reused process wall time of 6.58–7.26ms belongs to preparation and all four policies; it is not attributed to individual policies. These are few repetitions of one parent, without statistical generalization.

Each reused block runs four different policies once after setup. It does not measure preparation amortization across repeated calls to the same model. Its model interval excludes setup already paid by that block.

Actual Chi execution comprised 45 precheck and 344 timed trials, totaling 389 trials and 4,685 explicit API calls. Actual model ranking occurred once in the anchor and eight times in timed pipelines, totaling **nine real calls**. Separate model-free synthetic controls covered seven controller roots and eleven worker roots. Full records are in the [compressed observations](OBSERVATIONS.actual.public.v1.json.gz); the [derived summary](SUMMARY.actual.public.v1.json) retains individual costs and paired ratios.

## Comparison

The policies are BM25, fixed order, lexical order and the existing model order. Each retains all five candidates. A known contract mismatch stops checking that candidate and moves to the next. Incomplete capture or execution failure remains `unavailable`; it cannot become a pass. Only a candidate passing all nine checks can be `verified`.

A typed disagreement in normal termination or a return value is a known mismatch only when the explicit call protocol and value captures are complete. Missing, truncated or nonreturned calls withhold verification instead.

A separate precheck reruns all 45 combinations using actual Chi APIs and compares the full typed results with the prior observations. A separate anchor checks the model score bits and the four baseline orders. Every visited check inside a timed pipeline executes Chi APIs again. There is no cached verdict substituting for verification.

The two lifetimes are:

- Fresh process per pipeline: four policies in four order blocks, sixteen pipelines. Input reading and model preparation are included in that pipeline.
- Prepared state reused within a process: four fresh block workers each run four policies, sixteen pipelines. Each block records its input/model/feature setup cost separately.

Williams orders balance each policy across launch positions one through four and balance directed adjacent policy pairs within blocks. A fresh process does not imply a cold operating-system file cache.

## Reading the measurements

Candidate visits and explicit Chi API calls show additional work, but APIs have different costs. Call counts alone do not establish a speedup. The record also includes pipeline time, Go allocation bytes and allocations, and externally measured complete process time.

Process time includes startup, garbage collection and result encoding. Within-process measurement places garbage collection and memory-stat reads outside the interval; verification and trace collection remain inside. Go allocation measurements are neither total resident memory (RSS) nor GPU memory. Reused preparation is reported separately rather than treated as free.

Measured workers use `GOMAXPROCS=2` and `GOMEMLIMIT=96MiB`. The latter is a Go garbage-collector memory target, not an enforced cap on total process memory.

## Reproduction and evidence

The public Go controller in `run` verifies source/input/model SHA-256 values and the frozen plan, then uses a bounded temporary workspace. The caller explicitly supplies an existing model file. Model weights are not included in GitHub.

With Go 1.27.1, audit the saved evidence without reading a model file or launching new verification workers:

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/cost-pilot/run \
  -saved experiments/short-claim/next-cohort-chi-audit/cost-pilot/OBSERVATIONS.actual.public.v1.json.gz
```

To rerun the actual pilot, set `RIIDO_MODEL_FILE` to the local file for the pinned model below. Output creation refuses to overwrite an existing result.

```sh
go run ./experiments/short-claim/next-cohort-chi-audit/cost-pilot/run \
  -model "$RIIDO_MODEL_FILE" -out chi-cost-local.json.gz
```

Model: [existing failed FP32 artifact at its immutable JooYoon revision](https://huggingface.co/JooYoon/riidolaya-shortclaim-data-effect-failed-79/tree/5bef215895b69d3f2ef4b82bb3f1970279d67f46). Its size is 32,792 bytes and SHA-256 is `dff05140098845943ece87c3ab8a31a15564a4b170a59ab017ee393501158004`.

The source and plan are sealed before the first pilot. The original observer is reused with only its `main` declaration renamed. All candidates, unvisited markers, call traces and failure reasons are retained. Synthetic controls reject absent calls and altered results; those controls are not model performance examples.

Read the parent [Chi experiment](../README.en.md) and notices together. Selected Chi MIT and Armon notices are retained in the parent `source` directory. Permission for this finite source execution is managed separately from admission of the entire source family to training.
