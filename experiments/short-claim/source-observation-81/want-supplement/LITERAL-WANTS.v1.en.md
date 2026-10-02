# Source81: 32 literal Wants sealed before execution

[WANTS.v1.json](WANTS.v1.json) records finite observation expectations authored by reading source for 32 original inputs across four behavior goals. It contains no results obtained by running the original functions or an equivalent behavior oracle. Original full input objects, order, `goal_id`, JSON pointers, and `expected: null` are preserved. There is no separate original fixture ID; `ordinal` is position metadata.

The observation route was fixed first in [Root's scope confirmation](source81-root-scope-confirmation.v1.json). The earlier [channel proposal](CHANNEL-PROPOSAL.v1.json) and [historical handoff](HANDOFF.v1.json) remain unchanged. Their original missing-source status was not rewritten retrospectively. This preparation checked the bytes and SHA values of 57 source/module/license assets supplied by the new Source82 handoff `bb0db6e0…`; its 48 proposed Go 1.27.1 file selections are still not compiler or linkage evidence.

The Want author differs from the Source81/82 source-first author. This is AI-assisted, nonblind work informed by source and prior proposals. It claims neither human-blind judgment, independent source lineage, nor training-truth approval. `Got` and training truth, role, weight, and acceptable sets remain null. New parents, labels, roles, fitting, and model calls are 0; all readiness, qualification, production, and final flags are false.

## Channels to observe

Each hook fixture uses one fresh factory and one DecodeHookExec call. Destinations are valid zero reflect values; defined Time and defined slice are new named types, not aliases. The original `source_type: int, input_literal: "7"` denotes builtin `int(7)`. No full Decoder, weak hook, or cached hook route is included.

Hook results preserve type and string, int, or ordered string-slice channels separately. Every actual time.Time return, including a zero value returned with an error, gets one Date, Clock, Nanosecond, and Zone call. String, Format, and MarshalJSON are excluded. Inapplicable return channels are null.

Every pflag fixture creates a fresh `scope81` FlagSet with ContinueOnError, registers `items` with StringSliceVar and empty usage, then performs one Lookup. One GetSlice snapshot follows registration and each fixed operation. An owned copy preserves nil versus nonnil empty; the public Flag.Changed field is read directly. No additional GetStringSlice, Value.String, Type, or alias mutation is included. The initial snapshot has no operation error return channel, so `operation_error: null`; a successful operation has `error_nil: true`. These states are different.

For a returned error, observe one outer Error, one supported Unwrap, and one nonnil cause Error. Read public time.ParseError/csv.ParseError fields and CSV sentinel identities directly. Do not separately call CSV ParseError.Err.Error. Unexpected actual types must retain their type and unsupported detail. Primary, getter, and error-method panics remain separate; only primitive strings become panic text. Other panic payloads must not invoke Error or String.

## Source-derived return expectations

The time layout is `2006-01-02`. Only the first four fixtures are expected to return time.Time; the others retain their original string or int.

| Fixture position | Input / destination | Expected return |
|---|---|---|
| 0 | `2024-02-29` → time.Time | Date 2024/2/29, Clock 0/0/0, Nanosecond 0, UTC/0, nil error |
| 1 | `2024-02-30` → time.Time | Zero time Date 1/1/1; outer timeParseError and cause ParseError |
| 2 | Empty string → time.Time | Same zero time; outer timeParseError and cause ParseError |
| 3 | `bad` → time.Time | Same zero time; outer timeParseError and cause ParseError |
| 4 / 5 / 6 | Valid date → string / *time.Time / defined Time | Original date string, nil error, null time channel |
| 7 | int `7` → time.Time | int 7, nil error, null time channel |

The invalid day's outer text is exactly `parsing time : day out of range`: the wrapper adds a space before the leading colon in ParseError.Message. Its cause text is `parsing time "2024-02-30": day out of range`, with empty LayoutElem and ValueElem. Empty and `bad` inputs have LayoutElem=`2006` and empty Message. Their common outer text is `parsing time as "2006-01-02": cannot parse as "2006"`; cause text, Value, and ValueElem still preserve the input. Every exact string and public field is a JSON literal.

The comma-slice goal expects, in order, `[]` (nonnil), `[a]`, `[a,a,b]`, `[a,"",a]`, and `["",""]`. Destinations []int and defined []string retain the original `a,b` string; int input remains int 7. Duplicates and empty items are preserved.

Pflag always starts at `[default]`, Changed=false. The first successful Set replaces; later successful Sets append; errors preserve both values and state. An empty first Set produces a nonnil empty slice and Changed=true. Input `a,"unterminated` is expected to yield CSV StartLine=1, Line=1, Column=16, and ErrQuote, wrapped in InvalidValueError. A first success after an error still replaces. An error after success keeps the previous value. Input `"a,b",a` yields `["a,b",a]`.

Replace and Append do not change the public Flag.Changed field. Replace null produces nil; Replace [] produces nonnil empty; Append `x,x` appends one literal string. A first Set after Replace or Append replaces the previous value. If Set has already succeeded, a subsequent Set appends to a replacement. All 42 snapshots explicitly preserve these distinctions. No mutation-based general ownership guarantee is tested.

## Expected budget and unmeasured work

| Explicit dispatch | Expected count |
|---|---:|
| Hook factory / DecodeHookExec | 16 / 16 |
| NewFlagSet / StringSliceVar / Lookup | 16 each |
| GetSlice / Set / Replace / Append | 42 / 18 / 6 / 2 |
| Outer Error / Unwrap / cause Error | 6 each |
| Date / Clock / Nanosecond / Zone | 4 each |
| Total | 182 |

Per-fixture totals for the four goals are `[6,9,9,9,2,2,2,2]`, `[2,2,2,2,2,2,2,2]`, `[6,8,6,8,9,11,11,6]`, and `[6,6,6,6,8,8,10,10]`. Root's conservative fixed maximum is 338. The expectations comprise 148 original route dispatches plus 12 upstream error methods, giving 160 original public dispatches and 22 explicit stdlib dispatches. None is an actual execution count.

The internal hook called by HookExec, registration's Value.String, cause.Error within wrapper formatting, and CSV/time helpers are not dynamically instrumented in this explicit budget. Initialization is source evidence only; actual init calls and returns remain null. Original compile, import, init, API, tests, go list, worker, and controller execution are 0.

## Checks, failures, and rights

[Metadata checks](LITERAL-CHECKS.v1.json) bind all 32 original objects by pointer, order, raw/canonical SHA, and nulls; all 57 Source82 files; 23 historical source-evidence ranges; and 17 new stdlib ranges. Eight cached Go 1.27.1 assets comprise six source files plus VERSION and LICENSE. Of two stdlib helper compile attempts, the first failed due to a nested map type annotation. Its source was preserved; only that annotation changed before one successful metadata main execution. No behavior oracle or original test ran. [LITERAL-LEDGER.v1.json](LITERAL-LEDGER.v1.json) preserves this history.

Rights references remain the [pinned mapstructure MIT LICENSE](https://github.com/go-viper/mapstructure/blob/52aa5c6dc1d27226460807054ca2107b2d54fb2d/LICENSE), [pinned pflag BSD-3 LICENSE](https://github.com/spf13/pflag/blob/c966cfef47379dcb01e7929504d66d94b540945b/LICENSE), and actual Go notices in Source81/82. Root MIT/BSD terms do not relicense copied Go code. The cached Go 1.27.1 LICENSE has a separate recorded SHA; no new full source redistribution occurred here. Family, alias, helper, and author-flow diversity is not cleared. These 32 finite fixtures are neither 32 independent parents nor the protected final target of at least 2,400 fresh examples per domain.
