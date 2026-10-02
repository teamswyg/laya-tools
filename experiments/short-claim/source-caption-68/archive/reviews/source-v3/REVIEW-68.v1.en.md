# Independent pre-execution source and finite-Want review68

The v3 plan agrees with the pinned default flags and the statically read paths for the three version inputs and two names. Its counter wording was corrected. **This review does not approve full finite-result coverage by every short caption.** In particular, the complete NewVersion sentence says coercion is attempted; it does not enumerate or guarantee every success bit and normalized string. No original API observation has run yet.

The reviewer is `semantic_review60_prep`, a different reader from the root PLAN68 author and the upstream wording authors. This is not blind review: the reviewer previously read60/61, independently checked63 wording, authored64 projection, checked65 metadata, and saw source/metadata navigation. This report and its proposed English request drafts are AI-assisted reviewer-authored artifacts; they do not receive an independent author review merely by appearing here.

The final plan binding is `PLAN-68.v3.ko.md`, SHA `6a3b9eb06fcd1223392dab767b7e739304927f14d239879d5cc0dbe4ac057831`. Compared with the initially assigned v2 SHA `d1de515f6b5fc5946a9fee3b74a6f1c7767483b5747326e8d65e3065dfeed4c1`, only one worker-paragraph line differs; request, caption and Want descriptions are unchanged. Root's source-reading correction of v1 DetailedNewVersionErrors=false before any reported execution remains historical evidence. Both real declarations are true; DetailedNewVersionErrors controls detailed errors only on the CoerceNewVersion=false path.

## Complete source choices and budgets

Byte ranges start at zero and exclude the end. Prefixes, backticks, LF and grammar symbols remain exact. The JSON contains full raw text, source-file SHA, quote SHA and line ranges.

| Proposed choice | Original range | Bytes / normalized words | Prior63 location |
|---|---|---:|---|
| Complete Strict doc.go sentence | doc.go [386,497), lines13–15 |111 /17| `/quotes/0` |
| Complete NewVersion doc.go sentence | doc.go [498,590), lines15–16 |92 /16| New68 selection; absent from63 quote list |
| Whole g-star grammar item | match.go [197,257), line15 |60 /7| `/quotes/17` |
| Whole g-globstar grammar item | match.go [258,307), line16 |49 /5| `/quotes/18` |
| Complete g-mid-component sentence | match.go [1298,1457), lines40–42 |159 /26| `/quotes/24` |

All fit512B/32words. The following complete NewVersion example sentence is161B/36words and remains unselected, intact. No words were removed to pass the cap. Both1078B MIT notices are preserved separately as full original bytes. All15 original manifest artifact pins and full notices were checked. These are preservation facts, not individual or human-only authorship certification.

## What source reading supports, and what captions omit

StrictNewVersion's three-part SplitN and digit-only checks support the predeclared accepted Want `[true,false,false]` for `1.2.3`, `v1.2.3`, `1.2`. These are not executable Got values. The only-valid sentence faithfully states a restriction; it is not a promise to accept every valid version or to specify every exact error/no-panic field.

With CoerceNewVersion=true, NewVersion reaches the anchored loose regex and coerceNewVersion. The regex allows an optional v and omitted patch, omitted patch becomes zero, and String formats the numeric three-part version. Thus accepted `[true,true,true]` and strings `1.2.3,1.2.3,1.2.0` have bounded static support. **NewVersion is a new observation scope.** Strict-only57/59/63 evidence does not prove it previously ran. Exact String values are observation metadata, not a requested caption promise. Nevertheless, the selected sentence only says attempts; prefix/missing-patch success bits must not be retroactively declared full finite caption coverage.

Match uses slash, validation=true and caseInsensitive=false. Plain-star backtracking cannot cross slash; segment-start double star followed by slash permits zero or nested directories. A dot after ** falls through to ordinary-star handling. For fixed patterns `src/*.go`, `src/**/*.go`, `src/**.go` and names `src/main.go`, `src/lib/main.go`, the predeclared matched Wants `[true,false]`, `[true,true]`, `[true,false]` agree with these source paths.

g-star and g-globstar are grammar items, so shared `src/`, `.go`, whole-name slash matching and exact candidate-instance bindings must be explicit. The actual g-mid-component example is `path/to/**.txt`, not a literal enumeration of `src/**.go`. A bounded structurally equivalent source link is possible, but the quote is not a universal placement/error/panic/Windows/filesystem contract. err=nil, supported and no-panic remain separately checked observation fields.

## Proposed English inputs and connection boundary

Existing model inputs are not translated. These are new English **proposal drafts**, not finalized input/Want/worker freezes or new supervision truths:

1. `For inputs 1.2.3, v1.2.3, and 1.2, accept only the complete unprefixed version; reject both other inputs.` —105B/21words.
2. `For inputs 1.2.3, v1.2.3, and 1.2, accept all three, coercing the prefix or missing patch into a semantic version.` —114B/24words.
3. `Using whole-name slash matching with src/ and .go fixed, match both src/main.go and src/lib/main.go, allowing zero or nested directories.` —137B/25words.
4. `Using whole-name slash matching with src/ and .go fixed, match src/main.go but reject src/lib/main.go; wildcard matching must stay within one path segment.` —155B/28words.

The semver entrypoints, Version, String observer, validators and regex/init/flags stay one connected family. All three glob pattern instances, Match, recursive helpers, validation and sentinel stay another. Two requests or extra wrappers do not create independent families. No graph union or role allocation into existing membership ran. Preserving utils.go's compile closure does not authorize filesystem calls.

Prior g01–g05 and Strict prefix/missing-patch observations have exact original pointers for overlap navigation. They are not copied into new Got or new labels. All three NewVersion inputs and the mid-component main position are new observation positions. Existing72/216, unknown21,17groups/known16, roots and prior review states remain unchanged.

## Remaining concrete execution boundary

v3 distinguishes twelve primary entrypoints, Strict3/New3/Match6, from three NewVersion String observers. Strict.String is not called. It makes no individual initializer/helper instrumentation claim and includes initialization in child CPU/RSS. The synthetic test package imports no upstream code.

The existing plan requires final exact English input bytes, caption-to-entrypoint/pattern bindings, Want/observer schema, and full source/license/compile-closure/worker/plan pins before the first actual run. This report does not approve readiness, fit, generalization or LLM savings. Actual use of these quotes may document a different preexisting wording origin from project-authored descriptions, but two related families, function-name shortcuts and reused literals do not prove human-only authorship, independent populations or training eligibility.

One AI-assisted source/text review and one metadata serialization succeeded. The first synthetic test attempt failed to compile because this reviewer's helper missed a comma; after correction all three tests passed with race. The first failed source and output remain preserved. The ledger and receipt record real attempts/failures. Original APIs/init/AST/parser/formatter, actual Features, roles, fit, model/judge/paid APIs, protected-final reads, shared writes, Git and external publication are all zero. These counters describe separate processes/APIs, not zero AI-agent participation or zero total collaboration cost; ordinary Codex collaboration cost was not measured.
