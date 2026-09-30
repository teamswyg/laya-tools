# Real MPS training pilot — 2026-09-30

[한국어](mps-pilot.ko.md) · [Issue #11](https://github.com/teamswyg/laya-tools/issues/11)

This goes beyond the synthetic tensor readiness probe: **the original Laya checkpoint was actually trained on MPS**. The encoder and separate action head stayed frozen; 26,248,193 head/type_emb/scorer parameters were eligible for updates. Both tasks started independently from the same original. Python remains maintainer-only; the user runtime is Go.

## Experiment and results

Used 88 newly authored English synthetic development examples: 48 difficulty, 40 decomposition. See [dataset limitations](../benchmarks/training/README.md) and [all measured JSON](../benchmarks/results/mps-pilot-20260930). The existing 36 goldens were excluded from training. These examples do not replace the planned independent 400-group collection.

Compared learning rates 1e-5/5e-5 with up to two epochs, then extended to five epochs. Runs restart from the base and select checkpoints on validation NLL. Temperature is fitted separately on the development calibration partition within 0.5–5.0. The existing 0.9 confidence gate was not lowered.

| Task | Original development evaluation | Selected tuned model | Accepted at 0.9 | Selection |
|---|---:|---:|---:|---|
| Difficulty | 6/9 | 6/9 | 0/9 | 5e-5, epoch 5 |
| Decomposition | 5/8 | 4/8 | 0/8 | 5e-5, epoch 2 |

Longer decomposition training worsened validation loss, so epoch 2 remained selected. **Neither improved correct-answer counts nor increased usable recommendations was established.** The decomposition regression is preserved. Baseline probabilities use upstream calibration; candidate probabilities use pilot calibration, so probability-metric gains cannot be attributed to weight learning alone.

Conceptual patterns overlap across splits. The first test results were observed before extending training, so all evaluation remains development-only, not independent repository generalization, coding success, or savings evidence. Human difficulty estimates do not establish an automatic execution oracle.

## Memory and storage

Ran one model at a time on a 24 GiB Mac. Extended runs peaked at about 2.75 GiB process RSS and about 3.04 GiB step-end sampled MPS driver allocation. These overlap in unified memory and must not be added; samples are not continuous GPU peaks. Approximately 42 s for difficulty and 32 s for decomposition include loading, evaluation, hashes and checkpoint writes, not throughput benchmarking.

Each run checks a 5 GiB process RSS ceiling, 20% system memory availability floor, 30 GiB free disk and 20-minute limit, plus an MPS allocator cap. No other apps were terminated or global settings changed. The same settings may stop under different workstation pressure.

Store the approximately 804 MiB original once. Publish approximately 100 MiB FP32 replacement head parameters per specialist, avoiding complete encoder duplication. Optimizer/RNG checkpoints remain local and are never automatically uploaded. Check free space before every experiment; the tools do not automatically delete user files or old checkpoints.

## Reproduction

Tested Python 3.14.7 / PyTorch 2.14.0 / Laya SDK 0.3.21; [pinned package versions](../scripts/training/requirements-pilot-macos-arm64.txt) are not a full wheel-hash supply-chain lock. Original artifact hashes appear in release provenance.

```sh
python3 -m venv .cache/mps-training-venv
.cache/mps-training-venv/bin/python -m pip install -r scripts/training/requirements-pilot-macos-arm64.txt
.cache/mps-training-venv/bin/python scripts/training/train_pilot.py \
  --base .cache/training/base --task difficulty \
  --out .cache/training/new-difficulty-version --epochs 5 --minutes 20
```

The base directory requires model.safetensors, rl_agent_config.json, encoder/config.json and tokenizer files from revision `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`. The weight SHA-256 is verified. Existing output directories are refused. Use `--task decomposition` for the other task.

Verified finite loss/gradients, actual head updates, hashes of all frozen parameters, and saved-head inference reload parity with CPU fallback disabled. Optimizer state is saved, but **training-resume equivalence is not verified**. ONNX/INT8/Go conversion remains pending.

## Hugging Face and subsequent versions

Candidate release names: `JooYoon/riidolaya-difficulty-pilot-v0.1` and `JooYoon/riidolaya-decomposition-pilot-v0.1`. Issue #11 records actual publication and immutable revisions. Release v0.1 is distinct from the local extended-run name v0.2.

`package_pilot.py` bundles license, original model cards, NOTICE, modifications, fixtures, metrics, requirements and hashes. `publish_pilot.py` permits only explicit filenames and rejects optimizer state, extra files, symlinks and tampered weights. Publishing requires successful CI for the exact source commit and authenticated JooYoon identity. Existing versioned repositories cannot be overwritten. Tokens come only from the environment/official Hub credential store and are never embedded in artifacts or logs.

Reviewed Laya and ModernBERT model cards declaring Apache-2.0 and the SDK license. New data is original; no noncommercial/share-alike dataset was added. Preserve full terms and attribution based on declared licenses; this is not independent proof of all upstream pretraining rights.

These are head replacement weights, not LoRA adapters or standalone models. Combine with the exact base in FP32. `evaluate_head.py` provides a maintainer CPU reference. The default Go runtime model and automatic execution stay unchanged.

Next prioritize evidence-backed public tasks, label improvement and family-separated data rather than unlimited epochs. Subsequent releases get new versions and the same integrity/CI/license gates. No unattended training/publication schedule has been created.

[Subsequent PDCA tuning and next evaluation preparation](pdca-tuning.en.md)
