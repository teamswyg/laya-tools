# Tiny decision-head research — tinyhead-01

[한국어](README.ko.md) · [Issue #15](https://github.com/teamswyg/laya-tools/issues/15) · [Parent #11](https://github.com/teamswyg/laya-tools/issues/11) · [Public HF collection](https://huggingface.co/collections/JooYoon/riidolaya-public-research-6abcbd5ddb1917912fc5de38)

This route investigates cheaper decisions about which model should handle a task. It does not build a text-generating LLM or change the default router.

## What actually ran

The parent is the v0.2 fixed-tier head over the frozen Laya encoder. We folded LayerNorm's learned affine transform into the linear layer, then converted that head to FP32, row-scaled INT8 and ternary weights. Ternary weights use {-1, 0, +1} with a scale per output row. A presence bitmap and compact sign stream store the symbols. This independent Go implementation draws on the BITCOS layout idea; it is not compatible with the paper's kernels or a standard model file format.

Python remains only for maintainer export of frozen PyTorch/MPS encoder features from reviewed public synthetic text. Go implements head inference, compression, validation-loss selection, metrics, timing, model loading and package/CI checks. This is **post-training quantization**, not encoder retraining or ternary QAT. Outputs are probabilities for fast / standard / strong, with no automatic model switching or task execution.

Replacing Python does not automatically accelerate MPS kernels. The benefit here is a Python-free head runtime and reproducible local compression/evaluation tools.

## Results and boundaries

[Full results](results.json), [pre-registered plan](plan.json), [measurement scope](measurement.json). Apple M4 Pro, Go 1.27.1. Latencies below are single-run means over 100,000 warm head predictions on cached features. No statistically significant speed ranking is claimed.

| Variant | Artifact bytes | Head latency | Changed winners / 264 | Compression gate |
|---|---:|---:|---:|---|
| Original v0.2 safetensors | 20,772 | Not timed here | Reference | Parent |
| Folded FP32 | 12,332 | 4.26 µs | 0 | Pass |
| Row INT8 | 3,116 | 4.21 µs | 0 | Pass |
| Ternary threshold 0.5 | 660 | 5.91 µs | 0 | Pass |
| Ternary threshold 0.7 | 616 | 5.84 µs | 1 | Fail |
| Ternary threshold 1.0 | 571 | 5.14 µs | 1 | Fail |
| Ternary threshold 1.3 | 533 | 4.76 µs | 1 | Fail |

All variants allocate zero bytes per prediction. The 660-byte head is about 96.8% smaller than its parent file, but slower than FP32. Unpacking may explain the cost; this profile does not establish instruction-level causation. There is no SIMD implementation yet.

Only validation NLL selected threshold 0.5. All train 180 / validation 36 / calibration 24 / former test 24 cases are previously used or viewed synthetic development data. Passing means no unacceptable compression regression on those cases, not independent generalization, coding success, token savings, Korean performance or repository selection. The confidence gate remains 0.9; no new strong→fast errors occurred. Probabilities changed even when winners did not, so broader validation is necessary.

**The 660-byte head is not a complete router.** It still requires the original tokenizer and Laya encoder. Feature export peaked at approximately 2.60 GiB process RSS. The separate Go evaluation process includes the feature JSON, runtime and profiler; it is not full Laya memory. MPS driver memory overlaps with process memory and must not be added to RSS.

The selected ternary weights are 39.7% zero. Symbols occupy 616 bytes and headers/scales/biases another 44. Symbols alone use approximately 1.604 bits/weight, **not below 1.58**. Continuous five-trit packing would take 615 symbol bytes, one byte less; padding each 128-weight block would take 624. Smaller failed candidates are retained, not promoted by relaxing accuracy criteria.

## Reproduction and usage

Regular users should keep using `riidolaya`. These are separate research tools. Once assets are downloaded, local experiments require no HF network connection.

```sh
# Only the first feature export needs the maintainer PyTorch/MPS environment.
.cache/mps-training-venv/bin/python scripts/training/export_tiny_features.py \
  --base .cache/training/base \
  --package .cache/training/pdca-06-package-v0.2 \
  --out .cache/tinyhead/features.json

# All subsequent compression and evaluation is Go. out must not exist.
go run ./cmd/riido-tinybench \
  --features .cache/tinyhead/features.json \
  --out .cache/tinyhead/my-run --iterations 100000

# Always require exact-source CI immediately before publication.
go run ./cmd/riido-hubcheck --package .cache/tinyhead/package --require-ci

go test -race ./internal/tinyhead ./internal/researchbundle
go test ./internal/tinyhead -run '^$' -bench BenchmarkPredict -benchmem
```

A reusable feature cache must identify encoder/tokenizer revisions, pooling and input hashes. This experiment pins parent and dataset hashes and exports only the 264 public synthetic cases. It does not collect private repositories or arbitrary user prompts. Optional `--cpuprofile .cache/tinyhead/local.pprof` produces a local-only profile; disable it for timing comparisons.

The experimental `RDTINY01` format has a 44-byte little-endian header: 8-byte magic, uint32 width, uint32 kind, FP32 temperature, three FP32 biases and three FP32 scales. Payload is row-major 3×width weights: FP32 bytes, signed INT8 bytes, or a continuous presence bitmap followed by positive-sign bits in nonzero order. Padding is zero. Prediction uses float64 accumulation, LayerNorm epsilon 1e-5 and stable softmax. The loader validates bounds, finite numbers, padding and exact payload length. This is not a promised stable public API.

Models are immutable and each caller owns scratch space, avoiding locks. The hot path uses contiguous byte slices and fixed `[3]float64` outputs. Maps exist only in report/manifest preparation. A full ECS architecture is not justified for this small linear classifier.

## Research applied

Sources checked 2026-09-30; pinned code revisions and licenses are in [sources.json](sources.json). No third-party code, additional model weights or external training corpora were copied into this implementation.

- [Laya](https://huggingface.co/convaiinnovations/laya): current parent; retain pinned revisions and Apache-2.0 notices. Its arbitrary-choice head differs from our fixed-tier classifier.
- [Laya-MLX](https://github.com/mizorewww/laya-mlx): candidate for the next Apple Silicon encoder comparison. Its FP16 port and validation work are useful, but another Mac's timings are not ours. Compare the same inputs, pooling, calibration and quality.
- [JEV / TypeSafe](https://typesafe.ai/): informs bounded decision outsourcing. API access does not grant model/data redistribution rights. No paid calls, response training or weight copying took place.
- [Gestalt-Lab/jeff](https://github.com/Gestalt-Lab/jeff): Qwen3-4B adapter; code/adapter Apache-2.0. Adopt separate reporting of calibration and accuracy, without downloading a larger model for this small experiment.
- [firelex/jeff](https://github.com/firelex/jeff): a different project, with MIT code and Apache-2.0-labelled weights. Its leakage checks and preserved failures are useful practices. [Training sources](https://github.com/firelex/jeff/blob/main/docs/data-sources.md) have mixed licenses and are not imported here.
- [BitNet](https://github.com/microsoft/BitNet) and [b1.58 research](https://arxiv.org/abs/2402.17764): references for ternary models and dedicated execution. Reformatting dense Laya weights does not establish an equally capable BitNet model.
- [BITCOS](https://arxiv.org/html/2609.16338v1), discovered through the [linked article](https://news.hada.io/topic?id=33822): inspired the bitmap/sign representation. Do not transfer Intel kernel speed claims to Apple Silicon. Count actual padding, scales and headers.

This is a redistribution-scope review, not verification of all upstream pretraining rights. Any future code reuse needs the pinned source's LICENSE and NOTICE preserved first.

## Next routes and model family

| Route | Relationship | Next question / gate |
|---|---|---|
| Current fixed head | Laya encoder → v0.2 → FP32/INT8/ternary siblings | Same features and labels; compression regression only. Default deployment is separate |
| Ternary QAT | v0.2 features → newly trained ternary head | Recover quality at greater sparsity. Train-only updates, validation-only selection, calibration-only temperature |
| Encoder-free student | Public text → word/character hashes → small linear classifier | Main hypothesis for reducing total memory. Go training; compare lexical rules and abstain out of scope. Do not call it an LLM |
| Smaller encoder execution | Existing encoder → INT8 or MLX FP16 | Measure cold/warm end-to-end latency, RSS/MPS and quality together; verify graph/provider support |

Before another learning/generalization claim, freeze a **new unseen final set grouped by task family/source**: documentation edits, local function changes, failing-test diagnosis, API contract changes and concurrency/data-integrity work. Include easy/hard pairs sharing vocabulary, Korean/English, missing information and out-of-domain cases. Expected tiers are hypotheses, separate from observed tests passed, retries and costs. New real-task outcome labels have not yet been collected.

A student must distinguish original labels from teacher probabilities to avoid inheriting every teacher error. Evaluate accepted precision, coverage, strong→fast mistakes, calibration, multiple seeds and new repositories. Decomposition and repository routing remain sibling heads with their own evidence; difficulty success does not validate them. Run heavyweight experiments serially to respect a 24 GiB Mac's memory/thermal budget. No world-first or lowest-hardware claim is established.

## Public workspace

The [HF experiment package](https://huggingface.co/JooYoon/riidolaya-tinyhead-experiment-v0.1) is published after exact-source CI succeeds. GitHub stores code, plans, small results and hashes. HF stores six heads, public synthetic features/reference probabilities, license/notice and manifest. The immutable Hub commit and verification record belong in [#15](https://github.com/teamswyg/laya-tools/issues/15). Existing v0.1/v0.2 releases are preserved.

The workflow is local experiment → explicitly staged public files → Go integrity check → secret scan → exact-source CI → HF upload → pinned-revision download and all-file hash verification. Publication is independent, so Hub outages do not block training/evaluation. Integrity checks do not replace data-rights review. Never upload raw profiles, credentials, local absolute paths, internal project names or real user inputs. No hosted Space service or paid GPU was provisioned.
