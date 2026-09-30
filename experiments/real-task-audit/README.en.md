# Auditing real GitHub task sources — experiment28

**The nominal2,594 rows contain2,575 distinct normalized request strings across53 repositories.** This exceeds2,400 numerically, but does not establish independent final evaluation or model performance.

## Acquired sources

Under the [precommitted plan](plan-28.json), download only the test Parquet and README from [pinned SWE-bench Full](https://huggingface.co/datasets/SWE-bench/SWE-bench/tree/c6fe717fd7a4c3ac1daa4055a4fd082c6a1d28a2) and [pinned Multilingual](https://huggingface.co/datasets/SWE-bench/SWE-bench_Multilingual/tree/846e647b9f33c0b51b739d005d13d85493c9af09), locally. No development split, containers or repository checkouts were acquired. Both selected files per source passed remote checksum verification; this is not verification of a complete Hub clone.

| Source | Rows / distinct IDs | Distinct requests | Repositories | Repository + pre-fix commit groups |
|---|---:|---:|---:|---:|
| Full | 2,294 | 2,275 | 12 | 2,174 |
| Multilingual | 300 | 300 | 41 | 300 |
| Combined | 2,594 | 2,575 | 53 | 2,474 |

Requests are lowercased and whitespace-normalized before exact deduplication. Full contains19 additional rows repeating normalized request text; this is not a count of19 duplicate groups. No IDs or normalized requests are shared across these two sources. No empty IDs/requests or malformed repository/commit fields were found.

Against all14,291 rows of the prior CodeSearchNet projection (11,678 distinct normalized requests), exact request overlap is zero and repository overlap is zero. This does not rule out semantic similarities, code copies or pretrained-model contamination.

## Why counts are insufficient

Django accounts for850 rows; Full's2,294 rows come from Python projects. Coverage is not balanced merely because the total is large. Some tasks share a pre-fix snapshot. Future splitting must group connected request duplicates and snapshots, and specify repository-level reporting before evaluation. Related rows must not silently enter both training and evaluation.

These are real issue-resolution tasks. They provide richer context than function-description retrieval, but **patch files are not exhaustive relevance labels, and patch size is not ground truth for model difficulty or decomposition.** Public benchmark tasks may already be known to large models. Newly acquired for this project does not mean unseen by every model.

## Answer separation and implementation

The maintainer converter verifies SHA-256, row count and column types, then reads128-row batches of exactly four columns: `instance_id`, `repo`, `base_commit`, `problem_statement`. Patches, test patches, hints, difficulty, test lists, execution scripts and image fields are excluded from the request projection. They exist in the original Parquet but were neither executed nor supplied to a model.

A small maintainer-only script uses existing PyArrow25.0.1 for Parquet decoding. Go performs validation, counts, deduplication, prior-source overlap and membership hashing with sorted arrays. No runtime Python dependency or new Go dependency is introduced. Public output contains aggregates and an overall membership hash, not raw requests or per-row hashes.

## Licensing and source discrepancies

The [pinned benchmark source LICENSE](https://github.com/SWE-bench/SWE-bench/blob/02e7a74ffd0b707aab73d203fe87bdc7c76afc8e/LICENSE) is MIT. Multilingual's card advertises MIT; Full's pinned card metadata has no license field. This does not establish blanket permission to redistribute every upstream repository's code or issue content under those terms. **Per-source conditions remain unresolved.** This audit is not approval for raw republication or a new trained-model release. Git contains our analysis code, aggregates, hashes and explanations only.

The [official FAQ](https://www.swebench.com/SWE-bench/faq/) advertises42 Multilingual repositories; this pinned file contains41. Full's prose describes JSON-encoded test lists, while both actual Parquet schemas use `list<string>`. The converter checks the real schema. [Source notes and hashes](source-notes-28.json) preserve these differences.

## Verification, resources and next step

The aggregate report is byte-identical on replay. Synthetic tests cover normalization, duplicate identities, cross-source overlap, malformed fields, rejection of answer columns, absence of raw text in output and row-order-independent membership hashes. Converter tests verify answer/execution-field exclusion and source-hash rejection. Full Go race tests, vet and formatting passed.

The two Parquet files total35,169,743 bytes; request projections total5,433,482 bytes. A Go audit including the prior corpus took0.64s wall,0.28s user and75,776,000 bytes peak RSS (about72.3MiB) on M4 Pro CPU. This excludes downloading, Python conversion, training, inference and GPU work.

```sh
# After downloading pinned Parquet files, project each once using the maintainer environment.
python scripts/training/project_real_task_audit.py --source full --input .cache/real-task-full-28/data/test-00000-of-00001.parquet --output .cache/real-task-full-28.jsonl
python scripts/training/project_real_task_audit.py --source multilingual --input .cache/real-task-multilingual-28/data/test-00000-of-00001.parquet --output .cache/real-task-multilingual-28.jsonl
go run ./cmd/riido-taskaudit --out .cache/real-task-audit-new
```

[Aggregate results](results-28.json). Next verify connected groups, source conditions and pre-fix snapshot access costs before separately freezing evaluation-only membership and the input contract. No training, inference or agent execution occurred. Existing final reserves remain unused. No new model/HF release. **Acquiring2,575 candidate requests is not completing independent evaluation.** Not production ready; no LLM savings established.
