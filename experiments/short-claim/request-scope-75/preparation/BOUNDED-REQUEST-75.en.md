# Preparing explicit behavior scopes and captions: 75

This is a **request-version and supervision-eligibility proposal** for the same three parent identities and the same nine code candidates in their original order. Actual truth, roles, weights and pair eligibility remain `null`. Nothing has entered the runtime corpus or training. Original candidate v2, caption74 and readiness artifacts, plus this preparation's v1/v2, remain unchanged. Translations and A/B captions do not add parents.

## Requests visible to the model

| Parent | English request | Bytes / local word count |
| --- | --- | --- |
| datasize | Find the nonnil receiver method parsing ASCII integers and binary byte units: return syntax or bits errors with receiver zero, and range errors with receiver uint64 maximum. | 173 / 27 |
| query | Find struct query function: exported nonignored default fields retain duplicate primitive slice order; omit empty slices, keep empty strings without omitempty; included named nonnil non-time structs use brackets; exclude custom encoders. | 237 / 32 |
| shlex | Find the string splitter that keeps empty quoted words and returns completed tokens before an unclosed quote or trailing escape error. | 134 / 21 |

The first two use a separate scoped request version; shlex is unchanged. An independent ASCII identifier-boundary counter checked the 512-byte/32-word budget. Original Normalize, Validate and Features were not called.

The word `included` matters for query. The original implementation checks omission through `omitempty` and `IsZero` **before nested recursion**. Bracket nesting therefore concerns included named nonnil non-time structs. Flattened anonymous embeddings, nil pointers, special time formatting and custom Encoder implementations are outside that nesting assertion. The request concerns default primitive field/slice representation, without claiming general delimiter or option behavior.

## Code behavior and caption information are different

The source-reading proposal is that `UnmarshalText`, `Values` and `Split` satisfy their respective scoped requests; the other six directly selected APIs contradict a required shape or policy. `Encoder` declares an interface, rather than providing a concrete callable implementation. No unseen implementation or adapter is assumed. The wrapper bodies directly show delegation in `Parse`/`MustParse` and the single return versus accumulated list distinction in `Next`/`Split`. This is not a new API observation or a formal proof over all inputs.

A is the original short caption; B is the unchanged caption74 version. Each requested condition has separate supported, contradicted, omitted or unknown evidence. Facts established by hidden full code never fill information absent from the model-visible caption.

| Proposal | A | B | Common A/B intersection |
| --- | ---: | ---: | ---: |
| Captions sufficient for BCE supervision | 4 | 9 | 4 |
| Eligible positive candidates | 0 | 3 | 0 |
| Eligible positive-negative endpoint pairs | 0 | 6 | 0 |

A's four eligible candidates are negatives with decisive visible contradictions: receiver-free `Parse`, panic rather than returned error in `MustParse`, the `Encoder` interface and the single-word return of `Lexer.Next`. A signature or general one-line summary of the positive functions does not establish error-state changes, duplicate order, empty quotes or a completed-token prefix. B's sufficient positive behavior or decisive negative evidence remains a **proposal pending review**.

## A BCE mask does not enforce pair exclusion

Existing pair-ranking72 may include zero-BCE-weight endpoints in known-truth parents. This proposal records BCE weight and pair endpoint eligibility separately. Runtime enforcement has not been implemented. Before future use, a caller or pair plan must explicitly exclude insufficient endpoints, or conservatively exclude that parent's pair supervision. Existing72 specifications, numbers and results remain unchanged and are not retroactively invalidated.

A comparison isolating caption effects needs the same frozen mask and weights in both arms. This slice's common intersection has no eligible positive, so it **is not a trainable paired comparison dataset**. Arm-specific masks change both caption information and supervision coverage and constitute a different experiment. Underlying code truth, acceptable candidates, no_answer and unknown evaluation denominators remain intact either way.

## Sources and handoff

The retained sources are fixed revisions of [datasize](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/datasize.go) ([MIT](https://github.com/c2h5oh/datasize/blob/aa82cc1e65004e2b59a6e44d26f774ca961b24d8/LICENSE)), [go-querystring](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/query/encode.go) ([BSD 3-Clause](https://github.com/google/go-querystring/blob/965d79f2113ea0ff039d29828a08a616a0223d48/LICENSE)) and [shlex](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/shlex.go) ([Apache 2.0](https://github.com/google/shlex/blob/e7afc7fbc51079733e9468cdfd1efcd7d196cd1d/COPYING)). Original license and copyright-notice pins are retained; no relicensing claim is made.

A stdlib-only metadata tool prepared the new JSON while checking 3 raw source file hashes, 26 evidence spans, 9 candidate declaration spans and existing IDs, order and caption hashes. Actual original source AST, APIs, initialization, observer, roles, models, fitting, protected-final and remote executions are 0. The preparation ledger records 3 successful metadata versions and 0 failures, preserving source-scope amendments.

The author is an AI-assisted source reader previously exposed to Wants and results. This is not a blind assessment, proof of independent authorship, or complete semantic validation. Whole repository families, aliases and helpers stay together. The 2,400-per-domain final target, source-diversity limitations and existing training requirements remain unchanged. Later review governs corpus inclusion; writing this proposal or passing CI does not itself create training eligibility.
