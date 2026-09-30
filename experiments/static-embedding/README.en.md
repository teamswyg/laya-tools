# Static semantic embeddings 04: Go parity and packed execution, weak code relevance

[한국어](README.ko.md) · [Frozen plan](plan-04.json) · [Results](results-04.json) · [Source hashes](source-audit.json)

**A pure-Go static embedding runner now matches upstream and keeps ternary weights packed in memory. CoSQA validation quality remains below our lexical control, so no promotion.** Reserve1 (2,403) and reserve2 (2,401) remain unscored.

## Implementation and provenance

[MinishLab Model2Vec](https://github.com/MinishLab/model2vec)'s [potion-base-2M](https://huggingface.co/minishlab/potion-base-2M) is a general-English static embedding whose model card declares MIT. Pin model revision `389b9f64be5aa4ae7a6bc6fe95ef20ce485ae5da` and reference implementation commit `38c7bc9c93808ee15602265ba96a9b8fda58783f`. This is a separate comparison family, not a port/replacement of Laya's encoder.

Go implements WordPiece/BertNormalizer, averages64-dimensional token vectors and normalizes. No automatic CLS/SEP; unknown IDs are removed as upstream does, while literal special tokens are handled separately. Empty or all-unknown input yields a zero vector, not meaningful relevance evidence. Inputs must be valid UTF-8 and at most4,096 bytes; oversized inputs fail instead of silently truncating.

The existing Laya NFC/ByteLevel/BPE tokenizer is incompatible and is not reused. Exact weight/tokenizer hashes are mandatory. Runtime vocabulary uses sorted arrays and weights are contiguous; a temporary map is used during JSON loading only. Inference has no maps, mutable cache or locks; calls own scratch/results.

## Reference checks

- 34 newly authored public sentences cover accents, Korean/Chinese/Japanese, literal specials, controls, emoji, empty input and long words. All token IDs match; max FP32 absolute error8.94e-8 versus tolerance2e-6.
- 1,214 texts from all607 validation query/code pairs also match IDs exactly; max error1.19e-7. This tokenizer check includes the54 pairs outside the earlier lexical scope. Relevance AUC still uses the same553 eligible pairs. No reserve is used.
- Pinned Python Model2Vec is a maintainer oracle only. User execution, downloading and actual CI tests require no Python/ONNX/Torch. Only34 original public reference cases enter Git; CoSQA source/vector traces remain local.
- Race tests, CGO-disabled execution, explicit packed arithmetic examples, corrupt assets and input-budget failures are covered. CI downloads verified real assets and checks saved reference outputs, not mocks alone.

## Compression and measured cost

| Mode | In-memory weights plus scales | Validation AUC | Replay short-text encode p95 |
|---|---:|---:|---:|
| FP32 | 7,559,168B | 0.5289 | 5.04µs |
| INT8, per-token scale | 2,007,904B | 0.5286 | 3.75µs |
| Ternary, per-token scale | 590,560B (576.72KiB) | 0.5394 | 4.83µs |
| BM25 control | separate | 0.5378 | not measured here |
| Previous16-coefficient matching FP32 | separate | 0.5729 | not measured here |

These are point estimates on the same553 eligible validation pairs, not confidence intervals or fresh final results. Report54 exclusions explicitly. Primary FP32 misses the prespecified AUC≥0.60 and ≥0.03 above BM25 conditions. Do not replace it post hoc with ternary because of a small apparent gain. No task-specific fitting occurred this cycle.

Ternary uses per-row mean absolute value as scale and retains signs at magnitude≥0.7×scale. Two bits/coefficient plus a four-byte scale per64 coefficients gives **2.5bits/coefficient**, not1.58. Packed bits are read directly, but accumulation is float32; no GPU or specialized SIMD kernel. Startup reads the7.56MB FP32 source then packs it, so disk storage is not590KB.

Apple M4 Pro/24GiB/Go1.27.1,2,000 calls on one original public sentence. First p95 FP32/INT8/ternary:4.21/3.63/4.67µs; replay values above. Each912B/op and31allocs/op. Replay live Go heap after GC: approximately8.44/3.14/1.80MiB. **First whole-process maxRSS including loading was31.73/33.58/32.33MiB: compressed modes did not lower this peak.** Distinguish table size, live heap and process peak; tables exclude vocabulary/runtime. Short-text encode cost is not64-candidate retrieval or end-to-end agent latency.

## Human/agent usage

```sh
# Once: download immutable public assets and verify SHA256
go run ./cmd/riido-staticprobe --stage setup
# Offline thereafter: one JSON object in, vector/token IDs/table bytes out
printf '%s\n' '{"text":"parse HTTP response headers"}' | \
  go run ./cmd/riido-staticprobe --mode ternary
# Requires the previously prepared pinned local CoSQA source
go run ./cmd/riido-staticprobe --stage validation --mode fp32 \
  --out .cache/static-validation-new.json
```

Validation output must be a new file. This is an experimental vector supplier; it changes neither the default `riido-hints` model nor Codex registration. Maintainer reference generator: `scripts/training/static_reference.py`; users and CI consume saved public references.

## Next steps and licensing scope

Plain mean-vector cosine did not transfer general-English semantics to this code-relevance task. Retain the lexical control while testing identifier normalization, a small learned query/code alignment layer and, if justified, task-specific static embedding tuning. Extra memory without quality gains is not a reason to adopt a model. Go packed execution is now available, but candidate retrieval and LLM savings remain unverified.

MIT is the model author's declaration, not proof of all upstream-data rights or absence of pretraining overlap. Preserve the CoSQA distinction between MIT code and C-UDA data; no raw rows/labels are redistributed. Upstream weights remain local and were not copied to Git or our HF account. New-model publication requires provenance, verification and license/content review.
