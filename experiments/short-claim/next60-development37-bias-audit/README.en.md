# Candidate-order comparison on 37 development requests

The tiny `riidolaya` claim/hint model aims to propose a useful verification order. We first measured inexpensive nonlearned controls. **Lexical ordering is the strongest current control; a new model has not demonstrated utility beyond it.**

| Nonlearned control | Primary top-1 positive | Primary unit checks to first positive | Equal-parent permutation mean checks |
| --- | ---: | ---: | ---: |
| Fixed input order | 6/37 | 71 | 1.9595 |
| BM25 | 26/37 | 51 | 1.3784 |
| Lexical ordering | 29/37 | 48 | 1.3378 |
| Narrow rule | 26/37 | 51 | 1.3784 |

Always choosing the second candidate already succeeds on28/37. Positive positions `[6,28,3]` are biased, so primary accuracy alone is not learning evidence. The narrow rule does not support these requests and falls back to BM25 on all247 calls. It is not independent rule performance.

The37 primary orders and210 exhaustive permutations are separate series:247 actual calls,988 rankings and2,928 active scores. **These are not247 independent samples.** Equal-parent means neutralize the different permutation counts for two/three candidates. Ties preserve current display order, every candidate survives, and no per-request winner is substituted as the reference control.

The diagnostic assumes cost1 per candidate check. Lexical ordering's equal-parent total49.5 versus a perfect-order lower bound37 leaves approximately25.25% hypothetical headroom. This is not an achieved model gain. Verifier, preprocessing, inference and fallback costs must all be observed before claiming5% whole-work utility. This experiment measures no time, token, money or CPU/GPU improvements.

[results.json](results.json) contains the actual report; [POSTCLOSE](POSTCLOSE.actual.public.v1.json) pins the closed file; [INDEPENDENT-READBACK](INDEPENDENT-READBACK.actual.public.v2.json) records a separate Node matrix recomputation. That implementation checks saved-score ordering, permutations and aggregates. It does not independently establish BM25 score correctness or actual verifier cost. Existing predicate unknowns remain in their original evidence and do not become negative labels.

The table preserves the Mac execution. Linux BM25/narrow-rule permutation means are1.3649 because of near-tie rounding; lexical remains best on both. This environment difference is not an improvement. [The scope supplement](READBACK-SCOPE.en.md) lists which exported aggregates the Node readback did and did not separately check.

Reproduce from this repository with Go1.27.1:

```sh
bash scripts/verify-next60-bias37.sh
```

The command checks owned controls, reads the public37-row file and runs the nonlearned comparison once into a fresh output. It starts no original library, model or Fit. Absolute and relative score tolerances remain1e-12 each. Each platform's saved actual scores must independently reconstruct its complete permutation orders, ties, costs and matrix digest exactly. Common data, labels, denominators and no-training authority must match. Tolerance never changes actual scores or rankings.

The first Linux CI failed the original requirement that every aggregate match the Mac report. [The Linux actual report](LINUX-AMD64.actual.public.v1.json) differs by at most approximately5.33e-15 in score, yet one request's BM25 pair becomes exactly tied and changes permutation visits. [Independent recomputation](LINUX-INDEPENDENT-READBACK.actual.public.v1.json) confirms the Linux matrix and aggregates against its own scores. [The difference receipt](PLATFORM-READBACK.actual.public.v1.json) preserves the failure and both inputs. This is not learned improvement or measured work savings. Well-separated pair ordering must still agree; near-tie changes are reported separately.

An oversized result remains a failure. If storage completes successfully, a bounded record preserves errors and actual counters. Write/sync failures remain separate errors. This maintainer source is not automatically registered with the default CLI or Codex.

[The separate20-to-60 collection plan](NEXT-COHORT.en.md) follows. Existing37 rows, the new cohort and protected2,400/domain evaluation remain distinct. New Fits, independent groups and model activations in this step:0.

[한국어](README.ko.md) · [HF37 publication record](../next60-hf37-publication/README.en.md)
