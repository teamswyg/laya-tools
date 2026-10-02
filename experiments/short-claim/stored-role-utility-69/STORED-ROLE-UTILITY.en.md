# Utility by role from stored rankings

**One standard-library Go aggregation passed**, joining the stored rankings from 58 to the existing roles from 66. No new rankings or baselines were executed. Primary scope A retains **all 51 original known/no_answer requests** and reproduces the original best-control cost of **91 checks for BM25**, versus **73 for the oracle**.

This is a cheap integer derivation from the same original 72-request development corpus. It is not combined-76 utility, a training result, generalization validation, model accuracy, CPU/RSS measurement, or LLM savings evidence. Training_ready remains false.

## Rules frozen before the aggregation outcome

The pre-execution `execution-plan.json` SHA256 is `5219abe9c13d25d653b251930b83647013a0a87ae551290c43f90d1de938dca9`. An exclusive invocation receipt and result file were created in a fresh directory before the single aggregation. Failures and retries were both 0.

- **Primary A:** retain original whole-function truth, acceptable sets, and every candidate. Caption-fidelity masks affect a future loss only; they do not make candidate checks free.
- **Diagnostic B:** separately aggregate known/no_answer parents whose entire original candidate sets have true loss masks. B never replaces A, 58, or training-readiness conditions, and no gate is applied to B.
- Select **one control with the lowest aggregate check cost** within each role or the overall scope. Ties use fixed_order → bm25 → lexical_ordered → narrow_rule. Never select the best control separately for each parent.
- Answerable requests stop at the first acceptable candidate; no_answer requests exhaust every candidate, including under the oracle. Preserve all 21 unknown requests and 63 unknown positions, excluding them only from truth-cost and Top1/Top3 denominators.
- Top1/Top3 use answerable-request denominators. The existing 5% necessary comparison uses exact integers in A: `20 × (best checks − oracle checks) ≥ best checks`. It is neither sufficient readiness nor a new threshold.

The full outcomes from 58, roles, groups, and masks were already observed. The new freeze prevents adapting these derivation rules to their results; it does not make this blinded validation.

## A: every original whole-function request

| Scope | Known: answerable + no_answer | Retained unknown requests/positions | Fixed / BM25 / lexical / narrow checks | Single best | Oracle | Exact headroom fraction | Existing 5% necessary comparison |
|---|---:|---:|---|---|---:|---|---|
| Train role | 30: 20 + 10 | 14 / 42 | 60 / 51 / 53 / 51 | BM25 51 | 42 | 9/51 ≈ 17.6471% | Pass |
| Validation role | 12: 8 + 4 | 4 / 12 | 25 / 23 / 22 / 23 | Lexical 22 | 17 | 5/22 ≈ 22.7273% | Pass |
| Calibration role | 9: 6 + 3 | 3 / 9 | 24 / 17 / 17 / 17 | BM25 17 | 14 | 3/17 ≈ 17.6471% | Pass |
| Overall | 51: 34 + 17 | 21 / 63 | 109 / 91 / 92 / 91 | BM25 91 | 73 | 18/91 ≈ 19.7802% | Pass |

These fractions describe the remaining possible reduction in checks if truth were known. They do not establish that a fitted model reaches the oracle or saves resources. The sum of role-specific minima, 51+22+17=90, mixes different controls and therefore **does not replace the overall single-control baseline of 91**.

| A scope | Top1: fixed / BM25 / lexical / narrow | Top1 denominator | Top3 for each of the four controls |
|---|---|---:|---|
| Train | 5 / 12 / 11 / 12 | 20 | 20/20 |
| Validation | 1 / 4 / 4 / 4 | 8 | 8/8 |
| Calibration | 0 / 4 / 4 / 4 | 6 | 6/6 |
| Overall | 6 / 20 / 19 / 20 | 34 | 34/34 |

The stored narrow_rule control fell back to BM25 on all 72 original inputs. This aggregation did not create new successful narrow-rule cases.

## B: separate complete-caption parent diagnostic

| Scope | Included known: answerable + no_answer | Excluded known requests/positions | Fixed / BM25 / lexical / narrow checks | Single best | Oracle | Headroom |
|---|---:|---:|---|---|---:|---|
| Train | 21: 13 + 8 | 9 / 29 | 43 / 33 / 35 / 33 | BM25 33 | 29 | 4/33 ≈ 12.1212% |
| Validation | 11: 7 + 4 | 1 / 4 | 23 / 21 / 21 / 21 | BM25 21 | 16 | 5/21 ≈ 23.8095% |
| Calibration | 5: 3 + 2 | 4 / 12 | 12 / 8 / 8 / 8 | BM25 8 | 8 | 0/8 = 0% |
| Overall | 37: 23 + 14 | 14 / 45 | 78 / 62 / 64 / 62 | BM25 62 | 53 | 9/62 ≈ 14.5161% |

B includes 108 known candidate positions and excludes 45 from that diagnostic. It still records the original 72/216 and unknown 21/63. All 14 excluded requests and 45 positions remain in A's check-cost denominator. The 0% calibration headroom in B did not introduce a new stop gate or sample floor.

| B scope | Top1: fixed / BM25 / lexical / narrow | Top1 denominator | Top3 for each of the four controls |
|---|---|---:|---|
| Train | 2 / 9 / 8 / 9 | 13 | 13/13 |
| Validation | 1 / 4 / 3 / 4 | 7 | 7/7 |
| Calibration | 0 / 3 / 3 / 3 | 3 | 3/3 |
| Overall | 3 / 16 / 14 / 16 | 23 | 23/23 |

## Input and execution evidence

The tool hashed and decoded the same bytes of 7 files: result58, plan58, actual66, mask-role audit69, and the original probes56/probes56b/results56b. It joined all 72 parents and 216 positions to original order, IDs, nullable labels, and acceptable sets, checking 288 stored permutations, 288 stored control costs, and 72 oracle costs. Unknown cost 0 does not mean unknown truth was relabeled negative.

The original source freeze `3c2bde94db7661bf5cdb268f0a2ede0f0b962a94` and input freeze `bb2e6d2d137ce0626ccbfba1ee591a9814ec8153` from 58 were preserved. All 9 source files from its original plan match current bytes and SHA. Key files are `pkg/shortclaim/baseline.go`, SHA `ffb9f29ddc911d16b03a07cfe5c91bb99d8ae7f1d8aa1ea51edd651aee4bb00e`; `pkg/shortclaim/input.go`, SHA `3b296ca0835c4397af7444a2f956aeae59638dd350cbc861bf36dd81bc210ecb`; and `internal/lexicalhint/features.go`, SHA `20224bcab1ea2d6e5203a2efb7772adef0786eea67f43a003a5b2db896804557`. There was no new Git-blob lookup or trusted-compiler proof. The scope is a binding to the stored 58 provenance and unchanged current source bytes.

The actual result `official-attempt-1/results.json` is **211,990 bytes**, SHA256 `f0e766a82d81b3c3897a510f212629c39e987a7413a6e6e541ccfae6467431fa`. The receipt is **282 bytes**, SHA256 `00bb2c3781ad8c4144162d7973a4298257acef40481e0a783a5cd7bf536c173f`. The result retains each of the 72 parents' original rankings, acceptable sets, role, A costs, and B inclusion/exclusion reason.

Synthetic race checks passed 6 tests initially, 6 after adding explicit diagnostic fallback/exclusion denominators, and 1 additional affected denominator test: 3 invocations and 13 named-test executions, with 0 failures. Vet passed twice; the worker build and plan seal each passed once. The worker uses Go 1.27.1, CGO 0, and trimpath, with no executable dependency on the original repository packages.

New Baselines, ranking, Features, Project, Assign, candidate/upstream API, fit, and separate model calls were all 0. The original 1 Assign execution was preserved. Ordinary Codex collaboration cost was neither measured nor claimed to be zero. Shared repository files, inputs, roles, truth, masks, and old frozen plans were not changed.

The next first Project is a separately frozen execution that creates Prepared from original raw text through Validate and binds every original truth/mask/role. This derivation alone does not establish training readiness or a model-selection policy. Combining the two source families from 68 requires a new whole-membership mapping and a fixed seed1729 plan, rather than appending requests to old roles. The separate target of at least 2,400 protected-final requests per domain remains outstanding; it is not a minimum for every development fit.
