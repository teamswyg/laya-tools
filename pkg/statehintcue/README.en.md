# Experimental question punctuation cue

People and agents can try `riidolaya question-cue` without installing a model.

```sh
riidolaya question-cue --role prose --text "Should I check the cache again?"
riidolaya question-cue --role metadata --text "Summary"
riidolaya question-cue --jsonl < requests.jsonl
```

Each JSONL line has two string fields, such as `{"text":"Should I check the cache again?","role":"prose"}`. Private text can arrive over stdin and is not echoed in the output. `cue.question_punctuation=true` means the limited lexical rules observed a question mark. Discard partial cues when `status` is `guarded` or `unknown`. Existing `state-hint` models, controls and display criteria remain separate.

JSON escapes follow Go's standard JSON decoding. String validation applies to decoded logical text, without a claim to preserve the original escape spelling or Unicode scalar representation.

`Inspect(ctx, text, role)` observes literal `?` or fullwidth `？` in caller-declared
prose. It uses no learned model, dictionary, confidence or probability. A marker
does not establish grammatical question intent, a desired label, permission,
owner identity or successful work. The caller's `prose` / `metadata` / `code`
role is an explicit input classification, not owner proof.

All roles require valid UTF-8, no NUL and at most 4096 bytes. An explicit supported
role and non-nil context are required. Metadata and code return no cue. Prose
scans the whole input and checks cancellation throughout, including after a
candidate marker. Invalid, cancelled or unbalanced input returns no partial cue.

The deliberately small lexical rules are:

- Literal backtick runs open code and only a run of exactly the same length
  closes it. Inline code and fences use this same rule, without line-position,
  indentation or backslash-escape interpretation. Other run lengths inside code
  are content. This is not a full Markdown parser; unmatched delimiters reject
  the whole observation.
- `http://`, `https://`, `ws://`, `wss://` (ASCII case-insensitive) and `/` begin
  excluded URI-like spans at input start, after Unicode whitespace, or after an
  opening bracket/quote, `:` or `=`, or a matched closing code delimiter. A span
  ends at whitespace or a backtick.
  Adjacent closing brackets and punctuation remain excluded conservatively;
  other schemes and arbitrary URI syntax are not parsed.
- A question mark at the end of a URI-like token is ignored, even if the whole
  sentence could be a question. A separate `?` after whitespace or a later
  prose word ending in `?` is observed. No grammar is inferred to recover a
  question with no outside marker.

The result serializes fixed source/guard values and booleans only. It never
returns text, URLs or hashes. `ModelUsed`, `ActualVerified`,
`StateChangeProposed` and `MutationExecuted` are always false. The source name
`unlearned_question_punctuation_v1` distinguishes this experiment from frozen
model predictions and the existing speech-act rules; those remain unchanged.

Tests and benchmarks use invented public lexical fixtures. They verify the
bounded scanning contract, not independent semantic accuracy or production
latency. Inspection is O(n), has no shared cache or locks, and uses standard Go
only. Integration, owner authorization and any annotation decision remain
outside this package.

[Development cost records](COSTS.development.json) cover one repeated invented fixture only. Zero Go allocations do not imply zero process memory or measured model/GPU cost.
