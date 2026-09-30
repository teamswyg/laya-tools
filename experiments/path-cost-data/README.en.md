# Experiment 43: per-task search cost data

Follow-up: [expanded checkpoint costs/source readiness](EXPANDED-45.en.md) and the [FP32 plan requiring at least 2,400 tasks each for training and validation](../path-cost-claim/README.en.md). Results below belong to the frozen experiment43 checkpoint and are not replaced by later acquisition.

**Keep all 13,021 development tasks in the denominator and execute real search on the fixed checkpoint's 2,248 catalogs. On 2,236 scoreable tasks, total first-target pages decrease from 26,226 to 24,781, but 302 tasks get worse versus 143 that improve. Only five validation tasks are scoreable, so this cannot establish model quality or select a model.**

## Purpose

A tiny claim model needs both helpful and harmful examples to learn when auxiliary search might be useful. This experiment prepares that numeric cost data in Go without fitting a model. It does not run Laya inference, ternary training or paid LLM calls.

Commit the [plan](plan-43.json) first. The [results](results-43.json) contain public aggregates. Per-task numbers and bookkeeping identities remain in private caches. Result examples contain no raw queries, target paths, patches or model weights.

## Denominators and separation

| Role | Fixed tasks | Available catalogs | Scoreable |
|---|---:|---:|---:|
| Training | 7,335 | 2,243 | 2,231 |
| Validation | 5,686 | 5 | 5 |
| Total | 13,021 | 2,248 | 2,236 |

The checkpoint lacks catalogs for 10,773 tasks. Twelve available tasks lack old-file targets and receive no cost label. The 2,402 final tasks never reach search, feature extraction or scoring. The source projection includes other roles and duplicate siblings: their bytes and JSON are read for integrity, but only development representatives reach the search callback.

Status gives checkpoint unavailability precedence. Available tasks distinguish ranking errors, target parse failures, unsupported targets, no old targets, unmapped targets and scored outcomes. `ParseFailures=0` therefore does not mean the full source has no parse failures. The 294 development failures in [experiment 39](../development-labels/README.en.md) are all unavailable at this checkpoint. Neither missing data nor failures become negative training labels.

Pin the SHA-256 of experiment 40's second collection checkpoint. Concurrently acquired catalogs outside that checkpoint stay unavailable for this run. A required cache missing or changing digest causes an error, without network replacement. This reproduces partial acquisition; it does not shrink the full cohort.

## Paired comparison

Use the query and complete file candidates for baseline search, then capture the fixed 16 features and baseline order before passing gold to scoring. Execute auxiliary search for every available catalog and interleave its order with baseline first. Gold is used only after both rankings exist. Labels are parsed earlier, but never enter the ranker or feature function.

A page contains 20 candidates. Moving the first target from rank 21 to 20 changes two pages to one, yielding +1; the reverse yields -1. Assign costs only when all old-file targets map to candidates. `Hit1/10/100` counts tasks with at least one target within that rank; `All10` counts tasks with every target in the top ten. Divide by the corresponding role/repository's `Scored` for rates, and report scoreable coverage against all tasks alongside them.

| Measure | Baseline | Always auxiliary |
|---|---:|---:|
| Total first-target pages | 26,226 | 24,781 |
| Hit1 | 201 | 201 |
| Hit10 | 796 | 848 |
| Hit100 | 1,607 | 1,655 |
| All10 | 442 | 459 |

There are 143 improvements, 302 regressions and 1,791 ties. Actual baseline rank, auxiliary index build and auxiliary rank attempts each total 2,248, with no fallback or failure in this run. The twelve tasks without old targets still execute search before target inspection, so execution counts exceed scored counts.

Qiskit contributes 1,307 tasks and Airflow 444; a few repositories dominate. DataDog worsens from 580 to 646 pages and conda from 436 to 443. Validation worsens from 28 to 29 pages, but five tasks cannot support generalization. Do not enable the policy by default from the aggregate reduction.

## Implementation and verification

From the repository root, run `go run ./cmd/riido-pathcostdata --out .cache/path-cost-data-43`. The pinned private caches from experiments 36, 37, 39 and 40 must already exist; the output directory must be new. Missing inputs are not automatically downloaded.

Verify the entire query projection hash before streaming one row at a time. Enforce 96MiB file and 1MiB line limits, 19,008 source rows, development roles, repository/commit identities, duplicates and missing members; verify the second-pass digest too. Keep callback outputs unpublished until every check succeeds. Reuse the already computed baseline order instead of ranking it twice. Model inputs contain numeric features, not path names, repository names or gold identities.

Owned synthetic tests cover role leakage, file mutation, duplicates/missing members, target mismatches, page-boundary gains/losses, corrupt/missing caches and full denominators. Compare baseline/auxiliary rankings with the existing independent reference. Verify actual public reports and private examples are byte-identical across ordinary replay and race execution. Require full race tests, vet, formatting, secret scanning and CI before publication.

Acquisition runs concurrently, so do not report these runs as isolated speed or peak-memory benchmarks. Candidate pages do not measure LLM tokens, money or completed coding tasks. This result does not approve source eligibility, fitting new weights, final evaluation, model release or production activation.

## Next work

Address lagging validation-repository acquisition first. Keep membership and final protection fixed; change collection scheduling using availability rather than search outcomes. Once validation coverage and source review are sufficient, separately precommit fitting and threshold-selection rules. File-search evidence does not validate model up/down routing, repository selection or task decomposition.
