# State-hint classification training protocol

This experiment identifies a message's single intent so a Go policy can suggest
labels, emoji, and allowed state changes. The intent order is `question`,
`blocker`, `reference`, `progress`, `completion_report`, `cancel_request`,
`planned`, `unclear`. A completion report records what the message claims;
independent events and product policy determine whether completion is allowed.

## Original fixture contract

`templates.json` contains 96 original language-specific template families under
Apache-2.0: eight intents, two locales (`ko`, `en`), six templates per combination.
Each family expands to 25 variants by replacing only `{project}` and `{item}`.
Names and IDs are visibly fictional. No real prompts, private source, user
messages, external annotation data, or paid model calls are used.

The canonical `corpus.jsonl` row has `id`, `group_id`, `locale`, `intent`, `split`,
`text`, and `origin`. `origin` is `original_authored_synthetic_development`.
`id` is the family ID followed by `-00` through `-24`; `group_id` is the family
ID (at most 12 characters). For Korean, replacements are
`가상프로젝트-00`/`가상항목-00`; for English they are
`FictionalProject-00`/`FictionalItem-00`, with the matching index.

Within every locale and intent, three families belong to training, one to
validation, one to calibration, and one to test. Therefore the 2,400 rows are:

| Partition | Language-specific families | Rows | Rows per locale |
| --- | ---: | ---: | ---: |
| Training | 48 | 1,200 | 600 |
| Validation | 16 | 400 | 200 |
| Calibration | 16 | 400 | 200 |
| Test | 16 | 400 | 200 |

Group separation prevents variants of a template entering multiple partitions.
Text duplicates and conflicting family metadata are rejected. Similar concepts
still occur across partitions. Korean and English templates are paired
paraphrases with aligned partitions, not 96 independent real-world events.
The 25 fictional name substitutions provide very little linguistic variation.
These are development fixtures, not 2,400 independently annotated product cases.
`unclear` includes negated completion, quoted requests, hypothetical completion,
contradictory claims, missing context, and ambiguous current situations.

## Pretrained reference fine-tuning

`scripts/training/train_statehint.py` uses the existing pinned Laya 0.3.21 /
PyTorch 2.14.0 maintainer environment. It loads the original English
`convaiinnovations/laya` checkpoint at revision
`55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`, requiring weight SHA-256
`891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c`.
Configuration and tokenizer files must also match the pre-recorded plan hashes.
Hub/network loading is disabled. Python is a maintainer tool; the product
runtime remains Go.

Only `scorer.3.weight` and `scorer.3.bias` can train: 1,025 parameters in the
original pretrained final scoring linear layer. The encoder, decision head,
type embedding, preceding scorer layers, and action head remain frozen in
evaluation mode. One MPS forward per row captures the eight candidate vectors
entering the final linear layer. Reusing these frozen vectors for epochs is
mathematically the same final-linear objective as repeatedly running the
unchanged preceding model, subject to the recorded numerical reconstruction
check. It does not replace the backbone needed for Laya inference.

This is supervised cross-entropy fine-tuning with AdamW. Both learning-rate
trials start from the same pretrained final linear, rather than from random
weights. Every trial uses 20 epochs, batch size 32, weight decay 0.01, gradient
norm limit 1.0, and fixed seed 1729. Learning rates are fixed before execution at
0.001 and 0.003. Validation NLL selects the weights and epoch. Calibration rows
alone fit a scalar temperature over 0.5 through 5.0 in steps of 0.05.

The 0.9 confidence threshold is fixed. The selected parameter file and
temperature are saved in `LOCK.json` before any final-test forward. Test rows
then undergo one frozen feature extraction and final evaluation. Data integrity
checks read partition metadata before training; final-test metrics never enter
gradient updates, selection, or calibration. Test observations turn future
revisions of these fixtures into development evidence.

`results.json` reports raw and calibrated baseline/tuned metrics for validation,
calibration, and test, both pooled and separately for Korean and English:
confusion matrices, accuracy, NLL, Brier score, ten-bin ECE, accepted precision,
coverage, and acceptance excluding `unclear`. Raw tuned metrics help distinguish
weight changes from temperature fitting. The upstream baseline's eight-choice
temperature bucket is retained as a separate comparison. A high-confidence
`unclear` result is still an abstention from a concrete state-changing intent.

The English checkpoint's Korean capability is unqualified. Korean results must
stay visible and cannot inherit an English success claim. A multilingual
checkpoint is a separate later experiment with new provenance and language
evaluation. The existing pilot showed actual head updates without establishing
an accuracy improvement; this experiment must also retain unfavorable results.

## Fixed execution limits and outputs

The plan schema is `riido-statehint-training-plan-v1`. Required fields are
`source_sha256`, `reference_sha256` (the reused `train_pilot.py` helpers),
`data_sha256`, `templates_sha256`, `base_sha256`, `base_revision`,
`learning_rates`, `epochs`, `seed`, `batch_size`, `confidence_threshold`,
`extraction_minutes`, `training_minutes`, `overall_minutes`, and `base_files`.
`base_files` pins `rl_agent_config.json`, `encoder/config.json`,
`tokenizer/tokenizer.json`, and `tokenizer/tokenizer_config.json`. The plan is
written before execution and hash-checked by the trainer.

Feature extraction is capped at ten minutes, cached-feature training at ten
minutes, and the complete run at twenty minutes. Resource checks enforce a
5 GiB process RSS ceiling, 20% system memory availability floor, 30 GiB free
disk floor, and 25% MPS allocator fraction. MPS is required, FP32 is used, and
CPU fallback is disabled. No failed or out-of-memory run is automatically
retried. The recorded per-row cost separates full feature-extraction wall time
from synchronized model-forward time. Step-end GPU allocation samples miss
transient peaks; RSS and MPS memory overlap on Apple Silicon and are not added.

The output directory must be new and local. The parameter delta is
`scorer.safetensors` (approximately 4 KiB, hard cap 8 KiB). The trainer also
writes `rl_agent_config.delta.json`, `LOCK.json`, and `results.json`. It writes
no full backbone copy, feature cache, or optimizer checkpoint and publishes
nothing. The configuration delta explicitly replaces the `choice:6-10`
temperature so a stale inherited bucket cannot hide the fitted temperature.
Combining this delta with a base is a maintainer operation, not a standalone
model load.

Example invocation from the repository root after the plan is finalized:

```sh
.cache/mps-training-venv/bin/python scripts/training/train_statehint.py \
  --base .cache/training/base \
  --plan experiments/state-hints/plan-v1.json \
  --out .cache/training/state-hints-new-version
```

The cheap Go classifier is an independent supervised classifier over hashed
text features. Its own warm-start/retraining path must retain the same data
splits, taxonomy, feature schema, and calibration discipline. Training that
classifier is not pretrained Laya fine-tuning. Multilabel application labels
need a separate sigmoid/BCE task after the single-intent baseline is measured.

## Method references

- [PyTorch transfer learning](https://docs.pytorch.org/tutorials/beginner/transfer_learning_tutorial.html): freeze preceding pretrained layers and train a final classifier.
- [PyTorch cross-entropy](https://docs.pytorch.org/docs/2.14/generated/torch.nn.CrossEntropyLoss.html): supervised multiclass logit objective.
- [Laya's official Apple Silicon fine-tuning script](https://github.com/NandhaKishorM/laya/blob/main/notebooks/laya_finetune_typed_decisions_mps.py): upstream training reference; its broader RLCD/encoder updates differ from this bounded linear-only experiment.
- [Guo et al., On Calibration of Modern Neural Networks](https://arxiv.org/abs/1706.04599): temperature scaling for probability calibration.
- [GroupKFold](https://scikit-learn.org/stable/modules/generated/sklearn.model_selection.GroupKFold.html): separate related examples by groups rather than individual rows.
