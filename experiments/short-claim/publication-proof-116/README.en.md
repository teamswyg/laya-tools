# Verified publication of the seven-request version

[PR115](https://github.com/teamswyg/laya-tools/pull/115) passed
[all four required checks](https://github.com/teamswyg/laya-tools/actions/runs/37079477122)
and was automatically merged by GitHub Actions. The
[seven-request Hugging Face version](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-7-finite-v1)
was published and independently checked at its pinned commit.

All 112 owned files were downloaded and compared byte for byte, with an exact
owned path set. Checksums passed for the 110 payload files, excluding only the
root manifest and checksum list. Hub-managed `.gitattributes` remains unchanged
and is excluded from the owned payload. The viewer initially returned HTTP 500.
A later HTTP 200 response matched every field of all seven source JSONL rows.
The previous two- and three-row tags were preserved.

- `CI-MERGE115.v1.json`: Required checks, automatic merge, and head/merge tree comparison.
- `HF-PUBLICATION.v3.json`: Actual upload, pinned download, new tag, and viewer verification.
- `HF-FILE-MANIFEST.v3.json`, `HF-SHA256SUMS.v3`: Published file inventory and checksums.

The data contain seven requests, twenty labels, thirty-four fixed inputs, and
ninety-eight candidate observations. No additional model training occurred.
This publication record does not establish improved model accuracy, GPU
execution, or lower Codex usage. CI115 covers its source; later Hub cards and
publication metadata have separate verification records.
