# PR93, Wiki and Hugging Face publication verification

[PR93](https://github.com/teamswyg/laya-tools/pull/93) passed Linux/macOS, secrets and quality at exact head `9c0eca98ac9e3f40b0eadafb21422a349fb732d8`. Squash merge `9d204c2c700505658c108297d5fa769a835a60c3` was confirmed on 2026-10-02. Historical pending records remain preserved; [final CI status](CI-PR93.v1.json) is appended separately.

Four files from that head—Korean and English explanations, Home and Sidebar—were copied to Wiki. After push/fetch, remote Git contents matched the source bytes. The [Wiki proof](WIKI-PUBLICATION-93.v1.json) records its commit and file hashes.

The [Hugging Face dataset](https://huggingface.co/datasets/JooYoon/riidolaya-public-claim-preparation-69/tree/d80075c6160a53d5426019cf018016a3b02017bd) is the **original72 preparation snapshot**. Twenty project files, 7,960,062 B, were downloaded again at the exact commit and matched byte for byte. [Publication proof](HF-PUBLICATION-69.v1.json) and [collection/tag proof](HF-COLLECTION-69.v1.json) distinguish staging, publication and readback. Fourteen existing collection items were preserved; one new dataset reference was added.

This archive has no fitted coefficients and reports zero fits. Its 72 rows and 216 candidates are not 2,400 independent tasks or a protected final set. Project-owned public fixtures and derivatives use Apache-2.0; upstream implementation and pretrained weights were neither redistributed nor relicensed. Publication does not prove Laya-encoder/GPU execution or LLM savings. Later 76-request preparation and fitting do not overwrite this snapshot.

Historical HF-stage v1 review and latest-card delta remain in the [review folder](hf-stage-qa/STAGE-QA-69.en.md). Previous card bytes were unavailable, so a complete byte diff was not proved. One tag-check helper failure from an incorrect target assumption remains recorded; dataset info resolved at the tag confirms the actual upload commit.

[한국어](README.ko.md)
