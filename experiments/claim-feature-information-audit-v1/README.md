# Bounded feature-information audit v1

[한국어](README.ko.md) · [Results and original probes](results.json) · [Exact extraction helper](replay/main.go)

Development diagnostic only, 2026-10-08. Exactly 12 newly authored, unlabelled
standalone probe pairs (24 strings) were extracted through
`statehintwide.ExtractContextual`; 23 inputs were accepted and one exceeded the
byte limit. All inputs are development-exposed and excluded from final 2400 and
selectors. These are feature probes, not claim gold, independent semantic cases,
accuracy measurements, or a model qualification. No model was loaded, scored,
predicted, fitted, selected, or modified; no protected cases were inspected.

The byte-exact measured extraction helper is published in `replay/main.go`,
alongside `results.json`. All 24 strings are newly authored public-safe unlabelled
development probes, with no protected work data. The results retain every
original input, its UTF-8 SHA-256, byte/rune/word counts, activated and nonzero
bin counts, dense feature SHA-256, and comparisons.
Feature SHA-256 is over all 2048 float32 bit patterns in ascending bin order,
encoded little-endian. No model artifact or protected corpus is published.
Execution used Go 1.27.1 darwin/arm64 with `GOWORK=off`, `GOTOOLCHAIN=local`,
`GOPROXY=off`, `GOMAXPROCS=2`, and `GOMEMLIMIT=128MiB`.

## Fixed contract and source evidence

The schema is `utf8-word12-char2345-fnv1a-signed-logtf-l2-2048-v2`.

* [features.go, lines 122–133](../../pkg/statehintwide/features.go#L122):
  reject more than 4096 **UTF-8 bytes**, invalid UTF-8, NUL, or a missing
  workspace; otherwise read the entire string and lowercase each Unicode rune.
  The extractor does not truncate an oversized input or accept a prefix.
* [features.go, lines 135–155](../../pkg/statehintwide/features.go#L135):
  words are maximal runs of Unicode letters/numbers. Emit every word and every
  adjacent word pair; punctuation does not reset the previous-word state.
  There is no additional token, sentence, clause, or cue cap.
* [features.go, lines 157–165](../../pkg/statehintwide/features.go#L157):
  emit all overlapping character 2–5-grams over the complete lowercased rune
  sequence, including whitespace and punctuation. These lengths count runes,
  not bytes. No offsets, full sequence, or longer explicit n-grams are stored.
* [features.go, lines 93–120](../../pkg/statehintwide/features.go#L93) and
  [166–179](../../pkg/statehintwide/features.go#L166): namespace-prefixed
  FNV-1a uses UTF-8 bytes; all terms share 2048 bins, with sign from the hash high
  bit. Sum signed occurrences in each bin, apply signed `log1p(abs(total))`,
  then L2-normalize. Opposite signs can cancel a bin to zero.
* [contextual_api.go, lines 13–44](../../pkg/statehintwide/contextual_api.go#L13):
  the view borrows the workspace and exposes activated bins (including zeros)
  and word count. The helper copies numeric values before any reuse.
* [claims/model.go, lines 214–235](../../pkg/statehintclaims/model.go#L214) and
  [272–288](../../pkg/statehintclaims/model.go#L272): Scores/Predict accept text
  only and use this extractor; annotation, event facts, task state and provenance
  are absent from the numeric features. Word count also supports the no-word
  control. [Lines 180–186](../../pkg/statehintclaims/model.go#L180) accumulate
  model products in sparse iteration order; that order is a numerical caveat
  below, not a queried model result.

Korean inflections are retained as whole space-delimited letter runs and local
character fragments, without morphology analysis or stemming. Lowercasing drops
original case. There is no Unicode normalization, dedicated question/negation/
quotation parser, speaker attribution, tense field, or syntactic scope field.

## Actual extraction results

`Changed bins` compares normalized dense float32 bit patterns. Global L2 scaling
can change many bins after a local edit; this count is not a count of linguistic
cues. L2 deltas below are rounded for display; JSON retains full precision.

| Pair | New unlabelled framing | Words A/B | Changed bins | L2 delta |
| --- | --- | ---: | ---: | ---: |
| p01 | English final `.` / `?` | 6/6 | 114 | 0.253461 |
| p02 | English insertion of `not` | 6/7 | 148 | 0.474338 |
| p03 | English addition of enclosing quotes | 6/6 | 112 | 0.249712 |
| p04 | English current / future wording | 6/6 | 150 | 0.836990 |
| p05 | Korean `닦았습니다.` / `닦았습니까?` | 3/3 | 60 | 0.614283 |
| p06 | Korean `접었습니다.` / `접지 않았습니다.` | 3/4 | 80 | 0.888244 |
| p07 | Korean `다듬고 있습니다.` / `다듬을 예정입니다.` | 4/4 | 100 | 0.937158 |
| p08 | Swap amber/cobalt clauses between quoted and direct text | 17/17 | 0 | 0 |
| p09 | Swap amber/cobalt clauses between acceptance and negated acceptance text | 28/28 | 0 | 0 |
| p10 | Swap Korean color clauses between quoted and direct text | 16/16 | 0 | 0 |
| p11 | 4096-byte synthetic input: final `.` / `?` | 2048/2048 | 8 | 0.081180 |
| p12 | Valid Korean UTF-8: 4096 / 4097 bytes | 372/rejected | — | — |

The seven local edits are distinguishable in these probes. This establishes
retention of some lexical/punctuation information, not correct recognition,
generalization or linear separability. p01 A has 112 activated bins but 110
nonzero bins: two signed aggregate bins canceled completely. No particular cue
is alleged to have been canceled.

The three scope permutations have identical dense float32 bits, activated-bin
sets, and word count, with different sparse iteration order. Their dense digests
are respectively:

* p08: `cda7a7d237f11bd7184a9a8fce62f9101b213ee3eca73f34a82b9ae2cf9ca551`
* p09: `52e2751f83f73c55587f3e279af465feca31bbc600ce35aa79678d186b702fef`
* p10: `f01283910a6ba0277ec95a04229a3e67e343912f051450bd264d15d5ab3a7a23`

Each pair swaps two complete clauses that share their first/last four runes and
first/last word. Consequently the same local word/character fragment multisets
remain while clause-to-attribution placement changes. This is a dense-vector
collision and evidence that the vector is not a full sequence encoding. These
particular object swaps may be intended invariances for the three generic claim
heads; no independent target distinction or classification error is established.
Different summation order can affect floating-point results, so this audit does
**not** assert bit-identical Scores/Predict outputs. No classifier was queried.

p11 is exactly `("q " repeated 2047 times) + "q."` versus the same input ending
in `?`: both are 4096 ASCII bytes and all 2048 words are counted. The terminal
cue changes eight bins. p12 is 93 complete repeats of
`청록색 고리를 차례로 닦습니다. `, three padding spaces and a period (4096 bytes,
1678 runes), versus the same valid UTF-8 string with `!` appended (4097 bytes).
The latter returns `statehint: invalid or oversized input`. Rejected-input zero
fields in the JSON are placeholders, not extraction measurements.

## Bounded implications

The byte limit is the input cap. The maximal accepted word count is 2048:
2048 one-byte letter/number words plus 2047 one-byte separators require 4095
bytes, while 2049 words require at least 4097. No smaller token cap or cue cap
was found. Korean consumes more bytes per Hangul rune, so its usable rune budget
is smaller; this is byte budgeting, not Korean word truncation.

For the future fixed-contract, data-only sibling experiment, local cue coverage
and Korean inflection coverage remain plausible data questions. Attribution/scope
relations outside the local fragments are a separate representation hypothesis
only when independently justified target distinctions require them. Hash
aggregation and normalization can further dilute or cancel terms. These probes
neither select a new feature contract nor establish that a feature change is
needed; the fixed future contract is unchanged.

## Replay

From the repository root with Go 1.27.1:

```sh
GOWORK=off GOTOOLCHAIN=local GOPROXY=off GOMAXPROCS=2 GOMEMLIMIT=128MiB \
  go run ./experiments/claim-feature-information-audit-v1/replay
```

This reads only its embedded original probes and calls feature extraction. It
does not load or query a model, read a dataset, or modify the feature contract.
The output should reproduce the recorded numeric results on the same source
contract. Rejected-input placeholder fields remain non-measurements. The
publication copy was not used to create new probes or classifier results.
