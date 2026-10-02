# Source80 saved-result integrity check

Root's one recorded native run **matches all24 frozen predictions**. Original input/order, full Wants/Got, partial UUIDs, receiver/null channels, counters and file pins agree. The reviewer, checkpoint_cli_45, authored the Wants; this is nonblind author-side post-run record integrity, not independent oracle approval or new source semantics. No original API, observer/controller, Go test/build, model/Features/Fit was rerun.

| Finite goal | Normal primary returns | Nonnil errors | Matches |
|---|---:|---:|---:|
| UUID Parse | 8 | 3 | 8 |
| UUID.Scan | 8 | 2 | 8 |
| Original int Ordinal | 8 | No error-return channel | 8 |

Primary reserved/returned is24/24; explicit Error reserved/returned is5/5, split into one upstream URNPrefixError and four stdlib errorString calls. Unclassified errors, primary/Error panics and differences are0. Every Got channel was checked against the unchanged full Want rather than trusting reported matches alone. raw32's fe failure slot versus standard36's00, Scan receiver preservation on errors/empty inputs, typed nil versus nonnil empty bytes, and Ordinal's null error channel remain intact. There are three goals from two source families, not24 independent parents or training labels.

Worker final, partial and root stdout are each67,635bytes SHA256 `ce793a8633c75520e226bc894c2126caecac85cda7c809e56fd41792c9a877a6`; all three are byte-exact. Frozen-plan SHA `27186964abe2387f6dc8e5f33d3188bfb5dd5c4b56bf0cf7cdf4914a9b16f81a` differs from the prior v2 draft only by its frozen=false→true bytes. Want remains `c93ac2b7…`, input `a6110ab8…`, binary `55f662cb…`. The33 result source pins equal the frozen plan, and all33 actual source file SHA checks passed. Reservation retains zero direct-call counters and null records.

Root's saved ledger reports Start1, joined Wait, exit0, retry0 and no timeout/overflow/start/wait failure. The generic76 controller ledger schema is preserved and bound to source80's worker/plan/binary SHA. The root preflight metadata helper had one compilation failure in two build attempts before native execution; its original record remains preserved. Root's actual native run was once. This QA used one saved-metadata jq checker invocation, success1/failure0.

Aggregates extracted from private OS logs describe the **whole child**: real1.21seconds, user0.02, system0.07, maximum RSS18,104,320bytes (17.265625MiB), peak footprint15,680,016bytes; controller wall1.217474375seconds. This includes initialization, setup,24 observations, checkpoint/fsync and persistence. It cannot establish per-function costs or model-inference performance. Individual Go heap/GPU usage was not measured. The256MiB Go soft heap is not an OS RSS hard cap.

Four namespace MustParse→Parse sites are static startup evidence. Actual initializer callback/return counts remain null and individual instrumentation false; internal Scan recursion/Parse/fmt are likewise uninstrumented. Labels/roles/weights/actual parents/model/Features/Fit remain0, and training-ready/production-ready/qualification/protected-final remainfalse. This does not approve all-input safety, RFC validation, source diversity,2400-case success or model quality.

[MECHANICS.v1.json](MECHANICS.v1.json) and [RECEIPT.v1.json](RECEIPT.v1.json) contain safe aggregates and hashes only. The private frozen plan/argument record, raw OS log/stdout, native binary and checker helper are omitted from public copying; their original SHA and omission reasons are retained. Original Wants, source and actual results were not modified.
