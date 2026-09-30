# Low-cost feature learning 03: 16 coefficients improve validation, reserves preserved

[한국어](LEXICAL-03.ko.md) · [Pre-run plan](lexical-plan-03.json) · [All results and hashes](lexical-results-03.json) · [Previous real-data training](TRAIN-02.en.md)

**Replacing 8,192 cross-feature coefficients with 16 lexical coefficients improved validation ordering. The primary AUC of 0.5729 still misses the prespecified 0.60 threshold. Reserve1 (2,403) and reserve2 (2,401) remain unscored; no promotion or default integration.** This is a cheap lexical control for a semantic model to beat, not semantic understanding.

## Changes

The previous probe hashed many query-word/code-word combinations into a small table. This control learns 12 features including bias: token overlap ratios, set similarity, adjacent bigram matches, character-trigram overlap and declaration-name matches. Identifier splitting aligns spellings such as `parseHTTPResponse` and `parse_http_response`. An additional variant adds four length features, exposing possible source-length shortcuts. The matching variant leaves these four slots zero.

Declaration-name extraction heuristically reads a token after `def`, `func` or `function`; it is not a language parser. Quoted keywords and Go receiver declarations can confuse it. Other features use the entire code input, so this heuristic does not truncate model input. Evaluation uses CoSQA Python pairs, not Go repository tasks.

Training and scoring are Go. The shared BCE trainer now supports explicit dimensions while retaining its 8,192-dimensional default and rejecting out-of-range feature indices. The new feature contract is `riido-lexical-features-v1`; weights are incompatible with existing `.hbin` features. Experimental JSON exports cannot be passed directly to `riido-hints --model`.

## Matched comparison

The partition and scope remain unchanged: 11,263 of12,189 development pairs train; 553 of607 validation pairs select/diagnose. Length exclusions are926/54. Every number below uses the same553 eligible validation pairs. This is not a comparison against the earlier2,400-pair final results.

| Method | Validation AUC seed1729 | seed2718 |
|---|---:|---:|
| BM25 | 0.5378 | same control |
| Token overlap | 0.5502 | same control |
| Matching FP32 | 0.5729 | 0.5730 |
| Matching INT8 | 0.5730 | 0.5732 |
| Matching ternary PTQ | 0.5701 | 0.5701 |
| Matching ternary STE | 0.5772 | 0.5762 |
| Matching+length FP32 | 0.5786 | 0.5781 |
| Matching+length INT8 | 0.5786 | 0.5781 |
| Matching+length ternary PTQ | 0.5835 | 0.5835 |
| Matching+length ternary STE | 0.5762 | 0.5762 |

Two variants × two seeds × two learning rates × two training modes yield16 configurations, each100 epochs, batch128/L2=0.0001. Minimum validation BCE selects epoch/rate separately per variant/seed/mode. Derived INT8/PTQ exports produce16 artifacts overall. AUC is a post-selection diagnostic; shrinking score magnitudes can preserve ordering, so small PTQ loss differences do not establish improved accuracy.

The pre-run primary is matching FP32 seed1729. Considering a fresh final requires **validation AUC≥0.60 and ≥0.03 above matched BM25**. Only the second condition passed. We do not replace the primary with the best exploratory variant or lower the threshold. These are point estimates on reused validation data, not confirmed statistical improvement, candidate-retrieval gains or LLM savings.

## Cost and reproduction

- Apple M4 Pro/24GiB/Go1.27.1 CPU: complete run wall1.61s/userCPU1.27s/maxRSS58,638,336B (55.92MiB)/swaps0. Includes loading, partitioning, features,16 fits and exports. The previous225.84MiB run differs in features/configuration count/epochs; this is not a like-for-like speed comparison.
- Readable JSON exports including provenance/mode/weights are509–712B. Sixteen float64 coefficients occupy128B, excluding parsing, tokens and process overhead. **JSON is not packed1.58-bit storage.** Ternary describes values/training; execution uses floating point.
- Public short Go-example feature microbenchmark: three run means5,469/5,407/5,410ns/op,6,360B/op,134allocs/op. Not p95,64-candidate retrieval or end-to-end agent cost. Allocation reduction remains possible, but quality and actual retrieval validation take priority.
- All16 model SHA256 hashes reproduced. Tests cover camel/snake/Unicode, length ablation, finite empty inputs, held-out mutation isolation, dimension bounds and exact default-dimension compatibility. Raw source, models and profiles stay outside Git.

```sh
go run ./cmd/riido-lextrain --out .cache/lexical-training-new
go test ./internal/lexicalhint -run '^$' -bench BenchmarkFeatures -benchmem
```

Requires the pinned local CoSQA source and a new output directory. A changed plan hash is rejected. Exit0 only means execution success; `PrimaryReadyForFinal=false`, `Promotion=false`.

## Next PDCA

Shared matching features transfer better on this validation set than indiscriminately adding word crosses. Synonyms and relationships between descriptions and code behavior remain weak. Keep this control and investigate static semantic embeddings or compact shared lexical representations under a separate frozen plan, measuring incremental memory and quality. Collision/capacity ablations may help isolate causes. Do not announce practical improvement before an independent final evaluation.

Preserve the CoSQA distinction between MIT code and C-UDA data; use local computation and public aggregates. No weights were committed or uploaded to HF pending format-specific verification and license/content review. This is separate research from the original Laya router, not replacement of its encoder with16 coefficients.
