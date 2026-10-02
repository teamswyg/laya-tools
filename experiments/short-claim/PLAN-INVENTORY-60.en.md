# Stage 60: source declaration and historical digest rebinding — proposed plan

Stage 60 prepares historical digest links and a later caption-review protocol. **It does not approve captions, create truth, or train models.** Code, input, binary, and protocol freezes remain pending; actual inventory execution is 0. This proposal is not a freeze or completed review.

The eventual purpose is evidence for a small, non-decisive hint model. “Does this caption faithfully describe the implementation?” and “Does that implementation satisfy the request?” are different questions. An `odd-positive` implementation can fail the task while its caption about retaining positive odd integers remains faithful. Never copy correct/wrong controls, acceptable/rejected outcomes, or ranking scores into fidelity assessments.

## Existing scope and fixed inputs

There are 36 legacy and 24 typed source roots. The **typed** inventory has 204 relations, 70 distinct component IDs, and 53 distinct formatted SHA values, excluding the 36 legacy roots. The unique raw SHA count is uncomputed; do not assume 53. These are declaration/reference counts, not independent tasks or training samples.

Pin the exact bytes and SHA values of these 4 original Go files and the historical result. Do not change original candidates, helpers, types, or sentinels.

| File | Bytes | SHA256 |
|---|---:|---|
| `internal/behaviorprobe/data.go` | 28420 | `b377343a64c019c205f69865286a21ce5d91dda3051c31aa5c69b75a40300559` |
| `internal/typedbehavior/spec.go` | 1228 | `013a486ecc1d28913c2dda8cf74b3068dcb193d8b77e534dc62e2f87b9ee6b80` |
| `internal/typedbehavior/state.go` | 19402 | `35324ed595bcf384c484f33d98c6866164bebb2b636986ee8ea51bb979bdad7c` |
| `internal/typedbehavior/flow.go` | 18242 | `ea7f09d31e15e561173932ac0bb61db17746cdce929e0f976d97994d282d898b` |
| `experiments/short-claim/results-56b.json` | 266818 | `4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4` |

## What the AST inventory will verify

Verify pins before parsing top-level declarations. Preserve source IDs, roots, component IDs, relation order, and function/method/type/global/constant/sentinel node kinds. Retain historical formatting recipes and block scope, including const/iota and multiple value declarations.

Record 0-based, end-exclusive byte spans and physical lines, ignoring `//line` virtual positions. Distinguish raw, formatted, scanner-normalized, and serialized literal-table SHA. Rebind historical code/component/bundle recipes and standard-import/Go records; digest agreement is not semantic approval.

Use AST only: no new type checking, object closure, `SourcePins`, importer, registry, callback, or API calls. Do not execute original package initializers or candidates. Bind/Generate, features/baselines/ranking, models, and fits are excluded. Rebinding helper/type/sentinel records does not approve new semantic closure.

The [Go module](../../internal/sourceinventory/README.md) checks correspondence between AST identifier names and declaration descriptors. It does not prove general shadow resolution or actual runtime dispatch binding. The caller's GoVersion string is distinct from compiler verification; the later runner must bind build info, binary bytes and compiled-source pins.

The fixed policy is **no cache**: calculate each original root and component relation separately. Separate attempted/completed physical file hashing, parsing, formatting, scanner normalization, and bundle calls from logical relation comparisons. Physical completed means the call returned, including an error return. ParsesSucceeded, expected digest agreement and relation completion are separate counters. Preserve all 204 relations even when distinct enum IDs share one declaration block. A later cache version must independently specify raw identity plus recipe, hits/misses, actual calls, and unchanged results.

The [preparation recipe](source-inventory-recipe-60.json) derives expected successful work counts from retained metadata: 4 parses, 60 roots, 204 component relations, 264 format calls, 112 normalizations, and 24 bundles. **These are expectations, not observed original-inventory results.** The 52 typed function/method relations plus 60 roots account for 112 normalizations. Never substitute 70 distinct IDs or 53 digests for physical call counts. Unique raw declarations remain uncomputed. Operation counts are not speed, CPU, RSS, or GPU measurements.

Test boundaries include changed pins, missing/duplicate declarations, wrong roots/components/node kinds, local declarations, physical positions, repeated relations, and recipe differences. Automated binding checks do not approve caption meaning.

## Keep the existing review axes

Retain the 4 canonical axes in the [stage 59 recipe](content-review-recipe-59.json). Do not silently replace their names or state schema.

| Existing axis | Review question |
|---|---|
| `source_fidelity` | Does the caption faithfully describe the implementation, including incorrect behavior? |
| `request_contract_coverage` | What scope does the request state, and what does the caption cover, omit, or contradict? |
| `observation_fields` | Are return, error, state, order, ownership, and no-panic observations expressed within their scope? |
| `explicit_negative_boundaries` | Are negation, inclusion/exclusion, malformed inputs, and unsupported scope explicit? |

Keep the state vocabulary `pending`, `consistent_with_scoped_evidence`, `omits_required_scope`, `contradicts_scoped_evidence`, and `unsupported_or_uncertain`. **Every actual content state remains pending in this stage.** Scoped assessments may be recorded only under a later separately frozen protocol, and never promoted into new truth labels or global approval. Unknown truth and pending review are different states.

A later row links original parent/candidate positions and IDs, source, request/caption pointers, UTF-8 SHA/bytes and clause spans, contract/source pins, helper/type/sentinel/observer/literal declarations, input/observation fields, negative boundaries, uncertainty, and reviewer/record SHA. Metadata stays outside features. Preserve original prose within 512 raw bytes and 32 normalized words; never truncate to pass. Aggregate checked/failed is not per-vector Got.

## Protocol examples over the 3 fixed families

These families are fixed before this preparation, rather than selected from favorable ranking results. This is a planned review scope, not actual review records.

| Family | Parents / positions / source IDs / distinct texts / finite cases | Original positions |
|---|---|---|
| stable-odd | 4 / 12 / 3 / 4 / 5 | `probes-56.json` parents 16–19 |
| atomic-commit | 4 / 12 / 4 / 4 / 18 | `probes-56b.json` parents 0–3 |
| error-identity | 4 / 12 / 4 / 4 / 11 | `probes-56b.json` parents 4–7 |
| Total | 12 / 36 / 11 / 12 / 34 | Preserve original order and relations |

`probe-05-ambiguous`, `typed56b-p01-4`, and `typed56b-p02-4` remain unknown. The 34 cases are existing literals; repeated positions do not multiply independent requests.

- stable-odd: Separate negative values, duplicates, and order for `odd-correct/positive/reversed`. The checker's slices.Equal does not distinguish nil from empty. Aliasing, capacity, and allocation volume are not observed. Legacy panic/input mutation becomes unknown.
- atomic-commit: Link the 4 roots and `decodeAtomic/atomicConfig/ErrSyntax` plus observers. Separate Returned, After, the exact sentinel enum, and Panicked. Leave the meaning of “or zero if invalid” unresolved between prefix digits and complete syntax. A nonnil destination premise does not establish no-panic behavior for nil destinations.
- error-identity: Link the 4 roots and `identityFacts/identityCause/Error`, fixtures, and observers. Distinguish returned facts from post-call Cause.Code. Preserve uncertainty about “Cause pointers are nonnil” and “error chain” preservation. Typed nil, custom Is/As/Unwrap, and whole-chain immutability are outside the verified observations.

## Execution and completion boundaries

Run the first inventory only after its own code, support protocol, inputs, build recipe, binary, and plan freezes. Current original source execution is 0. Success verifies declaration/historical digest links, not all 216 content reviews, new behavior observations, or utility. Content judgment remains a later scope.

Preserve original 72/216, all 21 unknowns, 17 groups including 16 labeled groups, relation order, and truth. Record fixed failure/incomplete states on pin, span, or binding mismatch. Leave insufficient interpretation or observation coverage pending. Add no new 15-repository gate or 2400-sample prerequisite for every fit. Existing 17/16 already passes its existing minimum. The target of 2400 distinct protected-final requests is a separate generalization goal; do not read or reassign protected final or the CoSQA reserve.

This preparation records a Go module, synthetic checks, and a proposed plan for code review and CI. Original inventory/content records and new SourcePins/Bind/Generate/candidate/model/paid calls, labels, roles, fits, weights, approvals, and activation remain 0. All 738 content/closure review entries from59 remain pending. Existing repository regression and native model integration checks continue; their replays are not new60 collection or model-performance experiments. Add no human approval steps.
