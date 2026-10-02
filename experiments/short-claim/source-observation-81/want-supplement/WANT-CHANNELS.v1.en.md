# Source81: proposed Want channels and call routes

This is a **pre-execution observation-scope proposal**. It preserves the source-first document's 4 goals, 32 complete input objects, order and nulls while defining what later Wants should distinguish. No final literal Want, Got, candidate-satisfaction labels, parents, roles or weights are created. Original package compile/import/init/API/tests/go list and observer/model execution remain 0. Root can confirm the route and channels before separately sealing literal Wants.

The original proposal SHA is `dcd7fbd62cbc53d35608069c1c179aeb97353acf2d7c9aaebc858d8db22f1962`; the source-first handoff SHA is `14a03815901b4a0b0b47adc0f84c25cc56f00e32c56d854bb3fa97b58d96c0b4`. Fixtures have no separate IDs, so original goal_id plus `/goals/G/input_only_fixtures/I` identify them. A convenience ordinal is not an original fixture ID. The 32 fixtures are not 32 independent problems or new parents.

## Hook return observations

For each mapstructure fixture, the proposed route creates one hook through its factory and calls DecodeHookExec once. The anonymous factory result is not directly asserted to named DecodeHookFuncType. The whole decoder, weak hook and cached helper are excluded. Destinations use valid zero reflection values and true defined types: `defined Time` and `defined []string` are not aliases. Original `source_type=int,input_literal="7"` is explicitly interpreted as builtin int 7 without rewriting the original object. Nil sources, invalid reflection and named-string inputs are not added.

Return channels separate normal return, interface nil, actual dynamic type, tagged value, error and panic. Tags are string/int/[]string/time.Time or unsupported. String, integer and slice values retain complete content and nilness. For every returned time.Time, including an error return, Date, Clock, Nanosecond and Zone would each be called once to retain date, clock, nanoseconds, zone name and offset. Zero time returned with an error is not discarded. time.Time.String/Format/MarshalJSON are excluded to avoid implicit formatting. A direct hook return is not described as destination assignment.

Source shows that the time hook parses only builtin string with exactly time.Time as destination. Other destinations and int sources return original data. The slice condition is `t == reflect.SliceOf(f)`, not merely Slice kind. Empty builtin string returns a nonnil empty []string, and splitting retains order, duplicates and empty interior items. These are source branch findings, not observed results or proofs for all reflection inputs.

## pflag state observations

Each fixture uses a fresh NewFlagSet(..., ContinueOnError) and StringSliceVar. FlagSet name `scope81`, flag name `items`, empty usage, no shorthand/deprecation/custom normalizer are **proposals awaiting Root confirmation**. Names affect error text and must be fixed before literal Wants. Lookup occurs once after registration. The returned Flag pointer's Changed field is read directly each step. SliceValue.GetSlice is called after registration and after every operation; the observer copies an owned snapshot while separately preserving nilness. Original input/default/Replace arguments are not changed.

GetStringSlice, additional direct Value.String/Type calls and extra alias mutations are excluded. GetStringSlice takes a different String/CSV reparsing route. Registration's internal Value.String call is documented as an uninstrumented source path, not added as an independently measured callback. A potentially aliasing GetSlice view remains distinct from an observer-owned snapshot.

Source shows first successful Set replaces, later successful Set appends, and a CSV error returns before modifying value or internal changed. Replace/Append neither parse CSV nor change internal changed. Public Flag.Changed is updated by successful FlagSet.Set. The unexported changed field is not claimed as directly observed. Later fixed operations and snapshots remain present after an error. Replace(null), Replace([]), empty Set, duplicates and commas are not merged or normalized.

## Error and panic channels

Each returned error separates nil, actual `%T` type and outer Error text. If it supports Unwrap, the proposal calls that interface explicitly **once** and retains the returned cause's nilness, type and Error text separately. Public time.ParseError Layout/Value/LayoutElem/ValueElem/Message and csv.ParseError StartLine/Line/Column plus ErrQuote/ErrBareQuote/ErrFieldCount identity comparisons use direct fields, without extra Error/Unwrap calls. pflag GetFlag/GetValue getters are also excluded.

The time hook's private timeParseError versus cause time.ParseError, and FlagSet.Set's InvalidValueError versus cause csv.ParseError, are distinct channels. Unexpected types retain their actual type and unsupported details. Pinned fmt source shows `%T` does not call Error/String. However, pflag InvalidValueError.Error can format csv.ParseError.Error internally, which can format sentinel.Error. Those internal source-path calls remain uninstrumented/null. The explicit-call total is not a total of every internal invocation.

Panic remains distinct from returned errors. Only type and primitive string text are recorded, without arbitrary Error/String calls. Undispatched channels remain null; failed getters/error observations retain null plus phase/reason. Original inputs, fixtures and guards are not adapted to expectations, and Wants will not be rewritten after actual outcomes.

## Proposed call budget

Fixed inputs contain 12 operations in the Set goal and 14 in the mutator goal. Sixteen registration snapshots plus 26 transition snapshots imply 42 GetSlice calls.

| Explicit upstream route | Proposed count |
|---|---:|
| Hook factory / DecodeHookExec | 16 / 16 |
| NewFlagSet / StringSliceVar / Lookup | 16 / 16 / 16 |
| GetSlice / FlagSet.Set / Replace / Append | 42 / 18 / 6 / 2 |
| Route subtotal | 148 |
| Expected outer Error / outer Unwrap | 6 / 6 |
| Expected cause Error (stdlib) | 6 |
| Expected time getters (stdlib) | 16 |
| Expected explicit observer dispatch total | 182 |

Source suggests 3 time errors and 3 CSV errors; these are not measurements. A conservative bound allows at most one outer Error, Unwrap and cause Error for each of 42 potentially error-returning actions, plus four time getters for all 16 hook returns: **338 calls** (148+126+64). This remains a safety-budget proposal before Root confirmation, not an achieved or observed count. Nil causes, unsupported Unwrap, type differences and panics retain skip reasons and nullable channels. This does not measure total CPU or every reflection, input-decode, formatting and internal registration call.

## Remaining boundaries

The source handoff contains 12 newly acquired original bodies plus one historical Go LICENSE, totaling 13 preserved assets. Runtime Go bodies comprise 3 mapstructure files and 4 pflag files. This is not a compiler closure. Acquisition/selection/notices/full initializer review remain pending for 5 mapstructure and 38 pflag bodies. This work neither performs nor claims completion of the separate source82 task. Selected files expose a pflag CommandLine NewFlagSet initializer site and a mapstructure reflect TypeOf/Elem initializer expression; actual initializer callback/return counts remain null.

Go 1.27.1 time/format.go, time/time.go, time/zoneinfo.go, encoding/csv/reader.go, fmt/print.go and LICENSE/VERSION were read and byte-pinned. Date-only Parse's UTC route and zero time's nil Location route remain distinct from Local timezone loading. These are fixed cached-standard-library references, not a newly attested official Git revision or a complete stdlib closure proof. Full upstream MIT/BSD and copied-Go notices remain referenced in the existing archive without relicensing. mapstructure/wordwrap, pflag/Cobra and Go author-flow independence remain unresolved.

This is a separate Want author, but AI-assisted/nonblind collaboration already exposed to source-first notes and earlier results. No human-blind, independent-source, generalization or training-eligibility claim is made. Literal Want/Got, truth, roles and weights remain null. New actual parents, Features, Project, Fit, paid/model/native executions remain 0. The [channel proposal](CHANNEL-PROPOSAL.v1.json) and [handoff](HANDOFF.v1.json) bind exact original pointers, pins and scope.
