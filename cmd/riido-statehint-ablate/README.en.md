# Development feature ablation

This separate maintainer command compares the **existing**
`statehintwide.Simple` and `statehintwide.Contextual` models, both with 2,048
feature bins. It uses the original 840 training families and the 120 development
validation families whose results were **previously exposed**. These results are development diagnostics,
not a fresh evaluation or evidence of deployment qualification. The V4 run,
v1 SDK, model assets and fitting driver are unchanged.

Simple uses word unigrams and character 2/3-grams with count normalization.
Contextual adds word pairs and character 4/5-grams, signed hashing and log-TF.
This is a comparison of that feature bundle, not an isolated causal test of
word order. Both modes start fresh with learning rate .02, 40 epochs, batch 32,
decay .001 and seed 1729. Temperature stays 1. There is no calibration,
adaptive arm, model selection, final-test input or promotion.

Use Go 1.27.1. Supply workspace-relative regular JSONL files and their exact
lowercase SHA-256 digests from the approved local inputs. Replace the digest
placeholders below. There is no download or whole-manifest input.

```sh
go run ./cmd/riido-statehint-ablate \
  --train .cache/statehint-development/train840.jsonl --train-sha256 TRAIN_SHA256 \
  --validation .cache/statehint-development/validation120.jsonl --validation-sha256 VALIDATION_SHA256 \
  --check

go run ./cmd/riido-statehint-ablate \
  --train .cache/statehint-development/train840.jsonl --train-sha256 TRAIN_SHA256 \
  --validation .cache/statehint-development/validation120.jsonl --validation-sha256 VALIDATION_SHA256 \
  --fit --out .cache/statehint-feature-ablation/run01
```

Exactly one of `--check` and `--fit` is required. Check verifies pins, bounded
corpus parsing, paired locales, 840/120 family quotas, 105/15 families per
intent and disjoint family/lineage/row/text identities. It creates no outputs
and makes no model calls. The caller remains responsible for approved original
synthetic provenance; checksums and structural checks are not semantic or
rights verification. Only `text` enters model features; expected intent is the
training target. Annotation fields are not features.

Fit requires a new single-name run directory under the ignored
`.cache/statehint-feature-ablation` anchor. The anchor and run are 0700;
files are 0600 and existing files/runs cannot be overwritten. Outputs retain
the exact two input byte streams, fixed recipe, both `.rsh` models, complete
eight-probability vectors in validation source order and both full family
reports. Saved models are reread, checksum-validated and compared with the
trained model on every validation prediction field. Partial outputs and a
failure report remain available if a run fails.

Reports reuse `statehintfamily.Evaluate`: confidence .9, margin .05, false
P/C/Q costs 1/10/3 and target-miss cost averaged over the paired KO/EN rows
(one missed target row adds .5 per family), coverage .2, support
5 families/2 declared lineages, completion precision .98 and severity cost
strictly below .375. Even an eligible exposed-validation report remains a
development result; the command always reports no deployment qualification.
All failures and undefined precision/support cases remain visible.

The v1 reference artifact is 32,960 bytes; each existing wide artifact is
65,728 bytes. The v1 value is a size reference, **not a loaded or retrained
control**. Fit and validation prediction durations describe this local run;
they do not measure native/GPU memory or establish a speed/quality improvement.
Comparing these development results with earlier 32 KiB model results is also
development-only. No Laya, LoRA or ternary-model training is performed.

Tests use owned temporary metadata fixtures, including actual tiny fits for
both modes, exact reload/report parity, checksum/budget failures, split reuse
and private exclusive output checks. They do not establish model quality.

## Observed development run

The fixed two-mode run on the original training data and previously exposed
development validation produced the following results. Gated proposals use the
unchanged P/C/Q display gate; completion support is out of 15 families per locale.

| Mode | Eight-intent correct rows | Correct gated proposals / proposals | Correct completion families KO / EN | Eligible |
|---|---:|---:|---:|---|
| Simple2048 | 176/240 (73.33%) | 35/36 | 3/15 · 1/15 | No |
| Contextual2048 | 179/240 (74.58%) | 40/40 | 3/15 · 0/15 | No |

Both trained models are 65,728 bytes and passed exact save/load prediction parity
with 2,120 updates each. The complete local two-mode execution took 1.46 seconds
with observed maximum process RSS 18,022,400 bytes; no GPU runtime was used.
These single-run resource observations are not portable performance guarantees.
Both modes remain unqualified, especially for English completion support.
This exposed-validation feature-bundle comparison is not a fresh evaluation,
does not establish a context-only improvement and does not select or promote a model.

[한국어](README.ko.md)
