# Experiment 45: expanded checkpoint costs and source readiness

**Compare actual search costs for 3,369 tasks. Total candidate pages decrease from 33,897 to 31,921, while the 107 validation tasks worsen from 499 to 506. Validation coverage remains insufficient for fitting, and some historical sources have restrictive conditions. This does not approve a default policy or model release.**

Freeze the [plan](plan-45.json) before inspecting new costs. [Results](results-45.json) use only the 3,388 catalogs after experiment44's first fair batch. New caches from ongoing acquisition remain outside this checkpoint. Preserve experiment43's checkpoint and results.

## Denominators and comparison

| Role | All tasks | Catalogs | Scoreable | Baseline pages | Helper pages |
|---|---:|---:|---:|---:|---:|
| Training | 7,335 | 3,281 | 3,262 | 33,398 | 31,415 |
| Validation | 5,686 | 107 | 107 | 499 | 506 |
| Total | 13,021 | 3,388 | 3,369 | 33,897 | 31,921 |

Keep 9,633 unavailable catalogs, one available-task parse failure and 18 available tasks without old-file targets in coverage. The full target corpus has 294 parse failures; the others are unavailable at this checkpoint. The [target join](join-45.json) maps all 10,640 paths from 3,369 tasks to regular files at their exact snapshots. Gold never enters ranking inputs.

There are 171 improvements, 417 regressions and 2,781 ties. Validation has five improvements, 13 regressions and 89 ties; Hit10 declines from 39 to 38. Aggregate gains do not establish generalization or actual LLM savings. A candidate page is a proxy for reading 20 candidates until the first target. This runs auxiliary search always and is not a trained-model result.

## Source conditions joined to costs

The [source assessment](source-assessment-45.json) verifies pinned distribution provenance, 132 historical document records and 11,269 snapshot references. [Repository summary](source-summary-45.json) and [readiness](readiness-45.json) are public aggregates; per-task identities, raw text, paths and document bodies remain private.

The official [SWE-bench README](https://github.com/SWE-bench/SWE-bench/blob/02e7a74ffd0b707aab73d203fe87bdc7c76afc8e/README.md) describes preprocessed training data and the project's MIT license. This supports review of policy41's narrow local numeric scope, not proof of every historical issue author's rights. Global `IssueTextLicenseResolved` and `SourceEligibilityComplete` remain false.

Do not classify historical Prefect using its current Apache label alone. Keep 112 snapshots under the [early EULA](https://raw.githubusercontent.com/PrefectHQ/prefect/8f047c87cf30278eeaf66d3769c1bc5c52f81515/LICENSE.md) and 90 under [server-scoped conditions](https://raw.githubusercontent.com/PrefectHQ/prefect/cade26c01b363d42b1e3833a2a1044733acddcef/LICENSE) pending. Missing document candidates affect another twelve Prefect and two Dagster snapshots. Keyword hits are not license decisions.

There are 3,172 catalog-backed conditional local numeric candidates. Joining successful costs/features yields **3,048 training and 107 validation** rows. These are not public-release approvals and do not authorize code/text generation or raw redistribution. The [next FP32 plan](../path-cost-claim/README.en.md) requires at least 2,400 rows in each role; validation falls short, so corpus fitting has not started.

The subsequent [second fair acquisition batch](collection-batch2-44.json) terminated normally after 2,000 requests. Aggregates, private inventories and failures reproduce in an offline replay. Catalog coverage reaches **4,378: 3,614 training / 764 validation**. The [third batch](collection-batch3-44.json) also terminates after 2,000 requests and reproduces offline, reaching **5,369: 3,971 training / 1,398 validation**.

These are acquired catalogs, not newly source/target/cost-qualified fitting rows. Do not overwrite experiment45's costs or source assessment with later acquisition. The validation fitting gate remains unmet.

## Replaying and implementation

Override plan file, plan hash, checkpoint file and checkpoint hash together. Reject partial overrides; default invocation preserves experiment43.

```sh
go run ./cmd/riido-pathcostdata \
  --out .cache/path-cost-data-45-new \
  --plan experiments/path-cost-data/plan-45.json \
  --plan-sha256 ed9e766d3c456573cc42df96566a69aed3e544f06cf4884653ce056948dd7617 \
  --checkpoint .cache/development-join-45/evidence.json \
  --checkpoint-sha256 382ce75674e32a2e0adb7e8908d6cab92ff7e768f4f4ed63823952c03c957766
```

Pinned private inputs must already exist. Verify plan source hashes, schema and production=false; record actual digests in reports. Even a root-only unavailable-catalog row must verify its root digest without substituting a newly acquired catalog. Preserve streaming queries, 16 features, page20 and final protection.

Replay target joins, new costs and source assessment offline. Public costs and private numeric examples reproduce byte-for-byte in ordinary replay and race execution. Synthetic tests cover override combinations, wrong digests/plans, missing/changed roots and catalog non-access. Acquisition runs concurrently, so do not report isolated latency/RSS from these runs.

Add the pinned upstream dataset reference to the Hugging Face [research collection](https://huggingface.co/collections/JooYoon/riidolaya-public-research-6abcbd5ddb1917912fc5de38). No new model weights are released. Next: expand validation, review new sources and prepare the ten-candidate FP32 runner; separately compare ternary variants after useful signal.
