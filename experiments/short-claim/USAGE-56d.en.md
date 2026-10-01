# Checking the cost of a resident hint tool56d

[한국어](USAGE-56d.ko.md) · [Observed results](RESULTS-56d.en.md) · [Implementation boundaries](IMPLEMENTATION-RESIDENT-56d.en.md)

This development tool checks the cost of keeping a small hint process running while an agent sends requests. Inputs are public authored fixtures. Experiment56d records actual JSONL resource costs of **Go baselines without a learned model**. Suggested order is a hint about what to verify first; it neither removes candidates nor approves execution.

## Try the hint processor

Build from the repository root with Go1.27.1. Neither Python nor a GPU is required.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o bin/riido-shortclaim ./cmd/riido-shortclaim
bin/riido-shortclaim --baseline lexical_ordered < examples/shortclaim/order.json
bin/riido-shortclaim --stream --baseline lexical_ordered < examples/shortclaim/resident.jsonl
```

The first command processes one formatted JSON file; the second processes an actual single-line JSONL fixture. Send one JSON line and receive one response line in streaming mode. Continue sending lines to the same process; end input to stop. `verification_order` keeps every original candidate and status remains `unverified_heuristic`. Malformed JSON or candidate-limit violations are not converted into successful hints.

This measurement does not establish Laya ONNX inference costs or reduced Codex token usage. Codex and riido-daemon integrations are separate options; this command registers neither automatically.

## Prepare measurement data

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o bin/riido-residentperf ./cmd/riido-residentperf
bin/riido-residentperf --stage prepare --out .cache/my-resident-preparation
```

Use a new output directory. `corpus.json` contains exact transmission bytes and original-parent links; `preparation.json` contains counts, hashes and source pins; `build-recipe.json` fixes build settings. Preparation performs no child resource collection, new truth audit, training or model call. Normalized feature hashes select duplicate projections; they are not safe result cache keys for requests with different raw grammar.

## Read or reproduce the published replay

Read the [machine plan](execution-plan-56d.json), [linked corpus](wire-corpus-56d.json), [before-data freeze](freeze-56d.json), [raw observation](results-56d.json) and [summary](summary-56d.json) together. Source commit `60b03500fab417470d48d1a096be4ca1f43fafc2` and pre-replay plan commit `891e52d83e9aaf39c87cb5ef55c6d492724bebf1` were separately frozen.

The published plan binds **darwin/arm64 and both actual Go1.27.1 binary SHAs**. Other operating systems, toolchains or extra build options cannot reuse it unchanged. Build under the [fixed recipe](build-recipe-56d.json) environment and verify hashes before using this command shape:

```sh
bin/riido-residentperf --stage replay \
  --plan experiments/short-claim/execution-plan-56d.json \
  --plan-sha256 8c0b7a5b697e52d4afffbb5e4ca2f9b5d3c992a31d4eae4f9767bf57e2dbfed2 \
  --corpus experiments/short-claim/wire-corpus-56d.json \
  --binary bin/riido-shortclaim \
  --out .cache/my-separate-resident-replay
```

A new replay is a separate development observation. Do not pool it into56d or replace its first observation. Different bytes/environment require a separately frozen plan and version rather than editing hashes to force acceptance. The CLI checks commit syntax; the published separate freeze record reports verification of31 actual Git blobs at that commit.

Read phase counts, EOF, process termination, reader cleanup, errors and stream hashes alongside success counts. A failure leaves later rows unstarted, with no automatic retry. Existing output is never overwritten. This is neither a training nor an operational approval tool.

## What the numbers establish

On this Mac, child peak RSS was9.21875–10.515625MiB and per-row timed wire RTT p95 ranged0.021875–0.050333ms. Retain the first response577.13ms separately. Controller peak33.109375MiB includes preparation and measurement work. Adding these separate peaks does not establish simultaneous memory use.

Select48 original three-candidate parents from72; exclude the other24. Repeating these48 inputs25,080 times does not supply2400 independent golden requests. Every `narrow_rule` request in this corpus is outside its supported grammar and uses BM25 fallback. These observations establish neither successful narrow-grammar path costs nor semantic accuracy.

Caching, SIMD, concurrency and model experiments need separate comparisons. The current record provides resource costs for small nonlearned baselines; real task evaluations must separately establish LLM savings and model usefulness.
