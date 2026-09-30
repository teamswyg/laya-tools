# Experiment 46 plan: a tiny auxiliary-search claim model

**All ten candidates were fitted on 4,456 training and 2,599 validation examples; usefulness gates failed.** Start with the [results and resource observation](RESULTS-46.en.md). Below is the explanation of the [first FP32 plan](plan-46.json), committed before cost inspection. Ten candidates are not ten evaluated tasks. Neither the plan nor its gates were changed after observing results.

## What do 24 and 2,400 count?

The initial [difficulty evaluation](../../docs/pdca-results.en.md) has only 24 actual tasks and remains a small pilot, insufficient to establish generalization or production utility. Distinguish the later search experiment's [24 model configurations](../call-change-audit/README.en.md) from its 2,948 actual questions. Repeated seeds, representations or choice orders do not create independent tasks.

| Quantity | Next-experiment requirement | Purpose |
|---|---:|---|
| Training rows | At least 2,400 qualified tasks | Fit coefficients |
| Development validation rows | At least 2,400 qualified tasks | Compare epochs, call penalties and thresholds |
| Final evaluation rows | 2,402 protected tasks; acquisition/source review separate | Confirm after freezing model and execution conditions |
| Model configurations | Five penalties × two seeds = ten | Candidates evaluated on the same data |

2,400 is a minimum for starting comparison, not sufficient evidence to declare success. Preserve existing connected duplicate groups and repository boundaries. Text/ID deduplication does not prove semantic independence or absence of pretraining contamination. Report benefit/harm/tie counts, per-repository results and rare failures separately; retain insufficient-evidence conclusions for small subgroups.

Hypothetical sample-size illustration: with zero failures in independent, same-distribution binomial trials, the one-sided 95% upper failure bound is `1−0.05^(1/n)`: about **11.73%** for 24 trials and **0.125%** for 2,400. These are not observed results or guarantees for our data. Repeated observation, correlated tasks and distribution shift invalidate direct use of this simple illustration. See [NIST exact binomial confidence limits](https://www.itl.nist.gov/div898/software/dataplot/refman2/auxillar/exacbici.htm).

Inputs are the fixed 16 numeric features from baseline search. Raw queries, repository names, file paths and gold ranks never enter model inputs. Output is a ranking score about auxiliary-search value. The sigmoid of a cost-weighted loss is not an actual success probability. Execution policy and fallback to baseline belong to a separate path.

## Readiness

Require at least 2,400 source-qualified successful cost rows in both training and validation, with at least 100 in each of five validation repositories. Keep all 7,335 training and 5,686 validation tasks in coverage. Missing data, invalid targets and source conditions do not become negative labels. Protect the 2,402 final tasks until model and threshold are sealed and final evaluation is separately prepared.

The [earlier experiment 45 readiness join](../path-cost-data/readiness-45.json) had **3,048 training and 107 validation** rows, so fitting did not start. Additional acquisition produced an [actual input seal](input-46.json) with **4,456 training and 2,599 validation** examples, meeting both gates. These are operational candidates for a narrow local numeric scope, not proof of all historical issue rights or public-release permission.

## Objective and candidate count

Five call penalties `{0, 0.25, 1, 4, 16}` and two seeds `{1729, 2718}` produce **ten FP32 candidates**. Keep features, learning rate, regularization, 100 epochs and batch64 fixed. Minimum validation loss chooses the epoch, with earlier epoch on ties. Do not select only a favorable seed.

Target is `baseline pages − helper pages − penalty > 0`, weighted by the absolute difference. Penalties are page-equivalent experimental values, not money or token prices. Divide every weight by the fixed constant 5,016 derived from 100,000-candidate/page20 bounds. Constant scaling preserves the trainer's normalized weighted loss and update interpretation. Do not clip large examples after inspecting their outcomes.

Validation threshold selection includes disabled baseline and minimizes page cost under a 90% helper cap. Keep equal-score groups intact. Apply fixed call-count, penalty, threshold and epoch tie-breaks. Compare baseline-only, always-helper and a baseline-gap heuristic with the same cap too.

## Required evidence

Both seeds must reduce validation pages by at least 5% versus baseline and 1% versus always-helper. Pooled and equally weighted repository Hit10 must not decline, and no validation repository may worsen total pages by more than 5%. Record failure rather than adaptively searching this same data family. These are pilot advancement gates.

There are 16 FP32 coefficients, with Go-owned contiguous training columns. Disclose float64 feature or accumulation arithmetic separately from coefficient storage. Separate head bytes from feature/search pipeline costs. The complete runner's isolated RSS target is at most 256MiB; report exceedance without hiding work or removing rows.

## Implementation preparation and remaining work

`cmd/riido-pathclaim`, `internal/pathinput` and `internal/pathclaim` join reviewed source, measured costs, fixed conditions and model formats. All ten actual trials and every epoch were executed; reproduction and resources are recorded in the [results](RESULTS-46.en.md). The maintainer runner reuses the same binary after input sealing and does not automatically score final, register a production policy or publish weights. Raw user/source text and model files are excluded from Git.

After useful FP32 signal, separately compare FP32/INT8/ternary PTQ/ternary training with identical inputs and budgets. Packing 16 coefficients alone need not accelerate the whole pipeline. File search still needs actual agent completion and total usage evidence; model up/down routing, dependency-aware decomposition and repository selection require their own golden sets. First-target pages cannot replace the full project objective.
