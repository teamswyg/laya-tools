# Ternary weights for a tiny claim model

This isolated experiment tests smaller storage and repeated inference for development-comment response, activity and completion hints. The target is riidolaya’s Go classifier trained on original synthetic data. It neither converts pretrained Laya weights nor distills Laya outputs; encoder comparisons remain separate records. Preserve the original float32 research model. Small files alone do not prove semantic quality or faster processing.

## What changes

Weights become `-1, 0, +1`, independently of the three output classes. Each of the three categorical heads retains its own probabilities. One absmean scale is stored per head matrix; effective weights are scale times trit. Features, biases and scales remain floating point. This is a weight-only classifier adaptation, not a BitNet LLM or complete W1.58/A8 implementation.

Five trits fit in one byte because `3^5=243≤256`. Complete groups occupy **1.6 physical bits per weight**, distinct from `log2(3)≈1.585`. The 18,432-weight payload is3,687bytes; the RQT format including biases/scales/header/checksum is4,023bytes. Report artifact size, in-memory model size, Go heap and whole-process RSS separately.

Inference decodes only addressed active-feature trits; it never reconstructs a complete float matrix. Models are read-only and workspaces caller-owned, without global caches or inference locks. Feature extraction and float arithmetic remain; measure CPU performance before claiming acceleration.

## Four fixed comparisons

1. Unchanged published float32 reference.
2. Matched float32 continuation: same parent, data/order and40extraepochs as QAT.
3. One deterministic post-training ternary projection (PTQ), without Fit.
4. One warm-start ternary-aware continuation (QAT) from the original float master parameters, fresh optimizer moments and step zero. Preserve inherited2,120training steps separately.

QAT uses a ternary forward pass with float master parameters and a fixed identity STE gradient-copying heuristic. It is not the true derivative or a claim of matching official BitNet training. Fixed AdamW0.001/decay0.01/batch32/seed1729/40epochs/T1 preserves the existing Go weight-decay/bias semantics. The matched float continuation was included before any experiment outputs to distinguish quantization from merely extra optimization. No hyperparameter sweep.

Per head: `s=max(mean(abs(W)),1e-5)`, `t=clip(RoundToEven(W/s),-1,1)`. Exact ±0.5 ties map to zero. Float64 means followed by float32 scales specify our deterministic Go adaptation without claiming bitwise PyTorch parity.

## Basis and usage

References: [BitNet b1.58 paper](https://arxiv.org/html/2402.17764v1#S2), [pinned Microsoft converter](https://github.com/microsoft/BitNet/blob/85a507c67dd2e361f94152c1e8babff2cb4e7daf/utils/convert-hf-to-gguf-bitnet.py#L955-L960), [identity STE](https://arxiv.org/pdf/1308.3432), and the distinct [clipped STE](https://arxiv.org/pdf/1602.02830). No MIT source implementation is copied; code, data and original weights use Apache-2.0 provenance. Shout-out to [laya.tools](https://laya.tools/).

Use Go1.27.1 and explicit local `riidolaya claims --model MODEL.rqt --jsonl`; RSC remains supported. RQT records mode, parent SHA, inherited steps and new optimizer updates. There is no auto-registration, download or application mutation. Outputs remain research previews.

Keep0.9confidence/0.05margin gates. Comparisons on already-exposed240validation rows are diagnostics, not new testing, selection or calibration. Private work is neither trained on nor published; old/new calibration-test bodies remain sealed. Raw profiles stay local. Model binaries belong in immutable HF research versions, outside Git.

## Actual measurements: 2026-10-07

We ran one float continuation and one QAT continuation on the same 1,680 original synthetic training rows. Each added 2,120 optimizer updates for 4,240 total historical updates. PTQ added zero fits. These two runs bring the historical actual Fit count to 20. Sealed calibration and test data were not used.

| Condition | Artifact | Per-model storage | Diagnostic mean CE |
| --- | ---: | ---: | ---: |
| Original float | 73,988 B | 73,784 B | 0.416372 |
| Matched float continuation | 73,988 B | 73,784 B | 0.352619 |
| PTQ | 4,023 B | 3,880 B | 0.558685 |
| QAT | 4,023 B | 3,880 B | 0.428378 |

Lower CE is better. QAT improved on PTQ but underperformed the matched float continuation. QAT emitted no Korean completion proposals. Its four English completion proposals were correct but below the predeclared minimum support. PTQ emitted no positive proposal in any locale/head cell. **None of the conditions qualifies for operational use.**

A profile identified repeated checked packed-value access. A 1,215-byte shared lookup table replaced that access without changing learned weights, training, or arithmetic order; exact bit-parity tests passed. Single before/after measurements on the same 16 fixed synthetic inputs, 20,000 iterations each, showed QAT at approximately 17.70→7.21µs and PTQ at 17.64→7.12µs. In the latter round original float took 7.19µs and matched float 6.18µs. These observations suggest comparable speed to original float, not superior CPU speed or a latency distribution.

The 3,880-byte per-model storage excludes the 1,215-byte shared decoder and the 30,744-byte caller workspace. Whole-process peak RSS after optimization was about 10.0–10.7MiB across conditions, so the roughly 19-fold model-storage reduction is not a 19-fold process-memory reduction. Reused-workspace prediction tests measured zero allocations; some whole-process benchmark counters also included one unattributed 16-byte allocation. This Go experiment used neither GPU nor MPS.

All model/fixture SHA pins and prediction checksums matched exactly before and after optimization; no additional refit was performed. See the [public aggregate record](RESULTS.development.json) for exact values, denominators, and measurement scopes. Real-work comment text, individual predictions, and raw profiles are excluded.
