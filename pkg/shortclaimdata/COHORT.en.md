# Cohort input that preserves unknowns

`LoadCohortRow` is an **opt-in Go API** for the next development cohort with exactly five or eight candidates. The existing `LoadDevelopmentRow`, finite37 data and saved comparisons are unchanged. This addition introduces no new observed requests or trained model.

An unverified candidate must not become a negative training label. Ambiguity must not erase facts already recorded. This API keeps text, declared truth, eligibility and provenance separate.

| Input | Audit | Conservative learning bridge |
|---|---|---|
| T / `label:true` | Known positive | Used when other conditions pass |
| F / `label:false` | Known negative | Used when other conditions pass |
| U / `label:null` | Unknown | Entire parent withheld |
| Ambiguous interpretation | Original T/F/null retained | Entire parent withheld |
| Known label with weight zero | Truth retained | Zero weight within an otherwise selected complete parent |
| All eligibility zero or calibration | Truth retained | Entire parent withheld |

`Class()` returns structural summaries: `known_none`, `known_one`, `known_many`, `known_all`, `unknown_containing`, or `ambiguous`. F/F/U cannot become `known_none`. All-positive parents are retained. These summaries describe supplied labels, not verified semantic truth. An opaque evidence digest identifies an external record; this reader does not interpret its predicates, certify rights, discover ancestry, or establish qualification.

## Use

Freeze `CohortBinding` separately from reviewed source, finite observations, mask policy and whole-component role records. Include the exact row SHA-256 and expected metadata. Deriving expected hashes from a row after reading it is not external binding verification.

```go
// frozen is a shortclaimdata.CohortBinding fixed before this invocation.
example, err := shortclaimdata.LoadCohortRow(rowReader, frozen)
if err != nil { return err }
input := example.Input()         // Only request/candidate text is normalized.
audit := example.Supervision() // Fixed arrays in candidate order; U has LabelKnown=false.
metadata := example.Metadata() // Role/provenance metadata; never feature text.
```

The caller must invoke this API explicitly. It registers no Codex integration, opens no path, and runs no router, inference, feature extraction or training. Accessors return value copies. Returned examples retain immutable strings and fixed arrays, with no shared mutable map, slice or lock. For JSONL, supply one independently bounded row at a time.

## Wire contract

One JSON object is limited to 16 KiB. Text reuses the existing limits of 512 bytes and 32 normalized words per request/candidate. Candidate count is exactly five or eight; input order is retained. Every field is mandatory. Duplicate/escaped duplicate keys, case aliases, unknown fields, invalid Unicode and trailing JSON are rejected. The row digest binds exact bytes including whitespace. Errors contain no supplied text or local path.

- Twelve top-level fields: `schema`, `stable_id`, `request`, `candidates`, `role`, `whole_group`, `source_family`, `source_revision`, `text_revision`, `feature_policy`, `interpretation`, `bindings`.
- `schema`: `riido-shortclaim-cohort-row-v1`; `feature_policy`: `request_and_candidates_text_only`.
- `interpretation`: `unambiguous` or `ambiguous`.
- Each candidate: `metadata_id`, `text`, `state` (T/F/U), `label` (true/false/null), `loss_weight` (integer literal 0/1), `evaluation_eligible` (boolean).
- `bindings`: `source_sha256`, `evidence_sha256`, `mask_policy_sha256`, `roles_groups_sha256`. Each is 64 lowercase hex characters; zero digests are rejected.
- Roles: `development_train`, `development_validation`, `development_calibration`. Whole group is 1–65,535. Source revision is 40 lowercase hex characters.

Train uses `loss_weight` for gradient participation and has evaluation eligibility false. Validation has loss zero and uses `evaluation_eligible` for validation-loss participation. Calibration has both disabled. U and ambiguity also disable both, without deleting known T/F labels. This contract is separately versioned from the earlier Golden Reader/mask source proposals; their files and pins remain unchanged.

## Existing projection bridge

Maintainer API `internal/claimfit.ProjectCohort` accepts at most 60 rows. **Before selection**, it checks every row for duplicate IDs, declared group/family role conflicts, and common role-plan/mask-policy digests. Withheld rows cannot hide a declared role leak. This verifies known declarations, not the transitive ancestry graph.

Only fully known, unambiguous, non-calibration parents with at least one eligible row reach existing `claimfit.Project`. U is never removed or changed to false. Train loss masks and validation evaluation masks become the existing role-specific sample weights. Selected parents retain every candidate and truthful labels at zero weight. All-positive and all-negative parents retain BCE rows; pair-loss definedness follows the existing pairlearn contract.

`Audit` retains every original example. `SelectedAuditIndices` maps projected parent indexes back to audit indexes. Zero selected parents is a valid audit-only result, not training readiness. Positive train/validation weights, source duplication, role separation, rights and semantic qualification require separate checks. This API calls neither `Fit` nor `NLL` on empty data.

The next step is a separate 20-parent cohort balancing candidate counts and zero/one/multiple positives, unknowns and ambiguity, followed by observation and whole-family review before expansion. This change demonstrates no new independent requests, model quality, Codex savings, GPU execution or protected2,400 evaluation result.
