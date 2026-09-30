# Laya code relevance signal probe 17 / 17b

**The current INT8 code checkpoint does not pass the CoSQA relevance gate.** Documented-format validation AUC is 0.5040 versus matched BM25 0.5378. These outputs are not established as useful teacher labels for distillation. Defaults and weights remain unchanged.

## Why two plans

Initial plan-17 put query and code in state and asked a generic `choice` question. While it ran, inspection of the model card and existing search caller established that the code derivative was trained for `noul` relevance. **Before opening either aggregate result**, separately commit plan-17b to test that documented format. Preserve the original run as a different protocol diagnostic.

17b follows the existing `internal/search` instruction/options:

- State: `File: [unavailable]` followed by code. CoSQA has no actual path metadata; do not invent a filename.
- Instruction: `Is this source code relevant to the software change: "{query}"?`
- Options: `false: no, the statement does not hold`, `true: yes, the statement holds`.
- Primary score is true probability in that order. Reversed order and averaged scores are diagnostics only.

The source code model targets change-to-code relevance. CoSQA natural-language search without paths differs from that training context. This experiment does not rule out all Laya checkpoints, FP32 models or latent representations.

## Scope and results

Read the full source to verify its SHA-256 and original partition, and fit BM25 statistics on development code only. **Native model inference uses validation only.** Of 607 validation questions, 553 meet the existing 4,096-byte/64-word scope; 54 do not. No additional tokenizer exclusions occurred. Neither reserve nor the previously used final split was model-scored.

| Protocol | Primary AUC | Forward AUC | Reverse AUC | Averaged-score AUC | Order decision disagreements |
|---|---:|---:|---:|---:|---:|
| 17 generic choice (diagnostic) | 0.5373 | 0.5181 | 0.5481 | 0.5373 | 2/553 |
| 17b documented noul | 0.5040 | 0.5040 | 0.5277 | 0.5214 | 217/553 |

Matched BM25 AUC is 0.5378 in both runs. Mean absolute order probability gap is 0.2644 for 17b. Reversed options differ from the training order; do not select/average them after observing results to claim improvement. Both precommitted gates, AUC≥0.60 and ≥0.03 above BM25, fail. Scores are not established as calibrated out-of-domain probabilities or action authority. This is exploratory evidence on repeatedly used validation data, not an independent final evaluation.

## Runtime and reproduction

Verify hashes of the existing `code-int8-v2` model/tokenizer/config and macOS ARM64 ONNX Runtime 1.30.0. Model SHA-256: `0e3665bcf1aa7d8b224af93b5e10d8c86796227f7741da92f6c8a786364a44db`. Actual inference runs in Go through ONNX Runtime CPU, four threads and a 512-token maximum. No MPS/GPU execution or new training.

17b performs 553 × two orders = 1,106 native calls. Single-call median is 62.777 ms, p95 95.269 ms. External measurements: 80.73 s wall, 307.10 s user CPU, 1,577,156,608 bytes peak RSS. Summed CPU time can exceed wall time with multiple threads. The application's internal total of 80.322 s has a different measurement boundary. Run 17 took 80.44 s wall and 1,574,371,328 bytes peak RSS. These are full-process measurements, not tiny head sizes.

Use one native session sequentially. Check one-hour and sampled 8-GiB RSS limits between calls; this does not impose a within-call memory hard limit. Progress logs contain counts only. No source text, queries or per-row predictions are published. Results include source/partition/model/plan hashes and a complete score-trace hash. Full native repeatability has not been established.

```sh
go run ./cmd/riido-layaprobe --protocol noul --model-dir .cache/models-v2/code --runtime /path/to/pinned/libonnxruntime.1.30.0.dylib
# Separate diagnostic reproducing the initial general-choice protocol:
go run ./cmd/riido-layaprobe --protocol choice --model-dir .cache/models-v2/code --runtime /path/to/pinned/libonnxruntime.1.30.0.dylib
```

Unit tests verify split isolation, positive-option mapping in both orders, the documented forward-score primary and matched exclusions. Test doubles are not represented as native results.

Next verify upstream-reference parity and fitness for actual usage tasks before scaling training or ternary distillation. Preserve the distinction between CoSQA data conditions and checkpoint Apache notices. No new model or source corpus was published to Git/HF.
