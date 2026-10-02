# Correcting the pair-weight explanation in preparation 75

**In current72, a zero BCE sampleweight also gives zero learning contribution to every pair using that endpoint.** The earlier v3 preparation and KO/EN notes conflated endpoints/pairs retained in the audit with actual rank-gradient/NLL contribution. Original v1/v2/v3, documents, training specifications and results remain intact. This separate v4 and addendum are the current explanation; historical numbers are not changed or invalidated.

Reading the pinned implementation and actual driver establishes:

- [ranking.go 265–272](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L265): pair weight is the product of positive/negative endpoint `SampleWeights`. Zero-weight pairs remain in audit arrays.
- [344–360](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L344) skips zero-weight gradients. [372–390](https://github.com/teamswyg/laya-tools/blob/d506f58ccf9629e2d2b6ca3cd1766ab20789bf64/internal/pairlearn/ranking.go#L372) also skips zero-weight pair NLL contributions.
- Saved driver `fit.go` line 118 forwards projection Development/Validation unchanged. Lines 205–213 verify original 0/1 sampleweights against LossEligible. `main.go` lines 371–380 passes the same datasets to `FitWithRankingTrace`. It does not separately unmask pairs.

Exact file hashes, byte/line ranges and span hashes are in v4's `factual_correction_of_preparation_v3.source_evidence`. The actual frozen plan SHA is `17afbf5c8d14fdd34f7c627d7b43b9582793984d010f52118f408b287bbc7daf`. The peer receipt SHA is `c9c5df587cf8b9ac287692b1ac843ddd0fd6489dbbf23fcd4e6be088f0381f6e`. That reader is independent of this request/mask author, but authored caption74 and is nonblind. We did not recalculate pair/training outcomes. The stored audit's 16 training and 2 validation zero-weight pairs are referenced from the parent/peer's existing-result check.

BCE eligibility and pair endpoint eligibility can be recorded as separate concepts. However, **no new runtime implementation is needed to honor current72's same zero-weight exclusion.** A separate endpoint policy is a future objective-design hypothesis. Remaining v4 fields such as `enforcement_implemented: false` concern an unimplemented standalone75 endpoint-flag policy, not absence of the existing sampleweight-product behavior.

Scoped requests and source-code satisfaction proposals are unchanged. A has 4 sufficient negatives and 0 sufficient positives; B has 3 positives and 6 negatives. Applying the proposed 0/1 sampleweights hypothetically through current72 gives **nonzero pair proposals** of A 0, B 6 and common intersection 0. No actual pair generation or training occurred. The common intersection still lacks positives, so this slice alone does not support a trainable paired-comparison claim.

All 3 parents, 9 candidates, order, original A/B text, actual truth/role/weight/pair `null` fields and evaluation denominators stay intact. v4 is a factual correction and metadata handoff, with 0 original API, Features, Project, Fit, model, role, shared or remote mutations. It creates no new training qualification or corpus inclusion.
