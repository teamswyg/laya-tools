# Verified merges and Wiki publication

[PR97](https://github.com/teamswyg/laya-tools/pull/97) passed Linux, macOS, secrets and quality at exact head `ad61cd1a8b95434e2f2ed106aa154f3dcbe44621`, then automatically merged as `2c51606ba6da7442fd4db3d9667daa8c573c2b83`. Preserve the [CI](CI-97.v1.json) and [actual run](RUN-97.v1.json) records.

After merge, the bilingual Source-Observation73 and Compact-Storage73 pages plus Home and Sidebar—6 files/59,678B—were published to Wiki. A fresh fetch of Wiki commit `152ff088d6655bf0084aed07b9422db1c6ac81c3` matched the CI-head documents byte-for-byte. The [preparation ledger](WIKI-COPY-97.v1.json) retains its historical remote-pending state; the [actual publication/readback record](WIKI-PUBLICATION-97.v1.json) is separate. Earlier PR94/96 Wiki records remain preserved.

[PR98](https://github.com/teamswyg/laya-tools/pull/98) also passed all4 required checks at head `6e4a6d84008a9200fedf51b9bfd7d36834245c27` and automatically merged as `7cea49090727555924bf22679ffeccc4f38c9865`. Preserve [CI](CI-98.v1.json), [run](RUN-98.v1.json) and root's full-tree equality check. Its FNV prefix change preserves feature and score bits while reducing observed calculation time on local public fixtures; allocations are unchanged. This does not demonstrate model utility, RSS, GPU or real LLM cost improvements.

Publication verification did not repeat original behavioral observations, fitting, benchmarks or HF uploads. Model bytes remain in the earlier HF research archives and outside Git.

[한국어](README.ko.md)
