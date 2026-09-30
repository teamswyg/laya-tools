# Real relevance training 02: compact exports, no useful transfer yet

[한국어](TRAIN-02.ko.md) · [Pre-run plan](train-plan-02.json) · [Complete aggregates and model hashes](train-results-02.json) · [Previous 2,400-pair evaluation](PAIR-01.en.md)

**Actual CoSQA binary relevance training ran in Go. Validation AUC ranged from 0.475 to 0.513, providing insufficient ranking signal. No default promotion; reserve1 (2,403 pairs) and reserve2 (2,401) were not scored this cycle.** Deferring the final test is a development judgment after validation, not a preregistered statistical pass/fail rule. Smaller ternary files do not demonstrate better quality.

## Training contract

New `internal/pairlearn` and `cmd/riido-pairtrain` minimize binary cross entropy on query/code relevance, separate from the original synthetic candidate-ranking trainer. This is neither generative training nor tuning the Laya encoder. The independent compact probe uses 8,192 coefficients over signed hashed word/bigram crosses and overlap. It has no intercept; scores are not validated probabilities.

The previous partition is unchanged. Of 12,189 development pairs, 11,263 fit the input scope and 926 were excluded from training. Of 607 validation pairs, 553 were used for selection and 54 excluded. Results below cover **eligible validation pairs only**. Future final evaluation must retain long inputs using BM25 fallback. All source rows are read for provenance and grouping, but reserve features, scores, and labels are not used for fitting or selection.

Two seeds × two learning rates × FP32/ternary STE gave eight configurations, each trained for 20 epochs with batch128 and L2=0.0001. Minimum validation loss selected epoch and learning rate separately per mode and seed. Selected FP32 weights yielded INT8 and ternary PTQ exports; independently selected STE completed eight exports. AUC is a post-selection diagnostic, not an independent test.

## Findings

AUC measures whether relevant pairs rank above irrelevant pairs; approximately 0.5 indicates little useful ordering on this dataset. These validation point estimates have no confidence intervals, so small differences do not establish superiority.

| Method | Validation AUC seed1729 | Validation AUC seed2718 | Export bytes |
|---|---:|---:|---:|
| FP32 | 0.4937 | 0.5045 | 32,792 |
| INT8 | 0.4935 | 0.5053 | 8,216 |
| Ternary PTQ | 0.5028 | 0.5133 | 1,633 / 1,636 |
| Ternary STE | 0.4751 | 0.5062 | 1,639 / 1,630 |

Selected FP32 training loss dropped to approximately 0.680, while validation remained approximately 0.693. Constant probability 0.5 has loss approximately 0.693147; a constant development positive-rate predictor has validation loss 0.693824. Successful optimization is not useful transfer. This indicates weak generalization here, not proof that only a large model can work.

## Resources and reproducibility

- Apple M4 Pro, 24GiB RAM, Go1.27.1, CPU. First complete run: wall3.87s, userCPU3.52s, maxRSS236,814,336B (225.84MiB), swaps0. Includes source loading, feature preparation, eight fits and exports; not single-inference cost.
- Sparse features use contiguous offset, uint16 index and float64 value columns. Logical index/value payload: 70,556,830B (67.29MiB), excluding capacity overhead, offsets, labels, source and runtime. A per-dataset 320MB payload ceiling bounds the cache.
- Separate profiling replay: sampled live Go heap approximately 80.94MiB, preparation arrays approximately 95.48%; CPU flat samples approximately 35.83% sparse scoring, 21.50% Fit, 4.67% preparation. Do not sum overlapping cumulative times. pprof is neither total RSS nor GPU measurement. No GPU or SIMD acceleration was implemented/used.
- Single-owner training arrays require no locks or shared mutable state. Global ternary scale is recomputed each batch; backward uses identity STE. Packed files still decode to 64KiB of float64 coefficients.
- Profiling and AUC-diagnostic replays reproduced all eight original model SHA256 hashes. Aggregates include source, plan, partition and model hashes. Raw profiles and weights remain outside Git.

```sh
go run ./cmd/riido-pairtrain --out .cache/pair-training-new \
  --cpu-profile .cache/pair-training-cpu.pprof \
  --heap-profile .cache/pair-training-heap.pprof
```

Prepare the pinned local CoSQA source as described in the previous report. Output directory must be new. There is no final-evaluation stage in this command; exit0 means execution success only. Exports share the existing `riido-hints --model ... --sha256 ...` format but are not recommended for production.

## Next experiment and publication

Next compare learned low-dimensional lexical/length controls and ablations separating cross-feature collisions from overfitting. A separately licensed static semantic embedding experiment is a possible subsequent branch. Record repeated validation use; seal settings before a single reserve1 evaluation. Candidate retrieval, actual task outcomes and LLM-call savings still require separate evidence.

The CoSQA authors release code under MIT and data under C-UDA. This cycle used local computation and public aggregates only. No source rows, labels or model binaries were published to GitHub. New-weight Hugging Face publication requires provenance, license/content review and a model verification path; this report does not claim publication.
