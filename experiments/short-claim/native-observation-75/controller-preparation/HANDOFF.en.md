# Wrap/Marshal 75: preparing a controller for one attempt

This new controller leaves the original first-three controller unchanged. It uses only Go's standard library and was built once with Go 1.27.1, CGO 0, trimpath and buildvcs=false. **Neither this controller binary nor the original observer binary has been executed.** Preparation has 0 original API/init, role, label, model, fit or remote-publication executions.

The parent retains execution authority. Options are `--draft`, `--draft-sha256`, `--handoff`, `--handoff-sha256` and `--control-dir`. Actual paths are private arguments and never appear in diagnostics. The controller changes only the draft's single `Frozen: false` token to `true`, storing new bytes and SHA. It never overwrites the original draft or handoff.

Preflight reads and verifies caller-fixed draft/handoff hashes, 14 ordered source/license/module/document pins, the Want seal, equality of the 24 Wants and handoff Specification, and binary SHA/build metadata for Go/OS/architecture/CGO/trimpath/module replacements. Additional Go sources are refused. None of these checks invokes an original function or observer. Pins bind provenance; they are not independent proof of a trusted compiler.

The execution sequence is:

1. Exclusively create the outside directory, frozen plan, invocation and before-start ledger, then sync files and directories.
2. Verify the child directory is absent and create it exclusively. Exclusively store and sync the worker's outside-reservation. The controller does not create `results.json`; the worker reserves it.
3. Start one child through `/usr/bin/time -l` with only CPU 1, a 256 MiB soft Go heap setting and minimal PATH/locale. No credentials or inherited user environment are forwarded. The controller also sets CPU 1 and the same soft heap limit.
4. At 60 seconds, SIGKILL the process group and join exactly one Wait. No automatic retry. Return within exactly 60 seconds is not guaranteed after Kill/Wait.
5. Retain hashes of both final and partial bytes. Validate every saved record's ordered Probe/Want, Got-versus-Matches and panic/return/dispatch arithmetic against the fixed Wants. Completed observations with differences and all 24 Wants matching are distinct.

An empty result or refusal before Start never invents API counts. Start failure, abnormal exit after Start, timeout, valid partial prefixes and invalid final output are distinguished. A valid partial prefix can preserve counts after a corrupt final without hiding that final error. Dispatch reservation is intent, not proof of an in-flight return. Completed observations preserve differences and panics.

Darwin real/user/sys, maximum RSS and footprint fields are nullable. RSS uses bytes and covers child initialization, preflight, wrapper/API work and checkpoint I/O, excluding the controller. A soft Go heap setting is not a total RSS hard limit. Raw stdout/stderr, private plans and paths remain private; the safe ledger contains fixed diagnostic categories, counters, hashes and observations.

One synthetic/metadata-only race invocation (14 tests), two vet invocations and one final compile-only build passed. After race, only main's CPU/heap settings were added and checked by final vet/build. Stub lifecycle tests start no process. Reading source/Want/binary metadata executes no original initialization/API. The controller author does not claim independent upstream semantics, actual execution or Want accuracy. These 24 finite probes cover 2 behavior goals, not 24 independent parents. Training, caption, parent-label and broad-truth qualifications remain false.
