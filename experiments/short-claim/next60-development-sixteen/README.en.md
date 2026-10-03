# Sixteen development requests

This public supervision tests whether a small model can supply useful hints
about descriptions of code-task candidates. Each request has one request text
and two or three candidate captions. The count of sixteen refers to distinct
semantic requests, not input variants or candidate positions.

| Item | Current value |
|---|---:|
| Distinct requests | 16 |
| Candidate labels | 47: 16 positive, 31 negative |
| Frozen inputs / candidate observations | 80 / 236 |
| Data file | 22,097 bytes, approximately 22KB |
| This materialization / actual Go reader validation | One / 16 rows passed |
| Additional training or model inference | Zero |

The existing seven rows are preserved byte for byte. Nine rows append the exact
request and caption text proposed before execution. Their labels follow one
actual native trial, independent comparison of saved results, and a separate
finite Root adoption. Unknown predicates are never converted to false.

`data/train.jsonl` is the training-input format. `ROOT-QUALIFICATION.v1.json`
records admission of the new nine. `MATERIALIZATION.v1.json` and
`INPUT-VALIDATION.v1.json` are actual generation and reader results. Labels do
not guarantee every possible input. The existing three connected source groups
are reused; no unseen-source generalization or model improvement is established.

Only request and candidate-caption text may enter features. IDs, revisions,
labels, groups, observations, and weights are bookkeeping. Host paths,
credentials, private user work, and binaries are excluded.

Fourteen of sixteen positives occupy candidate index 1. Always choosing that
position would select the positive parent candidate in 87.5% of these requests.
Always calling every candidate negative would label 31/47, approximately 66.0%,
correctly. These arithmetic baselines use different units and are not measured model accuracy. Future
evaluation needs candidate-order permutations, position-only and lexical
baselines, independent validation, and eventually approximately 2,400 protected
semantic requests per claimed domain.

No new corpus fit starts before the checkpoint of thirty qualified distinct
requests. Sixteen rows are not a 2,400-request evaluation; paraphrases cannot
fill the gap. Failed models remain inactive and the historical 79 are preserved.

CI reproduces and validates this file with Go using
`bash scripts/verify-next60-development.sh data`. Mode `sixteen` checks synthetic
failures in the materializer and validator. These steps launch neither the
original observation worker nor model training. Native evidence and complete
notices are in [`next60-nine-actual-observation`](../next60-nine-actual-observation/).

This version is currently prepared locally for a PR. The verified published
Hugging Face version is
[`next60-7-finite-v1`](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-7-finite-v1).
After this change passes CI, sixteen will be published as a new immutable
version with file-by-file downloaded verification.
