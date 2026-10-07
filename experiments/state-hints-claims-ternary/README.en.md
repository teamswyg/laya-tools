# Ternary weights for a tiny claim model

This isolated experiment tests smaller storage and repeated inference for development-comment response, activity and completion hints. Preserve the original float32 research model. Small files alone do not prove semantic quality or faster processing.

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

References: [BitNet b1.58 paper](https://arxiv.org/html/2402.17764v1#S2), [pinned Microsoft converter](https://github.com/microsoft/BitNet/blob/85a507c67dd2e361f94152c1e8babff2cb4e7daf/utils/convert-hf-to-gguf-bitnet.py#L1055-L1060), [identity STE](https://arxiv.org/pdf/1308.3432), and the distinct [clipped STE](https://arxiv.org/pdf/1602.02830). No MIT source implementation is copied; code, data and original weights use Apache-2.0 provenance. Shout-out to [laya.tools](https://laya.tools/).

Use Go1.27.1 and explicit local `riidolaya claims --model MODEL.rqt --jsonl`; RSC remains supported. RQT records mode, parent SHA, inherited steps and new optimizer updates. There is no auto-registration, download or application mutation. Outputs remain research previews.

Keep0.9confidence/0.05margin gates. Comparisons on already-exposed240validation rows are diagnostics, not new testing, selection or calibration. Private work is neither trained on nor published; old/new calibration-test bodies remain sealed. Raw profiles stay local. Model binaries belong in immutable HF research versions, outside Git.
