# Safe reuse of prepared features

[한국어](README.ko.md) · [Scope](SCOPE.md)

The previous [actual cost pilot](../prepared-storage-preview/README.en.md)
reduced computation and allocation when one unchanged request was reused.
This research gate checks stale inputs and results before considering product
integration. It adds no public API or default routing policy.

Reuse requires the exact raw query, candidate count and ordered texts. ID-only
changes return current IDs. A different View recomputes scores. Smaller result
sets clear inactive output capacity, and failed calls preserve previous contents.
Concurrent calls require separate output buffers; runtime alias protection is
not provided.

Run from a checkout with Go1.27.1 and a C compiler for the race detector:

```sh
bash experiments/short-claim/next-cohort-xxhash-audit/prepared-contract-preview/verify.sh
```

The launcher verifies pinned bytes, then runs race tests and vet in an owned
temporary Go module. It requires no network, Python or downloaded weights and
cleans temporary files. Deterministic FP32, int8 and ternary PTQ numeric weights
are created in memory and scored by actual Go code. This is neither learned
Laya inference nor ternary training.

Controls cover 1/5/8 candidates, Unicode/repeated words/punctuation/empty text,
input bounds, exact feature and score bits, ties, invalidation, current IDs,
changed models, output shrink/growth, logical budgets, duplicate features and
concurrent reads. The final index-zero feature is word overlap, not a constant
bias. Synthetic fixtures are not independent Golden requests or quality scores.

The private test input permits empty text; existing product validation remains
unchanged and needs a separate integration gate for any future public API.
Storage arithmetic counts the owner once, cloned text lengths and feature
backing capacity once. Some budget checks follow temporary feature construction,
so this is not a peak-memory bound. Model, caller buffers, temporary allocation,
allocator rounding, RSS and GPU memory are outside that arithmetic.

The earlier failed recommendation and unresolved lineage remain. These controls
establish no semantic quality, SIMD/cache-hit attribution or whole-task/Codex
usage savings. New training roles, labels, fitting and protected evaluation are zero.

The first actual run passed all seven race tests, eight named input subcases and vet. [Execution](EXECUTION.actual.public.v1.json) · [Readback](READBACK.actual.public.v1.json) · [Complete output](STDOUT.actual.public.txt). Whole-repository race checks also passed, with all existing test results cached. The 2.292s compilation-inclusive duration is not inference performance.
