# Scoped caption review for stage 60 — preparation proposal

Review whether existing captions faithfully describe their implementations, separately from whether those implementations satisfy the requested task. **This proposal fixes the method and evidence locations before assessment. Actual content assessments and approvals are 0; every state remains `pending`.** The first assessment requires a separately frozen protocol and successful original-declaration rebinding under the stage60 execution protocol.

The fixed families are `stable-odd`, `atomic-commit`, and `error-identity`: 12 existing parents, 36 candidate positions, 11 source IDs, 12 distinct candidate texts, and 34 finite cases. Repeated text or positions do not create independent requests. The scope corresponds to **123 preparation entries** within the existing 738: 12 request entries, 36 candidate entries each for closure/fidelity/coverage, and 3 contract-observation entries. Neither the selected 123 nor the full 738 has been assessed.

## Keep the four questions separate

| Existing axis | Review question |
|---|---|
| `source_fidelity` | Does the caption faithfully describe this implementation, including intentionally incorrect behavior? |
| `request_contract_coverage` | Which finite contract conditions do the request and caption cover, omit, or contradict? |
| `observation_fields` | Are return, error, state, order, ownership, and panic claims expressed within the actual observation scope? |
| `explicit_negative_boundaries` | Are inclusion/exclusion, malformed inputs, unsupported inputs, and scope boundaries explicit? |

Preserve the stage59 vocabulary: `pending`, `consistent_with_scoped_evidence`, `omits_required_scope`, `contradicts_scoped_evidence`, and `unsupported_or_uncertain`. Consistency on one axis does not pass another. A caption about retaining only positive odd integers may faithfully describe its implementation while contradicting a request to retain negative odds too. Never copy correct/wrong controls, acceptable candidates, failure counts, or ranking scores into fidelity judgments or training features.

## Evidence needed before a state can change

Mechanical checks cover original file/text SHA, byte counts and JSON pointers, candidate order/IDs, contract references, and distinct raw/formatted/normalized/bundle digest links. The successful stage60 report must bind actual runner/toolchain/binary/input/plan provenance and preserve partial failure counters. **Digest agreement or green CI alone does not approve meaning.**

An independent reader then links each original text clause to authored control flow and write order, required roots/helpers/types/methods/sentinels, observers, and literal fields. Record scoped standard-library assumptions too. A closure entry can be scoped as consistent only after binding and reading account for its authored dependencies within that scope. This does not claim general object resolution or whole-program closure analysis.

Each later assessment must record exact text byte ranges, source/declaration references, input scope, observation fields, exclusions, uncertainty, reviewer, and immutable record SHA. Predeclared rules use `omits_required_scope` for required omissions, `contradicts_scoped_evidence` for explicit disagreement, and `unsupported_or_uncertain` for competing interpretations or claims beyond observations. Those states have not yet been assigned. An axis cannot roll up to consistency while a material clause on that axis remains unresolved. Do not produce one all-purpose approved-candidate scalar.

## Boundaries to examine in the three families

- **Stable odd filtering:** Separate negatives, duplicates, order, exclusion of evens, and input changes. Existing `slices.Equal` observations do not distinguish nil from empty slices, and do not inspect aliasing, capacity, or allocation count. Candidate panic or input mutation is unknown in this legacy checker; do not replace that policy with a typed contract's mismatch policy.
- **Atomic configuration commit:** Separate successful return from destination writes, and partial returned values from partial destination mutation on failure. Link exact four-byte ASCII `DD,F`, `Count/Enabled`, exact `ErrSyntax`, `Returned/After/Error/Panicked`, and fresh nonnil destinations. Leave competing readings of `or zero if invalid`—invalid prefix digits versus invalid complete syntax—unresolved before assessment. A nonnil premise does not promise nil-destination safety.
- **Error identity:** Separate equal messages from target identity, wrapping/joining, direct comparisons, text matching, and `Is/As`. Returned original code and post-return `Cause.Code` mutation are distinct observations. Some fixtures have no cause or nil Input, so preserve the scope question around `Cause pointers are nonnil`. Checking `Cause.Code` does not observe whole-chain immutability for `without changing the error chain`. Typed nil and custom `Is/As/Unwrap` lie outside the fixtures.

The three ambiguous requests remain unknown. Preserve all 72 parents, 216 candidates, 21 unknowns, 17 connected groups, and 16 groups containing known parents. Resolving some caption evidence does not create truth labels, assign train/validation/calibration roles, or make `training_ready` true. Keep the separate target of 2400 distinct protected-final requests per domain; add no 2400-every-development-fit prerequisite or new 15-repository condition.

## Artifacts and next sequence

The [protocol JSON](content-review-protocol-60.json) holds checklists, state vocabulary, and future assessment-record fields. The [reference preparation JSON](content-review-scope-references-60.json) contains original stage59 text pointers/SHA and retained authored-component descriptors. Its rows are **selection/reference scaffolding**, not actual semantic assessment records. No literal table was reified, no candidate ran, and no new candidate-by-vector Got was created. Aggregate failures are not per-vector Got.

Next: freeze the protocol and execution evidence → verify successful declaration links → perform the scoped independent reading → record resolved axes alongside unsupported clauses. This preparation performs 0 original inventory, AST/formatter, SourcePins, Bind, Generate, candidate API, model/paid evaluations, labels, roles, fits, or weights. Add no human approval step; preserve the existing CI merge gate. Code reading alone establishes no cost, memory, or speed improvement.
