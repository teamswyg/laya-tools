# Seven development requests for the next round

These **seven requests and 20 candidate labels (seven positive, 13 negative)** were admitted after actual finite observations and separate comparison. Four requests extend the first three rows, which remain byte-identical. The 34 inputs and 98 observations are separate from the request count.

See the [data](data/train.jsonl) and [Root qualification record](ROOT-QUALIFICATION.v1.json). New request text, candidate descriptions and order match the frozen preparation. Existing connected source groups 76/77/78 are reused for development training. They are not new independent sources or final evaluation data.

[Go validation](INPUT-VALIDATION.v1.json) checks seven strict reader loads, four frozen-input correspondences, unit weights and the exact three-row prefix. Features, scores, projection, additional Fit and original replay are zero. The validator does not grant semantic truth or qualification.

Run `bash scripts/verify-next60-development.sh data` from the repository with Go 1.27.1. This check executes neither a model nor Python. Consult the PR checks for the actual CI result.

The current pinned public HF version is the [three-request tag](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-3-finite-v1). The seven-row release is planned after CI, under a new tag without moving previous tags. Model bodies are excluded from GitHub.

The historical 79 requests, inactive failed models and three cumulative Fits remain unchanged. No new corpus Fit runs before the 30/60 checkpoints. [Literal corrections for ten further requests](../next60-catalog10-literal-correction/TRANSITION.en.md) remain unobserved and unlabelled; six inputs requiring report-channel instrumentation remain on hold.
