# Repository-level train/validation/final partition — experiment 37

**Freeze 7,335 training, 5,686 validation and 2,402 new final-evaluation representatives.** Use separate repository sets and keep related tasks together. Final outcomes have not been scored or used for training/model selection. Partition completion does not establish model quality or license approval.

Commit the [rules](plan-37.json) before repository collection and partition execution. Reverify experiment 36's 19,008 source rows and 15,423-candidate fingerprint. Protect all 2,594 previous evaluation rows and prior etcd/Kubernetes/LXD repositories. Preserve the original 2,400-task selection.

| Role | Representatives | Repositories | Intended use |
|---|---:|---:|---|
| Train | 7,335 | 20 | Future parameter fitting |
| Validation | 5,686 | 5 | Model, threshold and stopping selection |
| New final | 2,402 | 10 | Evaluation after selection is complete |
| Total | 15,423 | 35 | Every representative assigned exactly once |

Do not split repositories to force a ratio. Sort whole units by SHA-256 of length-framed canonical repository names, fill final to at least 2,400, then validation to at least 2,400, and assign the remainder to training, requiring at least 4,800. Keeping pandas' 3,639 representatives together expands validation to 5,686. Do not change order, seed or minima after observing the assignment. Validation remains imbalanced, so later results must include repository-specific and equally weighted repository metrics.

Fetch **91 public GitHub repository identities**: 35 training, 53 evaluation and three prior-search repositories. Validate positive numeric IDs, canonical names, explicit public/nonprivate status and fork status. Forks require public parent/source metadata and share their source-ID family. None of the 91 observed repositories were marked as forks. This does not detect copies unrecognized by GitHub.

Observed name changes are `google/jax → jax-ml/jax`, `Lightning-AI/lightning → lightning-ai/pytorch-lightning`, `tiangolo/fastapi → fastapi/fastapi` and prior `lxc/lxd → lxc/incus`. Normalize name casing. Inconsistent canonical names or fork families for one numeric ID abort. Protect matching evaluation/prior IDs/families and reapply transitive task exclusion and representative rules. The candidate count remains 15,423.

Each canonical repository is indivisible. Connect repositories sharing a GitHub network family or a normalized request in any original training row. **Include all 3,585 nonrepresentative sibling rows in these checks.** The actual corpus retains 35 units. Unselected siblings stay with their repository's role and cannot become extra training examples for another split.

An independent sorted-key verifier checks cross-role repository/network/request overlaps, task duplication/omission and role counts. Synthetic tests cover renames, forks, sibling bridges, permutation invariance, missing/private/malformed identities, incorrect task roles and shortfalls without approved membership output. These exact-string checks do not guarantee semantic deduplication, undisclosed fork/alias detection or absence of pretraining contamination.

Initial collection used 91 requests. Enforce at least 300ms between request starts, 30s per request, 2MiB responses and at most 128 calls per invocation. Persist validated responses atomically in 0700/0600 caches; corrupt caches are not silently refreshed. Freeze the collected inventory hash so later GitHub metadata changes cannot silently repartition the same experiment.

```sh
# Reproduce from fixed local inputs and metadata cache
go run ./cmd/riido-trainpartition --out .cache/training-partition-new
# Collection option; reject a changed frozen inventory before partitioning
go run ./cmd/riido-trainpartition --fetch-identities --out .cache/training-partition-collected
```

Ordinary offline replay and full-data race execution reproduce public aggregates and private `membership.json` byte-for-byte. Complete offline M4 Pro CPU execution took 3.70s wall, 3.44s user, 0.06s system and maximum RSS 206,864,384 bytes (about 197.3MiB). Exclude collection network time, model inference and training from these measurements. [Full aggregates and normalized public identities](results-37.json).

License fields are **current GitHub classification hints**, not historical pre-fix license text or reviewed issue/patch permissions. Keep `HistoricalLicensesReviewed=false` and `TrainingApproved=false`. Publish no requests, patches, task memberships or raw API responses; publish aggregates and normalized public repository metadata only.

Next, preserve these roles while reviewing historical licenses/provenance and preparing catalogs and search-cost labels for training/validation. Keep final outcomes unscored until model selection finishes. This prepares a file-search hint task, not validated difficulty, decomposition or model routing. No model training or Hugging Face release occurs in this stage.
