# Laya implementation and conversion diagnostic 18

**Go and Python INT8 outputs agree on these 12 inputs. Original-weight FP32 versus INT8 maximum drift is 0.45 percentage points in normal order and 20.76 pp in reverse order.** This is an implementation diagnostic without quality labels. It does not replace an independent ≥2,400-question evaluation or establish usefulness, probability calibration, or superior original-model quality.

## Comparisons

Six original synthetic states cover code, Korean, whitespace, Unicode, special tokens, and truncation; each uses both option orders. Inputs match the preceding local diagnostic. Commit the [plan](plan-18.json) and [asset hashes](assets-18.json) before observing outputs. Neither program opens training, validation, final, or reserve corpora.

1. Compare Go token IDs and marker positions with installed `laya==0.3.21` `build_sequence`.
2. Execute the same INT8 ONNX file through Go and Python ONNX Runtime 1.30.0 on CPU, applying the same temperature. The reference uses upstream sequence construction and direct ORT execution; this is not an end-to-end test of Python `ONNXAgent` APIs, hooks, or batching.
3. Run the original F16-stored weights from `tindang/laya-code@25f97e5a2ec5f8cf7218a4f67504367d8832e1fe` as FP32 CPU through the installed Laya implementation with strict state loading. No Python code from the model repository is downloaded or executed.

| Metric | Result |
|---|---:|
| Exact token/marker matches | 12/12 |
| Maximum Go/Python INT8 probability difference | 1.73 × 10⁻¹⁸ |
| Go/Python winner disagreements | 0/12 |
| Normal-order maximum original FP32/INT8 difference | 0.004498, or 0.45 pp |
| Reverse-order maximum original FP32/INT8 difference | 0.207607, or 20.76 pp |
| Overall maximum original FP32/INT8 probability difference | 0.207607, or 20.76 pp |
| Mean per-case maximum original FP32/INT8 difference | 0.019952, or 2.00 pp |
| Original FP32/INT8 winner disagreements | 0/12 |

All Go replay comparison metrics match after excluding elapsed time. The entire Python reference was not repeated. Its file hash includes timing and can change on regeneration.

There are six cases per order. The overall maximum occurs in reverse order; do not cite it as normal-use maximum drift. Order stratification was added after observing aggregate results as descriptive analysis, not a new quality gate.

These cases do not support a Go-port explanation for the relevance failure. However, FP32/INT8 drift combines **PyTorch execution, ONNX export, quantization and kernel differences**. Without an FP32 ONNX control it is not an isolated quantization measurement. Preserving 12 winners does not prove preservation of real-query ranking, AUC, or threshold decisions.

## Resources and reproduction

Apple M4 Pro, macOS ARM64, four CPU threads, 512-token maximum; programs run sequentially. Python reference: 5.69 s external wall, peak RSS 3,143,892,992 bytes (~2.93 GiB). Go comparator: 2.31 s external wall, peak RSS 1,600,077,824 bytes (~1.49 GiB). Python executes both INT8 and original weights while Go executes INT8 only; these are not language-performance comparisons. One external timing run, full-process RSS, no GPU or Go-heap claim.

Check 15-minute and sampled 8-GiB RSS limits between calls, not as hard within-call limits. Add the ORT 1.30.0 wheel to the maintainer environment only. Eight selected original-checkpoint files (~846 MB) pass HF hash verification at the pinned revision. Four unselected remote files and local Hub metadata produce missing/extra warnings; used assets have separate pinned hashes. Preserve the Apache-2.0 LICENSE/NOTICE locally. No new model release.

```sh
hf download tindang/laya-code README.md LICENSE NOTICE rl_agent_config.json encoder/config.json tokenizer/tokenizer.json tokenizer/tokenizer_config.json model.safetensors --revision 25f97e5a2ec5f8cf7218a4f67504367d8832e1fe --local-dir .cache/laya-code-source-18
python scripts/training/laya_parity_reference.py --model-dir .cache/models-v2/code --source-dir .cache/laya-code-source-18 --out .cache/parity18-reference.json
go run ./cmd/riido-parityprobe --reference .cache/parity18-reference.json --model-dir .cache/models-v2/code --runtime /path/to/pinned/libonnxruntime.1.30.0.dylib
```

Use the existing maintainer environment with `onnxruntime==1.30.0`; Python is not a user-runtime requirement. [Results](results-18.json) record package versions and aggregates. Choose a new output path to regenerate; the script refuses overwrite. The comparator checks matching asset pins, valid probabilities, exact inputs, and absolute probability tolerance 10⁻⁵. Passing implementation parity is separate from model quality.

Next evaluate original FP32 on **the same validation queries, input protocol and eligibility scope as 17b** to determine whether drift changes relevance discrimination before choosing a training/distillation path. Preserve the unused 2,403/2,401-question reserves.
