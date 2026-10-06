# Evaluating cached scores by scenario family

`Evaluate(corpus, predictions)` consumes cached scores in original row order and
evaluates paired Korean/English scenarios. It constructs, fits and predicts no
model and changes no data. Reports contain no text, IDs or lineage names. Valid
declared score source/lineage does not establish model provenance, semantic truth,
rights or statistical independence.

Eight columns preserve the original intent order; display arrays use
progress/completion-report/question/none, locale arrays ko/en. The actual gate
uses the original eight-column winner, learned source, positive training steps,
no guard, confidence >=.9 and margin >=.05. None probability sums the other five
columns; it is not an abstention bit. Its argmax never replaces the actual winner.
NLL uses a disclosed1e-15 floor and counts clipping; Brier sums squared deviations.
Average two rows within each family, then average families.

Any wrong completion/question/progress proposal within a family incurs10/3/1,
plus the paired mean missed-target-display cost. Eligibility requires >=.2 correct
coverage, >=5 correct families and >=2 correct **declared lineages** for every
display class/locale. Completion proposals need a nonzero denominator and >=.98
precision; cost must be strictly below.375. With15rows per class/locale, five
families means33.3%correct coverage. The driver owns full split quotas and original
ancestry/semantic/rights review. `Eligible` is neither production approval nor a
ground-truth certificate.

Untrained and rule scores have descriptive raw confusion/loss metrics but never
pass the learned display gate. Rule one-hot scores are not calibrated empirical
probabilities. The report makes no calibration-success claim; a control's forced
all-abstain cost is not utility evidence. Original hand-computed arithmetic
fixtures use no learned model or actual product accuracy evaluation.

[한국어](README.ko.md)
