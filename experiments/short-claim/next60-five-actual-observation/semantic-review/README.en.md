# Actual comparison of saved results for the next five tasks

We compared Root's saved first execution once using the sealed standard-library Go comparator. Across 23 fixed inputs and three candidates each, the 69 rows contain 42 satisfied, 21 unsatisfied, and six unknown outcomes. The 228 predicates contain 185 known matches, 32 known mismatches, and eleven unknowns. All five reference candidates satisfy every Wanted predicate on their 23 inputs. Nine candidates have known counterexamples; one error-writer baseline candidate remains unknown.

The table uses zero-based fixture positions. T means satisfied, F means unsatisfied through a known counterexample, and U means unknown. Every reference row is T. The full report retains each predicate's Wanted, observation, presence and rule, plus actual error type/logical kind/Error-method return/text/candidate return/panic channels.

| Task | Baseline row vector | Baseline counterexamples | Negative row vector | Negative counterexamples |
| --- | --- | --- | --- | --- |
| mapstructure nil config | F F F T | 0, 1, 2 | T F F T | 1, 2 |
| Cobra error writer | U U U U | None | T F F T | 1, 2 |
| Cobra required annotation | F F F T T | 0, 1, 2 | F F F T T | 0, 1, 2 |
| Cobra context preflight | F T T T | 0 | F T T T | 0 |
| Humanize BigBytes precision | F F F T U U | 0, 1, 2 | F F F T T T | 0, 1, 2 |

The six unknown-only rows are the error-writer baseline's four inputs and the BigBytes baseline's negative-size/invalid-precision inputs. `PrintErr` exposes no returned n/error, so the actual spy-writer count was never substituted for a returned value. BigBytes baseline return errors are also unavailable. Thus the error-writer baseline is not a false label. The BigBytes baseline has independent known text counterexamples while two rows remain U.

Two baseline candidate panics were observed for nil config and empty required annotation. Their rows contradict Wanted panic=false, with unavailable error observations retained as U. The nil-Result baseline returned logical kind unclassified_error instead of declared invalid_result. Its negative candidate changed three nil metadata slices into empty slices, violating unchanged. The error-writer negative reported short/failing writes as n=3 and nil error. Canceled-context comparison candidates ran unwanted initializer/pre-run/run hooks once each. BigBytes baseline rendered 1.5/1.5/1.0 instead of 1.50/1.501/1.000000000000001; its negative mutated the input despite correct formatting.

All three candidates for the `missing_required` fixture had normal tracked Error returns and exact text `required flag(s) "item" not set`. Their actual Go type is `*errors.errorString`. The logical kind does not establish a new Go sentinel or typed error. Comparison never replayed errors.Is, Error, callbacks, or original candidates.

The comparator ran exactly once, exited zero, and produced zero stderr bytes. Saved input is 142,918 bytes with SHA-256 `bfaa50f313a851348069e7d2fec4cd6680dd6dab963683285fe307f8b384af1e`. The report is 86,140 bytes with SHA-256 `c4db2ae8f33b29161af66fba5468edfd507a2b1c579c65c811a0c746478bd0bc`. We checked saved-row consistency with 67 normal returns, two candidate panics, 23 Error returns, and final sequence 184. External process and journal Sync success belong to Root's separate evidence; this comparator does not establish them. Internal API and original-startup counts remain null.

The reviewer did not author Wanted, candidates, or observer, but knew earlier source review and correction feedback, so this is nonblind. The reviewer authored the comparator. Its twenty synthetic tests and vet passed before source sealing; Root independently reviewed all of json.go/compare.go before authorizing actual comparison. Those preparation files remain unchanged. Synthetic passing controls are separate from actual candidate/model evidence.

This comparison performed zero worker replays, original imports/init/API calls, model/training/projection/scoring operations, protected-data reads, or Git/HF/network work. Actual comparator/worker memory and GPU performance measurements are absent here. labels_assigned=false and qualified=false remain unchanged. Root separately decides candidate adoption, labels, and training from known positive/counterexample evidence. No general correctness outside these inputs, model quality, cost savings, license guarantee, or new public-CI success is claimed.

[Detailed comparison](FINITE-COMPARISON.v1.json) · [Receipt](RECEIPT.v1.json)
