# laya-tools

Local Laya inference in a Go binary: bounded code search and **experimental, opt-in Codex model routing**. Built first for Apple Silicon macOS, with Linux amd64 CI and releases. No Python process, model API, account, or API key is needed at runtime.

This is an early prototype, not a demonstrated Codex cost reduction product. Laya classifies decisions; Codex still does the coding. See [measurements](docs/measurements.md) and the [Korean design explanation](docs/design.ko.md).

## Quick start

Download the matching `laya-v…-darwin-arm64.tar.gz` or `linux-amd64` archive from [releases](https://github.com/teamswyg/laya-tools/releases), check it against `SHA256SUMS`, and extract it. Or build with Go 1.27 and a C compiler:

```sh
go build -trimpath -o bin/laya ./cmd/laya
./bin/laya setup
./bin/laya search --root /path/to/repository --json 'where are redirect headers removed?'
```

`setup` downloads SHA-256-pinned public weights and ONNX Runtime. Initial model download is about 455 MiB compressed / 575 MiB installed. It verifies both the archive and installed files. Subsequent search/routing is local. macOS cache: `~/Library/Caches/laya-tools`; Linux: `~/.cache/laya-tools`. Override with `LAYA_CACHE`.

One Go executable orchestrates the work; the cached native inference library and model files remain separate. This is not a self-contained 10 MB AI model. See measured whole-process memory below.

## Code search

```sh
laya search --root . 'where is the retry policy implemented?'
laya search --root . --lexical --json 'retry timeout'  # no model loaded
laya setup --checkpoint code
laya search --checkpoint code --candidates 8 --limit 3 'redirect authentication'
laya search --candidate-query 'gzip decoder' 'gzip 압축을 해제하는 코드'
```

Git supplies tracked/untracked and ignore-aware file listing at repository roots. Install `rg` for non-root directories. Hidden files, common generated folders, symlinks, obvious credential filenames, binary files, and files over 256 KiB are skipped. Only listed code/text extensions are read. Limits: 32 MiB source, 50,000 chunks, 64 candidates, 512 model tokens per candidate. Results include paths, exact line ranges, excerpts, lexical/relevance scores, and a truncation flag. `--candidate-query` only changes lexical retrieval; Laya still reads the original question.

BM25 retrieves candidates; Laya reranks them; overlapping output windows are suppressed. No embeddings database or background index is created. Each request reindexes the current files, avoiding stale edits; the model stays warm in `serve`/`mcp`. Missing native assets fall back to lexical search with an explicit warning. A zero lexical match cannot be recovered by reranking. English Laya's Korean accuracy is unvalidated.

## Optional router

Nothing changes your Codex configuration automatically. Model IDs are supplied by the user; no price assumptions or hardcoded model catalog.

```sh
export LAYA_FAST_MODEL='your-fast-model-id'
export LAYA_STANDARD_MODEL='your-standard-model-id'
export LAYA_STRONG_MODEL='your-strong-model-id'
laya route --json 'Fix a spelling mistake in this comment'
laya codex --dry-run 'Investigate a concurrency bug'
laya codex 'Investigate a concurrency bug'
laya codex --model 'your-explicit-model-id' 'Implement the feature'
```

`route` recommends only. `codex` chooses once for a **new** interactive CLI session, releases Laya memory, then launches your installed `codex --model … -- PROMPT`. Empty strong-model settings preserve Codex's configured default. Explicit model selection wins. It never rewrites credentials, proxies API traffic, alters sandbox/approval settings, or switches a running conversation's model. The resulting Codex session communicates with its normal provider.

Low confidence (default 0.9), truncated inputs, inference failure, missing tier models, and Korean requests preserve the strong/default model. Probabilities are not calibrated task-success guarantees. The base checkpoint is recommended for routing; `code` is a relevance finetune. Automatic quality-based retries and cost accounting are future work, not implemented features.

## Humans and agents

Human output by default; stable JSON with `--json`; errors go to stderr with nonzero exit status. Put flags before the query. Pass `-` to read a prompt from stdin.

```sh
laya serve --root /path/to/repository
# send one JSON object per line; one response per line:
{"id":1,"op":"search","query":"redirect authentication"}
{"id":2,"op":"route","query":"Fix a typo"}
```

`serve` keeps one model in memory and processes requests sequentially. `laya mcp --root /path/to/repository` exposes stdio tools `search_code` and `route_model`. Optional Codex registration (uses an absolute executable path):

```sh
codex mcp add laya -- /absolute/path/to/laya mcp --root /absolute/path/to/repository
```

Registration is opt-in. The running model may use the search tool, but the routing tool can only recommend a model for a new task. No daemon, launch agent, open network port, or app settings are installed automatically.

## Measure, don't assume

```sh
laya bench --iterations 30 --threads 4
laya bench --cpu-profile cpu.pprof --heap-profile heap.pprof --ort-profile ort-trace
go tool pprof -top cpu.pprof
/usr/bin/time -l laya bench --iterations 30  # macOS process peak RSS
```

Go pprof covers Go allocations and CPU samples; native inference often appears as `runtime.cgocall`/unknown. It does **not** report GPU memory or all native allocations. ORT trace records actual provider execution; Instruments is needed for GPU counters. Do not publish raw profiles from private workspaces: profiles may contain local paths. `--provider coreml` is experimental and currently fails for the dynamic exported graph on the tested Mac. CPU INT8 is the working default. The FP32 graph is a developer export, not downloaded by normal setup.

## CI as the approval gate

[CI](.github/workflows/ci.yml) runs race-enabled tests, formatting, vet, real native model inference/tokenizer parity on Linux and macOS, and a redacted secret scan. The `quality` status is required by branch protection; reviewer count is zero. Trusted same-repository PRs queue automatically for squash merge after checks. Fork PRs run unprivileged tests and are not auto-approved. The privileged automerge workflow never checks out PR code. CI failures stop the loop for another code/test iteration.

Version tags run fresh tests and native inference before publishing binaries and checksums. There is no human approval environment. CI is a deterministic gate, not an AI correctness proof. Public model releases are separately pinned in [the manifest](internal/assets/manifest.json).

## Development

```sh
go test -race ./...
go vet ./...
go test -bench . -benchmem ./internal/search
```

Native tests additionally require `LAYA_MODEL_DIR` (contains `model.onnx`, `tokenizer.json`, `config.json`) and `LAYA_RUNTIME`. These are set in CI after `setup`. For reference export, see [build instructions](docs/model-build.md). Python is used **only by model maintainers for export/reference comparison**, never by the installed Go CLI.

Apache-2.0. Third-party sources and model revisions: [NOTICE](NOTICE).
