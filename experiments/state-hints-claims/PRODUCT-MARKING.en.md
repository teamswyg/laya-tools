# Product review of the mark contract v1

2026-10-07 · Product-criteria proposal · Human references not collected · No readiness claim

## Conclusion and source boundary

The three attributes broadly fit the goal of cheap Go hints for response requests, current development reports and completion reports. The meaning of the visible marks still needs human confirmation. The proposed first trial unit is **a claim in a version-bound comment**. A mark representing whole-task state needs a separate adapter contract and human criteria.

Sources read: public [issue155](https://github.com/teamswyg/laya-tools/issues/155), `experiments/state-hints-claims/RUBRIC.ko.md` and `.en.md`, and `pkg/statehintclaims/model.go`. The issue records 23 actual research Fits, no operationally qualified model, zero application writes and default OFF. The recent MLP failed every frozen research condition; completion precision was 24/32. Zero linear completion proposals establish neither precision nor product success. No private raw material, model or evaluation file was read for this audit.

## Where marks and learned labels differ

| Candidate mark | Actual current attribute | Contract to verify in the first trial |
| --- | --- | --- |
| ❓ Question | `response_requested`: an answer, explanation, confirmation result or choice is requested. Includes requests without a question mark; an imperative with no response request is not automatically included. | Is “response requested” the useful visible meaning? Is it useful when the addressee is unresolved? Do not equate the attribute with grammatical questions. |
| 🛠️ Progress | `current_activity_claimed`: a direct report that identifiable development work is currently underway; includes directly reported third-party work. | Attach author and task/subtask scope. Distinguish a report from observed execution, and bind event version so an old report does not silently become current task state. |
| ✅ Complete | `completion_claimed`: task/subtask completion or completed verification is reported; tentative and directly reported third-party work qualify. | Does the mark communicate “completion reported”? Do not expand a completed subcheck to whole-task completion. Keep report presence separate from success evidence. |

Each head has true/false/unknown, and multiple heads may be true. False means absence of a positive claim, not unfinished work or absent activity. Log semantic unknown separately from confidence/margin abstention. Preserve clearly asserted conflicting reports and record the conflict; respect explicit corrections or retractions. A reply after completion does not automatically reopen or undo the task. Review a new response request, same-scope correction and other-subtask report separately. Handling simultaneous positives in a catalog that permits only one mark remains human-pending.

## Minimum input for a first private read-only trial

Collect consecutive new comments from a specified development task and time window. An initial 30–50 comments is a proposed workflow check, not a new qualification threshold. Preserve privately:

- Original body and verified whole-body extraction, locale, comment ID, author ID (local pseudonyms are sufficient), posting time, and boundaries separating quoted/code content.
- Team/task ID, a minimal task/subtask scope description, relevant author/addressee relationship to the task, necessary preceding conversation and reply target. Keep unavailable relationships explicitly unverified.
- Reader provenance, full opaque revision/event version, edit/deletion indicator, collection time, and verification binding body to that revision. Changed text is a new comment version; invalidate attachment candidates from the older version.
- Whether the current catalog can represent response/activity/completion reports, and whether placement is on the comment or task. If execution/test evidence exists, retain its source, time and scope in separate fact fields. Its absence does not turn a completion report into false.

Current `Scores`/`Predict` accept text only; provenance, event facts and task state do not enter numeric features. The context above belongs to the **adapter and human review**. The current model does not provide author, scope, addressee or evidence spans. Missing fields mean that integration evidence is missing; masking every semantic outcome as unknown cannot establish a useful product.

## Local logging and human references

Before collection, record mark names, units, handling of simultaneous positives and the trial window. A person familiar with the work labels each head true/false/unknown before seeing predictions, records author/scope/evidence spans, and judges the mark useful, unnecessary or misleading with a reason. Where feasible, obtain an independent second review and retain original judgments and disagreements. Items without human participation stay human_pending. Repeated AI annotation alone does not establish human references or usefulness.

Text-free logs retain local event pseudonyms, revision-binding status, model/rubric version, each winner/state/confidence/margin/unknown_reason, proposal/abstention/conflict/invalidation reason, read/extraction/inference time and resources, and whether a human reference exists. Raw bodies, evidence spans and native IDs stay in restricted local reference files. Group related thread/task events and translated pairs; record exposure history separately from research materials. Do not open sealed calibration/final-test material or reuse exposed evaluations for selection/per-case tuning.

Keep .9 confidence/.05 margin unchanged. Report proposal precision and support with denominators by locale/head, together with true recall, coverage, helpful proposals and missed/misleading costs. All-abstention is not product success. Preserve the issue’s existing completion .98 precision plan, at least five correct events/two declared lineages per target intent/locale, 10/3/1 error costs and abstention comparison. Those values come from an existing adjudication plan; numeric passing or declared IDs do not automatically establish three-head operational qualification or independence.

## Concrete experimental choices

| Choice | Cost and benefit | Evidence scope and limit |
| --- | --- | --- |
| A. Frozen text-only private comment shadow | Smallest start: human review of proposals from consecutive new comments at existing gates. | Measures claim detection, abstention, misunderstandings and coverage in actual writing. Missing context, task-state adjudication and app integration remain unresolved. |
| B. Context-bearing reader shadow | Attach author, reply, scope and revision; humans judge identical comments with and without context. Adds read cost. | Shows which missing inputs change useful marks, and checks replay/edit/post-completion replies. Does not prove a better new model or qualify application writes. |
| C. Explicit events alongside comment hints | If trusted execution/completion events already exist, compare their separate fact path with comment hints. Requires contract checks. | Shows report-versus-fact differences and where hints add value. Does not supply truth for eventless scopes or whole-task completion. |

Proposed sequence: A for practical mark criteria, then B when missing context is identified. C is conditional on existing trusted events. This review documents a proposal from public sources. Already authorized private read-only shadow does not need a new approval workflow. Default OFF for automatic app writes is the existing research posture, not an invented permanent user prohibition. Inference-model routing, new Fits, code/app changes are not established by this review.

[contract table](PRODUCT-MARKING.contract.csv) separates source constraints, assumptions and human/integration-pending items by criterion ID without raw text. Established means explicitly recorded in a source; proposed and pending states remain unverified. The table contains no human labels, utterances or performance measurements.
