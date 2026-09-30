# Experiment 44: acquire validation data alongside training data

**Add an optional acquisition order that does not wait for every training repository before visiting validation. A 300-request batch adds 146 catalogs, including 102 validation tasks. Validation availability rises from five to 107, still far below its 5,686 tasks. This is not model fitting or an accuracy improvement.**

## Motivation

[Experiment 43](../path-cost-data/README.en.md) prepares search costs but has only five scoreable validation tasks. Increasing the global catalog count alone does not provide independent validation. Existing role/repository/ID order spends the request budget on earlier training repositories.

Commit the [plan](plan-44.json) before implementation and new requests. The optional `--order fair` alternates training and validation, rotating lexicographically sorted repositories within each role and yielding one ID-sorted task from each. Skip exhausted repositories; drain the remaining role when one ends. The scheduler receives no targets, search scores or query text. Its deterministic partial prefixes are not random or guaranteed representative samples.

## Preserved boundaries

Keep all 7,335 training and 5,686 validation members, totaling 13,021, and exclude the 2,402 final tasks from collection. The [actual membership probe](schedule-44.json) verifies no lost/duplicate members and all 25 development repositories within the first 40 visits. Reuse existing caches and preserve default `serial` order and `sample` mode. The new order is valid only in `all` mode and requires the pinned amendment hash.

Preserve request starts at least 300ms apart, 30-second request timeouts, at most 2,000 new requests per invocation, existing file/path/tree bounds and license-evidence scope. Do not traverse symlinks or submodule contents. This adds neither cohort selection nor training approval.

## Actual acquisition

Start from the previous serial batch's terminal [3,242-catalog checkpoint](collection-resume3-38.json). Check 491 remaining API requests, then limit this batch to 300.

| Role | All tasks | Previous catalogs | New catalogs | Added |
|---|---:|---:|---:|---:|
| Training | 7,335 | 3,237 | 3,281 | 44 |
| Validation | 5,686 | 5 | 107 | 102 |
| Total | 13,021 | 3,242 | 3,388 | 146 |

Keep all 9,633 unavailable catalogs in the denominator of [results](results-44.json). Having 107 validation catalogs does not establish that all have scoreable targets or eligible training sources. Historical license-candidate availability is not automatic permission. Experiment 43 retains its pinned 2,248-catalog checkpoint and unchanged results.

## Running and verifying

From the repository root with the existing private membership and caches:

```sh
go run ./cmd/riido-trainingcatalog --mode all --order fair --request-budget 300 --out .cache/catalog-fair-new
```

Use `--offline` and a new output directory to replay without network. An additional `schedule.json` records order version, amendment hash and membership hash. Public aggregate format stays unchanged. Private failure evidence follows visitation order; sort by identity and failure stage when comparing different schedules.

Owned synthetic tests cover exact order, uneven repository sizes, role exhaustion, input-order invariance, input immutability and rejection of duplicate IDs, final roles and repositories crossing roles. Before new requests, old/new ordering over identical caches produces byte-identical public aggregates and license inventory plus identical normalized failure evidence. The new 300-request batch reproduces report, license inventory, failure evidence and scheduling metadata byte-for-byte offline. Run full race tests, vet, formatting and redacted secret scans.

Do not report acquisition speed or memory as model inference performance. Next, increase validation coverage and verify target joins and source conditions for new data. Do not repeatedly select models on this partial evidence without adequate validation and a separate fitting plan.
