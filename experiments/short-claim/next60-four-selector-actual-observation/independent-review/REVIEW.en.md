# Review of the saved four-task results

We compared 57 candidate observations over 19 frozen inputs with their unchanged Wants. Candidate 1 satisfies every frozen input for each of the four tasks. Candidates 0 and 2 each have observed counterexamples. Overall there are **40 satisfied observations, 17 known mismatches and no unknown comparisons**.

| Task | Candidate 0: satisfied/mismatch | Candidate 1: satisfied/mismatch | Candidate 2: satisfied/mismatch |
| --- | ---: | ---: | ---: |
| Strict integer comma grammar | 3/3 | 6/0 | 3/3 |
| Reject matching Name-key collisions | 2/2 | 4/0 | 2/2 |
| Independent map snapshot | 2/2 | 4/0 | 2/2 |
| Exact fractional byte sizes | 3/2 | 5/0 | 4/1 |

This is finite behavior evidence for four requests. The 19 inputs and 57 observations are not additional independent requests. Candidate names supply no satisfaction authority: the comparison checks error presence and every required amount, key and map state. [Detailed comparison](FINITE-COMPARISON.v2.json) records three candidate comparisons for each of the 19 inputs.

For comma grammar, both original-backed designs accept `12,34` and `1,,234`. On positive int64 overflow they return an error but also return `9223372036854775807`, where the frozen contract requires zero. Checking error presence alone would miss that mismatch. The two designs call the same original function and are not two independent implementations.

For key collisions, both `NAME/name` and `NAME/Name` must be rejected. Candidate 1 preserves `old`, returns the sorted conflicting keys and reports ambiguity even when values are equal or an exact-case key exists. We do not claim that the original map fallback will always select the same key. The missing required ambiguity error establishes a counterexample regardless of the selected name.

For snapshots, both mutation directions were checked. Candidate 1 prevents snapshot mutations from changing the live map and live mutations from changing the snapshot. The original and outer-struct-copy designs share storage and fail this isolation. Nil and nonnil empty maps remain distinct: the missing flag returns a nonnil empty snapshot plus `*pflag.NotExistError`, while the separately registered `items` live/legacy map stays nil. These results are specific to the pinned pflag revision whose getter returns shared storage.

For byte sizes, `MaxUint64.0` and `MaxUint64-0.1` are valid, but the original floating path rejects them. The truncate-first design instead admits `MaxUint64+0.9`. The exact candidate checks range before truncation. Large amounts are compared as exact decimal strings without floating conversion. Five decimal probes do not establish all grammar, Unicode, units or whole-number execution equivalence.

An optional null Want supplies no predicate; it is different from an unavailable observed channel. The comparison checks `error_present=false` as well as true, required error kind/type, return amounts, ordered keys and complete maps including nil and entries shape. No frozen Want imposes error-text equality. None of these 57 comparisons has an unavailable required channel; uninstrumented original internal/init call counts still remain null.

The reviewer did not author the Wants, candidates, observer, native binding or outside controller. The reviewer inspected source and Wants before execution and knew candidate design names, so this is not blinded evaluation. The first saved-JSON filter failed because of field-list binding precedence. Its source, empty output and diagnostic were retained; a fresh filter corrected the metadata comparison. No original execution was replayed and no Want changed.

Root's record reports one Start and one joined Wait, 57 returned candidate wrappers, 14 returned direct error methods and no panics. Maximum worker RSS of 8,273,920 bytes and approximately 1.078 seconds of whole-child time measure the numeric observation/recording worker. **They do not measure model inference, GPU performance or Codex savings.**

This review supplies finite candidate satisfaction only. Root separately decides labels, roles, weights, whole groups and publication. It establishes no new training, utility improvement, completion of the 30/60-request training checkpoints or collection of 2,400 protected semantic requests per claimed domain. See the [review record](SEMANTIC-REVIEW.v1.json).
