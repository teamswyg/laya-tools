# Saved-record comparison for 77 v2

All 24 saved observations match the expectations sealed before execution. This check compared stored JSON, file hashes, counters, and recorded OS values without running original functions or the observer again. A Go 1.27.1 standard-library checker ran once and passed, with 0 failures and 0 retries.

| Check | Stored evidence and comparison |
|---|---|
| Finite observations | 12 logfmt, 4 Escape, and 8 Unescape cases; 24/24 Want matches |
| Tracked direct callbacks | 217 reserved, 217 returned, 0 panics; 209 original-package public calls + 8 standard-library error-string calls |
| After EOF | 7 EOF cases without an error skip further repeated calls and retain null observations |
| Sticky errors | Repeated observations in 5 logfmt error cases preserve error types, text, and positions |
| Byte observations | Nil/empty distinctions, lengths, hex, 15 successful key/value pairs, ordering, duplicates, and partial records agree with the sealed values |
| File bindings | All fields of 24 fixtures and Want/Got objects, 26 source pins, the zero-call reservation, and the original v1 Want hash were checked |
| Execution record | 1 root Start, joined Wait, exit 0; 0 retries, timeouts, and output overflows |

Every JSON field of each fixture was compared. JSON numeric text was preserved instead of converting numbers to floating point. The 78 byte states and 24 error states count observation occurrences, including the same error at several observation locations. They do not represent 78 distinct samples or 24 error cases.

The final `results.json`, `partial.json`, and root-captured stdout are exactly the same 43,604 bytes, with SHA256 `e5305785e8aca293b8dcb5d9f684ca824e0aeb05b10bbe568e387357b941a99a`. Only the draft plan's `frozen: false` token changed to `true`. The root reused the generic execution-ledger schema from 76. That fact is retained, and the ledger is bound to the actual 77 arguments, minimal environment, binary, plan, and Want hashes. A shared schema name was not treated as proof that different experiments are identical.

The recorded Darwin whole-child maximum RSS is **29,835,264 bytes = 28.453125 MiB**. Peak footprint is 22,807,104 bytes; OS real/user/system times are 5.20/0.17/0.46 seconds. Controller wall time is 5.210141667 seconds; the worker's internal pre-final-write snapshot is 4.817438167 seconds. Its heap snapshot is 4,083,152 bytes, and heap system bytes are 24,707,072. Cumulative allocation of 51,355,616 bytes is not simultaneous resident memory. These records include process startup, I/O, and persistence before and after each call. No subtraction estimates pure API or startup time. A soft Go heap setting is not a hard OS RSS limit.

This was a **native Go CPU observation**, not Laya or GPU inference. The 217 tracked calls cover the designated public APIs and error-string calls, not every internal helper or standard-library call. Original-package initialization occurs before main and was not dynamically instrumented.

The checker authored the earlier source-first/Want review and had already seen the result summary. The checker did not author the observer or root controller, but this saved-record comparison is not a blind semantic evaluation or an independent runtime reproduction. The original v1 EOF-usage finding and hashes remain unchanged; no claim is made that its panic path was reproduced by execution.

These 24 finite observations of two behavior goals are not 24 independent parent requests, training labels, or roles. New model calls, fitting, Features, Project, role assignment, original API/observer/test reruns, shared-repository changes, and remote mutations are all 0. `qualification`, `training_ready`, `production_ready`, and `protected_final` remain false. This observation does not replace the protected target of 2,400 fresh final examples per domain or establish generalization.

Numbers and bindings are recorded in [NUMERIC-RESULT.v1.json](NUMERIC-RESULT.v1.json) and [RECEIPT.v1.json](RECEIPT.v1.json); the execution history is in [ATTEMPT-LEDGER.v1.json](ATTEMPT-LEDGER.v1.json). The checker's source containing host paths, its private binary, raw OS logs, and private plans are excluded from the public-safe bundle.
