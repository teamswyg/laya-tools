# Critical review of a new prefix-only contract68

Narrowing the target to **allowing or rejecting the leading v in the exact input `v1.2.3`** can connect the intended property to existing upstream wording. The introductory bullet is a semver **library capability**, not a NewVersion function comment. Preserve that attribution level and make the source binding to NewVersion explicit.

The whole original bullet has these raw bytes; the actual proposed caption retains its two-space indentation, dash, backticks and final LF:

```text
  - Optionally work with a `v` prefix
```

It is doc.go line9, byte range `[260,298)`,38B/6normalized words, SHA `5780602667a40d7af64f4f879c65194d0e046d78ff78c614f73c7dd6a012453b`. The full doc.go SHA is `ad987af91a2a2c8dd5d01a38364263c53d294918670bed23edfad865d14afc22`. It fits512B/32words without truncation or new paraphrase. The JSON also preserves the full1078B original MIT notice.

Attribution is not a guessed function-name substitution. The same doc.go parsing paragraph explains that StrictNewVersion parses valid-v2 versions while NewVersion coerces a version with a leading v. Pinned NewVersion→coerceNewVersion→anchored loose regex also has the optional `v?` path. Thus an explicit candidate contract that **instantiates the package-level optional-prefix capability as pinned NewVersion(v1.2.3)** has source/documentation support. Strict's existing complete sentence, digit-only checks and the source comment distinguishing a leading v from SemVer itself support the opposite property for this fixed input.

Under that binding, this narrow prefix-permission/rejection semantic property has contextual coverage. The bullet alone does not guarantee all inputs of a named NewVersion function, nor permission in every semver API. If the candidate-to-NewVersion binding is missing or merely invented by the report, leave the coverage gap unresolved. Function binding is provenance, not added replacement text in the raw runtime caption.

Proposed English requests follow. They are AI-assisted reviewer-authored drafts, not root's final contract freeze or an independent author approval:

1. `For v1.2.3, reject the leading v prefix when parsing a semantic version.` —72B/14words.
2. `For v1.2.3, allow the optional leading v prefix when parsing a semantic version.` —80B/15words.

The model target is this literal's prefix behavior only. Existing three-literal Wants and String auxiliary observations can remain, but missing-patch success, exact strings, errors, panic and resource behavior are outside this caption target. Expected Strict rejection/New acceptance is static source reasoning, not observed Got or new supervision truth. The `1.2.3` control remains an auxiliary observation position for parsing the nonprefix portion.

Both opposite requests and both entrypoints share Version, validators, regex/init/flags and String observer in one whole family. Glob scope is not enlarged. A hypothesis of obtaining another preexisting wording origin remains separate from human-only authorship, independent populations or training readiness.

The immutable v3 report SHA `e1b2eb0e247fd22d98734d14dce535b4d85bf261f1d3eb75344a05d2bf65b814` was checked as input. Its generic-coercion fullcoverage limits and all v3 artifacts remain unchanged. No final v4 plan has been read or approved here. Root's final exact input/quote binding, target/aux schema and source/worker freeze must precede the separate actual execution record.

This phase had one successful private metadata serialization and zero failures. Original APIs/init/AST/parser/formatter, actual features, roles, fit, inference, judge and paid calls remain zero. New truth/acceptable sets, shared writes, Git and publication are zero. Ordinary AI-assisted Codex review cost was not measured; zero separate API/process execution is not zero total AI cost.
