# Static alignment learning 05: extra learned features do not beat lexical control

[한국어](README.ko.md) · [Frozen plan](plan-05.json) · [Complete results/hashes](results-05.json) · [Retrieval-source audit](retrieval-source-audit.json)

**Identifier preprocessing and a small learned layer over verified Go static embeddings did not improve the primary: AUC0.5728 versus lexical0.5729.** Additional memory is not justified; no default adoption. Reserve1 (2,403) and reserve2 (2,401) remain unscored. Future results on other data will not erase this failure.

## What trained

Train on11,263 eligible development pairs out of12,189; select on553 of607 validation pairs. Exclusions926/54 retain the previous scope. Increasing representation size does not silently change the cohort. Post-normalization inputs exceeding4,096 bytes error rather than truncate.

- **16 coefficients:** exactly the previous matching control; four length slots stay zero.
- **17 coefficients:** add cosine of normalized query/code static means. Preprocessing splits camel/snake names, lowercases and joins letter/digit tokens. It does not remove keywords.
- **145 coefficients:** add64 elementwise products scaled by8 and64 absolute differences. The head learns which coordinate agreements/differences matter. This is neither a full64×64 mapping nor encoder fine-tuning.

Compare FP32/ternary encoders, FP32/ternary-STE heads, two seeds and two learning rates; lexical needs no encoder. Forty configurations each run100 epochs, batch128/L2=0.0001. Minimum validation BCE selects epoch/rate per configuration family. FP32-derived INT8/PTQ exports yield40 files overall. Original pinned CoSQA grouping/source and Model2Vec revision remain unchanged.

## Results

AUC on the same553 eligible validation pairs. Reused validation point estimates are not independent final results or established statistical improvements.

| Features / encoder / head | seed1729 | seed2718 |
|---|---:|---:|
| BM25 | 0.5378 | same control |
| Lexical / none / FP32 | 0.5729 | 0.5730 |
| Lexical+cosine / FP32 / FP32 | 0.5748 | 0.5745 |
| Lexical+cosine / ternary / FP32 | 0.5742 | 0.5743 |
|145 coefficients / FP32 / FP32, **primary** | 0.5728 | 0.5723 |
|145 coefficients / ternary / FP32 | 0.5682 | 0.5592 |
|145 coefficients / FP32 / ternary STE | 0.5594 | 0.5567 |
|145 coefficients / ternary / ternary STE | 0.5565 | 0.5545 |

JSON contains every export including INT8/PTQ. The primary was frozen as `normalized_alignment / fp32 / fp32 / seed1729` before execution. It misses the AUC≥0.60 and ≥0.03 over BM25 conditions for considering a fresh final. Do not replace it post hoc with the slightly higher17-feature variant.

All40 file hashes reproduce. Eight lexical controls have exactly the previous coefficient arrays; this verifies the comparison baseline did not change. Metadata differs, so this does not mean the old and new file hashes match.

## Cost and use

Apple M4 Pro/24GiB/Go1.27.1 CPU: full loading/preparation/40 fits/exports wall8.84s, userCPU8.58s, maxRSS106,676,224B (101.73MiB), swaps0. Not inference or GPU cost. Deployment would require encoder weights/vocabulary in addition to the head; head size is not whole-model size.

```sh
go run ./cmd/riido-staticprobe --stage setup
go run ./cmd/riido-aligntrain --out .cache/alignment-training-new
```

Requires the previously prepared pinned CoSQA source and a new output directory. Changed plan hashes are rejected. No final/reserve training or selection stage exists. `.align.json` is an experimental contract, incompatible with `riido-hints --model` `.hbin`. No Codex default or execution authority changes.

Tests check unchanged lexical features, held-out mutation isolation, original/post-normalization budgets, common17/145-feature prefixes and the actual embedding-product/cosine relationship. macOS/Linux CI exercises the contract with real pinned encoder assets.

## Next direction: distinguish classification from retrieval

Several representation comparisons failed to establish sufficient improvement on this classification task. Instead of endlessly increasing coefficients/rates on the same validation set, audit data for the intended candidate-ordering use. Retain the classification failures and≥2,400-evaluation requirement.

[CoSQA+ authors](https://github.com/DeepSoftwareAnalytics/CoSQA_Plus) and [paper](https://arxiv.org/abs/2406.11589v7) address multiple suitable candidates and test-based annotations. Inspected revision `1ca3343833a4036a4c42c852eec24b62c1f280d4` had no explicit LICENSE and linked data through Drive. Do not adopt it for training/redistribution until terms are established. Its data-quality discussion does not invalidate our poor scores.

Downloaded the9,751,451B Go test Parquet from the [CodeSearchNet HF distribution](https://huggingface.co/datasets/code-search-net/code_search_net), revision `bd0cf261e357a3eb5c8fba490d23ec1a1cd59555`:14,291 functions,213 repositories,11,680 unique descriptions. No model scores were computed. Documentation is not equivalent to real user queries; unjudged other functions are not automatically negative. Measure known-function rank separately from actual agent-task outcomes.

The [original authors](https://github.com/github/CodeSearchNet#licenses) and HF card require individual source licenses. Initially checked Apache-2.0 root licenses at the actual data commits for kubernetes/test-infra, lxc/lxd and etcd-io/etcd. This does not clear vendored files or all213 repositories. Audit generated code, duplicates, vendor/separate notices, repository grouping and query-text leakage before freezing a retrieval plan. Do not label the entire source MIT.

Raw rows, labels and weights stay outside Git. Training/runtime are Go; local maintainer-only Parquet metadata inspection used Python. No raw source or unverified new model was uploaded to HF. The objective remains evidence for classification, retrieval and actual work separately, not an easier substitute benchmark that hides failure.
