# Difficulty-model PDCA and preparation for the next tuning round

[한국어](pdca-tuning.ko.md) · [Execution checklist](https://github.com/teamswyg/laya-tools/issues/11)

## What this experiment measures

Train a **decision model** to classify a task description as fast, standard, or
strong. It does not generate code or decomposition plans. Difficulty improvement
comes first; decomposition remains a separate task.

Current labels follow an authored difficulty rubric. They do not establish which
Codex model actually succeeds or how much subscription usage routing saves.
Passing the existing confidence gate of 0.9 never grants execution authority.

## The repeatable loop

1. Plan: freeze family-separated train/validation/calibration/test data and record plan/data hashes in the issue before training.
2. Do: train two seeds independently from the same original checkpoint; select only by validation loss.
3. Check: freeze both candidates, then evaluate the new final, legacy English cases and choice-order permutations.
4. Act: preserve failures. Viewed final data becomes development data; each new cycle needs a new final.

Two seeds are repeated training on the same data, not two independent datasets.
Family separation still shares authorship, style and conceptual patterns.
Repeatedly checking new small tests until one passes can introduce selection
bias; publish all failures and the number of attempts. This is not independent
repository generalization.

## Unchanged initial-improvement criteria

- Final accuracy at least 80%, no regression against the original.
- Strong recall at least 87.5%; zero strong-to-fast errors.
- At the unchanged 0.9 gate, coverage at least 25% and accepted precision at least 90%.
- At least 10 percentage points accuracy gain OR 10% relative Brier reduction.
- At most 10% winner flips under alternative choice orders.
- At most 5 percentage points legacy English accuracy regression.
- Both seeds pass every condition.

The original and tuned model use the same calibration data and temperature grid.
Calibration has 12 cases in cycles 1–3 and 24 fresh cases in cycle 4; both are small. Six choice-order evaluations of one case
are not six independent examples.

## Reproduction and resource limits

Frozen plans and fixtures live in [PDCA-01](../benchmarks/training/pdca-01),
[PDCA-02](../benchmarks/training/pdca-02) and
[PDCA-03](../benchmarks/training/pdca-03) and [PDCA-04](../benchmarks/training/pdca-04).
Python is maintainer-only; the user runtime stays Go.

Freeze encoder/action-head weights and train head/type_emb/scorer on MPS with
CPU fallback disabled. Check frozen hashes, finite losses/gradients, checkpoint
reload and artifact hashes. Run one model at a time with a 5 GiB RSS ceiling,
20% system memory availability floor, 30 GiB free disk and 20 minutes per seed.
RSS and sampled MPS driver memory overlap in unified memory; do not add them.

CI checks frozen plans and family boundaries without torch or model downloads.
Those offline checks are distinct from actually rerunning MPS training.
Released weights are approximately 100 MiB replacement heads that require the
exact original; they cannot replace the current Go ONNX artifact directly.
Publication requires a new immutable version and successful CI for the exact
source commit.

## Prepared next-round inputs

[Six public task candidates](../benchmarks/training/public-task-candidates.json)
reference an immutable public commit in this repository. They cover supplied
comment edits, bounded conditions, concurrency and permission handling. They
share two code families, so they are not six independent families. Each records
an expected tier, rationale and completion conditions. Model execution outcomes
remain empty; these candidates are excluded from current training.

Next steps:

1. Collect new public task families and Korean cases; review reasons and ambiguity.
2. Record success, retries, latency and actual usage under identical task/budget/verification conditions for each execution profile.
3. Target the least costly successful profile; leave tasks without measured outcomes unresolved.
4. Evaluate decomposition separately using parent/child dependencies, integration failures and total cost.
5. After Go ONNX conversion, measure CPU/MPS/ONNX decision/probability parity, p50/p95, RSS and native memory.
6. Compare old and new versions side by side before considering a default-model change.

Record origins and licenses. Do not collect or publish private code, customer
prompts or credentials. No new paid model executions or unattended recurring
training schedule have been started.

Cycle 4 adds cross-order Jensen–Shannon divergence to training loss and the validation selection objective. Inference remains one pass in one option order.
