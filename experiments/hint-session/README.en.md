# Reusable search session 13

`riido-hints --session --limit 20` computes a query's ranking once. Send the existing catalog request as one JSON line to receive the first page. Subsequent input lines contain only `cursor` (copied from the response's `next_cursor`) and `limit`. Responses are compact JSON lines. The existing single-request mode remains available.

```sh
go build -o riido-hints ./cmd/riido-hints
./riido-hints --session --limit 20 --identifier-hints
```

Example first input:

```json
{"snapshot_id":"public-demo","query":"read config","documents":[{"id":"a","text":"read"},{"id":"b","text":"config"},{"id":"c","text":"func ReadConfig() {}"}]}
```

This small example finishes in one page. For larger catalogs, send another line, replacing the placeholder with the returned cursor:

```json
{"cursor":"COPIED_NEXT_CURSOR","limit":20}
```

Close input to terminate the process. Agents should use stdin/stdout pipes and flush after each input line. If the final page has arrived, close input. A static file containing only the initial request also exits normally at EOF.

## Ownership and limits

Each process owns one immutable query/snapshot result. After initialization it retains candidate IDs, ranks, scores and response metadata. The application writes no source or results to files. Temporary source/index objects become eligible for Go garbage collection after initialization; immediate erasure is not guaranteed. There is no shared cache or new shared lock. Results remain in memory while the process stays open; callers control its lifetime.

Start a new process for changed queries, permissions or catalogs. Callers must supply authorized candidates only. Permission changes are not automatically reflected; terminate the session when permissions change. A cursor checks consistency, not authentication. Invalid cursors, unknown fields, invalid page limits and oversized messages terminate with an error. Previously delivered valid responses remain valid; no partial response is emitted for a malformed request.

Initial input is at most 16 MiB including newline; continuation lines are at most 1,024 bytes. Existing limits on 4,096 candidates, source bytes and query size remain. Page limits are 1..4,096; session startup rejects `--cursor`. There is no automatic expiry or network server.

## Evidence

Tests compare first and subsequent pages with stateless output. A separate local test reconstructs all 3,009 public-source candidates over 151 pages under both policies and checks exact IDs, ranks and scores. Raw source stays outside Git.

```sh
RIIDO_PUBLIC_SESSION_FIXTURE=/path/to/public-catalog.json go test -race ./cmd/riido-hints
```

results-13.json records a one-query complete traversal transport measurement. It is not 2,948 independent questions or evidence of LLM billing savings. Reading only the first page yields much less benefit from ranking reuse. Initialization includes a one-time internal full-response JSON conversion.

| One query, all 151 pages | Repeated stateless requests | Session |
|---|---:|---:|
| Raw BM25 total time | 11.045 s | 0.072 s |
| Helper total time | 20.412 s | 0.138 s |
| Input bytes | 360,069,617 | 2,405,664 |

Session first pages took 0.068/0.132 s, with peak RSS including initialization of 33,505,280/40,992,768 bytes respectively. One run per condition on Apple M4 Pro, CPU only. Total time includes driver JSON handling, process startup and IPC. Stateless RSS was not measured here, so no memory reduction is claimed. Compact session JSON also removes indentation, contributing to output-byte differences. The helper's first page remains slower than raw search.
