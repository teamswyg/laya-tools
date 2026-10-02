# Qualifying finite supervision and training membership

Root compared every actual candidate channel with frozen Want72, then separately reviewed the original requests, complete candidate functions and captions. Matching all72 expectations verifies predicted behavior; task truth identifies candidates satisfying the fixed request and reference behavior.

| Request | Acceptable candidate index | Actual counterexamples in other candidates |
|---|---|---|
| Preserve partial UUID results and errors | 1 | index0:2, index2:3 |
| Preserve existing state and scanning errors | 0 | index1:6, index2:2 |
| Preserve English ordinal teen exceptions | 2 | index0:4, index1:3 |

All nine captions faithfully describe their full candidate functions and postprocessing. Faithfully described incorrect behavior can therefore be a known negative. The three requests yield three positive and six negative labels, nine eligible loss masks and unit sample weights. [Task/caption/channel/eligibility review](ROOT-ELIGIBILITY.v1.json) records the axes separately. This is AI-assisted nonblind Root review, not independent human-blind judgment.

The [train append manifest](QUALIFIED-TRAIN-APPEND.v1.json) binds original76 projection and actual70 role hashes, adding only parent indices76,77,78. Existing group IDs end at74; new caller groups75 for UUID and76 for Ordinal are distinct from parent indices and role reassignment output. UUID Parse/Scan, shared helpers/errors and every variant stay connected. Both new whole groups join development_train. Original76 truth, candidate order, masks, weights, roles, validation and calibration remain unchanged. Whole-corpus roleplan.Assign was not rerun.

Public Go input validation and79-parent projection have not executed. The manifest is the fixed input for that next step; model qualification/training_ready/production_ready remain false. Prior source53/55 exposure is recorded. Absence of these families in the declared old graph is not source-independence or heldout proof. Labels cover eight fixed inputs per request and pinned revisions, and do not enter the final2,400-request evaluation.

Read the [saved-result review](ROOT-SAVED-REVIEW.v1.json), [execution preflight](ROOT-PREFLIGHT.v1.json), [candidate preparation](../preparation/PILOT.v1.en.md) and [next fixed FP32 plan](../../data-effect-fit-79/plan/PLAN.v1.en.md). Historical Want/Got supervision-null snapshots are preserved; new truth appears only in this separate manifest.
