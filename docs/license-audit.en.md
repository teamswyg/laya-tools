# License audit and redistribution corrections

[한국어](license-audit.ko.md) · English

Reviewed 2026-09-30: pinned models, adapted code, Go dependencies linked into the CLI, and the separately downloaded native runtime. **Modification, commercial use, and redistribution are feasible under the identified Apache-2.0/MIT/BSD-3-Clause terms**, with notice obligations. This review does not guarantee rights to every training datum or patent.

## Sources and scope

| Component | Evidence | Distribution treatment |
|---|---|---|
| Laya base | [Pinned card](https://huggingface.co/convaiinnovations/laya/blob/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851/README.md): `license: apache-2.0` | No root LICENSE/NOTICE at that revision. Include the declared Apache-2.0 text, attribution, and conversion notices |
| laya-code | [Pinned NOTICE](https://huggingface.co/tindang/laya-code/blob/25f97e5a2ec5f8cf7218a4f67504367d8832e1fe/NOTICE), LICENSE and card at the same revision | Apache-2.0; preserve Tin Dang/Convai Innovations/ModernBERT attribution verbatim |
| ModernBERT-large | [Card](https://huggingface.co/answerdotai/ModernBERT-large) and laya-code NOTICE | Retain encoder/tokenizer attribution; declaration does not independently establish training rights |
| Laya SDK/exporter | [Pinned LICENSE](https://github.com/NandhaKishorM/laya/blob/6d942c92081fbc139e736bbd9ac0023223c29b7f/LICENSE) | Apache-2.0; exporter and sequence-formatting attribution in NOTICE |
| system-one-router | [Pinned LICENSE](https://github.com/mmornati/system-one-router/blob/a437d00bca33a4c10038b37a5efe04dc0e3d40bb/LICENSE) | Apache-2.0; identify adapted structure and modifications |
| pi-pignon | [Pinned LICENSE](https://github.com/siiick/pi-pignon/blob/4d97d1a35134b813bac6a8e345a1bf9b0bbf0bdb/LICENSE) | MIT; retain Nicolas Chaintron's copyright and permission text after the Go port |
| regexp2 v1.11.5 | [LICENSE](https://github.com/dlclark/regexp2/blob/v1.11.5/LICENSE) | MIT, Doug Clark; bundle full text with binaries |
| onnxruntime_go v1.36.0 | [LICENSE](https://github.com/yalue/onnxruntime_go/blob/v1.36.0/LICENSE) | MIT, Nathan Otterness; also preserve Microsoft C-header MIT attribution |
| golang.org/x/text v0.25.0 | [LICENSE](https://github.com/golang/text/blob/v0.25.0/LICENSE), PATENTS | Include BSD-3-Clause and additional patent grant |
| Go runtime | Actual Go 1.27.1 LICENSE/PATENTS | Include BSD-3-Clause and patent grant |
| ONNX Runtime 1.30.0 | [Official LICENSE](https://github.com/microsoft/onnxruntime/blob/v1.30.0/LICENSE), distribution ThirdPartyNotices.txt | Setup already verifies/installs both; also bundle MIT text for linked C-header attribution |

laya-codex/keel remain references with no implementation copied. fast-laya-compaction was not ported because its license was not established. A directory's open-source description alone is not permission to copy.

## Obligations

[Apache-2.0](https://www.apache.org/licenses/LICENSE-2.0) §4 requires a license copy, prominent modification notices, relevant original attribution, and preservation of required NOTICE attribution. Our Apache-2.0 license does not replace upstream notices. §6 grants no general trademark license. We use `riidolaya` and attribution without suggesting official affiliation or endorsement.

[MIT](https://opensource.org/license/mit) allows modification, sale, and redistribution with copyright and permission text in copies or substantial portions. Translating TypeScript to Go does not remove pi-pignon attribution obligations. Headers identify the origin and distributions contain the full text.

BSD-3-Clause requires notices/disclaimers in source and binary distributions and prohibits unauthorized endorsement using authors' names. A source URL alone should not replace the required binary-distribution materials.

These are not strong-copyleft obligations requiring all original code to adopt that license. No GPL/AGPL component was found in the current linked inventory; this does not pre-approve future dependencies.

## Findings and corrections

1. **The models-v1 code bundle omitted the upstream NOTICE.** models-v2 preserves it verbatim and adds base attribution; LICENSE/card alone were not treated as sufficient.
2. ONNX `doc_string` identifies conversion/modification, with MODIFICATIONS.md and PROVENANCE.json alongside it. Serialized graph hashes verify the graph is unchanged. New URLs/checksums preserve old archive immutability.
3. **Early CLI bundles linked to some dependency sources without including full notices.** Current binaries include `licenses/`. Older releases received supplemental notices and guidance to update; this does not erase the historical state of those distributions.
4. `internal/compliance` checks linked module versions against the reviewed inventory, notice hashes, Go distribution notices, and required model notice files. It does not automatically approve new licenses or infer compatibility.

See [NOTICE](../NOTICE), [license files](../licenses), [inventory](../licenses/manifest.json), and the pinned [artifact manifest](../internal/assets/manifest.json). Go/module texts were copied from the actual corresponding distributions.

## Remaining limits

**A public license declaration is separate from verification of training-data rights.** The laya-code card also acknowledges uncertainty around source-code training and weights. We have not independently established the base model's entire training-data rights chain or the licensor's authority. Being a classifier rather than a text generator does not establish that every rights issue disappears.

Fine-tuning on Riido private code, user requests, or customer data needs a separate decision about rights, purpose, and disclosure. This preview publishes original synthetic fixtures only; no private repositories/customer data were trained on or redistributed.

This is a technical distribution review of specified versions, not a warranty concerning contracts, trademarks, training data, or patents. Commercial commitments requiring rights warranties or indemnification need legal review of those additional obligations.
