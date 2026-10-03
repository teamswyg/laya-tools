# Independent review of the twenty-one-row materializer

We reviewed the sealed Go helper and tests, plus Root's saved generated rows. No concrete defect requiring a patch was found in this closed input/output scope. Review ran no Go, reader, original or model operation.

The helper copies truth already assigned by Root. Its `source_candidate_position==1` assertion checks the frozen artifact pattern; the output label itself comes from the explicit Root truth pointer. Unknown-only negatives are rejected. Known-counterexample negatives may retain unknown predicates. Positive candidates require every row satisfied and no unknowns.

The error-writer baseline is excluded from training. Its reference moves to selected position zero while source position one and metadata_id=reference-design remain in the adoption ledger. New candidate counts are 3/2/3/3/3. All five request/caption texts match both Wanted and adoption. Labels, IDs, role, unit weights, groups, upstream SHA and text revision match as well.

We independently checked the 22,097-byte previous prefix hash and final LF. Generated output is 28,800 bytes, SHA-256 `f8d9beb88393d2c1d2d8f690e366342157fd13bee25abb79078287e62c94b828`. Counts are 21 requests, 61 labels, 21 positive, 40 negative and 103 inputs. Selected candidate observations total 301; full original execution evidence including excluded candidates totals 305. The new five contribute 65 selected versus 69 original observations.

An always-select-position-one parent control is 18/21, approximately 85.7%. An always-negative candidate-label control is 40/61. They use different units and are not model accuracy. Position, candidate IDs and caption-style shortcuts require separate future controls.

Fixed arrays and typed slices suffice; this single-invocation helper needs no shared map or lock. All three complete input hashes are checked before composition, binding even Root metadata omitted by typed projections. Output uses a fresh exclusive path and preserves existing or partial files. Complete is set only after file and parent-directory Sync succeed. This is no power-loss or storage-hardware guarantee. The memory setting is a soft limit, not measured RSS/GPU/SIMD performance.

Saved authored synthetic logs contain four top-level plus sixteen passing subtests; we recounted them without reexecution. No float-specific synthetic test is claimed. Numeric metadata decode into int fields; no concrete numeric gap remains in the current fully hash-pinned contract.

The reviewer did not author the helper, Wanted, candidates or Root adoption. This is nonblind review with earlier source/data knowledge. The reviewer authored the five-task saved comparator and the twenty-one-row reader validator, so this is not independent review of their own validator. Root's actual reader execution is separate evidence. Earlier helper/QA pending statements remain historical. No generic license approval, general correctness, model improvement, cost savings or new CI success is established.

[Detailed review](REVIEW.v1.json)
