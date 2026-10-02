# Saved original72 projection69 runtime result review

Read-only comparisons passed for the original result's pins, array bindings, excluded scopes and resource records. Root performed the original execution once. This review is not an execution replay, new ground truth, or approval of training, performance or diversity readiness.

The exact original result is 7,699,810 bytes with SHA256 `348af320a8bdcbee433f3102cc72aba06404ddc568a9ae9ddf6ff8f8d2980a4f`. It binds to root's frozen plan SHA `1811c6c098468acd48e774075263103af4bc87a6f42e398cc7920abe2edb8d84`, reservation, pre-start markers and completion ledgers. Draft→frozen changes only `frozen:false→true`. The 30 execution input/source/binary pins total 8,158,159 bytes; all 42 references total 15,882,779 bytes. Their actual file SHA/byte counts were checked. Public records omit private paths and raw time logs.

## Preserved data

| Item | development | validation | calibration |
|---|---:|---:|---:|
| Parents | 44 | 16 | 12 |
| Candidate positions | 132 | 48 | 36 |
| Unknown candidates | 42 | 12 | 9 |
| Fit-array rows | 90 | 36 | 0 |
| Retained zero-weight rows | 18 | 1 | 0 |
| Whole groups | 11 | 3 | 3 |
| Groups containing positive-weight rows | 9 | 3 | 0 |

All 72 parents/216 candidates preserve the original 34 known, 17 no_answer and 21 unknown parents. Candidate counts are 2/3/4 for 12/48/12 parents, retaining original order. The 28 nonpositional rows in the original role JSON are joined by original index without reordering that original JSON. Raw request/caption/IDs, the full ordered acceptable sets, groups/roles, loss masks and nullable labels were compared with pinned inputs. All 63 unknown candidate labels preserve null meaning. Calibration's 36 candidates remain in RuntimeParent while contributing no fit rows or feature extraction. Retaining known rows with zero weight creates no new negative truth.

## Stored arrays and diagnostics

Checks cover sparse offsets, index/value lengths, row-wise index uniqueness, the 8192 dimension bound, finite values, canonical parent/candidate order and label/group/weight bindings. Development/validation have 64,411/33,278 feature entries. All 19 zero-weight rows retain their original features and labels. Whole-group role leakage was checked.

Diagnostic arrays contain only 72/35 positive-weight rows, copied exactly from their fit-array rows, including bit-identical indices and float64 values. They have 35,223/32,960 feature entries. Diagnostic `Data.Excluded=0/0` and `UnknownAuditCandidates=42/12` retain their separate meanings. Diagnostic positive/negative labels are 20/52 and 10/25. No AUC was calculated.

The 72 stored normalized requests and 216 captions were compared using a stdlib copy of the already pinned lexicalhint string-splitting recipe. These are 288 stored-string checks; original NormalizeText/Validate/Features/Project calls remain zero. The mathematical generation of feature values was not recalculated. The review does not independently prove those feature semantics.

## Resource meanings

| Record | Value | Scope |
|---|---:|---|
| Owned arrays/headers/strings PayloadBytes | 1,774,364 bytes ≈ 1.692 MiB | Recomputed from saved lengths and matching Go ABI struct widths; allocator/scratch excluded |
| Original serialized JSON | 7,699,810 bytes ≈ 7.343 MiB | Result file size |
| Whole-child peak RSS | 56,573,952 bytes ≈ 53.953 MiB | macOS time observed bytes, not a hard cap |
| Whole-child peak footprint | 49,857,016 bytes ≈ 47.547 MiB | macOS observed bytes |
| Go heap soft setting | 268,435,456 bytes = 256 MiB | Configuration, not measured heap |
| Whole-child displayed real/user/sys | 0.46/0.07/0.01 seconds | Entire process including reading, validation, Project and serialization |
| Controller wall | 0.465949875 seconds | Controller-observed whole waiting interval |

Owned payload and serialized output satisfy their separate existing 64 MiB bounds. The recorded original direct Validate count of 72, Project count of 1 and returned FeatureScans of 252 match two passes over 126 rows. Project's internal ValidatePrepared/normalization calls were not individually instrumented and were not newly estimated. Isolated Project latency, actual Go heap, GPU use and AI-assisted preparation cost are unmeasured. These figures are not training or model inference benchmarks.

## Failures, author exposure and limits

The independent helper's first attempt refused with `runtime69_parent_binding` because the auditor incorrectly assumed three candidates for every parent. This was not an original data defect. The initial source/tests/module, invocation and failure record are preserved. The corrected second metadata helper compares each parent's original candidate count and passed. Both synthetic race runs passed; the latest suite has 6 top-level tests/6 subtests. Both vet runs passed. Additional failed reads/tools/tests: zero. Helper compile/tool preparation time is separate from original execution time.

The reviewer did not author driver69 or claimfit.Project, but prior pairlearn trace, roleplan62 and scope67 authorship and earlier nonblind artifact exposure remain disclosed. This is not independent source authorship or blind semantic evaluation. Saved-state comparisons cannot prove concurrent-mutation immutability, compiler hermeticity, wording fidelity, source diversity or model usefulness. Original truth/masks/roles are unchanged. New numeric gates, roles, labels, seeds, models, fit, paid APIs, native execution, protected-final reads, shared edits, Git and publication all remain zero. `training_ready=false` and `source_diversity_cleared=false` remain preserved.

Related: [runtime review receipt](RECEIPT-RUNTIME-PROJECTION-69.v1.json), [array/pin/resource mechanics](MECHANICS-RUNTIME-PROJECTION-69.v1.json), [audit ledger](LEDGER-RUNTIME-PROJECTION-69.v1.json), [first helper refusal](HELPER-ATTEMPT-1-REFUSAL.json).
