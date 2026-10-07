# Shared linear probe over frozen Laya features

This original Go research package learns 1,024 shared float32 weights over eight
frozen 1,024-dimensional option vectors. Scores are `s[i] = dot(w, phi[i])`.
One shared bias would cancel in softmax, so it is fixed at zero. Eight independent
class weight vectors and a text encoder are not included.

`Store` borrows an immutable `io.ReaderAt`; `ReadRow` decodes one 32,768-byte
little-endian row into caller-owned `Workspace`. It does not read the whole
55MB feature file into Go memory. Callers own file lifetime and verify source,
shape, row identity and SHA-256 before fitting. The command provides that binding.

`Model.ScoreFeatures(*Row, *Workspace)` accepts features, never text. Reuse a
workspace serially; concurrent scoring needs separate workspaces and an otherwise
immutable model. Fit/temperature changes require exclusive model ownership.
Novel text still requires the original large Laya backbone. A small head is not
an ultra-low-memory standalone text model.

`Fit` always starts from fresh zero weights and a fresh mini-batch AdamW optimizer.
Its CE gradient is the sum across all eight choices: `sum_i (p[i]-y[i])*phi[i]`.
Defaults are 40 epochs, batch32, lr0.001, decay0.01, seed1729 and temperature1.
Inputs and numeric failures leave the previous model unchanged. Caller labels
are training targets; this package establishes no annotation truth or rights.

RSP v1 is a separate **4,320-byte feature-head format**: 4,096 weight bytes plus a
192-byte header and 32-byte corruption checksum. It pins feature schema,
eight-intent order and the fixed base/instruction contract, temperature, seed and
this head's actual Go update count. It does not encode or invent the backbone's
unknown training history. It cannot load through v1/v2 text `.rsh` or v3 MLP
`.rsm` loaders. Use this package's `Save` and `Load`; checksum is not authenticity.

Tests use original small feature matrices, verify shared gradients with finite
differences, a known AdamW step, row streaming and bounds, actual toy training,
transactional/fresh fitting, and exact trained artifact/prediction reload. They
are code/numerical checks, not native Laya or semantic performance evidence.

The first actual fixed study completed1,326 fit rows and1,680 updates, but failed
its display-support goal. It scored118/354 on exposed internal development and
92/240 on exposed validation, with **zero gated P/C/Q proposals in both locales**.
Precision is undefined, not a perfect result. The4,320-byte trained head passed
exact reload parity; that execution success is not semantic qualification.
See the command's [actual report](../../cmd/riido-statehint-shared-probe/README.en.md)
for metrics/resource distinctions. No selection, calibration, final test or
promotion followed, and novel text still requires the large backbone.
