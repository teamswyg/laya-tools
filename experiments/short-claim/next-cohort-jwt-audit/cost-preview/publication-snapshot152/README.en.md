# Publication 152 file snapshot

This directory retains five shared repository files, byte for byte, from JWT cache cost publication commit [`1f6b84c`](https://github.com/teamswyg/laya-tools/tree/1f6b84c91e1931001f1ffa6dddc95e9c9c9ac092). It preserves the original checksum evidence when later work changes the README or CI.

`files/` contains the two root READMEs, CI configuration, security scanner configuration, and verification script from that commit. The archived configurations and script are data, not active configuration or executable verification. Original copyright and attribution text is unchanged. The repository's Apache-2.0 license applies.

`ORIGINAL-SHA256SUMS.publication152.txt` is byte-identical to the existing 45-row experiment manifest. `SHA256SUMS` checks only the five snapshots. The current verifier preserves the original manifest and resolves exactly these five rows to snapshots; the other 40 rows still check current experiment and replay code. Additional external paths are rejected.

The current `scripts/verify-jwt-cost-preview.sh` is a new verifier with this resolution step. `files/scripts/verify-jwt-cost-preview.sh` retains the original verifier. Current shared documentation, CI, and security changes require current repository checks; this snapshot does not approve them. It changes neither the first failed CI record nor experimental results. Verification needs no Git history or network access.
