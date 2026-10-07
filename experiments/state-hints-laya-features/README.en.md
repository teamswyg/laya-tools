# Frozen training-feature extraction

This offline maintainer tool extracts the original pinned train840 /1,680 KO/EN
rows for a separate true-label supervised Go probe. It performs no Fit, loss,
teacher pseudo-labeling, calibration, final-test/private-task reads or model
selection. The product runtime remains Go; Python is a maintainer exception.
The old frozen-reference experiment and all original corpus bytes stay unchanged.

The base/checkpoint, four config/tokenizer pins and original V1 QUESTION/order
are identical to the frozen reference. Its definitions still differ from the
V4 scope rubric. The separate sidecar copies original expected_intent only as
metadata; the tokenizer and backbone receive text-derived IDs plus the fixed
QUESTION, never target labels, units or split assignments.

```sh
python experiments/state-hints-laya-features/extract_training.py \
  --base BASE_DIR --train TRAIN840.jsonl \
  --plan PROBE_PLAN.json --plan-sha256 PLAN_SHA \
  --reference-plan FROZEN_REFERENCE_PLAN.json --check
```

`--check` verifies all byte pins and paired840-family/class105 identities, then
runs local tokenizer preflight over **all1,680rows before creating a backbone**.
It compares frozen512-token/192-head preprocessing with an uncut tokenizer-only
sequence. Any input or option truncation fails the stage with zero backbone
forwards; no row is skipped, shortened, relabeled or retried with a larger model
budget. Tokenizer imports can load library definitions but do not instantiate
or execute a model. Check mode writes no features or output files.

Only after that succeeds may ROOT authorize
`--run --out .cache/statehint-laya-features/NEW-RUN`. One frozen MPS/FP32 worker
makes exactly1,680forwards, capturing8×1,024 pre-final option features per row.
Prepared token IDs are reused, and one32KiB feature row is written at a time.
No delta scorer, predicted teacher label, full backbone copy or55MiB feature
buffer is made. MPS is mandatory; offline loading/fallback0,5GiB peak process
RSS,25% MPS fraction and20% system availability are fixed. Existing foundation
and extraction caches plus64MiB reserved output must fit the512MiB task-cache
cap before preflight. Other cache files are statted only, not opened.

`features.f32le` is exactly55,050,240bytes. `sidecar.json` uses
`riido-statehint-laya-feature-sidecar-v1`, base/corpus/instruction/feature hashes,
frozen intent order, rows_count1680, shape[1680,8,1024], dtype
`float32_little_endian`, and ordered rows containing only id,locale,text_sha256,
expected_intent. Go verifies those labels/IDs against the pinned corpus and
applies the unchanged663fit/177dev group split. Validation features reuse the
previously captured file separately; this tool never opens validation.

The new directory is0700 and files0600; total output is capped at64MiB. Numeric
or extraction failure retains partial private features and `FAILED.json`, with
no success sidecar or head Fit. Preflight failure exits before an output/cache
exists. Resource samples are not transient GPU peaks, and unified RSS/MPS
memory overlaps. Features and row metadata must never be uploaded or committed.
The separate Go head requires the frozen842MB backbone for novel text and is
not a standalone low-memory semantic model.

```sh
PYTHONDONTWRITEBYTECODE=1 python -m unittest discover \
  -s experiments/state-hints-laya-features -p 'test_*.py'
```

Owned tokenizer/forward stubs test exact streaming bytes/order, label isolation,
all-row preflight rejection, finite captures, source pairs/pins and cache/output
budgets. They make no real model calls and provide no pretrained accuracy or
MPS proof. ROOT performs actual tokenizer preflight and extraction only after
source review and resource checks. Apache-2.0 source/config provenance is preserved.

## Actual extraction and supervised probe run01

ROOT completed all 1,680 frozen MPS forwards after tokenizer preflight passed
121–312 tokens with no input/option truncation. The cache is 55,050,240 bytes and
producer output 56,300,062 bytes, below 64 MiB. Original expected labels stayed
metadata; no teacher predictions or MPS weight updates were used. The MPS
process exited before the separate pure Go head Fit.

| Phase | Whole-process elapsed | Peak process RSS | Scope |
| --- | ---: | ---: | --- |
| Training feature extraction |128.79s|2,976,038,912bytes (2.77GiB)|One MPS/FP32 backbone,1,680rows |
| Go shared-head run |1.72s|24,182,784bytes (23.06MiB)|Previously cached features; zero backbone calls |

Extraction worker time was 125.865 s, including loading/guards/extraction; native
full-base forwards totaled 69.980 s, mean 41.65 ms. External extraction CPU user/system
were 19.17 s/51.80 s. Sampled MPS driver/allocated maxima were 2,174,435,328 /
1,688,613,632 bytes and minimum system availability 52%. Samples miss transient
peaks and overlap RSS in unified memory; do not sum them. These are one local
run's distinct phase measurements, not a matched speedup or novel-text latency.

The Go head used 663 fit families/1,326 rows, original AI-reviewed targets,
fresh zero weights, 40 epochs/batch32/rate.001/decay.01/seed1729/T1 and 1,680 updates.
Its separate 4,320-byte artifact had exact trained Prediction and byte reload
parity. The already exposed diagnostic outcomes remained unsuccessful:

| Diagnostic | Raw eight-class correct | Gated P/C/Q proposals | Eligibility |
| --- | ---: | ---: | --- |
| Internal dev177families |118/354 (33.33%)|0|Fails coverage/support and undefined completion precision |
| Exposed validation120families |92/240 (38.33%)|0|Same failures; also equals the all-abstain cost comparator |

All per-locale P/C/Q precisions are undefined at zero proposals; numeric zero
sentinels are not 0% or 100% precision. Selection, promotion, calibration and
final-test access stayed zero. No useful completion gain or low-memory
novel-text qualification was established. The small head still requires the
842,609,210-byte frozen backbone for new text. Features and row metadata remain
private and are never HF/Git assets; existing studies/code/gates remain unchanged.
