# Two additional requests: native observation, comparison, development adoption

We executed a bounded string-slice task and an all-errors argument-validation task against real Go libraries. Each request has five frozen inputs and three candidates: the original, an authored reference, and an authored negative. The resulting **30 observations support two development requests**; they are not 30 independent training requests.

| Scope | Observed result |
| --- | --- |
| Bounded string-slice / CSV behavior | 15 observations, 15 normal returns |
| Combined validation errors / callback order | 15 observations, 15 normal returns |
| Satisfied observations | 23 |
| Observations with a known counterexample | 7 |
| Unknown overall satisfaction | 0 |
| Individual predicates | 125 satisfied / 15 unsatisfied / 1 unknown |

The original validation candidate's `B_code` is unavailable for one input and remains `observable_unavailable`. Other predicates in that observation have known callback-count and error-type contradictions, making the overall observation unsatisfied. The unknown predicate was not coerced to false.

Each reference satisfies the known conditions for all five inputs; every other candidate has a known counterexample. Root separately adopted these two finite requests into development data. Combined with the previous 21 requests, the pool contains **23 requests, 67 candidate labels, and 113 input variants**. It retains 335 observations and selects 331 for candidate supervision; the previous four unknown-observation exclusions remain unchanged.

## Reading the evidence

- [FIXTURES.v1.json](FIXTURES.v1.json) and [WANTS.v1.json](WANTS.v1.json): inputs and conditions frozen before execution.
- [SLICE-OBSERVATIONS.v1.json](SLICE-OBSERVATIONS.v1.json) and [VALIDATORS-OBSERVATIONS.v1.json](VALIDATORS-OBSERVATIONS.v1.json): typed outputs of the real worker processes.
- [COMPARISON.v1.json](COMPARISON.v1.json): comparison of saved observations against Wants; this record has no adoption authority.
- [ROOT-QUALIFICATION.v1.json](ROOT-QUALIFICATION.v1.json): separate Root adoption, with exact requests, candidates, and selected observations.
- [FIRST-TRIALS.v1.json](FIRST-TRIALS.v1.json): the historical **pre-comparison** first-execution receipt; its pending comparison and `qualified=false` remain intact.

For each worker Root verified one Start, one Wait, exit0, and reaping. Maximum worker RSS was 6,537,216 bytes for string slices and 6,488,064 bytes for validators. Controller start-to-wait time was approximately 0.651s and 0.489s. These measurements include process startup and durable output of small observation workers. **They do not measure Laya inference, GPU use, or routing savings.** The Go heap limit is soft, and the RSS limit is checked after execution.

## Agent and CI use

Run the saved-result check from the repository root:

```sh
bash scripts/verify-next60-extra-two.sh
```

It checks saved-file fingerprints, recomputes all 30 observations and 141 predicates with the Go array-based comparator, and compares every resulting field with the stored report. Synthetic controls cover truncated input, duplicate fields, mutated conditions, and unknown handling. It launches neither original candidates nor models. Private journals, persistence acknowledgment, and reaping were verified during Root's first collection; public CI does not claim to replay those private journal acknowledgments.

`source/` archives the authored Go observer, writer, outside controller, and comparator as `.go.txt`. CI restores only authored modules into a temporary directory. The public bundle excludes original library bodies, binaries, model weights, host paths, and raw journals. `native-*` records require separately acquired originals and are not restored by CI.

## Meaning for the next training checkpoint

This is development material for a small model that supplies hints about candidate constraint satisfaction, not a prose-generation benchmark. The reference occupies position2 in 20/23 requests, while 44/67 candidate labels are negative. These use different denominators and are not model accuracy. Candidate-order perturbation and semantic-group splitting remain necessary.

This adoption performs no new Fit, feature extraction, model scoring, or activation. The next Fit requires **30 qualified distinct requests**, followed by a separate 60-request development checkpoint and 2,400 protected evaluation requests per claimed domain. A 23-request pool is progress in preparation; it does not establish cost or time savings on user work.

Authored material follows the repository Apache-2.0 terms. Original pflag, Cobra, and selected Go notices are retained in [notices/](notices/). Local research execution review does not confer blanket source, binary, or model redistribution clearance.
