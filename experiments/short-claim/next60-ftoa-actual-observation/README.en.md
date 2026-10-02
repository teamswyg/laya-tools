# Direct rounding of the original float: first real execution

Training a small claim supplier requires evidence that a candidate satisfies the request. This experiment compares a selected public Go formatting function with two authored candidates on identical inputs. It validates data; model training and inference performance are separate experiments.

Rounding `1.375` to two fractional places should produce `1.38`. The original first formats six places and then truncates, producing `1.37`. The precise candidate rounds the original represented float directly. The negative control rounds to six places, parses that intermediate, and rounds again, losing a sufficiently small value.

| Input and precision | Expected | Original | Precise | Negative control |
|---|---|---|---|---|
| 1.125, 2 places | 1.12 | 1.12 | 1.12 | 1.12 |
| 1.375, 2 places | 1.38 | 1.37 | 1.38 | 1.38 |
| 1.999, 2 places | 2 | 1.99 | 2 | 2 |
| 0.000000125, 9 places | 0.000000125 | 0 | 0.000000125 | 0 |
| Negative zero, 2 places | 0 | -0 | 0 | 0 |
| Positive infinity | nonfinite error | No error return channel | nonfinite error | nonfinite error |

Six inputs × three candidates are 18 observations and **one distinct request**. Six independent integer-oracle checks all agreed with the frozen expectations. An unavailable original error channel is never converted into success or an absent error. The observation file retains null role, truth, weight and labels; later qualification is recorded separately.

On Go 1.27.1 / macOS arm64, the worker started once and was waited once. Retries, panics, model calls and protected evaluation reads were zero. Whole-child wall time was **0.378019625 seconds** and Darwin reported direct-child maximum RSS of **5,816,320 B, about 5.55 MiB** after exit. The Go heap snapshot was 207,056 B. The 0.000067333-second observation loop excludes startup, provenance checks and output persistence. These are measurements of a numeric validation worker, not Laya/claim-model inference, GPU usage or whole-machine memory.

Limits remained one Go execution thread, 256 MiB soft heap / observed RSS budget, 300 seconds whole-child time, 12 MiB result and 1 MiB outside records. RSS is an after-exit gate, not a hard memory cap. Storage accounting preserves and charges prior failures and binaries. The new 32 MiB archive allowance is 12 MiB result + 1 MiB outside + 19 MiB control records; it is separate from inference memory. Earlier storage failures are not retroactively passed.

Before execution, Root reviewed actual passing CI, selected-file scope, notices, binary settings, storage reservations and frozen inputs. Only the selected formatting file is imported, not the whole upstream package. The pinned Go compiler/generated headers and Darwin system libraries are stated trust boundaries. No independent complete compiler, OS or ancestral legal closure, or third-party binary redistribution clearance, is claimed.

Public files contain structured observations from frozen public inputs and path-free verification information. They exclude upstream bodies, binaries, private paths, credentials and profiling traces. Selected original attribution and hashes are in the [preparation record](../next60-ftoa-singlefile-preparation/README.en.md).
