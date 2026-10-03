# 30 public code requests for a small claim model

This is the **next development training data** for a small model intended to run repeatedly with very little memory and offer useful hints about code candidates. A model proposal remains a hint; tests and observed evidence decide the final result. This package contains no newly trained model or model weights. It is not a new training or inference result, or a measured reduction in Codex costs.

## What is included?

| Item | Current data |
|---|---:|
| Parent requests | 30 |
| Selected candidate labels | 88: 30 suitable, 58 unsuitable |
| Frozen finite inputs | 146 |
| Full candidate observations | 434 |
| Observations of candidates selected for training | 430 |
| Preserved unknown predicates | 8 |
| Actual Go Reader returns and value matches | 30 / 30 |
| Saved Reader checkpoints checked by Root | 60 |
| Connected development source groups | 6 |

Each request has two or three candidates. One original error-writer candidate has unavailable required channels and is excluded from training. This explains the difference between 434 full observations and 430 selected-candidate observations. Unknown is not converted to unsuitable. A known counterexample and unknown predicates can coexist; both remain recorded.

The earlier 23-row and 27-row JSONL files remain exact byte prefixes of the current file. See the [prefix check](evidence/PREFIX-CORRESPONDENCE.v1.json). Only `request` and candidate `text` in the [data](data/train.jsonl) are model inputs. IDs, labels, source names and revisions, group numbers, roles and observation evidence are excluded from features.

## How do I read and use it?

1. Read each JSONL request and its candidates. Labels were adopted **only for that request's frozen finite inputs**. They do not guarantee correctness for all inputs or another repository.
2. Check the [actual materialization summary](MATERIALIZATION.v1.json) and [actual Go Reader summary](READER.v1.json). These are exact copies of safe original summaries. Materialization and reading are different stages, so their `LoadDevelopmentRow` call counts and other stage fields differ.
3. Inspect the final three requests through [111 Wanted and observed predicates](evidence/REMAINING-THREE-PREDICATES.v1.json). Distinguish `satisfied`, `unsatisfied` and `unknown`. Full Got traces and original journals are withheld.
4. Read the [Root admission evidence](qualification/REMAINING-THREE.v1.json) with the [source group decision](qualification/GROUP-DECISION.mode30.v2.json). The earlier [Sub/body](qualification/SUB-BODY.v2.json) and [two INI requests](qualification/INI-TWO.v1.json) are also retained.

States such as `pending` in group decisions and source-preparation archives describe their individual seal times. Their bytes were not edited to appear current. The [public correspondence summary](evidence/ROOT-DATA30-PUBLIC-CORRESPONDENCE.v1.json) separately records the actual completed scope for 30 rows. This new package does not yet claim a GitHub CI pass or completed publication.

The public repository's `pkg/shortclaimdata.LoadDevelopmentRow(io.Reader)` is the Go interface for reading one JSONL row at a time. The [archived Go sources](source/) and [source correspondence](SOURCE-DERIVATION.v1.json) document materialization and validation. `.go.txt` and `go.mod.txt` are inert source archives. They do not replace the actual module links in a local workspace. The Reader modfile containing a physical local path and original dependency bodies are omitted.

## What was checked, and what remains?

The final three requests were collected once, producing 42 candidate observations; only saved values were compared. Observation rows contain 29 satisfied, 11 unsatisfied and two unknown results. Their predicates contain 89 satisfied, 20 unsatisfied and two unknown results. These are different units from the nine final candidate labels for those three requests. All required predicates are known true for each of the three reference candidates; the other six candidates have known counterexamples. Maximum RSS of 18,268,160 bytes and roughly 2.52 seconds in the [collection summary](evidence/NATIVE42-PUBLIC-SUMMARY.v1.json) cover the whole child process, including startup and source checks. These are not measurements of Laya, GPU or small-model inference.

The Reader returned and matched all 30 rows. Root also checked 60 alternating saved checkpoints in its private Reader journal. The public summary cites that verification; it does not replay or independently reconstruct the journal. Data-file-writing 0/false fields in `READER.v1.json` mean that the Reader did not generate data, not that its saved checkpoints failed.

The selected positive candidate occupies the second position in 27 of 30 requests. Always choosing that position yields a **90% parent-selection position control**. This is not model accuracy. Generalization or utility requires further position and wording controls and splits that keep an entire source family together.

The current Hugging Face release contains [23 requests](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-23-finite-v1); publication of this 30-request package remains pending. New Fit calls and new models are zero. Historical totals remain three Fit calls and three logical models, with overall limits of eight Fit calls and 16 models. Thirty requests form a collection checkpoint. The next main training step separately requires about 60 collected requests, frozen group/role decisions and a readiness decision. A 30-only exploratory fit also remains pending a separate admission. This does not claim completion of the 2,400 protected evaluations or a demonstrated 5% cost or utility improvement.

## History and licenses

The [failure history](history/FAILED-ATTEMPTS.public.v1.json) preserves the first strict-config rejection of pretty JSON and the earlier data27 stage-sequencing failure. Later successes do not erase those failures or turn them into successful attempts.

This publication preparation was performed by a nonblind contributor who previously authored some writers, data helpers and validators. It is not a new independent label decision or independent review of that contributor's own code. No comparison, training or original API execution was repeated. Original source bodies, binaries, model weights, personal paths, tokens and original journals are excluded. Derived summaries have new schemas and source hashes; edited metadata is not presented as byte-identical to an old seal.

Owned code is under [Apache-2.0](LICENSE). Referenced project terms and retained full notices are listed in [NOTICE](NOTICE.md) and [notices](notices/). This does not certify rights for every model, dataset or historical copied body.
