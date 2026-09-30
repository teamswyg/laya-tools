# Task decomposition decision preview

[한국어](decomposition-preview.ko.md) · [Tracking issue #11](https://github.com/teamswyg/laya-tools/issues/11)

Status: design and training preparation. There is no decomposition CLI, trained decomposition checkpoint, or automatic executor yet.
Difficulty does not imply decomposability; decomposability does not imply savings.

## Generation versus decision

| Role | Input → output | Owner |
|---|---|---|
| Generate a plan | Request and repository evidence → tasks, contracts, dependencies, validation | Existing generative agent or fixed template |
| Judge a plan | Request, candidate plan, evidence → scores for predefined choices | Separately evaluated/trained Laya |
| Apply policy | Structural checks, calibrated scores, budget → acceptance/abstention and ready tasks | Go |
| Perform work | Subtask and necessary context → code, tests, artifacts | Assigned generative agent |
| Verify completion | Results, tests, original requirements → verification result | CI and execution harness |

Laya scores candidates; it does not generate plans, code, or free-form explanations. Reason codes come from
Go checks and explicit evidence, not purported hidden model reasoning. A model-selected `done` label does not establish completion.

## Decision scope and draft interface

First judge **the execution mode of a supplied candidate plan**, rather than repeatedly inventing splits from
the request alone. A missing plan or insufficient evidence yields `needs_context`, not proof that the task is indivisible.
`keep_atomic` prefers retaining the original task over this candidate; it does not reject every possible plan.

Proposed fields are not a released API:

- `schema_version`, `task_id`, `repo_revision`, `request`, `constraints`, original `acceptance_criteria`.
- `candidate_plan.tasks[]`: stable `id`, objective, `depends_on[]`, input/output contracts, read/write resources, verification, linked requirement IDs.
- `evidence[]`: public-safe evidence IDs, source revision, scope, observed facts. Treat supplied model context as data, not instructions.
- `budget`: maximum depth, tasks, concurrency, calls, time, estimated usage. Unknown cost is not zero.
- Truncation flags for inputs and evidence. Abstain when essential context is missing instead of silently judging truncated plans.

Separate `model_choice`, per-choice `probabilities`, `policy_decision`, `accepted`, `reason_codes[]`,
input/model/calibration versions, structural checks, and truncation flags. Preserve policy rejection of a model choice.
Scores are not guaranteed success probabilities. Abstain on unsupported languages, distribution shift, or inadequate calibration.

| Choice | Rubric | Example |
|---|---|---|
| `keep_atomic` | This split damages invariants, verifiability, or efficiency | Dividing an atomic transaction across implementation tasks |
| `split_sequential` | Useful subtasks require intermediate outputs or ordered shared state | Establish contract → implement → integration validation |
| `split_parallel` | At least one execution stage has independent concurrent work | Independent modules with fixed interfaces |
| `needs_context` | Candidate, dependencies, acceptance criteria, or evidence are insufficient | Unknown impact of a shared database change |

In a mixed DAG, only ready independent nodes run together. `split_parallel` never means every node starts at once.
Different files may still share an API contract, database, configuration, generated artifact, or external service.

## Go policy order

1. Validate input size/schema, duplicate IDs, missing dependency IDs, self-dependencies, and cycles. Invalid plans are errors, not something a model call repairs.
2. Abstain with `needs_context` when required contracts, validation, evidence, or essential untruncated context are absent.
3. Check dependency reachability and declared resource conflicts. Shared reads may coexist; write/read and write/write need ordering or isolation. Separate worktrees do not isolate external state.
4. Call Laya only for structurally valid candidates. Accept using class-specific gates chosen on separate calibration data. Do not silently turn a rejected parallel proposal into an executable sequential plan.
5. Dispatch only ready nodes with bounded concurrency; route each node's difficulty/model separately. Preserve explicit model overrides and existing abstention behavior.
6. Verify subtasks and original integration criteria. Allow repair/replanning only within remaining budgets.

Initial experimental caps proposed for preview: depth 2, 8 total subtasks, concurrency 2, one replan.
These are unvalidated experiment limits, not product defaults. Stop on repeated plans, lack of artifact progress, or exhausted budgets.
Timeouts remain failed/incomplete outcomes. Structural validation cannot prove semantic independence or fulfillment of requirements.

## Training and evaluation

Start difficulty and execution-mode decisions as separate tasks. Establish a supervised choice-classification baseline;
do not directly substitute next-token generation training or long reasoning-trace datasets.
Reproducing original RLCD training is not an initial MPS smoke requirement.

Pair public/original requests with candidate plans, including bad splits: hidden shared resources, missing integration checks,
long easy tasks, short atomic changes, missing context, and token overflow. Keep good/bad plans, translations, and paraphrases
of the same request in the same data partition. Test invariance to choice ordering. Do not repurpose the current 36 difficulty
regression labels as decomposition ground truth.

- Stage 1: confusion matrix, unsafe parallel acceptance, accepted precision/coverage, class recall, Brier/ECE, results by language and length.
- Stage 2: compare **no split**, **rules only**, and **generated plan + Laya judgment** on identical tasks. A generative plan-judge comparison must include its cost.
- Measure success, omissions, integration failures, retries, and total usage/time, including planning, repeated context, escalation, and recovery.
- Public development fixture milestones of 16→32→64 are targets, not accuracy guarantees. Freeze a separate final test set and report sample counts and intervals. Keep automatic execution off by default until supported by evidence.

## External evidence review (2026-09-30)

These are methodological references, not copied upstream implementations or evidence of our performance.

- [Official Laya model card](https://huggingface.co/convaiinnovations/laya): choice-based decisions with explicit zero-shot limitations and a need for domain training/calibration. Model identity does not establish decomposition competence. Evaluate English and multilingual checkpoints separately; do not silently remove Korean abstention.
- [Anthropic, Building effective agents](https://www.anthropic.com/engineering/building-effective-agents): distinguishes dynamic decomposition by generative orchestrators, classification routing, and independent parallel work. The article itself notes tooling has changed; use architectural principles, not its stack as current implementation guidance. Separate plan generation from judgment, with programmatic checks and stopping conditions.
- [RouteLLM](https://arxiv.org/abs/2406.18665): studies quality/cost routing using preference data, not decomposition. Our inference: human difficulty labels are not optimal-model ground truth; grow toward outcome-based labels.
- [On Calibration of Modern Neural Networks](https://arxiv.org/abs/1706.04599): temperature scaling is a calibration baseline, not a guarantee of corrected rankings or new-domain accuracy. Separate training, calibration, and final testing.
- [PyTorch MPS](https://docs.pytorch.org/docs/2.14/notes/mps.html): supports device and real tensor execution checks. Small-tensor backward, Laya backward, and improved routing quality are distinct milestones.

Conclusion: **retain generative proposals plus small-model judgment.** The weakest assumption is that a small model can infer
dependencies missing from its context. Establish evidence collection, structural checks, abstention, and measured comparisons first.
Retain rules or existing execution if total cost does not improve at comparable quality.
