# PDCA: inspect training references before another Fit

This stage inspects what the existing public training corpus contains before training another model. It pairs the [product mark review](../state-hints-claims/PRODUCT-MARKING.en.md) with a training-only audit, keeping comment claims separate from actual task state.

## Executed audit

The Go 1.27.1 command audited 1,680 original rows, 840 bilingual families and 280 related groups. Each locale has 840 rows. False denotes clear absence of a positive claim, not unfinished work.

| Head | Korean true / false / unknown | English true / false / unknown |
| --- | --- | --- |
| Response requested | 206 / 544 / 90 | 205 / 545 / 90 |
| Activity reported | 284 / 499 / 57 | 284 / 499 / 57 |
| Completion reported | 318 / 456 / 66 | 316 / 458 / 66 |

Unknown comprises about 6.8% of activity and 7.9% of completion training references. The actual operational distribution remains unknown. No row has at most 16 Unicode code points; 17–32 counts are 24 Korean and 0 English. Bytes and code points are not semantic complexity or model tokens. These observations do not establish why a model failed. Exact Unicode-lowercase/whitespace-normalized duplicates and conflicting targets are zero; that does not establish semantic uniqueness or group independence. See [aggregate results](RESULTS.development.json).

## Separate reference consistency review

A frozen hash order of group IDs selected 40 whole groups / 240 rows without labels, model scores or errors. A new AI role judged each locale with original targets hidden and sealed every first judgment before comparison. Agreement was 716/720 head decisions (99.44%); exact three-head agreement was 236/240 rows (98.33%). Four decisions in 2 bilingual families / 2 groups disagree. Both vote sets remain unchanged; no automatic relabeling occurred.

This measures **AI reference consistency**, not model accuracy, human gold or usability. Correlated AI judgments and repeated synthetic scaffolds remain. It is not a semantic review of all 1,680 rows. Original evidence/reasons were unavailable in the pinned comparison export, so only original targets were compared; [the aggregate record](REFERENCE-CONSISTENCY.development.json) retains that limitation. The result does not justify widespread relabeling.

## Reproduce and proceed

The [audit command](../../cmd/riido-statehint-claims-training-audit) decodes only one fixed-SHA public training corpus and emits aggregates. It does not use models, evaluation material or real-work text.

```sh
bash scripts/verify-claims-training-audit.sh
```

The script downloads our own public training asset at an immutable Hugging Face revision, runs the Go audit and compares its output with the frozen aggregate. It performs no Fit, model inference, registration or application writes. One whole-process audit observation was 0.61s wall/0.04s user CPU/13,205,504B peak RSS, including reading, hashing, JSON decoding, aggregation and reporting. Raw profiles are not published.

Next work verifies coverage of short/ambiguous comments, useful human marking criteria and actual context separately. Freeze a new hypothesis, data and fresh evaluation before another Fit. Preserve existing exposed-evaluation non-reuse, calibration/final-test seals and qualification rules. This stage consumes 0 new Fits; cumulative research remains 23, with no qualified model. CI reproduces audit-code behavior, not product usefulness.

Public training asset: `JooYoon/riidolaya-development-claim-hints-mlp16-v0.1`, revision `31705b576851edba30f10c0e26fd22793ffc6f56`, SHA `c5b5a0adbc721d73213d4565c6a7c430094f31b8e7502fb5249e7ed40a3da837`. Preserve original data, weights and earlier failures.
