# Prepared feature contract scope

This is a research-only, private implementation inside an owned test source.
It adds no public/runtime API and changes none of the existing nine pinned
library files, learned weights, routing defaults or feature calculations.

The contract binds an exact raw query, candidate count and ordered raw texts.
IDs are metadata from the current call; a new View recomputes scores. Features
retain index/value bits, order, duplicates, zeros and the final overlap feature
at index zero. Its value varies by text; it is not a constant bias.
Caller-owned outputs can grow up to eight slots and successful reuse clears
their inactive capacity. Failed calls preserve output contents. Concurrent calls
require separate output storage; there is no runtime alias registry or guarantee.

Owned synthetic texts and deterministic numeric weights test these mechanics.
FP32, int8 and ternary PTQ are storage representations exercised in memory;
this does not train a ternary model, run Laya/transformer/GPU inference or read
Hugging Face assets. Tests, repetitions and candidate counts are not independent
Golden requests or semantic quality measurements. Exact API call totals are not
instrumented and must not be inferred as a complete attempt/panic ledger.

The private input layer permits empty/punctuation text to probe Features;
existing Rank/shortclaim validation is unchanged. Any future public integration
must apply its established validation separately. Bounds are at most eight
candidates, 512 bytes and 32 delimited words per text, and 64 bytes per nonempty
unique ID. These bounds do not replace the existing Rank limits.

The logical budget counts the owner struct once, cloned text lengths once and
tight feature backing capacities once. Headers/offsets already inside the owner
are diagnostic subparts, not additional bytes. Model/View, caller input/output,
allocator rounding and temporary feature construction are excluded. Exact
feature-sized budget rejection happens after temporary construction: this is
not a peak/heap/RSS/native/GPU limit. GOMEMLIMIT is a soft Go heap target.

No timing, SIMD/cache-hit attribution, semantic promotion, learning roles,
new labels, Fit, protected evaluation or default activation is part of this gate.
Previous failed recommendations and unresolved source lineage remain unchanged.
