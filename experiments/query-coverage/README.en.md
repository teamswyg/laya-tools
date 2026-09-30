# Query coverage from the existing index 24

**New features are cheap without another index, but individually fail to improve harm-versus-benefit discrimination.** Their mean repository AUCs are 0.4709–0.4933. No new model, operating threshold or automatic routing is enabled.

## Implementation

`hintsearch.Index.QueryCoverage(query, ids)` returns unique query-term presence fractions for up to 20 selected candidates. Binary-search the existing sorted vocabulary and document posting arrays; do not rescan raw code or use a helper index/model.

Tokenization matches BM25: lowercase Unicode letters/digits, deduplicated query words. Punctuation-only queries contain zero terms. Reject invalid/duplicate IDs, more than 20 candidates, empty queries and queries exceeding 8,192 bytes. Callers supply a permission-filtered index and IDs belonging to that index.

The index gains no persistent fields. Candidate counters/results use fixed arrays, while query tokenization allocates temporary memory. Immutable reads need no shared locks. This is not an allocation-free operation.

Four research features are first-candidate term coverage, mean and population standard deviation across the baseline top min(20,n), and the fraction of unique query terms absent from the baseline vocabulary. Preserve the original 12 features and four score-distribution features; store new values in a separate array **before helper retrieval or gold-rank calculation**. Lexical presence is neither semantic understanding nor correctness probability.

## Signal on 2,948 queries

The [plan](plan-24.json) was committed before execution. As in audit 23, only the training repository selects each feature's orientation. Values below are unweighted means of three repository AUCs, not accuracy.

| New feature | Harm/rest | Benefit/rest | Harm/benefit |
|---|---:|---:|---:|
| First-candidate coverage | 0.5157 | 0.5742 | 0.4933 |
| Top-20 mean coverage | 0.5675 | 0.5696 | 0.4709 |
| Top-20 coverage stddev | 0.5210 | 0.5403 | 0.4745 |
| Query vocabulary absence | 0.4909 | 0.4932 | 0.4756 |

Each full target counts all 2,948 distinct queries once. Harm/benefit is a **478-case subgroup**, excluding page ties, not a separate ≥2,400-case final test. There remain 146 harms, 332 benefits and 2,470 ties. Publish tied-score AP, prevalence and every fold alongside AUC.

Do not flip feature direction after seeing low evaluation AUC. Individual weakness does not exclude useful interactions. As precommitted, next freeze a separate **16-versus-20-feature training comparison**, adding all four features while holding objective, splits and selection fixed. Do not cherry-pick features or simultaneously change the loss. If the joint comparison also fails, these features lack support for default integration.

## Resources and validation

Apple M4 Pro/macOS/Go 1.27.1 synthetic fixture: 3,009 documents, 20 selected candidates, four lowercase query words (two known, two absent). Index construction is outside timing.

| Operation | Three runs | Allocation |
|---|---|---|
| Coverage for 20 candidates | 681.6 / 650.1 / 640.3 ns | 64 B/op, 1 alloc/op |
| Baseline ranking, reused arrays | 11,883 / 11,825 / 11,867 ns | 120 B/op, 3 allocs/op |

These operations are not substitutes: coverage adds work after baseline ranking. No end-to-end speedup is established. This fixed short warm-cache query does not bound longer inputs, case conversion or more known terms. The coverage microbenchmark excludes the research-feature mean/stddev reduction and JSON/IPC.

CPU samples attribute 82.61% cumulative time to the coverage method and 36.65% flat to integer posting binary search in that synthetic run. This is not an application-wide/GPU bottleneck measurement. Raw profiles remain local.

One complete source-check/retrieval/feature/audit process took wall 1.85s, user CPU 1.50s, peak RSS 91,717,632 bytes (~87.5MiB). This is whole-process cost, not per-function memory or proof of RSS improvement over prior experiments.

Full report replay is byte-identical. Audit 23 also replays byte-identically after shared preparation changes. Tests compare posting-based coverage with direct synthetic text inspection and cover duplicates, case/Korean text, unknown words, invalid IDs, concurrent readers, unchanged rankings and feature values. All prior 16-feature diagnostics remain unchanged.

```sh
go run ./cmd/riido-coverageaudit --out .cache/query-coverage-24
go test ./pkg/hintsearch -run '^$' -bench BenchmarkCoverageAndRanking -benchmem
```

[Full results](results-24.json). Repeated development data from three repositories is not fresh ≥2,400 real-user evidence. Agent completion and LLM savings remain unverified. CoSQA reserves remain unscored; no new model weights or HF model repository are produced.
