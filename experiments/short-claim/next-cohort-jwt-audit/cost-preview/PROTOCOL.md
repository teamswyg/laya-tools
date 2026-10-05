# JWT operational cache preview: author-only protocol

This package is authored, unformatted, uncompiled and unexecuted. Root must retain
the complete source, supply the exact local JWT source, qualify the actual fixed
Go 1.27.1 Darwin arm64 selected closure/notices, build and pin the binary, then
freeze a separate one-shot run before execution. Authoring is not permission to
initialize JWT or execute any cache/API calls.

## Fixed evidence and finite controls

The input is exactly 13,474 bytes, SHA-256
`47686828bbb0cd4d273db3b2087311357ea853b767f86fd787953ed76d452ed6`.
It is the already prepared public RSA key and 13 fictional signed tokens used
by the first exhaustive observer. Runtime reads that exact input from bounded
stdin, accepts no arguments and performs no file/network/environment reads.
It generates no keys or signatures. All input bytes are already public fixtures;
the preparation private key is neither an input nor available to this worker.

The JWT source is revision
`73c870b18e68b6e654b2b03f485aa3c9fab32cea`. `go.mod` uses a local replacement
that Root must populate with the exact reviewed 25-file MIT source package.
The nominal required version `v5.0.0` does not describe the replacement revision.
JWT MIT notices and the fixed SDK license/patent/selected source notices must be
retained by Root. No source-family-wide or training-rights certification is made.

The whole JWT-go lineage retains its existing exposed `development_validation`
role. Four requests and 20 fixture positions represent 13 unique fixtures, not
160 independent tasks or a clean held-out set. No labels/rows are admitted and
no model, Fit, default behavior, GPU or paid API is involved.

Candidate positions remain the eight original source options:
leeway 2s, issued-at check, required expiration, required not-before,
any audience alpha/beta, all audiences alpha/beta, issuer-A, subject user-A.
The common requirements remain RS256, strict raw URL-base64 decoding,
RegisteredClaims and a fresh fixed-time callback at Unix 1700000000.

Fixed source order is `[0,1,2,3,4,5,6,7]`. The manual-first choices were frozen
before the original API observations: `[4,5,2,0]` for the four parents,
followed by all remaining candidates in source order. These are exposed manually
authored semantic controls. They are not model predictions, data fitting or
evidence of generalization. BM25 and Jaccard are absent and require separate
source/config freezes before any future operational evaluation.

The literal Wanted vectors, in the fixed five-fixture order, are:

| Parent | Wanted |
| --- | --- |
| lantern | true, true, true, false, false |
| harbor | false, false, true, false, false |
| moss | false, true, false, false, true |
| cove | true, false, true, false, true |

These literals encode the predeclared request contracts. Expiration is strict
at its boundary; not-before is inclusive. The contracts are not used to obtain
an authentication result or to populate the cache. They determine fail-fast
fixture mismatch and candidate completion after an observed fresh policy result.

## Two scoped paths

The uncached reference constructs a fresh parser with all common options and
one candidate option and invokes the original ParseWithClaims on a fresh zero
RegisteredClaims plus a fresh public-key copy. Its measured parse duration
includes JSON decoding, signature verification and policy validation. Those
internal phases cannot be split without instrumentation or changing the source;
separate uncached authentication/policy times are therefore JSON null.

The cached path uses a bounded five-entry array reset for every parent/control
and round. It performs lazy authentication/strict parsing through the exact
source ParseWithClaims with WithoutClaimsValidation. The full exact key includes
the token string, complete public-key DER bytes, RS256, decoder/claims identity
and source revision. The cache does not include a candidate, Wanted value,
option-validation result, final Token.Valid result, expiration decision or time
policy result. An authentication error is unknown and is never a cached refusal.

After authentication, the cache holds decoded RegisteredClaims only. It copies
the audience slice and all three NumericDate pointers into owned values when
storing, then makes another deep copy for every fresh candidate validation.
Absent/nil dates stay nil. A fresh NewValidator receives the common fixed-time
option and that candidate option on every trial. ValidMethods and StrictDecoding
are authentication-parser requirements; applying them to NewValidator does not
replace the preceding authentication/decoder checks.

Policy errors receive an explicit ErrTokenInvalidClaims wrapper before bounded
sentinel normalization, matching the uncached parser's generic wrapper plus
policy causes. Error messages/tree shapes are not equated. Only exact recognized
policy leaves produce observed false; unknown crypto/input/custom leaves,
contradictions, capacity/deadline or unwrap failures remain unknown. The walker
is limited to 32 nodes/depth 8 and catches panic; it does not interrupt a hostile
Unwrap implementation. The pinned source has bounded standard wrappers and the
outer process deadline remains necessary.

Each trial hashes canonical claims and checks the immutable cache/public-key
values. These diagnostic checks are charged and add overhead; this is not a
minimal production cache implementation. Cache values contain no Token or
approval Boolean. This shortcut is scoped to these exact fixed authenticated
tokens, typed RegisteredClaims and policy-only options. It is not a general
JWT application cache: key rotation, revocation, custom claims validators,
decoder/global changes, token/key/source/method changes, time changes or other
parser options require invalidation/requalification and fresh policy validation.

## Order and measurement

First, all 160 parent/fixture/candidate combinations run in original
parent-then-fixture-then-candidate order. Each combination records the uncached
and cached result, including known/accept state, normalized error mask and
canonical claims SHA. Unknown results or any parity mismatch prevent the
operational comparison. This phase itself performs native API calls and must
be counted separately. Its interleaved per-parent complete path durations are
not separately attributable and are JSON null; additive measured phases and
all constructor/callback/API counters are retained.

Only after exhaustive equivalence succeeds do three fixed rounds run both
source/manual orders with both uncached/cached paths. In rounds 0 and 2 controls
are fixed then manual, and modes uncached then cached. Round 1 reverses control
and mode order. The within-control candidate order and within-candidate fixture
order never change. A candidate stops at the first observed Wanted mismatch;
a parent stops at the first candidate matching all five fixtures. Unknown
results halt with incomplete status instead of becoming negative labels.
There are 48 planned parent work records. Every parent/control gets a new cache.

There are no explicit warmup rounds, but exhaustive equivalence has already
warmed the process before operational rounds. Three deterministic rounds are
not independent replications or a randomized causal speed study. Absolute
timings include timer/count/hash/copy overhead. No timings from the earlier
160-call observer are summed or presented as a cache speedup.

Counters distinguish direct NewParser, NewValidator, its internal NewParser,
ParseWithClaims, Validate, option construction/application, key/time callbacks,
cache lookup/comparison/hit/miss, claims/key copies and fixture/candidate checks.
RSA verification is inside the original parser; its internal call count is
unobserved rather than claimed as an independently instrumented count.
Measured setup, lookup, copy, cache store, order construction, combined uncached
parse, cached auth/parse, fresh policy and outcome qualification costs are
reported. Complete operational parent work includes initialization and residual
overhead. Input/key setup and total main wall are also reported; output marshal,
write and imported package initialization are outside total main wall.

## Bounded execution and retention prerequisites

Runtime asserts Go 1.27.1 Darwin arm64 and default one-second JWT TimePrecision.
Main sets GOMAXPROCS=1 and a soft 64 MiB Go memory target. This is not a hard RSS
limit. The planned main budget is five seconds checked between finite calls;
it does not preempt a single crypto call. Root must use the reviewed one-shot
runner with an outer 30-second child wall deadline, fixed environment, no retry,
no warmup execution, stdout 192 KiB, stderr 4 KiB and separate first raw capture.
Whole-child RSS/CPU/wall include startup and come from the runner, not this code.
The public result contains no local source/host paths or private-key material.

Root must archive the first complete raw streams and parent receipt before
interpreting parity, candidate success or performance. A truncated/nonzero/
unknown response is not success and must never trigger an unrecorded rerun.
The native outcome and performance counters in the author receipts are all zero.

The first uncorrected authored source is preserved in FIRST-AUTHOR-FULL. Static
corrections before freeze fixed an unnamed-return timing assignment and added
generic invalid-claims wrapper parity and complete cache-store/key ownership
accounting. They were source review findings, not discovered by execution.
