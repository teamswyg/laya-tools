# Separate visible description information from code satisfaction

The model receives a request and a short caption. Missing conditions in a caption do not make the underlying function incorrect. This preparation keeps the same three requests and nine code candidates, comparing original captions A with behavioral captions B.

Use [v4](preparation/BOUNDED-REQUEST-SUPERVISION.v4.json) and [final handoff v2](preparation/FINAL-HANDOFF.v2.json). The [peer review](review-v4/REVIEW.v1.en.md) supports three satisfying and six incompatible direct selected-API proposals under explicit request domains. Existing training truth, roles, weights and groups remain null. Requests contain27/32/21 words; the512-byte/32-word limits are unchanged.

A [separate scoped diagnostic rubric](../caption-inference-76/rubric/riido-caption-inference76-root-truth.en.md) evaluates all nine candidates in both arms, retaining information-poor A candidates. It does not qualify training labels, independent final evaluation or universal behavioral truth. Caption arms, translations and repeated calls remain the same three logical requests.

Versions1–3 and their original notes are preserved. Preparation v3 incorrectly said zero-weight rows still contribute to pair loss. The [correction](preparation/FACTUAL-CORRECTION-75.en.md) records that actual72 uses the product of endpoint weights, making the gradient and NLL contribution zero when either endpoint is zero. Audit retention and learning contribution are different. Historical code, fit results and weights were neither changed nor rerun.

This three-request slice has no common A/B eligible positive pair, so it cannot establish a paired training comparison. A fixed-model caption diagnostic remains possible. This does not claim that fitting the entire historical corpus is mathematically impossible.

[한국어](README.ko.md)
