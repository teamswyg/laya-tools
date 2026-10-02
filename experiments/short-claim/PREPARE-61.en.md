# 61: Prepare content review of the remaining existing captions

This draft fixes the next scope for comparing existing candidate captions with code. Only stored JSON hashes and metadata references have been checked. All content assessments are `pending`; the parent task will handle publication, protocol freeze and independent content reading. New samples, truth, roles, fits and model calls are all 0.

Selection excludes `stable-odd`, `atomic-commit` and `error-identity` from the complete 59 family list, independently of observed scores. Original parent and candidate order is retained. Selected parent indices are 0–15, 20–47 and 56–71. The source rebinding already performed in 60 has not been rerun.

| Prepared scope | Actual stored metadata count |
|---|---:|
| Families / existing parents / candidate positions | 15 / 60 / 180 |
| Parents with 2 / 3 / 4 candidates | 11 / 38 / 11 |
| Existing known / no_answer / unknown parents | 28 / 14 / 18 |
| Source IDs / distinct candidate caption SHAs | 49 / 60 |
| Contracts / retained literal cases | 15 / 114 |
| Existing 59 review slots | 615 |
| Component relations / IDs / formatted SHAs for selected sources | 172 / 56 / 39 |

The 615 slots are 60 requests + 180 candidate closure slots + 180 candidate fidelity slots + 180 candidate coverage slots + 15 contract observation slots. Neither four pending axes per candidate nor 60 distinct caption hashes are independent requests. The complete existing corpus remains 72 parents, 216 candidate positions, 17 groups, 16 groups containing known/no_answer and 21 unknown parents. The subset touches 14 existing groups; `closed-window` and `clamp-window` retain their shared group 8.

The families are retain-active, allow-owner, closed-window, quota-total, remove-first, compact-runs, rotate-left, nondecreasing, clamp-window, prefix-balance, transform-order, owned-snapshot, cancellation-lifecycle, quoted-delimiters and ancestor-cycle. Each family has 4 existing parents and 12 candidate positions, but individual parents have different candidate counts. All 4 quoted-delimiters parents retain their original unknown state.

The four pinned inputs are below. The exact bytes read were hash-checked before those same bytes were decoded as JSON.

| Input | SHA-256 |
|---|---|
| caption-coverage-59.json | `7c1bd449d533d3d46a9211dea0a436ce338fe8f16ee3197d554744c3c24b0a29` |
| source-inventory-60.json | `d99d635929555d4123bbaa84aac99d0fb29db345f8b813ed70074dd3e09cdbd4` |
| content-review-protocol-60.json | `6f4504ba04e0b5813e4dc58ec0f6dbe5101a1127e09a379172dd57d7ed11ff19` |
| content-review-recipe-59.json | `17f0abc9843e082d53e78f0eff79842d592db89f2532c0ece0e3a44814065cff` |

All four repository paths are under `experiments/short-claim/`. The new scope references original JSON pointers, text SHAs and byte counts instead of copying request/candidate prose. Each candidate's historical formatted root SHA and typed bundle SHA are joined to its stored 60 root. Component relation order and IDs are preserved. Multiple IDs sharing one raw enum declaration are not collapsed. This confirms metadata correspondence, not object binding or complete source closure.

The axes and vocabulary are unchanged from 59/60. The axes are `source_fidelity`, `request_contract_coverage`, `observation_fields` and `explicit_negative_boundaries`. The states are `pending`, `consistent_with_scoped_evidence`, `omits_required_scope`, `contradicts_scoped_evidence` and `unsupported_or_uncertain`. A caption can faithfully describe a wrong implementation, so fidelity and request fulfillment remain separate.

The next reading proceeds in original order. Each claim must reference exact prose byte ranges, the root and all relevant authored helpers/types/methods/sentinels, the observer, literal Input/Want, observation fields and exclusions. An axis cannot become consistent while any material omission, contradiction or competing interpretation remains. Historical truth unknown and future content uncertainty are distinct. A limited clear caption claim cannot promote original unknown truth to an answer. Do not invent per-vector Got or unobserved general guarantees.

`boundary-checklists-61.json` contains reading questions, not assessments. They cover inclusive bounds, mandatory adjacent-run collapse, operation order, snapshot nil/empty and sequential writes in both directions, cancellation result-before-release order, complete tokenizer error/zero-array observations, and ancestor cycles versus shared descendants. Legacy observer panic/input mutation remains unknown; a typed candidate panic is an observed mismatch. Keep this distinction and review standard-library/infrastructure assumptions and supported scope before clearing a claim.

The Go preparation tool reads stored metadata only. Synthetic race and vet checks passed; controls reject altered order, root/bundle pins, duplicate IDs, invalid states and literal counts. One private metadata invocation completed 4 input joins, 180 candidate joins and 15 contract joins, with 0 failures. The ledger separates attempts, completions and fixed failure codes. Synthetic refusal tests are separate from official content-review retries. Original AST/formatter/Rebind/SourcePins/Bind/Generate, candidate APIs, rankings, paid/model calls, fits and protected-final reads are all 0. No repository file was modified.

The larger reference file belongs to maintainer tooling and makes no hint-runtime memory or speed claim. Prior development scores and truth have been observed; this is not blind validation. The target remains scoped English Go claims, and the Korean explanatory document is not a language test. `synthetic_single_pipeline`, `no_roles_plan`, role assignment and authoring-flow diversity remain unresolved; `training_ready=false`. CI can check file/reference integrity but cannot substitute for content fidelity. No new human approval flow, 15-repository requirement or 2,400-sample minimum for every development fit is added. The separate protected final target of at least 2,400 independent requests per domain is preserved.
