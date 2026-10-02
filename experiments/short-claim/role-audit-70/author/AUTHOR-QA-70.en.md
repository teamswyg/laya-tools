# Author metadata QA of the actual 70 role result

One stdlib Go metadata check successfully joined the saved single-attempt result to its frozen inputs. The reviewer authored the 70 metadata and worker, so this is not independent role-algorithm or semantic validation. It did not rerun roles, MembershipDigest, OrderDigest, Assign, Features, Validate, Project, fitting, models or upstream APIs.

The actual result is 29,864 bytes with SHA `24de7b745747966107b00c6748a61dc920d97ed9cc9f98a8f22fd367cd8b6b70`. Its frozen execution-plan SHA `6bc94c340d3983f950349fb327453e5518b9c2980c8c09175ccb05886f2d94db` and proposal SHA `4e9d36db9e151f807cc2c79da72d863308369d0708a86524b32f717a334f8c21` match. All 35 byte pins match: plan, executable, 14 inputs, 15 sources, result, worker receipt/ledger and root ledger. The original 17 complete stored group objects remain unchanged; two whole upstream families bring the total to 19. All 76 parents and 226 candidate positions match the result's parent_roles by ID, order, group, truth state and proposed-only flag.

| Observed saved metadata | Train | Validation | Calibration |
|---|---:|---:|---:|
| Whole groups | 12 | 4 | 3 |
| Known-containing groups | 11 | 4 | 3 |
| Requests / candidate positions | 44 / 130 | 20 / 60 | 12 / 36 |
| Answerable / no-answer / unknown requests | 22 / 9 / 13 | 10 / 5 / 5 | 6 / 3 / 3 |
| Unknown null positions | 39 | 15 | 9 |
| Mask-qualified P / N | 22 / 51 | 13 / 31 | 5 / 18 |
| Masked known positions | 18 | 1 | 4 |
| Groups with any mask-qualified position | 10 | 4 | 3 |
| Positive-bearing groups, descriptive | 8 | 4 | 2 |

All 21 unknown requests and 63 null positions remain. Unknown-only group 64 stays train provenance, without promotion into labels or weights. All actual training weights are still null. The original 46 P/107 N labels, source68's proposed 5 P/5 N, acceptable positions and masks remain unchanged. Six stored-truth coverage requirements pass the original 9/3/3 floors. Mask-qualified groups 10/4/3 and positive-bearing groups 8/4/2 are descriptive, not new floors.

The saved fixed-seed assignment places both new source68 families and all four new requests in train. This role result does not prove independent upstream transfer in validation or clear authorship diversity. `training_ready=false` and `source_authoring_diversity_cleared=false` remain. No graph recomputation, favorable-seed search or new semantic labels occurred.

Saved worker counters record one completed compiled-identity check, one membership-set verification and one Assign. Module counters record 19 components validated, 18 known order hashes, six coverage checks passed and 19 assignments returned. The root ledger records one actual child, exit 0, no retry and no timeout. This QA checks stored numerical joins; it makes no new resource measurements.

The frozen plan's limits[0] retains the historical draft sentence stating `execution_authorized=false` and no execution yet. The original is unchanged. Actual top-level fields are true/frozen, and the receipt/result/final ledger record one completed execution. The invocation receipt's actual_assign_attempts 0 is its pre-dispatch reservation snapshot, not a final counter contradicting the later completed result.

The helper is `check.go`, its pre-invocation plan is `QA-PLAN-70.v1.json`, and the output is `AUTHOR-QA-70.v1.json` (8,925 bytes; SHA `a0e401bc3a3baf1ba1b34315ad0faf0361df2734b7ef4f32b837e0adb66527dc`). One invocation succeeded with no failures. Shared edits, new roles, fitting, upstream APIs, models, protected-final reads, commits and publications remain zero. One broad private-directory listing encountered an unrelated permission error, retained separately in the ledger. AI assisted preparation/reading; collaboration cost was not measured.
