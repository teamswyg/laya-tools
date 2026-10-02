# Connecting small claim hints to training data66–69

[한국어](https://github.com/teamswyg/laya-tools/wiki/Claim-Checks-66-69-KO) · [Previous preparation64–67](https://github.com/teamswyg/laya-tools/wiki/Claim-Preparation-64-67-EN)

The target is a very small model that reads a short request and candidate captions and **suggests what a person or agent should verify first**. For example, a task asking for nested paths can receive a promising pattern earlier. Actual code checks establish the final answer. Tasks without a suitable candidate or supported input continue through the existing workflow.

This stage prepares exact Go-array bindings for original text, truth, connected groups and loss masks. Actual role assignment and public-function observations each ran once. There is no new fit or demonstrated Codex saving at this stage.

| Completed work | Observation | Meaning for the next training step |
|---|---|---|
| [Role assignment66](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/role-execution-66/README.en.md) | Labeled groups10/3/3 in train/validation/calibration | Keep each connected code family in one role. |
| [Upstream captions68](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/source-caption-68/README.en.md) | Four English requests,10 candidates,2 code families;12 fixed entrypoints and3 String observations, all15 comparisons match | Bind pre-existing MIT wording to finite semantic targets. |
| [Loss-mask binding69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/mask-role-69/README.en.md) | Eligible35 positives/95 negatives,23 masked known candidates,63 unknown candidates | Preserve original candidates and truth while assigning zero loss weight to exclusions. |
| [Input bounds69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/input-boundary-69/README.en.md) | All72 original requests and216 captions fit existing512byte/32word limits | Project original text without shortening it or changing limits. |

## Which captions contribute to learning?

Function truth and caption fidelity are separate. A function can have the required behavior while its caption is incomplete. That candidate stays available, but its caption is excluded from loss. All21 unknown requests retain unknown truth and do not become negatives.

Groups with eligible loss contributions number9/3/3 in train/validation/calibration, preserving existing preparation minima. The Go projection is prepared to retain all90 known train rows and36 validation rows, assigning zero weight to18 and1 exclusions respectively. Calibration's27 known and9 unknown candidates remain audit material and do not enter fitting. Actual original72 projection proceeds under a separate frozen plan.

## How were the public captions checked?

The two Semver requests ask only whether leading `v` is accepted at exactly `v1.2.3`. The NewVersion quote is a package-level bullet rather than a function-specific comment; its parsing context and implementation path are recorded together. The two glob requests ask whether nested paths are included on two listed names. Proposed finite supervision contains5 positives and5 negatives. Auxiliary missing-patch and String observations do not extend the model target.

Actual function behavior, wording fidelity within a stated finite scope and sufficient authorship diversity are separate reviews. The original72 requests share a synthetic English authoring process, and the four upstream requests are AI-assisted drafts. Translations, comparison counts and candidate counts are not independent tasks. Original wording and full MIT notices are retained; model binaries stay out of GitHub.

## Reading resource measurements

Whole-child peak RSS was16.64MiB for roles66 and6.89MiB for public-function observations68. These describe preparation and observation processes. Go heap, GPU use, tiny-model inference and concurrent-serving costs need separate measurements. A rounded CPU time of0.00seconds does not mean no CPU work occurred. Raw logs, personal paths and local executables are excluded from public records.

## Next execution

[Stored rankings under frozen roles69](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/stored-role-utility-69/STORED-ROLE-UTILITY.en.md) retain all51 original function-truth requests. The best single validation control is lexical22checks versus17 for an answer-knowing oracle. This headroom meets the existing5% necessary condition and is not achieved model performance. Whole-scope costs remain a single BM25control91checks/oracle73; summing different role-specific controls to90does not replace that baseline. Complete-caption-only diagnostic B has no separate gate.

First project the original72 requests using frozen roles and masks, then check fixed-control verification costs with the same roles and original function truth. Appending the four upstream requests in a new version requires full code-family closure and a newly frozen role assignment. Do not change seed, roles, candidates or masks after seeing results. Compare INT8/PTQ children and ternary STE siblings after a useful FP32 parent. The2400distinct protected-final requests per domain remains a separate target.

[Progress issue19](https://github.com/teamswyg/laya-tools/issues/19) records actual CI, execution and publication status. Earlier preparation documents retain their historical unexecuted state.
