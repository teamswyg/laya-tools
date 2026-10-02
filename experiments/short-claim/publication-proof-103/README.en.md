# Preparation publication verification

[PR103](https://github.com/teamswyg/laya-tools/pull/103) merged automatically after all four required checks passed on head `6a68527e6d08ba5ee517faf5dd472c1392ec59a0`. Its actual merge `01ce47dcca16c8f82022c97354d82105b6819875` has the same complete Git tree. The [CI run](https://github.com/teamswyg/laya-tools/actions/runs/37015530233) passed Linux/macOS tests, secrets and quality.

Afterwards, four Wiki files totaling14,115 bytes—Home, Sidebar and the Korean/English training-data preparation guides—were copied from reviewed Git blobs. Following the secret scan, they were published and actual Wiki commit `83cf5e1eb5500861b4cc4bde5d6826b0d02c7b4b` was fetched to verify all four remote files byte for byte.

- [Merge status](CI-MERGE-OFFICIAL.v1.json) and [four actual CI jobs](CI-JOBS-OFFICIAL.v1.json).
- [Remote publication verification](WIKI-PUBLICATION-LEDGER.json) and [local copy verification](WIKI-COPY-LEDGER.json).
- The [preparation handoff](HANDOFF.v1.json) preserves its earlier zero-execution and pending-publication state.

This verifies publication. It does not replace function/candidate observations, fitting or model utility tests. The publisher helper with host paths and its binary remain outside Git.
