# Preparation records for a tiny claim model: the 79-request experiment

This archive prepares a controlled test of whether a small model can help choose which candidate to verify first. The intended output is a hint, rather than authority to approve code or deployments. The model should contribute to a verification system whose final conclusions remain independently checked.

## What changes in this experiment

Three new public requests are appended to the original 76 for training only. Each has three candidates, yielding nine new supervised labels: three positive and six negative. Public UUID and ordinal-function behavior was executed before labeling, and request and candidate-caption fidelity were separately reviewed. Passing the finite example inputs does not establish correctness for every possible input.

The recipe remains the first FP32 recipe: 8,192 features, seed 1,729, 50 epochs, batch 128, learning rate 0.1, L2 0.0001, pair loss 0, and the earliest strictly lowest validation binary cross-entropy (BCE). BCE learns whether a candidate satisfies its request from positive and negative labels. Training starts fresh, without inherited weights or result-dependent changes to the recipe. This path tests the effect of added data. Ternary-weight experiments remain separate and cannot be combined into this result.

The new requests enter training only. The existing validation and calibration views, role and group membership, ambiguous-label masks and original feature values are preserved. Unknown candidates are not converted into negative truth. Earlier work already exposed the existing validation data, so it is a regression view. A useful-model conclusion still requires enough genuinely new requests per domain and a separate protected final-evaluation design of at least 2,400 requests. Three requests or dozens of example inputs are not an equivalent evaluation population.

## Reading evaluation numbers

The model's order is compared with existing deterministic orders on the same validation data. The goal is to reduce candidate checks while preserving ranking quality. Check counts are simulations calculated from saved labels, not measured LLM calls, Codex tokens or billing savings. This path uses a small Go feature-based CPU model; it is not an original Laya GPU execution experiment.

## Accounting for retained bytes

The original total 64 MiB plan is not retroactively marked as passing. Even the confirmed narrow subset retains 67,943,460 B, exceeding 64 MiB by 834,596 B. That excess and the original files are preserved. The latest fixed selector snapshot verifies 2,776 paths and 206,707,066 B. It records two actual scientific fits, two logical model artifacts and six physical model-copy paths. Distinct paths count separately even when hashes or inodes match. Completeness is limited to the fixed selector and snapshot; it is not a whole-machine, deleted-history or lifetime-consumption proof.

The next stage plans a separately versioned policy with a 512 MiB retained per-path ceiling and a 64 MiB ceiling for newly created stage data, models and numeric records. The original eight-fit and sixteen-model ceilings remain unchanged, with the existing two fits and two models consumed. No cleanup or aliasing retrospectively removes the old excess. Public byte copies and verification records also require additive retained accounting. Freezing a resource policy does not activate a model or approve a public upload.

## Frozen historical preparation status

This bundle freezes preparation before the actual 79-request projection and new fit. Both new drivers' actual runs are zero at this snapshot. Any later Root execution must be added as separate results, resource measurements and readback evidence; historical preparation files must not be overwritten to imply an earlier execution.

`loader79-guard-v2` binds source and executable pins, saved synthetic checks and the independent source-review snapshot. `fit79-guard-v2` contains the fixed training recipe, saved synthetic checks and guards. Public-source CI and these private derivative tools' local synthetic checks are separate evidence. Source copies are inert `.go.txt` review text. Executables, model weights, private absolute paths and original corpus bodies are excluded. Retain the repository's Apache-2.0 LICENSE and NOTICE as well.

`plan56-consumption-audit/final-v3` is the latest retained-byte and consumption snapshot. The 2,768-path, 206,581,572 B registry in `history` predates the discovered selector omission and must not authorize execution. Its failure records and the later console-display correction retain their original bytes. `PUBLIC-COPY-LEDGER.v1.json` and `PUBLIC-FINAL-PINS.v1.json` verify the public copies.
