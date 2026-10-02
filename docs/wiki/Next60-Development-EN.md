# Preparing the next small-model learning round

The new subset has **two requests and five labels**: two positive, three negative, each with unit weight. They qualify existing drafts; no new fit has run. The old 79-request corpus and inactive models are unchanged.

[Batch preparation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-source-batch-preparation) selects four targets, six short of ten, with thirteen held. Its 19 inputs, 12 descriptions and 57 proposed calls are preparation counts, not executions or newly qualified requests.

The [rounding controller](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-ftoa-outside-preparation) passed 28 Root synthetic race tests with one skip. Original/oracle/model calls remain 0. Its test parent's 121,012,224-byte RSS is not worker or GPU memory. Author seals and later Root test records are preserved separately.

A separate fit comparison follows checkpoints 30/60. Utility gates and protected 2,400-request targets per claimed domain remain unchanged; fixtures are not independent samples. [PR111](https://github.com/teamswyg/laya-tools/pull/111) passed all four required CI checks and merged automatically.

The [HF development dataset](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-2-finite-v1) is published under the fixed tag `next60-2-finite-v1`. All 38 downloaded files matched the original package and all 36 payload checksums passed. The HF viewer server responded HTTP500 with a busy/not-ready message, so actual viewer acceptance remains unverified. Direct file download is verified. See the [publication record](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/publication-proof-111/HF-PUBLICATION.v1.json).

The [Go reader](https://github.com/teamswyg/laya-tools/tree/main/pkg/shortclaimdata) separates a row into model text, supervision arrays and source bookkeeping. Owned fixed arrays and value copies avoid shared mutable state and locks. Local race tests passed 46 records and vet passed. These check data structure and information separation; they do not demonstrate the small hint model's utility or memory savings.

[Qualified subset](Native2-Training-EN) · [Actual original observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)
