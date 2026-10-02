# 76 saved-result check

The saved scores, rankings, counters, and input bindings agree. A Go standard-library checker ran once, without new inference or upstream API execution. This is not a claim that the learned models performed well.

The same three requests each have three candidates. A and B are two versions of candidate descriptions. The check preserves 3 requests, 9 candidate positions, and 18 descriptions. The 36 model score rows are not 36 independent requests. The original plan still has `null` labels, weights, roles, and groups. A separately frozen development rubric specifies acceptable positions `[0]`, `[1]`, and `[2]` for the three requests. This check does not newly approve that rubric's semantic accuracy.

`checks` counts candidates inspected before reaching an acceptable candidate; lower is better. The following values were recomputed by sorting saved scores in descending order, preserving original candidate order for ties. Each arm has the same 3 answerable requests in its Top1 denominator.

| Model/control | A checks | B checks | A Top1 | B Top1 |
|---|---:|---:|---:|---:|
| Failed model 71 | 3 | 8 | 3/3 | 0/3 |
| Failed model 72 | 5 | 6 | 1/3 | 1/3 |
| fixed_order | 6 | 6 | 1/3 | 1/3 |
| BM25 | 8 | 3 | 0/3 | 3/3 |
| lexical_ordered | 8 | 3 | 0/3 | 3/3 |
| narrow_rule | 8 | 3 | 0/3 | 3/3 |

Model 71 has B−A checks of +5; model 72 has +1. narrow_rule falls back to BM25 with `unsupported_rule_request` for all 6 request/arm combinations. Its saved scores and orders are identical to BM25. Selecting the control with the fewest total checks per arm, using the fixed list order for ties, gives fixed_order for A and BM25 for B. This does not select a new model or arm. Every Top3 value is 3/3, which is trivial with only three candidates.

The check covers 12 model parent rankings, 24 control parent rankings, 108 saved scalar scores, and 36 model score-row joins. The 22 actual equal-score pairs preserve original order. No candidate is deleted by a weight or mask: each model/arm retains all 9 scored candidates. Saved counters report 2 model reads/Decode calls, 6 Prepare/Baselines calls, 36 Features/Score calls, and 0 Fit, Project, upstream API, training-label assignment, or paid judge calls. These are saved-counter checks, not fresh instrumentation of internal calls.

The stdout, final, and partial results are exactly the same 17,225 B. The 31,738 B public copy of the original plan is also byte-identical, and all 18 description lengths and hashes match. The original plan retains `frozen_for_execution=false` as historical metadata. Actual dispatch is separately bound to the frozen rubric and the root's before-start arguments/ledger. The zero-call reservation is distinguished from completed execution. Automatic retries remain 0.

The OS log reports whole-child maximum RSS of 10,551,296 B = 10.0625 MiB, peak footprint of 8,258,040 B, real time of 2.23 s, user CPU of 0.05 s, and system CPU of 0.17 s. Controller wall time is 2.24256375 s. The worker's 1.853498208 s snapshot precedes final writes. The Go heap snapshot is separate from OS RSS. No pure inference, startup, or pipe time is inferred by subtracting these observations.

The checker did not author this worker, but had prior source/controller involvement and outcome exposure; this review is not blind. It did not read model files or rerun Decode, Prepare, Features, Score, Baselines, Project, Fit, the worker/controller, upstream APIs, or tests. New candidate/API execution, training labels, roles, weights, and models remain 0. `qualification`, `training_ready`, `production_ready`, and `protected_final` remain false. This adds no independent final-generalization or cost-savings evidence.

Checkable outputs are the [numeric check](NUMERIC-RESULT.v1.json), [receipt](RECEIPT.v1.json), [predeclared invocation](INVOCATION.v1.json), and [attempt ledger](ATTEMPT-LEDGER.v1.json). Private checker source/binary, original logs, and host paths are excluded from public records.
