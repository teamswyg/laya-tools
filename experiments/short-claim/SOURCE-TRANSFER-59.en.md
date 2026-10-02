# Preparing two small public-source transfer properties 59

[한국어](SOURCE-TRANSFER-59.ko.md) · [Machine-readable source bindings](source-transfer-59.json) · [Next preparation](NEXT-STORED-58.en.md) · [Existing original observations 57](../public-behavior/RESULTS-57.en.md)

**This preparation proposes two properties and binds existing evidence. New requests, captions, truth, acceptable sets, role assignments and fits remain zero.** Read original source, documentation, licenses and stored results 57 without executing source APIs, Bind, evaluation, Baselines, models, paid calls or protected-final reads. Retain `planned_reference_only`, `execution_eligible=false` and `training_ready=false`. Creating documentation or recording pins does not complete executable contracts, caption review or authoring diversity.

58 found about 19.7802% ordering headroom between BM25's 91 checks and oracle's 73 on the existing 72 development requests. Existing 17 total groups and 16 labeled groups already meet the floor of 15. Do not split favorable groups or count the two external sources as new independent parents. Rankings, scores and group costs from 58 are already observed; future role rules cannot restore unobserved or blinded data.

## Proposed scope

| Proposal | Narrow question | API actually observed before | Unsupported scope so far |
|---|---|---|---|
| Strict parsing | Does it avoid coercion on listed finite version inputs and return the exact nil/error or preserved success fields? | `StrictNewVersion`, existing `s01–s16` in 57 | Permissive `NewVersion`, coercion flag changes, general correctness for all SemVer inputs |
| Slash-boundary matching | Do `*` and `/**/` distinguish single segments and zero/nested directories on specified `/` paths? | `Match`, existing `g01–g09` in 57 | `PathMatch`, Windows separators, filesystem traversal, complete glob grammar/error policies |

The initial target is **finite Go behavior hints from short English requests and English candidate captions**. The Korean guide is explanatory translation, not a Korean model-input performance test. New caption/independent expectation preparation, fidelity within the 512-byte and 32-normalized-word model-input limits, and candidate relationship audits remain future work. Do not use these summaries as completed model inputs or labels.

## Distinguish original files from behavioral dependencies

Use the [existing manifest](../../internal/publicbehavior/upstream-manifest.json). Preserve 15 original artifacts totaling 99,035 bytes; no new external download occurred. The JSON records exact revisions, modules, original/retained paths, sizes and SHAs. Do not extract or reformat original files.

The **semantic closure** links the entrypoint, behavior-determining types, constants, error sentinels, helpers and observed getters. The **compile closure** preserves the files needed to compile whole original files. Other functions compiled from the same file do not acquire call authority or observation evidence. This is a static source inventory; no new compile or replay validation ran.

| Source | Pinned revision and module | Original compile/attribution closure |
|---|---|---|
| [Masterminds/semver](https://github.com/Masterminds/semver/tree/61fc460d28283a91c53be65c2e0f20b494ac8ad9) | `61fc460d28283a91c53be65c2e0f20b494ac8ad9`, `github.com/Masterminds/semver/v3`, original Go 1.21 | `version.go`, `collection.go`, `constraints.go`, `doc.go`, `LICENSE.txt`, `metadata/upstream.go.mod.txt`: six files |
| [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar/tree/8b690afa33319b0a1869367f594e53977e38bc99) | `8b690afa33319b0a1869367f594e53977e38bc99`, `github.com/bmatcuk/doublestar/v4`, original Go 1.16 | `doublestar.go`, `glob.go`, `globoptions.go`, `globwalk.go`, `match.go`, `utils.go`, `validate.go`, `LICENSE`, `metadata/upstream.go.mod.txt`: nine files |

Preserve exact original module bytes under metadata names. Original Go directives are source metadata; the actual historical run in 57 used Go 1.27.1 on Darwin arm64. Recording a directive does not prove new execution under that language condition.

### Strict parsing dependencies

`StrictNewVersion` at [version.go line 89](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/version.go#L89) uses `Version`, `num/allowed`, `containsOnly`, `validatePrerelease`, `validateMetadata` and six original error sentinels. Bind `Original`, `String`, `Major`, `Minor`, `Patch`, `Prerelease` and `Metadata` getters when preserving the complete recorded success fields. Exact declaration lines and file pins are in JSON.

Bind the original [parsing documentation at line 13](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/doc.go#L13) and the description at `version.go` lines 84–88. `doc.go` lines 13–19 occupy original bytes `[328,853)`; the `version.go` description occupies `[3039,3397)`. Bytes are zero-based with an exclusive end; lines are one-based and inclusive. Span SHAs in JSON bind the exact bytes including LF.

`StrictNewVersion` does not read coercion flags. Preserve regex/global/init declarations through whole-file source pins without calling that evidence for `NewVersion`. The existing observer called only the strict API. Documentation or source-reading alone cannot add permissive results as Got; those remain unknown until a separate contract and observation plan exist.

### Slash-boundary dependencies

`Match` at [match.go line 52](https://github.com/bmatcuk/doublestar/blob/8b690afa33319b0a1869367f594e53977e38bc99/match.go#L52) passes `/`, validation=true and case-insensitive=false. Its semantic closure includes `matchWithSeparator`, recursive `doMatchWithSeparator`, `matchRune`, `isZeroLengthPattern`, escape/alternative search, `utils.go:indexNextAlt`, `validate.go:doValidatePattern` and original `ErrBadPattern`.

Documentation lines 9–51 occupy original bytes `[76,1977)`; the core grammar at lines 15–17 occupies `[197,373)`. JSON records span SHAs. Whole `utils.go` also contains `FilepathGlob`, introducing compile dependencies on glob/options/walk code. Preserve the seven original Go files. This does not mean filesystem APIs ran or received execution authority.

## Reference stored Want/Got exactly

`vectors[].checks` in [probes-57.json](../public-behavior/probes-57.json) contains predeclared Want. `audit.rows[].checks` in [results-57.json](../public-behavior/results-57.json) contains stored Want/Got/match evidence. The new JSON binds **exact JSON pointers, field names and types** instead of generating values or reserializing large integers.

- The 16 existing strict rows have 14 fields: `error_cause`, `error_func`, `error_kind`, `error_num`, `major/minor/patch`, `metadata`, `nil_version`, `original`, `panicked`, `prerelease`, `string` and `supported`.
- The nine existing glob rows have seven fields: four error fields plus `matched`, `panicked` and `supported`.
- Stored rows do not duplicate input. First verify that the stored record's `/probes_sha256` matches the pinned original probes bytes, then read input from `/vectors/N/input`. Resolve each reference's expected checks and stored row/checks/status/match pointers. Verify ID, property, source family and expectation kind together. Check field names and types before using a declared check index.
- Preserve `Want/Got` as `json.RawMessage` and exact uint64/bool/string types. Do not lose maximum uint64 values through float64 or ordinary numeric JSON roundtrips.
- Do not promote observations without Want, such as `error_message`, `Versions`, counters or default values, to new truth. References are cold metadata, not scorer features.

All 25 referenced rows were complete and matched their finite expectations in 57. This is a reference to existing observations, not a new replay, 25 independent requests or truth for new captions. Partial field agreement does not establish fidelity of a general function description.

## Preserve discrepancies and uncertainty

Strict API nilness, sentinels and error-check order are descriptive contracts, not error kinds prescribed by SemVer. `s06-core-overflow` is API rejection of valid decimal syntax outside its uint64 representation; do not relabel it SemVer-invalid. Successful parsing of a large prerelease number does not prove correct numeric comparison.

Existing comparison discrepancies `p08` and `p09` are a reversed pair of the same large-number behavior. `b05` preserves `Match("{a,[}","a")` returning true with no error, versus a stronger whole-pattern validation policy. Do not rewrite Got or synthesize a preceding `ValidatePattern` gate. JSON preserves exact row pointers for all three disagreements.

New inputs/candidates, broader Unicode/grammar, wrappers, flag changes or stronger policies remain unknown or outside scope until support is established. Unsupported operations, UTF-8/byte-bound rejection, panic and unclassified errors are not candidate-failure labels. New correct/wrong candidate controls have not been authored or executed.

## Authorship, grouping and licenses

Semver and doublestar are pre-existing public upstream code/documentation. MIT notices name Matt Butcher/Matt Farina for semver and Bob Matcuk for doublestar. This does not audit every function's authors or certify human-only authorship. Existing project data uses a collaborative AI-assisted synthetic English authoring workflow; no transfer captions exist yet. External code or a different reviewer alone does not resolve `synthetic_single_pipeline`.

Strict failure/error observations provide comparison dimensions with `atomic-commit`/`error-identity`; glob context/boundaries provide dimensions with `prefix-balance`/`quoted-delimiters`. Similarity alone establishes neither group equivalence nor independence. Transitively bind actual shared helpers/types/sentinels, copied code, aliases, duplicate captions, parent/sibling relations and translations. Preserve shared closures of other properties from each upstream. No new group count was computed or claimed.

Keep full original [semver LICENSE](https://github.com/Masterminds/semver/blob/61fc460d28283a91c53be65c2e0f20b494ac8ad9/LICENSE.txt) and [doublestar LICENSE](https://github.com/bmatcuk/doublestar/blob/8b690afa33319b0a1869367f594e53977e38bc99/LICENSE), including copyright notices. Recording source licenses does not replace future training-asset and weight-publication eligibility checks.

Preserve the existing 72 parents, 17 groups, 16 labeled groups, 21 unknowns and original records. Caption fidelity, role/coverage rules and a separate execution plan's automatic checks remain necessary; no new human approval flow is added. Models, fits, new HF weights and production activation remain zero. At least 2,400 distinct protected-final requests per domain remains a separate generalization target, not a prerequisite to every development fit. No protected-final/CoSQA reserve was read or reassigned.
