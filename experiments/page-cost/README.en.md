# Page-cost experiment 12

We searched 2,948 distinct documentation queries over 3,009 functions. The 24 model files are not 24 questions. These previously observed queries come from three repositories; this is not an independent final evaluation of real user requests.

The primary page size, frozen before scoring, is 20. We compare raw BM25 with baseline-first interleaving of identifier-normalized BM25.

| Aggregate metric | Baseline | Helper |
|---|---:|---:|
| Requests through the page containing the first known target | 9,488 | 6,287 |
| Candidates emitted through that page | 189,760 | 125,740 |
| Equal-cost search-work proxy | 9,488 | 12,574 |

Pages and emitted candidates fall approximately 33.7%, while the work proxy rises approximately 32.5%. Page wins/losses/ties are 332/146/2,470 questions; work-proxy wins/losses/ties are 183/2,661/104. Enabling the helper everywhere is not demonstrated to save CPU.

The current stateless CLI rebuilds indexes and rankings on every page. Baseline uses one index, helper two. The proxy counts baseline pages versus twice the helper pages, assuming equal index/search cost. Normalization, sorting, construction and I/O actually differ. This is not measured CPU time, latency, power, GPU cost, or billed tokens.

We assume an oracle recognizes the first known target and count the entire final page, capped at catalog size. Actual agent verification and task completion are unmeasured; other valid answers may be unjudged. results-12.json also reports fixed page sizes 1/5/10/20/50. Aggregate work proxy increases at every tested size.

## Implication for training

The old binary label, any rank improvement, misses page boundaries and extra search cost. Future losses must distinguish candidate output cost from measured search cost. An arbitrary exchange rate must not be presented as demonstrated efficiency. Measure a session that reuses rankings first, freeze cost labels for that architecture, then compare repository-separated training. No optimal learned policy is demonstrated here.

## Reproduction

Use the pinned source projection and ZIP verification from experiment 07. Raw data stays outside Git.

```sh
go run ./cmd/riido-interleaveprobe --page-cost > results.json
go test ./internal/retrievalbench
```

plan-12.json was committed before scoring. Tests cover page boundaries, a partial last page, reversals between page and work improvements, and invalid target ranks. Compare replay output byte for byte. This experiment trains or publishes no model and changes no default policy.
