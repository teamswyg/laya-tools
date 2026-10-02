# Outside controller for the first three-source observer

Private preparation is complete. Actual controller, observer, and original API executions remain zero. Only a stdlib Go controller and synthetic tests were built and checked; neither imports the original packages. The 21 fixed assets, Want, and binary were joined by byte hashes and metadata reads.

Root executes `first-three-controller` once. Its flags are `--draft`, `--draft-sha256`, `--handoff`, `--handoff-sha256`, and `--control-dir`. Exact expected hashes are in `CONTROLLER-HANDOFF.v1.json`; the outside output directory must not exist. This document is not an execution record.

The controller changes only one `Frozen` token from false to true. It fsyncs the new directory, O_EXCL reservation, and invocation ledger before starting the child through `/usr/bin/time -l`. The child receives CPU 1, a 256 MiB Go soft heap limit, and a minimal environment. At 60 seconds, the controller kills the process group and joins exactly one Wait. Kill/Wait latency remains visible; this is not a hard total RSS limit.

The ledger preserves final and partial hashes, nullable dispatch counters, and OS CPU/RSS. A corrupted final JSON cannot erase a valid partial prefix. Differences from the original Want remain even after all 24 calls complete; no automatic retry or relabeling occurs. The 24 probes are finite inputs for three behavior goals, not 24 distinct parent requests.

Ten synthetic tests, one metadata-read test, and one affected resource-parser test execution passed. Original imports/initialization/API executions, fitting, model calls, shared repository changes, and publication remain zero. Go test durations in the ledger are preparation records, not product performance measurements.

Controller author-side checks are distinct from independent observer mechanics review and the separate Want review. Compiler metadata and source/binary pins provide provenance, not a formal trusted-compiler proof. Caption, parent-label, generalization, and training qualification remain false. Raw logs and the original draft containing private absolute paths are not public assets.
