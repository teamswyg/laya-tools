# 59: A plan to bind captions to original text and literal source locations

[한국어](PLAN-REFERENCES-59.ko.md) · [Preparation after 58](NEXT-STORED-58.en.md) · [58 results](RESULTS-STORED-58.en.md)

**This phase binds the original text and evidence locations before judging whether a caption is faithful.** Read the existing 72 public requests and 216 candidate captions unchanged. Record the JSON locations, UTF-8 byte counts and hashes for 288 texts, and the literal source expression locations for 18 contracts. This creates no new requests, labels, model results or savings evidence.

The 51 requests with known outcomes in 58 contain 34 answerable and 17 no-answer requests. All 21 unknowns remain in the full denominator of 72. Preserve the original 2-, 3- and 4-candidate layouts, ordering, captions, truth and 17 groups. There are 16 groups containing known parents; quoted-delimiters group ID 64 has 4 unknown parents. IDs, hashes, groups, truth and review status are audit metadata, never scorer features.

## One metadata generation

`riido-captionref prepare` prepares source, input, support and execution-plan records. This mode makes 0 Bind and metadata-generation calls and checks 20 Git blobs. After freezing source, inputs, plan and binary, use `riido-captionref references` for one official metadata generation in a new output directory's `results.json`. The execution plan has the fixed path `experiments/short-claim/execution-plan-59.json`. Plan 1 official attempt and 0 retries. A failure preserves its partial record, fixed failure code and stage counters. Do not silently repeat that official attempt.

The generator binds the 3 frozen public raw files with `storedaudit.Bind` and reads Go ASTs to locate individual legacy contract elements and typed literal-array expressions. It does not execute original candidate functions or their package registries. A raw expression hash differs from a serialized literal-table hash. Locating an expression does not reify Input/Want values or observe per-vector candidate Got values and failures.

| Check | Frozen scope |
|---|---|
| Compiled sources | 5: CLI main, captionref references/provenance, storedaudit binding/provenance |
| Inputs | 6: the 3 original raw JSON files and the 3 literal source files data.go/state.go/flow.go |
| Execution support | 9: both plans, build recipe, oracle review, 2 CLI/generator tests, go.mod/go.sum/LICENSE |
| Official Git blob checks | 5 sources + 9 support + 6 inputs + 1 execution plan = 21 |
| Generated references | 72 parents, 216 candidate positions, 288 text references, 18 contract expressions, 3 literal source files |

Build with Go 1.27.1, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false` and `-p=1`. Inputs and results are bounded at 1 MiB; the binary is bounded at 64 MiB. Check size, regular files and safe paths before and during reads, and compare on-disk bytes/hashes with compiled sources. Each Git blob check has a 5-second budget. `GOMAXPROCS=1` and the 256 MiB Go heap soft limit are settings, not measured CPU, total RSS, GPU or latency results.

The output contains a provenance envelope, `generation_counters` and a `references` body. Success uses `references_generated_content_review_pending`; failure uses `incomplete` and a fixed `failure_code`. Reserve a new output with `O_EXCL` rather than replacing earlier evidence. A public `caption-coverage-59.json` is an unchanged copy of the original envelope. Writing this plan alone does not complete execution freezes or generate results.

After the official result, prepare a separate `frozen_test.go` for CI regression replay. Compare the entire `references` Report and `generation_counters` **exactly**, as bytes or structs, with no floating-point tolerance. Verify the original source, input, execution-plan, binary and native Darwin envelope pins separately. Linux/macOS CI binaries are not claimed to equal the historical binary. Replays add no official attempts or independent requests. Do not retroactively add the post-observation test to the original 9 support files.

## Accurate references still need content review

Keep every source-fidelity, request-contract coverage and observation-fields review `pending`; record `literal_payload_reified=false` and `training_ready=false`. Expression references alone do not approve complete source closure across functions, helpers, types and sentinels, or rebind the historical code/bundle digest recipes.

Stored truth and caption fidelity are separate. A caption can faithfully describe an incorrect implementation. The [content-review recipe](content-review-recipe-59.json) separates source fidelity, requested contract scope, observation fields, and explicit negative/boundary conditions. Independently read text spans and evidence, recording omissions, contradictions and unsupported scope. Successful hash/location checks or finding words do not approve meaning. Do not truncate difficult behavior into apparent correctness or turn unknown into incorrect/no-answer.

The [role recipe](role-recipe-59.json) remains `unassigned`. Preserve the existing whole-component floors of 9 train / 3 validation / 3 calibration groups and permit 0 transfer groups. The 16 groups containing known parents make the numeric floors feasible; actual assignments, coverage and statistical independence are not verified. Freeze future membership and seed before applying one fixed algorithm to whole components. Assign no seed or role membership now. Previously observed data from 58 cannot regain blinded status, and group scores must not select assignments. Add no new gate or requirement to place all 18 prototypes in every role.

Of the 3 outputs proposed by NEXT-STORED-58, this phase supplies the **reference-preparation part** of caption coverage. Semantic review, the 2 scoped source-transfer proposals and actual role assignments remain pending. The separate content/role recipes and prototype ledger are follow-up preparation records outside the 9 execution support files; the runtime does not read them. Preserve `no_roles_plan` and `synthetic_single_pipeline`.

## Preserve failures too

The [prototype ledger](prototype-ledger-59.json) retains 2 private preparation executions and 2 Bind calls. The first failed before output because 3 manually entered prototype names differed from the immutable originals. It produced 0 files; its executable SHA was not recorded. Correcting those names allowed the second run to create 1 reference output. This history and independent reference reviews are not official 59 attempts, semantic approval or new independent requests.

New labels, source API calls, rankings, model/paid calls, role assignments, fits, weights, protected-final reads and activation remain 0. Follow automatic checks, independent content reviews and existing CI policy without adding human approval. Stop on mismatched text/source pins, ordering or references, or exceeded bounds/unsupported scope. Preserve content gaps without changing truth, groups or criteria.

The goal of 2,400 distinct protected-final requests per domain remains separate generalization evidence. Do not add the existing 72 requests, 288 texts, 18 literal expressions or replay counts to that target, or read existing protected-final/CoSQA reserves.
