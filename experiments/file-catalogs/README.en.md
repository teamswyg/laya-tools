# Complete pre-fix catalogs for 2,400 tasks — experiment 33

**All 2,400 frozen tasks now have complete pre-fix file catalogs.** Do not remove large repositories or truncate files. Separate labels and an evaluator are prepared, but these findings establish candidate acquisition, not retrieval quality or LLM savings.

## Coverage

| Item | Result |
|---|---:|
| Selected / available / unavailable snapshots | 2,400 / 2,400 / 0 |
| Regular-file / symlink references | 8,039,516 / 5,300 |
| Submodule references | 743 |
| Maximum file candidates per snapshot | 30,223 |
| Maximum file-path byte total per snapshot | 2,838,643 |
| Snapshots exceeding the old 4,096-document bound | 1,008 |
| Snapshots exceeding the old 2MiB text bound | 5 |

References include files repeated across historical snapshots; these are not eight million distinct files. Symlinks remain paths and submodule target repositories are not traversed. `AllSelectedCatalogsAvailable=true` covers complete tree metadata within the selected repository. [Repository aggregates](results-33.json).

## Collection and recovery

Freeze the [initial plan](plan-33.json), recompute identical membership, and anchor each catalog to its pre-fix root tree ID. Download no source bodies. Only validated complete identities/types/paths/parents/unique entries enter the atomic compressed cache.

The initial run ended at its 2,000-request budget with1,995 catalogs. A405-request continuation reached2,395. The remaining five Babel snapshots exceeded the local8MiB response bound. Five diagnostic requests confirmed this;763 requests under the precommitted [subtree plan](subtree-plan-33.json) completed recovery. [Recovery mechanism and limits](REPAIR.en.md).

Never mark a partial response complete. Persist assembled roots only after all children validate. Keep8MiB network responses,2MiB non-recursive validation,100,000 total entries and16MiB total path bytes, with a separate32MiB decompressed assembled-cache ceiling. Ordinary failures/corrupt caches do not silently enter recovery.

```sh
# Prepare pinned source projections and root caches first.
go run ./cmd/riido-filecatalog --request-budget 2000 --out .cache/catalog-pass-1
# Preserve caches and continue into a new output directory.
go run ./cmd/riido-filecatalog --repair-subtrees --request-budget 1500 --out .cache/catalog-pass-2
# Verify complete caches without network requests.
go run ./cmd/riido-filecatalog --offline --out .cache/catalog-replay
```

Use existing GitHub CLI authentication and adjust the budget to remaining API allowance. Output directories must be new. Follow [experiment28](../real-task-audit/README.en.md) sources and [experiment29](../evaluation-groups/README.en.md) selection. Keep `failures.json` and caches private: they contain source identities/local diagnostics.

## Prepared evaluation components

- [Complete-candidate Go search and profiling](SCALING.en.md): up to100,000 paths, global BM25 and baseline-first merging without losing candidates. Existing small APIs retain their bounds.
- [Isolated labels](LABELS.en.md):2,398 tasks have old paths; two add only new files. Labels never enter queries, candidate construction or training.

Publish only original code, plans and aggregates. No training, model release, default activation or final-reserve use occurred. Next compare frozen search policies on actual file catalogs.

## Verification and resources

Collection, ordinary offline replay and an actual2,400-task race run match both public aggregate and private failure list byte for byte. Verify all-task availability and repository-sum conservation. Merge through full Go race/vet, formatting, redacted checks and Linux/macOS CI.

After concurrent race verification finished, one separately timed M4 Pro CPU cache replay took12.16s wall,12.62s user,0.61s sys and46,333,952 bytes maximum RSS (about44.2MiB). This excludes initial network collection, search indexing, models and GPU execution.

Recursive/assembled caches contain3,143 objects and320,112,467 logical bytes; non-recursive tree caches contain10 objects and223,221 bytes. Their combined320,335,688 bytes (about305.5MiB) cover only those two caches, not root metadata, original Parquet, models or the whole workspace. Shared subtrees make object count different from task count.
