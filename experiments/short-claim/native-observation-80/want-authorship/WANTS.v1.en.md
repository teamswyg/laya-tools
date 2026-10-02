# Pre-dispatch Wants 80: UUID Parse, Scan and Ordinal

These are **pre-execution predictions** derived from frozen public source and inputs. There are no actual returns yet; every `Got` is null. The 24 fixtures are Parse8, UUID.Scan8 and Ordinal8: three behavior goals from two source families. They are not24 independent parents or new training labels.

The frozen source-first80 input is `INPUT-DRAFTS.v1.json`, 8,265bytes, SHA256 `a6110ab85d52fa4b8c62af6aab26bd13a988f04a0a5e38ce1e24e26542fa165b`. [WANTS.v1.json](WANTS.v1.json) preserves each original fixture object, JSON pointer, ID and order. The source-first author is semantic_review60_prep and the Want author is checkpoint_cli_45. They have separate authorship, but this is AI-assisted nonblind development with prior project exposure, not blind validation or independent human source origin.

| Path | Narrow prediction from source |
|---|---|
| Valid raw32 | `00112233445566778899aabbccddeeff`, nil error |
| Late `ge` failure in raw32 | `00112233445566778899aabbccddfe00`, `invalid UUID format` |
| Valid standard36 | Same successful16 bytes, nil error |
| Late `ge` failure in standard36 | `00112233445566778899aabbccdd0000`, same error |
| Mixed-case URN | EqualFold path then successful16 bytes |
| Wrong URN prefix | All-zero16 bytes, `uuid.URNPrefixError`, `invalid urn prefix: "bad:uuid:"` |
| Both38-byte `{…}` and `[…]` | Parse does not check the outer characters; both predict successful16 bytes |
| Scan nil, empty string, typed nil bytes, nonnil empty bytes | Preserve the fresh initial receiver `ffeeddccbbaa99887766554433221100` |
| Scan valid text and raw16 bytes | Replace receiver with successful16 bytes |
| Scan late invalid text and bytes | Preserve receiver, `Scan: invalid UUID format` |
| Ordinal 0,1,2,3,11,12,13,112 | `0th`, `1st`, `2nd`, `3rd`, `11th`, `12th`, `13th`, `112th` |

raw32 assigns its failing slot **before** testing validity. xvalues maps `g` to255 and `e` to14; byte arithmetic computes `(255<<4)|14 = 0xfe`, so that slot is fe and the following slot remains00. standard36 checks validity before assignment, leaving both slots00. These partial returns must not become a common 'all-zero on error' value. Parse's38-byte path differs from the separate Validate bracket check; Validate is not called in this observation.

Every Scan fixture receives a fresh nonnil16-byte UUID literal. No MustParse or FromBytes setup is allowed. Nil interface, typed nil []byte and nonnil empty []byte remain distinct input metadata rather than collapsing into one null or empty value. A local partial Parse return is not committed after an error. Non16 bytes use the source's recursive Scan(string(src)) path; this remains one direct observation input.

Error type means `%T` on the returned error, without invoking Error for type collection. The message should use exactly one direct `Error()` call per nonnil returned error. The Go1.27.1 errors.New and fmt.Errorf bodies were also read. The two UUID-format errors and two Scan errors predict `*errors.errorString`; the one URN-prefix error predicts `uuid.URNPrefixError`. Scan's `%v` is not a `%w` wrapping contract. Error identity, Unwrap and errors.Is remain null/out of scope. Ordinal has no error-return channel, so even `error_nil` is null. UUID, receiver and output channels not belonging to an entrypoint are also explicit nulls.

The direct entrypoint budget is24. Including five direct Error observations gives a proposed tracked dispatch total29: one upstream UUID Error method and four standard-library Error methods. The source-readable three internal Parse paths, one Scan recursion and fmt formatting are not dynamically instrumented. Four static UUID namespace MustParse→Parse initialization sites are separate from actual initializer return counts, which remain null/uninstrumented. The future native child must start after an external reservation. Whole-child OS measurement includes initialization, observations and persistence.

Source scope is google/uuid revision `2d3c2a9cc518326daf99a383f07c4d3c44317e4d`:15 Darwin non-JS runtime files plus original go.mod and complete BSD-3-Clause license; and dustin/go-humanize revision `a1b4e66b9a6d890e9e15e7091cf16c8032367d6e`: **the one original Ordinal source slice** plus go.mod and complete MIT license. The latter is not the complete humanize package and adds no Ordinal64. source_bindings preserves file SHA and license pins. Any public copy must retain original notices and complete licenses.

Proposed CPU1, Go soft heap256MiB, external60seconds and output1MiB are for a future execution plan. Soft heap is not an OS RSS hard cap. Actual original compilation/import/init/API/tests, adapter, observer, model/Fit, roles/labels, shared edits and publication are all0. [CHECKS.v1.json](CHECKS.v1.json) is an author-side metadata check, not an original function test or independent semantic validation. There is no broad RFC, universal no-panic, diversity, training-ready, production or protected-final approval. Preserve v1 and record any observed mismatch without rewriting the prediction.
