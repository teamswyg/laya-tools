# Second fit: adding a ranking objective still fails utility

The small claim scorer provides hints about which candidate to check first. This trial adds within-request positive-versus-negative ranking to the existing candidate BCE objective. It retains 8,192 FP32 coefficients, seed1729, learning rate0.1, L2 0.0001, batch128 and at most50 epochs. The new objective uses one predeclared λ=1 trial. It is a sibling objective, not a compressed child of the first model.

| Same development validation | Strongest lexical_ordered control | First BCE model | Second BCE+ranking model |
|---|---:|---:|---:|
| Total candidate checks: lower is better | 27 | 31 | 33 |
| Top1 correct:10 answerable requests | 5 | 2 | 1 |
| Top3 correct: same10 requests | 10 | 10 | 10 |

Relative check reduction is **−22.2222%**. Check reduction and Top1 fail, so default behavior does not activate the model. `PublicationQualified=false` and `ProductionReady=false` remain unchanged. There was no retry, seed/λ search or threshold relaxation.

Selected epoch50 has train BCE0.6311540918774127 and validation BCE0.6885500655576609. Parent-mean pair loss is0.593613195205774/0.6927276831485621. Training loss improvement did not reduce actual candidate checks. The original earliest-minimum validation BCE epoch-selection rule remains in place across the50 recorded epochs.

The original76-request cached projection, roles, truth and masks are reused. Fit rows are91/45 known candidates, retaining18/1 zero-weight rows. Train/validation have56/28 candidate pairs, including16/2 zero-weight pairs;16/10 parents contribute active pair loss. Utility evaluates all candidates of15 known validation requests:10 answerable and5 no-answer;5 unknown requests are excluded. Calibration supplies no fit rows. This validation also informed the next objective after the first failure, so it is not an independent final test. Separate2,400 fresh final requests per domain and source/authoring diversity remain unmet.

This is the second cumulative corpus fit and one local trial with zero retries. Go1.27.1 CPU1 and a256MiB Go soft heap target produced outside observed peak RSS28,999,680B, footprint26,444,304B and controller wall2.969253125s. The worker includes fitting, verification and serialization while the same Mac also ran the full Go test suite. This is neither isolated serving latency nor a matched speed comparison. No GPU, Laya encoder or paid LLM call was used.

The new coefficient file is32,792B, SHA `539bd0de1c4b08af99645ebc113eeaa7a7aeaef8dcf4282d10e3b3336e737c6f`. The reference decoder expands coefficients into65,536B float64 values; file size is not total memory. Model bytes are excluded from Git. Original result SHA is `e01d7daf8a5f26cd5578674e87f996a578d747f6622221308da4624a71995220`.

Independent numerical review used saved coefficients and projection to verify scores, weighted BCE, selected-epoch pair loss, candidate orders and check counts. It reran no original trainer, feature, baseline or role API. Intermediate epoch coefficients, repeatability and generalization were not independently reconstructed.

[Original result](results.json) · [Outside execution ledger](ROOT-ACTUAL-LEDGER.v1.json) · [Independent numeric review](RECEIPT-SECOND-FIT-NUMERIC.v1.json) · [Source CI](SOURCE-CI.v1.json) · [한국어](README.ko.md)
