# Initial measurements — 2026-09-30

Hardware: Apple M4 Pro, 24 GiB unified memory, macOS 26.6.2 arm64. Production inference is the Go binary + ONNX Runtime 1.30.0. These are development measurements, not held-out accuracy or demonstrated Codex bill savings.

## Short decisions

One 47-token boolean decision, one warmup, 30 serial measured calls per CPU setting:

| Native CPU threads | Median warm latency | Session initialization |
|---|---:|---:|
| 1 | 40.29 ms | 667 ms |
| 2 | 32.71 ms | 647 ms |
| 4 (default) | 26.28 ms | 624 ms |
| 8 | 23.56 ms | 620 ms |

FP32 CPU (4 threads, 20 calls) median was 38.54 ms. INT8 is the working default; 8 threads was fastest in this short test, but 4 leaves more CPU capacity for the coding task. This is not proof that 4 or 8 is optimal for every sequence length or battery/thermal condition. Run the benchmark on your own machine.

The model file is about 572 MiB, compressed distribution about 455 MiB. Go HeapAlloc was roughly 12–14 MB. Whole-process peak RSS was about 1.39 GiB for INT8 and 1.48 GiB for FP32 in the separate `/usr/bin/time -l` runs. Do not substitute the Go heap number for application memory. A separate live RSS measurement is included in the raw JSON.

pprof confirms most CPU time lies in native calls (`runtime.cgocall` / unknown native frames). ORT profiling adds overhead; the latency table uses runs without profiling. Core ML MLProgram initialization failed with an axis/rank error on this dynamic exported graph. **No working GPU speed, GPU memory, or ANE utilization measurement is claimed.** Current CPU selection is the best working configuration among the tested paths, not a global comparison with every Metal/Core ML runtime.

## Code retrieval

Public corpus: encode/httpx 0.28.1 at commit `26d48e0634e6ee9cdc0533996db289ce4b430177`, package directory only (284,396 bytes). 18 questions were written before this Go implementation: 12 English, 6 Korean. Eight lexical candidates, fixed 32-line windows with 24-line stride, 50/50 normalized BM25 + Laya relevance, overlap suppression. No embedding search.

The initial strict metric is whether the returned window contains the target function's **definition line**. It can mark a useful window further inside the right function as a miss; it is not semantic answer correctness. Questions are development fixtures, not a random sample or held-out set.

| Method | English definition-line Hit@3 | English Hit@5 | English median query time | Korean Hit@3 |
|---|---:|---:|---:|---:|
| BM25 | 3/12 | 8/12 | 0.25 ms | 2/6 |
| BM25 + base Laya | 3/12 | 8/12 | 1,760 ms | 3/6 |
| BM25 + Laya Code r1 | 5/12 | 8/12 | 1,641 ms | 2/6 |

Definition-line Hit@1 was 0/12 English for all three methods. This stringent metric and fixed windows need a better relevance evaluation before product claims. Some windows truncate at 512 tokens, and two Korean-only queries produce no lexical candidates. A reranker cannot fix missing candidates. Indexing and model startup are recorded separately in raw results; these query times use a resident model and existing index, whereas the CLI's cold path also loads/indexes.

The observed benefit is limited: Code r1 improved top-3 definition coverage in this tiny set but added substantial latency. For literal identifiers or exact strings, prefer `--lexical` or ordinary grep. No end-to-end Codex token/cost/runtime comparison was run.

## Router smoke test

Twelve original requests range from comment typos to multi-region migration and include one Korean request. With the conservative 0.9 default threshold, the initial run retained the strong/default tier for all 12 requests. This demonstrates the fallback path; **it demonstrates zero model downgrades and no cost savings**. Lowering the threshold is an explicit experiment, not a justified production recommendation. Dedicated difficulty labels, independent tasks and downstream completion tests are needed before selecting a useful threshold.

## Reproduce

```sh
git clone --branch 0.28.1 --depth 1 https://github.com/encode/httpx.git /tmp/httpx-eval
go run ./cmd/eval --root /tmp/httpx-eval/httpx
# Add --model-dir and --runtime to evaluate an installed checkpoint.
laya bench --threads 4 --iterations 30
laya serve < benchmarks/router-queries.jsonl
```

Raw sanitized results are in [benchmarks/results](../benchmarks/results). They contain no local usernames, credentials, private source excerpts, or pprof files. Profiles remain local because they can contain host paths.
