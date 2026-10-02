# Reading the same saved result through JSON and compact

In this single comparison, the compact reader showed lower cumulative Go allocation and shorter observed execution time, but RAM savings were not established. Maximum RSS was 19,120,128 B for JSON and 19,300,352 B for compact: compact was 180,224 B higher. Final Go heap was also 104,848 B higher. This is a saved-format reading experiment, without rerunning models, fitting, original projection or features.

One fresh process read the existing 76-parent projection JSON, 8,390,462 B, and another read its existing compact blob, 1,978,694 B. The JSON reader first token-scans numeric arrays for lengths and the retained-payload cap, then allocates arrays in a second pass. Compact uses the unchanged existing Decode. Both paths perform the same all-array validation and strict metadata checks. Controller pin reads warmed both inputs first; the fixed order was JSON then compact, one sample each.

| Saved observation | JSON reader | compact reader |
| --- | ---: | ---: |
| OS real / user / sys, seconds | 0.49 / 0.08 / 0.01 | 0.05 / 0.01 / 0.00 |
| Controller-observed child wall, seconds | 0.50100075 | 0.058869916 |
| OS maximum RSS, B | 19,120,128 | 19,300,352 |
| OS peak footprint, B | 16,269,840 | 15,745,528 |
| Final Go HeapAlloc, B | 4,660,664 | 4,765,512 |
| Cumulative Go TotalAlloc, B | 44,840,120 | 21,608,800 |
| Owned array+metadata payload, B | 1,978,406 | 1,978,406 |

RSS describes the whole process, HeapAlloc is its final Go heap observation, and TotalAlloc is cumulative allocation during execution. Lower cumulative allocation does not mean lower peak RAM. Compact's peak footprint was smaller while RSS and final heap were larger, so this does not approve a general RAM benefit. Displayed `0.00` sys is rounded, not proof of zero CPU use. Go 256MiB is a soft target; the retained-payload 64MiB cap does not include tokenizer/scratch/slice headers/allocator/whole RSS and is not a hard memory bound.

Actual root records show one start per child, zero retries, exit0, no timeout/overflow and returned Wait. Completion was established from `BothCompleted`, `ArrayAndMetadataDigestsEqual` and the final state, not exit code alone. All 24 columns—six columns across four datasets—match in count/null/ordered numeric-bit digest, with matching source JSON and metadata SHA. All 24 actual columns are nonnil. Rows are 91/45/73/44 and NNZ are 67,104/37,975/37,916/37,657. This is not FP64-to-FP32 conversion, uint16 index narrowing or diagnostic deduplication.

A separate reviewer checked 16 saved-file pins and 184 conditions. That runtime QA checks reported count/null/digest fields and their record bindings; it is not a second source-array read or independent bit-digest recomputation. Earlier static source inspection and the root's actual array reads/digest computation are the evidence. The reviewer did not author the reader/controller but has prior codec/preflight exposure and is nonblind. This publication-copy author authored the original reader/controller; this staging is not an additional independent performance experiment.

The single warm-file whole-child time includes source/binary pin checks, reading, array validation, hashing and reporting. JSON performs two passes for safe cap enforcement. Shorter observed time is not promoted to causal format-only acceleration, cold I/O,128-to-512 scaling, serving performance, LLM token savings or model utility. The earlier storage-roundtrip 76.4% byte reduction combines removal of pretty formatting with binary storage; it is separate from this read measurement. Protected final2400 completion is not claimed.

Original input snapshot/compact blob/binaries/raw stdout/time/private plans/private go.mod/helpers are excluded from this archive. Plans are separately derived as a [public draft view](PUBLIC-DRAFT-PLAN.v1.json) and [public frozen view](PUBLIC-FROZEN-PLAN.v1.json), recording original SHA and four removed host-path fields. These views cannot be used as execution plans. Source is a `.go.txt` reference archive retaining original SHA, not a standalone executable module. Exact copies and derived changes are recorded in the [SAFE-COPY-LEDGER](SAFE-COPY-LEDGER.v1.json).

Read the [root controller result](actual/controller.results.json), [JSON result](actual/json-reader.results.json), [compact result](actual/compact-reader.results.json) and [independent runtime QA](independent-runtime/RECEIPT.v1.json) together. New research/reader/codec/source API, Features, Project, Fit, roles, labels, models, paid calls, HF and shared repo publication executions remain zero. TrainingReady/PerformanceApproved/Final2400Complete stay false. Zero separate model calls does not imply zero cost for this AI-assisted development/review collaboration.
