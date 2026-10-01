# PDCA 52: owned comparisons and preparation for 2,400 requests

[한국어](RESULTS-52.ko.md) · [Frozen plan](plan-52.json) · [Numeric results](results-52.json) · [Acquisition plan](../../docs/golden-set-acquisition.en.md)

Three distinct public development tasks were attempted with each explicit requested profile, `gpt-6-sol / low` and `gpt-6-luna / low`. **All six fixed CLI attempts were recorded.** Independent checks accepted all six declared candidate closures, but **only five attempts had zero exit and complete core usage**. One reached the fixed 120-second deadline with unknown usage.

This is a small development comparison from one repository and two code families. Two requests are exact comment edits. It is not the final 2,400-request routing evaluation or representative capability/savings evidence. Repeating prior requests does not create additional distinct final requests.

## Plan

Stopped [plan51](plan-51.json), original records, and parser reanalysis remain unchanged. We neither resumed it nor silently substituted its unsupported model. A separate plan52 was frozen after read-only catalog refresh in an isolated authenticated home with no prior settings/cache. It listed 6 Sol and Luna with low, and omitted 6.1 Sol. Hidden catalog entries and authentication are not published.

Catalog metadata does not attest access or backend identity. Completed actual attempts verified request-specific access for both requested profiles. No provider-attested identity appears in these CLI records, so every `observed_model` remains `unknown`. [Official catalog/access guidance](https://developers.openai.com/siwc/token-sharing-open-source/codex-app-server)

- Plan SHA: `a41adfef1c45067a9988e0d8d755e3b7d4687fd043b176f5ab6b258e9362d67d`
- Pre-outcome commit: `d202ff042106ad36209ef4cdee6ca1e1c9e64514`
- Executor: clean `-trimpath` build of unchanged main `146b02b9c37e6d90a11386050ccfdffc036c113e`
- Executor SHA: `db450397f2f64715bbbcb463990c058b3b0954df97197d2f325b1af82fb132a5`
- CLI `0.158.0`; executable, Go1.27.1, public base, and task specifications pinned before inference
- Six attempts maximum, concurrency one, main deadline120s, independent verification45s, no executor retry/resume/fallback

These authorized public self-use attempts used the existing ChatGPT login. No API-key billing, cloud job, endpoint, or purchased credits were added. Subscription quota/money and backend internal request/retry counts remain unknown.

## Do / Check

| Ordinal | Task / requested profile | Main ms | Verification ms | Closure | Termination / usage |
|---|---|---:|---:|---|---|
| 1 | preview comment / Sol6 | 26,644 | 1 | accepted | exit0 / complete |
| 2 | preview comment / Luna | 17,555 | 1 | accepted | exit0 / complete |
| 3 | budget comment / Luna | 17,211 | 4,140 | accepted | exit0 / complete |
| 4 | budget comment / Sol6 | 25,788 | 4,189 | accepted | exit0 / complete |
| 5 | minimum context / Sol6 | 119,929 | 4,536 | accepted | deadline / unknown |
| 6 | minimum context / Luna | 39,510 | 4,381 | accepted | exit0 / complete |

Verification times are exact recorded values. Main time excludes preflight, parent preparation, and independent verification. Acceptance covers only the declared source closure; added files and caches outside it are unassessed. The two preview comment acceptances have no behavioral tests; the remaining attempts ran pinned independent checks. Candidate-authored tests were not acceptance evidence.

| Ordinal | Input | Cached subset | Output | Optional reasoning |
|---|---:|---:|---:|---:|
| 1 | 54,198 | 31,360 | 407 | 0 |
| 2 | 31,362 | 23,808 | 222 | 0 |
| 3 | 41,851 | 33,792 | 234 | 0 |
| 4 | 65,186 | 52,992 | 428 | 0 |
| 5 | unknown | unknown | unknown | unknown |
| 6 | 98,881 | 83,200 | 1,145 | 0 |

Cached input and reasoning are subsets of input and output, respectively. The five complete attempts sum to291,478 input,225,152 cached input, and2,436 output: **known-subset totals only**. Missing ordinal5 prevents whole-pilot usage/spend accounting. It is not zero. All main-process time sums to246,637ms; verification sums to17,248ms. Remote model CPU/GPU/memory was unobserved.

Ordinal5 demonstrates the need to separate candidate acceptance from execution and usage. Its candidate passed independent checks, but the process did not exit successfully and no completed-turn usage arrived. It was not relabeled as coding inability, silently excluded to claim whole savings, or retried.

All started attempts recorded durable start/terminal evidence, captured-byte hashes/counts, process-group cleanup, and copied-authentication removal. Public `attempt-52-01` through `06` JSON was byte-identical to private durable terminal records. Raw traces, authentication, local absolute paths, and model weights remain private.

## Act

The two comment requests now support paired acceptance and usage observations. Luna had lower observed time/tokens on these two small cases; this is descriptive evidence, not workload savings, subscription efficiency, or proof that larger profiles are unnecessary. Unknown Sol6 usage on the behavioral request prevents full three-request cost comparison.

The previous three CPU Laya predictions all abstained at threshold0.9 and remain historical context. We did not reinterpret their configured 6.1 Sol standard mapping as a new 6 Sol prediction. No new Laya inference, fitting, threshold change, or operational registration occurred. Previous cold whole-process peak RSS near1.48GB does not meet the ultra-small-memory target.

Expansion adds a **separate versioned behavioral contract**. Its first request, `repo-keyword-language-guard`, extends the existing script/language guard to keywords of selected repository candidates. It binds public revision, files, LICENSE/NOTICE, specification, independent checks, and required package/test terminal passes. The verifier is checked using unchanged baseline, independently authored correct implementation, and seeded errors. **There is no actual model outcome for this new request yet.** These six attempts used the unchanged old executor instead.

Before publication, independent review found two wrongly accepted implementations that skipped middle keywords or the third and later selected candidates. The new contract was strengthened across every selected position for candidate counts1–8, all32 keyword positions, and four scripts; both wrong implementations became regression controls. The4,608 keyword assertions are **checks within one request's verifier**, not4,608 new golden requests. The original three frozen task specifications remain unchanged.

The [acquisition plan](../../docs/golden-set-acquisition.en.md) separates development12–24→120→at least240, independent training/selection/calibration, and at least2,400 final requests per claimed domain. It proposes six diverse Go task strata, at least30 final repositories with at most10% from one, linked-group/repository separation, at least4,800 CLI attempts for two-profile final comparison, and explicit missing-cost handling. These are targets, not acquired/executed results. Protected search final2,402 remains unscored; no new weights or production routing was activated.

## Executor cleanup failure found in CI

The first [PR73 CI](https://github.com/teamswyg/laya-tools/actions/runs/36795753610) passed Linux and secret checks but failed the macOS descendant-termination test. That test did not snapshot or accept a candidate whose cleanup was unconfirmed. This executor test failure is distinct from the original terminal records of the six actual model attempts.

Local repetition of the old implementation failed twice in20 trials. An independent controlled reproduction observed `EPERM` for a process group containing only an exited, unreaped zombie, followed by `ESRCH` after reaping. This agrees with [Apple XNU group signaling](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/kern/kern_sig.c). **The CI log itself does not establish its exact errno.**

The repair rechecks transient `EPERM` within the unchanged two-second bound. Only observed group absence via `ESRCH` permits cleanup success; persistent `EPERM` or a live group remains `cleanup_unknown`. Deadlines and acceptance conditions were not relaxed. Error-sequence tests, repeated real termination tests, and independent review passed; CI on the repaired commit remains a separate merge gate.
