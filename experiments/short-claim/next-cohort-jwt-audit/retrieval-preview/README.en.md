# Preparing a JWT candidate retrieval cost comparison

A small claim model helps only when the work it saves exceeds the cost of producing its hints. This preview first asks whether **cheap word retrieval can improve candidate order**. It proposes an order; the existing JWT verifier makes the final judgment.

[한국어](README.md) · [Previous cache experiment](../cost-preview/README.en.md) · [Protocol frozen before execution](PROTOCOL.md) · [Own-code provenance](SOURCE-DERIVATION.prospective.v1.json)

This document describes preparation. The new 160-pair comparison and 128 cost operations have not run. No new speed or memory improvement is claimed. There are no Laya, other model, training, or GPU calls.

## Comparisons

| Candidate order | Construction | Purpose |
| --- | --- | --- |
| Fixed | Original declaration order of eight candidates | Baseline cost |
| Predeclared first candidate | `[4, 5, 2, 0]` from the exposed request contracts | Ideal control that already knows a matching candidate |
| BM25 | Query/caption word frequency and document length | Retrieval cost and effect |
| Jaccard | Unique word intersection divided by union | Simpler word overlap comparison |

Each order runs with and without a cache. Four exposed requests run for four rounds. Control positions rotate by round, and cache-mode order reverses on odd rounds. This gives `4 controls × 2 cache modes × 4 requests × 4 rounds = 128` cost operations. Repeated development inputs are not 128 new independent samples.

First, the same 160 judgments compare cached and uncached acceptance, error classes, and claims fingerprints. Current reference results qualify operational parity; they never enter retrieval scoring.

## Hint boundaries

Retrieval reads only raw request text and candidate captions. It receives no expected answers, policy options, previous outcomes, or timings. Zero-score and tied candidates remain in the complete order; ties preserve declaration order. The verifier checks actual inputs and stops a candidate at its first mismatch with expected acceptance. Unexpected errors or altered results remain `unknown` and stop the batch.

For “allow alpha or beta,” the hint suggests checking a similarly described policy first. Word overlap does not approve that policy as correct. Future small claim models must preserve this boundary too.

## Costs

Each request repeats text normalization, token preparation, scoring, and sorting inside its complete-request timer. No prebuilt retrieval catalog hides preparation cost. Each request starts an empty five-slot fixed-array authentication cache. Lookup, deep copying, fresh policy validation, fingerprints, and parity qualification are included.

CPU, peak RSS, and lifetime are recorded for the complete process separately from phase timings and counters. Whole-process measurements cannot attribute CPU or memory to one retrieval control. A Go memory target does not limit total OS or GPU memory.

## CI and local execution

The local 512MiB retained-storage cap applies to declared experiment work/evidence paths. It is not a whole-machine disk or memory cap. Retaining prior evidence leaves insufficient room for a fresh compiler cache, so the planned path tests/builds in CI, verifies the produced file, and measures actual operations locally.

GitHub documents free standard runners for public repositories and arm64 `macos-15`. Actual jobs still verify the Go version, host architecture, and source hashes. [GitHub runner documentation](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)

CI uses synthetic inputs to check arithmetic, array ownership, unknown handling, and complete-output sizing. It does not invoke the worker main or run the original 160/128 verification operations. Actual execution is separate, with the complete first response, failure, and status retained before interpretation.

Developers choosing a fresh execution can build this standalone module on macOS arm64 with Go 1.27.1 and enough build storage.

```sh
cd experiments/short-claim/next-cohort-jwt-audit/retrieval-preview
GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off \
  CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags='-s -w' \
  -o retrieval-preview .
GOMAXPROCS=1 GOMEMLIMIT=64MiB ./retrieval-preview \
  < ../OBSERVER-INPUT.actual.public.v1.json
```

This performs new original verification. It is distinct from preparation, source qualification, or replaying saved evidence, and does not automatically reproduce the project's capture tool, time limits, or CPU/RSS records. No Codex or riido registration is required. People can read this guide and summaries; agents can inspect the same JSON records and source fingerprints.

## Scope and licenses

Only four already-exposed requests from one lineage are used. Independent 2,400-case evaluation, selection of other repositories, actual Agent token/cost savings, and trained-model generalization remain unproved. Results will inform how much preparation/inference cost a future model experiment can afford.

Own code and documentation are Apache-2.0 with existing Laya Tools attribution retained. The 25 copied JWT files preserve the MIT license and original notices from the [pinned upstream](https://github.com/golang-jwt/jwt/tree/73c870b18e68b6e654b2b03f485aa3c9fab32cea). Build materials provide Go SDK BSD and patent notices. This program depends on the Go standard library, pinned JWT, and own retrieval code. It includes no model weights, ORT, or Laya encoder.
