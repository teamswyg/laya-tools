# Development catalogs and historical license collection — experiment 38

**Acquire complete catalogs for 25 representative snapshots and verified license-candidate text for 24. Training/release approval remains incomplete.** Separately collect all 13,021 fixed training/validation tasks. Exclude the 2,402 final tasks from collection.

Freeze the [plan](plan-38.json) before acquisition. A Go reader verifies experiment 37's private membership hash and returns only 7,335 train and 5,686 validation tasks. Check role counts, schema, duplicate IDs, byte limits and digest; a final task reaching the collection loop aborts. Sample the lexicographically first task ID per repository without using size, license or outcome.

| Representative preflight | Result |
|---|---:|
| Repositories / snapshots | 25 / 25 |
| Pre-fix roots / complete catalogs acquired | 25 / 25 |
| File-path references, including symlinks | 43,214 |
| Symlink / submodule references | 116 / 3 |
| Largest file candidate catalog | 6,425 |
| Snapshots with verified license-candidate text | 24 |
| License path/object references / distinct text objects | 52 / 49 |
| Distinct license text bytes | 204,206 |
| API requests / root, catalog or license request failures | 101 / 0 |

[Sample aggregates](sample-results-38.json). File references are not distinct source bodies. Do not follow symlink/submodule targets. Download only bounded license-candidate bodies.

The selected historical Prefect snapshot has no path containing `licen`, `copying` or `notice` among all 107 file candidates/133 tree entries. Experiment 37's current GitHub hint is Apache-2.0, but do not substitute it for historical permission. README or other possible licensing locations have not been inspected, so this is not a finding that no license exists anywhere. [GitHub documentation](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/licensing-a-repository) notes possible README statements and explains file-based automatic detection.

The historical DataDog `LICENSE-3rdparty.csv` lists components under other terms, including LGPL-3.0-only. Airflow and pandas also have separate license directories. Do not apply a single root license to every file, third-party work, issue or model; neither does this establish automatic application of a component's terms to our model. [Historical DataDog notice object](https://api.github.com/repos/DataDog/integrations-core/git/blobs/42674643f6679e769b34cf44af22936fc4d7971f), [Apache-2.0 text](https://www.apache.org/licenses/LICENSE-2.0).

Inspect root LICENSE/COPYING/NOTICE and LICENCE spellings, plus direct files inside LICENSE(S)/LICENCE(S) directories. Skip symlinks and count deeper directories as unresolved. Verify UTF-8 text up to 256KiB by recomputing Git blob SHA-1 and recording SHA-256. Preserve existing catalog bounds: 8MiB responses, 32MiB assembled cache, 100,000 entries and 16MiB paths, with bounded subtree recovery for large responses.

```sh
# Representative snapshot per development repository
go run ./cmd/riido-trainingcatalog --mode sample --out .cache/training-sample-new
# Continue all development tasks using the same cache
go run ./cmd/riido-trainingcatalog --mode all --request-budget 512 --out .cache/training-all-new
# Network-free replay of acquired samples
go run ./cmd/riido-trainingcatalog --mode sample --offline --out .cache/training-sample-replay-new
```

Space request starts by at least 300ms, with 30s per request. Allow at most 2,000 new calls per invocation; the initial full run uses 512. Unavailable tasks remain in denominators after budget exhaustion; subsequent runs reuse caches. Failure counts include budget-deferred acquisition. Never relabel partial availability as complete readiness. `AllDevelopmentCatalogsAvailable` requires all 13,021 catalogs.

Online/offline/race sample aggregates, private object inventories and failure lists reproduce byte-for-byte. Full race/vet/format checks and synthetic role-isolation, order-invariance and count/maximum tests pass. A separately measured cache-only sample run on M4 Pro CPU took 0.10s wall with maximum RSS 34,488,320 bytes (about 32.9MiB). This excludes network, model inference and training.

Twenty-five samples cannot approve every historical snapshot, per-file/third-party condition or issue-text permission across 13,021 tasks. Keep `HistoricalLicensesReviewed=false`, `IssueTextLicenseResolved=false` and `TrainingApproved=false`. Publish original code, plans and repository aggregates only; requests, patches, complete paths, license bodies, memberships and profiles remain local. No outcome scoring, model training or Hugging Face release occurs in this stage.

The first full invocation terminated normally at 512 new requests. Including sample caches, **274/13,021 roots and complete catalogs** are available and **12,747 remain unavailable**. Available catalogs contain 396,820 file-path references, with a maximum of 8,180 candidates. License-candidate text exists for 271 snapshots; three have no matching candidate names. License tree/blob request failures are zero. [First invocation aggregates](collection-first-38.json), private object inventories and failure lists match an offline replay exactly. This records a completed first batch, not live status of later resumptions. Overall readiness remains false.
