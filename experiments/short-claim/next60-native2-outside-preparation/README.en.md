This prepares an **outside-process gate** for one finite original-function observation attempt: IPNet 3 candidates × 5 inputs and OrCompose 2 × 4, totaling 23 observations. Rows are not counted as 23 new semantic requests or training labels. Original worker builds/starts/initializers/functions, model APIs, protected reads, external publication and production execution are zero.

In-main readiness cannot stop imported initialization that occurs before main. This controller therefore checks source, vectors, CI, local rights, resources, directories and binaries before a single start. It imports only standard library and the existing authored pure protocol. `source/` contains inert copies; a separate private module contains the actual authored source. No original upstream package or model is imported.

The pin graph is intentionally one-way.

1. Root first freezes a static reservation containing closure/vector, static source/local-rights review, CI, resources and fixed directories. It excludes worker, NativePlan, Readiness and execution-binding hashes.
2. Readiness references the reservation SHA; NativePlan references Readiness. Root later builds the worker with three declared bindings.
3. A later execution binding attaches actual worker/controller binaries, NativePlan and source hashes. Its hash never feeds back into Readiness or reservation.
4. After full verification, an exclusive `start.once` and its directory are fsynced before invoking the exact worker with only `--plan`.

The source/CI/resource proof files must themselves avoid future worker/Readiness/NativePlan dependencies; Root checks this indirect cycle when selecting evidence. The controller verifies proof **bytes**, not the independent meaning of CI suitability or licensing conclusions. Root owns those source-bound interpretations.

| Boundary | Fixed production scope |
|---|---|
| Worker start | At most once at fixed reservation output paths; automatic restart 0 |
| Start/Wait | One Wait after successful Start; failed Start has Wait 0 |
| Go concurrency / soft heap | GOMAXPROCS=1 / GOMEMLIMIT=256MiB |
| Whole-child time / observed RSS | 300s / Darwin MaxRSS 256MiB |
| Result / outside file cap | 12MiB / 1MiB |
| Raw stdout / private stderr | Raw retention 0 / up to 64KiB prefix |
| Model / protected / public writes | Allowed 0 |

All absolute ancestors are checked for symlinks. SourceRoot, AttemptDir and outside are pairwise separate. Existing ancestor inode comparisons supplement pathname checks to detect physical case aliases. Fresh output leaf names use lowercase ASCII letters/digits/`-`/`_`, excluding absent-name case/Unicode-normalization aliases. Controller source must also be physically separate from outputs. An existing output directory is never reused.

Before starting, all source-closure pins, actual frozen draft fields, the 23-observation schedule and Readiness are checked. Future binary metadata checks require Go1.27.1, darwin/arm64, CGO0, trimpath, no VCS data and exactly three worker `-X` declarations. Additional worker linker flags require separately reviewed source/version changes. Declared settings and hashes are not compiler-independent or kernel execution-inode attestation. Root must instrument actual build-tool/cache closure later.

A separate process group permits SIGKILL requests on timeout, output limits or interruption. One Wait joins pipe drainers. Stdout is counted/hashed without raw retention; only a private stderr prefix is saved. Worker results/partials are never copied. After exit, result size/hash and its reservation-bound inner dispatch fence are read. Child completion does not validate JSON meaning, labels, correctness or model quality.

An in-progress `results.json` may have size zero between exclusive creation and first Write. Polling permits this normal window; after exit, output must be nonempty and satisfy cap/hash checks. The initially incorrect rejection and subsequent regression fix are preserved in `AUTHORING-LEDGER.v1.json`. Existing original controller and native2 observer seals remain unchanged.

Tests re-executed only this stdlib/authored test binary as fake children. Final 15 top tests PASS cover success, group timeout kill, stdout overflow, stderr prefix, failed Start/Wait0, open-before-write empty output and physical-path gates. Short fake-test limits are not exposed through production plans. Standalone production controller builds and original worker builds/starts remain zero; Go test compilation and fake execution are counted separately.

Darwin wait4 MaxRSS observes the direct child's whole startup/pin/observation/write lifetime as a post-wait gate. It is not a hard RSS limit, exhaustive group/native/GPU attribution or Go heap. GOMAXPROCS is not OS affinity. Failed cleanup can extend past 300s while pipes drain; successful whole-child wall must remain ≤300s. Controller SIGKILL may leave only a synced once record, with child completion and final totals unknown. There is no automatic restart.

Public templates are non-executable with false/empty pins. Preparation and resource plans do not activate execution, training, publication or a model. Root owns source/CI/local-rights/resource freeze, measured binary sizes, production builds and any single original observation. Later actual results must append a separate receipt without rewriting this historical zero-execution snapshot.
