# Preparing the next small claim model

The new development round now has **three requests and eight candidate labels**: three positive, five negative, each with unit weight. Three existing draft IDs have been finitely qualified. No new fit has run; the actual 79-request corpus and previous inactive models remain unchanged.

[Direct-rounding observation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-ftoa-actual-observation) completed one real execution. Six inputs × three candidates are 18 observations and one distinct request. The precise candidate satisfies all six inputs; witnessed counterexamples support negative labels for the other candidates. Unavailable original error channels remain unknown. This request joins the existing humanize Ordinal training group 76, adding no independent source family or held-out evaluation.

Whole-child wall time was about 0.378 seconds and Darwin reported direct-child maximum RSS of about 5.55 MiB after exit. The 207,056-byte Go heap value is a snapshot. These describe the original-code validation tool, not small-model inference, GPU/whole-machine memory or demonstrated cost savings.

See the [three-row subset and guide](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-development-three). The Go reader separates text, fixed label/weight arrays and metadata. Only request/candidate text enters features. Three reader loads and one legacy input check passed; CI adds the same correspondence check. [PR112](https://github.com/teamswyg/laya-tools/pull/112) passed all four required checks and automatically merged the reader.

The [previous immutable two-request HF release](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-2-finite-v1) passed a pinned download of all 38 files. Its first viewer response was HTTP500 and is preserved. A later HTTP200 response returned two rows matching the pinned originals. A subsequent three-row HF version will be linked after separate CI and publication evidence.

The [next batch preparation](https://github.com/teamswyg/laya-tools/tree/main/experiments/short-claim/next60-source-batch-preparation) has four requests, 19 inputs, 12 candidates and 57 proposed observations. These are not completed execution or qualification counts. Even qualifying all 20 drafts would leave ten distinct semantic requests needed for the new-round checkpoint of 30. Further requests will be selected from the existing public contract catalog; paraphrases do not become new requests.

A separate fit comparison follows checkpoints of 30 and 60. Protected evaluation still targets 2,400 distinct requests per claimed domain. The 5% utility-improvement gate remains unchanged; a model stays inactive until actual comparison passes.

[Go reader guide](https://github.com/teamswyg/laya-tools/tree/main/pkg/shortclaimdata) · [First two training requests](Native2-Training-EN) · [First original observations](Native2-Observation-EN) · [한국어](Next60-Development-KO)
