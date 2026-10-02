# Public behavior auditor usage57

`riido-publicaudit` is a Go maintainer tool for collecting evidence about small public behaviors. Humans and agents can inspect inputs, errors and source revisions. It does not choose models or approve code edits. General users continue using `riidolaya`/`riido-shortclaim`.

## Verify published observations

From the repository root with Go1.27.1, replay actual APIs and compare preserved outputs, sources and expectations on Linux or macOS. This does not run the historical Darwin binary on another OS or create official samples.

```sh
go test ./cmd/riido-publicaudit -run TestPublishedFiniteAuditReplay -count=1
```

Raw UTF-8 and hex in [probes](probes-57.json) must agree. [Original observations](results-57.json) distinguish normative, descriptive API and stronger policy expectations. `source_read_hypothesis` is a pre-observation prediction, never acceptance criteria. The tool does not fill model prose with derived outcomes or hidden source IDs.

## Official collection boundary

The original collection used this sequence. Build with Go1.27.1 Darwin/arm64, CGO0 and the three flags in the [recipe](build-recipe-57.json); the executable must match the plan's SHA. Existing files/directories are refused. Fresh output is not a global execution-count gate; the official one-attempt record belongs to the external experiment ledger.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o ./riido-publicaudit ./cmd/riido-publicaudit
./riido-publicaudit plan --repo . --source-commit 9e99914f6d1e35aa9d97413f80fc6e03768e97ce --out NEW_PLAN_FILE
./riido-publicaudit audit --repo . --input-commit 1462c705087f24666049f8bd31d317490c05dbeb --out NEW_RESULT_DIRECTORY
```

`audit` reads fixed repository `execution-plan-57.json` and `probes-57.json`. Creating a plan does not automatically replace them. Original frozen commits must be available locally; a checkout containing only squashed main history needs to fetch those public commits separately. Another OS/toolchain/binary requires a **separate experiment** with new plan/input freeze, preserving historical observations.

Only `strict_parse`, `compare`, `glob_match` and `glob_validate` are supported. No arbitrary source execution or filesystem globbing. Invalid UTF-8, oversized inputs, unlisted source, tampering, duplicate keys, extra fields and missing input freeze fail before observations. Known library errors are observations; panic/unclassified errors remain unknown.

Expectation disagreement is a valid research outcome and does not itself fail the process. Incomplete/unknown observations and storage failures return nonzero. Storage-failure diagnostics retain the already-started observation count; no automatic official retry occurs. This tool is not a CPU/RSS/GPU/latency profiler.
