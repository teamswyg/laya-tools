# Read-only review of the 69 existing-input boundary result

All 72 requests and 216 captions in original probes56/56b satisfy the existing decoded-input contract. Separate stdlib math recomputed raw bytes, normalized bytes, and normalized word counts for all 288 texts and matches every actual 69 result row. The totals agree on 72 supported, zero unsupported, zero out-of-bounds captions, and maximum request/caption counts of 32 words. The earlier static length suspicion did not materialize in the actual result or this review. Shortening/deleting text, changing bounds, or changing labels is unnecessary.

| Existing text-bound check | Request | Caption |
| --- | ---: | ---: |
| Text count | 72 | 216 |
| Maximum raw bytes | 236 | 252 |
| Maximum normalized bytes | 225 | 239 |
| Maximum normalized words | 32 | 32 |
| Texts at exactly 32 words | 8 | 24 |

The existing rules allow at most 512 raw bytes, at most 512 normalized bytes, and 1–32 normalized words. Exactly 32 is allowed; 33 is rejected. The pinned splitter lowercases Unicode letters/digits and turns other characters into token boundaries. It splits uppercase characters after lowercase/digits and at an uppercase-acronym boundary followed by lowercase. Token order and repetition remain; normalization does not remove keywords, deduplicate a word set, or truncate sentences. Actual input.go and lexicalhint/features.go source hashes bind this reading. A separate streaming token assembly reproduced those conditions without computing lexical features or rankings.

Every parent/candidate ID, order, candidate count, and original-input SHA matches. The existing decoded Validate contract's 1–8 candidates, 1–64-byte flat ASCII candidate IDs, duplicate-ID rejection, UTF8 checks, and fixed schema/provenance are also satisfied by current original inputs. All supported statuses and empty fixed errors match these conditions. This is a decoded Validate contract review; it is not a new experiment covering Load's separate 12KiB JSON wire limit, every arbitrary input, or forged Prepared values.

Exact pins agree for the 1,677-byte frozen plan, 88,420-byte actual result, 421-byte invocation, three source files, and two input files. One root audit process records 72 Validate calls and 288 explicit metric NormalizeText calls. Validate's internal normalization calls were not individually instrumented and are not added to a fabricated total. Project/Features, original behavior/role APIs, fits, models, paid trials, and protected-final reads remain zero. This review reran no original Validate/NormalizeText/Project/Features or actual role/source/fit functions.

Raw-text SHA witnesses in the new `MECHANICS-BOUNDARY-69.v1.json` hash original strings. Normalized-text SHA witnesses are from this separate recipe implementation, not digests output or recorded by the original function. Root69 records lengths, not normalized strings/digests. The source-recipe/length/status comparison and these separate reconstructions remain distinct. New public-candidate artifacts contain no raw text, normalized text, or feature vectors. Earlier 66/68 receipts and original results remain unchanged.

Contract support means runtime input-bound acceptance. It does not approve source/caption fidelity, supervision eligibility, loss masks, effective labeled groups, diversity, training, or practical hint quality. Existing 72 parents, 216 candidates, unknowns, labels, masks, groups, and roles remain unchanged. No new cap, scope, or numeric gate was introduced. Current original texts are accepted; future different texts still require the same existing checks.

This reviewer did not author the 69 root helper and has nonblind exposure through earlier 62/67 implementation and 66/68 reviews. The review does not prove source origin, human-only authorship, compiler/binary trust, or blinded final quality. One stdlib-only metadata helper passed, reading eight pinned files totaling 205,177 bytes. One uncached race invocation covering three synthetic tests and one vet invocation passed. Tests cover camel/acronym/digit/Unicode splitting, repeated tokens, exact 32/512 boundaries, Unicode normalization byte expansion, empty/invalid-UTF8 text, and flat IDs. Helper/test/vet and required-pin failures are zero.

One exploratory command guessed incorrect source locations and failed; the failure is retained. Reading actual pinned pkg/shortclaim/input.go and internal/lexicalhint/features.go recovered it without claiming original sources were absent. Shared edits, Git, publication, actual training, and additional original API execution remain zero. Ordinary AI-assisted collaboration cost is unmeasured. This review ran no resource benchmark and did not create root69 performance measurements.
