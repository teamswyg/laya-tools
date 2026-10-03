# Nine development tasks checked by an actual execution

This record checks code to produce supervision for a small claim model. It is
not Laya inference, a GPU run, model training, or evidence of cost savings.

Nine frozen public requests contain 46 inputs and three candidates each:
original, authored reference, and authored negative. The first native trial
saved 138 observations. A separate comparator then checked those saved results
against Wanted frozen before execution. No original replay or Wanted edit was
performed. The tenth unknown-token request remains on hold.

| Check | Result |
|---|---:|
| Distinct tasks / inputs / candidate observations | 9 / 46 / 138 |
| Fully satisfied / known mismatch / unknown observations | 102 / 35 / 1 |
| True / false / unknown comparison predicates | 1,246 / 72 / 29 |
| References satisfying every required frozen predicate | 9 |
| Original and negative candidates with known counterexamples | 18 |
| First native attempts / replays / panics | 1 / 0 / 0 |
| Worker peak OS memory | 8 MiB |
| Execution including recording and synchronization | About 1.51 seconds |

Memory is worker maximum RSS from macOS `/usr/bin/time -l`, not Go heap or GPU
memory. The observer explicitly counted 899 returned calls, including 42 error
method calls. Library-internal and startup calls remain unknown. Returned file
synchronization acknowledgements are not a universal power-loss guarantee.

A positive candidate has every required predicate known true on these inputs.
A negative candidate has at least one known counterexample. Unknown predicates
remain unknown; they are never converted to false. An original with an
unobservable channel can still have a known counterexample on another input.

These requests reuse the existing pflag and mapstructure connected families.
They provide no new unseen-source evaluation. All nine reference candidates
occupy index 1, and caption style may reveal supervision. Later evaluation must
include candidate-order permutations and position-only and lexical baselines.

`evidence/` contains public-fixture observations and the local outside receipt.
`independent-semantic-review/` compares saved observations. Its reviewer did not
author the original code, Wanted, or candidates; the review is explicitly
nonblind. Root separately adopts supervision in
[`next60-development-sixteen`](../next60-development-sixteen/).

`source/` archives owned Go source as `.go.txt`. CI checks synthetic failure
controls for the outside controller and comparator, re-compares saved public
observations, and validates data. CI does not run this original worker or
produce 138 new native observations. Original library bodies, binaries, raw
journals, host paths, and configurations are excluded. Exact upstream revisions
and complete notices are recorded in `evidence/ROOT-RIGHTS.frozen.v1.json` and
`notices/`.
