# Actual file localization — experiment 34

**Across 2,400 real repair tasks from 53 repositories, aggregate search pages fell about 9.0%, but 378 tasks worsened versus 118 improved. Withhold default activation and model promotion.** This is a lexical file-search comparison, not evidence of Laya/ternary-model quality or LLM savings.

Freeze the [plan](plan-34.json) before the first ranking. [Overall and repository results](results-34.json) reproduce byte-for-byte in a separate execution. Use the same 2,400 complete requests and complete pre-fix paths without truncating long queries or large repositories. No training, threshold tuning or best-result selection occurs.

| Policy | Hit@1 count | Hit@10 count | Hit@100 count | First-target page sum |
|---|---:|---:|---:|---:|
| Original-path BM25 | 171 | 769 | 1,509 | 45,377 |
| Identifier-normalized BM25 | 173 | 780 | 1,524 | 42,946 |
| Baseline-first interleave | 171 | 785 | 1,530 | 41,272 |

Hit-rate denominators include all 2,400 tasks. Primary interleaved Hit@10 rises from 32.04% to 32.71%, about 0.67 percentage points. Pages contain 20 candidates; costs apply to 2,398 tasks with existing-file targets: 118 improved, 378 worsened and 1,902 tied. Two new-file-only tasks remain in hit-rate denominators but are excluded from applicable costs. All 3,912 old-file labels map to candidates. Parsing failures, ranking failures and auxiliary fallbacks are all zero.

The precommitted gate of non-increasing total pages and non-decreasing Hit@10 passes. However, repository page sums improve for 16 repositories, worsen for 23 and tie for 14. Equally weighted repository Hit@10 decreases from 31.56% to 31.39%. Django contributes 3,296 of 4,105 net saved pages, about 80.3%. These are post-hoc descriptive analyses, not changed acceptance criteria or statistical-significance claims. The 2,400 tasks are not necessarily semantically independent.

`fileeval.Rank` accepts only requests and file paths. Labels enter `Measure` after ranking. Compare original search with existing `NormalizeText` identifier splitting; distinct file identities remain separate even when normalized text matches. Auxiliary input-limit failures fall back to baseline. Candidates include executable and symlink paths, with no extension, test or vendor filtering and no symlink traversal.

On Apple M4 Pro CPU, the entire offline process took 52.07 seconds wall, 54.33 seconds user CPU and 1.97 seconds system CPU, with maximum RSS 113,491,968 bytes (about 108.2MiB). The 39.41-second evaluation loop includes cache reads, both index builds, ranking, scoring and output. All three policies run within one process: this does not establish their separate runtime difference or pure inference time. No GPU, model inference or training was used.

```sh
go run ./cmd/riido-fileeval --out .cache/file-eval-new
```

Pinned query/label projections and all local catalogs are required. No network is used; any missing catalog aborts before ranking. This guard was verified against an incomplete cache. Publish only aggregates, not requests, patches, per-task labels, caches or profiles.

Page costs assume perfect equal-cost target recognition; changed files are not exhaustive relevance labels. Actual agent success, tokens, latency savings, difficulty, decomposition and repository selection remain unverified. Treat these observed results as development evaluation for subsequent tuning, with separate training data and a new final evaluation. Existing final reserves remain unscored. No model changed, so there is no Hugging Face release.
