# Independent materializer source review 125

Verdict: two required source corrections before execution. No further blocking semantic issue was found within the reviewed source and fixed Root-adoption trust boundary. This review does not certify compilation, Reader correspondence, durable materialization or successful exit.

1. source/cmd/main.go.txt:16 loses an observed Reader return when LoadDevelopmentRow returns an error. Its error branch returns Loaded{} with ReaderReturned=false. VerifyReader (library:245-249) consequently records one dispatch and zero returns even though the explicit project Reader call returned. Return Loaded{ReaderReturned:true} on that error path and retain zero additional-validation calls/matches.
2. source/library.go.txt:287-289 creates both exclusive empty files but syncs only directories before Load reads adoption/input files. Complete the requested durable reservation boundary by syncing train.jsonl and MATERIALIZATION.v1.json before the final output-directory sync and before Reserve returns. Propagate sync failures, close handles and preserve the fresh partial attempt without retry.

The separate patch note gives implementation and control requirements. Original author files were not changed.

Verified source seal: 12 files / 62,908 bytes, supplied HANDOFF and manifest pins, all nine payload entries and ten checksum entries. Previous33, prior metadata, Root qualification v2, saved-comparison content and the unchanged project Reader/input source pins match. Saved comparison was hashed only; its outcomes were not semantically re-evaluated.

Root qualification v2 covers two requests and six labels (+2/-4), all 33 observation ordinals, the frozen display maps, each positive's complete satisfied vector and each negative's known counterexample. All 36 unknown entries remain on selected candidates. Both requests preserve family83/development_train. No truth or role allocator is introduced.

The config digest and exact config shape precede fresh output reservation; actual input/adoption pins are checked afterward. Canonical absolute input paths, regular files, bounded stable reads and exact content hashes bind the inputs. Altered adoption/comparison data cannot bypass these fixed pins. Structural checks reject moved roles/groups, absent/null labels or weights, changed vectors/ordinals, positives with U and negatives without F. Missing channels are copied as U from the pinned saved judgment; this helper does not recompute or relabel them.

Composition copies all 45,390 previous bytes before appending, then checks the old prefix byte-for-byte. Existing33 source metadata independently totals 96 labels (+33/-63), 160 inputs, 467 selected observations and position histogram [4,28,1]. Source requires the resulting35 pool to be 102 labels (+35/-67), 171 inputs, 509 all original evidence observations / 500 selected observations and [5,28,2]. The two appended supervision objects copy Root labels, unit weights, captions, ordering and roles only.

All prior post23-scoped10 entries plus new36 are retained as scoped46; whole-history count is explicitly false. Excluded OriginalInt metadata, null label/weight, row vector and retained unknown fields are copied as raw JSON from the fixed prior receipt. Historical artifacts remain external preserved evidence.

The bridge calls unchanged LoadDevelopmentRow and separately ValidateInput. It compares raw request/candidate strings, candidate IDs, all supervision, all metadata and finite-scope fields, unused arrays, and complete normalized Prepared equality. Successful path requires 35 Reader attempts/returns/matches and 35 additional bridge validation calls; these do not erase Reader's own internal validation. Failure counters remain in the receipt, subject to correction 1. Successful receipt content alone has no completion authority: Root must require CLI exit0 and durable readback.

The fake sources cover F+U retention, U-only exclusion, all-known-T positives, field/order/weight/role changes, prefix and pin failures, metadata/unused tails, partial counters and exclusive existing-output rejection. They have not run. File/data/receipt caps, stable failure codes and no automatic retry are visible in source. Directory creation and O_EXCL prevent reuse of an existing output attempt; correction 2 completes reservation syncing.

Reviewer did not author this helper. Review occurred after author preparation and Root adoption, with exposed development evidence: nonblind development source review, not fresh blind or protected-evaluation proof. No Go, formatter, test, Reader, original/API/model/Fit or network operation; no shared repository or author-file writes. Only this fresh review package was written.
