# First actual Go fit-input preparation69

The original72 requests produced fitting and validation arrays in **one first Project call, zero retries and zero fits**. This is data preparation, not Laya inference or a new model-quality result. It follows the [input-boundary check](../input-boundary-69/README.en.md), [mask-role join](../mask-role-69/README.en.md) and [stored nonlearned utility check](../stored-role-utility-69/STORED-ROLE-UTILITY.en.md).

| Saved output | Train | Validation | Calibration |
|---|---:|---:|---:|
| Parent requests |44|16|12|
| Candidate positions |132|48|36|
| Fit-input rows |90|36|0|
| Zero-weight rows above |18|1|0|
| Unknown candidates retained only for audit |42|12|9|

All known train/validation candidates remain physically present. Candidates whose captions are unsuitable for loss receive zero weight without changing their truth or removing them from ranking comparisons. Unknowns and calibration stay outside fitting. Three calibration groups with eligible supervision and zero positive-weight groups **in projected fitting rows** are different denominators.

Direct input validation returned72 times and Project returned once. Two feature passes account for252 scans; internal normalization calls were not separately instrumented. Preserve the [outside actual ledger](ROOT-ACTUAL-LEDGER-69.v1.json) and [pre-start reservation](ROOT-INVOCATION-69.v1.json).

| Resource observation | Value | Scope |
|---|---:|---|
| Owned returned payload |1,774,364B, about1.69MiB|Existing Go ABI calculation for arrays and retained strings; excludes scratch allocation|
| Whole-child peak RSS |56,573,952B, about53.95MiB|OS observation including reading, validation, preparation and JSON output|
| Outside wall |0.465949875s|One development observation of that full preparation process|
| Original JSON |7,699,810B|Includes indentation and repeated diagnostic arrays; not a memory measurement|
| Public summary |97,255B|Derived raw reference and per-row hashes; omits full features|

CPU1, soft Go heap256MiB and an outside300s wall bound were configured. RSS was observed, not forcibly capped. Go heap, GPU memory, serving latency and LLM savings were not measured. AI-assisted preparation and review costs are also unmeasured.

The [public summary](COMPACT-ACTUAL-RESULT-69.v1.json) preserves original SHA `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`, supervision, row references, feature counts and hashes. It cannot reproduce training by itself and does not replace the full arrays. The original arrays and private execution plan remain outside Git; at this copy milestone HF publication of those arrays is still pending. No fitted coefficients are present. `SampleWeights` are supervision loss weights.

The [author array audit](FINDINGS-ACTUAL-69.v1.en.md) compares saved results using a separate standard-library Go checker. That reviewer authored drivers69/71, so this is not an independent review. The [separate pre-execution review](FINDINGS-PRECHECK-69.v3.en.md) preserves the diagnostic `Excluded` denominator issue, the original role DTO ordering issue, earlier refusals and fixes. The original DTO order remains unchanged; an original-index array lookup joins its rows. A precheck is separate from review of actual arrays.

The [separate actual-array review](FINDINGS-RUNTIME-PROJECTION-69.v1.en.md) also passed. Without rerunning originals, it checked30 execution pins, supervision,288 stored normalized strings, sparse columns and diagnostic subsets. Its first checker falsely assumed three candidates for every parent; that refusal is preserved. Only the checker changed to compare each original2/3/4-candidate input. The [runtime-review copy ledger](COPY-RUNTIME-QA-LEDGER-93.v1.json) preserves eight source assets and discloses the reviewer's earlier authorship in62/67. This is not feature re-extraction or training/source-diversity approval.

The [clerical ledger correctionv2](LEDGER-RUNTIME-PROJECTION-69.v2.json) is also preserved unchanged. The first refused checker's completed string prefix was4, reconstructed from source control flow, rather than the192 incorrectly inferred in v1. Its successful288 checks, original execution and primary success receipt remain unchanged. v1 was not overwritten; reconstructed arithmetic is not an instrumented counter.

The [copy ledger](COPY-LEDGER-93.v1.json) records13 byte-identical copies and readback. This does not clear training readiness, source diversity, utility or generalization. The next76-parent corpus separately adds four upstream68 requests; it does not rewrite original72 results. [한국어](README.ko.md).
