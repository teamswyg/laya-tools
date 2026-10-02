# Go role-partition module62

[한국어](README.ko.md) · [Go implementation](../../../internal/roleplan/role.go) · [Tests](../../../internal/roleplan/role_test.go)

This maintainer module partitions training, validation and calibration metadata. Related requests from shared code/helpers must stay together to avoid inflated validation results. It assigns each caller-frozen **whole connected component** as one unit. This is separate from runtime hint scoring or model routing.

The calling sequence is `VerifyAgainstFreeze` for frozen metadata, `VerifyMembershipSet` for declared members/relationships/digests/overlap, then `Assign` with that same component list and fixed seed/coverage plan. The actual17 original groups remain unassigned. Only synthetic controls were executed; package availability does not change `training_ready=false`.

`Assign` orders known-containing groups by SHA and allocates3:1:1, breaking largest-remainder ties train→validation→calibration. If the existing9/3/3 floors or predeclared coverage fail, **no assignments are returned.** Do not split groups or search another seed after observing results. Unknown-only components remain in training provenance, excluded from known-group counts; unknown members receive no labels. Zero transfer groups are permitted.

Arrays and contiguous slices replace repeated metadata lookup without adding maps or locks. Coverage IDs resolve once against a sorted component-ID index into contiguous original-index columns. Adjacent comparisons detect duplicates after sorting. No SIMD implementation is added for these string/metadata operations. Speed or RSS gains from the structural changes have not been measured.

Implementation limits are4096 components,256-byte seeds,512-byte IDs and65536 aggregate coverage references. Coverage IDs pass UTF-8/length checks before copies or sorts. Membership encoding preflights every length-prefixed byte and refuses sizes above64MiB before one buffer Grow. The complete membership set's encoded-byte sum also stays within64MiB. These are implementation bounds, not new scientific eligibility requirements or a64MiB whole-process RSS guarantee.

Independent review found that a reference-count cap alone did not sufficiently bound comparisons of long equal-prefix IDs. Moving byte/UTF-8 validation ahead of sorting resolved it. Empty/invalid-UTF-8/513-byte/1MiB references are refused while exact512-byte IDs are accepted. Preserve the [original finding](independent-findings-62.v1.json)/[ledger](independent-ledger-62.v1.json), [fix confirmation](independent-findings-62.v2.json)/[affected-check ledger](independent-ledger-62.v2.json), and separate [author port ledger](PORT-LEDGER-62.v2.json). `role-v1.go.txt` and `role-v1-test.go.txt` archive earlier code and are not build inputs.

Opaque digests and `HasKnownMembers` flags cannot establish complete relationships or truthful labels. Pointer/source/policy/full-graph binding and fixed seed/coverage remain the runner's responsibility. The membership [encoding specification](encoding-proposal-62.json) is still proposed. This establishes neither a new authoring origin nor blind evaluation. Errors are fixed strings, not an API promise of typed-sentinel identity. Actual roles, fits, weights and new labels remain0.
