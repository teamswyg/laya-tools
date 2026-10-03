# Finite comparison of saved Catalog9 results

For 46 fixed inputs across nine tasks, we compared the original, an authored reference candidate, and an authored negative candidate. The input was the 138 rows already saved by Root. This review did not execute the original task code again.

The row outcomes are 102 satisfied, 35 unsatisfied, and one unknown. The 1,347 individual Wanted predicates contain 1,246 known matches, 72 known mismatches, and 29 unknowns. All nine reference candidates satisfy every fixed input and every Wanted predicate for their respective tasks. Each of the nine original candidates and nine negative candidates has a known counterexample.

The table uses zero-based fixture positions. All reference candidates are satisfied, so it lists the actual counterexamples for the two other candidates. [The detailed report](FINITE-COMPARISON.v1.json) preserves every candidate's complete row vector and unknown positions.

| Task | Inputs | Original counterexample positions | Negative counterexample positions |
| --- | ---: | --- | --- |
| annotation-ownership | 5 | 0, 1, 2 | 1 |
| bounded-count | 6 | 1, 3, 4 | 3, 4 |
| single-assignment | 5 | 0, 1 | 1 |
| sensitive-default | 5 | 0, 1, 3 | 3 |
| deferred-function | 5 | 1, 2, 3 | 1, 2 |
| text-error-value | 5 | 0, 1 | 0, 1 |
| hooks-errors-is | 5 | 0, 3, 4 | 2 |
| remain-conflict | 5 | 0, 1, 4 | 1 |
| tag-precedence | 5 | 1 | 1 |

The one unknown-only row is the original candidate at fixture position 3 of annotation-ownership. That revision has no available getter observation channel: nine predicates are unknown and the remaining eight are known matches. We did not convert this into a failure. Other rows may contain both a known mismatch and unknown predicates. Such a row is unsatisfied while its unknown predicates remain recorded.

Comparison is limited to the predicates declared in Wanted. Additional Got metadata is ignored, while empty objects and the full key sets of `/maps/A` and `/maps/B` are compared. Array order, field presence, null, empty arrays, and empty objects remain distinct. Declared unavailable error/getter/flag-identity channels are unknown. An actual known value or type mismatch is false. Matching `message_observed=false` and `message=null` establishes observation-metadata equality; it does not prove that no nested Error method ran.

The reviewer did not author the candidates, Wanted, or observer. This is a nonblind review because the reviewer had already read the source and provided correction feedback. The reviewer authored a standard-library-only Go comparator, ran synthetic controls, and sealed it. Root then independently read that source, found no concrete blocker, and authorized the one actual comparison. Synthetic controls are not observations of actual candidates or models.

The comparator ran once, exited with code zero, and produced zero stderr bytes. The saved input result has SHA-256 `5b0634522473882a7823c33565d3db68a6f11efd28e2b3474a59f8d6d57bfb99`; the comparison report is 86,308 bytes. This comparison performed zero original-code imports, initialization, or API calls, zero original-worker replays, zero model execution or training, zero protected-data reads, and zero network operations.

Root separately reported whole-worker RSS of 8,388,608 bytes and elapsed time of 1.509219167 seconds for its original worker execution. Those are whole-worker measurements, not measurements by this comparator, model or GPU memory, or model inference speed. Original startup and nested-call counts remain unknown/null. External process and journal evidence belongs to Root's separate record; the comparator does not establish it.

This record is evidence of candidate satisfaction on these 46 fixed inputs. It makes no general parser-correctness, model-quality, cost-saving, license-guarantee, or new public-CI-success claim. `labels_assigned=false` and `qualified=false` remain unchanged. Root separately decides labels, roles, training adoption, and final finite qualification. The tenth unknown-token-report task's six inputs are excluded.
