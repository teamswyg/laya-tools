# Training rubric for three claim attributes

Scope version: **development-work reports v2**. Preserve the initial broad-action v1 rubric and its AI annotations as local research history. No model was fitted to those annotations.

The target is inexpensive detection of response requests, current activity claims and completion claims in development comments. It generates no prose and has no authority over actual task state. This study is separate from the existing eight-primary-intent classifier.

Judge each attribute separately as `true`, `false` or `unknown`. Several attributes may be true together. Separate output normalization does not imply statistical independence.

| Order | Attribute | Meaning |
| --- | --- | --- |
| 0 | `response_requested` | The message actually requests an answer, explanation, confirmation result, choice or other response. |
| 1 | `current_activity_claimed` | The author directly reports development work currently in progress: edits, implementation, review, verification, build or deployment. |
| 2 | `completion_claimed` | The author directly reports completion or completed verification of an identifiable development task or subtask. |

`true` means a positive request or work report is present, not that it is factually verified. Directly reported third-party work is included. A tentative completion report is also a completion claim; record its hedging separately. `false` means the absence of the corresponding positive request or work report is clear, not that the task is unfinished or activity does not exist. Use `unknown` when the message does not resolve the request/report intent, current timing or reported work scope. Never fill missing labels or evidence with false.

Generic program behavior, execution traces and state variables are not automatically development-work reports. “The handler caught an exception” alone does not report completed implementation or verification. “I finished verifying the exception-handling change” does. Directly reported currently running tests, builds or deployments qualify as activity; contradictions in past program traces do not themselves make a current work claim unknown.

Keep claim presence separate from factual adjudication. When conflicting positive work reports are each clearly asserted, preserve their true presence and document the conflict in the evidence. Respect explicit corrections, retractions and negation. Unknown concerns unclear request/report meaning; it does not resolve which actual task state is true. Do not invent different substeps to remove a same-scope contradiction.

Quotation, code examples, hypothetical framing, future plans and negation do not themselves count as positive assertions. Assess any separate direct assertion in the surrounding context. A bare reminder is an unknown response request when its response intent is unclear. An imperative that requests no response is not automatically a question.

An original fictional example, “The type check passed and I am running regression tests now,” reports completion and progress of different substeps together. It does not establish whole-task completion. Do not resolve same-scope contradictions by inventing different scopes.

Annotate from the actual text, not by converting `expected_intent`, prior predictions or word-detection rules. Judge both locales separately without assuming translated pairs have identical attributes. Positive and unknown judgments retain original UTF-8 evidence byte spans and a brief explanation. Clear absence can be explained from the whole context. AI reference annotation is not human gold; disclose prior exposure and correlated AI history.

Supervised training averages cross entropy across three categorical heads. Unknown is an explicitly trained category. The classification-loss definition follows the [official CrossEntropyLoss documentation](https://docs.pytorch.org/docs/2.14/generated/torch.nn.CrossEntropyLoss.html), with execution and training implemented in Go. Results distinguish a learned unknown winner from confidence-based abstention.

Preserve the original 0.9 confidence / 0.05 margin gates, evaluation data and models. The research preview uses the same numeric display gates but does not inherit qualification from the eight-intent evaluation. Development results from the frozen group split inside original training data do not qualify deployment; separate new evaluation and real-comment validation are required. Completion-claim outputs cannot authorize label, emoji or task-state writes.

Three output states are not 1.58-bit/ternary weights. The first experiment uses float32 weights. Weight compression and Laya-encoder comparisons remain separate studies. No private work text or credentials enter this public training material.
