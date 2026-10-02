# Go input validation reused once

Human and agent CLI/JSONL usage stays the same. Input normalization previously ran during both loading and ranking; a validated value now reuses that work.

```sh
riido-shortclaim --stream --baseline lexical_ordered < requests.jsonl
```

Go callers can use `shortclaim.LoadValidated(reader)` or `shortclaim.ValidateInput(input)`, then call `.Rank(kind)`. Validation checks input shape and normalization; it grants no correctness or execution approval. Existing `Rank(Prepared)` remains available.

The public8-candidate fixture observed ranking allocations **81→0** and full-request allocations **291→210**. Different profiling settings prevent a causal latency or whole-memory improvement claim. Full race/vet checks and original score/order/fallback regressions passed. Historical measurements retain their exact original source verification.

[Usage, measurements and limits](https://github.com/teamswyg/laya-tools/blob/main/experiments/short-claim/validated-input-runtime-71/README.en.md) · [Learning results](Second-Claim-Fit-72-EN) · [한국어](Validated-Input-71-KO)
