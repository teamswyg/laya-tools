# Audit joining existing roles and supervision masks

**One read-only metadata audit passed.** It joined the stored role result from 66 and mask result from 67 to the original 56/56b inputs and truth, and the complete stored whole-group membership from 65. It did not execute candidate functions or assign roles again. This is not a new semantic approval or a finding that training is ready.

The pre-execution plan is `execution-plan.json`, SHA256 `f5f3ad35942120f670813e272dd82ef2a9d78519a8585cd4a59bf1cb929bc940`. The result is `official-attempt-1/results.json`, **233,601 bytes**, SHA256 `7127cb9cc24ae8a1d7ada23f51599f27d2e90dcf39dbd60d510e0001aebb3569`. The result was created exclusively in a fresh directory. There were 0 retries.

## What was preserved

The audit checked all **72 requests and 216 candidate positions**: IDs, order, decoded UTF-8 request/caption byte hashes, source IDs, original truth, and acceptable candidate indices. It also checked all 17 whole groups and the roles of all 72 parents. Candidate counts followed each original request rather than assuming 3 candidates everywhere. Group 8 retains its connection between two prototypes, and group 64 remains unknown-only.

Stored whole-function truth and caption fidelity describe different things. A candidate's compliance with its original finite function contract remains valid even when its caption is incomplete. Consequently, a false loss mask did not change the original 1/0 label or remove the candidate. All 63 candidate labels belonging to the 21 unknown requests remain null.

Primary scope A preserves **all 51 original known/no_answer requests and all their candidates**. Masks affect only a future training loss. The 37 known requests whose complete candidate sets have eligible captions are separate scope B diagnostic counts; they do not replace A or the utility denominator from 58. This audit calculated no rankings or utility.

## Observed counts by role

| Item | Train | Validation | Calibration | Total |
|---|---:|---:|---:|---:|
| All requests | 44 | 16 | 12 | 72 |
| Answerable / no_answer / unknown | 20 / 10 / 14 | 8 / 4 / 4 | 6 / 3 / 3 | 34 / 17 / 21 |
| All candidate positions | 132 | 48 | 36 | 216 |
| Original known positions | 90 | 36 | 27 | 153 |
| Unknown null positions | 42 | 12 | 9 | 63 |
| Original positive / negative | 28 / 62 | 11 / 25 | 7 / 20 | 46 / 107 |
| Loss-eligible positive / negative | 20 / 52 | 10 / 25 | 5 / 18 | 35 / 95 |
| Masked known positive / negative | 8 / 10 | 1 / 0 | 2 / 2 | 11 / 12 |
| All whole groups | 11 | 3 | 3 | 17 |
| Stored known-containing groups | 10 | 3 | 3 | 16 |
| Groups with at least one eligible candidate | 9 | 3 | 3 | 15 |
| Groups with eligible positives: descriptive only | 7 | 3 | 2 | 12 |
| Complete-caption known parents: scope B diagnostic | 21 | 11 | 5 | 37 |

The existing stored-truth group floors of **9 / 3 / 3** are satisfied. Comparing any-eligible groups with those same floors also yields 9 / 3 / 3. No new floor was applied to the positive-bearing counts of 7 / 3 / 2. Group 52 has stored truth but no eligible candidates. Group 64 remains unknown-only provenance in the train role. No group was split, deleted, or reassigned.

## Check scope and execution ledger

The result's 12 explicit check counters sum to **792**. This is neither a count of every internal assertion nor a count of independent samples. The principal counters are 7 verified input files, 17 groups, 72 parent joins, 216 candidate joins, 72/216 request/caption references, 72 membership-parent joins, 72 role-parent joins, 17 group-role bindings, and 17 group-mask totals. The separate 68 proposal's 4 parents and 10 candidates were checked without combining them with the original counts.

The tool uses only the standard library, Go 1.27.1, CGO 0, and trimpath. A single race run passed all 6 synthetic unit tests. Vet, binary build, and plan preparation each passed once. Preparation and actual metadata-audit failures were 0. These checks did not execute original candidate APIs or the stored role-assignment function.

The original **1 Assign execution** from 66 was preserved. New calls to Assign, MembershipDigest, Project, Features, Normalize, Validate, Baselines, candidate/upstream APIs, the native worker, fit, or a separate model were all 0, as were new roles and labels. There are no new CPU/RSS/speed measurements or savings conclusions. This does not mean that ordinary Codex collaboration itself had zero cost.

## The next actual Go bridge

The separate 68 proposal contains **4 requests, 10 candidates, and 2 whole-source families, with 5 proposed positives and 5 proposed negatives**. Its independent QA metadata leaves actual training loss weights null and records 0 mutations to existing dataset labels. This audit checked the proposal's fixed order and scope but did not append it to the existing 72/216 or the roles from 66.

The parent's separate first input-bound audit also passed once. The SHA256 of `results-69-input-boundary.v1.json` is `53b492d0629f794785e592c90a96f3d86223dbf81891393d763a89e4a9d92d55`. A read-only check of its stored SHA and structure confirmed that all 72 parents and 216 candidates are supported with no errors. The existing 512-byte and 32 normalized-word contract was preserved; the maximum normalized request and caption word counts were each 32. The initial static concern did not become an observed failure. That separate audit made 72 Validate calls and 288 explicit metric Normalize calls; this agent did not repeat them. The frozen mask-role audit plan and input pins were not changed.

No blocker in the metadata join or input bounds was found for the first Project. The actual loader must obtain Prepared values from raw original text through `shortclaim.Validate`, then bind every original truth, acceptable set, mask, and role into `claimfit.Parent`. Directly filling exported Prepared arrays would bypass validation. Project remains at 0 calls in this report, and training readiness is still not established.

A combined execution would retain all members and relationships from 65, include the two whole families and four requests from 68 in their fixed order, and freeze a **new complete mapping**. The seed remains ASCII `1729`, with the existing 3:1:1 largest-remainder algorithm and no favorable seed search. New requests cannot simply be appended to the old role result. Original truth, masks, unknowns, negatives, and no_answer denominators remain intact. No new groups or scientific floors are invented.

An opaque membership hash establishes identity with recorded membership; it does not prove that external relationships are absent. Source and authoring-flow diversity remain uncleared. The separate target of at least 2,400 protected-final requests per domain remains outstanding and is not a minimum for every development fit.

## Input pins

| ID | Bytes | SHA256 |
|---|---:|---|
| probes56 | 49,941 | `0fe97dd65606f4239ef9187f1fc4d9c8dc2fcaebf342612b2071220896b92df0` |
| probes56b | 43,598 | `0d2bf6980353cefc4327af165c0f60a6455c1feea920f354dc621019ce3ed21e` |
| saved56b | 266,818 | `4128400c792d2151955a1d98f7d35aca6112d097a3ac20aa280a7357996636c4` |
| membership65 | 1,986,200 | `ae2794224ab8fdb03aa879b2522777beb5065005286b6394db3ba4a8abcdd1ce` |
| actual66 | 31,789 | `6bcf7a1e843991e78e00856232da5b12ab4d4208ebc5b1a55cf2337c1719ba06` |
| masks67 | 1,026,473 | `0ab835d8dca2e3157d9546e0059b39cd46c4c309b3c237489f74960df9cfcebf` |
| proposed68 | 17,101 | `ecc9de25fea14872257fc50af2ddb312db644d0e0b5e9ed07ffdc9f3fb119083` |

Original inputs, the shared repository, historical results, and the frozen bridge plan were not modified. This report and ledger are private preparation artifacts; no publication or commit was performed.
