# Short-claim experiment 56c: preparing truth and captions for one property

[한국어](PLAN-56c.ko.md) · [Next direction](NEXT-56c.en.md) · [56b results](RESULTS-56b.en.md)

**This document is a preparation design. It does not execute a new official property audit, role partition, ranking observation, fit, model call, performance measurement or final evaluation.** Writing or committing this document alone does not seal execution inputs, sources or observers. Once implementation and caption review are ready, freeze a separate execution manifest before outcomes. At the time of writing, 56b PR79 CI is in progress; this document claims neither completed verification nor merging.

## Intended user and use

The fictional user is an agent deciding which code to inspect first from short descriptions of Go functions. It may ask for a candidate that preserves a comma inside quotes or retains no partial tokens on syntax errors. The cheap tool proposes **inspection order only**; an external independent verifier checks the actual code. Preserve every candidate, original order and fallback. Do not delete low-scoring candidates or automatically grant execution approval.

A function satisfying one property may violate another. Scores are not correctness probabilities, whole-function certificates or execution approvals. The utility of repeated small hints and actual LLM usage must be tested separately.

## Preserve the starting evidence

56b's existing 72 requests, 17 total connected groups and 16 labeled connected groups are historical development evidence. Clearing the preparation minimum of 15 does not remove the coordinated synthetic English authoring process, missing role plan or missing diversity. It establishes neither statistical independence nor training readiness.

Preserve the 4 existing unknown delimiter requests and their original text, candidates, truth, groups, sealed manifests and results. Prepare scoped requests as separate development records with new IDs and original-parent links; do not overwrite the original unknown states. Preserve the historical 120-candidate public-source survey and protected-final/CoSQA reserves.

This plan authorizes preparation design only. Keep 0 roles, 0 partitions, 0 fits, 0 new weights, 0 model calls, 0 new rankings, 0 new performance runs and 0 paid calls. Protected-final access and scoring remain 0. Preserve 7 distinct actual coding requests, 20 records and exhausted paid-call ledgers.

## Initial scope: three properties and nine existing literals

Reuse only the 4 frozen sources in `internal/typedbehavior/flow.go`: `quoted-correct`, `quoted-literal-comma`, `quoted-inside-escape` and `quoted-partial-error`. Accept no new candidates, remote code, user callbacks or dynamic compilation input. Support only the closed registry with matching actual source and helper-closure hashes.

| Draft property | Existing literal inputs | Required observations and exclusions |
|---|---|---|
| `quoted-comma-content-v1` | `"a,b",c`, `ab"c,d"ef` — 2 | Compare successful return, Count and decoded tokens with literal expectations for these two valid ASCII inputs. Strip quotes and split commas outside quotes. This initial scope contains no backslashes and certifies no error behavior |
| `outside-escaped-comma-v1` | `a\,b,c`, `a\\,b` — 2 | Compare successful return, Count and decoded tokens for valid ASCII inputs without quotes. Backslash consumes the next byte as content. The comma after two backslashes is unescaped and splits. Escapes inside quotes and error behavior are outside scope |
| `syntax-zero-output-v1` | `"a,b`, `a\`, `a,b\`, `"a\`, `a,"b` — 5 | Every input independently designated a syntax error by the existing literals must return syntax error, Count=0 and the entire `[8]string{}`. Support only ASCII syntax-error inputs within 128 raw bytes, 8 tokens and 8 decoded bytes per token. Certify no ASCII/bounds-error behavior |

The total of 9 counts existing literals assigned to narrower contracts. It does not count 9 new independent requests, independent sources or new observations. Fixed literal expectations are neither model outputs nor independently acquired public-source requests. Each property claims finite literal observations only, not general correctness on inputs absent from the table. Extending to escaped commas inside quotes requires separately frozen literals, observations, captions and a plan.

Never generate expected Count, tokens or errors by executing a candidate. Review and transfer the existing independently authored literals, then hash the new table's inputs, observation fields, expectations and version together. Do not enter new acceptable sets or assertion counts as results before official observation.

## Whole-function labels and property labels differ

The following table is a **prediction from reading current code, not a new execution result**. A subsequent official audit must check it through property-specific literal observations.

| Existing source control | Existing whole-contract role | Quoted-comma prediction | Outside-escaped-comma prediction | Syntax zero-output prediction |
|---|---|---|---|---|
| `quoted-correct` | correct | Agrees | Agrees | Agrees |
| `quoted-literal-comma` | wrong | Disagrees | Disagrees | Disagrees: does not return syntax errors |
| `quoted-inside-escape` | wrong | Agrees | Disagrees | Disagrees: does not report trailing escapes outside quotes as errors |
| `quoted-partial-error` | wrong | Agrees | Agrees | Disagrees: retains partial output on syntax errors |

`quoted-partial-error` follows the correct successful parsing path but handles later errors differently. Initial non-ASCII or raw input exceeding 128 bytes returns an error before parsing and clears output. Do not expand “clears output on initial errors” into “clears output on all errors.”

Do not apply the existing audit's requirement that every whole-contract wrong source fail at least 1 vector to a property audit. Do not reuse whole-contract control roles, existing parent known/no-answer/unknown states or acceptable arrays as property labels. A source can satisfy a property, and a candidate set containing only whole-contract wrong sources can still contain a property answer. Preserve the existing 56b audit and prepare separate property observations.

`syntax-zero-output-v1` is non-vacuous. Do not inspect output only when `got.Error!=OK`. Each of the 5 fixed inputs must return the specified syntax error, so omitting the error is also a mismatch. Count=0 alone is insufficient: the whole token array, including unused slots, must be empty. A weaker property that does not check exact error kinds must be declared separately and must not be mixed with this property.

## Complete short English captions

The actual ranking input consists only of English request and candidate text. Every caption must be valid UTF-8, at most 512 raw bytes and 512 normalized bytes, and at most 32 normalized words. Apply the same normalization as `shortclaim.Validate`, not an approximate whitespace word count. Do not truncate, automatically summarize or fill missing meaning through hidden registry metadata.

The following are **drafts for subsequent caption review**. Their actual bounds and independent semantic fidelity have not yet been checked; this document assigns no READY status.

| Property | Draft English request |
|---|---|
| quoted comma | ASCII without backslashes; ≤128 input bytes, ≤8 tokens, ≤8 decoded bytes/token. Preserve commas inside balanced double quotes, strip quotes, split outside commas; never panic. |
| outside escape | ASCII, no quotes or trailing escape; ≤128 input bytes, ≤8 tokens, ≤8 decoded bytes/token. Backslash consumes next byte as content; split unescaped commas; never panic. |
| syntax zero-output | ASCII ≤128 input bytes; ≤8 tokens, ≤8 decoded bytes/token. Unclosed quotes or trailing escapes must return syntax error, zero Count and eight empty Tokens; never panic. |

Review separately whether the narrower scope can actually be communicated faithfully within 32 normalized words. Every candidate caption must accurately describe its source and the property's conditions and observations. Reusing original candidate text does not waive new property-fidelity review. If a new caption is needed, give it a separate version/hash and preserve the 56b text. Korean/English document translations add neither evaluation inputs nor independent requests.

## New contract, sources, features and groups

Separate new wire/result schemas from existing v2. Proposed names are `riido-scoped-property-probes-v1`, `riido-scoped-property-truth-v1` and `riido-scoped-property-preparation-plan-v1`; freeze actual structures, versions and limits in the subsequent execution manifest.

Each parent carries a new ID, original-parent ID, property ID/version, request, original candidate order and candidate source/code/bundle pins. Each property definition carries supported scope, literal inputs/expectations, comparison fields, error/panic/unknown policy, caption-review state and table hash. Bind both the actual compiled adapter source and the 4 original source closures. Do not assume the existing `SourceArtifacts()` automatically includes a new adapter file. Explicitly include adapters, literals, schemas, captions and observer implementation in new compiled provenance, and verify agreement between disk sources and the executable.

Project only request/candidate text into ranking features. Property IDs, original parents, source IDs, function types, code/bundle hashes, literals, ExpectedControl, failure counts, acceptable sets, roles and truth are not features. Changing metadata while preserving text must leave the projection unchanged. Independent verification and oracles belong to the audit path only; ranking must not look up truth.

Transitively join original parents, every positive/negative source, shared `scanQuoted` and authored helpers/types/globals, prototypes and copied code/captions. Include original unknown parents. Do not use property IDs as new prototypes/cores to split relations. More properties, paraphrases, translations, seeds, siblings or INT8/PTQ children do not create independent groups. Do not report new group or labeled-group counts as achievements or estimates before the official audit.

## Preparation READY checklist and stop conditions

These are **requirements still to satisfy**, not completed checks.

- [ ] Freeze the 3 scopes, exclusions, 9 literals, observation fields, independent expectations and error/panic policies.
- [ ] Pin the 4 original source/code/helper closures and new compiled observer source; verify the closed Go 1.27.1 support scope.
- [ ] Check every request/candidate with actual raw/normalized 512-byte and normalized 32-word rules, and complete independent semantic review. Hold overlong/incomplete captions without truncating them.
- [ ] Separate unknown, known and no-answer with reasons in the new schema. Unsupported inputs, incomplete descriptions and observer failures are unknown, not negative labels.
- [ ] Treat omitted required errors, residual slots and safely observed no-panic violations as property mismatches. Never treat missing observations as passes.
- [ ] Verify metadata exclusion, retained candidates/fallbacks and original-parent/all-source/helper relations.
- [ ] Preserve the 4 original 56b unknown delimiter requests and all existing inputs, results and sealed files exactly.
- [ ] Commit and seal actual implementation, inputs, caption review, sources, literals and execution manifest before a separately planned single official audit and CI replay.

Unsupported inputs and incomplete captions remain unknown. Safely captured candidate panics under supported no-panic contracts and incorrect returns are known mismatches. Corrupt JSON/hashes/source bindings, unresolved closures, unreliable observers, literal/caption disagreement, metadata leakage or missing relations stop preparation or the audit. Do not adapt expectations, captions, candidates or primary criteria to observed outcomes. Record the cause and a separate new plan/seal boundary when repairs are necessary. Use CI verification rather than human approval, while preserving separate semantic and source-rights reviews; CI success does not automatically replace them.

Prepared properties do not make training READY. Keep 0 fits without frozen diversity, training/validation/calibration roles, group partitions and new utility before fitting. Do not lower the minimum of 15 connected/labeled groups or the necessary utility condition of at least 5%. Do not read or label existing protected-final/CoSQA reserves. At least 2,400 distinct protected-final requests per domain remains a separate acquisition target.

## Subsequent utility and resource paths

A separate utility plan distinguishes answerable/no-answer/unknown and compares fixed order, BM25, lexical, narrow-rule and oracle under common cost denominators. Retain candidates and never turn unknowns into failed labels. Oracle headroom is neither achieved model benefit nor LLM-token savings. Prepare distinct public-source development requests, increasing from 120 to 240, alongside language/condition/authorship coverage, rights, pinned revisions, closed compilation and independent truth before freezing roles and selection rules ahead of outcomes.

A separate resource plan distinguishes identical versus distinct public requests in uncached resident JSONL processing. Separate startup, warmup, processing, pipes, controller work and whole-child RSS/CPU, then consider open-loop load with a bounded queue. Do not claim cache/lock/SIMD/GPU gains or treat small artifact bit counts as reduced whole-system memory or LLM costs.

Only after evidence of a useful FP32 parent, compare same-parent INT8/PTQ children and equal-budget ternary STE siblings under a separate plan. Do not execute the existing proposal of at most 8 fits and 16 derivative artifacts yet. New paid attempts require a separate execution plan and budget ledger. Publish only new models passing utility, actual source rights and CI checks as immutable Hugging Face versions; Git contains code, plans, public numbers and model references only. This document records neither a new model release nor completed release checks.
