# Input size checks for the next nine requests

The proposed behaviors cover annotation ownership, bounded counts, single
assignment, sensitive defaults, deferred callbacks, preserved error causes,
joined hook errors, remain-field conflicts, and tag precedence. This folder
records nine completed Go input validations for the first nine entries of the
[pinned ten-request proposal](../next60-catalog10-literal-correction/CATALOG10.v2.json).

Each request and candidate caption passed the 32-word and 512-byte limits.
The helper checks text size and normalization only. Original-code execution,
Want comparisons, model inference, feature extraction, scoring, and training
were all zero. These checks do not establish correctness or training eligibility.
The tenth request, unknown-token reporting, remains on instrumentation hold.

- `input.go.txt`: Root-authored helper source, archived as text rather than a runtime package.
- `INPUT-GATE.actual.v1.json`: Actual completions and per-request sizes; no private work inputs.

The current development data still contain [seven requests and twenty labels](../next60-development-seven/README.en.md).
The nine shape-checked proposals are not added to that count.
