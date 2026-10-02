# Using the stored-truth utility tool

[한국어](USAGE-STORED-58.ko.md) · [Results and interpretation](RESULTS-STORED-58.en.md)

`riido-storedutility` is a Go maintainer tool for checking candidate-ordering headroom. It ranks public request/candidate prose with four controls and compares with previously stored truth, requiring neither model weights nor Python. It executes no candidate source or Laya model. It does not activate routing or replace human approval.

## Easiest verification

From the repository root with Go 1.27.1:

```sh
go test ./internal/storedaudit ./cmd/riido-storedutility
```

Tests verify frozen input, plan and original-result hashes, then replay nonlearned rankings. CI runs the same tests on Linux/macOS. These repetitions add no official experiments or independent requests. They do not execute the historical Darwin binary on Linux.

In [results-58.json](results-58.json), `evaluation.metrics` contains control totals, `group_metrics` contains group costs, and `rows` contains per-request orders, acceptable candidates and unknown reasons. `possible_relative_gain` is an answer-knowing upper bound, not achieved performance. Read `training_ready=false` and `stop_reasons` alongside it. The narrow rule's 72/72 fallbacks mean every request used BM25.

## How official collection was produced

This describes the completed collection procedure. It requires separate actual 40-hex source and input commits. Retain the same executable from plan generation through evaluation and commit the plan before ranking.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o PRIVATE_BINARY ./cmd/riido-storedutility
PRIVATE_BINARY plan --repo . --source-commit SOURCE_COMMIT --out NEW_PLAN_FILE
# Copy NEW_PLAN_FILE bytes to execution-plan-58.json and freeze a separate input commit
PRIVATE_BINARY audit --repo . --input-commit INPUT_COMMIT --out NEW_RESULT_DIRECTORY
```

Uppercase names such as `PRIVATE_BINARY` are explanatory placeholders. The historical plan pins the actual Darwin arm64 executable hash, so another OS/build cannot rerun official evaluation against it. Use regression tests for ordinary verification. New experiments need new versioned plans, inputs and results without overwriting phase 58.

`plan` verifies metadata binding with zero ranking calls. `audit` checks actual Git bytes for nine sources, 13 support files, three inputs and one plan, reserving a new directory and result file before evaluation. Preserve partial failure records. Path reuse protection is not a host-wide duplicate-execution guarantee; maintain the separate one-attempt/zero-retry official ledger.

Keep every candidate and fallback. Never convert unknown into incorrect/no-answer. Truth is not passed to scorers; only request/candidate prose supplies features. See the [pre-observation plan](PLAN-STORED-58.en.md) for settings and file limits. Do not infer actual CPU/GPU memory or cost savings from this record.
