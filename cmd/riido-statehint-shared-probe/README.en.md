# One supervised Go head fit on frozen features

This maintainer study asks whether a shared linear head can learn V4 scope from
the original AI-authored/reviewed labels over frozen Laya representations. It
does not use teacher predictions as labels, run a backbone, or perform a sweep.
It needs complete, independently pinned feature extraction before `--fit`.

Train840 supplies 1,680 KO/EN rows; the frozen group split gives 663 fit families
(1,326 rows) and177 already exposed internal development families (354 rows).
The fit is one fresh-zero shared1,024-weight head: 40 epochs, batch32, lr0.001,
AdamW decay0.01, seed1729, temperature1 and1,680 updates. Shared bias is zero.
The existing 240-row exposed validation feature cache is reused, not re-extracted.

## Inputs and use

Use Go1.27.1 from the repository root. Every input is an explicit local relative
FILE/SHA pair: `train`, `split`, `features`, `sidecar`, `validation`,
`validation-features`, `validation-scores`, and `reference-report`.

```sh
go run ./cmd/riido-statehint-shared-probe \
  --train train840.jsonl --train-sha256 TRAIN_SHA \
  --split frozen-split.json --split-sha256 SPLIT_SHA \
  --features features.f32le --features-sha256 FEATURE_SHA \
  --sidecar sidecar.json --sidecar-sha256 SIDECAR_SHA \
  --validation exposed-validation120.jsonl --validation-sha256 VALIDATION_SHA \
  --validation-features validation-features.f32le --validation-features-sha256 VALIDATION_FEATURE_SHA \
  --validation-scores scores.json --validation-scores-sha256 SCORES_SHA \
  --reference-report reference-report.json --reference-report-sha256 REFERENCE_SHA \
  --check
```

`--check` verifies full source/split/feature hashes and row mappings without head
training, native calls or output files. Replace it with
`--fit --out .cache/statehint-shared-probe/NEW-RUN` for the single fixed fit.
Train/split/validation/validation-feature hashes are frozen constants; the new
training feature and sidecar hashes must come from the completed approved worker.
There are no recipe, calibration, final-test, new extraction or model-download flags.

The sidecar schema is `riido-statehint-laya-feature-sidecar-v1`. It binds the
original base/instruction, eight-intent order, corpus/feature hashes, exact
`[1680,8,1024]` shape, little-endian float32 dtype and feature-row order. Row
metadata has ID, locale, text SHA and original `expected_intent`. Go checks the
label equals the corpus annotation, while the backbone never receives it.
The existing body-free validation score rows provide that cache's row mapping.

Full feature hashes are streamed; fitting reads only one32KiB row at a time.
The 55,050,240-byte training file and7,864,320-byte validation file are not copied
into a second full Go array. Feature extraction is external and its MPS worker
must finish/release the backbone before this pure-Go head fit.

## Outputs and interpretation

A fresh0700 run directory contains exclusive0600 recipe, source-index/label
receipt, actual head, all-eight prediction records and aggregate reports.
Raw training/validation bodies and large feature files are not duplicated.
The4,320-byte RSP head uses `pkg/statehintsharedprobe.Load`; use `ScoreFeatures`
with the matching frozen vectors. It is incompatible with text `.rsh`/`.rsm`
loaders and cannot predict novel text without the large backbone.

Original/delta historical validation aggregates are retained as reference
comparisons on the same cached validation features. Their unknown backbone/head
history is not represented as new Go steps. Only the new fitted head's1,680
updates are known. The existing family evaluator may calculate a descriptive
eligibility flag; that is not deployment qualification. Both evaluation sets
are already exposed, selection weight is0, and no model is selected/promoted.
Confidence0.9/margin0.05 stay unchanged. AI labels are not human gold and a
completion report does not authorize a task transition.

Tests use original temporary features/corpora, verify labels and row order,
fit/dev isolation, exact shared gradients, numerical/file bounds and trained
artifact parity. The actual first-run failure is recorded below.

## Actual first run: failed display-support target

Complete frozen extraction and the single Go fit finished. The head used1,326
original fit rows,40 epochs and1,680 updates; mean training loss was1.799871.
The saved4,320-byte RSP artifact has SHA-256
`bcfe817a17bd6fa1a6fb2cb46272911129140a1a74facc2debd155cc36a3ba5b`.
All evaluated predictions and reserialized bytes matched after reload.

| Already exposed evaluation | Raw correct | Eight-intent NLL | Gated proposals | Severity cost |
| --- | ---: | ---: | ---: | ---: |
| Internal development,177 families | 118/354 | 1.760635 | 0 | 0.338983 |
| Validation diagnostic,120 families | 92/240 | 1.626921 | 0 | 0.375000 |

Both KO and EN produced zero progress/completion/question candidates at the
unchanged0.9/0.05 gates. Precision is undefined and correct coverage is zero;
zero wrong candidates are the result of complete abstention, not useful safety
or successful recognition. Both descriptive eligibility checks failed, including
the validation all-abstain comparator. This experiment did not achieve the
display-support goal. The original base89/240 and published delta88/240 remain
historical references with different temperatures; no teacher ranking or
positive product claim follows from the new92/240 raw result.

The Go fit itself took1.58 seconds; the externally measured head-only process
took1.72 seconds with peakRSS24,182,784 bytes. Prior extraction required1,680 MPS
forwards and55,050,240 feature bytes: worker125.87 seconds, native-forward total
69.98 seconds, externally measured whole process128.79 seconds and peakRSS
2,976,038,912 bytes. The small-head process does not include the large backbone
needed for new text. No further backbone calls, selection, calibration, final
test, promotion or parameter/temperature sweep occurred. The failed model and
records are preserved for research; deployment remains unqualified.
