# First real comparison of four requests

One execution compared original and alternative candidates for four distinct requests: 19 finite inputs, 12 candidate positions and 57 saved observations. These are not 57 independent requests or measured internal original-API calls.

| Request | Original candidate 0 | Alternative candidate 1 | Other candidate 2 |
| --- | ---: | ---: | ---: |
| Comma grammar | 3/6 | 6/6 | 3/6 |
| Key collision | 2/4 | 4/4 | 2/4 |
| Map snapshot | 2/4 | 4/4 | 2/4 |
| Fractional bytes | 3/5 | 5/5 | 4/5 |

Counts indicate satisfaction of the full frozen expectations. A colleague who did not author the implementations or expectations compared the saved observations: 40 satisfied, 17 known mismatches and zero unknown comparisons. Candidate and error-method panics were zero. Candidate 1 satisfies each request; eight other positions have counterexamples. The review is nonblind because the reviewer inspected source and expectations beforehand.

Whole-worker maximum RSS was **8,273,920B, about 7.89MiB**. Start, input checks, observations and durable writes took **1.077920459 seconds**; OS user/system times were 0.01/0.04 seconds. These include per-event synchronized writes and are not model inference speed, Go heap or GPU memory measurements. Saved counts are 142 synchronized events, 57 candidate returns and 14 error-method returns. Internal original API and startup counts remain uninstrumented `null`.

The [observations](OBSERVATIONS.v1.json), [outside result](OUTSIDE-RESULT.v1.json) and [independent comparison](independent-review/FINITE-COMPARISON.v2.json) are preserved. The large intermediate journal is referenced by digest; binaries and private host paths are excluded. The initial metadata-filter failure is preserved without replaying the originals.

Root admitted four requests for finite development training after semantic review and [input validation](../next60-development-seven/INPUT-VALIDATION.v1.json). The [current data](../next60-development-seven/README.en.md) contain seven requests and 20 labels. This does not establish all-input or unseen-source correctness, model improvement or cost savings.
