# What to do after the second fit failed

**Prioritize semantic and wording diversity in new development data.** These results do not justify another lambda/seed/threshold search or ternary compression. This claim model suggests the order in which candidates should be verified; it does not determine correctness, execution or approval. Neither fitted model currently supports default activation or an effectiveness claim.

| Same15 validation requests,45 candidates | Strong lexical control | First BCE | BCE+pair lambda1 |
|---|---:|---:|---:|
| Candidates checked | 27 | 31 | 33 |
| Acceptable candidate first | 5/10 | 2/10 | 1/10 |
| Acceptable within top3 | 10/10 | 10/10 | 10/10 |
| Validation BCE | n/a | 0.66580 | 0.68855 |

The five no-answer requests require11 checks regardless of order. The other10 answerable requests cost16→20→22 checks. Preserved top3 does not imply improved top1 or cost. Parents5,8,37 each worsened by one check; parent60 improved by one, for a net increase of two. Typed cost improved9→8, while legacy worsened22→25. Inputs, truth, masks, roles and controls stayed fixed and no features were regenerated, so this is a development comparison of one objective change.

Training pair NLL fell0.69246→0.59361. Validation moved0.69441→0.69273, close to the zero-margin reference `log(2)=0.69315`. Only10 of26 nonzero-weight validation pairs placed the positive higher; parent-normalized weighted accuracy was40%, with mean margin about0.00091. Training relationships were learned, but useful transfer to the four validation source groups was weak. This alone does not prove overfitting, convergence or an implementation bug. Both runs hit50 epochs=50 updates; independent synthetic tests checked loss arithmetic and exact nil/default behavior.

**Epoch selection alone is not a convincing explanation here.** Both validation BCE and pair NLL reached their minima at epoch50. Pair loss still differs from verification cost: it seeks to place every positive above every negative, whereas cost improves once any acceptable candidate appears early. No-answer order does not affect cost, although those rows remain original negative BCE supervision. These mismatches matter, but this run does not show that changing only the selector would help.

Saved-array counts show just16 training parents contributed pair gradients, out of31 known training parents with the original zero weights retained; validation had10 eligible parents. Training request vocabulary contained166 unique words and validation82, of which50(about61%) did not occur in training requests. Validation captions contained100 unique words,67 unseen in training captions. These are unique-vocabulary ratios, not token-frequency ratios. Whole-family splitting leaves validation concepts such as `verified`, `inclusive`, `limit` and `release` without direct training-request exposure. Both new upstream families were in training, so this run does not measure held-out upstream generalization.

The encoder-free linear features hash query×caption unigram/adjacent-bigram interactions into8192 signed bins and add one exact-token-recall feature. They represent words and local order without explicit logical scope, temporal meaning or synonyms. Punctuation stripping can lose distinctions when otherwise identical inputs differ only by operators such as `<` and `>`. A particular collision has not been established as the cause: only490 of6107 active validation bins were unseen in training(about8%), carrying5.87% of validation L2 mass. Shared bins do not prove shared semantic interactions. Compression, collisions and semantic insufficiency need separate experiments.

Three options remain:

1. **Expand development data — recommended.** Collect licensed public wording and contrasting candidates for condition scope, inclusive boundaries, before/after and negation/exception. Template repetitions or literal cases are not independent requests. Split by source/author/template families and retain ambiguous truth as unknown. Keep features and objective fixed in the next data experiment to distinguish causes.
2. **A small feature ablation — next.** Separately test a few explicit operator, negation and temporal-relation channels derived only from raw request/caption text. Source IDs, roles, old correct outcomes and acceptable labels never become features. Retain the fast lexical control and opt-in hint scope; change no runtime or model default now.
3. **Redesign the utility surrogate — later.** Predeclare a closer first-acceptable-cost objective/selector separately. Do not choose favorable objectives using this validation set and then claim victory on the same set. With pair lambda1 already failing, continued objective search is lower priority.

Retain the76 examples as development regression evidence. Obtain and freeze a **separate fresh-domain final golden set of at least2400 independent requests**, unused for model/objective selection, before an effectiveness claim. Preserve all candidates, unknown21, no-answer, zero-weight known rows and calibration no-fit. Reducing actual verification work comes before larger GPU use or compression.

Pins: first result `060eaa602f27f7fd9a67306e2fa2a530f0ef4c5c4548858a7818d9dfdb4bb28a`; second `e01d7daf8a5f26cd5578674e87f996a578d747f6622221308da4624a71995220`; cached projection `f68bff5f48747a66f038c17262568c1bd91251a8c81b3f46f7ae5090c8fe4b99`. Sources read: hintlearn/learn.go and pairlearn/learn.go·ranking.go at CI95 commit `d506f58ccf9629e2d2b6ca3cd1766ab20789bf64`, plus unchanged utility.go. [Saved statistics](SAVED-STATS.json) were produced once with stdlib-only loops. One jq display-expression typo was corrected; it was not a data/API failure. New fit/features/projection/roles/labels/inference/benchmark/HF executions in this analysis are zero; AI-assisted collaboration cost is unmeasured. This is a nonblind development analysis by the backend/runner author, not independent effectiveness approval or a new gate.
