# Scoped evaluation rubric for a fixed-model diagnostic

This freezes the request75 v4 source-satisfaction proposals and peer review as a rubric for one development diagnostic. The original proposal's null truth, roles, weights and groups remain unchanged. This does not admit labels into a training corpus or protected final set.

`riido-caption-inference76-root-truth.v1.json` is 22,615 bytes, SHA-256 `1cec5c108d09e78f3dd6a3465538bf9589df6d2f981edb4b12b736321aa648d6`. Requests, selected code, candidate order and scoped source clauses are fixed. Under the explicit input domain and direct selected-API requirement, candidate0 for datasize, candidate1 for query and candidate2 for shlex satisfy the rubric. The other six have incompatible direct return shapes or policies. No unseen adapter or implementation is assumed.

Arms A and B describe the same code. All nine candidates and the same rubric remain in both arms, including candidates with insufficient information in A. Features receive only the English request and selected caption. Truth, code, identifiers, evidence and candidate position are excluded from features.

The rubric supports an exploratory ordering comparison of two fixed models. It is not universal behavioral truth, independently human-authored data, unseen-source performance, or a protected 2,400-request evaluation. The root author is an AI-assisted, nonblind reviewer exposed to prior results and source reviews. Translations, caption arms and model scores do not increase the three logical parent requests.

Root checked the3×3 structure, satisfaction booleans against acceptable indices `[0] / [1] / [2]`, false qualification/training/production/final flags and reference hashes. `qualification:false` was added after initial projection and before sealing the final hash. Preparation performed zero Decode, feature, score, fit or upstream API calls. Actual inference requires a separate execution record.

[한국어](riido-caption-inference76-root-truth.ko.md)
