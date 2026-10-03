# Finite development data for a small claim and hint model: verified 33 requests

The target is a tiny **claim and hint model** that frequently suggests useful code candidates at low cost. Tests and counterexamples support the hint. This package builds development supervision; it does not report a new model's accuracy or Codex savings.

**Owned Go generated33 requests and96 labels; the actual project Reader completed33 calls,33 returns and33 value matches.** See the [materialization receipt](MATERIALIZATION.v1.json) and [Reader receipt](READER.v1.json). The previous30 rows remain byte-exact. Data:45,390 bytes, SHA-2567b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919. GitHub CI and HF33 publication are separate verification steps.

| Unit | New batch | Verified33-request pool |
|---|---:|---:|
| Semantic requests | 3 | 33 |
| Selected candidate labels | 8: positive3, negative5 | 96: positive33, negative63 |
| Fixed inputs | 14 | 160 |
| All original observations | 42 | 476 |
| Selected candidate observations | 37 | 467 |
| New Fits or model inference in this expansion | 0 | 0 |

42 executions are input variants and candidates for three requests, not42 independent tasks. The existing30 rows contain88 candidates; they are not padded to90. Of nine new candidates, one unknown-only candidate is excluded, adding eight labels.

## Three tested contracts

| Request | Observed distinction | Finite suitable position |
|---|---|---:|
| Complement within a requested width | A requested width64 with stored length65 must return length64 | First |
| Checked JSON integer extraction | Preserve exact large integer text and distinguish fractions and range errors | Second |
| Strict binary bitset decoding | Reject trailing/short bytes and tail padding, preserving the receiver | Third |

Existing APIs follow their existing contracts. We compare fulfillment of **new requested contracts**; a mismatch alone is not an upstream bug. Source metadata is for traceability. Only request and candidate text become model features.

Keep satisfied, contradicted and unknown distinct. The42 candidate-input rows contain **26 satisfied,11 contradicted,5 unknown**;318 primitive conditions contain **272 satisfied,31 contradicted,15 unknown**. Original GJSON Int has no requested error-return channel, so its five rows remain unknown and its candidate receives no label. Two binary-decoding rows combine a known receiver-mutation counterexample and an unknown error identity. The counterexample supports a negative label; the unknown condition is not converted to false.

Two of the new15 unknown conditions belong to selected candidates;13 belong to the excluded candidate. Combined with eight metadata entries **after row23** in the previous30 materialization, those scoped counts are10 and23. Earlier-prefix evidence keeps its own metadata. These are not whole-research-history totals.

## Usage for people and agents

data/train.jsonl contains one request per line. The exact previous30-row byte prefix is retained; three rows are appended. Selected candidates have known labels and unit weights. Exclusions and unknowns remain in qualification/ROOT-QUALIFICATION.v3.json and evidence/SAVED-COMPARISON.v1.json.

Owned Go sources and a verifier command are supplied for a checkout:

```sh
bash scripts/verify-next60-thirtythree.sh
```

The checks cover saved finite comparison, old-data preservation, generation, actual LoadDevelopmentRow value correspondence and authored failure controls. They execute no original candidate, Laya or claim model. The public saved-comparison check compares all42 full rows and318 primitive conditions with the original saved report. Only four opaque receipt pins are replaced by public path-free projections and excluded from that equality comparison. It does not recertify undistributed private admission receipts.

See the [previous30 usage guide](../next60-development-thirty/README.en.md) for Go Reader examples. The Reader uses fixed arrays for rows up to16KiB and eight candidates, without locks. Bitset stays in source group82; GJSON, Match and Pretty stay together in group83, all development_train. Flat metadata names bits-and-blooms-bitset and tidwall-gjson map explicitly to preserved upstream repository names.

## Measurements and next training

Original collection ran once. Whole-worker OS peak RSS was **17,612,800 bytes, about16.8MiB**, and wall time was **8.052 seconds**, including1,739 durable file/directory writes and ACKs. These are not model inference memory, GPU execution or speedup results. The stripped saved-comparer executable was **2,850,722 bytes**, which is not a model size. Go pprof and OS RSS have different measurement scopes.

The three new suitable candidates occupy different positions, but always choosing the second candidate still yields28/33 for this pool. Position, style, lexical controls and semantic duplication must be checked before training. **33 is a collection checkpoint, not an automatic Fit trigger.** Reassess whole-source-family readiness around60 requests; retain the protected2,400 requests per domain plan and5% total verification-work reduction target. Existing three Fits, three logical models and the inactive failed model remain unchanged. Ternary storage size and CPU utility are separate experiments.

Next collection preparation covers complete values per physical JSON line, preserved callback prefixes before invalid lines, and invalid UTF-8 append rejection. Existing ForEachLine already honors callback stopping; that behavior is not repackaged as a defect. Freeze new expected outputs before observation.

## Public scope and history

Owned code, descriptions and fixtures are Apache-2.0. Separate complete BSD/MIT notices for selected upstream APIs remain in evidence/notices. No original source bodies, model weights, execution binaries, raw journals, private task inputs or credentials are included. Original source/go.mod/notice hashes provide traceability. Model publication rights are separate.

history/CORRECTIONS.v1.json records the draft unknown-scope error, flat metadata correction, first comparer build's size-budget failure and subsequent verification. Failed files and old drafts remain private and recoverable. Labels have finite fixed-input support and do not guarantee correctness on unseen inputs.

[Issue19](https://github.com/teamswyg/laya-tools/issues/19) · [Previous immutable HF30](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1) · [한국어](README.ko.md)
