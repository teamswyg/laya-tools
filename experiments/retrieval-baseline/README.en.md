# Retrieval baseline 07: 2,948 distinct questions expose a tradeoff

[한국어](README.ko.md) · [Frozen plan](plan-07.json) · [Results](results-07.json) · [Reproduction and correction](reproduction-07.json)

**Identifier splitting helps descriptions without a function name, but reduces overall top-1 accuracy. Do not replace the baseline globally.** This experiment actually searches 2,948 distinct descriptions, not 24 questions multiplied by candidate counts or repeated runs. It is a documentation proxy from three repositories, not an independent final evaluation on 2,400 real user questions.

## Comparison and correction

The catalog contains 3,009 Go functions. Only code is indexed: no target description or repository path is appended. All functions attached to the same normalized description are known targets; other candidates remain unjudged. Candidate order is fixed by repository/path/code hash, independently of query text, target position and scores.

- `raw`: existing `pkg/hintsearch` BM25 with whole identifiers.
- `identifiers`: symmetric camelCase/snake_case splitting of code and query before BM25, without a learned model.

Both methods were fixed without training or parameter search. Plan and evaluator draft were committed as `33bd500` before the first corpus ranking. The first run exposed an evaluator defect: a lowercased query deduplication key was used as the actual preprocessing input, while documents retained case. That made identifier splitting asymmetric. The initial run is invalidated and recorded; commit `2b44d9e` fixes it with a regression test. Results below use the corrected evaluator. This corpus is now observed evidence, not a fresh final test for subsequent tuning.

## Results

Recall@10 is the fraction of questions with any known target in the first ten candidates. MRR averages the reciprocal of the first known-target rank; higher is better.

| Questions | Method | Recall@1 | Recall@10 | MRR | Mean first-target rank |
|---|---|---:|---:|---:|---:|
| All 2,948 | Raw BM25 | 41.79% | 74.83% | 0.5336 | 48.99 |
| All 2,948 | Identifier splitting | 39.21% | 77.27% | 0.5262 | 30.00 |
| Name present, 2,604 | Raw BM25 | 44.51% | 78.49% | 0.5640 | 19.12 |
| Name present, 2,604 | Identifier splitting | 40.48% | 79.22% | 0.5412 | 16.08 |
| Name absent, 344 | Raw BM25 | 21.22% | 47.09% | 0.3038 | 275.12 |
| Name absent, 344 | Identifier splitting | 29.65% | 62.50% | 0.4121 | 135.39 |

Name strata use the known target name at evaluation time. **A production router must not be given that oracle information.** Results on 344 name-absent questions are not evidence from 2,400 name-absent questions. Per-repository results are in JSON. No interval assuming all questions independent or broad repository-generalization claim is made.

## Source checks and exclusions

Downloaded commit-pinned ZIPs for three repositories whose root notices were checked previously. Verified all 3,095 code fragments verbatim against 717 source files without executing or extracting them.

Sequential, mutually exclusive exclusions: 63 generated-code markers, 14 nested-notice cases, eight other-license headers and one whole query occurring in code. Remaining 3,009 candidates correspond to 2,948 normalized distinct descriptions. Findings include lxd `shared/log15/LICENSE` and BSD-related headers in lxd `shared/cert.go` and etcd `pkg/pathutil/path.go`. Exclusion does not mean those licenses are unusable; they fall outside this experiment's uniform review scope.

Results record revisions, archive hashes and root-notice hashes. For files without headers, the checked root notice is source evidence, not exhaustive legal clearance. Generated-code detection and semantic-clone coverage remain incomplete. Source rows stay local. Documentation/name bias and concentration in three repositories persist. Neither CoSQA reserve was scored or used for training.

## Resources and reproduction

On Apple M4 Pro/Go1.27.1 CPU, the full process took 1.74s wall/1.41s user CPU, peak RSS 94,519,296 bytes (about 90.1MiB). This includes ZIP/source loading and checks, both indexes and all evaluation; it excludes downloads, Parquet projection and compilation. It is not steady-state server memory.

Query preprocessing plus ranking p95 was approximately 0.206ms raw and 0.209ms with splitting. Index build timings include code preprocessing. Target verification, JSON output and source loading are excluded from individual query latency. These measurements cover this 3,009-document catalog and machine, not large-repository performance or LLM savings. Replay reproduced all quality metrics and membership; timing varied.

Prepare the [pinned projection](../retrieval-audit/README.en.md), then the archives:

```sh
mkdir -p .cache/retrieval-source-07
curl -fL 'https://codeload.github.com/kubernetes/test-infra/zip/8125fbda10178887be5dff9e901d6a0a519b67bc' -o .cache/retrieval-source-07/test-infra.zip
curl -fL 'https://codeload.github.com/lxc/lxd/zip/7a41d14e4c1a6bc25918aca91004d594774dcdd3' -o .cache/retrieval-source-07/lxd.zip
curl -fL 'https://codeload.github.com/etcd-io/etcd/zip/616592d9ba993e3fe9798eef581316016df98906' -o .cache/retrieval-source-07/etcd.zip
go run ./cmd/riido-retrievalbench --stage audit
go run ./cmd/riido-retrievalbench --stage evaluate
```

Unexpected source/plan/membership hashes or counts fail closed. A GitHub archive serialization change does not justify ignoring a checksum mismatch. Archives are bounded to 50MiB and read members to 8MiB; no extraction occurs. Offline source preparation uses maps; ranking retains the existing immutable array index.

## Next hypothesis

This nonlearned comparison suggests a target for a tiny claim supplier: predict whether an auxiliary search could help now. A separate experiment can learn from signals available at runtime, preserving baseline candidates when hints fail. Never feed gold name strata or evaluation results into runtime inputs. Independent questions, repository separation and actual end-to-end task/time/usage evidence remain necessary before automatic adoption or usefulness claims.

A CPU pprof was collected from the actual full evaluation. Of 1,330ms sampled, `Rank` accounted for 42.86% cumulatively and `runtime.madvise` for 29.32% flat. This short single sample is not a speedup estimate. Next inspect caller-owned reusable rank buffers and sorting cost before SIMD. Raw profiles stay local. Use `--cpuprofile .cache/new-cpu.pprof` with a new local path; profiled quality metrics also matched.
