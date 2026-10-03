# Two INI tasks: actual observations and finite development adoption

We are building a small model that offers frequent **claim or hint** scores for candidate actions. This tranche binds quoted-value parsing and duplicate-section deletion to actual Go originals. Distinct semantic requests increased **25 → 27**, and candidate labels **73 → 79**. The published training file still contains **23 rows**. Source-group closure and generation/reader correspondence precede the next materialization.

| Task | Inputs × candidates | Satisfied / unsatisfied / unknown observations | True / false / unknown predicates |
|---|---:|---:|---:|
| Comment markers inside quotes | 5 × 3 | 9 / 6 / 0 | 21 / 7 / 2 |
| Duplicate-section deletion bounds | 5 × 3 | 10 / 5 / 0 | 32 / 5 / 2 |

Each reference satisfied all five inputs. Both the baseline and seeded negative had known counterexamples. For example, `k="a#b;c" # tail` must preserve markers inside quotes. Deletion must reject negative/out-of-range/missing-name requests while preserving valid duplicate deletion. After the original negative-index deletion, one getter panicked. Two later state predicates remain **unknown**, separate from the already observed error-kind contradiction. Panic itself is not a negative label.

[ROOT-QUALIFICATION.v1.json](ROOT-QUALIFICATION.v1.json) records separate adoption after actual comparison. Frozen texts, inputs and Wants were unchanged. Four unknown predicates remain unknown. Thirty observations and six candidates are only two independent requests. The cumulative pool is **27 requests,79 labels (27 positive/52 negative),132 inputs,392 original observations and388 selected observations**. New labels are finite development supervision with unit weights.

The quoted task covers five single-line inputs, not every parser-option combination. Deletion `x_indexes` denotes occurrences observed through public APIs, not private indexes. Preflight validation and delegated mutation do not establish concurrent atomicity. The adoption record declares these limits.

Each original worker started once and was reaped once. Quoted parsing produced125 frames, maximum OS RSS6,553,600B and0.9011s start-to-wait time. Deletion produced1499 frames,10,895,360B and6.5542s. These **collection-worker** measurements include startup and durable IO; they are not model/GPU/Go-heap/energy/Codex-savings measurements. Later comparisons consumed saved results without replaying originals.

`source/comparer` archives owned Go comparison code and synthetic failure controls as `.txt`, without upstream bodies. This command restores the source in a temporary directory and runs race/vet; it launches no original worker or model:

```sh
bash scripts/verify-next60-ini-two.sh
```

`PREDICATES.*.v1.json` contains reduced actual Wants/Gots. Raw journals, worker dumps and binaries remain private and are referenced by hashes. Initial collection files retain their historical pending/25-request state; later comparison and adoption records supply subsequent evidence. The earlier Sub/body qualification's HF21/CI-pending fields are also historical. [Separate publication proof](../publication-proof-120/README.en.md) records actual HF23 and PR119 merge.

**New training remains zero before30 qualified requests.** Three tasks remain: total INI input-byte budgets, bounded file reads and exclusive writes. Always choosing index1 gives24/27; always predicting negative gives52/79. These are controls with different denominators, not model accuracy. Position/wording controls, connected-source separation, a60-request checkpoint,2,400 protected requests per claimed domain and5% utility guard remain required.

Owned code/prose use Apache-2.0. Full INI Apache and comparable Go BSD notices are retained; a comparable notice is not clearance of every historical copied lineage. No model weights are included. [한국어](README.ko.md) · [Progress issue](https://github.com/teamswyg/laya-tools/issues/19) · [Public23 rows](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-23-finite-v1)
