# Baseline-first ordering 10: an opt-in nonlearned search helper

[한국어](README.ko.md) · [Full results](results-10.json)

**Keep the baseline's first candidate and interleave identifier-split suggestions afterward.** On 2,948 documentation-proxy questions this preserves Recall@1 and improves Recall@10/mean target rank. This is a nonlearned baseline, not a learned-model success. Defaults remain unchanged.

## Try it

No model download or Python is needed for this original public fixture:

```sh
printf '%s\n' '{"snapshot_id":"demo","query":"read config","documents":[{"id":"a","text":"read"},{"id":"b","text":"config"},{"id":"c","text":"func ReadConfig() {}"}]}' \
  | go run ./cmd/riido-hints --identifier-hints
```

Baseline order a,b,c becomes a,c,b. Candidate a stays first; splitting `ReadConfig` into `read config` introduces c second. Output contains IDs, ranks and original BM25 scores, not document text. Original scores need not decrease in the interleaved order. Status is `unverified`, policy `identifier_bm25_baseline_first`.

Omit the flag for baseline behavior. Combining it with external hints or `--model` is rejected. If normalized input exceeds existing limits or contains no searchable query words, retain the baseline with policy `bm25_identifier_input_out_of_scope`. Callers must supply only authorized documents; the tool does not crawl repositories or decide permissions. It currently emits all candidates, so output-token costs for large catalogs still need separate consideration.

## Boundaries

The Go API is `hintsearch.InterleaveBaselineFirst(baseline, hints)`. Partial hints are allowed; all baseline candidates remain without duplicates. A candidate at baseline rank r appears by min(n,2r−1), so baseline rank 3 appears by rank 5. This bounds candidate checks under accurate equal-cost verification, not wall time, tokens, multi-evidence tasks or hint construction. Preserving the first candidate does not prevent regressions later in the ranking.

## Corpus results

Same three pinned repositories, 3,009 functions and 2,948 distinct descriptions as 07/09. Exclusions, candidate hash ordering and known targets were unchanged. Implementation was committed before first corpus scoring; replay reproduced the complete JSON byte-for-byte. This previously observed documentation proxy is not a fresh independent user-query test.

| Policy | Recall@1 | Recall@10 | MRR | Mean first-target rank |
|---|---:|---:|---:|---:|
| Raw BM25 | 41.79% | 74.83% | 0.5336 | 48.99 |
| Auxiliary-first interleave | 39.21% | 81.51% | 0.5479 | 26.71 |
| Baseline-first interleave | 41.79% | 81.38% | 0.5601 | 26.70 |

Baseline-first improves 733 questions, worsens 674 and leaves 1,541 unchanged. Average gain is not universal gain. Per-repository metrics are in JSON. Learned models must beat this nonlearned control while accounting for additional cost. No experiment-09 failure was promoted and no HF weights were replaced.

## Validation and cost

Exhaustive tests for 1–7 candidates cover every hint permutation/partial length, first/full-candidate preservation and the 2r−1 bound. Existing auxiliary-first behavior remains tested. Re-running the experiment-09 trainer through the updated shared implementation reproduced its complete report. Actual CLI execution and tests cover the example, conflicting inputs and out-of-scope fallback.

Apple M4 Pro/Go1.27.1 CPU: the whole comparison took 3.02s wall/2.67s user CPU, peak RSS 94,683,136 bytes. This includes source validation and two complete search preparations for comparing policies; it is not single-CLI/service latency. The option builds an additional index and search, so end-to-end time savings are not guaranteed. No training, GPU or LLM call occurs in this comparison.

```sh
# Uses the pinned corpus prepared for experiment 07.
go run ./cmd/riido-interleaveprobe
go test -race ./pkg/hintsearch ./cmd/riido-hints ./internal/retrievalbench
```

Next training must account for inspection gain and auxiliary-search cost relative to this control. Actual user-task success, end-to-end time/token savings, independent questions and reserved CoSQA evaluation remain outstanding.
