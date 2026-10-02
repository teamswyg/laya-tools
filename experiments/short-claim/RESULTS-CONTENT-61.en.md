# Content review61: what can support learning?

Content readings are complete for the existing **72 requests and216 candidate positions**. Review60 covered123 original entries; review61 covers the remaining615. These738 entries are several views of the same corpus, not738 independent requests or model trials. Original truth, candidate order,21 unknown parents and17 total/16 known-containing groups are preserved.

| Original review entries |60|61|Combined|
|---|---:|---:|---:|
| Consistent with scoped evidence |78|446|524|
| Contradicts scoped evidence |17|94|111|
| Omits required scope |2|10|12|
| Unsupported or uncertain |26|65|91|
| Total |123|615|738|

Consistency here is not a candidate approval rate. The table combines source references, request wording, implementation fidelity, task coverage and contract observation. Looking only at caption fidelity,207 of216 positions agree with their own implementation and9 remain uncertain. Task coverage instead has35 consistent/111 contradictory/9 omitted/61 uncertain positions. A faithful description of incorrect code is possible, so fidelity and correctness must remain separate.

For a hint tool, this distinction has a practical consequence. A request to “remove adjacent duplicates” may meet a caption saying “remove all duplicates.” That caption can faithfully describe an incorrect implementation and potentially support the existing negative label. A phrase such as “zero if invalid,” however, leaves the meaning of invalid unclear. An omission or uncertainty verdict cannot become a new binary truth label. A frozen supervision scope must record what is withheld.

Review61 reads15 families,60 requests and180 candidate positions. It neither reruns nor overwrites review60's three families. Observation fields and negative boundaries add360 supplementary records, making975 records at61. Together60 and61 have1170 records but still72 requests. See [61 detailed findings](CONTENT-REVIEW-FINDINGS-61.en.md), [60 detailed findings](CONTENT-REVIEW-FINDINGS-60.en.md) and the [61 manifest](content-review-manifest-61.json) for evidence and competing interpretations.

The original59 file with738 `pending` entries remains an immutable historical snapshot. Read the new60/61 overlays alongside it; do not rewrite the old file as passed. The61 manifest's pending independent verification also describes its production-time state. A subsequent [separate verification receipt](content-review-mechanics-61.json) and [ledger](content-review-mechanics-ledger-61.json) supplement it. The independent checker ran twice: a legacy-null handling assumption failed once, then the correction passed. Assessment modifications/regenerations remained0. Checks matched31 JSON pins,975 record digests,984 quote ranges and the615+360 entry bijections. The [independent explanation](QA-FINDINGS-61.en.md) refers to `report.json`/`ledger.json`, archived publicly under the receipt/ledger names linked here. Matching hashes, quotes and references establishes integrity, not objective semantic correctness or model quality.

The readings were AI-assisted within this Codex collaboration. The reader differs from the original caption/fixture/CLI author but has prior review and outcome-metadata exposure; this is neither blind evaluation nor a different data-authoring origin. `model_calls=0` counts no separately launched judge inference/API/benchmark processes. It does not mean zero AI usage or cost for the collaboration; ordinary collaboration cost was not measured. Generation, build and checker-helper failures and corrections remain in separate execution ledgers.

The next Go path will keep original text/truth, supervision eligibility and actual fit inputs separate. It must preserve every acceptable candidate, retain established negatives for no-answer requests and generate no training labels for unknowns. Related source/helper derivatives follow one component role. Withheld supervision never deletes a candidate from ranking or limits fallback.

Complete readings alone do not establish training readiness. Actual source diversity, frozen roles/seed/coverage and a bounded loader/execution plan remain. This review assigns0 roles and produces0 fits, weights or protected-final reads; `training_ready=false`. Existing15-group/5% utility conditions and the2400-request-per-domain protected-final target remain. No new15-repository condition or2400-request prerequisite for every development fit is introduced.
