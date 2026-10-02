---
license: apache-2.0
language:
  - en
  - ko
pretty_name: Riidolaya Next60 finite development subset
tags:
  - riidolaya
  - development-only
  - finite-code-behavior
configs:
  - config_name: development
    data_files:
      - split: train
        path: data/train.jsonl
---

# Riidolaya Next60 finite development subset

[한국어](README.ko.md) · Proposed destination: `JooYoon/riidolaya-shortclaim-next60-development`.

This is an **unpublished preparation**, not a held-out benchmark or a trained model. `data/train.jsonl` contains exactly **2 request rows and 5 candidate labels**: 2 positive and 3 negative. It transcribes two existing Next60 draft IDs with their finite, source-informed development qualification. It establishes **0 new globally unique requests and 0 new source families** relative to that catalog. Do not treat it as independent held-out coverage.

| Stable request ID | Connected whole group | Candidates | Frozen inputs | Recorded candidate observations |
|---|---:|---:|---:|---:|
| next60-pflag-native-ipnet | 77: cobra/pflag | 3 | 5 IPv4 fixtures | 15 |
| next60-mapstructure-native-or | 78: mapstructure | 2 | 4 integer-hook fixtures | 8 |

The 9 fixtures and 23 saved observations are evidence scopes; they are **not 9 or 23 dataset rows**. Each candidate keeps sample weight1. Repeated vectors do not increase its training weight. Every row is `development_train`; groups77/78 are this project's connected-source identifiers, not a universal taxonomy. Existing source/fork/alias/template relations stay together. Cross-role conflicts must block a later fit, not be resolved by moving roles or searching seeds.

The Root qualification was AI-assisted and nonblind: sources and saved finite results informed captions before any model scoring or fitting of these texts. A positive label requires all frozen fixtures satisfied with no unknowns; a negative label is a witnessed mismatch within the complete scoped caption. This is not all-input correctness, arbitrary hook/object immutability, independent native reproduction, unseen-family generalization, or model utility. The exact authoring and label decisions are retained in [qualification](evidence/QUALIFIED-TRAIN-SUBSET.v1.json); earlier conditional proposals remain earlier proposals.

## How to use the rows

Encode only `request` and `candidates[].text`. `metadata_id`, stable IDs, labels, weights, roles, groups, revisions, scopes and hashes are supervision or bookkeeping, **never model features**. API names appear only as metadata IDs, outside the text. Preserve candidate order. The Boolean label belongs to each candidate, not to the whole request; five eligible candidates carry unit BCE weights. This package makes no pair-loss policy.

The expanded HF row is **not directly compatible with shortclaim.LoadValidated**: its extra fields need an explicit adapter. The exact original two-input payload is [INPUTS.v3.json](evidence/INPUTS.v3.json). The inert [Go conversion source](source/convert.go.txt) reads byte-pinned Inputs and qualification; it does not call a model or invent labels. Its future execution in a fresh Root-owned directory must reproduce the two-line draft byte-for-byte. It has not been compiled or executed here. JSON parsing, converter reproducibility, training-reader projection and the proposed viewer YAML still need Root validation. The viewer config intentionally selects only `data/train.jsonl`, excluding evidence JSON as rows; Hub acceptance has not been tested.

No new training occurred on these two requests. There was no pre-30/60 checkpoint training, combined81 projection/fit, protected validation/calibration/final read, LoRA, ternary or MPS training, production activation, or cost-saving measurement in this package preparation. The existing79 corpus and previous model snapshots remain unchanged/inactive. This package contains no model weights, binary, raw upstream bodies, copied upstream prose, private host paths, raw traces or real-user prompts.

## License scope and provenance

The YAML `apache-2.0` and exact project [LICENSE](LICENSE) identify the repository license for our own prose and finite annotations. They do **not** relicense included upstream notices, all source ancestry, ancestor models or model weights. The five complete [notice assets](evidence/NOTICE-MANIFEST.v1.json) retain pflag BSD-3-Clause, mapstructure MIT, and separate Go BSD notices. Raw upstream bodies are excluded; unresolved ignored Go2022 join ancestry remains unresolved and excluded. The prior [conditional rights review](evidence/rights-review/CONDITIONAL-ELIGIBILITY.v1.json) and [material scope](evidence/MATERIAL-SCOPE.v1.json) preserve their historical limits; this is not blanket legal clearance.

Sources: pflag revision `c966cfef47379dcb01e7929504d66d94b540945b`; mapstructure revision `52aa5c6dc1d27226460807054ca2107b2d54fb2d`. Local public source head for preparation: `acf514d3dd25db2e199e43da80ccf4477d8ff310`, [PR110](https://github.com/teamswyg/laya-tools/pull/110). Its CI/merge are **pending at preparation**, not asserted complete. [PR109](https://github.com/teamswyg/laya-tools/pull/109) records the earlier native observation publication, with its separate retained proof. Root must verify exact-head required CI and committed source pins, final public safety/license scope and JSONL/viewer metadata before a separate publication plan can create or upload this dataset. This preparation performed HF repo/auth/API/HTTP/upload/create/token reads0.
