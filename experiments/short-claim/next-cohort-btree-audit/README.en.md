# Find where a small hint can help before tuning

This experiment screens verification headroom before another model Fit. Eight original `google/btree` APIs cover inclusive/exclusive bounds, iteration direction, callback early termination and empty input. Sources, requests, exact document captions, literal Wants, roles and budgets were frozen before the first original trial, after synthetic failure controls.

There are **12 requests, one connected source family, 34 fixtures and 272 candidate observations**. The whole family was prospectively assigned group84 `development_train`, conservatively including the documented GoLLRB API lineage. Legacy 37 data and roles remain unchanged. This is neither 49 independent Golden examples nor a new validation family.

## Results

| Control | Checks to first acceptable | Top1 /10 answerable | Top3 /10 answerable |
|---|---:|---:|---:|
| fixed order | 54 | 3 | 5 |
| BM25 | 59 | 2 | 4 |
| lexical | 58 | 2 | 4 |
| narrow rule | 59 | 2 | 4 |

All 12 narrow-rule requests fall back to BM25. The optimistic oracle needs 26 checks: **51.85% count headroom** against 54. This is an opportunity, not an improvement achieved by a learned model.

The96 finite candidate states are 18 T/78 F/0 U. Parent classes are eight single-answer, one multi-answer, one all-positive and two no-answer. Both no-answer requests exhaust all eight candidates. Full Got, ordered Wants, callbacks, returns, states and costs were independently recomputed. There are 272 normal returns and 1, 088 actual callbacks. Each timing sample includes fresh-tree construction/insertion, iteration, callback observation and literal comparison.

For this inexpensive Go verifier, the optimistic additional hint allowance is only about **8.62 µs/request**; requiring 5% lower saved verifier work leaves about 7.85µs under the same-overhead assumption. Preparation plus all four controls was measured together at 854, 045 ns, not as individual control latency. One instrumented fixed-order sample does not establish causal speedup, CPU usage or LLM token savings.

## Prepared training inputs

[cohort.train.v1.jsonl](cohort.train.v1.jsonl) preserves exact original captions and 96 finite labels. Known training loss weights are 1; independent evaluation eligibility is false. Actual Go `LoadCohortRow` succeeded for 12 rows and `ProjectCohort` once. **Fit, new weights, Laya encoder and GPU/MPS calls are zero.** Structural Reader/Project success is not permission or readiness to train/activate a model.

Only the two no-answer requests use `with`; other bounded requests use `keeping`. Preserve this first run and vary wording independently of truth in future prospective acquisition. See the [English](NEXT-ABLATION-PLAN.en.md) / [Korean](NEXT-ABLATION-PLAN.ko.md) representation-ablation plan. Independent2400 evaluation and real agent savings remain unproved.

## Reproduce

From the repository root, with Go 1.27.1 and Node, verify saved evidence, synthetic failure controls and the actual Reader/Project. Node prepares research sources and compares saved records; the observer and learning-data interfaces are Go.

```sh
bash scripts/verify-btree-utility.sh
```

Also replay the original B-tree APIs and compare all272 Got values and control rankings. Fresh timings never replace the frozen first experiment. Linux/macOS CI uses this mode, with no model download or inference.

```sh
bash scripts/verify-btree-utility.sh replay
```

## Sources and records

The selected source/doc comments are Apache-2.0 from [pinned google/btree](https://github.com/google/btree/tree/aeba20f7a1e1315badec4eca4fdc9f754f5f880a). The [complete upstream license](LICENSE.upstream.txt) and exact five-file `upstream-source.json.gz` are preserved. Go 1.27.1 selects `btree_generic.go` and excludes legacy `btree.go`. This notice review is not a blanket certificate for every ancestor, unselected file or future model.

`FREEZE` precedes original trials; `READER-FREEZE` precedes the actual Reader/Project. `OBSERVATIONS` holds the first complete trace; `INDEPENDENT-READBACK` records independent calculations. `QUALIFICATION`, `MASK-POLICY` and `ROW-BINDINGS` bind finite labels, roles and interpretation. Local archive names in `RESOURCE` support resource auditing; the public package supplies the same complete sources, inputs and evidence as individual files. No model body or raw profile is committed to Git.

CI compares Got and finite judgments exactly. Scores use the existing platform policy of `1e-12` absolute and relative tolerance, while each platform independently reconstructs its exact score order, checks and costs. Well-separated score pairs must keep their order. Near-tie changes are reported without replacing the first Mac trace. First stdout, stderr and status are retained before acceptance and uploaded as CI artifacts even on failure. See the [policy](REPLAY-POLICY.public.v1.json) and 19 synthetic platform controls.
