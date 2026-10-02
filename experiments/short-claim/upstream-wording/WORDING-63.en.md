# Public upstream wording assets 63 — private preparation

This preparation extracted **25 complete sentences, definitions or grammar items** from the two retained upstreams without authoring new wording. Every quote meets the 512-byte/32-normalized-word limits: 1,818 bytes total, at most 26 words. Internal line breaks, comment prefixes and original typos are preserved. New requests/captions/candidates, expected values, acceptable sets, truth/unknown changes and training-readiness decisions are zero.

`quote-catalog-63.json` was generated using owned Go arrays. Each quote binds its repository, immutable revision, original/retained file, complete file bytes/SHA, zero-based half-open raw byte span/SHA, one-based inclusive lines, original URL and MIT license file/SHA. All 15 original artifacts/99,035 bytes were checked, and **both complete original MIT licenses** are included verbatim. Four documentation spans already bound in 59 and 57's 16 strict/9 slash rows/287 field references were checked. Reading these records is neither candidate execution nor a new replay of 57.

- Semver: `Masterminds/semver@61fc460d28283a91c53be65c2e0f20b494ac8ad9`, `version.go`/`doc.go`; wording about `StrictNewVersion` and its already observed getters.
- Doublestar: `bmatcuk/doublestar@8b690afa33319b0a1869367f594e53977e38bc99`, `match.go`; wording about slash-based `Match` only.
- Full original licenses and provenance are sidecar metadata. Source names, hashes, copyright notices and review states are not features. Preserving MIT notices does not approve future training or weight publication eligibility.

The quoted words are **pre-existing pinned upstream wording**; 57's inputs and finite Wants are **project AI-assisted authored material**. These origins remain separate. The MIT notices naming Matt Butcher/Matt Farina and Bob Matcuk are copyright attribution, not an audit of individual sentence authorship or proof of human-only authorship. Verification remains false/unknown. This catalog alone does not set `authoring_diversity_cleared=true` or clear `synthetic_single_pipeline`.

## Expressed and unexpressed observation fields

`s-doc-strict` describes strict semantic-version parsing, and `s-parse-result` describes a Version-or-error result. “Only valid versions” does not imply acceptance of every valid string. Error definitions express some input conditions and upstream sentinel names, but not project adapter tags such as `semver_empty`, nil returns, error defaults or error precedence. Seven getter sentences describe the meanings of `major/minor/patch/original/string/prerelease/metadata`; they do not enumerate each recorded literal value.

| Existing strict rows | Original wording links | What these words do not express |
|---|---|---|
| s01 empty | `s-error-empty` plus strict/parse sentences | `nil_version`, adapter tag, exact empty error subfields and error-default getter values |
| s02 missing patch | `s-error-invalid` plus strict sentence | Exact three-part grammar or a complete grammar explicitly rejecting this input |
| s03 v prefix | `s-error-characters` plus strict sentence | The specific v-prefix rejection grammar; permissive `NewVersion` results |
| s04 core leading zero | `s-error-leading-zero` | Actual nil/tag/precedence and full grammar identifying the relevant numeric segment |
| s05 empty minor | Generic parse/error sentence | The recorded **NumError/syntax** and `error_func/cause/num`; it is not replaced by an invalid-character sentinel |
| s06 core overflow | Generic parse/error sentence | uint64 bounds and NumError/range; valid decimal syntax is not relabeled SemVer-invalid |
| s07 invalid metadata / s08 empty metadata | `s-error-metadata` | Complete allowed-character/empty grammar and exact adapter/return fields |
| s09 empty prerelease / s11 empty prerelease segment | `s-error-prerelease` | Segment grammar, empty restrictions and exact returns |
| s10 numeric prerelease leading zero | `s-error-leading-zero` | Full numeric/non-numeric prerelease grammar; it is not relabeled with an invalid-prerelease tag |
| s12 error priority | `s-error-metadata` plus generic sentence | Metadata-first policy when several errors coexist |
| s13 zero version | Strict plus getter sentences | Exact 0.0.0 fields, nil=false and empty success-error subfields |
| s14 maximum core | Strict plus getter sentences | Maximum uint64 and overflow boundary; saved large integers remain RawMessage/uint64 |
| s15 fields preserved | All seven getter sentences including `s-get-original` | Exact saved values are not enumerated; complete String format, numeric and prerelease/metadata grammar |
| s16 overflow prerelease accepted | Strict plus getter sentences | Large numeric prerelease acceptance width, exact fields and **comparison correctness** |

The 16 strict rows retain their 14 checked fields: `error_cause/error_func/error_kind/error_num`, `major/metadata/minor/nil_version/original/panicked/patch/prerelease/string/supported`. Topic links indicate which field meaning is partly expressed. **No quote is a scalar approval of all 14 fields.** Original Want/Got/matches and pointers remain unchanged, with no new oracle. Observations lacking a declared Want, including `error_message`, counters and Versions, were not promoted to truth.

`g-star` describes non-path-separator sequences; `g-globstar` describes zero or more directories for `/**/`; `g-forward-slash` describes `/` splitting; `g-whole-name` describes matching the complete name. `g-mid-component` is an original complete 26-word sentence that explicitly states equivalent results for `path/to/**.txt` and `path/to/*.txt`. It is not a newly shortened caption.

| Existing slash rows | Original wording links | Scope limit |
|---|---|---|
| g01 one segment / g02 does not cross slash | `g-star`, slash splitting | The sentence does not enumerate each literal Want |
| g03 zero directory / g04 nested directory | `g-globstar`, `g-component` | The narrow `/**/` grammar item, not a complete glob-grammar guarantee |
| g05 mid-component double star | `g-mid-component`, `g-component`, `g-star` | Meaning of the pattern form stated by the original sentence; no expansion to every pattern policy |
| g06 trailing globstar zero | Entrypoint/whole-name/component wording | **The short `/**/` item does not expressly state `src/**` matching `src`**; preserve the finite observation only |
| g07 whole-name anchor | `g-whole-name`, `g-star` | Whole-name matching rather than substring matching |
| g08 star empty name | `g-star` | An any-sequence item, not an explicit statement of the empty-name literal Want |
| g09 globstar crosses slash | Entrypoint/whole-name/component wording | **The short `/**/` item does not expressly cover a whole-pattern `**` without slashes**; preserve the finite observation only |

The nine slash rows retain four error fields and `matched/panicked/supported`. Quotes partly express `matched` semantics, without guaranteeing saved no-panic, wrapper support or empty error subfields. `g-error` preserves the upstream ErrBadPattern description as context only. Existing malformed-pattern discrepancies were not erased or converted into a new strict-validation gate. No new `NewVersion`, `PathMatch`, filesystem or other-property behavior/expectations were created.

## Actual attempts and next-use boundary

Both wording-generator invocations exited 0; one independent byte/reference-verifier invocation also exited 0. Self-review after the first generator found invented row-ID aliases and overly broad sentinel/topic links. The first catalog is preserved as `quote-catalog-63.attempt1.json` and withheld. The corrected final catalog was generated into a new exclusive file. **One first self-review failure** is reported separately from zero helper execution failures. All three synthetic race commands passed: two actual test executions and one final cached test result. The last actual race run took 1.606s. All three vet commands passed. The verifier checked original bytes, 25 quotes, 16+9 rows and 287 fields; it is not complete semantic approval or a new independent-origin experiment.

`PREPARATION-LEDGER-63.json` records attempts and failures; `quote-verification-receipt-63.json` records mechanical checks. Original APIs/AST/formatter/Rebind/SourcePins/Bind/Generate/behavior/model/fit/paid/role/seed/newlabel/candidate/acceptable/Git/shared edits/publication remain 0. No upstream quote was arbitrarily shortened, translated or rewritten. These Korean/English notes are explanatory analysis, not model inputs.

The next use must freeze narrow input scope, full candidate preservation, supervision eligibility and group relationships under the existing plans. Exact upstream wording assets remain separate from training readiness, diversity clearance and demonstrated utility. No new gate, human-approval step or readiness decision was added.
