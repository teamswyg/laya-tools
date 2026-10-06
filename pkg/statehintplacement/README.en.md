# Shadow placement bound to opaque owner revisions

This Go library conditionally checks where progress-report, completion-report and question proposals belong in already-held content. It implements no actual HTTP/native reader, authentication, reaction, notification or state write. Test readers and scores are explicitly synthetic doubles.

`Anchor` separates opaque scope/Work/Content references, an **owner-boundary-provided work revision**, and a content SHA-256. Missing `OwnerContentRevision` stays empty. `TextDigest` hashes exact valid UTF-8 bytes up to4,096bytes without trimming, lowercasing or normalization. A derived digest proves no owner-issued revision, permission, admission or atomicity. Do not mint/replace revisions from timestamps, row numbers, hashes or extracted numeric components.

`RevisionReader.ReadAnchor(ctx, scope, workRef, contentRef)` receives neither expected revision/hash nor model-input text. Its current SHA must be derived from current content actually held by the trusted application boundary. If inaccessible or unknown, omit the digest and return unavailable. Do not reflect user-JSON or model-input hashes as current-content evidence. The actual native adapter enforcing this boundary is not implemented.

`ValidatePlacement(expected, current)` exactly compares scope, Work/Content references, opaque work revision, supplied content revision and SHA. Synthetic composite revisions such as `p12-w9` remain whole strings, without numeric extraction. Missing/invalid required metadata is unavailable; changed metadata is stale. Equality still means only a **conditional `placement_matched` observation**, with `actual_verified:false`.

`Adapter{Revisions, Catalog, Predictor}.Propose(ctx, Request)` verifies already-held input bytes and reads a current anchor, then validates the original catalog schema. Annotation bindings for the same scope/Work are limited to the three display intents with state planning disabled. Confidence0.9and margin0.05remain fixed; eight probabilities are not renormalized. A final anchor read discards proposals when changed or unknown. Owner opaque revisions are never inserted into the catalog's older numeric-version field.

All results remain shadow, with false mutation/state-change flags and unqualified catalog consistency. Matching separate reads does not prove one atomic snapshot. A future application must revalidate current catalog and permission before real placement. Native IDs/names/credentials and lookups stay in the private adapter; public code receives only caller opaque references. Results do not echo source text, work revisions or content digests.

Replays read current metadata again, with no private shared cache or lock. Existing matching labels/emoji produce an annotation no-op. Classifier cost differs from actual native metadata/read/UI costs; actual verification of the latter remains zero.
