# A finite check of three float-string candidates

The saved native run applies six frozen inputs to three candidates. The precise candidate satisfies all six complete contracts. The original has four confirmed output counterexamples; the double-rounding candidate has one. This supports a proposal for one finite development request. This review assigns no role, label, weight or qualified count.

| Candidate | Complete matches | Known mismatches | Unknown |
| --- | ---: | ---: | ---: |
| Original | 0 | 4 | 2 |
| Precise | 6 | 0 | 0 |
| Double-rounding | 5 | 1 | 0 |

The original's first text,1.12, is correct. Its API returns only a string, so the required error-return condition cannot be confirmed. All six error channels remain unavailable/null. The final infinity Want has text=null, meaning no text assertion. It does not require an empty string, and a +Inf output is not evidence that a nonfinite error was returned. Four other known text failures suffice for the finite negative proposal.

For example,1.375 at two fractional places should round to1.38, while the original returns1.37. At nine places, the precise candidate preserves0.000000125. The double-rounding candidate first rounds to six places, loses the value to0, and then increases precision. This single counterexample establishes failure of that finite contract. All six saved Want-to-integer-oracle consistency checks are true. The worker stores consistency Booleans, not full oracle return records; this reviewer did not rerun the oracle.

The four INPUTS.v2 sentences describe only these frozen inputs. API names are absent from the request and candidate texts. Candidate names, sources, roles and labels are inspection metadata and must stay out of model features. The initial INPUTS.v1 used slash-bearing IDs rejected by the existing identifier grammar. Root retained the actual observation IDs and replaced only model-input IDs with flat names. Source reading confirms unchanged text and candidate order. Actual Go input validation is a separate Root step and was not executed by this reviewer.

Six fixtures remain one existing request. The broad0–18-place API draft and its different historical wrong-seed description remain unqualified. Negative finite values,NaN,negative infinity,invalid precision and subnormals are outside this supervision. The humanize family is already development-exposed and should connect to the existing Ordinal whole group76 for development training. It is no new independent family, unseen validation or final-test evidence.

Root completed Start1/Wait1/retry0, with18 candidate observations and6 oracle entry calls. Whole-child wall0.378019625seconds, Darwin MaxRSS5,816,320bytes and Go heap snapshot207,056bytes describe the function-observation tool. Model calls are0. These are not model inference latency, memory or GPU measurements; the heap snapshot is not a peak.

The AI-assisted reviewer authored none of the original,candidates,oracle,observer or Wants. This is nonblind saved-result/source QA. Prior resource-metadata-helper authorship is separately disclosed. New function/model/fit/protected-data operations and shared-repository edits are0. The5% usefulness gate, protected evaluation rules, rights/release scope and Root's adoption remain separate.

- [All18 candidate-input comparisons](FINITE-COMPARISON.v1.json)
- [Pinned evidence and execution scope](SEMANTIC-REVIEW.v1.json)
- [Caption,ID and existing-family boundaries](CAPTION-COVERAGE.v1.json)
