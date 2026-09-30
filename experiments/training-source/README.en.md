# Separate training-source identity audit — experiment 36

**Acquire a new pool of 15,423 task groups across 35 repositories. This is not yet a training/validation/final split or model training.** The pool has no overlaps under the checked task-ID, normalized-request, snapshot and repository-name identities with the existing evaluation sources.

The pinned `SWE-bench/SWE-bench` revision has only test/dev splits. The [pinned original distribution](https://huggingface.co/datasets/princeton-nlp/SWE-bench/tree/e48e2bd1e9fecd5bbd641e9414ac59da9f2e69f6) provides 19,008 training rows. Freeze its byte size, SHA-256 and audit rules in the [plan](plan-36.json) before downloading. Do not silently follow latest main.

| Item | Result |
|---|---:|
| Original train rows / distinct IDs | 19,008 / 19,008 |
| Distinct normalized requests / repository-base snapshots | 18,257 / 16,138 |
| Connected task groups / representatives | 15,423 / 15,423 |
| Additional sibling rows | 3,585 |
| Direct shared IDs, requests and snapshots with evaluation | All 0 |
| Shared evaluation repository names | 0 |
| Training rows excluded by protected connections | 0 |
| Candidate repositories | 35 |

Connect rows sharing an ID, lowercase/whitespace-normalized request or repository/base commit, including transitive connections. Protect all 2,594 original evaluation rows, including siblings not selected into the 2,400-task evaluation, and prior docstring repositories etcd, Kubernetes and LXD. Exclude entire components touching evaluation or forbidden repositories; choose the lexicographically first training ID from each remaining component. Inspect no difficulty, labels or search outcomes. [All repository aggregates](results-36.json).

The pool remains imbalanced: pandas contributes 3,639 representatives (about 23.6%) and Qiskit 1,314. These checks do not prove absence of semantic paraphrases, fork/rename aliases or model-pretraining contamination. They use exact strings and the stated normalization/linkage rules. Do not treat all 15,423 as independent observations.

Python is limited to maintainer-only conversion in the existing PyArrow environment. Validate the pinned source hash, 19,008 rows and string fields, then extract only `instance_id`, `repo`, `base_commit` and `problem_statement` in 128-row batches. Patches, hints and test fields never enter the Go input, and dataset scripts are never executed. Projection/audit outputs are 0600; download and audit directories are 0700.

Actual deduplication, connectivity and representative selection use Go sorted arrays and union-find. Only the explicit training reader permits 96MiB and 32,768 rows; existing evaluation-reader limits of 32MiB/4,096 rows and product-runtime bounds remain unchanged. Keep the existing 1MiB per-line bound.

```sh
go run ./cmd/riido-trainaudit --out .cache/training-audit-new
```

Prepare the pinned projection locally through maintainer conversion. The command performs no network calls and writes aggregate `results.json` and private representative `candidates.json` into a new directory. Source Parquet is 106,606,519 bytes; the request projection is 41,811,016 bytes. Do not expand the entire original dataset in memory or install Docker images. The complete Go audit on M4 Pro CPU took 2.04s wall, 1.70s user, 0.04s system, maximum RSS 175,915,008 bytes (about 167.8MiB). These costs exclude download/Parquet conversion and are not model-inference costs.

Projection regeneration, ordinary replay and full-data race execution reproduce their respective files byte-for-byte. Original 2,400-task selection remains unchanged. Synthetic tests cover transitive exclusion, representatives, order invariance, input preservation, duplicate IDs, field isolation, unchanged old bounds, output permissions and no overwrite. PyArrow tests run locally for maintainers; no new Python runtime is added to CI.

Licensing: the pinned dataset card declares no explicit license. Neither a train split name nor the SWE-bench tooling's MIT license establishes blanket permission for upstream issues, patches and derived-model distribution. Review historical repository licenses, provenance and third-party material before determining training and release scope. `TrainingApproved=false` describes unfinished research review, not a legal conclusion. Publish only original code, plans and aggregates, not requests, patches, representative lists or model weights.

Next steps are license/repository-alias review, training/validation/new-final partitioning that keeps related tasks and repositories together, at least 2,400 final tasks, and separate candidate catalogs/training labels. The intended model offers a hint about whether auxiliary search may reduce cost. Changed-file labels do not establish difficulty, decomposition or model-routing ground truth. Existing evaluation sources were not used for training, existing final reserves remain unscored, and no Hugging Face model is released.
