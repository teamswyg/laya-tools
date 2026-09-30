# Encoder-free learning 01: INT8 promise, ternary failure

[한국어](README.ko.md) · [Frozen plan](PLAN.md) · [Selection](selection.json) · [Results](results.json)

This is an actually trained **small linear query/document relevance scorer in Go**. It uses no Laya weights, encoder or teacher outputs and is separate from difficulty routing and the earlier Laya-head QAT. Training and optional local CLI inference work, but general code-search usefulness remains unproven.

The documents “cache keeps active entries and removes inactive entries” and its opposite contain identical words in different order. The model learns associations between query retain/discard and document keeps/removes using word/bigram cross-features. BM25 ties these documents. Train/validation/final use 20/6/6 distinct domains, respectively 80/24/24 queries and 40/12/12 candidate documents. Final documents never enter training negatives. Grammar and verbs are shared deliberately: this is a one-author synthetic domain-noun transfer probe, not unseen grammar/operations, Korean or real-user evaluation. Now-viewed final data must become development data in subsequent tuning.

| Model | File bytes | Top-1 correct /24 (seed1729 /2718) | Mean inspections | Frozen gate |
|---|---:|---:|---:|---|
| BM25 | no weights | 12 /12 | 1.500 /1.500 | control |
| FP32 | 32,792 | 24 /24 | 1.000 /1.000 | pass |
| INT8 PTQ | 8,216 | 24 /24 | 1.000 /1.000 | pass |
| Ternary PTQ | 1,215 | 15 /12 | 1.583 /1.917 | fail |
| Ternary STE | 1,274 /1,272 | 16 /15 | 1.542 /1.625 | fail |

Frozen gates: recall@4 ≥95% and mean inspections ≤80% BM25. INT8 reduces inspections by 33.3% on this probe. Ternary is smaller but increases inspection cost and is not adopted. Under bounded BM25 interleaving, INT8 remains at 1.0; ternary STE yields 1.583/1.625. These are authored-label oracle inspections, not measured LLM calls saved.

A **post-evaluation diagnostic control** explicitly replaces retain→keeps and discard→removes, then counts word/bigram overlap. It obtains 18/24 and mean 1.25. It was not preregistered or used for model selection. Stronger grammar-specific rules may solve this narrow task without learning; model necessity is not established.

## Reproduce and use

```sh
mkdir -p .cache/semantic-learning
go run ./cmd/riido-hinttrain --out .cache/semantic-learning/local-train
go run ./cmd/riido-hinttrain --stage check \
  --selection .cache/semantic-learning/local-train/selection.json \
  --out .cache/semantic-learning/local-check
```

Output directories must be new. `check` writes research results; its exit status is not a deployment gate. Inspect per-model `Pass` and scope. Green CI does not establish real-task model quality.

```sh
go run ./cmd/riido-hints \
  --model .cache/semantic-learning/local-train/int8-1729.hbin \
  --sha256 7db5ca98a5fd4426cead07cce2bef897ccf164d3e60ecd731c8ef598f1ed65e8 \
  < examples/hints/request-model.json
```

The pinned hash is this revision/data's seed1729 INT8 artifact. Check selection provenance before using another version's hash. The example reuses an already-viewed final request and is not new quality evidence. Without a model, the CLI uses BM25. External hints and a model are mutually exclusive. Output remains unverified and records the model hash. Model input is limited to 256 candidates and 4096 bytes/64 words per query/document. Oversized inputs fall back to BM25 with `bm25_model_input_out_of_scope`, without truncation. Hash mismatch/corrupt weights fail.

## Implementation, resources and limits

8192 coefficients; signed hashed query/document unigram/bigram cross-features plus overlap ratio. Float shadows, softmax cross-entropy and SGD with weight decay. Two seeds × two rates × FP32/ternary STE = eight candidates, 60 epochs each. Select by validation mean rank, breaking ties with NLL. Ternary forward uses global mean absolute shadow scale, threshold0.7×scale and identity STE with detached scale. Failure applies to this method, not all ternary learning; scale heterogeneity is a follow-up hypothesis.

FP32/INT8 or presence-bitmap/nonzero-sign packed files decode into 8192 float64 coefficients (64KiB) for reference scoring. **Serialized size is not decoded RAM or a native ternary-kernel performance claim.** No GPU/SIMD claims.

Plan/data hashes were fixed before fitting. Eight selected models were stored before final evaluation. Training replay with the final file absent reproduced all eight hashes. Export/decode coefficient equality is checked.

Apple M4 Pro, Go1.27.1: whole training wall2.34s/userCPU1.18s, maximum RSS26,607,616B (25.4MiB), zero swaps. One INT8 CLI process: maximum RSS5,701,632B (5.44MiB), wall0.37s including startup/loading, zero swaps. This is not warm p95 or total agent memory. No GPU used.

Code and original synthetic data are published under the project's Apache-2.0 terms. No third-party weights/data were copied. Weights stay local and excluded from GitHub; a new HF release awaits a strict publication verifier for this format. Overall usefulness remains an active goal.

Updated requirement: prepare [at least2,400 evaluation requests](../semantic-scale/README.en.md). A pass on the24-case probe is not a deployment recommendation.
