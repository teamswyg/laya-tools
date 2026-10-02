# Validate once and reuse the input

The short-claim tool normalized input during loading and repeated that work when ranking. `ValidatedInput` retains the checked input in private fixed arrays, removing the repeated normalization. CLI usage and output remain compatible.

```sh
riido-shortclaim --stream --baseline lexical_ordered < requests.jsonl
```

Go callers can use `shortclaim.LoadValidated(reader)` or `shortclaim.ValidateInput(input)`, then `.Rank(kind)`. `.Prepared()` returns an independent value copy. The zero value cannot rank; the existing `Rank(Prepared)` still performs its original validation. This contract checks input shape and normalization, rather than truth or execution authority.

On the public eight-candidate fixture, ranking allocations/op changed **81→0** and full-request allocations **291→210**. The old observation enabled CPU profiling while the new single observation disabled it, so causal latency/RSS improvements remain unconfirmed. The increased ending Go heap is also recorded. This adds no maps, locks, SIMD, model, GPU or training, and provides no evidence of LLM token savings.

[Measurements and limitations](RESULTS-VALIDATED-INPUT-71.v1.en.md) · [Comparison values](COMPARISON-VALIDATED-INPUT-71.v1.json) · [한국어](README.ko.md)
