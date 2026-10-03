# Go learning-data preparation that preserves unknown candidates

This proposal connects “do not train unknown answers as negatives” to actual Go
sparse datasets. It stores admitted candidates in contiguous columns and separates
gradient eligibility from validation eligibility. It is not connected to a real
corpus or model yet.

With Go 1.27.1, run from the repository root:

```sh
bash scripts/verify-next60-mask-learning.sh
```

The offline check uses newly owned synthetic inputs. It runs no original observer,
Golden corpus reader, training, model download or publication. Sources stay as
`.go.txt`; verification creates a disposable module. Separate source review and
the real Reader adapter remain pending.

For `[true, false, false, false, unknown]`, only the first four candidates become
learning or evaluation rows. Unknown candidates never become label-zero,
weight-zero placeholders: another metric could count those rows despite zero loss
weights and distort its denominator.

Validation retains four known answers with evaluation weight 1 even when all
gradient flags are false. Tests call the existing Go `pairlearn.NLL` and
`pairlearn.AUC` on these columns. Calibration candidates and ambiguous parents
never enter learning or validation feature extraction; audit counts preserve them.
An empty split is never passed to NLL by these controls.

The structure accepts at most 60 parents with five or eight candidates each. This
is a synthetic structural bound, not evidence of 60 actual semantic requests.
The existing 35-request development corpus and earlier failed models stay intact.

State, label, role and group are not feature-function arguments. Only original
request and candidate text produce features. A separate reviewed adapter must
connect Reader-validated examples to this structure. Structural validation alone
cannot establish source provenance or semantic correctness.

Fixed arrays plan rows. A count and SHA-256 pass precedes exact allocation of
contiguous columns; altered values on the second feature pass are rejected.
Neither this bridge nor its plan uses maps or locks. The existing feature function
first materializes term cross products, so this proposal rejects texts above
32 words before feature extraction. It never silently truncates inputs. This new
bound is specific to the proposal and changes no existing model input policy.

Returned columns, arrays and headers have a calculated 8MiB limit. Temporary
features, the plan, allocator overhead, process RSS and GPU memory are outside
that count. It is not a tiny inference-head memory measurement.

Local arm64 diagnostics, 100 owned synthetic builds for each variant:

| Input | Time/build | Allocated/build | Allocations/build |
|---|---:|---:|---:|
| Five known candidates | 41.8μs | 12,360B | 178 |
| Four known + one unknown | 29.8μs | 9,832B | 144 |

These small-input preparation measurements depend on order and environment. They
prove no inference speed, whole verification-work or Codex cost improvement.
CI does not gate on benchmark timing.

[PR #127](https://github.com/teamswyg/laya-tools/pull/127) automatically merged after
four successful CI gates. Its earlier Reader/archive source controls must remain
distinct from this new proposal's CI results. Existing native inference controls
are separate; GPU execution remains unverified.

The current three shortclaim models on Hugging Face are inactive failed
experiments. Model79 is a scratch linear feature-hash model, not a Laya finetune
or a 1.58-bit training result. The newly published 35-request corpus trained none
of them. These three are not the entire JooYoon model workspace.

Next: the Reader adapter; independent role/group and unknown-semantics review;
new semantic examples; protected 2,400-request evaluation; measured whole
verification-work improvement. Structural all-positive support does not revise
the existing Reader cohort policy. GitHub contains owned Apache-2.0 sources and
public diagnostic records. Future models require separate immutable Hugging Face
versions linking qualified data, source, license and CI evidence.

[한국어](README.ko.md)
