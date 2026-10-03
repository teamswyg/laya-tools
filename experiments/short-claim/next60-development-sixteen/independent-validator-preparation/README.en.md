# Preparation for validation of sixteen development rows

This package contains source for validating the future sixteen-row artifact. It checks that the previous seven-row file's 9,619 bytes remain its exact prefix and that the first nine tasks in the pinned Catalog10 are appended in order. No actual sixteen-row artifact was created or executed here. Root's separately authored adoption record is pinned at 16,030 bytes with SHA-256 `b2d5a732d7f5d9238883060a881fc0c8b21b08c5b26004a0c27ef374be0aa9d4`.

The reviewer did not author the candidates, Wanted, observer, Root adoption, or training data. The review is nonblind because it follows earlier source review and saved-result comparison. The reviewer authored this validator, so the sealed source is prepared for Root's independent review before actual use. This is independent code and evidence checking, not a new request for human approval.

`module/validate.go` is a standard-library-only core. `source/reader-main.go.txt` is a CLI bridge that calls the existing `shortclaimdata.LoadDevelopmentRow`. Copy both into one `package main` directory as `.go` files and run them together inside the laya-tools module. The pure synthetic tests did not compile the bridge or execute the project reader.

Expected invocation follows. File names are placeholders; replace DATA_SHA with the exact hash of Root's completed artifact. The CLI accepts exactly seven options once each in `--name value` form. Duplicate options, `--name=value`, and positional arguments are rejected.

```sh
GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0 \
go run source/validate.go source/reader-main.go \
  --data TRAIN_JSONL --data-sha256 DATA_SHA \
  --previous-data PREVIOUS_SEVEN_JSONL --catalog CATALOG10_V2_JSON \
  --adoption ROOT_QUALIFICATION_JSON \
  --adoption-sha256 b2d5a732d7f5d9238883060a881fc0c8b21b08c5b26004a0c27ef374be0aa9d4 \
  --finite-comparison FINITE_COMPARISON_JSON
```

Preparation used exactly Go 1.27.1 darwin/arm64; select that version explicitly for actual use too. Each input must be a regular file of at most 128 KiB, and each data row is bounded at 16 KiB. Diagnostics contain neither supplied paths nor raw bodies. stdout is a fixed-structure JSON summary: exit zero for success, one for invalid inputs, and two for invalid arguments.

Checks run in this order: file pins, seven-row byte prefix, Root adoption and comparison correspondence, sixteen row contents and metadata, project-reader loads, and final counts. Each row has the same eleven exact fields as the reader. Candidates have metadata_id/text/label/sample_weight, and finite_scope has five mandatory fields. Missing fields and null, explicit false and empty arrays, remain distinct. Duplicate or case-alias keys, invalid Unicode escapes, edited request/caption text, non-unit weights, changed groups/revisions/roles, and expanded finite guarantees are rejected.

New stable IDs are `next60-` plus the catalog ID. Candidate metadata IDs follow the catalog order original/authored_reference/authored_negative. Actual labels must match false/true/false in Root's frozen adoption. A reference positive requires all fixed inputs satisfied and no unknowns. A negative requires a known counterexample; any concurrent unknown positions must remain identical to the saved comparison. Unknown alone never creates a negative label. Source families retain existing group77 pflag connectivity and group78 mapstructure; no new family or evaluation role is created.

Expected success totals are sixteen requests, forty-seven candidate labels (sixteen positive and thirty-one negative), eighty fixed inputs, and two hundred thirty-six candidate observations. The project reader completes sixteen calls, with zero feature, score, projection, Fit, model, or original-candidate calls. The validator's own labels_assigned and qualified fields stay false. It checks faithful materialization of Root's existing finite adoption; it does not reassess semantics or grant training authority.

One standard-library synthetic test run passed nine top-level tests and thirty-one subtests, with zero failures, skips, or stderr bytes. One vet run passed. Fake records and stub-reader calls are not actual input validation, project-reader execution, or model-quality evidence. A control also rejects fake records at production pin checks. Actual sixteen-row validation remains at zero runs.

Training interpretation needs separate controls. All nine new positives occupy candidate index one, and fourteen of the sixteen combined rows also do. A constant index-one choice therefore achieves 14/16, or 87.5%, on this fixed ordering; that is arithmetic, not a measured model score. Captions also contain style cues: reference descriptions restate the desired request, while negatives often use words such as `but`, `losing`, or `exposes`. Later evaluation should use separately pinned candidate permutations and compare fixed-order, majority-label, and lexical baselines. Preserve the original training rows. Neither those controls nor training were executed here.

[The plan](PLAN.v1.json) and [position/style leakage review](LEAKAGE-REVIEW.v1.json) describe the scope. Source-based validation establishes no model quality, cost saving, external process evidence, license guarantee, or new public CI success. No shared repository or other sealed folder was changed, and no model, protected-data, or network operation ran.
