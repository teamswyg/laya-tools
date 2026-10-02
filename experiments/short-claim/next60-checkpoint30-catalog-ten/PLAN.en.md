# Ten existing catalog contracts for the next checkpoint

This selects **ten existing IDs** from the public 120-task Source53 catalog, excluding the first 20 drafts. No new request name or fixture-derived request is created. The 52 finite input proposals and 156 three-candidate dispatch proposals are **unexecuted planning**. New qualification, labels, roles and weights remain 0.

Root now reports ftoa qualified, bringing the next60 subset to three requests/eight labels. The earlier QA-pending proposal remains an immutable historical snapshot. The old 79-request corpus and inactive models are unchanged.

| Order | Existing ID · source | Observable difference | Planned negative |
|---|---|---|---|
| 1 | `go53-pflag-annotation-ownership` · flag.go | Registration/getter slice mutation must not change stored annotations | Copy registration only; return borrowed getter slice |
| 2 | `go53-pflag-bounded-count` · count.go | Range 0..2 rejects increment, negative and oversized text without state change | Assign before range/error check |
| 3 | `go53-pflag-single-assignment` · flag.go | Long/shorthand repeated assignments preserve first value | Treat raw spellings as different flags |
| 4 | `go53-pflag-sensitive-default` · flag.go | Fixed help marker preserves actual default/parse behavior | Redact default but expose optional value |
| 5 | `go53-pflag-deferred-function` · func.go | Later parse failure invokes no deferred callbacks; success keeps argv order | Flush queue on failed parse from defer |
| 6 | `go53-pflag-text-error-value` · text.go/errors.go | Retain name and causes while removing synthetic supplied value from error | New redacted string loses cause |
| 7 | `go53-mapstructure-hooks-errors-is` · decode_hooks.go | Is/As finds failed causes and nil-success stops alternatives | Continue after nil-success |
| 8 | `go53-mapstructure-remain-conflict` · mapstructure.go | Reject direct/embedded output remain conflicts before unmatched assignment | Inspect direct fields only |
| 9 | `go53-mapstructure-tag-precedence` · mapstructure.go | New option distinguishes absent from explicitly empty tag | Collapse both using Get |
| 10 | `go53-pflag-unknown-token-report` · flag.go | Report actual unknown token coordinates, excluding recognized values/post-- tokens | Treat every dash-prefixed token as unknown |

[The JSON proposal](CATALOG10.v1.json) supplies 3–6 inputs/Wants per request, three concrete candidate directions, hashes and prerequisites. Wants were not inferred from runs, and are not approved truth/labels before Root semantic/caption/contract review. A missing new API is not a compile-only baseline failure: adapters must call actual original APIs and preserve state, error and callback evidence.

Source reading rejected two initial suggestions. Compose already preserves typed-nil pointers, so no new typed-nil improvement gap was demonstrated. Metadata iteration may happen to appear sorted in one run, while strict error names are already sorted; failure cannot be promised. Text error causes already support Is/As, so only supplied-value redaction is a new gap.

Dedup is bounded to allowed public metadata. Admitted UUID Parse/Scan and Humanize Ordinal are excluded, as are semver/doublestar non-training families. No direct pflag/mapstructure family appears in the allowed 79-request metadata, but this does not prove semantic independence from all 17 original authored components. The qualified native2 Or request requires message concatenation. The new Join coding contract requires cause objects, stays in the same hook family and does not reuse native2 labels.

The ten contracts share **two already-exposed groups**: connected Cobra/pflag group 77 and mapstructure group 78. They are not ten unseen families or heldout data. Reuse existing source closures and the durable before-dispatch/after-return writer for a real future batch. No global finalizer reset, OS files, service, real-time clock or model is needed. Original package initialization and unexpected failures must still enter actual accounting.

Start with the first four direct state/rendering APIs on the retained pflag package, then the callback/Join/decoder contracts. Instrumenting original argv coordinates is the later item. Do not replace the original parser with a fake or hide candidate state resets. Preserve separate pflag BSD, mapstructure MIT and copied Go 1.23.4 text.go notices, linked to the new actual compiler selection. Existing notice pins do not approve a new execution or grant blanket rights.

Current 3 + prepared 4 + prior six + these ten could produce 23 qualified requests. Registry 20 + these ten could reach 30 only if every contract qualifies. If the remaining seven registry holds are expensive or MPL-deferred, choose seven other existing catalog contracts using the same process rather than counting more fixtures. Keep development 30/60 checkpoints and protected 2,400 requests per claimed domain. Small code-behavior probes are not proof of learned utility, low RAM, GPU use or cost savings.

Go/build/test/helper/original/oracle/model/protected/HF/HTTP/shared modifications were 0 in this task. This is nonblind source planning by the earlier ftoa/controller/converter/reader author, not independent oracle, label, legal or compiler approval.
