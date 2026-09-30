# Real ternary decision-head training and PDCA

[한국어](README.ko.md) · [Issue #17](https://github.com/teamswyg/laya-tools/issues/17) · [Previous compression experiment](../tinyhead/README.en.md) · [Public HF collection](https://huggingface.co/collections/JooYoon/riidolaya-public-research-6abcbd5ddb1917912fc5de38)

We moved beyond compressing already-trained weights: **training predictions use actual packed ternary weights**. Two PDCA cycles ran 32 candidates. Selected heads occupy 572 bytes, or 1.490 bits per linear coefficient including header, scales and biases, divided by 3,072 coefficients. This is not the size or bit-width of the complete Laya model.

## Why decisions rather than generation

The output is three probabilities—fast / standard / strong—not prose. Laya is already an encoder-based decision model, so there is no generation decoder to remove here. Understanding input still requires the encoder. This experiment tests ternary training of the small head after it; neither the encoder nor activations were ternarized.

## What changed in training

1. Start from the same public v0.2 parent, folding LayerNorm's learned affine transform into its linear layer.
2. Keep float64 shadow weights in Go. For every training forward pass, threshold each row relative to its mean absolute weight, producing zeros or scaled {-1,+1} values.
3. Update shadow weights using an identity straight-through estimator (STE). Scale/threshold derivatives are detached. This approximate gradient is not the exact derivative of rounding or a reproduction of the original BitNet recipe.
4. Select checkpoints by validation NLL under a 600-byte budget. Do not read final data. After selecting both seeds, choose temperature using 24 calibration cases only. Keep the confidence gate at 0.9.
5. Seal model, plan and dataset hashes, then evaluate final in a separate command.

Training, evaluation and publication verification run in Go on CPU. Cached development features are reused; only the new final-feature export used Python/PyTorch MPS. Seeds change minibatch order, not independent random initialization: both start from the same parent.

## PDCA record

| Cycle | Plan and execution | Check | Act |
|---|---|---|---|
| 01: supervised labels | 2 seeds × 4 thresholds × 2 learning rates; 80 epochs each | Selected 507-byte heads, but already-confident training predictions produced tiny gradients. Validation 35/36; negligible improvement over matching PTQ | Do not claim meaningful learning progress. Freeze the next plan while final remains unopened |
| 02: preserve teacher margins | Same 16 combinations, 120 epochs each. Supervised CE plus centered teacher-logit MSE weighted 0.1 | 572 bytes, validation 36/36 versus 35/36 for matching pre-training PTQ; 46 ternary symbols changed | Seal both seeds and calibration, then open final once |

The second objective preserves differences between the parent's class scores. Correctly classified examples can therefore still teach information lost during quantization. The teacher is our frozen public parent head, with no external API or paid model calls. Teacher targets are computed only for 180 training rows. The exact auxiliary loss is `0.1 / (2×3) × Σ(student centered logit − teacher centered logit)²`.

[First plan](plan-01.json), [second plan](plan.json), [all 32 candidate summaries](development-cycles.json), [final results](results.json), [paired timings](benchmark.json), [resources/reproducibility](measurement.json). Full epoch traces live on HF as `development-cycle-01.json` and `selection.json`. Experimental descendants do not overwrite their parent.

## Final results and tradeoffs

The 36 new English synthetic final cases were authored and frozen before training, with 12 per class. Labels express expected difficulty from explicit scope, not measured coding outcomes. Existing train 180 / validation 36 / calibration 24 were reused. The former 24-case test was not used for this training, selection or calibration. The new final is now viewed development evidence for any later tuning.

| Metric | Parent FP32 | Previous 660B PTQ | New 572B QAT, both seeds |
|---|---:|---:|---:|
| Correct | 33/36 | 33/36 | 34/36 |
| Correct among accepted, confidence ≥0.9 | 32/33 | 31/31 | 32/33 |
| Coverage | 91.7% | 86.1% | 91.7% |
| Strong recall | 12/12 | 12/12 | 12/12 |
| Strong→fast mistakes | 0 | 0 | 0 |
| NLL, lower is better | 0.171 | 0.128 | 0.179 |
| Brier, lower is better | 0.111 | 0.088 | 0.109 |

Accuracy and size improved, but **the new model does not dominate all quality metrics**. Probability quality and accepted precision worsened versus previous PTQ. Both seeds pass the pre-registered criteria, but one extra correct case out of 36 does not establish statistical superiority. The Wilson 95% interval for 34/36 is approximately 81.9–98.5%. No default-router promotion occurs.

Remaining errors over-classify two simple documentation edits as standard. The rate-limit unit typo is deferred at approximately 0.852 confidence; the watcher-documentation glob change is wrongly accepted at approximately 0.953. Future real-task evaluation must measure unnecessary escalation cost as well as unsafe downgrades. We did not adjust weights or temperature after reading these final errors.

## Below 1.58 bits: what is being counted

Selected heads have approximately 62.63% zeros. The three symbols are not equally frequent. Storage is 384 bytes of presence bitmap + 144 bytes of compact signs + 44 bytes of header/scales/biases = 572 bytes.

- Symbols: 528×8÷3,072 = 1.375 bits/weight.
- Complete file: 572×8÷3,072 = 1.4896 bits/weight.
- The denominator is the **linear layer's 3,072 coefficients**, not the encoder, tokenizer, activations or workspace.
- Scales/biases remain FP32; normalization/accumulation use FP64. Training also holds floating shadow weights and gradients. Inference-file size is not training memory.

M4 Pro, Go 1.27.1, same 36 cached feature vectors, five rotated-order repeats of 30,000 calls; medians:

| Head | Bytes | Head-only latency |
|---|---:|---:|
| Folded FP32 | 12,332 | 4.23 µs |
| Previous ternary PTQ | 660 | 5.73 µs |
| QAT seed 1729 | 572 | 5.01 µs |
| QAT seed 2718 | 572 | 5.03 µs |

This is about 12% less time and 13.3% fewer bytes than previous ternary PTQ, while still slower than FP32. Five samples are not a system-wide performance guarantee. All predictions allocate zero bytes. No SIMD/GPU kernel was added, and these figures do not establish end-to-end encoder latency improvements.

The complete second 16-candidate sweep took approximately 3.25 seconds and 24.1 MiB peak process RSS. That is **small Go head training over cached features**. New-feature extraction used a separate MPS process with approximately 2.51 GiB peak RSS. Processes ran sequentially; RSS and MPS driver memory must not be added together.

## Reproduction

Use the immutable HF revision recorded in the issue. GitHub stores no models or feature arrays. In the existing maintainer environment:

```sh
# PyTorch/MPS is needed only for new encoder features.
.cache/mps-training-venv/bin/python scripts/training/export_qat_features.py \
  --base .cache/training/base \
  --package .cache/training/pdca-06-package-v0.2 \
  --out .cache/ternary-qat/final-features.json

# train rejects a final argument.
go run ./cmd/riido-qat --stage train \
  --development .cache/tinyhead/features.json \
  --out .cache/ternary-qat/new-training

# Run separately after both seeds have been selected and sealed.
go run ./cmd/riido-qat --stage check \
  --development .cache/tinyhead/features.json \
  --final .cache/ternary-qat/final-features.json \
  --selection .cache/ternary-qat/new-training/selection.json \
  --out .cache/ternary-qat/new-check

# Recompute quality and require exact-source CI before publication.
go run ./cmd/riido-hubcheck --package .cache/ternary-qat/package --require-ci
```

Output directories must be new. Use the HF bundle's `development-features.json` directly as `--development` to reproduce training without Python. Use `--plan experiments/ternary-qat/plan-01.json` for cycle 01. `--stage bench` accepts the same data/selection arguments and measures timing only. Publication is separate, so Hub availability never blocks local training.

The bundle verifier goes beyond hashes: it runs downloaded heads on final features to recompute all 36 probabilities, accuracy, acceptance and quality gates. Do not publish raw profiles, internal names, actual user prompts, credentials or local absolute paths. Retain LICENSE and the parent NOTICE.

## Next PDCA

- Do not reuse these 36 cases as a new final. Freeze new families/sources before further tuning.
- Expand train/validation/calibration with complex vocabulary in genuinely simple tasks and other scope/wording contrasts. Move beyond one author's templates toward observed outcomes on public coding tasks.
- Investigate overly easy calibration cases selecting the lowest temperature, 0.5. Do not reduce the confidence threshold to inflate coverage.
- Vary fewer factors per experiment to isolate threshold, learning rate, teacher weight and calibration effects. Consider stricter pre-registered limits on NLL and accepted-precision regression for the next cycle.
- Test total memory reductions through an encoder-free student or low-bit encoder training/execution separately. Head success does not prove whole-Laya ternarization will preserve capability.
- Publish versioned HF descendants linked to parent/data/code/CI/hashes, preserving earlier models and failures. Continuous publication does not mean automatic production promotion.

## References and licensing

This is an independent Go implementation of mathematical ideas. No external training corpus or third-party implementation was copied. Existing Apache-2.0 parent assets retain LICENSE/NOTICE.

- [Ternary Weight Networks](https://arxiv.org/abs/1605.04711): threshold/scaling precedent.
- [Straight-through estimator research](https://arxiv.org/abs/1308.3432): background for approximate gradients; our identity/stop-gradient choices are experimental.
- [Knowledge distillation](https://arxiv.org/abs/1503.02531): transferring teacher information; our centered-logit MSE mixture does not reproduce that paper's results.
- [BitNet b1.58](https://arxiv.org/abs/2402.17764): low-bit training direction, not a whole-Transformer implementation here.
- [BITCOS](https://arxiv.org/html/2609.16338v1): bitmap/sign storage exploiting unequal symbol frequency. Intel performance claims are not transferred to this Mac.
