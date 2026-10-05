# JWT policy search costs: compare caching and simple ordering first

Before adding a tiny model, we measured how much repeated work caching and a simple candidate order could remove. In this experiment, **reusing signature verification and decoded claims made fixed-order search faster.** When the first candidate already matched the request, nothing was reused and caching added cost. There were no model calls or new training runs.

[한국어](README.md) · [Full first cost response](OBSERVATIONS.actual.public.v1.json.gz) · [Analysis](ANALYSIS.actual.public.v1.json) · [Execution and resources](EXECUTION.actual.public.v1.json) · [Original JWT observations](../README.en.md)

## What is reused?

A JWT is a signed token. Claims are its contents, such as audience and expiration, not-before and issued-at times. We separate verifying the signature from checking whether those contents satisfy the current policy. For example, comparing “allow alpha or beta” with “require both alpha and beta” can reuse signature verification, but each policy must still be checked.

The cache is a fixed five-entry array, cleared for every request, control and round. An entry is reusable only when the entire token, entire public-key DER, signing method, decoder/claims profile and pinned source revision match. On first use, an actual `ParseWithClaims` verifies the signature and decodes the token before storing its claims. Audience slices and all three NumericDate objects are copied both when storing and before reuse. Every candidate gets a fresh `Validator` and the fixed clock. Final `Token.Valid` values and policy acceptance/rejection results are never cached.

Unexpected errors, input differences, cache mutation, panics and deadline expiration remain **unknown**. They are not converted into policy rejection or evidence that a candidate is wrong; the operational search stops. This shortcut is scoped to these fixed public tokens, `RegisteredClaims` and the selected policy options. It does not prove equivalent reuse for other claims implementations or policies.

## What happened in the first execution?

We made one separate cost-worker execution. First, it paired the uncached and cached paths for each of the original 160 observations. **All 160 pairs matched in acceptance, normalized error classes and canonical claim fingerprints, and also matched the original observations.** This stage made 160 reference `ParseWithClaims` calls, 20 cache-authentication calls and 160 fresh policy `Validate` calls.

It then ran four requests through the following four controls for three rounds. `manual-first` uses first candidates `[4, 5, 2, 0]`, declared from the request contracts before observing outcomes. It is neither model prediction nor a learned order. A candidate stops at its first fixture mismatch; search stops at the first candidate whose acceptance matches Wanted on all five fixtures.

| Control | Total time for 12 request records across three rounds |
| --- | ---: |
| Fixed source order, uncached | 5.141583 ms |
| Fixed source order, cached | 2.711623 ms |
| Predeclared candidate first, uncached | 2.428458 ms |
| Predeclared candidate first, cached | 2.593877 ms |

These are **four requests × three rounds per control**, not single-request latencies. They include lookup, deep copies, option setup, order construction, fingerprint checks and remaining request work. All 48 searches finished at the same matching candidates `[4, 5, 2, 0]`, with no unknown results.

Across four requests in each round, fixed order performs 43 fixture checks. Without caching, these require 43 `ParseWithClaims` calls. With caching, they require 20 authentication parses and 43 fresh policy validations, with 23 cache hits. The predeclared-first control finishes at its first candidate: both modes require 20 parses and **have zero cache hits**. Caching adds lookup, storage, copies and a separate policy check in that case.

## Is there enough room for a model?

With caching already enabled, changing only the order produced different time differences across rounds. All values below are totals across four requests. Positive means the predeclared order was faster; negative means it was slower.

| Round | Fixed cached − predeclared-first cached | Fixed cached − predeclared-first uncached |
| --- | ---: | ---: |
| 0 | +61,497 ns | +107,457 ns |
| 1 | −44,209 ns | +66,125 ns |
| 2 | +100,458 ns | +109,583 ns |

The last column compares against an ideal control that checks the matching candidate first and skips caching on these inputs. **It is not a guaranteed inference budget.** A model would need to cover feature extraction, inference and order construction while producing a greater practical benefit, and its ordering accuracy would need separate validation. For now, this supports adding inexpensive nonlearned retrieval controls first.

All operational rounds followed the 160-pair check in the same process, after the code and runtime had already been used. The middle round reversed control scheduling, but this is not randomization or independent replication. Thirteen distinct fictional tokens are reused across 20 fixture positions in four requests. This is an exposed `development_validation` diagnostic for the whole JWT lineage, not 48 new independent tasks or a 2,400-task evaluation. There were no corpus admissions, model activation, training, GPU calls or paid calls.

## Timing and memory scope

The execution used an Apple M4 Pro, macOS arm64 and Go 1.27.1. `GOMAXPROCS=1` permits one Go execution slot; it does not enforce a single OS thread.

| Measurement | Observation | Included scope |
| --- | ---: | --- |
| Complete per-request work | Four control totals above | Setup, lookup, copies, checks and result qualification inside each request |
| `main` elapsed time | 23.256416 ms | Input setup, 160-pair comparison and 48 searches; excludes import initialization and output marshaling/writing |
| Whole-child lifecycle wall time | 608.614250 ms | Process start through pipe/wait cleanup; excludes persistence of the raw capture |
| Whole-child CPU time | user 29.161 ms + system 8.955 ms | OS lifetime totals including initialization, work and output |
| Whole-child peak RSS | 14,532,608 B, about 13.86 MiB | OS maximum for the process executing every control |

CPU and RSS were not measured separately for each cache mode or model. RSS is not Go heap usage or cache size. The difference between the two elapsed times does not isolate pure startup cost. `GOMEMLIMIT=64MiB` is a soft Go memory target; the 256 MiB RSS criterion is checked after execution. Internal signature-verification call counts were not observed. We compare controls within this execution rather than adding the previous observer's timings into a hypothetical saving.

## Inspect the saved evidence

From the repository root, replay the saved first response to check result equivalence, search stopping, counters and cost totals. This does not execute the original JWT API, a model or training.

```sh
go run ./experiments/short-claim/next-cohort-jwt-audit/cost-replay
```

Only if you want a new native execution, run the separate module below. It requires **Go 1.27.1 selected on macOS arm64**. The upstream source is supplied locally in `jwt-source`, so no module or model download is required.

```sh
cd experiments/short-claim/next-cohort-jwt-audit/cost-preview
GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOTELEMETRY=off \
  CGO_ENABLED=0 GOMAXPROCS=1 GOMEMLIMIT=64MiB \
  go run . < ../OBSERVER-INPUT.actual.public.v1.json
```

This command compiles and makes a **new API execution**. It is separate from saved-evidence replay and does not automatically reproduce the limits or CPU/RSS capture supplied by the first execution's separate process runner. The `author-only` wording in [PROTOCOL](PROTOCOL.md) and source comments preserves their frozen preparation state before execution. Read actual execution status and measurements from [the execution record](EXECUTION.actual.public.v1.json) and [the first response](OBSERVATIONS.actual.public.v1.json.gz). This document does not claim CI passage or completed external publication for this cost packet.

## Sources and licenses

The upstream is [this pinned golang-jwt/jwt commit](https://github.com/golang-jwt/jwt/tree/73c870b18e68b6e654b2b03f485aa3c9fab32cea). All 25 copied files and the [full MIT license and original copyright notice](jwt-source/LICENSE) are preserved. Our code, documentation and fictional inputs are Apache-2.0. See [source pins](SOURCE-PINS.public.v1.json) and [fixed Go SDK notices](../TOOLCHAIN-NOTICES.v1.json). This does not redistribute the complete SDK or certify rights across the whole lineage or for model training.

The public key and tokens are fictional test material, not real user credentials. The public first stdout preserves all 125,979 bytes unchanged inside gzip. The full process capture containing private host paths remains private; its scope is described in the public execution record.
