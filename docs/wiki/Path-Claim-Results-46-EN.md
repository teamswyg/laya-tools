# Beyond 24 tasks: tiny claim-head training on thousands of examples

The initial 24 tasks were a pilot for checking behavior. This experiment used **4,456 training and 2,599 validation examples**, meeting the fixed minimum of 2,400 in each role. All ten models—five helper-cost penalties × two seeds—were fitted. Repeating a task does not add a new example.

The model aims to hint that “auxiliary search may help here.” It neither generates sentences nor replaces the user's decision. Inputs are 16 generic numbers from baseline search. Raw text, paths, task IDs and repository names do not enter coefficient inputs.

**The current family failed its usefulness gates.**

| Validation: 2,599 tasks | Baseline | Learned selection 1729 | Learned selection 2718 |
|---|---:|---:|---:|
| Total pages to first target | 18,023 | 18,010 | 17,978 |
| Helper selections | 0 | 55 | 1,248 |
| Tasks with a target in top 10 | 1,014 | 1,013 | 1,007 |

Page reductions of about 0.07–0.25% missed the required 5%, and early target retrieval declined slightly. Repository harm and seed instability also appeared. Even an oracle knowing every beneficial case could save at most 3.11% with this helper. Neither the gate nor the winner-selection procedure was changed afterward.

Validation counts are Lightning 287, Google Cloud Python 672, NumPy 679, pandas 634 and cryptography 327. The 2,402 protected final tasks remain unscored. Validation selects candidates and is development evidence, not final performance.

One isolated warm execution of input verification, all ten fits, policy comparisons and artifact writing took **1.45 seconds with about 108.6MiB peak RSS**, without a GPU. Coefficient payloads are 64 bytes, but the catalogs, caches and search system are not. This duration is not single-inference latency.

Pages and selections replay stored search outcomes. Actually skipped helper work, LLM tokens, Codex usage and task completion savings are unproven. Failed models were neither registered for production nor released as new weights. A separate experiment should first examine helper headroom and features that distinguish benefit from harm. Ternary comparisons remain conditional on a useful FP32 signal.

[Detailed results, every candidate, provenance and memory](https://github.com/teamswyg/laya-tools/blob/main/experiments/path-cost-claim/RESULTS-46.en.md) · [Preregistered plan](https://github.com/teamswyg/laya-tools/blob/main/experiments/path-cost-claim/plan-46.json) · [Input seal](https://github.com/teamswyg/laya-tools/blob/main/experiments/path-cost-claim/input-46.json) · [한국어](https://github.com/teamswyg/laya-tools/wiki/Path-Claim-Results-46-KO)
