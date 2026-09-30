# External evaluation01: 2,400 pairs do not establish improvement

[한국어](PAIR-01.ko.md) · [Frozen plan](pair-plan.json) · [Aggregate results](pair-results-01.json)

**The actual 2,400-pair evaluation does not substantiate practical improvement for the models that achieved24/24 on the synthetic probe. No model is promoted.** This is binary query/code relevance, not repository selection, candidate retrieval or LLM cost savings.

## Partition and procedure

Pin the audited CoSQA SHA256. Connect rows sharing a lowercased/whitespace-normalized query or whitespace-normalized exact code. Code normalization may merge whitespace differences inside literals, conservatively grouping distinct snippets. It does not detect every renamed/semantic clone. Repository metadata is absent, so repository isolation is unproven.

Sort groups by a fixed seeded canonical-key hash, independent of labels, length and model scores. Assign whole groups until each split minimum is met.

| Split | Pairs | Groups | Use this cycle |
|---|---:|---:|---|
| development |12,189|3,754|BM25 corpus statistics only; no weight training|
| validation |607|175|unused|
| calibration |604|176|binary score thresholds|
| final |2,400|724|evaluated|
| reserve1 |2,403|716|not trained/scored|
| reserve2 |2,401|722|not trained/scored|

Reserve content is read for grouping/hash checks but not feature learning, scoring or performance selection; it is not literally an unread file. Future weights may use development, selection validation, thresholds calibration. The now-viewed final is development evidence for future design decisions.

Freeze all eight archived synthetic models and source/plan/membership hashes before evaluation; recompute identity before scoring. Final labels comprise1,171 relevant and1,229 irrelevant pairs. Weights were not trained on CoSQA, but thresholds use604 labels, so this is not wholly uncalibrated zero-shot classification.

## Results

Balanced accuracy averages positive and negative class recall. Keep all2,400 rows; the151 out-of-scope model inputs (6.29%) use BM25 scores and its calibration threshold.

| Method | Balanced accuracy |95% interval of difference vs BM25|
|---|---:|---:|
|BM25|52.62%|reference|
|Token overlap|53.22%|−1.19 to +2.43pp|
|FP32 seed1729|54.15%|−0.31 to +3.31pp|
|INT8 seed1729|54.31%|−0.17 to +3.49pp|
|FP32 seed2718|54.19%|−0.31 to +3.34pp|
|INT8 seed2718|54.09%|−0.39 to +3.22pp|
|Ternary PTQ seed1729|49.81%|−5.45 to −0.28pp|
|Ternary STE seed1729|49.61%|−5.39 to −0.84pp|
|Ternary PTQ seed2718|50.05%|−5.19 to +0.002pp|
|Ternary STE seed2718|51.36%|−3.87 to +1.38pp|

500 paired connected-group bootstrap replicates produce percentile intervals. Do not treat the2,400 rows as independent. Grouping addresses within-group dependence, not residual semantic similarity between groups, source bias or multiplicity across eight inspected models. No model has a positive lower bound; some ternary intervals are entirely below BM25.

JSON includes confusion matrices, accuracy/precision/recall/AUC and in/out-of-scope subgroups. A model's `EligibleAUC` uses2,249 cases and is not directly comparable with BM25's all2,400 AUC; compare `InScopeMetrics` for matched support. Raw scores are not probabilities.

## Resources and reproduction

Frozen manifest SHA256: `19001225d9af3dc59ab474ff37313fbe82673d4a5b5c425f43941c63095998a8`. Replay produced byte-identical case traces. Traces remain local to avoid redistributing source labels; published results contain hashes and aggregates.

Apple M4 Pro, Go1.27.1, CPU: first whole-process wall0.78s, userCPU0.43s, maximum RSS61,521,920B (58.67MiB), no swaps. This includes source loading, partition verification, eight models, two controls and bootstrap; not single-inference latency or total agent memory. JSON `Seconds` starts after loading and differs from OS wall time.

```sh
mkdir -p .cache/pair-eval
# Prepare the pinned source and local synthetic models as documented previously.
go run ./cmd/riido-paireval --input .cache/semantic-scale/cosqa-all.json \
  --selection .cache/semantic-learning/local-train/selection.json \
  --out .cache/pair-eval/frozen
go run ./cmd/riido-paireval --stage evaluate \
  --input .cache/semantic-scale/cosqa-all.json \
  --selection .cache/semantic-learning/local-train/selection.json \
  --frozen .cache/pair-eval/frozen/frozen.json --out .cache/pair-eval/evaluated
```

Directories must be new. Training duration fields may change selection/frozen hashes on replay; verify coefficient hashes and identical-input case traces separately. Exit0 means evaluation completed, not quality passed. Every result remains `Promotion=false`.

## Next PDCA

The narrow synthetic training distribution does not establish that all small models are incapable. Train relevance on development12,189, select using validation607, calibrate, then evaluate reserve1's2,403 only after choices are fixed. Do not tune against the now-viewed final and claim a fresh improvement on it. Actual candidate retrieval and agent outcomes still need separate evaluation. Do not recommend or auto-apply these failed models.
