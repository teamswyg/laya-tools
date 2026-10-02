# 77 v1 Want comparison before execution

One item needs correction before execution. **For logfmt-03..07, the extra ScanKeyval call after EOF has no current record and has a source-visible panic path.** Their v1 Wants expect successful return without panic, so the current observation scope and Wants disagree. No panic was reproduced and no upstream API was executed.

The evidence is pinned [logfmt decode.go](https://github.com/go-logfmt/logfmt/blob/804e98fff868b206344991c57a8182172e5ba41e/decode.go#L54), lines 54–87, and v1 `pure/run.go`, lines 148–158. When ScanRecord returns false at EOF, it does not reset `dec.pos`. The fixed Go 1.27.1 Scanner source stores a nil token on the final EOF split. ScanKeyval then slices that nil line at the previous positive position. Empty input in logfmt-01 and empty records in logfmt-02 have position 0. logfmt-08..12 have a nonnil error guard and are different paths.

Root selected a separate v2 scope that does not observe current-record ScanKeyval/Key/Value after EOF with nil Err, and records null/skipped explicitly. Sticky checks guarded by a nonnil error can be preserved separately. Original v1 Wants, source, and prior SOURCE-FIRST-77 notes remain unchanged. This review does not claim that v2 has already passed review.

The other 19 full fixture Wants, plus the pre-EOF values/errors of the five affected fixtures, agree with the source reading. This is finite AI-assisted source interpretation, not observed runtime agreement or new training-truth approval.

| Scope | Source-reading conclusion |
|---|---|
| logfmt-01..02 | Empty input/records, nil values, nil error at EOF |
| logfmt-03..07 | Duplicate-key order; nil bare/equal-empty/quoted-empty values; final line without newline; quoted escapes/Korean/Unicode pair agree before EOF, but the extra EOF call contradicts the Want |
| logfmt-08..11 | Successful prefix items are separate from the failing Key/Value; SyntaxError Msg/Line/Pos and message agree at byte positions 6/11/12/14 on line 1 |
| logfmt-12 | Small Scanner buffer produces token-too-long error and the error guard applies |
| percent-01..04 | Nil input, ASCII unreserved/reserved characters, 00/7f/80/ff bytes, uppercase percent escapes, input unchanged |
| percent-05..12 | Nil empty-success output, plus preserved, mixed-case hex/00/ff, truncated escape/invalid hex/raw non-ASCII return nil bytes with errors |

[rfc2396.go](https://github.com/vincent-petithory/dataurl/blob/d1553a71de50473073e188aa79cebf7f993f20fe/rfc2396.go#L75) returns nil bytes even after a late error; it does not return an accumulated prefix. Invalid-hex messages use the percent rune rather than the invalid byte, displaying `0x25`. The review preserves that actual source behavior. Errors are newly created errorString values, not shared sentinels.

Key/Value may borrow buffers valid only until the next ScanRecord. The observer's immediate nil/length/hex copies are distinct from upstream ownership. These fixtures do not establish actual alias-mutation behavior or complete UTF-8/error/Reader coverage. Arbitrary Readers, files/networks, full DecodeString/goroutine paths, and raw invalid UTF-8 are outside scope.

WANTS24.v1 has 39,675 B and SHA `b21af96b142b023d2d85c8448cfcba6fa29775f2454f5749898b2db18c808af0`. Its 12 logfmt, 4 Escape, and 8 Unescape fixtures total 24 finite cases for two behavior goals. Its mechanical v1 call budgets total 259 planned calls, not actual execution. The 24 cases are not 24 parents, independent problems, or training labels.

The reviewer authored SOURCE-FIRST-77 before seeing the Wants, but did not author this observer or its Wants. Prior source/category exposure means this is not human-blind review. Prior KO/EN notes/proposal hashes and the current Want hash are preserved together in the [receipt](RECEIPT.v1.json); hashes alone do not prove an independent timestamp. All 16 source hashes and both MIT notices plus the Go BSD NOTICE were compared again. Original import/init/API/observer/test execution, models, Features/Project/Fit, role assignment, and shared/public mutations remain 0. Training, production, and final qualification remain false.
