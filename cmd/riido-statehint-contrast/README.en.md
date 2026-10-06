# Controlled completion contrast

Maintainer-only Go command comparing two fresh `statehintwide.Contextual2048`
arms. It uses the frozen original train840 split: 663 fit families and 177
internal-development families. Both arms have 1,454 rows and 1,840 updates:

- `class_matched_control`: the 663 fit families plus 64 extra **whole KO/EN
  family draws** from that same fit set: 32 completion, 8 reference, 8 progress,
  8 planned and 8 blocker. Ranking is SHA-256 of
  `completion-contrast-class-matched-control-1729:` plus family ID, with family
  ID breaking a digest tie. Each class cycles only when its candidate list is
  too short. Repeated draws are sampling weights, not new families.
- `matched_semantic_contrast`: the same 663 fit families plus 64 source-derived
  augmentation families with exactly the same class and locale counts.

Both start with fresh weights and optimizer. The recipe is 40 epochs, batch 32,
learning rate 0.02, AdamW decay 0.001, seed 1729 and temperature 1. No parent model or
calibration input exists. Annotation fields are never model features; only
text and fit-side expected intents form training samples.

```sh
go run ./cmd/riido-statehint-contrast \
  --train train840.jsonl --train-sha256 TRAIN_SHA256 \
  --augmentation augmentation64.jsonl --augmentation-sha256 AUGMENTATION_SHA256 \
  --split split.json --split-sha256 SPLIT_SHA256 \
  --source-overlay source-groups.json --source-overlay-sha256 AUDIT_SHA256 \
  --check
```

The original train and split digests must match the frozen study pins embedded
in the command. Augmentation and audit require explicit lowercase SHA-256
pins. `--check` validates all inputs before making any output or model call.
`--fit --out .cache/statehint-completion-contrast/NEW-RUN` executes the fixed
comparison after its actual inputs have been authorized. Exactly one mode is
required. There are no default inputs, recipe overrides, calibration, final-test,
whole-manifest or parent-model flags.

An optional `--validation validation120.jsonl --validation-sha256 SHA` is
previously exposed development data. Both arms receive the same descriptive
diagnostic. Its selection weight and promotion count are always zero.

The required source audit has schema
`statehint-completion-contrast-source-groups-v1`, top-level
`augmentation_sha256`, `train840_sha256`, `frozen_internal_split_sha256` and 64
`entries`. Each entry binds `family_id`, `pair_id`, `expected_intent`,
`effective_leakage_group_id`, `source_connected_train_family_ids`,
`source_connected_train_group_ids`, `all_selected_source_component_members`
and `row_ids`/`text_sha256` arrays in KO/EN order. Additional provenance fields
are retained in the exact audit copy. Every listed source group must include
all its original selected members, all must be frozen fit-side, and connected
augmentation families must share a group. Every matched pair has one completion
family and one reference/progress/planned/blocker family. The command never
rewrites the original840 groups or moves an internal-dev group.

Selection uses only the 177 internal-dev families and the unchanged
`statehintfamily.Evaluate` gate and costs: confidence 0.9, margin 0.05, correct
coverage 0.2, support 5 families/2 declared lineages per P/C/Q and locale,
completion precision 0.98 and severity cost below 0.375. Among eligible arms,
lower severity cost wins, then lower eight-intent NLL, then fixed arm order.
If neither is eligible, no arm is selected and the previous parent remains
unchanged. The command never promotes a model or claims deployment qualification.

Each new run directory is 0700 and every output is 0600. Existing runs/files are
not overwritten. Outputs include exact pinned source/split/audit byte streams,
the fixed recipe, the control draw ledger, both 65,728-byte models, complete
eight-probability vectors, internal-dev and optional diagnostic reports, and
source pins. Saved models are reread and checked for byte serialization parity
and exact prediction parity on every evaluated row. Malformed inputs create no
output; a fitting/output failure records `FAILED.json` and retains clearly
marked partial artifacts rather than a completed report.

Fit/prediction time and post-fit Go heap allocation are reported separately.
Go heap is **not OS RSS**: execution should also measure process peak RSS with
the host's process-resource tool. This command has no native/GPU runtime and
makes no speed or accuracy claim before an actual controlled run.

The augmentation is correlated AI-authored, source-derived TRAINONLY
supervision with prior heldout-author exposure disclosed. Checksums and
structural checks do not verify semantic labels or establish human truth,
independent incidents, a pair-mechanism causal effect or fresh evaluation.
Tests use owned artificial fixtures and tiny real fits solely for isolation,
sampling, IO and serialization correctness.

## Observed controlled development run

The fixed run used the 663/177 internal split and the accepted 64-family,
128-row TRAINONLY augmentation. Both arms trained on 1,454 rows for 1,840
updates and passed exact artifact reload parity.

| Arm | Internal-dev eight-intent correct | Correct gated / proposed | Completion precision | Severity cost | Eligible |
|---|---:|---:|---:|---:|---|
| Class-matched control | 317/354 | 73/74 | 1.00 | 0.138418 | Yes; selected |
| Semantic contrast | 317/354 | 72/73 | 0.96 | 0.192090 | No; completion precision failed |

Control was selected by the frozen internal-dev rule. Contrast's completion
precision was below the unchanged 0.98 requirement. The exposed validation
diagnostic cannot override this selection:

| Arm | Diagnostic eight-intent correct | Correct gated / proposed | Correct completion families KO / EN | Diagnostic eligible |
|---|---:|---:|---:|---|
| Class-matched control | 172/240 | 39/39 | 4/15 · 0/15 | No |
| Semantic contrast | 183/240 | 48/48 | 7/15 · 2/15 | No |

Both remained unqualified on that diagnostic, including insufficient English
completion support. Its higher contrast counts are descriptive and do not
establish a causal or fresh-evaluation gain. No calibration or final-test
forward occurred and no model was promoted. Each artifact was 65,728 bytes;
the host observed 1.86 seconds total elapsed time and 26,165,248 bytes maximum
process RSS for this single run. These observations are not portable resource
guarantees. Source-derived supervision and correlated AI labels remain the
limitations described above.

[한국어](README.ko.md)
