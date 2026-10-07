# Aggregate frozen Laya reference scores in Go

This command reads an explicitly pinned, previously exposed validation corpus
and an existing body-free Laya score file. It **never runs a model, loads weights,
trains, calibrates or selects an arm**. The separate offline reference worker
captures features with MPS and scores both frozen final linear layers on CPU;
this Go command verifies the resulting score bindings and reports
descriptive outcomes for humans and agents.

The references are the existing pretrained Laya base and its published v0.1
final-linear parameter replacement. Both still require the original backbone
when producing new native scores. The 4,252-byte delta is not a standalone Go
model. Its legacy option definitions differ from the newer V4 scope rubric;
these are transfer diagnostics, not fresh evaluation or deployment evidence.

## Use

From the repository root with Go **1.27.1**, substitute the approved relative
paths and their exact SHA-256 values:

```sh
go run ./cmd/riido-statehint-laya-reference \
  --validation exposed-validation120.jsonl --validation-sha256 VALIDATION_SHA \
  --scores scores.json --scores-sha256 SCORES_SHA \
  --check
```

`--check` verifies inputs and arithmetic without output files. To save the
aggregate, replace it with
`--out .cache/statehint-laya-reference-report/NEW-RUN`. The run directory must
be new (0700); `report.json` is written once (0600). The command also emits JSON
to standard output. Errors do not print corpus text or supplied arguments.

The corpus must contain exactly 120 paired KO/EN families, 240 prose rows and
15 families per intent, using the frozen V4 rubric and `validation` partition.
It is capped at 16 MiB, with the existing parser's bounded text requirements.
Scores are capped at 2 MiB. No calibration, final-test, model or Fit flag exists.

## Score contract and checks

The confirmed producer schema is `riido-statehint-laya-reference-scores-v1` with
`status: reference_only`. It declares the validation hash, fixed eight-intent
order, base/delta origins and weight pins, historical temperatures, instruction
and feature-cache provenance hashes, and unknown training steps
(`training_steps_known: false`, `training_steps: null`). Each row contains only
ID, locale, exact UTF-8 text hash, both eight-logit/eight-probability vectors and
`truncated: false`. It contains no raw prose, expected labels or family metadata.

| Reference | Origin | Temperature | Weight SHA-256 |
| --- | --- | ---: | --- |
| Original base | `base_pretrained` | 1.0000158548355103 | `891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c` |
| Published v0.1 delta | `published_v1_delta` | 0.65 | `e7037a4c92460dd0c77facb643a26a2f3ce99240e6a59bd78c15ad91df314fcc` |

The original temperature is its historical eight-choice bucket; 0.65 is the
delta's historical fitted value. Neither is fitted on V4. The adapter verifies
finite vectors, eight values, probability range/sum and agreement with stable
softmax of logits at the declared temperature (tolerance 1e-6). Producer
probabilities are retained unchanged. Score row order may vary; sorted ID lookup
binds every row to corpus ID, locale and exact text hash. Duplicate, missing,
truncated, incompatible or over-budget inputs fail.

The instruction must match its fixed UTF-8 SHA-256
`255725ec00a4526fefedfad729ddd75acc9681181b81fc0f8f96b89b0a400fd3`.
Missing unknown-step fields are rejected; numeric training steps are unsupported.
The weight/provenance fields are producer declarations and hashes, not proof of
native execution or publisher authentication. Keep the worker's source, resource
and execution receipts separately. This adapter reads no feature cache or model.

## Reading the report

Arrays use frozen eight-intent order; display candidates use
**progress / completion_report / question**, and locales use **KO / EN**.
Raw exact ties choose the first class in frozen order, matching Laya/Torch;
their zero margin cannot pass the unchanged **confidence0.9 / margin0.05** gates.
The other five intents produce no candidate.

The report includes raw eight-way confusion/accuracy/NLL (log floor 1e-15),
gated counts, correct support and coverage per locale, and distinct wrong
candidate families and declared lineages. Zero proposals mean **undefined
precision**, represented by a false `precision_defined` flag, not success or
precision zero. Family and language rows are not independent product samples.

Reference training steps stay unknown. The command does not fabricate SDK
`Learned` steps or call the learned-family qualification evaluator. Eligibility
is not assessed; its false contract flag is not an empirical rejection verdict.
Selection weight, calibration/test/adapter model calls, promotion and deployment
qualification remain zero/false. A completion label is a claim, not verified task
completion. Original fake-logit tests cover row/locale/hash binding, numerical
checks, zero support, paired wrong-completion counting and private outputs; they
are not native inference or semantic performance evidence.

## Actual frozen reference run

One offline FP32 MPS worker completed 240 encoder forwards on the already
exposed 120-family / 240-row validation corpus. Both final linears used the same
captured features. The Go aggregate completed without any model calls.

| Measure | Original base | Published v0.1 delta |
| --- | ---: | ---: |
| Raw correct, all /240 | 89 | 88 |
| Raw correct, KO /120 | 16 | 16 |
| Raw correct, EN /120 | 73 | 72 |
| Correct gated / all proposals | 12/12 | 11/11 |
| EN correct gated progress / completion / question | 8 / 4 / 0 | 7 / 4 / 0 |
| KO gated proposals | 0 | 0 |
| Eight-intent NLL | 1.738268 | 1.616183 |

English completion support was four families out of 15 targets (26.7%), spanning
three declared lineages, with four correct out of four proposed completions.
There were no wrong gated candidate families. Korean produced no candidates,
so its candidate precisions are undefined. This small English completion signal
does not qualify a teacher: overall accuracy is poor, the legacy definitions
differ from V4, and the data were previously exposed. The NLL comparison includes
both different weights and historical temperatures; it does not isolate a head
improvement. Neither reference was selected, qualified, trained or promoted.

The worker elapsed time was 23.44 seconds, including model loading and checks;
the synchronized native-forward total was 13.23 seconds (mean 55.12 ms per row).
The separately recorded whole-process time was 24.24 seconds. Peak process RSS
was 2,887,172,096 bytes. Maximum sampled MPS driver/allocated memory was
2,166,046,720 / 1,688,613,632 bytes; system availability stayed at least 51%.
These MPS samples miss transient peaks and overlap RSS in unified memory, so
they are not added. The private feature cache was 7,864,320 bytes; maximum
original-native reconstruction error was 0.0000019073486328125. There was no
Fit, calibration, final-test access or new model download. Training steps remain
unknown and eligibility remains unassessed.
