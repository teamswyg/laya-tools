# Second tiny claim model: a useful failure before promotion

The small model suggests which code candidate to check first. Actual verification decides correctness, execution and approval. Existing Go ranking remains available; models and Codex integration are optional.

This trial added “put good candidates ahead within the same request” to candidate correctness learning. On the same validation requests, checks were **control27 → first model31 → second33**, and Top1 correct was **5→2→1**. Utility failed and default behavior does not activate either model.

Only16 requests contributed active ranking loss. Training loss decreased, but validation relationships were barely distinguished. Next comes development-data diversity across public code behavior and wording. Preserve the existing76 as regression data and keep separate2,400 final requests per domain out of model selection. Compression or repeated training has not been shown to solve this problem.

One CPU fit produced a32,792B coefficient file and observed whole-worker peak RSS about27.66MiB. The worker includes fitting, verification and serialization; these are not Laya-encoder, GPU-inference or Codex-token-savings measurements. Model bytes stay outside Git.

[Results and numeric review](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/second-ranking-fit-72/README.en.md) · [Why the next experiment changes](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/second-ranking-fit-72/ANALYSIS.en.md) · [Go input reuse](Validated-Input-71-EN) · [한국어](Second-Claim-Fit-72-KO)
