# Using the behavioral truth auditor

[한국어](USAGE-56b.ko.md) · [Pre-observation plan](PLAN-56b.en.md) · [Results](RESULTS-56b.en.md) · [Runtime hint tool](USAGE-56.en.md)

`riido-typedaudit` is a maintainer Go tool for preparing learning evidence. It compares observed values, error identity, input mutation, memory ownership and acquire/cleanup ordering against independent finite literal tables. Incomplete code descriptions remain `unknown`. Requests sharing sources, candidates, counterexamples or helpers are connected transitively.

The auditor is not a model router or ranking model. Runtime hints belong to the separate `riido-shortclaim` tool. This audit downloads or trains no model and changes neither Codex integration nor default settings.

Build from the repository root using Go1.27.1. No Python or API key is required.

```sh
mkdir -p bin .cache
CGO_ENABLED=0 go build -trimpath -o bin/riido-typedaudit ./cmd/riido-typedaudit
```

Preparation writes a new fixture from authored requests fixed in source. It executes no candidates, group audit or ranking. Use a new directory name: an existing output directory is rejected.

```sh
./bin/riido-typedaudit --stage prepare --out .cache/typed56b-preparation-example
```

`probes.json` contains original text and fixed source/bundle digests. There is no automatic label generation for arbitrary work code. Known labels require the exact reviewed request/candidate captions and source bindings. Fitting the caption budget does not establish semantic completeness.

The [results document](RESULTS-56b.en.md) provides the official replay command and exact plan digest. Auditing verifies the plan, source manifest, compiled sources and both input files, then decodes the same verified input bytes. It writes legacy48-request truth, new24-request truth and combined72-request relations to a new `results.json`, without overwriting existing files or output directories.

`whole_group_gate_pass` differs from `labeled_group_gate_pass`. Unknown-only groups retain their relation edges but cannot satisfy the labeled minimum. Even if both pass, `training_execution_ready` remains false: this plan assigns no roles and authorizes no fitting, weights, model calls, final evaluation or production activation.

CPU execution uses1 thread with a256MiB Go-heap soft limit. This is neither a total-process RSS/GPU memory cap nor a measured inference result. Distinguish offline `go/types` auditing, which needs the Go installation's standard-library type metadata, from the small fixed-array runtime hint tool. Do not put this maintainer auditor in the path for repeated low-cost production requests.

People can read public unknown reasons and group relations. Agents can inspect fixed error codes, exit status and JSON. Errors do not echo input, caller-selected personal paths or code. Raw profiles, real work inputs, credentials and weights do not belong in Git.
