# PDCA54: compare one actual parser behavior request

[한국어](RESULTS-54.ko.md)

One new request bounds JSONL event-envelope keys to64 entries and128 decoded UTF-8
bytes, rejecting duplicates. Sol6/Luna low each ran twice. All four candidates
passed75 independent terminal tests, exited zero and reported complete core usage.
The tests and repetitions do not add unique requests.

## What was fixed before outcomes

[Plan54](plan-54.json) was committed at `063f274f2a3a1776b793447cdb4b51be09a93743`
before outcomes: Sol6/Luna/Luna/Sol6, concurrency1, main120s plus independent
verification45s, without executor retry/resume/fallback. Only the existing local
ChatGPT login was used; no new API key, paid job or endpoint was created.

The executor was built cleanly from CI-passed `4373f9f6b20dba922b1cf5309d079e50a4591aeb`;
the public task base and contract bind revision `ee72334166e2962b0821d5198a50fdd13f92ab29`.
The new external-repository contracts were not used for this fixed comparison.

[The native Laya prediction](routing-predictions-54.json) was also committed before
coding, at `5aae164b788c638ad6b9f296b7346567f68cb977`. Unchanged INT8 base, CPU1thread
and threshold0.9 proposed standard with score **0.610351**, therefore abstained.
Configured fallback was strong/Astra; Astra was not executed. One cold whole-process
observation took **1,183ms**, app observation1,151.896ms, maximum RSS
**1,508,032,512bytes** (about1.51GB/1.40GiB). This includes native loading; it is
not warm latency, Go heap or GPU resources. The score is not calibrated coding-success
probability, and this memory does not meet the ultra-small-resource target.

## Actual observations

| Order | Requested profile | Main ms | Verification ms | Input | Included cached input | Output | Included reasoning | Outcome |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| 1 | Sol6 low | 66,431 | 4,750 | 163,569 | 141,568 | 2,036 | 48 | Accepted, zero exit, complete usage |
| 2 | Luna low | 28,981 | 4,470 | 78,701 | 66,304 | 663 | 0 | Accepted, zero exit, complete usage |
| 3 | Luna low | 28,488 | 4,503 | 85,762 | 71,424 | 665 | 0 | Accepted, zero exit, complete usage |
| 4 | Sol6 low | 61,574 | 4,730 | 157,577 | 144,256 | 1,918 | 109 | Accepted, zero exit, complete usage |

Input485,609 includes cached423,552; output5,282 includes reasoning157. Sum of
main times185,474ms and separate verification18,453ms excludes parent preparation.
Luna had smaller reported input/output and observed main time in both repetitions
of this request. Uncached input sums were Luna26,735 versus Sol6 35,322. Subscription
consumption, billing and internal provider calls/retries remain unknown, so these
numbers **do not establish monetary savings**.

[Numeric records and hashes](results-54.json) match each private executor terminal
`record.json` byte-for-byte. All four observed process-group cleanup and removal of
copied auth; candidate-written tests did not determine acceptance. Requested models
are explicit, while provider-attested served identity remains `unknown`. Sources
1/4 are identical; sources2/3 differ from each other. Passing a shared contract does
not mean identical implementations.

## Integration into the real tool

The first zero-exit accepted mutable `pkg/taskoutcome/summary.go` was integrated
unchanged, SHA256 `64bbcfd17d3321b74f1b679142baa79906ded759458a245ade6a546172e7deb5`.
A fixed `[64]string` array resets per event and checks decoded-key count, UTF-8
byte length and duplicates. Nested items remain opaque, and unknown names/values
are not retained in Summary. No shared state or locks were added.

The six previously frozen independent external-package tests were copied into
runtime regression with test names changed only. Candidate-authored tests were
not adopted. Independent review confirmed the final missing-type condition is
equivalent under the original type validation. Error precedence when one input
violates both a size and duplicate rule is outside the frozen assessment. This
code change does not validate routing utility or cost savings.

The first Linux CI exposed an incorrect test expectation: a formatter-normalized
UUID baseline without the new API requires isolated compilation, so a platform
without supported isolation must report `verifier_unknown` rather than rejection.
The test now checks that unknown outcome and actual isolated rejection on macOS
separately. Product decisions, frozen specs, parser observations and original
source files remain unchanged.

The next macOS CI reached the default ten-minute timeout for the test package
that aggregates multiple cold isolated compiler checks. Its current control had
run for four seconds. Only the outer CI race-suite budget is explicitly twenty
minutes, with a thirty-minute test-job cap. Each candidate still has the same
45-second default / 60-second maximum verification timeout, ten-second inner
test timeout, output limits, isolation and acceptance criteria. This change is
not a model execution-time or performance observation.

## Scale and next steps

Actual cumulative development observations now cover **5 distinct requests,
16 CLI records,3 code families and1 repository**. Earlier failures, support errors
and timeout remain unchanged. The8 local candidates and [120 public Go candidates](../../docs/public-go-acquisition-53.en.md)
are neither execution nor label counts. [Independent contracts](../../docs/public-go-contracts-54.en.md)
are prepared for2 of those120 external candidates; external model attempts,
training and final eligibility remain0.

24 development probes cannot justify general performance claims. Expand to120+
independently verifiable development requests across repositories and240+ requests
for usable training-signal assessment. Separately acquire and protect **at least
2,400 distinct final requests per claimed domain**. Previously read requests,
repetitions, translations and sibling tasks do not inflate final counts. The
[scale guide](../../docs/golden-set-scale.en.md) and [acquisition plan](../../docs/golden-set-acquisition.en.md)
describe grouping and minimum execution budgets. This cycle performs no fitting,
calibration, new weight release, production activation or protected-final scoring.
