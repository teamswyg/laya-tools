# Tiny hint models: do better descriptions help?

[한국어](Claim-Diagnostic-76-KO) · [Getting started](Getting-Started-EN)

This diagnostic asks whether fuller descriptions of the same code candidates help a tiny model order independent checks. It fixes three requests and nine candidates, comparing original comment/signature descriptions A with source-behavior descriptions B. Both previously failed models stay unchanged. There is no retraining or default router change.

| Ordering method | Simulated checks with A | Simulated checks with B |
|---|---:|---:|
| First FP32 model71 | 3 | 8 |
| Model72 with an added ranking loss | 5 | 6 |
| Original order | 6 | 6 |
| BM25 | 8 | 3 |
| Words and adjacent word pairs | 8 | 3 |

Lower means the predeclared acceptable candidate appears earlier. These values are **sums of rank positions computed from a known rubric**. They are not counts of actual verifier, Codex or LLM executions. Top3 is always3/3 with three candidates, so it provides no useful discrimination.

On this small slice, B helps cheap retrieval and worsens the existing models. B also changes length and wording structure; this does not establish a causal effect of information alone. Before another fit, we plan additional public cases involving return versus panic, conditional state changes and negation: similar words with different behavior. This is not a protected final evaluation or completion of the2,400-request target for each domain.

One actual run completed36 feature/score calls across the two models. Whole-child maximum RSS was10.0625MiB and elapsed time2.23seconds. Startup, file checks, controls and saved checkpoints are included; these are not single-inference latency or demonstrated RAM savings against another experiment. There were no GPU or paid model calls.

In parallel source preparation, one actual Go observation of wordwrap and godotenv matched24 predeclared finite cases across two behavior goals. Those cases are not24 independent requests or new training labels.

Users can try the [small behavior-hint tool](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/USAGE-56.en.md) without a model download. It retains every candidate and proposes a verification order; actual code checks and execution remain separate. Failed models are not activated as defaults.

[Actual scores, review and licensing](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/caption-inference-76) · [Finite observations75](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/native-observation-75) · [Immutable first-model HF archive](https://huggingface.co/JooYoon/riidolaya-shortclaim-fp32-failed-71/tree/32b8f4579065247be0f71c83e1e7c143845e7b33) · [Immutable second-model HF archive](https://huggingface.co/JooYoon/riidolaya-shortclaim-rank-bce-failed-72/tree/d090b00e9a5dab00d5372dfd6412c9aee0c60b7b)
