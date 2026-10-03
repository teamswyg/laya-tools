# Development data for a tiny claim and hint model: 35 verified requests

The target is a tiny model that frequently proposes code-behavior hints at low cost. For example, it might suggest that a candidate can continue to a later line after an error; an actual check then verifies that claim. A hint alone grants neither execution approval nor correctness.

The public pool contains **35 requests and 102 labels: 35 positive, 67 negative**, with 171 inputs, 509 all-candidate observations and 500 selected-candidate observations. The previous 33 rows remain byte-exact. See the [Go materialization and Reader receipt](https://github.com/teamswyg/laya-tools/blob/310bdeef76014204841076c01cf9ee60cdfe7375/experiments/short-claim/next60-development-thirtyfive/MATERIALIZATION.v1.json) and [usage/evidence](https://github.com/teamswyg/laya-tools/blob/310bdeef76014204841076c01cf9ee60cdfe7375/experiments/short-claim/next60-development-thirtyfive/README.en.md).

The two additions check complete JSON values per line and invalid UTF-8 rejection before append. Satisfaction, counterexample and unknown remain distinct. An unavailable error channel stays unknown. A known counterexample accompanying unknown conditions supports a negative label only through that counterexample. Existing API contracts and new requested contracts differ; a mismatch alone is not an upstream defect.

## Usage for people and agents

This command downloads **one data file**. No model download, inference or training is needed; Codex integration is optional. One line is one request, candidates contain code descriptions, and labels express finite observed support or contradiction. Unknown conditions and excluded candidates remain in separate qualification/evidence files.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-35-finite-v1/next60-development-thirtyfive/data/train.jsonl \
  --type dataset --revision 7621cd34d298b4e2dfb59def7a9adaf6e45fef52 \
  --local-dir ./riidolaya-next60-data
```

From a repository checkout, the following reproduces all saved comparisons, data bytes and actual project Reader values for 35 requests. It launches no original observer, model or new Fit.

```sh
bash scripts/verify-next60-thirtyfive.sh
```

## Actual publication and checks

[PR126](https://github.com/teamswyg/laya-tools/pull/126) passed all four required [CI jobs](https://github.com/teamswyg/laya-tools/actions/runs/37123177961), and the bot merged the identical source tree. Linux/macOS compare 33 saved rows/309 predicates and all 35 data rows. Existing native Laya inference passed separately. **GPU execution is unverified.**

The [fixed Hugging Face release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/7621cd34d298b4e2dfb59def7a9adaf6e45fef52) and next60-35-finite-v1 tag are published. Inventory is 720 owned files/721 including attributes. Every **84 new/changed files plus managed attributes**, 888,897 bytes, was directly read at the fixed commit and matched. All old 639 paths and eight tag refs remain. This does not claim another download of all old files at the new commit.

The viewer matched **every field and order for 35 rows, partial=false, zero truncated cells**. Response revision and top-level truncation fields are unavailable and remain unknown. The viewer follows current main, separate from immutable file proof.

## Criteria for the next training round

35 is a development collection checkpoint. Always choosing position 2 already succeeds on 28/35; a model must first demonstrate meaningful utility beyond simple controls. The new Golden 20-to-60 coverage plan is a separate cohort, covering five/eight candidates and zero/one/many positives, unknowns and ambiguous policies. Whole source families, shared helpers, translations and variants must stay within one evaluation role.

A proposed Reader preserves unknown-only labels/weights as null and withholds ambiguous parents from training. Connecting it directly to the current training projection could turn unknowns into negatives; masks must survive transformation, loss and evaluation first. New Fits and model calls are zero. Protected evaluation of 2,400 requests per domain and a 5% actual total verification-work reduction remain unfulfilled. Ternary storage size and useful CPU inference must be proven separately.

Owned source, text and fixtures use Apache-2.0; existing complete upstream notices remain. No upstream source bodies, weights, binaries, private inputs, credentials or raw journals are published. Model publication rights require separate review. [Issue19](https://github.com/teamswyg/laya-tools/issues/19) · [한국어](Next60-Development-KO).

<details>
<summary>이전 33개 시점 안내 원문 / Previous 33-request guide, verbatim</summary>

# Finite development data for a small claim and hint model: verified 33 requests

The target is a tiny **claim and hint model** that frequently suggests useful code candidates at low cost. Tests and counterexamples support the hint. This package builds development supervision; it does not report a new model's accuracy or Codex savings.

**Owned Go generated33 requests and96 labels; the actual project Reader completed33 calls,33 returns and33 value matches.** See the [materialization receipt](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/MATERIALIZATION.v1.json) and [Reader receipt](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/READER.v1.json). The previous30 rows remain byte-exact. Data:45,390 bytes, SHA-2567b28ae6d119884c902b32ba781b3514f1d3774ca3a60cea16fe4ce11a3aad919. PR124 passed Linux/macOS reproduction and all four required CI jobs, then the bot merged it. Actual publication below was checked afterwards.

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

See the [previous30 usage guide](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirty/README.en.md) for Go Reader examples. The Reader uses fixed arrays for rows up to16KiB and eight candidates, without locks. Bitset stays in source group82; GJSON, Match and Pretty stay together in group83, all development_train. Flat metadata names bits-and-blooms-bitset and tidwall-gjson map explicitly to preserved upstream repository names.

## Measurements and next training

Original collection ran once. Whole-worker OS peak RSS was **17,612,800 bytes, about16.8MiB**, and wall time was **8.052 seconds**, including1,739 durable file/directory writes and ACKs. These are not model inference memory, GPU execution or speedup results. The stripped saved-comparer executable was **2,850,722 bytes**, which is not a model size. Go pprof and OS RSS have different measurement scopes.

The three new suitable candidates occupy different positions, but always choosing the second candidate still yields28/33 for this pool. Position, style, lexical controls and semantic duplication must be checked before training. **33 is a collection checkpoint, not an automatic Fit trigger.** Reassess whole-source-family readiness around60 requests; retain the protected2,400 requests per domain plan and5% total verification-work reduction target. Existing three Fits, three logical models and the inactive failed model remain unchanged. Ternary storage size and CPU utility are separate experiments.

Next collection preparation covers complete values per physical JSON line, preserved callback prefixes before invalid lines, and invalid UTF-8 append rejection. Existing ForEachLine already honors callback stopping; that behavior is not repackaged as a defect. Freeze new expected outputs before observation.

## Public scope and history

Owned code, descriptions and fixtures are Apache-2.0. Separate complete BSD/MIT notices for selected upstream APIs remain in evidence/notices. No original source bodies, model weights, execution binaries, raw journals, private task inputs or credentials are included. Original source/go.mod/notice hashes provide traceability. Model publication rights are separate.

history/CORRECTIONS.v1.json records the draft unknown-scope error, flat metadata correction, first comparer build's size-budget failure and subsequent verification. Failed files and old drafts remain private and recoverable. Labels have finite fixed-input support and do not guarantee correctness on unseen inputs.

[Issue19](https://github.com/teamswyg/laya-tools/issues/19) · [Previous immutable HF30](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1) · [한국어](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/experiments/short-claim/next60-development-thirtythree/README.ko.md)

## Actual publication and a small download

The [immutable HF33 release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/2a1c2224e89f60eef9e381f98c0117206f60e0d2) is public. All **82 new/changed files,842,356 bytes**, plus managed attributes, were downloaded at the fixed commit and compared. Inventory is638 owned files/639 including attributes; previous paths and seven tag refs remain. The earlier558-file whole-release proof belongs to HF30; old files were not all downloaded again at this commit.

The first viewer request returned500 while the server prepared data. The second matched **every field/order of33 rows,partial=false,zero truncated cells**. Response revision and top-level truncation fields are unavailable and remain unknown. The viewer follows current main, separate from fixed-commit file proof.

Download **one data file** below. People and agents can read it without downloading a model or running training.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-33-finite-v1/next60-development-thirtythree/data/train.jsonl \
  --type dataset --revision 2a1c2224e89f60eef9e381f98c0117206f60e0d2 \
  --local-dir ./riidolaya-next60-data
```

[Go Reader](https://github.com/teamswyg/laya-tools/blob/511e8ac7c213f3b48568025eb9d83882d525bcdb/pkg/shortclaimdata) example. Import bytes and github.com/teamswyg/laya-tools/pkg/shortclaimdata.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

Input, supervision and provenance remain separate. Reading does not score or train. The next input bridge passed eight Go tests covering synthetic1/7/33/60-row collections, sparse groups and invalid plans/pins/row bindings. It has not admitted the real new corpus to training or called Fit. Exhaustive candidate-order bias auditing is in preparation.

<details>
<summary>Previous HF30 guide verbatim — “current” and counts below describe that stage</summary>

# Development data for a small claim/hint model

**30 semantic requests and88 candidate labels were verified and published on HF.** Read the [30-row data and guide](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-development-thirty/README.en.md) and [immutable HF30 release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1). All558 owned files and every field/order of30 viewer rows matched; older tags were preserved. See [publication and actual CI evidence](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/publication-proof-122/README.en.md). New claim-model/Fit calls for this expansion are zero; Codex savings remain unproven.

The goal is a **claim/hint model** cheap enough to call frequently with very little CPU and memory. It might suggest “this candidate seems useful for the request.” Tests, search and verification support the final choice. Reducing total work by narrowing candidates matters more than generating prose.

| Unit | Current verified/published30 | Previous HF23 |
|---|---:|---:|
| Distinct semantic requests | 30 | 23 |
| Candidate labels | 88:30 positive,58 negative | 67:23 positive,44 negative |
| Fixed input variants | 146 | 113 |
| Full original observations | 434 | 335 |
| Observations of selected training candidates | 430 | 331 |
| New training/model inference in this expansion | 0 | 0 |

Input variants and candidate executions do not increase independent request counts. All30 rows use development_train and weight1. Previous23-row and separately verified27-row files remain byte-for-byte prefixes. Source families use groups76–81; further retryablehttp collection is on hold. New-source generalization remains unverified.

## Use by people and agents

Read the [guide](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-development-thirty/README.en.md), then consume data/train.jsonl one line at a time. It is **41,428bytes**, SHA-256 **9cdfb758f03adc34a9fb5e00e3c1525921df26d0912f6b554e4ed4c890fd10a2**. Rows contain request/candidate text, labels, weights and provenance.

With HF CLI installed, download **one data file at the fixed commit** as follows. Reading requires neither training nor model download. Run the full verification command from this repository checkout root.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-30-finite-v1/next60-development-thirty/data/train.jsonl \
  --type dataset --revision 89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d \
  --local-dir ./riidolaya-next60-data
```

The [Go reader](https://github.com/teamswyg/laya-tools/tree/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/pkg/shortclaimdata) uses bounded arrays and immutable values:16KiB maximum row, at most8 candidates, no locks. Import bytes and github.com/teamswyg/laya-tools/pkg/shortclaimdata for this example.

```go
example, err := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(line))
if err != nil {
    return err
}
input := example.Input()
supervision := example.Supervision()
```

Only **request and candidate text** enter model features. IDs, sources, groups, revisions and verification results are provenance or supervision. Reading runs no scoring, training or inference. Arrays alone are not evidence of a measured speedup.

The following command reproduces saved data and checks the actual project Reader and111 saved predicates. It passed locally and on Linux/macOS in PR121 CI. Existing native Laya checks passed separately; this command itself executes no model or original functions.

```sh
bash scripts/verify-next60-thirty.sh
```

## What was actually verified

Materialized30 rows passed **30 Reader calls,30 returns and30 matching values**, with **60 saved checkpoints** before and after calls. Inputs, candidates, labels, weights and unused array slots are compared. **Eight unknown predicates** remain unknown. A negative candidate needs a separate known counterexample; unknown is never silently converted to false.

The last three requests cover an INI total input byte budget, bounded file reading and exclusive file writing. Original Go code produced **42 observations** from14 inputs×3 candidates. The111 fixed predicates were89 satisfied,20 known mismatches and2 unknown. Reference candidates were adopted because observations satisfied fixed inputs, not because of their names. These checks do not guarantee every parser option, filesystem concurrency or all possible inputs.

The observation collector peaked at about **17.4MiB OS RSS**, taking about **2.52seconds** including startup, source verification and durable storage. These are not Laya/claim-model inference memory, GPU execution or Codex savings measurements. Original code ran once. The first saved-comparer configuration-format failure was preserved; only reading saved files was repeated.

[22 proposed new-source contracts](https://github.com/teamswyg/laya-tools/blob/560d5b2cfcdf6f1c23f8b624dbdca3a258ca462c/experiments/short-claim/next60-new-source-preview/README.en.md) add zero qualified requests. Three proposals are held because existing behavior suffices or current inputs do not distinguish candidates. Go-list dependency selection and zero text duplicates do not establish execution, semantic independence or training readiness.

## Criteria for the next training round

Always choosing the second candidate succeeds on **27/30=90%** of requests. Always choosing a negative label matches **58/88≈65.9%** of candidate labels. Denominators differ; these are simple controls, not model accuracy. Position, caption-style and lexical controls are needed because high scores can arise without understanding meaning. Existing data and failure records are not rewritten to improve scores.

**30 is a collection checkpoint, not an automatic training trigger.** New source families require dependency, shared-helper and semantic-duplicate audits. Whole-family training, validation and calibration roles are fixed before observations. Readiness is reassessed around60 requests. Existing79 plus30 gives109 arithmetically, not109 independent examples before lineage and deduplication checks.

A later model must reduce total verification work by at least **5%** against fixed-order, lexical and rule controls. Protected evaluation with **2,400 requests per claim domain** remains planned. Historical3 Fits,3 logical models and the failed model's inactive state are unchanged. Ternary storage is separate: smaller storage alone does not establish CPU speed or useful acceleration.

[Issue19](https://github.com/teamswyg/laya-tools/issues/19) records progress and actual publication/CI results. Owned documentation/source use Apache-2.0; full upstream notices retain their licenses. Original runtime bodies, model weights, private inputs, credentials and raw journals are excluded. Source-license checks do not provide blanket clearance for the complete model lineage.

[First two training requests](Native2-Training-EN) · [First original observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)

</details>

</details>
