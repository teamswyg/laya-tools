# Original-weight Laya code validation 19

**Original-weight FP32 execution also fails the current relevance gate.** Forward-order AUC is 0.5008, compared with fixed INT8 0.5040 and matched BM25 0.5378. Removing INT8 conversion alone does not establish improvement for this checkpoint, protocol and task. Do not adopt the current head outputs as useful distillation targets.

## Matched protocol

Commit and freeze the [plan](plan-19.json) before inference. Use exactly the CoSQA source, grouped partition and 607 validation cases of experiment 17b. **553 distinct questions** meet the existing 4,096-byte/64-word scope; 54 do not. No extra tokenizer exclusions. Two option orders produce 1,106 calls, not 1,106 distinct questions.

Load F16-stored original weights from `tindang/laya-code@25f97e5a2ec5f8cf7218a4f67504367d8832e1fe` into FP32 CPU. Verify model, config, tokenizer and installed Laya source hashes against experiment 18 pins. Do not execute Python code from the remote model repository.

State is `File: [unavailable]` plus code; instruction is `Is this source code relevant to the software change: "{query}"?`; normal options are false/true. Actual paths are unavailable and explicitly marked. **Forward true probability is primary**; reversed and averaged scores are diagnostics. This checkpoint learned change-to-code relevance; general natural-language code search is not assumed to be the same task.

| Method | Forward AUC | Reverse AUC | Averaged-score AUC | Order decision disagreements |
|---|---:|---:|---:|---:|
| Original FP32, this run | 0.5008 | 0.5212 | 0.5167 | 211/553 |
| INT8, fixed experiment 17b | 0.5040 | 0.5277 | 0.5214 | 217/553 |
| Matched BM25 | 0.5378 | n/a | n/a | n/a |

AUC measures discrimination between positive and negative items; near 0.5 indicates little ranking discrimination. Neither model meets precommitted AUC≥0.60 and BM25+0.03. Do not claim statistical superiority from their small difference. This is repeatedly used validation evidence, not an independent final test.

## Isolation and verification

Go verifies the full source and partition, then exports validation requests only. Requests contain query/code, tokens, options and truncation metadata, **no labels**. Python reconstructs every sequence with upstream `build_sequence` and requires exact token/marker equality before any model inference: 1,106/1,106 matched.

Python writes local predictions. Go reconstructs requests from the original source, compares the entire prepared file, per-request identity/order, model/plan/partition metadata, then uses the existing evaluator. BM25 corpus statistics use development only. Reserves are read for full-source hashing and group partitioning, never model-scored, trained on or used for selection. No raw text or per-row scores go to Git/HF.

Replaying the Go evaluation reproduces all non-timing metrics. Full original-model inference was not repeated. Unit checks cover validation/scope isolation, option order, changed identities/truncation metadata and exhausted score sequences.

## Resources and reproduction

Apple M4 Pro, macOS ARM64, four CPU threads, 512 tokens. Python external wall 85.46 s, user CPU 163.06 s, peak RSS 3,009,675,264 bytes (~2.80 GiB). Internal 83.974 s includes model preparation and 1,106 calls. Check 15-minute/8-GiB limits between calls, not as hard within-call limits.

Median call 70.444 ms and p95 96.377 ms cover Python tensor preparation, FP32 execution and probabilities. INT8 Go timing includes tokenization and has a different boundary; do not use these for a precise speed comparison. `Report.ProbeSeconds` is **Go replay/evaluation of stored scores**, not inference time. No GPU/MPS performance claim.

```sh
go run ./cmd/riido-sourceprobe --stage prepare
python scripts/training/laya_original_validation.py --requests .cache/source19-requests.json --source-dir .cache/laya-code-source-18 --out .cache/source19-scores.json
go run ./cmd/riido-sourceprobe --stage evaluate
```

Follow [experiment 18](../laya-parity/README.en.md) for pinned original assets and the maintainer environment. Python remains a reference tool, not a user-runtime requirement. Use new paths to repeat preparation/inference; align Go `--requests`/`--scores` with Python `--requests`/`--out`. [Aggregate results](results-19.json) include hashes, versions and resource measurements.

Next focus on task-aligned claim learning. Distillation premised on useful current-head outputs remains unsupported. This binary relevance test does not decide whether latent representations, other training tasks or candidate ranking for actual code changes can help. Preserve the unused 2,403/2,401-question reserves until a candidate meets validation criteria. No default changes, new training or HF model release.
