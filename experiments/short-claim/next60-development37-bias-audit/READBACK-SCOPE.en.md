Scope of saved-report recomputation
==================================

The existing Node receipt recomputes primary order, permutation visits/top1/tie pairs/tied rankings, integer parent-macro sums and the complete matrix digest from saved actual scores. It independently validates neither BM25 scoring formulas nor actual verifier cost.

Its phrase “all permutation metrics” is too broad if interpreted as complete JSON validation. It does not separately compare fallback summary arrays, pair denominators, some primary summaries, floating macro means, better/worse/equal counts or every schema/state field. Exact input SHA pins and a matching matrix narrow that execution's scope; they do not replace general input validation or complete report validation. The original receipts and actual reports are preserved.

The successor Go saved-report control reconstructs the complete report from each platform's own scores and checks score-dependent fields exactly. It does not overwrite the Mac report with Linux aggregates or force near-equal scores into ties. The older Node receipt's file write/Sync alone also does not claim a general durable transaction including parent-directory Sync.
