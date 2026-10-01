# Using the small behavior-hint tool

[한국어](USAGE-56.ko.md) · [Earlier preparation plan](PLAN-56.en.md) · [Execution plan](execution-plan-56a.json)

`riido-shortclaim` accepts a short request and candidate descriptions, suggesting **which candidate to verify first**. It preserves every candidate and does not certify completion, correctness or permission to act. This separate Go experiment uses neither the Laya encoder nor learned weights.

A person can provide a request such as removing expired entries while preserving order, with short candidate descriptions. An agent can inspect actual code and independent checks in the returned ID order. The tool cannot guarantee that descriptions faithfully represent code. It does not silently summarize full code or call an LLM.

## First run

From the repository, build with Go1.27.1. Python, model downloads and external API keys are unnecessary.

```sh
CGO_ENABLED=0 go build -trimpath -o bin/riido-shortclaim ./cmd/riido-shortclaim
./bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
```

`verification_order` contains candidate IDs and scores. The status is always `unverified_heuristic`; scores are not success probabilities. Lexical controls can propose a wrong order when polarity or negation matters, so independently verify the code. `input_sha256` identifies length-framed request, candidate, order and metadata strings. Raw text and provenance are not echoed, but **candidate IDs are output: use IDs suitable for disclosure**.

Requests and each candidate must be valid UTF-8, at most512 bytes and, after normalization, at most512 bytes and32 words. There are1–8 candidates. One JSON record is limited to12KiB, IDs to1–64 bytes and provenance to1–128 bytes of bounded ASCII identifiers. The schema is `riido-short-behavior-claim-v1`. Duplicate/unknown keys, invalid Unicode, excess lengths and unknown schema fail with fixed error codes. Inputs are never truncated. This is the implemented refinement of the preparation-only fallback proposal.

## Repeated requests

Send one JSON object per line with `--stream`. One resident process serves multiple requests, avoiding repeated process startup. An invalid record stops processing; later records are not served. Closing input ends normally. There is no result cache: each request performs actual computation.

```sh
./bin/riido-shortclaim --stream --baseline bm25 < public-requests.jsonl
```

`public-requests.jsonl` names a file you prepare with the same input schema. This opt-in preview does not change `riidolaya` routing or Codex defaults. Go callers can use `pkg/shortclaim.Load`, `Validate` and `Rank`. Because callers can construct or modify `Prepared`, `Rank` validates it again. There is no approval or automatic task-execution interface.

## Four controls

| Option | Meaning and scope |
|---|---|
| `fixed_order` | Original order |
| `bm25` | Candidate-local term and document frequency matching |
| `lexical_ordered` | Word matching and ordered bigram matching |
| `narrow_rule` | Rule matching only when both sides fully follow the disclosed `behavior-v1:` grammar; if any candidate is unsupported, the entire request falls back to BM25 |

```sh
./bin/riido-shortclaim --baseline narrow_rule < examples/shortclaim/rule.json
```

Rules accept1–4 clauses: `if [not] name [operator name] then [not] action [object] else [not] action [object]`. Operators `< <= > >= == !=` require surrounding whitespace. Actions are keep/remove/accept/reject/preserve/reverse. Condition negation swaps branches for comparison; action negation remains an exact flag. Clause and operand order are preserved. Rules do not infer ordinary prose/code meaning, synonyms, antonyms or general logical implications. Natural-language development data was not rewritten to fit these rules.

## Structure and measurement

Fixed arrays bound each request to8 candidates and32 words. Tokens retain string views; small score arrays and stable ordering preserve candidates. There is no shared cache, shared mutable state, shared lock or runtime map. Metadata and targets are excluded from features. This does not claim a complete ECS/SoA or SIMD system: measure bounded-array work before adopting such changes.

`--benchmark` measures preparation, ranking, input digest, serialization and a full Go request on an original eight-candidate fixture. Ranking combines revalidation, features, scores and ordering; it is not separate timing for each. `cmd/riido-shortperf` verifies a frozen plan and records OS process RSS, CPU and cold startup. The following result document provides commands and observations.

`--cpuprofile` is **local-only** maintainer instrumentation. Raw profiles may contain personal paths and stack information; never commit or upload them. Go HeapAlloc includes uncollected objects and differs from OS peak RSS. This experiment uses CPU only, without GPU, ORT or an encoder.

The48 development requests are not48 independent samples. Shared candidates, prototypes, source and core templates form transitive groups; natural-language fidelity is separate from finite Go truth tables. Fast execution alone proves neither semantic accuracy nor LLM savings and does not replace the target of2,400 distinct final requests per domain.
