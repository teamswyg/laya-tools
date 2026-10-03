# Actual publication and readback of30 rows

The [immutable HF30 release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1) was published at commit **89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d**. All **558 owned files/4,559,126 bytes** were downloaded from that commit and matched the local publication bytes. The one Hub-managed .gitattributes file/2,504 bytes is unchanged from23. Previous2,3,7,16,21,23 tag refs did not move.

See the [file/tag/viewer receipt](HF30-PUBLICATION.actual.public.v3.json) and [viewer acceptance](HF30-VIEWER-ACCEPTANCE.actual.public.v1.json). The viewer returned **30 rows**, every field and row order matched, num_rows_total=30, partial=false, and zero truncated cells. Commit and top-level truncated fields are absent and remain null. Current viewer correspondence is separate from immutable-commit all-files verification.

[PR121](https://github.com/teamswyg/laya-tools/pull/121) passed all four required Linux/macOS/secrets/quality checks in [actual CI](https://github.com/teamswyg/laya-tools/actions/runs/37103825806) and was merged by the CI bot. The [source/merge-tree and job receipt](CI-MERGE121.actual.public.v1.json) distinguishes the successful new30-row offline reproduction step from existing native Laya checks. GPU execution was not verified.

The first three viewer requests returned HTTP500; attempt4 returned200 and passed value checks. Failed local assumptions about an absent top-level truncated field and a stale viewer filename are preserved. A fresh verifier using the actual schema and accepted response exited0. No original source or model reran to correct these local checks.

The data contains **30 requests/88 labels(30 positive,58 negative)**, all development_train. Read the [usage guide](../next60-development-thirty/README.en.md) and [22 proposed new-source contracts](../next60-new-source-preview/README.en.md). New claim-model/Fit calls for this data expansion are zero; the historical3 Fits/3 logical models and failed-model inactive state are unchanged. Thirty is a collection checkpoint, not proof of training readiness, generalization, Codex savings or model/GPU performance.

Public files contain owned prose, predicate summaries and checksums. Credentials, private host paths/inputs, upstream runtime bodies, model weights and raw journals are excluded. [한국어](README.ko.md)
