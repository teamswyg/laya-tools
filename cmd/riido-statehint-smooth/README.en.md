# Fixed MLP label-smoothing study

This maintainer command compares two fresh copies of the same
Contextual2048 → 16 ReLU → 8 softmax Go MLP. The only planned training difference
is the target distribution: ordinary hard-label cross entropy (**α=0**) versus
uniform label smoothing (**α=0.05**):

`q[c] = (1 − α) × one_hot[c] + α/8`, with cross-entropy gradient `p − q`.

The correct class receives 0.95625 and each other class 0.00625 at α=0.05.
This known regularization method may reduce some overconfident errors while also
reducing correct gated support. Neither benefit nor an explanation for earlier
completion errors is assumed. This is an isolated development experiment, not
Laya fine-tuning, ternary training or a qualified product update.

Both arms use the exact previous control preparation: the frozen original
train840 corpus, **663 whole-lineage fit families**, the same 64 class-matched
whole KO/EN-family draws, and the same **1,454 ordered training rows**. Draws add
32 completion families and eight each of blocker, reference, progress and
planned. The 177 internal development families have already been exposed; they
are neither fresh qualification data nor independent human gold. No semantic
augmentation or previously trained parent model is loaded.

## Run

Use Go **1.27.1** from the repository root. Supply the approved local original
corpus and frozen split using repository-relative paths:

```sh
go run ./cmd/riido-statehint-smooth \
  --train train840.jsonl \
  --train-sha256 586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e \
  --split frozen-split.json \
  --split-sha256 f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9 \
  --check
```

`--check` validates inputs, split and control draws without fitting, model calls
or output files. Replace it with
`--fit --out .cache/statehint-label-smoothing/NEW-RUN` for the experiment.
The directory must be new. Each arm uses 40 epochs, batch32, lr0.02, AdamW
decay0.001, seed1729, temperature1 and **1,840 updates**. Architecture,
initialization and shuffle recipe stay the same. There is no alpha sweep or
recipe override flag.

The α=0 arm must reproduce the previous actual MLP artifact SHA-256 exactly:

`43d41cb923b33bbbd96879488376ad5c3f600bdd53860d6d6fba852751c0c0b4`

If it does not, the command retains partial failure evidence and aborts **before
fitting α=0.05**. It never changes or bypasses this pin to force a result.

Optional `--validation exposed-validation120.jsonl --validation-sha256 SHA`
evaluates previously exposed balanced diagnostic data after both fixed fits;
its labels and scores have zero training, calibration or selection weight.
Calibration, final-test, augmentation, parent-model and alpha override inputs
are unsupported. The existing family evaluator and **confidence0.9 / margin0.05**
gates remain unchanged. Research selection uses eligible internal development
severity cost, then eight-intent NLL, then fixed arm order (α=0 first).
There is no deployment, promotion or fresh generalization claim.

## Records and loading

A private 0700 run directory contains exclusive 0600 files: exact source inputs,
whole-family draws, the ordered label/text-hash training receipt, all eight
evaluation probabilities in original row order, aggregate reports, both actual
trained models and exact reload evidence. Raw source/evaluation records stay
local. Publish only separately reviewed public aggregates and assets.

Both models are **RSM v3**, 131,872 bytes, with 32,920 float32 parameters.
Load them with `pkg/statehintmlp.Load(io.Reader)` and predict with its caller-owned
`Workspace`. The shared result type is `statehint.Prediction`. v1 CLI/SDK and v2
linear loaders cannot load this format. The artifact contains inference weights;
the recipe and each arm's fit report record label smoothing. The command verifies
saved SHA-256, every original/reloaded prediction and byte-identical
reserialization. The checksum detects corruption, not publisher authenticity.

Fit/prediction wall time and Go heap after fit are reported. Go heap is not OS
peak RSS; measure total process peak RSS externally, such as `/usr/bin/time -l`
on a prebuilt executable on macOS. This is pure Go CPU work, with no GPU claim.
One sequential run does not establish a stable performance benchmark.

Original temporary tests verify exact control preparation, objective wiring
against direct package fits, default α=0 behavior, actual baseline-mismatch
abort, annotation isolation and both trained models' exact save/load parity.
Those tiny fits verify code execution, not semantic performance.

## Actual first development run

The fixed run completed on both actual training arms. The α=0 artifact exactly
reproduced the required previous SHA-256; both 131,872-byte v3 artifacts passed
exact prediction and byte reload checks. Each arm used the same 1,454 training
rows and 1,840 updates. No calibration, final-test or promotion occurred.

| Measure | Hard CE, α=0 | Smoothing, α=0.05 |
| --- | ---: | ---: |
| Already exposed internal development: raw eight-intent correct | 322/354 | 305/354 |
| Internal development: correct gated / all gated proposals | 91/95 | 39/39 |
| Internal development: correct completion / completion proposals | 35/37 | 14/14 |
| Internal development eligibility | Failed | Eligible; research selection only |
| Exposed diagnostic: raw eight-intent correct | 174/240 | 158/240 |
| Exposed diagnostic: correct gated / all gated proposals | 52/55 | 11/11 |
| Exposed diagnostic: correct completion families, KO / EN | 5 / 1 | 0 / 0 |
| Exposed diagnostic qualification | Failed | Failed |

Smoothing had no gated errors in either evaluated corpus, and internal correct
completion support was KO5 / EN9. It traded the hard arm's two internal wrong
completion families for lower coverage. On the exposed diagnostic it made **zero
completion proposals**: completion precision is **undefined**, not zero or a
perfect success rate. These results provide no useful diagnostic improvement or
fresh deployment evidence. The research selector chose α=0.05 using internal
development only; the diagnostic did not influence selection.

Mean training loss is not directly comparable between the arms because hard CE
and smoothed CE use different target distributions. The complete sequential run
took about 2.20 seconds with externally measured process peak RSS 27,197,440
bytes; these are whole-run measurements, not per-arm or GPU measurements.
