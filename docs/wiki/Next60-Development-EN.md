# Development data for a small claim/hint model

**30 semantic requests and88 candidate labels were verified and published on HF.** Read the [30-row data and guide](../../experiments/short-claim/next60-development-thirty/README.en.md) and [immutable HF30 release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-30-finite-v1). All558 owned files and every field/order of30 viewer rows matched; older tags were preserved. See [publication and actual CI evidence](../../experiments/short-claim/publication-proof-122/README.en.md). New claim-model/Fit calls for this expansion are zero; Codex savings remain unproven.

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

Read the [guide](../../experiments/short-claim/next60-development-thirty/README.en.md), then consume data/train.jsonl one line at a time. It is **41,428bytes**, SHA-256 **9cdfb758f03adc34a9fb5e00e3c1525921df26d0912f6b554e4ed4c890fd10a2**. Rows contain request/candidate text, labels, weights and provenance.

With HF CLI installed, download **one data file at the fixed commit** as follows. Reading requires neither training nor model download. Run the full verification command from this repository checkout root.

```sh
hf download JooYoon/riidolaya-shortclaim-next60-development \
  releases/next60-30-finite-v1/next60-development-thirty/data/train.jsonl \
  --type dataset --revision 89c0c7c9ddc3135e37e88e9da5d51ac55dff0e1d \
  --local-dir ./riidolaya-next60-data
```

The [Go reader](../../pkg/shortclaimdata) uses bounded arrays and immutable values:16KiB maximum row, at most8 candidates, no locks. Import bytes and github.com/teamswyg/laya-tools/pkg/shortclaimdata for this example.

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

[22 proposed new-source contracts](../../experiments/short-claim/next60-new-source-preview/README.en.md) add zero qualified requests. Three proposals are held because existing behavior suffices or current inputs do not distinguish candidates. Go-list dependency selection and zero text duplicates do not establish execution, semantic independence or training readiness.

## Criteria for the next training round

Always choosing the second candidate succeeds on **27/30=90%** of requests. Always choosing a negative label matches **58/88≈65.9%** of candidate labels. Denominators differ; these are simple controls, not model accuracy. Position, caption-style and lexical controls are needed because high scores can arise without understanding meaning. Existing data and failure records are not rewritten to improve scores.

**30 is a collection checkpoint, not an automatic training trigger.** New source families require dependency, shared-helper and semantic-duplicate audits. Whole-family training, validation and calibration roles are fixed before observations. Readiness is reassessed around60 requests. Existing79 plus30 gives109 arithmetically, not109 independent examples before lineage and deduplication checks.

A later model must reduce total verification work by at least **5%** against fixed-order, lexical and rule controls. Protected evaluation with **2,400 requests per claim domain** remains planned. Historical3 Fits,3 logical models and the failed model's inactive state are unchanged. Ternary storage is separate: smaller storage alone does not establish CPU speed or useful acceleration.

[Issue19](https://github.com/teamswyg/laya-tools/issues/19) records progress and actual publication/CI results. Owned documentation/source use Apache-2.0; full upstream notices retain their licenses. Original runtime bodies, model weights, private inputs, credentials and raw journals are excluded. Source-license checks do not provide blanket clearance for the complete model lineage.

[First two training requests](Native2-Training-EN) · [First original observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)
