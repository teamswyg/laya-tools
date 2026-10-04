# Development diagnostic of symbols and relations

This phase tests whether a very small claim or hint can suggest which candidate to check first. It retains every candidate and grants no approval. We separate recovering symbols from using a small relation rule. No new training was performed.

## Observed diagnostic

The existing12 B-tree requests and96 finite labels are unchanged. They form one source family, group84, `development_train`, and were visible during grammar design. Ranking sees only raw request/candidate text; saved labels are consulted afterward. The existing34fixtures,272original observations, labels, masks and roles remain unchanged.

| Control | Counterfactual candidate checks | Sum of prior single-observation verification costs |
| --- | ---: | ---: |
| Current display order | 54 | 187,673 ns |
| Unlearned relation agreement | 26 | 85,793 ns |

All9 requests containing both true and false candidates rank a true one first. The empty-tree request with8true candidates is reported separately; both no-answer requests check all8. Two true candidates with different ranges remain true for the early-stop request. The26checks equal this packet's minimum count, but its work sum differs from the fastest-true oracle80,292ns.

This is a recomputation on exposed saved data, not an observed execution speedup from reordered source calls or a Codex token reduction. Since a simple relation control solves this task, learning must offer added value to justify its cost. Independent2,400-case evaluation and transfer to another lineage remain unproved.

## CPU and Go allocation

On Apple M4 Pro, Go1.27.1 and one CPU, one supported text was repeated eight times. Each measurement ran300ms, repeated3times. The table reports medians; complete output is in `BENCHMARK.actual.public.v1.txt`.

| Extraction API | Median per8candidates | Go B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Legacy word hash | 319,241 ns | 130,240 | 320 |
| Symbol-retaining hash | 432,449 ns | 171,456 | 352 |
| Legacy reused token plans and scratch | 299,022 ns | 704 | 24 |
| Relation array including raw parsing | 2,895 ns | 0 | 0 |

Symbol retention adds information but is slower and allocates more in this fixture. The relation array is small and fast but represents much narrower meaning. The warm legacy control excludes initial token/scratch preparation; relations are parsed every time. These APIs do not offer equal semantic/model quality, and one repeated sentence is not a production input distribution. Go allocation is distinct from RSS, native or GPU memory. No SIMD or GPU acceleration was verified.

## Use and reproduce

The separate Go executable has [human and agent instructions](../../../cmd/riido-hintpreview/README.en.md). From the repository root:

```sh
bash scripts/verify-text-representations.sh
go test ./internal/hintrelation -run '^$' -bench '^BenchmarkRepresentations$' -benchmem -count=3 -cpu=1 -benchtime=300ms
```

Source, raw text, saved observations and budget were frozen in `FREEZE.public.v1.json`. All first12responses/statuses and complete benchmark output were retained before interpretation. `RESULTS.actual.public.v1.json.gz` contains every candidate's fingerprints, relations, order and finite truth. Replay calls no original source API or model. Synthetic controls check exact legacy float64 bits without symbols, input/Unicode boundaries, grammar failures, partial recognition, array ownership and stable ties.

The new source and grammar are authored under Apache-2.0. Candidate descriptions and prior observations follow the [pinned B-tree source and complete notices](../next-cohort-btree-audit/SOURCE-PINS.public.v1.json). No new model/library license was introduced. Model bodies and private data are absent from Git.

## Conditions for the next training comparison

A is legacy word hash8192, B is symbol-retaining8192, and C is a separate8-dimensional relation representation. A future Fit must freeze the same raw text, labels, roles, order, seed, objective and epochs, while disclosing C's capacity/density differences. Never reinterpret legacy8192 CSR or RIIDOH01 weights. Coverage is not truth, loss or evaluation eligibility.

Actual70 assigns development-validation groups4/8/12/60:20parents,60candidates, with44eligible,1known masked and15unknown after existing masks. These were already exposed during71/72/79selection and are not independent validation. Reuse for a development ablation requires a newly pinned bounded adapter, source/helper lineage review and whole-family roles. Training on one B-tree family has a separate scope from the former9train-group readiness criterion. Next, adopt a new source's entire family as validation before observation and separate wording, truth and candidate position. Establish FP32 utility before separate INT8 or ternary experiments.
