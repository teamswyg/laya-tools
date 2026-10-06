# Same-data linear / small MLP capacity study

This maintainer command asks whether a small nonlinear representation helps the
eight-intent state hint model. Both arms receive the **same 1,454 ordered training
rows** and use the same Contextual2048 features. One arm is the existing fresh
zero-initialized linear softmax model; the other is a fresh 16-unit ReLU MLP with
Glorot initialization. This compares an **architecture bundle including
initialization**, not an isolated activation, word-order effect, Laya fine-tune,
ternary model or product improvement.

The original frozen train840 corpus is split by the existing whole-lineage rule:
663 fit families / 177 internal development families. The command repeats the
exact previous control draw: 64 whole KO/EN families from fit only, ranked by the
existing seed1729 SHA recipe, with 32 completion families and eight each of
blocker, reference, progress and planned. There is no semantic augmentation input
and no previously trained parent artifact. The 177 development families have
already been exposed; this is a development comparison, not fresh qualification
or independent human gold.

## Run

Use Go **1.27.1** from the repository root. Supply your approved local corpus and
frozen split; this repository does not publish raw development/diagnostic data.
Replace the example paths with local repository-relative paths. Input hashes are
fixed; they are not knobs for trying another dataset.

```sh
go run ./cmd/riido-statehint-capacity \
  --train train840.jsonl \
  --train-sha256 586f1862241bf0e43734494911d503b6aaedac979215f9fdd802f43e5c7a660e \
  --split frozen-split.json \
  --split-sha256 f5f661243684b0b65e171dbbc78c024c31264dc1c00481e1ba04926da39dede9 \
  --check
```

`--check` validates the fixed corpus, split and draw without a model call or output
directory. Replace it with `--fit --out .cache/statehint-capacity/NEW-RUN` to run
the fixed experiment. A new output directory is required; old runs are preserved.
For both arms the recipe is 40 epochs, batch32, lr0.02, AdamW decay0.001, seed1729,
temperature1 and **1,840 updates**. There are no recipe override flags.

Optional `--validation exposed-validation120.jsonl --validation-sha256 SHA`
evaluates that previously exposed, balanced diagnostic after each fixed fit. Its
labels and scores never train, retune, calibrate or select an arm. Calibration,
final-test, augmentation and parent-model inputs are unsupported.

The existing family evaluator and **confidence0.9 / margin0.05** gates stay
unchanged. Eligible internal development results are ranked by severity cost,
then eight-intent NLL, then fixed arm order (linear before MLP). Selection is a
research result; even an eligible result does not deploy, promote or establish
fresh generalization. Failed runs keep partial artifacts and `FAILED.json`.

## Outputs and loading

Outputs use a private 0700 directory and exclusive 0600 files: the pinned source
inputs, exact whole-family draws, ordered sample label/text-hash receipt, all
eight probabilities in original evaluation row order, aggregate reports and
actual trained artifacts. Keep this complete run directory local; it contains
raw source and evaluation metadata. Publish only separately reviewed public
aggregates and assets.

| Arm | Artifact | Bytes | Correct Go loader |
| --- | --- | ---: | --- |
| Contextual linear | RSH v2, `.rsh` | 65,728 | `pkg/statehintwide.Load(io.Reader)` |
| 16-unit ReLU MLP | RSM v3, `.rsm` | 131,872 | `pkg/statehintmlp.Load(io.Reader)` |

Both loaders return the shared eight-intent `statehint.Prediction` type. Use the
matching package's caller-owned `Workspace` for prediction; workspaces have
different types. The MLP has 32,920 float32 parameters. v3 is **not compatible**
with the v1 CLI/SDK or v2 linear loader. The command verifies the saved bytes,
reloads through the correct loader, compares every original/reloaded prediction
exactly, and verifies byte-identical reserialization. Artifact SHA-256 provides
corruption detection, not authentication.

The report records fit and prediction wall times plus Go heap after fit. Go heap
is not total memory or OS peak RSS; measure process peak RSS externally for the
whole run (on macOS, build once and use `/usr/bin/time -l` on the executable).
This is pure Go CPU execution, with no GPU claim. Sequential arm timings and a
single run do not establish a stable performance benchmark.

Synthetic temporary tests cover split/draw class matching, development and
diagnostic annotation isolation, both actually trained formats and exact reload
parity, shared prediction/evaluator types, input budgets and fresh private output.
Those tiny fits verify execution; they are not semantic performance evidence.

## Observed development run

Each actual arm trained on the identical ordered 1,454 rows for 1,840 updates.
The linear artifact exactly reproduced the previous control SHA-256. Internal
dev is already exposed; this is not a fresh qualification evaluation.

| Model | Internal raw correct | Internal gated correct/proposals | Completion precision | Internal gate | Diagnostic correct | Diagnostic gated correct/proposals | Diagnostic completion KO/EN |
|---|---:|---:|---:|---|---:|---:|---:|
| Linear | 317/354 | 73/74 | 26/26 | Pass/research selection | 172/240 | 39/39 | 4/15·0/15 |
| MLP16 | 322/354 | 91/95 | 35/37 (94.59%) | Fail | 174/240 | 52/55 | 5/15·1/15 |

Two MLP false-completion families/lineages, one per locale, failed the fixed
98% precision gate. Internal NLL improved; global overconfidence is not proved.
Both exposed diagnostics remain unqualified due to English completion support.
Diagnostic scores did not alter internal selection.

All trained reload Prediction fields and reserialized bytes matched exactly.
The whole two-arm run took 1.95s (user CPU 1.87s), with maximum process RSS
24,887,296 bytes (~23.73MiB). This is one whole-process observation, not a
controlled memory improvement against earlier runs. GPU, calibration, final
test and operational promotion calls remain zero. Prior results are preserved.
