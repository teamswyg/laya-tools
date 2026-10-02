# Preparing the next small-model learning round

The new subset has **two requests and five labels**: two positive, three negative, each with unit weight. They qualify existing drafts; no new fit has run. The old 79-request corpus and inactive models are unchanged.

[Batch preparation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-source-batch-preparation) selects four targets, six short of ten, with thirteen held. Its 19 inputs, 12 descriptions and 57 proposed calls are preparation counts, not executions or newly qualified requests.

The [rounding controller](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-ftoa-outside-preparation) passed 28 Root synthetic race tests with one skip. Original/oracle/model calls remain 0. Its test parent's 121,012,224-byte RSS is not worker or GPU memory. Author seals and later Root test records are preserved separately.

A separate fit comparison follows checkpoints 30/60. Utility gates and protected 2,400-request targets per claimed domain remain unchanged; fixtures are not independent samples. PR111 CI and the [planned HF dataset](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development) release are pending; the link does not claim publication.

[Qualified subset](Native2-Training-EN) · [Actual original observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)
