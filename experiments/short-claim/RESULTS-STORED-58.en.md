# Ordering headroom measured from stored truth

[한국어](RESULTS-STORED-58.ko.md) · [Pre-observation plan](PLAN-STORED-58.en.md) · [Usage](USAGE-STORED-58.en.md) · [Original record](results-58.json) · [Next preparation](NEXT-STORED-58.en.md)

**Better ordering still has room to help. The best fixed baseline required 91 checks; an answer-knowing oracle required 73. The gap is 18 checks, about 19.8%.** This supports further preparation of a tiny hint model. It is neither an improvement achieved by a model nor measured Codex savings. No model or training ran in this phase.

If only the second of three candidates satisfies a request, the original order needs two checks. A hint placing that candidate first could need one. A request with no acceptable candidate still requires every check. Preserve that cost so hints cannot delete candidates or replace independent verification.

## Comparison

Reuse all 72 existing public development requests and 216 candidate captions: 34 answerable, 17 no-answer and 21 unknown. Unknowns retain their original reasons, relations and full denominator, without becoming incorrect answers or entering cost/Top metrics. Cost covers 51 requests with known outcomes.

| Fixed control | Checks for 51 known requests | Checks for 34 answerable requests | Top1 correct | Top3 correct | Rule fallback / all 72 |
|---|---:|---:|---:|---:|---:|
| Original order | 109 | 70 | 6 / 34 | 34 / 34 | 0 / 72 |
| BM25 | **91** | **52** | **20 / 34** | 34 / 34 | 0 / 72 |
| Lexical/order features | 92 | 53 | 19 / 34 | 34 / 34 | 0 / 72 |
| Narrow rule | 91 | 52 | 20 / 34 | 34 / 34 | **72 / 72** |
| Answer-knowing oracle | **73** | **34** | Not compared | Not compared | Not applicable |

The narrow rule falls back to BM25 on every request, so this is no separate rule improvement. Its tie with BM25 resolves to BM25 using the frozen control-array order. Never select a favorable control per request to construct the best baseline.

With only two to four candidates per request, Top3 easily saturates. The shared 34/34 result is not broad semantic-understanding evidence. Answerable-only headroom is 18/52, about 34.6%; the predeclared gate uses no-answer-inclusive **18/91 = 19.7802%**. The exact integer test `100*18 >= 5*91` passes the necessary 5% floor.

Existing 17 connected groups and 16 labeled groups already pass the operating floor of 15. The mean of per-group mean known checks across 16 labeled groups is about 1.760417 for BM25 and 1.437500 for oracle. This additional diagnostic does not replace the global gate. Preserve every group result in the original record.

## Retained execution and preparation failures

**One official attempt, zero retries** completed every request. Distinguish 72 dispatches, 72 validated inputs, 72 Baselines calls, 288 completed control rankings and 72 completed rows. The 288 rankings are not independent requests. Original candidate/source API executions, model/paid calls, fits, new weights, assigned roles and protected-final reads are all zero.

The [collection ledger](collection-ledger-58.json) separates official execution from preparation attempts. Local full race tests, vet, Go formatting and maintainer publication/PDCA guards passed. Verify remote CI outcomes separately afterward.

Three plan-generation attempts preceded official ranking. The first stopped with `storedutility_noncanonical_json` because the build recipe used different JSON escaping. Fix the recipe and add a test of the actual checked-in file. The second mistakenly supplied a nonexistent commit argument and stopped with `storedutility_git_pin_mismatch`. The third succeeded with the actual source commit. **Both preparation failures occurred before ranking and remain recorded; neither was hidden or relabeled as an official retry.**

| Frozen item | Record |
|---|---|
| Source freeze | `3c2bde94db7661bf5cdb268f0a2ede0f0b962a94` |
| Input freeze | `bb2e6d2d137ce0626ccbfba1ee591a9814ec8153` |
| [Execution plan](execution-plan-58.json), 5,402 bytes | SHA256 `c85c048831c6a3807656fccdba967251574b88b5764cbf708c6b38575d8e1893` |
| [Official result](results-58.json), 184,532 bytes | SHA256 `48bbb4dd10ee5cf34f4edb7e746f0a0f7c85afd53b6513b62ea8442dac2768e6` |
| Actual executable, 4,872,258 bytes | SHA256 `fb7153b55b9351fcfb69ef881e1fb49a2ab69fc184adccebc5f6d2b6cc5510c8` |
| Runtime | Go 1.27.1, CGO 0, Darwin arm64 |

Verify 26 actual Git blobs before ranking: nine sources, 13 support files, three stored inputs and one plan. An independent read-only review also passed. Keep the executable outside Git. Preserve predeclared limits and Linux/macOS score comparison tolerance.

## What this permits

Retain `training_ready=false`. One synthetic English authoring workflow, missing roles and missing training execution plan prevent immediate fitting or production routing. First examine caption fidelity within the small input budget and transfer to different public sources. Then separately freeze a role plan assigning every member of each connected component together to exactly one of training, development-validation or calibration. Previously observed development data cannot become unseen validation or final evaluation by renaming a role.

No new CPU/RSS/GPU/latency, token or cost measurement ran. GOMAXPROCS 1 and a 256MiB Go heap soft limit are settings. Previous resource results retain their original scope. CI replays check the same result without adding official experiments or independent samples. Preserve earlier 56a/b/c/d/e/57 evidence and pending/unknown states.

See the [next preparation plan](NEXT-STORED-58.en.md) for work and stop conditions. The goal of 2,400 distinct protected-final requests per domain supports later generalization/final utility claims; it is not a universal prerequisite to every development fit. No new weights exist, so this phase publishes zero Hugging Face models. Remote CI is not yet run at document-writing time; record actual checks and merging in the issue afterward.
