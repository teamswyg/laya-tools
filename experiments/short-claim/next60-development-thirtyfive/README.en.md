# 35 development requests for small claim/hint routing

This development data contains 35 requests, candidate descriptions and 102 candidate labels: 35 positive and 67 negative. All bytes of the previous33 rows are preserved. Two GJSON requests were added after Root adopted their frozen finite evidence. One actual Go materialization exited0, and the project Reader matched every input, supervision and metadata value across35 rows.

The goal is a small claim/hint model that suggests which candidate to verify first from a request and short descriptions. An agent or person could optionally use that signal and attach further verification to decisions, generation and execution. These data do not establish model accuracy, savings or low-memory inference. This addition performed zero new Fits or model runs.

| Item | Verified count |
| --- | ---: |
| Development requests / candidate labels | 35 / 102 |
| Positive / negative labels | 35 / 67 |
| Frozen inputs | 171 |
| Full original saved observation evidence / selected-candidate observations | 509 / 500 |
| Requests with the positive at positions0,1,2 | 5,28,2 |

The additions validate one complete JSON value per nonblank physical line while preserving earlier callbacks and callback stop, and reject malformed UTF-8 before writing to the destination buffer. Their11 fixed inputs,6 candidates and33 observations support6 labels (+2/−4). The309 saved predicate judgments are255 true,18 false and36 unknown. Inputs, observations and predicates do not become extra training rows or weights. The original APIs offer different contracts; a mismatch with these requests is not a claim that the library is buggy.

All36 new unknown predicates remain alongside the prior33 post23 scoped10, giving scoped46. No whole-history unknown total is claimed. Excluded OriginalInt metadata retains null labels and weights. A known counterexample can support a negative while other predicates remain unknown; unknown-only evidence does not become negative.

Read [data/train.jsonl](data/train.jsonl) with [MATERIALIZATION.v1.json](MATERIALIZATION.v1.json) and [Root qualification v2](qualification/ROOT-QUALIFICATION.actual.public.v2.json). Reader attempts, returns and full matches are35 each. Another35 ValidateInput calls separately check normalized Prepared values. Files were exclusively reserved before adoption/input validation, and the previous 33-row byte prefix matched. An initial saved-metadata check used a wrong field-name assumption; its failure is preserved and a corrected check verified the existing files. It did not repeat Go materialization or original observation.

The project Go Reader exposes request/candidate input through `.Input()`, labels and unit weights through `.Supervision()`, and source, role and finite scope through `.Metadata()`. Future features should use only request/candidate description text, excluding supervision, metadata and observation evidence. The [35-row verification script](../../../scripts/verify-next60-thirtyfive.sh) reproduces the complete saved predicates and all 35 data rows and values. It does not launch the original worker or fit a model; actual CI results are recorded separately.

HF and Wiki currently publish33 requests. The [immutable HF33 tag](https://huggingface.co/datasets/JooYoon/riidolaya-shortclaim-next60-development/tree/next60-33-finite-v1) remains intact; HF35 awaits CI and separate publication. Both additions retain existing GJSON group83 development_train, adding no independent source family. Future position/style comparisons need prospective baselines. The preparation target of about60 does not automatically authorize Fit. Domain-specific protected2400 evaluation and the5% necessary utility gate remain separate requirements for final claims. This document grants no broader source rights.

[Progress: Issue19](https://github.com/teamswyg/laya-tools/issues/19) · [한국어](README.ko.md)
