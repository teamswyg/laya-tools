# What to prepare after finding ordering headroom

[한국어](NEXT-STORED-58.ko.md) · [58 results](RESULTS-STORED-58.en.md) · [Pre-observation plan](PLAN-STORED-58.en.md) · [Public sources 57](../public-behavior/RESULTS-57.en.md)

**58 found room that could justify preparing a model. It did not establish readiness to train one.** The best fixed baseline, BM25, required 91 checks across 51 requests with known outcomes; an answer-knowing oracle required 73. The gap, 18/91 or about 19.7802%, passed the necessary 5% utility floor. Original order required 109 checks, lexical/order features 92, and the narrow rule 91. All 72 narrow-rule requests fell back to BM25. This is neither an improvement achieved by a model nor measured execution savings.

The existing 17 connected groups and 16 labeled groups also already meet the floor of 15. The next task is not to split favorable groups or inflate counts. **Prepare evidence that captions faithfully describe behavior, transfer to other authoring origins is testable, and training, development validation and calibration have defined roles.**

This document proposes preparation 59. New executions, labels, role assignments, acquired sources, fits, weights, model/paid calls and protected-final reads remain zero; retain `training_ready=false`. Fits remain zero because preparation conditions are incomplete. No new human approval process is introduced. Later execution eligibility follows a predeclared plan, automatic checks and CI gates within the existing user authorization and merge policy.

## Three preparation outputs

The names below are proposed future deliverables. Writing this document does not mean those files exist or their reviews are complete.

| Proposed output | Contents | Acceptance for preparation | Stop or abstention condition |
|---|---|---|---|
| `caption-coverage-59.json` | Bind all existing 72 parents and 216 captions to 18 prototypes, input scope, observation fields, source and literal contracts | Independently review each caption's scope and evidence, recording omissions and ambiguity | Hold mismatched captions, budget violations or unsupported observations. Do not change stored truth to fit prose |
| `source-transfer-59.json` | Propose one scoped property from each already acquired semver/glob source in 57 | Bind original revision, licenses, files/documentation, helper relations and independent expected-value design | Stop execution and labeling while rights, closure, caption fidelity or normative-versus-actual behavior remain unresolved |
| `role-recipe-59.json` | Define whole-component rules for development training, development validation, development calibration and source-transfer diagnostics | Freeze relationships, deterministic rules, seed, required per-role coverage and shortage conditions before execution | Hold assignments and fits if components cross roles or coverage is insufficient. Never reselect groups or seeds for favorable scores |

### 1. Check caption fidelity first

The current data contains 34 answerable, 17 no-answer and 21 unknown parents. Even the 51 known outcomes do not certify general function understanding from short prose. Read each request and candidate caption together with its required behavior and the input scope actually checked. Bind relevant observations beyond return values: error identity, call order, atomic state changes, ownership and input preservation. Check that negation, inclusive/exclusive boundaries, order and error conditions are explicit.

Record original parent/candidate IDs and order, text SHA, contract/source pins, declared scope, independent review status and remaining uncertainty. Do not use 58 rankings or per-group gains to select prose during review. Scorers continue to receive only the request and ordered candidate captions. IDs, truth, sources, groups, roles and review status remain outside features.

Keep the existing 512-byte and 32-normalized-word limits. Do not truncate difficult behavior into an apparently correct caption. A narrower property requires a separate version linked to its original parent, source and connected component. Revised prose, translations, negatives and reversed inputs are not new independent parents. Preserve existing 56a/b captions, labels, 17 groups and all 21 unknowns. Unknown does not become incorrect or no-answer.

### 2. Prepare a small transfer test across authoring origins

The current 72 parents come from one collaborative synthetic English authoring workflow. Adding a different reviewing agent does not by itself resolve `synthetic_single_pipeline`. Use the two original code/documentation sources and MIT notices already acquired in 57 to distinguish pre-existing authoring evidence from captions created in this project.

| Existing public source | Proposed next scoped property | Boundary to retain |
|---|---|---|
| [Pinned Masterminds/semver source](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9), [parsing documentation](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/doc.go) | Describe strict parsing versus permissive version coercion on explicitly listed finite inputs | Keep permissive `NewVersion` separate from the strict contract. Freeze a separate observation plan before executing another API |
| [Pinned doublestar source](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99), [Match documentation](https://github.com/bmatcuk/doublestar/blob/8b690afa33319b0a1869367f594e53977e38bc99/match.go) | Distinguish the directory scope of `*` and `/**/` on specified `/` paths | Do not extend evidence to the entire glob grammar or platform-dependent `PathMatch` performance |

Bind original documentation locations/bytes and the new caption authoring process separately. Preserve revisions, file SHAs, module identities and MIT copyright/notices from the [existing source manifest](../../internal/publicbehavior/upstream-manifest.json). Design independent literal expectations and supported/unknown policies before observation; do not generate expected values by copying implementation behavior. A short phrase from documentation is not automatically a complete contract.

These are **two scoped proposals**, with zero new parents, truth or acceptable sets so far. Do not convert 57's 62 observations into 62 requests or count wrappers sharing helpers as separate families. Preserve 57's large-number comparison discrepancies and stronger glob error policy. Distinguish normative expectations, observed API behavior and stronger requested policies. External provenance does not justify breaking relationships to synthetic prototypes.

The initial target is **finite Go behavior hints from short English requests and English candidate captions**. The Korean document is an explanatory translation, not a Korean model-input performance test. Korean inputs or broader natural language require a separate caption-fidelity, budget and utility plan; translations of the same request remain in the same connected component. These two sources do not establish diversity across the whole public Go ecosystem.

### 3. Assign roles to whole connected components

58 rankings, scores and group costs are **already observed**. Freezing future membership hashes and seed rules prevents selecting favorable groups or seeds to change assignments. It does not restore an unobserved dataset or recreate blinded validation.

Transitively join shared prototypes, semantic cores, copied code, authored helpers/types/error sentinels, duplicate requests/captions, negatives, parent/sibling relations and translations. Preserve reasons and source pins for existing standard-library and observer infrastructure exceptions; do not add exclusions to inflate group counts. Derive a stable SHA from each complete component's sorted membership and source relationships. One component always receives one role.

Distinguish development roles such as `development_train`, `development_validation`, `development_calibration` and `transfer_development`. First specify deterministic rules, a fixed seed, required behavioral/prose/language coverage per role and shortage conditions. Renaming previously observed data cannot make it unseen validation or protected final. Report actual assignment counts only after checking relationships and coverage; do not present proposed ratios or roles as acquired evidence.

Keep unknown-only components in the relationship graph and full denominator. Exclude the 21 unknowns from training labels, cost and Top metrics, and distinguish groups with known parents. If the required roles lack coverage, prepare real source evidence instead of shrinking the scope or lowering floors. A seed cannot establish statistical independence between examples from one authoring workflow.

## Only then prepare a training execution plan

After the three outputs and independent content review are complete, separately freeze dataset/source/group/role SHAs, controls, utility/quality/abstention gates, memory/storage budgets, fit counts and publication eligibility. This document authorizes no new fit count or execution. Eligibility means passing automatic checks and CI within existing user authorization, not adding human approval.

1. **Primary FP32 candidate:** Begin a development comparison with a small coefficient model as a proposal. Record an actual trained baseline checkpoint with the same roles, inputs and candidate-preservation rules. Preparation does not guarantee suitability or performance.
2. **INT8/PTQ child:** Bind a compressed derivative to that same FP32 checkpoint. It is neither a new independent fit nor acquired data. Compare ordering, quality, abstention and actual runtime resources separately.
3. **Ternary STE sibling:** Treat separate optimization on the same data and roles as a sibling experiment. It does not replace FP32/INT8 results. A 1.58-bit representation alone cannot prove total memory or CPU/GPU improvement. Distinguish fit, seed and artifact counts from request counts.

Retain `no_roles_plan`, `synthetic_single_pipeline` and zero fits for now. Preserve failures, unknowns, retries and every candidate; stop preparation or training when necessary conditions fail. Production activation, default-model replacement and new HF weight publication are outside this phase. If weights are later produced, separately check source/training-asset/weight licenses and publication eligibility. Keep model bodies outside Git.

The target of **at least 2,400 distinct protected-final requests per domain** remains separate evidence for generalization and final utility claims. It is not a universal prerequisite to every development fit. Do not read or reassign existing protected-final/CoSQA reserves, or fill final counts with repetitions, prose variants, translations or model lineages. Writing this next-step document adds no execution, completed CI, acquired data or training outcome.
