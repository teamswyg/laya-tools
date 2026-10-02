# Conversion80 saved-result comparison

All 72 stored candidate observations match the 72 Wants sealed before execution. One Go native run observed the existing 3 requests and 9 code candidates against 24 original inputs. This was not Laya, GPU inference, new model training, or 72 independent requests.

The checker did not rerun original code or the observer. It compared every full literal, all 11 return channels, nil versus empty inputs, candidate/input order, Scan wrapper argument receiver, and unobserved panic text. Final, partial and stdout are byte-exact at SHA `ee6b993dd5495e901ea9b516f96278eee76821bdb1e41ffaaac0ba0f2433ab8c`.

Of 72 candidate dispatches, 69 returned normally and 3 expected ParsePanicOnError panics were recorded. Explicit Error observation on returned errors added 10 calls: 2 URN and 8 standard errorString calls. Thus 82 callbacks were tracked. The 56 inner original API mappings and 4 startup sites remain static evidence; their dynamically measured counts remain null. Error-valued panics incurred no additional Error/String/Format call.

The checker compared 37 closed file pins, including 20 originals, with the result. Only the frozen false-to-true token differs between draft and frozen plan. The initial reservation retained static56, zero execution counters and null inner/init counts. Root started once, retried zero times, exited zero and joined Wait, with no timeout or overflow.

The stored whole-child OS RSS was 19,529,728 B (18.625 MiB), and footprint was 16,728,544 B. OS real/user/system were 2.84/0.34/0.18 seconds; outside-controller wall was 2.8445567499999997 seconds. These include startup, loader, sync writes and candidates, rather than pure function cost. The 256 MiB Go heap setting is soft and does not establish a hard total-RSS cap.

One stdlib metadata checker attempt passed; zero failed. The receipt's 6,571 completed checks count decoding, shape and pin comparisons, rather than new semantic answers. The terminal count of 6,575 also includes final output fsync checks. This reviewer authored the observer/controller and saw the outcome beforehand. The review is nonblind saved-file mechanics QA, not an independent source oracle, runtime actor, blind semantic review or training approval. The original Wanted truth/role/weight/mask/label/Got nulls and preparation flags remain unchanged. Root's later SAT/eligibility judgments are separate records.

[Comparison](qa/NUMERIC.v1.json) · [Receipt](qa/RECEIPT.v1.json) · [Attempt ledger](qa/LEDGER.v1.json)

This folder is a public-preparation stage, with no shared Git changes, external publication or original code rerun. It preserves the [actual result](actual/results.json), [sealed Wants](want/WANTS.v1.json), [observer handoff](observer/HANDOFF.v2.en.md), and [original skeleton](skeleton/SKELETON.v1.en.md). Pending handoff status, reservation zero counters and preflight execution0 are historical pre-execution evidence. The [Root actual ledger](actual/ROOT-ACTUAL-LEDGER.v1.json) separately records completion and preserves its reused generic76 schema. The present scope is conversion80.

The [COPY ledger](COPY-LEDGER.v1.json) separates byte-exact copies from derived Markdown with link remapping only. Exact Markdown bytes also remain in `.md.txt`. The [MANIFEST](MANIFEST.v1.json), [SHA256SUMS](SHA256SUMS), and [omissions](OMISSIONS.v1.json) track original/public hashes and exclusion reasons. PRIVATE plans, binaries, helpers and raw OS/test logs are SHA-only. Go/modules are inert `.go.txt`/`.mod.txt`; full MIT/BSD/Apache notices and source slice scope are listed in [NOTICE](NOTICE.md). [한국어](README.ko.md)

Root's later SAT, caption fidelity, eligibility and corpus append records are separate; this observer snapshot is not retroactively changed. There is no claim of 128/256 independent development problems or 2,400 protected final requests per domain. One private helper authoring syntax failure is preserved in the [original ledger](AUTHORING-FAILURE.v1.json); it was not a native execution failure.

The exact-copy helper stopped once at a metadata check of a not-yet-created MANIFEST link after copying every original payload. The [failure ledger](COPY-FAILURE.v1.json) and [separate finalization ledger](COPY-FINALIZATION.v2.json) preserve this. The old COPY ledger failure0 is a historical pre-failure snapshot; the new ledger records the final status. No original code, observer or saved Go QA was rerun.
