# Actual observations for the first two native requests

Two requests from the existing 20-request draft and nine frozen inputs were observed across five original Go candidates in one child process: 23 observations, one Start, one Wait, zero exit, zero panics and zero unknowns. They are distinct within this run; they are not two newly authored requests or two new source families. PR108 passed all four required CI jobs on attempt 1 and merged automatically; the checked head and merge trees match.

| Request | Candidate | Finite inputs satisfied |
|---|---|---:|
| Parse CIDR into network state and preserve prior state on error | `ipNetValue.Set` | 5/5 |
| Same request | `ipValue.Set` | 0/5 |
| Same request | `ipMaskValue.Set` | 0/5 |
| Alternative hook composition stopping at first success | `OrComposeDecodeHookFunc` | 4/4 |
| Same request | `ComposeDecodeHookFunc` | 0/4 |

These are nine satisfied observations and fourteen mismatches by other candidates, **not 9/23 model accuracy**. Twenty-three observations are not 23 independent requests. Saved results were checked against original source separately. The [source review](../next60-native2-actual-semantic-review/REVIEW.en.md) discloses that its author also wrote the comparison script and was not blinded.

The behavioral distinctions matter for small claim models. `OrCompose` stops when an error is absent even if the returned value is `nil`. Empty hook composition produces a real nonnil error with an empty message. `Compose` stops at the first error and can pass a preceding `nil` return into the next hook as an invalid source value. Function names or error strings alone miss these distinctions. `Flag.Changed=false` after direct `Value.Set` is diagnostic and was not added as a new Want criterion.

Whole-child wall time including startup and pin reads was 0.37042475s; child CPU was 0.008281s user and 0.009383s system. OS maximum RSS was 9,551,872B (about 9.11MiB). Go heap after observation was 3,053,392B (about 2.91MiB). The 0.000144334s observation segment excludes startup/pins and is a single measurement, not a latency distribution or throughput result. The separate controller OS RSS of 12,091,392B is not added to the child RSS. GPU use was not measured and model inference count is zero.

Explicit upstream entrypoints were called 76 times and authored observation callbacks 9 times; init/internal/standard-library calls were not exhaustively instrumented. The 26,102B native result and 8,765B outside records were verified without rerunning original code. Held v1 plans, binaries and prior seals remain preserved.

New training-qualified requests are still zero. Finite satisfaction is followed by separate training/distribution rights, connected groups, roles, weights and data bindings. Actual corpus Fits remain 3/8 and logical models 3/16; utility-failed models 71/72/79 remain inactive. This evidence does not establish Laya inference, Codex savings or improved model utility.

Public artifacts contain owned comparison facts and measurement summaries. Original package bodies, host paths, PIDs, executable/model payloads and raw logs are withheld. [Measurements](OBSERVATION.v1.json) · [Finite comparison](FINITE-COMPARISON.v3.json) · [Actual CI/merge](PR108-ACTUAL-CI-MERGE.v1.json).

The [controller review](CONTROLLER-REVIEW.v1.json) is Root's transcription of a separate saved-receipt review, not a new native run. It preserves a pre-start accounting correction: the initial extra-control estimate was too small, so all four additional files were charged before starting. The comparison file grants no publication, training or activation authority; publication here is separately limited to owned facts. Neither review approves raw-source distribution or training.
