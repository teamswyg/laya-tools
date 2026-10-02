# Using the source reference runner for stage 60

`riido-sourceinventory` is a maintainer tool that connects 60 stored source definitions and 204 authored component relations to original Go declarations. It does not execute models or candidate functions. Whether a caption accurately describes an implementation belongs to the [separate content review](CONTENT-REVIEW-PROTOCOL-60.en.md).

## What is pinned before execution

The five original inputs are four Go source files and `results-56b.json`, with fixed full-file SHA and byte counts. Five compiled Go files are embedded in the binary and compared with current disk and Git bytes. Sixteen support files include this usage guide, build recipe, tests and content-review protocol. This list does not prove a complete semantic dependency closure.

Build with Go **1.27.1**, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, and `-p=1`. The actual binary SHA, size, platform and build information enter the plan; private binary/input paths do not. `GOMAXPROCS=1` and a Go heap soft limit of 256MiB are settings, not measured memory or speed.

## Prepare → freeze input → execute

Run these examples from the repository root. Replace `<...>` with actual experiment values. Keep the same binary and source commit between preparation and execution.

```sh
CGO_ENABLED=0 go build -trimpath -buildvcs=false -p=1 -o .cache/riido-sourceinventory-60 ./cmd/riido-sourceinventory
.cache/riido-sourceinventory-60 prepare --source-commit <source-freeze-commit> --out .cache/new-plan-60.json
```

`prepare` checks bytes and Git records for inputs, compiled files and support files without invoking original AST parsing, registry lookup, formatting or inventory. It verifies 26 Git blobs. Review the new plan and record it as `execution-plan-60.json` in a separate input commit, then use the same binary.

```sh
.cache/riido-sourceinventory-60 inventory --input-commit <input-freeze-commit> --out .cache/new-run-60/results.json
```

`inventory` rechecks inputs, plan, binary and 27 Git blobs, then projects only required strings from stored metadata. It does not convert large integers, correctness flags, rankings or training roles into arithmetic or features. It exclusively reserves a new directory and result file before **one** original rebinding call. Automatic retries are **zero**. Existing output paths and symlinks are rejected before that call.

## Reading success and failure

Successful scope requires four files, 60 roots in their original order, 204 component relations and agreement with historical hashes. Returned counters are compared against the predeclared expectation of four parses, 264 formatting calls, 112 normalization calls and 24 bundles. Expectations never populate observed counters. Raw declaration span hashes are computed for each relation; unique span counts do not replace physical calls.

A returned failure preserves partial results, actual counters and a fixed error code before exiting unsuccessfully. A claimed successful report with a scope mismatch also fails while preserving that report. A panic that prevents a return marks inner counters unavailable. Disk failure or process termination cannot guarantee a complete result file. There is no host-global lock; the experiment plan and ledger govern retries.

`metadata_rebound_content_review_pending` means byte/declaration correspondence succeeded. It does not imply object binding, complete closure discovery, caption approval, model quality or savings. `content_review=pending` and `training_ready=false` remain. Publish only safe result JSON and bilingual explanations; keep local binaries, paths and model weights out of Git.

## Post-execution checks

The first official plan allows one original inventory attempt and zero retries. Existing CI synthetic controls and historical frozen checks are separate. Future regression replays must be recorded separately, never retroactively included in the first-run ledger. Compare the original result's raw declaration SHA/spans, root/component order, formatted/normalized/bundle hashes and actual counters. Declaration correspondence does not supply semantic ground truth.
