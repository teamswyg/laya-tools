# Independent review of the Native2 once-controller

This is a static review of the outside controller intended to run **one later** finite set of 23 original Go observations. Two concrete implementation problems were corrected in v4; no remaining concrete source blocker was identified. Eligibility to execute still depends on Root separately freezing local rights, source selection, CI, resources and the execution plan. This report is neither that authorization nor approval of Want semantics or evidence that original observations succeeded.

The reviewer performed zero original builds, starts, initialization or API observations, zero production-controller builds/starts, zero fake-child runs, zero model operations, zero new HTTP requests and zero external publications. A second independent reviewer kept the same boundary. We read the author's fake-test logs without rerunning them. The existing first20 contract-review and Native2 observer-review seals were unchanged.

## Why freeze the plan in two stages?

Putting a not-yet-created executable hash into an earlier plan, then embedding that plan hash back into the executable, would make each hash depend on the other. This implementation uses a directional sequence:

1. Root's **static reservation** references the already frozen source closure, vectors, source/local-rights review, CI evidence, resource limits and three directories. It excludes future worker, NativePlan, Readiness and final-binding hashes.
2. Readiness and NativePlan reference the reservation. The worker binary declares the frozen closure/vector/readiness hashes in its build settings.
3. The final **execution binding** connects the actual reservation, plan, worker/controller binaries and controller source. A self-binary hash is not embedded back into its own source.
4. The controller checks the links and caps, exclusively creates and syncs `start.once`, checks again and makes one Start attempt.

No cycle was found in this sequence. Unknown fields and duplicate JSON keys are rejected. Model, protected-evaluation and public-write capabilities remain false; the scheduled count is fixed at 23.

## Controls checked in the source

| Area | Source behavior checked | Scope |
|---|---|---|
| Preflight | Checks closure/vector/Readiness/NativePlan and related file bytes/SHA, source inventory and declared Go build settings | Reads files and metadata without original imports or initialization |
| Directories | Absolute clean paths, no symlink ancestors, physical separation of source/attempt/outside and controller-source/output | Compares existing directory inodes; fresh output leaf names are lowercase ASCII |
| Once-only start | Fresh directories, O_EXCL files, file and directory sync, then a second preflight and direct worker `--plan` execution | No shell, help invocation or automatic retry |
| Start/reap | Failed Start means Wait0; successful Start means Wait1. Timeout/output errors request group kill and collect Wait | A main failure after Start does not prove original init0 |
| stdout | Counts/hashes received bytes, stores0 raw bytes, requests termination after exceeding12MiB | Extra delivered bytes are still counted; bytes not delivered after forced pipe closure can be unknown |
| stderr | Counts/hashes all received bytes, retains a private prefix≤64KiB with prefix SHA and explicit truncation/I/O facts | Private log bytes are not copied into public evidence |
| Worker result | Checks/hashes the fence and result in the worker directory; copies0 raw/partial result bytes outside | Does not interpret result JSON semantics or the correctness of23 Wants |
| Resources | CPU1 and Go soft heap256MiB environment before startup,300s timer, post-Wait whole-child wall and Darwin MaxRSS>0/≤256MiB gates | Does not substitute Go heap/GPU telemetry; RSS is not a hard limit |
| Receipts | Outside metadata/logs≤1MiB, failed/partial files retained, no automatic cleanup/retry | Storage failure has a separate exit; Root must gate on exit and readback together |

Each stream writer owns its state in one goroutine, and snapshots are read after `Cmd.Wait`, so this path does not need locks. Fake-child tests re-execute only the Go test binary and do not import original packages or initialize models.

## The two findings and their fixes

First, the actual observer opens its result file with O_EXCL and then performs its first write. A zero-byte file is normal during that interval. The earlier running poll rejected size0 and could consume the one permitted observation by terminating a valid worker. v4 distinguishes **size0 allowed while running / nonempty, fence and hash required after completion**. A fake-file regression follows the same Open→Write sequence.

Second, string-only path comparison can miss that `source` and `SOURCE` refer to the same location on a case-insensitive macOS volume. v4 compares existing endpoints with opposite directory ancestors by inode, requires lowercase ASCII for fresh output leaf names and rejects aliases with the same physical parent. This was a gap in the promised physical separation, not evidence that Root's actual chosen paths overlapped.

The author's v1/v2/v3/v4 logs contain12/13/14/15 passing top-level tests respectively. Original observations remain0 in all four runs. The four fake-child Start/Wait pairs in each run and deliberate failed Starts in v2–v4 are controller tests, not scientific retries. The15 v4 passes do not establish production preflight, actual compiler inputs, original initialization or Want semantics.

## Conditions that remain owned by Root

Root must read the **contents** of the CI, rights and resource evidence and freeze that interpretation in Readiness. Those proof bodies must also remain static, without back-references to future worker, NativePlan, Readiness or final-binding hashes. The controller checks exact IDs, bytes and SHA for those files; it does not independently interpret their CI status or legal/resource meaning. The true declarations represent Root's responsibility for choosing and reading the correct evidence.

Root must record actual compiler selection, build command/environment/tool inputs and pin the worker/controller binaries. Hashes and `debug/buildinfo` declarations establish useful connections but are not compiler-independent attestation of compiled inputs. Rechecks detect ordinary changes; they do not prevent hostile concurrent changes through the kernel exec boundary. Root must retain ownership of frozen files and paths.

Once the marker exists, post-marker pin rejection, Start/Wait failure and receipt failure remain consumed, preserved attempts. Do not erase the marker or automatically retry with a new name. If the outside process is killed, only synced intent may remain and child completion may be unknown; Root must inspect that state. An invalid CLI/input before a safe output location is established can return exit2 without writing a receipt, so the parent invocation ledger is also necessary.

The300s cap is the successful whole-child wall gate; failed cleanup is not guaranteed to return within300s. Darwin MaxRSS measures the direct child after Wait, not the total of every descendant, the controller or GPU memory. `GOMAXPROCS=1` is not OS CPU affinity. Failed result files are not copied outside; any needed post-failure inspection must observe state/bytes at the original worker output location.

## Alternatives

1. **Current v4 plus Root-frozen evidence, build receipts and one-start plan** best fits these23 finite observations. Keep the two fixes and explicit evidence responsibilities; Root owns actual execution.
2. **Parse every proof's meaning inside the controller** may suit broader automation but requires additional frozen CI/rights/resource schemas. No new need for that framework is established here. Avoid representing a pin check as semantic validation.
3. **Defer execution for stronger isolation or kernel/toolchain attestation** fits hostile mutation or aggregate-descendant resource requirements. Treat it as a separate experiment broader than this local one-start contract.

The accompanying JSON records source anchors, pins, regression evidence and remaining conditions. This static review produced no original-observation or performance-improvement measurements.
