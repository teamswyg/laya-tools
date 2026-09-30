# Candidate pagination 11: read only what is needed first

[한국어](README.ko.md) · [Measurements](results-11.json)

**`riido-hints --limit 20` emits 20 candidates at a time.** Remaining candidates stay reachable through `page.next_cursor`. Default output remains complete; ranking, models and auxiliary-search policies are unchanged.

## Usage

Save the JSON request from the [original fixture](../baseline-first/README.en.md) as `catalog.json`. No model is required.

```sh
go run ./cmd/riido-hints --identifier-hints --limit 2 < catalog.json > page-1.json
```

Copy `page.next_cursor` from that response and send the same request/options:

```sh
go run ./cmd/riido-hints --identifier-hints --limit 2 \
  --cursor 'NEXT_CURSOR_FROM_PAGE_1' < catalog.json
```

Absence of `next_cursor` means completion. `inspection_rank` remains global, not page-local. Metadata includes total candidates, offset, returned count and a consistency digest. Page size is 1–4,096 and may change between requests. Omit `--limit` or use zero for original full output; cursors require a positive limit.

Changes to query, snapshot, document IDs/content/order, external hints, model hash or search options invalidate the cursor. Changes to complete computed ordering/scores also invalidate it. Do not silently combine pages from different inputs/policies. Invalid cursors produce no partial output.

A cursor is not authentication: it contains a position and request/result digest, is unsigned and is not stored by a server. Callers must keep supplying the same authorized catalog. Pagination does not grant access or crawl repositories. Without server-side caching, each page rebuilds the index and recomputes ranking.

## Actual CLI measurement

Used experiment 10's audited 3,009-function catalog and its first hash-ordered question. This verifies transport, not a new 2,948-question quality evaluation.

| Output | Bytes |
|---|---:|
| Original complete response | 486,628 |
| First 20-candidate page | 3,878 |
| All 151 pages combined | 558,544 |

Initial output decreased by **99.20%**. Reading every page adds metadata and produces about 14.8% more bytes than full output. All 151 separate process calls, catalog retransmissions and searches took about 20.60s; this is neither first-page latency nor server throughput. The output advantage depends on stopping after finding sufficient candidates. Actual LLM tokens, billing and task-success savings were not measured.

Concatenated IDs/global ranks/scores exactly matched complete output. With pagination disabled, the response remained byte-identical to the old CLI. Raw requests contain public source code and stayed local rather than entering Git.

## Checks and next work

Regression tests cover exact reconstruction, changed page sizes, search policies, rejected snapshot/query/document changes, invalid positions/encoding and extra fields. Existing model/hint/input-bound tests remain.

```sh
go test -race ./cmd/riido-hints
```

Next measure how many pages a real agent reads before verifying a target, and the cost of repeated search. Pagination is not universally faster. Learned claims must be compared against controls that account for pages and auxiliary-search cost, not rank alone.
