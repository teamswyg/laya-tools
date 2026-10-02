# Caption and implementation review of three families

We reviewed whether existing descriptions faithfully convey their implementations within a fixed scope. **An incorrect implementation can have a faithful caption.** Positive-only filtering and reverse-order filtering are accurately described, while conflicting with requests to retain negative odds in original order. We did not combine those questions into an approved-candidate scalar.

The review follows the frozen [protocol](content-review-protocol-60.json), [reference preparation](content-review-scope-references-60.json), and first original-declaration result SHA `d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4`. Source freeze: `cc8502f872b047e02961322ff1fcdce4af46e59c`; input freeze: `c0f279da4c2410699977b502ae4ca2fddd3d3ca8`. The preselected scope is 12 parents, 36 candidate positions, 11 source IDs, 12 distinct texts, and 34 retained literal cases across three families.

The reader did not author the original captions, literal fixtures, or CLI. This is not a blind assessment: earlier reference preparation exposed source IDs and stored outcome metadata. Actual judgments use original prose, source, contracts and observation fields only. Correct/wrong controls, old acceptable candidates, failure counts, and rankings are not copied into fidelity judgments.

## Recorded scope and states

| Existing stage59 entry kind | Consistent within scope | Contradictory | Omitted | Uncertain | Total |
|---|---:|---:|---:|---:|---:|
| Request/contract scope | 6 | 0 | 0 | 6 | 12 |
| Scoped candidate source dependencies | 36 | 0 | 0 | 0 | 36 |
| Caption/implementation fidelity | 30 | 0 | 0 | 6 | 36 |
| Caption/request coverage | 3 | 17 | 2 | 14 | 36 |
| Finite contract/observation fields | 3 | 0 | 0 | 0 | 3 |
| Total | **78** | **17** | **2** | **26** | **123** |

Consistency means `consistent_with_scoped_evidence`, not global candidate approval, training eligibility, or general closure proof. The original stage59 file remains an immutable snapshot with 738 pending entries. This separate overlay assesses 123 of those entries; 615 are untouched.

To retain all four axes, 72 supplemental observation/negative-boundary records are included. Observation states are 27 omissions and 9 uncertain; negative-boundary states are 9 consistent, 18 omissions, and 9 uncertain. These records do not increase original parents or the738 entry count. All 195 record objects retain text ranges, source references, observation scope, uncertainty, and record SHA. `source_closure_only` records are distinct from caption-fidelity assessments.

## Differences the reading exposes

- **Stable odd filtering:** Captions for stable retention, positive-only retention, and reversed retention faithfully describe their respective implementations. Positive-only and reversed descriptions conflict with the requested negative/order conditions. Their implemented inclusion/exclusion/order boundaries are understandable, while input-change and legacy panic/input-mutation unknown conditions are omitted. No nil/empty, aliasing, capacity, or allocation observation is inferred.
- **Atomic commit:** Descriptions of writing Count before full validation and returning successfully without destination writes are faithful. They conflict respectively with failure preservation and successful commit. The normal-commit caption covers core behavior but does not independently state the nonnil premise, detailed ASCII/flag grammar, field mapping, or no-panic condition; those omissions are recorded rather than filled from known code.
- **Error identity:** Message search, direct comparison, and returning original facts before incrementing the stored code are faithfully described. Message equality versus identity, wrapped/joined traversal versus direct comparison, and returned value versus later mutation are separate request-coverage problems. Mentioning `Is/As` does not approve whole-chain immutability.

Two captions retain material unresolved clauses. In partial-result atomic handling, `or zero if invalid` may refer to invalid prefix digits or invalid complete syntax. In the error caption, `without changing the error chain` can promise more than the observed `Cause.Code`. Their six repeated positions remain uncertain for fidelity. Exact original clause byte ranges are recorded, without rewriting either sentence.

The request clause `Cause pointers are nonnil` is also unresolved: it may concern present causes, the `errors.As` output-address argument, or exclusion of absent-cause fixtures. We did not choose a replacement interpretation. Preserve the three selected ambiguous requests, all21 original unknowns,17 groups, and16 groups containing known parents. Unknown is provenance uncertainty about intent, not a new learned label.

## Integrity checks and next work

Read the [assessment JSON](content-review-assessments-60.json), [source evidence catalog](content-review-evidence-60.json), and [byte receipt](content-review-byte-receipt-60.json) together. Separate byte verification confirms25 existing declaration references,8 observer/type text spans,3 literal-expression spans, and195 record quotes/ranges/required fields/SHA values.25 is a reference count including repeated root/component declarations, not25 distinct original declarations. Byte verification establishes correspondence, not independent proof of semantic interpretation.

There was1 actual semantic review and1 successful assessment-generation invocation, with0 generation failures. Earlier metadata preparation's3 invocations are separately recorded. The first byte verifier stopped on its own missing `+1` in inclusive line-number arithmetic; only the verifier changed, and its second invocation passed. Assessment and catalog bytes were neither regenerated nor altered. Additional original inventory, AST/formatter, SourcePins, Bind, Generate, candidate/model/paid evaluations, and fits are0. No new per-vector Got was fabricated.

Next, consider separately versioned clarifications of unresolved clauses and apply the same review to other families. Any changed prose must retain links to its original family without changing old truth. Roles, truth labels, fits, weights, and activation remain unchanged; `training_ready` is false. No cost, speed, or memory gain is inferred from this review. Preserve the existing CI merge gate without adding a human-approval step.

A separate reader produced the [artifact mechanics report](content-review-mechanics-60.json) and [attempt ledger](content-review-mechanics-ledger-60.json), checking195 records,123 original-entry joins,72 supplemental axes, quote ranges, references and hashes. The first invocation's340 failures came from the checker treating axis objects as strings and the `new_labels=0` execution counter as a learning feature. Only the checker changed; the second invocation had0 failures and0 artifact mutations. Its `private/` input paths are symbolic provenance names; those files are now public beside this document.

The added CI archive guard checks frozen bytes, joins and retained uncertainty. It performs no new semantic review or original behavior experiment. Passing mechanical checks does not prove interpretation, independent data origin or training eligibility.
