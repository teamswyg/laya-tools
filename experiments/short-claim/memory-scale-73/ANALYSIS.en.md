# Sparse-array memory: current layout and scaling boundaries

**Feature indices already use uint16.** `pairlearn.Dataset` is CSR/SoA: separate row offsets, indices, FP64 values, labels, groups, and weights. Ranking's parent/positive-row/negative-row/weight columns are also SoA. One run owns its arrays without shared locks. Text/audit metadata remain in parent structs with eight fixed candidate slots.

The saved 76-request/226-candidate projection owns **1,928,154 B = 1.84 MiB**, reproduced exactly from saved column lengths and mirrored 64-bit struct sizes. Train91/validation45 rows contain 105,079 entries; diagnostic CSR copies add 75,573. Index+value bytes total **1,806,520 B, 93.7% of payload**. Rows do not store 8192 dense values: fit rows average 772.6 stored entries, including zero values after collision accumulation.

The table extrapolates **the same candidate, unknown/calibration, mask, and text-length composition as the existing 76 requests**. It is arithmetic, not a generated corpus or observed runtime. N counts the entire development pool.

| Requests N | Owned payload estimate | Same-format JSON file estimate |
|---:|---:|---:|
| 128 | 3.10 MiB | 13.48 MiB |
| 256 | 6.19 MiB | 26.95 MiB |
| 512 | 12.38 MiB | 53.91 MiB |
| 2400 | 58.04 MiB | 252.69 MiB |

This composition reaches owned 64 MiB at roughly 2646 requests, but **the private fitter's fixed 64 MiB input-file contract is reached first, at roughly 607**; no automatic increase is proposed. Unknown parents retain nullable audit text, calibration has no fit rows, and masked known rows leave only the diagnostic CSR. Different fractions change the estimates.

For all-known train/validation requests with all weights one, using current mean entries/row and maximum string byte allowances, 2400 requests with 2/4/8 candidates require **80.18/156.34/308.65 MiB**. The corresponding 64 MiB budgets allow means of about **596/268/105 entries/row**. This is a scenario, not a prediction. The 32-word contract yields ≤63 unigram+bigram terms and ≤3969 crosses plus recall: **≤3970 entries/row**. At that upper bound even 128 × 8 candidates exceed 64 MiB.

64-bit payload is `1096+576N+retainedTextBytes+10F+48R+32` B, with F/R totaling all four CSR datasets. Scratch 32 B/fit-row, temporary features, JSON parsing, allocator/runtime costs are separate. Ranking reserves `48(fitRows+2)+32pairs+48epochs`: about **2.05 MiB** at 2400 × 8 candidates and maximum 16 pairs/parent, under an independent 64 MiB cap. FP32 file 32,792 B, decoded coefficients 65,536 B, and core training arrays around 4 × 64 KiB also differ. **Go soft 256 MiB is not a hard RSS cap.** Historical projection JSON 8,390,462 B/RSS 47.7 MiB cannot establish future RSS by multiplication.

Lossless uint16-parent/uint8-candidate RowRef columns save only **3289 B** today; uint32 offsets save **1028 B**. **uint16 offsets fail: training already has 67,104 entries.** Arbitrary external IDs require inverse mappings preserving original cross-role checks. Diagnostic reference views could avoid **755,730 B** of duplicated features but require compatible APIs. FP32 values change numerical/schema semantics, potentially affecting gradients/ties/order; they remain a separate ablation, not an immediate lossless change.

The next minimal option is **compact/sharded storage preserving FP64 values**, exact value/index round trips, and text/order/null/truth/masks/roles/loss. It addresses file overhead, not guaranteed array capacity. Code/schema/limit changes remain zero; new development data remains the priority. **The separate fresh-domain final ≥2400 set can be evaluated one JSONL record at a time; it need not all enter training memory.** These estimates approve neither final completeness nor performance/readiness; protected-final reads zero.

Sources: [projection](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/claimfit/projection.go), [Dataset/trainer](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/learn.go), [ranking](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go). `CALCULATIONS.json` preserves source/input hashes and assumptions. This analysis used one saved-array calculator run, zero failures, and one reading-search path error; original application APIs, training/inference, Project/Features, benchmarks, shared edits, and external publication were all zero. AI-assisted collaboration cost was not measured.
