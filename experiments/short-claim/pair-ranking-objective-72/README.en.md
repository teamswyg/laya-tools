# A sibling experiment for ordering candidates within a request

The first FP32 fit required 31 checks on development validation, compared with 27 for the lexical/order rule. This follow-up preserves per-candidate BCE and adds a **same-request positive-above-negative loss** with fixed λ=1. It changes the objective in a separate experiment; the second corpus fit and its utility result have not yet run.

Opt in through `FitWithRanking` or `FitWithRankingTrace` in `internal/pairlearn`. Default APIs and valid λ=0 results preserve the original implementation. Pair losses average within each request, then across eligible requests, so extra candidates do not dominate. Masked pairs remain in the audit with weight zero; no-answer requests remain in BCE. Excluding unknown truth and calibration is the caller's responsibility. Request IDs and supervision weights never become model features.

New arrays use owned SoA columns without adding maps or locks. Their 64 MiB payload cap is not a process RSS cap. The planned single follow-up keeps the seed, features, truth, masks, roles, epoch selector and utility gates unchanged. Because the failed validation informed this experiment, any improvement is a development signal. Final evaluation still requires at least 2,400 distinct requests per domain from fresh source groups. This is an objective sibling, rather than a compressed child of the first FP32 model. Ternary compression remains a separate phase.

Synthetic gradient, final-minibatch and compatibility checks passed. Independent review also checked literal pair weights, the mean gradient over 231 batches and exact parity with historical code in 12 configurations. These checks do not establish corpus utility or generalization. There is no new CLI activation, Codex registration, GPU execution or claim of LLM savings.

[Implementation and PDCA](author/PDCA.en.md) · [Independent review](independent/FINDINGS.en.md) · [한국어](README.ko.md)
