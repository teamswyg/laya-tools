# Twenty-one finite development requests

[한국어](README.ko.md) · [Five actual observations](../next60-five-actual-observation/README.en.md)

This development data prepares the next training round for a tiny claim model. Its output is a hint about a candidate; tests and verified evidence determine the final outcome. Reading or validating this dataset does not train or execute a model.

The first sixteen rows remain byte-identical. Five distinct requests produce 21 requests and 61 candidate labels:21 positive and40 negative. The [28,800-byte JSONL](data/train.jsonl) is supported by103 frozen inputs and305 original evidence observations. Selected candidates account for301 observations. Input variants and candidates are not counted as distinct requests.

The error-writer baseline return is unknown, so that candidate and its four observations are excluded from the training selection. Unknown is not relabeled as false. Other candidates retain both known counterexamples and separate unknown conditions. See [Root adoption](ROOT-QUALIFICATION.v2.json), [actual materialization](MATERIALIZATION.v1.json), and [actual Go reader validation](INPUT-VALIDATION.v1.json).

The Go materializer checks the pinned inputs and adoption, then syncs the file and parent directory. A separate validator assigns no labels: it loads21 rows through the real reader and checks request/candidate text, IDs, provenance, labels, weights and counts. Features, Score, Project, Fit, original candidate and model calls are all0. The materializer reviewer did not author the candidates, Wants or materializer, but did author the comparer and reader validator; the review is nonblind. It is not an independent review of their own validator.

From the repository root, these commands reproduce saved-result comparison, data generation/reader validation and owned synthetic failure controls. They require Go1.27.1 and a C compiler for race checks. They never replay the original observation worker.

```sh
bash scripts/verify-next60-development.sh five
bash scripts/verify-next60-development.sh data
bash scripts/verify-next60-development.sh twentyone
```

CI uses the same checks. These added steps contain no model training or inference; the existing Laya native-inference step in the full CI is separate. Owned sources are archived as inert `.go.txt` files and copied to temporary modules for checks. `independent-validator-preparation/` preserves preparation history; the actual validation artifact records completion.

Selecting the second candidate gives an arithmetic18/21≈85.7% positional baseline. Calling all candidates negative gives40/61≈65.6%, with a different denominator. Neither is measured model accuracy. Source groups76,77,78 are reused, so no new-source generalization is established. No new corpus Fit starts before30 qualified requests. Later checkpoints remain60 requests,2,400 protected evaluations per claimed domain and a5% utility guard against simple baselines. Model improvement and actual Codex savings remain unproven.

This folder contains no model weights, executables, private host paths or raw journals. The previously published [immutable HF16 version](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-16-finite-v1) is available separately. This README does not claim that the21-row dataset has been published to HF. Owned materials follow repository Apache-2.0; full upstream notices are retained in the five-observation folder.
