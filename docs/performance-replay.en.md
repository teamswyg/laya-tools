# GitHub Actions performance replay

[한국어](performance-replay.ko.md) · English

Open [Performance replay](https://github.com/teamswyg/laya-tools/actions/workflows/performance.yml) and select **Run workflow**. Prefer reviewed code on the default branch. This does not run on a schedule or on every PR.

The current base INT8 model runs on CPU on standard Linux x64 and macOS ARM runners. Existing CI has already passed native inference on both platforms. GitHub documents standard hosted runners as free for public repositories; the listed Linux 16 GB and macOS ARM 7 GB memory exceed the approximately 1.4 GiB RSS measured locally. Shared runner performance varies: use actual reports. Artifact storage has separate policies; only small JSON and Markdown reports are retained for seven days.

- [Runner specifications](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)

## Coverage

| Category | Fixed inputs and outputs to inspect |
|---|---|
| Native inference | Cold load, five timed inferences after warmup, Go heap |
| Code search | Same keywords against this public repository, lexical vs Laya |
| Complexity routing | Easy, ordinary, difficult English tasks and Korean abstention policy |
| Model switch plans | Synthetic catalog downgrade and upgrade |
| Repository selection | Existing six synthetic repositories and 24 queries, lexical vs Laya |

Each OS runs 11 scenarios sequentially. Each scenario starts a fresh process, so total time includes model loading. Embedded native warm timings exclude loading. Repository per-request timings include the first inference. Native-error fallback to keyword-only execution does not count as success. A successful run does not certify that model classifications match intended difficulty.

Here, open loop means **fixed replay without adapting inputs, models, or thresholds to outputs**. This is not an arrival-rate open-loop load test. No automatic tuning, paid Codex calls, or unlimited retries occur. Coverage is the base checkpoint on CPU; the code checkpoint, GPU/CoreML, and concurrent service throughput are outside this suite.

## Reading results

Inspect the Actions Summary and `performance-<OS>` artifacts. Tables include wall time, process CPU time, and peak RSS. JSON embeds original command results, including abstentions, probabilities, and repository evaluation aggregates. CPU time sums work across cores and can exceed wall time. RSS includes Go and native allocations, but is not GPU memory. Raw pprof profiles are not uploaded.

The OS jobs use different hardware: do not compare their absolute numbers as performance regression gates. These are small development samples, not evidence of production accuracy, Codex savings, or meaningful p95 latency. Results are collected without speed thresholds blocking merges.

Only one runner executes at a time, with a 20-minute job limit per OS and a two-minute limit per scenario. Failed scenarios are recorded; the remaining fixed scenarios continue and the final job fails. Job timeout or build/download failure can prevent a report; inspect Actions logs.

## Local reproduction

Run from the repository root with Go 1.27.1, a C compiler, and the CPU runtime.

```sh
mkdir -p .cache/performance
go build -trimpath -o bin/riidolaya ./cmd/riidolaya
go build -trimpath -o bin/repoeval ./cmd/repoeval
go build -trimpath -o bin/perfsuite ./cmd/perfsuite
./bin/riidolaya setup
./bin/perfsuite --repeats 1 --output .cache/performance/report.json --summary .cache/performance/summary.md
```

`--repeats` accepts 1–3, never increased based on model answers. Use public/original fixtures only; never add real work prompts or private repositories to this public workflow.
