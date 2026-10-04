# Compare small hint representations

`riido-hintpreview` is a Go experiment showing how the same text becomes features: legacy word hashing, hashing that retains comparison symbols, and a small interval relation array. It needs no model download or Python. It performs no generation, training, or router change.

```sh
go build -trimpath -o bin/riido-hintpreview ./cmd/riido-hintpreview
./bin/riido-hintpreview <<'JSON'
{"request":"Return keys in increasing order, keeping x >= lower and x < upper.","candidates":["AscendRange calls the iterator for every value in the tree within the range [greaterOrEqual, lessThan), until iterator returns false.","DescendRange calls the iterator for every value in the tree within the range [lessOrEqual, greaterThan), until iterator returns false."]}
JSON
```

People can read candidate indexes in `agreement_order`; agents can consume the JSON as a suggested checking order. Indexes are zero-based positions in the input. No candidate is removed or executed. Each recognized relation match scores +1 and conflict −1. These scores are not probabilities or truth verdicts. Ties retain current display order.

The `legacy` and `symbol` SHA values fingerprint the ordered feature array: uint16 LE index followed by float64 LE value. Valid text without retained symbols matches legacy values and order exactly. Retaining `<` versus `<=` can change the fingerprint. Both hashes have8192 dimensions but different meanings, so legacy weights must not be reused.

`relation_columns` contains eight columns: match/conflict for direction, lower bound, upper bound and stopping. Eight candidates' float32 feature payload is256B. This excludes the program, text, stack, JSON and total RSS. Coverage in `grammar_recognition` is a recognition bit mask: direction1, lower2, upper4, stop8. Failure to recognize text does not make it false.

The scoped grammar accepts `Return keys/all keys ... in increasing/decreasing order` with explicit `keeping/with x >= lower`, `x < upper` and related bounds. Reversed operands are supported. Candidate text uses the complete traversal-comment grammar in the example. Descending intervals have the upper bound on the left and lower bound on the right. Unknown endpoint words retain partial recognition; empty endpoints, nesting and extra commas are unsupported. This is not a general language interpreter for negation, quotations, conditionals or Korean prose.

Stop counts are absent, and `stop_conflict` is always0 in the current supported grammar. Unrecognized requests, such as an empty-tree instruction, produce all-zero scores and retain display order. Labels, roles and training masks are unaffected.

Input is limited to16KiB JSON,512 UTF-8 bytes per string,32 words,64 tokens, and1–8 candidates. Relation word counts use ASCII word runs before camel-case splitting; symbol hashing counts Unicode words. Both JSON fields are required; null, duplicate keys, extra fields and invalid Unicode are rejected. Errors return exit1 with stderr and do not echo source text. This opt-in tool is independent of existing code search and Codex routing defaults.

See the [actual development diagnostic and resource measurements](../../experiments/short-claim/text-representation-preview/README.en.md).
