# Packed weights module publication verification

[PR101](https://github.com/teamswyg/laya-tools/pull/101) passed Linux and macOS tests, the secret scan and the quality gate, then merged automatically. The tested head and the merge commit have identical complete trees. The [actual CI response](riido101-run-success-proof.json) and [merge response](riido101-ci-final.json) are retained. No human approval or protection bypass was used.

The public module is a Go library that reads compressed weights and scores a small hint model. **Its size does not describe the complete Laya model.** Tests comparing its calculations with the existing formats passed. The public module itself has had zero benchmark executions; measurements from a separate private prototype remain distinct.

After verifying required CI success and the merge, four Wiki files containing Korean and English guidance were published, totaling 10,638 bytes. Fetching remote Wiki commit `cac8e0d727eea36d5b7608977ea863904e7d474b` and checking every file's bytes and hashes passed in one attempt. The [preparation ledger](riido101-wiki-copy-ledger.json) retains its historical pending state; the [publication ledger](riido101-wiki-publication-ledger.json) records completion.

[Korean Wiki](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-KO) · [English Wiki](https://github.com/teamswyg/laya-tools/wiki/Packed-Weights-78-EN) · [Exact copy ledger](PUBLIC-COPY-LEDGER.v1.json) · [한국어](README.ko.md)

This publication verification reran no model, original API, training or benchmark and made no Hugging Face upload. Four public responses retain their exact original 14,452 bytes. Local paths, binaries, private logs and publication helpers were excluded.
