# Local MPS training preparation

[한국어](mps-training.ko.md) · [Plan issue #11](https://github.com/teamswyg/laya-tools/issues/11)

End users need only the Go binary. Python here is exclusively for maintainer training/conversion experiments; it is not added to the `riidolaya` runtime or default CI.

## What was actually verified

On 2026-09-30, an isolated Python 3.14.7 / PyTorch 2.14.0 environment on Apple Silicon with 24 GiB unified memory completed **three synthetic tensor training steps on MPS**. With CPU fallback disabled, the probe verified real forward/backward execution, finite gradients, changed trainable head weights, and unchanged frozen encoder weights. The [sanitized result JSON](../scripts/training/mps-probe-20260930.json) contains no local paths.

- Loss: 1.45736 → 1.41734 → 1.37812, on the same tiny synthetic batch; this is not accuracy or generalization evidence.
- Synchronized three-step interval: about 1.36 seconds, including gradient checks and memory sampling but excluding environment setup. There is no warmup; this tiny probe is not a training throughput benchmark.
- Process peak RSS: about 430 MiB. Maximum sampled step-end MPS driver allocation: about 18.7 MiB.
- Step-end samples do not measure transient GPU peaks. Unified memory overlaps, so RSS and MPS allocations must not be added.

**This did not load or train Laya.** Original weights were not found in the repository cache; no large download or full training was started. Tiny operation success establishes neither complete ModernBERT operation support nor its training memory requirements. PyTorch emits a NumPy initialization warning in this minimal environment; this probe does not use NumPy conversion.

## Reproduce the readiness checks

Run from the repository root. The requirements file pins versions observed for this macOS arm64 probe; it is not a complete Laya training dependency lock.

```sh
python3 -m venv .cache/mps-training-venv
.cache/mps-training-venv/bin/python -m pip install -r scripts/training/requirements-probe-macos-arm64.txt
.cache/mps-training-venv/bin/python scripts/training/mps_probe.py --steps 3
.cache/mps-training-venv/bin/python -m unittest discover -s scripts/training -p 'test_*.py'
```

The probe uses at most ten steps, FP32, fixed tiny tensors, and a 25% MPS recommended working-set allocation limit. It fails when MPS is unavailable or CPU fallback is enabled. It downloads no model, makes no network calls, and writes no checkpoints. The two unit tests use synthetic modules to check the policy contract, not actual Laya integration.

## Remaining gates before actual Laya training

1. Obtain original weights, encoder, tokenizer, and configuration at the [pinned revision](https://huggingface.co/convaiinnovations/laya/tree/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851); record per-file SHA-256, provenance, and licenses. Do not train from the INT8 ONNX distribution. [Preparation metadata](../scripts/training/BASE_MODEL.json) records the reviewed baseline.
2. Lock complete training dependencies and artifact hashes. The reviewed `laya==0.3.21` wheel was downloaded for source inspection, not installed for execution. Check official input formatting against Go formatting and candidate ordering first.
3. Apply `head_policy.py` to train only `head`, `type_emb`, and `scorer`. **Freeze the entire encoder, including final norm, and `act_head`.** The separate action head controls answering/escalation and requires its own targets. Apply the policy after `model.train()`; it sets the frozen encoder to evaluation mode to disable dropout. Reapply after any subsequent `model.train()` call.
4. Begin with one real Laya MPS forward/backward step on a public synthetic input: FP32, batch one, short sequence, compile off, CPU fallback off. Verify finite loss/gradients, updated intended weights, and unchanged frozen weights before extending to 10/50 steps. Do not retry OOM indefinitely.
5. Only after evaluation shows head-only training insufficient, compare gradual encoder-layer unfreezing, activation checkpointing, and gradient accumulation. Separate optimization needs of a fully frozen encoder from those of training upper encoder layers.
6. Split train/validation/calibration/final test by task family. Keep the existing 36 development goldens out of training. Evaluate difficulty and decomposition suitability separately. Laya scores/selects candidates; an upstream generative agent creates decomposition proposals.
7. Calibrate on a separate split and check that stale `temperature_by_options` does not override new calibration. Compare decisions, probabilities, and abstention across CPU/MPS → FP32 ONNX → INT8 → Go before releasing a new immutable artifact version. Do not overwrite the current model.

Full original-model training, resume verification, new-domain accuracy, ONNX conversion, and Go parity have not been performed. Environment readiness is not evidence of successful Laya training or cost savings.

Sources: [PyTorch MPS documentation](https://docs.pytorch.org/docs/2.14/notes/mps.html), [Laya SDK 0.3.21](https://pypi.org/project/laya/0.3.21/), [external MPS experiment log](https://github.com/pilotspace/laya-codex/blob/580bc73c2ed95fd319db93ef725f30bf35047428/finetune/runs/r1-setup.md). External full-model training evidence and this repository's synthetic probe are different results.
