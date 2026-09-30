# CoNaLa source and distinct-question audit 16

**2,879 rows are not 2,879 distinct original questions.** The pinned release contains 2,074 question IDs and 2,089 normalized original intents. It cannot alone establish a 2,400-question independent user evaluation set. No training, retrieval scoring, model selection or source-code execution occurred.

| Count | Train | Test | Combined unique count |
|---|---:|---:|---:|
| Rows | 2,379 | 500 | 2,879 |
| Question IDs | 1,710 | 364 | 2,074 |
| Original intents | 1,724 | 365 | 2,089 |
| Nonempty rewritten intents | 2,211 | 451 | 2,662 |
| Outer-trimmed exact snippets | 2,326 | 490 | 2,811 |

There are 102 empty rewrites and 2,758 rows with nonempty rewrites differing from the original. An ID can have multiple wording variants, so both ID and intent counts are reported. Normalization only lowercases and collapses Unicode whitespace; it does not establish semantic uniqueness.

The splits share no IDs or normalized original/rewritten intents, but share five exact snippets. Components linked by equal ID, normalized original intent or exact snippet total 2,056, with a largest component of 14 rows. Five cross-split components contain 29 rows. This describes overlap, not benchmark misconduct. Internal Python indentation is preserved.

## Source and rights

The [official description](https://conala-corpus.github.io/) distinguishes Stack Overflow titles (`intent`) from rewrites incorporating code variable names and arguments (`rewritten_intent`). On 2026-09-30 the official ZIP URL `https://www.phontron.com/download/conala-corpus-v1.1.zip` returned HTTP 200 but `text/html` and `index.html`. It was not treated as a ZIP.

Instead, download only the README and two curated JSONL files from the authors' lab's [pinned neulab/conala revision](https://huggingface.co/datasets/neulab/conala/tree/fbc749f1c537e5c3834e93b15784302e331debe2). Despite `.json` names, they contain JSONL; source data totals 627,077 bytes. The roughly 600k mined corpus and Python dataset script were neither downloaded nor executed. Byte identity with the unavailable original ZIP is unverified; results refer specifically to this HF revision.

The HF card states MIT, while [Stack Overflow's official guidance](https://stackoverflow.com/help/licensing) describes CC BY-SA terms by contribution date. Card metadata is not treated as blanket clearance for source contributions or additional annotations. This audit establishes neither redistribution rights nor model-publication suitability. No source text, question IDs or row-level data are redistributed to Git/HF. The Go auditor is independently implemented without copying baseline code.

## Reproduction and next choice

plan-16.json was committed before the audit. Verify source SHA-256 and expected row counts. Readers enforce 2 MiB/file, 256 KiB/line and bounded row counts; unknown fields and malformed JSON fail. Aggregate output replays byte-identically. Tests cover connected overlap, empty rewrites, Python indentation and invalid input. Arrays, sorting and array-based component union require no shared locks.

```sh
hf download neulab/conala README.md data/conala-paired-train.json data/conala-paired-test.json --type dataset --revision fbc749f1c537e5c3834e93b15784302e331debe2 --local-dir .cache/conala-16
go run ./cmd/riido-conalaaudit
```

The built auditor took 0.35 s wall, 0.01 s user CPU and 11,796,480 bytes peak RSS on Apple M4 Pro. This is one process execution, not throughput or model inference speed. GPU was unused.

Retain CoNaLa as a smaller original-query auxiliary evaluation candidate, not a standalone 2,400-question final set. For larger sources, do not inflate independent counts with rows, rewrites or repetitions. Next, audit rights and answer quality in other real-question sources or deduplicate multiple sources. No model has been evaluated, and existing unused evaluation reserves remain unscored.
