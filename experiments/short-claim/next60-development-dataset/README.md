---
license: apache-2.0
language:
  - en
pretty_name: Riidolaya Next60 finite development subset
tags:
  - riidolaya
  - development-only
  - finite-code-behavior
configs:
  - config_name: development
    default: true
    data_files:
      - split: train
        path: data/train.jsonl
---

# Riidolaya Next60 finite development subset

[한국어 안내](README.ko.md)

This **development data release candidate** contains **2 requests and 5 candidate labels**: 2 positive and 3 negative. It helps develop a small claim scorer that checks which code-behavior description satisfies a request within a stated, finite scope. It is not a trained model or an independent evaluation benchmark. Publication is still pending the separate dataset CI, final checks and a Root publication plan.

The data rows are **English only**. A Korean guide is provided for people; its presence does not make the training rows bilingual.

| Request | Whole source group | Candidate labels | Evidence scope |
|---|---:|---:|---|
| next60-pflag-native-ipnet | 77: connected cobra/pflag | 3 | 5 frozen IPv4 inputs, 15 saved observations |
| next60-mapstructure-native-or | 78: mapstructure | 2 | 4 frozen integer-hook configurations, 8 saved observations |

The first request checks masked CIDR networks, error-state preservation and rejection of bare addresses. The second checks retrying hooks on unchanged input, stopping even at successful nil output, and collecting failed messages with line feeds. The **9 fixtures and 23 observations are evidence, not extra dataset rows**. Candidate order is preserved and each candidate has sample weight1; repeated observations do not increase its weight.

These are two existing Next60 draft IDs. They add **0 new globally unique request IDs and 0 new source families** relative to that catalog. Both belong only to `development_train`, with connected source groups77/78 kept whole. Captions and labels were AI-assisted, source/observation-informed and nonblind. A positive label requires all frozen cases satisfied with no unknowns; a negative label records a scoped mismatch. This does not establish all-input correctness, unseen-family generalization or an independent held-out result.

## Use the text, keep metadata separate

Read `data/train.jsonl` as one row per request. For model features, use **only `request` and `candidates[].text`**. Candidate Boolean labels and unit weights are supervision. `metadata_id`, stable IDs, role, group, source revision, scope and hashes are bookkeeping and must not become model features. API names stay in metadata IDs, outside the text. Do not feed the serialized whole row to an encoder.

The expanded row needs an explicit adapter; it is not directly accepted by `shortclaim.LoadValidated`. Root reconstructed the two original owned inputs from request text, candidate text and metadata IDs, found semantic equality with the pinned original inputs, and passed **LoadValidated2 + Validate2, Fit0**. The original [two-input payload](evidence/INPUTS.v3.json) and [exact validation record](evidence/INPUT-VALIDATION.v1.json) are included. This validates the input representation, not model quality or the future training-reader integration.

Root also executed the pinned Go converter **once**. Its output matched the two-line draft byte-for-byte: 2,597 bytes, SHA256 `843e4776823fe44ad14f0c7a8b6bdf29827d92596ef3d8bf1dd31d498e87ca1c`. The [exact Root validation receipt](evidence/ROOT-VALIDATION.v1.json) records Root's actual work; this package author's execution remains0. The [converter source](source/convert.go.txt) is retained as text, not a binary.

The viewer metadata selects only `data/train.jsonl` as `development/train`; it does not treat evidence JSON as rows. Its shape was checked against the official [dataset-card guide](https://huggingface.co/docs/hub/datasets-cards) and [repository-structure guide](https://huggingface.co/docs/datasets/repository_structure). Actual Hub/viewer acceptance remains pending.

## Provenance, limits and license

Root verified PR110's four required CI jobs, merge `e8382e83cb1ac2ec5af42d62c0936a548f53166d` and identical tree `f331956e35acf0556818caa021f097f16e00b24c`. The six wiki pages were verified at commit `832f44ffea58dfd7afd2b2168cffa3cf1e7f8ae5`; see the [exact retained receipt](evidence/PR110-CI-WIKI-PUBLICATION.v1.json). This is data/source publication evidence, not a model-quality result. Earlier pending cards are kept unchanged under `preparation/`; the original preparation provenance remains a historical snapshot.

No new Fit, combined81 training, pre-30/60 checkpoint training, paid model call or protected evaluation was performed for this release. Existing79 data and previous inactive models remain unchanged. This package claims no LoRA, ternary, MPS, GPU, production-readiness or Codex cost-saving result, and no completed 2,400-example collection. Cross-role source-family conflicts must block later training, not be fixed by moving roles or searching seeds.

The YAML `apache-2.0` and project [LICENSE](LICENSE) cover our own prose and finite annotations under the repository license. They do **not** relicense upstream notices, all source ancestry, ancestor models or weights. The [five complete notices](evidence/NOTICE-MANIFEST.v1.json) preserve pflag BSD-3-Clause, mapstructure MIT and separate Go BSD notices. Raw upstream bodies, private paths, binaries, model weights, raw traces and real-user prompts are excluded. The unresolved ignored Go2022 join ancestry stays unresolved/excluded; prior [conditional rights](evidence/rights-review/CONDITIONAL-ELIGIBILITY.v1.json) and [material-scope](evidence/MATERIAL-SCOPE.v1.json) records retain their limits.

Proposed destination: `JooYoon/riidolaya-shortclaim-next60-development`, immutable tag `next60-2-finite-v1`. [RootPublicationPlan](RootPublicationPlan.json) keeps publication blocked until exact dataset CI111, final safety/rights checks and Root's separate authorization. This package author performed Go/helper/model/HF/auth/API/HTTP/upload/create/token reads0.
