# Short-claim experiment56b: behavioral truth and connected-group audit

[한국어](RESULTS-56b.ko.md) · [Pre-observation plan](PLAN-56b.en.md) · [Usage](USAGE-56b.en.md) · [Numeric results](results-56b.json) · [Next experiment56c](NEXT-56c.en.md)

**The audit of72 authored requests found17 total groups and16 labeled groups. Both exceed the preparation minimum15, but training is not ready.** There are34 answerable,17 no-answer and21 unknown requests. Role splits, fitting, new weights, model calls, ranking, performance runs and final evaluation are all0 in this stage. Development data from one synthetic authoring pipeline do not establish general semantic understanding or actual LLM savings.

## What changed

Beyond [56a](RESULTS-56a.en.md)'s integer-array checks, the maintainer tool `riido-typedaudit` observes error identity, state commits, memory ownership, cancellation/cleanup order, quotes/escapes and graphs. For example, identical initial values can still violate snapshot ownership if later mutation of the original changes the snapshot. Equal error messages do not establish the same underlying cause.

This auditor is separate from the [small hint tool](USAGE-56.en.md) that suggests verification order. It compares concrete Go behavior with independently authored literal expectations and groups related data together. It uses no large encoder, Laya inference, Python or paid service.

## Scope frozen before observation

Inputs, source and the execution plan were sealed first in [commit c27712c](https://github.com/teamswyg/laya-tools/commit/c27712c18e1f79e8a373b9db1aa2564c4918f06a). Unit control checks had already run during preparation; **one official whole-cohort truth/group observation was collected after freezing**. Subsequent CI replay checks reproducibility and adds no samples. Preserve the [preparation plan](PLAN-56b.en.md)'s pre-observation wording and the original56a record as historical snapshots.

| Frozen artifact | SHA-256 |
|---|---|
| [Execution plan](execution-plan-56b.json) | `2fe7aeccd72198fda67f2b0ad9e067a2092b465bef252ba72fb464662cf40a82` |
| [Original24 new requests](probes-56b.json) | `0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e` |
| [Original48 legacy requests](probes-56.json) | `0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0` |
| [Official results](results-56b.json),266,818 bytes | `4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4` |

The plan pins23 implementation/test files and both inputs. The executor also checks filesystem source against compiled source and decodes exactly the verified input bytes. It preserves the original56a truth record exactly and records the stricter combined grouping separately.

## Data and truth counts

| Dataset | Original requests | Candidate captions | Answerable | No answer | Unknown |
|---|---:|---:|---:|---:|---:|
| Legacy56a | 48 | 144 | 24 | 12 | 12 |
| New56b | 24 | 72 | 10 | 5 | 9 |
| Combined | **72** | **216** | **34** | **17** | **21** |

The51 labeled requests are **34 answerable +17 no-answer**. JSON's `known_parents` includes no-answer requests; legacy36 and new15 are not answerable counts. No-answer means no candidate is acceptable within the finite contract. Unknown means the request is ambiguous or its short caption cannot faithfully express the full behavior. All21 unknown requests remain in the original denominator72 and relation audit.

The new code controls comprise6 correct functions and18 wrong functions. Candidate outputs did not generate the expectation tables. The6 correct functions had0 violations on their specified vectors; each of the18 wrong functions had at least1 observed violation.

| New behavioral family | Independent literal vectors | Recorded assertions across4 sources per family |
|---|---:|---:|
| Atomic state commit | 18 | 72 |
| Error identity | 11 | 44 |
| Owned snapshot | 6 | 24 |
| Cancellation/resource lifetime | 12 | 48 |
| Quoted/escaped delimiters | 23 | 92 |
| Cycles/shared DAGs | 10 | 40 |
| Total | **80** | **320** |

New320 + legacy204 = **524 recorded source assertions**. This is neither a request count, an independent training-sample count nor the total number of actual function calls. Observer support and post-return alias checks may make additional calls. Passing finite tables does not prove function correctness for all inputs.

Separate source execution validation from independent reading of natural-language fidelity. Each request/candidate has at most512 raw bytes,512 normalized bytes and32 normalized words; content is never truncated. Five families passed development-caption review, but **all4 delimiter requests remain unknown:3 incomplete captions and1 ambiguous policy**. Passing code controls does not mean32 words conveyed every ASCII, error-kind, capacity and partial-output exception. Translating documentation creates no new evaluation requests.

## Why passing group minima does not start training

| Separate preparation gate | Observed | Frozen minimum | Result |
|---|---:|---:|---|
| Total connected groups | 17 | 15 | Pass |
| Labeled connected groups | 16 | 15 | Pass |

The18 prototypes form17 groups. Transitively connect shared prototypes, behavioral cores, request/candidate captions, negative sources, copied functions and authored helpers/types/globals. Unknown-only groups remain in total grouping but do not satisfy the labeled minimum. Standard-library and observer support are excluded only from group connections under the predeclared policy, while remaining in provenance. No groups were split and no minima were lowered.

These counts do not establish statistical independence, diverse authorship or training eligibility. Data are English requests from one coordinated synthetic pipeline; `no_roles_plan` and `synthetic_single_pipeline` stop training preparation. Training/development/final roles and partitions0, fits0, new weights0, model calls0, ranking0, performance runs0 and activation0 remain unchanged. The existing paid coding ledger also remains20 records covering7 distinct requests.

## A packaged-executable problem found during validation

The actual `-trimpath` executable failed with `source_typecheck_failed`. Relying on GOROOT information supplied by the test runner could leave a packaged executable unable to locate standard-library type metadata. Offline provenance checking now obtains export metadata from installed Go or an already cached **exact Go1.27.1**. Downloads are disabled and global `build.Default` is unchanged. The CGO0/trimpath packaged preparation smoke check passed with the GOROOT environment hint removed.

Local full race, vet, CGO0 package checks, publication/PDCA guards, formatting, diff checks and the redacted full Git secret scan passed. **CI has not run at this document's preparation time**; no CI success status or link is claimed. Later CI will record a gate that reproduces the same frozen results.

Preparation is CPU-only, using Go1.27.1, GOMAXPROCS=1 and a256MiB Go-heap soft limit. These are not measured OS thread counts, a total-RSS hard cap, GPU usage or observed latency. Offline `go/types` metadata costs are not the resident costs of the small fixed-array hint tool. There are no new performance, caching or concurrency observations.

## Reproduction and the next experiment

Run from the repository root containing the frozen sources with installed or cached Go1.27.1. Choose an output directory that does not exist. This replays a finite audit; it neither fits nor calls a model.

```sh
mkdir -p .cache
GOTOOLCHAIN=go1.27.1 GOPROXY=off GOSUMDB=off CGO_ENABLED=0 go build -trimpath -buildvcs=false -o .cache/riido-typedaudit-56b ./cmd/riido-typedaudit
./.cache/riido-typedaudit-56b --stage audit --plan experiments/short-claim/execution-plan-56b.json --plan-sha256 2fe7aeccd72198fda67f2b0ad9e067a2092b465bef252ba72fb464662cf40a82 --input experiments/short-claim/probes-56b.json --legacy experiments/short-claim/probes-56.json --out .cache/my-typed-audit-56b
```

The next hypothesis asks about **one property with a fixed scope** instead of an entire function. For example, check only whether a valid graph's shared child is rejected as a cycle; do not promote this into whole-function correctness. First freeze a new input contract, literal truth, unknown conditions, caption fidelity and source relations, then prepare diverse public provenance plus separate nonlearned-utility and role-split plans. Preserve parent/translation/counterexample/sibling relationships and INT8/PTQ child relationships from the same fit. Future1.58-bit comparisons still have0 fits now.

Preserve the separate goal of **at least2,400 distinct final-evaluation requests per domain**, apart from training/development data. Final eligibility, protected-final/CoSQA access and scoring are0 here. There are no new weights to publish to HF; only maintenance of existing collection references is planned. Keep model bodies, personal paths, real work inputs, credentials and raw profiles out of Git. Resident streaming, load, total CPU/RSS and cache invalidation belong to a separate resource stage.
