# Preparing a real 2,400-request evaluation

[한국어](golden-set-acquisition.ko.md) · [Current counts and split principles](golden-set-scale.en.md)

A small pilot of about 24 requests is development data for checking behavior. The next performance assessment targets **at least 2,400 distinct final requests per claimed domain**. Selecting only 24 easy examples from a pool of 2,400 does not constitute that final evaluation. The old, already inspected 24 remain development data.

This is an acquisition and execution plan, not a report that 2,400 actual routing labels exist. Relevance pairs, file-search requests, actual model-selection requests, repository-selection requests, and decomposition parent requests are counted separately.

## Prepare acceptance before inference

Each coding request needs public provenance and revision, usage and redistribution scope, a task family, mutable files, precise completion conditions, independent checks, and their digests. An unchanged baseline must fail, an independently authored correct implementation must pass, and seeded wrong implementations must fail. Candidate-authored tests or the existing suite alone cannot establish requested completion.

The current development comparison covers three contracts from one repository; two are comment changes. Repeating them does not create 2,400 distinct requests. Add behavioral tasks through separate versioned definitions that bind source closure, revision, independent checks, and LICENSE/NOTICE. Preserve previously frozen plans and acceptance contracts.

The initial coding scope is **public Go tasks**. Generalization to other languages, private repositories, or arbitrary Codex work needs separate evidence. The following counts are acquisition targets, not acquired records.

| Stage | Request target | Advancement condition |
|---|---:|---|
| Measurement preparation | 12–24 | Connect start, termination, failure, independent acceptance, and whole-attempt usage. |
| Scope expansion | 120 | Acquire distinct behavioral and bug families across public repositories. |
| Learnability screening | At least 240 | Observe paired differences, both-success/both-failure cases, and variability. |
| Training, selection, calibration | Separate data for each | Freeze counts after checking label costs and group sizes; do not select on final. |
| Final evaluation | **At least 2,400** | Freeze policy, splits, contracts, controls, missingness, and execution budget first. |

These stage sizes are project proposals, not statistical guarantees or automatic permission to fit. When useful signal is absent, revise provenance or objectives before increasing training. Data inspected for revisions becomes development data.

## Request diversity and linked groups

The initial Go final acquisition proposal is below. Confirm feasibility and rights, then freeze a separate plan **before reading outcomes**. These chosen proportions do not describe the natural prevalence of user work.

| Task kind | Proposed requests |
|---|---:|
| Bug fixes involving errors, nil, and boundaries | 600 |
| API, configuration, and schema changes | 480 |
| Parsing, text, and search behavior | 480 |
| Cancellation, state transitions, and retry contracts | 360 |
| Concurrency, memory, and allocation contracts | 240 |
| Integration and compatibility across components | 240 |
| Total | **2,400** |

Acquisition targets at least 30 public final repositories, with no repository contributing over 10%. These targets alone do not establish independence. Group code/template/bug derivatives, translations, parents/children/siblings, forks, and copied source. For unseen-repository generalization, split whole repositories and exclude final repositories from training, selection, and calibration. Report request, repository, and linked-group counts, plus acquired and excluded records per stratum.

Authored fast/standard/strong tiers are **difficulty hypotheses**. Labels derive from actual requested-profile attempts, independent completion, and whole costs on the same request. Do not force larger models to outperform smaller ones.

## Execution budget and labels

Comparing two profiles on all 2,400 requests requires at least **4,800 CLI attempts**. Retries, escalation, child tasks, and repeated experiments add executions, not unique requests. Before full execution, freeze invocation, time, concurrency, stop conditions, and observable account-use bounds in a separate plan. Preparing public sources and checks does not launch this inference batch.

Preselect development requests and repetition counts to inspect variability. Preserve all four paired outcomes:

- Both succeed: a candidate for completing with lower whole cost.
- Smaller fails, larger succeeds: escalation may help.
- Smaller succeeds, larger fails: evidence against a fixed capability ordering.
- Both fail: neither profile is forced into a successful label.

A single outcome supports that attempt, not a permanent capability label. Unknown usage is not zero. A service profile-support error is not coding inability. Include failed attempts, retries, escalation, abstention, and independent verification; cached input and reasoning are subsets of input and output, respectively. Unobserved subscription quota, money, and backend request counts remain unknown.

Prepare fixed explicit profiles, rules, and an abstention-preserving proposed policy as controls. Select thresholds on separate calibration and keep them frozen on final. [RouteLLM](https://arxiv.org/html/2406.18665v4) informs learning from cross-model preferences and quality/cost tradeoffs. Its conversational preference labels are different from our coding acceptance checks, and its savings are not project results.

## Separate domains

Repository selection needs 2,400 original requests with accessible candidate snapshots, permissions, allowed repository sets, multiple valid answers, and no-answer/abstention conditions. Candidate counts and repeated catalog sizes are not request counts. Measure candidate inclusion and final recommendation quality separately.

Decomposition uses **2,400 parent requests**. Children, siblings, and alternative plans stay with their parent group. Separate structural validity from actual execution utility. Compare no decomposition, rules, and model proposals using parent completion and total proposal, execution, retry, merge, and integration-check costs. Checking 2,400 valid DAGs alone cannot prove faster completion.

## Reading the difference between 24 and 2,400

For independent trials with the same error probability and zero observed errors, the exact binomial one-sided 95% upper error bound is `1 - 0.05^(1/n)`: approximately **11.7%** for 24 and **0.125%** for 2,400. This is an illustrative calculation, not observed project accuracy. [NIST on intervals for small samples and rare failures](https://www.itl.nist.gov/div898/handbook/prc/section2/prc241.htm) motivates reporting uncertainty alongside accuracy.

Do not apply this independent-trial calculation directly to 2,400 linked template variants. If only 40 of 2,400 requests receive recommendations, recommendation error uses denominator 40. Even zero errors in those 40 yields about a 7.2% upper bound under the same assumptions. Report recommendation/abstention coverage, errors within recommendations, and failures by repository, language, and task kind. The 2,400 floor starts evidence collection; it does not repair bias or incorrect acceptance checks.
