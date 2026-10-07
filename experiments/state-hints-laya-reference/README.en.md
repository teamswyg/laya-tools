# Frozen Laya reference on exposed validation

This is an offline maintainer experiment, not a Python product dependency or a
new training run. It compares the cached English Laya base at revision
`55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851` with the already published v0.1 final
linear replacement. The 842,609,210-byte backbone and 4,252-byte delta remain
unchanged. No downloads, weight updates, distillation, calibration, selection,
promotion or application mutations are performed.

The eight choices and original V1 instruction are preserved exactly. Their
semantics differ from the newer V4 current-unit/scope rubric, notably the old
criterion for quoted/hypothetical unclear cases. This is a cross-definition
reference, not a qualified teacher or a fair isolated model-capacity comparison.
The English checkpoint's Korean capability remains unqualified.

Only the already exposed validation120 / 240 KO/EN rows are accepted, with a
fixed SHA and paired identity/class checks. No train, calibration, final-test,
whole-manifest or private-task input exists. The worker receives text only;
expected labels and scope metadata cannot affect its forwards. Predictions and
features stay private. No actual reference model call has been made by the
owned stub tests.

```sh
# Use the existing pinned Laya0.3.21 / Torch2.14.0 maintainer environment.
python experiments/state-hints-laya-reference/evaluate_reference.py \
  --base BASE_DIR --delta DELTA.safetensors \
  --validation EXPOSED_VALIDATION.jsonl \
  --plan PLAN.json --plan-sha256 PLAN_SHA --check
```

`--check` verifies exact base/delta/config/tokenizer/data/plan pins and bounded
inputs without importing Torch/Laya, calling models or writing output. For a
separately authorized reference execution, replace it with
`--run --out .cache/statehint-laya-reference/NEW-RUN`. MPS is mandatory; CPU
fallback and network loading are disabled. The command does not import or
invoke the old trainer, which can access other partitions and fit parameters.

One worker captures eight 1,024-dimensional FP32 vectors entering the original
last scoring linear from exactly one frozen forward per row, at most 240
encoder calls. The base and delta scores reuse these same vectors. Native base
logits must reconstruct within .001 absolute error. Serialized scores are CPU final-linear computations; the native full-base
MPS logits serve only as a parity reference. The original weights are never
replaced to obtain delta scores. Both heads score on CPU from the captured
vectors; no second backbone or encoder retry is required.

The base uses its pinned historical `choice:6-10` temperature
1.0000158548355103; the delta uses its previously locked 0.65. Both are recorded
before forwards. There is no raw-temperature diagnostic, sweep or new
calibration. Body-free `scores.json` records all eight logits/probabilities,
rowID/locale/textSHA and explicit pretrained/delta origins. Reference training
steps are unknown/null, not invented Go optimizer history. The companion Go
reference command computes descriptive family/locale P/C/Q gates at .9/.05;
it does not present these origins as a qualified Go state classifier.

The fresh output directory has mode 0700 and every file 0600. `features.f32le`
is exactly 7,864,320 bytes, in input-row / frozen-option-order / feature order.
Metadata includes its hash and schema; all output is capped at 16 MiB, with no
full backbone copy, new weights or raw prompt export. Resource checks require
peak process RSS at most 5 GiB, system availability at least 20%, and the MPS
allocator fraction .25; driver allocation is also sampled against that fraction.
The private receipt distinguishes native forward timing from full reference
elapsed time. MPS samples miss transient peaks, and RSS/MPS overlap in unified
memory, so they are not added. Failure retains explicit private `FAILED.json`
and partial features with no successful qualification; there is no automatic retry.

Run the owned arithmetic stub tests without model dependencies:

```sh
PYTHONDONTWRITEBYTECODE=1 python -m unittest discover \
  -s experiments/state-hints-laya-reference -p 'test_*.py'
```

These tests verify once-only extraction/two scoring heads, full feature bytes,
label/annotation isolation, finite scores, reconstruction/truncation failures,
fixed-plan negative cases, exact bounded pins and private exclusive outputs.
They are pipeline tests and provide no evidence of actual pretrained accuracy
or MPS execution. Actual reference execution is separate from these stub tests; run01 is
documented below after ROOT's resource preflight and execution. Model/code licenses remain Apache-2.0; the Go product
runtime and existing study artifacts remain unchanged.

## Actual frozen reference run01

ROOT completed one Torch MPS/FP32 worker on the previously exposed 120-family,
240-row, 43-lineage validation set. One frozen backbone made 240 forwards; both
serialized scores came from CPU last-linears over the same captured features.

| Reference | Inherited temperature | Raw correct /240 | KO correct /120 | EN correct /120 | EN gated P/C/Q proposals (all correct) |
| --- | ---: | ---: | ---: | ---: | --- |
| Pretrained base |1.0000158548355103|89|16|73|8/4/0|
| Published v0.1 delta |0.65|88|16|72|7/4/0|

Korean had zero gated P/C/Q proposals; those precisions are undefined. English
completion has only 4/15 correct families (26.7% coverage) across 3 declared lineages,
with 4/4 correct proposals for each reference. English question precision is also
undefined with zero proposals. Eligibility remains unassessed and training
steps unknown/null. No arm was selected, qualified or promoted; no Fit,
calibration, final-test or private-task calls occurred. Eight-class NLL changed
1.738268→1.616183 under different inherited temperatures; this does not isolate
head improvement. These results do not establish a useful or qualified teacher.

The one whole-process observation was 24.24 s (CPU user 4.76 s/system 6.78 s), with
peak RSS 2,887,172,096 bytes (2.69 GiB). Worker time was 23.442 s including loading,
guards and extraction; synchronized native full-base forwards totaled 13.229 s,
mean 55.12 ms, range33.77–401.02ms. These are separate timings, not a portable
product latency or controlled speedup comparison.

Sampled MPS driver/allocated maxima were 2,166,046,720 / 1,688,613,632 bytes;
minimum system availability was 51%. MPS samples miss transient peaks and overlap
RSS in unified memory, so they are not summed. The exact 7,864,320-byte feature
cache and all producer output totaled 8,300,158 bytes, below 16 MiB. Inputs/options
were untruncated (151–433 tokens); maximum native reconstruction error was
1.9073486328125e-6, within the unchanged .001 guard. The Go product runtime and
all existing weights remain unchanged. Further model calls or a supervised
probe require a separate prospective study; no teacher-label distillation is
justified by this reference.
