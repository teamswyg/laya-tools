# Stage 60: original source correspondence results

The original inventory succeeded in **one attempt, zero retries**. All 60 source definitions and 204 authored component relations agreed with historical hashes. This supplies exact original spans for later evidence review. **It is not caption approval, a training result or evidence of model improvement.**

| Observed work | Result |
|---|---:|
| Original files / successful parses | 4 / 4 |
| Root definitions / completed matches | 60 / 60 |
| Component relations / completed matches | 204 / 204 |
| Raw span hashing and formatting calls | 264 each |
| Normalization / bundle calls | 112 / 24 |
| Git blobs verified before original inventory | 27 |
| Distinct component IDs | 70 |
| Distinct raw spans across all 264 references | 89 |

The 89 spans were counted from the result by path/start/end byte tuples; distinct raw and formatted SHA values also numbered 89 each. This does not replace 264 physical calls with 89. Repetition may motivate a later cache experiment, but no speed or memory savings were measured. Current policy is `no_cache_each_relation_v1`.

## Which execution produced this result

The [plan](execution-plan-60.json) pins source freeze `cc8502f872b047e02961322ff1fcdce4af46e59c` and input freeze `c0f279da4c2410699977b502ae4ca2fddd3d3ca8`. The actual Go1.27.1/darwin-arm64/CGO0/trimpath/no-VCS binary was 5,432,642 bytes, SHA `831b73221b06dd757572521979e4f6fe0e15f1f5ccb4801dc0eddecaf05390c2`. This is executable size, not model size or memory use.

The [original record](source-inventory-60.json) is **174,651 bytes**, SHA `d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4`. The plan is 5,643 bytes, SHA `555dd524cfa5da5b0bc6eecbd9e13a21dfab2a7683978fa4b949765077779a4c`. Success followed checks of returned files, root/component order, hashes and actual counters. Object binding and semantic closure discovery were not performed.

## Preserved failures and validation

There were two prepare calls. The first stopped with `inventory60_build_recipe_invalid` because literal `<private_binary>` differed from Go JSON's canonical HTML-escaped encoding. Original inventory, AST and formatter calls were zero, and no plan output was produced. Only the support file's JSON encoding changed; the second prepare succeeded using the same binary. This preparation failure is retained separately from original inventory retries.

An independent reference checker ran twice. Its first incorrect assumption that every parent had three candidates was fixed. The second run passed 14,338 metadata checks, preserving original two- and four-candidate parents, positions and ordering. It made no semantic verdicts. The [execution ledger](execution-ledger-60.json) and [preparation review](runner-review-60.json) distinguish these activities.

A regression check added after observation verifies archived result/input/support hashes and all 264 raw span hashes/physical lines. It does not rerun original inventory, AST or formatting. Synthetic controls, historical checks and future CI are not retroactively included in the first original attempt count. CI merging had not started when this document was written.

## Next work toward training

Independent content review has started for `stable-odd`, `atomic-commit` and `error-identity` under the [frozen rules](CONTENT-REVIEW-PROTOCOL-60.en.md) and [reader scope](content-reader-plan-60.json). It covers 123 of the existing 738 entries, separating source fidelity, request conditions, observed fields and excluded scope. Repeated captions and positions are not new independent requests. This document does not include or assume completed review findings.

All content states in this stage remain pending and `training_ready=false`. Existing 72 parents, 216 candidate positions, 21 unknowns and 17 connected groups are preserved. New labels, roles, fits, model/paid trials, weights and protected-final reads are zero. Existing qualified Hugging Face releases remain; source correspondence is not published as new model performance. See the [usage guide](USAGE-INVENTORY-60.en.md).
