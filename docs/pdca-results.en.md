# PDCA measurements — all attempts

[한국어](pdca-results.ko.md) · [Method and limitations](pdca-tuning.en.md)

Accepted means calibrated confidence at least 0.9. Each final has 24 cases; two seeds repeat training on the same data.

| Cycle / seed | Correct: base → tuned | Strong correct | Accepted correct/accepted | Order flips | Overall |
|---|---:|---:|---:|---:|---|
| [pdca-01 / 1729](../benchmarks/results/pdca-01/seed1729-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/9 | 34/120 | FAIL |
| [pdca-01 / 2718](../benchmarks/results/pdca-01/seed2718-evaluation.json) | 16/24 → 18/24 | 6/8 | 9/10 | 29/120 | FAIL |
| [pdca-02 / 1729](../benchmarks/results/pdca-02/seed1729-evaluation.json) | 16/24 → 16/24 | 2/8 | 0/0 | 25/120 | FAIL |
| [pdca-02 / 2718](../benchmarks/results/pdca-02/seed2718-evaluation.json) | 16/24 → 18/24 | 4/8 | 3/4 | 30/120 | FAIL |
| [pdca-03 / 1729](../benchmarks/results/pdca-03/seed1729-evaluation.json) | 16/24 → 22/24 | 7/8 | 5/5 | 14/120 | FAIL |
| [pdca-03 / 2718](../benchmarks/results/pdca-03/seed2718-evaluation.json) | 16/24 → 21/24 | 8/8 | 6/6 | 19/120 | FAIL |

Order flips compare five alternative orders of the same 24 cases (120 comparisons), not 120 independent tasks. Divide accepted count by 24 for coverage.

PDCA-01 raw reports store 9/10 precision as float32 0.899999976. Later cycles use integer-count division with the same 0.9 gate. Other failed criteria mean the historical overall failure is unchanged.

PDCA-04 stopped using validation only, without completing both seeds or running the final. Its [partial history](../benchmarks/results/pdca-04/stopped.json) is preserved. PDCA-05 reuses only that unexecuted final/calibration; viewed PDCA-01/02/03 finals are not re-sealed.

The transparent keyword reference scored 19/24 in PDCA-02/03. It has no calibrated confidence or acceptance policy; it is retained to avoid assuming a trained model is necessary. Adaptive repeated testing, small samples and authored labels limit repository/cost generalization.
