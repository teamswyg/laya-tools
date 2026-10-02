# Independent review of explicit request scopes75

The v3 request scopes, source-satisfaction proposals and caption-specific BCE masks are supported by the inspected evidence. **One pair-learning explanation needs factual correction.** Keep existing v1/v2/v3 and historical training results unchanged; record a separate correction version.

The reviewer did not author request75/masks, but did author caption74. Independence covers the request/supervision proposal, not B-caption authorship or whole-pipeline blindness. This is AI-assisted source/text review. No original API, Normalize, Features, role or fit was executed.

Datasize explicitly requests a nonnil receiver, ASCII integers/binary byte units, receiver0 for syntax/bits errors and uint64 maximum for range errors. UnmarshalText:203–216 supports this. Parse operates on a fresh local; MustParse panics on errors, contradicting required receiver/returned-error behavior. The former 'distinct states' ambiguity remains historically intact while the new request narrows the scope.

Query's32words visibly include exported/nonignored/default primitives, custom exclusion and **included named nonnil non-time** structs. The source checks omitempty/isEmptyValue (:185–187,:321–343) before recursion (:264–268). 'Included' restricts the bracket assertion to fields that survive omission. Anonymous embedding flattening (:168–175), nil pointers, special time handling (:259–262) and custom callbacks (:194–205) are outside that nesting assertion. Default primitives follow index-order Add (:248–254), empty-slice omission (:216–220) and retained empty strings without omitempty (:271,:316). This supports conditional source satisfaction. It does not clear a broader standalone Values caption unconditionally: this is interpretation alongside the request, with historical caption74 fidelity/coverage states unchanged.

Shlex's request is unchanged. Split:403–415 accumulates words and returns the completed prefix with a later error, unlike a single-return Next. Successfully returned empty WordTokens are appended. Shared-helper facts can be read directly, without treating selected Next as a list-returning adapter.

| Sufficient-supervision proposal | A | B | Common |
| --- | ---: | ---: | ---: |
| Positive | 0 | 3 | 0 |
| Negative | 4 | 6 | 4 |
| Nonzero positive×negative pairs | 0 | 6 | 0 |

A's sufficient negatives are Parse, MustParse, Encoder and Lexer.Next: required shape/policy contradictions are visible. Other A captions omit material behavior or lack sufficient counterevidence and are conservatively masked. Omission was not substituted for source failure or automatic unknown-to-negative conversion. B's3P6N are bounded supervision **proposals**, not actual labels/weights. Arm-specific masks change caption information and supervision coverage together; they do not isolate caption effects. Common positive0 also prevents calling this slice a ready paired-learning dataset.

The correction is grounded in the actual second-fit d506 source. ranking.go:267 multiplies endpoint SampleWeights; :347–348 skips weight0 gradient, and :378–379 skips weight0 NLL. Saved driver fit.go:118 passes the same projected train/validation data, and main.go:373 forwards them unchanged to FitWithRankingTrace. fit.go:205–213 checks unchanged1/0 weights against LossEligible. No separate pair unmasking path exists. Therefore **BCE sampleweight0 makes that endpoint's pair training contribution0**. Physical audit retention of rows/pairs differs from loss/gradient contribution. Requiring a new implementation merely to block those contributions is inaccurate. A future policy distinct from BCE can be designed separately if desired. Existing72 values, weights and results remain unchanged.

One stdlib metadata-check attempt passed. It compared nine historical/current request budget rows, the same9candidate positions/two captions, preserved evidence26/source3/license references and all actual truth/role/weight/pair null fields. Word counts are v1=27/30/21, v2=27/32/21 and v3=27/32/21, within512B/32words. The original normalizer was not called. Only request+selected caption is a model input; code, metadata, observations and masks remain excluded. This preparation does not approve corpus inclusion, source diversity, training readiness or final2400per-domain qualification.
