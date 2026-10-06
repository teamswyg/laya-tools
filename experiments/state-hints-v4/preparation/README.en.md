# V4 preparation: development progress, completion-report and question hints

The target is inexpensive semantic assistance from a tiny Go model. Completion
reports remain an author's assertion, not proof of successful work or authority
to change task state. Lowering Codex models or reasoning levels is outside this stage.

The proposal is **1200 distinct original scenarios × one Korean and one English
wording = 2400 rows**. Train/selection/calibration/final-test partitions contain
840/120/120/120 families. Translations and derived contrasts stay together.
Changing names, sentence order or filler does not create a new scenario.
Related scenarios share a leakage lineage. The count 1200 does not establish
statistical independence or product ground truth.

The same eight intents are question, blocker, reference, progress,
completion_report, cancel_request, planned and unclear. Each has 150 families,
split 105/15/15/15. API, types, tests, builds, concurrency, cache, UI, docs and
deployment are mixed across intents. Both languages use the same visible-input
rubric: the current unit, ongoing work, required remainder and primary reply
request must be recoverable from the text. Headers/code use separate role controls.

Before any model scoring, the rubric was frozen and the first 20 bilingual
scenarios were authored. Independent AI review disputed the question/unclear
interpretation of one family. The original 40 rows were retained, and its two
rows excluded from the prepared subset. The remaining 19 families / 38 rows are
AI-reviewed preparation, not a human-verified golden set or the complete 2400
corpus. No new fit, final evaluation or operational activation has occurred.

The [Go structure checker](../../../internal/statehintcorpus/README.en.md) checks
translation pairs, declared lineages and bounds, emitting counts only. Semantic
labels and rights remain separately reviewed. A new fitting driver must freeze
complete source/data/split hashes and selection rules, then open final-test text
after selection. Existing V3 code/results, models and thresholds are preserved.

The proposed known-method training reuses the Go classifier's cross-entropy and
AdamW: three independent warm starts from frozen v0.2 at learning rates
.005/.01/.02, plus a fresh .02 arm. Parent weights inherently include historic
learning; old exposed evaluation rows and private work text are not resampled.
This is not Laya neural-network, LoRA or ternary-weight training.

Validation-only selection emphasizes false completion costs. Per family, the
cost adds 10/3/1 for any wrong completion/question/progress proposal plus missed
target-row cost. These are predeclared engineering weights, not universal
optimal values from a paper. Each display class/locale must have correct gated
support from at least 5 families and 2 **declared lineages**, completion precision
at least .98 and cost below the all-abstain comparator. Distinct lineage IDs still
need a semantic ancestry audit.

The .9 confidence/.05 margin gate stays fixed. Only one eligible selected arm
receives shared scalar temperature calibration on a separate partition. If
validation or calibration qualifies no candidate, the final-test seal stays
closed. A successful candidate will be saved/reloaded before one final stage.
The [Go selection/calibration driver](../../../internal/statehintfit/README.en.md)
and [family evaluator](../../../internal/statehintfamily/README.en.md) are now
implemented. Current checks use original arithmetic/metadata and stage callbacks,
not actual new training quality. Complete-data review, concrete locks and the
real fitting run remain preparation work.

[한국어](README.ko.md) · [Public tracking issue](https://github.com/teamswyg/laya-tools/issues/155)
