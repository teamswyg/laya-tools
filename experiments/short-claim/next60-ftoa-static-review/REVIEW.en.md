# Static review of sealed ftoa preparation

Execution readiness remains **false**. The 22-file, 63,996-byte inventory and selected source/Want pins were checked. No code, Want or old seal changed; Go, compilation, functions and models were not run. I authored the integer oracle, so this is not independent oracle QA. It is a nonblind source review separate from the worker, candidates and Want author.

Four boundaries remain before a launch.

- main89–94 syncs the once marker and empty result, but intermediate calls live only in memory. After kill/timeout, original call counts are **unknown**, not zero.
- main39/86 does not reject compiler_selection_pending. Seven inventory entries are not complete compiler-selection proof; Root's external frozen evidence must supply that gate.
- The new 18-row ftoa schema is incompatible with the existing 23-row native2 controller. Preparation correctly declares this BLOCKED.
- Nonfinite Want null text, actual Go empty-string return and baseline unavailable error channel need explicit comparison semantics. OracleChecks is not candidate satisfaction.

The 986-byte original renders six places and clips, returning a string only. Authored precise rounds at requested precision; flawed rounds to six places, parses, then rounds again. Signature and negative-control changes are declared separately; the six original draft Wants stay intact.

The exact fractions 9/8 and 11/8 exercise opposite tie parity. Carry at 1.999, nine-place small-number rendering, negative zero and Inf policy agree with static reading. These are not observed outputs. Actual compiler-produced decimal bits, compilation and general correctness were not checked. The source-static prediction table in [RECEIPT.v1.json](RECEIPT.v1.json) is not truth or labels.

Whole-package/compiler closure, CI, rights and resource evidence remain pending. A selected file and MIT notice do not clear the whole package. This is one existing development request, six inputs and 18 planned candidate observations, not new parents, families or training promotion. The [ledger](LEDGER.v1.json) records actual reading and limitations.
