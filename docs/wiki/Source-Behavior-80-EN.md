# Public function verification and training data expansion

All 24 UUID Parse, Scan and Ordinal inputs matched expectations fixed before execution. One run, without retries, recorded 24 entrypoints and 5 explicit error-message calls: 29 total. See [complete results and evidence](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/native-observation-80/ACTUAL-RESULTS.en.md).

This prepares evidence for a small claim model. “Every returned byte is zero on error” conflicts with actual partial UUID results. Such differences must be verified before creating requests, candidate claims and training truth. New truth, roles and parent requests remain zero; these are not 24 independent tasks.

Whole-program RSS17.265625 MiB/wall1.21s include initialization, hashing, JSON and storage. They are not per-function or model-inference costs. Laya/GPU inference and Go heap were not measured; no RAM saving comparison was performed.

[The next source record](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/source-observation-81/README.en.md) proposes four goals and 32 input drafts. Source, expectation and real behavior verification precede training-data design. Both existing learned models remain inactive because they increased inspection work. Final evaluation on 2,400 independent requests remains incomplete.

Full upstream BSD/MIT notices are retained; local paths, binaries and raw logs are omitted. [Current usage](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.en.md) · [한국어](Source-Behavior-80-KO)
