The hypothesis that richer captions automatically improve a learned ranker was not supported on these three development goals. With requests, code truth, candidate order and model coefficients fixed, caption B worsened model71 from 3 to 8 total checks and model72 from 5 to 6. Existing lexical controls improved from 8 to 3. Supplying useful descriptions and having a particular model use them well are separate questions.

Checks sum the position of the first acceptable candidate; fewer is better. All nine candidates remained in both arms. Top1 below has three answerable parents as its denominator.

| Fixed method | A original-caption checks → B behavior-caption checks | Top1 A → B |
|---|---:|---:|
| FP32 model71 | 3 → 8 | 3/3 → 0/3 |
| FP32 model72, BCE+rank λ1 | 5 → 6 | 1/3 → 1/3 |
| Original candidate order | 6 → 6 | 1/3 → 1/3 |
| BM25 / lexical_ordered | 8 → 3 | 0/3 → 3/3 |

narrow_rule used its existing BM25 fallback on prose requests; this is not separate successful semantic reasoning. Three candidates per parent make every method's Top3=3/3 uninformative. Model72 also varied by goal: datasize worsened 1→3 checks, query stayed 2→2, and shlex improved 2→1. An aggregate average hides these differences. Model71's perfect A result does not establish understanding or actual-work qualification.

Source reading establishes that the scorer builds words and adjacent bigrams, hashes signed request×caption associations into 8192 columns, and sums them with fixed coefficients. Longer text changes both associations and normalization. It is not an interpreter of return-versus-panic, negation or conditional state changes. [Pinned feature implementation](https://github.com/teamswyg/laya-tools/blob/7cea49090727555924bf22679ffeccc4f38c9865/internal/hintlearn/learn.go#L118).

Possible explanations include a mismatch between training phrasing and B's behavior descriptions, or missing learned distinctions among candidates that share vocabulary. Hash collisions and length effects are structurally possible too. B changes content, length and style, so this run does not isolate information from phrasing. It cannot identify the dominant cause: no new coefficient attribution, collision or training-distribution measurement was performed, and neither objective nor features changed.

The recommended next step is **prepare new comparison problems before another fit**. Use new public behaviors and requests chosen before inspecting model scores, with similar vocabulary but meaningful differences such as returning an error versus panicking, changing state only under a condition versus always changing it, or negation/exception boundaries. Captions must remain faithful to real source; flipping a sentence is not permission to invent a false caption. Keep aliases/helpers in whole families. Freeze requests, captions, scoped source rubric and existing comparator rules before a once-only fixed-model diagnostic. Do not revise truth or retain only favorable cases after seeing scores.

An alternative, after repeated failure conditions appear on those new problems, is a separately versioned ablation of small clause/negation/return-behavior features. More epochs, new seeds or lambda/threshold searching now have weak support. On current B, lexical controls already reach the oracle lower bound of 3 checks: additional headroom is 0. Re-fitting this slice cannot beat that lower bound. Its common A/B training-eligible intersection still has zero positive pairs; this fixed-model diagnostic does not make it a trainable paired-learning comparison.

Resource numbers are a single root-collected whole-process sample: peak RSS 10,551,296 B (10.0625 MiB), displayed OS real 2.23 s/user 0.05 s/system 0.17 s; external wrapper wall 2.24256375 s. Worker snapshot wall 1.853498208 s, final Go HeapAlloc 2,506,064 B and cumulative TotalAlloc 5,806,904 B have different scopes. Validation, checkpoints and fsync are included. Dividing total runtime by 36 would not establish pure inference latency or LLM savings. OS values are root-reported; this analysis did not independently reread raw OS logs.

This is AI-assisted, nonblind development analysis by the B-caption and runner author. Three exploratory source-rubric parents are not universal truth, training labels or an independent held-out golden set. Preserve historical validation failures, both 32,792-byte failed HF models, unknown/no_answer/mask/role records and false qualification/production gates. Different new behaviors, requests without prior score exposure and the protected 2,400-example final evaluation remain necessary. New scoring/Features/Normalize/Decode/Fit/native/HF executions in this analysis are 0.

Evidence: saved single-run result SHA `0e7373b024f0747dbe076875a0f4d7df845a35cf8db967cdea731c14b0dee6fd`, source commit `7cea49090727555924bf22679ffeccc4f38c9865`, feature-source SHA `d0c906d1283617d0778039ff8bc4c6fff6cb0af3a813c48df088424d78bebb71`. Only saved result/counters and fixed feature source were read. No rerun, model selection or publication was performed.
