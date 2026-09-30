# Bounded streaming of development path targets — experiment 39

**All 13,021 development tasks were accounted for: 12,669 yielded supported old-file paths, while 292 blank patches, two oversized patches and 58 parsed tasks without old paths remain separate.** This is target preparation, not model training or a quality improvement. The 2,402 final tasks and unselected siblings are excluded.

The [original plan](plan-39.json) was committed first. Inputs are the pinned experiment-36 Parquet and experiment-37 role membership. Development IDs are filtered at Arrow level before selected patch records are materialized into maintainer application objects. This does not claim that Parquet internals never decompress other rows. Final-role patches are not analyzed, labeled or scored.

The first projection hit its planned 16MiB row limit: one JSON row was **117,856,291 bytes**. A diagnostic replay reproduced the failure; incomplete outputs were removed. Before real target analysis, an [oversize amendment](oversize-amendment-39.json) retained the limits rather than raising them. Arrow computes selected patches' UTF-8 lengths before materializing strings. Above the existing 2MiB parser limit, only ID, original byte count, an oversize flag and `patch:null` are emitted. These tasks have unknown targets and count as failures, never valid empty targets.

| Category | Training | Validation | Total |
|---|---:|---:|---:|
| Fixed tasks | 7,335 | 5,686 | 13,021 |
| Supported tasks with old paths | 7,294 | 5,375 | 12,669 |
| Whitespace-only patches | 0 | 292 | 292 |
| Oversized patches | 1 | 1 | 2 |
| Parsed tasks without old paths | 40 | 18 | 58 |
| Old-path references | 21,715 | 15,399 | 37,114 |
| Maximum old paths per task | 265 | 137 | 265 |

[Full per-repository report](results-39.json). The 294 parse failures comprise 292 blank and two oversized patches. Oversized original texts were 8,020,190 and 102,783,478 bytes. Parsed targets contain 2,401 new-file blocks. There are zero unsupported blocks in parsed patches, not zero parse failures. The 58 cases without old paths have not been scored as search successes or failures.

A separate Go development reader streams the entire SHA-256 verification **before** decoding a file bounded at 256MiB, then processes JSONL rows bounded at 16MiB. It also verifies the second pass's digest. Binary search over sorted development IDs and a visited array reject missing, duplicate and out-of-scope tasks. The existing evaluation reader retains its 32MiB/4,096-row contract. Per-record raw patches are released and only path strings are copied; no whole-file raw buffer, global lock or cache is added. No SIMD speedup is claimed.

Python/PyArrow remains maintainer-only Parquet tooling; no Python dependency is added to the runtime, Go CLI or CI. No patch application or dataset code execution occurs. Raw inputs, patches and per-task targets remain local 0600 files. GitHub receives only code, plans and aggregates.

```sh
# Private projection in the existing maintainer PyArrow environment
python scripts/training/project_development_labels.py \
  --input .cache/training-source-36/data/train-00000-of-00001.parquet \
  --membership .cache/training-partition-37/membership.json \
  --output .cache/development-patches-new.jsonl
# Verify the pinned digest/plans and write aggregates/private targets to a new directory
go run ./cmd/riido-developmentlabels \
  --projection .cache/development-patches-new.jsonl \
  --out .cache/development-labels-new
```

The final private projection has 13,021 rows and 110,290,640 bytes; SHA-256 is `6bed55ca053705e59b1362f982b86cbd977fa3d216312a018b1657dfb9fd284b`. Reprojection reproduced identical bytes. Normal replay and actual full race execution also reproduced identical public aggregates and private targets. Synthetic checks cover excluded/final IDs, missing/duplicate identities, schema/hash/size boundaries, contradictory oversize metadata, blank patches, input preservation and independent path storage. Full race, vet, formatting and redacted scans gate publication.

Acquisition was running separately, so this step does not report timings as isolated performance. Arrow's memory while decompressing large chunks differs from the Go reader's memory. The bounded structure is an implementation property, not a measured RSS improvement claim.

Next, join targets against complete pre-fix file catalogs. This report does not establish that all 37,114 paths exist in candidate catalogs. Targets are for learning/evaluation only, never inference features. Historical licenses, issue-text permissions and training approval remain unresolved. No difficulty, decomposition, model-routing or actual LLM savings claim follows from these targets. No new model or Hugging Face release is produced.
