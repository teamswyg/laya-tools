# Original JWT observations and the next cost comparison

We need actual program outcomes and costs before deciding whether a router helps. We prepared 13 fictional signed tokens and applied eight candidate policies to five inputs for each of four requests. The original JWT library's `ParseWithClaims` **actually ran 160 times**. We preserved the first complete response, failure state and process resources. Model executions and new Fits remain zero.

[한국어](README.md) · [Full first API response](OBSERVATIONS.actual.public.v1.json.gz) · [Execution record](EXECUTION.actual.public.v1.json) · [Unchanged requests and expectations](INPUTS.reviewed.v2.json)

## What was observed

| Fictional request | Requested policy | Expected acceptance of five inputs | Candidate matching all five inputs |
| --- | --- | --- | --- |
| Lantern | Addressed to alpha or beta | T T T F F | WithAudience |
| Harbor | Addressed to both alpha and beta | F F T F F | WithAllAudiences |
| Moss | Expiration must be present | F T F F T | WithExpirationRequired |
| Cove | Two seconds of time tolerance | T F T F T | WithLeeway |

Here T means accepting an input, which differs from a candidate matching the request. The API accepted 100 tokens and rejected 60. Across 32 request/candidate combinations there were **four matches, 28 mismatches and zero unknowns**. Comparisons require completed signature/decoding observations and intact input guards. We did not change the original expectations.

These are finite five-input matches. They do not prove every time/issuer/subject boundary, all JWT implementations or real-work correctness. We authored the examples after reading public source, so the entire JWT lineage belongs to `development_validation`. This is neither an independent holdout nor 2,400 independent tasks, and no rows are admitted directly to training. We preserve the preparation snapshot's null role; a separate [family/control freeze](FAMILY-CONTROLS.freeze.v1.json) was adopted before outcomes.

## Memory and time

This was the first run on Apple M4 Pro, Go 1.27.1, with one CPU thread. Original observation peak OS RSS was **12,517,376B ≈ 11.94MiB**, with **23.689ms** lifetime child CPU. Time inside `main`, including input qualification, 160 trials and result preparation, was **9.607ms**. Start through process/pipe cleanup took **598.010ms**. Startup appears material for this small workload. These observations cover different scopes; their difference is not an isolated startup measurement.

RSS is neither Go heap nor model memory. This is Go CPU execution and measures no Laya inference, GPU use or agent token savings. `GOMEMLIMIT=64MiB` is a soft Go memory target; the 256MiB RSS gate is retrospective. Neither is a hard RSS guarantee. No profiles or dumps were taken while the private key was live.

## Why training does not follow immediately

Better ordering can reduce checks. But safely reusing authenticated and decoded claims within a request may leave only cheap policy validation. Even a tiny model is unhelpful when it costs more than that remaining work.

The next controls are fixed order, an order authored from request wording before observations, and simple retrieval. We will separate repeated authentication/decoding from exact token/key/algorithm/decoder/source scoped claims reuse. The cache must never store final `Token.Valid` or policy decisions. Audience slices and timestamp objects are copied, and every candidate policy is freshly validated. Actual cost must include cache lookup, copy, ranking and initial setup. Summing saved times from these 160 observations cannot establish an operational speedup.

## Check the saved response

Compare the preserved first API response against unchanged expectations in Go. This command starts no original API, model or training.

```sh
go run ./experiments/short-claim/next-cohort-jwt-audit/replay
```

`observer` is a separate Go module using the exact 25 upstream files in `testdata/jwt-source`. No network or model download is required. Running it again is a new original observation, distinct from reading saved evidence. `runner` is a separate Mac arm64 module for first child capture and CPU/RSS measurement. Their owned tests exercise boundaries and error handling, and cannot establish every actual OS lifecycle behavior.

## Sources and licenses

The upstream source is [golang-jwt/jwt at the fixed commit](https://github.com/golang-jwt/jwt/tree/73c870b18e68b6e654b2b03f485aa3c9fab32cea). Copied files retain the [complete MIT license and original copyright notices](observer/testdata/jwt-source/LICENSE). Our observer source, fictional inputs and documentation use Apache-2.0. Selected installed Go SDK fingerprints and relevant notices were reviewed; we do not claim to archive or redistribute the complete SDK. Model training and distribution rights require separate review.

The public key and tokens are fictional fixtures generated for this experiment. No real user tokens, private code, credentials or model weights are included. The first private captures contain host paths and remain private; the public API response is a separate byte-for-byte copy. Preparation-helper failures and original execution are distinguished in the [execution record](EXECUTION.actual.public.v1.json).
