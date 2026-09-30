# Laya routing suitability and GitHub repository preview

[한국어](repository-routing-preview.ko.md) · English

## What is feasible, and what remains unproven

Laya classifies text against a finite set of choices. Repository names, roles, and short summaries fit that interface. **Well-formed output is not the same as a correct choice:** the model can select the wrong repository without inventing an out-of-catalog name.

Repository selection asks which codebase owns a feature. Coding-model selection predicts which model will complete the task adequately. The latter needs labels for completion, failure cost, retries, and usage. Short tasks are not necessarily easy, and Laya's probabilities are not coding-model success probabilities.

The [upstream README](https://github.com/NandhaKishorM/laya) distinguishes English base, multilingual, and typed-decisions checkpoints. **Our installed root base is English ModernBERT, 421M parameters, 512 tokens.** Multilingual/long-context claims do not automatically apply. A different tokenizer and graph need separate validation; simply substituting files is unsupported.

The [pinned base card](https://huggingface.co/convaiinnovations/laya/blob/55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851/README.md) discusses domain calibration, confident errors across languages, and wording effects. Calibration on a general benchmark does not establish Riido routing accuracy. The [laya-code card](https://huggingface.co/tindang/laya-code/blob/25f97e5a2ec5f8cf7218a4f67504367d8832e1fe/README.md) says choice/score were not trained; this preview uses base instead.

## A design for Riido

```text
Accessible repository catalog supplied by Riido
  → local index of names, summaries, keywords
  → at most 4 keyword candidates by default
  → optional Laya choice, including none/ambiguous
  → preview candidate / experimental suggestion / abstention
  → no execution adapter yet
```

Do not put all source or READMEs into every inference. Supply role summaries and aliases already known to Riido. First determine whether keyword candidates suffice, then compare Laya on ambiguous requests.

`pkg/reporouter` is a standard-library-only public Go package. It snapshots the catalog and performs no per-request GitHub fetch, clone, embedding generation, or file scan. Native inference is a separate adapter. Text, JSON, and persistent JSONL interfaces are available.

The caller must filter the catalog by access permissions before providing it. Disabled repositories are excluded. Routing neither verifies nor grants access. Descriptions are not executed as instructions, but malicious descriptions may still influence classification; restricting outputs to catalog IDs does not solve that issue.

## Usage

The example catalog is fictional, not Riido's real repositories. Run from the repository root:

```sh
riidolaya repo-preview --catalog examples/repositories/catalog.json --json 'refund invoices'
riidolaya repo-preview --catalog examples/repositories/catalog.json --laya --json 'refund invoices'
riidolaya repo-serve --catalog examples/repositories/catalog.json --laya
```

JSONL input:

```json
{"id":1,"query":"Fix login sessions"}
{"id":2,"query":"로그인 비밀번호 수정"}
```

Catalog:

```json
[
  {
    "name": "example/identity-api",
    "summary": "Authentication service: login and access tokens.",
    "keywords": ["auth", "로그인", "인증"],
    "disabled": false
  }
]
```

Call `reporouter.New(repos)` then `idx.Preview(query, reporouter.DefaultConfig(), judge)`. A nil judge needs no model. Custom adapters implement `Judge.Choose`.

**Every result has `preview: true` and is not execution authorization.**

- `candidate`: lexical evidence, not a calibrated automatic selection. Model-disabled, unvalidated-language, and inference-failure paths state that explicitly.
- `suggest`: experimental Laya choice passes confidence 0.9 and a 0.15 top-two probability margin; still a preview.
- `abstain`: no candidates, none/ambiguous, insufficient confidence, truncated input, or invalid inference. `suggested` is empty.

The CLI uses four candidates; the public API allows up to eight. Including none stays within the engine's 2–10 options. Exceeding option token budgets flags truncation and withholds suggestions. We fixed previously unreported option truncation without changing token-ID construction.

Non-Latin letters in the query or candidate summaries skip the English model and return lexical candidates. Korean keywords remain usable. This is not complete language identification; other languages using Latin letters are not thereby validated.

Limits: CLI JSON 1 MiB; public API metadata 4 MiB / 4,096 repositories; summary 2,048 bytes; 32 keywords of up to 128 bytes; query 8 KiB. The model's separate 512-token budget still applies.

## First measurement: no demonstrated benefit from Laya

M4 Pro/24 GiB, Go 1.27, four CPU threads, INT8 base, 2026-09-30. Six fictional repositories and 24 original questions: 12 English/6 Korean single-repository positives, four out-of-scope and two cross-repository cases. These are easy development examples with overlapping vocabulary, not independent or production evaluation.

| Metric | Lexical only | Laya enabled |
|---|---:|---:|
| Positive lexical Top-1 | 18/18 | Same candidate stage: 18/18 |
| Gold repository in shortlist | 18/18 | 18/18 |
| Actual model judgments | 0 | 15; Korean/no-candidate cases skipped |
| Raw model correctness before thresholds | N/A | 8/15 |
| Model suggestions passing policy | N/A | 0/24 |
| Remaining candidate/suggestion on six negative cases | 3/6 | 0/6 |
| Request median, excluding model load | 0.001 ms | 39.669 ms |
| Request p95, excluding model load | 0.002 ms | 54.040 ms |
| Model load | None | 941.815 ms |
| Peak process RSS | About 11.5 MiB | About 1,436 MiB / 1.40 GiB |
| Go heap at evaluation end | About 0.54 MiB | About 12.7 MiB |

Timings include all 24 cases, including skipped inference, and include the first inference but exclude model creation. RSS came from separate `/usr/bin/time -l` processes. Go heap depends on GC timing and excludes native memory. GPU execution/memory was not measured.

Zero accepted suggestions do not establish high precision: **coverage is zero and accepted precision is undefined**. The model suppressed negative candidates but also abstained on all positive English tasks. We did not lower thresholds to make results look better.

Lexical search also suggested candidates for camera-history and cross-repository work. Its 18/18 positive score on easy fixtures is insufficient for automatic execution.

Pure-Go synthetic candidate-search microbenchmarks, excluding index construction/model inference:

| Repositories | Time/request | Allocation/request |
|---:|---:|---:|
| 10 | 1.114 µs | 768 B |
| 100 | 12.772 µs | 2,400 B |
| 1,000 | 170.631 µs | 16,992 B |

Allocations are not resident-index memory. Catalog size, text length, and distinct words affect memory. No background daemon is installed. Keyword mode loads no model. `repo-serve --laya` loads lazily for the first eligible request and retains it until exit; separate warm processes do not share model memory.

Raw results: [lexical](../benchmarks/results/repo-preview-lexical.json), [Laya](../benchmarks/results/repo-preview-laya.json).

```sh
go build -o bin/repoeval ./cmd/repoeval
/usr/bin/time -l ./bin/repoeval
/usr/bin/time -l ./bin/repoeval --laya
go test ./pkg/reporouter -run '^$' -bench . -benchmem
```

## Next validation steps

1. Connect Riido's existing names, roles, and aliases to a local catalog; compare existing decisions with previews without publishing private inputs.
2. Label repositories and separate English/Korean, similar names, overlapping roles, new repositories, no-match, and multi-repository cases. Measure shortlist recall separately; Laya cannot recover an omitted candidate.
3. Compare lexical, Laya-assisted, and optionally other small classifiers/embeddings on identical questions. Track p50/p95, cold start, RSS, coverage, and wrong suggestions.
4. Consider domain training/calibration or multilingual ONNX only after demonstrating value. Separate training, calibration, and final evaluation data; verify tokenizer/graph compatibility.
5. Evaluate coding-model selection separately using successful completion, recovery, retries, and total usage. Current numbers do not establish subscription savings.

For now, connect **lexical preview first and keep Laya as a measurement option**. Automatic work assignment and live Codex model switching are not implemented.
