# A small behavior-hint contract using actual upstream wording68

[한국어](README.ko.md) · [Frozen contract](contract-68.v4.json) · [Exact-copy manifest](archive-manifest-68.json)

This is a small step beyond using only descriptions authored in this project: pre-existing public library wording becomes actual candidate text. It records **source-caption binding, fixed invocations and one native Go API observation**, not a trained-model result. Four related English requests retain two whole source families, rather than four independent groups. Original72requests/216candidates/unknown21/17groups, labels and masks remain unchanged.

## The bounded requests

Two semver requests ask only whether the leading `v` is rejected or allowed on the exact input `v1.2.3`. Candidates are `StrictNewVersion` and `NewVersion`. Strict uses the complete valid-v2 sentence; New uses the whole original package-level optional-prefix bullet. The bullet is not a function-specific comment. Its attribution uses the following parsing context and the pinned coercing NewVersion source path. It does not promise every semver function, every input, errors, panics or String values.

Two glob requests allow or reject a nested path under explicit whole-name slash matching, shared `src/`/`.go` premises and the fixed names `src/main.go` and `src/lib/main.go`. Patterns are `src/*.go`, `src/**/*.go`, `src/**.go`. Complete star, standalone-globstar and mid-component quotations are preserved. The original txt example is not rewritten to go; only the same structural placement is attributed. Filesystem traversal, Windows and arbitrary patterns are outside this experiment.

The [contract](contract-68.v4.json) preserves exact English requests, quotation bytes, order and SHA. `target_Want_not_Got` describes each candidate's expected behavior. Both opposite semver requests share `[false],[true]`; their requested goals and proposed acceptable sets `[0]`/`[1]` remain separate. Missing-patch acceptance and NewVersion.String are auxiliary observations, not additional model targets.

## Earlier limits remain visible

| Record | Meaning |
|---|---|
| [v1](archive/plans/PLAN-68.v1.ko.md) | Original plan whose incorrect DetailedNewVersionErrors expectation was caught before execution; preserved unchanged. |
| [v2](archive/plans/PLAN-68.v2.ko.md) | Both flag expectations corrected to true; still no native execution. |
| [v3](archive/plans/PLAN-68.v3.ko.md) | Twelve entrypoints separated from three String observers. Generic-attempts wording could not promise all finite acceptance/normalization; the [held review](archive/reviews/source-v3/INDEPENDENT-REVIEW-68.v1.json) retains that limit. |
| [v4](archive/plans/PLAN-68.v4.ko.md) | A new narrower prefix contract, preserving the [prefix review](archive/reviews/prefix/PREFIX-ONLY-CRITICAL-REVIEW-68.v1.json), [final static review](archive/reviews/final-v4/FINAL-V4-REVIEW-68.v1.json) and [worker-code review](archive/reviews/worker/INDEPENDENT-WORKER-REVIEW-68.v1.json). |

A successful later v4 observation does not retroactively clear v3's insufficient caption. Requests were AI-assisted drafts selected by root. Upstream wording predates those drafts and has a different wording origin; this does not authenticate individual/human-only authorship, generalization or authoring-diversity clearance. Prior reviewer authorship and nonblind metadata exposure remain disclosed in the receipts.

## The actual observation

The [result](archive/actual/results-68.json) and [actual ledger](archive/actual/ROOT-ACTUAL-EXECUTION-LEDGER-68.v1.json) report one first native execution, zero retries and exit0. Direct API adapters count Strict3/New3/Match6=12, with three separate NewVersion.String observers. All fifteen frozen Want comparisons—twelve booleans and three strings—match. Flags remained true before and after. Individual initialization and internal-helper counts were not instrumented.

The checkpoint reviewer, who did not author the requests, provided an unchanged [precheck](archive/qa/precheck/FINDINGS-PRECHECK-68.v1.en.md) and [runtime Got review](archive/qa/runtime/FINDINGS-RUNTIME-68.v1.en.md). This nonblind reader had prior development/precheck exposure and did not rerun original APIs. Observed-derived acceptable sets `[0]`, `[1]`, `[1]`, `[0,2]` match the earlier proposals. The [supervision proposal](archive/qa/runtime/PROPOSED-OBSERVED-SUPERVISION-68.v1.json) retains five positive/five negative positions and proposed masks only for the narrow scope. Actual training weights are null and existing dataset mutations remain zero. Eight unique target API observations are reused across ten candidate positions; seven auxiliaries bring comparisons to fifteen, not model accuracy. This archive copier only copies that separate review and makes no new semantic assessment.

| First-execution measurement | Recorded value | Scope |
|---|---:|---|
| Controller elapsed wall |0.366244042s|Includes controller start/wait overhead. |
| macOS time real |0.36s|Rounded OS output. |
| User / system CPU |0.00 / 0.00s|Rounded to the tool's printed precision, not proof of zero CPU work. |
| Peak RSS |7,225,344B ≈6.89MiB|OS maximum resident set; neither Go heap nor GPU memory. |
| Peak memory footprint |4,735,480B ≈4.52MiB|Separate OS footprint metric, not interchangeable with RSS. |

One cold worker lifetime includes package initialization, metadata guards, APIs, observers and output. These numbers are not isolated API latency, training, model inference, GPU measurements or a general performance distribution. The external controller reserved the attempt before exec and applied a300-second timeout. The256MiB RSS bound was observational; no hard RSS limit was enforced. GOMAXPROCS1 and the256MiB Go heap soft limit are settings, not measured consumption or hard RSS guarantees.

## Archive scope and licenses

The [manifest](archive-manifest-68.json) records each immutable copied asset's SHA, byte size and original snapshot name. Historical receipts keep their original meaning and names. `.go` and `go.mod` files are `.go.txt`/`.mod.txt` archives, not a public runtime port. Historical failed sources and ledgers remain; raw native/synthetic logs are excluded, with their existing hashes retained in the omission list. Models, binaries and controller source containing local absolute paths are not published.

Fifteen upstream originals totaling99035B reference the retained [MIT manifest](../../../internal/publicbehavior/upstream-manifest.json), [complete semver notice](../../../internal/publicbehavior/testdata/upstream/semver/LICENSE.txt) and [complete doublestar notice](../../../internal/publicbehavior/testdata/upstream/doublestar/LICENSE). They are not copied again. [Catalog63](../upstream-wording/quote-catalog-63.json) retains origin, revision and raw quotation provenance. New maintainer Go code uses Apache-2.0.

Within this68 experiment and archive copy, features, roles, fitting, model/judge/paid trials and protected-final execution remain zero; training_ready=false. Zero separate model/API processes do not mean zero AI-assisted Codex use or collaboration cost; that cost was not measured. This connects public wording with small finite native observations without proving training eligibility, broad accuracy, token savings or the utility of a minimal-memory model.
